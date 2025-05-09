package batch

import "registration/internal/access_code"

type AccessCodeTester struct {
	AccessCodes map[string]*access_code.AccessCodeInfo
}

func (a *AccessCodeTester) Test(c *Change) ChangeTestResult {
	if c.AccessCode == "" {
		return ChangeTestResult{c, nil}
	}

	info, ok := a.AccessCodes[c.AccessCode]
	if !ok || !info.IsValid() {
		return ChangeTestResult{c, []ChangeError{{Code: ERROR_ACCESS_CODE, Detail: "access code not found"}}}
	}

	return ChangeTestResult{c, nil}
}
