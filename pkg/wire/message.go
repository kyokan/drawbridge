package wire

import (
	"io"
	"github.com/lightningnetwork/lnd/lnwire"
	"encoding/binary"
	"github.com/go-errors/errors"
	"github.com/kyokan/drawbridge/pkg/crypto"
	"net"
	"github.com/btcsuite/btcd/btcec"
	"math/big"
	"github.com/kyokan/drawbridge/internal/conv"
	"reflect"
	"github.com/ethereum/go-ethereum/common"
)

const MsgPrefix = 0xbeef

const UnknownMsg = "<unknown>"

const (
	MsgInit             lnwire.MessageType = 16
	MsgOpenChannel                         = 32
	MsgAcceptChannel                       = 33
	MsgFundingCreated                      = 34
	MsgFundingSigned                       = 35
	MsgFundingLocked                       = 36
	MsgInitiateSwap                        = 900
	MsgSwapAccepted                        = 901
	MsgInvoiceGenerated                    = 902
	MsgInvoiceExecuted                     = 903
)

func readElement(r io.Reader, element interface{}) error {
	var err error
	switch e := element.(type) {
	case **crypto.PublicKey:
		var b [btcec.PubKeyBytesLenCompressed]byte
		if _, err = io.ReadFull(r, b[:]); err != nil {
			return err
		}

		pub, err := crypto.PublicFromBytes(b[:])
		if err != nil {
			return err
		}
		*e = pub
	case **net.TCPAddr:
		addr := &net.TCPAddr{}
		_, err = ReadTCPAddr(r, addr)
		if err != nil {
			return err
		}
		*e = addr
	case *string:
		var l [2]byte
		if _, err := io.ReadFull(r, l[:]); err != nil {
			return err
		}
		strLen := binary.BigEndian.Uint16(l[:])
		strBytes := make([]byte, strLen)
		if _, err := io.ReadFull(r, strBytes); err != nil {
			return err
		}
		str := string(strBytes)
		*e = str
	case *[32]byte:
		var b [32]byte
		if _, err := io.ReadFull(r, b[:]); err != nil {
			return err
		}
		*e = b
	case *common.Hash:
		var b [32]byte
		if _, err := io.ReadFull(r, b[:]); err != nil {
			return err
		}
		*e = b
	case *[]byte:
		buf, err := readByteSliceLike(r)
		if err != nil {
			return err
		}
		*e = buf
	case *crypto.Signature:
		buf, err := readByteSliceLike(r)
		if err != nil {
			return err
		}
		*e = buf
	case **big.Int:
		var b [32]byte
		if _, err := io.ReadFull(r, b[:]); err != nil {
			return err
		}

		num, err := conv.BytesToBig(b[:])
		if err != nil {
			return err
		}

		*e = num
	case *uint16:
		var b [2]byte
		if _, err := io.ReadFull(r, b[:]); err != nil {
			return err
		}
		*e = binary.BigEndian.Uint16(b[:])
	case *uint64:
		var b [8]byte
		if _, err := io.ReadFull(r, b[:]); err != nil {
			return err
		}
		*e = binary.BigEndian.Uint64(b[:])
	default:
		return errors.New("reading unknown element type " + reflect.TypeOf(element).String())
	}

	return nil
}

func readByteSliceLike(r io.Reader) ([]byte, error) {
	var l [2]byte
	if _, err := io.ReadFull(r, l[:]); err != nil {
		return nil, err
	}
	byteLen := binary.BigEndian.Uint16(l[:])
	buf := make([]byte, byteLen)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

func writeElement(w io.Writer, element interface{}) error {
	switch e := element.(type) {
	case *crypto.PublicKey:
		if e == nil {
			return errors.New("cannot write nil pubkey")
		}

		var b [33]byte
		serialized := e.SerializeCompressed()
		copy(b[:], serialized)
		if _, err := w.Write(b[:]); err != nil {
			return err
		}
	case *net.TCPAddr:
		WriteTCPAddr(w, e)
	case string:
		strBytes := []byte(e)
		var l [2]byte
		binary.BigEndian.PutUint16(l[:], uint16(len(strBytes)))
		if _, err := w.Write(l[:]); err != nil {
			return err
		}

		if _, err := w.Write(strBytes); err != nil {
			return err
		}
	case [32]byte:
		if _, err := w.Write(e[:]); err != nil {
			return err
		}
	case common.Hash:
		if _, err := w.Write(e[:]); err != nil {
			return err
		}
	case []byte:
		if err := writeByteSliceLike(w, e); err != nil {
			return err
		}
	case crypto.Signature:
		if err := writeByteSliceLike(w, e); err != nil {
			return err
		}
	case *big.Int:
		if _, err := w.Write(conv.BigToBytes(e)); err != nil {
			return err
		}
	case uint16:
		var b [2]byte
		binary.BigEndian.PutUint16(b[:], e)
		if _, err := w.Write(b[:]); err != nil {
			return err
		}
	case uint64:
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], e)
		if _, err := w.Write(b[:]); err != nil {
			return err
		}
	default:
		return errors.New("writing unknown element type " + reflect.TypeOf(element).String())
	}

	return nil
}

