---
title: /moai
weight: 20
draft: false
---

완전 자율 자동화 명령어입니다. 사용자가 목표를 제공하면 MoAI가 **plan → run → sync** 파이프라인을 자율적으로 실행합니다.

{{< callout type="info" >}}
  **한 줄 요약**: `/moai`는 "완전 자율 자동화" 명령어입니다. 사용자는 원하는
  기능을 자연어로 설명하기만 하면, MoAI가 SPEC 생성부터 구현, 문서화까지 **모든
  과정을 자동으로** 수행합니다.
{{< /callout >}}

{{< callout type="info" >}}
**슬래시 커맨드 지원**: MoAI의 모든 서브커맨드는 스킬로 래핑되어 있어, `/moai`만 입력하면 사용 가능한 서브커맨드 목록이 표시됩니다. 각 서브커맨드는 `/moai:fix`, `/moai:loop`, `/moai:review` 등의 형식으로 바로 실행할 수도 있습니다.
{{< /callout >}}

## 개요

`/moai`는 MoAI-ADK의 **완전 자율 자동화 워크플로우** 명령어입니다. 하위 명령어를 따로 실행할 필요 없이, 단 한 번의 명령으로 전체 개발 프로세스가 자동화됩니다:

1. **SPEC 생성** (manager-spec)
2. **DDD/TDD 구현** (manager-develop — quality.yaml의 development_mode에 따라)
3. **문서 동기화** (manager-docs)

## Analyze-First 라우팅

v3부터 `/moai`의 기본 라우팅은 **Analyze-First** — 언어 독립적 의도 분석입니다. 영어 키워드 매칭이 아니라 요청의 의미를 분류하므로, 어떤 `conversation_language`로 요청해도 같은 품질로 라우팅됩니다.

라우팅은 다음 순서로 진행됩니다:

1. **의도 분석**: 사용자 요청의 의도를 분류 (입력 언어와 무관)
2. **컨텍스트 충분성 확인**: 불충분하면 Socratic 인터뷰로 명확화
3. **실행 계획 구성**: 스킬 / 에이전트 / 동적 워크플로우 체인 선택
4. **오케스트레이션 모드 선택** (Phase 4): solo-sequential / parallel-subagents / dynamic-workflow

즉 `/moai "로그인 버그 고쳐줘"`처럼 서브커맨드 없이 자연어만 입력해도, 의도 분석을 거쳐 알맞은 워크플로우 (수정이면 fix 계열, 신규 기능이면 plan→run→sync 파이프라인)로 연결됩니다.

## 사용법

```bash
# 기본 사용법
> /moai "구현하고 싶은 기능 설명"

# 워크트리와 함께
> /moai "기능 설명" --worktree

# 브랜치와 함께
> /moai "기능 설명" --branch

# 루프 모드 활성화
> /moai "기능 설명" --loop

# 기존 SPEC 재개
> /moai --resume SPEC-AUTH-001
```

## 지원 플래그

| 플래그              | 설명                             | 예시                           |
| ------------------- | -------------------------------- | ------------------------------ |
| `--loop`            | 구현 후 자동 반복 수정 활성화    | `/moai "기능" --loop`          |
| `--max N`           | 최대 반복 횟수 지정 (기본값 100) | `/moai "기능" --loop --max 10` |
| `--branch`          | 자동 feature 브랜치 생성         | `/moai "기능" --branch`        |
| `--pr`              | 완료 후 자동 PR 생성             | `/moai "기능" --pr`            |
| `--resume SPEC-XXX` | 기존 SPEC 작업 재개              | `/moai --resume SPEC-AUTH-001` |
| `--team`            | 에이전트 팀 모드 강제            | `/moai "기능" --team`          |
| `--solo`            | 하위 에이전트 모드 강제          | `/moai "기능" --solo`          |

### --loop 플래그

구현이 완료된 후 자동으로 반복 수정을 실행하여 모든 오류를 수정합니다:

```bash
> /moai "JWT 인증 시스템" --loop
```

이 옵션을 사용하면:

