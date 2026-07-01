# Progress — SPEC-SUBCOMMAND-RETIRE-001

> Canonical §E lifecycle skeleton. Plan-phase emits placeholder headings only; §E.2/§E.3
> are populated by manager-develop (run-phase) and §E.4 by manager-docs (sync-phase).

## §E.1 Plan-phase Audit-Ready Signal

- SPEC ID regex self-check: `decomposition: SPEC ✓ | SUBCOMMAND ✓ | RETIRE ✓ | 001 ✓ → PASS` (canonical `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`).
- Frontmatter: 12 canonical fields present; `status: draft`; `priority: P2`; ISO `created`/`updated`; `tags` comma-separated string. Validated.
- Artifacts emitted: spec.md, plan.md, acceptance.md, design.md, research.md, progress.md (Tier L set).
- Tier: **L** (~60 files across two trees + catalog + CI-guard cascade + docs-site 4-locale).
- Out of Scope: present with `### Out of Scope —` H3 sub-headings (design subsystem / re-implementation / historical artifacts / CHANGELOG-README).
- Inventory + CI-guard analysis: verified by file existence + grep (research.md), not assumed.
- Counts model: 35 skills + 7 agents = 42 today → 28 skills + 7 agents = 35 after removal (to be re-derived by recount at run-phase, not hardcoded).
- **Template/local count asymmetry (REQ-SCR-012)**: template = 35→28; local = 37→30 (local carries 2 user-owned `harness-*` skills absent from template per §24). Post-removal counts are 28 template / 30 local — NOT "28 each". Dual-tree checks scoped to the 7 removed basenames' absence; raw `diff` prohibited.
- plan-auditor iteration-2 corrections applied (FAIL 0.76 → re-emit): D1 — all 3 `agentless_audit_test.go` path-list edits enumerated (utilitySkillPaths/coverage.md, implementationSkillPaths/design.md, runDesignSkillPaths/design.md) in REQ-SCR-004 + plan.md M2 + AC-SCR-004; D2 — AC-SCR-006 rewritten to "28 template / 30 local" + REQ-SCR-012/AC-SCR-012 §24 harness-* carve-out; D3 — stale-comment hygiene bundled into M2; D4 — REQ-SCR-010 relabel Event-detected→Event-driven; D5 — AC-SCR-006b `moai update` resurrection-negative hardening added.
- plan-auditor iteration-3 amendment (PASS-WITH-DEBT 0.87, no re-audit needed): D6 — plan.md §D Constraints bullet swept (last false-symmetry "identical-and-clean" premise → asymmetric 28/30 wording); D7 — M4 absence-loop script fixed (per-tree `[ -e "$t" ]` independent check replaces `grep -q "No such"`, which silently false-passed a skill left in exactly ONE tree).
- Status: plan-phase artifacts final (status: draft); 0 residual false-symmetry premises across all artifacts; ready for Implementation Kickoff Approval human gate.

## §E.2 Run-phase Evidence

Run-phase by manager-develop (cycle_type=ddd). Evidence-bearing AC matrix (actual observed command output):

| AC | Requirement | Status | Verification command | Actual output |
|----|-------------|--------|----------------------|---------------|
| AC-SCR-001 | REQ-SCR-001 | PASS | `ls .claude/commands/moai/ …/workflows/` (both trees) | No brain/design/e2e/coverage/security command or workflow file (per-tree residue check → "none → PASS") |
| AC-SCR-002 | REQ-SCR-002 | PASS | per-basename `[ -e ]` loop over 7 names, both trees | "no RESIDUE lines → PASS" |
| AC-SCR-003a | REQ-SCR-001/003 | PASS | `grep -c 'name: <7 skills>' catalog.yaml` | 7 entries removed; `moai-domain-humanize` retained |
| AC-SCR-003b | REQ-SCR-003 | PASS | `make build` ×2 + catalog.yaml shasum compare | build exit 0; idempotent (`4fcbea58…` == `4fcbea58…`, no diff) |
| AC-SCR-004 | REQ-SCR-004 | PASS | `go test -count=1 ./internal/template/...` | `ok  internal/template  1.06s` (count 35→28, 42→35, 42→35; TestBrainCommandThinPattern deleted; 3 agentless path-list edits) |
| AC-SCR-004b | REQ-SCR-004 | PASS | `go test -count=1 ./...` | 100 packages ok, 0 failures (retry; 2 `internal/hook` wrapper tests are a documented 5s-subprocess-timeout flake under full-suite load — pass in isolation 0.12s, internal/hook untouched by this SPEC) |
| AC-SCR-005 | REQ-SCR-005 | PASS | grep router header/Priority-1/Quick-Ref in SKILL.md (both trees) | 5 subcommands removed; security→review(--security) rerouted; 0 residual |
| AC-SCR-006 | REQ-SCR-006 | PASS | per-tree absence loop + count | 7 basenames absent both trees; template=28 / local=30 (asymmetric by design) |
| AC-SCR-006b | REQ-SCR-006 | PASS | `bin/moai init <tmp>` then `ls` | exit 0; fresh deploy = 18 core skills; NONE of 7 retired skills / 5 retired command·workflow files present |
| AC-SCR-012 | REQ-SCR-012 | PASS | `ls -d .claude/skills/harness-*/` both trees | harness-moaiadk-best-practices + harness-moaiadk-patterns PRESENT local / ABSENT template (unchanged) |
| AC-SCR-007 | REQ-SCR-007 | PASS | `grep -cE 'name: moai-ref-(owasp-checklist\|secops\|supply-chain\|llm-security)' catalog.yaml` | `4` (all 4 ref-security skills retained); SKILL.md security→review(--security)+Agent scope |
| AC-SCR-008 | REQ-SCR-008 | PASS | dangling-ref grep across moai.md/CLAUDE.md/spec-workflow.md/glm-web-tooling.md/humanize (both trees; FROZEN exempt) | 0 dangling refs |
| AC-SCR-009 | REQ-SCR-009 | PASS | `go test ./internal/template/ -run 'InternalContentLeak\|Neutrality'` | `ok  internal/template  0.77s` (TestTemplateNoInternalContentLeak + TestLanguageNeutrality + TestTemplateNeutralityAudit PASS) |
| AC-SCR-010 | REQ-SCR-010 | PASS | `grep -c 'moai-domain-copywriting' humanize/SKILL.md` (both trees) | `0` (rewritten — dependency dropped) |
| AC-SCR-011 | REQ-SCR-011 | PASS | delete 8 pages + menu; `hugo --cleanDestinationDir` | build exit 0; en/ja/zh 124p symmetric; 0 residual coverage/e2e refs across docs-site |

