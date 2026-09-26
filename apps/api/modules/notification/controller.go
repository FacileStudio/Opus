package notification

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
)

// Controller adapts HTTP requests for notifications and preferences.
type Controller struct {
	service *Service
}

func newController(service *Service) *Controller {
	return &Controller{service: service}
}

func (controller *Controller) list(ctx context.Context, userID int64) ([]NotificationResponse, error) {
	items, err := controller.service.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	responses := make([]NotificationResponse, 0, len(items))
	for i := range items {
		responses = append(responses, toNotificationResponse(&items[i]))
	}
	return responses, nil
}

func (controller *Controller) create(ctx context.Context, userID int64, req *CreateNotificationRequest) (*NotificationResponse, error) {
	req.Type = strings.TrimSpace(req.Type)
	if req.Type == "" {
		return nil, errors.Invalid("type is required")
	}
	n, err := controller.service.Create(ctx, userID, req)
	if err != nil {
		return nil, err
	}
	resp := toNotificationResponse(n)
	return &resp, nil
}

func (controller *Controller) markAsRead(ctx context.Context, userID, notificationID int64) (*NotificationResponse, error) {
	if notificationID <= 0 {
		return nil, errors.Invalid("invalid notification id")
	}
	n, err := controller.service.MarkAsRead(ctx, userID, notificationID)
	if err != nil {
		return nil, err
	}
	resp := toNotificationResponse(n)
	return &resp, nil
}

func (controller *Controller) markAllAsRead(ctx context.Context, userID int64) (*BulkResultResponse, error) {
	count, err := controller.service.MarkAllAsRead(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &BulkResultResponse{Success: true, Count: count}, nil
}

func (controller *Controller) clearAll(ctx context.Context, userID int64) (*BulkResultResponse, error) {
	count, err := controller.service.ClearAll(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &BulkResultResponse{Success: true, Count: count}, nil
}

func (controller *Controller) getPreferences(ctx context.Context, userID int64, email string) (*NotificationPreferencesResponse, error) {
	return controller.service.GetPreferences(ctx, userID, email)
}

func (controller *Controller) updatePreferences(ctx context.Context, userID int64, email string, req *UpdatePreferencesRequest) (*NotificationPreferencesResponse, error) {
	return controller.service.UpdatePreferences(ctx, userID, email, req)
}

func (controller *Controller) upsertWorkspaceRule(ctx context.Context, userID, workspaceID int64, email string, req *UpsertWorkspaceRuleRequest) (*NotificationPreferencesResponse, error) {
	if workspaceID <= 0 {
		return nil, errors.Invalid("invalid workspace id")
	}
	if req.ProjectMode != "all" && req.ProjectMode != "selected" {
		return nil, errors.Invalid("projectMode must be 'all' or 'selected'")
	}
	return controller.service.UpsertWorkspaceRule(ctx, userID, workspaceID, email, req)
}

func (controller *Controller) deleteWorkspaceRule(ctx context.Context, userID, workspaceID int64, email string) (*NotificationPreferencesResponse, error) {
	if workspaceID <= 0 {
		return nil, errors.Invalid("invalid workspace id")
	}
	return controller.service.DeleteWorkspaceRule(ctx, userID, workspaceID, email)
}

func toNotificationResponse(n *schemas.Notification) NotificationResponse {
	var eventData any
	if n.EventData != "" {
		json.Unmarshal([]byte(n.EventData), &eventData)
	}
	return NotificationResponse{
		ID:           n.ID,
		UserID:       n.UserID,
		Title:        n.Title,
		Content:      n.Content,
		Type:         n.Type,
		EventData:    eventData,
		IsRead:       n.IsRead,
		ResourceID:   n.ResourceID,
		ResourceType: n.ResourceType,
		CreatedAt:    n.CreatedAt,
	}
}
