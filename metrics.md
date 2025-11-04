# S3 Plugin Metrics - Complete Grafana Guide

This guide provides ready-to-use PromQL queries for S3 Storage Plugin metrics, including panel configuration details (legend, min step, units).

---

## Overview

The S3 plugin exposes two primary metrics to track file operations and errors:

- `rr_s3_operations_total` - Counter tracking all S3 operations by type, bucket, and status
- `rr_s3_errors_total` - Counter tracking errors by bucket and error type

---

## 1. Operation Metrics

### 1.1 Total Operations Per Second

**Query:**

```promql
sum(rate(rr_s3_operations_total[5m]))
```

**Configuration:**

- **Legend:** `Total OPS`
- **Min Step:** `15s`
- **Unit:** `ops (operations/sec)`
- **Panel Type:** Graph
- **Description:** Overall S3 operation rate across all buckets

---

### 1.2 Operations Per Second by Bucket

**Query:**

```promql
sum by (bucket) (rate(rr_s3_operations_total[5m]))
```

**Configuration:**

- **Legend:** `{{bucket}}`
- **Min Step:** `15s`
- **Unit:** `ops (operations/sec)`
- **Panel Type:** Graph or Bar gauge
- **Description:** Operation rate grouped by bucket

---

### 1.3 Operations Per Second by Type

**Query:**

```promql
sum by (operation) (rate(rr_s3_operations_total[5m]))
```

**Configuration:**

- **Legend:** `{{operation}}`
- **Min Step:** `15s`
- **Unit:** `ops (operations/sec)`
- **Panel Type:** Graph (stacked area) or Pie chart
- **Description:** Operation distribution by type (write, read, delete, copy, move, list, exists, get_metadata, set_visibility, get_url)

---

### 1.4 Operations Per Second by Status

**Query:**

```promql
sum by (status) (rate(rr_s3_operations_total[5m]))
```

**Configuration:**

- **Legend:** `{{status}}`
- **Min Step:** `15s`
- **Unit:** `ops (operations/sec)`
- **Panel Type:** Graph (stacked area)
- **Description:** Operation rate grouped by status (success, error)

---

### 1.5 Success Rate Percentage

**Query:**

```promql
sum(rate(rr_s3_operations_total{status="success"}[5m])) / sum(rate(rr_s3_operations_total[5m])) * 100
```

**Configuration:**

- **Legend:** `Success Rate`
- **Min Step:** `15s`
- **Unit:** `percent (0-100)`
- **Panel Type:** Gauge or Graph
- **Thresholds:** Red < 95%, Yellow 95-99%, Green > 99%
- **Description:** Percentage of successful S3 operations

---

### 1.6 Total Operations Count

**Query:**

```promql
sum(rr_s3_operations_total)
```

**Configuration:**

- **Legend:** `Total Operations`
- **Min Step:** `1m`
- **Unit:** `short`
- **Panel Type:** Stat
- **Description:** Cumulative count of all S3 operations since start

---

## 2. Bucket Analysis

### 2.1 Most Active Buckets (by Total Operations)

**Query:**

```promql
topk(10, sum by (bucket) (rate(rr_s3_operations_total[5m])))
```

**Configuration:**

- **Legend:** `{{bucket}}`
- **Min Step:** `15s`
- **Unit:** `ops (operations/sec)`
- **Panel Type:** Bar gauge (horizontal) or Table
- **Description:** Top 10 buckets by operation rate

---

### 2.2 Most Active Buckets (by Write Operations)

**Query:**

```promql
topk(10, sum by (bucket) (rate(rr_s3_operations_total{operation="write"}[5m])))
```

**Configuration:**

- **Legend:** `{{bucket}}`
- **Min Step:** `15s`
- **Unit:** `ops (operations/sec)`
- **Panel Type:** Bar gauge or Table
- **Description:** Top 10 buckets by write operation rate

---

### 2.3 Most Active Buckets (by Read Operations)

**Query:**

```promql
topk(10, sum by (bucket) (rate(rr_s3_operations_total{operation="read"}[5m])))
```

**Configuration:**

- **Legend:** `{{bucket}}`
- **Min Step:** `15s`
- **Unit:** `ops (operations/sec)`
- **Panel Type:** Bar gauge or Table
- **Description:** Top 10 buckets by read operation rate

---

### 2.4 Bucket Performance Table

**Query 1 (Total OPS):**

