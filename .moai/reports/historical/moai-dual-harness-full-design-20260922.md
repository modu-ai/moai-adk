# MoAI-ADK 듀얼 하네스 전체 설계

> 설계 제안 v1.0 · 2026-09-22 · Claude Code + Codex · 구현 완료를 뜻하지 않음

## 01. 결론과 설계 목표

**공통 정책·워크플로·검증·상태는 MoAI가 소유하고, Claude Code와 Codex는 실행 환경으로 연결한다.** 기존 Go 코어와 배포기를 확장한다. 별도 범용 워크플로 엔진이나 새로운 분산 데이터베이스를 만들지 않는다.

목표는 같은 작업에 대해 같은 필수 산출물, 같은 승인 범위, 같은 검증 조건, 같은 완료 판정을 제공하는 것이다. 화면·도구 이름·모델 응답 문장까지 일치시킬 필요는 없다. 단, 기능을 생략한 상태를 동등 지원으로 표시하지 않는다.

- 제품 범위: 현재 MoAI 명령 스킬 17개, 공통 스킬 35개, 전문 에이전트 12개와 설치·업데이트·훅·워크트리·칸반·팩토리·MCP·진단.
- 1차 인증 환경: Claude Code CLI, Codex CLI, 두 CLI를 섞은 팩토리. 각 환경은 별도 통과해야 한다.
- Desktop·Web·원격 실행: 범위에서 삭제하지 않고 별도 프로파일로 관리한다. 로컬 훅·파일 접근이 없는 환경은 동일 기능의 실행 경로가 검증되기 전까지 미인증이다.
- 이 문서의 경로·타입·명령 중 ‘제안’ 표시는 앞으로 구현할 계약이다. 기존 제품 사용법으로 취급하지 않는다.

## 02. 기준과 확인된 문제

| 구분 | 기준과 상태 |
|---|---|
| 감사 기준 | 로컬 develop `7f86971fcb8d289a37880f7279fd5eb7c90f55fa` |
| 설계 시점 HEAD | develop `3f3ffbb57633b5e5aed65ba7696bffd82044f65d`; 작성 중 병합 진행 |
| 변경 확인 범위 | 감사 이후 codexadapter·codexwiring·harness_fs.go·codex_launcher.go의 커밋 차이는 없음. 전체 트리 재감사는 아님 |
| 기존 구현 | 스킬 게시·에이전트 생성·실행기·기본 훅 번역·MCP 설정 및 관련 로컬 테스트 |
| F1 | Codex 전용 배포에 없는 CLAUDE.md·.claude/rules·워크플로 경로를 스킬이 참조 |
| F2 | Codex Stop 생성 설정이 일반 stop 하나이며 Claude의 별도 goal·quality·review·learning 체인과 다름 |
| F3 | 감사 기준 이벤트 12개 중 8개 적응. compact·permission·interrupt 처리 미완료 |
| F4 | Codex launcher의 새 worktree 생성·칸반 진입, 일부 역할 권한·질문·디자인 기능 동등성 미완료 |
| 실운영 Gap | 실제 Codex 연결 handshake 실패 1건; Kanban 검증 권한 제한·timeout; 모델 E2E 전수 미실행 |

앞선 감사의 수치는 그 커밋에서 측정한 값이다. 설치 바이너리나 이후 병합본의 현재 상태로 일반화하지 않는다. 기존 9월 21일 설계의 ‘중립 원본 + 어댑터’ 방향을 계승하되, 배포 참조 완결성과 훅의 기능별 연결을 필수 승인 조건으로 강화한다.

## 03. 동등 지원의 계약

하네스는 모델이 파일·도구·세션을 다루는 실행 환경이다. MoAI의 기능 동등성은 호스트 기능의 개수를 맞추는 대신 아래 계약으로 판정한다.

| 축 | 두 하네스에서 지켜야 할 계약 | 허용되는 차이 |
|---|---|---|
| 정책 | 모든 필수 의무에 적용 위치·강제 장치·검증 항목이 존재 | 지시문 파일 형식 |
| 작업 | 같은 입력·저장소에서 같은 필수 artifact와 상태 전이 | 모델의 설명과 중간 탐색 |
| 권한 | 사용자 승인 범위 밖의 작업과 권한 확대를 거부 | 승인 UI의 모양 |
| 검증 | 같은 증거를 같은 기준으로 평가; 미실행은 성공 아님 | 테스트 실행 도구 |
| 상태 | 카드·SPEC·worktree·dispatch의 소유자와 전이가 일관됨 | 호스트 내부 session ID |
| 복구 | 취소·오류·재시작 후 중복 반영이나 권한 재사용이 없음 | 세션 재개 방식 |

