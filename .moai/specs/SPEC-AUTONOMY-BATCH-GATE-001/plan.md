# SPEC-AUTONOMY-BATCH-GATE-001 — Plan

> 이 문서는 구현 계획이다. 요구사항은 `spec.md`, 검증은 `acceptance.md`, 설계 결정은 `design.md`, 근거와 측정은 `research.md`, 열린 결정은 `decision-index.md`가 소유한다. 시간 추정 없이 우선순위와 순서만 쓴다.

## §A Context

- 에픽 소속 없음(단독 SPEC). 형제: `SPEC-LANE-STALL-WATCHDOG-001`(`auto-semantics.md` 저작), `SPEC-AUTONOMY-KICKOFF-CALIB-001`, `SPEC-AUTONOMY-GATE-REWIRE-001`, `SPEC-AUTONOMY-ESCALATION-001`.
- 위치: 워크트리 `.moai/worktrees/t1344`, 브랜치 `WT-batch-approval-gate`, 0.4.0 개정 시점 HEAD `08692e732`(SPEC 전용 커밋 둘이 로컬 `develop` 끝 위에 있고 소스 파일은 `develop`과 같다). **Tier L**(오케스트레이터 판정 R1, `spec.md` §A.5): `spec.md` + `plan.md` + `acceptance.md` + `design.md` + `research.md` (+ `spec-compact.md`, `decision-index.md`, `progress.md`).
- 변경 성격: **문서(규칙·스킬) + 리더 공지 문장(제품 Go 파일 4개: `internal/hook/session_start_{kanban,factory}{,_i18n}.go`) + 테스트 Go 파일 3개**(훅 2, 템플릿 1). 제품 Go 변경은 리더 공지 문자열에 한정된다(`design.md` §D.1·§D.7). 0.1.0의 "제품 Go 없음" 전제는 Q5 판정(2026-10-02)으로 철회됐다.
- 근거와 측정(증거 요약 E1–E17, 임시 측정, 추가 측정 원문, 0.3.0에서 새로 확인한 전제, 계획 시점 재계수, 출처 조사의 세 결과)은 **`research.md`**가 가진다. 이 문서는 그 ID(E13, R3.1 등)만 인용한다.
- 요약: plan→run Kickoff 행에서 운영자 결정을 기다리는 카드들을 한 번에 보여 주고 질문 하나로 승인받는 *제시 형식*을 `auto-semantics.md` §9.2로 정본화하고, 반대 증거 의무와 fail-closed 제외 규칙을 못 박는다. 효과의 크기는 미측정이므로(G-1) 결과 수치는 약속하지 않는다.

## §B Known Issues (이 SPEC에 해당하는 항목만)

- **B3 서브에이전트 경계**: 훅은 질문 도구를 호출하지 않는다. 이 트리에서 패키지 전역 grep 가드 문구는 거짓이다(`research.md` E13). 리더 공지 문장(REQ-BGS-016)은 질문 도구 이름을 담지 않고, 알림 소스 4파일의 `AskUserQuestion` 토큰 수는 0을 유지한다(AC-015 — `TestNoUserInteraction`과 같은 파일 범위·주석 제외 형태).
- **B4 frontmatter**: `created:`/`updated:`/`tags:` 정본 12필드, `phase`는 릴리스 목표 라벨. `design.md`·`research.md`는 status 축에서 무상태라 `status:`를 두지 않는다.
- **B6 spec-lint 제목 규약**: `### Out of Scope — <topic>` H3 + `-` 항목.
- **B8/B10 작업 트리 위생·PRESERVE**: 아래 §D 보존 목록 밖은 건드리지 않는다.
- **B13(상시 로딩 편집은 끝에)**: 이 SPEC의 상시 로딩 파일 편집은 `kanban-dispatch.md` 한 곳(M3)이다. `auto-semantics.md`는 `paths:` 범위라 세션 prefix에 상시 로딩되지 않으므로(`design.md` §D.6) M2의 편집은 이 제약에 걸리지 않는다. 그래서 **실행 순서를 M0 → M1 → M2 → M4 → M5 → M3로 한다 — M3가 마지막이다**(캐시 지시문 3). M 번호는 식별자이고 실행 순서가 아니다. M3는 M2에만 의존하므로 M4·M5 뒤로 미뤄도 선행 조건이 깨지지 않는다.

## §C Pre-flight (run-phase 시작 시 실행)

```bash
git branch --show-current                    # WT-batch-approval-gate
git rev-parse --short HEAD                   # 계획 마감 이후 HEAD 기록
go test -count=1 -v -run '^TestAlwaysLoadedTokenBudget$' ./internal/config/
go test -count=1 -v -run '^TestRuleTemplateMirrorDrift$' ./internal/template/
go test -count=1 -v -run '^(TestSubSkillLOCCeiling|TestEntryRouterLOCCeiling)$' ./internal/skills/
go test -count=1 -v -run '^(TestAutoRankDoctrineAmendment|TestAutoRankMirrorParity|TestSpecAssembly_RewrittenToCLIPath|TestSpecAssembly_NoNewInternalTokens)$' ./internal/cli/
go test -count=1 -v -run '^(TestImplementationKickoffApprovalPreservedBeforeGoal|TestTemplateNoInternalContentLeak|TestContractModeEmitterSites)$' ./internal/template/
# 리더 공지 GREEN 기준선 — 명령 전문은 research.md R3.1 E17 (23개 이름, ./internal/hook/)
```

기준선이 전부 `--- PASS`여야 시작한다. 하나라도 빨가면 이 SPEC 이전의 문제이므로 blocker로 보고한다. 플랜 단계에서 측정한 기준선: 예산·미러·LOC·보존 구간(`research.md` E9~E9c)과 리더 공지 63개(E17). `internal/template` 쪽 세 테스트(`TestImplementationKickoffApprovalPreservedBeforeGoal`, `TestTemplateNoInternalContentLeak`, `TestContractModeEmitterSites`)와 `TestSpecAssembly_NoNewInternalTokens`의 변경 전 기준선은 이번에도 돌리지 않았다 — Gap. run-phase에서 첫 명령으로 측정한다. `internal/hook` 패키지 전체는 분 단위 스위트이므로 레인 로컬에서는 위 이름 선택만 돌리고 전체는 CI에 맡긴다(전체를 돌려야 하면 `moai slot acquire`로 자원 임대를 먼저 잡는다 — `gitflow-lane-protocol.md` §8).

## §D Constraints

