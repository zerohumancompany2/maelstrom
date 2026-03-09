package datasource

import (
	"encoding/json"
	"fmt"
	"github.com/maelstrom/v3/pkg/security"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

type localDisk struct {
	path               string
	xattrNamespace     string
	allowedForBoundary []security.BoundaryType
	taintMode          string
	alwaysTaintAs      []string
}

func NewLocalDisk(config map[string]any) (DataSource, error) {
	path := config["path"].(string)
	xattrNS := "user.maelstrom"
	if ns, ok := config["xattrNamespace"].(string); ok {
		xattrNS = ns
	}

	allowedForBoundary := []security.BoundaryType{}
	if allowed, ok := config["allowedForBoundary"].([]security.BoundaryType); ok {
		allowedForBoundary = allowed
	}

	taintMode := "inheritFromXattr"
	if tm, ok := config["taintMode"].(string); ok {
		taintMode = tm
	}

	alwaysTaintAs := []string{}
	if taintMode != "" && strings.HasPrefix(taintMode, "alwaysTaintAs=") {
		taintValue := strings.TrimPrefix(taintMode, "alwaysTaintAs=")
		alwaysTaintAs = strings.Split(taintValue, ",")
		for i := range alwaysTaintAs {
			alwaysTaintAs[i] = strings.TrimSpace(alwaysTaintAs[i])
		}
	}

	return &localDisk{
		path:               path,
		xattrNamespace:     xattrNS,
		allowedForBoundary: allowedForBoundary,
		taintMode:          taintMode,
		alwaysTaintAs:      alwaysTaintAs,
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
	if d.taintMode == "none" {
		return []string{}, nil
	}

	if d.taintMode != "" && strings.HasPrefix(d.taintMode, "alwaysTaintAs=") {
		return d.alwaysTaintAs, nil
	}

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
	if len(d.allowedForBoundary) == 0 {
		return nil
	}
	for _, allowed := range d.allowedForBoundary {
		if allowed == boundary {
			return nil
		}
	}
	return fmt.Errorf("boundary %s not allowed for this datasource", boundary)
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
