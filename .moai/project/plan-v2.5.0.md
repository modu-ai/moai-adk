# moai-adk v2.5.0 업데이트 플랜

> **작성일**: 2026-02-18
> **최종 업데이트**: 2026-02-18 (v2.1.30 항목 추가, 공식 문서 검증 완료)
> **대상 버전**: v2.5.0 (현재: v2.4.7)
> **기준**: Claude Code v2.1.30 ~ v2.1.45 릴리즈 노트 분석 + 공식 문서 검증
> **상태**: DRAFT - 승인 대기

---

## 1. 요약 (Executive Summary)

moai-adk v2.5.0은 Claude Code v2.1.30~v2.1.45에서 도입된 핵심 기능들을 MoAI 템플릿 시스템에 통합하는 **기능 강화 릴리즈**이다.

### 핵심 목표

1. **Agent Teams 프로덕션 준비 완료**: TeammateIdle/TaskCompleted hook에 실질적 품질 검증 로직 구현, 팀 리소스 관리 CLI 명령 추가
2. **Agent Persistent Memory 지원**: agent 정의 파일에 `memory` frontmatter 필드 지원, 템플릿 에이전트에 적절한 scope 할당
3. **Task 메트릭 모니터링**: Task tool 결과에 포함되는 토큰/도구사용/시간 메트릭 활용, `/moai insights` 분석
4. **관측성 강화**: `/debug` 명령 연동, hook 디버깅 가이드, MCP OAuth 설정 가이드
5. **사용자 경험 개선**: MoAI 전용 spinner tips, 플러그인 관리 설정, hook 품질 검증 템플릿

### 이번 세션에서 이미 완료된 사항 (v2.5.0 범위에서 제외)

- SubagentStop CLI 버그 수정 (`moai hook subagent-stop` 미등록 문제)
- SessionEnd hook에 고아 팀 디렉토리 및 tmux 세션 정리 로직 추가
- 고아 리소스 정리 완료 (tmux 세션 45개, 팀 디렉토리 18개)

### 영향 범위

| 영역 | 변경 파일 수 (예상) | 위험도 |
|------|---------------------|--------|
| Go 소스 코드 (`internal/`) | 8-12개 | 중간 |
| 템플릿 파일 (`internal/template/templates/`) | 25-35개 | 낮음 |
| 설정 파일 (`.moai/config/`) | 2-3개 | 낮음 |
| 테스트 파일 (`*_test.go`) | 6-10개 | 낮음 |

---

## 2. 기능 로드맵 (Feature Roadmap)

### CRITICAL 우선순위

#### 2.1 Agent Teams 프로덕션 준비

**현재 상태**: hook wrapper 스크립트와 CLI 명령은 등록되어 있으나, 실제 품질 검증 로직이 없다.

**목표 변경사항**:

| 항목 | 현재 상태 | 목표 상태 | 복잡도 |
|------|-----------|-----------|--------|
| TeammateIdle hook 로직 | pass-through (exit 0) | LSP 오류 검증, lint 검사, 테스트 커버리지 확인 후 exit 2로 블록 | 높음 |
| TaskCompleted hook 로직 | pass-through (exit 0) | SPEC acceptance criteria 검증, 파일 소유권 검사 | 높음 |
| `moai team list` 명령 | 미존재 | 활성 팀/팀메이트 목록 표시 | 중간 |
| `moai team cleanup` 명령 | 미존재 | 고아 팀 디렉토리/tmux 세션 정리 | 중간 |
| 파일 소유권 전략 | workflow.yaml에 문서만 존재 | 팀메이트별 파일 패턴 매핑 구현 | 높음 |
| 팀 수명주기 가이드 | 미존재 | 문제 해결 가이드 및 디버깅 팁 | 낮음 |

**영향 파일**:
- `internal/hook/teammate_idle.go` (신규)
- `internal/hook/task_completed.go` (신규)
- `internal/cli/team.go` (신규)
- `internal/template/templates/.claude/hooks/moai/handle-teammate-idle.sh.tmpl` (업데이트)
- `internal/template/templates/.claude/hooks/moai/handle-task-completed.sh.tmpl` (업데이트)
- `.moai/config/sections/workflow.yaml` 스키마 확장

#### 2.2 Agent Persistent Memory 지원 (v2.1.33)

**현재 상태**: 일부 agent 정의 파일(`team-analyst.md`, `expert-backend.md` 등)에 `memory` frontmatter가 이미 추가되어 있다.

**목표 변경사항**:

| 항목 | 현재 상태 | 목표 상태 | 복잡도 |
|------|-----------|-----------|--------|
| agent 정의 일관성 | 일부 agent만 memory 필드 보유 | 전체 27개 agent에 적절한 scope 할당 | 낮음 |
| memory scope 전략 | 비일관적 | 체계적 전략 문서화 (아래 표 참조) | 낮음 |
| `moai init` memory 디렉토리 | 미생성 | `.claude/agent-memory/` 디렉토리 자동 생성 | 낮음 |
| `.gitignore` 업데이트 | memory 디렉토리 미포함 | local scope 디렉토리 제외 규칙 추가 | 낮음 |

