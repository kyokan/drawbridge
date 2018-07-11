package txout

import (
	"math/big"
	"github.com/ethereum/go-ethereum/common"
	"io"
	"github.com/kyokan/drawbridge/internal/conv"
)

type OfferedHTLC struct {
	Delay *big.Int
	RedemptionAddress common.Address
	TimeoutAddress common.Address
	PaymentHash common.Hash
}

func (o *OfferedHTLC) OutputType() OutputType {
	return OutputOfferedHTLC
}

func (o *OfferedHTLC) Decode (r io.Reader, pver uint32) error {
	var delayBuf [32]byte
	if _, err := io.ReadFull(r, delayBuf[:]); err != nil {
	    return err
	}
	delay, err := conv.BytesToBig(delayBuf[:])
	if err != nil {
		return err
	}
	o.Delay = delay

	var redeemer common.Address
	var timeout common.Address
	var hash common.Hash
	if _, err := io.ReadFull(r, redeemer[:]); err != nil {
	    return err
	}
	if _, err := io.ReadFull(r, timeout[:]); err != nil {
	    return err
	}
	if _, err := io.ReadFull(r, hash[:]); err != nil {
	    return err
	}
	o.RedemptionAddress = redeemer
	o.TimeoutAddress = timeout
	o.PaymentHash = hash
	return nil
}

func (o *OfferedHTLC) Encode(w io.Writer, pver uint32) error {
	var b [1]byte
	b[0] = byte(OutputOfferedHTLC)
	if _, err := w.Write(b[:]); err != nil {
	    return err
	}
	delayBytes := conv.BigToBytes(o.Delay)
	if _, err := w.Write(delayBytes); err != nil {
	    return err
	}
	if _, err := w.Write(o.RedemptionAddress.Bytes()); err != nil {
	    return err
	}
	if _, err := w.Write(o.TimeoutAddress.Bytes()); err != nil {
	    return err
	}
	if _, err := w.Write(o.PaymentHash.Bytes()); err != nil {
	    return err
	}
	return nil
}