package handlers

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	TypeSigninAward    = "signin:award"
	TypeLotterySpin    = "lottery:spin"
	TypeLotteryNotify  = "lottery:notify"
	TypeSeckillPersist = "seckill:persist"
	TypeCouponNotify   = "coupon:notify"
)

// Register wires task-type → handler mappings to an asynq.ServeMux.
// Concrete logic for each handler lives in mall-workflow-rpc; this binary
// is the runtime that dequeues tasks. The handlers below are stubs that
// log the payload — Phase P2/P3 will replace them with real workflow calls.
func Register(mux *asynq.ServeMux) {
	mux.HandleFunc(TypeSigninAward, stubHandler(TypeSigninAward))
	mux.HandleFunc(TypeLotterySpin, stubHandler(TypeLotterySpin))
	mux.HandleFunc(TypeLotteryNotify, stubHandler(TypeLotteryNotify))
	mux.HandleFunc(TypeSeckillPersist, stubHandler(TypeSeckillPersist))
	mux.HandleFunc(TypeCouponNotify, stubHandler(TypeCouponNotify))
}

func stubHandler(taskType string) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var payload map[string]any
		_ = json.Unmarshal(t.Payload(), &payload)
		logx.WithContext(ctx).Infow("asynq task received (stub)",
			logx.Field("type", taskType),
			logx.Field("payload", payload),
		)
		return nil
	}
}
