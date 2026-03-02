package datasources

import (
	"testing"
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
