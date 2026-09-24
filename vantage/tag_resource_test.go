package vantage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	testingresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

func TestTagResourceSchema(t *testing.T) {
	var resp resource.SchemaResponse
	TagResource{}.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	tagKey := resp.Schema.Attributes["tag_key"].(schema.StringAttribute)
	if !tagKey.Required {
		t.Fatal("tag_key must be required")
	}

	for _, name := range []string{"hidden", "preferred"} {
		attribute := resp.Schema.Attributes[name].(schema.BoolAttribute)
		if !attribute.Optional || !attribute.Computed {
			t.Fatalf("%s must be optional and computed", name)
		}
	}

	providers := resp.Schema.Attributes["providers"].(schema.SetAttribute)
	if !providers.Computed {
		t.Fatal("providers must be computed")
	}
}

func TestTagResourceUpdateSetsSettingsAndReadsReturnedTag(t *testing.T) {
	ctx := context.Background()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPut || req.URL.Path != "/v2/tags" {
			http.NotFound(w, req)
			return
		}

		var update modelsv2.UpdateTag
		if err := json.NewDecoder(req.Body).Decode(&update); err != nil {
			t.Fatalf("decode update: %v", err)
		}
		if len(update.TagKeys) != 1 || update.TagKeys[0] != "app" {
			t.Fatalf("tag_keys = %#v, want [app]", update.TagKeys)
		}
		if update.Hidden == nil || !*update.Hidden {
			t.Fatalf("hidden = %#v, want true", update.Hidden)
		}
		if update.Preferred == nil || *update.Preferred {
			t.Fatalf("preferred = %#v, want false", update.Preferred)
		}

		writeTagsResponse(t, w, []*modelsv2.Tag{
			{TagKey: "app", Hidden: true, Preferred: false, Providers: []string{"aws", "azure"}},
		})
	}))
	defer srv.Close()

	resource := TagResource{client: clientForServer(t, srv.URL)}
	tag, err := resource.updateTag(ctx, TagResourceModel{
		TagKey:    types.StringValue("app"),
		Hidden:    types.BoolValue(true),
		Preferred: types.BoolValue(false),
	})
	if err != nil {
		t.Fatalf("updateTag: %v", err)
	}
	if tag.TagKey != "app" || !tag.Hidden || tag.Preferred {
		t.Fatalf("tag = %#v", tag)
	}
}

func TestTagResourceReadUsesExactTagKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet || req.URL.Path != "/v2/tags" {
			http.NotFound(w, req)
			return
		}
		if got := req.URL.Query().Get("search_query"); got != "app" {
			t.Fatalf("search_query = %q, want app", got)
		}

		writeTagsResponse(t, w, []*modelsv2.Tag{
			{TagKey: "application", Providers: []string{"aws"}},
			{TagKey: "app", Hidden: true, Providers: []string{"aws"}},
		})
	}))
	defer srv.Close()

	resource := TagResource{client: clientForServer(t, srv.URL)}
	tag, err := resource.readTag(context.Background(), "app")
	if err != nil {
		t.Fatalf("readTag: %v", err)
	}
	if tag.TagKey != "app" {
		t.Fatalf("tag key = %q, want app", tag.TagKey)
	}
}

func TestTagResourceClearResetsSettings(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var update modelsv2.UpdateTag
		if err := json.NewDecoder(req.Body).Decode(&update); err != nil {
			t.Fatalf("decode update: %v", err)
		}
		if update.Hidden == nil || *update.Hidden || update.Preferred == nil || *update.Preferred {
			t.Fatalf("clear update = %#v, want hidden and preferred false", update)
		}
		writeTagsResponse(t, w, []*modelsv2.Tag{{TagKey: "app", Providers: []string{"aws"}}})
	}))
	defer srv.Close()

	resource := TagResource{client: clientForServer(t, srv.URL)}
	if err := resource.clearTag(context.Background(), "app"); err != nil {
		t.Fatalf("clearTag: %v", err)
	}
}

func TestFindTagReturnsNotFoundWithoutExactMatch(t *testing.T) {
	_, err := findTag([]*modelsv2.Tag{{TagKey: "application"}}, "app")
	if !errors.Is(err, errTagNotFound) {
		t.Fatalf("error = %v, want errTagNotFound", err)
	}
}

func TestAccVantageTag_basic(t *testing.T) {
	resourceName := "vantage_tag.test"

	testingresource.Test(t, testingresource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []testingresource.TestStep{
			{
				Config: testAccVantageTagConfig("app", true, false),
				Check: testingresource.ComposeTestCheckFunc(
					testingresource.TestCheckResourceAttr(resourceName, "tag_key", "app"),
					testingresource.TestCheckResourceAttr(resourceName, "hidden", "true"),
					testingresource.TestCheckResourceAttr(resourceName, "preferred", "false"),
					testingresource.TestCheckResourceAttrSet(resourceName, "providers.#"),
				),
			},
			{
				Config: testAccVantageTagConfig("app", false, true),
				Check: testingresource.ComposeTestCheckFunc(
					testingresource.TestCheckResourceAttr(resourceName, "hidden", "false"),
					testingresource.TestCheckResourceAttr(resourceName, "preferred", "true"),
				),
			},
			{
				Config:             testAccVantageTagConfig("app", false, true),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccVantageTagConfig(tagKey string, hidden, preferred bool) string {
	return fmt.Sprintf(`
resource "vantage_tag" "test" {
  tag_key  = %[1]q
  hidden   = %[2]t
  preferred = %[3]t
}
`, tagKey, hidden, preferred)
}

func writeTagsResponse(t *testing.T, w http.ResponseWriter, tags []*modelsv2.Tag) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(&modelsv2.Tags{Tags: tags}); err != nil {
		t.Fatalf("encode response: %v", err)
	}
}
