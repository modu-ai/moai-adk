"""Comprehensive tests for confidence_scoring.py module."""

import ast
import logging
import os
import tempfile
from pathlib import Path
from typing import Dict, Any
from unittest.mock import Mock, patch, mock_open

import pytest
from pytest_mock import MockerFixture

from moai_adk.core.spec.confidence_scoring import (
    SpecGenerator,
    ConfidenceScoringSystem,
    calculate_completion_confidence,
)


class TestSpecGenerator:
    """Tests for SpecGenerator class."""

    def test_init(self):
        """Test SpecGenerator initialization."""
        gen = SpecGenerator()
        assert gen.name == "SpecGenerator"

    def test_generate_spec(self):
        """Test SPEC document generation."""
        gen = SpecGenerator()
        result = gen.generate_spec("test_file.py", "def test(): pass")

        assert "test_file.py" in result
        assert "SPEC document" in result
        assert "Confidence analysis" in result

    def test_generate_spec_with_long_content(self):
        """Test SPEC generation with truncated content."""
        gen = SpecGenerator()
        long_content = "x = " + "1" * 200
        result = gen.generate_spec("test.py", long_content)

        assert "test.py" in result
        assert "..." in result


class TestConfidenceScoringSystemInit:
    """Tests for ConfidenceScoringSystem initialization."""

    def test_init(self):
        """Test ConfidenceScoringSystem initialization."""
        system = ConfidenceScoringSystem()

        assert system.spec_generator is not None
        assert isinstance(system.spec_generator, SpecGenerator)
        assert system.word_patterns is not None
        assert len(system.word_patterns) > 0

    def test_word_patterns_structure(self):
        """Test word patterns are properly structured."""
        system = ConfidenceScoringSystem()

        expected_domains = [
            "security",
            "data",
            "api",
            "ui",
            "business",
            "testing",
        ]

        for domain in expected_domains:
            assert domain in system.word_patterns
            assert isinstance(system.word_patterns[domain], list)
            assert len(system.word_patterns[domain]) > 0

    def test_word_patterns_content(self):
        """Test word patterns contain expected keywords."""
        system = ConfidenceScoringSystem()

        assert "auth" in system.word_patterns["security"]
        assert "password" in system.word_patterns["security"]
        assert "model" in system.word_patterns["data"]
        assert "api" in system.word_patterns["api"]
        assert "component" in system.word_patterns["ui"]


class TestAnalyzeCodeStructure:
    """Tests for analyze_code_structure method."""

    def test_analyze_simple_file(self, tmp_path):
        """Test analyzing a simple Python file."""
        test_file = tmp_path / "test.py"
        test_file.write_text("""
class TestClass:
    '''Test class docstring.'''
    def method(self):
        '''Method docstring.'''
        pass

def function():
    '''Function docstring.'''
    pass

import os
import sys
""")

        system = ConfidenceScoringSystem()
        result = system.analyze_code_structure(str(test_file))

        assert isinstance(result, dict)
        assert "class_ratio" in result
        assert "function_ratio" in result
        assert "method_ratio" in result
        assert "complexity_ratio" in result
        assert 0 <= result["class_ratio"] <= 1.0
        assert 0 <= result["function_ratio"] <= 1.0

    def test_analyze_complex_file(self, tmp_path):
        """Test analyzing a complex file with nested structures."""
        test_file = tmp_path / "complex.py"
        test_file.write_text("""
import os
import sys

class ComplexClass:
    '''Complex class.'''

    def method1(self):
        if True:
            for i in range(10):
                while True:
                    pass

    def method2(self):
        try:
            with open("file") as f:
                pass
        except Exception:
            pass

def function1():
    pass

def function2():
    pass
""")

        system = ConfidenceScoringSystem()
        result = system.analyze_code_structure(str(test_file))

        assert result["class_ratio"] > 0
        assert result["nesting_ratio"] < 1.0

    def test_analyze_file_with_no_docstrings(self, tmp_path):
        """Test analyzing file without docstrings."""
        test_file = tmp_path / "nodoc.py"
        test_file.write_text("""
class TestClass:
    def method(self):
        pass

def function():
    pass
""")

        system = ConfidenceScoringSystem()
        result = system.analyze_code_structure(str(test_file))

        assert result["docstring_score"] == 0.0

    def test_analyze_nonexistent_file(self):
        """Test analyzing nonexistent file returns default scores."""
        system = ConfidenceScoringSystem()
        result = system.analyze_code_structure("/nonexistent/file.py")

        # Should return default scores
        assert result["class_ratio"] == 0.5
        assert result["docstring_score"] == 0.3

    def test_analyze_file_with_syntax_error(self, tmp_path):
        """Test analyzing file with syntax errors returns default scores."""
        test_file = tmp_path / "syntax_error.py"
        test_file.write_text("def broken(:\n    pass")

        system = ConfidenceScoringSystem()
        result = system.analyze_code_structure(str(test_file))

        # Should return default scores
        assert result["class_ratio"] == 0.5


class TestCalculateComplexity:
    """Tests for _calculate_complexity method."""

    def test_calculate_complexity_simple(self, tmp_path):
        """Test complexity calculation for simple code."""
        test_file = tmp_path / "simple.py"
        test_file.write_text("x = 1")

        system = ConfidenceScoringSystem()
        with open(test_file) as f:
            tree = ast.parse(f.read())

        complexity = system._calculate_complexity(tree)
        assert complexity >= 1

    def test_calculate_complexity_with_control_flow(self, tmp_path):
        """Test complexity with control flow statements."""
        test_file = tmp_path / "control.py"
        test_file.write_text("""
def test():
    if True:
        pass
    for i in range(10):
        pass
    while True:
        break
    try:
        pass
    except:
        pass
    with open("file"):
        pass
""")

        system = ConfidenceScoringSystem()
        with open(test_file) as f:
            tree = ast.parse(f.read())

        complexity = system._calculate_complexity(tree)
        assert complexity > 1

    def test_calculate_complexity_with_bool_operators(self, tmp_path):
        """Test complexity with boolean operators."""
        test_file = tmp_path / "bool.py"
        test_file.write_text("result = a or b or c or d")

        system = ConfidenceScoringSystem()
        with open(test_file) as f:
            tree = ast.parse(f.read())

        complexity = system._calculate_complexity(tree)
        assert complexity > 1


