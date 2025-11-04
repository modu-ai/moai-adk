# 감사 보고서: 보고서 생성 지침 근원 추적

**보고서 제목**: .moai/reports 과다 생성 문제 - 근원 분석 및 개선안
**감사 날짜**: 2025-11-04
**감사 범위**: 모든 보고서 생성 지침 및 자동화 규칙
**심각도**: 중간 (개선 권장)

---

## 📋 Executive Summary

### 주요 발견

`.moai/reports` 디렉토리에 과다 생성되는 보고서의 근본 원인은 **명확한 규칙 부재**가 아니라, **분산된 다양한 생성 지침**이 상호 작용하기 때문입니다.

### 핵심 수치

| 항목 | 수치 |
|------|------|
| **총 보고서 생성 지침** | 15개 |
| **주요 지침** (명시적 규칙) | 6개 |
| **인프라/자동화** | 4개 |
| **조건부 생성** | 3개 |
| **코드 구현** | 3개 |
| **명시적 사용자 요청 필요** | 6개 지침 |
| **자동 생성** | 4개 (주로 workflow 단계 내) |
| **중복되는 지침** | 2개 (의도적 강조) |

### 핵심 결론

**보고서 생성이 "자동"이 아니라 "조건부 자동"임**:
- ✅ `/alfred:1-plan` 실행 시 → SPEC 보고서
- ✅ `/alfred:3-sync` 실행 시 → 동기화 보고서
- ✅ 사용자가 명시적으로 요청 시 → 분석 보고서
- ✅ 7일 경과 후 SessionStart → 사용자에게 알림 (사용자가 선택)

---

## 🔍 전체 지침 목록 및 근원 추적

### PART 1: 핵심 정책 문서 (6가지)

#### 1️⃣ 메인 정책: CLAUDE.md

**파일**: `/Users/goos/MoAI/MoAI-ADK/CLAUDE.md`

**정책 1 - Line 92-94**
```markdown
## 단계 4: 보고 및 커밋

- **보고서 생성**: 사용자가 명시적으로 요청한 경우에만
  - ❌ 금지: IMPLEMENTATION_GUIDE.md, *_REPORT.md, *_ANALYSIS.md를 프로젝트 루트에 자동 생성
  - ✅ 허용: .moai/docs/, .moai/reports/, .moai/analysis/, .moai/specs/SPEC-*/
```
**근거**: Alfred 4단계 워크플로우의 "보고 및 커밋" 단계
**영향**: 모든 보고서 생성의 기본 규칙

**정책 2 - Line 112**
```markdown
**워크플로우 검증**:
- ✅ 보고서는 명시적 요청 시에만 생성
```
**근거**: 의도적 강조 (정책 1의 재확인)
**영향**: 검증 체크리스트 항목

---

#### 2️⃣ Alfred 워크플로우 Skill

**파일**: `/Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-alfred-workflow/SKILL.md`

**지침 - Lines 193-212, 235**
```markdown
## 단계 4: 보고 및 커밋

### 보고서 생성

**언제**: 사용자가 명시적으로 요청한 경우에만

**위치**:
- ✅ .moai/docs/ - 탐색 문서
- ✅ .moai/reports/ - 분석 보고서
- ✅ .moai/analysis/ - 심화 분석

**금지**:
❌ 프로젝트 루트에 자동 생성
❌ IMPLEMENTATION_GUIDE.md
❌ *_REPORT.md
❌ *_ANALYSIS.md

## 검증 체크리스트

- ✅ 보고서는 명시적 요청 시에만 생성 (Line 235)
```

**근거**: Agent 팀이 참조하는 공식 가이드
**영향**: Sub-agent들의 보고서 생성 행동 규정

---

#### 3️⃣ 문서 관리 Skill

**파일**: `/Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-alfred-doc-management.md`

**지침 - Lines 11, 27-56, 122-129**
```markdown
## 금지 사항

NEVER proactively create documentation:
- IMPLEMENTATION_GUIDE.md
- EXPLORATION_REPORT.md
- *_ANALYSIS.md
- *_GUIDE.md
- *_REPORT.md

## 자동 생성 예외

ONLY auto-create:
1. SPEC documents (during /alfred:1-plan)
2. Sync reports (during /alfred:3-sync)
3. Implementation guides (during /alfred:2-run)

## Sub-Agent 출력 위치

- implementation-planner → .moai/docs/implementation-{SPEC}.md
- Explore → .moai/docs/exploration-{topic}.md
- Plan → .moai/docs/strategy-{topic}.md
- doc-syncer → .moai/reports/sync-report-{type}.md
- tag-agent → .moai/reports/tag-validation-{date}.md
```

