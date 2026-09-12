# SPEC Review Report: SPEC-MOAI-GATEWAY-001 — AS4 및 API 출력 정책 변경분

Iteration: 3/3 (B2 완료 게이트 명확화 확인)
Verdict: PASS
Overall Score: 0.9375

## Claim — 3차 현재 판정

Reasoning context ignored per M1 Context Isolation.

이번 PASS의 범위는 기존 B2 발견의 완료 의무·연구 의존성·다른 구현의 진행 경계를 명확히 한 변경분이다. native fork의 기술 연결 방법이나 제품 실행을 검증 완료했다는 뜻이 아니다. 전체 AS4 제품 완료는 계속 보류된다. 2차에서 해결된 B1/MP6/MP7의 판정을 다시 수행한 것으로 표시하지 않는다.

spec.md:570~573은 native Agent fork의 부모 이력·분기 위치·격리를 다시 명시적 양성 의무로 두고, 가용성 전제 실패를 NOT-RUN으로 기록하며 실제 증거 전 AS4 및 전체 완료를 금지한다. plan.md:35~40은 실제 설치 가용성 확인→가용한 경로의 양성/음성 증거→전체 AS-013 판정 순서를 명시한다. 독립 가능한 resume·모델·압축·일반 자식·명시 세션 분기는 그 연구와 병행한다. 이로써 기존 요구 삭제를 통한 PASS 경로가 닫혔다.

## Regression Check — 3차

| 발견 | 현재 상태 | 근거 |
|---|---|---|
| AS4-B2 | RESOLVED — 계획의 의무 유지 및 연구 게이트 처분 | spec.md:560,570~573; design.md:508~514; plan.md:35~40; acceptance.md:611~616가 native fork 양성·음성 의무와 전체 완료 보류를 일치시킨다. 기술 구현 및 가용성은 별도의 명시적 미완료 게이트로 유지된다. |
| AS4-B1 | 2차 RESOLVED 유지 | 이번 변경은 B2 의무·게이트 명확화에 한정된다. 압축 설계의 새로운 제품 PASS를 주장하지 않는다. |
| AS4-MP6 / AS4-MP7 | 2차 RESOLVED 유지 | 이번에는 별도 전수 재검증 없이 기존 수정 판정과 현재 lint를 기록한다. |

## Must-Pass Results — 3차 범위

이번에는 이전 발견 B2 및 그 직접 회귀만 확인했다. MP-1/3의 새로운 번호·frontmatter 변경은 없고 lint No findings이며, 요구사항 층에 `The gateway shall`이 명시되어 MP-2 변경분은 충족한다(spec.md:570). MP-4는 동일 단일 언어 범위로 N/A다. MP-5/6/7은 2차 판정에서 새로운 변경이 없는 범위이며 전수 재실행 PASS로 표시하지 않는다. MP-8은 이번 B2 변경분에 release-blocking RED-now 셀이 없어 N/A다. 이전 B2의 의무 삭제 여부를 확인하는 변경분 밖의 미관측 항목은 아래 Gaps로 남긴다.

## Category Scores — 3차 범위

| Dimension | Score | Rubric Band | Evidence |
|---|---:|---:|---|
| Clarity | 1.00 | 1.00 | native fork 의무와 NOT-RUN/완료 보류/독립 구현 진행이 명확하다(spec.md:570~573). |
| Completeness | 0.75 | 0.75 | 구현 가능한 경로와 native fork 가용성·연결 실증 게이트를 분리한 진행 계획이 있다(plan.md:35~40). native fork 기술 설계 자체는 아직 연구 게이트 밖으로 나오지 않았다. |
| Testability | 1.00 | 1.00 | 실제 분기 전 사실·분기 후 혼선·계정/귀속 거절 및 NOT-RUN 판정이 명시됐다(acceptance.md:611~616). |
| Traceability | 1.00 | 1.00 | REQ-MG-015의 복원된 native fork 의무와 AS-013의 추가 시나리오가 연결된다(spec.md:570~573, acceptance.md:603,611~616). |

## Defects Found — 3차 현재 목록

기존 B2의 완료 의무 유지·게이트 처분 변경분에서 미해결 문서 결함을 발견하지 않았다. native fork 실행 가능성·정확한 부모 결속은 미확인 연구/제품 게이트이며 PASS가 아니다. 해당 게이트가 해소될 때까지 AS4 및 전체 지원 완료를 주장할 수 없다.

## Evidence — 3차

```text
go run ./cmd/moai spec lint SPEC-MOAI-GATEWAY-001
✓ No findings — all SPEC documents are valid
exit: 0

git rev-parse HEAD
81c1d58f9cf7045594ee61d5e4ff380948ce9eba
```

`nl -ba`와 `sed`로 spec.md:568~575, plan.md:29~42, design.md:508~516, acceptance.md:609~619를 직접 읽었다. 핵심 원문:

