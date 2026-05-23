---
id: SPEC-V3R6-CHANGELOG-CLEANUP-001
title: "CHANGELOG.md [Unreleased] hallucination + duplicate entry cleanup"
version: "0.2.0"
status: implemented
created: 2026-05-23
updated: 2026-05-23
author: MoAI Maintainer
priority: P2
phase: "v3.0.0-rc2"
module: "CHANGELOG.md"
lifecycle: spec-anchored
tags: "changelog, sync, hallucination, batch-sync, manager-docs"
tier: S
issue_number: null
depends_on: []
related_specs: [SPEC-V3R6-SESSION-HANDOFF-AUTO-001, SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001, SPEC-V3R6-HOOK-ASYNC-EXPAND-001, SPEC-V3R6-HOOK-CWD-LEAK-AUDIT-001, SPEC-V3R6-CI-BASELINE-DRIFT-001]
---

# SPEC-V3R6-CHANGELOG-CLEANUP-001: CHANGELOG.md `[Unreleased]` 환각·중복 엔트리 정리

## Section A — Pre-existing State Survey (codebase audit prior to scoping)

본 섹션은 BATCH-SYNC 커밋 `9e96e21b4` (2026-05-23, parallel session) 가 `CHANGELOG.md` `[Unreleased]` 섹션에 환각 + 중복 엔트리를 추가한 사실을 확정하고, 인접 4개 sibling SPEC 엔트리의 AC 카운트 드리프트를 점검한 결과를 5개 사실 (Five Facts) 형태로 정리한다. 모든 사실은 orchestrator pre-flight 단계에서 working-tree HEAD `97a36b5a2` (main) 에 대해 verified 되었으며, 본 SPEC 본문 외 어떤 코드도 재조사하지 않는다.

### A.1 — Fact 1: 라인 65 환각 카탈로그 (7-row drift)

CHANGELOG.md 라인 65-65+N에 위치한 SPEC-V3R6-SESSION-HANDOFF-AUTO-001 엔트리는 7개의 검증 가능한 환각을 포함한다. 본 표는 orchestrator pre-flight 가 working-tree HEAD `97a36b5a2` 에 대해 검증한 결과를 verbatim 으로 인용한다.

| Field | Hallucinated value | Verified actual value |
|-------|-------------------|----------------------|
| Path | `internal/handoff/` | `internal/hook/handoff/` (hook prefix missing) |
| Files listed | `{package,atomic_write,parser}.go` | single file `persist.go` only |
| Block 6 label | `감정` | `머지 후` |
| Volume | "10 files +556/-3" | ~1,140 LOC (persist.go 401 + persist_test.go 823 = 1,224 LOC across 2 NEW files) |
| Test count | "15 functions" | 18 functions |
| State file | `.moai/state/session-handoff.json` (JSON) | `.moai/state/session-handoff/pending.md` (Markdown + YAML frontmatter) |
| Supersede mechanism | `.supersede` file in `~/.claude/projects/{hash}/session/` | `[SUPERSEDED by ...]` marker prefix prepended to existing MEMORY.md line via `applySupersedeMarker()` |

검증 명령: `ls /Users/goos/MoAI/moai-adk-go/internal/hook/handoff/` → `persist.go` + `persist_test.go` 2개 파일만 존재; `ls /Users/goos/MoAI/moai-adk-go/internal/handoff/` → No such file or directory; `wc -l internal/hook/handoff/*.go` → 401 + 823 = 1,224 LOC; `grep -cE "^func Test" internal/hook/handoff/persist_test.go` → 18.

### A.2 — Fact 2: 라인 34 정확성 (canonical SSOT)

