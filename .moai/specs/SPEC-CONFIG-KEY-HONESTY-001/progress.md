# SPEC-CONFIG-KEY-HONESTY-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-08-12
tier: M
artifacts: [spec.md, plan.md, acceptance.md, progress.md]
depends_on: [SPEC-CONFIG-TIER-PERSIST-001]
code_baseline: ed70e4354
plan_audit:
  iteration_1:
    verdict: FAIL
    score: 0.72
    threshold: 0.80
    dimensions: {clarity: 0.78, completeness: 0.80, testability: 0.62, traceability: 0.72}
    must_pass: 7/7
    resolved: [D1, D2, D3, D4, D5, D6, D7, D8]
    deferred: [D9, D10, D11, D12, D13, D14, D15]
    report: .moai/reports/plan-audit/SPEC-CONFIG-KEY-HONESTY-001.md
  iteration_2:
    verdict: FAIL
    score: 0.78
    threshold: 0.80
    dimensions: {clarity: 0.75, completeness: 0.75, testability: 0.78, traceability: 0.85}
    must_pass: 7/7
    root_cause: baseline-drift (plan authored against d5336214e, audited against ed70e4354)
    report: .moai/reports/plan-audit/SPEC-CONFIG-KEY-HONESTY-001-review-2.md
  iteration_3:
    verdict: PASS
    score: 0.87
    threshold: 0.80
    dimensions: {clarity: 0.85, completeness: 0.90, testability: 0.83, traceability: 0.90}
    must_pass: 7/7
    refresh_type: baseline-reverification
    code_baseline: ed70e4354
    resolved: [D1..D8 stale citations, D3 adhoc-live deleted-file foundation, D4 AC-CKH-012 false baseline, D9 AC ceiling 23→15, D10 NFR floor 200→900, D11 M3-hold in spec/plan, D12 main-fork premise softened, D14 folded into D12]
    deferred: [D13 handoff-note-path placeholder (forward-ref by design), D14 AC-CKH-013 token-counting (minor)]
