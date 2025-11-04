---
title: Dynamic Document Reference for spec-builder
id: SPEC-FEATURE-001
type: feature
priority: medium
status: closed
resolution: replaced_by_skill
closed_date: 2025-11-04
affects: spec-builder agent, project structure
discussion: #130
---

# SPEC-FEATURE-001: Dynamic Document Reference System

## HISTORY

### v1.0.0 (2025-11-04) - CLOSED
- **Status**: Closed
- **Reason**: Dynamic Document Reference 기능이 Skill 시스템(`moai-foundation-product`, `moai-foundation-structure`)으로 대체됨
- **Resolution**: Skill 기반 아키텍처가 더 유연한 문서 참조 제공으로 자연스럽게 해결됨
- **Author**: @Goos
- **Note**: 추가 문서 참조는 `/alfred:1-plan` 실행 시 spec-builder가 자동으로 `.moai/project/` 디렉토리 문서 스캔

---

## Problem Statement

**Issue**: spec-builder only references 3 hardcoded files:
- `.moai/project/product.md`
- `.moai/project/structure.md`
- `.moai/project/tech.md`

**Limitation**: Cannot reference additional project documents like:
- `ux-journey.md` (UX/UI requirements)
- `testing-strategy.md` (QA requirements)
- `api-spec.md` (API requirements)
- `deployment-plan.md` (DevOps requirements)

**Impact**: Users must manually merge UX/UI/other requirements into existing documents, reducing flexibility.

**Severity**: Medium (feature enhancement)

## User Story

As a **product manager working on UX-heavy projects**,
I want to **create separate documents for different requirement types**,
So that **spec-builder automatically references all relevant context when generating SPECs**.

### Example Use Case

```
.moai/project/
├── product.md          # Business requirements
├── structure.md        # System architecture
├── tech.md             # Technology stack
├── ux-journey.md       # User journey maps (NEW)
├── testing-strategy.md # QA requirements (NEW)
└── api-spec.md         # API contracts (NEW)
```

When running `/alfred:1-plan "Onboarding UX"`, spec-builder should automatically reference:
- `product.md` (business context)
- `structure.md` (system context)
- `ux-journey.md` (UX requirements) ← NEW!

## Solution

### Approach: Dynamic Document Scanning

Replace hardcoded file list with **automatic scanning** of `.moai/project/*.md` files.

### Design Principles

1. **Convention over Configuration**: Auto-discover all `.md` files
2. **Prioritization**: Known files (product/structure/tech) have priority
3. **Extensibility**: Users can add any number of documents
4. **Backward Compatibility**: Existing 3-file setup still works

### Implementation Design

#### 1. Update spec-builder Agent Instructions

**File**: `src/moai_adk/templates/.claude/agents/alfred/spec-builder.md`

**Current**:
```markdown
## Reference Documents (Auto-loaded)

When creating SPECs, always read these files:
1. `.moai/project/product.md` - Business requirements
2. `.moai/project/structure.md` - System architecture
3. `.moai/project/tech.md` - Technology stack
```

