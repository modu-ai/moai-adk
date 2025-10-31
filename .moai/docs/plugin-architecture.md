# MoAI-ADK v1.0 Plugin Architecture Guide

## 📋 개요

이 문서는 MoAI-ADK v1.0의 5개 공식 플러그인의 아키텍처, 설계 패턴, 및 확장 가이드를 제공합니다.

**대상 독자**:
- 플러그인 사용자 (새로운 기능 이해)
- 개발자 (플러그인 커스터마이징/확장)
- 팀 리더 (플러그인 운영)

---

## 🎯 5개 플러그인 개요

### 플러그인별 역할

```
┌─────────────────────────────────────────────────────────────┐
│                    MoAI-ADK v1.0                            │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  PM Plugin         → SPEC 문서 자동 생성 (EARS 패턴)        │
│  ↓                                                          │
│  UI/UX Plugin      → 프론트엔드 컴포넌트 설정 (shadcn/ui)    │
│  Backend Plugin    → API 프로젝트 초기화 (FastAPI)          │
│  Frontend Plugin   → React 프로젝트 초기화 (상태관리)       │
│  DevOps Plugin     → 배포 파이프라인 설정 (Docker/K8s)      │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### 작업 흐름 (Workflow)

```
1. PM Plugin 실행
   ↓ (SPEC 생성)
2. UI/UX + Frontend Plugin 실행
   ↓ (UI 프로토타입)
3. Backend Plugin 실행
   ↓ (API 구현)
4. DevOps Plugin 실행
   ↓ (배포 설정)
5. 완성!
```

---

## 🏗️ 플러그인 표준 아키텍처

모든 플러그인은 다음 구조를 따릅니다:

```
moai-alfred-{name}/
├── {name}_plugin/
│   ├── __init__.py              # 모듈 초기화 및 export
│   └── commands.py              # 커맨드 클래스 및 로직
├── tests/
│   ├── __init__.py
│   └── test_commands.py         # 단위 및 통합 테스트
├── README.md                    # 사용 설명서
├── USAGE.md                     # 실제 사용 예시 (선택)
└── CHANGELOG.md                 # 버전 변경 사항
```

### 핵심 클래스 구조

모든 플러그인은 다음을 포함합니다:

#### 1. CommandResult (데이터 클래스)

```python
@dataclass
class CommandResult:
    """커맨드 실행 결과"""
    success: bool                       # 성공/실패
    project_dir: Optional[Path]         # 생성된 디렉토리 (또는 config_dir)
    files_created: List[str]            # 생성된 파일 목록
    message: str                        # 사용자 메시지
    error: Optional[str] = None         # 에러 메시지
```

#### 2. Command 클래스

```python
class SetupCommand:
    """플러그인 커맨드 구현"""

    # 검증 메서드 (validation)
    def validate_input(self, value: str) -> bool:
        """입력값 검증"""
        if not value:
            raise ValueError("Invalid input")
        return True

    # 생성 메서드 (creation)
    def create_output_structure(self, path: Path) -> None:
        """출력 디렉토리 및 파일 생성"""
        pass

    # 메인 실행 메서드
    def execute(self, **kwargs) -> CommandResult:
        """메인 오케스트레이션"""
        try:
            # 검증 → 생성 → 반환
            self.validate_input(kwargs['arg1'])
            self.create_output_structure(...)
            return CommandResult(success=True, ...)
        except Exception as e:
            return CommandResult(success=False, error=str(e))
```

---

## 📊 플러그인별 상세 설계

### 1️⃣ PM Plugin 아키텍처

**목적**: SPEC 문서 자동 생성

**클래스 구조**:
- `InitPMCommand` (1개)

**주요 메서드**:
1. `validate_project_name()` - 프로젝트명 검증
2. `validate_template()` - 템플릿 검증
3. `validate_risk_level()` - 위험도 검증
4. `generate_spec_id()` - SPEC ID 자동 생성
5. `create_spec_directory()` - 디렉토리 생성
6. `create_spec_file()` - SPEC 파일 생성
7. `create_plan_file()` - 계획 파일 생성
8. `create_acceptance_file()` - 인수 기준 생성
9. `create_charter_file()` - 헌장 생성
10. `create_risk_matrix()` - 위험 매트릭스 생성
11. `execute()` - 메인 오케스트레이션

**생성 파일**:
```
.moai/specs/SPEC-{ID}/
├── spec.md           # SPEC 문서 (YAML + EARS)
├── plan.md           # 5단계 구현 계획
├── acceptance.md     # 인수 기준 및 메트릭
├── charter.md        # 프로젝트 헌장 (선택)
└── risk-matrix.json  # 위험 분석 (선택)
```

**확장 포인트**:
- 새로운 템플릿 추가 (VALID_TEMPLATES에 추가)
- 추가 YAML 섹션 (create_charter_file 수정)
- 위험 매트릭스 레이아웃 변경

---

### 2️⃣ UI/UX Plugin 아키텍처

**목적**: shadcn/ui 컴포넌트 라이브러리 설정

**클래스 구조**:
- `SetupShadcnUICommand` (1개)

**주요 메서드**:
1. `validate_project_name()` - 프로젝트명 검증
2. `validate_framework()` - 프레임워크 검증
3. `validate_components()` - 컴포넌트 리스트 검증
4. `create_config_directory()` - 설정 디렉토리 생성
5. `create_components_json()` - components.json 생성
6. `create_tailwind_config()` - tailwind.config.js 생성
7. `create_component_structure()` - 컴포넌트 디렉토리 생성
8. `execute()` - 메인 오케스트레이션

**생성 파일**:
```
{project}/
├── components.json           # shadcn/ui 설정
├── tailwind.config.js        # Tailwind CSS 설정
└── src/components/ui/        # 19개 컴포넌트 디렉토리
    ├── button.jsx
    ├── card.jsx
    ├── modal.jsx
    └── ...