1. SPEC 생성
2. DDD 구현
3. **자동 루프 실행** (LSP 오류, 테스트 실패, 커버리지 부족 해결)
4. 문서 동기화
5. PR 생성

{{< callout type="info" >}}
  `--loop` 옵션은 **구현 후 정리 작업을 완전히 자동화**하여 생산성을
  극대화합니다.
{{< /callout >}}

### --team / --solo 플래그와 오케스트레이션 모드

플래그 없이 실행하면 MoAI가 작업 규모를 보고 오케스트레이션 모드를 자동 선택합니다:

**자동 선택 기준** (플래그 없을 때):

- 영향 도메인 >= 3개 → 병렬 실행
- 수정 파일 >= 10개 → 병렬 실행
- 복잡도 점수 >= 7 → 병렬 실행
- 그 외 → 하위 에이전트 모드 (순차 실행)

| 플래그 | 동작 |
| ------ | ---- |
| `--team` | 에이전트 팀 모드 강제 |
| `--solo` | 하위 에이전트 모드 강제 (순차 실행) |
| (없음) | 복잡도 기반 자동 선택 |

{{< callout type="warning" >}}
**v3.0.0-rc11 변경**: Agent Teams 정적 오케스트레이션 계층은 **은퇴**했습니다. `--team`을 강제해도 `MODE_TEAM_UNAVAILABLE`과 함께 하위 에이전트 모드로 폴백합니다. 병렬 실행은 병렬 하위 에이전트 팬아웃과 동적 워크플로우 2종 (plan-phase 연구 병렬 팬아웃, sync-phase 4차원 품질 평가)이 담당하며, 네이티브 teammate 런타임 (`moai cg`의 tmux pane)은 그대로 유지됩니다.
{{< /callout >}}

병렬 실행은 에이전트마다 독립 컨텍스트 윈도우를 쓰므로 토큰 사용량이 늘어납니다. 단순한 단일 도메인 작업에는 `--solo` (순차)가 더 경제적입니다 — 규모 기반 자동 선택이 기본값인 이유입니다.

## 실행 과정

`/moai`가 내부적으로 수행하는 전체 과정입니다:

```mermaid
flowchart TD
    A["명령어 실행<br/>/moai '기능 설명'"] --> B{--resume?}
    B -->|예| C["SPEC 로드<br/>이어서 작업"]
    B -->|아니오| D["Phase 0<br/>병렬 탐색"]

    subgraph D["Phase 0: 병렬 탐색 (15-30초)"]
        D1["Explore 하위 에이전트<br/>코드베이스 분석"]
        D2["Research 하위 에이전트<br/>외부 문서 조사"]
        D3["Quality 하위 에이전트<br/>품질 기준선 확인"]
    end

    D --> E{"단일 도메인?"}
    E -->|예| F["전문가 에이전트에<br/>직접 위임"]
    E -->|아니오| G["Phase 1 계속"]

    C --> G["Phase 1<br/>SPEC 생성"]
    G --> H["manager-spec 호출"]
    H --> I["EARS 형식 SPEC 생성"]
    I --> J[".moai/specs/SPEC-XXX/spec.md"]

    J --> K["Phase 2<br/>DDD 구현"]

    K --> L["manager-develop 호출<br/>DDD/TDD 순환 (quality.yaml에 따라)"]
    L --> M{"구현 완료?"}
    M -->|아니오| L
    M -->|예| N{"--loop?"}

    N -->|예| O["자동 루프 실행"]
    O --> P["모든 문제 해결"]
    N -->|아니오| P

    P --> Q["Phase 3<br/>문서 동기화"]

    Q --> R["manager-docs 호출<br/>문서 생성"]
    R --> S{"--pr?"}
    S -->|예| T["PR 생성"]
    S -->|아니오| U["완료 신호"]
    T --> U
```

**핵심 포인트:**

