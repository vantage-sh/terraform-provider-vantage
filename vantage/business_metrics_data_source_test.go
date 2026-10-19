package vantage

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/datasource_business_metrics"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

func TestCostReportTokensListForDataSourceLabelFilters(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name    string
		filters string
		want    map[string][]string
	}{
		{
			name:    "multiple keys and values",
			filters: `{"team":["platform","data"],"environment":["production"]}`,
			want: map[string][]string{
				"team":        {"platform", "data"},
				"environment": {"production"},
			},
		},
		{name: "empty", filters: `{}`},
		{name: "null", filters: `null`},
		{name: "omitted"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			var payload modelsv2.AttachedCostReportForBusinessMetric
			payloadJSON := `{"calculation_type":"unit_cost","cost_report_token":"rprt_123","unit_scale":"per_unit"`
			if tc.filters != "" {
				payloadJSON += `,"label_filters":` + tc.filters
			}
			payloadJSON += `}`
			if err := json.Unmarshal([]byte(payloadJSON), &payload); err != nil {
				t.Fatal(err)
			}
			list, diags := costReportTokensListForDataSource(ctx, []*modelsv2.AttachedCostReportForBusinessMetric{&payload})
			if diags.HasError() {
				t.Fatalf("building cost report tokens: %v", diags)
			}
			if len(list.Elements()) != 1 {
				t.Fatalf("cost report token count = %d, want 1", len(list.Elements()))
			}
			token, ok := list.Elements()[0].(datasource_business_metrics.CostReportTokensWithMetadataValue)
			if !ok {
				t.Fatalf("cost report token = %T, want CostReportTokensWithMetadataValue", list.Elements()[0])
			}
			object, diags := token.ToObjectValue(ctx)
			if diags.HasError() {
				t.Fatalf("reading cost report token: %v", diags)
			}
			filters, ok := object.Attributes()["label_filters"].(types.Map)
			if !ok {
				t.Fatalf("label_filters = %T, want types.Map", object.Attributes()["label_filters"])
			}
			if tc.want == nil {
				if !filters.IsNull() {
					t.Fatalf("label_filters = %v, want null", filters)
				}
				return
			}
			var got map[string][]string
			if d := filters.ElementsAs(ctx, &got, false); d.HasError() {
				t.Fatalf("reading label_filters: %v", d)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("label_filters = %#v, want %#v", got, tc.want)
			}
		})
	}
}
