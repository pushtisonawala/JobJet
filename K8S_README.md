# JobJet Kubernetes Integration

Minimal Kubernetes controller that watches for `JobDefinition` custom resources and enqueues them to JobJet.

## 🎯 What This Does

1. **Watches** for `JobDefinition` CRs in your k3s cluster
2. **Enqueues** jobs to JobJet when a CR is created  
3. **Updates status** progressively: `Pending → Running → Succeeded`

## 📦 What's Included

```
JobJet/
├── jobdefinition-crd.yaml        # Custom Resource Definition
├── job.yaml                       # Example JobDefinition resource
├── pkg/k8s/controller.go         # Controller logic (SharedInformer)
└── cmd/controller/main.go        # Main entry point
```

## 🚀 Quick Start

### 1. Install the CRD

```bash
kubectl apply -f jobdefinition-crd.yaml
```

Verify it's installed:
```bash
kubectl get crds | grep jobdefinition
```

### 2. Start the Controller

```bash
go run cmd/controller/main.go
```

Expected output:
```
JobJet Kubernetes Controller
=============================
Using kubeconfig: C:\Users\YourName\.kube\config
✓ Connected to Kubernetes cluster
Starting JobDefinition controller...
Waiting for cache sync...
Cache synced. Controller is running.
```

### 3. Create a JobDefinition

In **another terminal**, apply the example job:

```bash
kubectl apply -f job.yaml
```

### 4. Observe the Controller Logs

You should see:
```
▶ JobDefinition created: default/example-job
  Queue: email-queue
  Enqueueing job to JobJet...
[JobJet] Enqueue called for: example-job → JobID: job-example-job-12345
  ✓ Job enqueued with ID: job-example-job-12345
  Updating status.state to 'Running'...
  ✓ Status updated to 'Running'
  Job example-job completed (simulated)
  Updating status.state to 'Succeeded'...
  ✓ Status updated to 'Succeeded'
```

### 5. Check the Status

```bash
kubectl get jobdefinitions -o yaml
```

You should see:
```yaml
status:
  state: Succeeded
```

Or use the shorthand:
```bash
kubectl get jd
```

## 📝 How It Works

### Architecture

```
┌─────────────────┐
│   kubectl apply │
│   (job.yaml)    │
└────────┬────────┘
         │
         ▼
┌─────────────────────────┐
│  Kubernetes API Server  │
│  (JobDefinition CR)     │
└────────┬────────────────┘
         │
         │ watch
         ▼
┌─────────────────────────┐
│  SharedInformer         │
│  (local cache)          │
└────────┬────────────────┘
         │
         │ ADD event
         ▼
┌─────────────────────────┐
│  handleAdd()            │
│  1. Enqueue to JobJet   │
│  2. Update status       │
└─────────────────────────┘
```

### Key Components

**Dynamic Client** (`k8s.io/client-go/dynamic`)
- Works with unstructured data (no code generation)
- Perfect for learning and prototyping

**SharedInformer** (`k8s.io/client-go/tools/cache`)
- Maintains local cache of resources
- Watches for changes efficiently
- Reduces API server load

**Status Subresource**
- Defined in CRD with `subresources.status`
- Updated using `UpdateStatus()` not `Update()`
- Separates spec changes from status changes

## 🔧 Customization

### Change Watched Namespace

In [cmd/controller/main.go](cmd/controller/main.go#L45):

```go
// Watch only default namespace
controller := k8s.NewController(dynamicClient, enqueuer, "default")

// Watch all namespaces (current)
controller := k8s.NewController(dynamicClient, enqueuer, "")
```

### Integrate Real JobJet

Replace `MockJobJetEnqueuer` in [cmd/controller/main.go](cmd/controller/main.go#L19-L27) with your actual implementation:

```go
import "jobjet/queue"

type RealJobJetEnqueuer struct {
    queue *queue.Queue
}

func (r *RealJobJetEnqueuer) Enqueue(jobName string) (string, error) {
    return r.queue.Enqueue(jobName)
}
```

### Adjust Simulated Delay

In [pkg/k8s/controller.go](pkg/k8s/controller.go#L122):

```go
time.Sleep(5 * time.Second)  // Change this value
```

## 🧪 Testing Different Scenarios

### Test Multiple Jobs

```bash
for i in {1..3}; do
  kubectl create -f - <<EOF
apiVersion: jobjet.dev/v1
kind: JobDefinition
metadata:
  name: test-job-$i
spec:
  queue: test-queue
  payload:
    type: test
    retries: 1
EOF
done
```

### Test Different Namespaces

```bash
kubectl create namespace jobjet-test
kubectl apply -f job.yaml -n jobjet-test
```

### Watch Status Changes in Real-Time

```bash
kubectl get jd -w
```

## 🐛 Troubleshooting

### Controller can't connect to cluster

**Error:** `Failed to build kubeconfig`

**Solution:** Ensure `~/.kube/config` exists and points to your k3s cluster

### CRD not found

**Error:** `no matches for kind "JobDefinition"`

**Solution:** Install the CRD first:
```bash
kubectl apply -f jobdefinition-crd.yaml
```

### Status not updating

**Issue:** Status field remains empty

**Check:**
1. CRD has `subresources.status` (line 47 in `jobdefinition-crd.yaml`)
2. Controller uses `UpdateStatus()` not `Update()` (line 147 in `controller.go`)

## 📚 Learning Resources

### Why Dynamic Client?
- **No code generation** required (unlike controller-runtime)
- Works with `map[string]interface{}` via `unstructured.Unstructured`
- Great for prototyping and learning

### Why SharedInformer?
- **Caches resources locally** → reduces API calls
- **Watches efficiently** → only receives changes
- **Handles reconnection** automatically

### Why Status Subresource?
- **Separates concerns** → spec vs status
- **Different RBAC** permissions possible
- **Prevents race conditions** between spec and status updates

## 🎓 Next Steps

1. **Add validation** → Check required fields before enqueueing
2. **Handle failures** → Retry logic, DLQ integration
3. **Add finalizers** → Clean up when JobDefinition is deleted
4. **Metrics** → Expose Prometheus metrics for jobs processed
5. **Leader election** → Run multiple controller replicas

## ⚠️ Constraints Followed

✅ **No kubebuilder** - Used raw client-go  
✅ **No controller-runtime** - Built with SharedInformer  
✅ **No code generation** - Used dynamic client  
✅ **Explicit and minimal** - Every line is commented  
✅ **Learning-focused** - WHY comments everywhere