```promql
sum by (bucket) (rate(rr_s3_operations_total[5m]))
```

**Query 2 (Write OPS):**

```promql
sum by (bucket) (rate(rr_s3_operations_total{operation="write"}[5m]))
```

**Query 3 (Read OPS):**

```promql
sum by (bucket) (rate(rr_s3_operations_total{operation="read"}[5m]))
```

**Query 4 (Error %):**

```promql
sum by (bucket) (rate(rr_s3_operations_total{status="error"}[5m])) / sum by (bucket) (rate(rr_s3_operations_total[5m])) * 100
```

**Configuration:**

- **Legend:** N/A (Table columns)
- **Min Step:** `15s`
- **Unit:**
    - Query 1-3: `ops (operations/sec)`
    - Query 4: `percent (0-100)`
- **Panel Type:** Table
- **Column Names:** `Bucket`, `Total OPS`, `Write OPS`, `Read OPS`, `Error Rate %`
- **Description:** Comprehensive bucket performance overview

---

### 2.5 Read/Write Ratio by Bucket

**Query:**

```promql
sum by (bucket) (rate(rr_s3_operations_total{operation="read"}[5m])) / sum by (bucket) (rate(rr_s3_operations_total{operation="write"}[5m]))
```

**Configuration:**

- **Legend:** `{{bucket}}`
- **Min Step:** `15s`
- **Unit:** `short` (ratio)
- **Panel Type:** Graph or Table
- **Description:** Read-to-write ratio per bucket (higher = more reads than writes)

---

## 3. Operation Type Analysis

### 3.1 Write Operations Rate

**Query:**

```promql
sum(rate(rr_s3_operations_total{operation="write"}[5m]))
```

**Configuration:**

- **Legend:** `Writes/sec`
- **Min Step:** `15s`
- **Unit:** `ops (operations/sec)`
- **Panel Type:** Graph
- **Description:** Total file upload rate

---

### 3.2 Read Operations Rate

**Query:**

```promql
sum(rate(rr_s3_operations_total{operation="read"}[5m]))
```

**Configuration:**

- **Legend:** `Reads/sec`
- **Min Step:** `15s`
- **Unit:** `ops (operations/sec)`
- **Panel Type:** Graph
- **Description:** Total file download rate

---

### 3.3 Delete Operations Rate

**Query:**

```promql
sum(rate(rr_s3_operations_total{operation="delete"}[5m]))
```

**Configuration:**

- **Legend:** `Deletes/sec`
- **Min Step:** `15s`
- **Unit:** `ops (operations/sec)`
- **Panel Type:** Graph
- **Description:** Total file deletion rate

---

### 3.4 List Operations Rate

**Query:**

```promql
sum(rate(rr_s3_operations_total{operation="list"}[5m]))
```

**Configuration:**

- **Legend:** `Lists/sec`
- **Min Step:** `15s`
- **Unit:** `ops (operations/sec)`
- **Panel Type:** Graph
- **Description:** Total object listing rate

---

### 3.5 Copy Operations Rate

**Query:**

```promql
sum(rate(rr_s3_operations_total{operation="copy"}[5m]))
```

**Configuration:**

- **Legend:** `Copies/sec`
- **Min Step:** `15s`
- **Unit:** `ops (operations/sec)`
- **Panel Type:** Graph
- **Description:** Total file copy rate

---

### 3.6 Move Operations Rate

**Query:**

```promql
sum(rate(rr_s3_operations_total{operation="move"}[5m]))
```

**Configuration:**

- **Legend:** `Moves/sec`
- **Min Step:** `15s`
- **Unit:** `ops (operations/sec)`
- **Panel Type:** Graph
- **Description:** Total file move rate

---

### 3.7 Exists Check Rate

**Query:**

```promql
sum(rate(rr_s3_operations_total{operation="exists"}[5m]))
```

**Configuration:**

- **Legend:** `Exists checks/sec`
- **Min Step:** `15s`
- **Unit:** `ops (operations/sec)`
- **Panel Type:** Graph
- **Description:** Total file existence check rate

---

### 3.8 Metadata Operations Rate

**Query:**

```promql
sum(rate(rr_s3_operations_total{operation="get_metadata"}[5m]))
```

**Configuration:**

- **Legend:** `Metadata ops/sec`
- **Min Step:** `15s`
- **Unit:** `ops (operations/sec)`
- **Panel Type:** Graph
- **Description:** Total metadata retrieval rate

---

