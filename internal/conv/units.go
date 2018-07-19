package conv

import (
	"math/big"
	"errors"
)

var WeiPerSatoshi = big.NewInt(10000000000)

func SatoshiToWei(sats *big.Int) *big.Int {
	return sats.Mul(sats, WeiPerSatoshi)
}

func WeiToSatoshi(wei *big.Int) *big.Int {
	return wei.Div(wei, WeiPerSatoshi)
}

func StringToBig(num string) (*big.Int, error) {
	out, success := big.NewInt(0).SetString(num, 10)

	if !success {
		return nil, errors.New("cannot convert " + num + " to bigint")
	}

	return out, nil
}