package models

type GenerateAnswerRequest struct {
	ContextPackage ContextPackage
	WebSearch      bool
}

type GenerateAnswerResponse struct {
	Text       string         `json:"text"`
	Model      string         `json:"model"`
	ResponseID string         `json:"response_id,omitempty"`
	Usage      map[string]any `json:"usage,omitempty"`
}
