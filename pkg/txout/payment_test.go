package txout

import (
	"testing"
	"github.com/ethereum/go-ethereum/common"
	"bytes"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"math/big"
)

func TestPayment_Encode(t *testing.T) {
	addr := common.HexToAddress("0x08e4f70109ccc5135f50cc359d24cb7686247df4")
	payment := NewPayment(addr)
	var buf bytes.Buffer
	payment.Encode(&buf, 0)
	hex := hexutil.Encode(buf.Bytes())
	assert.Equal(t, "0x0108e4f70109ccc5135f50cc359d24cb7686247df4", hex)
}

func TestPayment_Decode(t *testing.T) {
	addr := common.HexToAddress("0x08e4f70109ccc5135f50cc359d24cb7686247df4")
	input, err := hexutil.Decode("0x08e4f70109ccc5135f50cc359d24cb7686247df4")

	if err != nil {
		t.Error(err)
	}

	reader := bytes.NewReader(input)
	payment := &Payment{}
	err = payment.Decode(reader, 0)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, addr, payment.Recipient)
}

func TestPaymentWitness_Encode(t *testing.T) {
	witness := NewPaymentWitness()
	var b bytes.Buffer
	err := witness.Encode(&b)

	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, "0x00", hexutil.Encode(b.Bytes()))
}

func TestPaymentWitness_SigData(t *testing.T) {
	addrA := common.HexToAddress("0x627306090abab3a6e1400e9345bc60c78a8bef57")
	addrB := common.HexToAddress("0xf17f52151ebef6c7334fad080c5704d77216b732")

	req := &SpendRequest{
		InputID: big.NewInt(8),
		Witness: NewPaymentWitness(),
		Values: []*big.Int{
			big.NewInt(1000),
			big.NewInt(99000),
		},
		Outputs: []Output{
			NewPayment(addrB),
			NewPayment(addrA),
		},
	}

	data, err := SigData(req)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, "0x99565010abc704fa632de769e499a546c027e4250e1da7432d634c71724d16ed", hexutil.Encode(data))
}
