# SPEC-V3R6-DOCS-PREREWRITE-FLAG-001 — Implementation Plan

## A. Context (핀, 출처, 방법)

### A.1 Measurement pin

- 트리: `.moai/worktrees/t1726`, 브랜치 `WT-factory-run-clear`, 플랜 작성 시점 HEAD `74ad197380` (로컬 main 팁과 동일 확인).
- 모든 계수·RED-now 관측은 이 트리, 이 실행에서 수행했다(verification-claim-integrity §2 귀속). 원시 기록: acceptance.md 증거 장부 E-01~E-04, 레인 전제 기록 `.moai/reports/t1726/progress.md`.
- 카드: t1726 [FLAG] 플래그 인용 `--factory-run`·`--clear-policy` 정리. Class C, docs-only. 부모 인덱스 SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001의 11개 자식 중 FLAG.

### A.2 Q18 인용 분류 (계획이 적용하는 기본 결정)

Q18(제품 수준)은 decision-index.md의 열린 행이다. 이 계획은 기본값 (a) — **대체된 소비 인용만 정리, 살아 있는 인용 보존** — 을 적용해 작성됐고, 그 적용은 다음 분류로 구체화된다(트리 74ad197380에서 직접 확인한 코드 근거 포함):

- `--factory-run`: LIVE. 파서 `internal/cli/factory.go:153-163`(RunID 선택자 :126), 재개 검증 `internal/cli/lane_resume.go:61`. 12개 표면의 모든 `--factory-run` 인용은 선택자·합류·이어받기라는 살아 있는 동작을 서술한다 → 전부 유지.
- `--clear-policy`: LIVE(레인 전용 선택, REQ-SD-020 — `internal/cli/factory.go:75`, `:127`, `:836`, `lane_resume.go:62`). 그러나 relaunch 값의 **소비**는 카드 t1554로 대체됐다(`internal/cli/factory_card.go:2484-2485`, `factoryRelaunchSupersededNote`: 카드 소비는 통합 `moai todo --auto` 엔진으로, 런처는 레인 세션 하나를 띄운다).
- 따라서 STALE은 세 부류이다(iter-1 D2 원문 대조 판정 반영): **STALE-1** README 4종 74행의 "카드를 순회하는 `--clear-policy relaunch` 루프" 인용, **STALE-2** factory-mode 4 로캘 77행 `relaunch` 셀의 카드별 새 세션 기동 서술, **STALE-3** 레인 `/clear` 안내 서술 — factory-mode 77행 `clear-each` 셀("카드마다 `/clear`를 요청")·134행("그 레인을 `/clear`합니다"), README 4종 87행("the lane asks for a `/clear`"). STALE-3 판정 근거: `factoryPrintClearPolicyLine`(`internal/cli/factory_card.go:2527-2551`)은 clear-each에 세션 안 카드 전환(`ExitWorktree(keep)` → `moai factory next` → `EnterWorktree`)을 출력하고 godoc이 "No value instructs a session clear"로 끝나며, `TestSD_AC020_ClearPolicies`(`internal/cli/factory_m6_test.go:74-76`)가 complete 출력의 `/clear` 지시를 금한다. launchers 37행의 값 나열과 `clear-when-full`의 임계치 조건 서술은 LIVE → 유지(AC020은 complete 출력의 clear 지시를 금할 뿐 문서의 값 나열을 금하지 않는다).
- Q13(운영자 결정 d-20261019T015908Z-220a): 교체 대상이 존재하면 치환, 없으면 삭제. 세 STALE 부류 모두 현재 교체 대상이 존재하므로 기본 처분은 **치환**이다 — STALE-1·2는 단일 레인 세션이 `moai todo --auto` 큐를 소비한다는 현재 소비 서술로, STALE-3는 세션 안 카드 전환(`ExitWorktree(keep)` → 임대 → `EnterWorktree`; 이전 카드 대화 맥락은 비권위) 서술로 치환한다. 치환 문장은 `--clear-policy` 플래그를 계속 이름 붙인다.
- Q17(같은 결정): 4 로캘을 같은 변경 세트로 함께 수정. 로캘별 분산 편집 금지, 단일 커밋 허용.
- Q18이 운영자 판정으로 기본값에서 벗어나면(예: (b) 전면 제거) M1의 처분표가 판정에 맞춰 다시 쓰이고 M2 분량만 늘어난다 — 인벤토리와 검증 명령은 재사용 가능하다.

