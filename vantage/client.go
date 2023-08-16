package vantage

import (
	"net/url"

	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	vantagev1 "github.com/vantage-sh/vantage-go/vantagev1/vantage"
)

type Client struct {
	V1   *vantagev1.Vantage
	Auth runtime.ClientAuthInfoWriter
}

func NewClient(host, token string) (*Client, error) {
	parsedURL, err := url.Parse(host)
	if err != nil {
		return nil, err
	}

	transport := vantagev1.DefaultTransportConfig()
	transport.WithHost(parsedURL.Host)
	transport.WithSchemes([]string{parsedURL.Scheme})
	v1 := vantagev1.NewHTTPClientWithConfig(strfmt.Default, transport)
	bearerTokenAuth := httptransport.BearerToken(token)
	return &Client{
		V1:   v1,
		Auth: bearerTokenAuth,
	}, nil
}

func strPtr(str string) *string {
	return &str
}
