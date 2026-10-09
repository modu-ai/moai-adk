---
id: SPEC-RELUP-DUALAXIS-001
title: "acceptance — release-update 하네스 CC+Codex 이중 축 정착"
created: 2026-10-09
author: manager-spec
tier: M
---

# acceptance: SPEC-RELUP-DUALAXIS-001

## §A 범위·판정 규율

- 모든 릴리스 블로킹 AC는 verification-completeness.md §2 two-cell 규율을 따른다: **RED-now 셀**(구현 전 트리에서 적색 관측 — 단일 읽기전용 명령 + 그 명령의 축자 stdout + exit code 독립 필드 + 트리 SHA 핀)과 **green-path 셀**(어느 마일스톤이 뒤집는지 + 녹색 출력 형태)이 쌍으로 존재한다.
- **트리 핀**: 전체 RED-now 관측은 `2aab5f797` (worktree `.moai/worktrees/t1579`, branch `WT-high-10-07`)에서 수행했다. 문서 수준 핀이 이 문서이며, 개별 AC 핀이 없는 한 이 핀이 구속한다.
- **RED의 올바른 이유**: 각 RED는 "이 SPEC이 바꿀 표면이 오늘 비어 있다"는 이유로 적색이다 — 구현이 그 표면을 채우면 뒤집힌다. 선존재 파일이 못 만지는 wrong-reason red는 없다.
- **명령 형태**: 전부 단일 호출(파이프·리다이렉트·`&&`·`;`·서브셸 없음). grep 일치 0개는 exit 1이다 — "빈 출력 + exit 1"은 완전한 관측이다.
- **비재현 관측 처분**: 재실행 불가능한 관측(예: internal/ 0힛 — 이미 녹색인 부재 클레임)은 회귀 가드로 분류하고 릴리스 블로킹에서 제외한다(undecidable disposition).

## §D AC Matrix

| AC | 분류 | 주장 (Then) | RED-now (LED) | Green path |
|----|------|-------------|---------------|------------|
| AC-RDX-001 | 블로킹 | manifest `domain` 필드가 `Codex CLI upstream change tracking`을 명명 (domain 키 행 스코프 — CX-3) | LED-001 (0/1) | M4 → ≥1/0 |
| AC-RDX-002 | 블로킹 | manifest `domain` 필드가 best-practices 축을 명명 (동일 스코프) | LED-002 (0/1) | M4 → ≥1/0 |
| AC-RDX-003 | 블로킹 | runner에 codex 렌즈 셀렉터가 정의되고 top-level 디스패치 블록에서 호출된다 (`selectCodexSweepTargets(args)` ≥2 — CX-2; 실질 생성 면 plan §E3-P2) | LED-003 (0/1) | M2 → ≥2/0 |
| AC-RDX-004 | 블로킹 | runner에 커밋 복원 폴백 앵커 `CODEX_COMMITS_FALLBACK` 존재 | LED-004 (0/1) | M2 → ≥1/0 |
| AC-RDX-005 | 블로킹 | runner에 6테마 체크리스트 앵커 `CODEX_THEME_CHECKLIST` 존재 | LED-005 (0/1) | M2 → ≥1/0 |
| AC-RDX-006 | 블로킹 | specialist가 `last-codex-version.json` 스키마를 문서화 | LED-006 (0/1) | M1 → ≥1/0 |
| AC-RDX-007 | 블로킹 | specialist가 시드 `rust-v0.161.0`을 기술 | LED-007 (0/1) | M1 → ≥1/0 |
| AC-RDX-008 | 블로킹 | specialist에 BP 상시 절차 섹션 존재 | LED-008 (0/1) | M3 → ≥1/0 |
| AC-RDX-009 | 블로킹 | specialist Phase 3 URL 세트가 `code.claude.com` 캐노니컬을 명명 | LED-009 (0/1) | M3 → ≥1/0 |
| AC-RDX-010 | 블로킹 | specialist가 `HTML proposal report` 산출물을 명명 | LED-010 (0/1) | M3 → ≥1/0 |
| AC-RDX-014 | 블로킹 | specialist 스키마 블록이 codex 상태 파일 부재 시 기본값(`rust-v0.161.0` + 경고)을 문서화 | LED-006 공유 (0/1) | M1 → LED-006 ≥1 + 블록 내 기본값 기술 |
| AC-RDX-015 | 블로킹 | runner 체크리스트 블록이 alpha watch 규범(watch 관찰목록 전용, 안정 탑재 시에만 채택)을 담는다 | LED-005 공유 (0/1) | M2 → LED-005 ≥1 + 블록 내 watch 규범 기술 |
| AC-RDX-016 | 블로킹 | specialist BP 절차가 `source-first` 원문-패치 선행 강제를 명명 | LED-014 (0/1) | M3 → ≥1/0 |
| AC-RDX-011 | 회귀 가드 | `internal/`에 `last-codex-version` 참조 0힛 유지 (Go 라이터 부재 보존) | — (오늘 녹색 — 부재 클레임, 비재현) | 유지 조건: run-phase 전체 |
| AC-RDX-012 | 회귀 가드 | specialist의 `last-cc-version.json` 문서화 ≥3힛 유지 (CC 축 절차 보존) | — (오늘 녹색 3힛) | 유지 조건: run-phase 전체 |
| AC-RDX-013 | 회귀 가드 | manifest의 `hns-release-update-run.js` 참조 1힛 유지 + `sprint_contract` dimensions·thresholds 판독 기준선 일치 (LED-013 + LED-015 — CX-3 판독면) | — (오늘 녹색: 1힛 + LED-015 기준선 출력) | 유지 조건: run-phase 전체 |

