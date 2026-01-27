package k8s

import (
	"context"
	"fmt"
	"log"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
)

// JobDefinition GVR (Group Version Resource)
// This identifies our custom resource in the Kubernetes API
var JobDefinitionGVR = schema.GroupVersionResource{
	Group:    "jobjet.dev",
	Version:  "v1",
	Resource: "jobdefinitions",
}

// JobJetEnqueuer defines the interface to enqueue jobs
// In production, this would be implemented by your JobJet queue system
type JobJetEnqueuer interface {
	Enqueue(jobName string) (jobID string, err error)
}

// Controller watches JobDefinition resources and enqueues them to JobJet
type Controller struct {
	dynamicClient dynamic.Interface
	enqueuer      JobJetEnqueuer
	namespace     string // namespace to watch (use "" for all namespaces)
}

// NewController creates a new Kubernetes controller
func NewController(dynamicClient dynamic.Interface, enqueuer JobJetEnqueuer, namespace string) *Controller {
	return &Controller{
		dynamicClient: dynamicClient,
		enqueuer:      enqueuer,
		namespace:     namespace,
	}
}

// Run starts the controller and blocks until stopCh is closed
func (c *Controller) Run(stopCh <-chan struct{}) error {
	log.Println("Starting JobDefinition controller...")

	// Create a dynamic informer factory
	// This factory creates informers for dynamic (unstructured) resources
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(
		c.dynamicClient,
		30*time.Second, // resync period - full list/watch every 30s
		c.namespace,    // namespace filter
		nil,            // no label/field selectors
	)

	// Get an informer for our JobDefinition resource
	// SharedInformer maintains a local cache and watches for changes
	informer := factory.ForResource(JobDefinitionGVR).Informer()

	// Register event handlers
	// WHY: These callbacks are invoked when the informer detects changes
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// Called when a new JobDefinition is created
			c.handleAdd(obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			// Called when a JobDefinition is modified
			// We mostly ignore updates unless needed for reconciliation
			log.Println("Update event received (ignored for now)")
		},
		DeleteFunc: func(obj interface{}) {
			// Called when a JobDefinition is deleted
			log.Println("Delete event received (ignored for now)")
		},
	})

	// Start the informer
	// WHY: This begins watching the API server and populating the cache
	factory.Start(stopCh)

	// Wait for cache sync
	// WHY: Before processing events, we must ensure the cache is fully populated
	log.Println("Waiting for cache sync...")
	if !cache.WaitForCacheSync(stopCh, informer.HasSynced) {
		return fmt.Errorf("timed out waiting for cache sync")
	}

	log.Println("Cache synced. Controller is running.")

	// Block until stopCh is closed
	<-stopCh
	log.Println("Shutting down controller...")
	return nil
}

// handleAdd processes new JobDefinition resources
func (c *Controller) handleAdd(obj interface{}) {
	// Convert the object to unstructured format
	// WHY: Dynamic client works with unstructured.Unstructured, not typed structs
	u, ok := obj.(*unstructured.Unstructured)
	if !ok {
		log.Printf("ERROR: unexpected object type: %T\n", obj)
		return
	}

	name := u.GetName()
	namespace := u.GetNamespace()
	log.Printf("▶ JobDefinition created: %s/%s\n", namespace, name)

	// Extract spec fields (optional - for logging/validation)
	spec, found, err := unstructured.NestedMap(u.Object, "spec")
	if !found || err != nil {
		log.Printf("ERROR: spec not found in %s/%s\n", namespace, name)
		return
	}
	queue, _, _ := unstructured.NestedString(spec, "queue")
	log.Printf("  Queue: %s\n", queue)

	// Step 1: Enqueue job to JobJet
	log.Printf("  Enqueueing job to JobJet...\n")
	jobID, err := c.enqueuer.Enqueue(name)
	if err != nil {
		log.Printf("ERROR: failed to enqueue job: %v\n", err)
		c.updateStatus(namespace, name, "Failed")
		return
	}
	log.Printf("  ✓ Job enqueued with ID: %s\n", jobID)

	// Step 2: Update status to "Running"
	c.updateStatus(namespace, name, "Running")

	// Step 3: Simulate async work completion
	// WHY: In real scenarios, JobJet would notify us when job completes
	// For learning purposes, we simulate a delay then mark as Succeeded
	go func() {
		time.Sleep(5 * time.Second)
		log.Printf("  Job %s completed (simulated)\n", name)
		c.updateStatus(namespace, name, "Succeeded")
	}()
}

// updateStatus updates the .status.state field of a JobDefinition
// CRITICAL: Must use UpdateStatus, not Update, because status is a subresource
func (c *Controller) updateStatus(namespace, name, state string) {
	log.Printf("  Updating status.state to '%s'...\n", state)

	// Get the resource client for our namespace
	resourceClient := c.dynamicClient.Resource(JobDefinitionGVR).Namespace(namespace)

	// Fetch the current resource
	// WHY: We need the latest resourceVersion to avoid conflicts
	ctx := context.Background()
	obj, err := resourceClient.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		log.Printf("ERROR: failed to get JobDefinition: %v\n", err)
		return
	}

	// Update the status field
	// WHY: unstructured.SetNestedField works with map[string]interface{} objects
	if err := unstructured.SetNestedField(obj.Object, state, "status", "state"); err != nil {
		log.Printf("ERROR: failed to set status.state: %v\n", err)
		return
	}

	// Write the status back using UpdateStatus (not Update!)
	// WHY: Kubernetes separates spec and status updates for CRDs with status subresource
	_, err = resourceClient.UpdateStatus(ctx, obj, metav1.UpdateOptions{})
	if err != nil {
		log.Printf("ERROR: failed to update status: %v\n", err)
		return
	}

	log.Printf("  ✓ Status updated to '%s'\n", state)
}
