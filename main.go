package main

import (
	"fmt"
	"log"
	"log/slog"
	"net"
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
	msgCh chan []byte
	
}

func NewServer(cfg Config) *Server{

	if len(cfg.ListenAddr) == 0 {// why length of litstenaAddr
		cfg.ListenAddr= defaultListenAddr

	}
	return &Server{
		Config: cfg,
		peers: make(map[*Peer]bool),
		addPeerCh: make( chan *Peer),
		quitCh: make( chan struct{}),
		msgCh: make(chan []byte),
	}
}

func(s *Server) Start() error { // why returning error 

	ln, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return err
	}
	s.ln = ln //??

	go s.loop() // why go?
	
	slog.Info("server running", "listenAddr" , s.ListenAddr)//slog.Info := gives certain metadeta instead of just printing a value like time and date of connection establishment 
	return s.acceptLoop()
}

func (s *Server) handleRawMessage(rawMsg []byte) error{ 
	fmt.Print(string(rawMsg))
	return nil
}

func (s *Server) loop() {//explaination
	for {
		select{ // why select and what is select?is this like switch case?

		case rawMsg := <- s.msgCh:
			if err := s.handleRawMessage(rawMsg); err != nil{
				slog.Error("raw message error", "err", err)//slog functions ?
			}
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
peer := NewPeer(conn, s.msgCh)
s.addPeerCh <- peer
slog.Info("new peer connected", "remoteAddr", conn.RemoteAddr())
 if err := peer.readLoop(); err != nil {
	slog.Error("peer read error", "err", err, "remote Addr", conn.RemoteAddr())
 }
}

func main() {
server := NewServer((Config{}))
log.Fatal(server.Start())
}