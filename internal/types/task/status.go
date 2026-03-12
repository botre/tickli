package task

import (
	"encoding/json"
	"github.com/gookit/color"
)

type Status int

var (
	StatusNormal   Status = 0
	StatusComplete Status = 2
)

func (s *Status) UnmarshalJSON(data []byte) error {
	var status int
	if err := json.Unmarshal(data, &status); err != nil {
		return err
	}
	switch status {
	case int(StatusNormal), int(StatusComplete):
		*s = Status(status)
	default:
		*s = StatusNormal
	}
	return nil
}

func (s Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(int(s))
}

// ColorString returns the status icon with ANSI color for display
func (s Status) ColorString() string {
	switch s {
	case StatusComplete:
		return color.Green.Sprint("☑")
	case StatusNormal:
		return color.White.Sprint("☐")
	default:
		return color.Red.Sprint("☒")
	}
}

// String returns a plain text label (used by cobra for help/defaults)
func (s Status) String() string {
	switch s {
	case StatusComplete:
		return "completed"
	case StatusNormal:
		return "normal"
	default:
		return "unknown"
	}
}
