package testutil
import "errors"

// MockApplicationContext is a test double for ApplicationContext.
type MockApplicationContext struct {
	Data      map[string]any
	Taints    map[string][]string
	Ns        string
	GetErr    error
	SetErr    error
}

// NewMockApplicationContext creates a new mock context.
func NewMockApplicationContext() *MockApplicationContext {
	return &MockApplicationContext{
		Data:   make(map[string]any),
		Taints: make(map[string][]string),
		Ns:     "test-namespace",
	}
}

// Get retrieves a value and its taints.
func (m *MockApplicationContext) Get(key string, callerBoundary string) (any, []string, error) {
	if m.GetErr != nil {
		return nil, nil, m.GetErr
	}
	val, exists := m.Data[key]
	if !exists {
		return nil, nil, errors.New("key not found")
	}
	return val, m.Taints[key], nil
}

// Set stores a value with taints.
func (m *MockApplicationContext) Set(key string, value any, taints []string, callerBoundary string) error {
	if m.SetErr != nil {
		return m.SetErr
	}
	m.Data[key] = value
	m.Taints[key] = taints
	return nil
}

// Namespace returns the namespace.
func (m *MockApplicationContext) Namespace() string {
	return m.Ns
}
