package batch

import "fmt"

type InventoryLimiter struct {
	MaxCounts map[string]int
	CurCounts map[string]int
}

func (i *InventoryLimiter) Test(c *Change) ChangeTestResult {
	addedOpts := c.GetAddedOptions()
	for _, opt := range addedOpts {
		max, maxOk := i.MaxCounts[opt]
		cur := i.CurCounts[opt]

		if maxOk && cur < max {
			i.CurCounts[opt] = cur + 1
		} else if maxOk && cur >= max {
			return ChangeTestResult{c, []ChangeError{{Code: ERROR_LIMIT, Detail: fmt.Sprintf("option limit exceeded: %s", opt)}}}
		}
	}
	return ChangeTestResult{c, nil}
}
