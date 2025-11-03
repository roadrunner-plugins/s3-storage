package s3

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"go.uber.org/zap"
)

// Operations handles all S3 file operations
type Operations struct {
	plugin *Plugin
	log    *zap.Logger
}

// NewOperations creates a new Operations instance
func NewOperations(plugin *Plugin, log *zap.Logger) *Operations {
	return &Operations{
		plugin: plugin,
		log:    log,
	}
}

// Write uploads a file to S3
func (o *Operations) Write(ctx context.Context, req *WriteRequest, resp *WriteResponse) error {
	// Track operation for graceful shutdown
	o.plugin.TrackOperation()
	defer o.plugin.CompleteOperation()

	start := time.Now()

	// Validate request
	if err := o.validatePathname(req.Pathname); err != nil {
		return err
	}

	// Get bucket
	bucket, err := o.plugin.buckets.GetBucket(req.Bucket)
	if err != nil {
		return NewBucketNotFoundError(req.Bucket)
	}

	// Acquire semaphore
	bucket.Acquire()
	defer bucket.Release()

	// Determine visibility
	visibility := req.Visibility
	if visibility == "" {
		visibility = bucket.GetVisibility()
	}

	// Get full S3 key
	key := bucket.GetFullPath(req.Pathname)

	// Detect content type
	contentType := o.detectContentType(req.Pathname, req.Content)

	// Prepare upload input
	putInput := &s3.PutObjectInput{
		Bucket:      aws.String(bucket.Config.Bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(req.Content),
		ACL:         types.ObjectCannedACL(visibility),
		ContentType: aws.String(contentType),
	}

	// Add custom metadata if provided
	if len(req.Config) > 0 {
		metadata := make(map[string]string)
		for k, v := range req.Config {
			metadata[k] = v
		}
		putInput.Metadata = metadata
	}

	// Use upload manager for better performance with large files
	uploader := manager.NewUploader(bucket.Client, func(u *manager.Uploader) {
		u.PartSize = bucket.Config.PartSize
		u.Concurrency = bucket.Config.Concurrency
	})

	// Upload file
	result, err := uploader.Upload(ctx, putInput)
	if err != nil {
		o.log.Error("failed to upload file",
			zap.String("bucket", req.Bucket),
			zap.String("pathname", req.Pathname),
			zap.Error(err),
		)
		return NewS3OperationError("upload", err)
	}

	// Get metadata for response
	headResult, err := bucket.Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket.Config.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		o.log.Warn("failed to get object metadata after upload",
			zap.String("bucket", req.Bucket),
			zap.String("pathname", req.Pathname),
			zap.Error(err),
		)
		// Don't fail the operation, just return without metadata
		resp.Success = true
		resp.Pathname = req.Pathname
		resp.Size = int64(len(req.Content))
		resp.LastModified = time.Now().Unix()
		return nil
	}

	resp.Success = true
	resp.Pathname = req.Pathname
	resp.Size = *headResult.ContentLength
	resp.LastModified = headResult.LastModified.Unix()

	o.log.Debug("file uploaded successfully",
		zap.String("bucket", req.Bucket),
		zap.String("pathname", req.Pathname),
		zap.Int64("size", resp.Size),
		zap.Duration("duration", time.Since(start)),
	)

	_ = result // Use result to avoid unused variable warning

	return nil
}

// Read downloads a file from S3
func (o *Operations) Read(ctx context.Context, req *ReadRequest, resp *ReadResponse) error {
	o.plugin.TrackOperation()
	defer o.plugin.CompleteOperation()

	start := time.Now()

	// Validate request
	if err := o.validatePathname(req.Pathname); err != nil {
		return err
	}

	// Get bucket
	bucket, err := o.plugin.buckets.GetBucket(req.Bucket)
	if err != nil {
		return NewBucketNotFoundError(req.Bucket)
	}

	bucket.Acquire()
	defer bucket.Release()

	// Get full S3 key
	key := bucket.GetFullPath(req.Pathname)

	// Download file
	result, err := bucket.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket.Config.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var nsk *types.NoSuchKey
		if errors.As(err, &nsk) {
			return NewFileNotFoundError(req.Pathname)
		}
		o.log.Error("failed to download file",
			zap.String("bucket", req.Bucket),
			zap.String("pathname", req.Pathname),
			zap.Error(err),
		)
		return NewS3OperationError("download", err)
	}
	defer result.Body.Close()

	// Read content
	content, err := io.ReadAll(result.Body)
	if err != nil {
		o.log.Error("failed to read file content",
			zap.String("bucket", req.Bucket),
			zap.String("pathname", req.Pathname),
			zap.Error(err),
		)
		return NewS3OperationError("read content", err)
	}

	resp.Content = content
	resp.Size = *result.ContentLength
	resp.MimeType = *result.ContentType
	resp.LastModified = result.LastModified.Unix()

	o.log.Debug("file downloaded successfully",
		zap.String("bucket", req.Bucket),
		zap.String("pathname", req.Pathname),
		zap.Int64("size", resp.Size),
		zap.Duration("duration", time.Since(start)),
	)

	return nil
}

