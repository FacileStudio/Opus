// Package asset implements asset and image upload API documentation.
package asset

import documentation "github.com/FacileStudio/Opus/apps/api/internal/documentation"

func assetRoutes() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "GET",
			Path:         "/asset/{id}",
			Summary:      "Download asset",
			Description:  "Streams the asset object. Requires workspace membership, public project, or valid share token.",
			Auth:         "optional",
			PathParams:   []documentation.Field{{Name: "id", Type: "string", Description: "Asset ID"}},
			ResponseBody: "binary data",
			Errors: []documentation.Error{
				{Status: 401, Code: "unauthorized", Description: "Authentication required or invalid share token"},
				{Status: 404, Code: "not_found", Description: "Asset not found"},
			},
		},
		{
			Method:       "PUT",
			Path:         "/task/image-upload/{id}",
			Summary:      "Create task image upload URL",
			Description:  "Generates an upload destination and required headers for task assets.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "int64", Description: "Task ID"}},
			RequestBody:  `{"filename":"photo.png","contentType":"image/png","size":1024,"surface":"description"}`,
			ResponseBody: `{"key":"workspace/1/project/1/task/1/descriptions/photo.png","uploadUrl":"/api/asset/upload/...","headers":{"Content-Type":"image/png"}}`,
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid payload or exceeds size limit"},
				{Status: 403, Code: "permission_denied", Description: "Not a workspace member"},
				{Status: 404, Code: "not_found", Description: "Task not found"},
			},
		},
		{
			Method:       "POST",
			Path:         "/task/image-upload/{id}/finalize",
			Summary:      "Finalize task image upload",
			Description:  "Confirms an uploaded asset and records it in the database.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "int64", Description: "Task ID"}},
			RequestBody:  `{"key":"...","filename":"photo.png","contentType":"image/png","size":1024,"surface":"description"}`,
			ResponseBody: `{"id":"ast_123","url":"/api/asset/ast_123"}`,
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid payload or mismatched key"},
				{Status: 403, Code: "permission_denied", Description: "Not a workspace member"},
				{Status: 404, Code: "not_found", Description: "Task not found"},
			},
		},
	}
}

// Documentation returns the asset module API reference metadata.
func Documentation() documentation.Module {
	return documentation.Module{
		Name:        "asset",
		Description: "Asset storage, direct uploads, and access-controlled binary streaming.",
		Routes:      assetRoutes(),
	}
}