기능 상태는 `native`, `adapted`, `blocked`, `unverified`로 나눈다. native/adapted도 실제 테스트 증거가 있어야 해당 프로파일의 지원 완료로 표시한다. required 기능에 blocked/unverified가 하나라도 있으면 전체 완료 판정은 실패다. 순차 실행으로 결과를 얻는 것은 가능하지만 병렬 실행 동등성까지 통과한 것으로 세지 않는다.

## 04. 전체 구조

공통 원본을 두 실행 환경에 맞춰 생성하고, 실행 시에는 공통 서비스가 정책과 상태를 판단한다. 하네스 어댑터는 입력·출력 및 호출 방식을 변환한다.

```text
중립 카탈로그: policies / workflows / skills / agents / features
                │ 결정적 생성 + 참조 검증
                ├─ 공통 배포: AGENTS.md + .moai/policies + .moai/workflows
                ├─ Claude 배포: CLAUDE.md + .claude/{rules,skills,agents,settings}
                └─ Codex 배포: .agents/skills + .codex/{agents,config,hooks}

Claude adapter ─┐
               ├─ 기존 MoAI Go 서비스 ── 프로젝트 상태·검증 증거
Codex adapter ──┘   hook / goal / gate / worktree / kanban / sessionmsg
```

경계는 네 가지다. 카탈로그는 ‘무엇을 해야 하는가’, 생성기는 ‘어디에 배포하는가’, 어댑터는 ‘어떻게 호출하는가’, 기존 Go 서비스는 ‘허용·진행·완료 여부’를 담당한다. 모델이 임의로 다음 상태를 확정하지 못하게 한다.

## 05. 원본과 배포 디렉터리

**제안 구조.** 현재 원본을 한 번에 삭제하지 않는다. 먼저 동일 산출물을 만드는 생성기로 전환하고, 의미·해시·참조 검사 후 원본 소유권을 이동한다.

```text
internal/template/catalog/            # 중립 원본: 새로 정리
  policies/                          # standing·scoped·host·enforced
  workflows/                         # plan·run·sync 및 명령별 절차
  skills/                            # SKILL 본문·references·scripts
  agents/                            # 본문·역할·required capabilities
  features.yaml                      # ID·핵심 여부·대체 경로·검증 연결
internal/template/agentemit/           # 기존 변환기 재사용
internal/template/templates/          # 생성된 배포 트리, 기존 embed 유지

project/
  AGENTS.md                          # 작고 필수적인 공통 계약
  CLAUDE.md                          # Claude 선택 시에만 생성
  .moai/policies/                    # 양쪽이 읽는 긴 규칙
  .moai/workflows/                   # 양쪽이 읽는 공통 절차
  .moai/config/sections/             # 기존 설정 체계 유지
  .claude/{rules,skills,agents}/      # Claude 선택 시 생성
  .agents/skills/                    # Codex 선택 시 생성
  .codex/{agents,config.toml,hooks.json}
```

`AGENTS.md`에 저장소 전체 문서를 합치지 않는다. 필수 불변식만 직접 싣고 상세 규칙은 명시적으로 읽는 경로를 둔다. Codex 스킬 본문에서 Claude 내부 경로를 자동 추측하지 않게 한다. 스킬의 scripts/references는 패키지 내부 상대경로로, 공통 정책·워크플로는 `.moai/` 절대적 프로젝트 상대경로로 정규화한다.

두 종류의 manifest를 구분한다. 기존 배포 manifest는 파일 소유·해시·업데이트를 담당한다. feature manifest는 의미적 의무·필요 능력·테스트 연결을 담당한다. 같은 정보를 다른 저장소에 중복 저장하지 않는다.

## 06. 정책 전달과 참조 완결성

| 정책 종류 | 공통 원본의 역할 | Claude 전달 | Codex 전달 | 완료 조건 |
|---|---|---|---|---|
| standing | 승인·증거·공유 Git 불변식 | CLAUDE.md에서 공통 계약 연결 | 루트 AGENTS.md | 실제 읽힌 계약·바이트 예산 검사 |
| scoped | 경로별 규칙 | path-scoped rules | 명시적 파일 작업 preflight + 범위 정보 | 동일 대상 파일에 같은 의무 적용 |
| workflow | 작업 절차 | skill → .moai/workflows | skill → .moai/workflows | 참조 대상 존재·읽기 테스트 |
| host | 제품 전용 절차 | Claude 출력에만 | Codex 출력에만 | 다른 하네스에서 호출되지 않음 |
| enforced | 반드시 막아야 할 행위 | CLI/hook/권한 경계 | CLI/hook/권한 경계 | 우회·오류 주입 테스트 |

