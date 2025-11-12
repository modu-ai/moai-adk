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
        self.assertIsInstance(self.engine, EARSEngine)
        self.assertIn('web', self.engine.domains)
        self.assertIn('mobile', self.engine.domains)
        self.assertIn('ml', self.engine.domains)
        self.assertIn('api', self.engine.domains)

    def test_detect_domain_web(self):
        """Test domain detection for web applications."""
        domain = self.engine._detect_domain(self.web_analysis)
        self.assertEqual(domain, 'web')
        self.assertIn('flask', self.engine.domains[domain]['keywords'])
        self.assertIn('http', self.engine.domains[domain]['keywords'])

    def test_detect_domain_ml(self):
        """Test domain detection for machine learning."""
        domain = self.engine._detect_domain(self.ml_analysis)
        self.assertEqual(domain, 'ml')
        self.assertIn('tensorflow', self.engine.domains[domain]['keywords'])
        self.assertIn('model', self.engine.domains[domain]['keywords'])

    def test_detect_domain_fallback(self):
        """Test fallback domain detection."""
        analysis = {
            'file_path': os.path.join(self.test_dir, 'unknown.py'),
            'domain_keywords': ['custom', 'unique']
        }
        domain = self.engine._detect_domain(analysis)
        self.assertEqual(domain, 'general')

    def test_generate_spec_content_web(self):
        """Test spec content generation for web domain."""
        spec_content = self.engine._generate_spec_content(self.web_analysis, 'web')

        # Check required sections
        self.assertIn('Overview', spec_content)
        self.assertIn('Environment', spec_content)
        self.assertIn('Assumptions', spec_content)
        self.assertIn('Requirements', spec_content)
        self.assertIn('Specifications', spec_content)

        # Check web-specific content
        self.assertIn('Web Application', spec_content)
        self.assertIn('Flask', spec_content)
        self.assertIn('HTTP', spec_content)
        self.assertIn('REST API', spec_content)

        # Check extracted information
        self.assertIn('APIServer', spec_content)
        self.assertIn('User', spec_content)
        self.assertIn('create_user', spec_content)
        self.assertIn('get_user', spec_content)

    def test_generate_spec_content_ml(self):
        """Test spec content generation for ML domain."""
        spec_content = self.engine._generate_spec_content(self.ml_analysis, 'ml')

        # Check ML-specific content
        self.assertIn('Machine Learning', spec_content)
        self.assertIn('TensorFlow', spec_content)
        self.assertIn('Neural Network', spec_content)
        self.assertIn('Model Training', spec_content)

        # Check extracted information
        self.assertIn('NeuralNetwork', spec_content)
        self.assertIn('Dataset', spec_content)
        self.assertIn('train_model', spec_content)
        self.assertIn('evaluate_model', spec_content)

    def test_generate_plan_content(self):
        """Test implementation plan generation."""
        plan_content = self.engine._generate_plan_content(self.web_analysis, 'web')

        # Check required sections
        self.assertIn('Implementation Plan', plan_content)
        self.assertIn('Phase 1: Environment Setup', plan_content)
        self.assertIn('Phase 2: Core Implementation', plan_content)
        self.assertIn('Phase 3: Testing and Validation', plan_content)
        self.assertIn('Phase 4: Documentation and Deployment', plan_content)

        # Check web-specific tasks
        self.assertIn('Flask application setup', plan_content)
        self.assertIn('API endpoints implementation', plan_content)
        self.assertIn('User authentication system', plan_content)

    def test_generate_acceptance_criteria(self):
        """Test acceptance criteria generation."""
        acceptance = self.engine._generate_acceptance_criteria(self.web_analysis, 'web')

        # Check required sections
        self.assertIn('Acceptance Criteria', acceptance)
        self.assertIn('Functional Requirements', acceptance)
        self.assertIn('Non-Functional Requirements', acceptance)
        self.assertIn('Technical Requirements', acceptance)

        # Check web-specific criteria
        self.assertIn('API endpoints', acceptance)
        self.assertIn('HTTP status codes', acceptance)
        self.assertIn('Response formats', acceptance)

    def test_generate_complete_spec(self):
        """Test complete SPEC generation."""
        result = self.engine.generate_complete_spec(self.web_analysis)

        # Check all required files are generated
        self.assertIn('spec_md', result)
        self.assertIn('plan_md', result)
        self.assertIn('acceptance_md', result)

        # Check content validity
        self.assertIn('Implementation Plan', result['plan_md'])
        self.assertIn('Acceptance Criteria', result['acceptance_md'])

        # Check file paths
        self.assertIn('web_service.py', result['spec_md'])

    def test_spec_validation_high_quality(self):
        """Test SPEC validation for high-quality generated specs."""
        result = self.engine.generate_complete_spec(self.web_analysis)
        validation = self.engine.validate_generated_spec(result)

        # High-quality spec should pass validation
        self.assertGreater(validation['quality_score'], 0.8)
        self.assertTrue(validation['ears_compliance'])
        self.assertIn('suggestions', validation)

    def test_spec_validation_low_quality(self):
        """Test SPEC validation for low-quality specs."""
        # Create low-quality spec
        low_quality_spec = {
            'spec_md': 'Invalid spec content',
            'plan_md': '',
            'acceptance_md': ''
        }
        validation = self.engine.validate_generated_spec(low_quality_spec)

        # Low-quality spec should fail validation
        self.assertLess(validation['quality_score'], 0.5)
        self.assertFalse(validation['ears_compliance'])

    def test_domain_specific_templates(self):
        """Test domain-specific template generation."""
        web_result = self.engine.generate_complete_spec(self.web_analysis)
        ml_result = self.engine.generate_complete_spec(self.ml_analysis)

        # Web spec should contain web-specific content
        self.assertIn('Flask', web_result['spec_md'])
        self.assertIn('REST API', web_result['spec_md'])

        # ML spec should contain ML-specific content
        self.assertIn('TensorFlow', ml_result['spec_md'])
        self.assertIn('Neural Network', ml_result['spec_md'])

        # Content should be different between domains
        self.assertNotEqual(web_result['spec_md'], ml_result['spec_md'])

    def test_error_handling_invalid_analysis(self):
        """Test error handling for invalid analysis data."""
        invalid_analysis = {
            'file_path': 'invalid.py',
            'domain_keywords': []
        }

        # Should not raise exceptions
        result = self.engine.generate_complete_spec(invalid_analysis)
        self.assertIn('spec_md', result)
        self.assertIn('plan_md', result)
        self.assertIn('acceptance_md', result)

    def test_meta_information_generation(self):
        """Test meta information generation."""
        result = self.engine.generate_complete_spec(self.web_analysis)

        # Check meta information
        self.assertIn('SPEC-ID', result['spec_md'])
        self.assertIn('Generated', result['spec_md'])
        self.assertIn('Analysis Time', result['spec_md'])
        self.assertIn('Domain', result['spec_md'])

    def test_traceability_tags(self):
        """Test traceability tag generation."""
        result = self.engine.generate_complete_spec(self.web_analysis)

        # Check for traceability tags

    def test_file_path_extraction(self):
        """Test file path extraction and inclusion."""
        result = self.engine.generate_complete_spec(self.web_analysis)

        # Should include the actual file path
        self.assertIn('web_service.py', result['spec_md'])

    def test_execution_time_tracking(self):
        """Test execution time tracking."""
        result = self.engine.generate_complete_spec(self.web_analysis)

        # Should track execution time
        self.assertIn('Analysis Time', result['spec_md'])
        self.assertIn('0.5', result['spec_md'])  # From test data

    def test_confidence_threshold_integration(self):
        """Test integration with confidence scoring."""
        # High confidence analysis
        high_conf_result = self.engine.generate_complete_spec(self.web_analysis)
        high_conf_validation = self.engine.validate_generated_spec(high_conf_result)

        # Low confidence analysis
        low_conf_result = self.engine.generate_complete_spec(self.low_quality_analysis)
        low_conf_validation = self.engine.validate_generated_spec(low_conf_result)

        # High confidence should yield better quality score
        self.assertGreater(high_conf_validation['quality_score'],
                          low_conf_validation['quality_score'])

    def test_template_variable_substitution(self):
        """Test template variable substitution."""
        result = self.engine.generate_complete_spec(self.web_analysis)

        # Check that variables are properly substituted
        self.assertIn('APIServer', result['spec_md'])  # From structure_info
        self.assertIn('create_user', result['spec_md'])  # From structure_info
        self.assertIn('flask', result['spec_md'])  # From domain_keywords

    def test_custom_domain_template(self):
        """Test custom domain template generation."""
        # Create custom domain configuration
        custom_domain = {
            'name': 'blockchain',
            'keywords': ['ethereum', 'smart-contract', 'web3'],
            'templates': {
                'spec': '''# BLOCKCHAIN SPEC

## Overview
{overview}

## Environment
{environment}

## Requirements
{requirements}

## Specifications
{specifications}''',
                'plan': '''# BLOCKCHAIN PLAN
{plan}''',
                'acceptance': '''# BLOCKCHAIN ACCEPTANCE
{acceptance}'''
            }
        }

        # Add custom domain to engine
        self.engine.domains['blockchain'] = custom_domain

        # Test custom domain generation
        analysis = {
            'file_path': os.path.join(self.test_dir, 'smart_contract.py'),
            'domain_keywords': ['ethereum', 'smart-contract', 'web3'],
            'structure_info': {
                'classes': ['SmartContract'],
                'functions': ['deploy', 'interact'],
                'imports': ['web3', 'eth-utils']
            }
        }

        result = self.engine.generate_complete_spec(analysis)
        self.assertIn('BLOCKCHAIN', result['spec_md'])
        self.assertIn('SmartContract', result['spec_md'])

    @patch('moai_adk.core.spec.ears_template_engine.EARSEngine')
    def test_integration_with_confidence_scoring(self, mock_engine_class):
        """Test integration with confidence scoring system."""
        mock_engine = Mock()
        mock_engine_class.return_value = mock_engine
        mock_engine.generate_complete_spec.return_value = {
            'spec_md': '# Test Spec\n## Overview\nTest content',
            'plan_md': '# Test Plan\n## Steps\n1. Step 1',
            'acceptance_md': '# Test Acceptance\n## Criteria\n1. Criteria 1'
        }

        # Test integration
        analysis = {
            'file_path': 'test.py',
            'confidence_score': 0.85,
            'structure_score': 0.8,
            'domain_accuracy': 0.9
        }

        result = mock_engine.generate_complete_spec(analysis)
        self.assertIsInstance(result, dict)
        self.assertIn('spec_md', result)

    def test_ears_format_compliance(self):
        """Test EARS format compliance."""
        result = self.engine.generate_complete_spec(self.web_analysis)

        # Check EARS sections are present
        self.assertIn('Overview', result['spec_md'])
        self.assertIn('Environment', result['spec_md'])
        self.assertIn('Assumptions', result['spec_md'])
        self.assertIn('Requirements', result['spec_md'])
        self.assertIn('Specifications', result['spec_md'])

        # Check plan sections
        self.assertIn('Implementation Plan', result['plan_md'])
        self.assertIn('Phase', result['plan_md'])

        # Check acceptance sections
        self.assertIn('Acceptance Criteria', result['acceptance_md'])
        self.assertIn('Requirements', result['acceptance_md'])


if __name__ == '__main__':
    unittest.main()