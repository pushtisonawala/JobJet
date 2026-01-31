package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
)

var (
	followFlag     bool
	containerFlag  string
	sinceFlag      string
	tailFlag       int
	timestampsFlag bool
)

var logsCmd = &cobra.Command{
	Use:   "logs <job-id>",
	Short: "Fetch and stream logs for a JobJet job",
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
		if len(resources.Pods) == 0 {
			return fmt.Errorf("no pods found for jobdefinition '%s'", jobDef)
		}
		pod := resources.Pods[0] // newest pod
		fmt.Fprintf(os.Stderr, "Using pod: %s\n", pod.Name)

		// 4. Prepare log options
		opts := &corev1.PodLogOptions{
			Follow:     followFlag,
			Timestamps: timestampsFlag,
		}
		if containerFlag != "" {
			opts.Container = containerFlag
		}
		if sinceFlag != "" {
			dur, err := time.ParseDuration(sinceFlag)
			if err != nil {
				return fmt.Errorf("invalid --since duration: %v", err)
			}
			opts.SinceSeconds = int64Ptr(int64(dur.Seconds()))
		}
		if tailFlag > 0 {
			opts.TailLines = int64Ptr(int64(tailFlag))
		}

		req := clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, opts)
		stream, err := req.Stream(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to stream logs: %v", err)
		}
		defer stream.Close()

		ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer cancel()

		buf := make([]byte, 4096)
		for {
			select {
			case <-ctx.Done():
				return nil
			default:
				n, err := stream.Read(buf)
				if n > 0 {
					os.Stdout.Write(buf[:n])
				}
				if err != nil {
					return nil
				}
			}
		}
	},
}

func int64Ptr(i int64) *int64 { return &i }

func init() {
	logsCmd.Flags().BoolVarP(&followFlag, "follow", "f", false, "Stream logs continuously")
	logsCmd.Flags().StringVar(&containerFlag, "container", "", "Container name")
	logsCmd.Flags().StringVar(&sinceFlag, "since", "", "Fetch logs since duration (e.g. 10m)")
	logsCmd.Flags().IntVar(&tailFlag, "tail", 0, "Number of last lines")
	logsCmd.Flags().BoolVar(&timestampsFlag, "timestamps", false, "Show timestamps")
	rootCmd.AddCommand(logsCmd)
}