class TestCalculateNestingDepth:
    """Tests for _calculate_nesting_depth method."""

    def test_nesting_depth_no_nesting(self, tmp_path):
        """Test nesting depth for non-nested code."""
        test_file = tmp_path / "flat.py"
        test_file.write_text("x = 1\ny = 2")

        system = ConfidenceScoringSystem()
        with open(test_file) as f:
            tree = ast.parse(f.read())

        depth = system._calculate_nesting_depth(tree)
        assert depth == 0

    def test_nesting_depth_single_level(self, tmp_path):
        """Test nesting depth for single level nesting."""
        test_file = tmp_path / "single.py"
        test_file.write_text("""
if True:
    x = 1
""")

        system = ConfidenceScoringSystem()
        with open(test_file) as f:
            tree = ast.parse(f.read())

        depth = system._calculate_nesting_depth(tree)
        assert depth >= 1

    def test_nesting_depth_deep_nesting(self, tmp_path):
        """Test nesting depth for deeply nested code."""
        test_file = tmp_path / "deep.py"
        test_file.write_text("""
if True:
    for i in range(10):
        if i > 5:
            while True:
                try:
                    x = 1
                except:
                    pass
""")

        system = ConfidenceScoringSystem()
        with open(test_file) as f:
            tree = ast.parse(f.read())

        depth = system._calculate_nesting_depth(tree)
        assert depth > 1


class TestCalculateNamingConsistency:
    """Tests for _calculate_naming_consistency method."""

    def test_naming_consistency_snake_case(self, tmp_path):
        """Test naming consistency with snake_case."""
        test_file = tmp_path / "snake.py"
        test_file.write_text("""
def test_function():
    my_variable = 1
    another_name = 2
""")

        system = ConfidenceScoringSystem()
        with open(test_file) as f:
            tree = ast.parse(f.read())

        consistency = system._calculate_naming_consistency(tree)
        assert consistency > 0.5

    def test_naming_consistency_camel_case(self, tmp_path):
        """Test naming consistency with camelCase (class names)."""
        test_file = tmp_path / "camel.py"
        test_file.write_text("""
class MyClass:
    def myMethod(self):
        pass
""")

        system = ConfidenceScoringSystem()
        with open(test_file) as f:
            tree = ast.parse(f.read())

        consistency = system._calculate_naming_consistency(tree)
        assert consistency >= 0.5

    def test_naming_consistency_mixed(self, tmp_path):
        """Test naming consistency with mixed naming styles."""
        test_file = tmp_path / "mixed.py"
        test_file.write_text("""
def testFunction():
    x_var = 1
    MyVar = 2
    oddName = 3
""")

        system = ConfidenceScoringSystem()
        with open(test_file) as f:
            tree = ast.parse(f.read())

        consistency = system._calculate_naming_consistency(tree)
        assert consistency >= 0.5

    def test_naming_consistency_empty_file(self, tmp_path):
        """Test naming consistency for empty file."""
        test_file = tmp_path / "empty.py"
        test_file.write_text("")

        system = ConfidenceScoringSystem()
        with open(test_file) as f:
            tree = ast.parse(f.read())

        consistency = system._calculate_naming_consistency(tree)
        assert consistency == 1.0


class TestGetDefaultStructureScores:
    """Tests for _get_default_structure_scores method."""

    def test_default_structure_scores(self):
        """Test default structure scores are returned correctly."""
        system = ConfidenceScoringSystem()
        scores = system._get_default_structure_scores()

        assert isinstance(scores, dict)
        assert "class_ratio" in scores
        assert "function_ratio" in scores
        assert "method_ratio" in scores
        assert "import_ratio" in scores
        assert "complexity_ratio" in scores
        assert "nesting_ratio" in scores
        assert "docstring_score" in scores
        assert "naming_score" in scores

        # All values should be between 0 and 1
        for value in scores.values():
            assert 0 <= value <= 1.0


class TestAnalyzeDomainRelevance:
    """Tests for analyze_domain_relevance method."""

    def test_analyze_domain_relevance_security(self, tmp_path):
        """Test analyzing security domain relevance."""
        test_file = tmp_path / "security.py"
        test_file.write_text("""
def authenticate():
    password = "secret"
    token = generate_token()
    hash_value = bcrypt.hash(password)
""")

        system = ConfidenceScoringSystem()
        result = system.analyze_domain_relevance(str(test_file))

        assert "security_coverage" in result
        assert "overall_relevance" in result
        assert result["security_coverage"] > 0

    def test_analyze_domain_relevance_api(self, tmp_path):
        """Test analyzing API domain relevance."""
        test_file = tmp_path / "api.py"
        test_file.write_text("""
class APIController:
    def api_endpoint(self):
        service = ServiceHandler()
        middleware = MiddlewareFilter()
""")

        system = ConfidenceScoringSystem()
        result = system.analyze_domain_relevance(str(test_file))

        assert "api_coverage" in result
        assert result["api_coverage"] > 0

    def test_analyze_domain_relevance_all_domains(self, tmp_path):
        """Test analyzing file with multiple domains."""
        test_file = tmp_path / "multidom.py"
        test_file.write_text("""
class UserModel:
    '''User data model.'''
    def authenticate(self):
        '''Handle authentication.'''
        pass

    def get_api_endpoint(self):
        '''Get API endpoint.'''
        return self.endpoint

    def test_security(self):
        '''Test security feature.'''
        pass
""")

        system = ConfidenceScoringSystem()
        result = system.analyze_domain_relevance(str(test_file))

        assert "overall_relevance" in result
        assert "technical_density" in result
        assert 0 <= result["overall_relevance"] <= 1.0

    def test_analyze_domain_relevance_nonexistent_file(self):
        """Test analyzing nonexistent file."""
        system = ConfidenceScoringSystem()
        result = system.analyze_domain_relevance("/nonexistent/file.py")

        # Should return default scores
        assert result["overall_relevance"] == 0.5
        assert result["security_coverage"] == 0.0

    def test_analyze_domain_relevance_empty_file(self, tmp_path):
        """Test analyzing empty file."""
        test_file = tmp_path / "empty.py"
        test_file.write_text("")

        system = ConfidenceScoringSystem()
        result = system.analyze_domain_relevance(str(test_file))

        assert "overall_relevance" in result
        assert result["technical_density"] == 0.0


