package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Job struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Status     string                 `json:"status"`
	RetryCount int                    `json:"retry_count"`
	MaxRetry   int                    `json:"max_retry"`
	CreatedAt  string                 `json:"createdat"`
	Payload    map[string]interface{} `json:"payload"`
}

type QueueResponse struct {
	Count    int      `json:"count"`
	JobQueue []string `json:"job_queue"`
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List jobs in the queue",
	Long:  `List all jobs currently in the JobJet queue with their status and details.`,
	Run: func(cmd *cobra.Command, args []string) {
		listJobs()
	},
}

func listJobs() {
	apiURL := viper.GetString("api-url")
	if !strings.HasPrefix(apiURL, "http") {
		apiURL = "http://" + apiURL
	}

	// Try to get queue status from the debug endpoint
	resp, err := http.Get(fmt.Sprintf("%s/debug/jobqueue", apiURL))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to JobJet API: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "API returned status %d\n", resp.StatusCode)
		os.Exit(1)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response: %v\n", err)
		os.Exit(1)
	}

	var queueResp QueueResponse
	if err := json.Unmarshal(body, &queueResp); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
		os.Exit(1)
	}

	outputFormat := viper.GetString("output")

	switch outputFormat {
	case "json":
		fmt.Println(string(body))
	case "yaml":
		printYAML(&queueResp)
	default:
		printTable(&queueResp)
	}
}

func printTable(queueResp *QueueResponse) {
	if queueResp.Count == 0 {
		fmt.Println("No jobs in queue")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "JOB ID\tTYPE\tSTATUS\tRETRIES\tCREATED")
	fmt.Fprintln(w, "------\t----\t------\t-------\t-------")

	for _, jobStr := range queueResp.JobQueue {
		var job Job
		if err := json.Unmarshal([]byte(jobStr), &job); err != nil {
			fmt.Fprintf(w, "ERROR\t%s\t-\t-\t-\n", err.Error())
			continue
		}

		shortID := job.ID
		if len(shortID) > 8 {
			shortID = shortID[:8] + "..."
		}

		createdAt := job.CreatedAt
		if len(createdAt) > 19 {
			createdAt = createdAt[:19]
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%d/%d\t%s\n",
			shortID,
			job.Type,
			job.Status,
			job.RetryCount,
			job.MaxRetry,
			createdAt,
		)
	}

	w.Flush()
	fmt.Printf("\nTotal jobs: %d\n", queueResp.Count)
}

func printYAML(queueResp *QueueResponse) {
	fmt.Printf("count: %d\n", queueResp.Count)
	fmt.Println("jobs:")

	for i, jobStr := range queueResp.JobQueue {
		var job Job
		if err := json.Unmarshal([]byte(jobStr), &job); err != nil {
			fmt.Printf("  - error: %s\n", err.Error())
			continue
		}

		fmt.Printf("  - id: %s\n", job.ID)
		fmt.Printf("    type: %s\n", job.Type)
		fmt.Printf("    status: %s\n", job.Status)
		fmt.Printf("    retries: %d/%d\n", job.RetryCount, job.MaxRetry)
		fmt.Printf("    created: %s\n", job.CreatedAt)
		if i < len(queueResp.JobQueue)-1 {
			fmt.Println()
		}
	}
}

func init() {
	rootCmd.AddCommand(listCmd)
}
