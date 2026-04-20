ARCHIVED — SPEC-I18N-001

이 SPEC은 moai-docs 레포(archived)에서 이관된 원본 문서입니다.
현재 유효한 명세는 SPEC-DOCS-SITE-001 (Hextra 전환 포함)을 참조하세요.

이관 날짜: 2026-04-17
이관 근거: REQ-DS-28 (SPEC-I18N-001 흡수 프로세스)
원본 위치: modu-ai/moai-adk-docs:.moai/specs/SPEC-I18N-001/

---

# SPEC-I18N-001: Complete Multilingual Translation of MoAI-ADK Documentation Site

## Metadata

| Field       | Value                                              |
| ----------- | -------------------------------------------------- |
| SPEC ID     | SPEC-I18N-001                                      |
| Title       | Complete Multilingual Translation of Documentation |
| Created     | 2026-02-07                                         |
| Status      | Planned                                            |
| Priority    | High                                               |
| Assigned    | expert-frontend (implementation)                   |
| Labels      | i18n, documentation, translation, nextra           |

---

## Environment

The MoAI-ADK documentation site is built with **Nextra 4** and **Next.js 16**, configured for 4 locales: Korean (ko), English (en), Chinese (zh), and Japanese (ja).

### Infrastructure (Complete)

- Directory structure: `content/ko/`, `content/en/`, `content/zh/`, `content/ja/`
- `_meta.ts` navigation files: Fully translated for all 4 locales
- `theme.config.tsx`: i18n configuration complete
- `middleware.ts`: Locale detection, cookie persistence, Accept-Language parsing, and redirect working
- Language selector: Nextra built-in i18n switcher functional
- Build: Passes successfully with all 220 pages (55 per locale x 4 locales) returning HTTP 200

### Source Language

Korean (`content/ko/`) is the canonical source with 55 fully authored MDX files across 8 sections.

### Current Translation Status

| Locale       | Translated | Untranslated | Current State of Untranslated Files           |
| ------------ | ---------- | ------------ | --------------------------------------------- |
| ko (Korean)  | 55/55      | 0            | Original source language                       |
| en (English) | 13/55      | 42           | Untranslated files contain Korean text copies  |
| zh (Chinese) | 0/55       | 55           | Mix of English and Korean text copies          |
| ja (Japanese)| 0/55       | 55           | Mix of English and Korean text copies          |

**Total files requiring translation: 152**

### File Inventory by Section

| Section            | Files | en Status     | zh Status     | ja Status     |
| ------------------ | ----- | ------------- | ------------- | ------------- |
| index.mdx (root)   | 1     | Translated    | Untranslated  | Untranslated  |
| getting-started/   | 7     | Translated    | Untranslated  | Untranslated  |
| core-concepts/     | 5     | Translated    | Untranslated  | Untranslated  |
| workflow-commands/  | 5     | Untranslated  | Untranslated  | Untranslated  |
| utility-commands/   | 5     | Untranslated  | Untranslated  | Untranslated  |
| advanced/          | 9     | Untranslated  | Untranslated  | Untranslated  |
| claude-code/       | 15    | Untranslated  | Untranslated  | Untranslated  |
| worktree/          | 4     | Untranslated  | Untranslated  | Untranslated  |
| moai-rank/         | 4     | Untranslated  | Untranslated  | Untranslated  |

---

## Assumptions

1. **Korean as source of truth**: All translations are derived from `content/ko/` as the canonical source. The existing 13 English translations (getting-started + core-concepts) were already translated from Korean and are correct.

2. **AI-assisted translation**: Translations will be performed by Claude using the Korean source files, with human review recommended for production deployment.

3. **MDX structure preservation**: All MDX-specific syntax (imports, Callout components, code blocks, frontmatter, Mermaid diagrams, links) must be preserved exactly during translation. Only human-readable prose and headings are translated.

4. **No new content creation**: This SPEC covers translation of existing content only. No new pages or sections are created.

5. **_meta.ts files are complete**: All navigation labels in `_meta.ts` files are already translated for all 4 locales and require no changes.

