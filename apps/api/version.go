package main

import (
	"net/http"

	"github.com/FacileStudio/tronc/httpjson"
)

// version is the build's release, set by the APP_VERSION build arg through
// -ldflags "-X main.version=...".
var version = "dev"

func versionHandler(w http.ResponseWriter, _ *http.Request) {
	httpjson.WriteJSON(w, http.StatusOK, map[string]string{"version": version})
}
