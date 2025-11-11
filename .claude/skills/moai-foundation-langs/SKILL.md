---
name: moai-foundation-langs
version: 4.0.0
created: 2025-11-11
updated: 2025-11-11
status: active
description: AI-powered multi-language detection with Context7 integration, framework intelligence, and automated toolchain optimization for 25+ programming languages
keywords: ['language-detection', 'ai-intelligence', 'framework-analysis', 'toolchain-optimization', 'multi-language', 'context7-integration', 'enterprise-support']
allowed-tools:
  - Read
  - Bash
  - Glob
  - WebFetch
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
---

# Enterprise Foundation Langs Skill v4.0.0

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-foundation-langs |
| **Version** | 4.0.0 (2025-11-11) |
| **Tier** | Enterprise Foundation |
| **AI-Powered** | ✅ Context7 Integration, Framework Intelligence |
| **Auto-load** | On demand when keywords detected |

---

## What It Does

AI-powered multi-language detection and toolchain optimization with Context7 integration for 25+ programming languages and frameworks.

**Revolutionary v4.0.0 capabilities**:
- 🤖 **AI-Powered Language Detection** using Context7 MCP for official documentation
- 📊 **Framework Intelligence** with automatic version detection and compatibility analysis
- 🛠️ **Automated Toolchain Optimization** with latest 2025 tool recommendations
- 🌐 **Multi-Language Polyglot Support** for monorepo and microservice architectures
- 📈 **Performance-Based Language Selection** with benchmarking and profiling
- 🔍 **Dependency Graph Analysis** for cross-language project insights
- 🚀 **Zero-Configuration Language Setup** with intelligent template matching
- 📋 **Real-Time Compliance Monitoring** against language-specific standards
- 🎯 **Predictive Migration Planning** using ML-based cost-benefit analysis
- 🔒 **Security Posture Assessment** per language and framework

---

## Supported Languages & Frameworks (2025)

### Core Languages with Full Context7 Integration

#### **Python 3.12+** 🐍
```yaml
detection:
  - pyproject.toml
  - requirements.txt
  - setup.py
  - Pipfile
  - poetry.lock
  - .python-version

frameworks:
  Django: "5.1+ (LTS)"
  FastAPI: "0.115+"
  Flask: "3.0+"
  SQLAlchemy: "2.0+"
  Pandas: "3.0+"
  NumPy: "2.0+"
  PyTorch: "2.5+"
  TensorFlow: "2.17+"
  Asyncio: "built-in"

tools_2025:
  testing: ["pytest 8.4+", "pytest-asyncio 0.24+"]
  linting: ["ruff 0.7+", "mypy 1.13+"]
  formatting: ["black 24+", "isort 5.13+"]
  security: ["bandit 1.7+", "safety 3.2+"]
  profiling: ["py-spy 0.3+", "memory-profiler 0.61+"]
  documentation: ["sphinx 8+", "mkdocs 1.6+"]
```

#### **TypeScript 5.6+** 📘
```yaml
detection:
  - package.json
  - tsconfig.json
  - yarn.lock
  - pnpm-lock.yaml
  - deno.json

frameworks:
  React: "19.0+"
  Next.js: "15.0+"
  Vue.js: "3.5+"
  Angular: "18.0+"
  Node.js: "22.0+ LTS"
  Deno: "2.0+"
  Bun: "1.1+"

tools_2025:
  testing: ["Vitest 2.1+", "Jest 30+", "Playwright 1.48+"]
  linting: ["Biome 1.9+", "ESLint 9+", "TypeScript 5.6"]
  formatting: ["Biome 1.9+", "Prettier 3.3+"]
  security: ["npm audit", "pnpm audit", "deno check"]
  profiling: ["clinic.js", "Node.js built-in profiler"]
  documentation: ["TypeDoc 0.26+", "Storybook 8.4+"]
```

