package client

import (
	"context"
	"fmt"
	"os"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// SubmitJobK8s creates a JobDefinition CRD in Kubernetes
func SubmitJobK8s(handler string, payload map[string]interface{}, namespace string) (string, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			home, _ := os.UserHomeDir()
			kubeconfig = home + "/.kube/config"
		}
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return "", fmt.Errorf("failed to load kubeconfig: %v", err)
		}
	}
	dyn, err := dynamic.NewForConfig(config)
	if err != nil {
		return "", fmt.Errorf("failed to create dynamic client: %v", err)
	}
	gvr := schema.GroupVersionResource{
		Group:    "jobjet.dev",
		Version:  "v1",
		Resource: "jobdefinitions",
	}
	name := fmt.Sprintf("%s-%d", handler, time.Now().UnixNano())
	cr := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "jobjet.dev/v1",
			"kind":       "JobDefinition",
			"metadata": map[string]interface{}{
				"name": name,
			},
			"spec": map[string]interface{}{
				"queue":   handler,
				"payload": payload,
			},
		},
	}
	res, err := dyn.Resource(gvr).Namespace(namespace).Create(context.TODO(), cr, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create JobDefinition: %v", err)
	}
	return res.GetName(), nil
}
