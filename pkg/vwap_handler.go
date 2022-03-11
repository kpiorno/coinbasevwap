package pkg

import (
	"context"
	"vwap/pkg/dtos"
)

//VWAPHandler defines the interface for the main vwap calculator handler
type VWAPHandler interface {
	Subscribe(ctx context.Context, productIds ...string) (<-chan *dtos.ProductAvgs, error)
	Close()
}
