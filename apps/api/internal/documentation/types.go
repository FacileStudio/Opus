// Package docs carries the route registry apiref turns into the OpenAPI
// reference served at /docs.
package docs

import "github.com/FacileStudio/tronc/apiref"

// Response is the top-level API reference registry returned by the docs endpoint.
type Response = apiref.Registry

// Module groups the routes belonging to one API module in the docs registry.
type Module = apiref.Module

// Route describes a single documented API endpoint.
type Route = apiref.Route

// Field describes one documented request or response field.
type Field = apiref.Field

// Error describes one documented error response.
type Error = apiref.Error
