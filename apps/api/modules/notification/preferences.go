package notification

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"strings"
	"time"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
	"gorm.io/gorm"
)

func maskValue(val string) *string {
	trimmed := strings.TrimSpace(val)
	if trimmed == "" {
		return nil
	}
	if len(trimmed) > 8 {
		res := trimmed[:4] + "…" + trimmed[len(trimmed)-4:]
		return &res
	}
	res := "••••"
	return &res
}

// GetPreferences fetches delivery preferences and workspace rules for a user.
func (service *Service) GetPreferences(ctx context.Context, userID int64, userEmail string) (*NotificationPreferencesResponse, error) {
	var pref schemas.UserNotificationPreference
	err := service.db.WithContext(ctx).First(&pref, "user_id = ?", userID).Error
	if err != nil && !stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.Internal("failed to query notification preferences", err)
	}

	type ruleRow struct {
		ID                 int64
		WorkspaceID        int64
		WorkspaceName      string
		IsActive           bool
		EmailEnabled       bool
		NtfyEnabled        bool
		GotifyEnabled      bool
		WebhookEnabled     bool
		ProjectMode        string
		SelectedProjectIDs string
		CreatedAt          time.Time
		UpdatedAt          time.Time
	}

	var rows []ruleRow
	service.db.WithContext(ctx).
		Table("user_notification_workspace_rules").
		Select("user_notification_workspace_rules.*, workspaces.name as workspace_name").
		Joins("JOIN workspaces ON workspaces.id = user_notification_workspace_rules.workspace_id").
		Where("user_notification_workspace_rules.user_id = ?", userID).
		Order("user_notification_workspace_rules.created_at asc").
		Scan(&rows)

	rules := make([]WorkspaceRuleResponse, 0, len(rows))
	for _, r := range rows {
		var selectedIDs []string
		if r.SelectedProjectIDs != "" {
			json.Unmarshal([]byte(r.SelectedProjectIDs), &selectedIDs)
		}
		rules = append(rules, WorkspaceRuleResponse{
			ID:                 r.ID,
			WorkspaceID:        r.WorkspaceID,
			WorkspaceName:      r.WorkspaceName,
			IsActive:           r.IsActive,
			EmailEnabled:       r.EmailEnabled,
			NtfyEnabled:        r.NtfyEnabled,
			GotifyEnabled:      r.GotifyEnabled,
			WebhookEnabled:     r.WebhookEnabled,
			ProjectMode:        r.ProjectMode,
			SelectedProjectIDs: selectedIDs,
			CreatedAt:          r.CreatedAt.Format(time.RFC3339),
			UpdatedAt:          r.UpdatedAt.Format(time.RFC3339),
		})
	}

	return formatPreferences(&pref, rules, userEmail), nil
}

func formatPreferences(pref *schemas.UserNotificationPreference, rules []WorkspaceRuleResponse, email string) *NotificationPreferencesResponse {
	var emailPtr *string
	if email != "" {
		emailPtr = &email
	}
	var createdStr, updatedStr *string
	if !pref.CreatedAt.IsZero() {
		c := pref.CreatedAt.Format(time.RFC3339)
		u := pref.UpdatedAt.Format(time.RFC3339)
		createdStr = &c
		updatedStr = &u
	}
	var ntfyServerURL, ntfyTopic, gotifyServerURL, webhookURL *string
	if pref.NtfyServerURL != "" {
		ntfyServerURL = &pref.NtfyServerURL
	}
	if pref.NtfyTopic != "" {
		ntfyTopic = &pref.NtfyTopic
	}
	if pref.GotifyServerURL != "" {
		gotifyServerURL = &pref.GotifyServerURL
	}
	if pref.WebhookURL != "" {
		webhookURL = &pref.WebhookURL
	}
	return &NotificationPreferencesResponse{
		EmailAddress: emailPtr,
		EmailEnabled: pref.EmailEnabled,
		NtfyPreferences: NtfyPreferences{
			NtfyEnabled:         pref.NtfyEnabled,
			NtfyConfigured:      pref.NtfyServerURL != "" && pref.NtfyTopic != "",
			NtfyServerURL:       ntfyServerURL,
			NtfyTopic:           ntfyTopic,
			NtfyTokenConfigured: pref.NtfyToken != "",
			MaskedNtfyToken:     maskValue(pref.NtfyToken),
		},
		GotifyPreferences: GotifyPreferences{
			GotifyEnabled:         pref.GotifyEnabled,
			GotifyConfigured:      pref.GotifyServerURL != "" && pref.GotifyToken != "",
			GotifyServerURL:       gotifyServerURL,
			GotifyTokenConfigured: pref.GotifyToken != "",
			MaskedGotifyToken:     maskValue(pref.GotifyToken),
		},
		WebhookPreferences: WebhookPreferences{
			WebhookEnabled:          pref.WebhookEnabled,
			WebhookConfigured:       pref.WebhookURL != "",
			WebhookURL:              webhookURL,
			WebhookSecretConfigured: pref.WebhookSecret != "",
			MaskedWebhookSecret:     maskValue(pref.WebhookSecret),
		},
		Workspaces: rules,
		CreatedAt:  createdStr,
		UpdatedAt:  updatedStr,
	}
}

