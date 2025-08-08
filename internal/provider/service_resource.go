package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ploi/terraform-provider-ploicloud/internal/client"
)

var _ resource.Resource = &ServiceResource{}
var _ resource.ResourceWithImportState = &ServiceResource{}

func NewServiceResource() resource.Resource {
	return &ServiceResource{}
}

type ServiceResource struct {
	client *client.Client
}

type ServiceResourceModel struct {
	ID            types.Int64  `tfsdk:"id"`
	ApplicationID types.Int64  `tfsdk:"application_id"`
	Name          types.String `tfsdk:"service_name"`
	Type          types.String `tfsdk:"type"`
	Version       types.String `tfsdk:"version"`
	Settings      types.Map    `tfsdk:"settings"`
	Replicas      types.Int64  `tfsdk:"replicas"`
	CPURequest    types.String `tfsdk:"cpu_request"`
	MemoryRequest types.String `tfsdk:"memory_request"`
	StorageSize   types.String `tfsdk:"storage_size"`
	Extensions    types.List   `tfsdk:"extensions"`
	Command       types.String `tfsdk:"command"`
	Status        types.String `tfsdk:"status"`
}

func (r *ServiceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service"
}

func (r *ServiceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Ploi Cloud application service (database, cache, etc.)",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Service ID",
			},
			"application_id": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Application ID this service belongs to",
			},
			"service_name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Service name (auto-generated if not provided)",
			},
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Service type (mysql, postgresql, redis, valkey, rabbitmq, mongodb, minio, sftp, worker)",
				Validators: []validator.String{
					stringvalidator.OneOf("mysql", "postgresql", "redis", "valkey", "rabbitmq", "mongodb", "minio", "sftp", "worker"),
				},
			},
			"version": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Service version",
			},
			"settings": schema.MapAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Service-specific settings (can be configured, auto-generated values will be preserved)",
			},
			"replicas": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Number of replicas for the service (applicable to worker-type services)",
			},
			"cpu_request": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "CPU request for the service (e.g., '250m', '1')",
			},
			"memory_request": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Memory request for the service (e.g., '256Mi', '1Gi')",
			},
			"storage_size": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Storage size for the service (e.g., '1Gi', '10Gi')",
			},
			"extensions": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Extensions for PostgreSQL services (e.g., ['uuid-ossp', 'pgcrypto'])",
			},
			"command": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Command to run for worker services (e.g., 'php artisan queue:work'). Only applicable to worker type services.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Service status",
			},
		},
	}
}

func (r *ServiceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *ServiceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ServiceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	service := r.toAPIModel(&data)

	created, err := r.client.CreateService(service)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create service, got error: %s", err))
		return
	}

	// Ensure ApplicationID is preserved from the request
	created.ApplicationID = service.ApplicationID
	r.fromAPIModel(created, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ServiceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	service, err := r.client.GetService(data.ApplicationID.ValueInt64(), data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read service, got error: %s", err))
		return
	}

	if service == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	r.fromAPIModel(service, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ServiceResourceModel
	var state ServiceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Copy IDs from state to the plan data
	data.ID = state.ID
	data.ApplicationID = state.ApplicationID

	// Convert to API model and update
	service := r.toAPIModel(&data)
	
	updated, err := r.client.UpdateService(data.ApplicationID.ValueInt64(), data.ID.ValueInt64(), service)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update service, got error: %s", err))
		return
	}

	// Ensure ApplicationID is preserved from the request
	updated.ApplicationID = service.ApplicationID
	r.fromAPIModel(updated, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ServiceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteService(data.ApplicationID.ValueInt64(), data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete service, got error: %s", err))
		return
	}
}

func (r *ServiceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ".")
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid Import ID", "Import ID must be in the format 'application_id.service_id'")
		return
	}

	applicationID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", "Application ID must be a valid integer")
		return
	}

	serviceID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", "Service ID must be a valid integer")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("application_id"), applicationID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), serviceID)...)
}

func (r *ServiceResource) toAPIModel(data *ServiceResourceModel) *client.ApplicationService {
	service := &client.ApplicationService{
		ApplicationID: data.ApplicationID.ValueInt64(),
		Type:          data.Type.ValueString(),
		Version:       data.Version.ValueString(),
	}

	if !data.ID.IsNull() {
		service.ID = data.ID.ValueInt64()
	}

	if !data.Name.IsNull() && data.Name.ValueString() != "" {
		service.Name = data.Name.ValueString()
	}

	// Include settings if provided in configuration
	if !data.Settings.IsNull() {
		settingsMap := make(map[string]string)
		data.Settings.ElementsAs(context.Background(), &settingsMap, false)
		service.Settings = client.FlexibleSettingsFromMap(settingsMap)
	}

	// Include resource configuration if provided
	if !data.Replicas.IsNull() {
		service.Replicas = data.Replicas.ValueInt64()
	}
	
	if !data.CPURequest.IsNull() && data.CPURequest.ValueString() != "" {
		service.CPURequest = data.CPURequest.ValueString()
	}
	
	if !data.MemoryRequest.IsNull() && data.MemoryRequest.ValueString() != "" {
		service.MemoryRequest = data.MemoryRequest.ValueString()
	}
	
	if !data.StorageSize.IsNull() && data.StorageSize.ValueString() != "" {
		service.StorageSize = data.StorageSize.ValueString()
	}
	
	// Handle extensions list for PostgreSQL services
	if !data.Extensions.IsNull() {
		var extensions []string
		data.Extensions.ElementsAs(context.Background(), &extensions, false)
		service.Extensions = extensions
	}

	// Handle command for worker services
	if !data.Command.IsNull() && data.Command.ValueString() != "" {
		service.Command = data.Command.ValueString()
	}

	return service
}

func (r *ServiceResource) fromAPIModel(service *client.ApplicationService, data *ServiceResourceModel) {
	data.ID = types.Int64Value(service.ID)
	data.ApplicationID = types.Int64Value(service.ApplicationID)
	data.Name = types.StringValue(service.Name)
	data.Type = types.StringValue(service.Type)
	// Handle Version field - convert empty string to null, non-empty to string value
	if service.Version != "" {
		data.Version = types.StringValue(service.Version)
	} else {
		data.Version = types.StringNull()
	}
	data.Status = types.StringValue(service.Status)

	// Set resource configuration
	data.Replicas = types.Int64Value(service.Replicas)
	data.CPURequest = types.StringValue(service.CPURequest)
	data.MemoryRequest = types.StringValue(service.MemoryRequest)
	data.StorageSize = types.StringValue(service.StorageSize)

	// Handle extensions list
	if len(service.Extensions) > 0 {
		extensions := make([]types.String, len(service.Extensions))
		for i, ext := range service.Extensions {
			extensions[i] = types.StringValue(ext)
		}
		data.Extensions, _ = types.ListValueFrom(context.Background(), types.StringType, extensions)
	} else {
		data.Extensions = types.ListNull(types.StringType)
	}

	// Handle command field
	if service.Command != "" {
		data.Command = types.StringValue(service.Command)
	} else {
		data.Command = types.StringNull()
	}

	if len(service.Settings) > 0 {
		settingsMap := make(map[string]types.String)
		for k, v := range service.Settings.ToMap() {
			settingsMap[k] = types.StringValue(v)
		}
		data.Settings, _ = types.MapValueFrom(context.Background(), types.StringType, settingsMap)
	}
}