Codex의 디렉터리 기반 지시문 탐색을 Claude의 파일 glob 적용과 동일하다고 가정하지 않는다. 현재 저장소의 단일 AGENTS 계약을 유지하며 자동 nested AGENTS 생성은 기본안에서 제외한다. 별도 제품 환경에서 도입하려면 규칙 충돌·상속·예산을 독립 검증한다.

참조 검사는 필수 읽기, 선택 기능, 예시, 런타임 생성 경로를 구분한다. 필수 읽기의 미존재·상대경로 탈출·잘못된 fragment는 배포 실패다. 예시 속 경로는 결함 개수에 포함하지 않는다. 모든 기존 의무 ID는 생성 위치와 검증 ID에 연결되어야 한다.

## 07. 기능 탐지와 실행 경로 선택

버전 문자열은 힌트이고 실제 사용 가능성은 탐지 결과다. 프로젝트가 요청한 기능, 호스트 능력, 조직 정책을 교차해 실행 계획을 만든다. 임의의 최신 모델 ID나 지원되지 않는 TOML 키를 추측해서 넣지 않는다.

```yaml
# 제안: features.yaml의 한 항목
id: workflow.goal.continue
required: true
requires: [turn.complete.observe, turn.continue.request]
implementations:
  claude: hook_stop
  codex: hook_stop_or_supervised_session
on_unavailable: block
acceptance: [AC-HOOK-02, AC-GOAL-01]
```

탐지 레코드는 executable 경로·버전·파일 해시, OS, CLI/Desktop/Web 표면, 설정 digest, 탐지 시각, 관측 방식과 결과를 가진다. 바이너리·설정·권한 변경 시 캐시를 무효화한다. 메타데이터 검사와 실제 세션 검사를 분리하고, 모델 호출이 필요한 탐지는 비용과 실행 범위를 표시한다.

네이티브 기능이 없을 때는 이미 검증된 공통 경로만 선택한다. 안전한 대체 경로가 없으면 시작 전에 차단한다. ‘코드 있음’, ‘설정 생성됨’, ‘신뢰됨’, ‘이벤트 발화됨’, ‘효과가 검증됨’을 진단에서 별도로 표시한다.

## 08. 훅을 기능 단위로 공통화

F2의 해법은 이벤트 이름을 추가하는 데서 끝나지 않는다. 현재 Claude 쉘 훅이 수행하는 기능을 목록화하고 같은 Go 처리 함수를 양쪽 등록기가 호출하도록 한다. 기존 `internal/hook`, `internal/codexadapter`, `internal/codexwiring`를 확장한다.

**제안 처리 순서:** 입력 검증 → 세션·worktree 귀속 확인 → 중복 이벤트 확인 → 필수 동기 검증 → 상태 반영 → 선택 관찰. 사용자 취소는 별도 최우선 경로로 처리한다.

| 기능 체인 | 공통 계약 | 실패 처리 |
|---|---|---|
| pre-tool | 경로·권한·금지 작업 판단 | 금지면 차단, ask를 허용으로 바꾸지 않음 |
| post-tool | 실행 결과·변경 경로·검증 근거 기록 | 실패를 성공 artifact로 기록하지 않음 |
| session-start/end | 등록·준비 상태·정리·handoff | 부분 초기화는 재진입 가능하게 기록 |
| stop: required gate | 품질·필수 리뷰·증거 검사 | 불충족이면 완료 거부 |
| stop: goal | 목표 조건·잔여 예산·취소 상태 확인 | 미달이면 지원되는 경로로만 지속 |
| stop: advisory | 학습·보안 관찰·진단 | 실패 기록 후 비차단; 보안 차단 정책은 pre-tool에서 강제 |
| subagent | 부모·자식·scope·종료 근거 연결 | 고아 결과는 승인·완료 근거로 사용하지 않음 |
| compact/interrupt | checkpoint·취소·재개 | 이벤트 미지원이면 실행 전 대체 경로 필요 |

결과는 `allow`, `deny`, `needs_input`, `retryable_error`, `fatal_error`로 정규화한다. deny가 다른 하네스에서 allow가 되는 변환은 금지한다. 표현 불가능한 승인 요구는 일시 중단하고 사용자 입력 경로로 보낸다.