```text
spec.md:570 The gateway shall native Agent fork의 부모 이력 상속·분기 위치 보존·자식 격리를 양성 기능 의무로 유지해야 한다.
spec.md:572 해당 양성·음성 증거를 확보하기 전 AS4 및 전체 지원 완료를 보류해야 한다. 일반 자식이나 명시 세션 분기의 성공으로
spec.md:573 이를 대체해서는 안 된다. 독립적으로 구현 가능한 다른 AS4 경로는 계속 진행한다.
plan.md:39 순서는 가용성 확인과 독립적으로 진행 가능한 resume·모델·압축·일반 자식·명시 세션 분기 구현을 병행하고,
plan.md:40 마지막 완료 판정에서 native fork 증거를 포함해 AS-013 전체를 확인하는 것이다.
acceptance.md:614 설치본에서 native fork를 찾지 못하면 전제 실패 NOT-RUN으로 기록하고 AS4 및 전체 지원 완료를 보류한다.
acceptance.md:615 이 양성 의무는 유지되며 일반 non-fork 또는 --fork-session 성공으로 대체하지 않는다. 가용성 조사 및 native fork
```

## Baseline-attribution — 3차

동일 legacy WT `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`, HEAD `81c1d58f9cf7045594ee61d5e4ff380948ce9eba`. 다음 명령의 출력으로 실제 dirty 문서를 고정했다.

```text
shasum -a 256 .moai/specs/SPEC-MOAI-GATEWAY-001/{spec,design,plan,acceptance}.md
ccdbdd091cd2ee26292725abc6b1220c25f9a6385046bf27efd196188b48e086  .moai/specs/SPEC-MOAI-GATEWAY-001/spec.md
db1c4a8c455ba061e7d8047d1021ed59b52c439e5afa9ec23ac414acc55eab03  .moai/specs/SPEC-MOAI-GATEWAY-001/design.md
58b509852be7bdcaf1c0ee34d07c11e8f866d9c49d19fc2ca314e1344a45e0a3  .moai/specs/SPEC-MOAI-GATEWAY-001/plan.md
82231badae807bbbb18f1a65935ff0f6ccfb8f0f18bcae4d4b12482572b3ac1c  .moai/specs/SPEC-MOAI-GATEWAY-001/acceptance.md
```

## Gaps — 3차

native fork 가용성·부모 결속·실제 분기/재개를 실행하지 않았다. 다른 AS4 구현과 제품 E2E, Windows 실행도 이번 감사 대상이 아니다. 기존 probe 및 공식 문서를 다시 실행/열람한 것으로 표시하지 않는다. 전체 SPEC 재감사는 수행하지 않았으며 현재 PASS는 지정된 발견의 수정 확인이다.

## Residual-risk — 3차

미해결 기능 연구가 남아 있어 다른 AS4 경로를 구현해도 AS4 전체를 닫을 수 없다. 구현 중 새로운 native 표면을 채택할 때 정확한 계정·부모·분기 경계가 입증되어야 한다. 단순 NOT-RUN 표시는 안전한 완료 보류이며 기능 지원을 대신하지 않는다.

## Recommendation — 3차

독립 구현 가능한 AS4 경로는 현재 계획대로 진행한다. native fork는 별도 가용성·연결 실증 게이트로 유지하고 그 양성/음성 증거 전 전체 완료를 보류한다. 이번 네 발견의 감사는 여기서 동결하며 추가 범위로 반복하지 않는다.

## Iteration history — 최종

| Iteration | Verdict | Score | 처리 |
|---|---|---:|---|
| 1 | FAIL | 0.8125 | B1/B2/MP6/MP7 |
| 2 | FAIL | 0.875 | B1/MP6/MP7 해결, B2 양성 의무 유지 불명확 |
| 3 | PASS — 변경분 계획 판정 | 0.9375 | B2 양성 의무·전체 완료 보류·독립 구현 진행 경계 명확화; native fork 제품 게이트는 미완료 |

---

# 보존된 2차 감사

Iteration: 2/3 (기존 네 발견 및 직접 회귀)
Verdict: FAIL
Overall Score: 0.875

## Claim — 2차 현재 판정

Reasoning context ignored per M1 Context Isolation.

이 절이 현재 판정이며 아래 1차 본문은 이력이다. AS4-B1, AS4-MP6, AS4-MP7은 해결되었다. AS4-B2는 일반 non-fork 자식 및 launcher 명시 분기의 설계가 구체화되었으나, 기존 native Agent fork 양성 의무를 별도 Gap으로 옮겨 전체 완료 게이트에서 유지하는지 불명확하다. 따라서 이전 발견이 완전히 해결됐다고 판정하지 않는다. 점수 상승이 미해결 발견을 상쇄하지 않는다.

## Regression Check — 2차

