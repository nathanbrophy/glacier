# Specs

Specs are Glacier's source of truth. Every code change, identity change, or process change is first a spec. The Docs & Identity agent (Magpie) generates public-site content and reference documentation directly from accepted specs.

## Authoring a spec

1. **Pick the next ID.** List existing specs and increment:
   ```sh
   ls specs/ | grep -E '^[0-9]{4}-' | sort | tail -1
   ```
   IDs are zero-padded four-digit, monotonically increasing, never reused on supersede.

2. **Copy the template.** From the repo root:
   ```sh
   cp specs/_template.md specs/NNNN-<slug>.md
   ```
   Replace `NNNN` with the new ID and `<slug>` with a kebab-case title.

3. **Fill the front matter.** Set `id`, `title`, `slug`, `owner-agent`, `created`, `last-updated`. Set `status: proposed`. Populate `reviewers` per the sign-off matrix in [`0000-spec-process.md`](0000-spec-process.md). `## Open Questions` may be non-empty during `proposed` and `in-review`.

4. **Write the spec.** Every required section in the template must be filled. Section headers are stable anchors — do not rename or reorder them. Public-marked sections feed the docs site verbatim, so write them in end-user voice; Internal-marked sections are engineering-only.

5. **Move to `in-review`.** Set `status: in-review` and open a PR. Reviewers update their `signed-off-at` ISO timestamp under `reviewers:` when satisfied.

6. **Move to `accepted`.** When all required reviewers have signed off **and** `## Open Questions` is empty (this is non-negotiable), set `status: accepted`. Implementation may now begin.

7. **Move to `implemented`.** When the code or content for the spec has merged, set `status: implemented` and add the merging commit hash(es) to `implementing-commits`.

8. **Move to `verified`.** When the spec's `## Verification` section has been executed end-to-end and the outcome matches expectations, set `status: verified` and `verified-at` to the ISO timestamp.

## Spec lifecycle (state machine)

```
proposed → in-review → accepted → implemented → verified
                          ▲
                          ╰─ HARD GATE: implementation may not begin until status === accepted.
```

## Sign-off matrix (short form)

Full detail in [`0000-spec-process.md`](0000-spec-process.md).

| Spec type | Required reviewers |
|---|---|
| Identity / brand / voice | Magpie (owner), Otter |
| Framework architecture / cross-cutting | Otter (owner), Lynx, Falcon |
| Component (CLI, mock, httpmock, sandbox, primitives) | Otter (owner), Lynx, Falcon |
| Testing infrastructure | Lynx (owner), Otter, Falcon |
| Process / governance change | Otter (owner), Magpie |
| Research artifact (`research/`) | Owner only — no gate |

## Layout

```
specs/
├── README.md                this file
├── _template.md             canonical template (copy this)
├── 0000-spec-process.md     meta-spec: defines the spec process itself
├── NNNN-<slug>.md           subsequent specs, in order
├── research/                ungated research artifacts (UX surveys, etc.)
└── superseded/              retired specs, preserved by ID with `superseded-by` populated
```