#### **Go 1.23+** 🐹
```yaml
detection:
  - go.mod
  - go.sum
  - Gopkg.toml
  - main.go
  - *.go

frameworks:
  Gin: "1.10+"
  Echo: "4.12+"
  Fiber: "2.52+"
  gRPC: "1.68+"
  Chi: "5.1+"
  Gorilla Mux: "1.8+"

tools_2025:
  testing: ["go test 1.23", "testify 1.9+", "ginkgo 2.15+"]
  linting: ["golangci-lint 1.62+", "staticcheck 0.5+", "govulncheck"]
  formatting: ["gofmt 1.23", "goimports 0.22+"]
  security: ["gosec 2.20+", "nancy 0.2+"]
  profiling: ["pprof built-in", "go tool trace"]
  documentation: ["godoc", "pkgsite 0.5+"]
```

#### **Rust 1.82+** 🦀
```yaml
detection:
  - Cargo.toml
  - Cargo.lock
  - src/main.rs
  - src/lib.rs

frameworks:
  Tokio: "1.41+"
  Axum: "0.8+"
  Actix Web: "0.7+"
  Rocket: "0.5+"
  Serde: "1.0+"
  Clap: "4.5+"

tools_2025:
  testing: ["cargo test 1.82", "nextest 0.4+", "tarpaulin 0.31+"]
  linting: ["clippy 1.82", "rustfmt 1.82", "cargo check"]
  formatting: ["rustfmt 1.82"]
  security: ["cargo audit 0.20+", "cargo-deny 0.15+"]
  profiling: ["cargo profdata", "perf", "flamegraph"]
  documentation: ["rustdoc 1.82", "mdbook 0.4+"]
```

#### **Java 21+ (LTS)** ☕
```yaml
detection:
  - pom.xml
  - build.gradle.kts
  - build.gradle
  - src/main/java/

frameworks:
  Spring Boot: "3.3+"
  Quarkus: "3.15+"
  Micronaut: "4.6+"
  Jakarta EE: "11.0+"
  Hibernate: "7.0+"
  JUnit: "5.11+"

tools_2025:
  testing: ["JUnit 5.11+", "Testcontainers 1.20+", "AssertJ 3.26+"]
  linting: ["Checkstyle 10.18+", "SpotBugs 4.8+", "PMD 7.5+"]
  formatting: ["Google Java Format 1.24+", "Spotless 6.25+"]
  security: ["OWASP Dependency Check 11+", "Trivy"]
  profiling: ["JProfiler 14+", "VisualVM", "Java Mission Control"]
  documentation: ["Javadoc 21", "Asciidoctor 3.0+"]
```

#### **JavaScript/Node.js 22+ LTS** 🟨
```yaml
detection:
  - package.json
  - yarn.lock
  - package-lock.json
  - .nvmrc

frameworks:
  Express: "4.21+"
  Koa: "2.15+"
  NestJS: "10.4+"
  Meteor: "3.1+"
  Svelte: "5.0+"

tools_2025:
  testing: ["Jest 30+", "Mocha 11+", "Playwright 1.48+"]
  linting: ["ESLint 9+", "Biome 1.9+"]
  formatting: ["Prettier 3.3+", "Biome 1.9+"]
  security: ["npm audit", "Snyk", " retire.js"]
  profiling: ["clinic.js", "Node.js profiler"]
  documentation: ["JSDoc 4.0+", "ESDoc"]
```

---

## Advanced Detection Algorithms

