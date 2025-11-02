package s3

import (
	"go.uber.org/zap"
)

// rpc implements the RPC interface exposed to PHP via goridge
type rpc struct {
	plugin *Plugin
	log    *zap.Logger
}

// RegisterBucketRequest represents the request to register a new bucket
type RegisterBucketRequest struct {
	Name        string            `json:"name"`
	Region      string            `json:"region"`
	Endpoint    string            `json:"endpoint"`
	Bucket      string            `json:"bucket"`
	Prefix      string            `json:"prefix"`
	Credentials BucketCredentials `json:"credentials"`
	Visibility  string            `json:"visibility"`
}

// RegisterBucketResponse represents the response from bucket registration
type RegisterBucketResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ListBucketsRequest represents the request to list all buckets
type ListBucketsRequest struct{}

// ListBucketsResponse represents the response with all bucket names
type ListBucketsResponse struct {
	Buckets []string `json:"buckets"`
	Default string   `json:"default"`
}

// WriteRequest represents a file write/upload request
type WriteRequest struct {
	Bucket     string            `json:"bucket"`
	Pathname   string            `json:"pathname"`
	Content    []byte            `json:"content"`
	Config     map[string]string `json:"config,omitempty"`
	Visibility string            `json:"visibility,omitempty"`
}

// WriteResponse represents the response from a write operation
type WriteResponse struct {
	Success      bool   `json:"success"`
	Pathname     string `json:"pathname"`
	Size         int64  `json:"size"`
	LastModified int64  `json:"last_modified"`
}

// ReadRequest represents a file read/download request
type ReadRequest struct {
	Bucket   string `json:"bucket"`
	Pathname string `json:"pathname"`
}

// ReadResponse represents the response from a read operation
type ReadResponse struct {
	Content      []byte `json:"content"`
	Size         int64  `json:"size"`
	MimeType     string `json:"mime_type"`
	LastModified int64  `json:"last_modified"`
}

// ExistsRequest represents a file existence check request
type ExistsRequest struct {
	Bucket   string `json:"bucket"`
	Pathname string `json:"pathname"`
}

// ExistsResponse represents the response from an exists check
type ExistsResponse struct {
	Exists bool `json:"exists"`
}

// DeleteRequest represents a file deletion request
type DeleteRequest struct {
	Bucket   string `json:"bucket"`
	Pathname string `json:"pathname"`
}

// DeleteResponse represents the response from a delete operation
type DeleteResponse struct {
	Success bool `json:"success"`
}

// CopyRequest represents a file copy request
type CopyRequest struct {
	SourceBucket   string            `json:"source_bucket"`
	SourcePathname string            `json:"source_pathname"`
	DestBucket     string            `json:"dest_bucket"`
	DestPathname   string            `json:"dest_pathname"`
	Config         map[string]string `json:"config,omitempty"`
	Visibility     string            `json:"visibility,omitempty"`
}

// CopyResponse represents the response from a copy operation
type CopyResponse struct {
	Success      bool   `json:"success"`
	Pathname     string `json:"pathname"`
	Size         int64  `json:"size"`
	LastModified int64  `json:"last_modified"`
}

// MoveRequest represents a file move request (copy + delete)
type MoveRequest struct {
	SourceBucket   string            `json:"source_bucket"`
	SourcePathname string            `json:"source_pathname"`
	DestBucket     string            `json:"dest_bucket"`
	DestPathname   string            `json:"dest_pathname"`
	Config         map[string]string `json:"config,omitempty"`
	Visibility     string            `json:"visibility,omitempty"`
}

// MoveResponse represents the response from a move operation
type MoveResponse struct {
	Success      bool   `json:"success"`
	Pathname     string `json:"pathname"`
	Size         int64  `json:"size"`
	LastModified int64  `json:"last_modified"`
}

// GetMetadataRequest represents a request to get file metadata
type GetMetadataRequest struct {
	Bucket   string `json:"bucket"`
	Pathname string `json:"pathname"`
}

// GetMetadataResponse represents file metadata
type GetMetadataResponse struct {
	Size         int64  `json:"size"`
	MimeType     string `json:"mime_type"`
	LastModified int64  `json:"last_modified"`
	Visibility   string `json:"visibility"`
	ETag         string `json:"etag,omitempty"`
}

