package main

import (
	// "fmt"
	"net"
	"log/slog"
	"log"
)

const defaultListenAddr = ":5001"

type Config struct {
	ListenAddr string 
}

type Server struct {
	Config
	ln      net.Listener
	peers   map[*Peer]bool
	addPeerCh chan *Peer // what is chan
	quitCh chan struct{}
}

func NewServer(cfg Config) *Server{

	if len(cfg.ListenAddr) == 0 {
		cfg.ListenAddr= defaultListenAddr

	}
	return &Server{
		Config: cfg,
		peers: make(map[*Peer]bool),
		addPeerCh: make( chan *Peer),
		quitCh: make( chan struct{}),

	}
}

func(s *Server) Start() error { // why returning error 

	ln, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return err
	}
	s.ln = ln

	go s.loop()
	
	slog.Info("server running", "listenAddr" , s.ListenAddr)
	return s.acceptLoop()
}

func (s *Server) loop() {//explaination
	for {
		select{ 
		case <- s.quitCh:
			return  
		case peer := <- s.addPeerCh: 
		s.peers[peer] = true

	}
}
}

func (s *Server) acceptLoop() error { // why returning error? why pointers to server??
	for{
		conn, err := s.ln.Accept()
		if err != nil {
			slog.Error("accept error", "err", err)
			continue
		}
		go s.handleConn( conn ) //why go? maybe to handle multiple connections in the backend #Concurrency
	}
}

func (s *Server) handleConn(conn net.Conn) {
peer := NewPeer(conn)
s.addPeerCh <- peer

 peer.readLoop()
}

func main() {
server := NewServer((Config{}))
log.Fatal(server.Start())
}