1. **Template-First, 바이트 동일 미러.** `.claude/...`와 `internal/template/templates/.claude/...`를 함께 편집하고 `make build`로 임베드한다. 편집 대상 파일 중 `kanban-dispatch.md`(`:177`)와 `moai.md`(`:215,245,253,283-284`)는 라이브↔미러가 이미 다르다 — 그 줄들은 그대로 둔다.
2. **언어.** 새 독트린 문장은 영어(지침 문서 언어 정책). SPEC 서술만 한국어.
3. **템플릿 중립성.** 미러 복사본에 카드 id(`tNNN`), SPEC ID, 날짜를 쓰지 않는다(`TestTemplateNoInternalContentLeak`). 배치 식별자 형식은 자리표시자(`YYYYMMDDTHHMMSSZ`)만 쓴다. 형제 SPEC도 같은 방식으로 개념 토큰만 썼다.
4. **줄·바이트·예산 한도(REQ-BGS-013에서 옮김; AC-016이 검증).** `run.md` 199/200줄, `spec-assembly.md` 597/600줄 — 새 줄을 추가하지 않고 *기존 줄을 바꿔 쓴다*(`TestEntryRouterLOCCeiling`, `TestSubSkillLOCCeiling`); `run.md`는 `git diff --numstat develop...HEAD`에서 추가 줄 수 = 삭제 줄 수. 상시 로딩 `kanban-dispatch.md`는 라이브·미러 각각 바이트 순증가를 1,000 미만으로 하고(`rule-authoring.md`), 추가 줄 수는 3 이하로 한다. **이 측정의 왼쪽 끝은 읽는 시점의 `git merge-base develop HEAD`다 — 리터럴 base SHA로 재지 않는다**(`gitflow-lane-protocol.md` §8: 흡수하는 순간 리터럴 핀 범위에 다른 카드의 커밋이 들어온다). 한계: 병합 전에만 유효하다 — 병합 뒤에는 merge-base가 카드 tip이 되어 범위가 비고 공허하게 통과하므로 병합 뒤 근거는 병합 트리와 카드 브랜치 트리의 동일성이다. 대조: `git diff --name-only develop...HEAD`가 한 줄 이상이어야 하고 0줄이면 판정은 PASS가 아니라 '측정 불가'다. 계획 시점 측정(고정 SHA에서 잰 바이트 수)은 `research.md` R5에만 둔다. `TestAlwaysLoadedTokenBudget` 측정 여유는 13,373토큰(`research.md` E9b). `workflows/moai.md`(라이브 284 / 미러 282줄)에는 줄 수 테스트가 걸려 있지 않다(E14) — 그래도 M4는 두 사본의 줄 수를 바꾸지 않는다(`:144`·`:240`은 한 줄을 한 줄로 바꿔 쓴다).
5. **고정 문구.** `TestImplementationKickoffApprovalPreservedBeforeGoal`: `run.md`는 "Implementation Kickoff Approval", "AskUserQuestion", `/goal`을 이 순서로 유지하고 "regardless of"·"score-independent" 문구도 유지. `TestSpecAssembly_RewrittenToCLIPath`: `spec-assembly.md`는 `[HARD] The Implementation Kickoff Approval`, `moai plan render-html`, `Fail-open`을 유지. `TestSpecAssembly_NoNewInternalTokens`: 금지 토큰 추가 금지.
6. **보존 구간.** `kanban-dispatch.md`의 "Promotion is the operator's act, always." → "The self-dispatch lane exception." 구간은 라이브·미러 바이트 동일을 유지하고, "never picks for the operator"·"queue ADMISSION stays the operator's" 금지 문구를 지킨다(`TestAutoRankDoctrineAmendment`, `TestAutoRankMirrorParity`). 이 구간에 `:31`의 "batch approval" 문장이 들어 있으므로 *그 줄은 편집 금지*다.
7. **계약 모드 문구.** `TestJevAmendmentLinkage`(`internal/contract/kickoff/activation_test.go:208`)가 MCP 카탈로그·workflow.yaml의 contract-mode Kickoff/Jev 예외 문자열 제거를 막는다 — 건드리지 않는다. `TestContractModeEmitterSites`(`internal/template/contract_mode_guided_test.go:695`)도 같다.
8. **Frozen 구역 보존.** `orchestration-mode-selection.md:18`의 `[ZONE:Frozen]` 조항은 편집하지 않는다. 이 SPEC이 만지는 어떤 조항도 `zone-registry.md`에 항목이 없으므로(`research.md` E9e) `moai constitution amend`는 필요 없다.
9. **명령 형식.** 워크트리 가드 때문에 검증 명령은 단순 명령·리터럴 경로로 쓴다(`worktree-integration.md` § Refused Commands). 따옴표 없는 `--exclude=*_test.go`는 셸 글롭에 걸려 실패하므로 `--exclude='*_test.go'`로 쓴다.
10. **도구 출처(VCI §2.2).** 이 플랜의 `go test`는 트리에서 직접 빌드한 결과다. `moai` 설치본(`v3.2.0-rc.23`, 커밋 `d194083fb`)은 트리 HEAD보다 오래됐으므로 `moai spec lint`는 트리에서 빌드하고 커밋을 스탬프한 바이너리로 돌려 판정 빌드 커밋을 함께 적는다(`research.md` R1: 0.3.0의 판정 빌드 커밋 `72e09d27b` = 트리 HEAD). 첫 빌드의 Go VCS 스탬프가 HEAD의 조상이 아니어서 `-ldflags="-X …/pkg/version.Commit=<sha>"`로 다시 빌드했다.
11. **리더 공지 Go 제약.** 새 문장은 기존 테스트가 이미 고정한 다음을 깨지 않는다(REQ-BGS-018). (a) 블록 구조: 칸반 공지는 정확히 5개 블록이고 3번째 블록은 명령 3줄뿐이다(`TestKanbanLeadNoticeBlockLayout`) → 문장은 *기존 블록 안의 한 줄*로 합치고 새 블록을 만들지 않는다. (b) 역할 용어: 팩토리·칸반 리더 공지의 en에는 단어 `lead`(대소문자 무시, 단어 경계)가 없고, ko에는 `리드`가 없으며, `worker-<n>`·`-f worker`·`-f agent`가 없고, 두 공지는 로케일별 leader·lane 용어를 계속 담는다(`TestRoleNamingM3NoticesCarryLeaderLaneTerms`). (c) todo 비활성 시 칸반 공지에 `moai todo`가 없다(`TestSessionStartKanbanRespectsTodoDisabled`) → 문장에 `moai todo`를 쓰지 않는다. (d) SPEC 미설정 시 `SPEC-` 부분 문자열이 없고(`TestKanbanLeadNoticeOmitsSPECWhenUnset`), 대소문자 무시 `epic`이 없다(`TestKanbanLeadNoticeBacklogSummaryCountsQueuedOnly`). (e) 팩토리 공지에 `moai todo list`·`.moai/state/kanban/backlog.json`·`poll the backlog queue`·`No model override`·`ANTHROPIC_DEFAULT_*_MODEL`이 없다(`TestFactoryLeadNoticeCarriesDispatchDiscipline`). (f) 새 필드는 형식 문자열이 아니다(`%s`·`%d` 동사 없음). (g) 질문 도구 이름 없음(B3). (h) 카드 id·SPEC ID·날짜 없음, 네 로케일 동시 갱신(`TestKanbanLocalesCoverEveryField`, `TestKanbanNoticePreservesProtocolTokensInEveryLocale`). (i) 포인터 토큰 `.claude/rules/moai/workflow/auto-semantics.md` §9.2는 네 로케일에서 번역하지 않는 주소다(기존 명령 토큰과 같은 취급). (j) 이름 토큰 `batch gate summary`도 네 로케일에서 그대로 둔다(REQ-BGS-016, AC-015).
12. **`workflows/moai.md` 편집 방식.** 라이브와 미러는 `:215`·`:245`·`:253`·`:283-284`에서 이미 다르다(`research.md` E15). M4는 각 사본에 같은 `old_string` 치환을 *따로* 적용하고, 한쪽을 다른 쪽으로 복사하거나 재동기화하지 않는다. 편집 뒤 `diff`의 헌크 헤더가 편집 전과 같은 넷이어야 한다(AC-014).
13. **정본 전용 어휘(AC-016).** `counter_refs=`와 `searched=`는 정본 두 사본(`auto-semantics.md` 라이브·미러) 밖의 어떤 규칙·스킬·훅 소스 비테스트 파일에도 쓰지 않는다. 포인터는 §9.2를 *가리킬* 뿐 서식을 옮겨 적지 않는다.
14. **기준선 파일의 위치.** 기준선은 추적되는 `.moai/specs/SPEC-AUTONOMY-BATCH-GATE-001/baseline-gate-rounds.md`에만 둔다. `.moai/reports/` 아래는 `.gitignore:235`가 무시하므로 쓰지 않고, 강제 추가(`git add -f`)도 하지 않는다(운영자 지시 2026-09-14, `.gitignore:233-246`).

## §E Self-Verification (run-phase 완료 보고에 담을 항목)

