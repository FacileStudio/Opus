package workspace

import (
	"context"
	"fmt"
	"testing"

	"github.com/FacileStudio/Opus/apps/api/internal/events"

	"github.com/FacileStudio/antenne-client/go/catalog"
	envelope "github.com/FacileStudio/antenne-client/go/envelope"
	"github.com/FacileStudio/antenne-client/go/outbox"

	"gorm.io/gorm"
)

func seedErasureTraces(t *testing.T, db *gorm.DB, f erasureFixture) {
	t.Helper()
	statements := []struct {
		query string
		args  []any
	}{
		{"UPDATE users SET github_id = 'gh-alice', image = 'https://sso.test/avatars/sub-alice' WHERE id = ?", []any{f.alice}},
		{"UPDATE tasks SET actor_id = ? WHERE id = ?", []any{f.alice, f.sharedTask}},
		{"INSERT INTO activities (task_id, user_id, type, content) VALUES (?, ?, 'comment', 'shared note')", []any{f.sharedTask, f.alice}},
		{"INSERT INTO time_entries (task_id, user_id, description, start_time) VALUES (?, ?, 'pairing', now())", []any{f.sharedTask, f.alice}},
		{"INSERT INTO notifications (user_id, title, content) VALUES (?, 'Reminder', 'due soon')", []any{f.alice}},
		{`INSERT INTO mcp_actions (id, user_id, kind, target, digest, summary, expires_at)
			VALUES ('act-1', ?, 'delete_task', 'task:1', 'd', 'Delete Ship it', now() + interval '1 hour')`, []any{fmt.Sprint(f.alice)}},
		{`INSERT INTO task_pull_requests (task_id, repo_full_name, pr_number, title, url, state, author_github_id)
			VALUES (?, 'facile/opus', 7, 'Fix it', 'https://github.com/facile/opus/pull/7', 'open', 'gh-alice')`, []any{f.sharedTask}},
		{`INSERT INTO antenne_outbox (channel, payload, idempotency_key, sent_at) VALUES
			('task.updated', '{"payload":{"actor_email":"alice@facile.studio"}}', 'k1', now()),
			('project.created', '{"payload":{"space":{"owner_subject":"sub-alice"}}}', 'k2', now()),
			('task.updated', '{"payload":{"actor_email":"bob@facile.studio"}}', 'k3', now())`, nil},
	}
	for _, s := range statements {
		if err := db.Exec(s.query, s.args...).Error; err != nil {
			t.Fatalf("seed %q: %v", s.query, err)
		}
	}
	queueUnsentEvents(t, db)
}

func queueUnsentEvents(t *testing.T, db *gorm.DB) {
	t.Helper()
	queued := []struct {
		channel, key string
		event        any
	}{
		{"task.updated", "u1", envelope.Event[events.TaskPayload]{Payload: events.TaskPayload{
			Title: "Ship it", ActorEmail: "Alice@Facile.Studio", OwnerEmail: "alice@facile.studio"}}},
		{"task.created", "u2", envelope.Event[events.TaskPayload]{Payload: events.TaskPayload{
			Title: "Ship it", ActorSubject: "sub-alice"}}},
		{"project.created", "u3", envelope.Event[events.ProjectPayload]{Payload: events.ProjectPayload{
			Name: "Side", Space: &events.SpacePayload{Personal: true, OwnerSubject: "sub-alice", OwnerEmail: "alice@facile.studio"}}}},
		{"task.updated", "u4", envelope.Event[events.TaskPayload]{Payload: events.TaskPayload{
			Title: "Review", ActorEmail: "bob@facile.studio", ActorSubject: "sub-bob"}}},
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("sql db: %v", err)
	}
	tx, err := sqlDB.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	for _, q := range queued {
		if err := outbox.Enqueue(context.Background(), tx, catalog.Channel(q.channel), q.event, q.key); err != nil {
			t.Fatalf("enqueue %s: %v", q.key, err)
		}
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}
	if n := count(t, db, "SELECT count(*) FROM antenne_outbox WHERE sent_at IS NULL"); n != 4 {
		t.Fatalf("queued %d unsent events, want 4", n)
	}
}

func textColumns(t *testing.T, db *gorm.DB) [][2]string {
	t.Helper()
	var rows []struct{ TableName, ColumnName string }
	err := db.Raw(`SELECT table_name, column_name FROM information_schema.columns
		WHERE table_schema = current_schema() AND data_type IN ('text', 'character varying', 'jsonb', 'json')`).Scan(&rows).Error
	if err != nil || len(rows) == 0 {
		t.Fatalf("list text columns: %d, %v", len(rows), err)
	}
	out := make([][2]string, len(rows))
	for i, r := range rows {
		out[i] = [2]string{r.TableName, r.ColumnName}
	}
	return out
}

func TestErasureLeavesNoTextNamingThePerson(t *testing.T) {
	service, db := newPostgresService(t)
	f := seedErasure(t, service, db)
	seedErasureTraces(t, db, f)
	erase := eraseAccount(t, service, db, &fakeAssets{})
	if err := erase(context.Background(), testIssuer+"/", "sub-alice"); err != nil {
		t.Fatalf("EraseAccount: %v", err)
	}

	needles := []string{"alice@facile.studio", "Alice", "sub-alice", "gh-alice", fmt.Sprint(f.alice)}
	for _, column := range textColumns(t, db) {
		for _, needle := range needles[:4] {
			query := fmt.Sprintf(`SELECT count(*) FROM %q WHERE %q::text ILIKE '%%' || ? || '%%'`, column[0], column[1])
			if needle == "Alice" {
				query = fmt.Sprintf(`SELECT count(*) FROM %q WHERE %q::text LIKE '%%' || ? || '%%'`, column[0], column[1])
			}
			if n := count(t, db, query, needle); n != 0 {
				t.Errorf("%s.%s: %d rows mention %q", column[0], column[1], n, needle)
			}
		}
	}
	if n := count(t, db, "SELECT count(*) FROM mcp_actions WHERE user_id = ?", needles[4]); n != 0 {
		t.Errorf("mcp_actions: %d rows for the erased id", n)
	}
	if n := count(t, db, "SELECT count(*) FROM antenne_outbox WHERE idempotency_key IN ('u1', 'u2', 'u3')"); n != 0 {
		t.Errorf("antenne_outbox: %d queued events naming the person survive", n)
	}
	references := map[string]int64{
		"activity author":   count(t, db, "SELECT count(*) FROM activities WHERE user_id IS NULL AND content = 'shared note'"),
		"time entry author": count(t, db, "SELECT count(*) FROM time_entries WHERE user_id IS NULL AND description = 'pairing'"),
		"task assignee":     count(t, db, "SELECT count(*) FROM tasks WHERE id = ? AND actor_id IS NULL", f.sharedTask),
		"pull request":      count(t, db, "SELECT count(*) FROM task_pull_requests WHERE title = 'Fix it' AND author_github_id IS NULL"),
		"others' events":    count(t, db, "SELECT count(*) FROM antenne_outbox WHERE idempotency_key = 'k3'"),
		"others' queued":    count(t, db, "SELECT count(*) FROM antenne_outbox WHERE idempotency_key = 'u4' AND sent_at IS NULL"),
	}
	for name, n := range references {
		if n != 1 {
			t.Errorf("%s: %d anonymised rows kept, want 1", name, n)
		}
	}
	checkErasureSurvivors(t, db, f)
}
