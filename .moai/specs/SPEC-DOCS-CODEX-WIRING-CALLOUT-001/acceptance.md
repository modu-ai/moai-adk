---
id: SPEC-DOCS-CODEX-WIRING-CALLOUT-001
title: "Acceptance — docs-site moai doctor 페이지 Codex Wiring 콜아웃 4로케일 반영"
version: "0.1.0"
created: 2026-09-08
updated: 2026-09-08
author: manager-spec
priority: P2
phase: "v3.1.4"
module: "docs-site/content"
tags: "docs, docs-site, i18n, codex, doctor"
tier: M
---

# Acceptance: SPEC-DOCS-CODEX-WIRING-CALLOUT-001

## §A 검증 규약 (모든 AC에 공통)

- 부재 판정 명령은 `/usr/bin/grep` — 이 리포 셸 `grep` 은 ugrep 래퍼로 부재 판정을 무효화할 수 있다 (REQ-DWC-014).
- RED-now 셀의 귀속 baseline: 워크트리 t535, branch `WT-docs-codex-callout`, tree SHA `a849d99d2` (full `a849d99d2421cc43fb15fef35c61c74da279cb03`), 관측일 2026-09-08. 이 문서에 개별 pin이 없는 AC 는 이 문서 수준 pin에 묶인다 (verification-completeness §2.1).
- 각 AC는 RED-now 셀 + green 경로 셀의 쌍이다. green 경로는 명시한 마일스톤이 뒤집는다.
- AC-DWC-004..014 의 RED-now 셀은 §D 도입부의 baseline 관측(부재 0히트·H2 7×4·대조군 2×4·절 순서)을 **공유 셀**로 참조해 유지된다 — 각 셀이 RED를 다시 쓰지 않는 것은 의도다: 같은 관측을 반복 인용하면 실측이 사본으로 희석된다. 공유 셀은 문서 수준 pin(`a849d99d2`)에 묶이고 모두 이번 트리에서 재실행 가능한 형태다 (plan-audit iter1 D4 명시).

## §D AC Matrix (14 — Tier M 상한 16 이내)

### AC-DWC-001 — 부재 역전 (RED-now → M4)

- RED-now: `/usr/bin/grep -rn 'Codex Wiring' docs-site/` → stdout 없음, **exit 1** (baseline `a849d99d2`, 2026-09-08). 빠진 이유: docs-site에 Codex Wiring 콜아웃이 아직 없어서다 — 셸 grep 왜곡이 아니라 실제 부재다 (대조군 AC-DWC-002 참조).
- Given docs-site 4로케일 doctor.md 에 콜아웃이 착지한 tree / When `/usr/bin/grep -rn 'Codex Wiring' docs-site/` 실행 / Then exit 0, 히트가 최소 4개(로케일당 doctor.md 1개 이상)이며 전부 새 절 안이다.

### AC-DWC-002 — 대조군 (베이스라인 유지)

- Given baseline `a849d99d2` / When 로케일당 평문 단일 명령 `/usr/bin/grep -c 'Home Disk Usage' docs-site/content/<l>/cli-reference/doctor.md` (l = ko, en, ja, zh — 워크트리 가드가 루프·복합형을 거부하므로 네 명령으로 나눠 실행) / Then 4로케일 각 2히트 (관측값). Then 콜아웃 작업 후에도 이 값이 유지된다 — 작업이 기존 절을 훼손하지 않았다는 음성 증거.

### AC-DWC-003 — H2 패리티 7→8

- RED-now: `grep -c '^## '` ×4로케일 = **7·7·7·7** (baseline `a849d99d2`). 빠진 이유: 새 절이 아직 없어서다.
- Given M1–M3 착지 / When H2 카운트 재측정 / Then 4로케일 모두 8.

### AC-DWC-004 — 절 제목·배지 형태

- Given 각 로케일 doctor.md / When `/usr/bin/grep -n 'Codex Wiring' docs-site/content/<l>/cli-reference/doctor.md` / Then H2 행이 `## Codex Wiring 진단|check|診断|诊断 {{< new-badge v3.1.4 >}}` 형태이고, 이 절 범위 안의 `v3.1.3` 문자열은 0히트다 (REQ-DWC-002·008 — 배지를 v3.1.3으로 바꾸는 뮤턴트를 이 AC가 잡는다).

### AC-DWC-005 — 절 배치

- Given 각 로케일 doctor.md / When Hook Delivery 절 헤딩 행 번호 H_hook, 새 절 헤딩 행 번호 H_new, "종료 코드"/"Exit codes" 헤딩 행 번호 H_exit을 관측 / Then `H_hook < H_new < H_exit`가 4로케일 모두 성립한다 (REQ-DWC-003).

### AC-DWC-006 — 크로스링크

- Given 각 로케일 doctor.md / When `/usr/bin/grep -c 'advanced/codex-dual-harness' docs-site/content/<l>/cli-reference/doctor.md` / Then 로케일별 자기 경로(`/ko/`, `/en/`, `/ja/`, `/zh/` 접두) 히트 ≥1. 그리고 링크 대상 파일이 4로케일 모두 실재한다 (baseline 관측: `docs-site/content/{ko,en,ja,zh}/advanced/codex-dual-harness.md` 존재 확인 완료).

### AC-DWC-007 — 수정 지시문 코드 충실 (4종)

