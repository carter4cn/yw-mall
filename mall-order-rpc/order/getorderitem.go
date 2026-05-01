package order

import "fmt"

// GetOrderItemReq is the request message for GetOrderItem.
type GetOrderItemReq struct {
	OrderItemId int64 `protobuf:"varint,1,opt,name=order_item_id,json=orderItemId,proto3" json:"order_item_id,omitempty"`
}

func (x *GetOrderItemReq) Reset()         {}
func (x *GetOrderItemReq) String() string  { return fmt.Sprintf("order_item_id:%d", x.OrderItemId) }
func (*GetOrderItemReq) ProtoMessage()     {}

// GetOrderItemResp is the response message for GetOrderItem.
type GetOrderItemResp struct {
	OrderItemId int64 `protobuf:"varint,1,opt,name=order_item_id,json=orderItemId,proto3" json:"order_item_id,omitempty"`
	OrderId     int64 `protobuf:"varint,2,opt,name=order_id,json=orderId,proto3" json:"order_id,omitempty"`
	UserId      int64 `protobuf:"varint,3,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	ProductId   int64 `protobuf:"varint,4,opt,name=product_id,json=productId,proto3" json:"product_id,omitempty"`
	Quantity    int64 `protobuf:"varint,5,opt,name=quantity,proto3" json:"quantity,omitempty"`
	OrderStatus int32 `protobuf:"varint,6,opt,name=order_status,json=orderStatus,proto3" json:"order_status,omitempty"`
	CreateTime  int64 `protobuf:"varint,7,opt,name=create_time,json=createTime,proto3" json:"create_time,omitempty"`
}

func (x *GetOrderItemResp) Reset()        {}
func (x *GetOrderItemResp) String() string { return fmt.Sprintf("order_item_id:%d", x.OrderItemId) }
func (*GetOrderItemResp) ProtoMessage()    {}
