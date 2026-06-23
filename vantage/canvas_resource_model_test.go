package vantage

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

func TestCanvasDataObject(t *testing.T) {
	ctx := context.Background()

	t.Run("nil data", func(t *testing.T) {
		obj, d := canvasDataObject(ctx, nil)
		if d.HasError() {
			t.Fatalf("unexpected diagnostics: %v", d)
		}
		if !obj.IsNull() {
			t.Fatalf("expected null data object, got %#v", obj)
		}
	})

	t.Run("table and error", func(t *testing.T) {
		errorMessage := "refresh failed"
		obj, d := canvasDataObject(ctx, &modelsv2.CanvasData{
			Error: &errorMessage,
			Table: map[string]any{
				"columns": []map[string]string{{"key": "provider"}},
				"rows":    [][]string{{"aws"}},
			},
		})
		if d.HasError() {
			t.Fatalf("unexpected diagnostics: %v", d)
		}

		attrs := obj.Attributes()
		if got := attrs["error"].(types.String).ValueString(); got != errorMessage {
			t.Fatalf("error = %q, want %q", got, errorMessage)
		}

		tableJSON := attrs["table"].(types.String).ValueString()
		if tableJSON == "" {
			t.Fatal("expected encoded table json")
		}
	})
}

func TestCanvasTableJSON(t *testing.T) {
	t.Run("nil table", func(t *testing.T) {
		value, d := canvasTableJSON(nil)
		if d.HasError() {
			t.Fatalf("unexpected diagnostics: %v", d)
		}
		if !value.IsNull() {
			t.Fatalf("expected null table, got %#v", value)
		}
	})
}
