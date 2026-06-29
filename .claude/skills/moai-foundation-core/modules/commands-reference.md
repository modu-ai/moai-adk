# Commands Reference - MoAI-ADK Core Commands

Purpose: Complete reference for MoAI-ADK's 6 core commands used in SPEC-First DDD workflow.

Last Updated: 2025-11-25
Version: 2.0.0

---

## Quick Reference (30 seconds)

MoAI-ADK provides 6 core commands for SPEC-First DDD execution:

| Command            | Purpose                | Phase         |
| ------------------ | ---------------------- | ------------- |
| `/moai project`  | Project initialization | Setup         |
| `/moai plan`     | SPEC generation        | Planning      |
| `/moai run`      | DDD implementation     | Development   |
| `/moai sync`     | Documentation sync     | Documentation |
| `/moai feedback` | Feedback collection    | Improvement   |

Required Workflow:
```
1. /moai project # Initialize
2. /moai plan "description" # Generate SPEC
3. /clear # Clear context (REQUIRED)
4. /moai run SPEC-001 # Implement
5. /moai sync SPEC-001 # Document
6. /moai feedback # Improve
```

Critical Rule: Execute `/clear` after `/moai plan` (saves 45-50K tokens)

---

## Implementation Guide (5 minutes)

### `/moai project` - Project Initialization

Purpose: Initialize project structure and generate configuration

Agent Delegation: `manager-docs` (project-doc scaffolding)

Usage:
```bash
/moai project
/moai project --with-git
```

What It Does:
1. Creates `.moai/` directory structure
2. Generates `config.json` with default settings
3. Initializes Git repository (if `--with-git` flag provided)
4. Sets up MoAI-ADK workflows

Output:
- `.moai/` directory
- `.moai/config/config.yaml`
- `.moai/state/` (empty, ready for session state)
- `.moai/logs/` (empty, ready for logging)

Next Step: Ready for SPEC generation via `/moai plan`

Example:
```
User: /moai project
MoAI: Project initialized successfully.
 - .moai/config/config.yaml created
 - Git workflow set to 'manual' mode
 Ready for SPEC generation.
```

---

### `/moai plan` - SPEC Generation

Purpose: Generate SPEC document in GEARS format (current; EARS retained as legacy reference for the 6-month backward-compat window)

Agent Delegation: `manager-spec`

Usage:
```bash
/moai plan "Implement user authentication endpoint (JWT)"
/moai plan "Add dark mode toggle to settings page"
```

What It Does:
1. Analyzes user request
2. Generates GEARS format SPEC document (current; EARS for legacy SPECs)
3. Creates `.moai/specs/SPEC-XXX/` directory
4. Saves `spec.md` with requirements

GEARS Format (5 patterns; current notation):
- Ubiquitous: "The <subject> shall <behavior>"
- Event-driven: "When <event>, the <subject> shall <behavior>"
- State-driven: "While <state>, the <subject> shall <behavior>"
- Where (capability gate): "Where <capability or feature flag>, the <subject> shall <behavior>"
- Event-detected (replaces the deprecated conditional modality): "When <undesired-condition-detected>, the <subject> shall <response>"

Unified compound clause: `[Where ...][While ...][When ...] The <subject> shall <behavior>` — any subset may chain. `<subject>` is generalized (any noun: system, component, service, agent, function, artifact). See `.claude/skills/moai-workflow-spec/SKILL.md` § "GEARS Format" for the canonical authoring guide.

EARS Format (legacy reference, 6-month backward-compat — expires 2026-11-22):
- WHEN (trigger conditions)
- WHILE (state-driven conditions; preferred over the deprecated conditional modality)
- THE SYSTEM SHALL (functional requirements)
- WHERE (constraints)
- UBIQUITOUS (quality requirements)

Output:
- `.moai/specs/SPEC-001/spec.md` (GEARS document; EARS for legacy SPECs)
- SPEC ID assigned (auto-incremented)

CRITICAL: Execute `/clear` immediately after completion
- Saves 45-50K tokens
- Prepares clean context for implementation

Example:
```
User: /moai plan "Implement user authentication endpoint (JWT)"
MoAI: SPEC-001 generated successfully.
 Location: .moai/specs/SPEC-001/spec.md

 IMPORTANT: Execute /clear now to free 45-50K tokens.
```

