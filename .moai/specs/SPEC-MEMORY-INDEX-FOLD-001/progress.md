# SPEC-MEMORY-INDEX-FOLD-001 — Progress

## Phase 1 (Research) — SKIP rationale

Phase 1 리서치는 스폰 전에 레인 오케스트레이터가 완료함 (카드 t1065 dispatch 본문 + plan-phase 스팟체크). 완료된 리서치:

- 읽기 절단 한도 24,986 chars / 경고 목표 17,510 chars — 카드 전제로 FIXED (재도출 금지)
- 전제 변천 측정: 09-21 카드 작성 시 22,810 chars (91.3%) → 리드 세션 09-22 병행 정비로 13,408 chars (53.7%), 기제 = 훅 문구 재작성 −11,490 chars
- 유실 3건 확정: before-image(/tmp/t1065_before.md, 리드 트랜스크립트 348c3ff2 @ 11:21 KST) vs current 링크 타깃 집합 diff — plan-phase 13:1x 재측정에서 정확히 3건, 추가 유실 없음 재확인
- 증거 스테이징 완료: `/tmp/t1065_before.md`, `/tmp/t1065_diff.txt` → 본 SPEC `evidence/` 디렉터리로 복사 (plan-phase, 13:21 KST)

따라서 run-phase에서의 재리서치는 이중 작업 — SKIP 확정.

## Plan-audit repairs 기록 (iter-1, PASS-WITH-DEBT 0.90 → 수리 완료)

- **D1 재번호**: REQ-A..E → REQ-MIF-001..005 (spec.md §2/§3, plan.md §B/§F/§G, 잔여 토큰 grep 0건 확인).
- **D1 발견 — lint AC 라인 형태 규약**: 재번호 직후 lint가 REQ 5건 전부에 `CoverageIncomplete`를 냈다. 원인 실측: inline 커버리지 추출은 AC 라인(리스트 항목, `AC-…:` 콜론 형태) **같은 줄**의 `maps REQ-…` 키워드만 인식한다(`internal/spec/ears.go` `ExtractRequirementMappings` — `maps\s+(REQ-...)`; H3 괄호 참조 `(REQ-MIF-00X)`는 미인식). spec.md §3을 리스트-라인 형태(`- AC-00X: 제목 — maps REQ-MIF-00X`)로 재구성하여 해소. Given/When/Then 본문은 유지.
- **D1 긍정 대조 (양성 대조 실측 기록)**:
  1. baseline: `go run ./cmd/moai spec lint SPEC-MEMORY-INDEX-FOLD-001` → `✓ No findings — all SPEC documents are valid` (exit 0)
  2. 변이: AC-004 라인에서 `maps REQ-MIF-004` 토큰 제거 → lint → **관측된 발화줄(verbatim)**: `WARNING   CoverageIncomplete   .../SPEC-MEMORY-INDEX-FOLD-001/spec.md   45   REQ REQ-MIF-004 is not referenced by any AC` (`0 error(s), 1 warning(s)`) — lint의 REQ 계층이 5개 요구사항을 실제로 보고 커버리지 갭을 잡음을 입증. (1차 변이 시도는 토큰을 줄에 남겨 무력 변이였음을 관측하고 즉시 올바른 변이로 재수행)
  3. 복원 후: `✓ No findings — all SPEC documents are valid` (exit 0)
- **D2**: AC-002의 sha 주장을 레인 자기 쓰기 창으로 재범위 + 전체-실행-창 비교 명시적 미청구 (동시 리드 편집 시 위양성); plan §G M2에 유실 집합 ≠ 3건 분기(내용 보존 논리로 계속 + 중복 보고); 재베이스라인을 2개 대상 인덱스 파일까지 확장(3개 파일 모두).
- **D3**: AC-003 검증 템플릿을 `grep -cF "<line>"` → `python3` containment check로 교체 (유실 3라인에 백틱 존재 실측 — `evidence/t1065_diff.txt` 2-4행; 이중따옴표 셸 명령치환 위해).
- **D4**: 고아 서술을 "referenced from nowhere" → "어떤 색인 라인도 해당 토픽 파일을 링크하지 않음(no index LINE LINKS)"; t1019는 9월 시대로 정정(t223–t444만 7~8월).
- **D5**: 급감 분해(−11,490 / −614 / +2,115)를 APPROXIMATE로 표기 — spec.md HISTORY·REQ-MIF-002, plan.md §B, `measurements-plan.md` 근사치 고지 추가 (폐합 갭 ≈622/176자, 99−3+10=106≠108).
- **감사 residual**: plan §G M1 선두 단계로 `moai memory doctor` 추가 (two-stores 아티팩트 대응, 출력 progress.md 기록 + 불일치 시 리드 에스컬레이션).

