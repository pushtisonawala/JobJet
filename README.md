#  JobJet - Distributed Job Queue System

> **A powerful, scalable job queue system with Kubernetes native support, distributed tracing, and comprehensive monitoring.**

JobJet is a modern job queue system that can run both as a standalone application and in Kubernetes. It features a powerful CLI tool, real-time monitoring with Jaeger and Prometheus, and supports various job types from simple email sending to complex data processing tasks.

##  Table of Contents

- [ Features](#-features)
- [ Architecture](#️-architecture)
- [ Quick Start](#-quick-start)
- [ Installation](#-installation)
- [ Usage Examples](#️-usage-examples)
- [ Monitoring & Observability](#-monitoring--observability)
- [ Kubernetes Integration](#️-kubernetes-integration)
- [ JobJet CLI](#-jobjet-cli)
- [ Configuration](#-configuration)
- [ API Reference](#-api-reference)
- [ Contributing](#-contributing)

##  Features

- ** Multiple Deployment Modes**: Run standalone or in Kubernetes
- ** Powerful CLI**: `jobjet` command-line tool for job management
- ** Distributed Tracing**: Full observability with Jaeger integration
- ** Metrics & Monitoring**: Prometheus metrics and Grafana dashboards
- ** High Performance**: Redis-backed queue with efficient job processing
- ** Retry Logic**: Automatic retry with exponential backoff
- ** Multiple Job Types**: Email, notifications, data processing, and more
- ** Kubernetes Native**: Custom Resource Definitions (CRDs) support
- ** Dead Letter Queue**: Handle failed jobs gracefully

##  Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   JobJet CLI    │    │   REST API      │    │   Web UI        │
│   (jobjet)      │    │   (Port 8000)   │    │   (Optional)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
         ┌───────────────────────▼───────────────────────┐
         │              Job Queue System                 │
         │              (Redis Backend)                  │
         └───────────────────────┬───────────────────────┘
                                 │
    ┌─────────────────┬──────────┼──────────┬─────────────────┐
    │                 │          │          │                 │
┌───▼───┐    ┌────▼────┐    ┌───▼───┐    ┌▼────┐    ┌──────▼──────┐
│Worker1│    │ Worker2 │    │Worker3│    │ ... │    │ K8s Controller│
│       │    │         │    │       │    │     │    │              │
└───────┘    └─────────┘    └───────┘    └─────┘    └─────────────┘
     │             │             │          │              │
     └─────────────┼─────────────┼──────────┼──────────────┘
                   │             │          │
         ┌─────────▼─────────────▼──────────▼─────────┐
         │           Monitoring Stack                 │
         │   Jaeger + Prometheus + Grafana           │
         └───────────────────────────────────────────┘
```

##  Quick Start

### Prerequisites
- Go 1.21+ 
- Docker & Docker Compose
- Kubernetes cluster (for K8s features)
- Redis server
- MongoDB (for job persistence)

### 1. Clone and Setup
```bash
git clone <repository-url>
cd jobqueue

# Install dependencies
go mod tidy
```

### 2. Start Infrastructure
```bash
# Start Redis, MongoDB, and Jaeger
docker-compose up -d
```

### 3. Start the Job Queue System
```bash
# Terminal 1: Start the API server
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318 go run main.go

# Terminal 2: Start a worker
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318 go run worker2/main.go
```

### 4. Build and Use the CLI
```bash
# Build the CLI
cd jobjet-cli
go build -o jobjet .

# Submit your first job!
./jobjet submit email --payload '{"to": "user@example.com", "subject": "Welcome!", "body": "Hello from JobJet!"}'
```
<img width="946" height="801" alt="image" src="https://github.com/user-attachments/assets/2e15f6c2-1650-46bc-a8f4-97793fbcea6f" />


 **Congratulations!** Your JobJet system is now running. Visit http://localhost:16686 to see your job traces in Jaeger!

## Installation

### Option 1: Standalone Installation

```bash
# 1. Start required services
docker-compose up -d redis mongodb jaeger

# 2. Run the application
go run main.go
```

### Option 2: Docker Installation

```bash
# Build and run with Docker
docker-compose up -d
```

### Option 3: Kubernetes Installation

```bash
# Apply CRDs and deploy
kubectl apply -f jobdefinition-crd.yaml
kubectl apply -f jobqueue-app-deployment.yaml
kubectl apply -f jobqueue-controller-deployment.yaml
```

## 🛠️ Usage Examples

### Example 1: Sending an Email Job

**Using the API:**
```bash
curl -X POST http://localhost:8000/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "type": "email",
    "payload": {
      "to": "customer@example.com",
      "subject": "Order Confirmation",
      "body": "Thank you for your order! Your items will be shipped soon."
    }
  }'
```

**Using the CLI:**
```bash
./jobjet submit email --payload '{
  "to": "customer@example.com", 
  "subject": "Order Confirmation",
  "body": "Thank you for your order!"
}'
```

### Example 2: Data Processing Job

```bash
./jobjet submit data-processing --payload '{
  "dataset": "user_analytics",
  "operation": "aggregate",
  "timeframe": "last_30_days",
  "output_format": "parquet"
}'
```

### Example 3: Image Processing Job

```bash
./jobjet submit image-processing --payload '{
  "image_url": "https://example.com/photo.jpg",
  "operations": ["resize:800x600", "compress:80", "watermark"],
  "output_bucket": "processed-images"
}'
```

### Example 4: Notification Job

```bash
./jobjet submit notification --payload '{
  "user_id": 12345,
  "message": "Your subscription expires in 3 days",
  "channels": ["email", "sms", "push"],
  "priority": "high"
}'

```
<img width="684" height="290" alt="image" src="https://github.com/user-attachments/assets/1bab99e5-a433-4eac-b0bf-5878ab939a77" />



##  Monitoring & Observability

### Jaeger Tracing

JobJet includes comprehensive distributed tracing:

1. **Start Jaeger:**
   ```bash
   docker-compose -f monitoring/jaeger-compose.yml up -d
   ```

2. **View Traces:**
   - Open http://localhost:16686
   - Search for services: `jaeger-all-in-one`
   - Filter by operation: `POST /jobs`, `process_job`, `job_execution`
     

### Prometheus Metrics

Available metrics:
- `jobqueue_jobs_total` - Total jobs processed
- `jobqueue_jobs_duration` - Job processing time
- `jobqueue_queue_length` - Current queue length
- `jobqueue_retry_attempts` - Retry attempts per job

Access metrics at: http://localhost:2112/metrics
<img width="1816" height="899" alt="image" src="https://github.com/user-attachments/assets/1dccd845-ca74-43be-bb1b-85c829b4d55d" />


### Grafana Dashboard

```bash
# Start Grafana
docker-compose up -d grafana

# Access dashboard
open http://localhost:3000
# Login: admin/admin
```
<img width="1510" height="786" alt="image" src="https://github.com/user-attachments/assets/c5d5e52f-3812-4dc1-9b02-13dc4ce64994" />

##  Kubernetes Complete Setup & Integration

### Prerequisites
- Kubernetes cluster (kind, minikube, or cloud)
- kubectl configured and working
- Docker for building images

### Step 1: Create Kubernetes Cluster
```bash
# Option 1: Using kind (recommended for local development)
kind create cluster --name jobjet

# Option 2: Using minikube
minikube start

# Verify cluster connection
kubectl cluster-info
```

### Step 2: Install JobJet Custom Resource Definitions (CRDs)
```bash
# Apply the JobDefinition CRD
kubectl apply -f jobdefinition-crd.yaml

# Verify CRD installation
kubectl get crd | grep jobdefinition
```

### Step 3: Setup External Services
```bash
# Start Redis and MongoDB outside K8s (easier for development)
docker-compose up -d redis mongodb

# Or deploy them in Kubernetes
kubectl apply -f external-services.yaml
```

### Step 4: Build and Load Container Images
```bash
# Build all images
docker build -t jobqueue-app:latest .
docker build -t jobqueue-worker:latest -f worker2/Dockerfile .
docker build -t jobqueue-controller:latest -f controller.Dockerfile .

# For kind clusters, load images
kind load docker-image jobqueue-app:latest --name jobjet
kind load docker-image jobqueue-worker:latest --name jobjet
kind load docker-image jobqueue-controller:latest --name jobjet
```

### Step 5: Deploy JobJet Components
```bash
# Deploy API server
kubectl apply -f jobqueue-app-deployment.yaml

# Deploy Controller (processes JobDefinitions)
kubectl apply -f jobqueue-controller-deployment.yaml

# Deploy Workers
kubectl apply -f jobqueue-worker2-deployment.yaml

# Check deployment status
kubectl get pods -w
```
<img width="898" height="400" alt="image" src="https://github.com/user-attachments/assets/eddaba93-46b1-4499-b3e1-46a0d0d5ed12" />


### Step 6: Expose Services
```bash
# Port forward to access the API
kubectl port-forward service/jobqueue-app 8000:8000 &

# Or use LoadBalancer/NodePort in cloud environments
```

### Kubernetes Job Examples

#### Example 1: Simple Email Job
```bash
# Create a JobDefinition using kubectl
kubectl apply -f - <<EOF
apiVersion: jobjet.dev/v1
kind: JobDefinition
metadata:
  name: welcome-email-job
  namespace: default
spec:
  queue: email
  payload:
    to: "new-user@company.com"
    subject: "Welcome to Our Platform!"
    body: "Thank you for joining us. Your account is now active."
    priority: "high"
EOF
```

#### Example 2: Batch Data Processing Job
```bash
kubectl apply -f - <<EOF
apiVersion: jobjet.dev/v1
kind: JobDefinition
metadata:
  name: daily-report-job
spec:
  queue: data-processing
  payload:
    dataset: "user_analytics"
    operation: "daily_summary"
    date: "2026-02-03"
    output_format: "csv"
    destination: "s3://reports/daily/"
EOF
```

#### Example 3: Multiple Jobs at Once
```bash
kubectl apply -f - <<EOF
apiVersion: jobjet.dev/v1
kind: JobDefinition
metadata:
  name: user-notification-batch
spec:
  queue: notification
  payload:
    campaign_id: "welcome_series_001"
    user_segments: ["new_users", "trial_users"]
    message: "Don't forget to complete your profile!"
    channels: ["email", "push"]
---
apiVersion: jobjet.dev/v1
kind: JobDefinition
metadata:
  name: image-resize-job
spec:
  queue: image-processing
  payload:
    image_urls: [
      "https://example.com/photo1.jpg",
      "https://example.com/photo2.jpg"
    ]
    operations: ["resize:1024x768", "compress:85", "watermark"]
    output_bucket: "processed-images"
EOF
```
<img width="700" height="626" alt="image" src="https://github.com/user-attachments/assets/7b075f5b-7e53-4dbc-bcbf-6bd4c02c0044" />


### Monitoring Kubernetes Jobs

```bash
# List all JobDefinitions
kubectl get jobdefinitions

# Get detailed information
kubectl describe jobdefinition welcome-email-job

# Check if Jobs were created from JobDefinitions
kubectl get jobs

# View job pods and logs
kubectl get pods -l jobdefinition=welcome-email-job
kubectl logs -l jobdefinition=welcome-email-job

# Monitor job controller logs
kubectl logs -l app=jobqueue-controller -f
```
<img width="874" height="828" alt="image" src="https://github.com/user-attachments/assets/fb739ab4-137a-4553-8cc9-43f70d3aa6e3" />


### Scaling in Kubernetes

```bash
# Scale workers horizontally
kubectl scale deployment jobqueue-worker --replicas=5

# Set up horizontal pod autoscaling
kubectl autoscale deployment jobqueue-worker \
  --min=2 --max=10 --cpu-percent=70

# Scale the API service
kubectl scale deployment jobqueue-app --replicas=3
```
<img width="870" height="635" alt="image" src="https://github.com/user-attachments/assets/00df7611-5788-424a-9382-d732946d1720" />


##  JobJet CLI Complete Guide

### Installation & Setup
```bash
# Navigate to CLI directory
cd jobjet-cli

# Build the CLI
go build -o jobjet .

# Make it globally accessible (optional)
sudo cp jobjet /usr/local/bin/  # Linux/Mac
# Or add to PATH on Windows

# Verify installation
./jobjet --help
```

### CLI Configuration

#### Option 1: Configuration File
```bash
# Create config file
mkdir -p ~/.config/jobjet
cat > ~/.config/jobjet/config.yaml << EOF
api-url: "http://localhost:8000"
output: "table"
timeout: 30s
namespace: "default"
kubeconfig: "$HOME/.kube/config"
EOF
```

#### Option 2: Environment Variables
```bash
export JOBJET_API_URL="http://localhost:8000"
export JOBJET_NAMESPACE="production"
export JOBJET_OUTPUT="json"
export KUBECONFIG="$HOME/.kube/config"
```

### Complete CLI Command Examples

#### `submit` - Job Submission Examples

**Basic Job Submission:**
```bash
# Email job
./jobjet submit email --payload '{
  "to": "customer@example.com",
  "subject": "Order Shipped!",
  "body": "Your order #12345 has been shipped and will arrive tomorrow."
}'

# Notification job
./jobjet submit notification --payload '{
  "user_id": 789,
  "message": "Your subscription expires in 3 days",
  "channels": ["email", "sms", "push"],
  "priority": "high"
}'
```

**Advanced Job Submission:**
```bash
# Data processing with options
./jobjet submit data-processing \
  --payload '{
    "dataset": "sales_2026_q1",
    "operation": "aggregate_revenue", 
    "groupby": ["region", "product_category"],
    "date_range": {"start": "2026-01-01", "end": "2026-03-31"}
  }' \
  --priority 8 \
  --timeout 1800 \
  --retries 3

# Image processing job
./jobjet submit image-processing \
  --payload '{
    "source_bucket": "user-uploads",
    "images": ["profile_123.jpg", "cover_456.png"],
    "operations": [
      {"type": "resize", "width": 800, "height": 600},
      {"type": "compress", "quality": 85},
      {"type": "watermark", "text": "© 2026 Company"}
    ],
    "output_bucket": "processed-images"
  }'
```

**Job Submission from File:**
```bash
# Create job file
cat > complex-job.json << EOF
{
  "campaign_name": "spring_promo_2026",
  "target_segments": ["premium_users", "active_users"],
  "email_template": "spring_promo_v2",
  "personalization": {
    "include_name": true,
    "include_recommendations": true
  },
  "schedule": {
    "send_time": "2026-02-04T09:00:00Z",
    "timezone_aware": true
  },
  "analytics": {
    "track_opens": true,
    "track_clicks": true,
    "attribution_campaign": "spring_2026"
  }
}
EOF

# Submit from file
./jobjet submit marketing --payload @complex-job.json
```

#### `list` - Queue Monitoring Examples

```bash
# Basic list
./jobjet list

# JSON output for scripting
./jobjet list --output json

# YAML output
./jobjet list --output yaml

# Filter by specific API endpoint
./jobjet list --api-url http://production-api:8000
```

#### `logs` - Job Log Management

```bash
# View recent logs
./jobjet logs email-1770131662641564100

# Stream logs in real-time
./jobjet logs notification-1770131674494908000 --follow

# Show timestamps
./jobjet logs data-processing-1770131700000000 --timestamps

# Last 100 lines
./jobjet logs image-processing-1770131720000000 --tail 100

# Logs since 10 minutes ago
./jobjet logs analytics-1770131740000000 --since 10m

# Logs with specific container (if multiple)
./jobjet logs batch-job-1770131760000000 --container worker
```

#### `describe` - Detailed Job Information

```bash
# Basic job description
./jobjet describe email-1770131662641564100

# JSON format for API integration
./jobjet describe notification-1770131674494908000 --output json

# YAML format
./jobjet describe data-processing-1770131700000000 --output yaml
```
<img width="649" height="607" alt="image" src="https://github.com/user-attachments/assets/4af1e4a0-1d00-46e4-a936-66dd9cc44089" />


#### Advanced CLI Usage Examples

**Batch Operations:**
```bash
# Submit multiple jobs in sequence
for region in us-east us-west eu-central; do
  ./jobjet submit analytics --payload "{
    \"region\": \"$region\",
    \"report_type\": \"daily_summary\",
    \"date\": \"$(date -I)\"
  }"
done

# Monitor multiple jobs
job_ids=($(./jobjet list --output json | jq -r '.jobs[].id'))
for job_id in "${job_ids[@]}"; do
  echo "=== Job: $job_id ==="
  ./jobjet describe "$job_id"
done
```

**Production Monitoring Scripts:**
```bash
#!/bin/bash
# monitor-jobs.sh - Monitor job queue health

# Check queue length
queue_info=$(./jobjet list --output json)
queue_length=$(echo "$queue_info" | jq '.count')

if [ "$queue_length" -gt 100 ]; then
  echo " WARNING: Queue length is $queue_length"
  # Scale workers
  kubectl scale deployment jobqueue-worker --replicas=10
fi

# Check for failed jobs
failed_jobs=$(./jobjet list --output json | jq '.jobs[] | select(.status == "failed")')
if [ -n "$failed_jobs" ]; then
  echo "Failed jobs detected:"
  echo "$failed_jobs" | jq -r '.id'
fi
```

## 🔍 Jaeger Distributed Tracing Complete Setup

### Jaeger Installation & Configuration

#### Option 1: Docker Compose (Recommended for Development)
```bash
# Start Jaeger with proper configuration
docker-compose -f monitoring/jaeger-compose.yml up -d

# Verify Jaeger is running
curl -s http://localhost:16686/api/services | jq '.'
```

#### Option 2: Kubernetes Deployment
```bash
# Deploy Jaeger operator
kubectl create namespace observability
kubectl apply -f https://github.com/jaegertracing/jaeger-operator/releases/download/v1.51.0/jaeger-operator.yaml -n observability

# Deploy Jaeger instance
kubectl apply -f - <<EOF
apiVersion: jaegertracing.io/v1
kind: Jaeger
metadata:
  name: jobjet-tracing
  namespace: observability
spec:
  strategy: allInOne
  allInOne:
    image: jaegertracing/all-in-one:1.52
    options:
      log-level: debug
  storage:
    type: memory
    options:
      memory:
        max-traces: 10000
  ui:
    options:
      dependencies:
        menuEnabled: false
EOF

# Port forward to access UI
kubectl port-forward -n observability service/jobjet-tracing-query 16686:16686
```

### Configure Applications for Tracing

#### Environment Variables for All Components
```bash
# Set these for all JobJet components
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
export OTEL_RESOURCE_ATTRIBUTES="service.name=jobqueue-api,service.version=1.0.0"
export OTEL_TRACE_SAMPLER=always_on

# For Kubernetes deployments, add to deployment manifests
env:
- name: OTEL_EXPORTER_OTLP_ENDPOINT
  value: "http://jaeger-collector:4318"
- name: OTEL_RESOURCE_ATTRIBUTES
  value: "service.name=jobqueue-api,service.version=1.0.0,k8s.namespace=default"
```

### Start All Components with Tracing
```bash
# Terminal 1: Start API with tracing
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318 \
OTEL_RESOURCE_ATTRIBUTES="service.name=jobqueue-api" \
go run main.go

# Terminal 2: Start Worker with tracing  
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318 \
OTEL_RESOURCE_ATTRIBUTES="service.name=jobqueue-worker" \
go run worker2/main.go

# Terminal 3: Start Controller with tracing
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318 \
OTEL_RESOURCE_ATTRIBUTES="service.name=jobqueue-controller" \
go run k8s-integration/controller/main.go
```

### Generate Traces with Real Examples

#### Example 1: Complete Email Job Trace
```bash
# Submit job that will create full trace
curl -X POST http://localhost:8000/jobs \
  -H "Content-Type: application/json" \
  -H "X-Trace-Id: email-trace-$(date +%s)" \
  -d '{
    "type": "email",
    "payload": {
      "to": "customer@example.com",
      "subject": "Welcome Email with Tracing",
      "body": "This email job will create a complete distributed trace!",
      "template_id": "welcome_template_v2",
      "personalization": {
        "user_name": "John Doe",
        "account_type": "premium"
      }
    },
    "priority": 5,
    "retries": 3
  }'
