# grape-swagger currently emits map[string][]string fields as unstructured
# objects. Restore the missing value schema before Terraform code generation so
# label_filters is generated as map(list(string)), not an empty nested object.
walk(
  if type == "object"
    and (.properties? | type) == "object"
    and (.properties.label_filters? | type) == "object"
    and .properties.label_filters.type? == "object"
  then
    .properties.label_filters.additionalProperties = {
      "type": "array",
      "items": {
        "type": "string"
      }
    }
  else
    .
  end
)
