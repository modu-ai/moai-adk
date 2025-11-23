# MoAI-ADK 로컬 Claude Code 개발 가이드

## 개발 워크플로우

### 1.1 작업 위치 규칙

**모든 개발 작업은 다음 위치에서 수행:**

```
/Users/goos/MoAI/MoAI-ADK/src/moai_adk/
├── .claude/                 # Claude Code 설정
├── .moai/                   # MoAI 프로젝트 메타데이터
├── templates/               # 프로젝트 템플릿
└── [여타 소스 코드]
```

**작업 후 로컬 프로젝트로 동기화:**

```
/Users/goos/MoAI/MoAI-ADK/
├── .claude/                 # 동기화됨
├── .moai/                   # 동기화됨
└── [소스 코드 및 문서]
```

### 1.2 개발 사이클

```
1. 소스 프로젝트에서 작업 (/src/moai_adk/...)
   ↓
2. 로컬 프로젝트에 동기화 (./)
   ↓
3. 로컬 프로젝트에서 테스트 및 검증
   ↓
4. Git 커밋 (로컬 프로젝트에서)
```

---

## 파일 동기화 규칙

### 2.1 동기화 대상 디렉토리

**자동 동기화 필요 영역:**

```
src/moai_adk/.claude/    ↔  .claude/
src/moai_adk/.moai/      ↔  .moai/
src/moai_adk/templates/  ↔  ./
```

### 2.2 동기화 제외 (로컬 전용)

**절대 동기화하지 않을 파일:**

```
.claude/commands/moai/99-release.md          # 로컬 릴리스 커맨드만
.claude/settings.local.json                  # 개인 설정
.claude/hooks/                               # 개발용 hooks (패키지에 포함 금지)
.CLAUDE.local.md                             # 이 파일
.moai/cache/                                 # 캐시 파일
.moai/logs/                                  # 로그 파일
.moai/config/config.json                     # 개인 프로젝트 설정
```

### 2.3 동기화 도구

**사용할 도구:**

```bash
# 수동 동기화 (rsync 사용)
rsync -avz \
  --exclude=".DS_Store" \
  --exclude="*.pyc" \
  --exclude="__pycache__" \
  --exclude=".cache" \
  src/moai_adk/.claude/ .claude/

rsync -avz \
  --exclude=".DS_Store" \
  --exclude="*.pyc" \
  --exclude="__pycache__" \
  --exclude="cache/" \
  --exclude="logs/" \
  --exclude="config/" \
  src/moai_adk/.moai/ .moai/
```

### 2.4 스크립트 기반 동기화

**동기화 스크립트 위치:**

```
.moai/scripts/sync-from-src.sh
```

**실행:**

```bash
bash .moai/scripts/sync-from-src.sh
```

---

## 코드 작성 표준

### 3.1 언어 규칙

**모든 코드 작업:**

