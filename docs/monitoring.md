# MiniCMS Monitoring & Metrics Specification

## Overview

MiniCMS includes a comprehensive monitoring and metrics system designed to provide real-time insights into application performance, system health, and operational metrics. The system follows industry best practices and integrates seamlessly with popular monitoring tools like Prometheus, Grafana, and Kubernetes.

## Architecture

### Core Components

- **MetricsCollector**: Central hub for all application metrics
- **HealthChecker**: Component health monitoring and readiness checks
- **Middleware Integration**: Automatic instrumentation of HTTP requests
- **Background Services**: Continuous system metrics collection

### Metric Types

The system supports four primary metric types:

1. **Counter**: Monotonically increasing values (e.g., total requests)
2. **Gauge**: Values that can increase or decrease (e.g., memory usage)
3. **Histogram**: Distribution tracking with configurable buckets (e.g., response times)
4. **Timer**: Convenient wrapper for duration measurement

## Available Metrics

### HTTP Metrics

| Metric Name | Type | Description | Labels |
|-------------|------|-------------|---------|
| `http_requests_total` | Counter | Total number of HTTP requests processed | - |
| `http_request_duration_ms` | Histogram | HTTP request duration in milliseconds | - |
| `http_requests_in_flight` | Gauge | Current number of requests being processed | - |
| `http_response_size_bytes` | Histogram | HTTP response size distribution | - |
| `http_errors_total` | Counter | Total number of HTTP errors (4xx, 5xx) | - |
| `route_not_found_total` | Counter | Total number of 404 responses | - |

**Histogram Buckets** (milliseconds): 1, 5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000, 10000

### Application Performance Metrics

| Metric Name | Type | Description | Labels |
|-------------|------|-------------|---------|
| `files_total` | Gauge | Total number of files managed by FileManager | - |
| `file_operations_total` | Counter | Total number of file operations performed | - |
| `file_processing_duration_ms` | Histogram | Time to process files through plugins | - |
| `file_watcher_events_total` | Counter | Total file system events processed | - |

### Plugin System Metrics

| Metric Name | Type | Description | Labels |
|-------------|------|-------------|---------|
| `plugins_registered` | Gauge | Number of registered plugins | - |
| `plugin_execution_duration_ms` | Histogram | Plugin execution time distribution | - |
| `plugin_errors_total` | Counter | Total number of plugin execution errors | - |

### Routing Metrics

| Metric Name | Type | Description | Labels |
|-------------|------|-------------|---------|
| `routes_total` | Gauge | Total number of registered routes | - |
| `route_rebuild_duration_ms` | Histogram | Time taken to rebuild routes | - |

### System Metrics

| Metric Name | Type | Description | Labels |
|-------------|------|-------------|---------|
| `go_routines_count` | Gauge | Number of active Go routines | - |
| `memory_usage_bytes` | Gauge | Current memory allocation in bytes | - |
| `uptime_seconds` | Gauge | Application uptime in seconds | - |

### Security Metrics

| Metric Name | Type | Description | Labels |
|-------------|------|-------------|---------|
| `rate_limit_hits_total` | Counter | Total requests subject to rate limiting | - |
| `rate_limit_blocks_total` | Counter | Total requests blocked by rate limiter | - |

## HTTP Endpoints

### Metrics Endpoints

#### `/metrics`
Returns metrics in JSON format suitable for custom dashboards and applications.

**Response Format:**
```json
{
  "timestamp": "2025-01-15T10:30:00Z",
  "metrics": {
    "http_requests_total": 1234,
    "http_request_duration": {
      "buckets": {
        "1": 50,
        "5": 120,
        "10": 200,
        "25": 300,
        "50": 450,
        "100": 600,
        "250": 700,
        "500": 750,
        "1000": 800,
        "2500": 820,
        "5000": 825,
        "10000": 830
      },
      "sum": 15420.5,
      "count": 830
    },
    "memory_usage_bytes": 67108864,
    "uptime_seconds": 3600
  }
}
```

#### `/metrics/prometheus`
Returns metrics in Prometheus exposition format for direct integration with Prometheus monitoring.

