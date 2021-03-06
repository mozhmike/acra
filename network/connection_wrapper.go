package network

import (
	"net"
)

type ConnectionWrapper interface {
	WrapClient(id []byte, conn net.Conn) (net.Conn, error)
	WrapServer(conn net.Conn) (net.Conn, []byte, error) // conn, ClientId, error
}
