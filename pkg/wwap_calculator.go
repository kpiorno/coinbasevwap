package pkg

import (
	"context"
	"vwap/pkg/dtos"
)

//VWAPCalculator defines the interface for the vwap calculation
type VWAPCalculator interface {
	CalcAvg(ctx context.Context, responseChan <-chan *dtos.Response) (<-chan *dtos.ProductAvgs, error)
	Close()
}
