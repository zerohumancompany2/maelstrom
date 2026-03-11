package gateway

import (
	"slices"
	"testing"
)

func TestGatewaySecurity_AllDataTaintedAtBorder(t *testing.T) {
	tainter := &BorderTainter{
		DefaultTaints: []string{"USER_SUPPLIED", "OUTER_BOUNDARY"},
	}

	// All data entering the runtime is tainted at the border (arch-v1.md L276)
	testData := map[string]any{
		"message": "Hello, agent!",
		"source":  "user:web",
	}

	taintedData, err := tainter.TaintInboundData(testData, "user:web")
	if err != nil {
		t.Fatalf("Expected no error tainting inbound data, got %v", err)
	}

	// Verify data is tainted (arch-v1.md L276)
	taintedMap, ok := taintedData.(map[string]any)
	if !ok {
		t.Fatal("Expected tainted data to be map")
	}

	if taintedMap["taints"] == nil {
		t.Error("Expected taints to be attached to data")
	}

	// Verify taints are attached as soon as data is ingested (arch-v1.md L276)
	taints := taintedMap["taints"].([]string)
	if !slices.Contains(taints, "USER_SUPPLIED") {
		t.Error("Expected USER_SUPPLIED taint")
	}

	if !slices.Contains(taints, "OUTER_BOUNDARY") {
		t.Error("Expected OUTER_BOUNDARY taint")
	}

	// No untainted information exists inside the runtime (arch-v1.md L276)
	if len(taints) == 0 {
		t.Error("Expected all data to have at least one taint")
	}

	// Test with different source types
	testCases := []struct {
		source        string
		expectedTaint string
	}{
		{"user:web", "USER_SUPPLIED"},
		{"user:api", "USER_SUPPLIED"},
		{"system:heartbeat", "SYSTEM_GENERATED"},
		{"tool:registry", "TOOL_OUTPUT"},
	}

	for _, tc := range testCases {
		taintedData, err := tainter.TaintInboundData(map[string]any{"data": "test"}, tc.source)
		if err != nil {
			t.Fatalf("Expected no error for source %s, got %v", tc.source, err)
		}

		taintedMap := taintedData.(map[string]any)
		taints := taintedMap["taints"].([]string)

		if !slices.Contains(taints, tc.expectedTaint) {
			t.Errorf("Expected taint %s for source %s", tc.expectedTaint, tc.source)
		}
	}
}
