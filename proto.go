// RESP protocol parsing logic
package main

import (
	"bytes"
	"fmt"
	"io"

	"github.com/tidwall/resp"
)

// *3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n :

// *3: Array of 3 elements.

// $3: Bulk string of length 3.

// SET, foo, bar: The actual data.

type Command interface {
	// allows us to define different Redis commands (GET, SET..etc )
}

type SetCommand struct {
	key, val string
}

type GetCommand struct {
	key string
}

func parseCommand(raw string) (Command, error) {
	rd := resp.NewReader(bytes.NewBufferString(raw))

	for {
		v, _, err := rd.ReadValue() //??
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// if REDIS client has returned an array (eg: ["SET","key", "val"])
		if v.Type() == resp.Array {
			args := v.Array() // what does array function do
			if len(args) == 0 {
				continue
			}

			// getting command
			cmdName := args[0].String()

			switch cmdName {
			case "SET":
				if len(args) != 3 {
					return nil, fmt.Errorf("Invalid SET Command")
				}
				return SetCommand{
					key: args[1].String(),
					val: args[2].String(),
				}, nil

			case "GET":
				if len(args) != 2 {
					return nil, fmt.Errorf("Invalid GET Command")
				}
				return GetCommand{
					key: args[1].String(), // why returning key and not value

				}, nil
			}
		}
	}
	return nil, fmt.Errorf("unknown Command")
}
