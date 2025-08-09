package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	httpClient  *http.Client
	apiToken    string
	apiEndpoint string
	logger      *Logger
}

// Logger provides structured logging for API requests and responses
type Logger struct {
	enabled bool
	debug   bool
}

// LogEntry represents a structured log entry for an API call
type LogEntry struct {
	Timestamp    time.Time `json:"timestamp"`
	Method       string    `json:"method"`
	URL          string    `json:"url"`
	RequestBody  string    `json:"request_body,omitempty"`
	StatusCode   int       `json:"status_code"`
	ResponseBody string    `json:"response_body,omitempty"`
	Error        string    `json:"error,omitempty"`
	Duration     time.Duration `json:"duration"`
}

// DetailedError provides enhanced error information with actionable feedback
type DetailedError struct {
	StatusCode int                 `json:"status_code"`
	Message    string              `json:"message"`
	Errors     map[string][]string `json:"errors,omitempty"`
	Suggestion string              `json:"suggestion,omitempty"`
	DocsLink   string              `json:"docs_link,omitempty"`
}

func NewClient(apiToken string, apiEndpoint *string) *Client {
	endpoint := "https://cloud.ploi.io/api/v1"
	if apiEndpoint != nil && *apiEndpoint != "" {
		endpoint = *apiEndpoint
	}

	// Initialize logger based on environment variables
	logger := &Logger{
		enabled: os.Getenv("TF_LOG") == "DEBUG" || os.Getenv("PLOI_DEBUG") == "1",
		debug:   os.Getenv("TF_LOG") == "DEBUG" || os.Getenv("PLOI_DEBUG") == "1",
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiToken:    apiToken,
		apiEndpoint: endpoint,
		logger:      logger,
	}
}

func (c *Client) doRequest(method, path string, body interface{}) (*http.Response, error) {
	return c.doRequestWithRetry(method, path, body, 3)
}

func (c *Client) doRequestWithRetry(method, path string, body interface{}, maxRetries int) (*http.Response, error) {
	var lastResp *http.Response
	var lastErr error
	
	for attempt := 0; attempt <= maxRetries; attempt++ {
		start := time.Now()
		
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
		var bodyBytes []byte
		var requestBodyStr string

		url := c.apiEndpoint + path
		
		if body != nil {
			bodyBytes, err = json.Marshal(body)
			if err != nil {
				c.logRequest(method, url, "", 0, "", fmt.Sprintf("failed to marshal request body: %v", err), time.Since(start))
				return nil, fmt.Errorf("failed to marshal request body: %w", err)
			}
			if bodyBytes == nil {
				bodyBytes = []byte{}
			}
			requestBodyStr = c.sanitizeBody(string(bodyBytes))
			req, err = http.NewRequest(method, url, bytes.NewReader(bodyBytes))
		} else {
			req, err = http.NewRequest(method, url, nil)
		}
		
		if err != nil {
			c.logRequest(method, url, requestBodyStr, 0, "", fmt.Sprintf("failed to create HTTP request: %v", err), time.Since(start))
			return nil, fmt.Errorf("failed to create HTTP request: %w", err)
		}
		
		if req == nil {
			c.logRequest(method, url, requestBodyStr, 0, "", "request is nil after creation", time.Since(start))
			return nil, fmt.Errorf("request is nil after creation")
		}

		req.Header.Set("Authorization", "Bearer "+c.apiToken)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			c.logRequest(method, url, requestBodyStr, 0, "", fmt.Sprintf("failed to execute HTTP request: %v", err), time.Since(start))
			
			if attempt < maxRetries {
				backoffDuration := time.Duration(attempt+1) * time.Second
				c.logRequest(method, url, requestBodyStr, 0, "", fmt.Sprintf("retrying in %v (attempt %d/%d)", backoffDuration, attempt+1, maxRetries+1), time.Since(start))
				time.Sleep(backoffDuration)
				continue
			}
			return nil, fmt.Errorf("failed to execute HTTP request after %d attempts: %w", maxRetries+1, err)
		}

		// Read response body for logging and error handling
		var responseBodyStr string
		if resp.Body != nil {
			bodyBytes, readErr := io.ReadAll(resp.Body)
			resp.Body.Close()
			if readErr == nil {
				responseBodyStr = string(bodyBytes)
				// Recreate the response body for the caller
				resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			}
		}

		// Log the completed request
		var errorMsg string
		if resp.StatusCode >= 400 {
			errorMsg = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, resp.Status)
		}
		
		// Check if we should retry based on status code
		if resp.StatusCode >= 500 && resp.StatusCode < 600 && attempt < maxRetries {
			lastResp = resp
			backoffDuration := time.Duration(attempt+1) * time.Second
			c.logRequest(method, url, requestBodyStr, resp.StatusCode, responseBodyStr, fmt.Sprintf("%s - retrying in %v (attempt %d/%d)", errorMsg, backoffDuration, attempt+1, maxRetries+1), time.Since(start))
			time.Sleep(backoffDuration)
			continue
		}
		
		c.logRequest(method, url, requestBodyStr, resp.StatusCode, responseBodyStr, errorMsg, time.Since(start))
		return resp, nil
	}

	// If we get here, all retries failed
	if lastResp != nil {
		return lastResp, nil
	}
	return nil, lastErr
}