```

#### Example 2: Complex Multi-Service Trace
```bash
# Submit data processing job that touches multiple services
./jobjet submit data-processing --payload '{
  "pipeline_name": "user_analytics_trace",
  "steps": [
    {"service": "data_extractor", "source": "postgresql://users"},
    {"service": "data_transformer", "operations": ["clean", "normalize"]},
    {"service": "ml_processor", "model": "user_segmentation_v3"},
    {"service": "result_writer", "destination": "s3://analytics/results/"}
  ],
  "trace_metadata": {
    "experiment_id": "trace_test_001",
    "user_id": "analyst_123"
  }
}'
```

#### Example 3: Batch Job Tracing
```bash
# Submit multiple related jobs to see trace correlation
for i in {1..5}; do
  ./jobjet submit notification --payload "{
    \"batch_id\": \"batch_trace_$(date +%s)\",
    \"user_id\": $((1000 + i)),
    \"message\": \"Batch notification $i of 5\",
    \"channels\": [\"email\", \"push\"],
    \"correlation_id\": \"batch_trace_$(date +%s)\"
  }"
  sleep 1
done
```

### Jaeger UI Navigation & Analysis

#### Access and Navigate Jaeger UI
```bash
# Open Jaeger UI
open http://localhost:16686

# Or use curl to check services programmatically
curl -s http://localhost:16686/api/services | jq '.data[].name'
```

#### Key Traces to Look For

1. **Service List** - You should see:
   - `jobqueue-api` - API request traces
   - `jobqueue-worker` - Job processing traces  
   - `jobqueue-controller` - Kubernetes controller traces

2. **Operations to Filter By:**
   - `POST /jobs` - Job submission
   - `process_job` - Job execution
   - `job_execution` - Individual job processing
   - `redis_operation` - Queue operations
   - `k8s_controller_reconcile` - Controller operations

3. **Trace Analysis Examples:**
   ```bash
   # Find traces for specific operations
   # In Jaeger UI, use these filters:
   
   # All email jobs in last hour
   Service: jobqueue-api
   Operation: POST /jobs
   Tags: job.type="email"
   Lookback: 1h
   
   # Failed job processing
   Service: jobqueue-worker
   Operation: process_job
   Tags: error=true
   
   # Long-running jobs (> 5 seconds)
   Min Duration: 5s
   ```
   <img width="1894" height="776" alt="image" src="https://github.com/user-attachments/assets/9892ac6b-98cc-4d4b-a352-e045e33e7e80" />


#### Advanced Jaeger Queries
```bash
# Query Jaeger API directly
# Get all services
curl http://localhost:16686/api/services

