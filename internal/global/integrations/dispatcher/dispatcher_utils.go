package dispatcher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/injunweb/backend-server/env"
)

type ClientPayload map[string]interface{}

func TriggerDispatch(repo, eventType string, clientPayload ClientPayload) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/dispatches", env.GITHUB_OWNER, repo)

	payload := map[string]interface{}{
		"event_type":     eventType,
		"client_payload": clientPayload,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+env.GITHUB_DISPATCH_TOKEN)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
