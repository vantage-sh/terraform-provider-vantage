package vantage

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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

var errTagNotFound = errors.New("tag not found")

type TagResource struct {
	client *Client
}

type TagResourceModel struct {
	TagKey    types.String `tfsdk:"tag_key"`
	Hidden    types.Bool   `tfsdk:"hidden"`
	Preferred types.Bool   `tfsdk:"preferred"`
	Providers types.Set    `tfsdk:"providers"`
}

func NewTagResource() resource.Resource {
	return &TagResource{}
}

func (r *TagResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tag"
}

func (r TagResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages visibility and preference settings for a standard tag dimension.",
		Attributes: map[string]schema.Attribute{
			"tag_key": schema.StringAttribute{
				MarkdownDescription: "The standard tag key to manage.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"hidden": schema.BoolAttribute{
				MarkdownDescription: "Whether the tag is hidden from the Vantage UI.",
				Optional:            true,
				Computed:            true,
			},
			"preferred": schema.BoolAttribute{
				MarkdownDescription: "Whether the tag is marked as preferred in the Vantage UI.",
				Optional:            true,
				Computed:            true,
			},
			"providers": schema.SetAttribute{
				MarkdownDescription: "Providers that expose the tag key.",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (r *TagResource) ConfigValidators(context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.AtLeastOneOf(
			path.MatchRoot("hidden"),
			path.MatchRoot("preferred"),
		),
	}
}

func (r TagResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TagResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tag, err := r.updateTag(ctx, data)
	if err != nil {
		handleError("Create Tag Resource", &resp.Diagnostics, err)
		return
	}

	resp.Diagnostics.Append(data.applyPayload(ctx, tag)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r TagResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TagResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tag, err := r.readTag(ctx, state.TagKey.ValueString())
	if errors.Is(err, errTagNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		handleError("Read Tag Resource", &resp.Diagnostics, err)
		return
	}

	resp.Diagnostics.Append(state.applyPayload(ctx, tag)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r TagResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TagResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tag, err := r.updateTag(ctx, data)
	if err != nil {
		handleError("Update Tag Resource", &resp.Diagnostics, err)
		return
	}

	resp.Diagnostics.Append(data.applyPayload(ctx, tag)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r TagResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state TagResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.clearTag(ctx, state.TagKey.ValueString()); err != nil {
		handleError("Delete Tag Resource", &resp.Diagnostics, err)
	}
}

func (r *TagResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*Client)
}

func (r TagResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tag_key"), req.ID)...)
}

func (r TagResource) readTag(ctx context.Context, tagKey string) (*modelsv2.Tag, error) {
	limit := int32(1000)
	params := tagsv2.
		NewGetTagsParams().
		WithContext(ctx).
		WithLimit(&limit).
		WithSearchQuery(&tagKey)
	out, err := r.client.V2.Tags.GetTags(params, r.client.Auth)
	if err != nil {
		return nil, err
	}

	return findTag(out.Payload.Tags, tagKey)
}

func (r TagResource) updateTag(ctx context.Context, data TagResourceModel) (*modelsv2.Tag, error) {
	update := &modelsv2.UpdateTag{TagKeys: []string{data.TagKey.ValueString()}}
	if !data.Hidden.IsNull() && !data.Hidden.IsUnknown() {
		update.Hidden = data.Hidden.ValueBoolPointer()
	}
	if !data.Preferred.IsNull() && !data.Preferred.IsUnknown() {
		update.Preferred = data.Preferred.ValueBoolPointer()
	}

	params := tagsv2.NewUpdateTagParams().WithContext(ctx).WithUpdateTag(update)
	out, err := r.client.V2.Tags.UpdateTag(params, r.client.Auth)
	if err != nil {
		return nil, err
	}

	return findTag(out.Payload.Tags, data.TagKey.ValueString())
}

func (r TagResource) clearTag(ctx context.Context, tagKey string) error {
	value := false
	params := tagsv2.NewUpdateTagParams().WithContext(ctx).WithUpdateTag(&modelsv2.UpdateTag{
		TagKeys:   []string{tagKey},
		Hidden:    &value,
		Preferred: &value,
	})
	_, err := r.client.V2.Tags.UpdateTag(params, r.client.Auth)
	return err
}

func findTag(tags []*modelsv2.Tag, tagKey string) (*modelsv2.Tag, error) {
	for _, tag := range tags {
		if tag.TagKey == tagKey {
			return tag, nil
		}
	}

	return nil, errTagNotFound
}

func (m *TagResourceModel) applyPayload(ctx context.Context, tag *modelsv2.Tag) diag.Diagnostics {
	providers, diags := types.SetValueFrom(ctx, types.StringType, tag.Providers)
	if diags.HasError() {
		return diags
	}

	m.TagKey = types.StringValue(tag.TagKey)
	m.Hidden = types.BoolValue(tag.Hidden)
	m.Preferred = types.BoolValue(tag.Preferred)
	m.Providers = providers
	return diags
}