`manager-develop-prompt-template.md` §E의 E1–E8을 따른다. 이 SPEC에 특화된 항목:
- E1 AC 이진 PASS/FAIL 행렬(`acceptance.md` 기준, 명령 + 원문 출력).
- RED 원문: 가드 테스트가 정본 절 부재 상태에서 실패하는 출력(E8, TDD).
- 변이 증거: 앵커를 뺀 변이 본문 각각의 거부 출력(AC-007: 하위 테스트 `--- PASS` 46개 이상, 변이 거부 줄 45개, `[no tests to run]` 없음).
- 기준선 증거: M0 종료 때 AC-008 V1–V5(존재, `git check-ignore -v`의 exit 1, 네 요소 레이블 줄, B 하나, B의 단독 경로), 끝 점검 때 V6–V8(기준선 파일을 건드린 커밋은 B뿐, 정본 절 커밋 D 하나, B가 D의 조상).
- 상시 로딩 `wc -c` 전/후(읽는 시점의 `git merge-base develop HEAD` 기준)와 `TestAlwaysLoadedTokenBudget` 출력, `git diff --numstat develop...HEAD`의 포인터 네 곳 값과 범위 비공허 대조, 정본 전용 어휘 두 토큰의 `grep -rlF` 출력(AC-016).
- 미러 동일성(`TestRuleTemplateMirrorDrift`)과 LOC 한도 출력.
- 리더 공지(M5): AC-015의 RED 원문(파일 부재·포인터 토큰 0·이름 토큰 0), GREEN 출력(하위 테스트 `--- PASS` 17개 이상, `=== RUN`과 같은 수, 변이 거부 줄), 알림 소스 4파일의 `AskUserQuestion` 토큰 수 0, 기존 고정 테스트 63개(E17)의 재실행 결과, `git diff --numstat`(AC-009 읽기 단계).
- `workflows/moai.md`(M4): 편집 뒤 `diff` 헌크 헤더 넷과 `wc -l`(284/282) 출력(AC-014).
- Tier 재계수(`spec.md` §A.5): `git diff --name-only develop...HEAD -- . ':(exclude).moai/specs' ':(exclude).moai/reports'`의 출력 줄 수가 열거된 17과 일치하는지.

## §F Milestones

우선순위 표기만 쓴다. **M 번호는 식별자이고 실행 순서는 §B의 B13 줄이 정한다: M0 → M1 → M2 → M4 → M5 → M3.** 변경 가능성이 큰 순서(검토 우선순위)는 아래 표다.

| 검토 우선순위 | 마일스톤 | 이유 |
|---|---|---|
| 1 (가장 바뀔 가능성 큼) | M2 정본 절 | 사용자 대면 흐름(요약 보고·단일 질문·빼내기)과 반대 증거 형식, 차단 토큰 — 운영자 판정(Q1·Q2·Q3·Q6·Q7·Q10·Q13)에 직접 걸린다 |
| 2 | M5 리더 공지 문장 | 세션 시작 때 리더와 운영자가 읽는 문구(네 로케일). 명칭(Q10)이나 정본 위치가 바뀌면 같이 바뀐다. 실행은 M2 뒤 |
| 3 | M4 낡은 표현 정합 | 운영자가 범위 안으로 판정(Q4, 2026-10-02). 문구가 §9.1·§9.2에 종속된다. 실행은 M2 뒤. 범위 확대는 Q12 |
| 4 | M3 포인터 | 정본 위치가 정해지면 기계적. 상시 로딩 편집이라 실행은 마지막 |
| 5 (가장 기계적) | M0·M1 | 기준선 기록과 테스트 — 결정에 덜 민감 |

### M0 — 기준선 기록 (Priority High, 선행 필수)

- 범위를 *먼저* 정의한다: 어느 전사본 디렉터리(프로젝트 디렉터리 406개 중 어느 접두사), 어느 기간, 게이트 질문 분류 방식(질문 문구가 §9 게이트 행 이름을 포함하는가), 중복 처리.
- 산출물 **`.moai/specs/SPEC-AUTONOMY-BATCH-GATE-001/baseline-gate-rounds.md`**(추적되는 SPEC 디렉터리 안): 명령 원문, 관측 출력, 분류 방식, 한계를 줄 머리 레이블 `Command:`, `Observed output:`, `Classification method:`, `Limits:`로 싣는다 — 각각 같은 줄에 비어 있지 않은 본문이 오고(AC-008이 이 줄을 센다), 긴 출력은 아래 펜스 블록에 두고 레이블 줄에는 그 위치를 적는다. **이 파일만 담은 단독 커밋**으로 남긴다(VCI §2.3 — 기준선과 변경이 같은 커밋이면 순서를 커밋 그래프로 증명할 수 없다). 이 커밋은 `spec.md`를 건드리지 않는다: `draft → in-progress` 전이는 이어지는 M1 커밋이 싣는다. 파일은 `ComputeHash`의 계획 산출물 집합 밖이라 캐시된 감사 판정을 무효화하지 않는다(`research.md` R4.3).
- Exit: 기준선 파일이 존재하고, 무시되지 않으며, 네 요소 레이블 줄을 모두 싣고, 기준선 커밋은 그 파일 하나만 담는다.
  - 판정은 AC-008의 M0 종료 점검 V1–V5다. 기준선 커밋이 정본 절 커밋의 조상이라는 판정과 `-S` 순서는 정본 절 커밋이 M2에서 생긴 뒤에만 가능하므로 M0 종료 조건이 아니라 끝 점검 V6–V8이다(§F 끝 점검).

### M1 — 가드 테스트, RED 먼저 (Priority High)

- 새 파일 `internal/template/batch_gate_summary_doctrine_test.go`(테스트 전용). 정본 절을 템플릿 복사본에서 잘라 내 아래 앵커 표를 점검하는 검사 함수를 두고, 표 구동 하위 테스트로 (a) 실제 절은 위반 0, (b) 앵커를 하나씩 뺀(또는 반대로 쓴) 변이 본문은 *해당* 앵커 위반을 보고함을 확인한다.
- RED: 정본 절이 아직 없으므로 첫 실행은 실패해야 한다(원문 출력을 progress §E.2에 보존).
- 비공허: 변이가 거부될 때마다 `t.Logf("mutant %s rejected: reported=%s", id, reported)` 꼴의 한 줄을 찍고, 실제 절 하위 테스트는 `real_section violations=0`을 찍는다. AC-007이 이 줄의 수로 빈 하위 테스트를 거부한다.
- Exit: 테스트 파일이 있고 정본 절이 없는 상태에서 첫 실행이 FAIL이며 그 원문이 progress §E.2에 있다. GREEN 전환은 M2의 Exit다.
- 앵커 문구는 M2에서 확정한다. 각 앵커는 `spec.md` §C의 요구사항에서 파생되며 요구사항이 담은 열린 행의 초안 읽기(`spec.md` §B.8 표)를 같이 담는다 — 그 행의 판정이 다르면 앵커도 같이 바뀐다.

앵커 표(45행; 하위 테스트 `--- PASS` 수 = 45 + 실제 절 1 = 46 — AC-007):

