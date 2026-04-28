---
id: SPEC-ASKUSER-ENFORCE-001
status: completed
updated: "2026-04-25"
---

# SPEC-ASKUSER-ENFORCE-001 Progress

## Implementation Status

All 18 tasks completed. All 20 ACs verified.

## Phase Completion

| Phase | Status | Commit | Notes |
|-------|--------|--------|-------|
| A: askuser-protocol.md 신규 생성 | DONE | f137f6cc7 | 7개 섹션, 262 lines |
| B: CLAUDE.md §7/§8 보강 | DONE | cbde1d420 | ToolSearch step + cross-ref |
| C: moai-constitution.md 강화 | DONE | fed6367d4 | HARD rule + cross-ref |
| D: agent-common-protocol.md 양방향 codify | DONE | 9e150c8fa | Blocker report format |
| E: output-styles/moai/moai.md §3/§10 갱신 | DONE | 9397f30b7 | ToolSearch step + free-form anti-pattern |
| F: Template 미러 동기화 + make build | DONE | 30e63ed74 + 76f6cf252 | 5 mirrors synced, make build exit 0 |
| G: Memory dead lesson SUPERSEDED | DONE | (memory file, outside repo) | SUPERSEDED + GRADUATED markers |
| H: 통합 검증 + Lint | DONE | (see below) | All AC-AUE-012 pass, golangci-lint 0 issues |

## Acceptance Criteria Status

| AC | Status | Evidence |
|----|--------|---------|
| AC-AUE-001 | PASS | 5 files contain AskUserQuestion-only channel rule |
| AC-AUE-002 | PASS | ToolSearch(query: "select:AskUserQuestion") in 4+ files |
| AC-AUE-003 | PASS | ≤4Q/≤4O/(권장)/conversation_language in askuser-protocol.md + moai.md |
| AC-AUE-004 | PASS | description standards + bias prevention in askuser-protocol.md |
| AC-AUE-005 | PASS | Blocker report format + re-delegation in agent-common-protocol.md + askuser-protocol.md |
| AC-AUE-006 | PASS | 4 triggers + 5 exceptions in CLAUDE.md §7/§8 + askuser-protocol.md |
| AC-AUE-007 | PASS | free-form prohibited + Other option in askuser-protocol.md + output-styles |
| AC-AUE-008-1 | PASS | CLAUDE.md: AskUserQuestion ≥24 hits, ToolSearch ×2, Socratic ×6 |
| AC-AUE-008-2 | PASS | moai-constitution.md: ToolSearch ×1, askuser-protocol ×1 |
| AC-AUE-008-3 | PASS | agent-common-protocol.md: orchestrator+ToolSearch, missing inputs ×2, askuser-protocol ×1 |
| AC-AUE-008-4 | PASS | output-styles/moai/moai.md: AskUserQuestion ×6, ToolSearch×1, free-form prohibited ×1 |
| AC-AUE-008-5 | PASS | askuser-protocol.md: 7 sections, 262+ lines |
| AC-AUE-009-1 | PASS | 5 diff -q commands all exit 0 |
| AC-AUE-009-2 | PASS | make build exit 0, go build exit 0 |
| AC-AUE-009-3 | PASS | go:embed all:templates in embed.go (no separate embedded.go file in this project) |
| AC-AUE-010 | PASS | [SUPERSEDED by SPEC-ASKUSER-ENFORCE-001] in memory file |
| AC-AUE-011 | PASS | [GRADUATED to SPEC-ASKUSER-ENFORCE-001] in MEMORY.md |
| AC-AUE-012 | PASS | All 14 integration checks exit 0 |

## TRUST 5 Gate Summary

| Pillar | Status | Evidence |
|--------|--------|---------|
| Tested | PASS | All AC grep checks pass, go build, go vet, golangci-lint 0 issues |
| Readable | PASS | Single canonical source (askuser-protocol.md), 4 cross-refs, consistent vocab |
| Unified | PASS | 5 mirrors byte-for-byte identical, go:embed templates synced |
| Secured | PASS | No credentials, no OWASP impact |
| Trackable | PASS | All commits reference SPEC-ASKUSER-ENFORCE-001 + task IDs |

## Dogfood Self-Validation

The SPEC's own documents (.moai/specs/SPEC-ASKUSER-ENFORCE-001/) comply with their own REQs:
- REQ-AUE-008 file-marker existence: All 5 core files contain required markers
- AC-AUE-012 bash sequence: All checks exit 0
- golangci-lint: 0 issues
- go vet: 0 issues
