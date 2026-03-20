// Core server logic, accepts connection and manages Peers
package main

import (
	"fmt"
	"log/slog"
	"net"
	"time"
)

const defaultListenAddr = ":5001"

type Config struct {
	ListenAddr string
}

type Server struct {
	Config
	ln        net.Listener
	peers     map[*Peer]bool
	addPeerCh chan *Peer    // Channels Pipeline to send data from one go routine to another safely...
	quitCh    chan struct{} // Sends signal to quit server
	msgCh     chan Message  // saves data and client information coming from clients
	kv        *KV           //stores key value pair

}

func NewServer(cfg Config) *Server {

	if len(cfg.ListenAddr) == 0 { // ListenAddr: IP + Port
		cfg.ListenAddr = defaultListenAddr  

	}
	return &Server{
		Config:    cfg,
		peers:     make(map[*Peer]bool),
		addPeerCh: make(chan *Peer),
		quitCh:    make(chan struct{}),
		msgCh:     make(chan Message),
		kv:        NewKV(),
	}
}

func (s *Server) Start() error {
	//TCP listener

	ln, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return err
	}
	s.ln = ln //??

	go s.loop() // goroutine; runs in background w/o blocking the main thread

	slog.Info("server running", "listenAddr", s.ListenAddr) //slog.Info := gives certain metadeta instead of just printing a value like time and date of connection establishment

	// acceptLoop called in the end because it is a blocking function( it has a for loop )
	return s.acceptLoop() // for accepting the connection
}

func (s *Server) loop() { //explaination
	for { // infinite loop to listen infinitely

		select { // select : Switch case for channels, thread safe that is why no need for locks while updating maps

		case rawMsg := <-s.msgCh:
			if err := s.handleMessage(rawMsg); err != nil {
				slog.Error("raw message error", "err", err) 
			}

		case <-s.quitCh:
			return

		case peer := <-s.addPeerCh:
			s.peers[peer] = true
			slog.Info("peer added to internal map", "remoteAddr", peer.conn.RemoteAddr())

		}
	}
}

func (s *Server) acceptLoop() error { // why returning error? why pointers to server??
	for {
		//New Connection acceptance
		conn, err := s.ln.Accept()
		if err != nil {
			slog.Error("accept error", "err", err)
			continue
		}

		// starting a new go routine for every connection
		// this way the server can handle multiple clients at the same time (Concurrecny)
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {

	peer := NewPeer(conn, s.msgCh)
	s.addPeerCh <- peer
	slog.Info("new peer connected", "remoteAddr", conn.RemoteAddr())

	if err := peer.readLoop(); err != nil {
		slog.Error("peer read error", "err", err, "remote Addr", conn.RemoteAddr())
	}
}

func (s *Server) handleMessage(msg Message) error {
	cmd, err := parseCommand(string(msg.data)) //reflect on string since msg.data is []byte ; returns command which is a empty interface 
	if err != nil {
		msg.peer.Send([]byte(fmt.Sprintf("-ERR %s \r\n", err.Error()))) // msg.peer -> sending client error info; Sprintf -> formatted error message to convert in byte and send it to client.
	}

	switch v := cmd.(type) { // .(type)-> "Type Switch" or "Type Assertion" checks the type of struct in cmd 
	case SetCommand:

		s.kv.Set([]byte(v.key), []byte(v.val))

		if v.ex > 0 {
			//time.AfterFunc schedules a function to run after a duration
			//runs in it's own go routine so it doesn't block the server 
			time.AfterFunc(time.Duration(v.ex)*time.Second, func(){
				s.kv.Delete([]byte(v.key))
			})
		}
		msg.peer.Send([]byte("+Ok\r\n"))
		// time.AfterFunc is a neat and simple approach
		//but wht if:
		// there are 10 mil keys with diff expiring time? this will waste RAM and CPU scheduling time
		//Better Approach
		// Passive: when a client tries to GET a key, check if it is expired, if yes then delete it and return nil
		// Active: A background threat checks 20 random keys with an associated TTL, 10 times each second and deletes the expired ones  
	case GetCommand:

		val, ok := s.kv.Get([]byte(v.key))
		if !ok {
			msg.peer.Send([]byte("$01\r\n"))
		} else {
			respMsg := fmt.Sprintf("$%d\r\n%s\r\n", len(val), val)
			msg.peer.Send([]byte(respMsg))
		}

	case DelCommand:
		deleted := s.kv.Delete([]byte(v.key))
		resp := ":0\r\n"
		if deleted {
			resp = ":1\r\n"
		}
		msg.peer.Send([]byte(resp))
	}

	return nil

}
