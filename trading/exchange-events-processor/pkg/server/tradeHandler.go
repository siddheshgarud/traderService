package server

import (
	"context"
	"encoding/json"
	"exchange-events-processor/pkg/kafka/consumer"
	"exchange-events-processor/pkg/model"
	"fmt"
)

func TradeHandler(ctx context.Context, message consumer.Message) error {
	var trade model.Trade
	if err := json.Unmarshal(message.Value, &trade); err != nil {
		fmt.Printf("Failed to unmarshal trade: %v\n", err)
		return err
	}
	fmt.Printf("Trade Received: ClientID=%s OrderID=%s Qty=%.2f Price=%.2f Remark=%s TradeID=%s PartialFilledQty=%.2f\n",
		trade.ClientID, trade.OrderID, trade.Qty, trade.Price, trade.Remark, trade.TradeID, trade.PartialFilledQty)
	return nil
}
