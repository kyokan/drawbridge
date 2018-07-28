package wire

import (
	"testing"
	"math/rand"
	"net"
	"encoding/binary"
	"bytes"
	"github.com/stretchr/testify/assert"
)

func randTCP4Addr(r *rand.Rand) (*net.TCPAddr, error) {
	var ip [4]byte
	if _, err := r.Read(ip[:]); err != nil {
		return nil, err
	}

	var port [2]byte
	if _, err := r.Read(port[:]); err != nil {
		return nil, err
	}

	addrIP := net.IP(ip[:])
	addrPort := int(binary.BigEndian.Uint16(port[:]))

	return &net.TCPAddr{IP: addrIP, Port: addrPort}, nil
}

func randTCP6Addr(r *rand.Rand) (*net.TCPAddr, error) {
	var ip [16]byte
	if _, err := r.Read(ip[:]); err != nil {
		return nil, err
	}

	var port [2]byte
	if _, err := r.Read(port[:]); err != nil {
		return nil, err
	}

	addrIP := net.IP(ip[:])
	addrPort := int(binary.BigEndian.Uint16(port[:]))

	return &net.TCPAddr{IP: addrIP, Port: addrPort}, nil
}

func TestReadWriteTCP4Addr(t *testing.T) {
	var buf bytes.Buffer
	r := rand.New(rand.NewSource(10))
	addr, err := randTCP4Addr(r)
	if err != nil {
		t.Fatalf(err.Error())
	}

	out := &net.TCPAddr{}
	err = WriteTCPAddr(&buf, addr)
	if err != nil {
		t.Fatalf(err.Error())
	}
	_, err = ReadTCPAddr(&buf, out)
	if err != nil {
		t.Fatalf(err.Error())
	}
	assert.Equal(t, addr, out)
}

func TestReadWriteTCP6Addr(t *testing.T) {
	var buf bytes.Buffer
	r := rand.New(rand.NewSource(10))
	addr, err := randTCP6Addr(r)
	if err != nil {
		t.Fatalf(err.Error())
	}

	out := &net.TCPAddr{}
	err = WriteTCPAddr(&buf, addr)
	if err != nil {
		t.Fatalf(err.Error())
	}
	_, err = ReadTCPAddr(&buf, out)
	if err != nil {
		t.Fatalf(err.Error())
	}
	assert.Equal(t, addr, out)
}