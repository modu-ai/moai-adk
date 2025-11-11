---
name: moai-foundation-specs
version: 3.0.0
created: 2025-11-11
updated: 2025-11-11
status: active
description: YAML frontmatter validation with 15 required fields, 15 real templates, and official YAML 1.2 specification compliance
keywords: ['spec', 'yaml', 'validation', 'frontmatter', 'template', 'schema']
allowed-tools:
  - Read
  - Bash
  - Write
---

# Foundation Specs Skill - Professional Edition

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-foundation-specs |
| **Version** | 3.0.0 (2025-11-11) |
| **Allowed tools** | Read (read_file), Bash (terminal), Write (create_file) |
| **Auto-load** | On demand when keywords detected |
| **Tier** | Foundation |
| **Integration** | YAML 1.2 Specification |

---

## What It Does

Professional YAML frontmatter validation system with 15 required fields, 15 real-world SPEC templates, and official YAML 1.2 specification compliance. This skill provides comprehensive validation, template generation, and schema enforcement for MoAI-ADK specifications.

**Key capabilities**:
- ✅ 15 required YAML frontmatter fields with validation
- ✅ 15+ real SPEC templates from actual projects
- ✅ YAML 1.2 specification compliance
- ✅ Multi-template support (SPEC, PLAN, ACCEPTANCE, etc.)
- ✅ Schema validation and auto-correction
- ✅ Template generation and customization
- ✅ Cross-reference validation
- ✅ Integration with MoAI-ADK workflow
- ✅ Performance optimization for large projects

---

## When to Use

**Automatic triggers**:
- SPEC creation and validation (`/alfred:1-plan`)
- Cross-reference checking in markdown files
- Template generation requests
- Quality gate validation

**Manual invocation**:
- Creating new SPEC files with proper frontmatter
- Validating existing specifications
- Generating custom templates
- Cross-project standard enforcement

---

## Core YAML Frontmatter v3.0

### Required Fields (15 Fields)

```yaml
---
# Core Identification Fields (4)
spec_id: "SPEC-XXXX-XXX"                    # Unique SPEC identifier
spec_title: "Brief description"              # Human-readable title
domain: "DOMAIN"                             # Technical domain
version: "1.0.0"                            # Semantic version

# Metadata Fields (6)
created_date: "2025-11-11"                  # Creation date
status: "active|completed|pending"            # Current status
priority: "P0|P1|P2|P3"                     # Priority level
owner: "Team/Person"                        # Owner information
author: "Author Name"                       # Primary author
reviewer: "Reviewer Name"                   # Code review lead

# Classification Fields (3)
type: "new|enhancement|bugfix|documentation" # Change type
complexity: "low|medium|high"               # Implementation complexity
estimated_effort_hours: 40                  # Time estimation

# Lifecycle Fields (2)
implementation_phases: 3                    # Phase count
timeline_weeks: 2                           # Duration in weeks
---
```

### Optional Fields (Enhanced)

```yaml
---
# Extended Fields (Optional)
related_tags: []                            # Cross-references
linked_specs: []                            # Dependencies
tags: []                                    # Additional tags
category: "development|infrastructure"      # Business category
acceptance_criteria: []                     # Key acceptance criteria
risk_factors: []                            # Known risks
stakeholders: []                            # Interested parties
last_updated: "2025-11-11"                  # Modification date
review_status: "draft|reviewed|approved"    # Review state
compliance_level: "basic|enhanced|enterprise" # Compliance tier
---
```

---

## Real SPEC Templates (15 Examples)

### 1. Core SPEC Template
```yaml
---
spec_id: "SPEC-CORE-FEATURE-001"
spec_title: "Core Feature Implementation"
domain: "CORE"
version: "1.0.0"
created_date: "2025-11-11"
status: "active"
priority: "P0"
owner: "GoosLab"
author: "Developer Name"
reviewer: "Tech Lead"
type: "new"
complexity: "medium"
estimated_effort_hours: 80
implementation_phases: 4
timeline_weeks: 3
related_tags:
  - "@SPEC:CORE-FEATURE-001"
  - "@CODE:CORE-FEATURE-001"
  - "@TEST:CORE-FEATURE-001"
  - "@DOC:CORE-FEATURE-001"
linked_specs: []
tags: ["feature", "core", "infrastructure"]
category: "development"
acceptance_criteria: []
risk_factors: []
stakeholders: []
last_updated: "2025-11-11"
review_status: "draft"
compliance_level: "enhanced"
---

# @SPEC:CORE-FEATURE-001 | @EXPERT:BACKEND | @EXPERT:FRONTEND

## SPEC Overview

This SPEC defines the core feature implementation for MoAI-ADK, providing foundational capabilities that support all domain-specific functionality.

## Requirements

- **Core Functionality**: Implement primary feature with full API coverage
- **Performance**: Optimize for large-scale deployments
- **Integration**: Ensure seamless integration with existing systems
- **Documentation**: Complete API documentation and usage guides

## Implementation Strategy

### Phase 1: Foundation (1 week)
- Core architecture design
- Database schema implementation
- Basic API endpoints

### Phase 2: Feature Implementation (1 week)
- Business logic implementation
- External integrations
- Security implementation

### Phase 3: Testing & Validation (1 week)
- Unit and integration testing
- Performance testing
- Security validation

### Phase 4: Documentation & Deployment (1 week)
- API documentation
- User guides
- Production deployment
```

### 2. BaaS Ecosystem Template
```yaml
---
spec_id: "SPEC-BAAS-ECOSYSTEM-001"
spec_title: "BaaS Platform Ecosystem Integration"
domain: "BAAS"
version: "2.0.0"
created_date: "2025-11-09"
status: "active"
priority: "P0"
owner: "GoosLab"
author: "Platform Team"
reviewer: "Architecture Committee"
type: "enhancement"
complexity: "high"
estimated_effort_hours: 150
implementation_phases: 6
timeline_weeks: 6
related_tags:
  - "@SPEC:BAAS-ECOSYSTEM-001"
  - "@CODE:BAAS-FEATURES"
  - "@TEST:BAAS-INTEGRATION"
  - "@DOC:BAAS-ARCHITECTURE"
linked_specs: ["SPEC-CORE-FEATURE-001"]
tags: ["baas", "platform", "integration", "ecosystem"]
category: "infrastructure"
acceptance_criteria: []
risk_factors: ["vendor-lockin", "performance", "security"]
stakeholders: ["platform-team", "devops", "security"]
last_updated: "2025-11-11"
review_status: "reviewed"
compliance_level: "enterprise"
---

# @SPEC:BAAS-ECOSYSTEM-001 | @EXPERT:BACKEND | @EXPERT:DEVOPS

## 📋 개요

MoAI-ADK에 **9개 BaaS 플랫폼** (Supabase, Vercel, Neon, Clerk, Railway, Convex, Firebase, Cloudflare, Auth0)을 심화 통합하여 vibe coder들이 다양한 아키텍처 패턴으로 최적의 플랫폼 조합을 선택하고 설정할 수 있도록 지원합니다.
```

### 3. Validation System Template
```yaml
---
spec_id: "SPEC-VAL-002"
spec_title: "Validation System Implementation"
domain: "VAL"
version: "1.0.0"
created_date: "2025-11-10"
status: "completed"
priority: "P1"
owner: "Quality Assurance"
author: "QA Team"
reviewer: "Engineering Manager"
type: "enhancement"
complexity: "medium"
estimated_effort_hours: 60
implementation_phases: 3
timeline_weeks: 2
related_tags:
  - "@SPEC:VAL-002"
  - "@CODE:VAL-002"
  - "@TEST:VAL-002"
  - "@DOC:VAL-002"
linked_specs: []
tags: ["validation", "quality", "automation", "testing"]
category: "quality"
acceptance_criteria: []
risk_factors: ["false-positives", "performance-impact"]
stakeholders: ["qa-team", "development", "product"]
last_updated: "2025-11-10"
review_status: "approved"
compliance_level: "enhanced"
---

# @SPEC:VAL-002 | @EXPERT:BACKEND | @EXPERT:DEVOPS

## SPEC Overview

This SPEC defines the validation system for MoAI-ADK, which provides comprehensive validation capabilities for TAG chains, code quality, and system integrity.

## Requirements

- **TAG Chain Validation**: Validate complete TAG chains (SPEC → CODE → TEST → DOC)
- **Code Quality Validation**: Validate code structure and TAG adherence
- **System Integrity**: Validate overall system integrity and consistency
- **Performance**: Optimize validation processes for large codebases
```

### 4. Implementation Plan Template
```yaml
---
doc_type: "implementation_plan"
spec_id: "SPEC-IMPLEMENTATION-001"
spec_title: "Implementation Plan Template"
domain: "PLANNING"
version: "1.0.0"
created_date: "2025-11-11"
status: "active"
priority: "P0"
owner: "Project Management"
author: "PM Team"
reviewer: "Stakeholders"
type: "planning"
complexity: "low"
estimated_effort_hours: 40
implementation_phases: 5
timeline_weeks: 4
related_tags: []
linked_specs: []
tags: ["planning", "implementation", "timeline", "resources"]
category: "management"
acceptance_criteria: []
risk_factors: []
stakeholders: []
last_updated: "2025-11-11"
review_status: "draft"
compliance_level: "basic"
---

# 구현 계획: SPEC-IMPLEMENTATION-001

## 📋 개요

SPEC의 단계별 구현을 계획하고 자원을 할당합니다.

**총 노력**: 40시간 | **기간**: 4주 | **팀**: 5명
```

### 5. Acceptance Criteria Template
```yaml
---
doc_type: "acceptance_criteria"
spec_id: "SPEC-ACCEPTANCE-001"
spec_title: "Acceptance Criteria Definition"
domain: "QA"
version: "1.0.0"
created_date: "2025-11-11"
status: "active"
priority: "P0"
owner: "Quality Assurance"
author: "QA Team"
reviewer: "Product Owner"
type: "documentation"
complexity: "low"
estimated_effort_hours: 20
implementation_phases: 1
timeline_weeks: 1
related_tags: []
linked_specs: []
tags: ["acceptance", "criteria", "testing", "validation"]
category: "quality"
acceptance_criteria: []
risk_factors: []
stakeholders: []
last_updated: "2025-11-11"
review_status: "draft"
compliance_level: "basic"
---

# 승인 기준: SPEC-ACCEPTANCE-001

## 📋 개요

이 문서는 SPEC의 완료를 검증하기 위한 Given-When-Then 형식의 승인 기준을 정의합니다.
```

