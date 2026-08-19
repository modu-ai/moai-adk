# t115 Review Verdict — MUST-FIX (잔존 누락: docs-site 홈 4로케일)

- reviewer: review lane session (release-v311 hub), 2026-08-17
- target: WT-t115 @ 99b1e5638 (본체, 17파일 +297/-212) + d1e7914a9 (증거)
- lens: 기본 4-관점 (docs-sync — 보안 비대상)
- 재측 환경: t115 트리에서 이번 리뷰 실행이 직접 측정

## Verdict

**MUST-FIX — 부분 반려.** 카드가 갱신한 17파일의 클레임은 전부 재측 일치하나, **카드의 본 목적(보드 6→5컬럼 서술 제거)이 docs-site 홈 `content/<locale>/_index.md` 4개 로케일 전부에 잔존**한다. evidence의 "잔존 스윕(전 표면) → 0건" 주장은 이번 재측으로 기각. 수정은 소규모(4파일 각 1문장절) — 반영 후 재검 경로 권장.

## MUST-FIX 상세

`git diff 47963ec0b..99b1e5638 -- docs-site/content/{en,ko,ja,zh}/_index.md` → **변경 0** (카드가 홈을 전혀 건드리지 않음). 그 결과 현재 트리에서:

| 로케일 | 위치 | 잔존 서술 (실측) |
|---|---|---|
| en | `content/en/_index.md:33` | "The board has **six columns**, `backlog → plan → run → review → sync → done`" |
| ko | `content/ko/_index.md:34` | "보드는 … `backlog → plan → run → review → sync → done` **여섯 칸**이고" |
| ja | `content/ja/_index.md:31` | "…の**6列**で" |
| zh | `content/zh/_index.md:31` | "看板有 … **六列**" |

공식 사이트(adk.mo.ai.kr) **홈페이지**가 6컬럼·review 세션 체계를 사용자에게 서술 — 카드가 지우려던 바로 그 낡은 아키텍처가 정문에 남는다. 홈은 4-로케일 패리티 의무 범위 안의 정본 문서다.

**수정**: 4로케일 `_index.md` 칸반 절 1문장씩 — 컬럼 나열을 5컬럼(`backlog → plan → sync` 3작업열 구조에 맞게)으로, review 노드 언급 제거, 개수 표현(six columns/여섯 칸/6列/六列) 정합. 수정 후 잔존 스윕(아래 R2 패턴)·hugo 재빌드·4-로케일 diff 확인이면 재검 충분.

## 클레임 재측 결과 (리드 지시 5종 + 부가)

| 클레임 | evidence | 이번 재측 | 판정 |
|---|---|---|---|
| README H2 12 ×4 | 12/12/12/12 | 12/12/12/12 | ✓ |
| README 펜스 42 ×4 | 42/42/42/42 | 42/42/42/42 | ✓ |
| README 줄수 755 ×4 | 755×4 | 755×4 | ✓ |
| 템플릿 미러 diff 0 | IDENTICAL | MIRROR-IDENTICAL, 13,575B 일치 | ✓ |
| mermaid LR/RL 0 | 0건 | 0건 | ✓ |
| URL 블랙리스트 0 | 0건 | 0건 | ✓ |
| hugo clean | 4288ms exit 0 | 4447ms exit 0, sitemap 572B | ✓ |
| PNG IHDR 3200×1760 | 3200×1760 | IHDR `0c80 06e0` = 3200×1760, 203,681B, docs-site 사본 cmp 동일 | ✓ |
| **잔존 스윕 0건** | 0건 | **_index 4로케일 6컬럼 잔존 (위 표)** | **✗ 기각** |

준거 코드 사실 재현 ✓: `internal/kanban/column.go:21-25` 5상수("review removed per t113"), `bootstrap.go:37` `CompanionRoles = {"plan","run","sync"}`("t97 retired the review role"), `bootstrap.go:15` t56 네이밍("carry no run id") — evidence 준거 측정 정확.

## 스윕 기각의 원인 추정 (R2)

워커 스윕 패턴은 `six-column`(하이픈)·`여섯 칸`·`6列`·`六列`·`review-<run-id>`로 제시됐으나 ko/ja/zh의 실제 표현(`여섯 칸`·`6列`·`六列`)이 패턴에 있었음에도 미검출 — 스윕 실행의 대상 경로에서 `_index.md`가 누락됐거나 grep 범위가 8개 갱신 페이지 주변으로 좁았을 가능성. en의 "six columns"(공백)는 하이픈 패턴이 본래 못 잡는 형태. 재검 시 패턴 교정 필요:

```
grep -rn 'six columns\|six-column\|여섯 칸\|여섯칸\|6列\|六列\|→ review →' \
  README*.md docs-site/content/**/*.md .claude/rules/moai/workflow/kanban-dispatch*.md
```
(컬럼 나열 `→ review →` 패턴이 개수 표현과 독립적으로 잡는 2차 안전망)

