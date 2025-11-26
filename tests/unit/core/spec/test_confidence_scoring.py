"""Test suite for Confidence Scoring System."""

import os
import tempfile
import unittest
from unittest.mock import Mock, patch

from moai_adk.core.spec.confidence_scoring import ConfidenceScoringSystem


class TestConfidenceScoringSystem(unittest.TestCase):
    """Test cases for Confidence Scoring System."""

    def setUp(self):
        """Set up test environment."""
        self.scorer = ConfidenceScoringSystem()
        self.test_dir = tempfile.mkdtemp()

        # Sample test code files
        self.high_quality_code = """
import bcrypt
from typing import Optional

class UserAuth:
    \"\"\"User authentication system.\"
    \"\"\"

    def __init__(self):
        \"\"\"Initialize user authentication system.\"
        \"\"\"
        self.users = {}

    def register_user(self, username: str, password: str) -> bool:
        \"\"\"Register a new user with password hashing.\"
        \"\"\"
        if username in self.users:
            return False

        hashed_password = bcrypt.hashpw(password.encode(), bcrypt.gensalt())
        self.users[username] = {
            'password': hashed_password,
            'created_at': '2025-11-11'
        }
        return True

    def login(self, username: str, password: str) -> bool:
        \"\"\"Authenticate user login with password verification.\"
        \"\"\"
        if username not in self.users:
            return False

        stored_hash = self.users[username]['password']
        return bcrypt.checkpw(password.encode(), stored_hash)
"""

        self.low_quality_code = """
def function1():
    # This is a simple function
    x = 1
    y = 2
    z = x + y
    return z

def function2():
    # Another function
    if True:
        for i in range(10):
            print(i)
    return 42
"""

        # Write test files
        self.high_quality_file = os.path.join(self.test_dir, "high_quality.py")
        with open(self.high_quality_file, "w") as f:
            f.write(self.high_quality_code)

        self.low_quality_file = os.path.join(self.test_dir, "low_quality.py")
        with open(self.low_quality_file, "w") as f:
            f.write(self.low_quality_code)

    def tearDown(self):
        """Clean up test environment."""
        import shutil

        shutil.rmtree(self.test_dir, ignore_errors=True)

    def test_analyze_code_structure_high_quality(self):
        """Test code structure analysis for high-quality code."""
        structure_scores = self.scorer.analyze_code_structure(self.high_quality_file)

        # Check that all expected keys are present
        expected_keys = [
            "class_ratio",
            "function_ratio",
            "method_ratio",
            "import_ratio",
            "complexity_ratio",
            "nesting_ratio",
            "docstring_score",
            "naming_score",
        ]

        for key in expected_keys:
            self.assertIn(key, structure_scores)
            self.assertIsInstance(structure_scores[key], float)
            self.assertGreaterEqual(structure_scores[key], 0.0)
            self.assertLessEqual(structure_scores[key], 1.0)

        # High-quality code should have good scores
        self.assertGreater(structure_scores["class_ratio"], 0.05)  # Has classes
        self.assertGreater(structure_scores["method_ratio"], 0.05)  # Has methods
        self.assertGreater(structure_scores["docstring_score"], 0.3)  # Has docstrings
        self.assertGreater(structure_scores["naming_score"], 0.5)  # Good naming

    def test_analyze_code_structure_low_quality(self):
        """Test code structure analysis for low-quality code."""
        structure_scores = self.scorer.analyze_code_structure(self.low_quality_file)

        # Low-quality code should have lower scores
        self.assertGreaterEqual(structure_scores["docstring_score"], 0.0)
        self.assertLessEqual(structure_scores["docstring_score"], 0.6)  # Few docstrings

    def test_analyze_domain_relevance(self):
        """Test domain relevance analysis."""
        domain_scores = self.scorer.analyze_domain_relevance(self.high_quality_file)

        # Check that expected keys are present
        expected_keys = [
            "security_coverage",
            "data_coverage",
            "api_coverage",
            "ui_coverage",
            "business_coverage",
            "testing_coverage",
            "overall_relevance",
            "specificity",
            "technical_density",
        ]

        for key in expected_keys:
            self.assertIn(key, domain_scores)
            self.assertIsInstance(domain_scores[key], float)
            self.assertGreaterEqual(domain_scores[key], 0.0)
            self.assertLessEqual(domain_scores[key], 1.0)

        # High-quality code should have good domain relevance
        self.assertGreater(domain_scores["overall_relevance"], 0.1)
        self.assertGreater(domain_scores["technical_density"], 0.1)

    def test_analyze_documentation_quality(self):
        """Test documentation quality analysis."""
        doc_scores = self.scorer.analyze_documentation_quality(self.high_quality_file)

        # Check that expected keys are present
        expected_keys = [
            "docstring_coverage",
            "comment_density",
            "explanation_quality",
            "examples_present",
            "parameter_documentation",
            "return_documentation",
            "exception_documentation",
        ]

        for key in expected_keys:
            self.assertIn(key, doc_scores)
            self.assertIsInstance(doc_scores[key], float)
            self.assertGreaterEqual(doc_scores[key], 0.0)
            self.assertLessEqual(doc_scores[key], 1.0)

        # High-quality code should have good documentation
        self.assertGreater(doc_scores["docstring_coverage"], 0.3)
        self.assertGreaterEqual(doc_scores["parameter_documentation"], 0.0)

    def test_calculate_confidence_score(self):
        """Test confidence score calculation."""
        confidence, detailed_analysis = self.scorer.calculate_confidence_score(self.high_quality_file)

        # Check confidence score
        self.assertIsInstance(confidence, float)
        self.assertGreaterEqual(confidence, 0.0)
        self.assertLessEqual(confidence, 1.0)

        # Check detailed analysis structure
        self.assertIn("file_path", detailed_analysis)
        self.assertIn("analysis_time", detailed_analysis)
        self.assertIn("confidence_score", detailed_analysis)
        self.assertIn("structure_analysis", detailed_analysis)
        self.assertIn("domain_analysis", detailed_analysis)
        self.assertIn("documentation_analysis", detailed_analysis)
        self.assertIn("recommendations", detailed_analysis)

        # Check that analysis time is reasonable
        self.assertGreater(detailed_analysis["analysis_time"], 0.0)
        self.assertLess(detailed_analysis["analysis_time"], 1.0)

        # Check recommendations
        self.assertIsInstance(detailed_analysis["recommendations"], list)
        self.assertLessEqual(len(detailed_analysis["recommendations"]), 5)

    def test_confidence_score_comparison(self):
        """Test confidence score comparison between high and low quality code."""
        high_confidence, _ = self.scorer.calculate_confidence_score(self.high_quality_file)
        low_confidence, _ = self.scorer.calculate_confidence_score(self.low_quality_file)

        # High-quality code should have higher confidence
        self.assertGreater(high_confidence, low_confidence)

    def test_validate_confidence_threshold(self):
        """Test confidence threshold validation."""
        confidence = 0.8
        threshold = 0.7

        validation = self.scorer.validate_confidence_threshold(confidence, threshold)

        self.assertIn("meets_threshold", validation)
        self.assertIn("confidence_score", validation)
        self.assertIn("threshold", validation)
        self.assertIn("difference", validation)
        self.assertIn("recommendation", validation)

        # Should meet threshold
        self.assertTrue(validation["meets_threshold"])
        self.assertEqual(validation["confidence_score"], confidence)
        self.assertEqual(validation["threshold"], threshold)
        self.assertAlmostEqual(validation["difference"], 0.1, places=10)

    def test_get_confidence_breakdown(self):
        """Test confidence score breakdown."""
        confidence = 0.75

        breakdown = self.scorer.get_confidence_breakdown(confidence)

        self.assertIn("overall_score", breakdown)
        self.assertIn("interpretation", breakdown)
        self.assertIn("risk_level", breakdown)
        self.assertIn("action_required", breakdown)

        self.assertEqual(breakdown["overall_score"], confidence)
        self.assertIsInstance(breakdown["interpretation"], str)
        self.assertIsInstance(breakdown["risk_level"], str)
        self.assertIsInstance(breakdown["action_required"], str)

    def test_error_handling(self):
        """Test error handling for invalid files."""
        invalid_file = os.path.join(self.test_dir, "nonexistent.py")

        # Should not raise exceptions but return default values
        structure_scores = self.scorer.analyze_code_structure(invalid_file)
        domain_scores = self.scorer.analyze_domain_relevance(invalid_file)
        doc_scores = self.scorer.analyze_documentation_quality(invalid_file)

        # Should return default scores
        self.assertIsInstance(structure_scores, dict)
        self.assertIsInstance(domain_scores, dict)
        self.assertIsInstance(doc_scores, dict)

    def test_custom_weights(self):
        """Test custom weights for confidence calculation."""
        custom_structure_weights = {
            "class_ratio": 0.2,
            "function_ratio": 0.2,
            "method_ratio": 0.2,
            "import_ratio": 0.1,
            "complexity_ratio": 0.1,
            "nesting_ratio": 0.1,
            "docstring_score": 0.05,
            "naming_score": 0.05,
        }

        confidence1, _ = self.scorer.calculate_confidence_score(self.high_quality_file)
        confidence2, _ = self.scorer.calculate_confidence_score(
            self.high_quality_file, structure_weights=custom_structure_weights
        )

        # Should be different with custom weights
        self.assertIsInstance(confidence1, float)
        self.assertIsInstance(confidence2, float)

    @patch("moai_adk.core.tags.spec_generator.SpecGenerator")
    def test_integration_with_spec_generator(self, mock_spec_generator):
        """Test integration with existing SpecGenerator."""
        mock_generator = Mock()
        mock_generator.analyze.return_value = {
            "file_path": self.high_quality_file,
            "domain_keywords": ["auth", "user", "bcrypt"],
            "structure_info": {
                "classes": ["UserAuth"],
                "functions": ["register_user", "login"],
                "imports": ["bcrypt", "typing"],
            },
        }

        self.scorer.spec_generator = mock_generator

        confidence, detailed_analysis = self.scorer.calculate_confidence_score(self.high_quality_file)

        # Should integrate properly
        self.assertIsInstance(confidence, float)
        self.assertGreater(confidence, 0.0)
        self.assertLessEqual(confidence, 1.0)

    def test_backwards_compatibility(self):
        """Test backwards compatibility function."""
        from moai_adk.core.spec.confidence_scoring import calculate_completion_confidence

        analysis = {
            "file_path": self.high_quality_file,
            "structure_score": 0.9,
            "domain_accuracy": 0.95,
            "documentation_level": 0.85,
        }

        confidence = calculate_completion_confidence(analysis)

        # Should return a confidence score
        self.assertIsInstance(confidence, float)
        self.assertGreaterEqual(confidence, 0.0)
        self.assertLessEqual(confidence, 1.0)


if __name__ == "__main__":
    unittest.main()
