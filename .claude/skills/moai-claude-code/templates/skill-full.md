---
name: {skill-name}
description: {Comprehensive description under 200 characters - must clearly indicate when Claude should autonomously invoke this skill}
model: haiku
allowed-tools:
  - Read
  - Write
  - Grep
  - Glob
  - Bash
---

# {Skill Title}

> {One-sentence compelling summary of value proposition}

---

## 🎯 이 스킬의 목적

{Comprehensive explanation covering:
- Problem statement and context
- How this skill addresses the problem
- Unique value proposition
- Integration with broader workflows}

**문제**: {Detailed problem description with examples}
**해결**: {Comprehensive solution approach}
**효과**: {Measurable benefits and improvements}

---

## 🏗️ MoAI-ADK 통합

### Alfred 자동 선택 조건

Alfred는 다음 조건에서 이 스킬을 자동으로 활성화합니다:

- {Specific automatic trigger condition 1 with context}
- {Specific automatic trigger condition 2 with keywords}
- {Specific automatic trigger condition 3 with workflow state}

### 워크플로우 위치

```
/alfred:1-spec → /alfred:2-build → /alfred:3-sync
                                        ↑
                                  이 스킬 자동 활성화
                                  ({when activated})
```

**통합 시점**:
- **Phase**: {Which phase of MoAI-ADK workflow}
- **트리거**: {What triggers automatic invocation}
- **역할**: {What this skill contributes to the workflow}

---

## 📋 핵심 기능

### 1. {Major Feature 1 Name}

{Detailed multi-paragraph description of this feature}

**구현 방법**:
```{language}
# {Implementation detail 1}
{code-example-1}

# {Implementation detail 2}
{code-example-2}
```

**산출물**:
- **{Output 1}**: {Detailed description with format}
- **{Output 2}**: {Description with validation criteria}
- **{Output 3}**: {Description with usage notes}

**검증**:
```bash
# {Verification method}
{verification-command}
```

---

### 2. {Major Feature 2 Name}

{Comprehensive feature description}

**알고리즘**:
1. {Step 1 of algorithm}
2. {Step 2 with details}
3. {Step 3 and expected outcome}

**구현 예시**:
```{language}
{detailed-code-example}
```

---

### 3. {Major Feature 3 Name}

{Feature description with use cases}

**사용 시나리오**:
- **{Scenario A}**: {When and why to use}
- **{Scenario B}**: {Alternative use case}

---

## 💡 사용 패턴

### 패턴 1: 수동 호출

**사용자 요청 예시**:
```
"{skill-name} 실행해주세요"
"{natural-language-trigger-phrase}"
```

**Alfred 동작**:
1. {What Alfred does in step 1}
2. {What Alfred does in step 2}
3. {Final action and result}

---

### 패턴 2: 자동 활성화

**트리거 조건**: {When automatic activation occurs}

**Alfred 감지 시나리오**:
```
사용자: "{example-user-request}"
→ Alfred 분석: {how Alfred recognizes this needs the skill}
→ 자동 실행: {what happens automatically}
→ 결과: {what user receives}
```

---

### 패턴 3: 커맨드 통합

**연관 커맨드**: `/{command-name}`

**통합 흐름**:
```
/{command-name} 실행
  ↓
{When skill is invoked during command}
  ↓
이 스킬 자동 호출
  ↓
{What skill contributes}
  ↓
커맨드 계속 진행
```

---

## ⚙️ 설정 및 구성

### 설정 파일 위치

`.moai/config.json`에서 설정:

```json
{
  "{skill-config-section}": {
    "{option1}": {default-value},
    "{option2}": {default-value},
    "{option3}": {
      "{sub-option1}": {value},
      "{sub-option2}": {value}
    }
  }
}
```

### 설정 옵션 상세

| 옵션 | 타입 | 기본값 | 필수 | 설명 |
|-----|------|-------|------|------|
| `{option1}` | {type} | `{default}` | ✅/⚠️ | {Comprehensive description} |
| `{option2}` | {type} | `{default}` | ⚠️ | {What this controls} |
| `{option3}` | {type} | `{default}` | ⚠️ | {Usage notes} |

### 환경변수 (선택적)

```bash
# {Environment variable 1}
export {VAR_NAME_1}="{value}"

# {Environment variable 2}
export {VAR_NAME_2}="{value}"
```

---

## 📁 디렉토리 구조

```
.claude/skills/{skill-name}/
├── SKILL.md                # 메인 스킬 정의 (this file)
├── reference.md            # 상세 참조 문서
├── examples.md             # 실전 예제 모음
├── scripts/                # 유틸리티 스크립트
│   ├── {helper-1}.py
│   └── {helper-2}.py
└── templates/              # 템플릿 파일
    ├── {template-1}.txt
    └── {template-2}.json
```

### 추가 파일 설명

