---
id: NNNN
title: <Title in Title Case>
slug: <kebab-case-slug>
status: proposed                   # proposed | in-review | accepted | implemented | verified
owner-agent: <otter|magpie|lynx|gopher|octopus|falcon>
created: YYYY-MM-DD
last-updated: YYYY-MM-DD
supersedes: []                     # list of spec IDs this replaces, e.g. [0003]
superseded-by: null                # populated when this spec is replaced
reviewers:
  - { agent: <agent>,  required: true,  signed-off-at: null }
  - { agent: <agent>,  required: true,  signed-off-at: null }
  # add optional reviewers as needed:
  # - { agent: <agent>, required: false, signed-off-at: null }
implementing-commits: []           # populated at status=implemented
verified-at: null                  # populated at status=verified
docs-extract:                      # which sections Magpie pulls into public docs
  - public-summary
  - mental-model
  - api
  - examples
  - faq
---

# <Title>

<!--
  Section headers below are STABLE ANCHORS. Magpie extracts content by header,
  so do not rename or reorder them. Doing so is a process change requiring its
  own spec.

  Sections marked **Public** are extracted by Magpie for the public site.
  Sections marked **Internal** are engineering-only and never appear in published docs.
-->

## Public Summary

<!-- **Public.** One paragraph in end-user voice. The canonical description for the site and README. -->

## Mental Model

<!-- **Public.** The conceptual frame a developer should hold while using this. Mermaid diagrams welcome. Source for the "Concepts" page on the site. -->

## Goals

<!-- **Internal.** Bulleted list. -->

## Non-Goals

<!-- **Internal.** Bulleted list. What this spec deliberately excludes. -->

## Architecture

<!-- **Internal.** Mermaid diagram + prose. Package layout, data flow, lifecycle. -->

## Schema

<!-- **Internal.** Go types with invariants stated as `// invariant: ...` comments on each field. -->

## API

<!--
  **Public.** Every exported symbol introduced by this spec.
  For each: signature, doc comment (which becomes godoc), preconditions, postconditions,
  error contract, concurrency notes (goroutine-safe? blocking?), lifecycle hooks.
  Magpie extracts signatures + doc comments verbatim to the API reference page.
-->

## Examples

<!--
  **Public.** Runnable Go examples in fenced ```go blocks.
  Each example is self-contained and `go test ./...`-compatible (valid Example functions).
  Magpie transcludes verbatim into tutorials.
-->

## Test Matrix

<!--
  **Internal.** Owned by Lynx.
  Table: scenario × input × expected outcome × covered-by-test-name.
-->

| Scenario | Input | Expected | Covered by |
|---|---|---|---|
|  |  |  |  |

## Dependency Justification

<!--
  **Internal.** Owned by Falcon.
  One row per new direct dependency. The empty table is the goal.
  Required answers: license, last-release-date, maintainer count, alternatives considered, why we don't roll our own.
-->

| Module | Version | License | Last release | Maintainers | Alternatives considered | Why we can't roll our own |
|---|---|---|---|---|---|---|
|  |  |  |  |  |  |  |

## Security & Supply-Chain Notes

<!-- **Internal.** Untrusted-input handling, sandboxing implications, secrets handling, vuln-scan considerations. -->

## Migration & Compatibility

<!-- **Internal.** Only required when this spec changes an accepted spec. Describes the migration path. Delete this section if not applicable. -->

## FAQ

<!-- **Public.** Anticipated user questions with answers. Magpie extracts to the public docs FAQ. -->

## Decisions & Rationale

<!-- **Internal.** Why-this-and-not-that for non-obvious choices. Folded-in ADR. -->

## Open Questions

<!--
  **Internal.** Unresolved items.
  MUST be empty before this spec moves to `accepted` (per CLAUDE.md core directive 1 / D11).
-->

## Verification

<!-- **Internal.** Concrete steps to prove the change works end-to-end. Run when the spec moves to `verified`. -->
