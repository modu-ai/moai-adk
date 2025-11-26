"""
TRUST 4 Principles Implementation

Comprehensive implementation of TRUST principles for MoAI-ADK:
- Test First: Comprehensive testing strategy
- Readable: Code clarity and maintainability
- Unified: Consistent architecture and patterns
- Secured: Security best practices and validation

Features:
- Automated principle scoring
- Validation checklists
- Quality gate enforcement
- Enterprise-grade security
- Complete audit trails
"""

import re
from dataclasses import dataclass, field
from enum import Enum
from pathlib import Path
from typing import Any, Dict, List


class TrustPrinciple(Enum):
    """TRUST 4 principles enumeration"""

    TEST_FIRST = "test_first"
    READABLE = "readable"
    UNIFIED = "unified"
    SECURED = "secured"


class ComplianceLevel(Enum):
    """Compliance level enumeration"""

    CRITICAL = "critical"
    HIGH = "high"
    MEDIUM = "medium"
    LOW = "low"
    NONE = "none"


@dataclass
class PrincipleScore:
    """Individual principle scoring"""

    principle: TrustPrinciple
    score: float  # 0.0 - 100.0
    compliance_level: ComplianceLevel
    issues: List[str] = field(default_factory=list)
    recommendations: List[str] = field(default_factory=list)
    metrics: Dict[str, Any] = field(default_factory=dict)


@dataclass
class TrustAssessment:
    """Complete TRUST assessment"""

    principle_scores: Dict[TrustPrinciple, PrincipleScore]
    overall_score: float
    compliance_level: ComplianceLevel
    passed_checks: int
    total_checks: int
    audit_trail: List[Dict[str, Any]] = field(default_factory=list)
    timestamp: str = field(default_factory=lambda: "2025-11-11")


