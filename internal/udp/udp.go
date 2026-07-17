package udp

import (
	"net"
	"time"
)

type Conn interface {
	Write(b []byte) (int, error)
	Read(b []byte) (int, error)
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
	Close() error
}

func NewConn(addr *net.UDPAddr) (Conn, error) {
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func NewFakeUDPConn() Conn {
	return &fakeUDPConn{}
}

type fakeUDPConn struct{}

func (f *fakeUDPConn) Write(b []byte) (int, error) {
	return len(b), nil
}

func (f *fakeUDPConn) Read(b []byte) (int, error) {
	return 0, nil
}

func (f *fakeUDPConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (f *fakeUDPConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func (f *fakeUDPConn) Close() error {
	return nil
}
