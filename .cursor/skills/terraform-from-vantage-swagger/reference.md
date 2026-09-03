# Vantage Swagger-to-Terraform Reference

## Fixed response object

Use a dedicated entity when keys are known:

```ruby
class DefaultForecast < Vantage::V1::Entities::Base
  def self.entity_name
    "DefaultForecast"
  end

  expose_as_required(
    :kind,
    nullable: false,
    documentation: {values: %w[baseline report_forecast]}
  )
  expose :report_forecast_token
end
```

Expose it from the parent:

```ruby
expose_as_required(
  :default_forecast,
  nullable: false,
  using: Entities::DefaultForecast
) do |report, _|
  report.default_forecast_selection
end
```

Expected Swagger:

```json
{
  "$ref": "#/definitions/DefaultForecast"
}
```

Expected go-swagger result:

```go
DefaultForecast *DefaultForecast `json:"default_forecast"`
```

## Dynamic map response

Use `additional_properties` when keys are data:

```ruby
documentation: {
  type: Hash,
  additional_properties: {
    type: "array",
    items: {type: "string"},
    minItems: 1
  }
}
```

Expected Swagger:

```json
{
  "type": "object",
  "additionalProperties": {
    "type": "array",
    "items": {"type": "string"},
    "minItems": 1
  }
}
```

Expected Go:

```go
LabelFilters map[string][]string `json:"label_filters,omitempty"`
```

`grape-swagger-entity` may discard `additional_properties` from response
entities. Verify the actual Swagger output. In Core, preserve this metadata in
the response entity attribute parser rather than accepting `interface{}`.

## Optional request field

When a generated client must omit an optional request key:

```ruby
documentation: {
  x: {omitempty: true}
}
```

Verify the generated tag:

```go
LabelFilter []string `json:"label_filter,omitempty"`
```

This matters when two fields are mutually exclusive: a serialized `null` can
still count as present to request validation.

## Applying a generated nested object

```go
value, d := resource_example.NewNestedValue(
  resource_example.NestedValue{}.AttributeTypes(ctx),
  map[string]attr.Value{
    "kind":  types.StringValue(payload.Kind),
    "token": types.StringPointerValue(payload.Token),
  },
)
diags.Append(d...)
if d.HasError() {
  return diags
}
state.Nested = value
```

Explicitly set the generated null value when the API returns nil:

```go
state.Nested = resource_example.NewNestedValueNull()
```

## Typed map conversion

API → Terraform:

```go
mapValue := types.MapNull(types.ListType{ElemType: types.StringType})
if payload.LabelFilters != nil {
  mapValue, d = types.MapValueFrom(
    ctx,
    types.ListType{ElemType: types.StringType},
    payload.LabelFilters,
  )
  diags.Append(d...)
}
```

Terraform → API:

```go
items := map[string][]string{}
d := value.(types.Map).ElementsAs(ctx, &items, false)
diags.Append(d...)
if d.HasError() {
  return request
}
request.LabelFilters = items
```

## Update-only field after create

```go
created, err := client.Create(createParams, auth)
// handle error

payload := created.Payload
if fieldConfigured {
  updateParams := NewUpdateParams().
    WithToken(created.Payload.Token).
    WithBody(updateModel)
  updated, err := client.Update(updateParams, auth)
  // handle error
  payload = updated.Payload
}

diags.Append(state.applyPayload(ctx, payload)...)
```

## Acceptance test shape

```go
Steps: []resource.TestStep{
  {
    Config: config("initial"),
    Check: resource.ComposeTestCheckFunc(
      resource.TestCheckResourceAttr(resourceName, "field", "initial"),
    ),
  },
  {
    Config: config("updated"),
    Check: resource.ComposeTestCheckFunc(
      resource.TestCheckResourceAttr(resourceName, "field", "updated"),
    ),
  },
  {
    Config:             config("updated"),
    PlanOnly:           true,
    ExpectNonEmptyPlan: false,
  },
}
```
