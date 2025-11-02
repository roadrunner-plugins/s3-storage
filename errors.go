package s3

// ErrorCode represents structured error codes for S3 operations
type ErrorCode string

const (
	// ErrBucketNotFound indicates the requested bucket doesn't exist
	ErrBucketNotFound ErrorCode = "BUCKET_NOT_FOUND"

	// ErrFileNotFound indicates the requested file doesn't exist
	ErrFileNotFound ErrorCode = "FILE_NOT_FOUND"

	// ErrInvalidConfig indicates invalid configuration
	ErrInvalidConfig ErrorCode = "INVALID_CONFIG"

	// ErrS3Operation indicates an S3 operation failed
	ErrS3Operation ErrorCode = "S3_OPERATION_FAILED"

	// ErrPermissionDenied indicates insufficient permissions
	ErrPermissionDenied ErrorCode = "PERMISSION_DENIED"

	// ErrInvalidPathname indicates invalid file path
	ErrInvalidPathname ErrorCode = "INVALID_PATHNAME"

	// ErrBucketAlreadyExists indicates bucket is already registered
	ErrBucketAlreadyExists ErrorCode = "BUCKET_ALREADY_EXISTS"

	// ErrInvalidVisibility indicates invalid visibility value
	ErrInvalidVisibility ErrorCode = "INVALID_VISIBILITY"

	// ErrOperationTimeout indicates operation exceeded timeout
	ErrOperationTimeout ErrorCode = "OPERATION_TIMEOUT"
)

// S3Error represents a structured error returned to PHP
type S3Error struct {
	// Code is the error code
	Code ErrorCode `json:"code"`

	// Message is the human-readable error message
	Message string `json:"message"`

	// Details contains additional error context (optional)
	Details string `json:"details,omitempty"`
}

// Error implements the error interface
func (e *S3Error) Error() string {
	if e.Details != "" {
		return string(e.Code) + ": " + e.Message + " (" + e.Details + ")"
	}
	return string(e.Code) + ": " + e.Message
}

// NewS3Error creates a new S3Error
func NewS3Error(code ErrorCode, message string, details string) *S3Error {
	return &S3Error{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// NewBucketNotFoundError creates a bucket not found error
func NewBucketNotFoundError(bucketName string) *S3Error {
	return NewS3Error(
		ErrBucketNotFound,
		"Bucket not found",
		"bucket: "+bucketName,
	)
}

// NewFileNotFoundError creates a file not found error
func NewFileNotFoundError(pathname string) *S3Error {
	return NewS3Error(
		ErrFileNotFound,
		"File not found",
		"pathname: "+pathname,
	)
}

// NewInvalidConfigError creates an invalid config error
func NewInvalidConfigError(reason string) *S3Error {
	return NewS3Error(
		ErrInvalidConfig,
		"Invalid configuration",
		reason,
	)
}

// NewS3OperationError creates an S3 operation error
func NewS3OperationError(operation string, err error) *S3Error {
	return NewS3Error(
		ErrS3Operation,
		"S3 operation failed: "+operation,
		err.Error(),
	)
}

// NewPermissionDeniedError creates a permission denied error
func NewPermissionDeniedError(operation string) *S3Error {
	return NewS3Error(
		ErrPermissionDenied,
		"Permission denied",
		"operation: "+operation,
	)
}

// NewInvalidPathnameError creates an invalid pathname error
func NewInvalidPathnameError(pathname string, reason string) *S3Error {
	return NewS3Error(
		ErrInvalidPathname,
		"Invalid pathname",
		"pathname: "+pathname+", reason: "+reason,
	)
}
