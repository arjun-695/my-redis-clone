// RESP protocol parsing logic
package main

import (
	"fmt"
	"strconv" //for String to integer conversion
	"strings"

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
	ex       int // expiration in seconds, 0 means no expiration
}

type GetCommand struct {
	key string
}

type DelCommand struct {
	key string
}

func parseCommand(args []resp.Value) (Command, error) {

	if len(args) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	// getting command
	cmdName := strings.ToUpper(args[0].String())

	switch cmdName {
	case "SET":
		if len(args) != 3 && len(args) != 5 {
			return nil, fmt.Errorf("Invalid SET Command")
		}
		cmd := SetCommand{
			key: args[1].String(),
			val: args[2].String(),
		}
		if len(args) == 5 {
			if args[3].String() != "EX" {
				return nil, fmt.Errorf("expected EX argument")
			}
			ex, err := strconv.Atoi(args[4].String())
			if err != nil {
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
			key: args[1].String(), // client wants to get thevalue of key passed
		}, nil

	case "DEL":
		if len(args) != 2 {
			return nil, fmt.Errorf("Invalid DEL command")
		}
		return DelCommand{
			key: args[1].String(),
		}, nil
	}
	return nil, fmt.Errorf("unknown Command")
}
