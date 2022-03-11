package dtos

import (
	"math/big"
	"time"
)

//Response defines the response from Coinbase matches
type Response struct {
	Type         string     `json:"match"`
	TradeId      int64      `json:"trade_id"`
	Sequence     int64      `json:"sequence"`
	MakerOrderId string     `json:"maker_order_id"`
	TakerOrderId string     `json:"taker_order_id"`
	Time         time.Time  `json:"time"`
	ProductId    string     `json:"product_id"`
	Size         *big.Float `json:"size"`
	Price        *big.Float `json:"price"`
	Side         string     `json:"side"`
	Error        `json:",inline"`
}
