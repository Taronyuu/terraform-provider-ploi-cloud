package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ploi/terraform-provider-ploicloud/internal/client"
)

var _ provider.Provider = &PloiCloudProvider{}

type PloiCloudProvider struct {
	version string
}

type PloiCloudProviderModel struct {
	ApiToken    types.String `tfsdk:"api_token"`
	ApiEndpoint types.String `tfsdk:"api_endpoint"`
}

func (p *PloiCloudProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "ploicloud"
	resp.Version = p.version
}

func (p *PloiCloudProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The Ploi Cloud provider allows you to manage your Ploi Cloud applications and services using Terraform.",
		Attributes: map[string]schema.Attribute{
			"api_token": schema.StringAttribute{
				MarkdownDescription: "The API token for Ploi Cloud authentication. Can also be set with the PLOICLOUD_API_TOKEN environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"api_endpoint": schema.StringAttribute{
				MarkdownDescription: "The API endpoint for Ploi Cloud. Defaults to https://cloud.ploi.io/api/v1.",
				Optional:            true,
			},
		},
	}
}

func (p *PloiCloudProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config PloiCloudProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiToken := os.Getenv("PLOICLOUD_API_TOKEN")
	if !config.ApiToken.IsNull() {
		apiToken = config.ApiToken.ValueString()
	}

	if apiToken == "" {
		resp.Diagnostics.AddError(
			"Missing API Token Configuration",
			"While configuring the provider, the API token was not found in the "+
				"PLOICLOUD_API_TOKEN environment variable or provider configuration "+
				"block api_token attribute.",
		)
		return
	}

	apiEndpoint := config.ApiEndpoint.ValueStringPointer()

	client := client.NewClient(apiToken, apiEndpoint)

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *PloiCloudProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewApplicationResource,
		NewServiceResource,
		NewDomainResource,
		NewSecretResource,
		NewVolumeResource,
		NewWorkerResource,
	}
}

func (p *PloiCloudProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewApplicationDataSource,
		NewTeamDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &PloiCloudProvider{
			version: version,
		}
	}
}