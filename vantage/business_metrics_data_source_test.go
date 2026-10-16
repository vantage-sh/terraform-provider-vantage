package vantage

import (
	"context"
	"reflect"
	"testing"

	"github.com/vantage-sh/terraform-provider-vantage/vantage/datasource_business_metrics"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

func TestCostReportTokensListForDataSourceLabelFilters(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	costReportToken := "rprt_123"
	want := map[string][]string{
		"team": {"platform", "data"},
	}

	list, diags := costReportTokensListForDataSource(ctx, []*modelsv2.AttachedCostReportForBusinessMetric{
		{
			CalculationType: "unit_cost",
			CostReportToken: &costReportToken,
			LabelFilter:     []string{},
			LabelFilters:    want,
			UnitScale:       "per_unit",
		},
	})
	if diags.HasError() {
		t.Fatalf("building cost report tokens: %v", diags)
	}
	if len(list.Elements()) != 1 {
		t.Fatalf("cost report token count = %d, want 1", len(list.Elements()))
	}

	token, ok := list.Elements()[0].(datasource_business_metrics.CostReportTokensWithMetadataValue)
	if !ok {
		t.Fatalf("cost report token is %T, want CostReportTokensWithMetadataValue", list.Elements()[0])
	}

	got := map[string][]string{}
	if d := token.LabelFilters.ElementsAs(ctx, &got, false); d.HasError() {
		t.Fatalf("reading label_filters: %v", d)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("label_filters = %#v, want %#v", got, want)
	}
}