- Given 새 절 본문 / When 4 지시문 존재 검사 / Then (1) `moai init --agent codex`, (2) `codex /hooks` (재신뢰 맥락), (3) `moai update --templates-only --force --yes`, (4) 스테일 항목 조언("제거 또는 스킬 파일 복원")이 로케일 관례 번역으로 존재한다. 그리고 `moai clean --codex-skills` 이 doctor 지시문이 아니라 별도 동사로만 등장한다 (REQ-DWC-004 — 뮤턴트 "doctor가 clean을 가리킴"을 잡는다).

### AC-DWC-008 — fatal 서술 정확성

- Given 새 절 본문 / When fatal 문단 검사 / Then `enabled` bare TOML 불리언(`enabled = true`/`enabled = false`) 요구, codex `0.153.4` 실측 한정, "이 검사의 유일한 fatal" 서술, "그 외 발견은 조언형이며 종료 코드에 영향 없음" 서술이 4로케일 모두 존재한다 (REQ-DWC-005).

### AC-DWC-009 — un-nagging 서술

- Given 새 절 본문 / When claude-only 미배선 케이스 검사 / Then "코덱스 없는 환경의 claude-only 프로젝트는 조용히(정보성으로) 건너뛴다"는 서술이 4로케일 모두 존재한다 (REQ-DWC-006).

### AC-DWC-010 — 사실·수치 verbatim 보존

- Given 4로케일 새 절 / When 명령·버전·TOML 토큰 추출 / Then `moai init --agent codex`, `moai update --templates-only --force --yes`, `moai clean --codex-skills`, `enabled = true`, `0.153.4` 토큰이 로케일 간 동일하다 (REQ-DWC-009 — 번역 금지 토큰).

### AC-DWC-011 — 문서 관례 스윕

- Given 4로케일 diff / When (1) 새 절 범위 이모지 스캔 0히트, (2) 금지 URL(`docs.moai-ai.dev`, `adk.moai.com`, `adk.moai.kr`) 0히트, (3) Mermaid LR/RL 0히트, (4) 강조-괄호 간격 위반 0건 / Then 전부 충족 (REQ-DWC-009).

### AC-DWC-012 — hugo 빌드 무경고

- Given M4 / When `cd docs-site && hugo --minify` / Then exit 0, 신규 warning 0 (REQ-DWC-013).

### AC-DWC-013 — 쓰기 표면 범위

- Given run-phase 커밋 후보 diff / When `git diff --stat` / Then 변경 파일이 `docs-site/content/{ko,en,ja,zh}/cli-reference/doctor.md` 정확히 4개이고 `.moai/specs/SPEC-DOCS-CODEX-WIRING-CALLOUT-001/**` 외 다른 경로가 없다 (REQ-DWC-010).

### AC-DWC-014 — 4로케일 동일 변경 집합

- Given M4 / When 로케일별 새 절 존재 검사 / Then ko·en·ja·zh 네 파일이 같은 변경 집합(같은 커밋 후보 diff) 안에 있고 부분 로케일 착지가 아니다 (REQ-DWC-001).

## §D.1 심각도·추적

- release-blocking: AC-DWC-001, 003, 004, 005, 006, 007, 008, 013, 014 (결함 본질·허위 귀속 방지·코드 충실)
- regression-guard: AC-DWC-002, 009, 010, 011, 012 (베이스라인 유지·관례 스윕 — RED가 규약상 항상 재실행 가능하나 일부는 음성 증거 성격)
- 추적: AC-DWC-001→REQ-DWC-012·014 / 002→REQ-DWC-014 (대조군·baseline 유지는 부재 검증 규약의 양성 needle 축) / 003→REQ-DWC-002 / 004→REQ-DWC-002·008 / 005→REQ-DWC-003 / 006→REQ-DWC-007 / 007→REQ-DWC-004·011 / 008→REQ-DWC-005 / 009→REQ-DWC-006 / 010→REQ-DWC-009 / 011→REQ-DWC-009 / 012→REQ-DWC-013 / 013→REQ-DWC-010 / 014→REQ-DWC-001
- REQ-DWC-015 (t538 표면 미흡수) 는 AC-DWC-013 (쓰기 표면 4파일 한정) 을 통해 간접 적용된다 — 4파일 밖 변경 금지가 곧 t538 소관 표면의 미흡수 집행이다. plan-audit iter1 D1 에서 선언된다.

## §D.2 경계 케이스

- baseline 이 이미 Codex Wiring 절을 가진 경우(병합 흡수 등): AC-DWC-001 RED-now가 성립하지 않으므로 run 진입 전 blocker 보고 — 이 SPEC 은 no-op으로 종결 후보다.
- 로케일별 헤딩 행 번호가 다른 경우: AC-DWC-005는 절대 행이 아니라 상대 순서(H_hook < H_new < H_exit)로 판정한다.
- ugrep 래퍼 환경: 부재 판정 재실행은 반드시 `/usr/bin/grep`; `grep -c` 카운트는 순수 카운트라 래퍼 영향이 없음을 확인한 뒤 사용한다.

## §D.3 Definition of Done

- 14 AC 전수 PASS (regression-guard 포함) + hugo 빌드 무경고 + 4파일 범위 판정.
- run-phase 완료 보고가 VCI §3 5단 형식의 §E 증거를 동반한다.
