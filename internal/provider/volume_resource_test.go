package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ploi/terraform-provider-ploicloud/internal/client"
)

func TestVolumeResource_Schema(t *testing.T) {
	r := NewVolumeResource()
	
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}
	
	r.Schema(context.Background(), req, resp)

	if resp.Schema.Attributes == nil {
		t.Fatal("Schema attributes should not be nil")
	}
}

func TestVolumeResource_StorageClass_toAPIModel(t *testing.T) {
	resource := &VolumeResource{}
	
	tests := []struct {
		name              string
		data              *VolumeResourceModel
		expectedStorageClass string
	}{
		{
			name: "volume with fast-ssd storage class",
			data: &VolumeResourceModel{
				ID:            types.Int64Value(1),
				ApplicationID: types.Int64Value(100),
				Name:          types.StringValue("fast-volume"),
				Size:          types.Int64Value(10),
				MountPath:     types.StringValue("/var/lib/data"),
				StorageClass:  types.StringValue("fast-ssd"),
			},
			expectedStorageClass: "fast-ssd",
		},
		{
			name: "volume with standard storage class",
			data: &VolumeResourceModel{
				ID:            types.Int64Value(2),
				ApplicationID: types.Int64Value(100),
				Name:          types.StringValue("standard-volume"),
				Size:          types.Int64Value(50),
				MountPath:     types.StringValue("/var/lib/storage"),
				StorageClass:  types.StringValue("standard"),
			},
			expectedStorageClass: "standard",
		},
		{
			name: "volume with ssd storage class",
			data: &VolumeResourceModel{
				ID:            types.Int64Value(3),
				ApplicationID: types.Int64Value(100),
				Name:          types.StringValue("ssd-volume"),
				Size:          types.Int64Value(25),
				MountPath:     types.StringValue("/var/lib/cache"),
				StorageClass:  types.StringValue("ssd"),
			},
			expectedStorageClass: "ssd",
		},
		{
			name: "volume with null storage class",
			data: &VolumeResourceModel{
				ID:            types.Int64Value(4),
				ApplicationID: types.Int64Value(100),
				Name:          types.StringValue("default-volume"),
				Size:          types.Int64Value(20),
				MountPath:     types.StringValue("/var/lib/default"),
				StorageClass:  types.StringNull(),
			},
			expectedStorageClass: "",
		},
		{
			name: "volume with empty storage class",
			data: &VolumeResourceModel{
				ID:            types.Int64Value(5),
				ApplicationID: types.Int64Value(100),
				Name:          types.StringValue("empty-class-volume"),
				Size:          types.Int64Value(15),
				MountPath:     types.StringValue("/var/lib/empty"),
				StorageClass:  types.StringValue(""),
			},
			expectedStorageClass: "",
		},
		{
			name: "volume with custom storage class",
			data: &VolumeResourceModel{
				ID:            types.Int64Value(6),
				ApplicationID: types.Int64Value(100),
				Name:          types.StringValue("custom-volume"),
				Size:          types.Int64Value(100),
				MountPath:     types.StringValue("/var/lib/custom"),
				StorageClass:  types.StringValue("nvme-local"),
			},
			expectedStorageClass: "nvme-local",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resource.toAPIModel(tt.data)
			
			if result.StorageClass != tt.expectedStorageClass {
				t.Errorf("Expected StorageClass '%s', got '%s'", tt.expectedStorageClass, result.StorageClass)
			}
			
			// Verify other fields are preserved
			if result.ID != tt.data.ID.ValueInt64() {
				t.Errorf("Expected ID %d, got %d", tt.data.ID.ValueInt64(), result.ID)
			}
			if result.ApplicationID != tt.data.ApplicationID.ValueInt64() {
				t.Errorf("Expected ApplicationID %d, got %d", tt.data.ApplicationID.ValueInt64(), result.ApplicationID)
			}
			if result.Name != tt.data.Name.ValueString() {
				t.Errorf("Expected Name '%s', got '%s'", tt.data.Name.ValueString(), result.Name)
			}
			if result.Size != tt.data.Size.ValueInt64() {
				t.Errorf("Expected Size %d, got %d", tt.data.Size.ValueInt64(), result.Size)
			}
			if result.MountPath != tt.data.MountPath.ValueString() {
				t.Errorf("Expected MountPath '%s', got '%s'", tt.data.MountPath.ValueString(), result.MountPath)
			}
		})
	}
}

