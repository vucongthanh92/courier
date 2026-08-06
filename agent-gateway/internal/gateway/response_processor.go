package gateway

import (
	"strings"
	"unicode/utf8"

	"github.com/vucongthanh92/courier/agent-gateway/helper/constants"
	"github.com/vucongthanh92/courier/agent-gateway/internal/domain/models"
)

func BuildAssistantResponseParts(text string, maxRunes int, baseMetadata map[string]any) []models.AssistantMessagePart {
	text = strings.TrimSpace(text)
	if text == "" {
		text = "Mình chưa thể tạo câu trả lời vào lúc này."
	}
	if maxRunes <= 0 {
		maxRunes = constants.DefaultMaxMessageRunes
	}

	chunks := splitRunes(text, maxRunes)
	parts := make([]models.AssistantMessagePart, 0, len(chunks))
	for idx, chunk := range chunks {
		metadata := cloneMetadata(baseMetadata)
		metadata["part_index"] = idx + 1
		metadata["part_total"] = len(chunks)
		metadata["is_truncated"] = len(chunks) > 1
		parts = append(parts, models.AssistantMessagePart{
			Body:     chunk,
			Index:    idx + 1,
			Total:    len(chunks),
			Metadata: metadata,
		})
	}
	return parts
}

func splitRunes(text string, maxRunes int) []string {
	if utf8.RuneCountInString(text) <= maxRunes {
		return []string{text}
	}

	runes := []rune(text)
	chunks := make([]string, 0, (len(runes)/maxRunes)+1)
	for start := 0; start < len(runes); start += maxRunes {
		end := start + maxRunes
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, strings.TrimSpace(string(runes[start:end])))
	}
	return chunks
}

func cloneMetadata(input map[string]any) map[string]any {
	output := map[string]any{}
	for key, value := range input {
		output[key] = value
	}
	return output
}