6. **Technical terms preserved**: Product names (MoAI-ADK, Claude Code, SPEC, DDD, TRUST 5), command names (`/moai plan`, `/moai run`), code identifiers, and file paths remain untranslated in all locales.

7. **Build must remain passing**: At no point during the translation process should the build break. Translations are applied file-by-file, each independently valid.

---

## Requirements

### REQ-1: English Translation Completion (42 files)

**When** a user visits the English documentation site, **the system shall** display all 55 pages with content fully translated into natural English prose.

- REQ-1.1: **When** a user navigates to any page under `content/en/advanced/`, **the system shall** display content in English, not Korean.
- REQ-1.2: **When** a user navigates to any page under `content/en/claude-code/`, **the system shall** display content in English, not Korean.
- REQ-1.3: **When** a user navigates to any page under `content/en/workflow-commands/`, **the system shall** display content in English, not Korean.
- REQ-1.4: **When** a user navigates to any page under `content/en/utility-commands/`, **the system shall** display content in English, not Korean.
- REQ-1.5: **When** a user navigates to any page under `content/en/worktree/`, **the system shall** display content in English, not Korean.
- REQ-1.6: **When** a user navigates to any page under `content/en/moai-rank/`, **the system shall** display content in English, not Korean.

### REQ-2: Chinese Translation (55 files)

**When** a user visits the Chinese documentation site, **the system shall** display all 55 pages with content fully translated into Simplified Chinese.

- REQ-2.1: The system shall display all section headings, paragraph text, list items, table content, callout messages, and frontmatter metadata in Simplified Chinese.
- REQ-2.2: **While** translating to Chinese, the system shall use standard Simplified Chinese (zh-CN) conventions, not Traditional Chinese (zh-TW).

### REQ-3: Japanese Translation (55 files)

**When** a user visits the Japanese documentation site, **the system shall** display all 55 pages with content fully translated into Japanese.

- REQ-3.1: The system shall display all section headings, paragraph text, list items, table content, callout messages, and frontmatter metadata in Japanese.
- REQ-3.2: **While** translating to Japanese, the system shall use appropriate honorific and formal register suitable for technical documentation.

### REQ-4: MDX Structure Preservation

The system shall preserve all MDX structural elements unchanged during translation.

- REQ-4.1: The system shall not modify `import` statements at the top of MDX files.
- REQ-4.2: The system shall not modify JSX component syntax (e.g., `<Callout>`, `<Steps>`, `<Tabs>`).
- REQ-4.3: The system shall not modify code blocks (fenced with triple backticks), including code comments in English.
- REQ-4.4: The system shall not modify internal links (`[text](/path)`) except translating the visible link text.
- REQ-4.5: The system shall preserve Mermaid diagram syntax, translating only node labels where they contain prose.
- REQ-4.6: The system shall translate YAML frontmatter `title` and `description` fields into the target language.

### REQ-5: Technical Term Consistency

The system shall maintain consistent technical terminology across all translations.

- REQ-5.1: Product names shall remain in English across all locales: MoAI-ADK, Claude Code, Nextra, Next.js, SPEC, DDD, TDD, TRUST 5, EARS.
- REQ-5.2: Command names shall remain in their original form: `/moai plan`, `/moai run`, `/moai sync`, `/moai project`, `/moai fix`, `/moai loop`, `/moai feedback`.
- REQ-5.3: File paths and code identifiers shall remain unchanged: `.claude/`, `.moai/`, `CLAUDE.md`, `config.yaml`.
- REQ-5.4: **If** a technical term has a widely-accepted localized form in the target language, **then** the system shall use the localized form followed by the English original in parentheses on first occurrence only.

### REQ-6: Build Integrity

The system shall not introduce build failures during the translation process.

- REQ-6.1: The system shall verify that `npm run build` passes after each batch of translations.
- REQ-6.2: **If** a translated file causes a build error, **then** the system shall revert that file and report the error before proceeding.

### REQ-7: Unwanted Behaviors

