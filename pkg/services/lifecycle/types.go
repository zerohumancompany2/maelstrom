package lifecycle

import (
	"time"

	"github.com/maelstrom/v3/pkg/mail"
)

type RuntimeInfo struct {
	ID           string
	DefinitionID string
	Boundary     mail.BoundaryType
	ActiveStates []string
	IsRunning    bool
}

type StateTransition struct {
	From      string
	To        string
	Timestamp time.Time
}
