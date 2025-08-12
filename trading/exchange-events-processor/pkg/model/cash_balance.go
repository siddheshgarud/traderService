package model

type CashBalance struct {
	ClientID       string  `json:"client_id"`
	PayIn          float64 `json:"pay_in"`
	Payout         float64 `json:"payout"`
	OpeningBalance float64 `json:"opening_balance"`
	MarginUtilized float64 `json:"margin_utilized"`
}
