package calculator

import (
	"context"
	"log"
	"math/big"
	"time"
	"vwap/pkg"
	"vwap/pkg/dtos"
)

var _ pkg.VWAPCalculator = &CoinbaseVWAPCalculator{}

const (
	slidingWindow = 200
	unmarshalErr  = "unmarshal_error"
	errorType     = "error"
)

type AvgData struct {
	points              []*dtos.Response
	currentPos          int64
	CalculatedVwap      *big.Float
	totalWeightedValues *big.Float
	totalWeights        *big.Float
}

func (a *AvgData) Add(point *dtos.Response) {
	// This code was refactored like this way in contrary
	// to iterate all elements each time a new data arrives
	mul := new(big.Float)
	mul.Mul(point.Price, point.Size)
	a.totalWeightedValues.Add(a.totalWeightedValues, mul)
	a.totalWeights.Add(a.totalWeights, point.Size)
	if a.currentPos >= slidingWindow {
		price, size := a.points[0].Price, a.points[0].Size
		mul := new(big.Float)
		mul.Mul(price, size)
		a.points = append(a.points[1:], point)
		a.totalWeightedValues.Sub(a.totalWeightedValues, mul)
		a.totalWeights.Sub(a.totalWeights, size)
	} else {
		a.points[a.currentPos] = point
		a.currentPos++
	}
	a.CalculatedVwap = new(big.Float).Quo(a.totalWeightedValues, a.totalWeights)
}

type CoinbaseVWAPCalculator struct {
	exit        chan struct{}
	currentTime time.Time
	delay       float64
	maxDelay    float64
	productAvgs map[string]*AvgData
}

func (c *CoinbaseVWAPCalculator) calcAvg(data *dtos.Response) {
	var avgdata *AvgData
	var ok bool
	if avgdata, ok = c.productAvgs[data.ProductId]; !ok {
		avgdata = &AvgData{
			points:              make([]*dtos.Response, slidingWindow),
			CalculatedVwap:      new(big.Float),
			totalWeightedValues: new(big.Float),
			totalWeights:        new(big.Float),
		}
		c.productAvgs[data.ProductId] = avgdata
	}
	avgdata.Add(data)
}

// checkDelay checks if it is time to send the calculated avg
func (c *CoinbaseVWAPCalculator) checkDelay() bool {
	var update bool
	diff := time.Since(c.currentTime)
	c.delay += diff.Seconds()
	if c.delay >= c.maxDelay {
		update = true
		c.delay -= c.maxDelay
	}
	c.currentTime = time.Now()
	return update
}

// sendProductAvgs converts the currently calculated avg into the final data type
func (c *CoinbaseVWAPCalculator) sendProductAvgs(productAvgs chan<- *dtos.ProductAvgs) {
	response := &dtos.ProductAvgs{
		Products: make(map[string]*big.Float),
	}
	for k, v := range c.productAvgs {
		response.Products[k] = v.CalculatedVwap
	}
	productAvgs <- response
}

// CalcAvg processes all the coinbase responses in real-time, calculates the avg and sends the computed avg.
func (c *CoinbaseVWAPCalculator) CalcAvg(ctx context.Context, responseChan <-chan *dtos.Response) (<-chan *dtos.ProductAvgs, error) {
	response := make(chan *dtos.ProductAvgs)
	go func() {
		for {
			select {
			case <-c.exit:
				return
			case <-ctx.Done():
				return
			case msg := <-responseChan:
				//ignores empty responses
				//TODO: Handle unmarshalErr cases
				if *msg == (dtos.Response{}) || msg.Type == unmarshalErr {
					continue
				}
				if msg.Type == errorType {
					log.Printf(msg.Error.Message)
					return
				}
				c.calcAvg(msg)
				if c.checkDelay() {
					c.sendProductAvgs(response)
				}
				time.Sleep(time.Millisecond * 10)
			}
		}
	}()
	return response, nil
}

func (c *CoinbaseVWAPCalculator) Close() {
	c.exit <- struct{}{}
}

func NewCoinbaseCalculator(maxDelay float64) *CoinbaseVWAPCalculator {
	return &CoinbaseVWAPCalculator{
		exit:        make(chan struct{}),
		productAvgs: make(map[string]*AvgData),
		currentTime: time.Now(),
		maxDelay:    maxDelay,
	}
}