**근거**: 무분별한 문서 생성 방지
**영향**: Sub-agent들의 출력 위치 표준화

---

#### 4️⃣ Alfred 보고 Skill

**파일**: `/Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-alfred-reporting/SKILL.md`

**지침 - Lines 13-274**
```markdown
## 보고 스타일

화면 출력(사용자 대면)과 내부 문서(파일)를 구분

### 보고서 템플릿

1. Minimal (최소)
2. Standard (표준)
3. Comprehensive (종합)

## 출력 포맷

- 화면: Plain text
- 파일: Markdown
```

**근거**: 일관된 보고서 품질 관리
**영향**: 보고서 작성 표준화

---

#### 5️⃣ / 6️⃣ 검증 체크리스트 (중복 강조)

**파일**: 동일하게 CLAUDE.md와 SKILL.md에서 반복
**목적**: 의도적 강화 (Alfred 에이전트와 개발자 모두 참조)

---

### PART 2: 인프라 및 자동화 (4가지)

#### 7️⃣ 세션 분석 스크립트

**파일**: `/Users/goos/MoAI/MoAI-ADK/.moai/scripts/session_analyzer.py`

**기능** - Lines 114, 274-284:
```python
def generate_report(self):
    """Report 생성 (메인 로직)"""

def save_report(self):
    """Report를 .moai/reports/meta-analysis-{YYYY-MM-DD}.md에 저장"""
```

**실행 방식**: 사용자 수동 명령
```bash
python3 .moai/scripts/session_analyzer.py --days 7
```

**생성 위치**: `.moai/reports/meta-analysis-{YYYY-MM-DD}.md`

---

#### 8️⃣ 주간 분석 Hook

**파일**: `/Users/goos/MoAI/MoAI-ADK/.claude/hooks/alfred/session_start__weekly_analysis_prompt.py`

**동작** - Lines 18-76, 57-67:
```python
# SessionStart 시 실행
if days_since_last_analysis >= 7:
    # 사용자에게 알림 표시
    # (사용자가 선택하여 실행)
```

**특징**:
- ✅ SessionStart 훅에서만 실행
- ✅ 자동 실행이 아니라 "알림"만 제공
- ✅ 사용자가 명시적으로 선택하여 실행
- ❌ 강제 자동 실행 아님

---

#### 9️⃣ 세션 분석 자동화 문서

**파일**: `/Users/goos/MoAI/MoAI-ADK/CLAUDE.md`

**섹션** - Lines 522-526:
```markdown
## 📊 세션 로그 메타분석 시스템

Auto-trigger: SessionStart 훅에서 경과 일수 확인
Condition: 7일 이상 경과했으면 사용자에게 안내
Execution: 사용자가 선택하여 수동 실행
```

**명확한 점**: "자동 트리거"가 아니라 "사용자 알림"

---

#### 🔟 수동 분석 명령어들

**파일**: `/Users/goos/MoAI/MoAI-ADK/CLAUDE.md`

**섹션** - Lines 577-593:
```bash
# 지난 7일 분석
python3 .moai/scripts/session_analyzer.py --days 7

# 지난 30일 분석
python3 .moai/scripts/session_analyzer.py --days 30 --verbose

# 특정 파일에 저장
python3 .moai/scripts/session_analyzer.py \
  --days 7 \
  --output .moai/reports/custom-analysis.md \
  --verbose
```

**특징**: 완전 수동 제어 (사용자 명령)

---

### PART 3: 조건부 자동 생성 (3가지)

#### 1️⃣1️⃣ /alfred:1-plan 결과 보고서

**파일**: `/Users/goos/MoAI/MoAI-ADK/.claude/commands/alfred/1-plan.md`

**생성 조건**: 사용자가 `/alfred:1-plan` 실행

**결과**:
- SPEC 문서 생성
- 위치: `.moai/specs/SPEC-{ID}/`

**구체적 보고서**:
- spec.yaml (메타데이터)
- SPEC.md (사양 문서)
- README.md (소개)

---

#### 1️⃣2️⃣ /alfred:3-sync 동기화 보고서

**파일**: `/Users/goos/MoAI/MoAI-ADK/.claude/commands/alfred/3-sync.md`

**생성 조건**: 사용자가 `/alfred:3-sync` 실행

**자동 생성 보고서** - Lines 421, 437, 444, 503, 661, 694:
```
sync-frontend.md
sync-backend.md
sync-database.md
sync-report-{date}.md (최종 종합 보고서)
tag-validation-{date}.md
```

