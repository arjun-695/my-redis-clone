package main

import (
	"fmt"
	"log"
	"net"
	"os"
	// "time"
)

func main() {
	/*
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
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Server Response: %q\n", string(buf[:n]))

	cmdGet := "*2\r\n$3\r\nGET\r\n$4\r\nname\r\n"
	fmt.Println("Sending: GET name")
	conn.Write([]byte(cmdGet))

	// 5. GET ka response read karein
	n2, _ := conn.Read(buf)
	fmt.Println("Server Reply:", string(buf[:n2]))

	conn.Write([]byte("*2\r\n$3\r\nDEL\r\n$4\r\nname\r\n"))
	n, _ = conn.Read(buf)
	fmt.Printf("   -> DEL Response: %q (Expected :1 means deleted)\n\n", string(buf[:n]))

	fmt.Println("   -> Action: Setting 'token' with 2 seconds TTL...")
	conn.Write([]byte("*5\r\n$3\r\nSET\r\n$5\r\ntoken\r\n$6\r\nsecret\r\n$2\r\nEX\r\n$1\r\n2\r\n"))
	n, _ = conn.Read(buf)
	fmt.Printf("   -> SET EX Response: %q\n", string(buf[:n]))

	conn.Write([]byte("*2\r\n$3\r\nGET\r\n$5\r\ntoken\r\n"))
	n, _ = conn.Read(buf)
	fmt.Printf("   -> Immediate GET: %q (Should contain data)\n", string(buf[:n]))

	fmt.Println("   ... Waiting 3 seconds for expiration ...")
	time.Sleep(3 * time.Second)

	conn.Write([]byte("*2\r\n$3\r\nGET\r\n$5\r\ntoken\r\n"))
	n, _ = conn.Read(buf)
	fmt.Printf("   -> Delayed GET: %q (Expected $-1 meaning NULL)\n", string(buf[:n]))

	fmt.Println("\n All Tests Completed!") */

	if len(os.Args) < 2 {
		fmt.Println("Usage: go run client/main.go [step1 | step2]")
		return
	}

	conn, err := net.Dial("tcp", "localhost:5001")
	if err != nil {
		log.Fatal("Connection Failed:", err)
	}
	defer conn.Close()

	if os.Args[1] == "step1" {
		// Save data
		fmt.Println("--- STEP 1: Saving Data ---")
		sendCommand(conn, "*3\r\n$3\r\nSET\r\n$6\r\nresume\r\n$10\r\nredis_done\r\n")
		fmt.Println("Data saved. Now go to the server terminal, press CTRL+C to kill it.")
		fmt.Println("Then start the server again and run: go run client/main.go step2")
	} else if os.Args[1] == "step2" {
		// Read data back after crash
		fmt.Println("--- STEP 2: Checking Recovery ---")
		sendCommand(conn, "*2\r\n$3\r\nGET\r\n$6\r\nresume\r\n")
	}
}
func sendCommand(conn net.Conn, cmd string) {
	_, err := conn.Write([]byte(cmd))
	if err != nil {
		log.Fatal(err)
	}

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Server says: %q\n", string(buf[:n]))
}