## §D.1 시나리오 (Given-When-Then — 블로킹 13종)

- **AC-RDX-001** — **Given** manifest.json이 CC 단일 domain 문자열을 담은 상태로, **When** LED-001 명령(`domain` 키 행 스코프)을 실행하면, **Then** 일치 개수가 1 이상이다 (domain 필드가 codex 축을 명명 — source_request의 동일 문구는 매치 제외, CX-3).
- **AC-RDX-002** — **Given** 동일 상태로, **When** LED-002 명령(동일 스코프)을 실행하면, **Then** 일치 개수가 1 이상이다 (domain 필드가 best-practices 축을 명명).
- **AC-RDX-003** — **Given** runner가 CC 렌즈만 fan-out하는 상태로, **When** LED-003 명령을 실행하면, **Then** `selectCodexSweepTargets(args)` 출현이 2 이상이고 제2 출현은 top-level 디스패치 블록의 병합 호출이다 (정의 단독·주석 mutant는 1로 좌초 — CX-2). 실질 target 생성은 plan §E3-P2 어댑터가 시드 `codexDeltas`에 대해 실측한다 (mutant M-2 봉쇄).
- **AC-RDX-004** — **Given** runner에 커밋 복원 절차가 없는 상태로, **When** LED-004 명령을 실행하면, **Then** `CODEX_COMMITS_FALLBACK` 앵커가 1 이상 관측된다.
- **AC-RDX-005** — **Given** runner에 테마 관찰목록이 없는 상태로, **When** LED-005 명령을 실행하면, **Then** `CODEX_THEME_CHECKLIST` 앵커가 1 이상 관측된다.
- **AC-RDX-006** — **Given** specialist 본문에 codex 상태 스키마가 없는 상태로, **When** LED-006 명령을 실행하면, **Then** `last-codex-version.json`이 1 이상 관측된다.
- **AC-RDX-007** — **Given** AC-RDX-006이 충족된 상태에서도 시드가 빠질 수 있으므로(mutant M-3), **When** LED-007 명령을 실행하면, **Then** `rust-v0.161.0`이 1 이상 관측된다.
- **AC-RDX-008** — **Given** specialist에 BP 축이 없는 상태로, **When** LED-008 명령을 실행하면, **Then** best-practice 섹션이 1 이상 관측된다.
- **AC-RDX-009** — **Given** Phase 3 URL 세트가 docs.anthropic.com 구형 나열인 상태로, **When** LED-009 명령을 실행하면, **Then** `code.claude.com`이 1 이상 관측된다.
- **AC-RDX-010** — **Given** BP 산출물이 명명되지 않은 상태로, **When** LED-010 명령을 실행하면, **Then** `HTML proposal report`가 1 이상 관측된다.
- **AC-RDX-014** — **Given** codex 상태 파일이 존재하지 않는 다음 스윕 실행을 상정하는 상태로, **When** specialist의 스키마 블록(LED-006이 잡는 블록)을 읽으면, **Then** 부재 시 기본값 `rust-v0.161.0` + 경고 절차가 기술돼 있다 (REQ-RDX-004 — 스키마 문서화만으로 통과하는 mutant를 잡는 AC-RDX-006/007의 제3 쌍).
- **AC-RDX-015** — **Given** alpha 테마가 안정에 미탑재 상태로, **When** runner의 체크리스트 블록(LED-005이 잡는 블록)을 읽으면, **Then** watch 관찰목록 규범("alpha 테마는 채택 아님 — 안정 탑재 시에만 채택 판정")이 기술돼 있다 (REQ-RDX-009 — 1차 스윕 watch 판정의 절차화).
- **AC-RDX-016** — **Given** BP 절차에 원문 선행 강제가 없는 상태로, **When** LED-014 명령을 실행하면, **Then** `source-first` 리터럴이 1 이상 관측된다 (REQ-RDX-013 — mutant M-4의 기계 판정면).

