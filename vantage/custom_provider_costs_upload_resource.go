package vantage

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	integrationsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/integrations"
)

var (
	_ resource.Resource                = (*CustomProviderCostsUploadResource)(nil)
	_ resource.ResourceWithConfigure   = (*CustomProviderCostsUploadResource)(nil)
	_ resource.ResourceWithImportState = (*CustomProviderCostsUploadResource)(nil)
)

type CustomProviderCostsUploadResource struct{ client *Client }

func NewCustomProviderCostsUploadResource() resource.Resource {
	return &CustomProviderCostsUploadResource{}
}

type CustomProviderCostsUploadResourceModel struct {
	IntegrationToken types.String `tfsdk:"integration_token"`
	CsvContent       types.String `tfsdk:"csv_content"`
	AutoTransform    types.Bool   `tfsdk:"auto_transform"`
	Token            types.String `tfsdk:"token"`
	Id               types.String `tfsdk:"id"`
	ImportStatus     types.String `tfsdk:"import_status"`
	StartDate        types.String `tfsdk:"start_date"`
	EndDate          types.String `tfsdk:"end_date"`
	Amount           types.String `tfsdk:"amount"`
	Filename         types.String `tfsdk:"filename"`
}

func (r *CustomProviderCostsUploadResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*Client)
}

func (r *CustomProviderCostsUploadResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_provider_costs_upload"
}

func (r *CustomProviderCostsUploadResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"integration_token": schema.StringAttribute{
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				MarkdownDescription: "The token of the Custom Provider integration to upload costs for.",
			},
			"csv_content": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				MarkdownDescription: "CSV content to upload as costs data.",
			},
			"auto_transform": schema.BoolAttribute{
				Optional:            true,
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
				MarkdownDescription: "When true, attempts to automatically transform the CSV to match the FOCUS format.",
			},
			"token": schema.StringAttribute{
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "Unique token of the costs upload.",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "Same as token.",
			},
			"import_status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The import status of the upload (e.g. processing, complete, error).",
			},
			"start_date": schema.StringAttribute{
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "The start date of the costs in the upload.",
			},
			"end_date": schema.StringAttribute{
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "The end date of the costs in the upload.",
			},
			"amount": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The total amount of costs in the upload.",
			},
			"filename": schema.StringAttribute{
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "The filename of the uploaded costs file.",
			},
		},
		MarkdownDescription: "Uploads a CSV of costs for a Custom Provider integration.",
	}
}

func (r *CustomProviderCostsUploadResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("token"), req, resp)
}

func (r *CustomProviderCostsUploadResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CustomProviderCostsUploadResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	csvReader := runtime.NamedReader("costs.csv", strings.NewReader(data.CsvContent.ValueString()))

	params := integrationsv2.NewCreateUserCostsUploadViaCsvParams()
	params.SetIntegrationToken(data.IntegrationToken.ValueString())
	params.SetCsv(csvReader)

	if !data.AutoTransform.IsNull() && !data.AutoTransform.IsUnknown() {
		v := data.AutoTransform.ValueBool()
		params.SetAutoTransform(&v)
	}

	out, err := r.client.V2.Integrations.CreateUserCostsUploadViaCsv(params, r.client.Auth, integrationsv2.WithContentTypeMultipartFormData)
	if err != nil {
		handleError("Create Custom Provider Costs Upload", &resp.Diagnostics, err)
		return
	}

	data.Token = types.StringValue(out.Payload.Token)
	data.Id = types.StringValue(out.Payload.Token)
	data.ImportStatus = types.StringValue(out.Payload.ImportStatus)
	data.StartDate = types.StringValue(out.Payload.StartDate)
	data.EndDate = types.StringValue(out.Payload.EndDate)
	data.Amount = types.StringValue(out.Payload.Amount)
	data.Filename = types.StringValue(out.Payload.Filename)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CustomProviderCostsUploadResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// There is no individual GET endpoint for a costs upload; preserve existing state.
	var state CustomProviderCostsUploadResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *CustomProviderCostsUploadResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All fields are RequiresReplace, so Update is never called in practice.
	var plan CustomProviderCostsUploadResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CustomProviderCostsUploadResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CustomProviderCostsUploadResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The SDK incorrectly types the user_costs_upload_token path parameter as
	// int32. The API actually accepts a string. We bypass the SDK by submitting
	// a raw ClientOperation through the existing transport so that auth,
	// timeouts, and user-agent headers are all preserved.
	op := &runtime.ClientOperation{
		ID:                 "deleteUserCostsUpload",
		Method:             "DELETE",
		PathPattern:        "/integrations/{integration_token}/costs/{user_costs_upload_token}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"https"},
		Params: &deleteUploadStringParams{
			integrationToken:     state.IntegrationToken.ValueString(),
			userCostsUploadToken: state.Token.ValueString(),
		},
		Reader:   &deleteUploadReader{},
		AuthInfo: r.client.Auth,
	}

	if _, err := r.client.V2.Transport.Submit(op); err != nil {
		handleError("Delete Custom Provider Costs Upload", &resp.Diagnostics, err)
	}
}

// deleteUploadStringParams writes both path parameters as plain strings,
// bypassing the SDK's incorrect int32 type for user_costs_upload_token.
type deleteUploadStringParams struct {
	integrationToken     string
	userCostsUploadToken string
}

func (p *deleteUploadStringParams) WriteToRequest(r runtime.ClientRequest, _ strfmt.Registry) error {
	if err := r.SetPathParam("integration_token", p.integrationToken); err != nil {
		return err
	}
	return r.SetPathParam("user_costs_upload_token", p.userCostsUploadToken)
}

// deleteUploadReader accepts a 204 No Content or 200 OK response from the
// delete endpoint and returns an error for anything else.
type deleteUploadReader struct{}

func (r *deleteUploadReader) ReadResponse(response runtime.ClientResponse, _ runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200, 204:
		return nil, nil
	case 404:
		return nil, nil // already gone — treat as success
	default:
		return nil, fmt.Errorf("unexpected status %d from delete costs upload", response.Code())
	}
}

