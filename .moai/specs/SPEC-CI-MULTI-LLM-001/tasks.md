---
spec_id: SPEC-CI-MULTI-LLM-001
version: 1.0.0
created: 2026-04-27
updated: 2026-04-27
---

# SPEC-CI-MULTI-LLM-001 — Task Decomposition

## Wave 1 — CLI Skeleton + Runner Domain (M1)

| Task ID | Description | Requirements | Dependencies | Planned Files | Status |
|---------|-------------|-------------|-------------|---------------|--------|
| T-01 | GitHub Cobra Command Group Extension | REQ-CI-001, CI-001.1, CI-001.2, CI-001.3 | None | `internal/cli/github.go` (수정), `internal/cli/github_test.go` (신규) | pending |
| T-02 | Runner Installer Domain | REQ-CI-003, CI-003.1, CI-003.2, CI-003.3 | T-01 | `internal/github/runner/installer.go`, `internal/github/runner/installer_test.go` | pending |
| T-03 | Runner Registrar Domain | REQ-CI-004, CI-004.1, CI-004.2, CI-004.3 | T-02 | `internal/github/runner/registrar.go`, `internal/github/runner/registrar_test.go` | pending |
| T-04 | Runner Service Manager (launchd + systemd stub) | REQ-CI-003.4, CI-005.3 | T-03 | `internal/github/runner/service.go`, `internal/github/runner/service_test.go` | pending |
| T-05 | Runner Version Checker | REQ-CI-005, CI-005.1, CI-005.2 | T-02 | `internal/github/runner/version.go`, `internal/github/runner/version_test.go` | pending |
| T-06 | Runner CLI Subcommands (Thin Cobra Wiring) | REQ-CI-001 (runner subtree) | T-01~T-05 | `internal/cli/github_runner.go`, `internal/cli/github_runner_test.go` | pending |

## Wave 2 — LLM Auth Bootstrap (M2)

| Task ID | Description | Requirements | Dependencies | Planned Files | Status |
|---------|-------------|-------------|-------------|---------------|--------|
| T-07 | Secret Manager (gh secret Wrapper) | REQ-SEC-002 | T-01 | `internal/github/secret.go`, `internal/github/secret_test.go` | pending |
| T-08 | Claude Auth Handler | REQ-CI-008, CI-008.1, CI-008.2 | T-07 | `internal/github/auth/claude.go`, `internal/github/auth/claude_test.go` | pending |
| T-09 | Codex Auth Handler (Private Guard) | REQ-CI-009, CI-009.1, CI-009.2, REQ-CI-007, CI-007.1, CI-007.2, REQ-SEC-001 | T-07 | `internal/github/auth/codex.go`, `internal/github/auth/codex_test.go` | pending |
| T-10 | Gemini Auth Handler | REQ-CI-010, CI-010.1, CI-010.2 | T-07 | `internal/github/auth/gemini.go`, `internal/github/auth/gemini_test.go` | pending |
| T-11 | GLM Auth Handler | REQ-CI-011, CI-011.1, CI-011.2 | T-07 | `internal/github/auth/glm.go`, `internal/github/auth/glm_test.go` | pending |

## Wave 3 — Workflow Templates + Composite Actions (M3)

| Task ID | Description | Requirements | Dependencies | Planned Files | Status |
|---------|-------------|-------------|-------------|---------------|--------|
| T-12 | detect-language Composite Action | REQ-CI-012, CI-012.2, CI-012.3 | None | `internal/template/templates/.github/actions/detect-language/action.yml.tmpl`, `internal/github/workflow/detect_language_test.go` | pending |
| T-13 | setup-glm-env Composite Action | REQ-CI-011.1, CI-011.2 | T-11 | `internal/template/templates/.github/actions/setup-glm-env/action.yml.tmpl` | pending |
| T-14 | codex-bootstrap Composite Action | REQ-CI-009.1, CI-021, REQ-SEC-006 | T-09 | `internal/template/templates/.github/actions/codex-bootstrap/action.yml.tmpl` | pending |
| T-15 | claude.yml.tmpl (Issue Trigger) | REQ-CI-014, SEC-009, SEC-005, SEC-010 | T-12 | `internal/template/templates/.github/workflows/claude.yml.tmpl` | pending |
| T-16 | claude-code-review.yml.tmpl (PR Trigger) | REQ-CI-014, CI-018, CI-018.1, SEC-005, SEC-010 | T-12 | `internal/template/templates/.github/workflows/claude-code-review.yml.tmpl` | pending |
| T-17 | codex-review.yml.tmpl | REQ-CI-007, SEC-001, SEC-006, SEC-005, SEC-009, SEC-010 | T-12, T-14 | `internal/template/templates/.github/workflows/codex-review.yml.tmpl` | pending |
| T-18 | gemini-review.yml.tmpl | REQ-CI-014, CI-006, CI-006.1, SEC-005, SEC-009, SEC-010 | T-12 | `internal/template/templates/.github/workflows/gemini-review.yml.tmpl` | pending |
| T-19 | glm-review.yml.tmpl | REQ-CI-014, SEC-005, SEC-009, SEC-010 | T-12, T-13 | `internal/template/templates/.github/workflows/glm-review.yml.tmpl` | pending |
| T-20 | llm-panel.yml.tmpl (PR Open Auto Panel) | REQ-CI-013, CI-013.1~013.4, CI-015, SEC-010 | T-15~T-19 | `internal/template/templates/.github/workflows/llm-panel.yml.tmpl` | pending |
| T-21 | github-actions.yaml.tmpl (Config Template) | REQ-CI-015, CI-020, CI-020.1 | T-01 | `internal/template/templates/.moai/config/sections/github-actions.yaml.tmpl` | pending |
| T-22 | Template Validator | REQ-CI-018, CI-018.1, CI-021, SEC-003, SEC-005 | T-15~T-19 | `internal/github/workflow/validator.go`, `internal/github/workflow/validator_test.go` | pending |