CHANGELOG.md 라인 34-39 는 SPEC-V3R6-SESSION-HANDOFF-AUTO-001 의 정확한 엔트리를 이미 포함한다. 이 엔트리는 직전 sync commit `1deacc6b3` (`/moai sync SPEC-V3R6-SESSION-HANDOFF-AUTO-001` manager-docs Tier S minimal) 에 의해 추가되었으며, 본 SPEC 의 cleanup 범위에서 보존 (PRESERVE) 대상이다. 결론: 라인 65 엔트리는 라인 34 의 중복 (duplicate) 이면서 동시에 환각 (hallucination) 이다. 권장 해결: 라인 65 엔트리 전체 삭제, 라인 34 는 byte-identical 보존.

### A.3 — Fact 3: 4 sibling SPEC AC 카운트 드리프트 점검 (acceptance.md as SSOT)

CHANGELOG.md 라인 57/59/61/63 의 4개 sibling SPEC 엔트리는 각각 "X/X AC PASS" 형식으로 AC 카운트를 명시한다. Orchestrator pre-flight 가 각 SPEC 의 acceptance.md (canonical SSOT) 에서 unique AC ID 카운트를 추출한 결과는 다음과 같다.

| CHANGELOG line | SPEC | Claimed AC count | acceptance.md SSOT count | Drift? |
|----------------|------|------------------|--------------------------|--------|
| 57 | HOOK-OBSERVE-OPT-IN-001 | 7/7 | 7 (AC-HOI-001~007) | NO — accurate |
| 59 | HOOK-ASYNC-EXPAND-001 | 12/12 | **8** (AC-HAE-001~008) | **YES — over-counts by 4** |
| 61 | HOOK-CWD-LEAK-AUDIT-001 | 8/8 | **7** (AC-HCWA-001~007) | **YES — over-counts by 1** |
| 63 | CI-BASELINE-DRIFT-001 | 8/8 | 8 (AC-CBD-001~008) | NO — accurate |
| 65 | SESSION-HANDOFF-AUTO-001 (duplicate) | 10/10 | 10 (AC-SHA-001~010) | (deletion target) |

검증 명령 (HOOK-ASYNC-EXPAND): `grep -oE "AC-HAE-[0-9]+" .moai/specs/SPEC-V3R6-HOOK-ASYNC-EXPAND-001/acceptance.md | sort -u | wc -l` → `8`. 검증 명령 (CI-BASELINE-DRIFT): `grep -oE "AC-CBD-[0-9]+" .moai/specs/SPEC-V3R6-CI-BASELINE-DRIFT-001/acceptance.md | sort -u | wc -l` → `8`. acceptance.md 가 SSOT 인 이유: progress.md 는 implementation phase 동안 추가되는 AC (예: AC-SHA-011 path-injection guard, M3 deferred) 를 일시적으로 포함할 수 있으나, acceptance.md 는 plan-phase 의 최종 AC 집합을 canonical 로 정의한다. CHANGELOG 엔트리는 plan-phase AC 카운트 (acceptance.md) 를 인용해야 하며, deferred AC 는 본문 안에 별도 명시 (예: "Deferred: AC-XXX-NNN") 한다.

### A.4 — Fact 4: BATCH-SYNC root cause (manager-docs prompt 결함)

라인 65 환각의 발생 원인은 BATCH-SYNC 커밋 `9e96e21b4` 의 manager-docs spawn prompt 가 implementation file 들을 `Read` 도구로 직접 검증하지 않고 SPEC plan.md 만 참조한 것에 있다. 결과: SPEC plan.md 의 미래형 표현 ("internal/handoff/ 패키지 신규 생성" 같은 plan-phase placeholder) 이 sync-phase 의 CHANGELOG entry 에 그대로 옮겨지면서, 실제 구현된 코드 (`internal/hook/handoff/persist.go` 단일 파일) 와 mismatch 가 발생했다. 6분 전 parallel session 에 의한 정확한 sync (`1deacc6b3`) 가 라인 34 에 이미 정확한 엔트리를 추가했음에도, BATCH-SYNC 는 CHANGELOG.md 의 기존 라인을 `grep` 으로 중복 탐지하지 않은 채 또 다른 엔트리를 추가했다.

