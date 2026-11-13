package github

import (
	"net/http"
	"strconv"

	"github.com/FacileStudio/Opus/apps/api/internal/authcontext"
	"github.com/FacileStudio/Opus/apps/api/internal/middleware"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
	"github.com/go-chi/chi/v5"
)

// RegisterRepoRoutes mounts the linked-repository and issues endpoints.
func RegisterRepoRoutes(api chi.Router, repos *Repos, authService middleware.Authenticator) {
	authed := api.With(middleware.RequireAuth(authService))
	authed.Get("/project/{projectId}/github-repos", func(w http.ResponseWriter, r *http.Request) {
		handleListRepos(w, r, repos)
	})
	authed.Post("/project/{projectId}/github-repos", func(w http.ResponseWriter, r *http.Request) {
		handleLinkRepo(w, r, repos)
	})
	authed.Delete("/project/{projectId}/github-repos/{repoId}", func(w http.ResponseWriter, r *http.Request) {
		handleUnlinkRepo(w, r, repos)
	})
	authed.Get("/project/{projectId}/github-issues", func(w http.ResponseWriter, r *http.Request) {
		handleListIssues(w, r, repos)
	})
}

func pathID(r *http.Request, name string) (int64, error) {
	id, err := strconv.ParseInt(chi.URLParam(r, name), 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.Invalid("invalid " + name)
	}
	return id, nil
}

func callerAndProject(r *http.Request) (int64, int64, error) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		return 0, 0, errors.Unauthorized("missing auth token")
	}
	projectID, err := pathID(r, "projectId")
	return identity.UserID, projectID, err
}

func handleListRepos(w http.ResponseWriter, r *http.Request, repos *Repos) {
	userID, projectID, err := callerAndProject(r)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	out, err := repos.List(r.Context(), userID, projectID)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, out)
}

func handleLinkRepo(w http.ResponseWriter, r *http.Request, repos *Repos) {
	userID, projectID, err := callerAndProject(r)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	var body LinkRepoRequest
	if err := httpjson.DecodeJSON(w, r, &body); err != nil {
		httpjson.WriteError(w, err)
		return
	}
	out, err := repos.Link(r.Context(), userID, projectID, body.Repo)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	httpjson.WriteJSON(w, http.StatusCreated, out)
}

func handleUnlinkRepo(w http.ResponseWriter, r *http.Request, repos *Repos) {
	userID, projectID, err := callerAndProject(r)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	repoID, err := pathID(r, "repoId")
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	if err := repos.Unlink(r.Context(), userID, projectID, repoID); err != nil {
		httpjson.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func handleListIssues(w http.ResponseWriter, r *http.Request, repos *Repos) {
	userID, projectID, err := callerAndProject(r)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	out, err := repos.Issues(r.Context(), userID, projectID, r.URL.Query().Get("state"))
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, out)
}
