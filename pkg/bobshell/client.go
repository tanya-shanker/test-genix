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
	fmt.Println("🔧 [Bob Shell] Building prompt...")

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
	fmt.Printf("🔧 [Bob Shell] Prompt length: %d characters\n", len(prompt))

	// Find bob CLI executable
	fmt.Println("🔧 [Bob Shell] Looking for Bob CLI executable...")
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
			return nil, fmt.Errorf("bob CLI not found in PATH or common locations. Please install Bob CLI")
		}
	}
	fmt.Printf("🔧 [Bob Shell] Found Bob CLI at: %s\n", bobPath)

	// Call Bob Shell CLI using non-interactive mode with -p flag
	// Usage: bob -p "prompt"
	fmt.Println("🔧 [Bob Shell] Executing Bob CLI command...")
	fmt.Println("🔧 [Bob Shell] Command: bob -p \"<prompt>\"")
	cmd := exec.Command(bobPath, "-p", prompt)

	// Set API key environment variable
	apiKeyLen := len(c.APIKey)
	if apiKeyLen > 0 {
		fmt.Printf("🔧 [Bob Shell] API key configured (length: %d chars)\n", apiKeyLen)
	} else {
		fmt.Println("⚠️  [Bob Shell] WARNING: API key is empty!")
	}
	cmd.Env = append(os.Environ(), fmt.Sprintf("BOBSHELL_API_KEY=%s", c.APIKey))

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	fmt.Println("🔧 [Bob Shell] Waiting for Bob CLI response...")
	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ [Bob Shell] Bob CLI failed: %v\n", err)
		fmt.Printf("❌ [Bob Shell] Stderr: %s\n", stderr.String())
		return nil, fmt.Errorf("bob CLI error: %w, stderr: %s", err, stderr.String())
	}
	fmt.Println("✅ [Bob Shell] Bob CLI completed successfully")

	// Parse response
	responseText := strings.TrimSpace(stdout.String())
	fmt.Printf("🔧 [Bob Shell] Response length: %d characters\n", len(responseText))

	response := &MessageResponse{
		Content: []ContentBlock{
			{
				Type: "text",
				Text: responseText,
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
