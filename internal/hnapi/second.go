package hnapi

import (
	"encoding/json"
	"strconv"
	"time"
)

var (
	_ json.Unmarshaler = (*TimestampSecond)(nil)
	_ json.Marshaler   = TimestampSecond{}
)

// TimestampSecond implements json encoding/decoding using seconds since EPOCH.
type TimestampSecond time.Time

func (ts TimestampSecond) String() string {
	return ts.ToTime().String()
}

// ToTime converts TimestampSecond back to time.Time.
func (ts TimestampSecond) ToTime() time.Time {
	return time.Time(ts)
}

// UnmarshalJSON implements json.Unmarshaler.
func (ts *TimestampSecond) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*ts = TimestampSecond{}
		return nil
	}

	s, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return err
	}
	*ts = TimestampSecond(time.Unix(s, 0))
	return nil
}

// MarshalJSON implements json.Marshaler.
func (ts TimestampSecond) MarshalJSON() ([]byte, error) {
	t := ts.ToTime()
	if t.IsZero() {
		return []byte("null"), nil
	}

	return []byte(strconv.FormatInt(t.Unix(), 10)), nil
}
