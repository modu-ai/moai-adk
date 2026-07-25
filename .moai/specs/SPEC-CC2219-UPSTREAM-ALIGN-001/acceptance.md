# acceptance.md — SPEC-CC2219-UPSTREAM-ALIGN-001

## §D AC Matrix

All greps exclude `.claude/worktrees/`, `.moai/specs/`, `.moai/reports/`, `CHANGELOG.md` unless stated. "Live+T" = live surface AND its `internal/template/templates/` mirror.

### Child A — GD-1 (nesting)

- **AC-GD1-001** (verification: after the M1 rewrite, the stale default-off claims are absent):
  `grep -rn 'default changed to \*\*off\*\*\|defaults to `0` = nesting disabled\|runtime default-off' CLAUDE.md .claude/rules/ .claude/agents/ internal/template/templates/CLAUDE.md internal/template/templates/.claude/` → **0 matches**.
- **AC-GD1-002**: `grep -rn 'double guarantee' CLAUDE.md .claude/rules/ .claude/agents/ internal/template/templates/` → **0 matches** (framing removed per REQ-GD1-002).
- **AC-GD1-003**: Rewritten surfaces state default-ON + `CLAUDE_CODE_MAX_SUBAGENT_SPAWN_DEPTH=1` disables: `grep -rln 'MAX_SUBAGENT_SPAWN_DEPTH=1' CLAUDE.md .claude/rules/moai/development/agent-authoring.md internal/template/templates/CLAUDE.md internal/template/templates/.claude/rules/moai/development/agent-authoring.md` returns all four files.
- **AC-GD1-004** (probe caveat): the CLAUDE.md §4 nesting note contains a caveat that the depth-3 ceiling is changelog-sourced (single depth-1 probe) — grep for a `single depth-1` (or equivalent) caveat token in CLAUDE.md returns ≥1, Live+T.
- **AC-GD1-005** (pilot decision): EITHER `sync-auditor.md` frontmatter `tools:` no longer contains `Agent` (option a) OR `sync-auditor.md` body's nesting-pilot rationale references tool-restricted children only (`Explore`) and contains no `mode: "plan"` justification (option b). Decision provenance recorded in progress.md §E.2. Live+T.
- **AC-GD1-006**: CLAUDE.md §4 (or agent-authoring.md) carries a dated supersession note naming SPEC-SUBAGENT-NESTING-DOCTRINE-001 — live surface only; the template mirror carries the neutral mechanism text WITHOUT the SPEC ID (§25 neutrality). Verify: SPEC-ID grep inside `internal/template/templates/` → 0.

### Child A — GD-2 (mode parameter)

- **AC-GD2-001**: `grep -rn 'mode: "plan"' CLAUDE.md .claude/rules/ .claude/agents/ internal/template/templates/CLAUDE.md internal/template/templates/.claude/` → **0 matches** on live+template doctrine surfaces.
- **AC-GD2-002**: Read-only grounding present: `sync-auditor.md` and `worktree-integration.md` read-only prose references tool restriction (`Explore` / tools-list omission), verified by grep for the new grounding token on each rewritten file (Live+T).
- **AC-GD2-003**: Provenance caveat present: at least one child-A surface states the deprecation is changelog/doc-sourced (2.1.213), not runtime-observed.

### Child C — GD-4 (Opus 5)

- **AC-GD4-001**: `grep -n '"opus"' internal/template/model_policy.go` shows the alias resolving to a `claude-opus-5` const; `ModelDeprecatedCanonicalIDs` contains a `claude-opus-4-8` row. `go test ./internal/template/...` green (verbatim output cited).
- **AC-GD4-002** (@MX:ANCHOR closure): all three fan-in consumers (`expandModelString`, `normalizeModel`, `modelOptions`) compile and their tests pass — `go build ./... && go test ./internal/cli/... ./internal/template/... ./internal/settings/...` exit 0 (adjust pkg paths to actual fan-in locations re-measured at M2).
- **AC-GD4-003a**: `grep -rn 'claude-opus-4-8' internal/web/ --include='*.go' | grep -v _test.go` → **0 matches**.
- **AC-GD4-003b**: `appbar_context_test.go` asserts the Opus 5 id and `go test ./internal/web/...` exits 0 (verbatim output cited).
- **AC-GD4-004** (naming sweep): `grep -rn 'opus = Opus 4.8\|opus` resolves to Opus 4.8' .claude/rules/` → 0; `context-window-management.md` table contains an `Opus 5` 1M row and no bare ambiguous `Opus` 256K label (Live+T).
- **AC-GD4-005**: `quality.yaml.tmpl` no longer claims `xhigh`/`max` require Opus 4.7; effort-availability text matches report §4 GD-4 S5 quotes.

