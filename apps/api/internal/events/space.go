package events

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/FacileStudio/Opus/apps/api/schemas"
)

// SpaceOf resolves where a workspace's projects live for Sablier. It reports
// false for a missing workspace and for one that is neither linked to a
// Registre space nor someone's personal workspace: those never leave Opus.
func (p *Publisher) SpaceOf(ctx context.Context, workspaceID int64) (SpacePayload, bool) {
	var workspaces []schemas.Workspace
	err := p.db.WithContext(ctx).Select("id", "registre_space_id", "personal_owner_id").
		Where("id = ?", workspaceID).Limit(1).Find(&workspaces).Error
	if err != nil || len(workspaces) == 0 {
		return SpacePayload{}, false
	}
	workspace := workspaces[0]
	if workspace.RegistreSpaceID != nil && *workspace.RegistreSpaceID != "" {
		return SpacePayload{RegistreSpaceID: *workspace.RegistreSpaceID}, true
	}
	if workspace.PersonalOwnerID == nil {
		return SpacePayload{}, false
	}
	var emails []string
	p.db.WithContext(ctx).Model(&schemas.User{}).Where("id = ?", *workspace.PersonalOwnerID).Limit(1).Pluck("email", &emails)
	space := SpacePayload{Personal: true, OwnerSubject: p.subjectOf(ctx, *workspace.PersonalOwnerID)}
	if len(emails) > 0 {
		space.OwnerEmail = strings.ToLower(emails[0])
	}
	return space, true
}

func (p *Publisher) subjectOf(ctx context.Context, userID int64) string {
	if p.config.Issuer == "" {
		return ""
	}
	var subjects []string
	p.db.WithContext(ctx).Table("porte_identities").Where("user_id = ? AND provider = ?", userID, p.config.Issuer).
		Limit(1).Pluck("subject", &subjects)
	if len(subjects) == 0 {
		return ""
	}
	return subjects[0]
}

func (p *Publisher) ensureFacileID(ctx context.Context, model any, id int64, current string) string {
	if current != "" || id == 0 {
		return current
	}
	minted := schemas.NewFacileID()
	err := p.db.WithContext(ctx).Model(model).Where("id = ? AND (facile_id IS NULL OR facile_id = '')", id).
		Update("facile_id", minted).Error
	if err != nil {
		p.logger.Warn("antenne: failed to assign a facile_id", slog.Int64("id", id), slog.Any("error", err))
		return ""
	}
	var stored []string
	p.db.WithContext(ctx).Model(model).Where("id = ?", id).Limit(1).Pluck("facile_id", &stored)
	if len(stored) == 0 {
		return ""
	}
	return stored[0]
}

// link stamps synced_at without touching updated_at, so last-writer-wins
// still compares real edits.
func (p *Publisher) link(ctx context.Context, model any, id int64) {
	if p.db == nil || id == 0 {
		return
	}
	p.db.WithContext(context.WithoutCancel(ctx)).Model(model).Where("id = ?", id).UpdateColumn("synced_at", time.Now())
}

func (p *Publisher) prepareProject(ctx context.Context, payload *ProjectPayload) bool {
	if p.db == nil {
		return true
	}
	space, ok := p.SpaceOf(ctx, payload.WorkspaceID)
	if !ok {
		return false
	}
	payload.Space = &space
	payload.FacileID = p.ensureFacileID(ctx, &schemas.Project{}, payload.ID, payload.FacileID)
	return payload.FacileID != ""
}

func (p *Publisher) prepareTask(ctx context.Context, payload *TaskPayload) bool {
	if p.db == nil {
		return true
	}
	if _, ok := p.SpaceOf(ctx, payload.WorkspaceID); !ok {
		return false
	}
	payload.FacileID = p.ensureFacileID(ctx, &schemas.Task{}, payload.ID, payload.FacileID)
	payload.ProjectFacileID = p.ensureFacileID(ctx, &schemas.Project{}, payload.ProjectID, payload.ProjectFacileID)
	if payload.ActorID != nil && payload.ActorSubject == "" {
		payload.ActorSubject = p.subjectOf(ctx, *payload.ActorID)
	}
	return payload.FacileID != "" && payload.ProjectFacileID != ""
}
