# moai-core-rules Reference ( .0)

## 공식 자료 & 링크 (November 2025 Enterprise Standard)

### Architecture & Design Patterns

| 자료 | 링크 | 설명 |
|-----|------|------|
| **Agentic AI Design Patterns 2025** | https://www.infoq.com/articles/agentic-ai-architecture-framework/ | Enterprise agent architecture patterns |
| **ReAct Framework** | https://arxiv.org/abs/2210.03629 | Reasoning + Acting patterns |
| **Multi-Agent Orchestration** | https://infoq.com/news/2025/10/ai-agent-orchestration/ | Agent coordination patterns |

### TDD & Testing

| 자료 | 링크 | 설명 |
|-----|------|------|
| **pytest Documentation** | https://docs.pytest.org/ | Test framework & best practices |
| **TDD Best Practices** | https://martinfowler.com/bliki/TestDrivenDevelopment.html | Martin Fowler on TDD |
| **Coverage.py** | https://coverage.readthedocs.io/ | Code coverage measurement |

### Code Quality Standards

| 자료 | 링크 | 설명 |
|-----|------|------|
| **SOLID Principles** | https://en.wikipedia.org/wiki/SOLID | Object-oriented design principles |
| **Clean Code** | https://www.oreilly.com/library/view/clean-code-a/9780136083238/ | Robert Martin's clean code guide |
| **OWASP Top 10** | https://owasp.org/www-project-top-ten/ | Security vulnerabilities |

### Git & Version Control

| 자료 | 링크 | 설명 |
|-----|------|------|
| **Conventional Commits** | https://www.conventionalcommits.org/ | Commit message standard |
| **Git Flow** | https://nvie.com/posts/a-successful-git-branching-model/ | Branching model |
| **Semantic Versioning** | https://semver.org/ | Version numbering standard |

---

## 아키텍처 규칙 매트릭스

### Command vs Agent vs Skill

| 계층 | 책임 | 금지 | 필수 |
|-----|------|------|------|
| **Command** | 조율만 | bash 직접 실행, Skill 호출 | Task() 위임 |
| **Agent** | 도메인 전문, 계획/실행 | 상태 유지, 다른 Agent 직접 호출 | Skill 호출, 분석/결정 |
| **Skill** | 재사용 가능 playbook | Task() 실행, 다른 Skill 호출 | Read, Glob, Grep |

---

## 10 Mandatory Skills 참조 매트릭스

| # | Skill | 호출 조건 | 반환값 |
|---|-------|---------|--------|
| 1 | moai-foundation-trust | TRUST 5 검증 필요 | Score + violations |
| 3 | moai-foundation-specs | SPEC 작성/검증 | Template + validation |
| 4 | moai-foundation-ears | EARS 형식 필요 | Format guide |
| 5 | moai-foundation-git | Git 워크플로우 | Branch + commit rules |
| 6 | moai-foundation-langs | 언어 감지 | Language + framework |
| 7 | moai-essentials-debug | 디버깅 필요 | Error analysis |
| 8 | moai-essentials-refactor | 코드 개선 | Refactoring suggestions |
| 9 | moai-essentials-perf | 성능 최적화 | Optimization tips |
| 10 | moai-essentials-review | 코드 리뷰 | Quality report |

---

## AskUserQuestion 패턴 체크리스트

### 사용 여부 결정 트리

```
User Intent Clear?
├─ YES → 계속 진행
└─ NO → AskUserQuestion 사용
    ├─ Tech stack unclear? → 옵션 제시
    ├─ Architecture decision? → 트레이드오프 설명
    ├─ Intent ambiguous? → 명확화 질문
    ├─ Component impact? → 영향도 확인
    └─ Resource constraints? → 제약 조건 확인
```

### 올바른 형식

```python
AskUserQuestion({
  question: "명확한 질문 (한글)",
  header: "카테고리",
  multiSelect: false,
  options: [
    {
      label: "옵션 1",
      description: "설명"
    },
    {
      label: "옵션 2",
      description: "설명"
    }
  ]
})
```