**Memory Scope 전략**:

| Agent 카테고리 | Scope | 이유 |
|---------------|-------|------|
| Manager agents (8) | `project` | 프로젝트별 워크플로우 학습, VCS로 팀 공유 |
| Expert agents (8) | `project` | 프로젝트별 도메인 패턴 학습 |
| Builder agents (3) | `user` | 사용자별 생성 패턴 학습, 크로스 프로젝트 |
| Team agents - plan (3) | `project` | 프로젝트 컨텍스트 유지 |
| Team agents - run (5) | `project` | 구현 패턴 학습 |

**영향 파일**:
- `internal/template/templates/.claude/agents/moai/*.md` (전체 27개)
- `internal/template/templates/.gitignore.tmpl`
- `internal/template/deployer.go` (memory 디렉토리 생성 로직)

---

### HIGH 우선순위

#### 2.3 Task(agent_type) 구문 지원 (v2.1.33)

**현재 상태**: agent-authoring.md에 문서화만 되어 있고, 실제 coordinator agent 템플릿에는 적용 안 됨.

**목표 변경사항**:

| 항목 | 현재 상태 | 목표 상태 | 복잡도 |
|------|-----------|-----------|--------|
| MoAI coordinator agent | `tools` 필드에 Task만 포함 | `tools: Task(expert-backend, expert-frontend, ...), Read, Bash` | 낮음 |
| agent-authoring.md | "not currently applicable" 표기 | 실제 사용 예시 추가 | 낮음 |
| settings.json.tmpl | 해당 없음 | 해당 없음 (agent 정의에서만 사용) | - |

> **참고**: Task(agent_type) 구문은 `claude --agent` 로 실행되는 메인 스레드 agent에만 적용된다. MoAI의 subagent들은 이 제한의 영향을 받지 않지만, coordinator agent 패턴 문서화는 유용하다.

**영향 파일**:
- `internal/template/templates/.claude/rules/moai/development/agent-authoring.md`
- coordinator agent 문서 (해당 시 추가)

#### 2.4 Claude Sonnet 4.6 지원 (v2.1.45)

**현재 상태**: workflow.yaml의 `default_model`이 `sonnet`으로 설정됨. 모델명에 버전이 없음.

**목표 변경사항**:

| 항목 | 현재 상태 | 목표 상태 | 복잡도 |
|------|-----------|-----------|--------|
| workflow.yaml model 필드 | `sonnet` (버전 미명시) | `sonnet` 유지 (Claude Code가 최신 해석) | 없음 |
| investigation 패턴 model | `haiku` | `haiku` 유지 (비용 최적화) | 없음 |
| agent 정의 model 필드 | `inherit` / `haiku` | 변경 없음 (Claude Code가 해석) | 없음 |
| 문서 업데이트 | Sonnet 4.6 미언급 | CLAUDE.md, workflow 문서에 Sonnet 4.6 참조 추가 | 낮음 |
| Fast mode 문서화 | 미문서화 | Fast mode 활용 가이드 추가 | 낮음 |

> **결론**: Claude Code는 `sonnet`/`haiku`/`opus` 키워드를 최신 모델로 자동 해석하므로, workflow.yaml의 값 변경은 불필요하다. 문서에 Sonnet 4.6의 존재와 fast mode를 명시하면 충분하다.

**영향 파일**:
- `internal/template/templates/CLAUDE.md` (문서 참조)
- `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md` (문서 참조)

#### 2.5 spinnerTipsOverride 지원 (v2.1.45)

**현재 상태**: settings.json.tmpl에 `"spinnerTipsEnabled": true`만 설정됨.

**목표 변경사항**:

| 항목 | 현재 상태 | 목표 상태 | 복잡도 |
|------|-----------|-----------|--------|
| spinnerTipsOverride | 미존재 | MoAI 전용 spinner 메시지 추가 | 낮음 |
| settings.json.tmpl | `spinnerTipsEnabled: true` | + `spinnerTipsOverride: [...]` | 낮음 |
| settings.go | 해당 없음 | SpinnerTipsOverride 필드 추가 | 낮음 |

**MoAI 전용 Spinner Tips 예시**:
```json
"spinnerTipsOverride": [
  "Analyzing SPEC requirements...",
  "Running DDD analysis cycle...",
  "Checking TRUST 5 quality gates...",
  "Delegating to expert agent...",
  "Generating characterization tests...",
  "Synchronizing documentation...",
  "Planning implementation strategy...",
  "Validating acceptance criteria..."
]
```

**영향 파일**:
- `internal/template/templates/.claude/settings.json.tmpl`
- `internal/template/settings_test.go`

---

### MEDIUM 우선순위

#### 2.6 Plugin Management 지원 (v2.1.45)

**현재 상태**: `enabledPlugins`와 `extraKnownMarketplaces`가 settings.json에 없음.