동기 훅의 제한 안에서 끝나지 않는 검증은 실행 단계에서 수행해 receipt를 남긴다. 종료 훅은 HEAD·dirty diff digest·검증 설정·명령·도구 버전과 receipt를 비교한다. 코드가 바뀌면 기존 통과는 무효다. 오래 걸리는 검사를 훅에서 무한 반복하지 않는다.

## 09. 목표 반복·중단·승인

새로운 목표 엔진을 만들지 않고 기존 goal 평가기를 호출한다. 훅만으로 지속이 보장되지 않는 호스트는 MoAI가 시작한 세션에 한해 감독 실행 경로를 제공한다. 이 경로의 thread/turn 재개가 실측되기 전에는 미지원이다.

```text
created → preflight → ready → running → verifying → completed
                         ↘ waiting_input → ready
                         ↘ failed
                         ↘ paused / cancelled
```

completed는 검사기가 확정한다. 사용자 중단은 goal 미달보다 우선하며 자동 재개하지 않는다. waiting_input은 시간이 지나도 ready로 바뀌지 않는다. 예산 소진은 성공이 아니며 상태·증거·남은 일을 남긴다. 재시도 횟수·총 실행 예산·같은 실패 반복 한도를 기존 설정과 통합한다.

질문 요청은 question_id, choices, 필요한 정보, 만료 및 run 귀속을 가진다. 호스트 질문 도구가 있으면 쓰고, 없으면 MoAI의 명시적 사용자 입력 경로를 제공한다. 응답은 사용자 세션과 run에 결합하고 다른 에이전트 메시지를 사용자 승인으로 취급하지 않는다. 기존 승인은 동일 범위에서 유지하고 scope가 바뀌면 변경된 범위만 다시 확인한다.

## 10. 스킬·에이전트·모델

| 대상 | 제안 |
|---|---|
| 스킬 35개 | 중립 본문 하나에서 양쪽 생성. 스크립트·references까지 참조 완결성 검사 |
| 명령 17개 | 각 호스트에서 발견 가능한 얇은 entry 생성; 존재하지 않는 CLI verb로 대체하지 않음 |
| 에이전트 12개 | 역할·입출력·쓰기 범위·필수 능력을 중립 정의로 관리; 기존 agentemit 확장 |
| 도구 권한 | prose만으로 read-only를 주장하지 않음. 실제 sandbox/권한/실행 경계로 확인 |
| 모델 | 역할별 요구와 사용자 선택을 호스트 모델 설정에 매핑. 미확인 ID는 거부/명시 선택 |
| 위임 | native subagent 또는 검증된 worker 세션; 순차 대체는 병렬 기능 통과로 세지 않음 |
| 동적 워크플로 | 공통 단계·검증 계약을 동일하게 적용; Claude JS는 실행 최적화로 유지 가능 |
| DesignSync | 공통 요청·결과·수정본 hash 계약과 connector capability 필요. 없으면 해당 기능 차단 |

역할별 권한을 호스트가 표현하지 못하면 더 넓은 권한을 조용히 부여하지 않는다. 외부 worker의 sandbox로 강제할 수 있는지 먼저 검증하고 불가능하면 해당 역할을 차단한다. 에이전트가 만든 결과는 공통 검증 이후에만 부모 run의 완료 증거가 된다.

## 11. 17개 명령의 공통 계약

아래는 목표 검증 표다. 현재 E2E 통과 목록이 아니다. 각 행을 Claude/Codex에서 같은 fixture로 실행한다.

| 명령 | 공통 산출물·동작 | 검증 기준 |
|---|---|---|
| plan | SPEC·계획·인수 기준·요구 추적 | 필수 항목·상태·독립 검토 근거 |
| run | 범위 내 구현·관련 테스트·진행 증거 | 요구 위반 fixture 차단·회귀 검사 |
| sync | 문서·변경 기록·검증·PR 흐름 | 정확한 HEAD와 CI/review 근거 |
| project | 프로젝트 문서·설정 | 배포 프로파일 보존·참조 완결성 |
| fix | 국소 수정·진단 해소 | 오류 재현→해결 확인 |
| loop | 제한된 반복 수정 | 반복 한도·사용자 취소·잔여 오류 |
| gate | lint·format·type·test 등 구성된 검사 | exit와 로그 귀속·empty pass 차단 |
| review | 구조화된 findings·판정 | 빈 결과/실패 연결을 pass로 해석하지 않음 |
| clean | 승인 범위의 불필요 코드 정리 | 사용자 변경 보존·동작 회귀 확인 |
| codemaps | 코드 구조 문서 | 실제 경로·의존성 추적 |
| mx | 태그 후보·수정·검증 | 프로젝트 계약·기존 내용 보존 |
| e2e | 사용자 여정·실행 결과 | 실행 표면·브라우저·실제 결과 귀속 |
| feedback | 검토 가능한 피드백·외부 등록 | 사용자 권한·중복 등록 방지·readback |
| harness | 생성된 구성·실행 계약 | 두 하네스 산출물 또는 명시적 차단 |
| goal | 목표 평가·지속·종료 | 미달 차단·취소 우선·예산 종료 |
| gtd | 수집→정리→검토→실행 기록 | 기존 저장소 계약·중복 수집 방지 |
| todo | 카드 조회·선택·상태 갱신 | 공식 CLI/API·카드 소유권·경합 처리 |

