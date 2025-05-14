package batch_test

import (
	"registration/internal/access_code"
	"registration/internal/batch"
	"registration/internal/registration"
	"testing"
	"time"
)

func TestChange(t *testing.T) {
	old := registration.Registration{RegistrationFields: registration.RegistrationFields{
		Id:      "1",
		EventId: "a",
		Version: 1,
		Status:  registration.STATUS_PENDING,
		Options: registration.NewOptions("a", "b"),
	}}
	new := registration.Registration{RegistrationFields: registration.RegistrationFields{
		Id:      "1",
		EventId: "a",
		Version: 1,
		Status:  registration.STATUS_CREATED,
		Options: registration.NewOptions("a", "b"),
	}}

	change := &batch.Change{Old: old, New: new}

	limiter := batch.InventoryLimiter{
		MaxCounts: map[string]int{"a": 2, "b": 2},
		CurCounts: map[string]int{"a": 1, "b": 1},
	}

	res := change.Test(limiter.Test)
	if len(res.Errors) != 0 {
		t.Errorf("expected no error, got %v", res.Errors)
	}
}

func TestChangeVersionErr(t *testing.T) {
	old := registration.Registration{RegistrationFields: registration.RegistrationFields{
		Id:      "1",
		EventId: "a",
		Version: 1,
		Status:  registration.STATUS_CREATED,
	}}
	new := registration.Registration{RegistrationFields: registration.RegistrationFields{
		Id:      "1",
		EventId: "a",
		Version: 2,
		Status:  registration.STATUS_CREATED,
	}}

	change := &batch.Change{Old: old, New: new}
	res := change.Test()
	if len(res.Errors) != 1 || res.Errors[0].Code != batch.ERROR_VERSION {
		t.Errorf("expected error, got %v", res.Errors)
	}
}

func TestChangeIDErr(t *testing.T) {
	old := registration.Registration{RegistrationFields: registration.RegistrationFields{
		Id:      "1",
		Version: 1,
		Status:  registration.STATUS_CREATED,
	}}
	new := registration.Registration{RegistrationFields: registration.RegistrationFields{
		Id:      "2",
		Version: 1,
		Status:  registration.STATUS_CREATED,
	}}

	change := &batch.Change{Old: old, New: new}
	res := change.Test()
	if len(res.Errors) != 1 || res.Errors[0].Code != batch.ERROR_ID {
		t.Errorf("expected error, got %v", res.Errors)
	}
}

func TestChangeEventErr(t *testing.T) {
	old := registration.Registration{RegistrationFields: registration.RegistrationFields{
		Id:      "1",
		EventId: "a",
		Version: 1,
		Status:  registration.STATUS_CREATED,
	}}
	new := registration.Registration{RegistrationFields: registration.RegistrationFields{
		Id:      "1",
		EventId: "b",
		Version: 1,
		Status:  registration.STATUS_CREATED,
	}}

	change := &batch.Change{Old: old, New: new}
	res := change.Test()
	if len(res.Errors) != 1 || res.Errors[0].Code != batch.ERROR_EVENT {
		t.Errorf("expected error, got %v", res.Errors)
	}
}

func TestChangeCancelErr(t *testing.T) {
	old := registration.Registration{RegistrationFields: registration.RegistrationFields{
		Id:      "1",
		EventId: "a",
		Version: 1,
		Status:  registration.STATUS_CANCELED,
	}}
	new := registration.Registration{RegistrationFields: registration.RegistrationFields{
		Id:      "1",
		EventId: "a",
		Version: 1,
		Status:  registration.STATUS_CREATED,
	}}

	change := &batch.Change{Old: old, New: new}
	res := change.Test()
	if len(res.Errors) != 1 || res.Errors[0].Code != batch.ERROR_STATUS {
		t.Errorf("expected error, got %v", res.Errors)
	}
}

