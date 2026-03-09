package persistence

import (
	"github.com/maelstrom/v3/pkg/security"
	"testing"
)

func TestPersistence_RefusesBoundaryViolation(t *testing.T) {
	p := &Persistence{
		taintPolicy: &security.TaintPolicy{},
		dataSource:  nil,
	}

	err := p.ValidateTaintPolicy([]string{"INNER_ONLY", "SECRET"}, security.OuterBoundary)

	if err == nil {
		t.Fatal("expected error for boundary violation, got nil")
	}

	if err.Error() != "taint INNER_ONLY is forbidden on boundary outer" {
		t.Fatalf("expected taint policy violation error, got: %v", err)
	}
}

func TestPersistence_EnforcesTaintPolicy(t *testing.T) {
	panic("not implemented")
}

func TestPersistence_AllowsCompliantWrites(t *testing.T) {
	panic("not implemented")
}
