package dto

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"time"
)

type Date struct {
	time.Time
}

const dateLayout = "2006-01-02"

func (d *Date) UnmarshalJSON(b []byte) error {
	s := string(b)

	if s == "null" || s == `""` || s == "" {
		d.Time = time.Time{}
		return nil
	}

	s, err := strconv.Unquote(s)
	if err != nil {
		return fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}

	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}

	d.Time = t
	return nil
}

func (d Date) MarshalJSON() ([]byte, error) {
	if d.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(`"` + d.Time.Format(dateLayout) + `"`), nil
}

func (d *Date) Scan(value any) error {
	if value == nil {
		d.Time = time.Time{}
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		d.Time = v
		return nil
	case string:
		t, err := time.Parse(dateLayout, v)
		if err != nil {
			return err
		}
		d.Time = t
		return nil
	default:
		return fmt.Errorf("cannot scan type %T into Date", value)
	}
}

func (d Date) Value() (driver.Value, error) {
	if d.Time.IsZero() {
		return nil, nil
	}
	return d.Time, nil
}

func (d Date) IsZero() bool {
	return d.Time.IsZero()
}

func (d *Date) ToTimePtr() *time.Time {
	if d == nil {
		return nil
	}
	if d.IsZero() {
		return nil
	}
	return &d.Time
}

func FromTimePtr(t *time.Time) *Date {
	if t == nil {
		return nil
	}
	return &Date{Time: *t}
}
