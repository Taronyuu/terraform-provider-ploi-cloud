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

var _ resource.Resource = &SecretResource{}
var _ resource.ResourceWithImportState = &SecretResource{}

func NewSecretResource() resource.Resource {
	return &SecretResource{}
}

type SecretResource struct {
	client *client.Client
}

type SecretResourceModel struct {
	ApplicationID types.Int64  `tfsdk:"application_id"`
	Key           types.String `tfsdk:"key"`
	Value         types.String `tfsdk:"value"`
}

func (r *SecretResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret"
}

func (r *SecretResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an environment variable secret for a Ploi Cloud application",

		Attributes: map[string]schema.Attribute{
			"application_id": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Application ID this secret belongs to",
			},
			"key": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Environment variable key (must be uppercase with underscores)",
			},
			"value": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "Environment variable value",
			},
		},
	}
}

func (r *SecretResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SecretResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SecretResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	secret := r.toAPIModel(&data)

	// Try to create the secret first
	created, err := r.client.CreateSecret(secret)
	if err != nil {
		// If creation failed due to existing secret, try to update it instead
		if strings.Contains(err.Error(), "already exists") {
			updated, updateErr := r.client.UpdateSecret(data.ApplicationID.ValueInt64(), data.Key.ValueString(), secret)
			if updateErr != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create or update secret, create error: %s, update error: %s", err, updateErr))
				return
			}
			r.fromAPIModel(updated, &data)
		} else {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create secret, got error: %s", err))
			return
		}
	} else {
		r.fromAPIModel(created, &data)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SecretResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SecretResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	secret, err := r.client.GetSecret(data.ApplicationID.ValueInt64(), data.Key.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read secret, got error: %s", err))
		return
	}

	if secret == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	r.fromAPIModel(secret, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SecretResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SecretResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	secret := r.toAPIModel(&data)

	updated, err := r.client.UpdateSecret(data.ApplicationID.ValueInt64(), data.Key.ValueString(), secret)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update secret, got error: %s", err))
		return
	}

	r.fromAPIModel(updated, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SecretResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SecretResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteSecret(data.ApplicationID.ValueInt64(), data.Key.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete secret, got error: %s", err))
		return
	}
}

func (r *SecretResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ".")
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid Import ID", "Import ID must be in the format 'application_id.secret_key'")
		return
	}

	applicationID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", "Application ID must be a valid integer")
		return
	}

	secretKey := parts[1]

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("application_id"), applicationID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("key"), secretKey)...)
}

func (r *SecretResource) toAPIModel(data *SecretResourceModel) *client.ApplicationSecret {
	return &client.ApplicationSecret{
		ApplicationID: data.ApplicationID.ValueInt64(),
		Key:           data.Key.ValueString(),
		Value:         data.Value.ValueString(),
	}
}

func (r *SecretResource) fromAPIModel(secret *client.ApplicationSecret, data *SecretResourceModel) {
	// Only update ApplicationID if it's not zero, otherwise preserve the planned value
	if secret.ApplicationID != 0 {
		data.ApplicationID = types.Int64Value(secret.ApplicationID)
	}
	
	data.Key = types.StringValue(secret.Key)
	
	// Don't update the value if API returns masked value "********"
	// The API masks secret values for security, so preserve the original planned value
	if secret.Value != "" && secret.Value != "********" {
		data.Value = types.StringValue(secret.Value)
	}
}