### 3.9 Visibility Change Rate

**Query:**

```promql
sum(rate(rr_s3_operations_total{operation="set_visibility"}[5m]))
```

**Configuration:**

- **Legend:** `Visibility changes/sec`
- **Min Step:** `15s`
- **Unit:** `ops (operations/sec)`
- **Panel Type:** Graph
- **Description:** Total ACL change rate

---

### 3.10 URL Generation Rate

**Query:**

```promql
sum(rate(rr_s3_operations_total{operation="get_url"}[5m]))
```

**Configuration:**

- **Legend:** `URL gens/sec`
- **Min Step:** `15s`
- **Unit:** `ops (operations/sec)`
- **Panel Type:** Graph
- **Description:** Total URL generation rate (public/presigned)

---

### 3.11 Operation Distribution (Pie Chart)

**Query:**

```promql
sum by (operation) (rate(rr_s3_operations_total[5m])) / sum(rate(rr_s3_operations_total[5m])) * 100
```

**Configuration:**

- **Legend:** `{{operation}}`
- **Min Step:** `15s`
- **Unit:** `percent (0-100)`
- **Panel Type:** Pie chart
- **Description:** Percentage breakdown of operations by type

---

### 3.12 Write vs Read Operations (Stacked)

**Query 1 (Writes):**

```promql
sum(rate(rr_s3_operations_total{operation="write"}[5m]))
```

**Query 2 (Reads):**

```promql
sum(rate(rr_s3_operations_total{operation="read"}[5m]))
```

**Configuration:**

- **Legend:**
    - Query 1: `Writes`
    - Query 2: `Reads`
- **Min Step:** `15s`
- **Unit:** `ops (operations/sec)`
- **Panel Type:** Graph (stacked area)
- **Description:** Visual comparison of write vs read operations

---

## 4. Error Tracking

### 4.1 Total Error Rate

**Query:**

```promql
sum(rate(rr_s3_errors_total[5m]))
```

**Configuration:**

- **Legend:** `Errors/sec`
- **Min Step:** `15s`
- **Unit:** `errors/sec`
- **Panel Type:** Graph
- **Description:** Total S3 errors per second (all types)

---

### 4.2 Error Rate Percentage

**Query:**

```promql
sum(rate(rr_s3_operations_total{status="error"}[5m])) / sum(rate(rr_s3_operations_total[5m])) * 100
```

**Configuration:**

- **Legend:** `Error Rate`
- **Min Step:** `15s`
- **Unit:** `percent (0-100)`
- **Panel Type:** Gauge or Graph
- **Thresholds:** Green < 1%, Yellow 1-5%, Red > 5%
- **Description:** Percentage of operations that result in errors

---

### 4.3 Error Rate by Type

**Query:**

```promql
sum by (error_type) (rate(rr_s3_errors_total[5m]))
```

**Configuration:**

- **Legend:** `{{error_type}}`
- **Min Step:** `15s`
- **Unit:** `errors/sec`
- **Panel Type:** Graph (stacked) or Pie chart
- **Description:** Errors grouped by classification (BUCKET_NOT_FOUND, FILE_NOT_FOUND, S3_OPERATION_FAILED, etc.)

---

### 4.4 Error Rate by Bucket

**Query:**

```promql
sum by (bucket) (rate(rr_s3_errors_total[5m]))
```

**Configuration:**

- **Legend:** `{{bucket}}`
- **Min Step:** `15s`
- **Unit:** `errors/sec`
- **Panel Type:** Graph or Table
- **Description:** Errors grouped by bucket

---

### 4.5 Most Error-Prone Buckets (by Count)

**Query:**

```promql
topk(10, sum by (bucket) (rate(rr_s3_errors_total[5m])))
```

**Configuration:**

- **Legend:** `{{bucket}}`
- **Min Step:** `15s`
- **Unit:** `errors/sec`
- **Panel Type:** Bar gauge or Table
- **Description:** Buckets with highest error rate

---

### 4.6 Most Error-Prone Buckets (by Percentage)

**Query:**

```promql
topk(10, sum by (bucket) (rate(rr_s3_operations_total{status="error"}[5m])) / sum by (bucket) (rate(rr_s3_operations_total[5m])) * 100)
```

**Configuration:**

- **Legend:** `{{bucket}}`
- **Min Step:** `15s`
- **Unit:** `percent (0-100)`
- **Panel Type:** Bar gauge or Table
- **Description:** Buckets with highest error percentage

