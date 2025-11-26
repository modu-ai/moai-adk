"""Confidence Scoring System for Auto-SPEC Generation."""

import ast
import logging
import re
import time
from typing import Any, Dict, List, Tuple

# Configure logging
logger = logging.getLogger(__name__)


# SpecGenerator: Placeholder for spec generation functionality
class SpecGenerator:
    """Placeholder SpecGenerator class for confidence scoring."""

    def __init__(self):
        self.name = "SpecGenerator"

    def generate_spec(self, file_path: str, content: str) -> str:
        """Generate a basic SPEC document."""
        return f"SPEC document for {file_path}\n\nConfidence analysis: {content[:100]}..."


class ConfidenceScoringSystem:
    """
    Advanced confidence scoring system for auto-generated SPECs.

    This system analyzes code structure, domain relevance, and documentation
    quality to provide confidence scores for SPEC auto-generation.
    """

    def __init__(self):
        self.spec_generator = SpecGenerator()
        self.word_patterns = {
            "security": [
                "auth",
                "login",
                "password",
                "encrypt",
                "security",
                "bcrypt",
                "hash",
                "token",
            ],
            "data": [
                "model",
                "entity",
                "schema",
                "database",
                "persistence",
                "storage",
                "cache",
            ],
            "api": [
                "api",
                "endpoint",
                "route",
                "controller",
                "service",
                "handler",
                "middleware",
            ],
            "ui": [
                "ui",
                "interface",
                "component",
                "widget",
                "layout",
                "theme",
                "display",
            ],
            "business": ["business", "logic", "process", "workflow", "rule", "policy"],
            "testing": [
                "test",
                "mock",
                "fixture",
                "assertion",
                "verification",
                "validation",
            ],
        }

    def analyze_code_structure(self, file_path: str) -> Dict[str, float]:
        """
        Analyze code structure quality.

        Args:
            file_path: Path to the Python code file

        Returns:
            Dictionary with structure scores
        """
        try:
            with open(file_path, "r", encoding="utf-8") as f:
                code_content = f.read()

            # Parse AST
            tree = ast.parse(code_content)

            structure_scores = {
                "class_count": 0,
                "function_count": 0,
                "method_count": 0,
                "import_count": 0,
                "complexity_score": 0.0,
                "nesting_depth": 0,
                "docstring_coverage": 0.0,
                "naming_consistency": 0.0,
            }

            # Analyze AST nodes
            for node in ast.walk(tree):
                if isinstance(node, ast.ClassDef):
                    structure_scores["class_count"] += 1
                    # Check if class has docstring
                    if ast.get_docstring(node):
                        structure_scores["docstring_coverage"] += 0.1
                    # Count methods
                    for child in ast.walk(node):
                        if isinstance(child, ast.FunctionDef):
                            structure_scores["method_count"] += 1

                elif isinstance(node, ast.FunctionDef):
                    # Check if function is not a method (not inside a class)
                    is_method = False
                    for parent in ast.walk(tree):
                        if isinstance(parent, ast.ClassDef):
                            for child in parent.body:
                                if child == node:
                                    is_method = True
                                    break

                    if not is_method:
                        structure_scores["function_count"] += 1

                    # Check if function has docstring
                    if ast.get_docstring(node):
                        structure_scores["docstring_coverage"] += 0.1

                elif isinstance(node, ast.Import) or isinstance(node, ast.ImportFrom):
                    structure_scores["import_count"] += 1

            # Calculate complexity score (cyclomatic complexity approximation)
            structure_scores["complexity_score"] = self._calculate_complexity(tree)

            # Calculate nesting depth
            structure_scores["nesting_depth"] = self._calculate_nesting_depth(tree)

            # Calculate naming consistency
            structure_scores["naming_consistency"] = self._calculate_naming_consistency(tree)

            # Normalize scores (0-1 range)
            max_classes = 5  # Reasonable upper bound
            max_functions = 20
            max_methods = 50
            max_imports = 15
            max_complexity = 10
            max_nesting = 5
            max_docstring = 1.0

            normalized_scores = {
                "class_ratio": min(structure_scores["class_count"] / max_classes, 1.0),
                "function_ratio": min(structure_scores["function_count"] / max_functions, 1.0),
                "method_ratio": min(structure_scores["method_count"] / max_methods, 1.0),
                "import_ratio": min(structure_scores["import_count"] / max_imports, 1.0),
                "complexity_ratio": max(0, 1.0 - structure_scores["complexity_score"] / max_complexity),
                "nesting_ratio": max(0, 1.0 - structure_scores["nesting_depth"] / max_nesting),
                "docstring_score": min(structure_scores["docstring_coverage"] / max_docstring, 1.0),
                "naming_score": structure_scores["naming_consistency"],
            }

            return normalized_scores

        except Exception as e:
            logger.error(f"Error analyzing code structure: {e}")
            return self._get_default_structure_scores()

    def _calculate_complexity(self, tree: ast.AST) -> float:
        """Calculate approximate cyclomatic complexity."""
        complexity = 1  # Base complexity

        for node in ast.walk(tree):
            if isinstance(node, (ast.If, ast.For, ast.While, ast.Try, ast.With)):
                complexity += 1
            elif isinstance(node, ast.BoolOp):
                complexity += len(node.values) - 1

        return complexity

    def _calculate_nesting_depth(self, tree: ast.AST) -> int:
        """Calculate maximum nesting depth."""
        max_depth = 0

        def _visit(node, depth):
            nonlocal max_depth
            max_depth = max(max_depth, depth)

            for child in ast.iter_child_nodes(node):
                if isinstance(child, (ast.If, ast.For, ast.While, ast.Try, ast.With)):
                    _visit(child, depth + 1)
                else:
                    _visit(child, depth)

        _visit(tree, 0)
        return max_depth

    def _calculate_naming_consistency(self, tree: ast.AST) -> float:
        """Calculate naming consistency score."""
        names = []

        for node in ast.walk(tree):
            if isinstance(node, ast.Name):
                names.append(node.id)
            elif isinstance(node, ast.FunctionDef):
                names.append(node.name)
            elif isinstance(node, ast.ClassDef):
                names.append(node.name)

        if not names:
            return 1.0

        # Check naming patterns
        snake_case_count = 0
        camel_case_count = 0

        for name in names:
            if re.match(r"^[a-z]+(?:_[a-z]+)*$", name):
                snake_case_count += 1
            elif re.match(r"^[A-Z][a-zA-Z0-9]*$", name):
                camel_case_count += 1

        # Calculate consistency
        total_names = len(names)
        if total_names == 0:
            return 1.0

        snake_case_ratio = snake_case_count / total_names
        camel_case_ratio = camel_case_count / total_names

        # Favor consistency over specific style
        if snake_case_ratio > 0.7 or camel_case_ratio > 0.7:
            return 0.9
        elif snake_case_ratio > 0.5 or camel_case_ratio > 0.5:
            return 0.7
        else:
            return 0.5

    def _get_default_structure_scores(self) -> Dict[str, float]:
        """Get default structure scores for error cases."""
        return {
            "class_ratio": 0.5,
            "function_ratio": 0.5,
            "method_ratio": 0.5,
            "import_ratio": 0.5,
            "complexity_ratio": 0.7,
            "nesting_ratio": 0.7,
            "docstring_score": 0.3,
            "naming_score": 0.5,
        }

    def analyze_domain_relevance(self, file_path: str) -> Dict[str, float]:
        """
        Analyze domain relevance and keyword patterns.

        Args:
            file_path: Path to the code file

        Returns:
            Dictionary with domain relevance scores
        """
        try:
            with open(file_path, "r", encoding="utf-8") as f:
                content = f.read()

            # Normalize content
            content = content.lower()

            # Check for domain-specific patterns
            domain_scores = {}

            for domain, patterns in self.word_patterns.items():
                matches = 0
                for pattern in patterns:
                    if pattern in content:
                        matches += 1
                # Normalize by number of patterns
                domain_scores[f"{domain}_coverage"] = matches / len(patterns)

            # Calculate overall domain relevance
            total_relevance = sum(domain_scores.values())
            domain_scores["overall_relevance"] = min(total_relevance / len(self.word_patterns), 1.0)

            # Calculate domain specificity (how focused the code is)
            max_domain = max(domain_scores.values()) if domain_scores.values() else 0
            domain_scores["specificity"] = max_domain

            # Calculate technical vocabulary density
            technical_words = len(
                re.findall(
                    r"\b(?:api|endpoint|service|controller|model|entity|schema|database|cache|auth|login|password|token|session|user|admin|customer|product|order|payment|billing|subscription|plan|feature|function|method|class|interface|abstract|extends|implements|override|virtual|static|dynamic|async|await|promise|callback|event|handler|middleware|filter|validator|transformer|processor|worker|thread|queue|job|task|cron|scheduler|config|setting|env|variable|constant|property|attribute|field|column|table|index|constraint|foreign|primary|unique|notnull|default|check|trigger|procedure|function|stored|view|materialized|temp|temporary|permanent|persistent|volatile|in-memory|file-based|disk-based|cloud|distributed|clustered|load-balanced|scalable|high-availability|fault-tolerant|redundant|backup|restore|migration|version|branch|merge|conflict|resolve|commit|push|pull|fork|clone|repository|github|gitlab|bitbucket|ci|cd|pipeline|workflow|deployment|staging|production|development|testing|unit|integration|e2e|performance|load|stress|security|vulnerability|attack|breach|authentication|authorization|encryption|decryption|hash|salt|pepper|session|cookie|jwt|oauth|ldap|saml|rbac|abac|detection|prevention|monitoring|logging|tracing|metrics|analytics|dashboard|report|chart|graph|visualization|ui|ux|frontend|backend|fullstack|mobile|web|desktop|cross-platform|native|hybrid|responsive|adaptive|progressive|spa|pwa|ssr|csr|mvc|mvvm|riot|angular|react|vue|ember|backbone|knockout|jquery|vanilla|plain|pure|framework|library|package|module|bundle|dependency|require|import|export|include|extend|inherit|compose|aggregate|delegate|proxy|facade|adapter|bridge|decorator|singleton|factory|builder|prototype|command|observer|strategy|state|chain|iterator|visitor|mediator|composite|flyweight|proxy|interpreter|template|method|abstract|factory|builder|prototype|singleton|adapter|bridge|composite|decorator|facade|flyweight|proxy|chain|command|iterator|mediator|memento|observer|state|strategy|template|visitor)\b",
                    content,
                )
            )
            total_words = len(content.split())

            if total_words > 0:
                domain_scores["technical_density"] = min(technical_words / total_words, 1.0)
            else:
                domain_scores["technical_density"] = 0.0

            return domain_scores

        except Exception as e:
            logger.error(f"Error analyzing domain relevance: {e}")
            return {
                "security_coverage": 0.0,
                "data_coverage": 0.0,
                "api_coverage": 0.0,
                "ui_coverage": 0.0,
                "business_coverage": 0.0,
                "testing_coverage": 0.0,
                "overall_relevance": 0.5,
                "specificity": 0.5,
                "technical_density": 0.3,
            }

    def analyze_documentation_quality(self, file_path: str) -> Dict[str, float]:
        """
        Analyze documentation quality.

        Args:
            file_path: Path to the code file

        Returns:
            Dictionary with documentation scores
        """
        try:
            with open(file_path, "r", encoding="utf-8") as f:
                content = f.read()

            doc_scores = {
                "docstring_coverage": 0.0,
                "comment_density": 0.0,
                "explanation_quality": 0.0,
                "examples_present": 0.0,
                "parameter_documentation": 0.0,
                "return_documentation": 0.0,
                "exception_documentation": 0.0,
            }

            # Parse AST for docstring analysis
            tree = ast.parse(content)

            total_functions = 0
            documented_functions = 0
            documented_classes = 0
            total_classes = 0

            for node in ast.walk(tree):
                if isinstance(node, ast.FunctionDef):
                    total_functions += 1
                    if ast.get_docstring(node):
                        documented_functions += 1
                        # Check for parameter documentation
                        if ":" in ast.get_docstring(node) or "param" in ast.get_docstring(node):
                            doc_scores["parameter_documentation"] = 0.8
                        if "return" in ast.get_docstring(node) or "->" in ast.get_docstring(node):
                            doc_scores["return_documentation"] = 0.8
                        if "raise" in ast.get_docstring(node) or "exception" in ast.get_docstring(node):
                            doc_scores["exception_documentation"] = 0.8

                elif isinstance(node, ast.ClassDef):
                    total_classes += 1
                    if ast.get_docstring(node):
                        documented_classes += 1

            # Calculate docstring coverage
            if total_functions > 0:
                doc_scores["docstring_coverage"] = documented_functions / total_functions
            if total_classes > 0:
                class_coverage = documented_classes / total_classes
                doc_scores["docstring_coverage"] = max(doc_scores["docstring_coverage"], class_coverage)

            # Calculate comment density
            lines = content.split("\n")
            comment_lines = 0
            code_lines = 0

            for line in lines:
                stripped = line.strip()
                if stripped.startswith("#"):
                    comment_lines += 1
                elif stripped and not stripped.startswith('"""') and not stripped.startswith("'''"):
                    code_lines += 1

            if code_lines > 0:
                doc_scores["comment_density"] = min(comment_lines / code_lines, 1.0)

            # Check for examples in docstrings
            total_docstring_length = len(ast.get_docstring(tree) or "")
            if total_docstring_length > 0:
                example_count = len(re.findall(r">>>|Example:|example:|\b\d+\.\s", content))
                doc_scores["examples_present"] = min(example_count / 3, 1.0)

            # Calculate explanation quality based on docstring content
            docstring_content = ast.get_docstring(tree) or ""
            if docstring_content:
                # Check for good explanation indicators
                explanation_indicators = [
                    "provides",
                    "allows",
                    "enables",
                    "implements",
                    "handles",
                    "processes",
                    "manages",
                ]
                explanation_count = sum(1 for indicator in explanation_indicators if indicator in docstring_content)
                doc_scores["explanation_quality"] = min(explanation_count / len(explanation_indicators), 1.0)

            return doc_scores

        except Exception as e:
            logger.error(f"Error analyzing documentation quality: {e}")
            return {
                "docstring_coverage": 0.3,
                "comment_density": 0.2,
                "explanation_quality": 0.3,
                "examples_present": 0.0,
                "parameter_documentation": 0.2,
                "return_documentation": 0.2,
                "exception_documentation": 0.1,
            }

    def calculate_confidence_score(
        self,
        file_path: str,
        structure_weights: Dict[str, float] = None,
        domain_weights: Dict[str, float] = None,
        doc_weights: Dict[str, float] = None,
    ) -> Tuple[float, Dict[str, Any]]:
        """
        Calculate overall confidence score for auto-SPEC generation.

        Args:
            file_path: Path to the code file
            structure_weights: Weights for structure analysis
            domain_weights: Weights for domain analysis
            doc_weights: Weights for documentation analysis

        Returns:
            Tuple of (confidence_score, detailed_analysis)
        """
        start_time = time.time()

        # Default weights
        default_structure_weights = {
            "class_ratio": 0.1,
            "function_ratio": 0.1,
            "method_ratio": 0.1,
            "import_ratio": 0.1,
            "complexity_ratio": 0.15,
            "nesting_ratio": 0.15,
            "docstring_score": 0.15,
            "naming_score": 0.15,
        }

        default_domain_weights = {
            "overall_relevance": 0.3,
            "specificity": 0.2,
            "technical_density": 0.3,
            "security_coverage": 0.1,
            "data_coverage": 0.1,
        }

        default_doc_weights = {
            "docstring_coverage": 0.3,
            "comment_density": 0.2,
            "explanation_quality": 0.2,
            "examples_present": 0.1,
            "parameter_documentation": 0.1,
            "return_documentation": 0.1,
        }

        # Use provided weights or defaults
        structure_weights = structure_weights or default_structure_weights
        domain_weights = domain_weights or default_domain_weights
        doc_weights = doc_weights or default_doc_weights

        # Analyze code
        structure_analysis = self.analyze_code_structure(file_path)
        domain_analysis = self.analyze_domain_relevance(file_path)
        doc_analysis = self.analyze_documentation_quality(file_path)

        # Calculate weighted scores
        structure_score = sum(structure_analysis[key] * structure_weights.get(key, 0) for key in structure_analysis)

        domain_score = sum(domain_analysis[key] * domain_weights.get(key, 0) for key in domain_analysis)

        doc_score = sum(doc_analysis[key] * doc_weights.get(key, 0) for key in doc_analysis)

        # Final confidence score (weighted average)
        total_weights = sum(structure_weights.values()) + sum(domain_weights.values()) + sum(doc_weights.values())

        final_confidence = (structure_score + domain_score + doc_score) / total_weights

        # Round to 2 decimal places
        final_confidence = round(final_confidence, 2)

        # Create detailed analysis
        detailed_analysis = {
            "file_path": file_path,
            "analysis_time": time.time() - start_time,
            "confidence_score": final_confidence,
            "structure_analysis": {
                "score": round(structure_score, 2),
                "details": structure_analysis,
                "weights": structure_weights,
            },
            "domain_analysis": {
                "score": round(domain_score, 2),
                "details": domain_analysis,
                "weights": domain_weights,
            },
            "documentation_analysis": {
                "score": round(doc_score, 2),
                "details": doc_analysis,
                "weights": doc_weights,
            },
            "recommendations": self._generate_recommendations(structure_analysis, domain_analysis, doc_analysis),
        }

        return final_confidence, detailed_analysis

    def _generate_recommendations(
        self, structure_analysis: Dict, domain_analysis: Dict, doc_analysis: Dict
    ) -> List[str]:
        """Generate improvement recommendations."""
        recommendations = []

        # Structure recommendations
        if structure_analysis.get("docstring_score", 0) < 0.5:
            recommendations.append("Add more docstrings to improve documentation coverage")

        if structure_analysis.get("complexity_ratio", 0) < 0.7:
            recommendations.append("Consider refactoring complex functions to improve maintainability")

        if structure_analysis.get("naming_score", 0) < 0.7:
            recommendations.append("Improve naming consistency (use consistent naming convention)")

        # Domain recommendations
        if domain_analysis.get("overall_relevance", 0) < 0.6:
            recommendations.append("Add domain-specific terminology to improve relevance")

        if domain_analysis.get("technical_density", 0) < 0.3:
            recommendations.append("Increase technical vocabulary for better specification")

        # Documentation recommendations
        if doc_analysis.get("examples_present", 0) < 0.5:
            recommendations.append("Add usage examples in docstrings for better understanding")

        if doc_analysis.get("parameter_documentation", 0) < 0.5:
            recommendations.append("Document function parameters and return values")

        return recommendations[:5]  # Return top 5 recommendations

    def validate_confidence_threshold(
        self, confidence_score: float, threshold: float = 0.7, strict_mode: bool = False
    ) -> Dict[str, Any]:
        """
        Validate confidence score against threshold.

        Args:
            confidence_score: Calculated confidence score
            threshold: Minimum confidence threshold
            strict_mode: Whether to use strict validation

        Returns:
            Validation result
        """
        if strict_mode:
            # Strict mode: all scores must meet threshold
            meets_threshold = confidence_score >= threshold
        else:
            # Normal mode: average score meets threshold
            meets_threshold = confidence_score >= threshold

        validation_result = {
            "meets_threshold": meets_threshold,
            "confidence_score": confidence_score,
            "threshold": threshold,
            "difference": confidence_score - threshold,
            "recommendation": self._get_threshold_recommendation(confidence_score, threshold),
        }

        return validation_result

    def _get_threshold_recommendation(self, confidence_score: float, threshold: float) -> str:
        """Get recommendation based on confidence score."""
        if confidence_score >= threshold:
            if confidence_score >= 0.9:
                return "Excellent confidence level - auto-spec generation recommended"
            elif confidence_score >= 0.8:
                return "Good confidence level - auto-spec generation recommended"
            else:
                return "Acceptable confidence level - auto-spec generation recommended"
        else:
            if confidence_score >= 0.6:
                return "Marginal confidence level - manual review recommended"
            elif confidence_score >= 0.4:
                return "Low confidence level - significant improvements needed"
            else:
                return "Very low confidence level - complete redesign recommended"

    def get_confidence_breakdown(self, confidence_score: float) -> Dict[str, Any]:
        """Get detailed breakdown of confidence score components."""
        return {
            "overall_score": confidence_score,
            "interpretation": self._interpret_confidence_score(confidence_score),
            "risk_level": self._get_risk_level(confidence_score),
            "action_required": self._get_action_required(confidence_score),
        }

    def _interpret_confidence_score(self, score: float) -> str:
        """Interpret confidence score meaning."""
        if score >= 0.9:
            return "Excellent - Very high likelihood of generating a quality SPEC"
        elif score >= 0.8:
            return "Good - High likelihood of generating a quality SPEC"
        elif score >= 0.7:
            return "Acceptable - Moderate likelihood of generating a quality SPEC"
        elif score >= 0.6:
            return "Marginal - Low likelihood of generating a quality SPEC"
        elif score >= 0.4:
            return "Poor - Very low likelihood of generating a quality SPEC"
        else:
            return "Very Poor - Extremely low likelihood of generating a quality SPEC"

    def _get_risk_level(self, score: float) -> str:
        """Get risk level based on confidence score."""
        if score >= 0.8:
            return "Low"
        elif score >= 0.6:
            return "Medium"
        elif score >= 0.4:
            return "High"
        else:
            return "Critical"

    def _get_action_required(self, score: float) -> str:
        """Get required action based on confidence score."""
        if score >= 0.7:
            return "Auto-generate SPEC"
        elif score >= 0.5:
            return "Generate SPEC with manual review"
        else:
            return "Do not auto-generate - require manual creation"


# Utility function for backwards compatibility
def calculate_completion_confidence(analysis: Dict[str, Any]) -> float:
    """
    Backwards compatibility function.

    Args:
        analysis: Code analysis result

    Returns:
        Confidence score
    """
    scorer = ConfidenceScoringSystem()

    # Extract file path from analysis or use default
    file_path = analysis.get("file_path", "dummy_file.py")

    # Calculate confidence score
    confidence, detailed_analysis = scorer.calculate_confidence_score(file_path)

    return confidence
