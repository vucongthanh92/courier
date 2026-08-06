package gateway

import "testing"

func TestBuildAssistantResponsePartsSplitsLongResponses(t *testing.T) {
	parts := BuildAssistantResponseParts("abcdef", 2, map[string]any{"status": "completed"})
	if len(parts) != 3 {
		t.Fatalf("expected 3 parts, got %d", len(parts))
	}
	for _, part := range parts {
		if len([]rune(part.Body)) > 2 {
			t.Fatalf("part body exceeds max runes: %q", part.Body)
		}
		if part.Total != 3 {
			t.Fatalf("expected total 3, got %d", part.Total)
		}
		if part.Metadata["status"] != "completed" {
			t.Fatalf("expected metadata status to be preserved")
		}
	}
}

func TestBuildAssistantResponsePartsUsesFallbackForEmptyText(t *testing.T) {
	parts := BuildAssistantResponseParts(" ", 4000, nil)
	if len(parts) != 1 {
		t.Fatalf("expected 1 part, got %d", len(parts))
	}
	if parts[0].Body == "" {
		t.Fatal("expected fallback body")
	}
}