## 12. 워크트리와 실행기

MoAI 카드 작업의 worktree lifecycle owner는 MoAI 하나로 고정한다. 호스트가 별도로 생성·정리한 트리와 혼용하지 않는다. 기존 worktree 서비스와 launcher를 재사용하고 새 Git 조작 경로를 여러 곳에 만들지 않는다.

- 공통 launch 요청: harness, project_root, worktree_ref, create_if_missing, run_id, resume_ref, permissions, model_policy.
- 기존 `moai cc`·`moai codex`는 같은 worktree 서비스로 위임하되 호스트 인자 변환만 다르게 한다.
- 새 카드의 base는 원격 기본 브랜치에서 해석한다. 의존 브랜치 병합은 새 트리 내부에서만 수행한다.
- 공유 기본 checkout은 branch 전환·stash·reset·merge를 하지 않는다. 호출 직전 branch/HEAD 확인과 경합 검사를 유지한다.
- WT 이름·카드 traceability·세션 귀속을 보존한다. 통합·원격 반영 증거 전 자동 삭제하지 않는다.
- 기존 작업 중 하네스를 바꿀 때는 두 writer를 동시에 붙이지 않는다. checkpoint→소유권 반환→대상 하네스 획득→검증 후 재개한다.

**제안 UX:** `moai codex -w <name>`의 생성 의미를 Claude와 맞추려면 공통 worktree 서비스 구현 후 명시적으로 변경한다. 현재 명령의 의미는 기존 트리 선택이다. 새 `-k` 지원도 같은 원칙으로 추가하고 현재 가능하다고 안내하지 않는다.

## 13. 팩토리·세션 통신·작업 지시

기존 `internal/kanban`과 `internal/sessionmsg`를 사용한다. 세션 메시지는 알림이고, 실제 작업 권한은 공통 dispatch 레코드에 둔다. 카드당 writer 한 명을 보장한다.

| 필드 | 계약 |
|---|---|
| dispatch_id / run_id / card_id | 작업·실행·카드 식별자를 분리 |
| source_session / target_session | 호스트 ID를 MoAI 세션 ID에 매핑 |
| scope / authority_ref | 허용 경로·행동과 사용자 권한 근거 |
| worktree_ref / expected_head | 실행 장소·출발점 검증 |
| lease / fencing_token | 이전 worker의 뒤늦은 반영 거부 |
| idempotency_key / attempt | 재전송과 새 시도 구분 |
| artifact_refs / evidence_refs | 결과와 검증 연결 |

메시지 전달은 중복 가능성을 전제로 한다. 수신자는 이미 처리한 dispatch를 다시 적용하지 않는다. 완료·ack 전에 결과를 영속화한다. worker가 끊겨 lease가 끝나도 기존 writer가 살아 있는지 확인하고 fencing으로 반영 권한을 회수한 뒤 재할당한다. 프로세스 존재만으로 카드 소유권을 확정하지 않는다.

혼합 팩토리의 인증 조합은 Claude→Claude, Codex→Codex, Claude→Codex, Codex→Claude 네 가지다. 전달 성공·작업 시작·결과 완료·통합 완료를 서로 다른 상태로 기록한다. UI의 메시지 도착은 완료 증거가 아니다.

## 14. 설정·MCP·상태의 소유권

| 범위 | 소유 내용 | 제약 |
|---|---|---|
| 시스템/조직 | 정책·허용 능력·강제 제한 | 프로젝트나 에이전트가 완화 불가 |
| 사용자 | 인증·개인 선호·개인 스킬 | 프로젝트 배포가 임의 변경하지 않음 |
| 프로젝트 | 활성 하네스·workflow·quality·역할 설정 | 버전 관리와 명시적 override |
| worktree/run | 선택 카드·현재 실행·임시 receipt | 다른 run과 분리 |
| 호스트 | UI·네이티브 승인·연결 프로필 | 공통 모델에 없는 필드는 그대로 보존 |

