package registration_test

import (
	"bytes"
	"encoding/json"
	"reflect"
	"registration/internal/registration"
	"testing"
	"time"
)

func TestRegMarshal(t *testing.T) {
	b := []byte("{\"id\": \"test\", \"event_id\": \"test-event\", \"version\": 1, \"status\": \"created\", \"date_created\": \"2020-01-01T12:00:00.123-05:00\", \"misc\": true, \"options\": [\"a\",\"b\",\"a\"]}")

	var reg *registration.Registration
	err := json.Unmarshal(b, &reg)
	if err != nil {
		t.Fatal(err)
	}

	exTime, err := time.Parse(time.RFC3339Nano, "2020-01-01T12:00:00.123-05:00")
	if err != nil {
		panic(err)
	}

	expected := &registration.Registration{
		RegistrationFields: registration.RegistrationFields{
			Id:          "test",
			EventID:     "test-event",
			Version:     1,
			Status:      registration.STATUS_CREATED,
			DateCreated: exTime,
			Options:     registration.NewOptions("a", "b"),
		},
		ExtraData: map[string]any{
			"misc": true,
		},
	}

	if !reflect.DeepEqual(*reg, *expected) {
		t.Fatalf("expected %v, got %v", expected, reg)
	}
}

func TestRegUnmarshal(t *testing.T) {
	exTime, err := time.Parse(time.RFC3339Nano, "2020-01-01T12:00:00.123-05:00")
	if err != nil {
		panic(err)
	}

	reg := &registration.Registration{
		RegistrationFields: registration.RegistrationFields{
			Id:          "test",
			EventID:     "test-event",
			Status:      registration.STATUS_CREATED,
			Version:     1,
			DateCreated: exTime,
			Options:     registration.NewOptions("a", "b", "a"),
		},
		ExtraData: map[string]any{
			"misc": true,
		},
	}

	asBytes, err := json.Marshal(reg)
	if err != nil {
		t.Fatal(err)
	}

	expBytes := []byte("{\"date_created\":\"2020-01-01T12:00:00.123-05:00\",\"event_id\":\"test-event\",\"id\":\"test\",\"misc\":true,\"options\":[\"a\",\"b\"],\"status\":\"created\",\"version\":1}")
	if !bytes.Equal(asBytes, expBytes) {
		t.Fatalf("expected %s, got %s", expBytes, asBytes)
	}
}
