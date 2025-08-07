package server

import (
	"github.com/getkin/kin-openapi/openapi3"
)

type HandlerOption func(*TenantHandler)

func NewServerHandler(publicURL string, swagger *openapi3.T, opts ...HandlerOption) *TenantHandler {
	handler := &TenantHandler{
		PublicURL: publicURL,
		Swagger:   swagger,
	}
	for _, opt := range opts {
		opt(handler)
	}
	return handler
}
