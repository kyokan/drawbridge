package wire

import (
	"testing"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common"
)

func TestMultisig_Hashable(t *testing.T) {
	expected := "0x92a383356d4cc150ea3f18853b46f61d1a34d4553fc1d77faf4c8d7c80529eeba9096246fb68566f5cd946475650bcec96145c17c3dbf2753705a92af0db269f627306090abab3a6e1400e9345bc60c78a8bef57f17f52151ebef6c7334fad080c5704d77216b732"
	multisig := makeMultisig()
	actual := hexutil.Encode(multisig.Hashable())

	if expected != actual {
		t.Errorf("Expected %s, but got %s", expected, actual)
	}
}

func TestMultisig_Hash(t *testing.T) {
	expected := "0x17c5573674d583a72ce017c2c19e2b68c9770e3b24f39fedc600c8b743c4e69b"
	multisig := makeMultisig()
	hash := multisig.Hash()
	actual := hexutil.Encode(hash[:])

	if expected != actual {
		t.Errorf("Expected %s, but got %s", expected, actual)
	}
}

func makeMultisig() (*Multisig) {
	inputASlice, _ := hexutil.Decode("0x92a383356d4cc150ea3f18853b46f61d1a34d4553fc1d77faf4c8d7c80529eeb")
	inputBSlice, _ := hexutil.Decode("0xa9096246fb68566f5cd946475650bcec96145c17c3dbf2753705a92af0db269f")
	signerA := common.HexToAddress("0x627306090abab3a6e1400e9345bc60c78a8bef57")
	signerB := common.HexToAddress("0xf17f52151ebef6c7334fad080c5704d77216b732")

	var inputA [32]byte
	var inputB [32]byte

	copy(inputA[:], inputASlice)
	copy(inputB[:], inputBSlice)

	return &Multisig{
		InputA:  inputA,
		InputB:  inputB,
		SignerA: signerA,
		SignerB: signerB,
	}
}