func TestVolumeResource_StorageClass_fromAPIModel(t *testing.T) {
	resource := &VolumeResource{}
	
	tests := []struct {
		name                 string
		volume               *client.ApplicationVolume
		expectedStorageClass types.String
	}{
		{
			name: "volume with fast-ssd storage class from API",
			volume: &client.ApplicationVolume{
				ID:            1,
				ApplicationID: 100,
				Name:          "fast-volume",
				Size:          10,
				MountPath:     "/var/lib/data",
				StorageClass:  "fast-ssd",
				ResizeStatus:  "completed",
			},
			expectedStorageClass: types.StringValue("fast-ssd"),
		},
		{
			name: "volume with standard storage class from API",
			volume: &client.ApplicationVolume{
				ID:            2,
				ApplicationID: 100,
				Name:          "standard-volume",
				Size:          50,
				MountPath:     "/var/lib/storage",
				StorageClass:  "standard",
				ResizeStatus:  "completed",
			},
			expectedStorageClass: types.StringValue("standard"),
		},
		{
			name: "volume with empty storage class from API",
			volume: &client.ApplicationVolume{
				ID:            3,
				ApplicationID: 100,
				Name:          "no-class-volume",
				Size:          25,
				MountPath:     "/var/lib/default",
				StorageClass:  "",
				ResizeStatus:  "completed",
			},
			expectedStorageClass: types.StringValue(""),
		},
		{
			name: "volume with custom storage class from API",
			volume: &client.ApplicationVolume{
				ID:            4,
				ApplicationID: 100,
				Name:          "custom-volume",
				Size:          100,
				MountPath:     "/var/lib/custom",
				StorageClass:  "nvme-local",
				ResizeStatus:  "pending",
			},
			expectedStorageClass: types.StringValue("nvme-local"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var data VolumeResourceModel
			resource.fromAPIModel(tt.volume, &data)
			
			if !data.StorageClass.Equal(tt.expectedStorageClass) {
				t.Errorf("Expected StorageClass %v, got %v", tt.expectedStorageClass, data.StorageClass)
			}
			
			// Verify other fields are set correctly
			expectedID := types.Int64Value(tt.volume.ID)
			if !data.ID.Equal(expectedID) {
				t.Errorf("Expected ID %v, got %v", expectedID, data.ID)
			}
			
			expectedApplicationID := types.Int64Value(tt.volume.ApplicationID)
			if !data.ApplicationID.Equal(expectedApplicationID) {
				t.Errorf("Expected ApplicationID %v, got %v", expectedApplicationID, data.ApplicationID)
			}
			
			expectedName := types.StringValue(tt.volume.Name)
			if !data.Name.Equal(expectedName) {
				t.Errorf("Expected Name %v, got %v", expectedName, data.Name)
			}
			
			expectedSize := types.Int64Value(tt.volume.Size)
			if !data.Size.Equal(expectedSize) {
				t.Errorf("Expected Size %v, got %v", expectedSize, data.Size)
			}
			
			expectedMountPath := types.StringValue(tt.volume.MountPath)
			if !data.MountPath.Equal(expectedMountPath) {
				t.Errorf("Expected MountPath %v, got %v", expectedMountPath, data.MountPath)
			}
			
			expectedResizeStatus := types.StringValue(tt.volume.ResizeStatus)
			if !data.ResizeStatus.Equal(expectedResizeStatus) {
				t.Errorf("Expected ResizeStatus %v, got %v", expectedResizeStatus, data.ResizeStatus)
			}
		})
	}
}

