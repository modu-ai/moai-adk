---
description: 프로젝트 진행 상황 모니터링 대시보드
argument-hint: [--detail|--export] [output-path]
allowed-tools: Read, Bash, Task
---

# 📊 MoAI-ADK 프로젝트 대시보드

프로젝트의 현재 상태를 한눈에 파악할 수 있는 종합 모니터링 대시보드입니다. 파이프라인 진행률, SPEC 현황, TAG 시스템 건강도, Constitution 준수 현황 등을 시각적으로 표시합니다.

> Claude Code 화면에 컬러풀한 대시보드가 출력되며, 실시간 스냅샷을 제공합니다.

## 사용법

### 기본 사용법
```bash
/moai:7-dashboard
```
프로젝트의 핵심 상태 정보를 요약하여 표시합니다.

### 상세 정보 표시
```bash
/moai:7-dashboard --detail
```
더 많은 세부 정보와 통계를 포함하여 표시합니다.

### 내보내기
```bash
/moai:7-dashboard --export
/moai:7-dashboard --export reports/dashboard.html
```
대시보드를 HTML 또는 Markdown 파일로 내보냅니다.

## 자연어 체이닝 오케스트레이션

🤖 MoAI-ADK 프로젝트 대시보드를 표시합니다. 현재 상태를 종합적으로 분석하여 시각적으로 출력하겠습니다.

### 단계별 진행

1. **프로젝트 상태 수집**: .moai/indexes/ 디렉토리의 상태 파일들을 읽어 현재 상황을 파악합니다.
2. **Git 정보 수집**: 현재 브랜치, 최근 커밋, 변경사항 등을 수집합니다.
3. **대시보드 렌더링**: Python 스크립트를 실행하여 Rich 라이브러리로 시각화합니다.
4. **결과 출력**: Claude Code 화면에 컬러풀한 대시보드를 표시합니다.

### 모드별 처리

#### 기본 모드
```
🎯 기본 대시보드 표시 중...

📊 핵심 지표 수집:
├── 파이프라인 진행률 분석
├── SPEC 현황 요약
├── TAG 시스템 건강도
├── Constitution 준수 상태
└── 최근 활동 내역

💻 시각화 렌더링:
├── Rich 라이브러리 활용
├── 프로그레스 바 생성
├── 테이블 포맷팅
└── 컬러 코딩 적용
```

#### 상세 모드 (--detail)
```
🔍 상세 대시보드 표시 중...

📈 확장 지표 수집:
├── 파일별 변경 내역
├── 개별 SPEC 세부 상태
├── TAG 카테고리별 분석
├── Constitution 위반 세부사항
├── 테스트 커버리지 상세
├── 성능 지표 분석
└── 의존성 상태 확인

📋 추가 정보:
├── 브랜치별 커밋 히스토리
├── 활성 TODO 항목
├── 경고 및 권고사항
└── 최적화 제안
```

#### 내보내기 모드 (--export)
```
💾 대시보드 내보내기 중...

📁 파일 생성:
├── HTML 버전: 웹 브라우저용
├── Markdown 버전: 문서 통합용
├── 스타일시트: CSS 적용
└── 자산 파일: 이미지, 아이콘

📤 출력 위치:
├── 기본: ./dashboard-{timestamp}.html
├── 지정: 사용자 정의 경로
└── 권한: 파일 접근성 확인
```

## 대시보드 레이아웃

### 헤더 섹션
```
╭─────────────────────── 🗿 MoAI-ADK Dashboard ───────────────────────╮
│ 프로젝트: {프로젝트명} v{버전} | 단계: {현재단계} | {현재시각}        │
╰──────────────────────────────────────────────────────────────────────╯
```

### 파이프라인 진행률
```
📊 개발 파이프라인 진행 상황
┌──────────────────────────────────────────────────────────────────────┐
│ SPECIFY  ████████████████████ 100% [2/2] ✅                         │
│ PLAN     ████████░░░░░░░░░░░░  40% [1/2] ← 현재                      │
│ TASKS    ░░░░░░░░░░░░░░░░░░░░   0% [0/0]                             │
│ IMPLEMENT░░░░░░░░░░░░░░░░░░░░   0% [0/0]                             │
│ SYNC     ░░░░░░░░░░░░░░░░░░░░   0% [0/0]                             │
└──────────────────────────────────────────────────────────────────────┘
```

### SPEC 현황 테이블
```
📋 SPEC 현황
┌──────────┬─────────────────────────┬──────────┬─────────┬──────────┐
│ SPEC ID  │ 제목                    │ 상태     │ 진행률  │ 담당자   │
├──────────┼─────────────────────────┼──────────┼─────────┼──────────┤
│ SPEC-001 │ 프로젝트 초기화         │ ✅ 완료  │ 100%    │ System   │
│ SPEC-002 │ 명령어 시스템           │ ✅ 완료  │ 100%    │ System   │
│ SPEC-004 │ 대시보드 기능           │ 🔄 진행  │  60%    │ User     │
└──────────┴─────────────────────────┴──────────┴─────────┴──────────┘
```

