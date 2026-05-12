// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"sync"

	"mall-api/internal/svc"
	"mall-api/internal/types"
	logisticspb "mall-logistics-rpc/logistics"
	orderpb "mall-order-rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type OrderDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrderDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderDetailLogic {
	return &OrderDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OrderDetailLogic) OrderDetail(req *types.OrderDetailReq) (*types.OrderDetailResp, error) {
	var (
		ord       *orderpb.GetOrderResp
		shipments *logisticspb.ListShipmentsByOrderResp
		ordErr    error
		wg        sync.WaitGroup
	)

	wg.Add(2)
	go func() {
		defer wg.Done()
		ord, ordErr = l.svcCtx.OrderRpc.GetOrder(l.ctx, &orderpb.GetOrderReq{Id: req.Id})
	}()
	go func() {
		defer wg.Done()
		shipments, _ = l.svcCtx.LogisticsRpc.ListShipmentsByOrder(l.ctx, &logisticspb.ListShipmentsByOrderReq{OrderId: req.Id})
	}()
	wg.Wait()

	if ordErr != nil {
		return nil, ordErr
	}

	items := make([]types.CreateOrderItem, 0, len(ord.Items))
	for _, item := range ord.Items {
		items = append(items, types.CreateOrderItem{
			ProductId:   item.ProductId,
			ProductName: item.ProductName,
			Price:       item.Price,
			Quantity:    item.Quantity,
		})
	}

	resp := &types.OrderDetailResp{
		Id:           ord.Id,
		OrderNo:      ord.OrderNo,
		UserId:       ord.UserId,
		TotalAmount:  ord.TotalAmount,
		Status:       ord.Status,
		Items:        items,
		CreateTime:   ord.CreateTime,
		PayTime:      ord.PayTime,
		ShipTime:     ord.ShipTime,
		CompleteTime: ord.CompleteTime,
		CancelTime:   ord.CancelTime,
		CancelReason: ord.CancelReason,
	}

	if shipments != nil {
		for _, s := range shipments.Shipments {
			resp.Shipments = append(resp.Shipments, protoShipmentToType(s))
		}
	}

	return resp, nil
}
