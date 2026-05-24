---
id: SPEC-V3R6-SPEC-LINT-CLEANUP-001
title: "spec-lint MissingExclusions baseline cleanup — Acceptance Criteria"
version: "0.1.0"
status: implemented
created: 2026-05-25
updated: 2026-05-25
author: GOOS행님
priority: P2
phase: "v3.0.0"
module: ".moai/specs"
lifecycle: spec-anchored
tags: "spec-lint, missing-exclusions, baseline-cleanup, h3-pattern, retroactive, tier-s, acceptance"
sync_commit_sha: "0d777471c21f36f827752608ea6b7bcceea09fd8"
---

# SPEC-V3R6-SPEC-LINT-CLEANUP-001 — Acceptance Criteria

## §D. AC Matrix (7 ACs — REQ-SLC-001..007)

| # | AC ID | REQ Covered | Severity | Phase | Verification Command (binary PASS/FAIL) | Expected Output |
|---|-------|------------|---------|-------|---------------------------------------|-----------------|
| 1 | AC-SLC-001 | REQ-SLC-001 | Must-Pass | plan | `grep -E '^\| [0-9]+ \| SPEC-V3R6-' .moai/specs/SPEC-V3R6-SPEC-LINT-CLEANUP-001/spec.md \| wc -l` | `8` (8개 sibling SPEC row enumerated in spec.md §2.2) |
| 2 | AC-SLC-002 | REQ-SLC-002 | Must-Pass | plan | `grep -c 'Out of Scope — ' .moai/specs/SPEC-V3R6-SPEC-LINT-CLEANUP-001/spec.md` | `≥3` (§3.1 minimum form codeblock + §3 self-compliance H3 + §5 examples; canonical pattern codified) |
| 3 | AC-SLC-003 | REQ-SLC-003 | Must-Pass | plan | `grep -E '분류 [AB]' .moai/specs/SPEC-V3R6-SPEC-LINT-CLEANUP-001/spec.md \| wc -l` | `≥5` (분류 A/분류 B taxonomy applied across §2.2 table + §3.2 + §3.3 + plan.md mapping) |
| 4 | AC-SLC-004 | REQ-SLC-004 | Must-Pass | run | `git diff --name-only main..<run-phase-head> -- '.moai/specs/' \| grep -v 'SPEC-V3R6-SPEC-LINT-CLEANUP-001'` | only 8 paths matching `.moai/specs/SPEC-V3R6-{CI-BASELINE-DRIFT-001\|HOOK-CWD-LEAK-AUDIT-001\|LEGACY-CLEANUP-001\|LEGACY-CLEANUP-002\|LEGACY-CLEANUP-003\|PROMPT-CACHE-001\|SESSION-HANDOFF-AUTO-001\|TEMPLATE-NEUTRALITY-AUDIT-001}/spec.md` |
| 5 | AC-SLC-005 | REQ-SLC-005 | Must-Pass | run | for-each `<sibling>`: `git diff main..<run-phase-head> -- .moai/specs/<sibling>/spec.md \| grep -E '^-' \| grep -v '^---' \| grep -v '^-$' \| grep -vE '^-[[:space:]]*(###\|---)' ` | 0 lines (no deletion of non-empty list item bodies — only `###` heading insertions and `-` item insertions; semantic drift = 0) |
| 6 | AC-SLC-006 | REQ-SLC-006 | Must-Pass | run | `~/go/bin/moai spec lint 2>&1 \| grep -c MissingExclusions` | `0` (baseline failure cleared) |
| 7 | AC-SLC-007 | REQ-SLC-007 | Should-Pass (Optional) | plan | `grep -c '### §5\.[0-9]\+ Out of Scope' .moai/specs/SPEC-V3R6-SPEC-LINT-CLEANUP-001/spec.md` | `≥1` (본 spec.md self-compliance — §5에 H3 sub-heading 1개 이상 존재) |

## §D.1 Severity Legend

- **Must-Pass**: AC가 FAIL이면 phase 전체가 FAIL. plan-auditor PASS verdict 차단 (plan-phase) 또는 run-phase 종료 차단 (run-phase).
- **Should-Pass (Optional)**: AC가 FAIL이어도 phase는 진행 가능하나 self-compliance demerit로 plan-auditor verdict에서 부분 감점 (-0.02 ~ -0.05).

## §D.2 Phase Mapping

