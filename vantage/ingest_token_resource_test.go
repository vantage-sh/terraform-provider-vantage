package vantage

import (
	"context"
	"encoding/json"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

func TestIngestTokenLifecycle(t *testing.T) {
	ctx := context.Background()
	name := "test"
	retention := int32(30)
	var revoked, missingFeed bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == "POST" && r.URL.Path == "/v2/ingest_tokens":
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body["feed_token"] != "feed" || body["name"] != name || body["retention_days"] != float64(retention) || body["expires_at"] != nil {
				t.Errorf("unexpected create body: %+v, %v", body, err)
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(modelsv2.IngestTokenCreation{Token: "tok", FeedToken: "feed", Name: &name, RetentionDays: &retention, ExpiresAt: "2027-01-01T00:00:00Z", PlaintextToken: "secret", Provider: "test", Last4: "cret", IssuedAt: "2026-01-01T00:00:00Z"})
		case r.Method == "GET" && r.URL.Path == "/v2/ingest_tokens":
			if missingFeed {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"errors":["not found"]}`))
				return
			}
			if r.URL.Query().Get("feed_token") != "feed" {
				t.Error("missing feed token")
			}
			if r.URL.Query().Get("page") == "1" {
				w.Write([]byte(`{"tokens":[],"links":{"next":"next"}}`))
				return
			}
			json.NewEncoder(w).Encode(modelsv2.IngestTokens{Tokens: []*modelsv2.IngestToken{{Token: "tok", FeedToken: "feed", Name: &name, RetentionDays: &retention, ExpiresAt: "2027-01-01T00:00:00Z", Provider: "test", Last4: "cret", IssuedAt: "2026-01-01T00:00:00Z"}}})
		case r.Method == "POST" && r.URL.Path == "/v2/ingest_tokens/tok/revoke":
			revoked = true
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	r := &ingestTokenResource{client: clientForServer(t, srv.URL)}
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	s := schemaResp.Schema
	model := ingestTokenModel{FeedToken: types.StringValue("feed"), Name: types.StringValue(name), RetentionDays: types.Int64Value(30), ExpiresAt: types.StringUnknown(), ID: types.StringUnknown(), Token: types.StringUnknown(), PlaintextToken: types.StringUnknown(), Provider: types.StringUnknown(), Last4: types.StringUnknown(), IssuedAt: types.StringUnknown(), CreatedBy: types.StringUnknown()}
	plan := tfsdk.Plan{Schema: s}
	if d := plan.Set(ctx, &model); d.HasError() {
		t.Fatal(d)
	}
	created := resource.CreateResponse{State: tfsdk.State{Schema: s}}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, &created)
	if created.Diagnostics.HasError() {
		t.Fatal(created.Diagnostics)
	}
	var got ingestTokenModel
	if d := created.State.Get(ctx, &got); d.HasError() {
		t.Fatal(d)
	}
	if got.Token.ValueString() != "tok" || got.PlaintextToken.ValueString() != "secret" || got.ExpiresAt.ValueString() != "2027-01-01T00:00:00Z" {
		t.Fatalf("unexpected create state: %+v", got)
	}
	read := resource.ReadResponse{State: created.State}
	r.Read(ctx, resource.ReadRequest{State: created.State}, &read)
	if read.Diagnostics.HasError() {
		t.Fatal(read.Diagnostics)
	}
	if d := read.State.Get(ctx, &got); d.HasError() {
		t.Fatal(d)
	}
	if got.PlaintextToken.ValueString() != "secret" || got.RetentionDays.ValueInt64() != 30 {
		t.Fatalf("unexpected refreshed state: %+v", got)
	}
	for _, tc := range []struct {
		name, expiry types.String
		invalid      bool
	}{
		{types.StringValue(" padded "), types.StringValue("2027-01-01T00:00:00Z"), true},
		{types.StringValue(""), types.StringNull(), true},
		{types.StringValue("test"), types.StringValue("bad"), true},
		{types.StringValue("test"), types.StringValue("2027-01-01T00:00:00.5Z"), true},
		{types.StringUnknown(), types.StringUnknown(), false},
		{types.StringNull(), types.StringValue("2027-01-01T00:00:00Z"), false},
	} {
		configModel := model
		configModel.Name, configModel.ExpiresAt = tc.name, tc.expiry
		configPlan := tfsdk.Plan{Schema: s}
		if d := configPlan.Set(ctx, &configModel); d.HasError() {
			t.Fatal(d)
		}
		config := tfsdk.Config{Schema: s, Raw: configPlan.Raw}
		validation := resource.ValidateConfigResponse{}
		r.ValidateConfig(ctx, resource.ValidateConfigRequest{Config: config}, &validation)
		if validation.Diagnostics.HasError() != tc.invalid {
			t.Fatalf("validation name=%v expiry=%v: %v", tc.name, tc.expiry, validation.Diagnostics)
		}
	}
	missingFeed = true
	missing := resource.ReadResponse{State: read.State}
	r.Read(ctx, resource.ReadRequest{State: read.State}, &missing)
	if missing.Diagnostics.HasError() || !missing.State.Raw.IsNull() {
		t.Fatalf("missing feed: %v, state %v", missing.Diagnostics, missing.State.Raw)
	}
	for _, input := range []struct{ name, expiry string }{{"", "2027-01-01T00:00:00Z"}, {" padded ", "2027-01-01T00:00:00Z"}, {"test", "2027-01-01T00:00:00.5Z"}} {
		invalid := model
		invalid.Name = types.StringValue(input.name)
		invalid.ExpiresAt = types.StringValue(input.expiry)
		invalidPlan := tfsdk.Plan{Schema: s}
		if d := invalidPlan.Set(ctx, &invalid); d.HasError() {
			t.Fatal(d)
		}
		result := resource.CreateResponse{State: tfsdk.State{Schema: s}}
		r.Create(ctx, resource.CreateRequest{Plan: invalidPlan}, &result)
		if !result.Diagnostics.HasError() {
			t.Fatalf("expected invalid input rejection: %+v", input)
		}
	}
	deleted := resource.DeleteResponse{State: read.State}
	r.Delete(ctx, resource.DeleteRequest{State: read.State}, &deleted)
	if deleted.Diagnostics.HasError() || !revoked {
		t.Fatalf("delete: %v, revoked %v", deleted.Diagnostics, revoked)
	}
}

func TestIngestTokenModifyPlanExpiry(t *testing.T) {
	ctx := context.Background()
	r := &ingestTokenResource{}
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	s := schemaResp.Schema
	null := tftypes.NewValue(s.Type().TerraformType(ctx), nil)
	now := time.Now().UTC()
	expired := now.Add(-24 * time.Hour).Truncate(time.Second).Format(time.RFC3339)
	future := now.Add(24 * time.Hour).Truncate(time.Second).Format(time.RFC3339)
	tooLate := now.AddDate(0, 19, 0).Truncate(time.Second).Format(time.RFC3339)
	base := ingestTokenModel{FeedToken: types.StringValue("feed"), Name: types.StringValue("name"), ExpiresAt: types.StringValue(expired), RetentionDays: types.Int64Null()}
	makeState := func(m ingestTokenModel) tfsdk.State {
		state := tfsdk.State{Schema: s}
		if d := state.Set(ctx, &m); d.HasError() {
			t.Fatal(d)
		}
		return state
	}
	makePlan := func(m ingestTokenModel) tfsdk.Plan {
		plan := tfsdk.Plan{Schema: s}
		if d := plan.Set(ctx, &m); d.HasError() {
			t.Fatal(d)
		}
		return plan
	}
	with := func(expiry, name types.String) ingestTokenModel {
		m := base
		m.ExpiresAt = expiry
		m.Name = name
		return m
	}
	nullState := tfsdk.State{Schema: s, Raw: null}
	destroy := tfsdk.Plan{Schema: s, Raw: null}
	cases := []struct {
		name    string
		state   tfsdk.State
		plan    tfsdk.Plan
		invalid bool
	}{
		{"new expired", nullState, makePlan(base), true},
		{"new beyond lifetime", nullState, makePlan(with(types.StringValue(tooLate), base.Name)), true},
		{"new valid", nullState, makePlan(with(types.StringValue(future), base.Name)), false},
		{"replacement expired", makeState(base), makePlan(with(base.ExpiresAt, types.StringValue("new"))), true},
		{"replacement beyond lifetime", makeState(base), makePlan(with(types.StringValue(tooLate), base.Name)), true},
		{"unchanged expired", makeState(base), makePlan(base), false},
		{"destroy expired", makeState(base), destroy, false},
		{"unknown expiry", nullState, makePlan(with(types.StringUnknown(), base.Name)), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := resource.ModifyPlanResponse{Plan: tc.plan}
			r.ModifyPlan(ctx, resource.ModifyPlanRequest{State: tc.state, Plan: tc.plan, Config: tfsdk.Config{Schema: s, Raw: tc.plan.Raw}}, &resp)
			if resp.Diagnostics.HasError() != tc.invalid {
				t.Fatalf("diagnostics: %v", resp.Diagnostics)
			}
		})
	}
	configured := with(types.StringNull(), types.StringValue("new"))
	configuredPlan := makePlan(configured)
	replacement := makePlan(with(base.ExpiresAt, types.StringValue("new")))
	response := resource.ModifyPlanResponse{Plan: replacement}
	r.ModifyPlan(ctx, resource.ModifyPlanRequest{State: makeState(base), Plan: replacement, Config: tfsdk.Config{Schema: s, Raw: configuredPlan.Raw}}, &response)
	if response.Diagnostics.HasError() {
		t.Fatal(response.Diagnostics)
	}
	var result ingestTokenModel
	if d := response.Plan.Get(ctx, &result); d.HasError() {
		t.Fatal(d)
	}
	if !result.ExpiresAt.IsUnknown() {
		t.Fatalf("omitted expiry must be regenerated: %v", result.ExpiresAt)
	}

}

func TestIngestTokenMaxExpiryClampsCalendarMonth(t *testing.T) {
	for _, tc := range []struct{ issued, want string }{
		{"2024-08-31T12:34:56Z", "2026-02-28T12:34:56Z"},
		{"2026-08-31T12:34:56Z", "2028-02-29T12:34:56Z"},
		{"2026-01-30T12:34:56Z", "2027-07-30T12:34:56Z"},
	} {
		issued, err := time.Parse(time.RFC3339, tc.issued)
		if err != nil {
			t.Fatal(err)
		}
		if got := ingestTokenMaxExpiry(issued).Format(time.RFC3339); got != tc.want {
			t.Fatalf("%s: got %s, want %s", tc.issued, got, tc.want)
		}
	}
}