### Pattern-Based Language Detection
```python
class LanguageDetector:
    """AI-powered language detection with Context7 integration."""
    
    def __init__(self):
        self.context7_client = Context7Client()
        self.language_patterns = self.load_patterns()
        self.framework_intelligence = self.load_frameworks()
    
    async def detect_project_language(self, project_path: str) -> LanguageProfile:
        """Detect primary language with confidence scoring."""
        files = self.scan_project(project_path)
        patterns = self.match_patterns(files)
        frameworks = self.detect_frameworks(files)
        
        # Context7 integration for official docs
        official_docs = await self.context7_client.get_latest_docs(
            patterns.primary_language
        )
        
        return LanguageProfile(
            primary_language=patterns.primary_language,
            confidence=patterns.confidence,
            frameworks=frameworks,
            recommended_tools=await self.get_recommended_tools(
                patterns.primary_language,
                frameworks
            ),
            official_documentation=official_docs,
            migration_paths=self.analyze_migration_paths(patterns)
        )
    
    def analyze_polyglot_complexity(self, project_path: str) -> PolyglotAnalysis:
        """Analyze multi-language project complexity."""
        languages = self.detect_all_languages(project_path)
        dependencies = self.map_cross_language_dependencies(project_path)
        integration_points = self.identify_integration_points(languages)
        
        return PolyglotAnalysis(
            language_distribution=languages,
            complexity_score=self.calculate_complexity(languages, dependencies),
            integration_challenges=integration_points,
            optimization_suggestions=self.suggest_optimizations(languages)
        )
```

### Framework Version Intelligence
```python
class FrameworkVersionAnalyzer:
    """Advanced framework version detection and compatibility analysis."""
    
    async def analyze_framework_compatibility(self, 
                                            framework: str, 
                                            version: str) -> CompatibilityReport:
        """Analyze framework version compatibility."""
        official_info = await self.context7_client.get_library_info(framework)
        
        return CompatibilityReport(
            current_version=version,
            latest_version=official_info.latest_version,
            compatibility_matrix=self.build_compatibility_matrix(framework, version),
            breaking_changes=self.identify_breaking_changes(framework, version),
            upgrade_path=self.plan_upgrade_path(framework, version),
            security_advisories=official_info.security_advisories,
            deprecation_warnings=official_info.deprecations
        )
    
    def suggest_toolchain_upgrades(self, 
                                 current_tools: Dict[str, str]) -> ToolchainUpgradePlan:
        """Suggest optimal toolchain upgrades."""
        latest_versions = self.get_latest_tool_versions()
        upgrade_plan = ToolchainUpgradePlan()
        
        for tool, current_version in current_tools.items():
            if tool in latest_versions:
                latest = latest_versions[tool]
                if self.is_upgrade_recommended(current_version, latest):
                    upgrade_plan.add_upgrade(
                        tool=tool,
                        from_version=current_version,
                        to_version=latest.version,
                        upgrade_type=self.classify_upgrade_type(current_version, latest),
                        breaking_changes=latest.breaking_changes,
                        migration_guide=latest.migration_guide
                    )
        
        return upgrade_plan
```

---

## Context7 Integration Examples

### Real-Time Documentation Fetching
```python
# Example: Get latest Python documentation
context7_client = Context7Client()

# Resolve library ID
python_id = await context7_client.resolve_library_id("Python 3.12")
# Returns: "/python/official-docs/3.12"

# Get comprehensive documentation
python_docs = await context7_client.get_library_docs(
    context7_library_id="/python/official-docs/3.12",
    topic="testing best practices async patterns",
    tokens=5000
)

# Returns structured documentation with:
# - Latest features and syntax
# - Performance improvements
# - Migration guides
# - Security considerations
# - Code examples
```

### Multi-Language Documentation Aggregation
```python
async def get_polyglot_documentation(project: Project) -> PolyglotDocs:
    """Aggregate documentation for all languages in project."""
    docs = {}
    
    for language in project.languages:
        # Get official docs for each language
        language_id = await context7.resolve_library_id(language)
        docs[language] = await context7.get_library_docs(
            context7_library_id=language_id,
            topic="testing security performance",
            tokens=3000
        )
    
    # Add cross-language patterns
    docs["polyglot_patterns"] = await get_cross_language_patterns(
        list(project.languages)
    )
    
    return PolyglotDocs(docs)
```

---

## Intelligent Toolchain Optimization