- **reference.md**: {What additional documentation it contains}
- **examples.md**: {What examples are provided}
- **scripts/**: {What utility scripts do}
- **templates/**: {What templates are included}

---

## ✅ 검증 체크리스트

### 실행 전 검증

- [ ] {Pre-execution check 1}
- [ ] {Pre-execution check 2}
- [ ] {Pre-execution check 3}

### 실행 후 검증

- [ ] {Post-execution validation 1 with criteria}
- [ ] {Post-execution validation 2 with expected state}
- [ ] {Post-execution validation 3 with deliverable}
- [ ] {MoAI-ADK 워크플로우 통합 확인}

### 검증 명령어

```bash
# {Validation script 1}
uv run .claude/skills/{skill-name}/scripts/validate.py

# {Validation check 2}
{verification-command}

# {Integration test}
{integration-test-command}
```

---

## 🚨 에러 처리

### 에러 분류

#### 1. {Error Category 1}

**증상**: {How this error manifests}

**원인**:
- {Possible cause 1}
- {Possible cause 2}

**해결 방법**:
```bash
# {Solution step 1}
{command-1}

# {Solution step 2}
{command-2}
```

---

#### 2. {Error Category 2}

**증상**: {Error description}

**디버깅**:
```bash
# {How to debug}
{debug-command}
```

**수정**:
1. {Fix step 1}
2. {Fix step 2}

---

### 로깅 및 디버깅

**로그 위치**: `{log-file-path}`

**로그 레벨 설정**:
```bash
# {How to enable debug logging}
{logging-config-command}
```

**로그 확인**:
```bash
# {How to view logs}
tail -f {log-file-path}
```

---

## 🔗 연관 에이전트/커맨드

### 연관 커맨드

- **/{command-1}** - {How this skill supports the command}
- **/{command-2}** - {Integration point}

### 연관 에이전트

- **@agent-{agent-1}** - {How they work together}
- **@agent-{agent-2}** - {Collaboration scenario}

### 연관 스킬

- **{skill-1}** - {Complementary functionality}
- **{skill-2}** - {When to use together}

---

## 📊 성능 및 메트릭

### 성능 특성

- **실행 시간**: {Typical execution time}
- **메모리 사용**: {Expected memory usage}
- **디스크 I/O**: {File operations count}
- **네트워크**: {External API calls if any}

### 최적화 팁

1. **{Optimization 1}**: {How to improve performance}
2. **{Optimization 2}**: {Configuration tweak}
3. **{Optimization 3}**: {Best practice}

---

## 🎓 베스트 프랙티스

### 1. {Practice Category 1}

**권장 사항**:
```{language}
# {Good practice example}
{recommended-code}
```

**피해야 할 사항**:
```{language}
# {Anti-pattern example}
{avoid-this-code}
```

**이유**: {Why this is best practice}

---

### 2. {Practice Category 2}

**팁**: {Helpful tip}

**예시**:
```bash
{example-of-best-practice}
```

---

### 3. {Practice Category 3}

**주의사항**: {Important consideration}

---

## 📖 실전 예제

### 예제 1: {Common Use Case}

**목적**: {What this example demonstrates}

**입력**:
```{format}
{example-input}
```

**실행**:
```bash
{commands-to-run}
```

**출력**:
```{format}
{example-output}
```

**설명**: {What happened and why}

---

### 예제 2: {Advanced Use Case}

**목적**: {Advanced scenario}

**시나리오**: {Detailed scenario description}

**구현**:
```{language}
{implementation-code}
```

**결과**: {What you achieve}

---

### 예제 3: {Edge Case}

**상황**: {Unusual but important scenario}

**처리 방법**: {How skill handles this}

---

## 🔧 커스터마이제이션

### 확장 포인트

이 스킬을 프로젝트에 맞게 커스터마이즈할 수 있는 영역:

1. **{Extension Point 1}**
   - 파일: `{file-to-modify}`
   - 수정 방법: {How to customize}

2. **{Extension Point 2}**
   - 설정: `{config-key}`
   - 옵션: {Available options}

### 플러그인 시스템 (고급)

```python
# {How to create plugins for this skill}
{plugin-example-code}
```

---

## 📚 참고 자료

### 공식 문서
- **Claude Code Skills**: https://docs.claude.com/en/docs/claude-code/skills
- **{Related Doc}**: {URL}

### MoAI-ADK 리소스
- **개발 가이드**: `.moai/memory/development-guide.md`
- **SPEC 메타데이터**: `.moai/memory/spec-metadata.md`

### 커뮤니티
- **GitHub Issues**: {Link}
- **디스커션**: {Link}

---

## 🔄 업데이트 로그

### v1.0.0 (Initial)
- {Feature 1 introduced}
- {Feature 2 implemented}
- {Initial release notes}

---

**Template Level**: Full
**Best For**: Production MoAI-ADK integration, enterprise workflows
**Features**:
- Alfred 자동 선택
- 워크플로우 통합
- 상세 설정
- 검증 자동화
- 에러 처리
- 성능 최적화

**Directory Structure**: Full (SKILL.md + reference.md + examples.md + scripts/ + templates/)
**Estimated Setup Time**: 45-60 minutes
**Maintenance**: Regular updates as workflow evolves
**Support**: Full MoAI-ADK integration support

---

이 스킬은 {domain}에서 최고 수준의 자동화를 제공합니다.
