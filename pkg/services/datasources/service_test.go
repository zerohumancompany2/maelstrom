package datasources

import (
	"testing"
)

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
