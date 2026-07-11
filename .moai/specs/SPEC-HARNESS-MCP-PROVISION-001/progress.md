# Progress — SPEC-HARNESS-MCP-PROVISION-001

> Lifecycle progress ledger. §E is parser-load-bearing (era.go string-matches the
> literal `§E.2` / `§E.3` / `§E.4` heading tokens + `sync_commit_sha`). Do NOT rename
> the §E.N headings. Plan-phase populates §E.1 only; §E.2-§E.4 are placeholder
> headings owned by manager-develop (run) and manager-docs (sync).

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-07-11
plan_version: 0.1.3
tier: M
artifacts: [spec.md, plan.md, acceptance.md, progress.md]
req_count: 11
ac_count: 16
depends_on: [SPEC-PROJECT-HARNESS-BRIDGE-001]
depends_on_state: "SATISFIED — SPEC-PROJECT-HARNESS-BRIDGE-001 status: completed (measured 2026-07-11). Phase 0.5 depends_on gate does NOT block run-phase entry."
open_clarifications: []
resolved_clarifications:
  - marker: "mcp-matrix config surface"
    decision: "standalone template DATA RESOURCE (internal/template/templates/.moai/config/sections/mcp-matrix.yaml, read as prose-context by project/doc-generation.md) — NOT a new Go config section; no typed loader / struct field. Keeps the SPEC doc/config-only."
  - marker: "doctor manifest-mcp validate-vs-tolerate"
    decision: "TOLERATE-ONLY, zero Go change. Verified: doctor.go + applier.go use plain json.Unmarshal with no DisallowUnknownFields; v4manifest/validate.go checks only required fields. AC-HMP-010 encodes documented-tolerance grep + regression guard grep -c DisallowUnknownFields internal/harness/v4manifest/*.go == 0. Active mcp-block validation deferred to a follow-up Go SPEC."