**Response Format:**
```
# TYPE http_requests_total counter
http_requests_total 1234

# TYPE http_request_duration_ms histogram
http_request_duration_ms_bucket{le="1.0"} 50
http_request_duration_ms_bucket{le="5.0"} 120
http_request_duration_ms_bucket{le="10.0"} 200
http_request_duration_ms_bucket{le="+Inf"} 830
http_request_duration_ms_sum 15420.50
http_request_duration_ms_count 830

# TYPE memory_usage_bytes gauge
memory_usage_bytes 67108864
```

### Health Check Endpoints

#### `/health`
Comprehensive health check that executes all registered health checks and returns detailed status.

**Response Format:**
```json
{
  "status": "healthy",
  "timestamp": "2025-01-15T10:30:00Z",
  "last_update": "2025-01-15T10:29:45Z",
  "checks": {
    "file_manager": {
      "name": "file_manager",
      "status": "healthy",
      "message": "",
      "last_checked": "2025-01-15T10:29:45Z",
      "duration": "2ms"
    },
    "file_watcher": {
      "name": "file_watcher",
      "status": "healthy",
      "message": "",
      "last_checked": "2025-01-15T10:29:45Z",
      "duration": "1ms"
    },
    "memory": {
      "name": "memory",
      "status": "healthy",
      "message": "",
      "last_checked": "2025-01-15T10:29:45Z",
      "duration": "0ms"
    },
    "goroutines": {
      "name": "goroutines",
      "status": "healthy",
      "message": "",
      "last_checked": "2025-01-15T10:29:45Z",
      "duration": "0ms"
    }
  }
}
```

**Status Codes:**
- `200 OK`: Application is healthy or degraded but operational
- `503 Service Unavailable`: Application is unhealthy or not ready

**Health Status Values:**
- `healthy`: All checks passing
- `degraded`: Some non-critical checks failing
- `unhealthy`: Critical checks failing
- `unknown`: Unable to determine status

#### `/health/live`
Simple liveness probe for Kubernetes and container orchestration.

**Response Format:**
```json
{
  "status": "alive",
  "timestamp": "2025-01-15T10:30:00Z"
}
```

**Status Codes:**
- `200 OK`: Application process is running

#### `/health/ready`
Readiness probe indicating if the application is ready to receive traffic.

**Response Format:**
```json
{
  "status": "ready",
  "timestamp": "2025-01-15T10:30:00Z"
}
```

**Status Codes:**
- `200 OK`: Application is ready to serve requests
- `503 Service Unavailable`: Application is not ready

## Health Checks

### Built-in Health Checks

#### File Manager Health Check
- **Name**: `file_manager`
- **Purpose**: Verifies FileManager is initialized and accessible
- **Failure Conditions**:
  - FileManager is nil
  - Root directory is inaccessible
  - Cannot retrieve file list

#### File Watcher Health Check
- **Name**: `file_watcher`
- **Purpose**: Ensures file watching system is operational
- **Failure Conditions**:
  - FileWatcher is nil
  - FileWatcher is not running
  - No directories being watched

#### Plugin Manager Health Check
- **Name**: `plugin_manager`
- **Purpose**: Validates plugin system functionality
- **Failure Conditions**:
  - PluginManager is nil
  - No plugins registered

#### Memory Health Check
- **Name**: `memory`
- **Purpose**: Monitors memory usage for potential leaks
- **Failure Conditions**:
  - Memory usage exceeds 1GB threshold

#### Goroutine Health Check
- **Name**: `goroutines`
- **Purpose**: Detects potential goroutine leaks
- **Failure Conditions**:
  - Goroutine count exceeds 1000

### Custom Health Checks

Health checks can be registered programmatically:

```go
import "cms/core"

// Register a custom health check
core.GlobalHealthChecker.RegisterCheck("database", func(ctx context.Context) error {
    // Custom health check logic
    if !isDatabaseConnected() {
        return fmt.Errorf("database connection failed")
    }
    return nil
})
```

## Integration Examples

### Prometheus Configuration

Add the following job to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'minicms'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics/prometheus'
    scrape_interval: 30s
