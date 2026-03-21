// Individual Client Connection logic
package main

import (
	// "fmt"
	"net"

	"github.com/tidwall/resp"
)

// encapsulates the raw data and the sender (peer)
type Message struct {
	cmd  Command
	peer *Peer // client details
}

type Peer struct {
	conn  net.Conn
	msgCh chan Message
	delCh chan *Peer //
}

func NewPeer(conn net.Conn, msgCh chan Message, delCh chan *Peer) *Peer {
	return &Peer{
		conn:  conn,
		msgCh: msgCh,
		delCh: delCh,
	}
}

// to write data back to the client
func (p *Peer) Send(b []byte) error { //?
	_, err := p.conn.Write(b)
	return err
}
func (p *Peer) readLoop() error {
	defer func() {
		p.delCh <- p
		p.conn.Close() //closing the socket
	}()
	rd := resp.NewReader(p.conn)

	for {
		v, _, err := rd.ReadValue()
		if err != nil {
			return err //client disconnected or network error
		}

		if v.Type() == resp.Array {
			cmd, err := parseCommand(v.Array())
			if err != nil {
				p.Send( []byte("-ERR " + err.Error() + "\r\n"))
				continue
			} 
		
		// sending data to main loop of server
		p.msgCh <- Message{
			cmd: cmd,
			peer: p,
			} //  data flow how it is read and sent in server's main loop
	}
}
}
