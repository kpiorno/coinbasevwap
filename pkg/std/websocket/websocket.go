package websocket

import (
	"context"
	"encoding/json"
	"time"
	"vwap/pkg"
	"vwap/pkg/dtos"

	"golang.org/x/net/websocket"
)

var _ pkg.Websocket = &StdWebsocket{}

const (
	unmarshalErr = "unmarshal_error"
)

type StdWebsocket struct {
	ws   *websocket.Conn
	exit chan struct{}
}

func (s *StdWebsocket) Connect(url string) error {
	origin := "http://localhost/"
	ws, err := websocket.Dial(url, "", origin)
	s.ws = ws
	return err
}

func (s *StdWebsocket) Subscribe(ctx context.Context, request *dtos.Subscription) (<-chan *dtos.Response, error) {
	payload, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	if _, err := s.ws.Write([]byte(payload)); err != nil {
		return nil, err
	}
	// let's be garbagged collected for the moment
	responseChan := make(chan *dtos.Response)
	if _, err := s.ws.Write([]byte(payload)); err != nil {
		return nil, err
	}
	go func() {
		defer s.ws.Close()
		for {
			select {
			case <-s.exit:
				return
			case <-ctx.Done():
				return
			default:
				var msg = make([]byte, 2048)
				var n int
				if n, err = s.ws.Read(msg); err != nil {
					s.dispatchError(responseChan, err.Error())
					continue
				}
				res := &dtos.Response{}
				err := json.Unmarshal(msg[:n], res)
				if err != nil {
					s.dispatchError(responseChan, err.Error())
					continue
				}
				responseChan <- res
				time.Sleep(time.Millisecond * 10)
			}
		}
	}()
	return responseChan, nil

}

func (s *StdWebsocket) Close() {
	s.exit <- struct{}{}
}

func (s *StdWebsocket) dispatchError(responseChan chan *dtos.Response, msg string) {
	responseChan <- &dtos.Response{
		Type: unmarshalErr,
		Error: dtos.Error{
			Message: msg,
		},
	}
}
func NewStdWebsocket() *StdWebsocket {
	return &StdWebsocket{
		exit: make(chan struct{}),
	}
}
