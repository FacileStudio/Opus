package notification

import "time"

// NotificationResponse is the JSON wire representation of an in-app notification.
type NotificationResponse struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"userId"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	Type         string    `json:"type"`
	EventData    any       `json:"eventData,omitempty"`
	IsRead       bool      `json:"isRead"`
	ResourceID   *string   `json:"resourceId,omitempty"`
	ResourceType *string   `json:"resourceType,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
}

// CreateNotificationRequest carries the payload to record an in-app notification.
type CreateNotificationRequest struct {
	Title             *string `json:"title,omitempty"`
	Message           *string `json:"message,omitempty"`
	Content           *string `json:"content,omitempty"`
	Type              string  `json:"type"`
	EventData         any     `json:"eventData,omitempty"`
	ResourceID        *string `json:"resourceId,omitempty"`
	ResourceType      *string `json:"resourceType,omitempty"`
	RelatedEntityID   *string `json:"relatedEntityId,omitempty"`
	RelatedEntityType *string `json:"relatedEntityType,omitempty"`
}

// BulkResultResponse reports the outcome of a batch notification mutation.
type BulkResultResponse struct {
	Success bool  `json:"success"`
	Count   int64 `json:"count"`
}

// WorkspaceRuleResponse describes notification settings scoped to one workspace.
type WorkspaceRuleResponse struct {
	ID                 int64    `json:"id"`
	WorkspaceID        int64    `json:"workspaceId"`
	WorkspaceName      string   `json:"workspaceName"`
	IsActive           bool     `json:"isActive"`
	EmailEnabled       bool     `json:"emailEnabled"`
	NtfyEnabled        bool     `json:"ntfyEnabled"`
	GotifyEnabled      bool     `json:"gotifyEnabled"`
	WebhookEnabled     bool     `json:"webhookEnabled"`
	ProjectMode        string   `json:"projectMode"`
	SelectedProjectIDs []string `json:"selectedProjectIds"`
	CreatedAt          string   `json:"createdAt"`
	UpdatedAt          string   `json:"updatedAt"`
}

// NtfyPreferences describes ntfy push notification configuration.
type NtfyPreferences struct {
	NtfyEnabled         bool    `json:"ntfyEnabled"`
	NtfyConfigured      bool    `json:"ntfyConfigured"`
	NtfyServerURL       *string `json:"ntfyServerUrl"`
	NtfyTopic           *string `json:"ntfyTopic"`
	NtfyTokenConfigured bool    `json:"ntfyTokenConfigured"`
	MaskedNtfyToken     *string `json:"maskedNtfyToken"`
}

// GotifyPreferences describes gotify push notification configuration.
type GotifyPreferences struct {
	GotifyEnabled         bool    `json:"gotifyEnabled"`
	GotifyConfigured      bool    `json:"gotifyConfigured"`
	GotifyServerURL       *string `json:"gotifyServerUrl"`
	GotifyTokenConfigured bool    `json:"gotifyTokenConfigured"`
	MaskedGotifyToken     *string `json:"maskedGotifyToken"`
}

// WebhookPreferences describes outgoing webhook notification configuration.
type WebhookPreferences struct {
	WebhookEnabled          bool    `json:"webhookEnabled"`
	WebhookConfigured       bool    `json:"webhookConfigured"`
	WebhookURL              *string `json:"webhookUrl"`
	WebhookSecretConfigured bool    `json:"webhookSecretConfigured"`
	MaskedWebhookSecret     *string `json:"maskedWebhookSecret"`
}

// NotificationPreferencesResponse describes global and workspace notification settings.
type NotificationPreferencesResponse struct {
	EmailAddress *string `json:"emailAddress"`
	EmailEnabled bool    `json:"emailEnabled"`
	NtfyPreferences
	GotifyPreferences
	WebhookPreferences
	Workspaces []WorkspaceRuleResponse `json:"workspaces"`
	CreatedAt  *string                 `json:"createdAt"`
	UpdatedAt  *string                 `json:"updatedAt"`
}

// UpdatePreferencesRequest carries global notification delivery configuration.
type UpdatePreferencesRequest struct {
	EmailEnabled    *bool   `json:"emailEnabled,omitempty"`
	NtfyEnabled     *bool   `json:"ntfyEnabled,omitempty"`
	NtfyServerURL   *string `json:"ntfyServerUrl,omitempty"`
	NtfyTopic       *string `json:"ntfyTopic,omitempty"`
	NtfyToken       *string `json:"ntfyToken,omitempty"`
	GotifyEnabled   *bool   `json:"gotifyEnabled,omitempty"`
	GotifyServerURL *string `json:"gotifyServerUrl,omitempty"`
	GotifyToken     *string `json:"gotifyToken,omitempty"`
	WebhookEnabled  *bool   `json:"webhookEnabled,omitempty"`
	WebhookURL      *string `json:"webhookUrl,omitempty"`
	WebhookSecret   *string `json:"webhookSecret,omitempty"`
}

// UpsertWorkspaceRuleRequest carries workspace notification override settings.
type UpsertWorkspaceRuleRequest struct {
	IsActive           bool     `json:"isActive"`
	EmailEnabled       bool     `json:"emailEnabled"`
	NtfyEnabled        bool     `json:"ntfyEnabled"`
	GotifyEnabled      bool     `json:"gotifyEnabled"`
	WebhookEnabled     bool     `json:"webhookEnabled"`
	ProjectMode        string   `json:"projectMode"`
	SelectedProjectIDs []string `json:"selectedProjectIds"`
}
