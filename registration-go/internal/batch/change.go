package batch

import (
	"fmt"
	"registration/internal/registration"
	"slices"
)

var (
	ERROR_VERSION     = "version"
	ERROR_ID          = "id"
	ERROR_EVENT       = "event"
	ERROR_STATUS      = "status"
	ERROR_LIMIT       = "limit"
	ERROR_ACCESS_CODE = "access-code"
)

type Change struct {
	Old        *registration.Registration
	New        *registration.Registration
	AccessCode string
}

func (c *Change) GetAddedOptions() []string {
	if c.Old.Status == registration.STATUS_PENDING && c.New.Status == registration.STATUS_CREATED {
		return slices.Clone(c.New.Options)
	} else if c.New.Status == registration.STATUS_CREATED {
		return slices.DeleteFunc(c.New.Options, func(o string) bool { return slices.Contains(c.Old.Options, o) })
	} else {
		return nil
	}
}

// Test whether the change can be applied.
func (c *Change) Test(extraTests ...ChangeTest) ChangeTestResult {
	allTests := []ChangeTest{checkVersion, checkChangedID, checkChangedEvent, checkInvalidStatus}
	allTests = append(allTests, extraTests...)

	cur := ChangeTestResult{c, nil}
	for _, test := range allTests {
		cur = bindTest(cur, test)
	}
	return cur
}

type ChangeError struct {
	Code   string `json:"code"`
	Detail string `json:"detail,omitempty"`
}

func (e *ChangeError) Error() string {
	return e.Detail
}

type ChangeTestResult struct {
	Change *Change
	Errors []ChangeError
}

type ChangeTest func(c *Change) ChangeTestResult

func bindTest(r ChangeTestResult, t ChangeTest) ChangeTestResult {
	res := t(r.Change)
	fullErrs := append(r.Errors, res.Errors...)
	return ChangeTestResult{res.Change, fullErrs}
}

func checkVersion(c *Change) ChangeTestResult {
	if c.New.Version != c.Old.Version {
		return ChangeTestResult{c, []ChangeError{{Code: ERROR_VERSION, Detail: fmt.Sprintf("expected %v, got %v", c.Old.Version, c.New.Version)}}}
	}
	return ChangeTestResult{c, nil}
}

func checkChangedID(c *Change) ChangeTestResult {
	if c.New.Id != c.Old.Id {
		return ChangeTestResult{c, []ChangeError{{Code: ERROR_ID, Detail: "cannot change ID"}}}
	}
	return ChangeTestResult{c, nil}
}

func checkChangedEvent(c *Change) ChangeTestResult {
	if c.New.EventID != c.Old.EventID {
		return ChangeTestResult{c, []ChangeError{{Code: ERROR_EVENT, Detail: "cannot change event ID"}}}
	}
	return ChangeTestResult{c, nil}
}

func checkInvalidStatus(c *Change) ChangeTestResult {
	if c.New.Status != registration.STATUS_PENDING && c.New.Status != registration.STATUS_CREATED && c.New.Status != registration.STATUS_CANCELED {
		return ChangeTestResult{c, []ChangeError{{Code: ERROR_STATUS, Detail: "invalid status"}}}
	}
	if c.Old.Status == registration.STATUS_CANCELED && c.New.Status != registration.STATUS_CANCELED {
		return ChangeTestResult{c, []ChangeError{{Code: ERROR_STATUS, Detail: "registration is canceled"}}}
	}
	if c.New.Status == registration.STATUS_PENDING && c.Old.Status != registration.STATUS_PENDING {
		return ChangeTestResult{c, []ChangeError{{Code: ERROR_STATUS, Detail: "registration is already in a terminal state"}}}
	}
	return ChangeTestResult{c, nil}
}
