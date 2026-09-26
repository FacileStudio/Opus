// Package column implements the column management API.
package column

import "time"

// ColumnResponse is the wire shape for a board column.
type ColumnResponse struct {
	ID        int64     `json:"id"`
	ProjectID int64     `json:"projectId"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Position  int       `json:"position"`
	Icon      string    `json:"icon"`
	Color     string    `json:"color"`
	IsFinal   bool      `json:"isFinal"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// CreateColumnRequest is the body of POST /column/{projectId}.
type CreateColumnRequest struct {
	Name    string  `json:"name"`
	Icon    *string `json:"icon,omitempty"`
	Color   *string `json:"color,omitempty"`
	IsFinal *bool   `json:"isFinal,omitempty"`
}

// UpdateColumnRequest is the body of PUT /column/{id}.
type UpdateColumnRequest struct {
	Name    *string `json:"name,omitempty"`
	Icon    *string `json:"icon,omitempty"`
	Color   *string `json:"color,omitempty"`
	IsFinal *bool   `json:"isFinal,omitempty"`
}

// ColumnOrderItem defines the position update for a single column.
type ColumnOrderItem struct {
	ID       int64 `json:"id"`
	Position int   `json:"position"`
}

// ReorderColumnsRequest is the body of PUT /column/reorder/{projectId}.
type ReorderColumnsRequest struct {
	Columns []ColumnOrderItem `json:"columns"`
}