| 앵커 | 점검하는 불변식 | REQ | AC |
|---|---|---|---|
| A01 | 요약은 plan→run Kickoff 행에만 적용(2회차 판정 R1) | 001 | 006 |
| A02 | 레인은 교차 카드 배치를 만들지 않음 | 001 | 006 |
| A03 | 준비된 카드를 붙잡지 않음 | 002 | 006 |
| A04 | 동일 판정 주체는 영향 카드를 전부 적어 한 번 질문 | 003 | 006 |
| A05 | 보고서 머리가 Kickoff 행을 적고 행 필드 목록(카드 id·SPEC id·독립 판정의 최종 반복+반복 식별자+점수+여유·해시 불변·기록 참조·`counter_refs=`) | 004 | 002 |
| A06 | keep-set·리더 보유 권한 점검 결과와 분류 근거 필드 | 004 | 002 |
| A07 | 반대 증거 행이 먼저 | 004 | 002 |
| A08 | 보고서가 질문보다 앞섬 | 004 | 002 |
| A09 | 질문은 정확히 하나이고 한 호출에 요약 질문은 하나뿐 | 005 | 002 |
| A10 | 승인은 나열된 승인 가능 행에만 | 005 | 002 |
| A11 | 나중에 온 행·예약 행·차단 행은 덮이지 않는다는 진술 | 005 | 002 |
| A12 | 빼내기는 자동 자유 입력에 카드 id를 적는 방식(행 수 무관) | 006 | 002 |
| A13 | 빼내도 나머지 행의 승인 유지 | 006 | 002 |
| A14 | `counter_refs=` 필드 | 007 | 003 |
| A15 | 반대 증거 원천 목록 여섯 | 007 | 003 |
| A16 | `counter_refs=none`에는 `searched=`와 공백 없는 단일 토큰, 보고서와 기록 양쪽 | 008 | 003 |
| A17 | `searched=` 없는 `counter_refs=none`은 반대 증거 진술이 아님 | 008 | 003 |
| A18 | 승인된 행마다 한 줄 자기 결정 기록: §10 세 필드를 순서대로, 이어서 `counter_refs=` | 009 | 003 |
| A19 | `ladder_path`의 게이트 행 슬러그 뒤 `;batch=<id>`, `<id>` 형식·한 요약 안 동일·결정 보드에 같은 값이 있으면 다음 빈 초 | 009 | 003 |
| A20 | 기록 없는 행은 미승인 | 009 | 003 |
| A21 | 기록 직전 재확인은 REQ-011의 네 조건(판정·audit-ready·해시·열린 차단) 전부 | 019 | 003 |
| A22 | 표류한 행은 거부하고 운영자에게 알림 | 019 | 003 |
| A23 | keep-set 범주 1 환경상 불가능 | 010 | 004 |
| A24 | keep-set 범주 2 운영자 보유 | 010 | 004 |
| A25 | keep-set 범주 3 외부 공유 시스템의 되돌릴 수 없는 조작 | 010 | 004 |
| A26 | 리더 보유 권한 여섯 항목 모두 | 010 | 004 |
| A27 | contract 모드: 서명이 게이트, 요약 행 없음 | 012 | 004 |
| A28 | 판정 참조는 현재 산출물의 최종 반복 판정에 결속 | 011 | 005 |
| A29 | 승인 가능은 PASS·audit-ready·해시 불변·열린 차단 없음 *넷 모두*이고 그 밖의 행은 차단으로 보고되어 승인에서 빠진다(보고 위치는 Q3) | 011 | 005 |
| A30 | PASS-WITH-DEBT는 차단 | 011 | 005 |
| A31 | BYPASSED는 차단 | 011 | 005 |
| A32 | FAIL은 차단 | 011 | 005 |
| A33 | INCONCLUSIVE는 차단 | 011 | 005 |
| A34 | 부재한 판정은 차단 | 011 | 005 |
| A35 | audit-ready 미기록은 차단 | 011 | 005 |
| A36 | 판정 뒤 해시 변경은 차단 | 011 | 005 |
| A37 | 열린 차단은 차단 | 011 | 005 |
| A38 | 결과 수치("176"류) 부재(부재 앵커) | — | 007 |
| A39 | Kickoff 외 게이트 행(sync 차단 승인·카드 선택·factory decide 행들)은 요약 행이 아니고 운영자 답이 필요하면 개별 질문 | 001 | 006 |
| A40 | 공유 판정 질문은 차단 행·예약 행을 이름에 올리지 않음 | 003 | 006 |
| A41 | 질문 문구가 읽기 규칙을 적음: 적힌 카드 id는 빼내고 나열된 나머지는 승인 | 006 | 002 |
| A42 | 읽을 수 없는 입력·표에 없는 카드 id는 아무것도 승인하지 않고 같은 질문을 다시 냄 | 006 | 002 |
| A43 | 기록 직전 재확인은 REQ-010 분류(운영자 hold 포함)도 포함 | 019 | 003 |
| A44 | 판정은 독립(plan-auditor가 낸 것)이어야 함 — 작성 세션의 자기 진술 PASS는 차단 | 004, 011 | 005 |
| A45 | 승인은 Kickoff의 다른 조건을 약화하지 않음: 카드마다 티어·모드 선호·PR 전략·체인 범위가 run 진입 전 디스크에 있음 | 020 | 002 |

### M2 — 정본 절 `auto-semantics.md` §9.2 (Priority High)

- 위치: §9.1 바로 뒤. 제목 `### 9.2 The batch gate summary`. 분량 목표는 3KB 안팎(상한 없음, 단순 유지).
- 내용: 앵커 표 A01–A45가 정한다(최종 문구는 run-phase가 작성하고 앵커를 모두 만족한다). 앵커 표가 말하지 않는 것만 여기에 둔다 — 질문 채널이 없는 하니스에서는 같은 요약이 blocker 보고서로 나간다(G-6); `--auto` 배치 권한 부여와 별개라는 구별 문장; 권고 라벨은 현행 `recommendation_mode` 규약 그대로; 카드 id·SPEC ID·날짜를 쓰지 않는다(제약 3).
- §10에 한 문장 추가: 요약으로 승인된 결정은 `counter_refs=`와 `;batch=<id>`를 싣는다(서식은 §9.2). §9.1 끝에 §9.2 포인터 한 줄.
- 템플릿 먼저 → 라이브 동일 내용 → `make build`. 미러 동일성 확인.
- Exit: AC-001의 두 grep이 `1`이고, `TestBatchGateSummaryDoctrine`이 AC-007의 점검(`--- PASS` 46개 이상·`--- FAIL` 0·`=== RUN`과 같은 수·`[no tests to run]` 없음·변이 거부 줄 45개)을 만족하며, `TestRuleTemplateMirrorDrift`가 `--- PASS`다.

### M4 — 낡은 표현 정합 (Priority Medium, 무조건 — Q4 범위 안, 2026-10-02; **M2 뒤에 실행**)

- 선행: M2 커밋이 브랜치에 있어야 한다(다시 쓴 문구가 §9.1·§9.2를 가리킨다).
- `spec-assembly.md:202-208`(7줄)을 같은 줄 수로 다시 쓴다. 보존: 첫 줄의 `[HARD] The Implementation Kickoff Approval` 시작 문구(`TestSpecAssembly_RewrittenToCLIPath`가 템플릿 사본에서 고정), `moai plan render-html` 단계(`:197`), `Fail-open` 문단(`:210` 이후), HTML 보고서가 게이트를 *대체하지 않고 보강만 한다*는 취지와 `(권장)`/pull 라벨 설명. 정정: "stays MANDATORY and score-independent"와 "A plan-auditor PASS or a high skip-eligible score does NOT substitute for the gate."를 `auto-semantics.md` §9.1(기본은 자율 전환, keep-set은 운영자 답)에 맞게 쓰고 §9.2를 가리킨다. 597줄 불변(한도 600, `TestSubSkillLOCCeiling`은 라이브 사본만 읽는다). 새 내부 토큰(SPEC ID·REQ/AC 토큰) 추가 금지(`TestSpecAssembly_NoNewInternalTokens`).
- `workflows/moai.md:144`(파이프라인 게이트 2)와 `:240`(Step 11.3)을 각각 *한 줄을 한 줄로* 바꿔 쓴다. 정정 대상: `:144`의 "Score-independent: a plan-auditor PASS or skip-eligible score never bypasses it"와 `:240`의 "score-independent". 같은 줄의 보존 서술(게이트를 한 파이프라인 진입당 한 번 제시, 병합 라운드가 질문 둘을 한 호출에 싣는 문장, 파생 완료 조건이 run 진입을 허락하지 않는다는 문장)은 유지한다. 두 줄의 레이블 `HUMAN GATE`는 keep-set 형태에서만 문자 그대로 참이다 — 함께 정정할지는 run-phase가 §9.1 문구와 대조해 정하고 이유를 progress §E.2에 남긴다(정정 범위를 임의로 넓히지 않는다). 뒤따르는 contract-mode 블록(`contract-pipeline-gates`, `contract-merged-round`)은 건드리지 않는다(`contract_mode_blocks_test.go:54`가 고정). 적용은 두 사본에 *각각* 같은 `old_string`으로(제약 12). 줄 수 284/282 불변, 편집 뒤 `diff` 헌크 헤더가 `215c215`·`245c245`·`253d252`·`283,284c282` 넷 그대로(AC-014).
- 이 마일스톤은 두 파일만 고친다(REQ-BGS-014). 같은 종류의 문구를 가진 다른 파일(아래 표면 목록의 영향·표지만 분류)은 이 SPEC이 고치지 않는다 — 확대 여부는 Q12.
- Exit: AC-014의 grep 다섯 줄·`wc -l`·`diff` 헌크 헤더·`TestSpecAssembly_*` 두 테스트가 통과한다(§9.2가 이미 있어야 하므로 M2 뒤).
- 검증: `TestSpecAssembly_RewrittenToCLIPath`, `TestSpecAssembly_NoNewInternalTokens`, `TestSubSkillLOCCeiling`, `TestImplementationKickoffApprovalPreservedBeforeGoal`(run.md만 읽으므로 영향이 없음을 확인), `TestContractModeEmitterSites`와 contract-mode 블록 테스트, 새 문구에 `manager-tdd`·`manager-ddd` 금지(고아 참조 스캔), AC-014의 grep·`diff`·`wc -l`.

