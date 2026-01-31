package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type JobResources struct {
	Job  *batchv1.Job
	Pods []corev1.Pod
}

// resolveJobResources finds the K8s Job and all Pods for a given jobdefinition label
func resolveJobResources(ctx context.Context, clientset *kubernetes.Clientset, jobDefinition string) (*JobResources, error) {
	namespace := os.Getenv("JOBJET_NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}
	labelSelector := fmt.Sprintf("jobdefinition=%s", jobDefinition)

	// Find Job
	jobList, err := clientset.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %w", err)
	}
	var job *batchv1.Job
	if len(jobList.Items) > 0 {
		job = &jobList.Items[0]
	}

	// Find Pods
	podList, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}
	pods := podList.Items
	// Sort pods by creationTimestamp (newest first)
	sort.Slice(pods, func(i, j int) bool {
		return pods[i].CreationTimestamp.Time.After(pods[j].CreationTimestamp.Time)
	})

	if job == nil && len(pods) == 0 {
		return nil, fmt.Errorf("no Kubernetes Job or Pods found for jobdefinition '%s'", jobDefinition)
	}

	return &JobResources{Job: job, Pods: pods}, nil
}
