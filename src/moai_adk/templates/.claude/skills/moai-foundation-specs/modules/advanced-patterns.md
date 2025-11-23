# Advanced SPEC Patterns

**Version**: 4.0.0
**Focus**: Version management, template systems, automated validation

---

## Advanced SPEC Patterns

### Pattern 1: Version Management & Backwards Compatibility

**Semantic Versioning for SPEC Documents**:
```yaml
# SPEC-001/spec.md - Version tracking
---
spec_id: SPEC-001
title: User Authentication System
version: 2.1.3  # Major.Minor.Patch
previous_versions:
  - version: 2.1.2
    date: 2025-11-15
    changes: "Fixed OAuth token expiry handling"
    status: deprecated

compatibility:
  min_sdk_version: 3.12.0
  breaking_changes:
    - version: 2.0.0
      description: "Changed JWT signature algorithm from HS256 to RS256"
      migration_guide: "link to migration doc"

deprecation:
  scheduled_removal: 2025-12-31
  successor_spec: SPEC-015
---

## Requirements (v2.1.3)

### SPEC-001-REQ-01: OAuth2 Integration
**Version**: 1.3.0 (added in v2.0.0)
**Status**: Active
**Deprecated since**: None (use SPEC-015 for newer patterns)

### SPEC-001-REQ-02: JWT Token Validation
**Version**: 2.1.0 (changed in v2.1.0)
**Status**: Active
**Breaking change**: RS256 required (v2.0.0+)
```

### Pattern 2: SPEC Template System

**Automated SPEC Generation**:
```python
# scripts/spec_generator.py
import jinja2
from pathlib import Path
from dataclasses import dataclass

@dataclass
class SpecTemplate:
    name: str
    category: str
    description: str

class SpecTemplateEngine:
    """Generate SPEC documents from templates"""

    TEMPLATES = {
        'rest_api': '''---
spec_id: {{ spec_id }}
title: {{ title }}
category: REST API
---

## Requirements

### {{ spec_id }}-REQ-01: API Endpoints
**Pattern**: Ubiquitous
**Statement**: The system SHALL provide REST endpoints at /api/v1/{{ resource }}

### {{ spec_id }}-REQ-02: Authentication
**Pattern**: Event-driven
**Statement**: WHEN a request includes Authorization header, the system SHALL validate JWT token

### {{ spec_id }}-REQ-03: Error Handling
**Pattern**: Unwanted
**Statement**: IF the request is invalid, THEN the system SHALL return 400 with error details
''',
        'auth_system': '''---
spec_id: {{ spec_id }}
title: {{ title }}
category: Authentication
---

## Requirements

### {{ spec_id }}-REQ-01: User Registration
**Statement**: The system SHALL register users with email and password

### {{ spec_id }}-REQ-02: Multi-Factor Authentication
**Statement**: The system SHALL support TOTP-based 2FA

### {{ spec_id }}-REQ-03: Session Management
**Statement**: The system SHALL invalidate sessions after 1 hour of inactivity
'''
    }

    def generate_spec(self, template_name: str, **context) -> str:
        """Generate SPEC from template"""
        template_text = self.TEMPLATES.get(template_name)
        if not template_text:
            raise ValueError(f"Template {template_name} not found")

        template = jinja2.Template(template_text)
        return template.render(**context)

# Usage
engine = SpecTemplateEngine()
spec_content = engine.generate_spec(
    'rest_api',
    spec_id='SPEC-042',
    title='Product Catalog API',
    resource='products'
)
```

### Pattern 3: Automated SPEC Validation