---

### 4.7 Bucket Not Found Errors

**Query:**

```promql
sum(rate(rr_s3_errors_total{error_type="BUCKET_NOT_FOUND"}[5m]))
```

**Configuration:**

- **Legend:** `Bucket Not Found`
- **Min Step:** `15s`
- **Unit:** `errors/sec`
- **Panel Type:** Graph
- **Description:** Rate of bucket not found errors

---

### 4.8 File Not Found Errors

**Query:**

```promql
sum(rate(rr_s3_errors_total{error_type="FILE_NOT_FOUND"}[5m]))
```

**Configuration:**

- **Legend:** `File Not Found`
- **Min Step:** `15s`
- **Unit:** `errors/sec`
- **Panel Type:** Graph
- **Description:** Rate of file not found errors

---

### 4.9 S3 Operation Failed Errors

**Query:**

```promql
sum(rate(rr_s3_errors_total{error_type="S3_OPERATION_FAILED"}[5m]))
```

**Configuration:**

- **Legend:** `S3 Operation Failed`
- **Min Step:** `15s`
- **Unit:** `errors/sec`
- **Panel Type:** Graph
- **Description:** Rate of S3 SDK operation failures

---

### 4.10 Permission Denied Errors

**Query:**

```promql
sum(rate(rr_s3_errors_total{error_type="PERMISSION_DENIED"}[5m]))
```

**Configuration:**

- **Legend:** `Permission Denied`
- **Min Step:** `15s`
- **Unit:** `errors/sec`
- **Panel Type:** Graph
- **Description:** Rate of permission/access denied errors

---

### 4.11 Invalid Pathname Errors

**Query:**

```promql
sum(rate(rr_s3_errors_total{error_type="INVALID_PATHNAME"}[5m]))
```

**Configuration:**

- **Legend:** `Invalid Pathname`
- **Min Step:** `15s`
- **Unit:** `errors/sec`
- **Panel Type:** Graph
- **Description:** Rate of invalid pathname errors

---

### 4.12 Operation Timeout Errors

**Query:**

```promql
sum(rate(rr_s3_errors_total{error_type="OPERATION_TIMEOUT"}[5m]))
```

**Configuration:**

- **Legend:** `Timeouts`
- **Min Step:** `15s`
- **Unit:** `errors/sec`
- **Panel Type:** Graph
- **Thresholds:** Any value > 0 requires investigation
- **Description:** Rate of operation timeout errors

---

### 4.13 Error Distribution (Pie Chart)

**Query:**

```promql
sum by (error_type) (rate(rr_s3_errors_total[5m])) / sum(rate(rr_s3_errors_total[5m])) * 100
```

**Configuration:**

- **Legend:** `{{error_type}}`
- **Min Step:** `15s`
- **Unit:** `percent (0-100)`
- **Panel Type:** Pie chart
- **Description:** Percentage breakdown of errors by type

---

### 4.14 Error Heatmap (Bucket vs Error Type)

**Query:**

```promql
sum by (bucket, error_type) (rate(rr_s3_errors_total[5m]))
```

**Configuration:**

- **Legend:** N/A (heatmap)
- **Min Step:** `30s`
- **Unit:** `errors/sec`
- **Panel Type:** Heatmap
- **Description:** Visual correlation between buckets and error types

---

## 5. Combined Operation & Error Analysis

### 5.1 Operations and Errors (Dual Axis)

**Query 1 (Operations - Left Axis):**

```promql
sum(rate(rr_s3_operations_total[5m]))
```

**Query 2 (Errors - Right Axis):**

```promql
sum(rate(rr_s3_errors_total[5m]))
```

**Configuration:**

- **Legend:**
    - Query 1: `Operations/sec`
    - Query 2: `Errors/sec`
- **Min Step:** `15s`
- **Unit:**
    - Left Axis: `ops (operations/sec)`
    - Right Axis: `errors/sec`
- **Panel Type:** Graph (dual Y-axis)
- **Description:** Correlation between operation rate and error rate

---

### 5.2 Success vs Error Rate (Stacked)

**Query 1 (Success):**

```promql
sum(rate(rr_s3_operations_total{status="success"}[5m]))
```

**Query 2 (Error):**

```promql
sum(rate(rr_s3_operations_total{status="error"}[5m]))
```

**Configuration:**

- **Legend:**
    - Query 1: `Success`
    - Query 2: `Error`