- **Phase 0 (병렬 탐색)**: 세 에이전트가 동시에 실행되어 2-3배 속도 향상
- **단일 도메인 라우팅**: 단순 작업은 전문가 에이전트에 직접 위임하여 SPEC 건너뜀
- **완료 신호**: 작업 완료 시 완료 보고서에 작업 완료를 명시

## Phase별 상세

### Phase 0: 병렬 탐색 (선택적)

세 에이전트가 **동시에** 실행되어 프로젝트 맥락을 빠르게 파악합니다:

| 에이전트     | 역할            | 작업                                     |
| ------------ | --------------- | ---------------------------------------- |
| **Explore**  | 코드베이스 분석 | 관련 파일, 아키텍처 패턴, 기존 구현 발견 |
| **Research** | 외부 문서 조사  | 공식 문서, API 문서, 유사 구현 예시      |
| **Quality**  | 품질 기준선     | 테스트 커버리지, 린트 상태, 기술 부채    |

**속도 향상:** 병렬 실행으로 순차 실행 대비 2-3배 빠름 (15-30초 vs 45-90초)

**단일 도메인 라우팅:**

- 단일 도메인 작업 (예: "SQL 최적화"): SPEC 생성 없이 전문가 에이전트에 직접 위임
- 다중 도메인 작업: 전체 워크플로우 진행

### Phase 1: SPEC 생성

**manager-spec** 하위 에이전트가 EARS 형식 SPEC 문서를 생성합니다:

- .moai/specs/SPEC-XXX/spec.md
- EARS 형식 요구사항
- Given-When-Then 인수 기준
- conversation_language로 작성된 콘텐츠

### Phase 2: DDD/TDD 구현 루프

**manager-develop** 하위 에이전트가 SPEC을 기반으로 구현을 수행합니다:

- DDD 순환: ANALYZE-PRESERVE-IMPROVE (기존 코드 리팩토링)
- TDD 순환: RED-GREEN-REFACTOR (새 기능 개발)
- 도메인 컨텍스트 자동 주입 (백엔드, 프론트엔드, 보안, 데이터베이스 등)

**quality.yaml development_mode 설정:**

- `development_mode: ddd` → DDD 순환 사용 (기존 코드 개선)
- `development_mode: tdd` → TDD 순환 사용 (새 기능 개발, 기본값)

**루프 동작 (--loop 또는 loop.enabled가 true일 때):**

```
문제가 존재 AND 반복 < 최대값:
  1. 진단 실행 (LSP 오류, 테스트 실패, 커버리지)
  2. manager-develop에 수정 위임
  3. 수정 결과 검증
  4. 완료 조건 충족 여부 확인
  5. 완료 문장 감지 시 루프 종료
```

### Phase 3: 문서 동기화

**manager-docs** 하위 에이전트가 구현과 문서를 동기화합니다:

- API 문서 생성
- README 업데이트
- CHANGELOG 추가
- 성공 시 작업 완료를 명시

## TODO 관리

**[HARD] TodoWrite 도구 필수:** 모든 작업 추적으로 TodoWrite 사용 필수

- 이슈 발견 시: TodoWrite (pending 상태)
- 작업 시작 전: TodoWrite (in_progress 상태)
- 작업 완료 후: TodoWrite (completed 상태)
- TODO 목록을 텍스트로 출력 금지

## 완료 신호

모든 워크플로우 단계가 성공적으로 완료되면, MoAI는 완료 보고서(배너/산문)에 작업 완료를 명시하여 결과를 명확히 합니다.

## LLM 모드 라우팅

토크노믹스의 핵심 장치입니다. llm.yaml 설정에 따라 단계별로 Claude와 GLM을 자동 라우팅합니다 — 전략·계획은 Claude가, 대량 구현은 저비용 GLM이 담당하는 하이브리드가 가능합니다.

| 모드          | Plan 단계      | Run 단계       |
| ------------- | -------------- | -------------- |
| `claude-only` | Claude         | Claude         |
| `hybrid`      | Claude         | GLM (worktree) |
| `glm-only`    | GLM (worktree) | GLM (worktree) |

## 실전 예시

### 예시: JWT 인증 시스템 완전 자동화

