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
	grayText := pterm.NewStyle(pterm.FgGray).Sprint("g - generate | h - help | q - quit")
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
			pterm.Info.Println("Commands:")
			pterm.Info.Println("g - Generate password/passphrase")
			pterm.Info.Println("h - Show this help")
			pterm.Info.Println("q - Quit")
			continue
		case "g":
			if err := s.showPasswordGenerator(); err != nil {
				return err
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
			// Base display is just the name
			displayStr := item.Name

			// If it's a note (has notes but no login credentials), show (Note)
			if item.Notes != "" && item.Login.Username == "" && item.Login.Password == "" {
				displayStr = fmt.Sprintf("%s %s", displayStr, pterm.FgGray.Sprint("(Note)"))
			} else if item.Login.Username != "" {
				// If it has a username, show it
				displayStr = fmt.Sprintf("%s %s", displayStr, pterm.FgGray.Sprint(fmt.Sprintf("(%s)", item.Login.Username)))
			}
			
			options[i] = displayStr
		}
		options[len(items)] = "Cancel"

		// Show selection prompt
		selectedStr, err := pterm.DefaultInteractiveSelect.
				WithOptions(options).
				WithMaxHeight(10).
				WithDefaultText("Select item (↑/↓ arrows to move, enter to select)").
				Show()
		if err != nil {
			return fmt.Errorf("selection failed: %w", err)
		}

		if selectedStr == "Cancel" {
			continue
		}

		// Find the selected index
		var selectedIndex int
		for i, opt := range options {
			if i < len(items) && opt == selectedStr {
				selectedIndex = i
				break
			}
		}

		// Get selected item
		selectedItem := items[selectedIndex]

		// Create options slice only with non-empty fields
		var copyOptions []string
		if selectedItem.Login.Username != "" {
			copyOptions = append(copyOptions, fmt.Sprintf("Username: %s", selectedItem.Login.Username))
		}
		if selectedItem.Login.Password != "" {
			copyOptions = append(copyOptions, fmt.Sprintf("Password: %s", selectedItem.Login.Password))
		}
		if selectedItem.Notes != "" {
			copyOptions = append(copyOptions, fmt.Sprintf("Notes: %s", selectedItem.Notes))
		}
		if len(selectedItem.Login.Uris) > 0 && selectedItem.Login.Uris[0].Uri != "" {
			copyOptions = append(copyOptions, fmt.Sprintf("URL: %s", selectedItem.Login.Uris[0].Uri))
		}

		// Add metadata option
		copyOptions = append(copyOptions, "Show Metadata")
		copyOptions = append(copyOptions, "Cancel")

		// Show copy options
		action, err := pterm.DefaultInteractiveSelect.
			WithOptions(copyOptions).
				WithDefaultText("Choose action (↑/↓ arrows to move, enter to select)").
				Show()
		if err != nil {
			return fmt.Errorf("action selection failed: %w", err)
		}

		switch action {
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
		case fmt.Sprintf("URL: %s", selectedItem.Login.Uris[0].Uri):
			if err := clipboard.WriteAll(selectedItem.Login.Uris[0].Uri); err != nil {
				return fmt.Errorf("failed to copy URL: %w", err)
			}
			pterm.Success.Printf("URL for %s copied to clipboard!\n", selectedItem.Name)
		case "Show Metadata":
			// Format metadata
			var metadata strings.Builder
			metadata.WriteString(fmt.Sprintf("Created: %s\n", selectedItem.CreationDate))
			metadata.WriteString(fmt.Sprintf("Last Modified: %s\n", selectedItem.RevisionDate))
			if selectedItem.Login.PasswordRevisionDate != "" {
				metadata.WriteString(fmt.Sprintf("Password Last Modified: %s\n", selectedItem.Login.PasswordRevisionDate))
			}
			if len(selectedItem.PasswordHistory) > 0 {
				metadata.WriteString(fmt.Sprintf("Password Last Used: %s\n", selectedItem.PasswordHistory[0].LastUsedDate))
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