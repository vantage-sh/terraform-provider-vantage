// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package resource_resource_report

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func ResourceReportResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"created_at": schema.StringAttribute{
				Computed:            true,
				Description:         "The date and time, in UTC, the report was created. ISO 8601 Formatted.",
				MarkdownDescription: "The date and time, in UTC, the report was created. ISO 8601 Formatted.",
			},
			"created_by_token": schema.StringAttribute{
				Computed:            true,
				Description:         "The token for the User or Team who created this ResourceReport.",
				MarkdownDescription: "The token for the User or Team who created this ResourceReport.",
			},
			"filter": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The VQL filter for the ResourceReport.",
				MarkdownDescription: "The VQL filter for the ResourceReport.",
			},
			"title": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The title of the ResourceReport.",
				MarkdownDescription: "The title of the ResourceReport.",
			},
			"token": schema.StringAttribute{
				Computed:            true,
				Description:         "The token of the report",
				MarkdownDescription: "The token of the report",
			},
			"user_token": schema.StringAttribute{
				Computed:            true,
				Description:         "The token for the User who created this ResourceReport.",
				MarkdownDescription: "The token for the User who created this ResourceReport.",
			},
			"workspace_token": schema.StringAttribute{
				Required:            true,
				Description:         "The token of the Workspace to add the ResourceReport to.",
				MarkdownDescription: "The token of the Workspace to add the ResourceReport to.",
			},
		},
	}
}

type ResourceReportModel struct {
	CreatedAt      types.String `tfsdk:"created_at"`
	CreatedByToken types.String `tfsdk:"created_by_token"`
	Filter         types.String `tfsdk:"filter"`
	Title          types.String `tfsdk:"title"`
	Token          types.String `tfsdk:"token"`
	UserToken      types.String `tfsdk:"user_token"`
	WorkspaceToken types.String `tfsdk:"workspace_token"`
}