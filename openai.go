package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/r3labs/sse"
)

type Payload struct {
	Model       string    `json:"model"`
	Prompt      string    `json:"prompt,omitempty"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float32   `json:"temperature"`
	Stop        []string  `json:"stop,omitempty"`
	Messages    []Message `json:"messages"`
	Stream      bool      `json:"stream"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Response struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
		Index        int    `json:"index"`
	} `json:"choices"`
}

type OpenAIClient struct {
	apiKey    string
	url       string
	model     string
	maxTokens int

	messages []Message

	httpClient *http.Client
}

func promptForModel(model string) []Message {
	switch model {
	case "gpt-4":
		return []Message{
			{Role: "system", Content: "You are a terminal assistant. Turn the natural language instructions into a terminal command. By default always only output code, and in a code block. However, if the user is clearly asking a question then answer it very briefly and well."},
			{Role: "user", Content: "print hi"},
			{Role: "assistant", Content: "```bash\necho \"hi\"\n```"},
		}
	}
	// default for gpt-3.5-turbo
	return []Message{
		{Role: "system", Content: "You are a terminal assistant. Turn the natural language instructions into a terminal command. By default always only output code, and in a code block. DO NOT OUTPUT ADDITIONAL REMARKS ABOUT THE CODE YOU OUTPUT. Do not repeat the question the users asks. Do not add explanations for your code. Do not output any non-code words at all. Just output the code. Short is better. However, if the user is clearly asking a general question then answer it very briefly and well."},
		{Role: "user", Content: "get the current time from some website"},
		{Role: "assistant", Content: "```bash\ncurl -s http://worldtimeapi.org/api/ip | jq '.datetime'\n```"},
		{Role: "user", Content: "print hi"},
		{Role: "assistant", Content: "```bash\necho \"hi\"\n```"},
	}
}

func NewClient(apiKey string, modelOverride string) *OpenAIClient {
	model := "gpt-3.5-turbo"
	if modelOverride != "" {
		model = modelOverride
	}

	return &OpenAIClient{
		apiKey:    apiKey,
		url:       "https://api.openai.com/v1/chat/completions",
		model:     model,
		maxTokens: 256,

		messages: promptForModel(model),

		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

func (c *OpenAIClient) Query(query string) (string, error) {
	messages := c.messages
	messages = append(messages, Message{Role: "user", Content: query})

	payload := Payload{
		Model:       c.model,
		Messages:    messages,
		Temperature: 0,
		MaxTokens:   c.maxTokens,
		Stream:      true,
	}
	// message, err := c.call(payload)
	message, err := c.callStream(payload)
	if err != nil {
		return "", err
	}
	c.messages = append(c.messages, message)
	return message.Content, nil
}

func (c *OpenAIClient) call(payload Payload) (Message, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return Message{}, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", c.url, bytes.NewReader(payloadBytes))
	if err != nil {
		return Message{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Message{}, fmt.Errorf("failed to make the API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Message{}, fmt.Errorf("API request failed: %s", resp.Status)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Message{}, fmt.Errorf("failed to read response body: %w", err)
	}

	var response Response
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return Message{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(response.Choices) == 0 {
		return Message{}, fmt.Errorf("no completions found in response")
	}

	return response.Choices[0].Message, nil
}

func (c *OpenAIClient) callStream(payload Payload) (Message, error) {
	fmt.Println("callStream")
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return Message{}, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", c.url, bytes.NewReader(payloadBytes))
	if err != nil {
		return Message{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Accept", "text/event-stream")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Message{}, fmt.Errorf("failed to make the API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		client := sse.NewClient(c.url)
		err = client.SubscribeRaw(func(msg *sse.Event) {
			responseText := string(msg.Data)

			if responseText == "[DONE]" {
				fmt.Println("Stream finished.")
				return
			}

			fmt.Println("Event received:", responseText)
		})

		if err != nil {
			panic(err)
		}
	} else {
		responseBody, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("Error:", resp.StatusCode, string(responseBody))
	}

	fmt.Println(req)

	// client := sse.NewClient(c.url)
	// client.Connection = c.httpClient
	// fmt.Println("subscribing...")
	// err = client.SubscribeRaw(func(msg *sse.Event) {
	// 	// Got some data!
	// 	fmt.Println("dat!")
	// 	fmt.Println(msg.Data)

	// })
	// if err != nil {
	// 	return Message{}, fmt.Errorf("failed to subscribe: %w", err)
	// }
	// fmt.Println("subscribed")
	return Message{}, nil
}

// func (openaiClient *OpenAIClient) handleSSEEvents(callback EventCallback, w http.ResponseWriter, r *http.Request) {
// 	client := sse.NewClient("https://api.openai.com/v1/chat/completions")

// 	payload := Payload{
// 		Model:       openaiClient.model,
// 		Messages:    openaiClient.messages,
// 		Temperature: 0,
// 		MaxTokens:   openaiClient.maxTokens,
// 		Stop:        []string{"[DONE]"},
// 	}

// 	payloadBytes, err := json.Marshal(payload)
// 	if err != nil {
// 		http.Error(w, "Error marshaling payload", http.StatusInternalServerError)
// 		return
// 	}

// 	req, err := http.NewRequest("POST", openaiClient.url, bytes.NewReader(payloadBytes))
// 	if err != nil {
// 		http.Error(w, "Error creating request", http.StatusInternalServerError)
// 		return
// 	}

// 	req.Header.Set("Content-Type", "application/json")
// 	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", openaiClient.apiKey))
// 	req.Header.Set("Accept", "text/event-stream")

// 	// Set the new HTTP client with the Transport
// 	client.Client.Transport = openaiClient.httpClient.Transport

// 	w.Header().Set("Cache-Control", "no-cache")
// 	w.Header().Set("Content-Type", "text/event-stream")
// 	w.Header().Set("Access-Control-Allow-Origin", "*")
// 	w.Header().Set("Connection", "keep-alive")
// 	w.WriteHeader(http.StatusOK)

// 	client.Subscribe("completion", req.URL.String(), req.Headers, func(msg *sse.Event) {
// 		if string(msg.Data) == "[DONE]" {
// 			return
// 		}

// 		var parsed Response
// 		if err := json.Unmarshal(msg.Data, &parsed); err == nil {
// 			if len(parsed.Choices) > 0 {
// 				callback(parsed.Choices[0].Message.Content)
// 			}
// 		}
// 	})
// }
