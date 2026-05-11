package logic

import (
	"context"

	"mall-shop-rpc/internal/svc"
	"mall-shop-rpc/shop"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListLevelApplicationsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListLevelApplicationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLevelApplicationsLogic {
	return &ListLevelApplicationsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListLevelApplicationsLogic) ListLevelApplications(in *shop.ListLevelApplicationsReq) (*shop.ListLevelApplicationsResp, error) {
	page, size := normPage(in.Page, in.PageSize)
	offset := (page - 1) * size

	where := "1=1"
	args := []any{}
	if in.Status >= 0 {
		where += " AND status=?"
		args = append(args, in.Status)
	}

	var total int64
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &total,
		"SELECT COUNT(*) FROM shop_level_application WHERE "+where, args...); err != nil {
		return nil, err
	}

	listArgs := append(args, size, offset)
	var rows []*levelApplicationRow
	if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows,
		"SELECT "+levelApplicationCols+" FROM shop_level_application WHERE "+where+" ORDER BY id DESC LIMIT ? OFFSET ?", listArgs...); err != nil {
		return nil, err
	}

	out := make([]*shop.ShopLevelApplication, 0, len(rows))
	for _, r := range rows {
		out = append(out, toLevelApplicationProto(r))
	}
	return &shop.ListLevelApplicationsResp{Applications: out, Total: total}, nil
}