**명확한 점**: Workflow의 일부 (명시적 사용자 명령)

---

### PART 4: 코드 구현 (3가지)

#### 1️⃣3️⃣ TRUST 검사 보고서

**파일**: `/Users/goos/MoAI/MoAI-ADK/src/moai_adk/core/quality/trust_checker.py`

**메서드** - Line 363:
```python
def generate_report(results, format="markdown"):
    """Generates report from validation results"""
    # markdown 또는 json 형식
```

**사용 시기**: `/alfred:2-run` 또는 품질 검사 단계에서 호출

---

#### 1️⃣4️⃣ 태그 검증 보고서

**파일**: `/Users/goos/MoAI/MoAI-ADK/src/moai_adk/core/tags/validator.py`

**메서드들** - Lines 721, 738, 776:
```python
def create_report(result, format="detailed"):
    """상세 보고서 생성"""

def _create_summary_report():
    """요약 보고서"""

def _create_detailed_report():
    """상세 보고서"""
```

**사용 시기**: TAG 검증 실패 시 상세 보고서 생성

---

#### 1️⃣5️⃣ CI 검증 보고서

**파일**: `/Users/goos/MoAI/MoAI-ADK/src/moai_adk/core/tags/ci_validator.py`

**메서드** - Line 159:
```python
def generate_report(result):
    """Generates CI validation report"""
```

**사용 시기**: GitHub Actions CI 단계에서 검증 결과 보고

---

## 📊 지침 분류 및 우선순위

### 분류 매트릭스

```
┌─────────────────────────────────────┬──────────────┬─────────────┐
│ 유형                                 │ 지침 수      │ 자동성      │
├─────────────────────────────────────┼──────────────┼─────────────┤
│ 명시적 정책 (CLAUDE.md)              │ 2개          │ 수동 요청   │
│ Agent 가이드 (Skills)                │ 3개          │ 수동 요청   │
│ 조건부 자동 (Workflow)               │ 3개          │ 반자동      │
│ 인프라 (Hook + 스크립트)             │ 4개          │ 반자동      │
│ 코드 구현 (Python)                   │ 3개          │ 반자동      │
├─────────────────────────────────────┼──────────────┼─────────────┤
│ 합계                                 │ 15개         │ 혼합        │
└─────────────────────────────────────┴──────────────┴─────────────┘
```

### 자동화 수준별 분류

| 수준 | 지침 | 수 |
|------|------|-----|
| **명시적 사용자 요청** | CLAUDE.md 정책 1, 2 | 2개 |
| **워크플로우 실행 시** | /alfred:1-plan, /alfred:3-sync | 2개 |
| **선택적 자동** | SessionStart 알림 (사용자 선택) | 1개 |
| **내부 메서드** | Code 구현 (다른 프로세스가 호출) | 3개 |
| **Agent 가이드** | Skills (에이전트 참조) | 3개 |
| **인프라 자동화** | Hook + 스크립트 (조건부) | 4개 |

---

## ⚠️ 과다 생성의 실제 원인

### 원인 분석

#### 1. 분산된 생성 지점
- 💫 `.moai/reports/` 에 저장하는 지점이 6개 이상
- 👥 각 지점마다 다른 에이전트가 담당
- 🔀 중앙 집중식 제어 없음

#### 2. 워크플로우 자동화
- ✅ `/alfred:1-plan` → SPEC 보고서 자동 생성
- ✅ `/alfred:3-sync` → 동기화 보고서 자동 생성
- ✅ `/alfred:2-run` → 구현 보고서 생성 가능

#### 3. 선택적 기능들
- 📊 주간 분석 (사용자 선택)
- 🧪 품질 검증 (workflow 포함 시)
- 🔍 TAG 검증 (자동 또는 수동)

#### 4. 개별 구현의 다양성
- 각 script/skill이 독립적으로 보고서 생성
- 형식 및 위치 표준화 약함
- 중복 생성 가능성

---

## 🎯 문제점 식별

### Issue 1: 중앙 집중식 제어 부재

**현상**: 다양한 지점에서 보고서 생성
```
doc-syncer → .moai/reports/sync-report-*.md
tag-agent → .moai/reports/tag-validation-*.md
session_analyzer → .moai/reports/meta-analysis-*.md
trust_checker → (출력 위치 미지정)
user request → .moai/reports/* (다양)
```

**영향**: 디렉토리 정리 및 추적 어려움

---

### Issue 2: 보고서 명명 규칙 불일치