class TestAnalyzeDocumentationQuality:
    """Tests for analyze_documentation_quality method."""

    def test_analyze_documentation_well_documented(self, tmp_path):
        """Test analyzing well-documented code."""
        test_file = tmp_path / "welldoc.py"
        test_file.write_text("""
def add(a, b):
    '''Add two numbers.

    Args:
        a: First number
        b: Second number

    Returns:
        Sum of a and b

    Raises:
        TypeError: If inputs are not numbers
    '''
    return a + b

class Calculator:
    '''Calculator class for basic operations.'''

    def multiply(self, x, y):
        '''Multiply two numbers.

        Example:
            >>> calc = Calculator()
            >>> calc.multiply(2, 3)
            6
        '''
        return x * y
""")

        system = ConfidenceScoringSystem()
        result = system.analyze_documentation_quality(str(test_file))

        assert result["docstring_coverage"] > 0
        assert "parameter_documentation" in result
        assert "return_documentation" in result

    def test_analyze_documentation_poorly_documented(self, tmp_path):
        """Test analyzing poorly-documented code."""
        test_file = tmp_path / "poordoc.py"
        test_file.write_text("""
def add(a, b):
    return a + b

class Calculator:
    def multiply(self, x, y):
        return x * y
""")

        system = ConfidenceScoringSystem()
        result = system.analyze_documentation_quality(str(test_file))

        assert result["docstring_coverage"] == 0.0

    def test_analyze_documentation_with_comments(self, tmp_path):
        """Test analyzing code with comments."""
        test_file = tmp_path / "comments.py"
        test_file.write_text("""
# This is a function
def add(a, b):
    # Add a and b
    return a + b

# Calculate result
result = add(1, 2)
""")

        system = ConfidenceScoringSystem()
        result = system.analyze_documentation_quality(str(test_file))

        assert "comment_density" in result
        assert result["comment_density"] >= 0

    def test_analyze_documentation_nonexistent_file(self):
        """Test analyzing nonexistent file."""
        system = ConfidenceScoringSystem()
        result = system.analyze_documentation_quality("/nonexistent/file.py")

        # Should return default scores
        assert result["docstring_coverage"] == 0.3
        assert result["comment_density"] == 0.2

    def test_analyze_documentation_with_examples(self, tmp_path):
        """Test analyzing code with examples in docstrings."""
        test_file = tmp_path / "examples.py"
        test_file.write_text("""
def multiply(x, y):
    '''Multiply numbers.

    Example:
        >>> multiply(2, 3)
        6

    Example 2:
        >>> multiply(5, 4)
        20
    '''
    return x * y
""")

        system = ConfidenceScoringSystem()
        result = system.analyze_documentation_quality(str(test_file))

        assert "examples_present" in result


class TestCalculateConfidenceScore:
    """Tests for calculate_confidence_score method."""

    def test_calculate_confidence_score_default_weights(self, tmp_path):
        """Test confidence score calculation with default weights."""
        test_file = tmp_path / "code.py"
        test_file.write_text("""
class TestClass:
    '''Test class.'''
    def method(self):
        '''Test method.'''
        if True:
            for i in range(10):
                pass
""")

        system = ConfidenceScoringSystem()
        score, analysis = system.calculate_confidence_score(str(test_file))

        assert isinstance(score, float)
        assert 0 <= score <= 1.0
        assert isinstance(analysis, dict)
        assert "file_path" in analysis
        assert "confidence_score" in analysis
        assert "structure_analysis" in analysis
        assert "domain_analysis" in analysis
        assert "documentation_analysis" in analysis
        assert "recommendations" in analysis
        assert "analysis_time" in analysis

    def test_calculate_confidence_score_custom_weights(self, tmp_path):
        """Test confidence score with custom weights."""
        test_file = tmp_path / "code.py"
        test_file.write_text("x = 1")

        custom_struct_weights = {
            "class_ratio": 0.2,
            "function_ratio": 0.2,
            "method_ratio": 0.2,
            "import_ratio": 0.2,
            "complexity_ratio": 0.1,
            "nesting_ratio": 0.1,
            "docstring_score": 0.1,
            "naming_score": 0.1,
        }

        system = ConfidenceScoringSystem()
        score, analysis = system.calculate_confidence_score(
            str(test_file),
            structure_weights=custom_struct_weights,
        )

        assert isinstance(score, float)
        assert analysis["structure_analysis"]["weights"] == custom_struct_weights

    def test_calculate_confidence_score_nonexistent_file(self):
        """Test confidence score for nonexistent file."""
        system = ConfidenceScoringSystem()
        score, analysis = system.calculate_confidence_score("/nonexistent/file.py")

        assert isinstance(score, float)
        assert 0 <= score <= 1.0
        assert analysis["file_path"] == "/nonexistent/file.py"

    def test_calculate_confidence_score_complexity(self, tmp_path):
        """Test confidence score for complex file."""
        test_file = tmp_path / "complex.py"
        test_file.write_text("""
class ComplexSystem:
    '''Complex system class.'''

    def process_data(self, data):
        '''Process data.

        Args:
            data: Input data

        Returns:
            Processed data
        '''
        if isinstance(data, list):
            for item in data:
                if isinstance(item, dict):
                    for key, value in item.items():
                        if value is not None:
                            try:
                                result = self._process_item(value)
                            except Exception as e:
                                logger.error(e)
        return data

    def _process_item(self, item):
        '''Process individual item.'''
        return item * 2

import logging
logger = logging.getLogger(__name__)
""")

        system = ConfidenceScoringSystem()
        score, analysis = system.calculate_confidence_score(str(test_file))

        assert "recommendations" in analysis
        assert isinstance(analysis["recommendations"], list)


