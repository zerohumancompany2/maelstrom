package datasources

import (
	"testing"

	"github.com/maelstrom/v3/pkg/datasource"
)

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

	err := svc.ValidateAccess("/path/to/file.txt", datasource.InnerBoundary)
	if err != nil {
		t.Fatalf("ValidateAccess failed: %v", err)
	}

	err = svc.ValidateAccess("/path/to/file.txt", datasource.OuterBoundary)
	if err != nil {
		t.Fatalf("ValidateAccess for outer failed: %v", err)
	}
}
