# MoAI-ADK Agent 시스템

## 🤖 에이전트 개요
MoAI-ADK는 개발 파이프라인의 각 단계를 자동화하기 위해 집중된 역할을 가진 에이전트들을 제공합니다. 모든 에이전트는 `.claude/agents/moai/` 디렉터리에 존재하며, 커맨드가 명시적으로 위임할 때만 동작합니다.

| 에이전트 | 역할 | 주요 커맨드 |
| --- | --- | --- |
| `project-manager` | 프로젝트 킥오프 인터뷰, 설정/브레인스토밍 옵션 관리 | `/moai:0-project` |
| `cc-manager` | Claude Code 권한/훅/환경 점검 | `/moai:0-project` 실행 직후 자동 |
| `spec-builder` | SPEC 자동 제안 및 문서 생성 | `/moai:1-spec` |
| `code-builder` | TDD 구현(RED→GREEN→REFACTOR) | `/moai:2-build` |
| `doc-syncer` | 문서/PR 동기화, TAG 인덱스 관리 | `/moai:3-sync` |
| `git-manager` | 체크포인트/브랜치/커밋/동기화 전담 | `/moai:git:*` 계열 |
| `codex-bridge`* | Codex CLI headless 호출 (선택) | 브레인스토밍 활성화 시 |
| `gemini-bridge`* | Gemini CLI headless 호출 (선택) | 브레인스토밍 활성화 시 |


asterisk(*) 표시된 에이전트는 `.moai/config.json.brainstorming` 설정이 활성화되어 있고 CLI가 설치된 경우에만 사용됩니다.

## 🧭 에이전트별 상세 설명

### project-manager
- `/moai:0-project` 실행 시 신규/레거시 감지, product/structure/tech 인터뷰 진행
- Codex/Gemini CLI 설치 여부를 확인하고 사용자의 동의가 있으면 설치/로그인 방법만 안내 (자동 실행 금지)
- 외부 브레인스토밍 사용 여부를 결정하고 `.moai/config.json.brainstorming`(`enabled`, `providers`) 값을 업데이트하며 `providers`에는 항상 "claude"를 포함

### cc-manager
- Guard/훅 권한을 점검하고 Claude Code 환경을 최적화
- project-manager 이후 실행되어 설정 충돌을 예방

### spec-builder
- 프로젝트 문서를 요약하여 SPEC 후보를 제안하고 문서를 생성
- 팀 모드에서는 GitHub Issue로 변환할 수 있도록 정보를 제공

### code-builder
- TDD 사이클을 강제하며 단계별 체크포인트/커밋 전략을 수행
- 외부 브레인스토밍이 활성화된 경우 codex-bridge/gemini-bridge 출력과 Claude 제안을 비교하여 최적안을 선택 (예: `Task: use codex-bridge to run "codex exec --model gpt-5-codex 'List refactor risks'"`)

### doc-syncer
- 코드 ↔ 문서 일관성을 유지하고 16-Core @TAG를 업데이트
- 팀 모드에서 Draft → Ready 전환, 리뷰어 할당 등의 PR 절차를 보조

### git-manager
- 브랜치 생성, 체크포인트, 커밋, 동기화, 롤백 등 모든 Git 작업을 담당
- `/moai:git:*` 명령에서 직접 호출되며 다른 에이전트가 Git 명령을 수행하지 않도록 보장

### codex-bridge (선택)
- `codex exec --model gpt-5-codex "..."` 형태의 headless 호출을 담당
- 프로젝트 컨텍스트를 요약하고 결과를 구조화하여 Claude 세션에 보고
- 사용 예: `Task: use codex-bridge to run "codex exec --model gpt-5-codex 'Summarize design risks'"`
- 설치되지 않은 경우 `npm install -g @openai/codex` 또는 `brew install codex` 명령을 안내만 함 (자동 설치 금지)

### gemini-bridge (선택)
- `gemini -m gemini-2.5-pro -p "..." --output-format json` 호출을 담당
- JSON 출력으로 구조화된 브레인스토밍/리뷰 결과를 제공
- 사용 예: `Task: use gemini-bridge to run "gemini -m gemini-2.5-pro -p 'List alternative solution paths' --output-format json"`
- 설치되지 않은 경우 `npm install -g @google/gemini-cli` 또는 `brew install gemini-cli` 명령을 안내만 함 (자동 설치 금지)

## ⚙️ 브레인스토밍 설정(.moai/config.json)

```json
{
  "brainstorming": {
    "enabled": true,
    "providers": ["claude", "codex", "gemini"]
  }
}
```

- `enabled`: `true` 인 경우 외부 AI 브레인스토밍을 활성화
- `providers`: 사용할 엔진을 배열로 지정하며 항상 "claude"를 포함하고 필요에 따라 `codex`, `gemini`를 추가
- project-manager가 `/moai:0-project` 인터뷰 중에 사용 여부를 묻고 값을 갱신
- 다른 커맨드는 이 설정을 읽어 외부 브리지 에이전트를 호출할지 판단합니다.

## 🔄 실행 흐름 요약

```mermaid
flowchart TD
    A[/moai:0-project] -->|project-manager| B[/moai:1-spec]
    B -->|spec-builder| C[/moai:2-build]
    C -->|code-builder| D[/moai:3-sync]
    B -. git tasks .-> G[git-manager]
    C -. git tasks .-> G
    D -. git tasks .-> G
```

외부 브레인스토밍이 활성화된 경우, 각 단계에서 필요한 시점에 `codex-bridge` 와 `gemini-bridge` 가 추가로 호출됩니다.

## 🧪 에이전트 사용 시 체크리스트
- 각 에이전트는 **단일 책임**만 수행하고 다른 에이전트를 직접 호출하지 않습니다.
- 커맨드에서 순차/병렬 흐름을 명시적으로 관리하여 의존성 충돌을 방지합니다.
- 외부 CLI 호출 전에는 항상 `which`/`--version` 으로 설치 여부를 확인하고, 사용자 동의 없이 설치하거나 로그인하지 않습니다.

## 🔧 커스터마이징 가이드

### 모델 변경
- 에이전트 단위 모델 지정은 각 에이전트 마크다운의 `model` 필드에서 수행합니다.
- project-manager, spec-builder, code-builder, doc-syncer, git-manager, codex-bridge, gemini-bridge 기본값은 `sonnet` 입니다.

### 설정 정책 문서화
```markdown
# .claude/memory/team_conventions.md
- 외부 브레인스토밍은 `/moai:0-project update` 인터뷰에서 명시적으로 승인한다.
- Codex/Gemini CLI는 개발자 로컬 환경에만 설치하고 CI에서는 비활성화한다.
```

## 📚 참고 문서
- [Codex CLI 공식 문서](https://developers.openai.com/codex/cli/)
- [OpenAI Codex GitHub 레포](https://github.com/openai/codex)
- [Gemini CLI 공식 가이드](https://developers.google.com/gemini-code-assist/docs/gemini-cli)
- [Google Gemini CLI GitHub 레포](https://github.com/google-gemini/gemini-cli)