---

### `/moai run` - DDD Implementation

Purpose: Execute ANALYZE-PRESERVE-IMPROVE cycle

Agent Delegation: `manager-develop` (cycle_type ddd/tdd/autofix)

Usage:
```bash
/moai run SPEC-001
/moai run SPEC-002
```

What It Does:
1. Reads SPEC document
2. Executes DDD cycle in 3 phases:
 - ANALYZE: Understand requirements and existing behavior
 - PRESERVE: Ensure existing behavior is protected with tests
 - IMPROVE: Implement improvements incrementally
3. Validates TRUST 5 quality gates
4. Generates implementation report

DDD Process:
```
Phase 1 (ANALYZE):
 - Understand requirements from SPEC
 - Analyze existing codebase behavior
 - Identify areas of change

Phase 2 (PRESERVE):
 - Create characterization tests for existing behavior
 - Ensure all tests pass before changes
 - Run tests → ALL PASS

Phase 3 (IMPROVE):
 - Implement changes incrementally
 - Validate behavior preservation
 - Optimize code structure
 - Run tests → ALL PASS (maintained)
```

Output:
- Implemented code (in source directories)
- Test files (in test directories)
- Quality report (TRUST 5 validation)

Requirement: Test coverage ≥ 85% (TRUST 5)

Example:
```
User: /moai run SPEC-001
MoAI: DDD implementation cycle started for SPEC-001.

 ANALYZE: Requirements analyzed, 12 acceptance criteria identified
 PRESERVE: Existing behavior protected, characterization tests created
 IMPROVE: Implementation complete, all tests passing

 Test Coverage: 92% ( meets 85% threshold)
 TRUST 5: All gates passed
```

---

### `/moai sync` - Documentation Synchronization

Purpose: Auto-generate API documentation and project artifacts

Agent Delegation: `manager-docs`

Usage:
```bash
/moai sync SPEC-001
/moai sync SPEC-002
```

What It Does:
1. Reads implemented code
2. Generates API documentation (OpenAPI format)
3. Creates architecture diagrams
4. Produces project completion report

Output:
- API documentation (OpenAPI/Swagger format)
- Architecture diagrams (Mermaid)
- `.moai/docs/SPEC-001/` directory
- Project report

Example:
```
User: /moai sync SPEC-001
MoAI: Documentation synchronized for SPEC-001.

 Generated:
 - API documentation: .moai/docs/SPEC-001/api.yaml
 - Architecture diagram: .moai/docs/SPEC-001/architecture.md
 - Completion report: .moai/docs/SPEC-001/report.md
```

---

### `/moai feedback` - Improvement Feedback Collection

Purpose: Error analysis and improvement suggestions

Agent Delegation: orchestrator (creates a GitHub issue)

Usage:
```bash
/moai feedback
/moai feedback --analyze SPEC-001
```

What It Does:
1. Analyzes errors encountered during workflow
2. Collects improvement suggestions
3. Reports to MoAI-ADK development team
4. Proposes error recovery strategies

Use Cases:
- Errors: When errors occur during any workflow phase
- Improvements: When MoAI-ADK enhancements are identified
- Analysis: Post-implementation review

Example:
```
User: /moai feedback
MoAI: Collecting feedback for recent session.

 Errors: 2 permission issues detected
 Improvements: 1 token optimization suggestion

 Feedback submitted to MoAI-ADK development team.
```

---

## Advanced Implementation (10+ minutes)

<!-- Release section removed (dev-only §21 leak): release is the maintainer-only
     97-series release-update command, not a user-facing command, so it must not
     appear in this user-facing reference. -->

### Context Initialization Rules

Rule 1: Execute `/clear` AFTER `/moai plan` (mandatory)
- SPEC generation uses 45-50K tokens
- `/clear` frees this context for implementation phase
- Prevents context overflow

Rule 2: Execute `/clear` when context > 150K tokens
- Monitor context usage via `/context` command
- Prevents token limit exceeded errors

Rule 3: Execute `/clear` after 50+ conversation messages
- Accumulated context from conversation history
- Reset for fresh context

