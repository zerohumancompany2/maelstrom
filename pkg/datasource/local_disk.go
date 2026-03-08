package datasource

import (
	"github.com/maelstrom/v3/pkg/security"
	"os"
	"path/filepath"
)

type localDisk struct {
	path           string
	xattrNamespace string
}

func NewLocalDisk(config map[string]any) (DataSource, error) {
	path := config["path"].(string)
	xattrNS := "user.maelstrom"
	if ns, ok := config["xattrNamespace"].(string); ok {
		xattrNS = ns
	}

	return &localDisk{
		path:           path,
		xattrNamespace: xattrNS,
	}, nil
}

func (d *localDisk) TagOnWrite(path string, taints []string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		return err
	}

	return nil
}

func (d *localDisk) GetTaints(path string) ([]string, error) {
	return []string{}, nil
}

func (d *localDisk) ValidateAccess(boundary security.BoundaryType) error {
	return nil
}

func init() {
	Register("localDisk", NewLocalDisk)
}
