package conv

import (
	"github.com/roasbeef/btcd/btcec"
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func BTCKeyToETHKey(key *btcec.PublicKey) (*ecdsa.PublicKey) {
	return key.ToECDSA()
}

func ETHKeyToBTCKey(key *ecdsa.PublicKey) (*btcec.PublicKey) {
	return (*btcec.PublicKey)(key)
}

func PubKeyToHex(key *btcec.PublicKey) string {
	return hexutil.Encode(key.SerializeCompressed())
}