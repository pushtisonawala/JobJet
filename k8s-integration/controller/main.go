package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	gvr = schema.GroupVersionResource{
		Group:    "batch.jobjet.dev",
		Version:  "v1alpha1",
		Resource: "jobjettasks",
	}

	// JobJet API URL (same server that runs on :8000)
	jobjetURL = getEnv("JOBJET_URL", "http://localhost:8000")
)

func main() {
	log.Println("🚀 JobJet K8s Controller starting...")
	log.Printf("📡 Connecting to JobJet API: %s", jobjetURL)

	// Load kubeconfig from default location
	kubeconfig := os.Getenv("HOME") + "/.kube/config"
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("Failed to load kubeconfig: %v", err)
	}

	// Create Kubernetes dynamic client
	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create K8s client: %v", err)
	}

	// Create informer to watch JobJetTask resources
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(
		dynClient,
		30*time.Second,
		metav1.NamespaceAll,
		nil,
	)

	informer := factory.ForResource(gvr).Informer()

	// Register event handlers
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			u := obj.(*unstructured.Unstructured)
			log.Printf("✨ New JobJetTask detected: %s/%s", u.GetNamespace(), u.GetName())
			handleTask(dynClient, u)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			u := newObj.(*unstructured.Unstructured)

			// Only process if status is empty (not yet submitted)
			phase, _, _ := unstructured.NestedString(u.Object, "status", "phase")
			if phase == "" {
				log.Printf("🔄 Processing updated JobJetTask: %s/%s", u.GetNamespace(), u.GetName())
				handleTask(dynClient, u)
			}
		},
	})

	// Start the informer
	stopCh := make(chan struct{})
	defer close(stopCh)

	log.Println("✅ Controller running. Watching for JobJetTask resources...")
	log.Println("   Press Ctrl+C to exit")

	factory.Start(stopCh)
	factory.WaitForCacheSync(stopCh)

	// Block forever
	<-stopCh
}

func handleTask(client dynamic.Interface, obj *unstructured.Unstructured) {
	ctx := context.Background()

	// Skip if already processed
	phase, _, _ := unstructured.NestedString(obj.Object, "status", "phase")
	if phase != "" {
		return
	}

	// Extract handler (job type)
	handler, found, err := unstructured.NestedString(obj.Object, "spec", "handler")
	if !found || err != nil {
		log.Printf("❌ Missing handler field")
		updateTaskStatus(ctx, client, obj, "Failed", "", "Missing required field: handler")
		return
	}

	// Extract payload
	payloadMap, found, err := unstructured.NestedMap(obj.Object, "spec", "payload")
	if !found || err != nil {
		log.Printf("❌ Missing or invalid payload field")
		updateTaskStatus(ctx, client, obj, "Failed", "", "Missing or invalid payload")
		return
	}

	log.Printf("📤 Submitting job: type=%s", handler)

	// Submit to JobJet API
	jobID, err := submitToJobJetAPI(handler, payloadMap)
	if err != nil {
		log.Printf("❌ Failed to submit: %v", err)
		updateTaskStatus(ctx, client, obj, "Failed", "", err.Error())
		return
	}

	log.Printf("✅ Job submitted successfully: jobID=%s", jobID)
	updateTaskStatus(ctx, client, obj, "Pending", jobID, "Job submitted to JobJet")
}

// Calls your existing JobJet API at http://localhost:8000/jobs
func submitToJobJetAPI(jobType string, payload map[string]interface{}) (string, error) {
	// Build request matching YOUR API format
	reqBody := map[string]interface{}{
		"type":    jobType, // Maps to req.Type in your controller
		"payload": payload, // Maps to req.Payload (EmailPayload)
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	log.Printf("→ POST %s/jobs", jobjetURL)
	log.Printf("  Body: %s", string(jsonData))

	// Make HTTP POST request
	resp, err := http.Post(
		jobjetURL+"/jobs",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Your API returns 202 Accepted
	if resp.StatusCode != http.StatusAccepted {
		bodyBytes := make([]byte, 1024)
		n, _ := resp.Body.Read(bodyBytes)
		return "", fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes[:n]))
	}

	// Parse your API response
	var apiResp struct {
		Message     string `json:"message"`
		JobID       string `json:"job_id"`
		RedisAddr   string `json:"redis_addr"`
		JobQueueLen int64  `json:"job_queue_len"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	log.Printf("← API Response: message=%s, jobID=%s, queueLen=%d, redis=%s",
		apiResp.Message, apiResp.JobID, apiResp.JobQueueLen, apiResp.RedisAddr)

	return apiResp.JobID, nil
}

func updateTaskStatus(ctx context.Context, client dynamic.Interface,
	obj *unstructured.Unstructured, phase, jobID, message string) {

	status := map[string]interface{}{
		"phase":       phase,
		"message":     message,
		"submittedAt": time.Now().Format(time.RFC3339),
	}

	if jobID != "" {
		status["jobID"] = jobID
	}

	unstructured.SetNestedMap(obj.Object, status, "status")

	// Update the status subresource in K8s
	_, err := client.Resource(gvr).
		Namespace(obj.GetNamespace()).
		UpdateStatus(ctx, obj, metav1.UpdateOptions{})

	if err != nil {
		log.Printf("⚠️  Failed to update K8s status: %v", err)
	} else {
		log.Printf("✓ Updated K8s status: phase=%s", phase)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}