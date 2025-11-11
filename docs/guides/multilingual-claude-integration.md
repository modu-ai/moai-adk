# 다국어 Claude CLI 통합 가이드

MoAI-ADK의 강화된 다국어 지원 기능과 Claude CLI 통합을 사용하는 방법에 대한 가이드입니다.

## 개요

MoAI-ADK v0.23.0+는 다음과 같은 고급 다국어 기능을 제공합니다:

- **확장된 언어 지원**: 12개 언어 (영어, 한국어, 일본어, 스페인어, 프랑스어, 독일어, 중국어, 포르투갈어, 러시아어, 이탈리아어, 아랍어, 힌디어)
- **자동 번역**: Claude CLI를 사용한 설명 자동 번역
- **변수 치환**: 템플릿 변수를 사용한 동적 콘텐츠 생성
- **Headless 실행**: 비대화형 자동화 지원

## 지원 언어

| 코드 | 영어명 | 원어민 표기 | 어족 |
|------|--------|-------------|------|
| en | English | English | Indo-European |
| ko | Korean | 한국어 | Koreanic |
| ja | Japanese | 日本語 | Japonic |
| es | Spanish | Español | Indo-European |
| fr | French | Français | Indo-European |
| de | German | Deutsch | Indo-European |
| zh | Chinese | 中文 | Sino-Tibetan |
| pt | Portuguese | Português | Indo-European |
| ru | Russian | Русский | Indo-European |
| it | Italian | Italiano | Indo-European |
| ar | Arabic | العربية | Afro-Asiatic |
| hi | Hindi | हिन्दी | Indo-European |

## 기본 사용법

### 1. 언어 목록 확인

```bash
# 지원하는 모든 언어 목록
moai-adk language list

# JSON 형식으로 출력
moai-adk language list --json-output
```

### 2. 특정 언어 정보 확인

```bash
# 한국어 정보
moai-adk language info ko

# 상세 정보 포함
moai-adk language info es --detail
```

### 3. 템플릿 렌더링

```bash
# 변수 파일로 템플릿 처리
moai-adk language render-template template.md variables.json

# 언어 지정
moai-adk language render-template template.md variables.json --language ko

# 출력 파일 지정
moai-adk language render-template template.md variables.json -o output.md
```

## 변수 파일 예제

### variables.json
```json
{
  "PROJECT_NAME": "MyApp",
  "CONVERSATION_LANGUAGE": "ko",
  "CONVERSATION_LANGUAGE_NAME": "한국어",
  "CODEBASE_LANGUAGE": "python",
  "PROJECT_DESCRIPTION": "React 기반 웹 애플리케이션"
}
```

### template.md
```markdown
# {{PROJECT_NAME}}

**언어**: {{CONVERSATION_LANGUAGE_NAME}}
**설명**: {{PROJECT_DESCRIPTION}}
**프로그래밍 언어**: {{CODEBASE_LANGUAGE}}

이 프로젝트는 {{CONVERSATION_LANGUAGE_NAME}}로 개발됩니다.
```

## 다국어 설명 생성

### 자동 번역

```bash
# 기본 설명을 여러 언어로 번역
moai-adk language translate-descriptions "Create React component for user authentication" \
  --target-languages "ko,ja,es,fr,de" \
  -o descriptions.json
```

### 결과 (descriptions.json)
```json
{
  "base": {
    "en": "Create React component for user authentication",
    "ko": "사용자 인증을 위한 React 컴포넌트 생성",
    "ja": "ユーザー認証のためのReactコンポーネント作成",
    "es": "Crear componente de React para autenticación de usuario",
    "fr": "Créer un composant React pour l'authentification utilisateur",
    "de": "React-Komponente für Benutzerauthentifizierung erstellen"
  }
}
```

## Claude CLI 통합

### Headless 모드 실행

```bash
# 변수를 사용한 Claude CLI 실행
moai-adk language execute \
  "Create {{PROJECT_NAME}} project in {{CONVERSATION_LANGUAGE_NAME}}" \
  --variables variables.json \
  --language ko \
  --output-format json
```

### Dry Run 모드

```bash
# 실행하지 않고 명령어만 확인
moai-adk language execute \
  "Generate {{CODEBASE_LANGUAGE}} code for {{PROJECT_NAME}}" \
  --variables variables.json \
  --dry-run
```

## 프로그래밍 방식 사용

### Python API 예제

```python
from pathlib import Path
from src.moai_adk.core.claude_integration import ClaudeCLIIntegration
from src.moai_adk.core.language_config import get_native_name

# Claude 통합 초기화
claude_integration = ClaudeCLIIntegration()

# 변수 정의
variables = {
    "PROJECT_NAME": "MyKoreanApp",
    "CONVERSATION_LANGUAGE": "ko",
    "CONVERSATION_LANGUAGE_NAME": get_native_name("ko"),
    "CODEBASE_LANGUAGE": "python"
}

# 템플릿 처리
template = "Create {{PROJECT_NAME}} using {{CODEBASE_LANGUAGE}} with {{CONVERSATION_LANGUAGE_NAME}} documentation"
result = claude_integration.process_template_command(
    template,
    variables,
    print_mode=True,
    output_format="json"
)

if result["success"]:
    print("Command executed successfully")
    print(f"Output: {result['stdout']}")
else:
    print(f"Error: {result.get('error', 'Unknown error')}")
```

### 다국어 에이전트 생성

```python
# 다국어 지원 에이전트 생성
agent_config = claude_integration.create_agent_with_multilingual_support(
    agent_name="korean-code-reviewer",
    base_description="Code reviewer specialized in Korean projects",
    tools=["Read", "Write", "Bash(git:*)"],
    model="sonnet",
    target_languages=["en", "ko", "ja"]
)

print(f"Agent created with {len(agent_config['descriptions'])} language descriptions")
```

