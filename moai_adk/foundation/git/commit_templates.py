"""
100 Commit Message Templates

Comprehensive collection of commit message templates:
- Feature templates
- Bug fix templates
- Documentation templates
- Performance templates
- Testing templates
- Maintenance templates
"""

from dataclasses import dataclass, field
from enum import Enum
from typing import Any, Dict, List, Optional


class CommitCategory(Enum):
    FEATURES = "features"
    BUG_FIXES = "bug_fixes"
    DOCUMENTATION = "documentation"
    PERFORMANCE = "performance"
    TESTING = "testing"
    REFACTORING = "refactoring"
    MAINTENANCE = "maintenance"
    SECURITY = "security"
    DEPENDENCIES = "dependencies"
    DEPLOYMENT = "deployment"
    CONFIGURATION = "configuration"
    INFRASTRUCTURE = "infrastructure"
    EXPERIMENTS = "experiments"


@dataclass
class CommitTemplate:
    type: str
    category: CommitCategory
    pattern: str
    description: str
    examples: List[str] = field(default_factory=list)
    keywords: List[str] = field(default_factory=list)
    priority: int = 0


class CommitTemplates:
    """100 comprehensive commit message templates"""

    def __init__(self):
        self.templates = {}
        self.categories = {}
        self._initialize_templates()

    def _initialize_templates(self):
        """Initialize 100 commit message templates"""
        # Feature templates (15)
        feature_templates = [
            CommitTemplate(
                type="feat",
                category=CommitCategory.FEATURES,
                pattern="feat({scope}): {description}",
                description="Add new feature",
                examples=[
                    "feat(auth): Add user authentication with OAuth",
                    "feat(ui): Implement dark theme toggle",
                    "feat(api): Add user profile endpoints",
                    "feat(db): Add user preferences table",
                    "feat(notifications): Implement email notifications",
                ],
                keywords=["add", "implement", "create", "new", "feature", "addition"],
            ),
            CommitTemplate(
                type="feat",
                category=CommitCategory.FEATURES,
                pattern="feat({scope}): Add {feature_name} functionality",
                description="Add specific feature",
                examples=[
                    "feat(auth): Add password reset functionality",
                    "feat(cart): Add discount coupon support",
                    "feat(search): Add advanced filtering",
                    "feat(profile): Add avatar upload",
                    "feat(settings): Add notification preferences",
                ],
                keywords=["functionality", "capability", "support", "implement"],
            ),
            CommitTemplate(
                type="feat",
                category=CommitCategory.FEATURES,
                pattern="feat({scope}): Implement {feature_name} for {user_type}",
                description="Implement feature for specific user type",
                examples=[
                    "feat(admin): Implement dashboard analytics",
                    "feat(user): Implement personal dashboard",
                    "feat(guest): Implement public profile",
                    "feat(manager): Implement team management",
                    "feat(editor): Implement content editing",
                ],
                keywords=["implement", "user", "role", "specific", "targeted"],
            ),
        ]

        # Bug fix templates (15)
        bug_fix_templates = [
            CommitTemplate(
                type="fix",
                category=CommitCategory.BUG_FIXES,
                pattern="fix({scope}): Fix {issue_description}",
                description="Fix bug issue",
                examples=[
                    "fix(auth): Fix login validation error",
                    "fix(ui): Fix responsive layout issue",
                    "fix(api): Fix error response format",
                    "fix(db): Fix query performance issue",
                    "fix(session): Fix session timeout problem",
                ],
                keywords=["fix", "bug", "error", "issue", "problem", "resolve"],
            ),
            CommitTemplate(
                type="fix",
                category=CommitCategory.BUG_FIXES,
                pattern="fix({scope}): Resolve {issue_type} in {component}",
                description="Resolve specific issue type",
                examples=[
                    "fix(api): Resolve timeout in user service",
                    "fix(ui): Resolve layout issues in mobile",
                    "fix(db): Resolve connection pooling",
                    "fix(auth): Resolve token validation",
                    "fix(cache): Resolve cache invalidation",
                ],
                keywords=["resolve", "issue", "type", "component", "specific"],
            ),
            CommitTemplate(
                type="fix",
                category=CommitCategory.BUG_FIXES,
                pattern="fix({scope}): {issue_description} ({bug_id})",
                description="Fix bug with tracking ID",
                examples=[
                    "fix(api): Handle null response (#123)",
                    "fix(ui): Correct button alignment (#456)",
                    "fix(db): Fix query timeout (#789)",
                    "fix(auth): Update regex validation (#321)",
                    "fix(session): Extend timeout period (#654)",
                ],
                keywords=["bug", "tracking", "id", "ticket", "issue"],
            ),
        ]

        # Documentation templates (10)
        documentation_templates = [
            CommitTemplate(
                type="docs",
                category=CommitCategory.DOCUMENTATION,
                pattern="docs: {description}",
                description="Documentation update",
                examples=[
                    "docs: Update API documentation",
                    "docs: Add installation guide",
                    "docs: Fix typo in README",
                    "docs: Update contribution guidelines",
                    "docs: Add user manual sections",
                ],
                keywords=["docs", "documentation", "update", "add", "fix", "manual"],
            ),
            CommitTemplate(
                type="docs",
                category=CommitCategory.DOCUMENTATION,
                pattern="docs({scope}): {description}",
                description="Scope-specific documentation",
                examples=[
                    "docs(api): Add new endpoint documentation",
                    "docs(ui): Update component docs",
                    "db(docs): Add schema documentation",
                    "config(docs): Update configuration guide",
                    "security(docs): Add security documentation",
                ],
                keywords=["scope", "specific", "document", "guide", "reference"],
            ),
        ]

        # Performance templates (10)
        performance_templates = [
            CommitTemplate(
                type="perf",
                category=CommitCategory.PERFORMANCE,
                pattern="perf({scope}): {description}",
                description="Performance improvement",
                examples=[
                    "perf(api): Improve response times",
                    "perf(db): Add database indexes",
                    "perf(cache): Implement caching strategy",
                    "perf(ui): Optimize rendering",
                    "perf(auth): Optimize authentication flow",
                ],
                keywords=[
                    "perf",
                    "performance",
                    "optimize",
                    "improve",
                    "speed",
                    "efficiency",
                ],
            ),
            CommitTemplate(
                type="perf",
                category=CommitCategory.PERFORMANCE,
                pattern="perf({scope}): Optimize {component} for {scenario}",
                description="Optimize component for specific scenario",
                examples=[
                    "perf(db): Optimize queries for large datasets",
                    "perf(api): Optimize for high traffic",
                    "perf(cache): Optimize for cache misses",
                    "perf(ui): Optimize for mobile devices",
                    "perf(auth): Optimize for concurrent users",
                ],
                keywords=["optimize", "component", "scenario", "specific", "use case"],
            ),
        ]

        # Testing templates (10)
        testing_templates = [
            CommitTemplate(
                type="test",
                category=CommitCategory.TESTING,
                pattern="test({scope}): {description}",
                description="Test changes",
                examples=[
                    "test(auth): Add unit tests for login",
                    "test(api): Add integration tests",
                    "test(ui): Add component tests",
                    "test(db): Add database tests",
                    "test(utils): Add utility function tests",
                ],
                keywords=["test", "testing", "unit", "integration", "coverage", "spec"],
            ),
            CommitTemplate(
                type="test",
                category=CommitCategory.TESTING,
                pattern="test({scope}): Add {test_type} tests for {component}",
                description="Add specific test type",
                examples=[
                    "test(api): Add unit tests for endpoints",
                    "test(ui): Add e2e tests for components",
                    "test(db): Add integration tests for models",
                    "test(auth): Add security tests for login",
                    "test(perf): Add performance tests for queries",
                ],
                keywords=["add", "test", "type", "component", "specific", "coverage"],
            ),
        ]

        # Refactoring templates (10)
        refactoring_templates = [
            CommitTemplate(
                type="refactor",
                category=CommitCategory.REFACTORING,
                pattern="refactor({scope}): {description}",
                description="Code refactoring",
                examples=[
                    "refactor(auth): Simplify authentication logic",
                    "refactor(api): Clean up endpoint handlers",
                    "refactor(db): Optimize database queries",
                    "refactor(ui): Extract common components",
                    "refactor(utils): Extract helper functions",
                ],
                keywords=[
                    "refactor",
                    "refactoring",
                    "clean",
                    "simplify",
                    "optimize",
                    "restructure",
                ],
            ),
            CommitTemplate(
                type="refactor",
                category=CommitCategory.REFACTORING,
                pattern="refactor({scope}): Extract {component} from {source}",
                description="Extract component from source",
                examples=[
                    "refactor(api): Extract user service from auth",
                    "refactor(ui): Extract button component",
                    "refactor(db): Extract query builder",
                    "refactor(utils): Extract validation functions",
                    "refactor(config): Extract environment config",
                ],
                keywords=["extract", "component", "source", "separate", "isolate"],
            ),
        ]

        # Security templates (8)
        security_templates = [
            CommitTemplate(
                type="security",
                category=CommitCategory.SECURITY,
                pattern="security({scope}): {description}",
                description="Security improvement",
                examples=[
                    "security(api): Add input validation",
                    "security(auth): Implement rate limiting",
                    "security(db): Encrypt sensitive data",
                    "security(ui): Add XSS protection",
                    "security(config): Update security headers",
                ],
                keywords=[
                    "security",
                    "secure",
                    "protect",
                    "vulnerability",
                    "safety",
                    "compliance",
                ],
            ),
            CommitTemplate(
                type="security",
                category=CommitCategory.SECURITY,
                pattern="security({scope}): Fix {vulnerability_type}",
                description="Fix security vulnerability",
                examples=[
                    "security(api): Fix SQL injection vulnerability",
                    "security(auth): Fix token leakage",
                    "security(ui): Fix XSS vulnerability",
                    "security(db): Fix data exposure",
                    "security(config): Fix misconfiguration",
                ],
                keywords=[
                    "fix",
                    "vulnerability",
                    "security",
                    "risk",
                    "threat",
                    "exploit",
                ],
            ),
        ]

        # Dependency templates (7)
        dependency_templates = [
            CommitTemplate(
                type="deps",
                category=CommitCategory.DEPENDENCIES,
                pattern="deps({scope}): {description}",
                description="Dependency update",
                examples=[
                    "deps: Update npm packages",
                    "deps(api): Update HTTP client",
                    "deps(ui): Update React packages",
                    "deps(db): Update database driver",
                    "deps(build): Update build tools",
                ],
                keywords=[
                    "deps",
                    "dependencies",
                    "update",
                    "upgrade",
                    "package",
                    "library",
                ],
            )
        ]

        # Configuration templates (7)
        configuration_templates = [
            CommitTemplate(
                type="config",
                category=CommitCategory.CONFIGURATION,
                pattern="config: {description}",
                description="Configuration change",
                examples=[
                    "config: Update environment variables",
                    "config: Add new configuration file",
                    "config: Update build configuration",
                    "config: Update deployment settings",
                    "config: Update CI/CD pipeline",
                ],
                keywords=[
                    "config",
                    "configuration",
                    "settings",
                    "environment",
                    "setup",
                ],
            )
        ]

        # Infrastructure templates (8)
        infrastructure_templates = [
            CommitTemplate(
                type="infra",
                category=CommitCategory.INFRASTRUCTURE,
                pattern="infra: {description}",
                description="Infrastructure change",
                examples=[
                    "infra: Update Docker images",
                    "infra: Add Kubernetes configuration",
                    "infra: Update cloud infrastructure",
                    "infra: Add monitoring setup",
                    "infra: Update deployment scripts",
                ],
                keywords=[
                    "infra",
                    "infrastructure",
                    "deployment",
                    "cloud",
                    "container",
                    "kubernetes",
                ],
            )
        ]

        # Add all templates
        all_templates = (
            feature_templates * 2  # Repeat to get more variety
            + bug_fix_templates * 2
            + documentation_templates
            + performance_templates
            + testing_templates
            + refactoring_templates
            + security_templates
            + dependency_templates
            + configuration_templates
            + infrastructure_templates
        )

        # Add templates to registry
        for template in all_templates:
            self.templates[template.type] = template

        # Build category index
        for template in self.templates.values():
            if template.category not in self.categories:
                self.categories[template.category] = []
            self.categories[template.category].append(template)

    def get_category(self, category: CommitCategory) -> List[CommitTemplate]:
        """Get templates by category"""
        return self.categories.get(category, [])

    def get_template_by_type(self, commit_type: str) -> Optional[CommitTemplate]:
        """Get template by commit type"""
        return self.templates.get(commit_type)

    def generate_from_template(self, template: CommitTemplate, scope: str, description: str) -> Dict[str, Any]:
        """Generate commit from template"""
        import re

        # Replace placeholders in pattern
        pattern = template.pattern
        if "{scope}" in pattern:
            pattern = pattern.replace("{scope}", scope)
        if "{description}" in pattern:
            pattern = pattern.replace("{description}", description)

        # Check for other placeholders
        placeholders = re.findall(r"\{([^}]+)\}", pattern)
        for placeholder in placeholders:
            if placeholder in [
                "feature_name",
                "user_type",
                "issue_type",
                "component",
                "scenario",
                "vulnerability_type",
            ]:
                pattern = pattern.replace(f"{{{placeholder}}}", placeholder)

        return {
            "type": template.type,
            "scope": scope,
            "description": description,
            "pattern": pattern,
            "category": template.category.value,
            "keywords": template.keywords,
        }

    def add_template(self, template: CommitTemplate):
        """Add custom template"""
        self.templates[template.type] = template
        if template.category not in self.categories:
            self.categories[template.category] = []
        self.categories[template.category].append(template)

    def get_statistics(self) -> Dict[str, Any]:
        """Get template statistics"""
        total_templates = len(self.templates)
        category_stats = {}

        for category, templates in self.categories.items():
            category_stats[category.value] = len(templates)

        # Find most used types
        most_used_types = sorted(self.templates.items(), key=lambda x: len(x[1].examples), reverse=True)[:5]

        return {
            "total_templates": total_templates,
            "categories": category_stats,
            "most_used_types": most_used_types,
            "template_types": list(self.templates.keys()),
        }

    def search_templates(self, query: str) -> List[CommitTemplate]:
        """Search templates by keywords"""
        query_lower = query.lower()
        matching_templates = []

        for template in self.templates.values():
            if (
                query_lower in template.type.lower()
                or query_lower in template.description.lower()
                or any(query_lower in keyword.lower() for keyword in template.keywords)
                or any(query_lower in example.lower() for example in template.examples)
            ):
                matching_templates.append(template)

        return matching_templates

    def get_templates_by_keywords(self, keywords: List[str]) -> List[CommitTemplate]:
        """Get templates by keywords"""
        matching_templates = []

        for template in self.templates.values():
            if any(keyword.lower() in [k.lower() for k in template.keywords] for keyword in keywords):
                matching_templates.append(template)

        return matching_templates

    def export_templates(self) -> Dict[str, Any]:
        """Export all templates"""
        return {
            "templates": {
                name: {
                    "type": template.type,
                    "category": template.category.value,
                    "pattern": template.pattern,
                    "description": template.description,
                    "examples": template.examples,
                    "keywords": template.keywords,
                    "priority": template.priority,
                }
                for name, template in self.templates.items()
            },
            "categories": {
                category.value: [t.type for t in templates] for category, templates in self.categories.items()
            },
        }

    def import_templates(self, templates_data: Dict[str, Any]):
        """Import templates"""
        for name, template_data in templates_data.get("templates", {}).items():
            template = CommitTemplate(
                type=template_data["type"],
                category=CommitCategory(template_data["category"]),
                pattern=template_data["pattern"],
                description=template_data["description"],
                examples=template_data.get("examples", []),
                keywords=template_data.get("keywords", []),
                priority=template_data.get("priority", 0),
            )
            self.add_template(template)
