# Commands Reference - MoAI-ADK Core Commands

Purpose: Complete reference for MoAI-ADK's 6 core commands used in SPEC-First DDD workflow.

Last Updated: 2025-11-25
Version: 2.0.0

---

## Quick Reference (30 seconds)

MoAI-ADK provides 6 core commands for SPEC-First DDD execution:

| Command            | Purpose                | Phase         |
| ------------------ | ---------------------- | ------------- |
| `/moai:0-project`  | Project initialization | Setup         |
| `/moai:1-plan`     | SPEC generation        | Planning      |
| `/moai:2-run`      | DDD implementation     | Development   |
| `/moai:3-sync`     | Documentation sync     | Documentation |
| `/moai:9-feedback` | Feedback collection    | Improvement   |
| `/moai:99-release` | Production deployment  | Release       |

Required Workflow:
```
1. /moai:0-project # Initialize
2. /moai:1-plan "description" # Generate SPEC
3. /clear # Clear context (REQUIRED)
4. /moai:2-run SPEC-001 # Implement
5. /moai:3-sync SPEC-001 # Document
6. /moai:9-feedback # Improve
```

Critical Rule: Execute `/clear` after `/moai:1-plan` (saves 45-50K tokens)

---

## Implementation Guide (5 minutes)

### `/moai:0-project` - Project Initialization

Purpose: Initialize project structure and generate configuration

Agent Delegation: `workflow-project`

Usage:
```bash
/moai:0-project
/moai:0-project --with-git
```

What It Does:
1. Creates `.moai/` directory structure
2. Generates `config.json` with default settings
3. Initializes Git repository (if `--with-git` flag provided)
4. Sets up MoAI-ADK workflows

Output:
- `.moai/` directory
- `.moai/config/config.yaml`
- `.moai/memory/` (empty, ready for session state)
- `.moai/logs/` (empty, ready for logging)

Next Step: Ready for SPEC generation via `/moai:1-plan`

Example:
```
User: /moai:0-project
MoAI: Project initialized successfully.
 - .moai/config/config.yaml created
 - Git workflow set to 'manual' mode
 Ready for SPEC generation.
```

---

### `/moai:1-plan` - SPEC Generation

Purpose: Generate SPEC document in EARS format

Agent Delegation: `workflow-spec`

Usage:
```bash
/moai:1-plan "Implement user authentication endpoint (JWT)"
/moai:1-plan "Add dark mode toggle to settings page"
```

What It Does:
1. Analyzes user request
2. Generates EARS format SPEC document
3. Creates `.moai/specs/SPEC-XXX/` directory
4. Saves `spec.md` with requirements

EARS Format (5 sections):
- WHEN (trigger conditions)
- IF (preconditions)
- THE SYSTEM SHALL (functional requirements)
- WHERE (constraints)
- UBIQUITOUS (quality requirements)

Output:
- `.moai/specs/SPEC-001/spec.md` (EARS document)
- SPEC ID assigned (auto-incremented)

CRITICAL: Execute `/clear` immediately after completion
- Saves 45-50K tokens
- Prepares clean context for implementation

Example:
```
User: /moai:1-plan "Implement user authentication endpoint (JWT)"
MoAI: SPEC-001 generated successfully.
 Location: .moai/specs/SPEC-001/spec.md

 IMPORTANT: Execute /clear now to free 45-50K tokens.
```

---

### `/moai:2-run` - DDD Implementation

Purpose: Execute ANALYZE-PRESERVE-IMPROVE cycle

Agent Delegation: `workflow-ddd`

Usage:
```bash
/moai:2-run SPEC-001
/moai:2-run SPEC-002
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
User: /moai:2-run SPEC-001
MoAI: DDD implementation cycle started for SPEC-001.

 ANALYZE: Requirements analyzed, 12 acceptance criteria identified
 PRESERVE: Existing behavior protected, characterization tests created
 IMPROVE: Implementation complete, all tests passing

 Test Coverage: 92% ( meets 85% threshold)
 TRUST 5: All gates passed
```

---

### `/moai:3-sync` - Documentation Synchronization

Purpose: Auto-generate API documentation and project artifacts

Agent Delegation: `workflow-docs`

Usage:
```bash
/moai:3-sync SPEC-001
/moai:3-sync SPEC-002
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
User: /moai:3-sync SPEC-001
MoAI: Documentation synchronized for SPEC-001.

 Generated:
 - API documentation: .moai/docs/SPEC-001/api.yaml
 - Architecture diagram: .moai/docs/SPEC-001/architecture.md
 - Completion report: .moai/docs/SPEC-001/report.md
```

---

### `/moai:9-feedback` - Improvement Feedback Collection