기존 `.moai/config/sections` 체계를 유지한다. 사용자 데이터 저장 위치는 현재 저장소 API를 따라가며 새 설계가 임의로 `.moai/state`나 SQLite에 직접 쓰지 않는다. 중앙화가 필요하면 별도 migration으로 다룬다.

MCP의 공통 의미와 도구 schema는 하나로 유지하고 Claude/Codex 등록 형식만 생성한다. tools/list 발견, read-only 호출, 쓰기 승인, timeout/cancel, 잘못된 응답의 판정을 분리 검증한다. API 키·토큰은 생성 템플릿·메시지·보고서에 넣지 않는다.

## 15. 설치·업데이트·롤백

세 배포 프로파일 `claude`, `gpt`, `both`를 유지한다. 모델 이름과 하네스 이름은 내부적으로 분리하고 기존 CLI 별칭 호환성을 지킨다.

1. 현재 파일·manifest·활성 프로파일을 읽고 배포 diff를 계산한다.
2. 임시 영역에 모든 파일을 생성하고 schema·필수 참조·의무 coverage를 검사한다.
3. 잠금 후 대상 파일 해시를 다시 확인한다. 사용자 변경이면 충돌로 보고하고 중단한다.
4. 사용자 파일과 관리 파일을 구분해 적용한다. 파일별 atomic rename + transaction journal로 중단 후 복구한다. 다중 파일 변경 전체가 자동 원자적이라고 주장하지 않는다.
5. readback 후 설치 세대를 확정한다. 실패 시 journal과 이전 해시를 기준으로 복구하며 이후 사용자 변경은 덮어쓰지 않는다.
6. 훅·MCP 신뢰가 필요한 경우 `installed-untrusted`로 표시한다. 실제 실행 확인 후에만 `operational`로 올린다.

`claude → both`, `gpt → both`, `both → 단일`, 구버전→신버전, 수정된 사용자 파일, 손상된 설정, symlink 경계, 설치 중단을 모두 검사한다. 제거는 관리 소유·해시가 확인된 파일만 대상으로 하고 기존 사용자 파일은 보존한다.

## 16. 진단·관찰·사용자 경험

`doctor`와 `codex status`는 현재 파일 존재뿐 아니라 지원 계약을 보여주도록 확장한다. 런타임의 검증을 받지 않은 설치를 초록색 완료로 표시하지 않는다.

- 준비 상태: binary → config → catalog → trust → connection → event firing → workflow acceptance.
- 결과 귀속: commit, dirty_digest, binary_digest, host_version, OS, profile, run_id, test_id, verdict, evidence_ref.
- 오류 표현: 설치 오류, 연결 오류, 정책 차단, 사용자 입력 필요, 테스트 실패, 미실행을 구분한다.
- 로그: 기본은 메타데이터·digest·경로 참조. 원문 프롬프트·인증 정보 저장은 피한다.
- 상태줄: 공통 상태 모델을 호스트 표시 능력에 맞춰 표현한다. 표현할 수 없는 상세 정보는 `status`에서 확인하게 한다.

## 17. 변경 파일과 구현 책임

아래 새 경로는 제안이며, 기존 패키지로 해결 가능하면 별도 패키지로 분리하지 않는다.

| 영역 | 기존 재사용 지점 | 필요한 변경 |
|---|---|---|
| 카탈로그 | internal/template, catalog.yaml | 중립 본문·의무/기능 ID·참조 manifest |
| 배포 | harness_fs.go, deployer.go, manifest | 공통 .moai 리소스 배포·참조 검증·세대 복구 |
| 에이전트 | internal/template/agentemit | 중립 입력·권한 손실 보고·검증 |
| 훅 | internal/hook, codexadapter, codexwiring | 공통 기능 목록·동일 체인·결정 변환 |
| 실행기 | cli/cc.go, codex_launcher.go, launcher.go | 공통 launch/worktree 요청·기능 preflight |
| 목표·품질 | 기존 goal 및 gate 처리 | 공통 receipt·중단·검증된 지속 경로 |
| 팩토리 | internal/kanban, internal/sessionmsg | dispatch 소유권·중복·재개·혼합 조합 |
| 진단 | cli/doctor_codex.go, codex_readiness.go | 설정 존재와 실운영 상태 분리 |
| 문서 | docs-site와 생성 템플릿 | 지원 프로파일·현재/제안 분리·사용 예시 |

### 감사 결과와 구현·검증 연결

