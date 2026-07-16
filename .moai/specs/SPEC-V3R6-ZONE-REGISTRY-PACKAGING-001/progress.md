# Progress — SPEC-V3R6-ZONE-REGISTRY-PACKAGING-001

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
spec_id: SPEC-V3R6-ZONE-REGISTRY-PACKAGING-001
tier: S
phase: plan
iteration: 2               # plan-audit iter-1 FAIL 0.62 → D1(BLOCKING)/D2/D3 resolved
artifacts:
  spec_md: present        # §A context (A.4.1 six-guard inventory + A.4.2 D1 exemption), §B 7 GEARS requirements, §C exclusions, §D AC map
  plan_md: present        # reversibility-ordered (Strategy A/B decision leads), M1-M4 milestones (M2 = both guard-test edits), §F @MX targets
  acceptance_md: present  # AC-ZRP-001..007, edge cases, DoD
requirements: 7           # REQ-ZRP-001..007 (GEARS) — +REQ-ZRP-007 governance-token file-level exemption
acceptance_criteria: 7    # AC-ZRP-001..007 — +AC-ZRP-007; AC-ZRP-004 D3 fix (dropped unobservable spec-lint sub-assertion); AC-ZRP-005 six-guard
needs_clarification: 0
neutralization_strategy: A   # sanitized-pair; neutralize mirror only (source untouched)
forbidden_tokens_found: 4    # 2 SPEC-IDs (lines 1009,1017) + 2 dates (lines 629,630); all other §25 classes = 0
const_token_occurrences: 119 # legitimate registry content (115 entries + 4 xrefs); §C forbids removal → file-level exemption (REQ-ZRP-007)
ci_guards_analyzed:          # SIX guards (iter-1 missed #5/#6)
  - internal_content_leak_test.go        # C1 whole-tree SPEC-ID = default gate; C1b (CONST-V3R) is skillBodyScoped → NOT firing on .claude/rules/
  - template_neutrality_audit_test.go    # binary classes all 0; C2 file-allowListSet already includes zone-registry (line 138) — precedent for D1
  - rule_template_mirror_test.go         # opt-in allowlist; zone-registry NOT enrolled → divergent mirror OK
  - sanitized_pair_parity_test.go        # REQ-ZRP-006 registration target; CONST not normalized but identical both sides → no drift
  - rule_provenance_audit_test.go        # D1 governance-token: 119 CONST fire whole-tree, unconditional; isPedagogicallyAllowed per-(path,token) → file-level carve-out (REQ-ZRP-007)
  - rule_date_provenance_audit_test.go   # D2 date: 20[0-9]{2} broader than leak-strict; only 2 dates (629/630) → removed by neutralization → green, no exemption
