# progress.md — SPEC-JEV-SKILL-SUGGESTION-001

Card: t1340 | Class C | Worktree: `.moai/worktrees/t1340` | Branch: `WT-jev-skill-suggest` | Base: develop @ `7a713a9a8`

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-09-30
tier: M
artifacts: spec.md, plan.md, acceptance.md, research.md, progress.md
spec_id_check: PASS (SPEC-JEV-SKILL-SUGGESTION-001 — regex-match-count=1)
frontmatter_check: 12 canonical fields + depends_on + related_specs + tier
red_now_observed: true (4 deliverable-absence greps all 0 + REQ-JSK prefix unique, this tree)
needs_clarification_count: 0
key_finding: dispatch premise corrected — consumer B withdrawn by SPEC-JEV-GUARD-001; design is a no-call-path guidance skill
plan_audit_iter1: CONDITIONAL 0.875 (Tier M 0.80) — D1/D2 blocking + D3/D4/D5 optional, all five applied in the 0.1.1 annotation commit; delta re-audit scope = the two amended AC cells
next: delta re-audit, then Implementation Kickoff Approval gate, then /moai run SPEC-JEV-SKILL-SUGGESTION-001
```

## §E.2 Run-phase Evidence

Convention (VCI §3 + plan-audit D4): every evidence entry carries (a) the
command, (b) its verbatim output, (c) the tree HEAD SHA measured against, and
(d) that command's exit code — `grep -c` zero-hits exit 1, so code and count
travel as a pair.

### M1 — guard test first (RED observed before any deliverable exists)

Pre-flight baseline (tree `235fcfd12`): `go build ./...` → exit 0
(observed `BUILD_EXIT_0`); skill file absent (`ls .claude/skills/ | grep
moai-jev-skill-suggestion` → 0 hits, exit 1); catalog entry absent
(`grep -c "moai-jev-skill-suggestion" internal/template/catalog.yaml` → `0`,
exit 1).

E8 (RED evidence) — tree `235fcfd12` (pre-M1-commit), exit 1:

```
$ go test ./internal/cli/ -run '^TestJevSkillSuggestionSkill(CarriesNoCallPath|CopiesStayIdentical)$' -count=1
--- FAIL: TestJevSkillSuggestionSkillCarriesNoCallPath (0.00s)
    jev_skill_suggestion_skill_test.go:61: read ../../.claude/skills/moai-jev-skill-suggestion/SKILL.md: open ../../.claude/skills/moai-jev-skill-suggestion/SKILL.md: no such file or directory
--- FAIL: TestJevSkillSuggestionSkillCopiesStayIdentical (0.00s)
    jev_skill_suggestion_skill_test.go:81: read ../../.claude/skills/moai-jev-skill-suggestion/SKILL.md: open ../../.claude/skills/moai-jev-skill-suggestion/SKILL.md: no such file or directory
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.051s
FAIL
```

RED is an **existence RED**: both tests fail on the skill file's absence
(read-fail) — the right stated reason. The positive controls read and matched
before the failure line, proving the scan fires on files that legitimately
carry the tokens; the sweep counts exactly 2 tests (the anchored selector), so
the green after M2 will not be an empty sweep. Full output also persisted at
`/tmp/t1340-red.log` in this run.

### M2 — skill body + mirror (GREEN, sentinel sweep, neutrality)

Deliverables: `internal/template/templates/.claude/skills/moai-jev-skill-suggestion/SKILL.md`
(template source, authored first) + `.claude/skills/moai-jev-skill-suggestion/SKILL.md`
(local copy, `cp`-mirrored). 115 lines, English body, `user-invocable: true`.

| # | Claim | Command | Observed output (verbatim) | Exit | Tree |
|---|-------|---------|---------------------------|------|------|
| G1 | Both copies byte-identical | `cmp .claude/skills/.../SKILL.md internal/template/templates/.claude/skills/.../SKILL.md` | (no output — identical) | 0 | this run |
| G2 | Guard test green after deliverables (RED→GREEN pair closes) | `go test ./internal/cli/ -run '^TestJevSkillSuggestionSkill(CarriesNoCallPath\|CopiesStayIdentical)$' -count=1` | `ok  	github.com/modu-ai/moai-adk/internal/cli	0.846s` (final re-run after the one sentinel prose fix; first green was `1.263s`) | 0 | this run |
| G3 | All 13 sentinels (AC-JSK-004..007) ≥ 1 on both copies | `grep -c` loop over the sentinel set × 2 copies | each sentinel → `1` (both copies identical counts; `NO SIGNAL` → `2`, `function-hook` ci → `1`, `a signal a person` → `1` after one prose fix) | 0 per grep | this run |
| G4 | Neutrality + withdrawn-name absent (AC-JSK-003) on both copies | `grep -Ec 'SPEC-...-\|REQ-...-\|t1340\|[0-9a-f]{40}\|20[0-9]{2}-...' <copy>` + `grep -c 'jev-suggest' <copy>` | `0` / `0` per copy (grep -Ec zero-hit exits 1 — code and count recorded as a pair) | 1 per grep | this run |

One sentinel prose fix during M2: the first body draft carried "a labelled
model **signal a person** reads", which misses the AC-JSK-004 sentinel
`a signal a person` for want of its leading article; rephrased to "A
suggestion is a signal a person reads: a labelled model answer…" in the
template source first, then re-mirrored (`cp`) and re-verified — identity,
sentinels, and guard test all re-observed after the fix.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — manager-develop 소관>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs 소관. sync_commit_sha: (pending-backfill at sync commit)>_