### TAG 시스템 건강도
```
🏷️ TAG 시스템 (건강도: 100% ✅)
├── STEERING: 3개 (@VISION, @STRUCT, @TECH)
├── SPEC: 3개 (@REQ:*)
├── IMPLEMENTATION: 0개
├── QUALITY: 0개
└── 전체 연결성: 완전 (끊어진 링크 0개)
```

### Constitution 준수 현황
```
⚖️ Constitution 5원칙 준수 현황
✅ Simplicity    프로젝트 수 1/3개 (67% 여유)
✅ Architecture  모든 기능이 라이브러리 구조
⚠️ Testing      커버리지 65% (목표: 80%, 부족: 15%)
✅ Observability 구조화 로깅 활성화됨
✅ Versioning    0.1.25 (MAJOR.MINOR.BUILD 준수)
```

### Git 상태 정보
```
🔀 Git 상태
├── 현재 브랜치: develop
├── 최근 커밋: 95e5356 - 🚀 feat: enhance session_start_notice (22분 전)
├── 변경사항: 수정 6개, 삭제 2개, 미추적 4개
├── 리모트 동기화: ↑2 ↓0 (2개 커밋 앞섬)
└── 작업 상태: 🟡 진행 중 (커밋 필요)
```

### 추천 액션
```
💡 다음 단계 추천
1. 🚀 /moai:3-plan SPEC-004  # Constitution 검증 및 계획 수립
2. 📊 pytest --cov=80        # 테스트 커버리지 목표 달성
3. 🔄 git add . && git commit # 변경사항 커밋
4. 📝 /moai:6-sync auto      # 문서 동기화

⚠️ 주의사항
- 미추적 파일 4개 확인 필요
- Testing Constitution 위반 해결 권장
```

### 데이터 수집 단계

**1단계: 데이터 수집 스크립트 실행**

기본 모드에서는 `collect_dashboard_data.py` 스크립트를 실행하여 JSON 형태로 모든 상태 정보를 수집합니다:

```bash
python3 .moai/scripts/collect_dashboard_data.py --format compact
```

**2단계: JSON 파싱 및 분석**

수집된 JSON 데이터를 파싱하여 다음 정보를 추출합니다:
- 프로젝트 메타데이터 (이름, 버전, 브랜치 등)
- 파이프라인 진행 상황 (각 단계별 완료 여부 및 진행률)
- SPEC 현황 (ID, 제목, 상태, 우선순위)
- TAG 시스템 (총 태그 수, 카테고리별 분포, 건강도)
- Constitution 준수 현황 (5원칙 각각의 상태)
- Git 상태 (브랜치, 최근 커밋, 변경사항)
- 추천 액션 및 경고사항

**3단계: 구조화된 출력 렌더링**

파싱된 데이터를 기반으로 유니코드 박스 문자와 이모지를 사용하여 읽기 쉬운 대시보드를 생성합니다.

## 에러 처리

### 프로젝트 미초기화
```markdown
❌ ERROR: MoAI 프로젝트가 초기화되지 않았습니다.

필수 구조 누락:
- .moai/ 디렉토리
- .claude/commands/moai/ 디렉토리
- .claude/agents/moai/ 디렉토리

해결 방법:
먼저 프로젝트를 초기화해주세요:
> /moai:1-project init
```

### 상태 파일 손상
```markdown
⚠️ WARNING: 상태 파일이 손상되었습니다.

손상된 파일:
- .moai/indexes/state.json (JSON 구문 오류)
- .moai/indexes/tags.json (파일 없음)

복구 방법:
> /moai:6-sync force --rebuild-index
```

### Git 저장소 없음
```markdown
📁 INFO: Git 저장소가 감지되지 않았습니다.

Git 정보 없이 제한된 대시보드가 표시됩니다:
- 브랜치 정보: 없음
- 커밋 히스토리: 없음
- 변경사항: 감지 불가

권장사항:
> git init && git add . && git commit -m "Initial commit"
```

## 성능 최적화

### 캐시 활용
```
🚀 성능 최적화 기능:

1. 상태 파일 캐시: 5초간 메모리 보관
2. Git 정보 캐시: 변경 감지 시에만 갱신
3. 레이아웃 캐시: 터미널 크기별 저장
4. 컬러 캐시: 테마별 미리 생성
```

### 지연 로딩
```python
# 기본 정보만 즉시 로드
dashboard.load_essential()

# 상세 정보는 --detail 옵션 시에만 로드
if args.detail:
    dashboard.load_detailed_metrics()
    dashboard.load_performance_data()
```

## 사용자 정의

### 테마 설정
```bash
# 다크 테마 (기본값)
/moai:7-dashboard --theme=dark

# 라이트 테마
/moai:7-dashboard --theme=light

# 컬러 없음 (CI/CD용)
/moai:7-dashboard --theme=plain
```