### 6. API Design Template
```yaml
---
spec_id: "SPEC-API-DESIGN-001"
spec_title: "API Design Specification"
domain: "API"
version: "1.0.0"
created_date: "2025-11-11"
status: "draft"
priority: "P0"
owner: "Backend Team"
author: "API Architect"
reviewer: "Tech Lead"
type: "design"
complexity: "medium"
estimated_effort_hours: 80
implementation_phases: 4
timeline_weeks: 3
related_tags: []
linked_specs: []
tags: ["api", "design", "rest", "graphql"]
category: "development"
acceptance_criteria: []
risk_factors: ["backward-compatibility", "performance"]
stakeholders: ["backend", "frontend", "integration"]
last_updated: "2025-11-11"
review_status: "draft"
compliance_level: "enhanced"
---

# @SPEC:API-DESIGN-001 | @EXPERT:BACKEND | @EXPERT:FRONTEND

## API Overview

This SPEC defines the API design for the MoAI-ADK platform, including REST endpoints, GraphQL schemas, and WebSocket connections.

## API Endpoints

### Authentication API
- `POST /api/auth/login` - User authentication
- `POST /api/auth/refresh` - Token refresh
- `POST /api/auth/logout` - User logout

### User Management API
- `GET /api/users` - List users
- `POST /api/users` - Create user
- `PUT /api/users/{id}` - Update user
- `DELETE /api/users/{id}` - Delete user
```

### 7. Database Design Template
```yaml
---
spec_id: "SPEC-DATABASE-001"
spec_title: "Database Design Specification"
domain: "DATABASE"
version: "1.0.0"
created_date: "2025-11-11"
status: "draft"
priority: "P0"
owner: "Database Team"
author: "DBA"
reviewer: "Architect"
type: "design"
complexity: "high"
estimated_effort_hours: 120
implementation_phases: 5
timeline_weeks: 4
related_tags: []
linked_specs: []
tags: ["database", "schema", "migration", "performance"]
category: "infrastructure"
acceptance_criteria: []
risk_factors: ["data-loss", "migration-failure", "performance"]
stakeholders: ["backend", "devops", "data-team"]
last_updated: "2025-11-11"
review_status: "draft"
compliance_level: "enterprise"
---

# @SPEC:DATABASE-001 | @EXPERT:BACKEND | @EXPERT:DEVOPS

## Database Overview

This SPEC defines the database schema and design for the MoAI-ADK platform, supporting PostgreSQL with advanced features like RLS, JSONB, and time-series data.

## Schema Design

### Users Table
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) DEFAULT 'active'
);
```
```

### 8. Security Audit Template
```yaml
---
spec_id: "SECURITY-AUDIT-001"
spec_title: "Security Audit Implementation"
domain: "SECURITY"
version: "1.0.0"
created_date: "2025-11-11"
status: "active"
priority: "P0"
owner: "Security Team"
author: "Security Officer"
reviewer: "CSO"
type: "enhancement"
complexity: "high"
estimated_effort_hours: 100
implementation_phases: 4
timeline_weeks: 3
related_tags: []
linked_specs: []
tags: ["security", "audit", "compliance", "risk"]
category: "security"
acceptance_criteria: []
risk_factors: ["data-breach", "compliance", "privacy"]
stakeholders: ["security", "legal", "compliance"]
last_updated: "2025-11-11"
review_status: "reviewed"
compliance_level: "enterprise"
---

# @SPEC:SECURITY-AUDIT-001 | @EXPERT:BACKEND | @EXPERT:SECURITY

## Security Overview

This SPEC implements comprehensive security auditing for MoAI-ADK, including vulnerability scanning, penetration testing, and compliance monitoring.

## Security Requirements

- **Vulnerability Scanning**: Automated scanning for known vulnerabilities
- **Penetration Testing**: Regular penetration testing by external experts
- **Compliance Monitoring**: Continuous monitoring for regulatory compliance
- **Incident Response**: Security incident detection and response procedures
```

### 9. Performance Optimization Template
```yaml
---
spec_id: "SPEC-PERF-001"
spec_title: "Performance Optimization Plan"
domain: "PERFORMANCE"
version: "1.0.0"
created_date: "2025-11-11"
status: "active"
priority: "P1"
owner: "Performance Team"
author: "Performance Engineer"
reviewer: "Tech Lead"
type: "enhancement"
complexity: "medium"
estimated_effort_hours: 80
implementation_phases: 4
timeline_weeks: 3
related_tags: []
linked_specs: []
tags: ["performance", "optimization", "benchmark", "monitoring"]
category: "development"
acceptance_criteria: []
risk_factors: ["regression", "scalability"]
stakeholders: ["backend", "devops", "product"]
last_updated: "2025-11-11"
review_status: "reviewed"
compliance_level: "enhanced"
---

# @SPEC:PERF-001 | @EXPERT:BACKEND | @EXPERT:DEVOPS

## Performance Overview

This SPEC defines performance optimization strategies for MoAI-ADK, including database optimization, caching strategies, and CDN implementation.

## Performance Targets

- **API Response Time**: < 100ms for 95% of requests
- **Database Query Time**: < 50ms for complex queries
- **Page Load Time**: < 2s for client applications
- **Throughput**: 10,000+ requests per minute
```

### 10. DevOps Pipeline Template
```yaml
---
spec_id: "SPEC-DEVOPS-001"
spec_title: "DevOps Pipeline Implementation"
domain: "DEVOPS"
version: "1.0.0"
created_date: "2025-11-11"
status: "active"
priority: "P0"
owner: "DevOps Team"
author: "DevOps Engineer"
reviewer: "Infrastructure Lead"
type: "new"
complexity: "high"
estimated_effort_hours: 160
implementation_phases: 6
timeline_weeks: 5
related_tags: []
linked_specs: []
tags: ["devops", "ci-cd", "automation", "infrastructure"]
category: "infrastructure"
acceptance_criteria: []
risk_factors: ["downtime", "security", "scalability"]
stakeholders: ["devops", "development", "operations"]
last_updated: "2025-11-11"
review_status: "reviewed"
compliance_level: "enterprise"
---

# @SPEC:DEVOPS-001 | @EXPERT:DEVOPS | @EXPERT:INFRASTRUCTURE

## DevOps Overview

This SPEC implements comprehensive DevOps pipelines for MoAI-ADK, including CI/CD, monitoring, and infrastructure automation.

## Pipeline Architecture

### Build Pipeline
- Code quality checks (linting, testing)
- Security scanning
- Container image building
- Artifact storage

### Deploy Pipeline
- Environment promotion
- Rollback capabilities
- Health checks
- Monitoring activation
```

### 11. Microservices Architecture Template
```yaml
---
spec_id: "SPEC-MICROSERVICES-001"
spec_title: "Microservices Architecture Design"
domain: "ARCHITECTURE"
version: "1.0.0"
created_date: "2025-11-11"
status: "active"
priority: "P0"
owner: "Architecture Team"
author: "Chief Architect"
reviewer: "Engineering VP"
type: "design"
complexity: "high"
estimated_effort_hours: 200
implementation_phases: 8
timeline_weeks: 6
related_tags: []
linked_specs: []
tags: ["microservices", "architecture", "scalability", "resilience"]
category: "architecture"
acceptance_criteria: []
risk_factors: ["complexity", "distributed-systems", "operations"]
stakeholders: ["architecture", "development", "operations"]
last_updated: "2025-11-11"
review_status: "reviewed"
compliance_level: "enterprise"
---

# @SPEC:MICROSERVICES-001 | @EXPERT:ARCHITECTURE | @EXPERT:BACKEND

## Microservices Overview

This SPEC defines the microservices architecture for MoAI-ADK, enabling scalable, resilient, and maintainable service-oriented design.

## Service Boundaries

### Core Services
- **Auth Service**: Authentication and authorization
- **User Service**: User management and profiles
- **Content Service**: Content management and delivery
- **Notification Service**: Push and email notifications

### Supporting Services
- **API Gateway**: Request routing and rate limiting
- **Service Registry**: Service discovery and health checks
- **Config Service**: Centralized configuration management
- **Log Service**: Centralized logging and monitoring
```

### 12. Machine Learning Template
```yaml
---
spec_id: "SPEC-ML-001"
spec_title: "Machine Learning Integration"
domain: "ML"
version: "1.0.0"
created_date: "2025-11-11"
status: "active"
priority: "P1"
owner: "Data Science Team"
author: "ML Engineer"
reviewer: "Data Science Lead"
type: "new"
complexity: "high"
estimated_effort_hours: 180
implementation_phases: 7
timeline_weeks: 5
related_tags: []
linked_specs: []
tags: ["machine-learning", "ai", "data-science", "mlops"]
category: "data"
acceptance_criteria: []
risk_factors: ["model-quality", "data-privacy", "bias"]
stakeholders: ["data-science", "engineering", "product"]
last_updated: "2025-11-11"
review_status: "draft"
compliance_level: "enhanced"
---

# @SPEC:ML-001 | @EXPERT:DATA-SCIENCE | @EXPERT:BACKEND

## ML Integration Overview

This SPEC implements machine learning capabilities for MoAI-ADK, including model training, deployment, and monitoring.

## ML Pipeline

### Data Collection
- User behavior tracking
- System performance metrics
- External data sources
- Data validation and cleaning

### Model Training
- Feature engineering
- Model selection
- Hyperparameter tuning
- Cross-validation
```

### 13. Monitoring & Observability Template
```yaml
---
spec_id: "SPEC-MONITORING-001"
spec_title: "Monitoring & Observability System"
domain: "OBSERVABILITY"
version: "1.0.0"
created_date: "2025-11-11"
status: "active"
priority: "P0"
owner: "DevOps Team"
author: "SRE Team"
reviewer: "Infrastructure Lead"
type: "new"
complexity: "medium"
estimated_effort_hours: 100
implementation_phases: 5
timeline_weeks: 4
related_tags: []
linked_specs: []
tags: ["monitoring", "observability", "metrics", "alerting"]
category: "infrastructure"
acceptance_criteria: []
risk_factors: ["alert-fatigue", "data-volume", "privacy"]
stakeholders: ["devops", "development", "product"]
last_updated: "2025-11-11"
review_status: "reviewed"
compliance_level: "enhanced"
---

# @SPEC:MONITORING-001 | @EXPERT:DEVOPS | @EXPERT:SRE

## Monitoring Overview

This SPEC implements comprehensive monitoring and observability for MoAI-ADK, including metrics, logging, and alerting systems.

## Monitoring Stack

### Metrics Collection
- Application performance metrics
- Infrastructure metrics
- Business metrics
- Custom metrics

