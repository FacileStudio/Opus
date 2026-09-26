package activity

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
)

// Controller validates incoming payloads and maps activity entities to DTOs.
type Controller struct {
	service *Service
}

func newController(service *Service) *Controller {
	return &Controller{service: service}
}

func (controller *Controller) list(ctx context.Context, userID, taskID int64) ([]ActivityResponse, error) {
	if taskID <= 0 {
		return nil, errors.Invalid("invalid task id")
	}
	records, err := controller.service.ListForTask(ctx, userID, taskID)
	if err != nil {
		return nil, err
	}
	responses := make([]ActivityResponse, 0, len(records))
	for i := range records {
		responses = append(responses, toActivityResponse(&records[i]))
	}
	return responses, nil
}

func (controller *Controller) create(ctx context.Context, authUserID int64, req *CreateActivityRequest) (*ActivityResponse, error) {
	if req.TaskID <= 0 {
		return nil, errors.Invalid("invalid task id")
	}
	activityType := strings.TrimSpace(req.Type)
	if activityType == "" {
		return nil, errors.Invalid("type is required")
	}
	params := resolveCreateParams(authUserID, activityType, req)
	activity, err := controller.service.Create(ctx, authUserID, params)
	if err != nil {
		return nil, err
	}
	resp := toActivityResponse(activity)
	return &resp, nil
}

func resolveCreateParams(authUserID int64, activityType string, req *CreateActivityRequest) ActivityParams {
	targetUserID := req.UserID
	if targetUserID == nil && authUserID > 0 {
		uid := authUserID
		targetUserID = &uid
	}
	content := ""
	if req.Message != nil {
		content = *req.Message
	} else if req.Content != nil {
		content = *req.Content
	}
	eventData := ""
	if len(req.EventData) > 0 && string(req.EventData) != "null" {
		eventData = string(req.EventData)
	}
	return ActivityParams{
		TaskID:    req.TaskID,
		UserID:    targetUserID,
		Type:      activityType,
		Content:   content,
		EventData: eventData,
	}
}

func (controller *Controller) createComment(ctx context.Context, userID int64, req *CommentActivityRequest) (*ActivityResponse, error) {
	if req.TaskID <= 0 {
		return nil, errors.Invalid("invalid task id")
	}
	content := strings.TrimSpace(req.Comment)
	if content == "" {
		content = strings.TrimSpace(req.Content)
	}
	if content == "" {
		return nil, errors.Invalid("comment is required")
	}
	activity, err := controller.service.CreateComment(ctx, userID, req.TaskID, content)
	if err != nil {
		return nil, err
	}
	resp := toActivityResponse(activity)
	return &resp, nil
}

func (controller *Controller) updateComment(ctx context.Context, userID int64, req *UpdateCommentActivityRequest) (*ActivityResponse, error) {
	if req.ActivityID <= 0 {
		return nil, errors.Invalid("invalid activity id")
	}
	content := strings.TrimSpace(req.Comment)
	if content == "" {
		content = strings.TrimSpace(req.Content)
	}
	if content == "" {
		return nil, errors.Invalid("comment is required")
	}
	activity, err := controller.service.UpdateComment(ctx, userID, req.ActivityID, content)
	if err != nil {
		return nil, err
	}
	resp := toActivityResponse(activity)
	return &resp, nil
}

func (controller *Controller) deleteComment(ctx context.Context, userID int64, req *DeleteCommentActivityRequest) (*ActivityResponse, error) {
	if req.ActivityID <= 0 {
		return nil, errors.Invalid("invalid activity id")
	}
	activity, err := controller.service.DeleteComment(ctx, userID, req.ActivityID)
	if err != nil {
		return nil, err
	}
	resp := toActivityResponse(activity)
	return &resp, nil
}

func cleanNewlines(s string) string {
	for strings.Contains(s, "\n\n") {
		s = strings.ReplaceAll(s, "\n\n", "\n")
	}
	return s
}

func toActivityResponse(activity *schemas.Activity) ActivityResponse {
	resp := ActivityResponse{
		ID:        activity.ID,
		TaskID:    activity.TaskID,
		UserID:    activity.UserID,
		Type:      activity.Type,
		EventData: json.RawMessage("null"),
		CreatedAt: activity.CreatedAt,
		UpdatedAt: activity.UpdatedAt,
	}
	if activity.Content != "" {
		cleaned := cleanNewlines(activity.Content)
		resp.Content = &cleaned
	}
	if activity.EventData != "" {
		resp.EventData = json.RawMessage(activity.EventData)
	}
	return resp
}