| Phase | Must-Pass ACs | Should-Pass ACs | 평가 시점 |
|-------|--------------|----------------|---------|
| plan | AC-SLC-001, AC-SLC-002, AC-SLC-003 | AC-SLC-007 | plan-auditor 시점 + spec.md/plan.md write 완료 후 |
| run | AC-SLC-004, AC-SLC-005, AC-SLC-006 | — | manager-develop run-phase 종료 시점 |

## §D.3 Traceability Matrix (REQ ↔ AC ↔ verification)

| REQ ID | EARS Pattern | AC ID | Verification Pattern |
|--------|-------------|-------|---------------------|
| REQ-SLC-001 | Ubiquitous (SHALL enumerate 8 SPECs) | AC-SLC-001 | grep + wc count = 8 |
| REQ-SLC-002 | Ubiquitous (SHALL codify minimum form) | AC-SLC-002 | grep H3 pattern occurrences ≥3 |
| REQ-SLC-003 | Ubiquitous (SHALL classify A/B) | AC-SLC-003 | grep classification labels |
| REQ-SLC-004 | State-Driven (run-phase WHILE …) | AC-SLC-004 | git diff --name-only filter |
| REQ-SLC-005 | Unwanted (SHALL NOT change semantic) | AC-SLC-005 | git diff non-empty deletion count = 0 |
| REQ-SLC-006 | Event-Driven (WHEN run-phase 완료) | AC-SLC-006 | moai spec lint count = 0 |
| REQ-SLC-007 | Optional (WHERE possible self-compliance) | AC-SLC-007 | self-grep on own spec.md |

100% REQ ↔ AC bidirectional coverage. plan-auditor S5 traceability dimension PASS 기대.

## §D.4 Edge Cases

| 시나리오 | AC 영향 | 대응 |
|--------|--------|------|
| run-phase 중 parallel session이 신규 SPEC을 작성하면서 `MissingExclusions` failure 추가 | AC-SLC-006 FAIL 가능 (count > 0) | (i) parallel session의 신규 failure가 본 SPEC scope 8개와 분리 가능한 경우: PASS-WITH-NOTE 처리, (ii) 분리 불가능한 경우: re-delegate run-phase로 신규 SPEC 추가 처리. 사용자 결정 필요 |
| 8개 sibling SPEC 중 일부가 archive 대상으로 결정되어 `.moai/specs/_archive/`로 이동된 경우 | AC-SLC-004 path 변경 | archive된 SPEC의 spec.md는 더 이상 `moai spec lint` 대상 외. AC-SLC-006는 자동 PASS. AC-SLC-004 path 매칭은 운영 시점 `.moai/specs/` 잔존 SPEC 기준으로 평가 |
| lint tool 버전 변경으로 rule 알고리즘이 변경됨 | AC-SLC-006 verification 동작 변경 가능 | `~/go/bin/moai spec lint --version` 출력 기록 + lint rule code 변경 시 본 SPEC scope 재평가 |
| H3 텍스트가 `### Out of Scope` (대소문자 보존) vs `### out of scope` (lowercase) | OutOfScopeRule는 case-insensitive 적용 (lint.go:683 `strings.ToLower`) | run-phase 양쪽 모두 허용. canonical pattern은 `Out of Scope` Title Case 권장 |

## §D.5 Definition of Done

본 SPEC plan-phase는 다음을 모두 만족할 때 DONE으로 간주한다:

- [ ] 4개 artifact (`spec.md`, `plan.md`, `acceptance.md`, `progress.md`) 작성 완료
- [ ] AC-SLC-001 + AC-SLC-002 + AC-SLC-003 (plan-phase Must-Pass) 모두 PASS
- [ ] AC-SLC-007 (plan-phase Should-Pass) PASS (자기 준수)
- [ ] plan-auditor iter-1 verdict ≥ 0.75 (Tier S 최소) 또는 PASS-WITH-DEBT 명문 정리
- [ ] 12-field frontmatter schema 4 파일 모두 준수
- [ ] SPEC ID regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` 매칭 (verified via Pre-Write Self-Check Protocol)
- [ ] `MissingExclusions` self-compliance — 본 spec.md §5에 H3 sub-heading + `-` list item 존재

본 SPEC run-phase는 별도 사이클에서 다음을 모두 만족할 때 DONE으로 간주한다:

- [ ] AC-SLC-004 + AC-SLC-005 + AC-SLC-006 (run-phase Must-Pass) 모두 PASS
- [ ] `moai spec lint` baseline `MissingExclusions` count = 0
- [ ] 8개 sibling spec.md 외 file 변경 0건
- [ ] commit message에 SPEC-V3R6-SPEC-LINT-CLEANUP-001 attribution + 변경 SPEC ID 명시
