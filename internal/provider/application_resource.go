package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ploi/terraform-provider-ploicloud/internal/client"
)

var _ resource.Resource = &ApplicationResource{}
var _ resource.ResourceWithImportState = &ApplicationResource{}

func NewApplicationResource() resource.Resource {
	return &ApplicationResource{}
}

type ApplicationResource struct {
	client *client.Client
}

type ApplicationResourceModel struct {
	ID                 types.Int64    `tfsdk:"id"`
	Name               types.String   `tfsdk:"name"`
	Type               types.String   `tfsdk:"type"`
	ApplicationVersion types.String   `tfsdk:"application_version"`
	Runtime            *RuntimeModel  `tfsdk:"runtime"`
	BuildCommands      types.List     `tfsdk:"build_commands"`
	InitCommands       types.List     `tfsdk:"init_commands"`
	Settings           *SettingsModel `tfsdk:"settings"`
	PHPExtensions      types.List     `tfsdk:"php_extensions"`
	PHPSettings        types.List     `tfsdk:"php_settings"`
	URL                types.String   `tfsdk:"url"`
	Status             types.String   `tfsdk:"status"`
	NeedsDeployment    types.Bool     `tfsdk:"needs_deployment"`
	CustomManifests    types.String   `tfsdk:"custom_manifests"`
	RepositoryURL      types.String   `tfsdk:"repository_url"`
	RepositoryOwner    types.String   `tfsdk:"repository_owner"`
	RepositoryName     types.String   `tfsdk:"repository_name"`
	DefaultBranch      types.String   `tfsdk:"default_branch"`
	SocialAccountID    types.Int64    `tfsdk:"social_account_id"`
	Region             types.String   `tfsdk:"region"`
	CloudProvider      types.String   `tfsdk:"cloud_provider"`
}

type RuntimeModel struct {
	PHPVersion    types.String `tfsdk:"php_version"`
	NodeJSVersion types.String `tfsdk:"nodejs_version"`
}

type SettingsModel struct {
	HealthCheckPath  types.String `tfsdk:"health_check_path"`
	SchedulerEnabled types.Bool   `tfsdk:"scheduler_enabled"`
	Replicas         types.Int64  `tfsdk:"replicas"`
	CPURequest       types.String `tfsdk:"cpu_request"`
	MemoryRequest    types.String `tfsdk:"memory_request"`
}

func (r *ApplicationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

func (r *ApplicationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Ploi Cloud application",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Application ID",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Application name",
			},
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Application type (laravel, wordpress, statamic, craftcms, nodejs)",
				Validators: []validator.String{
					stringvalidator.OneOf("laravel", "wordpress", "statamic", "craftcms", "nodejs"),
				},
			},
			"application_version": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Application version (e.g., 11.x for Laravel)",
			},
			"build_commands": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Build commands to run during image build",
			},
			"init_commands": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Initialization commands to run before starting the application",
			},
			"php_extensions": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "PHP extensions to install",
			},
			"php_settings": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "PHP ini settings",
			},
			"url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Application URL",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Application status",
			},
			"needs_deployment": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the application needs deployment",
			},
			"custom_manifests": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Custom Kubernetes manifests in YAML format",
			},
			"repository_url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Repository URL",
			},
			"repository_owner": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Repository owner",
			},
			"repository_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Repository name",
			},
			"default_branch": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("main"),
				MarkdownDescription: "Default git branch",
			},
			"social_account_id": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Social account ID for git integration",
			},
			"region": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("default"),
				MarkdownDescription: "Region to deploy the application",
			},
			"cloud_provider": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("default"),
				MarkdownDescription: "Cloud provider",
			},
		},

		Blocks: map[string]schema.Block{
			"runtime": schema.SingleNestedBlock{
				MarkdownDescription: "Runtime configuration",
				Attributes: map[string]schema.Attribute{
					"php_version": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "PHP version (7.4, 8.0, 8.1, 8.2, 8.3, 8.4)",
						Validators: []validator.String{
							stringvalidator.OneOf("7.4", "8.0", "8.1", "8.2", "8.3", "8.4"),
						},
					},
					"nodejs_version": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Node.js version (18, 20, 22, 24)",
						Validators: []validator.String{
							stringvalidator.OneOf("18", "20", "22", "24"),
						},
					},
				},
			},
			"settings": schema.SingleNestedBlock{
				MarkdownDescription: "Application settings",
				Attributes: map[string]schema.Attribute{
					"health_check_path": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString("/"),
						MarkdownDescription: "Health check path",
					},
					"scheduler_enabled": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
						MarkdownDescription: "Enable Laravel scheduler",
					},
					"replicas": schema.Int64Attribute{
						Optional:            true,
						Computed:            true,
						Default:             int64default.StaticInt64(1),
						MarkdownDescription: "Number of replicas",
					},
					"cpu_request": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString("250m"),
						MarkdownDescription: "CPU request",
					},
					"memory_request": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString("512Mi"),
						MarkdownDescription: "Memory request",
					},
				},
			},
		},
	}
}

