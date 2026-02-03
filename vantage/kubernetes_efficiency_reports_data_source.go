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
	KubernetesEfficiencyReports []kubernetesEfficiencyReportDataModel `tfsdk:"kubernetes_efficiency_reports"`
}

type kubernetesEfficiencyReportDataModel struct {
	AggregatedBy   types.String `tfsdk:"aggregated_by"`
	CreatedAt      types.String `tfsdk:"created_at"`
	DateBucket     types.String `tfsdk:"date_bucket"`
	DateInterval   types.String `tfsdk:"date_interval"`
	Default        types.Bool   `tfsdk:"default"`
	EndDate        types.String `tfsdk:"end_date"`
	Filter         types.String `tfsdk:"filter"`
	Groupings      types.String `tfsdk:"groupings"`
	StartDate      types.String `tfsdk:"start_date"`
	Title          types.String `tfsdk:"title"`
	Token          types.String `tfsdk:"token"`
	Id             types.String `tfsdk:"id"`
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

	reports := []kubernetesEfficiencyReportDataModel{}
	for _, ker := range out.Payload.KubernetesEfficiencyReports {
		report := kubernetesEfficiencyReportDataModel{
			AggregatedBy:   types.StringValue(ker.AggregatedBy),
			CreatedAt:      types.StringValue(ker.CreatedAt),
			DateBucket:     types.StringValue(ker.DateBucket),
			DateInterval:   types.StringPointerValue(ker.DateInterval),
			Default:        types.BoolValue(ker.Default),
			EndDate:        types.StringPointerValue(ker.EndDate),
			Groupings:      types.StringPointerValue(ker.Groupings),
			StartDate:      types.StringPointerValue(ker.StartDate),
			Title:          types.StringValue(ker.Title),
			Token:          types.StringValue(ker.Token),
			Id:             types.StringValue(ker.Token),
			UserToken:      types.StringValue(ker.UserToken),
			WorkspaceToken: types.StringValue(ker.WorkspaceToken),
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
