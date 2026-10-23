package poolsync

import (
	"testing"

	"github.com/FacileStudio/Opus/apps/api/schemas"
)

func columns(slugs ...string) []schemas.Column {
	out := make([]schemas.Column, len(slugs))
	for i, slug := range slugs {
		out[i] = schemas.Column{ID: int64(i + 1), Slug: slug, Position: i, IsFinal: slug == "shipped" || slug == "done"}
	}
	return out
}

func TestResolveStatus(t *testing.T) {
	defaults := columns("to-do", "in-progress", "in-review", "done")
	custom := columns("backlog", "doing", "qa", "polish", "shipped")
	tests := []struct {
		name    string
		columns []schemas.Column
		wire    string
		current string
		want    string
		found   bool
	}{
		{"exact slug on default workflow", defaults, "in-review", "to-do", "in-review", true},
		{"to-do lands on first open column", custom, "to-do", "doing", "backlog", true},
		{"in-progress lands on second open column", custom, "in-progress", "backlog", "doing", true},
		{"in-review lands on third open column", custom, "in-review", "backlog", "qa", true},
		{"done lands on the final column", custom, "done", "doing", "shipped", true},
		{"unknown value is left alone", custom, "blocked", "doing", "", false},
		{"empty is left alone", custom, "", "doing", "", false},
		{"equivalent status is not moved", custom, "in-review", "qa", "", false},
		{"equivalent beyond rank is not moved", custom, "in-review", "polish", "", false},
		{"no final column keeps done unmapped", columns("a", "b"), "done", "a", "", false},
		{"two open columns clamp in-review", columns("a", "b", "done"), "in-review", "a", "b", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, found := resolveStatus(tc.columns, tc.wire, tc.current)
			if found != tc.found || got.slug != tc.want {
				t.Errorf("resolveStatus(%q, %q) = %q/%v, want %q/%v", tc.wire, tc.current, got.slug, found, tc.want, tc.found)
			}
		})
	}
}

func TestDefaultStatus(t *testing.T) {
	if got := defaultStatus(columns("backlog", "shipped")); got.slug != "backlog" || got.columnID == nil {
		t.Errorf("default = %+v, want backlog column", got)
	}
	if got := defaultStatus(nil); got.slug != "to-do" || got.columnID != nil {
		t.Errorf("default without columns = %+v, want bare to-do", got)
	}
}
