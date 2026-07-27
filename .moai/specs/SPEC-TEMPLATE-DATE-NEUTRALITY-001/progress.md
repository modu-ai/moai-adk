# SPEC-TEMPLATE-DATE-NEUTRALITY-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-07-25
plan_iteration: 2
tier: L
artifacts: [spec.md, plan.md, acceptance.md, design.md, research.md, classify.sh]
requirements: 20
acceptance_criteria: 23
open_questions: 0
deferred_questions: 1
```

Plan-phase artifacts authored in the isolated worktree `.claude/worktrees/debt-clear` on branch `spec/template-date-neutrality`, based on `origin/main` at `c7309aeb6`.

**Iteration 2** — revised after plan-audit FAIL (0.55, Tier L threshold 0.85). Four user decisions absorbed as binding requirements (hybrid carve-out, DC-1 preserve, mirror-stamp preserve, CI isolated target); the fifth marker re-posed as an M2 step rather than a pre-Kickoff gate.

All 23 acceptance-criteria judgment commands were executed during plan phase; baselines are recorded verbatim in `acceptance.md` §B. The classification recipe is committed as `classify.sh` and reproduces the guard's 135-finding set exactly (verified by set diff).

Zero clarification markers remain in `plan.md`. One question is explicitly **deferred** to M2 (internal-incident prose disposition) because it requires the per-row inventory that M2 produces; it is not a blocker on Implementation Kickoff Approval.

## §E.2 Run-phase Evidence

> M1-M3 evidence rows are not recorded in this section; the entries below cover M4-M5 only.

### M4-M5 — carve-out + report cap (2026-07-27)

Commits: `0888cf9ad` (M4), `b933c28e4` (M5). Single file changed: `internal/template/internal_content_leak_test.go`.

| AC | Status | Judgment command | Actual output |
|---|---|---|---|
| AC-TDN-009 | PASS | `awk '/^func (collectLeakViolations\|isPedagogicallyAllowed\|isDateAllowlisted)/,/^}/' internal/template/internal_content_leak_test.go \| grep -cE 'LineStart\|LineEnd\|lineNo\|LineNumber'` | `0` (window 68 lines, all 3 functions captured; injection probe returned `1`, revert returned `0`) |
| AC-TDN-010 half 1 | PASS | `grep -c "limit := 50" internal/template/internal_content_leak_test.go` | `0` |
| AC-TDN-010 half 2 | PASS | injection recipe on `NOTICE.md`, then `grep -oE 'full listing: [^ ]+'` | `full listing: /var/folders/.../T/moai-template-leak-4123967095.log`; `test -f` succeeded; probe reverted |
| AC-TDN-012 | PASS | `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/ -run TestTemplateNoInternalContentLeak -count=1` | `ok`, `exit=0` (baseline was `78 occurrences, mode=strict`) |
| AC-TDN-002 (sanity) | PASS | narrow tier, same target | `ok`, `exit=0` |
| AC-TDN-013 (sanity) | PASS | `go test ./internal/template/ -run TestTemplateNeutralityAudit -count=1` | `ok`, `exit=0` |

Carve-out accounting — the 78 strict findings mapped 1:1 onto the 78 distinct `(file, date)` PRESERVE pairs of `triage.tsv`, all of class `S1-internal-date`; the S2 short-sha class contributed 0 both before and after:

| Mechanism | Categories | Findings carved |
|---|---|---|
| structural gate | DC-1 (48) + DC-4 (3) | 51 |
| content-anchored allowlist (27 entries) | DC-3 (9) + DC-2b (11) + DC-5 (7) | 27 |

Over-carving probes (each injected, observed, reverted; `git status --porcelain` clean after each):

| Probe | Expectation | Result |
|---|---|---|
| dated prose line in a skill body | flagged | flagged (`match=2029-12-31`) |
| different date inside an allowlisted file | flagged | flagged (`match=2027-05-05`) — allowlist is `(file, date)`-pinned, not file-wide |
| date repeated on a gated frontmatter line and a prose line | flagged | flagged (`match=2026-01-11`) — the structural gate does not poison per-file dedup |
| frontmatter-shaped line inside a fence | flagged | flagged (`match=2029-11-11`) — fence tracking holds |
| future attribution line in `NOTICE.md` (AC-TDN-015 A) | clean | clean |
| `**Import Date (probe)**: 2027-09-09` in `NOTICE.md` | clean | clean |
| future frontmatter bump (AC-TDN-015 B) | clean | clean |
| arbitrary dated comment in `NOTICE.md` | flagged | flagged — DC-4 gate is `(file AND attribution-line-shape)`, not whole-file |

Regression + build + lint: `go test -count=1 ./...` exit 0 (105 ok / 0 FAIL / 3 no-test); `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `golangci-lint run --timeout=2m` exit 0, `0 issues` (no new, no pre-existing baseline); `make build` exit 0 with `catalog.yaml` SHA256 unchanged (`999046874371ce42d930b097af68b7101812790d4e8963675dacf0bc8b895225` before and after), as expected for a `_test.go`-only change.

## §F Phase 4 Mode Selection

**Input parameters**

| Signal | Value |
|---|---|
| tier | L |
| scope (files) | 67 template files carry >=1 REMOVE row; 116 files carry >=1 finding |
| domain count | 2 (template markdown/yaml tree; Go guard file + CI workflow) |
| file language mix | markdown / yaml / tmpl (M3), Go (M4-M5), yaml workflow (M6) |
| concurrency benefit | LOW — M3 carries two distinct transform shapes plus a per-edit surrounding-block judgment (research.md gap G5) |

**Mode evaluation**

| Mode | Selected | Rationale |
|---|---|---|
| 1 trivial | no | 92 line-level deletions across 67 files plus a Go structural change; not a single-line edit |
| 2 background | no | write work; the orchestrator needs each milestone's result before sequencing the next |
| 3 agent-team | no | RETIRED (orchestration-mode-selection.md §C.1) |
| 4 parallel | no | coding-heavy and write-scoped; the Anthropic coding-task parallelism caveat routes this away from parallel fan-out |
| 5 sub-agent | **yes** | sequential `manager-develop` per milestone; each milestone's exit condition gates the next |
| 6 workflow | no | file count clears the ~30 soft boundary, but M3 is NOT a single uniform transform rule: 80 DC-2a stamp deletions and 12 bespoke DC-5 edits are two shapes, and the G5 orphan-block check requires per-edit judgment. §B.2 tie-breaker (multi-rule / semantic work prefers Mode 5) applies |

**Decision: sub-agent**

**Justification.** Mode 6 was evaluated seriously because the raw file count (67) clears the `~30` soft boundary, but the Mode 6 capability gate in §C.3 requires a *single uniform mechanical transform rule with no inter-file dependency*. M3 fails that on two counts: it carries two distinct edit shapes (a standalone footer-stamp line delete versus a mid-paragraph date-phrase excision), and every deletion needs the surrounding block inspected for the orphaned-header regression `research.md` gap G5 names, which no acceptance criterion mechanizes. Per the §B.2 tie-breaker, multi-rule work at any file count stays on Mode 5. Milestones M4-M6 are additionally sequential by construction (M5's report-cap change is untestable until M4 drives the strict tier to zero findings), so a sequential sub-agent chain matches the dependency graph.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
