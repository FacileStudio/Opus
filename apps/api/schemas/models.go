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

// Invitation is a pending request to join a workspace, sent by email. It
// becomes a workspace_member row when accepted; the status column is what
// makes accept, reject and cancel all single-row updates.
type Invitation struct {
	ID          int64      `gorm:"column:id;primaryKey;autoIncrement"`
	WorkspaceID int64      `gorm:"column:workspace_id;not null;index"`
	Email       string     `gorm:"column:email;not null;index"`
	Role        string     `gorm:"column:role;not null;default:member"`
	Status      string     `gorm:"column:status;not null;default:pending"`
	InviterID   int64      `gorm:"column:inviter_id;not null;index"`
	CreatedAt   time.Time  `gorm:"column:created_at;not null"`
	ExpiresAt   *time.Time `gorm:"column:expires_at"`
}

// Project is a container for tasks, columns and views within a workspace.
type Project struct {
	ID          int64      `gorm:"column:id;primaryKey;autoIncrement"`
	WorkspaceID int64      `gorm:"column:workspace_id;not null;index"`
	Slug        string     `gorm:"column:slug;not null"`
	Icon        string     `gorm:"column:icon;default:Layout"`
	Name        string     `gorm:"column:name;not null"`
	Description string     `gorm:"column:description"`
	IsPublic    bool       `gorm:"column:is_public;default:false"`
	ArchivedAt  *time.Time `gorm:"column:archived_at"`
	CreatedAt   time.Time  `gorm:"column:created_at;not null"`
}

// Column is a stage in a project's workflow board.
type Column struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement"`
	ProjectID int64     `gorm:"column:project_id;not null;index"`
	Name      string    `gorm:"column:name;not null"`
	Slug      string    `gorm:"column:slug;not null"`
	Position  int       `gorm:"column:position;not null;default:0"`
	Icon      string    `gorm:"column:icon"`
	Color     string    `gorm:"column:color"`
	IsFinal   bool      `gorm:"column:is_final;not null;default:false"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`
}

// Task is a work unit within a project and workspace.
type Task struct {
	ID          int64      `gorm:"column:id;primaryKey;autoIncrement"`
	ProjectID   int64      `gorm:"column:project_id;not null;index"`
	WorkspaceID int64      `gorm:"column:workspace_id;not null;index"`
	ColumnID    *int64     `gorm:"column:column_id;index"`
	ActorID     *int64     `gorm:"column:actor_id;index"`
	Number      int        `gorm:"column:number;not null;default:1"`
	Position    int        `gorm:"column:position;not null;default:0"`
	Title       string     `gorm:"column:title;not null"`
	Description string     `gorm:"column:description"`
	Status      string     `gorm:"column:status;not null;default:to-do"`
	Priority    string     `gorm:"column:priority;not null;default:low"`
	StartDate   *time.Time `gorm:"column:start_date"`
	DueDate     *time.Time `gorm:"column:due_date;index"`
	CreatedAt   time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;not null"`
}

// Label tags tasks within a workspace or specific task.
type Label struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement"`
	WorkspaceID int64     `gorm:"column:workspace_id;not null;index"`
	TaskID      *int64    `gorm:"column:task_id;index"`
	Name        string    `gorm:"column:name;not null"`
	Color       string    `gorm:"column:color;not null"`
	CreatedAt   time.Time `gorm:"column:created_at;not null"`
	UpdatedAt   time.Time `gorm:"column:updated_at;not null"`
}

// Comment is a discussion entry on a task.
type Comment struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement"`
	TaskID    int64     `gorm:"column:task_id;not null;index"`
	UserID    int64     `gorm:"column:user_id;not null;index"`
	Content   string    `gorm:"column:content;not null"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`
}

// TaskRelation defines a dependency between two tasks.
type TaskRelation struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement"`
	SourceTaskID int64     `gorm:"column:source_task_id;not null;index"`
	TargetTaskID int64     `gorm:"column:target_task_id;not null;index"`
	RelationType string    `gorm:"column:relation_type;not null"`
	CreatedAt    time.Time `gorm:"column:created_at;not null"`
}

// Activity records state changes and audit events on a task.
type Activity struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement"`
	TaskID    int64     `gorm:"column:task_id;not null;index"`
	UserID    *int64    `gorm:"column:user_id;index"`
	Type      string    `gorm:"column:type;not null"`
	Content   string    `gorm:"column:content"`
	EventData string    `gorm:"column:event_data"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`
}