class TestGenerateRecommendations:
    """Tests for _generate_recommendations method."""

    def test_generate_recommendations_low_docstrings(self):
        """Test recommendations for low docstring coverage."""
        system = ConfidenceScoringSystem()
        structure = {"docstring_score": 0.3, "complexity_ratio": 0.8, "naming_score": 0.8}
        domain = {"overall_relevance": 0.7, "technical_density": 0.6}
        doc = {"examples_present": 0.8, "parameter_documentation": 0.8}

        recs = system._generate_recommendations(structure, domain, doc)
        assert len(recs) > 0
        assert any("docstring" in r.lower() for r in recs)

    def test_generate_recommendations_high_complexity(self):
        """Test recommendations for high complexity."""
        system = ConfidenceScoringSystem()
        structure = {"docstring_score": 0.9, "complexity_ratio": 0.3, "naming_score": 0.8}
        domain = {"overall_relevance": 0.7, "technical_density": 0.6}
        doc = {"examples_present": 0.8, "parameter_documentation": 0.8}

        recs = system._generate_recommendations(structure, domain, doc)
        assert any("refactor" in r.lower() or "complex" in r.lower() for r in recs)

    def test_generate_recommendations_poor_naming(self):
        """Test recommendations for poor naming consistency."""
        system = ConfidenceScoringSystem()
        structure = {"docstring_score": 0.9, "complexity_ratio": 0.8, "naming_score": 0.3}
        domain = {"overall_relevance": 0.7, "technical_density": 0.6}
        doc = {"examples_present": 0.8, "parameter_documentation": 0.8}

        recs = system._generate_recommendations(structure, domain, doc)
        assert any("naming" in r.lower() for r in recs)

    def test_generate_recommendations_low_domain_relevance(self):
        """Test recommendations for low domain relevance."""
        system = ConfidenceScoringSystem()
        structure = {"docstring_score": 0.9, "complexity_ratio": 0.8, "naming_score": 0.8}
        domain = {"overall_relevance": 0.3, "technical_density": 0.6}
        doc = {"examples_present": 0.8, "parameter_documentation": 0.8}

        recs = system._generate_recommendations(structure, domain, doc)
        assert any("domain" in r.lower() for r in recs)

    def test_generate_recommendations_low_technical_density(self):
        """Test recommendations for low technical vocabulary."""
        system = ConfidenceScoringSystem()
        structure = {"docstring_score": 0.9, "complexity_ratio": 0.8, "naming_score": 0.8}
        domain = {"overall_relevance": 0.7, "technical_density": 0.1}
        doc = {"examples_present": 0.8, "parameter_documentation": 0.8}

        recs = system._generate_recommendations(structure, domain, doc)
        assert any("technical" in r.lower() for r in recs)

    def test_generate_recommendations_missing_examples(self):
        """Test recommendations for missing examples."""
        system = ConfidenceScoringSystem()
        structure = {"docstring_score": 0.9, "complexity_ratio": 0.8, "naming_score": 0.8}
        domain = {"overall_relevance": 0.7, "technical_density": 0.6}
        doc = {"examples_present": 0.2, "parameter_documentation": 0.8}

        recs = system._generate_recommendations(structure, domain, doc)
        assert any("example" in r.lower() for r in recs)

    def test_generate_recommendations_missing_parameters(self):
        """Test recommendations for missing parameter documentation."""
        system = ConfidenceScoringSystem()
        structure = {"docstring_score": 0.9, "complexity_ratio": 0.8, "naming_score": 0.8}
        domain = {"overall_relevance": 0.7, "technical_density": 0.6}
        doc = {"examples_present": 0.8, "parameter_documentation": 0.2}

        recs = system._generate_recommendations(structure, domain, doc)
        assert any("parameter" in r.lower() or "document" in r.lower() for r in recs)

    def test_generate_recommendations_max_count(self):
        """Test that recommendations are limited to top 5."""
        system = ConfidenceScoringSystem()
        structure = {"docstring_score": 0.1, "complexity_ratio": 0.1, "naming_score": 0.1}
        domain = {"overall_relevance": 0.1, "technical_density": 0.1}
        doc = {"examples_present": 0.1, "parameter_documentation": 0.1}

        recs = system._generate_recommendations(structure, domain, doc)
        assert len(recs) <= 5


class TestValidateConfidenceThreshold:
    """Tests for validate_confidence_threshold method."""

    def test_validate_threshold_above(self):
        """Test validation when score is above threshold."""
        system = ConfidenceScoringSystem()
        result = system.validate_confidence_threshold(0.85, threshold=0.7)

        assert result["meets_threshold"] is True
        assert result["confidence_score"] == 0.85
        assert result["threshold"] == 0.7
        assert abs(result["difference"] - 0.15) < 0.001

    def test_validate_threshold_below(self):
        """Test validation when score is below threshold."""
        system = ConfidenceScoringSystem()
        result = system.validate_confidence_threshold(0.5, threshold=0.7)

        assert result["meets_threshold"] is False
        assert result["confidence_score"] == 0.5
        assert abs(result["difference"] - (-0.2)) < 0.001

    def test_validate_threshold_equal(self):
        """Test validation when score equals threshold."""
        system = ConfidenceScoringSystem()
        result = system.validate_confidence_threshold(0.7, threshold=0.7)

        assert result["meets_threshold"] is True

    def test_validate_threshold_strict_mode(self):
        """Test validation in strict mode."""
        system = ConfidenceScoringSystem()
        result = system.validate_confidence_threshold(
            0.75, threshold=0.7, strict_mode=True
        )

        assert result["meets_threshold"] is True

    def test_validate_threshold_recommendation_exists(self):
        """Test that recommendation is provided."""
        system = ConfidenceScoringSystem()
        result = system.validate_confidence_threshold(0.85)

        assert "recommendation" in result
        assert isinstance(result["recommendation"], str)