## Plan-audit iter-2 마무리 (D6–D8 closure)

- D6: AC-005 Given의 스테일 범위 단편 `REQ-MIF-001..D` → `REQ-MIF-001..REQ-MIF-004` (spec.md).
- D7: AC-004의 `**Given** pre-append captures (…)` 마커 복원 — 긍정 대조군 변이 중 1차 무력 변이 시도가 `\n\n**Given** pre-append captures` 절을 소거해 본문이 maps 라인에 융합된 상태였던 것을 분리 복원 (다른 4개 AC와 동일 형식).
- D8: §5 Out of Scope의 스테일 서술 `t223~t1019, 7~8월 시대` → 수정 형태로 정렬 (t223–t444 7~8월, t1019 9월 — 링크 부재 축, REQ-MIF-005 참조).
- 3건 치환 후 lint 재실행: `go run ./cmd/moai spec lint SPEC-MEMORY-INDEX-FOLD-001` → `✓ No findings — all SPEC documents are valid`, exit 0.

## Plan-phase 측정 기록 (wc -m 로케일 함정)

- `LC_ALL=en_US.UTF-8 wc -m` 이 이 머신(darwin)에서 바이트 수(18456)를 반환함을 plan-phase에서 실측 — char 권위는 `python3 len()`, 바이트는 `wc -c`. run-phase 재계측도 이 이원화를 따른다.
- plan-phase 시점 측정: 13,443 chars / 18,456 bytes (mtime 2026-09-22T13:00:33 KST) — 한도의 53.8%, 목표의 76.8%. 참고값이며 run-phase 재계측값이 유일 유효 베이스라인.

## Decision Point 1 기록 (DP1)

- 운영자 60초 타임아웃 → 오케스트레이터 best-judgment continuation으로 PROCEED (기록basis: 수정 방향이 원 dispatch가 명령한 full plan→run→sync transit보다 엄격히 협소 — 본 기록은 progress.md 보존 대상으로 지정됨).
- NEEDS CLARIFICATION 3건 전부 제안대로 확정: (1) verdict.md = 본 워크트리 `.moai/reports/t1065/verdict.md`, (2) 접기 드리프트 = 보고-에스컬레이션(카드 내 접기 금지), (3) verbatim zero-rewording + 빈 줄 분리 블록 EOF 추가.

## Kickoff 게이트 기록 (Implementation Kickoff Approval)

- 2026-09-22 14:12 KST AskUserQuestion 60초 무응답 ×1 — [HARD] 게이트라 조용한 진행 금지, 정지 상태로 리드 서지 보고 (t1064 선례 준수).
- 재개 경로: 운영자의 직접 응답 또는 리드의 운영자-응답 중계(출처 명시 시). 리드 단독 판단은 게이트를 넘지 못한다 — 리드가 할 수 있는 것은 배차 정정 또는 카드 철회(취소)다.
- 대기 중인 선택지: 승인(권장) / 승인+goal 부착 / 보류(draft 정지) / 취소(전제 소멸, 유실 수복은 별도 소카드 분리).
- **승인 수신 (run 진입)**: 리드 크로스-세션 메시지로 운영자 응답 중계 — «card: t1065 · Kickoff 승인(운영자 "전부 승인" 09-22)» (출처: 리드 세션, uds:/tmp/cc-socks/55591.sock). 출처 명시된 운영자-응답 중계로 본 progress.md가 정의한 재개 경로 충족 → run 진입.

## §F Phase 4 Mode Selection