**1단계: 명령어 실행**

```bash
> /moai "JWT 기반 사용자 인증 시스템: 회원가입, 로그인, 토큰 갱신" --worktree --loop --pr
```

**2단계: Phase 0 - 병렬 탐색**

```
[병렬 탐색 시작]
  Explore 하위 에이전트: src/auth/ 분석 중...
  Research 하위 에이전트: JWT best practices 조사 중...
  Quality 하위 에이전트: 테스트 커버리지 32% 확인...

[탐색 완료 - 23초]
  발견 파일: 4개
  권장 라이브러리: PyJWT, bcrypt
  기준선: LSP 0 오류, 커버리지 32%
```

**3단계: Phase 1 - SPEC 생성**

```
[manager-spec 호출]
  SPEC ID: SPEC-AUTH-001
  요구사항: 5개 (EARS 형식)
  인수 기준: 3개 시나리오

  사용자 승인: 완료
```

**4단계: Phase 2 - DDD 구현**

```
[manager-spec]
  작업 분해: 7개 태스크
  전략 계획 완료

[manager-develop]
  ANALYZE: 코드 구조 분석 완료
  PRESERVE: 특성화 테스트 12개 작성
  IMPROVE: 7개 태스크 구현 완료

[sync-auditor]
  TRUST 5: 모든 기둥 통과
  커버리지: 89%
  상태: PASS
```

**5단계: 자동 루프 (--loop)**

```
[루프 시작 - 반복 1/100]
  진단: 타입 오류 2개 발견
  수정: manager-develop 하위 에이전트에 위임
  검증: 모든 오류 해결됨

[루프 종료 - 1회 반복]
  완료 조건 충족!
```

**6단계: Phase 3 - 문서 동기화**

```
[manager-docs]
  API 문서: docs/api/auth.md 생성
  README: 사용법 섹션 업데이트
  CHANGELOG: v1.1.0 항목 추가
  SPEC-AUTH-001: ACTIVE → COMPLETED
```

**7단계: 완료**

```
[완료]
  SPEC: SPEC-AUTH-001
  커밋: 7개
  테스트: 36/36 통과
  커버리지: 89%
  PR: #42 생성 (Draft → Ready)

<moai:COMPLETE />
```

## 자주 묻는 질문

### Q: `/moai`와 하위 명령어의 차이는 무엇인가요?

| 명령어       | 범위        | 사용 시점                    |
| ------------ | ----------- | ---------------------------- |
| `/moai`      | 전체 자동화 | 빠른 완전 자동화 원할 때     |
| `/moai plan` | SPEC 생성만 | SPEC을 먼저 검토하고 싶을 때 |
| `/moai run`  | 구현만      | SPEC이 이미 있을 때          |
| `/moai sync` | 문서화만    | 구현 후 문서만 업데이트할 때 |

### Q: --loop 플래그를 언제 사용해야 하나요?

구현 후 자동으로 모든 오류를 수정하고 싶을 때 사용합니다. 특히 대규모 리팩토링 후 정리 작업에 유용합니다.

### Q: 단일 도메인 라우팅이란 무엇인가요?

단일 도메인 작업 (예: "SQL 쿼리 최적화")은 SPEC 생성 없이 해당 도메인 전문가 에이전트에 직접 위임하여 시간을 절약합니다.

### Q: 영어가 아닌 언어로 요청해도 되나요?

네. Analyze-First 라우팅은 언어 독립적 의도 분석이므로, 한국어·일본어·중국어 등 어떤 언어로 요청해도 동일하게 동작합니다.

## 관련 문서

- [/moai plan](/workflow-commands/moai-plan) - SPEC 생성 상세
- [/moai run](/workflow-commands/moai-run) - DDD 구현 상세
- [/moai sync](/workflow-commands/moai-sync) - 문서 동기화 상세
- [/moai loop](/utility-commands/moai-loop) - 반복 수정 루프 상세
- [/moai fix](/utility-commands/moai-fix) - 일회성 자동 수정 상세
