// Individual Client Connection logic
package main

import (
	// "fmt"
	"net"
)

type Peer struct {
	conn  net.Conn
	msgCh chan []byte
}

func NewPeer(conn net.Conn, msgCh chan []byte) *Peer {
	return &Peer{
		conn:  conn,
		msgCh: msgCh,
	}
}
func (p *Peer) readLoop() error {
	buf := make([]byte, 1024) // Buffer size 1024 bytes
	for {
		n, err := p.conn.Read(buf)
		if err != nil {
			return err
		}
		//
		msgBuf := make([]byte, n)
		copy(msgBuf, buf[:n])

		// sending data to main loop of server
		p.msgCh <- msgBuf //  data flow how it is read and sent in server's main loop
	}
}
