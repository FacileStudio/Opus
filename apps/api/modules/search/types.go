package search

import "time"

// SearchMetadata groups container and author context for a search result item.
type SearchMetadata struct {
	ProjectID     *string `json:"projectId,omitempty"`
	ProjectName   *string `json:"projectName,omitempty"`
	ProjectSlug   *string `json:"projectSlug,omitempty"`
	WorkspaceID   *string `json:"workspaceId,omitempty"`
	WorkspaceName *string `json:"workspaceName,omitempty"`
	UserID        *string `json:"userId,omitempty"`
	UserName      *string `json:"userName,omitempty"`
}

// SearchItem is one result item across tasks, projects, workspaces, or comments.
type SearchItem struct {
	ID             string    `json:"id"`
	Type           string    `json:"type"`
	Title          string    `json:"title"`
	Description    *string   `json:"description,omitempty"`
	Content        *string   `json:"content,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	RelevanceScore int       `json:"relevanceScore"`
	TaskNumber     *int      `json:"taskNumber,omitempty"`
	Priority       *string   `json:"priority,omitempty"`
	Status         *string   `json:"status,omitempty"`
	SearchMetadata
}

// SearchResponse is the JSON wire response for a global search query.
type SearchResponse struct {
	Results     []SearchItem `json:"results"`
	TotalCount  int          `json:"totalCount"`
	SearchQuery string       `json:"searchQuery"`
}

// SearchFilter carries normalized search query arguments.
type SearchFilter struct {
	Query       string
	Type        string
	WorkspaceID *int64
	ProjectID   *int64
	Limit       int
}
