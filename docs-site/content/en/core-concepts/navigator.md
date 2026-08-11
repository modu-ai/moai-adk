---
title: Navigator Binding Tokens
weight: 25
draft: false
---
# Navigator Binding Tokens

When code and documentation point to each other, an agent that fixes one side can immediately pull in the other side's context. **Navigator Binding Tokens** are three authoring tokens that connect design decisions, code symbols, and SPECs into a single addressable graph. These tokens combine to produce one artifact: `.moai/project/navigator/nav-graph.json`.

## The Three Token Families

The navigator integration layer merges three binding-token families into one graph.

| Family | Token form | Author surface | Purpose |
|--------|------------|----------------|---------|
| `NAV:DEC` | `@NAV:DEC-<id>` | Design docs (`.moai/project/*.md`, `.moai/docs/**/*.md`) | Link a design decision to a SPEC or symbol |
| `NAV:SYM` | `@NAV:SYM:<symbol>` | Code comments + Design docs | Link a doc location to a named code symbol |
| `MX:SPEC` | `@MX:SPEC:<SPEC-ID>` | Code comments (sub-line of an `@MX:` tag) | Link a code location to a SPEC |

`MX:SPEC` is already covered by the [MX tag system](/en/advanced/mx-tags/). The navigator integration layer **consumes** the mx-scanner's `SpecAssociator` output — it does NOT re-scan. So don't author this token here; follow the existing MX tag rules instead.

## When to Author These Tokens

### Author `@NAV:DEC-<id>` when

- A design decision in `.moai/project/tech.md`, `structure.md`, `product.md` or under `.moai/docs/` corresponds to a SPEC or a named code symbol.
- You want future code edits to surface the design context that motivated them.

### Author `@NAV:SYM:<symbol>` when

- A doc location or code comment should bind to a named code symbol so a reader of the graph can navigate from the doc to the code (or symbol to symbol).

Do NOT author `@MX:SPEC:` here — that's the mx-scanner surface. Re-authoring it is unnecessary.

## Token Grammar

Both tokens MUST NOT carry empty values. The scanner skips items with empty values and writes a diagnostic warning to `.moai/logs/navigator-sync.log` but does NOT abort the graph build — fail-open.

### `@NAV:DEC-<id>`

`<id>` MUST match `[A-Z][A-Z0-9-]*` (uppercase ASCII plus digits and internal hyphens). Consistent with SPEC-ID domain tokens. The `@NAV:DEC-` prefix is the unambiguous discriminator — the id alone never appears without it.

### `@NAV:SYM:<symbol>`

`<symbol>` MUST match `[A-Za-z_][A-Za-z0-9_.]*` (identifier-shaped, language-neutral). A package-qualified form (`pkg.ParseHeader`) is conventional; a bare form (`ParseHeader`) is also accepted and resolves by suffix match against the existing symbol set.

## Scan Root

The navigator integration layer scans these surfaces:

- **Design docs**: `.moai/project/{product,structure,tech}.md` + `.moai/docs/**/*.md`.
- **Code** (for `@NAV:SYM` only): Go `*.go` files excluding `*_test.go` and `vendor/`, plus the design-doc surface above.

The layer does NOT scan:
- `.moai/specs/` — already covered by the mx-scanner body-based association.
- `.moai/reports/`, `.moai/state/` — ephemeral / runtime state.
- The three existing Navigator chains' source code (consumer-only).

## Output — `nav-graph.json`

A single artifact at `.moai/project/navigator/nav-graph.json` with the shape:

```json
{
  "provenance": { "extract_commit_sha": "...", "captured_at": "..." },
  "nodes": [
    { "entity_type": "decision|spec|symbol", "identifier": "...", "display_name": "..." }
  ],
  "edges": [
    { "edge_type": "dec-edge|spec-edge|sym-edge", "source_node": "...", "target_node": "...", "source_path": "...", "line_number": 0 }
  ]
}
```

`entity_type` is one of `decision | spec | symbol`; `edge_type` is one of `dec-edge | spec-edge | sym-edge`.

The artifact is **byte-stable**: two runs on the same git HEAD produce byte-identical output (no wall-clock timestamp). Because the timestamp isn't recorded, results are identical regardless of who ran it or when. Auditing and reproducibility rest on this property.

{{< callout type="info" >}}
**fail-open** — The graph build always exits 0. Even with malformed tokens, it doesn't abort — it writes a diagnostic warning and builds the graph from the healthy portions.
{{< /callout >}}

## Authoring Example

The simplest form: pointing from a design doc to decisions and symbols, and echoing the same decisions and symbols in code comments.

Design doc (`tech.md`):

```markdown
# Tech

The session layer adopts OAuth2 for delegated access.

Decision @NAV:DEC-AUTH-STRATEGY: OAuth2 over client-credentials.

The header parser (see @NAV:SYM:pkg.ParseHeader) extracts the bearer token.
```

Code (`auth/auth.go`):

```go
package auth

// @NAV:DEC-AUTH-STRATEGY: implement OAuth2 client-credentials flow.
// @NAV:SYM:auth.ParseBearer extracts the bearer token from the Authorization header.
func ParseBearer(h string) string { ... }
```

From these two files, the graph builds three nodes (decision `AUTH-STRATEGY`, symbol `pkg.ParseHeader`, symbol `auth.ParseBearer`) and the edges between them. A reader of the graph can move freely from design docs to code and from code to design rationale.

## Forward Compatibility

The token grammar, the binding-record 5-field shape, and the graph schema are all forward-compatible (additive only). Later milestones MAY add fields; existing fields keep their names and shapes. Tokens you author once remain valid long-term.

## Related Documents

- [MX tag system](/en/advanced/mx-tags/) — The source rule for `@MX:SPEC` tokens. The navigator integration layer consumes this output.
- [SPEC-based development](/en/core-concepts/spec-based-dev/) — SPEC lifecycle and the parent context of `@MX:SPEC`.
- [Agent guide](/en/advanced/agent-guide/) — How agents traverse between code comments and design docs.
