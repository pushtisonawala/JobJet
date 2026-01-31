package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var describeCmd = &cobra.Command{
	Use:   "describe <job-id>",
	Short: "Describe a JobJet job and related Kubernetes resources",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		jobDef := args[0]
		if strings.TrimSpace(jobDef) == "" {
			return fmt.Errorf("job-id is required")
		}

		clientset, err := getKubeClient()
		if err != nil {
			return fmt.Errorf("failed to create k8s client: %v", err)
		}

		resources, err := resolveJobResources(cmd.Context(), clientset, jobDef)
		if err != nil {
			return err
		}

		fmt.Printf("Job: %s\n", jobDef)
		queue := ""
		if resources.Job != nil {
			if q, ok := resources.Job.Labels["queue"]; ok {
				queue = q
			}
		}
		if queue == "" && len(resources.Pods) > 0 {
			if q, ok := resources.Pods[0].Labels["queue"]; ok {
				queue = q
			}
		}
		fmt.Printf("Queue: %s\n", queue)
		if resources.Job != nil {
			fmt.Printf("K8s Job: %s\n", resources.Job.Name)
		} else {
			fmt.Printf("K8s Job: (not found)\n")
		}
		fmt.Printf("Pods:\n")
		for _, pod := range resources.Pods {
			fmt.Printf("  - %s (%s)\n", pod.Name, pod.Status.Phase)
		}
		retries := 0
		// removed unused failures variable
		var startTime, completionTime string
		if resources.Job != nil {
			if resources.Job.Status.StartTime != nil {
				startTime = resources.Job.Status.StartTime.Time.Format("2006-01-02 15:04:05")
			}
			if resources.Job.Status.CompletionTime != nil {
				completionTime = resources.Job.Status.CompletionTime.Time.Format("2006-01-02 15:04:05")
			}
			retries = int(resources.Job.Status.Failed)
		}
		fmt.Printf("Retries: %d\n", retries)
		fmt.Printf("Start Time: %s\n", startTime)
		fmt.Printf("Completion Time: %s\n", completionTime)

		fmt.Println("\nEvents:")
		for _, pod := range resources.Pods {
			events, err := clientset.CoreV1().Events(pod.Namespace).List(cmd.Context(), metav1.ListOptions{
				FieldSelector: fmt.Sprintf("involvedObject.kind=Pod,involvedObject.name=%s", pod.Name),
			})
			if err != nil {
				fmt.Printf("  Failed to fetch events for pod %s: %v\n", pod.Name, err)
				continue
			}
			if len(events.Items) == 0 {
				fmt.Printf("  No events for pod %s\n", pod.Name)
				continue
			}
			for _, ev := range events.Items {
				if isInterestingEvent(ev) {
					fmt.Printf("  [%s] %s: %s\n", ev.LastTimestamp.Time.Format(time.RFC3339), ev.Reason, ev.Message)
				}
			}
		}
		return nil
	},
}

func isInterestingEvent(ev corev1.Event) bool {
	badReasons := []string{"FailedScheduling", "FailedMount", "Failed", "BackOff", "CrashLoopBackOff", "ErrImagePull", "ImagePullBackOff", "OOMKilled"}
	for _, r := range badReasons {
		if strings.Contains(ev.Reason, r) || strings.Contains(ev.Message, r) {
			return true
		}
	}
	return false
}

func init() {
	rootCmd.AddCommand(describeCmd)
}
