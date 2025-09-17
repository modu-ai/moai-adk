# MoAI-ADK 대화형 마법사

## 🧙‍♂️ 대화형 마법사 개요

`/moai:1-project` 명령어는 10단계 Q&A 시스템을 통해 프로젝트를 완전히 설정합니다.

## 10단계 맥락 자산 구축 시스템

### 1. 프로젝트 기본 정보
```
Q: 프로젝트 이름을 입력하세요
A: [사용자 입력]

Q: 프로젝트 설명을 간단히 입력하세요
A: [사용자 입력]
```

### 2. 기술 스택 선택
```
Q: 주 프로그래밍 언어를 선택하세요
   1) TypeScript/JavaScript
   2) Python
   3) Rust
   4) Go
   5) 기타
A: [1-5 선택]
```

### 3. 프레임워크 선택
```
Q: 사용할 프레임워크를 선택하세요
   [언어별 동적 옵션 표시]
   TypeScript: Next.js, React, Vue.js, Express
   Python: Django, FastAPI, Flask
A: [동적 선택]
```

### 4. 프로젝트 유형
```
Q: 프로젝트 유형을 선택하세요
   1) 웹 애플리케이션
   2) API 서버
   3) 데스크톱 앱
   4) 모바일 앱
   5) 라이브러리/패키지
A: [1-5 선택]
```

### 5. 데이터베이스 선택
```
Q: 데이터베이스를 선택하세요
   1) PostgreSQL
   2) MySQL
   3) MongoDB
   4) SQLite
   5) 사용 안함
A: [1-5 선택]
```

### 6. 배포 환경
```
Q: 배포 환경을 선택하세요
   1) AWS
   2) Google Cloud
   3) Azure
   4) Vercel
   5) 온프레미스
A: [1-5 선택]
```

### 7. 팀 규모
```
Q: 팀 규모를 선택하세요
   1) 개인 프로젝트
   2) 소규모 팀 (2-5명)
   3) 중규모 팀 (6-15명)
   4) 대규모 팀 (16명+)
A: [1-4 선택]
```

### 8. 개발 방법론
```
Q: 개발 방법론을 선택하세요
   1) Agile/Scrum
   2) Kanban
   3) Waterfall
   4) 자유 형식
A: [1-4 선택]
```

### 9. 품질 요구사항
```
Q: 품질 요구사항을 선택하세요 (복수 선택)
   1) 높은 테스트 커버리지 (90%+)
   2) 엄격한 코드 품질 (linting)
   3) 자동화된 보안 검사
   4) 성능 모니터링
A: [1-4 복수 선택]
```

### 10. 특별 요구사항
```
Q: 특별한 요구사항이 있다면 입력하세요
A: [자유 입력]
```

## 동적 분기 로직

### 언어별 분기
```python
if language == "typescript":
    frameworks = ["Next.js", "React", "Vue.js", "Express", "Nest.js"]
elif language == "python":
    frameworks = ["Django", "FastAPI", "Flask", "Streamlit"]
elif language == "rust":
    frameworks = ["Actix Web", "Rocket", "Warp", "Axum"]
```

### 프로젝트 유형별 분기
```python
if project_type == "web_app":
    questions.extend(frontend_questions)
elif project_type == "api_server":
    questions.extend(api_questions)
elif project_type == "library":
    questions.extend(library_questions)
```

## 자동 스캔 및 프리필

### 기존 코드 스캔
```python
def scan_existing_project():
    """기존 프로젝트 파일 스캔"""
    scanned_info = {}

    # package.json 스캔
    if Path("package.json").exists():
        scanned_info["language"] = "typescript"
        scanned_info["dependencies"] = scan_package_json()

    # pyproject.toml 스캔
    if Path("pyproject.toml").exists():
        scanned_info["language"] = "python"
        scanned_info["dependencies"] = scan_pyproject_toml()

    # Git 정보 스캔
    if Path(".git").exists():
        scanned_info["git_remote"] = get_git_remote()

    return scanned_info
```

### 자동 프리필
```python
def prefill_answers(scanned_info):
    """스캔 결과로 답변 사전 입력"""
    prefilled = {}

    if "language" in scanned_info:
        prefilled["language"] = scanned_info["language"]

    if "next" in scanned_info.get("dependencies", []):
        prefilled["framework"] = "Next.js"

    return prefilled
```

## 생성 결과

### Steering 문서 생성
- **product.md**: 제품 비전과 전략
- **structure.md**: 아키텍처 설계
- **tech.md**: 기술 스택 선정

### 프로젝트 구조 생성
```
프로젝트/
├── docs/           # 문서 디렉토리
├── src/            # 소스 코드
├── tests/          # 테스트 디렉토리
├── .env.example    # 환경 변수 템플릿
└── README.md       # 프로젝트 소개
```

### 설정 파일 생성
- 언어별 설정 파일 (tsconfig.json, pyproject.toml 등)
- CI/CD 파이프라인 (.github/workflows/)
- 개발 도구 설정 (.eslintrc, .prettierrc 등)

## 레거시 Steering 파일명 마이그레이션

이전 버전에서 생성된 `.moai/steering/vision.md`, `architecture.md`, `techstack.md`는 더 이상 감지되지 않습니다.
표준 파일명(`product.md`, `structure.md`, `tech.md`)으로 마이그레이션하세요.

```bash
# 드라이런(계획만 출력)
python scripts/migrate_steering_filenames.py

# 실제 적용(기존 타겟이 있을 경우 백업 후 덮어쓰기)
python scripts/migrate_steering_filenames.py --apply --force
```

## 마법사 재실행

### 설정 변경/점진적 업데이트
설정/수정/추가는 이제 별도 서브커맨드 없이 `/moai:1-project`에서 대화형으로 처리합니다.

대화형 마법사는 **맞춤형 프로젝트 설정**과 **자동화된 구조 생성**을 통해 완벽한 개발 환경을 제공합니다.
