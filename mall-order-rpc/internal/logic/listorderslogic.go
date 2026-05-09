package logic

import (
	"context"
	"fmt"

	"mall-order-rpc/internal/model"
	"mall-order-rpc/internal/svc"
	"mall-order-rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListOrdersLogic {
	return &ListOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListOrdersLogic) ListOrders(in *order.ListOrdersReq) (*order.ListOrdersResp, error) {
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	page := in.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize

	whereClause := "`user_id` = ?"
	args := []interface{}{in.UserId}

	// status -1 means all; otherwise filter by status
	if in.Status >= 0 {
		whereClause += " AND `status` = ?"
		args = append(args, in.Status)
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM `order` WHERE %s", whereClause)
	var total int64
	if err := l.svcCtx.SqlConn.QueryRowCtx(l.ctx, &total, countQuery, args...); err != nil {
		return nil, err
	}

	listQuery := fmt.Sprintf(
		"SELECT `id`, `order_no`, `user_id`, `total_amount`, `status`, `create_time`, `address_id`, `receiver_name`, `receiver_phone`, `receiver_province`, `receiver_city`, `receiver_district`, `receiver_detail` FROM `order` WHERE %s ORDER BY `id` DESC LIMIT ? OFFSET ?",
		whereClause,
	)
	listArgs := append(args, pageSize, offset)

	var rows []orderRow
	if err := l.svcCtx.SqlConn.QueryRowsCtx(l.ctx, &rows, listQuery, listArgs...); err != nil {
		return nil, err
	}

	pbOrders := make([]*order.GetOrderResp, 0, len(rows))
	for _, o := range rows {
		var items []*model.OrderItem
		if err := l.svcCtx.SqlConn.QueryRowsCtx(l.ctx, &items,
			"SELECT `id`, `order_id`, `product_id`, `product_name`, `price`, `quantity`, `create_time` FROM `order_item` WHERE `order_id` = ?",
			o.Id,
		); err != nil {
			return nil, err
		}

		pbItems := make([]*order.OrderItem, 0, len(items))
		for _, item := range items {
			pbItems = append(pbItems, &order.OrderItem{
				ProductId:   int64(item.ProductId),
				ProductName: item.ProductName,
				Price:       item.Price,
				Quantity:    int32(item.Quantity),
			})
		}

		pbOrders = append(pbOrders, &order.GetOrderResp{
			Id:               int64(o.Id),
			OrderNo:          o.OrderNo,
			UserId:           int64(o.UserId),
			TotalAmount:      o.TotalAmount,
			Status:           int32(o.Status),
			Items:            pbItems,
			CreateTime:       o.CreateTime.Unix(),
			AddressId:        o.AddressId,
			ReceiverName:     o.ReceiverName,
			ReceiverPhone:    o.ReceiverPhone,
			ReceiverProvince: o.ReceiverProvince,
			ReceiverCity:     o.ReceiverCity,
			ReceiverDistrict: o.ReceiverDistrict,
			ReceiverDetail:   o.ReceiverDetail,
		})
	}

	return &order.ListOrdersResp{
		Orders: pbOrders,
		Total:  total,
	}, nil
}
