package logic

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"mall-order-rpc/internal/svc"
	"mall-order-rpc/order"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// orderRowForRefund is the slim projection of `order` used during refund submit.
// go-zero sqlx maps columns by `db:"..."` tag; defining the type at package
// level (rather than function-local) avoids "not matching destination to scan".
type orderRowForRefund struct {
	Id          int64  `db:"id"`
	OrderNo     string `db:"order_no"`
	UserId      int64  `db:"user_id"`
	TotalAmount int64  `db:"total_amount"`
	Status      int64  `db:"status"`
	ShopId      int64  `db:"shop_id"`
}

// refundRow is the wire shape of a refund_request row used by every read query.
type refundRow struct {
	Id                     int64  `db:"id"`
	OrderId                int64  `db:"order_id"`
	OrderNo                string `db:"order_no"`
	UserId                 int64  `db:"user_id"`
	ShopId                 int64  `db:"shop_id"`
	Amount                 int64  `db:"amount"`
	Reason                 string `db:"reason"`
	Evidence               string `db:"evidence"`
	Items                  string `db:"items"`
	Status                 int64  `db:"status"`
	MerchantUserId         int64  `db:"merchant_user_id"`
	MerchantRemark         string `db:"merchant_remark"`
	MerchantHandleTime     int64  `db:"merchant_handle_time"`
	AdminId                int64  `db:"admin_id"`
	AdminRemark            string `db:"admin_remark"`
	AdminHandleTime        int64  `db:"admin_handle_time"`
	AppealReason           string `db:"appeal_reason"`
	AppealTime             int64  `db:"appeal_time"`
	RefundNo               string `db:"refund_no"`
	RefundCompleteTime     int64  `db:"refund_complete_time"`
	CreateTime             int64  `db:"create_time"`
	RefundType             int64  `db:"refund_type"`
	ReturnTrackingNo       string `db:"return_tracking_no"`
	ReturnCarrier          string `db:"return_carrier"`
	ReturnShipTime         int64  `db:"return_ship_time"`
	ReturnReceivedTime     int64  `db:"return_received_time"`
	ReturnInspectionPassed int64  `db:"return_inspection_passed"`
	ExchangeNewOrderId     int64  `db:"exchange_new_order_id"`
}

// refundColumns lists all SELECT columns matching refundRow (kept in one place
// so adding a column = bumping one line).
const refundColumns = "id, order_id, order_no, user_id, shop_id, amount, reason, evidence, items, status, merchant_user_id, merchant_remark, merchant_handle_time, admin_id, admin_remark, admin_handle_time, appeal_reason, appeal_time, refund_no, refund_complete_time, create_time, refund_type, return_tracking_no, return_carrier, return_ship_time, return_received_time, return_inspection_passed, exchange_new_order_id"

func toRefundProto(r *refundRow) *order.RefundRequest {
	out := &order.RefundRequest{
		Id:                     r.Id,
		OrderId:                r.OrderId,
		OrderNo:                r.OrderNo,
		UserId:                 r.UserId,
		ShopId:                 r.ShopId,
		Amount:                 r.Amount,
		Reason:                 r.Reason,
		Status:                 int32(r.Status),
		MerchantUserId:         r.MerchantUserId,
		MerchantRemark:         r.MerchantRemark,
		MerchantHandleTime:     r.MerchantHandleTime,
		AdminId:                r.AdminId,
		AdminRemark:            r.AdminRemark,
		AdminHandleTime:        r.AdminHandleTime,
		AppealReason:           r.AppealReason,
		AppealTime:             r.AppealTime,
		RefundNo:               r.RefundNo,
		RefundCompleteTime:     r.RefundCompleteTime,
		CreateTime:             r.CreateTime,
		RefundType:             int32(r.RefundType),
		ReturnTrackingNo:       r.ReturnTrackingNo,
		ReturnCarrier:          r.ReturnCarrier,
		ReturnShipTime:         r.ReturnShipTime,
		ReturnReceivedTime:     r.ReturnReceivedTime,
		ReturnInspectionPassed: int32(r.ReturnInspectionPassed),
		ExchangeNewOrderId:     r.ExchangeNewOrderId,
	}
	if r.Evidence != "" {
		_ = json.Unmarshal([]byte(r.Evidence), &out.Evidence)
	}
	if r.Items != "" {
		var raw []struct {
			SkuId    int64  `json:"skuId"`
			SkuName  string `json:"skuName"`
			Quantity int32  `json:"quantity"`
			Amount   int64  `json:"amount"`
		}
		if err := json.Unmarshal([]byte(r.Items), &raw); err == nil {
			for _, it := range raw {
				out.Items = append(out.Items, &order.RefundItem{
					SkuId:    it.SkuId,
					SkuName:  it.SkuName,
					Quantity: it.Quantity,
					Amount:   it.Amount,
				})
			}
		}
	}
	return out
}

// refundTypeOrDefault returns the request's refund_type, falling back to 1
// (仅退款 refund_only) when callers send 0 / unset.
func refundTypeOrDefault(t int32) int32 {
	if t == 0 {
		return 1
	}
	return t
}

// generateRefundNo returns a globally-unique refund identifier of the form
// `RF<unix><4-digit-random>`, suitable for both DB persistence and external
// channel calls.
func generateRefundNo() string {
	n, err := rand.Int(rand.Reader, big.NewInt(10000))
	suffix := int64(0)
	if err == nil {
		suffix = n.Int64()
	}
	return fmt.Sprintf("RF%d%04d", time.Now().Unix(), suffix)
}

// loadRefundForUpdate selects a refund row inside a transaction with row lock,
// constraining the caller's status assumption.
func loadRefundForUpdate(ctx context.Context, tx sqlx.Session, refundId int64, expectStatus int) (*refundRow, error) {
	var r refundRow
	q := "SELECT " + refundColumns + " FROM refund_request WHERE id = ? FOR UPDATE"
	if err := tx.QueryRowCtx(ctx, &r, q, refundId); err != nil {
		return nil, err
	}
	if int(r.Status) != expectStatus {
		return nil, fmt.Errorf("refund status mismatch: got %d want %d", r.Status, expectStatus)
	}
	return &r, nil
}

// loadRefundById is a non-locking read helper.
func loadRefundById(ctx context.Context, svcCtx *svc.ServiceContext, id int64) (*refundRow, error) {
	var r refundRow
	q := "SELECT " + refundColumns + " FROM refund_request WHERE id = ? LIMIT 1"
	if err := svcCtx.SqlConn.QueryRowCtx(ctx, &r, q, id); err != nil {
		return nil, err
	}
	return &r, nil
}
