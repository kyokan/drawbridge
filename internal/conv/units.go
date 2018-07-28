package conv

import (
	"math/big"
	"errors"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func HexToBig(hex string) (*big.Int, error) {
	return hexutil.DecodeBig(hex)
}

func StringToBig(num string) (*big.Int, error) {
	out, success := big.NewInt(0).SetString(num, 10)

	if !success {
		return nil, errors.New("cannot convert " + num + " to bigint")
	}

	return out, nil
}

func BytesToBig(b []byte) (*big.Int, error) {
	hex := hexutil.Encode(b)
	num, ok := math.ParseBig256(hex)
	if !ok {
		return nil, errors.New("invalid bignum")
	}
	return num, nil
}

func BigToBytes(n *big.Int) []byte {
	return math.PaddedBigBytes(math.U256(n), 32)
}

func BigToHex(n *big.Int) string {
	return hexutil.EncodeBig(n)
}
