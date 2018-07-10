package eth

import "math/big"

var WeiPerSatoshi = big.NewInt(10000000000)

func SatoshiToWei(sats *big.Int) *big.Int {
	return sats.Mul(sats, WeiPerSatoshi)
}

func WeiToSatoshi(wei *big.Int) *big.Int {
	return wei.Div(wei, WeiPerSatoshi)
}