// SetVisibilityRequest represents a request to change file visibility
type SetVisibilityRequest struct {
	Bucket     string `json:"bucket"`
	Pathname   string `json:"pathname"`
	Visibility string `json:"visibility"`
}

// SetVisibilityResponse represents the response from visibility change
type SetVisibilityResponse struct {
	Success bool `json:"success"`
}

// GetPublicURLRequest represents a request to generate a public URL
type GetPublicURLRequest struct {
	Bucket    string `json:"bucket"`
	Pathname  string `json:"pathname"`
	ExpiresIn int64  `json:"expires_in,omitempty"` // Seconds, 0 for permanent
}

// GetPublicURLResponse represents the response with a public URL
type GetPublicURLResponse struct {
	URL       string `json:"url"`
	ExpiresAt int64  `json:"expires_at,omitempty"` // Unix timestamp
}

// RegisterBucket registers a new bucket dynamically via RPC
func (r *rpc) RegisterBucket(req *RegisterBucketRequest, resp *RegisterBucketResponse) error {
	r.log.Info("registering bucket via RPC",
		zap.String("name", req.Name),
		zap.String("bucket", req.Bucket),
		zap.String("region", req.Region),
	)

	// Create bucket configuration from request
	cfg := &BucketConfig{
		Region:      req.Region,
		Endpoint:    req.Endpoint,
		Bucket:      req.Bucket,
		Prefix:      req.Prefix,
		Credentials: req.Credentials,
		Visibility:  req.Visibility,
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		resp.Success = false
		resp.Message = "Invalid configuration: " + err.Error()
		return NewInvalidConfigError(err.Error())
	}

	// Register bucket
	if err := r.plugin.buckets.RegisterBucket(r.plugin.ctx, req.Name, cfg); err != nil {
		resp.Success = false
		resp.Message = "Failed to register bucket: " + err.Error()
		return err
	}

	resp.Success = true
	resp.Message = "Bucket registered successfully"
	return nil
}

// ListBuckets lists all registered buckets
func (r *rpc) ListBuckets(req *ListBucketsRequest, resp *ListBucketsResponse) error {
	resp.Buckets = r.plugin.buckets.ListBuckets()
	resp.Default = r.plugin.buckets.GetDefaultBucketName()
	return nil
}

// Write uploads a file to S3
func (r *rpc) Write(req *WriteRequest, resp *WriteResponse) error {
	return r.plugin.operations.Write(r.plugin.ctx, req, resp)
}

// Read downloads a file from S3
func (r *rpc) Read(req *ReadRequest, resp *ReadResponse) error {
	return r.plugin.operations.Read(r.plugin.ctx, req, resp)
}

// Exists checks if a file exists in S3
func (r *rpc) Exists(req *ExistsRequest, resp *ExistsResponse) error {
	return r.plugin.operations.Exists(r.plugin.ctx, req, resp)
}

// Delete deletes a file from S3
func (r *rpc) Delete(req *DeleteRequest, resp *DeleteResponse) error {
	return r.plugin.operations.Delete(r.plugin.ctx, req, resp)
}

// Copy copies a file within or between buckets
func (r *rpc) Copy(req *CopyRequest, resp *CopyResponse) error {
	return r.plugin.operations.Copy(r.plugin.ctx, req, resp)
}

// Move moves a file within or between buckets
func (r *rpc) Move(req *MoveRequest, resp *MoveResponse) error {
	return r.plugin.operations.Move(r.plugin.ctx, req, resp)
}

// GetMetadata retrieves file metadata
func (r *rpc) GetMetadata(req *GetMetadataRequest, resp *GetMetadataResponse) error {
	return r.plugin.operations.GetMetadata(r.plugin.ctx, req, resp)
}

// SetVisibility changes file visibility (ACL)
func (r *rpc) SetVisibility(req *SetVisibilityRequest, resp *SetVisibilityResponse) error {
	return r.plugin.operations.SetVisibility(r.plugin.ctx, req, resp)
}

// GetPublicURL generates a public or presigned URL for a file
func (r *rpc) GetPublicURL(req *GetPublicURLRequest, resp *GetPublicURLResponse) error {
	return r.plugin.operations.GetPublicURL(r.plugin.ctx, req, resp)
}
