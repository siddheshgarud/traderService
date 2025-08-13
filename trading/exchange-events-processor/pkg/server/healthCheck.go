package server

import (
	"encoding/json"
	"net/http"

	"exchange-events-processor/pkg/kafka/consumer"
)

// HealthHandler returns a simple health status.
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// KafkaHealthHandler returns the health of the Kafka consumer.
func KafkaHealthHandler(w http.ResponseWriter, r *http.Request) {
	if consumer.GlobalConsumer == nil {
		http.Error(w, `{"error":"consumer not initialized"}`, http.StatusServiceUnavailable)
		return
	}
	h := consumer.GlobalConsumer.Health()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(h)
}
