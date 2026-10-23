package comment

import (
	"context"
	"strconv"
	"strings"

	"github.com/FacileStudio/Opus/apps/api/internal/realtime"
	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
)

// Controller validates requests and maps comment models to API responses.
type Controller struct {
	service *Service
}

func newController(service *Service) *Controller {
	return &Controller{service: service}
}

func (controller *Controller) list(ctx context.Context, userID, taskID int64) ([]CommentResponse, error) {
	if taskID <= 0 {
		return nil, errors.Invalid("invalid task id")
	}
	records, err := controller.service.ListForTask(ctx, userID, taskID)
	if err != nil {
		return nil, err
	}
	responses := make([]CommentResponse, 0, len(records))
	for _, record := range records {
		responses = append(responses, toRecordResponse(record))
	}
	return responses, nil
}

func (controller *Controller) create(ctx context.Context, userID, taskID int64, req *CreateCommentRequest) (*CommentResponse, error) {
	if taskID <= 0 {
		return nil, errors.Invalid("invalid task id")
	}
	content := strings.TrimSpace(req.Content)
	if content == "" {
		return nil, errors.Invalid("content is required")
	}
	comment, err := controller.service.Create(ctx, userID, taskID, content)
	if err != nil {
		return nil, err
	}
	controller.service.notify(ctx, realtime.BroadcastMessage{Type: realtime.TypeCommentUpdated, TaskID: strconv.FormatInt(comment.TaskID, 10)}, comment.TaskID)
	return toCommentResponse(comment), nil
}

func (controller *Controller) update(ctx context.Context, userID, commentID int64, req *UpdateCommentRequest) (*CommentResponse, error) {
	if commentID <= 0 {
		return nil, errors.Invalid("invalid comment id")
	}
	content := strings.TrimSpace(req.Content)
	if content == "" {
		return nil, errors.Invalid("content is required")
	}
	comment, err := controller.service.Update(ctx, userID, commentID, content)
	if err != nil {
		return nil, err
	}
	controller.service.notify(ctx, realtime.BroadcastMessage{Type: realtime.TypeCommentUpdated, TaskID: strconv.FormatInt(comment.TaskID, 10)}, comment.TaskID)
	return toCommentResponse(comment), nil
}

func (controller *Controller) delete(ctx context.Context, userID, commentID int64) (*CommentResponse, error) {
	if commentID <= 0 {
		return nil, errors.Invalid("invalid comment id")
	}
	comment, err := controller.service.Delete(ctx, userID, commentID)
	if err != nil {
		return nil, err
	}
	controller.service.notify(ctx, realtime.BroadcastMessage{Type: realtime.TypeCommentUpdated, TaskID: strconv.FormatInt(comment.TaskID, 10)}, comment.TaskID)
	return toCommentResponse(comment), nil
}

func toRecordResponse(record CommentRecord) CommentResponse {
	return CommentResponse{
		ID:        record.Comment.ID,
		TaskID:    record.Comment.TaskID,
		UserID:    record.Comment.UserID,
		Content:   record.Comment.Content,
		CreatedAt: record.Comment.CreatedAt,
		UpdatedAt: record.Comment.UpdatedAt,
		User: &CommentUserResponse{
			Name:  record.AuthorName,
			Image: record.AuthorImage,
		},
	}
}

func toCommentResponse(comment *schemas.Comment) *CommentResponse {
	return &CommentResponse{
		ID:        comment.ID,
		TaskID:    comment.TaskID,
		UserID:    comment.UserID,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt,
		UpdatedAt: comment.UpdatedAt,
	}
}