- ✅ **영문으로만 작성**
- ✅ 변수명: camelCase 또는 snake_case (언어별 관례)
- ✅ 함수명: camelCase (JavaScript/Python) 또는 PascalCase (C#/Java)
- ✅ 클래스명: PascalCase (모든 언어)
- ✅ 상수명: UPPER_SNAKE_CASE (모든 언어)

**주석과 문서:**

- ✅ **모든 주석은 영문**
- ✅ JSDoc, docstring 등 모두 영문
- ✅ Commit messages: 영문 (또는 한글 + 영문 혼용 시 format: 영문)

**이 파일 (@CLAUDE.local.md):**

- ✅ **한글로 작성** (로컬 작업 지침이므로)
- ✅ Git 추적 대상

### 3.2 주석 표준 (영문)

- 모든 코드, 출력메시지, 주석은 영문으로 작성

### 3.3 금지 사항

```python
# ❌ WRONG - Korean comments
def calculate_score():  # 점수 계산
    score = 100  # 최종 점수
    return score

# ✅ CORRECT - English comments
def calculate_score():  # Calculate final score
    score = 100  # Final score value
    return score
```

---

## 로컬 전용 파일 관리

### 5.1 로컬 전용 파일 목록

**절대 패키지에 동기화하지 않을 파일:**

| 파일                  | 위치                     | 용도               | Git 추적 |
| --------------------- | ------------------------ | ------------------ | -------- |
| `99-release.md`       | `.claude/commands/moai/` | 로컬 릴리스 커맨드 | ✅ Yes   |
| `CLAUDE.local.md`     | 루트                     | 로컬 개발 지침     | ✅ Yes   |
| `settings.local.json` | `.claude/`               | 개인 설정          | ❌ No    |
| `cache/`              | `.moai/`                 | 캐시 파일          | ❌ No    |
| `logs/`               | `.moai/`                 | 로그 파일          | ❌ No    |
| `config/config.json`  | `.moai/`                 | 개인 설정          | ❌ No    |

### 5.2 로컬 릴리스 커맨드

**.claude/commands/moai/99-release.md (로컬만):**

```markdown
# Local Release Management

This command is only for local development and testing.
It manages MoAI-ADK package releases locally.

## Features

- Version management
- Pre-release testing
- Local deployment simulation
- Changelog generation

## Usage

> /moai:99-release

This command is NOT synchronized to the package.
```

---

### 6.3 Git 작업 체크리스트

**커밋 전:**

- [ ] 모든 코드가 영문으로 작성됨
- [ ] 주석과 docstring이 영문임
- [ ] 로컬 전용 파일이 포함되지 않음
- [ ] 테스트가 통과함
- [ ] Linting이 통과함 (ruff, pylint, etc.)

**푸시 전:**

- [ ] 브랜치가 최신 개발 버전으로 rebase됨
- [ ] 커밋이 논리적 단위로 정리됨
- [ ] 커밋 메시지가 표준 포맷을 따름

**PR 전:**

- [ ] 문서가 동기화됨
- [ ] SPEC이 업데이트됨 (필요시)
- [ ] 변경사항이 설명됨

---

## 자주 사용하는 명령어

### 동기화

```bash
# 소스에서 로컬로 동기화
bash .moai/scripts/sync-from-src.sh

# 특정 디렉토리만 동기화
rsync -avz src/moai_adk/.claude/ .claude/
rsync -avz src/moai_adk/.moai/ .moai/
```

### 검증

```bash
# 코드 품질 확인
ruff check src/
mypy src/

# 테스트 실행
pytest tests/ -v --cov

# 문서 검증
python .moai/tools/validate-docs.py
```

---

## CLAUDE.md 작성 및 유지보수 가이드

### 개요

이 가이드는 MoAI-ADK의 CLAUDE.md 파일을 작성하고 유지보수하는 방법을 설명합니다.
MoAI 프레임워크 자체를 개발하는 개발자를 위한 문서입니다.

---

### CLAUDE.md의 본질

**중요**: CLAUDE.md는 **코드가 아닙니다**. CLAUDE.md는 **Alfred의 기본 실행 지침**입니다.

- ✅ **용도**: Claude Code agents를 위한 orchestration 규칙
- ❌ **용도 아님**: 사용자 가이드, 구현 가이드, 튜토리얼
- 👥 **대상**: Claude Code (agents, commands, hooks)
- ❌ **대상 아님**: 최종 사용자

**CLAUDE.md vs. 다른 문서**:

| 문서               | 용도                | 대상              |
| ------------------ | ------------------- | ----------------- |
| CLAUDE.md          | Alfred 실행 규칙    | Agents/Commands   |
| README.md          | 프로젝트 개요       | End users         |
| Skill SKILL.md     | 패턴/지식 캡슐      | Agents/Developers |
| .moai/memory/\*.md | 실행 규칙 참고 문서 | Agents/Developers |
| CLAUDE.local.md    | 로컬 작업 지침      | Local developers  |

---

### 1. CLAUDE.md 구조 표준

모든 CLAUDE.md 파일은 다음 8개 섹션을 **필수**로 포함해야 합니다:

#### I. 목적 & 범위 (Required)

```markdown
# [PROJECT]: Claude Code 실행 가이드

**목적**: [PROJECT]의 Super Agent Orchestrator 실행 매뉴얼
**대상**: Claude Code (agents, commands), 최종 사용자 아님
**철학**: [철학 문구]
```

**반드시 포함할 것**:

- ✅ 명확한 목적 선언
- ✅ "대상: Claude Code agents"
- ✅ "NOT for end users"
- ✅ 범위 내/외 명시

#### II. 핵심 원칙 (Required)

3-5개의 기본 운영 규칙:

```markdown
## 핵심 원칙

1. **[원칙명]** - 설명
2. **[원칙명]** - 설명
3. **[원칙명]** - 설명
```

#### III. 설정 통합 (조건부)

Config.json과의 연결:

```markdown
## 설정 통합

이 문서가 읽는 config 필드:

- `github.spec_git_workflow` - Git 워크플로우 스타일
- `constitution.test_coverage_target` - 품질 게이트 임계값

### Config 필드 명세

**필드**: `github.spec_git_workflow`

- **위치**: config.json → github.spec_git_workflow
- **타입**: String (enum)
- **가능값**: develop_direct, feature_branch, per_spec
- **기본값**: develop_direct
- **우선순위**: Priority 1 (최상위)
- **영향**: Git branch 생성 여부 제어
```

#### IV. Auto-Trigger 규칙 (조건부)

Agent/Command가 자동으로 실행되는 조건:

````markdown
## Agent: [AGENT_NAME] - Auto-Trigger 규칙

### Trigger 활성화 포인트

| Phase | 이벤트       | 조건 | Config 필드                    | 위임 패턴 |
| ----- | ------------ | ---- | ------------------------------ | --------- |
| PLAN  | /moai:1-plan | 항상 | language.conversation_language | 직접 호출 |
| RUN   | /moai:2-run  | 항상 | constitution.enforce_tdd       | 직접 호출 |

### Trigger 로직 (Pseudo-code)

```python
def should_trigger(event, config):
    if event.type == "moai:1-plan":
        return True  # 항상 trigger
    elif event.type == "vague_request":
        return measure_clarity(event) < 70%
    return False
```

### 전달 Context

Trigger 시 다음 정보 전달:

1. `user_request` - 원본 사용자 요청
2. `current_phase` - 현재 phase (PLAN/RUN/SYNC)
3. `config` - 사용자 config.json
4. `previous_results` - 이전 phase 결과 (있는 경우)
````

#### V. 위임 계층 (Required)

어떤 agent를 언제 호출할지:

```markdown
## 위임 계층

- **spec-builder**: SPEC 생성 및 분석

  - 조건: /moai:1-plan 실행
  - Context: 사용자 요청 + config

- **git-manager**: Git 브랜치 생성
  - 조건: spec_git_workflow != "develop_direct"
  - Context: SPEC ID + git config

### 위임 오류 처리

git-manager 호출 실패 시:

1. 로그 남기기
2. 사용자에게 AskUserQuestion으로 선택 제시
3. 선택 기반 retry 또는 skip
```

#### VI. 품질 게이트 (Required)

TRUST 5 또는 유사 기준:

```markdown
## 품질 게이트 (TRUST 5)

### Test-first

**기준**: ≥ 85% 테스트 커버리지
**검증**: pytest --cov=src/ | grep "Coverage"
**실패**: PR 차단, 커버리지 갭 보고

### Readable

**기준**: 명확한 네이밍 (모호한 약자 없음)
**검증**: ruff linter 자동 검사
**실패**: 경고 (차단 아님)

### Unified

**기준**: 프로젝트 패턴 준수 (일관된 스타일)
**검증**: black, isort 자동 체크
**실패**: 자동 포맷 또는 경고

### Secured

**기준**: OWASP 보안 검사 통과
**검증**: security-expert agent 검수 (필수)
**실패**: PR 차단

### Trackable

**기준**: 명확한 commit 메시지 + 테스트 증거
**검증**: Git commit message regex 검증
**실패**: 메시지 포맷 제안
```

#### VII. 참고 문서 (Required)

외부 문서 참조:

```markdown
## 참고 문서

### 필수 참조

- @.moai/memory/execution-rules.md - 실행 제약사항
- @.moai/memory/agents.md - Agent 카탈로그
- @.moai/config/config.json - Config 스키마

### 권장 참조

- Skill("moai-spec-intelligent-workflow") - SPEC 결정 로직
- Skill("moai-cc-configuration") - Config 관리
- @.moai/memory/token-optimization.md - 토큰 예산
```

**참조 형식 (반드시 이 형식 사용)**:

- ✅ `@.moai/memory/agents.md` (파일 참조)
- ✅ `Skill("moai-cc-commands")` (Skill 참조)
- ✅ `/moai:1-plan` (Command 참조)
- ❌ `.moai/memory/agents.md` (@ 누락)
- ❌ `moai-cc-commands` (Skill() 미포장)

#### VIII. 빠른 참조 & 예제 (Required)

실제 사용 예제:

````markdown
## 예제 시나리오 1: Personal + develop_direct

**설정**:

```json
{
  "git_strategy": { "mode": "personal" },
  "github": { "spec_git_workflow": "develop_direct" }
}
```
````

**예상 동작**:

- ✅ /moai:1-plan SPEC 파일 생성
- ✅ git-manager 호출 안됨
- ✅ 브랜치 생성 안됨
- ✅ 현재 브랜치에서 직접 커밋 가능

---

### 2. 금지 사항 (CLAUDE.md에 포함하면 안됨)

❌ **절대 포함하지 말 것**:

- ❌ 사용자 가이드 또는 튜토리얼
- ❌ 구현 코드 예제 (흐름도 제외)
- ❌ 마케팅 언어
- ❌ Skills/memory/에 이미 있는 내용 복제
- ❌ API 구현 상세 (Skills 참조)
- ❌ 하드코딩된 시크릿이나 자격증명

---

### 3. 작성 스타일 가이드라인

#### 톤 & 음성

- ✅ 직접적, 기술적, 명확
- ✅ 명령조: "Alfred MUST NOT directly execute tasks" (소극적 아님)
- ✅ 완전성 > 간결성
- ✅ 용어 첫 사용 시 정의

**나쁜 예**:

```text
Alfred는 아마도 작업을 실행해야 할 것 같습니다.
```

**좋은 예**:

```text
Alfred DOES NOT execute tasks directly. Alfred DELEGATES to specialized agents.
```

#### 기술 명확성

| 상황 | 형식 |
|------|------|
| 결정 매트릭스 (3개 이상) | 표 사용 |
| 복잡한 로직 | ASCII 흐름도 또는 Pseudo-code |
| Config 예제 | 전체 JSON/YAML 블록 |
| 규칙/제약사항 | 불릿 리스트 |
| 순차 절차 | 번호 리스트 |

**Pseudo-code 사용 OK인 경우**:

```python
# OK: 결정 로직 보여줌
if config["spec_git_workflow"] == "develop_direct":
    TRIGGER_GIT_MANAGER = False
else:
    TRIGGER_GIT_MANAGER = True
```

**구현 코드는 Skills 참조**:

```markdown
# WRONG

def validate_configuration(config):
schema = ConfigSchema()
return schema.validate(config)

# RIGHT

검증은 moai-cc-configuration Skill에서 처리합니다.
자세한 내용: @.moai/memory/configuration-validation.md
```

---

### 4. CLAUDE.md 검증 체크리스트

**CLAUDE.md 커밋 전 필수 확인**:

- [ ] **목적 명확**: "Alfred의 기본 실행 지침"으로 시작
- [ ] **대상 명시**: "Claude Code agents를 위한 문서"
- [ ] **8개 섹션**: 모두 포함 (또는 조건부 섹션 제외 정당화)
- [ ] **복제 없음**: Skills/memory/와 중복 내용 없음
- [ ] **Config 참조 유효**: 모든 필드가 schema에 존재
- [ ] **Agent 이름 정확**: .claude/agents/에 존재하는 agent만
- [ ] **외부 참조 형식**: `@.moai/` 또는 `Skill()` 형식
- [ ] **예제 유효성**: JSON/YAML 예제가 문법적으로 정확
- [ ] **시크릿 없음**: API 키, 자격증명 없음
- [ ] **종료 명시**: "Claude Code 실행을 위한 문서"로 종료

---

### 5. 메모리/참고 문서 표준

`.moai/memory/` 문서 구조:

```markdown
# [제목]

**목적**: 한 줄 목적 (30자 이내)
**대상**: [Agents / Humans / Developers]
**최종 업데이트**: YYYY-MM-DD
**버전**: X.Y.Z

## 빠른 참조 (30초)

한 단락 요약. Agents가 이 부분 먼저 읽습니다.

---

## 구현 가이드 (5분)

구조화된 구현 지침:

### 기능

- 기능 1
- 기능 2

### 사용 시기

- 경우 1에 사용
- 경우 2에 사용

### 핵심 패턴

- 패턴 1
- 패턴 2

---

## 고급 구현 (10분 이상)

깊이 있는 설명, 복잡한 시나리오, 엣지 케이스

---

## 참고 & 예제

완전한 예제, 코드 스니펫, 상세 참조
```

---

### 6. Skill SKILL.md 표준

```markdown
---
name: moai-[domain]-[skill-name]
description: [한 줄 설명 - 15단어 이내]
---

## 빠른 참조 (30초)

한 단락.

---

## 구현 가이드

### 기능

[기능 목록]

### 사용 시기

[사용 케이스]

### 핵심 패턴

[패턴과 예제]

---

## 고급 구현 (Level 3)

[복잡한 패턴, 엣지 케이스]

---

## 참고 & 자료

[완전한 API 참조, 예제, 링크]
```

**Skill 명명 규칙**:

```text
moai-cc-[기능명]           # Claude Code 관련
moai-foundation-[개념]     # 공유 개념
moai-[언어]-[기능]         # 언어별 기능
```

예:

- moai-cc-commands (Claude Code commands)
- moai-foundation-trust (TRUST 5 프레임워크)
- moai-lang-python (Python 특화)

---

### 7. Agents가 CLAUDE.md 읽는 방식

Agents는 다음 순서로 정보를 추출합니다:

1. **나는 무엇을 할 수 있나?** (Permissions 섹션)

   - 도구 허용/차단 목록
   - 최대 토큰 예산
   - 실행 제약사항

2. **나는 언제 자동 실행되나?** (Auto-trigger 섹션)

   - Trigger 조건
   - 이벤트 타입
   - Config 의존성

3. **누구를 호출하나?** (Delegation 섹션)

   - 호출할 Sub-agents
   - 각 호출 시점
   - 전달할 Context

4. **성공은 어떻게 아나?** (Quality gate 섹션)
   - Pass 기준
   - Fail 처리
   - 검증 단계

---

### 8. Config 필드 참조 패턴

CLAUDE.md에서 config를 참조할 때 사용할 형식:

```markdown
### Config: github.spec_git_workflow

**필드 경로**: config.json → github → spec_git_workflow
**타입**: String (enum)
**가능값**: develop_direct, feature_branch, per_spec
**기본값**: develop_direct
**우선순위**: Priority 1 (최상위)

**영향**:

- Git branch 생성 여부 제어
- git-manager auto-trigger 결정
- PHASE 3 실행 여부 결정

**검증 규칙**:

- 반드시 enum 값 중 하나
- 누락 시: 기본값 develop_direct 사용
- 유효하지 않은 값: 경고 후 기본값 사용

**관련 필드**:

- `git_strategy.mode` (fallback)
- `github.spec_git_workflow_configured` (validation flag)
```

---

### 9. 업데이트 & 유지보수

#### 버전 관리

- CLAUDE.md 변경사항을 semantic versioning으로 태그
- root의 CHANGELOG.md에서 주요 변경사항 기록
- 필요시 frontmatter에 버전 명시

#### 검토 프로세스 (병합 전)

1. **명확성 검토**: 큰 소리로 읽어보기 (모호함 확인)
2. **Agent 테스트**: Agents가 규칙을 명확히 추출 가능한가?
3. **Config 검증**: Config 참조가 schema와 일치하나?
4. **참조 확인**: 외부 참조가 실제로 존재하나?
5. **예제 검증**: 예제가 그대로 실행 가능한가?

#### 오래된 내용 아카이빙

오래된 CLAUDE.md 섹션은:

- `.moai/archive/CLAUDE.md.[날짜]`로 이동
- 활성 CLAUDE.md에는 현재 규칙만 유지

---

## 참고 자료

### 공식 문서

- [Claude Code 공식 문서](https://code.claude.com/docs)
- [Claude Code CLI 레퍼런스](https://code.claude.com/docs/en/cli-reference)
- [Claude Code 설정 가이드](https://code.claude.com/docs/en/settings)
- [MCP 통합 가이드](https://code.claude.com/docs/en/mcp)

### MoAI-ADK 문서

- [CLAUDE.md](./CLAUDE.md) - Claude Code 실행 가이드
- [.moai/memory/](./. moai/memory/) - 참고 문서
- [README.md](./README.md) - 프로젝트 개요

### 관련 Skill

- `moai-cc-claude-md` - CLAUDE.md 작성 가이드
- `moai-cc-hooks` - Claude Code Hooks 시스템
- `moai-cc-skills-guide` - Skill 개발 가이드
- `moai-cc-configuration` - 설정 관리 가이드

---

## 업데이트 이력

| 날짜       | 버전  | 변경사항  |
| ---------- | ----- | --------- |
| 2025-11-22 | 1.0.0 | 초기 작성 |
| -          | -     | -         |

---

**작성자**: GOOS님
**프로젝트**: MoAI-ADK
**상태**: ✅ 활성 문서

---

## 마크다운 표준 & 패턴

### CommonMark 호환성 규칙 (필수)

**규칙**: 괄호는 반드시 bold 마커 **밖에** 있어야 CommonMark 렌더링이 정상 작동합니다.

#### 금지된 패턴:
```markdown
❌ **Text(description)**next - 렌더링 실패
❌ **텍스트(설명)**다음 - 렌더링 실패
❌ **文本(description)**next - 렌더링 실패
```

**이유**: CommonMark는 `()`를 구두점 문자로 취급합니다. 괄호가 bold 마커 `**` 안에 있으면 delimiter run 규칙을 위반하여 마크업이 제대로 닫히지 않습니다.

#### 허용된 패턴:
```markdown
✅ **Text**(description) - 권장 (공백 없음)
✅ **Text** (description) - 허용 (공백 있음)
✅ **текст**(описание) - 모든 언어에서 작동
✅ **文本**(描述) - 모든 언어에서 작동
```

### 구현 규칙

**문서 생성 시**:
마크다운 파일을 생성하거나 수정할 때:

```python
import re

# 검증: 나쁜 패턴 거부
def validate_markdown_pattern(content: str) -> bool:
    bad_pattern = r'\*\*[^*]+\([^)]+\)\*\*[^\s*]'
    return not bool(re.search(bad_pattern, content))

# 정규화: 공백 수정
def normalize_bold_parentheses(content: str) -> str:
    # **text** (desc) → **text**(desc)
    pattern = r'\*\*([^*]+)\*\*\s+\(([^)]+)\)'
    return re.sub(pattern, r'**\1**(\2)', content)
```

### 적용 범위

**필수 적용**:
- README.md (모든 언어 버전)
- API 문서
- 기술 가이드
- docs-manager 에이전트 출력
- doc-syncer 문서 동기화
- 모든 마크다운 관련 스킬

**검증 포인트**:
- 문서 생성: 사전 검증
- 문서 동기화: 사후 정규화
- 수동 편집: markdownlint 검증
- CI/CD: 자동화된 마크다운 검증

### 다국어 지원

이 규칙은 **모든 언어**에 적용되어 일관된 CommonMark 렌더링을 보장합니다:

- 영어: `**Feature**(description)`
- 한국어: `**기능**(설명)`
- 일본어: `**機能**(説明)`
- 중국어: `**功能**(描述)`
- 러시아어: `**Функция**(описание)`

**모든 언어는 동일한 패턴을 따릅니다**: bold와 괄호 사이에 공백 없이 `**Text**(details)` 형식으로 작성합니다.
- '/Users/goos/MoAI/MoAI-ADK/.claude/settings.json'  -> 항상 변수가 치환된 값으로 사용해야 한다.