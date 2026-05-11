package logic

import (
	"context"

	"mall-payment-rpc/internal/svc"
	"mall-payment-rpc/payment"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListBillRecordsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListBillRecordsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListBillRecordsLogic {
	return &ListBillRecordsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

type billRow struct {
	Id         int64  `db:"id"`
	ShopId     int64  `db:"shop_id"`
	Type       string `db:"type"`
	Amount     int64  `db:"amount"`
	OrderId    int64  `db:"order_id"`
	Remark     string `db:"remark"`
	CreateTime int64  `db:"create_time"`
}

func (l *ListBillRecordsLogic) ListBillRecords(in *payment.ListBillRecordsReq) (*payment.ListBillRecordsResp, error) {
	var total int64
	if err := l.svcCtx.SqlConn.QueryRowCtx(l.ctx, &total,
		"SELECT COUNT(*) FROM bill_record WHERE shop_id = ?", in.ShopId); err != nil {
		return nil, err
	}

	page, pageSize := in.Page, in.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	} else if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	rows := []*billRow{}
	if err := l.svcCtx.SqlConn.QueryRowsCtx(l.ctx, &rows,
		"SELECT id, shop_id, type, amount, order_id, remark, create_time FROM bill_record WHERE shop_id = ? ORDER BY create_time DESC, id DESC LIMIT ? OFFSET ?",
		in.ShopId, pageSize, offset); err != nil {
		return nil, err
	}

	out := make([]*payment.BillRecord, 0, len(rows))
	for _, r := range rows {
		out = append(out, &payment.BillRecord{
			Id:         r.Id,
			ShopId:     r.ShopId,
			Type:       r.Type,
			Amount:     r.Amount,
			OrderId:    r.OrderId,
			Remark:     r.Remark,
			CreateTime: r.CreateTime,
		})
	}
	return &payment.ListBillRecordsResp{Records: out, Total: total}, nil
}
