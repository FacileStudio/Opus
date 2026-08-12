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

// CreateWorkspaceRequest is the body of POST /workspace.
type CreateWorkspaceRequest struct {
	Name        string  `json:"name"`
	Slug        string  `json:"slug,omitempty"`
	Description *string `json:"description,omitempty"`
}

// UpdateWorkspaceRequest is the body of PATCH /workspace/{id}. Every field is
// a pointer so an omitted key is not silently cleared.
type UpdateWorkspaceRequest struct {
	Name        *string `json:"name,omitempty"`
	Slug        *string `json:"slug,omitempty"`
	Description *string `json:"description,omitempty"`
}

// UpdateMemberRoleRequest is the body of PATCH /workspace/{id}/members/{userId}.
type UpdateMemberRoleRequest struct {
	Role string `json:"role"`
}

// InviteMemberRequest is the body of POST /workspace/{id}/members. The role
// defaults to member when omitted.
type InviteMemberRequest struct {
	Email string `json:"email"`
	Role  string `json:"role,omitempty"`
}

// MessageResponse is the bare {message} body the mutating endpoints answer
// when there is no resource to return.
type MessageResponse struct {
	Message string `json:"message"`
}

// FullWorkspaceResponse is what GET /workspace/{id}/full answers: the
// workspace row plus its members and its invitations, flattened for the
// frontend's organization view.
type FullWorkspaceResponse struct {
	Workspace   *WorkspaceResponse  `json:"workspace"`
	Members     []MemberResponse    `json:"members"`
	Invitations []InvitationResponse `json:"invitations"`
}

// InvitationResponse is one invitation row joined to its workspace and its
// inviter, so the frontend can render the invite without a second request.
// InviterID stays internal: the cancel check reads it, the wire does not.
type InvitationResponse struct {
	ID            int64      `json:"id"`
	Email         string     `json:"email"`
	Role          string     `json:"role"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"createdAt"`
	ExpiresAt     *time.Time `json:"expiresAt"`
	WorkspaceID   int64      `json:"workspaceId"`
	WorkspaceName string     `json:"workspaceName"`
	InviterName   string     `json:"inviterName"`
	InviterID     int64      `json:"-"`
}

// AcceptInvitationResponse is what POST /invitation/{id}/accept answers. The
// workspace lets the accept page navigate straight into the workspace the
// invitation belonged to.
type AcceptInvitationResponse struct {
	Workspace *WorkspaceResponse `json:"workspace,omitempty"`
	Message   string             `json:"message"`
}
