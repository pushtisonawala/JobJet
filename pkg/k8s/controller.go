package k8s

import (
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

// JobEnqueuer interface for enqueuing jobs
type JobEnqueuer interface {
	Enqueue(handler string, payload interface{}) error
}

// Controller watches JobDefinition CRDs and enqueues jobs
type Controller struct {
	client    dynamic.Interface
	enqueuer  JobEnqueuer
	namespace string
	stopCh    chan struct{}
}

// JobDefinition represents the structure of our custom resource
type JobDefinition struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              JobDefinitionSpec `json:"spec,omitempty"`
}

type JobDefinitionSpec struct {
	Handler string                 `json:"handler"`
	Payload map[string]interface{} `json:"payload"`
}

// NewController creates a new Kubernetes controller
func NewController(client dynamic.Interface, enqueuer JobEnqueuer, namespace string) *Controller {
	return &Controller{
		client:    client,
		enqueuer:  enqueuer,
		namespace: namespace,
		stopCh:    make(chan struct{}),
	}
}

// Run starts the controller and blocks until stopped
func (c *Controller) Run(stopCh <-chan struct{}) error {
	log.Println("Starting Kubernetes controller...")

	// Define the JobDefinition CRD GroupVersionResource
	gvr := schema.GroupVersionResource{
		Group:    "jobjet.io",
		Version:  "v1",
		Resource: "jobdefinitions",
	}

	// Create a dynamic informer for JobDefinition CRDs
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(
		c.client,
		time.Second*30,
		c.namespace,
		nil,
	)

	informer := factory.ForResource(gvr).Informer()

	// Add event handlers
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			c.handleJobDefinition(obj, "ADDED")
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			c.handleJobDefinition(newObj, "UPDATED")
		},
		DeleteFunc: func(obj interface{}) {
			c.handleJobDefinition(obj, "DELETED")
		},
	})

	// Start the informer
	factory.Start(stopCh)

	// Wait for caches to sync
	log.Println("Waiting for cache sync...")
	if !cache.WaitForCacheSync(stopCh, informer.HasSynced) {
		return fmt.Errorf("failed to sync cache")
	}

	log.Println("Controller is now watching for JobDefinition CRDs...")

	// Block until stop signal is received
	<-stopCh
	log.Println("Stopping controller...")
	return nil
}

// handleJobDefinition processes JobDefinition CRD events
func (c *Controller) handleJobDefinition(obj interface{}, eventType string) {
	// Convert unstructured object to our JobDefinition
	unstructuredObj, ok := obj.(*unstructured.Unstructured)
	if !ok {
		log.Printf("Failed to cast object to unstructured: %v", obj)
		return
	}

	// Extract the spec from the unstructured object
	spec, found, err := unstructured.NestedMap(unstructuredObj.Object, "spec")
	if err != nil || !found {
		log.Printf("Failed to get spec from JobDefinition: %v", err)
		return
	}

	// Extract handler and payload
	handler, _, _ := unstructured.NestedString(spec, "handler")
	payload, _, _ := unstructured.NestedMap(spec, "payload")

	name := unstructuredObj.GetName()
	namespace := unstructuredObj.GetNamespace()

	log.Printf("%s JobDefinition: %s/%s (handler: %s)", eventType, namespace, name, handler)

	// Only enqueue jobs for ADDED events
	if eventType == "ADDED" {
		if err := c.enqueuer.Enqueue(handler, payload); err != nil {
			log.Printf("Failed to enqueue job: %v", err)
			return
		}
		log.Printf("Successfully enqueued job: %s with handler: %s", name, handler)
	}
}
