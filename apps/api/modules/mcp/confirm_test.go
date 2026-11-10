package mcp

import (
	"testing"

	"github.com/FacileStudio/Opus/apps/api/modules/project"
)

func TestProjectQuestionCountsTasks(t *testing.T) {
	cases := []struct {
		stats *project.ProjectStatistics
		want  string
	}{
		{nil, `Delete the project "Site" for good?`},
		{&project.ProjectStatistics{TotalTasks: 0}, `Delete the project "Site" for good?`},
		{&project.ProjectStatistics{TotalTasks: 1}, `Delete the project "Site" and its 1 task for good?`},
		{&project.ProjectStatistics{TotalTasks: 2}, `Delete the project "Site" and its 2 tasks for good?`},
	}
	for _, c := range cases {
		got := projectQuestion(project.ProjectResponse{Name: "Site", Statistics: c.stats})
		if got != c.want {
			t.Errorf("got %q, want %q", got, c.want)
		}
	}
}
