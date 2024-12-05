package bitwarden

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Client struct {
	Session string
}

type Item struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Notes string `json:"notes"`
	Login struct {
			Username string `json:"username"`
			Password string `json:"password"`
	} `json:"login"`
}

// NewClient creates a new Bitwarden client with the given session
func NewClient(session string) *Client {
	return &Client{
		Session: session,
	}
}

// Login performs the login operation and returns the session key
func (c *Client) Login(email, password string) (string, error) {
	cmd := exec.Command("bw", "login", email, password, "--raw")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("login failed: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// Search searches for items in the vault
func (c *Client) Search(query string) ([]Item, error) {
	if c.Session == "" {
		return nil, fmt.Errorf("no session provided")
	}

	cmd := exec.Command("bw", "list", "items", "--search", query, "--session", c.Session, "--raw")
	
	// Set environment variables for faster execution
	env := os.Environ()
	env = append(env, fmt.Sprintf("BW_SESSION=%s", c.Session))
	cmd.Env = env
	
	// Ensure we're not using a shell
	cmd.Stderr = nil
	
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	var items []Item
	if err := json.Unmarshal(output, &items); err != nil {
		return nil, fmt.Errorf("failed to parse items: %w", err)
	}

	return items, nil
}

// Unlock unlocks the vault using the master password
func (c *Client) Unlock(password string) (string, error) {
	cmd := exec.Command("bw", "unlock", password, "--raw")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("unlock failed: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// Status checks if the vault is locked/unlocked
func (c *Client) Status() (string, error) {
	// If we have a valid session, we're unlocked
	if c.ValidateSession() {
		return "unlocked", nil
	}
	return "locked", nil
}

// Add a new method to validate session
func (c *Client) ValidateSession() bool {
	if c.Session == "" {
		return false
	}

	cmd := exec.Command("bw", "list", "items", "--search", "", "--session", c.Session, "--raw")
	return cmd.Run() == nil
} 