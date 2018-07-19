package conv

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-errors/errors"
)

func HexToBytes32(hex string) ([32]byte, error) {
	var out [32]byte
	b, err := hexutil.Decode(hex)

	if err != nil {
		return out, err
	}

	if len(b) != 32 {
		return out, errors.New("byte length must be exactly 32")
	}

	copy(out[:], b)
	return out, nil
}