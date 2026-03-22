package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

const sysInstructn = `You are a translator that converts English into the Redis Serialization Protocol (RESP). 
Our custom database ONLY supports the following 5 commands:
1. SET key value [EX seconds]
2. GET key
3. DEL key
4. VSET key float1 float2 float3...
5. VSEARCH float1 float2 float3... LIMIT n

Rules:
- Output ONLY the raw RESP string. No markdown formatting, no backticks, no explanation.
- Use \r\n for line breaks.

Examples:
Input: "save car as bmw for 10 seconds"
Output: *5\r\n$3\r\nSET\r\n$3\r\ncar\r\n$3\r\nbmw\r\n$2\r\nEX\r\n$2\r\n10\r\n

Input: "find 2 vectors similar to 0.5 0.2"
Output: *5\r\n$7\r\nVSEARCH\r\n$3\r\n0.5\r\n$3\r\n0.2\r\n$5\r\nLIMIT\r\n$1\r\n2\r\n`

func main() {
	apiKey := os.Getenv("Gemini_API_KEY")
	if apiKey == "" {
		log.Fatal("Error: Gemini_API_KEY environment variable is not set. \n ")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatal("Failed to create Gemini client", err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-2.5-flash")
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(sysInstructn)},
	}

	fmt.Println("Welcome to AI powered Redis CLI ")
	fmt.Println("Type 'exit' to quit.")

	conn, err := net.Dial("tcp", "localhost:5001")
	if err != nil {
		log.Fatal("Could not connect to custom Redis Server:", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Ai-CLI >")
		userInput, _ := reader.ReadString('\n')
		userInput = strings.TrimSpace(userInput)

		if userInput == "exit" {
			break
		}
		if userInput == "" {
			continue
		}

		//ask gemini to translate
		respCommand := askGemini(ctx, model, userInput)

		respCommand = sanitizeLLMOutput(respCommand)

		fmt.Printf("\n[Gemini Translated to RESP] -> %q\n", respCommand)

		// sending raw RESP to GO Server

		_, err := conn.Write([]byte(respCommand))
		if err != nil {
			fmt.Println("Error sending to Server:", err)
			continue
		}

		// Read Serverr response
		buf := make([]byte, 2048) //increased buffer for vector serach results
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading from server: ", err)
			continue
		}
		fmt.Printf("Server: %q\n\n", string(buf[:n]))
	}
}

func askGemini(ctx context.Context, model *genai.GenerativeModel, prompt string) string{
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return fmt.Sprintf("-ERR gemini Api failed: %v\r\n", err)
	}
	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		part := resp.Candidates[0].Content.Parts[0]
		if text, ok := part.(genai.Text); ok {
			return string(text)
		}
	}
return "-ERR NO response from Gemini\r\n"
}

// function to fix formatting issues for raw TCP streaming 
func sanitizeLLMOutput(output string) string{
	output = strings.TrimSpace(output)

	//remove markdown backticks
	output = strings.TrimPrefix(output, "```resp")
	output = strings.TrimPrefix(output, "```text")
	output = strings.TrimPrefix(output, "```")
	output = strings.TrimSuffix(output, "```")
	output = strings.TrimSpace(output)

	// might return literal "\r\n" characters as string instead of actual carriage returns

	output = strings.ReplaceAll(output, "\\r\\n", "\r\n")

	return output
}
