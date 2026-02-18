package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_recommendation_view"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

type recommendationViewResourceModel resource_recommendation_view.RecommendationViewModel

func (m *recommendationViewResourceModel) applyPayload(ctx context.Context, payload *modelsv2.RecommendationView) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Token = types.StringPointerValue(payload.Token)
	m.Id = types.StringPointerValue(payload.Token)
	m.Title = types.StringPointerValue(payload.Title)
	m.WorkspaceToken = types.StringPointerValue(payload.WorkspaceToken)
	m.CreatedAt = types.StringPointerValue(payload.CreatedAt)
	m.CreatedBy = types.StringPointerValue(payload.CreatedBy)
	m.StartDate = types.StringPointerValue(payload.StartDate)
	m.EndDate = types.StringPointerValue(payload.EndDate)
	m.TagKey = types.StringPointerValue(payload.TagKey)
	m.TagValue = types.StringPointerValue(payload.TagValue)

	// Handle list fields - convert from API arrays to Terraform lists
	providerIds, d := types.ListValueFrom(ctx, types.StringType, payload.ProviderIds)
	if d.HasError() {
		diags.Append(d...)
		return diags
	}
	m.ProviderIds = providerIds

	billingAccountIds, d := types.ListValueFrom(ctx, types.StringType, payload.BillingAccountIds)
	if d.HasError() {
		diags.Append(d...)
		return diags
	}
	m.BillingAccountIds = billingAccountIds

	accountIds, d := types.ListValueFrom(ctx, types.StringType, payload.AccountIds)
	if d.HasError() {
		diags.Append(d...)
		return diags
	}
	m.AccountIds = accountIds

	regions, d := types.ListValueFrom(ctx, types.StringType, payload.Regions)
	if d.HasError() {
		diags.Append(d...)
		return diags
	}
	m.Regions = regions

	return diags
}

func (m *recommendationViewResourceModel) toCreateModel(ctx context.Context, diags *diag.Diagnostics) *modelsv2.CreateRecommendationView {
	dst := &modelsv2.CreateRecommendationView{
		Title:          m.Title.ValueStringPointer(),
		WorkspaceToken: m.WorkspaceToken.ValueStringPointer(),
	}

	// Handle optional string fields
	if !m.StartDate.IsNull() && !m.StartDate.IsUnknown() {
		dst.StartDate = m.StartDate.ValueString()
	}

	if !m.EndDate.IsNull() && !m.EndDate.IsUnknown() {
		dst.EndDate = m.EndDate.ValueString()
	}

	if !m.TagKey.IsNull() && !m.TagKey.IsUnknown() {
		dst.TagKey = m.TagKey.ValueString()
	}

	if !m.TagValue.IsNull() && !m.TagValue.IsUnknown() {
		dst.TagValue = m.TagValue.ValueString()
	}

	// Handle list fields - default to empty arrays per AGENTS.md guidelines
	if !m.ProviderIds.IsNull() && !m.ProviderIds.IsUnknown() {
		var providerIds []string
		d := m.ProviderIds.ElementsAs(ctx, &providerIds, false)
		if d.HasError() {
			diags.Append(d...)
			return nil
		}
		dst.ProviderIds = providerIds
	} else {
		dst.ProviderIds = []string{}
	}

	if !m.BillingAccountIds.IsNull() && !m.BillingAccountIds.IsUnknown() {
		var billingAccountIds []string
		d := m.BillingAccountIds.ElementsAs(ctx, &billingAccountIds, false)
		if d.HasError() {
			diags.Append(d...)
			return nil
		}
		dst.BillingAccountIds = billingAccountIds
	} else {
		dst.BillingAccountIds = []string{}
	}

	if !m.AccountIds.IsNull() && !m.AccountIds.IsUnknown() {
		var accountIds []string
		d := m.AccountIds.ElementsAs(ctx, &accountIds, false)
		if d.HasError() {
			diags.Append(d...)
			return nil
		}
		dst.AccountIds = accountIds
	} else {
		dst.AccountIds = []string{}
	}

	if !m.Regions.IsNull() && !m.Regions.IsUnknown() {
		var regions []string
		d := m.Regions.ElementsAs(ctx, &regions, false)
		if d.HasError() {
			diags.Append(d...)
			return nil
		}
		dst.Regions = regions
	} else {
		dst.Regions = []string{}
	}

	return dst
}

func (m *recommendationViewResourceModel) toUpdateModel(ctx context.Context, diags *diag.Diagnostics) *modelsv2.UpdateRecommendationView {
	dst := &modelsv2.UpdateRecommendationView{
		Title: m.Title.ValueString(),
	}

	// Handle optional string fields
	if !m.StartDate.IsNull() && !m.StartDate.IsUnknown() {
		dst.StartDate = m.StartDate.ValueString()
	}

	if !m.EndDate.IsNull() && !m.EndDate.IsUnknown() {
		dst.EndDate = m.EndDate.ValueString()
	}

	if !m.TagKey.IsNull() && !m.TagKey.IsUnknown() {
		dst.TagKey = m.TagKey.ValueString()
	}

	if !m.TagValue.IsNull() && !m.TagValue.IsUnknown() {
		dst.TagValue = m.TagValue.ValueString()
	}

	// Handle list fields - default to empty arrays per AGENTS.md guidelines
	if !m.ProviderIds.IsNull() && !m.ProviderIds.IsUnknown() {
		var providerIds []string
		d := m.ProviderIds.ElementsAs(ctx, &providerIds, false)
		if d.HasError() {
			diags.Append(d...)
			return nil
		}
		dst.ProviderIds = providerIds
	} else {
		dst.ProviderIds = []string{}
	}

	if !m.BillingAccountIds.IsNull() && !m.BillingAccountIds.IsUnknown() {
		var billingAccountIds []string
		d := m.BillingAccountIds.ElementsAs(ctx, &billingAccountIds, false)
		if d.HasError() {
			diags.Append(d...)
			return nil
		}
		dst.BillingAccountIds = billingAccountIds
	} else {
		dst.BillingAccountIds = []string{}
	}

	if !m.AccountIds.IsNull() && !m.AccountIds.IsUnknown() {
		var accountIds []string
		d := m.AccountIds.ElementsAs(ctx, &accountIds, false)
		if d.HasError() {
			diags.Append(d...)
			return nil
		}
		dst.AccountIds = accountIds
	} else {
		dst.AccountIds = []string{}
	}

	if !m.Regions.IsNull() && !m.Regions.IsUnknown() {
		var regions []string
		d := m.Regions.ElementsAs(ctx, &regions, false)
		if d.HasError() {
			diags.Append(d...)
			return nil
		}
		dst.Regions = regions
	} else {
		dst.Regions = []string{}
	}

	return dst
}
