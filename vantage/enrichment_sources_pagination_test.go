package vantage

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/datasource_enrichment_sources"
)

func TestEnrichmentSourcesPagination(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("page") == "1" {
			w.Write([]byte(`{"enrichment_sources":[{"token":"first"}],"links":{"next":"next"}}`))
			return
		}
		w.Write([]byte(`{"enrichment_sources":[{"token":"second"}]}`))
	}))
	defer srv.Close()
	ctx := context.Background()
	ds := &enrichmentSourcesDataSource{client: clientForServer(t, srv.URL)}
	var sr datasource.SchemaResponse
	ds.Schema(ctx, datasource.SchemaRequest{}, &sr)
	initial := tfsdk.State{Schema: sr.Schema}
	if d := initial.Set(ctx, &enrichmentSourcesDataSourceModel{EnrichmentSources: types.ListNull(datasource_enrichment_sources.EnrichmentSourcesValue{}.Type(ctx))}); d.HasError() {
		t.Fatal(d)
	}
	config := tfsdk.Config{Schema: sr.Schema, Raw: initial.Raw}
	resp := datasource.ReadResponse{State: tfsdk.State{Schema: sr.Schema}}
	ds.Read(ctx, datasource.ReadRequest{Config: config}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatal(resp.Diagnostics)
	}
	var got enrichmentSourcesDataSourceModel
	if d := resp.State.Get(ctx, &got); d.HasError() {
		t.Fatal(d)
	}
	if len(got.EnrichmentSources.Elements()) != 2 {
		t.Fatalf("got %d sources, want 2", len(got.EnrichmentSources.Elements()))
	}
}
