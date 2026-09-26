package search

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/FacileStudio/tronc/errors"
)

// Controller adapts search HTTP queries and validates filters.
type Controller struct {
	service *Service
}

func newController(service *Service) *Controller {
	return &Controller{service: service}
}

func parseLimit(val string) int {
	if val == "" {
		return 20
	}
	parsed, err := strconv.Atoi(val)
	if err != nil || parsed <= 0 {
		return 20
	}
	if parsed > 50 {
		return 50
	}
	return parsed
}

func parseOptionalID(val string) *int64 {
	if val == "" {
		return nil
	}
	parsed, err := strconv.ParseInt(val, 10, 64)
	if err != nil || parsed <= 0 {
		return nil
	}
	return &parsed
}

func (controller *Controller) search(ctx context.Context, userID int64, r *http.Request) (*SearchResponse, error) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		return nil, errors.Invalid("query is required")
	}

	searchType := strings.TrimSpace(r.URL.Query().Get("type"))
	if searchType == "" {
		searchType = "all"
	}

	filter := SearchFilter{
		Query:       q,
		Type:        searchType,
		WorkspaceID: parseOptionalID(r.URL.Query().Get("workspaceId")),
		ProjectID:   parseOptionalID(r.URL.Query().Get("projectId")),
		Limit:       parseLimit(r.URL.Query().Get("limit")),
	}

	return controller.service.Search(ctx, userID, filter)
}
