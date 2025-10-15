package core

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

// MetricType represents different types of metrics
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
	MetricTypeTiming    MetricType = "timing"
)

// Metric represents a single metric measurement
type Metric struct {
	Name      string                 `json:"name"`
	Type      MetricType             `json:"type"`
	Value     float64                `json:"value"`
	Labels    map[string]string      `json:"labels,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Counter represents a monotonically increasing counter
type Counter struct {
	value int64
	name  string
	help  string
}

// NewCounter creates a new counter
func NewCounter(name, help string) *Counter {
	return &Counter{
		name: name,
		help: help,
	}
}

// Inc increments the counter by 1
func (c *Counter) Inc() {
	atomic.AddInt64(&c.value, 1)
}

// Add adds the given value to the counter
func (c *Counter) Add(value int64) {
	atomic.AddInt64(&c.value, value)
}

// Get returns the current counter value
func (c *Counter) Get() int64 {
	return atomic.LoadInt64(&c.value)
}

// Gauge represents a value that can go up and down
type Gauge struct {
	value int64
	name  string
	help  string
}

// NewGauge creates a new gauge
func NewGauge(name, help string) *Gauge {
	return &Gauge{
		name: name,
		help: help,
	}
}

// Set sets the gauge to the given value
func (g *Gauge) Set(value int64) {
	atomic.StoreInt64(&g.value, value)
}

// Inc increments the gauge by 1
func (g *Gauge) Inc() {
	atomic.AddInt64(&g.value, 1)
}

// Dec decrements the gauge by 1
func (g *Gauge) Dec() {
	atomic.AddInt64(&g.value, -1)
}

// Add adds the given value to the gauge
func (g *Gauge) Add(value int64) {
	atomic.AddInt64(&g.value, value)
}

// Get returns the current gauge value
func (g *Gauge) Get() int64 {
	return atomic.LoadInt64(&g.value)
}

// Histogram tracks distribution of values
type Histogram struct {
	mu      sync.RWMutex
	buckets map[float64]int64
	sum     float64
	count   int64
	name    string
	help    string
}

// NewHistogram creates a new histogram with predefined buckets
func NewHistogram(name, help string) *Histogram {
	// Default buckets for HTTP response times (in milliseconds)
	buckets := map[float64]int64{
		1:     0,
		5:     0,
		10:    0,
		25:    0,
		50:    0,
		100:   0,
		250:   0,
		500:   0,
		1000:  0,
		2500:  0,
		5000:  0,
		10000: 0,
	}

	return &Histogram{
		buckets: buckets,
		name:    name,
		help:    help,
	}
}

// Observe records a new observation
func (h *Histogram) Observe(value float64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.sum += value
	h.count++

	// Update buckets
	for bucket := range h.buckets {
		if value <= bucket {
			h.buckets[bucket]++
		}
	}
}

// GetBuckets returns the current bucket counts
func (h *Histogram) GetBuckets() map[float64]int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make(map[float64]int64, len(h.buckets))
	for k, v := range h.buckets {
		result[k] = v
	}
	return result
}

// GetSum returns the sum of all observations
func (h *Histogram) GetSum() float64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.sum
}

// GetCount returns the number of observations
func (h *Histogram) GetCount() int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.count
}

// Timer helps measure durations
type Timer struct {
	start time.Time
	hist  *Histogram
}

// NewTimer creates a new timer that will observe to the given histogram
func NewTimer(hist *Histogram) *Timer {
	return &Timer{
		start: time.Now(),
		hist:  hist,
	}
}

// ObserveDuration records the duration since timer creation
func (t *Timer) ObserveDuration() {
	duration := float64(time.Since(t.start).Nanoseconds()) / 1e6 // Convert to milliseconds
	t.hist.Observe(duration)
}

// MetricsCollector manages all metrics for the application
type MetricsCollector struct {
	mu sync.RWMutex

	// HTTP metrics
	HTTPRequestsTotal       *Counter
	HTTPRequestDuration     *Histogram
	HTTPRequestsInFlight    *Gauge
	HTTPResponseSize        *Histogram
	HTTPErrorsTotal         *Counter

	// File system metrics
	FilesTotal              *Gauge
	FileProcessingDuration  *Histogram
	FileWatcherEvents       *Counter
	FileOperationsTotal     *Counter

	// Plugin metrics
	PluginExecutionDuration *Histogram
	PluginErrorsTotal       *Counter
	PluginsRegistered       *Gauge

	// Route metrics
	RoutesTotal             *Gauge
	RouteRebuildDuration    *Histogram
	RouteNotFoundTotal      *Counter

	// System metrics
	GoRoutinesCount         *Gauge
	MemoryUsage             *Gauge
	UptimeSeconds           *Gauge

	// Rate limiting metrics
	RateLimitHits           *Counter
	RateLimitBlocks         *Counter

	// Custom metrics registry
	customMetrics map[string]interface{}
	startTime     time.Time
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		// HTTP metrics
		HTTPRequestsTotal:       NewCounter("http_requests_total", "Total number of HTTP requests"),
		HTTPRequestDuration:     NewHistogram("http_request_duration_ms", "HTTP request duration in milliseconds"),
		HTTPRequestsInFlight:    NewGauge("http_requests_in_flight", "Current number of HTTP requests being processed"),
		HTTPResponseSize:        NewHistogram("http_response_size_bytes", "HTTP response size in bytes"),
		HTTPErrorsTotal:         NewCounter("http_errors_total", "Total number of HTTP errors"),

		// File system metrics
		FilesTotal:              NewGauge("files_total", "Total number of files managed"),
		FileProcessingDuration:  NewHistogram("file_processing_duration_ms", "File processing duration in milliseconds"),
		FileWatcherEvents:       NewCounter("file_watcher_events_total", "Total number of file watcher events"),
		FileOperationsTotal:     NewCounter("file_operations_total", "Total number of file operations"),

		// Plugin metrics
		PluginExecutionDuration: NewHistogram("plugin_execution_duration_ms", "Plugin execution duration in milliseconds"),
		PluginErrorsTotal:       NewCounter("plugin_errors_total", "Total number of plugin errors"),
		PluginsRegistered:       NewGauge("plugins_registered", "Number of registered plugins"),

		// Route metrics
		RoutesTotal:             NewGauge("routes_total", "Total number of routes"),
		RouteRebuildDuration:    NewHistogram("route_rebuild_duration_ms", "Route rebuild duration in milliseconds"),
		RouteNotFoundTotal:      NewCounter("route_not_found_total", "Total number of 404 responses"),

		// System metrics
		GoRoutinesCount:         NewGauge("go_routines_count", "Number of Go routines"),
		MemoryUsage:             NewGauge("memory_usage_bytes", "Memory usage in bytes"),
		UptimeSeconds:           NewGauge("uptime_seconds", "Application uptime in seconds"),

		// Rate limiting metrics
		RateLimitHits:           NewCounter("rate_limit_hits_total", "Total number of rate limit hits"),
		RateLimitBlocks:         NewCounter("rate_limit_blocks_total", "Total number of rate limit blocks"),

		customMetrics: make(map[string]interface{}),
		startTime:     time.Now(),
	}
}

// RegisterCustomMetric registers a custom metric
func (mc *MetricsCollector) RegisterCustomMetric(name string, metric interface{}) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.customMetrics[name] = metric
}

// GetCustomMetric retrieves a custom metric
func (mc *MetricsCollector) GetCustomMetric(name string) (interface{}, bool) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	metric, exists := mc.customMetrics[name]
	return metric, exists
}

// UpdateSystemMetrics updates system-level metrics
func (mc *MetricsCollector) UpdateSystemMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	mc.GoRoutinesCount.Set(int64(runtime.NumGoroutine()))
	mc.MemoryUsage.Set(int64(memStats.Alloc))
	mc.UptimeSeconds.Set(int64(time.Since(mc.startTime).Seconds()))
}

// GetAllMetrics returns all current metric values
func (mc *MetricsCollector) GetAllMetrics() map[string]interface{} {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	// Update system metrics before returning
	mc.UpdateSystemMetrics()

	metrics := map[string]interface{}{
		// HTTP metrics
		"http_requests_total":         mc.HTTPRequestsTotal.Get(),
		"http_requests_in_flight":     mc.HTTPRequestsInFlight.Get(),
		"http_errors_total":           mc.HTTPErrorsTotal.Get(),
		"http_request_duration":       mc.getHistogramData(mc.HTTPRequestDuration),
		"http_response_size":          mc.getHistogramData(mc.HTTPResponseSize),

		// File system metrics
		"files_total":                 mc.FilesTotal.Get(),
		"file_watcher_events_total":   mc.FileWatcherEvents.Get(),
		"file_operations_total":       mc.FileOperationsTotal.Get(),
		"file_processing_duration":    mc.getHistogramData(mc.FileProcessingDuration),

		// Plugin metrics
		"plugins_registered":          mc.PluginsRegistered.Get(),
		"plugin_errors_total":         mc.PluginErrorsTotal.Get(),
		"plugin_execution_duration":   mc.getHistogramData(mc.PluginExecutionDuration),

		// Route metrics
		"routes_total":                mc.RoutesTotal.Get(),
		"route_not_found_total":       mc.RouteNotFoundTotal.Get(),
		"route_rebuild_duration":      mc.getHistogramData(mc.RouteRebuildDuration),

		// System metrics
		"go_routines_count":           mc.GoRoutinesCount.Get(),
		"memory_usage_bytes":          mc.MemoryUsage.Get(),
		"uptime_seconds":              mc.UptimeSeconds.Get(),

		// Rate limiting metrics
		"rate_limit_hits_total":       mc.RateLimitHits.Get(),
		"rate_limit_blocks_total":     mc.RateLimitBlocks.Get(),
	}

	// Add custom metrics
	for name, metric := range mc.customMetrics {
		metrics[name] = metric
	}

	return metrics
}

func (mc *MetricsCollector) getHistogramData(hist *Histogram) map[string]interface{} {
	return map[string]interface{}{
		"buckets": hist.GetBuckets(),
		"sum":     hist.GetSum(),
		"count":   hist.GetCount(),
	}
}

// MetricsMiddleware creates a Gin middleware for collecting HTTP metrics
func (mc *MetricsCollector) MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip metrics for the metrics endpoint itself
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		start := time.Now()
		mc.HTTPRequestsInFlight.Inc()
		timer := NewTimer(mc.HTTPRequestDuration)

		// Process request
		c.Next()

		// Record metrics
		duration := time.Since(start)
		timer.ObserveDuration()
		mc.HTTPRequestsInFlight.Dec()
		mc.HTTPRequestsTotal.Inc()

		// Record response size
		if c.Writer.Size() > 0 {
			mc.HTTPResponseSize.Observe(float64(c.Writer.Size()))
		}

		// Record errors
		if c.Writer.Status() >= 400 {
			mc.HTTPErrorsTotal.Inc()
			if c.Writer.Status() == 404 {
				mc.RouteNotFoundTotal.Inc()
			}
		}

		// Log slow requests
		if duration > 1*time.Second {
			Info("Slow request detected: %s %s took %v", c.Request.Method, c.Request.URL.Path, duration)
		}
	}
}

// MetricsHandler returns an HTTP handler for the /metrics endpoint
func (mc *MetricsCollector) MetricsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		metrics := mc.GetAllMetrics()
		c.JSON(http.StatusOK, gin.H{
			"timestamp": time.Now(),
			"metrics":   metrics,
		})
	}
}

// PrometheusHandler returns metrics in Prometheus format
func (mc *MetricsCollector) PrometheusHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

		metrics := mc.GetAllMetrics()
		output := mc.formatPrometheusMetrics(metrics)

		c.String(http.StatusOK, output)
	}
}

func (mc *MetricsCollector) formatPrometheusMetrics(metrics map[string]interface{}) string {
	var output string

	for name, value := range metrics {
		switch v := value.(type) {
		case int64:
			output += fmt.Sprintf("# TYPE %s gauge\n", name)
			output += fmt.Sprintf("%s %d\n", name, v)
		case float64:
			output += fmt.Sprintf("# TYPE %s gauge\n", name)
			output += fmt.Sprintf("%s %.2f\n", name, v)
		case map[string]interface{}:
			// Handle histogram data
			if buckets, ok := v["buckets"].(map[float64]int64); ok {
				output += fmt.Sprintf("# TYPE %s histogram\n", name)
				for bucket, count := range buckets {
					output += fmt.Sprintf("%s_bucket{le=\"%.1f\"} %d\n", name, bucket, count)
				}
				if sum, ok := v["sum"].(float64); ok {
					output += fmt.Sprintf("%s_sum %.2f\n", name, sum)
				}
				if count, ok := v["count"].(int64); ok {
					output += fmt.Sprintf("%s_count %d\n", name, count)
				}
			}
		}
		output += "\n"
	}

	return output
}

// StartMetricsCollector starts background metric collection
func (mc *MetricsCollector) StartMetricsCollector(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			mc.UpdateSystemMetrics()
		}
	}
}

// Global metrics collector instance
var GlobalMetrics = NewMetricsCollector()

// Convenience functions for global metrics
func RecordHTTPRequest() {
	GlobalMetrics.HTTPRequestsTotal.Inc()
}

func RecordFileOperation() {
	GlobalMetrics.FileOperationsTotal.Inc()
}

func RecordFileWatcherEvent() {
	GlobalMetrics.FileWatcherEvents.Inc()
}

func RecordPluginError() {
	GlobalMetrics.PluginErrorsTotal.Inc()
}

func RecordRateLimitHit() {
	GlobalMetrics.RateLimitHits.Inc()
}

func RecordRateLimitBlock() {
	GlobalMetrics.RateLimitBlocks.Inc()
}

func SetFilesCount(count int64) {
	GlobalMetrics.FilesTotal.Set(count)
}

func SetRoutesCount(count int64) {
	GlobalMetrics.RoutesTotal.Set(count)
}

func SetPluginsCount(count int64) {
	GlobalMetrics.PluginsRegistered.Set(count)
}

func NewFileProcessingTimer() *Timer {
	return NewTimer(GlobalMetrics.FileProcessingDuration)
}

func NewPluginExecutionTimer() *Timer {
	return NewTimer(GlobalMetrics.PluginExecutionDuration)
}

func NewRouteRebuildTimer() *Timer {
	return NewTimer(GlobalMetrics.RouteRebuildDuration)
}