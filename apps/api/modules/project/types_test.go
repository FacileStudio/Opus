package project

import (
	"encoding/json"
	"testing"
)

// TestCreateProjectRequestAcceptsStringWorkspaceID pins the exact payload the
// web client sends for POST /project: the frontend normalises ids to strings,
// and tronc's strict decoder turned workspaceId:"7" into 400 invalid JSON body
// before wireid existed.
func TestCreateProjectRequestAcceptsStringWorkspaceID(t *testing.T) {
	var req CreateProjectRequest
	payload := `{"name":"Website","slug":"website","workspaceId":"7"}`
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		t.Fatalf("Unmarshal(%s): %v", payload, err)
	}
	if req.WorkspaceID != 7 {
		t.Fatalf("WorkspaceID = %d, want 7", req.WorkspaceID)
	}

	var numeric CreateProjectRequest
	if err := json.Unmarshal([]byte(`{"name":"Website","workspaceId":7}`), &numeric); err != nil {
		t.Fatalf("Unmarshal numeric id: %v", err)
	}
	if numeric.WorkspaceID != 7 {
		t.Fatalf("numeric WorkspaceID = %d, want 7", numeric.WorkspaceID)
	}
}
