package github

import (
	"context"
	"sort"

	"github.com/FacileStudio/Opus/apps/api/internal/githubapp"
	"github.com/FacileStudio/tronc/errors"
)

// Issues returns the issues of every repository linked to a project, pull
// requests excluded, at most 100 per repository. Any project member may read
// them. A repository GitHub refuses is listed under Unavailable with the
// reason, so one uninstalled repository does not hide the others.
func (r *Repos) Issues(ctx context.Context, userID, projectID int64, state string) (*IssuesResponse, error) {
	if state == "" {
		state = "open"
	}
	if state != "open" && state != "closed" {
		return nil, errors.Invalid("state is open or closed")
	}
	if err := r.requireApp(); err != nil {
		return nil, err
	}
	if err := r.requireRole(ctx, userID, projectID); err != nil {
		return nil, err
	}
	rows, err := r.linked(ctx, projectID)
	if err != nil {
		return nil, err
	}
	out := &IssuesResponse{Issues: []IssueResponse{}, Unavailable: []UnavailableRepo{}}
	for _, row := range rows {
		issues, err := r.app.Issues(ctx, row.RepoFullName, state)
		if err != nil {
			out.Unavailable = append(out.Unavailable, UnavailableRepo{Repo: row.RepoFullName, Reason: appError(err, row.RepoFullName).Error()})
			continue
		}
		for _, issue := range issues {
			out.Issues = append(out.Issues, toIssueResponse(row.RepoFullName, issue))
		}
	}
	sort.SliceStable(out.Issues, func(i, j int) bool {
		return out.Issues[i].UpdatedAt.After(out.Issues[j].UpdatedAt)
	})
	return out, nil
}

func toIssueResponse(repo string, issue githubapp.Issue) IssueResponse {
	labels := make([]IssueLabel, 0, len(issue.Labels))
	for _, label := range issue.Labels {
		labels = append(labels, IssueLabel(label))
	}
	response := IssueResponse{
		Repo:      repo,
		Number:    issue.Number,
		Title:     issue.Title,
		URL:       issue.HTMLURL,
		State:     issue.State,
		Comments:  issue.Comments,
		Labels:    labels,
		CreatedAt: issue.CreatedAt,
		UpdatedAt: issue.UpdatedAt,
	}
	if issue.User != nil {
		response.Author = &IssueAuthor{Login: issue.User.Login, AvatarURL: issue.User.AvatarURL, URL: issue.User.HTMLURL}
	}
	return response
}
