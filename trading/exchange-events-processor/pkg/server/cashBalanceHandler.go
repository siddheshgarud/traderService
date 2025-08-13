package server

import (
	"context"
	"encoding/json"
	"exchange-events-processor/pkg/kafka/consumer"
	"exchange-events-processor/pkg/model"
	"log"
)

// CashBalanceHandler processes incoming cash balance messages.
func CashBalanceHandler(ctx context.Context, message consumer.Message) error {
	var cb model.CashBalance
	if err := json.Unmarshal(message.Value, &cb); err != nil {
		log.Printf("CashBalanceHandler: failed to unmarshal cash balance: %v", err)
		return err
	}
	log.Printf("CashBalance Received: ClientID=%s PayIn=%.2f Payout=%.2f OpeningBalance=%.2f MarginUtilized=%.2f",
		cb.ClientID, cb.PayIn, cb.Payout, cb.OpeningBalance, cb.MarginUtilized)
	return nil
}
