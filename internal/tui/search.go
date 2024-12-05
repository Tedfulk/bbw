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
	
	// Show search prompt
	query, err := pterm.DefaultInteractiveTextInput.
		WithMultiLine(false).
		WithDefaultText(pterm.FgMagenta.Sprint("Search")).
		WithTextStyle(pterm.NewStyle(pterm.FgMagenta)).
		Show()
	if err != nil {
		return fmt.Errorf("failed to get search input: %w", err)
	}

	// Search for items
	items, err := s.client.Search(query)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if len(items) == 0 {
		pterm.Warning.Println("No items found")
		return nil
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
		return nil
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

	return nil
} 