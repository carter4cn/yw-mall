package worker

import (
	"context"
	"encoding/json"
	"time"

	"mall-logistics-rpc/internal/logic"
	"mall-logistics-rpc/internal/svc"
	"mall-logistics-rpc/logistics"

	kgo "github.com/segmentio/kafka-go"
	"github.com/zeromicro/go-zero/core/logx"
)

// orderShippedEvent must match the JSON envelope written by mall-order-rpc MarkShipped.
type orderShippedItem struct {
	OrderItemId int64 `json:"orderItemId"`
	ProductId   int64 `json:"productId"`
	Quantity    int32 `json:"quantity"`
}
type orderShippedEvent struct {
	OrderId    int64              `json:"orderId"`
	UserId     int64              `json:"userId"`
	TrackingNo string             `json:"trackingNo"`
	Carrier    string             `json:"carrier"`
	Items      []orderShippedItem `json:"items"`
	ShippedAt  int64              `json:"shippedAt"`
}

type OrderShippedWorker struct {
	svcCtx *svc.ServiceContext
	reader *kgo.Reader
}

func NewOrderShippedWorker(svcCtx *svc.ServiceContext) *OrderShippedWorker {
	return &OrderShippedWorker{
		svcCtx: svcCtx,
		reader: kgo.NewReader(kgo.ReaderConfig{
			Brokers:        svcCtx.Config.Kafka.Brokers,
			Topic:          svcCtx.Config.Kafka.Topic,
			GroupID:        svcCtx.Config.Kafka.Group,
			MinBytes:       1,
			MaxBytes:       10e6,
			CommitInterval: time.Second,
		}),
	}
}

// Start launches the consumer loop in a new goroutine. It returns immediately.
func (w *OrderShippedWorker) Start(ctx context.Context) {
	go func() {
		defer w.reader.Close()
		for {
			m, err := w.reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				logx.Errorf("logistics worker: kafka read failed: %v", err)
				time.Sleep(time.Second)
				continue
			}
			if err := w.handle(ctx, m.Value); err != nil {
				logx.Errorf("logistics worker: handle order.shipped failed: %v", err)
			}
		}
	}()
}

func (w *OrderShippedWorker) handle(ctx context.Context, payload []byte) error {
	var ev orderShippedEvent
	if err := json.Unmarshal(payload, &ev); err != nil {
		return err
	}
	items := make([]*logistics.ShipmentItemRef, 0, len(ev.Items))
	for _, it := range ev.Items {
		items = append(items, &logistics.ShipmentItemRef{
			OrderItemId: it.OrderItemId,
			ProductId:   it.ProductId,
			Quantity:    it.Quantity,
		})
	}
	resp, err := logic.NewCreateShipmentLogic(ctx, w.svcCtx).CreateShipment(&logistics.CreateShipmentReq{
		OrderId:    ev.OrderId,
		UserId:     ev.UserId,
		TrackingNo: ev.TrackingNo,
		Carrier:    ev.Carrier,
		Items:      items,
	})
	if err != nil {
		return err
	}
	w.subscribeWithRetry(ctx, resp.ShipmentId, ev.Carrier, ev.TrackingNo)
	return nil
}

func (w *OrderShippedWorker) subscribeWithRetry(ctx context.Context, shipmentId int64, carrier, trackingNo string) {
	max := w.svcCtx.Config.Subscribe.MaxRetries
	if max <= 0 {
		max = 3
	}
	backoff := time.Duration(w.svcCtx.Config.Subscribe.InitialBackoffMs) * time.Millisecond
	if backoff <= 0 {
		backoff = time.Second
	}
	var lastErr error
	for i := 0; i < max; i++ {
		if err := w.svcCtx.Kuaidi100.Subscribe(ctx, carrier, trackingNo); err == nil {
			_, _ = w.svcCtx.DB.ExecCtx(ctx,
				"UPDATE shipment SET subscribe_status=1 WHERE id=?", shipmentId)
			return
		} else {
			lastErr = err
		}
		time.Sleep(backoff)
		backoff *= 2
	}
	_, _ = w.svcCtx.DB.ExecCtx(ctx,
		"UPDATE shipment SET subscribe_status=2 WHERE id=?", shipmentId)
	_, _ = w.svcCtx.DB.ExecCtx(ctx,
		"INSERT INTO shipment_track(shipment_id, track_time, location, description, state_kuaidi100, state_internal) VALUES (?, NOW(), '', ?, 255, 5)",
		shipmentId, "subscribe_failed: "+lastErr.Error())
	logx.Errorf("logistics worker: subscribe failed after %d retries: %v", max, lastErr)
}