- **Min Step:** `15s`
- **Unit:** `ops (operations/sec)`
- **Panel Type:** Graph (stacked area)
- **Description:** Visual comparison of successful vs failed operations

---

### 5.3 Write Operation Success Rate

**Query:**

```promql
sum(rate(rr_s3_operations_total{operation="write",status="success"}[5m])) / sum(rate(rr_s3_operations_total{operation="write"}[5m])) * 100
```

**Configuration:**

- **Legend:** `Write Success Rate`
- **Min Step:** `15s`
- **Unit:** `percent (0-100)`
- **Panel Type:** Graph or Gauge
- **Thresholds:** Red < 95%, Yellow 95-99%, Green > 99%
- **Description:** Success rate for write operations only

---

### 5.4 Read Operation Success Rate

**Query:**

```promql
sum(rate(rr_s3_operations_total{operation="read",status="success"}[5m])) / sum(rate(rr_s3_operations_total{operation="read"}[5m])) * 100
```

**Configuration:**

- **Legend:** `Read Success Rate`
- **Min Step:** `15s`
- **Unit:** `percent (0-100)`
- **Panel Type:** Graph or Gauge
- **Thresholds:** Red < 95%, Yellow 95-99%, Green > 99%
- **Description:** Success rate for read operations only

---

### 5.5 Operation Success Rate by Type (Table)

**Query 1 (Write):**

```promql
sum(rate(rr_s3_operations_total{operation="write",status="success"}[5m])) / sum(rate(rr_s3_operations_total{operation="write"}[5m])) * 100
```

**Query 2 (Read):**

```promql
sum(rate(rr_s3_operations_total{operation="read",status="success"}[5m])) / sum(rate(rr_s3_operations_total{operation="read"}[5m])) * 100
```

**Query 3 (Delete):**

```promql
sum(rate(rr_s3_operations_total{operation="delete",status="success"}[5m])) / sum(rate(rr_s3_operations_total{operation="delete"}[5m])) * 100
```

**Query 4 (List):**

```promql
sum(rate(rr_s3_operations_total{operation="list",status="success"}[5m])) / sum(rate(rr_s3_operations_total{operation="list"}[5m])) * 100
```

**Configuration:**

- **Legend:** N/A (Table rows)
- **Min Step:** `15s`
- **Unit:** `percent (0-100)`
- **Panel Type:** Table
- **Row Names:** `Write`, `Read`, `Delete`, `List`
- **Description:** Success rate breakdown by operation type

---

## 6. Advanced Analytics

### 6.1 Operation Rate Trend (Hour over Hour)

**Query:**

```promql
sum(rate(rr_s3_operations_total[1h])) / sum(rate(rr_s3_operations_total[1h] offset 24h))
```

**Configuration:**

- **Legend:** `HoH Change`
- **Min Step:** `5m`
- **Unit:** `short` (ratio)
- **Panel Type:** Graph or Stat
- **Description:** Current hour traffic vs same hour yesterday (1.0 = same, 2.0 = double)

---

### 6.2 Error Burst Detection

**Query:**

```promql
sum(rate(rr_s3_errors_total[1m])) > 2 * avg_over_time(sum(rate(rr_s3_errors_total[1m]))[10m:1m])
```

**Configuration:**

- **Legend:** `Error Burst`
- **Min Step:** `15s`
- **Unit:** `bool` (0 or 1)
- **Panel Type:** Graph (binary)
- **Thresholds:** Red when value = 1
- **Description:** Detects sudden spikes in errors (>2x baseline)

---

### 6.3 Operations Per Bucket (Distribution)

**Query:**

```promql
sum by (bucket) (rr_s3_operations_total)
```

**Configuration:**

- **Legend:** `{{bucket}}`
- **Min Step:** `1m`
- **Unit:** `short`
- **Panel Type:** Pie chart or Bar gauge
- **Description:** Total cumulative operations per bucket

---

### 6.4 Bucket Activity Timeline (Heatmap)

**Query:**

```promql
sum by (bucket) (rate(rr_s3_operations_total[5m]))
```

**Configuration:**

- **Legend:** N/A (heatmap)
- **Min Step:** `30s`
- **Unit:** `ops (operations/sec)`
- **Panel Type:** Heatmap
- **Description:** Visual activity pattern across buckets over time

---

### 6.5 Write-Heavy vs Read-Heavy Buckets

**Query:**

