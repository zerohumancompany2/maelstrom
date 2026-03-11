package datasource

import (
	"github.com/maelstrom/v3/pkg/security/taint"
)

type DataSource interface {
	Read(path string) (any, []string, error)
	Write(path string, data any, taints []string) error
}

type FileDataSource interface {
	DataSource
	Read(path string) (any, []string, error)
	Write(path string, data any, taints []string) error
}

type ObjectDataSource interface {
	DataSource
	GetObject(bucket, key string) (any, []string, error)
	PutObject(bucket, key string, data any, taints []string) error
}

type MemoryDataSource interface {
	DataSource
	Get(key string) (any, []string, error)
	Put(key string, data any, taints []string) error
}

type NetworkDataSource interface {
	DataSource
	Fetch(url string) (any, []string, error)
}

type Config struct {
	TaintEngine *taint.TaintEngine
}

func NewConfig() *Config {
	return &Config{
		TaintEngine: taint.NewTaintEngine(),
	}
}
