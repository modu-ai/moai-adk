---
# Report Metadata Header Template
# Copy this YAML frontmatter to all new reports

report_type: sync|analysis|validation|audit|implementation|test|plan|completion|regression
# Required: One of the valid report types
# sync: 동기화 결과 보고서
# analysis: 분석 보고서
# validation: 검증/검사 보고서
# audit: 감사 보고서
# implementation: 구현 결과 보고서
# test: 테스트 결과 보고서
# plan: 계획 및 전략 문서
# completion: 완료/최종 보고서
# regression: 회귀 분석 보고서

generated_by: alfred|user|system
# Required: Who/what generated this report

generated_at: "2025-11-04T11:00:00Z"
# Required: ISO 8601 timestamp with Z suffix (UTC)

purpose: "Brief description of report purpose"
# Required: One-line summary of what this report covers

scope: Full|Partial|Specific
# Required: Scope of coverage
# Full: Entire system/phase
# Partial: Subset of system
# Specific: Specific component/area

status: Complete|Incomplete|InProgress|Failed
# Required: Completion status

spec_id: SPEC-TRANSLATION-001
# Optional: Associated SPEC document ID (if applicable)

retention_days: 30
# Optional: Days to retain this report (default: 30)
# Use 90+ for important reports, spec-related reports
# Use values in permanent_tags for indefinite retention

tags:
  - translation
  - implementation
  - analysis
# Optional: Tags for categorization and search
# Examples: translation, implementation, performance, security, hooks, skills, etc.

related_documents:
  - path: "src/moai_adk/templates/.claude/commands/alfred/0-project.md"
    section: "STEP 2.1.4"
  - path: ".moai/docs/runtime-translation-flow.md"
    description: "Complete translation flow documentation"
# Optional: Related files or documentation references

author: "🎩 Alfred@MoAI"
# Optional: Author name or Alfred information

version: "1.0"
# Optional: Report version

---

# [Report Title Goes Here]

## 📋 Executive Summary

Brief summary of findings, key metrics, and recommendations.

## 📊 Key Findings

- Finding 1
- Finding 2
- Finding 3

## 📈 Metrics and Data

| Metric | Value | Status |
|--------|-------|--------|
| Item 1 | Value | ✅ |
| Item 2 | Value | ⚠️  |

## 🎯 Recommendations

1. Recommendation 1
2. Recommendation 2
3. Recommendation 3

## 📝 Implementation Details

Detailed analysis and implementation information.

## ✅ Conclusion

Final conclusions and next steps.

---

## Report Metadata

- **Generated**: {{generated_at}}
- **Type**: {{report_type}}
- **Status**: {{status}}
- **Retention**: {{retention_days}} days
- **Tags**: {{tags}}

Generated with Claude Code 🤖
