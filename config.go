package s3

import (
	"fmt"
)

// Config represents the plugin configuration from .rr.yaml
type Config struct {
	// Default bucket name to use when none specified
	Default string `mapstructure:"default"`

	// Servers contains S3 server definitions (credentials and endpoints)
	Servers map[string]*ServerConfig `mapstructure:"servers"`

	// Buckets contains bucket definitions that reference servers
	Buckets map[string]*BucketConfig `mapstructure:"buckets"`
}

// ServerConfig represents S3 server configuration (credentials and endpoint)
type ServerConfig struct {
	// Region is the AWS region (e.g., "us-east-1", "fra1" for DigitalOcean)
	Region string `mapstructure:"region"`

	// Endpoint is the S3 endpoint URL (required for S3-compatible services)
	// Example: "https://fra1.digitaloceanspaces.com"
	// Leave empty for AWS S3 (will use default AWS endpoint)
	Endpoint string `mapstructure:"endpoint"`

	// Credentials contains authentication credentials for this server
	Credentials ServerCredentials `mapstructure:"credentials"`
}

// ServerCredentials contains S3 authentication credentials
type ServerCredentials struct {
	// Key is the Access Key ID
	Key string `mapstructure:"key"`

	// Secret is the Secret Access Key
	Secret string `mapstructure:"secret"`

	// Token is the Session Token (optional, for temporary credentials)
	Token string `mapstructure:"token"`
}

// BucketConfig represents a single bucket configuration
type BucketConfig struct {
	// Server is the reference to a server defined in the servers section
	Server string `mapstructure:"server"`

	// Bucket is the actual S3 bucket name
	Bucket string `mapstructure:"bucket"`

	// Prefix is the path prefix for all operations (optional)
	// Example: "uploads/" - all files will be stored under this prefix
	Prefix string `mapstructure:"prefix"`

	// Visibility defines default ACL: "public" or "private"
	Visibility string `mapstructure:"visibility"`

	// MaxConcurrentOperations limits concurrent operations per bucket (default: 100)
	MaxConcurrentOperations int `mapstructure:"max_concurrent_operations"`

	// PartSize defines multipart upload part size in bytes (default: 5MB)
	PartSize int64 `mapstructure:"part_size"`

	// Concurrency defines number of goroutines for multipart uploads (default: 5)
	Concurrency int `mapstructure:"concurrency"`
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if len(c.Servers) == 0 {
		return fmt.Errorf("at least one server must be configured")
	}

	if len(c.Buckets) == 0 {
		return fmt.Errorf("at least one bucket must be configured")
	}

	// Validate each server configuration
	for name, server := range c.Servers {
		if err := server.Validate(); err != nil {
			return fmt.Errorf("invalid configuration for server '%s': %w", name, err)
		}
	}

	// Validate each bucket configuration
	for name, bucket := range c.Buckets {
		if err := bucket.Validate(c.Servers); err != nil {
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

// Validate validates a server configuration
func (sc *ServerConfig) Validate() error {
	if sc.Region == "" {
		return fmt.Errorf("region is required")
	}

	if sc.Credentials.Key == "" {
		return fmt.Errorf("credentials.key is required")
	}

	if sc.Credentials.Secret == "" {
		return fmt.Errorf("credentials.secret is required")
	}

	return nil
}

// Validate validates a bucket configuration
func (bc *BucketConfig) Validate(servers map[string]*ServerConfig) error {
	if bc.Server == "" {
		return fmt.Errorf("server reference is required")
	}

	// Validate server reference exists
	if _, exists := servers[bc.Server]; !exists {
		return fmt.Errorf("referenced server '%s' not found in configuration", bc.Server)
	}

	if bc.Bucket == "" {
		return fmt.Errorf("bucket name is required")
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

// GetServerConfig returns the server configuration for this bucket
func (bc *BucketConfig) GetServerConfig(servers map[string]*ServerConfig) (*ServerConfig, error) {
	server, exists := servers[bc.Server]
	if !exists {
		return nil, fmt.Errorf("server '%s' not found", bc.Server)
	}
	return server, nil
}
