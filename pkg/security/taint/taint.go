package taint

import (
	"sync"
)

type TaintType string

const (
	TaintUserSupplied TaintType = "USER_SUPPLIED"
	TaintExternal     TaintType = "EXTERNAL"
	TaintInnerOnly    TaintType = "INNER_ONLY"
	TaintSecret       TaintType = "SECRET"
	TaintPII          TaintType = "PII"
)

type Taint struct {
	Type   TaintType
	Source string
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

func (t TaintSet) ToSlice() []string {
	result := make([]string, 0, len(t))
	for k := range t {
		result = append(result, k)
	}
	return result
}

type TaintEngine struct {
	mu       sync.Mutex
	taintMap map[string]TaintSet
}

func NewTaintEngine() *TaintEngine {
	return &TaintEngine{
		taintMap: make(map[string]TaintSet),
	}
}

func (e *TaintEngine) Attach(data any, source string) (any, []string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	taints := e.determineTaints(source)
	return e.attachTaintsToData(data, taints)
}

func (e *TaintEngine) determineTaints(source string) []string {
	taints := make([]string, 0)

	switch {
	case source == "user" || source == "user_input" || source == "user_message":
		taints = append(taints, string(TaintUserSupplied))
	case source == "external" || source == "external_api" || source == "external_service":
		taints = append(taints, string(TaintExternal))
	case source == "inner" || source == "inner_db" || source == "inner_store":
		taints = append(taints, string(TaintInnerOnly))
	case source == "secret" || source == "secrets_store" || source == "secret_key":
		taints = append(taints, string(TaintSecret))
	case source == "pii" || source == "pii_store" || source == "user_data":
		taints = append(taints, string(TaintPII))
	default:
		taints = append(taints, string(TaintExternal))
	}

	return taints
}

func (e *TaintEngine) attachTaintsToData(data any, taints []string) (any, []string) {
	if data == nil {
		return nil, taints
	}

	switch v := data.(type) {
	case map[string]any:
		return e.attachTaintsToMap(v, taints), taints
	case *mailMail:
		return e.attachTaintsToMail(v, taints), taints
	default:
		return data, taints
	}
}

func (e *TaintEngine) attachTaintsToMap(m map[string]any, taints []string) map[string]any {
	result := make(map[string]any)

	for k, val := range m {
		switch v := val.(type) {
		case map[string]any:
			result[k] = e.attachTaintsToMap(v, taints)
		case []any:
			result[k] = e.attachTaintsToSlice(v, taints)
		default:
			result[k] = val
		}
	}

	merged := e.mergeTaints(m, taints)
	if len(merged) > 0 {
		result["_taints"] = merged
	}

	return result
}

func (e *TaintEngine) attachTaintsToSlice(slice []any, taints []string) []any {
	result := make([]any, len(slice))
	for i, elem := range slice {
		switch v := elem.(type) {
		case map[string]any:
			result[i] = e.attachTaintsToMap(v, taints)
		default:
			result[i] = elem
		}
	}
	return result
}

func (e *TaintEngine) mergeTaints(m map[string]any, newTaints []string) []string {
	existing := make(map[string]bool)

	if existingTaints, ok := m["_taints"].([]string); ok {
		for _, t := range existingTaints {
			existing[t] = true
		}
	}

	for _, t := range newTaints {
		existing[t] = true
	}

	merged := make([]string, 0, len(existing))
	for t := range existing {
		merged = append(merged, t)
	}

	return merged
}

func (e *TaintEngine) Propagate(src any, dest any) any {
	e.mu.Lock()
	defer e.mu.Unlock()

	srcTaints := e.extractTaints(src)
	if len(srcTaints) == 0 {
		return dest
	}

	result, _ := e.attachTaintsToData(dest, srcTaints)
	return result
}

func (e *TaintEngine) extractTaints(data any) []string {
	if data == nil {
		return nil
	}

	switch v := data.(type) {
	case map[string]any:
		if taints, ok := v["_taints"].([]string); ok {
			return taints
		}
	case *mailMail:
		return v.Taints
	}

	return nil
}

func (e *TaintEngine) Strip(data any, taints []string) any {
	e.mu.Lock()
	defer e.mu.Unlock()

	if data == nil {
		return nil
	}

	stripSet := make(map[string]bool)
	for _, t := range taints {
		stripSet[t] = true
	}

	switch v := data.(type) {
	case map[string]any:
		return e.stripTaintsFromMap(v, stripSet)
	case *mailMail:
		return e.stripTaintsFromMail(v, stripSet)
	default:
		return data
	}
}

func (e *TaintEngine) stripTaintsFromMap(m map[string]any, stripSet map[string]bool) map[string]any {
	result := make(map[string]any)

	if taints, ok := m["_taints"].([]string); ok {
		remaining := make([]string, 0)
		for _, t := range taints {
			if !stripSet[t] {
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
		case map[string]any:
			result[k] = e.stripTaintsFromMap(v, stripSet)
		case []any:
			result[k] = e.stripTaintsFromSlice(v, stripSet)
		default:
			result[k] = val
		}
	}

	return result
}

func (e *TaintEngine) stripTaintsFromSlice(slice []any, stripSet map[string]bool) []any {
	result := make([]any, len(slice))
	for i, elem := range slice {
		switch v := elem.(type) {
		case map[string]any:
			result[i] = e.stripTaintsFromMap(v, stripSet)
		default:
			result[i] = elem
		}
	}
	return result
}

type mailMail struct {
	Taints []string
}

func (e *TaintEngine) attachTaintsToMail(mail *mailMail, taints []string) *mailMail {
	existing := make(map[string]bool)
	for _, t := range mail.Taints {
		existing[t] = true
	}
	for _, t := range taints {
		existing[t] = true
	}
	merged := make([]string, 0, len(existing))
	for t := range existing {
		merged = append(merged, t)
	}
	mail.Taints = merged
	return mail
}

func (e *TaintEngine) stripTaintsFromMail(mail *mailMail, stripSet map[string]bool) *mailMail {
	remaining := make([]string, 0)
	for _, t := range mail.Taints {
		if !stripSet[t] {
			remaining = append(remaining, t)
		}
	}
	mail.Taints = remaining
	return mail
}
