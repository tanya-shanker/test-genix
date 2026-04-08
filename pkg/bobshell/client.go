package bobshell

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Client represents a Bob Shell CLI client
type Client struct {
	APIKey string
}

// NewClient creates a new Bob Shell CLI client
func NewClient(apiKey string) *Client {
	return &Client{
		APIKey: apiKey,
	}
}

// Message represents a message in the conversation
type Message struct {
	Role    string
	Content string
}

// MessageRequest represents a request to Bob Shell
type MessageRequest struct {
	Model       string
	MaxTokens   int
	Messages    []Message
	System      string
	Temperature float64
}

// MessageResponse represents Bob Shell's response
type MessageResponse struct {
	Content []ContentBlock
	Model   string
	Role    string
}

// ContentBlock represents a content block in the response
type ContentBlock struct {
	Type string
	Text string
}

// CreateMessage sends a message to Bob Shell CLI and returns the response
func (c *Client) CreateMessage(req MessageRequest) (*MessageResponse, error) {
	// Build the prompt from messages
	var promptBuilder strings.Builder

	// Add system message if present
	if req.System != "" {
		promptBuilder.WriteString(req.System)
		promptBuilder.WriteString("\n\n")
	}

	// Add user messages
	for _, msg := range req.Messages {
		if msg.Role == "user" {
			promptBuilder.WriteString(msg.Content)
		}
	}

	prompt := promptBuilder.String()

	// Create temporary file for prompt
	tmpFile, err := os.CreateTemp("", "bob-prompt-*.txt")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(prompt); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("failed to write prompt: %w", err)
	}
	tmpFile.Close()

	// Find bob CLI executable
	bobPath, err := exec.LookPath("bob")
	if err != nil {
		// Try common installation paths
		possiblePaths := []string{
			"/usr/local/bin/bob",
			"/usr/bin/bob",
			os.Getenv("HOME") + "/.bob/bin/bob",
		}
		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				bobPath = path
				break
			}
		}
		if bobPath == "" {
			return nil, fmt.Errorf("bob CLI not found in PATH or common locations. Please install Bob CLI: https://github.com/IBM/bob-cli")
		}
	}

	// Call Bob Shell CLI
	cmd := exec.Command(bobPath, "ask", "--file", tmpFile.Name())

	// Set API key environment variable
	cmd.Env = append(os.Environ(), fmt.Sprintf("BOBSHELL_API_KEY=%s", c.APIKey))

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("bob CLI error: %w, stderr: %s", err, stderr.String())
	}

	// Parse response
	response := &MessageResponse{
		Content: []ContentBlock{
			{
				Type: "text",
				Text: strings.TrimSpace(stdout.String()),
			},
		},
		Model: req.Model,
		Role:  "assistant",
	}

	return response, nil
}

// ExtractText extracts text from the response
func (r *MessageResponse) ExtractText() string {
	if len(r.Content) == 0 {
		return ""
	}
	return r.Content[0].Text
}

// Made with Bob
