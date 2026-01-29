# Enterprise Code Quality Examples

Comprehensive examples demonstrating the moai-foundation-quality skill in action across various scenarios and use cases.

---

## Example 1: Basic Quality Gate Validation

Scenario: Validate code quality before deployment

```python
from moai_core_quality import QualityOrchestrator, QualityConfig

# Initialize quality orchestrator with TRUST 5 validation
config = QualityConfig(
 trust5_enabled=True,
 quality_threshold=0.85,
 languages=["python", "typescript"]
)

quality_orchestrator = QualityOrchestrator(config)

# Run comprehensive quality analysis
result = await quality_orchestrator.analyze_codebase(
 path="src/",
 languages=["python", "typescript"],
 quality_threshold=0.85
)

print(f"Overall Quality Score: {result.overall_score:.2f}")
print(f"TRUST 5 Validation: {' PASSED' if result.trust5_validation.passed else ' FAILED'}")

# Check individual TRUST 5 principles
for principle, validation in result.trust5_validation.principles.items():
 status = "" if validation.passed else ""
 print(f"{status} {principle.title()}: {validation.score:.2f}")

# Generate quality report
await quality_orchestrator.generate_report(
 result=result,
 output_path="quality-report.html",
 format="html"
)
```

Output:
```
Overall Quality Score: 0.87
TRUST 5 Validation: PASSED
 Testable: 0.92
 Readable: 0.84
 Unified: 0.88
 Secured: 0.91
 Trackable: 0.80
```

---

## Example 2: Proactive Security Analysis

Scenario: Identify and fix security vulnerabilities in web application

```python
from moai_core_quality import ProactiveSecurityScanner

security_scanner = ProactiveSecurityScanner(
 context7_integration=True,
 owasp_compliance=True
)

# Focus on security vulnerabilities
security_issues = await security_scanner.scan_security_issues(
 codebase="src/web_application/",
 scan_types=[
 "sql_injection",
 "xss_vulnerabilities",
 "authentication_bypass",
 "sensitive_data_exposure",
 "dependency_vulnerabilities"
 ]
)

print(f"Found {len(security_issues.critical_issues)} critical security issues:")
for issue in security_issues.critical_issues:
 print(f" {issue.title}")
 print(f" Location: {issue.file}:{issue.line}")
 print(f" Description: {issue.description}")
 print(f" Fix: {issue.recommended_fix}")
 print()

# Auto-fix low-risk issues
auto_fixed = await security_scanner.auto_fix_issues(
 issues=security_issues.low_risk_issues,
 backup=True
)

print(f"Auto-fixed {len(auto_fixed)} security issues")

# Generate security report
await security_scanner.generate_security_report(
 issues=security_issues,
 output_path="security-analysis.json",
 format="sarif" # GitHub Actions compatible
)
```

Sample Security Issues Found:
```json
{
 "critical_issues": [
 {
 "title": "SQL Injection in User Authentication",
 "file": "src/auth/user_service.py",
 "line": 45,
 "severity": "critical",
 "description": "Direct SQL query construction with user input",
 "recommended_fix": "Use parameterized queries or ORM",
 "owasp_category": "A03:2021 – Injection"
 }
 ],
 "medium_issues": [
 {
 "title": "Missing Rate Limiting",
 "file": "src/api/endpoints.py",
 "line": 120,
 "severity": "medium",
 "description": "API endpoint lacks rate limiting protection",
 "recommended_fix": "Implement rate limiting middleware"
 }
 ]
}
```

---

## Example 3: Performance Optimization Analysis

Scenario: Identify performance bottlenecks and optimization opportunities

```python
from moai_core_quality import PerformanceAnalyzer

performance_analyzer = PerformanceAnalyzer(
 context7_integration=True,
 benchmark_comparison=True
)

# Analyze performance issues
performance_result = await performance_analyzer.analyze_performance(
 codebase="src/",
 focus_areas=[
 "database_queries",
 "algorithm_complexity",
 "memory_usage",
 "api_response_time",
 "concurrent_operations"
 ]
)

print("Performance Analysis Results:")
print(f"Overall Performance Score: {performance_result.overall_score:.2f}")

# Performance hotspots
print("\n Performance Hotspots:")
for hotspot in performance_result.hotspots:
 print(f" {hotspot.file}:{hotspot.line} - {hotspot.issue}")
 print(f" Impact: {hotspot.performance_impact}")
 print(f" Suggested Fix: {hotspot.optimization}")

# Benchmark comparison
if performance_result.benchmark_comparison:
 print(f"\n Benchmark Percentiles:")
 for metric, percentile in performance_result.benchmark_comparison.items():
 print(f" {metric}: {percentile}th percentile")

# Generate optimization recommendations
recommendations = await performance_analyzer.generate_optimization_plan(
 issues=performance_result.issues,
 priority="high",
 estimated_impact=True
)

print(f"\n Optimization Plan:")
for i, rec in enumerate(recommendations, 1):
 print(f"{i}. {rec.title}")
 print(f" Priority: {rec.priority}")
 print(f" Estimated Impact: {rec.estimated_impact}")
 print(f" Effort: {rec.estimated_effort}")
```

