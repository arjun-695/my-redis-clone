// Individual Client Connection logic
package main

import (
	// "fmt"
	"net"
)

//encapsulates the raw data and the sender (peer)
type Message struct {
	data [] byte 
	peer *Peer // client details 
}

type Peer struct {
	conn  net.Conn
	msgCh chan Message
}

func NewPeer(conn net.Conn, msgCh chan Message) *Peer {
	return &Peer{
		conn:  conn,
		msgCh: msgCh,
	}
}

// to write data back to the client 
func (p *Peer) Send (b []byte) error { //?
	_, err := p.conn.Write(b)
	return err
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
		p.msgCh <- Message{
			data : msgBuf,
			peer : p }//  data flow how it is read and sent in server's main loop
	}
}
