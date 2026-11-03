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
// RegistreSpaceID links it to the Registre shared space it mirrors; a linked
// workspace's name and members belong to Registre.
type Workspace struct {
	ID              int64     `gorm:"column:id;primaryKey;autoIncrement"`
	Name            string    `gorm:"column:name;not null"`
	Slug            string    `gorm:"column:slug;uniqueIndex;not null"`
	Logo            string    `gorm:"column:logo"`
	Description     string    `gorm:"column:description"`
	CreatedAt       time.Time `gorm:"column:created_at;not null"`
	RegistreSpaceID *string   `gorm:"column:registre_space_id;uniqueIndex"`
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
	ID          int64 `gorm:"column:id;primaryKey;autoIncrement"`
	WorkspaceID int64 `gorm:"column:workspace_id;not null;index"`
	FacileRef
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
	ID int64 `gorm:"column:id;primaryKey;autoIncrement"`
	FacileRef
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

// ShareLink provides public read-only access to a project.
type ShareLink struct {
	ID             int64      `gorm:"column:id;primaryKey;autoIncrement"`
	ProjectID      int64      `gorm:"column:project_id;not null;index"`
	Token          string     `gorm:"column:token;uniqueIndex;not null"`
	Label          string     `gorm:"column:label"`
	ExpiresAt      *time.Time `gorm:"column:expires_at"`
	IsActive       bool       `gorm:"column:is_active;not null;default:true"`
	LastAccessedAt *time.Time `gorm:"column:last_accessed_at"`
	AccessCount    int        `gorm:"column:access_count;not null;default:0"`
	CreatedBy      int64      `gorm:"column:created_by;not null;index"`
	CreatedAt      time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;not null"`
}

// TimeEntry logs work duration against a task.
type TimeEntry struct {
	ID          int64      `gorm:"column:id;primaryKey;autoIncrement"`
	TaskID      int64      `gorm:"column:task_id;not null;index"`
	UserID      *int64     `gorm:"column:user_id;index"`
	Description string     `gorm:"column:description"`
	StartTime   time.Time  `gorm:"column:start_time;not null"`
	EndTime     *time.Time `gorm:"column:end_time"`
	Duration    int        `gorm:"column:duration;not null;default:0"`
	CreatedAt   time.Time  `gorm:"column:created_at;not null"`
}

// WorkflowRule triggers automated task actions based on integration events.
type WorkflowRule struct {
	ID              int64     `gorm:"column:id;primaryKey;autoIncrement"`
	ProjectID       int64     `gorm:"column:project_id;not null;index"`
	IntegrationType string    `gorm:"column:integration_type;not null"`
	EventType       string    `gorm:"column:event_type;not null"`
	ColumnID        int64     `gorm:"column:column_id;not null;index"`
	CreatedAt       time.Time `gorm:"column:created_at;not null"`
	UpdatedAt       time.Time `gorm:"column:updated_at;not null"`
}

// Notification represents an in-app alert for a user.
type Notification struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement"`
	UserID       int64     `gorm:"column:user_id;not null;index"`
	Title        string    `gorm:"column:title"`
	Content      string    `gorm:"column:content"`
	Type         string    `gorm:"column:type;not null;default:info"`
	EventData    string    `gorm:"column:event_data"`
	IsRead       bool      `gorm:"column:is_read;not null;default:false"`
	ResourceID   *string   `gorm:"column:resource_id"`
	ResourceType *string   `gorm:"column:resource_type"`
	CreatedAt    time.Time `gorm:"column:created_at;not null"`
}

// UserNotificationPreference stores global notification delivery settings.
type UserNotificationPreference struct {
	ID              int64     `gorm:"column:id;primaryKey;autoIncrement"`
	UserID          int64     `gorm:"column:user_id;uniqueIndex;not null"`
	EmailEnabled    bool      `gorm:"column:email_enabled;not null;default:false"`
	NtfyEnabled     bool      `gorm:"column:ntfy_enabled;not null;default:false"`
	NtfyServerURL   string    `gorm:"column:ntfy_server_url"`
	NtfyTopic       string    `gorm:"column:ntfy_topic"`
	NtfyToken       string    `gorm:"column:ntfy_token"`
	GotifyEnabled   bool      `gorm:"column:gotify_enabled;not null;default:false"`
	GotifyServerURL string    `gorm:"column:gotify_server_url"`
	GotifyToken     string    `gorm:"column:gotify_token"`
	WebhookEnabled  bool      `gorm:"column:webhook_enabled;not null;default:false"`
	WebhookURL      string    `gorm:"column:webhook_url"`
	WebhookSecret   string    `gorm:"column:webhook_secret"`
	CreatedAt       time.Time `gorm:"column:created_at;not null"`
	UpdatedAt       time.Time `gorm:"column:updated_at;not null"`
}

// UserNotificationWorkspaceRule overrides preferences per workspace.
type UserNotificationWorkspaceRule struct {
	ID                 int64     `gorm:"column:id;primaryKey;autoIncrement"`
	UserID             int64     `gorm:"column:user_id;not null;index"`
	WorkspaceID        int64     `gorm:"column:workspace_id;not null;index"`
	IsActive           bool      `gorm:"column:is_active;not null;default:true"`
	EmailEnabled       bool      `gorm:"column:email_enabled;not null;default:false"`
	NtfyEnabled        bool      `gorm:"column:ntfy_enabled;not null;default:false"`
	GotifyEnabled      bool      `gorm:"column:gotify_enabled;not null;default:false"`
	WebhookEnabled     bool      `gorm:"column:webhook_enabled;not null;default:false"`
	ProjectMode        string    `gorm:"column:project_mode;not null;default:all"`
	SelectedProjectIDs string    `gorm:"column:selected_project_ids"`
	CreatedAt          time.Time `gorm:"column:created_at;not null"`
	UpdatedAt          time.Time `gorm:"column:updated_at;not null"`
}

// TaskReminderSent tracks which reminder types have already been dispatched.
type TaskReminderSent struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement"`
	TaskID       int64     `gorm:"column:task_id;not null;uniqueIndex:idx_task_reminder_type"`
	ReminderType string    `gorm:"column:reminder_type;not null;uniqueIndex:idx_task_reminder_type"`
	SentAt       time.Time `gorm:"column:sent_at;not null"`
}

// Asset represents an uploaded file or media asset.
type Asset struct {
	ID          string    `gorm:"primaryKey;size:64" json:"id"`
	WorkspaceID int64     `gorm:"not null;index" json:"workspace_id"`
	ProjectID   int64     `gorm:"not null;index" json:"project_id"`
	TaskID      *int64    `gorm:"index" json:"task_id,omitempty"`
	ActivityID  *int64    `gorm:"index" json:"activity_id,omitempty"`
	ObjectKey   string    `gorm:"not null;uniqueIndex;size:512" json:"object_key"`
	Filename    string    `gorm:"not null;size:255" json:"filename"`
	MimeType    string    `gorm:"not null;size:100" json:"mime_type"`
	Size        int64     `gorm:"not null" json:"size"`
	Kind        string    `gorm:"not null;default:'image';size:32" json:"kind"`
	Surface     string    `gorm:"not null;default:'description';size:32" json:"surface"`
	CreatedBy   *int64    `gorm:"index" json:"created_by,omitempty"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName returns the database table name for Asset.
func (Asset) TableName() string {
	return "asset"
}
