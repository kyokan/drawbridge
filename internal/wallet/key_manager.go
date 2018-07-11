package wallet

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/core/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"math/big"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/kyokan/drawbridge/pkg/crypto"
)

type KeyManager struct {
	key     *ecdsa.PrivateKey
	chainId *big.Int
}

func NewKeyManager(keyHex string, chainId *big.Int) (*KeyManager, error) {
	key, err := ethcrypto.HexToECDSA(keyHex)

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

func (c *KeyManager) SignData(data []byte) (crypto.Signature, error) {
	hash := crypto.GethHash(data)
	res, err := ethcrypto.Sign(hash, c.key)

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

func (c *KeyManager) PublicKey() *crypto.PublicKey {
	k, err := crypto.PublicFromOtherPublic(c.key.Public())

	if err != nil {
		// should never happen since the private key is
		// validated upon struct construction
		panic(err)
	}

	return k
}
