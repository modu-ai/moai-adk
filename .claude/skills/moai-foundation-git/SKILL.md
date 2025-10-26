---
name: moai-foundation-git
version: 2.0.0
created: 2025-10-22
updated: 2025-10-22
status: active
description: GitFlow automation and PR policy enforcement for MoAI-ADK workflows.
keywords: ['git', 'gitflow', 'pr', 'automation']
allowed-tools:
  - Read
  - Bash
---

# Foundation Git Skill

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-foundation-git |
| **Version** | 2.0.0 (2025-10-22) |
| **Allowed tools** | Read (read_file), Bash (terminal) |
| **Auto-load** | On demand when keywords detected |
| **Tier** | Foundation |

---

## What It Does

GitFlow automation and PR policy enforcement for MoAI-ADK workflows.

**Key capabilities**:
- ✅ Best practices enforcement for foundation domain
- ✅ TRUST 5 principles integration
- ✅ Latest tool versions (2025-10-22)
- ✅ TDD workflow support

---

## When to Use

**Automatic triggers**:
- Related code discussions and file patterns
- SPEC implementation (`/alfred:2-run`)
- Code review requests

**Manual invocation**:
- Review code for TRUST 5 compliance
- Design new features
- Troubleshoot issues

---

## Tool Version Matrix (2025-10-22)

| Tool | Version | Purpose | Status |
|------|---------|---------|--------|
| **Git** | 2.47.0 | Primary | ✅ Current |
| **GitHub CLI** | 2.63.0 | Primary | ✅ Current |

---

## Inputs

- Language-specific source directories
- Configuration files
- Test suites and sample data

## Outputs

- Test/lint execution plan
- TRUST 5 review checkpoints
- Migration guidance

## Failure Modes

- When required tools are not installed
- When dependencies are missing
- When test coverage falls below 85%

## Dependencies

- Access to project files via Read/Bash tools
- Integration with `moai-foundation-langs` for language detection
- Integration with `moai-foundation-trust` for quality gates

---

## References (Latest Documentation)

_Documentation links updated 2025-10-22_

---

## Changelog

- **v2.0.0** (2025-10-22): Major update with latest tool versions, comprehensive best practices, TRUST 5 integration
- **v1.0.0** (2025-03-29): Initial Skill release

---

## Works Well With

- `moai-foundation-trust` (quality gates)
- `moai-alfred-code-reviewer` (code review)
- `moai-essentials-debug` (debugging support)

---

## Best Practices

✅ **DO**:
- Follow foundation best practices
- Use latest stable tool versions
- Maintain test coverage ≥85%
- Document all public APIs

❌ **DON'T**:
- Skip quality gates
- Use deprecated tools
- Ignore security warnings
- Mix testing frameworks

---

## 🤖 Alfred Git Commit Signature Standard

**All Git commits and operations follow this signature**:

```
🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: 🎩 Alfred@[MoAI](https://adk.mo.ai.kr)
```

**Applied to**:
- ✅ Feature commits (TDD RED → GREEN → REFACTOR)
- ✅ Release commits (version bumps, tags)
- ✅ Hotfix commits (urgent bug fixes)
- ✅ Merge commits (PR merges, branch integrations)
- ✅ All Git automation operations

**Rationale**: Alfred is the MoAI SuperAgent orchestrating all Git operations in MoAI-ADK, ensuring accountability and traceability across all development workflows.
