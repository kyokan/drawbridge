package wire

import (
	"io"
	"net"
	"github.com/go-errors/errors"
	"encoding/binary"
)

const tcp4Addr = 1

const tcp6Addr = 2

func WriteTCPAddr(w io.Writer, addr *net.TCPAddr) error {
	if addr == nil {
		return errors.New("cannot write nil TCPAddr")
	}

	if addr.IP.To4() != nil {
		var descriptor [1]byte
		descriptor[0] = uint8(tcp4Addr)
		if _, err := w.Write(descriptor[:]); err != nil {
			return err
		}

		var ip [4]byte
		copy(ip[:], addr.IP.To4())
		if _, err := w.Write(ip[:]); err != nil {
			return err
		}
	} else {
		var descriptor [1]byte
		descriptor[0] = uint8(tcp6Addr)
		if _, err := w.Write(descriptor[:]); err != nil {
			return err
		}
		var ip [16]byte
		copy(ip[:], addr.IP.To16())
		if _, err := w.Write(ip[:]); err != nil {
			return err
		}
	}

	var port [2]byte
	binary.BigEndian.PutUint16(port[:], uint16(addr.Port))
	if _, err := w.Write(port[:]); err != nil {
		return err
	}

	return nil
}

func ReadTCPAddr(r io.Reader, addr *net.TCPAddr) (uint, error) {
	bytesRead := uint(0)

	var descriptor [1]byte
	if _, err := io.ReadFull(r, descriptor[:]); err != nil {
		return bytesRead, err
	}

	addrType := uint8(descriptor[0])

	if addrType == tcp4Addr {
		var ip [4]byte
		if _, err := io.ReadFull(r, ip[:]); err != nil {
			return bytesRead, err
		}
		bytesRead += 4

		var port [2]byte
		if _, err := io.ReadFull(r, port[:]); err != nil {
			return bytesRead, err
		}
		bytesRead += 2

		addr.IP = net.IP(ip[:])
		addr.Port = int(binary.BigEndian.Uint16(port[:]))
	} else if addrType == tcp6Addr {
		var ip [16]byte
		if _, err := io.ReadFull(r, ip[:]); err != nil {
			return bytesRead, err
		}
		bytesRead += 16

		var port [2]byte
		if _, err := io.ReadFull(r, port[:]); err != nil {
			return bytesRead, err
		}
		bytesRead += 2

		addr.IP = net.IP(ip[:])
		addr.Port = int(binary.BigEndian.Uint16(port[:]))
	} else {
		return bytesRead, errors.New("unknown address type")
	}

	return bytesRead, nil
}