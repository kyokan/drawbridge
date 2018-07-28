package txout

import (
	"github.com/lightningnetwork/lnd/lnwire"
	"io"
	"math/big"
	"github.com/go-errors/errors"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto/sha3"
)

type OutputType uint8

const (
	OutputPayment         OutputType = 0x01
	OutputMultisig        OutputType = 0x02
	OutputCommitmentLocal OutputType = 0x03
)

func (o OutputType) String() string {
	switch o {
	case OutputPayment:
		return "OutputPayment"
	case OutputMultisig:
		return "OutputMultisig"
	case OutputCommitmentLocal:
		return "OutputCommitmentLocal"
	default:
		return "<unknown output>"
	}
}

type Witness interface {
	Encode(w io.Writer) error
}

type Output interface {
	lnwire.Serializable

	OutputType() OutputType
}

type SpendRequest struct {
	InputID *big.Int
	Witness Witness
	Values []*big.Int
	Outputs []Output
}

func SigData(req *SpendRequest) ([]byte, error) {
	if req.InputID == nil || req.Witness == nil || len(req.Values) == 0 || len(req.Outputs) == 0 {
		return nil, errors.New("input, witness, values and outputs are required")
	}

	if len(req.Values) != len(req.Outputs) {
		return nil, errors.New("number of values must match number of outputs")
	}

	hash := sha3.NewKeccak256()
	hash.Write(math.PaddedBigBytes(math.U256(req.InputID), 32))
	if err := req.Witness.Encode(hash); err != nil {
		return nil, err
	}

	for i := 0; i < len(req.Outputs); i++ {
		out := req.Outputs[i]
		value := req.Values[i]

		if _, err := hash.Write(math.PaddedBigBytes(math.U256(value), 32)); err != nil {
			return nil, err
		}

		if err := out.Encode(hash, 0); err != nil {
			return nil, err
		}
	}

	return hash.Sum(nil), nil
}
