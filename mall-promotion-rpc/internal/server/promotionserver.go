// Promotion gRPC server. 嵌入 UnimplementedPromotionServer 让未实现的 RPC
// 自动返 codes.Unimplemented；本文件按 story 增量实现具体 method。
package server

import (
	"context"

	"mall-promotion-rpc/internal/logic"
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

// ===== S1.2 活动管理 =====

func (s *PromotionServer) CreateActivity(ctx context.Context, in *promotion.CreateActivityReq) (*promotion.CreateActivityResp, error) {
	return logic.CreateActivity(ctx, s.svcCtx, in)
}

func (s *PromotionServer) GetActivity(ctx context.Context, in *promotion.GetActivityReq) (*promotion.GetActivityResp, error) {
	return logic.GetActivity(ctx, s.svcCtx, in)
}

func (s *PromotionServer) UpdateActivity(ctx context.Context, in *promotion.UpdateActivityReq) (*promotion.OkResp, error) {
	return logic.UpdateActivity(ctx, s.svcCtx, in)
}

func (s *PromotionServer) ChangeActivityStatus(ctx context.Context, in *promotion.ChangeActivityStatusReq) (*promotion.OkResp, error) {
	return logic.ChangeActivityStatus(ctx, s.svcCtx, in)
}

func (s *PromotionServer) ListActivities(ctx context.Context, in *promotion.ListActivitiesReq) (*promotion.ListActivitiesResp, error) {
	return logic.ListActivities(ctx, s.svcCtx, in)
}
