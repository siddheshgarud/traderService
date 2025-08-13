package server

import (
	"context"
	"encoding/json"
	"exchange-events-processor/pkg/kafka/consumer"
	"exchange-events-processor/pkg/model"
	"log"
)

// TradeHandler processes incoming trade messages.
func TradeHandler(ctx context.Context, message consumer.Message) error {
	var trade model.Trade
	if err := json.Unmarshal(message.Value, &trade); err != nil {
		log.Printf("TradeHandler: failed to unmarshal trade: %v", err)
		return err
	}
	log.Printf("Trade Received: ClientID=%s OrderID=%s Qty=%.2f Price=%.2f Remark=%s TradeID=%s PartialFilledQty=%.2f",
		trade.ClientID, trade.OrderID, trade.Qty, trade.Price, trade.Remark, trade.TradeID, trade.PartialFilledQty)
	return nil
}