**목표 변경사항**:

| 항목 | 현재 상태 | 목표 상태 | 복잡도 |
|------|-----------|-----------|--------|
| enabledPlugins | 미존재 | 빈 배열로 초기화 (사용자 추가 가능 구조) | 낮음 |
| extraKnownMarketplaces | 미존재 | 조직별 플러그인 저장소 URL 지원 | 낮음 |
| 플러그인 문서 | 스킬 레퍼런스에만 존재 | 설정 가이드 추가 | 낮음 |

**영향 파일**:
- `internal/template/templates/.claude/settings.json.tmpl`
- `internal/template/settings_test.go`
- `internal/template/templates/.claude/skills/moai-foundation-claude/reference/claude-code-settings-official.md`

#### 2.7 Hook 품질 검증 템플릿

**현재 상태**: TeammateIdle과 TaskCompleted hook은 등록만 되어 있고, 검증 로직이 없다.

**목표 변경사항**:

| 항목 | 검증 항목 | exit code |
|------|-----------|-----------|
| TeammateIdle hook | LSP 오류 수 == 0, 타입 오류 수 == 0, lint 오류 수 == 0 | 0 (허용) / 2 (블록 + stderr 메시지) |
| TaskCompleted hook | SPEC acceptance criteria 검증, 파일 소유권 무결성 | 0 (승인) / 2 (거부 + stderr 메시지) |

**기술적 상세**:

TeammateIdle 검증 흐름:
```
stdin JSON 수신
  -> agent name, team name 추출
  -> moai hook teammate-idle 실행
  -> LSP 진단 수집 (lsp.CollectDiagnostics)
  -> quality gate 검증 (quality.ValidateGates)
  -> 검증 통과: exit 0
  -> 검증 실패: stderr에 실패 사유 출력 + exit 2
```

TaskCompleted 검증 흐름:
```
stdin JSON 수신
  -> task description, agent name 추출
  -> moai hook task-completed 실행
  -> 관련 SPEC 파일 로드
  -> acceptance criteria 매칭
  -> 파일 변경 목록 추출
  -> 파일 소유권 검증 (workflow.yaml의 file_ownership)
  -> 검증 통과: exit 0
  -> 검증 실패: stderr에 거부 사유 출력 + exit 2
```

**영향 파일**:
- `internal/hook/teammate_idle.go` (신규)
- `internal/hook/task_completed.go` (신규)
- `internal/hook/teammate_idle_test.go` (신규)
- `internal/hook/task_completed_test.go` (신규)

#### 2.8 `moai hook subagent-stop` 템플릿 동기화

**현재 상태**: 배포된 스크립트는 수정 완료되었으나, 템플릿 소스 파일과의 동기화 상태 확인 필요.

**목표 변경사항**:

| 항목 | 현재 상태 | 목표 상태 | 복잡도 |
|------|-----------|-----------|--------|
| template 파일 | `handle-subagent-stop.sh.tmpl` 존재 | 내용 검증 및 필요시 업데이트 | 낮음 |
| `make build` | 수동 실행 필요 | 템플릿 변경 후 자동 재빌드 확인 | 낮음 |

**영향 파일**:
- `internal/template/templates/.claude/hooks/moai/handle-subagent-stop.sh.tmpl`
- `internal/template/embedded.go` (`make build` 후 재생성)

---

### LOW 우선순위

#### 2.9 문서 업데이트

| 문서 | 변경 사항 | 복잡도 |
|------|-----------|--------|
| Agent Teams 알려진 제한사항 가이드 | 트러블슈팅 섹션 추가, 일반적 실패 패턴 문서화 | 낮음 |
| Hook 디버깅 가이드 | exit code 2 피드백 패턴, stderr 메시지 형식 | 낮음 |
| Windows ARM64 지원 노트 | v2.1.41에서 추가된 Windows ARM64 지원 명시 | 낮음 |
| CHANGELOG.md | v2.5.0 엔트리 작성 | 낮음 |

**영향 파일**:
- `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md`
- `internal/template/templates/.claude/rules/moai/development/agent-authoring.md`
- `CHANGELOG.md`
- `README.md` / `README.ko.md`

---

### HIGH 우선순위 (v2.1.30 신규 추가)

#### 2.10 Task 메트릭 모니터링 (v2.1.30)

**공식 문서 검증**: Task tool 결과에 `metrics` 객체가 포함됨.

```json
{
  "status": "completed",
  "output": "...",
  "metrics": {
    "tokensUsed": 12450,
    "toolUses": 8,
    "durationSeconds": 45.2
  }
}
```

**목표 변경사항**:

| 항목 | 현재 상태 | 목표 상태 | 복잡도 |
|------|-----------|-----------|--------|
| PostToolUse hook 메트릭 수집 | Task 결과 무시 | tokensUsed/toolUses/durationSeconds 파싱 및 JSONL 로깅 | 중간 |
| `.moai/logs/task-metrics.jsonl` | 미존재 | Task 메트릭 누적 로그 | 낮음 |
| `quality.yaml` 토큰 예산 모니터링 | 없음 | `token_budget.per_task_warning: 20000` 임계값 설정 | 낮음 |
| `/moai insights` 서브커맨드 | 미존재 | 메트릭 분석 리포트 (평균 토큰, 느린 에이전트) | 높음 |

