# SPEC-V3R5-STATUSLINE-V2145-001 Progress Tracker

## Status Summary

| Phase | Status | Owner | Notes |
|-------|--------|-------|-------|
| Plan | completed | manager-spec | 4 artifact files, 1040 LOC, 16 REQs + 20 ACs (1:1 traceability) — committed `d61a37a45` on main (admin override: no plan-PR per user directive 2026-05-20) |
| Run (M1 Hotfix) | completed | orchestrator (direct) | DEBUG default OFF + dead echo cleanup — committed `1d21d78c2` on main |
| Run (M2 Feature) | completed | manager-develop (cycle_type=tdd) | PR segment + RepoInfo + 52 sub-tests, 87.8% coverage — committed `52a0fe302` on main |
| Run (M3 Docs) | completed | manager-docs (zh) + orchestrator (ko/en/ja interrupted-batch + _meta.yaml updates) | 4-locale parity achieved (334 lines each, 10-12 KB) + _meta.yaml × 4 + orphan dirs (`{}/`, `internal/hook/.moai/`) cleanup |
| Sync | completed | manager-docs (via /moai sync --merge --pr) | sync branch + squash PR; STATUSLINE-only scope (LATE-BRANCH-001 work shipped via separate sync PR) |

## Milestone Progress

### M1 — Disappearing Hotfix (✅ COMPLETED 2026-05-20)

- [x] AC-SLV-001 — DEBUG_STATUSLINE default 0 in rendered `.moai/status_line.sh` (line 18)
- [x] AC-SLV-002 — Debug fork guarded by explicit `export DEBUG_STATUSLINE=1` opt-in
- [x] AC-SLV-003 — Dead `echo ""` lines (8 occurrences) removed from `internal/template/templates/.moai/status_line.sh.tmpl`
- [ ] AC-SLV-004 — `statusLine.padding` documented in all 4 docs-site locales **→ deferred to M3 batch**
- [ ] AC-SLV-005 — No user-specific absolute path in rendered wrapper **→ deferred (OQ-1 D1, not addressed in M1 hotfix scope; tracked for future cleanup)**

### M2 — PR Segment Addition (✅ COMPLETED 2026-05-20)

- [x] AC-SLV-010 — `StdinData.PR *PRInfo` field exists with json tag `pr,omitempty` + `PRInfo {Number, URL, ReviewState}` struct (`internal/statusline/types.go:71,80-84`)
- [x] AC-SLV-011 — `WorkspaceInfo.Repo *RepoInfo` field + `RepoInfo {Host, Owner, Name}` struct (`internal/statusline/types.go:167-170`)
- [x] AC-SLV-012 — PR segment default off (`isPREnabled` returns false when `segments.pr` absent; default config `pr: false`)
- [x] AC-SLV-013 — PR segment render format `#<number> ⌥<state>` (renderer.go `renderPRSegment` + tests verify `#1023 ⌥approved` literals)
- [x] AC-SLV-014 — Review-state color coding for 5 cases (approved=green, pending=yellow, changes_requested=red, draft=gray, unknown=passthrough)
- [x] AC-SLV-015 — No segment emitted when PR absent (nil PR or Number==0)
- [x] AC-SLV-016 — `SegmentPR = "pr"` constant in types.go + cross-file references in builder.go (3 sites) and renderer.go (3 sites)
- [x] AC-SLV-017 — Coverage ≥85% on changed files (achieved 87.8% on `internal/statusline/`; `renderPRSegment`/`isPREnabled`/`prReviewStateColor` all 100%)

### M3 — docs-site 4-locale sync (⏸️ DEFERRED)

- [ ] AC-SLV-020 — Korean canonical section exists **→ deferred**
- [ ] AC-SLV-021 — 4-locale parity (ko/en/ja/zh) **→ deferred**
- [ ] AC-SLV-022 — docs CI passes + URL blacklist clean **→ deferred**

