package handler

import (
	"context"
	"errors"
	pkg "vwap/pkg"
	"vwap/pkg/dtos"
)

var _ pkg.VWAPHandler = &CoinbaseHandler{}

const (
	url = "wss://ws-feed.exchange.coinbase.com"
)

type CoinbaseHandler struct {
	websocket      pkg.Websocket
	vwapCalculator pkg.VWAPCalculator
}

// createSubscriptionPayload it's a helper function for creating the subscription payload
func (c *CoinbaseHandler) createSubscriptionPayload(productIds []string) *dtos.Subscription {
	return &dtos.Subscription{
		Type:       "subscribe",
		ProductIds: productIds,
		Channels:   []string{"matches"},
	}
}

// Subscribe function subscribes to the coinbase match channel in order to process responses
func (c *CoinbaseHandler) Subscribe(ctx context.Context, productIds ...string) (<-chan *dtos.ProductAvgs, error) {
	if len(productIds) == 0 {
		return nil, errors.New("no product id provided")
	}
	err := c.websocket.Connect(url)
	if err != nil {
		return nil, err
	}
	subscription := c.createSubscriptionPayload(productIds)
	websocketChan, err := c.websocket.Subscribe(ctx, subscription)
	if err != nil {
		return nil, err
	}
	responseChan, err := c.vwapCalculator.CalcAvg(ctx, websocketChan)
	if err != nil {
		return nil, err
	}

	return responseChan, nil
}

func (c *CoinbaseHandler) Close() {
	go c.websocket.Close()
	go c.vwapCalculator.Close()
}

func NewCoinbaseHandler(websocket pkg.Websocket, vwapCalculator pkg.VWAPCalculator) *CoinbaseHandler {
	return &CoinbaseHandler{
		websocket:      websocket,
		vwapCalculator: vwapCalculator,
	}
}
