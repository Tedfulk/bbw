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
	env     []string
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
	// Set up environment variables once
	env := os.Environ()
	if session != "" {
		env = append(env, fmt.Sprintf("BW_SESSION=%s", session))
	}
	
	return &Client{
		Session: session,
		env:     env,
	}
}

// Login performs the login operation and returns the session key
func (c *Client) Login(email, password string) (string, error) {
	cmd := exec.Command("bw", "login", email, password, "--raw")
	cmd.Env = c.env
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("login failed: %w: %s", err, string(output))
	}

	newSession := strings.TrimSpace(string(output))
	c.Session = newSession
	c.env = append(os.Environ(), fmt.Sprintf("BW_SESSION=%s", newSession))

	return newSession, nil
}

// Search searches for items in the vault
func (c *Client) Search(query string) ([]Item, error) {
	if c.Session == "" {
		return nil, fmt.Errorf("no session provided")
	}

	cmd := exec.Command("bw", "list", "items", "--search", query, "--raw")
	cmd.Env = c.env
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
	cmd.Env = c.env
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("unlock failed: %w", err)
	}

	newSession := strings.TrimSpace(string(output))
	c.Session = newSession
	c.env = append(os.Environ(), fmt.Sprintf("BW_SESSION=%s", newSession))

	return newSession, nil
}

// Status checks if the vault is locked/unlocked
func (c *Client) Status() (string, error) {
	// If we have a valid session, we're unlocked
	if c.ValidateSession() {
		return "unlocked", nil
	}
	return "locked", nil
}

// ValidateSession checks if the session is valid using a lightweight status check
func (c *Client) ValidateSession() bool {
	if c.Session == "" {
		return false
	}

	cmd := exec.Command("bw", "status")
	cmd.Env = c.env
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	var status map[string]interface{}
	if err := json.Unmarshal(output, &status); err != nil {
		return false
	}

	if statusStr, ok := status["status"].(string); ok {
		return statusStr == "unlocked"
	}
	return false
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

// SyncVault syncs the vault with the server
func (c *Client) SyncVault() error {
	cmd := exec.Command("bw", "sync", "--session", c.Session)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}
	return nil
}

// CheckUpdates checks for Bitwarden CLI updates
func (c *Client) CheckUpdates() error {
	cmd := exec.Command("bw", "update")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("update check failed: %w", err)
	}
	return nil
}

// GetStatus returns detailed status information
func (c *Client) GetStatus() (map[string]interface{}, error) {
	cmd := exec.Command("bw", "status")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("status check failed: %w", err)
	}

	var status map[string]interface{}
	if err := json.Unmarshal(output, &status); err != nil {
		return nil, fmt.Errorf("failed to parse status: %w", err)
	}

	return status, nil
}

// helper function to run commands with the client's environment
func (c *Client) runCommand(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.Env = c.env
	return cmd
} 