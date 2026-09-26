package asset

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/FacileStudio/Opus/apps/api/internal/authcontext"
	"github.com/FacileStudio/Opus/apps/api/internal/middleware"
	"github.com/FacileStudio/Opus/apps/api/internal/storage"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes mounts all asset and image upload endpoints onto the Chi router.
func RegisterRoutes(router chi.Router, service *Service, authService middleware.Authenticator) {
	router.Route("/asset", func(sub chi.Router) {
		registerAssetDownloadRoutes(sub, service, authService)
	})
	router.Route("/task/image-upload", func(sub chi.Router) {
		registerTaskUploadRoutes(sub, service, authService)
	})
}

func registerAssetDownloadRoutes(router chi.Router, service *Service, authService middleware.Authenticator) {
	router.With(middleware.OptionalAuth(authService)).Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		assetID := chi.URLParam(r, "id")
		shareToken := r.URL.Query().Get("share_token")
		var callerID int64
		if identity, ok := authcontext.IdentityFromContext(r.Context()); ok {
			callerID = identity.UserID
		}
		download, err := service.controller.getAsset(r.Context(), callerID, assetID, shareToken)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		defer download.Body.Close()
		writeAssetStream(w, download)
	})
	router.Put("/upload/*", func(w http.ResponseWriter, r *http.Request) {
		key := chi.URLParam(r, "*")
		if key == "" {
			key = strings.TrimPrefix(r.URL.Path, "/asset/upload/")
			key = strings.TrimPrefix(key, "/api/asset/upload/")
		}
		if err := service.controller.storeDirectUpload(key, r.Body); err != nil {
			httpjson.WriteError(w, err)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}

func writeAssetStream(w http.ResponseWriter, download *AssetDownload) {
	cacheControl := "private, max-age=120"
	if download.IsPublic {
		cacheControl = "public, max-age=300"
	}
	w.Header().Set("Cache-Control", cacheControl)
	w.Header().Set("Content-Disposition", storage.BuildContentDisposition(download.Filename))
	w.Header().Set("Content-Type", download.MimeType)
	if download.ContentLength > 0 {
		w.Header().Set("Content-Length", strconv.FormatInt(download.ContentLength, 10))
	}
	if !download.LastModified.IsZero() {
		w.Header().Set("Last-Modified", download.LastModified.UTC().Format(http.TimeFormat))
	}
	w.WriteHeader(http.StatusOK)
	if _, err := io.Copy(w, download.Body); err != nil {
		return
	}
}

func registerTaskUploadRoutes(router chi.Router, service *Service, authService middleware.Authenticator) {
	router.Use(middleware.RequireAuth(authService))
	router.Put("/{id}", func(w http.ResponseWriter, r *http.Request) {
		taskID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid task id"))
			return
		}
		var req CreateImageUploadRequest
		if decErr := json.NewDecoder(r.Body).Decode(&req); decErr != nil {
			httpjson.WriteError(w, errors.Invalid("invalid request body"))
			return
		}
		identity, _ := authcontext.IdentityFromContext(r.Context())
		resp, err := service.controller.createImageUpload(r.Context(), identity.UserID, taskID, req)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
	router.Post("/{id}/finalize", func(w http.ResponseWriter, r *http.Request) {
		taskID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid task id"))
			return
		}
		var req FinalizeImageUploadRequest
		if decErr := json.NewDecoder(r.Body).Decode(&req); decErr != nil {
			httpjson.WriteError(w, errors.Invalid("invalid request body"))
			return
		}
		identity, _ := authcontext.IdentityFromContext(r.Context())
		resp, err := service.controller.finalizeImageUpload(r.Context(), identity.UserID, taskID, req)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}
