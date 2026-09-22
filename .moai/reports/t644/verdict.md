# t644 verdict — Stop-hook shared-snapshot claim

## Claim under test

`quality-gates-quality.md` (local line 49, template line 43, identical text) said three sync-phase
consumers, including the `sync-phase-quality-gate.sh` Stop hook, "all consume this single snapshot keyed by
HEAD SHA rather than each independently re-executing `go test` / `golangci-lint` / `go vet` / `go test -cover`".

## Verdict: FALSE for the Stop hook — sentence corrected, hook behavior unchanged

Evidence, read at base `eb50af5a8` (local and template hook copies byte-identical, `cmp` exit 0):

- `sync-phase-quality-gate.sh:248-259` `consume_snapshot` runs `moai verify check --key-current` and prints
  `hit` / `miss` / `unavailable`.
- `:414-415` is the only use of that result: `SNAPSHOT_STATUS=$(consume_snapshot)` then
  `log_gate_event "snapshot_status=$SNAPSHOT_STATUS"`. `grep -n SNAPSHOT_STATUS` finds only these two lines,
  so the result never reaches the verdict.
- `:446-455` then always runs its own fast checks (for Go: `go vet ./...`, `go build ./...`); the header
  (`:6-11`, `:19`) states that the test suite, coverage, and golangci-lint are not run by this hook.
- `internal/cli/verify.go:180-218` (`verify check`) only loads and reports freshness; it records nothing.

So the hook reads the snapshot for logging only, re-runs vet itself, never ran the other three commands,
and never records into the snapshot.

## Change

The sentence now names two consumers (sync-auditor Evidence cells, sync-audit-4dim judges) and states
the hook's actual behavior in a separate sentence, in language-neutral wording (template copy is
distributed). The `moai` catalog hash was regenerated (`e870bc11` → `5f112723`) because the template
skill tree changed.

`git diff --stat` against `eb50af5a8`: 3 files, 3 insertions, 3 deletions — the two copies plus
`internal/template/catalog.yaml`. This is one line in each of two copies plus a generated hash line, not a
single-file change; whether that still counts as Class A is the lead's call. CI has not run (lanes do not
push).

## Verification (this tree, working copy on top of `eb50af5a8`)

- `MOAI_TEMPLATE_LEAK_STRICT=1 go test -v ./internal/template/ -run '^(TestTemplateNoInternalContentLeak|TestCatalogHashCoversSkillSubfiles|TestManifestHashFormat|TestAllSkillsInCatalog|TestCatalogReferencesValid)$'`
  → exit 0, 5 × `--- PASS`.
- `go test -count=1 ./internal/template/...` → exit 0 (template, agentemit, commandemit ok).
- Mutant: old hash `e870bc11` put back → `TestCatalogHashCoversSkillSubfiles` FAIL; restored.

## Primary-checkout uncommitted edit (not touched)

The primary checkout holds an uncommitted edit to the local copy (base blob `289836bf6`, older than
develop's `2a6ecd006`). It rewrites the "Launch three background tasks simultaneously:" line into the
shared-bounded-queue wording. That wording already exists in develop's local copy (line 58), so the
edit is a stale copy of a change that has landed. It does not touch the claim sentence; no overlap
with this card.

## Gaps

- The other two consumers (sync-auditor Evidence cells, sync-audit-4dim judges) were not verified; the
  sentence keeps the original claim for them unchanged.
- The local and template copies differ elsewhere (e.g. the bounded-queue paragraph exists only in the
  local copy); not this card's scope.
- CI not run.
