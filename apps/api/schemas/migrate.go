package schemas

import "gorm.io/gorm"

// Migrate brings the Go backend's own schema up to date. AutoMigrate only
// adds what is missing and never drops a column, which is the right shape for
// a dev-cycle schema the Drizzle migration runner no longer owns. Goose takes
// over at cutover (Phase 6), from a pg_dump of the live database.
func Migrate(db *gorm.DB) error {
	models := []any{
		&User{},
		&Workspace{},
		&WorkspaceMember{},
		&Invitation{},
		&Project{},
		&Column{},
		&Task{},
		&Label{},
		&Comment{},
		&TaskRelation{},
		&Activity{},
		&ShareLink{},
		&TimeEntry{},
		&WorkflowRule{},
		&Notification{},
		&UserNotificationPreference{},
		&UserNotificationWorkspaceRule{},
		&TaskReminderSent{},
		&Asset{},
	}
	if err := db.AutoMigrate(models...); err != nil {
		return err
	}
	return AdoptPorte(db)
}
