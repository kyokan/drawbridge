package wire

import (
	"testing"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"math/big"
)

func TestEncumbrance_Hashable(t *testing.T) {
	e := makeEncumbrance()
	expected := "0x927f83791a88ff72c6cfe4eeb9dd96a801fbc48f9e3225d9e3e9dbea67b57e8600000000000000000000000000000000000000000000000000000000000003e800000000000000000000000000000000000000000000000000000000000173180000000000000000000000000000000000000000000000000000000000001388ea05319122ecf34a553669191848370ff785fe00ee6f01d3d9a8e4be7eee5249"
	actual := hexutil.Encode(e.Hashable())

	if expected != actual {
		t.Errorf("Expected %s, got: %s", expected, actual)
	}
}

func TestEncumbrance_Hash(t *testing.T) {
	e := makeEncumbrance()
	expected := "0x9a3af8416df249e86d4e508718ab492a00471ca8fb23992a5dc56a158e885cc3"
	hash := e.Hash()
	actual := hexutil.Encode(hash[:])

	if expected != actual {
		t.Errorf("Expected %s, got: %s", expected, actual)
	}
}

func makeEncumbrance() (*Encumbrance) {
	inputIdSlice, _ := hexutil.Decode("0x927f83791a88ff72c6cfe4eeb9dd96a801fbc48f9e3225d9e3e9dbea67b57e86")
	hashLockSlice, _ := hexutil.Decode("0xea05319122ecf34a553669191848370ff785fe00ee6f01d3d9a8e4be7eee5249")

	var inputId [32]byte
	var hashLock [32]byte

	copy(inputId[:], inputIdSlice)
	copy(hashLock[:], hashLockSlice)

	return &Encumbrance{
		InputId:  inputId,
		LockTime: big.NewInt(1000),
		ValueA:   big.NewInt(95000),
		ValueB:   big.NewInt(5000),
		HashLock: hashLock,
	}
}
