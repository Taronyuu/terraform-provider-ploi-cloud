package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	httpClient  *http.Client
	apiToken    string
	apiEndpoint string
}

func NewClient(apiToken string, apiEndpoint *string) *Client {
	endpoint := "https://cloud.ploi.io/api/v1"
	if apiEndpoint != nil && *apiEndpoint != "" {
		endpoint = *apiEndpoint
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiToken:    apiToken,
		apiEndpoint: endpoint,
	}
}

func (c *Client) doRequest(method, path string, body interface{}) (*http.Response, error) {
	var bodyReader *bytes.Reader

	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(method, c.apiEndpoint+path, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return c.httpClient.Do(req)
}

func (c *Client) CreateApplication(app *Application) (*Application, error) {
	resp, err := c.doRequest("POST", "/applications", app)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("failed to create application: %s", resp.Status)
		}
		return nil, fmt.Errorf("failed to create application: %s", errResp.Message)
	}

	var result SingleResponse[Application]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (c *Client) GetApplication(id int64) (*Application, error) {
	resp, err := c.doRequest("GET", fmt.Sprintf("/applications/%d", id), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("failed to get application: %s", resp.Status)
		}
		return nil, fmt.Errorf("failed to get application: %s", errResp.Message)
	}

	var result SingleResponse[Application]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (c *Client) UpdateApplication(id int64, app *Application) (*Application, error) {
	resp, err := c.doRequest("PUT", fmt.Sprintf("/applications/%d", id), app)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("failed to update application: %s", resp.Status)
		}
		return nil, fmt.Errorf("failed to update application: %s", errResp.Message)
	}

	var result SingleResponse[Application]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (c *Client) DeleteApplication(id int64) error {
	resp, err := c.doRequest("DELETE", fmt.Sprintf("/applications/%d", id), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return fmt.Errorf("failed to delete application: %s", resp.Status)
		}
		return fmt.Errorf("failed to delete application: %s", errResp.Message)
	}

	return nil
}

func (c *Client) DeployApplication(id int64) error {
	resp, err := c.doRequest("POST", fmt.Sprintf("/applications/%d/deploy", id), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return fmt.Errorf("failed to deploy application: %s", resp.Status)
		}
		return fmt.Errorf("failed to deploy application: %s", errResp.Message)
	}

	return nil
}

func (c *Client) CreateService(service *ApplicationService) (*ApplicationService, error) {
	resp, err := c.doRequest("POST", fmt.Sprintf("/applications/%d/services", service.ApplicationID), service)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("failed to create service: %s", resp.Status)
		}
		return nil, fmt.Errorf("failed to create service: %s", errResp.Message)
	}

	var result SingleResponse[ApplicationService]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (c *Client) GetService(applicationID, serviceID int64) (*ApplicationService, error) {
	resp, err := c.doRequest("GET", fmt.Sprintf("/applications/%d/services/%d", applicationID, serviceID), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("failed to get service: %s", resp.Status)
		}
		return nil, fmt.Errorf("failed to get service: %s", errResp.Message)
	}

	var result SingleResponse[ApplicationService]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (c *Client) UpdateService(applicationID, serviceID int64, service *ApplicationService) (*ApplicationService, error) {
	resp, err := c.doRequest("PUT", fmt.Sprintf("/applications/%d/services/%d", applicationID, serviceID), service)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("failed to update service: %s", resp.Status)
		}
		return nil, fmt.Errorf("failed to update service: %s", errResp.Message)
	}

	var result SingleResponse[ApplicationService]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (c *Client) DeleteService(applicationID, serviceID int64) error {
	resp, err := c.doRequest("DELETE", fmt.Sprintf("/applications/%d/services/%d", applicationID, serviceID), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return fmt.Errorf("failed to delete service: %s", resp.Status)
		}
		return fmt.Errorf("failed to delete service: %s", errResp.Message)
	}

	return nil
}

func (c *Client) CreateDomain(domain *ApplicationDomain) (*ApplicationDomain, error) {
	resp, err := c.doRequest("POST", fmt.Sprintf("/applications/%d/domains", domain.ApplicationID), domain)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("failed to create domain: %s", resp.Status)
		}
		return nil, fmt.Errorf("failed to create domain: %s", errResp.Message)
	}

	var result SingleResponse[ApplicationDomain]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (c *Client) GetDomain(applicationID, domainID int64) (*ApplicationDomain, error) {
	resp, err := c.doRequest("GET", fmt.Sprintf("/applications/%d/domains/%d", applicationID, domainID), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("failed to get domain: %s", resp.Status)
		}
		return nil, fmt.Errorf("failed to get domain: %s", errResp.Message)
	}

	var result SingleResponse[ApplicationDomain]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (c *Client) DeleteDomain(applicationID, domainID int64) error {
	resp, err := c.doRequest("DELETE", fmt.Sprintf("/applications/%d/domains/%d", applicationID, domainID), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return fmt.Errorf("failed to delete domain: %s", resp.Status)
		}
		return fmt.Errorf("failed to delete domain: %s", errResp.Message)
	}

	return nil
}