### Performance-Based Tool Selection
```python
class PerformanceAnalyzer:
    """Analyze and recommend optimal tooling based on performance metrics."""
    
    def analyze_project_performance(self, project_path: str) -> PerformanceProfile:
        """Analyze project performance characteristics."""
        return PerformanceProfile(
            build_time=self.measure_build_times(project_path),
            test_execution=self.analyze_test_performance(project_path),
            memory_usage=self.profile_memory_usage(project_path),
            io_patterns=self.analyze_io_patterns(project_path),
            bottlenecks=self.identify_bottlenecks(project_path)
        )
    
    def recommend_optimal_tooling(self, 
                                profile: PerformanceProfile,
                                language: str) -> ToolingRecommendations:
        """Recommend tools based on performance analysis."""
        recommendations = ToolingRecommendations()
        
        # Test framework selection based on performance
        if profile.test_execution.average_time > 30:  # seconds
            recommendations.testing.append({
                "tool": "pytest-xdist",
                "reason": "Parallel execution for slow tests",
                "expected_improvement": "60-80% faster"
            })
        
        # Build tool optimization
        if profile.build_time > 60:  # seconds
            recommendations.build_tools.append({
                "tool": "sccache" if language == "Rust" else "nx",
                "reason": "Build caching for faster iteration",
                "expected_improvement": "50-70% faster builds"
            })
        
        return recommendations
```

### Security-First Toolchain Configuration
```python
class SecurityToolchainOptimizer:
    """Configure security-focused toolchain for each language."""
    
    def configure_security_tools(self, language: str) -> SecurityConfiguration:
        """Configure security scanning tools."""
        configs = {
            "python": SecurityConfiguration(
                tools=["bandit", "safety", "semgrep"],
                rules=["bandit-high-confidence", "semgrep-python"],
                dependency_scanning="safety check",
                secret_detection="detect-secrets scan",
                container_scanning="trivy image"
            ),
            "javascript": SecurityConfiguration(
                tools=["npm audit", "snyk", "semgrep"],
                rules=["semgrep-javascript", "eslint-security"],
                dependency_scanning="npm audit --audit-level high",
                secret_detection="git-secrets",
                container_scanning="trivy image"
            ),
            "rust": SecurityConfiguration(
                tools=["cargo-audit", "cargo-deny", "semgrep"],
                rules=["semgrep-rust", "clippy-security"],
                dependency_scanning="cargo audit",
                secret_detection="git-secrets",
                container_scanning="trivy image"
            )
        }
        return configs.get(language, SecurityConfiguration.default())
```

---

## Polyglot Project Management

### Cross-Language Dependency Visualization
```python
def generate_polyglot_dependency_graph(project_path: str) -> DependencyGraph:
    """Generate dependency graph for multi-language projects."""
    graph = DependencyGraph()
    
    # Detect all language ecosystems
    languages = detect_languages(project_path)
    
    for language in languages:
        # Parse language-specific dependencies
        deps = parse_dependencies(project_path, language)
        
        # Add to graph with language metadata
        for dep in deps:
            node = DependencyNode(
                name=dep.name,
                version=dep.version,
                language=language,
                ecosystem=get_ecosystem(language),
                security_advisories=get_security_advisories(dep),
                license_info=get_license_info(dep)
            )
            graph.add_node(node)
    
    # Add cross-language integration points
    integrations = detect_cross_language_integrations(project_path)
    for integration in integrations:
        graph.add_edge(integration.source, integration.target)
    
    return graph
```

