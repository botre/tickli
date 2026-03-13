package utils

import (
	"fmt"
	"time"

	"github.com/sho0pi/naturaltime"
)

var DefaultDuration = 1 * time.Hour

func ParseTimeExpression(expr string) (*naturaltime.Range, error) {
	p, err := naturaltime.New()
	if err != nil {
		return nil, err
	}
	currentTime := time.Now()
	r, err := p.ParseRange(expr, currentTime)
	if err != nil {
		return nil, err
	}

	if !r.IsAllDay() {
		return r, nil
	}
	// Checks if the user just specified the date but no the time.
	if r.Start().Hour() == currentTime.Hour() && r.Start().Minute() == currentTime.Minute() && r.Start().Second() == currentTime.Second() {
		return r, nil
	}
	r.Duration = DefaultDuration
	return r, nil
}

// TruncateToDate strips the time component, keeping only the date at midnight.
func TruncateToDate(t time.Time) time.Time {
	if t.IsZero() {
		return t
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// ParseFlexibleTime parses ISO 8601 timestamps, accepting both +00:00 and +0000 offset formats.
func ParseFlexibleTime(value string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t, nil
	}
	if t, err := time.Parse("2006-01-02T15:04:05-0700", value); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("invalid time format %q: expected ISO 8601 (e.g. 2025-02-18T15:04:05Z or 2025-02-18T15:04:05+02:00)", value)
}
