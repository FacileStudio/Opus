package events

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func outboxPublisher(t *testing.T) (*Publisher, *fakeTransport, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(&schemas.User{}, &schemas.Workspace{}, &schemas.Project{}, &schemas.AntenneOutbox{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	fake := &fakeTransport{connected: true}
	return New(Config{URL: "http://localhost", Secret: "sec"}, WithTransport(fake), WithStore(db)), fake, db
}

func TestOutboxRetriesInOrder(t *testing.T) {
	pub, fake, db := outboxPublisher(t)
	owner := schemas.User{Email: "Owner@Facile.studio"}
	db.Create(&owner)
	personal := schemas.Workspace{Name: "Personal", Slug: "personal", CreatedAt: time.Now(), PersonalOwnerID: &owner.ID}
	shared := schemas.Workspace{Name: "Local", Slug: "local", CreatedAt: time.Now()}
	db.Create(&personal)
	db.Create(&shared)
	mine := schemas.Project{WorkspaceID: personal.ID, Name: "Mine", Slug: "mine"}
	hidden := schemas.Project{WorkspaceID: shared.ID, Name: "Hidden", Slug: "hidden"}
	db.Create(&mine)
	db.Create(&hidden)

	ctx := context.Background()
	pub.EmitProjectCreated(ctx, NewProjectPayload(mine, ""))
	pub.EmitProjectCreated(ctx, NewProjectPayload(hidden, ""))
	mine.Name = "Mine v2"
	pub.EmitProjectUpdated(ctx, NewProjectPayload(mine, ""))

	fake.emitErr = errors.New("pool down")
	pub.drainOutbox(ctx)
	var rows []schemas.AntenneOutbox
	db.Order("id").Find(&rows)
	if len(rows) != 2 || rows[0].Attempts != 1 || rows[1].Attempts != 0 || rows[0].Channel != ChannelProjectCreated {
		t.Fatalf("failed drain must keep both rows and stop at the first: %+v", rows)
	}

	fake.emitErr = nil
	pub.drainOutbox(ctx)
	msgs := fake.getMessages()
	if len(msgs) != 2 || msgs[0].channel != ChannelProjectCreated || msgs[1].channel != ChannelProjectUpdated {
		t.Fatalf("messages out of order: %+v", msgs)
	}
	var sent struct {
		IdempotencyKey string         `json:"idempotency_key"`
		Payload        ProjectPayload `json:"payload"`
	}
	if err := json.Unmarshal(msgs[0].payload.(json.RawMessage), &sent); err != nil {
		t.Fatalf("decode: %v", err)
	}
	space := sent.Payload.Space
	if space == nil || !space.Personal || space.OwnerEmail != "owner@facile.studio" || space.RegistreSpaceID != "" {
		t.Errorf("personal space not described: %+v", space)
	}
	if want := "project.created:" + *mine.FacileID + ":"; len(sent.IdempotencyKey) <= len(want) || sent.IdempotencyKey[:len(want)] != want {
		t.Errorf("idempotency key = %q, want prefix %q", sent.IdempotencyKey, want)
	}
	if n := db.Find(&[]schemas.AntenneOutbox{}).RowsAffected; n != 0 {
		t.Errorf("outbox not emptied: %d rows", n)
	}
}
