package mcp

import mcpkit "github.com/FacileStudio/tronc/mcp"

// PendingItem is one action waiting on the confirm page.
type PendingItem struct {
	mcpkit.PendingItem
	Target *PendingTarget `json:"target,omitempty"`
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
