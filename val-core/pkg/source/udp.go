package source

import (
	"fmt"
	"net"
)

type UDPConnection struct {
	Conn *net.UDPConn
	Port int
}

func NewUDP(port int) (*UDPConnection, error) {
	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP("0.0.0.0"),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		return nil, fmt.Errorf("failed to bind UDP port %d - %w", port, err)
	}

	return &UDPConnection{
		Conn: conn,
		Port: port,
	}, nil
}

func (u *UDPConnection) Close() error {
	if u.Conn == nil {
		return nil
	}

	return u.Conn.Close()
}
