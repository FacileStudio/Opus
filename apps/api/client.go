package main

import (
	"os"
	"path/filepath"

	"github.com/FacileStudio/tronc/spa"
	"github.com/go-chi/chi/v5"
)

// appShell is the React app's document once the image renames it to make room
// for the public site, whose own index.html then answers the landing pages.
const appShell = "app.html"

// sitePages are the signed-out pages the SvelteKit site serves. Every other
// path stays the React app's.
var sitePages = []string{"/", "/downloads"}

// mountClient serves the built client. A client dir holding app.html carries
// both builds: the site answers sitePages through its index.html, and the
// React shell is the fallback for everything else. A dir without it is the
// React build alone, served as before.
func mountClient(router chi.Router, clientDir string) {
	index := "index.html"
	if info, err := os.Stat(filepath.Join(clientDir, appShell)); err == nil && !info.IsDir() {
		index = appShell
		site := spa.Handler(spa.Config{Dir: clientDir})
		for _, page := range sitePages {
			router.Handle(page, site)
		}
	}
	router.Handle("/*", spa.Handler(spa.Config{Dir: clientDir, Index: index}))
}
