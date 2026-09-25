package models

type AuditLogInput struct {
	ActorType    string
	ActorID      *string
	Action       string
	ResourceType string
	ResourceID   string
	IP           *string
	Metadata     []byte
}
