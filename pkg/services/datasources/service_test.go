package datasources

import (
	"testing"
)

func TestDatasources_TagOnWrite(t *testing.T) {
	svc := NewDatasourceService()

	svc.Register("localDisk", &LocalDiskDatasource{})

	err := svc.TagOnWrite("/path/to/file.txt", []string{"confidential", "internal"})
	if err != nil {
		t.Fatalf("TagOnWrite failed: %v", err)
	}
}
