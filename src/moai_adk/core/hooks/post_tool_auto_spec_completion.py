# @CODE:HOOK-POST-AUTO-SPEC-001
"""PostToolUse Hook for Automated SPEC Completion System."""

import os
import json
import re
import time
import hashlib
from typing import Dict, Any, List, Optional
from pathlib import Path
import logging

from moai_adk.core.tags.spec_generator import SpecGenerator
# BaseHook: Simplified base hook class for auto-spec completion
class BaseHook:
    """Base hook class for auto-spec completion."""
    def __init__(self):
        self.name = "PostToolAutoSpecCompletion"
        self.description = "PostToolUse Hook for Automated SPEC Completion System"


# Configure logging
logger = logging.getLogger(__name__)


class PostToolAutoSpecCompletion(BaseHook):
    """
    PostToolUse Hook for automated SPEC completion.

    This hook detects code file changes after Write/Edit/MultiEdit tools
    and automatically generates complete SPEC documents in EARS format.
    """

    def __init__(self):
        super().__init__()
        self.spec_generator = SpecGenerator()
        self.auto_config = self._get_auto_spec_config()

        # Track processed files to avoid duplicates
        self.processed_files = set()

    def _get_auto_spec_config(self) -> Dict[str, Any]:
        """Get auto-spec completion configuration."""
        try:
            from moai_adk.core.config.config_manager import ConfigManager
            config = ConfigManager()
            return config.get_value('tags.policy.auto_spec_completion', {
                'enabled': True,
                'min_confidence': 0.7,
                'auto_open_editor': True,
                'supported_languages': ['python', 'javascript', 'typescript', 'go'],
                'excluded_patterns': ['test_', 'spec_', '__tests__']
            })
        except ImportError:
            return {
                'enabled': True,
                'min_confidence': 0.7,
                'auto_open_editor': True,
                'supported_languages': ['python', 'javascript', 'typescript', 'go'],
                'excluded_patterns': ['test_', 'spec_', '__tests__']
            }

    def should_trigger_spec_completion(self, tool_name: str, tool_args: Dict[str, Any]) -> bool:
        """
        Determine if spec completion should be triggered.

        Args:
            tool_name: Name of the tool that was executed
            tool_args: Arguments passed to the tool

        Returns:
            True if spec completion should be triggered
        """
        # Check if auto-spec completion is enabled
        if not self.auto_config.get('enabled', True):
            logger.debug("Auto-spec completion is disabled")
            return False

        # Only trigger for Write/Edit/MultiEdit tools
        if tool_name not in ['Write', 'Edit', 'MultiEdit']:
            logger.debug(f"Tool {tool_name} does not trigger spec completion")
            return False

        # Extract file paths from tool arguments
        file_paths = self._extract_file_paths(tool_args)

        if not file_paths:
            logger.debug("No file paths found in tool arguments")
            return False

        # Check if any file is a supported language
        supported_files = []
        for file_path in file_paths:
            if self._is_supported_file(file_path):
                supported_files.append(file_path)
            else:
                logger.debug(f"File {file_path} is not supported for auto-spec completion")

        if not supported_files:
            logger.debug("No supported files found")
            return False

        # Check for excluded patterns
        excluded_files = []
        for file_path in supported_files:
            if self._is_excluded_file(file_path):
                excluded_files.append(file_path)

        # Filter out excluded files
        target_files = [f for f in supported_files if f not in excluded_files]

        if not target_files:
            logger.debug("All files are excluded from auto-spec completion")
            return False

        return True

    def _extract_file_paths(self, tool_args: Dict[str, Any]) -> List[str]:
        """Extract file paths from tool arguments."""
        file_paths = []

        # Handle Write tool
        if 'file_path' in tool_args:
            file_paths.append(tool_args['file_path'])

        # Handle Edit tool
        if 'file_path' in tool_args:
            file_paths.append(tool_args['file_path'])

        # Handle MultiEdit tool
        if 'edits' in tool_args:
            for edit in tool_args['edits']:
                if 'file_path' in edit:
                    file_paths.append(edit['file_path'])

        # Remove duplicates and resolve relative paths
        unique_paths = []
        for path in file_paths:
            if path not in unique_paths:
                abs_path = os.path.abspath(path)
                unique_paths.append(abs_path)

        return unique_paths

    def _is_supported_file(self, file_path: str) -> bool:
        """Check if file is supported for auto-spec completion."""
        # Get file extension
        file_ext = os.path.splitext(file_path)[1].lower()

        # Map extensions to languages
        supported_extensions = {
            '.py': 'python',
            '.js': 'javascript',
            '.jsx': 'javascript',
            '.ts': 'typescript',
            '.tsx': 'typescript',
            '.go': 'go'
        }

        if file_ext not in supported_extensions:
            return False

        # Check if language is supported
        language = supported_extensions[file_ext]
        supported_languages = self.auto_config.get('supported_languages', [])
        return language in supported_languages

    def _is_excluded_file(self, file_path: str) -> bool:
        """Check if file should be excluded from auto-spec completion."""
        file_name = os.path.basename(file_path)
        file_dir = os.path.basename(os.path.dirname(file_path))

        excluded_patterns = self.auto_config.get('excluded_patterns', [])

        for pattern in excluded_patterns:
            # Check filename patterns
            if re.search(pattern, file_name):
                return True
            # Check directory patterns
            if re.search(pattern, file_dir):
                return True

        return False

    def detect_code_changes(self, tool_name: str, tool_args: Dict[str, Any],
                           result: Any) -> List[str]:
        """
        Detect code changes from tool execution.

        Args:
            tool_name: Name of the tool that was executed
            tool_args: Arguments passed to the tool
            result: Result from tool execution

        Returns:
            List of affected file paths
        """
        file_paths = []

        # Write tool creates new files
        if tool_name == 'Write':
            if 'file_path' in tool_args:
                file_paths.append(tool_args['file_path'])

        # Edit tool modifies existing files
        elif tool_name == 'Edit':
            if 'file_path' in tool_args:
                file_paths.append(tool_args['file_path'])

        # MultiEdit tool can modify multiple files
        elif tool_name == 'MultiEdit':
            if 'edits' in tool_args:
                for edit in tool_args['edits']:
                    if 'file_path' in edit:
                        file_paths.append(edit['file_path'])

        # Convert to absolute paths
        abs_paths = [os.path.abspath(path) for path in file_paths]

        # Filter out already processed files
        new_paths = [path for path in abs_paths if path not in self.processed_files]

        # Add to processed files
        self.processed_files.update(new_paths)

        return new_paths

    def calculate_completion_confidence(self, analysis: Dict[str, Any]) -> float:
        """
        Calculate confidence score for SPEC completion.

        Args:
            analysis: Code analysis result

        Returns:
            Confidence score between 0.0 and 1.0
        """
        # Default confidence if analysis is incomplete
        if not analysis:
            return 0.5

        structure_score = analysis.get('structure_score', 0.5)
        domain_accuracy = analysis.get('domain_accuracy', 0.5)
        documentation_level = analysis.get('documentation_level', 0.5)

        # Weighted calculation
        # Structure clarity: 30%
        # Domain accuracy: 40%
        # Documentation level: 30%
        confidence = (
            structure_score * 0.3 +
            domain_accuracy * 0.4 +
            documentation_level * 0.3
        )

        return min(max(confidence, 0.0), 1.0)

    def generate_complete_spec(self, analysis: Dict[str, Any], file_path: str) -> Dict[str, str]:
        """
        Generate complete SPEC documents in EARS format.

        Args:
            analysis: Code analysis result
            file_path: Path to the analyzed file

        Returns:
            Dictionary containing spec.md, plan.md, and acceptance.md
        """
        spec_id = self._generate_spec_id(file_path)
        file_name = os.path.basename(file_path)

        # Generate basic spec content
        spec_md = self._generate_spec_content(analysis, spec_id, file_name)
        plan_md = self._generate_plan_content(analysis, spec_id, file_name)
        acceptance_md = self._generate_acceptance_content(analysis, spec_id, file_name)

        return {
            'spec_id': spec_id,
            'spec_md': spec_md,
            'plan_md': plan_md,
            'acceptance_md': acceptance_md
        }

    def _generate_spec_id(self, file_path: str) -> str:
        """Generate unique SPEC ID from file path."""
        # Extract meaningful name from file path
        file_name = os.path.basename(file_path)
        name_parts = file_name.split('_')

        # Convert to uppercase and join
        meaningful_name = ''.join(part.upper() for part in name_parts if part)

        # Add hash to ensure uniqueness
        file_hash = hashlib.md5(file_path.encode()).hexdigest()[:4]

        return f"{meaningful_name}-{file_hash}"

    def _generate_spec_content(self, analysis: Dict[str, Any], spec_id: str, file_name: str) -> str:
        """Generate main spec.md content."""
        template = f"""---
@META: {{
  "id": "SPEC-{spec_id}",
  "title": "Auto-generated SPEC for {file_name}",
  "title_en": "Auto-generated SPEC for {file_name}",
  "version": "1.0.0",
  "status": "pending",
  "created": "{time.strftime('%Y-%m-%d')}",
  "author": "@alfred-auto",
  "reviewer": "",
  "category": "FEATURE",
  "priority": "MEDIUM",
  "tags": ["auto-generated", "{spec_id}"],
  "language": "ko",
  "estimated_complexity": "auto"
}}
---

# @SPEC:{spec_id}: Auto-generated SPEC for {file_name}
## Auto-generated SPEC for {file_name}

### 개요 (Overview)

{analysis.get('description', 'This spec was auto-generated based on code analysis.')}

### 환경 (Environment)

- **프로젝트**: MoAI-ADK Auto-generated SPEC
- **언어**: {analysis.get('language', 'Python')}
- **파일**: {file_name}
- **생성 방식**: 자동 분석 기반
- **상태**: 검토 필요

### 가정 (Assumptions)

1. 코드 구조는 명확하게 정의되어 있음
2. 도메인 전문 용어가 사용되었을 것으로 추정
3. 표준 개발 방식을 따랐을 것으로 가정
4. 생성된 SPEC은 사용자 검토 후 최종화될 것으로 예상

### 요구사항 (Requirements)

#### 보편적 요구사항 (Ubiquitous Requirements)

- **REQ-001**: 시스템은 {file_name}의 기능을 수행해야 함
- **REQ-002**: 생성된 기능은 안정적이어야 함
- **REQ-003**: 코드는 유지보수 가능한 형태로 작성되어야 함
- **REQ-004**: 테스트는 기능적 요구사항을 충족해야 함
- **REQ-005**: 코드는 프로젝트 코딩 표준을 준수해야 함

#### 상태 기반 요구사항 (State-driven Requirements)

{analysis.get('state_requirements', '- **REQ-006**: 시스템은 초기 상태에서 목표 상태로 전환해야 함')}

#### 이벤트 기반 요구사항 (Event-driven Requirements)

{analysis.get('event_requirements', '- **REQ-007**: 사용자 입력 시 시스템이 반응해야 함')}

### 명세 (Specifications)

{analysis.get('specifications', '- **SPEC-001**: 시스템은 요구사항을 구현해야 함')}

### 추적성 (Traceability)

- **@SPEC:{spec_id}** ← **@CODE:HOOK-POST-AUTO-SPEC-001** (후크 자동 생성)
- **@SPEC:{spec_id}** → **@TEST:{spec_id}** (테스트)
- **@SPEC:{spec_id}** → **@CODE:{spec_id}** (구현)

### 편집 가이드 (Edit Guide)

**사용자 검토 권항목:**
1. ✅ 기술적 명확성 확인
2. ✅ 요구사항 구체화
3. ✅ 도메인 전문 용어 검토
4. ✅ 상태 및 이벤트 요구사항 정의
5. ✅ 명세사항 상세화

**품질 개선 제안:**
- 도메인 전문 용어 추가
- 사용자 사례 구체화
- 성능 요구사항 정의
- 보안 요구사항 추가
"""
        return template

    def _generate_plan_content(self, analysis: Dict[str, Any], spec_id: str, file_name: str) -> str:
        """Generate plan.md content."""
        return f"""---
@META: {{
  "id": "PLAN-{spec_id}",
  "spec_id": "SPEC-{spec_id}",
  "title": "Auto-generated Implementation Plan for {file_name}",
  "version": "1.0.0",
  "status": "pending",
  "created": "{time.strftime('%Y-%m-%d')}",
  "author": "@alfred-auto"
}}
---

# @PLAN:{spec_id}: Auto-generated Implementation Plan
## Auto-generated Implementation Plan for {file_name}

### 구현 단계 (Implementation Phases)

#### 1단계: 기본 구조 검토 (Priority: High)

- [ ] 코드 구조 분석 완료
- [ ] 핵심 기능 식별
- [ ] 의존성 확인
- [ ] 테스트 환경 설정

#### 2단계: 요구사항 상세화 (Priority: Medium)

- [ ] 보편적 요구사항 구체화
- [ ] 상태 기반 요구사항 정의
- [ ] 이벤트 기반 요구사항 검토
- [ ] 성능 요구사항 설정

#### 3단계: 구현 계획 수립 (Priority: Medium)

- [ ] 모듈 아키텍처 설계
- [ ] 인터페이스 정의
- [ ] 데이터 구조 설계
- [ ] 오류 처리 계획

#### 4단계: 테스트 전략 수립 (Priority: High)

- [ ] 단위 테스트 계획
- [ ] 통합 테스트 계획
- [ ] 사용자 스토리 기반 테스트
- [ ] 자동화 테스트 구현

### 기술적 접근 방식 (Technical Approach)

#### 아키텍처 설계

```
{analysis.get('architecture', 'User Input → Validation → Business Logic → Data Processing → Output')}
    ↓
[Core Components] → [External Services] → [Data Layer]
```

#### 핵심 컴포넌트

1. **{analysis.get('main_component', 'Main Class')}**: 주요 비즈니스 로직 처리
2. **{analysis.get('service_component', 'Service Layer')}**: 외부 서비스 연동
3. **{analysis.get('data_component', 'Data Layer')}**: 데이터 처리 및 저장
4. **{analysis.get('component_4', 'Validation Layer')}**: 입력값 검증 및 유효성 검사

#### 의존성 관리

**기존 모듈 활용:**
- {analysis.get('existing_modules', '표준 라이브러리 활용')}

**신규 모듈 추가:**
- {analysis.get('new_modules', '필요에 따라 추가')}

### 성공 기준 (Success Criteria)

#### 기능적 기준

- ✅ 모든 요구사항 구현 완료
- ✅ 테스트 커버리지 85% 이상
- ✅ 성능 목표 충족
- ✅ 사용자 요구사항 충족

#### 성능 기준

- ✅ 응답 시간 {analysis.get('performance_target', '1초 이내')}
- ✅ 메모리 사용량 최적화
- ✅ 병렬 처리 지원
- ✅ 확장성 확인

#### 품질 기준

- ✅ 코드 품질 검증 통과
- ✅ 보안 스캐닝 통과
- ✅ 문서 완성도 확인
- ✅ 유지보수 용이성 검증

### 다음 단계 (Next Steps)

1. **즉시 실행**: 기본 구조 검토 (1-2일)
2. **주간 목표**: 요구사항 상세화 (3-5일)
3. **2주 목표**: 구현 완료 (7-14일)
4. **배포 준비**: 테스트 및 검증 (14-16일)
"""

    def _generate_acceptance_content(self, analysis: Dict[str, Any], spec_id: str, file_name: str) -> str:
        """Generate acceptance.md content."""
        return f"""---
@META: {{
  "id": "ACCEPT-{spec_id}",
  "spec_id": "SPEC-{spec_id}",
  "title": "Auto-generated Acceptance Criteria for {file_name}",
  "version": "1.0.0",
  "status": "pending",
  "created": "{time.strftime('%Y-%m-%d')}",
  "author": "@alfred-auto"
}}
---

# @ACCEPT:{spec_id}: Auto-generated Acceptance Criteria
## Auto-generated Acceptance Criteria for {file_name}

### 검수 기준 (Acceptance Criteria)

#### 기본 기능 검수 (Basic Functionality)

**필수 조건 (Must-have):**
- [ ] {analysis.get('must_have_1', '시스템이 정상적으로 구동되어야 함')}
- [ ] {analysis.get('must_have_2', '사용자 인터페이스가 올바르게 표시되어야 함')}
- [ ] {analysis.get('must_have_3', '데이터 처리 로직이 정상적으로 작동해야 함')}

**필수 조건 (Should-have):**
- [ ] {analysis.get('should_have_1', '사용자 경험이 원활해야 함')}
- [ ] {analysis.get('should_have_2', '성능 목표를 충족해야 함')}

#### 성능 검수 (Performance Testing)

**성능 요구사항:**
- [ ] 응답 시간: {analysis.get('response_time', '1초 이내')}
- [ ] 동시 접속자: {analysis.get('concurrent_users', '100명')} 이상 지원
- [ ] 메모리 사용량: {analysis.get('memory_usage', '100MB 이하')}
- [ ] CPU 사용률: {analysis.get('cpu_usage', '50% 이하')}

**부하 테스트:**
- [ ] 기능 부하 테스트 통과
- [ ] 장기 안정성 테스트 통과
- [ ] 회복 테스트 통과

#### 보안 검수 (Security Testing)

**보안 요구사항:**
- [ ] {analysis.get('security_req_1', '인증 및 권한 검증 통과')}
- [ ] {analysis.get('security_req_2', '입력값 검증 통과')}
- [ ] {analysis.get('security_req_3', 'SQL 인젝션 방어 통과')}

**취약점 테스트:**
- [ ] OWASP Top 10 검사 통과
- [ ] 보안 스캐닝 통과
- [ ] 권한 설정 검증 통과

#### 호환성 검수 (Compatibility Testing)

**브라우저 호환성:**
- [ ] Chrome 최신 버전
- [ ] Firefox 최신 버전
- [ ] Safari 최신 버전
- [ ] Edge 최신 버전

**디바이스 호환성:**
- [ ] 데스크톱 (1920x1080)
- [ ] 태블릿 (768x1024)
- [ ] 모바일 (375x667)

#### 사용자 검수 (User Acceptance Testing)

**사용자 시나리오:**
- [ ] {analysis.get('user_scenario_1', '일반 사용자 시나리오 테스트 통과')}
- [ ] {analysis.get('user_scenario_2', '관리자 시나리오 테스트 통과')}
- [ ] {analysis.get('user_scenario_3', '에러 처리 시나리오 테스트 통과')}

**사용자 피드백:**
- [ ] 사용자 만족도 80% 이상
- [ ] 기능 사용 편의성 평가
- [ ] 디자인 및 UI/UX 검증

### 검수 절차 (Validation Process)

#### 1단계: 단위 테스트 (Unit Tests)

- [ ] 개발자 테스트 완료
- [ ] 코드 리뷰 통과
- [ ] 자동화 테스트 통과
- [ ] 코드 커버리지 85% 이상

#### 2단계: 통합 테스트 (Integration Tests)

- [ ] 모듈 간 통합 테스트
- [ ] API 연동 테스트
- [ ] 데이터베이스 통합 테스트
- [ ] 외부 서비스 연동 테스트

#### 3단계: 시스템 테스트 (System Tests)

- [ ] 전체 시스템 기능 테스트
- [ ] 성능 테스트
- [ ] 보안 테스트
- [ ] 안정성 테스트

#### 4단계: 사용자 테스트 (User Tests)

- [ ] 내부 사용자 테스트
- [ ] 실제 사용자 테스트
- [ ] 피드백 수집 및 반영
- [ ] 최종 검수 승인

### 검수 템플릿 (Validation Templates)

#### 기능 검수 템플릿

| 기능 ID | 기능명 | 예상 결과 | 실제 결과 | 상태 | 비고 |
|---------|--------|-----------|-----------|------|------|
| FUNC-001 | 기능 1 | 성공 | 테스트 중 | 진행중 | 설명 |
| FUNC-002 | 기능 2 | 성공 | 성공 | 통과 | 설명 |
| FUNC-003 | 기능 3 | 성공 | 실패 | 실패 | 설명 |

#### 성능 검수 템플릿

| 테스트 항목 | 목표치 | 측정치 | 상태 | 비고 |
|-------------|--------|--------|------|------|
| 응답 시간 | 1초 | 0.8초 | 통과 | 설명 |
| 메모리 사용 | 100MB | 85MB | 통과 | 설명 |
| CPU 사용률 | 50% | 45% | 통과 | 설명 |

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
- 최종 검수자: @stakeholder
"""

    def validate_generated_spec(self, spec_content: Dict[str, str]) -> Dict[str, Any]:
        """
        Validate quality of generated spec.

        Args:
            spec_content: Dictionary with spec.md, plan.md, acceptance_md

        Returns:
            Validation result with quality metrics
        """
        quality_score = 0.0
        suggestions = []

        # Check EARS format compliance
        ears_compliance = self._check_ears_compliance(spec_content)
        quality_score += ears_compliance * 0.4

        # Check completeness
        completeness = self._check_completeness(spec_content)
        quality_score += completeness * 0.3

        # Check content quality
        content_quality = self._check_content_quality(spec_content)
        quality_score += content_quality * 0.3

        # Generate suggestions
        if ears_compliance < 0.9:
            suggestions.append("EARS 형식을 완벽하게 준수하도록 개선이 필요합니다.")

        if completeness < 0.8:
            suggestions.append("요구사항과 명세사항을 더 구체화해야 합니다.")

        if content_quality < 0.7:
            suggestions.append("도메인 전문 용어와 기술적 내용을 추가해야 합니다.")

        return {
            'quality_score': min(max(quality_score, 0.0), 1.0),
            'ears_compliance': ears_compliance,
            'completeness': completeness,
            'content_quality': content_quality,
            'suggestions': suggestions
        }

    def _check_ears_compliance(self, spec_content: Dict[str, str]) -> float:
        """Check EARS format compliance."""
        spec_md = spec_content.get('spec_md', '')

        required_sections = ['개요 (Overview)', '환경 (Environment)', '가정 (Assumptions)',
                           '요구사항 (Requirements)', '명세 (Specifications)']

        found_sections = 0
        for section in required_sections:
            if section in spec_md:
                found_sections += 1

        return found_sections / len(required_sections)

    def _check_completeness(self, spec_content: Dict[str, str]) -> float:
        """Check content completeness."""
        spec_md = spec_content.get('spec_md', '')
        plan_md = spec_content.get('plan_md', '')
        acceptance_md = spec_content.get('acceptance_md', '')

        # Check minimum content length
        total_length = len(spec_md) + len(plan_md) + len(acceptance_md)
        length_score = min(total_length / 2000, 1.0)  # 2000 chars as baseline

        # Check for content diversity
        has_requirements = '요구사항' in spec_md
        has_planning = '구현 계획' in plan_md
        has_acceptance = '검수' in acceptance_md

        diversity_score = 0.0
        if has_requirements:
            diversity_score += 0.3
        if has_planning:
            diversity_score += 0.3
        if has_acceptance:
            diversity_score += 0.4

        return (length_score + diversity_score) / 2

    def _check_content_quality(self, spec_content: Dict[str, str]) -> float:
        """Check content quality."""
        spec_md = spec_content.get('spec_md', '')

        # Check for technical terms
        technical_indicators = ['API', '데이터', '인터페이스', '모듈', '컴포넌트', '아키텍처']
        technical_score = sum(1 for term in technical_indicators if term in spec_md) / len(technical_indicators)

        # Check for specificity
        has_requirements = re.search(r'REQ-\d+', spec_md)
        has_specifications = re.search(r'SPEC-\d+', spec_md)

        specificity_score = 0.0
        if has_requirements:
            specificity_score += 0.5
        if has_specifications:
            specificity_score += 0.5

        return (technical_score + specificity_score) / 2

    def create_spec_files(self, spec_id: str, content: Dict[str, str],
                         base_dir: str = ".moai/specs") -> bool:
        """
        Create SPEC files in the correct directory structure.

        Args:
            spec_id: SPEC identifier
            content: Dictionary with spec_md, plan_md, acceptance_md
            base_dir: Base directory for specs

        Returns:
            True if files were created successfully
        """
        try:
            # Create spec directory
            spec_dir = os.path.join(base_dir, f"SPEC-{spec_id}")
            os.makedirs(spec_dir, exist_ok=True)

            # Create files
            files_to_create = [
                ('spec.md', content.get('spec_md', '')),
                ('plan.md', content.get('plan_md', '')),
                ('acceptance.md', content.get('acceptance_md', ''))
            ]

            for filename, content_text in files_to_create:
                file_path = os.path.join(spec_dir, filename)
                with open(file_path, 'w', encoding='utf-8') as f:
                    f.write(content_text)

                logger.info(f"Created spec file: {file_path}")

            return True

        except Exception as e:
            logger.error(f"Failed to create spec files: {e}")
            return False

    def execute(self, tool_name: str, tool_args: Dict[str, Any],
               result: Any = None) -> Dict[str, Any]:
        """
        Execute the auto-spec completion hook.

        Args:
            tool_name: Name of the tool that was executed
            tool_args: Arguments passed to the tool
            result: Result from tool execution

        Returns:
            Execution result
        """
        start_time = time.time()

        try:
            # Check if we should trigger spec completion
            if not self.should_trigger_spec_completion(tool_name, tool_args):
                return {
                    'success': False,
                    'message': 'Auto-spec completion not triggered',
                    'execution_time': time.time() - start_time
                }

            # Detect code changes
            changed_files = self.detect_code_changes(tool_name, tool_args, result)

            if not changed_files:
                return {
                    'success': False,
                    'message': 'No code changes detected',
                    'execution_time': time.time() - start_time
                }

            # Process each changed file
            results = []
            for file_path in changed_files:
                try:
                    # Analyze the code file
                    analysis = self.spec_generator.analyze(file_path)

                    # Calculate confidence
                    confidence = self.calculate_completion_confidence(analysis)

                    # Skip if confidence is too low
                    min_confidence = self.auto_config.get('min_confidence', 0.7)
                    if confidence < min_confidence:
                        logger.info(f"Confidence {confidence} below threshold {min_confidence}")
                        continue

                    # Generate complete spec
                    spec_content = self.generate_complete_spec(analysis, file_path)

                    # Validate quality
                    validation = self.validate_generated_spec(spec_content)

                    # Create spec files
                    spec_id = spec_content['spec_id']
                    created = self.create_spec_files(spec_id, spec_content)

                    results.append({
                        'file_path': file_path,
                        'spec_id': spec_id,
                        'confidence': confidence,
                        'quality_score': validation['quality_score'],
                        'created': created
                    })

                    logger.info(f"Auto-generated SPEC for {file_path}: {spec_id}")

                except Exception as e:
                    logger.error(f"Error processing {file_path}: {e}")
                    results.append({
                        'file_path': file_path,
                        'error': str(e)
                    })

            # Generate summary
            successful_creations = [r for r in results if r.get('created', False)]
            failed_creations = [r for r in results if not r.get('created', False)]

            execution_result = {
                'success': len(successful_creations) > 0,
                'generated_specs': successful_creations,
                'failed_files': failed_creations,
                'execution_time': time.time() - start_time
            }

            # Add notification message
            if successful_creations:
                execution_result['message'] = f"Auto-generated {len(successful_creations)} SPEC(s)"
            elif failed_creations:
                execution_result['message'] = f"Auto-spec completion attempted but no specs created"
            else:
                execution_result['message'] = "No files required auto-spec completion"

            return execution_result

        except Exception as e:
            logger.error(f"Error in auto-spec completion: {e}")
            return {
                'success': False,
                'error': str(e),
                'execution_time': time.time() - start_time
            }