```promql
(sum by (bucket) (rate(rr_s3_operations_total{operation="write"}[5m])) > sum by (bucket) (rate(rr_s3_operations_total{operation="read"}[5m])))
```

**Configuration:**

- **Legend:** `{{bucket}}`
- **Min Step:** `15s`
- **Unit:** `bool` (0 or 1)
- **Panel Type:** Graph or Table
- **Description:** Identifies write-heavy buckets (1 = more writes than reads)

---

### 6.6 Most Reliable Bucket

**Query:**

```promql
bottomk(1, sum by (bucket) (rate(rr_s3_operations_total{status="error"}[5m])) / sum by (bucket) (rate(rr_s3_operations_total[5m])) * 100)
```

**Configuration:**

- **Legend:** `{{bucket}}`
- **Min Step:** `15s`
- **Unit:** `percent (0-100)`
- **Panel Type:** Stat
- **Description:** Bucket with lowest error rate

---

### 6.7 Least Reliable Bucket

**Query:**

```promql
topk(1, sum by (bucket) (rate(rr_s3_operations_total{status="error"}[5m])) / sum by (bucket) (rate(rr_s3_operations_total[5m])) * 100)
```

**Configuration:**

- **Legend:** `{{bucket}}`
- **Min Step:** `15s`
- **Unit:** `percent (0-100)`
- **Panel Type:** Stat
- **Thresholds:** Red > 5%, Yellow 1-5%, Green < 1%
- **Description:** Bucket with highest error rate

---

## 7. Dashboard Layout Recommendations

### Row 1: Key Metrics Overview (4 panels)

1. **Total Operations/sec** - Stat panel
2. **Success Rate %** - Gauge with thresholds
3. **Total Errors/sec** - Stat panel with threshold colors
4. **Active Buckets** - Stat (count of buckets with ops > 0)

### Row 2: Operation Analysis (2 panels)

1. **Operations by Type** - Stacked area graph
2. **Operations by Bucket** - Graph (time series)

### Row 3: Success vs Errors (2 panels)

1. **Success vs Error Rate** - Stacked area graph
2. **Operation Success Rate by Type** - Table

### Row 4: Error Analysis (2 panels)

1. **Errors by Type** - Pie chart or Stacked area
2. **Most Error-Prone Buckets** - Bar gauge

### Row 5: Bucket Performance (1 panel)

1. **Bucket Performance Table** - Table with multiple queries (Total OPS, Write OPS, Read OPS, Error %)

### Row 6: Advanced (2 panels)

1. **Error Heatmap (Bucket vs Type)** - Heatmap
2. **Read/Write Ratio by Bucket** - Bar gauge

---

## 8. Unit Reference Guide

### Standard Grafana Units

**Rate Units:**

- `ops (operations/sec)` - for operation rates
- `errors/sec` - for error rates

**Percentage:**

- `percent (0-100)` - displays as 95%
- `percentunit (0.0-1.0)` - displays 0.95 as 95%

**Count:**

- `short` - auto-formats large numbers (1K, 1M)
- `none` - raw number

**Boolean:**

- `bool` - 0 or 1
- `bool_yes_no` - displays as Yes/No

---

## 9. Common Threshold Configurations

### Error Rate Thresholds

```
Green: < 1%
Yellow: 1-5%
Red: > 5%
```

### Success Rate Thresholds

```
Red: < 95%
Yellow: 95-99%
Green: > 99%
```

### Error Burst Detection

```
Red: value = 1 (burst detected)
Green: value = 0 (normal)
```

---

## 10. Alert Rules (Prometheus)

### Critical Alerts

**High Error Rate:**

```yaml
- alert: S3HighErrorRate
  expr: sum(rate(rr_s3_operations_total{status="error"}[5m])) / sum(rate(rr_s3_operations_total[5m])) * 100 > 5
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: "S3 plugin error rate above 5%"
    description: "Error rate is {{ $value }}% (threshold: 5%)"
```

**Bucket Not Found:**

```yaml
- alert: S3BucketNotFoundErrors
  expr: sum(rate(rr_s3_errors_total{error_type="BUCKET_NOT_FOUND"}[5m])) > 0
  for: 2m
  labels:
    severity: critical
  annotations:
    summary: "S3 bucket not found errors detected"
    description: "Configuration issue: bucket references don't exist"
```

**Permission Denied:**

