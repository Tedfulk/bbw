package tui

import (
	"fmt"

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
		options[i] = fmt.Sprintf("%s %s", item.Name, pterm.FgGray.Sprint(fmt.Sprintf("(%s)", item.Login.Username)))
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

	// Get password for selected item
	selectedItem := items[selectedIndex]

	// Fetch password ahead of time
	password, err := s.client.GetPassword(selectedItem.ID)
	if err != nil {
		return fmt.Errorf("failed to get password: %w", err)
	}

	// Show copy options
	action, err := pterm.DefaultInteractiveSelect.
		WithOptions([]string{
			fmt.Sprintf("Username: %s", selectedItem.Login.Username),
			fmt.Sprintf("Password: %s", password),
			"Cancel",
		}).
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
	case fmt.Sprintf("Password: %s", password):
		if err := clipboard.WriteAll(password); err != nil {
			return fmt.Errorf("failed to copy password: %w", err)
		}
		pterm.Success.Printf("Password for %s copied to clipboard!\n", selectedItem.Name)
	}

	return nil
} 