### M5 — 리더 공지 문장 (Priority Medium, 무조건 — Q5 범위 안, 2026-10-02; **M2 뒤에 실행**)

- 선행: M2 커밋이 브랜치에 있어야 한다(포인터가 가리킬 정본 절이 먼저 있어야 한다). 훅 Go 파일은 세션 로딩 파일이 아니므로 `cache-aware-execution.md` 지시문 3의 "상시 로딩 파일 편집은 끝으로" 제약은 이 마일스톤에 걸리지 않는다. M3·M4와의 순서는 M2에만 의존한다.
- **두 리더 공지에 모두 싣는다**(근거 `design.md` §D.7, `research.md` E12b). 레인·동반 공지는 받지 않는다.
- 제품 Go 대상, 정확히 넷:
  - `internal/hook/session_start_kanban_i18n.go`: 구조체(`:37-51`)에 문장 필드 하나(후보 이름 `leaderGateSummary`), 네 로케일 값(en `:60-83`, ko `:84-105`, ja `:106-127`, zh `:128-149`).
  - `internal/hook/session_start_kanban.go`: `kanbanLeaderNotice`(`:116-197`)의 문맥 블록 (e)(`:177-194`)에서 설정 줄 뒤에 그 필드를 한 줄로 합친다. 새 블록 없음.
  - `internal/hook/session_start_factory_i18n.go`: 구조체(`:35-67`)에 같은 필드, 네 로케일 값(en `:74-124`, ko `:125-169`, ja `:170-214`, zh `:215-258`).
  - `internal/hook/session_start_factory.go`: `factoryLeaderNotice`(`:185-249`)의 규율 블록 (e)(`:222`)의 `strings.Join` 목록에 그 필드를 더한다.
- 테스트 Go 대상: 기존 `session_start_kanban_i18n_test.go`의 `TestKanbanLocalesCoverEveryField` 필드 표(`:27-41`)에 새 필드를 덧붙인다(약화 없음). 새 `internal/hook/session_start_leader_gate_notice_test.go`: 두 리더 공지 × 네 로케일을 렌더해 검사 함수에 넣는다(포인터 토큰 존재; 금지 토큰 부재 — 질문 도구 이름, `SPEC-`, 카드 id 꼴, ISO 날짜, `moai todo`, 단어 `lead`, `리드`, `epic`). 레인·동반 공지에 포인터가 없음을 확인한다. 변이 본문(로케일 하나 누락, 질문 도구 이름 포함, 포인터 누락, 이름 토큰 누락, 카드 id 포함, 한쪽 리더 공지만 보유)이 각각 거부됨을 보인다(거부마다 `mutant <이름> rejected: reported=<이름>` 줄 — AC-015). 알림 소스 4파일의 주석 제외 토큰 스캔(`TestNoUserInteraction` 형태).
- 문장의 모양은 `design.md` §D.7이 가진다(최종 문구는 run-phase가 쓴다): 한 문장, 이름 토큰 `batch gate summary`와 포인터 `.claude/rules/moai/workflow/auto-semantics.md` §9.2, 정본의 서식·제외·반대 증거 의무는 *적지 않는다*(REQ-BGS-016 "restate 금지").
- 순서: RED 먼저. 테스트 파일을 먼저 만들어 실패(포인터 부재)를 관측하고 원문을 progress §E.2에 남긴 뒤, 필드와 값을 더해 GREEN으로 뒤집는다. 제품 Go와 테스트는 한 커밋이어도 된다(VCI §2.3은 기준선에만 걸린다).
- 비용 공시: 리더 세션의 `startup`에서만 문장 한 줄이 늘고 비리더 세션은 0바이트다. 렌더 길이의 전/후는 run-phase에서 공시한다(플랜 단계에서는 측정하지 않았다 — UNVERIFIED).
- 검증: E17 명령 전체 PASS(63개 + 새 하위 테스트), 새 테스트 PASS(하위 테스트 `--- PASS` 17개 이상·`=== RUN`과 같은 수·`[no tests to run]` 없음·변이 거부 줄), 4파일 토큰 수 0, 변경 파일에 `gofmt -l`·`go vet`·CI 버전 golangci-lint.
- Exit: AC-015의 Verify 네 줄(새 테스트 출력 계수, `AskUserQuestion` 토큰 `:0`, 포인터 토큰 `:4`, 이름 토큰 `:4`)이 통과하고 기존 고정 테스트 63개가 그대로 `--- PASS`다.

### M3 — 포인터 (Priority Medium; **실행은 M5 뒤, 마지막** — 상시 로딩 파일 편집)

- 선행: M2 커밋(정본 절 존재). 상시 로딩 `kanban-dispatch.md`는 이 마일스톤에서만 편집한다(B13).
- `kanban-dispatch.md` Boundaries "No gate bypass." 불릿 안에서 문장을 다듬어 §9.2를 가리킨다(상시 로딩, 제약 4의 한도 안, 보존 구간 `:29-33` 밖). 미러 동시 편집, `:177` 사전 차이는 유지.
- `run.md:137`의 기존 문장에서 "§9.1" 인용을 "§9.1–9.2"로 바꾸는 식의 *줄 수 불변* 편집만. 고정 문구·순서 유지.
- 정본 서식을 옮겨 적지 않는다(제약 13). 검증: 예산·LOC·AutoRank·Kickoff 보존 테스트 전부와 AC-016의 (a)–(d).
- Exit: AC-016의 (a)–(d)(읽는 시점의 merge-base, 범위 비공허 대조 포함)와 AC-010의 가드 전부가 통과한다. 병합 전에 판정한다.

### 끝 점검 — 마지막 마일스톤(M3) 뒤, 완료 보고 전

마일스톤 종료 조건이 아니다. 앞 마일스톤 전부의 산출물이 있어야 판정되는 점검이다(마일스톤 종료가 뒤 마일스톤의 산출물에 기대면 그 종료는 도달 불가능하다 — 2회차 감사 N1).
- AC-008 V6–V8: 기준선 파일을 건드린 커밋은 B 하나, 정본 절 커밋 D 하나, B가 D의 조상.
- AC-009: 바뀐 `.go` 파일이 허용 일곱 줄.
- AC-010: 고정 가드·예산 GREEN. Tier 재계수가 17(`spec.md` §A.5).

## §G Anti-Patterns

- 요약을 *승인 수단*으로 오해해 행별 증거 판독을 건너뛰는 것 — 요약은 제시 형식이다. 행마다 §9.1 증거(최신 판정 PASS·해시 불변·차단 없음)를 리더가 읽는다.
- PASS-WITH-DEBT나 BYPASSED 판정의 행을 승인 가능 행에 넣는 것(REQ-BGS-011).
- 응답 뒤 기록을 쓸 때 판정·해시를 다시 읽지 않는 것(REQ-BGS-019).
- 대기 카드를 모으려고 준비된 카드를 지연시키는 것(REQ-BGS-002).
- keep-set·리더 보유 권한 행을 "일괄이 편하니까" 넣는 것(REQ-BGS-010).
- `none`만 적고 검색 내용을 빼는 것(REQ-BGS-008).
- 기준선과 변경을 한 커밋에 담는 것(VCI §2.3), 또는 기준선을 `.moai/reports/` 아래에 두고 강제 추가하는 것(제약 14).
- 새 줄을 늘려 `run.md`·`spec-assembly.md` 한도를 넘기는 것.
- 리더 공지 문장에 정본의 서식·제외·반대 증거 규칙을 옮겨 적는 것(REQ-BGS-016 "restate 금지") — 두 벌이 되는 순간 갈라진다. 같은 이유로 포인터 표면에 `counter_refs=`·`searched=`를 쓰는 것(제약 13).
- 리더 공지 문장을 레인·동반 공지나 레인 규칙 필드에 넣는 것, 또는 한쪽 리더 공지에만 넣는 것.
- 공지에 새 블록을 만들어 `TestKanbanLeadNoticeBlockLayout`의 5블록 구조를 깨는 것.
- `moai.md`의 라이브와 미러를 맞추려고 한쪽을 복사하거나 재동기화해 사전 차이(`:215`·`:245`·`:253`·`:283-284`)를 건드리는 것.
- 패키지 전역 `internal/hook` grep을 통과 기준으로 쓰는 것 — 이 트리에서 그 가드 문구는 거짓이다(E13).
- 이 SPEC이 고치지 않는 형제 파일(Q12)을 조용히 같이 고치는 것.
- Kickoff 외 게이트 행(sync 차단 승인, factory decide 행들, 카드 선택)을 요약에 넣거나, 공유 판정 1회 질문에 차단·예약 행을 이름 올리는 것(REQ-BGS-001·003).
- 리터럴 base SHA를 "이 카드가 무엇을 바꿨는가"의 왼쪽 끝으로 쓰는 것(`gitflow-lane-protocol.md` §8; 제약 4).
- 마일스톤 종료 조건을 뒤 마일스톤이 만드는 산출물에 거는 것 — 그 판정은 끝 점검이다.
- 하위 테스트 `--- PASS` 줄 수만 세고 변이 거부 줄 수와 `=== RUN` 일치를 안 보는 것(빈 하위 테스트가 통과한다 — AC-007·AC-015).

