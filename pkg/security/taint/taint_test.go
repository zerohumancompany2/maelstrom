package taint

import (
	"sync"
	"testing"
)

func TestTaintEngine_Attach_UserSource(t *testing.T) {
	engine := NewTaintEngine()
	data := map[string]any{"key": "value"}
	result, taints := engine.Attach(data, "user")

	if len(taints) != 1 {
		t.Fatalf("expected 1 taint, got %d", len(taints))
	}
	if taints[0] != string(TaintUserSupplied) {
		t.Errorf("expected USER_SUPPLIED taint, got %s", taints[0])
	}

	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatal("expected result to be map[string]any")
	}
	if _, hasTaints := resultMap["_taints"]; !hasTaints {
		t.Error("expected _taints field in result")
	}
}

func TestTaintEngine_Attach_ExternalSource(t *testing.T) {
	engine := NewTaintEngine()
	data := map[string]any{"key": "value"}
	_, taints := engine.Attach(data, "external_api")

	if len(taints) != 1 {
		t.Fatalf("expected 1 taint, got %d", len(taints))
	}
	if taints[0] != string(TaintExternal) {
		t.Errorf("expected EXTERNAL taint, got %s", taints[0])
	}
}

func TestTaintEngine_Attach_InnerSource(t *testing.T) {
	engine := NewTaintEngine()
	data := map[string]any{"key": "value"}
	_, taints := engine.Attach(data, "inner_db")

	if len(taints) != 1 {
		t.Fatalf("expected 1 taint, got %d", len(taints))
	}
	if taints[0] != string(TaintInnerOnly) {
		t.Errorf("expected INNER_ONLY taint, got %s", taints[0])
	}
}

func TestTaintEngine_Attach_SecretSource(t *testing.T) {
	engine := NewTaintEngine()
	data := map[string]any{"key": "value"}
	_, taints := engine.Attach(data, "secrets_store")

	if len(taints) != 1 {
		t.Fatalf("expected 1 taint, got %d", len(taints))
	}
	if taints[0] != string(TaintSecret) {
		t.Errorf("expected SECRET taint, got %s", taints[0])
	}
}

func TestTaintEngine_Attach_PII_Source(t *testing.T) {
	engine := NewTaintEngine()
	data := map[string]any{"key": "value"}
	_, taints := engine.Attach(data, "pii_store")

	if len(taints) != 1 {
		t.Fatalf("expected 1 taint, got %d", len(taints))
	}
	if taints[0] != string(TaintPII) {
		t.Errorf("expected PII taint, got %s", taints[0])
	}
}

func TestTaintEngine_Attach_NilData(t *testing.T) {
	engine := NewTaintEngine()
	result, taints := engine.Attach(nil, "user")

	if result != nil {
		t.Error("expected nil result for nil data")
	}
	if len(taints) != 1 {
		t.Errorf("expected 1 taint, got %d", len(taints))
	}
}

func TestTaintEngine_Propagate_MapToMap(t *testing.T) {
	engine := NewTaintEngine()
	src := map[string]any{
		"_taints": []string{"USER_SUPPLIED"},
		"data":    "source",
	}
	dest := map[string]any{"data": "destination"}

	result := engine.Propagate(src, dest)
	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatal("expected result to be map[string]any")
	}

	taints, hasTaints := resultMap["_taints"].([]string)
	if !hasTaints {
		t.Error("expected _taints field in result")
	}
	if len(taints) != 1 || taints[0] != "USER_SUPPLIED" {
		t.Errorf("expected USER_SUPPLIED taint, got %v", taints)
	}
}

func TestTaintEngine_Propagate_NoTaints(t *testing.T) {
	engine := NewTaintEngine()
	src := map[string]any{"data": "source"}
	dest := map[string]any{"data": "destination"}

	result := engine.Propagate(src, dest)
	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatal("expected result to be map[string]any")
	}

	if _, hasTaints := resultMap["_taints"]; hasTaints {
		t.Error("expected no _taints field when source has no taints")
	}
}

func TestTaintEngine_Propagate_NilSource(t *testing.T) {
	engine := NewTaintEngine()
	dest := map[string]any{"data": "destination"}

	result := engine.Propagate(nil, dest)
	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatal("expected result to be map[string]any")
	}

	if _, hasTaints := resultMap["_taints"]; hasTaints {
		t.Error("expected no _taints field when source is nil")
	}
}

func TestTaintEngine_Propagate_NilDest(t *testing.T) {
	engine := NewTaintEngine()
	src := map[string]any{"_taints": []string{"USER_SUPPLIED"}}

	result := engine.Propagate(src, nil)
	if result != nil {
		t.Error("expected nil result for nil dest")
	}
}