**영향 파일**:
- `internal/hook/post_tool.go` (Task 메트릭 파싱 로직 추가)
- `internal/template/templates/.moai/config/sections/quality.yaml.tmpl` (token_budget 섹션 추가)
- `internal/cli/insights.go` (신규: `/moai insights` 명령)

#### 2.11 `/debug` 명령 연동 (v2.1.30)

**공식 문서 검증**: `/debug`는 세션 상태, hook 실행 로그, tool 트레이스를 표시. `claude --debug "hooks,api"` 필터 지원.

**목표 변경사항**:

| 항목 | 현재 상태 | 목표 상태 | 복잡도 |
|------|-----------|-----------|--------|
| hook 디버깅 가이드 | 없음 | `/debug` 사용법 + `claude --debug hooks` 방법 문서화 | 낮음 |
| CLAUDE.md 트러블슈팅 섹션 | 없음 | 5가지 일반적 문제 + `/debug` 활용법 | 낮음 |
| `expert-debug` 에이전트 | `/debug` 미언급 | 에이전트 지시에 `/debug` 활용 패턴 추가 | 낮음 |

**영향 파일**:
- `internal/template/templates/CLAUDE.md` (트러블슈팅 섹션 추가)
- `internal/template/templates/.claude/agents/moai/expert-debug.md` (디버그 패턴 추가)

---

### MEDIUM 우선순위 (v2.1.30 신규 추가)

#### 2.12 MCP OAuth 자격증명 가이드 (v2.1.30)

**공식 문서 검증**: `.mcp.json`의 `oauth` 필드 지원:

```json
{
  "mcpServers": {
    "slack": {
      "type": "http",
      "url": "https://mcp.slack.com/mcp",
      "oauth": {
        "clientId": "your-client-id",
        "callbackPort": 8080
      }
    }
  }
}
```

**목표 변경사항**:

| 항목 | 현재 상태 | 목표 상태 | 복잡도 |
|------|-----------|-----------|--------|
| MCP OAuth 문서 | 없음 | `.moai/docs/MCP_OAUTH_SETUP.md` 가이드 신규 생성 | 낮음 |
| 예시 `.mcp.json` 템플릿 | 없음 | Slack, GitHub, Sentry OAuth 예시 포함 | 낮음 |

**영향 파일**:
- `internal/template/templates/.moai/docs/MCP_OAUTH_SETUP.md` (신규)

#### 2.13 PDF pages 파라미터 에이전트 가이드 (v2.1.30)

**공식 문서 검증**: Read tool에 `pages: "1-5"` 파라미터 지원. 10페이지 초과 PDF는 `@` 참조 시 경량 참조만 반환.

**목표 변경사항**:

| 항목 | 현재 상태 | 목표 상태 | 복잡도 |
|------|-----------|-----------|--------|
| team-researcher 에이전트 지시 | PDF 처리 미언급 | "50페이지 초과 PDF는 pages 파라미터 사용" 추가 | 낮음 |
| expert-debug 에이전트 지시 | PDF 처리 미언급 | PDF 기반 에러 로그 분석 패턴 추가 | 낮음 |
| moai-workflow-jit-docs 스킬 | PDF 처리 미언급 | pages 파라미터 예시 추가 | 낮음 |

**영향 파일**:
- `internal/template/templates/.claude/agents/moai/team-researcher.md`
- `internal/template/templates/.claude/skills/moai/moai-workflow-jit-docs/` 관련 파일

---

## 3. 필요한 SPEC 문서 목록

각 주요 기능에 대해 SPEC 문서를 작성하여 체계적으로 구현한다.

| SPEC ID | 제목 | 우선순위 | 의존성 | 예상 복잡도 | v2.1.x 기준 |
|---------|------|----------|--------|------------|-------------|
| SPEC-TEAM-001 | Agent Teams 품질 검증 Hook 구현 | CRITICAL | 없음 | 높음 | v2.1.33 |
| SPEC-MEMORY-001 | Agent Persistent Memory 템플릿 정규화 | HIGH | 없음 | 낮음 | v2.1.33 |
| SPEC-MONITOR-001 | Task 메트릭 Hook 수집 (로깅만) | HIGH | 없음 | 낮음 | **v2.1.30** |
| SPEC-SETTINGS-001 | settings.json 기능 확장 (spinnerTips, plugins) | HIGH | 없음 | 낮음 | v2.1.45 |
| SPEC-DOCS-001 | v2.5.0 문서 업데이트 번들 | LOW | 전체 기능 완료 후 | 낮음 | v2.1.30~45 |

