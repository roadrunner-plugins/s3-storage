package s3

import (
	"fmt"
)

// Config represents the plugin configuration from .rr.yaml
type Config struct {
	// Default bucket name to use when none specified
	Default string `mapstructure:"default"`

	// Buckets contains pre-configured bucket definitions
	Buckets map[string]*BucketConfig `mapstructure:"buckets"`
}

// BucketConfig represents a single bucket configuration
type BucketConfig struct {
	// Region is the AWS region (e.g., "us-east-1")
	Region string `mapstructure:"region"`

	// Endpoint is the S3 endpoint URL (for custom S3-compatible services)
	// Leave empty for AWS S3 (will use default AWS endpoint)
	Endpoint string `mapstructure:"endpoint"`

	// Bucket is the actual S3 bucket name
	Bucket string `mapstructure:"bucket"`

	// Prefix is the path prefix for all operations (optional)
	// Example: "uploads/" - all files will be stored under this prefix
	Prefix string `mapstructure:"prefix"`

	// Credentials contains AWS credentials
	Credentials BucketCredentials `mapstructure:"credentials"`

	// Visibility defines default ACL: "public" or "private"
	Visibility string `mapstructure:"visibility"`

	// MaxConcurrentOperations limits concurrent operations per bucket (default: 100)
	MaxConcurrentOperations int `mapstructure:"max_concurrent_operations"`

	// PartSize defines multipart upload part size in bytes (default: 5MB)
	PartSize int64 `mapstructure:"part_size"`

	// Concurrency defines number of goroutines for multipart uploads (default: 5)
	Concurrency int `mapstructure:"concurrency"`
}

// BucketCredentials contains AWS authentication credentials
type BucketCredentials struct {
	// Key is the AWS Access Key ID
	Key string `mapstructure:"key"`

	// Secret is the AWS Secret Access Key
	Secret string `mapstructure:"secret"`

	// Token is the AWS Session Token (optional, for temporary credentials)
	Token string `mapstructure:"token"`
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if len(c.Buckets) == 0 {
		return fmt.Errorf("at least one bucket must be configured")
	}

	// Validate each bucket configuration
	for name, bucket := range c.Buckets {
		if err := bucket.Validate(); err != nil {
			return fmt.Errorf("invalid configuration for bucket '%s': %w", name, err)
		}
	}

	// Validate default bucket exists if specified
	if c.Default != "" {
		if _, exists := c.Buckets[c.Default]; !exists {
			return fmt.Errorf("default bucket '%s' not found in configuration", c.Default)
		}
	}

	return nil
}

// Validate validates a single bucket configuration
func (bc *BucketConfig) Validate() error {
	if bc.Region == "" {
		return fmt.Errorf("region is required")
	}

	if bc.Bucket == "" {
		return fmt.Errorf("bucket name is required")
	}

	if bc.Credentials.Key == "" {
		return fmt.Errorf("credentials.key is required")
	}

	if bc.Credentials.Secret == "" {
		return fmt.Errorf("credentials.secret is required")
	}

	if bc.Visibility != "" && bc.Visibility != "public" && bc.Visibility != "private" {
		return fmt.Errorf("visibility must be 'public' or 'private', got '%s'", bc.Visibility)
	}

	// Set defaults
	if bc.Visibility == "" {
		bc.Visibility = "private"
	}

	if bc.MaxConcurrentOperations <= 0 {
		bc.MaxConcurrentOperations = 100
	}

	if bc.PartSize <= 0 {
		bc.PartSize = 5 * 1024 * 1024 // 5MB default
	}

	if bc.Concurrency <= 0 {
		bc.Concurrency = 5
	}

	return nil
}

// GetVisibility returns the ACL string for S3 operations
func (bc *BucketConfig) GetVisibility() string {
	if bc.Visibility == "public" {
		return "public-read"
	}
	return "private"
}

// GetFullPath returns the full path including prefix
func (bc *BucketConfig) GetFullPath(pathname string) string {
	if bc.Prefix == "" {
		return pathname
	}
	return bc.Prefix + pathname
}
