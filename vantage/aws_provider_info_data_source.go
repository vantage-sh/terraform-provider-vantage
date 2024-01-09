package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &awsProviderInfoDataSource{}
	_ datasource.DataSourceWithConfigure = &awsProviderInfoDataSource{}
)

func NewAwsProviderInfoDataSource() datasource.DataSource {
	return &awsProviderInfoDataSource{}
}

type awsProviderInfoDataSource struct {
	client *Client
}

type awsProviderInfoDataSourceModel struct {
	ExternalID                types.String `tfsdk:"external_id"`
	IamRoleARN                types.String `tfsdk:"iam_role_arn"`
	RootPolicy                types.String `tfsdk:"root_policy"`
	AutopilotPolicy           types.String `tfsdk:"autopilot_policy"`
	CloudwatchMetricsPolicy   types.String `tfsdk:"cloudwatch_metrics_policy"`
	AdditionalResourcesPolicy types.String `tfsdk:"additional_resources_policy"`
}

func (d *awsProviderInfoDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_aws_provider_info"
}

func (d *awsProviderInfoDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"iam_role_arn": schema.StringAttribute{
				Computed: true,
				Description: "The IAM role that Vantage assumes into your account.",
				MarkdownDescription: "The IAM role that Vantage assumes into your account.",
			},
			"external_id": schema.StringAttribute{
				Computed: true,
				Description: "The Vantage external ID to authenticate your account.",
				MarkdownDescription: "The Vantage external ID to authenticate your account.",
				Sensitive: true,
			},
			"root_policy": schema.StringAttribute{
				Computed: true,
				Description: "The policy that allows Vantage to acces billing information.",
				MarkdownDescription: "The policy that allows Vantage to manage autopilot.",
			},
			"autopilot_policy": schema.StringAttribute{
				Computed: true,
				Description: "The policy that allows Vantage to manage autopilot",
				MarkdownDescription: "The policy that allows Vantage to manage autopilot.",
			},
			"cloudwatch_metrics_policy": schema.StringAttribute{
				Computed: true,
				Description: "The policy that allows Vantage to retrieve cloudwatch metrics from your AWS account.",
				MarkdownDescription: "The policy that allows Vantage to retrieve cloudwatch metrics from your AWS account.",
			},
			"additional_resources_policy": schema.StringAttribute{
				Computed: true,
				Description: "The policy that allows Vantage to list and describe resources from your AWS account.",
				MarkdownDescription: "The policy that allows Vantage to list and describe resources from your AWS account.",
			},
		},
	}
}

func (d *awsProviderInfoDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state awsProviderInfoDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	out, err := d.client.V1.Integrations.GetIntegrationsAWSInfo(nil, d.client.Auth)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Vantage AwsProviderInfo",
			err.Error(),
		)
		return
	}

	if !out.IsSuccess() {
		resp.Diagnostics.AddError(
			"Unable to Vantage AwsProviderInfo",
			out.Error(),
		)
		return
	}

	state.ExternalID = types.StringValue(out.Payload.ExternalID)
	state.IamRoleARN = types.StringValue(out.Payload.IamRoleArn)
	state.RootPolicy = types.StringValue(out.Payload.Policies.Root)
	state.AutopilotPolicy = types.StringValue(out.Payload.Policies.Autopilot)
	state.CloudwatchMetricsPolicy = types.StringValue(out.Payload.Policies.Cloudwatch)
	state.AdditionalResourcesPolicy = types.StringValue(out.Payload.Policies.Resources)

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *awsProviderInfoDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*Client)
}
