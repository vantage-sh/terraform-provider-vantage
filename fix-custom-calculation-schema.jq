# Generate list schemas; CRUD helpers must also normalize scalar API readback.
def usage_unit_list:
  . as $source
  | if .type == null or .type == "object" or .type == "array"
    then . else error("Unexpected usage-unit outer type") end
  | (.oneOf // ."x-oneOf" // []) as $branches
  | if ($branches | all(.type == "string" or .type == "array" or .type == "null"))
    then . else error("Unexpected usage-unit union type") end
  | (if .type == "array" and ($branches | length) == 0 then . else
      [$branches[] | select(.type == "array")]
      | if length == 1 then .[0] else error("Expected one usage-unit array schema") end
    end) as $array
  | if $array.items.type != "string" then error("Expected usage-unit string items") else . end
  | if $source.items != null and ($array.items * $source.items) != ($source.items * $array.items)
    then error("Conflicting usage-unit item schemas") else . end
  | ($array * $source)
  | if .items.type != "string" then error("Expected usage-unit string items") else . end
  | del(.oneOf, ."x-oneOf", .properties)
  | .type = "array"
  | .nullable = true
  | .minItems = ([1, ($source.minItems // 0), ($array.minItems // 0)] | max)
  | if .minItems > 1 and ($branches | any(.type == "string"))
    then error("Cannot represent scalar usage units with an array minimum above one") else . end
  | ([$source.maxItems, $array.maxItems] | map(select(. != null))) as $maximums
  | if ($maximums | length) > 0 then .maxItems = ($maximums | min) else . end;

def filter_lists:
  if (.properties | type) == "object" then
    .properties |= with_entries(
      if .key == "usage_unit" or .key == "usage_unit_numerator" or .key == "usage_unit_denominator"
      then .value |= usage_unit_list else . end
    )
  else . end;

# Attachment tokens are unordered and de-duplicated by the API.
def attachment_set($name):
  if .properties[$name]? then
    .properties[$name] |= (
      if .type == "array" and .items.type == "string" then .format = "set"
      else error("Expected attachment token string array") end
    )
  else . end;

if (.components.schemas | type) == "object" then
  .components.schemas |= with_entries(
    if .key == "CustomCalculationFilter" then .value |= filter_lists
    elif .key == "CustomCalculation" or .key == "createCustomCalculation" or .key == "updateCustomCalculation"
    then .value |= attachment_set("cost_report_tokens")
    elif .key == "CostReport" or .key == "createCostReport" or .key == "updateCostReport"
    then .value |= attachment_set("custom_calculation_tokens")
    else . end
  )
else . end
