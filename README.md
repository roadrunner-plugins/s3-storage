# RoadRunner S3 Storage Plugin

A high-performance Go plugin for RoadRunner that provides S3-compatible object storage operations via RPC interface.

## Features

- ✅ **Multiple Bucket Support**: Configure multiple S3 buckets (AWS S3, MinIO, DigitalOcean Spaces, etc.)
- ✅ **Dynamic Configuration**: Register buckets at runtime via RPC
- ✅ **Full S3 Operations**: Upload, download, copy, move, delete, metadata operations
- ✅ **Concurrent Operations**: Built-in goroutine management and connection pooling
- ✅ **Public URL Generation**: Generate public and presigned URLs
- ✅ **Visibility Control**: Manage file ACLs (public/private)
- ✅ **Large File Support**: Multipart upload for files > 5MB
- ✅ **Graceful Shutdown**: Proper context cancellation and operation tracking

## Installation

```bash
go get github.com/roadrunner-server/s3-plugin
```

## Configuration

### Basic Configuration (.rr.yaml)

```yaml
s3:
  # Default bucket to use when none specified
  default: uploads

  # Pre-configured buckets at startup
  buckets:
    uploads:
      region: us-east-1
      endpoint: https://s3.amazonaws.com  # Optional, for custom S3-compatible services
      bucket: my-uploads-bucket
      prefix: "uploads/"  # Optional path prefix
      credentials:
        key: ${AWS_ACCESS_KEY_ID}
        secret: ${AWS_SECRET_ACCESS_KEY}
        token: ${AWS_SESSION_TOKEN}  # Optional for temporary credentials
      visibility: public  # or "private"
      max_concurrent_operations: 100  # Optional, default: 100
      part_size: 5242880  # Optional, default: 5MB (for multipart uploads)
      concurrency: 5  # Optional, default: 5 (goroutines for multipart)

    private-docs:
      region: eu-west-1
      bucket: company-documents
      credentials:
        key: ${AWS_ACCESS_KEY_ID}
        secret: ${AWS_SECRET_ACCESS_KEY}
      visibility: private
```

### MinIO Configuration Example

```yaml
s3:
  default: minio-storage
  buckets:
    minio-storage:
      region: us-east-1  # MinIO requires a region value
      endpoint: http://localhost:9000
      bucket: my-bucket
      credentials:
        key: minioadmin
        secret: minioadmin
      visibility: public
```

## PHP Usage

### Basic Operations

```php
use Spiral\Goridge\RPC\RPC;

$rpc = new RPC(/* connection */);

// Upload a file
$response = $rpc->call('s3.Write', [
    'bucket' => 'uploads',
    'pathname' => 'images/photo.jpg',
    'content' => base64_encode(file_get_contents('photo.jpg')),
    'visibility' => 'public'  // Optional
]);
// Returns: ['success' => true, 'pathname' => '...', 'size' => 12345, 'last_modified' => 1234567890]

// Download a file
$response = $rpc->call('s3.Read', [
    'bucket' => 'uploads',
    'pathname' => 'images/photo.jpg'
]);
// Returns: ['content' => base64_data, 'size' => 12345, 'mime_type' => 'image/jpeg', 'last_modified' => 1234567890]

// Check if file exists
$response = $rpc->call('s3.Exists', [
    'bucket' => 'uploads',
    'pathname' => 'images/photo.jpg'
]);
// Returns: ['exists' => true]

// Delete a file
$response = $rpc->call('s3.Delete', [
    'bucket' => 'uploads',
    'pathname' => 'images/photo.jpg'
]);
// Returns: ['success' => true]

// Get file metadata
$response = $rpc->call('s3.GetMetadata', [
    'bucket' => 'uploads',
    'pathname' => 'images/photo.jpg'
]);
// Returns: ['size' => 12345, 'mime_type' => 'image/jpeg', 'last_modified' => 1234567890, 'visibility' => 'public']

// Get public URL (permanent)
$response = $rpc->call('s3.GetPublicURL', [
    'bucket' => 'uploads',
    'pathname' => 'images/photo.jpg',
    'expires_in' => 0  // 0 for permanent URL
]);
// Returns: ['url' => 'https://...']

// Get presigned URL (expires in 1 hour)
$response = $rpc->call('s3.GetPublicURL', [
    'bucket' => 'uploads',
    'pathname' => 'images/photo.jpg',
    'expires_in' => 3600  // seconds
]);
// Returns: ['url' => 'https://...', 'expires_at' => 1234567890]
```

### Advanced Operations

```php
// Copy file within same bucket
$response = $rpc->call('s3.Copy', [
    'source_bucket' => 'uploads',
    'source_pathname' => 'images/photo.jpg',
    'dest_bucket' => 'uploads',
    'dest_pathname' => 'images/photo-copy.jpg',
    'visibility' => 'private'  // Optional
]);

// Move file between buckets
$response = $rpc->call('s3.Move', [
    'source_bucket' => 'uploads',
    'source_pathname' => 'temp/photo.jpg',
    'dest_bucket' => 'private-docs',
    'dest_pathname' => 'archive/photo.jpg'
]);

// Change file visibility
$response = $rpc->call('s3.SetVisibility', [
    'bucket' => 'uploads',
    'pathname' => 'images/photo.jpg',
    'visibility' => 'private'  // or 'public'
]);
```

