package github

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/FacileStudio/Opus/apps/api/internal/middleware"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes mounts GitHub integration endpoints into the router.
func RegisterRoutes(api chi.Router, service *Service, authService middleware.Authenticator) {
	api.Route("/integrations/github", func(r chi.Router) {
		r.Post("/webhook", func(w http.ResponseWriter, r *http.Request) {
			handleWebhook(w, r, service)
		})
	})

	api.Route("/project/{projectId}/task/{taskId}/pull-requests", func(r chi.Router) {
		r.Use(middleware.RequireAuth(authService))
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			handleListPullRequests(w, r, service)
		})
	})
}

func handleWebhook(w http.ResponseWriter, r *http.Request, service *Service) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 10<<20))
	if err != nil {
		httpjson.WriteError(w, errors.Invalid("failed to read webhook body"))
		return
	}
	signature := r.Header.Get("X-Hub-Signature-256")
	if !verifySignature(body, signature, service.secret) {
		httpjson.WriteError(w, errors.Unauthorized("invalid webhook signature"))
		return
	}
	var payload WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		httpjson.WriteError(w, errors.Invalid("invalid json payload"))
		return
	}
	if err := service.ProcessWebhook(r.Context(), &payload); err != nil {
		httpjson.WriteError(w, errors.Internal("failed to process webhook", err))
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func handleListPullRequests(w http.ResponseWriter, r *http.Request, service *Service) {
	taskID, err := strconv.ParseInt(chi.URLParam(r, "taskId"), 10, 64)
	if err != nil || taskID <= 0 {
		httpjson.WriteError(w, errors.Invalid("invalid task id"))
		return
	}
	rows, err := service.ListPullRequestsByTask(r.Context(), taskID)
	if err != nil {
		httpjson.WriteError(w, errors.Internal("failed to list pull requests", err))
		return
	}
	responses := make([]PullRequestResponse, 0, len(rows))
	for _, row := range rows {
		responses = append(responses, toPullRequestResponse(row))
	}
	httpjson.WriteJSON(w, http.StatusOK, responses)
}
