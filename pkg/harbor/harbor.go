package harbor

import (
	"fmt"
	"io"
	"net/http"

	"github.com/injunweb/backend-server/internal/config"
)

func RepositoryExists(repoName string) (bool, error) {
	url := fmt.Sprintf("%s/projects/%s/repositories/%s", config.AppConfig.HarborURL, config.AppConfig.HarborProjectName, repoName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %v", err)
	}

	req.SetBasicAuth(config.AppConfig.HarborUsername, config.AppConfig.HarborPassword)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("failed to check repository. Status code: %d, Response: %s", resp.StatusCode, body)
	}

	return true, nil
}

func DeleteRepository(repoName string) error {
	url := fmt.Sprintf("%s/projects/%s/repositories/%s", config.AppConfig.HarborURL, config.AppConfig.HarborProjectName, repoName)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.SetBasicAuth(config.AppConfig.HarborUsername, config.AppConfig.HarborPassword)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete repository. Status code: %d, Response: %s", resp.StatusCode, body)
	}

	fmt.Printf("Repository %s/%s successfully deleted.\n", config.AppConfig.HarborProjectName, repoName)
	return nil
}