## §H Cross-References

- `.claude/rules/moai/workflow/auto-semantics.md` §7, §9, §9.1, §10, §11 — 정본 기반
- `.claude/rules/moai/workflow/kanban-dispatch.md` — 리더 책임, 보존 구간, Boundaries, 리더 보유 권한(`:106`)
- `.claude/rules/moai/core/askuser-protocol.md` § Report-Before-Ask Gate, § Recommendation mode, Structural Constraints(선택지 상한)
- `.claude/rules/moai/core/verification-claim-integrity.md` §1, §2.2, §2.3
- `.claude/rules/moai/development/verification-completeness.md` §1.1, §2, §2.1, §3
- `internal/hook/session_start_{kanban,factory}{,_i18n}.go` — 리더 공지 조립과 로케일 표(M5); `internal/hook/CLAUDE.md` § Conventions(서브에이전트 경계 문구, E13)
- `.claude/rules/moai/development/rule-authoring.md` — 비호출 세션 비용 공시 형식(리더 공지 문장의 R-3에 준용)
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier, § Report Persistence
- 형제: `.moai/specs/SPEC-LANE-STALL-WATCHDOG-001/spec.md` §A.7

## 표면 목록 (D6 형제 점검 결과, 0.3.0에서 Kickoff 언급 38개 파일 전부를 줄 단위로 읽어 분류)

점검 방법: `grep -rIl -i "kickoff" .claude/skills .claude/rules .claude/agents .moai/docs` → 38개 파일. 파일마다 `grep -n -i "kickoff" <파일>`로 줄을 뽑아 읽었다. 분류(영향 A·표지만 B·무관 C)의 기준과 줄별 원문은 `research.md` R4.4가 가진다. 분류는 읽기 판정이며 기계 점검이 아니다.

| 분류 | 파일 | 처분 |
|---|---|---|
| **편집(항상)** | `.claude/rules/moai/workflow/auto-semantics.md` (+미러) | §9.2 신설, §9.1 끝·§10 한 줄 (M2) |
| **편집(항상)** | `.claude/rules/moai/workflow/kanban-dispatch.md` (+미러) | Boundaries 불릿 포인터(상시 로딩, 보존 구간 밖) (M3, 마지막) |
| **편집(항상)** | `.claude/skills/moai/workflows/run.md` (+미러) | `:137` 줄 수 불변 포인터 (M3) |
| **편집(항상)** | `internal/template/batch_gate_summary_doctrine_test.go` | 신규 테스트 전용 파일 (M1) |
| **편집(항상, Q4 범위 안)** | `.claude/skills/moai/workflows/plan/spec-assembly.md:202-208` (+미러) | 낡은 표현 정합 (M4) |
| **편집(항상, Q4 범위 안)** | `.claude/skills/moai/workflows/moai.md:144,240` (+미러, 사전 차이 줄 미접촉) | 낡은 표현 정합 (M4) |
| **편집(항상, Q5 범위 안)** | `internal/hook/session_start_kanban_i18n.go`, `session_start_kanban.go`, `session_start_factory_i18n.go`, `session_start_factory.go` | 리더 공지 문장 하나, 두 리더 공지 × 4 로케일 (M5) |
| **편집(항상, Q5 범위 안)** | `internal/hook/session_start_kanban_i18n_test.go`(덧붙임), `internal/hook/session_start_leader_gate_notice_test.go`(신규) | 필드 빈값 표 + 문장 규칙 점검 (M5) |
| **영향 A — 읽고 변경 없음(Q12)** | `.claude/agents/harness/workflow-specialist.md:72-77`, `.claude/skills/hns-moaiadk-patterns/SKILL.md:150`, `.claude/rules/moai/workflow/session-handoff-examples.md:237`, `.claude/skills/moai/workflows/project/doc-generation.md:384,392` | 카드별 운영자 답을 현재형으로 단정하는 문구가 있다. Q4가 이름 붙인 두 파일 밖이라 고치지 않는다. 확대 여부는 decision-index Q12 |
| **표지만 B — 읽고 변경 없음** | `.claude/skills/moai/workflows/plan.md:53,73`, `.claude/skills/moai-workflow-spec/SKILL.md:173,191`, `.claude/skills/moai/workflows/design.md:48,52`, `session-handoff-examples.md:242,250`, `.claude/skills/moai/SKILL.md:166,372` | "HUMAN GATE"/"mandatory human gate" 표지 또는 조건 순서 문장이며 게이트가 충족돼야 한다는 뜻으로 읽힌다. `moai/SKILL.md`는 두 사본이 이미 달라(`diff -q` exit 1) 손대지 않는다. 확대 여부는 Q12 |
| **무관 C — 읽고 변경 없음** | `goal.md`, `factory.md`, `run/{mode-orchestration,task-decomposition,phase-execution}.md`, `goal-directive.md`, `goal-directive-detail.md`, `session-handoff.md`, `spec-workflow.md`, `cadence-bridge.md`, `askuser-protocol.md`, `orchestration-mode-selection.md`(`:18` Frozen), `contract-autonomy.md`, `manager-develop.md`, `manager-design.md`, `plan-auditor.md`, `session-handoff-appendix.md`, `dynamic-workflows.md`, `archived-agent-rejection.md`, `coding-standards.md`, `moai-mcp-tools-catalogue.md`, `e2e.md`, `harness-build-entry.md`, `moai-kanban-foreman/SKILL.md:73`, `.moai/docs/jev-local-operations.md` | 이미 §9.1을 인용하거나 선행 조건 인용이거나 다른 게이트이거나 contract 모드 정본이다(줄별 사유: `research.md` R4.4) |
| **로컬 전용, 비배포** | `.moai/docs/jev-local-operations.md`, `.claude/rules/local/*` | 편집 없음(3등급 독트린의 집은 그대로) |

## 요약 승인의 적용 방식 (문서 수준)

- 범위는 plan→run Kickoff 행이다. 승인된 SPEC마다 run-phase 진입을 허락하고 카드마다 §10 한 줄 기록(`;batch=<id>`, `counter_refs=` 포함)을 남긴다. 빼낸 행은 개별 질문. 응답을 받은 뒤 기록 직전에 행마다 REQ-BGS-011의 네 조건과 REQ-BGS-010 분류를 다시 읽고 표류한 행은 기록하지 않고 거부한다(REQ-BGS-019). 승인된 카드의 선호는 run 진입 전에 디스크에 있어야 한다(REQ-BGS-020).
- `moai factory decide <card>... --gate kickoff`는 별도 게이트 행(`factory decide: kickoff approve/reject`)이라 요약 행이 아니다(2회차 판정 R1, Q13). 이 SPEC은 factory 경로의 일괄 승인을 정의하지 않는다.
- 결정 기록은 디스크 결정 보드(§11, 홈 표면)에 줄로 쓴다.
- Go 판단(제품 Go는 리더 공지 문자열 한정, 배치 기구의 Go 강제는 Q8로 열림, 테스트 전용 Go 3파일)의 근거와 기각한 대안은 `design.md` §D.1이 가진다.