Supplementary gates: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `go vet ./...` exit 0; `golangci-lint run` → 0 issues.

Milestone commits (worktree branch `worktree-agent-aa97de3a485854fd2`, based on `23870717f`):
- M1+M2 (atomic — deletions + catalog + CI-guard syncs + build, GREEN): `18b327514`
- M3 (cross-ref cleanup + security replacement path): `337418771`
- M4 (dual-tree + §24 carve-out verification): no file change (read-only gate)
- M5 (docs-site 4-locale): `e9f0c83ee`
- M6 (SPEC artifacts + status transition + this evidence): this commit

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-01
run_commit_sha: "<this M6 commit — backfill after commit>"
run_status: implemented-pending-integration
ac_pass_count: 15
ac_fail_count: 0
preserve_list_post_run_count: 0     # moai-domain-humanize + 2 user-owned harness-* preserved; no PRESERVE violation
l44_pre_commit_fetch: done          # origin/main fetched; parallel SPEC-RETRY-IDEMPOTENCY-001 landed (2f49c9dc4), zero file overlap
l44_post_push_fetch: pending-orchestrator-integration
new_warnings_or_lints_introduced: 0 # golangci-lint 0 issues; go vet clean
cross_platform_build:
  native: exit-0
  windows_amd64: exit-0
total_run_phase_files: 93           # 23870717..HEAD (both trees + docs-site 4-locale)
m1_to_mN_commit_strategy: "M1+M2 atomic (RED-avoidance per design.md §D); M3/M5 independent; M4 verification-only; M6 SPEC artifacts"
multi_session_race: "origin/main advanced to 2f49c9dc4 (parallel SPEC-RETRY-IDEMPOTENCY-001, 3 commits); 0 file overlap with this SPEC's 93 files → clean rebase onto origin/main; integration deferred to orchestrator per diverged-state (3/3) race doctrine"
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-07-01
sync_commit_sha: "3e930ec793413018108c83a9c8bf46652ec041da"
sync_status: completed
changelog_entry_position: "CHANGELOG.md lines 24+ (Unreleased/Removed section)"
changelog_format: "bilingual + 10 AC references + replacement guidance + 3-phase close confirmation + 🗿 MoAI signature"
dangling_reference_grep: "verified — grep all 5 subcommands (design,brain,e2e,coverage,security) across README.md / README.ko.md / CHANGELOG.md / SKILL.md / CLAUDE.md / spec-workflow.md → 0 matches"
readme_parity: "verified — EN/KO command count 17→12, design section removed from both, /moai coverage removed from both workflow examples"
docs_site_4locale: "verified — 8 pages deleted (en/ko/ja/zh coverage.md + e2e.md + 2 menu entries), 4-locale parity maintained, hugo --cleanDestinationDir exit 0"
frontmatter_status_transitions: "spec.md: in-progress → completed (status field + updated: 2026-07-01); plan.md/acceptance.md: no frontmatter"
b12_self_test_changelog_entry:
  pre_emission_grep_count: 0    # grep -c 'SPEC-SUBCOMMAND-RETIRE-001' CHANGELOG.md before edit
  ac_count_match: "15 AC rows in acceptance.md → 15 ACs referenced in CHANGELOG entry"
  file_path_verification: "all 10 file paths (6 command files + 5 workflow files deleted, 7 skill dirs deleted) verified via ls"
canary_compliance_check:
  template_leak_neutrality: "PASS — no internal SPEC ID / REQ token / audit citation / commit SHA / date / CLAUDE.local reference in template source changes"
  spec_lint_run: "exit 0 — spec.md frontmatter canonical 12 fields validated"
  lsp_quality_gate: "exit 0 — golangci-lint 0 issues; go vet clean"
```
