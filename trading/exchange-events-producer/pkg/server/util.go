package server

import (
	"math/rand"
	"time"

	"github.com/google/uuid"
)

func randomId() string {
	return uuid.New().String()
}

func randomClientId() string {
	return uuid.New().String()
}

func randomOrderId() string {
	return uuid.New().String()
}

func randomTradeId() string {
	return uuid.New().String()
}

func randomQty() int {
	return rand.Intn(1000) + 1
}

func randomPartialFilledQty() int {
	return rand.Intn(1000)
}

func randomPrice() float64 {
	return rand.Float64() * 1000
}

func randomPayIn() float64 {
	return rand.Float64() * 10000
}

func randomPayout() float64 {
	return rand.Float64() * 10000
}

func randomOpeningBalance() float64 {
	return rand.Float64() * 100000
}

func randomMarginUtilized() float64 {
	return rand.Float64() * 50000
}

func randomRemark() string {
	remarks := []string{
		"OK", "Pending", "Rejected", "Filled", "Partial", "Manual", "Auto", "Review",
	}
	return remarks[rand.Intn(len(remarks))]
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