- **입력 파라미터**: tier S · 범위 ≈3파일(메모리-dir 2차 색인 2곳 + MEMORY.md sha 감시, 읽기전용) + SPEC 아티팩트 3종 · 도메인 수 1(메모리-색인 유지보수, repo 코드 0) · 언어 믹스 100% markdown/데이터 · 동시성 이익 LOW(순차 추가 + 살아있는 리드 편집기와의 재기준 프로토콜) · Agent Teams 사전조건 해당 없음.
- **모드 평가**: `direct` 미선정(쓰기는 전문가 위임 대상 — 계측·추가·검증의 격리 이익) · `serial` **선정** · `fanout` 미선정(단일 도메인, 병렬 리서치 없음, 동시 스폰이 살아있는 편집기·재기준 프로토콜과 경합) · `sweep` 미선정(3줄 추가는 고볼륨 기계 변환이 아님).
- **Decision: serial**
- **근거**: 단일 도메인의 순차 변이 작업이고 쓰기 창마다 sha 재기준이 필요해 병렬 스폰이 정확성을 해친다. Anthropic의 코딩-작업 병렬성 경고(대부분의 구현 작업은 진병렬화 이익이 낮다)의 아날로그로, 측정→추가→검증의 의존 사슬이 직렬을 요구한다. 경계 사례 없음.
- run-phase 실행 전 계약: Phase 1 게이트는 iter-2 판정 이후 D6-D8 치환으로 해시가 변해 스킵 계약 조건 3 불충족 → 감사인 보증(델타=정확히 3건 치환)에 따른 **델타 한정 재감사(iter-3)로 재실행**. iter-3 판정은 `.moai/reports/t1065/plan-audit-iter3.md`에 착륙.

## §E.1 Plan-phase Audit-Ready Signal

- status: `draft` 설정 완료 (4 artifacts 중 Tier S 3종: spec.md + plan.md + progress.md; acceptance.md는 Tier S 생략, AC는 spec.md §3 인라인)
- SPEC ID regex 체크: Bash 실행 `PASS` 출력 인용 (plan-phase 13:0x)
- ID uniqueness: `SPEC-MEMORY-DIET-001`, `SPEC-MEMORY-STORE-RECONCILE-001` 외 충돌 없음 확인
- Frontmatter 12 필드 + tier: 스키마 준수 (형제 SPEC-MEMORY-DIET-001 양식 대조)
- Out of Scope: `### Out of Scope —` H3 3개 + 불릿 충족 (OutOfScopeRule)
- SPEC lint 최종: `go run ./cmd/moai spec lint SPEC-MEMORY-INDEX-FOLD-001` → `✓ No findings — all SPEC documents are valid`, exit 0 (plan-audit 수리 후, 긍정 대조 복원 확인 포함)
- plan_complete_at: 2026-09-22T14:09:18+09:00
- plan_status: audit-ready

## §E.2 Run-phase Evidence

- run 진입: 2026-09-22 22:0x KST · HEAD `00e761af8` · branch `WT-index-fold` · worktree `.claude/worktrees/t1065`
- 측정 규율: char 권위 = `python3 len()`, 바이트 교차검증 = `wc -c` (`wc -m` 미사용 — 이 머신에서 바이트 반환 실측, plan §B)
- Phase 1 Plan Audit Gate: SKIPPED — 3 조건 충족 (iter-3 PASS 0.94 ≥ Tier S 0.75 · plan-artifact hash 불변 · depends_on 부재 → trivially PASS). 근거: `.moai/reports/t1065/plan-audit-iter3.md`
- Kickoff: 운영자 "전부 승인" (09-22, 리드 중계) — 본 파일 상단 Kickoff 게이트 기록 참조

### M1 — 재계측 + NO-ADDITIONAL-FOLD 판정 (REQ-MIF-001, REQ-MIF-002)

**Step 0 — `moai memory doctor`** (exit 0). verbatim (`evidence/run/m1/memory_doctor.txt`):

