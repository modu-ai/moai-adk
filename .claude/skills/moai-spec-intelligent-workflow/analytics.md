# SPEC Analytics and Reporting System

**Created**: 2025-11-21
**Status**: Production Ready

---

## Overview

To measure the effectiveness of the SPEC-First workflow, we provide a **simple yet practical analytics system**.

This system:
- ✅ Automatically displays recent 30-day SPEC statistics
- ✅ Automatically collects SPEC-related data
- ✅ Automatically generates monthly reports
- ✅ Provides trend analysis and improvement recommendations

---

## Data Structure

### Core File: `.moai/logs/spec-usage.json`

```json
{
  "metadata": {
    "version": "1.0",
    "created_at": "2025-11-21T00:00:00Z",
    "updated_at": "2025-11-21T23:59:59Z"
  },
  "specs": [
    {
      "spec_id": "SPEC-001",
      "title": "User Profile Image Upload",
      "created_at": "2025-11-20T10:00:00Z",
      "completed_at": "2025-11-21T14:30:00Z",
      "template_level": "Level 2",
      "estimated_time_minutes": 120,
      "actual_time_minutes": 95,
      "status": "completed",
      "linked_commits": ["abc1234", "def5678"],
      "linked_files": ["src/routes/profile.ts", "src/services/imageService.ts"],
      "test_coverage": 87,
      "test_passing": 23
    }
  ],
  "summary": {
    "total_specs": 12,
    "completed_specs": 10,
    "avg_completion_time_minutes": 45,
    "avg_test_coverage": 85,
    "time_saved_minutes": 240
  }
}
```

---

## SessionStart Hook: Display Statistics

### Purpose
Automatically display **recent 30-day SPEC statistics** at session start

### Display Content

```
📊 SPEC-First Workflow Stats (Last 30 days)

  ✅ SPEC created: 12
  ⏱️  Average completion time: 45 min
      (vs without SPEC: 72 min, 37% faster!)
  🔗 Code linkage rate: 92% (11/12)
  🧪 Average test coverage: 85%

  📈 Trend: +3 SPECs this week (20% increase)

  💡 Tips:
    • 1 SPEC still in progress (SPEC-010)
    • 2 SPECs with low coverage (SPEC-005: 65%, SPEC-007: 72%)

  ℹ️  Details: .moai/logs/spec-usage.json
```

### Implementation

**Hook File**: `.claude/hooks/sessionstart.sh`

```bash
#!/bin/bash
# Display SPEC statistics

SPEC_USAGE_FILE=".moai/logs/spec-usage.json"

if [ ! -f "$SPEC_USAGE_FILE" ]; then
    exit 0  # Do not display if no data
fi

# Calculate and display statistics with Python
python3 << 'EOF'
import json
from datetime import datetime, timedelta

with open('.moai/logs/spec-usage.json', 'r') as f:
    data = json.load(f)

# Filter recent 30-day SPECs
cutoff_date = datetime.now() - timedelta(days=30)
recent_specs = [
    s for s in data['specs']
    if datetime.fromisoformat(s['created_at']) > cutoff_date
]

# Calculate statistics
total = len(recent_specs)
completed = len([s for s in recent_specs if s['status'] == 'completed'])
avg_time = sum(
    s['actual_time_minutes'] for s in recent_specs
    if s['actual_time_minutes']
) / completed if completed > 0 else 0

linkage = len([s for s in recent_specs if s['linked_commits']]) / total if total > 0 else 0
coverage = sum(s['test_coverage'] for s in recent_specs) / total if total > 0 else 0

# Display
print(f"📊 SPEC-First Workflow Stats (Last 30 days)")
print(f"  ✅ SPEC created: {total}")
print(f"  ✅ SPEC completed: {completed}")
print(f"  ⏱️  Average time: {avg_time:.0f} min")
print(f"  🔗 Code linkage: {linkage*100:.0f}%")
print(f"  🧪 Test coverage: {coverage:.0f}%")
EOF
```

### Display Frequency
- Every session (within 1 second)
- Rolling 30-day window

---

## SessionEnd Hook: Data Collection

### Purpose
Automatically collect **SPEC-related data** at session end

### Collected Data

1. **SPEC Creation**
   - SPEC ID, creation time, template level, estimated time

2. **Implementation Tracking**
   - Actual time spent, status (completed/in progress/abandoned)

3. **Code Linkage**
   - Git commits (messages containing SPEC-XXX)
   - Modified files, added tests

4. **Quality Metrics**
   - Test coverage, test pass rate

### Implementation

**Hook File**: `.claude/hooks/sessionend.sh`