# Get operations for jobqueue-api
curl "http://localhost:16686/api/services/jobqueue-api/operations"

# Search traces with parameters
curl "http://localhost:16686/api/traces?service=jobqueue-api&operation=POST%20/jobs&limit=20&start=$(date -d '1 hour ago' +%s)000000&end=$(date +%s)000000"

# Find traces by tags
curl "http://localhost:16686/api/traces?service=jobqueue-worker&tag=job.type:email&limit=10"
```


### Trace Correlation Examples

#### Example 1: Full Request Lifecycle Trace
```bash
# This single request should create spans across all services
curl -X POST http://localhost:8000/jobs \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: full-lifecycle-trace-001" \
  -d '{
    "type": "complex_workflow",
    "payload": {
      "workflow_steps": [
        {"step": "validate_data", "timeout": 30},
        {"step": "process_payment", "timeout": 60},
        {"step": "send_confirmation", "timeout": 15},
        {"step": "update_inventory", "timeout": 45}
      ]
    }
  }'

# Expected trace flow:
# jobqueue-api -> Redis (job submission)
# -> jobqueue-worker (job pickup)
# -> Multiple child spans for each workflow step
# -> Final span for job completion
```

#### Example 2: Error Tracing
```bash
# Submit a job that will likely fail to see error traces
./jobjet submit email --payload '{
  "to": "invalid-email-address",
  "subject": "This should fail",
  "body": "Testing error tracing",
  "force_error": true
}'

