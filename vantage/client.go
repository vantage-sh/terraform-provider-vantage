package vantage

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"

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

func NewClient(host, token string, debug bool) (*Client, error) {
	parsedURL, err := url.Parse(host)
	if err != nil {
		return nil, err
	}

	v1Cfg := vantagev1.DefaultTransportConfig()
	v1Cfg.WithHost(parsedURL.Host)
	v1Cfg.WithSchemes([]string{parsedURL.Scheme})
	transportv1 := httptransport.New(v1Cfg.Host, v1Cfg.BasePath, v1Cfg.Schemes)
	transportv1.Transport = userAgentTripper(transportv1.Transport, userAgent)
	transportv1.SetDebug(debug)
	v1 := vantagev1.New(transportv1, strfmt.Default)

	v2Cfg := vantagev2.DefaultTransportConfig()
	v2Cfg.WithHost(parsedURL.Host)
	v2Cfg.WithSchemes([]string{parsedURL.Scheme})
	transportv2 := httptransport.New(v2Cfg.Host, v2Cfg.BasePath, v2Cfg.Schemes)
	transportv2.Transport = userAgentTripper(transportv2.Transport, userAgent)
	transportv2.SetDebug(debug)
	v2 := vantagev2.New(transportv2, strfmt.Default)

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

func handleError(action string, d *diag.Diagnostics, err error) {
	d.AddError(
		fmt.Sprintf("Unable to %s", action),
		"An unexpected error occurred while attempting to contact the API. "+
			"Please retry the operation or report this issue to the provider developers.\n\n"+
			"Connection Error: "+err.Error(),
	)
}

func handleBadRequest(action string, d *diag.Diagnostics, mErr *modelsv2.Errors) {
	d.AddError(
		"Unable to "+action,
		"One or more of your fields contained invalid input.\n"+strings.Join(mErr.Errors, "\n"),
	)
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
