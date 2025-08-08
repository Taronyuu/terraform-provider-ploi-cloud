package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ploi/terraform-provider-ploicloud/internal/client"
)

var _ resource.Resource = &WorkerResource{}
var _ resource.ResourceWithImportState = &WorkerResource{}

func NewWorkerResource() resource.Resource {
	return &WorkerResource{}
}

type WorkerResource struct {
	client *client.Client
}

type WorkerResourceModel struct {
	ID            types.Int64  `tfsdk:"id"`
	ApplicationID types.Int64  `tfsdk:"application_id"`
	Name          types.String `tfsdk:"name"`
	Command       types.String `tfsdk:"command"`
	Type          types.String `tfsdk:"type"`
	Replicas      types.Int64  `tfsdk:"replicas"`
	MemoryRequest types.String `tfsdk:"memory_request"`
	CPURequest    types.String `tfsdk:"cpu_request"`
	Status        types.String `tfsdk:"status"`
}

func (r *WorkerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_worker"
}

func (r *WorkerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		DeprecationMessage:  "Worker resources are deprecated. Use ploi-cloud_service with type 'worker' instead. See migration guide: https://docs.ploi.io/terraform-provider/migration/workers",
		MarkdownDescription: "**DEPRECATED**: Worker resources are deprecated. Use `ploi-cloud_service` with `type = \"worker\"` instead.\n\nThis resource is deprecated because workers are now handled as services in the Ploi Cloud API. Please migrate your configurations to use the service resource instead.\n\n## Migration Guide\n\nReplace your worker resource:\n\n```hcl\nresource \"ploicloud_worker\" \"queue\" {\n  application_id = ploicloud_application.app.id\n  name           = \"queue-worker\"\n  command        = \"php artisan queue:work\"\n  replicas       = 2\n}\n```\n\nWith a service resource:\n\n```hcl\nresource \"ploicloud_service\" \"queue\" {\n  application_id = ploicloud_application.app.id\n  service_name   = \"queue-worker\"\n  type           = \"worker\"\n  command        = \"php artisan queue:work\"\n  replicas       = 2\n}\n```",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Worker ID",
			},
			"application_id": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Application ID this worker belongs to",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Worker name",
			},
			"command": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Command to run for this worker (e.g., php artisan queue:work)",
			},
			"type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("queue"),
				MarkdownDescription: "Worker type (defaults to 'queue')",
			},
			"replicas": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1),
				MarkdownDescription: "Number of worker replicas",
			},
			"memory_request": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Memory request for the worker (e.g., '256Mi', '1Gi')",
			},
			"cpu_request": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "CPU request for the worker (e.g., '250m', '1')",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Worker status",
			},
		},
	}
}

func (r *WorkerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *WorkerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data WorkerResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Prevent creation of new worker resources and provide migration guidance
	resp.Diagnostics.AddError(
		"Worker Resource Deprecated",
		fmt.Sprintf("Worker resources are deprecated and cannot be created. Please use the 'ploicloud_service' resource instead.\n\n"+
			"Migration Guide:\n"+
			"Replace this worker configuration:\n\n"+
			"resource \"ploicloud_worker\" \"example\" {\n"+
			"  application_id = %d\n"+
			"  name           = \"%s\"\n"+
			"  command        = \"%s\"\n"+
			"  replicas       = %d\n"+
			"}\n\n"+
			"With this service configuration:\n\n"+
			"resource \"ploicloud_service\" \"example\" {\n"+
			"  application_id = %d\n"+
			"  service_name   = \"%s\"\n"+
			"  type           = \"worker\"\n"+
			"  command        = \"%s\"\n"+
			"  replicas       = %d\n"+
			"}\n\n"+
			"Documentation: https://docs.ploi.io/terraform-provider/migration/workers",
			data.ApplicationID.ValueInt64(),
			data.Name.ValueString(),
			data.Command.ValueString(),
			data.Replicas.ValueInt64(),
			data.ApplicationID.ValueInt64(),
			data.Name.ValueString(),
			data.Command.ValueString(),
			data.Replicas.ValueInt64()),
	)
}

func (r *WorkerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data WorkerResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	worker, err := r.client.GetWorker(data.ApplicationID.ValueInt64(), data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read worker, got error: %s", err))
		return
	}

	if worker == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	r.fromAPIModel(worker, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data WorkerResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	worker := r.toAPIModel(&data)

	updated, err := r.client.UpdateWorker(data.ApplicationID.ValueInt64(), data.ID.ValueInt64(), worker)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update worker, got error: %s", err))
		return
	}

	r.fromAPIModel(updated, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data WorkerResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteWorker(data.ApplicationID.ValueInt64(), data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete worker, got error: %s", err))
		return
	}
}

func (r *WorkerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ".")
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid Import ID", "Import ID must be in the format 'application_id.worker_id'")
		return
	}

	applicationID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", "Application ID must be a valid integer")
		return
	}

	workerID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", "Worker ID must be a valid integer")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("application_id"), applicationID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), workerID)...)
}

func (r *WorkerResource) toAPIModel(data *WorkerResourceModel) *client.Worker {
	worker := &client.Worker{
		ApplicationID: data.ApplicationID.ValueInt64(),
		Name:          data.Name.ValueString(),
		Command:       data.Command.ValueString(),
		Replicas:      data.Replicas.ValueInt64(),
	}

	if !data.ID.IsNull() {
		worker.ID = data.ID.ValueInt64()
	}
	
	if !data.Type.IsNull() && data.Type.ValueString() != "" {
		worker.Type = data.Type.ValueString()
	}
	
	if !data.MemoryRequest.IsNull() && data.MemoryRequest.ValueString() != "" {
		worker.MemoryRequest = data.MemoryRequest.ValueString()
	}
	
	if !data.CPURequest.IsNull() && data.CPURequest.ValueString() != "" {
		worker.CPURequest = data.CPURequest.ValueString()
	}

	return worker
}

func (r *WorkerResource) fromAPIModel(worker *client.Worker, data *WorkerResourceModel) {
	data.ID = types.Int64Value(worker.ID)
	data.ApplicationID = types.Int64Value(worker.ApplicationID)
	data.Name = types.StringValue(worker.Name)
	data.Command = types.StringValue(worker.Command)
	data.Type = types.StringValue(worker.Type)
	data.Replicas = types.Int64Value(worker.Replicas)
	data.MemoryRequest = types.StringValue(worker.MemoryRequest)
	data.CPURequest = types.StringValue(worker.CPURequest)
	data.Status = types.StringValue(worker.Status)
}