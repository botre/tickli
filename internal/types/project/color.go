package project

import (
	"encoding/json"
	"fmt"
	"github.com/gookit/color"
	"github.com/spf13/cobra"
	"regexp"
	"strings"
)

var DefaultColor = Color(color.HEX("#000000"))

type Color color.RGBColor

var ColorCompletion = []cobra.Completion{
	cobra.CompletionWithDesc(DefaultColor.String(), "🎨Default color"),
	cobra.CompletionWithDesc("#EC6665", "❤️Red"),
	cobra.CompletionWithDesc("#F2B04A", "🧡Orange"),
	cobra.CompletionWithDesc("#FFD866", "💛Yellow"),
	cobra.CompletionWithDesc("#5CD0A7", "💚Green"),
	cobra.CompletionWithDesc("#9BECEC", "🩵Cyan"),
	cobra.CompletionWithDesc("#4AA6EF", "💙Blue"),
	cobra.CompletionWithDesc("#CF66F6", "💜Purple"),
	cobra.CompletionWithDesc("#EC70A5", "💖Pink"),
	cobra.CompletionWithDesc("#FDF8DC", "🤍White"),
}

var ColorCompletionFunc = cobra.FixedCompletions(ColorCompletion, cobra.ShellCompDirectiveNoFileComp)

func (c *Color) UnmarshalJSON(data []byte) error {

	var colorStr string
	if err := json.Unmarshal(data, &colorStr); err != nil {
		return err
	}

	if colorStr == "" {
		*c = DefaultColor
	} else {
		*c = Color(color.HEX(colorStr))
	}
	return nil
}

func (c Color) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

func (c Color) Sprint(a ...any) string {
	return color.RGBColor(c).Sprint(a...)
}

func (c Color) String() string {
	return "#" + strings.ToUpper(color.RGBColor(c).Hex())
}

func (c *Color) Set(s string) error {
	// Validate the hex color format
	s = strings.TrimSpace(s)

	// This pattern matches both 3-digit and 6-digit hex colors, with optional "#" prefix
	hexPattern := regexp.MustCompile(`^#?([0-9A-Fa-f]{3}|[0-9A-Fa-f]{6})$`)
	if !hexPattern.MatchString(s) {
		return fmt.Errorf("invalid hex color format: must be a 3 or 6-digit hex color code (e.g., '#F18' or '#F18181')")
	}

	// Add the "#" back if it was missing
	if !strings.HasPrefix(s, "#") {
		s = "#" + s
	}

	*c = Color(color.HEX(s))
	return nil
}

func (c *Color) Type() string {
	return "ProjectColor"
}
