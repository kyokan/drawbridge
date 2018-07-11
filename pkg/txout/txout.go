package txout

import (
	"github.com/lightningnetwork/lnd/lnwire"
	"io"
	"math/big"
	"github.com/go-errors/errors"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/ethereum/go-ethereum/common"
	"github.com/kyokan/drawbridge/internal/conv"
	"github.com/kyokan/drawbridge/pkg/crypto"
	"bytes"
	"encoding/binary"
)

type OutputType uint8

const (
	OutputPayment         OutputType = 0x01
	OutputMultisig                   = 0x02
	OutputCommitmentLocal            = 0x03
	OutputOfferedHTLC                = 0x04
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
	InputID common.Hash
	Witness Witness
	Values  []*big.Int
	Outputs []Output
}

func SigData(req *SpendRequest) ([]byte, error) {
	if req.Witness == nil || len(req.Values) == 0 || len(req.Outputs) == 0 {
		return nil, errors.New("input, witness, values and outputs are required")
	}

	if len(req.Values) != len(req.Outputs) {
		return nil, errors.New("number of values must match number of outputs")
	}

	hash := sha3.NewKeccak256()
	if _, err := hash.Write(req.InputID[:]); err != nil {
		return nil, err
	}

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

func WireData(req *SpendRequest, sig crypto.Signature) ([]byte, []byte, error) {
	var inputsWire bytes.Buffer
	if _, err := inputsWire.Write(req.InputID[:]); err != nil {
		return nil, nil, err
	}

	var witBuf bytes.Buffer
	if err := req.Witness.Encode(&witBuf); err != nil {
		return nil, nil, err
	}
	if _, err := witBuf.Write(sig.Bytes()); err != nil {
		return nil, nil, err
	}
	if witBuf.Len() > math.MaxUint16 {
		return nil, nil, errors.New("witness too long")
	}

	var pad [14]byte
	var l [2]byte
	binary.BigEndian.PutUint16(l[:], uint16(witBuf.Len()))
	if _, err := inputsWire.Write(pad[:]); err != nil {
		return nil, nil, err
	}
	if _, err := inputsWire.Write(l[:]); err != nil {
		return nil, nil, err
	}
	if _, err := inputsWire.Write(witBuf.Bytes()); err != nil {
		return nil, nil, err
	}

	var outputsWire bytes.Buffer
	for i := 0; i < len(req.Outputs); i++ {
		value := req.Values[i]
		out := req.Outputs[i]
		if _, err := outputsWire.Write(conv.BigToBytes(value)); err != nil {
			return nil, nil, err
		}
		if err := out.Encode(&outputsWire, 0); err != nil {
			return nil, nil, err
		}
	}

	return inputsWire.Bytes(), outputsWire.Bytes(), nil
}

func GenOutputIDs(req *SpendRequest) ([]common.Hash, error) {
	var inputId common.Hash
	h := sha3.NewKeccak256()
	if _, err := h.Write(req.InputID[:]); err != nil {
		return nil, err
	}
	h.Sum(inputId[:0])

	var ids []common.Hash
	for i := 0; i < len(req.Outputs); i++ {
		outH := sha3.NewKeccak256()
		out := req.Outputs[i]
		value := req.Values[i]
		outH.Write(inputId[:])
		if err := out.Encode(outH, 0); err != nil {
			return nil, err
		}
		if _, err := outH.Write(conv.BigToBytes(value)); err != nil {
			return nil, err
		}
		if _, err := outH.Write(conv.BigToBytes(big.NewInt(int64(i)))); err != nil {
			return nil, err
		}

		var id common.Hash
		outH.Sum(id[:0])
		ids = append(ids, id)
	}
	return ids, nil
}