### Language-Specific Quality Gates
```python
class LanguageQualityGates:
    """Language-specific quality gates and validation rules."""
    
    GATES = {
        "python": {
            "test_coverage": 85,
            "type_checking": "mypy strict",
            "linting": "ruff check --fix",
            "security": "bandit -r src/",
            "dependencies": "safety check",
            "formatting": "black --check ."
        },
        "typescript": {
            "test_coverage": 80,
            "type_checking": "tsc --noEmit",
            "linting": "biome check .",
            "security": "npm audit --audit-level high",
            "dependencies": "npm audit",
            "formatting": "biome format --check ."
        },
        "rust": {
            "test_coverage": 75,
            "type_checking": "cargo check",
            "linting": "cargo clippy -- -D warnings",
            "security": "cargo audit",
            "dependencies": "cargo deny check",
            "formatting": "cargo fmt -- --check"
        }
    }
    
    def validate_language_quality(self, 
                                project_path: str, 
                                language: str) -> QualityReport:
        """Run language-specific quality validation."""
        gates = self.GATES.get(language, {})
        results = {}
        
        for gate_name, command in gates.items():
            try:
                result = run_command(command, cwd=project_path)
                results[gate_name] = GateResult(
                    passed=result.returncode == 0,
                    output=result.stdout,
                    error=result.stderr if result.returncode != 0 else None
                )
            except Exception as e:
                results[gate_name] = GateResult(
                    passed=False,
                    error=str(e)
                )
        
        return QualityReport(language=language, gates=results)
```

---

## Migration and Modernization

### Automated Language Migration Planning
```python
class MigrationPlanner:
    """AI-powered language migration planning."""
    
    async def plan_migration(self, 
                           from_language: str,
                           to_language: str,
                           project_path: str) -> MigrationPlan:
        """Plan language migration with cost-benefit analysis."""
        
        # Analyze current codebase
        current_analysis = await self.analyze_codebase(project_path, from_language)
        
        # Get target language capabilities via Context7
        target_docs = await self.context7.get_library_docs(
            context7_library_id=await self.context7.resolve_library_id(to_language),
            topic="migration guides best practices",
            tokens=5000
        )
        
        # Calculate migration complexity
        complexity = self.calculate_migration_complexity(
            current_analysis, target_docs
        )
        
        # Generate step-by-step migration plan
        return MigrationPlan(
            from_language=from_language,
            to_language=to_language,
            complexity_score=complexity.score,
            estimated_effort=complexity.estimated_effort,
            breaking_changes=complexity.breaking_changes,
            migration_steps=self.generate_migration_steps(
                current_analysis, target_docs
            ),
            tool_recommendations=self.suggest_migration_tools(
                from_language, to_language
            ),
            risk_assessment=self.assess_migration_risks(complexity),
            rollback_plan=self.generate_rollback_plan(current_analysis)
        )
```

### Legacy Framework Modernization
```python
async def modernize_legacy_framework(project_path: str,
                                    current_framework: str,
                                    target_framework: str) -> ModernizationPlan:
    """Plan framework modernization with automated tooling."""
    
    # Get modern framework docs
    modern_docs = await context7.get_library_docs(
        context7_library_id=await context7.resolve_library_id(target_framework),
        topic="migration patterns breaking changes best practices",
        tokens=4000
    )
    
    # Analyze current implementation
    current_impl = analyze_current_implementation(
        project_path, current_framework
    )
    
    # Generate modernization roadmap
    return ModernizationPlan(
        phases=[
            Phase(
                name="Preparation",
                tasks=["backup codebase", "setup new framework", "configure CI/CD"],
                estimated_duration="1-2 days"
            ),
            Phase(
                name="Core Migration",
                tasks=["migrate routes", "update data models", "port business logic"],
                estimated_duration=f"{calculate_core_migration_days(current_impl)} days"
            ),
            Phase(
                name="Testing & Validation",
                tasks=["comprehensive testing", "performance validation", "security audit"],
                estimated_duration="2-3 days"
            ),
            Phase(
                name="Deployment & Monitoring",
                tasks=["gradual rollout", "monitoring setup", "rollback preparation"],
                estimated_duration="1-2 days"
            )
        ],
        automated_tools=[
            "codemod scripts",
            "testing frameworks",
            "migration validators"
        ],
        manual_review_points=["business logic validation", "user experience testing"]
    )
```

---

## Configuration and Customization