## §D.2 추적성 (AC ↔ REQ)

| AC | REQ | mutant 봉쇄 |
|----|-----|-------------|
| AC-RDX-001/002 | REQ-RDX-010 | M-1 (source_request 기만) — 필드 스코프로 봉쇄 강화 (CX-3) |
| AC-RDX-003 | REQ-RDX-006 | M-2 (주석·미연결 정의) — ≥2 앵커 + §E3-P2 실측 생성 면 (CX-2) |
| AC-RDX-004 | REQ-RDX-007 | — (M-2 공유 봉쇄면) |
| AC-RDX-005 | REQ-RDX-008 | — |
| AC-RDX-006 | REQ-RDX-001/002 | — |
| AC-RDX-007 | REQ-RDX-002 | M-3 (seed 누락) |
| AC-RDX-008 | REQ-RDX-012 | — (섹션 존재면) |
| AC-RDX-009 | REQ-RDX-014 | — |
| AC-RDX-010 | REQ-RDX-014 | — |
| AC-RDX-014 | REQ-RDX-004 | M-3의 제3 쌍 (스키마 문서화 + 시드 + 부재 기본값 3중 분해) |
| AC-RDX-015 | REQ-RDX-009 | alpha-채택 오표기 mutant 봉쇄 |
| AC-RDX-016 | REQ-RDX-013 | M-4 (BP 껍데기 섹션 — 기계 판정면 확보) |
| AC-RDX-011 | REQ-RDX-005 | M-5 (Go 침입) |
| AC-RDX-012 | REQ-RDX-003 (보존 축) | — |
| AC-RDX-013 | REQ-RDX-011 (골격 보존) | threshold 편집은 LED-015 판독면에 걸린다 (CX-3) |

## §D.3 증거 원장 (Evidence Ledger — 트리 `2aab5f797`, 2026-10-09 관측)

각 행: 명령은 축자 그대로 단일 실행됐고, stdout·exit code는 같은 실행에서 관측했다. 경로는 워크트리 루트 기준.

| LED | 명령 (단일 호출) | stdout (축자) | exit | 판정 |
|-----|------------------|---------------|------|------|
| LED-001 | `grep -c '"domain".*Codex CLI upstream change tracking' .claude/commands/harness/release-update/manifest.json` | `0` | `1` | RED (AC-001) — domain 필드 스코프 (CX-3 재앵커) |
| LED-002 | `grep -c '"domain".*best-practices axis' .claude/commands/harness/release-update/manifest.json` | `0` | `1` | RED (AC-002) — 동일 재앵커 (CX-3) |
| LED-003 | `grep -c "selectCodexSweepTargets(args)" .claude/workflows/hns-release-update-run.js` | `0` | `1` | RED (AC-003) — 디스패치 호출 앵커, 착지 후 ≥2 (CX-2 재앵커) |
| LED-004 | `grep -c "CODEX_COMMITS_FALLBACK" .claude/workflows/hns-release-update-run.js` | `0` | `1` | RED (AC-004) |
| LED-005 | `grep -c "CODEX_THEME_CHECKLIST" .claude/workflows/hns-release-update-run.js` | `0` | `1` | RED (AC-005) |
| LED-006 | `grep -c "last-codex-version.json" .claude/agents/harness/hns-release-update-specialist.md` | `0` | `1` | RED (AC-006) |
| LED-007 | `grep -c "rust-v0.161.0" .claude/agents/harness/hns-release-update-specialist.md` | `0` | `1` | RED (AC-007) |
| LED-008 | `grep -ci "best-practice" .claude/agents/harness/hns-release-update-specialist.md` | `0` | `1` | RED (AC-008) |
| LED-009 | `grep -c "code.claude.com" .claude/agents/harness/hns-release-update-specialist.md` | `0` | `1` | RED (AC-009) |
| LED-010 | `grep -c "HTML proposal report" .claude/agents/harness/hns-release-update-specialist.md` | `0` | `1` | RED (AC-010) |
| LED-014 | `grep -c "source-first" .claude/agents/harness/hns-release-update-specialist.md` | `0` | `1` | RED (AC-016) |
| LED-015 | `python3 -c "import json;d=json.load(open('.claude/commands/harness/release-update/manifest.json'));sc=d['sprint_contract'];print(sc['dimensions'],sc['thresholds'])"` | `['Functionality', 'Consistency'] {'Functionality': 0.85, 'Consistency': 0.8}` | `0` | 회귀 가드 기준선 — 출력 불변 유지가 PASS (AC-013, CX-3 판독면) |
| LED-011 | `grep -rn "last-codex-version" internal/` | (출력 없음) | `1` | 회귀 가드 기준선 — 0힛 유지가 PASS (AC-011) |
| LED-012 | `grep -c "last-cc-version.json" .claude/agents/harness/hns-release-update-specialist.md` | `3` | `0` | 회귀 가드 기준선 — ≥3 유지가 PASS (AC-012) |
| LED-013 | `grep -c "hns-release-update-run.js" .claude/commands/harness/release-update/manifest.json` | `1` | `0` | 회귀 가드 기준선 — 1 유지가 PASS (AC-013) |

