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

type VSetCommand struct{
	key string
	vector []float64
}

type VSearchCommand struct {
	vector []float64
	limit int
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
	
		case "VSET":
			if len(args) < 3 {
				return nil, fmt.Errorf("VSET requires a key and at least one vector value")
			}
			key := args[1].String()
			var vec []float64
			for i := 2; i< len(args); i++ {
				val, err := strconv.ParseFloat(args[i].String(),64)
				if err != nil {
					return nil, fmt.Errorf("invalid vector float: %v ", err)
				}
				vec = append(vec, val)
			}
			return VSetCommand{key: key, vector: vec}, nil

		case "VSEARCH" :
			if len(args) < 4 { // VSEARCH + values + LIMIT + n
				return nil , fmt.Errorf("VSEARCH requires vector, Limit, and a number") 
			}

			limit := 1 //default
			var vec []float64

			//Find Limit Keyword
			limitIdx := -1
			for i :=1; i< len(args); i++{
				if strings.ToUpper(args[i].String()) == "LIMIT"{
					limitIdx = i
					break
				}
			}

			if limitIdx != -1 {
				if limitIdx == len(args)-1 {
					return nil, fmt.Errorf("Syntax error, expected LIMIT [number]")
				}
				limitStr := args[limitIdx+1].String()
				parsedLimit, err := strconv.Atoi(limitStr)
				if err == nil {
					limit = parsedLimit
				}
			} else {
				limitIdx = len(args)
			}

			for i := 1; i < limitIdx; i++ {
				val, err := strconv.ParseFloat(args[i].String(), 64)
				if err != nil {
					return nil, fmt.Errorf("invalid vector float: %v", err)
				}
				vec = append(vec, val)
			}
			return VSearchCommand{vector: vec, limit: limit}, nil
	}
	return nil, fmt.Errorf("unknown Command")
}