Performance Issues Identified:
```python
# Example output
Performance Analysis Results:
Overall Performance Score: 0.73

 Performance Hotspots:
 src/database/user_repository.py:89 - N+1 query problem
 Impact: High (200ms+ additional queries per request)
 Suggested Fix: Implement eager loading or batch queries

 src/algorithms/sort_service.py:34 - O(n²) algorithm
 Impact: Medium (exponential time complexity)
 Suggested Fix: Use built-in sort or implement O(n log n) algorithm

 Benchmark Percentiles:
 database_efficiency: 45th percentile
 algorithm_efficiency: 78th percentile
 memory_usage: 62nd percentile
```

---

## Example 4: Multi-Language Project Quality Analysis

Scenario: Analyze quality across Python, TypeScript, and Go codebase

```python
from moai_core_quality import MultiLanguageQualityAnalyzer

multi_lang_analyzer = MultiLanguageQualityAnalyzer(
 languages=["python", "typescript", "go"],
 context7_integration=True
)

# Analyze each language with specific rules
language_configs = {
 "python": {
 "style_guide": "pep8",
 "linter": "ruff",
 "formatter": "black",
 "type_checker": "mypy",
 "test_framework": "pytest"
 },
 "typescript": {
 "style_guide": "google",
 "linter": "eslint",
 "formatter": "prettier",
 "type_checker": "tsc",
 "test_framework": "jest"
 },
 "go": {
 "style_guide": "gofmt",
 "linter": "golangci-lint",
 "formatter": "gofmt",
 "type_checker": "go vet",
 "test_framework": "go test"
 }
}

results = await multi_lang_analyzer.analyze_multilang_project(
 project_root="src/",
 language_configs=language_configs
)

print("Multi-Language Quality Report:")
print("=" * 50)

for lang, result in results.items():
 print(f"\n {lang.upper()}:")
 print(f" Quality Score: {result.overall_score:.2f}")
 print(f" Test Coverage: {result.test_coverage:.1%}")
 print(f" Code Issues: {len(result.issues)}")

 # Language-specific issues
 if result.language_specific_issues:
 print(" Language-Specific Issues:")
 for issue in result.language_specific_issues[:3]: # Show top 3
 print(f" - {issue.title} ({issue.severity})")

# Cross-language consistency analysis
consistency = await multi_lang_analyzer.analyze_consistency(results)
print(f"\n Cross-Language Consistency: {consistency.overall_score:.2f}")

if consistency.inconsistencies:
 print("Inconsistencies Found:")
 for inconsistency in consistency.inconsistencies:
 print(f" - {inconsistency.description}")
```

Sample Multi-Language Output:
```
Multi-Language Quality Report:
==================================================

 PYTHON:
 Quality Score: 0.86
 Test Coverage: 92.3%
 Code Issues: 15
 Language-Specific Issues:
 - Unused imports (low)
 - Missing type hints (medium)
 - Long parameter lists (medium)

 TYPESCRIPT:
 Quality Score: 0.89
 Test Coverage: 87.1%
 Code Issues: 8
 Language-Specific Issues:
 - Any types used (medium)
 - Missing null checks (high)

 GO:
 Quality Score: 0.91
 Test Coverage: 85.7%
 Code Issues: 5
 Language-Specific Issues:
 - Unhandled errors (medium)
 - Inefficient string concatenation (low)

 Cross-Language Consistency: 0.78
```

---

## Example 5: CI/CD Pipeline Integration

Scenario: Integrate quality gates into GitHub Actions pipeline

