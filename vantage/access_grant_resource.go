package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
	accessgrantsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/access_grants"
)

type AccessGrantResource struct {
	client *Client
}

func NewAccessGrantResource() resource.Resource {
	return &AccessGrantResource{}
}

type AccessGrantResourceModel struct {
	Token         types.String `tfsdk:"token"`
	ResourceToken types.String `tfsdk:"resource_token"`
	TeamToken     types.String `tfsdk:"team_token"`
	Access        types.String `tfsdk:"access"`
}

func (r *AccessGrantResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_access_grant"
}

func (r AccessGrantResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"resource_token": schema.StringAttribute{
				MarkdownDescription: "Token of the resource being granted.",
				Required:            true,
			},
			"team_token": schema.StringAttribute{
				MarkdownDescription: "Token of the team being granted.",
				Required:            true,
			},
			"access": schema.StringAttribute{
				MarkdownDescription: "Access level of the grant. Must be either `allowed` or `denied`.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("allowed", "denied"),
				},
			},
			"token": schema.StringAttribute{
				Computed:            true,
				Required:            false,
				MarkdownDescription: "Token of the access grant.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		MarkdownDescription: "Manages an AccessGrant.",
	}
}

func (r AccessGrantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *AccessGrantResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := accessgrantsv2.NewCreateAccessGrantParams()
	body := &modelsv2.PostAccessGrants{
		ResourceToken: data.ResourceToken.ValueStringPointer(),
		TeamToken:     data.TeamToken.ValueStringPointer(),
		Access:        data.Access.ValueString(),
	}
	params.WithAccessGrants(body)
	out, err := r.client.V2.AccessGrants.CreateAccessGrant(params, r.client.Auth)
	if err != nil {
		handleError("Create Access Grant Resource", &resp.Diagnostics, err)
		return
	}

	data.Token = types.StringValue(out.Payload.Token)
	data.ResourceToken = types.StringValue(out.Payload.ResourceToken)
	data.TeamToken = types.StringValue(out.Payload.TeamToken)
	data.Access = types.StringValue(out.Payload.Access)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r AccessGrantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state *AccessGrantResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := accessgrantsv2.NewGetAccessGrantParams()

	params.SetAccessGrantToken(state.Token.ValueString())
	out, err := r.client.V2.AccessGrants.GetAccessGrant(params, r.client.Auth)
	if err != nil {
		if _, ok := err.(*accessgrantsv2.GetAccessGrantNotFound); ok {
			resp.State.RemoveResource(ctx)
			return
		}

		handleError("Get Saved Filter Resource", &resp.Diagnostics, err)
		return
	}

	state.Token = types.StringValue(out.Payload.Token)
	state.ResourceToken = types.StringValue(out.Payload.ResourceToken)
	state.TeamToken = types.StringValue(out.Payload.TeamToken)
	state.Access = types.StringValue(out.Payload.Access)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r AccessGrantResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *AccessGrantResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := accessgrantsv2.NewUpdateAccessGrantParams()

	params.WithAccessGrantToken(data.Token.ValueString())
	model := &modelsv2.PutAccessGrants{
		Access: data.Access.ValueStringPointer(),
	}

	params.WithAccessGrants(model)
	out, err := r.client.V2.AccessGrants.UpdateAccessGrant(params, r.client.Auth)
	if err != nil {
		handleError("Update Saved Filter Resource", &resp.Diagnostics, err)
		return
	}

	data.Access = types.StringValue(out.Payload.Access)
	data.ResourceToken = types.StringValue(out.Payload.ResourceToken)
	data.TeamToken = types.StringValue(out.Payload.TeamToken)
	data.Token = types.StringValue(out.Payload.Token)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r AccessGrantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state *AccessGrantResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := accessgrantsv2.NewDeleteAccessGrantParams()
	params.SetAccessGrantToken(state.Token.ValueString())
	_, err := r.client.V2.AccessGrants.DeleteAccessGrant(params, r.client.Auth)
	if err != nil {
		handleError("Delete Saved Filter Resource", &resp.Diagnostics, err)
	}
}

// Configure adds the provider configured client to the data source.
func (r *AccessGrantResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*Client)
}