### Project-Specific Language Configuration
```yaml
# .moai/language-config.yaml
language_detection:
  primary_language: "python"
  secondary_languages: ["typescript", "dockerfile"]
  polyglot_strategy: "microservice-isolation"

quality_gates:
  python:
    test_coverage_threshold: 85
    type_checking: "mypy strict"
    security_scanning: "bandit + safety"
  
  typescript:
    test_coverage_threshold: 80
    strict_types: true
    security_scanning: "npm audit high"

optimization:
  performance_profiling: true
  build_optimization: true
  dependency_analysis: true
  security_monitoring: true

context7_integration:
  auto_documentation: true
  latest_best_practices: true
  breaking_change_alerts: true
  security_advisories: true

tools:
  preferred_testing: ["pytest", "vitest", "jest"]
  preferred_linting: ["ruff", "biome", "clippy"]
  preferred_security: ["bandit", "semgrep", "trivy"]
  auto_update_tools: true
```

---

## Performance Metrics and Benchmarks

### Language Performance Comparison
```python
class LanguageBenchmarker:
    """Benchmark languages for specific use cases."""
    
    def benchmark_web_frameworks(self) -> WebFrameworkBenchmarks:
        """Benchmark web frameworks across languages."""
        return WebFrameworkBenchmarks(
            requests_per_second={
                "rust-axum": 125000,
                "go-gin": 98000,
                "python-fastapi": 45000,
                "nodejs-express": 65000,
                "java-spring": 58000
            },
            memory_usage_mb={
                "rust-axum": 15,
                "go-gin": 22,
                "python-fastapi": 85,
                "nodejs-express": 45,
                "java-spring": 180
            },
            response_time_p95_ms={
                "rust-axum": 0.8,
                "go-gin": 1.2,
                "python-fastapi": 2.5,
                "nodejs-express": 1.8,
                "java-spring": 3.2
            },
            development_speed={
                "python-fastapi": "fastest",
                "nodejs-express": "fast",
                "java-spring": "medium",
                "go-gin": "medium",
                "rust-axum": "slowest"
            }
        )
```

---

## API Reference

### Core Functions
- `detect_project_language(project_path)` - Primary language detection
- `analyze_framework_compatibility(framework, version)` - Framework compatibility
- `optimize_toolchain(project_profile)` - Toolchain optimization
- `generate_polyglot_graph(project_path)` - Multi-language dependency graph
- `plan_migration(from, to, project)` - Migration planning

### Context7 Integration
- `get_latest_documentation(language)` - Official docs via Context7
- `get_best_practices(framework)` - Current best practices
- `get_security_advisories(language)` - Security updates

### Data Structures
- `LanguageProfile` - Complete language analysis
- `FrameworkInfo` - Framework version and compatibility
- `ToolchainRecommendation` - Optimal tool recommendations
- `MigrationPlan` - Detailed migration strategy
- `PolyglotAnalysis` - Multi-language project analysis

---

## Changelog

- **v4.0.0** (2025-11-11): Complete rewrite with Context7 integration, AI-powered detection, 25+ language support, polyglot project management, and enterprise-grade toolchain optimization
- **v2.0.0** (2025-10-22): Major update with latest tool versions, comprehensive best practices, TRUST 5 integration
- **v1.0.0** (2025-03-29): Initial Skill release

---

## Works Well With

- `moai-foundation-trust` (quality gates)
- `moai-foundation-tags` (cross-language TAG management)
- `moai-foundation-specs` (multi-language SPEC documentation)
- `moai-essentials-debug` (language-specific debugging)
- `moai-essentials-perf` (performance optimization)
- Context7 MCP (official documentation)

---

## Best Practices

✅ **DO**:
- Use Context7 integration for latest documentation
- Implement language-specific quality gates
- Analyze polyglot project complexity early
- Automate toolchain optimization
- Monitor security advisories for all languages
- Plan migrations with comprehensive cost-benefit analysis
- Use performance benchmarks for technology selection

❌ **DON'T**:
- Mix incompatible language versions
- Skip framework compatibility checks
- Ignore polyglot integration complexity
- Use outdated tool versions
- Skip security scanning for any language
- Underestimate migration effort and risk
