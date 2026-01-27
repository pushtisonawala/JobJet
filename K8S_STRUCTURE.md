# Project Structure

```
JobJet/
│
├── 📦 Kubernetes Integration (NEW)
│   ├── pkg/k8s/
│   │   └── controller.go          # Core controller logic
│   │                               # - SharedInformer setup
│   │                               # - Event handlers (ADD/UPDATE/DELETE)
│   │                               # - Status update logic
│   │
│   ├── cmd/controller/
│   │   └── main.go                 # Controller entry point
│   │                               # - Kubeconfig loading
│   │                               # - Dynamic client creation
│   │                               # - Graceful shutdown
│   │
│   ├── jobdefinition-crd.yaml      # Custom Resource Definition
│   │                               # Defines JobDefinition schema
│   │
│   ├── job.yaml                    # Example JobDefinition resource
│   │                               # For testing the controller
│   │
│   ├── setup-k8s.sh                # Setup script (Linux/Mac)
│   ├── setup-k8s.bat               # Setup script (Windows)
│   ├── K8S_README.md               # Quick start guide
│   └── K8S_LEARNING_GUIDE.md       # Deep dive educational guide
│
├── 📂 Existing JobJet Code
│   ├── main.go                     # Main application
│   ├── queue/                      # Queue implementation
│   ├── worker/                     # Job processing
│   ├── db/                         # Database layer
│   ├── models/                     # Data models
│   ├── controllers/                # HTTP controllers
│   ├── metrics/                    # Prometheus metrics
│   └── ...                         # Other components
│
└── 🐳 Infrastructure
    ├── docker-compose.yaml
    ├── Dockerfile
    ├── k8s/                        # (existing K8s manifests)
    └── grafana/                    # Monitoring dashboards
```

## File Purposes

### Core Controller Files

| File | Lines | Purpose |
|------|-------|---------|
| `pkg/k8s/controller.go` | ~170 | Watches JobDefinitions, handles events, updates status |
| `cmd/controller/main.go` | ~90 | Initializes controller, loads config, runs main loop |
| `jobdefinition-crd.yaml` | ~55 | Tells Kubernetes what a JobDefinition looks like |
| `job.yaml` | ~10 | Example JobDefinition for testing |

### Documentation

| File | Purpose |
|------|---------|
| `K8S_README.md` | Quick start, testing instructions, troubleshooting |
| `K8S_LEARNING_GUIDE.md` | Deep concepts, code flow, common pitfalls, next steps |

### Scripts

| File | Purpose |
|------|---------|
| `setup-k8s.sh` | Automated CRD installation (bash) |
| `setup-k8s.bat` | Automated CRD installation (Windows) |

## How They Connect

```
┌─────────────────────────────────────────────────────────────┐
│                    kubectl apply -f job.yaml                │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│              Kubernetes API Server                          │
│              (validates against jobdefinition-crd.yaml)     │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            │ Watch API
                            ▼
┌─────────────────────────────────────────────────────────────┐
│              cmd/controller/main.go                         │
│              - Loads kubeconfig                             │
│              - Creates dynamic client                       │
│              - Initializes pkg/k8s/controller.go            │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│              pkg/k8s/controller.go                          │
│              - SharedInformer watches JobDefinitions        │
│              - handleAdd() called on new resources          │
│              - Calls JobJet.Enqueue() (your existing code)  │
│              - Updates .status.state via UpdateStatus()     │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│              Your Existing JobJet System                    │
│              - queue/queue.go: Enqueues job                 │
│              - worker/worker.go: Processes job              │
│              - db/jobs.go: Stores in database               │
└─────────────────────────────────────────────────────────────┘
```

## Integration Points

### Where Controller Calls JobJet

```go
// pkg/k8s/controller.go line ~105
jobID, err := c.enqueuer.Enqueue(name)
```

**Currently:** Uses `MockJobJetEnqueuer` for demonstration

**To integrate:** Replace with your real queue:

```go
// cmd/controller/main.go
import "jobjet/queue"

realQueue := queue.NewQueue(redisClient) // Your existing queue
controller := k8s.NewController(dynamicClient, realQueue, "")
```

### What You Need to Implement

```go
// Your queue package should implement:
type JobJetEnqueuer interface {
    Enqueue(jobName string) (jobID string, err error)
}
```

That's it! The controller is designed to plug into your existing system with minimal changes.

## Dependencies Added

```go
// go.mod (new dependencies)
k8s.io/client-go v0.31.0        // Kubernetes client library
k8s.io/apimachinery v0.31.0     // Common Kubernetes types
```

**Total added code:** ~260 lines of Go + ~65 lines of YAML

**Zero generated code** - everything is explicit and readable!