d1_exemption_mechanism: file-level-single-path  # governanceTokenFileAllowlist{zone-registry.md}; NOT class-wide; parallel to neutrality C2 allowListSet
deploy_mechanism: generic-fs-walk        # deployer.go fs.WalkDir + embed.go all:templates; no catalog entry needed
guard_test_edits_in_scope: 2             # sanitizedPairPaths +1 line; governanceTokenFileAllowlist +1 path (both additive, file-scoped)
mx_targets: none-hard                     # doctrine-markdown fix; 1 optional @MX:NOTE on registry addition
spec_id_precheck: PASS                    # ^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$
```

_Plan-phase artifacts revised (iter-2). D1/D2/D3 resolved. Ready for plan-auditor re-review._

## §E.2 Run-phase Evidence

Strategy A (sanitized-pair; source untouched) implemented as `cycle_type=tdd`. RED baseline
first: a byte-verbatim mirror + pair registration made the guards FAIL exactly as the six-guard
analysis predicted (leak: 2 SPEC-IDs; date: 2 dates; governance: 121 tokens), then GREEN via
neutralization + single-path exemption. Verbatim command output persisted under
`.moai/state/verify/` (worktree). All evidence rows below are actually-observed outputs.

| AC | Requirement | Command / Evidence | Actual Output | Status |
|----|-------------|--------------------|---------------|--------|
| — (RED) | guards detect leaks | `go test -run 'TestRuleProvenanceAudit\|TestTemplateNoInternalContentLeak\|TestRuleDateProvenance' …` on verbatim mirror → `.moai/state/verify/00-RED-baseline.log` | 3 guards FAIL: C1 2×SPEC-ID; RULE_DATE_PROVENANCE_LEAK 2×date; RULE_GOVERNANCE_TOKEN_LEAK 121 (119 CONST+2 SPEC); `TestSanitizedPairParity` PASS | PASS (RED observed) |
| AC-ZRP-001 | REQ-ZRP-001 | `grep -cE` leak/date/binary classes on mirror → `01-mirror-tokens.log` | C1 SPEC-ID=0; ISO-date(202[6-9])=0; broad-date=0; neutrality-binary=0 | PASS |
| AC-ZRP-002 | REQ-ZRP-002 | `test -f` + `grep -coE '^- id: CONST-'` mirror vs source → `01-mirror-tokens.log` | mirror EXISTS; CONST entries 115==115; CONST-V3R occurrences=119 | PASS |
| AC-ZRP-003 | REQ-ZRP-003 | `make build` → `09-make-build.log`; embedded-FS proven transitively by AC-004 | make build exit=0; `bin/moai` built (v3.0.0-rc12); catalog.yaml unchanged (no cascade) | PASS |
| AC-ZRP-004 | REQ-ZRP-004 | `bin/moai init $(mktemp -d)/proj` + 3 CLI cmds → `10-moai-init.log`..`13-spec-lint.log` | deployed file present (40254B, 119 CONST, 0 SPEC-ID leak); `constitution list` exit 0 returns CONST-V3R2-001.. table, no load error; `doctor` exit 0 → `✓ Constitution Registry  registry OK — 115 entries (71 Frozen, 44 Evolvable)`, `not found`=0; `spec lint` exit 0, registry file EXISTS (absence-skip unreachable) | PASS |
| AC-ZRP-005 | REQ-ZRP-005 | `go test ./internal/template/...` (default) → `03-template-pkg.log`; strict tier → `04`/`06`; pinned D1/D2 → `05` | default full package exit=0 (all six guards green); pinned `TestRuleProvenanceAudit`/`TestRuleDateProvenance`/`…RecurrenceBackstop`/`TestSanitizedPairParity` all PASS; **strict tier exit=1 for 132 PRE-EXISTING unrelated S1-date leaks (0 in this mirror; count 132 WITH mirror == 132 WITHOUT mirror — provably mirror-neutral)** | PASS-WITH-DEBT |
| AC-ZRP-006 | REQ-ZRP-006 | `go test -run TestSanitizedPairParity …` → `05-pinned-guards.log` | PASS, no SANITIZED_PAIR_PARITY_DRIFT (4-token reword within tolerance, net one-sided=0) | PASS |
| AC-ZRP-007 | REQ-ZRP-007 | `grep -A3 governanceTokenFileAllowlist` → `14-single-path.log`; guard + backstop → `03`/`05` | allowlist = exactly 1 path key (`zone-registry.md`); `TestRuleProvenanceAudit` PASS with 119 CONST present; `TestRuleProvenanceRecurrenceBackstop` PASS (raw-regex `CONST-V3R6-099` leaky case unaffected → still fires elsewhere) | PASS |

Invariants:

| Invariant | Evidence | Status |
|-----------|----------|--------|
| Source `.claude/rules/moai/core/zone-registry.md` unchanged (Strategy A) | `git diff --stat` source → empty → `02-source-unchanged.log` | PASS |
| No CONST-* entry added/removed/semantically changed | 115 entries + 119 occurrences preserved (source==mirror) | PASS |
| Exemption single-path, not class-wide | 1 allowlist key; governance guard still fires elsewhere (backstop) | PASS |
| No catalog.yaml / manifest cascade | `git diff --stat internal/template/catalog.yaml` empty post-`make build` | PASS |
| `go build ./...` / `go vet ./internal/template/...` clean | exit 0 / exit 0 → `07-go-build.log` / `08-go-vet.log` | PASS |

Neutralization diff summary (4 forbidden tokens, mirror only):
1. `2026-05-04` (src line 629) — removed; comment prose kept (`new workflow rules;`).
2. `2026-05-09` (src line 630) — removed; comment prose kept (`model-specific threshold revision:`).
3. `SPEC-V3R6-HARNESS-RUNTIME-RECOVERY-001` (src line 1009) — reworded to `# (first V3R6 modern-era entry)`.
4. `SPEC-V3R6-HOOK-RECOVERY-SIGNAL-001` (src line 1017 clause tail) — reworded to `…deferred to a future SPEC)`; rule meaning intact.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-17
run_commit_sha: 07c5fe706e4aee9d807175a6c1c8f54cd522bdd9
run_status: complete
ac_pass_count: 6                # AC-ZRP-001/002/003/004/006/007 full PASS
ac_fail_count: 0
ac_pass_with_debt_count: 1      # AC-ZRP-005 (strict-tier sub-assertion is pre-existing out-of-scope debt)
preserve_list_post_run_count: 0 # no behavior-preservation list (packaging fix)
l44_pre_commit_fetch: performed-by-orchestrator   # pre-spawn sync check owned by orchestrator
l44_post_push_fetch: n/a-not-pushed               # run-phase commit stays on feature branch in worktree
new_warnings_or_lints_introduced: 0               # go vet clean; go build clean
cross_platform_build:
  darwin_arm64: pass            # make build exit 0 on host (darwin)
  note: single-platform host build; no cross-compile invoked (Tier S doctrine-markdown packaging)
