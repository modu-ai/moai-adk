---
id: SPEC-WEB-WRITE-SAFETY-001
created: 2026-09-07
updated: 2026-09-07
---

# Plan — SPEC-WEB-WRITE-SAFETY-001 (v0.1.0)

Tier M · Class B (결함, 원인 미특정 — 조사-선결) · cycle_type: **tdd** (RED-first reproduction)

## §A Context

- **카드**: t517 — `moai web`이 Save 없이 추적 config 파일을 재기록하는 결함.
- **작업 위치**: 워크트리 `.claude/worktrees/t517` (브랜치 `WT-web-write-safety`, 베이스 `0b1e27877` = origin/develop). primary checkout 무접촉.
- **SPEC 산출물**: `.moai/specs/SPEC-WEB-WRITE-SAFETY-001/{spec,plan,acceptance,progress}.md` (Tier M 3-artifact + progress skeleton).
- **스코프 SSOT**: spec.md §3.1 (쓰기 시점·범위·충실도·파서 4축) + §3.2 경계(t509 codex 패널 제외, t510 쓰기 위치 제외).
- **근거 상태**: 리드 실측 관측(O1: 무저장 재기록 2건, O2: llm.yaml 미귀속 기록)을 ground truth로 취급하되, **본 세션 재측정은 아니다** — M1이 같은 재현을 본 카드에서 다시 관측해 채택한다(carry-over 방지).

### §A.1 전달 근거의 검증 상태 (리드 지시 — 확인/미확인 구분)

| 리드가 전달한 근거 | 검증 상태 |
|---|---|
| `handleSave` 8단계 순차 쓰기·롤백 없음, 단일 파일 temp+rename | **직접 확인** — handlers.go:350-558 판독. 쓰기 호출은 9개(ledger 1건 advisory 포함); "8단계"는 리드 카운트 기준 |
| git-strategy dirty-gate 존재 (`manager.go:206-221`) | **직접 확인** — 판독. dirty 플래그 리셋(`:230`)까지 확인 |
| llm.yaml typed 경로 무조건 재기록 (`manager.go:223-226`) | **직접 확인** — 판독. gate 없음 |
| `r.PostFormValue` 첫 값 반환 (`schemaform.go:323,329,339`) | **직접 확인** — 320-352 판독 (bool/int/float/text 분기 전부) |
| feedback.yaml이 "seam" 섹션이라는 것 | **직접 확인** — `sectionroute.go:99` `"feedback": RouteSeam` + `projectconfig.go:160-164` seam/typed 이원 구조 |
| `ConfigManager.Save()` 저장 목록에 feedback.yaml **없음** | **직접 확인** (리드 미언급 추가 발견) — feedback.yaml 기록 주체는 `Save()` 밖 별도 경로 |
| 서버 기동 경로의 config 쓰기 부재 | **직접 확인** — `server.go:94-141` 판독 + `internal/cli/web.go` grep 무매치 |
| 관측 사실 자체 (파일 2건 변경·mtime·복구 경과) | **전달 근거(미재측정)** — M1에서 본 카드가 재관측 |
| primary checkout llm.yaml 오늘 기록 | **전달 근거(미귀속)** — M2의 1급 귀속 대상 |

## §B Known Issues