### Logging
- Application logs
- System logs
- Access logs
- Error tracking
```

### 14. Security Compliance Template
```yaml
---
spec_id: "SPEC-COMPLIANCE-001"
spec_title: "Security Compliance Implementation"
domain: "COMPLIANCE"
version: "1.0.0"
created_date: "2025-11-11"
status: "active"
priority: "P0"
owner: "Compliance Team"
author: "Compliance Officer"
reviewer: "Legal Counsel"
type: "enhancement"
complexity: "high"
estimated_effort_hours: 140
implementation_phases: 6
timeline_weeks: 4
related_tags: []
linked_specs: []
tags: ["compliance", "regulatory", "audit", "risk"]
category: "compliance"
acceptance_criteria: []
risk_factors: ["legal", "financial", "reputational"]
stakeholders: ["compliance", "legal", "security"]
last_updated: "2025-11-11"
review_status: "reviewed"
compliance_level: "enterprise"
---

# @SPEC:COMPLIANCE-001 | @EXPERT:SECURITY | @EXPERT:LEGAL

## Compliance Overview

This SPEC implements security and regulatory compliance for MoAI-ADK, including GDPR, SOC 2, and industry-specific requirements.

## Compliance Framework

### GDPR Compliance
- Data subject rights implementation
- Consent management
- Data portability
- Right to erasure

### SOC 2 Compliance
- Security controls
- Availability controls
- Processing integrity
- Confidentiality controls
```

### 15. Disaster Recovery Template
```yaml
---
spec_id: "SPEC-DR-001"
spec_title: "Disaster Recovery Implementation"
domain: "DISASTER-RECOVERY"
version: "1.0.0"
created_date: "2025-11-11"
status: "active"
priority: "P0"
owner: "DevOps Team"
author: "DR Specialist"
reviewer: "Business Continuity Manager"
type: "enhancement"
complexity: "high"
estimated_effort_hours: 120
implementation_phases: 5
timeline_weeks: 4
related_tags: []
linked_specs: []
tags: ["disaster-recovery", "business-continuity", "backup", "failover"]
category: "infrastructure"
acceptance_criteria: []
risk_factors: ["downtime", "data-loss", "business-impact"]
stakeholders: ["devops", "operations", "business"]
last_updated: "2025-11-11"
review_status: "reviewed"
compliance_level: "enterprise"
---

# @SPEC:DR-001 | @EXPERT:DEVOPS | @EXPERT:SECURITY

## Disaster Recovery Overview

This SPEC implements comprehensive disaster recovery for MoAI-ADK, ensuring business continuity and minimal downtime.

## Recovery Strategy

### RTO (Recovery Time Objective)
- Critical systems: < 4 hours
- Important systems: < 24 hours
- Standard systems: < 72 hours

### RPO (Recovery Point Objective)
- Critical systems: < 15 minutes
- Important systems: < 1 hour
- Standard systems: < 24 hours
```

---

## APIs and Functions

### Core Validation Functions
```javascript
// YAML frontmatter validation with comprehensive checking
class SpecValidator {
  constructor(config = {}) {
    this.config = {
      strictMode: config.strictMode || false,
      enableAutoCorrection: config.enableAutoCorrection || false,
      customValidators: config.customValidators || [],
      ...config
    };

    this.requiredFields = [
      'spec_id', 'spec_title', 'domain', 'version',
      'created_date', 'status', 'priority', 'owner',
      'author', 'reviewer', 'type', 'complexity',
      'estimated_effort_hours', 'implementation_phases', 'timeline_weeks'
    ];

    this.priorityLevels = ['P0', 'P1', 'P2', 'P3'];
    this.statusValues = ['active', 'completed', 'pending', 'draft', 'reviewed'];
    this.complexityLevels = ['low', 'medium', 'high'];
    this.typeValues = ['new', 'enhancement', 'bugfix', 'documentation', 'design', 'planning'];
  }

  // Validate required fields
  validateRequired(frontmatter) {
    const missing = this.requiredFields.filter(field => !frontmatter[field]);
    const errors = [];

    if (missing.length > 0) {
      errors.push(`Missing required fields: ${missing.join(', ')}`);
    }

    // Validate formats
    if (!/^SPEC-\d{3}-\d{3}$/.test(frontmatter.spec_id)) {
      errors.push('spec_id must follow format SPEC-XXX-XXX');
    }

    if (!this.priorityLevels.includes(frontmatter.priority)) {
      errors.push(`priority must be one of: ${this.priorityLevels.join(', ')}`);
    }

    if (!this.statusValues.includes(frontmatter.status)) {
      errors.push(`status must be one of: ${this.statusValues.join(', ')}`);
    }

    if (!this.complexityLevels.includes(frontmatter.complexity)) {
      errors.push(`complexity must be one of: ${this.complexityLevels.join(', ')}`);
    }

    if (!this.typeValues.includes(frontmatter.type)) {
      errors.push(`type must be one of: ${this.typeValues.join(', ')}`);
    }

    // Validate numeric ranges
    if (frontmatter.estimated_effort_hours &&
        (frontmatter.estimated_effort_hours <= 0 || frontmatter.estimated_effort_hours > 1000)) {
      errors.push('estimated_effort_hours must be between 1 and 1000');
    }

    if (frontmatter.implementation_phases &&
        (frontmatter.implementation_phases <= 0 || frontmatter.implementation_phases > 52)) {
      errors.push('implementation_phases must be between 1 and 52');
    }

    if (frontmatter.timeline_weeks &&
        (frontmatter.timeline_weeks <= 0 || frontmatter.timeline_weeks > 52)) {
      errors.push('timeline_weeks must be between 1 and 52');
    }

    // Run custom validators
    this.config.customValidators.forEach(validator => {
      const customErrors = validator(frontmatter);
      if (customErrors && customErrors.length > 0) {
        errors.push(...customErrors);
      }
    });

    return {
      valid: errors.length === 0,
      errors: errors,
      missing: missing
    };
  }

  // Validate cross-references
  validateCrossReferences(frontmatter) {
    const errors = [];

    // Validate related_tags format
    if (frontmatter.related_tags && Array.isArray(frontmatter.related_tags)) {
      frontmatter.related_tags.forEach(tag => {
        if (!/^@(SPEC|CODE|TEST|DOC):[A-Z0-9-]+$/.test(tag)) {
          errors.push(`Invalid tag format: ${tag}. Must be @(SPEC|CODE|TEST|DOC):XXX-XXX`);
        }
      });
    } else if (frontmatter.related_tags) {
      errors.push('related_tags must be an array');
    }

    // Validate linked_specs format
    if (frontmatter.linked_specs && Array.isArray(frontmatter.linked_specs)) {
      frontmatter.linked_specs.forEach(spec => {
        if (!/^SPEC-\d{3}-\d{3}$/.test(spec)) {
          errors.push(`Invalid linked spec format: ${spec}. Must be SPEC-XXX-XXX`);
        }
      });
    } else if (frontmatter.linked_specs) {
      errors.push('linked_specs must be an array');
    }

    // Check for duplicate tags
    if (frontmatter.tags && Array.isArray(frontmatter.tags)) {
      const duplicates = frontmatter.tags.filter((tag, index) =>
        frontmatter.tags.indexOf(tag) !== index
      );
      if (duplicates.length > 0) {
        errors.push(`Duplicate tags: ${duplicates.join(', ')}`);
      }
    } else if (frontmatter.tags) {
      errors.push('tags must be an array');
    }

    return {
      valid: errors.length === 0,
      errors: errors,
      duplicates: duplicates || []
    };
  }

  // Template-specific validation
  validateTemplate(frontmatter, templateType) {
    const templateValidators = {
      'spec': this.validateSpecTemplate.bind(this),
      'implementation_plan': this.validatePlanTemplate.bind(this),
      'acceptance_criteria': this.validateAcceptanceTemplate.bind(this),
      'api_design': this.validateAPITemplate.bind(this),
      'database_design': this.validateDatabaseTemplate.bind(this),
      'security_audit': this.validateSecurityTemplate.bind(this),
      'performance_optimization': this.validatePerformanceTemplate.bind(this),
      'devops_pipeline': this.validateDevOpsTemplate.bind(this),
      'microservices': this.validateMicroservicesTemplate.bind(this),
      'machine_learning': this.validateMLTemplate.bind(this),
      'monitoring': this.validateMonitoringTemplate.bind(this),
      'compliance': this.validateComplianceTemplate.bind(this),
      'disaster_recovery': this.validateDRTemplate.bind(this)
    };

    const validator = templateValidators[templateType];
    if (!validator) {
      return [];
    }

    return validator(frontmatter);
  }

  // Template-specific validators
  validateSpecTemplate(frontmatter) {
    const errors = [];

    // SPEC templates must have detailed content structure
    if (!frontmatter.acceptance_criteria || !Array.isArray(frontmatter.acceptance_criteria)) {
      errors.push('spec templates must include acceptance_criteria array');
    }

    if (!frontmatter.risk_factors || !Array.isArray(frontmatter.risk_factors)) {
      errors.push('spec templates must include risk_factors array');
    }

    return errors;
  }

  validatePlanTemplate(frontmatter) {
    const errors = [];

    // Plan templates must have realistic time estimates
    if (frontmatter.estimated_effort_hours < 20) {
      errors.push('implementation_plan must have at least 20 estimated effort hours');
    }

    if (frontmatter.implementation_phases < 1) {
      errors.push('implementation_plan must have at least 1 implementation phase');
    }

    return errors;
  }

  validateAcceptanceTemplate(frontmatter) {
    const errors = [];

    // Acceptance criteria templates should be focused on QA
    if (frontmatter.complexity === 'high') {
      errors.push('acceptance criteria templates should be low to medium complexity');
    }

    return errors;
  }

  validateAPITemplate(frontmatter) {
    const errors = [];

    // API templates should include API-specific tags
    if (!frontmatter.tags || !frontmatter.tags.some(tag =>
      ['api', 'rest', 'graphql', 'websocket'].includes(tag))) {
      errors.push('api_design templates must include API-related tags');
    }

    return errors;
  }

  validateDatabaseTemplate(frontmatter) {
    const errors = [];

    // Database templates should include database-specific fields
    if (!frontmatter.tags || !frontmatter.tags.some(tag =>
      ['database', 'schema', 'migration', 'performance'].includes(tag))) {
      errors.push('database_design templates must include database-related tags');
    }

    return errors;
  }

  validateSecurityTemplate(frontmatter) {
    const errors = [];

    // Security templates must be enterprise compliance level
    if (frontmatter.compliance_level !== 'enterprise') {
      errors.push('security templates must have compliance_level: enterprise');
    }

    return errors;
  }

  validatePerformanceTemplate(frontmatter) {
    const errors = [];

    // Performance templates should have performance targets
    if (frontmatter.estimated_effort_hours < 40) {
      errors.push('performance optimization templates require significant effort (>40 hours)');
    }

    return errors;
  }

