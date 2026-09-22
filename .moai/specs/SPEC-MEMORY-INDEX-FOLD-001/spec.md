---
id: SPEC-MEMORY-INDEX-FOLD-001
title: "MEMORY.md 인덱스 유실 3건 수복 + 접기 판정 기록 — 세션 기억 색인 무결성 (카드 t1065)"
version: "1.0.0"
status: completed
created: 2026-09-22
updated: 2026-09-22
author: GOOS행님
priority: P2
phase: "v14.4.0 target"
module: "~/.moai/claude-profiles/moai-adk/projects/-Users-goos-MoAI-moai-adk-go/memory (repo 외부) + .moai/reports/t1065"
lifecycle: spec-anchored
tags: "memory-index, index-integrity, fold-judgment, restoration, evidence-bearing-report, card-t1065"
tier: S
---

## HISTORY

- 2026-09-22 — v1.0.0 — manager-spec — Plan-phase 산출물 작성 (Tier S, 3 artifacts: spec.md + plan.md + progress.md §E skeleton; AC는 §3 인라인). 카드 t1065의 원 처방(접기)은 리드 세션의 09-22 병행 인덱스 정비(99항목 중 89개 재작성, 훅 문구 약 −11,490 chars [APPROXIMATE — 분해 수치는 근사치])로 이미 과잉 이행됨(91.3% → 53.8%). 본 SPEC의 실제 작업은 (1) 재계측 보고, (2) 무접기 판정 기록, (3) 정비 과정에서 색인에서 유실된 3개 포인터 라인의 2차 색인 복원, (4) 무손상 기계 검증, (5) 5섹션 증거 보고. `status: draft`.

---

## 1. 배경 (Background)

대상 파일은 저장소 외부의 세션 기억 색인이다:
`/Users/goos/.moai/claude-profiles/moai-adk/projects/-Users-goos-MoAI-moai-adk-go/memory/MEMORY.md`

- **읽기 절단 한도**: 24,986 chars (고정 전제)
- **경고 목표**: 17,510 chars (고정 전제)
- **단위**: chars (`python3 len()`) — 바이트 아님. `wc -m`은 이 머신에서 바이트를 반환함이 실측됨 (progress.md § 기재)
- **처방**: FOLD(2차 색인 파일로 라인 이전)만 허용, 삭제 금지
- **전제 변천**: 09-21 카드 작성 시점 22,810 chars / 38,281 bytes = 한도의 91.3% → 리드 세션의 09-22 병행 정비 후 13,408 chars / 18,412 bytes = 53.7%. **카드의 두 임계값은 이미 충족 상태.**
- **잔여 결함 (본 카드의 실작업)**: 정비 과정에서 색인 라인 3건이 어느 파일에도 보존되지 않음. 토픽 파일은 모두 존재(사실은 생존), 색인 발견 가능성만 소실:
  1. `project_lead_batch_20260921.md` 포인터 (리드 배치 09-21 — 인계 본문은 `.moai/reports/lead/handoff-20260921.md`)
  2. `feedback_a_requirement_map_no_instrument_reads.md` 포인터
  3. `feedback_a_falsifying_row_needs_the_survivor_on_the_far_side.md` 포인터

유실 라인의 원문은 `evidence/t1065_diff.txt` § DROPPED ENTRIES에, 정비 전 전체 이미지는 `evidence/t1065_before.md`에 보존되어 있다(/tmp 의존 제거를 위해 SPEC 증거 디렉터리로 복사 완료, plan-phase 시행).

## 2. GEARS 요구사항 (Requirements)

### REQ-MIF-001 — 재계측 보고

**When** the run phase begins, the lane shall re-measure `MEMORY.md` — chars via `python3 len()` (char authority) cross-checked against bytes via `wc -c` — and record a timestamped side-by-side table against the 24,986-char limit and the 17,510-char target with percentages, **before any write to any memory file**.

### REQ-MIF-002 — 접기 판정 기록

The lane shall record a NO-ADDITIONAL-FOLD judgment **when** the measured chars ≤ 17,510, citing (a) the premise-evolution chain (91.3% → 53.8%, mechanism = the lead session's 09-22 parallel index maintenance, hook-text rewrite ≈ −11,490 chars, APPROXIMATE — 급감 분해는 양끝점에 폐합되지 않는 근사치), (b) the card's conservative-keep clause, and (c) the card's citation-required clause ("오래됨" is not a discriminator). The lane **shall not** edit `MEMORY.md` under any circumstance within this SPEC.

### REQ-MIF-003 — 유실 3건 수복

The lane shall append the 3 dropped index lines verbatim — zero rewording, sourced from `evidence/t1065_diff.txt` § DROPPED ENTRIES — as a blank-line-separated block at EOF of the proper 2nd-level index:

- 항목 2 (`feedback_a_requirement_map_no_instrument_reads.md`) + 항목 3 (`feedback_a_falsifying_row_needs_the_survivor_on_the_far_side.md`) → `feedback_index_lessons_202609.md`
- 항목 1 (`project_lead_batch_20260921.md`) → `project_card_archive_2026_09.md`

Each append shall be a pure addition (additive doctrine); no existing line of either target file shall be modified or removed.

### REQ-MIF-004 — 무손상 검증

**When** the appends complete, the lane shall verify mechanically, in one parallel read-only batch: (a) each of the 3 lines matches verbatim (`grep -cF` = 1 in its target), (b) the `MEMORY.md` sha256 captured immediately before the append window equals the sha256 captured immediately after (untouched), (c) all 3 link targets exist on disk, (d) each target index's diff against its pre-append capture contains exactly the appended block and nothing else.

### REQ-MIF-005 — 보고

The lane shall write `.moai/reports/t1065/verdict.md` in the 5-section evidence-bearing format (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk), including (a) the premise-evolution narrative for the lead and (b) the 20 pre-existing orphan card files as a report-only list t223–t444는 7~8월 시대, t1019는 9월 시대 — 결함 축은 존재가 아니라 도달성: **어떤 색인 라인도 해당 토픽 파일을 링크하지 않음** (no index LINE LINKS the topic file; t1019는 다른 항목의 주석에 텍스트로 언급되지만 링크는 없음); 본 카드 내 복원 금지).

