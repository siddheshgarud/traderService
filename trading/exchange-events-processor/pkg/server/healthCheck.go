package server

import (
	"encoding/json"
	"net/http"

	"exchange-events-processor/pkg/kafka/consumer"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func KafkaHealthHandler(w http.ResponseWriter, r *http.Request) {
	if consumer.GlobalConsumer == nil {
		http.Error(w, "consumer not initialized", http.StatusServiceUnavailable)
		return
	}
	h := consumer.GlobalConsumer.Health()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(h)
}
