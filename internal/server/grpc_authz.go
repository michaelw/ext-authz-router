package server

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	envoy_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_service_auth_v3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	envoy_type_v3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	grpcstatus "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
)

// AuthzGRPCServer implements the Envoy external authorization gRPC service
type AuthzGRPCServer struct {
	envoy_service_auth_v3.UnimplementedAuthorizationServer
	handler *AuthzHandler
}

// NewAuthzGRPCServer creates a new gRPC authorization server
func NewAuthzGRPCServer(handler *AuthzHandler) *AuthzGRPCServer {
	return &AuthzGRPCServer{
		handler: handler,
	}
}

// Check implements the authorization check
func (s *AuthzGRPCServer) Check(ctx context.Context, req *envoy_service_auth_v3.CheckRequest) (*envoy_service_auth_v3.CheckResponse, error) {
	// Extract request information
	httpReq := req.GetAttributes().GetRequest().GetHttp()
	if httpReq == nil {
		return s.denyResponse(codes.InvalidArgument, "missing HTTP request"), nil
	}

	// Extract namespace from cookie
	namespaceID := s.GetCookieOrHeader(COOKIE_NAME, httpReq.GetHeaders())

	// If no namespace cookie or header, redirect to namespace selection
	if namespaceID == "" {
		// Check if this is a browser request or API request
		acceptHeader := httpReq.GetHeaders()["accept"]
		if strings.Contains(acceptHeader, "text/html") {
			// Browser-ish request - redirect to namespace selection page
			originalURL := fmt.Sprintf("%s://%s%s",
				httpReq.GetScheme(),
				httpReq.GetHost(),
				httpReq.GetPath())
			redirectURL := fmt.Sprintf("%s?redirect_to=%s",
				s.handler.PublicURL,
				url.QueryEscape(originalURL))

			return s.redirectResponse(redirectURL), nil
		} else {
			// API request - return 401 with WWW-Authenticate header
			return s.unauthorizedResponse("Missing namespace identifier. Provide namespace via 'x-namespace' header or 'namespace' cookie."), nil
		}
	}

	// Check if namespace exists in configuration
	s.handler.configLock.RLock()
	namespace, ok := s.handler.currentConfig.Namespaces[namespaceID]
	s.handler.configLock.RUnlock()

	if !ok {
		return s.denyResponse(codes.PermissionDenied, fmt.Sprintf("unauthorized namespace ID: %v", namespaceID)), nil
	}

	// Allow request and set x-namespace header
	return s.allowResponse(namespace.Target), nil
}

// GetCookieOrHeader extracts the namespace value from a cookie or header
func (s *AuthzGRPCServer) GetCookieOrHeader(name string, headers map[string]string) string {
	cookieHeader, ok := headers["cookie"]
	if ok {
		// Parse cookies to find namespace
		cookies := strings.Split(cookieHeader, ";")
		for _, cookie := range cookies {
			cookie = strings.TrimSpace(cookie)
			if after, ok := strings.CutPrefix(cookie, name+"="); ok {
				return after
			}
		}
	}

	if namespace, ok := headers["x-"+name]; ok {
		return namespace
	}

	return ""
}

// allowResponse creates a successful authorization response
func (s *AuthzGRPCServer) allowResponse(namespace string) *envoy_service_auth_v3.CheckResponse {
	return &envoy_service_auth_v3.CheckResponse{
		Status: &grpcstatus.Status{Code: int32(codes.OK)},
		HttpResponse: &envoy_service_auth_v3.CheckResponse_OkResponse{
			OkResponse: &envoy_service_auth_v3.OkHttpResponse{
				Headers: []*envoy_core_v3.HeaderValueOption{
					{
						Header: &envoy_core_v3.HeaderValue{
							Key:   "x-backend",
							Value: namespace,
						},
					},
				},
			},
		},
	}
}

// redirectResponse creates a redirect response
func (s *AuthzGRPCServer) redirectResponse(location string) *envoy_service_auth_v3.CheckResponse {
	return &envoy_service_auth_v3.CheckResponse{
		Status: &grpcstatus.Status{Code: int32(codes.Unauthenticated)},
		HttpResponse: &envoy_service_auth_v3.CheckResponse_DeniedResponse{
			DeniedResponse: &envoy_service_auth_v3.DeniedHttpResponse{
				Status: &envoy_type_v3.HttpStatus{Code: envoy_type_v3.StatusCode_Found},
				Headers: []*envoy_core_v3.HeaderValueOption{
					{
						Header: &envoy_core_v3.HeaderValue{
							Key:   "location",
							Value: location,
						},
					},
				},
			},
		},
	}
}

// denyResponse creates a denial response
func (s *AuthzGRPCServer) denyResponse(code codes.Code, message string) *envoy_service_auth_v3.CheckResponse {
	return &envoy_service_auth_v3.CheckResponse{
		Status: &grpcstatus.Status{
			Code:    int32(code),
			Message: message,
		},
		HttpResponse: &envoy_service_auth_v3.CheckResponse_DeniedResponse{
			DeniedResponse: &envoy_service_auth_v3.DeniedHttpResponse{
				Status: &envoy_type_v3.HttpStatus{Code: envoy_type_v3.StatusCode_Forbidden},
				Body:   message,
			},
		},
	}
}

// unauthorizedResponse creates a 401 response for API clients
func (s *AuthzGRPCServer) unauthorizedResponse(message string) *envoy_service_auth_v3.CheckResponse {
	return &envoy_service_auth_v3.CheckResponse{
		Status: &grpcstatus.Status{
			Code:    int32(codes.Unauthenticated),
			Message: message,
		},
		HttpResponse: &envoy_service_auth_v3.CheckResponse_DeniedResponse{
			DeniedResponse: &envoy_service_auth_v3.DeniedHttpResponse{
				Status: &envoy_type_v3.HttpStatus{Code: envoy_type_v3.StatusCode_Unauthorized},
				Headers: []*envoy_core_v3.HeaderValueOption{
					{
						Header: &envoy_core_v3.HeaderValue{
							Key:   "www-authenticate",
							Value: `Custom realm="namespace-required", error="missing_namespace", error_description="Provide namespace via x-namespace header or namespace cookie"`,
						},
					},
				},
				Body: message,
			},
		},
	}
}
