package server

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "ext-authz-router-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test with valid YAML config
	t.Run("valid YAML config", func(t *testing.T) {
		yamlConfig := `
namespaces:
  awesome-penguin:
    target: red
  cool-otter:
    target: blue
  test-namespace:
    target: green
`
		configPath := filepath.Join(tempDir, "config.yaml")
		err := os.WriteFile(configPath, []byte(yamlConfig), 0644)
		if err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}

		// Create handler with test config path
		swagger := &openapi3.T{}
		handler := NewServerHandler("http://test", swagger, WithConfig(configPath))

		// Check if config was loaded correctly
		handler.configLock.RLock()
		namespaces := handler.currentConfig.Namespaces
		handler.configLock.RUnlock()

		expectedNamespaces := map[string]string{
			"awesome-penguin": "red",
			"cool-otter":      "blue",
			"test-namespace":  "green",
		}

		if len(namespaces) != len(expectedNamespaces) {
			t.Errorf("Expected %d namespaces, got %d", len(expectedNamespaces), len(namespaces))
		}

		for k, v := range expectedNamespaces {
			if namespaces[k].Target != v {
				t.Errorf("Expected namespace %s to have value %s, got %s", k, v, namespaces[k].Target)
			}
		}
	})

	// Test with current config file format (should now work with our fix)
	t.Run("current config format (fixed)", func(t *testing.T) {
		yamlConfig := `namespaces:
  awesome-penguin:
    target: red
  cool-otter:
    target: blue`

		configPath := filepath.Join(tempDir, "config.yaml")
		err := os.WriteFile(configPath, []byte(yamlConfig), 0644)
		if err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}

		// Create handler with test config path
		swagger := &openapi3.T{}
		handler := &AuthzHandler{
			PublicURL:  "http://test",
			Swagger:    swagger,
			configPath: configPath,
		}

		// This should now succeed with our YAML fix
		err = handler.loadConfig()
		if err != nil {
			t.Errorf("Expected loadConfig to succeed with YAML format, but it failed: %v", err)
		}

		// Verify the config was loaded correctly
		expectedNamespaces := map[string]string{
			"awesome-penguin": "red",
			"cool-otter":      "blue",
		}

		if len(handler.currentConfig.Namespaces) != len(expectedNamespaces) {
			t.Errorf("Expected %d namespaces, got %d", len(expectedNamespaces), len(handler.currentConfig.Namespaces))
		}

		for k, v := range expectedNamespaces {
			if handler.currentConfig.Namespaces[k].Target != v {
				t.Errorf("Expected namespace %s to have value %s, got %s", k, v, handler.currentConfig.Namespaces[k].Target)
			}
		}
	})

	// Test with invalid config file
	t.Run("invalid config file", func(t *testing.T) {
		invalidConfig := `invalid yaml content: [[[`

		configPath := filepath.Join(tempDir, "invalid.yaml")
		err := os.WriteFile(configPath, []byte(invalidConfig), 0644)
		if err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}

		handler := &AuthzHandler{
			configPath: configPath,
		}

		err = handler.loadConfig()
		if err == nil {
			t.Error("Expected loadConfig to fail with invalid YAML, but it succeeded")
		}
	})

	// Test with missing config file
	t.Run("missing config file", func(t *testing.T) {
		nonExistentPath := filepath.Join(tempDir, "does-not-exist.yaml")

		handler := &AuthzHandler{
			configPath: nonExistentPath,
		}

		err = handler.loadConfig()
		if err == nil {
			t.Error("Expected loadConfig to fail with missing file, but it succeeded")
		}
	})
}
