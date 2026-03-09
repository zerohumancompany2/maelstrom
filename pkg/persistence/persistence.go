package persistence

import (
	"github.com/maelstrom/v3/pkg/datasource"
	"github.com/maelstrom/v3/pkg/security"
)

type Persistence struct {
	taintPolicy *security.TaintPolicy
	dataSource  datasource.DataSource
}

func NewPersistence(taintPolicy *security.TaintPolicy, dataSource datasource.DataSource) *Persistence {
	return &Persistence{
		taintPolicy: taintPolicy,
		dataSource:  dataSource,
	}
}

func (p *Persistence) Write(data any, taints []string) error {
	panic("not implemented")
}

func (p *Persistence) Read(key string) (any, []string, error) {
	panic("not implemented")
}

func (p *Persistence) ValidateTaintPolicy(taints []string, boundary security.BoundaryType) error {
	panic("not implemented")
}