- **B1 (쓰기 경로 미특정 — M-a)**: handleSave는 POST 게이트이고 기동 경로는 쓰기가 없다(C7). 그런데 썼다. spec.md §1.4의 4가지 후보 기제(htmx 자동 POST / lazy GET 쓰기 / CLI 래퍼 접촉 / 외부 동시 작성자) 중 어느 것도 배제되지 않았다.
- **B2 (gate 우회 기제 미특정 — M-b)**: dirty-gate가 있는데 git-strategy.yaml이 바뀌었다. 저장 흐름에서 `SetSection(git_strategy)`이 조용히 불리는지, gate 밖 경로인지 M3가 특정한다.
- **B3 (feedback.yaml 기록 주체)**: `Save()` 저장 목록에 없다(C1). yamlpatch seam 경로(`applySchemaEdits`)가 추정 주체이며, seam 경로의 빈 줄 보존 실패가 M-b의 두 번째 축이다.
- **B4 (llm.yaml 미귀속 기록)**: primary checkout llm.yaml이 오늘 쓰였다(O2). C3(무조건 재기록)이 그럴듯한 기제이나 측정된 귀속이 아니다. M2에 포함한다.
- **B5 (이진 최신성)**: 관측에 쓰인 `moai` 바이너리와 run-phase 검증 바이너리의 빌드 시점이 다를 수 있다. 수리 후 검증은 반드시 재빌드 후 수행한다(`verification-completeness.md` §1.3).
- **B6 (병렬 세션)**: 워크트리·primary 주변에 다른 레인이 활성일 수 있다. M1 재현 중 외부 쓰기 혼입을 막기 위해 격리 트리에서 수행하고, 관측 시각대의 다른 세션 활동을 기록한다.

## §C Pre-flight

1. **앵커 재검증**: spec.md §1.3의 file:line 앵커는 트리 `0b1e27877` 실측값. run-phase 착수 시 content-token 기준으로 재검증한다(라인 번호 드리프트 무시, 코드 토큰 기준).
2. **바이너리 빌드**: 검증용 `moai`를 본 카드 브랜치에서 빌드하고 SHA를 기록한다. `moai version` 출력 + 빌드 커밋을 M1 증거에 귀속.
3. **격리 트리 준비**: fresh worktree(또는 전용 fixture 트리)를 만들고 `.moai/config/`가 존재하는 상태로 세팅한다. primary checkout은 절대 대상이 아니다.
4. **스냅샷 복사**: 재현 전 `.moai/config/sections/` 전체를 스냅샷 디렉터리로 `cp`. 이후 모든 판정은 스냅샷과의 diff로 한다.
5. **동시 세션 기록**: 재현 시각대에 이 머신에서 활성인 다른 moai 세션 유무를 기록한다(B6 혼입 배제 근거).

## §D Constraints

- **`git restore` 금지**: 본 카드에서 `git restore`를 도구로 쓰지 않는다(관측 시 1회 복구 목적으로 쓰인 것을 일반화하지 않는다). 복구가 필요하면 스냅샷 복사본으로 되돌리고, 워크트리는 어차피 폐기 가능한 격리 트리다.
- **primary checkout 무접촉**: 재현·검증·커밋 전부 워크트리 안에서. 종료 시 primary `git status`가 시작 시와 동일함을 비교로 증명한다.
- **경계**: t509(codex 패널)·t510(쓰기 위치) 영역에 수리·리팩터를 넣지 않는다. 발견하면 리드에 회신.
- **측정 귀속**: 모든 판정은 커맨드 + 출력 + 트리 SHA. 전달 수치(O1/O2)를 그대로 판정 근거로 쓰지 않는다.

## §E Self-Verification

run-phase 종료 시 다음을 보여야 한다:

- E1: M1 재현 증거 — 스냅샷 diff(RED) + 커맨드/출력/exit code/트리 SHA 4요소
- E2: M-a/M-b 측정 보고서 — 후보 기제 4가지 각각의 소거 또는 채택 근거
- E3: 수리 후 같은 재현 절차 GREEN + 비편집 섹션 diff 없음 + git-strategy 양성 통제(SetSection 주입 시 재기록 관측)
- E4: 부재-가드 AC 전부의 RED-first 셀 + 뮤턴트 포착 기록
- E5: 변경 패키지 테스트 통과 출력 (`go test ./internal/web/... ./internal/config/...` — 전체 스위트 아님, CI 몫)
- E6: primary checkout 무접촉 증명

## §F Milestones

