
generate:
	$(eval VANTAGE_HOST ?= https://api.vantage.sh)
	@echo "Downloading OAS3 version of our swagger file"
	curl "${VANTAGE_HOST}/v2/swagger.json" > tmp-swagger.json
	cat tmp-swagger.json | curl --header 'Content-Type: application/json' --data @'-' https://converter.swagger.io/api/convert | jq > tmp.json
	@echo "Generating spec from OAS3 swagger file"
	tfplugingen-openapi generate --config generator.yaml --output spec.json tmp.json
	jq -f add-ids.jq spec.json > spec.json.tmp && mv spec.json.tmp spec.json
	@echo "Generating Terraform code from OpenAPI spec"
	tfplugingen-framework generate resources --input spec.json --output vantage
	tfplugingen-framework generate data-sources --input spec.json --output vantage
	rm spec.json tmp-swagger.json tmp.json

scaffold-resource:
ifndef RESOURCE
	$(error RESOURCE is not set)
endif
	tfplugingen-framework scaffold resource --output-dir vantage --package vantage \
	  --name ${RESOURCE} \
	  --force

# DATASOURCE should be plural (ie "folders")
scaffold-datasource:
ifndef DATASOURCE
	$(error DATASOURCE is not set)
endif
	tfplugingen-framework scaffold data-source --output-dir vantage --package vantage \
	  --name ${DATASOURCE} \
	  --force

.PHONY: docs
docs:
	go generate ./...

test:
	go test -v ./... -count=1