```
/Users/goos/.moai/claude-profiles/moai-adk/projects/-Users-goos-MoAI-moai-adk-go/memory  (CLAUDE_CONFIG_DIR — primary checkout)
  topic files : 1089 (cap 50)
  index lines : 129
  findings    :
    MEMORY_ORPHAN_NOT_INDEXED      102
    MEMORY_INDEX_DUPLICATE_ENTRY   5
    MEMORY_TOPIC_COUNT_OVER_CAP    1
    → 1089 topic files; cap is 50 — archive 1039 into _archive/
(워크트리 접미 스토어 2개: not present — 생략 없음, 전문은 evidence 파일 참조)
```

판독: 세션 로드 스토어 = CLAUDE_CONFIG_DIR primary checkout (`/Users/goos/.moai/claude-profiles/moai-adk/projects/-Users-goos-MoAI-moai-adk-go/memory`) = **SPEC 대상 경로와 일치** → 진행. 워크트리 접미 스토어 2개는 `not present` — 비대상 확인. (dispatch 기재 128 lines → 실측 129: 리드 세션 병행 편집 드리프트. 본 실행의 유일 유효 베이스라인은 아래 22:09:22 재계측이다.)

**재계측 표 (AC-001)**:

| 측정 시각 (ISO-8601 KST) | chars (python3 len) | bytes (wc -c) | lines | 한도 24,986 대비 | 목표 17,510 대비 |
|---|---|---|---|---|---|
| 2026-09-22T22:09:22+09:00 | 15,972 | 22,009 | 129 | 63.9% | 91.2% |

측정 명령 + verbatim 출력:

```
$ python3 -c "print('CHARS_PY_LEN=' + str(len(open('<MEM>/MEMORY.md').read())))"
CHARS_PY_LEN=15972
$ wc -c <MEM>/MEMORY.md
   22009 /Users/goos/.moai/claude-profiles/moai-adk/projects/-Users-goos-MoAI-moai-adk-go/memory/MEMORY.md
$ wc -l <MEM>/MEMORY.md
     129 /Users/goos/.moai/claude-profiles/moai-adk/projects/-Users-goos-MoAI-moai-adk-go/memory/MEMORY.md
```

증거 파일: `.moai/reports/t1065/evidence/run/m1/` (memory_doctor.txt · chars_pylen.txt · bytes_wcc.txt · measurement_timestamp.txt). 사전 안정성 확인: 22:07:47 1차 측정(15,972)과 22:09:22 정본 측정이 동일 — 측정 안정.

**NO-ADDITIONAL-FOLD 판정**: **접지 않는다.** 측정 chars 15,972 ≤ 17,510 (REQ-MIF-002 발동 조건의 부정 충족).

근거 사슬:

1. **전제-변천**: 09-21 카드 작성 22,810 chars (91.3%) → 09-22 리드 병행 정비 13,408 (53.7%, 11:28) → plan-phase 13,443 (53.8%, 13:1x) → 본 실행 15,972 (63.9%, 22:09). 급감 기제 = 리드 세션의 인덱스 정비(99항목 중 89개 재작성, 훅 문구 재작성 ≈ −11,490 chars — **APPROXIMATE**, 분해 수치는 양끝점에 폐합되지 않는 근사치). 13,443→15,972 재증가분은 리드 세션이 낮 동안 오픈 카드 항목을 재추가한 것 — 인덱스 흡입 규율(moai-memory.md § Admission)상 오픈 작업의 MEMORY.md 적재는 정상 운영이며 결함이 아니다.
2. **conservative-keep 조항**: 목표 이내 상태에서의 추가 접기는 재수리 루프(접기→리필→재접기)만 만든다 — 카드의 보수적 유지 원칙.
3. **citation-required 조항**: "오래됨"은 유실 판별 근거가 아니다. 접힌 항목은 2차 색인 경유로 도달 가능해야 하며, 본 카드는 살아 있는 링크를 제거하지 않는다.

MEMORY.md 편집 금지 (REQ-MIF-002 후반부): 본 카드 전 기간 `MEMORY.md` 무편집. AC-002의 sha 창 검증은 M2 쓰기 창 직전/직후로 집행 (아래 M3).

### M2 — 유실 3건 수복 (REQ-MIF-003)

