package tui

import (
	"fmt"
	"strings"

	"github.com/Tedfulk/bbw/internal/bitwarden"
	"github.com/atotto/clipboard"
	"github.com/pterm/pterm"
)

type SearchUI struct {
	client *bitwarden.Client
}

func NewSearchUI(client *bitwarden.Client) *SearchUI {
	return &SearchUI{
		client: client,
	}
}

// Create helper function for display string
func createDisplayString(item bitwarden.Item) string {
	displayStr := item.Name
	if item.Notes != "" && item.Login.Username == "" && item.Login.Password == "" {
		displayStr = fmt.Sprintf("%s %s", displayStr, pterm.FgGray.Sprint("(Note)"))
	} else if item.Login.Username != "" {
		displayStr = fmt.Sprintf("%s %s", displayStr, pterm.FgGray.Sprint(fmt.Sprintf("(%s)", item.Login.Username)))
	}
	return displayStr
}

func (s *SearchUI) Show() error {
	// Show centered title box
	boxContent := pterm.DefaultBox.
		WithBoxStyle(pterm.NewStyle(pterm.FgCyan)).
		WithTopPadding(1).
		WithBottomPadding(1).
		WithLeftPadding(10).
		WithRightPadding(10).
		Sprint("Better Bitwarden")
	
	pterm.DefaultCenter.Println(boxContent)
	
	// Show centered commands
	grayText := pterm.NewStyle(pterm.FgGray).Sprint("g - generate | h - help | s - status | u - update | q - quit")
	pterm.DefaultCenter.Println(grayText)
	pterm.Println()

	for {
		// Show search prompt
		
		query, err := pterm.DefaultInteractiveTextInput.
			WithMultiLine(false).
			WithDefaultText(pterm.FgMagenta.Sprint("Search")).
			WithTextStyle(pterm.NewStyle(pterm.FgMagenta)).
			Show()
		if err != nil {
			return fmt.Errorf("failed to get search input: %w", err)
		}

		switch query {
		case "q":
			return nil
		case "h":
			helpText := "Commands:\n\n" +
				"g - Generate password/passphrase\n" +
				"h - Show this help\n" +
				"s - Show vault status\n" +
				"u - Update CLI and sync vault\n" +
				"q - Quit"

			boxContent := pterm.DefaultBox.
				WithBoxStyle(pterm.NewStyle(pterm.FgCyan)).
				WithTitle("Help").
				WithTitleTopCenter().
				WithTopPadding(1).
				WithBottomPadding(1).
				WithLeftPadding(3).
				WithRightPadding(3).
				Sprint(helpText)

			pterm.DefaultCenter.Println(boxContent)
			continue
		case "g":
			if err := s.showPasswordGenerator(); err != nil {
				return err
			}
			continue
		case "u":
			pterm.Info.Println("Checking for updates and syncing vault...")
			if err := s.client.CheckUpdates(); err != nil {
				pterm.Warning.Printf("Update check failed: %v\n", err)
			}
			if err := s.client.SyncVault(); err != nil {
				pterm.Warning.Printf("Sync failed: %v\n", err)
			} else {
				pterm.Success.Println("Vault synced successfully!")
			}
			continue
		case "s":
			status, err := s.client.GetStatus()
			if err != nil {
				pterm.Warning.Printf("Failed to get status: %v\n", err)
			} else {
				pterm.DefaultBox.
					WithBoxStyle(pterm.NewStyle(pterm.FgCyan)).
					WithTitle("Vault Status").
					WithTitleTopCenter().
					Println(
						pterm.FgWhite.Sprintf("Server URL: %v\n", status["serverUrl"]) +
						pterm.FgWhite.Sprintf("Last Sync: %v\n", status["lastSync"]) +
						pterm.FgWhite.Sprintf("Email: %v\n", status["userEmail"]) +
						pterm.FgWhite.Sprintf("User ID: %v\n", status["userId"]) +
						pterm.FgGreen.Sprintf("Status: %v", status["status"]),
					)
			}
			continue
		case "":
			continue
		}

		// Search for items
		items, err := s.client.Search(query)
		if err != nil {
			return fmt.Errorf("search failed: %w", err)
		}

		if len(items) == 0 {
			pterm.Warning.Println("No items found")
			continue
		}

		// Create options for selection
		options := make([]string, len(items)+1)
		for i, item := range items {
			options[i] = createDisplayString(item)
		}
		options[len(items)] = "Cancel"

		// Show selection prompt
		selectedStr, _ := pterm.DefaultInteractiveSelect.
				WithOptions(options).
				WithMaxHeight(15).
				WithDefaultText("Select item (↑/↓ arrows to move, enter to select)").
				Show()

		if selectedStr == "Cancel" {
			continue
		}

		// Find the matching item
		var selectedItem bitwarden.Item
		for _, item := range items {
			if createDisplayString(item) == selectedStr {
				selectedItem = item
				break
			}
		}

		if err := s.handleItemSelection(selectedItem); err != nil {
			return fmt.Errorf("error handling item selection: %w", err)
		}
	}
}

