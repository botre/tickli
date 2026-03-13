package theme

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette — a cohesive set of colors used throughout the TUI.
// Inspired by the existing TickTick priority colors with additions for UI chrome.
type Palette struct {
	// Brand
	Primary   lipgloss.Color
	Secondary lipgloss.Color
	Accent    lipgloss.Color

	// Priorities (matching TickTick's scheme)
	PriorityNone   lipgloss.Color
	PriorityLow    lipgloss.Color
	PriorityMedium lipgloss.Color
	PriorityHigh   lipgloss.Color

	// Status
	Success lipgloss.Color
	Warning lipgloss.Color
	Error   lipgloss.Color
	Info    lipgloss.Color

	// Surface
	Base    lipgloss.Color
	Surface lipgloss.Color
	Overlay lipgloss.Color
	Muted   lipgloss.Color
	Subtle  lipgloss.Color
	Text    lipgloss.Color
	SubText lipgloss.Color
}

// Theme holds all styled components for the TUI.
// Pass this through your model tree for consistent styling.
type Theme struct {
	Palette Palette

	// App chrome
	Logo       lipgloss.Style
	Header     lipgloss.Style
	StatusBar  lipgloss.Style
	HelpBar    lipgloss.Style
	Breadcrumb lipgloss.Style

	// Content
	Title       lipgloss.Style
	Subtitle    lipgloss.Style
	Body        lipgloss.Style
	Muted       lipgloss.Style
	Bold        lipgloss.Style
	Description lipgloss.Style

	// List
	ListItem         lipgloss.Style
	SelectedItem     lipgloss.Style
	ActiveItem       lipgloss.Style
	DimmedItem       lipgloss.Style
	FilterPrompt     lipgloss.Style
	FilterMatch      lipgloss.Style
	FilterMatchCount lipgloss.Style

	// Task specific
	TaskTitle      lipgloss.Style
	TaskContent    lipgloss.Style
	TaskDue        lipgloss.Style
	TaskDueOverdue lipgloss.Style
	TaskTag        lipgloss.Style
	TaskProject    lipgloss.Style

	// Priority styles
	PriorityNone   lipgloss.Style
	PriorityLow    lipgloss.Style
	PriorityMedium lipgloss.Style
	PriorityHigh   lipgloss.Style

	// Status indicators
	StatusComplete lipgloss.Style
	StatusPending  lipgloss.Style

	// Panels and cards
	Card       lipgloss.Style
	CardActive lipgloss.Style
	Panel      lipgloss.Style
	Divider    lipgloss.Style

	// Form elements
	FormLabel    lipgloss.Style
	FormField    lipgloss.Style
	FormButton   lipgloss.Style
	FormSelected lipgloss.Style

	// Messages
	SuccessMessage lipgloss.Style
	ErrorMessage   lipgloss.Style
	WarningMessage lipgloss.Style
	InfoMessage    lipgloss.Style

	// Spinner
	Spinner lipgloss.Style
}

// Default returns the default tickli theme.
func Default() Theme {
	hasDarkBG := lipgloss.HasDarkBackground()

	p := Palette{
		Primary:   lipgloss.Color("#7C3AED"),
		Secondary: lipgloss.Color("#6D28D9"),
		Accent:    lipgloss.Color("#A78BFA"),

		PriorityNone:   lipgloss.Color("#6B7280"),
		PriorityLow:    lipgloss.Color("#4772F9"),
		PriorityMedium: lipgloss.Color("#FAA80B"),
		PriorityHigh:   lipgloss.Color("#D52B24"),

		Success: lipgloss.Color("#10B981"),
		Warning: lipgloss.Color("#F59E0B"),
		Error:   lipgloss.Color("#EF4444"),
		Info:    lipgloss.Color("#3B82F6"),
	}

	if hasDarkBG {
		p.Base = lipgloss.Color("#0F0F14")
		p.Surface = lipgloss.Color("#1A1A24")
		p.Overlay = lipgloss.Color("#252532")
		p.Muted = lipgloss.Color("#6B7280")
		p.Subtle = lipgloss.Color("#374151")
		p.Text = lipgloss.Color("#F9FAFB")
		p.SubText = lipgloss.Color("#9CA3AF")
	} else {
		p.Base = lipgloss.Color("#FFFFFF")
		p.Surface = lipgloss.Color("#F9FAFB")
		p.Overlay = lipgloss.Color("#F3F4F6")
		p.Muted = lipgloss.Color("#9CA3AF")
		p.Subtle = lipgloss.Color("#D1D5DB")
		p.Text = lipgloss.Color("#111827")
		p.SubText = lipgloss.Color("#6B7280")
	}

	return newTheme(p)
}