## 4-관점 평가

| 관점 | 판정 | 근거 |
|---|---|---|
| Functionality | **MUST-FIX** | 갱신분 자체는 정확(아래 Craft/Consistency)이나 목표 달성이 부분 — 홈 4로케일 잔존 + 검증 클레임(스윕 0건) 기각 |
| Security | N/A (비대상) | docs-only, 신규 입력·비밀 경로 없음 |
| Craft | PASS | 재측 8종 전부 일치 — 패리티·미러·빌드·이미지 파이프라인(IHDR 2x 렌더·사본 동일) 충실. SVG 소스 신규 확보로 재생성 가능성 확보 |
| Consistency | PASS(갱신분 한정) | 4-로케일 동시 갱신 원칙 준수(8페이지), 준거 코드 사실에 정확히 결합, 미러·make build 완료. 단 홈 미갱신으로 사이트 전체 기준 패리티는 깨짐(MUST-FIX 본체) |

## Gaps

- Vercel 프리뷰 렌더링(그림·mermaid 실제 표시) — push 후 관측 (evidence와 동일 갭).
- ja/zh 문체 사람 검수 (evidence와 동일).
- 이번 재측은 스윑 재실행 외 전 항목 커버 — 재검 시 MUST-FIX 수정분 + R2 스윕이면 충분.

## 라이더 (재검 시 반영)

- **R1 = MUST-FIX 본체**: `_index.md` 4로케일 칸반 절 5컬럼 정합.
- **R2**: 잔존 스윕 패턴 교정(위 코드블록) — 재검 시 `_index` 포함 전 content 대상 실행.
- **R3(사소, 기존 유지)**: PNG 파일명 "five-sessions" 재해석(ja/zh 검수와 함께 후속 검토 여지).

---

# 2차 재견 판정 (재작업 a916c8137, 2026-08-18) — MUST-FIX 유지

## 2차 재측 결과

| 항목 | evidence 주장 | 이번 재측 | 판정 |
|---|---|---|---|
| 교정 스윕 전 대상 → 0건 | 0건 | 0건 (동일 패턴·동일 대상 재현) | ✓ |
| R1 _index 보드 문장 5컬럼 + review-in-sync-gate | 반영 | en:32 "five columns … Review is not a column: the sync gate absorbs the verdict" + ko/ja/zh 대응 확인 | ✓ |
| R2 web-console ×4 ChainRoles 정합 | 반영 | 4로케일:55 `lead → plan → run → sync`, `viewmodel_ops.go:46` ChainRoles 실측 대조 일치 | ✓ |
| R2 multi-llm ×4 3컴패니언 | 반영 | `plan -> run -> sync` 체인·"three companion" 확인 | ✓ |
| hugo 재빌드 clean | 4301ms | 3887ms, exit 0 | ✓ |

## 2차 MUST-FIX (신규 발견 — R1의 미완)

**`_index.md` ×4 히어로 소개 문장(en:19 / ko:20 / ja:17 / zh:17)에 "네 동반 세션 + review" 잔존** (실측):

- en:19 "…across **five terminals**: a lead session drives the chain while **four companion sessions** each own a single column — `plan`, `run`, `**review**`, `sync` — … no session carries **four phases'** worth…"
- ko:20 "터미널 **다섯 개** … **네 개의 동반 세션**이 `plan`·`run`·`**review**`·`sync` … **네 단계**치 이력"
- ja:17 "5つのターミナル … **4つの同伴セッション**が`plan`・`run`・`**review**`・`sync` … **4フェーズ**分"
- zh:17 "**5 个终端** … **4 个伴随会话**各自负责 `plan`、`run`、`**review**`、`sync` … **四个阶段**"

재작업이 보드 문장(:30-34)만 고치고 소개 문장(상단)은 놓침. 준거 사실(`bootstrap.go:37` CompanionRoles 3개)과 직접 모순 — 정정: **터미널 넷(리드+3)·세 동반 세션·plan/run/sync·세 단계치 이력**. 이 부류가 계속 새로 나오는 구조적 이유: 스윕 패턴이 컬럼 개수 표현만 잡고 **컴패니언 수·세션 수·단계 수 표현**을 안 잡음. 스윕에 추가:

```
grep -rEn 'four companion|five terminals|네 개의 동반|터미널 다섯|4つの同伴|5つのターミナル|4 个伴随|四个伴随|5 个终端|`review`' \
    docs-site/content/ README*.md
```
(마지막 `review` 백틱 패턴이 컴패니언 나열 잔존을 직접 잡는 2차 안전망)

## 허브 판단 (워커 이관 질의 — README:301 `plan → run → verify → sync`)

**판단: 범위 확장하여 이번 재작업에 포함할 것 (README ×4 각 1문장).** 실측 근거:

1. **verify 스테이지의 코드 원천 부재 확정** — `internal/` 전역에서 스테이지/단계로서의 verify 정의 0건. verify 실재물은 전부 별개 개념: `internal/verify` 스냅샷 스토어+CLI, web 콘솔의 스냅샷 통계 라벨, 칸반 record의 `VerifyRung`(검증 결과의 엄격도 라벨 PRIMARY/FALLBACK/DEGRADED — record.go:27-38, "verify stage"는 주석의 일반 명사 용법).
2. **`kanban_chain` 골 프리셋도 코드 부재** — `internal/`·`cmd/`·템플릿 전역 0건. goal 패키지에 preset 정의 자체가 없음. README:301의 "arms a `kanban_chain` goal preset"은 원천 없는 주장.
3. 현행 체계의 단계 사실은 `plan → run → sync`(CompanionRoles/체인)뿐 — multi-llm에서 동일 프리셋 표기를 이미 3단계로 고친 것과 정합.

수정: `plan → run → verify → sync` → `plan → run → sync`, `kanban_chain` 프리셋 주장은 실재 골 무장 서술로 정정 또는 제거. **Origin-Trail Chain 서술은 유지** — append-only JSONL 계보·하트비트는 `internal/kanban/record.go` 체계로 실재.

## 2차 판정

**MUST-FIX 유지** — 수정 2건: (1) `_index` ×4 소개 문장 컴패니언/터미널/단계 수 정합 + (2) README ×4:301 verify·`kanban_chain` 정정(위 허브 판단). 반영 후 스윕(교정 패턴 2종)·hugo·README 패리티(줄수 유지 시 755 기준 재확인) 재검이면 통합 승인 가능.

---

# 3차 재견 판정 (재작업 72cc1b220, 2026-08-18) — **PASS (통합 승인)**

## 3차 재측 결과 (전항 이번 실행 실측)

| 항목 | 주장 | 재측 | 판정 |
|---|---|---|---|
| 지정1 _index 히어로 소개 ×4 | 터미널 넷·세 동반·세 단계 | en:19 "four terminals… three companion… plan/run/sync… three phases" + ko:20·ja:17·zh:17 전부 정합, alt 텍스트까지 갱신 | ✓ |
| 지정2 README ×4:301 | plan→run→sync·프리셋 제거 | 4로케일 전부 "`plan → run → sync` … under the lead session's coordination", `kanban_chain`·verify 소거, Origin-Trail 유지 | ✓ |
| 스윕-1 (6컬럼·kanban_chain·verify→sync) 전 표면 | 0건 | 0건 (README×4 + content 전체 + 룰 2 + 미러 대상) | ✓ |
| 스윕-2 (컴패니언/터미널/단계 수) 전 표면 | 0건 | 0건 | ✓ |
| README 패리티 | H2 12 ×4·755줄 ×4 | 12/12/12/12, 755×4 | ✓ |
| hugo 재빌드 | clean | 3307ms exit 0 | ✓ |

## 허브 판단 — 범위 밖 자발 정정 3표면: **전건 수용**

판정 지정 범위 밖이나 **동일 결함 근원**(kanban_chain/verify 낡은 서술)의 잔존 제거이고 전부 실재 코드 근거로 교체됐음을 실측 확인:

- **(a) advanced/kanban-mode ×4 — 수용**: `kanban_chain` 프리셋 제거 → "launcher arms the kanban-mode environment (`MOAI_KANBAN` chain seed) … SessionStart notice announces the run id and companion launch commands" (:179, 4로케일 전부 MOAI_KANBAN 1건씩). 코드 근거 실측: `internal/config/envkeys.go:160` `EnvMoaiKanban = "MOAI_KANBAN"`. 원천 없는 주장을 실재 메커니즘(환경 시드+SessionStart 안내) 서술로 교체 — 올바른 정정 방향.
- **(b) workflow-commands/moai-goal ×4 — 수용**: :122 "bundles `/moai goal`'s infinite-duration goal into a `plan → run → sync` chain" — 세 단계화 정합, kanban_chain 잔존 없음.
- **(c) workflow-commands/moai-run ko/en — 수용**: en:538·ko:540 "세 페이즈 `plan → run → sync` 자동 체이닝" 정합. **ja/zh는 해당 문단 부재 주장도 실측 일치** — 두 파일에서 세 단계 체이닝 서술 0건(수정 불필요한 정당한 생략, 커밋 18파일 목록과도 정합: ja/zh moai-run 미포함).

## 3차 판정

**PASS — 통합 승인.** 2차 MUST-FIX 2건 반영 완료, 스윕 2종 0건, 패리티·빌드 유지, 범위 밖 자발 정정 3표면은 전부 수용(동일 근원 제거 + 코드 근거 교체). 잔여 위험은 1차 판정문과 동일(Vercel 프리뷰 실측·ja/zh 문체 사람 검수).
