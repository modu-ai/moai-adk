"""
PM Plugin Commands - /init-pm command implementation

@CODE:PM-INIT-CMD-001:COMMANDS
"""

from dataclasses import dataclass
from pathlib import Path
from typing import Optional, List
import json
import yaml
from datetime import datetime


# @CODE:PM-COMMAND-RESULT-001:RESULT
@dataclass
class CommandResult:
    """Result object for command execution"""
    success: bool
    spec_dir: Path
    files_created: List[str]
    message: str
    error: Optional[str] = None


class InitPMCommand:
    """
    /init-pm command implementation

    Generates EARS SPEC templates for project management
    """

    # Validation constants
    MIN_PROJECT_NAME_LENGTH = 3
    MAX_PROJECT_NAME_LENGTH = 50
    VALID_TEMPLATES = ["moai-spec", "enterprise", "agile"]
    VALID_RISK_LEVELS = ["low", "medium", "high"]

    def __init__(self):
        """Initialize PM Plugin command"""
        self.templates_dir = Path(__file__).parent / "templates"

    def validate_project_name(self, project_name: str) -> bool:
        """
        Validate project name format

        @CODE:PM-VALIDATE-NAME-001:VALIDATION

        Args:
            project_name: Project name to validate

        Returns:
            True if valid

        Raises:
            ValueError: If project name is invalid
        """
        if not project_name:
            raise ValueError("Project name cannot be empty")

        if len(project_name) < self.MIN_PROJECT_NAME_LENGTH:
            raise ValueError(
                f"Project name must be at least {self.MIN_PROJECT_NAME_LENGTH} characters"
            )

        if len(project_name) > self.MAX_PROJECT_NAME_LENGTH:
            raise ValueError(
                f"Project name cannot exceed {self.MAX_PROJECT_NAME_LENGTH} characters"
            )

        # Check for lowercase letters, numbers, hyphens only
        if not all(c.islower() or c.isdigit() or c == "-" for c in project_name):
            raise ValueError(
                "Project name must contain only lowercase letters, numbers, and hyphens"
            )

        # Cannot start or end with hyphen
        if project_name.startswith("-") or project_name.endswith("-"):
            raise ValueError("Project name cannot start or end with a hyphen")

        # Check for consecutive hyphens
        if "--" in project_name:
            raise ValueError("Project name cannot contain consecutive hyphens")

        return True

    def validate_template(self, template: str) -> bool:
        """
        Validate template name

        @CODE:PM-VALIDATE-TEMPLATE-001:VALIDATION
        """
        if template not in self.VALID_TEMPLATES:
            raise ValueError(
                f"Invalid template: {template}\nSupported templates: {', '.join(self.VALID_TEMPLATES)}"
            )
        return True

    def validate_risk_level(self, risk_level: str) -> bool:
        """
        Validate risk level

        @CODE:PM-VALIDATE-RISK-001:VALIDATION
        """
        if risk_level not in self.VALID_RISK_LEVELS:
            raise ValueError(
                f"Invalid risk level: {risk_level}\nSupported levels: {', '.join(self.VALID_RISK_LEVELS)}"
            )
        return True

    def generate_spec_id(self, project_name: str) -> str:
        """
        Generate SPEC ID from project name

        @CODE:PM-SPEC-ID-001:SPEC

        Format: SPEC-{PROJECT}-001
        """
        spec_name = project_name.upper()
        return f"SPEC-{spec_name}-001"

    def create_spec_directory(self, output_dir: Path, spec_id: str) -> Path:
        """
        Create SPEC directory structure

        @CODE:PM-SPEC-DIR-001:DIRECTORY
        """
        spec_dir = output_dir / ".moai" / "specs" / spec_id

        if spec_dir.exists():
            raise FileExistsError(f"SPEC already exists: {spec_dir}")

        spec_dir.mkdir(parents=True, exist_ok=False)
        return spec_dir

    def create_spec_file(
        self,
        spec_dir: Path,
        project_name: str,
        spec_id: str,
        template: str
    ) -> Path:
        """
        Create spec.md with EARS format

        @CODE:PM-SPEC-FILE-001:SPEC
        """
        # @CODE:PM-YAML-FRONTMATTER-CREATION-001:YAML
        frontmatter = {
            "spec_id": spec_id,
            "title": f"{project_name.replace('-', ' ').title()} - Project Specification",
            "version": "1.0.0-dev",
            "status": "In Development",
            "owner": "GOOS🪿",
            "created": datetime.now().strftime("%Y-%m-%d"),
            "tags": ["spec", "project", template],
            "language": "en"
        }

        # Create EARS content
        ears_content = f"""
## 🎯 EARS Requirements

### Ubiquitous Behaviors (Core Features)

**Feature 1: Project Setup**
- GIVEN the project is initialized
- WHEN project charter is created
- THEN stakeholder roles are defined

### Event-Driven Behaviors

**Event 1: Project Milestone**
- WHEN project milestone is reached
- THEN status is updated

### State-Driven Behaviors

**State 1: Project Status**
- GIVEN the project is in planning phase
- WHEN all requirements are documented
- THEN project can move to design phase

### Optional Behaviors

**Optional 1: Risk Assessment**
- GIVEN risk assessment is enabled
- WHEN risks are identified
- THEN mitigation plan is created (optional)

### Unwanted Behaviors

**Unwanted 1: Unauthorized Access**
- GIVEN a user without permissions
- WHEN attempting to modify project
- THEN access is denied with error message
"""

        spec_file = spec_dir / "spec.md"

        with open(spec_file, "w") as f:
            f.write("---\n")
            f.write(yaml.dump(frontmatter, default_flow_style=False, sort_keys=False))
            f.write("---\n")
            f.write(ears_content)

        return spec_file

    def create_plan_file(self, spec_dir: Path, project_name: str, spec_id: str) -> Path:
        """
        Create plan.md with implementation plan

        @CODE:PM-PLAN-FILE-001:PLAN
        """
        plan_content = f"""---
spec_id: {spec_id}
title: {project_name.replace('-', ' ').title()} - Implementation Plan
version: 1.0.0-dev
status: In Development
created: {datetime.now().strftime("%Y-%m-%d")}
---

# Implementation Plan

## Phase 1: Kickoff (Week 1)

### Activities
1. Stakeholder alignment meeting
2. Project charter finalization
3. Resource allocation
4. Risk assessment

### Deliverables
- [ ] Project charter signed
- [ ] Team roster confirmed
- [ ] Risk register created

---

## Phase 2: Design (Week 2-3)

### Activities
1. Architecture design
2. Technology selection
3. API contract definition

### Deliverables
- [ ] Architecture diagram
- [ ] Technology stack approved
- [ ] API specification

---

## Phase 3: Implementation (Week 4-8)

### Activities
1. Development sprint planning
2. Code implementation
3. Unit testing
4. Integration testing

### Deliverables
- [ ] Feature development complete
- [ ] Test coverage ≥85%
- [ ] Build pipeline established

---

## Phase 4: Validation (Week 9-10)

### Activities
1. UAT execution
2. Bug fixes
3. Performance testing

### Deliverables
- [ ] UAT complete
- [ ] All critical bugs fixed
- [ ] Performance acceptable

---

## Phase 5: Release (Week 11-12)

### Activities
1. Release preparation
2. Deployment
3. Post-launch monitoring

### Deliverables
- [ ] Production deployment
- [ ] Documentation complete
- [ ] Support plan ready

---

## Timeline

| Phase | Duration | Start | End |
|-------|----------|-------|-----|
| Kickoff | 1 week | - | - |
| Design | 2 weeks | - | - |
| Implementation | 5 weeks | - | - |
| Validation | 2 weeks | - | - |
| Release | 2 weeks | - | - |
| **Total** | **12 weeks** | - | - |

---

## Resources

### Team
- Project Manager: TBD
- Architect: TBD
- Development Team: TBD
- QA Lead: TBD

### Budget
- Development: TBD
- Infrastructure: TBD
- Testing: TBD

---

## Risks

See risk-matrix.json for detailed risk assessment.

---

**Plan Author**: Alfred PM Agent
**Last Updated**: {datetime.now().strftime("%Y-%m-%d")}
"""

        plan_file = spec_dir / "plan.md"
        plan_file.write_text(plan_content)
        return plan_file

    def create_acceptance_file(self, spec_dir: Path, spec_id: str) -> Path:
        """
        Create acceptance.md with acceptance criteria

        @CODE:PM-ACCEPTANCE-FILE-001:ACCEPTANCE
        """
        acceptance_content = f"""---
spec_id: {spec_id}
title: Acceptance Criteria
version: 1.0.0-dev
created: {datetime.now().strftime("%Y-%m-%d")}
---

# Acceptance Criteria

## ✅ Project Completion Criteria

### Functional Requirements
- [ ] All features implemented per SPEC
- [ ] All acceptance criteria met
- [ ] No critical bugs remaining
- [ ] Performance meets SLA

### Quality Requirements
- [ ] Test coverage ≥85%
- [ ] Code review approved
- [ ] Linting: 0 errors
- [ ] Type checking: 100% (mypy strict)

### Documentation Requirements
- [ ] README.md complete
- [ ] API documentation complete
- [ ] Architecture documentation complete
- [ ] Deployment guide complete

### Deployment Requirements
- [ ] Build pipeline passes
- [ ] Deployment checklist complete
- [ ] Rollback plan documented
- [ ] Monitoring configured

---

## 📊 Quality Metrics

| Metric | Target | Status |
|--------|--------|--------|
| Test Coverage | ≥85% | ⏳ |
| Linting | 0 errors | ⏳ |
| Type Safety | 100% | ⏳ |
| Code Review | Approved | ⏳ |
| Deployment | Verified | ⏳ |

---

## 🔍 Sign-Off

| Role | Name | Date | Signature |
|------|------|------|-----------|
| Project Manager | - | - | - |
| Tech Lead | - | - | - |
| QA Lead | - | - | - |
| Product Owner | - | - | - |

---

**Created**: {datetime.now().strftime("%Y-%m-%d")}
"""

        acceptance_file = spec_dir / "acceptance.md"
        acceptance_file.write_text(acceptance_content)
        return acceptance_file

    def create_charter_file(self, spec_dir: Path, project_name: str, spec_id: str) -> Path:
        """
        Create charter.md with project governance

        @CODE:PM-CHARTER-FILE-001:CHARTER
        """
        charter_content = f"""---
spec_id: {spec_id}
title: Project Charter
version: 1.0.0-dev
created: {datetime.now().strftime("%Y-%m-%d")}
---

# Project Charter

## 📋 Project Overview

### Project Name
{project_name.replace('-', ' ').title()}

### Project ID
{spec_id}

### Project Manager
TBD

### Sponsor
TBD

---

## 🎯 Business Case

### Objective
Define the primary business objective for this project.

### Expected Benefits
- Benefit 1
- Benefit 2
- Benefit 3

### Success Criteria
- [ ] Measurable criteria 1
- [ ] Measurable criteria 2
- [ ] Measurable criteria 3

---

## 👥 Stakeholders

### Stakeholder Matrix

| Stakeholder | Role | Responsibility | Contact |
|-----------|------|-----------------|---------|
| Sponsor | Executive Sponsor | Strategic direction | - |
| PM | Project Manager | Execution | - |
| Tech Lead | Technical Lead | Architecture | - |
| QA Lead | Quality Lead | Testing | - |

---

## 📊 Budget & Schedule

### Timeline
- **Start Date**: TBD
- **End Date**: TBD
- **Duration**: TBD

### Budget
- **Total Budget**: TBD
- **Contingency**: TBD
- **Reserve**: TBD

---

## ⚠️ Risk Management

See risk-matrix.json for detailed risk assessment.

---

## 🔐 Governance

### Approval Chain
1. Project Manager reviews
2. Sponsor approves
3. Stakeholders confirm

### Decision Authority
- Strategic decisions: Sponsor
- Technical decisions: Tech Lead
- Budget decisions: Sponsor
- Schedule decisions: Project Manager

---

**Charter Created**: {datetime.now().strftime("%Y-%m-%d")}
"""

        charter_file = spec_dir / "charter.md"
        charter_file.write_text(charter_content)
        return charter_file

    def create_risk_matrix(self, spec_dir: Path, risk_level: str) -> Path:
        """
        Create risk-matrix.json with risk assessment

        @CODE:PM-RISK-CREATION-001:RISK
        """
        # Risk levels based on specified level
        risk_counts = {
            "low": 3,
            "medium": 6,
            "high": 10
        }

        risks = []
        risk_count = risk_counts.get(risk_level, 6)

        for i in range(1, risk_count + 1):
            risks.append({
                "id": f"RISK-{i:03d}",
                "description": f"Risk {i}: Technical/Process/Resource risk",
                "category": ["Technical", "Process", "Resource"][i % 3],
                "probability": ["Low", "Medium", "High"][i % 3],
                "impact": ["Low", "Medium", "High"][(i + 1) % 3],
                "mitigation": f"Mitigation strategy for risk {i}",
                "owner": "TBD",
                "status": "Identified"
            })

        risk_data = {
            "spec_id": spec_dir.name,
            "created": datetime.now().isoformat(),
            "risk_level": risk_level,
            "total_risks": len(risks),
            "risks": risks
        }

        risk_file = spec_dir / "risk-matrix.json"
        with open(risk_file, "w") as f:
            json.dump(risk_data, f, indent=2)

        return risk_file

    def execute(
        self,
        project_name: str,
        output_dir: Path,
        template: str = "moai-spec",
        risk_level: str = "medium",
        skip_charter: bool = False
    ) -> CommandResult:
        """
        Execute /init-pm command

        @CODE:PM-EXECUTE-001:MAIN

        Args:
            project_name: Project name (lowercase, hyphens)
            output_dir: Output directory for SPEC files
            template: Template to use (moai-spec, enterprise, agile)
            risk_level: Risk level (low, medium, high)
            skip_charter: Skip charter.md creation

        Returns:
            CommandResult with success/failure status
        """
        # Validation (may raise exceptions)
        self.validate_project_name(project_name)
        self.validate_template(template)
        self.validate_risk_level(risk_level)

        try:
            # Generate SPEC ID
            spec_id = self.generate_spec_id(project_name)

            # Create SPEC directory
            spec_dir = self.create_spec_directory(Path(output_dir), spec_id)

            # Create files
            files_created = []

            # spec.md
            spec_file = self.create_spec_file(spec_dir, project_name, spec_id, template)
            files_created.append(str(spec_file.relative_to(output_dir)))

            # plan.md
            plan_file = self.create_plan_file(spec_dir, project_name, spec_id)
            files_created.append(str(plan_file.relative_to(output_dir)))

            # acceptance.md
            acceptance_file = self.create_acceptance_file(spec_dir, spec_id)
            files_created.append(str(acceptance_file.relative_to(output_dir)))

            # charter.md (unless skipped)
            if not skip_charter:
                charter_file = self.create_charter_file(spec_dir, project_name, spec_id)
                files_created.append(str(charter_file.relative_to(output_dir)))

            # risk-matrix.json
            risk_file = self.create_risk_matrix(spec_dir, risk_level)
            files_created.append(str(risk_file.relative_to(output_dir)))

            message = f"✅ Project '{project_name}' initialized successfully\n"
            message += f"📁 Location: {spec_dir}\n"
            message += f"📊 Risk Level: {risk_level}\n"
            message += f"📋 Template: {template}\n"
            message += f"📝 Files created: {len(files_created)}"

            return CommandResult(
                success=True,
                spec_dir=spec_dir,
                files_created=files_created,
                message=message
            )

        except FileExistsError:
            # Re-raise specific validation errors for caller to handle
            raise
        except Exception as e:
            return CommandResult(
                success=False,
                spec_dir=None,
                files_created=[],
                message=f"❌ Error initializing project",
                error=str(e)
            )


# @CODE:PM-MODULE-INIT-001:MODULE
# Create module-level command instance
init_pm = InitPMCommand()