# In Jaeger UI, look for:
# - Spans with error=true tags
# - Exception details in span logs
# - Retry attempts in trace timeline
```

This complete setup gives you full observability into your JobJet system with detailed tracing across all components!

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `REDIS_URL` | Redis connection string | `127.0.0.1:6379` |
| `MONGO_URL` | MongoDB connection string | `mongodb://localhost:27017` |
| `PORT` | API server port | `8000` |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | OpenTelemetry endpoint | `http://localhost:4318` |
| `JOBJET_NAMESPACE` | Kubernetes namespace | `default` |

### Job Types and Handlers

JobJet supports multiple job types out of the box:

1. **Email Jobs** (`email`)
   ```json
   {
     "to": "user@example.com",
     "subject": "Subject line",
     "body": "Email content",
     "attachments": ["file1.pdf"]
   }
   ```

2. **Notification Jobs** (`notification`)
   ```json
   {
     "user_id": 12345,
     "message": "Your order is ready!",
     "channels": ["email", "sms", "push"],
     "priority": "high"
   }
   ```

3. **Data Processing Jobs** (`data-processing`)
   ```json
   {
     "dataset": "sales_data",
     "operation": "aggregate",
     "filters": {"date_range": "last_7_days"},
     "output": {"format": "csv", "destination": "s3://reports/"}
   }
   ```

