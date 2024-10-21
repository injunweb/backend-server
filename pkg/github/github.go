package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/injunweb/backend-server/internal/config"
	"github.com/injunweb/backend-server/internal/models"
)

func TriggerWriteValuesWorkflow(app models.Application) error {
	payload := map[string]interface{}{
		"event_type": "write-values",
		"client_payload": map[string]string{
			"appName":         app.Name,
			"git":             app.GitURL,
			"branch":          app.Branch,
			"port":            fmt.Sprintf("%d", app.Port),
			"primaryHostname": app.PrimaryHostname,
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON payload: %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.github.com/repos/injunweb/gitops-repo/dispatches",
		bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create GitHub request: %v", err)
	}

	req.Header.Set("Authorization", "token "+config.AppConfig.GithubToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to dispatch GitHub event: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub dispatch failed with status: %s", resp.Status)
	}

	return nil
}

func TriggerRemovePipelineWorkflow(app models.Application) error {
	payload := map[string]interface{}{
		"event_type": "remove-pipeline",
		"client_payload": map[string]string{
			"appName": app.Name,
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON payload: %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.github.com/repos/injunweb/gitops-repo/dispatches",
		bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create GitHub request: %v", err)
	}

	req.Header.Set("Authorization", "token "+config.AppConfig.GithubToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to dispatch GitHub event: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub dispatch failed with status: %s", resp.Status)
	}

	return nil
}

func TriggerAddExtraHostnameWorkflow(app models.Application, hostname string) error {
	payload := map[string]interface{}{
		"event_type": "add-extra-hostname",
		"client_payload": map[string]string{
			"appName":  app.Name,
			"hostname": hostname,
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON payload: %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.github.com/repos/injunweb/gitops-repo/dispatches",
		bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create GitHub request: %v", err)
	}

	req.Header.Set("Authorization", "token "+config.AppConfig.GithubToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to dispatch GitHub event: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub dispatch failed with status: %s", resp.Status)
	}

	return nil
}

func TriggerDeleteExtraHostnameWorkflow(app models.Application, hostname string) error {
	payload := map[string]interface{}{
		"event_type": "delete-extra-hostname",
		"client_payload": map[string]string{
			"appName":  app.Name,
			"hostname": hostname,
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON payload: %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.github.com/repos/injunweb/gitops-repo/dispatches",
		bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create GitHub request: %v", err)
	}

	req.Header.Set("Authorization", "token "+config.AppConfig.GithubToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to dispatch GitHub event: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub dispatch failed with status: %s", resp.Status)
	}

	return nil
}

func TriggerUpdatePrimaryHostname(app models.Application, hostname string) error {
	payload := map[string]interface{}{
		"event_type": "update-primary-hostname",
		"client_payload": map[string]string{
			"appName":  app.Name,
			"hostname": hostname,
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON payload: %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.github.com/repos/injunweb/gitops-repo/dispatches",
		bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create GitHub request: %v", err)
	}

	req.Header.Set("Authorization", "token "+config.AppConfig.GithubToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to dispatch GitHub event: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub dispatch failed with status: %s", resp.Status)
	}

	return nil
}
