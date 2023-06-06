package vantage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type vantageClient struct {
	host  string
	token string
}

func newClient(host, token string) *vantageClient {
	return &vantageClient{
		host:  host,
		token: token,
	}
}

func (v *vantageClient) Ping() (string, error) {
	uri, err := url.JoinPath(v.host, "/v1/ping")
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", v.token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	out, err := io.ReadAll(resp.Body)
	return string(out), err
}

type AwsProviderInfoResult struct {
	ExternalID string `json:"external_id"`
	IamRoleARN string `json:"iam_role_arn"`
	Policies   struct {
		Root       string `json:"root"`
		Autopilot  string `json:"autopilot"`
		Cloudwatch string `json:"cloudwatch"`
		Resources  string `json:"resources"`
	} `json:"policies"`
}

func (v *vantageClient) AwsProviderInfo() (*AwsProviderInfoResult, error) {
	uri, err := url.JoinPath(v.host, "/v1/integrations/aws/info")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", v.token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	out := AwsProviderInfoResult{}

	switch resp.StatusCode {
	case http.StatusOK:
		err = json.NewDecoder(resp.Body).Decode(&out)
		return &out, err
	default:
		return nil, fmt.Errorf("failed to create provider credential: %d", resp.StatusCode)
	}
}

// AwsProviderResourceAPIModel describes the API data model.
type AwsProviderResourceAPIModel struct {
	Id              int `json:"id"`
	CrossAccountARN string `json:"cross_account_arn"`
	BucketARN       string `json:"bucket_arn"`
}

func (v *vantageClient) AddAwsProvider(in AwsProviderResourceAPIModel) (*AwsProviderResourceAPIModel, error) {
	uri, err := url.JoinPath(v.host, "/v1/integrations/aws")
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, uri, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", v.token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	out := AwsProviderResourceAPIModel{}
	switch resp.StatusCode {
	case http.StatusNotFound:
		return &out, err
	case http.StatusCreated:
		err = json.NewDecoder(resp.Body).Decode(&out)
		return &out, err
	default:
		return nil, fmt.Errorf("failed to create provider credential: %d", resp.StatusCode)
	}
}

func (v *vantageClient) UpdateAwsProvider(in AwsProviderResourceAPIModel) (*AwsProviderResourceAPIModel, error) {
	uri, err := url.JoinPath(v.host, fmt.Sprintf("/v1/integrations/aws/%d", in.Id))
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPut, uri, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", v.token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	out := AwsProviderResourceAPIModel{}
	switch resp.StatusCode {
	case http.StatusNotFound:
		return &out, err
	case http.StatusOK:
		err = json.NewDecoder(resp.Body).Decode(&out)
		return &out, err
	default:
		return nil, fmt.Errorf("failed to create provider credential: %d", resp.StatusCode)
	}
}

func (v *vantageClient) GetAwsProvider(id int) (*AwsProviderResourceAPIModel, error) {
	uri, err := url.JoinPath(v.host, fmt.Sprintf("/v1/integrations/aws/%d", id))
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", v.token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	out := AwsProviderResourceAPIModel{}
	switch resp.StatusCode {
	case http.StatusNotFound:
		return nil, err
	case http.StatusOK:
		err = json.NewDecoder(resp.Body).Decode(&out)
		return &out, err
	default:
		return nil, fmt.Errorf("failed to fetch provider credential: %d", resp.StatusCode)
	}
}

func (v *vantageClient) DeleteAwsProvider(id int) error {
	uri, err := url.JoinPath(v.host, fmt.Sprintf("/v1/integrations/aws/%s", id))
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodDelete, uri, nil)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", v.token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	switch resp.StatusCode {
	case http.StatusNoContent:
		return nil
	default:
		return fmt.Errorf("failed to delete provider credential: %d", resp.StatusCode)
	}
}
