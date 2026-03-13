package forms

import (
	"github.com/charmbracelet/huh"
)

// RunSelect displays a styled selection prompt.
func RunSelect(title string, options []string) (int, error) {
	opts := make([]huh.Option[int], len(options))
	for i, o := range options {
		opts[i] = huh.NewOption(o, i)
	}

	var selected int
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title(title).
				Options(opts...).
				Value(&selected),
		),
	).WithTheme(huh.ThemeDracula())

	err := form.Run()
	if err != nil {
		return -1, err
	}

	return selected, nil
}
