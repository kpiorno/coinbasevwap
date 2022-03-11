package calculator

import (
	"context"
	"math/big"
	"sync"
	"testing"
	"vwap/pkg/dtos"

	"github.com/stretchr/testify/assert"
)

func TestCoinbaseVWAPCalculator_CalcAvg(t *testing.T) {
	const (
		success = iota
		responseError
		maxSlidingWindowReached
	)
	ctx := context.Background()
	type args struct {
		ctx          context.Context
		responseChan <-chan *dtos.Response
	}
	tests := []struct {
		name     string
		args     args
		testType int
	}{
		{
			name: "test calc avg success",
			args: args{
				ctx: ctx,
			},
			testType: success,
		},
		{
			name: "test response error",
			args: args{
				ctx: ctx,
			},
			testType: responseError,
		},
		{
			name: "test max sliding window reached",
			args: args{
				ctx: ctx,
			},
			testType: maxSlidingWindowReached,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCoinbaseCalculator(0)
			switch tt.testType {
			case success:
				tt.args.responseChan = func() <-chan *dtos.Response {
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
				response, err := c.CalcAvg(tt.args.ctx, tt.args.responseChan)
				assert.NotNil(t, response)
				assert.NoError(t, err)
			case responseError:
				tt.args.responseChan = func() <-chan *dtos.Response {
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
				response, err := c.CalcAvg(tt.args.ctx, tt.args.responseChan)
				assert.NotNil(t, response)
				assert.NoError(t, err)
			case maxSlidingWindowReached:
				var wg sync.WaitGroup

				tt.args.responseChan = func() <-chan *dtos.Response {
					cbCount := slidingWindow + 10
					wg.Add(cbCount)
					coinbaseResponses := make([]*dtos.Response, cbCount)
					for i := 0; i < cbCount; i++ {
						coinbaseResponses[i] = &dtos.Response{

							ProductId: "BTC-USD",
							Type:      "match",
							Price:     big.NewFloat(4.0),
							Size:      big.NewFloat(4.0),
						}
					}
					response := make(chan *dtos.Response)
					go func() {
						for _, res := range coinbaseResponses {
							wg.Done()
							response <- res
						}
					}()
					return response
				}()
				response, err := c.CalcAvg(tt.args.ctx, tt.args.responseChan)
				go func() {
					for range response {
						continue
					}
				}()
				wg.Wait()
				assert.NotNil(t, response)
				assert.NoError(t, err)
			}

		})
	}
}
