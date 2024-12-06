package main

import (
	"os"

	"github.com/Tedfulk/bbw/internal/bitwarden"
	"github.com/Tedfulk/bbw/internal/config"
	"github.com/Tedfulk/bbw/internal/tui"
	"github.com/pterm/pterm"
)

func main() {
	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		pterm.Error.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Create Bitwarden client
	client := bitwarden.NewClient(cfg.Session)

	// Only check status and unlock if we don't have a valid session
	if !client.ValidateSession() {
		if cfg.Password == "" {
			if err := firstTimeSetup(cfg); err != nil {
				pterm.Error.Printf("Setup failed: %v\n", err)
				os.Exit(1)
			}
			client = bitwarden.NewClient(cfg.Session)
		} else {
			session, err := client.Unlock(cfg.Password)
			if err != nil {
				pterm.Error.Printf("Failed to unlock vault: %v\n", err)
				os.Exit(1)
			}
			cfg.Session = session
			if err := config.SaveConfig(cfg); err != nil {
				pterm.Error.Printf("Failed to save config: %v\n", err)
				os.Exit(1)
			}
			client = bitwarden.NewClient(session)
		}
	}

	// Create and show search UI
	searchUI := tui.NewSearchUI(client)
	if err := searchUI.Show(); err != nil {
		pterm.Error.Printf("Search failed: %v\n", err)
		os.Exit(1)
	}
}

func firstTimeSetup(cfg *config.Config) error {
	pterm.Info.Println("First time setup")

	email, err := pterm.DefaultInteractiveTextInput.
		WithDefaultText("Enter your Bitwarden email").
		Show()
	if err != nil {
		return err
	}

	password, err := pterm.DefaultInteractiveTextInput.
		WithDefaultText("Enter your master password").
		WithMask("*").
		Show()
	if err != nil {
		return err
	}

	client := bitwarden.NewClient("")
	
	// Check if already logged in
	status, err := client.GetStatus()
	if err != nil {
		return err
	}
	
	var session string
	if status["userEmail"] == email {
		// Already logged in, just unlock
		session, err = client.Unlock(password)
	} else {
		// New login needed
		session, err = client.Login(email, password)
	}
	if err != nil {
		return err
	}

	cfg.Email = email
	cfg.Password = password
	cfg.Session = session

	return config.SaveConfig(cfg)
} 