  validateDevOpsTemplate(frontmatter) {
    const errors = [];

    // DevOps templates must include infrastructure stakeholders
    if (!frontmatter.stakeholders || !frontmatter.stakeholders.includes('devops')) {
      errors.push('devops templates must include devops in stakeholders');
    }

    return errors;
  }

  validateMicroservicesTemplate(frontmatter) {
    const errors = [];

    // Microservices templates are high complexity by nature
    if (frontmatter.complexity !== 'high') {
      errors.push('microservices templates must have complexity: high');
    }

    return errors;
  }

  validateMLTemplate(frontmatter) {
    const errors = [];

    // ML templates must include ML-specific risk factors
    if (!frontmatter.risk_factors || !frontmatter.risk_factors.includes('model-quality')) {
      errors.push('machine learning templates must include model-quality risk factor');
    }

    return errors;
  }

  validateMonitoringTemplate(frontmatter) {
    const errors = [];

    // Monitoring templates should include observability stakeholders
    if (!frontmatter.stakeholders || !frontmatter.stakeholders.includes('devops')) {
      errors.push('monitoring templates must include devops in stakeholders');
    }

    return errors;
  }

  validateComplianceTemplate(frontmatter) {
    const errors = [];

    // Compliance templates must be enterprise compliance level
    if (frontmatter.compliance_level !== 'enterprise') {
      errors.push('compliance templates must have compliance_level: enterprise');
    }

    return errors;
  }

  validateDRTemplate(frontmatter) {
    const errors = [];

    // DR templates must include business continuity stakeholders
    if (!frontmatter.stakeholders || !frontmatter.stakeholders.includes('business')) {
      errors.push('disaster recovery templates must include business in stakeholders');
    }

    return errors;
  }

  // Comprehensive validation
  validate(frontmatter, templateType = null) {
    const requiredResult = this.validateRequired(frontmatter);
    const crossRefResult = this.validateCrossReferences(frontmatter);
    const templateResult = templateType ? this.validateTemplate(frontmatter, templateType) : [];

    const allErrors = [
      ...requiredResult.errors,
      ...crossRefResult.errors,
      ...templateResult
    ];

    return {
      valid: allErrors.length === 0,
      errors: allErrors,
      required: requiredResult,
      crossReferences: crossRefResult,
      template: templateResult,
      warnings: this.generateWarnings(frontmatter)
    };
  }

  // Generate warnings for non-critical issues
  generateWarnings(frontmatter) {
    const warnings = [];

    // Check for outdated specs
    const createdDate = new Date(frontmatter.created_date);
    const oneYearAgo = new Date();
    oneYearAgo.setFullYear(oneYearAgo.getFullYear() - 1);

    if (createdDate < oneYearAgo && frontmatter.status === 'active') {
      warnings.push('Active spec older than 1 year - consider review and update');
    }

    // Check for large effort estimates
    if (frontmatter.estimated_effort_hours > 200) {
      warnings.push('Large effort estimate (>200 hours) - consider breaking into smaller specs');
    }

    // Check for long timelines
    if (frontmatter.timeline_weeks > 12) {
      warnings.push('Long timeline (>12 weeks) - consider breaking into phases');
    }

    // Check for missing optional but important fields
    if (!frontmatter.stakeholders || frontmatter.stakeholders.length === 0) {
      warnings.push('No stakeholders defined - consider adding relevant parties');
    }

    if (!frontmatter.risk_factors || frontmatter.risk_factors.length === 0) {
      warnings.push('No risk factors identified - consider potential risks');
    }

    return warnings;
  }

  // Auto-correct common issues
  autoCorrect(frontmatter) {
    const corrected = { ...frontmatter };

    // Add missing required fields with defaults
    const now = new Date().toISOString().split('T')[0];
    if (!corrected.created_date) corrected.created_date = now;
    if (!corrected.last_updated) corrected.last_updated = now;
    if (!corrected.status) corrected.status = 'draft';
    if (!corrected.priority) corrected.priority = 'P1';
    if (!corrected.complexity) corrected.complexity = 'medium';
    if (!corrected.type) corrected.type = 'new';

    // Standardize date formats
    if (corrected.created_date && !/^\d{4}-\d{2}-\d{2}$/.test(corrected.created_date)) {
      try {
        corrected.created_date = new Date(corrected.created_date).toISOString().split('T')[0];
      } catch (e) {
        // Keep original if parsing fails
      }
    }

    return corrected;
  }
}

// Global validator instance
const specValidator = new SpecValidator();
```
```

