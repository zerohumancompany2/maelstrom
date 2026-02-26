package statechart

import (
	"time"
)

// RuntimeID uniquely identifies a ChartRuntime instance within an Engine.
type RuntimeID string

// ControlCmd represents lifecycle commands for a ChartRuntime.
type ControlCmd string

const (
	CmdStart             ControlCmd = "start"
	CmdPause             ControlCmd = "pause"
	CmdResume            ControlCmd = "resume"
	CmdStop              ControlCmd = "stop"
	CmdSnapshot          ControlCmd = "snapshot"
	CmdRestore           ControlCmd = "restore"
	CmdInjectEvent       ControlCmd = "injectEvent"
	CmdReplaceDefinition ControlCmd = "replaceDefinition"
)

// Event drives internal transitions within a Chart.
type Event struct {
	Type          string
	Payload       any
	CorrelationID string
	Source        string
	TargetPath    string // optional: "region:foo", "child:bar", etc.
}

// RuntimeContext provides read-only context about a running chart.
type RuntimeContext struct {
	ChartID      string
	RuntimeID    string
	ParentID     *string
	ActiveStates []string
	StartTime    time.Time
	Metrics      map[string]float64
}

// Snapshot captures the serialized state of a ChartRuntime for persistence.
type Snapshot struct {
	RuntimeID        RuntimeID
	DefinitionID     string
	ParentID         *string
	ActiveStates     []string
	EventQueue       []Event
	RuntimeContext   RuntimeContext
	ApplicationState map[string]any // serialized app context
}

// NodeType classifies the runtime behavior of a Node based on its children.
type NodeType int

const (
	NodeTypeAtomic NodeType = iota
	NodeTypeCompound
	NodeTypeParallel
)

// String returns the string representation of NodeType.
func (nt NodeType) String() string {
	switch nt {
	case NodeTypeAtomic:
		return "atomic"
	case NodeTypeCompound:
		return "compound"
	case NodeTypeParallel:
		return "parallel"
	default:
		return "unknown"
	}
}

// Transition defines a state transition triggered by an event.
type Transition struct {
	Event   string
	Guard   string // name of registered guard, empty if none
	Actions []string
	Target  string
}

// Node is the unified state primitive.
type Node struct {
	ID            string
	Children      map[string]*Node
	RegionNames   []string // for parallel nodes: names of regions
	Transitions   []Transition
	EntryActions  []string
	ExitActions   []string
	IsInitial     bool // default child for compound states
}

// NodeType derives the runtime type from the node's structure.
func (n *Node) NodeType() NodeType {
	switch len(n.Children) {
	case 0:
		return NodeTypeAtomic
	case 1:
		return NodeTypeCompound
	default:
		return NodeTypeParallel
	}
}

// ChartDefinition is the immutable, hydrated definition of a statechart.
type ChartDefinition struct {
	ID           string
	Version      string
	Root         *Node
	InitialState string // initial state path (e.g., "root/child1")
}

// RuntimeState represents the lifecycle state of a ChartRuntime.
type RuntimeState int

const (
	RuntimeStateCreated RuntimeState = iota
	RuntimeStateRunning
	RuntimeStatePaused
	RuntimeStateStopped
)

// String returns the string representation of RuntimeState.
func (rs RuntimeState) String() string {
	switch rs {
	case RuntimeStateCreated:
		return "created"
	case RuntimeStateRunning:
		return "running"
	case RuntimeStatePaused:
		return "paused"
	case RuntimeStateStopped:
		return "stopped"
	default:
		return "unknown"
	}
}
