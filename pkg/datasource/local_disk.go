package datasource

import (
	"encoding/json"
	"github.com/maelstrom/v3/pkg/security"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
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

	jsonData, err := json.Marshal(taints)
	if err != nil {
		return err
	}

	attrName := d.xattrNamespace + ".taints"
	if err := unix.Lsetxattr(path, attrName, jsonData, 0); err != nil {
		return d.writeSidecar(path, taints)
	}

	return nil
}

func (d *localDisk) GetTaints(path string) ([]string, error) {
	attrName := d.xattrNamespace + ".taints"
	dest := make([]byte, 4096)
	n, err := unix.Lgetxattr(path, attrName, dest)
	if err != nil {
		sidecarPath := path + ".maelstrom"
		jsonData, err := os.ReadFile(sidecarPath)
		if err != nil {
			return []string{}, nil
		}
		var taints []string
		if err := json.Unmarshal(jsonData, &taints); err != nil {
			return []string{}, nil
		}
		return taints, nil
	}

	var taints []string
	if err := json.Unmarshal(dest[:n], &taints); err != nil {
		return []string{}, err
	}

	return taints, nil
}

func (d *localDisk) ValidateAccess(boundary security.BoundaryType) error {
	return nil
}

func (d *localDisk) writeSidecar(path string, taints []string) error {
	jsonData, err := json.Marshal(taints)
	if err != nil {
		return err
	}
	return os.WriteFile(path+".maelstrom", jsonData, 0644)
}

func init() {
	Register("localDisk", NewLocalDisk)
}
