# SPEC-I18N-001: Implementation Plan

## SPEC Reference

| Field   | Value       |
| ------- | ----------- |
| SPEC ID | SPEC-I18N-001 |
| Title   | Complete Multilingual Translation of Documentation |

---

## Technical Approach

### AI-Assisted Translation Workflow

Each file follows a deterministic translation pipeline:

1. **Read source file** from `content/ko/` (for English) or `content/en/` (for Chinese/Japanese)
2. **Parse MDX structure** to separate translatable prose from non-translatable syntax
3. **Translate prose** using Claude with context-aware prompts that include:
   - The full source file content
   - A glossary of MoAI-specific terms and their approved translations
   - Instructions to preserve MDX structure
4. **Write translated file** to the target locale directory, replacing the existing placeholder
5. **Validate** that the output is valid MDX (import statements intact, JSX balanced, frontmatter valid)

### Translation Quality Guidelines

**English**: Clear, concise technical English. Use active voice. Follow Nextra documentation conventions.

**Chinese (Simplified)**: Use standard Simplified Chinese characters. Technical documentation register. Follow conventions of established Chinese developer documentation (e.g., Vue.js Chinese docs, React Chinese docs).

**Japanese**: Use desu/masu form for politeness. Follow conventions of established Japanese developer documentation (e.g., MDN Japanese docs). Use katakana for foreign technical loanwords where standard.

### Glossary of Key Terms

| English Term         | Chinese (zh)           | Japanese (ja)           | Notes                     |
| -------------------- | ---------------------- | ----------------------- | ------------------------- |
| Agent                | 代理 (Agent)           | エージェント (Agent)    | Keep English in parens    |
| Skill                | 技能 (Skill)           | スキル (Skill)          | Keep English in parens    |
| Hook                 | 钩子 (Hook)            | フック (Hook)           | Keep English in parens    |
| Workflow             | 工作流                 | ワークフロー             | Common localized terms    |
| Command              | 命令                   | コマンド                 | Common localized terms    |
| Plugin               | 插件 (Plugin)          | プラグイン (Plugin)     | Keep English in parens    |
| Builder              | 构建器 (Builder)       | ビルダー (Builder)      | Keep English in parens    |
| Getting Started      | 开始使用               | はじめに                 | Section titles            |
| Best Practices       | 最佳实践               | ベストプラクティス        | Section titles            |
| Troubleshooting      | 故障排除               | トラブルシューティング    | Section titles            |
| Installation         | 安装                   | インストール             | Common localized terms    |
| Configuration        | 配置                   | 設定                    | Common localized terms    |
| Quick Start          | 快速入门               | クイックスタート          | Section titles            |

Product names always in English: MoAI-ADK, Claude Code, Nextra, Next.js, SPEC, DDD, TDD, TRUST 5, EARS, Git Worktree.

---

## Milestones

### Phase 1: Complete English Translation

**Priority**: Primary Goal
**Scope**: 42 files in `content/en/`
**Source**: `content/ko/` (Korean originals)
**Dependency**: None (can start immediately)

| Milestone   | Section           | Files | Source Path                  | Target Path                  |
| ----------- | ----------------- | ----- | ---------------------------- | ---------------------------- |
| M1.1        | workflow-commands  | 5     | content/ko/workflow-commands/ | content/en/workflow-commands/ |
| M1.2        | utility-commands   | 5     | content/ko/utility-commands/  | content/en/utility-commands/  |
| M1.3        | claude-code        | 15    | content/ko/claude-code/       | content/en/claude-code/       |
| M1.4        | advanced           | 9     | content/ko/advanced/          | content/en/advanced/          |
| M1.5        | worktree           | 4     | content/ko/worktree/          | content/en/worktree/          |
| M1.6        | moai-rank          | 4     | content/ko/moai-rank/         | content/en/moai-rank/         |

**Build Validation**: Run `npm run build` after each milestone (M1.1 through M1.6).

### Phase 2: Chinese Translation

**Priority**: Secondary Goal
**Scope**: 55 files in `content/zh/`
**Source**: `content/en/` (completed English translations)
**Dependency**: Phase 1 must be complete

