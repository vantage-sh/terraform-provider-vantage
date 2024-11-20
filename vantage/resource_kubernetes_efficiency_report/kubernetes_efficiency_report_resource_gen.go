// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package resource_kubernetes_efficiency_report

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func KubernetesEfficiencyReportResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"aggregated_by": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The column by which the costs are aggregated.",
				MarkdownDescription: "The column by which the costs are aggregated.",
				Validators: []validator.String{
					stringvalidator.OneOf(
						"idle_cost",
						"amount",
						"cost_efficiency",
					),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				Description:         "The date and time, in UTC, the report was created. ISO 8601 Formatted.",
				MarkdownDescription: "The date and time, in UTC, the report was created. ISO 8601 Formatted.",
			},
			"date_bin": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The date bin of the KubernetesEfficiencyReport.",
				MarkdownDescription: "The date bin of the KubernetesEfficiencyReport.",
				Validators: []validator.String{
					stringvalidator.OneOf(
						"day",
						"week",
						"month",
					),
				},
				Default: stringdefault.StaticString("day"),
			},
			"date_bucket": schema.StringAttribute{
				Computed:            true,
				Description:         "How costs are grouped and displayed in the KubernetesEfficiencyReport. Possible values: day, week, month.",
				MarkdownDescription: "How costs are grouped and displayed in the KubernetesEfficiencyReport. Possible values: day, week, month.",
			},
			"date_interval": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The date interval of the KubernetesEfficiencyReport. Incompatible with 'start_date' and 'end_date' parameters. Defaults to 'this_month' if start_date and end_date are not provided.",
				MarkdownDescription: "The date interval of the KubernetesEfficiencyReport. Incompatible with 'start_date' and 'end_date' parameters. Defaults to 'this_month' if start_date and end_date are not provided.",
				Validators: []validator.String{
					stringvalidator.OneOf(
						"this_month",
						"last_7_days",
						"last_30_days",
						"last_month",
						"last_3_months",
						"last_6_months",
						"custom",
						"last_12_months",
						"last_24_months",
						"last_36_months",
						"next_month",
						"next_3_months",
						"next_6_months",
						"next_12_months",
						"year_to_date",
					),
				},
			},
			"default": schema.BoolAttribute{
				Computed:            true,
				Description:         "Indicates whether the KubernetesEfficiencyReport is the default report.",
				MarkdownDescription: "Indicates whether the KubernetesEfficiencyReport is the default report.",
			},
			"end_date": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The end date of the KubernetesEfficiencyReport. ISO 8601 Formatted. Incompatible with 'date_interval' parameter.",
				MarkdownDescription: "The end date of the KubernetesEfficiencyReport. ISO 8601 Formatted. Incompatible with 'date_interval' parameter.",
			},
			"filter": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The filter query language to apply to the KubernetesEfficiencyReport. Additional documentation available at https://docs.vantage.sh/vql.",
				MarkdownDescription: "The filter query language to apply to the KubernetesEfficiencyReport. Additional documentation available at https://docs.vantage.sh/vql.",
			},
			"groupings": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "Grouping values for aggregating costs on the KubernetesEfficiencyReport. Valid groupings: cluster_id, namespace, labeled, category, label, label:<label_name>.",
				MarkdownDescription: "Grouping values for aggregating costs on the KubernetesEfficiencyReport. Valid groupings: cluster_id, namespace, labeled, category, label, label:<label_name>.",
			},
			"start_date": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The start date of the KubernetesEfficiencyReport. ISO 8601 Formatted. Incompatible with 'date_interval' parameter.",
				MarkdownDescription: "The start date of the KubernetesEfficiencyReport. ISO 8601 Formatted. Incompatible with 'date_interval' parameter.",
			},
			"title": schema.StringAttribute{
				Required:            true,
				Description:         "The title of the KubernetesEfficiencyReport.",
				MarkdownDescription: "The title of the KubernetesEfficiencyReport.",
			},
			"token": schema.StringAttribute{
				Computed:            true,
				Description:         "The token of the report",
				MarkdownDescription: "The token of the report",
			},
			"user_token": schema.StringAttribute{
				Computed:            true,
				Description:         "The token for the User who created this KubernetesEfficiencyReport.",
				MarkdownDescription: "The token for the User who created this KubernetesEfficiencyReport.",
			},
			"workspace_token": schema.StringAttribute{
				Required:            true,
				Description:         "The Workspace in which the KubernetesEfficiencyReport will be created.",
				MarkdownDescription: "The Workspace in which the KubernetesEfficiencyReport will be created.",
			},
		},
	}
}

type KubernetesEfficiencyReportModel struct {
	AggregatedBy   types.String `tfsdk:"aggregated_by"`
	CreatedAt      types.String `tfsdk:"created_at"`
	DateBin        types.String `tfsdk:"date_bin"`
	DateBucket     types.String `tfsdk:"date_bucket"`
	DateInterval   types.String `tfsdk:"date_interval"`
	Default        types.Bool   `tfsdk:"default"`
	EndDate        types.String `tfsdk:"end_date"`
	Filter         types.String `tfsdk:"filter"`
	Groupings      types.List   `tfsdk:"groupings"`
	StartDate      types.String `tfsdk:"start_date"`
	Title          types.String `tfsdk:"title"`
	Token          types.String `tfsdk:"token"`
	UserToken      types.String `tfsdk:"user_token"`
	WorkspaceToken types.String `tfsdk:"workspace_token"`
}