4. **Image Processing Jobs** (`image-processing`)
   ```json
   {
     "image_url": "https://example.com/image.jpg",
     "operations": ["resize:800x600", "compress:85"],
     "output_bucket": "processed-images"
   }
   ```

## API Reference

### Submit Job
```http
POST /jobs
Content-Type: application/json

{
  "type": "job-type",
  "payload": { ... },
  "priority": 5,
  "retries": 3,
  "timeout": 300
}
```

**Response:**
```json
{
  "job_id": "uuid-here",
  "message": "job queued",
  "job_queue_len": 5,
  "redis_addr": "127.0.0.1:6379"
}
```

### Debug Endpoints

#### Check Redis Status
```http
GET /debug/redis
```

#### View Job Queue
```http
GET /debug/jobqueue
```

### Health Check
```http
GET /health
```

##  Troubleshooting

### Common Issues

#### 1. "Connection refused" errors
```bash
# Check if Redis is running
redis-cli ping

# Check if MongoDB is running  
mongosh --eval "db.stats()"

# Restart services
docker-compose restart redis mongodb
```

#### 2. Jobs not processing
```bash
# Check worker logs
docker-compose logs worker

# Check queue length
curl http://localhost:8000/debug/redis

# Restart worker
docker-compose restart worker
```

#### 3. Tracing not working
```bash
# Check Jaeger is running
curl http://localhost:16686/api/services

# Verify OTLP endpoint
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
```

