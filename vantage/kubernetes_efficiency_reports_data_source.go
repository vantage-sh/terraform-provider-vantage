package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/datasource_kubernetes_efficiency_reports"
	kerv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/kubernetes_efficiency_reports"
)

var _ datasource.DataSource = (*kubernetesEfficiencyReportsDataSource)(nil)
var _ datasource.DataSourceWithConfigure = &kubernetesEfficiencyReportsDataSource{}

func NewKubernetesEfficiencyReportsDataSource() datasource.DataSource {
	return &kubernetesEfficiencyReportsDataSource{}
}

type kubernetesEfficiencyReportsDataSource struct {
	client *Client
}

// Configure implements datasource.DataSourceWithConfigure.
func (d *kubernetesEfficiencyReportsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*Client)
}

type kubernetesEfficiencyReportsDataSourceModel struct {
	KubernetesEfficiencyReports []kubernetesEfficiencyReportModel `tfsdk:"kubernetes_efficiency_reports"`
}

type kubernetesEfficiencyReportModel struct {
	AggregatedBy   types.String `tfsdk:"aggregated_by"`
	CreatedAt      types.String `tfsdk:"created_at"`
	DateBucket     types.String `tfsdk:"date_bucket"`
	DateInterval   types.String `tfsdk:"date_interval"`
	Default        types.Bool   `tfsdk:"default"`
	EndDate        types.String `tfsdk:"end_date"`
	Groupings      types.String `tfsdk:"groupings"`
	StartDate      types.String `tfsdk:"start_date"`
	Title          types.String `tfsdk:"title"`
	Token          types.String `tfsdk:"token"`
	UserToken      types.String `tfsdk:"user_token"`
	WorkspaceToken types.String `tfsdk:"workspace_token"`
}

func (d *kubernetesEfficiencyReportsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_efficiency_reports"
}

func (d *kubernetesEfficiencyReportsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_kubernetes_efficiency_reports.KubernetesEfficiencyReportsDataSourceSchema(ctx)
}

func (d *kubernetesEfficiencyReportsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data kubernetesEfficiencyReportsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	params := kerv2.NewGetKubernetesEfficiencyReportsParams()
	out, err := d.client.V2.KubernetesEfficiencyReports.GetKubernetesEfficiencyReports(params, d.client.Auth)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Get Vantage Kubernetes Efficiency Reports",
			err.Error(),
		)
		return
	}

	reports := []kubernetesEfficiencyReportModel{}
	for _, fcr := range out.Payload.KubernetesEfficiencyReports {
		report := kubernetesEfficiencyReportModel{
			AggregatedBy:   types.StringValue(fcr.AggregatedBy),
			CreatedAt:      types.StringValue(fcr.CreatedAt),
			DateBucket:     types.StringValue(fcr.DateBucket),
			DateInterval:   types.StringValue(fcr.DateInterval),
			Default:        types.BoolValue(fcr.Default),
			EndDate:        types.StringValue(fcr.EndDate),
			Groupings:      types.StringValue(fcr.Groupings),
			StartDate:      types.StringValue(fcr.StartDate),
			Title:          types.StringValue(fcr.Title),
			Token:          types.StringValue(fcr.Token),
			UserToken:      types.StringValue(fcr.UserToken),
			WorkspaceToken: types.StringValue(fcr.WorkspaceToken),
		}
		reports = append(reports, report)
	}

	data.KubernetesEfficiencyReports = reports
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

}
