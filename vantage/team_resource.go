package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
	teamsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/teams"
)

type TeamResource struct {
	client *Client
}

func NewTeamResource() resource.Resource {
	return &TeamResource{}
}

type TeamResourceModel struct {
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	WorkspaceTokens types.List   `tfsdk:"workspace_tokens"`
	UserTokens      types.List   `tfsdk:"user_tokens"`
	UserEmails      types.List   `tfsdk:"user_emails"`
	Token           types.String `tfsdk:"token"`
}

func (r *TeamResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team"
}

func (r TeamResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the team.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the team.",
				Optional:            true,
			},
			"workspace_tokens": schema.ListAttribute{
				MarkdownDescription: "Workspace tokens to add the team to.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"user_tokens": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "User tokens.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"user_emails": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "User emails.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"token": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique team identifier.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		MarkdownDescription: "Manages a Team.",
	}
}

func (r TeamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *TeamResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := teamsv2.NewCreateTeamParams()

	var userTokens []types.String
	if !data.UserTokens.IsNull() && !data.UserTokens.IsUnknown() {
		userTokens = make([]types.String, 0, len(data.UserTokens.Elements()))
		resp.Diagnostics.Append(data.UserTokens.ElementsAs(ctx, &userTokens, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	var userEmails []types.String
	if !data.UserEmails.IsNull() && !data.UserEmails.IsUnknown() {
		userEmails = make([]types.String, 0, len(data.UserEmails.Elements()))
		resp.Diagnostics.Append(data.UserEmails.ElementsAs(ctx, &userEmails, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	var workspaceTokens []types.String
	if !data.WorkspaceTokens.IsNull() && !data.WorkspaceTokens.IsUnknown() {
		workspaceTokens = make([]types.String, 0, len(data.WorkspaceTokens.Elements()))
		resp.Diagnostics.Append(data.WorkspaceTokens.ElementsAs(ctx, &workspaceTokens, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	rt := &modelsv2.PostTeams{
		Name:            data.Name.ValueStringPointer(),
		Description:     data.Description.ValueString(),
		UserTokens:      fromStringsValue(userTokens),
		UserEmails:      fromStringsValue(userEmails),
		WorkspaceTokens: fromStringsValue(workspaceTokens),
	}

	params.WithTeams(rt)
	out, err := r.client.V2.Teams.CreateTeam(params, r.client.Auth)
	if err != nil {
		if e, ok := err.(*teamsv2.CreateTeamBadRequest); ok {
			handleBadRequest("Create Team Resource", &resp.Diagnostics, e.GetPayload())
			return
		}

		handleError("Create Team Resource", &resp.Diagnostics, err)
		return
	}

	data.Token = types.StringValue(out.Payload.Token)
	data.Name = types.StringValue(out.Payload.Name)
	data.Description = types.StringValue(out.Payload.Description)
	if out.Payload.WorkspaceTokens != nil {
		workspaceTokensValue := make([]types.String, 0, len(out.Payload.WorkspaceTokens))
		for _, token := range out.Payload.WorkspaceTokens {
			workspaceTokensValue = append(workspaceTokensValue, types.StringValue(token))
		}
		list, diag := types.ListValueFrom(ctx, types.StringType, workspaceTokensValue)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}
		data.WorkspaceTokens = list
	}

	if out.Payload.UserTokens != nil {
		userTokensValue := make([]types.String, 0, len(out.Payload.UserTokens))
		for _, token := range out.Payload.UserTokens {
			userTokensValue = append(userTokensValue, types.StringValue(token))
		}
		list, diag := types.ListValueFrom(ctx, types.StringType, userTokensValue)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}
		data.UserTokens = list
	}

	if out.Payload.UserEmails != nil {
		userEmailsValue := make([]types.String, 0, len(out.Payload.UserEmails))
		for _, email := range out.Payload.UserEmails {
			userEmailsValue = append(userEmailsValue, types.StringValue(email))
		}
		list, diag := types.ListValueFrom(ctx, types.StringType, userEmailsValue)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}
		data.UserEmails = list
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r TeamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state *TeamResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := teamsv2.NewGetTeamParams()
	params.SetTeamToken(state.Token.ValueString())
	out, err := r.client.V2.Teams.GetTeam(params, r.client.Auth)
	if err != nil {
		if _, ok := err.(*teamsv2.GetTeamNotFound); ok {
			resp.State.RemoveResource(ctx)
			return
		}

		handleError("Get Team Resource", &resp.Diagnostics, err)
		return
	}

	state.Token = types.StringValue(out.Payload.Token)
	state.Name = types.StringValue(out.Payload.Name)
	state.Description = types.StringValue(out.Payload.Description)

	userTokens, diag := types.ListValueFrom(ctx, types.StringType, out.Payload.UserTokens)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}
	state.UserTokens = userTokens

	userEmails, diag := types.ListValueFrom(ctx, types.StringType, out.Payload.UserEmails)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}
	state.UserEmails = userEmails

	workspaceTokensValue, diag := types.ListValueFrom(ctx, types.StringType, out.Payload.WorkspaceTokens)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}
	state.WorkspaceTokens = workspaceTokensValue

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r TeamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *TeamResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	params := teamsv2.NewUpdateTeamParams()
	params.WithTeamToken(data.Token.ValueString())

	userTokensList, diag := types.ListValueFrom(ctx, types.StringType, data.UserTokens)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	var userTokens []string
	userTokensList.ElementsAs(ctx, userTokens, false)

	userEmailsList, diag := types.ListValueFrom(ctx, types.StringType, data.UserEmails)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}
	var userEmails []string
	userEmailsList.ElementsAs(ctx, userEmails, false)

	workspaceTokensList, diag := types.ListValueFrom(ctx, types.StringType, data.WorkspaceTokens)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}
	var workspaceTokens []string
	workspaceTokensList.ElementsAs(ctx, workspaceTokens, false)

	model := &modelsv2.PutTeams{
		Name:            data.Name.ValueString(),
		Description:     data.Description.ValueString(),
		UserTokens:      userTokens,
		UserEmails:      userEmails,
		WorkspaceTokens: workspaceTokens,
	}

	params.WithTeams(model)
	out, err := r.client.V2.Teams.UpdateTeam(params, r.client.Auth)
	if err != nil {
		handleError("Update Team Resource", &resp.Diagnostics, err)
		return
	}

	data.Name = types.StringValue(out.Payload.Name)
	data.Description = types.StringValue(out.Payload.Description)
	workspaceTokensValue, diag := types.ListValueFrom(ctx, types.StringType, out.Payload.WorkspaceTokens)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}
	data.WorkspaceTokens = workspaceTokensValue

	userTokensValue, diag := types.ListValueFrom(ctx, types.StringType, out.Payload.UserTokens)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}
	data.UserTokens = userTokensValue

	userEmailsValue, diag := types.ListValueFrom(ctx, types.StringType, out.Payload.UserEmails)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}
	data.UserEmails = userEmailsValue

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r TeamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state *TeamResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := teamsv2.NewDeleteTeamParams()
	params.SetTeamToken(state.Token.ValueString())
	_, err := r.client.V2.Teams.DeleteTeam(params, r.client.Auth)
	if err != nil {
		handleError("Delete Team Resource", &resp.Diagnostics, err)
	}
}

// Configure adds the provider configured client to the data source.
func (r *TeamResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*Client)
}