```yaml
# .github/workflows/quality-gate.yml
name: Quality Gate

on:
 pull_request:
 branches: [ main ]
 push:
 branches: [ main ]

jobs:
 quality-analysis:
 runs-on: ubuntu-latest

 steps:
 - uses: actions/checkout@v3

 - name: Setup Python
 uses: actions/setup-python@v4
 with:
 python-version: '3.11'

 - name: Install dependencies
 run: |
 pip install moai-foundation-quality
 npm install -g eslint prettier

 - name: Run Quality Analysis
 run: |
 python -c "
 import asyncio
 from moai_core_quality import QualityOrchestrator

 async def run_quality_analysis():
 orchestrator = QualityOrchestrator()
 result = await orchestrator.analyze_codebase(
 path='src/',
 languages=['python', 'typescript'],
 quality_threshold=0.85
 )

 # Set GitHub Actions outputs
 with open('quality_output.txt', 'w') as f:
 f.write(f'quality_score={result.overall_score:.2f}\n')
 f.write(f'quality_passed={result.trust5_validation.passed}\n')
 f.write(f'test_coverage={result.trust5_validation.principles[\"testable\"].score:.2f}\n')

 # Generate detailed report
 await orchestrator.generate_report(
 result=result,
 output_path='quality-report.html',
 format='html'
 )

 # Exit with error if quality gate fails
 if not result.trust5_validation.passed:
 print(' Quality gate failed!')
 exit(1)
 print(' Quality gate passed!')

 asyncio.run(run_quality_analysis())
 "

 - name: Upload Quality Report
 uses: actions/upload-artifact@v3
 with:
 name: quality-report
 path: quality-report.html

 - name: Set Quality Outputs
 run: |
 while read line; do
 echo "::set-output name=${line%%=*}::${line#*=}"
 done < quality_output.txt
```

Python Quality Analysis Script:
```python
# quality_pipeline.py
import asyncio
import json
import sys
from pathlib import Path
from moai_core_quality import QualityOrchestrator, CIIntegration

async def run_pipeline_quality_check():
 """Run quality check for CI/CD pipeline"""

 # Configuration
 config = {
 "codebase_path": "src/",
 "languages": ["python", "typescript"],
 "quality_threshold": 0.85,
 "fail_on_threshold": True,
 "generate_reports": True,
 "output_formats": ["json", "html", "junit"]
 }

 # Initialize quality orchestrator
 orchestrator = QualityOrchestrator()
 ci_integration = CIIntegration()

 print(" Starting quality analysis...")

 # Run quality analysis
 result = await orchestrator.analyze_codebase(
 path=config["codebase_path"],
 languages=config["languages"],
 quality_threshold=config["quality_threshold"]
 )

 # Generate CI/CD outputs
 await ci_integration.generate_pipeline_outputs(
 result=result,
 output_dir="quality-reports/",
 formats=config["output_formats"]
 )

 # Quality gate validation
 if not result.trust5_validation.passed:
 print(" Quality gate failed!")

 # Print failed principles
 failed_principles = [
 name for name, validation in result.trust5_validation.principles.items()
 if not validation.passed
 ]

 print(f"Failed principles: {', '.join(failed_principles)}")

 # Exit with error if configured
 if config["fail_on_threshold"]:
 sys.exit(1)

 print(f" Quality gate passed! Score: {result.overall_score:.2f}")

 # Create summary for CI/CD dashboard
 summary = {
 "quality_score": result.overall_score,
 "quality_passed": result.trust5_validation.passed,
 "test_coverage": result.trust5_validation.principles["testable"].score,
 "security_score": result.trust5_validation.principles["secured"].score,
 "issues_found": len(result.proactive_analysis.recommendations),
 "critical_issues": len([
 r for r in result.proactive_analysis.recommendations
 if r.severity == "critical"
 ])
 }

 # Save summary for dashboard
 with open("quality-summary.json", "w") as f:
 json.dump(summary, f, indent=2)

 print(" Quality summary saved to quality-summary.json")

if __name__ == "__main__":
 asyncio.run(run_pipeline_quality_check())
```

---

## Example 6: Custom Quality Rules and Policies

Scenario: Define project-specific quality rules and policies