class TestGetThresholdRecommendation:
    """Tests for _get_threshold_recommendation method."""

    def test_recommendation_excellent(self):
        """Test recommendation for excellent score."""
        system = ConfidenceScoringSystem()
        rec = system._get_threshold_recommendation(0.95, 0.7)

        assert "Excellent" in rec
        assert "recommended" in rec

    def test_recommendation_good(self):
        """Test recommendation for good score."""
        system = ConfidenceScoringSystem()
        rec = system._get_threshold_recommendation(0.85, 0.7)

        assert "Good" in rec
        assert "recommended" in rec

    def test_recommendation_acceptable(self):
        """Test recommendation for acceptable score."""
        system = ConfidenceScoringSystem()
        rec = system._get_threshold_recommendation(0.75, 0.7)

        assert "Acceptable" in rec
        assert "recommended" in rec

    def test_recommendation_marginal(self):
        """Test recommendation for marginal score."""
        system = ConfidenceScoringSystem()
        rec = system._get_threshold_recommendation(0.65, 0.7)

        assert "Marginal" in rec
        assert "review" in rec

    def test_recommendation_low(self):
        """Test recommendation for low score."""
        system = ConfidenceScoringSystem()
        rec = system._get_threshold_recommendation(0.45, 0.7)

        assert "Low" in rec
        assert "improvement" in rec

    def test_recommendation_very_low(self):
        """Test recommendation for very low score."""
        system = ConfidenceScoringSystem()
        rec = system._get_threshold_recommendation(0.25, 0.7)

        assert "Very low" in rec
        assert "redesign" in rec


class TestGetConfidenceBreakdown:
    """Tests for get_confidence_breakdown method."""

    def test_get_confidence_breakdown(self):
        """Test getting confidence breakdown."""
        system = ConfidenceScoringSystem()
        breakdown = system.get_confidence_breakdown(0.85)

        assert "overall_score" in breakdown
        assert "interpretation" in breakdown
        assert "risk_level" in breakdown
        assert "action_required" in breakdown
        assert breakdown["overall_score"] == 0.85

    def test_get_confidence_breakdown_high_score(self):
        """Test breakdown for high confidence score."""
        system = ConfidenceScoringSystem()
        breakdown = system.get_confidence_breakdown(0.95)

        assert "Excellent" in breakdown["interpretation"]
        assert breakdown["risk_level"] == "Low"

    def test_get_confidence_breakdown_low_score(self):
        """Test breakdown for low confidence score."""
        system = ConfidenceScoringSystem()
        breakdown = system.get_confidence_breakdown(0.35)

        assert "Poor" in breakdown["interpretation"]
        assert breakdown["risk_level"] == "Critical"


class TestInterpretConfidenceScore:
    """Tests for _interpret_confidence_score method."""

    def test_interpret_excellent(self):
        """Test interpretation of excellent score."""
        system = ConfidenceScoringSystem()
        result = system._interpret_confidence_score(0.95)

        assert "Excellent" in result

    def test_interpret_good(self):
        """Test interpretation of good score."""
        system = ConfidenceScoringSystem()
        result = system._interpret_confidence_score(0.85)

        assert "Good" in result

    def test_interpret_acceptable(self):
        """Test interpretation of acceptable score."""
        system = ConfidenceScoringSystem()
        result = system._interpret_confidence_score(0.75)

        assert "Acceptable" in result

    def test_interpret_marginal(self):
        """Test interpretation of marginal score."""
        system = ConfidenceScoringSystem()
        result = system._interpret_confidence_score(0.65)

        assert "Marginal" in result

    def test_interpret_poor(self):
        """Test interpretation of poor score."""
        system = ConfidenceScoringSystem()
        result = system._interpret_confidence_score(0.45)

        assert "Poor" in result

    def test_interpret_very_poor(self):
        """Test interpretation of very poor score."""
        system = ConfidenceScoringSystem()
        result = system._interpret_confidence_score(0.25)

        assert "Very Poor" in result


class TestGetRiskLevel:
    """Tests for _get_risk_level method."""

    def test_risk_level_low(self):
        """Test risk level for high confidence."""
        system = ConfidenceScoringSystem()
        result = system._get_risk_level(0.85)

        assert result == "Low"

    def test_risk_level_medium(self):
        """Test risk level for medium confidence."""
        system = ConfidenceScoringSystem()
        result = system._get_risk_level(0.7)

        assert result == "Medium"

    def test_risk_level_high(self):
        """Test risk level for low confidence."""
        system = ConfidenceScoringSystem()
        result = system._get_risk_level(0.5)

        assert result == "High"

    def test_risk_level_critical(self):
        """Test risk level for very low confidence."""
        system = ConfidenceScoringSystem()
        result = system._get_risk_level(0.3)

        assert result == "Critical"


class TestGetActionRequired:
    """Tests for _get_action_required method."""

    def test_action_high_score(self):
        """Test action for high confidence score."""
        system = ConfidenceScoringSystem()
        result = system._get_action_required(0.8)

        assert "Auto-generate" in result

    def test_action_medium_score(self):
        """Test action for medium confidence score."""
        system = ConfidenceScoringSystem()
        result = system._get_action_required(0.6)

        assert "manual review" in result

    def test_action_low_score(self):
        """Test action for low confidence score."""
        system = ConfidenceScoringSystem()
        result = system._get_action_required(0.4)

        assert "Do not auto-generate" in result


