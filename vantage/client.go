package vantage

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"
	"time"

	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	vantagev1 "github.com/vantage-sh/vantage-go/vantagev1/vantage"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
	vantagev2 "github.com/vantage-sh/vantage-go/vantagev2/vantage"
)

const userAgent = "tf-provider-vantage"

type Client struct {
	V1   *vantagev1.Vantage
	V2   *vantagev2.Vantage
	Auth runtime.ClientAuthInfoWriter
}

// timeoutTransport wraps a runtime.ClientTransport and sets the request timeout
// on every operation so the go-openapi default (30s) is overridden by the provider's configured timeout.
type timeoutTransport struct {
	inner   runtime.ClientTransport
	timeout time.Duration
}

func (t *timeoutTransport) Submit(operation *runtime.ClientOperation) (interface{}, error) {
	originalParams := operation.Params
	operation.Params = runtime.ClientRequestWriterFunc(func(req runtime.ClientRequest, reg strfmt.Registry) error {
		if err := originalParams.WriteToRequest(req, reg); err != nil {
			return err
		}
		return req.SetTimeout(t.timeout)
	})
	return t.inner.Submit(operation)
}

func NewClient(host, token string, debug bool, timeout time.Duration) (*Client, error) {
	parsedURL, err := url.Parse(host)
	if err != nil {
		return nil, err
	}

	v1Cfg := vantagev1.DefaultTransportConfig()
	v1Cfg.WithHost(parsedURL.Host)
	v1Cfg.WithSchemes([]string{parsedURL.Scheme})
	httpClientV1 := &http.Client{
		Timeout:   timeout,
		Transport: userAgentTripper(http.DefaultTransport, userAgent),
	}
	transportv1 := httptransport.NewWithClient(v1Cfg.Host, v1Cfg.BasePath, v1Cfg.Schemes, httpClientV1)
	transportv1.SetDebug(debug)
	v1 := vantagev1.New(&timeoutTransport{inner: transportv1, timeout: timeout}, strfmt.Default)

	v2Cfg := vantagev2.DefaultTransportConfig()
	v2Cfg.WithHost(parsedURL.Host)
	v2Cfg.WithSchemes([]string{parsedURL.Scheme})
	httpClientV2 := &http.Client{
		Timeout:   timeout,
		Transport: userAgentTripper(http.DefaultTransport, userAgent),
	}
	transportv2 := httptransport.NewWithClient(v2Cfg.Host, v2Cfg.BasePath, v2Cfg.Schemes, httpClientV2)
	transportv2.SetDebug(debug)
	v2 := vantagev2.New(&timeoutTransport{inner: transportv2, timeout: timeout}, strfmt.Default)

	bearerTokenAuth := httptransport.BearerToken(token)
	return &Client{
		V1:   v1,
		V2:   v2,
		Auth: bearerTokenAuth,
	}, nil
}

func userAgentTripper(inner http.RoundTripper, userAgent string) http.RoundTripper {
	version := "unknown"
	modified := false
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, s := range info.Settings {
			switch s.Key {
			case "vcs.revision":
				version = s.Value[:7]
			case "vcs.modified":
				modified = s.Value == "true"
			}
		}
	}
	agent := userAgent + "/" + version
	if modified {
		agent = agent + "+"
	}
	return &roundtripper{
		inner: inner,
		agent: agent,
	}
}

type roundtripper struct {
	inner http.RoundTripper
	agent string
}

func (ug *roundtripper) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("User-Agent", ug.agent)
	return ug.inner.RoundTrip(r)
}

// ptrStringOrEmpty returns a Terraform StringValue from a *string, falling back
// to an empty string (rather than null) when the pointer is nil. Use this for
// API fields that are semantically "optional empty string" so that the provider
// doesn't produce null where Terraform planned for "".
func ptrStringOrEmpty(s *string) types.String {
	if s == nil {
		return types.StringValue("")
	}
	return types.StringValue(*s)
}

func handleError(action string, d *diag.Diagnostics, err error) {
	d.AddError(
		fmt.Sprintf("Unable to %s", action),
		"An unexpected error occurred while attempting to contact the API. "+
			"Please retry the operation or report this issue to the provider developers.\n\n"+
			"Connection Error: "+err.Error(),
	)
}

