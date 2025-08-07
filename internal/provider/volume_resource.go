package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ploi/terraform-provider-ploicloud/internal/client"
)

var _ resource.Resource = &VolumeResource{}
var _ resource.ResourceWithImportState = &VolumeResource{}

func NewVolumeResource() resource.Resource {
	return &VolumeResource{}
}

type VolumeResource struct {
	client *client.Client
}

type VolumeResourceModel struct {
	ID            types.Int64  `tfsdk:"id"`
	ApplicationID types.Int64  `tfsdk:"application_id"`
	Name          types.String `tfsdk:"name"`
	Size          types.Int64  `tfsdk:"size"`
	MountPath     types.String `tfsdk:"mount_path"`
	StorageClass  types.String `tfsdk:"storage_class"`
	ResizeStatus  types.String `tfsdk:"resize_status"`
}

func (r *VolumeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume"
}

func (r *VolumeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a persistent volume for a Ploi Cloud application",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Volume ID",
			},
			"application_id": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Application ID this volume belongs to",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Volume name",
			},
			"size": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Volume size in GB",
			},
			"mount_path": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Path where the volume is mounted in the container",
			},
			"storage_class": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Storage class for the volume",
			},
			"resize_status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Volume resize status",
			},
		},
	}
}

func (r *VolumeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *VolumeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VolumeResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	volume := r.toAPIModel(&data)

	created, err := r.client.CreateVolume(volume)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create volume, got error: %s", err))
		return
	}

	r.fromAPIModel(created, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VolumeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VolumeResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	volume, err := r.client.GetVolume(data.ApplicationID.ValueInt64(), data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read volume, got error: %s", err))
		return
	}

	if volume == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	r.fromAPIModel(volume, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VolumeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VolumeResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	volume := r.toAPIModel(&data)

	updated, err := r.client.UpdateVolume(data.ApplicationID.ValueInt64(), data.ID.ValueInt64(), volume)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update volume, got error: %s", err))
		return
	}

	r.fromAPIModel(updated, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VolumeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VolumeResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteVolume(data.ApplicationID.ValueInt64(), data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete volume, got error: %s", err))
		return
	}
}

func (r *VolumeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ".")
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid Import ID", "Import ID must be in the format 'application_id.volume_id'")
		return
	}

	applicationID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", "Application ID must be a valid integer")
		return
	}

	volumeID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", "Volume ID must be a valid integer")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("application_id"), applicationID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), volumeID)...)
}

func (r *VolumeResource) toAPIModel(data *VolumeResourceModel) *client.ApplicationVolume {
	volume := &client.ApplicationVolume{
		ApplicationID: data.ApplicationID.ValueInt64(),
		Name:          data.Name.ValueString(),
		Size:          data.Size.ValueInt64(),
		MountPath:     data.MountPath.ValueString(),
	}

	if !data.ID.IsNull() {
		volume.ID = data.ID.ValueInt64()
	}
	
	if !data.StorageClass.IsNull() && data.StorageClass.ValueString() != "" {
		volume.StorageClass = data.StorageClass.ValueString()
	}

	return volume
}

func (r *VolumeResource) fromAPIModel(volume *client.ApplicationVolume, data *VolumeResourceModel) {
	data.ID = types.Int64Value(volume.ID)
	data.ApplicationID = types.Int64Value(volume.ApplicationID)
	data.Name = types.StringValue(volume.Name)
	data.Size = types.Int64Value(volume.Size)
	data.MountPath = types.StringValue(volume.MountPath)
	data.StorageClass = types.StringValue(volume.StorageClass)
	data.ResizeStatus = types.StringValue(volume.ResizeStatus)
}