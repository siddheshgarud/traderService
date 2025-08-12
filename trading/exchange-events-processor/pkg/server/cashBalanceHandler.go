package server

import (
	"context"
	"encoding/json"
	"exchange-events-processor/pkg/kafka/consumer"
	"exchange-events-processor/pkg/model"
	"fmt"
)

func CashBalanceHandler(ctx context.Context, message consumer.Message) error {
	var cb model.CashBalance
	if err := json.Unmarshal(message.Value, &cb); err != nil {
		fmt.Printf("Failed to unmarshal cash balance: %v\n", err)
		return err
	}
	fmt.Printf("CashBalance Received: ClientID=%s PayIn=%.2f Payout=%.2f OpeningBalance=%.2f MarginUtilized=%.2f\n",
		cb.ClientID, cb.PayIn, cb.Payout, cb.OpeningBalance, cb.MarginUtilized)
	return nil
}