```python
from moai_core_quality import QualityOrchestrator, CustomQualityRule, QualityPolicy

# Define custom quality rules for enterprise project
class EnterpriseNamingRule(CustomQualityRule):
 """Enterprise naming convention rule"""

 def __init__(self):
 super().__init__(
 name="enterprise_naming_conventions",
 description="Enforce enterprise naming standards",
 severity="medium"
 )

 async def validate(self, codebase: str):
 patterns = {
 "api_endpoints": r"^[a-z]+_[a-z_]+$", # get_user_profile
 "database_models": r"^[A-Z][a-zA-Z]*Model$", # UserProfileModel
 "service_classes": r"^[A-Z][a-zA-Z]*Service$", # UserService
 "constants": r"^[A-Z][A-Z_]*$" # MAX_RETRY_COUNT
 }

 violations = []
 for pattern_name, pattern in patterns.items():
 pattern_violations = await self._check_pattern(
 codebase, pattern, pattern_name
 )
 violations.extend(pattern_violations)

 return self.create_result(
 passed=len(violations) == 0,
 details={"violations": violations},
 recommendations=[
 f"Fix {len(violations)} naming convention violations"
 ]
 )

class SecurityPolicyRule(CustomQualityRule):
 """Enterprise security policy rule"""

 def __init__(self):
 super().__init__(
 name="enterprise_security_policy",
 description="Enforce enterprise security standards",
 severity="high"
 )

 async def validate(self, codebase: str):
 security_violations = []

 # Check for hardcoded secrets
 secret_violations = await self._detect_hardcoded_secrets(codebase)
 security_violations.extend(secret_violations)

 # Check for insecure direct object references
 idor_violations = await self._detect_idor_patterns(codebase)
 security_violations.extend(idor_violations)

 # Check for missing authentication
 auth_violations = await self._detect_missing_auth(codebase)
 security_violations.extend(auth_violations)

 return self.create_result(
 passed=len(security_violations) == 0,
 details={"security_violations": security_violations},
 recommendations=[
 "Fix security policy violations immediately"
 ]
 )

# Create enterprise quality policy
enterprise_policy = QualityPolicy(
 name="enterprise_quality_policy",
 description="Enterprise-grade quality standards",
 custom_rules=[
 EnterpriseNamingRule(),
 SecurityPolicyRule()
 ],
 quality_thresholds={
 "overall": 0.90,
 "testable": 0.95,
 "secured": 0.95,
 "readable": 0.85,
 "unified": 0.90,
 "trackable": 0.85
 }
)

# Apply policy to quality analysis
orchestrator = QualityOrchestrator()
orchestrator.apply_policy(enterprise_policy)

result = await orchestrator.analyze_codebase(
 path="enterprise_app/",
 languages=["python", "typescript"],
 policy=enterprise_policy
)

print(f"Enterprise Quality Score: {result.overall_score:.2f}")
print(f"Policy Compliance: {' COMPLIANT' if result.policy_compliance.passed else ' NON-COMPLIANT'}")
```

---

## Example 7: Technical Debt Analysis and Management

Scenario: Analyze and manage technical debt across codebase

```python
from moai_core_quality import TechnicalDebtAnalyzer

debt_analyzer = TechnicalDebtAnalyzer(
 context7_integration=True,
 debt_categories=[
 "code_complexity",
 "code_duplication",
 "lack_of_testing",
 "outdated_dependencies",
 "poor_documentation",
 "performance_issues"
 ]
)

# Comprehensive technical debt analysis
debt_analysis = await debt_analyzer.analyze_technical_debt(
 codebase="src/",
 languages=["python", "javascript"],
 include_historical_trends=True
)

print("Technical Debt Analysis")
print("=" * 40)

# Overall debt metrics
print(f"Total Technical Debt: {debt_analysis.total_debt_hours} hours")
print(f"Debt Ratio: {debt_analysis.debt_ratio:.2%}")
print(f"Debt Interest: {debt_analysis.monthly_debt_interest} hours/month")

# Debt by category
print("\n Debt by Category:")
for category, debt in debt_analysis.debt_by_category.items():
 print(f" {category}: {debt.hours} hours ({debt.percentage:.1f}%)")
 print(f" Files affected: {debt.files_affected}")
 print(f" Priority: {debt.priority}")

# Hotspots (files with most debt)
print("\n Technical Debt Hotspots:")
for hotspot in debt_analysis.debt_hotspots[:5]:
 print(f" {hotspot.file}: {hotspot.total_debt} hours")
 print(f" Issues: {len(hotspot.issues)}")
 print(f" Main issue: {hotspot.primary_issue.title}")

# Debt repayment roadmap
repayment_plan = await debt_analyzer.create_repayment_plan(
 total_budget=80, # 80 hours per sprint
 priority_strategy="highest_roi"
)

print("\n Debt Repayment Plan:")
for i, item in enumerate(repayment_plan.items[:10], 1):
 print(f"{i}. {item.title}")
 print(f" Effort: {item.estimated_hours} hours")
 print(f" ROI: {item.roi_score}")
 print(f" Files: {', '.join(item.files)}")

# Generate technical debt report
await debt_analyzer.generate_debt_report(
 analysis=debt_analysis,
 output_path="technical-debt-report.html",
 include_roadmap=True,
 format="html"
)

# Track debt over time
if debt_analysis.historical_data:
 print("\n Debt Trend:")
 for period in debt_analysis.historical_data[-6:]: # Last 6 periods
 print(f" {period.date}: {period.total_debt} hours")
```

