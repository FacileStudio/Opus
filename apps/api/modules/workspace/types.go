package workspace

import "time"

// WorkspaceResponse is the workspace shape GET / returns. Fields match the
// legacy controller's select, ids int64 on the wire.
type WorkspaceResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Logo        string    `json:"logo"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
}

// MemberResponse is one row of GET /{workspaceId}/members. ID is the member
// user's id, matching the legacy join's select of userTable.id.
type MemberResponse struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Image string `json:"image"`
	Role  string `json:"role"`
}

// UploadLogoResponse is what POST /{workspaceId}/logo answers.
type UploadLogoResponse struct {
	URL     string `json:"url"`
	Message string `json:"message"`
}

// DeleteLogoResponse is what DELETE /{workspaceId}/logo answers.
type DeleteLogoResponse struct {
	Message string `json:"message"`
}
