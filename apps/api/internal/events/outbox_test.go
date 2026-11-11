package events

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/FacileStudio/Opus/apps/api/internal/database/dbtest"
	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/antenne-client/go/outbox"
	"gorm.io/gorm"
)

func outboxPublisher(t *testing.T) (*Publisher, *fakeTransport, *gorm.DB) {
	t.Helper()
	db := dbtest.Open(t)
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
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
	if _, err := pub.drainer.Drain(ctx); err == nil {
		t.Fatal("a refused emit must fail the pass")
	}
	var unsent []string
	db.Raw("SELECT channel FROM antenne_outbox WHERE sent_at IS NULL ORDER BY id").Scan(&unsent)
	if len(unsent) != 2 || unsent[0] != ChannelProjectCreated || len(fake.getMessages()) != 0 {
		t.Fatalf("failed drain must keep both rows and stop at the first: %v", unsent)
	}

	fake.emitErr = nil
	if _, err := pub.drainer.Drain(ctx); err != nil {
		t.Fatalf("drain: %v", err)
	}
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
	var left int64
	db.Raw("SELECT count(*) FROM antenne_outbox WHERE sent_at IS NULL").Scan(&left)
	if left != 0 {
		t.Errorf("outbox not drained: %d rows unsent", left)
	}
}

func outboxShape(t *testing.T, db *gorm.DB, schema string) []string {
	t.Helper()
	var shape []string
	err := db.Raw(`SELECT column_name || ' ' || data_type || ' ' || is_nullable || ' ' || replace(coalesce(column_default, ''), ? || '.', '')
		FROM information_schema.columns WHERE table_schema = ? AND table_name = 'antenne_outbox'
		UNION ALL SELECT replace(indexdef, ? || '.', '') FROM pg_indexes WHERE schemaname = ? AND tablename = 'antenne_outbox'
		ORDER BY 1`, schema, schema, schema, schema).Scan(&shape).Error
	if err != nil || len(shape) == 0 {
		t.Fatalf("read the outbox shape of %s: %v, %v", schema, shape, err)
	}
	return shape
}

func TestConvertedOutboxMatchesTheLibraryMigration(t *testing.T) {
	db := dbtest.Open(t)
	var current string
	db.Raw("SELECT current_schema()").Scan(&current)
	up, _, _ := strings.Cut(strings.SplitN(outbox.Migration, "-- +goose Up", 2)[1], "-- +goose Down")
	const reference = "opus_outbox_reference"
	statements := []string{"DROP SCHEMA IF EXISTS " + reference + " CASCADE", "CREATE SCHEMA " + reference,
		"SET LOCAL search_path TO " + reference, up}
	err := db.Transaction(func(tx *gorm.DB) error {
		for _, statement := range statements {
			if err := tx.Exec(statement).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("apply the library migration: %v", err)
	}
	t.Cleanup(func() { db.Exec("DROP SCHEMA IF EXISTS " + reference + " CASCADE") })
	got, want := outboxShape(t, db, current), outboxShape(t, db, reference)
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("antenne_outbox differs from outbox.Migration:\n got %q\nwant %q", got, want)
	}
}

func TestUnconfiguredPublisherQueuesNothing(t *testing.T) {
	db := dbtest.Open(t)
	pub := New(Config{}, WithStore(db))
	if err := pub.EmitWorkspaceCreated(context.Background(), WorkspacePayload{ID: 1, Slug: "studio"}); err != nil {
		t.Fatalf("emit: %v", err)
	}
	var queued int64
	db.Raw("SELECT count(*) FROM antenne_outbox").Scan(&queued)
	if queued != 0 || pub.drainer != nil {
		t.Fatalf("an unconfigured publisher queued %d rows", queued)
	}
}

func TestDrainedEventsKeepThePinnedBytes(t *testing.T) {
	pub, fake, _ := outboxPublisher(t)
	samples := wireSamples()
	names := make([]string, 0, len(samples))
	for name, payload := range samples {
		names = append(names, name)
		channel, _, _ := strings.Cut(name, ".minimal")
		pub.enqueue(context.Background(), channel, "pin:"+name, json.RawMessage(wireEnvelope(t, name, payload)))
	}
	if _, err := pub.drainer.Drain(context.Background()); err != nil {
		t.Fatalf("drain: %v", err)
	}
	msgs := fake.getMessages()
	if len(msgs) != len(names) {
		t.Fatalf("sent %d messages, want %d", len(msgs), len(names))
	}
	for i, name := range names {
		want, err := os.ReadFile(filepath.Join("testdata", "events", name+".json"))
		if err != nil {
			t.Fatal(err)
		}
		if got := string(msgs[i].payload.(json.RawMessage)) + "\n"; got != string(want) {
			t.Errorf("%s left the outbox as\n%s\nwant\n%s", name, got, want)
		}
	}
}
