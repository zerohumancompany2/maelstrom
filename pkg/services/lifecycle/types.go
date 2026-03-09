package lifecycle

import "github.com/maelstrom/v3/pkg/mail"

type RuntimeInfo struct {
	ID           string
	DefinitionID string
	Boundary     mail.BoundaryType
	ActiveStates []string
	IsRunning    bool
}