func handleBadRequest(action string, d *diag.Diagnostics, mErr *modelsv2.Errors) {
	detail := "One or more of your fields contained invalid input."
	if mErr != nil {
		detail += "\n" + strings.Join(mErr.Errors, "\n")
	}
	d.AddError("Unable to "+action, detail)
}

func handleForbidden(action string, d *diag.Diagnostics, mErr *modelsv2.Errors) {
	message := "The API returned a 403 Forbidden response."
	if mErr != nil && len(mErr.Errors) > 0 {
		message = strings.Join(mErr.Errors, "\n")
	}
	d.AddError("Unable to "+action, message)
}

// apiNotFound selects how a 404 response is applied.
type apiNotFound int

const (
	// apiNotFoundError records a not-found diagnostic and keeps state.
	// Use this for create, update, and data source reads.
	apiNotFoundError apiNotFound = iota
	// apiNotFoundRemove drops the resource from state and records no diagnostic.
	// Use this for resource reads.
	apiNotFoundRemove
	// apiNotFoundIgnore treats 404 as success. Use this for deletes.
	apiNotFoundIgnore
)

type statusCoder interface {
	Code() int
}

type errorsPayload interface {
	GetPayload() *modelsv2.Errors
}

// handleAPIError classifies a Vantage API error and records one diagnostic.
// It returns true when err is non-nil and the caller must stop.
// remove runs only for apiNotFoundRemove.
//
// Every resource and data source uses this function so status codes mean the
// same thing everywhere:
//   - 400 and 422 are invalid input
//   - 402, 403, and 406 keep the API message
//   - 404 follows notFound
//   - every other error stays a connection error
func handleAPIError(action string, d *diag.Diagnostics, err error, notFound apiNotFound, remove func()) bool {
	if err == nil {
		return false
	}

	code, hasCode := apiErrorCode(err)
	payload := apiErrorPayload(err)
	switch code {
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		handleBadRequest(action, d, payload)
	case http.StatusForbidden:
		handleForbidden(action, d, payload)
	case http.StatusPaymentRequired:
		addAPIStatusError(action, d, "The API returned a 402 Payment Required response.", payload)
	case http.StatusNotAcceptable:
		addAPIStatusError(action, d, "The API returned a 406 Not Acceptable response.", payload)
	case http.StatusNotFound:
		switch notFound {
		case apiNotFoundIgnore:
			return true
		case apiNotFoundRemove:
			if remove != nil {
				remove()
				return true
			}
		}
		addAPIStatusError(action, d, "The API returned a 404 Not Found response.", payload)
	default:
		if hasCode && code >= 400 && code < 500 && payload != nil && len(payload.Errors) > 0 {
			addAPIStatusError(action, d, fmt.Sprintf("The API returned a %d response.", code), payload)
			return true
		}
		handleError(action, d, err)
	}
	return true
}

func apiErrorCode(err error) (int, bool) {
	var coder statusCoder
	if errors.As(err, &coder) {
		return coder.Code(), true
	}
	var apiErr *runtime.APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code, true
	}
	return 0, false
}

func apiErrorPayload(err error) *modelsv2.Errors {
	var payload errorsPayload
	if errors.As(err, &payload) {
		return payload.GetPayload()
	}
	return nil
}

func addAPIStatusError(action string, d *diag.Diagnostics, summary string, mErr *modelsv2.Errors) {
	if mErr != nil && len(mErr.Errors) > 0 {
		summary = strings.Join(mErr.Errors, "\n")
	}
	d.AddError("Unable to "+action, summary)
}

func toStringsValue(s []string) []basetypes.StringValue {
	out := []basetypes.StringValue{}
	for _, str := range s {
		out = append(out, types.StringValue(str))
	}

	return out
}

func fromStringsValue(s []types.String) []string {
	out := []string{}
	for _, str := range s {
		out = append(out, str.ValueString())
	}

	return out
}

func stringListFrom(in []string) (types.List, diag.Diagnostics) {
	values := make([]types.String, 0, len(in))
	for _, v := range in {
		values = append(values, types.StringValue(v))
	}
	return types.ListValueFrom(context.Background(), types.StringType, values)
}
