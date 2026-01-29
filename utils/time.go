package utils

import (
	"errors"
	"time"
)

var (
	ErrInvalidTimeFormat = errors.New("invalid time format, expected RFC3339")
	ErrInvalidDateRange  = errors.New("start time must be before end time")
)

func ValidateTimeRange(startStr, endStr, timezone string) (start, end time.Time, err error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	start, err = time.ParseInLocation(time.RFC3339, startStr, loc)
	if err != nil {
		return time.Time{}, time.Time{}, ErrInvalidTimeFormat
	}

	end, err = time.ParseInLocation(time.RFC3339, endStr, loc)
	if err != nil {
		return time.Time{}, time.Time{}, ErrInvalidTimeFormat
	}

	if start.After(end) {
		return time.Time{}, time.Time{}, ErrInvalidDateRange
	}

	return start, end, nil
}
