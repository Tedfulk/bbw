package bitwarden

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

type Client struct {
	Session string
}

type Item struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
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
	cmd := exec.Command("bw", "login", email, password)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("login failed: %w", err)
	}

	// Extract session key using regex
	re := regexp.MustCompile(`export BW_SESSION="([^"]+)"`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) < 2 {
		return "", fmt.Errorf("could not find session key in output")
	}

	return matches[1], nil
}

// Search searches for items in the vault
func (c *Client) Search(query string) ([]Item, error) {
	if c.Session == "" {
		return nil, fmt.Errorf("no session provided")
	}

	cmd := exec.Command("bw", "list", "items", "--search", query, "--session", c.Session)
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

// GetPassword retrieves the password for a specific item
func (c *Client) GetPassword(id string) (string, error) {
	if c.Session == "" {
		return "", fmt.Errorf("no session provided")
	}

	cmd := exec.Command("bw", "get", "password", id, "--session", c.Session)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get password: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
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
	cmd := exec.Command("bw", "status")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("status check failed: %w", err)
	}

	var status struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(output, &status); err != nil {
		return "", fmt.Errorf("failed to parse status: %w", err)
	}

	return status.Status, nil
} 