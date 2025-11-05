# Reference

## Document Location Matrix

| Document Type | Primary Location | File Pattern | Auto-Create? | User Approval Required? |
|---------------|------------------|--------------|--------------|------------------------|
| SPEC Documents | `.moai/specs/SPEC-*/` | `spec.md`, `plan.md`, `acceptance.md` | ✅ Yes (during /alfred:1-plan) | ❌ No |
| Implementation Guides | `.moai/docs/` | `implementation-{SPEC-ID}.md` | ⚡ Contextual | ✅ Yes (if not explicitly requested) |
| Exploration Reports | `.moai/docs/` | `exploration-{topic}.md` | ⚡ Contextual | ✅ Yes (if not explicitly requested) |
| Strategy Documents | `.moai/docs/` | `strategy-{topic}.md` | ⚡ Contextual | ✅ Yes (if not explicitly requested) |
| Sync Reports | `.moai/reports/` | `sync-report-{date}.md` | ✅ Yes (during /alfred:3-sync) | ❌ No |
| Analysis Reports | `.moai/analysis/` | `*-analysis.md` | ⚡ Contextual | ✅ Yes (if not explicitly requested) |
| Public Documentation | Project Root | `README.md`, `CHANGELOG.md` | ❌ No | ✅ Yes (always) |
| Memory/Session Data | `.moai/memory/` | `session-state.json` | ✅ Yes (runtime) | ❌ No |

## Directory Structure Standards

### Complete MoAI Directory Layout

```
.moai/
├── config.json                    # Project configuration
├── docs/                          # Internal documentation
│   ├── implementation-*.md        # Implementation guides
│   ├── exploration-*.md           # Investigation reports
│   ├── strategy-*.md              # Strategic documents
│   └── guide-*.md                 # How-to guides
├── specs/                         # SPEC document containers
│   ├── SPEC-ID-001/               # Individual SPEC directories
│   │   ├── spec.md               # Requirements specification
│   │   ├── plan.md               # Implementation plan
│   │   └── acceptance.md         # Acceptance criteria
│   └── SPEC-ID-002/
├── reports/                       # Generated reports
│   ├── sync-report-YYYYMMDD.md   # Synchronization reports
│   └── tag-validation-YYYYMMDD.md # Tag validation reports
├── analysis/                      # Technical analysis
│   ├── performance-analysis.md    # Performance studies
│   ├── architecture-analysis.md   # Architecture reviews
│   └── security-analysis.md       # Security assessments
└── memory/                        # Runtime data (session state)
    └── session-state.json         # Current session context
```

## Forbidden Patterns

### Root Directory Restrictions

**ABSOLUTELY FORBIDDEN in project root:**
- `*_GUIDE.md` (e.g., `IMPLEMENTATION_GUIDE.md`)
- `*_REPORT.md` (e.g., `ANALYSIS_REPORT.md`)
- `*_ANALYSIS.md` (e.g., `PERFORMANCE_ANALYSIS.md`)
- `*_EXPLORATION.md` (e.g., `TECH_EXPLORATION.md`)
- `IMPLEMENTATION.md` (without proper location)

**ALLOWED in project root (only):**
- `README.md` - Official project documentation
- `CHANGELOG.md` - Version history
- `CONTRIBUTING.md` - Contribution guidelines
- `LICENSE` - License file

### Misplacement Prevention

**Wrong Location → Correct Location:**

| Wrong Location | Correct Location | Reason |
|----------------|------------------|--------|
| `./sync-report.md` | `.moai/reports/sync-report-YYYYMMDD.md` | Sync reports belong in reports/ |
| `./implementation-SPEC-001.md` | `.moai/docs/implementation-SPEC-001.md` | Internal docs go in docs/ |
| `./exploration-topic.md` | `.moai/docs/exploration-topic.md` | Explorations are internal docs |
| `./SPEC-001.md` | `.moai/specs/SPEC-001/spec.md` | SPECs need directory structure |
| `.moai/docs/sync-report.md` | `.moai/reports/sync-report-YYYYMMDD.md` | Don't mix report types |

## Naming Conventions

### Standard Patterns

1. **Implementation Guides**: `implementation-{SPEC-ID}.md`
   - Examples: `implementation-SPEC-001.md`, `implementation-user-auth.md`

2. **Exploration Reports**: `exploration-{topic}.md`
   - Examples: `exploration-performance-optimization.md`, `exploration-database-migration.md`

3. **Strategy Documents**: `strategy-{topic}.md`
   - Examples: `strategy-scalability.md`, `strategy-security.md`

4. **Sync Reports**: `sync-report-{YYYYMMDD}.md`
   - Examples: `sync-report-20251105.md`, `sync-report-20251104.md`

