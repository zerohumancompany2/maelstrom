package persistence

import (
	"fmt"

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
	if err := p.ValidateTaintPolicy(taints, security.OuterBoundary); err != nil {
		return err
	}

	switch v := data.(type) {
	case map[string]interface{}:
		for k := range v {
			if err := p.dataSource.TagOnWrite(k, taints); err != nil {
				return err
			}
		}
	default:
		if err := p.dataSource.TagOnWrite("default", taints); err != nil {
			return err
		}
	}

	return nil
}

func (p *Persistence) Read(key string) (any, []string, error) {
	panic("not implemented")
}

func (p *Persistence) ValidateTaintPolicy(taints []string, boundary security.BoundaryType) error {
	for _, taint := range taints {
		if taint == "INNER_ONLY" && (boundary == security.DMZBoundary || boundary == security.OuterBoundary) {
			return fmt.Errorf("taint %s is forbidden on boundary %s", taint, boundary)
		}
		if taint == "SECRET" && (boundary == security.DMZBoundary || boundary == security.OuterBoundary) {
			return fmt.Errorf("taint %s is forbidden on boundary %s", taint, boundary)
		}
		if taint == "PII" && boundary == security.OuterBoundary {
			return fmt.Errorf("taint %s is forbidden on boundary %s", taint, boundary)
		}
	}
	return nil
}