func (c *Client) CreateApplication(app *Application) (*Application, error) {
	resp, err := c.doRequest("POST", "/applications", app)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp, "create application")
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
		return nil, c.handleErrorResponse(resp, "get application")
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
		return nil, c.handleErrorResponse(resp, "update application")
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
		return c.handleErrorResponse(resp, "delete application")
	}

	return nil
}

func (c *Client) DeployApplication(id int64) error {
	resp, err := c.doRequest("POST", fmt.Sprintf("/applications/%d/deploy", id), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return fmt.Errorf("failed to deploy application: %s", resp.Status)
		}
		return fmt.Errorf("failed to deploy application: %s", errResp.Message)
	}

	return nil
}

func (c *Client) CreateService(service *ApplicationService) (*ApplicationService, error) {
	// Validate service before making API request
	if err := c.ValidateServiceRequest(service); err != nil {
		return nil, err
	}

	resp, err := c.doRequest("POST", fmt.Sprintf("/applications/%d/services", service.ApplicationID), service)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp, "create service")
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

// logRequest logs API request details with sanitized sensitive information
func (c *Client) logRequest(method, url, requestBody string, statusCode int, responseBody, errorMsg string, duration time.Duration) {
	if !c.logger.enabled {
		return
	}

	// Create log entry for structured logging (can be used for external log systems)
	_ = LogEntry{
		Timestamp:    time.Now(),
		Method:       method,
		URL:          c.sanitizeURL(url),
		RequestBody:  requestBody,
		StatusCode:   statusCode,
		ResponseBody: responseBody,
		Error:        errorMsg,
		Duration:     duration,
	}

	if c.logger.debug {
		// Detailed logging for debug mode
		log.Printf("[DEBUG] Ploi API Request: %s %s", method, c.sanitizeURL(url))
		if requestBody != "" {
			log.Printf("[DEBUG] Request Body: %s", requestBody)
		}
		if statusCode > 0 {
			log.Printf("[DEBUG] Response Status: %d", statusCode)
			if responseBody != "" {
				log.Printf("[DEBUG] Response Body: %s", responseBody)
			}
		}
		if errorMsg != "" {
			log.Printf("[DEBUG] Error: %s", errorMsg)
		}
		log.Printf("[DEBUG] Duration: %v", duration)
	} else {
		// Compact logging for normal mode
		if errorMsg != "" {
			log.Printf("[ERROR] Ploi API %s %s failed: %s (took %v)", method, c.sanitizeURL(url), errorMsg, duration)
		} else {
			log.Printf("[INFO] Ploi API %s %s: %d (took %v)", method, c.sanitizeURL(url), statusCode, duration)
		}
	}
}

// sanitizeToken masks API token for logging
func (c *Client) sanitizeToken(token string) string {
	if len(token) <= 8 {
		return "***"
	}
	return token[:4] + "***" + token[len(token)-4:]
}

// sanitizeURL removes sensitive information from URL for logging
func (c *Client) sanitizeURL(url string) string {
	// Remove any potential sensitive information from query parameters
	if strings.Contains(url, "?") {
		parts := strings.Split(url, "?")
		return parts[0] + "?[params sanitized]"
	}
	return url
}

// sanitizeBody sanitizes request/response body for logging
func (c *Client) sanitizeBody(body string) string {
	// For now, just return the body as-is since we're not storing actual secrets in service configs
	// In the future, we could add more sophisticated sanitization
	return body
}

// handleErrorResponse processes error responses and returns detailed error information
func (c *Client) handleErrorResponse(resp *http.Response, operation string) error {
	var errResp ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		return fmt.Errorf("failed to %s: %s", operation, resp.Status)
	}

	detailedErr := &DetailedError{
		StatusCode: resp.StatusCode,
		Message:    errResp.Message,
		DocsLink:   "https://docs.ploi.io/cloud",
	}

	// Convert error map to detailed format
	if len(errResp.Errors) > 0 {
		detailedErr.Errors = make(map[string][]string)
		for field, value := range errResp.Errors {
			switch v := value.(type) {
			case string:
				detailedErr.Errors[field] = []string{v}
			case []interface{}:
				messages := make([]string, len(v))
				for i, msg := range v {
					if str, ok := msg.(string); ok {
						messages[i] = str
					} else {
						messages[i] = fmt.Sprintf("%v", msg)
					}
				}
				detailedErr.Errors[field] = messages
			case []string:
				detailedErr.Errors[field] = v
			default:
				detailedErr.Errors[field] = []string{fmt.Sprintf("%v", v)}
			}
		}
	}

	// Add specific suggestions based on status code
	switch resp.StatusCode {
	case 422:
		detailedErr.Suggestion = c.generateValidationSuggestion(operation, detailedErr.Errors)
	case 404:
		detailedErr.Suggestion = "Check that the resource exists and the ID is correct"
	case 401:
		detailedErr.Suggestion = "Check that your API token is valid and has the required permissions"
	case 403:
		detailedErr.Suggestion = "Check that your API token has permission to perform this operation"
	case 500, 502, 503, 504:
		detailedErr.Suggestion = "This appears to be a server error. Please try again in a few moments"
	}

	return fmt.Errorf("failed to %s: %s\nSuggestion: %s\nDocumentation: %s",
		operation, detailedErr.Message, detailedErr.Suggestion, detailedErr.DocsLink)
}