Why `/clear` is critical:
```
Without /clear:
 SPEC generation: 50K tokens
 Implementation: 100K tokens
 Total: 150K tokens (approaching 200K limit)

With /clear:
 SPEC generation: 50K tokens
 /clear: 0K tokens (reset)
 Implementation: 100K tokens
 Total: 100K tokens (50K budget remaining)
```

---

### Command Delegation Patterns

Each command delegates to a specific agent:

| Command            | Agent              | Phase scope                 |
| ------------------ | ------------------ | --------------------------- |
| `/moai project`  | `manager-docs`     | Project-doc scaffolding     |
| `/moai plan`     | `manager-spec`     | Plan-phase                  |
| `/moai run`      | `manager-develop`  | Run-phase (ddd/tdd/autofix) |
| `/moai sync`     | `manager-docs`     | Sync-phase                  |
| `/moai feedback` | orchestrator       | Creates a GitHub issue      |

Delegation Flow:
```
User executes command
 ↓
MoAI receives command
 ↓
Command processor agent invoked
 ↓
Agent executes workflow
 ↓
Results reported to user
```

---

### Token Budget by Command

| Command        | Average Tokens | Phase Budget                          |
| -------------- | -------------- | ------------------------------------- |
| `/moai plan` | 45-50K         | Planning Phase (30K allocated)        |
| `/moai run`  | 80-100K        | Implementation Phase (180K allocated) |
| `/moai sync` | 20-25K         | Documentation Phase (40K allocated)   |
| Total          | 145-175K       | 250K per feature                      |

Optimization:
- Use Haiku 4.5 for `/moai run` (fast, cost-effective)
- Use Sonnet 4.5 for `/moai plan` (high-quality SPEC)
- Execute `/clear` between phases (critical)

---

### Error Handling

Common Errors:

| Error                     | Command                | Solution                                    |
| ------------------------- | ---------------------- | ------------------------------------------- |
| "Project not initialized" | `/moai plan`         | Run `/moai project` first                 |
| "SPEC not found"          | `/moai run SPEC-999` | Verify SPEC ID exists                       |
| "Token limit exceeded"    | Any                    | Execute `/clear` immediately                |
| "Test coverage < 85%"     | `/moai run`          | `manager-develop` adds the missing tests    |

Recovery Pattern:
```bash
# Error: Token limit exceeded
1. /clear # Reset context
2. /moai run SPEC-001 # Retry with clean context
```

---

### Workflow Variations

Standard Workflow (Full SPEC):
```
/moai project → /moai plan → /clear → /moai run → /moai sync
```

Quick Workflow (No SPEC for simple tasks):
```
/moai project → Direct implementation (for 1-2 file changes)
```

Iterative Workflow (Multiple SPECs):
```
/moai plan "Feature A" → /clear → /moai run SPEC-001 → /moai sync SPEC-001
/moai plan "Feature B" → /clear → /moai run SPEC-002 → /moai sync SPEC-002
```

---

### Integration with Git Workflow

Commands automatically integrate with Git based on `config.json` settings:

Manual Mode (Local Git):
- `/moai plan`: Prompts for branch creation
- `/moai run`: Auto-commits to local branch
- No auto-push

Personal Mode (GitHub Individual):
- `/moai plan`: Auto-creates feature branch + auto-push
- `/moai run`: Auto-commits + auto-push
- `/moai sync`: Suggests PR creation (user choice)

Team Mode (GitHub Team):
- `/moai plan`: Auto-creates feature branch + Draft PR
- `/moai run`: Auto-commits + auto-push
- `/moai sync`: Prepares PR for team review

---

## Works Well With

Skills:
- [moai-foundation-core](../SKILL.md) - Parent skill
- [moai-foundation-context](../../moai-foundation-context/SKILL.md) - Token budget management

Other Modules:
- [spec-first-ddd.md](spec-first-ddd.md) - Detailed SPEC-First DDD process
- [token-optimization.md](token-optimization.md) - /clear execution strategies
- [agents-reference.md](agents-reference.md) - Agent catalog

Agents:
- [manager-spec](agents-reference.md) - `/moai plan`
- [manager-develop](agents-reference.md) - `/moai run`
- [manager-docs](agents-reference.md) - `/moai sync`, `/moai project`

---

Maintained by: MoAI-ADK Team
Status: Production Ready
