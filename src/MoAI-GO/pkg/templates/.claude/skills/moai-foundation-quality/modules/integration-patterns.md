# Integration Patterns

Enterprise integration patterns for CI/CD pipelines, GitHub Actions, and Quality-as-Service APIs.

## CI/CD Pipeline Integration

```python
async def quality_gate_pipeline():
 """Integrate quality validation into CI/CD pipeline"""

 # Initialize quality orchestrator
 quality_orchestrator = QualityOrchestrator.from_config("quality-config.yaml")

 # Run comprehensive quality analysis
 quality_result = await quality_orchestrator.analyze_codebase(
 path="src/",
 languages=["python", "typescript"],
 quality_threshold=0.85
 )

 # Quality gate validation
 if not quality_result.trust5_validation.passed:
 print(" Quality gate failed!")
 print(f"Overall score: {quality_result.overall_score:.2f}")

 # Print failed principles
 for principle, result in quality_result.trust5_validation.principles.items():
 if not result.passed:
 print(f" {principle}: {result.score:.2f} (threshold: 0.80)")

 # Exit with error code
 sys.exit(1)

 # Check for critical security issues
 critical_issues = [
 issue for issue in quality_result.proactive_analysis.recommendations
 if issue.severity == "critical" and issue.category == "security"
 ]

 if critical_issues:
 print(f" Found {len(critical_issues)} critical security issues!")
 for issue in critical_issues:
 print(f" - {issue.description}")
 sys.exit(1)

 print(" Quality gate passed!")
 print(f"Overall quality score: {quality_result.overall_score:.2f}")

 # Generate quality report
 await generate_quality_report(quality_result, output_path="quality-report.json")
```

## GitHub Actions Integration

```python
async def github_actions_quality_check():
 """Quality check for GitHub Actions workflow"""

 # Parse inputs
 github_token = os.getenv("GITHUB_TOKEN")
 repo_path = os.getenv("GITHUB_WORKSPACE", ".")
 pr_number = os.getenv("PR_NUMBER")

 # Run quality analysis
 quality_orchestrator = QualityOrchestrator()
 quality_result = await quality_orchestrator.analyze_codebase(
 path=repo_path,
 languages=["python", "javascript", "typescript"]
 )

 # Post comment on PR if quality issues found
 if pr_number and quality_result.overall_score < 0.85:
 comment = generate_pr_quality_comment(quality_result)
 await post_github_comment(github_token, pr_number, comment)

 # Set output for GitHub Actions
 print(f"::set-output name=quality_score::{quality_result.overall_score}")
 print(f"::set-output name=quality_passed::{quality_result.trust5_validation.passed}")
```

## Quality-as-Service REST API

```python
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel

app = FastAPI(title="Code Quality Analysis API")

class QualityAnalysisRequest(BaseModel):
 repository_url: str
 languages: List[str]
 quality_threshold: float = 0.85

class QualityAnalysisResponse(BaseModel):
 analysis_id: str
 overall_score: float
 trust5_validation: Dict
 recommendations: List[Dict]
 analysis_completed_at: datetime

@app.post("/api/quality/analyze", response_model=QualityAnalysisResponse)
async def analyze_quality(request: QualityAnalysisRequest):
 """API endpoint for quality analysis"""

 try:
 # Clone and analyze repository
 with tempfile.TemporaryDirectory() as temp_dir:
 await clone_repository(request.repository_url, temp_dir)

 quality_orchestrator = QualityOrchestrator()
 quality_result = await quality_orchestrator.analyze_codebase(
 path=temp_dir,
 languages=request.languages,
 quality_threshold=request.quality_threshold
 )

 return QualityAnalysisResponse(
 analysis_id=str(uuid.uuid4()),
 overall_score=quality_result.overall_score,
 trust5_validation=quality_result.trust5_validation.dict(),
 recommendations=[rec.dict() for rec in quality_result.proactive_analysis.recommendations],
 analysis_completed_at=datetime.now(UTC)
 )

 except Exception as e:
 raise HTTPException(status_code=500, detail=str(e))
```

## Cross-Project Quality Benchmarking

```python
class QualityBenchmarking:
 """Cross-project quality benchmarking and comparison"""

 def __init__(self, benchmark_database: str):
 self.benchmark_db = benchmark_database
 self.comparison_metrics = [
 "code_coverage",
 "security_score",
 "maintainability_index",
 "technical_debt_ratio",
 "duplicate_code_percentage"
 ]

 async def benchmark_project(self, project_path: str, project_metadata: Dict) -> BenchmarkResult:
 """Benchmark project quality against similar projects"""

 # Analyze current project
 current_metrics = await self._analyze_project_quality(project_path)

 # Find comparable projects from database
 comparable_projects = await self._find_comparable_projects(project_metadata)

 # Calculate percentiles and rankings
 benchmark_comparison = await self._calculate_benchmark_comparison(
 current_metrics, comparable_projects
 )

 # Generate improvement recommendations based on top performers
 improvement_recommendations = await self._generate_benchmark_recommendations(
 current_metrics, benchmark_comparison
 )

 return BenchmarkResult(
 project_metrics=current_metrics,
 benchmark_comparison=benchmark_comparison,
 industry_percentiles=benchmark_comparison.percentiles,
 improvement_roadmap=improvement_recommendations,
 competitive_analysis=self._analyze_competitive_position(
 current_metrics, benchmark_comparison
 )
 )

 async def _find_comparable_projects(self, project_metadata: Dict) -> List[ProjectMetrics]:
 """Find projects with similar characteristics for comparison"""

 query = {
 "language": project_metadata.get("language"),
 "project_type": project_metadata.get("type", "web_application"),
 "team_size_range": self._get_team_size_range(project_metadata.get("team_size", 5)),
 "industry": project_metadata.get("industry", "technology")
 }

 # Query benchmark database
 comparable_projects = await self.benchmark_db.find_projects(query)

 return comparable_projects[:50] # Limit to top 50 comparable projects
```

## Related

- [TRUST 5 Validation](trust5-validation.md)
- [Proactive Analysis](proactive-analysis.md)
- [Best Practices Engine](best-practices.md)
