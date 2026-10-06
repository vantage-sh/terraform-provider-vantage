# TipTap widget content is free-form JSON (nested empty objects in OAS).
# Represent it as a JSON string in Terraform so tfplugingen-framework does not
# emit colliding ContentType/ContentValue declarations for nested nodes.
def content_string:
  {
    type: "string",
    description: "JSON-encoded TipTap document for a free text widget. Example: {\"type\":\"doc\",\"content\":[...]}",
    nullable: true
  };

def rewrite_widget_content:
  if type == "object" and has("properties") and (.properties|type=="object") and (.properties|has("content")) then
    .properties.content = content_string
  else
    .
  end;

.components.schemas.DashboardWidget |= rewrite_widget_content
| .components.schemas.createDashboard.properties.widgets.items |= rewrite_widget_content
| .components.schemas.updateDashboard.properties.widgets.items |= rewrite_widget_content