### A.3 Method limits

- 문서 전용: Go 빌드·테스트·커버리지는 이 SPEC의 검증 대상이 아니다(변경이 Go 코드에 닿지 않음을 AC-DPF-006이 음(-)으로 보장).
- 계측은 grep 라인 카운트와 원문 인용이 전부다. 문장 단위 의미 판단(STALE-2 셀의 어느 절이 대체된 부분인가)은 M1 처분표에서 줄별로 기록하고 M3에서 diff 대조로 검증한다.

## B. Known Issues (docs-only 필터 — manager-develop 템플릿 B1~B12 중 해당)

- **B2 계열 — 대체 정책과의 충돌**: 정리 대상 인용이 가리키는 소비 경로는 t1554가 대체했다. 문서가 옛 소비를 현재로 서술하는 것이 이 카드의 결함 자체다. 코드 쪽 노트(`factoryRelaunchSupersededNote`)는 유지되는 것이 전제며, 이 SPEC은 그 문구를 바꾸지 않는다.
- **B6 — 표제 관례**: spec.md의 Out of Scope는 h2 아래 `### Out of Scope — <topic>` h3 소절과 `-` 항목을 갖춰야 한다(MissingExclusions lint). 작성 시 반영했다.
- **B8 — 작업 트리 위생**: 커밋은 명시적 경로스펙으로만. `.moai/state/`, `.moai/harness/` 등 런타임 관리 파일을 만지지 않는다.
- **B10 — 범위 규율 (PRESERVE)**: 12개 표면의 플래그 인용 외 줄은 편집 금지 — `factory-mode.md` ×4의 SWEEPA 담당 줄, README 4종의 README 자식 담당 줄(부재 명령 인용 등). 파일 공유, 줄 불공유(D3).
- **B11 — 블로커 프로토콜**: 런 페이즈에서 Q18 판정이 도착해 기본값과 다르면 재위임 없이 처분표만 갱신 가능한지 판단하고, REQ 텍스트 변경이 필요하면 블로커 보고로 복귀한다.
- B1(크로스 플랫폼 빌드)·B3(서브에이전트 경계)·B4(프론트매터)·B5(CI 3단)·B7(observer)·B9(직접 push)·B12(CHANGELOG): 이 문서 전용 SPEC에 해당 없거나 싱크 페이즈 소관 — 해당 없음으로 필터.

## C. Pre-flight

런 페이즈 시작 전 manager-develop 실행:

```bash
git rev-parse --short HEAD          # 74ad197380 이후 흡수 커밋 여부 확인
git status --short                  # 깨끗한 트리 확인
grep -c -- "--clear-policy relaunch" README.md README.ko.md README.ja.md README.zh.md   # RED-now 재확인 (기대: 각 1)
grep -c -- "--factory-run" docs-site/content/ko/advanced/factory-mode.md README.md      # 보존 baseline 재확인 (기대: 각 2)
```

HEAD가 74ad197380에서 움직였으면(흡수 병합) RED-now 셀을 새 트리로 재측정·재핀한다(verification-completeness §4).

## D. Constraints (PRESERVE + 금지)

- **PRESERVE**: `internal/`, `cmd/` 전체. 12개 표면의 플래그 인용 외 전체 줄(특히 factory-mode.md의 SWEEPA 담당 줄, README의 README 자식 담당 줄). `internal/cli/factory_card.go`의 `factoryRelaunchSupersededNote` 문구.
- **금지**: `git add -A`/`git add .` 대신 명시 경로스펙. force-push. `--no-verify`. 런 페이즈에서 로캘별 분산 커밋(Q17 위반). 시간 예측 문구.
- **변경 허용 파일(정확히 12개)**: `docs-site/content/ko|en|ja|zh/advanced/factory-mode.md`, `docs-site/content/ko|en|ja|zh/cli-reference/launchers.md`, `README.md`, `README.ko.md`, `README.ja.md`, `README.zh.md`.
- 편집은 인용 줄(및 그 문장이 속한 표 셀/문장)에 한정한다. 표 구조와 문서의 다른 절을 재배치하지 않는다.

### D.1 Risks