// Exists checks if a file exists in S3
func (o *Operations) Exists(ctx context.Context, req *ExistsRequest, resp *ExistsResponse) error {
	o.plugin.TrackOperation()
	defer o.plugin.CompleteOperation()

	// Validate request
	if err := o.validatePathname(req.Pathname); err != nil {
		return err
	}

	// Get bucket
	bucket, err := o.plugin.buckets.GetBucket(req.Bucket)
	if err != nil {
		return NewBucketNotFoundError(req.Bucket)
	}

	bucket.Acquire()
	defer bucket.Release()

	// Get full S3 key
	key := bucket.GetFullPath(req.Pathname)

	// Check if object exists
	_, err = bucket.Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket.Config.Bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		var nsk *types.NoSuchKey
		var nf *types.NotFound
		if errors.As(err, &nsk) || errors.As(err, &nf) {
			resp.Exists = false
			return nil
		}
		// Other errors should be returned
		o.log.Error("failed to check file existence",
			zap.String("bucket", req.Bucket),
			zap.String("pathname", req.Pathname),
			zap.Error(err),
		)
		return NewS3OperationError("head object", err)
	}

	resp.Exists = true
	return nil
}

// Delete deletes a file from S3
func (o *Operations) Delete(ctx context.Context, req *DeleteRequest, resp *DeleteResponse) error {
	o.plugin.TrackOperation()
	defer o.plugin.CompleteOperation()

	// Validate request
	if err := o.validatePathname(req.Pathname); err != nil {
		return err
	}

	// Get bucket
	bucket, err := o.plugin.buckets.GetBucket(req.Bucket)
	if err != nil {
		return NewBucketNotFoundError(req.Bucket)
	}

	bucket.Acquire()
	defer bucket.Release()

	// Get full S3 key
	key := bucket.GetFullPath(req.Pathname)

	// Delete object
	_, err = bucket.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket.Config.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		o.log.Error("failed to delete file",
			zap.String("bucket", req.Bucket),
			zap.String("pathname", req.Pathname),
			zap.Error(err),
		)
		return NewS3OperationError("delete", err)
	}

	resp.Success = true

	o.log.Debug("file deleted successfully",
		zap.String("bucket", req.Bucket),
		zap.String("pathname", req.Pathname),
	)

	return nil
}