func writeByteSliceLike(w io.Writer, thing []byte) error {
	var l [2]byte
	binary.BigEndian.PutUint16(l[:], uint16(len(thing)))
	if _, err := w.Write(l[:]); err != nil {
		return err
	}

	if _, err := w.Write(thing[:]); err != nil {
		return err
	}

	return nil
}

func readElements(r io.Reader, elements ...interface{}) error {
	for _, element := range elements {
		err := readElement(r, element)
		if err != nil {
			return err
		}
	}

	return nil
}

func writeElements(w io.Writer, elements ...interface{}) error {
	for _, element := range elements {
		err := writeElement(w, element)
		if err != nil {
			return err
		}
	}

	return nil
}

func makeEmptyMessage(msgType lnwire.MessageType) (lnwire.Message, error) {
	var msg lnwire.Message

	switch msgType {
	case MsgInit:
		msg = &Init{}
	case lnwire.MsgPing:
		msg = &lnwire.Ping{}
	case lnwire.MsgPong:
		msg = &lnwire.Pong{}
	case MsgOpenChannel:
		msg = &OpenChannel{}
	case MsgAcceptChannel:
		msg = &AcceptChannel{}
	case MsgFundingCreated:
		msg = &FundingCreated{}
	case MsgFundingSigned:
		msg = &FundingSigned{}
	case MsgFundingLocked:
		msg = &FundingLocked{}
	case MsgInitiateSwap:
		msg = &InitiateSwap{}
	case MsgSwapAccepted:
		msg = &SwapAccepted{}
	case MsgInvoiceGenerated:
		msg = &InvoiceGenerated{}
	case MsgInvoiceExecuted:
		msg = &InvoiceExecuted{}
	default:
		return nil, errors.New("unknown message")
	}

	return msg, nil
}

func WriteMessage(w io.Writer, msg lnwire.Message) (int, error) {
	totalBytes := 0

	var prefix [2]byte
	binary.BigEndian.PutUint16(prefix[:], uint16(MsgPrefix))

	n, err := w.Write(prefix[:])

	if err != nil {
		return totalBytes, err
	}

	totalBytes += n

	n, err = lnwire.WriteMessage(w, msg, 0)
	if err != nil {
		return totalBytes, err
	}

	totalBytes += n

	return n, nil
}

func ReadMessage(r io.Reader, pver uint32) (lnwire.Message, error) {
	var pfx [2]byte
	if _, err := io.ReadFull(r, pfx[:]); err != nil {
		return nil, err
	}

	prefix := binary.BigEndian.Uint16(pfx[:])

	if prefix != MsgPrefix {
		return nil, errors.New("invalid prefix")
	}

	var mType [2]byte
	if _, err := io.ReadFull(r, mType[:]); err != nil {
		return nil, err
	}

	msgType := lnwire.MessageType(binary.BigEndian.Uint16(mType[:]))

	msg, err := makeEmptyMessage(msgType)
	if err != nil {
		return nil, err
	}

	if err := msg.Decode(r, pver); err != nil {
		return nil, err
	}

	return msg, nil
}

func MessageName(msgType lnwire.MessageType) string {
	t := msgType.String()
	if t != UnknownMsg {
		return t
	}

	switch msgType {
	case MsgInitiateSwap:
		return "MsgInitiateSwap"
	case MsgSwapAccepted:
		return "MsgSwapAccepted"
	case MsgInvoiceGenerated:
		return "MsgInvoiceGenerated"
	case MsgInvoiceExecuted:
		return "MsgInvoiceExecuted"
	default:
		return UnknownMsg
	}
}
