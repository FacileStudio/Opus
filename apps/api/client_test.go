package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func clientDir(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, body := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func fetchClient(t *testing.T, dir, path string) (int, string) {
	t.Helper()
	router := chi.NewRouter()
	mountClient(router, dir)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
	return recorder.Code, strings.TrimSpace(recorder.Body.String())
}

func TestTheSiteAnswersTheLandingPagesAndTheAppTheRest(t *testing.T) {
	dir := clientDir(t, map[string]string{
		"index.html":             "site",
		"app.html":               "app",
		"_app/immutable/site.js": "site-js",
		"assets/app.js":          "app-js",
	})
	cases := map[string]string{
		"/":                       "site",
		"/downloads":              "site",
		"/dashboard":              "app",
		"/auth/sign-in":           "app",
		"/downloads/x":            "app",
		"/_app/immutable/site.js": "site-js",
		"/assets/app.js":          "app-js",
	}
	for path, want := range cases {
		if code, body := fetchClient(t, dir, path); code != http.StatusOK || body != want {
			t.Errorf("GET %s = %d %q, want 200 %q", path, code, body, want)
		}
	}
}

func TestAReactOnlyBuildKeepsItsIndexEverywhere(t *testing.T) {
	dir := clientDir(t, map[string]string{"index.html": "app"})
	for _, path := range []string{"/", "/downloads", "/dashboard"} {
		if code, body := fetchClient(t, dir, path); code != http.StatusOK || body != "app" {
			t.Errorf("GET %s = %d %q, want 200 %q", path, code, body, "app")
		}
	}
}
