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
