package server

import (
	"context"
	"encoding/json"
	"exchange-events-processor/pkg/kafka/consumer"
	"exchange-events-processor/pkg/model"
	"fmt"
)

func OrderHandler(ctx context.Context, message consumer.Message) error {
	var order model.Order
	if err := json.Unmarshal(message.Value, &order); err != nil {
		fmt.Printf("Failed to unmarshal order: %v\n", err)
		return err
	}
	fmt.Printf("Order Received: ClientID=%s OrderID=%s Qty=%.2f Price=%.2f Remark=%s\n",
		order.ClientID, order.OrderID, order.Qty, order.Price, order.Remark)
	return nil
}
