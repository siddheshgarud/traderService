package model

type Order struct {
	ClientID string  `json:"client_id"`
	OrderID  string  `json:"order_id"`
	Qty      float64 `json:"qty"`
	Price    float64 `json:"price"`
	Remark   string  `json:"remark"`
}
