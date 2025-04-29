package listtools

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

type MCPMessage string

type VerboseKey struct{}

const (
	Handshake1 MCPMessage = `{"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{"tools":true,"prompts":false,"resources":true,"logging":false,"roots":{"listChanged":false}},"clientInfo":{"name":"cursor-vscode","version":"1.0.0"}},"jsonrpc":"2.0","id":0}`
	Handshake2 MCPMessage = `{"method":"notifications/initialized","jsonrpc":"2.0"}`
	ListTools  MCPMessage = `{"method":"tools/list","jsonrpc":"2.0","id":1}`
)

type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Client struct {
	baseURL          string
	messagesEndpoint string
	messages         chan string
	httpClient       *http.Client
	toolsMessage     chan string
	closed           bool
	closeMux         sync.Mutex
}

// NewClient creates a new MCP client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:      baseURL,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
		messages:     make(chan string),
		toolsMessage: make(chan string),
	}
}

// InitiateConnection starts an SSE connection and returns a channel for messages
func (c *Client) InitiateConnection(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/sse", c.baseURL), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/event-stream") {
		return fmt.Errorf("unexpected content type: %s, expected text/event-stream", contentType)
	}

	log.Println("Processing SSE stream")
	go c.processSSEStream(ctx, resp)
	return nil
}

// processSSEStream processes the SSE stream and sends messages to the channel
func (c *Client) processSSEStream(ctx context.Context, resp *http.Response) {
	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)
	var dataBuffer strings.Builder
	inMessage := false

	for {
		select {
		case <-ctx.Done():
			return
		default:
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}

			line = strings.TrimSpace(line)

			// Empty line marks the end of a message
			if line == "" {
				if inMessage && dataBuffer.Len() > 0 {
					log.Println("Sending message: " + dataBuffer.String())
					c.messages <- dataBuffer.String()
					log.Println("Sent message: " + dataBuffer.String())
					dataBuffer.Reset()
					inMessage = false
				}
				continue
			}

			// Skip comments
			if strings.HasPrefix(line, ":") {
				continue
			}

			// Handle data lines
			if strings.HasPrefix(line, "data:") {
				data := strings.TrimPrefix(line, "data:")
				data = strings.TrimSpace(data)

				// If this isn't the first data line, add a newline
				if inMessage && dataBuffer.Len() > 0 {
					dataBuffer.WriteString("\n")
				}

				dataBuffer.WriteString(data)
				inMessage = true
			}
		}
	}
}

// SendMessage sends a message to the server
func (c *Client) SendMessage(ctx context.Context, message MCPMessage) error {
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/%s", c.baseURL, c.messagesEndpoint), strings.NewReader(string(message)))
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}

	defer resp.Body.Close()
	return nil
}

func (c *Client) Close() {
	c.closeMux.Lock()
	defer c.closeMux.Unlock()

	if !c.closed {
		close(c.messages)
		close(c.toolsMessage)
		c.closed = true
	}
}

// ParseToolsResponse parses the tools response from the server
func ParseToolsResponse(toolsResponse string) ([]Tool, error) {
	type ToolResponse struct {
		Name        string          `json:"name"`
		Description string          `json:"description,omitempty"`
		InputSchema json.RawMessage `json:"inputSchema"`
	}

	type Response struct {
		Result struct {
			Tools []ToolResponse `json:"tools"`
		} `json:"result"`
	}

	var response Response
	err := json.Unmarshal([]byte(toolsResponse), &response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tools response: %v", err)
	}

	tools := make([]Tool, 0, len(response.Result.Tools))
	for _, t := range response.Result.Tools {
		if t.Description == "" {
			t.Description = "No description provided"
		}
		tools = append(tools, Tool{
			Name:        t.Name,
			Description: t.Description,
		})
	}

	return tools, nil
}

// parseMessagesEndpoint parses the messages endpoint from the message
func parseMessagesEndpoint(message string) string {
	re := regexp.MustCompile(`\/(messages[^\s]+)`)
	matches := re.FindStringSubmatch(message)
	if len(matches) == 0 {
		return ""
	}

	log.Printf("Found messages endpoint: %s", matches[1])
	return matches[1]
}

// FetchToolsResponse fetches the tools response from the server
func FetchToolsResponse(ctx context.Context, client *Client) (string, error) {
	log.Println("Initiating connection")
	err := client.InitiateConnection(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to initiate connection: %v", err)
	}

	var error error

	go func() {
		log.Println("Waiting for messages")
		for {
			select {
			case message := <-client.messages:
				log.Printf("Received message: %s", message)
				if strings.Contains(message, "session") {
					client.messagesEndpoint = parseMessagesEndpoint(message)
				}

				if strings.Contains(message, `"result":{"tools":`) {
					client.toolsMessage <- message
				}
			case <-ctx.Done():
				error = fmt.Errorf("context deadline exceeded: %v", ctx.Err())
				return
			}
		}
	}()

	if error != nil {
		return "", error
	}

	for {
		if client.messagesEndpoint != "" {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	log.Printf("Successfully got the messagesEndpoint: %s", client.messagesEndpoint)

	if err := client.SendMessage(ctx, Handshake1); err != nil {
		return "", fmt.Errorf("failed to send handshake1: %v", err)
	}

	if err := client.SendMessage(ctx, Handshake2); err != nil {
		return "", fmt.Errorf("failed to send handshake2: %v", err)
	}

	if err := client.SendMessage(ctx, ListTools); err != nil {
		return "", fmt.Errorf("failed to send list_tools: %v", err)
	}

	for {
		select {
		case toolsResponse := <-client.toolsMessage:
			if toolsResponse == "" {
				log.Println("Empty tools response received")
				continue
			}

			return toolsResponse, nil
		case <-ctx.Done():
			log.Printf("Context deadline exceeded: %v", ctx.Err())
			return "", fmt.Errorf("context deadline exceeded: %v", ctx.Err())
		}
	}
}