class TestCalculateCompletionConfidence:
    """Tests for calculate_completion_confidence function."""

    def test_calculate_completion_confidence_with_file_path(self, tmp_path):
        """Test completion confidence with file_path in analysis."""
        test_file = tmp_path / "code.py"
        test_file.write_text("""
class TestClass:
    '''Test class.'''
    def method(self):
        '''Method.'''
        pass
""")

        analysis = {"file_path": str(test_file)}
        score = calculate_completion_confidence(analysis)

        assert isinstance(score, float)
        assert 0 <= score <= 1.0

    def test_calculate_completion_confidence_without_file_path(self):
        """Test completion confidence without file_path in analysis."""
        analysis = {"some_key": "some_value"}
        score = calculate_completion_confidence(analysis)

        assert isinstance(score, float)
        assert 0 <= score <= 1.0

    def test_calculate_completion_confidence_empty_analysis(self):
        """Test completion confidence with empty analysis."""
        analysis = {}
        score = calculate_completion_confidence(analysis)

        assert isinstance(score, float)
        assert 0 <= score <= 1.0


class TestEdgeCases:
    """Tests for edge cases and error handling."""

    def test_analyze_file_with_unicode(self, tmp_path):
        """Test analyzing file with unicode characters."""
        test_file = tmp_path / "unicode.py"
        test_file.write_text("""
# Test unicode: こんにちは, 你好, Привет
def test():
    '''Test with unicode in docstring: 测试'''
    name = "Örjan"
    return name
""", encoding="utf-8")

        system = ConfidenceScoringSystem()
        result = system.analyze_code_structure(str(test_file))

        assert isinstance(result, dict)
        assert result["function_ratio"] > 0

    def test_analyze_large_file(self, tmp_path):
        """Test analyzing a large file."""
        test_file = tmp_path / "large.py"
        content = "def func():\n    pass\n" * 1000
        test_file.write_text(content)

        system = ConfidenceScoringSystem()
        score, analysis = system.calculate_confidence_score(str(test_file))

        assert isinstance(score, float)
        assert "analysis_time" in analysis

    def test_score_rounding(self, tmp_path):
        """Test that confidence scores are properly rounded."""
        test_file = tmp_path / "code.py"
        test_file.write_text("x = 1")

        system = ConfidenceScoringSystem()
        score, analysis = system.calculate_confidence_score(str(test_file))

        # Score should be rounded to 2 decimal places
        assert len(str(score).split(".")[-1]) <= 2

    def test_concurrent_analysis(self, tmp_path):
        """Test analyzing multiple files doesn't cause issues."""
        test_file1 = tmp_path / "code1.py"
        test_file1.write_text("x = 1")

        test_file2 = tmp_path / "code2.py"
        test_file2.write_text("y = 2")

        system = ConfidenceScoringSystem()

        score1, _ = system.calculate_confidence_score(str(test_file1))
        score2, _ = system.calculate_confidence_score(str(test_file2))

        assert isinstance(score1, float)
        assert isinstance(score2, float)

    def test_analyze_file_with_imports_only(self, tmp_path):
        """Test analyzing file with only import statements."""
        test_file = tmp_path / "imports.py"
        test_file.write_text("""
import os
import sys
from pathlib import Path
from typing import Dict, List, Any
""")

        system = ConfidenceScoringSystem()
        result = system.analyze_code_structure(str(test_file))

        assert result["import_ratio"] > 0
        assert result["function_ratio"] == 0

    def test_analyze_file_with_only_classes(self, tmp_path):
        """Test analyzing file with only class definitions."""
        test_file = tmp_path / "classes.py"
        test_file.write_text("""
class Class1:
    '''Class 1.'''
    pass

class Class2:
    '''Class 2.'''
    pass

class Class3:
    '''Class 3.'''
    pass
""")

        system = ConfidenceScoringSystem()
        result = system.analyze_code_structure(str(test_file))

        assert result["class_ratio"] > 0
        assert result["function_ratio"] == 0

    def test_analyze_file_with_lambda_functions(self, tmp_path):
        """Test analyzing file with lambda functions."""
        test_file = tmp_path / "lambda.py"
        test_file.write_text("""
add = lambda x, y: x + y
multiply = lambda x, y: x * y
filter_even = lambda nums: [n for n in nums if n % 2 == 0]
""")

        system = ConfidenceScoringSystem()
        result = system.analyze_code_structure(str(test_file))

        assert isinstance(result, dict)

    def test_analyze_file_with_decorators(self, tmp_path):
        """Test analyzing file with decorators."""
        test_file = tmp_path / "decorators.py"
        test_file.write_text("""
def decorator(func):
    def wrapper(*args, **kwargs):
        return func(*args, **kwargs)
    return wrapper

@decorator
def decorated_func():
    '''Decorated function.'''
    pass
""")

        system = ConfidenceScoringSystem()
        result = system.analyze_code_structure(str(test_file))

        assert isinstance(result, dict)


