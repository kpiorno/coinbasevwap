package pkg

import (
	"context"
	"vwap/pkg/dtos"
)

//Websocket defines the interface for the websocket client
type Websocket interface {
	Connect(url string) error
	Subscribe(ctx context.Context, request *dtos.Subscription) (<-chan *dtos.Response, error)
	Close()
}
