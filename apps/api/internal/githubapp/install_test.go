package githubapp_test

import (
	"context"
	"testing"
	"time"

	"github.com/FacileStudio/Opus/apps/api/internal/githubapp"
	"github.com/FacileStudio/Opus/apps/api/internal/githubapp/githubapptest"
)

func TestInstallURLIsCachedAndAFailureIsRetriedLater(t *testing.T) {
	fake := githubapptest.New(t)
	ctx := context.Background()
	now := time.Now()
	githubapp.SetClock(fake.Client, func() time.Time { return now })
	fake.AppDown = true
	if got := fake.Client.InstallURL(ctx); got != "" {
		t.Fatalf("lookup failed but got %q", got)
	}
	if got := fake.Client.InstallURL(ctx); got != "" || fake.AppCalls() != 1 {
		t.Fatalf("a failure must back off: %q after %d calls", got, fake.AppCalls())
	}
	fake.AppDown = false
	now = now.Add(2 * time.Minute)
	want := "https://github.com/apps/facile-opus/installations/new"
	for range 2 {
		if got := fake.Client.InstallURL(ctx); got != want {
			t.Fatalf("install url = %q", got)
		}
	}
	if fake.AppCalls() != 2 {
		t.Fatalf("a success must be cached: %d calls", fake.AppCalls())
	}
	var off *githubapp.Client
	if off.InstallURL(ctx) != "" {
		t.Fatal("a nil client has no install url")
	}
}
