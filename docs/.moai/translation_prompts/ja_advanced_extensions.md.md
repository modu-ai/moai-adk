Translate the following Korean markdown document to Japanese.

**CRITICAL RULES:**
1. Preserve ALL markdown structure (headers, code blocks, links, tables, diagrams)
2. Keep ALL code blocks and technical terms UNCHANGED
3. Maintain the EXACT same file structure and formatting
4. Translate ONLY Korean text content
5. Keep ALL @TAG references unchanged (e.g., @SPEC:AUTH-001)
6. Preserve ALL file paths and URLs
7. Keep ALL emoji and icons as-is
8. Maintain ALL frontmatter (YAML) structure

**Source File:** /Users/goos/MoAI/MoAI-ADK/docs/src/ko/advanced/extensions.md
**Target Language:** Japanese
**Target File:** /Users/goos/MoAI/MoAI-ADK/docs/src/ja/advanced/extensions.md

**Content to Translate:**

# 확장 및 커스터마이제이션 가이드

MoAI-ADK를 프로젝트에 맞게 커스터마이즈하는 방법입니다.

## :bullseye: 확장 가능한 영역

1. **Custom Skills**: 새로운 도메인 스킬 추가
2. **Custom Agents**: 특화된 에이전트 생성
3. **Custom Hooks**: 프로젝트별 자동화 Hook
4. **Custom Commands**: Alfred 커맨드 확장

## :hammer_and_wrench: Custom Skills 만들기

### Skill 구조

```
.claude/skills/
└── custom-skill/
    ├── index.md           # Skill 문서
    ├── examples.md        # 사용 예시
    ├── reference.md       # API 명세
    └── templates/         # 프롬프트 템플릿
```

### Skill 작성 예시

```markdown
# moai-custom-mlops

기계학습 파이프라인 구축 및 배포 스킬

## 용도
- ML 모델 학습
- 모델 평가 및 검증
- 프로덕션 배포

## 포함 내용
- MLflow 통합
- Kubeflow 배포
- 모델 서빙 패턴
```

### Skill 등록

```bash
# Skill 메타데이터 추가
# .moai/config.json:
{
  "custom_skills": {
    "moai-custom-mlops": {
      "version": "1.0",
      "author": "team",
      "enabled": true
    }
  }
}
```

## 👥 Custom Agents 만들기

### Agent 구조

```
.claude/agents/
└── custom-agent/
    ├── agent.py          # Agent 구현
    ├── prompts.md        # 프롬프트
    └── tools.json        # 도구 목록
```

### Agent 예시

```python
# .claude/agents/ml-expert/agent.py

class MLExpert:
    """머신러닝 전문가 에이전트"""

    def __init__(self):
        self.skills = [
            "moai-domain-ml",
            "moai-lang-python"
        ]

    def analyze_data(self, dataset):
        """데이터 분석"""
        # 분석 로직
        pass

    def train_model(self, data, params):
        """모델 학습"""
        # 학습 로직
        pass

    def evaluate(self, model, test_data):
        """모델 평가"""
        # 평가 로직
        pass
```

### Agent 활성화

```bash
# 커스텀 에이전트 추가
# .moai/config.json:
{
  "custom_agents": {
    "ml-expert": {
      "enabled": true,
      "activation_keywords": ["machine learning", "mlops", "model"]
    }
  }
}
```

## 🔧 Custom Hooks 만들기

### Hook 구조

```bash
.claude/hooks/
├── custom_pre_tool.sh        # Pre-tool Hook
└── custom_post_tool.sh       # Post-tool Hook
```

### Hook 예시

```bash
#!/bin/bash
# .claude/hooks/custom_post_tool.sh

# 모든 Python 파일 생성 후 자동 포매팅
if [[ "$TOOL_NAME" == "Write" && "$FILE_PATH" == *.py ]]; then
    black "$FILE_PATH"
    ruff check "$FILE_PATH"
fi

# Git 커밋 후 자동 푸시
if [[ "$TOOL_NAME" == "Bash" && "$COMMAND" == *"git commit"* ]]; then
    echo "Auto-pushing after commit..."
    git push
fi
```

### Hook 등록

```json
{
  "hooks": {
    "custom_post_tool": ".claude/hooks/custom_post_tool.sh"
  }
}
```

## 📝 Custom Commands 만들기

### Command 파일 구조

```
.claude/commands/
└── custom-deploy.md
```

### Command 파일 예시

```markdown
# /custom-deploy

배포 자동화 커맨드

이 커맨드는 다음을 수행합니다:
1. 빌드 실행
2. 테스트 실행
3. 프로덕션 배포

## 사용법

/custom-deploy [environment] [version]

### 예시

/custom-deploy production v1.0.0
```

### Command 실행

```
사용자: /custom-deploy production

Alfred:
1. 빌드 실행
2. 테스트 검증
3. 배포 확인
4. 모니터링 설정

완료!
```

## 🔄 Integration Points

### Alfred와의 통합

```python
# Alfred가 custom agent 활성화
if "mlops" in spec.keywords:
    activate(ml_expert)  # Custom agent
    activate(backend_expert)  # Built-in agent
```

### Skills와의 통합

```python
# Custom skill 로드
Skill("moai-custom-mlops")
Skill("moai-domain-backend")
```

## 📈 Extension 예시: CI/CD 자동화

### 목표

개발부터 배포까지 자동화된 파이프라인

### 구현

```bash
# Custom hook: .claude/hooks/custom_post_tool.sh

# Git commit 후 자동 CI/CD
if [[ "$COMMAND" == *"git commit"* ]]; then
    echo "🚀 CI/CD 파이프라인 시작..."

    # 1. 빌드
    docker build -t app:latest .

    # 2. 테스트
    docker run app:latest pytest

    # 3. 배포
    kubectl apply -f k8s/

    echo "✅ 배포 완료"
fi
```

## :bullseye: Best Practices

### Custom Skill 만들 때

```
✅ DO:
- 명확한 문서 작성
- 다양한 예시 제공
- 다른 Skill과 조합 가능하게 설계
- 버전 관리 (semantic versioning)

:x: DON'T:
- Alfred의 핵심 로직 수정
- 보안 취약점 무시
- 하드코딩된 경로/값
- 테스트 없는 코드
```

### Custom Agent 만들 때

```
✅ DO:
- 특정 도메인에 집중
- 기존 에이전트와 협력 설계
- 명확한 활성화 조건 정의
- 에러 처리 포함

:x: DON'T:
- 너무 많은 책임 담기
- Alfred와 중복되는 기능
- 순환 참조 생성
- 하드코딩된 설정
```

______________________________________________________________________

**다음**: [보안 고급 가이드](security.md) 또는 [성능 최적화](performance.md)


**Instructions:**
- Translate the content above to Japanese
- Output ONLY the translated markdown content
- Do NOT include any explanations or comments
- Maintain EXACT markdown formatting
- Preserve ALL code blocks exactly as-is