> 순서 원칙: **재현+측정 먼저, 수리는 측정 뒤, 회귀 테스트는 수리 뒤**. 결정-가역성 순으로 M4 내부에서는 바뀔 가능성이 큰 쓰기-게이트 설계 결정을 앞세우고 기계적 파서 수정을 뒤로 보낸다.

### M1 — RED-first 실물 재현 (Priority High)

- 격리 트리에서 `moai web` 기동 → Save 제출 없이 탐색/대기 → 종료 → 스냅샷과 diff.
- **3단계 분해로 판별력 확보**: (a) 기동만(브라우저 미접속), (b) 페이지 렌더(GET), (c) 탐색·폴링 유지. 어느 단계에서 쓰기가 발생하는지 구분 기록 — M-a의 1차 판별 증거가 된다.
- 산출: RED-now 증거 4요소(커맨드·출력·exit code·트리 SHA)를 acceptance.md 각 부재-가드 AC의 RED 셀로 귀속.
- 완료 판정: 무저장 상태에서 `.moai/config/**` 변경이 **본 카드 재현에서 직접 관측**됨(O1 단순 인용 아님).

### M2 — 첫 측정 A: 무저장 쓰기 경로 귀속 (M-a 게이트, Priority High)

- spec.md §1.4의 4가지 후보 기제를 각각 검증/소거:
  1. htmx 자동 POST — 렌더된 HTML에서 `hx-post`/`hx-trigger` 자동 제출 패턴 조사 + 요청 로그 관측
  2. lazy GET 쓰기 — GET 라우트 핸들러 코드 추적 + 라우트 순회 요청 후 diff
  3. CLI 래퍼 기동 접촉 — `moai web` 기동 경로(strace류 관측 또는 코드 추적)로 config 파일 개방/기록 확인
  4. 외부 동시 작성자 — 재현 중 다른 세션 활동 기록과 대조, 배제 불가하면 blocker report
- **llm.yaml 귀속(O2) 포함**: primary checkout llm.yaml 기록의 작성 시각대와 사용자 활동(콘솔 Save 유무)을 대조해 C3 기제와 일치하는지 판정. primary에는 쓰지 않고 기록·관측만.
- **6종 판별 증거(D2)**: 무저장 재현 결과 `Save()` 6종 섹션(user/language/quality/git-convention/git-strategy/llm) 중 어느 파일이 **내용 변경**됐는지 판별해 기록한다. O1은 git-strategy·feedback 2종만 내용 변경임을 시사 — 무조건 5종이 함께 기록됐는가(내용 동일 round-trip 여부 포함)가 "전체 `Save()` 통과 vs 부분 경로만 통과"를 가르는 교차 판별점이며, 이 판별 없이는 M-a 귀속이 과소 특정될 수 있다.
- 산출: 쓰기 경로 특정 보고서(커맨드+출력+트리 SHA 귀속). M4는 이 결론을 인용해서만 설계한다.
- **[HARD] 이 측정이 완료되기 전까지 어떤 수리 코드도 작성하지 않는다.**

### M3 — 첫 측정 B: gate 우회 + seam 충실도 원인 (M-b 게이트, Priority High)

- git-strategy.yaml: 저장 흐름 전체를 추적해 `SetSection(git_strategy)` 호출 유무·`gitStrategyDirty` 플래그 변화를 특정. 가능하면 단일 유닛 테스트로 우회 기제를 환원해 재현.
- feedback.yaml: "seam"의 정의를 코드에서 정리(`sectionroute.go` RouteSeam 계약)하고, yamlpatch 경로가 빈 줄을 삭제하는 지점을 특정. golden-file round-trip 테스트로 환원.
- 산출: 두 원인 진술 + (가능하면) 결함을 환원한 유닛 테스트 — 이것이 M4 수리의 RED 테스트가 된다.
- **[HARD] 이 측정이 완료되기 전까지 어떤 수리 코드도 작성하지 않는다.**

### M4 — 수리 (Priority High, M2+M3 완료 후에만 착수)

