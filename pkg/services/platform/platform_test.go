package platform

import (
	"testing"

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
