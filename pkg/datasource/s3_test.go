package datasource

import (
	"strings"
	"testing"

	"github.com/maelstrom/v3/pkg/security"
)

func TestS3DataSource_Config(t *testing.T) {
	config := map[string]any{
		"bucket":   "my-bucket",
		"region":   "us-east-1",
		"endpoint": "https://s3.us-east-1.amazonaws.com",
	}

	ds, err := NewS3DataSource(config)
	if err != nil {
		t.Fatalf("NewS3DataSource failed: %v", err)
	}

	if ds == nil {
		t.Fatal("DataSource should not be nil")
	}

	s3ds, ok := ds.(*s3DataSource)
	if !ok {
		t.Fatal("DataSource should be *s3DataSource")
	}

	if s3ds.bucket != "my-bucket" {
		t.Errorf("bucket mismatch: got %q, want %q", s3ds.bucket, "my-bucket")
	}

	if s3ds.region != "us-east-1" {
		t.Errorf("region mismatch: got %q, want %q", s3ds.region, "us-east-1")
	}
}

func TestS3DataSource_ValidateAccess_Allowed(t *testing.T) {
	config := map[string]any{
		"bucket":             "my-bucket",
		"region":             "us-east-1",
		"allowedForBoundary": []security.BoundaryType{security.InnerBoundary, security.DMZBoundary},
	}

	ds, err := NewS3DataSource(config)
	if err != nil {
		t.Fatalf("NewS3DataSource failed: %v", err)
	}

	err = ds.ValidateAccess(security.DMZBoundary)
	if err != nil {
		t.Errorf("ValidateAccess should allow dmz boundary, got error: %v", err)
	}

	err = ds.ValidateAccess(security.InnerBoundary)
	if err != nil {
		t.Errorf("ValidateAccess should allow inner boundary, got error: %v", err)
	}
}

func TestS3DataSource_ValidateAccess_Denied(t *testing.T) {
	config := map[string]any{
		"bucket":             "my-bucket",
		"region":             "us-east-1",
		"allowedForBoundary": []security.BoundaryType{security.InnerBoundary},
	}

	ds, err := NewS3DataSource(config)
	if err != nil {
		t.Fatalf("NewS3DataSource failed: %v", err)
	}

	err = ds.ValidateAccess(security.OuterBoundary)
	if err == nil {
		t.Error("ValidateAccess should deny outer boundary, got no error")
	} else if !strings.Contains(err.Error(), "outer") {
		t.Errorf("error message should contain boundary info, got: %v", err)
	}
}

func TestS3DataSource_ValidateAccess_NoRestriction(t *testing.T) {
	config := map[string]any{
		"bucket": "my-bucket",
		"region": "us-east-1",
	}

	ds, err := NewS3DataSource(config)
	if err != nil {
		t.Fatalf("NewS3DataSource failed: %v", err)
	}

	err = ds.ValidateAccess(security.OuterBoundary)
	if err != nil {
		t.Errorf("ValidateAccess should allow all boundaries when no restriction, got error: %v", err)
	}
}

func TestS3DataSource_TagOnWrite(t *testing.T) {
	config := map[string]any{
		"bucket": "my-bucket",
		"region": "us-east-1",
	}

	ds, err := NewS3DataSource(config)
	if err != nil {
		t.Fatalf("NewS3DataSource failed: %v", err)
	}

	key := "documents/secret.txt"
	taints := []string{"SECRET", "INNER_ONLY"}

	err = ds.TagOnWrite(key, taints)
	if err != nil {
		t.Errorf("TagOnWrite failed: %v", err)
	}

	s3ds, ok := ds.(*s3DataSource)
	if !ok {
		t.Fatal("DataSource should be *s3DataSource")
	}

	storedTaints, ok := s3ds.tags[key]
	if !ok {
		t.Error("Tags should be stored for key")
	}

	if len(storedTaints) != len(taints) {
		t.Errorf("taint length mismatch: got %d, want %d", len(storedTaints), len(taints))
	}

	for i, expected := range taints {
		if storedTaints[i] != expected {
			t.Errorf("taint[%d] mismatch: got %q, want %q", i, storedTaints[i], expected)
		}
	}
}

func TestS3DataSource_GetTaints(t *testing.T) {
	config := map[string]any{
		"bucket": "my-bucket",
		"region": "us-east-1",
	}

	ds, err := NewS3DataSource(config)
	if err != nil {
		t.Fatalf("NewS3DataSource failed: %v", err)
	}

	s3ds, ok := ds.(*s3DataSource)
	if !ok {
		t.Fatal("DataSource should be *s3DataSource")
	}

	expectedTaints := []string{"PII", "TOOL_OUTPUT"}
	s3ds.tags["documents/file.txt"] = expectedTaints

	key := "documents/file.txt"
	taints, err := ds.GetTaints(key)
	if err != nil {
		t.Errorf("GetTaints failed: %v", err)
	}

	if len(taints) != len(expectedTaints) {
		t.Errorf("taint length mismatch: got %d, want %d", len(taints), len(expectedTaints))
	}

	for i, expected := range expectedTaints {
		if taints[i] != expected {
			t.Errorf("taint[%d] mismatch: got %q, want %q", i, taints[i], expected)
		}
	}
}

func TestS3DataSource_GetTaints_Empty(t *testing.T) {
	config := map[string]any{
		"bucket": "my-bucket",
		"region": "us-east-1",
	}

	ds, err := NewS3DataSource(config)
	if err != nil {
		t.Fatalf("NewS3DataSource failed: %v", err)
	}

	key := "nonexistent/file.txt"
	taints, err := ds.GetTaints(key)
	if err != nil {
		t.Errorf("GetTaints should not return error for missing key, got: %v", err)
	}

	if len(taints) != 0 {
		t.Errorf("GetTaints should return empty slice for missing key, got %d items", len(taints))
	}
}
