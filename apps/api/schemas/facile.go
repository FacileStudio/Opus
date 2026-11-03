package schemas

import (
	"crypto/rand"
	"time"

	"gorm.io/gorm"
)

// FacileRef carries the stable cross-app identifier shared with Sablier.
type FacileRef struct {
	FacileID *string `gorm:"column:facile_id;uniqueIndex"`
}

// NewFacileID mints a facile_id.
func NewFacileID() string {
	return rand.Text()
}

// BeforeCreate mints the facile_id when the caller did not supply one.
func (r *FacileRef) BeforeCreate(*gorm.DB) error {
	if r.FacileID == nil || *r.FacileID == "" {
		minted := NewFacileID()
		r.FacileID = &minted
	}
	return nil
}

// PoolProcessedEvent is the idempotency ledger of inbound pool events.
type PoolProcessedEvent struct {
	IdempotencyKey string    `gorm:"column:idempotency_key;primaryKey"`
	ProcessedAt    time.Time `gorm:"column:processed_at;not null;index;autoCreateTime"`
}

// TableName pins the ledger table name.
func (PoolProcessedEvent) TableName() string { return "pool_processed_events" }

// TimeEntrySourceSablier marks a time entry mirrored from Sablier, which Opus
// serves read-only.
const TimeEntrySourceSablier = "sablier"

// AntenneOutbox is one outbound pool event waiting for delivery. The worker
// sends rows in id order and deletes each once the pool accepted it.
type AntenneOutbox struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement"`
	Channel   string    `gorm:"column:channel;not null"`
	Payload   string    `gorm:"column:payload;not null"`
	Attempts  int       `gorm:"column:attempts;not null;default:0"`
	LastError string    `gorm:"column:last_error;not null;default:''"`
	CreatedAt time.Time `gorm:"column:created_at;not null;autoCreateTime"`
}

// TableName pins the outbox table name.
func (AntenneOutbox) TableName() string { return "antenne_outbox" }