**Multi-Layer Validation System**:
```python
class SpecValidator:
    """Comprehensive SPEC document validation"""

    VALIDATION_RULES = {
        'structure': [
            'Must have spec_id',
            'Must have title',
            'Must have requirements section',
            'Version must follow semantic versioning'
        ],
        'requirements': [
            'Each requirement must have unique ID (SPEC-001-REQ-XX)',
            'Each requirement must follow EARS pattern',
            'Each requirement must have acceptance criteria',
            'Complexity must be documented'
        ],
        'quality': [
            'No circular dependencies between SPECs',
            'Traceability to user stories',
            'Risk assessment for each requirement',
            'Effort estimation'
        ]
    }

    async def validate_spec(self, spec_file: Path) -> ValidationResult:
        """Validate SPEC against all rules"""
        with open(spec_file) as f:
            spec = yaml.safe_load(f.read())

        errors = []
        warnings = []

        # Structure validation
        for rule in self.VALIDATION_RULES['structure']:
            if not self._check_structure(spec, rule):
                errors.append(f"Structure: {rule}")

        # Requirements validation
        for req in spec.get('requirements', []):
            for rule in self.VALIDATION_RULES['requirements']:
                if not self._validate_requirement(req, rule):
                    errors.append(f"Requirement: {rule} in {req.get('id')}")

        # Quality checks
        quality_checks = self._check_quality(spec)
        warnings.extend(quality_checks)

        return ValidationResult(
            valid=len(errors) == 0,
            errors=errors,
            warnings=warnings,
            score=self._calculate_quality_score(spec)
        )
```

### Pattern 4: SPEC Dependency Tracking

**Requirements Traceability Matrix (RTM)**:
```yaml
# .moai/specs/RTM.yaml - Traceability Matrix
dependencies:
  SPEC-001:  # User Authentication
    depends_on: []
    blocks: [SPEC-002, SPEC-003]
    related_to: [SPEC-015, SPEC-018]

  SPEC-002:  # OAuth2 Integration
    depends_on: [SPEC-001]
    blocks: [SPEC-010]
    related_to: [SPEC-003, SPEC-004]

  SPEC-003:  # Session Management
    depends_on: [SPEC-001]
    blocks: []
    related_to: [SPEC-002, SPEC-010]

# Automated dependency checking
verification:
  - name: No Circular Dependencies
    rule: "No SPEC can depend on specs that depend on it"

  - name: Dependency Order
    rule: "Implement SPECs in dependency order"

  - name: Completeness
    rule: "All blocking SPECs must be completed first"
```

### Pattern 5: SPEC Review Workflow Automation

**GitHub Workflow for SPEC Review**:
```yaml
# .github/workflows/spec-review.yml
name: SPEC Review Process

on:
  pull_request:
    paths:
      - '.moai/specs/SPEC-*.yaml'
  push:
    branches: [main]
    paths:
      - '.moai/specs/SPEC-*.yaml'

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Validate SPEC Format
        run: |
          python scripts/validate_spec.py .moai/specs/

      - name: Check Traceability
        run: |
          python scripts/check_traceability.py .moai/specs/

      - name: Generate RTM
        run: |
          python scripts/generate_rtm.py .moai/specs/

      - name: Comment RTM on PR
        if: github.event_name == 'pull_request'
        uses: actions/github-script@v6
        with:
          script: |
            const fs = require('fs');
            const rtm = fs.readFileSync('.moai/specs/RTM.md', 'utf8');
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: rtm
            });
```

### Pattern 6: SPEC Search and Navigation

**Full-Text Search and Indexing**:
```python
from elasticsearch import Elasticsearch
from pathlib import Path

class SpecSearchEngine:
    """Search SPEC documents efficiently"""

    def __init__(self):
        self.es = Elasticsearch([{'host': 'localhost', 'port': 9200}])
        self.index_name = 'moai-specs'

    def index_specs(self, specs_dir: Path):
        """Index all SPEC files"""
        for spec_file in specs_dir.glob('SPEC-*/spec.md'):
            with open(spec_file) as f:
                content = f.read()

            self.es.index(
                index=self.index_name,
                id=spec_file.parent.name,
                body={
                    'spec_id': spec_file.parent.name,
                    'title': self._extract_title(content),
                    'requirements': self._extract_requirements(content),
                    'content': content,
                    'indexed_at': datetime.now().isoformat()
                }
            )

    def search(self, query: str, filters: dict = None) -> list:
        """Search SPEC documents"""
        search_body = {
            'query': {
                'multi_match': {
                    'query': query,
                    'fields': ['title^2', 'requirements', 'content']
                }
            }
        }

        if filters:
            # Add filters for category, status, etc.
            search_body['query'] = {
                'bool': {
                    'must': search_body['query'],
                    'filter': [{'term': filters}]
                }
            }

        results = self.es.search(index=self.index_name, body=search_body)
        return [hit['_source'] for hit in results['hits']['hits']]
```

---

**Last Updated**: 2025-11-22
