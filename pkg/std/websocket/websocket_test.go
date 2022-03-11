package websocket

import (
	"context"
	"encoding/json"
	"log"
	"math/big"
	"net/http/httptest"
	"strings"
	"testing"
	"vwap/pkg/dtos"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/websocket"
)

func connectTest(ws *websocket.Conn) {
}

func subscribeResponseError(ws *websocket.Conn) {
	var msg = make([]byte, 2048)
	if n, err := ws.Read(msg); err != nil {
		log.Print(msg[:n])
	}
	response := "bad response{"
	if _, err := ws.Write([]byte(response)); err != nil {
		log.Fatal(err)
	}
}

func subscribeSuccessResponse(ws *websocket.Conn) {
	var msg = make([]byte, 2048)
	if n, err := ws.Read(msg); err != nil {
		log.Print(msg[:n])
	}
	response, _ := json.Marshal(&dtos.Response{
		Type:      "match",
		ProductId: "BTC-USD",
		Price:     big.NewFloat(4.0),
		Size:      big.NewFloat(4.0),
	})
	if _, err := ws.Write([]byte(response)); err != nil {
		log.Fatal(err)
	}
}

func TestStdWebsocket_Websocket(t *testing.T) {
	const (
		success = iota
		subscriptionSuccess
		subscriptionResponseError
	)
	tests := []struct {
		name     string
		testType int
	}{
		{
			name:     "test success connection",
			testType: success,
		},
		{
			name:     "test subscription success",
			testType: subscriptionSuccess,
		},
		{
			name:     "test subscription response error",
			testType: subscriptionResponseError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.testType {
			case success:
				server := httptest.NewServer(websocket.Handler(connectTest))
				defer server.Close()
				s := NewStdWebsocket()
				u := "ws" + strings.TrimPrefix(server.URL, "http")
				err := s.Connect(u)
				assert.NoError(t, err)
			case subscriptionSuccess:
				server := httptest.NewServer(websocket.Handler(subscribeSuccessResponse))
				defer server.Close()
				s := NewStdWebsocket()
				u := "ws" + strings.TrimPrefix(server.URL, "http")
				err := s.Connect(u)
				assert.NoError(t, err)
				response, err := s.Subscribe(context.Background(), &dtos.Subscription{
					Type:       "subscribe",
					ProductIds: []string{"BTC-USD", "ETH-USD", "ETH-BTC"},
					Channels:   []string{"matches"},
				})
				assert.NotNil(t, response)
				assert.NoError(t, err)
			case subscriptionResponseError:
				server := httptest.NewServer(websocket.Handler(subscribeResponseError))
				defer server.Close()
				s := NewStdWebsocket()
				u := "ws" + strings.TrimPrefix(server.URL, "http")
				err := s.Connect(u)
				assert.NoError(t, err)
				response, err := s.Subscribe(context.Background(), &dtos.Subscription{
					Type:       "subscribe",
					ProductIds: []string{"BTC-USD", "ETH-USD", "ETH-BTC"},
					Channels:   []string{"matches"},
				})
				assert.NotNil(t, response)
				assert.NoError(t, err)
			}
		})
	}
}
