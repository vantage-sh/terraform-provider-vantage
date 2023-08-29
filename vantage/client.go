package vantage

import (
	"fmt"
	"net/url"

	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	vantagev1 "github.com/vantage-sh/vantage-go/vantagev1/vantage"
	vantagev2 "github.com/vantage-sh/vantage-go/vantagev2/vantage"
)

type Client struct {
	V1   *vantagev1.Vantage
	V2   *vantagev2.Vantage
	Auth runtime.ClientAuthInfoWriter
}

func NewClient(host, token string) (*Client, error) {
	parsedURL, err := url.Parse(host)
	if err != nil {
		return nil, err
	}

	transportv1 := vantagev1.DefaultTransportConfig()
	transportv1.WithHost(parsedURL.Host)
	transportv1.WithSchemes([]string{parsedURL.Scheme})
	v1 := vantagev1.NewHTTPClientWithConfig(strfmt.Default, transportv1)

	transportv2 := vantagev2.DefaultTransportConfig()
	transportv2.WithHost(parsedURL.Host)
	transportv2.WithSchemes([]string{parsedURL.Scheme})
	v2 := vantagev2.NewHTTPClientWithConfig(strfmt.Default, transportv2)
	bearerTokenAuth := httptransport.BearerToken(token)
	return &Client{
		V1:   v1,
		V2:   v2,
		Auth: bearerTokenAuth,
	}, nil
}

func handleError(action string, d *diag.Diagnostics, err error) {
	d.AddError(
		fmt.Sprintf("Unable to %s", action),
		"An unexpected error occurred while attempting to contact the API. "+
			"Please retry the operation or report this issue to the provider developers.\n\n"+
			"Connection Error: "+err.Error(),
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
