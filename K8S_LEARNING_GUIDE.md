# 🎓 JobJet K8s Integration - Learning Guide

## What You Built

A **minimal Kubernetes controller** that demonstrates core concepts without framework magic.

## Architecture Overview

```
┌──────────────┐
│ kubectl      │  You create a JobDefinition
└──────┬───────┘
       │
       ▼
┌──────────────────────┐
│ Kubernetes API       │  Stores the custom resource
│ (JobDefinition CRD)  │
└──────┬───────────────┘
       │
       │ Watch API
       ▼
┌──────────────────────┐
│ SharedInformer       │  Caches resources locally
│ (in your controller) │  Notifies on changes
└──────┬───────────────┘
       │
       │ ADD event
       ▼
┌──────────────────────┐
│ Event Handler        │  Your custom logic:
│ handleAdd()          │  1. Enqueue to JobJet
│                      │  2. Update status → Running
│                      │  3. Simulate work (5s)
│                      │  4. Update status → Succeeded
└──────────────────────┘
```

## Key Concepts Explained

### 1. **Dynamic Client** (No Code Generation!)

```go
dynamicClient.Resource(JobDefinitionGVR).Namespace(ns).Get(...)
```

**Why?** Works with `map[string]interface{}` instead of typed structs.
- No need for code generators like `deepcopy-gen`, `client-gen`
- Perfect for prototypes and learning
- Trade-off: No compile-time type safety

### 2. **SharedInformer** (Efficient Watching)

```go
informer := factory.ForResource(JobDefinitionGVR).Informer()
informer.AddEventHandler(cache.ResourceEventHandlerFuncs{...})
```

**Why?** Instead of polling the API every second:
- Maintains a **local cache** of resources
- Uses **long-polling watch** (HTTP/2 streams)
- Only sends **deltas** when things change
- Multiple handlers can share the same cache

**Lifecycle:**
1. Initial **LIST**: Get all existing JobDefinitions
2. **WATCH**: Keep connection open for updates
3. **Cache sync**: Populate local store
4. **Events**: Fire ADD/UPDATE/DELETE handlers

### 3. **Status Subresource** (Critical!)

```yaml
# In CRD:
subresources:
  status: {}
```

```go
// In code:
resourceClient.UpdateStatus(ctx, obj, metav1.UpdateOptions{})
```

