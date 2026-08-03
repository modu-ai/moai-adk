# moai-workflow-project Reference

Progressive Disclosure Level 2: Extended documentation for advanced users and integrators.

---

## Design Surface (language-neutral contracts)

This skill is delivered as a distributed Go binary plus embedded templates (no Python surface exists in the repository). The capabilities below are described as design contracts — the runtime that exposes them is the Go `moai` CLI, invoked through slash commands and `moai` subcommands. Treat this section as a capability map, not an importable API.

### Capabilities

Project lifecycle:
- Purpose: Unified surface for project management operations
- Entry: `moai init <project>` / `moai project` (CLI), `/moai project` (slash command)
- Operations:
  - Complete project setup — documentation scaffolding, language init, template optimization
  - SPEC-driven doc generation — produce documentation from a SPEC artifact
  - Template optimization — analyze and refine embedded templates
  - Project status report — comprehensive state summary
  - Language configuration update — conversation / agent-prompt / docs / code-comment language

Documentation generation:
- Purpose: Template-based documentation generation
- Operations:
  - Generate documentation set for a detected project type and language
  - Update documentation from SPEC data
  - Multi-format export (md, html, pdf)
  - Auto-detect project type from project markers

Language initialization:
- Purpose: Language detection and configuration
- Operations:
  - Detect project language from project markers
  - Set up multilingual documentation structure
  - Localize agent prompts for a configured language
  - Token-cost-impact analysis per language

Template optimization:
- Purpose: Template analysis and optimization
- Operations:
  - Comprehensive template analysis
  - Apply optimizations (size / performance / complexity)
  - Backup templates before mutation
  - Restore templates from a backup

---

## Configuration Options

### Project Configuration Schema

```yaml
project:
  name: "string"              # Project display name
  type: "string"              # web_application, mobile_app, cli_tool, library, ml_project
  initialized_at: "datetime"  # ISO 8601 timestamp
  version: "string"           # Semantic version

language:
  conversation_language: "string"   # en, ko, ja, zh, es, fr, de
  agent_prompt_language: "string"   # english (cost-optimized) or localized
  documentation_language: "string"  # Primary documentation language
  code_comments: "string"           # Code comment language

menu_system:
  version: "string"           # Menu system version
  fully_initialized: boolean  # Complete initialization status
  modules_enabled: []         # List of enabled modules
```

### Optimization Options Schema

```yaml
backup_first: boolean                    # Create backup before optimization
apply_size_optimizations: boolean        # Reduce template file sizes
apply_performance_optimizations: boolean # Improve template performance
apply_complexity_optimizations: boolean  # Reduce template complexity
preserve_functionality: boolean          # Maintain all existing features
max_complexity_score: number             # Maximum allowed complexity (1-10)
```

### Language Configuration Presets

Supported Languages with Token Impact:

| Language | Code | Locale       | Token Cost Impact |
|----------|------|--------------|-------------------|
| English  | en   | en_US.UTF-8  | 0% (baseline)     |
| Korean   | ko   | ko_KR.UTF-8  | +20%              |
| Japanese | ja   | ja_JP.UTF-8  | +25%              |
| Chinese  | zh   | zh_CN.UTF-8  | +15%              |
| Spanish  | es   | es_ES.UTF-8  | +5%               |
| French   | fr   | fr_FR.UTF-8  | +5%               |
| German   | de   | de_DE.UTF-8  | +5%               |

---

## Integration Patterns

### Pattern 1: SPEC-Driven Documentation Workflow

Integration with `/moai plan` and `/moai sync` (language-neutral sequence):

1. `/moai plan "..."` generates a SPEC artifact under `.moai/specs/SPEC-XXX/`.
2. The orchestrator reads the SPEC and emits the documentation scaffolding via `/moai project`:
   - Feature documentation with the SPEC's requirements
   - API documentation with endpoint details (when applicable)
   - Architecture documentation
   - Multilingual versions when the project configures multiple documentation languages
3. `/moai sync SPEC-XXX` refreshes docs and opens the PR.

### Pattern 2: CI/CD Documentation Gate

Run a documentation-completeness check in CI as a non-blocking advisory (no Python helper required):

- Step 1: `moai project --status` emits a machine-readable status summary (docs completion, language configured, templates optimized, warnings, errors).
- Step 2: A CI step parses the summary and annotates the run when `docs_completion` is below the project's threshold or any error is present.
- Step 3: Treat the check as advisory (warn-on-failure) — documentation drift should not gate a build.

