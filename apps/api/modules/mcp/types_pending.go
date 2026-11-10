package mcp

import "time"

// PendingItem is one action waiting on the confirm page.
type PendingItem struct {
	ID        string         `json:"id"`
	Kind      string         `json:"kind"`
	ClientID  string         `json:"client_id"`
	Summary   string         `json:"summary"`
	CreatedAt time.Time      `json:"created_at"`
	ExpiresAt time.Time      `json:"expires_at"`
	State     string         `json:"state"`
	Target    *PendingTarget `json:"target,omitempty"`
}

// PendingTarget is the task or project an action would delete, as it is now.
type PendingTarget struct {
	Type        string `json:"type"`
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Ref         string `json:"ref,omitempty"`
	ProjectName string `json:"project_name,omitempty"`
	TaskCount   int    `json:"task_count,omitempty"`
	URL         string `json:"url"`
}

// PendingList is the body of GET /api/mcp/pending.
type PendingList struct {
	Actions []PendingItem `json:"actions"`
}

// PendingResult is the body answered when an action is confirmed or
// dismissed.
type PendingResult struct {
	Status string `json:"status"`
}

// EmptyRequest documents a route that takes no body.
type EmptyRequest struct{}
