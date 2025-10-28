# {{PROJECT_NAME}} GitFlow PR

> Full GitFlow Transparency — the agent auto-fills information

## 📝 SPEC Info

- Related SPEC: `SPEC-AUTH-001` (e.g., JWT authentication system)
- Directory: `{{SPEC_DIR}}/SPEC-AUTH-001/`
{% if ENABLE_TAG_SYSTEM -%}
- @TAG Links: @SPEC:AUTH-001 @CODE:AUTH-001 (auto-tagging)
{% endif -%}

## ✅ SPEC Quality Checks

- [ ] YAML Front Matter: 7 required fields (id, version, status, created, updated, author, priority)
- [ ] HISTORY Section: Record versioned change log (include v0.0.1 INITIAL)
- [ ] EARS Requirements: Ubiquitous, Event-driven, State-driven, Optional, Constraints
{% if ENABLE_TAG_SYSTEM -%}
- [ ] @SPEC:ID TAG: Include TAG in doc and check duplicates (`rg "@SPEC:<ID>" -n`)
{% endif -%}

{% if ENABLE_ALFRED_COMMANDS -%}
## 🤖 Automated Validation Status

<!-- The checklist below is auto-updated by the agent -->
<!-- /alfred:1-plan → create feature branch → Draft PR -->
<!-- /alfred:2-run → implement via TDD → auto-check checkboxes -->
<!-- /alfred:3-sync → synchronize documents → Ready for Review -->

- [ ] spec-builder: Complete EARS spec and create feature branch
- [ ] code-builder: Finish TDD RED-GREEN-REFACTOR
- [ ] doc-syncer: Sync Living Documents and mark PR Ready
{% endif -%}

{% if ENABLE_TRUST_5 -%}
## 📊 Quality Metrics (auto-calculated)

- TRUST 5 Principles: ✅ Compliant
- Test Coverage: XX% (target ≥ 85%)
{% if ENABLE_TAG_SYSTEM -%}
- @TAG Traceability: 100%
{% endif -%}
{% endif -%}

## 🌍 Locale Settings

- Project Language: {{CONVERSATION_LANGUAGE}}
- Commit Messages: <!-- generated automatically according to locale -->

## 🎯 Changes

<!-- auto-fills TDD results -->

### 🔴 RED (Test Authoring)
- Test File: `{{TEST_DIR}}/auth/service.test.ts`
- Test Description: [describe the failing test]

### 🟢 GREEN (Implementation)
- Implementation File: `src/auth/service.ts`
- Implementation Done: [describe functionality]

### ♻️ REFACTOR (Improvements)
- Refactoring Details: [code quality improvements]

## 📚 Documentation Sync

<!-- auto-filled by doc-syncer -->

- [ ] Update README
- [ ] Sync API docs
{% if ENABLE_TAG_SYSTEM -%}
- [ ] Update TAG index
{% endif -%}
- [ ] Update HISTORY section (SPEC docs)

---

🚀 {{PROJECT_NAME}}: Professional development via a 3-stage pipeline!

Reviewers: Check quality compliance and SPEC metadata completeness only.

