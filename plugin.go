package s3

import (
	"context"
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/roadrunner-server/endure/v2/dep"
	"github.com/roadrunner-server/errors"
	"go.uber.org/zap"
)

const (
	// PluginName is the name of the S3 plugin
	PluginName = "s3"
)

// Plugin represents the main S3 storage plugin structure
type Plugin struct {
	// Configuration provider
	cfg Configurer

	// Logger
	log *zap.Logger

	// Bucket manager holds all registered buckets
	buckets *BucketManager

	// Operations handler for S3 operations
	operations *Operations

	// Metrics exporter for Prometheus integration
	metrics *metricsExporter

	// Context for graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc

	// WaitGroup for tracking ongoing operations
	wg sync.WaitGroup

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// Configurer interface for configuration loading
type Configurer interface {
	// UnmarshalKey takes a single key and unmarshal it into a Struct
	UnmarshalKey(name string, out interface{}) error
	// Has checks if config section exists
	Has(name string) bool
}

// Logger interface for logging
type Logger interface {
	NamedLogger(name string) *zap.Logger
}

// Init initializes the plugin with dependencies
func (p *Plugin) Init(cfg Configurer, log Logger) error {
	const op = errors.Op("s3_plugin_init")
	if !cfg.Has(PluginName) {
		return errors.E(op, errors.Disabled)
	}

	p.cfg = cfg
	p.log = log.NamedLogger(PluginName)
	p.ctx, p.cancel = context.WithCancel(context.Background())

	// Initialize metrics exporter
	p.metrics = newMetricsExporter()

	// Initialize bucket manager
	p.buckets = NewBucketManager(p.log)

	// Initialize operations handler
	p.operations = NewOperations(p, p.log)

	// Load static configuration from .rr.yaml
	var config Config
	if err := cfg.UnmarshalKey(PluginName, &config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Set server configurations in bucket manager
	p.buckets.SetServers(config.Servers)

	// Register buckets from static configuration
	for name, bucketCfg := range config.Buckets {
		p.log.Debug("registering bucket from config",
			zap.String("name", name),
			zap.String("bucket", bucketCfg.Bucket),
			zap.String("server", bucketCfg.Server),
		)

		if err := p.buckets.RegisterBucket(p.ctx, name, bucketCfg); err != nil {
			// Log error but don't fail initialization - allow other buckets to work
			p.log.Error("failed to register bucket",
				zap.String("name", name),
				zap.Error(err),
			)
			continue
		}
	}

	// Set default bucket if specified
	if config.Default != "" {
		if err := p.buckets.SetDefault(config.Default); err != nil {
			p.log.Warn("failed to set default bucket",
				zap.String("default", config.Default),
				zap.Error(err),
			)
		}
	}

	p.log.Info("S3 plugin initialized",
		zap.Int("servers", len(config.Servers)),
		zap.Int("buckets", len(config.Buckets)),
		zap.String("default", config.Default),
	)

	return nil
}

// Serve starts the plugin (long-running service)
func (p *Plugin) Serve() chan error {
	errCh := make(chan error, 1)

	// This plugin doesn't have background workers, but implements Service interface
	// for proper lifecycle management
	p.log.Debug("S3 plugin serving")

	return errCh
}

// Stop gracefully stops the plugin
func (p *Plugin) Stop(ctx context.Context) error {
	p.log.Debug("stopping S3 plugin")

	// Cancel all ongoing operations
	p.cancel()

	// Wait for all operations to complete or timeout
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		p.log.Debug("all S3 operations completed")
	case <-ctx.Done():
		p.log.Warn("shutdown timeout reached, forcing stop")
	}

	// Close all S3 clients
	if err := p.buckets.CloseAll(); err != nil {
		p.log.Error("error closing bucket clients", zap.Error(err))
		return err
	}

	p.log.Debug("S3 plugin stopped")
	return nil
}

// Name returns the plugin name
func (p *Plugin) Name() string {
	return PluginName
}

// Weight returns plugin weight for dependency resolution
// Higher weight = initialized later
func (p *Plugin) Weight() uint {
	return 10
}

// RPC returns the RPC interface exposed to PHP
func (p *Plugin) RPC() interface{} {
	return &rpc{
		plugin: p,
		log:    p.log,
	}
}

// Collects declares the plugin's dependencies
func (p *Plugin) Collects() []*dep.In {
	return []*dep.In{
		dep.Fits(func(pp any) {
			p.cfg = pp.(Configurer)
		}, (*Configurer)(nil)),
		dep.Fits(func(pp any) {
			p.log = pp.(Logger).NamedLogger(PluginName)
		}, (*Logger)(nil)),
	}
}

// GetBucketManager returns the bucket manager (for internal use)
func (p *Plugin) GetBucketManager() *BucketManager {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.buckets
}

// GetContext returns the plugin context
func (p *Plugin) GetContext() context.Context {
	return p.ctx
}

// TrackOperation adds an operation to the wait group
func (p *Plugin) TrackOperation() {
	p.wg.Add(1)
}

// CompleteOperation marks an operation as complete
func (p *Plugin) CompleteOperation() {
	p.wg.Done()
}

// MetricsCollector implements the StatProvider interface for Prometheus metrics integration
// This method is called by the metrics plugin during its Serve phase to register all collectors
func (p *Plugin) MetricsCollector() []prometheus.Collector {
	if p.metrics == nil {
		return nil
	}
	return p.metrics.getCollectors()
}
