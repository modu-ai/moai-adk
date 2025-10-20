---
name: alfred:0-project
description: 프로젝트 문서 초기화 - product/structure/tech.md 생성 및 언어별 최적화 설정 (Sub-agents 기반 리팩토링)
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash(ls:*)
  - Bash(grep:*)
  - Task
---

# 📋 MoAI-ADK 0단계: 프로젝트 문서 초기화 (Sub-agents 조율)

## 🎯 커맨드 목적

프로젝트 환경을 자동 분석하여 product/structure/tech.md 문서를 생성/갱신하고 언어별 최적화 설정을 구성합니다.

**핵심 변경사항** (v0.4.0+):
- ✅ 6개 Sub-agents 조율 방식으로 리팩토링 (991 lines → 300 lines)
- ✅ Alfred는 조율자 역할만 수행 (Task tool 활용)
- ✅ 복잡한 로직은 Sub-agents로 위임

## 📋 실행 흐름 (2단계)

### Phase 1: 분석 및 계획 수립

1.0 **백업 확인** (Alfred 직접)
1.1 **백업 병합** (조건부, backup-merger 호출)
1.2 **언어 감지** (language-detector 호출)
1.3 **프로젝트 인터뷰** (project-interviewer 호출)
1.4 **사용자 승인 대기**

### Phase 2: 실행 (사용자 승인 후)

2.1 **문서 생성** (document-generator 호출)
2.2 **config.json 생성** (Alfred 직접)
2.3 **품질 검증** (선택적, trust-checker 호출)

### Phase 3: 최적화 (선택적)

3.1 **기능 선택** (feature-selector 호출)
3.2 **템플릿 최적화** (template-optimizer 호출)
3.3 **완료 보고**

## 🔗 연관 에이전트 (6개 Sub-agents)

| 에이전트 | 모델 | 역할 | 호출 시점 |
|---------|------|------|----------|
| **backup-merger** 📦 | Sonnet | 백업 병합 | 백업 존재 시 |
| **language-detector** 🔍 | Haiku | 언어 감지 | 항상 |
| **project-interviewer** 💬 | Sonnet | 요구사항 수집 | 신규/레거시 분석 |
| **document-generator** 📝 | Haiku | 문서 생성 | Phase 2 시작 |
| **feature-selector** 🎯 | Haiku | 스킬 선택 | Phase 3 (선택적) |
| **template-optimizer** ⚙️ | Haiku | 템플릿 최적화 | Phase 3 (선택적) |

## ⚠️ 금지 사항

**절대 하지 말아야 할 작업**:
- ❌ `.claude/memory/` 디렉토리에 파일 생성
- ❌ `.claude/commands/alfred/*.json` 파일 생성
- ❌ 기존 문서 불필요한 덮어쓰기
- ❌ 날짜와 수치 예측 ("3개월 내", "50% 단축" 등)

**사용해야 할 표현**:
- ✅ "우선순위 높음/중간/낮음"
- ✅ "즉시 필요", "단계적 개선"
- ✅ 현재 확인 가능한 사실

---

## 🚀 Phase 1: 분석 및 계획 수립

### 1.0 백업 디렉토리 확인 (Alfred 직접)

**목적**: moai-adk init 재초기화 후 백업 파일 존재 여부 확인

Alfred가 직접 실행:
```bash
# 백업 존재 확인
ls -t .moai-backups/ | head -1

# config.json의 optimized 플래그 확인
grep "optimized" .moai/config.json
```

**백업 존재 조건**:
- `.moai-backups/` 디렉토리 존재
- 최신 백업 폴더에 `.moai/project/*.md` 파일 존재
- `config.json`의 `optimized: false` (재초기화 직후)

**백업 존재 시 사용자 선택**:
```markdown
백업 파일(.moai-backups/{timestamp}/)이 발견되었습니다. 어떻게 처리하시겠습니까?

1️⃣ **병합**: 백업 파일의 사용자 커스터마이징을 최신 템플릿에 병합 (권장)
2️⃣ **새로 작성**: 백업 무시하고 새로운 인터뷰 시작
3️⃣ **건너뛰기**: 현재 파일 유지 (작업 종료)
```

**응답 처리**:
- "병합" → 1.1 백업 병합 워크플로우
- "새로 작성" → 1.2 언어 감지
- "건너뛰기" → 작업 종료

**백업 없음** → 1.2 언어 감지로 바로 진행

---

### 1.1 백업 병합 워크플로우 (조건부)

**조건**: 사용자가 "병합" 선택 시

