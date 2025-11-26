"""Quality Validator for Auto-Generated SPECs."""

import logging
import re
import time
from typing import Any, Dict, List

from moai_adk.core.spec.confidence_scoring import ConfidenceScoringSystem

# Configure logging
logger = logging.getLogger(__name__)

# Traceability tag patterns
TRACEABILITY_TAGS = ["@SPEC:", "@REQ:", "@TEST:", "@IMPL:", "@DOC:"]


class QualityValidator:
    """
    Quality Validator for auto-generated SPEC documents.

    This validator ensures that auto-generated SPECs meet quality standards,
    follow EARS format, and are complete and accurate.
    """

    def __init__(self, config: Dict[str, Any] = None):
        """Initialize the quality validator."""
        self.config = config or {}
        self.confidence_scorer = ConfidenceScoringSystem()

        # Quality thresholds from config
        self.min_ears_compliance = self.config.get("min_ears_compliance", 0.85)
        self.min_confidence_score = self.config.get("min_confidence_score", 0.7)
        self.min_content_length = self.config.get("min_content_length", 500)
        self.max_review_suggestions = self.config.get("max_review_suggestions", 10)

        # Quality metrics weights
        self.quality_weights = self.config.get(
            "quality_weights",
            {
                "ears_compliance": 0.3,
                "content_completeness": 0.25,
                "technical_accuracy": 0.2,
                "clarity_score": 0.15,
                "traceability": 0.1,
            },
        )

        # Validation rules
        self.validation_rules = {
            "required_sections": [
                "Overview",
                "Environment",
                "Assumptions",
                "Requirements",
                "Specifications",
                "Traceability",
            ],
            "required_plan_sections": [
                "Implementation Phases",
                "Technical Approach",
                "Success Criteria",
                "Next Steps",
            ],
            "required_acceptance_sections": [
                "Acceptance Criteria",
                "Validation Process",
                "Completion Criteria",
            ],
            "technical_keywords": [
                "API",
                "Database",
                "Authentication",
                "Security",
                "Performance",
                "Scalability",
                "Testing",
            ],
        }

    def validate_spec_quality(
        self, spec_content: Dict[str, str], code_analysis: Dict[str, Any] = None
    ) -> Dict[str, Any]:
        """
        Validate the quality of a generated SPEC.

        Args:
            spec_content: Dictionary with spec_md, plan_md, and acceptance_md
            code_analysis: Original code analysis for reference

        Returns:
            Quality validation results with scores and recommendations
        """
        start_time = time.time()

        logger.info("Starting SPEC quality validation")

        # Initialize validation results
        validation_result: Dict[str, Any] = {
            "validation_time": 0.0,
            "overall_score": 0.0,
            "quality_grade": "F",
            "passed_checks": [],
            "failed_checks": [],
            "recommendations": [],
            "metrics": {},
            "details": {},
        }

        try:
            # Validate EARS format compliance
            ears_result = self._validate_ears_format(spec_content["spec_md"])
            validation_result["details"]["ears_compliance"] = ears_result

            # Validate content completeness
            completeness_result = self._validate_content_completeness(spec_content)
            validation_result["details"]["content_completeness"] = completeness_result

            # Validate technical accuracy
            technical_result = self._validate_technical_accuracy(spec_content, code_analysis)
            validation_result["details"]["technical_accuracy"] = technical_result

            # Validate clarity and readability
            clarity_result = self._validate_clarity(spec_content)
            validation_result["details"]["clarity_score"] = clarity_result

            # Validate traceability
            traceability_result = self._validate_traceability(spec_content)
            validation_result["details"]["traceability"] = traceability_result

            # Calculate overall quality score
            overall_score = self._calculate_overall_score(validation_result["details"])
            validation_result["overall_score"] = overall_score

            # Determine quality grade
            validation_result["quality_grade"] = self._determine_quality_grade(overall_score)

            # Check if SPEC meets minimum quality standards
            meets_standards = self._check_quality_standards(validation_result)
            validation_result["meets_minimum_standards"] = meets_standards

            # Generate recommendations
            recommendations = self._generate_recommendations(validation_result)
            validation_result["recommendations"] = recommendations

            # Compile check results
            validation_result["passed_checks"] = self._compile_passed_checks(validation_result["details"])
            validation_result["failed_checks"] = self._compile_failed_checks(validation_result["details"])

        except Exception as e:
            logger.error(f"Error during SPEC validation: {str(e)}")
            validation_result["error"] = str(e)

        # Set validation time
        validation_result["validation_time"] = time.time() - start_time

        logger.info(f"SPEC quality validation completed in {validation_result['validation_time']:.2f}s")

        return validation_result

    def _validate_ears_format(self, spec_md: str) -> Dict[str, Any]:
        """Validate EARS format compliance."""
        logger.info("Validating EARS format compliance")

        required_sections = self.validation_rules["required_sections"]
        section_scores = {}
        missing_sections = []

        # Check each required section
        for section in required_sections:
            if section in spec_md:
                section_scores[section] = 1.0
            else:
                section_scores[section] = 0.0
                missing_sections.append(section)

        # Calculate overall compliance
        overall_compliance = sum(section_scores.values()) / len(required_sections)

        # Check for META information
        has_meta = "---" in spec_md and "title:" in spec_md

        # Check for tags
        has_tags = self._check_traceability_tags(spec_md)

        # Check for proper structure
        has_proper_headings = self._check_heading_structure(spec_md)

        return {
            "overall_compliance": round(overall_compliance, 2),
            "section_scores": section_scores,
            "missing_sections": missing_sections,
            "has_meta_info": has_meta,
            "has_tags": has_tags,
            "has_proper_structure": has_proper_headings,
            "total_sections": len(required_sections),
            "present_sections": sum(1 for score in section_scores.values() if score > 0),
        }

    def _validate_content_completeness(self, spec_content: Dict[str, str]) -> Dict[str, Any]:
        """Validate content completeness across all SPEC files."""
        logger.info("Validating content completeness")

        results = {
            "spec_completeness": 0.0,
            "plan_completeness": 0.0,
            "acceptance_completeness": 0.0,
            "overall_completeness": 0.0,
        }

        # Validate spec.md completeness
        spec_md = spec_content.get("spec_md", "")
        spec_completeness = self._assess_section_completeness(spec_md)
        results["spec_completeness"] = spec_completeness

        # Validate plan.md completeness
        plan_md = spec_content.get("plan_md", "")
        plan_completeness = self._assess_plan_completeness(plan_md)
        results["plan_completeness"] = plan_completeness

        # Validate acceptance.md completeness
        acceptance_md = spec_content.get("acceptance_md", "")
        acceptance_completeness = self._assess_acceptance_completeness(acceptance_md)
        results["acceptance_completeness"] = acceptance_completeness

        # Calculate overall completeness
        overall_completeness = (spec_completeness + plan_completeness + acceptance_completeness) / 3
        results["overall_completeness"] = round(overall_completeness, 2)

        return results

    def _validate_technical_accuracy(
        self, spec_content: Dict[str, str], code_analysis: Dict[str, Any] = None
    ) -> Dict[str, Any]:
        """Validate technical accuracy of the SPEC."""
        logger.info("Validating technical accuracy")

        spec_md = spec_content.get("spec_md", "")

        # Check for technical keywords presence
        technical_keywords = self.validation_rules["technical_keywords"]
        keyword_presence = self._check_technical_keywords(spec_md, technical_keywords)

        # Check consistency with code analysis
        consistency_score = 1.0
        if code_analysis:
            consistency_score = self._check_code_consistency(spec_md, code_analysis)

        # Check for technical specifications
        has_technical_specs = self._check_technical_specifications(spec_md)

        # Check for realistic estimates
        has_realistic_estimates = self._check_realistic_estimates(spec_md)

        return {
            "keyword_presence": keyword_presence,
            "code_consistency": consistency_score,
            "has_technical_specs": has_technical_specs,
            "has_realistic_estimates": has_realistic_estimates,
            "overall_accuracy": round(
                (
                    keyword_presence
                    + consistency_score
                    + (1.0 if has_technical_specs else 0.0)
                    + (1.0 if has_realistic_estimates else 0.0)
                )
                / 4,
                2,
            ),
        }

    def _validate_clarity(self, spec_content: Dict[str, str]) -> Dict[str, Any]:
        """Validate clarity and readability of the SPEC."""
        logger.info("Validating clarity and readability")

        spec_md = spec_content.get("spec_md", "")

        # Check for proper language use
        language_quality = self._check_language_quality(spec_md)

        # Check for clarity requirements
        clarity_requirements = self._check_clarity_requirements(spec_md)

        # Check for ambiguity
        ambiguity_score = self._check_ambiguity(spec_md)

        # Check for consistency
        consistency_score = self._check_consistency(spec_md)

        return {
            "language_quality": language_quality,
            "clarity_requirements": clarity_requirements,
            "ambiguity_score": ambiguity_score,
            "consistency_score": consistency_score,
            "overall_clarity": round(
                (language_quality + clarity_requirements + (1.0 - ambiguity_score) + consistency_score) / 4,
                2,
            ),
        }

    def _validate_traceability(self, spec_content: Dict[str, str]) -> Dict[str, Any]:
        """Validate traceability in the SPEC."""
        logger.info("Validating traceability")

        spec_md = spec_content.get("spec_md", "")

        # Check for basic traceability elements
        has_traceability = "@SPEC:" in spec_md or "Requirements:" in spec_md

        return {
            "has_traceability_elements": has_traceability,
            "overall_traceability": 1.0 if has_traceability else 0.5,
        }

    def _calculate_overall_score(self, details: Dict[str, Any]) -> float:
        """Calculate overall quality score."""
        weights = self.quality_weights

        overall_score = (
            details["ears_compliance"]["overall_compliance"] * weights["ears_compliance"]
            + details["content_completeness"]["overall_completeness"] * weights["content_completeness"]
            + details["technical_accuracy"]["overall_accuracy"] * weights["technical_accuracy"]
            + details["clarity_score"]["overall_clarity"] * weights["clarity_score"]
            + details["traceability"]["overall_traceability"] * weights["traceability"]
        )

        return round(overall_score, 2)

    def _determine_quality_grade(self, score: float) -> str:
        """Determine quality grade based on score."""
        if score >= 0.9:
            return "A"
        elif score >= 0.8:
            return "B"
        elif score >= 0.7:
            return "C"
        elif score >= 0.6:
            return "D"
        else:
            return "F"

    def _check_quality_standards(self, validation_result: Dict[str, Any]) -> bool:
        """Check if SPEC meets minimum quality standards."""
        overall_score = validation_result["overall_score"]
        ears_compliance = validation_result["details"]["ears_compliance"]["overall_compliance"]

        meets_overall = overall_score >= self.min_confidence_score
        meets_ears = ears_compliance >= self.min_ears_compliance

        return meets_overall and meets_ears

    def _generate_recommendations(self, validation_result: Dict[str, Any]) -> List[str]:
        """Generate improvement recommendations."""
        recommendations = []
        details = validation_result["details"]

        # EARS compliance recommendations
        if details["ears_compliance"]["overall_compliance"] < 1.0:
            missing = details["ears_compliance"]["missing_sections"]
            recommendations.append(f"Add missing EARS sections: {', '.join(missing[:3])}")

        # Content completeness recommendations
        if details["content_completeness"]["overall_completeness"] < 0.8:
            recommendations.append("Expand content sections with more detailed requirements")

        # Technical accuracy recommendations
        if details["technical_accuracy"]["overall_accuracy"] < 0.8:
            recommendations.append("Add more technical specifications and details")

        # Clarity recommendations
        if details["clarity_score"]["overall_clarity"] < 0.7:
            recommendations.append("Improve language clarity and reduce ambiguity")

        # Traceability recommendations
        if details["traceability"]["overall_traceability"] < 0.8:
            recommendations.append("Add proper traceability tags and relationships")

        # Limit recommendations
        return recommendations[: self.max_review_suggestions]

    def _compile_passed_checks(self, details: Dict[str, Any]) -> List[str]:
        """Compile list of passed quality checks."""
        passed = []

        if details["ears_compliance"]["overall_compliance"] >= self.min_ears_compliance:
            passed.append("EARS format compliance")

        if details["content_completeness"]["overall_completeness"] >= 0.8:
            passed.append("Content completeness")

        if details["technical_accuracy"]["overall_accuracy"] >= 0.7:
            passed.append("Technical accuracy")

        if details["clarity_score"]["overall_clarity"] >= 0.7:
            passed.append("Clarity and readability")

        if details["traceability"]["overall_traceability"] >= 0.7:
            passed.append("Traceability")

        return passed

    def _compile_failed_checks(self, details: Dict[str, Any]) -> List[str]:
        """Compile list of failed quality checks."""
        failed = []

        if details["ears_compliance"]["overall_compliance"] < self.min_ears_compliance:
            failed.append("EARS format compliance")

        if details["content_completeness"]["overall_completeness"] < 0.8:
            failed.append("Content completeness")

        if details["technical_accuracy"]["overall_accuracy"] < 0.7:
            failed.append("Technical accuracy")

        if details["clarity_score"]["overall_clarity"] < 0.7:
            failed.append("Clarity and readability")

        if details["traceability"]["overall_traceability"] < 0.7:
            failed.append("Traceability")

        return failed

    # Helper methods
    def _check_heading_structure(self, spec_md: str) -> bool:
        """Check if proper heading structure exists."""
        # Look for proper markdown heading structure
        heading_pattern = r"^#+\s+.*$"
        headings = re.findall(heading_pattern, spec_md, re.MULTILINE)
        return len(headings) >= 5  # At least 5 headings

    def _assess_section_completeness(self, content: str) -> float:
        """Assess completeness of a section."""
        if len(content) < self.min_content_length:
            return 0.0

        # Check for key indicators of completeness
        completeness_indicators = [
            r"##\s+.*",  # Subheadings
            r"\*\*.*\*\*",  # Bold text
            r"-\s+.*",  # Lists
            r"`[^`]+`",  # Code snippets
            r"\d+\.",  # Numbered lists
        ]

        score = 0.0
        for indicator in completeness_indicators:
            matches = re.findall(indicator, content)
            if matches:
                score += 0.2

        return min(score, 1.0)

    def _assess_plan_completeness(self, plan_content: str) -> float:
        """Assess completeness of implementation plan."""
        if not plan_content:
            return 0.0

        # Check for plan-specific elements
        plan_indicators = [
            r"Phase",  # Phases
            r"Priority",  # Priorities
            r"Task",  # Tasks
            r"\[\s*\]",  # Checkboxes
            r"Phase 1",  # Phase indicators
        ]

        score = 0.0
        for indicator in plan_indicators:
            if re.search(indicator, plan_content):
                score += 0.2

        return min(score, 1.0)

    def _assess_acceptance_criteria_completeness(self, acceptance_content: str) -> float:
        """Assess completeness of acceptance criteria."""
        if not acceptance_content:
            return 0.0

        # Check for acceptance-specific elements
        acceptance_indicators = [
            r"Acceptance",  # Acceptance
            r"Criteria",  # Criteria
            r"Pass",  # Pass
            r"Fail",  # Fail
            r"Test",  # Test
        ]

        score = 0.0
        for indicator in acceptance_indicators:
            if re.search(indicator, acceptance_content):
                score += 0.2

        return min(score, 1.0)

    def _check_technical_keywords(self, spec_md: str, keywords: List[str]) -> float:
        """Check for presence of technical keywords."""
        found_keywords = []
        for keyword in keywords:
            if keyword in spec_md:
                found_keywords.append(keyword)

        return len(found_keywords) / len(keywords)

    def _check_code_consistency(self, spec_md: str, code_analysis: Dict[str, Any]) -> float:
        """Check consistency between SPEC and code analysis."""
        # Extract key elements from code analysis
        classes = code_analysis.get("structure_info", {}).get("classes", [])
        functions = code_analysis.get("structure_info", {}).get("functions", [])

        # Check if these elements are mentioned in the SPEC
        class_mentions = sum(1 for cls in classes if cls in spec_md)
        function_mentions = sum(1 for func in functions if func in spec_md)

        total_elements = len(classes) + len(functions)
        if total_elements == 0:
            return 1.0

        consistency = (class_mentions + function_mentions) / total_elements
        return min(consistency, 1.0)

    def _check_technical_specifications(self, spec_md: str) -> bool:
        """Check for presence of technical specifications."""
        technical_patterns = [
            r"API\s+Endpoint",
            r"Database\s+Schema",
            r"Authentication",
            r"Performance\s+Requirement",
            r"Security\s+Requirement",
            r"Scalability\s+Plan",
            r"Testing\s+Strategy",
        ]

        return any(re.search(pattern, spec_md) for pattern in technical_patterns)

    def _check_realistic_estimates(self, spec_md: str) -> bool:
        """Check for realistic time/effort estimates."""
        # Look for realistic time estimates
        time_patterns = [r"1-2\s*days", r"3-5\s*days", r"1-2\s*weeks", r"\d+\s*hours"]

        return any(re.search(pattern, spec_md) for pattern in time_patterns)

    def _check_language_quality(self, spec_md: str) -> float:
        """Check language quality."""
        # Simple language quality checks
        sentences = spec_md.split(".")
        avg_sentence_length = sum(len(s.split()) for s in sentences) / len(sentences)

        # Optimal sentence length is around 15-20 words
        if 15 <= avg_sentence_length <= 25:
            return 1.0
        elif 10 <= avg_sentence_length <= 30:
            return 0.8
        else:
            return 0.5

    def _check_clarity_requirements(self, spec_md: str) -> float:
        """Check clarity requirements."""
        clarity_indicators = [
            r"Clear",
            r"Specific",
            r"Measurable",
            r"Achievable",
            r"Relevant",
        ]

        found_indicators = sum(1 for indicator in clarity_indicators if re.search(indicator, spec_md))

        return min(found_indicators / len(clarity_indicators), 1.0)

    def _check_ambiguity(self, spec_md: str) -> float:
        """Check for ambiguous language."""
        ambiguity_indicators = [
            r"degree",
            r"extent",
            r"approximately",
            r"about",
            r"around",
        ]

        ambiguous_count = sum(1 for indicator in ambiguity_indicators if re.search(indicator, spec_md))

        # Normalize by content length
        content_length = len(spec_md.split())
        ambiguity_ratio = ambiguous_count / max(content_length / 100, 1)

        return min(ambiguity_ratio, 1.0)

    def _check_consistency(self, spec_md: str) -> float:
        """Check for consistency in terminology."""
        # Check for consistent terminology
        # This is a simplified check - in practice, you'd use more sophisticated NLP
        sentences = spec_md.split(".")
        if len(sentences) < 2:
            return 1.0

        # Simple consistency check: look for repeated terms
        words = spec_md.lower().split()
        unique_words = set(words)
        consistency_ratio = len(unique_words) / len(words)

        return consistency_ratio

    def _check_traceability_tags(self, spec_md: str) -> bool:
        """Check for traceability tags."""
        return any(tag in spec_md for tag in TRACEABILITY_TAGS)

    def _check_tag_formatting(self, spec_md: str) -> float:
        """Check proper tag formatting."""
        # Look for properly formatted tags
        tag_pattern = r"@[A-Z]+:[A-Za-z0-9\-]+"
        matches = re.findall(tag_pattern, spec_md)

        # Count total tags to check ratio
        total_tags = len(re.findall(r"@[A-Z]+:", spec_md))

        if total_tags == 0:
            return 1.0

        return len(matches) / total_tags

    def _check_traceability_relationships(self, spec_md: str) -> float:
        """Check traceability relationships."""
        # Look for traceability indicators
        traceability_indicators = [
            r"←|→",  # Arrows for relationships
            r"Relationship",
            r"Connect",
            r"Trace",
        ]

        found_indicators = sum(1 for indicator in traceability_indicators if re.search(indicator, spec_md))

        return min(found_indicators / len(traceability_indicators), 1.0)

    def _assess_acceptance_completeness(self, acceptance_content: str) -> float:
        """Assess completeness of acceptance criteria section."""
        if not acceptance_content:
            return 0.0

        # Check for acceptance-specific elements
        acceptance_indicators = [
            r"Acceptance",  # Acceptance
            r"Criteria",  # Criteria
            r"Pass",  # Pass
            r"Fail",  # Fail
            r"Test",  # Test
        ]

        score = 0.0
        for indicator in acceptance_indicators:
            if re.search(indicator, acceptance_content):
                score += 0.2

        return min(score, 1.0)

    def generate_quality_report(self, validation_result: Dict[str, Any]) -> str:
        """Generate a quality report string."""
        report = f"""
# Quality Validation Report

## Overall Quality Score: {validation_result["overall_score"]:.2f}/1.0
## Quality Grade: {validation_result["quality_grade"]}
## Validation Time: {validation_result["validation_time"]:.2f}s

## Summary
- Passed Checks: {len(validation_result["passed_checks"])}
- Failed Checks: {len(validation_result["failed_checks"])}
- Recommendations: {len(validation_result["recommendations"])}

## Detailed Metrics
"""

        # Add detailed metrics
        for metric_name, metric_value in validation_result["metrics"].items():
            report += f"- {metric_name}: {metric_value:.2f}\n"

        # Add recommendations
        if validation_result["recommendations"]:
            report += "\n## Recommendations\n"
            for i, rec in enumerate(validation_result["recommendations"], 1):
                report += f"{i}. {rec}\n"

        # Add quality determination
        meets_standards = validation_result.get("meets_minimum_standards", False)
        status = "PASSED" if meets_standards else "FAILED"
        report += f"\n## Quality Status: {status}\n"

        return report
