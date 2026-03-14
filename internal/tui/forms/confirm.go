package forms

import (
	"github.com/charmbracelet/huh"

	"github.com/botre/tickli/internal/tui/theme"
)

// RunConfirm displays a styled confirmation prompt.
func RunConfirm(title, description string) (bool, error) {
	var confirmed bool

	t := theme.Default()
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(title).
				Description(description).
				Affirmative("Yes").
				Negative("No").
				Value(&confirmed),
		),
	).WithTheme(huhTheme(t))

	err := form.Run()
	if err != nil {
		return false, err
	}

	return confirmed, nil
}