> **아키텍처 결정 (2026-02-18)**: CLI 서브커맨드(`moai team list/cleanup`, `moai insights`) 불필요.
> - `moai team cleanup`: SessionEnd 훅에서 자동 처리 (이미 구현됨)
> - `moai team list` / insights: 슬래시 커맨드도 불필요. 필요 시 직접 `~/.claude/teams/` 확인
> - Go 바이너리는 훅 핸들러 역할에만 집중 (단일 책임)

### SPEC-TEAM-001: Agent Teams 품질 검증 Hook 구현

**범위**:
- `internal/hook/teammate_idle.go`: TeammateIdle 이벤트 핸들러 구현
- `internal/hook/task_completed.go`: TaskCompleted 이벤트 핸들러 구현
- LSP 진단 수집 연동 (`internal/lsp/` 활용)
- TRUST 5 품질 게이트 연동 (`internal/core/quality/` 활용)
- SPEC acceptance criteria 파서 구현
- 파일 소유권 검증 로직 (workflow.yaml의 team 섹션 참조)
- exit code 2 + stderr 메시지 패턴 구현
- 단위 테스트 및 계약 테스트

**예상 TAG 체인**:
1. TAG-001: TeammateIdle hook 핸들러 스캐폴딩
2. TAG-002: LSP 진단 수집 연동
3. TAG-003: 품질 게이트 검증 로직
4. TAG-004: TaskCompleted hook 핸들러 스캐폴딩
5. TAG-005: SPEC acceptance criteria 파서
6. TAG-006: 파일 소유권 검증
7. TAG-007: 통합 테스트

### SPEC-TEAM-002: `moai team` CLI 명령 구현

**범위**:
- `internal/cli/team.go`: team 하위 명령 등록
- `moai team list`: 활성 팀/팀메이트 목록 표시
- `moai team cleanup`: 고아 리소스 정리
- `.moai/teams/` 레지스트리 연동
- tmux 세션 관리 연동

**예상 TAG 체인**:
1. TAG-001: team 명령 스캐폴딩 (Cobra subcommand)
2. TAG-002: `moai team list` 구현
3. TAG-003: `moai team cleanup` 구현
4. TAG-004: 테스트 및 문서

### SPEC-MEMORY-001: Agent Persistent Memory 템플릿 정규화

**범위**:
- 전체 27개 agent 정의 파일의 `memory` frontmatter 일관성 확보
- `.gitignore.tmpl`에 `.claude/agent-memory-local/` 제외 규칙 추가
- `moai init`에서 memory 디렉토리 자동 생성
- agent-authoring.md에 memory scope 가이드라인 업데이트

**예상 TAG 체인**:
1. TAG-001: agent 정의 파일 memory 필드 정규화
2. TAG-002: .gitignore 및 deployer 업데이트
3. TAG-003: 문서 업데이트

### SPEC-MONITOR-001: Task 메트릭 모니터링 + `/moai insights` (신규, v2.1.30)

**범위**:
- `internal/hook/post_tool.go`: Task tool 결과에서 `metrics` 객체 파싱
- `.moai/logs/task-metrics.jsonl`: JSONL 형식으로 메트릭 누적 (session_id, agent_type, tokensUsed, toolUses, durationSeconds)
- `quality.yaml.tmpl`: `token_budget.per_task_warning` 임계값 추가
- `internal/cli/insights.go`: `/moai insights` 서브커맨드 (평균 토큰, 느린 에이전트, 비용 추정)

**예상 TAG 체인**:
1. TAG-001: PostToolUse hook에 Task 메트릭 파싱 추가
2. TAG-002: JSONL 메트릭 로거 구현
3. TAG-003: quality.yaml 토큰 예산 스키마 추가
4. TAG-004: `moai insights` 명령 구현

### SPEC-SETTINGS-001: settings.json 기능 확장

**범위**:
- `spinnerTipsOverride` 배열 추가 (MoAI 전용 메시지 8-10개)
- `enabledPlugins` 빈 배열 지원
- `extraKnownMarketplaces` 지원
- settings_test.go 업데이트

**예상 TAG 체인**:
1. TAG-001: settings.json.tmpl에 신규 필드 추가
2. TAG-002: 테스트 업데이트
3. TAG-003: 문서 업데이트

### SPEC-DOCS-001: v2.5.0 문서 업데이트 번들

**범위**:
- Agent Teams 트러블슈팅 가이드
- Hook 디버깅 가이드 (exit code 2 패턴)
- `/debug` 명령 사용법 + CLAUDE.md 트러블슈팅 섹션 (v2.1.30)
- MCP OAuth 설정 가이드 `.moai/docs/MCP_OAUTH_SETUP.md` (v2.1.30)
- PDF pages 파라미터 에이전트 지시 업데이트 (v2.1.30)
- Windows ARM64 지원 노트 (v2.1.41)
- CHANGELOG.md v2.5.0 엔트리
- README 버전 업데이트

---

## 4. 구현 순서 (Implementation Order)

### Phase 1: 기반 작업 (SPEC-MEMORY-001 + SPEC-SETTINGS-001)