class TrustPrinciplesValidator:
    """TRUST 4 principles validator and analyzer"""

    def __init__(self):
        self.principle_weights = {
            TrustPrinciple.TEST_FIRST: 0.30,
            TrustPrinciple.READABLE: 0.25,
            TrustPrinciple.UNIFIED: 0.25,
            TrustPrinciple.SECURED: 0.20,
        }

        # Test First validation patterns
        self.test_patterns = {
            "unit_tests": r"def test_.*\(|class Test.*:",
            "integration_tests": r"@integration_test|test_integration_",
            "coverage_directive": r"# pragma: cover|@pytest\.mark\.cover",
            "assertion_count": r"\bassert\b.*>=|\bshould_\w+\(",
            "test_docstrings": r'def test_\w+.*\n\s*""".*?"""',
        }

        # Readability validation patterns
        self.readability_patterns = {
            "function_length": r"def \w+\([^)]*\):(.*\n){1,50}",
            "class_length": r"class \w+:",
            "variable_naming": r"\b[a-z_][a-z0-9_]*\b",
            "constant_naming": r"\b[A-Z_][A-Z0-9_]*\b",
            "docstrings": r'def \w+.*\n\s*""".*?"""',
            "type_hints": r": \w+\[?\]?]|: Optional\[|: Dict\[|: List\[",
        }

        # Unified validation patterns
        self.unified_patterns = {
            "import_structure": r"^(import|from)\s+\w+",
            "naming_convention": r"\b\w+(?:Error|Exception|Manager|Service|Handler|Validator)\b",
            "file_structure": r"^(class|def)\s+",
            "error_handling": r"except\s+\w+:|raise\s+\w+\(",
            "logging_pattern": r"logger\.\w+|logging\.\w+",
        }

        # Security validation patterns
        self.security_patterns = {
            "sql_injection": r'(cursor\.execute|db\.query)\s*\(\s*["\'][^"\']*%[^"\']*["\']',
            "xss_prevention": r"escape\(|sanitize\(|validate\(",
            "auth_check": r"@login_required|@permission_required|is_authenticated",
            "input_validation": r"re\.match\(|validator\.|form\.is_valid\(\)",
            "secret_management": r"\.env|SECRET_KEY|API_KEY|PASSWORD",
            "https_enforcement": r"https://|SECURE_SSL_REDIRECT|HSTS",
        }

        # Trackability validation patterns
        self.trackability_patterns = {
            "commit_messages": r"^(feat|fix|docs|style|refactor|test|chop)\(",
            "issue_references": r"#\d+|closes #\d+|fixes #\d+",
            "documentation_links": r"\[.*?\]\(.*?\.md\)",
            "version_tracking": r"v\d+\.\d+\.\d+|\d+\.\d+\.\d+",
        }

    def validate_test_first(self, project_path: str) -> PrincipleScore:
        """Validate Test First principle"""
        issues = []
        recommendations = []
        metrics = {}

        try:
            project_dir = Path(project_path)
            python_files = list(project_dir.rglob("*.py"))

            test_files = [f for f in python_files if f.name.startswith("test_")]
            source_files = [f for f in python_files if not f.name.startswith("test_")]

            # Test file ratio
            test_ratio = len(test_files) / max(len(source_files), 1)

            # Analyze test coverage indicators
            total_test_patterns = 0
            found_test_patterns = 0

            for file_path in python_files:
                try:
                    content = file_path.read_text(encoding="utf-8")

                    for pattern_name, pattern in self.test_patterns.items():
                        total_test_patterns += 1
                        if re.search(pattern, content, re.MULTILINE | re.DOTALL):
                            found_test_patterns += 1
                except Exception as e:
                    issues.append(f"Error analyzing {file_path}: {str(e)}")

            pattern_coverage = found_test_patterns / max(total_test_patterns, 1)

            # Calculate score
            score = min(100, (test_ratio * 30) + (pattern_coverage * 70))

            metrics = {
                "test_files": len(test_files),
                "source_files": len(source_files),
                "test_ratio": round(test_ratio, 2),
                "pattern_coverage": round(pattern_coverage, 2),
                "total_patterns": total_test_patterns,
                "found_patterns": found_test_patterns,
            }

            # Generate recommendations
            if test_ratio < 0.5:
                recommendations.append("Increase test file coverage (aim for 1:1 ratio with source files)")

            if pattern_coverage < 0.7:
                recommendations.append("Add comprehensive test patterns and assertions")

            if not any("integration" in str(issue) for issue in issues):
                recommendations.append("Add integration tests for critical workflows")

            # Determine compliance level
            if score >= 90:
                compliance = ComplianceLevel.CRITICAL
            elif score >= 75:
                compliance = ComplianceLevel.HIGH
            elif score >= 60:
                compliance = ComplianceLevel.MEDIUM
            elif score >= 40:
                compliance = ComplianceLevel.LOW
            else:
                compliance = ComplianceLevel.NONE

        except Exception as e:
            issues.append(f"Error validating Test First principle: {str(e)}")
            score = 0
            compliance = ComplianceLevel.NONE

        return PrincipleScore(
            principle=TrustPrinciple.TEST_FIRST,
            score=round(score, 2),
            compliance_level=compliance,
            issues=issues,
            recommendations=recommendations,
            metrics=metrics,
        )

    def validate_readable(self, project_path: str) -> PrincipleScore:
        """Validate Readable principle"""
        issues = []
        recommendations = []
        metrics = {}

        try:
            project_dir = Path(project_path)
            python_files = list(project_dir.rglob("*.py"))

            total_functions = 0
            long_functions = 0
            functions_with_docstrings = 0
            functions_with_type_hints = 0
            total_classes = 0
            classes_with_docstrings = 0

            for file_path in python_files:
                try:
                    content = file_path.read_text(encoding="utf-8")
                    content.split("\n")

                    # Analyze functions
                    function_matches = re.finditer(r"def\s+(\w+)\([^)]*\):", content)
                    for match in function_matches:
                        total_functions += 1

                        # Count lines in function
                        func_start = match.start()
                        func_lines = 0
                        indent_level = None

                        for i, line in enumerate(content[func_start:].split("\n")[1:], 1):
                            if line.strip() == "":
                                continue

                            current_indent = len(line) - len(line.lstrip())
                            if indent_level is None:
                                indent_level = current_indent
                            elif current_indent <= indent_level and line.strip():
                                break
                            func_lines += 1

                        if func_lines > 50:
                            long_functions += 1
                            issues.append(f"Long function in {file_path.name}: {match.group(1)} ({func_lines} lines)")

                        # Check for docstring
                        func_content = content[func_start : func_start + 1000]  # Check next 1000 chars
                        if '"""' in func_content.split("\n")[1:5]:
                            functions_with_docstrings += 1

                        # Check for type hints
                        if "->" in match.group(0) or ":" in match.group(0):
                            functions_with_type_hints += 1

                    # Analyze classes
                    class_matches = re.finditer(r"class\s+(\w+)[\(\:]:", content)
                    for match in class_matches:
                        total_classes += 1

                        # Check for docstring
                        class_start = match.start()
                        class_content = content[class_start : class_start + 500]  # Check next 500 chars
                        if '"""' in class_content.split("\n")[1:3]:
                            classes_with_docstrings += 1

                except Exception as e:
                    issues.append(f"Error analyzing {file_path}: {str(e)}")

            # Calculate scores
            docstring_ratio = functions_with_docstrings / max(total_functions, 1)
            type_hint_ratio = functions_with_type_hints / max(total_functions, 1)
            long_function_ratio = 1 - (long_functions / max(total_functions, 1))
            class_docstring_ratio = classes_with_docstrings / max(total_classes, 1)

            score = docstring_ratio * 30 + type_hint_ratio * 25 + long_function_ratio * 25 + class_docstring_ratio * 20

            metrics = {
                "total_functions": total_functions,
                "functions_with_docstrings": functions_with_docstrings,
                "functions_with_type_hints": functions_with_type_hints,
                "long_functions": long_functions,
                "total_classes": total_classes,
                "classes_with_docstrings": classes_with_docstrings,
                "docstring_ratio": round(docstring_ratio, 2),
                "type_hint_ratio": round(type_hint_ratio, 2),
            }

            # Generate recommendations
            if docstring_ratio < 0.8:
                recommendations.append("Add docstrings to all functions and methods")

            if type_hint_ratio < 0.8:
                recommendations.append("Add type hints for better code clarity")

            if long_function_ratio < 0.9:
                recommendations.append("Refactor long functions (keep under 50 lines)")

            # Determine compliance level
            if score >= 90:
                compliance = ComplianceLevel.CRITICAL
            elif score >= 80:
                compliance = ComplianceLevel.HIGH
            elif score >= 70:
                compliance = ComplianceLevel.MEDIUM
            elif score >= 60:
                compliance = ComplianceLevel.LOW
            else:
                compliance = ComplianceLevel.NONE

        except Exception as e:
            issues.append(f"Error validating Readable principle: {str(e)}")
            score = 0
            compliance = ComplianceLevel.NONE

        return PrincipleScore(
            principle=TrustPrinciple.READABLE,
            score=round(score, 2),
            compliance_level=compliance,
            issues=issues,
            recommendations=recommendations,
            metrics=metrics,
        )

    def validate_unified(self, project_path: str) -> PrincipleScore:
        """Validate Unified principle"""
        issues = []
        recommendations = []
        metrics = {}

        try:
            project_dir = Path(project_path)
            python_files = list(project_dir.rglob("*.py"))

            unified_patterns_found = 0
            total_patterns = len(self.unified_patterns)
            file_count = len(python_files)

            naming_violations = 0
            error_handling_count = 0
            logging_count = 0

            for file_path in python_files:
                try:
                    content = file_path.read_text(encoding="utf-8")

                    # Check unified patterns
                    for pattern_name, pattern in self.unified_patterns.items():
                        if re.search(pattern, content, re.MULTILINE):
                            unified_patterns_found += 1

                    # Check naming violations
                    if re.search(r"class\s+[a-z]", content):  # Class starting with lowercase
                        naming_violations += 1

                    # Count error handling
                    error_matches = re.findall(r"except\s+\w+:", content)
                    error_handling_count += len(error_matches)

                    # Count logging usage
                    logging_matches = re.findall(r"logger\.\w+|logging\.\w+", content)
                    logging_count += len(logging_matches)

                except Exception as e:
                    issues.append(f"Error analyzing {file_path}: {str(e)}")

            # Calculate scores
            pattern_coverage = unified_patterns_found / (total_patterns * file_count)
            error_handling_ratio = error_handling_count / max(file_count, 1)
            logging_ratio = logging_count / max(file_count, 1)
            naming_quality = 1 - (naming_violations / max(file_count, 1))

            score = pattern_coverage * 40 + error_handling_ratio * 25 + logging_ratio * 20 + naming_quality * 15

            metrics = {
                "files_analyzed": file_count,
                "unified_patterns_found": unified_patterns_found,
                "pattern_coverage": round(pattern_coverage, 2),
                "error_handling_count": error_handling_count,
                "logging_count": logging_count,
                "naming_violations": naming_violations,
            }

            # Generate recommendations
            if pattern_coverage < 0.7:
                recommendations.append("Improve code structure consistency across files")

            if error_handling_count < file_count * 0.5:
                recommendations.append("Add comprehensive error handling")

            if logging_count < file_count * 0.3:
                recommendations.append("Add proper logging for better debugging")

            # Determine compliance level
            if score >= 85:
                compliance = ComplianceLevel.CRITICAL
            elif score >= 75:
                compliance = ComplianceLevel.HIGH
            elif score >= 65:
                compliance = ComplianceLevel.MEDIUM
            elif score >= 55:
                compliance = ComplianceLevel.LOW
            else:
                compliance = ComplianceLevel.NONE

        except Exception as e:
            issues.append(f"Error validating Unified principle: {str(e)}")
            score = 0
            compliance = ComplianceLevel.NONE

        return PrincipleScore(
            principle=TrustPrinciple.UNIFIED,
            score=round(score, 2),
            compliance_level=compliance,
            issues=issues,
            recommendations=recommendations,
            metrics=metrics,
        )

    def validate_secured(self, project_path: str) -> PrincipleScore:
        """Validate Secured principle"""
        issues = []
        recommendations = []
        metrics = {}

        try:
            project_dir = Path(project_path)
            python_files = list(project_dir.rglob("*.py"))

            security_issues = []
            security_patterns_found = 0
            high_risk_patterns = 0

            for file_path in python_files:
                try:
                    content = file_path.read_text(encoding="utf-8")

                    # Check for security patterns
                    for pattern_name, pattern in self.security_patterns.items():
                        matches = re.findall(pattern, content, re.MULTILINE | re.DOTALL)
                        if matches:
                            security_patterns_found += len(matches)

                            if pattern_name in ["sql_injection", "secret_management"]:
                                high_risk_patterns += len(matches)
                                for match in matches:
                                    security_issues.append(f"High-risk pattern in {file_path.name}: {pattern_name}")

                    # Check for hardcoded secrets (basic pattern)
                    secret_patterns = [
                        r'password\s*=\s*["\'][^"\']+["\']',
                        r'api_key\s*=\s*["\'][^"\']+["\']',
                        r'secret\s*=\s*["\'][^"\']+["\']',
                        r'token\s*=\s*["\'][^"\']+["\']',
                    ]

                    for pattern in secret_patterns:
                        if re.search(pattern, content, re.IGNORECASE):
                            security_issues.append(f"Potential hardcoded secret in {file_path.name}")
                            high_risk_patterns += 1

                except Exception as e:
                    issues.append(f"Error analyzing {file_path}: {str(e)}")

            # Calculate security score
            if high_risk_patterns > 0:
                base_score = max(0, 100 - (high_risk_patterns * 20))
            else:
                base_score = 100

            # Bonus points for security best practices
            security_bonus = 0
            if security_patterns_found > 0:
                security_bonus = min(20, security_patterns_found * 2)

            score = min(100, base_score + security_bonus)

            metrics = {
                "security_patterns_found": security_patterns_found,
                "high_risk_patterns": high_risk_patterns,
                "security_issues": len(security_issues),
                "files_analyzed": len(python_files),
            }

            # Generate recommendations
            if high_risk_patterns > 0:
                recommendations.append("Address high-risk security patterns immediately")
                issues.extend(security_issues[:5])  # Add first 5 issues

            if security_patterns_found < 10:
                recommendations.append("Implement more security validation patterns")

            recommendations.extend(
                [
                    "Use environment variables for secrets",
                    "Implement input validation and sanitization",
                    "Add authentication and authorization checks",
                ]
            )

            # Determine compliance level
            if score >= 95:
                compliance = ComplianceLevel.CRITICAL
            elif score >= 85:
                compliance = ComplianceLevel.HIGH
            elif score >= 75:
                compliance = ComplianceLevel.MEDIUM
            elif score >= 60:
                compliance = ComplianceLevel.LOW
            else:
                compliance = ComplianceLevel.NONE

        except Exception as e:
            issues.append(f"Error validating Secured principle: {str(e)}")
            score = 0
            compliance = ComplianceLevel.NONE

        return PrincipleScore(
            principle=TrustPrinciple.SECURED,
            score=round(score, 2),
            compliance_level=compliance,
            issues=issues,
            recommendations=recommendations,
            metrics=metrics,
        )

    def assess_project(self, project_path: str) -> TrustAssessment:
        """Perform complete TRUST assessment"""
        principle_scores = {}

        # Validate each principle
        principle_scores[TrustPrinciple.TEST_FIRST] = self.validate_test_first(project_path)
        principle_scores[TrustPrinciple.READABLE] = self.validate_readable(project_path)
        principle_scores[TrustPrinciple.UNIFIED] = self.validate_unified(project_path)
        principle_scores[TrustPrinciple.SECURED] = self.validate_secured(project_path)

        # Calculate overall score
        overall_score = 0.0
        for principle, score in principle_scores.items():
            weight = self.principle_weights[principle]
            overall_score += score.score * weight

        # Calculate passed checks
        total_checks = sum(len(score.metrics) for score in principle_scores.values())
        passed_checks = sum(
            sum(
                1
                for metric_value in score.metrics.values()
                if isinstance(metric_value, (int, float)) and metric_value > 0
            )
            for score in principle_scores.values()
        )

        # Determine overall compliance level
        score_levels = [score.compliance_level for score in principle_scores.values()]
        if all(level == ComplianceLevel.CRITICAL for level in score_levels):
            overall_compliance = ComplianceLevel.CRITICAL
        elif all(level in [ComplianceLevel.CRITICAL, ComplianceLevel.HIGH] for level in score_levels):
            overall_compliance = ComplianceLevel.HIGH
        elif all(
            level in [ComplianceLevel.CRITICAL, ComplianceLevel.HIGH, ComplianceLevel.MEDIUM] for level in score_levels
        ):
            overall_compliance = ComplianceLevel.MEDIUM
        elif all(level != ComplianceLevel.NONE for level in score_levels):
            overall_compliance = ComplianceLevel.LOW
        else:
            overall_compliance = ComplianceLevel.NONE

        # Create audit trail
        audit_trail = []
        for principle, score in principle_scores.items():
            audit_trail.append(
                {
                    "principle": principle.value,
                    "score": score.score,
                    "compliance_level": score.compliance_level.value,
                    "issues_count": len(score.issues),
                    "recommendations_count": len(score.recommendations),
                    "metrics": score.metrics,
                }
            )

        return TrustAssessment(
            principle_scores=principle_scores,
            overall_score=round(overall_score, 2),
            compliance_level=overall_compliance,
            passed_checks=passed_checks,
            total_checks=total_checks,
            audit_trail=audit_trail,
        )

    def generate_report(self, assessment: TrustAssessment) -> str:
        """Generate comprehensive TRUST assessment report"""
        report = []
        report.append("# TRUST 4 Principles Assessment Report")
        report.append(f"Generated: {assessment.timestamp}")
        report.append(f"Overall Score: {assessment.overall_score}/100")
        report.append(f"Compliance Level: {assessment.compliance_level.value.upper()}")
        report.append(f"Passed Checks: {assessment.passed_checks}/{assessment.total_checks}")
        report.append("")

        # Principle breakdown
        report.append("## Principle Breakdown")
        report.append("")

        for principle, score in assessment.principle_scores.items():
            report.append(f"### {principle.value.replace('_', ' ').title()}")
            report.append(f"**Score**: {score.score}/100")
            report.append(f"**Compliance**: {score.compliance_level.value.upper()}")

            if score.issues:
                report.append("**Issues**:")
                for issue in score.issues[:5]:  # Show first 5 issues
                    report.append(f"- {issue}")

            if score.recommendations:
                report.append("**Recommendations**:")
                for rec in score.recommendations[:5]:  # Show first 5 recommendations
                    report.append(f"- {rec}")

            if score.metrics:
                report.append("**Metrics**:")
                for key, value in score.metrics.items():
                    report.append(f"- {key}: {value}")

            report.append("")

        # Summary and next steps
        report.append("## Summary")
        report.append("")

        if assessment.overall_score >= 80:
            report.append("✅ **EXCELLENT**: Project meets TRUST principles at a high level")
        elif assessment.overall_score >= 70:
            report.append("🟡 **GOOD**: Project mostly follows TRUST principles with some areas for improvement")
        elif assessment.overall_score >= 60:
            report.append("🟠 **NEEDS IMPROVEMENT**: Project has significant gaps in TRUST principles")
        else:
            report.append("❌ **CRITICAL**: Project requires immediate attention to TRUST principles")

        report.append("")
        report.append("## Next Steps")
        report.append("")

        # Collect all recommendations
        all_recommendations = []
        for score in assessment.principle_scores.values():
            all_recommendations.extend(score.recommendations)

        if all_recommendations:
            report.append("### Priority Recommendations")
            for i, rec in enumerate(set(all_recommendations[:10]), 1):  # Top 10 unique recommendations
                report.append(f"{i}. {rec}")

        return "\n".join(report)


# Convenience functions for common use cases
def validate_project_trust(project_path: str = ".") -> TrustAssessment:
    """Quick TRUST validation for a project"""
    validator = TrustPrinciplesValidator()
    return validator.assess_project(project_path)


def generate_trust_report(project_path: str = ".") -> str:
    """Generate TRUST assessment report"""
    validator = TrustPrinciplesValidator()
    assessment = validator.assess_project(project_path)
    return validator.generate_report(assessment)
