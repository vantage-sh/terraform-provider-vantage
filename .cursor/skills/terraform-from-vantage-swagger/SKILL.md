---
name: terraform-from-vantage-swagger
description: Builds or updates Vantage Terraform resources and data sources from Core Grape Swagger endpoints, including fixing response schemas, regenerating and releasing vantage-go, wiring provider models, and adding drift-safe acceptance tests. Use when adding API fields, resources, data sources, or regenerating Terraform schemas from Vantage Swagger.
---

# Terraform from Vantage Swagger

Use this workflow across `vantage-sh/core`, `vantage-sh/vantage-go`, and
`vantage-sh/terraform-provider-vantage`. Do not skip a layer: generated Terraform
types are only as accurate as Core Swagger and vantage-go.

Read [reference.md](reference.md) for concrete schema and model-mapping patterns.

## 1. Establish scope

1. Identify the API endpoint and supported operations: create, read, update,
   delete, and list.
2. Compare every request and response field, including nested properties,
   enums, nullability, defaults, arrays, maps, and conditional requirements.
3. Decide whether the task is:
   - resource only;
   - data source only; or
   - both.
4. Create dedicated worktrees for every repository that will change.

## 2. Validate Core Swagger first

Inspect the live or locally generated `/v2/swagger.json`, not just endpoint Ruby.

For every field, verify:

- fixed objects have `properties` or a `$ref`;
- dynamic maps have typed `additionalProperties`;
- optional request fields generate `x-omitempty: true` when omission matters;
- enums and required fields are represented;
- create, update, and read definitions agree where they should;
- arrays identify their item type and minimum size.

Do not proceed with `interface{}`-shaped response fields. Fix Core first.

### Core response rules

- Fixed nested object: define a `Grape::Entity` and expose it with `using:`.
- Dynamic map: document `additional_properties` with its value schema.
- Add a Swagger regression test that fetches `/v2/swagger.json` and asserts the
  exact `$ref`, properties, enum, or `additionalProperties`.
- Preserve the runtime JSON response shape.

Run:

```bash
PARALLEL_WORKERS=0 bundle exec rails test <relevant-api-test>
bundle exec rubocop <changed-ruby-files>
bundle exec codeownership validate
```

Generate Swagger locally through the Rails app and inspect it before touching
vantage-go:

```ruby
response = Rack::MockRequest.new(Rails.application).get("/v2/swagger.json")
raise "swagger status #{response.status}" unless response.status == 200
File.write("/tmp/vantage-swagger.json", response.body)
```

Merge Core changes before publishing the SDK. If multiple Core fixes are
required, wait for all of them.

## 3. Regenerate and release vantage-go

Regenerate from the merged Core Swagger, not a stale public deployment. Serve a
local Swagger file on a random free port and set `VANTAGE_HOST` for generation.
Never use port 3000.

After generation, inspect generated Go structs and JSON tags:

- no unexpected `interface{}`;
- pointer vs value semantics match nullability;
- optional fields use `omitempty`;
- map types are concrete, for example `map[string][]string`;
- required nested objects are concrete pointers;
- enum constants and validators exist.

Run:

```bash
go test ./...
```

Create a vantage-go PR and wait for it to merge. Then:

1. verify the requested version tag does not exist;
2. tag the merged commit;
3. create the GitHub release with generated notes;
4. verify `go list -m` can resolve the new version.

Never tag an unmerged feature commit.

## 4. Configure provider generation

Update `generator.yaml`:

- add resource CRUD paths under `resources`;
- add list/read paths under `data_sources`;
- add token aliases and intentional ignores only;
- remove ignores once the SDK supports those fields.

Run:

```bash
make generate
```

Generated `resource_*` and `datasource_*` packages must not be hand-edited.
Revert collateral generated changes outside the requested resource/data source,
then verify regeneration is reproducible.

If a generated data-source package is not imported by the handwritten data
source, remove it from `generator.yaml` instead of committing dead code.

## 5. Wire the resource model

Every configurable field needs all three paths:

1. `applyPayload`: API response → Terraform state;
2. `toCreateModel`: Terraform plan → create request;
3. `toUpdateModel`: Terraform plan → update request.

Handle SDK diagnostics from `ElementsAs`, `MapValueFrom`, and generated value
constructors. Never discard conversion errors.

### Drift rules

- API `nil` must clear stale state.
- Preserve null when the config omitted Optional+Computed nested objects if
  capturing API defaults would cause drift.
- Do not retain plan/state merely because the API omitted a field.
- Preserve empty-array behavior required by the API, except when an empty
  sibling field violates mutual exclusion.
- Do not introduce generated defaults that overwrite existing remote values.
- Use `UseStateForUnknown` only where preserving known state is intentional.

### Update-only API fields

If a field is accepted by update but not create:

1. override the generated schema to make it configurable;
2. create the resource;
3. immediately perform a follow-up update when the field is configured;
4. apply state from the update response;
5. test both initial creation and later updates.

### Generated nested values

List elements may be generated custom values rather than `types.Object`. Handle
the generated value type explicitly and optionally support `types.Object` for
unit tests. Never use an unchecked type assertion.

## 6. Wire data sources separately

Do not reuse resource models for data sources unless their schemas are exactly
identical. Data sources usually expose additional read-only fields.

Use generated `New*Value` constructors for nested custom object types. Generic
`types.ObjectValue` can produce an incompatible object type.

## 7. Test behavior, not only compilation

For each new configurable field, add:

1. create coverage;
2. update coverage;
3. `PlanOnly: true, ExpectNonEmptyPlan: false`.

Also add focused unit tests for model conversion and nil/state reconciliation.

For arrays/maps:

- test multiple values;
- test empty/omitted behavior;
- test order when order is meaningful;
- test mutually exclusive legacy/new forms;
- use a real compatible API object when validation depends on backend type.

If acceptance coverage needs a pre-existing specialized object, read its token
from an explicit environment variable and skip with a precise message when
unset. Keep unit coverage for payload mapping unconditional.

Run at minimum:

```bash
go vet ./...
go test -count=1 ./...
TF_ACC=1 go test -count=1 -run '<relevant-acceptance-tests>' -v ./vantage/
go generate ./...
```

## 8. Final verification

Before handoff:

- the new vantage-go release is resolvable without a local `replace`;
- `go.mod` references the released version;
- no local paths remain in `go.mod`;
- no unrelated generated files changed;
- resource/data-source docs match the final schema;
- generated files are reproducible;
- all relevant unit and acceptance tests pass;
- commits are pushed to the intended feature branch.
