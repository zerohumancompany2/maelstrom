package datasource

import (
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

	key := "documents/file.txt"
	taints, err := ds.GetTaints(key)
	if err != nil {
		t.Errorf("GetTaints failed: %v", err)
	}

	if taints == nil {
		t.Error("GetTaints should return non-nil slice")
	}
}
