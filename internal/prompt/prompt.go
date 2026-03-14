package prompt

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

var reader = bufio.NewReader(os.Stdin)

// IsInteractive returns true when stdin is a terminal.
// Commands should skip confirmation prompts when this is false,
// enabling scripted and agentic usage without --force.
func IsInteractive() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd()))
}

// String prompts the user for text input. If defaultVal is non-empty, it is
// shown in brackets and returned when the user presses Enter without typing.
func String(label, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("%s [%s]: ", label, defaultVal)
	} else {
		fmt.Printf("%s: ", label)
	}
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultVal
	}
	return input
}

