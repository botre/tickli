package forms

import (
	"github.com/charmbracelet/huh"
)

// RunConfirm displays a styled confirmation prompt.
func RunConfirm(title, description string) (bool, error) {
	var confirmed bool

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(title).
				Description(description).
				Affirmative("Yes").
				Negative("No").
				Value(&confirmed),
		),
	).WithTheme(huh.ThemeDracula())

	err := form.Run()
	if err != nil {
		return false, err
	}

	return confirmed, nil
}
