package main

import "log"

// handles configuration and server

func main() {
	cfg := Config{
		ListenAddr: ":5001",
	}
	server := NewServer(cfg)

	log.Fatal(server.Start()) // crash if we can't start the server

}