## 3. 인라인 수용기준 (Acceptance Criteria — Tier S inline)

- AC-001: 재계측 표 존재 — maps REQ-MIF-001

**Given** the run phase has started **When** the lane re-measures `MEMORY.md` with `python3 -c "len(open(path).read())"` and `wc -c` **Then** a timestamped table records both figures side-by-side against 24,986 / 17,510 with percentages, and the measurement timestamp is ISO-8601 KST.

- AC-002: 판정 기록 + 자기 쓰기 창 내 MEMORY.md 무편집 — maps REQ-MIF-002

**Given** measured chars ≤ 17,510 **When** the fold judgment is written **Then** it cites the premise-evolution chain and the two card clauses, and the lane's OWN write window shows zero edits to `MEMORY.md` — the sha256 captured immediately before the append window equals the sha256 captured immediately after (re-baseline semantics per REQ-MIF-004; the whole-run-window sha equality is explicitly NOT claimed — concurrent lead-session edits would spuriously fail it).

- AC-003: 3 라인 verbatim 존재 — maps REQ-MIF-003

**Given** the evidence-dir copy of the dropped lines **When** the lane runs a `python3` containment check (`line in pathlib.Path(target).read_text()` → True) for each of the 3 lines — the lines contain backticks, so double-quoted shell interpolation (`grep -cF "<line>"`) is prohibited (command-substitution hazard) **Then** all 3 return True, each in its designated target file, as a blank-line-separated block at EOF.

- AC-004: 순수 추가 검증 — maps REQ-MIF-004

**Given** pre-append captures (sha256 of `MEMORY.md`; full copies of both target indexes) **When** the post-append diff is computed **Then** each index diff contains exactly the appended block and nothing else, and all 3 link-target files exist (`test -f` exit 0).

- AC-005: 판정서 존재 — maps REQ-MIF-005

**Given** REQ-MIF-001..REQ-MIF-004 are complete **When** `.moai/reports/t1065/verdict.md` is read **Then** it carries all 5 sections non-empty, names the 20 orphans explicitly, and every cited evidence path resolves at read time.

### 동시성 위험 관리 (cross-cutting)

리드 세션이 `MEMORY.md`를 계속 편집 중임이 관측됨(plan-phase 중 13:00:33 KST +44 bytes 편집 실측). 모든 재계측·해시는 쓰기 직전/직후에 재획득하며 — `MEMORY.md`뿐 아니라 2개 대상 인덱스 파일(`feedback_index_lessons_202609.md`, `project_card_archive_2026_09.md`)에도 동일 적용 — 실행 중 변경 감지 시 REQ-MIF-001를 재베이스라인하고 델타를 보고한다 — 두 베이스라인의 측정값을 혼용하지 않는다.

## 4. 제약 (Constraints)

- 기억 디렉터리는 저장소 **외부** — 절대 경로 필수; 워크트리 상대경로 규칙 미적용.
- `MEMORY.md` 자체는 본 카드에서 **편집 금지** — 모든 쓰기는 2차 색인 파일에 대한 추가만.
- 저장소 코드 변경 없음 — 쓰기 범위는 본 워크트리(.moai/specs/, .moai/reports/) + 기억 디렉터리 추가뿐.
- 접기 판정 드리프트: 실행 중 chars > 17,510으로 상승 시에도 카드 내 접기 금지 — 보고 후 리드 에스컬레이션 (DP1 확정).

## 5. 제외 (Out of Scope)

### Out of Scope — MEMORY.md 자체 편집 및 접기

- `MEMORY.md`에 대한 어떤 편집(접기, 재작성, 정리)도 본 SPEC 범위 밖 — headroom 이득 보존.

### Out of Scope — 기존 고아 카드 파일

- 20개 기존 고아 카드 파일(t223–t444는 7~8월 시대, t1019는 9월 시대 — 결함 축은 링크 부재: 어떤 색인 라인도 해당 토픽 파일을 링크하지 않음, REQ-MIF-005 참조)의 복원 또는 정리 — report-only로 판정서에 열거만.

### Out of Scope — 저장소 코드 변경

- `internal/`, `pkg/`, `cmd/`, 템플릿 등 저장소 소스에 대한 일체의 변경.
