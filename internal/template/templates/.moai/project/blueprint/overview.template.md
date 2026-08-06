<!--
  Module Overview (Blueprint Layer) — author-facing template.

  This file is the per-module overview document the blueprint layer composes
  from `module_tree.json` entries. It is a SCAFFOLD only: a human author
  fills each section, then the file is treated as authored content (the
  blueprint layer does NOT auto-replace a human-edited overview).

  Section structure follows the Kiro Design 7-section module overview. Each
  heading below is a slot; replace the placeholder prose with module-specific
  content. Delete this comment block before authoring.
-->
# {{module_display_name}}

<!--
  Provenance placeholder. The blueprint layer fills `last_updated_commit`
  with the git SHA of the run that last authored or refined the prose. It is
  a git baseline (NOT a wall-clock timestamp), so two readers at the same
  commit see the same value. Leave the placeholder literally until the
  blueprint layer stamps it.
-->
- **Package path**: `{{module_package_path}}`
- **Layer**: {{module_layer}}
- **Responsibility**: {{module_responsibility_one_liner}}
- **Last updated commit**: `{{last_updated_commit}}`

## Component Architecture

Describe the internal components of this module and how they collaborate.
Name the significant types, the boundaries between them, and the reason
the module is split this way.

<!-- example shape:
- `Client` — entry point; owns the connection lifecycle.
- `Router` — dispatches inbound messages to per-method handlers.
- `Transport` — abstracts the wire (stdio, tcp, in-memory).
-->

## Data Flow

Trace the path of a representative request through this module, from the
entry point to the side effect and back. Name every transform and every
hand-off. A reader who finishes this section should be able to draw the
flow on paper.

<!-- example shape:
Caller -> Client.Call -> Router.dispatch -> Handler -> Transport.Write -> wire
wire   -> Transport.Read -> Router.route   -> Handler -> reply to Caller
-->

## Data Model

List the canonical types this module owns and the invariants each one
guarantees. Call out which fields are required, which are optional, and
which carry representation invariants a caller can assume.

<!-- example shape:
- `Config` (struct): command, args, root URI. Required: command. Optional: args.
- `Event` (sum type): tagged union over {Started, Stopped, Error}.
-->

## Error Handling

Enumerate the error modes this module can surface and the recovery
contract for each. A reader should learn which errors are retryable, which
are fatal, and which are programming bugs.

<!-- example shape:
- `ErrConnClosed` — caller may reconnect.
- `ErrInvalidMsg` — programming bug; do not retry; fix the caller.
-->

## Test Strategy

Describe how this module is verified. Name the layers (unit / integration
/ characterization), the fixtures they share, and the behaviors that are
locked by characterization tests.

<!-- example shape:
- Unit: table-driven over `Client` lifecycle states.
- Characterization: snapshot of Router dispatch order.
-->

## Implementation Approach

Capture the design decisions a maintainer must understand before editing
this module. Record the chosen approach, the rejected alternatives, and
the reason the chosen approach won.

<!-- example shape:
Approach: JSON-RPC over stdio. Rejected: HTTP — adds a transport we do not
need for a local subprocess.
-->

## Migration

Describe how this module evolves when its public surface changes. Name
the compat window, the deprecation path, and where the successor module
lives (if any).

<!-- example shape:
v1 callers: keep until the next major. Migration: rename `Dial` to `Connect`.
-->
