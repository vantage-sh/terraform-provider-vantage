package vantage

import (
	"net/http"
	"strings"
	"testing"

	"github.com/go-openapi/runtime"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
	costalertsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/cost_alerts"
)

type fakeAPIError struct {
	code    int
	payload *modelsv2.Errors
	wrapped error
}

func (e *fakeAPIError) Error() string {
	if e.wrapped != nil {
		return e.wrapped.Error()
	}
	return "fake api error"
}

func (e *fakeAPIError) Unwrap() error { return e.wrapped }

func (e *fakeAPIError) Code() int { return e.code }

func (e *fakeAPIError) GetPayload() *modelsv2.Errors { return e.payload }

func TestHandleAPIErrorUsesGeneratedSDKTypes(t *testing.T) {
	notFound := costalertsv2.NewGetCostAlertNotFound()
	notFound.Payload = &modelsv2.Errors{Errors: []string{"cost alert not found"}}

	var diags diag.Diagnostics
	removed := false
	if !handleAPIError("Read Cost Alert", &diags, notFound, apiNotFoundRemove, func() {
		removed = true
	}) {
		t.Fatal("expected the generated 404 to stop the caller")
	}
	if !removed {
		t.Fatal("expected read 404 to remove state")
	}
	if diags.HasError() {
		t.Fatalf("read 404 recorded a diagnostic: %v", diags)
	}

	badRequest := costalertsv2.NewCreateCostAlertBadRequest()
	badRequest.Payload = &modelsv2.Errors{Errors: []string{"title is required"}}
	diags = diag.Diagnostics{}
	if !handleAPIError("Create Cost Alert", &diags, badRequest, apiNotFoundError, nil) {
		t.Fatal("expected the generated 400 to stop the caller")
	}
	if len(diags) != 1 || !strings.Contains(diags[0].Detail(), "title is required") {
		t.Fatalf("bad request diagnostic = %v", diags)
	}
}

func TestHandleAPIError(t *testing.T) {
	payload := &modelsv2.Errors{Errors: []string{"name is required", "workspace is missing"}}

	tests := []struct {
		name       string
		err        error
		notFound   apiNotFound
		wantStop   bool
		wantRemove bool
		wantDetail string
	}{
		{
			name:     "nil error",
			notFound: apiNotFoundError,
		},
		{
			name:       "bad request",
			err:        &fakeAPIError{code: http.StatusBadRequest, payload: payload},
			notFound:   apiNotFoundError,
			wantStop:   true,
			wantDetail: "invalid input",
		},
		{
			name:       "unprocessable entity",
			err:        &fakeAPIError{code: http.StatusUnprocessableEntity, payload: payload},
			notFound:   apiNotFoundError,
			wantStop:   true,
			wantDetail: "name is required",
		},
		{
			name:       "forbidden uses payload",
			err:        &fakeAPIError{code: http.StatusForbidden, payload: payload},
			notFound:   apiNotFoundError,
			wantStop:   true,
			wantDetail: "workspace is missing",
		},
		{
			name:       "forbidden without payload",
			err:        &fakeAPIError{code: http.StatusForbidden},
			notFound:   apiNotFoundError,
			wantStop:   true,
			wantDetail: "403 Forbidden",
		},
		{
			name:       "payment required",
			err:        &fakeAPIError{code: http.StatusPaymentRequired, payload: payload},
			notFound:   apiNotFoundError,
			wantStop:   true,
			wantDetail: "name is required",
		},
		{
			name:       "not acceptable",
			err:        &fakeAPIError{code: http.StatusNotAcceptable, payload: &modelsv2.Errors{Errors: []string{"bad accept"}}},
			notFound:   apiNotFoundError,
			wantStop:   true,
			wantDetail: "bad accept",
		},
		{
			name:       "create not found keeps a diagnostic",
			err:        &fakeAPIError{code: http.StatusNotFound, payload: payload},
			notFound:   apiNotFoundError,
			wantStop:   true,
			wantDetail: "name is required",
		},
		{
			name:       "read not found removes state",
			err:        &fakeAPIError{code: http.StatusNotFound, payload: payload},
			notFound:   apiNotFoundRemove,
			wantStop:   true,
			wantRemove: true,
		},
		{
			name:     "delete not found is success",
			err:      &fakeAPIError{code: http.StatusNotFound, payload: payload},
			notFound: apiNotFoundIgnore,
			wantStop: true,
		},
		{
			name:       "generic client error with payload",
			err:        &fakeAPIError{code: http.StatusConflict, payload: payload},
			notFound:   apiNotFoundError,
			wantStop:   true,
			wantDetail: "name is required",
		},
		{
			name:       "runtime api error not found",
			err:        runtime.NewAPIError("[GET /widgets] getWidget", "missing", http.StatusNotFound),
			notFound:   apiNotFoundRemove,
			wantStop:   true,
			wantRemove: true,
		},
		{
			name:       "server error stays a connection error",
			err:        &fakeAPIError{code: http.StatusInternalServerError},
			notFound:   apiNotFoundError,
			wantStop:   true,
			wantDetail: "Connection Error",
		},
		{
			name:       "uncoded error stays a connection error",
			err:        &fakeAPIError{code: 0},
			notFound:   apiNotFoundIgnore,
			wantStop:   true,
			wantDetail: "Connection Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			removed := false
			stop := handleAPIError("Create Widget", &diags, tt.err, tt.notFound, func() {
				removed = true
			})
			if stop != tt.wantStop {
				t.Fatalf("stop = %v, want %v", stop, tt.wantStop)
			}
			if removed != tt.wantRemove {
				t.Fatalf("removed = %v, want %v", removed, tt.wantRemove)
			}
			if tt.wantDetail == "" {
				if diags.HasError() {
					t.Fatalf("unexpected diagnostics: %v", diags)
				}
				return
			}
			if !diags.HasError() {
				t.Fatal("expected a diagnostic")
			}
			detail := diags[0].Detail()
			if !strings.Contains(detail, tt.wantDetail) {
				t.Fatalf("detail %q does not contain %q", detail, tt.wantDetail)
			}
			if !strings.Contains(diags[0].Summary(), "Unable to Create Widget") {
				t.Fatalf("summary = %q", diags[0].Summary())
			}
		})
	}
}