**LED-001/002/003 재앵커 근거**: plan-audit iter1(CX-2/CX-3)으로 위 세 행의 명령을 교체했다 — 재측정은 본 트리에서 수행했으며, 하네스 표면은 `2aab5f797` 핀 이후 `.moai/` 전용 변경으로 바이트 동일해 재관측이 충실하다. 나머지 LED 행은 원본 그대로다. LED-015의 세미콜론은 인용된 python 프로그램 내부의 것 — 셸 구분자가 아니므로 단일 호출 규약을 유지한다.

**보조 관측 (동일 트리)**: `grep -c "Codex" manifest.json` → `0`/exit 1 · `grep -ci "codex" runner` → `0`/exit 1 · `grep -c "HTML" specialist.md` → `0`/exit 1 · `grep -rn "last-cc-version" internal/` → 출력 없음/exit 1 (Go 라이터 부재 — 상태 파일이 하네스 계층 소유임의 근거, spec.md §1.1 M4). `git rev-parse --short HEAD` → `2aab5f797`.

## §D.4 간접 검증 항목

- **러너 파스**: plan §E3 CommonJS require() 스모크 — `node --check`의 무음 통과 한계 보강.
- **manifest JSON 타당성**: plan §E4 python json.load — domain 문자열 편집 후 문법 훼손 잡기.
- **spec-lint**: plan §E6 — `### Out of Scope —` h3 (MissingExclusions)·frontmatter 12 필드 (FrontmatterInvalid) 0건.
- **시드 신선도**: plan §C — run-phase 착지 직전 npm/gh 재측정으로 `rust-v0.161.0` 유효성 재판정 (0.162 승격 대응).

## §D.5 종결 게이트 (Definition of Done)

1. 블로킹 AC 13종 전부 GREEN (RED-now가 대응 마일스톤에서 뒤집힘 — exit code 포함 관측).
2. 회귀 가드 3종 기준선 유지 (LED-011/012/013 변화 없음).
3. spec.md REQ-RDX-001..014 전부 구현 대응물 존재 — REQ↔AC 추적성 §D.2 공백 없음.
4. Go 트리 변경 0 (git diff --name-only가 3개 하네스 파일만 반환).
5. `[NEEDS CLARIFICATION]` 마커 0개 — 열린 판단은 전부 spec.md §1.2 결정 기록으로 봉쇄.

## §D.6 선향 체크 (착지 후 다음 스윕이 검증할 것)

- 다음 `/harness:release-update` 실행이 codex 렌즈를 실제로 fan-out하는지 — 첫 실행 관측까지 렌즈 프롬프트의 커밋 복원 실효성은 미검증으로 남는다(spec.md §7).
- `last-codex-version.json` 첫 생성이 Phase 7a-codex 절차를 따르는지 — 기계 로컬이라 본문 문서화가 규범면.
- 0.162 alpha → stable 승격 시 6테마 관찰목록 승계가 절차대로 동작하는지.
