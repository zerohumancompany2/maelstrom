package datasources

import (
	"testing"
)

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
