package mcp

import (
	"testing"

	"github.com/FacileStudio/tronc/mcp/mcptest"
)

// TestConformance runs tronc's conformance kit on the endpoint, with both
// deletes as the sensitive tools.
func TestConformance(t *testing.T) {
	f := newFixture(t)
	mcptest.Run(t, mcptest.Config{
		URL:   f.resource,
		Token: f.token,
		Confirm: []mcptest.ConfirmCase{
			{Tool: toolDeleteTask, Setup: f.freshTask},
			{Tool: toolDeleteProject, Setup: f.freshProject},
		},
	})
}
