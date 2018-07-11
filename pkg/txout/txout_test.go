package txout

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/kyokan/drawbridge/internal/wallet"
	)

var km *wallet.KeyManager

func init() {
	k, err := wallet.NewKeyManager("c87509a1c067bbde78beb793e6fa76530b6382a4c0241e5e4a9ec0a0f44dc0d3", big.NewInt(0))
	if err != nil {
		panic(err)
	}
	km = k
}

func TestGenOutputIDs(t *testing.T) {
	req := dummySpendReq()
	ids, err := GenOutputIDs(req)
	if err != nil {
		t.Fatalf(err.Error())
	}

	assert.Equal(t, len(ids), 2)
	assert.Equal(t, "0x40ad09dd47198ce4fe0a433bad89ff91e3f87ab46a134a4e53dccfcddf10316f", hexutil.Encode(ids[0][:]))
	assert.Equal(t, "0x412152eeb2d6afc5819231fed25490a7034e54ad2cb810ff557396c46444801a", hexutil.Encode(ids[1][:]))
}

func TestWireData_SpendWithChange(t *testing.T) {
	req := dummySpendReq()
	sigData, err := SigData(req)
	if err != nil {
		t.Fatalf(err.Error())
	}
	sig, err := km.SignData(sigData)
	if err != nil {
		t.Fatalf(err.Error())
	}
	in, out, err := WireData(req, sig)
	if err != nil {
		t.Fatalf(err.Error())
	}

	assert.Equal(
		t,
		"0xf2f452833095a6d4a81f0845f5712a67a9bcbec74cad1c1c5c151f2fa62a59c30000000000000000000000000000004200bc88d58947f5e46e15d0ca2baed4e141d31705f087f55936d4c2bc7d4aec549773cca06d396c60063e120fbc5575aead8eb8e235f39dc370d2293abc8d060db201",
		hexutil.Encode(in),
	)
	assert.Equal(
		t,
		"0x00000000000000000000000000000000000000000000000000000000000003e801f17f52151ebef6c7334fad080c5704d77216b73200000000000000000000000000000000000000000000000000000000000182b801627306090abab3a6e1400e9345bc60c78a8bef57",
		hexutil.Encode(out),
	)
}

func TestWireData_SingleOutputMultisig(t *testing.T) {
	addrA := common.HexToAddress("0x627306090abab3a6e1400e9345bc60c78a8bef57")
	addrB := common.HexToAddress("0xf17f52151ebef6c7334fad080c5704d77216b732")
	inputId := common.HexToHash("0xca89f758cdbc84fa247709df12b479ed8cf2f0eae5af8e820277dc86fc37054e")

	req := &SpendRequest{
		InputID: inputId,
		Witness: NewPaymentWitness(),
		Values: []*big.Int{
			big.NewInt(100000),
		},
		Outputs: []Output{
			NewMultisig(addrA, addrB),
		},
	}
	sigData, err := SigData(req)
	if err != nil {
		t.Fatalf(err.Error())
	}
	sig, err := km.SignData(sigData)
	if err != nil {
		t.Fatalf(err.Error())
	}
	in, out, err := WireData(req, sig)
	if err != nil {
		t.Fatalf(err.Error())
	}

	assert.Equal(
		t,
		"0xca89f758cdbc84fa247709df12b479ed8cf2f0eae5af8e820277dc86fc37054e00000000000000000000000000000042008260ac6c2295d34ef1176d6d151df6147d1ce4e44819160b3c66ee41aedde2e465d739aa0f624dc863d12d675d65a8d659dea441e135eee56c7e7a83f2d2344601",
		hexutil.Encode(in),
	)
	assert.Equal(
		t,
		"0x00000000000000000000000000000000000000000000000000000000000186a002627306090abab3a6e1400e9345bc60c78a8bef57f17f52151ebef6c7334fad080c5704d77216b732",
		hexutil.Encode(out),
	)
}

func dummySpendReq() *SpendRequest {
	addrA := common.HexToAddress("0x627306090abab3a6e1400e9345bc60c78a8bef57")
	addrB := common.HexToAddress("0xf17f52151ebef6c7334fad080c5704d77216b732")
	inputId := common.HexToHash("0xf2f452833095a6d4a81f0845f5712a67a9bcbec74cad1c1c5c151f2fa62a59c3")

	return &SpendRequest{
		InputID: inputId,
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
}
