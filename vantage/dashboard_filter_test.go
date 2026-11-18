package vantage

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

func TestDashboardFilterRequests(t *testing.T) {
	for _, value := range []types.String{types.StringNull(), types.StringValue(""), types.StringValue(`vantage.provider = 'aws'`)} {
		t.Run(value.String(), func(t *testing.T) {
			model := &dashboardModel{Filter: value}
			var diags diag.Diagnostics
			create := model.toCreate(context.Background(), &diags)
			update := model.toUpdate(context.Background(), &diags)
			if diags.HasError() {
				t.Fatal(diags)
			}
			if value.IsNull() && create.Filter != nil {
				t.Fatal("create must omit unset filter")
			}
			if !value.IsNull() && (create.Filter == nil || *create.Filter != value.ValueString()) {
				t.Fatal("create lost filter")
			}
			encoded, err := json.Marshal(update)
			if err != nil {
				t.Fatal(err)
			}
			var body map[string]interface{}
			if err := json.Unmarshal(encoded, &body); err != nil {
				t.Fatal(err)
			}
			if body["filter"] != value.ValueString() {
				t.Fatalf("update filter = %#v, want %q", body["filter"], value.ValueString())
			}
		})
	}
}

func TestDashboardFilterState(t *testing.T) {
	normalized := `vantage.provider = 'aws'`
	changed := `vantage.provider = 'gcp'`
	baseline, _ := json.Marshal(normalized)
	configured := types.StringValue(`(vantage.provider = 'aws')`)
	for _, tc := range []struct {
		name     string
		remote   *string
		previous []byte
		want     types.String
	}{
		{"normalized", &normalized, baseline, configured},
		{"remote edit", &changed, baseline, types.StringValue(changed)},
		{"remote clear", nil, baseline, types.StringNull()},
		{"import", &normalized, nil, types.StringValue(normalized)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := dashboardFilterState(configured, tc.remote, tc.previous); !got.Equal(tc.want) {
				t.Fatalf("got %s, want %s", got, tc.want)
			}
		})
	}
	model := &dashboardModel{Filter: configured}
	if diags := model.applyPayload(context.Background(), &modelsv2.Dashboard{}); diags.HasError() {
		t.Fatal(diags)
	}
	if !model.Filter.IsNull() {
		t.Fatal("absent API filter must clear stale state")
	}
	value, diags := dashboardDataSourceValueFromPayload(context.Background(), &modelsv2.Dashboard{Filter: &normalized})
	if diags.HasError() {
		t.Fatal(diags)
	}
	if value.Filter.ValueString() != normalized {
		t.Fatal("data source lost filter")
	}
}
