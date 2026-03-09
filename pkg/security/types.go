package security

import (
	"fmt"

	"github.com/maelstrom/v3/pkg/mail"
)

type BoundaryType string

const (
	InnerBoundary BoundaryType = "inner"
	DMZBoundary   BoundaryType = "dmz"
	OuterBoundary BoundaryType = "outer"
)

type ContextBlock struct {
	Name           string
	Source         string
	Content        string
	Strategy       string
	N              int
	MaxTokens      int
	Eviction       string
	StreamCommit   bool
	Priority       int
	TaintPolicy    TaintPolicy
	BoundaryFilter BoundaryType
}

type TaintPolicy struct {
	RedactMode         string
	RedactRules        []RedactRule
	AllowedForBoundary []BoundaryType
}

type RedactRule struct {
	Taint       string
	Replacement string
}

type TaintSet map[string]bool

func (t TaintSet) Add(taint string) {
	t[taint] = true
}

func (t TaintSet) Has(taint string) bool {
	if t == nil {
		return false
	}
	return t[taint]
}

func (t TaintSet) Union(other TaintSet) TaintSet {
	result := make(TaintSet)
	for k, v := range t {
		result[k] = v
	}
	for k, v := range other {
		result[k] = v
	}
	return result
}

type TaintMap map[string][]string

func (tm TaintMap) Add(runtimeID, taint string) {
	tm[runtimeID] = append(tm[runtimeID], taint)
}

func (tm TaintMap) Merge(other TaintMap) {
	for k, v := range other {
		tm[k] = append(tm[k], v...)
	}
}

func assemble(blocks []*ContextBlock) []*ContextBlock {
	sorted := make([]*ContextBlock, len(blocks))
	copy(sorted, blocks)

	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].Priority < sorted[i].Priority {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted
}

func staticSource(block *ContextBlock) string {
	return block.Content
}

func sessionSource(block *ContextBlock, messages []string) string {
	n := block.N
	if n <= 0 {
		n = 10
	}
	result := ""
	for i := 0; i < len(messages) && i < n; i++ {
		if i > 0 {
			result += ", "
		}
		result += messages[i]
	}
	return result
}

func memorySource(block *ContextBlock, query string) string {
	return "queried: " + query
}

func toolSource(block *ContextBlock, tools []string, callerBoundary BoundaryType) string {
	var filtered []string
	for _, tool := range tools {
		toolBoundary := getToolBoundary(tool)
		if isBoundaryVisible(toolBoundary, callerBoundary) {
			filtered = append(filtered, tool)
		}
	}
	return joinStrings(filtered, ",")
}

func getToolBoundary(tool string) BoundaryType {
	if tool == "innerDbQuery" {
		return InnerBoundary
	}
	return DMZBoundary
}

func isBoundaryVisible(toolBoundary, callerBoundary BoundaryType) bool {
	switch callerBoundary {
	case InnerBoundary:
		return toolBoundary == InnerBoundary
	case DMZBoundary:
		return toolBoundary == DMZBoundary
	case OuterBoundary:
		return toolBoundary == DMZBoundary
	default:
		return false
	}
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

type TaintEngine interface {
	Mark(obj any, taints []string) (any, error)
	MarkRead(path string, taints []string) ([]string, error)
	Propagate(obj any, newTaints []string) (any, error)
	CheckForbidden(taints []string, boundary BoundaryType) error
	Redact(obj any, rules []RedactRule) (any, error)
	ReportTaints(chartID string) (TaintMap, error)
	AttachTaint(obj any, taints []string) (any, error)
}

type taintEngineImpl struct {
	taints TaintMap
}

func NewTaintEngine() TaintEngine {
	return &taintEngineImpl{
		taints: make(TaintMap),
	}
}

func (e *taintEngineImpl) Mark(obj any, taints []string) (any, error) {
	return obj, nil
}

func (e *taintEngineImpl) MarkRead(path string, taints []string) ([]string, error) {
	return taints, nil
}

func (e *taintEngineImpl) Propagate(obj any, newTaints []string) (any, error) {
	return obj, nil
}

func (e *taintEngineImpl) CheckForbidden(taints []string, boundary BoundaryType) error {
	return nil
}

func (e *taintEngineImpl) Redact(obj any, rules []RedactRule) (any, error) {
	return obj, nil
}

func (e *taintEngineImpl) ReportTaints(chartID string) (TaintMap, error) {
	return e.taints, nil
}

func (e *taintEngineImpl) AttachTaint(obj any, taints []string) (any, error) {
	if obj == nil {
		return nil, fmt.Errorf("cannot attach taints to nil object")
	}

	switch v := obj.(type) {
	case *mail.Mail:
		existing := make(map[string]bool)
		for _, t := range v.Metadata.Taints {
			existing[t] = true
		}
		for _, t := range taints {
			existing[t] = true
		}
		merged := make([]string, 0, len(existing))
		for t := range existing {
			merged = append(merged, t)
		}
		v.Metadata.Taints = merged
		return v, nil
	case map[string]interface{}:
		return e.attachTaintToMap(v, taints)
	default:
		return obj, nil
	}
}

func (e *taintEngineImpl) attachTaintToMap(m map[string]interface{}, taints []string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for k, val := range m {
		switch v := val.(type) {
		case map[string]interface{}:
			nested, err := e.attachTaintToMap(v, taints)
			if err != nil {
				return nil, err
			}
			result[k] = nested
		default:
			result[k] = val
		}
	}
	result["_taints"] = taints
	return result, nil
}

type BoundaryService interface {
	ValidateAndSanitize(mail any, src, tgt BoundaryType) (any, error)
	MarkTaint(obj any, taints []string) (any, error)
	CheckForbidden(taints []string, boundary BoundaryType) error
	NamespaceIsolate(chartID, operation string) (IsolatedView, error)
	ReportTaints(chartID string) (TaintMap, error)
	PrepareContextForBoundary(chartID string, boundary BoundaryType) error
}

type IsolatedView struct {
	ChartID   string
	Operation string
	Namespace string
	Isolated  bool
}

type boundaryServiceImpl struct {
	engine TaintEngine
}

func NewBoundaryService(engine TaintEngine) BoundaryService {
	return &boundaryServiceImpl{engine: engine}
}

func (s *boundaryServiceImpl) ValidateAndSanitize(mail any, src, tgt BoundaryType) (any, error) {
	return mail, nil
}

func (s *boundaryServiceImpl) MarkTaint(obj any, taints []string) (any, error) {
	return obj, nil
}

func (s *boundaryServiceImpl) CheckForbidden(taints []string, boundary BoundaryType) error {
	return nil
}

func (s *boundaryServiceImpl) NamespaceIsolate(chartID, operation string) (IsolatedView, error) {
	return IsolatedView{
		ChartID:   chartID,
		Operation: operation,
		Isolated:  true,
	}, nil
}

func (s *boundaryServiceImpl) ReportTaints(chartID string) (TaintMap, error) {
	return make(TaintMap), nil
}

func (s *boundaryServiceImpl) PrepareContextForBoundary(chartID string, boundary BoundaryType) error {
	return nil
}