func (s *SearchUI) handleItemSelection(selectedItem bitwarden.Item) error {
	// Create list of actions
	actions := []string{
		fmt.Sprintf("Username: %s", selectedItem.Login.Username),
		fmt.Sprintf("Password: %s", selectedItem.Login.Password),
	}
	if selectedItem.Notes != "" {
		actions = append(actions, fmt.Sprintf("Notes: %s", selectedItem.Notes))
	}
	
	// Handle URI display
	urlAction := "URL: No URI available"
	if len(selectedItem.Login.Uris) > 0 && selectedItem.Login.Uris[0].Uri != "" {
		urlAction = fmt.Sprintf("URL: %s", selectedItem.Login.Uris[0].Uri)
	}

	actions = append(actions, "Show Metadata", "Cancel")

	selectedAction, err := pterm.DefaultInteractiveSelect.
			WithOptions(actions).
			WithMaxHeight(25).
			WithDefaultText("Select action to copy to clipboard").
			Show()
	if err != nil {
		return fmt.Errorf("failed to get action: %w", err)
	}

	if selectedAction == "Cancel" {
		return nil
	}

	// Handle the selected action
	switch selectedAction {
	case fmt.Sprintf("Username: %s", selectedItem.Login.Username):
		if err := clipboard.WriteAll(selectedItem.Login.Username); err != nil {
			return fmt.Errorf("failed to copy username: %w", err)
		}
		pterm.Success.Printf("Username for %s copied to clipboard!\n", selectedItem.Name)
	case fmt.Sprintf("Password: %s", selectedItem.Login.Password):
		if err := clipboard.WriteAll(selectedItem.Login.Password); err != nil {
			return fmt.Errorf("failed to copy password: %w", err)
		}
		pterm.Success.Printf("Password for %s copied to clipboard!\n", selectedItem.Name)
	case fmt.Sprintf("Notes: %s", selectedItem.Notes):
		if err := clipboard.WriteAll(selectedItem.Notes); err != nil {
			return fmt.Errorf("failed to copy notes: %w", err)
		}
		pterm.Success.Printf("Notes for %s copied to clipboard!\n", selectedItem.Name)
	case urlAction:
		if len(selectedItem.Login.Uris) > 0 && selectedItem.Login.Uris[0].Uri != "" {
			if err := clipboard.WriteAll(selectedItem.Login.Uris[0].Uri); err != nil {
				return fmt.Errorf("failed to copy URL: %w", err)
			}
			pterm.Success.Printf("URL for %s copied to clipboard!\n", selectedItem.Name)
		} else {
			pterm.Info.Println("No URI available to copy")
		}
	case "Show Metadata":
		// Format metadata
		var metadata strings.Builder
		metadata.WriteString(fmt.Sprintf("Created: %s\n", selectedItem.CreationDate))
		metadata.WriteString(fmt.Sprintf("Last Modified: %s\n", selectedItem.RevisionDate))
		
		// Handle password revision date
		if selectedItem.Login.PasswordRevisionDate != "" {
			metadata.WriteString(fmt.Sprintf("Password Last Modified: %s\n", selectedItem.Login.PasswordRevisionDate))
		}
		
		// Handle password history
		if len(selectedItem.PasswordHistory) > 0 && selectedItem.PasswordHistory[0].LastUsedDate != "" {
			metadata.WriteString(fmt.Sprintf("Password Last Used: %s\n", selectedItem.PasswordHistory[0].LastUsedDate))
		}
		
		// Handle URIs
		if len(selectedItem.Login.Uris) > 0 && selectedItem.Login.Uris[0].Uri != "" {
			metadata.WriteString(fmt.Sprintf("URI: %s\n", selectedItem.Login.Uris[0].Uri))
		} else {
			metadata.WriteString("URI: No URI available\n")
		}

		// Print metadata
		pterm.Info.Println("Metadata for", selectedItem.Name)
		fmt.Println(metadata.String())

		// Ask if user wants to copy metadata
		copyMetadata, _ := pterm.DefaultInteractiveConfirm.
			WithDefaultText("Copy metadata to clipboard?").
			Show()
			
		if copyMetadata {
			if err := clipboard.WriteAll(metadata.String()); err != nil {
				return fmt.Errorf("failed to copy metadata: %w", err)
			}
			pterm.Success.Printf("Metadata for %s copied to clipboard!\n", selectedItem.Name)
		}
	}

	return nil
}

func (s *SearchUI) showPasswordGenerator() error {
	// Show generator options
	genType, err := pterm.DefaultInteractiveSelect.
		WithOptions([]string{
			"Password with Special (-lusn --length 18 --minSpecial 2 --minNumber 2)",
			"Password without Special (-lun --length 18 --minNumber 2)",
			"Passphrase (5 words 1 number)",
			"Cancel",
		}).
		WithDefaultText("Select generator type").
		Show()
	if err != nil {
		return err
	}

	switch genType {
	case "Password with Special (-lusn --length 18 --minSpecial 2 --minNumber 2)":
		password, err := s.client.GeneratePassword(18, true)
		if err != nil {
			return err
		}
		if err := clipboard.WriteAll(password); err != nil {
			return fmt.Errorf("failed to copy password: %w", err)
		}
		pterm.Success.Printf("Generated password: %s\n", password)
		pterm.Success.Println("Password copied to clipboard!")

	case "Password without Special (-lun --length 18 --minNumber 2)":
		password, err := s.client.GeneratePassword(18, false)
		if err != nil {
			return err
		}
		if err := clipboard.WriteAll(password); err != nil {
			return fmt.Errorf("failed to copy password: %w", err)
		}
		pterm.Success.Printf("Generated password: %s\n", password)
		pterm.Success.Println("Password copied to clipboard!")

	case "Passphrase (5 words 1 number)":
		passphrase, err := s.client.GeneratePassphrase(5, true)
		if err != nil {
			return err
		}
		if err := clipboard.WriteAll(passphrase); err != nil {
			return fmt.Errorf("failed to copy passphrase: %w", err)
		}
		pterm.Success.Printf("Generated passphrase: %s\n", passphrase)
		pterm.Success.Println("Passphrase copied to clipboard!")
	}
	return nil
} 