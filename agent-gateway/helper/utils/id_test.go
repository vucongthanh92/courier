package utils

import "testing"

func TestStableUUID(t *testing.T) {
	first := StableUUID("assistant-123:user")
	second := StableUUID("assistant-123:user")

	if first != second {
		t.Fatalf("expected stable UUID, got %q and %q", first, second)
	}
	if len(first) != 36 {
		t.Fatalf("expected UUID length 36, got %d: %q", len(first), first)
	}
	if first[14] != '5' {
		t.Fatalf("expected version 5 UUID, got %q", first)
	}
	if first[8] != '-' || first[13] != '-' || first[18] != '-' || first[23] != '-' {
		t.Fatalf("expected dashed UUID format, got %q", first)
	}
}