```

**19개 지원 컴포넌트**:
- 입력: button, input, textarea, select, checkbox, switch
- 디스플레이: card, badge, label, table
- 컨테이너: modal, dialog, dropdown, accordion
- 고급: form, pagination, tabs, calendar, tooltip

**확장 포인트**:
- 새로운 컴포넌트 추가 (AVAILABLE_COMPONENTS에 추가)
- 프레임워크 지원 추가 (VALID_FRAMEWORKS에 추가)
- 커스텀 설정 파일 생성

---

### 3️⃣ Backend Plugin 아키텍처

**목적**: FastAPI 기반 백엔드 프로젝트 초기화

**클래스 구조**:
- `InitFastAPICommand` (FastAPI 초기화)
- `DBSetupCommand` (데이터베이스 설정)
- `ResourceCRUDCommand` (REST API 생성)

**주요 메서드** (InitFastAPICommand):
1. `validate_project_name()` - 프로젝트명 검증
2. `validate_database()` - 데이터베이스 검증
3. `create_project_directory()` - 디렉토리 생성
4. `create_fastapi_structure()` - FastAPI 구조 생성
5. `execute()` - 메인 오케스트레이션

**생성 파일**:
```
{project}/
├── main.py                  # FastAPI 메인 앱
├── requirements.txt         # 의존성
├── .env.example             # 환경 변수 템플릿
├── app/
│   ├── __init__.py
│   ├── database.py          # DB 연결
│   ├── models/
│   │   └── {resource}.py
│   └── routes/
│       └── {resource}.py
└── database/
    └── {database}.db        # 로컬 DB (SQLite)
```

**지원 데이터베이스**:
- PostgreSQL (psycopg2)
- MySQL (pymysql)
- SQLite (기본 내장)
- MongoDB (pymongo)

**확장 포인트**:
- 새로운 데이터베이스 드라이버 추가
- 인증 전략 변경 (JWT, OAuth)
- 추가 미들웨어 통합

---

### 4️⃣ Frontend Plugin 아키텍처

**목적**: React 프로젝트 초기화 및 상태 관리 설정

**클래스 구조**:
- `InitReactCommand` (React 초기화)
- `SetupStateCommand` (상태 관리 설정)
- `SetupTestingCommand` (테스팅 설정)

**주요 메서드** (InitReactCommand):
1. `validate_project_name()` - 프로젝트명 검증
2. `validate_framework()` - 프레임워크 검증
3. `create_project_directory()` - 디렉토리 생성
4. `create_react_structure()` - React 구조 생성
5. `execute()` - 메인 오케스트레이션

**생성 파일**:
```
{project}/
├── src/
│   ├── App.jsx
│   ├── main.jsx
│   ├── App.css
│   └── store/
│       └── useStore.js
├── public/
│   └── index.html
├── package.json
├── vite.config.js
└── __tests__/
    └── App.test.jsx
```

**지원 옵션**:
- 프레임워크: React, Vite, Next.js
- 상태관리: Context, Zustand, Redux, Recoil
- 테스팅: Vitest, Jest
- 스타일링: CSS, Tailwind

**확장 포인트**:
- 새로운 상태 관리 라이브러리 추가
- 컴포넌트 제너레이터 확장
- 테스팅 템플릿 커스터마이징

---

### 5️⃣ DevOps Plugin 아키텍처

**목적**: Docker, CI/CD, Kubernetes 인프라 설정

**클래스 구조**:
- `SetupDockerCommand` (Docker 설정)
- `SetupCICommand` (CI/CD 파이프라인)
- `SetupK8sCommand` (Kubernetes 설정)

**주요 메서드** (SetupDockerCommand):
1. `validate_app_type()` - 앱 타입 검증
2. `create_dockerfile()` - Dockerfile 생성
3. `execute()` - 메인 오케스트레이션

**생성 파일**:
```
{project}/
├── Dockerfile                    # 컨테이너 이미지
├── .dockerignore                # 무시 파일 목록
├── docker-compose.yml           # 멀티컨테이너 (선택)
├── .github/workflows/
│   └── ci.yml                   # GitHub Actions
├── .gitlab-ci.yml               # GitLab CI
└── k8s/
    ├── deployment.yaml          # K8s Deployment
    ├── service.yaml             # K8s Service
    └── ingress.yaml             # Ingress (선택)