**유실 집합 재도출** (`evidence/run/m2/derive_lost_set.py` → `.out`): before-image 링크 타깃 105건 vs 현재 MEMORY.md 타깃 120건 → 차집합 **LOST_COUNT=3** — 기대 3건과 정확히 일치, 중복·불일치 0:

```
LOST=feedback_a_falsifying_row_needs_the_survivor_on_the_far_side.md
LOST=feedback_a_requirement_map_no_instrument_reads.md
LOST=project_lead_batch_20260921.md
```

(ADDED_SINCE_BEFORE=18건은 리드 세션의 낮 동안 신규 항목 추가 — 유실과 무관, 변경 없음.)

**verbatim 추출** (`evidence/run/m2/extract_dropped.py` → `.out`): `evidence/t1065_diff.txt` § DROPPED ENTRIES에서 `[U] ` 주석 접두를 벗겨 3라인 추출 — `EXTRACTED_OK fb_lines=2 ar_lines=1`, LINE_BYTES 374 (lead_batch) / 342 (requirement_map) / 363 (falsifying_row). 스크립트가 기대 집합 불일치 시 FATAL로 중단하는 가드 포함.

**쓰기 창 (22:16:12 → 22:16:51 KST)**:

| 파일 | pre sha256 (22:16:12) | post sha256 (22:16:51) | 변화 |
|---|---|---|---|
| `MEMORY.md` | `7b26aa1e a9615f` (전체 `7b26aa1ea5ae042e…31a9615f`) | `7b26aa1e a9615f` (동일) | **변화 없음** — 쓰기 창 내 무편집 (AC-002) |
| `feedback_index_lessons_202609.md` | `8ca72546 b54454a0d9` | `c7606e4e 685fe8005` | +2라인 추가 블록 |
| `project_card_archive_2026_09.md` | `a3ad58c9 810ec8bc` | `127edaa1 a00e31b5d` | +1라인 추가 블록 |

(위 표의 축약 표기는 요약용이며 전체 해시는 `evidence/run/pre/sha256_before.txt` · `evidence/run/post/sha256_after.txt`가 정본이다.)

**추가 내용**: Edit 툴로 각 2차 색인 EOF에 빈 줄 분리 블록 순수 추가 — `feedback_a_requirement_map_no_instrument_reads.md` + `feedback_a_falsifying_row_needs_the_survivor_on_the_far_side.md` 2라인 → `feedback_index_lessons_202609.md`, `project_lead_batch_20260921.md` 1라인 → `project_card_archive_2026_09.md`. 기존 라인 수정·삭제 0 (additive doctrine). 전문 사본: `evidence/run/pre/` · `evidence/run/post/` (3파일씩 full copy).

M3 무손상 기계 검증 결과는 §E.2 M3에 기재.

### M3 — 무손상 검증 + 판정서 (REQ-MIF-004, REQ-MIF-005)

단일 턴 병렬 read-only 배치 (2026-09-22 22:1x KST, HEAD `bff24dc6b`). verbatim 원문: `.moai/reports/t1065/evidence/run/m3/` · 판정서: `.moai/reports/t1065/verdict.md`.

**(a) containment ×3** (`m3/containment_check.py` → `containment.out`, exit 0):

```
CONTAINMENT target=project_lead_batch_20260921.md result=True
CONTAINMENT target=feedback_a_requirement_map_no_instrument_reads.md result=True
CONTAINMENT target=feedback_a_falsifying_row_needs_the_survivor_on_the_far_side.md result=True
CHECKED=3
VERDICT=ALL_CONTAINED
```

python3 containment — 백틱 포함 라인이라 셸 인용 템플릿 배제 (D3). CHECKED=3 상수로 빈 스윕 통과 차단.

**(b) MEMORY.md sha 창** (AC-002): `diff pre/MEMORY.md post/MEMORY.md` → **empty, exit 0** (`m3/diff_memory.txt`). sha256 `7b26aa1ea5ae042e…31a9615f` 전/후 항등 (`m3/sha_compare.txt` — MEMORY.md 라인은 diff 부재로 항등 입증). 레인 자기 쓰기 창 내 무편집. 전체-실행-창 항등은 청구하지 않음(AC-002 재베이스라인 의미론).