```
SPEC-MEMORY-001 ────→ SPEC-SETTINGS-001
(agent memory 정규화)    (settings.json 확장)
        \                    /
         \                  /
          +──────+─────────+
                 |
                 v
           make build
           (embedded.go 재생성)
```

**이유**: 두 SPEC 모두 독립적이며 템플릿 변경만 수반한다. 병렬 진행 가능.

**예상 소요**: 각 1-2 세션

### Phase 2: 핵심 기능 (SPEC-TEAM-001)

```
SPEC-TEAM-001
(Hook 품질 검증)
      |
      |  TAG-001~003: TeammateIdle
      |  TAG-004~006: TaskCompleted
      |  TAG-007: 통합 테스트
      |
      v
    make build + go test -race ./...
```

**이유**: Agent Teams 프로덕션 준비의 핵심. Go 소스 코드 변경이 가장 많다.

**예상 소요**: 3-4 세션

### Phase 3: CLI 확장 (SPEC-TEAM-002)

```
SPEC-TEAM-002
(moai team 명령)
      |
      |  의존: SPEC-TEAM-001의 팀 레지스트리 구조
      |
      v
    make build + go test -race ./...
```

**이유**: SPEC-TEAM-001에서 정의한 팀 데이터 구조에 의존한다.

**예상 소요**: 1-2 세션

### Phase 4: 문서 및 릴리즈 (SPEC-DOCS-001)

```
SPEC-DOCS-001
(문서 업데이트 번들)
      |
      |  의존: 모든 기능 구현 완료
      |
      v
    CHANGELOG.md + README + 릴리즈 노트
```

**예상 소요**: 1 세션

### 전체 의존성 다이어그램

```
Phase 1 (병렬)                Phase 2           Phase 3         Phase 4
+──────────────────────+    +──────────────+   +──────────────+  +──────────────+
|SPEC-MEMORY-001       |    |SPEC-TEAM-001 |   |SPEC-TEAM-002 |  |SPEC-DOCS-001 |
|SPEC-SETTINGS-001     | →  |              | → |              |→ |  (+ v2.1.30  |
|SPEC-MONITOR-001 (신규)|    |              |   |              |  |   문서 추가) |
+──────────────────────+    +──────────────+   +──────────────+  +──────────────+
```

---

## 5. Breaking Changes

### v2.5.0에는 Breaking Changes가 없다

| 변경 항목 | 하위 호환성 | 이유 |
|-----------|------------|------|
| spinnerTipsOverride 추가 | 호환 | 신규 필드, 기존 설정 영향 없음 |
| enabledPlugins 추가 | 호환 | 신규 필드, 빈 배열 기본값 |
| agent memory 필드 추가 | 호환 | `memory`는 선택적 frontmatter 필드 |
| TeammateIdle hook 로직 | 호환 | 기존 exit 0 → 조건부 exit 2 (팀 모드에서만 영향) |
| TaskCompleted hook 로직 | 호환 | 기존 exit 0 → 조건부 exit 2 (팀 모드에서만 영향) |
| moai team 명령 추가 | 호환 | 신규 서브커맨드, 기존 명령 영향 없음 |

### 잠재적 행동 변화 (주의 필요)

1. **TeammateIdle hook이 exit 2를 반환할 수 있음**: 팀 모드에서 품질 게이트를 통과하지 못한 팀메이트는 idle 상태로 전환되지 않고 계속 작업하도록 지시받는다. 이는 의도된 동작이며 팀 품질을 향상시키기 위한 것이다.

2. **TaskCompleted hook이 exit 2를 반환할 수 있음**: 팀 모드에서 acceptance criteria를 충족하지 않은 작업은 완료로 인정되지 않는다. 팀메이트에게 stderr를 통해 구체적인 거부 사유가 전달된다.

---

## 6. 마이그레이션 가이드 (v2.4.x → v2.5.0)

### 자동 마이그레이션 (`moai update`)

v2.4.x에서 v2.5.0으로의 업데이트는 `moai update` 명령으로 자동 처리된다.

**자동 처리 항목**:
1. 바이너리 교체 (기존 self-update 메커니즘)
2. 템플릿 파일 3-way merge (manifest 기반)
3. settings.json.tmpl 신규 필드 추가
4. agent 정의 파일 memory 필드 추가
5. hook 스크립트 업데이트

**사용자 확인 필요 항목**:
1. `.claude/settings.json`이 user_modified 상태인 경우 수동 머지 필요 가능
2. 커스텀 agent 정의 파일은 자동 업데이트 대상이 아님 (user_created)

### 수동 마이그레이션 단계

`moai update`를 사용하지 않는 경우:

```bash
# 1. 최신 바이너리 설치
go install github.com/modu-ai/moai-adk-go/cmd/moai@v2.5.0

# 2. 템플릿 재배포 (기존 설정 유지)
moai update -t

# 3. 변경 확인
moai doctor
```

### 설정 파일 변경사항