**Why separate?**
- Prevents race conditions (user updates spec, controller updates status simultaneously)
- Allows different RBAC permissions (users can't fake "Succeeded" status)
- Status changes don't trigger new controller events

**Wrong way:**
```go
resourceClient.Update(...)  // ❌ Overwrites both spec and status
```

**Right way:**
```go
resourceClient.UpdateStatus(...)  // ✅ Only touches .status field
```

### 4. **Unstructured Data** (Working Without Types)

```go
u := &unstructured.Unstructured{}
queue, _, _ := unstructured.NestedString(u.Object, "spec", "queue")
unstructured.SetNestedField(u.Object, "Running", "status", "state")
```

**Why?** Because we don't have Go structs for JobDefinition.

**Think of it as:**
```go
// Conceptually this:
type JobDefinition struct {
    Spec struct {
        Queue string
    }
    Status struct {
        State string
    }
}

// Becomes this:
map[string]interface{}{
    "spec": map[string]interface{}{
        "queue": "email-queue",
    },
    "status": map[string]interface{}{
        "state": "Running",
    },
}
```

## Code Flow Walkthrough

### Step 1: Controller Startup

```go
// cmd/controller/main.go
config := clientcmd.BuildConfigFromFlags("", kubeconfig)
dynamicClient := dynamic.NewForConfig(config)
controller := k8s.NewController(dynamicClient, enqueuer, "")
controller.Run(stopCh)
```

**What happens:**
1. Load `~/.kube/config` (contains cluster URL + auth)
2. Create dynamic client (can talk to any API resource)
3. Create controller instance
4. Start watching

### Step 2: Informer Initialization

```go
// pkg/k8s/controller.go
factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(...)
informer := factory.ForResource(JobDefinitionGVR).Informer()
informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
    AddFunc: c.handleAdd,
})
factory.Start(stopCh)
cache.WaitForCacheSync(stopCh, informer.HasSynced)
```

**What happens:**
1. Factory creates informers on-demand
2. We request an informer for `jobdefinitions.jobjet.dev/v1`
3. Register callbacks for ADD/UPDATE/DELETE
4. Start watching (API server connection)
5. Wait for initial LIST to complete

### Step 3: Resource Created

```bash
kubectl apply -f job.yaml
```

**Kubernetes side:**
1. API server validates against CRD schema
2. Stores in etcd
3. Sends watch event to connected clients

**Controller side:**
1. Informer receives ADD event
2. Updates local cache
3. Calls `handleAdd(obj)`

### Step 4: Event Handler

```go
func (c *Controller) handleAdd(obj interface{}) {
    u := obj.(*unstructured.Unstructured)
    name := u.GetName()
    
    // 1. Enqueue to JobJet
    jobID, _ := c.enqueuer.Enqueue(name)
    
    // 2. Update status
    c.updateStatus(namespace, name, "Running")
    
    // 3. Simulate work
    go func() {
        time.Sleep(5 * time.Second)
        c.updateStatus(namespace, name, "Succeeded")
    }()
}
```

**What happens:**
1. Extract resource metadata
2. Call JobJet's enqueue function
3. Write "Running" to `.status.state`
4. Spawn goroutine to wait and update to "Succeeded"

### Step 5: Status Update

```go
func (c *Controller) updateStatus(namespace, name, state string) {
    // Get latest version
    obj, _ := resourceClient.Get(ctx, name, metav1.GetOptions{})
    
    // Modify status field
    unstructured.SetNestedField(obj.Object, state, "status", "state")
    
    // Write back (MUST use UpdateStatus!)
    resourceClient.UpdateStatus(ctx, obj, metav1.UpdateOptions{})
}
```

**Why GET first?**
- Need latest `resourceVersion` to avoid conflicts
- If someone else modified it, API server rejects our update
- We'd need to retry with fresh version

## Common Pitfalls (Avoided!)

### ❌ Using Update Instead of UpdateStatus

```go
// This overwrites spec too!
resourceClient.Update(ctx, obj, metav1.UpdateOptions{})
```

**Result:** Status changes, but may also reset spec fields.

### ❌ Forgetting Status Subresource in CRD

```yaml
# Missing this in CRD:
subresources:
  status: {}
```

**Result:** UpdateStatus() fails with 404 error.

### ❌ Not Waiting for Cache Sync

```go
factory.Start(stopCh)
// Missing: cache.WaitForCacheSync(...)
// Code continues immediately
```

**Result:** Event handlers fire before cache is populated, may miss existing resources.

### ❌ Blocking Event Handlers

```go
AddFunc: func(obj interface{}) {
    time.Sleep(5 * time.Minute)  // ❌ Blocks all events!
}
```

**Result:** No other events processed until this returns.

**Fix:** Use goroutines for async work.

## Testing Checklist

- [ ] CRD installs without errors
- [ ] Controller connects to cluster
- [ ] Controller logs "Cache synced"
- [ ] Creating JobDefinition triggers ADD event
- [ ] Status changes: (empty) → Running → Succeeded
- [ ] `kubectl get jd` shows correct state
- [ ] Controller handles multiple jobs
- [ ] Ctrl+C shuts down gracefully

## What You Learned

✅ How to watch custom resources without kubebuilder  
✅ Dynamic client vs typed client trade-offs  
✅ SharedInformer caching and efficiency  
✅ Status subresource pattern  
✅ Unstructured data manipulation  
✅ Resource versioning and conflicts  
✅ Graceful shutdown handling

## Next Steps for Production

1. **Work Queue** - Use `client-go/util/workqueue` for retry logic
2. **Leader Election** - Run multiple controller replicas safely
3. **Finalizers** - Clean up external resources on deletion
4. **RBAC** - Define ServiceAccount + Roles for in-cluster deployment
5. **Metrics** - Expose Prometheus metrics for monitoring
6. **Error Handling** - Exponential backoff, max retries, DLQ
7. **Logging** - Structured logging (logr, zap)
8. **Testing** - Unit tests with fake clients, integration tests with envtest

## Resources

- [client-go examples](https://github.com/kubernetes/client-go/tree/master/examples)
- [sample-controller](https://github.com/kubernetes/sample-controller) - Official minimal example
- [Programming Kubernetes](https://www.oreilly.com/library/view/programming-kubernetes/9781492047094/) - Book

## Congratulations! 🎉

You've built a Kubernetes controller from scratch without magic frameworks. This knowledge transfers directly to understanding how kubebuilder and Operator SDK work under the hood.