본 root cause 는 후속 SPEC `SPEC-V3R6-BATCH-SYNC-WORKFLOW-001` (별도, Tier M, 본 SPEC 의 §F Deferred) 에서 BATCH-SYNC sync-workflow 구조적 재설계로 완전히 해소될 예정이다. 본 SPEC 의 범위는 (a) 라인 65 삭제, (b) 라인 59/61 sibling AC 카운트 정정, (c) manager-docs spawn prompt 에 "Read implementation files" 단계 + duplicate detection 단계 추가하는 standing-rule guard 뿐이다.

### A.5 — Fact 5: PRESERVE 대상 (untouched scope)

본 SPEC 의 run-phase 변경 범위는 CHANGELOG.md 라인 59 (HOOK-ASYNC-EXPAND AC count) + 라인 61 (HOOK-CWD-LEAK AC count) + 라인 65 (SESSION-HANDOFF-AUTO duplicate entry 전체 삭제) 만이다. 다음 파일/라인은 PRESERVE (read-only 검증만) 대상으로 절대 수정하지 않는다.

- CHANGELOG.md 라인 1-33 (v3.0.0-rc1 섹션 전체)
- CHANGELOG.md 라인 34-39 (SESSION-HANDOFF-AUTO-001 정확한 엔트리, byte-identical 보존)
- CHANGELOG.md 라인 40-56 (Wave 6 Korean→English translation + 3 update SPEC + docs-site Hextra→geekdoc 마이그레이션 + Harness Autonomy W3 entries)
- CHANGELOG.md 라인 57 (HOOK-OBSERVE-OPT-IN-001 7/7, accurate)
- CHANGELOG.md 라인 58 + 60 + 62 + 64 (entry separator blank lines)
- CHANGELOG.md 라인 63 (CI-BASELINE-DRIFT-001 8/8, accurate)
- CHANGELOG.md 라인 66-80 (`### Changed (Hook opt-in context)` 서브섹션 라인 67-71 + `audit_test.go` 라인 71-72 + `### Tooling (Standing Rules + Output Style)` 라인 73-80)
- CHANGELOG.md 라인 81+ (output-style v5.2.0 / v2.20.0-rc1 / 그 이하 Security / Added 모든 leg entries)
- 5개 sibling SPEC 디렉토리 (`.moai/specs/SPEC-V3R6-{HOOK-OBSERVE-OPT-IN,HOOK-ASYNC-EXPAND,HOOK-CWD-LEAK-AUDIT,CI-BASELINE-DRIFT,SESSION-HANDOFF-AUTO}-001/`) 의 모든 파일 (spec.md, plan.md, acceptance.md, progress.md) — orchestrator pre-flight 가 acceptance.md 만 read-only 로 SSOT 검증, run-phase 는 어떤 sibling SPEC 파일도 수정 금지

## Section B — EARS Requirements

### REQ-CHL-001 — Ubiquitous: Line 65 hallucination entry deletion

The system shall remove the entire SESSION-HANDOFF-AUTO-001 entry at CHANGELOG.md line 65 (including its bullet sub-lines that contain the 7 hallucinated details enumerated in §A.1). The deletion shall be a single contiguous range removal without affecting any other line.

### REQ-CHL-002 — Event-Driven: Sibling AC count reconciliation

WHEN the run-phase executes the sibling AC reconciliation task, the system shall update CHANGELOG.md line 59 (HOOK-ASYNC-EXPAND-001 "AC 12/12 PASS" → "AC 8/8 PASS") and line 61 (HOOK-CWD-LEAK-AUDIT-001 "AC 8/8 PASS" → "AC 7/7 PASS"), AND shall NOT modify line 57 (HOOK-OBSERVE-OPT-IN-001 7/7) or line 63 (CI-BASELINE-DRIFT-001 8/8) — both verified accurate against their acceptance.md SSOT.

