# @TEST:SPEC-EARS-TEMPLATE-001
"""Test suite for EARS Template Engine."""

import unittest
import tempfile
import os
import json
from unittest.mock import Mock, patch, MagicMock
from pathlib import Path

from moai_adk.core.spec.ears_template_engine import EARSTemplateEngine


class TestEARSTemplateEngine(unittest.TestCase):
    """Test cases for EARS Template Engine."""

    def setUp(self):
        """Set up test environment."""
        self.engine = EARSTemplateEngine()
        self.test_dir = tempfile.mkdtemp()

        # Sample code analysis results for different domains
        self.web_analysis = {
            'file_path': os.path.join(self.test_dir, 'web_service.py'),
            'analysis_time': 0.5,
            'domain_keywords': ['flask', 'http', 'api', 'request', 'response'],
            'structure_info': {
                'classes': ['APIServer', 'User'],
                'functions': ['app.route', 'create_user', 'get_user'],
                'imports': ['flask', 'request', 'jsonify']
            },
            'technical_density': 0.85,
            'overall_relevance': 0.9
        }

        self.ml_analysis = {
            'file_path': os.path.join(self.test_dir, 'model_training.py'),
            'analysis_time': 1.2,
            'domain_keywords': ['tensorflow', 'model', 'training', 'accuracy', 'dataset'],
            'structure_info': {
                'classes': ['NeuralNetwork', 'Dataset'],
                'functions': ['train_model', 'evaluate_model', 'predict'],
                'imports': ['tensorflow', 'numpy', 'sklearn']
            },
            'technical_density': 0.95,
            'overall_relevance': 0.95
        }

        self.low_quality_analysis = {
            'file_path': os.path.join(self.test_dir, 'simple_script.py'),
            'analysis_time': 0.1,
            'domain_keywords': ['print', 'simple', 'basic'],
            'structure_info': {
                'classes': [],
                'functions': ['main', 'calculate'],
                'imports': ['sys']
            },
            'technical_density': 0.2,
            'overall_relevance': 0.3
        }

    def tearDown(self):
        """Clean up test environment."""
        import shutil
        shutil.rmtree(self.test_dir, ignore_errors=True)

    def test_engine_initialization(self):
        """Test EARS engine initialization."""
        self.assertIsInstance(self.engine, EARSTemplateEngine)
        self.assertIn('auth', self.engine.domain_templates)
        self.assertIn('api', self.engine.domain_templates)
        self.assertIn('data', self.engine.domain_templates)
        self.assertIn('ui', self.engine.domain_templates)

    def test_determine_domain_web(self):
        """Test domain detection for web applications."""
        domain = self.engine._determine_domain(self.web_analysis)
        self.assertEqual(domain, 'api')  # web analysis should be detected as api domain

    def test_determine_domain_ml(self):
        """Test domain detection for machine learning."""
        domain = self.engine._determine_domain(self.ml_analysis)
        self.assertEqual(domain, 'data')  # ml analysis should be detected as data domain

    def test_determine_domain_fallback(self):
        """Test fallback domain detection."""
        analysis = {
            'file_path': os.path.join(self.test_dir, 'unknown.py'),
            'domain_keywords': ['custom', 'unique']
        }
        domain = self.engine._determine_domain(analysis)
        self.assertEqual(domain, 'general')  # Should fallback to general

    def test_generate_complete_spec(self):
        """Test complete SPEC generation."""
        result = self.engine.generate_complete_spec(
            self.web_analysis,
            self.web_analysis['file_path']
        )

        # Check all required files are generated
        self.assertIn('spec_md', result)
        self.assertIn('plan_md', result)
        self.assertIn('acceptance_md', result)

        # Check content validity
        self.assertIn('@META:', result['spec_md'])
        self.assertIn('Implementation Plan', result['plan_md'])
        self.assertIn('Acceptance Criteria', result['acceptance_md'])

        # Check file paths
        self.assertIn('web_service.py', result['spec_md'])

    def test_generate_complete_spec_with_ml(self):
        """Test complete SPEC generation for ML domain."""
        result = self.engine.generate_complete_spec(
            self.ml_analysis,
            self.ml_analysis['file_path']
        )

        # Check all required files are generated
        self.assertIn('spec_md', result)
        self.assertIn('plan_md', result)
        self.assertIn('acceptance_md', result)

        # Check content validity
        self.assertIn('@META:', result['spec_md'])
        self.assertIn('Implementation Plan', result['plan_md'])
        self.assertIn('Acceptance Criteria', result['acceptance_md'])

    def test_validate_ears_compliance(self):
        """Test EARS compliance validation."""
        # Generate a spec first
        result = self.engine.generate_complete_spec(
            self.web_analysis,
            self.web_analysis['file_path']
        )

        validation = self.engine._validate_ears_compliance(result)

        # Should return validation results
        self.assertIn('ears_compliance', validation)
        self.assertIn('section_scores', validation)
        self.assertIn('suggestions', validation)
        self.assertIn('total_sections', validation)
        self.assertIn('present_sections', validation)

        # Compliance should be between 0 and 1
        self.assertGreaterEqual(validation['ears_compliance'], 0.0)
        self.assertLessEqual(validation['ears_compliance'], 1.0)

    def test_validate_ears_compliance_low_quality(self):
        """Test EARS compliance validation for low-quality spec."""
        # Create low-quality spec
        low_quality_spec = {
            'spec_md': 'Invalid spec content',
            'plan_md': '',
            'acceptance_md': ''
        }

        validation = self.engine._validate_ears_compliance(low_quality_spec)

        # Low-quality spec should have low compliance
        self.assertLess(validation['ears_compliance'], 1.0)
        self.assertEqual(len(validation['present_sections']), 0)

    def test_error_handling_invalid_analysis(self):
        """Test error handling for invalid analysis data."""
        invalid_analysis = {
            'file_path': 'invalid.py',
            'domain_keywords': []
        }

        # Should not raise exceptions
        result = self.engine.generate_complete_spec(
            invalid_analysis,
            invalid_analysis['file_path']
        )
        self.assertIn('spec_md', result)
        self.assertIn('plan_md', result)
        self.assertIn('acceptance_md', result)

    def test_extract_information_from_analysis(self):
        """Test information extraction from analysis."""
        extraction = self.engine._extract_information_from_analysis(
            self.web_analysis,
            self.web_analysis['file_path']
        )

        # Check extracted information
        self.assertIn('file_name', extraction)
        self.assertIn('domain_keywords', extraction)
        self.assertIn('structure_info', extraction)
        self.assertIn('technical_density', extraction)
        self.assertIn('overall_relevance', extraction)

    def test_generate_spec_id(self):
        """Test SPEC ID generation."""
        extraction = self.engine._extract_information_from_analysis(
            self.web_analysis,
            self.web_analysis['file_path']
        )

        domain = self.engine._determine_domain(extraction)
        spec_id = self.engine._generate_spec_id(extraction, domain)

        # Should generate valid spec id
        self.assertIsInstance(spec_id, str)
        self.assertTrue(spec_id.startswith('WEB'))
        self.assertGreater(len(spec_id), 3)

    def test_detect_framework(self):
        """Test framework detection."""
        # Flask detection
        framework = self.engine._detect_framework(self.web_analysis)
        self.assertEqual(framework, 'Flask')

        # Custom framework detection
        custom_analysis = {
            'imports': ['custom_lib', 'another_lib']
        }
        framework = self.engine._detect_framework(custom_analysis)
        self.assertEqual(framework, 'Custom')

    def test_detect_language(self):
        """Test language detection."""
        # Python detection
        language = self.engine._detect_language(self.web_analysis)
        self.assertEqual(language, 'Python')

        # JavaScript detection
        js_analysis = {
            'imports': ['react', 'express']
        }
        language = self.engine._detect_language(js_analysis)
        self.assertEqual(language, 'JavaScript')

    def test_template_cache(self):
        """Test template caching mechanism."""
        # Generate spec twice
        result1 = self.engine.generate_complete_spec(
            self.web_analysis,
            self.web_analysis['file_path']
        )

        result2 = self.engine.generate_complete_spec(
            self.web_analysis,
            self.web_analysis['file_path']
        )

        # Results should be the same
        self.assertEqual(result1['spec_md'], result2['spec_md'])
        self.assertEqual(result1['plan_md'], result2['plan_md'])
        self.assertEqual(result1['acceptance_md'], result2['acceptance_md'])

    def test_confidence_scorer_integration(self):
        """Test integration with confidence scorer."""
        self.assertIsNotNone(self.engine.confidence_scorer)
        self.assertEqual(self.engine.confidence_scorer.__class__.__name__, 'ConfidenceScoringSystem')

    def test_domain_templates_access(self):
        """Test access to domain templates."""
        self.assertIsInstance(self.engine.domain_templates, dict)
        self.assertGreater(len(self.engine.domain_templates), 0)

    def test_ear_templates_access(self):
        """Test access to EARS templates."""
        self.assertIsInstance(self.engine.ears_templates, dict)
        self.assertGreater(len(self.engine.ears_templates), 0)

    def test_spec_content_contains_required_sections(self):
        """Test that generated spec contains required sections."""
        result = self.engine.generate_complete_spec(
            self.web_analysis,
            self.web_analysis['file_path']
        )

        spec_content = result['spec_md']

        # Check for required sections
        self.assertIn('개요 (Overview)', spec_content)
        self.assertIn('환경 (Environment)', spec_content)
        self.assertIn('가정 (Assumptions)', spec_content)
        self.assertIn('요구사항 (Requirements)', spec_content)
        self.assertIn('명세 (Specifications)', spec_content)
        self.assertIn('추적성 (Traceability)', spec_content)

    def test_plan_content_contains_phases(self):
        """Test that generated plan contains implementation phases."""
        result = self.engine.generate_complete_spec(
            self.web_analysis,
            self.web_analysis['file_path']
        )

        plan_content = result['plan_md']

        # Check for implementation phases
        self.assertIn('1단계: 요구사항 분석', plan_content)
        self.assertIn('2단계: 설계', plan_content)
        self.assertIn('3단계: 개발', plan_content)
        self.assertIn('4단계: 테스트', plan_content)
        self.assertIn('5단계: 배포', plan_content)

    def test_acceptance_content_contains_criteria(self):
        """Test that generated acceptance contains criteria."""
        result = self.engine.generate_complete_spec(
            self.web_analysis,
            self.web_analysis['file_path']
        )

        acceptance_content = result['acceptance_md']

        # Check for acceptance criteria
        self.assertIn('검수 기준 (Acceptance Criteria)', acceptance_content)
        self.assertIn('기본 기능 검수', acceptance_content)
        self.assertIn('성능 검수', acceptance_content)
        self.assertIn('보안 검수', acceptance_content)

    def test_meta_information_inclusion(self):
        """Test that meta information is included."""
        result = self.engine.generate_complete_spec(
            self.web_analysis,
            self.web_analysis['file_path']
        )

        # Check meta information in spec
        self.assertIn('@META:', result['spec_md'])
        self.assertIn('id', result['spec_md'])
        self.assertIn('spec_id', result['spec_md'])
        self.assertIn('title', result['spec_md'])
        self.assertIn('version', result['spec_md'])
        self.assertIn('status', result['spec_md'])
        self.assertIn('created', result['spec_md'])
        self.assertIn('author', result['spec_md'])
        self.assertIn('domain', result['spec_md'])

    def test_custom_config_override(self):
        """Test custom configuration override."""
        custom_config = {
            'custom_param': 'test_value',
            'another_param': 'another_value'
        }

        result = self.engine.generate_complete_spec(
            self.web_analysis,
            self.web_analysis['file_path'],
            custom_config
        )

        # Should not raise exceptions
        self.assertIn('spec_md', result)
        self.assertIn('plan_md', result)
        self.assertIn('acceptance_md', result)


if __name__ == '__main__':
    unittest.main()