Sample Technical Debt Output:
```
Technical Debt Analysis
========================================
Total Technical Debt: 245 hours
Debt Ratio: 12.3%
Debt Interest: 18 hours/month

 Debt by Category:
 code_complexity: 78 hours (31.8%)
 Files affected: 23
 Priority: high
 lack_of_testing: 65 hours (26.5%)
 Files affected: 15
 Priority: high
 code_duplication: 42 hours (17.1%)
 Files affected: 8
 Priority: medium
 outdated_dependencies: 35 hours (14.3%)
 Files affected: 12
 Priority: medium
 performance_issues: 25 hours (10.2%)
 Files affected: 6
 Priority: low

 Technical Debt Hotspots:
 src/services/user_service.py: 45 hours
 Issues: 8
 Main issue: Complex method with high cyclomatic complexity
 src/utils/data_processor.py: 32 hours
 Issues: 5
 Main issue: Duplicate code blocks
```

---

## Example 8: Real-time Quality Monitoring

Scenario: Set up real-time quality monitoring with alerts

```python
from moai_core_quality import RealTimeQualityMonitor, AlertConfiguration

# Configure real-time monitoring
alert_config = AlertConfiguration(
 webhook_url="https://hooks.slack.com/services/...",
 email_recipients=["dev-team@company.com"],
 alert_thresholds={
 "quality_degradation": 0.05, # 5% drop triggers alert
 "security_score": 0.85,
 "test_coverage": 0.80,
 "critical_vulnerabilities": 0
 },
 notification_channels=["slack", "email", "dashboard"]
)

# Initialize real-time monitor
quality_monitor = RealTimeQualityMonitor(
 codebase_path="src/",
 check_interval=300, # 5 minutes
 alert_config=alert_config,
 context7_integration=True
)

print(" Starting real-time quality monitoring...")

# Start monitoring in background
import asyncio
monitoring_task = asyncio.create_task(
 quality_monitor.start_monitoring()
)

try:
 # Monitor for 24 hours
 await asyncio.sleep(86400)
except KeyboardInterrupt:
 print("\n⏹ Stopping quality monitoring...")
 monitoring_task.cancel()
 await monitoring_task

# Generate monitoring report
monitoring_report = await quality_monitor.generate_monitoring_report(
 duration_hours=24
)

print(f"Monitoring Summary:")
print(f" Quality checks performed: {monitoring_report.total_checks}")
print(f" Alerts triggered: {monitoring_report.alerts_triggered}")
print(f" Average quality score: {monitoring_report.avg_quality_score:.2f}")
print(f" Quality trends: {monitoring_report.quality_trend}")
```

Custom Alert Handler:
```python
class CustomQualityAlertHandler:
 """Custom handler for quality alerts"""

 def __init__(self, slack_webhook: str, jira_api: str):
 self.slack_webhook = slack_webhook
 self.jira_api = jira_api

 async def handle_quality_alert(self, alert: QualityAlert):
 """Handle quality alerts with custom logic"""

 # Create JIRA ticket for critical issues
 if alert.severity == "critical":
 jira_ticket = await self.create_jira_ticket(alert)
 print(f" Created JIRA ticket: {jira_ticket}")

 # Send Slack notification with rich formatting
 slack_message = self.format_slack_message(alert)
 await self.send_slack_notification(slack_message)

 # Log to quality dashboard
 await self.log_to_dashboard(alert)

 def format_slack_message(self, alert: QualityAlert) -> dict:
 """Format alert for Slack notification"""

 colors = {
 "critical": "#ff0000",
 "high": "#ff9900",
 "medium": "#ffff00",
 "low": "#00ff00"
 }

 return {
 "attachments": [{
 "color": colors.get(alert.severity, "#808080"),
 "title": f" Quality Alert: {alert.title}",
 "text": alert.description,
 "fields": [
 {"title": "Severity", "value": alert.severity.upper(), "short": True},
 {"title": "Quality Score", "value": f"{alert.current_score:.2f}", "short": True},
 {"title": "File", "value": alert.location, "short": True}
 ],
 "footer": "Quality Monitor",
 "ts": alert.timestamp.timestamp()
 }]
 }

# Register custom alert handler
quality_monitor.register_alert_handler(
 CustomQualityAlertHandler(
 slack_webhook="...",
 jira_api="..."
 )
)
```