func (c *Client) CreateSecret(secret *ApplicationSecret) (*ApplicationSecret, error) {
	resp, err := c.doRequest("POST", fmt.Sprintf("/applications/%d/secrets", secret.ApplicationID), secret)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("failed to create secret: %s", resp.Status)
		}
		return nil, fmt.Errorf("failed to create secret: %s", errResp.Message)
	}

	var result SingleResponse[ApplicationSecret]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (c *Client) GetSecret(applicationID int64, key string) (*ApplicationSecret, error) {
	resp, err := c.doRequest("GET", fmt.Sprintf("/applications/%d/secrets/%s", applicationID, key), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("failed to get secret: %s", resp.Status)
		}
		return nil, fmt.Errorf("failed to get secret: %s", errResp.Message)
	}

	var result SingleResponse[ApplicationSecret]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (c *Client) UpdateSecret(applicationID int64, key string, secret *ApplicationSecret) (*ApplicationSecret, error) {
	resp, err := c.doRequest("PUT", fmt.Sprintf("/applications/%d/secrets/%s", applicationID, key), secret)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("failed to update secret: %s", resp.Status)
		}
		return nil, fmt.Errorf("failed to update secret: %s", errResp.Message)
	}

	var result SingleResponse[ApplicationSecret]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (c *Client) DeleteSecret(applicationID int64, key string) error {
	resp, err := c.doRequest("DELETE", fmt.Sprintf("/applications/%d/secrets/%s", applicationID, key), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return fmt.Errorf("failed to delete secret: %s", resp.Status)
		}
		return fmt.Errorf("failed to delete secret: %s", errResp.Message)
	}

	return nil
}

func (c *Client) CreateVolume(volume *ApplicationVolume) (*ApplicationVolume, error) {
	resp, err := c.doRequest("POST", fmt.Sprintf("/applications/%d/volumes", volume.ApplicationID), volume)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("failed to create volume: %s", resp.Status)
		}
		return nil, fmt.Errorf("failed to create volume: %s", errResp.Message)
	}

	var result SingleResponse[ApplicationVolume]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (c *Client) GetVolume(applicationID, volumeID int64) (*ApplicationVolume, error) {
	resp, err := c.doRequest("GET", fmt.Sprintf("/applications/%d/volumes/%d", applicationID, volumeID), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("failed to get volume: %s", resp.Status)
		}
		return nil, fmt.Errorf("failed to get volume: %s", errResp.Message)
	}

	var result SingleResponse[ApplicationVolume]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (c *Client) UpdateVolume(applicationID, volumeID int64, volume *ApplicationVolume) (*ApplicationVolume, error) {
	resp, err := c.doRequest("PUT", fmt.Sprintf("/applications/%d/volumes/%d", applicationID, volumeID), volume)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("failed to update volume: %s", resp.Status)
		}
		return nil, fmt.Errorf("failed to update volume: %s", errResp.Message)
	}

	var result SingleResponse[ApplicationVolume]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (c *Client) DeleteVolume(applicationID, volumeID int64) error {
	resp, err := c.doRequest("DELETE", fmt.Sprintf("/applications/%d/volumes/%d", applicationID, volumeID), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return fmt.Errorf("failed to delete volume: %s", resp.Status)
		}
		return fmt.Errorf("failed to delete volume: %s", errResp.Message)
	}

	return nil
}

func (c *Client) CreateWorker(worker *Worker) (*Worker, error) {
	resp, err := c.doRequest("POST", fmt.Sprintf("/applications/%d/services", worker.ApplicationID), worker)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("failed to create worker: %s", resp.Status)
		}
		return nil, fmt.Errorf("failed to create worker: %s", errResp.Message)
	}

	var result SingleResponse[Worker]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (c *Client) GetWorker(applicationID, workerID int64) (*Worker, error) {
	resp, err := c.doRequest("GET", fmt.Sprintf("/applications/%d/services/%d", applicationID, workerID), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("failed to get worker: %s", resp.Status)
		}
		return nil, fmt.Errorf("failed to get worker: %s", errResp.Message)
	}

	var result SingleResponse[Worker]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (c *Client) UpdateWorker(applicationID, workerID int64, worker *Worker) (*Worker, error) {
	resp, err := c.doRequest("PUT", fmt.Sprintf("/applications/%d/services/%d", applicationID, workerID), worker)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("failed to update worker: %s", resp.Status)
		}
		return nil, fmt.Errorf("failed to update worker: %s", errResp.Message)
	}

	var result SingleResponse[Worker]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (c *Client) DeleteWorker(applicationID, workerID int64) error {
	resp, err := c.doRequest("DELETE", fmt.Sprintf("/applications/%d/services/%d", applicationID, workerID), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return fmt.Errorf("failed to delete worker: %s", resp.Status)
		}
		return fmt.Errorf("failed to delete worker: %s", errResp.Message)
	}

	return nil
}