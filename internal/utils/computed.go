package utils

import (
	"time"

	"github.com/botre/tickli/internal/types"
)

// ComputeFields populates all computed fields on a task before output.
func ComputeFields(t *types.Task) {
	computeDuration(t)
}

func computeDuration(t *types.Task) {
	start := time.Time(t.StartDate)
	due := time.Time(t.DueDate)
	if !start.IsZero() && !due.IsZero() && !start.Equal(due) {
		t.Duration = due.Sub(start).String()
	}
}