| 감사 항목 | 설계 절 | 구현 단계 | 완료 검사 |
|---|---|---|---|
| F1 필수 참조 누락 | 05–06 | M1 | AC-POL-01, AC-TPL-01/02 |
| F2 별도 훅 체인 누락 | 08–09 | M2 | AC-HOOK-01/02, AC-GOAL-01 |
| F3 이벤트·승인 의미 차이 | 07–09 | M2 | AC-HOOK-02, AC-OBS-01 |
| F4 launcher·권한·협업 차이 | 10–13 | M3 | AC-AGENT-01, AC-WT-01, AC-FACT-01 |
| 실제 연결·운영 Gap | 14–16, 19 | M4–M5 | AC-MCP-01, AC-WF-01 |

## 18. 단계별 구현 계획

| 단계 | 우선순위 | 산출물 | 종료 조건 | 선행 조건 |
|---|---|---|---|---|
| M0 기준 고정 | High | 기능/의무 inventory, 현재 동작 fixture | 모든 기능에 owner·evidence·gap 부여 | 없음 |
| M1 배포 완결성 | High | 중립 정책/워크플로, 두 출력기 | Codex-only 필수 참조 0건 누락 | M0 |
| M2 훅·검증 동등성 | High | 공통 체인·결정 변환·goal 지속 | required gate 실패가 양쪽 완료를 차단 | M0; 통합은 M1 |
| M3 실행·협업 | High | worktree·질문·dispatch·mixed factory | 소유권·중복·취소·권한 테스트 통과 | M1, M2 |
| M4 전환·운영 | Medium | migration·rollback·doctor·문서 | 모든 전환 행과 복구 시나리오 통과 | M1–M3 |
| M5 인증·출시 | High | 버전별 실운영 결과·지원표 | required AC 전부 PASS, NOT_RUN 0 | M4 |

단계 순서만 제시한다. 구현 기간이나 완료 날짜는 추정하지 않는다. 각 단계는 develop 기준의 새 격리 worktree에서 수행하며, 결과 반영 전 branch/HEAD와 기존 작업을 재확인한다.

## 19. 인수 기준과 테스트 전략

| ID | 기계적으로 판정할 조건 | 양성·음성 검증 |
|---|---|---|
| AC-POL-01 | 모든 required 의무가 양쪽 적용 경로·검사에 연결 | 의무 하나 삭제한 변형은 실패 |
| AC-TPL-01 | claude/gpt/both 신규 배포의 필수 참조 전부 해석 | 규칙 누락·잘못된 경로·탈출 거부 |
| AC-TPL-02 | 원본 동일 시 생성 결과 동일 | 재생성 diff 0; 사용자 수정 보존 |
| AC-HOOK-01 | feature 체인의 양쪽 효과 동등 | 이벤트 입력 golden + 실제 발화 |
| AC-HOOK-02 | deny/needs_input이 allow로 바뀌지 않음 | 표현 불가·timeout·손상 출력 주입 |
| AC-GOAL-01 | 미달 지속·충족 종료·취소 우선·예산 종료 | 반복 차단/호스트 override도 성공으로 오인하지 않음 |
| AC-AGENT-01 | 역할 권한과 산출물 계약 준수 | read-only 역할의 쓰기 시도 차단 |
| AC-WT-01 | 새/기존 트리·base·소유권·정리 보호 | 동시 writer·미통합 삭제 거부 |
| AC-MSG-01 | 중복·재시작에도 결과 한 번만 반영 | duplicate, lost ack, stale fencing token |
| AC-FACT-01 | 혼합 4조합의 카드 전 과정 완료 | worker 중단·재할당·늦은 응답 |
| AC-MCP-01 | 연결·도구 호출·승인·취소 결과 정확 | handshake 실패·RPC 오류는 inconclusive/fail |
| AC-MIG-01 | 지정된 모든 전환/복구 유형에서 사용자 데이터 보존 | 중단 지점 주입·손상 파일·충돌 |
| AC-WF-01 | 17개 명령 × 두 CLI의 required 계약 충족 | 성공 사례와 대표 실패 사례 |
| AC-OBS-01 | 모든 PASS가 실행 환경·코드·로그에 귀속 | skip/empty run/낡은 receipt 차단 |

설정에는 Mac/Linux/Windows와 symlink/copy 변형을 반영한다. 모든 운영체제를 지원 완료로 공표하려면 해당 런타임 검증이 필요하다. Desktop/Web도 독립 프로파일로 검사한다.

로컬에서는 변경 범위 테스트만 실행하고 전체 회귀는 CI에 맡긴다. 실제 모델 E2E는 격리 저장소·명시적 호출 예산·cleanup 보장 아래 실행한다. 실패와 미실행을 분리 집계하며 제품 공통 기능에서 skip을 pass로 세지 않는다.

