package wallet

import (
	"testing"
	"math/big"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func TestKeyManager_SignData(t *testing.T) {
	km, err := NewKeyManager("c87509a1c067bbde78beb793e6fa76530b6382a4c0241e5e4a9ec0a0f44dc0d3", big.NewInt(1))

	if err != nil {
		t.Fatalf("Got error: %s", err)
	}

	data, _ := hexutil.Decode("0x9a3af8416df249e86d4e508718ab492a00471ca8fb23992a5dc56a158e885cc3")
	expected := "0x677ce6711527af70e94556bd48b078fed08b4a3b38e97bd7cc80a6a7f73f2774322ca8bfbba71c7d9201c602ca63272338f490b02ba678738855feba5d471a3c00"
	actualBytes, err := km.SignData(data)
	actual := hexutil.Encode(actualBytes)

	if err != nil {
		t.Errorf("Got error: %s", err)
	}

	if expected != actual {
		t.Errorf("Expected %s, but got %s", expected, actual)
	}
}
