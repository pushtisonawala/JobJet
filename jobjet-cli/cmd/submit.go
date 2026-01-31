package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"jobjet-cli/pkg/client"
	"jobjet-cli/pkg/spinner"
)

var (
	payloadFlag  string
	priorityFlag int
	retriesFlag  int
	timeoutFlag  int
)

var submitCmd = &cobra.Command{
	Use:   "submit HANDLER",
	Short: "Submit a new job",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		handler := args[0]
		if payloadFlag == "" {
			fmt.Fprintln(os.Stderr, "--payload is required")
			os.Exit(1)
		}

		var payloadData []byte
		var err error
		if strings.HasPrefix(payloadFlag, "@") {
			filePath := strings.TrimPrefix(payloadFlag, "@")
			payloadData, err = ioutil.ReadFile(filePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to read payload file: %v\n", err)
				os.Exit(1)
			}
		} else {
			payloadData = []byte(payloadFlag)
		}

		var js map[string]interface{}
		if err := json.Unmarshal(payloadData, &js); err != nil {
			fmt.Fprintln(os.Stderr, "Invalid JSON payload:", err)
			os.Exit(1)
		}

		spin := spinner.New("Submitting job to Kubernetes...")
		spin.Start()
		ns := os.Getenv("JOBJET_NAMESPACE")
		if ns == "" {
			ns = "default"
		}
		jobID, err := client.SubmitJobK8s(handler, js, ns)
		spin.Stop()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to submit job to K8s: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Job submitted to Kubernetes!\nJobDefinition: %s\nQueue: %s\n", jobID, handler)
		fmt.Printf("To stream logs: jobjet logs %s --follow\n", jobID)
	},
}

func init() {
	rootCmd.AddCommand(submitCmd)
	submitCmd.Flags().StringVar(&payloadFlag, "payload", "", "Job payload as JSON string or @file.json (required)")
	submitCmd.Flags().IntVar(&priorityFlag, "priority", 5, "Job priority (default 5)")
	submitCmd.Flags().IntVar(&retriesFlag, "retries", 3, "Max retries (default 3)")
	submitCmd.Flags().IntVar(&timeoutFlag, "timeout", 300, "Timeout in seconds (default 300)")
	submitCmd.MarkFlagRequired("payload")
}
