package s3

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewS3Error(t *testing.T) {
	err := NewS3Error(ErrBucketNotFound, "test message", "test details")
	assert.Equal(t, ErrBucketNotFound, err.Code)
	assert.Equal(t, "test message", err.Message)
	assert.Equal(t, "test details", err.Details)
}

func TestS3Error_Error(t *testing.T) {
	tests := []struct {
		name    string
		err     *S3Error
		wantMsg string
	}{
		{
			name: "with details",
			err: &S3Error{
				Code:    ErrBucketNotFound,
				Message: "Bucket not found",
				Details: "bucket: test-bucket",
			},
			wantMsg: "BUCKET_NOT_FOUND: Bucket not found (bucket: test-bucket)",
		},
		{
			name: "without details",
			err: &S3Error{
				Code:    ErrFileNotFound,
				Message: "File not found",
				Details: "",
			},
			wantMsg: "FILE_NOT_FOUND: File not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			assert.Equal(t, tt.wantMsg, got)
		})
	}
}

func TestNewBucketNotFoundError(t *testing.T) {
	err := NewBucketNotFoundError("test-bucket")
	assert.Equal(t, ErrBucketNotFound, err.Code)
	assert.Contains(t, err.Error(), "Bucket not found")
	assert.Contains(t, err.Details, "test-bucket")
}

func TestNewFileNotFoundError(t *testing.T) {
	err := NewFileNotFoundError("path/to/file.txt")
	assert.Equal(t, ErrFileNotFound, err.Code)
	assert.Contains(t, err.Error(), "File not found")
	assert.Contains(t, err.Details, "path/to/file.txt")
}

func TestNewInvalidConfigError(t *testing.T) {
	err := NewInvalidConfigError("missing region")
	assert.Equal(t, ErrInvalidConfig, err.Code)
	assert.Contains(t, err.Error(), "Invalid configuration")
	assert.Contains(t, err.Details, "missing region")
}

func TestNewS3OperationError(t *testing.T) {
	originalErr := assert.AnError
	err := NewS3OperationError("upload", originalErr)
	assert.Equal(t, ErrS3Operation, err.Code)
	assert.Contains(t, err.Error(), "S3 operation failed: upload")
	assert.Contains(t, err.Details, originalErr.Error())
}

func TestNewPermissionDeniedError(t *testing.T) {
	err := NewPermissionDeniedError("PutObject")
	assert.Equal(t, ErrPermissionDenied, err.Code)
	assert.Contains(t, err.Error(), "Permission denied")
	assert.Contains(t, err.Details, "PutObject")
}

func TestNewInvalidPathnameError(t *testing.T) {
	err := NewInvalidPathnameError("../file.txt", "contains ..")
	assert.Equal(t, ErrInvalidPathname, err.Code)
	assert.Contains(t, err.Error(), "Invalid pathname")
	assert.Contains(t, err.Details, "../file.txt")
	assert.Contains(t, err.Details, "contains ..")
}

func TestErrorCodes(t *testing.T) {
	// Verify all error codes are defined
	codes := []ErrorCode{
		ErrBucketNotFound,
		ErrFileNotFound,
		ErrInvalidConfig,
		ErrS3Operation,
		ErrPermissionDenied,
		ErrInvalidPathname,
		ErrBucketAlreadyExists,
		ErrInvalidVisibility,
		ErrOperationTimeout,
	}

	// Ensure no empty codes
	for _, code := range codes {
		assert.NotEmpty(t, code)
	}

	// Ensure codes are unique
	seen := make(map[ErrorCode]bool)
	for _, code := range codes {
		assert.False(t, seen[code], "duplicate error code: %s", code)
		seen[code] = true
	}
}
