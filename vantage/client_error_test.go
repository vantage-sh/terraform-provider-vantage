package vantage

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/go-openapi/runtime"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// stubClientResponse is the response shape the generated client hands to
// runtime.NewAPIError for a status it does not model.
type stubClientResponse struct {
	code int
	body io.ReadCloser
}

func (r stubClientResponse) Code() int                    { return r.code }
func (r stubClientResponse) Message() string              { return "" }
func (r stubClientResponse) GetHeader(_ string) string    { return "" }
func (r stubClientResponse) GetHeaders(_ string) []string { return nil }
func (r stubClientResponse) Body() io.ReadCloser          { return r.body }

func TestHandleErrorReportsAPIResponseBody(t *testing.T) {
	body := `{"errors":["Duration in days can't be blank"]}`
	err := runtime.NewAPIError(
		"[PUT /budget_alerts/{budget_alert_token}] updateBudgetAlert",
		stubClientResponse{code: 400, body: io.NopCloser(strings.NewReader(body))},
		400,
	)

	var diags diag.Diagnostics
	handleError("Update Budget Alert", &diags, err)

	if !diags.HasError() {
		t.Fatal("expected an error diagnostic")
	}
	detail := diags.Errors()[0].Detail()
	if !strings.Contains(detail, "status 400") {
		t.Errorf("detail does not report the status: %q", detail)
	}
	if !strings.Contains(detail, "Duration in days can't be blank") {
		t.Errorf("detail does not carry the API response body: %q", detail)
	}
}

// TestHandleErrorWithoutBody checks that an APIError whose body cannot be read
// still reports the status rather than printing an empty object.
func TestHandleErrorWithoutBody(t *testing.T) {
	err := runtime.NewAPIError("[PUT /x] op", stubClientResponse{code: 500}, 500)

	var diags diag.Diagnostics
	handleError("Update Budget Alert", &diags, err)

	detail := diags.Errors()[0].Detail()
	if !strings.Contains(detail, "status 500") {
		t.Errorf("detail does not report the status: %q", detail)
	}
	if strings.Contains(detail, "{}") {
		t.Errorf("detail still prints the empty response object: %q", detail)
	}
}

// TestHandleErrorPlainError keeps the original wording for a transport failure,
// which really is a connection problem.
func TestHandleErrorPlainError(t *testing.T) {
	var diags diag.Diagnostics
	handleError("Update Budget Alert", &diags, errors.New("dial tcp: connection refused"))

	detail := diags.Errors()[0].Detail()
	if !strings.Contains(detail, "Connection Error: dial tcp: connection refused") {
		t.Errorf("detail lost the underlying error: %q", detail)
	}
}