func TestTaintEngine_Strip_SingleTaint(t *testing.T) {
	engine := NewTaintEngine()
	data := map[string]any{
		"_taints": []string{"USER_SUPPLIED", "EXTERNAL"},
		"data":    "value",
	}

	result := engine.Strip(data, []string{"USER_SUPPLIED"})
	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatal("expected result to be map[string]any")
	}

	taints, hasTaints := resultMap["_taints"].([]string)
	if !hasTaints {
		t.Error("expected _taints field in result")
	}
	if len(taints) != 1 || taints[0] != "EXTERNAL" {
		t.Errorf("expected EXTERNAL taint only, got %v", taints)
	}
}

func TestTaintEngine_Strip_AllTaints(t *testing.T) {
	engine := NewTaintEngine()
	data := map[string]any{
		"_taints": []string{"USER_SUPPLIED"},
		"data":    "value",
	}

	result := engine.Strip(data, []string{"USER_SUPPLIED"})
	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatal("expected result to be map[string]any")
	}

	if _, hasTaints := resultMap["_taints"]; hasTaints {
		t.Error("expected no _taints field when all taints are stripped")
	}
}

func TestTaintEngine_Strip_NonExistentTaint(t *testing.T) {
	engine := NewTaintEngine()
	data := map[string]any{
		"_taints": []string{"USER_SUPPLIED"},
		"data":    "value",
	}

	result := engine.Strip(data, []string{"NONEXISTENT"})
	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatal("expected result to be map[string]any")
	}

	taints, hasTaints := resultMap["_taints"].([]string)
	if !hasTaints {
		t.Error("expected _taints field in result")
	}
	if len(taints) != 1 || taints[0] != "USER_SUPPLIED" {
		t.Errorf("expected USER_SUPPLIED taint unchanged, got %v", taints)
	}
}

func TestTaintEngine_Strip_NilData(t *testing.T) {
	engine := NewTaintEngine()
	result := engine.Strip(nil, []string{"USER_SUPPLIED"})

	if result != nil {
		t.Error("expected nil result for nil data")
	}
}

func TestTaintEngine_Attach_NestedMap(t *testing.T) {
	engine := NewTaintEngine()
	data := map[string]any{
		"nested": map[string]any{
			"key": "value",
		},
	}

	result, _ := engine.Attach(data, "user")
	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatal("expected result to be map[string]any")
	}

	if _, hasTaints := resultMap["_taints"]; !hasTaints {
		t.Error("expected _taints field in root result")
	}
}

func TestTaintEngine_Attach_MergeExistingTaints(t *testing.T) {
	engine := NewTaintEngine()
	data := map[string]any{
		"_taints": []string{"EXTERNAL"},
		"data":    "value",
	}

	result, _ := engine.Attach(data, "user")
	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatal("expected result to be map[string]any")
	}

	taints, ok := resultMap["_taints"].([]string)
	if !ok {
		t.Fatal("expected _taints to be []string")
	}
	if len(taints) != 2 {
		t.Errorf("expected 2 taints, got %d", len(taints))
	}
}

func TestTaintEngine_ConcurrentAttach(t *testing.T) {
	engine := NewTaintEngine()
	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			data := map[string]any{"key": "value"}
			engine.Attach(data, "user")
		}()
	}

	wg.Wait()
}

func TestTaintEngine_TaintSet_Operations(t *testing.T) {
	ts := make(TaintSet)
	ts.Add("USER_SUPPLIED")
	ts.Add("EXTERNAL")

	if !ts.Has("USER_SUPPLIED") {
		t.Error("expected USER_SUPPLIED in taint set")
	}
	if !ts.Has("EXTERNAL") {
		t.Error("expected EXTERNAL in taint set")
	}
	if ts.Has("PII") {
		t.Error("expected PII not in taint set")
	}

	slice := ts.ToSlice()
	if len(slice) != 2 {
		t.Errorf("expected 2 taints in slice, got %d", len(slice))
	}
}

func TestTaintEngine_Propagate_NestedMap(t *testing.T) {
	engine := NewTaintEngine()
	src := map[string]any{
		"_taints": []string{"SECRET"},
		"nested": map[string]any{
			"data": "source",
		},
	}
	dest := map[string]any{
		"nested": map[string]any{
			"data": "destination",
		},
	}

	result := engine.Propagate(src, dest)
	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatal("expected result to be map[string]any")
	}

	taints, hasTaints := resultMap["_taints"].([]string)
	if !hasTaints {
		t.Error("expected _taints field in result")
	}
	if len(taints) != 1 || taints[0] != "SECRET" {
		t.Errorf("expected SECRET taint, got %v", taints)
	}
}
