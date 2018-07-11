package types

import (
	"math/big"
	"github.com/kyokan/drawbridge/pkg/crypto"
)

type Channel struct {
	OurFundingAddress   *crypto.PublicKey
	TheirFundingAddress *crypto.PublicKey
	FundingAmount       *big.Int
	OurSignature        crypto.Signature
	TheirSignature      crypto.Signature
	InputID             [32]byte
}