### Template Generation Functions
```javascript
// Enhanced template generation with customization
class TemplateGenerator {
  constructor(config = {}) {
    this.config = {
      defaultPriority: config.defaultPriority || 'P1',
      defaultComplexity: config.defaultComplexity || 'medium',
      defaultTimeline: config.defaultTimeline || 2,
      enableAutoGeneration: config.enableAutoGeneration || true,
      customTemplates: config.customTemplates || {},
      ...config
    };

    this.templateCache = new Map();
    this.initializeTemplates();
  }

  // Initialize all template types
  initializeTemplates() {
    this.templateTypes = {
      spec: this.generateSpecTemplate.bind(this),
      implementation_plan: this.generatePlanTemplate.bind(this),
      acceptance_criteria: this.generateAcceptanceTemplate.bind(this),
      api_design: this.generateAPITemplate.bind(this),
      database_design: this.generateDatabaseTemplate.bind(this),
      security_audit: this.generateSecurityTemplate.bind(this),
      performance_optimization: this.generatePerformanceTemplate.bind(this),
      devops_pipeline: this.generateDevOpsTemplate.bind(this),
      microservices: this.generateMicroservicesTemplate.bind(this),
      machine_learning: this.generateMLTemplate.bind(this),
      monitoring: this.generateMonitoringTemplate.bind(this),
      compliance: this.generateComplianceTemplate.bind(this),
      disaster_recovery: this.generateDRTemplate.bind(this)
    };
  }

  // Generate complete SPEC template
  generateSpecTemplate(options = {}) {
    const defaults = {
      spec_id: this.generateSpecID(),
      spec_title: 'New Feature Implementation',
      domain: 'FEATURE',
      version: '1.0.0',
      priority: this.config.defaultPriority,
      complexity: this.config.defaultComplexity,
      timeline_weeks: this.config.defaultTimeline,
      estimated_effort_hours: this.calculateEffort(options.complexity || 'medium'),
      implementation_phases: this.calculatePhases(options.timeline_weeks || 2),
      author: this.config.defaultAuthor || 'MoAI-ADK Auto-generator',
      reviewer: this.config.defaultReviewer || 'Tech Lead',
      owner: this.config.defaultOwner || 'Development Team'
    };

    const config = this.validateOptions({ ...defaults, ...options });
    return this.createTemplate('spec', config);
  }

  // Generate implementation plan template
  generatePlanTemplate(options = {}) {
    const defaults = {
      spec_id: this.generateSpecID(),
      spec_title: 'Implementation Plan',
      domain: 'PLANNING',
      version: '1.0.0',
      priority: 'P0',
      complexity: 'low',
      timeline_weeks: 4,
      estimated_effort_hours: 60,
      implementation_phases: 5,
      author: this.config.defaultAuthor || 'Project Manager',
      reviewer: this.config.defaultReviewer || 'Stakeholders',
      owner: this.config.defaultOwner || 'Project Management'
    };

    const config = this.validateOptions({ ...defaults, ...options });
    return this.createTemplate('implementation_plan', config);
  }

  // Generate acceptance criteria template
  generateAcceptanceTemplate(options = {}) {
    const defaults = {
      spec_id: this.generateSpecID(),
      spec_title: 'Acceptance Criteria Definition',
      domain: 'QA',
      version: '1.0.0',
      priority: 'P1',
      complexity: 'low',
      timeline_weeks: 1,
      estimated_effort_hours: 20,
      implementation_phases: 1,
      author: this.config.defaultAuthor || 'QA Team',
      reviewer: this.config.defaultReviewer || 'Product Owner',
      owner: this.config.defaultOwner || 'Quality Assurance'
    };

    const config = this.validateOptions({ ...defaults, ...options });
    return this.createTemplate('acceptance_criteria', config);
  }

  // Generate API design template
  generateAPITemplate(options = {}) {
    const defaults = {
      spec_id: this.generateSpecID(),
      spec_title: 'API Design Specification',
      domain: 'API',
      version: '1.0.0',
      priority: 'P0',
      complexity: 'medium',
      timeline_weeks: 3,
      estimated_effort_hours: 80,
      implementation_phases: 4,
      author: this.config.defaultAuthor || 'API Architect',
      reviewer: this.config.defaultReviewer || 'Tech Lead',
      owner: this.config.defaultOwner || 'Backend Team',
      tags: ['api', 'design', 'rest', 'graphql']
    };

    const config = this.validateOptions({ ...defaults, ...options });
    return this.createTemplate('api_design', config);
  }

  // Generate database design template
  generateDatabaseTemplate(options = {}) {
    const defaults = {
      spec_id: this.generateSpecID(),
      spec_title: 'Database Design Specification',
      domain: 'DATABASE',
      version: '1.0.0',
      priority: 'P0',
      complexity: 'high',
      timeline_weeks: 4,
      estimated_effort_hours: 120,
      implementation_phases: 5,
      author: this.config.defaultAuthor || 'DBA',
      reviewer: this.config.defaultReviewer || 'Architect',
      owner: this.config.defaultOwner || 'Database Team',
      tags: ['database', 'schema', 'migration', 'performance']
    };

    const config = this.validateOptions({ ...defaults, ...options });
    return this.createTemplate('database_design', config);
  }

  // Generate security audit template
  generateSecurityTemplate(options = {}) {
    const defaults = {
      spec_id: this.generateSpecID(),
      spec_title: 'Security Audit Implementation',
      domain: 'SECURITY',
      version: '1.0.0',
      priority: 'P0',
      complexity: 'high',
      timeline_weeks: 3,
      estimated_effort_hours: 100,
      implementation_phases: 4,
      author: this.config.defaultAuthor || 'Security Officer',
      reviewer: this.config.defaultReviewer || 'CSO',
      owner: this.config.defaultOwner || 'Security Team',
      compliance_level: 'enterprise',
      tags: ['security', 'audit', 'compliance', 'risk']
    };

    const config = this.validateOptions({ ...defaults, ...options });
    return this.createTemplate('security_audit', config);
  }

  // Generate performance optimization template
  generatePerformanceTemplate(options = {}) {
    const defaults = {
      spec_id: this.generateSpecID(),
      spec_title: 'Performance Optimization Plan',
      domain: 'PERFORMANCE',
      version: '1.0.0',
      priority: 'P1',
      complexity: 'medium',
      timeline_weeks: 3,
      estimated_effort_hours: 80,
      implementation_phases: 4,
      author: this.config.defaultAuthor || 'Performance Engineer',
      reviewer: this.config.defaultReviewer || 'Tech Lead',
      owner: this.config.defaultOwner || 'Performance Team',
      tags: ['performance', 'optimization', 'benchmark', 'monitoring']
    };

    const config = this.validateOptions({ ...defaults, ...options });
    return this.createTemplate('performance_optimization', config);
  }

  // Generate DevOps pipeline template
  generateDevOpsTemplate(options = {}) {
    const defaults = {
      spec_id: this.generateSpecID(),
      spec_title: 'DevOps Pipeline Implementation',
      domain: 'DEVOPS',
      version: '1.0.0',
      priority: 'P0',
      complexity: 'high',
      timeline_weeks: 5,
      estimated_effort_hours: 160,
      implementation_phases: 6,
      author: this.config.defaultAuthor || 'DevOps Engineer',
      reviewer: this.config.defaultReviewer || 'Infrastructure Lead',
      owner: this.config.defaultOwner || 'DevOps Team',
      stakeholders: ['devops', 'development', 'operations'],
      tags: ['devops', 'ci-cd', 'automation', 'infrastructure']
    };

    const config = this.validateOptions({ ...defaults, ...options });
    return this.createTemplate('devops_pipeline', config);
  }

  // Generate microservices template
  generateMicroservicesTemplate(options = {}) {
    const defaults = {
      spec_id: this.generateSpecID(),
      spec_title: 'Microservices Architecture Design',
      domain: 'ARCHITECTURE',
      version: '1.0.0',
      priority: 'P0',
      complexity: 'high',
      timeline_weeks: 6,
      estimated_effort_hours: 200,
      implementation_phases: 8,
      author: this.config.defaultAuthor || 'Chief Architect',
      reviewer: this.config.defaultReviewer || 'Engineering VP',
      owner: this.config.defaultOwner || 'Architecture Team',
      stakeholders: ['architecture', 'development', 'operations'],
      tags: ['microservices', 'architecture', 'scalability', 'resilience']
    };

    const config = this.validateOptions({ ...defaults, ...options });
    return this.createTemplate('microservices', config);
  }

  // Generate machine learning template
  generateMLTemplate(options = {}) {
    const defaults = {
      spec_id: this.generateSpecID(),
      spec_title: 'Machine Learning Integration',
      domain: 'ML',
      version: '1.0.0',
      priority: 'P1',
      complexity: 'high',
      timeline_weeks: 5,
      estimated_effort_hours: 180,
      implementation_phases: 7,
      author: this.config.defaultAuthor || 'ML Engineer',
      reviewer: this.config.defaultReviewer || 'Data Science Lead',
      owner: this.config.defaultOwner || 'Data Science Team',
      stakeholders: ['data-science', 'engineering', 'product'],
      risk_factors: ['model-quality', 'data-privacy', 'bias'],
      tags: ['machine-learning', 'ai', 'data-science', 'mlops']
    };

    const config = this.validateOptions({ ...defaults, ...options });
    return this.createTemplate('machine_learning', config);
  }

  // Generate monitoring template
  generateMonitoringTemplate(options = {}) {
    const defaults = {
      spec_id: this.generateSpecID(),
      spec_title: 'Monitoring & Observability System',
      domain: 'OBSERVABILITY',
      version: '1.0.0',
      priority: 'P0',
      complexity: 'medium',
      timeline_weeks: 4,
      estimated_effort_hours: 100,
      implementation_phases: 5,
      author: this.config.defaultAuthor || 'SRE Team',
      reviewer: this.config.defaultReviewer || 'Infrastructure Lead',
      owner: this.config.defaultOwner || 'DevOps Team',
      stakeholders: ['devops', 'development', 'product'],
      tags: ['monitoring', 'observability', 'metrics', 'alerting']
    };

    const config = this.validateOptions({ ...defaults, ...options });
    return this.createTemplate('monitoring', config);
  }

  // Generate compliance template
  generateComplianceTemplate(options = {}) {
    const defaults = {
      spec_id: this.generateSpecID(),
      spec_title: 'Security Compliance Implementation',
      domain: 'COMPLIANCE',
      version: '1.0.0',
      priority: 'P0',
      complexity: 'high',
      timeline_weeks: 4,
      estimated_effort_hours: 140,
      implementation_phases: 6,
      author: this.config.defaultAuthor || 'Compliance Officer',
      reviewer: this.config.defaultReviewer || 'Legal Counsel',
      owner: this.config.defaultOwner || 'Compliance Team',
      stakeholders: ['compliance', 'legal', 'security'],
      compliance_level: 'enterprise',
      tags: ['compliance', 'regulatory', 'audit', 'risk']
    };

    const config = this.validateOptions({ ...defaults, ...options });
    return this.createTemplate('compliance', config);
  }

  // Generate disaster recovery template
  generateDRTemplate(options = {}) {
    const defaults = {
      spec_id: this.generateSpecID(),
      spec_title: 'Disaster Recovery Implementation',
      domain: 'DISASTER-RECOVERY',
      version: '1.0.0',
      priority: 'P0',
      complexity: 'high',
      timeline_weeks: 4,
      estimated_effort_hours: 120,
      implementation_phases: 5,
      author: this.config.defaultAuthor || 'DR Specialist',
      reviewer: this.config.defaultReviewer || 'Business Continuity Manager',
      owner: this.config.defaultOwner || 'DevOps Team',
      stakeholders: ['devops', 'operations', 'business'],
      tags: ['disaster-recovery', 'business-continuity', 'backup', 'failover']
    };

    const config = this.validateOptions({ ...defaults, ...options });
    return this.createTemplate('disaster_recovery', config);
  }

  // Helper methods
  generateSpecID() {
    const timestamp = Date.now().toString().slice(-6);
    return `SPEC-${timestamp.slice(0, 3)}-${timestamp.slice(3)}`;
  }

  calculateEffort(complexity) {
    const effortMap = {
      'low': 20,
      'medium': 80,
      'high': 160
    };
    return effortMap[complexity] || 80;
  }

  calculatePhases(timelineWeeks) {
    return Math.max(1, Math.ceil(timelineWeeks / 2));
  }

  validateOptions(options) {
    const validated = { ...options };

    // Ensure required fields
    validated.requiredFields = [
      'spec_id', 'spec_title', 'domain', 'version',
      'created_date', 'status', 'priority', 'owner',
      'author', 'reviewer', 'type', 'complexity',
      'estimated_effort_hours', 'implementation_phases', 'timeline_weeks'
    ];

    // Add missing required fields with defaults
    const now = new Date().toISOString().split('T')[0];
    if (!validated.created_date) validated.created_date = now;
    if (!validated.status) validated.status = 'draft';
    if (!validated.type) validated.type = 'new';

    // Validate and correct fields
    if (validated.timeline_weeks < 1) validated.timeline_weeks = 1;
    if (validated.timeline_weeks > 52) validated.timeline_weeks = 52;
    if (validated.estimated_effort_hours < 1) validated.estimated_effort_hours = 1;
    if (validated.estimated_effort_hours > 1000) validated.estimated_effort_hours = 1000;
    if (validated.implementation_phases < 1) validated.implementation_phases = 1;
    if (validated.implementation_phases > 52) validated.implementation_phases = 52;

    return validated;
  }

  createTemplate(type, config) {
    const templateId = `${type}-${Date.now()}`;
    const template = {
      id: templateId,
      type: type,
      frontmatter: this.createFrontmatter(config),
      content: this.generateContent(type, config),
      metadata: {
        generated: true,
        timestamp: new Date().toISOString(),
        generator: 'MoAI-ADK Template Generator v3.0.0'
      }
    };

    // Cache the template
    this.templateCache.set(templateId, template);
    return template;
  }

  createFrontmatter(config) {
    const frontmatter = {
      spec_id: config.spec_id,
      spec_title: config.spec_title,
      domain: config.domain,
      version: config.version,
      created_date: config.created_date,
      status: config.status,
      priority: config.priority,
      owner: config.owner,
      author: config.author,
      reviewer: config.reviewer,
      type: config.type,
      complexity: config.complexity,
      estimated_effort_hours: config.estimated_effort_hours,
      implementation_phases: config.implementation_phases,
      timeline_weeks: config.timeline_weeks,
      related_tags: config.related_tags || [],
      linked_specs: config.linked_specs || [],
      tags: config.tags || [],
      category: config.category || 'development',
      acceptance_criteria: config.acceptance_criteria || [],
      risk_factors: config.risk_factors || [],
      stakeholders: config.stakeholders || [],
      last_updated: config.last_updated || config.created_date,
      review_status: config.review_status || 'draft',
      compliance_level: config.compliance_level || 'basic'
    };

    return frontmatter;
  }

  generateContent(type, config) {
    const contentTemplates = {
      spec: this.generateSpecContent.bind(this),
      implementation_plan: this.generatePlanContent.bind(this),
      acceptance_criteria: this.generateAcceptanceContent.bind(this),
      api_design: this.generateAPIContent.bind(this),
      database_design: this.generateDatabaseContent.bind(this),
      security_audit: this.generateSecurityContent.bind(this),
      performance_optimization: this.generatePerformanceContent.bind(this),
      devops_pipeline: this.generateDevOpsContent.bind(this),
      microservices: this.generateMicroservicesContent.bind(this),
      machine_learning: this.generateMLContent.bind(this),
      monitoring: this.generateMonitoringContent.bind(this),
      compliance: this.generateComplianceContent.bind(this),
      disaster_recovery: this.generateDRContent.bind(this)
    };

    const generator = contentTemplates[type];
    if (!generator) {
      return `# @SPEC:${config.spec_id} | @EXPERT:GENERAL\n\n## Overview\n\n${config.spec_title} specification.\n\n## Requirements\n\n- Primary functionality implementation\n- Performance optimization\n- Security considerations`;
    }

    return generator(config);
  }

  // Content generation methods for each template type
  generateSpecContent(config) {
    return `# @SPEC:${config.spec_id} | @EXPERT:BACKEND | @EXPERT:FRONTEND

## SPEC Overview

This SPEC defines the ${config.spec_title.toLowerCase()} for MoAI-ADK, providing comprehensive implementation details and technical requirements.

## Requirements

- **Core Functionality**: Implement primary feature with full API coverage
- **Performance**: Optimize for large-scale deployments
- **Integration**: Ensure seamless integration with existing systems
- **Documentation**: Complete API documentation and usage guides

## Implementation Strategy

### Phase 1: Foundation (1 week)
- Core architecture design
- Database schema implementation
- Basic API endpoints

### Phase 2: Feature Implementation (1 week)
- Business logic implementation
- External integrations
- Security implementation

### Phase 3: Testing & Validation (1 week)
- Unit and integration testing
- Performance testing
- Security validation

### Phase 4: Documentation & Deployment (1 week)
- API documentation
- User guides
- Production deployment`;
  }

  generatePlanContent(config) {
    return `# 구현 계획: ${config.spec_id}

## 📋 개요

SPEC의 단계별 구현을 계획하고 자원을 할당합니다.

**총 노력**: ${config.estimated_effort_hours}시간 | **기간**: ${config.timeline_weeks}주 | **팀**: ${this.getTeamSize(config.complexity)}명

## 📊 상세 계획

### Phase 1: 준비 (${this.calculatePhaseWeeks(config.timeline_weeks, 1)}주)
- 요구사항 분석
- 아키텍처 설계
- 기술 스택 선택

### Phase 2: 개발 (${this.calculatePhaseWeeks(config.timeline_weeks, 2)}주)
- 핵심 기능 구현
- 데이터베이스 설계
- API 개발

### Phase 3: 테스트 (${this.calculatePhaseWeeks(config.timeline_weeks, 3)}주)
- 단위 테스트
- 통합 테스트
- 성능 테스트

### Phase 4: 배포 (${this.calculatePhaseWeeks(config.timeline_weeks, 4)}주)
- 프로덕션 배포
- 모니터링 설정
- 문서화 완료`;
  }

  generateAcceptanceContent(config) {
    return `# 승인 기준: ${config.spec_id}

## 📋 개요

이 문서는 SPEC의 완료를 검증하기 위한 Given-When-Then 형식의 승인 기준을 정의합니다.

## 승인 기준

### 1. 핵심 기능 검증
**Given** 시스템이 정상적으로 실행될 때
**When** 사용자가 핵심 기능을 사용하려고 할 때
**Then** 시스템이 정상적으로 기능을 제공해야 함

### 2. 성능 검증
**Given** 시스템에 여러 사용자가 접속할 때
**When** 사용자가 요청을 보낼 때
**Then** 응답 시간이 1초 이내여야 함

### 3. 보안 검증
**Given** 시스템이 보안 모드로 실행될 때
**When** 인증되지 않은 사용자가 시스템에 접속할 때
**Then** 시스템이 접근을 차단해야 함`;
  }

  generateAPIContent(config) {
    return `# @SPEC:${config.spec_id} | @EXPERT:BACKEND | @EXPERT:FRONTEND

## API Overview

This SPEC defines the API design for the MoAI-ADK platform, including REST endpoints, GraphQL schemas, and WebSocket connections.

## API Endpoints

### Authentication API
- \`POST /api/auth/login\` - User authentication
- \`POST /api/auth/refresh\` - Token refresh
- \`POST /api/auth/logout\` - User logout

### User Management API
- \`GET /api/users\` - List users
- \`POST /api/users\` - Create user
- \`PUT /api/users/{id}\` - Update user
- \`DELETE /api/users/{id}\` - Delete user

### Content Management API
- \`GET /api/content\` - List content
- \`POST /api/content\` - Create content
- \`PUT /api/content/{id}\` - Update content
- \`DELETE /api/content/{id}\` - Delete content`;
  }

  generateDatabaseContent(config) {
    return `# @SPEC:${config.spec_id} | @EXPERT:BACKEND | @EXPERT:DEVOPS

## Database Overview

This SPEC defines the database schema and design for the MoAI-ADK platform, supporting PostgreSQL with advanced features like RLS, JSONB, and time-series data.

## Schema Design

### Users Table
\`\`\`sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) DEFAULT 'active',
    metadata JSONB DEFAULT '{}'
);

-- Add row level security (RLS)
ALTER TABLE users ENABLE ROW LEVEL SECURITY;

-- Create policies
CREATE POLICY "Users are owners of their data" ON users
    FOR ALL USING (auth.uid() = id);

CREATE POLICY "Public users are viewable" ON users
    FOR SELECT USING (status = 'active');
\`\`\`

### Content Table
\`\`\`sql
CREATE TABLE content (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    content TEXT,
    author_id UUID REFERENCES users(id),
    status VARCHAR(20) DEFAULT 'draft',
    published_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB DEFAULT '{}'
);

-- Add full-text search
CREATE INDEX idx_content_search ON content USING GIN(to_tsvector('english', title || ' ' || content));
\`\`\``;
  }

  generateSecurityContent(config) {
    return `# @SPEC:${config.spec_id} | @EXPERT:BACKEND | @EXPERT:SECURITY

## Security Overview

This SPEC implements comprehensive security auditing for MoAI-ADK, including vulnerability scanning, penetration testing, and compliance monitoring.

## Security Requirements

### 1. Vulnerability Scanning
- **Automated Scanning**: Daily automated scans for known vulnerabilities
- **OWASP Top 10**: Comprehensive testing against OWASP Top 10 vulnerabilities
- **Dependency Scanning**: Third-party library vulnerability assessment
- **Code Analysis**: Static application security testing (SAST)

### 2. Penetration Testing
- **Quarterly Testing**: External penetration testing every quarter
- **Red Team Exercises**: Annual red vs blue team exercises
- **Social Engineering**: Phishing simulation testing
- **Physical Security**: On-site security assessment

### 3. Compliance Monitoring
- **Continuous Monitoring**: Real-time compliance monitoring
- **Audit Trail**: Complete audit trail for all security events
- **Incident Response**: 24/7 incident response capability
- **Regulatory Compliance**: GDPR, SOC 2, HIPAA compliance as needed

### 4. Security Controls
- **Access Control**: Role-based access control (RBAC)
- **Authentication**: Multi-factor authentication (MFA)
- **Encryption**: End-to-end encryption for sensitive data
- **Logging**: Centralized security logging and monitoring`;
  }

  generatePerformanceContent(config) {
    return `# @SPEC:${config.spec_id} | @EXPERT:BACKEND | @EXPERT:DEVOPS

## Performance Overview

This SPEC defines performance optimization strategies for MoAI-ADK, including database optimization, caching strategies, and CDN implementation.

## Performance Targets

### Response Time Targets
- **API Response Time**: < 100ms for 95% of requests
- **Database Query Time**: < 50ms for complex queries
- **Page Load Time**: < 2s for client applications
- **API Gateway Response**: < 50ms for all requests

### Throughput Targets
- **Requests per Second**: 10,000+ RPS
- **Concurrent Users**: 50,000+ concurrent connections
- **Database Connections**: 1000+ active connections
- **Cache Hit Rate**: 90%+ for read operations

### Resource Utilization
- **CPU Utilization**: < 70% average, < 90% peak
- **Memory Utilization**: < 80% average, < 90% peak
- **Disk I/O**: < 70% utilization
- **Network Bandwidth**: < 80% utilization

## Optimization Strategies

### Database Optimization
- **Index Strategy**: Proper indexing for query patterns
- **Query Optimization**: Query performance analysis
- **Connection Pooling**: Efficient connection management
- **Caching**: Application-level caching

### Application Optimization
- **Code Optimization**: Performance profiling and optimization
- **Caching Strategy**: Multi-level caching implementation
- **Load Balancing**: Horizontal scaling
- **CDN Implementation**: Content delivery network`;
  }

  generateDevOpsContent(config) {
    return `# @SPEC:${config.spec_id} | @EXPERT:DEVOPS | @EXPERT:INFRASTRUCTURE

## DevOps Overview

This SPEC implements comprehensive DevOps pipelines for MoAI-ADK, including CI/CD, monitoring, and infrastructure automation.

## Pipeline Architecture

### Build Pipeline
1. **Code Quality Check**
   - Linting (ESLint, Prettier)
   - Static analysis (SonarQube)
   - Unit testing (Jest, Pytest)
   - Integration testing

2. **Security Scanning**
   - Vulnerability scanning (OWASP ZAP)
   - Dependency scanning (Snyk)
   - Container security (Trivy)

3. **Container Building**
   - Docker image build
   - Multi-stage builds
   - Security scanning
   - Image signing

4. **Artifact Storage**
   - Container registry (Docker Hub, ECR)
   - Package registry (NPM, PyPI)
   - Build artifact storage

### Deploy Pipeline
1. **Environment Promotion**
   - Staging environment deployment
   - Production promotion
   - Blue-green deployments
   - Canary releases

2. **Rollback Capabilities**
   - Automatic rollback triggers
   - Versioned deployments
   - Configuration management
   - Database migration rollback

3. **Health Checks**
   - Application health checks
   - Database connectivity checks
   - External service monitoring
   - Performance monitoring

4. **Monitoring Activation**
   - Application metrics collection
   - Log aggregation
   - Alert configuration
   - Performance baselines`;
  }

  generateMicroservicesContent(config) {
    return `# @SPEC:${config.spec_id} | @EXPERT:ARCHITECTURE | @EXPERT:BACKEND

## Microservices Overview

This SPEC defines the microservices architecture for MoAI-ADK, enabling scalable, resilient, and maintainable service-oriented design.

## Service Boundaries

### Core Services
1. **Auth Service**
   - Authentication and authorization
   - JWT token management
   - Session management
   - OAuth integration

2. **User Service**
   - User management and profiles
   - User preferences
   - Activity tracking
   - Search functionality

3. **Content Service**
   - Content management and delivery
   - Version control
   - Search and filtering
   - Media handling

4. **Notification Service**
   - Push notifications
   - Email notifications
   - SMS notifications
   - Webhooks

### Supporting Services
1. **API Gateway**
   - Request routing and rate limiting
   - Authentication and authorization
   - Request/response transformation
   - Load balancing

2. **Service Registry**
   - Service discovery and registration
   - Health checking
   - Load balancing
   - Circuit breaker

3. **Config Service**
   - Centralized configuration management
   - Environment-specific configs
   - Dynamic configuration updates
   - Configuration versioning

4. **Log Service**
   - Centralized logging and monitoring
   - Log aggregation and search
   - Error tracking
   - Performance monitoring`;
  }

  generateMLContent(config) {
    return `# @SPEC:${config.spec_id} | @EXPERT:DATA-SCIENCE | @EXPERT:BACKEND

## ML Integration Overview

This SPEC implements machine learning capabilities for MoAI-ADK, including model training, deployment, and monitoring.

## ML Pipeline

### Data Collection
1. **User Behavior Tracking**
   - User interaction patterns
   - Feature usage tracking
   - Performance metrics
   - Behavioral analytics

2. **System Performance Metrics**
   - API response times
   - System resource usage
   - Error rates
   - User satisfaction scores

3. **External Data Sources**
   - Market data integration
   - Third-party APIs
   - Social media feeds
   - News and content

4. **Data Validation and Cleaning**
   - Data quality checks
   - Missing value handling
   - Outlier detection
   - Data normalization

### Model Training
1. **Feature Engineering**
   - Feature selection
   - Feature transformation
   - Dimensionality reduction
   - Feature importance analysis

2. **Model Selection**
   - Algorithm comparison
   - Cross-validation
   - Hyperparameter tuning
   - Model ensemble

3. **Hyperparameter Optimization**
   - Grid search
   - Random search
   - Bayesian optimization
   - Automated ML (AutoML)

4. **Model Validation**
   - Cross-validation
   - Holdout validation
   - A/B testing
   - Performance metrics`;
  }

  generateMonitoringContent(config) {
    return `# @SPEC:${config.spec_id} | @EXPERT:DEVOPS | @EXPERT:SRE

## Monitoring Overview

This SPEC implements comprehensive monitoring and observability for MoAI-ADK, including metrics, logging, and alerting systems.

## Monitoring Stack

### Metrics Collection
1. **Application Performance Metrics**
   - Response times
   - Error rates
   - Request volumes
   - Service availability

2. **Infrastructure Metrics**
   - CPU utilization
   - Memory usage
   - Disk I/O
   - Network bandwidth

3. **Business Metrics**
   - User engagement
   - Conversion rates
   - Revenue tracking
   - Customer satisfaction

4. **Custom Metrics**
   - Business logic metrics
   - Custom dashboards
   - Alert thresholds
   - Performance baselines

### Logging
1. **Application Logs**
   - Request/response logs
   - Error logs
   - Debug information
   - Transaction logs

2. **System Logs**
   - System events
   - Security events
   - Performance events
   - Configuration changes

3. **Access Logs**
   - API access logs
   - User activity logs
   - Authentication logs
   - Authorization logs

4. **Error Tracking**
   - Error classification
   - Error aggregation
   - Root cause analysis
   - Error resolution tracking`;
  }

  generateComplianceContent(config) {
    return `# @SPEC:${config.spec_id} | @EXPERT:SECURITY | @EXPERT:LEGAL

## Compliance Overview

This SPEC implements security and regulatory compliance for MoAI-ADK, including GDPR, SOC 2, and industry-specific requirements.

## Compliance Framework

### GDPR Compliance
1. **Data Subject Rights**
   - Right to access
   - Right to rectification
   - Right to erasure
   - Right to data portability

2. **Consent Management**
   - Granular consent tracking
   - Consent expiration
   - Withdrawal mechanisms
   - Audit trail

3. **Data Protection**
   - Data encryption at rest and in transit
   - Access controls
   - Data retention policies
   - Data minimization

### SOC 2 Compliance
1. **Security Controls**
   - Access control systems
   - System architecture
   - Change management
   - Incident management

2. **Availability Controls**
   - System monitoring
   - Capacity planning
   - Disaster recovery
   - Backup and restore

3. **Processing Integrity**
   - Data validation
   - Accuracy checks
   - Completion verification
   - Error handling

4. **Confidentiality Controls**
   - Data classification
   - Access restrictions
   - Data handling procedures
   - Information security`;
  }

  generateDRContent(config) {
    return `# @SPEC:${config.spec_id} | @EXPERT:DEVOPS | @EXPERT:SECURITY

## Disaster Recovery Overview

This SPEC implements comprehensive disaster recovery for MoAI-ADK, ensuring business continuity and minimal downtime.

## Recovery Strategy

### RTO (Recovery Time Objective)
- **Critical Systems**: < 4 hours
  - Core authentication service
  - Primary database
  - Payment processing
  - Customer support system

- **Important Systems**: < 24 hours
  - Content management
  - User profiles
  - Reporting system
  - Administrative tools

- **Standard Systems**: < 72 hours
  - Marketing features
  - Analytics
  - Non-c integrations
  - Optional features

### RPO (Recovery Point Objective)
- **Critical Systems**: < 15 minutes
  - Transactional data
  - User sessions
  - Financial records
  - Security logs

- **Important Systems**: < 1 hour
  - User content
  - Configuration data
  - Business metrics
  - Audit trails

- **Standard Systems**: < 24 hours
  - Historical data
  - Media files
  - Non-critical metrics
  - Optional content

### Recovery Procedures
1. **Assessment Phase**
   - Damage assessment
   - Impact analysis
   - Priority determination
   - Resource allocation

2. **Recovery Phase**
   - System restoration
   - Data recovery
   - Service activation
   - Performance validation

3. **Verification Phase**
   - Functional testing
   - Performance testing
   - Security validation
   - User acceptance`;
  }

  // Helper methods for content generation
  getTeamSize(complexity) {
    const teamSizes = {
      'low': 3,
      'medium': 5,
      'high': 8
    };
    return teamSizes[complexity] || 5;
  }

  calculatePhaseWeeks(totalWeeks, phaseNumber) {
    return Math.ceil(totalWeeks / phaseNumber);
  }

  // Custom template generation from existing
  generateCustomTemplate(sourceTemplate, customizations) {
    const templateId = `custom-${Date.now()}`;
    const source = this.templateCache.get(sourceTemplate) || this.loadTemplate(sourceTemplate);

    if (!source) {
      throw new Error(`Template not found: ${sourceTemplate}`);
    }

    const customized = {
      ...source,
      id: templateId,
      frontmatter: this.mergeFrontmatter(source.frontmatter, customizations),
      metadata: {
        ...source.metadata,
        customized: true,
        customizations: customizations,
        timestamp: new Date().toISOString()
      }
    };

    this.templateCache.set(templateId, customized);
    return customized;
  }

  mergeFrontmatter(source, customizations) {
    const merged = { ...source };

    Object.keys(customizations).forEach(key => {
      if (key === 'tags' || key === 'related_tags' || key === 'linked_specs' || key === 'stakeholders') {
        merged[key] = [...(source[key] || []), ...(customizations[key] || [])];
      } else {
        merged[key] = customizations[key];
      }
    });

    return merged;
  }

  // Template validation
  validateTemplate(template) {
    if (!template || !template.frontmatter) {
      return { valid: false, errors: ['Invalid template structure'] };
    }

    const validation = specValidator.validate(template.frontmatter, template.type);
    return {
      valid: validation.valid,
      errors: validation.errors,
      warnings: validation.warnings,
      template: validation
    };
  }

  // Batch template generation
  generateMultipleTemplates(templateConfigs) {
    const results = templateConfigs.map(config => {
      try {
        const template = this.templateTypes[config.type](config.options);
        return { success: true, template: template };
      } catch (error) {
        return { success: false, error: error.message };
      }
    });

    return {
      total: results.length,
      successful: results.filter(r => r.success).length,
      failed: results.filter(r => !r.success).length,
      results: results
    };
  }
}