// UpdatePreferences creates or modifies global notification preferences for a user.
func (service *Service) UpdatePreferences(ctx context.Context, userID int64, userEmail string, req *UpdatePreferencesRequest) (*NotificationPreferencesResponse, error) {
	var pref schemas.UserNotificationPreference
	err := service.db.WithContext(ctx).First(&pref, "user_id = ?", userID).Error
	now := time.Now()
	if err != nil {
		if !stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.Internal("failed to query notification preferences", err)
		}
		pref = schemas.UserNotificationPreference{UserID: userID, CreatedAt: now, UpdatedAt: now}
	}
	applyPreferenceUpdates(&pref, req, now)
	if err := service.db.WithContext(ctx).Save(&pref).Error; err != nil {
		return nil, errors.Internal("failed to save notification preferences", err)
	}
	return service.GetPreferences(ctx, userID, userEmail)
}

func applyPreferenceUpdates(pref *schemas.UserNotificationPreference, req *UpdatePreferencesRequest, now time.Time) {
	pref.UpdatedAt = now
	if req.EmailEnabled != nil {
		pref.EmailEnabled = *req.EmailEnabled
	}
	if req.NtfyEnabled != nil {
		pref.NtfyEnabled = *req.NtfyEnabled
	}
	if req.NtfyServerURL != nil {
		pref.NtfyServerURL = *req.NtfyServerURL
	}
	if req.NtfyTopic != nil {
		pref.NtfyTopic = *req.NtfyTopic
	}
	if req.NtfyToken != nil {
		pref.NtfyToken = *req.NtfyToken
	}
	if req.GotifyEnabled != nil {
		pref.GotifyEnabled = *req.GotifyEnabled
	}
	if req.GotifyServerURL != nil {
		pref.GotifyServerURL = *req.GotifyServerURL
	}
	if req.GotifyToken != nil {
		pref.GotifyToken = *req.GotifyToken
	}
	if req.WebhookEnabled != nil {
		pref.WebhookEnabled = *req.WebhookEnabled
	}
	if req.WebhookURL != nil {
		pref.WebhookURL = *req.WebhookURL
	}
	if req.WebhookSecret != nil {
		pref.WebhookSecret = *req.WebhookSecret
	}
}

// UpsertWorkspaceRule creates or updates notification rules for a specific workspace.
func (service *Service) UpsertWorkspaceRule(ctx context.Context, userID, workspaceID int64, userEmail string, req *UpsertWorkspaceRuleRequest) (*NotificationPreferencesResponse, error) {
	var count int64
	service.db.WithContext(ctx).
		Model(&schemas.WorkspaceMember{}).
		Where("user_id = ? AND workspace_id = ?", userID, workspaceID).
		Count(&count)
	if count == 0 {
		return nil, errors.Forbidden("You don't have access to this workspace")
	}

	selectedJSON, _ := json.Marshal(req.SelectedProjectIDs)
	var rule schemas.UserNotificationWorkspaceRule
	err := service.db.WithContext(ctx).
		Where("user_id = ? AND workspace_id = ?", userID, workspaceID).
		First(&rule).Error
	now := time.Now()
	if err == nil {
		rule.IsActive = req.IsActive
		rule.EmailEnabled = req.EmailEnabled
		rule.NtfyEnabled = req.NtfyEnabled
		rule.GotifyEnabled = req.GotifyEnabled
		rule.WebhookEnabled = req.WebhookEnabled
		rule.ProjectMode = req.ProjectMode
		rule.SelectedProjectIDs = string(selectedJSON)
		rule.UpdatedAt = now
		if err := service.db.WithContext(ctx).Save(&rule).Error; err != nil {
			return nil, errors.Internal("failed to update workspace rule", err)
		}
	} else if stderrors.Is(err, gorm.ErrRecordNotFound) {
		rule = schemas.UserNotificationWorkspaceRule{
			UserID:             userID,
			WorkspaceID:        workspaceID,
			IsActive:           req.IsActive,
			EmailEnabled:       req.EmailEnabled,
			NtfyEnabled:        req.NtfyEnabled,
			GotifyEnabled:      req.GotifyEnabled,
			WebhookEnabled:     req.WebhookEnabled,
			ProjectMode:        req.ProjectMode,
			SelectedProjectIDs: string(selectedJSON),
			CreatedAt:          now,
			UpdatedAt:          now,
		}
		if err := service.db.WithContext(ctx).Create(&rule).Error; err != nil {
			return nil, errors.Internal("failed to create workspace rule", err)
		}
	} else {
		return nil, errors.Internal("failed to query workspace rule", err)
	}

	return service.GetPreferences(ctx, userID, userEmail)
}

// DeleteWorkspaceRule removes notification rules for a workspace.
func (service *Service) DeleteWorkspaceRule(ctx context.Context, userID, workspaceID int64, userEmail string) (*NotificationPreferencesResponse, error) {
	var rule schemas.UserNotificationWorkspaceRule
	err := service.db.WithContext(ctx).
		Where("user_id = ? AND workspace_id = ?", userID, workspaceID).
		First(&rule).Error
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("Workspace rule not found")
		}
		return nil, errors.Internal("failed to query workspace rule", err)
	}
	if err := service.db.WithContext(ctx).Delete(&rule).Error; err != nil {
		return nil, errors.Internal("failed to delete workspace rule", err)
	}
	return service.GetPreferences(ctx, userID, userEmail)
}
