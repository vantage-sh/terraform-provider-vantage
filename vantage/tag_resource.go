package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
	tagsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/tags"
)

var (
	_ resource.Resource                     = (*TagResource)(nil)
	_ resource.ResourceWithConfigure        = (*TagResource)(nil)
	_ resource.ResourceWithImportState      = (*TagResource)(nil)
	_ resource.ResourceWithConfigValidators = (*TagResource)(nil)
)

type TagResource struct {
	client *Client
}

func NewTagResource() resource.Resource {
	return &TagResource{}
}

type TagResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Key       types.String `tfsdk:"key"`
	Preferred types.Bool   `tfsdk:"preferred"`
	Hidden    types.Bool   `tfsdk:"hidden"`
}

func (r *TagResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tag"
}

func (r *TagResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages whether a tag key is preferred and/or hidden in the Vantage UI.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The tag key, used as the resource ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The tag key to manage.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"preferred": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the tag key is marked as preferred in the Vantage UI.",
			},
			"hidden": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the tag key is hidden from the Vantage UI.",
			},
		},
	}
}

func (r *TagResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.AtLeastOneOf(
			path.MatchRoot("preferred"),
			path.MatchRoot("hidden"),
		),
	}
}

func (r *TagResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*Client)
}

func (r *TagResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TagResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tag, err := r.updateTag(ctx, data)
	if err != nil {
		handleError("Create Tag", &resp.Diagnostics, err)
		return
	}

	data.applyTag(tag)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TagResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TagResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tag, found, err := r.findTag(ctx, data.Key.ValueString())
	if err != nil {
		handleError("Read Tag", &resp.Diagnostics, err)
		return
	}
	if !found {
		// Tags may be preferred/hidden without appearing in GET /tags yet
		// (for example virtual-config-only keys). Keep state rather than
		// forcing recreate on every plan.
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	data.applyTag(tag)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TagResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TagResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tag, err := r.updateTag(ctx, data)
	if err != nil {
		handleError("Update Tag", &resp.Diagnostics, err)
		return
	}

	data.applyTag(tag)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TagResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TagResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	falseVal := false
	params := tagsv2.NewUpdateTagParams().WithContext(ctx).WithUpdateTag(&modelsv2.UpdateTag{
		TagKey:    data.Key.ValueString(),
		Preferred: &falseVal,
		Hidden:    &falseVal,
	})
	_, err := r.client.V2.Tags.UpdateTag(params, r.client.Auth)
	if err != nil {
		handleError("Delete Tag", &resp.Diagnostics, err)
	}
}

func (r *TagResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("key"), req, resp)
}

func (r *TagResource) updateTag(ctx context.Context, data TagResourceModel) (*modelsv2.Tag, error) {
	body := &modelsv2.UpdateTag{
		TagKey: data.Key.ValueString(),
	}
	if !data.Preferred.IsNull() && !data.Preferred.IsUnknown() {
		body.Preferred = data.Preferred.ValueBoolPointer()
	}
	if !data.Hidden.IsNull() && !data.Hidden.IsUnknown() {
		body.Hidden = data.Hidden.ValueBoolPointer()
	}

	params := tagsv2.NewUpdateTagParams().WithContext(ctx).WithUpdateTag(body)
	out, err := r.client.V2.Tags.UpdateTag(params, r.client.Auth)
	if err != nil {
		return nil, err
	}

	key := data.Key.ValueString()
	for _, tag := range out.Payload.Tags {
		if tag != nil && tag.TagKey == key {
			return tag, nil
		}
	}

	// Response can omit virtual-config-only keys; synthesize from the request.
	tag := &modelsv2.Tag{
		TagKey: key,
	}
	if body.Preferred != nil {
		tag.Preferred = *body.Preferred
	} else if !data.Preferred.IsNull() && !data.Preferred.IsUnknown() {
		tag.Preferred = data.Preferred.ValueBool()
	}
	if body.Hidden != nil {
		tag.Hidden = *body.Hidden
	} else if !data.Hidden.IsNull() && !data.Hidden.IsUnknown() {
		tag.Hidden = data.Hidden.ValueBool()
	}
	return tag, nil
}

func (r *TagResource) findTag(ctx context.Context, key string) (*modelsv2.Tag, bool, error) {
	search := key
	limit := int32(1000)
	params := tagsv2.NewGetTagsParams().WithContext(ctx).WithSearchQuery(&search).WithLimit(&limit)
	out, err := r.client.V2.Tags.GetTags(params, r.client.Auth)
	if err != nil {
		return nil, false, err
	}

	for _, tag := range out.Payload.Tags {
		if tag != nil && tag.TagKey == key {
			return tag, true, nil
		}
	}
	return nil, false, nil
}

func (m *TagResourceModel) applyTag(tag *modelsv2.Tag) {
	m.ID = types.StringValue(tag.TagKey)
	m.Key = types.StringValue(tag.TagKey)
	m.Preferred = types.BoolValue(tag.Preferred)
	m.Hidden = types.BoolValue(tag.Hidden)
}