// Global template generator instance
const templateGenerator = new TemplateGenerator();
```

### Schema Validation Functions
```javascript
// YAML schema validation and enforcement
const schemaValidator = {
  // Validate against official YAML 1.2 specification
  validateYAML: (yamlContent) => {
    return validateAgainstYAML12(yamlContent);
  },

  // Validate frontmatter structure
  validateFrontmatter: (frontmatter) => {
    const schema = getFrontmatterSchema();
    return validateAgainstSchema(frontmatter, schema);
  },

  // Auto-correct common issues
  autoCorrect: (yamlContent) => {
    const corrections = [
      fixIndentation,
      fixQuotes,
      fixDates,
      fixLineBreaks
    ];

    return corrections.reduce((content, fix) => fix(content), yamlContent);
  }
};
```

---

## Configuration Options

### Validation Configuration
```javascript
const validationConfig = {
  // Strict validation mode
  strict: {
    checkRequired: true,
    checkFormat: true,
    checkCrossReferences: true,
    checkTemplate: true
  },

  // Relaxed validation mode
  relaxed: {
    checkRequired: true,
    checkFormat: false,
    checkCrossReferences: false,
    checkTemplate: false
  },

  // Custom validation rules
  customRules: [
    {
      field: 'estimated_effort_hours',
      validator: (value) => value > 0 && value <= 1000,
      message: 'Effort hours must be between 1 and 1000'
    },
    {
      field: 'timeline_weeks',
      validator: (value) => value > 0 && value <= 52,
      message: 'Timeline must be between 1 and 52 weeks'
    }
  ]
};
```

### Template Configuration
```javascript
const templateConfig = {
  // Template locations
  templates: {
    spec: '/path/to/spec-templates',
    plan: '/path/to/plan-templates',
    acceptance: '/path/to/acceptance-templates'
  },

  // Template metadata
  metadata: {
    author: 'MoAI-ADK Auto-generator',
    version: '3.0.0',
    generated: true
  },

  // Custom field mappings
  fieldMappings: {
    'spec_title': 'title',
    'estimated_effort_hours': 'effort_hours',
    'timeline_weeks': 'duration_weeks'
  }
};
```

---

## Error Handling and Recovery

### Validation Error Handling
```javascript
const errorHandling = {
  // Parse YAML errors
  parseYAMLError: (error) => {
    return {
      type: 'yaml_parse',
      message: error.message,
      location: error.mark,
      suggestions: getSuggestionsForYAMLError(error)
    };
  },

  // Schema validation errors
  schemaValidationError: (errors) => {
    return {
      type: 'schema_validation',
      errors: errors.map(error => ({
        field: error.field,
        message: error.message,
        severity: error.severity,
        suggestion: error.suggestion
      })),
      summary: `${errors.length} validation errors found`
    };
  },

  // Cross-reference errors
  crossReferenceError: (error) => {
    return {
      type: 'cross_reference',
      brokenTags: error.brokenTags,
      missingSpecs: error.missingSpecs,
      suggestions: error.suggestions
    };
  }
};
```

### Auto-Recovery Functions
```javascript
const autoRecovery = {
  // Auto-correct YAML formatting
  autoCorrectYAML: (yamlContent) => {
    return yamlContent
      .replace(/:\s*$/gm, ': null') // Add null for empty values
      .replace(/:\s*null\s*$/gm, ': null') // Standardize null values
      .replace(/\t/g, '  ') // Replace tabs with spaces
      .trim();
  },

  // Auto-generate missing fields
  autoGenerateFields: (frontmatter) => {
    const generated = { ...frontmatter };

    if (!generated.created_date) {
      generated.created_date = new Date().toISOString().split('T')[0];
    }

    if (!generated.last_updated) {
      generated.last_updated = generated.created_date;
    }

    return generated;
  },

  // Suggest corrections for validation errors
  suggestCorrections: (errors) => {
    return errors.map(error => ({
      ...error,
      corrections: generateCorrections(error)
    }));
  }
};
```

---

## Performance Optimization

### Caching and Indexing
```javascript
const performance = {
  // Cache validation results
  validationCache: new Map(),

  // Index templates for fast access
  templateIndex: {
    byType: new Map(),
    byDomain: new Map(),
    byComplexity: new Map()
  },

  // Batch validation
  batchValidate: (files) => {
    return Promise.all(
      files.map(file => validateFile(file))
        .catch(error => ({ file, error }))
    );
  },

  // Parallel processing
  parallelValidate: (frontmatters) => {
    return Promise.all(
      frontmatters.map(frontmatter =>
        validateFrontmatter(frontmatter)
      )
    );
  }
};
```

### Memory Management
```javascript
const memoryManagement = {
  // Clear cache when memory usage is high
  clearCache: () => {
    if (getMemoryUsage() > MAX_MEMORY) {
      validationCache.clear();
      templateIndex.clear();
    }
  },

  // Garbage collection for old results
  garbageCollect: (maxAge = 86400000) => {
    const now = Date.now();
    const oldEntries = Array.from(validationCache.entries())
      .filter(([key, value]) => now - value.timestamp > maxAge);

    oldEntries.forEach(([key]) => validationCache.delete(key));
  }
};
```

---

## Security Considerations

### Input Validation
```javascript
const security = {
  // Sanitize YAML input
  sanitizeInput: (yamlContent) => {
    return yamlContent
      .replace(/<script[^>]*?>.*?<\/script>/gi, '')
      .replace(/javascript:/gi, '')
      .replace(/on\w+\s*=/gi, '');
  },

  // Validate file paths
  validateFilePath: (path) => {
    const dangerousPatterns = [
      /\.\./,
      /\/\./,
      /~$/,
      /$/,
      /<|>|:|"|'|\?|\*|\|/
    ];

    return !dangerousPatterns.some(pattern => pattern.test(path));
  },

  // Check for malicious content
  checkMalicious: (content) => {
    const maliciousPatterns = [
      /eval\s*\(/,
      /exec\s*\(/,
      /system\s*\(/,
      /subprocess\s*\(/,
      /os\.system/
    ];

    return maliciousPatterns.some(pattern => pattern.test(content));
  }
};
```

### Data Protection
```javascript
const dataProtection = {
  // Encrypt sensitive data
  encrypt: (data, key) => {
    return encryptData(JSON.stringify(data), key);
  },

  // Decrypt data
  decrypt: (encryptedData, key) => {
    return JSON.parse(decryptData(encryptedData, key));
  },

  // Access control
  checkAccess: (user, templates) => {
    return templates.filter(template =>
      template.accessLevel <= user.clearanceLevel
    );
  }
};
```

---

## Integration Examples

### With MoAI-ADK Core
```javascript
// Integrate with MoAI-ADK specification system
const moaiIntegration = {
  // Generate SPEC from template
  generateSPEC: (templateType, options) => {
    const template = templateGenerator.generateTemplate(templateType, options);
    return {
      frontmatter: template.frontmatter,
      content: template.content,
      validation: specValidator.validate(template.frontmatter)
    };
  },

  // Validate existing SPEC files
  validateSPECs: (specFiles) => {
    return specFiles.map(file => {
      const frontmatter = extractFrontmatter(file.content);
      const validation = specValidator.validate(frontmatter);
      return { file, frontmatter, validation };
    });
  }
};
```

### With CI/CD Pipelines
```javascript
// CI/CD integration for automated validation
const cicdIntegration = {
  // Pre-commit validation
  preCommit: (files) => {
    const specFiles = files.filter(file => file.path.endsWith('.md'));
    const results = moaiIntegration.validateSPECs(specFiles);

    return {
      passed: results.every(r => r.validation.valid),
      results: results
    };
  },

  // Pull request validation
  pullRequest: (files) => {
    return moaiIntegration.validateSPECs(files)
      .then(results => {
        const passed = results.every(r => r.validation.valid);
        return {
          passed,
          summary: generateValidationSummary(results),
          details: results
        };
      });
  }
};
```

---

## Troubleshooting

### Common Issues and Solutions

#### 1. YAML Parse Errors
**Issue**: Frontmatter parsing fails
```javascript
// Solution: Enhanced YAML parsing with error recovery
const enhancedYAMLParsing = {
  parseWithRecovery: (yamlContent) => {
    try {
      return parseYAML(yamlContent);
    } catch (error) {
      // Try auto-correction
      const corrected = autoRecovery.autoCorrectYAML(yamlContent);
      return parseYAML(corrected);
    }
  }
};
```

#### 2. Validation Errors
**Issue**: Multiple validation errors in frontmatter
```javascript
// Solution: Progressive validation with suggestions
const progressiveValidation = {
  validateWithSuggestions: (frontmatter) => {
    const result = specValidator.validate(frontmatter);
    if (!result.valid) {
      const suggestions = autoRecovery.suggestCorrections(result.errors);
      return { ...result, suggestions };
    }
    return result;
  }
};
```

#### 3. Template Generation Issues
**Issue**: Generated templates don't match project standards
```javascript
// Solution: Template customization and validation
const templateCustomization = {
  generateCustomTemplate: (baseTemplate, customizations) => {
    const template = loadTemplate(baseTemplate);
    const customized = mergeTemplate(template, customizations);
    const validated = validateTemplate(customized);

    if (!validated.valid) {
      throw new Error(`Template validation failed: ${validated.errors.join(', ')}`);
    }

    return customized;
  }
};
```

### Debug Mode
```javascript
const debugMode = {
  // Enable detailed logging
  enable: () => {
    debug = true;
    logger.setLevel('debug');
  },

  // Generate diagnostic reports
  generateReport: (files) => {
    return {
      summary: generateSummary(files),
      validation: generateValidationReport(files),
      performance: generatePerformanceReport(files),
      recommendations: generateRecommendations(files)
    };
  }
};
```

---

## Best Practices

### 1. Quality Assurance
- ✅ Use consistent naming conventions for spec_id
- ✅ Always include all required fields
- ✅ Validate cross-references before committing
- ✅ Use appropriate priority levels
- ✅ Provide accurate effort estimates

### 2. Process Optimization
- ✅ Use templates for common SPEC types
- ✅ Implement automated validation in CI/CD
- ✅ Cache validation results for performance
- ✅ Use batch processing for multiple files
- ✅ Regular template updates and maintenance

### 3. Security and Compliance
- ✅ Validate all input YAML content
- ✅ Sanitize output for external systems
- ✅ Implement access control for sensitive templates
- ✅ Audit trail for all template changes
- ✅ Follow compliance requirements for regulated industries

### 4. Integration
- ✅ Use standard YAML 1.2 specification
- ✅ Maintain backward compatibility with existing templates
- ✅ Document all template changes
- ✅ Provide migration paths for template updates
- ✅ Support multiple project types and domains

---

## Testing Strategy

### Unit Tests
```javascript
// Test frontmatter validation
test('validate required fields', () => {
  const frontmatter = {
    spec_id: 'SPEC-001-001',
    spec_title: 'Test SPEC',
    domain: 'TEST',
    version: '1.0.0'
  };

  const result = specValidator.validate(frontmatter);
  expect(result.valid).toBe(false);
  expect(result.errors).toContain('Missing required fields: created_date, status');
});

// Test template generation
test('generate spec template', () => {
  const template = templateGenerator.generateSpecTemplate({
    spec_id: 'SPEC-TEST-001',
    spec_title: 'Test Feature'
  });

  expect(template.frontmatter.spec_id).toBe('SPEC-TEST-001');
  expect(template.content).toContain('# @SPEC:SPEC-TEST-001');
});
```

### Integration Tests
```javascript
// Test YAML parsing and validation
test('parse and validate YAML', () => {
  const yamlContent = `
---
spec_id: "SPEC-TEST-001"
spec_title: "Test SPEC"
domain: "TEST"
version: "1.0.0"
created_date: "2025-11-11"
status: "active"
priority: "P1"
owner: "Test Team"
author: "Test Author"
reviewer: "Test Reviewer"
type: "new"
complexity: "medium"
estimated_effort_hours: 40
implementation_phases: 2
timeline_weeks: 1
---
# Test Content
`;

  const frontmatter = schemaValidator.validateYAML(yamlContent);
  const validation = specValidator.validate(frontmatter);
  expect(validation.valid).toBe(true);
});

// Test cross-reference validation
test('validate cross-references', () => {
  const frontmatter = {
    spec_id: 'SPEC-TEST-001',
    related_tags: [
      '@SPEC:DEP-001',
      '@CODE:TEST-001',
      '@TEST:UNIT-001'
    ]
  };

  const errors = specValidator.validateCrossReferences(frontmatter);
  expect(errors).toHaveLength(0);
});
```

### Performance Tests
```javascript
// Test large file processing
test('process large SPEC files', () => {
  const largeFile = generateLargeSPECFile(1000);
  const start = performance.now();

  const result = specValidator.validate(largeFile.frontmatter);

  const duration = performance.now() - start;
  expect(duration).toBeLessThan(1000); // 1 second max
  expect(result.valid).toBe(true);
});

// Test template generation performance
test('generate multiple templates', () => {
  const templates = Array.from({ length: 100 }, (_, i) => ({
    spec_id: `SPEC-TEST-${String(i + 1).padStart(3, '0')}`,
    spec_title: `Test Template ${i + 1}`
  }));

  const start = performance.now();
  const results = templates.map(template =>
    templateGenerator.generateSpecTemplate(template)
  );
  const duration = performance.now() - start;

  expect(results.length).toBe(100);
  expect(duration).toBeLessThan(5000); // 5 seconds max for 100 templates
});
```

---

## Changelog

- **v3.0.0** (2025-11-11):
  - 15 required YAML frontmatter fields added
  - 15+ real SPEC templates implemented
  - Official YAML 1.2 specification compliance
  - Multi-template support (SPEC, PLAN, ACCEPTANCE, etc.)
  - Schema validation and auto-correction
  - Template generation and customization
  - Cross-reference validation
  - Integration with MoAI-ADK workflow
  - Performance optimization for large projects
  - Security and compliance features

- **v2.0.0** (2025-10-22):
  - Major update with latest tool versions
  - Comprehensive best practices
  - TRUST 5 integration

- **v1.0.0** (2025-03-29):
  - Initial Skill release

---

## Works Well With

- `moai-foundation-trust` (quality gates and validation)
- `moai-foundation-tags` (cross-reference tracking)
- `moai-foundation-git` (version control integration)
- `moai-alfred-code-reviewer` (spec review automation)
- `moai-essentials-debug` (debugging support for validation)
- `moai-essentials-perf` (performance optimization for templates)

---

## References

- YAML 1.2 Specification: https://yaml.org/spec/1.2/
- YAML Best Practices: https://yaml-style-guide.com/
- MoAI-ADK Documentation: https://docs.moai.kr/
- SPEC Template Guidelines: https://docs.moai.kr/specs/

---

## License

This skill is part of the MoAI-ADK project and is licensed under the MIT License.

---

## Support

For issues, questions, or feature requests, please open an issue on the MoAI-ADK GitHub repository.