### 다국어 명령어 생성

```python
# 다국어 지원 명령어 생성
command_config = claude_integration.create_command_with_multilingual_support(
    command_name="generate-korean-docs",
    base_description="Generate project documentation in Korean",
    argument_hint=["project_path", "output_format"],
    tools=["Read", "Write", "Edit"],
    model="haiku",
    target_languages=["en", "ko", "ja", "es"]
)

print(f"Command created with descriptions in: {list(command_config['descriptions'].keys())}")
```

## JSON 스트림 처리

### 스트림 입력 처리

```python
import json

# JSON 스트림 입력 처리
input_data = {
    "messages": [
        {"role": "user", "content": "Create {{PROJECT_NAME}} in {{CONVERSATION_LANGUAGE_NAME}}"}
    ],
    "settings": {
        "model": "claude-sonnet-4-5-20250929",
        "language": "{{CONVERSATION_LANGUAGE}}"
    }
}

variables = {
    "PROJECT_NAME": "한국어프로젝트",
    "CONVERSATION_LANGUAGE": "ko",
    "CONVERSATION_LANGUAGE_NAME": "한국어"
}

processed_input = claude_integration.process_json_stream_input(input_data, variables)
print(json.dumps(processed_input, ensure_ascii=False, indent=2))
```

### 스트리밍 실행

```python
# 실시간 스트리밍 실행
result = claude_integration.execute_headless_command(
    prompt_template="Analyze {{PROJECT_NAME}} codebase",
    variables=variables,
    input_format="stream-json",
    output_format="stream-json"
)

if result["success"]:
    print("Streaming output:")
    for line in result["stdout"]:
        print(f"> {line}")
```

## 설정 유효성 검사

```bash
# MoAI-ADK 설정 파일의 언어 구성 검사
moai-adk language validate-config .moai/config.json

# 언어 코드 유효성 검사 포함
moai-adk language validate-config .moai/config.json --validate-languages
```

## 고급 기능

### RTL 언어 지원

```python
from src.moai_adk.core.language_config import is_rtl_language

# RTL 언어 확인
if is_rtl_language("ar"):
    print("아랍어는 RTL(오른쪽에서 왼쪽) 언어입니다")
```

### 언어 패밀리 분석

```python
from src.moai_adk.core.language_config import get_language_family

family = get_language_family("ko")  # "koreanic"
print(f"한국어는 {family} 어족에 속합니다")
```

### 최적 모델 선택

```python
from src.moai_adk.core.language_config import get_optimal_model

optimal_model = get_optimal_model("ko")
print(f"한국어에 최적화된 Claude 모델: {optimal_model}")
```

## 템플릿 변수 참조

MoAI-ADK에서 사용할 수 있는 표준 템플릿 변수:

| 변수 | 설명 | 예시 |
|------|------|------|
| `PROJECT_NAME` | 프로젝트 이름 | `MyApp` |
| `PROJECT_DESCRIPTION` | 프로젝트 설명 | `React 웹 애플리케이션` |
| `PROJECT_OWNER` | 프로젝트 소유자 | `개발팀` |
| `CONVERSATION_LANGUAGE` | 대화 언어 코드 | `ko`, `en`, `ja` |
| `CONVERSATION_LANGUAGE_NAME` | 대화 언어 원어민 표기 | `한국어`, `English`, `日本語` |
| `CODEBASE_LANGUAGE` | 코드베이스 언어 | `python`, `typescript`, `java` |
| `PROJECT_MODE` | 프로젝트 모드 | `personal`, `team` |

## 문제 해결

### 일반적인 오류

1. **언어 코드를 찾을 수 없음**
   ```bash
   # 지원하는 언어 코드 확인
   moai-adk language list
   ```

2. **템플릿 변수가 치환되지 않음**
   ```bash
   # 변수 파일 확인
   cat variables.json | jq .

   # Dry run으로 처리 결과 확인
   moai-adk language execute "Template: {{VAR}}" --variables variables.json --dry-run
   ```

3. **Claude CLI 실행 실패**
   ```bash
   # Claude CLI 설치 확인
   claude --version

   # 권한 확인
   ls -la $(which claude)
   ```

### 디버깅 팁

1. **상세 로그 활성화**
   ```bash
   # Claude CLI 디버그 모드
   claude --debug api,hooks --print "your command"
   ```

2. **JSON 출력 확인**
   ```bash
   # JSON 형식으로 출력하여 구조 확인
   moai-adk language execute "command" --output-format json
   ```

3. **템플릿 변수 추적**
   ```python
   # 사용된 변수 확인
   result = claude_integration.process_template_command(template, variables)
   print(f"Variables used: {result['variables_used']}")
   print(f"Processed command: {result['processed_command']}")
   ```

## 모범 사례

1. **변수 파일 관리**
   - 프로젝트별로 별도의 변수 파일 유지
   - 민감 정보는 환경 변수로 관리
   - 버전 관리에 포함시키기

2. **언어 일관성**
   - 프로젝트 전체에서 일관된 언어 코드 사용
   - 설정 파일과 변수 파일 간的一致性 유지

3. **템플릿 설계**
   - 재사용 가능한 템플릿 구조 설계
   - 필요한 변수만 명시적으로 정의
   - 기본값 제공

4. **오류 처리**
   - 항상 실행 결과 확인
   - 롤백 전략 준비
   - 상세한 로깅

이 가이드를 통해 MoAI-ADK의 다국어 기능을 최대한 활용하여 국제화된 프로젝트를 효율적으로 관리할 수 있습니다.