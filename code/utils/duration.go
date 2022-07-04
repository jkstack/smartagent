package utils

import (
	"time"
)

type Duration time.Duration

// MarshalKV marshal duration
func (d Duration) MarshalKV() (string, error) {
	return time.Duration(d).String(), nil
}

// UnmarshalKV unmarshal duration
func (d *Duration) UnmarshalKV(value string) error {
	pd, err := time.ParseDuration(value)
	if err != nil {
		return err
	}
	*d = Duration(pd)
	return nil
}

// String format to string
func (d Duration) String() string {
	return time.Duration(d).String()
}

// Duration get duration
func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}
