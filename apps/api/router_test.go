package main

import (
	"testing"
)

func TestReferenceConfig(t *testing.T) {
	cfg := referenceConfig()
	if cfg.Title != "Opus API" {
		t.Fatalf("expected title 'Opus API', got '%s'", cfg.Title)
	}

	expectedModules := map[string]bool{
		"auth":         false,
		"workspace":    false,
		"invitation":   false,
		"project":      false,
		"column":       false,
		"task":         false,
		"taskrelation": false,
		"label":        false,
		"comment":      false,
		"activity":     false,
		"sharelink":    false,
		"timeentry":    false,
		"workflowrule": false,
		"notification": false,
		"search":       false,
	}

	for _, mod := range cfg.Registry.Modules {
		expectedModules[mod.Name] = true
		if len(mod.Routes) == 0 {
			t.Errorf("module %s has 0 routes registered", mod.Name)
		}
	}

	for name, found := range expectedModules {
		if !found {
			t.Errorf("expected module %s was not found in documentation registry", name)
		}
	}
}
