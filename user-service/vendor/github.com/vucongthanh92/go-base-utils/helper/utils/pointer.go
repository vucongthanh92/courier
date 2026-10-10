package utils

// Uint64Ptr returns a pointer to the given uint64, or nil if the value is 0.
func Uint64Ptr(s uint64) *uint64 {
	if s == 0 {
		return nil
	}
	return &s
}

// StrPtr returns a pointer to the given string, or nil if the string is empty.
func StringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func StringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