**(c) 링크 타깃 존재**: `test -f` ×3 → `ALL_3_LINK_TARGETS_EXIST`.

**(d) 인덱스 diff** (`m3/diff_feedback.txt` · `m3/diff_archive.txt`):

- feedback: `301a302,305` — 추가 4행(빈 줄 + 2라인 + 빈 줄) = 추가 블록 그 자체.
- archive: `343a344,345` — 추가 2행(빈 줄 + 1라인) = 추가 블록 그 자체.
- **관측 1건씩 (숨기지 않고 기재)**: 각 파일 `8c8` 헝크 — frontmatter `modified:` 타임스탬프 변경 (`13:16:31.033Z`/`13:16:38.280Z` UTC = 22:16:31/38 KST = 레인 쓰기 창 초 단위 내부, 간격 7초 = 두 Edit 호출 간격). 귀속: 런타임 메모리 서브시스템의 write-time 스탬핑 — 레인의 Edit 페이로드는 EOF 블록만 포함(§M2 추출본이 정본). **내용 라인(인덱스 항목) 수정·삭제 0** — additive doctrine 축은 영향 없음. AC-004 문언("exactly the appended block and nothing else")에 대한 관측 편차로 판정서 §4 Gaps #2 에 기록.

**(e) 판정서**: `.moai/reports/t1065/verdict.md` — 5섹션 완비 + 전제-변천 부록 A + 고아 20건 report-only 부록 B (t223~t444 19건 7~8월 시대, t1019 1건 9월 시대 — 결함 축=도달성, 링크 부재).

**범위 외 관측 (처분 기록)**: `.moai/reports/*` 는 gitignore(2026-09-14 운영자 지시, `.gitignore:235`) → M3 커밋은 progress.md 델타만 stage; 판정서·증거는 로컬 아티팩트(경로 상시 해석 가능). `-f` 강제 추가 미실시(운영자 지시 정면 위반 방지).

## §E.3 Run-phase Audit-Ready Signal

```
run_complete_at: 2026-09-22T22:21:18+09:00
run_commit_sha: 43b974f0b   # M3 커밋 (D3 backfill — backfill 커밋이 본 행을 실음)
run_status: complete
ac_pass_count: 4                      # AC-001, AC-002, AC-003, AC-005
ac_pass_with_observation: 1           # AC-004 — runtime `modified:` metadata line (§E.2 M3 (d) + verdict §4 Gaps #2)
ac_fail_count: 0
preserve_list_post_run_count: MEMORY.md 전체(129행, 카드 전 기간 편집 0) + 기존 인덱스 라인 전체(feedback 301행, archive 343행) 무손상
l44_pre_commit_fetch: N/A — 레인 push 없음(git-flow: develop push는 리드 일괄, 2026-09-02)
l44_post_push_fetch: N/A — 동상 (push 주체 아님)
new_warnings_or_lints_introduced: 0   # 코드 변경 0 — lint 대상 자체가 없음 (미실행을 0으로 오인 아님: N/A)
cross_platform_build: N/A             # 저장소 코드 변경 0
total_run_phase_files: SPEC 디렉터리 5파일(M1 커밋) + progress.md 델타(M2/M3) + 기억 디렉터리 2파일 추가 수정 + 로컬 증거 ~25파일(.moai/reports/t1065/, gitignored)
m1_to_mN_commit_strategy: 마일스톤별 커밋 — M1 d83e07033 · M2 bff24dc6b · M3 (backfill 커밋 참조)
```

메모리 저장소 doctor 결과 요약: 로드 스토어 = CLAUDE_CONFIG_DIR primary checkout (SPEC 대상 일치, index lines 129); 워크트리 접미 스토어 2개 not present. Phase 1 Plan Audit Gate: SKIPPED (3 조건 충족 — §E.2 헤더 참조).

## §E.4 Sync-phase Audit-Ready Signal

```
sync_complete_at: 2026-09-22T22:53:57+09:00   # STAGE 2 close 커밋 시점 확정 (sync-audit PASS-WITH-DEBT 0.94 수신 후)
sync_commit_sha: pending-backfill-sync   # D3 placeholder — 커밋은 자기 해시를 참조할 수 없으므로 STAGE 3 backfill 커밋이 실제 SHA를 기입
sync_status: completed                   # sync-audit PASS-WITH-DEBT 0.94 (차단 결함 0) — close 커밋에서 확정
```

