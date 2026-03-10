package datasources

import (
	"fmt"

	"github.com/maelstrom/v3/pkg/datasource"
	"github.com/maelstrom/v3/pkg/mail"
	"github.com/maelstrom/v3/pkg/security"
)

type DatasourceService interface {
	ID() string
	Get(name string) (datasource.DataSource, error)
	List() []string
	TagOnWrite(path string, taints []string) error
	GetTaints(path string) ([]string, error)
	ValidateAccess(path string, boundary security.BoundaryType) error
	Register(name string, ds datasource.DataSource) error
	AttachTaintsOnRead(m *mail.Mail, path string) *mail.Mail
	HandleMail(mail mail.Mail) error
	Start() error
	Stop() error
}

type datasourceService struct {
	datasources map[string]datasource.DataSource
	taints      map[string][]string
}

func NewDatasourceService() DatasourceService {
	return &datasourceService{
		datasources: make(map[string]datasource.DataSource),
		taints:      make(map[string][]string),
	}
}

func (s *datasourceService) ID() string {
	return "sys:datasources"
}

func (s *datasourceService) Get(name string) (datasource.DataSource, error) {
	ds, ok := s.datasources[name]
	if !ok {
		return nil, fmt.Errorf("datasource %q not found", name)
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
	s.taints[path] = make([]string, len(taints))
	copy(s.taints[path], taints)
	return nil
}

func (s *datasourceService) GetTaints(path string) ([]string, error) {
	taints, ok := s.taints[path]
	if !ok {
		return []string{}, nil
	}
	result := make([]string, len(taints))
	copy(result, taints)
	return result, nil
}

func (s *datasourceService) ValidateAccess(path string, boundary security.BoundaryType) error {
	taints, _ := s.GetTaints(path)
	for _, taint := range taints {
		switch boundary {
		case security.InnerBoundary:
			return nil
		case security.DMZBoundary:
			if taint == "inner" {
				return fmt.Errorf("DMZ boundary cannot access inner taints")
			}
		case security.OuterBoundary:
			if taint == "inner" {
				return fmt.Errorf("Outer boundary cannot access inner taints")
			}
		}
	}
	return nil
}

func (s *datasourceService) Register(name string, ds datasource.DataSource) error {
	if _, exists := s.datasources[name]; exists {
		return fmt.Errorf("datasource %q already registered", name)
	}
	s.datasources[name] = ds
	return nil
}

func (s *datasourceService) AttachTaintsOnRead(m *mail.Mail, path string) *mail.Mail {
	taints, _ := s.GetTaints(path)
	existing := make(map[string]bool)
	for _, t := range m.Metadata.Taints {
		existing[t] = true
	}
	for _, t := range taints {
		if !existing[t] {
			m.Metadata.Taints = append(m.Metadata.Taints, t)
			existing[t] = true
		}
	}
	return m
}

func (s *datasourceService) HandleMail(m mail.Mail) error {
	return nil
}

func (s *datasourceService) Start() error {
	return nil
}

func (s *datasourceService) Stop() error {
	return nil
}

type LocalDiskDatasource struct{}

func (d *LocalDiskDatasource) TagOnWrite(path string, taints []string) error {
	return nil
}

func (d *LocalDiskDatasource) GetTaints(path string) ([]string, error) {
	return nil, nil
}

func (d *LocalDiskDatasource) ValidateAccess(boundary security.BoundaryType) error {
	return nil
}

type S3Datasource struct{}

func (d *S3Datasource) TagOnWrite(path string, taints []string) error {
	return nil
}

func (d *S3Datasource) GetTaints(path string) ([]string, error) {
	return nil, nil
}

func (d *S3Datasource) ValidateAccess(boundary security.BoundaryType) error {
	return nil
}