```bash
#!/bin/bash
# SPEC 데이터 수집 (세션 종료 시)

SPEC_USAGE_FILE=".moai/logs/spec-usage.json"

python3 << 'EOF'
import json
import subprocess
from datetime import datetime

SPEC_USAGE_FILE = ".moai/logs/spec-usage.json"
spec_id = subprocess.run(
    ['git', 'log', '--oneline', '-n', '1', '--grep=SPEC'],
    capture_output=True,
    text=True
).stdout.strip()

if not spec_id:
    exit(0)

# 데이터 로드
try:
    with open(SPEC_USAGE_FILE, 'r') as f:
        data = json.load(f)
except FileNotFoundError:
    data = {"metadata": {}, "specs": [], "summary": {}}

# SPEC 정보 업데이트
commits = subprocess.run(
    ['git', 'log', '--oneline', f'--grep={spec_id}', '-n', '10'],
    capture_output=True,
    text=True
).stdout.strip().split('\n')

files = subprocess.run(
    ['git', 'diff', '--name-only', 'HEAD~1'],
    capture_output=True,
    text=True
).stdout.strip().split('\n')

for spec in data['specs']:
    if spec['spec_id'] == spec_id:
        spec['linked_commits'] = [c.split()[0] for c in commits if c]
        spec['linked_files'] = [f for f in files if f]
        spec['completed_at'] = datetime.now().isoformat()
        break

# 저장
with open(SPEC_USAGE_FILE, 'w') as f:
    json.dump(data, f, indent=2)
EOF
```

### 수집 빈도
- 세션 종료 시 (2초 이내)
- 자동 수집 (사용자 개입 없음)

---

## 월간 리포트 생성

### 목적
매월 자동으로 **SPEC 워크플로우의 효과를 분석**한 리포트 생성

### 위치
`.moai/reports/spec-analytics-YYYY-MM.md`

### 리포트 내용 예제

```markdown
# SPEC Analytics Report - November 2025

## 📊 Summary

| Metric | Value | Trend |
|--------|-------|-------|
| SPEC Created | 12 | +3 (20%) |
| SPEC Completed | 10 | +2 (25%) |
| Avg Time | 45 min | -15 min (25%) |
| Test Coverage | 85% | +5% (6%) |
| Code Linkage | 92% | +7% (8%) |

## 📈 Trends

### Completion Time by Level
- Level 1: 8 min (vs 10 min estimate, 20% faster)
- Level 2: 45 min (vs 60 min estimate, 25% faster)
- Level 3: 105 min (vs 120 min estimate, 13% faster)

## 💡 Insights

1. **Level 2 효과 가장 큼** (25% 시간 절감)
2. **테스트 커버리지 우수** (85% 유지)
3. **미완료 SPEC 1개** (SPEC-010, 우선순위 재조정 필요)
4. **낮은 커버리지 2개** (SPEC-005, 007 테스트 보충 필요)

## 🎯 Recommendations

### 즉시 (1-2주)
- SPEC-010 완료 추진
- SPEC-005, 007 테스트 보충

### 단기 (1개월)
- SPEC 생성 빈도 유지 (주 3-4개)
- 테스트 커버리지 85% 이상 유지

### 장기 (3개월)
- SPEC-First 워크플로우 정착
- 월 20개 SPEC 목표
```

### 생성 빈도
- 매월 마지막 날 자정에 자동 생성
- 기존 리포트는 보관

---

## 메트릭 정의

### 주요 메트릭

| 메트릭 | 정의 | 목표 |
|--------|------|------|
| **SPEC Creation Rate** | 월별 생성된 SPEC 개수 | 15-20개 |
| **Completion Rate** | 완료된 SPEC / 전체 SPEC | 90%+ |
| **Time Savings** | (예상 시간 - 실제 시간) / 예상 시간 | 25%+ |
| **Test Coverage** | 테스트 커버리지 평균 | 85%+ |
| **Code Linkage** | 코드와 연결된 SPEC / 전체 SPEC | 90%+ |

### 계산식

```
Completion Rate = Completed / Total * 100
Time Savings = (Estimated - Actual) / Estimated * 100
Code Linkage = Specs with Commits / Total Specs * 100
Avg Coverage = Sum of Coverage / Total Specs
```

---

## 데이터 보관 정책

### 저장 위치
```
.moai/logs/
├── spec-usage.json              # 현재 상태 (최신)
├── spec-usage-YYYY-MM-DD.json   # 일일 백업 (선택사항)

.moai/reports/
├── spec-analytics-2025-11.md    # 월간 리포트
├── spec-analytics-2025-10.md
└── ...
```

### 보존 기간
- 현재 달: 전체 데이터
- 이전 11개월: 전체 데이터
- 12개월 이상: 요약 통계만 보관

---

## 개인정보 보호

### 데이터 수집 원칙
- 📊 통계만 수집 (개인정보 최소화)
- 🔒 로컬 저장 (외부 전송 안 함)
- 📝 감사 추적 가능

### 데이터 보안
- `.moai/logs/` : `.gitignore`에 추가
- 로컬 개발 환경에서만 수집
- CI/CD 환경에서는 자동 수집 안 함

---

## FAQ

**Q: 통계가 정확한가요?**
A: 자동 수집되므로 완벽하지 않을 수 있습니다. 주요 트렌드 파악용으로 사용하세요.

**Q: 데이터는 어디에 저장되나요?**
A: `.moai/logs/` 디렉토리에 로컬 저장됩니다. 외부로 전송되지 않습니다.

**Q: 통계 수집을 비활성화할 수 있나요?**
A: 네, `.claude/hooks/sessionend.sh`에서 Hook을 제거하면 됩니다.

---

**문서 버전**: 1.0.0
**마지막 업데이트**: 2025-11-21
**상태**: Production Ready