func TestChangePendingErr(t *testing.T) {
	old := registration.Registration{RegistrationFields: registration.RegistrationFields{
		Id:      "1",
		EventId: "a",
		Version: 1,
		Status:  registration.STATUS_CREATED,
	}}
	new := registration.Registration{RegistrationFields: registration.RegistrationFields{
		Id:      "1",
		EventId: "a",
		Version: 1,
		Status:  registration.STATUS_PENDING,
	}}

	change := &batch.Change{Old: old, New: new}
	res := change.Test()
	if len(res.Errors) != 1 || res.Errors[0].Code != batch.ERROR_STATUS {
		t.Errorf("expected error, got %v", res.Errors)
	}
}

func TestChangeOptionLimitErr(t *testing.T) {
	old := registration.Registration{RegistrationFields: registration.RegistrationFields{
		Id:      "1",
		EventId: "a",
		Version: 1,
		Status:  registration.STATUS_PENDING,
		Options: registration.NewOptions("a", "b"),
	}}
	new := registration.Registration{RegistrationFields: registration.RegistrationFields{
		Id:      "1",
		EventId: "a",
		Version: 1,
		Status:  registration.STATUS_CREATED,
		Options: registration.NewOptions("a", "b"),
	}}

	change := &batch.Change{Old: old, New: new}

	limiter := batch.InventoryLimiter{
		MaxCounts: map[string]int{"a": 2, "b": 2},
		CurCounts: map[string]int{"a": 1, "b": 2},
	}

	res := change.Test(limiter.Test)
	if len(res.Errors) != 1 || res.Errors[0].Code != batch.ERROR_LIMIT {
		t.Errorf("expected error, got %v", res.Errors)
	}
}

func TestChangeInvalidAccessCodeErr(t *testing.T) {
	old := registration.Registration{RegistrationFields: registration.RegistrationFields{
		Id:      "1",
		EventId: "a",
		Version: 1,
		Status:  registration.STATUS_PENDING,
	}}
	new := registration.Registration{RegistrationFields: registration.RegistrationFields{
		Id:      "1",
		EventId: "a",
		Version: 1,
		Status:  registration.STATUS_CREATED,
	}}

	change := &batch.Change{Old: old, New: new, AccessCode: "test"}

	codeTester := &batch.AccessCodeTester{
		AccessCodes: map[string]access_code.AccessCode{},
	}

	res := change.Test(codeTester.Test)

	if len(res.Errors) != 1 || res.Errors[0].Code != batch.ERROR_ACCESS_CODE {
		t.Errorf("expected error, got %v", res.Errors)
	}
}

func TestChangeExpiredAccessCodeErr(t *testing.T) {
	old := registration.Registration{RegistrationFields: registration.RegistrationFields{
		Id:      "1",
		EventId: "a",
		Version: 1,
		Status:  registration.STATUS_PENDING,
	}}
	new := registration.Registration{RegistrationFields: registration.RegistrationFields{
		Id:      "1",
		EventId: "a",
		Version: 1,
		Status:  registration.STATUS_CREATED,
	}}

	change := &batch.Change{Old: old, New: new, AccessCode: "test"}

	codeTester := &batch.AccessCodeTester{
		AccessCodes: map[string]access_code.AccessCode{
			"test": {
				Used:        false,
				DateExpires: time.Now().Add(-10 * time.Second),
			},
		},
	}

	res := change.Test(codeTester.Test)

	if len(res.Errors) != 1 || res.Errors[0].Code != batch.ERROR_ACCESS_CODE {
		t.Errorf("expected error, got %v", res.Errors)
	}
}

func TestValidAccessCode(t *testing.T) {
	old := registration.Registration{RegistrationFields: registration.RegistrationFields{
		Id:      "1",
		EventId: "a",
		Version: 1,
		Status:  registration.STATUS_PENDING,
	}}
	new := registration.Registration{RegistrationFields: registration.RegistrationFields{
		Id:      "1",
		EventId: "a",
		Version: 1,
		Status:  registration.STATUS_CREATED,
	}}

	change := &batch.Change{Old: old, New: new, AccessCode: "test"}

	codeTester := &batch.AccessCodeTester{
		AccessCodes: map[string]access_code.AccessCode{
			"test": {
				Used:        false,
				DateExpires: time.Now().Add(10 * time.Second),
			},
		},
	}

	res := change.Test(codeTester.Test)

	if len(res.Errors) != 0 {
		t.Errorf("expected no error, got %v", res.Errors)
	}
}