Purpose: Error analysis and improvement suggestions

Agent Delegation: `core-quality`

Usage:
```bash
/moai:9-feedback
/moai:9-feedback --analyze SPEC-001
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
User: /moai:9-feedback
MoAI: Collecting feedback for recent session.

 Errors: 2 permission issues detected
 Improvements: 1 token optimization suggestion

 Feedback submitted to MoAI-ADK development team.
```

---

### `/moai:99-release` - Production Deployment

Purpose: Production deployment workflow

Agent Delegation: `infra-devops`

Usage:
```bash
/moai:99-release
```

What It Does:
1. Validates all TRUST 5 quality gates
2. Runs full test suite
3. Builds production artifacts
4. Deploys to production environment

Note: This command is local-only and NOT synchronized to the package template. It's for local development and testing.

---

## Advanced Implementation (10+ minutes)

### Context Initialization Rules

Rule 1: Execute `/clear` AFTER `/moai:1-plan` (mandatory)
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

| Command            | Agent              | Agent Type              |
| ------------------ | ------------------ | ----------------------- |
| `/moai:0-project`  | `workflow-project` | Tier 1 (Always Active)  |
| `/moai:1-plan`     | `workflow-spec`    | Tier 1 (Always Active)  |
| `/moai:2-run`      | `workflow-ddd`     | Tier 1 (Always Active)  |
| `/moai:3-sync`     | `workflow-docs`    | Tier 1 (Always Active)  |
| `/moai:9-feedback` | `core-quality`     | Tier 2 (Auto-triggered) |
| `/moai:99-release` | `infra-devops`     | Tier 3 (Lazy-loaded)    |

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
| `/moai:1-plan` | 45-50K         | Planning Phase (30K allocated)        |
| `/moai:2-run`  | 80-100K        | Implementation Phase (180K allocated) |
| `/moai:3-sync` | 20-25K         | Documentation Phase (40K allocated)   |
| Total          | 145-175K       | 250K per feature                      |

Optimization:
- Use Haiku 4.5 for `/moai:2-run` (fast, cost-effective)
- Use Sonnet 4.5 for `/moai:1-plan` (high-quality SPEC)
- Execute `/clear` between phases (critical)

---

### Error Handling

Common Errors:

| Error                     | Command                | Solution                                    |
| ------------------------- | ---------------------- | ------------------------------------------- |
| "Project not initialized" | `/moai:1-plan`         | Run `/moai:0-project` first                 |
| "SPEC not found"          | `/moai:2-run SPEC-999` | Verify SPEC ID exists                       |
| "Token limit exceeded"    | Any                    | Execute `/clear` immediately                |
| "Test coverage < 85%"     | `/moai:2-run`          | `core-quality` auto-generates missing tests |

Recovery Pattern:
```bash
# Error: Token limit exceeded
1. /clear # Reset context
2. /moai:2-run SPEC-001 # Retry with clean context
```

---

### Workflow Variations

Standard Workflow (Full SPEC):
```
/moai:0-project → /moai:1-plan → /clear → /moai:2-run → /moai:3-sync
```

Quick Workflow (No SPEC for simple tasks):
```
/moai:0-project → Direct implementation (for 1-2 file changes)
```

Iterative Workflow (Multiple SPECs):
```
/moai:1-plan "Feature A" → /clear → /moai:2-run SPEC-001 → /moai:3-sync SPEC-001
/moai:1-plan "Feature B" → /clear → /moai:2-run SPEC-002 → /moai:3-sync SPEC-002
```

---

### Integration with Git Workflow

Commands automatically integrate with Git based on `config.json` settings:

Manual Mode (Local Git):
- `/moai:1-plan`: Prompts for branch creation
- `/moai:2-run`: Auto-commits to local branch
- No auto-push

Personal Mode (GitHub Individual):
- `/moai:1-plan`: Auto-creates feature branch + auto-push
- `/moai:2-run`: Auto-commits + auto-push
- `/moai:3-sync`: Suggests PR creation (user choice)

Team Mode (GitHub Team):
- `/moai:1-plan`: Auto-creates feature branch + Draft PR
- `/moai:2-run`: Auto-commits + auto-push
- `/moai:3-sync`: Prepares PR for team review

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
- [workflow-project](agents-reference.md#tier-1-command-processors) - `/moai:0-project`
- [workflow-spec](agents-reference.md#tier-1-command-processors) - `/moai:1-plan`
- [workflow-ddd](agents-reference.md#tier-1-command-processors) - `/moai:2-run`
- [workflow-docs](agents-reference.md#tier-1-command-processors) - `/moai:3-sync`

---

Maintained by: MoAI-ADK Team
Status: Production Ready