// Copy copies a file within or between buckets
func (o *Operations) Copy(ctx context.Context, req *CopyRequest, resp *CopyResponse) error {
	o.plugin.TrackOperation()
	defer o.plugin.CompleteOperation()

	start := time.Now()

	// Validate request
	if err := o.validatePathname(req.SourcePathname); err != nil {
		return err
	}
	if err := o.validatePathname(req.DestPathname); err != nil {
		return err
	}

	// Get source bucket
	sourceBucket, err := o.plugin.buckets.GetBucket(req.SourceBucket)
	if err != nil {
		return NewBucketNotFoundError(req.SourceBucket)
	}

	// Get destination bucket
	destBucket, err := o.plugin.buckets.GetBucket(req.DestBucket)
	if err != nil {
		return NewBucketNotFoundError(req.DestBucket)
	}

	// Acquire semaphores
	sourceBucket.Acquire()
	defer sourceBucket.Release()
	if req.SourceBucket != req.DestBucket {
		destBucket.Acquire()
		defer destBucket.Release()
	}

	// Get full S3 keys
	sourceKey := sourceBucket.GetFullPath(req.SourcePathname)
	destKey := destBucket.GetFullPath(req.DestPathname)

	// Prepare copy source
	copySource := fmt.Sprintf("%s/%s", sourceBucket.Config.Bucket, sourceKey)

	// Determine visibility
	visibility := req.Visibility
	if visibility == "" {
		visibility = destBucket.GetVisibility()
	}

	// Copy object
	_, err = destBucket.Client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(destBucket.Config.Bucket),
		Key:        aws.String(destKey),
		CopySource: aws.String(copySource),
		ACL:        types.ObjectCannedACL(visibility),
	})
	if err != nil {
		o.log.Error("failed to copy file",
			zap.String("source_bucket", req.SourceBucket),
			zap.String("source_pathname", req.SourcePathname),
			zap.String("dest_bucket", req.DestBucket),
			zap.String("dest_pathname", req.DestPathname),
			zap.Error(err),
		)
		return NewS3OperationError("copy", err)
	}

	// Get metadata for response
	headResult, err := destBucket.Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(destBucket.Config.Bucket),
		Key:    aws.String(destKey),
	})
	if err == nil {
		resp.Size = *headResult.ContentLength
		resp.LastModified = headResult.LastModified.Unix()
	}

	resp.Success = true
	resp.Pathname = req.DestPathname

	o.log.Debug("file copied successfully",
		zap.String("source_bucket", req.SourceBucket),
		zap.String("source_pathname", req.SourcePathname),
		zap.String("dest_bucket", req.DestBucket),
		zap.String("dest_pathname", req.DestPathname),
		zap.Duration("duration", time.Since(start)),
	)

	return nil
}

// Move moves a file within or between buckets (copy + delete)
func (o *Operations) Move(ctx context.Context, req *MoveRequest, resp *MoveResponse) error {
	// First, copy the file
	copyReq := &CopyRequest{
		SourceBucket:   req.SourceBucket,
		SourcePathname: req.SourcePathname,
		DestBucket:     req.DestBucket,
		DestPathname:   req.DestPathname,
		Config:         req.Config,
		Visibility:     req.Visibility,
	}
	copyResp := &CopyResponse{}

	if err := o.Copy(ctx, copyReq, copyResp); err != nil {
		return err
	}

	// Then delete the source file
	deleteReq := &DeleteRequest{
		Bucket:   req.SourceBucket,
		Pathname: req.SourcePathname,
	}
	deleteResp := &DeleteResponse{}

	if err := o.Delete(ctx, deleteReq, deleteResp); err != nil {
		o.log.Error("failed to delete source file after copy",
			zap.String("bucket", req.SourceBucket),
			zap.String("pathname", req.SourcePathname),
			zap.Error(err),
		)
		// Return error but note that copy succeeded
		return fmt.Errorf("copy succeeded but delete failed: %w", err)
	}

	resp.Success = true
	resp.Pathname = copyResp.Pathname
	resp.Size = copyResp.Size
	resp.LastModified = copyResp.LastModified

	return nil
}

// GetMetadata retrieves file metadata
func (o *Operations) GetMetadata(ctx context.Context, req *GetMetadataRequest, resp *GetMetadataResponse) error {
	o.plugin.TrackOperation()
	defer o.plugin.CompleteOperation()

	// Validate request
	if err := o.validatePathname(req.Pathname); err != nil {
		return err
	}

	// Get bucket
	bucket, err := o.plugin.buckets.GetBucket(req.Bucket)
	if err != nil {
		return NewBucketNotFoundError(req.Bucket)
	}

	bucket.Acquire()
	defer bucket.Release()

	// Get full S3 key
	key := bucket.GetFullPath(req.Pathname)

	// Get object metadata
	result, err := bucket.Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket.Config.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var nsk *types.NoSuchKey
		var nf *types.NotFound
		if errors.As(err, &nsk) || errors.As(err, &nf) {
			return NewFileNotFoundError(req.Pathname)
		}
		o.log.Error("failed to get file metadata",
			zap.String("bucket", req.Bucket),
			zap.String("pathname", req.Pathname),
			zap.Error(err),
		)
		return NewS3OperationError("head object", err)
	}

	resp.Size = *result.ContentLength
	if result.ContentType != nil {
		resp.MimeType = *result.ContentType
	}
	resp.LastModified = result.LastModified.Unix()
	if result.ETag != nil {
		resp.ETag = *result.ETag
	}

	// Determine visibility from ACL (if available)
	resp.Visibility = "private" // Default

	return nil
}

