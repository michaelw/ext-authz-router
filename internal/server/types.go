package server

import (
	"github.com/getkin/kin-openapi/openapi3"

	"github.com/michaelw/ext-authz-router/api"
)

var _ api.StrictServerInterface = (*TenantHandler)(nil)

type TenantHandler struct {
	PublicURL string
	Swagger   *openapi3.T
	// Add any dependencies or configuration needed for the handler

	// Embed Unimplemented to provide default implementations
	api.Unimplemented
}
