package channel

import (
	"context"
	"fmt"
	"time"
)

// MockChannel is the always-available dev channel. It does not call any real
// payment provider; Pay returns a confirm URL the C-side will POST back to.
type MockChannel struct{}

func (m *MockChannel) Name() string { return "mock" }

func (m *MockChannel) Pay(_ context.Context, req *PayReq) (*PayResp, error) {
	return &PayResp{
		Channel: "mock",
		PayURL:  fmt.Sprintf("/api/payment/mock-confirm/%d", req.OrderID),
	}, nil
}

func (m *MockChannel) Refund(_ context.Context, req *RefundReq) (*RefundResp, error) {
	return &RefundResp{
		RefundNo: fmt.Sprintf("MOCK_REFUND_%d_%d", req.OrderID, time.Now().Unix()),
		Status:   "success",
	}, nil
}
