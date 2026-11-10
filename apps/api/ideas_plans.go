package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/porte/directory"
	"gorm.io/gorm/clause"
)

// planStateID is the forfait_state row Opus keeps its plans in.
const planStateID = 1

// planSaveTimeout bounds one save of the plan state.
const planSaveTimeout = 10 * time.Second

// loadPlans restores the plans saved before the last restart. Until it or a
// snapshot brings an authoritative revision, ai.access reads as off.
func (i *ideas) loadPlans(ctx context.Context) {
	var rows []schemas.ForfaitState
	if err := i.db.WithContext(ctx).Where("id = ?", planStateID).Find(&rows).Error; err != nil {
		i.logger.Error("plans: saved entitlements not read, AI stays off until Registre answers", slog.Any("error", err))
		return
	}
	if len(rows) == 0 {
		return
	}
	document := []byte(rows[0].Document)
	if err := i.plans.Import(document); err != nil {
		i.logger.Error("plans: saved entitlements not restored, AI stays off until Registre answers", slog.Any("error", err))
		return
	}
	i.markSynced(document)
}

// savePlans applies a directory snapshot and saves the resulting state.
func (i *ideas) savePlans(snapshot directory.Snapshot) {
	i.plans.Apply(snapshot)
	document, err := i.plans.Export()
	if err != nil {
		i.logger.Error("plans: entitlements not exported", slog.Any("error", err))
		return
	}
	i.markSynced(document)
	ctx, cancel := context.WithTimeout(context.Background(), planSaveTimeout)
	defer cancel()
	row := schemas.ForfaitState{ID: planStateID, Document: string(document)}
	err = i.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"document", "updated_at"}),
	}).Create(&row).Error
	if err != nil {
		i.logger.Error("plans: entitlements not saved", slog.Any("error", err))
	}
}

// markSynced records whether document holds an authoritative revision.
func (i *ideas) markSynced(document []byte) {
	var state struct {
		Revision int64 `json:"entitlements_revision"`
	}
	if json.Unmarshal(document, &state) == nil && state.Revision > 0 {
		i.synced.Store(true)
	}
}