notes: >
  SPEC 2 of the 3-SPEC Project-Harness Pipeline Epic. Doc/config-only (markdown+yaml);
  no Go code. Phase 0.5 Depends_on pre-flight is SATISFIED —
  SPEC-PROJECT-HARNESS-BRIDGE-001 is status: completed (measured 2026-07-11); the gate
  does NOT block run-phase entry. (v0.1.0-v0.1.2 claimed BRIDGE-001 was "currently draft,
  so this gate WILL block" — an unobserved state claim, corrected at v0.1.3 per
  verification-claim-integrity.md §1.1.)
  Plan-audit fixes applied at v0.1.1: both clarifications resolved (markers struck from
  plan.md); AC-HMP-014 added for REQ-HMP-003 (write-on-approval coverage gap);
  AC-HMP-010 de-vacuumed (grep+guard, dropped repo-wide doctor smoke); Epic artifact
  numbering corrected (MCP fragment = artifact 7 OPTIONAL; verify skill = artifact 6
  mandatory, SPEC-HARNESS-VERIFY-PROMOTE-001); AC-HMP-015 added for the harness-builder
  "exactly 5" prose reconciliation.
  v0.1.2: acceptance.md §C hardened against the token-presence-vs-reachability failure
  mode — 7 sub-checks were measurably vacuous (passing at baseline with ZERO
  implementation: AC-002=6, AC-003=7, AC-005=2, AC-007 additive=2 / env-var=1,
  AC-009=8 (headline conditional-emission!), AC-010=1); every AC now carries a positive
  check that FAILS on the unmodified tree, compound-alternation-as-sole-evidence is
  eliminated (one grep per clause), positional clauses are verified inside
  heading-delimited section ranges (§C.0 extractors), and two broken commands are fixed
  (multi-file `grep -c` printed file:count pairs, never a scalar; parity `diff` ran
  without an existence precondition). REQ set, AC↔REQ mapping, GWT scenarios, and scope
  are UNCHANGED (11 REQ / 15 AC) — only HOW each AC is verified changed.
  v0.1.3 (plan-audit iter-2, 5 fixes): D1 MUST-FIX — the /moai project ROUTER
  (.claude/skills/moai/workflows/project.md) was entirely OUT OF SCOPE (measured:
  project.md appeared 0x in spec.md, 0x in acceptance.md, 0x in plan.md §F milestones).
  An implementation passing all 15 prior ACs would have left the router still documenting
  "Phase 3.5 -> Phase 3.7" with Phase 3.6 nowhere in it — the same cross-file
  reachability gap this SPEC exists to close, one file over. Precedent confirms it is not
  hypothetical: grep -c "Phase 3.2" project.md == 0, i.e. BRIDGE-001 already shipped this
  exact drift. Fixes: plan.md §F M2 gains the router (+ template mirror) as an edit target
  with BOTH router surfaces named (Phase Routing Table row + Invocation Flow diagram
  line); user-approved scope addition backfills the missing Phase 3.2 row + diagram line
  (description READ from doc-generation.md, not invented); new AC-HMP-016 pins router
  registration (Phase 3.6 >= 2 occurrences, mechanical 3.5<3.6<3.7 row ordering, Phase 3.2
  backfill >= 2, sub-skill pointer); AC-HMP-011 parity loop extended to the router pair
  (3 -> 4 files). D2 — AC-HMP-012's prohibition filter did not cover the parenthetical
  negation form "(NOT `.moai/specs/`)" that plan.md §F M2 step 2 itself prescribes, so an
  implementer copying plan.md's own wording would false-FAIL; filter widened with |NOT
  and re-verified against 3 controls. D3 — AC-HMP-015 checked only 1 of 6 stale
  "5 artifact types" sites, leaving the H2 heading "(the 5 artifact types)" standing above
  a 7-artifact section while the AC PASSED; added a second reverse delta on the heading
  and enumerated all 6 sites in plan.md §F M3 step 1. D4 — struck the stale
  "moai harness doctor" from acceptance.md §E Tested bullet (contradicted AC-HMP-010 and
  §E's own Doctor-tolerance bullet; the smoke was dropped at v0.1.1). D5 — corrected the
  false BRIDGE-001 "currently draft" dependency-state claim (it is completed). D6 —
  AC-HMP-004/013 row greps now guard the missing-file case (grep -c on an absent file
  emits nothing and exits 2 — not a scalar; same non-evaluable class fixed in AC-HMP-010).
  REQ set unchanged (11 REQ); AC count 15 -> 16.
```

## §E.2 Run-phase Evidence

Run-phase executed with cycle_type=ddd (ANALYZE-PRESERVE-IMPROVE — behavior-preserving
doc/config insertion). Baseline synced first: the L1 worktree was 8 commits behind main and
carried v0.1.1 artifacts; fast-forwarded to main (`f2a71dcf8..82b758d43`) so run-phase
implemented against the audited v0.1.3 plan (16 AC, router AC-016). All edits Template-First
(edit → mirror → `make build`); every touched skill/config file byte-identical between the
local `.claude/` / `.moai/` tree and `internal/template/templates/`.

### Per-milestone what-changed

| Milestone | Commit | What changed |
|-----------|--------|--------------|
| M1 (REQ-HMP-004) | `8d69b5886` | Created `.moai/config/sections/mcp-matrix.yaml` (both trees, byte-identical): per-project-type matrix (web-frontend / mobile / backend-db rows + `universal_starter` fallback), project-type-keyed + 16-language neutral, no typed Go loader. Transitioned spec.md `draft → in-progress`. |
| M2 (REQ-HMP-001/002/003/005/006/007) | `28be79f2d` | `doc-generation.md`: inserted `## Phase 3.6 MCP Server Provisioning` between Phase 3.5 (LSP) and Phase 3.7 (dev-mode) — stack detect via `harness-spec.yaml` `external_systems`/`ui_surface`, matrix select (3-5 cap, vendor-maintained), orchestrator AskUserQuestion approval (subagent returns blocker report; per-server approval for credentialed servers), additive `.mcp.json` write at repo-root project scope with `${VAR}` secrets. `project.md` router: registered Phase 3.6 in the Phase Routing Table + Invocation Flow diagram (ordered 3.5 < 3.6 < 3.7); backfilled the missing Phase 3.2 (harness-spec.yaml emission) row + diagram line. Agent Chain Summary Phase 3.6 line added. |
| M3 (REQ-HMP-008/009/010) | `b16264f83` | `harness-builder.md`: added `### Artifact 7` (optional `.mcp.json` fragment via `artifact_type=mcp-server`) under GENERATE Output Contract — conditional emission (emit iff MCP need from `harness-spec.yaml`; else byte-identical omission), doctor-tolerant optional manifest `mcp` block (TOLERATE-ONLY, zero code change). Reconciled all six stale "5 artifact types" sites to the canonical order (5 base + verify skill artifact 6 + optional MCP fragment artifact 7), including the H2 heading and body sentence. |
| M4 (REQ-HMP-011 + discovered cascade) | `4245bf2f4` | Registered `mcp-matrix` in `internal/config/audit_loader_completeness_test.go` `acknowledgedUnloadedSections` allowlist (the codebase's sanctioned mechanism for an intentional-no-loader section — matches the SPEC's resolved "data resource, no typed loader" clarification). See Residual-risk note below. |

### AC PASS/FAIL matrix (16 AC, verbatim §C commands run post-implementation)

| AC | REQ | Status | Verification command | Actual output | Baseline → post |
|----|-----|--------|----------------------|---------------|-----------------|
| AC-HMP-001 | REQ-HMP-001 | PASS | `grep -c '^## Phase 3.6' $DG` + mechanical 3.5<3.6<3.7 | `1` ; `ORDER_OK (154<192<256)` | 0/ORDER_FAIL → 1/ORDER_OK |
| AC-HMP-002 | REQ-HMP-001 | PASS | `p36 \| grep -c -F mcp-matrix.yaml/external_systems/ui_surface` | `1 / 1 / 1` | 0/0/0 → 1/1/1 |
| AC-HMP-003 | REQ-HMP-002 | PASS | `p36 \| grep -c AskUserQuestion/subagent/blocker report` | `1 / 1 / 1` | 0/0/0 → 1/1/1 |
| AC-HMP-004 | REQ-HMP-004 | PASS | `test -f $MX/$MXT` + 4 row greps + pointer + `- { name:` absence | `MX_LOCAL_OK MX_TPL_OK` ; rows `1/1/1/1` ; pointer `1` ; `- { name:` in DG `0` | MISSING/0×4 → OK/1×4 |
| AC-HMP-005 | REQ-HMP-005 | PASS | `p36 \| grep -c 3-5 / vendor-maintained` | `1 / 1` | 0/0 → 1/1 |
| AC-HMP-006 | REQ-HMP-006 | PASS | `p36 \| grep -c credential/per-server/auto-write` | `1 / 1 / 1` | 0/0/0 → 1/1/1 |
| AC-HMP-007 | REQ-HMP-007 | PASS | `p36 \| grep -c additive/clobber/.mcp.json/'${'/literal` + specs-write absence | `1 / 1 / 1 / 1 / 1` ; absence `0` | 0×5 → 1×5 |
| AC-HMP-008 | REQ-HMP-008 | PASS | `grep -c '^### Artifact 7' $HB` + `a7 \| grep -c artifact_type=mcp-server/.mcp.json` | `1 / 1 / 1` | 0/0/0 → 1/1/1 |
| AC-HMP-009 | REQ-HMP-009 | PASS | `a7 \| grep -c byte-identical/omit/external_systems` | `1 / 1 / 1` | 0/0/0 → 1/1/1 |
| AC-HMP-010 | REQ-HMP-010 | PASS | `a7 \| grep -c doctor/tolerat` + `DisallowUnknownFields` absence (v4manifest + decode sites) | `1 / 1` ; `0 / 0` | 0/0 clause → 1/1 ; guard stays 0/0 |
| AC-HMP-011 | REQ-HMP-011 | PASS | 4-pair existence-checked `diff -q` | `PARITY OK ×4` | MISSING(mcp-matrix)+OK×3 → PARITY OK×4 |
| AC-HMP-012 | REQ-HMP-011 | PASS | prohibition-aware `.moai/specs/` write grep + `p36 .mcp.json` | unfiltered `0` ; `.mcp.json` `1` | 0/0 → 0/1 |
| AC-HMP-013 | REQ-HMP-011 | PASS | `make build` + `go test ./internal/template/...` + SPEC-ID leak + template rows + no-privileged-lang | `exit=0` ; `ok internal/template 1.435s` ; leak `0` ; rows `1/1/1` ; privileged `0` | leak 0 preserved; rows 0→1 |
| AC-HMP-014 | REQ-HMP-003 | PASS | `p36 \| grep -c '(on\|upon\|once\|after) approval'/project scope/repo-root` | `1 / 1 / 1` | 0/0/0 → 1/1/1 |
| AC-HMP-015 | REQ-HMP-008 | PASS | `grep -c 'exactly 5 artifact types'` + `'(the 5 artifact types)'` (reverse) + `goc \| grep -c artifact 7/optional` | `0 / 0` (reverse) ; `1 / 1` | 1/1 → 0/0 ; 0/0 → 1/1 |
| AC-HMP-016 | REQ-HMP-001 | PASS | router `Phase 3.6` ≥2 + mechanical row order + `Phase 3.2` ≥2 + 3.6→doc-generation.md | `2` ; `ROUTER_ORDER_OK (55<56<57)` ; `2` ; `1` | 0/FAIL/0/0 → 2/OK/2/1 |

### PRESERVE invariants (all still hold at exit — DDD behavior preservation)

| Invariant | Status | Evidence |
|-----------|--------|----------|
| NO-SPEC scope guard (project flow never writes `.moai/specs/**`) | PASS | AC-012 prohibition-aware grep `0`; existing Phase 3.2 `MUST NOT` guard intact; Phase 3.6 write target is repo-root `.mcp.json` |
| Pre-MCP harness GENERATE byte-identical when no MCP need declared | PASS | Artifact 7 is a CONDITIONAL branch (AC-009 `byte-identical`/`omit` clause); no unconditional write added |
| `builder-harness` `artifact_type=mcp-server` internals unchanged | PASS | `.claude/agents/moai/builder-harness.md` not modified (reused, not reimplemented) |
| `harness-spec.yaml` schema + adaptive interview (BRIDGE-001-owned) unchanged | PASS | consumed via `external_systems`/`ui_surface`/`verification`; no schema edit |
| Lenient `json.Unmarshal` manifest decoder (doctor tolerance) unchanged | PASS | `DisallowUnknownFields` count `0` in `internal/harness/v4manifest/` + both decode sites |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: audit-ready
run_complete_at: 2026-07-11
run_commit_sha: 4245bf2f4   # M4 (final implementation commit); progress.md evidence commit lands after this
cycle_type: ddd
ac_pass_count: 16
ac_fail_count: 0
ac_total: 16
preserve_list_post_run_count: 5   # all 5 PRESERVE invariants still hold at exit
new_warnings_or_lints_introduced: 0   # go vet clean; gofmt clean; template neutrality + internal-content-leak guards green
cross_platform_build:
  host_darwin: "go build ./... exit 0"
  windows: "GOOS=windows GOARCH=amd64 go build ./... exit 0"
non_regression:
  full_suite: "go test ./... exit 0 — 96 ok packages, 0 FAIL, 3 no-test-file"
  neutrality: "go test ./internal/template/... — ok (template-neutrality-check + internal_content_leak_test green)"
  spec_id_leak_in_template_tree: 0
doctor_tolerance: "documented-tolerance grep (a7 doctor=1, tolerat=1) + DisallowUnknownFields==0 in v4manifest/*.go and both decode sites (applier.go, doctor.go)"
total_run_phase_files: 11   # mcp-matrix.yaml(x2) + doc-generation.md(x2) + project.md(x2) + harness-builder.md(x2) + spec.md + audit_loader_completeness_test.go + progress.md
m1_to_mN_commit_strategy: "5 milestone commits M1..M4 + this progress.md evidence commit on the isolated L1 worktree branch; direct-to-branch (no per-milestone push); orchestrator owns push at session end (parallel-session commits stacked)"
milestone_commits:
  M1: 8d69b5886   # externalize MCP matrix to config + spec.md draft->in-progress
  M2: 28be79f2d   # /moai project Phase 3.6 + router registration
  M3: b16264f83   # harness GENERATE optional artifact 7 + doctor tolerance
  M4: 4245bf2f4   # acknowledge mcp-matrix as intentional no-loader section (discovered cascade)
l44_pre_commit_fetch: "N/A — orchestrator owns push/fetch this session (parallel-session commits stacked); Pre-Spawn Sync Check is orchestrator-side"
l44_post_push_fetch: "N/A — push deferred to orchestrator"
baseline_sync: "L1 worktree fast-forwarded f2a71dcf8..82b758d43 (8 commits behind main) before run-phase, so implemented against audited v0.1.3 (16 AC)"
scope_note: "M4 touched one Go test file (audit_loader_completeness_test.go allowlist) — a mechanical cascade of the M1 config-file addition, NOT MCP feature code. See §E.2 Residual-risk / manager-develop E-report. plan.md §F M4's 'no Go code changed' assertion did not anticipate the TestAuditLoaderCompleteness guard."
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — owned by manager-docs; carries sync_commit_sha>_