### REQ-CHL-003 — Event-Driven: manager-docs sync-workflow guard insertion

WHEN the run-phase reaches the standing-rule guard task, the system shall insert a new "Read implementation files" enforcement step into `.claude/rules/moai/development/manager-develop-prompt-template.md` Section B (existing B1-B11) as a new bullet B12, instructing every manager-docs CHANGELOG-emission delegation prompt to (a) `Read` the actual implementation files referenced in the SPEC before drafting CHANGELOG entries, AND (b) `grep -c "<SPEC-ID>" CHANGELOG.md` before appending to detect prior sync entries from parallel sessions.

### REQ-CHL-004 — Unwanted: Line 34 entry preservation guard

The system shall NOT modify CHANGELOG.md line 34-39 (existing accurate SESSION-HANDOFF-AUTO-001 entry from sync commit `1deacc6b3`). Any change to this byte range — including whitespace, line-ending, or character substitution — is a regression.

### REQ-CHL-005 — State-Driven: Out-of-scope line preservation

WHILE the run-phase executes any of M1/M2/M3 milestones, the system shall preserve byte-identical content for CHANGELOG.md lines 1-33 (v3.0.0-rc1 section), line 57 (HOOK-OBSERVE-OPT-IN-001 entry), line 63 (CI-BASELINE-DRIFT-001 entry), and lines 81+ (Tooling / Standing Rules section). Diff scope shall be strictly limited to lines 59, 61, and the line-65-anchored deletion range.

### REQ-CHL-006 — Optional: Sibling SPEC file read-only access

WHERE the run-phase verifies sibling AC counts, the system shall access `.moai/specs/SPEC-V3R6-HOOK-ASYNC-EXPAND-001/acceptance.md` and `.moai/specs/SPEC-V3R6-CI-BASELINE-DRIFT-001/acceptance.md` in read-only mode (Read tool only, no Edit/Write). No sibling SPEC file shall be modified by any milestone.

## Section C — Edge Cases

### EC-CHL-001 — 동시 편집 race (concurrent CHANGELOG.md edits)

본 SPEC 의 run-phase 가 진행 중인 동안 parallel session 이 또 다른 CHANGELOG.md 편집을 시도할 가능성이 존재한다. 완화: M1 시작 전 `git fetch origin main && git log --oneline -5` 로 race 탐지, 만약 새 커밋이 발견되면 manager-develop 가 blocker report 반환 (subagent boundary: AskUserQuestion 금지), orchestrator 가 사용자에게 rebase/abort 옵션 제시.

### EC-CHL-002 — 부분 AC 카운트 정정 실패 (partial reconciliation)

M2 의 line 59 와 line 61 정정 중 하나만 성공하고 다른 하나가 실패하는 시나리오. 완화: M2 단일 commit 으로 두 라인을 atomic 하게 변경 (sed 또는 Edit 도구 2-call 시퀀스), AC-CHL-002 의 binary verify 가 두 라인을 동시 검증 (둘 다 PASS 해야 commit). 부분 실패 시 manager-develop 가 두 Edit 모두 revert 후 blocker report.

### EC-CHL-003 — 라인 34 엔트리 정확성 자체 의심 (line 34 drift)

만약 line 34 엔트리도 추후 환각으로 판명되면 본 SPEC 의 PRESERVE 가정이 무효화된다. 완화: orchestrator pre-flight 가 line 34 엔트리의 7개 사실 필드 (Path / Files / Block 6 / Volume / Test count / State file / Supersede mechanism) 를 line 65 와 동일한 검증 명령으로 cross-check 완료 (§A.2 결론: line 34 는 정확). 만약 본 SPEC run-phase 중 line 34 의 추가 부정확성이 발견되면 별도 SPEC `SPEC-V3R6-CHANGELOG-LINE34-DRIFT-001` 분기 신설 (현 SPEC scope 밖).