5. **Analysis Reports**: `{topic}-analysis.md`
   - Examples: `performance-analysis.md`, `security-analysis.md`

6. **Guide Documents**: `guide-{topic}.md`
   - Examples: `guide-alfred-workflows.md`, `guide-spec-creation.md`

### SPEC Directory Naming

```
SPEC-{category}-{number}-{description}
Examples:
- SPEC-FE-001-user-authentication
- SPEC-BE-002-api-rate-limiting
- SPEC-INF-003-container-orchestration
```

## Decision Tree Logic

### Document Creation Flowchart

```python
def determine_document_location(doc_type, user_request=None):
    """
    Determine correct location and approval requirements for document creation
    """
    
    # Step 1: Check if user explicitly requested
    if user_request:
        return handle_explicit_request(doc_type, user_request)
    
    # Step 2: Check if this is part of automated workflow
    if is_workflow_automated(doc_type):
        return handle_automated_workflow(doc_type)
    
    # Step 3: Require user approval
    return request_user_approval(doc_type)

def handle_explicit_request(doc_type, request):
    """User explicitly requested document creation"""
    location = get_standard_location(doc_type)
    return {
        "action": "create",
        "location": location,
        "approval": "granted",
        "reason": "user_explicit_request"
    }

def handle_automated_workflow(doc_type):
    """Document is part of standard workflow"""
    if doc_type in ["spec", "sync_report"]:
        return {
            "action": "auto_create",
            "location": get_standard_location(doc_type),
            "approval": "not_required",
            "reason": "standard_workflow"
        }
    else:
        return request_user_approval(doc_type)

def request_user_approval(doc_type):
    """Ask user before creating document"""
    location = get_standard_location(doc_type)
    
    question = f"Create {doc_type} at {location}?"
    options = [
        "YES - Create the document",
        "NO - Skip documentation"
    ]
    
    return AskUserQuestion(question=question, options=options)
```

## Integration Points

### Alfred Command Integration

| Command | Document Creation | Location | Auto-Approved? |
|---------|-------------------|----------|----------------|
| `/alfred:0-project` | Project setup documentation | `.moai/docs/` | ❌ Ask user |
| `/alfred:1-plan` | SPEC documents | `.moai/specs/SPEC-*/` | ✅ Auto-create |
| `/alfred:2-run` | Implementation guides (optional) | `.moai/docs/` | ❌ Ask user |
| `/alfred:3-sync` | Sync reports | `.moai/reports/` | ✅ Auto-create |

### Sub-Agent Integration

| Sub-Agent | Default Output | File Pattern | Location |
|-----------|----------------|--------------|----------|
| `implementation-planner` | Implementation guide | `implementation-{SPEC}.md` | `.moai/docs/` |
| `Explore` | Exploration report | `exploration-{topic}.md` | `.moai/docs/` |
| `Plan` | Strategy document | `strategy-{topic}.md` | `.moai/docs/` |
| `doc-syncer` | Sync report | `sync-report-{type}.md` | `.moai/reports/` |
| `tag-agent` | Tag validation | `tag-validation-{date}.md` | `.moai/reports/` |
| `spec-builder` | SPEC package | `spec.md`, `plan.md`, `acceptance.md` | `.moai/specs/SPEC-*/` |
| `tdd-implementer` | Implementation notes | `implementation-{SPEC}.md` | `.moai/docs/` |

## Validation Rules

### Pre-creation Validation

```python
def validate_document_creation(location, content, doc_type):
    """Validate document before creation"""
    
    # Rule 1: Check location is allowed
    if not is_location_allowed(location):
        raise ValueError(f"Document location {location} is not allowed for {doc_type}")
    
    # Rule 2: Check filename follows convention
    if not follows_naming_convention(location):
        raise ValueError(f"Filename {location} does not follow naming conventions")
    
    # Rule 3: Check directory structure exists
    ensure_directory_structure(location)
    
    # Rule 4: Validate content type matches location
    if not content_matches_location_type(content, doc_type, location):
        raise ValueError(f"Content type doesn't match location {location}")
    
    return True
```

### Post-creation Validation

```python
def validate_created_document(file_path):
    """Validate document after creation"""
    
    # Check file exists
    if not os.path.exists(file_path):
        return False, "File was not created"
    
    # Check file has content
    if os.path.getsize(file_path) == 0:
        return False, "File is empty"
    
    # Check file follows expected structure
    with open(file_path, 'r') as f:
        content = f.read()
    
    if not has_proper_structure(content, file_path):
        return False, "File structure is invalid"
    
    return True, "Document created successfully"
```