**Deferral rationale**: Per user directive 2026-05-20, M3 docs-site work is postponed to a batched documentation sync iteration (separate from this SPEC's run-phase). M2's code-level implementation is self-documenting via the type/const names + test fixtures. The PR segment is opt-in (default OFF), so the absence of user-facing docs does not silently change behavior for existing users — they must explicitly toggle `segments.pr: true` to see the new behavior.

### Cross-Milestone (✅ PASS where applicable)

- [x] AC-SLV-100 — Zero-regression on existing 14 segments (full statusline test suite PASS in 5.683s)
- [x] AC-SLV-101 — golangci-lint baseline does not regress (`0 issues` on `internal/statusline/...`)
- [x] AC-SLV-102 — `go build ./...` succeeds (cross-package build verified)
- [x] AC-SLV-103 — Conventional Commits format (all 3 commits follow `<type>(<scope>): <description>` pattern with MoAI signature)

## Commit Log

| Date | SHA | Branch | Author | Summary |
|------|-----|--------|--------|---------|
| 2026-05-20 | `d61a37a45` | main | manager-spec | feat(spec): plan-phase (4 SPEC files + research file, 1040 LOC) |
| 2026-05-20 | `1d21d78c2` | main | orchestrator | fix(statusline): M1 hotfix — DEBUG default OFF + dead echo cleanup |
| 2026-05-20 | `52a0fe302` | main | manager-develop | feat(statusline): M2 — PR segment + RepoInfo (52 sub-tests, 87.8% coverage) |

## PR Links

| Phase | PR # | Status | Merged commit |
|-------|------|--------|---------------|
| Plan | — | bypassed (admin override per user directive) | `d61a37a45` direct main |
| Run M1 | — | bypassed | `1d21d78c2` direct main |
| Run M2 | — | bypassed | `52a0fe302` direct main |
| M3 docs | — | deferred to batched docs work | n/a |
| Sync | — | n/a (main-direct workflow) | n/a |

## Open Issues / Blockers

- **AC-SLV-004 + AC-SLV-005 + AC-SLV-020/021/022** deferred to batched docs work. Track via separate task or absorb into next docs-batch SPEC.
- **OQ-1 D1** (rendered wrapper hardcoded path → `$HOME` runtime expansion) NOT applied in M1 hotfix. Symptom (statusline disappearing) is independently resolved by AC-SLV-001/002/003. Path normalization is portability concern only; safe to defer.

## Re-planning Triggers

None during run-phase. All iterations succeeded on first attempt:

| Iteration | AC PASS count | NEW errors introduced | Stagnation flag |
|-----------|---------------|----------------------|-----------------|
| M1 (single pass) | 3/3 in-scope | 0 | no |
| M2 (TDD single pass) | 8/8 in-scope + 4 cross | 0 | no |

## Lessons Capture

To be reviewed for memory promotion via auto-memory `lessons.md`:

- **Lesson candidate (workflow)**: User can opt to skip plan-PR + run-PR + sync-PR pipeline for low-risk maintenance SPECs by directing "메인 브랜치 직접 진행". Admin override pattern previously used for W1/W2 chicken-and-egg lint baselines applies here even without that specific reason — generalized to "user-discretion bypass for lightweight SPECs".
- **Lesson candidate (statusline debugging)**: DEBUG instrumentation defaults that survive into production are the #1 root cause of statusline timeout-induced disappearance. Pre-merge checklist for shell-wrapper SPECs: verify every `DEBUG_*:-N` default is `0`.
- **Lesson candidate (cross-version JSON schema)**: When upstream (CC) adds new JSON fields to a contract (statusline stdin), `grep -rn '<field-name>' internal/<package>/` over the parsing package is the canonical detection method. Verified zero matches → confident gap analysis.

## Handoff Notes

**For next docs batch**: M3 includes 4 new locale files (`docs-site/content/{ko,en,ja,zh}/advanced/statusline.md`) and 4 `_meta.yaml` updates. Reference SPEC plan.md §2.3 and acceptance.md AC-SLV-020/021/022. The M2-implemented `PRInfo` / `RepoInfo` / `SegmentPR` are self-documenting via test fixtures; docs should describe user-facing concepts (when PR badge appears, color meanings, opt-in toggle), not internal Go API.

**For future statusline SPECs**: Path normalization (rendered wrapper hardcoded `/Users/goos/go/bin/moai` → `$HOME/go/bin/moai`) tracked as low-priority cleanup. Activation criterion: when next fork of moai-adk-go encounters portability issue OR `moai update` regenerates the rendered file from a template that uses `$HOME`.

**For SPEC lifecycle status**: This SPEC remains `in-progress` (status field in spec.md) until M3 docs sync completes. On M3 completion, transition to `completed` with version bump 0.1.0 → 0.2.0 → 1.0.0 (or per project convention).