func newTheme(p Palette) Theme {
	return Theme{
		Palette: p,

		// App chrome
		Logo: lipgloss.NewStyle().
			Foreground(p.Primary).
			Bold(true),

		Header: lipgloss.NewStyle().
			Foreground(p.Text).
			Background(p.Primary).
			Bold(true).
			Padding(0, 1),

		StatusBar: lipgloss.NewStyle().
			Foreground(p.SubText).
			Background(p.Surface).
			Padding(0, 1),

		HelpBar: lipgloss.NewStyle().
			Foreground(p.Muted),

		Breadcrumb: lipgloss.NewStyle().
			Foreground(p.SubText).
			Faint(true),

		// Content
		Title: lipgloss.NewStyle().
			Foreground(p.Text).
			Bold(true),

		Subtitle: lipgloss.NewStyle().
			Foreground(p.SubText),

		Body: lipgloss.NewStyle().
			Foreground(p.Text),

		Muted: lipgloss.NewStyle().
			Foreground(p.Muted),

		Bold: lipgloss.NewStyle().
			Foreground(p.Text).
			Bold(true),

		Description: lipgloss.NewStyle().
			Foreground(p.SubText).
			Italic(true),

		// List
		ListItem: lipgloss.NewStyle().
			Foreground(p.Text).
			Padding(0, 0, 0, 2),

		SelectedItem: lipgloss.NewStyle().
			Foreground(p.Primary).
			Bold(true).
			Padding(0, 0, 0, 1).
			Border(lipgloss.ThickBorder(), false, false, false, true).
			BorderForeground(p.Primary),

		ActiveItem: lipgloss.NewStyle().
			Foreground(p.Accent),

		DimmedItem: lipgloss.NewStyle().
			Foreground(p.Muted),

		FilterPrompt: lipgloss.NewStyle().
			Foreground(p.Primary).
			Bold(true),

		FilterMatch: lipgloss.NewStyle().
			Foreground(p.Accent).
			Underline(true),

		FilterMatchCount: lipgloss.NewStyle().
			Foreground(p.Muted),

		// Task specific
		TaskTitle: lipgloss.NewStyle().
			Foreground(p.Text).
			Bold(true),

		TaskContent: lipgloss.NewStyle().
			Foreground(p.SubText),

		TaskDue: lipgloss.NewStyle().
			Foreground(p.Info),

		TaskDueOverdue: lipgloss.NewStyle().
			Foreground(p.Error).
			Bold(true),

		TaskTag: lipgloss.NewStyle().
			Foreground(p.Accent).
			Background(p.Overlay).
			Padding(0, 1),

		TaskProject: lipgloss.NewStyle().
			Foreground(p.Secondary).
			Bold(true),

		// Priorities
		PriorityNone: lipgloss.NewStyle().
			Foreground(p.PriorityNone),

		PriorityLow: lipgloss.NewStyle().
			Foreground(p.PriorityLow),

		PriorityMedium: lipgloss.NewStyle().
			Foreground(p.PriorityMedium),

		PriorityHigh: lipgloss.NewStyle().
			Foreground(p.PriorityHigh),

		// Status
		StatusComplete: lipgloss.NewStyle().
			Foreground(p.Success),

		StatusPending: lipgloss.NewStyle().
			Foreground(p.SubText),

		// Panels
		Card: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(p.Subtle).
			Padding(1, 2),

		CardActive: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(p.Primary).
			Padding(1, 2),

		Panel: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(p.Subtle).
			Padding(0, 1),

		Divider: lipgloss.NewStyle().
			Foreground(p.Subtle),

		// Forms
		FormLabel: lipgloss.NewStyle().
			Foreground(p.Text).
			Bold(true),

		FormField: lipgloss.NewStyle().
			Foreground(p.Text).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(p.Subtle).
			Padding(0, 1),

		FormButton: lipgloss.NewStyle().
			Foreground(p.Base).
			Background(p.Primary).
			Padding(0, 2).
			Bold(true),

		FormSelected: lipgloss.NewStyle().
			Foreground(p.Primary).
			Bold(true),

		// Messages
		SuccessMessage: lipgloss.NewStyle().
			Foreground(p.Success).
			Bold(true),

		ErrorMessage: lipgloss.NewStyle().
			Foreground(p.Error).
			Bold(true),

		WarningMessage: lipgloss.NewStyle().
			Foreground(p.Warning).
			Bold(true),

		InfoMessage: lipgloss.NewStyle().
			Foreground(p.Info),

		// Spinner
		Spinner: lipgloss.NewStyle().
			Foreground(p.Primary),
	}
}