total_run_phase_files: 3        # 1 new mirror (md) + 2 additive Go test edits (progress.md excluded — audit artifact)
m1_to_mN_commit_strategy: single-commit   # Tier S; one feat() run-phase commit on feature branch
neutralization_strategy: A      # sanitized-pair; source untouched
forbidden_tokens_neutralized: 4 # 2 SPEC-IDs + 2 dates (mirror only)
const_tokens_preserved: 119     # all CONST-V3R occurrences kept green via single-path exemption
six_guards_default_tier: green  # go test ./internal/template/... exit 0
strict_tier_status: pre-existing-red-132-unrelated-dates-mirror-neutral  # §C out-of-scope; not introduced by this SPEC
```

**Residual risk / debt (AC-ZRP-005 PASS-WITH-DEBT)**: the `MOAI_TEMPLATE_LEAK_STRICT=1` strict-tier
leak command exits 1 because of **132 pre-existing S1-internal-date leaks across ~40 unrelated
template files** (agents, skills, NOTICE.md, etc.). This is provably NOT introduced by this SPEC:
the strict-tier occurrence count is **132 with the mirror and 132 without it** (evidence
`04-strict-leak.log` vs `06-strict-without-mirror.log`), and 0 of the 132 are in this mirror
(mirror has 0 ISO dates post-neutralization). SPEC §C explicitly scopes this SPEC's strict-tier
obligation to "the mirror is clean under both tiers" (satisfied) and out-of-scopes any strict-tier
enforcement flip. Fixing the 132 dates (40+ files) is a separate, non-Tier-S concern. The default
tier — the gate that runs on every `go test ./...` — is fully green. This is an AC-scoping defect
(the plan-auditor assumed the strict baseline was green; it was already red) surfaced for a follow-up
SPEC, NOT a defect in this deliverable and NOT resolved by weakening any guard.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-07-17
sync_commit_sha: pending-backfill-see-followup-commit
sync_status: complete
changelog_entry_position: "[Unreleased] › Fixed (new subsection)"
frontmatter_status_transitions:
  spec_md: "in-progress → completed (merged in-progress → implemented → completed on sync commit)"
  progress_md: "in-progress → completed"
b12_self_test_a: "grep -c 'SPEC-V3R6-ZONE-REGISTRY-PACKAGING' CHANGELOG.md == 1 after emission (PASS — single entry, no duplicate)"
b12_self_test_b: "acceptance.md SSOT = 7 AC groups (AC-ZRP-001..007); CHANGELOG references 6 PASS + 1 PASS-WITH-DEBT = 7 (PASS — match)"
b12_self_test_c: "file paths in CHANGELOG verified via ls (.claude/rules/moai/core/zone-registry.md, internal/template/templates/.claude/rules/moai/core/zone-registry.md present) (PASS)"
canary_compliance_check:
  spec_lint_delta: "not independently re-run in sync-phase (Tier S doctrine-markdown packaging fix; no code touched)"
  touched_pkg_tests: "unchanged from §E.3 (go test ./internal/template/... exit 0)"
```

Sync-phase scope (Tier S minimal): CHANGELOG `[Unreleased] › Fixed` entry + spec.md
frontmatter `status: in-progress → completed` (`updated:` refreshed to 2026-07-17) +
this §E.4 block. No README/docs-site work (internal packaging fix, not user-facing
feature). `sync_commit_sha` is backfilled in a tiny follow-up commit per the
SHA-placeholder self-reference exemption (spec-frontmatter-schema.md § Status
Transition Ownership Matrix), matching the run-phase e738aca9f precedent.

## §F Phase 4 Mode Selection

Input parameters: tier=S; scope≈4 files (1 markdown mirror + 2 additive Go test edits + neutralization of 1 source doctrine file); domain count=1 (template packaging); file language mix=markdown mirror + Go test edits; concurrency benefit=LOW (coding-heavy, sequential).

Mode evaluation:
- trivial: not selected — multi-file semantic change, not a typo
- background: not selected — has writes
- agent-team: not selected — RETIRED
- parallel: not selected — single domain, coding-heavy (Anthropic coding-task parallelism caveat)
- workflow: not selected — <30 files, not a uniform mechanical transform
- sub-agent: SELECTED — default fallback for Tier S coding work

Decision: sub-agent

Justification: A Tier S packaging fix touching one domain (template mirror + two additive guard-test edits + neutralization) is coding-heavy with low parallelism benefit; per Anthropic's coding-task parallelism caveat the sequential sub-agent (Mode 5) is the correct envelope. Implementation Kickoff Approval was obtained (explicit user approval) before run-phase entry. cycle_type=tdd per quality.yaml development_mode.
