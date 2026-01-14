# MoAI-ADK 기술 스택

> **최종 업데이트**: 2026-01-15
> **버전**: 4.1.0

---

## 런타임 환경

### Python
- **지원 버전**: Python 3.11, 3.12, 3.13, 3.14
- **목표 버전**: Python 3.14 (개발용)
- **패키지 매니저**: uv (권장), pip (지원)

---

## 핵심 의존성

### CLI 프레임워크
| 패키지 | 버전 | 목적 |
|---------|---------|---------|
| click | >=8.1.0 | CLI 명령 프레임워크 |
| rich | >=13.0.0 | 리치 텍스트 포맷팅 및 터미널 UI |
| pyfiglet | >=1.0.2 | ASCII 아트 텍스트 생성 |
| questionary | >=2.0.0 | 대화형 CLI 프롬프트 |
| InquirerPy | >=0.3.4 | 퍼지 검색이 있는 현대적 CLI 프롬프트 |

### Git 통합
| 패키지 | 버전 | 목적 |
|---------|---------|---------|
| gitpython | >=3.1.45 | Git 작업 및 저장소 관리 |

### 설정 & 템플릿팅
| 패키지 | 버전 | 목적 |
|---------|---------|---------|
| pyyaml | >=6.0 | YAML 파싱 및 직렬화 |
| jinja2 | >=3.0.0 | 코드 생성용 템플릿 엔진 |

### HTTP & 비동기
| 패키지 | 버전 | 목적 |
|---------|---------|---------|
| requests | >=2.28.0 | REST API용 HTTP 클라이언트 |
| aiohttp | >=3.13.2 | 비동기 HTTP 클라이언트 |

### 시스템 & 유틸리티
| 패키지 | 버전 | 목적 |
|---------|---------|---------|
| psutil | >=7.1.3 | 시스템 및 프로세스 유틸리티 |
| packaging | >=21.0 | 버전 파싱 및 비교 |

### AI 통합
| 패키지 | 버전 | 목적 |
|---------|---------|---------|
| google-genai | >=1.0.0 | 이미지 생성용 Gemini API (nano-banana) |
| pillow | >=10.0.0 | 이미지 처리 |
| anthropic | >=0.40.0 | Claude API SDK (선택사항, 웹 채팅용) |

---

## 개발 의존성

### 테스트
| 패키지 | 버전 | 목적 |
|---------|---------|---------|
| pytest | >=8.4.2 | 테스트 프레임워크 |
| pytest-cov | >=7.0.0 | 커버리지 보고 |
| pytest-xdist | >=3.8.0 | 병렬 테스트 실행 |
| pytest-asyncio | >=1.2.0 | 비동기 테스트 지원 |
| pytest-mock | >=3.15.1 | 모킹 지원 |

### 코드 품질
| 패키지 | 버전 | 목적 |
|---------|---------|---------|
| ruff | >=0.1.0 | 빠른 Python 린터 및 포맷터 |
| mypy | >=1.7.0 | 정적 타입 검사 |
| black | >=24.0.0 | 코드 포맷팅 |
| types-PyYAML | >=6.0.0 | PyYAML용 타입 스텁 |

### 보안
| 패키지 | 버전 | 목적 |
|---------|---------|---------|
| pip-audit | >=2.7.0 | 의존성 취약점 스캔 |
| bandit | >=1.8.0 | Python용 보안 린터 |

---

## 외부 도구

### AST-Grep
- **도구**: ast-grep (sg CLI)
- **목적**: 구조적 코드 검색 및 리팩토링
- **언어**: 40개 이상 프로그래밍 언어 지원
- **용도**: 보안 스캔, 코드 패턴 매칭, 대규모 리팩토링

### Claude Code
- **플랫폼**: Anthropic Claude Code CLI
- **버전**: 최신
- **목적**: AI 기반 개발 어시스턴트
- **통합**: `.claude/` 설정 디렉토리를 통해

---

## 빌드 시스템

### 패키지 빌딩
- **빌드 백엔드**: hatchling
- **설정**: pyproject.toml (PEP 517/518)

### 배포
- **패키지 이름**: moai-adk
- **CLI 진입점**:
  - `moai-adk` - 메인 CLI
  - `moai` - moai-adk 별칭
  - `moai-worktree` - Git 워크트리 관리

---

## 품질 설정

### Ruff (린팅)
```toml
line-length = 120
target-version = "py314"
select = ["E", "F", "W", "I", "N"]
```

### 커버리지
```toml
fail_under = 85
precision = 2
show_missing = true
```

### MyPy
```toml
python_version = "3.14"
ignore_missing_imports = true
```

---

## 아키텍처 패턴

### 멀티레이어 모듈러 아키텍처
1. **CLI 레이어**: Click 기반 명령형 인터페이스
2. **코어 레이어**: 45개 모듈로 구성된 비즈니스 로직
3. **파운데이션 레이어**: TRUST 5, Git 작업
4. **Claude Code 통합 레이어**: 21개 에이전트, 48개 스킬