| 위험 | 영향 | 완화 |
|---|---|---|
| Q18 운영자 판정이 기본값 (a)가 아님 | M2 분량 변화 (전면 제거 시 인용 24줄 전부) | M1 처분표를 판정 파라미터로 작성 — 분류·검증 명령 재사용, REQ 텍스트는 (a)/(b) 공통 |
| 파일 공유 카드(SWEEPA, README 자식)와의 병합 충돌 | 같은 파일 다른 줄 — 텍스트 충돌 가능성 낮지만 0 아님 | 부모 Appendix A 착지 순서 권장(README 먼저); 줄 범위 편집 엄수 |
| 치환 문장이 `moai todo --auto` 소비를 부정확히 서술 | 새로운 부정확 인용 유입 | M3에서 치환 문장을 `factoryRelaunchSupersededNote` 원문과 대조 |
| 부분 로캘 편집 (패리티 깨짐) | Q17 위반 + 독자 간 정보 어긋남 | AC-DPF-005 카운트 대칭 검사 + 단일 변경 세트 규칙 |
| HEAD 이동으로 RED-now 핀 실효 | 계측 귀속 상실 | Pre-flight 재측정·재핀 절차 (§C) |

## E. Self-Verification deliverables (manager-develop §E 보고 형식)

- **E1** AC PASS/FAIL 매트릭스 (AC-DPF-001~007, 각 행에 명령+실측 출력).
- **E2** grep 스윕 결과: STALE 문구 0 확인(AC-DPF-001·002·007 명령), 보존 카운트(AC-DPF-003·004).
- **E3** 4 로캘 패리티 카운트 대칭 (AC-DPF-005).
- **E4** 음(-) 검증(다중 상태, iter-1 D3 강화): 두 단일 명령으로 네 상태를 전부 열거한다 — `git diff --name-only <base>..HEAD -- internal/ cmd/`(커밋된 변경)와 `git status --porcelain -- internal/ cmd/`(스테이지·작업 트리 수정·언트래킹) 모두 빈 출력 (AC-DPF-006). `<base>`는 판정 시점의 감사 대상 커밋으로 재핀하고 그 SHA를 증거 행에 기록한다.
- **E5** diff 범위 확인(허용 목록 대조, iter-2 D8 확대): 변경 경로 전체 집합(커밋+스테이지+작업 트리+언트래킹 — E4의 두 명령, 전체 `git status --porcelain`, 그리고 경로스펙 없는 `git diff --name-only <base>..HEAD` — internal/·cmd/ 밖의 커밋된 변경도 보인다)이 §D의 12개 허용 파일 안에 있다. 목록 밖 경로 하나라도 있으면 FAIL.
- **E6** 커밋 상태: 단일 런 페이즈 커밋(Q17 허용), Conventional Commits + `Authored-By-Agent` 트레일러.
- Go 커버리지/빌드/린트 항목은 문서 전용이라 해당 없음 — E 항목에서 "N/A (docs-only)"로 명시한다.

## F. Milestones (의사결정 가역성 순 — 판단이 먼저, 기계적 편집은 뒤)

### F.1 M1 — Q18 인용 분류 확정 + 파일/줄 인벤토리 (Priority High)

- **진입 선결 조건(Q18, iter-1 D1)**: Q18 운영자 판정이 도착해 있어야 M1(런 페이즈)을 시작한다. 판정은 리더가 운영자에게 라우팅한다(decision-index Q18 행 — FOUNDER·product-level·`Operator verdict:` 공석, 이 행이 채워질 때까지 자율 Kickoff는 막힌다). planner 기본값 (a)는 이 계획의 잠정 입력일 뿐이며 자율 진입 승인이 아니므로, 판정 없이 기본값으로 진행하는 경로는 없다.
- 판정이 (a)이면 아래 분류가 그대로 처분표가 되고, (a)와 다르면(예: (b) 전면 제거) 처분표를 판정에 맞춰 갱신한다(행은 판정으로 닫힌다).
- 현재 HEAD에서 §A.2 분류표를 재측정해 일치 확인(acceptance.md E-01~E-04 재실행).
- Q13 처분표 작성: 줄별로 REPLACE/DELETE와, REPLACE의 경우 현재 메커니즘 문구(STALE-1·2: 단일 레인 세션이 `moai todo --auto` 큐를 소비 / STALE-3: 세션 안 카드 전환)를 기록. 기본값 (a)의 REPLACE 집합: README 74 ×4(STALE-1), factory-mode 77 ×4(STALE-2 relaunch 셀 + STALE-3 clear-each 셀 — 같은 줄), factory-mode 134 ×4, README 87 ×4(STALE-3) — 실질 16줄.
- 산출물: 처분표(이 SPEC 디렉터리 내 기록 또는 커밋 메시지 본문).