func TestVolumeResource_StorageClass_BackwardCompatibility(t *testing.T) {
	resource := &VolumeResource{}
	
	// Test that existing volume configurations without storage_class still work
	data := &VolumeResourceModel{
		ID:            types.Int64Value(1),
		ApplicationID: types.Int64Value(100),
		Name:          types.StringValue("legacy-volume"),
		Size:          types.Int64Value(20),
		MountPath:     types.StringValue("/var/lib/legacy"),
		// StorageClass is null/unset (backward compatibility)
		StorageClass: types.StringNull(),
	}
	
	result := resource.toAPIModel(data)
	
	// Verify basic fields are preserved
	if result.ID != 1 {
		t.Errorf("Expected ID 1, got %d", result.ID)
	}
	if result.ApplicationID != 100 {
		t.Errorf("Expected ApplicationID 100, got %d", result.ApplicationID)
	}
	if result.Name != "legacy-volume" {
		t.Errorf("Expected Name 'legacy-volume', got %s", result.Name)
	}
	if result.Size != 20 {
		t.Errorf("Expected Size 20, got %d", result.Size)
	}
	if result.MountPath != "/var/lib/legacy" {
		t.Errorf("Expected MountPath '/var/lib/legacy', got %s", result.MountPath)
	}
	
	// Verify storage_class has default/empty value
	if result.StorageClass != "" {
		t.Errorf("Expected StorageClass to be empty, got %s", result.StorageClass)
	}
}

func TestVolumeResource_StorageClass_CommonValues(t *testing.T) {
	resource := &VolumeResource{}
	
	// Test common storage class values
	commonStorageClasses := []string{
		"fast-ssd",
		"standard",
		"ssd",
		"hdd",
		"nvme",
		"local-ssd",
		"network-ssd",
		"balanced",
		"premium",
		"economy",
	}
	
	for i, storageClass := range commonStorageClasses {
		t.Run(fmt.Sprintf("storage_class_%s", storageClass), func(t *testing.T) {
			data := &VolumeResourceModel{
				ID:            types.Int64Value(int64(i + 1)),
				ApplicationID: types.Int64Value(100),
				Name:          types.StringValue(fmt.Sprintf("volume-%s", storageClass)),
				Size:          types.Int64Value(10),
				MountPath:     types.StringValue("/var/lib/data"),
				StorageClass:  types.StringValue(storageClass),
			}
			
			result := resource.toAPIModel(data)
			
			if result.StorageClass != storageClass {
				t.Errorf("Expected StorageClass '%s', got '%s'", storageClass, result.StorageClass)
			}
			
			// Test round-trip conversion
			var convertedData VolumeResourceModel
			apiVolume := &client.ApplicationVolume{
				ID:            result.ID,
				ApplicationID: result.ApplicationID,
				Name:          result.Name,
				Size:          result.Size,
				MountPath:     result.MountPath,
				StorageClass:  result.StorageClass,
				ResizeStatus:  "completed",
			}
			
			resource.fromAPIModel(apiVolume, &convertedData)
			
			if !convertedData.StorageClass.Equal(data.StorageClass) {
				t.Errorf("Round-trip conversion failed for %s: expected %v, got %v", 
					storageClass, data.StorageClass, convertedData.StorageClass)
			}
		})
	}
}

func TestVolumeResource_StorageClass_ComputedBehavior(t *testing.T) {
	resource := &VolumeResource{}
	
	// Test that storage_class behaves as optional + computed field
	tests := []struct {
		name         string
		inputClass   types.String
		apiResponse  string
		expectedResult types.String
	}{
		{
			name:           "user provides storage class",
			inputClass:     types.StringValue("fast-ssd"),
			apiResponse:    "fast-ssd",
			expectedResult: types.StringValue("fast-ssd"),
		},
		{
			name:           "user provides null, API returns default",
			inputClass:     types.StringNull(),
			apiResponse:    "standard",
			expectedResult: types.StringValue("standard"),
		},
		{
			name:           "user provides empty, API returns computed",
			inputClass:     types.StringValue(""),
			apiResponse:    "premium",
			expectedResult: types.StringValue("premium"),
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate API response with computed storage class
			apiVolume := &client.ApplicationVolume{
				ID:            1,
				ApplicationID: 100,
				Name:          "test-volume",
				Size:          10,
				MountPath:     "/var/lib/data",
				StorageClass:  tt.apiResponse,
				ResizeStatus:  "completed",
			}
			
			var data VolumeResourceModel
			data.StorageClass = tt.inputClass // Set initial value
			
			resource.fromAPIModel(apiVolume, &data)
			
			if !data.StorageClass.Equal(tt.expectedResult) {
				t.Errorf("Expected StorageClass %v, got %v", tt.expectedResult, data.StorageClass)
			}
		})
	}
}

