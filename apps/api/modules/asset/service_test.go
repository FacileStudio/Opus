package asset

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/FacileStudio/Opus/apps/api/internal/database/dbtest"
	"github.com/FacileStudio/Opus/apps/api/internal/storage"
	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/porte"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type testAuth struct {
	userID int64
	email  string
}

func (a *testAuth) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := porte.WithIdentity(r.Context(), porte.Identity{UserID: a.userID, SessionID: 1})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *testAuth) Optional(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.userID > 0 {
			ctx := porte.WithIdentity(r.Context(), porte.Identity{UserID: a.userID, SessionID: 1})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *testAuth) IdentityForUser(_ context.Context, userID int64) (int64, string, error) {
	return userID, a.email, nil
}

func setupTestEnv(t *testing.T) (*gorm.DB, *Service, string) {
	t.Helper()
	db := dbtest.Open(t)
	tempDir, err := os.MkdirTemp("", "opus_asset_test_*")
	if err != nil {
		t.Fatalf("temp dir: %v", err)
	}
	assetStorage := storage.NewAssetStorage(tempDir, nil)
	service := NewService(db, assetStorage)
	return db, service, tempDir
}

func seedFixture(t *testing.T, db *gorm.DB) (int64, int64, int64) {
	t.Helper()
	create := func(row any) {
		if err := db.Create(row).Error; err != nil {
			t.Fatalf("seed %T: %v", row, err)
		}
	}
	user := schemas.User{Email: "test@facile.studio", Name: "Tester"}
	create(&user)
	workspace := schemas.Workspace{Name: "Workspace", Slug: "workspace"}
	create(&workspace)
	create(&schemas.WorkspaceMember{WorkspaceID: workspace.ID, UserID: user.ID, Role: "admin"})
	project := schemas.Project{WorkspaceID: workspace.ID, Name: "Project", Slug: "project"}
	create(&project)
	task := schemas.Task{WorkspaceID: workspace.ID, ProjectID: project.ID, Title: "Task"}
	create(&task)
	return user.ID, project.ID, task.ID
}

func executeCreateUpload(t *testing.T, r http.Handler, taskID int64) CreateImageUploadResponse {
	t.Helper()
	uploadReq := CreateImageUploadRequest{
		Filename:    "test-image.png",
		ContentType: "image/png",
		Size:        100,
		Surface:     "description",
	}
	reqBody, _ := json.Marshal(uploadReq)
	req := httptest.NewRequest(http.MethodPut, "/task/image-upload/"+strconv.FormatInt(taskID, 10), bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("create upload status %d: %s", w.Code, w.Body.String())
	}
	var resp CreateImageUploadResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil || resp.Key == "" {
		t.Fatalf("invalid create upload response: %v", err)
	}
	return resp
}

func executeDirectAndFinalize(t *testing.T, r http.Handler, taskID int64, upload CreateImageUploadResponse, data []byte) string {
	t.Helper()
	key := upload.Key
	putReq := httptest.NewRequest(http.MethodPut, strings.TrimPrefix(upload.UploadURL, "/api"), bytes.NewReader(data))
	putW := httptest.NewRecorder()
	r.ServeHTTP(putW, putReq)
	if putW.Code != http.StatusOK {
		t.Fatalf("direct upload status %d: %s", putW.Code, putW.Body.String())
	}
	finReq := FinalizeImageUploadRequest{
		Key:         key,
		Filename:    "test-image.png",
		ContentType: "image/png",
		Size:        int64(len(data)),
		Surface:     "description",
	}
	body, _ := json.Marshal(finReq)
	postReq := httptest.NewRequest(http.MethodPost, "/task/image-upload/"+strconv.FormatInt(taskID, 10)+"/finalize", bytes.NewReader(body))
	postReq.Header.Set("Content-Type", "application/json")
	postW := httptest.NewRecorder()
	r.ServeHTTP(postW, postReq)
	if postW.Code != http.StatusOK {
		t.Fatalf("finalize upload status %d: %s", postW.Code, postW.Body.String())
	}
	var finResp FinalizeImageUploadResponse
	if err := json.Unmarshal(postW.Body.Bytes(), &finResp); err != nil || finResp.ID == "" {
		t.Fatalf("invalid finalize response: %v", err)
	}
	return finResp.ID
}

func TestAssetUploadAndDownload(t *testing.T) {
	db, service, tempDir := setupTestEnv(t)
	defer os.RemoveAll(tempDir)
	userID, projectID, taskID := seedFixture(t, db)

	auth := &testAuth{userID: userID, email: "test@facile.studio"}
	r := chi.NewRouter()
	RegisterRoutes(r, service, auth)

	upload := executeCreateUpload(t, r, taskID)
	directData := []byte("image binary payload")
	assetID := executeDirectAndFinalize(t, r, taskID, upload, directData)

	getReq := httptest.NewRequest(http.MethodGet, "/asset/"+assetID, nil)
	getW := httptest.NewRecorder()
	r.ServeHTTP(getW, getReq)
	if getW.Code != http.StatusOK {
		t.Fatalf("asset download status %d: %s", getW.Code, getW.Body.String())
	}
	if !bytes.Equal(getW.Body.Bytes(), directData) {
		t.Fatalf("expected body %s, got %s", directData, getW.Body.Bytes())
	}

	testShareTokenAccess(t, db, r, projectID, assetID)
	testUnauthorizedAccess(t, service, assetID)
}

func testShareTokenAccess(t *testing.T, db *gorm.DB, r http.Handler, projectID int64, assetID string) {
	t.Helper()
	expires := time.Now().Add(1 * time.Hour)
	link := schemas.ShareLink{
		ProjectID: projectID,
		Token:     "test_share_tok",
		IsActive:  true,
		ExpiresAt: &expires,
	}
	db.Create(&link)

	getReq := httptest.NewRequest(http.MethodGet, "/asset/"+assetID+"?share_token=test_share_tok", nil)
	getW := httptest.NewRecorder()
	r.ServeHTTP(getW, getReq)
	if getW.Code != http.StatusOK {
		t.Fatalf("share token download status %d", getW.Code)
	}
}

func testUnauthorizedAccess(t *testing.T, service *Service, assetID string) {
	t.Helper()
	anonAuth := &testAuth{userID: 0}
	anonR := chi.NewRouter()
	RegisterRoutes(anonR, service, anonAuth)

	getReq := httptest.NewRequest(http.MethodGet, "/asset/"+assetID+"?share_token=invalid", nil)
	getW := httptest.NewRecorder()
	anonR.ServeHTTP(getW, getReq)
	if getW.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized access status %d", getW.Code)
	}
}
