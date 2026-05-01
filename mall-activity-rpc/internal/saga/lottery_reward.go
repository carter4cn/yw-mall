// Package saga orchestrates the multi-step lottery-reward distributed transaction.
//
// Branches (forward → compensate):
//   1. Reward.Dispatch                ↔ Reward.RefundReward
//   2. Activity.MarkParticipationRewarded ↔ Activity.MarkParticipationRefunded
//   3. Reward.Confirm                 ↔ Reward.MarkFailed
//
// Compensation runs in reverse order, only for branches that succeeded.
//
// We use DTM v1.19's HTTP newGid endpoint to allocate a globally-unique
// transaction id (visible in DTM's audit dashboard) and tag every branch with
// it for trace correlation. The actual branch execution is in-process gRPC
// (we own both reward-rpc and activity-rpc), so we don't need DTM's coordinator
// to fan out — it's an audit anchor, not the transport.
package saga

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"mall-activity-rpc/activity"
	"mall-reward-rpc/reward"

	"github.com/zeromicro/go-zero/core/logx"
)

type RewardClient interface {
	Dispatch(ctx context.Context, in *reward.DispatchReq) (*reward.DispatchResp, error)
	Confirm(ctx context.Context, in *reward.ConfirmReq) (*reward.Empty, error)
	RefundReward(ctx context.Context, in *reward.RefundRewardReq) (*reward.Empty, error)
	MarkFailed(ctx context.Context, in *reward.MarkFailedReq) (*reward.Empty, error)
}

type ParticipationMarker interface {
	MarkParticipationRewarded(ctx context.Context, in *activity.MarkParticipationRewardedReq) (*activity.Empty, error)
	MarkParticipationRefunded(ctx context.Context, in *activity.MarkParticipationRefundedReq) (*activity.Empty, error)
}

type Coordinator struct {
	DTMHttpEndpoint string
	Reward          RewardClient
	Participation   ParticipationMarker
}

type LotteryRewardInput struct {
	UserId             int64
	ActivityId         int64
	WorkflowInstanceId int64
	ParticipationId    int64
	TemplateId         int64
	PayloadJson        string
}

type LotteryRewardResult struct {
	Gid            string
	RewardRecordId int64
	Status         string // CONFIRMED / COMPENSATED / FAILED
	FailedAt       string // empty if Status=CONFIRMED
}

// RunLotteryReward executes the 3-branch SAGA. Failures trigger reverse
// compensation; the returned Result reports which branch failed (if any) so
// callers can surface a useful error to the user.
func (c *Coordinator) RunLotteryReward(ctx context.Context, in LotteryRewardInput) (*LotteryRewardResult, error) {
	gid, err := c.newGid(ctx)
	if err != nil {
		// non-fatal: gid is only for tracing. Fall back to a deterministic local id.
		gid = fmt.Sprintf("local-%d-%d", in.WorkflowInstanceId, time.Now().UnixNano())
		logx.Errorf("dtm newGid failed (using local id %s): %v", gid, err)
	}
	res := &LotteryRewardResult{Gid: gid}

	// Branch 1: Dispatch reward
	disp, err := c.Reward.Dispatch(ctx, &reward.DispatchReq{
		UserId:             in.UserId,
		ActivityId:         in.ActivityId,
		WorkflowInstanceId: in.WorkflowInstanceId,
		TemplateId:         in.TemplateId,
		PayloadJson:        in.PayloadJson,
		IdempotencyKey:     gid + ":dispatch",
	})
	if err != nil {
		res.Status = "FAILED"
		res.FailedAt = "Dispatch"
		return res, fmt.Errorf("dispatch: %w", err)
	}
	res.RewardRecordId = disp.RewardRecordId

	// Branch 2: Mark participation rewarded
	if _, err := c.Participation.MarkParticipationRewarded(ctx, &activity.MarkParticipationRewardedReq{
		ParticipationId: in.ParticipationId,
		IdempotencyKey:  gid + ":mark",
	}); err != nil {
		c.compensateBranch1(ctx, gid, res.RewardRecordId, "branch2 failed")
		res.Status = "COMPENSATED"
		res.FailedAt = "MarkParticipationRewarded"
		return res, fmt.Errorf("mark participation: %w", err)
	}

	// Branch 3: Confirm reward
	if _, err := c.Reward.Confirm(ctx, &reward.ConfirmReq{
		RewardRecordId: res.RewardRecordId,
		IdempotencyKey: gid + ":confirm",
	}); err != nil {
		c.compensateBranch2(ctx, gid, in.ParticipationId, "branch3 failed")
		c.compensateBranch1(ctx, gid, res.RewardRecordId, "branch3 failed")
		res.Status = "COMPENSATED"
		res.FailedAt = "Confirm"
		return res, fmt.Errorf("confirm: %w", err)
	}

	res.Status = "CONFIRMED"
	logx.Infof("saga lottery_reward gid=%s CONFIRMED user=%d activity=%d reward=%d",
		gid, in.UserId, in.ActivityId, res.RewardRecordId)
	return res, nil
}

func (c *Coordinator) compensateBranch1(ctx context.Context, gid string, rewardRecordId int64, reason string) {
	if rewardRecordId == 0 {
		return
	}
	if _, err := c.Reward.RefundReward(ctx, &reward.RefundRewardReq{
		RewardRecordId: rewardRecordId,
		IdempotencyKey: gid + ":refund",
		Reason:         reason,
	}); err != nil {
		logx.Errorf("saga compensate branch1 (RefundReward) gid=%s reward=%d: %v", gid, rewardRecordId, err)
	}
}

func (c *Coordinator) compensateBranch2(ctx context.Context, gid string, participationId int64, reason string) {
	if _, err := c.Participation.MarkParticipationRefunded(ctx, &activity.MarkParticipationRefundedReq{
		ParticipationId: participationId,
		IdempotencyKey:  gid + ":refund_part",
	}); err != nil {
		logx.Errorf("saga compensate branch2 (MarkParticipationRefunded) gid=%s participation=%d: %v",
			gid, participationId, err)
	}
}

// newGid asks DTM for a globally-unique transaction id. Used as a correlation
// tag across branches (idempotency keys, log lines). DTM is best-effort here:
// if unreachable, we fall back to a local id.
func (c *Coordinator) newGid(ctx context.Context) (string, error) {
	if c.DTMHttpEndpoint == "" {
		return "", fmt.Errorf("DTMHttpEndpoint empty")
	}
	httpCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(httpCtx, http.MethodGet, c.DTMHttpEndpoint+"/newGid", nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("dtm newGid status=%d body=%s", resp.StatusCode, string(body))
	}
	var v struct {
		DtmResult string `json:"dtm_result"`
		Gid       string `json:"gid"`
	}
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&v); err != nil {
		return "", err
	}
	if v.Gid == "" {
		return "", fmt.Errorf("dtm newGid empty gid: %s", string(body))
	}
	return v.Gid, nil
}