### EC-CHL-004 — BATCH-SYNC 커밋의 추가 환각 누적 (cascading drift)

BATCH-SYNC 커밋 `9e96e21b4` 가 SESSION-HANDOFF-AUTO 외에 다른 SPEC 엔트리에도 환각을 추가했을 가능성. 완화: §A.3 의 4 sibling 점검이 정확히 이 cascading drift 를 탐지하기 위한 cross-check 였으며, 결과는 line 59 + line 61 의 AC 카운트 드리프트만 발견. 추가 드리프트 발견 시 본 SPEC scope 확장 대신 별도 follow-up SPEC 신설 (§F Deferred 참조).

## Section D — Quality Gate (binary verification commands)

본 SPEC 의 run-phase 완료 후 orchestrator 가 단일 turn 내 parallel batch 로 실행하는 binary verify 명령. 모든 명령은 exit 0 = PASS, exit !=0 = FAIL.

```bash
# Q1. 라인 65 환각 카탈로그의 7개 필드 모두 CHANGELOG.md 에서 제거됨
grep -c 'internal/handoff/{package,atomic_write,parser}' CHANGELOG.md  # → 0
grep -c '`감정`' CHANGELOG.md                                            # → 0
grep -c '10 files +556/-3' CHANGELOG.md                                  # → 0
grep -c '15 functions' CHANGELOG.md                                      # → 0

# Q2. 라인 59 + 라인 61 AC 카운트 정정 (sibling SSOT 일치)
test "$(grep -oE 'AC 8/8 PASS' CHANGELOG.md | wc -l | tr -d ' ')" -ge 2
grep -c 'AC 12/12 PASS' CHANGELOG.md                                     # → 0
grep -c 'AC 8/8 PASS.*HOOK-ASYNC-EXPAND' CHANGELOG.md                    # → 1 (line 59)

# Q3. manager-develop-prompt-template.md Section B 에 B12 가드 삽입됨
grep -c 'Read implementation files' .claude/rules/moai/development/manager-develop-prompt-template.md  # → ≥1
grep -c 'B12' .claude/rules/moai/development/manager-develop-prompt-template.md                         # → ≥1

# Q4. 라인 34 엔트리 byte-identical 보존
test "$(sed -n '34,39p' CHANGELOG.md | sha256sum | awk '{print $1}')" = "$EXPECTED_SHA256_LINE_34_39"

# Q5. 5 sibling SPEC 디렉토리 read-only 보장
git diff --name-only HEAD~3..HEAD .moai/specs/SPEC-V3R6-{HOOK-OBSERVE-OPT-IN,HOOK-ASYNC-EXPAND,HOOK-CWD-LEAK-AUDIT,CI-BASELINE-DRIFT,SESSION-HANDOFF-AUTO}-001/ | wc -l  # → 0
```

(Q4 의 `$EXPECTED_SHA256_LINE_34_39` 는 run-phase M0 pre-flight 가 baseline 으로 capture 후 progress.md 에 기록. AC-CHL-004 의 binary verify 가 동일 sha256 을 재검증.)

## Section E — Out of Scope

### E.0 Out of Scope Items (canonical bullet list for spec-lint)

- Codemaps 재생성 (`/moai codemaps` 호출은 본 SPEC sync-phase 에서도 생략)
- CHANGELOG.md 라인 75-79 (Tooling / Standing Rules + Output Style 섹션) 의 standing-rule entries (orthogonal scope)
- CHANGELOG.md 라인 81+ 의 모든 standing-rule + 도구/output-style 변경 기록
- Sibling SPEC body (spec.md / plan.md / acceptance.md / progress.md) 자체의 AC 카운트 정정 (acceptance.md SSOT 기준 이미 정확)
- BATCH-SYNC sync-workflow 구조적 재설계 (별도 후속 SPEC `SPEC-V3R6-BATCH-SYNC-WORKFLOW-001` 에서 다룸)
- CHANGELOG.md 라인 34-39 의 추가 audit (본 SPEC merge 후 별도 chore 로 수행, 드리프트 발견 시 별도 SPEC 신설)

