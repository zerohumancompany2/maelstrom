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
	tags               map[string][]string
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
		tags:               make(map[string][]string),
	}, nil
}

func (s *s3DataSource) TagOnWrite(key string, taints []string) error {
	s.tags[key] = make([]string, len(taints))
	copy(s.tags[key], taints)
	return nil
}

func (s *s3DataSource) GetTaints(key string) ([]string, error) {
	taints, ok := s.tags[key]
	if !ok {
		return []string{}, nil
	}
	return taints, nil
}

func (s *s3DataSource) ValidateAccess(boundary security.BoundaryType) error {
	if len(s.allowedForBoundary) == 0 {
		return nil
	}
	for _, allowed := range s.allowedForBoundary {
		if allowed == boundary {
			return nil
		}
	}
	return fmt.Errorf("boundary %s not allowed for this datasource", boundary)
}

func init() {
	Register("s3", NewS3DataSource)
}
