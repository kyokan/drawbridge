package eth

import (
	"github.com/roasbeef/btcd/btcec"
	"crypto/ecdsa"
)

func BTCKeyToETHKey(key *btcec.PublicKey) (*ecdsa.PublicKey) {
	return key.ToECDSA()
}

func ETHKeyToBTCKey(key *ecdsa.PublicKey) (*btcec.PublicKey) {
	return (*btcec.PublicKey)(key)
}