package logic

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	mathrand "math/rand"
	"strconv"
	"sync"
	"time"

	"mall-activity-rpc/activity"
	"mall-activity-rpc/internal/lua"
	"mall-activity-rpc/internal/model"
	"mall-activity-rpc/internal/saga"
	"mall-activity-rpc/internal/svc"
	"mall-reward-rpc/reward"
	"mall-risk-rpc/risk"
	"mall-rule-rpc/rule"
	"mall-workflow-rpc/workflow"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc"
)

type ParticipateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewParticipateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ParticipateLogic {
	return &ParticipateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Participate is the universal entry point for all 4 activity types. It:
//  1. loads the activity, asserts it is PUBLISHED
//  2. evaluates rule_set (if configured) for eligibility
//  3. dispatches to per-type handler (signin/lottery/seckill/coupon)
//  4. starts a workflow_instance and fires the per-type initial trigger
//  5. writes a participation_record (idempotent via idempotency_key)
//
// Returns ParticipateResp with the workflow_instance_id and a status hint.
func (l *ParticipateLogic) Participate(in *activity.ParticipateReq) (*activity.ParticipateResp, error) {
	a, err := l.svcCtx.ActivityModel.FindOne(l.ctx, uint64(in.ActivityId))
	if err != nil {
		return nil, fmt.Errorf("activity %d: %w", in.ActivityId, err)
	}
	if a.Status != "PUBLISHED" {
		return nil, fmt.Errorf("activity status=%s, not participable", a.Status)
	}

	// idempotency_key = sha1(user_id|activity_id|client_request_id_or_minute)
	idemKey := makeIdempotencyKey(in)
	if existing, err := l.svcCtx.ParticipationRecordModel.FindOneByIdempotencyKey(l.ctx, idemKey); err == nil && existing != nil {
		// already participated under this key → return idempotent result without
		// re-charging risk counters (the original request already paid that cost).
		return &activity.ParticipateResp{
			ParticipationId:    int64(existing.Id),
			WorkflowInstanceId: existing.WorkflowInstanceId,
			Status:             existing.Status,
			DetailJson:         existing.PayloadJson.String,
		}, nil
	}

	// Risk checks run before any side-effect (Lua decrements, DB writes) so a
	// rejected request is cheap. Signin is exempt — it's a low-value daily flow
	// where rate limiting would just create false-positive support tickets.
	if a.Type != "signin" {
		if err := l.applyRiskChecks(in, a.Type); err != nil {
			return nil, err
		}
	}

	// rule eligibility check
	if a.RuleSetId > 0 {
		ruleCtx := l.buildRuleContext(int64(a.Id), a.Type, in.UserId)
		ev, err := l.svcCtx.RuleRpc.EvaluateRuleSet(l.ctx, &rule.EvaluateRuleSetReq{
			RuleSetId: a.RuleSetId,
			Context:   ruleCtx,
		})
		if err != nil {
			return nil, fmt.Errorf("rule eval: %w", err)
		}
		if !ev.Result {
			return nil, fmt.Errorf("not eligible: rule %d failed", ev.FirstFailedRuleId)
		}
	}

	// per-type dispatch returns initial trigger + payload + status hint
	trigger, payload, statusHint, err := l.dispatch(a, in)
	if err != nil {
		return nil, err
	}
	payloadBytes, _ := json.Marshal(payload)

	// start workflow instance
	wfStart, err := l.svcCtx.WorkflowRpc.StartInstance(l.ctx, &workflow.StartInstanceReq{
		DefinitionId: a.WorkflowDefinitionId,
		ActivityId:   int64(a.Id),
		UserId:       in.UserId,
		PayloadJson:  string(payloadBytes),
	})
	if err != nil {
		return nil, fmt.Errorf("workflow.StartInstance: %w", err)
	}

	// fire the per-type initial trigger (e.g. check_in, claim, spin, buy)
	if trigger != "" {
		if _, err := l.svcCtx.WorkflowRpc.Fire(l.ctx, &workflow.FireReq{
			InstanceId: wfStart.InstanceId,
			Trigger:    trigger,
		}); err != nil {
			return nil, fmt.Errorf("workflow.Fire(%s): %w", trigger, err)
		}
	}

	// persist participation_record
	res, err := l.svcCtx.ParticipationRecordModel.Insert(l.ctx, &model.ParticipationRecord{
		ActivityId:         int64(a.Id),
		UserId:             in.UserId,
		Sequence:           1,
		WorkflowInstanceId: wfStart.InstanceId,
		Status:             statusHint,
		PayloadJson:        sql.NullString{String: string(payloadBytes), Valid: true},
		IdempotencyKey:     idemKey,
	})
	if err != nil {
		return nil, fmt.Errorf("insert participation: %w", err)
	}
	pid, _ := res.LastInsertId()

	// Reward dispatch.
	//   • lottery WON: runs the 3-branch SAGA (Dispatch → MarkRewarded → Confirm)
	//     so a downstream failure is rolled back atomically.
	//   • signin / coupon: single-step Dispatch is sufficient (no cross-record
	//     invariant to maintain), so we skip the SAGA's overhead.
	//   • seckill: reward dispatch deferred to its own SAGA in workflow-rpc once
	//     order pre-creation lands.
	if rewardable(a.Type, statusHint) {
		switch a.Type {
		case "lottery":
			l.runLotterySaga(int64(a.Id), in.UserId, wfStart.InstanceId, pid, payload, payloadBytes)
		default:
			l.dispatchReward(a.Type, int64(a.Id), in.UserId, wfStart.InstanceId, payload)
		}
	}

	return &activity.ParticipateResp{
		ParticipationId:    pid,
		WorkflowInstanceId: wfStart.InstanceId,
		Status:             statusHint,
		DetailJson:         string(payloadBytes),
	}, nil
}

// rewardable decides whether this terminal-status participation should fire a
// reward dispatch directly. Lottery losers and rejected/seckill paths skip.
func rewardable(activityType, status string) bool {
	switch activityType {
	case "signin":
		return status == "CHECKED_IN"
	case "lottery":
		return status == "WON"
	case "coupon":
		return status == "ISSUED"
	}
	return false
}

// dispatchReward calls reward-rpc.Dispatch keyed by workflow_instance_id so
// retries collapse to the same reward_record. Failures are logged but do not
// fail the participate call — the activity already happened, the reward will
// be reconciled by the outbox relay or replayed manually.
func (l *ParticipateLogic) dispatchReward(activityType string, activityId, userId, workflowInstanceId int64, payload map[string]any) {
	tplCode := l.templateCodeFor(activityType)
	if tplCode == "" {
		l.Logger.Errorf("no reward template configured for activity type %q", activityType)
		return
	}
	templateId, err := l.resolveTemplateId(tplCode)
	if err != nil {
		l.Logger.Errorf("reward template lookup %q: %v", tplCode, err)
		return
	}
	body, _ := json.Marshal(payload)
	if _, err := l.svcCtx.RewardRpc.Dispatch(l.ctx, &reward.DispatchReq{
		UserId:             userId,
		ActivityId:         activityId,
		WorkflowInstanceId: workflowInstanceId,
		TemplateId:         templateId,
		PayloadJson:        string(body),
	}); err != nil {
		l.Logger.Errorf("reward dispatch (workflow=%d type=%s): %v", workflowInstanceId, activityType, err)
	}
}

// templateIdCache memoises code->id resolutions across requests. The map is
// process-local; re-seeding reward templates produces new ids, which means a
// service restart is the simplest invalidation path. That trade-off is fine
// because templates change rarely.
var templateIdCache sync.Map

func (l *ParticipateLogic) resolveTemplateId(code string) (int64, error) {
	if v, ok := templateIdCache.Load(code); ok {
		return v.(int64), nil
	}
	res, err := l.svcCtx.RewardRpc.ListTemplates(l.ctx, &reward.ListTemplatesReq{PageSize: 200})
	if err != nil {
		return 0, err
	}
	for _, t := range res.Templates {
		templateIdCache.Store(t.Code, t.Id)
	}
	if v, ok := templateIdCache.Load(code); ok {
		return v.(int64), nil
	}
	return 0, fmt.Errorf("template %q not found among %d templates", code, len(res.Templates))
}

// runLotterySaga drives the 3-branch SAGA. Errors are logged but do not
// fail the participate response — the participation itself already succeeded;
// the reward state will be visible in /api/activity/my/rewards once the SAGA
// resolves (CONFIRMED on success, REFUNDED on compensation).
func (l *ParticipateLogic) runLotterySaga(activityId, userId, workflowInstanceId, participationId int64, payload map[string]any, payloadBytes []byte) {
	tplCode := l.templateCodeFor("lottery")
	templateId, err := l.resolveTemplateId(tplCode)
	if err != nil {
		l.Logger.Errorf("saga: lottery template %q lookup: %v", tplCode, err)
		return
	}
	coord := &saga.Coordinator{
		DTMHttpEndpoint: l.svcCtx.Config.Dtm.Server,
		Reward:          sagaRewardAdapter{rpc: l.svcCtx.RewardRpc},
		Participation:   sagaParticipationAdapter{local: l},
	}
	res, err := coord.RunLotteryReward(l.ctx, saga.LotteryRewardInput{
		UserId:             userId,
		ActivityId:         activityId,
		WorkflowInstanceId: workflowInstanceId,
		ParticipationId:    participationId,
		TemplateId:         templateId,
		PayloadJson:        string(payloadBytes),
	})
	if err != nil {
		l.Logger.Errorf("saga lottery_reward gid=%v failed at %s: %v", res, res.FailedAt, err)
		return
	}
	l.Logger.Infof("saga lottery_reward gid=%s CONFIRMED reward_record=%d", res.Gid, res.RewardRecordId)
}

// sagaRewardAdapter bridges the saga.RewardClient interface (plain ctx args)
// to the goctl-generated rewardclient (which requires ...grpc.CallOption).
type sagaRewardAdapter struct {
	rpc interface {
		Dispatch(ctx context.Context, in *reward.DispatchReq, opts ...grpc.CallOption) (*reward.DispatchResp, error)
		Confirm(ctx context.Context, in *reward.ConfirmReq, opts ...grpc.CallOption) (*reward.Empty, error)
		RefundReward(ctx context.Context, in *reward.RefundRewardReq, opts ...grpc.CallOption) (*reward.Empty, error)
		MarkFailed(ctx context.Context, in *reward.MarkFailedReq, opts ...grpc.CallOption) (*reward.Empty, error)
	}
}

func (a sagaRewardAdapter) Dispatch(ctx context.Context, in *reward.DispatchReq) (*reward.DispatchResp, error) {
	return a.rpc.Dispatch(ctx, in)
}
func (a sagaRewardAdapter) Confirm(ctx context.Context, in *reward.ConfirmReq) (*reward.Empty, error) {
	return a.rpc.Confirm(ctx, in)
}
func (a sagaRewardAdapter) RefundReward(ctx context.Context, in *reward.RefundRewardReq) (*reward.Empty, error) {
	return a.rpc.RefundReward(ctx, in)
}
func (a sagaRewardAdapter) MarkFailed(ctx context.Context, in *reward.MarkFailedReq) (*reward.Empty, error) {
	return a.rpc.MarkFailed(ctx, in)
}

// sagaParticipationAdapter bridges the saga.ParticipationMarker interface to
// the in-process MarkParticipation*Logic implementations. Going through the
// gRPC client would also work, but in-process is faster and avoids a self-call
// loopback hop.
type sagaParticipationAdapter struct {
	local *ParticipateLogic
}

func (a sagaParticipationAdapter) MarkParticipationRewarded(ctx context.Context, in *activity.MarkParticipationRewardedReq) (*activity.Empty, error) {
	return NewMarkParticipationRewardedLogic(ctx, a.local.svcCtx).MarkParticipationRewarded(in)
}

func (a sagaParticipationAdapter) MarkParticipationRefunded(ctx context.Context, in *activity.MarkParticipationRefundedReq) (*activity.Empty, error) {
	return NewMarkParticipationRefundedLogic(ctx, a.local.svcCtx).MarkParticipationRefunded(in)
}

func (l *ParticipateLogic) templateCodeFor(activityType string) string {
	switch activityType {
	case "signin":
		return l.svcCtx.Config.RewardTemplates.Signin
	case "lottery":
		return l.svcCtx.Config.RewardTemplates.Lottery
	case "coupon":
		return l.svcCtx.Config.RewardTemplates.Coupon
	}
	return ""
}

// dispatch runs the per-type hot-path Lua atomic operation (if any) and
// returns the FSM trigger name + a JSON-friendly payload + status hint.
func (l *ParticipateLogic) dispatch(a *model.Activity, in *activity.ParticipateReq) (trigger string, payload map[string]any, status string, err error) {
	payload = map[string]any{
		"user_id":     in.UserId,
		"activity_id": int64(a.Id),
		"type":        a.Type,
	}
	switch a.Type {
	case "signin":
		return "check_in", payload, "CHECKED_IN", nil

	case "lottery":
		// pick a random number in [0, 100); the Lua script handles weighting
		rand := mathrand.Intn(100)
		key := fmt.Sprintf("activity:%d:prizes", a.Id)
		v, e := l.svcCtx.Redis.EvalCtx(l.ctx, lua.LotteryPick, []string{key}, strconv.Itoa(rand))
		if e != nil {
			return "", payload, "REJECTED", fmt.Errorf("lottery_pick: %w", e)
		}
		idx, _ := v.(int64)
		if idx < 0 {
			payload["prize_index"] = -1
			return "lose", payload, "LOST", nil
		}
		payload["prize_index"] = idx
		return "win", payload, "WON", nil

	case "seckill":
		var skuId int64 = 1
		var qty int64 = 1
		if v, ok := readInt(in.PayloadJson, "sku_id"); ok {
			skuId = v
		}
		if v, ok := readInt(in.PayloadJson, "quantity"); ok && v > 0 {
			qty = v
		}
		stockKey := fmt.Sprintf("activity:%d:stock", a.Id)
		dedupKey := fmt.Sprintf("activity:%d:dedup", a.Id)
		v, e := l.svcCtx.Redis.EvalCtx(l.ctx, lua.SeckillDecr,
			[]string{stockKey, dedupKey},
			strconv.FormatInt(skuId, 10),
			strconv.FormatInt(in.UserId, 10),
			strconv.FormatInt(qty, 10),
		)
		if e != nil {
			return "", payload, "REJECTED", fmt.Errorf("seckill_decr: %w", e)
		}
		left, _ := v.(int64)
		switch left {
		case -1:
			return "", payload, "REJECTED", fmt.Errorf("already participated")
		case -2:
			return "", payload, "REJECTED", fmt.Errorf("sku not found")
		case -3:
			return "", payload, "REJECTED", fmt.Errorf("stock empty")
		}
		payload["sku_id"] = skuId
		payload["quantity"] = qty
		payload["stock_left"] = left
		return "buy", payload, "RESERVED", nil

	case "coupon":
		var maxPerUser int64 = 1
		if a.ConfigJson.Valid {
			if v, ok := readInt(a.ConfigJson.String, "max_per_user"); ok && v > 0 {
				maxPerUser = v
			}
		}
		stockKey := fmt.Sprintf("activity:%d:coupon_stock", a.Id)
		userKey := fmt.Sprintf("activity:%d:user_claims", a.Id)
		v, e := l.svcCtx.Redis.EvalCtx(l.ctx, lua.CouponClaim,
			[]string{stockKey, userKey},
			strconv.FormatInt(in.UserId, 10),
			strconv.FormatInt(maxPerUser, 10),
		)
		if e != nil {
			return "", payload, "REJECTED", fmt.Errorf("coupon_claim: %w", e)
		}
		left, _ := v.(int64)
		switch left {
		case -1:
			return "", payload, "REJECTED", fmt.Errorf("per-user limit reached")
		case -2:
			return "", payload, "REJECTED", fmt.Errorf("stock empty")
		}
		payload["stock_left"] = left
		return "claim", payload, "ISSUED", nil
	}
	return "", payload, "REJECTED", fmt.Errorf("unsupported activity type %q", a.Type)
}

// applyRiskChecks runs blacklist + rate limit + (when present) token verify.
// Returns a non-nil error to short-circuit Participate. Errors are formatted
// to match the errorx codes in mall-common (7005/7006/7007); mall-api maps
// the message back to the right HTTP code.
func (l *ParticipateLogic) applyRiskChecks(in *activity.ParticipateReq, activityType string) error {
	uidStr := strconv.FormatInt(in.UserId, 10)

	// 1. Blacklist
	if bl, err := l.svcCtx.RiskRpc.CheckBlacklist(l.ctx, &risk.CheckBlacklistReq{
		SubjectType:  "user",
		SubjectValue: uidStr,
	}); err == nil && bl.Blacklisted {
		return fmt.Errorf("user blacklisted: %s", bl.Reason)
	}

	// 2. Rate limit (per user-per-activity)
	rl, err := l.svcCtx.RiskRpc.RateLimit(l.ctx, &risk.RateLimitReq{
		ActivityId:    in.ActivityId,
		SubjectType:   "user",
		SubjectValue:  uidStr,
		MaxCount:      10,
		WindowSeconds: 60,
	})
	if err == nil && !rl.Allowed {
		return fmt.Errorf("rate limit exceeded; reset_at=%d", rl.ResetAt)
	}

	// 3. Token verify (only when caller supplied one — getActivity issued it for
	// lottery/seckill). Empty token is allowed for now since the demo flow
	// doesn't push tokens through every path; tightening to "required for
	// lottery/seckill" is a one-line change once the UI plumbs it through.
	if in.Token != "" {
		v, err := l.svcCtx.RiskRpc.VerifyToken(l.ctx, &risk.VerifyTokenReq{
			UserId:     in.UserId,
			ActivityId: in.ActivityId,
			Token:      in.Token,
			Consume:    true,
		})
		if err != nil {
			return fmt.Errorf("token verify: %w", err)
		}
		if !v.Valid {
			return fmt.Errorf("invalid participation token: %s", v.Reason)
		}
	}
	return nil
}

func (l *ParticipateLogic) buildRuleContext(activityId int64, activityType string, userId int64) *rule.RuleContext {
	// lightweight context — full enrichment (tier/country/risk score/etc.)
	// arrives in P5 once mall-risk-rpc + user-rpc.GetUser fan-out is wired.
	now := time.Now().Unix()

	var todayCount int64
	dayStart := time.Now().Truncate(24 * time.Hour).Unix()
	_ = l.svcCtx.DB.QueryRowCtx(l.ctx, &todayCount,
		"SELECT COUNT(*) FROM `participation_record` WHERE activity_id=? AND user_id=? AND UNIX_TIMESTAMP(create_time) >= ?",
		activityId, userId, dayStart)

	var totalCount int64
	_ = l.svcCtx.DB.QueryRowCtx(l.ctx, &totalCount,
		"SELECT COUNT(*) FROM `participation_record` WHERE activity_id=? AND user_id=?",
		activityId, userId)

	return &rule.RuleContext{
		UserId:                  userId,
		ActivityId:              activityId,
		ActivityType:            activityType,
		CurrentTime:             now,
		ParticipationCountToday: todayCount,
		ParticipationCountTotal: totalCount,
		UserTier:                "regular",
	}
}

// makeIdempotencyKey builds a stable hash so duplicate requests within the
// same logical operation (typically same client_request_id) collapse to the
// same participation_record. When the caller does not supply a request id,
// we fall back to a per-minute bucket — this is intentionally coarse so a
// double-click doesn't create two records, while different minutes are
// considered fresh attempts.
func makeIdempotencyKey(in *activity.ParticipateReq) string {
	src := in.ClientRequestId
	if src == "" {
		src = fmt.Sprintf("u:%d|a:%d|m:%d", in.UserId, in.ActivityId, time.Now().Unix()/60)
	}
	h := sha1.Sum([]byte(src))
	return hex.EncodeToString(h[:])
}

func readInt(payloadJson, key string) (int64, bool) {
	if payloadJson == "" {
		return 0, false
	}
	m := map[string]any{}
	if err := json.Unmarshal([]byte(payloadJson), &m); err != nil {
		return 0, false
	}
	v, ok := m[key]
	if !ok {
		return 0, false
	}
	switch x := v.(type) {
	case float64:
		return int64(x), true
	case int64:
		return x, true
	case int:
		return int64(x), true
	case string:
		n, err := strconv.ParseInt(x, 10, 64)
		if err == nil {
			return n, true
		}
	}
	return 0, false
}