func (r *ApplicationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ApplicationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ApplicationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	app := r.toAPIModel(&data)

	created, err := r.client.CreateApplication(app)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create application, got error: %s", err))
		return
	}

	r.fromAPIModel(created, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ApplicationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ApplicationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	app, err := r.client.GetApplication(data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read application, got error: %s", err))
		return
	}

	if app == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	r.fromAPIModel(app, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ApplicationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ApplicationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	app := r.toAPIModel(&data)

	updated, err := r.client.UpdateApplication(data.ID.ValueInt64(), app)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update application, got error: %s", err))
		return
	}

	r.fromAPIModel(updated, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ApplicationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ApplicationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteApplication(data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete application, got error: %s", err))
		return
	}
}

func (r *ApplicationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", "Import ID must be a valid integer")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

func (r *ApplicationResource) toAPIModel(data *ApplicationResourceModel) *client.Application {
	app := &client.Application{
		Name:               data.Name.ValueString(),
		Type:               data.Type.ValueString(),
		ApplicationVersion: data.ApplicationVersion.ValueString(),
		CustomManifests:    data.CustomManifests.ValueString(),
		RepositoryURL:      data.RepositoryURL.ValueString(),
		RepositoryOwner:    data.RepositoryOwner.ValueString(),
		RepositoryName:     data.RepositoryName.ValueString(),
		DefaultBranch:      data.DefaultBranch.ValueString(),
		Region:             data.Region.ValueString(),
		Provider:           data.CloudProvider.ValueString(),
	}

	if !data.ID.IsNull() {
		app.ID = data.ID.ValueInt64()
	}

	if !data.SocialAccountID.IsNull() {
		app.SocialAccountID = data.SocialAccountID.ValueInt64()
	}

	if data.Runtime != nil {
		if !data.Runtime.PHPVersion.IsNull() {
			app.PHPVersion = data.Runtime.PHPVersion.ValueString()
		}
		if !data.Runtime.NodeJSVersion.IsNull() {
			app.NodeJSVersion = data.Runtime.NodeJSVersion.ValueString()
		}
	}

	if data.Settings != nil {
		if !data.Settings.HealthCheckPath.IsNull() {
			app.HealthCheckPath = data.Settings.HealthCheckPath.ValueString()
		}
		if !data.Settings.SchedulerEnabled.IsNull() {
			app.SchedulerEnabled = data.Settings.SchedulerEnabled.ValueBool()
		}
		if !data.Settings.Replicas.IsNull() {
			app.Replicas = data.Settings.Replicas.ValueInt64()
		}
		if !data.Settings.CPURequest.IsNull() {
			app.CPURequest = data.Settings.CPURequest.ValueString()
		}
		if !data.Settings.MemoryRequest.IsNull() {
			app.MemoryRequest = data.Settings.MemoryRequest.ValueString()
		}
	}

	if !data.BuildCommands.IsNull() {
		elements := make([]types.String, 0, len(data.BuildCommands.Elements()))
		data.BuildCommands.ElementsAs(context.Background(), &elements, false)
		for _, elem := range elements {
			app.BuildCommands = append(app.BuildCommands, elem.ValueString())
		}
	}

	if !data.InitCommands.IsNull() {
		elements := make([]types.String, 0, len(data.InitCommands.Elements()))
		data.InitCommands.ElementsAs(context.Background(), &elements, false)
		for _, elem := range elements {
			app.InitCommands = append(app.InitCommands, elem.ValueString())
		}
	}

	if !data.PHPExtensions.IsNull() {
		elements := make([]types.String, 0, len(data.PHPExtensions.Elements()))
		data.PHPExtensions.ElementsAs(context.Background(), &elements, false)
		for _, elem := range elements {
			app.PHPExtensions = append(app.PHPExtensions, elem.ValueString())
		}
	}

	if !data.PHPSettings.IsNull() {
		elements := make([]types.String, 0, len(data.PHPSettings.Elements()))
		data.PHPSettings.ElementsAs(context.Background(), &elements, false)
		for _, elem := range elements {
			app.PHPSettings = append(app.PHPSettings, elem.ValueString())
		}
	}

	return app
}

func (r *ApplicationResource) fromAPIModel(app *client.Application, data *ApplicationResourceModel) {
	data.ID = types.Int64Value(app.ID)
	data.Name = types.StringValue(app.Name)
	data.Type = types.StringValue(app.Type)
	data.ApplicationVersion = types.StringValue(app.ApplicationVersion)
	data.URL = types.StringValue(app.URL)
	data.Status = types.StringValue(app.Status)
	data.NeedsDeployment = types.BoolValue(app.NeedsDeployment)
	data.CustomManifests = types.StringValue(app.CustomManifests)
	data.RepositoryURL = types.StringValue(app.RepositoryURL)
	data.RepositoryOwner = types.StringValue(app.RepositoryOwner)
	data.RepositoryName = types.StringValue(app.RepositoryName)
	data.DefaultBranch = types.StringValue(app.DefaultBranch)
	data.Region = types.StringValue(app.Region)
	data.CloudProvider = types.StringValue(app.Provider)

	if app.SocialAccountID != 0 {
		data.SocialAccountID = types.Int64Value(app.SocialAccountID)
	}

	if data.Runtime == nil {
		data.Runtime = &RuntimeModel{}
	}
	data.Runtime.PHPVersion = types.StringValue(app.PHPVersion)
	data.Runtime.NodeJSVersion = types.StringValue(app.NodeJSVersion)

	if data.Settings == nil {
		data.Settings = &SettingsModel{}
	}
	data.Settings.HealthCheckPath = types.StringValue(app.HealthCheckPath)
	data.Settings.SchedulerEnabled = types.BoolValue(app.SchedulerEnabled)
	data.Settings.Replicas = types.Int64Value(app.Replicas)
	data.Settings.CPURequest = types.StringValue(app.CPURequest)
	data.Settings.MemoryRequest = types.StringValue(app.MemoryRequest)

	if len(app.BuildCommands) > 0 {
		elements := make([]types.String, len(app.BuildCommands))
		for i, cmd := range app.BuildCommands {
			elements[i] = types.StringValue(cmd)
		}
		data.BuildCommands, _ = types.ListValueFrom(context.Background(), types.StringType, elements)
	}

	if len(app.InitCommands) > 0 {
		elements := make([]types.String, len(app.InitCommands))
		for i, cmd := range app.InitCommands {
			elements[i] = types.StringValue(cmd)
		}
		data.InitCommands, _ = types.ListValueFrom(context.Background(), types.StringType, elements)
	}

	if len(app.PHPExtensions) > 0 {
		elements := make([]types.String, len(app.PHPExtensions))
		for i, ext := range app.PHPExtensions {
			elements[i] = types.StringValue(ext)
		}
		data.PHPExtensions, _ = types.ListValueFrom(context.Background(), types.StringType, elements)
	}

	if len(app.PHPSettings) > 0 {
		elements := make([]types.String, len(app.PHPSettings))
		for i, setting := range app.PHPSettings {
			elements[i] = types.StringValue(setting)
		}
		data.PHPSettings, _ = types.ListValueFrom(context.Background(), types.StringType, elements)
	}
}