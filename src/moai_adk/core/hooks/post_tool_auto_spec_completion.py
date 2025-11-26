"""PostToolUse Hook for Automated SPEC Completion System."""

import hashlib
import logging
import os
import re
import time
from typing import Any, Dict, List


# SpecGenerator: Placeholder for spec generation functionality
class SpecGenerator:
    """Placeholder SpecGenerator class for auto-spec completion."""

    def __init__(self):
        self.name = "SpecGenerator"

    def generate_spec(self, file_path: str, content: str) -> str:
        """Generate a basic SPEC document."""
        return f"SPEC document for {file_path}\n\nContent analysis:\n{content[:200]}..."

    def analyze(self, file_path: str) -> Dict[str, Any]:
        """Analyze code file for SPEC generation."""
        return {
            "file_path": file_path,
            "structure_info": {},
            "domain_keywords": [],
        }


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
            return config.get_value(
                "auto_spec_completion",
                {
                    "enabled": True,
                    "min_confidence": 0.7,
                    "auto_open_editor": True,
                    "supported_languages": ["python", "javascript", "typescript", "go"],
                    "excluded_patterns": ["test_", "spec_", "__tests__"],
                },
            )
        except ImportError:
            return {
                "enabled": True,
                "min_confidence": 0.7,
                "auto_open_editor": True,
                "supported_languages": ["python", "javascript", "typescript", "go"],
                "excluded_patterns": ["test_", "spec_", "__tests__"],
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
        if not self.auto_config.get("enabled", True):
            logger.debug("Auto-spec completion is disabled")
            return False

        # Only trigger for Write/Edit/MultiEdit tools
        if tool_name not in ["Write", "Edit", "MultiEdit"]:
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
        if "file_path" in tool_args:
            file_paths.append(tool_args["file_path"])

        # Handle Edit tool
        if "file_path" in tool_args:
            file_paths.append(tool_args["file_path"])

        # Handle MultiEdit tool
        if "edits" in tool_args:
            for edit in tool_args["edits"]:
                if "file_path" in edit:
                    file_paths.append(edit["file_path"])

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
            ".py": "python",
            ".js": "javascript",
            ".jsx": "javascript",
            ".ts": "typescript",
            ".tsx": "typescript",
            ".go": "go",
        }

        if file_ext not in supported_extensions:
            return False

        # Check if language is supported
        language = supported_extensions[file_ext]
        supported_languages = self.auto_config.get("supported_languages", [])
        return language in supported_languages

    def _is_excluded_file(self, file_path: str) -> bool:
        """Check if file should be excluded from auto-spec completion."""
        file_name = os.path.basename(file_path)
        file_dir = os.path.basename(os.path.dirname(file_path))

        excluded_patterns = self.auto_config.get("excluded_patterns", [])

        for pattern in excluded_patterns:
            # Check filename patterns
            if re.search(pattern, file_name):
                return True
            # Check directory patterns
            if re.search(pattern, file_dir):
                return True

        return False

    def detect_code_changes(self, tool_name: str, tool_args: Dict[str, Any], result: Any) -> List[str]:
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
        if tool_name == "Write":
            if "file_path" in tool_args:
                file_paths.append(tool_args["file_path"])

        # Edit tool modifies existing files
        elif tool_name == "Edit":
            if "file_path" in tool_args:
                file_paths.append(tool_args["file_path"])

        # MultiEdit tool can modify multiple files
        elif tool_name == "MultiEdit":
            if "edits" in tool_args:
                for edit in tool_args["edits"]:
                    if "file_path" in edit:
                        file_paths.append(edit["file_path"])

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

        structure_score = analysis.get("structure_score", 0.5)
        domain_accuracy = analysis.get("domain_accuracy", 0.5)
        documentation_level = analysis.get("documentation_level", 0.5)

        # Weighted calculation
        # Structure clarity: 30%
        # Domain accuracy: 40%
        # Documentation level: 30%
        confidence = structure_score * 0.3 + domain_accuracy * 0.4 + documentation_level * 0.3

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
            "spec_id": spec_id,
            "spec_md": spec_md,
            "plan_md": plan_md,
            "acceptance_md": acceptance_md,
        }

    def _generate_spec_id(self, file_path: str) -> str:
        """Generate unique SPEC ID from file path."""
        # Extract meaningful name from file path
        file_name = os.path.basename(file_path)
        name_parts = file_name.split("_")

        # Convert to uppercase and join
        meaningful_name = "".join(part.upper() for part in name_parts if part)

        # Add hash to ensure uniqueness
        file_hash = hashlib.md5(file_path.encode(), usedforsecurity=False).hexdigest()[:4]

        return f"{meaningful_name}-{file_hash}"

    def _generate_spec_content(self, analysis: Dict[str, Any], spec_id: str, file_name: str) -> str:
        """Generate main spec.md content."""
        template = f"""---
  "id": "SPEC-{spec_id}",
  "title": "Auto-generated SPEC for {file_name}",
  "title_en": "Auto-generated SPEC for {file_name}",
  "version": "1.0.0",
  "status": "pending",
  "created": "{time.strftime("%Y-%m-%d")}",
  "author": "@alfred-auto",
  "reviewer": "",
  "category": "FEATURE",
  "priority": "MEDIUM",
  "tags": ["auto-generated", "{spec_id}"],
  "language": "en",
  "estimated_complexity": "auto"
}}
---

## Auto-generated SPEC for {file_name}

### Overview

{analysis.get("description", "This spec was auto-generated based on code analysis.")}

### Environment

- **Project**: MoAI-ADK Auto-generated SPEC
- **Language**: {analysis.get("language", "Python")}
- **File**: {file_name}
- **Generation Method**: Automatic analysis-based
- **Status**: Review required

### Assumptions

1. Code structure is clearly defined
2. Domain-specific terminology is expected to be used
3. Standard development practices are assumed to be followed
4. Generated SPEC will be finalized after user review

### Requirements

#### Ubiquitous Requirements

- **REQ-001**: System must perform the functionality of {file_name}
- **REQ-002**: Generated functionality must be stable
- **REQ-003**: Code must be written in a maintainable form
- **REQ-004**: Tests must satisfy functional requirements
- **REQ-005**: Code must comply with project coding standards

#### State-driven Requirements

{analysis.get("state_requirements", "- **REQ-006**: System must transition from initial state to target state")}

#### Event-driven Requirements

{analysis.get("event_requirements", "- **REQ-007**: System must respond when user input occurs")}

### Specifications

{analysis.get("specifications", "- **SPEC-001**: System must implement requirements")}

### Traceability


### Edit Guide

**User Review Recommendations:**
1. ✅ Verify technical clarity
2. ✅ Specify requirements
3. ✅ Review domain-specific terminology
4. ✅ Define state and event requirements
5. ✅ Detail specifications

**Quality Improvement Suggestions:**
- Add domain-specific terminology
- Specify user cases
- Define performance requirements
- Add security requirements
"""
        return template

    def _generate_plan_content(self, analysis: Dict[str, Any], spec_id: str, file_name: str) -> str:
        """Generate plan.md content."""
        return f"""---
  "id": "PLAN-{spec_id}",
  "spec_id": "SPEC-{spec_id}",
  "title": "Auto-generated Implementation Plan for {file_name}",
  "version": "1.0.0",
  "status": "pending",
  "created": "{time.strftime("%Y-%m-%d")}",
  "author": "@alfred-auto"
}}
---

## Auto-generated Implementation Plan for {file_name}

### Implementation Phases

#### Phase 1: Basic Structure Review (Priority: High)

- [ ] Complete code structure analysis
- [ ] Identify core functionality
- [ ] Verify dependencies
- [ ] Set up test environment

#### Phase 2: Requirements Specification (Priority: Medium)

- [ ] Specify ubiquitous requirements
- [ ] Define state-driven requirements
- [ ] Review event-driven requirements
- [ ] Set performance requirements

#### Phase 3: Implementation Planning (Priority: Medium)

- [ ] Design module architecture
- [ ] Define interfaces
- [ ] Design data structures
- [ ] Plan error handling

#### Phase 4: Test Strategy Development (Priority: High)

- [ ] Plan unit tests
- [ ] Plan integration tests
- [ ] User story-based testing
- [ ] Implement test automation

### Technical Approach

#### Architecture Design

```
{analysis.get("architecture", "User Input → Validation → Business Logic → Data Processing → Output")}
    ↓
[Core Components] → [External Services] → [Data Layer]
```

#### Core Components

1. **{analysis.get("main_component", "Main Class")}**: Handle primary business logic
2. **{analysis.get("service_component", "Service Layer")}**: Integrate external services
3. **{analysis.get("data_component", "Data Layer")}**: Process and store data
4. **{analysis.get("component_4", "Validation Layer")}**: Validate input and check validity

#### Dependency Management

**Utilize Existing Modules:**
- {analysis.get("existing_modules", "Utilize standard libraries")}

**Add New Modules:**
- {analysis.get("new_modules", "Add as needed")}

### Success Criteria

#### Functional Criteria

- ✅ All requirements implemented
- ✅ Test coverage 85% or higher
- ✅ Performance targets met
- ✅ User requirements satisfied

#### Performance Criteria

- ✅ Response time {analysis.get("performance_target", "within 1 second")}
- ✅ Memory usage optimized
- ✅ Parallel processing supported
- ✅ Scalability verified

#### Quality Criteria

- ✅ Code quality verification passed
- ✅ Security scanning passed
- ✅ Documentation completeness verified
- ✅ Maintainability validated

### Next Steps

1. **Immediate**: Basic structure review (1-2 days)
2. **Weekly Goal**: Requirements specification (3-5 days)
3. **2-Week Goal**: Implementation completion (7-14 days)
4. **Release Preparation**: Testing and validation (14-16 days)
"""

    def _generate_acceptance_content(self, analysis: Dict[str, Any], spec_id: str, file_name: str) -> str:
        """Generate acceptance.md content."""
        return f"""---
  "id": "ACCEPT-{spec_id}",
  "spec_id": "SPEC-{spec_id}",
  "title": "Auto-generated Acceptance Criteria for {file_name}",
  "version": "1.0.0",
  "status": "pending",
  "created": "{time.strftime("%Y-%m-%d")}",
  "author": "@alfred-auto"
}}
---

## Auto-generated Acceptance Criteria for {file_name}

### Acceptance Criteria

#### Basic Functionality

**Must-have:**
- [ ] {analysis.get("must_have_1", "System must operate normally")}
- [ ] {analysis.get("must_have_2", "User interface must display correctly")}
- [ ] {analysis.get("must_have_3", "Data processing logic must function properly")}

**Should-have:**
- [ ] {analysis.get("should_have_1", "User experience must be smooth")}
- [ ] {analysis.get("should_have_2", "Performance targets must be met")}

#### Performance Testing

**Performance Requirements:**
- [ ] Response time: {analysis.get("response_time", "within 1 second")}
- [ ] Concurrent users: support {analysis.get("concurrent_users", "100 users")} or more
- [ ] Memory usage: {analysis.get("memory_usage", "100MB or less")}
- [ ] CPU utilization: {analysis.get("cpu_usage", "50% or less")}

**Load Testing:**
- [ ] Functional load testing passed
- [ ] Long-term stability testing passed
- [ ] Recovery testing passed

#### Security Testing

**Security Requirements:**
- [ ] {analysis.get("security_req_1", "Authentication and authorization verification passed")}
- [ ] {analysis.get("security_req_2", "Input validation passed")}
- [ ] {analysis.get("security_req_3", "SQL injection protection passed")}

**Vulnerability Testing:**
- [ ] OWASP Top 10 inspection passed
- [ ] Security scanning passed
- [ ] Permission settings verification passed

#### Compatibility Testing

**Browser Compatibility:**
- [ ] Chrome latest version
- [ ] Firefox latest version
- [ ] Safari latest version
- [ ] Edge latest version

**Device Compatibility:**
- [ ] Desktop (1920x1080)
- [ ] Tablet (768x1024)
- [ ] Mobile (375x667)

#### User Acceptance Testing

**User Scenarios:**
- [ ] {analysis.get("user_scenario_1", "General user scenario testing passed")}
- [ ] {analysis.get("user_scenario_2", "Administrator scenario testing passed")}
- [ ] {analysis.get("user_scenario_3", "Error handling scenario testing passed")}

**User Feedback:**
- [ ] User satisfaction 80% or higher
- [ ] Feature usability evaluation
- [ ] Design and UI/UX verification

### Validation Process

#### Phase 1: Unit Tests

- [ ] Developer testing completed
- [ ] Code review passed
- [ ] Automated testing passed
- [ ] Code coverage 85% or higher

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
- [ ] Final acceptance approval

### Validation Templates

#### Functionality Validation Template

| Function ID | Function Name | Expected Result | Actual Result | Status | Notes |
|-------------|---------------|-----------------|---------------|--------|-------|
| FUNC-001 | Function 1 | Success | Testing | In Progress | Description |
| FUNC-002 | Function 2 | Success | Success | Passed | Description |
| FUNC-003 | Function 3 | Success | Failed | Failed | Description |

#### Performance Validation Template

| Test Item | Target | Measured | Status | Notes |
|-----------|--------|----------|--------|-------|
| Response time | 1s | 0.8s | Passed | Description |
| Memory usage | 100MB | 85MB | Passed | Description |
| CPU utilization | 50% | 45% | Passed | Description |

### Completion Criteria

#### Pass Criteria

- ✅ All required functionality validation passed
- ✅ Performance requirements met
- ✅ Security testing passed
- ✅ User acceptance passed
- ✅ Documentation validation completed

#### Reporting

- [ ] Validation report created
- [ ] Identified issues documented
- [ ] Improvements defined
- [ ] Acceptance approval document prepared

**Validation Team:**
- Developer: @developer
- QA: @qa_engineer
- Product Owner: @product_owner
- Final Approver: @stakeholder
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
            suggestions.append("Improvement needed to fully comply with EARS format.")

        if completeness < 0.8:
            suggestions.append("Requirements and specifications need to be more detailed.")

        if content_quality < 0.7:
            suggestions.append("Domain-specific terminology and technical content need to be added.")

        return {
            "quality_score": min(max(quality_score, 0.0), 1.0),
            "ears_compliance": ears_compliance,
            "completeness": completeness,
            "content_quality": content_quality,
            "suggestions": suggestions,
        }

    def _check_ears_compliance(self, spec_content: Dict[str, str]) -> float:
        """Check EARS format compliance."""
        spec_md = spec_content.get("spec_md", "")

        required_sections = [
            "Overview",
            "Environment",
            "Assumptions",
            "Requirements",
            "Specifications",
        ]

        found_sections = 0
        for section in required_sections:
            if section in spec_md:
                found_sections += 1

        return found_sections / len(required_sections)

    def _check_completeness(self, spec_content: Dict[str, str]) -> float:
        """Check content completeness."""
        spec_md = spec_content.get("spec_md", "")
        plan_md = spec_content.get("plan_md", "")
        acceptance_md = spec_content.get("acceptance_md", "")

        # Check minimum content length
        total_length = len(spec_md) + len(plan_md) + len(acceptance_md)
        length_score = min(total_length / 2000, 1.0)  # 2000 chars as baseline

        # Check for content diversity
        has_requirements = "Requirements" in spec_md
        has_planning = "Implementation Plan" in plan_md
        has_acceptance = "Acceptance" in acceptance_md

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
        spec_md = spec_content.get("spec_md", "")

        # Check for technical terms
        technical_indicators = [
            "API",
            "data",
            "interface",
            "module",
            "component",
            "architecture",
        ]
        technical_score = sum(1 for term in technical_indicators if term in spec_md) / len(technical_indicators)

        # Check for specificity
        has_requirements = re.search(r"REQ-\d+", spec_md)
        has_specifications = re.search(r"SPEC-\d+", spec_md)

        specificity_score = 0.0
        if has_requirements:
            specificity_score += 0.5
        if has_specifications:
            specificity_score += 0.5

        return (technical_score + specificity_score) / 2

    def create_spec_files(self, spec_id: str, content: Dict[str, str], base_dir: str = ".moai/specs") -> bool:
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
                ("spec.md", content.get("spec_md", "")),
                ("plan.md", content.get("plan_md", "")),
                ("acceptance.md", content.get("acceptance_md", "")),
            ]

            for filename, content_text in files_to_create:
                file_path = os.path.join(spec_dir, filename)
                with open(file_path, "w", encoding="utf-8") as f:
                    f.write(content_text)

                logger.info(f"Created spec file: {file_path}")

            return True

        except Exception as e:
            logger.error(f"Failed to create spec files: {e}")
            return False

    def execute(self, tool_name: str, tool_args: Dict[str, Any], result: Any = None) -> Dict[str, Any]:
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
                    "success": False,
                    "message": "Auto-spec completion not triggered",
                    "execution_time": time.time() - start_time,
                }

            # Detect code changes
            changed_files = self.detect_code_changes(tool_name, tool_args, result)

            if not changed_files:
                return {
                    "success": False,
                    "message": "No code changes detected",
                    "execution_time": time.time() - start_time,
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
                    min_confidence = self.auto_config.get("min_confidence", 0.7)
                    if confidence < min_confidence:
                        logger.info(f"Confidence {confidence} below threshold {min_confidence}")
                        continue

                    # Generate complete spec
                    spec_content = self.generate_complete_spec(analysis, file_path)

                    # Validate quality
                    validation = self.validate_generated_spec(spec_content)

                    # Create spec files
                    spec_id = spec_content["spec_id"]
                    created = self.create_spec_files(spec_id, spec_content)

                    results.append(
                        {
                            "file_path": file_path,
                            "spec_id": spec_id,
                            "confidence": confidence,
                            "quality_score": validation["quality_score"],
                            "created": created,
                        }
                    )

                    logger.info(f"Auto-generated SPEC for {file_path}: {spec_id}")

                except Exception as e:
                    logger.error(f"Error processing {file_path}: {e}")
                    results.append({"file_path": file_path, "error": str(e)})

            # Generate summary
            successful_creations = [r for r in results if r.get("created", False)]
            failed_creations = [r for r in results if not r.get("created", False)]

            execution_result = {
                "success": len(successful_creations) > 0,
                "generated_specs": successful_creations,
                "failed_files": failed_creations,
                "execution_time": time.time() - start_time,
            }

            # Add notification message
            if successful_creations:
                execution_result["message"] = f"Auto-generated {len(successful_creations)} SPEC(s)"
            elif failed_creations:
                execution_result["message"] = "Auto-spec completion attempted but no specs created"
            else:
                execution_result["message"] = "No files required auto-spec completion"

            return execution_result

        except Exception as e:
            logger.error(f"Error in auto-spec completion: {e}")
            return {
                "success": False,
                "error": str(e),
                "execution_time": time.time() - start_time,
            }
