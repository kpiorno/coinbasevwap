package handler

import (
	"context"
	"errors"
	"math/big"
	"testing"
	"vwap/pkg/coinbase/calculator"
	"vwap/pkg/dtos"
	"vwap/pkg/mocks"

	"github.com/stretchr/testify/assert"
)

func TestCoinbaseHandler_Subscribe(t *testing.T) {
	ctx := context.Background()
	const (
		success = iota
		websocketConnectError
		websocketSubscribeError
		websocketResponseError
		calculatorCalcAvgError
		noProductIdProvidedError
	)
	type args struct {
		ctx        context.Context
		productIds []string
	}
	commonArgs := args{
		ctx:        ctx,
		productIds: []string{"BTC-USD", "ETH-USD", "ETH-BTC"},
	}
	tests := []struct {
		name     string
		args     args
		testType int
	}{
		{
			name:     "test subscription sucessfully",
			args:     commonArgs,
			testType: success,
		},
		{
			name:     "test websocket connect error",
			args:     commonArgs,
			testType: websocketConnectError,
		},
		{
			name:     "test websocket subscribe error",
			args:     commonArgs,
			testType: websocketSubscribeError,
		},
		{
			name:     "test websocket response error",
			args:     commonArgs,
			testType: websocketResponseError,
		},
		{
			name:     "test calculator calcAvg error",
			args:     commonArgs,
			testType: calculatorCalcAvgError,
		},
		{
			name: "test no product id provided error",
			args: args{
				ctx:        ctx,
				productIds: []string{},
			},
			testType: noProductIdProvidedError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.testType {
			case success:
				websocket := &mocks.Websocket{}
				vwapCalculator := calculator.NewCoinbaseCalculator(0.0)
				c := &CoinbaseHandler{
					websocket:      websocket,
					vwapCalculator: vwapCalculator,
				}
				responseChan := func() <-chan *dtos.Response {
					coinbaseResponses := []*dtos.Response{
						{
							ProductId: "BTC-USD",
							Type:      "match",
							Price:     big.NewFloat(4.0),
							Size:      big.NewFloat(4.0),
						},
						{
							ProductId: "BTC-USD",
							Type:      "match",
							Price:     big.NewFloat(4.0),
							Size:      big.NewFloat(4.0),
						},
					}
					response := make(chan *dtos.Response)
					go func() {
						for _, res := range coinbaseResponses {
							response <- res
						}
					}()
					return response
				}()

				websocket.On("Connect", url).Return(nil)
				websocket.On("Subscribe", ctx, &dtos.Subscription{
					Type:       "subscribe",
					ProductIds: tt.args.productIds,
					Channels:   []string{"matches"},
				}).Return(responseChan, nil)
				websocket.On("Close").Return()
				productAvgs, err := c.Subscribe(tt.args.ctx, tt.args.productIds...)
				assert.NotNil(t, productAvgs)
				assert.NoError(t, err)
				vwapCalc := <-productAvgs
				assert.NotNil(t, vwapCalc.Products["BTC-USD"])
				c.Close()
			case websocketConnectError:
				websocket := &mocks.Websocket{}
				vwapCalculator := calculator.NewCoinbaseCalculator(0.0)
				c := &CoinbaseHandler{
					websocket:      websocket,
					vwapCalculator: vwapCalculator,
				}
				websocket.On("Connect", url).Return(errors.New(""))
				productAvgs, err := c.Subscribe(tt.args.ctx, tt.args.productIds...)
				assert.Nil(t, productAvgs)
				assert.Error(t, err)
			case websocketSubscribeError:
				websocket := &mocks.Websocket{}
				vwapCalculator := calculator.NewCoinbaseCalculator(0.0)
				c := &CoinbaseHandler{
					websocket:      websocket,
					vwapCalculator: vwapCalculator,
				}
				websocket.On("Connect", url).Return(nil)
				websocket.On("Subscribe", ctx, &dtos.Subscription{
					Type:       "subscribe",
					ProductIds: tt.args.productIds,
					Channels:   []string{"matches"},
				}).Return(nil, errors.New(""))
				productAvgs, err := c.Subscribe(tt.args.ctx, tt.args.productIds...)
				assert.Nil(t, productAvgs)
				assert.Error(t, err)
			case websocketResponseError:
				websocket := &mocks.Websocket{}
				vwapCalculator := calculator.NewCoinbaseCalculator(0.0)
				c := &CoinbaseHandler{
					websocket:      websocket,
					vwapCalculator: vwapCalculator,
				}
				responseChan := func() <-chan *dtos.Response {
					coinbaseResponses := []*dtos.Response{
						{
							Type: "error",
							Error: dtos.Error{
								Message: "unexpected error",
							},
						},
					}
					response := make(chan *dtos.Response)
					go func() {
						for _, res := range coinbaseResponses {
							response <- res
						}
					}()
					return response
				}()

				websocket.On("Connect", url).Return(nil)
				websocket.On("Subscribe", ctx, &dtos.Subscription{
					Type:       "subscribe",
					ProductIds: tt.args.productIds,
					Channels:   []string{"matches"},
				}).Return(responseChan, nil)
				productAvgs, err := c.Subscribe(tt.args.ctx, tt.args.productIds...)
				assert.NotNil(t, productAvgs)
				assert.NoError(t, err)
			case calculatorCalcAvgError:
				websocket := &mocks.Websocket{}
				vwapCalculator := &mocks.VWAPCalculator{}
				c := &CoinbaseHandler{
					websocket:      websocket,
					vwapCalculator: vwapCalculator,
				}
				responseChan := make(<-chan *dtos.Response)
				websocket.On("Connect", url).Return(nil)
				websocket.On("Subscribe", ctx, &dtos.Subscription{
					Type:       "subscribe",
					ProductIds: tt.args.productIds,
					Channels:   []string{"matches"},
				}).Return(responseChan, nil)
				vwapCalculator.On("CalcAvg", ctx, responseChan).Return(nil, errors.New(""))
				productAvgs, err := c.Subscribe(tt.args.ctx, tt.args.productIds...)
				assert.Nil(t, productAvgs)
				assert.Error(t, err)
			case noProductIdProvidedError:
				websocket := &mocks.Websocket{}
				vwapCalculator := &mocks.VWAPCalculator{}
				c := &CoinbaseHandler{
					websocket:      websocket,
					vwapCalculator: vwapCalculator,
				}
				productAvgs, err := c.Subscribe(tt.args.ctx, tt.args.productIds...)
				assert.Nil(t, productAvgs)
				assert.Error(t, err)
			}
		})
	}
}
