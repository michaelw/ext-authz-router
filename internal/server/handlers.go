package server

import (
	"bytes"
	"context"
	_ "embed"
	"log"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/michaelw/ext-authz-router/api"
)

const (
	COOKIE_NAME       = "namespace"
	COOKIE_DOMAIN     = "int.kube"
	COOKIE_EXPIRATION = 24 * time.Hour

	REDIRECT_URL = "http://namespaces.int.kube/"

	CONFIG_PATH = "/app/config/config.yaml"
)

var (
	//go:embed assets/index.html
	namespaceSelectionSPA string
)

type HandlerOption func(*AuthzHandler)

func NewServerHandler(publicURL string, swagger *openapi3.T, opts ...HandlerOption) *AuthzHandler {
	handler := &AuthzHandler{
		PublicURL:  publicURL,
		Swagger:    swagger,
		configPath: CONFIG_PATH,
	}
	for _, opt := range opts {
		opt(handler)
	}
	log.Printf("I: configuration file: %v", handler.configPath)
	if err := handler.loadConfig(); err != nil {
		log.Printf("E: failed to load configuration: %v", err)
	}

	return handler
}

// Get handles GET / - Namespace selection UI
func (h *AuthzHandler) Get(ctx context.Context, request api.GetRequestObject) (api.GetResponseObject, error) {
	var buf bytes.Buffer
	if _, err := buf.WriteString(namespaceSelectionSPA); err != nil {
		return nil, err
	}

	return api.Get200TexthtmlResponse{
		Body:          &buf,
		ContentLength: int64(buf.Len()),
	}, nil
}

// GetNamespaces handles GET /namespaces - Returns available namespaces
func (h *AuthzHandler) GetNamespaces(ctx context.Context, request api.GetNamespacesRequestObject) (api.GetNamespacesResponseObject, error) {
	h.configLock.RLock()
	namespaces := h.currentConfig.Namespaces
	h.configLock.RUnlock()

	ns := map[string]api.NamespaceAttributes{}
	for id, attrs := range namespaces {
		desc := attrs.Description
		if desc == "" {
			desc = id
		}
		ns[id] = api.NamespaceAttributes{
			Description: &desc,
		}
	}

	return api.GetNamespaces200JSONResponse{
		Namespaces: ns,
	}, nil
}

// PostSubmit handles POST /namespace - Set namespace cookie
func (h *AuthzHandler) PostSubmit(ctx context.Context, request api.PostSubmitRequestObject) (api.PostSubmitResponseObject, error) {
	var namespace string

	// Support both JSON and form data
	if request.JSONBody != nil {
		namespace = request.JSONBody.Value
	} else if request.FormdataBody != nil {
		namespace = request.FormdataBody.Value
	}

	if namespace == "" {
		return api.PostSubmit400JSONResponse{}, nil
	}

	h.configLock.RLock()
	_, ok := h.currentConfig.Namespaces[namespace]
	h.configLock.RUnlock()

	if !ok {
		return api.PostSubmit400JSONResponse{}, nil
	}

	redirectTo := REDIRECT_URL
	if request.Params.RedirectTo != nil {
		redirectTo = *request.Params.RedirectTo
	}

	return api.PostSubmit302JSONResponse{
		Headers: api.PostSubmit302ResponseHeaders{
			Location:  redirectTo,
			SetCookie: COOKIE_NAME + "=" + namespace + "; Path=/; Domain=" + COOKIE_DOMAIN + "; Expires=" + time.Now().Add(COOKIE_EXPIRATION).UTC().Format(time.RFC1123) + "; HttpOnly", // XXX, also secure
		},
	}, nil
}
