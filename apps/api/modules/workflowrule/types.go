package workflowrule

import "time"

// WorkflowRuleResponse is the JSON wire representation of an automated workflow rule.
type WorkflowRuleResponse struct {
	ID              int64     `json:"id"`
	ProjectID       int64     `json:"projectId"`
	IntegrationType string    `json:"integrationType"`
	EventType       string    `json:"eventType"`
	ColumnID        int64     `json:"columnId"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// UpsertWorkflowRuleRequest carries the payload for configuring an automated rule.
type UpsertWorkflowRuleRequest struct {
	IntegrationType string `json:"integrationType"`
	EventType       string `json:"eventType"`
	ColumnID        int64  `json:"columnId"`
}