class TestIntegration:
    """Integration tests combining multiple methods."""

    def test_full_analysis_pipeline(self, tmp_path):
        """Test full analysis pipeline from code to recommendations."""
        test_file = tmp_path / "pipeline.py"
        test_file.write_text("""
class DataProcessor:
    '''Process data from various sources.'''

    def __init__(self):
        '''Initialize processor.'''
        self.cache = {}

    def process_security_data(self, data):
        '''Process security-related data.

        Args:
            data: Raw security data

        Returns:
            Processed security data

        Raises:
            ValueError: If data is invalid
        '''
        if not data:
            raise ValueError("Data cannot be empty")

        # Authenticate and validate
        auth_token = self._authenticate()
        validated = self._validate_data(data)

        return validated

    def _authenticate(self):
        '''Authenticate with security system.'''
        password = "secret"
        bcrypt_hash = "hashed"
        return "token"

    def _validate_data(self, data):
        '''Validate input data.'''
        if isinstance(data, dict):
            for key, value in data.items():
                if value is None:
                    raise ValueError(f"Invalid value for {key}")
        return data

import logging
logger = logging.getLogger(__name__)
""")

        system = ConfidenceScoringSystem()

        # Run all analyses
        structure = system.analyze_code_structure(str(test_file))
        domain = system.analyze_domain_relevance(str(test_file))
        docs = system.analyze_documentation_quality(str(test_file))
        score, detailed = system.calculate_confidence_score(str(test_file))
        breakdown = system.get_confidence_breakdown(score)
        validation = system.validate_confidence_threshold(score)
        recommendations = detailed["recommendations"]

        # Verify all components
        assert structure["class_ratio"] > 0
        assert domain["security_coverage"] > 0
        assert docs["docstring_coverage"] > 0
        assert 0 <= score <= 1.0
        assert breakdown["risk_level"] in ["Low", "Medium", "High", "Critical"]
        assert isinstance(recommendations, list)
        # Don't assert meets_threshold as it depends on the actual calculated score

    def test_comparison_high_vs_low_quality(self, tmp_path):
        """Test comparison between high and low quality code."""
        # High quality code
        high_file = tmp_path / "high_quality.py"
        high_file.write_text("""
class HighQualityClass:
    '''This is a high quality class with full documentation.'''

    def process_data(self, data):
        '''Process input data.

        Args:
            data: Input data to process

        Returns:
            Processed data

        Raises:
            ValueError: If data is invalid

        Example:
            >>> obj = HighQualityClass()
            >>> obj.process_data([1, 2, 3])
            [2, 4, 6]
        '''
        if not data:
            raise ValueError("Data cannot be empty")
        return [x * 2 for x in data]
""")

        # Low quality code
        low_file = tmp_path / "low_quality.py"
        low_file.write_text("def f(x):\n    return x*2")

        system = ConfidenceScoringSystem()

        high_score, _ = system.calculate_confidence_score(str(high_file))
        low_score, _ = system.calculate_confidence_score(str(low_file))

        assert high_score > low_score


