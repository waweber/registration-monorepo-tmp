package access_code

import "time"

type AccessCodeInfo struct {
	Used        bool
	DateExpires time.Time
}

func (a *AccessCodeInfo) IsValid() bool {
	return a.IsValidAt(time.Now())
}

func (a *AccessCodeInfo) IsValidAt(t time.Time) bool {
	return !a.Used && t.Before(a.DateExpires)
}