#### 4. Kubernetes jobs not creating
```bash
# Check CRDs are installed
kubectl get crd | grep jobdefinition

# Check controller logs
kubectl logs -l app=jobqueue-controller

# Verify JobDefinitions
kubectl get jobdefinitions
```

### Debugging Tips

1. **Enable Debug Logging:**
   ```bash
   export LOG_LEVEL=debug
   go run main.go
   ```

2. **Check All Services:**
   ```bash
   docker-compose ps
   kubectl get pods
   ```

3. **Monitor Metrics:**
   ```bash
   curl http://localhost:2112/metrics | grep jobqueue
   ```

## Performance Tuning

### Redis Configuration
```bash
# Increase memory limit
redis-cli CONFIG SET maxmemory 2gb
redis-cli CONFIG SET maxmemory-policy allkeys-lru
```

### Worker Scaling
```bash
# Run multiple workers
for i in {1..5}; do
  PORT=$((8001 + $i)) go run worker2/main.go &
done
```

### Kubernetes Scaling
```bash
# Scale workers
kubectl scale deployment jobqueue-worker --replicas=5

# Horizontal Pod Autoscaler
kubectl autoscale deployment jobqueue-worker --min=2 --max=10 --cpu-percent=70
```

##  Contributing

We welcome contributions! Here's how to get started:

1. **Fork the repository**
2. **Create a feature branch:** `git checkout -b feature/amazing-feature`
3. **Make your changes** and add tests
4. **Run the test suite:** `go test ./...`
5. **Submit a pull request**

### Development Setup

```bash
# Install development dependencies
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run tests
go test -v ./...

# Format code
goimports -w .
golangci-lint run
```

##  License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙋‍♂️ Support

- **Documentation:** [Wiki](../../wiki)
- **Issues:** [GitHub Issues](../../issues)
- **Discussions:** [GitHub Discussions](../../discussions)

---

**Built with ❤️ using Go, Kubernetes, and modern observability tools.**

> 💡 **Tip:** Start with the Quick Start guide, then follow the Kubernetes, CLI, and Jaeger sections for complete setup!
