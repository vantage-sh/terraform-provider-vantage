package vantage

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ provider.Provider = &vantageProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New() provider.Provider {
	return &vantageProvider{}
}

// vantageProvider is the provider implementation.
type vantageProvider struct{}

type vantageProviderModel struct {
	Host     types.String `tfsdk:"host"`
	APIToken types.String `tfsdk:"api_token"`
}

// Metadata returns the provider type name.
func (p *vantageProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "vantage"
}

// Schema defines the provider-level schema for configuration data.
func (p *vantageProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Optional: true,
			},
			"api_token": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}

// Configure prepares a vantage API client for data sources and resources.
func (p *vantageProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config vantageProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown Vantage API Host",
			"The provider cannot create the Vantage API client as there is an unknown configuration value for the Vantage API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the VANTAGE_HOST environment variable.",
		)
	}

	if config.APIToken.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Unknown Vantage API Token",
			"The provider cannot create the Vantage API client as there is an unknown configuration value for the Vantage API token. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the VANTAGE_API_TOKEN environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	host := os.Getenv("VANTAGE_HOST")
	apiToken := os.Getenv("VANTAGE_API_TOKEN")

	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	if !config.APIToken.IsNull() {
		apiToken = config.APIToken.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if host == "" {
		host = "https://api.vantage.sh"
	}

	if apiToken == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Missing Vantage API Token",
			"The provider cannot create the Vantage API client as there is a missing or empty value for the Vantage API token. "+
				"Set the Token value in the configuration or use the VANTAGE_API_TOKEN environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := NewClient(host, apiToken)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Failed to build API Client",
			"The provider cannot create the Vantage API client, likely due to an error parsing the host configuration.  "+err.Error(),
		)

		return
	}
	// Make the Vantage client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client
}

// DataSources defines the data sources implemented in the provider.
func (p *vantageProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewAwsProviderInfoDataSource,
		NewSavedFiltersDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *vantageProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAwsProviderResource,
		NewFolderResource,
		NewSavedFilterResource,
		NewCostReportResource,
		NewDashboardResource,
		NewSegmentResource,
		NewTeamResource,
	}
}