---

## Example 9: Quality Benchmarking and Comparison

Scenario: Benchmark project quality against industry standards

```python
from moai_core_quality import QualityBenchmarking

# Initialize benchmarking service
benchmarking = QualityBenchmarking(
 industry_database="enterprise_web_apps",
 comparison_criteria=["project_size", "team_size", "industry"]
)

# Analyze current project quality
project_metadata = {
 "language": ["python", "typescript"],
 "project_type": "web_application",
 "team_size": 8,
 "industry": "fintech",
 "lines_of_code": 50000
}

benchmark_result = await benchmarking.benchmark_project(
 codebase="src/",
 project_metadata=project_metadata
)

print(" Quality Benchmarking Results")
print("=" * 40)

# Current project metrics
print(f"Current Project Quality Score: {benchmark_result.project_metrics.overall_score:.2f}")

# Industry comparison
print(f"\n Industry Percentiles:")
for metric, percentile in benchmark_result.industry_percentiles.items():
 percentile_emoji = "" if percentile >= 90 else "" if percentile >= 75 else "" if percentile >= 50 else ""
 print(f" {percentile_emoji} {metric}: {percentile}th percentile")

# Competitive analysis
print(f"\n Competitive Position:")
competitive_pos = benchmark_result.competitive_analysis
print(f" Overall Rank: {competitive_pos.overall_rank}/100")
print(f" Top Quartile: {'' if competitive_pos.is_top_quartile else ''}")
print(f" Industry Leader: {'' if competitive_pos.is_industry_leader else ''}")

# Improvement roadmap based on top performers
print(f"\n Improvement Roadmap:")
for i, recommendation in enumerate(benchmark_result.improvement_roadmap[:5], 1):
 print(f"{i}. {recommendation.area}")
 print(f" Target: {recommendation.target_score}")
 print(f" Top Performer Score: {recommendation.top_performer_score}")
 print(f" Key Actions: {', '.join(recommendation.key_actions[:3])}")

# Generate comprehensive benchmark report
await benchmarking.generate_benchmark_report(
 result=benchmark_result,
 output_path="quality-benchmark.html",
 include_industry_trends=True,
 format="html"
)
```

---

## Example 10: Automated Code Refactoring Suggestions

Scenario: Generate and apply automated refactoring suggestions

```python
from moai_core_quality import RefactoringSuggester

refactoring_engine = RefactoringSuggester(
 auto_apply_safe_refactors=True,
 context7_integration=True,
 backup_before_changes=True
)

# Analyze code for refactoring opportunities
refactoring_opportunities = await refactoring_engine.analyze_refactoring_opportunities(
 codebase="src/",
 languages=["python"],
 focus_areas=[
 "code_duplication",
 "long_methods",
 "large_classes",
 "complex_conditionals",
 "magic_numbers",
 "unused_variables"
 ]
)

print(" Refactoring Opportunities Analysis")
print("=" * 45)

# Group opportunities by type and priority
opportunity_groups = refactoring_engine.group_opportunities(
 refactoring_opportunities
)

for category, opportunities in opportunity_groups.items():
 print(f"\n {category.title()}:")

 high_priority = [opp for opp in opportunities if opp.priority == "high"]
 medium_priority = [opp for opp in opportunities if opp.priority == "medium"]

 if high_priority:
 print(f" High Priority ({len(high_priority)}):")
 for opp in high_priority[:3]: # Show top 3
 print(f" - {opp.description}")
 print(f" File: {opp.file}:{opp.line}")
 print(f" Effort: {opp.estimated_effort} minutes")
 print(f" Impact: {opp.estimated_impact}")

 if medium_priority:
 print(f" 🟡 Medium Priority ({len(medium_priority)}):")
 for opp in medium_priority[:3]:
 print(f" - {opp.description}")

# Generate refactoring plan
refactoring_plan = await refactoring_engine.create_refactoring_plan(
 opportunities=refactoring_opportunities,
 budget_hours=8, # 8 hours allocated for refactoring
 risk_tolerance="medium"
)

print(f"\n Refactoring Plan:")
print(f"Total planned refactors: {len(refactoring_plan.planned_refactors)}")
print(f"Estimated time: {refactoring_plan.total_estimated_hours} hours")
print(f"Expected quality improvement: +{refactoring_plan.quality_improvement:.2f}")

# Apply safe refactoring automatically
safe_refactors = [
 opp for opp in refactoring_opportunities
 if opp.risk_level == "low" and opp.auto_safe
]

if safe_refactors:
 print(f"\n Applying {len(safe_refactors)} safe refactoring operations...")

 applied_refactors = await refactoring_engine.apply_refactoring(
 refactors=safe_refactors,
 create_backup=True,
 dry_run=False
 )

 print(f"Successfully applied {len(applied_refactors)} refactoring operations:")

 for refactor in applied_refactors:
 print(f" {refactor.description}")
 print(f" File: {refactor.file}")
 print(f" Lines changed: {refactor.lines_changed}")
else:
 print("\n No safe automatic refactoring opportunities found")

# Generate before/after quality comparison
if safe_refactors:
 quality_comparison = await refactoring_engine.compare_quality(
 before_path="src/",
 after_path="src/", # Same path after refactoring
 refactors=applied_refactors
 )

 print(f"\n Quality Impact:")
 print(f"Before: {quality_comparison.before_score:.2f}")
 print(f"After: {quality_comparison.after_score:.2f}")
 print(f"Improvement: +{quality_comparison.improvement:.2f}")
```

