---
description: "Author-facing NAV:DEC / NAV:SYM binding-token reference for the BAS integration layer"
paths: "**/.moai/project/*.md,**/.moai/docs/**/*.md,**/*.go,**/nav-tokens.md"
---

# Navigator Binding Tokens

> Author-facing token reference for the BAS (Blueprint-Anchored Synchronization)
> integration layer. Documents the two NEW author-facing tokens this layer
> introduces. The third family (`@MX:SPEC`) is already documented under
> `mx-tag-protocol.md` and is consumed (not re-scanned) by the integration
> layer.

## The binding-token trio

The integration layer joins three SSOT binding-token families into a single
addressable graph (`nav-graph.json`):

| Family | Token form | Author surface | Purpose |
|--------|------------|----------------|---------|
| `NAV:DEC` | `@NAV:DEC-<id>` | design docs (`.moai/project/*.md`, `.moai/docs/**/*.md`) | link a design decision to a SPEC or symbol |
| `NAV:SYM` | `@NAV:SYM:<symbol>` | code comments + design docs | link a doc location to a named code symbol |
| `MX:SPEC` | `@MX:SPEC:<SPEC-ID>` | code comments (sub-line of an `@MX:` tag) | link a code location to a SPEC (consumed from the mx-scanner; NOT re-scanned here) |

## When to author these tokens

Author `@NAV:DEC-<id>` when:
- A design decision in `tech.md` / `structure.md` / `product.md` or under
  `.moai/docs/` corresponds to a SPEC or a named code symbol.
- You want future code edits to surface the design context that motivated
  them.

Author `@NAV:SYM:<symbol>` when:
- A doc location or code comment should bind to a named code symbol so a
  reader of the graph can navigate from the doc to the code (or symbol to
  symbol).

Do NOT author `@MX:SPEC:` here — that token is the existing mx-scanner
surface (see `mx-tag-protocol.md`). The integration layer CONSUMES the
mx-scanner's `SpecAssociator` output; re-authoring `@MX:SPEC` is unnecessary.

## Token grammar

- `@NAV:DEC-<id>` — `<id>` MUST match `[A-Z][A-Z0-9-]*` (uppercase ASCII +
  digits + internal hyphens). Consistent with SPEC-ID domain tokens. The
  `@NAV:DEC-` prefix is the unambiguous discriminator — the id alone never
  appears without it.
- `@NAV:SYM:<symbol>` — `<symbol>` MUST match `[A-Za-z_][A-Za-z0-9_.]*`
  (identifier-shaped, language-neutral). A package-qualified form
  (`pkg.ParseHeader`) is conventional; a bare form (`ParseHeader`) is
  accepted and resolves by suffix match against the existing symbol set.

## Example

```markdown
# Tech

The session layer adopts OAuth2 for delegated access.

Decision @NAV:DEC-AUTH-STRATEGY: OAuth2 over client-credentials.

The header parser (see @NAV:SYM:pkg.ParseHeader) extracts the bearer token.
```

```go
package auth

// @NAV:DEC-AUTH-STRATEGY: implement OAuth2 client-credentials flow.
// @NAV:SYM:auth.ParseBearer extracts the bearer token from the Authorization header.
func ParseBearer(h string) string { ... }
```

## Malformed tokens

A token with an empty `<id>` or `<symbol>` (`@NAV:DEC-` / `@NAV:SYM:`) is
skipped with a diagnostic warning written to `.moai/logs/navigator-sync.log`
and does NOT abort the graph build. The build is fail-open — exit code 0
always.

## Scan root

The integration layer scans these surfaces:
- Design docs: `.moai/project/{product,structure,tech}.md` + `.moai/docs/**/*.md`.
- Code (for `@NAV:SYM` only): Go `*.go` files excluding `*_test.go` and
  `vendor/`, plus the design-doc surface above.

The layer does NOT scan:
- `.moai/specs/` — already covered by the mx-scanner body-based association.
- `.moai/reports/`, `.moai/state/` — ephemeral / runtime state.
- The three existing Navigator chains' source code (consumer-only).

## Output

A single artifact at `.moai/project/navigator/nav-graph.json` with the shape:

```json
{
  "provenance": { "extract_commit_sha": "...", "captured_at": "..." },
  "nodes": [ { "entity_type": "decision|spec|symbol", "identifier": "...", "display_name": "..." } ],
  "edges": [ { "edge_type": "dec-edge|spec-edge|sym-edge", "source_node": "...", "target_node": "...", "source_path": "...", "line_number": N } ]
}
```

The artifact is byte-stable: two runs on the same git HEAD produce
byte-identical output (no wall-clock timestamp).

## Forward compatibility

The token grammar, the binding-record 5-field shape, and the graph schema
are forward-compatible (additive only). Later milestones MAY add fields;
existing fields keep their names and shapes.

## Cross-references

- `mx-tag-protocol.md` — the existing `@MX:` tag system (the integration
  layer consumes the `@MX:SPEC` association source from there).
- `verification-claim-integrity.md` — every binding record is attributable
  to a git baseline via the `provenance` block.
