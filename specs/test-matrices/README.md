# Test matrices

> Lynx-owned exhaustive test matrices for Glacier's 14 v0 packages. **Input artifacts** for component-spec authoring per spec 0002 §25. Producing `go test ./...` exit-code-0 → 100% release confidence is the standard these matrices uphold.

## Files

| File | Scope | Test files | Test rows |
|---|---|---|---|
| [`kernel.md`](kernel.md) | `option`, `errs`, `log`, `assert` (+ `assert/require`), `term` | ~56 | ~560+ |
| [`mid.md`](mid.md) | `concur`, `fluent`, `conf`, `fixture`, `obs` | ~60 | ~510+ |
| [`leaves.md`](leaves.md) | `cli` (+ `cli/gen`), `mock`, `httpmock`, `httpc` | ~60 | ~410+ |

**Framework total:** ~170 test files, ~1480+ test rows across 14 packages, plus runnable godoc Examples for every public-API symbol.

## How these matrices were produced

Three Lynx instances reviewed the framework-shape plan ([`mongoose-spec-0002-framework-shape.md`](../../../../.claude/plans/mongoose-spec-0002-framework-shape.md)) in parallel — one per tier — and produced exhaustive test matrices satisfying the test-first commitment from spec 0002 §25. Each matrix was built from:

- The package's interview content (§21.X subsections of the framework plan).
- The post-review amendments (§23 of the plan) — recalibrated performance targets, locked concurrency behaviors, generics fixes, lifecycle Close additions, naming disambiguations.
- The cross-cutting test-discipline rules (§25 of the plan).
- Lynx's own charter — the sharp-eyed defect hunter who sees failure modes other agents glance past.

The matrices are **persisted as agent-output artifacts** in the framework's session cache:

- Kernel: `C:\Users\natha\.claude\projects\C--Users-natha-Projects-mongoose\898eac2a-39a3-490d-9d0c-c39ad1bd8720\tool-results\toolu_016S7q8M4ESJPZgGffTovN8z.json`
- Mid: `C:\Users\natha\.claude\projects\C--Users-natha-Projects-mongoose\898eac2a-39a3-490d-9d0c-c39ad1bd8720\tool-results\toolu_01QegE6UENy2mVZPJHAEzWrY.json`
- Leaves: `C:\Users\natha\.claude\projects\C--Users-natha-Projects-mongoose\898eac2a-39a3-490d-9d0c-c39ad1bd8720\tool-results\toolu_01DiiSzyUEwNgPGAw6mbeEdz.json`

The committed `.md` files in this directory are the human-readable, repo-tracked form. The cache files are the immutable source-of-truth from the agent reviews.

## How to use these matrices

When authoring a component spec (e.g., `0011-cli.md`), the spec's `## Test Matrix` section pulls the corresponding rows from the appropriate matrix file:

1. Locate the package's section in the matching matrix file (e.g., `## Package: cli/` in `leaves.md`).
2. Copy the test-matrix table verbatim into the component spec's `## Test Matrix` section.
3. Add any spec-specific elaborations (e.g., test-data fixtures, expected error messages) inline.
4. Submit for Lynx signoff alongside the rest of the spec.

Lynx verifies that the spec's matrix equals or exceeds the matrix file's coverage. Component specs may *add* tests (positive direction); they may not silently drop tests Lynx originally listed.

## Mutability

These matrix files are **read-only after Lynx signoff**. Changes require:

1. Lynx countersignature on the change.
2. A justification entry in the amending spec's `## Decisions & Rationale` explaining why a test was added, removed, or restructured.
3. A PR diff that surfaces the matrix change to all required spec reviewers.

The matrix files do **not** follow the spec lifecycle (proposed → accepted → verified). They are framework-internal engineering artifacts, parallel to the agent charter files in `.claude/agents/`. Their canonical owner is Lynx.

## What's covered

Each matrix covers, for every package in scope:

- **Unit tests** for every public symbol (positive, negative, edge, error path).
- **Benchmarks** per the recalibrated D35 / §23.13 targets.
- **Fuzz targets** for parsers and external-input boundaries.
- **Race-detector tests** for every concurrent code path.
- **Property-based tests** for algebraic identities and invariants.
- **Edge cases** drawn from each package's interview (E-rows in §21.X).
- **Cross-platform tests** where platform behavior matters.
- **Lifecycle tests** for Close idempotency and multi-resource cleanup.
- **Generics-fix verification** per §23.17.
- **Spec traceability** — every test references its spec section.

Plus the cross-package **integration tests** that exercise the framework's composition story (5+ packages together) — captured in spec 0002 §25.9.

## Test-helper dogfooding

Every test in Glacier uses Glacier's own test helpers:

- `assert/` for value assertions
- `assert/require/` for halt-on-fail variants
- `fixture/` for golden files, mock filesystem, deterministic clock, stdout capture, leak detection
- `mock/` for interface mocks (runtime or codegen)
- `httpmock/` for HTTP transport stubs

The kernel packages face a bootstrap problem (`assert/`'s tests can't naively use `assert/`); resolved by a documented bootstrap subset of bare-`if` checks for the most foundational primitives. See `kernel.md` "Bootstrap discipline" for the canonical list.

External test libraries (`testify`, `quicktest`, `is`, `goconvey`, etc.) are **forbidden in Glacier's own tests**.