| Milestone   | Section           | Files | Source Path                  | Target Path                  |
| ----------- | ----------------- | ----- | ---------------------------- | ---------------------------- |
| M2.1        | index + getting-started + core-concepts | 13 | content/en/ | content/zh/ |
| M2.2        | workflow-commands  | 5     | content/en/workflow-commands/ | content/zh/workflow-commands/ |
| M2.3        | utility-commands   | 5     | content/en/utility-commands/  | content/zh/utility-commands/  |
| M2.4        | claude-code        | 15    | content/en/claude-code/       | content/zh/claude-code/       |
| M2.5        | advanced           | 9     | content/en/advanced/          | content/zh/advanced/          |
| M2.6        | worktree + moai-rank | 8   | content/en/                   | content/zh/                   |

### Phase 3: Japanese Translation

**Priority**: Secondary Goal
**Scope**: 55 files in `content/ja/`
**Source**: `content/en/` (completed English translations)
**Dependency**: Phase 1 must be complete
**Parallelism**: Phase 3 may run in parallel with Phase 2

| Milestone   | Section           | Files | Source Path                  | Target Path                  |
| ----------- | ----------------- | ----- | ---------------------------- | ---------------------------- |
| M3.1        | index + getting-started + core-concepts | 13 | content/en/ | content/ja/ |
| M3.2        | workflow-commands  | 5     | content/en/workflow-commands/ | content/ja/workflow-commands/ |
| M3.3        | utility-commands   | 5     | content/en/utility-commands/  | content/ja/utility-commands/  |
| M3.4        | claude-code        | 15    | content/en/claude-code/       | content/ja/claude-code/       |
| M3.5        | advanced           | 9     | content/en/advanced/          | content/ja/advanced/          |
| M3.6        | worktree + moai-rank | 8   | content/en/                   | content/ja/                   |

### Phase 4: Final Validation

**Priority**: Final Goal
**Scope**: All 220 pages across 4 locales
**Dependency**: Phases 1, 2, and 3 complete

| Milestone   | Task                                    |
| ----------- | --------------------------------------- |
| M4.1        | Full build validation (`npm run build`) |
| M4.2        | Verify all 220 pages return HTTP 200    |
| M4.3        | Spot-check translation quality (5 pages per locale) |
| M4.4        | Verify language switcher works correctly across all pages |

---

## Architecture Design Direction

### No Architectural Changes Required

The existing i18n infrastructure is fully functional:

- `middleware.ts` handles locale detection and routing
- Nextra 4 provides built-in i18n support via directory-based locale structure
- `_meta.ts` files provide translated navigation for all locales
- Language selector is operational

The work is purely content translation with no code changes needed.

### File Naming Convention

All translated files maintain identical filenames to their source:

```
content/ko/advanced/agent-guide.mdx    (Korean source)
content/en/advanced/agent-guide.mdx    (English translation)
content/zh/advanced/agent-guide.mdx    (Chinese translation)
content/ja/advanced/agent-guide.mdx    (Japanese translation)
```

### Execution Context

Each phase can be executed as a separate `/moai run SPEC-I18N-001` session:

- Session 1: Phase 1 (English) - 42 files
- Session 2: Phase 2 (Chinese) - 55 files
- Session 3: Phase 3 (Japanese) - 55 files
- Session 4: Phase 4 (Validation) - verification only

Due to the large file count, each phase may require multiple sessions with `/clear` between section batches.

---

## Risks and Mitigations

| Risk                                  | Likelihood | Impact | Mitigation                                                  |
| ------------------------------------- | ---------- | ------ | ----------------------------------------------------------- |
| MDX syntax broken during translation  | Medium     | High   | Validate each file individually; build after each section   |
| Inconsistent terminology across files | Medium     | Medium | Use glossary; translate section-by-section for consistency  |
| Token budget exceeded per session     | High       | Low    | Break into section-level batches; /clear between batches    |
| Mermaid diagram labels lost           | Low        | Medium | Parse Mermaid blocks separately; translate only node labels |
| Existing English translations inconsistent with new ones | Low | Medium | Review existing 13 files for glossary compliance after Phase 1 |
| Build failure from untranslated frontmatter | Low   | High   | Translate frontmatter title/description as first step per file |

---

## Dependencies

- `content/ko/` must remain stable during translation (no concurrent content updates)
- Node.js and npm must be available for build validation
- No external translation services or APIs required (Claude performs all translations)