```

### Grafana Dashboard

Key panels to include:

1. **Request Rate**: `rate(http_requests_total[5m])`
2. **Response Time**: `histogram_quantile(0.95, rate(http_request_duration_ms_bucket[5m]))`
3. **Error Rate**: `rate(http_errors_total[5m]) / rate(http_requests_total[5m])`
4. **Memory Usage**: `memory_usage_bytes`
5. **Active Connections**: `http_requests_in_flight`

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: minicms
spec:
  template:
    spec:
      containers:
      - name: minicms
        image: minicms:latest
        ports:
        - containerPort: 8080
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

### Alerting Rules

Example Prometheus alerting rules:

```yaml
groups:
- name: minicms
  rules:
  - alert: HighErrorRate
    expr: rate(http_errors_total[5m]) / rate(http_requests_total[5m]) > 0.1
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "MiniCMS high error rate"
      description: "Error rate is {{ $value | humanizePercentage }}"

  - alert: HighMemoryUsage
    expr: memory_usage_bytes > 1073741824  # 1GB
    for: 2m
    labels:
      severity: warning
    annotations:
      summary: "MiniCMS high memory usage"
      description: "Memory usage is {{ $value | humanizeBytes }}"

  - alert: ServiceDown
    expr: up{job="minicms"} == 0
    for: 1m
    labels:
      severity: critical
    annotations:
      summary: "MiniCMS service is down"
      description: "MiniCMS has been down for more than 1 minute"
```

## Configuration

### Metrics Collection

Metrics collection is automatically enabled and runs in the background. The collection interval for system metrics is 30 seconds.

### Health Check Frequency

Health checks are executed:
- **On-demand**: When `/health` endpoint is accessed
- **Periodically**: Every 60 seconds in the background
- **Individual checks**: Can be executed via API

### Rate Limiting Configuration

The built-in rate limiter can be configured in the router setup:

```go
// Create rate limiter with custom limit (requests per minute)
rateLimiter := core.NewRateLimiter(120) // 120 requests per minute
router.Use(rateLimiter.Middleware())
```

## Performance Considerations

### Overhead

The monitoring system is designed for minimal performance impact:

- **Counters/Gauges**: Use atomic operations (nanosecond overhead)
- **Histograms**: Simple bucket increment operations
- **Health Checks**: Execute in separate goroutines
- **Background Collection**: Runs every 30 seconds, not per-request

### Memory Usage

- **Metric Storage**: Minimal memory footprint using atomic values
- **Rate Limiter**: Automatic cleanup of old client entries
- **Health Check Results**: Cached and updated periodically

### Scalability

The system scales with your application:
- All metrics operations are thread-safe
- No blocking operations in request path
- Configurable collection intervals
- Optional components can be disabled if needed

## Troubleshooting

### Common Issues

#### Metrics Not Updating
- Verify background metrics collector is running
- Check for panics in application logs
- Ensure endpoints are accessible

#### Health Checks Failing
- Review individual check error messages in `/health` response
- Verify all components are properly initialized
- Check system resource availability

#### High Memory Usage Alerts
- Review `memory_usage_bytes` metric trends
- Check for goroutine leaks using `go_routines_count`
- Monitor file cache size via `files_total`

### Debug Information

Enable debug logging to see detailed metrics information:

```go
core.GlobalLogger.SetLevel(core.LogLevelDebug)
```

This will provide detailed logging of:
- Metric updates
- Health check executions
- Background service operations
- Rate limiting decisions

## Security Considerations

### Endpoint Security

Monitor endpoints expose operational data that could be sensitive:

- Consider restricting access to metrics endpoints
- Use authentication/authorization if required
- Monitor access patterns to these endpoints

### Rate Limiting

The rate limiting system provides basic DDoS protection:

- Per-IP tracking with configurable limits
- Automatic cleanup prevents memory leaks
- Metrics track both hits and blocks

### Data Privacy

Metrics do not include:
- User data or personal information
- Request content or parameters
- Authentication tokens or credentials
- File content (only metadata and counts)

## Future Enhancements

### Planned Features

- **Custom Metric Labels**: Support for dimensional metrics
- **Distributed Tracing**: Integration with OpenTelemetry
- **Advanced Alerting**: Built-in alert manager
- **Metric Persistence**: Optional local storage for metrics
- **Dashboard UI**: Built-in metrics visualization

### Extension Points

The monitoring system is designed for extensibility:

- Custom metric types can be added
- Health checks are fully pluggable
- Middleware can be extended for additional instrumentation
- Export formats can be customized