## Wave 4 — Integration + Docs (M4)

| Task ID | Description | Requirements | Dependencies | Planned Files | Status |
|---------|-------------|-------------|-------------|---------------|--------|
| T-23 | GitHub Init Command (Integrated Bootstrap) | REQ-CI-002, CI-002.1~002.3 | T-06, T-07~T-11, T-21 | `internal/cli/github_init.go`, `internal/cli/github_init_test.go` | pending |
| T-24 | GitHub Auth Command (Thin Wrapper) | REQ-CI-008~011 | T-08~T-11 | `internal/cli/github_auth.go`, `internal/cli/github_auth_test.go` | pending |
| T-25 | GitHub Workflow Command | REQ-CI-016 | T-22, T-21 | `internal/cli/github_workflow.go`, `internal/cli/github_workflow_test.go` | pending |
| T-26 | GitHub Status Command | REQ-CI-017, SEC-004 | T-05, T-07, T-22 | `internal/cli/github_status.go`, `internal/cli/github_status_test.go` | pending |
| T-27 | Doctor Integration (Runner Version Check) | REQ-CI-005 | T-05 | `internal/cli/doctor.go` (수정), `internal/cli/doctor_test.go` (수정) | pending |
| T-28 | SessionStart Hook Integration | REQ-CI-005 | T-05 | `internal/cli/hook.go` (수정), `internal/cli/hook_test.go` (수정) | pending |
| T-29 | Network Egress Validation + Audit Log Setup | REQ-SEC-007, CI-007.1, CI-007.2, SEC-008 | T-02, T-11 | `internal/github/runner/egress.go`, `internal/github/runner/egress_test.go` | pending |
| T-30 | docs-site 4-Locale Guides | REQ-CI-012, CLAUDE.local.md §17 | T-23 | `docs-site/content/{ko,en,ja,zh}/guides/multi-llm-ci.md` | pending |

---

## REQ Coverage Matrix

| REQ ID | Tasks | Covered |
|--------|-------|---------|
| REQ-CI-001 | T-01, T-06 | YES |
| REQ-CI-002 | T-23 | YES |
| REQ-CI-003 | T-02, T-04 | YES |
| REQ-CI-004 | T-03 | YES |
| REQ-CI-005 | T-05, T-27, T-28 | YES |
| REQ-CI-006 | T-18, T-21 | YES |
| REQ-CI-007 | T-09, T-17 | YES |
| REQ-CI-008 | T-08, T-24 | YES |
| REQ-CI-009 | T-09, T-14 | YES |
| REQ-CI-010 | T-10, T-24 | YES |
| REQ-CI-011 | T-11, T-13, T-24 | YES |
| REQ-CI-012 | T-12, T-30 | YES |
| REQ-CI-013 | T-20 | YES |
| REQ-CI-014 | T-15, T-16, T-17, T-18, T-19 | YES |
| REQ-CI-015 | T-21, T-20 | YES |
| REQ-CI-016 | T-25 | YES |
| REQ-CI-017 | T-26 | YES |
| REQ-CI-018 | T-16, T-22 | YES |
| REQ-CI-019 | T-12~T-22 (all templates) | YES |
| REQ-CI-020 | T-21 | YES |
| REQ-CI-021 | T-14, T-22 | YES |
| REQ-SEC-001 | T-09, T-17 | YES |
| REQ-SEC-002 | T-07 | YES |
| REQ-SEC-003 | T-22 | YES |
| REQ-SEC-004 | T-26 | YES |
| REQ-SEC-005 | T-15~T-19, T-22 | YES |
| REQ-SEC-006 | T-14, T-17 | YES |
| REQ-SEC-007 | T-29 | YES |
| REQ-SEC-008 | T-29 | YES |
| REQ-SEC-009 | T-15~T-19 (acceptance criteria) | YES |
| REQ-SEC-010 | T-15~T-20 (acceptance criteria) | YES |

**Coverage: 31/31 REQs (100%)**

---

## File Count Summary

| Category | New Files | Modified Files | Total |
|----------|-----------|----------------|-------|
| Go source (internal/github/) | 15 | 0 | 15 |
| Go source (internal/cli/) | 8 | 4 | 12 |
| Go test files | 14 | 4 | 18 |
| Workflow templates (.yml.tmpl) | 6 | 0 | 6 |
| Composite actions | 3 | 0 | 3 |
| Config templates | 1 | 0 | 1 |
| docs-site (4 locales) | 4 | 0 | 4 |
| **Total** | **51** | **8** | **59** |
