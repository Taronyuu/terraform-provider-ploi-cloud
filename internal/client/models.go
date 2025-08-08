package client

import (
	"encoding/json"
	"time"
)

type Application struct {
	ID                 int64               `json:"id,omitempty"`
	Name               string              `json:"name"`
	Type               string              `json:"application_type"`
	ApplicationVersion string              `json:"application_version,omitempty"`
	PHPVersion         string              `json:"php_version,omitempty"`
	NodeJSVersion      string              `json:"nodejs_version,omitempty"`
	BuildCommands      []string            `json:"build_commands,omitempty"`
	InitCommands       []string            `json:"init_commands,omitempty"`
	PHPExtensions      []string            `json:"php_extensions,omitempty"`
	PHPSettings        []string            `json:"php_settings,omitempty"`
	HealthCheckPath    string              `json:"health_check_path,omitempty"`
	SchedulerEnabled   bool                `json:"scheduler_enabled,omitempty"`
	Replicas           int64               `json:"replicas,omitempty"`
	CPURequest         string              `json:"cpu_request,omitempty"`
	MemoryRequest      string              `json:"memory_request,omitempty"`
	StartCommand       string              `json:"start_command,omitempty"`
	URL                string              `json:"url,omitempty"`
	Status             string              `json:"status,omitempty"`
	NeedsDeployment    bool                `json:"needs_deployment,omitempty"`
	CustomManifests    string              `json:"custom_manifests,omitempty"`
	RepositoryURL      string              `json:"repository_url,omitempty"`
	RepositoryOwner    string              `json:"repository_owner,omitempty"`
	RepositoryName     string              `json:"repository_name,omitempty"`
	DefaultBranch      string              `json:"default_branch,omitempty"`
	SocialAccountID    int64               `json:"social_account_id,omitempty"`
	Region             string              `json:"region,omitempty"`
	Provider           string              `json:"provider,omitempty"`
	CreatedAt          time.Time           `json:"created_at,omitempty"`
	UpdatedAt          time.Time           `json:"updated_at,omitempty"`
	Domains            []ApplicationDomain `json:"domains,omitempty"`
	Secrets            []ApplicationSecret `json:"secrets,omitempty"`
	Services           []ApplicationService `json:"services,omitempty"`
	Volumes            []ApplicationVolume  `json:"volumes,omitempty"`
}

type ApplicationService struct {
	ID              int64             `json:"id,omitempty"`
	ApplicationID   int64             `json:"application_id"`
	Name            string            `json:"name,omitempty"`
	Type            string            `json:"type"`
	Version         string            `json:"version,omitempty"`
	Status          string            `json:"status,omitempty"`
	Settings        FlexibleSettings  `json:"settings,omitempty"`
	Command         string            `json:"command,omitempty"`
	Replicas        int64             `json:"replicas,omitempty"`
	CPURequest      string            `json:"cpu_request,omitempty"`
	MemoryRequest   string            `json:"memory_request,omitempty"`
	StorageSize     string            `json:"storage_size,omitempty"`
	Extensions      []string          `json:"extensions,omitempty"`
	DebugAccessPort int64             `json:"debug_access_port,omitempty"`
	CreatedAt       time.Time         `json:"created_at,omitempty"`
	UpdatedAt       time.Time         `json:"updated_at,omitempty"`
}

// FlexibleSettings can handle both map[string]string and empty arrays from the API
type FlexibleSettings map[string]string

func (fs *FlexibleSettings) UnmarshalJSON(data []byte) error {
	// First try to unmarshal as a map
	var m map[string]string
	if err := json.Unmarshal(data, &m); err == nil {
		*fs = FlexibleSettings(m)
		return nil
	}
	
	// If that fails, try to unmarshal as an array (which we'll ignore)
	var arr []interface{}
	if err := json.Unmarshal(data, &arr); err == nil {
		// Empty array case - initialize as empty map
		*fs = make(FlexibleSettings)
		return nil
	}
	
	// If both fail, initialize as empty map
	*fs = make(FlexibleSettings)
	return nil
}

func (fs FlexibleSettings) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string(fs))
}

// ToMap converts FlexibleSettings to map[string]string for easier access
func (fs FlexibleSettings) ToMap() map[string]string {
	return map[string]string(fs)
}

func FlexibleSettingsFromMap(m map[string]string) FlexibleSettings {
	return FlexibleSettings(m)
}

type ApplicationDomain struct {
	ID            int64     `json:"id,omitempty"`
	ApplicationID int64     `json:"application_id"`
	Domain        string    `json:"domain"`
	SSLStatus     string    `json:"ssl_status,omitempty"`
	CreatedAt     time.Time `json:"created_at,omitempty"`
	UpdatedAt     time.Time `json:"updated_at,omitempty"`
}

type ApplicationSecret struct {
	ApplicationID int64     `json:"application_id"`
	Key           string    `json:"key"`
	Value         string    `json:"value"`
	CreatedAt     time.Time `json:"created_at,omitempty"`
	UpdatedAt     time.Time `json:"updated_at,omitempty"`
}

type ApplicationVolume struct {
	ID            int64     `json:"id,omitempty"`
	ApplicationID int64     `json:"application_id"`
	Name          string    `json:"name"`
	Size          int64     `json:"size"`
	MountPath     string    `json:"path"`
	ResizeStatus  string    `json:"resize_status,omitempty"`
	StorageClass  string    `json:"storage_class,omitempty"`
	CreatedAt     time.Time `json:"created_at,omitempty"`
	UpdatedAt     time.Time `json:"updated_at,omitempty"`
}

type Worker struct {
	ID            int64     `json:"id,omitempty"`
	ApplicationID int64     `json:"application_id"`
	Name          string    `json:"name"`
	Command       string    `json:"command"`
	Type          string    `json:"type,omitempty"`
	Replicas      int64     `json:"replicas"`
	MemoryRequest string    `json:"memory_request,omitempty"`
	CPURequest    string    `json:"cpu_request,omitempty"`
	Status        string    `json:"status,omitempty"`
	CreatedAt     time.Time `json:"created_at,omitempty"`
	UpdatedAt     time.Time `json:"updated_at,omitempty"`
}

type Team struct {
	ID        int64     `json:"id,omitempty"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

type ErrorResponse struct {
	Message string                 `json:"message"`
	Errors  map[string]interface{} `json:"errors,omitempty"`
}

type ListResponse[T any] struct {
	Success bool           `json:"success,omitempty"`
	Message *string        `json:"message,omitempty"`
	Data    []T            `json:"data"`
	Links   map[string]string `json:"links,omitempty"`
	Meta    map[string]interface{} `json:"meta,omitempty"`
}

type SingleResponse[T any] struct {
	Success bool    `json:"success,omitempty"`
	Message *string `json:"message,omitempty"`
	Data    T       `json:"data"`
}