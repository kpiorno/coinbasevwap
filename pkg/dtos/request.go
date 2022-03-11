package dtos

//Subscription defines the payload that is going to use for subscribing to coinbase channeles
type Subscription struct {
	Type       string   `json:"type"`
	ProductIds []string `json:"product_ids"`
	Channels   []string `json:"channels"`
}