| 이전 발견 | 판정 | 현재 근거 |
|---|---|---|
| AS4-B1 압축 HTTP 사전 결속 | RESOLVED | design.md:461~484는 정상 요약 turn→실제 반환 요약→인증 PostCompact exact digest/scope/epoch→검증 wrapper rebase를 정의한다. 사전 HTTP 분류가 필요 없어졌다. prefix 정규화·완료 요청 재시도·통지 유실·변조 실패가 명시되었고 acceptance.md:592~601은 실제 요약·회상·새 프로세스·ToolSearch 양성과 음성을 유지한다. 보관된 실제 Claude 3회 probe의 비교 JSON을 직접 읽어 구조 근거를 확인했다. 제품 PASS는 아니다. |
| AS4-B2 부모 및 분기 결속 | PARTIALLY RESOLVED / UNRESOLVED | design.md:495~506과 acceptance.md:603~609는 일반 자식의 독립 thread와 launcher의 원본 family/exact prefix/completedTurnID/lastTurnId를 구체화한다. 그러나 기존 REQ의 Claude Agent fork는 현재 spec.md:560의 독립 대화·명시 세션 fork로 바뀌고, native fork는 design.md:508~511 및 acceptance.md:610의 별도 실증 Gap이 됐다. 원래 양성 의무가 전체 완료 조건에 남는다는 문장이 없다. |
| AS4-MP7 native cap marker | RESOLVED | plan.md:563~574가 공식 모델 자료, 명목값·계정 실측·UI 기준의 구분, 후속 실제 검증을 명시한다. 현재 plan/research marker 추출 결과 []. 해당 공식 Claude/GLM 모델 자료를 이번 감사에서 직접 열어 수치를 대조했다. |
| AS4-MP6 syscall 플랫폼 경계 | RESOLVED | spec.md:489~490에 !windows/windows build 제약이 추가되었으며 실제 internal/cli/launch_exec_posix.go 및 launch_exec_windows.go 첫 줄과 일치한다. |

## Must-Pass Results — 2차

- MP-1, MP-2, MP-3: 변경분에서 번호·GEARS 의무·frontmatter를 유지한다. spec.md:552~569의 When/Where/The gateway shall 의무와 현재 lint No findings를 확인했다. 1차의 전체 번호 추출 결과를 새로운 전수 측정으로 재표시하지 않는다.
- MP-4: N/A — 동일 Go gateway 범위(spec.md:11).
- MP-5: 변경분에 퇴역 SPEC의 새 의존성이 없다. 현재 조항의 직접 회귀만 확인했으며 1차 참조 전체를 재검사한 주장은 아니다.
- MP-6: PASS — spec.md:489~490 및 실제 플랫폼 파일 첫 줄 대조.
- MP-7: PASS — plan.md/research.md의 marker 목록 [].
- MP-8: N/A — 이번 변경분 AS-010~013은 release-blocking RED-now 분류가 없다. 미래 제품 E2E를 수행한 것처럼 취급하지 않았다.

## Category Scores — 2차

| Dimension | Score | Rubric Band | Evidence |
|---|---:|---:|---|
| Clarity | 0.75 | 0.75 | 압축 처리 규약은 구체화(design.md:471~484). native fork 완료 지위는 불명확(design.md:508~511). |
| Completeness | 0.75 | 0.75 | 사전 압축 분류 의존 해소와 명시 세션 분기 원장이 추가됨(design.md:461~484,501~506). 기존 native fork 양성 의무 처분은 남음. |
| Testability | 1.00 | 1.00 | 현재 AS-012/013의 실제 요약·회상·정확 분기 위치·재시도/변조 판정은 이진 검증 가능(acceptance.md:592~610). |
| Traceability | 1.00 | 1.00 | 변경 AS-012/013이 REQ-MG-015/017을 명시하며 현재 두 REQ가 존재한다(acceptance.md:592,603; spec.md:552,582). |

## Defects Found — 2차 현재 목록

D2. **AS4-B2 — 기존 native Agent fork의 양성 완료 의무가 대체됨** — spec.md:560; design.md:508~511; acceptance.md:603~610 — Severity: major — Class: blocking.

일반 non-fork 자식의 독립 맥락을 분기 이력 상속과 구별한 것은 공식 문서에 부합한다. 명시 --fork-session에서 원본 family 및 lastTurnId를 사용하는 설계도 별개의 양성 경로로 구체화됐다. 하지만 1차 spec.md:557의 “Claude Agent의 fork”와 AS-013:604~611의 실제 Agent fork는 현재 요구사항에서 독립 자식·명시 세션 분기로 바뀌었다. native Agent(fork)/subtask는 별도 실증 Gap이라고만 기재돼 기존 전체 완료 게이트가 유지되는지 확인할 수 없다.

