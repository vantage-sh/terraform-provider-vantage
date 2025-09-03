def add_id_after_token:
  if type == "array" then
    map(
      if .name == "token" then
        [., {
          "name": "id",
          "string": {
            "computed_optional_required": "computed",
            "description": ("The id of the " + (if .string.description and (.string.description | type) == "string" then (.string.description | split(" of the ")[1] // "resource") else "resource" end))
          }
        }]
      else
        [.]
      end
    ) | flatten
  else
    .
  end;

def process_nested_attributes:
  if has("list_nested") and .list_nested.nested_object.attributes then
    .list_nested.nested_object.attributes |= add_id_after_token
  else
    .
  end;

(.resources[]?) |= 
  if has("schema") and .schema.attributes and (.schema.attributes != null) then
    .schema.attributes |= add_id_after_token
  else
    .
  end |

(.datasources[]?) |= 
  if has("schema") and .schema.attributes and (.schema.attributes != null) then
    .schema.attributes |= map(process_nested_attributes)
  else
    .
  end
