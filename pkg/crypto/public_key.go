package crypto

import (
	"crypto/ecdsa"
	gocrypto "crypto"
	"github.com/ethereum/go-ethereum/common"
	"crypto/elliptic"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/roasbeef/btcd/btcec"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-errors/errors"
	"crypto/rand"
	"bytes"
)

type PublicKey struct {
	backing *ecdsa.PublicKey
}

func RandomPublicKey() (*PublicKey, error) {
	priv, err := ecdsa.GenerateKey(btcec.S256(), rand.Reader)

	if err != nil {
		return nil, err
	}

	return PublicFromOtherPublic(priv.Public())
}

func PublicFromCompressedHex(hex string) (*PublicKey, error) {
	b, err := hexutil.Decode(hex)

	if err != nil {
		return nil, err
	}

	backing, err := btcec.ParsePubKey(b, btcec.S256())

	if err != nil {
		return nil, err
	}

	return &PublicKey{
		backing: backing.ToECDSA(),
	}, nil
}

func PublicFromOtherPublic(pub gocrypto.PublicKey) (*PublicKey, error) {
	backing, ok := pub.(*ecdsa.PublicKey)

	if !ok {
		return nil, errors.New("public key is not an ECDSA public key")
	}

	return &PublicKey{
		backing: backing,
	}, nil
}

func PublicFromBTCEC(pub *btcec.PublicKey) (*PublicKey, error) {
	return &PublicKey{
		backing: pub.ToECDSA(),
	}, nil
}

func (p *PublicKey) ETHAddress() (common.Address) {
	pubBytes := elliptic.Marshal(btcec.S256(), p.backing.X, p.backing.Y)
	addrBytes := crypto.Keccak256(pubBytes[1:])[12:]
	return common.BytesToAddress(addrBytes)
}

func (p *PublicKey) CompressedHex() string {
	return BTCECToCompressedHex(p.BTCEC())
}

func (p *PublicKey) BTCEC() (*btcec.PublicKey) {
	return (*btcec.PublicKey)(p.backing)
}

func (p *PublicKey) ECDSA() (*ecdsa.PublicKey) {
	return p.backing
}

func (p *PublicKey) Bytes() []byte {
	return elliptic.Marshal(btcec.S256(), p.backing.X, p.backing.Y)
}

func (p *PublicKey) Equal(other *PublicKey) bool {
	return bytes.Equal(p.Bytes(), other.Bytes())
}

func BTCECToCompressedHex(pub *btcec.PublicKey) string {
	return hexutil.Encode(pub.SerializeCompressed())
}
