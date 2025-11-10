# @CODE:SPEC-EARS-TEMPLATE-001
"""EARS Template Engine for Auto-Generated SPECs."""

import re
import json
import time
import random
from typing import Dict, Any, List, Optional, Tuple
from pathlib import Path
import logging

from moai_adk.core.spec.confidence_scoring import ConfidenceScoringSystem

# Configure logging
logger = logging.getLogger(__name__)


class EARSTemplateEngine:
    """
    EARS Template Engine for generating complete SPEC documents.

    This engine generates SPEC documents in EARS (Environment, Assumptions,
    Requirements, Specifications) format based on code analysis.
    """

    def __init__(self):
        self.confidence_scorer = ConfidenceScoringSystem()
        self.template_cache = {}

        # Domain-specific templates
        self.domain_templates = {
            'auth': {
                'description': 'User authentication and security system',
                'common_features': ['Login', 'Registration', 'Password Reset', 'Session Management'],
                'security_requirements': ['Encryption', 'Password Hashing', 'Rate Limiting'],
                'environment': 'Web Application with User Management'
            },
            'api': {
                'description': 'RESTful API service',
                'common_features': ['Endpoints', 'Authentication', 'Rate Limiting', 'Caching'],
                'technical_requirements': ['RESTful Design', 'JSON Format', 'HTTP Status Codes'],
                'environment': 'Microservice Architecture'
            },
            'data': {
                'description': 'Data processing and storage system',
                'common_features': ['Data Validation', 'Persistence', 'Backup', 'Migration'],
                'technical_requirements': ['Data Integrity', 'Performance', 'Scalability'],
                'environment': 'Database System with Analytics'
            },
            'ui': {
                'description': 'User interface and experience system',
                'common_features': ['Components', 'Navigation', 'Forms', 'Validation'],
                'experience_requirements': ['Responsive Design', 'Accessibility', 'Performance'],
                'environment': 'Web Frontend with React/Angular/Vue'
            },
            'business': {
                'description': 'Business logic and workflow system',
                'common_features': ['Process Management', 'Rules Engine', 'Notifications'],
                'business_requirements': ['Compliance', 'Audit Trail', 'Reporting'],
                'environment': 'Enterprise Application'
            }
        }

        # EARS section templates
        self.ears_templates = {
            'environment': {
                'template': '''### 환경 (Environment)

- **프로젝트**: {project_name}
- **언어**: {language}
- **프레임워크**: {framework}
- **패러다임**: {paradigm}
- **플랫폼**: {platform}
- **배포**: {deployment}
- **상태**: {status}
- **생성 방식**: 자동 분석 기반''',
                'required_fields': ['project_name', 'language', 'framework', 'paradigm']
            },
            'assumptions': {
                'template': '''### 가정 (Assumptions)

1. 시스템은 표준 개발 방식을 따랐을 것으로 가정
2. 사용자는 해당 도메인에 대한 기본 지식을 가지고 있을 것으로 가정
3. 시스템은 안정적이고 확장 가능한 아키텍처로 설계되었을 것으로 가정
4. 외부 의존성은 정상적으로 작동할 것으로 가정
5. 보안 요구사항은 업계 표준을 준수할 것으로 가정
6. 데이터 무결성은 유지될 것으로 가정
7. 사용자 인터페이스는 직관적으로 설계되었을 것으로 가정
8. 성능 요구사항은 충족될 것으로 가정''',
                'required_fields': []
            },
            'requirements': {
                'template': '''### 요구사항 (Requirements)

#### 보편적 요구사항 (Ubiquitous Requirements)

- **REQ-001**: 시스템은 {primary_function}의 기능을 수행해야 함
- **REQ-002**: 생성된 기능은 안정적이어야 함
- **REQ-003**: 코드는 유지보수 가능한 형태로 작성되어야 함
- **REQ-004**: 테스트는 기능적 요구사항을 충족해야 함
- **REQ-005**: 코드는 프로젝트 코딩 표준을 준수해야 함
- **REQ-006**: 시스템은 예외 상황을 적절히 처리해야 함
- **REQ-007**: 사용자 경험은 최적화되어야 함

#### 상태 기반 요구사항 (State-driven Requirements)

{state_requirements}

#### 이벤트 기반 요구사항 (Event-driven Requirements)

{event_requirements}

#### 선택적 요구사항 (Optional Requirements)

- **REQ-008**: 시스템은 성능 모니터링 기능을 포함해야 함
- **REQ-009**: 자동 백업 및 복원 기능이 필요할 수 있음
- **REQ-010**: 사용자 활동 로깅이 필요할 수 있음
- **REQ-011**: 다국어 지원이 필요할 수 있음
- **REQ-012**: 모바일 호환성이 필요할 수 있음''',
                'required_fields': ['primary_function']
            },
            'specifications': {
                'template': '''### 명세 (Specifications)

{technical_specs}

#### 기술적 명세사항

{technical_details}

#### 데이터 모델

{data_models}

#### API 명세

{api_specs}

#### 인터페이스 명세

{interface_specs}

#### 보안 명세

{security_specs}

#### 성능 명세

{performance_specs}

#### 확장성 명세

{scalability_specs}''',
                'required_fields': []
            }
        }

    def generate_complete_spec(self, code_analysis: Dict[str, Any],
                            file_path: str,
                            custom_config: Dict[str, Any] = None) -> Dict[str, str]:
        """
        Generate complete SPEC document in EARS format.

        Args:
            code_analysis: Code analysis result
            file_path: Path to the analyzed file
            custom_config: Custom configuration overrides

        Returns:
            Dictionary with spec.md, plan.md, and acceptance.md
        """
        start_time = time.time()

        # Extract information from code analysis
        extraction_result = self._extract_information_from_analysis(code_analysis, file_path)

        # Determine domain
        domain = self._determine_domain(extraction_result)

        # Generate SPEC ID
        spec_id = self._generate_spec_id(extraction_result, domain)

        # Generate content for each section
        spec_md_content = self._generate_spec_content(extraction_result, domain, spec_id, custom_config)
        plan_md_content = self._generate_plan_content(extraction_result, domain, spec_id, custom_config)
        acceptance_md_content = self._generate_acceptance_content(extraction_result, domain, spec_id, custom_config)

        # Validate content
        validation_result = self._validate_ears_compliance({
            'spec_md': spec_md_content,
            'plan_md': plan_md_content,
            'acceptance_md': acceptance_md_content
        })

        # Create result
        result = {
            'spec_id': spec_id,
            'domain': domain,
            'spec_md': spec_md_content,
            'plan_md': plan_md_content,
            'acceptance_md': acceptance_md_content,
            'validation': validation_result,
            'generation_time': time.time() - start_time,
            'extraction': extraction_result
        }

        return result

    def _extract_information_from_analysis(self, code_analysis: Dict[str, Any], file_path: str) -> Dict[str, Any]:
        """Extract information from code analysis."""
        extraction = {
            'file_path': file_path,
            'file_name': Path(file_path).stem,
            'file_extension': Path(file_path).suffix,
            'language': self._detect_language(file_path),
            'classes': [],
            'functions': [],
            'imports': [],
            'domain_keywords': [],
            'technical_indicators': [],
            'complexity': 'low',
            'architecture': 'simple'
        }

        # Extract from code_analysis
        if 'structure_info' in code_analysis:
            structure = code_analysis['structure_info']
            extraction['classes'] = structure.get('classes', [])
            extraction['functions'] = structure.get('functions', [])
            extraction['imports'] = structure.get('imports', [])

        if 'domain_keywords' in code_analysis:
            extraction['domain_keywords'] = code_analysis['domain_keywords']

        # Extract from AST analysis if available
        if hasattr(code_analysis, 'ast_info'):
            ast_info = code_analysis.ast_info
            # Additional extraction logic here

        # Determine complexity and architecture
        extraction['complexity'] = self._analyze_complexity(extraction)
        extraction['architecture'] = self._analyze_architecture(extraction)

        return extraction

    def _detect_language(self, file_path: str) -> str:
        """Detect programming language from file path."""
        extension = Path(file_path).suffix.lower()

        language_map = {
            '.py': 'Python',
            '.js': 'JavaScript',
            '.jsx': 'JavaScript',
            '.ts': 'TypeScript',
            '.tsx': 'TypeScript',
            '.go': 'Go',
            '.java': 'Java',
            '.cpp': 'C++',
            '.c': 'C',
            '.cs': 'C#',
            '.rb': 'Ruby',
            '.php': 'PHP',
            '.swift': 'Swift',
            '.kt': 'Kotlin'
        }

        return language_map.get(extension, 'Unknown')

    def _analyze_complexity(self, extraction: Dict[str, Any]) -> str:
        """Analyze code complexity."""
        class_count = len(extraction['classes'])
        function_count = len(extraction['functions'])

        if class_count > 5 or function_count > 20:
            return 'high'
        elif class_count > 2 or function_count > 10:
            return 'medium'
        else:
            return 'low'

    def _analyze_architecture(self, extraction: Dict[str, Any]) -> str:
        """Analyze system architecture."""
        imports = extraction['imports']

        # Check for architectural patterns
        if any('django' in imp.lower() for imp in imports):
            return 'mvc'
        elif any('react' in imp.lower() or 'vue' in imp.lower() for imp in imports):
            return 'frontend'
        elif any('fastapi' in imp.lower() or 'flask' in imp.lower() for imp in imports):
            return 'api'
        elif any('sqlalchemy' in imp.lower() or 'django' in imp.lower() for imp in imports):
            return 'data'
        else:
            return 'simple'

    def _determine_domain(self, extraction: Dict[str, Any]) -> str:
        """Determine the domain based on code analysis."""
        domain_keywords = extraction['domain_keywords']
        imports = extraction['imports']

        # Check for domain indicators
        domain_indicators = {
            'auth': ['auth', 'login', 'password', 'security', 'bcrypt', 'token'],
            'api': ['api', 'endpoint', 'route', 'controller', 'service'],
            'data': ['model', 'entity', 'schema', 'database', 'persistence'],
            'ui': ['ui', 'interface', 'component', 'view', 'template'],
            'business': ['business', 'logic', 'process', 'workflow', 'rule']
        }

        domain_scores = {}
        for domain, keywords in domain_indicators.items():
            score = sum(1 for keyword in keywords if any(keyword in kw for kw in domain_keywords))
            domain_scores[domain] = score

        # Return domain with highest score
        if domain_scores:
            return max(domain_scores, key=domain_scores.get)
        else:
            return 'general'

    def _generate_spec_id(self, extraction: Dict[str, Any], domain: str) -> str:
        """Generate unique SPEC ID."""
        file_name = extraction['file_name']
        domain_upper = domain.upper()

        # Clean file name
        clean_name = re.sub(r'[^a-zA-Z0-9]', '', file_name)

        # Generate hash for uniqueness
        import hashlib
        file_hash = hashlib.md5(f"{file_name}{domain}{time.time()}".encode()).hexdigest()[:4]

        return f"{domain_upper}-{clean_name[:8]}-{file_hash}"

    def _generate_spec_content(self, extraction: Dict[str, Any], domain: str,
                             spec_id: str, custom_config: Dict[str, Any] = None) -> str:
        """Generate main spec.md content."""
        config = custom_config or {}

        # Get domain template
        domain_info = self.domain_templates.get(domain, {
            'description': 'General system',
            'common_features': ['Standard Features'],
            'environment': 'General Purpose'
        })

        # Extract information
        primary_function = self._extract_primary_function(extraction, domain)
        state_requirements = self._generate_state_requirements(extraction, domain)
        event_requirements = self._generate_event_requirements(extraction, domain)
        technical_specs = self._generate_technical_specs(extraction, domain)

        # Generate template content
        spec_content = self._render_template(
            self.ears_templates['environment'],
            {
                'project_name': config.get('project_name', f'{domain.capitalize()} System'),
                'language': extraction['language'],
                'framework': config.get('framework', self._detect_framework(extraction)),
                'paradigm': config.get('paradigm', 'Object-Oriented'),
                'platform': config.get('platform', 'Web/Server'),
                'deployment': config.get('deployment', 'Cloud-based'),
                'status': config.get('status', 'Development'),
                **extraction
            }
        )

        # Add assumptions
        spec_content += "\n\n" + self._render_template(
            self.ears_templates['assumptions'],
            extraction
        )

        # Add requirements
        spec_content += "\n\n" + self._render_template(
            self.ears_templates['requirements'],
            {
                'primary_function': primary_function,
                'state_requirements': state_requirements,
                'event_requirements': event_requirements,
                **extraction
            }
        )

        # Add specifications
        spec_content += "\n\n" + self._render_template(
            self.ears_templates['specifications'],
            {
                'technical_specs': technical_specs,
                **self._generate_technical_details(extraction, domain),
                **extraction
            }
        )

        # Add traceability
        spec_content += self._generate_traceability(spec_id)

        # Add edit guide
        spec_content += self._generate_edit_guide(extraction, domain)

        # Add meta information
        spec_md = f"""---
@META: {{
  "id": "SPEC-{spec_id}",
  "title": "Auto-generated SPEC for {extraction['file_name']}",
  "title_en": "Auto-generated SPEC for {extraction['file_name']}",
  "version": "1.0.0",
  "status": "pending",
  "created": "{time.strftime('%Y-%m-%d')}",
  "author": "@alfred-auto",
  "reviewer": "",
  "category": "FEATURE",
  "priority": "MEDIUM",
  "tags": ["auto-generated", "{spec_id}", "{domain}"],
  "language": "ko",
  "estimated_complexity": "{extraction['complexity']}",
  "domain": "{domain}"
}}
---

# @SPEC:{spec_id}: Auto-generated SPEC for {extraction['file_name']}
## Auto-generated SPEC for {extraction['file_name']}

### 개요 (Overview)

{domain_info['description']}

{spec_content}
"""

        return spec_md

    def _render_template(self, template: Dict[str, str], context: Dict[str, Any]) -> str:
        """Render template with context."""
        template_text = template['template']

        # Replace placeholders
        for key, value in context.items():
            placeholder = f"{{{key}}}"
            template_text = template_text.replace(placeholder, str(value))

        return template_text

    def _extract_primary_function(self, extraction: Dict[str, Any], domain: str) -> str:
        """Extract primary function from code analysis."""
        classes = extraction['classes']
        functions = extraction['functions']

        if classes:
            return f"Manage {classes[0]} class and related operations"
        elif functions:
            return f"Execute {functions[0]} function and related operations"
        else:
            return f"Process data and perform {domain} operations"

    def _generate_state_requirements(self, extraction: Dict[str, Any], domain: str) -> str:
        """Generate state-based requirements."""
        base_requirements = [
            "- **REQ-006**: 시스템은 초기 상태에서 목표 상태로 전환해야 함",
            "- **REQ-007**: 상태 변화는 유효한 조건 하에서만 발생해야 함",
            "- **REQ-008**: 시스템은 각 상태의 무결성을 유지해야 함",
            "- **REQ-009**: 상태 변화는 로깅되어 추적 가능해야 함",
            "- **REQ-010**: 시스템은 오류 상태에서 복구 메커니즘을 제공해야 함"
        ]

        domain_specific = {
            'auth': [
                "- **AUTH-001**: 사용자는 미인증 상태에서 인증 상태로 전환될 수 있어야 함",
                "- **AUTH-002**: 인증 상태에서 시스템에 접근할 수 있어야 함",
                "- **AUTH-003**: 세션 만료 시 자동으로 미인증 상태로 전환되어야 함"
            ],
            'api': [
                "- **API-001**: API는 준비 상태, 실행 상태, 오류 상태를 가질 수 있어야 함",
                "- **API-002**: 오류 상태에서 적절한 오류 응답을 반환해야 함",
                "- **API-003**: 상태 변화는 이벤트로 알려져야 함"
            ],
            'data': [
                "- **DATA-001**: 데이터는 생성, 수정, 삭제 상태를 가질 수 있어야 함",
                "- **DATA-002**: 데이터 무결성은 항상 유지되어야 함",
                "- **DATA-003**: 데이터 백업 상태는 모니터링되어야 함"
            ]
        }

        result = "\n".join(base_requirements)
        if domain in domain_specific:
            result += "\n\n" + "\n".join(domain_specific[domain])

        return result

    def _generate_event_requirements(self, extraction: Dict[str, Any], domain: str) -> str:
        """Generate event-based requirements."""
        base_events = [
            "- **EVT-001**: 사용자 입력 이벤트에 반응해야 함",
            "- **EVT-002**: 시스템 내부 이벤트를 처리해야 함",
            "- **EVT-003**: 외부 서비스 이벤트를 수신해야 함",
            "- **EVT-004**: 이벤트 처리 오류는 적절히 처리해야 함",
            "- **EVT-005**: 이벤트 로그는 유지해야 함"
        ]

        domain_specific = {
            'auth': [
                "- **AUTH-EVT-001**: 로그인 이벤트를 처리해야 함",
                "- **AUTH-EVT-002**: 로그아웃 이벤트를 처리해야 함",
                "- **AUTH-EVT-003**: 비밀번호 변경 이벤트를 처리해야 함"
            ],
            'api': [
                "- **API-EVT-001**: API 요청 이벤트를 처리해야 함",
                "- **API-EVT-002**: 인증 이벤트를 처리해야 함",
                "- **API-EVT-003**: 레이트 리밋 이벤트를 처리해야 함"
            ],
            'data': [
                "- **DATA-EVT-001**: 데이터 저장 이벤트를 처리해야 함",
                "- **DATA-EVT-002**: 데이터 조회 이벤트를 처리해야 함",
                "- **DATA-EVT-003**: 데이터 삭제 이벤트를 처리해야 함"
            ]
        }

        result = "\n".join(base_events)
        if domain in domain_specific:
            result += "\n\n" + "\n".join(domain_specific[domain])

        return result

    def _generate_technical_specs(self, extraction: Dict[str, Any], domain: str) -> str:
        """Generate technical specifications."""
        technical_specs = [
            "#### 핵심 구현사항 (Core Implementation)",

            f"- **SPEC-001**: {extraction['classes'][0] if extraction['classes'] else 'Main'} 클래스를 구현해야 함",
            f"- **SPEC-002**: {extraction['functions'][0] if extraction['functions'] else 'Core'} 함수를 구현해야 함",
            "- **SPEC-003**: 입력 유효성 검증을 구현해야 함",
            "- **SPEC-004**: 오류 처리 메커니즘을 구현해야 함",
            "- **SPEC-005**: 로깅 시스템을 구현해야 함",

            "#### 확장성 (Extensibility)",
            "- **SPEC-006**: 플러그인 아키텍처 지원",
            "- **SPEC-007**: 설정 기반 기능 활성화/비활성화",
            "- **SPEC-008**: 테스트 가능한 설계",

            "#### 유지보수 (Maintainability)",
            "- **SPEC-009**: 코드 문서화",
            "- **SPEC-010**: 단위 테스트 커버리지",
            "- **SPEC-011**: 코드 품질 검증"
        ]

        return "\n".join(technical_specs)

    def _generate_technical_details(self, extraction: Dict[str, Any], domain: str) -> Dict[str, str]:
        """Generate technical details for specifications."""
        return {
            'technical_details': f"""#### 기술적 상세사항 (Technical Details)

- **구조**: {extraction['architecture'].title()} Architecture
- **복잡도**: {extraction['complexity'].title()}
- **언어**: {extraction['language']}
- **모듈 수**: {len(extraction['classes'])} 클래스, {len(extraction['functions'])} 함수
- **의존성**: {len(extraction['imports'])} 개의 외부 의존성

#### 데이터 모델 (Data Models)

{self._generate_data_models(extraction, domain)}

#### API 명세 (API Specification)

{self._generate_api_specs(extraction, domain)}

#### 인터페이스 명세 (Interface Specification)

{self._generate_interface_specs(extraction, domain)}

#### 보안 명세 (Security Specification)

{self._generate_security_specs(extraction, domain)}

#### 성능 명세 (Performance Specification)

{self._generate_performance_specs(extraction, domain)}

#### 확장성 명세 (Scalability Specification)

{self._generate_scalability_specs(extraction, domain)}""",
            'data_models': self._generate_data_models(extraction, domain),
            'api_specs': self._generate_api_specs(extraction, domain),
            'interface_specs': self._generate_interface_specs(extraction, domain),
            'security_specs': self._generate_security_specs(extraction, domain),
            'performance_specs': self._generate_performance_specs(extraction, domain),
            'scalability_specs': self._generate_scalability_specs(extraction, domain)
        }

    def _generate_data_models(self, extraction: Dict[str, Any], domain: str) -> str:
        """Generate data models section."""
        if extraction['classes']:
            models = []
            for class_name in extraction['classes'][:3]:  # Limit to 3 models
                models.append(f"""
**{class_name}**:
- 속성: ID, 생성일시, 상태
- 메서드: 생성, 수정, 삭제, 조회
- 관계: 다른 모델과의 연관관계""")
            return "\n".join(models)
        else:
            return "데이터 모델이 명시적으로 정의되지 않았습니다."

    def _generate_api_specs(self, extraction: Dict[str, Any], domain: str) -> str:
        """Generate API specifications."""
        if domain in ['api', 'auth']:
            return """
**RESTful API Endpoints**:
- `GET /api/{resource}`: 리소스 목록 조회
- `POST /api/{resource}`: 리소스 생성
- `PUT /api/{resource}/{id}`: 리소스 수정
- `DELETE /api/{resource}/{id}`: 리소스 삭제
- `GET /api/{resource}/{id}`: 특정 리소스 조회

**응답 형식**:
- 성공: `200 OK` + JSON 데이터
- 실패: `400 Bad Request`, `404 Not Found`, `500 Internal Server Error`"""
        else:
            return "API 명세는 해당 도메인에 적용되지 않습니다."

    def _generate_interface_specs(self, extraction: Dict[str, Any], domain: str) -> str:
        """Generate interface specifications."""
        if domain in ['ui', 'api']:
            return """
**사용자 인터페이스**:
- 웹 인터페이스: 반응형 디자인
- 모바일 인터페이스: 크로스 플랫폼 호환
- API 인터페이스: RESTful API

**인터랙션 패턴**:
- 사용자 입력 처리
- 실시간 업데이트
- 에러 상태 처리"""
        else:
            return "인터페이스 명세는 해당 도메인에 적용되지 않습니다."

    def _generate_security_specs(self, extraction: Dict[str, Any], domain: str) -> str:
        """Generate security specifications."""
        if domain in ['auth', 'api']:
            return """
**보안 요구사항**:
- 인증 및 인가
- 데이터 암호화
- 입력값 검증
- 접근 제어
- 로깅 모니터링

**보안 대책**:
- 비밀번호 해싱
- 세션 관리
- CSRF 방지
- XSS 방지
- SQL 인젝션 방지"""
        else:
            return "보안 명세는 기본적으로 적용됩니다."

    def _generate_performance_specs(self, extraction: Dict[str, Any], domain: str) -> str:
        """Generate performance specifications."""
        return """
**성능 요구사항**:
- 응답 시간: 1초 이내
- 동시 처리: 최대 1000 요청/초
- 메모리 사용: 최대 512MB
- 처리량: 99.9% 가용성

**성능 모니터링**:
- 응답 시간 모니터링
- 자원 사용량 모니터링
- 에러율 모니터링"""

    def _generate_scalability_specs(self, extraction: Dict[str, Any], domain: str) -> str:
        """Generate scalability specifications."""
        return """
**확장성 요구사항**:
- 수평 확장 지원
- 부하 분산
- 캐싱 전략
- 데이터베이스 샤딩

**확장성 계획**:
- 마이크로서비스 아키텍처
- 컨테이너화
- 오케스트레이션
- CDN 통합"""

    def _generate_traceability(self, spec_id: str) -> str:
        """Generate traceability section."""
        return f"""

### 추적성 (Traceability)

- **@SPEC:{spec_id}** ← **@CODE:HOOK-POST-AUTO-SPEC-001** (후크 자동 생성)
- **@SPEC:{spec_id}** ← **@CODE:CONFIDENCE-SCORING-001** (신뢰도 평가)
- **@SPEC:{spec_id}** ← **@CODE:SPEC-EARS-TEMPLATE-001** (템플릿 생성)
- **@SPEC:{spec_id}** → **@TEST:{spec_id}** (테스트)
- **@SPEC:{spec_id}** → **@CODE:{spec_id}** (구현)"""

    def _generate_edit_guide(self, extraction: Dict[str, Any], domain: str) -> str:
        """Generate edit guide section."""
        return f"""

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

**도메인별 검토 사항:**
- **{domain.upper()}**: {self._get_domain_specific_review(domain)}"""

    def _get_domain_specific_review(self, domain: str) -> str:
        """Get domain-specific review guidance."""
        domain_reviews = {
            'auth': '보안 요구사항 검토, 인증 흐름 확인, 세션 관리 검토',
            'api': 'API 설계 검토, 에러 처리 검토, 성능 검토',
            'data': '데이터 무결성 검토, 백업 및 복원 검토',
            'ui': '사용자 경험 검토, 접근성 검토, 성능 검토',
            'business': '비즈니스 규칙 검토, 컴�라이언스 검토'
        }
        return domain_reviews.get(domain, '일반적인 요구사항 검토')

    def _generate_plan_content(self, extraction: Dict[str, Any], domain: str,
                             spec_id: str, custom_config: Dict[str, Any] = None) -> str:
        """Generate plan.md content."""
        config = custom_config or {}

        # Generate implementation plan based on complexity and domain
        plan_content = f"""---
@META: {{
  "id": "PLAN-{spec_id}",
  "spec_id": "SPEC-{spec_id}",
  "title": "Auto-generated Implementation Plan for {extraction['file_name']}",
  "version": "1.0.0",
  "status": "pending",
  "created": "{time.strftime('%Y-%m-%d')}",
  "author": "@alfred-auto",
  "domain": "{domain}"
}}
---

# @PLAN:{spec_id}: Auto-generated Implementation Plan
## Auto-generated Implementation Plan for {extraction['file_name']}

### 구현 단계 (Implementation Phases)

#### 1단계: 요구사항 분석 (Priority: High)

- [ ] 기능 요구사항 상세화
- [ ] 비기능 요구사항 정의
- [ ] 성능 요구사항 설정
- [ ] 보안 요구사항 정의
- [ ] 사용자 스토리 작성

#### 2단계: 설계 (Priority: High)

- [ ] 아키텍처 설계 완료
- [ ] 데이터 모델 설계
- [ ] API 설계 완료
- [ ] 인터페이스 설계
- [ ] 데이터베이스 스키마 설계

#### 3단계: 개발 (Priority: Medium)

- [ ] 핵심 모듈 개발
- [ ] API 개발 완료
- [ ] 인터페이스 개발
- [ ] 데이터베이스 연동
- [ ] 보안 기능 구현

#### 4단계: 테스트 (Priority: High)

- [ ] 단위 테스트 구현
- [ ] 통합 테스트 구현
- [ ] 시스템 테스트 구현
- [ ] 성능 테스트
- [ ] 보안 테스트

#### 5단계: 배포 (Priority: Medium)

- [ ] 스테이징 환경 배포
- [ ] 배포 자동화 구현
- [ ] 모니터링 설정
- [ ] 문서화 완료
- [ ] 운영 가이드 작성

### 기술적 접근 방식 (Technical Approach)

#### 아키텍처 설계

```
{self._generate_architecture_diagram(extraction, domain)}
```

#### 핵심 컴포넌트

1. **{self._get_main_component(extraction, domain)}**: 주요 비즈니스 로직 처리
2. **{self._get_service_component(extraction, domain)}**: 외부 서비스 연동
3. **{self._get_data_component(extraction, domain)}**: 데이터 처리 및 저장
4. **{self._get_component_4(extraction, domain)}**: 검증 및 처리 계층

#### 의존성 관리

**기존 모듈 활용:**
- 표준 라이브러리 활용
- 기존 인프라 활용

**신규 모듈 추가:**
- {self._get_new_modules(extraction, domain)}

### 성공 기준 (Success Criteria)

#### 기능적 기준

- ✅ 모든 요구사항 구현 완료
- ✅ 테스트 커버리지 85% 이상
- ✅ 성능 목표 충족
- ✅ 사용자 요구사항 충족

#### 성능 기준

- ✅ 응답 시간 1초 이내
- ✅ 메모리 사용량 최적화
- ✅ 병렬 처리 지원
- ✅ 확장성 확인

#### 품질 기준

- ✅ 코드 품질 검증 통과
- ✅ 보안 스캐닝 통과
- ✅ 문서 완성도 확인
- ✅ 유지보수 용이성 검증

### 다음 단계 (Next Steps)

1. **즉시 실행**: 요구사항 분석 (1-2일)
2. **주간 목표**: 설계 완료 (3-5일)
3. **2주 목표**: 개발 완료 (7-14일)
4. **배포 준비**: 테스트 및 검증 (14-16일)
"""

        return plan_content

    def _generate_architecture_diagram(self, extraction: Dict[str, Any], domain: str) -> str:
        """Generate architecture diagram."""
        if domain == 'auth':
            return """
Client → [API Gateway] → [Auth Service] → [Database]
     ↑          ↓           ↓
  [UI Layer]  [Log Service] [Cache]
"""
        elif domain == 'api':
            return """
Client → [Load Balancer] → [API Gateway] → [Service 1]
                                   ↓
                              [Service 2]
                                   ↓
                              [Database]
"""
        elif domain == 'data':
            return """
[Application] → [Data Service] → [Database]
                ↑          ↓
           [Cache Layer] [Analytics]
"""
        else:
            return """
[Client] → [Service] → [Database]
    ↑           ↓
  [UI]      [Cache]
"""

    def _get_main_component(self, extraction: Dict[str, Any], domain: str) -> str:
        """Get main component name."""
        components = {
            'auth': 'AuthService',
            'api': 'APIController',
            'data': 'DataService',
            'ui': 'UIController',
            'business': 'BusinessLogic'
        }
        return components.get(domain, 'MainComponent')

    def _get_service_component(self, extraction: Dict[str, Any], domain: str) -> str:
        """Get service component name."""
        components = {
            'auth': 'UserService',
            'api': 'ExternalService',
            'data': 'PersistenceService',
            'ui': 'ClientService',
            'business': 'WorkflowService'
        }
        return components.get(domain, 'ServiceComponent')

    def _get_data_component(self, extraction: Dict[str, Any], domain: str) -> str:
        """Get data component name."""
        components = {
            'auth': 'UserRepository',
            'api': 'DataRepository',
            'data': 'DataAccessLayer',
            'ui': 'StateManagement',
            'business': 'DataProcessor'
        }
        return components.get(domain, 'DataComponent')

    def _get_component_4(self, extraction: Dict[str, Any], domain: str) -> str:
        """Get fourth component name."""
        components = {
            'auth': 'SecurityManager',
            'api': 'RateLimiter',
            'data': 'DataValidator',
            'ui': 'FormValidator',
            'business': 'RuleEngine'
        }
        return components.get(domain, 'ValidationComponent')

    def _get_new_modules(self, extraction: Dict[str, Any], domain: str) -> str:
        """Get new modules to be added."""
        modules = {
            'auth': '인증 모듈, 보안 모듈, 세션 관리 모듈',
            'api': '라우팅 모듈, 미들웨어 모듈, 인증 모듈',
            'data': '데이터베이스 모듈, 캐시 모듈, 백업 모듈',
            'ui': '컴포넌트 라이브러리, 상태 관리 모듈',
            'business': '비즈니스 규칙 모듈, 워크플로우 모듈'
        }
        return modules.get(domain, '표준 모듈')

    def _generate_acceptance_content(self, extraction: Dict[str, Any], domain: str,
                                   spec_id: str, custom_config: Dict[str, Any] = None) -> str:
        """Generate acceptance.md content."""
        config = custom_config or {}

        acceptance_content = f"""---
@META: {{
  "id": "ACCEPT-{spec_id}",
  "spec_id": "SPEC-{spec_id}",
  "title": "Auto-generated Acceptance Criteria for {extraction['file_name']}",
  "version": "1.0.0",
  "status": "pending",
  "created": "{time.strftime('%Y-%m-%d')}",
  "author": "@alfred-auto",
  "domain": "{domain}"
}}
---

# @ACCEPT:{spec_id}: Auto-generated Acceptance Criteria
## Auto-generated Acceptance Criteria for {extraction['file_name']}

### 검수 기준 (Acceptance Criteria)

#### 기본 기능 검수 (Basic Functionality)

**필수 조건 (Must-have):**
- [ ] 시스템이 정상적으로 구동되어야 함
- [ ] 사용자 인터페이스가 올바르게 표시되어야 함
- [ ] 데이터 처리 로직이 정상적으로 작동해야 함
- [ ] 오류 상황에 대한 적절한 처리가 필요함
- [ ] 로그 기록이 정상적으로 작동해야 함

**필수 조건 (Should-have):**
- [ ] 사용자 경험이 원활해야 함
- [ ] 성능 목표를 충족해야 함
- [ ] 보안 요구사항을 충족해야 함
- [ ] 접근성 표준을 준수해야 함

#### {domain.upper()} 도메인 특화 검수 ({domain.upper()} Domain Specific)

{self._generate_domain_specific_acceptance(domain)}

#### 성능 검수 (Performance Testing)

**성능 요구사항:**
- [ ] 응답 시간: 1초 이내
- [ ] 동시 접속자: 100명 이상 지원
- [ ] 메모리 사용량: 100MB 이하
- [ ] CPU 사용률: 50% 이하

**부하 테스트:**
- [ ] 기능 부하 테스트 통과
- [ ] 장기 안정성 테스트 통과
- [ ] 회복 테스트 통과

#### 보안 검수 (Security Testing)

**보안 요구사항:**
- [ ] 인증 및 권한 검증 통과
- [ ] 입력값 검증 통과
- [ ] SQL 인젝션 방어 통과
- [ ] CSRF 방어 통과
- [ ] XSS 방어 통과

**취약점 테스트:**
- [ ] OWASP Top 10 검사 통과
- [ ] 보안 스캐닝 통과
- [ ] 권한 설정 검증 통과

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

        return acceptance_content

    def _generate_domain_specific_acceptance(self, domain: str) -> str:
        """Generate domain-specific acceptance criteria."""
        domain_criteria = {
            'auth': """
