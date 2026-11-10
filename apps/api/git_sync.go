package main

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/FacileStudio/Opus/apps/api/internal/env"
	"github.com/FacileStudio/Opus/apps/api/internal/events"
	"github.com/FacileStudio/Opus/apps/api/modules/github"
	antenne "github.com/FacileStudio/antenne-client/go"
	"gorm.io/gorm"
)

func attachGitSync(db *gorm.DB, appEnv *env.Config, pub *events.Publisher, appLogger *slog.Logger) {
	if pub == nil {
		return
	}
	svc := github.NewService(db, appEnv.GitHubWebhookSecret)
	handler := func(payload json.RawMessage, meta antenne.EventMeta) {
		ghPayload, err := github.DecodePayload(payload)
		if err != nil {
			appLogger.Warn("antenne git event decode failed", slog.Any("error", err))
			return
		}
		if err := svc.ProcessWebhook(context.Background(), ghPayload); err != nil {
			appLogger.Error("antenne git event process failed", slog.Any("error", err))
		}
	}
	pub.Listen(events.ChannelGitPullRequest, handler)
	pub.Listen(events.ChannelGitPush, handler)
}