// SetVisibility changes file visibility (ACL)
func (o *Operations) SetVisibility(ctx context.Context, req *SetVisibilityRequest, resp *SetVisibilityResponse) error {
	o.plugin.TrackOperation()
	defer o.plugin.CompleteOperation()

	// Validate request
	if err := o.validatePathname(req.Pathname); err != nil {
		return err
	}

	if req.Visibility != "public" && req.Visibility != "private" {
		return NewS3Error(ErrInvalidVisibility, "visibility must be 'public' or 'private'", req.Visibility)
	}

	// Get bucket
	bucket, err := o.plugin.buckets.GetBucket(req.Bucket)
	if err != nil {
		return NewBucketNotFoundError(req.Bucket)
	}

	bucket.Acquire()
	defer bucket.Release()

	// Get full S3 key
	key := bucket.GetFullPath(req.Pathname)

	// Map visibility to ACL
	acl := types.ObjectCannedACLPrivate
	if req.Visibility == "public" {
		acl = types.ObjectCannedACLPublicRead
	}

	// Set ACL
	_, err = bucket.Client.PutObjectAcl(ctx, &s3.PutObjectAclInput{
		Bucket: aws.String(bucket.Config.Bucket),
		Key:    aws.String(key),
		ACL:    acl,
	})
	if err != nil {
		o.log.Error("failed to set file visibility",
			zap.String("bucket", req.Bucket),
			zap.String("pathname", req.Pathname),
			zap.String("visibility", req.Visibility),
			zap.Error(err),
		)
		return NewS3OperationError("put object acl", err)
	}

	resp.Success = true

	o.log.Debug("file visibility changed",
		zap.String("bucket", req.Bucket),
		zap.String("pathname", req.Pathname),
		zap.String("visibility", req.Visibility),
	)

	return nil
}

// GetPublicURL generates a public or presigned URL for a file
func (o *Operations) GetPublicURL(ctx context.Context, req *GetPublicURLRequest, resp *GetPublicURLResponse) error {
	o.plugin.TrackOperation()
	defer o.plugin.CompleteOperation()

	// Validate request
	if err := o.validatePathname(req.Pathname); err != nil {
		return err
	}

	// Get bucket
	bucket, err := o.plugin.buckets.GetBucket(req.Bucket)
	if err != nil {
		return NewBucketNotFoundError(req.Bucket)
	}

	// Get full S3 key
	key := bucket.GetFullPath(req.Pathname)

	// If no expiration, generate permanent public URL
	if req.ExpiresIn == 0 {
		// Generate public URL (assuming public-read ACL)
		endpoint := bucket.Config.Endpoint
		if endpoint == "" {
			endpoint = fmt.Sprintf("https://s3.%s.amazonaws.com", bucket.Config.Region)
		}
		resp.URL = fmt.Sprintf("%s/%s/%s", endpoint, bucket.Config.Bucket, key)
		return nil
	}

	// Generate presigned URL
	presignClient := s3.NewPresignClient(bucket.Client)
	presignResult, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket.Config.Bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(req.ExpiresIn) * time.Second
	})
	if err != nil {
		o.log.Error("failed to generate presigned URL",
			zap.String("bucket", req.Bucket),
			zap.String("pathname", req.Pathname),
			zap.Error(err),
		)
		return NewS3OperationError("presign get object", err)
	}

	resp.URL = presignResult.URL
	resp.ExpiresAt = time.Now().Add(time.Duration(req.ExpiresIn) * time.Second).Unix()

	return nil
}

