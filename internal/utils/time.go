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

// ParseFlexibleTime parses time values in multiple formats:
// - ISO 8601 with timezone (e.g. 2025-02-18T15:04:05+02:00)
// - ISO 8601 with compact offset (e.g. 2025-02-18T15:04:05+0200)
// - Plain date (e.g. 2025-02-18) — treated as midnight local time
// - Natural language (e.g. "tomorrow", "next friday 5pm") — returns start of the parsed range
func ParseFlexibleTime(value string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t, nil
	}
	if t, err := time.Parse("2006-01-02T15:04:05-0700", value); err == nil {
		return t, nil
	}
	if t, err := time.ParseInLocation("2006-01-02", value, time.Local); err == nil {
		return t, nil
	}
	if r, err := ParseTimeExpression(value); err == nil {
		return r.Start(), nil
	}
	return time.Time{}, fmt.Errorf("invalid time format %q: expected ISO 8601 (e.g. 2025-02-18T15:04:05Z), plain date (e.g. 2025-02-18), or natural language (e.g. 'tomorrow')", value)
}