## §I Audit iteration 1 disposition

플랜 감사 1회차(`.moai/reports/t1344/plan-audit-iter1.md`, 로컬 전용; FAIL, 0.70)의 결함 D1–D17을 하나씩 처리했다. 파일:줄은 0.3.0 작성 직후의 위치다. 오케스트레이터 판정 R1–R5는 표의 해당 행에 적었다. 이 표는 이번 개정의 처분 기록이고 요구사항·검증을 되풀이하지 않는다.

| 결함 | 처분 | 파일:줄 |
|---|---|---|
| D1 기준선 무시 경로 | R2 적용: 기준선을 추적되는 SPEC 디렉터리 `…/baseline-gate-rounds.md`로 이동. RED-3의 거짓 전제("보고서 아래는 추적 대상")를 `git check-ignore -v` 관측으로 정정(양성 대조 포함). AC-008·M0·REQ-BGS-015·§E 증거 목록을 새 경로로 다시 가리킴. 조상 판정이 같은 커밋도 통과시키는 점을 `git show --name-only`로 막음. `ComputeHash` 집합 밖임을 `planArtifactNames`로 확인 | acceptance.md:79, :202; plan.md:59, :78, :51; spec.md:152; research.md:185, :207 |
| D2 낡은 범위 문장 | R5 적용: Q8을 "제품 Go 네 파일 + 테스트 Go 셋, 배치 기구 자체의 Go 강제는 열림"으로 고침. 옛 전제 문자열 grep을 모든 산출물에 돌려 결과를 보고 | decision-index.md:54-59 |
| D3 부정 열거 | R3 적용: REQ-BGS-011을 양성형으로 재작성(최종 반복 판정 PASS + audit-ready + 해시 불변 + 열린 차단 없음). PASS-WITH-DEBT·BYPASSED·FAIL·INCONCLUSIVE·부재를 차단 토큰으로 명명. §9.1이 PASS-WITH-DEBT·BYPASSED에 침묵함을 사실로 기록하고 차단으로 다룸. REQ-BGS-004의 판정 참조를 최종 반복에 결속. AC-005 앵커 열 개(A28–A37)와 토큰별 변이 | spec.md:145, :132; acceptance.md:56; plan.md:121-130; design.md:75; research.md:200 |
| D4 예약 목록 미정의 | R4 적용: "user-facing behavior changes" 삭제. 예약 목록을 keep-set 세 범주와 `kanban-dispatch.md:106`의 여섯 항목(원문을 직접 읽어 확인)으로 한정하고 REQ-BGS-010·설계·Q6이 같은 원천을 가리키게 함. §B.5는 목록을 되풀이하지 않는 색인으로 바꿈. Q6은 열린 채 | spec.md:144, :91; decision-index.md:40; design.md:75; research.md:238 |
| D5 미커버 조항 | 신설 AC-016: 정본 전용 어휘 두 토큰의 `grep -rlF`가 정본 두 사본만 내는지(포인터만 남김 점검, 변이 포함), 포인터 네 곳의 `§9.2`, 바이트 순증가와 `git diff --numstat` 줄 증가(1회차에는 리터럴 base SHA가 왼쪽 끝이었고 2회차 N5가 읽는 시점 merge-base로 바꿨다 — §J). REQ-BGS-013 → AC-016 직접 매핑. RED-7 신설 | acceptance.md:163, :251, :309; plan.md:41 |
| D6 Tier | R1 적용: `tier: L`. 계수 규칙(경로 단위, 라이브·미러 각각, `.moai/specs/`·`.moai/reports/` 제외)과 재계수 명령을 §A.5에 못 박음. 열거 17 = 재계수 명령 값. `design.md`·`research.md` 추가(status 없음). Q11은 POLICY-COVERED로 두고 오케스트레이터 판정으로 재작성 | spec.md:14, :73; decision-index.md:75; design.md:1; research.md:1 |
| D7 승인-기록 TOCTOU | 신설 REQ-BGS-019: 응답 뒤 기록 직전에 행마다 최신 판정·해시를 다시 읽고 표류한 행은 승인으로 기록하지 않고 거부. AC-003(A21·A22)·AC-017(시나리오) | spec.md:162; acceptance.md:40, :175; plan.md:114 |
| D8 keep-set 분류 불가시 | REQ-BGS-004에 행별 keep-set·리더 보유 권한 점검 결과와 분류 근거 필드 추가. 제외 행을 나열하든 않든(Q3) 성립한다고 Q3·REQ-BGS-005에 명시 | spec.md:132, :133; design.md:45; decision-index.md:19; plan.md:99 |
| D9 빼내기 vs 선택지 상한 | REQ-BGS-006: 질문 채널의 자동 자유 입력에 빼낼 카드 id를 적는 방식으로 명시(행 수 무관). 상한(질문당 선택지 4개, 자동 "Other")을 원문으로 확인해 근거로 적음 | spec.md:134; design.md:45; research.md:242; acceptance.md:32 |
| D10 배치 식별자 | REQ-BGS-009에 형식 정의: `ladder_path=<게이트 행 슬러그>;batch=<YYYYMMDDTHHMMSSZ>`, 한 요약 안 동일·요약 간 구별(요약 간 구별은 2회차 N10이 결정 보드 점검으로 대체 — §J). 가드 앵커 A19, M-2는 `batch=`로 묶어 계산. 같은 초 충돌은 질문 직렬성에 기댄 가정으로 공시 | spec.md:140, :98; design.md:45; acceptance.md:40; plan.md:112 |
| D11 REQ-016 격자 | 에이전트용 사본이 구성상 영어임을 두 파일의 머리 주석으로 확인하고(`session_start_kanban.go:17-19`, `…_i18n.go:17-26`) REQ를 셋으로 분할: 016(문장 내용), 017(에이전트용 영어 사본·운영자용 4 로케일에 있음, 레인·동반에는 없음), 018(기존 고정 문자열 보존). AC-015·추적표 갱신 | spec.md:156-158, :54; acceptance.md:153; research.md:246 |
| D12 레거시 형태·HOW | REQ-BGS-010을 사건 구동형(When)으로 재작성. REQ-BGS-013에서 바이트·줄 수치와 이름 붙은 절을 빼고 plan 제약 4로, 검증은 AC-016으로 이동 | spec.md:144, :150; plan.md:41 |
| D13 decision-index 편향 | Q7·Q10을 세 형태 모두의 비용을 대칭으로 다시 씀. 초안의 이유는 design으로 이동하고 "판정 아님"을 표시 | decision-index.md:47, :68; design.md:69, :107 |
| D14 열린 행 읽기 | 각 요구사항 끝에 담은 열린 행 표지를 달고 §B.8에 요구사항→행 지도 표를 둠. 열린 행은 풀지 않음. 감사가 짚은 REQ-BGS-013의 읽기는 재작성으로 사라졌고 REQ-BGS-016·017·018·019는 각각 Q10·Q5 판정·Q8을 표시 | spec.md:103-117, :126-162 |
| D15 A5 재계수 불일치 | 열거에서 기준선 산출물을 뺌(`.moai/specs/` 아래라 계수 밖). 열거 17 = 재계수 명령 값을 계획 시점에 측정(기존 15 `git ls-files` + 신규 2 `ls` 부재) | spec.md:73; research.md:250 |
| D16 B13 대 M3 | M3(상시 로딩 `kanban-dispatch.md` 편집)을 마지막으로 옮김: 실행 순서 M0 → M1 → M2 → M4 → M5 → M3, M 번호는 식별자. `auto-semantics.md`가 `paths:` 범위라 M2는 B13에 걸리지 않음을 확인 | plan.md:19, :68, :172 |
| D17 형제 파일 | 38개 파일의 Kickoff 줄을 읽어 영향 A 네 파일·표지만 B 다섯 파일·무관 C로 분류(명령과 줄별 원문은 research). REQ-BGS-014는 두 파일만 덮는다고 못 박고 Out of Scope 절을 신설. 확대 여부는 Q12로 열어 둠 | plan.md:208; research.md:211; spec.md:190, :151; decision-index.md:82 |