### Pattern 3: Multi-Project Template Sharing

Optimized templates live in the distributed binary's embedded FS plus the per-project `.moai/` tree; share them via version control, not a runtime API:

1. Author canonical templates in the source repo (the SSOT).
2. Other projects consume them through `moai update` (re-render from the embedded catalog) or by vendoring the relevant `.moai/` files.
3. Backup-then-mutate: `moai update` preserves user-owned content (`hns-*` skills, project memory, specs) per the namespace separation contract — only template-managed surfaces are re-rendered.

### Pattern 4: Language-Aware Agent Delegation

When delegating to a sub-agent, the orchestrator passes the active `conversation_language` and `agent_prompt_language` (read from `.moai/config/sections/language.yaml`) in the spawn prompt. There is no library call — the delegation surface is the `Agent()` tool:

1. Resolve the active language settings from `language.yaml`.
2. Compose the spawn prompt in `agent_prompt_language` (English by default for cost control).
3. Inject the user's `conversation_language` so the agent's user-facing output respects it.
4. The token-cost overhead for non-English `conversation_language` follows the table in § Language Configuration Presets (Korean/Japanese/Chinese carry a measurable overhead; English is the baseline).

---

## Troubleshooting

### Common Issues

Issue: Documentation generation fails with template not found:
- Cause: Template directory missing or corrupted
- Solution: Re-run `moai update` to re-render templates from the embedded catalog, or restore the project's `.moai/` files from version control
- Prevention: Keep `.moai/` under version control so template state is recoverable

Issue: Language detection returns incorrect language:
- Cause: Insufficient language indicators in project files
- Solution: Manually set the language in `.moai/config/sections/language.yaml` (`conversation_language: ko`)
- Prevention: Include language-revealing comments / config markers in main source files

Issue: Template optimization causes functionality loss:
- Cause: Aggressive optimization removed necessary content
- Solution: Restore `.moai/` templates from version control (the prior commit)
- Prevention: Run optimization on a branch and verify behavior before merging

Issue: Multilingual documentation structure incomplete:
- Cause: Partial initialization or interrupted process
- Solution: Re-run `/moai project` to complete the documentation scaffolding
- Prevention: Ensure the process runs to completion; resume from `.moai/specs/<SPEC>/progress.md` after an interrupt

Issue: High token cost for non-English languages:
- Cause: Localized agent prompts increase token usage
- Solution: Use `agent_prompt_language: en` with `conversation_language: ko` for cost optimization
- Prevention: Configure language settings before heavy usage

### Diagnostic Commands

The diagnostic surface is the `moai` CLI, not a Python API. Run the status command and read its output:

```bash
# Full diagnostic report
moai project --status
```

The report surfaces: initialization state, language configuration, documentation completion %, template status, errors, and warnings.

### Log Locations

- Project logs: `.moai/logs/project.log`
- Template optimization logs: `.moai/logs/template-optimizer.log`
- Language initialization logs: `.moai/logs/language-init.log`

---

## External Resources

### Official Documentation

- MoAI-ADK Documentation: https://github.com/moai-adk/docs
- Claude Code Skills Guide: https://docs.anthropic.com/claude-code/skills
- SPEC-First DDD Methodology: See `moai-foundation-core/modules/spec-first-ddd.md`

### Related Skills

- moai-foundation-core - Core execution patterns and SPEC workflow
- moai-foundation-cc - Claude Code integration patterns
- moai-workflow-spec - SPEC workflow orchestration (plan / run / sync)
- moai-workflow-docs-claim-check - README / public-docs claim validation

### Template Resources

- Documentation Templates: `/templates/doc-templates/`
- Product Template: `/templates/doc-templates/product-template.md`
- Technical Template: `/templates/doc-templates/tech-template.md`
- Structure Template: `/templates/doc-templates/structure-template.md`

### Version History

| Version | Date       | Changes                                           |
|---------|------------|---------------------------------------------------|
| 2.0.0   | 2025-11-27 | Integrated modular architecture                   |
| 1.5.0   | 2025-11-20 | Added template optimization module                |
| 1.0.0   | 2025-11-15 | Initial release with documentation management     |

---

Status: Reference Documentation Complete
Skill Version: 2.0.0
