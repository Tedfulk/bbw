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
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Notes          string    `json:"notes"`
	CreationDate   string    `json:"creationDate"`
	RevisionDate   string    `json:"revisionDate"`
	PasswordHistory []struct {
		LastUsedDate string `json:"lastUsedDate"`
		Password     string `json:"password"`
	} `json:"passwordHistory"`
	Login struct {
		Username            string `json:"username"`
		Password            string `json:"password"`
		PasswordRevisionDate string `json:"passwordRevisionDate"`
		Uris               []struct {
			Uri string `json:"uri"`
		} `json:"uris"`
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

// GeneratePassword generates a random password
func (c *Client) GeneratePassword(length int, includeSpecial bool) (string, error) {
	args := []string{"generate", "--length", fmt.Sprintf("%d", length)}
	
	if includeSpecial {
		// -lusn means lowercase, uppercase, special, numbers
		args = append(args, "-lusn", "--minSpecial", "2", "--minNumber", "2")
	} else {
		// -lun means lowercase, uppercase, numbers (no special)
		args = append(args, "-lun", "--minNumber", "2")
	}

	cmd := exec.Command("bw", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("password generation failed: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GeneratePassphrase generates a random passphrase with specified options
func (c *Client) GeneratePassphrase(words int, includeNumber bool) (string, error) {
	// Always include number by default
	args := []string{"generate", "--passphrase", "--words", fmt.Sprintf("%d", words), "--separator", "empty", "--includeNumber"}
	
	cmd := exec.Command("bw", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("passphrase generation failed: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
} 