### 섹션 선택
```bash
# 특정 섹션만 표시
/moai:7-dashboard --sections=pipeline,specs,git

# 전체 섹션 (기본값)
/moai:7-dashboard --sections=all
```

이 명령어를 통해 MoAI-ADK 프로젝트의 진행 상황을 실시간으로 모니터링하고, 데이터 기반의 의사결정을 내릴 수 있습니다.

## 🔁 실행 프로세스 (필수)

**1단계: 데이터 수집**
데이터 수집 스크립트를 실행하여 현재 프로젝트 상태를 JSON으로 수집합니다.

**2단계: 데이터 파싱**
수집된 JSON 데이터를 파싱하여 대시보드 구성 요소를 준비합니다.

**3단계: 텍스트 렌더링**
파싱된 데이터를 기반으로 구조화된 대시보드를 텍스트로 렌더링합니다.

**4단계: 결과 출력**
Claude Code 화면에 최종 대시보드를 표시합니다.

### 실행 단계별 세부사항

#### 1단계: 데이터 수집
```bash
# 기본 모드
python3 .moai/scripts/collect_dashboard_data.py

# 상세 모드
python3 .moai/scripts/collect_dashboard_data.py --detail
```

#### 2-4단계: 파싱 및 렌더링
수집된 JSON 데이터에서 다음 섹션들을 순서대로 렌더링:
1. **헤더**: 프로젝트명, 버전, 브랜치, 현재 시간
2. **파이프라인**: SPECIFY → PLAN → TASKS → IMPLEMENT → SYNC 진행률
3. **SPEC 현황**: 각 SPEC별 상태 테이블
4. **TAG 시스템**: 카테고리별 태그 수 및 건강도
5. **Constitution**: 5원칙 준수 현황
6. **Git 상태**: 브랜치, 커밋, 변경사항
7. **추천 액션**: 다음 단계 추천사항
8. **경고사항**: 주의가 필요한 항목들

**중요**: 대시보드는 읽기 전용이며 프로젝트 상태를 변경하지 않습니다.

---

### 실제 실행 시작 👇

현재 프로젝트 상태를 수집하고 대시보드를 표시하겠습니다.

먼저 데이터 수집 스크립트를 실행하여 현재 상태를 가져오겠습니다:

```bash
python3 .moai/scripts/collect_dashboard_data.py --format compact
```

수집된 데이터를 기반으로 대시보드를 렌더링합니다:

## 📊 MoAI-ADK 프로젝트 대시보드

데이터를 수집하고 파싱하여 다음과 같은 구조로 대시보드를 표시합니다:

### 헤더 섹션
```
╭─────────────────────── 🗿 MoAI-ADK Dashboard ───────────────────────╮
│ 프로젝트: {project_name} v{version} | 브랜치: {branch} | {timestamp} │
╰────────────────────────────────────────────────────────────────────╯
```

### 파이프라인 진행률 바
각 단계별 진행 상황을 프로그레스 바로 표시:
- SPECIFY: ████████████████████ 100% ✅
- PLAN: ████████████████████ 100% ✅
- TASKS: ████████░░░░░░░░░░░░ 40% 🔄
- IMPLEMENT: ████████████████████ 100% ✅
- SYNC: ░░░░░░░░░░░░░░░░░░░░ 0% ⏳

### SPEC 현황 테이블
| SPEC ID | 제목 | 상태 | 우선순위 | 진행률 |
|---------|------|------|----------|--------|
| SPEC-001 | 프로젝트 초기화 | ✅ 완료 | P1 | 100% |
| SPEC-002 | 에이전트 최적화 | ✅ 완료 | P2 | 100% |
| SPEC-004 | 대시보드 기능 | 🔄 진행 | P4 | 60% |

### TAG 시스템 건강도
```
🏷️ TAG 시스템 (건강도: {health_score}% ✅)
├── SPEC: {spec_count}개
├── STEERING: {steering_count}개
├── IMPLEMENTATION: {impl_count}개
├── QUALITY: {quality_count}개
└── 전체 연결성: {traceability_status}
```

### Constitution 5원칙 준수 현황
```
⚖️ Constitution 5원칙 준수 현황
✅ Simplicity    {simplicity_status}  {simplicity_details}
✅ Architecture  {architecture_status}  {architecture_details}
⚠️ Testing       {testing_status}  {testing_details}
✅ Observability {observability_status}  {observability_details}
✅ Versioning    {versioning_status}  {versioning_details}
```

### Git 상태 정보
```
🔀 Git 상태
├── 현재 브랜치: {branch}
├── 최근 커밋: {last_commit}
├── 변경사항: 수정 {modified}개, 삭제 {deleted}개, 미추적 {untracked}개
└── 작업 상태: {work_status}
```

### 추천 액션
```
💡 다음 단계 추천
{recommendations_list}

⚠️ 주의사항
{warnings_list}
```

---

**실행 프로세스:**
1. 데이터 수집 스크립트 실행
2. JSON 응답 파싱
3. 각 섹션별 데이터 추출
4. 위 템플릿에 데이터 대입하여 최종 대시보드 렌더링