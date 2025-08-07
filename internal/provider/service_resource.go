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
	Type          types.String `tfsdk:"type"`
	Version       types.String `tfsdk:"version"`
	Settings      types.Map    `tfsdk:"settings"`
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
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Service type (mysql, postgresql, redis, valkey, rabbitmq, mongodb, minio, sftp)",
				Validators: []validator.String{
					stringvalidator.OneOf("mysql", "postgresql", "redis", "valkey", "rabbitmq", "mongodb", "minio", "sftp"),
				},
			},
			"version": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Service version",
			},
			"settings": schema.MapAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Service-specific settings (e.g., database name, storage size)",
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

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	service := r.toAPIModel(&data)

	updated, err := r.client.UpdateService(data.ApplicationID.ValueInt64(), data.ID.ValueInt64(), service)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update service, got error: %s", err))
		return
	}

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

	if !data.Settings.IsNull() {
		settings := make(map[string]string)
		data.Settings.ElementsAs(context.Background(), &settings, false)
		service.Settings = settings
	}

	return service
}

func (r *ServiceResource) fromAPIModel(service *client.ApplicationService, data *ServiceResourceModel) {
	data.ID = types.Int64Value(service.ID)
	data.ApplicationID = types.Int64Value(service.ApplicationID)
	data.Type = types.StringValue(service.Type)
	data.Version = types.StringValue(service.Version)
	data.Status = types.StringValue(service.Status)

	if len(service.Settings) > 0 {
		settingsMap := make(map[string]types.String)
		for k, v := range service.Settings {
			settingsMap[k] = types.StringValue(v)
		}
		data.Settings, _ = types.MapValueFrom(context.Background(), types.StringType, settingsMap)
	}
}