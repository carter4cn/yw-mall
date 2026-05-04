package logic

import (
	"context"
	"errors"

	"mall-logistics-rpc/internal/svc"
	"mall-logistics-rpc/logistics"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type CreateShipmentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateShipmentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateShipmentLogic {
	return &CreateShipmentLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *CreateShipmentLogic) CreateShipment(in *logistics.CreateShipmentReq) (*logistics.CreateShipmentResp, error) {
	if in.TrackingNo == "" || in.Carrier == "" {
		return nil, errors.New("logistics: tracking_no and carrier are required")
	}
	var newId int64
	err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		ret, err := session.ExecCtx(ctx,
			"INSERT INTO `shipment`(order_id, user_id, tracking_no, carrier, status, subscribe_status) VALUES (?,?,?,?,0,0)",
			in.OrderId, in.UserId, in.TrackingNo, in.Carrier)
		if err != nil {
			if isDuplicateKey(err) {
				return errors.New("logistics: tracking number already exists for this carrier")
			}
			return err
		}
		newId, _ = ret.LastInsertId()
		for _, it := range in.Items {
			if _, err := session.ExecCtx(ctx,
				"INSERT INTO `shipment_item`(shipment_id, order_item_id, product_id, quantity) VALUES (?,?,?,?)",
				newId, it.OrderItemId, it.ProductId, it.Quantity); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &logistics.CreateShipmentResp{ShipmentId: newId}, nil
}