```yaml
- alert: S3PermissionDenied
  expr: sum(rate(rr_s3_errors_total{error_type="PERMISSION_DENIED"}[5m])) > 0
  for: 2m
  labels:
    severity: critical
  annotations:
    summary: "S3 permission denied errors"
    description: "Credentials or IAM policy issue detected"
```

### Warning Alerts

**Elevated Error Rate:**

```yaml
- alert: S3ElevatedErrorRate
  expr: sum(rate(rr_s3_operations_total{status="error"}[5m])) / sum(rate(rr_s3_operations_total[5m])) * 100 > 1
  for: 10m
  labels:
    severity: warning
  annotations:
    summary: "S3 plugin error rate elevated"
    description: "Error rate is {{ $value }}% (threshold: 1%)"
```

**Timeout Errors:**

```yaml
- alert: S3OperationTimeouts
  expr: sum(rate(rr_s3_errors_total{error_type="OPERATION_TIMEOUT"}[5m])) > 0
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "S3 operation timeouts detected"
    description: "Network or S3 service performance issue"
```

---

## 11. Example Grafana Dashboard JSON

See `grafana_dashboard_example.json` for a complete pre-configured dashboard including all key metrics and recommended visualizations.

---

## 12. Troubleshooting Guide

### No Metrics Appearing

1. Verify metrics plugin is enabled in `.rr.yaml`:
   ```yaml
   metrics:
     address: 127.0.0.1:2112
   ```

2. Check metrics endpoint:
   ```bash
   curl http://localhost:2112/metrics | grep rr_s3
   ```

3. Ensure S3 plugin is performing operations (metrics only appear after first operation)

### High Error Rates

1. Check error types with:
   ```promql
   sum by (error_type) (rate(rr_s3_errors_total[5m]))
   ```

2. Identify problematic buckets:
   ```promql
   topk(5, sum by (bucket) (rate(rr_s3_errors_total[5m])))
   ```

3. Review RoadRunner logs for detailed error messages

### Permission Issues

- Check AWS credentials configuration
- Verify IAM policy has required permissions
- Review `PERMISSION_DENIED` error rate by bucket

### Performance Degradation

- Monitor operation rate trends (Hour over Hour)
- Check for error bursts
- Review specific operation types (write/read/list) for bottlenecks

---

## 13. Integration with Other RoadRunner Metrics

### Combined RoadRunner + S3 Dashboard

**HTTP Traffic vs S3 Operations:**

```promql
# HTTP RPS
sum(rate(rr_http_requests_total[5m]))

# S3 OPS
sum(rate(rr_s3_operations_total[5m]))
```

**Worker Pool vs S3 Activity:**

```promql
# Worker utilization
rr_http_worker_utilization_percent

# S3 write operations
sum(rate(rr_s3_operations_total{operation="write"}[5m]))
```

This correlation helps identify if S3 operations are causing worker pool pressure.

---

## 14. Best Practices

### Query Performance

- Use `rate()` over `irate()` for smoother graphs
- Set appropriate `[5m]` intervals based on traffic volume
- Use `topk()` to limit cardinality in busy systems

### Alerting

- Set up critical alerts for BUCKET_NOT_FOUND (config issues)
- Monitor PERMISSION_DENIED (security issues)
- Alert on error rate >5% sustained for 5 minutes

### Dashboard Organization

- Group metrics by concern (operations, errors, buckets)
- Use consistent color schemes across panels
- Include both rate and percentage views

### Retention

- Default Prometheus retention: 15 days
- Increase for long-term S3 usage analysis
- Consider recording rules for long-term aggregations

---

## 15. Recording Rules (Optional Optimization)

For high-traffic systems, pre-compute common queries:

```yaml
groups:
  - name: s3_recording_rules
    interval: 30s
    rules:
      - record: s3:operations:rate5m
        expr: sum(rate(rr_s3_operations_total[5m]))
      
      - record: s3:operations:rate5m:by_bucket
        expr: sum by (bucket) (rate(rr_s3_operations_total[5m]))
      
      - record: s3:errors:rate5m
        expr: sum(rate(rr_s3_errors_total[5m]))
      
      - record: s3:success_rate:percent
        expr: sum(rate(rr_s3_operations_total{status="success"}[5m])) / sum(rate(rr_s3_operations_total[5m])) * 100
```

Then query using: `s3:operations:rate5m` instead of full expression.

---

For additional support or questions about S3 plugin metrics, refer to the main RoadRunner documentation or open an issue on GitHub.
