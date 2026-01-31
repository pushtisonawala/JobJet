package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type Client struct {
	apiURL     string
	httpClient *http.Client
}

func NewClient(apiURL string) *Client {
	return &Client{
		apiURL:     apiURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

type submitJobRequest struct {
	Type     string          `json:"type"`
	Payload  json.RawMessage `json:"payload"`
	Priority int             `json:"priority"`
	Retries  int             `json:"retries"`
	Timeout  int             `json:"timeout"`
}

type submitJobResponse struct {
	Message string `json:"message"`
	JobID   string `json:"job_id"`
}

func (c *Client) SubmitJob(handler string, payload []byte, priority, retries, timeout int) (string, error) {
	reqBody := submitJobRequest{
		Type:     handler,
		Payload:  payload,
		Priority: priority,
		Retries:  retries,
		Timeout:  timeout,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal job request: %w", err)
	}
	url := fmt.Sprintf("%s/jobs", c.apiURL)
	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}
	var out submitJobResponse
	_ = json.Unmarshal(respBody, &out)
	if out.JobID != "" {
		return out.JobID, nil
	}
	// If no jobID, show the message or raw response as error
	if out.Message != "" {
		return "", fmt.Errorf("API error: %s", out.Message)
	}
	return "", fmt.Errorf("API error: %s", string(respBody))
}
