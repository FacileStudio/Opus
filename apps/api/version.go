package main

import (
	"net/http"

	"github.com/FacileStudio/tronc/httpjson"
)

// version is the build's release: the Dockerfile reads the root package.json through
// scripts/version.sh and sets it with -ldflags "-X main.version=...".
var version = "dev"

func versionHandler(w http.ResponseWriter, _ *http.Request) {
	httpjson.WriteJSON(w, http.StatusOK, map[string]string{"version": version})
}
