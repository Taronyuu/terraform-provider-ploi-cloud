package client

import "time"

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
	Type            string            `json:"type"`
	Version         string            `json:"version,omitempty"`
	Status          string            `json:"status,omitempty"`
	Settings        map[string]string `json:"settings,omitempty"`
	Command         string            `json:"command,omitempty"`
	Replicas        int64             `json:"replicas,omitempty"`
	DebugAccessPort int64             `json:"debug_access_port,omitempty"`
	CreatedAt       time.Time         `json:"created_at,omitempty"`
	UpdatedAt       time.Time         `json:"updated_at,omitempty"`
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
	CreatedAt     time.Time `json:"created_at,omitempty"`
	UpdatedAt     time.Time `json:"updated_at,omitempty"`
}

type Worker struct {
	ID            int64     `json:"id,omitempty"`
	ApplicationID int64     `json:"application_id"`
	Name          string    `json:"name"`
	Command       string    `json:"command"`
	Replicas      int64     `json:"replicas"`
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
	Message string            `json:"message"`
	Errors  map[string]string `json:"errors,omitempty"`
}

type ListResponse[T any] struct {
	Data  []T            `json:"data"`
	Links map[string]string `json:"links,omitempty"`
	Meta  map[string]interface{} `json:"meta,omitempty"`
}

type SingleResponse[T any] struct {
	Data T `json:"data"`
}