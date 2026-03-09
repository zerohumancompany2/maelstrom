package persistence

import (
	"testing"

	"github.com/maelstrom/v3/pkg/datasource"
	"github.com/maelstrom/v3/pkg/security"
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
	p := &Persistence{
		taintPolicy: &security.TaintPolicy{
			RedactMode: "strict",
		},
		dataSource: nil,
	}

	tests := []struct {
		name     string
		taints   []string
		boundary security.BoundaryType
		wantErr  bool
	}{
		{"INNER_ONLY on outer", []string{"INNER_ONLY"}, security.OuterBoundary, true},
		{"SECRET on DMZ", []string{"SECRET"}, security.DMZBoundary, true},
		{"PII on outer", []string{"PII"}, security.OuterBoundary, true},
		{"USER_SUPPLIED on outer", []string{"USER_SUPPLIED"}, security.OuterBoundary, false},
		{"TOOL_OUTPUT on DMZ", []string{"TOOL_OUTPUT"}, security.DMZBoundary, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := p.ValidateTaintPolicy(tt.taints, tt.boundary)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTaintPolicy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPersistence_AllowsCompliantWrites(t *testing.T) {
	ds := datasource.NewInMemoryDataSource()

	p := &Persistence{
		taintPolicy: &security.TaintPolicy{
			AllowedForBoundary: []security.BoundaryType{security.OuterBoundary},
		},
		dataSource: ds,
	}

	data := map[string]interface{}{"key": "value"}
	taints := []string{"USER_SUPPLIED"}

	err := p.Write(data, taints)
	if err != nil {
		t.Fatalf("Write() unexpected error: %v", err)
	}

	retrievedTaints, err := ds.GetTaints("key")
	if err != nil {
		t.Fatalf("GetTaints() unexpected error: %v", err)
	}

	if len(retrievedTaints) != 1 || retrievedTaints[0] != "USER_SUPPLIED" {
		t.Fatalf("expected taints [USER_SUPPLIED], got: %v", retrievedTaints)
	}
}
