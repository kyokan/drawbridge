package txout

import (
	"testing"
	"github.com/ethereum/go-ethereum/common"
	"bytes"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
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
	req := dummySpendReq()
	data, err := SigData(req)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, "0xa52ebd83481f00bc028d375b41c2099ac39f22f1c51a0e84874ce1a4c693ae0a", hexutil.Encode(data))
}
