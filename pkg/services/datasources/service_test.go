package datasources

import (
	"testing"

	"github.com/maelstrom/v3/pkg/security"
)

// TestDataSourceService_ID - arch-v1.md L473: sys:datasources service ID
func TestDataSourceService_ID(t *testing.T) {
	svc := NewDatasourceService()

	id := svc.ID()
	if id != "sys:datasources" {
		t.Errorf("Expected ID 'sys:datasources', got '%s'", id)
	}
}

// TestDataSourceService_Register - arch-v1.md L473: register data sources with duplicate detection
func TestDataSourceService_Register(t *testing.T) {
	svc := NewDatasourceService()

	// Register first datasource
	err := svc.Register("localDisk", &LocalDiskDatasource{})
	if err != nil {
		t.Fatalf("Register localDisk failed: %v", err)
	}

	// Register duplicate should fail
	err = svc.Register("localDisk", &S3Datasource{})
	if err == nil {
		t.Error("Expected error for duplicate registration, got nil")
	}
}

// TestDataSourceService_Get - arch-v1.md L490: get data source by name
func TestDataSourceService_Get(t *testing.T) {
	svc := NewDatasourceService()

	// Register a datasource
	err := svc.Register("localDisk", &LocalDiskDatasource{})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Get existing datasource
	ds, err := svc.Get("localDisk")
	if err != nil {
		t.Fatalf("Get existing datasource failed: %v", err)
	}
	if ds == nil {
		t.Error("Expected non-nil datasource")
	}

	// Get unknown datasource should return error
	_, err = svc.Get("unknown")
	if err == nil {
		t.Error("Expected error for unknown datasource, got nil")
	}
}

func TestDatasources_Register(t *testing.T) {
	svc := NewDatasourceService()

	err := svc.Register("localDisk", &LocalDiskDatasource{})
	if err != nil {
		t.Fatalf("Register localDisk failed: %v", err)
	}

	err = svc.Register("s3", &S3Datasource{})
	if err != nil {
		t.Fatalf("Register s3 failed: %v", err)
	}

	if len(svc.List()) != 2 {
		t.Errorf("Expected 2 datasources, got %d", len(svc.List()))
	}
}

func TestDatasources_Get(t *testing.T) {
	svc := NewDatasourceService()

	svc.Register("localDisk", &LocalDiskDatasource{})

	ds, err := svc.Get("localDisk")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if ds == nil {
		t.Fatal("Expected non-nil datasource")
	}
}

func TestDatasources_TagOnWrite(t *testing.T) {
	svc := NewDatasourceService()

	svc.Register("localDisk", &LocalDiskDatasource{})

	err := svc.TagOnWrite("/path/to/file.txt", []string{"confidential", "internal"})
	if err != nil {
		t.Fatalf("TagOnWrite failed: %v", err)
	}
}

func TestDatasources_GetTaints(t *testing.T) {
	svc := NewDatasourceService()

	svc.Register("localDisk", &LocalDiskDatasource{})

	taints, err := svc.GetTaints("/path/to/file.txt")
	if err != nil {
		t.Fatalf("GetTaints failed: %v", err)
	}

	if taints == nil {
		t.Fatal("Expected non-nil taints")
	}
}

func TestDatasources_ValidateAccess(t *testing.T) {
	svc := NewDatasourceService()

	svc.Register("localDisk", &LocalDiskDatasource{})

	err := svc.ValidateAccess("/path/to/file.txt", security.InnerBoundary)
	if err != nil {
		t.Fatalf("ValidateAccess failed: %v", err)
	}

	err = svc.ValidateAccess("/path/to/file.txt", security.OuterBoundary)
	if err != nil {
		t.Fatalf("ValidateAccess for outer failed: %v", err)
	}
}
