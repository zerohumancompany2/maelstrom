package statechart

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
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

// System event types for parallel region coordination.
const (
	// Parent → Region: Lifecycle control
	SysEnter  = "sys:enter"
	SysExit   = "sys:exit"
	SysPause  = "sys:pause"
	SysResume = "sys:resume"

	// Region → Parent: Coordination signals
	SysDone         = "sys:done"
	SysTransition   = "sys:transition"
	SysPanic        = "sys:panic"
	SysBackpressure = "sys:backpressure"
)

// IsSystem returns true if the event is a system event (has "sys:" prefix).
func (e Event) IsSystem() bool {
	return len(e.Type) > 4 && e.Type[:4] == "sys:"
}

// IsUser returns true if the event is a user/domain event.
func (e Event) IsUser() bool {
	return !e.IsSystem()
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
	ApplicationState map[string]any    // serialized app context
	RegionStates     map[string]string // Current state per region for parallel states
	IsParallel       bool              // Whether in parallel state at snapshot time
}

// RestoreOptions provides the definition and other required data for Restore.
type RestoreOptions struct {
	Definition ChartDefinition
	AppCtx     ApplicationContext
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
	ID           string
	Children     map[string]*Node
	RegionNames  []string // for parallel nodes: names of regions
	Transitions  []Transition
	EntryActions []string
	ExitActions  []string
	IsInitial    bool // default child for compound states
}

// NodeType derives the runtime type from the node's structure.
func (n *Node) NodeType() NodeType {
	// RegionNames explicitly indicates a parallel state
	if n.RegionNames != nil {
		return NodeTypeParallel
	}
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
	Spec         map[string]interface{}
}

// GetID returns the chart ID.
func (c ChartDefinition) GetID() string {
	return c.ID
}

// HydratorFunc transforms raw YAML bytes into a hydrated ChartDefinition.
type HydratorFunc func([]byte) (ChartDefinition, error)

// DefaultHydrator provides env substitution and template execution.
func DefaultHydrator() HydratorFunc {
	return func(content []byte) (ChartDefinition, error) {
		content, err := envSubstitute(content)
		if err != nil {
			return ChartDefinition{}, fmt.Errorf("env substitution: %w", err)
		}

		content, err = executeTemplates(content)
		if err != nil {
			return ChartDefinition{}, fmt.Errorf("template execution: %w", err)
		}

		var def ChartDefinition
		if err := yaml.Unmarshal(content, &def); err != nil {
			return ChartDefinition{}, fmt.Errorf("yaml parse: %w", err)
		}

		if err := validateChart(def); err != nil {
			return ChartDefinition{}, fmt.Errorf("validation: %w", err)
		}

		return def, nil
	}
}

func validateChart(def ChartDefinition) error {
	if def.ID == "" {
		return fmt.Errorf("chart ID is required")
	}
	if def.Version == "" {
		return fmt.Errorf("chart version is required")
	}
	return nil
}

func envSubstitute(input []byte) ([]byte, error) {
	result := string(input)

	for {
		start := strings.Index(result, "${")
		if start == -1 {
			break
		}
		end := strings.Index(result[start:], "}")
		if end == -1 {
			return nil, fmt.Errorf("unclosed ${ in template")
		}
		end += start

		varExpr := result[start+2 : end]
		defaultValue := ""
		hasDefault := false

		if idx := strings.Index(varExpr, ":-"); idx != -1 {
			defaultValue = varExpr[idx+2:]
			varExpr = varExpr[:idx]
			hasDefault = true
		}

		value := os.Getenv(varExpr)
		if value == "" {
			if !hasDefault {
				return nil, fmt.Errorf("required environment variable %s is not set", varExpr)
			}
			value = defaultValue
		}

		result = result[:start] + value + result[end+1:]
	}

	return []byte(result), nil
}

func executeTemplates(content []byte) ([]byte, error) {
	return content, nil
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
