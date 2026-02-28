package chart

import (
	"fmt"
	"os"
	"strings"
)

// HydratorFunc transforms raw YAML bytes into a hydrated ChartDefinition.
type HydratorFunc func([]byte) (ChartDefinition, error)

// ChartDefinition represents a hydrated chart ready for instantiation.
type ChartDefinition struct {
	ID      string
	Version string
	Spec    map[string]interface{}
}

// DefaultHydrator provides env substitution and template execution.
func DefaultHydrator() HydratorFunc {
	return func(content []byte) (ChartDefinition, error) {
		panic("not implemented")
	}
}

// envSubstitute replaces ${VAR} and ${VAR:-default} patterns.
func envSubstitute(input []byte) ([]byte, error) {
	result := string(input)

	// Simple ${VAR} substitution
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

		// Check for ${VAR:-default} syntax
		if idx := strings.Index(varExpr, ":-"); idx != -1 {
			defaultValue = varExpr[idx+2:]
			varExpr = varExpr[:idx]
		}

		value := os.Getenv(varExpr)
		if value == "" {
			value = defaultValue
		}

		result = result[:start] + value + result[end+1:]
	}

	return []byte(result), nil
}

// executeTemplates processes {{template "name"}} directives.
func executeTemplates(content []byte) ([]byte, error) {
	// Stub - would load and execute templates from a template registry
	return content, nil
}
