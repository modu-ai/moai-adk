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
# Collect SPEC data (at session end)

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

# Load data
try:
    with open(SPEC_USAGE_FILE, 'r') as f:
        data = json.load(f)
except FileNotFoundError:
    data = {"metadata": {}, "specs": [], "summary": {}}

# Update SPEC information
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

# Save
with open(SPEC_USAGE_FILE, 'w') as f:
    json.dump(data, f, indent=2)
EOF
```

### Collection Frequency
- At session end (within 2 seconds)
- Automatic collection (no user intervention)

---

## Monthly Report Generation

### Purpose
Automatically generate monthly reports **analyzing the effectiveness of the SPEC workflow**

### Location
`.moai/reports/spec-analytics-YYYY-MM.md`

### Report Content Example

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

1. **Level 2 most effective** (25% time savings)
2. **Excellent test coverage** (maintained at 85%)
3. **1 incomplete SPEC** (SPEC-010, requires priority adjustment)
4. **2 with low coverage** (SPEC-005, 007 need test supplementation)

## 🎯 Recommendations

### Immediate (1-2 weeks)
- Push to complete SPEC-010
- Supplement tests for SPEC-005, 007

### Short-term (1 month)
- Maintain SPEC creation frequency (3-4 per week)
- Maintain test coverage above 85%

### Long-term (3 months)
- Establish SPEC-First workflow
- Target 20 SPECs per month
```

### Generation Frequency
- Automatically generated at midnight on last day of month
- Existing reports are archived

---

## Metric Definitions

### Key Metrics

| Metric | Definition | Target |
|--------|------|------|
| **SPEC Creation Rate** | Number of SPECs created per month | 15-20 |
| **Completion Rate** | Completed SPECs / Total SPECs | 90%+ |
| **Time Savings** | (Estimated Time - Actual Time) / Estimated Time | 25%+ |
| **Test Coverage** | Average test coverage | 85%+ |
| **Code Linkage** | SPECs linked to code / Total SPECs | 90%+ |

### Calculation Formulas

```
Completion Rate = Completed / Total * 100
Time Savings = (Estimated - Actual) / Estimated * 100
Code Linkage = Specs with Commits / Total Specs * 100
Avg Coverage = Sum of Coverage / Total Specs
```

---

## Data Retention Policy

### Storage Locations
```
.moai/logs/
├── spec-usage.json              # Current state (latest)
├── spec-usage-YYYY-MM-DD.json   # Daily backup (optional)

.moai/reports/
├── spec-analytics-2025-11.md    # Monthly report
├── spec-analytics-2025-10.md
└── ...
```

### Retention Period
- Current month: Full data
- Previous 11 months: Full data
- 12+ months old: Summary statistics only

---

## Privacy Protection

### Data Collection Principles
- 📊 Collect statistics only (minimize personal information)
- 🔒 Local storage (no external transmission)
- 📝 Audit trail available

### Data Security
- `.moai/logs/` : Added to `.gitignore`
- Collection only in local development environment
- No automatic collection in CI/CD environments

---

## FAQ

**Q: Are the statistics accurate?**
A: Since they're automatically collected, they may not be perfect. Use them for identifying major trends.

**Q: Where is data stored?**
A: Stored locally in the `.moai/logs/` directory. Not transmitted externally.

**Q: Can statistics collection be disabled?**
A: Yes, remove the hook from `.claude/hooks/sessionend.sh`.

---

**Document Version**: 1.0.0
**Last Updated**: 2025-11-21
**Status**: Production Ready
