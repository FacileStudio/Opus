package column

import documentation "github.com/FacileStudio/Opus/apps/api/internal/documentation"

func columnCollectionRoutes() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "GET",
			Path:         "/column/{projectId}",
			Summary:      "Get all columns for a project",
			Description:  "List of columns ordered by position.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "projectId", Type: "int64", Description: "The project's id"}},
			ResponseBody: `[{"id":int64,"projectId":int64,"name":string,"slug":string,"position":int,"icon":string,"color":string,"isFinal":bool,"createdAt":string,"updatedAt":string}]`,
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "The caller is not a member of the workspace"},
				{Status: 404, Code: "not_found", Description: "Project not found"},
			},
		},
		{
			Method:       "POST",
			Path:         "/column/{projectId}",
			Summary:      "Create a new column in a project",
			Description:  "Creates a column positioned after the current last column.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "projectId", Type: "int64", Description: "The project's id"}},
			RequestBody:  `{"name":string,"icon":string?,"color":string?,"isFinal":bool?}`,
			ResponseBody: `{"id":int64,"projectId":int64,"name":string,"slug":string,"position":int,"icon":string,"color":string,"isFinal":bool,"createdAt":string,"updatedAt":string}`,
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Column name must contain at least one alphanumeric character"},
				{Status: 403, Code: "permission_denied", Description: "The caller is not a member of the workspace"},
				{Status: 409, Code: "already_exists", Description: "Column with slug already exists or is reserved"},
			},
		},
	}
}

func columnItemRoutes() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "PUT",
			Path:         "/column/{id}",
			Summary:      "Update a column",
			Description:  "Updates name, icon, color, or final flag of a column.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "int64", Description: "The column's id"}},
			RequestBody:  `{"name":string?,"icon":string?,"color":string?,"isFinal":bool?}`,
			ResponseBody: `{"id":int64,"projectId":int64,"name":string,"slug":string,"position":int,"icon":string,"color":string,"isFinal":bool,"createdAt":string,"updatedAt":string}`,
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "The caller is not a member of the workspace"},
				{Status: 404, Code: "not_found", Description: "Column not found"},
			},
		},
		{
			Method:       "DELETE",
			Path:         "/column/{id}",
			Summary:      "Delete a column",
			Description:  "Removes a column if it contains no tasks.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "int64", Description: "The column's id"}},
			ResponseBody: `{"id":int64,"projectId":int64,"name":string,"slug":string,"position":int,"icon":string,"color":string,"isFinal":bool,"createdAt":string,"updatedAt":string}`,
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "The caller is not a member of the workspace"},
				{Status: 404, Code: "not_found", Description: "Column not found"},
				{Status: 409, Code: "already_exists", Description: "Cannot delete column that contains tasks"},
			},
		},
	}
}

func columnOrderRoutes() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "PUT",
			Path:         "/column/reorder/{projectId}",
			Summary:      "Reorder columns in a project",
			Description:  "Applies new positions to a list of columns in the project.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "projectId", Type: "int64", Description: "The project's id"}},
			RequestBody:  `{"columns":[{"id":int64,"position":int}]}`,
			ResponseBody: `[{"id":int64,"projectId":int64,"name":string,"slug":string,"position":int,"icon":string,"color":string,"isFinal":bool,"createdAt":string,"updatedAt":string}]`,
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Column does not belong to this project"},
				{Status: 403, Code: "permission_denied", Description: "The caller is not a member of the workspace"},
			},
		},
	}
}

// Documentation returns the column module API reference metadata.
func Documentation() documentation.Module {
	return documentation.Module{
		Name:        "column",
		Description: "Columns represent stages in a project workflow board.",
		Routes:      append(columnCollectionRoutes(), append(columnItemRoutes(), columnOrderRoutes()...)...),
	}
}
