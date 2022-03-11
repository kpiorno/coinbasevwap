package dtos

import (
	"math/big"
)

//ProductAvgs defines the data struct that containts all the data points according to the sliding window
type ProductAvgs struct {
	Products map[string]*big.Float
}