**settings.json 신규 필드** (자동 추가됨):
```json
{
  "spinnerTipsOverride": [
    "Analyzing SPEC requirements...",
    "Running DDD analysis cycle...",
    "Checking TRUST 5 quality gates...",
    "Delegating to expert agent...",
    "Generating characterization tests...",
    "Synchronizing documentation...",
    "Planning implementation strategy...",
    "Validating acceptance criteria..."
  ],
  "enabledPlugins": [],
  "extraKnownMarketplaces": []
}
```

**workflow.yaml 변경 없음**: 기존 `default_model: sonnet` 설정은 Claude Code가 최신 모델로 자동 해석하므로 변경 불필요.

### 팀 모드 사용자 주의사항

v2.5.0부터 TeammateIdle/TaskCompleted hook이 품질 검증을 수행하므로:

1. **팀 모드 실행 전**: `moai doctor`로 LSP 서버 상태 확인 권장
2. **품질 게이트 실패 시**: 팀메이트에게 stderr로 구체적 실패 사유 전달
3. **비활성화 방법**: quality.yaml에서 `enforce_quality: false`로 설정하면 hook이 항상 exit 0 반환

---

## 7. 성공 지표 (Success Metrics)

### 릴리즈 게이트 체크리스트

v2.5.0 릴리즈 전 다음 기준을 모두 충족해야 한다:

| 기준 | 목표 | 측정 방법 |
|------|------|-----------|
| 전체 테스트 통과 | `go test -race ./...` 100% 통과 | CI 파이프라인 |
| 테스트 커버리지 | 85% 이상 (전체), 90% 이상 (신규 코드) | `go test -cover ./...` |
| 린트 통과 | `golangci-lint run ./...` 경고 0건 | CI 파이프라인 |
| Hook 계약 테스트 | 6개 플랫폼 타겟 전체 통과 | `go test ./internal/hook/...` |
| JSON 안전성 테스트 | settings.json 생성 → 파싱 → 재직렬화 일치 | `go test ./internal/template/...` |
| 바이너리 크기 | 30MB 이하 | `ls -la bin/moai` |
| 매뉴얼 QA | TeammateIdle hook exit 2 시나리오 검증 | 수동 테스트 |
| 매뉴얼 QA | TaskCompleted hook exit 2 시나리오 검증 | 수동 테스트 |
| 매뉴얼 QA | `moai team list` / `moai team cleanup` 동작 확인 | 수동 테스트 |
| `moai update` 호환성 | v2.4.7 → v2.5.0 업데이트 무손실 | 통합 테스트 |

### 품질 지표

| 지표 | v2.4.7 (현재) | v2.5.0 (목표) |
|------|--------------|--------------|
| Hook 이벤트 커버리지 | 15개 hook 등록, 검증 로직 0개 | 15개 hook 등록, 핵심 2개 검증 구현 |
| Agent memory 일관성 | 약 50% (일부만 설정) | 100% (전체 27개 agent) |
| settings.json 필드 완전성 | spinnerTips만 | spinnerTips + plugins + marketplaces |
| CLI 명령 수 | 12개 | 16개 (+moai team list, +moai team cleanup, +moai insights) |
| 팀 모드 프로덕션 준비 | 실험적 | 프로덕션 준비 완료 |
| Task 메트릭 수집 | 없음 | PostToolUse hook에서 자동 수집 + JSONL 로깅 |
| 관측성 | `/debug` 미연동 | CLAUDE.md 트러블슈팅 + MCP OAuth 가이드 |

### 회귀 방지 지표

| 지표 | 목표 | 측정 방법 |
|------|------|-----------|
| Hook 회귀 커밋 | 릴리즈 사이클당 0건 | Git log 분석 |
| settings.json 생성 실패 | 0건 | 계약 테스트 |
| 파괴적 업데이트 덮어쓰기 | 0건 | Manifest 기반 검증 |
| 팀 리소스 누수 | 0건 (고아 tmux/디렉토리) | SessionEnd hook 검증 |

---

## 8. 기술 스택 변경 요약

### 신규 의존성

**없음.** v2.5.0은 기존 의존성만으로 구현 가능하다.

### Go 소스 코드 변경 범위

| 패키지 | 변경 유형 | 파일 수 |
|--------|-----------|---------|
| `internal/hook/` | 신규 파일 2개 + 테스트 2개 | 4 |
| `internal/cli/` | 신규 파일 1개 (team.go) + 테스트 | 2 |
| `internal/template/` | deployer.go 수정, settings_test.go 수정 | 2 |
| `internal/config/` | types.go 확장 (file_ownership 스키마) | 1 |

### 템플릿 변경 범위

| 디렉토리 | 변경 유형 | 파일 수 |
|----------|-----------|---------|
| `.claude/agents/moai/` | memory 필드 정규화 | ~15개 (미설정 agent) |
| `.claude/settings.json.tmpl` | 신규 필드 추가 | 1 |
| `.claude/rules/moai/` | 문서 업데이트 | 3-5개 |
| `.claude/hooks/moai/` | 검증 완료 (변경 불필요 예상) | 0 |
| `.gitignore.tmpl` | memory 디렉토리 제외 | 1 |