func TestVolumeResource_StorageClass_WithOtherFields(t *testing.T) {
	resource := &VolumeResource{}
	
	// Test that storage_class works correctly with other volume fields
	data := &VolumeResourceModel{
		ID:            types.Int64Value(1),
		ApplicationID: types.Int64Value(100),
		Name:          types.StringValue("complete-volume"),
		Size:          types.Int64Value(50),
		MountPath:     types.StringValue("/var/lib/database"),
		StorageClass:  types.StringValue("fast-ssd"),
	}
	
	result := resource.toAPIModel(data)
	
	// Verify storage_class is preserved
	if result.StorageClass != "fast-ssd" {
		t.Errorf("Expected StorageClass 'fast-ssd', got '%s'", result.StorageClass)
	}
	
	// Verify other fields are also preserved
	if result.ID != 1 {
		t.Errorf("Expected ID 1, got %d", result.ID)
	}
	if result.ApplicationID != 100 {
		t.Errorf("Expected ApplicationID 100, got %d", result.ApplicationID)
	}
	if result.Name != "complete-volume" {
		t.Errorf("Expected Name 'complete-volume', got '%s'", result.Name)
	}
	if result.Size != 50 {
		t.Errorf("Expected Size 50, got %d", result.Size)
	}
	if result.MountPath != "/var/lib/database" {
		t.Errorf("Expected MountPath '/var/lib/database', got '%s'", result.MountPath)
	}
}

func TestVolumeResource_StorageClass_APIClientIntegration(t *testing.T) {
	// Mock server for testing API interactions
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		switch r.URL.Path {
		case "/applications/100/volumes":
			if r.Method == http.MethodPost {
				// Return a volume with storage_class
				response := `{
					"success": true,
					"data": {
						"id": 1,
						"application_id": 100,
						"name": "test-volume",
						"size": 10,
						"path": "/var/lib/data",
						"storage_class": "fast-ssd",
						"resize_status": "completed"
					}
				}`
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte(response))
			}
		case "/applications/100/volumes/1":
			if r.Method == http.MethodGet {
				response := `{
					"success": true,
					"data": {
						"id": 1,
						"application_id": 100,
						"name": "test-volume",
						"size": 10,
						"path": "/var/lib/data",
						"storage_class": "fast-ssd",
						"resize_status": "completed"
					}
				}`
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(response))
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create client with test server
	c := client.NewClient("test-token", &server.URL)
	
	// Test volume creation with storage_class
	volume := &client.ApplicationVolume{
		ApplicationID: 100,
		Name:          "test-volume",
		Size:          10,
		MountPath:     "/var/lib/data",
		StorageClass:  "fast-ssd",
	}
	
	created, err := c.CreateVolume(volume)
	if err != nil {
		t.Fatalf("Failed to create volume: %v", err)
	}
	
	// Verify response includes storage_class
	if created.StorageClass != "fast-ssd" {
		t.Errorf("Expected StorageClass 'fast-ssd', got '%s'", created.StorageClass)
	}
}

// Mock client for testing without network calls
type MockVolumeClient struct {
	volumes map[int64]*client.ApplicationVolume
	nextID  int64
}

func NewMockVolumeClient() *MockVolumeClient {
	return &MockVolumeClient{
		volumes: make(map[int64]*client.ApplicationVolume),
		nextID:  1,
	}
}

