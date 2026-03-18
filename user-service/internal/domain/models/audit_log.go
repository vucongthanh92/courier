package models

type AuditLogRequest struct {
	CreatorID uint64                 `json:"creator_id"`
	Action    string                 `json:"action"`
	IP        string                 `json:"ip"`
	UserAgent string                 `json:"user_agent"`
	Metadata  map[string]interface{} `json:"metadata"`
}
