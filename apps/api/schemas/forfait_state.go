package schemas

import "time"

// ForfaitState is the one row holding the plan entitlements last applied,
// as porte/forfait's Export document, so a restart keeps enforcing them.
type ForfaitState struct {
	ID        int       `gorm:"column:id;primaryKey;autoIncrement:false"`
	Document  string    `gorm:"column:document;type:text;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

// TableName pins the table name.
func (ForfaitState) TableName() string { return "forfait_state" }
