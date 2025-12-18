func (p *VantageProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		// ...existing integrations...
		NewCustomProviderResource,
		NewCustomProviderCostsUploadResource,
	}
}