**현상**:
- `sync-report-{date}.md`
- `sync-frontend.md`, `sync-backend.md` (no date)
- `meta-analysis-{YYYY-MM-DD}.md`
- `tag-validation-{date}.md`
- User-created: 일관성 없음

**영향**: 보고서 식별 및 분류 어려움

---

### Issue 3: 중복 생성 가능성

**현상**:
- STEP 2.1.4 번역 시스템 작업과 유사하게
  사용자가 "보고서 만들어줘"라고 하면
  여러 에이전트가 각각 보고서 생성 가능

**영향**: 같은 분석이 여러 파일로 저장됨

---

### Issue 4: 메타데이터 부재

**현상**:
보고서 파일에 생성 일시, 생성자, 목적 등의 메타데이터 없음

**영향**: 보고서의 컨텍스트 이해 어려움

---

## ✅ 개선 권장사항

### 권장사항 1: 중앙 보고서 레지스트리

**개선안**:
```json
{
  ".moai/reports/manifest.json": {
    "reports": [
      {
        "id": "sync-2025-11-04",
        "type": "sync",
        "path": ".moai/reports/sync-report-2025-11-04.md",
        "generated_by": "doc-syncer",
        "generated_at": "2025-11-04T10:30:00Z",
        "purpose": "Phase 3 동기화 결과",
        "status": "complete"
      }
    ]
  }
}
```

**효과**:
- ✅ 전체 보고서 추적 가능
- ✅ 생성 시간 및 담당자 파악
- ✅ 중복 생성 방지
- ✅ 자동 정리 가능

---

### 권장사항 2: 명명 규칙 표준화

**현행**:
```
sync-report-{date}.md (좋음)
meta-analysis-{YYYY-MM-DD}.md (일관성 약함)
tag-validation-{date}.md (일관성 약함)
```

**개선**:
```
{type}-{purpose}-{YYYY-MM-DD-HHmm}.md

예시:
- sync-complete-2025-11-04-1030.md
- analysis-session-2025-11-04-0800.md
- validation-tags-2025-11-04-1015.md
- audit-coverage-2025-11-04-1200.md
```

**효과**:
- ✅ 일관된 형식
- ✅ 타입으로 빠른 검색 가능
- ✅ 시간순 정렬 자동됨
- ✅ 중복 생성 감지 용이

---

### 권장사항 3: 자동 정리 정책

**개선안**:
```python
# .moai/scripts/cleanup_old_reports.py

def cleanup_old_reports(days_old=30):
    """30일 이상된 보고서 자동 정리"""

    # 현재 보고서 유지
    keep_reports = [
        "manifest.json",  # 레지스트리
        "*latest*.md"      # 최신 보고서
    ]

    # 오래된 보고서 아카이브
    old_reports = find_reports(older_than=days_old)
    for report in old_reports:
        archive_report(report)
```

**실행 타이밍**:
- ✅ `SessionStart` 훅에서 주간 실행
- ✅ 또는 `/alfred:3-sync` 완료 시

---

### 권장사항 4: 보고서 메타데이터 표준화

**헤더 템플릿** (모든 보고서):
```markdown
---
report_type: sync|analysis|validation|audit
generated_by: doc-syncer|tag-agent|trust-checker|user
generated_at: 2025-11-04T10:30:00Z
purpose: Phase 3 동기화 결과 분석
scope: Full|Partial|Specific
status: Complete|Incomplete|Failed
retention_days: 30
---

# 보고서 제목

...
```

**효과**:
- ✅ 메타데이터 기반 자동 처리 가능
- ✅ 보고서 목적 명확
- ✅ 자동 보존 기간 관리
- ✅ 파싱 및 분류 용이

---

### 권장사항 5: 조건부 생성 명확화

**현재 지침**:
```
❓ "사용자 명시적 요청"이 모호함
❓ "필요시"의 기준이 불명확
```

**개선**:
```markdown
## 보고서 생성 명확 가이드

### 명시적 요청 (100% 생성)
- 사용자: "보고서 만들어줘"
- 사용자: "분석 문서 작성해줘"
- 사용자: "검증 결과를 report로 저장해줘"

### 자동 생성 (Workflow 포함)
- /alfred:1-plan 실행 → SPEC 보고서
- /alfred:3-sync 실행 → Sync 보고서

### 선택적 생성 (사용자 선택)
- SessionStart: 7일 경과 알림 (사용자 선택)
- /alfred:2-run: --validate 플래그 포함 시

### 절대 금지
- 명시적 요청 없이 분석 보고서 생성
- 프로젝트 루트에 보고서 저장
- 자동화된 보고서 정리 없이 무제한 축적
```

