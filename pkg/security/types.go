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
	Taints         TaintSet
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

func (t *TaintSet) Add(taint string) {
	if *t == nil {
		*t = make(TaintSet)
	}
	(*t)[taint] = true
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
	StripTaint(obj any, forbiddenTaints []string) (any, []string, error)
	EnforceAllowedOnExit(taints []string, policy EnforcementPolicy) error
}

type EnforcementPolicy struct {
	AllowedOnExit []string
	Enforcement   string
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
	if obj == nil {
		return nil, fmt.Errorf("cannot propagate taints to nil object")
	}

	switch v := obj.(type) {
	case map[string]interface{}:
		return e.propagateTaintToMap(v, newTaints)
	default:
		return obj, nil
	}
}

func (e *taintEngineImpl) propagateTaintToMap(m map[string]interface{}, newTaints []string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	seen := make(map[string]bool)
	merged := make([]string, 0)

	if existing, ok := m["_taints"].([]string); ok {
		for _, t := range existing {
			if !seen[t] {
				seen[t] = true
				merged = append(merged, t)
			}
		}
	}

	for _, t := range newTaints {
		if !seen[t] {
			seen[t] = true
			merged = append(merged, t)
		}
	}

	result["_taints"] = merged

	for k, val := range m {
		if k == "_taints" {
			continue
		}
		switch v := val.(type) {
		case map[string]interface{}:
			nested, err := e.propagateTaintToMap(v, newTaints)
			if err != nil {
				return nil, err
			}
			result[k] = nested
		case []interface{}:
			result[k] = e.propagateTaintToSlice(v, newTaints)
		default:
			result[k] = val
		}
	}

	return result, nil
}

func (e *taintEngineImpl) propagateTaintToSlice(slice []interface{}, newTaints []string) []interface{} {
	result := make([]interface{}, len(slice))
	for i, elem := range slice {
		switch v := elem.(type) {
		case map[string]interface{}:
			nested, _ := e.propagateTaintToMap(v, newTaints)
			result[i] = nested
		default:
			result[i] = elem
		}
	}
	return result
}

func (e *taintEngineImpl) CheckForbidden(taints []string, boundary BoundaryType) error {
	for _, taint := range taints {
		if taint == "INNER_ONLY" && (boundary == DMZBoundary || boundary == OuterBoundary) {
			return fmt.Errorf("taint %s is forbidden on boundary %s", taint, boundary)
		}
		if taint == "SECRET" && (boundary == DMZBoundary || boundary == OuterBoundary) {
			return fmt.Errorf("taint %s is forbidden on boundary %s", taint, boundary)
		}
		if taint == "PII" && boundary == OuterBoundary {
			return fmt.Errorf("taint %s is forbidden on boundary %s", taint, boundary)
		}
	}
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

func (e *taintEngineImpl) StripTaint(obj any, forbiddenTaints []string) (any, []string, error) {
	if obj == nil {
		return nil, nil, fmt.Errorf("cannot strip taints from nil object")
	}

	forbiddenSet := make(map[string]bool)
	for _, t := range forbiddenTaints {
		forbiddenSet[t] = true
	}

	switch v := obj.(type) {
	case map[string]interface{}:
		return e.stripTaintFromMap(v, forbiddenSet)
	default:
		return obj, nil, nil
	}
}

func (e *taintEngineImpl) stripTaintFromMap(m map[string]interface{}, forbiddenSet map[string]bool) (map[string]interface{}, []string, error) {
	result := make(map[string]interface{})
	stripped := make([]string, 0)

	taints, ok := m["_taints"].([]string)
	if ok {
		remaining := make([]string, 0)
		for _, t := range taints {
			if forbiddenSet[t] {
				stripped = append(stripped, t)
			} else {
				remaining = append(remaining, t)
			}
		}
		if len(remaining) > 0 {
			result["_taints"] = remaining
		}
	}

	for k, val := range m {
		if k == "_taints" {
			continue
		}
		switch v := val.(type) {
		case map[string]interface{}:
			nested, nestedStripped, err := e.stripTaintFromMap(v, forbiddenSet)
			if err != nil {
				return nil, nil, err
			}
			result[k] = nested
			stripped = append(stripped, nestedStripped...)
		case []interface{}:
			result[k] = e.stripTaintFromSlice(v, forbiddenSet)
		default:
			result[k] = val
		}
	}

	return result, stripped, nil
}

func (e *taintEngineImpl) stripTaintFromSlice(slice []interface{}, forbiddenSet map[string]bool) []interface{} {
	result := make([]interface{}, len(slice))
	for i, elem := range slice {
		switch v := elem.(type) {
		case map[string]interface{}:
			nested, _, _ := e.stripTaintFromMap(v, forbiddenSet)
			result[i] = nested
		default:
			result[i] = elem
		}
	}
	return result
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

func (e *taintEngineImpl) EnforceAllowedOnExit(taints []string, policy EnforcementPolicy) error {
	allowedSet := make(map[string]bool)
	for _, t := range policy.AllowedOnExit {
		allowedSet[t] = true
	}

	for _, taint := range taints {
		if !allowedSet[taint] {
			return fmt.Errorf("taint %s is not allowed on exit", taint)
		}
	}
	return nil
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
	RuntimeID   string
	Operation   string
	Boundary    BoundaryType
	ContextData map[string]any
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
	return s.engine.CheckForbidden(taints, boundary)
}

func (s *boundaryServiceImpl) NamespaceIsolate(chartID, operation string) (IsolatedView, error) {
	return IsolatedView{
		RuntimeID: chartID,
		Operation: operation,
		Boundary:  DMZBoundary,
	}, nil
}

func (s *boundaryServiceImpl) ReportTaints(chartID string) (TaintMap, error) {
	result := make(TaintMap)
	engineImpl, ok := s.engine.(*taintEngineImpl)
	if !ok {
		return result, nil
	}
	for k, v := range engineImpl.taints {
		result[k] = make([]string, len(v))
		copy(result[k], v)
	}
	return result, nil
}

func (s *boundaryServiceImpl) PrepareContextForBoundary(chartID string, boundary BoundaryType) error {
	return nil
}
