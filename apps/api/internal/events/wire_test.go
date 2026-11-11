package events

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"

	envelope "github.com/FacileStudio/antenne-client/go/envelope"
)

var updateGolden = flag.Bool("update-golden", false, "rewrite testdata/events from the current envelopes")

func wireSamples() map[string]any {
	due := "2026-11-20"
	actor := int64(7)
	description := "Plan the launch"
	icon := "rocket"
	task := TaskPayload{
		ID: 42, FacileID: "01JTASK00000000000000000000", ProjectID: 3, ProjectFacileID: "01JPROJ00000000000000000000",
		WorkspaceID: 1, Title: "Ship it", Name: "Ship it", Status: "in-progress", Description: "Before Friday",
		Priority: "high", DueDate: &due, ActorID: &actor, ActorEmail: "camille@facile.studio",
		ActorSubject: "sub-camille", OwnerEmail: "owner@facile.studio", UpdatedAt: "2026-11-11T09:30:00.123Z",
	}
	project := ProjectPayload{
		ID: 3, FacileID: "01JPROJ00000000000000000000", WorkspaceID: 1, Name: "Launch", Slug: "launch",
		Description: &description, Icon: &icon, ActorEmail: "camille@facile.studio",
		Space:     &SpacePayload{RegistreSpaceID: "space-1", OwnerSubject: "sub-owner", OwnerEmail: "owner@facile.studio"},
		UpdatedAt: "2026-11-11T09:30:00Z",
	}
	samples := map[string]any{
		ChannelCommentCreated: CommentPayload{ID: 9, FacileID: "9", TaskID: 42, UserID: 7, Content: "Looks good", ActorEmail: "camille@facile.studio"},
		ChannelWorkspaceCreated: WorkspacePayload{ID: 1, FacileID: "studio", Name: "Studio", Slug: "studio",
			Description: &description, ActorEmail: "camille@facile.studio"},
		"task.minimal":    TaskPayload{ID: 1, ProjectID: 1, WorkspaceID: 1, Title: "t", Name: "t", Priority: "none"},
		"project.minimal": ProjectPayload{ID: 1, WorkspaceID: 1, Name: "p", Slug: "p"},
	}
	for _, channel := range []string{ChannelTaskCreated, ChannelTaskUpdated, ChannelTaskDeleted, ChannelTaskStatusChanged, ChannelTaskActorChanged} {
		samples[channel] = task
	}
	for _, channel := range []string{ChannelProjectCreated, ChannelProjectUpdated, ChannelProjectDeleted} {
		samples[channel] = project
	}
	return samples
}

func wireEnvelope(t *testing.T, name string, payload any) []byte {
	t.Helper()
	now := time.Date(2026, 11, 11, 9, 30, 0, 123456789, time.UTC)
	var raw []byte
	var err error
	switch p := payload.(type) {
	case TaskPayload:
		channel := name
		if name == "task.minimal" {
			channel = ChannelTaskUpdated
		}
		raw, err = json.Marshal(BuildEnvelope(channel, p.entityID(), envelope.ObjectTask, p, now))
	case ProjectPayload:
		channel := name
		if name == "project.minimal" {
			channel = ChannelProjectUpdated
		}
		raw, err = json.Marshal(BuildEnvelope(channel, p.entityID(), envelope.ObjectProject, p, now))
	case CommentPayload:
		raw, err = json.Marshal(BuildEnvelope(name, p.entityID(), ObjectComment, p, now))
	case WorkspacePayload:
		raw, err = json.Marshal(BuildEnvelope(name, p.entityID(), ObjectWorkspace, p, now))
	}
	if err != nil {
		t.Fatalf("%s: marshal: %v", name, err)
	}
	return raw
}

func TestPublishedEventBytesArePinned(t *testing.T) {
	samples := wireSamples()
	for _, channel := range AllChannels() {
		if _, ok := samples[channel]; !ok {
			t.Errorf("channel %s has no pinned sample", channel)
		}
	}
	for name, payload := range samples {
		got := wireEnvelope(t, name, payload)
		path := filepath.Join("testdata", "events", name+".json")
		if *updateGolden {
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, append(got, '\n'), 0o644); err != nil {
				t.Fatal(err)
			}
			continue
		}
		want, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if string(got)+"\n" != string(want) {
			t.Errorf("%s changed on the wire:\n got %s\nwant %s", name, got, want)
		}
	}
}
