"""EARS Template Engine for Auto-Generated SPECs."""

import logging
import re
import time
from pathlib import Path
from typing import Any, Dict

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
            "auth": {
                "description": "User authentication and security system",
                "common_features": [
                    "Login",
                    "Registration",
                    "Password Reset",
                    "Session Management",
                ],
                "security_requirements": [
                    "Encryption",
                    "Password Hashing",
                    "Rate Limiting",
                ],
                "environment": "Web Application with User Management",
            },
            "api": {
                "description": "RESTful API service",
                "common_features": [
                    "Endpoints",
                    "Authentication",
                    "Rate Limiting",
                    "Caching",
                ],
                "technical_requirements": [
                    "RESTful Design",
                    "JSON Format",
                    "HTTP Status Codes",
                ],
                "environment": "Microservice Architecture",
            },
            "data": {
                "description": "Data processing and storage system",
                "common_features": [
                    "Data Validation",
                    "Persistence",
                    "Backup",
                    "Migration",
                ],
                "technical_requirements": [
                    "Data Integrity",
                    "Performance",
                    "Scalability",
                ],
                "environment": "Database System with Analytics",
            },
            "ui": {
                "description": "User interface and experience system",
                "common_features": ["Components", "Navigation", "Forms", "Validation"],
                "experience_requirements": [
                    "Responsive Design",
                    "Accessibility",
                    "Performance",
                ],
                "environment": "Web Frontend with React/Angular/Vue",
            },
            "business": {
                "description": "Business logic and workflow system",
                "common_features": [
                    "Process Management",
                    "Rules Engine",
                    "Notifications",
                ],
                "business_requirements": ["Compliance", "Audit Trail", "Reporting"],
                "environment": "Enterprise Application",
            },
        }

        # EARS section templates
        self.ears_templates = {
            "environment": {
                "template": """### Environment

- **Project**: {project_name}
- **Language**: {language}
- **Framework**: {framework}
- **Paradigm**: {paradigm}
- **Platform**: {platform}
- **Deployment**: {deployment}
- **Status**: {status}
- **Generation Method**: Auto-analysis based""",
                "required_fields": [
                    "project_name",
                    "language",
                    "framework",
                    "paradigm",
                ],
            },
            "assumptions": {
                "template": """### Assumptions

1. System follows standard development practices
2. Users have basic domain knowledge
3. System is designed with stable and scalable architecture
4. External dependencies operate normally
5. Security requirements comply with industry standards
6. Data integrity is maintained
7. User interface is intuitively designed
8. Performance requirements are met""",
                "required_fields": [],
            },
            "requirements": {
                "template": """### Requirements

#### Ubiquitous Requirements

- **REQ-001**: System SHALL perform {primary_function} functionality
- **REQ-002**: Generated features SHALL be stable
- **REQ-003**: Code SHALL be maintainable
- **REQ-004**: Tests SHALL meet functional requirements
- **REQ-005**: Code SHALL comply with project coding standards
- **REQ-006**: System SHALL handle exceptional situations appropriately
- **REQ-007**: User experience SHALL be optimized

#### State-driven Requirements

{state_requirements}

#### Event-driven Requirements

{event_requirements}

#### Optional Requirements

- **REQ-008**: System SHALL include performance monitoring features
- **REQ-009**: Automatic backup and restore features MAY be required
- **REQ-010**: User activity logging MAY be required
- **REQ-011**: Multi-language support MAY be required
- **REQ-012**: Mobile compatibility MAY be required""",
                "required_fields": ["primary_function"],
            },
            "specifications": {
                "template": """### Specifications

{technical_specs}

#### Technical Specifications

{technical_details}

#### Data Models

{data_models}

#### API Specifications

{api_specs}

#### Interface Specifications

{interface_specs}

#### Security Specifications

{security_specs}

#### Performance Specifications

{performance_specs}

#### Scalability Specifications

{scalability_specs}""",
                "required_fields": [],
            },
        }

    def generate_complete_spec(
        self,
        code_analysis: Dict[str, Any],
        file_path: str,
        custom_config: Dict[str, Any] = None,
    ) -> Dict[str, str]:
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
        validation_result = self._validate_ears_compliance(
            {
                "spec_md": spec_md_content,
                "plan_md": plan_md_content,
                "acceptance_md": acceptance_md_content,
            }
        )

        # Create result
        result = {
            "spec_id": spec_id,
            "domain": domain,
            "spec_md": spec_md_content,
            "plan_md": plan_md_content,
            "acceptance_md": acceptance_md_content,
            "validation": validation_result,
            "generation_time": time.time() - start_time,
            "extraction": extraction_result,
        }

        return result  # type: ignore[return-value]

    def _extract_information_from_analysis(self, code_analysis: Dict[str, Any], file_path: str) -> Dict[str, Any]:
        """Extract information from code analysis."""
        extraction = {
            "file_path": file_path,
            "file_name": Path(file_path).stem,
            "file_extension": Path(file_path).suffix,
            "language": self._detect_language(file_path),
            "classes": [],
            "functions": [],
            "imports": [],
            "domain_keywords": [],
            "technical_indicators": [],
            "complexity": "low",
            "architecture": "simple",
        }

        # Extract from code_analysis
        if "structure_info" in code_analysis:
            structure = code_analysis["structure_info"]
            extraction["classes"] = structure.get("classes", [])
            extraction["functions"] = structure.get("functions", [])
            extraction["imports"] = structure.get("imports", [])

        if "domain_keywords" in code_analysis:
            extraction["domain_keywords"] = code_analysis["domain_keywords"]

        # Extract from AST analysis if available
        if hasattr(code_analysis, "ast_info"):
            pass
            # Additional extraction logic here

        # Determine complexity and architecture
        extraction["complexity"] = self._analyze_complexity(extraction)
        extraction["architecture"] = self._analyze_architecture(extraction)

        return extraction

    def _detect_language(self, file_path: str) -> str:
        """Detect programming language from file path."""
        extension = Path(file_path).suffix.lower()

        language_map = {
            ".py": "Python",
            ".js": "JavaScript",
            ".jsx": "JavaScript",
            ".ts": "TypeScript",
            ".tsx": "TypeScript",
            ".go": "Go",
            ".java": "Java",
            ".cpp": "C++",
            ".c": "C",
            ".cs": "C#",
            ".rb": "Ruby",
            ".php": "PHP",
            ".swift": "Swift",
            ".kt": "Kotlin",
        }

        return language_map.get(extension, "Unknown")

    def _analyze_complexity(self, extraction: Dict[str, Any]) -> str:
        """Analyze code complexity."""
        class_count = len(extraction["classes"])
        function_count = len(extraction["functions"])

        if class_count > 5 or function_count > 20:
            return "high"
        elif class_count > 2 or function_count > 10:
            return "medium"
        else:
            return "low"

    def _analyze_architecture(self, extraction: Dict[str, Any]) -> str:
        """Analyze system architecture."""
        imports = extraction["imports"]

        # Check for architectural patterns
        if any("django" in imp.lower() for imp in imports):
            return "mvc"
        elif any("react" in imp.lower() or "vue" in imp.lower() for imp in imports):
            return "frontend"
        elif any("fastapi" in imp.lower() or "flask" in imp.lower() for imp in imports):
            return "api"
        elif any("sqlalchemy" in imp.lower() or "django" in imp.lower() for imp in imports):
            return "data"
        else:
            return "simple"

    def _determine_domain(self, extraction: Dict[str, Any]) -> str:
        """Determine the domain based on code analysis."""
        domain_keywords = extraction["domain_keywords"]
        extraction["imports"]

        # Check for domain indicators
        domain_indicators = {
            "auth": ["auth", "login", "password", "security", "bcrypt", "token"],
            "api": ["api", "endpoint", "route", "controller", "service"],
            "data": ["model", "entity", "schema", "database", "persistence"],
            "ui": ["ui", "interface", "component", "view", "template"],
            "business": ["business", "logic", "process", "workflow", "rule"],
        }

        domain_scores = {}
        for domain, keywords in domain_indicators.items():
            score = sum(1 for keyword in keywords if any(keyword in kw for kw in domain_keywords))
            domain_scores[domain] = score

        # Return domain with highest score
        if domain_scores:
            return max(domain_scores, key=domain_scores.get)
        else:
            return "general"

    def _generate_spec_id(self, extraction: Dict[str, Any], domain: str) -> str:
        """Generate unique SPEC ID."""
        file_name = extraction["file_name"]
        domain_upper = domain.upper()

        # Clean file name
        clean_name = re.sub(r"[^a-zA-Z0-9]", "", file_name)

        # Generate hash for uniqueness
        import hashlib

        file_hash = hashlib.md5(f"{file_name}{domain}{time.time()}".encode(), usedforsecurity=False).hexdigest()[:4]

        return f"{domain_upper}-{clean_name[:8]}-{file_hash}"

    def _generate_spec_content(
        self,
        extraction: Dict[str, Any],
        domain: str,
        spec_id: str,
        custom_config: Dict[str, Any] = None,
    ) -> str:
        """Generate main spec.md content."""
        config = custom_config or {}

        # Get domain template
        domain_info = self.domain_templates.get(
            domain,
            {
                "description": "General system",
                "common_features": ["Standard Features"],
                "environment": "General Purpose",
            },
        )

        # Extract information
        primary_function = self._extract_primary_function(extraction, domain)
        state_requirements = self._generate_state_requirements(extraction, domain)
        event_requirements = self._generate_event_requirements(extraction, domain)
        technical_specs = self._generate_technical_specs(extraction, domain)

        # Generate template content
        spec_content = self._render_template(
            self.ears_templates["environment"],
            {
                "project_name": config.get("project_name", f"{domain.capitalize()} System"),
                "language": extraction["language"],
                "framework": config.get("framework", self._detect_framework(extraction)),
                "paradigm": config.get("paradigm", "Object-Oriented"),
                "platform": config.get("platform", "Web/Server"),
                "deployment": config.get("deployment", "Cloud-based"),
                "status": config.get("status", "Development"),
                **extraction,
            },
        )

        # Add assumptions
        spec_content += "\n\n" + self._render_template(self.ears_templates["assumptions"], extraction)

        # Add requirements
        spec_content += "\n\n" + self._render_template(
            self.ears_templates["requirements"],
            {
                "primary_function": primary_function,
                "state_requirements": state_requirements,
                "event_requirements": event_requirements,
                **extraction,
            },
        )

        # Add specifications
        spec_content += "\n\n" + self._render_template(
            self.ears_templates["specifications"],
            {
                "technical_specs": technical_specs,
                **self._generate_technical_details(extraction, domain),
                **extraction,
            },
        )

        # Add traceability
        spec_content += self._generate_traceability(spec_id)

        # Add edit guide
        spec_content += self._generate_edit_guide(extraction, domain)

        # Add meta information
        spec_md = f"""---
  "id": "SPEC-{spec_id}",
  "title": "Auto-generated SPEC for {extraction["file_name"]}",
  "title_en": "Auto-generated SPEC for {extraction["file_name"]}",
  "version": "1.0.0",
  "status": "pending",
  "created": "{time.strftime("%Y-%m-%d")}",
  "author": "@alfred-auto",
  "reviewer": "",
  "category": "FEATURE",
  "priority": "MEDIUM",
  "tags": ["auto-generated", "{spec_id}", "{domain}"],
  "language": "en",
  "estimated_complexity": "{extraction["complexity"]}",
  "domain": "{domain}"
}}
---

## Auto-generated SPEC for {extraction["file_name"]}

### Overview

{domain_info["description"]}

{spec_content}
"""

        return spec_md

    def _render_template(self, template: Dict[str, str], context: Dict[str, Any]) -> str:
        """Render template with context."""
        template_text = template["template"]

        # Replace placeholders
        for key, value in context.items():
            placeholder = f"{{{key}}}"
            # Handle both string and sequence values
            if isinstance(value, (list, tuple)):
                # Convert sequence to newline-separated string
                str_value = "\n".join(str(v) for v in value)
            else:
                str_value = str(value)
            template_text = template_text.replace(placeholder, str_value)

        return template_text

    def _extract_primary_function(self, extraction: Dict[str, Any], domain: str) -> str:
        """Extract primary function from code analysis."""
        classes = extraction["classes"]
        functions = extraction["functions"]

        if classes:
            return f"Manage {classes[0]} class and related operations"
        elif functions:
            return f"Execute {functions[0]} function and related operations"
        else:
            return f"Process data and perform {domain} operations"

    def _generate_state_requirements(self, extraction: Dict[str, Any], domain: str) -> str:
        """Generate state-based requirements."""
        base_requirements = [
            "- **REQ-006**: System SHALL transition from initial state to target state",
            "- **REQ-007**: State changes SHALL occur only under valid conditions",
            "- **REQ-008**: System SHALL maintain integrity of each state",
            "- **REQ-009**: State changes SHALL be logged and traceable",
            "- **REQ-010**: System SHALL provide recovery mechanism from error state",
        ]

        domain_specific = {
            "auth": [
                "- **AUTH-001**: User SHALL be able to transition from unauthenticated to authenticated state",
                "- **AUTH-002**: System SHALL be accessible in authenticated state",
                "- **AUTH-003**: System SHALL automatically transition to unauthenticated state on session expiry",
            ],
            "api": [
                "- **API-001**: API SHALL have ready, executing, and error states",
                "- **API-002**: System SHALL return appropriate error response in error state",
                "- **API-003**: State changes SHALL be notified as events",
            ],
            "data": [
                "- **DATA-001**: Data SHALL have create, update, and delete states",
                "- **DATA-002**: Data integrity SHALL be maintained at all times",
                "- **DATA-003**: Data backup state SHALL be monitored",
            ],
        }

        result = "\n".join(base_requirements)
        if domain in domain_specific:
            result += "\n\n" + "\n".join(domain_specific[domain])

        return result

    def _generate_event_requirements(self, extraction: Dict[str, Any], domain: str) -> str:
        """Generate event-based requirements."""
        base_events = [
            "- **EVT-001**: System SHALL respond to user input events",
            "- **EVT-002**: System SHALL handle internal system events",
            "- **EVT-003**: System SHALL receive external service events",
            "- **EVT-004**: Event processing errors SHALL be handled appropriately",
            "- **EVT-005**: Event logs SHALL be maintained",
        ]

        domain_specific = {
            "auth": [
                "- **AUTH-EVT-001**: System SHALL handle login events",
                "- **AUTH-EVT-002**: System SHALL handle logout events",
                "- **AUTH-EVT-003**: System SHALL handle password change events",
            ],
            "api": [
                "- **API-EVT-001**: System SHALL handle API request events",
                "- **API-EVT-002**: System SHALL handle authentication events",
                "- **API-EVT-003**: System SHALL handle rate limit events",
            ],
            "data": [
                "- **DATA-EVT-001**: System SHALL handle data save events",
                "- **DATA-EVT-002**: System SHALL handle data retrieval events",
                "- **DATA-EVT-003**: System SHALL handle data deletion events",
            ],
        }

        result = "\n".join(base_events)
        if domain in domain_specific:
            result += "\n\n" + "\n".join(domain_specific[domain])

        return result

    def _generate_technical_specs(self, extraction: Dict[str, Any], domain: str) -> str:
        """Generate technical specifications."""
        technical_specs = [
            "#### Core Implementation",
            f"- **SPEC-001**: {extraction['classes'][0] if extraction['classes'] else 'Main'} "
            "class SHALL be implemented",
            f"- **SPEC-002**: {extraction['functions'][0] if extraction['functions'] else 'Core'} "
            "function SHALL be implemented",
            "- **SPEC-003**: Input validation SHALL be implemented",
            "- **SPEC-004**: Error handling mechanism SHALL be implemented",
            "- **SPEC-005**: Logging system SHALL be implemented",
            "#### Extensibility",
            "- **SPEC-006**: Plugin architecture support",
            "- **SPEC-007**: Configuration-based feature enable/disable",
            "- **SPEC-008**: Testable design",
            "#### Maintainability",
            "- **SPEC-009**: Code documentation",
            "- **SPEC-010**: Unit test coverage",
            "- **SPEC-011**: Code quality validation",
        ]

        return "\n".join(technical_specs)

    def _generate_technical_details(self, extraction: Dict[str, Any], domain: str) -> Dict[str, str]:
        """Generate technical details for specifications."""
        return {
            "technical_details": f"""#### Technical Details

- **Architecture**: {extraction["architecture"].title()} Architecture
- **Complexity**: {extraction["complexity"].title()}
- **Language**: {extraction["language"]}
- **Module Count**: {len(extraction["classes"])} classes, {len(extraction["functions"])} functions
- **Dependencies**: {len(extraction["imports"])} external dependencies

#### Data Models

{self._generate_data_models(extraction, domain)}

#### API Specification

{self._generate_api_specs(extraction, domain)}

#### Interface Specification

{self._generate_interface_specs(extraction, domain)}

#### Security Specification

{self._generate_security_specs(extraction, domain)}

#### Performance Specification

{self._generate_performance_specs(extraction, domain)}

#### Scalability Specification

{self._generate_scalability_specs(extraction, domain)}""",
            "data_models": self._generate_data_models(extraction, domain),
            "api_specs": self._generate_api_specs(extraction, domain),
            "interface_specs": self._generate_interface_specs(extraction, domain),
            "security_specs": self._generate_security_specs(extraction, domain),
            "performance_specs": self._generate_performance_specs(extraction, domain),
            "scalability_specs": self._generate_scalability_specs(extraction, domain),
        }

    def _generate_data_models(self, extraction: Dict[str, Any], domain: str) -> str:
        """Generate data models section."""
        if extraction["classes"]:
            models = []
            for class_name in extraction["classes"][:3]:  # Limit to 3 models
                models.append(
                    f"""
**{class_name}**:
- Attributes: ID, created_at, status
- Methods: create, update, delete, retrieve
- Relations: Relationships with other models"""
                )
            return "\n".join(models)
        else:
            return "Data models are not explicitly defined."

    def _generate_api_specs(self, extraction: Dict[str, Any], domain: str) -> str:
        """Generate API specifications."""
        if domain in ["api", "auth"]:
            return """
**RESTful API Endpoints**:
- `GET /api/{resource}`: Retrieve resource list
- `POST /api/{resource}`: Create resource
- `PUT /api/{resource}/{id}`: Update resource
- `DELETE /api/{resource}/{id}`: Delete resource
- `GET /api/{resource}/{id}`: Retrieve specific resource

**Response Format**:
- Success: `200 OK` + JSON data
- Failure: `400 Bad Request`, `404 Not Found`, `500 Internal Server Error`"""
        else:
            return "API specification not applicable to this domain."

    def _generate_interface_specs(self, extraction: Dict[str, Any], domain: str) -> str:
        """Generate interface specifications."""
        if domain in ["ui", "api"]:
            return """
**User Interface**:
- Web Interface: Responsive design
- Mobile Interface: Cross-platform compatible
- API Interface: RESTful API

**Interaction Patterns**:
- User input handling
- Real-time updates
- Error state handling"""
        else:
            return "Interface specification not applicable to this domain."

    def _generate_security_specs(self, extraction: Dict[str, Any], domain: str) -> str:
        """Generate security specifications."""
        if domain in ["auth", "api"]:
            return """
**Security Requirements**:
- Authentication and authorization
- Data encryption
- Input validation
- Access control
- Logging and monitoring

**Security Measures**:
- Password hashing
- Session management
- CSRF prevention
- XSS prevention
- SQL injection prevention"""
        else:
            return "Security specifications apply by default."

    def _generate_performance_specs(self, extraction: Dict[str, Any], domain: str) -> str:
        """Generate performance specifications."""
        return """
**Performance Requirements**:
- Response time: Within 1 second
- Concurrent processing: Maximum 1000 requests/second
- Memory usage: Maximum 512MB
- Throughput: 99.9% availability

**Performance Monitoring**:
- Response time monitoring
- Resource usage monitoring
- Error rate monitoring"""

    def _generate_scalability_specs(self, extraction: Dict[str, Any], domain: str) -> str:
        """Generate scalability specifications."""
        return """
**Scalability Requirements**:
- Horizontal scaling support
- Load balancing
- Caching strategy
- Database sharding

**Scalability Plan**:
- Microservice architecture
- Containerization
- Orchestration
- CDN integration"""

    def _generate_traceability(self, spec_id: str) -> str:
        """Generate traceability section."""
        return """

### Traceability

**Requirements Traceability Matrix:**
- Functional requirements → Design specifications → Test cases
- Non-functional requirements → Architecture decisions → Validation tests
- Business requirements → User stories → Acceptance criteria

**Implementation Traceability:**
- Design specifications → Code modules → Unit tests
- API specifications → Endpoints → Integration tests
- Database design → Schemas → Data validation tests

**Change Management:**
- All changes tracked with timestamps
- Impact analysis documented
- Stakeholder approvals recorded"""

    def _generate_edit_guide(self, extraction: Dict[str, Any], domain: str) -> str:
        """Generate edit guide section."""
        return f"""

### Edit Guide

**User Review Checklist:**
1. [OK] Verify technical clarity
2. [OK] Specify requirements in detail
3. [OK] Review domain terminology
4. [OK] Define state and event requirements
5. [OK] Detail specifications

**Quality Improvement Suggestions:**
- Add domain-specific terminology
- Specify user cases in detail
- Define performance requirements
- Add security requirements

**Domain-Specific Review:**
- **{domain.upper()}**: {self._get_domain_specific_review(domain)}"""

    def _get_domain_specific_review(self, domain: str) -> str:
        """Get domain-specific review guidance."""
        domain_reviews = {
            "auth": "Review security requirements, verify authentication flow, review session management",
            "api": "Review API design, review error handling, review performance",
            "data": "Review data integrity, review backup and restore",
            "ui": "Review user experience, review accessibility, review performance",
            "business": "Review business rules, review compliance",
        }
        return domain_reviews.get(domain, "Review general requirements")

    def _generate_plan_content(
        self,
        extraction: Dict[str, Any],
        domain: str,
        spec_id: str,
        custom_config: Dict[str, Any] = None,
    ) -> str:
        """Generate plan.md content."""

        # Generate implementation plan based on complexity and domain
        plan_content = f"""---
id: "PLAN-{spec_id}"
spec_id: "SPEC-{spec_id}"
title: "Auto-generated Implementation Plan for {extraction["file_name"]}"
version: "1.0.0"
status: "pending"
created: "{time.strftime("%Y-%m-%d")}"
author: "@alfred-auto"
domain: "{domain}"
---
## Auto-generated Implementation Plan for {extraction["file_name"]}

### Implementation Phases

#### Phase 1: Requirements Analysis (Priority: High)

- [ ] Detail functional requirements
- [ ] Define non-functional requirements
- [ ] Set performance requirements
- [ ] Define security requirements
- [ ] Write user stories

#### Phase 2: Design (Priority: High)

- [ ] Complete architecture design
- [ ] Design data models
- [ ] Complete API design
- [ ] Design interfaces
- [ ] Design database schema

#### Phase 3: Development (Priority: Medium)

- [ ] Develop core modules
- [ ] Complete API development
- [ ] Develop interfaces
- [ ] Integrate database
- [ ] Implement security features

#### Phase 4: Testing (Priority: High)

- [ ] Implement unit tests
- [ ] Implement integration tests
- [ ] Implement system tests
- [ ] Performance testing
- [ ] Security testing

#### Phase 5: Deployment (Priority: Medium)

- [ ] Deploy to staging environment
- [ ] Implement deployment automation
- [ ] Configure monitoring
- [ ] Complete documentation
- [ ] Write operational guide

### Technical Approach

#### Architecture Design

```
{self._generate_architecture_diagram(extraction, domain)}
```

#### Core Components

1. **{self._get_main_component(extraction, domain)}**: Main business logic processing
2. **{self._get_service_component(extraction, domain)}**: External service integration
3. **{self._get_data_component(extraction, domain)}**: Data processing and storage
4. **{self._get_component_4(extraction, domain)}**: Validation and processing layer

#### Dependency Management

**Utilize Existing Modules:**
- Utilize standard libraries
- Utilize existing infrastructure

**Add New Modules:**
- {self._get_new_modules(extraction, domain)}

### Success Criteria

#### Functional Criteria

- ✅ All requirements implemented
- ✅ Test coverage above 85%
- ✅ Performance goals met
- ✅ User requirements satisfied

#### Performance Criteria

- ✅ Response time within 1 second
- ✅ Memory usage optimized
- ✅ Parallel processing supported
- ✅ Scalability verified

#### Quality Criteria

- ✅ Code quality validation passed
- ✅ Security scanning passed
- ✅ Documentation completeness verified
- ✅ Maintainability verified

### Next Steps

1. **Immediate**: Requirements analysis (1-2 days)
2. **Weekly Goal**: Complete design (3-5 days)
3. **2-Week Goal**: Complete development (7-14 days)
4. **Deployment Prep**: Testing and verification (14-16 days)
"""

        return plan_content

    def _generate_architecture_diagram(self, extraction: Dict[str, Any], domain: str) -> str:
        """Generate architecture diagram."""
        if domain == "auth":
            return """
Client → [API Gateway] → [Auth Service] → [Database]
     ↑          ↓           ↓
  [UI Layer]  [Log Service] [Cache]
"""
        elif domain == "api":
            return """
Client → [Load Balancer] → [API Gateway] → [Service 1]
                                   ↓
                              [Service 2]
                                   ↓
                              [Database]
"""
        elif domain == "data":
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
            "auth": "AuthService",
            "api": "APIController",
            "data": "DataService",
            "ui": "UIController",
            "business": "BusinessLogic",
        }
        return components.get(domain, "MainComponent")

    def _get_service_component(self, extraction: Dict[str, Any], domain: str) -> str:
        """Get service component name."""
        components = {
            "auth": "UserService",
            "api": "ExternalService",
            "data": "PersistenceService",
            "ui": "ClientService",
            "business": "WorkflowService",
        }
        return components.get(domain, "ServiceComponent")

    def _get_data_component(self, extraction: Dict[str, Any], domain: str) -> str:
        """Get data component name."""
        components = {
            "auth": "UserRepository",
            "api": "DataRepository",
            "data": "DataAccessLayer",
            "ui": "StateManagement",
            "business": "DataProcessor",
        }
        return components.get(domain, "DataComponent")

    def _get_component_4(self, extraction: Dict[str, Any], domain: str) -> str:
        """Get fourth component name."""
        components = {
            "auth": "SecurityManager",
            "api": "RateLimiter",
            "data": "DataValidator",
            "ui": "FormValidator",
            "business": "RuleEngine",
        }
        return components.get(domain, "ValidationComponent")

    def _get_new_modules(self, extraction: Dict[str, Any], domain: str) -> str:
        """Get new modules to be added."""
        modules = {
            "auth": "Authentication module, Security module, Session management module",
            "api": "Routing module, Middleware module, Authentication module",
            "data": "Database module, Cache module, Backup module",
            "ui": "Component library, State management module",
            "business": "Business rules module, Workflow module",
        }
        return modules.get(domain, "Standard modules")

    def _generate_acceptance_content(
        self,
        extraction: Dict[str, Any],
        domain: str,
        spec_id: str,
        custom_config: Dict[str, Any] = None,
    ) -> str:
        """Generate acceptance.md content."""

        acceptance_content = f"""---
  "id": "ACCEPT-{spec_id}",
  "spec_id": "SPEC-{spec_id}",
  "title": "Auto-generated Acceptance Criteria for {extraction["file_name"]}",
  "version": "1.0.0",
  "status": "pending",
  "created": "{time.strftime("%Y-%m-%d")}",
  "author": "@alfred-auto",
  "domain": "{domain}"
}}
---

## Auto-generated Acceptance Criteria for {extraction["file_name"]}

### Acceptance Criteria

#### Basic Functionality

**Must-have:**
- [ ] System SHALL operate normally
- [ ] User interface SHALL display correctly
- [ ] Data processing logic SHALL operate correctly
- [ ] Error situations SHALL be handled appropriately
- [ ] Logging SHALL operate correctly

**Should-have:**
- [ ] User experience SHALL be smooth
- [ ] Performance goals SHALL be met
- [ ] Security requirements SHALL be met
- [ ] Accessibility standards SHALL be complied with

#### {domain.upper()} Domain Specific

{self._generate_domain_specific_acceptance(domain)}

#### Performance Testing

**Performance Requirements:**
- [ ] Response time: Within 1 second
- [ ] Concurrent users: Support 100+ users
- [ ] Memory usage: Under 100MB
- [ ] CPU usage: Under 50%

**Load Testing:**
- [ ] Pass functional load tests
- [ ] Pass long-term stability tests
- [ ] Pass recovery tests

#### Security Testing

**Security Requirements:**
- [ ] Pass authentication and authorization validation
- [ ] Pass input validation
- [ ] Pass SQL injection defense
- [ ] Pass CSRF defense
- [ ] Pass XSS defense

**Vulnerability Testing:**
- [ ] Pass OWASP Top 10 checks
- [ ] Pass security scanning
- [ ] Pass permission configuration validation

### Validation Process

#### Phase 1: Unit Tests

- [ ] Complete developer testing
- [ ] Pass code review
- [ ] Pass automated testing
- [ ] Code coverage above 85%

#### Phase 2: Integration Tests

- [ ] Inter-module integration testing
- [ ] API integration testing
- [ ] Database integration testing
- [ ] External service integration testing

#### Phase 3: System Tests

- [ ] Full system functionality testing
- [ ] Performance testing
- [ ] Security testing
- [ ] Stability testing

#### Phase 4: User Tests

- [ ] Internal user testing
- [ ] Actual user testing
- [ ] Feedback collection and incorporation
- [ ] Final validation approval

### Validation Templates

#### Functional Validation Template

| Function ID | Function Name | Expected Result | Actual Result | Status | Notes |
|-------------|---------------|-----------------|---------------|--------|-------|
| FUNC-001 | Feature 1 | Success | Testing | In Progress | Description |
| FUNC-002 | Feature 2 | Success | Success | Passed | Description |
| FUNC-003 | Feature 3 | Success | Failed | Failed | Description |

#### Performance Validation Template

| Test Item | Target | Measured | Status | Notes |
|-----------|--------|----------|--------|-------|
| Response Time | 1 sec | 0.8 sec | Passed | Description |
| Memory Usage | 100MB | 85MB | Passed | Description |
| CPU Usage | 50% | 45% | Passed | Description |

### Completion Criteria

#### Pass Criteria

- ✅ All essential functionality validation passed
- ✅ Performance requirements met
- ✅ Security testing passed
- ✅ User validation passed
- ✅ Documentation validation completed

#### Reporting

- [ ] Write validation report
- [ ] Organize discovered issues list
- [ ] Define improvements
- [ ] Write validation approval document

**Validation Team:**
- Developer: @developer
- QA: @qa_engineer
- Product Owner: @product_owner
- Final Validator: @stakeholder
"""

        return acceptance_content

    def _generate_domain_specific_acceptance(self, domain: str) -> str:
        """Generate domain-specific acceptance criteria."""
        domain_criteria = {
            "auth": """
- **AUTH-001**: User login functionality validation
  - SHALL allow login with user ID and password
  - SHALL issue session token on success
  - SHALL display appropriate error message on failure
- **AUTH-002**: User registration functionality validation
  - SHALL allow new user registration
  - SHALL check for duplicate IDs correctly
  - MAY require email verification
- **AUTH-003**: Password change functionality validation
  - SHALL allow password change after verification
  - SHALL validate password complexity
  - SHALL send notification on change""",
            "api": """
- **API-001**: REST API functionality validation
  - SHALL operate CRUD operations correctly
  - SHALL return correct HTTP status codes
  - SHALL manage API versioning
- **API-002**: Authentication functionality validation
  - SHALL operate API key-based authentication
  - SHALL process JWT tokens correctly
  - SHALL implement permission-level access control
- **API-003**: Rate limiting functionality validation
  - SHALL operate request limits correctly
  - SHALL return appropriate errors when limit exceeded""",
            "data": """
- **DATA-001**: Data storage functionality validation
  - SHALL store data correctly
  - SHALL maintain data integrity
  - SHALL provide backup and restore functionality
- **DATA-002**: Data retrieval functionality validation
  - SHALL retrieve data accurately
  - SHALL meet query performance goals
  - SHALL operate indexing correctly
- **DATA-003**: Data management functionality validation
  - SHALL allow data modification
  - SHALL handle data deletion safely
  - MAY require data migration functionality""",
        }
        return domain_criteria.get(domain, "")

    def _validate_ears_compliance(self, spec_content: Dict[str, str]) -> Dict[str, Any]:
        """Validate EARS format compliance."""
        spec_md = spec_content.get("spec_md", "")

        # Check for required sections
        required_sections = [
            "Overview",
            "Environment",
            "Assumptions",
            "Requirements",
            "Specifications",
            "Traceability",
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
                suggestions.append(f"Required: {section} section must be included")

        return {
            "ears_compliance": round(overall_compliance, 2),
            "section_scores": section_scores,
            "suggestions": suggestions[:5],  # Top 5 suggestions
            "total_sections": len(required_sections),
            "present_sections": sum(1 for score in section_scores.values() if score > 0),
        }

    def _detect_framework(self, extraction: Dict[str, Any]) -> str:
        """Detect framework from imports."""
        imports = extraction["imports"]

        framework_indicators = {
            "Django": ["django"],
            "Flask": ["flask"],
            "FastAPI": ["fastapi"],
            "Spring": ["spring"],
            "Express": ["express"],
            "React": ["react"],
            "Angular": ["angular"],
            "Vue": ["vue"],
            "Next.js": ["next"],
        }

        for framework, indicators in framework_indicators.items():
            for imp in imports:
                if any(indicator in imp.lower() for indicator in indicators):
                    return framework

        return "Custom"
