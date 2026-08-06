package models

type SafetyEvaluationRequest struct {
	Text string `json:"text"`
}

type SafetyEvaluationResult struct {
	Decision    string   `json:"decision"`
	Category    string   `json:"category,omitempty"`
	Reasons     []string `json:"reasons,omitempty"`
	UserMessage string   `json:"user_message,omitempty"`
}
