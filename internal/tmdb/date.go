package tmdb

import (
	"encoding/json"
	"strings"
	"time"
)

type Date struct {
	Time  time.Time
	Valid bool
}

func (d *Date) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	s = strings.TrimSpace(s)
	if s == "" {
		d.Valid = false
		return nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		d.Valid = false
		return nil
	}
	d.Time = t
	d.Valid = true
	return nil
}

func (d Date) Ptr() *time.Time {
	if !d.Valid {
		return nil
	}
	t := d.Time
	return &t
}