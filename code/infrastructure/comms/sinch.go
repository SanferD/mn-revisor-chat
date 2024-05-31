package comms

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const sinchAPIURL = "https://us.sms.api.sinch.com/xms/v1/%s/batches"

type SinchHelper struct {
	apiToken                string
	client                  *http.Client
	projectID               string
	sinchVirtualPhoneNumber string
	timeout                 time.Duration
}

func InitializeSinchHelper(ctx context.Context, apiToken, projectID, sinchVirtualPhoneNumber string, contextTimeout time.Duration) (*SinchHelper, error) {
	if len(apiToken) == 0 || len(projectID) == 0 || len(sinchVirtualPhoneNumber) == 0 {
		return nil, fmt.Errorf("apiToken, projectID, or sinchVirtualPhoneNumber is not specified")
	}
	return &SinchHelper{
		apiToken:                apiToken,
		client:                  &http.Client{},
		projectID:               projectID,
		sinchVirtualPhoneNumber: sinchVirtualPhoneNumber,
		timeout:                 contextTimeout,
	}, nil
}

func (sh *SinchHelper) SendMessage(ctx context.Context, toPhoneNumber, messageContent string) error {
	url := fmt.Sprintf(sinchAPIURL, sh.projectID)

	payload := map[string]interface{}{
		"from": sh.sinchVirtualPhoneNumber,
		"to":   []string{toPhoneNumber},
		"body": messageContent,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, sh.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	sh.setupHeaders(req)

	resp, err := sh.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var responseBody bytes.Buffer
		_, err := responseBody.ReadFrom(resp.Body)
		if err != nil {
			return fmt.Errorf("received non-2xx response status: %s, and failed to read response body: %w", resp.Status, err)
		}
		return fmt.Errorf("received non-2xx response status: %s, response body: %s", resp.Status, responseBody.String())
	}
	return nil
}

func (sh *SinchHelper) setupHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+sh.apiToken)
	req.Header.Set("Content-Type", "application/json")
}
