generate:
	@echo "Downloading OAS3 version of our swagger file"
	curl https://converter.swagger.io/api/convert\?url\=https://api.vantage.sh/v2/swagger.json | jq > tmp.json
	@echo "Generating spec from OAS3 swagger file"
	tfplugingen-openapi generate --config generator.yaml --output spec.json tmp.json
	@echo "Generating Terraform code from OpenAPI spec"
	tfplugingen-framework generate resources --input spec.json --output vantage
	tfplugingen-framework generate data-sources --input spec.json --output vantage
	rm tmp.json
