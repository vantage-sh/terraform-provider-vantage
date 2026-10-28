package vantage

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

func TestFolderDataSources_dashboardFolderType(t *testing.T) {
	ctx := context.Background()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if got := req.URL.Query().Get("type"); got != "DashboardFolder" {
			t.Errorf("type filter = %q", got)
		}
		first := mockFolder("fldr_parent", "Dashboard Parent", "wrkspc_test", nil)
		second := mockFolder("fldr_child", "Dashboard Child", "wrkspc_test", &first.Token)
		first.Type, second.Type = "DashboardFolder", "DashboardFolder"
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(foldersResponse{Folders: []*modelsv2.Folder{first, second}})
	}))
	defer srv.Close()

	for _, single := range []bool{false, true} {
		var dataSource datasource.DataSource
		client := clientForServer(t, srv.URL)
		if single {
			dataSource = &folderLookupDataSource{client: client}
		} else {
			dataSource = &foldersDataSource{client: client}
		}
		var schemaResp datasource.SchemaResponse
		dataSource.Schema(ctx, datasource.SchemaRequest{}, &schemaResp)
		objectType := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
		values := make(map[string]tftypes.Value)
		for name, attributeType := range objectType.AttributeTypes {
			values[name] = tftypes.NewValue(attributeType, nil)
		}
		values["type"] = tftypes.NewValue(tftypes.String, "DashboardFolder")
		if single {
			values["title"] = tftypes.NewValue(tftypes.String, "Dashboard Child")
		}
		config := tfsdk.Config{Schema: schemaResp.Schema, Raw: tftypes.NewValue(objectType, values)}
		resp := datasource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
		dataSource.Read(ctx, datasource.ReadRequest{Config: config}, &resp)
		if resp.Diagnostics.HasError() {
			t.Fatal(resp.Diagnostics)
		}
		if single {
			var model folderLookupModel
			if d := resp.State.Get(ctx, &model); d.HasError() {
				t.Fatal(d)
			}
			if model.Type.ValueString() != "DashboardFolder" || model.Token.ValueString() != "fldr_child" {
				t.Fatalf("unexpected matched folder: %#v", model)
			}
		} else {
			var model foldersDataSourceModel
			if d := resp.State.Get(ctx, &model); d.HasError() {
				t.Fatal(d)
			}
			if len(model.Folders) != 2 {
				t.Fatalf("folders = %d, want 2", len(model.Folders))
			}
			for _, folder := range model.Folders {
				if folder.Type.ValueString() != "DashboardFolder" || folder.SavedFilterTokens.IsNull() || len(folder.SavedFilterTokens.Elements()) != 0 {
					t.Fatalf("unexpected listed folder: %#v", folder)
				}
			}
		}
	}
}
