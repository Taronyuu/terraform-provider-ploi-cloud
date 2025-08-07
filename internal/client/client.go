package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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
	if c == nil {
		return nil, fmt.Errorf("client is nil")
	}
	if c.httpClient == nil {
		return nil, fmt.Errorf("http client is nil")
	}
	if c.apiEndpoint == "" {
		return nil, fmt.Errorf("api endpoint is empty")
	}
	if c.apiToken == "" {
		return nil, fmt.Errorf("api token is empty")
	}

	var req *http.Request
	var err error

	url := c.apiEndpoint + path
	
	if body != nil {
		bodyBytes, jsonErr := json.Marshal(body)
		if jsonErr != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", jsonErr)
		}
		if bodyBytes == nil {
			bodyBytes = []byte{}
		}
		req, err = http.NewRequest(method, url, bytes.NewReader(bodyBytes))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	
	if req == nil {
		return nil, fmt.Errorf("request is nil after creation")
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Debug logging if TF_LOG=DEBUG or PLOI_DEBUG=1
	debug := os.Getenv("TF_LOG") == "DEBUG" || os.Getenv("PLOI_DEBUG") == "1"
	if debug {
		fmt.Printf("[DEBUG] Ploi API Request: %s %s\n", method, url)
		if body != nil {
			bodyBytes, _ := json.MarshalIndent(body, "", "  ")
			fmt.Printf("[DEBUG] Request Body:\n%s\n", string(bodyBytes))
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	if debug {
		fmt.Printf("[DEBUG] Response Status: %s\n", resp.Status)
		if resp.Body != nil {
			// Read response body for debugging
			bodyBytes, readErr := io.ReadAll(resp.Body)
			resp.Body.Close()
			if readErr == nil {
				fmt.Printf("[DEBUG] Response Body:\n%s\n", string(bodyBytes))
				// Recreate the response body for the caller
				resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			}
		}
	}

	return resp, nil
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

func (c *Client) UpdateApplication(id int64, updateData interface{}) (*Application, error) {
	resp, err := c.doRequest("PUT", fmt.Sprintf("/applications/%d", id), updateData)
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
	// Since the API doesn't support GET for individual services, 
	// we get the application and find the service in its services list
	app, err := c.GetApplication(applicationID)
	if err != nil {
		return nil, err
	}
	
	if app == nil {
		return nil, nil
	}
	
	// Find the service with matching ID
	for _, service := range app.Services {
		if service.ID == serviceID {
			// Ensure ApplicationID is set (it might not be in the nested response)
			service.ApplicationID = applicationID
			return &service, nil
		}
	}
	
	// Service not found
	return nil, nil
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
	// Get all secrets and filter by key since individual secret GET is not supported
	resp, err := c.doRequest("GET", fmt.Sprintf("/applications/%d/secrets", applicationID), nil)
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
			return nil, fmt.Errorf("failed to get secrets: %s", resp.Status)
		}
		return nil, fmt.Errorf("failed to get secrets: %s", errResp.Message)
	}

	var result ListResponse[ApplicationSecret]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Find the secret with the matching key
	for _, secret := range result.Data {
		if secret.Key == key {
			return &secret, nil
		}
	}

	return nil, nil // Secret not found
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
	resp, err := c.doRequest("POST", fmt.Sprintf("/applications/%d/workers", worker.ApplicationID), worker)
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
	resp, err := c.doRequest("GET", fmt.Sprintf("/applications/%d/workers/%d", applicationID, workerID), nil)
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
	resp, err := c.doRequest("PUT", fmt.Sprintf("/applications/%d/workers/%d", applicationID, workerID), worker)
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
	resp, err := c.doRequest("DELETE", fmt.Sprintf("/applications/%d/workers/%d", applicationID, workerID), nil)
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