### Sync 범위 요약 (docs-only close)

- 본 SPEC의 저장소 트리 델타 = `.moai/specs/SPEC-MEMORY-INDEX-FOLD-001/**` 유일 — 저장소 코드·README·docs-site 등 사용자 대면 표면 변경 0.
- 산출물 표면 2곳: (a) **외부 기억 저장소**(레포 외부 — 2차 색인 2파일 verbatim 추가 블록, MEMORY.md는 sha256 항등 무편집), (b) **SPEC 아티팩트 + 로컬 증거**(`.moai/reports/t1065/` — `.gitignore:235` 로컬 아티팩트, 커밋 대상 아님).
- **증거 경로 베이스 (sync-audit D1 흡수)**: 판정서·plan의 `evidence/t1065_diff.txt`·`evidence/t1065_before.md` 는 `.moai/specs/SPEC-MEMORY-INDEX-FOLD-001/` 기준(M1 커밋 트래킹), `evidence/run/…` 인용은 `.moai/reports/t1065/` 기준으로 해석한다.

### Run-phase 판정 요약

- **AC 5/5** — AC-001 / AC-002 / AC-003 / AC-005 PASS + **AC-004 PASS-with-observation**(런타임 `modified:` 메타데이터 헝크 — §E.2 M3 (d) + verdict §4 Gaps #2; 내용 라인 영향 0). 판정서: `.moai/reports/t1065/verdict.md`(로컬 gitignored).

### CHANGELOG 판정 (B12 discipline)

**결정: CHANGELOG 항목 없음(NO ENTRY).** 저장소 대면 변경이 없어 기록할 항목 자체가 없다 — 기억 저장소는 레포 외부이고 SPEC 아티팩트는 내부 프로세스 아티팩트다. B12 자가검증 3종 실측 (2026-09-22 22:31–22:34 KST, HEAD `6a5fab67c` 워크트리):

- **(a) 사전점검 grep** — 명령 + verbatim 출력:

```
$ grep -c 'SPEC-MEMORY-INDEX-FOLD-001' CHANGELOG.md
0
exit=1
```

  카운트 0 — 병렬 BATCH-SYNC 중복 항목 위험 없음. `exit=1`은 grep의 무매치 종료 코드.
- **(b) AC 수 일치** — 적용 대상 CHANGELOG 항목 부재로 불적용(N/A). 참고 실측: Tier S로 `acceptance.md` 미보유(AC는 spec.md §3 인라인)이므로 spec.md 대상으로 계산, 식별자는 정확히 5:

```
$ grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' .moai/specs/SPEC-MEMORY-INDEX-FOLD-001/spec.md | sort -u
AC-001
AC-002
AC-003
AC-004
AC-005
```

- **(c) 파일 경로 검증** — 기입할 CHANGELOG 항목이 없어 검증할 경로 클레임 없음(N/A).
- changelog_entry_position: N/A (no entry)

### 보류 항목 (후속 스테이지)

- **STAGE 2 — close 커밋**: 외부 sync-audit 판정 후 단일 close 커밋. spec.md frontmatter `in-progress → implemented → completed` 전이 + `updated:` 갱신을 4 artifacts에 원자 적용(§E.4 확정 동반). 현행 확인: spec.md:5 `status: in-progress`. 소유: manager-docs (spec-frontmatter-schema.md § Status Transition Ownership Matrix).
- **STAGE 3 — `sync_commit_sha` backfill**: `pending-backfill-sync` 자리에 close 커밋 실제 SHA 기입 (D3 SHA placeholder backfill exemption).

### Sync 부대 검증

- **MX Tag validation (sync 부단계)**: N/A — 저장소 코드 델타 0으로 @MX 대상 소스 파일 없음(§E.2 전 기간 repo 코드 편집 없음과 정합).
- **canary_compliance_check**: N/A — 본 SPEC은 자기 sync 테스트가 검증하는 선향 정책을 담지 않음(기억 색인 유지보수 카드).