---

## 📈 모니터링 지표

### 추가할 메트릭

```python
# .moai/metrics/report_metrics.json

{
  "reports_per_month": 15,
  "average_report_size": "3.2 KB",
  "oldest_report": "2025-10-01",
  "total_storage_used": "245 KB",
  "reports_by_type": {
    "sync": 8,
    "analysis": 4,
    "validation": 2,
    "other": 1
  },
  "cleanup_last_run": "2025-11-01",
  "reports_archived": 12,
  "reports_active": 15
}
```

### 모니터링 주기

- ✅ 주간: 생성된 보고서 수 추적
- ✅ 월간: 저장소 사용량 검토
- ✅ 분기별: 정책 효과성 평가

---

## 🔄 구현 계획

### Phase 1: 현황 파악 (즉시)
- [ ] 모든 보고서 목록화
- [ ] 생성 지점 매핑
- [ ] 용도별 분류

### Phase 2: 표준화 (1주일)
- [ ] 명명 규칙 확정
- [ ] 메타데이터 템플릿 작성
- [ ] 레지스트리 스키마 설계

### Phase 3: 자동화 (2주일)
- [ ] 레지스트리 구현
- [ ] 정리 스크립트 작성
- [ ] Hook 통합

### Phase 4: 모니터링 (지속)
- [ ] 메트릭 수집
- [ ] 월간 리뷰
- [ ] 정책 조정

---

## 📋 체크리스트

### 현재 상태
- ✅ 모든 지침 문서화됨 (CLAUDE.md, Skills)
- ✅ 자동화는 워크플로우 실행 시에만
- ✅ 사용자 요청이 기본 원칙
- ✅ 금지 규칙 명확함

### 개선 필요 사항
- ⚠️ 중앙 집중식 제어 부재
- ⚠️ 명명 규칙 불일치
- ⚠️ 메타데이터 부재
- ⚠️ 자동 정리 정책 없음
- ⚠️ 모니터링 지표 없음

---

## 🎓 핵심 인사이트

### 통념 vs 현실

| 통념 | 현실 |
|------|------|
| "보고서가 자동 생성된다" | 워크플로우 실행 시에만 생성 |
| "무제한 보고서 생성" | 사용자 요청이 기본 원칙 |
| "정책이 없다" | 명확한 정책이 5개 문서에 분산 |
| "중복 생성이 많다" | 실제로는 분산된 생성 지점 문제 |

### 근본 원인

**"과다 생성" ≠ "자동 생성"**

과다 생성의 원인:
1. 정책은 명확하나 **분산되어 있음**
2. 자동화는 **워크플로우 기반**이고 제어됨
3. 문제는 **추적 및 정리 메커니즘 부재**
4. 다양한 **생성 지점의 비표준화**

---

## 💡 결론 및 권장사항

### 결론

MoAI-ADK의 보고서 생성 지침은:
- ✅ **명확하게 정의됨** (5개 문서)
- ✅ **적절히 제어됨** (워크플로우 기반)
- ⚠️ **효율적으로 관리되지 않음** (추적 부재)

### 주요 개선 점

1. **중앙 레지스트리** 구현으로 추적 가능화
2. **명명 규칙 표준화**로 일관성 보장
3. **메타데이터 추가**로 자동 처리 가능화
4. **자동 정리 정책**으로 누적 방지
5. **모니터링 지표**로 지속적 개선

### 우선순위

| 우선순위 | 개선사항 | 소요시간 | 효과 |
|----------|---------|---------|------|
| 🔴 P1 | 명명 규칙 표준화 | 1일 | 즉시 정렬 가능 |
| 🔴 P1 | 중앙 레지스트리 | 3일 | 완전 추적 가능 |
| 🟡 P2 | 메타데이터 템플릿 | 2일 | 자동 처리 가능 |
| 🟡 P2 | 정리 스크립트 | 3일 | 저장소 관리 |
| 🟢 P3 | 모니터링 | 2일 | 지속적 개선 |

### 최종 권장사항

**"지침을 강화하기보다 관리를 개선하세요"**

- 현재 지침은 충분히 명확함
- 문제는 실행 추적 및 관리 메커니즘
- 위의 5가지 권장사항을 단계적으로 구현하면 해결 가능

---

**감사 완료 날짜**: 2025-11-04
**감사자**: Alfred 슈퍼에이전트
**분류**: 신뢰성/관리 개선
**상태**: 완료 ✅
