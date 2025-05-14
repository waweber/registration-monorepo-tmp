package registration

import (
	"encoding/json"
	"slices"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

const (
	STATUS_PENDING  = "pending"
	STATUS_CREATED  = "created"
	STATUS_CANCELED = "canceled"
)

type RegistrationFields struct {
	Id            string     `json:"id"`
	EventId       string     `json:"event_id"`
	Status        string     `json:"status"`
	Version       int        `json:"version"`
	DateCreated   time.Time  `json:"date_created"`
	DateUpdated   *time.Time `json:"date_updated,omitempty"`
	Options       Options    `json:"options"`
	FirstName     *string    `json:"first_name,omitempty"`
	LastName      *string    `json:"last_name,omitempty"`
	PreferredName *string    `json:"preferred_name,omitempty"`
	Nickname      *string    `json:"nickname,omitempty"`
	Number        *int       `json:"number,omitempty"`
	Email         *string    `json:"email,omitempty"`
	AccountId     *string    `json:"account_id,omitempty"`
	CheckedIn     *bool      `json:"checked_in,omitempty"`
	DateCheckedIn *time.Time `json:"date_checked_in,omitempty"`
}

type Options []string

func NewOptions(s ...string) Options {
	asMap := make(map[string]struct{})
	for _, opt := range s {
		asMap[opt] = struct{}{}
	}

	opts := make([]string, 0, len(asMap))

	for k := range asMap {
		opts = append(opts, k)
	}

	slices.Sort(opts)

	return opts
}

func (o *Options) UnmarshalJSON(data []byte) error {
	var asSlice []string
	err := json.Unmarshal(data, &asSlice)
	if err != nil {
		return err
	}
	*o = NewOptions(asSlice...)
	return nil
}

type Registration struct {
	RegistrationFields
	ExtraData map[string]any `json:"extra_data"`
}

var regFieldNames = map[string]struct{}{
	"id":              {},
	"event_id":        {},
	"status":          {},
	"version":         {},
	"date_created":    {},
	"date_updated":    {},
	"options":         {},
	"first_name":      {},
	"last_name":       {},
	"preferred_name":  {},
	"nickname":        {},
	"number":          {},
	"email":           {},
	"account_id":      {},
	"checked_in":      {},
	"date_checked_in": {},
}

func IsRegFieldName(field string) bool {
	_, isFieldName := regFieldNames[field]
	return isFieldName
}

func (f Registration) MarshalJSON() ([]byte, error) {
	asMapBytes, err := json.Marshal(f.RegistrationFields)
	if err != nil {
		return nil, err
	}

	// double-encode is an inefficient and hacky but lazy way to do this...
	var asMap map[string]any
	err = json.Unmarshal(asMapBytes, &asMap)
	if err != nil {
		return nil, err
	}

	for k, v := range f.ExtraData {
		if !IsRegFieldName(k) {
			asMap[k] = v
		}
	}

	finalBytes, err := json.Marshal(asMap)
	return finalBytes, err
}

func (f *Registration) UnmarshalJSON(data []byte) error {
	var asMap map[string]any
	err := json.Unmarshal(data, &asMap)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &f.RegistrationFields)
	if err != nil {
		return err
	}

	f.ExtraData = make(map[string]any)

	for k, v := range asMap {
		if !IsRegFieldName(k) {
			f.ExtraData[k] = v
		}
	}

	return nil
}

// Generate a new random ID
func NewRegistrationId() string {
	return gonanoid.MustGenerate("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz", 14)
}

// Validate a registration status string.
func ValidateStatus(status string) bool {
	return status == STATUS_PENDING || status == STATUS_CREATED || status == STATUS_CANCELED
}
