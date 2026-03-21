// Core server logic, accepts connection and manages Peers
package main

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"
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
	delPeerCh chan *Peer    // handles disconnection and prevents memory leak
	quitCh    chan struct{} // Sends signal to quit server
	msgCh     chan Message  // saves data and client information coming from clients
	kv        *KV           //stores key value pair
	aof       *AOF
}

func NewServer(cfg Config) *Server {

	if len(cfg.ListenAddr) == 0 { // ListenAddr: IP + Port
		cfg.ListenAddr = defaultListenAddr
	}

	aof, err := NewAOF("dump.aof")
	if err != nil {
		slog.Error("AOF creation Error", "err", err)
		os.Exit(1)
	}
	s := &Server{
		Config:    cfg,
		peers:     make(map[*Peer]bool),
		addPeerCh: make(chan *Peer),
		delPeerCh: make(chan *Peer),
		quitCh:    make(chan struct{}),
		msgCh:     make(chan Message),
		kv:        NewKV(),
		aof: 	   aof,
	}
	err = s.aof.ReadExisting(func(cmd Command) {
		s.executeCommand(cmd, nil) // peer = nil because we are reading from file
	})

	if err != nil{
		slog.Warn("Could not load AOF (might be a new Instance)", "err", err)
	} else {
		slog.Info("Persistace data loaded successfully from dump.aof")

	}

	return s
}

func (s *Server) Start() error {
	//TCP listener

	ln, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return err
	}
	s.ln = ln 

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

		case peer := <-s.delPeerCh:
			delete(s.peers, peer)
			slog.Info("peer disconnected", "active_connection", len(s.peers))
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

	peer := NewPeer(conn, s.msgCh, s.delPeerCh)
	s.addPeerCh <- peer
	slog.Info("new peer connected", "remoteAddr", conn.RemoteAddr())

	if err := peer.readLoop(); err != nil {
		slog.Error("peer read error", "err", err, "remote Addr", conn.RemoteAddr())
	}
}

func (s *Server) handleMessage(msg Message) error {
	switch v := msg.cmd.(type){
	case SetCommand: 
		if v.ex > 0 {
			s.aof.Write(SerializeCommand("SET", v.key, v.val, "EX", strconv.Itoa(v.ex)))

		} else {
			s.aof.Write(SerializeCommand("SET", v.key, v.val))
		}
	case DelCommand:
		s.aof.Write(SerializeCommand("DEL", v.key))
	}

	return s.executeCommand(msg.cmd, msg.peer)
}

func (s *Server) executeCommand(cmd Command, peer *Peer) error {

	switch v := cmd.(type) { // .(type)-> "Type Switch" or "Type Assertion" checks the type of struct in cmd
	case SetCommand:

		s.kv.Set([]byte(v.key), []byte(v.val))

		if v.ex > 0 {
			//time.AfterFunc schedules a function to run after a duration
			//runs in it's own go routine so it doesn't block the server
			time.AfterFunc(time.Duration(v.ex)*time.Second, func() {
				s.kv.Delete([]byte(v.key))
			})
		}

		if peer != nil {
			go peer.Send([]byte("+Ok\r\n")) //server loop will not wait for any slow client because of "go"
		}
		// time.AfterFunc is a neat and simple approach
		//but wht if:
		// there are 10 mil keys with diff expiring time? this will waste RAM and CPU scheduling time
		//Better Approach
		// Passive: when a client tries to GET a key, check if it is expired, if yes then delete it and return nil
		// Active: A background threat checks 20 random keys with an associated TTL, 10 times each second and deletes the expired ones
	case GetCommand:

		val, ok := s.kv.Get([]byte(v.key))
		if peer !=nil {
			if !ok {
			go peer.Send([]byte("$-1\r\n"))
		} else {
			respMsg := fmt.Sprintf("$%d\r\n%s\r\n", len(val), val)
			go peer.Send([]byte(respMsg))
		}}

	case DelCommand:
		deleted := s.kv.Delete([]byte(v.key))
		if peer != nil {
			resp := ":0\r\n"
		if deleted {
			resp = ":1\r\n"
		}
		go peer.Send([]byte(resp))}
	}

	return nil
}
