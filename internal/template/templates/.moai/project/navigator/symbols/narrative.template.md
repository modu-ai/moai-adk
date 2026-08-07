<!--
  Per-Symbol Narrative (Tier 3) — author-facing template.

  One file per named symbol (function, method, type, constant). The
  filename SHOULD match the symbol identifier (e.g. `ParseHeader.md`).
  Filled by an author; the symbol layer treats this as authored content
  and does NOT auto-replace a human-edited narrative.

  Slots below: replace the placeholders, then delete this comment.
-->
# {{symbol_display_name}}

<!--
  metadata.json sidecar (sibling file: metadata.json). The symbol layer
  reads this sidecar to attribute the narrative to a git baseline. Both
  fields are git-derived (NO wall-clock timestamp), so two readers at the
  same commit see the same values.

  Required shape:
  {
    "last_updated_commit": "<git SHA>",
    "symbol":              "<package-qualified identifier>"
  }
-->

- **Symbol**: `{{package_qualified_identifier}}`
- **Kind**: {{kind}}  <!-- one of: function | method | type | constant | variable | interface -->
- **Package**: `{{package_path}}`
- **Last updated commit**: `{{last_updated_commit}}`

## Docstring

The one-paragraph description a reader sees first. Answer: what does this
symbol do, what contract does it guarantee, and what does it NOT do?
Avoid restating the signature — the reader can see the signature.

## Call context

Where and why is this symbol called from? Name the significant callers
(typically 3-5), the condition under which each calls it, and the data
that crosses the boundary. A reader who finishes this section should
know when they would reach for this symbol and when they would not.

<!-- example shape:
- `pkg.Client.Call` — inbound request path; calls when forwarding to the
  transport.
- `pkg.Router.dispatch` — per-message dispatch; calls to resolve a method
  name to a handler.
-->

## Examples

Idiomatic call sites. One short block per common usage. Replace the
placeholder with the language-appropriate fenced code block.

<!-- example shape:
```go
h := ParseHeader(raw)
if h == nil {
    return ErrInvalidHeader
}
```
-->
