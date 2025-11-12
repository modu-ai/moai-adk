# 기여 가이드

MoAI-ADK 프로젝트에 기여해주셔서 감사합니다! 이 문서는 프로젝트에 효과적으로 기여하는 방법을 안내합니다.

## 기여 방법

### 1. 버그 리포트

버그를 발견하셨나요? 다음 정보를 포함하여 이슈를 등록해주세요:

- **환경 정보**: OS, Python 버전, MoAI-ADK 버전
- **재현 단계**: 버그를 재현할 수 있는 명확한 단계
- **예상 동작**: 원래 기대했던 동작
- **실제 동작**: 실제로 발생한 동작
- **스크린샷/로그**: 가능하면 관련 스크린샷이나 로그

### 2. 기능 제안

새로운 기능을 제안하고 싶으신가요?

- **Use Case**: 이 기능이 필요한 실제 사용 시나리오
- **제안 내용**: 기능의 상세한 설명
- **대안 고려**: 고려해본 다른 대안들
- **추가 컨텍스트**: 관련 스크린샷, 참고 자료 등

### 3. 코드 기여

코드를 직접 기여하시려면:

#### 개발 환경 설정

```bash
# 저장소 Fork 및 Clone
git clone https://github.com/YOUR_USERNAME/MoAI-ADK.git
cd MoAI-ADK

# 개발 환경 설정
uv venv
source .venv/bin/activate  # Windows: .venv\Scripts\activate

# 개발 의존성 설치
uv pip install -e ".[dev]"
```

#### GitFlow 워크플로우

MoAI-ADK는 GitFlow 전략을 따릅니다:

```bash
# feature 브랜치 생성
git checkout -b feature/SPEC-YOUR-FEATURE develop

# 개발 및 커밋
git add .
git commit -m "feat: Add your feature

🤖 Generated with [Claude Code](https://claude.com/claude-code)


# Push 및 PR 생성
git push origin feature/SPEC-YOUR-FEATURE
```

**중요**: PR은 항상 `develop` 브랜치를 타겟으로 합니다.

#### 코딩 스타일

- **PEP 8**: Python 코딩 스타일 가이드 준수
- **Type Hints**: 모든 함수에 타입 힌트 사용
- **Docstrings**: Google 스타일 docstring 작성
- **Tests**: 새로운 기능에는 테스트 포함 필수

```python
def example_function(param: str) -> dict[str, Any]:
    """
    함수의 간단한 설명.

    Args:
        param: 매개변수 설명

    Returns:
        반환값 설명

    Raises:
        ValueError: 발생 가능한 예외 설명
    """
    pass
```

#### 테스트 작성

TDD (Test-Driven Development) 원칙을 따릅니다:

```bash
# 테스트 실행
pytest tests/

# 커버리지 확인
pytest --cov=moai_adk tests/

# 특정 테스트만 실행
pytest tests/test_specific.py::test_function
```

#### 문서 업데이트

코드 변경 시 관련 문서도 함께 업데이트해주세요:

- **API 문서**: `docs/pages/ko/reference/`
- **가이드**: `docs/pages/ko/guides/`
- **튜토리얼**: `docs/pages/ko/tutorials/`

### 4. 문서 기여

문서 개선도 큰 도움이 됩니다:

- 오타 수정
- 설명 개선
- 예제 추가
- 번역 작업

```bash
# 문서 로컬 빌드
cd docs
npm install
npm run dev
```

## Pull Request 체크리스트

PR을 제출하기 전에 다음을 확인해주세요:

- [ ] 코드가 PEP 8 스타일을 따릅니다
- [ ] 모든 테스트가 통과합니다
- [ ] 새로운 기능에 테스트를 추가했습니다
- [ ] Type hints를 추가했습니다
- [ ] Docstring을 작성했습니다
- [ ] 관련 문서를 업데이트했습니다
- [ ] CHANGELOG.md를 업데이트했습니다
- [ ] `develop` 브랜치를 타겟으로 설정했습니다

## 커밋 메시지 규칙

[Conventional Commits](https://www.conventionalcommits.org/) 형식을 따릅니다:

```
<type>: <subject>

[optional body]

🤖 Generated with [Claude Code](https://claude.com/claude-code)

```

**Types:**
- `feat`: 새로운 기능
- `fix`: 버그 수정
- `docs`: 문서 변경
- `style`: 코드 스타일 (포매팅, 세미콜론 등)
- `refactor`: 코드 리팩토링
- `test`: 테스트 추가/수정
- `chore`: 빌드, 패키지 매니저 등

**예시:**

```
feat: Add MCP server integration support

- Add MCP server configuration during init
- Support interactive and CLI options
- Auto-install recommended MCP servers

🤖 Generated with [Claude Code](https://claude.com/claude-code)

```

## 코드 리뷰 프로세스

1. **자동 검사**: PR 생성 시 자동으로 실행됩니다
   - Linting (Ruff)
   - Type checking (mypy)
   - Tests (pytest)
   - Coverage report

2. **코드 리뷰**: 메인테이너가 코드를 검토합니다
   - 코드 품질
   - 테스트 커버리지
   - 문서 완성도
   - 아키텍처 적합성

3. **피드백**: 필요시 수정 요청을 받습니다

4. **머지**: 모든 검사 통과 및 승인 후 `develop`에 머지됩니다

## 질문이 있으신가요?

- **GitHub Issues**: 버그, 기능 제안, 질문
- **Discussions**: 일반적인 토론, 아이디어 공유
- **Ko-fi**: 프로젝트 후원 및 지원

## 행동 강령

모든 기여자는 다음을 준수해야 합니다:

- **존중**: 모든 참여자를 존중합니다
- **협력**: 건설적인 피드백을 제공합니다
- **포용**: 다양한 관점을 환영합니다
- **전문성**: 전문적인 태도를 유지합니다

## 라이선스

MoAI-ADK에 기여하는 모든 코드는 프로젝트의 라이선스(MIT License)를 따릅니다.

---

**감사합니다!** 여러분의 기여가 MoAI-ADK를 더 나은 프로젝트로 만듭니다. 🎩
