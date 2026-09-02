---
id: SPEC-CODEMAPS-ACCURACY-001
title: "codemaps 유령 인용 정밀 수리 — internal/factory→kanban·internal/bodp·ListActive API + cited-path-existence 재발 방지 축"
version: "0.1.0"
status: draft
created: 2026-09-02
updated: 2026-09-02
author: manager-spec
priority: P2
phase: "v3.1.4 target"
module: ".moai/project/codemaps, internal/graph, internal/cli, internal/template/templates/.claude/skills/moai/workflows/codemaps.md"
lifecycle: spec-anchored
tags: "codemaps, phantom-citation, graph-check, kanban-rename, t304, accuracy, template-first"
tier: M
---

# SPEC: codemaps 6개 문서의 유령 패키지 인용 수리와 기계적 재발 방지 축

## HISTORY

- 0.1.0 (2026-09-02): plan-phase 최초 작성. 카드 t304 (Class B, Tier M). 트리 origin/develop `65196a5a7` (worktree t304, branch WT-codemaps-accuracy)에서 레인 사전 조사가 완료된 상태로 착수 — 조사 결과는 §1에 그대로 반영했다. 병합 순서 제약: t432 (WT-codemaps-refresh @ `7e1c4d94f`, 미통합)가 같은 6개 codemaps 파일을 재생성하므로 이 카드의 파일 편집은 t432 병합 **이후**에 착지한다 (§3 M0).

## 1. 문제 — 측정된 형태

카드 전제: "codemaps 6개 문서가 존재하지 않는 패키지 6개를 서술한다 — 재스탬프가 덮어 감추는 기존 부정확". 원칙(t432 REQ-CMR-008 계승): **신선도 게이트 초록은 정확성의 증명이 아니다** — 재스탬프는 신선도 축만 지우고 부정확을 남긴다. 이 카드는 사실을 바로잡지 재스탬프하지 않는다.

레인 사전 조사(origin/develop `65196a5a7` 및 t432 트리 양쪽에서 독립 측정, t432 전수검사 보고서 `.claude/worktrees/t432/.moai/reports/t432/codemaps-accuracy-verification.md` 참조 — 읽기 전용):

- `.moai/project/codemaps/*.md`의 인용 `(internal|pkg|cmd)/` 경로: develop 기준 유니크 86개(t432 트리 102개). (t432 보고서 자체 집계 "100"과의 차이는 추출 정규화 산물이지 트리 변화가 아니다 — 부재 집합은 양 트리에서 동일하다.) **부재 집합 8개 — 양 트리에서 동일**:
  `internal/design`, `internal/evaluator`, `internal/factory`, `internal/migrate`, `internal/research`, `internal/state`, `internal/bodp`, `cmd/moai/main`.

### 1.1 부재 8개 분류표 (완결)