감사 주장 가운데 이번에 저장소에서 직접 다시 확인한 것: D1(`git check-ignore -v`), D2(옛 Q8 문장), D3(§9.1 문면과 두 토큰의 부재), D4(`kanban-dispatch.md:106` 여섯 항목), D5(AC-010에 바이트 Verify 부재), D6(열거 17), D9(선택지 상한), D11(머리 주석), D16(B13과 M3 순서). 나머지는 읽어서 확인했다. 감사 보고서와 저장소 사이의 어긋남은 하나다: D17은 미독 파일을 "열 개"라 적었으나 이전 표면 목록의 "개별 정독 안 함" 행에는 19개 파일이 있었다(방향은 같고 규모가 크다). 이번에 Kickoff를 언급하는 38개 파일 전부를 읽어 분류했다.

## §J Audit iteration 2 disposition

플랜 감사 2회차(`.moai/reports/t1344/plan-audit-iter2.md`, 로컬 전용; FAIL, 0.775, MP-9 실패)의 결함 N1–N15와 오케스트레이터 2회차 판정 R1–R4(근거는 `design.md` §D.8)를 하나씩 처리했다. 1회차 표(§I)의 D5 행이 말하는 고정 SHA 측정은 N5로 대체됐다. 이 표는 처분 기록이고 요구사항·검증을 되풀이하지 않는다.

| 결함 | 처분 | 위치 |
|---|---|---|
| N1 M0 종료 도달 불가 | R3: M0 종료를 {존재, 무시되지 않음, 네 요소 레이블 줄, 기준선 커밋은 그 파일만}으로 묶고 조상·`-S` 순서는 끝 점검 V6–V8로 옮김. 같은 문장을 plan M0 `Exit:`와 AC-008 M0 Exit에 그대로 인용. 마일스톤마다 `Exit:` 줄, §F 끝 점검 신설 | plan §F M0·끝 점검; acceptance AC-008, §D.3 |
| N2 sync 게이트 fail-closed 틈 | R1: 요약을 plan→run Kickoff 행에만 적용하고 나머지 행은 개별 질문. REQ-001·003·004·005·006·011·019, AC-006, 앵커 A01·A39 재작성, Q13 신설 | spec §C.1–C.4, §D; decision-index Q1·Q13; design §D.2·§D.8 |
| N3 1회 질문이 차단 행을 받음 | REQ-003에 "neither reserved nor blocked", 앵커 A40과 변이(차단 행을 이름에 올린 질문), AC-013에 차단 카드 I | spec REQ-003; plan 앵커 표; acceptance AC-006·AC-013 |
| N4 AC-008 placeholder 변이 | 네 요소 레이블 줄 점검 V3, 기준선 파일을 건드린 커밋은 하나 V6, `-S` 출현 커밋 정확히 한 줄 V7, 변이 8종 | acceptance AC-008; spec G-10 |
| N5 AC-016 리터럴 base SHA | R4: 읽는 시점 merge-base, 병합 전 한정 한계, 범위 비공허 대조. 리터럴은 `research.md` R5에만 | acceptance AC-016, RED-7; plan 제약 4·§E |
| N6 Q3 대 REQ-011·AC-011 | R2: REQ-011·AC-011을 "차단 행을 차단으로 보고하고 승인에서 뺀다"로 다시 씀. Q3를 "표 안/표 밖"으로 좁혀 정확하게 고침. §B.8·design 정렬 | spec REQ-004·011, §B.8; acceptance AC-011; decision-index Q3 |
| N7 빼내기 자유 입력 의미 | REQ-006에 읽기 규칙(적힌 id는 빼내고 나머지는 승인)과 실패 규칙(읽을 수 없는 입력·표에 없는 id는 승인 없이 재질문). 앵커 A41·A42, AC-012 변형 | spec REQ-006; acceptance AC-002·AC-012; design §D.3 |
| N8 독립 누락 | REQ-004·011에 "independent plan-audit verdict", 앵커 A44와 변이(자기 진술 PASS) | spec REQ-004·011; acceptance AC-005 |
| N9 재확인이 술어보다 좁음 | REQ-019가 REQ-011 네 조건 전부와 REQ-010 분류(hold 포함)를 다시 읽음. 앵커 A21·A43 | spec REQ-019; design §D.5 |
| N10 배치 식별자 유일성 | 한 호출에 요약 질문 하나(REQ-005) + 결정 보드에 같은 값이 있으면 다음 빈 초(REQ-009)로 가정을 없앰. 앵커 A09·A19 | spec REQ-005·009; design §D.3 |
| N11 AC-007 빈 하위 테스트 | `=== RUN`=`--- PASS`, `[no tests to run]` 없음, 변이 거부 줄 45개를 요구. M1에 로그 형식 | acceptance AC-007; plan M1 |
| N12 모집단·선호 배출 | §A.3에 운영자 대화로 남는 비예약 행과 선호 출처 문장, REQ-BGS-020 신설, 앵커 A45, G-1 | spec §A.3, REQ-020, G-1; design §D.3 |
| N13 한 곳에 한 사실 | 예약 목록은 REQ-010만 갖고 AC-004·design D.5·Q6·research R4.5는 포인터/원문 인용. 공지 문장의 모양은 design D.7만. 처분표는 plan §I·§J만, HISTORY 0.3.0 행은 요약. spec-compact는 REQ마다 한 줄 요지로 다시 씀 | 각 파일 |
| N14 공지 이름 미점검 | REQ-016이 이름 토큰 `batch gate summary`를 네 로케일에 그대로 둔다고 하고 AC-015가 토큰 `:4`와 변이를 점검. Q10에 반영 | spec REQ-016; acceptance AC-015; decision-index Q10 |
| N15 기록 서식 | 기록은 §10 세 필드 뒤 `counter_refs=`, 값은 공백 없는 토큰, `counter_refs=none`에는 `searched=<토큰>`이 보고서와 기록 양쪽에 필요. 앵커 A16–A18, M-3 갱신 | spec REQ-008·009, §B.6; design §D.3 |

같은 부류를 다른 곳에서도 찾은 결과(형제 스윕):
- S1 (N1 부류) 모든 마일스톤 종료 조건을 점검했다. M1·M2·M4·M5·M3의 `Exit:`는 앞 마일스톤의 산출물에만 기댄다. §E의 "기준선 증거(M0)" 불릿에도 조상 판정이 끼어 있어 M0 증거와 끝 점검으로 갈랐다.
- S2 (N11 부류) AC-015도 `--- PASS` 줄 수만 셌다(빈 하위 테스트 변이). 같은 비공허 점검(`=== RUN` 일치, `[no tests to run]` 없음, 변이 거부 줄)을 적용했다.
- S3 (N5 부류) 리터럴 SHA를 변경 범위의 왼쪽 끝으로 쓴 곳: AC-016 (c)(d), plan 제약 4, plan §E 불릿, RED-7 7c·7d, spec-compact AC-016. 모두 읽는 시점 merge-base로 맞췄다. 남은 리터럴 SHA(HEAD 핀, 계획 시점 측정 표)는 측정 트리 핀이며 변경 범위 판정의 왼쪽 끝이 아니다.
- S4 (N2 부류) "게이트 행"을 일반으로 쓴 문장: spec §A.3·B.1, design D.1·D.2, plan 요약·M2 개요·적용 방식, acceptance AC-002·006, spec-compact. 모두 Kickoff 범위로 맞췄다. `factory decide --gate kickoff` 적용 문장은 factory kickoff 행이 개별 질문이라는 R1과 충돌해 삭제했다.
- S5 (N13 부류) 같은 목록의 되풀이: keep-set 세 범주도 REQ-010 밖(design D.5, AC-004)에 있었다. 이름 레이블만 남기고 포인터로 바꿨다.
- S6 (N6 부류) 열린 행의 읽기가 요구사항 문장에서 반쯤 결정된 곳: Q3가 유일한 불일치였다. Q1·Q2·Q7·Q10·Q12는 각 요구사항에 draft reading 표지가 있고 열린 행이 같은 사실을 적는다.
