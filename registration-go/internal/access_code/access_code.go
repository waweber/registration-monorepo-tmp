package access_code

import "time"

type AccessCode struct {
	Code        string
	EventID     string
	DateCreated time.Time
	DateExpires time.Time
	Name        string
	Used        bool
	Options     map[string]any
}

func (a *AccessCode) IsValid() bool {
	return a.IsValidAt(time.Now())
}

func (a *AccessCode) IsValidAt(t time.Time) bool {
	return !a.Used && t.Before(a.DateExpires)
}
