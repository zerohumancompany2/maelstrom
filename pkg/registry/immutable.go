package registry

import (
	"errors"
	"sync"

	"github.com/maelstrom/v3/pkg/mail"
	"github.com/maelstrom/v3/pkg/services/persistence"
)

var ErrBoundaryChangeNotAllowed = errors.New("boundary change is not allowed - boundary is immutable")

type ImmutableBoundaryEnforcer struct {
	mu sync.Mutex
}

func NewImmutableBoundaryEnforcer() *ImmutableBoundaryEnforcer {
	return &ImmutableBoundaryEnforcer{}
}

func (e *ImmutableBoundaryEnforcer) EnforceBoundaryImmutability(oldBoundary, newBoundary mail.BoundaryType) error {
	if oldBoundary != newBoundary {
		return ErrBoundaryChangeNotAllowed
	}
	return nil
}

func (e *ImmutableBoundaryEnforcer) ValidateMigrationExcludesBoundary(migrationPolicy persistence.MigrationPolicy) error {
	return nil
}

func (e *ImmutableBoundaryEnforcer) PreserveBoundaryDuringHotReload(currentBoundary mail.BoundaryType) mail.BoundaryType {
	return currentBoundary
}
