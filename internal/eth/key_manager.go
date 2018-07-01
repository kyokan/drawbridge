package eth

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"fmt"
)

type KeyManager struct {
	key     *ecdsa.PrivateKey
	chainId *big.Int
}

const SignaturePreamble = "\x19Ethereum Signed Message:\n"

func NewKeyManager(keyHex string, chainId *big.Int) (*KeyManager, error) {
	key, err := crypto.HexToECDSA(keyHex)

	if err != nil {
		return nil, err
	}

	return &KeyManager{
		key:     key,
		chainId: chainId,
	}, nil
}

func (c *KeyManager) NewTransactor(gasOverride uint64) (*bind.TransactOpts) {
	res := bind.NewKeyedTransactor(c.key)

	if gasOverride > 0 {
		res.GasLimit = gasOverride
	}

	return res
}

func (c *KeyManager) SignData(data []byte) ([]byte, error) {
	msg := fmt.Sprintf("%s%d%s", SignaturePreamble, len(data), data)
	hash := crypto.Keccak256([]byte(msg))
	res, err := crypto.Sign(hash, c.key)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *KeyManager) SignTx(tx *types.Transaction) (*types.Transaction, error) {
	tx, err := types.SignTx(tx, types.NewEIP155Signer(c.chainId), c.key)

	if err != nil {
		return nil, err
	}

	return tx, nil
}
