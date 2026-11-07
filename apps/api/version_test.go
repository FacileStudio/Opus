package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestVersionScriptServesThePackageVersion(t *testing.T) {
	raw, err := os.ReadFile("../../package.json")
	if err != nil {
		t.Fatal(err)
	}
	var pkg struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(raw, &pkg); err != nil || pkg.Version == "" {
		t.Fatalf("package.json has no version: %v", err)
	}
	out, err := exec.Command("sh", "../../scripts/version.sh", "../../package.json").Output()
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(string(out)); got != "v"+pkg.Version {
		t.Fatalf("scripts/version.sh = %q, want %q", got, "v"+pkg.Version)
	}
}
