package platform

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/maelstrom/v3/pkg/statechart"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestPlatformServiceYAML_Schema validates YAML schema against spec
// Required fields present (apiVersion, kind, metadata, spec)
// Optional fields handled correctly
func TestPlatformServiceYAML_Schema(t *testing.T) {
	t.Run("validates required fields", func(t *testing.T) {
		yamlContent := `
apiVersion: maelstrom.dev/v1
kind: PlatformService
metadata:
  name: sys:gateway
spec:
  chartRef: gateway-v1
`
		var ps PlatformService
		err := yaml.Unmarshal([]byte(yamlContent), &ps)
		require.NoError(t, err)
		assert.Equal(t, "maelstrom.dev/v1", ps.APIVersion)
		assert.Equal(t, "PlatformService", ps.Kind)
		assert.Equal(t, "sys:gateway", ps.Metadata.Name)
		assert.Equal(t, "gateway-v1", ps.Spec.ChartRef)
	})

	t.Run("validates required metadata name", func(t *testing.T) {
		yamlContent := `
apiVersion: maelstrom.dev/v1
kind: PlatformService
metadata:
spec:
  chartRef: gateway-v1
`
		var ps PlatformService
		err := yaml.Unmarshal([]byte(yamlContent), &ps)
		require.NoError(t, err)
		// Validate should catch missing required field
		err = ps.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "metadata.name is required")
	})

	t.Run("validates required spec chartRef", func(t *testing.T) {
		yamlContent := `
apiVersion: maelstrom.dev/v1
kind: PlatformService
metadata:
  name: sys:gateway
spec: {}
`
		var ps PlatformService
		err := yaml.Unmarshal([]byte(yamlContent), &ps)
		require.NoError(t, err)
		err = ps.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "spec.chartRef is required")
	})

	t.Run("handles optional fields with defaults", func(t *testing.T) {
		yamlContent := `
apiVersion: maelstrom.dev/v1
kind: PlatformService
metadata:
  name: sys:gateway
spec:
  chartRef: gateway-v1
`
		var ps PlatformService
		err := yaml.Unmarshal([]byte(yamlContent), &ps)
		require.NoError(t, err)
		// Optional fields should have defaults
		assert.False(t, ps.Metadata.Core)
		assert.False(t, ps.Spec.RequiredForKernelReady)
		assert.Equal(t, int32(1), ps.Spec.GetReplicas())
	})

	t.Run("handles all optional fields", func(t *testing.T) {
		yamlContent := `
apiVersion: maelstrom.dev/v1
kind: PlatformService
metadata:
  name: sys:gateway
  core: true
  displayName: Gateway Service
spec:
  chartRef: gateway-v1
  requiredForKernelReady: true
  replicas: 3
  persistence:
    enabled: true
    snapshotEvery: "100 messages"
  expose:
    http:
      port: 8080
      tls: true
    tcp:
      port: 9090
`
		var ps PlatformService
		err := yaml.Unmarshal([]byte(yamlContent), &ps)
		require.NoError(t, err)
		assert.True(t, ps.Metadata.Core)
		assert.Equal(t, "Gateway Service", ps.Metadata.DisplayName)
		assert.True(t, ps.Spec.RequiredForKernelReady)
		assert.Equal(t, int32(3), ps.Spec.GetReplicas())
		assert.True(t, ps.Spec.Persistence.Enabled)
		assert.Equal(t, "100 messages", ps.Spec.Persistence.SnapshotEvery)
		assert.True(t, ps.Spec.Expose.HTTP.TLS)
		assert.Equal(t, int32(8080), ps.Spec.Expose.HTTP.Port)
		assert.Equal(t, int32(9090), ps.Spec.Expose.TCP.Port)
	})
}

// TestPlatformServiceYAML_ChartRegistryLoad tests ChartRegistry loading PlatformService YAMLs
// ChartRegistry loads PlatformService YAMLs from charts/platform-services/
// YAML files parsed correctly
// ChartDefinitions created from YAML
func TestPlatformServiceYAML_ChartRegistryLoad(t *testing.T) {
	t.Run("loads platform service yaml from directory", func(t *testing.T) {
		// Create temp directory with test YAML files
		tmpDir := t.TempDir()

		yamlContent := `
apiVersion: maelstrom.dev/v1
kind: PlatformService
metadata:
  name: sys:gateway
  core: true
spec:
  chartRef: gateway-v1
  requiredForKernelReady: true
  replicas: 2
`
		err := os.WriteFile(filepath.Join(tmpDir, "gateway.yaml"), []byte(yamlContent), 0644)
		require.NoError(t, err)

		registry := NewChartRegistry(tmpDir)
		services, err := registry.LoadPlatformServices()
		require.NoError(t, err)
		assert.Len(t, services, 1)
		assert.Equal(t, "sys:gateway", services[0].Metadata.Name)
		assert.True(t, services[0].Metadata.Core)
	})

	t.Run("parses multiple yaml files", func(t *testing.T) {
		tmpDir := t.TempDir()

		files := map[string]string{
			"gateway.yaml": `
apiVersion: maelstrom.dev/v1
kind: PlatformService
metadata:
  name: sys:gateway
spec:
  chartRef: gateway-v1
`,
			"admin.yaml": `
apiVersion: maelstrom.dev/v1
kind: PlatformService
metadata:
  name: sys:admin
spec:
  chartRef: admin-v1
`,
		}

		for name, content := range files {
			err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644)
			require.NoError(t, err)
		}

		registry := NewChartRegistry(tmpDir)
		services, err := registry.LoadPlatformServices()
		require.NoError(t, err)
		assert.Len(t, services, 2)
	})

	t.Run("creates chart definitions from yaml", func(t *testing.T) {
		tmpDir := t.TempDir()

		yamlContent := `
apiVersion: maelstrom.dev/v1
kind: PlatformService
metadata:
  name: sys:gateway
  core: true
spec:
  chartRef: gateway-v1
  requiredForKernelReady: true
`
		err := os.WriteFile(filepath.Join(tmpDir, "gateway.yaml"), []byte(yamlContent), 0644)
		require.NoError(t, err)

		registry := NewChartRegistry(tmpDir)
		services, err := registry.LoadPlatformServices()
		require.NoError(t, err)

		// Convert to ChartDefinition
		hydrator := statechart.DefaultHydrator()
		def, err := services[0].ToChartDefinition(hydrator)
		require.NoError(t, err)
		assert.NotEmpty(t, def.ID)
		assert.NotNil(t, def.Root)
	})

	t.Run("returns error for invalid yaml", func(t *testing.T) {
		tmpDir := t.TempDir()

		yamlContent := `
apiVersion: maelstrom.dev/v1
kind: PlatformService
metadata:
  name: sys:gateway
spec:
  chartRef: gateway-v1
invalid yaml content: [
`
		err := os.WriteFile(filepath.Join(tmpDir, "invalid.yaml"), []byte(yamlContent), 0644)
		require.NoError(t, err)

		registry := NewChartRegistry(tmpDir)
		_, err = registry.LoadPlatformServices()
		assert.Error(t, err)
	})
}
