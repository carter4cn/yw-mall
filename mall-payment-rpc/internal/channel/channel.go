// Package channel defines the PayChannel abstraction so mall-payment-rpc can
// route a payment intent to the configured concrete provider (mock today,
// wechat/alipay/unionpay in S11).
package channel

import "context"

type PayReq struct {
	OrderID int64
	OrderNo string
	UserID  int64
	Amount  int64 // cents
}

type PayResp struct {
	Channel  string // mock / wechat / alipay
	PayURL   string // 微信 H5 url 或支付宝 wap form
	QRCode   string // 扫码场景
	PrepayID string // 微信 prepay_id
}

type RefundReq struct {
	OrderID int64
	OrderNo string
	Amount  int64
	Reason  string
}

type RefundResp struct {
	RefundNo string
	Status   string // success / pending / failed
}

type PayChannel interface {
	Name() string
	Pay(ctx context.Context, req *PayReq) (*PayResp, error)
	Refund(ctx context.Context, req *RefundReq) (*RefundResp, error)
}