| ID | 경로 | 분류 | 판정 | 조치 |
|----|------|------|------|------|
| P1 | `internal/design` | 이미 수정됨 | 2026-08-31 resync(`dd817c44c`)가 modules.md 93행 인근 blockquote "존재하지 않음" 경고 노트로 전환 — **정당한 부정 인용** | 없음 (보존, REQ-CMA-006) |
| P2 | `internal/migrate` | 이미 수정됨 | modules.md 147행 인근 blockquote 노트 (실제 패키지는 `internal/migration` 단수형) | 없음 (보존) |
| P3 | `internal/state` | 이미 수정됨 | modules.md 218행 인근 blockquote 노트 (실제는 `internal/session`) | 없음 (보존) |
| P4 | `internal/research` | 이미 수정됨 | modules.md 232행 인근 blockquote 노트 | 없음 (보존) |
| P5 | `internal/evaluator` | 이미 수정됨 | modules.md 258행 인근 blockquote 노트 (SPEC-CLEANUP-EVALUATOR-001로 제거된 TDD RED 스캐폴드) | 없음 (보존) |
| P6 | `internal/factory` | **개명 미반영 — 잔여 양성 유령** | modules.md 158–162행(develop 기준)이 실재 패키지처럼 서술. 실제로는 SPEC-KANBAN-RENAME-001로 `internal/kanban`에 개명 — `record.go`(validateSessionID, record.go:202), `revision.go`(SuppressStep0551, revision.go:167-174), `integration_lock.go`가 전부 `internal/kanban`에 존재. modules.md 자체의 진입점 인용 `internal/cli/factory.go`·`internal/cli/launcher_blockcap_infinite.go`는 실존 | 절 재작성: 제목 `### internal/kanban`, 본문을 Kanban 모드 서술로 갱신 (REQ-CMA-003) |
| P7 | `internal/bodp` | **제거 미반영 — 양성 유령(develop)** | 커밋 `5792fc755`(#1278, worktree surface redesign)로 트리에서 제거. develop modules.md 95–97행이 양성 절로 잔존. t432 재생성은 dependencies.md 185행 인근 부정 각주로 전환(미병합) | t432 병합 후 각주를 modules.md 기존 5개 경고 노트와 같은 blockquote 형식으로 정렬해 보존 (REQ-CMA-004, D3) |
| P8 | `cmd/moai/main` | 정규화 산물 — 결함 아님 | overview.md 161행 `cmd/moai/main() → cli.Execute()`는 함수 호출-연쇄 표기이며 실체 `cmd/moai/main.go`는 실존. t432 정리 규칙(후행 슬래시 제거 → `.go` 접미 복원 → `cmd/moai/main`→`cmd/moai/main.go`)이 이미 처리 | 새 검사에 t432 정리 규칙 재사용 (REQ-CMA-002) |

### 1.2 식별자 드리프트 — data-flow.md `ListActive`

t432 전수검사에서 유일한 진짜 식별자 미적중(보고서 §3.1 #21). data-flow.md 3개 지점(develop 기준): 197행 mermaid 노드 `E["ListActive"]`, 214행 흐름 단계 "PreToolUse: ListActive 쿼리", 351–359행 "Registry (Session)" 인터페이스 블록. 실제 API(`internal/session/registry.go`, 이번 실행으로 직독 검증):

```go
// 리시버 메서드 (Registry)
Register(sessionID, specID, phase string) error   // registry.go:169
Heartbeat(sessionID string) error                 // registry.go:215
Deregister(sessionID string) error                // registry.go:241
Query(optSpecID string) ([]Entry, error)          // registry.go:266
// 패키지 함수
QueryActiveWork(optSpecID string) ([]Entry, error) // registry.go:261
```

data-flow.md의 `ListActive(spec string) ([]Session, error)`, `Register(spec, branch string)`는 파라미터·반환 타입까지 전부 실제 API가 아니다 (`Session` 타입도 부재 — `Entry`). t432는 REQ-CMR-004에 따라 기록만 하고 수정하지 않았다 — **본 수정은 t304 소관이다** (REQ-CMA-005).

### 1.3 재발 원인 — 생성기 입력 (R1 조사 결과)

codemaps 생성기는 기계 코드 생성기가 아니라 **스킬 구동(LLM 저작)**이다. `.claude/skills/moai/workflows/codemaps.md`(로컬)와 `internal/template/templates/.claude/skills/moai/workflows/codemaps.md`(템플릿 정본, 양본 확인)에서 재발 경로를 특정했다:

1. **Phase 2가 기존 문서를 병합 입력으로 읽는다** — "Analysis inputs: Existing .moai/project/codemaps/ content (if --force not set, for incremental updates)". `--force` 없는 재생성은 낡은 문서의 유령 이름을 문맥으로 흡수해 같은 유령을 재유도한다. 카드 경고("단순 재생성은 자동 해결이 아닐 수 있다")의 실체가 이것이다.
2. **Phase 4 "Verify all referenced files and modules actually exist"는 LLM 판단 권고문이다** — 실행 가능한 검증 명령, 부정 인용 규약, 기계적 게이트가 전혀 없다. 이 결함이 3개 그래프 레이어(codemaps / mx-index / edges) 전부를 통과했다 — 어느 축도 인용 경로의 존재를 검사하지 않는다.

### 1.4 t432 파생 MINOR (F1)

t432 로컬 전용 증거 파일 `.claude/worktrees/t432/.moai/reports/t432/codemaps-accuracy-verification.md` §3.1 제목(194행)은 "26항목", 표는 27행(251행은 "27항목"). 해당 파일은 lane-7 동결 트리의 untracked 로컬 파일 — **t304는 거기 쓰지 않는다**. 정정은 t304 자체 증거에 기록하고 리드에게 보고한다 (REQ-CMA-009).

## 2. 요구사항 (GEARS)

> GEARS 키워드(When/While/Where/shall/shall not)는 프로토콜 토큰으로 영문을 유지한다.

- **REQ-CMA-001** (Ubiquitous): Every positive citation of a `(internal|pkg|cmd)/` source path in `.moai/project/codemaps/*.md` shall name a path that exists in the working tree — the only permitted absent citations are the enumerated negative citations of §1.1 (P1–P5 warning notes, the P7 bodp removal note, and normalization artifacts resolved by the §1.1 P8 rules).
- **REQ-CMA-002** (When): When `moai graph check` runs, the citations check shall report a cited-path-existence metric as a layer report row and exit 1 when a positive citation names an absent path — applying the t432 normalization rules (trailing-slash strip, `.go`-suffix restore, `cmd/moai/main` → `cmd/moai/main.go`) and exempting paths on blockquote (`>`-prefixed) lines as negative-context citations (D1/D2).
- **REQ-CMA-003** (When): When a cited package is found renamed (`internal/factory` → `internal/kanban`, §1.1 P6), the harness shall rewrite the modules.md section under the real package heading with corrected content — not delete the description — and every entry point the rewritten section cites shall be an existing path.
- **REQ-CMA-004** (When): When a cited package is found removed from the tree (`internal/bodp`, §1.1 P7), the codemaps doc shall carry a negative citation note in the same blockquote format as the five existing warning notes, recording the removal — the note itself is the correct end state, not debt.
- **REQ-CMA-005** (Ubiquitous): The data-flow.md session-registry description shall cite the real `internal/session` API — no `ListActive`, no `Session` return type; the interface block shall show `Register(sessionID, specID, phase string) error`, `Heartbeat(sessionID string) error`, `Deregister(sessionID string) error`, `Query(optSpecID string) ([]Entry, error)` as Registry receiver methods, with the package-level `QueryActiveWork` where the call-chain notation names the package function — at all three drifted sites (mermaid node, flow step, interface block).
- **REQ-CMA-006** (While): While this SPEC is in run phase, the five existing warning notes (P1–P5: `internal/design`, `internal/migrate`, `internal/state`, `internal/research`, `internal/evaluator`) shall remain present and unflagged — they are correct negative citations, and no fix under this SPEC may remove or reword them into positive claims.
- **REQ-CMA-007** (When): When the codemaps skill is amended, the template source of truth (`internal/template/templates/.claude/skills/moai/workflows/codemaps.md`) shall be edited FIRST and the local mirror (`.claude/skills/moai/workflows/codemaps.md`) in the same change — amending Phase 2 to stop treating existing codemaps content as an authority for package existence, and Phase 4 to name the runnable existence-verification command (`moai graph check` citations row) and the negative-citation convention — negative citations MUST use blockquote form (a form mandate for negative citations, not a blockquote-exclusivity claim) — followed by `make build`.
- **REQ-CMA-008** (When): When run-phase work begins, the lane shall absorb `origin/develop` (picking up the t432 codemaps regeneration) into the card worktree and re-measure the phantom inventory against the post-absorb tree — every §1.1 classification cell re-observed by its named command before any codemaps file is edited, with the re-measured coordinates recorded in progress.md (VCI §2 baseline-integrity attribution).
- **REQ-CMA-009** (While): While this SPEC is active, the t432 worktree (`.claude/worktrees/t432`) shall not be written to by this card — the F1 §3.1 count correction (26→27) is recorded in t304's own evidence under `.moai/reports/t304/` and reported to the lead, never patched in the frozen tree.
- **REQ-CMA-010** (Event-detected / shall not): When a regeneration or restamp would leave a §1.1 positive phantom in place, the harness shall not treat freshness-gate green (stamp reachability, described-source-diff = 0) as accuracy proof — this SPEC's completion is decided by REQ-CMA-001/002 observations, not by the freshness verdict (t432 REQ-CMR-008 principle).

## 3. 설계 결정 (D1–D4, 명시적)

- **D1 — 재발 방지 메커니즘: 기계적 축으로 추가한다 (채택).** 카드 요구 (3)은 "검토"였고, 검토 결과: 3개 기존 레이어가 이 결함을 모두 통과했으므로 기계적 존재 검사만이 잡는다. 형태: `internal/graph`의 `CheckFreshness`가 반환하는 `CheckResult.Layers`에 citations 검사 행을 추가(레이어명 `citations`, metric 토큰 `positive-cited-path-absence`, threshold 0 — 양성 유령 1개면 red)하고 `internal/cli/graph_check.go`의 출력·`Failed()` 경로가 그대로 소비한다. exit-code 계약 0/1/2는 불변. 스크립트 수준/문서만 되돌리는 대안은 기각 — 스크립트는 CI 게이트(`moai graph check`)에 연결되지 않고, 문서 권고는 §1.3이 입증한 대로 이미 실패했다. Tier 예산 근거: `internal/graph` + `internal/cli` + 테스트로 국지적이며 새 하위시스템이 아니다.
- **D2 — 부정 인용 판별자: blockquote 행 면제 (채택) 및 위험 기록.** 판별자: 해당 행이 `>` 접두로 시작하면 그 행의 경로는 부정 문맥으로 간주해 면제; 양성 주장은 제목(`###`)·비인용 본문에만 유효. 근거: 기존 5개 경고 노트와 t432의 bodp 각주가 전부 blockquote 형식이다. **위음성 위험**: blockquote 안에 양성 주장을 쓰면 검사를 우회한다. 완화: REQ-CMA-007이 스킬에 "부정 인용은 반드시 blockquote 표기를 사용한다" 규약을 새겨 재생성 산출물을 규약에 맞게 유지한다 — 이는 blockquote 독점 주장이 아니다: 코퍼스에는 비부정 blockquote 주석(modules.md:267·dependencies.md:149)이 존재하며, 오늘의 존재 검사는 blockquote 행을 면제 축으로만 읽으므로 이들 주석의 내용에는 영향을 받지 않는다. 위양성 아님 — 양성 유령은 확실히 red.
- **D3 — bodp 각주 처리: 보존 (채택).** 제거 기록은 정확한 역사이며 5개 노트 선례와 일치한다. t432가 dependencies.md에 둔 부정 각주를 modules.md 경고 노트 형식(blockquote)과 정렬해 유지한다. 삭제 기각 — 삭제는 "제거된 패키지를 문서가 한 번 서술했었다"는 사실까지 지운다.
- **D4 — known-5 무처치 확인 (채택).** §1.1 P1–P5는 2026-08-31 resync가 이미 올바르게 고친 부정 인용이다. 이 SPEC의 AC는 이 노트들이 잔존하고 올바름을 검증할 뿐(REQ-CMA-006), 편집 대상이 아니다.

## 4. 제약

- **병합 순서 [HARD]**: 이 카드의 codemaps 파일 편집은 t432(WT-codemaps-refresh) 병합 이후에만 착지한다 — 같은 6개 파일을 t432 재생성이 다시 쓰므로, 흡수 전 편집은 병합 충돌 또는 덮어쓰기로 소실된다. run-phase 착수 전 `origin/develop` 흡수 + §1.1 재측정이 선행된다 (REQ-CMA-008).
- **Template-First [HARD]**: 스킬 편집은 템플릿 정본 먼저, 같은 커밋에서 로컬 미러, 편집 후 `make build`.
- 이 카드는 **사실 수정**이 목적이다 — 재스탬프·전면 재생성·스탬프 도달 축(t291 소관, 직교)은 범위 밖.
- `.moai/project/codemaps/` 편집은 run-phase 소관 — plan-phase(본 문서)는 편집하지 않는다.
- 언어 규약: codemaps 문서 본문은 기존 문서의 언어(한국어)를 따른다; Go 코드·주석·커밋 메시지는 영문.
- 게이트 계약: `moai graph check`의 0/1/2 exit-code 의미와 기존 3레이어 보고 형식은 변경하지 않는다 — citations 행은 추가이며 기존 행의 의미를 바꾸지 않는다.

## 5. Tier 분류

**Tier M.** 근거: 산출물이 문서 편집(modules.md·data-flow.md·dependencies.md) + 스킬 양본 편집 + Go 구현(`internal/graph`·`internal/cli` + 테스트)로 3개 이상 파일·마일스톤 5개, 단 신규 하위시스템이나 ≥10 파일 규모가 아니므로 Tier L 요건(≥3 마일스톤 AND ≥10 파일) 미충족. 카드 표기 "Tier S~M"에서 Go 축 채택(D1)으로 M 확정.

## 6. Out of Scope

### Out of Scope — 재스탬프·신선도 축
- 스탬프 도달성(stamp-reachability) 축 — t291 소관, 이 카드와 직교
- codemaps 재생성 실행(`--force` 전면 재생성) 및 신선도 스탬프 갱신 — 이 카드는 사실 수정이며 재스탬프가 완료 조건이 아니다 (REQ-CMA-010)

### Out of Scope — 문서 전면 재작성
- codemaps 6개 문서의 구조 개편·신규 섹션 추가·스타일 통일 — §1.1·§1.2의 유령 인용과 API 드리프트만 수정한다
- overview.md·entry-points.md의 서술 개선(P8 정규화 대상 외) — 결함 아닌 표기는 그대로 둔다

### Out of Scope — 타 카드 소관
- t432 트리·t432 증거 파일 직접 수정 (REQ-CMA-009 — F1 정정은 t304 증거에 기록)
- `internal/session` API 자체 변경 — 문서가 코드를 따라간다
- gate.yaml 임계값 재조정 — citations 축 threshold 0은 코드 기본값으로 정의된다

### Out of Scope — 검사 범위 확장
- `.moai/project/*.md`(product/structure/tech)·`.moai/docs/` 등 codemaps 외 문서의 경로 존재 검사 — 이 카드의 검사는 `.moai/project/codemaps/*.md`에 한정한다
