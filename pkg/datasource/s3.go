package datasource

import (
	"fmt"
	"github.com/maelstrom/v3/pkg/security"
)

type s3DataSource struct {
	bucket             string
	region             string
	endpoint           string
	allowedForBoundary []security.BoundaryType
}

func NewS3DataSource(config map[string]any) (DataSource, error) {
	bucket, ok := config["bucket"].(string)
	if !ok || bucket == "" {
		return nil, fmt.Errorf("bucket is required")
	}

	region, ok := config["region"].(string)
	if !ok || region == "" {
		return nil, fmt.Errorf("region is required")
	}

	endpoint, ok := config["endpoint"].(string)
	if !ok {
		endpoint = ""
	}

	allowedForBoundary := []security.BoundaryType{}
	if allowed, ok := config["allowedForBoundary"].([]security.BoundaryType); ok {
		allowedForBoundary = allowed
	}

	return &s3DataSource{
		bucket:             bucket,
		region:             region,
		endpoint:           endpoint,
		allowedForBoundary: allowedForBoundary,
	}, nil
}

func (s *s3DataSource) TagOnWrite(key string, taints []string) error {
	return fmt.Errorf("not implemented")
}

func (s *s3DataSource) GetTaints(key string) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *s3DataSource) ValidateAccess(boundary security.BoundaryType) error {
	return fmt.Errorf("not implemented")
}

func init() {
	Register("s3", NewS3DataSource)
}