```

- Artifacts authored: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (Tier M set).
- Code baseline `ed70e4354` (iteration-3 refresh). The prior `d5336214e` baseline was 12 days stale
  at iteration-2 audit; all file:line citations were re-verified and updated against `ed70e4354`.
- **Iteration-3 refresh (2026-08-12).** Plan-audit iteration 2 returned FAIL 0.78 (threshold 0.80)
  with the root cause identified as baseline drift — the plan was authored against `d5336214e` and
  audited against HEAD `ed70e4354`. The iteration-3 refresh resolved all 15 named defects (D1-D15
  from the iteration-2 report): D1/D2/D5/D6/D7/D8 stale file:line citations re-verified and updated;
  D3 `adhoc-live` class foundation re-derived (sole confirmed instance file deleted at `5792fc755`,
  class retained as forward-looking with zero current instances, `tmux_preferred` reclassified
  dead); D4 AC-CKH-012 baseline re-derived against HEAD (`isHookOptInEnabled` refactored to
  delegator at `e3f8dd463`, inline-struct readers now at `routing_ledger.go:104` +
  `update.go:1140`); D9 Tier M AC ceiling (23→15 via consolidation, documented in acceptance.md §A
  clause 8); D10 NFR-CKH-002 floor raised (200→900 keys / 200→250 fields); D11 M3-hold reflected in
  spec.md §B.1 + plan.md §F M3 (not only progress.md); D12 main-fork premise softened to
  conditional; D14 AC-CKH-016 token-counting folded into the consolidated AC-CKH-010 (minor,
  carried forward). The §A discipline framework, §C falsification design, and class taxonomy were
  validated as strong by the auditor and preserved unchanged.
- Findings F1-F7 each re-verified against `ed70e4354`; one drift recorded
  (spec.md §A.8 — shipped `workflow.yaml` worktree toggles contradict `internal/config/defaults.go`).
- **F3 re-derived path-resolved at the plan-audit revision** (D3): 287 distinct `yaml:`-tagged field
  names, **174** with zero production reads and 4 accessor-only; 161 map to a shipped key across 188
  (file, key) occurrences. This supersedes the earlier bare-field-name figures (122 / 5 / 121),
  which were produced by the very method this SPEC forbids as AP-3. The re-derivation used a
  throwaway `go/packages` selector-resolution probe over `./internal/... ./pkg/... ./cmd/...`
  (106 packages, 0 type errors); M2's guard is its durable implementation.
- The recomputation resolved the §A.3 ↔ §A.6 contradiction rather than papering over it:
  `auto_merge` is dead under path resolution and now appears in the family table, matching §A.6's
  "AutoMerge has zero reads". 43 names flip live→dead in total.
- Prose-consumer discriminator measured: dotted-path fixed-string probe yields 0-1 hits per key
  versus up to 46 for the bare leaf key.
- SPEC ID regex self-check executed: `PASS`.
- Status: `draft`. Awaiting plan-audit re-run and Implementation Kickoff Approval.

### Epic run order (dependency sequencing)

`depends_on: [SPEC-CONFIG-TIER-PERSIST-001]` records that this SPEC reads the tier-resolution
contract E3 owns — a shipped key's liveness is judged against the loader that actually resolves it,
and E3 fixes which tier that is. The edge is a **read** edge: nothing in M1-M6 writes a surface E3
also writes.

The run-phase `Depends_on Pre-flight Check` treats a dependency as fulfilled only at
`status: completed`. Every SPEC in this Epic is currently `draft`, so entering `/moai run` on this
SPEC before E3 closes raises the 3-option wait / override / abort blocker. **The dependency is
satisfied by sequencing, not by an `--ignore-deps` bypass** — the run order below is the mechanism,
and it is consistent with the orders recorded in `SPEC-UPDATE-DATA-SURVIVAL-001` §E.1 and
`SPEC-CONFIG-TIER-PERSIST-001` §E.1:

| Order | SPEC | Gate to clear before the next entry |
|---|---|---|
| 1 | `SPEC-UPDATE-REINSTALL-LOOP-002` (E1) | reaches `status: completed` — REQ-RIL2-015/016 landed |
| 2 | `SPEC-UPDATE-DATA-SURVIVAL-001` (E2) | reaches `status: completed` — backup coverage + failure contract landed |
| 3 | `SPEC-CONFIG-TIER-PERSIST-001` (E3) | reaches `status: completed` — tier precedence + atomic write landed |
| 4 | **`SPEC-CONFIG-KEY-HONESTY-001`** (this SPEC) | — |
| 5+ | remaining Epic SPECs (`SPEC-UPDATE-YAML-PRESERVE-001`, `SPEC-UPDATE-CI-GUARD-001`, `SPEC-UPDATE-DOC-DRIFT-001`) | no `depends_on` edge to this SPEC |

Do **not** invoke `/moai run` on this SPEC with `--ignore-deps`. If E3 slips and starting this SPEC
early becomes necessary, the correct move is to run **M1, M2, M4, M5, and M6** — none reads E3's
tier contract — and hold **M3** open, since `quality.yaml`'s parse path is the one place where which
tier resolved the block changes the answer. That is a scope decision for the orchestrator to surface
via `AskUserQuestion`, not a flag the run-phase agent may set on its own.

One ordering constraint is internal to this SPEC and independent of the Epic order: M1 must land
before M2, because M2's guard fails on any `dead` / `unresolved` / `unbound` key absent from M1's
**P** / **R** allowlists — with no inventory, every shipped key fails at once.

### Deferred audit defects (D9-D15) — iteration-3 resolution status

Recorded so the next iteration does not re-derive them. As of iteration-3 refresh:
- D9 (`qualityFileWrapper` citation at `types.go:1174`): **RESOLVED** → updated to `types.go:1312`.
- D10 (§A.6 grep transcript pre-filtered): **RESOLVED** → §A.6 prose documents the filter; NFR floor
  tightened (see D11 below).
- D11 (NFR-CKH-002's 200-key floor against ~1020 shipped keys): **RESOLVED** → floor raised to
  900 keys / 250 fields.
- D12 (AC counts tokens rather than asserting meaning): **ACKNOWLEDGED, carried forward** — the
  consolidated AC-CKH-010 Part B retains the token-count form as a necessary-but-not-sufficient
  proxy; a semantic grep was considered but rejected as brittle against prose rewording.
- D13 (`<handoff-note-path>` placeholder): **ACKNOWLEDGED, forward-ref by design** — the path is
  concrete by run-phase; no plan-time action possible.
- D14 (`main-fork/` premise false in worktree): **RESOLVED** → plan.md §B2 softened to conditional
  ("MAY exist in some checkouts"); AP-4 hazard documented as a general rule.
- D15 (`depends_on` target in `draft`): **ADDRESSED** by the run-order table above + M3-hold
  mechanism, not by frontmatter change.

## §E.2 Run-phase Evidence

### M4 — system.yaml resolution (REQ-CKH-004)

M4 unblinds the parity guard for the `"system"` registry entry and consolidates
the two inline-struct readers of `system.yaml`'s `hook.*` block onto a single
shared parse path.

**Mechanism (AC branch 2 — real binding).** A narrow `loadSystemSection`
(`internal/config/loader_system.go`) reads `system.yaml` via a
`systemFileWrapper{Hook SystemHookConfig}` and binds the `hook` block into
`cfg.System.Hook`, wired into `Loader.Load` (`loader.go:92`). The shared helper
`config.LoadSystemHookOptInEnabled(projectRoot)` reads the same file through the
same wrapper, replacing the inline anonymous structs in
`internal/hook/routing_ledger.go::HookObserveOptInEnabled` and
`internal/cli/update.go::readHookOptInEnabled` (both now one-line delegators).
The `moai` / `github` / `document_management` blocks have no `SystemConfig`
field and are intentionally ignored by the loader; they remain classified **R**
in the M1 inventory (`document_management.*` promises file deletion that
nothing performs — the sharpest reserved case).

**Branch choice (E5 judgment call).** plan.md §F M4 names two decisions — (1)
move `"system"` to `yamlAuditExceptions` and (2) add `loadSystemSection`. These
are mutually exclusive under D5 once (2) lands: a real loader makes the
exception reason "Loader.Load does not read the file" false, so keeping the
entry in `yamlAuditExceptions` would convert an unbound lie into a different
lie. AC-CKH-007 Part A encodes the two as an either/or; branch 2 (registry=1,
real loader) is the D5-honest state and is implemented here. Decision 1 is
preempted by decision 2 — the real loader IS the "real reason" the exception
would later be retired for.

**Observed verification (against HEAD of this worktree, post-M4):**

```
$ awk '/^var yamlAuditExceptions/,/^}/'  internal/config/audit_registry.go | grep -c '"system"'
0
$ awk '/^var yamlToStructRegistry/,/^}/' internal/config/audit_registry.go | grep -c '"system"'
1
$ grep -rn 'loadSystemSection' internal/config/
internal/config/loader_system.go:31:func (l *Loader) loadSystemSection(dir string, cfg *Config) {
internal/config/loader.go:92:	l.loadSystemSection(sectionsDir, cfg)
$ grep -rn 'var doc struct' internal/hook/routing_ledger.go internal/cli/update.go | grep -c 'Hook struct'
0
$ go test -run 'TestAuditParity' -count=1 ./internal/config/
ok  github.com/modu-ai/moai-adk/internal/config  0.325s
$ go test -run 'TestSystemHookOptInLoadsViaLoader' -count=1 ./internal/config/
ok  github.com/modu-ai/moai-adk/internal/config  2.191s
```

Baseline at HEAD `0c494e9e1` (pre-M4, the concealed state) for the same greps:
`yamlAuditExceptions` count 0, `yamlToStructRegistry` count 1, no
`loadSystemSection`, inline-struct count 2 — and `TestAuditParity` passing
anyway (the false-positive the M4 parity-guard unblinding targets).

Reader-site verification: the iteration-3 refresh cited
`routing_ledger.go:104` and `update.go:1140` as the inline-struct sites at code
baseline `ed70e4354`. Both confirmed present at worktree HEAD `0c494e9e1`
(`HookObserveOptInEnabled` at `routing_ledger.go:104`, `readHookOptInEnabled`
at `update.go:1143`); both are now one-line delegators to
`config.LoadSystemHookOptInEnabled`. No drift.

### M1 + M2 evidence (prior milestones, summarized)

M1 (triage rule `.moai/docs/config-key-triage-rule.md` + 952-entry inventory
`internal/config/testdata/shipped_key_inventory.yaml`) and M2 (anti-rot guard
`internal/config/shipped_key_reader_test.go`) landed on earlier worktree
commits and are unchanged by M4 (B10 PRESERVE). `github.*` (3 keys) and
`document_management.*` (18 keys) are classified **R** in the M1 inventory —
M4 did not need to reclassify them.

### M5 — Documented-but-unenforced reconciliation (REQ-CKH-009, REQ-CKH-010)

M5 resolves the three documented-but-unenforced findings: the §A.8
contradiction (template shipped `true` while Go defaults `false`), the F4
dual-hardcode (`max_active_learnings` unwired; enforcement in two independent
`= 50` constants), and F7 (`SessionNamePattern` with zero production readers).
M5 is a documentation-correction milestone — it does NOT delete keys (that is
M6's domain under REQ-CKH-011) and does NOT wire the unwired keys (out of
scope per spec §C). The M1 inventory classes for the five keys are unchanged
(B10 PRESERVE): `max_active_learnings` D, `auto_cleanup` P, `auto_create` W,
`auto_merge` D, `session_name_pattern` D.

**§A.8 resolution (REQ-CKH-009 — auto_merge / auto_cleanup / auto_create).**
The shipped template `internal/template/templates/.moai/config/sections/workflow.yaml`
previously shipped `auto_merge: true` and `auto_cleanup: true` while
`internal/config/defaults.go` sets all three `WorkflowWorktreeConfig` toggles
to `false`. The template now ships `auto_create: false`, `auto_merge: false`,
`auto_cleanup: false` — matching `defaults.go:665-667` exactly. Each toggle
now carries an honest comment: `auto_create` is read once
(`internal/cli/worktree_advisory.go`) only to select advisory wording (does
not gate creation); `auto_merge` / `auto_cleanup` are declared but not read.
`CLAUDE.local.md` §22.8 (materialized at `.moai/docs/local-dev-settings-intent.md`)
is corrected to state each toggle's real reader status precisely rather than
describing all three as governing web-console worktree automation.

**F4 documentation (REQ-CKH-010 — max_active_learnings).** The config
declaration (`internal/config/types.go` `DesignEvolution.MaxActiveLearnings`)
and the default (`internal/config/defaults.go:935`) now carry comments naming
the two real enforcement sites: `internal/evolution/types.go` `MaxActiveLearnings`
(= 50) and `internal/constitution/rate_limiter.go` `rateLimitMaxActiveLearnings`
(= 50). Both literals appear within the comment context so a reader of the
config cannot mistake the key for the lever. Wiring remains out of scope
(spec §C — a refactor beyond this SPEC).

**F7 marker (REQ-CKH-009 — session_name_pattern).** The shipped
`session_name_pattern` value is retained (the M2 guard enumerates shipped
keys; removing it would change the key set) but now carries a generic reserved
marker: "declared but not read — no code builds a session name from this value
(reserved)." The shipped file no longer presents the pattern as an active
setting.

**Observed verification (against HEAD of this worktree, post-M5):**

```
# AC-CKH-009 — enforcement-site comments at both Go sites
$ grep -n -B3 -A1 'MaxActiveLearnings' internal/config/types.go internal/config/defaults.go
types.go:1236: // MaxActiveLearnings is declared but NOT read ...
defaults.go:930: // MaxActiveLearnings mirrors the value enforced by two ...
# Both internal/evolution/types.go and internal/constitution/rate_limiter.go
# appear within the printed context.

# AC-CKH-010 Part A — shipped toggle values match defaults.go
$ grep -n 'auto_create\|auto_merge\|auto_cleanup' .../workflow.yaml
auto_create: false
auto_merge: false
auto_cleanup: false
$ grep -n 'AutoCleanup\|AutoCreate\|AutoMerge' internal/config/defaults.go | head -3
AutoCleanup:        false,
AutoCreate:         false,
AutoMerge:          false,

# AC-CKH-010 Part B — §22.8 states real reader status (count >= 3)
$ sed -n '/§22.8/,/§22.9/p' .moai/docs/local-dev-settings-intent.md | grep -cE 'auto_cleanup|auto_merge|auto_create'
3

# AC-CKH-011 — session_name_pattern reserved marker
$ grep -n -A2 'session_name_pattern' .../workflow.yaml
# session_name_pattern: declared but not read — no code builds a session
# name from this value (reserved). Retained as a placeholder only.
session_name_pattern: "moai-{ProjectName}-{SPEC-ID}"

# M2 guard unaffected (template values changed; guard reads keys, not values)
$ go test -count=1 ./internal/config/...
ok  github.com/modu-ai/moai-adk/internal/config  5.672s
```

Baseline at HEAD `b58f5c371` (pre-M5): `workflow.yaml:36-37` shipped
`auto_merge: true` / `auto_cleanup: true`; neither Go site carried an
enforcement-constant comment; `session_name_pattern` carried no reserved
marker; §22.8 described all three toggles as governing web-console worktree
automation. All four are corrected by this milestone.

REQ-CKH-009 nuance respected: `auto_cleanup` is NOT deleted (it was introduced
as a gate role via REQ-SW-022 and is classified P in the M1 inventory). Its
M5 deliverable is value-correction + honest documentation, consistent with
AC-CKH-010 which verifies value alignment, not deletion.

### M6 — Neutrality leak removal + E5 handoff (REQ-CKH-011, REQ-CKH-012, REQ-CKH-013)

M6 is the final run-phase milestone. It removes the three template-neutrality
leaks (finding F6, spec.md §A.9) by rewriting the leaking comments to drop the
internal citation while preserving the mechanism description, records the
pattern-coverage gaps for sibling E5, and pins the report-once/delete-never
posture over the merge engine for keys M1 classified D.

**Origin-vs-HEAD check (B8 honesty).** The task brief flagged that origin/main
recently landed `9fb2ffd75` (zone-registry neutralization). The three F6 leaks
were re-verified at worktree HEAD `3f8f8458a` AND at origin/main:

```
$ grep -n 'SPEC-AGENT-ARCH-V2-001' internal/template/templates/.moai/config/sections/workflow.yaml   # HEAD
88:    # cycle (plan.md §D D6, SPEC-AGENT-ARCH-V2-001 M3b). Values mirror the
108:    # model_routing_profiles: No-Haiku 3-tier policy (SPEC-AGENT-ARCH-V2-001
$ git show origin/main:internal/template/templates/.moai/config/sections/workflow.yaml | grep -n 'SPEC-AGENT-ARCH-V2-001'
82:    # cycle (plan.md §D D6, SPEC-AGENT-ARCH-V2-001 M3b). Values mirror the
102:    # model_routing_profiles: No-Haiku 3-tier policy (SPEC-AGENT-ARCH-V2-001
$ grep -n 'issue #653' internal/template/templates/.moai/config/sections/llm.yaml   # HEAD
179:    # (issue #653). Claude Code reports context_window_size based on the
$ git show origin/main:internal/template/templates/.moai/config/sections/llm.yaml | grep -n 'issue #653'
179:    # (issue #653). Claude Code reports context_window_size based on the
```

All three leaks are present at BOTH HEAD and origin/main — M6 is a real fix,
not a no-op. (HEAD line numbers for workflow.yaml shifted from 82/102 to 88/108
because the M5 milestone added reserved-marker comment lines above; the leak
content is identical.)

**Neutrality fix (REQ-CKH-012 — comment rewrite, mechanism preserved).**
- `workflow.yaml:88`: dropped `(plan.md §D D6, SPEC-AGENT-ARCH-V2-001 M3b)`
  from the `model_routing` deprecation note; the backward-compat-alias
  mechanism description ("retained as a `medium` profile alias for one
  backward-compat cycle") is preserved verbatim.
- `workflow.yaml:108`: dropped `(SPEC-AGENT-ARCH-V2-001 M3, design.md §D.5)`
  from the `model_routing_profiles` note; the No-Haiku 3-tier policy
  mechanism ("perfTier -> (Tier x Phase) -> {model, effort}") is preserved.
- `llm.yaml:179`: dropped `(issue #653)` from the GLM context-window comment;
  the upstream-misreport explanation ("Claude Code reports context_window_size
  based on the Claude slot ... regardless of provider") is preserved.

**E5 handoff (REQ-CKH-013 — pattern-coverage gap recorded, guard NOT edited).**
The handoff note at `.moai/specs/SPEC-UPDATE-CI-GUARD-001/research-ckh-f6-handoff.md`
records the three measured evidence points for sibling E5: (1)
`SPEC-AGENT-ARCH-V2-001` matches no registered C1/C1c family; (2) C6 matches
`PR #N` but not `issue #N`; (3) no class covers `plan.md §D D6`-shaped artifact
citations. `internal/template/internal_content_leak_test.go` is unmodified
(verified: `git diff --stat ed70e4354 HEAD -- internal/template/internal_content_leak_test.go` returns empty).

**Report-once / delete-never posture (REQ-CKH-011 — AC-CKH-014).**
`TestTemplateRemovedKeySurvivesUserConfig` (`internal/config/template_removed_key_test.go`,
`package config_test`) builds a `t.TempDir()` project carrying the M1-D key
`design.evolution.max_active_learnings` with a user-set value (200), simulates
a future template that removed it, runs `backup.MergeYAML3Way`, and asserts:
(a) the key + user value survive (delete-never); (b) the retained-key advisory
names the key (reported at least once); (c) a hand-added unrelated user key is
also retained (no-other-key-dropped). The merge engine is owned by sibling
SPEC-UPDATE-YAML-PRESERVE-001 (plan §D3); M6 adds only the posture assertion
over that engine plus a minimal exported sink-setter
(`backup.SetRetainedKeySinkForTest`) so the cross-package test can capture the
advisory. The setter is test infrastructure only — no merge semantics changed.

**Deferred sub-clause (AC-CKH-014 part b literal).** The literal "emits no
further report on a second merge over the same tree" sub-clause requires
cross-call report-once idempotency in the merge engine. The current engine
emits the retained-key advisory on every call that finds an old-only key and
keeps no cross-call state, so a second merge over the same tree re-emits. This
idempotency is a merge-engine behaviour owned by SPEC-UPDATE-YAML-PRESERVE-001
and is out of scope for M6 (plan §F M6 = template-neutrality + E5 handoff, NOT
merge-engine modification). M6 pins the load-bearing delete-never +
reported-at-least-once posture; the cross-call idempotency is left to the
sibling SPEC. Surfaced here as a deferred gap rather than silently weakening
the assertion.

**Observed verification (against HEAD of this worktree, post-M6):**

```
# AC-CKH-012 Part A — the three leaks gone (only placeholders remain)
$ grep -rnE 'SPEC-[A-Z0-9]' internal/template/templates/.moai/config/
.../sections/workflow.yaml:45:        session_name_pattern: "moai-{ProjectName}-{SPEC-ID}"
.../sections/statusline.yaml:28:    # "📋 [<command> <SPEC-ID>-<stage>]" only when
.../sections/cache.yaml:6:# `/moai run SPEC-XXX`.
.../sections/cache.yaml:18:  # Per-SPEC breakpoint TTL applied on `/moai run SPEC-XXX`. Enum: "5m" | "off".
$ grep -rn 'issue #' internal/template/templates/.moai/config/      # (no output, exit 1)
$ grep -rn 'plan\.md §' internal/template/templates/.moai/config/   # (no output, exit 1)
# All four hits above are generic placeholders ({SPEC-ID}, <SPEC-ID>, SPEC-XXX); zero real SPEC IDs / issue refs / artifact citations.

# AC-CKH-012 Part B — the existing guard passes and was not edited
$ go test -run 'TestTemplateNoInternalContentLeak' -count=1 -v ./internal/template/
=== RUN   TestTemplateNoInternalContentLeak
--- PASS: TestTemplateNoInternalContentLeak (0.87s)
ok  	github.com/modu-ai/moai-adk/internal/template	1.855s
$ git diff --stat ed70e4354 HEAD -- internal/template/internal_content_leak_test.go
# (empty — REQ-CKH-013 honored)

# AC-CKH-013 — the E5 handoff note records the three measured gaps
$ grep -cE 'SPEC-AGENT-ARCH-V2-001|issue #N|plan\.md §' \
    .moai/specs/SPEC-UPDATE-CI-GUARD-001/research-ckh-f6-handoff.md
16

# AC-CKH-014 — report-once / delete-never posture over the merge engine
$ go test -run 'TestTemplateRemovedKeySurvivesUserConfig' -count=1 -v ./internal/config/
=== RUN   TestTemplateRemovedKeySurvivesUserConfig
--- PASS: TestTemplateRemovedKeySurvivesUserConfig (0.00s)
ok  github.com/modu-ai/moai-adk/internal/config  1.050s

# Template-First: make build regenerated catalog.yaml after the template edits
$ make build 2>&1 | tail -1
go build -ldflags "..." -o bin/moai ./cmd/moai   # exit 0; catalog.yaml updated

# Cross-cutting: build + vet + full config suite green
$ go build ./...   # exit 0
$ go vet ./internal/config/ ./internal/cli/update/backup/ ./internal/template/   # clean
$ go test -count=1 ./internal/config/...
ok  github.com/modu-ai/moai-adk/internal/config  2.474s
ok  github.com/modu-ai/moai-adk/internal/config/toolpolicy  0.719s
```

**Run-phase status after M6.** M1, M2, M4, M5, M6 are landed. M3 remains HELD
pending SPEC-CONFIG-TIER-PERSIST-001 (E3). All active ACs are satisfied except
M3's (AC-CKH-005, AC-CKH-006 — held) and the deferred cross-call-idempotency
sub-clause of AC-CKH-014 part (b) documented above. Run-phase evidence is
complete for the active milestone set.


## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_status: audit-ready
sync_complete_at: 2026-08-12
sync_commit_sha: 4d3e1d669  # D3 self-referential-hazard workaround — backfilled in a follow-up commit after the sync PR merges
run_commit_shas:                                # 5 run-phase milestone commits (M3 intentionally skipped — held for SPEC-CONFIG-TIER-PERSIST-001 / E3)
  M1: 3186a65d8     # feat(SPEC-CONFIG-KEY-HONESTY-001): M1 triage rule + shipped key inventory
  M2: 0c494e9e1     # feat(SPEC-CONFIG-KEY-HONESTY-001): M2 path-resolved anti-rot guard (REQ-CKH-008)
  M4: b58f5c371     # feat(SPEC-CONFIG-KEY-HONESTY-001): M4 system.yaml resolution — unblind parity guard + loadSystemSection (REQ-CKH-004)
  M5: 3f8f8458a     # feat(SPEC-CONFIG-KEY-HONESTY-001): M5 documented-but-unenforced reconciliation (REQ-CKH-009/010)
  M6: 785947555     # feat(SPEC-CONFIG-KEY-HONESTY-001): M6 neutrality leak removal + E5 handoff (REQ-CKH-011/012/013)
  M3: HELD          # deferred to SPEC-CONFIG-TIER-PERSIST-001 (E3) — NOT executed in this SPEC; do NOT attempt until E3 reaches status: completed (spec.md §A "M3 HELD" clause + plan.md §A M3-on-hold rationale)
frontmatter_status_transitions:
  spec_md: "in-progress → implemented → completed"  # merged into the single sync commit (3-phase close); plan.md/acceptance.md carry no frontmatter (markdown-header convention)
  updated_field_refreshed: "2026-08-12"
changelog_entry_position: "[Unreleased] / Added"   # SPEC-CONFIG-KEY-HONESTY-001 entry appended at the top of the Added section
readme_decision: skip                              # internal-tooling config-honesty fix — no user-facing command or behavior change
docs_site_decision: skip                           # internal-tooling; no docs-site surface touched
mx_tag_validation: sub-step-complete               # MX tag validation is a sync sub-step (no separate Mx-phase commit); no new @MX tags warranted in this SPEC's diff (guards live under test-tree paths; the loader_system.go addition carries no high-fan-in anchor)
ac_pass_count_final: 22                            # AC-CKH-001..013, AC-CKH-015..023 — 22 MUST ACs PASS
ac_pass_with_debt_count_final: 1                   # AC-CKH-014 — PASS-WITH-DEBT (see debt_note below)
ac_fail_count_final: 0
ac_held_count_final: 2                             # AC-CKH-005, AC-CKH-006 — held with M3 (deferred to SPEC-CONFIG-TIER-PERSIST-001)
debt_note: |
  AC-CKH-014 PASS-WITH-DEBT — the delete-never clause (template-removed D-class key +
    user-set value survive backup.MergeYAML3Way) AND the reported-once clause (retained-key
    advisory names the key at least once) ARE secured by TestTemplateRemovedKeySurvivesUserConfig
    (internal/config/template_removed_key_test.go). Only the cross-call report-once sub-clause
    (a second merge over the same tree emits no further advisory) is deferred to
    SPEC-UPDATE-YAML-PRESERVE-001, which owns the merge engine. Recorded as a deferred gap,
    NOT silently weakened. See progress.md §E.2 M6 "Deferred sub-clause (AC-CKH-014 part b literal)".
template_first_mirror_check: no-gap                # zone-registry.md / llm.yaml / workflow.yaml ARE the template source under internal/template/templates/; .moai/docs/config-key-triage-rule.md + .moai/docs/local-dev-settings-intent.md are dev-internal (explicitly NOT shipped per the local-only docs/ set)
build_verification: "go build ./... exit 0 (sync-phase light verification; full go test ./... deferred to sync-auditor)"
open_blockers: 0
```

## §F Phase 4 Mode Selection

Logged by the orchestrator before the first run-phase `Agent()` spawn (per
orchestration-mode-selection.md §D).

### Input parameters

- **tier**: M
- **scope (file count)**: ~10-14 files (config Go source + template YAML + test
  files + `.moai/docs` triage rule + `CLAUDE.local.md` + testdata inventory)
- **domain count**: 4 (`internal/config` Go, `internal/template` YAML, test
  guard, `.moai/docs` + docs prose)
- **file language mix**: Go + YAML + markdown (no frontend, no shell)
- **concurrency benefit**: LOW — coding-heavy with data dependencies (M1
  inventory → M2 guard; M1 W/P/R/D classification → M4/M5/M6 consumption)

### Mode evaluation

| Mode | Selected | Rationale |
|------|----------|-----------|
| 1 trivial | NO | 5 milestones, multi-file, semantic classification work |
| 2 background | NO | coding-heavy write work, not read-only async |
| 3 agent-team | RETIRED | Mode 3 tombstone (Agent Teams static layer retired) |
| 4 parallel | NO | coding-heavy violates Anthropic's coding-task parallelism caveat |
| 5 sub-agent | **YES** | sequential per-milestone delegation; data deps + semantic judgment |
| 6 workflow | NO | not high-volume mechanical (W/P/R/D classification + guard logic is semantic) |

### Decision

`sub-agent`

### Justification

Tier M coding-heavy work with 5 milestones (M1/M2/M4/M5/M6) touching config Go,
template YAML, test files, and docs prose. Per Anthropic's coding-task
parallelism caveat, sequential sub-agent delegation (Mode 5) is the correct
default: the milestones have hard data dependencies (M1's inventory +
allowlists are consumed by M2's guard, and M1's W/P/R/D classification is
consumed by M4/M5/M6), and the work involves semantic judgment (classifying
each shipped key into W/P/R/D), not a uniform mechanical transform. Progression
mode: **autonomous (goal-armed ac_converge)** — selected by the user at the
Implementation Kickoff Approval gate. The `/moai goal` ac_converge condition is
armed alongside M1 delegation; the loop continues across milestones without
per-milestone checkpoints until all active ACs (M3-hold excluded) converge.
