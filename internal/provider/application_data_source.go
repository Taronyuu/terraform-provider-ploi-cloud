package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ploi/terraform-provider-ploicloud/internal/client"
)

var _ datasource.DataSource = &ApplicationDataSource{}

func NewApplicationDataSource() datasource.DataSource {
	return &ApplicationDataSource{}
}

type ApplicationDataSource struct {
	client *client.Client
}

type ApplicationDataSourceModel struct {
	ID                 types.Int64  `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Type               types.String `tfsdk:"type"`
	ApplicationVersion types.String `tfsdk:"application_version"`
	URL                types.String `tfsdk:"url"`
	Status             types.String `tfsdk:"status"`
	NeedsDeployment    types.Bool   `tfsdk:"needs_deployment"`
	Region             types.String `tfsdk:"region"`
	CloudProvider      types.String `tfsdk:"cloud_provider"`
}

func (d *ApplicationDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

func (d *ApplicationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Application data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Application identifier",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Application name",
			},
			"type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Application type",
			},
			"application_version": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Application version",
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
			"region": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Application region",
			},
			"cloud_provider": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Cloud provider",
			},
		},
	}
}

func (d *ApplicationDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *ApplicationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ApplicationDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	app, err := d.client.GetApplication(data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read application, got error: %s", err))
		return
	}

	if app == nil {
		resp.Diagnostics.AddError("Application Not Found", fmt.Sprintf("Application with ID %d not found", data.ID.ValueInt64()))
		return
	}

	data.ID = types.Int64Value(app.ID)
	data.Name = types.StringValue(app.Name)
	data.Type = types.StringValue(app.Type)
	data.ApplicationVersion = types.StringValue(app.ApplicationVersion)
	data.URL = types.StringValue(app.URL)
	data.Status = types.StringValue(app.Status)
	data.NeedsDeployment = types.BoolValue(app.NeedsDeployment)
	data.Region = types.StringValue(app.Region)
	data.CloudProvider = types.StringValue(app.Provider)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}