---

## Example 11: Quality Metrics Dashboard

Scenario: Create comprehensive quality metrics dashboard

```python
from moai_core_quality import QualityDashboard, MetricsCollector

# Initialize metrics collector
metrics_collector = MetricsCollector(
 collection_interval=3600, # 1 hour
 historical_retention_days=90
)

# Collect comprehensive quality metrics
metrics_data = await metrics_collector.collect_comprehensive_metrics(
 codebase="src/",
 languages=["python", "typescript"],
 include_historical_trends=True
)

# Create quality dashboard
dashboard = QualityDashboard(
 title="Enterprise Quality Dashboard",
 refresh_interval=300, # 5 minutes
 theme="enterprise"
)

# Configure dashboard widgets
dashboard.add_widget(
 widget_type="overview",
 title="Quality Overview",
 metrics=["overall_score", "test_coverage", "security_score"],
 visualization="gauge"
)

dashboard.add_widget(
 widget_type="trend",
 title="Quality Trend (30 days)",
 metrics=["overall_score"],
 visualization="line_chart",
 time_range="30d"
)

dashboard.add_widget(
 widget_type="distribution",
 title="Issues by Severity",
 metrics=["issue_count"],
 group_by="severity",
 visualization="pie_chart"
)

dashboard.add_widget(
 widget_type="comparison",
 title="Language Comparison",
 metrics=["quality_score"],
 group_by="language",
 visualization="bar_chart"
)

dashboard.add_widget(
 widget_type="hotspot",
 title="Quality Hotspots",
 metrics=["issue_density"],
 visualization="heatmap"
)

# Generate dashboard HTML
dashboard_html = await dashboard.generate_dashboard(
 metrics_data=metrics_data,
 output_path="quality-dashboard.html",
 interactive=True
)

print(" Quality Dashboard Generated:")
print(f" Location: quality-dashboard.html")
print(f" Widgets: {len(dashboard.widgets)}")
print(f" Metrics tracked: {len(metrics_data.metrics)}")
print(f" Time range: 30 days")
```

Dashboard Widget Examples:
```python
# Custom widget for executive summary
executive_widget = {
 "type": "executive_summary",
 "title": "Executive Quality Summary",
 "metrics": {
 "overall_health": "excellent", # Based on score ranges
 "trend": "improving",
 "critical_issues": 2,
 "technical_debt_hours": 125,
 "quality_investment_roi": "3.2x"
 },
 "alerts": [
 {"type": "warning", "message": "Test coverage decreased by 3%"},
 {"type": "info", "message": "New security policies implemented"}
 ]
}

# Real-time widget for live monitoring
live_monitoring_widget = {
 "type": "real_time_monitor",
 "title": "Live Quality Monitoring",
 "refresh_rate": 60, # seconds
 "metrics": [
 {"name": "active_developers", "current": 8},
 {"name": "commits_today", "current": 15},
 {"name": "quality_checks_passed", "current": 142},
 {"name": "quality_checks_failed", "current": 3}
 ]
}
```

