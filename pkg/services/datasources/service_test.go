package datasources

import (
	"testing"

	"github.com/maelstrom/v3/pkg/datasource"
)

func TestDatasources_ValidateAccess(t *testing.T) {
	svc := NewDatasourceService()

	svc.Register("localDisk", &LocalDiskDatasource{})

	err := svc.ValidateAccess("/path/to/file.txt", "inner")
	if err != nil {
		t.Fatalf("ValidateAccess failed: %v", err)
	}

	err = svc.ValidateAccess("/path/to/file.txt", datasource.OuterBoundary)
	if err != nil {
		t.Fatalf("ValidateAccess for outer failed: %v", err)
	}
}