- **쓰기 시점 게이트** (가장 바뀔 가능성이 큰 설계 결정 — 먼저 검토): M-a로 특정된 경로가 명시적 저장 동작 없이는 `.moai/config/**`에 쓰지 않도록 차단. 구현 형태(게이트 위치·방식)는 M-a/M-b 결론에 따라 확정 — 본 plan은 형태를 미리 못박지 않는다.
- **쓰기 범위 최소화**: 값 불변 섹션 재기록 제거(REQ-WWS-003). 무조건 재기록 exemplar는 llm.yaml 단독이 아니라 **5섹션 전부** — user.yaml(`manager.go:187`), language.yaml(`:192`), quality.yaml(`:197`), git-convention.yaml(`:202`), llm.yaml(`:224`); git-strategy만 gated 6번째다(C2).
- **포맷 충실도**: M-b로 특정된 seam 빈 줄 삭제 지점 수리 + golden round-trip으로 키 순서·주석·unknown key 보존 확인.
- **파서 중복 인지**: `schemaform.go`의 중복-값 무지 수리(거부 또는 문서화된 규칙). t509 패널에는 접촉하지 않는다.
- 완료 판정: M1과 동일한 재현 절차가 GREEN, 비편집 섹션 diff 없음, git-strategy 게이트 계약 회복.

### M5 — 회귀 가드 + 채택 (Priority High)

- 부재-가드 AC 전부를 2-cell로 채택: RED-now 셀(M1 증거, 트리 SHA 고정) + green path 셀(M4가 뒤집음).
- 뮤턴트 검증: 각 가드에 결함 재도입 변형(예: startup에 Save 주입, gate 우회 복원, 첫 값 채택 복원)을 적용해 RED 확인. **못 잡은 뮤턴트도 기록에 남긴다**(가드 경계의 문서화).
- git-strategy **양성 통제**: SetSection 주입 시 재기록이 관측됨을 확인 — gate가 살아 있음의 반대 방향 증거.
- 변경 패키지 테스트 통과 + 수리 후 재빌드 확인(B5).

## §G Anti-Patterns

- `git restore`를 복구 수단으로 습관화 — 스냅샷 복사본과 diff로 판정한다.
- 결함이 존재하는 데 가드가 GREEN인 채택(공허한 초록) — RED-first 없는 부재-가드는 채택이 아니다.
- 전달된 관측(O1/O2)을 본 카드 측정인 것처럼 인용 — M1에서 재관측한다.
- 셀렉터 0매치/빈 스윕을 GREEN으로 판독 — 스윕 대상 수를 먼저 확정한다.
- grep으로 "수리 착지" 판정 — pre-fix 형태 실패 + post-fix 형태 통과를 같은 트리에서 관측한다.
- primary checkout에서의 재현·검증 — 격리 트리만.
- t509/t510 영역 침범 — 발견 시 회신, 수리하지 않는다.

## §H Cross-References

- spec.md §1.4 (M-a 후보 기제 4가지) · acceptance.md §D (AC 매트릭스 + RED-first/mutant 셀 구조)
- `verification-completeness.md` §1-§2 (관측-실패 완결 축 + 2-cell 채택) · `verification-claim-integrity.md` §2 (baseline 귀속)
- SPEC-GITSTRATEGY-SAVE-ISOLATION-001 · SPEC-WEB-CONSOLE-011 · SPEC-WEB-CONSOLE-010 · SPEC-FEEDBACK-AUTO-SUBMIT-001
- t509 (codex 패널 경계) · t510 (쓰기 위치 축 경계)

## [NEEDS CLARIFICATION] — 없음

조사 항목(M-a/M-b)은 전부 run-phase 측정으로 소관이 명확하다. 유일한 분기: M2가 "외부 동시 작성자" 결론에 도달하면 수리 대상이 본 카드 코드가 아니므로 blocker report로 리드에 회신한다(M2 완료 판정에 포함).