**Alfred 작업**:
```python
# 의사코드
Task(
    subagent_type="backup-merger",
    description="백업 파일(.moai-backups/)과 최신 템플릿 병합",
    prompt=f"""
    백업 디렉토리: {backup_dir}
    병합 대상: product.md, structure.md, tech.md

    작업:
    1. 백업 파일 읽기
    2. 템플릿 기본값 탐지
    3. 사용자 커스터마이징 추출
    4. 최신 템플릿과 병합
    5. HISTORY 섹션 업데이트
    6. 버전 업데이트 (Patch 증가)
    """
)
```

**backup-merger 산출물**:
- 병합된 product/structure/tech.md
- 병합 보고서 (복원된 섹션, 교체된 섹션)
- 버전 업데이트 (v0.1.x → v0.1.x+1)

**병합 완료 후** → 1.2 언어 감지로 진행

---

### 1.2 언어 감지 (항상 실행)

**Alfred 작업**:
```python
# 의사코드
language_result = Task(
    subagent_type="language-detector",
    description="프로젝트 언어 및 프레임워크 자동 감지",
    prompt="""
    작업:
    1. 설정 파일 스캔 (pyproject.toml, package.json 등)
    2. 언어 감지
    3. 프레임워크 분석
    4. 도구 체인 추천 (LanguageInterface 표준)
    5. JSON 응답 생성
    """
)

# 결과 저장
detected_language = language_result["language"]
detected_framework = language_result["framework"]
```

**language-detector 산출물**:
```json
{
  "language": "Python",
  "framework": "FastAPI",
  "test_framework": "pytest",
  "linter": "ruff",
  "formatter": "black",
  "type_checker": "mypy",
  "package_manager": "uv",
  "version_requirement": ">=3.11"
}
```

---

### 1.3 프로젝트 인터뷰 (신규/레거시)

**Alfred 작업**:
```python
# 의사코드
interview_result = Task(
    subagent_type="project-interviewer",
    description="프로젝트 요구사항 인터뷰 수행",
    prompt=f"""
    프로젝트 유형: {project_type}  # greenfield or legacy
    언어: {detected_language}
    프레임워크: {detected_framework}

    작업:
    1. 프로젝트 유형별 질문 트리 선택
    2. 4단계 인터뷰 진행 (비즈니스 → 아키텍처 → 기술 → 팀)
    3. 모호한 응답 명확화
    4. JSON 응답 생성
    """
)

# 결과 저장
product_info = interview_result["product"]
structure_info = interview_result["structure"]
tech_info = interview_result["tech"]
team_info = interview_result["team"]
```

**project-interviewer 산출물**:
```json
{
  "product": {
    "user_segments": ["초급 개발자", "시니어 개발자"],
    "problems": ["테스트 없는 레거시", "요구사항 불일치"],
    "strategy": ["SPEC-First", "자동 TAG"],
    "success_metrics": ["커버리지 85%", "SPEC 준수 100%"]
  },
  "structure": {...},
  "tech": {...},
  "team": {...}
}
```

---

### 1.4 계획 보고서 생성 및 사용자 승인

**Alfred 작업**:
```markdown
## 📊 프로젝트 초기화 계획

### 감지된 환경
- **언어**: Python (>=3.11)
- **프레임워크**: FastAPI
- **프로젝트 유형**: 신규 프로젝트 (Greenfield)

### 수집된 정보
- **사용자층**: 2개 세그먼트
- **핵심 문제**: 2개
- **아키텍처**: 모놀리식 백엔드
- **팀 모드**: Personal (1인 개발)

### 생성할 문서
- ✅ `.moai/project/product.md` (EARS 구문 적용)
- ✅ `.moai/project/structure.md`
- ✅ `.moai/project/tech.md`
- ✅ `.moai/config.json`

### 다음 단계 옵션
1️⃣ **진행**: Phase 2 (문서 생성) 실행
2️⃣ **수정**: 인터뷰 재수행
3️⃣ **중단**: 작업 취소

응답: "진행" 또는 "수정" 또는 "중단"
```

**사용자 응답 처리**:
- "진행" → Phase 2로 진행
- "수정" → 1.3 재수행
- "중단" → 작업 종료

---

## 🚀 Phase 2: 실행 (사용자 승인 후)

### 2.1 문서 생성 (document-generator)

**Alfred 작업**:
```python
# 의사코드
doc_result = Task(
    subagent_type="document-generator",
    description="product/structure/tech.md 문서 생성",
    prompt=f"""
    입력 데이터:
    {json.dumps(interview_result)}

    작업:
    1. EARS 구문 적용
    2. YAML Front Matter 생성 (필수 7개 필드)
    3. HISTORY 섹션 추가 (v0.0.1 INITIAL)
    4. product.md 작성
    5. structure.md 작성
    6. tech.md 작성
    """
)
```

**document-generator 산출물**:
- `.moai/project/product.md` (EARS 적용률 85%)
- `.moai/project/structure.md`
- `.moai/project/tech.md`
- 문서 생성 보고서

---

