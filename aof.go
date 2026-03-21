package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/tidwall/resp"
)

// Append only file handles persistence to disk
type AOF struct {
	file *os.File
	mu   sync.Mutex
}

func NewAOF(path string) (*AOF, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	aof := &AOF{file: f}

	//background fsync: forces OS to write to disk every second to prevent data loss
	go func() {
		for {
			time.Sleep(time.Second)
			aof.mu.Lock()
			aof.file.Sync()
			aof.mu.Unlock()
		}
	}()
	return aof, nil

}

func (aof *AOF) Write(raw []byte) error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	_, err := aof.file.Write(raw)
	return err
}

func (aof *AOF) Close() error {
	aof.mu.Lock()
	defer aof.mu.Unlock()
	aof.file.Sync()
	return aof.file.Close()
}

// ReadExisting replays the file on startup using a callback function
func (aof *AOF) ReadExisting(callback func(Command)) error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	aof.file.Seek(0, 0)
	rd := resp.NewReader(aof.file)

	for {
		v, _, err := rd.ReadValue()
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		cmd, err := parseCommand(v.Array())
		if err == nil {
			callback(cmd)
		}
	}
	return nil
}

//SerializeCommand perfectly formats raw Strings back to Redis RESP format

func SerializeCommand(args ...string) []byte {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("*%d\r\n", len(args)))
	for _, arg := range args {
		buf.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg))
	}
	return buf.Bytes()
}
