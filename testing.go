package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type RequestPayload struct {
	Model       string  `json:"model"`
	Prompt      string  `json:"prompt"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
	Stream      bool    `json:"stream"`
}

type Choice struct {
	Delta struct {
		Content string `json:"content"`
	} `json:"delta"`
}

type ResponseData struct {
	Choices []Choice `json:"choices"`
}

func processStreaming(resp *http.Response, controller chan string) {
	defer close(controller)

	counter := 0
	streamReader := bufio.NewReader(resp.Body)
	for {
		line, err := streamReader.ReadString('\n')
		if err != nil {
			break
		}

		line = strings.TrimSpace(line)
		fmt.Println(line)

		if line == "data: [DONE]" {
			controller <- "[DONE]"
			break
		}

		if strings.HasPrefix(line, "data:") {
			payload := strings.TrimPrefix(line, "data:")

			var responseData ResponseData
			err = json.Unmarshal([]byte(payload), &responseData)
			if err != nil {
				fmt.Println("Error parsing data:", err)
				continue
			}

			content := responseData.Choices[0].Delta.Content
			fmt.Println("Content:", content)
			if counter < 2 && strings.Count(content, "\n") > 0 {
				continue
			}

			controller <- content
			counter++
		}
	}
}

func checkEndpoint() {
	apiEndpoint := "https://api.openai.com/v1/chat/completions"

	requestPayload := Payload{
		Model:       "gpt-3.5-turbo",
		Messages
		MaxTokens:   10,
		Temperature: 0.7,
		Stream:      true,
	}

	reqBytes, err := json.Marshal(requestPayload)
	if err != nil {
		panic(err)
	}

	reqBody := bytes.NewBuffer(reqBytes)

	req, err := http.NewRequest("POST", apiEndpoint, reqBody)
	if err != nil {
		panic(err)
	}

	apiKey := os.Getenv("OPENAI_API_KEY")

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", apiKey)) // Replace <YOUR-API-KEY> with your actual API key

	client := &http.Client{
		Timeout: time.Second * 60,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		controller := make(chan string)
		go processStreaming(resp, controller)

		for {
			msg, ok := <-controller
			if !ok {
				break
			}

			fmt.Println("P", msg)

			if msg == "[DONE]" {
				fmt.Println("Stream finished.")
			} else {
				fmt.Println("Data received:", msg)
			}
		}
	} else {
		fmt.Println("Error:", resp.StatusCode)
	}
}