빈 프로필 Claude 2.1.269의 `Agent type 'fork' not found` 기록은 그 조건에서 실증이 막혔다는 증거다. 전체 native fork 기능이 없다는 증거나 기존 지원 의무의 폐기 근거가 아니다. 공식 문서도 non-fork와 부모 이력 상속 fork를 별개로 설명한다. [공식 subagents 문서](https://code.claude.com/docs/en/sub-agents#what-loads-at-startup).

**Required fix:** native Agent fork의 기존 양성 의무와 전체 완료 Gate를 명확히 유지한다. 현재 실증 차단은 별도 가용성 Gap으로 남기되, 가용성을 확인할 단계·양성/음성 판정·완료 보류 조건을 기재한다. 또는 기능 지원 범위를 실제로 변경하기로 별도로 결정했다면 그 범위 변경과 남는 기능 제한을 명시해야 하며, 이를 이전 발견의 무조건 해소로 표시하지 않는다. 이 발견은 일반 자식/명시 세션 분기의 구현을 중단하라는 요구가 아니다.

## Evidence — 2차 현재 실행

```text
moai session current
cf1d7aa8-9d2f-491f-a48c-27e8d717276a

git rev-parse HEAD
81c1d58f9cf7045594ee61d5e4ff380948ce9eba
git branch --show-current
WT-unified-gateway

go run ./cmd/moai spec lint SPEC-MOAI-GATEWAY-001
✓ No findings — all SPEC documents are valid
exit: 0

head -8 internal/cli/launch_exec_posix.go
//go:build !windows
[이후 package/import 출력 생략]
head -8 internal/cli/launch_exec_windows.go
//go:build windows
[이후 package/import 출력 생략]
```

Python으로 plan/research의 `[NEEDS CLARIFICATION` 포함 행과 설치 ThreadForkParams의 lastTurnId를 직접 추출한 실제 출력:

```text
clarification markers: []
lastTurnId: {'description': 'Optional last turn id to fork through, inclusive.\n\nWhen specified, turns after `last_turn_id` are omitted from the fork. The referenced turn cannot be in progress.', 'type': ['string', 'null']}
```

`cat .moai/reports/t649/appserver-redesign/probe/compact-history-shape-comparison.json`에서 seed_user_canonical_equal, seed_system_text_equal, top_system_canonical_equal, compact_assistant_text_exact, compact_added_user_instruction, summary_hash_matches_PostCompact, followup_old_assistant_absent가 모두 true임을 읽었다. summary SHA-256은 `6fb58f445a90cbc893e201b16fcf8c876cff5f92937e12dead91ed11ebe6463f`, 새 질의 위치는 `messages[0].content[5].text`였다. 이 기록을 재실행한 결과로 표현하지 않는다.

이번 web open/find로 [Claude 모델 표](https://platform.claude.com/docs/en/models/overview)의 Opus/Sonnet 1M·Haiku 200K와 [GLM-5.1](https://docs.z.ai/guides/llm/glm-5.1), [GLM-4.7](https://docs.z.ai/guides/llm/glm-4.7)의 200K context·128K output·text 입력/출력을 확인했다. 명목 문서 수치 대조이며 계정 한도 실험이 아니다.

## Baseline-attribution — 2차

동일 legacy WT / WT-unified-gateway / HEAD 81c1d58f9cf7045594ee61d5e4ff380948ce9eba. 실제 파일 hash:

```text
4069787de0dbc4f69d9cccfd21964a25589084215c0f65a18750f26784d6fbba  spec.md
3ab6665a7467dd79c1b5aad9c7e991ed0bc4b1a59f7cfd4f9bddf09e74981047  design.md
6d46a7cbc320c947bd3a3dd3238880a59adcd2a83d2195c42dd60e1ff1d06395  plan.md
c1ca7267fef3acfd7fa0edad2c0fc9e61bd09c52b57431c2b8854bdfb175c5d5  acceptance.md
```

source_session_id는 개발자 지정 `01a08e7b-6aa0-7361-ab7e-ea8da1f02228`을 유지한다. 이번 하위 환경의 session current가 다른 UUID를 출력한 사실을 숨기거나 같은 값으로 정정하지 않았다.

## Gaps — 2차

probe 재실행, 실제 GPT 요약/회상, 모든 압축 모드, native fork 가용성/부모 결속, 제품 명시 --fork-session, Windows runtime을 실행하지 않았다. 동일 계정의 실제 모델 한도도 측정하지 않았다. 이번 감사는 이전 네 발견의 설계 수정과 직접 회귀만 대상으로 하며 AS3 코드는 읽거나 고치지 않았다.

## Residual-risk — 2차

요약 wrapper의 설치본별 파싱과 PostCompact IPC 인증·원장 갱신의 정확성은 제품 코드/E2E에서 판정해야 한다. prefix 원장으로 지정 시점의 세션 분기를 설계할 수 있다는 사실과 native Agent fork가 실제 작동한다는 사실은 다르다. 현재 문서가 후자를 완료 Gate에서 제거한 것인지 모호한 상태로 전체 기능 완료를 주장할 수 없다.

## Iteration history — 현재

| Iteration | Verdict | Score | 처리 |
|---|---|---:|---|
| 1 | FAIL | 0.8125 | B1/B2/MP6/MP7 |
| 2 | FAIL | 0.875 | B1/MP6/MP7 해결, B2 부분 해결·완료 의무 유지 불명확 |

---

# 보존된 1차 감사 및 당시 대안 검토

Iteration: 1/3 (AS4 변경분 감사)
Verdict: FAIL
Overall Score: 0.8125

## Claim

Reasoning context ignored per M1 Context Isolation.

대상은 SPEC 0.11.0의 AS4 재개·모델 변경·압축·fork 및 구독/API 출력 정책 변경분이다. 작성자의 판단이나 이전 감사 PASS를 판정 근거로 쓰지 않았다. spec.md, plan.md, acceptance.md, design.md, research.md에서 관련 조항과 교차 참조를 읽었다. 전체 과거 요구사항의 의미를 다시 인증하는 감사나 AS3 코드 감사가 아니다.

읽기 전에 설정한 실패 가능성은 번호·frontmatter·추적 단절, 생성 토큰 상한의 과장, 압축 예약의 잘못된 HTTP 소비, 빈 RPC 응답을 완료로 오인, 완료 응답 유실 후 중복 압축, 압축 후 입력 중복·이력 유실, 부모와 자식의 잘못된 결속, 재개 계정 혼선, 미지원 모델의 묵시적 대체였다.

API 정책 변경 자체는 문서 간 일치한다. 그러나 AS4의 압축 HTTP 결속과 실제 Claude 부모 fork를 구현할 연결 수단이 아직 정해지지 않았다. 명시 실패 규칙은 안전한 거절을 정의하지만 요구된 양성 기능의 구현 경로를 대신하지 않는다. 별도로 현행 문서의 MP-7과 MP-6 필수 조건 위반을 확인했다. 이 둘은 AS4가 새로 만든 결함이라고 주장하지 않는다.

## Must-Pass Results

| 항목 | 판정 | 근거 |
|---|---|---|
| MP-1 번호 | PASS | 현재 spec의 REQ-MG-001~026 정의를 추출했다. 중복·누락 0이며 퇴역 번호도 정의를 유지한다. 아래 명령 출력 참조. |
| MP-2 GEARS | PASS, 변경분 한정 | REQ-MG-015의 `Where GPT 대화가 App Server로 이어지는 경우, the gateway shall` 및 `When 인증된 압축 요청이 완료되면, the gateway shall`(spec.md:551,560), REQ-MG-017의 `The GPT adapter shall`(578)과 `The adapter shall`(605)이 요구사항 층의 명시적 의무다. AC의 Given/When/Then은 올바른 검증 층이다. 미변경 전체 REQ의 재인증은 범위 밖이다. |
| MP-3 frontmatter | PASS | spec.md:1~15의 12개 필수 필드와 tier:L. 현재 트리 lint 출력이 No findings이며 canonical schema의 필드 표와 대조했다. |
| MP-4 언어 중립 | N/A | Go 저장소의 gateway 구현 계약이다(spec.md:11). 범용 다언어 템플릿 변경이 아니다. |
| MP-5 D7 | PASS | 추출된 실제 참조 5개는 모두 draft이며 retired/superseded/archived 참조가 없다. SPEC-MOAI-PROXY-001은 존재하지 않는 역사 참조다. 아래 발견 목록 참조. |
| MP-6 D8 | FAIL | 현재 의무 spec.md:488 및 제약 808에 syscall.Exec가 있으나 literal //go:build 제약이나 명시적 syscall 면제 조항이 없다. POSIX 설명만으로 D8의 정해진 형식 조건을 충족하지 않는다. AS4 외 기존 필수 조건 위반, D4. |
| MP-7 clarification | FAIL | plan.md:563의 `[NEEDS CLARIFICATION: native Claude·GLM cap 근거]`가 남아 있다. 다음 줄의 “다른 구현을 멈추지 않고” 문구는 필수 clarification gate의 예외가 아니다. AS4 외 기존 위반, D3. |
| MP-8 RED-now | N/A, 이번 변경분 | AS-010~013 및 API 정책 시나리오에는 release-blocking 분류나 RED-now 셀이 없다. 기존 Windows release PR의 미래 실행 게이트를 이 변경분의 RED-now 셀로 재해석하지 않았다. 테스트 셀 재실행 PASS를 주장하지 않는다. |

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|---|---:|---:|---|
| Clarity | 0.75 | 0.75 | 출력 정책과 완료/실패는 명료하다(spec.md:605~608, design.md:470~479). 그러나 “유일하게 결속”의 판별 기준이 비어 있다(design.md:468). |
| Completeness | 0.50 | 0.50 | 재개·모델 변경·압축 결과의 계약과 다섯 문서가 있으나 실제 HTTP의 압축 결속 및 부모 분기 결속이라는 두 핵심 설계 입력이 없다(design.md:466~469,494~500). |
| Testability | 1.00 | 1.00 | AS-010~013의 사실 회상, 선택 model 일치, RPC 횟수, 변조 거절, 형제 격리 조건은 이진 판정 가능하다(acceptance.md:584~611). 테스트의 존재·통과를 주장하는 점수가 아니다. |
| Traceability | 1.00 | 1.00 | AS-010~013은 유효한 REQ-MG-015/017에 연결되며, API 정책은 REQ-MG-017과 acceptance.md:491~495 및 t650/t654 판정에 연결된다. 전체 AC 참조 추출에서 존재하지 않는 REQ 참조 0. |

## Defects Found

D1. **AS4-B1 — 압축 예약과 특정 HTTP의 연결 방식 부재** — design.md:466~469; acceptance.md:594~600 — Severity: major — Class: blocking.

인증된 PreCompact는 계정/family/agent/thread/epoch를 고정하지만 이후 HTTP가 바로 그 압축 요청임을 어떻게 확인할지 문서가 정하지 않는다. idle, 같은 범위의 singleflight, 이미 반영한 prefix를 고정하는 것만으로 요청 종류를 확인했다는 근거도 없다. 일반 요청이 예약 뒤 먼저 도착하거나 이전 HTTP가 재시도되는 경우, 둘 다 같은 인증·범위·이력 조건을 가질 수 있는지는 설계에서 판별해야 할 경합이다. 이는 관측된 런타임 오동작 주장이 아니라 현재 문서의 누락이다.

기존 probe는 PreCompact→HTTP 순서만 기록한다. hook 필드에 HTTP별 nonce/digest가 없고, 제품 결속 규약을 시험하지 않았다. “해당 HTTP와 유일하게 결속할 수 없으면 거절”만 구현하면 AS-012 양성 경로 전체가 영구 거절이어도 그 문장 자체는 충족한다.

**Required fix:** 실제 Claude의 제어 표면과 gateway 사이에서 예약이 특정 압축 HTTP를 식별하는 프로토콜을 design에 정한다. 식별 정보의 생산자·전달 경로·검증자·소비 시점, 일반 요청/재시도/다른 scope/만료 경쟁 시 처리, 응답 유실 뒤 재요청의 동일성 판정을 포함한다. 같은 scope의 idle/singleflight/prefix만으로 충분하다고 판단한다면 그 불변식이 실제 Claude에서 성립하는 관측과 반례 대조를 제시한다. 프롬프트 문구·도착 순서 추측으로 대체하거나 압축 양성 요구를 삭제하지 않는다.

D2. **AS4-B2 — 실제 Claude fork의 부모 및 분기 위치를 공급하는 경로 부재** — design.md:494~500; plan.md:32~34; acceptance.md:607~611 — Severity: major — Class: blocking.

현재 계획은 HTTP agent-id와 CLI parent_tool_use_id의 직접 결속을 확보하지 못했다고 적고, 명시적으로 인증된 부모 thread와 분기 경계가 “있는” 경우만 정의한다. 그러나 실제 Claude Agent에서 그 정보를 확보하는 제어 경로 또는 구현 전 검증 절차가 없다. 병렬 자식 ID가 안정적인 사실은 서로 다른 범위를 나누는 근거이며 정확한 부모·중첩 부모 및 분기 시점의 근거가 아니다. 따라서 정상 resume와 독립 자식 구현을 할 수 있더라도 요구된 실제 fork 완료까지의 설계가 닫히지 않는다.

**Required fix:** 실제 Claude Agent의 부모 도구 호출과 자식 HTTP, 부모 thread의 분기 경계를 인증된 방식으로 연결하는 표면을 확인하고 설계에 기록한다. 병렬·중첩·부모가 이미 다음 turn으로 이동한 상황을 포함한 양성/음성 절차를 둔다. 확인 전에는 해당 연구 의존성을 구현 착수 전의 명시 게이트로 배치하며, “독립 자식만 성공” 또는 “모호하면 모두 오류”로 fork 양성 요구를 통과시키지 않는다. 요구사항 축소를 해결책으로 삼지 않는다.

D3. **AS4-MP7 — 미해결 clarification marker** — plan.md:563~565 — Severity: critical — Class: blocking — Required fix: native Claude/GLM cap의 현재 경로 근거를 확인해 해결 내용을 기록하고 marker를 해결한다. 단순 이름 변경·삭제로 연구 의무를 없애지 않는다. 사용자 선호 질문이 아닌 기계적 조사라고 적은 사실은 확인했으며, 이 감사는 불필요한 재승인을 요구하지 않는다.

D4. **AS4-MP6 — 활성 syscall 의무의 D8 형식 누락** — spec.md:488,808 — Severity: critical — Class: blocking — Required fix: 기존 POSIX/Windows 구현의 플랫폼 분리 근거를 확인한 후 현재 해당 의무에 실제 //go:build 제약 또는 명시적 cross-platform exemption을 기재한다. 과거 HISTORY 일괄 수정은 요구하지 않는다. Windows 런타임 결함이나 빌드 실패가 관측됐다는 뜻은 아니다.

Optional: SPEC-MOAI-PROXY-001 참조 대상은 현행 디렉터리에 없으나 명칭 대응 및 과거 이력(spec.md:399 이하)으로 남은 참조다. 현재 AS4 기능의 의존성으로 판정하지 않았다.

## Evidence

현재 실행한 명령과 관측 출력이다. probe 결과 파일을 읽은 것과 probe를 이 감사에서 새로 실행한 것을 구분한다.

```text
moai session current
01a08e7b-6aa0-7361-ab7e-ea8da1f02228

git rev-parse --show-toplevel
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified
git rev-parse HEAD
81c1d58f9cf7045594ee61d5e4ff380948ce9eba

go run ./cmd/moai spec lint SPEC-MOAI-GATEWAY-001
✓ No findings — all SPEC documents are valid
exit: 0

rg -n 'NEEDS CLARIFICATION|release-blocking|RED-now' .moai/specs/SPEC-MOAI-GATEWAY-001/{plan,research,acceptance}.md
.moai/specs/SPEC-MOAI-GATEWAY-001/plan.md:563:[NEEDS CLARIFICATION: native Claude·GLM cap 근거] 실제 사용 경로의 기본 context와 계량 의미를 공식/현재 native

git check-ignore .moai/reports/t649/appserver-redesign/as4-plan-audit.md
[stdout empty]
exit: 1
```

구조 검사는 Python 표준 라이브러리로 `spec.md`의 `^\*\*(REQ-MG-\d+)\*\*` 정의를 추출하고 1~26과 대조했으며 acceptance.md의 REQ 참조 차집합, SPEC 참조 파일 존재·status를 출력했다. 관측 출력:

```text
REQ headers: REQ-MG-001,REQ-MG-002,REQ-MG-003,REQ-MG-004,REQ-MG-005,REQ-MG-006,REQ-MG-007,REQ-MG-008,REQ-MG-009,REQ-MG-010,REQ-MG-011,REQ-MG-012,REQ-MG-013,REQ-MG-014,REQ-MG-015,REQ-MG-016,REQ-MG-017,REQ-MG-018,REQ-MG-019,REQ-MG-020,REQ-MG-021,REQ-MG-022,REQ-MG-023,REQ-MG-024,REQ-MG-025,REQ-MG-026
REQ duplicate headers: []
REQ missing numbers: []
AC dangling references: []
release-blocking literal count: 0
SPEC-MOAI-CG-RETIRE-001 draft
SPEC-MOAI-GATEWAY-001 draft
SPEC-MOAI-GATEWAY-PICKER-001 draft
SPEC-MOAI-GATEWAY-TEAMMATE-001 draft
SPEC-MOAI-GPT-AUTH-001 draft
SPEC-MOAI-PROXY-001 ABSENT
spec syscall lines: [88, 90, 149, 151, 488, 808]
```

`compact-hook-result.json` 및 설치 schema JSON을 Python `json.loads`로 읽어 실제 보관 필드·순서를 확인한 출력:

```text
compact success: True
compact invocation exits: [0, 0]
compact hook field sets: [['hook_event_name', 'kind', 'session_id', 'time_ns', 'transcript_path_present', 'trigger'], ['hook_event_name', 'kind', 'session_id', 'source', 'time_ns', 'transcript_path_present'], ['hook_event_name', 'kind', 'session_id', 'time_ns', 'transcript_path_present', 'trigger']]
compact events: [('http_request', None, None, None), ('hook', 'PreCompact', 'manual', None), ('http_request', None, None, None), ('hook', 'SessionStart', None, 'compact'), ('hook', 'PostCompact', 'manual', None)]
compact params: ['threadId']
compact result fields: []
```

이는 보관된 probe의 관측 범위 확인이다. 이번 감사에서 실제 GPT/Claude 압축을 재실행한 증거가 아니다. `agent-identity-parallel-verdict.md`의 8개 요청 관측 및 정확한 parent_tool_use_id 결속 미확보 설명도 읽었다. 공식 hook 문서 보관본 3015~3046행은 PreCompact의 trigger/custom_instructions 입력과 차단 출력을 설명하며 HTTP 요청별 식별자 전달 근거를 제공하지 않는다. 문서 표면이 부족하다는 결론을 “다른 표면도 존재하지 않는다”로 확장하지 않았다.

## Baseline-attribution

측정 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`

HEAD: `81c1d58f9cf7045594ee61d5e4ff380948ce9eba`. dirty SPEC의 내용은 HEAD만으로 고정되지 않으므로 아래 실제 SHA-256을 함께 기록한다. 작성 메모의 판단은 사용하지 않았으며 그 변경 후 hash와 현재 읽은 파일 hash의 일치만 대조했다.

```text
shasum -a 256 .moai/specs/SPEC-MOAI-GATEWAY-001/{spec,plan,acceptance,design,research}.md
dc5b33a83e080b12173fc8a9c12a045c7c3a4a71bcae1873752c6480bc31401a  .moai/specs/SPEC-MOAI-GATEWAY-001/spec.md
07c03c8fef3d5b07cfe54f39f38e9832a0e9370cd34c5cb55c6f8847574ce97b  .moai/specs/SPEC-MOAI-GATEWAY-001/plan.md
7670b8ab80b22ed97cbcd65282a3ffb62ce02919136fe7d12fa0d2f200d0fd08  .moai/specs/SPEC-MOAI-GATEWAY-001/acceptance.md
19596742159e80eff605230e887635cd9a9d0e94e24cb3ed3948dc3cf05d1689  .moai/specs/SPEC-MOAI-GATEWAY-001/design.md
0622a38f0e5f00ca2066bcf1f6f8f4c1f9f8760412d6d07ab4c483af8d0c649d  .moai/specs/SPEC-MOAI-GATEWAY-001/research.md
```

## Gaps

- AS4 제품 구현, 실제 구독/API 생성, 압축 후 회상, interactive/auto/child compact, fork/resume, Windows 실행은 이 감사에서 수행하지 않았다.
- 기존 probe를 재실행하지 않았다. 파일 판독은 그 기록을 현재 관측한 증거이며 최신 제품 성공 증거가 아니다.
- 전체 미변경 REQ의 의미·모든 과거 AC 및 코드 전체를 재감사하지 않았다. 이 파일을 전체 SPEC 재인증 PASS로 사용할 수 없다.
- 외부 모델 교차 감사는 실행하지 않았다. `.moai/config`의 audit_model 검색에서 설정을 찾지 못했다. 현재 판정은 단일 독립 감사다.
- 공식 source 보조 검색 중 존재하지 않는 codex_message_processor.rs 경로 1개에서 rg exit 2가 발생했다. 이 실패를 공식 소스 검증 PASS로 사용하지 않았다.

## Residual-risk

바이트·취소 제한은 API 사용자의 실제 생성 토큰 비용 상한이 아니다. 승인된 문구는 그 차이를 숨기지 않지만 실제 UI와 취소 동작은 제품 검증이 필요하다. 이어쓰기 기준점이 Claude 압축 후 history에 보존되는지, Codex 압축이 합성 사실·도구 결과를 보존하는지도 별도 실증 대상이다. 문서의 명시 실패 원칙은 유실·혼선을 줄일 수 있으나 정상 기능 완료를 자동으로 성립시키지 않는다.

## Recommendation

1. D1과 D2의 연결 수단 및 실제 경로 관측을 먼저 확정한다. 양성 요구는 유지한다.
2. D3의 조사 의존성을 해결하고 D4의 활성 플랫폼 계약 표기를 고친다.
3. 같은 네 발견의 변경분과 회귀만 재감사한다. API 서버 출력 정책 변경은 현재 문구를 유지해도 된다(spec.md:605~608, design.md:505~507, acceptance.md:491~495, plan.md:35~36).

## Iteration history

| Iteration | Scope | Verdict | Score |
|---|---|---|---:|
| 1 | AS4 및 API 출력 정책, 현재 필수 조건 점검 | FAIL | 0.8125 |

## 대안 검토 — 미검증 설계 분석

오케스트레이터가 제안한 대안은 Claude의 압축 요청을 정상 App Server turn으로 처리하여 실제 사실 요약을 Claude에 돌려주고, 인증된 PostCompact로 Claude history의 기준을 다시 맞추는 방식이다. Codex 자체 자동 압축은 공식 runtime 소유로 두며 MoAI가 thread/compact/start를 추가 호출하지 않는다. 이 검토는 대안 승인이나 현재 FAIL의 해소 판정이 아니다.

- **inferred — 계약 대조 규칙:** design.md:469~474 및 AS-012:594~597의 “명시 compact RPC 1회 / 일반 생성 RPC 0회 / 비사실 안내 / 별도 요약 생성 금지”를 대안과 대조하여 변경 목록을 작성한다. 사용자 목표인 Claude UI·공식 App Server·재개·사실 보존은 이 네 구현 선택과 별도로 유지되는지 새 계약에서 판정한다. 구현 방식을 바꾸는 것만으로 기능 요구 축소라고 단정하지 않는다.
- **measured — 보관된 JSON의 구조 판독:** `python3`으로 compact-hook-result.json의 두 requests를 읽은 결과, 첫 요청은 user(text,text)+system(text), 둘째는 user(text,text)+system(string)+assistant(text)+user(string)이었다. 이 관측을 출발점으로 합성 원문을 캡처하여 이미 반영한 prefix와 새 요약 지시를 정확히 분리할 수 있는지 측정한다. 기존 shape는 내용·정규화 동등성의 증거가 아니다.
- **assumption — 후행 결속 후보:** 반환한 공개 요약의 digest와 인증된 PostCompact의 실제 compact_summary, scope/epoch를 연결하여 history rebase를 판정하는 실험을 수행한다. 원문 요약을 임의 명령이나 인증 토큰으로 신뢰하지 않고, 이미 반환한 응답과 일치함을 확인하는 절차를 검증한다. 이 후보는 현재 제품의 구현 사실이 아니다.
- **assumption — 요청 재처리:** PostCompact 지연·중복·유실, HTTP 완료 응답 유실, 다음 일반 요청 선행, 새 프로세스 재개를 주입하여 생성·입력 반영·rebase가 중복되지 않는지 측정한다. 후행 hook을 받기 전 완료/미완료 경계를 새 설계에 명시한다.
- **assumption — 실제 사용자 목표:** 수동·자동·자식 압축에서 실제 사실 요약 표시, 압축 전 사실·도구 결과 회상, 새 프로세스 resume, ToolSearch 후발 도구 사용을 검증한다. native reasoning은 Codex 안에 두고 공개 history를 통째로 재주입하지 않는 조건을 유지한다. 성공하면 사전 HTTP 압축 분류를 요구하지 않는 더 단순한 설계의 근거가 될 수 있으나, 현재는 미검증이다.

이 대안은 D2의 부모 fork 결속 문제를 해결하지 않는다. D3은 AS3 수리 중단 지시가 아니라 기존 전체 SPEC 종료 게이트의 잔여 사항이다.