### E.1 — Codemaps 재생성

본 SPEC 의 cleanup 작업은 CHANGELOG.md + 1개 rule template 만 수정하므로 codemap regeneration 대상이 아니다. `/moai codemaps` 호출은 본 SPEC sync-phase 에서도 생략한다.

### E.2 — Out of Scope (Standing-Rule Entries)

CHANGELOG.md 라인 75-79 (Tooling / Standing Rules + Output Style 섹션, commit `ba85955db` manager-develop-prompt-template B9/B10/B11 promotion) 와 라인 81+ 의 모든 standing-rule entry 는 본 SPEC 의 scope 와 orthogonal 하다. 이들은 (a) 환각이 아니며, (b) SPEC entry 가 아닌 도구/output-style 변경 기록이므로, AC 카운트 reconciliation 대상이 아니다. 본 SPEC 의 diff 는 이 영역을 절대 침범하지 않는다.

### E.3 — Sibling SPEC body 정정

라인 59/61 의 AC 카운트 드리프트가 CHANGELOG.md 에서 발견되었더라도, 해당 sibling SPEC 의 본문 (spec.md, plan.md, acceptance.md, progress.md) 은 acceptance.md SSOT 기준으로 이미 정확하다 (orchestrator pre-flight verified). 따라서 sibling SPEC 파일 자체의 정정은 불필요하며, 본 SPEC 의 변경 범위에서 명시적으로 제외한다.

### E.4 — BATCH-SYNC 워크플로우 구조적 재설계

§A.4 의 root cause 분석에서 식별된 BATCH-SYNC sync-workflow 결함 (단일 turn 내 다수 SPEC 동시 sync 시 implementation file 검증 누락 + duplicate detection 누락) 의 구조적 해소는 별도 SPEC `SPEC-V3R6-BATCH-SYNC-WORKFLOW-001` (Tier M, 후속) 에서 다룬다. 본 SPEC 은 단지 manager-docs spawn prompt 의 standing-rule guard (B12) 만 추가하여 1차 방어선을 구축한다.

## Section F — Deferred

### F.1 — BATCH-SYNC sync-workflow 구조적 SPEC

본 SPEC 의 §A.4 root cause 분석은 BATCH-SYNC 가 단일 turn 으로 다수 SPEC 을 동시 sync 할 때 implementation file 검증 + duplicate detection 이 누락된다는 결함을 식별했다. 본 SPEC 은 manager-docs spawn prompt 의 B12 standing-rule guard 만 추가하여 1차 방어선을 구축하며, 구조적 해소 (예: sync-workflow 의 multi-phase 분해, per-SPEC Read-verify gate, CHANGELOG duplicate hash registry 등) 는 후속 SPEC `SPEC-V3R6-BATCH-SYNC-WORKFLOW-001` (Tier M) 에서 다룬다. 후속 SPEC 의 trigger condition: 본 SPEC merge 후 BATCH-SYNC 가 2회 이상 다시 시도되거나, B12 guard 가 우회되는 사례가 1회라도 관찰되면 즉시 plan-phase 진입.

### F.2 — CHANGELOG line 34 entry follow-up audit

본 SPEC pre-flight 는 line 34 엔트리의 7개 사실 필드를 검증하여 정확함을 확인했으나, line 34 의 잔여 부정확성 (예: Block 6 label 의 미세한 격차, supersede mechanism 의 호출 chain 미세 오차) 가 추후 발견될 가능성에 대비하여, 본 SPEC merge 후 30일 이내 line 34 의 추가 audit 을 별도 chore (SPEC 미신설) 로 수행 권장. 추가 드리프트 발견 시 별도 SPEC `SPEC-V3R6-CHANGELOG-LINE34-DRIFT-001` 신설.