---

## TRUST 5 Gate Validation Checklist

### Test Coverage

```bash
# Coverage 측정
pytest --cov=src tests/

# 목표: ≥ 85%
# 확인: 모든 브랜치 커버
```

### Readable (가독성)

```bash
# 형식 검사
black --check src/
flake8 src/
pylint src/

# 타입 검사
mypy src/
```

### Unified (일관성)

```bash
# Import 순서
isort --check src/

# Linting
pylint src/

# 중복 확인
radon cc src/  # Cyclomatic complexity
```

### Secured (보안)

```bash
# 보안 스캔
bandit -r src/

# 의존성 취약점
safety check

# 비밀 검사
detect-secrets scan
```


```bash
grep -r "@[A-Z]+-[0-9]\+" .

# 체인 검증
# SPEC, TEST, CODE, COMMIT 모두 포함?
```

---

## 커밋 메시지 형식 예제

### RED (테스트 작성)

```

- Implement test_successful_login()
- Implement test_invalid_credentials()
- Implement test_expired_token()

All tests RED (expected to fail at this stage).
```

### GREEN (구현)

```

- Add authenticate_user() function
- Add JWT token generation
- Add credential validation
- Add error handling

All tests now GREEN (passing).
```

### REFACTOR (최적화)

```

- Extract token validation to separate function
- Add in-memory cache for user lookups
- Improve error messages for better UX
- Add type hints for clarity

All tests PASSING (quality improved, tests still pass).
```

---

## Agent Delegation 패턴

### 정책 결정 → Delegation

```python
# 패턴 1: 계획 필요
if planning_required:
    Task(
        subagent_type="plan-agent",
        description="Break down this feature",
        prompt="Create detailed plan with risks, timeline, dependencies"
    )

# 패턴 2: 코드 개발
if code_needed:
    Task(
        subagent_type="tdd-implementer",
        description="Implement feature X",
    )

# 패턴 3: 테스트
if testing_needed:
    Task(
        subagent_type="test-engineer",
        description="Write tests for feature X",
        prompt="Achieve 85%+ coverage, test edge cases"
    )

# 패턴 4: 문서화
if docs_needed:
    Task(
        subagent_type="doc-syncer",
        description="Document feature X",
    )

# 패턴 5: Git 커밋
if commit_needed:
    Task(
        subagent_type="git-manager",
        description="Commit changes",
    )
```

---

## TAG 무결성 검증 스크립트

```python
# tag_validator.py

import re
from pathlib import Path

TAG_PATTERN = r'@[A-Z]+-\d{3}'

def find_all_tags(directory):
    tags = {}
    for file in Path(directory).rglob('*'):
        if file.is_file() and file.suffix in ['.py', '.md', '.txt']:
            content = file.read_text()
            matches = re.findall(TAG_PATTERN, content)
            for tag in matches:
                if tag not in tags:
                    tags[tag] = []
                tags[tag].append(str(file))
    return tags

def validate_tag_chain(tags):
    """Validate complete TAG chains"""
    issues = []
    
    for tag, files in tags.items():
        has_spec = any('spec' in f.lower() for f in files)
        has_test = any('test' in f.lower() for f in files)
        has_code = any('src' in f for f in files)
        has_commit = any('.git' in f for f in files)
        
        if not has_spec:
            issues.append(f"{tag}: Missing SPEC definition")
        if not has_test:
            issues.append(f"{tag}: Missing TEST")
        if not has_code:
            issues.append(f"{tag}: Missing CODE")
    
    return issues

if __name__ == "__main__":
    tags = find_all_tags(".")
    issues = validate_tag_chain(tags)
    
    if issues:
        print("TAG Chain Issues Found:")
        for issue in issues:
            print(f"  - {issue}")
    else:
        print("All TAG chains valid!")
```

---

**최종 업데이트**: 2025-11-12
**버전**: 4.0.0
**유지보수자**: GoosLab
