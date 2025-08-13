package server

import (
	"context"
	"encoding/json"
	"exchange-events-processor/pkg/kafka/consumer"
	"exchange-events-processor/pkg/model"
	"log"
)

// OrderHandler processes incoming order messages.
func OrderHandler(ctx context.Context, message consumer.Message) error {
	var order model.Order
	if err := json.Unmarshal(message.Value, &order); err != nil {
		log.Printf("OrderHandler: failed to unmarshal order: %v", err)
		return err
	}
	log.Printf("Order Received: ClientID=%s OrderID=%s Qty=%.2f Price=%.2f Remark=%s",
		order.ClientID, order.OrderID, order.Qty, order.Price, order.Remark)
	return nil
}