**Updated**:
```markdown
## Reference Documents (Auto-discovered)

When creating SPECs, automatically scan and read:
1. **Priority documents** (always read first):
   - `.moai/project/product.md` - Business requirements
   - `.moai/project/structure.md` - System architecture
   - `.moai/project/tech.md` - Technology stack

2. **Additional documents** (context-specific):
   - `.moai/project/ux-journey.md` - UX/UI requirements
   - `.moai/project/testing-strategy.md` - QA requirements
   - `.moai/project/api-spec.md` - API contracts
   - `.moai/project/deployment-plan.md` - DevOps requirements
   - `.moai/project/*.md` - Any other markdown files

### Document Discovery Strategy

1. **Scan directory**: Read all `.md` files in `.moai/project/`
2. **Prioritize**: Load priority documents first (product → structure → tech)
3. **Categorize**: Infer document purpose from filename/content
4. **Contextualize**: Use relevant documents based on SPEC topic

### Filename Conventions

Suggested naming patterns for auto-categorization:
- `ux-*.md`, `ui-*.md` → UX/UI requirements
- `test-*.md`, `qa-*.md` → Testing requirements
- `api-*.md` → API specifications
- `deploy-*.md`, `ops-*.md` → DevOps/deployment
- `security-*.md` → Security requirements
- `compliance-*.md` → Compliance/legal requirements

### Example Workflow

User request: `/alfred:1-plan "User onboarding flow"`

spec-builder's document loading:
1. ✅ Read `product.md` (business context)
2. ✅ Read `structure.md` (system context)
3. ✅ Read `tech.md` (technical stack)
4. ✅ Detect `ux-journey.md` exists → Read UX requirements
5. ⏭️  Skip `api-spec.md` (not relevant to UX)
6. ⏭️  Skip `deployment-plan.md` (not relevant to UX)

Result: SPEC generated with full UX context!
```

#### 2. Create Document Scanner Utility

**File**: `src/moai_adk/core/project/document_scanner.py`

```python
"""Dynamic document scanner for spec-builder context loading."""

from pathlib import Path
from typing import Dict, List


class ProjectDocumentScanner:
    """Scan and categorize project documentation for spec-builder."""

    # Priority order for known documents
    PRIORITY_DOCS = [
        "product.md",
        "structure.md",
        "tech.md",
    ]

    # Document type categorization (filename patterns)
    CATEGORIES = {
        "ux": ["ux-", "ui-", "user-", "journey-"],
        "testing": ["test-", "qa-", "quality-"],
        "api": ["api-", "endpoint-", "rest-"],
        "devops": ["deploy-", "ops-", "infra-", "ci-"],
        "security": ["security-", "auth-", "permission-"],
        "compliance": ["compliance-", "legal-", "gdpr-"],
    }

    def __init__(self, project_path: Path):
        """Initialize scanner.

        Args:
            project_path: Project root directory
        """
        self.project_path = project_path
        self.docs_dir = project_path / ".moai" / "project"

    def scan_documents(self) -> Dict[str, List[Path]]:
        """Scan all markdown files in .moai/project/.

        Returns:
            Dictionary mapping categories to file paths:
            {
                "priority": [product.md, structure.md, tech.md],
                "ux": [ux-journey.md],
                "testing": [testing-strategy.md],
                ...
            }
        """
        if not self.docs_dir.exists():
            return {"priority": []}

        result = {"priority": []}

        # 1. Collect priority documents first
        for doc_name in self.PRIORITY_DOCS:
            doc_path = self.docs_dir / doc_name
            if doc_path.exists():
                result["priority"].append(doc_path)

        # 2. Categorize additional documents
        for doc_path in self.docs_dir.glob("*.md"):
            # Skip priority documents (already processed)
            if doc_path.name in self.PRIORITY_DOCS:
                continue

            # Categorize by filename pattern
            categorized = False
            for category, patterns in self.CATEGORIES.items():
                if any(doc_path.name.startswith(p) for p in patterns):
                    if category not in result:
                        result[category] = []
                    result[category].append(doc_path)
                    categorized = True
                    break

            # Uncategorized documents go to "other"
            if not categorized:
                if "other" not in result:
                    result["other"] = []
                result["other"].append(doc_path)

        return result

    def get_relevant_docs(self, topic: str) -> List[Path]:
        """Get relevant documents for a given SPEC topic.

        Args:
            topic: SPEC topic/title (e.g., "User onboarding UX")

        Returns:
            List of relevant document paths (priority docs always included)
        """
        all_docs = self.scan_documents()
        relevant = []

        # Always include priority documents
        relevant.extend(all_docs.get("priority", []))

        # Detect topic category and include relevant docs
        topic_lower = topic.lower()

        if any(keyword in topic_lower for keyword in ["ux", "ui", "user", "onboard", "journey"]):
            relevant.extend(all_docs.get("ux", []))

        if any(keyword in topic_lower for keyword in ["test", "qa", "quality"]):
            relevant.extend(all_docs.get("testing", []))

        if any(keyword in topic_lower for keyword in ["api", "endpoint", "rest"]):
            relevant.extend(all_docs.get("api", []))

        if any(keyword in topic_lower for keyword in ["deploy", "infra", "ci", "cd"]):
            relevant.extend(all_docs.get("devops", []))

        if any(keyword in topic_lower for keyword in ["security", "auth", "permission"]):
            relevant.extend(all_docs.get("security", []))

        # If no specific category detected, include all
        if len(relevant) == len(all_docs.get("priority", [])):
            for docs in all_docs.values():
                relevant.extend(docs)
            relevant = list(set(relevant))  # Deduplicate

        return relevant
```

#### 3. Update /alfred:1-plan Command

**File**: `src/moai_adk/templates/.claude/commands/alfred/1-plan.md`

Add document scanning step:

```markdown
## Phase 1: Context Loading

Before creating SPECs, load all relevant project documents:

1. **Scan documents**:
   ```python
   from moai_adk.core.project.document_scanner import ProjectDocumentScanner

   scanner = ProjectDocumentScanner(project_path)
   relevant_docs = scanner.get_relevant_docs(spec_title)
   ```

2. **Load documents**:
   - Read each document in order (priority first)
   - Extract key requirements
   - Build context for spec-builder

3. **Pass to spec-builder**:
   - Send all loaded documents as context
   - spec-builder generates SPEC with full context
```

## Testing Strategy

### Unit Tests

**File**: `tests/unit/test_document_scanner.py`

```python
import pytest
from pathlib import Path
from moai_adk.core.project.document_scanner import ProjectDocumentScanner


def test_scan_priority_documents(tmp_path):
    """Test scanning priority documents."""
    docs_dir = tmp_path / ".moai" / "project"
    docs_dir.mkdir(parents=True)

    (docs_dir / "product.md").write_text("# Product")
    (docs_dir / "structure.md").write_text("# Structure")
    (docs_dir / "tech.md").write_text("# Tech")

    scanner = ProjectDocumentScanner(tmp_path)
    result = scanner.scan_documents()

    assert len(result["priority"]) == 3
    assert all(doc.exists() for doc in result["priority"])


def test_scan_ux_documents(tmp_path):
    """Test scanning UX documents."""
    docs_dir = tmp_path / ".moai" / "project"
    docs_dir.mkdir(parents=True)

    (docs_dir / "ux-journey.md").write_text("# UX Journey")
    (docs_dir / "ui-components.md").write_text("# UI Components")

    scanner = ProjectDocumentScanner(tmp_path)
    result = scanner.scan_documents()

    assert "ux" in result
    assert len(result["ux"]) == 2


def test_get_relevant_docs_for_ux_topic(tmp_path):
    """Test retrieving relevant docs for UX topic."""
    docs_dir = tmp_path / ".moai" / "project"
    docs_dir.mkdir(parents=True)

    (docs_dir / "product.md").write_text("# Product")
    (docs_dir / "ux-journey.md").write_text("# UX Journey")
    (docs_dir / "api-spec.md").write_text("# API Spec")

    scanner = ProjectDocumentScanner(tmp_path)
    relevant = scanner.get_relevant_docs("User onboarding UX")

    # Should include product.md (priority) + ux-journey.md
    assert len(relevant) >= 2
    assert any("product.md" in str(doc) for doc in relevant)
    assert any("ux-journey.md" in str(doc) for doc in relevant)
    # Should NOT include api-spec.md (not relevant to UX)
    assert not any("api-spec.md" in str(doc) for doc in relevant)
```

### Integration Tests

**Test Scenarios**:

1. **Fresh Project**: Only priority docs exist
   - Expected: spec-builder references 3 priority docs

2. **UX Project**: Priority + ux-journey.md exist
   - Expected: spec-builder references 4 docs for UX SPECs

3. **Full Project**: All document types exist
   - Expected: spec-builder selectively references relevant docs

### Manual Testing

```bash
# 1. Create test documents
mkdir -p .moai/project
echo "# Product Requirements" > .moai/project/product.md
echo "# UX Journey" > .moai/project/ux-journey.md
echo "# API Spec" > .moai/project/api-spec.md

# 2. Run /alfred:1-plan with UX topic
/alfred:1-plan "User onboarding flow"

# 3. Verify spec-builder loaded UX documents
# Expected output: "Loaded 4 documents: product.md, structure.md, tech.md, ux-journey.md"
```

## Acceptance Criteria

- [ ] spec-builder auto-discovers all `.md` files in `.moai/project/`
- [ ] Priority documents (product/structure/tech) always loaded first
- [ ] Document categorization by filename pattern works
- [ ] Relevant documents selected based on SPEC topic
- [ ] Backward compatibility: existing 3-file setup still works
- [ ] Unit tests pass for all scanning scenarios
- [ ] Integration tests pass
- [ ] Documentation updated with filename conventions

## Implementation Plan

### Phase 1: Design (Day 1)
- Finalize filename convention patterns
- Design document scanner API
- Write test cases

### Phase 2: Core Scanner (Day 2)
- Implement `ProjectDocumentScanner` class
- Add categorization logic
- Write unit tests

### Phase 3: Agent Integration (Day 3)
- Update spec-builder agent instructions
- Update /alfred:1-plan command
- Add document loading step

### Phase 4: Testing (Day 4-5)
- Write integration tests
- Manual testing with various document setups
- Performance testing (large number of docs)

### Phase 5: Documentation (Day 6)
- Update README.md with document conventions
- Add examples for each document type
- Create migration guide
- Release v0.11.0

## User Documentation

### Recommended Document Structure

```
.moai/project/
├── product.md              # Business requirements (always loaded)
├── structure.md            # System architecture (always loaded)
├── tech.md                 # Technology stack (always loaded)
├── ux-journey.md           # User journey maps and UX flows
├── ui-components.md        # UI component library and design system
├── testing-strategy.md     # QA strategy and test plans
├── api-spec.md             # API contracts and endpoints
├── deployment-plan.md      # Deployment strategy and infrastructure
├── security-requirements.md # Security and compliance requirements
└── performance-goals.md    # Performance metrics and SLAs
```

### Filename Conventions

Follow these naming patterns for automatic categorization:

| Category | Prefix Examples | Purpose |
|----------|----------------|---------|
| UX/UI | `ux-`, `ui-`, `user-`, `journey-` | User experience and interface |
| Testing | `test-`, `qa-`, `quality-` | Quality assurance and testing |
| API | `api-`, `endpoint-`, `rest-` | API specifications |
| DevOps | `deploy-`, `ops-`, `infra-`, `ci-` | Deployment and operations |
| Security | `security-`, `auth-`, `permission-` | Security requirements |
| Compliance | `compliance-`, `legal-`, `gdpr-` | Legal and compliance |

## Risks and Mitigations

### Risk 1: Too many documents slow down spec-builder
**Impact**: Slow SPEC generation
**Mitigation**: Selective loading based on topic relevance

### Risk 2: Document categorization fails
**Impact**: Relevant docs not loaded
**Mitigation**: Fallback to loading all docs if no category match

### Risk 3: Breaking change for existing projects
**Impact**: Users need to adapt
**Mitigation**: Backward compatible - 3-file setup still works

## Related Issues

- Discussion #130: UX workflow support request
- Related to spec-builder agent design

## Success Metrics

- Users can add custom document types without framework changes
- SPEC generation time < 30 seconds (even with 10+ documents)
- Positive feedback on UX workflow support
- Adoption rate of additional document types

---

**Author**: debug-helper
**Created**: 2025-10-30
**Target Release**: v0.11.0