func (m *MockVolumeClient) CreateVolume(volume *client.ApplicationVolume) (*client.ApplicationVolume, error) {
	volume.ID = m.nextID
	volume.ResizeStatus = "completed"
	volume.CreatedAt = time.Now()
	volume.UpdatedAt = time.Now()
	
	// If no storage class provided, use default
	if volume.StorageClass == "" {
		volume.StorageClass = "standard"
	}
	
	m.volumes[volume.ID] = volume
	m.nextID++
	
	return volume, nil
}

func (m *MockVolumeClient) GetVolume(appID, volumeID int64) (*client.ApplicationVolume, error) {
	volume, exists := m.volumes[volumeID]
	if !exists {
		return nil, fmt.Errorf("volume not found")
	}
	return volume, nil
}

func TestVolumeResource_StorageClass_CRUDOperations(t *testing.T) {
	mockClient := NewMockVolumeClient()
	resource := &VolumeResource{client: nil} // We'll mock the client methods
	
	// Test Create with storage_class
	data := &VolumeResourceModel{
		ApplicationID: types.Int64Value(100),
		Name:          types.StringValue("test-volume"),
		Size:          types.Int64Value(25),
		MountPath:     types.StringValue("/var/lib/test"),
		StorageClass:  types.StringValue("premium"),
	}
	
	apiModel := resource.toAPIModel(data)
	created, err := mockClient.CreateVolume(apiModel)
	if err != nil {
		t.Fatalf("Failed to create volume: %v", err)
	}
	
	// Verify creation includes storage_class
	if created.ID == 0 {
		t.Error("Expected volume to have an ID after creation")
	}
	if created.StorageClass != "premium" {
		t.Errorf("Expected StorageClass 'premium', got '%s'", created.StorageClass)
	}
	
	// Test Read
	retrieved, err := mockClient.GetVolume(100, created.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve volume: %v", err)
	}
	
	if retrieved.StorageClass != "premium" {
		t.Errorf("Expected StorageClass 'premium', got '%s'", retrieved.StorageClass)
	}
}

func TestVolumeResource_StorageClass_DefaultBehavior(t *testing.T) {
	mockClient := NewMockVolumeClient()
	resource := &VolumeResource{client: nil}
	
	// Test Create without storage_class (should get default from API)
	data := &VolumeResourceModel{
		ApplicationID: types.Int64Value(100),
		Name:          types.StringValue("default-volume"),
		Size:          types.Int64Value(10),
		MountPath:     types.StringValue("/var/lib/default"),
		StorageClass:  types.StringNull(),
	}
	
	apiModel := resource.toAPIModel(data)
	created, err := mockClient.CreateVolume(apiModel)
	if err != nil {
		t.Fatalf("Failed to create volume: %v", err)
	}
	
	// Mock client should set default storage class
	if created.StorageClass != "standard" {
		t.Errorf("Expected default StorageClass 'standard', got '%s'", created.StorageClass)
	}
}

func TestVolumeResource_StorageClass_ConversionAccuracy(t *testing.T) {
	resource := &VolumeResource{}
	
	// Test round-trip conversion (terraform -> api -> terraform)
	originalData := &VolumeResourceModel{
		ApplicationID: types.Int64Value(100),
		Name:          types.StringValue("conversion-test"),
		Size:          types.Int64Value(15),
		MountPath:     types.StringValue("/var/lib/test"),
		StorageClass:  types.StringValue("nvme-ultra"),
	}
	
	// Convert to API model
	apiModel := resource.toAPIModel(originalData)
	
	// Convert back from API model (simulate API response)
	apiVolume := &client.ApplicationVolume{
		ID:            apiModel.ID,
		ApplicationID: apiModel.ApplicationID,
		Name:          apiModel.Name,
		Size:          apiModel.Size,
		MountPath:     apiModel.MountPath,
		StorageClass:  apiModel.StorageClass,
		ResizeStatus:  "completed",
	}
	
	var convertedData VolumeResourceModel
	resource.fromAPIModel(apiVolume, &convertedData)
	
	// Verify round-trip accuracy
	if !convertedData.StorageClass.Equal(originalData.StorageClass) {
		t.Errorf("Round-trip conversion failed: expected %v, got %v", 
			originalData.StorageClass, convertedData.StorageClass)
	}
}