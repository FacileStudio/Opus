package events

import (
	"context"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"gorm.io/gorm"
)

// TaskActorEmails maps each task id to the email of its assignee. Tasks with
// no assignee, or an assignee without an email, are absent from the result.
func TaskActorEmails(ctx context.Context, db *gorm.DB, tasks []schemas.Task) map[int64]string {
	userIDs := make([]int64, 0, len(tasks))
	for _, task := range tasks {
		if task.ActorID != nil {
			userIDs = append(userIDs, *task.ActorID)
		}
	}
	emails := make(map[int64]string, len(tasks))
	if len(userIDs) == 0 {
		return emails
	}
	var users []schemas.User
	db.WithContext(ctx).Select("id", "email").Where("id IN ?", userIDs).Find(&users)
	byUser := make(map[int64]string, len(users))
	for _, user := range users {
		byUser[user.ID] = user.Email
	}
	for _, task := range tasks {
		if task.ActorID != nil {
			emails[task.ID] = byUser[*task.ActorID]
		}
	}
	return emails
}
