package server

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

func WithConfig(path string) HandlerOption {
	return func(h *AuthzHandler) {
		h.configPath = path
	}
}

// loadConfig loads the namespace configuration from the YAML file
func (h *AuthzHandler) loadConfig() error {
	file, err := os.ReadFile(h.configPath)
	if err != nil {
		return err
	}
	var cfg AuthzConfig
	if err := yaml.Unmarshal(file, &cfg); err != nil {
		return err
	}
	h.configLock.Lock()
	defer h.configLock.Unlock()
	h.currentConfig = cfg
	log.Println("[config] reloaded")
	return nil
}
