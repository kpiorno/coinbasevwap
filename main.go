package main

import (
	"context"
	"fmt"
	"log"
	"vwap/pkg/coinbase/calculator"
	"vwap/pkg/coinbase/handler"
	"vwap/pkg/std/websocket"
)

const (
	avgDataDelay = 0.0
)

func main() {
	websocket := websocket.NewStdWebsocket()
	// The Delay time for sending the calculated average. Kindly change it as desired.
	vwapCalculator := calculator.NewCoinbaseCalculator(avgDataDelay)
	handler := handler.NewCoinbaseHandler(websocket, vwapCalculator)

	responseChan, err := handler.Subscribe(context.Background(), "BTC-USD", "ETH-USD", "ETH-BTC")
	if err != nil {
		log.Fatal(err)
		return
	}
	for {
		select {
		case <-context.Background().Done():
			return
		case msg := <-responseChan:
			fmt.Printf("Received: %v.\n", msg)
		}
	}
}
