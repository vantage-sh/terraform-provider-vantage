def add_id_after_token:
  if type == "array" then
    map(
      if .name == "token" then
        [., {
          "name": "id",
          "string": {
            "computed_optional_required": "computed",
            "description": ("The id of the " + (.string.description | split(" of the ")[1] // "resource"))
          }
        }]
      else
        [.]
      end
    ) | flatten
  else
    .
  end;

(.resources[]?, .datasources[]?) |= 
  if has("schema") and .schema.attributes and (.schema.attributes != null) then
    .schema.attributes |= add_id_after_token
  else
    .
  end