---

## Example 12: Enterprise Quality Governance

Scenario: Implement enterprise-wide quality governance policies

```python
from moai_core_quality import EnterpriseQualityGovernance, QualityPolicy, ComplianceReporter

# Define enterprise quality governance framework
governance = EnterpriseQualityGovernance(
 organization_name="Acme Corporation",
 quality_standards="ISO/IEC 25010",
 compliance_requirements=["SOC2", "GDPR", "OWASP"]
)

# Define quality policies for different project tiers
policies = {
 "critical": QualityPolicy(
 name="Critical Systems Policy",
 quality_threshold=0.95,
 mandatory_requirements=[
 "100% test coverage for critical paths",
 "Zero critical security vulnerabilities",
 "Code review by senior architect",
 "Performance testing with 99.9% SLA"
 ],
 approved_tools=["sonarqube", "coveralls", "snyk"],
 audit_frequency="weekly"
 ),

 "standard": QualityPolicy(
 name="Standard Systems Policy",
 quality_threshold=0.85,
 mandatory_requirements=[
 "85% test coverage",
 "No critical security vulnerabilities",
 "Code review by peer",
 "Basic performance testing"
 ],
 approved_tools=["eslint", "pytest", "bandit"],
 audit_frequency="monthly"
 ),

 "internal": QualityPolicy(
 name="Internal Tools Policy",
 quality_threshold=0.75,
 mandatory_requirements=[
 "70% test coverage",
 "Basic security review",
 "Documentation required"
 ],
 approved_tools=["pylint", "black"],
 audit_frequency="quarterly"
 )
}

# Register policies with governance framework
for tier, policy in policies.items():
 governance.register_policy(tier, policy)

# Project classification and policy application
projects = [
 {"name": "payment-processor", "tier": "critical", "path": "src/payment/"},
 {"name": "user-dashboard", "tier": "standard", "path": "src/dashboard/"},
 {"name": "internal-tools", "tier": "internal", "path": "src/tools/"}
]

compliance_results = []

for project in projects:
 # Apply appropriate policy
 policy = governance.get_policy_for_tier(project["tier"])

 # Check compliance
 compliance_check = await governance.check_compliance(
 codebase=project["path"],
 policy=policy,
 generate_report=True
 )

 compliance_results.append({
 "project": project["name"],
 "tier": project["tier"],
 "compliant": compliance_check.is_compliant,
 "score": compliance_check.quality_score,
 "violations": compliance_check.policy_violations,
 "remediation_plan": compliance_check.remediation_plan
 })

 print(f"\n {project['name']} ({project['tier'].upper()}):")
 print(f" Compliance: {' COMPLIANT' if compliance_check.is_compliant else ' NON-COMPLIANT'}")
 print(f" Quality Score: {compliance_check.quality_score:.2f}")
 print(f" Policy Threshold: {policy.quality_threshold}")

 if not compliance_check.is_compliant:
 print(f" Violations: {len(compliance_check.policy_violations)}")
 for violation in compliance_check.policy_violations[:3]:
 print(f" - {violation.requirement}: {violation.status}")

# Generate enterprise compliance report
compliance_reporter = ComplianceReporter()
enterprise_report = await compliance_reporter.generate_enterprise_report(
 compliance_results=compliance_results,
 governance_framework=governance,
 output_path="enterprise-compliance-report.html",
 include_executive_summary=True
)

print(f"\n Enterprise Compliance Summary:")
total_projects = len(compliance_results)
compliant_projects = sum(1 for r in compliance_results if r["compliant"])
compliance_rate = (compliant_projects / total_projects) * 100

print(f" Total Projects: {total_projects}")
print(f" Compliant Projects: {compliant_projects}")
print(f" Overall Compliance Rate: {compliance_rate:.1f}%")
print(f" Report Generated: enterprise-compliance-report.html")

# Schedule regular compliance audits
await governance.schedule_compliance_audits(
 frequency="monthly",
 notification_channels=["email", "slack"],
 auto_remediation=False
)

print(" Monthly compliance audits scheduled")
```

---

These examples demonstrate the comprehensive capabilities of the moai-foundation-quality skill across various enterprise scenarios, from basic quality validation to complex governance frameworks. Each example includes practical implementation details, output samples, and integration patterns suitable for production environments.
