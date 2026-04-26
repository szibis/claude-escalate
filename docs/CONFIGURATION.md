# Configuration Guide (v0.5.0)

## Overview

Claude Escalate is configured via a single `config.yaml` file. Default location: `~/.claude-escalate/config.yaml`

All settings support **live reload** — changes apply immediately without restarting the service.

## Quick Start

### Minimal Configuration

```yaml
gateway:
  port: 8080
  host: 127.0.0.1

claude_api:
  api_key: ${CLAUDE_API_KEY}

cache:
  enabled: true
  semantic_enabled: true

knowledge_graph:
  enabled: true
  index_paths: [./src, ./lib]
```

## Configuration Sections

### Gateway

```yaml
gateway:
  host: 127.0.0.1
  port: 8080
  log_level: "info"      # debug, info, warn, error
  request_timeout_seconds: 30
```

### Claude API

```yaml
claude_api:
  api_key: ${CLAUDE_API_KEY}
  default_model: "claude-3-5-haiku-20241022"
  timeout_seconds: 60
```

### Cache

```yaml
cache:
  enabled: true
  exact_dedup_enabled: true
  semantic_enabled: true
  semantic_threshold: 0.85      # 0-1 scale
  database: "~/.claude-escalate/cache.db"
  ttl_hours: 168
```

### Knowledge Graph

```yaml
knowledge_graph:
  enabled: true
  index_paths: [./src, ./lib, ./app]
  exclude_patterns: [node_modules, vendor, .venv]
  database: "~/.claude-escalate/graph.db"
  enable_file_watching: true
  debounce_ms: 500
```

### Input Optimization

```yaml
input_optimization:
  enabled: true
  layers:
    tool_stripping: true
    parameter_compression: true
    structured_formatting: true
    whitespace_removal: true
  aggressiveness: "moderate"    # conservative, moderate, aggressive
```

### Security

```yaml
security:
  enabled: true  # CANNOT BE DISABLED
  validate_sql_injection: true
  validate_xss: true
  validate_command_injection: true
  rate_limit_rps: 1000
```

### Metrics & OpenTelemetry

```yaml
metrics:
  enabled: true
  
  export_prometheus: true
  export_json: true
  export_otel: true
  export_logs: true
  
  # OpenTelemetry push endpoint (for production monitoring)
  otel:
    enabled: true
    endpoint: "http://localhost:4317"
    service_name: "claude-escalate"
    environment: "production"
    push_interval_seconds: 30
    
    # Headers for authentication (optional)
    headers:
      # Authorization: "Bearer YOUR_TOKEN"
  
  # Prometheus metrics
  prometheus:
    port: 8080
    path: "/metrics"
  
  # Alerting thresholds
  thresholds:
    cache_false_positive_limit: 0.005  # 0.5%
    min_cache_hit_rate: 0.30
```

### Dashboard

```yaml
dashboard:
  enabled: true
  port: 8080
  refresh_interval_ms: 1000
  dark_mode: "auto"  # auto, light, dark
```

## Environment Variables

Replace values with environment variables:

```bash
export CLAUDE_API_KEY="sk-ant-..."
export OTEL_ENDPOINT="http://otel-collector:4317"
```

```yaml
claude_api:
  api_key: ${CLAUDE_API_KEY}

metrics:
  otel:
    endpoint: ${OTEL_ENDPOINT:-http://localhost:4317}
```

## Examples

### Development

```yaml
gateway:
  log_level: "debug"

cache:
  semantic_threshold: 0.95

knowledge_graph:
  debounce_ms: 200
```

### Enterprise with OTEL Push

```yaml
metrics:
  otel:
    enabled: true
    endpoint: "http://datadog-agent:4317"
    headers:
      Authorization: "Bearer ${DD_API_KEY}"

security:
  rate_limit_per_ip_rpm: 500
```

### Strict Security

```yaml
cache:
  enabled: false

knowledge_graph:
  enabled: false
```

## Management

```bash
# Show configuration
claude-escalate config show

# Validate configuration
claude-escalate config validate

# Update setting (live reload)
claude-escalate config set cache.semantic_threshold 0.90

# View metrics
claude-escalate metrics --now

# Stream metrics to OTEL
# (Configured via config.yaml metrics.otel.enabled)
```

## OTEL Integration

Claude Escalate pushes metrics to OpenTelemetry Collector:

```yaml
metrics:
  otel:
    enabled: true
    endpoint: "http://localhost:4317"      # OTLP gRPC endpoint
    service_name: "claude-escalate"
    push_interval_seconds: 30
```

**Metrics pushed**:
- Request counts and latency
- Cache hit rates
- Token savings
- Security events
- Graph query performance

**Compatible with**:
- Datadog (via OTLP)
- Jaeger
- Prometheus (via OTLP)
- New Relic
- Any OpenTelemetry Collector

## Troubleshooting

```bash
# Validate configuration syntax
claude-escalate config validate --verbose

# Check OTEL connectivity
telnet localhost 4317

# View metrics
curl http://localhost:8080/metrics

# Check logs
tail -f ~/.claude-escalate/gateway.log
```
