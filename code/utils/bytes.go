package utils

import "github.com/dustin/go-humanize"

// Bytes yaml bytes
type Bytes uint64

// MarshalKV marshal bytes
func (b Bytes) MarshalKV() (string, error) {
	return humanize.Bytes(uint64(b)), nil
}

// UnmarshalKV unmarshal bytes
func (data *Bytes) UnmarshalKV(value string) error {
	n, err := humanize.ParseBytes(value)
	if err != nil {
		return err
	}
	*data = Bytes(n)
	return nil
}

// Bytes get bytes data
func (data *Bytes) Bytes() uint64 {
	return uint64(*data)
}

// String format to string
func (data Bytes) String() string {
	return humanize.Bytes(uint64(data))
}