### Dynamic Bucket Registration

```php
// Register a new bucket at runtime
$response = $rpc->call('s3.RegisterBucket', [
    'name' => 'dynamic-bucket',
    'region' => 'us-west-2',
    'endpoint' => 'https://s3.amazonaws.com',
    'bucket' => 'my-new-bucket',
    'prefix' => 'files/',
    'credentials' => [
        'key' => 'AKIAIOSFODNN7EXAMPLE',
        'secret' => 'wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY'
    ],
    'visibility' => 'public'
]);
// Returns: ['success' => true, 'message' => 'Bucket registered successfully']

// List all registered buckets
$response = $rpc->call('s3.ListBuckets', []);
// Returns: ['buckets' => ['uploads', 'private-docs', 'dynamic-bucket'], 'default' => 'uploads']
```

## Architecture

### Plugin Structure

```
s3/
├── plugin.go           # Main plugin with DI and lifecycle management
├── config.go           # Configuration structures and validation
├── bucket_manager.go   # Bucket registration and S3 client management
├── operations.go       # All S3 file operations implementation
├── rpc.go             # RPC interface definitions and handlers
├── errors.go          # Structured error types
└── go.mod             # Go module dependencies
```

### Concurrency Model

- **Per-Bucket Semaphores**: Limits concurrent operations per bucket (default: 100)
- **AWS SDK Connection Pooling**: Built-in HTTP connection reuse
- **Goroutine Tracking**: WaitGroup for graceful shutdown
- **Context Propagation**: All operations support cancellation

### Performance Characteristics

- **Small Files (< 1MB)**: Direct upload, 100+ ops/sec per bucket
- **Large Files (> 5MB)**: Multipart upload with configurable concurrency
- **Memory Usage**: Streams large files, minimal memory footprint
- **Concurrent Operations**: Supports 50+ simultaneous operations per bucket

## Error Handling

All RPC methods return structured errors:

```go
type S3Error struct {
    Code    string `json:"code"`     // Error code (e.g., "BUCKET_NOT_FOUND")
    Message string `json:"message"`  // Human-readable message
    Details string `json:"details"`  // Additional context
}
```

### Error Codes

| Code                    | Description                    |
|-------------------------|--------------------------------|
| `BUCKET_NOT_FOUND`      | Requested bucket doesn't exist |
| `FILE_NOT_FOUND`        | Requested file doesn't exist   |
| `INVALID_CONFIG`        | Invalid bucket configuration   |
| `S3_OPERATION_FAILED`   | S3 operation failed            |
| `PERMISSION_DENIED`     | Insufficient permissions       |
| `INVALID_PATHNAME`      | Invalid file path              |
| `BUCKET_ALREADY_EXISTS` | Bucket already registered      |
| `INVALID_VISIBILITY`    | Invalid visibility value       |

## Testing

```bash
# Run unit tests
go test ./...

# Run with coverage
go test -cover ./...

# Run integration tests (requires S3 credentials)
export AWS_ACCESS_KEY_ID=your_key
export AWS_SECRET_ACCESS_KEY=your_secret
go test -tags=integration ./...
```

## Integration with RoadRunner

### Plugin Registration

```go
// In your RoadRunner server
import (
    s3plugin "github.com/roadrunner-server/s3-plugin"
)

func main() {
    container := endure.New(/* config */)
    
    // Register S3 plugin
    container.Register(&s3plugin.Plugin{})
    
    // ... other plugins
    
    container.Init()
    container.Serve()
}
```

## Security Best Practices

1. **Credentials Management**
    - Use environment variables for secrets
    - Never commit credentials to version control
    - Rotate credentials regularly

2. **Access Control**
    - Use IAM roles when running on AWS infrastructure
    - Apply principle of least privilege
    - Use bucket policies for additional security

3. **Network Security**
    - Use HTTPS endpoints for production
    - Consider VPC endpoints for AWS S3
    - Implement proper CORS policies if needed

## Troubleshooting

### Common Issues

**"Bucket not found" error**

- Verify bucket name is correct in configuration
- Check that bucket is registered (use `ListBuckets` RPC)
- Ensure credentials have access to the bucket

**"Permission denied" error**

- Verify AWS credentials are correct
- Check IAM policy allows required S3 operations
- Ensure bucket policy doesn't block access

**Slow upload performance**

- Increase `concurrency` setting for multipart uploads
- Adjust `part_size` (larger parts = fewer API calls)
- Check `max_concurrent_operations` limit

**Memory usage too high**

- Files are streamed for large uploads
- Check if multiple large files are processed simultaneously
- Adjust `max_concurrent_operations` to limit parallelism

## Contributing

Contributions are welcome! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

MIT License - see LICENSE file for details

## Support

- **Documentation**: [RoadRunner Docs](https://docs.roadrunner.dev)
- **Issues**: [GitHub Issues](https://github.com/roadrunner-server/s3-plugin/issues)
- **Community**: [Discord](https://discord.gg/roadrunner)