- **AUTH-001**: 사용자 로그인 기능 검수
  - 사용자 ID와 비밀번호로 로그인할 수 있어야 함
  - 성공 시 세션 토큰이 발급되어야 함
  - 실패 시 적절한 오류 메시지가 표시되어야 함
- **AUTH-002**: 사용자 등록 기능 검수
  - 새 사용자를 등록할 수 있어야 함
  - 중복 ID 체크가 정상 작동해야 함
  - 이메일 인증이 필요할 수 있음
- **AUTH-003**: 비밀번호 변경 기능 검수
  - 기존 비밀번호 확인 후 변경 가능
  - 비밀번호 복잡도 검증이 필요함
  - 변경 시 알림 발송""",
            'api': """
- **API-001**: REST API 기능 검수
  - CRUD 연산이 정상 작동해야 함
  - HTTP 상태 코드가 올바르게 반환되어야 함
  - API 버전 관리가 필요함
- **API-002**: 인증 기능 검수
  - API 키 기반 인증이 작동해야 함
  - JWT 토큰 처리가 정상 작동해야 함
  - 권한 수준별 접근 제어가 필요함
- **API-003**: 레이트 리밋 기능 검수
  - 요청 제한이 정상 작동해야 함
  - 제한 초과 시 적절한 오류 반환""",
            'data': """
