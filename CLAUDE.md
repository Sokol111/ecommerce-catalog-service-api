# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this repo is

This is the **API contract repo** for `ecommerce-catalog-service` — the write side / source
of truth in a CQRS + event-driven system. It holds no business logic. It contains:

- **Protobuf sources** (`proto/`) — the only files you hand-edit.
- **Generated code** (`gen/go`, `gen/typescript`) — never hand-edit; regenerate instead.
- **Thin Go helpers** (`pkg/`) that build on the generated code: an `fx` gRPC client module
  and a Kafka topic registry.

Consumers: the catalog service (implements the RPC servers, produces the events), the query
services (consume the events into their read models), and the Nuxt UIs (import the generated
TS client).

## Golden rule: edit proto, then regenerate

**Never hand-edit anything under `gen/`.** All Go and TypeScript there is produced by `buf`
from `proto/`. Edit the `.proto` and run `make generate`. Hand edits are silently overwritten
on the next generation and break the release pipeline.

## Commands

```bash
make generate            # DEFAULT WORKFLOW: lint + generate TS + Go events + Go Connect/gRPC
make lint                # buf lint only (STANDARD rules)
make format              # buf format -w
make connect-breaking    # check proto for breaking changes against .git#branch=main
make tidy                # go mod tidy
make update-proto-deps   # buf dep update (refresh buf.lock)
make connect-install-tools   # install buf + protoc-gen-{go,connect-go,go-grpc} at pinned versions
make help                # list all targets grouped by category
```

`make generate` runs three independent generators (see `makefiles/`), each driven by its own
buf template:

- **`connect-generate`** → Go structs + Connect + gRPC from `proto/catalog/v1/` (template
  `buf.gen.yaml`) into `gen/go/catalog/v1/`.
- **`events-generate`** → Go structs from `proto/catalog/events/v1/` (template
  `buf.gen.events.yaml`) into `gen/go/catalog/events/v1/`.
- **`connect-ts-generate`** → TypeScript client + a generated `package.json`/`tsconfig.json`,
  then `npm run build`. The package is `@sokol111/ecommerce-catalog-service-api`, versioned
  from the `VERSION` file. Use `connect-ts-generate-fast` to skip the npm build (CI builds
  before publishing).

## Proto layout and conventions

Two proto trees under `proto/catalog/`, with distinct purposes and package names:

- `v1/` (`package catalog.v1`) — **synchronous RPC contracts**. Three services, one file each:
  `product.proto`, `category.proto`, `attribute.proto`. Each defines entities, request/response
  messages, and a `service`.
- `events/v1/` (`package catalog.events.v1`) — **Kafka event schemas** emitted by the catalog
  service: `*UpdatedEvent` / `*DeletedEvent` per aggregate.

Key modeling decision to preserve when editing events: **events carry only immutable references
(IDs, slugs) plus product-specific values — not mutable display data** (attribute names, option
names). Consumers join mutable master data from separate `*UpdatedEvent` streams. Don't add
denormalized display fields to product events.

Polymorphic attribute values are modeled as a protobuf `oneof` (`option_slug_value`,
`numeric_value`, `text_value`, `boolean_value`, and a `StringList` wrapper for repeated slugs)
— mirrored in both the RPC `AttributeValue` and the event `AttributeValue`.

The buf module (`buf.build/sokol111/catalog-api`) depends on `protovalidate`; validation rules
belong in the proto as protovalidate options.

## The `pkg/` helpers

- `pkg/client/grpc.go` — `client.Module()` returns an `fx.Option` wiring native gRPC clients
  for all three services (`ProductServiceClient`, `AttributeServiceClient`,
  `CategoryServiceClient`), reading config from koanf key `catalog.grpc`. This is how consumer
  services get a catalog client — import the module, don't construct clients by hand.
- `pkg/events/topics.go` — maps proto event message full-names to Kafka topic names
  (`catalog.product.events`, `catalog.category.events`, `catalog.attribute.events`) and exposes
  `TopicFor(msg)`. **When you add a new event message, register it in `topicMap`** or `TopicFor`
  panics at runtime for that type.

## Releasing (production only)

Versioning is release-then-bump. Pushing a change to the `VERSION` file on `master` triggers
`.github/workflows/release.yml`, which calls the shared
`ecommerce-infrastructure/.github/workflows/api-release.yml` — this tags the Go module and
publishes the TS package to GitHub Packages, running a breaking-change check first
(`skip_breaking` input to override). Consuming services then bump their `go.mod` dependency.

Note: this applies to CI/production. In local development everything resolves through the root
`go.work`, so proto changes are visible to consumers immediately with no release/bump.