### F.2 M2 — 4 로캘 + README 4종 편집 (Priority High)

- **하나의 변경 세트**로 12개 파일을 함께 편집한다(Q17; 단일 커밋 허용).
- README 4종 74행(STALE-1): `--clear-policy relaunch` 루프 인용을 현재 메커니즘 서술로 치환(Q13). `moai factory relaunch`(일회성 명령) 본문은 유지하고 대비절만 고친다. 로캘 간 동일 의미.
- factory-mode 4 로캘 77행(STALE-2+3): `relaunch` 셀의 "카드마다 새 세션 기동" 서술을 현재 소비(엔진이 레인 세션 하나를 띄우고 그 세션이 큐를 소비)로, `clear-each` 셀의 "카드마다 `/clear`를 요청" 서술을 현재 동작(세션 안 카드 전환 — clear 지시 없음)으로 치환. `--clear-policy <값>` 행 자체와 `clear-when-full` 셀(임계치 조건)은 유지.
- README 4종 87행·factory-mode 4 로캘 134행(STALE-3): 레인 `/clear` 안내를 현재 카드 전환 서술로 치환(Q13). 치환 문장은 `--clear-policy`를 계속 이름 붙인다(문서군별 플래그 카운트 유지).
- 그 외 줄 무변경. 74행의 README 본문 중 `moai factory relaunch` 설명, 65/67/36/37/72/99행 전부 무변경.
- 이 변경 세트가 AC-DPF-007(iter-2 D7, release-blocking)을 뒤집는다: 편집 후 E-05의 12개 STALE-3 문구가 0(출력 없음, exit 1)이 되고, 치환 줄은 `--clear-policy` 인용을 유지한다(AC-DPF-004의 E-04 카운트). factory-mode 100행 mermaid 라벨은 범위표 밖이라 무변경이다.

### F.3 M3 — 검증 (Priority High)

- AC-DPF-001~007 실행(acceptance.md의 명령과 기대 출력 그대로).
- 치환 문장 대조: 편집된 문장을 `internal/cli/factory_card.go` `factoryRelaunchSupersededNote` 취지와 대조.
- 선택: Hugo 툴체인이 환경에 있으면 문서 빌드 경고 확인(oss-docs 검증 레시피). 없으면 Gaps에 명시(침묵 생략 금지).

## G. Anti-Patterns

- 살아 있는 인용(`--factory-run` 전부, `--clear-policy` 선택 서술)까지 지우는 과잉 정리 — REQ-DPF-005 위반 변형.
- STALE 자리에 대체된 순환을 새 문구로 되살리는 치환 — Out of Scope.
- help-text 추가를 "문서 작업의 일부"로 포장해 `internal/`을 건드리는 것 — 별도 SPEC.
- 한 로캘만 먼저 고치는 분산 편집 — Q17 위반.
- 검증을 "grep 했을 것이다"로 보고하는 것 — 모든 AC 보고는 명령+실측 출력(verification-claim-integrity §1·§2).

## H. Cross-References

- 부모 인덱스: `.moai/specs/SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001/` (decision-index Q13·Q17·Q18, Appendix A FLAG 행).
- 형제: GATE(card t1717, Q18 입력 — SPEC-V3R6-DOCS-PREREWRITE-GATE-001은 계획 전, `.moai/specs/` 미존재, 그 플랜 페이즈가 이 참조를 소유한다), SWEEPA(t1721, factory-mode.md 공유), README 자식(t1720, README 공유).
- 코드 근거: `internal/cli/factory.go`, `internal/cli/lane_resume.go`, `internal/cli/factory_card.go`(`factoryRelaunchSupersededNote`).
- 전제 기록: `.moai/reports/t1726/progress.md`.
- 규칙: `.claude/rules/moai/development/verification-completeness.md`(두 칸 채택), `.claude/rules/moai/core/verification-claim-integrity.md`, `.moai/docs/docs-site-i18n-rules.md`(참고 사실 — Q17 운영자 결정이 로캘 의무의 근거다).