- **DATA-001**: 데이터 저장 기능 검수
  - 데이터가 정상적으로 저장되어야 함
  - 데이터 무결성이 유지되어야 함
  - 백업 및 복원 기능이 필요함
- **DATA-002**: 데이터 조회 기능 검수
  - 데이터 정확히 조회되어야 함
  - 쿼리 성능이 목표를 충족해야 함
  - 인덱싱이 정상 작동해야 함
- **DATA-003**: 데이터 관리 기능 검수
  - 데이터 수정이 가능해야 함
  - 데이터 삭제가 안전하게 처리되어야 함
  - 데이터 이관 기능이 필요할 수 있음"""
        }
        return domain_criteria.get(domain, "")

    def _validate_ears_compliance(self, spec_content: Dict[str, str]) -> Dict[str, Any]:
        """Validate EARS format compliance."""
        spec_md = spec_content.get('spec_md', '')

        # Check for required sections
        required_sections = [
            '개요 (Overview)',
            '환경 (Environment)',
            '가정 (Assumptions)',
            '요구사항 (Requirements)',
            '명세 (Specifications)',
            '추적성 (Traceability)'
        ]

        section_scores = {}
        for section in required_sections:
            if section in spec_md:
                section_scores[section] = 1.0
            else:
                section_scores[section] = 0.0

        # Calculate overall compliance
        overall_compliance = sum(section_scores.values()) / len(required_sections)

        # Generate suggestions
        suggestions = []
        for section, score in section_scores.items():
            if score < 1.0:
                suggestions.append(f"추가 필요: {section} 섹션을 포함해야 함")

        return {
            'ears_compliance': round(overall_compliance, 2),
            'section_scores': section_scores,
            'suggestions': suggestions[:5],  # Top 5 suggestions
            'total_sections': len(required_sections),
            'present_sections': sum(1 for score in section_scores.values() if score > 0)
        }

    def _detect_framework(self, extraction: Dict[str, Any]) -> str:
        """Detect framework from imports."""
        imports = extraction['imports']

        framework_indicators = {
            'Django': ['django'],
            'Flask': ['flask'],
            'FastAPI': ['fastapi'],
            'Spring': ['spring'],
            'Express': ['express'],
            'React': ['react'],
            'Angular': ['angular'],
            'Vue': ['vue'],
            'Next.js': ['next']
        }

        for framework, indicators in framework_indicators.items():
            for imp in imports:
                if any(indicator in imp.lower() for indicator in indicators):
                    return framework

        return 'Custom'