### Child D — GD-5/6/7

- **AC-D-001**: `native-invocation-model.md` `/code-review` and `/deep-research` rows no longer classify them as auto-invocable; Axis A L67/L87 annotated manual-only (Live+T).
- **AC-D-002**: CLAUDE.md §10 `/deep-research` prose + `dynamic-workflows.md` L78 note manual invocation only (Live+T).
- **AC-D-003**: `dynamic-workflows.md` size prose contains `unrestricted`, explicit `medium` default (<15 agents), and `workflowSizeGuideline`; `settings-management.md` key table contains a `workflowSizeGuideline` row (Live+T).
- **AC-D-004**: `grep -c 'DirectoryAdded' .claude/rules/moai/core/hooks-system.md internal/template/templates/.claude/rules/moai/core/hooks-system.md` → ≥1 each; entry notes no MoAI handler is wired.

### Child E — GD-8/9

- **AC-E-001**: `agent-authoring.md` fork section references `/subtask` for the in-session fork and `/fork` as background-session copy; the 2.1.212/2.1.213 upstream inconsistency is recorded (Live+T).
- **AC-E-002**: `skill-authoring.md` + `moai-foundation-cc/reference/claude-code-skills-official.md` document `context: fork` background-by-default + `background: false` opt-out (Live+T).

### Child F — docs-site/README

- **AC-F-001**: For each of the 4 stale-claim families (nesting default-off, `mode: "plan"` read-only, Opus 4.8 default-Opus, auto `/deep-research`): either 4-locale-parity edits landed in docs-site + README, or a recorded 0-match grep evidences no counterpart (REQ-F-002). Per-family evidence table in progress.md §E.2.
- **AC-F-002**: docs-site edits pass the harness verify recipe (warning-free hugo build + 4-locale parity checks) when any edit was made.

### Cross-cutting

- **AC-X-001**: Every edited `.claude/` file has a mirror edit or a documented local-only/sanitized-pair exemption; `make build` exits 0.
- **AC-X-002**: Neutrality guards green: `go test ./internal/template/ -run 'Leak|Neutrality'` (actual test names re-measured) exit 0; no SPEC ID/date/SHA introduced into templates.
- **AC-X-003** (GD-3 exclusion validity — branch-staleness-aware): `git show origin/main:.claude/settings.json | grep -c 'startup|resume|clear|compact|fork'` → ≥1 (PR #1146, commit 714270085, landed on main). The CURRENT checkout may predate that merge and still lack `fork` locally — that is NOT a failure; the AC is anchored to origin/main. After rebase onto origin/main, the local file matches. Observation only, no edit by this SPEC.
- **AC-X-004**: `moai spec lint` (or `go run ./cmd/moai spec lint`) reports 0 errors for this SPEC (repo-wide pre-existing findings are excluded via per-SPEC JSON filtering).
- **AC-X-005** (REQ-X-003 concurrency-safeguard preservation): `grep -c 'does not run two write-capable agents concurrently' CLAUDE.md` → ≥1 AND `grep -c 'does not run two write-capable agents concurrently' internal/template/templates/CLAUDE.md` → ≥1 (measured baseline: 1 each on 2026-07-25) — the safeguard prose survives every child-A edit verbatim.

## §D.1 Edge cases

- A mirror file that is a sanitized pair (not byte-identical): apply the semantic edit to both, keep the sanitization delta; do not force byte parity.
- Upstream publishes a correcting changelog entry mid-run: halt the affected milestone, return a blocker report with the new evidence.
- `opus[1m]` picker: if Opus 5 native 1M makes the `[1m]` suffix a no-op, record the decision (keep for back-compat vs drop) rather than silently removing.

## §D.2 Definition of Done

All AC rows PASS with verbatim command output cited in progress.md §E.2; M1 decision gate provenance recorded; build/test/neutrality/mirror guards green; docs-site family evidence table complete; GD-3 exclusion observation recorded.