- REQ-7.1: The system shall not use machine-translation artifacts such as awkward literal translations, inconsistent terminology, or unnatural sentence structures.
- REQ-7.2: The system shall not translate content inside code blocks or inline code spans.
- REQ-7.3: The system shall not change the file structure, file names, or directory hierarchy.
- REQ-7.4: The system shall not modify the existing 13 English translations (getting-started + core-concepts sections) unless errors are discovered.

---

## Specifications

### Translation Source Mapping

| Target Locale | Source for Translation | Rationale                                                    |
| ------------- | --------------------- | ------------------------------------------------------------ |
| en (English)  | content/ko/ (Korean)  | Korean is the canonical source; translate Korean to English   |
| zh (Chinese)  | content/en/ (English) | Translate from completed English to Chinese for accuracy      |
| ja (Japanese) | content/en/ (English) | Translate from completed English to Japanese for accuracy     |

**Translation order matters**: English must be completed first, then Chinese and Japanese can be translated from the English versions in parallel.

### Section Priority Order

Sections are prioritized by user impact and traffic importance:

| Priority | Section            | Files | Rationale                              |
| -------- | ------------------ | ----- | -------------------------------------- |
| P1       | getting-started/   | 7     | First contact for new users            |
| P1       | core-concepts/     | 5     | Foundational understanding             |
| P1       | index.mdx (root)   | 1     | Landing page                           |
| P2       | workflow-commands/  | 5     | Core workflow documentation            |
| P2       | utility-commands/   | 5     | Daily-use command reference            |
| P3       | claude-code/       | 15    | Claude Code integration reference      |
| P3       | advanced/          | 9     | Advanced user features                 |
| P4       | worktree/          | 4     | Specialized feature documentation      |
| P4       | moai-rank/         | 4     | Supplementary feature                  |

### Translation Execution Strategy

**Phase 1 - English Completion (42 files)**:
Translate remaining `content/en/` files from Korean source (`content/ko/`). This completes the English locale and provides a clean English source for zh/ja translations.

**Phase 2 - Chinese Translation (55 files)**:
Translate all `content/zh/` files from the completed English versions (`content/en/`). Process section-by-section following priority order.

**Phase 3 - Japanese Translation (55 files)**:
Translate all `content/ja/` files from the completed English versions (`content/en/`). Process section-by-section following priority order. Phase 2 and Phase 3 may be executed in parallel since they are independent.

### File Translation Process (per file)

1. Read the source file (ko for en translation, en for zh/ja translation)
2. Identify translatable content (prose, headings, frontmatter title/description, list items, table cells, callout text)
3. Identify non-translatable content (imports, JSX, code blocks, links paths, Mermaid syntax structure, inline code)
4. Translate all translatable content into the target language
5. Write the translated content to the target locale file, overwriting the existing placeholder
6. Verify the file is valid MDX by checking for syntax errors

### Quality Assurance Approach

- **Structural validation**: Each translated file must have the same line count (within 10% tolerance) as the source file
- **Component preservation**: All `import` statements and JSX components in the translated file must match the source file exactly
- **Build validation**: Run `npm run build` after completing each section to catch errors early
- **Link integrity**: All internal cross-references must point to valid locale-prefixed paths

---

## Traceability

| Requirement | Plan Reference         | Acceptance Reference       |
| ----------- | ---------------------- | -------------------------- |
| REQ-1       | Phase 1 (plan.md)     | AC-1 (acceptance.md)       |
| REQ-2       | Phase 2 (plan.md)     | AC-2 (acceptance.md)       |
| REQ-3       | Phase 3 (plan.md)     | AC-3 (acceptance.md)       |
| REQ-4       | All Phases (plan.md)  | AC-4 (acceptance.md)       |
| REQ-5       | All Phases (plan.md)  | AC-5 (acceptance.md)       |
| REQ-6       | All Phases (plan.md)  | AC-6 (acceptance.md)       |
| REQ-7       | All Phases (plan.md)  | AC-7 (acceptance.md)       |
