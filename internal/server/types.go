package server

import (
	"sync"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/michaelw/ext-authz-router/api"
)

var _ api.StrictServerInterface = (*AuthzHandler)(nil)

type AuthzConfig struct {
	Namespaces map[string]struct {
		Target      string `yaml:"target" json:"target"`
		Description string `yaml:"description,omitempty" json:"description,omitempty"`
	} `yaml:"namespaces" json:"namespaces"`
}

type AuthzHandler struct {
	PublicURL     string
	Swagger       *openapi3.T
	configLock    sync.RWMutex
	currentConfig AuthzConfig
	configPath    string
}