## 20. 위험·대안·결정 기록

| 위험 | 대응 설계 | 남는 검증 |
|---|---|---|
| 지시문 예산으로 핵심 계약 누락 | 작은 standing 계약 + 의무 coverage | 개인/상위 지시문 포함 실제 로드 |
| 모델이 절차를 생략 | 상태 전이·필수 검증을 CLI에서 강제 | 직접 도구 사용 우회 경계 |
| 호스트 버전 변화 | digest 기반 탐지·지원표·재인증 | TUI/비대화형 차이 |
| 훅 순서·중복 | 공통 체인·멱등 처리·세대 ID | 중복 이벤트·동시 도착 |
| 장시간 검증의 훅 timeout | 작업 중 검사·종료 시 receipt 확인 | dirty digest와 검사 범위 정확성 |
| 두 호스트가 같은 파일 수정 | lease·fencing·worktree 분리 | 호스트 외 수동 수정 감지 |
| 사용자 설정 덮어쓰기 | 소유 manifest·해시 재확인·journal | 중간 실패 복구 |
| 과도한 추상화 | 기존 서비스 확장·최소 인터페이스 | 각 추상화에 실제 두 소비자 존재 |

ADR-01: 중립 원본 하나와 결정적 생성. ADR-02: 기존 상태 저장 API 재사용. ADR-03: required 기능의 조용한 생략 금지. ADR-04: MoAI worktree lifecycle 단독 소유. ADR-05: 메시지와 작업 권한 분리. ADR-06: 네이티브 UI 동일성보다 작업·권한·검증 결과 동등성을 인증.

## 21. 검토할 결정과 문서의 한계

이 문서의 기본안은 제안이며 구현 착수 승인을 의미하지 않는다. 다음 결정은 구현 단계의 첫 검토에서 고정한다.

| 결정 | 기본 제안 | 변경 시 영향 |
|---|---|---|
| 제품 인증 범위 | 두 CLI + 혼합 팩토리 우선, Desktop/Web 별도 | 첫 릴리스의 인증 행 변경 |
| 필수 기능 범위 | 현재 MoAI 공통 기능 전체 | 축소 시 ‘전체 동등 지원’ 표현 사용 불가 |
| 권한을 표현 못하는 역할 | 안전한 worker 경계 없으면 차단 | 편의성보다 권한 계약 유지 |
| 새 worktree UX | 공통 서비스 이후 양쪽 의미 통일 | 기존 launcher 호환성 검사 |
| 규칙 위치 | .moai/policies; nested AGENTS 자동 생성 제외 | 기존 저장소 계약 변경 필요 여부 |

Claim: 전체 구현을 위한 목표 구조와 검증 기준을 설계했다. Evidence: 앞선 로컬 감사·현재 HEAD 및 관련 파일 diff·공식 문서 확인에 근거한다. Baseline: 감사 `7f86971f`, 설계 시점 `3f3ffbb5`. Gaps: 새 구조·명령·인수 테스트는 구현·실행되지 않았다. Residual risk: 실제 호스트 능력이 부족하면 일부 기능은 안전한 구현 경로를 추가 확보해야 하며, 그 전에는 전체 지원을 선언할 수 없다.

## 22. 근거 자료

- [직전 develop 전수 감사](/tmp/moai-codex-develop-audit-20260922.md) — 로컬 조사 artifact. 공유본에는 별도 첨부 필요.
- [기존 이중 하네스 설계](moai-dual-harness-design-20260921.md) / [기존 구현 계획](moai-dual-harness-implementation-plan-20260921.md).
- [OpenAI AGENTS.md 안내](https://learn.chatgpt.com/docs/agent-configuration/agents-md): 지시문 탐색·크기 제한. 공통 계약을 작게 유지하는 근거.
- [OpenAI Skills 안내](https://learn.chatgpt.com/docs/build-skills): SKILL.md와 프로젝트 .agents/skills 탐색. Codex 생성 표면의 근거.
- [Claude Code Memory 안내](https://code.claude.com/docs/en/memory): .claude/rules 및 경로별 적용. 규칙을 단순 스킬로 대체하지 않는 근거.
- [Claude Code Hooks 안내](https://code.claude.com/docs/en/hooks): Stop·취소·반복 제약. 무한 자동 지속을 가정하지 않는 근거.

공식 문서는 2026-09-22 열람했다. 문서의 가능성과 설치된 바이너리의 관측 결과는 별개다. 기존 설계의 제품 버전별 주장은 그대로 복사하지 않고 재인증 대상으로 남겼다.
