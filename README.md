# 🚀 JobJet - Distributed Job Queue System

> **A powerful, scalable job queue system with Kubernetes native support, distributed tracing, and comprehensive monitoring.**

JobJet is a modern job queue system that can run both as a standalone application and in Kubernetes. It features a powerful CLI tool, real-time monitoring with Jaeger and Prometheus, and supports various job types from simple email sending to complex data processing tasks.

## 📋 Table of Contents

- [✨ Features](#-features)
- [🏗️ Architecture](#️-architecture)
- [🚀 Quick Start](#-quick-start)
- [📦 Installation](#-installation)
- [🛠️ Usage Examples](#️-usage-examples)
- [📊 Monitoring & Observability](#-monitoring--observability)
- [☸️ Kubernetes Integration](#️-kubernetes-integration)
- [🎯 JobJet CLI](#-jobjet-cli)
- [🔧 Configuration](#-configuration)
- [📚 API Reference](#-api-reference)
- [🤝 Contributing](#-contributing)

## ✨ Features

- **🎯 Multiple Deployment Modes**: Run standalone or in Kubernetes
- **📱 Powerful CLI**: `jobjet` command-line tool for job management
- **🔍 Distributed Tracing**: Full observability with Jaeger integration
- **📊 Metrics & Monitoring**: Prometheus metrics and Grafana dashboards
- **⚡ High Performance**: Redis-backed queue with efficient job processing
- **🔄 Retry Logic**: Automatic retry with exponential backoff
- **🎛️ Multiple Job Types**: Email, notifications, data processing, and more
- **☸️ Kubernetes Native**: Custom Resource Definitions (CRDs) support
- **🚨 Dead Letter Queue**: Handle failed jobs gracefully

## 🏗️ Architecture

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

## 🚀 Quick Start

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

🎉 **Congratulations!** Your JobJet system is now running. Visit http://localhost:16686 to see your job traces in Jaeger!

## 📦 Installation

### Option 1: Standalone Installation

```bash
# 1. Start required services
docker-compose up -d redis mongodb jaeger

# 2. Run the application
go run main.go
git clone <repository-url>
cd jobqueue

# Start Redis and MongoDB services
docker-compose up -d redis mongo

# Build and start the main application
go run main.go
```

### 2. Full Kubernetes Setup

#### Step 1: Start Supporting Services

```bash
# Start Redis and MongoDB
docker-compose up -d redis mongo
```

#### Step 2: Setup Kubernetes Cluster

```bash
# For kind cluster
kind create cluster

# Verify cluster connection
kubectl cluster-info
```

#### Step 3: Install Custom Resource Definitions

```bash
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

## 📊 Monitoring & Observability

### Jaeger Tracing

JobJet includes comprehensive distributed tracing:

1. **Start Jaeger:**
   ```bash
   docker-compose -f monitoring/jaeger-compose.yml up -d
   ```

2. **View Traces:**
   - Open http://localhost:16686
   - Search for services: `jobqueue-api`, `jobqueue-worker`, `jobqueue-controller`
   - Filter by operation: `POST /jobs`, `process_job`, `job_execution`

### Prometheus Metrics

Available metrics:
- `jobqueue_jobs_total` - Total jobs processed
- `jobqueue_jobs_duration` - Job processing time
- `jobqueue_queue_length` - Current queue length
- `jobqueue_retry_attempts` - Retry attempts per job

Access metrics at: http://localhost:2112/metrics

### Grafana Dashboard

```bash
# Start Grafana
docker-compose up -d grafana

# Access dashboard
open http://localhost:3000
# Login: admin/admin
```

## ☸️ Kubernetes Integration

### 1. Install Custom Resource Definitions

```bash
kubectl apply -f jobdefinition-crd.yaml
```

### 2. Deploy JobJet Components

```bash
# Deploy API server
kubectl apply -f jobqueue-app-deployment.yaml

# Deploy Controller
kubectl apply -f jobqueue-controller-deployment.yaml

# Deploy Worker
kubectl apply -f jobqueue-worker2-deployment.yaml
```

### 3. Submit Jobs to Kubernetes

```bash
# Using JobJet CLI (automatically creates K8s JobDefinitions)
./jobjet submit email --payload '{"to": "test@k8s.com", "subject": "K8s Job", "body": "Running in Kubernetes!"}'

# Or apply YAML directly
kubectl apply -f - <<EOF
apiVersion: jobjet.dev/v1
kind: JobDefinition
metadata:
  name: my-email-job
spec:
  queue: email
  payload:
    to: "user@example.com"
    subject: "Hello from K8s"
    body: "This job is running in Kubernetes!"
EOF
```

### 4. Monitor Kubernetes Jobs

```bash
# List JobDefinitions
kubectl get jobdefinitions

# Describe a specific job
kubectl describe jobdefinition my-email-job

# Check actual K8s Jobs created
kubectl get jobs

# View job logs
kubectl logs -l jobdefinition=my-email-job
```

## 🎯 JobJet CLI

### Installation
```bash
cd jobjet-cli
go build -o jobjet .
```

### Commands Overview

#### `submit` - Submit a new job
```bash
# Basic usage
./jobjet submit <job-type> --payload '<json-payload>'

# With file input
./jobjet submit email --payload @email-job.json

# With options
./jobjet submit data-processing \
  --payload '{"dataset": "users"}' \
  --priority 8 \
  --timeout 600 \
  --retries 5
```

#### `list` - List jobs in queue
```bash
# List all jobs
./jobjet list

# JSON output
./jobjet list --output json

# YAML output  
./jobjet list --output yaml
```

#### `logs` - View job logs
```bash
# View logs
./jobjet logs <job-id>

# Follow logs (stream)
./jobjet logs <job-id> --follow

# Show timestamps
./jobjet logs <job-id> --timestamps

# Last N lines
./jobjet logs <job-id> --tail 50
```

#### `describe` - Describe a job
```bash
# Describe job details
./jobjet describe <job-id>

# JSON output
./jobjet describe <job-id> --output json
```

### Configuration

Create `~/.jobjet.yaml`:
```yaml
api-url: "http://localhost:8000"
output: "table"
timeout: 30s
```

Or use environment variables:
```bash
export JOBJET_API_URL="http://localhost:8000"
export JOBJET_NAMESPACE="production"
```

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

## 📚 API Reference

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

## 🚨 Troubleshooting

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

## 🏃‍♂️ Performance Tuning

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

## 🤝 Contributing

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

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙋‍♂️ Support

- **Documentation:** [Wiki](../../wiki)
- **Issues:** [GitHub Issues](../../issues)
- **Discussions:** [GitHub Discussions](../../discussions)

---

**Built with ❤️ using Go, Kubernetes, and modern observability tools.**

> 💡 **Tip:** Start with the Quick Start guide above, then dive into specific sections based on your needs!
# Submit an email job
curl -X POST http://localhost:8080/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "type": "email",
    "payload": {
      "to": "user@example.com",
      "subject": "Welcome to JobJet!",
      "body": "Your job queue is working perfectly!"
    }
  }'

# Response
{
  "job_id": "f491b0d0-7dd9-4714-82e3-5f6eb435c6dd",
  "job_queue_len": 1,
  "message": "job queued",
  "redis_addr": "redis:6379"
}
```

### 2. Kubernetes CRD Job Submission

Create a JobDefinition YAML file:

```yaml
# example-job.yaml
apiVersion: jobjet.dev/v1
kind: JobDefinition
metadata:
  name: email-notification-job
  namespace: default
spec:
  queue: "email-queue"
  payload:
    type: "email"
    to: "admin@company.com"
    subject: "System Alert"
    body: "Your Kubernetes JobJet system is operational!"
    priority: "high"
    retries: 3
```

```bash
# Apply the job definition
kubectl apply -f example-job.yaml

# Check job status
kubectl get jobdefinitions
```

### 3. Multiple Job Types

```yaml
# multiple-jobs.yaml
apiVersion: jobjet.dev/v1
kind: JobDefinition
metadata:
  name: notification-job
spec:
  queue: "notifications"
  payload:
    type: "push_notification"
    device_token: "abc123"
    message: "Your order is ready!"
    
---
apiVersion: jobjet.dev/v1
kind: JobDefinition
metadata:
  name: data-processing-job
spec:
  queue: "data-processing"
  payload:
    type: "analytics"
    dataset: "user_events"
    format: "parquet"
    date_range: "2026-02-01 to 2026-02-03"
    
---
apiVersion: jobjet.dev/v1
kind: JobDefinition
metadata:
  name: payment-retry-job
spec:
  queue: "payments"
  payload:
    type: "payment_retry"
    transaction_id: "txn_123456"
    amount: 99.99
    currency: "USD"
    retries: 5
```

## 🔍 Monitoring & Debugging

### Check Queue Status

```bash
# View job queue status
curl http://localhost:8080/debug/jobqueue

# Check Redis connection
curl http://localhost:8080/debug/redis

# Monitor Kubernetes resources
kubectl get pods,services,jobdefinitions
```

### View Logs

```bash
# Main application logs
kubectl logs -l app=jobqueue-app -f

# Controller logs
kubectl logs -l app=jobqueue-controller -f

# Worker logs
kubectl logs -l app=jobqueue-worker2 -f
```

### Prometheus Metrics

```bash
# Access metrics endpoint
curl http://localhost:8080/metrics
```

## ⚙️ Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `REDIS_ADDR` | Redis connection address | `127.0.0.1:6379` |
| `MONGO_URI` | MongoDB connection string | `mongodb://127.0.0.1:27017/jobjet` |
| `JOBJET_URL` | JobJet API URL for controller | `http://localhost:8000` |
| `GIN_MODE` | Gin framework mode | `debug` |

### Kubernetes Configuration

The system uses ConfigMaps and Secrets for production configuration:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: jobqueue-config
data:
  redis.addr: "redis:6379"
  mongo.uri: "mongodb://mongo:27017/jobjet"
```

## 🏃‍♂️ Development

### Local Development Setup

```bash
# Install dependencies
go mod download

# Start services
docker-compose up -d redis mongo

# Run the application
go run main.go

# Run tests
go test ./...

# Build for production
go build -o jobqueue main.go
```

### Adding New Job Types

1. **Define the job handler** in `worker/job_processor.go`
2. **Update the payload structure** in `models/jobs.go`
3. **Add processing logic** in the worker
4. **Test with both HTTP API and CRDs**

## 🚀 Production Deployment

### Kubernetes Manifests

All deployment manifests are included:

- `jobqueue-app-deployment.yaml` - Main API service
- `jobqueue-controller-deployment.yaml` - CRD controller
- `jobqueue-worker2-deployment.yaml` - Job processors
- `external-services.yaml` - Database connections
- `jobdefinition-crd.yaml` - Custom resource definition

### Scaling

```bash
# Scale workers
kubectl scale deployment jobqueue-worker2 --replicas=5

# Scale API service
kubectl scale deployment jobqueue-app --replicas=3
```

## 📊 Performance

- **Throughput**: 1000+ jobs/second with Redis clustering
- **Latency**: Sub-millisecond job pickup
- **Reliability**: At-least-once delivery with acknowledgments
- **Scalability**: Horizontal scaling with Kubernetes

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔗 Related Projects

- **Celery**: Python distributed task queue
- **BullMQ**: Redis-based queue for Node.js
- **Sidekiq**: Ruby background job processor
- **Kubernetes Jobs**: Native K8s batch processing

---

**Built with ❤️ for Kubernetes-native job processing**