### 2.2 config.json 생성 (Alfred 직접)

**Alfred 작업**:
```python
# 의사코드
config = {
    "project": {
        "name": project_name,
        "version": "0.0.1",
        "mode": team_info["mode"],  # personal or team
        "locale": "ko"
    },
    "optimized": False
}

Write(".moai/config.json", json.dumps(config, indent=2))
```

---

### 2.3 품질 검증 (선택적)

**Alfred 작업** (사용자 요청 시):
```python
# 의사코드
Task(
    subagent_type="trust-checker",
    description="프로젝트 초기 구조 TRUST 원칙 검증",
    prompt="""
    검증 대상:
    - .moai/project/*.md 필수 필드 완전성
    - YAML Front Matter 형식
    - HISTORY 섹션 존재
    - EARS 구문 적용률
    """
)
```

---

## 🚀 Phase 3: 최적화 (선택적)

### 3.1 기능 선택 (feature-selector)

**Alfred 작업**:
```python
# 의사코드
feature_result = Task(
    subagent_type="feature-selector",
    description="49개 스킬 중 3~9개 최적 선택",
    prompt=f"""
    프로젝트 정보:
    - 언어: {detected_language}
    - 프레임워크: {detected_framework}
    - 도메인: backend
    - 팀 우선순위: {team_info["priority_areas"]}

    작업:
    1. Tier 1 (Core) 5개 자동 포함
    2. Tier 2 (Language) 1개 선택
    3. Tier 3 (Domain) 0~3개 선택
    4. Tier 4 (Essentials) 선택적
    5. JSON 응답 생성
    """
)
```

**feature-selector 산출물**:
```json
{
  "selected_skills": [
    {"tier": 1, "name": "moai-claude-code"},
    {"tier": 1, "name": "moai-foundation-langs"},
    {"tier": 1, "name": "moai-foundation-specs"},
    {"tier": 1, "name": "moai-foundation-ears"},
    {"tier": 1, "name": "moai-foundation-tags"},
    {"tier": 2, "name": "moai-lang-python"},
    {"tier": 3, "name": "moai-domain-backend"},
    {"tier": 3, "name": "moai-domain-web-api"}
  ],
  "total_selected": 8,
  "recommendation": "49개 → 8개 (84% 감소)"
}
```

---

### 3.2 템플릿 최적화 (template-optimizer)

**Alfred 작업**:
```python
# 의사코드
optimize_result = Task(
    subagent_type="template-optimizer",
    description="CLAUDE.md 맞춤형 생성 및 불필요한 스킬 파일 삭제",
    prompt=f"""
    선택된 스킬: {feature_result["selected_skills"]}

    작업:
    1. 맞춤형 CLAUDE.md 생성
    2. 불필요한 스킬 파일 삭제 (41개)
    3. config.json 업데이트 (optimized: true)
    4. 최적화 보고서 생성
    """
)
```

**template-optimizer 산출물**:
- 맞춤형 CLAUDE.md
- 삭제된 스킬 목록 (41개)
- 디스크 절약량 (12.5 MB)
- config.json 업데이트

---

### 3.3 완료 보고

**Alfred 작업**:
```markdown
## ✅ 프로젝트 초기화 완료

### Phase 1: 분석 및 계획
- ✅ 언어 감지: Python + FastAPI
- ✅ 프로젝트 인터뷰: 12개 질문 완료

### Phase 2: 문서 생성
- ✅ product.md (EARS 적용률 85%)
- ✅ structure.md
- ✅ tech.md
- ✅ config.json

### Phase 3: 최적화
- ✅ 스킬 선택: 49개 → 8개 (84% 감소)
- ✅ CLAUDE.md 맞춤형 생성
- ✅ 디스크 절약: 12.5 MB

### 다음 단계
1. 프로젝트 문서 검토 (.moai/project/*.md)
2. SPEC 작성 시작 (/alfred:1-spec)
```

---

## 📋 에러 처리

**Sub-agent 호출 실패**:
```markdown
❌ language-detector 호출 실패: 설정 파일 없음
  → pyproject.toml 또는 package.json 생성 후 재시도
```

**사용자 응답 없음**:
```markdown
⚠️ 사용자 응답 없음 (3분 초과)
  → 기본값으로 진행: Personal 모드, 중간 우선순위
```

**백업 병합 충돌**:
```markdown
⚠️ 백업 병합 충돌: STRATEGY 섹션이 백업과 템플릿에서 다름
  → 어떻게 처리? (백업 우선 / 템플릿 우선)
```

---

**라인 수**: ~295 lines (목표 300 lines 달성)
**리팩토링 비율**: 991 lines → 295 lines (70% 감소)
**Sub-agents 활용**: 6개 (backup-merger, language-detector, project-interviewer, document-generator, feature-selector, template-optimizer)
