// Package schemas owns the Go backend's tables. The Go backend keeps its own
// schema on its own database: Opus's legacy API is the only suite app still on
// CUID2 text ids, and porte's auth contract is int64 end to end, so the port
// adopts the suite's int64 identity model rather than a surrogate mapping.
package schemas

import "time"

// User is one account. It is the app-side half of a porte identity: porte
// verifies the credential, this row carries the profile and the workspace
// memberships.
type User struct {
	ID            int64     `gorm:"column:id;primaryKey;autoIncrement"`
	Email         string    `gorm:"column:email;uniqueIndex;not null"`
	EmailVerified bool      `gorm:"column:email_verified;not null;default:false"`
	Name          string    `gorm:"column:name"`
	Image         string    `gorm:"column:image"`
	Color         string    `gorm:"column:color"`
	Locale        string    `gorm:"column:locale"`
	CreatedAt     time.Time `gorm:"column:created_at;not null"`
	UpdatedAt     time.Time `gorm:"column:updated_at;not null"`
}

// Workspace is a tenant: one team's projects, columns and tasks.
type Workspace struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement"`
	Name        string    `gorm:"column:name;not null"`
	Slug        string    `gorm:"column:slug;uniqueIndex;not null"`
	Logo        string    `gorm:"column:logo"`
	Description string    `gorm:"column:description"`
	CreatedAt   time.Time `gorm:"column:created_at;not null"`
}

// WorkspaceMember is one human's membership in one workspace.
type WorkspaceMember struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement"`
	WorkspaceID int64     `gorm:"column:workspace_id;not null;index"`
	UserID      int64     `gorm:"column:user_id;not null;index"`
	Role        string    `gorm:"column:role;not null;default:member"`
	JoinedAt    time.Time `gorm:"column:joined_at;not null"`
}
