package logic

import (
	"context"
	"fmt"

	"mall-reward-rpc/internal/svc"
)

// bustRewardRecordCache evicts the goctl-cached entries for a reward_record id
// after a status mutation, so the next FindOne / FindOneByIdempotencyKey reads
// fresh data instead of serving the now-stale row.
//
// The id-key is straightforward; the idempotency-key index requires us to look
// up the row to find the key string. We skip the lookup when it's not in cache.
func bustRewardRecordCache(ctx context.Context, sc *svc.ServiceContext, recordId int64) {
	idKey := fmt.Sprintf("cache:rewardRecord:id:%d", recordId)
	_, _ = sc.Redis.DelCtx(ctx, idKey)

	var idem string
	if err := sc.DB.QueryRowCtx(ctx, &idem,
		"SELECT idempotency_key FROM `reward_record` WHERE id=? LIMIT 1 FOR UPDATE", recordId); err == nil && idem != "" {
		_, _ = sc.Redis.DelCtx(ctx, "cache:rewardRecord:idempotencyKey:"+idem)
	}
}
