package s3

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
)

// BucketManager manages all S3 bucket clients
type BucketManager struct {
	// Map of bucket name to bucket instance
	buckets map[string]*Bucket

	// Default bucket name
	defaultBucket string

	// Logger
	log *zap.Logger

	// Mutex for thread-safe access
	mu sync.RWMutex
}

// Bucket represents a single S3 bucket with its client and configuration
type Bucket struct {
	// Name is the bucket identifier in the plugin
	Name string

	// Config is the bucket configuration
	Config *BucketConfig

	// Client is the AWS S3 client
	Client *s3.Client

	// Semaphore for limiting concurrent operations
	sem chan struct{}
}

// NewBucketManager creates a new bucket manager
func NewBucketManager(log *zap.Logger) *BucketManager {
	return &BucketManager{
		buckets: make(map[string]*Bucket),
		log:     log,
	}
}

// RegisterBucket registers a new bucket with S3 client initialization
func (bm *BucketManager) RegisterBucket(ctx context.Context, name string, cfg *BucketConfig) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	// Check if bucket already exists
	if _, exists := bm.buckets[name]; exists {
		return fmt.Errorf("bucket '%s' already registered", name)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid bucket configuration: %w", err)
	}

	// Create AWS configuration
	awsCfg, err := bm.createAWSConfig(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to create AWS config: %w", err)
	}

	// Create S3 client
	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			o.UsePathStyle = true // Required for MinIO and some S3-compatible services
		}
	})

	// Create bucket instance
	bucket := &Bucket{
		Name:   name,
		Config: cfg,
		Client: s3Client,
		sem:    make(chan struct{}, cfg.MaxConcurrentOperations),
	}

	// Store bucket
	bm.buckets[name] = bucket

	bm.log.Debug("bucket registered",
		zap.String("name", name),
		zap.String("bucket", cfg.Bucket),
		zap.String("region", cfg.Region),
		zap.String("endpoint", cfg.Endpoint),
	)

	return nil
}

// GetBucket retrieves a bucket by name
func (bm *BucketManager) GetBucket(name string) (*Bucket, error) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	bucket, exists := bm.buckets[name]
	if !exists {
		return nil, fmt.Errorf("bucket '%s' not found", name)
	}

	return bucket, nil
}

// GetDefaultBucket retrieves the default bucket
func (bm *BucketManager) GetDefaultBucket() (*Bucket, error) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	if bm.defaultBucket == "" {
		return nil, fmt.Errorf("no default bucket configured")
	}

	bucket, exists := bm.buckets[bm.defaultBucket]
	if !exists {
		return nil, fmt.Errorf("default bucket '%s' not found", bm.defaultBucket)
	}

	return bucket, nil
}

// SetDefault sets the default bucket
func (bm *BucketManager) SetDefault(name string) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if _, exists := bm.buckets[name]; !exists {
		return fmt.Errorf("bucket '%s' not found", name)
	}

	bm.defaultBucket = name
	bm.log.Debug("default bucket set", zap.String("name", name))
	return nil
}

// ListBuckets returns all registered bucket names
func (bm *BucketManager) ListBuckets() []string {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	names := make([]string, 0, len(bm.buckets))
	for name := range bm.buckets {
		names = append(names, name)
	}

	return names
}

// GetDefaultBucketName returns the default bucket name
func (bm *BucketManager) GetDefaultBucketName() string {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	return bm.defaultBucket
}

// RemoveBucket removes a bucket (used for dynamic buckets)
func (bm *BucketManager) RemoveBucket(name string) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if name == bm.defaultBucket {
		return fmt.Errorf("cannot remove default bucket '%s'", name)
	}

	if _, exists := bm.buckets[name]; !exists {
		return fmt.Errorf("bucket '%s' not found", name)
	}

	delete(bm.buckets, name)
	bm.log.Debug("bucket removed", zap.String("name", name))
	return nil
}

// CloseAll closes all bucket clients
func (bm *BucketManager) CloseAll() error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	// AWS SDK v2 doesn't require explicit client closing
	// But we clean up resources
	for name := range bm.buckets {
		close(bm.buckets[name].sem)
	}

	bm.buckets = make(map[string]*Bucket)
	bm.log.Debug("all bucket clients closed")
	return nil
}

// createAWSConfig creates AWS configuration from bucket config
func (bm *BucketManager) createAWSConfig(ctx context.Context, cfg *BucketConfig) (aws.Config, error) {
	// Create credentials provider
	credsProvider := credentials.NewStaticCredentialsProvider(
		cfg.Credentials.Key,
		cfg.Credentials.Secret,
		cfg.Credentials.Token,
	)

	// Load AWS config with custom credentials
	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credsProvider),
	)
	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return awsCfg, nil
}

// Acquire acquires a semaphore slot for the bucket
func (b *Bucket) Acquire() {
	b.sem <- struct{}{}
}

// Release releases a semaphore slot for the bucket
func (b *Bucket) Release() {
	<-b.sem
}

// GetFullPath returns the full S3 key including prefix
func (b *Bucket) GetFullPath(pathname string) string {
	return b.Config.GetFullPath(pathname)
}

// GetVisibility returns the ACL for the bucket
func (b *Bucket) GetVisibility() string {
	return b.Config.GetVisibility()
}