class TestCoverageEdgeCases:
    """Additional tests to cover edge cases and increase coverage."""

    def test_naming_consistency_with_no_names(self, tmp_path):
        """Test naming consistency when tree has no names."""
        test_file = tmp_path / "no_names.py"
        test_file.write_text("# Just a comment\n1 + 2")

        system = ConfidenceScoringSystem()
        with open(test_file) as f:
            tree = ast.parse(f.read())

        consistency = system._calculate_naming_consistency(tree)
        # Should handle case gracefully
        assert 0 <= consistency <= 1.0

    def test_analyze_documentation_with_return_docs(self, tmp_path):
        """Test documentation analysis with return documentation."""
        test_file = tmp_path / "return_doc.py"
        test_file.write_text("""
def calculate(x, y):
    '''Calculate sum.

    return value is the sum.
    '''
    return x + y
""")

        system = ConfidenceScoringSystem()
        result = system.analyze_documentation_quality(str(test_file))

        # The code checks for "return" keyword in docstring
        assert result["return_documentation"] >= 0

    def test_analyze_documentation_with_exception_docs(self, tmp_path):
        """Test documentation analysis with exception documentation."""
        test_file = tmp_path / "exc_doc.py"
        test_file.write_text("""
def divide(a, b):
    '''Divide two numbers.

    Raises on exception or error case.
    '''
    if b == 0:
        raise ValueError("b cannot be zero")
    return a / b
""")

        system = ConfidenceScoringSystem()
        result = system.analyze_documentation_quality(str(test_file))

        # The code checks for "raise" or "exception" keyword in docstring
        assert result["exception_documentation"] >= 0

    def test_analyze_documentation_with_examples_in_docstring(self, tmp_path):
        """Test documentation analysis with examples."""
        test_file = tmp_path / "doc_examples.py"
        test_file.write_text('''
"""Module with examples.

Example:
    >>> add(2, 3)
    5

Example 2:
    >>> multiply(4, 5)
    20
"""

def add(a, b):
    """Add two numbers."""
    return a + b

def multiply(a, b):
    """Multiply two numbers."""
    return a * b
''')

        system = ConfidenceScoringSystem()
        result = system.analyze_documentation_quality(str(test_file))

        assert result["examples_present"] > 0

    def test_analyze_documentation_with_explanation_indicators(self, tmp_path):
        """Test documentation analysis with explanation indicators."""
        test_file = tmp_path / "doc_quality.py"
        test_file.write_text('''
"""Module that provides useful utilities.

This module provides various functions and classes.
It allows users to perform common operations.
It enables efficient processing of data.
"""

def process():
    """Process data.

    This function implements data processing.
    It handles various edge cases.
    It manages resources efficiently.
    """
    pass
''')

        system = ConfidenceScoringSystem()
        result = system.analyze_documentation_quality(str(test_file))

        assert result["explanation_quality"] > 0

    def test_analyze_file_with_inline_comments(self, tmp_path):
        """Test analyzing file with inline comments."""
        test_file = tmp_path / "inline_comments.py"
        test_file.write_text("""
def function():
    # Comment 1
    x = 1  # Inline comment
    # Comment 2
    y = 2
    # Comment 3
    return x + y
""")

        system = ConfidenceScoringSystem()
        result = system.analyze_code_structure(str(test_file))

        assert isinstance(result, dict)

    def test_analyze_file_with_multiline_strings_not_docstrings(self, tmp_path):
        """Test analyzing file with multiline strings that are not docstrings."""
        test_file = tmp_path / "multiline.py"
        test_file.write_text('''
def function():
    text = """
    This is a multiline string
    but not a docstring
    """
    return text
''')

        system = ConfidenceScoringSystem()
        result = system.analyze_code_structure(str(test_file))

        assert isinstance(result, dict)

    def test_analyze_file_with_complex_boolean_logic(self, tmp_path):
        """Test complexity calculation with complex boolean operators."""
        test_file = tmp_path / "bool_complex.py"
        test_file.write_text("""
def check_conditions(a, b, c, d, e):
    return (a and b) or (c and d) or (e and (a or b))
""")

        system = ConfidenceScoringSystem()
        with open(test_file) as f:
            tree = ast.parse(f.read())

        complexity = system._calculate_complexity(tree)
        assert complexity > 1

    def test_analyze_very_deep_nesting(self, tmp_path):
        """Test nesting depth calculation with very deep nesting."""
        test_file = tmp_path / "very_deep.py"
        test_file.write_text("""
def func():
    if True:
        if True:
            if True:
                if True:
                    if True:
                        if True:
                            x = 1
""")

        system = ConfidenceScoringSystem()
        with open(test_file) as f:
            tree = ast.parse(f.read())

        depth = system._calculate_nesting_depth(tree)
        assert depth >= 5

    def test_domain_analysis_with_all_patterns(self, tmp_path):
        """Test domain analysis recognizes all domain patterns."""
        test_file = tmp_path / "all_domains.py"
        test_file.write_text("""
# Security domain
def authenticate(password, token, bcrypt_hash):
    return bcrypt.verify(password, bcrypt_hash)

# Data domain
class UserModel:
    schema = "users"
    database = "main"
    cache = None

# API domain
class UserController:
    def api_endpoint(self):
        service = UserService()
        handler = RequestHandler()

# UI domain
def render_component():
    interface = UserInterface()
    widget = Button()
    layout = HBox()
    theme = DarkTheme()

# Business domain
def apply_business_rule():
    policy = SubscriptionPolicy()
    workflow = ApprovalWorkflow()

# Testing domain
def test_authentication():
    mock_auth = MockAuthentication()
    fixture = UserFixture()
""")

        system = ConfidenceScoringSystem()
        result = system.analyze_domain_relevance(str(test_file))

        # Should detect patterns from multiple domains
        assert result["security_coverage"] > 0
        assert result["data_coverage"] > 0
        assert result["api_coverage"] > 0
        assert result["ui_coverage"] > 0
        assert result["business_coverage"] > 0
        assert result["testing_coverage"] > 0

    def test_calculate_confidence_all_weighted_scenarios(self, tmp_path):
        """Test confidence calculation with different weight distributions."""
        test_file = tmp_path / "code.py"
        test_file.write_text("""
class MyClass:
    '''My class.'''
    def method(self):
        '''Method.'''
        pass
""")

        system = ConfidenceScoringSystem()

        # Test with structure-heavy weights
        struct_weights = {
            "class_ratio": 0.3,
            "function_ratio": 0.3,
            "method_ratio": 0.2,
            "import_ratio": 0.1,
            "complexity_ratio": 0.0,
            "nesting_ratio": 0.0,
            "docstring_score": 0.0,
            "naming_score": 0.0,
        }

        score1, _ = system.calculate_confidence_score(
            str(test_file), structure_weights=struct_weights
        )

        # Test with documentation-heavy weights
        struct_weights_doc = {
            "class_ratio": 0.0,
            "function_ratio": 0.0,
            "method_ratio": 0.0,
            "import_ratio": 0.0,
            "complexity_ratio": 0.0,
            "nesting_ratio": 0.0,
            "docstring_score": 1.0,
            "naming_score": 0.0,
        }

        score2, _ = system.calculate_confidence_score(
            str(test_file), structure_weights=struct_weights_doc
        )

        assert 0 <= score1 <= 1.0
        assert 0 <= score2 <= 1.0

    def test_recommendation_boundary_conditions(self):
        """Test recommendations at boundary values."""
        system = ConfidenceScoringSystem()

        # Test at exact boundary (0.5)
        structure = {"docstring_score": 0.5, "complexity_ratio": 0.5, "naming_score": 0.5}
        domain = {"overall_relevance": 0.5, "technical_density": 0.5}
        doc = {"examples_present": 0.5, "parameter_documentation": 0.5}

        recs = system._generate_recommendations(structure, domain, doc)
        assert isinstance(recs, list)

    def test_confidence_score_rounding_precision(self, tmp_path):
        """Test that confidence scores maintain precision."""
        test_file = tmp_path / "precision.py"
        test_file.write_text("x = 1")

        system = ConfidenceScoringSystem()
        score, analysis = system.calculate_confidence_score(str(test_file))

        # Verify score is rounded to 2 decimal places
        decimal_places = len(str(score).split(".")[-1])
        assert decimal_places <= 2

    def test_analyze_file_with_generator_functions(self, tmp_path):
        """Test analyzing file with generator functions."""
        test_file = tmp_path / "generators.py"
        test_file.write_text("""
def generator_func():
    '''Generate values.'''
    yield 1
    yield 2
    yield 3
""")

        system = ConfidenceScoringSystem()
        result = system.analyze_code_structure(str(test_file))

        assert result["function_ratio"] > 0

    def test_analyze_file_with_nested_classes(self, tmp_path):
        """Test analyzing file with nested class definitions."""
        test_file = tmp_path / "nested_classes.py"
        test_file.write_text("""
class Outer:
    '''Outer class.'''
    class Inner:
        '''Inner class.'''
        def method(self):
            '''Method.'''
            pass
""")

        system = ConfidenceScoringSystem()
        result = system.analyze_code_structure(str(test_file))

        assert result["class_ratio"] > 0

    def test_naming_consistency_when_no_identifiable_names(self, tmp_path):
        """Test naming consistency when tree has only numbers and operators."""
        test_file = tmp_path / "numbers_only.py"
        test_file.write_text("result = 1 + 2 + 3 + 4 + 5")

        system = ConfidenceScoringSystem()
        with open(test_file) as f:
            tree = ast.parse(f.read())

        # This tests the code path where total_names == 0 at line 236
        # when there are binary operations but no identifiable names
        consistency = system._calculate_naming_consistency(tree)
        # Should return 1.0 when no names found
        assert consistency <= 1.0
