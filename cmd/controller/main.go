package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"jobjet/pkg/k8s"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// MockJobJetEnqueuer is a simple implementation for demonstration
// WHY: In production, this would call your actual JobJet queue system
type MockJobJetEnqueuer struct{}

func (m *MockJobJetEnqueuer) Enqueue(jobName string) (string, error) {
	// In production, replace this with actual JobJet.Enqueue call
	// For now, we just simulate success
	jobID := fmt.Sprintf("job-%s-%d", jobName, os.Getpid())
	log.Printf("[JobJet] Enqueue called for: %s → JobID: %s\n", jobName, jobID)
	return jobID, nil
}

func main() {
	log.Println("JobJet Kubernetes Controller")
	log.Println("=============================")

	// Step 1: Load kubeconfig
	// WHY: We need to authenticate with the Kubernetes API server
	kubeconfig := getKubeconfig()
	log.Printf("Using kubeconfig: %s\n", kubeconfig)

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("Failed to build kubeconfig: %v", err)
	}

	// Step 2: Create dynamic client
	// WHY: Dynamic client works with unstructured data (no code generation needed)
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create dynamic client: %v", err)
	}
	log.Println("✓ Connected to Kubernetes cluster")

	// Step 3: Create JobJet enqueuer
	// TODO: Replace MockJobJetEnqueuer with your actual JobJet implementation
	enqueuer := &MockJobJetEnqueuer{}

	// Step 4: Create and start the controller
	// Watching all namespaces (use "default" to watch only default namespace)
	controller := k8s.NewController(dynamicClient, enqueuer, "")

	// Step 5: Handle graceful shutdown
	// WHY: Allows the controller to clean up when receiving SIGINT/SIGTERM
	stopCh := make(chan struct{})
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("Received shutdown signal")
		close(stopCh)
	}()

	// Step 6: Run the controller (blocks until stopCh is closed)
	if err := controller.Run(stopCh); err != nil {
		log.Fatalf("Controller error: %v", err)
	}

	log.Println("Controller stopped gracefully")
}

// getKubeconfig returns the path to the kubeconfig file
// WHY: Tries in-cluster config first, then falls back to ~/.kube/config
func getKubeconfig() string {
	// Check for KUBECONFIG environment variable
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		return kubeconfig
	}

	// Default to ~/.kube/config
	if home := homedir.HomeDir(); home != "" {
		return filepath.Join(home, ".kube", "config")
	}

	log.Fatal("Unable to locate kubeconfig")
	return ""
}
