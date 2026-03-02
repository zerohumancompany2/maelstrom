package datasources

import (
	"github.com/maelstrom/v3/pkg/datasource"
)

type DatasourceService interface {
	Get(name string) (datasource.DataSource, error)
	List() []string
	TagOnWrite(path string, taints []string) error
	GetTaints(path string) ([]string, error)
	ValidateAccess(path string, boundary datasource.BoundaryType) error
	Register(name string, ds datasource.DataSource) error
}

type datasourceService struct {
	datasources map[string]datasource.DataSource
}

func NewDatasourceService() DatasourceService {
	return &datasourceService{
		datasources: make(map[string]datasource.DataSource),
	}
}

func (s *datasourceService) Get(name string) (datasource.DataSource, error) {
	ds, ok := s.datasources[name]
	if !ok {
		return nil, nil
	}
	return ds, nil
}

func (s *datasourceService) List() []string {
	names := make([]string, 0, len(s.datasources))
	for name := range s.datasources {
		names = append(names, name)
	}
	return names
}

func (s *datasourceService) TagOnWrite(path string, taints []string) error {
	return nil
}

func (s *datasourceService) GetTaints(path string) ([]string, error) {
	return []string{}, nil
}

func (s *datasourceService) ValidateAccess(path string, boundary datasource.BoundaryType) error {
	return nil
}

func (s *datasourceService) Register(name string, ds datasource.DataSource) error {
	s.datasources[name] = ds
	return nil
}

type LocalDiskDatasource struct{}

func (d *LocalDiskDatasource) TagOnWrite(path string, taints []string) error {
	return nil
}

func (d *LocalDiskDatasource) GetTaints(path string) ([]string, error) {
	return nil, nil
}

func (d *LocalDiskDatasource) ValidateAccess(boundary datasource.BoundaryType) error {
	return nil
}

type S3Datasource struct{}

func (d *S3Datasource) TagOnWrite(path string, taints []string) error {
	return nil
}

func (d *S3Datasource) GetTaints(path string) ([]string, error) {
	return nil, nil
}

func (d *S3Datasource) ValidateAccess(boundary datasource.BoundaryType) error {
	return nil
}