---

## 9. 위험 요소 및 대응 방안

| 위험 | 영향도 | 발생 확률 | 대응 방안 |
|------|--------|-----------|-----------|
| TeammateIdle hook이 너무 엄격하여 팀 워크플로우 차단 | 높음 | 중간 | quality.yaml에 `enforce_quality: false` 비활성화 옵션 제공 |
| LSP 서버 미설치 환경에서 hook timeout | 중간 | 중간 | LSP 사용 불가 시 graceful degradation (경고만 출력, exit 0) |
| 기존 사용자의 settings.json 커스터마이징과 충돌 | 중간 | 낮음 | 3-way merge + .conflict 파일 생성 패턴 유지 |
| Windows 환경에서 tmux 세션 관리 불가 | 낮음 | 높음 | Windows에서는 `moai team cleanup`이 디렉토리만 정리, tmux 스킵 |
| agent memory 디렉토리 권한 문제 | 낮음 | 낮음 | `moai init`에서 디렉토리 생성 시 적절한 권한 설정 |

---

## 10. 승인 요청

### 의사결정 필요 항목

1. **TeammateIdle hook의 기본 동작**:
   - 옵션 A: 기본 활성화 (검증 실패 시 exit 2) -- **권장**
   - 옵션 B: 기본 비활성화 (사용자가 opt-in)
   - 권장 이유: 팀 모드의 핵심 가치는 품질 보장이며, graceful degradation으로 LSP 미설치 환경도 지원

2. **`moai team` 명령의 하위 구조**:
   - 옵션 A: `moai team list` / `moai team cleanup` -- **권장**
   - 옵션 B: `moai team ls` / `moai team gc`
   - 권장 이유: 기존 `moai worktree` 명령과 일관된 네이밍 (`list`, `clean`)

3. **spinnerTipsOverride 메시지 언어**:
   - 옵션 A: 영어만 -- **권장**
   - 옵션 B: conversation_language에 따라 다국어
   - 권장 이유: Claude Code spinner는 영어 기반이며, 다국어는 복잡성 대비 가치가 낮음

### 승인 체크리스트

- [ ] 기능 로드맵 승인
- [ ] SPEC 문서 목록 승인
- [ ] 구현 순서 승인
- [ ] Breaking changes 평가 승인
- [ ] 마이그레이션 가이드 승인
- [ ] 성공 지표 승인

---

## 11. 다음 단계

승인 후 진행 순서:

1. **SPEC 문서 생성**: Phase 1의 SPEC-MEMORY-001과 SPEC-SETTINGS-001 작성
2. **Phase 1 구현**: 병렬로 agent memory 정규화와 settings.json 확장 진행
3. **Phase 2 구현**: SPEC-TEAM-001로 hook 품질 검증 로직 구현
4. **Phase 3 구현**: SPEC-TEAM-002로 `moai team` CLI 명령 구현
5. **Phase 4 문서**: SPEC-DOCS-001로 전체 문서 업데이트
6. **릴리즈 준비**: CHANGELOG 작성, 버전 태깅, goreleaser 실행

**예상 전체 소요**: 8-12 세션 (Phase별 2-4 세션)

---

> **참고**: 이 플랜은 Claude Code v2.1.30~v2.1.45 릴리즈 노트와 공식 문서(https://code.claude.com/docs)를 기반으로 한다. 추후 v2.1.46+에서 추가 기능이 출시되면 별도 업데이트 플랜으로 관리한다.

---

## 12. v2.1.30 항목 공식 문서 검증 요약

| 항목 | 공식 문서 확인 | 실제 스키마 | moai-adk 적용 |
|------|-------------|------------|--------------|
| `reducedMotion` | ✅ `prefersReducedMotion: false` | settings.json 사용자 설정 | ❌ 템플릿 불필요 (사용자 취향) |
| MCP OAuth 자격증명 | ✅ `.mcp.json`에 `oauth.clientId`, `oauth.callbackPort` | `--client-secret` 시스템 키체인 저장 | ✅ SPEC-DOCS-001 (가이드 문서) |
| PDF pages 파라미터 | ✅ Read tool `pages: "1-5"` | 10페이지+ PDF는 경량 참조 반환 | ✅ SPEC-DOCS-001 (에이전트 지시 업데이트) |
| Task 메트릭 | ✅ `metrics.tokensUsed`, `.toolUses`, `.durationSeconds` | Task tool 결과 자동 포함 | ✅ **SPEC-MONITOR-001 (신규)** |
| `/debug` 명령 | ✅ `claude --debug "hooks,api"` 필터 지원 | 세션 트레이스, hook 로그, tool 실행 추적 | ✅ SPEC-DOCS-001 (CLAUDE.md 트러블슈팅) |

---

**버전**: 1.0.0
**작성자**: manager-strategy
**최종 수정**: 2026-02-18
