package workspace

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/FacileStudio/porte/directory"
	portepg "github.com/FacileStudio/porte/pg"

	"gorm.io/gorm"
)

// erasedRegistre lists one erasure of subject, and serves it on every path.
func erasedRegistre(t *testing.T, subject string) string {
	t.Helper()
	at := time.Now().Add(-time.Minute).UTC()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if err := json.NewEncoder(w).Encode(map[string]any{"deleted": []directory.Deleted{{Subject: subject, DeletedAt: at}}}); err != nil {
			t.Errorf("deletion list: %v", err)
		}
	}))
	t.Cleanup(server.Close)
	return server.URL
}

func newReceiver(t *testing.T, service *Service, db *gorm.DB, issuer string) *Deletions {
	t.Helper()
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("sql handle: %v", err)
	}
	return NewDeletions(db, directory.Client{Issuer: issuer, ClientID: "opus", ClientSecret: "s"},
		eraseAccount(t, service, db, &fakeAssets{}), portepg.New(sqlDB), slog.New(slog.DiscardHandler))
}

func seedSubject(t *testing.T, db *gorm.DB, email, issuer, subject string) int64 {
	t.Helper()
	id := createUser(t, db, email, "Alice")
	if err := db.Exec("INSERT INTO porte_identities (user_id, provider, subject) VALUES (?, ?, ?)", id, issuer, subject).Error; err != nil {
		t.Fatalf("add identity: %v", err)
	}
	return id
}

// A notice handled before an account was re-created for its subject erases
// that account on the next delivery to the same running receiver, instead of
// skipping the subject as done.
func TestARepushedNoticeErasesAnAccountRecreatedAfterTheErasure(t *testing.T) {
	service, db := newPostgresService(t)
	ctx := context.Background()
	issuer := erasedRegistre(t, "sub-alice")
	seedSubject(t, db, "alice@facile.studio", issuer, "sub-alice")
	bob := seedSubject(t, db, "bob@facile.studio", issuer, "sub-bob")
	receiver := newReceiver(t, service, db, issuer)
	if err := receiver.receiver.Poll(ctx); err != nil {
		t.Fatalf("first notice: %v", err)
	}
	back := seedSubject(t, db, "alice@facile.studio", issuer, "sub-alice")
	if err := receiver.receiver.Poll(ctx); err != nil {
		t.Fatalf("repushed notice: %v", err)
	}
	if n := count(t, db, "SELECT count(*) FROM users WHERE id = ?", back); n != 0 {
		t.Errorf("the re-created account survived the repushed notice")
	}
	if n := count(t, db, "SELECT count(*) FROM users WHERE id = ?", bob); n != 1 {
		t.Errorf("the repushed notice touched another account")
	}
}

// A Registre snapshot that still lists a person the receiver erased neither
// provisions nor links them, and still provisions everyone else.
func TestTheSpaceSyncSkipsAnErasedMember(t *testing.T) {
	service, db := newPostgresService(t)
	ctx := context.Background()
	issuer := erasedRegistre(t, "sub-gone")
	receiver := newReceiver(t, service, db, issuer)
	if err := receiver.receiver.Poll(ctx); err != nil {
		t.Fatalf("notice: %v", err)
	}
	dir := NewDirectory(service, directory.Client{Issuer: issuer}, slog.New(slog.DiscardHandler))
	dir.RefuseErased(receiver)
	createUser(t, db, "linked@facile.studio", "Linked")
	members := []directory.Member{
		{Subject: "sub-gone", Email: "gone@facile.studio", Role: directory.RoleMember},
		{Subject: "sub-gone", Email: "linked@facile.studio", Role: directory.RoleMember},
		{Subject: "sub-kept", Email: "kept@facile.studio", Role: directory.RoleMember},
	}
	snapshot := directory.Snapshot{Version: 1, Spaces: []directory.Space{{ID: "space-studio", Name: "Studio", Members: members}}}
	if err := dir.Apply(ctx, snapshot); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if n := count(t, db, "SELECT count(*) FROM porte_identities WHERE subject = 'sub-gone'"); n != 0 {
		t.Errorf("the erased member got %d identities", n)
	}
	if n := count(t, db, "SELECT count(*) FROM users WHERE email = 'gone@facile.studio'"); n != 0 {
		t.Errorf("the erased member was provisioned")
	}
	if n := count(t, db, "SELECT count(*) FROM porte_identities WHERE subject = 'sub-kept'"); n != 1 {
		t.Errorf("the other member was not provisioned")
	}
}