```

**지원 언어**:
- Python (3.11-slim)
- Node.js (20-alpine)
- Go (1.21, 다단계 빌드)
- Java (OpenJDK 21)

**CI/CD 플랫폼**:
- GitHub Actions
- GitLab CI
- CircleCI

**확장 포인트**:
- 새로운 언어 Dockerfile 추가
- CI/CD 파이프라인 커스터마이징
- Kubernetes CRD 지원 추가

---

## 🔗 플러그인 간 연계

### 표준 사용 순서

```
1. PM Plugin로 SPEC 생성
   └─> 요구사항 명확화

2. UI/UX Plugin + Frontend Plugin으로 UI 구현
   └─> 프론트엔드 프로토타입

3. Backend Plugin으로 API 구현
   └─> REST API 엔드포인트

4. DevOps Plugin으로 배포 설정
   └─> Docker + CI/CD + K8s 설정
```

### 데이터 공유 패턴

플러그인 간 데이터 공유는 파일 시스템을 통해 이루어집니다:

```
PM Plugin 출력 (.moai/specs/)
  ↓ (SPEC 문서)
Backend/Frontend Plugin 입력 (API 설계, UI 구조)
  ↓
DevOps Plugin 입력 (배포 대상 설정)
```

---

## ✅ 품질 관리

### 테스트 전략

**각 플러그인은 다음 테스트를 포함합니다**:

1. **유닛 테스트**: 검증 메서드 단위 테스트
2. **통합 테스트**: 전체 워크플로우 테스트
3. **경계값 테스트**: 입력 범위 테스트
4. **에러 핸들링**: 예외 처리 테스트
5. **성능 테스트**: 실행 시간 검증

### 테스트 커버리지 기준

- **최소**: 85%
- **목표**: 95%+
- **모든 플러그인**: 94-100% 달성

### 타입 안전성

모든 플러그인은 `mypy strict` 모드에서:
- ✅ 0개 타입 오류
- ✅ Optional[Path] 타입 정확성
- ✅ Dict[str, Any] 명시적 선언

---

## 🔒 보안 고려사항

### 각 플러그인의 보안 정책

**PM Plugin**:
- SPEC ID 자동 생성으로 사용자 입력 최소화
- 파일 권한 기본값 (600)

**UI/UX Plugin**:
- 컴포넌트 화이트리스트 (19개만 허용)
- 프레임워크 검증으로 injection 방지

**Backend Plugin**:
- 데이터베이스 URL 환경변수 분리
- 기본 CORS 정책 (프로덕션에서 수정 필요)

**Frontend Plugin**:
- API 엔드포인트 환경변수 분리
- 상태 관리 보안 모범 사례

**DevOps Plugin**:
- 최소한의 base image (slim, alpine)
- 다단계 빌드로 최종 이미지 크기 최소화
- .dockerignore로 민감한 파일 제외

---

## 📈 확장성

### 새로운 플러그인 추가 방법

1. **디렉토리 구조 생성**:
   ```
   moai-alfred-newplugin/
   ├── newplugin_plugin/
   │   ├── __init__.py
   │   └── commands.py
   ├── tests/
   │   └── test_commands.py
   └── README.md
   ```

2. **CommandResult 정의**:
   ```python
   @dataclass
   class CommandResult:
       success: bool
       output_dir: Optional[Path]
       files_created: List[str]
       message: str
       error: Optional[str] = None
   ```

3. **Command 클래스 구현**:
   ```python
   class NewPluginCommand:
       def validate_input(self) -> bool:
           pass

       def create_output(self) -> None:
           pass

       def execute(self, **kwargs) -> CommandResult:
           pass
   ```

4. **테스트 작성** (최소 85% 커버리지)

5. **문서화** (README, USAGE, CHANGELOG)

### 플러그인 마켓플레이스 구조

```
moai-marketplace/
└── plugins/
    ├── moai-alfred-pm/           # 공식 플러그인
    ├── moai-alfred-uiux/
    ├── moai-alfred-backend/
    ├── moai-alfred-frontend/
    ├── moai-alfred-devops/
    └── (future community plugins)
```

---

## 📚 참고 자료

- SPEC-V1-001: 전체 요구사항 명세
- .moai/reports/: 플러그인 감시 및 테스트 보고서
- CHANGELOG.md: 버전 변경 사항
- 각 플러그인 README.md: 세부 사용 가이드

---

## 🤝 커뮤니티 기여

### 플러그인 기여 가이드

1. Feature branch 생성 (`feature/plugin-name`)
2. 테스트 작성 (TDD 원칙 준수)
3. Pull Request 제출
4. 코드 검토 완료
5. 마스터 브랜치 병합

### 질문 및 피드백

- GitHub Issues: 버그 및 기능 요청
- Discussions: 아이디어 및 제안
- CLAUDE.md: 프로젝트 지침 확인

---

**MoAI-ADK v1.0 플러그인 생태계는 지속적으로 성장하고 있습니다.**

각 플러그인은 프로덕션 준비 완료 상태이며, 팀 환경에서 즉시 사용 가능합니다.

Happy Plugin Development! 🚀

