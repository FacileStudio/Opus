package notification

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"time"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
	"gorm.io/gorm"
)

// Service provides notification persistence and lifecycle management.
type Service struct {
	db         *gorm.DB
	controller *Controller
}

// NewService instantiates a new notification service.
func NewService(db *gorm.DB) *Service {
	service := &Service{db: db}
	service.controller = newController(service)
	return service
}

// List returns notifications for a user ordered by creation time descending.
func (service *Service) List(ctx context.Context, userID int64) ([]schemas.Notification, error) {
	var notifications []schemas.Notification
	err := service.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at desc").
		Find(&notifications).Error
	if err != nil {
		return nil, errors.Internal("failed to list notifications", err)
	}
	return notifications, nil
}

// Create inserts a new notification record for a user.
func (service *Service) Create(ctx context.Context, userID int64, req *CreateNotificationRequest) (*schemas.Notification, error) {
	title := ""
	if req.Title != nil {
		title = *req.Title
	}
	content := ""
	if req.Content != nil {
		content = *req.Content
	} else if req.Message != nil {
		content = *req.Message
	}
	resID := req.ResourceID
	if resID == nil {
		resID = req.RelatedEntityID
	}
	resType := req.ResourceType
	if resType == nil {
		resType = req.RelatedEntityType
	}
	eventDataStr := ""
	if req.EventData != nil {
		b, _ := json.Marshal(req.EventData)
		eventDataStr = string(b)
	}
	n := schemas.Notification{
		UserID:       userID,
		Title:        title,
		Content:      content,
		Type:         req.Type,
		EventData:    eventDataStr,
		IsRead:       false,
		ResourceID:   resID,
		ResourceType: resType,
		CreatedAt:    time.Now(),
	}
	if err := service.db.WithContext(ctx).Create(&n).Error; err != nil {
		return nil, errors.Internal("failed to create notification", err)
	}
	return &n, nil
}

// MarkAsRead flags a notification as read if owned by the caller.
func (service *Service) MarkAsRead(ctx context.Context, userID, notificationID int64) (*schemas.Notification, error) {
	return service.setRead(ctx, userID, notificationID, true)
}

// MarkAsUnread flags a notification as unread if owned by the caller.
func (service *Service) MarkAsUnread(ctx context.Context, userID, notificationID int64) (*schemas.Notification, error) {
	return service.setRead(ctx, userID, notificationID, false)
}

// Delete removes a single notification owned by the caller.
func (service *Service) Delete(ctx context.Context, userID, notificationID int64) error {
	result := service.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Delete(&schemas.Notification{})
	if result.Error != nil {
		return errors.Internal("failed to delete notification", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.NotFound("Notification not found")
	}
	return nil
}

func (service *Service) setRead(ctx context.Context, userID, notificationID int64, read bool) (*schemas.Notification, error) {
	var n schemas.Notification
	err := service.db.WithContext(ctx).First(&n, "id = ? AND user_id = ?", notificationID, userID).Error
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("Notification not found")
		}
		return nil, errors.Internal("failed to query notification", err)
	}
	n.IsRead = read
	if err := service.db.WithContext(ctx).Save(&n).Error; err != nil {
		return nil, errors.Internal("failed to update notification", err)
	}
	return &n, nil
}

// MarkAllAsRead marks all unread notifications as read for the user.
func (service *Service) MarkAllAsRead(ctx context.Context, userID int64) (int64, error) {
	result := service.db.WithContext(ctx).
		Model(&schemas.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true)
	if result.Error != nil {
		return 0, errors.Internal("failed to mark notifications as read", result.Error)
	}
	return result.RowsAffected, nil
}

// ClearAll deletes all notifications belonging to the user.
func (service *Service) ClearAll(ctx context.Context, userID int64) (int64, error) {
	result := service.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&schemas.Notification{})
	if result.Error != nil {
		return 0, errors.Internal("failed to clear notifications", result.Error)
	}
	return result.RowsAffected, nil
}
