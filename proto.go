// RESP protocol parsing logic
package main

import (
	"bytes"
	"fmt"
	"io"
	"strconv" //for String to integer conversion

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
	ex int// expiration in seconds, 0 means no expiration
}

type GetCommand struct {
	key string
}

type DelCommand struct {
	key string
}

func parseCommand(raw string) (Command, error) {
	rd := resp.NewReader(bytes.NewBufferString(raw))

	for {
		v, _, err := rd.ReadValue() //rd - special library "tidwall/resp" ; ReadValue() reads the string sent by the client and returns an array 
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// if REDIS client has returned an array (eg: ["SET","key", "val"])
		if v.Type() == resp.Array {
			args := v.Array() // RESP data structure -> Go native array/slice
			if len(args) == 0 {
				continue
			}

			// getting command
			cmdName := args[0].String()

			switch cmdName {
			case "SET":
				if len(args) != 3 && len(args) != 5 {
					return nil, fmt.Errorf("Invalid SET Command")
				}
				cmd:= SetCommand{
					key: args[1].String(),
					val: args[2].String(),
				}

				if len(args) == 5{
					if args[3].String() != "EX" {
						return nil, fmt.Errorf("expected EX argument")
					}
					ex, err := strconv.Atoi(args[4].String())
					if err != nil{
						return nil, fmt.Errorf("invalid EX duration")
					}
					cmd.ex = ex
				}
				return cmd, nil

			case "GET":
				if len(args) != 2 {
					return nil, fmt.Errorf("Invalid GET Command")
				}
				return GetCommand{
					key: args[1].String(), // client wants to get the value of key passed 

				}, nil

			case "DEL":
				if len(args) != 2 {
					return nil, fmt.Errorf("Invalid DEL command")
				}
				return DelCommand{
					key: args[1].String(),
				}, nil
			}
		}
	}
	return nil, fmt.Errorf("unknown Command")
}
