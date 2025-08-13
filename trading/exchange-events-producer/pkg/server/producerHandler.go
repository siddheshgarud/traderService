package server

import (
	"context"
	"encoding/json"
	"exchange-events-producer/pkg/config"
	"exchange-events-producer/pkg/kafka/producer"
	"log"
	"net/http"
	"time"
)

type ProduceRandomRequest struct {
	OrderCount       int `json:"order_count"`
	CashBalanceCount int `json:"cash_balance_count"`
	TradesCount      int `json:"trades_count"`
}

type ProduceRandomResponse struct {
	OrderCount       int    `json:"order_count"`
	CashBalanceCount int    `json:"cash_balance_count"`
	TradesCount      int    `json:"trades_count"`
	Status           string `json:"status"`
}

func ProduceRandomHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req ProduceRandomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	var failed bool

	// Produce random orders
	for i := 0; i < req.OrderCount; i++ {
		order := map[string]interface{}{
			"client_id": randomClientId(),
			"order_id":  randomOrderId(),
			"qty":       randomQty(),
			"price":     randomPrice(),
			"remark":    randomRemark(),
		}
		val, _ := json.Marshal(order)
		msg := producer.ProducerMessage{
			Topic:     config.KafkaConfig.OrderTopic,
			Value:     val,
			Timestamp: time.Now(),
		}
		if err := producer.ProduceWithRetry(ctx, msg); err != nil {
			log.Printf("Failed to produce order message: %v", err)
			failed = true
		}
	}

	// Produce random cash balances
	for i := 0; i < req.CashBalanceCount; i++ {
		cash := map[string]interface{}{
			"client_id":       randomClientId(),
			"pay_in":          randomPayIn(),
			"payout":          randomPayout(),
			"opening_balance": randomOpeningBalance(),
			"margin_utilized": randomMarginUtilized(),
		}
		val, _ := json.Marshal(cash)
		msg := producer.ProducerMessage{
			Topic:     config.KafkaConfig.CashBalanceTopic,
			Value:     val,
			Timestamp: time.Now(),
		}
		if err := producer.ProduceWithRetry(ctx, msg); err != nil {
			log.Printf("Failed to produce cash balance message: %v", err)
			failed = true
		}
	}

	// Produce random trades
	for i := 0; i < req.TradesCount; i++ {
		trade := map[string]interface{}{
			"client_id":          randomClientId(),
			"order_id":           randomOrderId(),
			"qty":                randomQty(),
			"price":              randomPrice(),
			"remark":             randomRemark(),
			"trade_id":           randomTradeId(),
			"partial_filled_qty": randomPartialFilledQty(),
		}
		val, _ := json.Marshal(trade)
		msg := producer.ProducerMessage{
			Topic:     config.KafkaConfig.TradesTopic,
			Value:     val,
			Timestamp: time.Now(),
		}
		if err := producer.ProduceWithRetry(ctx, msg); err != nil {
			log.Printf("Failed to produce trade message: %v", err)
			failed = true
		}
	}

	resp := ProduceRandomResponse{
		OrderCount:       req.OrderCount,
		CashBalanceCount: req.CashBalanceCount,
		TradesCount:      req.TradesCount,
		Status:           "success",
	}
	if failed {
		resp.Status = "partial_failure"
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
