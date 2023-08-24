package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
	costsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/costs"
)

type CostReportFolderResource struct {
	client *Client
}

func NewCostReportFolderResource() resource.Resource {
	return &CostReportFolderResource{}
}

type CostReportFolderResourceModel struct {
	Title             types.String `tfsdk:"title"`
	ParentFolderToken types.String `tfsdk:"parent_folder_token"`
	Token             types.String `tfsdk:"token"`
}

func (r *CostReportFolderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cost_report_folder"
}

func (r CostReportFolderResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"title": schema.StringAttribute{
				MarkdownDescription: "Title of the folder",
				Required:            true,
			},
			"parent_folder_token": schema.StringAttribute{
				MarkdownDescription: "Token of the folder's parent folder",
				Optional:            true,
			},
			"token": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique folder identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		MarkdownDescription: "Manages a CostReportFolder.",
	}
}

func (r CostReportFolderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *CostReportFolderResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := costsv2.NewCreateCostReportFolderParams()
	rf := &modelsv2.PostReportsFolders{
		Title:             data.Title.ValueStringPointer(),
		ParentFolderToken: data.ParentFolderToken.ValueString(),
	}
	params.WithReportsFolders(rf)
	out, err := r.client.V2.Costs.CreateCostReportFolder(params, r.client.Auth)
	if err != nil {
		handleError("Create Cost Report Folder Resource", &resp.Diagnostics, err)
		return
	}

	data.Token = types.StringValue(out.Payload.Token)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r CostReportFolderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state *CostReportFolderResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := costsv2.NewGetCostReportFolderParams()
	params.SetFolderToken(state.Token.ValueString())
	out, err := r.client.V2.Costs.GetCostReportFolder(params, r.client.Auth)
	if err != nil {
		if _, ok := err.(*costsv2.GetCostReportFolderNotFound); ok {
			resp.State.RemoveResource(ctx)
			return
		}

		handleError("Get Cost Report Folder Resource", &resp.Diagnostics, err)
		return
	}

	state.Token = types.StringValue(out.Payload.Token)
	state.ParentFolderToken = types.StringValue(out.Payload.ParentFolderToken)
	state.Title = types.StringValue(out.Payload.Title)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r CostReportFolderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *CostReportFolderResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := costsv2.NewUpdateCostReportFolderParams()
	params.WithFolderToken(data.Token.ValueString())
	model := &modelsv2.PutReportsFolders{
		ParentFolderToken: data.ParentFolderToken.ValueString(),
		Title:             data.Title.ValueString(),
	}
	params.WithReportsFolders(model)
	out, err := r.client.V2.Costs.UpdateCostReportFolder(params, r.client.Auth)
	if err != nil {
		handleError("Update Cost Report Folder Resource", &resp.Diagnostics, err)
		return
	}

	pft := out.Payload.ParentFolderToken
	if pft != "" {
		data.ParentFolderToken = types.StringValue(pft)
	}
	data.Title = types.StringValue(out.Payload.Title)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r CostReportFolderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	panic("not implemented")
}

// Configure adds the provider configured client to the data source.
func (r *CostReportFolderResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*Client)
}
