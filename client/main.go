package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:5001") //what does dial function do
	if err != nil {
		log.Fatal("Connection Failed:", err)
	}
	defer conn.Close() //to close the connection

	// sending Command in RESP format
	cmd := "*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$5\r\nRahul\r\n"

	fmt.Println("Sending command:SET name Rahul")
	_, err = conn.Write([]byte(cmd))
	if err != nil {
		log.Fatal(err)
	}

	// Reading Server's Reply
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil{
		log.Fatal(err)
	}

	fmt.Printf("Server Response: %q\n", string(buf[:n]))
}
