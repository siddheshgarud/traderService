package model

type Order struct {
    ClientID string  `json:"client_id"`
    OrderID  string  `json:"order_id"`
    Price    float64 `json:"price"`
    Qty      float64 `json:"qty"`
    Remark   string  `json:"remark"`
}
