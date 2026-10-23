package notification

import (
	documentation "github.com/FacileStudio/Opus/apps/api/internal/documentation"
)

func notificationCoreRoutes() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "GET",
			Path:         "/notification",
			Summary:      "List notifications",
			Description:  "Get all in-app notifications for the authenticated user.",
			Auth:         "bearer",
			ResponseBody: `[{"id":int64,"title":string,"content":string,"isRead":bool}]`,
		},
		{
			Method:       "POST",
			Path:         "/notification",
			Summary:      "Create a notification",
			Description:  "Record an in-app alert for the user.",
			Auth:         "bearer",
			RequestBody:  `{"type":string,"title":string,"message":string}`,
			ResponseBody: `{"id":int64,"title":string,"content":string,"isRead":bool}`,
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Notification type is required"},
			},
		},
		{
			Method:       "PATCH",
			Path:         "/notification/{id}/read",
			Summary:      "Mark notification as read",
			Description:  "Flag a single notification as read.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "int64", Description: "Notification ID"}},
			ResponseBody: `{"id":int64,"isRead":true}`,
			Errors: []documentation.Error{
				{Status: 404, Code: "not_found", Description: "Notification not found"},
			},
		},
		{
			Method:       "PATCH",
			Path:         "/notification/{id}/unread",
			Summary:      "Mark notification as unread",
			Description:  "Flag a single notification as unread.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "int64", Description: "Notification ID"}},
			ResponseBody: `{"id":int64,"isRead":false}`,
			Errors: []documentation.Error{
				{Status: 404, Code: "not_found", Description: "Notification not found"},
			},
		},
		{
			Method:       "DELETE",
			Path:         "/notification/{id}",
			Summary:      "Delete notification",
			Description:  "Delete a single notification.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "int64", Description: "Notification ID"}},
			ResponseBody: `{"success":true,"count":1}`,
			Errors: []documentation.Error{
				{Status: 404, Code: "not_found", Description: "Notification not found"},
			},
		},
		{
			Method:       "PATCH",
			Path:         "/notification/read-all",
			Summary:      "Mark all notifications read",
			Description:  "Mark all pending notifications as read.",
			Auth:         "bearer",
			ResponseBody: `{"success":true,"count":int64}`,
		},
		{
			Method:       "DELETE",
			Path:         "/notification/clear-all",
			Summary:      "Clear all notifications",
			Description:  "Delete all notifications for the user.",
			Auth:         "bearer",
			ResponseBody: `{"success":true,"count":int64}`,
		},
	}
}

func preferenceRoutes() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "GET",
			Path:         "/notification-preferences",
			Summary:      "Get notification preferences",
			Description:  "Fetch global and workspace delivery channels.",
			Auth:         "bearer",
			ResponseBody: `{"emailEnabled":bool,"ntfyEnabled":bool,"workspaces":[]}`,
		},
		{
			Method:       "PUT",
			Path:         "/notification-preferences",
			Summary:      "Update notification preferences",
			Description:  "Update global delivery settings.",
			Auth:         "bearer",
			RequestBody:  `{"emailEnabled":bool,"ntfyEnabled":bool}`,
			ResponseBody: `{"emailEnabled":bool,"ntfyEnabled":bool}`,
		},
		{
			Method:       "PUT",
			Path:         "/notification-preferences/workspaces/{workspaceId}",
			Summary:      "Upsert workspace notification rule",
			Description:  "Configure workspace-specific notifications.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "workspaceId", Type: "int64", Description: "Workspace ID"}},
			RequestBody:  `{"isActive":bool,"projectMode":string}`,
			ResponseBody: `{"workspaces":[]}`,
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "Caller lacks workspace access"},
			},
		},
		{
			Method:       "DELETE",
			Path:         "/notification-preferences/workspaces/{workspaceId}",
			Summary:      "Delete workspace notification rule",
			Description:  "Remove workspace override rules.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "workspaceId", Type: "int64", Description: "Workspace ID"}},
			ResponseBody: `{"workspaces":[]}`,
			Errors: []documentation.Error{
				{Status: 404, Code: "not_found", Description: "Rule not found"},
			},
		},
	}
}

// Documentation returns API documentation metadata for notifications.
func Documentation() documentation.Module {
	return documentation.Module{
		Name:        "notification",
		Description: "In-app notifications and delivery preferences.",
		Routes:      append(notificationCoreRoutes(), preferenceRoutes()...),
	}
}