### 에이전트 아키텍처 (7-Tier)
1. **Tier 0**: 코어 에이전트 (Alfred 오케스트레이터)
2. **Tier 1**: 매니저 에이전트 (워크플로우 조정)
3. **Tier 2**: 전문가 에이전트 (도메인 전문가)
4. **Tier 3**: 빌더 에이전트 (생성 전문가)
5. **Tier 4**: 유틸리티 에이전트 (헬퍼 함수)

### 스킬 아키텍처
- **파운데이션 스킬**: 핵심 원칙 및 패턴
- **도메인 스킬**: 프론트엔드, 백엔드, 데이터베이스
- **언어 스킬**: Python, TypeScript, Go, Rust 등
- **플랫폼 스킬**: Firebase, Supabase, Vercel 등
- **워크플로우 스킬**: TDD, SPEC, 프로젝트 관리

### 설정 아키텍처
- **모듈형 섹션**: 관심사별 분리 (user, language, project, git, quality)
- **YAML 형식**: 사람이 읽기 쉬운 설정
- **템플릿 변수**: 배포용 `{{VARIABLE}}`
- **런타임 확인**: 환경 변수 및 동적 값

---

## 통합 지점

### Claude Code 통합
- `.claude/agents/`의 사용자 정의 에이전트
- `.claude/commands/`의 슬래시 명령
- `.claude/hooks/`의 후크
- `.claude/skills/`의 스킬

### GitHub 통합
- CI/CD용 워크플로우 템플릿
- SPEC에서 GitHub 이슈로 동기화
- 브랜치 보호 및 PR 자동화

### LSP 통합
- 진단용 Language Server Protocol
- 실시간 오류 감지
- 자동 수정 제안

---

## 보안 고려사항

### 의존성 보안
- 정기 `pip-audit` 스캔
- Bandit 정적 분석
- GitHub Dependabot 알림

### 코드 보안
- AST-grep 보안 패턴
- OWASP 취약점 감지
- XSS, CSRF, SQL 인젝션 방지

### 설정 보안
- 비밀용 `.env` 파일
- git-ignored 민감 파일
- 안전한 파일 권한 (0o600)

---

## 성능 최적화

### 토큰 효율성 (v4.1.0)
- 단순화된 컨텍스트 수집
- 조건부 스킬 자동 로딩
- 간단한 작업용 빠른 참조
- 긴 세션용 컨텍스트 압축

### 병렬 실행
- 모든 독립 작업은 기본적으로 병렬 실행
- 필요시 순차 실행용 `--sequential` 플래그
- 다단계 작업의 워크플로우 성능 3-4배 향상
- 시스템 리소스의 더 나은 활용

### 빌드 성능
- 병렬 테스트 실행 (pytest-xdist)
- 증분 타입 검사 (mypy)
- 빠른 린팅 (flake8 대신 ruff)

### 런타임 성능
- 가능한 곳에서 비동기 작업
- 구성 가능한 TTL로 캐싱
- 리소스 지연 로딩
- 더 빠른 세션 시작/종료를 위한 백그라운드 git 후크

---

## 프레임워크 및 라이브러리 버전

### 핵심 프레임워크
- Click 8.1+ - CLI 프레임워크
- Rich 13.0+ - 터미널 UI
- PyYAML 6.0+ - 설정 관리
- Jinja2 3.0+ - 템플릿 엔진
- GitPython 3.1+ - Git 작업

### 개발 도구
- pytest 8.4+ - 테스트 프레임워크
- ruff 0.1+ - 린터 및 포맷터
- mypy 1.7+ - 타입 검사기
- hatchling - 빌드 백엔드

### AI 및 ML
- Anthropic SDK 0.40+ - Claude API
- Google GenAI 1.0+ - Gemini API
- Pillow 10.0+ - 이미지 처리

---

## 플랫폼 지원

### 운영 체제
- Linux (Ubuntu 20.04+, Debian 11+, CentOS 8+)
- macOS (12+ Monterey, 13+ Ventura, 14+ Sonoma)
- Windows (10+, 11 with WSL2)

### Python 버전
- Python 3.11 (최소 지원)
- Python 3.12 (권장)
- Python 3.13 (최신 안정)
- Python 3.14 (개발 목표)

---

## 모니터링 및 로깅

### 로깅
- 구조화된 JSON 로그
- 로그 레벨: DEBUG, INFO, WARNING, ERROR, CRITICAL
- 로그 파일: `.moai/logs/`
- 로그 회전: 크기 및 날짜 기반

### 성능 모니터링
- 응답 시간 추적
- 메모리 사용량 모니터링
- 토큰 사용량 추적
- 에이전트 실행 시간 측정

### 분석
- 사용 패턴 추적
- 명령 사용 통계
- 에이전트 성공률
- `.moai/analytics/`에 저장
