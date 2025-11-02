package s3

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config with default bucket",
			config: Config{
				Default: "uploads",
				Buckets: map[string]*BucketConfig{
					"uploads": {
						Region: "us-east-1",
						Bucket: "my-bucket",
						Credentials: BucketCredentials{
							Key:    "key",
							Secret: "secret",
						},
						Visibility: "public",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid config without default",
			config: Config{
				Buckets: map[string]*BucketConfig{
					"uploads": {
						Region: "us-east-1",
						Bucket: "my-bucket",
						Credentials: BucketCredentials{
							Key:    "key",
							Secret: "secret",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "no buckets configured",
			config: Config{
				Buckets: map[string]*BucketConfig{},
			},
			wantErr: true,
			errMsg:  "at least one bucket must be configured",
		},
		{
			name: "default bucket not found",
			config: Config{
				Default: "nonexistent",
				Buckets: map[string]*BucketConfig{
					"uploads": {
						Region: "us-east-1",
						Bucket: "my-bucket",
						Credentials: BucketCredentials{
							Key:    "key",
							Secret: "secret",
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "default bucket 'nonexistent' not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestBucketConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  BucketConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: BucketConfig{
				Region: "us-east-1",
				Bucket: "my-bucket",
				Credentials: BucketCredentials{
					Key:    "key",
					Secret: "secret",
				},
				Visibility: "public",
			},
			wantErr: false,
		},
		{
			name: "missing region",
			config: BucketConfig{
				Bucket: "my-bucket",
				Credentials: BucketCredentials{
					Key:    "key",
					Secret: "secret",
				},
			},
			wantErr: true,
			errMsg:  "region is required",
		},
		{
			name: "missing bucket",
			config: BucketConfig{
				Region: "us-east-1",
				Credentials: BucketCredentials{
					Key:    "key",
					Secret: "secret",
				},
			},
			wantErr: true,
			errMsg:  "bucket name is required",
		},
		{
			name: "missing credentials key",
			config: BucketConfig{
				Region: "us-east-1",
				Bucket: "my-bucket",
				Credentials: BucketCredentials{
					Secret: "secret",
				},
			},
			wantErr: true,
			errMsg:  "credentials.key is required",
		},
		{
			name: "missing credentials secret",
			config: BucketConfig{
				Region: "us-east-1",
				Bucket: "my-bucket",
				Credentials: BucketCredentials{
					Key: "key",
				},
			},
			wantErr: true,
			errMsg:  "credentials.secret is required",
		},
		{
			name: "invalid visibility",
			config: BucketConfig{
				Region: "us-east-1",
				Bucket: "my-bucket",
				Credentials: BucketCredentials{
					Key:    "key",
					Secret: "secret",
				},
				Visibility: "invalid",
			},
			wantErr: true,
			errMsg:  "visibility must be 'public' or 'private'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				// Check defaults are set
				if tt.config.Visibility == "" {
					assert.Equal(t, "private", tt.config.Visibility)
				}
				assert.Greater(t, tt.config.MaxConcurrentOperations, 0)
				assert.Greater(t, tt.config.PartSize, int64(0))
				assert.Greater(t, tt.config.Concurrency, 0)
			}
		})
	}
}

func TestBucketConfig_GetVisibility(t *testing.T) {
	tests := []struct {
		name       string
		visibility string
		want       string
	}{
		{
			name:       "public visibility",
			visibility: "public",
			want:       "public-read",
		},
		{
			name:       "private visibility",
			visibility: "private",
			want:       "private",
		},
		{
			name:       "empty defaults to private",
			visibility: "",
			want:       "private",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &BucketConfig{Visibility: tt.visibility}
			got := cfg.GetVisibility()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBucketConfig_GetFullPath(t *testing.T) {
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
		{
			name:     "with nested path",
			prefix:   "uploads/",
			pathname: "images/photo.jpg",
			want:     "uploads/images/photo.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &BucketConfig{Prefix: tt.prefix}
			got := cfg.GetFullPath(tt.pathname)
			assert.Equal(t, tt.want, got)
		})
	}
}
