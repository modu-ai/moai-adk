# SPEC-SIMPLICITY-AUDIT-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase complete. Artifacts authored by `manager-spec`:

- `spec.md` — 12 canonical frontmatter fields + `era: V3R6` + `tier: S`; §A background, §B research (4 items, Read/Grep-backed), §C REQ-1..REQ-4 (GEARS), §3 inline AC (8 ACs, binary), §J exclusions (7 `### Out of Scope` H3 incl. the mandated LLM-as-judge benchmark exclusion), §H cross-references.
- `plan.md` — §A context, §B application-point verdict (`--lean` mode CHOSEN over standalone-skill / sync-auditor-dimension / anti-patterns-expansion) + §B.1 `@MX:DEBT` connect sub-verdict, §C pre-flight, §D Tier S decision, §E self-verification checklist, §F 5 milestones (M1-M5), §G irony-guard anti-patterns (AP-SA-001..008), §H cross-references.
- `progress.md` — this file.

Tier: **S** (2 files touched: `review.md` live + template mirror; zero Go; AC inline). plan-auditor PASS threshold: 0.75.

SPEC ID self-check: `SPEC ✓ | SIMPLICITY ✓ | AUDIT ✓ | 001 ✓ → PASS` against `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`.

Status: `draft`. Awaiting plan-auditor (Phase 0.5) + Implementation Kickoff Approval before `/moai run`.

_<plan-phase complete; run-phase pending>_

## §E.2 Run-phase Evidence

Run-phase cycle: TDD (RED-GREEN-REFACTOR adapted for skill-prose). RED = the AC verify-greps failed before the edit (`grep -c '\-\-lean' review.md` → 0, all 5 tags MISSING). GREEN = the same greps pass after the M1-M4 edits. REFACTOR = irony-guard self-check (§G AP-SA-001..008): the mechanism landed as one flag-mode of `review.md` + its mirror — no new skill, no new agent, no new config, no new lint rule, no new Go. Files touched: `review.md` (live) + its template mirror only.

### AC Binary PASS/FAIL Matrix

| AC | Status | Verify Command | Actual Output |
|----|--------|----------------|---------------|
| AC-SA-001 | PASS | `grep -c '\-\-lean' review.md`; `grep -c 'over-engineer' review.md`; check EXCLUDES boundary | `8` and `6` (both ≥1); `--lean` section line 191 states "by EXCLUDING correctness bugs, security findings, and performance findings" |
| AC-SA-002 | PASS | `for t in 'delete:' 'stdlib:' 'native:' 'yagni:' 'shrink:'; do grep -q "$t" review.md \|\| echo MISSING $t; done`; check output format | prints nothing (all 5 present); format `L<line>: <tag> <what to cut>. <replacement>. [path]` at line 223 |
| AC-SA-003 | PASS | `grep -q 'net: -' review.md && grep -q 'Lean already. Ship.' review.md` | both PRESENT (`net: -<N> lines possible` / `net: -<N> lines, -<M> deps possible` + literal `Lean already. Ship.`) |
| AC-SA-004 | PASS | `stdlib:`/`native:` generic phrasing + negative-token scan of the `--lean` section | line 216: `stdlib:` = "the language's standard library", `native:` = "a platform-native feature"; negative-token scan (npm/pip/cargo/gopls/node_modules/go.mod/package.json/requirements.txt/crates/pypi/nuget/gem) → PASS: no single-language bias token in `--lean` section |
| AC-SA-005 | PASS | check read-only/advisory/no-verdict + Phase 5 routing | line 253: "applies NO fixes, modifies NO files, and renders NO PASS/FAIL verdict"; "Remediation routes through the existing Phase 5 Next Steps" |
| AC-SA-006 | PASS | `grep -q 'mx query --kind DEBT' review.md && grep -q 'already tracked @MX:DEBT' review.md` | both PRESENT; line 241 states the link "READS ... but NEVER creates, modifies, or removes an `@MX:DEBT` marker" |
| AC-SA-007 | PASS | `grep -q 'anti-patterns.md' review.md && grep -q 'Enforce Simplicity' review.md`; not-duplicated check | both PRESENT; `--lean` section names the 2 Karpathy categories by reference only; full 6-rung ladder NOT inlined (ladder-rung-2 verbatim grep → 0) |
| AC-SA-008 | PASS | `diff -q .claude/skills/moai/workflows/review.md internal/template/templates/.claude/skills/moai/workflows/review.md`; template-neutrality grep | `PARITY: identical (full-file)`; `grep -nE 'SPEC-SIMPLICITY-AUDIT-001\|REQ-SA\|2026-06-' <template>` → no internal-token leak |

### D1-D4 minor-defect resolutions (plan-auditor iter-1)

- **D1 (PASS)** — `--repo` registered in the Supported Flags list (not just described in the mode body): `grep -c '\-\-repo' review.md` ≥ 1, present as a `- --repo:` bullet in `## Supported Flags`.
- **D2 (PASS)** — language-neutrality made mechanically checkable: the negative-token scan of the `--lean` section (AC-SA-004) finds no hardcoded single-language identifier.
- **D3 (PASS)** — the `--lean` doc covers BOTH scopes: a `### Scope (two variants — review vs audit split)` sub-section documents diff-scope (default) AND `--repo` repo-wide, mirroring the ponytail review-vs-audit split.
- **D4 (PASS)** — review.md is NOT enrolled in `workflowOptMirroredPaths`, so parity was confirmed by an ACTUAL `diff -q` (AC-SA-008 above), not a CI guard.

### Mechanical verification (observed outputs)

- Template neutrality guard: `go test ./internal/template/... -run 'TestTemplateNeutralityAudit|TestTemplateNoInternalContentLeak'` → `ok github.com/modu-ai/moai-adk/internal/template 0.737s`.
- Full template package (cascade check): `go test ./internal/template/...` → `ok ... 0.819s` (no cascading failures).
- spec-lint: `go run ./cmd/moai spec lint .moai/specs/SPEC-SIMPLICITY-AUDIT-001/spec.md` → `✓ No findings — all SPEC documents are valid`.
- Build sanity (zero-Go SPEC): `go build ./cmd/moai` → exit 0. Template tree is embedded at compile time via `//go:embed all:templates` (no separate `embedded.go` regen step needed).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-06-23
run_commit_sha: <placeholder — backfilled post-commit>
run_status: implemented
ac_pass_count: 8
ac_fail_count: 0
preserve_list_post_run_count: 0   # only review.md (live + mirror) modified; no PRESERVE-list file touched
l44_pre_commit_fetch: "git fetch origin main → 0 1 (local ahead by 1, clean); 9d7f7b266 ancestor of HEAD confirmed"
l44_post_push_fetch: <placeholder — backfilled post-push>
new_warnings_or_lints_introduced: 0   # spec-lint 0 findings; template neutrality guard ok; no Go change
cross_platform_build:
  go_build: "n/a — zero Go change (skill-prose + template mirror only); go build ./cmd/moai exit 0 (sanity)"
  goos_windows: "n/a — no syscall/platform code touched"
total_run_phase_files: 2   # review.md (live) + review.md (template mirror)
m1_to_mN_commit_strategy: "single run commit — M1-M4 are sequential edits to one section of one file (+ its mirror); plan.md §F notes no parallelism"
```

Tier: **S**. cycle_type=tdd. plan-auditor iter-1 PASS 0.89 (threshold 0.75); D1-D4 minor defects all resolved in this run. AC 8/8 PASS, 0 FAIL. Template parity identical, neutrality guard ok, spec-lint clean.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
