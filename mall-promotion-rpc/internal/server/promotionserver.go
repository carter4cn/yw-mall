// Promotion gRPC server. 当前 Phase 1 骨架阶段：嵌入 UnimplementedPromotionServer
// 拿到所有 RPC 的"未实现"兜底，后续 story 逐个覆盖。
package server

import (
	"mall-promotion-rpc/internal/svc"
	"mall-promotion-rpc/promotion"
)

type PromotionServer struct {
	promotion.UnimplementedPromotionServer
	svcCtx *svc.ServiceContext
}

func NewPromotionServer(svcCtx *svc.ServiceContext) *PromotionServer {
	return &PromotionServer{svcCtx: svcCtx}
}
