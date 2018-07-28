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
)

const MsgPrefix = 0xbeef

const (
	MsgInit          lnwire.MessageType = 16
	MsgOpenChannel                      = 32
	MsgAcceptChannel                    = 33
	MsgFundingCreated                   = 34
	MsgInitiateSwap                     = 900
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
	default:
		return errors.New("unknown element type")
	}

	return nil
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
		var l [2] byte
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
	default:
		return errors.New("unknown element type")
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
