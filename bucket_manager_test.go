package s3

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestBucketManager_RegisterBucket(t *testing.T) {
	log := zaptest.NewLogger(t)
	bm := NewBucketManager(log)
	ctx := context.Background()

	cfg := &BucketConfig{
		Region: "us-east-1",
		Bucket: "test-bucket",
		Credentials: BucketCredentials{
			Key:    "test-key",
			Secret: "test-secret",
		},
		Visibility:              "public",
		MaxConcurrentOperations: 100,
		PartSize:                5 * 1024 * 1024,
		Concurrency:             5,
	}

	t.Run("register new bucket", func(t *testing.T) {
		err := bm.RegisterBucket(ctx, "test", cfg)
		require.NoError(t, err)

		bucket, err := bm.GetBucket("test")
		require.NoError(t, err)
		assert.Equal(t, "test", bucket.Name)
		assert.Equal(t, cfg, bucket.Config)
		assert.NotNil(t, bucket.Client)
		assert.NotNil(t, bucket.sem)
	})

	t.Run("register duplicate bucket", func(t *testing.T) {
		err := bm.RegisterBucket(ctx, "test", cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already registered")
	})

	t.Run("register with invalid config", func(t *testing.T) {
		invalidCfg := &BucketConfig{
			Region: "", // Missing required field
			Bucket: "test-bucket",
		}
		err := bm.RegisterBucket(ctx, "invalid", invalidCfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid bucket configuration")
	})
}

func TestBucketManager_GetBucket(t *testing.T) {
	log := zaptest.NewLogger(t)
	bm := NewBucketManager(log)
	ctx := context.Background()

	cfg := &BucketConfig{
		Region: "us-east-1",
		Bucket: "test-bucket",
		Credentials: BucketCredentials{
			Key:    "test-key",
			Secret: "test-secret",
		},
	}
	cfg.Validate()

	_ = bm.RegisterBucket(ctx, "test", cfg)

	t.Run("get existing bucket", func(t *testing.T) {
		bucket, err := bm.GetBucket("test")
		require.NoError(t, err)
		assert.Equal(t, "test", bucket.Name)
	})

	t.Run("get non-existent bucket", func(t *testing.T) {
		_, err := bm.GetBucket("nonexistent")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestBucketManager_SetDefault(t *testing.T) {
	log := zaptest.NewLogger(t)
	bm := NewBucketManager(log)
	ctx := context.Background()

	cfg := &BucketConfig{
		Region: "us-east-1",
		Bucket: "test-bucket",
		Credentials: BucketCredentials{
			Key:    "test-key",
			Secret: "test-secret",
		},
	}
	cfg.Validate()

	_ = bm.RegisterBucket(ctx, "test", cfg)

	t.Run("set existing bucket as default", func(t *testing.T) {
		err := bm.SetDefault("test")
		require.NoError(t, err)
		assert.Equal(t, "test", bm.GetDefaultBucketName())

		bucket, err := bm.GetDefaultBucket()
		require.NoError(t, err)
		assert.Equal(t, "test", bucket.Name)
	})

	t.Run("set non-existent bucket as default", func(t *testing.T) {
		err := bm.SetDefault("nonexistent")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestBucketManager_ListBuckets(t *testing.T) {
	log := zaptest.NewLogger(t)
	bm := NewBucketManager(log)
	ctx := context.Background()

	cfg := &BucketConfig{
		Region: "us-east-1",
		Bucket: "test-bucket",
		Credentials: BucketCredentials{
			Key:    "test-key",
			Secret: "test-secret",
		},
	}
	cfg.Validate()

	t.Run("list empty buckets", func(t *testing.T) {
		buckets := bm.ListBuckets()
		assert.Empty(t, buckets)
	})

	t.Run("list registered buckets", func(t *testing.T) {
		_ = bm.RegisterBucket(ctx, "bucket1", cfg)
		_ = bm.RegisterBucket(ctx, "bucket2", cfg)

		buckets := bm.ListBuckets()
		assert.Len(t, buckets, 2)
		assert.Contains(t, buckets, "bucket1")
		assert.Contains(t, buckets, "bucket2")
	})
}

func TestBucketManager_RemoveBucket(t *testing.T) {
	log := zaptest.NewLogger(t)
	bm := NewBucketManager(log)
	ctx := context.Background()

	cfg := &BucketConfig{
		Region: "us-east-1",
		Bucket: "test-bucket",
		Credentials: BucketCredentials{
			Key:    "test-key",
			Secret: "test-secret",
		},
	}
	cfg.Validate()

	_ = bm.RegisterBucket(ctx, "test", cfg)
	_ = bm.SetDefault("test")

	t.Run("cannot remove default bucket", func(t *testing.T) {
		err := bm.RemoveBucket("test")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot remove default bucket")
	})

	_ = bm.RegisterBucket(ctx, "test2", cfg)

	t.Run("remove non-default bucket", func(t *testing.T) {
		err := bm.RemoveBucket("test2")
		require.NoError(t, err)

		_, err = bm.GetBucket("test2")
		require.Error(t, err)
	})

	t.Run("remove non-existent bucket", func(t *testing.T) {
		err := bm.RemoveBucket("nonexistent")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestBucket_SemaphoreOperations(t *testing.T) {
	bucket := &Bucket{
		Name: "test",
		Config: &BucketConfig{
			MaxConcurrentOperations: 2,
		},
		sem: make(chan struct{}, 2),
	}

	t.Run("acquire and release", func(t *testing.T) {
		// Acquire first slot
		bucket.Acquire()
		assert.Len(t, bucket.sem, 1)

		// Acquire second slot
		bucket.Acquire()
		assert.Len(t, bucket.sem, 2)

		// Release slots
		bucket.Release()
		assert.Len(t, bucket.sem, 1)

		bucket.Release()
		assert.Len(t, bucket.sem, 0)
	})
}

func TestBucket_GetFullPath(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		pathname string
		want     string
	}{
		{
			name:     "with prefix",
			prefix:   "uploads/",
			pathname: "file.txt",
			want:     "uploads/file.txt",
		},
		{
			name:     "without prefix",
			prefix:   "",
			pathname: "file.txt",
			want:     "file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bucket := &Bucket{
				Config: &BucketConfig{Prefix: tt.prefix},
			}
			got := bucket.GetFullPath(tt.pathname)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBucket_GetVisibility(t *testing.T) {
	tests := []struct {
		name       string
		visibility string
		want       string
	}{
		{
			name:       "public",
			visibility: "public",
			want:       "public-read",
		},
		{
			name:       "private",
			visibility: "private",
			want:       "private",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bucket := &Bucket{
				Config: &BucketConfig{Visibility: tt.visibility},
			}
			got := bucket.GetVisibility()
			assert.Equal(t, tt.want, got)
		})
	}
}