// ListObjects lists objects in a bucket with optional filtering and pagination
func (o *Operations) ListObjects(ctx context.Context, req *ListObjectsRequest, resp *ListObjectsResponse) error {
	o.plugin.TrackOperation()
	defer o.plugin.CompleteOperation()

	start := time.Now()

	// Get bucket
	bucket, err := o.plugin.buckets.GetBucket(req.Bucket)
	if err != nil {
		return NewBucketNotFoundError(req.Bucket)
	}

	bucket.Acquire()
	defer bucket.Release()

	// Set default max keys if not specified
	maxKeys := req.MaxKeys
	if maxKeys <= 0 {
		maxKeys = 1000
	}

	// Prepare prefix - include bucket prefix if configured
	prefix := bucket.GetFullPath(req.Prefix)

	// Prepare list objects input
	input := &s3.ListObjectsV2Input{
		Bucket:  aws.String(bucket.Config.Bucket),
		MaxKeys: aws.Int32(maxKeys),
	}

	// Add optional parameters
	if prefix != "" {
		input.Prefix = aws.String(prefix)
	}

	if req.Delimiter != "" {
		input.Delimiter = aws.String(req.Delimiter)
	}

	if req.ContinuationToken != "" {
		input.ContinuationToken = aws.String(req.ContinuationToken)
	}

	// List objects
	result, err := bucket.Client.ListObjectsV2(ctx, input)
	if err != nil {
		o.log.Error("failed to list objects",
			zap.String("bucket", req.Bucket),
			zap.String("prefix", req.Prefix),
			zap.Error(err),
		)
		return NewS3OperationError("list objects", err)
	}

	// Convert results to response format
	resp.Objects = make([]ObjectInfo, 0, len(result.Contents))
	for _, obj := range result.Contents {
		// Remove bucket prefix from key if present
		key := *obj.Key
		if bucket.Config.Prefix != "" && strings.HasPrefix(key, bucket.Config.Prefix) {
			key = strings.TrimPrefix(key, bucket.Config.Prefix)
		}

		objectInfo := ObjectInfo{
			Key:          key,
			Size:         *obj.Size,
			LastModified: obj.LastModified.Unix(),
		}

		if obj.ETag != nil {
			objectInfo.ETag = *obj.ETag
		}

		if obj.StorageClass != "" {
			objectInfo.StorageClass = string(obj.StorageClass)
		}

		resp.Objects = append(resp.Objects, objectInfo)
	}

	// Process common prefixes (directories)
	if len(result.CommonPrefixes) > 0 {
		resp.CommonPrefixes = make([]CommonPrefix, 0, len(result.CommonPrefixes))
		for _, cp := range result.CommonPrefixes {
			prefix := *cp.Prefix
			// Remove bucket prefix if present
			if bucket.Config.Prefix != "" && strings.HasPrefix(prefix, bucket.Config.Prefix) {
				prefix = strings.TrimPrefix(prefix, bucket.Config.Prefix)
			}

			resp.CommonPrefixes = append(resp.CommonPrefixes, CommonPrefix{
				Prefix: prefix,
			})
		}
	}

	// Set pagination info
	resp.IsTruncated = result.IsTruncated != nil && *result.IsTruncated
	if result.NextContinuationToken != nil {
		resp.NextContinuationToken = *result.NextContinuationToken
	}
	resp.KeyCount = *result.KeyCount

	o.log.Debug("objects listed successfully",
		zap.String("bucket", req.Bucket),
		zap.String("prefix", req.Prefix),
		zap.Int32("count", resp.KeyCount),
		zap.Bool("truncated", resp.IsTruncated),
		zap.Duration("duration", time.Since(start)),
	)

	return nil
}

// validatePathname validates a file pathname
func (o *Operations) validatePathname(pathname string) error {
	if pathname == "" {
		return NewInvalidPathnameError(pathname, "pathname cannot be empty")
	}

	if strings.HasPrefix(pathname, "/") {
		return NewInvalidPathnameError(pathname, "pathname cannot start with '/'")
	}

	if strings.Contains(pathname, "..") {
		return NewInvalidPathnameError(pathname, "pathname cannot contain '..'")
	}

	return nil
}

// detectContentType attempts to detect content type from filename and content
func (o *Operations) detectContentType(pathname string, content []byte) string {
	// Simple content type detection based on file extension
	ext := strings.ToLower(pathname[strings.LastIndex(pathname, ".")+1:])

	contentTypes := map[string]string{
		"jpg":  "image/jpeg",
		"jpeg": "image/jpeg",
		"png":  "image/png",
		"gif":  "image/gif",
		"webp": "image/webp",
		"svg":  "image/svg+xml",
		"pdf":  "application/pdf",
		"txt":  "text/plain",
		"html": "text/html",
		"css":  "text/css",
		"js":   "application/javascript",
		"json": "application/json",
		"xml":  "application/xml",
		"zip":  "application/zip",
		"mp4":  "video/mp4",
		"mp3":  "audio/mpeg",
	}

	if contentType, ok := contentTypes[ext]; ok {
		return contentType
	}

	return "application/octet-stream"
}