// generateValidationSuggestion provides helpful suggestions for validation errors
func (c *Client) generateValidationSuggestion(operation string, errors map[string][]string) string {
	if len(errors) == 0 {
		return "Check the API documentation for required fields and valid values"
	}

	suggestions := []string{}
	
	for field, messages := range errors {
		switch field {
		case "type":
			suggestions = append(suggestions, "Service type must be one of: mysql, postgresql, redis, valkey, rabbitmq, mongodb, minio, sftp")
		case "version":
			suggestions = append(suggestions, "Check that the version is supported for the selected service type")
		case "storage_size":
			suggestions = append(suggestions, "Storage size must be specified with units (e.g., '1Gi', '10Gi')")
		case "memory_request":
			suggestions = append(suggestions, "Memory request must be specified with units (e.g., '256Mi', '1Gi')")
		case "cpu_request":
			suggestions = append(suggestions, "CPU request must be specified correctly (e.g., '250m', '1', '2')")
		default:
			suggestions = append(suggestions, fmt.Sprintf("Field '%s': %s", field, strings.Join(messages, ", ")))
		}
	}

	if len(suggestions) > 0 {
		return strings.Join(suggestions, "; ")
	}

	return "Check the API documentation for required fields and valid values"
}

// ValidateServiceRequest validates service configuration before API request
func (c *Client) ValidateServiceRequest(service *ApplicationService) error {
	if service == nil {
		return fmt.Errorf("service cannot be nil")
	}

	if service.ApplicationID <= 0 {
		return fmt.Errorf("application_id must be greater than 0")
	}

	if service.Type == "" {
		return fmt.Errorf("service type is required")
	}

	// Validate service type
	validTypes := map[string]bool{
		"mysql":      true,
		"postgresql": true,
		"redis":      true,
		"valkey":     true,
		"rabbitmq":   true,
		"mongodb":    true,
		"minio":      true,
		"sftp":       true,
		"worker":     true,
	}

	if !validTypes[service.Type] {
		return fmt.Errorf("invalid service type '%s'. Must be one of: mysql, postgresql, redis, valkey, rabbitmq, mongodb, minio, sftp, worker", service.Type)
	}

	// Validate that worker services have a command (either direct field or in settings)
	if service.Type == "worker" {
		hasCommand := service.Command != ""
		if !hasCommand && len(service.Settings) > 0 {
			settingsMap := service.Settings.ToMap()
			if cmd, ok := settingsMap["command"]; ok && cmd != "" {
				hasCommand = true
			}
		}
		if !hasCommand {
			return fmt.Errorf("command is required for worker type services")
		}
	}

	// Validate resource specifications if provided
	if service.MemoryRequest != "" && !isValidResourceSpec(service.MemoryRequest, []string{"Mi", "Gi"}) {
		return fmt.Errorf("invalid memory_request format '%s'. Use format like '256Mi' or '1Gi'", service.MemoryRequest)
	}

	if service.CPURequest != "" && !isValidCPUSpec(service.CPURequest) {
		return fmt.Errorf("invalid cpu_request format '%s'. Use format like '250m', '1', or '2'", service.CPURequest)
	}

	if service.StorageSize != "" && !isValidResourceSpec(service.StorageSize, []string{"Mi", "Gi", "Ti"}) {
		return fmt.Errorf("invalid storage_size format '%s'. Use format like '1Gi' or '10Gi'", service.StorageSize)
	}

	return nil
}

// isValidResourceSpec validates Kubernetes resource specification format
func isValidResourceSpec(spec string, validUnits []string) bool {
	if spec == "" {
		return false
	}

	for _, unit := range validUnits {
		if strings.HasSuffix(spec, unit) {
			numberPart := strings.TrimSuffix(spec, unit)
			if _, err := strconv.ParseFloat(numberPart, 64); err == nil {
				return true
			}
		}
	}
	return false
}

// isValidCPUSpec validates CPU specification format
func isValidCPUSpec(spec string) bool {
	if spec == "" {
		return false
	}

	// Check for millicores (e.g., "250m")
	if strings.HasSuffix(spec, "m") {
		numberPart := strings.TrimSuffix(spec, "m")
		if _, err := strconv.ParseInt(numberPart, 10, 64); err == nil {
			return true
		}
	}

	// Check for whole cores (e.g., "1", "2")
	if _, err := strconv.ParseFloat(spec, 64); err == nil {
		return true
	}

	return false
}