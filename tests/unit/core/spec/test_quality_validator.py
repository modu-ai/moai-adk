# @TEST:QUALITY-VALIDATOR-001
"""Test suite for Quality Validator."""

import unittest
import tempfile
import os
import json
from unittest.mock import Mock, patch, MagicMock
from pathlib import Path

from moai_adk.core.spec.quality_validator import QualityValidator


class TestQualityValidator(unittest.TestCase):
    """Test cases for Quality Validator."""

    def setUp(self):
        """Set up test environment."""
        self.validator = QualityValidator()
        self.test_dir = tempfile.mkdtemp()

        # Sample high-quality SPEC content
        self.high_quality_spec = {
            'spec_md': '''# @SPEC:TEST-001: High Quality SPEC
## 개요 (Overview)
This is a comprehensive test specification for a high-quality system.

## 환경 (Environment)
- Python 3.10
- Flask Framework
- PostgreSQL Database
- Redis Cache

## 가정 (Assumptions)
- System will handle up to 1000 concurrent users
- Database connections are stable
- Network latency is under 100ms

## 요구사항 (Requirements)
- REQ-001: User authentication system
- REQ-002: Data validation
- REQ-003: Error handling

## 명세 (Specifications)
- SPEC-001: API endpoints implementation
- SPEC-002: Database schema design
- SPEC-003: Security measures

## 추적성 (Traceability)
- @SPEC:TEST-001 ← @CODE:HOOK-POST-AUTO-SPEC-001
- @SPEC:TEST-001 → @TEST:TEST-001
- @SPEC:TEST-001 → @CODE:TEST-001''',
            'plan_md': '''# @PLAN:TEST-001: Implementation Plan
## 구현 단계 (Implementation Phases)
### 1단계: 요구사항 분석 (Priority: High)
- [ ] 기능 요구사항 상세화
- [ ] 비기능 요구사항 정의

### 2단계: 설계 (Priority: High)
- [ ] 아키텍처 설계 완료
- [ ] 데이터 모델 설계

### 3단계: 개발 (Priority: Medium)
- [ ] 핵심 모듈 개발
- [ ] API 개발 완료

### 4단계: 테스트 (Priority: High)
- [ ] 단위 테스트 구현
- [ ] 통합 테스트 구현

### 5단계: 배포 (Priority: Medium)
- [ ] 스테이징 환경 배포
- [ ] 배포 자동화 구현

## 기술적 접근 방식 (Technical Approach)
- 아키텍처: Microservices
- 프레임워크: Flask
- 데이터베이스: PostgreSQL

## 성공 기준 (Success Criteria)
- 모든 요구사항 구현 완료
- 테스트 커버리지 85% 이상
- 성능 목표 충족

## 다음 단계 (Next Steps)
1. **즉시 실행**: 요구사항 분석 (1-2일)
2. **주간 목표**: 설계 완료 (3-5일)
3. **2주 목표**: 개발 완료 (7-14일)''',
            'acceptance_md': '''# @ACCEPT:TEST-001: Acceptance Criteria
## 검수 기준 (Acceptance Criteria)

### 기본 기능 검수 (Basic Functionality)

**필수 조건 (Must-have):**
- [ ] 시스템이 정상적으로 구동되어야 함
- [ ] 사용자 인터페이스가 올바르게 표시되어야 함
- [ ] 데이터 처리 로직이 정상적으로 작동해야 함

### API 도메인 특화 검수 (API Domain Specific)

- **API-001**: REST API 기능 검수
  - CRUD 연산이 정상 작동해야 함
  - HTTP 상태 코드가 올바르게 반환되어야 함

### 성능 검수 (Performance Testing)

**성능 요구사항:**
- [ ] 응답 시간: 1초 이내
- [ ] 동시 접속자: 100명 이상 지원

### 보안 검수 (Security Testing)

**보안 요구사항:**
- [ ] 인증 및 권한 검증 통과
- [ ] 입력값 검증 통과
- [ ] SQL 인젝션 방어 통과

### 검수 절차 (Validation Process)

#### 1단계: 단위 테스트 (Unit Tests)
- [ ] 개발자 테스트 완료
- [ ] 코드 리뷰 통과
- [ ] 자동화 테스트 통과

#### 2단계: 통합 테스트 (Integration Tests)
- [ ] 모듈 간 통합 테스트
- [ ] API 연동 테스트
- [ ] 데이터베이스 통합 테스트

#### 3단계: 시스템 테스트 (System Tests)
- [ ] 전체 시스템 기능 테스트
- [ ] 성능 테스트
- [ ] 보안 테스트

#### 4단계: 사용자 테스트 (User Tests)
- [ ] 내부 사용자 테스트
- [ ] 실제 사용자 테스트
- [ ] 피드백 수집 및 반영

### 검수 완료 기준 (Completion Criteria)

#### 통과 조건 (Pass Criteria)
- ✅ 모든 필수 기능 검수 통과
- ✅ 성능 요구사항 충족
- ✅ 보안 테스트 통과
- ✅ 사용자 검수 통과
- ✅ 문서 검수 완료

#### 보고서 작성 (Reporting)
- [ ] 검수 보고서 작성
- [ ] 발견된 이슈 목록 정리
- [ ] 개선 사항 정의
- [ ] 검수 승인서 작성

**검수 담당자:**
- 개발자: @developer
- QA: @qa_engineer
- 제품 책임자: @product_owner
- 최종 검수자: @stakeholder'''
        }

        # Sample low-quality SPEC content
        self.low_quality_spec = {
            'spec_md': '''# Low Quality SPEC
This is a poor quality specification.

Missing required sections like Overview, Environment, etc.

No technical details or proper structure.

Traceability tags are missing.

The content is too brief and incomplete.''',
            'plan_md': '''# Plan
No real plan here.
Just some random text.''',
            'acceptance_md': '''# Acceptance
No acceptance criteria.
Just placeholder text.'''
        }

        # Sample code analysis for consistency check
        self.code_analysis = {
            'structure_info': {
                'classes': ['User', 'Product', 'Order'],
                'functions': ['create_user', 'get_product', 'place_order'],
                'imports': ['flask', 'sqlalchemy', 'jwt']
            },
            'domain_keywords': ['api', 'auth', 'database', 'user', 'product']
        }

    def tearDown(self):
        """Clean up test environment."""
        import shutil
        shutil.rmtree(self.test_dir, ignore_errors=True)

    def test_validator_initialization(self):
        """Test validator initialization."""
        self.assertIsInstance(self.validator, QualityValidator)
        self.assertIsNotNone(self.validator.confidence_scorer)
        self.assertEqual(self.validator.min_ears_compliance, 0.85)
        self.assertEqual(self.validator.min_confidence_score, 0.7)
        self.assertIsInstance(self.validator.quality_weights, dict)

    def test_validate_spec_quality_high_quality(self):
        """Test quality validation for high-quality SPEC."""
        result = self.validator.validate_spec_quality(self.high_quality_spec)

        # Check basic structure
        self.assertIn('overall_score', result)
        self.assertIn('quality_grade', result)
        self.assertIn('validation_time', result)
        self.assertIn('passed_checks', result)
        self.assertIn('failed_checks', result)
        self.assertIn('recommendations', result)
        self.assertIn('details', result)

        # High-quality SPEC should have good score
        self.assertGreater(result['overall_score'], 0.8)
        self.assertEqual(result['quality_grade'], 'A')
        self.assertGreater(len(result['passed_checks']), 0)
        self.assertLess(len(result['failed_checks']), 3)

        # Check meets minimum standards
        self.assertTrue(result['meets_minimum_standards'])

    def test_validate_spec_quality_low_quality(self):
        """Test quality validation for low-quality SPEC."""
        result = self.validator.validate_spec_quality(self.low_quality_spec)

        # Low-quality SPEC should have poor score
        self.assertLess(result['overall_score'], 0.5)
        self.assertIn('F', result['quality_grade'])
        self.assertGreater(len(result['failed_checks']), 0)
        self.assertGreater(len(result['recommendations']), 0)

        # Check doesn't meet minimum standards
        self.assertFalse(result['meets_minimum_standards'])

    def test_validate_spec_quality_with_code_analysis(self):
        """Test quality validation with code analysis."""
        result = self.validator.validate_spec_quality(
            self.high_quality_spec,
            self.code_analysis
        )

        # Should have consistency check in details
        self.assertIn('technical_accuracy', result['details'])
        self.assertIn('code_consistency', result['details']['technical_accuracy'])

    def test_ears_format_validation(self):
        """Test EARS format validation."""
        result = self.validator._validate_ears_format(self.high_quality_spec['spec_md'])

        # Check structure
        self.assertIn('overall_compliance', result)
        self.assertIn('section_scores', result)
        self.assertIn('missing_sections', result)
        self.assertIn('has_meta_info', result)
        self.assertIn('has_tags', result)
        self.assertIn('has_proper_structure', result)

        # High-quality SPEC should have good compliance
        self.assertGreater(result['overall_compliance'], 0.9)
        self.assertEqual(len(result['missing_sections']), 0)
        self.assertTrue(result['has_meta_info'])
        self.assertTrue(result['has_tags'])
        self.assertTrue(result['has_proper_structure'])

    def test_ears_format_validation_low_quality(self):
        """Test EARS format validation for low-quality SPEC."""
        result = self.validator._validate_ears_format(self.low_quality_spec['spec_md'])

        # Low-quality SPEC should have poor compliance
        self.assertLess(result['overall_compliance'], 0.3)
        self.assertGreater(len(result['missing_sections']), 0)
        self.assertFalse(result['has_meta_info'])
        self.assertFalse(result['has_tags'])
        self.assertFalse(result['has_proper_structure'])

    def test_content_completeness_validation(self):
        """Test content completeness validation."""
        result = self.validator._validate_content_completeness(self.high_quality_spec)

        # Check structure
        self.assertIn('spec_completeness', result)
        self.assertIn('plan_completeness', result)
        self.assertIn('acceptance_completeness', result)
        self.assertIn('overall_completeness', result)

        # High-quality SPEC should have good completeness
        self.assertGreater(result['overall_completeness'], 0.8)
        self.assertGreater(result['spec_completeness'], 0.8)
        self.assertGreater(result['plan_completeness'], 0.8)
        self.assertGreater(result['acceptance_completeness'], 0.8)

    def test_content_completeness_validation_low_quality(self):
        """Test content completeness validation for low-quality SPEC."""
        result = self.validator._validate_content_completeness(self.low_quality_spec)

        # Low-quality SPEC should have poor completeness
        self.assertLess(result['overall_completeness'], 0.3)
        self.assertLess(result['spec_completeness'], 0.3)
        self.assertLess(result['plan_completeness'], 0.3)
        self.assertLess(result['acceptance_completeness'], 0.3)

    def test_technical_accuracy_validation(self):
        """Test technical accuracy validation."""
        result = self.validator._validate_technical_accuracy(
            self.high_quality_spec,
            self.code_analysis
        )

        # Check structure
        self.assertIn('keyword_presence', result)
        self.assertIn('code_consistency', result)
        self.assertIn('has_technical_specs', result)
        self.assertIn('has_realistic_estimates', result)
        self.assertIn('overall_accuracy', result)

        # High-quality SPEC should have good technical accuracy
        self.assertGreater(result['overall_accuracy'], 0.8)
        self.assertGreater(result['keyword_presence'], 0.5)
        self.assertGreater(result['code_consistency'], 0.5)
        self.assertTrue(result['has_technical_specs'])

    def test_clarity_validation(self):
        """Test clarity validation."""
        result = self.validator._validate_clarity(self.high_quality_spec)

        # Check structure
        self.assertIn('language_quality', result)
        self.assertIn('clarity_requirements', result)
        self.assertIn('ambiguity_score', result)
        self.assertIn('consistency_score', result)
        self.assertIn('overall_clarity', result)

        # High-quality SPEC should have good clarity
        self.assertGreater(result['overall_clarity'], 0.7)
        self.assertGreater(result['language_quality'], 0.5)
        self.assertGreater(result['clarity_requirements'], 0.5)
        self.assertLess(result['ambiguity_score'], 0.5)
        self.assertGreater(result['consistency_score'], 0.5)

    def test_traceability_validation(self):
        """Test traceability validation."""
        result = self.validator._validate_traceability(self.high_quality_spec['spec_md'])

        # Check structure
        self.assertIn('has_traceability_tags', result)
        self.assertIn('tag_formatting', result)
        self.assertIn('traceability_relationships', result)
        self.assertIn('overall_traceability', result)

        # High-quality SPEC should have good traceability
        self.assertTrue(result['has_traceability_tags'])
        self.assertGreater(result['tag_formatting'], 0.8)
        self.assertGreater(result['traceability_relationships'], 0.5)
        self.assertGreater(result['overall_traceability'], 0.8)

    def test_quality_grade_determination(self):
        """Test quality grade determination."""
        # Test A grade
        self.assertEqual(self.validator._determine_quality_grade(0.95), 'A')
        self.assertEqual(self.validator._determine_quality_grade(0.90), 'A')

        # Test B grade
        self.assertEqual(self.validator._determine_quality_grade(0.85), 'B')
        self.assertEqual(self.validator._determine_quality_grade(0.80), 'B')

        # Test C grade
        self.assertEqual(self.validator._determine_quality_grade(0.75), 'C')
        self.assertEqual(self.validator._determine_quality_grade(0.70), 'C')

        # Test D grade
        self.assertEqual(self.validator._determine_quality_grade(0.65), 'D')
        self.assertEqual(self.validator._determine_quality_grade(0.60), 'D')

        # Test F grade
        self.assertEqual(self.validator._determine_quality_grade(0.55), 'F')
        self.assertEqual(self.validator._determine_quality_grade(0.40), 'F')

    def test_quality_standards_check(self):
        """Test quality standards check."""
        # High-quality SPEC should pass
        high_result = self.validator.validate_spec_quality(self.high_quality_spec)
        self.assertTrue(high_result['meets_minimum_standards'])

        # Low-quality SPEC should fail
        low_result = self.validator.validate_spec_quality(self.low_quality_spec)
        self.assertFalse(low_result['meets_minimum_standards'])

    def test_recommendations_generation(self):
        """Test recommendations generation."""
        result = self.validator.validate_spec_quality(self.low_quality_spec)

        # Should have recommendations
        self.assertIsInstance(result['recommendations'], list)
        self.assertGreater(len(result['recommendations']), 0)

        # Check that recommendations are specific
        for rec in result['recommendations']:
            self.assertIsInstance(rec, str)
            self.assertGreater(len(rec), 0)

    def test_passed_failed_checks_compilation(self):
        """Test passed and failed checks compilation."""
        result = self.validator.validate_spec_quality(self.high_quality_spec)

        # Check passed checks
        self.assertIsInstance(result['passed_checks'], list)
        self.assertGreater(len(result['passed_checks']), 0)

        # Check failed checks
        self.assertIsInstance(result['failed_checks'], list)
        self.assertLess(len(result['failed_checks']), 3)

        # Check that checks are strings
        for check in result['passed_checks']:
            self.assertIsInstance(check, str)
            self.assertGreater(len(check), 0)

        for check in result['failed_checks']:
            self.assertIsInstance(check, str)
            self.assertGreater(len(check), 0)

    def test_validation_time_tracking(self):
        """Test validation time tracking."""
        result = self.validator.validate_spec_quality(self.high_quality_spec)

        # Should have reasonable validation time
        self.assertGreater(result['validation_time'], 0.0)
        self.assertLess(result['validation_time'], 5.0)  # Should be fast

    def test_custom_config(self):
        """Test custom configuration."""
        custom_config = {
            'min_ears_compliance': 0.90,
            'min_confidence_score': 0.80,
            'min_content_length': 1000,
            'max_review_suggestions': 5,
            'quality_weights': {
                'ears_compliance': 0.4,
                'content_completeness': 0.3,
                'technical_accuracy': 0.2,
                'clarity_score': 0.1,
                'traceability': 0.0
            }
        }

        validator = QualityValidator(custom_config)
        self.assertEqual(validator.min_ears_compliance, 0.90)
        self.assertEqual(validator.min_confidence_score, 0.80)
        self.assertEqual(validator.min_content_length, 1000)
        self.assertEqual(validator.max_review_suggestions, 5)
        self.assertEqual(validator.quality_weights['ears_compliance'], 0.4)
        self.assertEqual(validator.quality_weights['traceability'], 0.0)

    def test_quality_report_generation(self):
        """Test quality report generation."""
        result = self.validator.validate_spec_quality(self.high_quality_spec)
        report = self.validator.generate_quality_report(result)

        # Check report structure
        self.assertIsInstance(report, str)
        self.assertIn('Quality Validation Report', report)
        self.assertIn('Overall Quality Score', report)
        self.assertIn('Quality Grade', report)
        self.assertIn('Validation Time', report)
        self.assertIn('Summary', report)
        self.assertIn('Passed Checks', report)
        self.assertIn('Failed Checks', report)
        self.assertIn('Recommendations', report)
        self.assertIn('Quality Status', report)

        # Check that score is in report
        self.assertIn(str(result['overall_score']), report)
        self.assertIn(result['quality_grade'], report)

    def test_error_handling(self):
        """Test error handling."""
        # Test with invalid input
        result = self.validator.validate_spec_quality({})
        self.assertIn('overall_score', result)
        self.assertIn('quality_grade', result)

        # Test with None input
        result = self.validator.validate_spec_quality(None)
        self.assertIn('overall_score', result)
        self.assertIn('quality_grade', result)

    def test_heading_structure_check(self):
        """Test heading structure check."""
        # Good heading structure
        good_content = '''# Title
## Section 1
### Subsection 1.1
## Section 2
### Subsection 2.1'''
        self.assertTrue(self.validator._check_heading_structure(good_content))

        # Poor heading structure
        poor_content = '''# Title
Just some text
No proper headings'''
        self.assertFalse(self.validator._check_heading_structure(poor_content))

    def test_technical_keywords_check(self):
        """Test technical keywords check."""
        content_with_keywords = '''This system includes API endpoints, Database connections, Authentication mechanisms, Security features, and Performance optimization.'''
        keywords = ['API', 'Database', 'Authentication', 'Security', 'Performance']
        score = self.validator._check_technical_keywords(content_with_keywords, keywords)
        self.assertEqual(score, 1.0)  # All keywords found

        content_with_partial_keywords = '''This system has some basic API functionality.'''
        score = self.validator._check_technical_keywords(content_with_partial_keywords, keywords)
        self.assertLess(score, 1.0)  # Not all keywords found

    def test_code_consistency_check(self):
        """Test code consistency check."""
        # High consistency
        consistent_spec = '''The system includes User class with create_user method, Product class with get_product method, and Order class with place_order method.'''
        score = self.validator._check_code_consistency(consistent_spec, self.code_analysis)
        self.assertGreater(score, 0.8)

        # Low consistency
        inconsistent_spec = '''This is a generic system without specific classes or functions.'''
        score = self.validator._check_code_consistency(inconsistent_spec, self.code_analysis)
        self.assertLess(score, 0.5)

    def test_traceability_tags_check(self):
        """Test traceability tags check."""
        # Has tags
        with_tags = '''@SPEC:TEST-001 is related to @CODE:TEST-001 and @TEST:TEST-001'''
        self.assertTrue(self.validator._check_traceability_tags(with_tags))

        # No tags
        without_tags = '''This is a spec without any traceability tags.'''
        self.assertFalse(self.validator._check_traceability_tags(without_tags))

    def test_tag_formatting_check(self):
        """Test tag formatting check."""
        # Good formatting
        well_formatted = '''@SPEC:TEST-001, @TEST:TEST-001, @CODE:TEST-001, @ENV:TEST-001'''
        score = self.validator._check_tag_formatting(well_formatted)
        self.assertEqual(score, 1.0)

        # Poor formatting
        poorly_formatted = '''@SPEC:TEST-001, BAD-TAG:TEST-001, @TEST:TEST-001'''
        score = self.validator._check_tag_formatting(poorly_formatted)
        self.assertLess(score, 1.0)

        # No tags
        no_tags = '''No tags here.'''
        score = self.validator._check_tag_formatting(no_tags)
        self.assertEqual(score, 1.0)  # Should pass when no tags


if __name__ == '__main__':
    unittest.main()