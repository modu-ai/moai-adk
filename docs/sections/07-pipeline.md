# MoAI-ADK 4단계 파이프라인

## 📝 신규 4단계 워크플로우

MoAI-ADK는 개인/팀 모드에 최적화된 다음 흐름을 제공합니다.

```mermaid
flowchart LR
    A[/moai:0-project] --> B[/moai:1-spec]
    B --> C[/moai:2-build]
    C --> D[/moai:3-sync]
```

## 모델 사용 가이드

| 단계 | 권장 모델 | 비고 |
|------|-----------|------|
| `/moai:0-project` | `sonnet` | 프로젝트 문서 갱신 + 메모리 로드 |
| `/moai:1-spec` | `sonnet` | auto 제안(개인: 로컬 생성 / 팀: GitHub Issue) |
| `/moai:2-build` | `sonnet` | TDD(개인: 체크포인트 / 팀: 7단계 커밋) |
| `/moai:3-sync` | `haiku` | 문서/PR 동기화 + TAG 인덱스 갱신 |

> 💡 모델/운영 가이드 세부 내용은 프로젝트 루트 `CLAUDE.md`의 “모델 사용 가이드(opusplan)” 섹션을 참고하세요.

## 단계별 상세

### 1. 프로젝트 킥오프 - 문서 초기화
**명령어**: `/moai:0-project`
**담당**: project-manager 에이전트 (cc-manager 지원)

**목표**: product/structure/tech 문서를 인터랙티브하게 갱신하고 CLAUDE 메모리에 반영

**결과물**:
- `.moai/project/{product,structure,tech}.md` 갱신
- CLAUDE.md `@` 임포트 섹션 반영
- 외부 브레인스토밍 사용 시 `.moai/config.json.brainstorming` 업데이트

### 2. SPEC - auto 제안/생성
**명령어**: `/moai:1-spec`
**담당**: spec-builder 에이전트 (git-manager 협력)

**목표**: 프로젝트 문서를 읽고 주요 기능 SPEC을 자동 제안
- 개인: 승인 시 `.moai/specs/`에 일괄 생성
- 팀: 승인 시 GitHub Issue 생성 및 브랜치 템플릿 연동

**결과물**:
- SPEC-XXX 디렉터리와 핵심 문서(spec.md, acceptance.md 등)
- 또는 GitHub Issue(팀)

**외부 브레인스토밍(선택)**:
- `.moai/config.json.brainstorming.enabled` 가 `true` 이면 codex-bridge/gemini-bridge 에이전트에서 headless 분석 결과를 수집해 다양한 설계안을 비교합니다. (예: `Task: use codex-bridge to run "codex exec --model gpt-5-codex ..."`)

**보조 역할**:
- git-manager가 브랜치 네이밍/PR 초안 연계를 검토하고, 병렬 승인이 필요한 경우 충돌 가능성을 경고합니다.

### 3. BUILD - TDD 구현
**명령어**: `/moai:2-build`
**담당**: code-builder 에이전트 (git-manager, doc-syncer 협력)

**목표**: TDD 사이클로 구현 진행
- 개인: 자동 체크포인트(파일 변경 + 5분 주기)
- 팀: 7단계 커밋(RED→GREEN→REFACTOR)

**결과물**:
- 소스/테스트 코드
- 체크포인트 또는 구조화 커밋 히스토리

**외부 브레인스토밍(선택)**:
- 구현 전후로 codex-bridge/gemini-bridge 출력과 Claude 제안을 Self-Consistency 방식으로 비교해 최적안을 선택합니다.

**보조 역할**:
- git-manager가 스테이징/커밋 전략을 보조하고, doc-syncer가 3-sync를 위한 변경 요약을 수집합니다.

### 4. SYNC - 문서/PR 동기화
**명령어**: `/moai:3-sync`
**담당**: doc-syncer 에이전트 (git-manager 협력)

**목표**: Living Document 갱신 및 상태 보고
- TAG 인덱스 갱신 + `docs/status/sync-report.md` 생성
- 팀: Draft → Ready 전환, 리뷰어 할당(옵션)

**결과물**:
- 최신화된 문서/PR 상태
- `docs/sections/index.md` 갱신일 반영

**외부 브레인스토밍(선택)**:
- 문서 보완 아이디어나 리스크 분석이 필요하면 codex-bridge/gemini-bridge 결과를 동기화 리포트에 반영할 수 있습니다.

**보조 역할**:
- git-manager가 커밋/PR 상태를 점검하고, 필요한 경우 gh CLI를 통한 Ready 전환을 실행합니다.

## 단계별 Gate 체크포인트 (요약)

### Gate 1: SPECIFY → PLAN
**검증 항목**:
- EARS 형식 준수 여부
- @REQ 태그 완성도
- [NEEDS CLARIFICATION] 해결 여부

**통과 조건**:
```bash
✅ 모든 요구사항이 EARS 형식으로 작성됨
✅ 불명확한 요구사항이 모두 해결됨
✅ @REQ 태그가 올바르게 생성됨
```

### Gate 2: SPEC → BUILD
**검증 항목**:
- 개발 가이드 5원칙 준수
- 아키텍처 결정의 타당성
- 기술적 위험 요소 대응 계획

**통과 조건**:
```bash
✅ 개발 가이드 Check 통과
✅ ADR 문서 작성 완료
✅ @DESIGN 태그 체계 완성
```

### Gate 3: BUILD → SYNC
**검증 항목**:
- TDD 사이클 계획 완료
- 테스트 케이스 정의 완료
- 작업 우선순위 설정

**통과 조건**:
```bash
✅ 모든 작업이 테스트 우선으로 계획됨
✅ @TASK 태그 추적성 확보
✅ Sprint 계획 수립 완료
```

### Gate 4: SYNC → COMPLETE
**검증 항목**:
- 테스트 커버리지 80% 이상
- 모든 테스트 통과
- 코드 품질 기준 준수

**통과 조건**:
```bash
✅ 테스트 커버리지 ≥ 80%
✅ 모든 단위/통합 테스트 통과
✅ @TEST 태그 완성
```

## 실행 예시

```bash
# 개인 모드
/moai:0-project
/moai:1-spec
/moai:2-build
/moai:3-sync

# 팀 모드 (필요 시 먼저 `/moai:0-project update`로 팀 설정 확인)
/moai:0-project update
/moai:1-spec
/moai:git:branch --team SPEC-001
/moai:2-build SPEC-001
/moai:3-sync
```

## 파이프라인 상태 관리

### 상태 추적
```json
// .moai/indexes/state.json
{
  "current_stage": "TASKS",
  "specs": {
    "SPEC-001": {
      "stage": "IMPLEMENT",
      "progress": 75,
      "tests_passing": true,
      "coverage": 0.85
    }
  }
}
```

### 진행률 모니터링
```bash
# 현재 프로젝트 상태(요약)
moai status

# 상세 상태(버전/파일 카운트 등)
moai status -v

# 다른 경로의 프로젝트 상태 확인
moai status -p /path/to/project
```

파이프라인은 **체계적인 개발 프로세스**와 **자동화된 품질 보장**을 통해 안정적인 소프트웨어 개발을 지원합니다.
