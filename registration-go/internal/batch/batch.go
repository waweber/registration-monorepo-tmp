package batch

type BatchChange struct {
	Changes []Change
}

// Get all access codes used in this batch
func (b *BatchChange) GetAccessCodes() []string {
	var codes []string
	for _, c := range b.Changes {
		if c.AccessCode != "" {
			codes = append(codes, c.AccessCode)
		}
	}

	return codes
}
