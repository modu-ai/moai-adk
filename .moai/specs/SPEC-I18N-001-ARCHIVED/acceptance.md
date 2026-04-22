# SPEC-I18N-001: Acceptance Criteria

## SPEC Reference

| Field   | Value       |
| ------- | ----------- |
| SPEC ID | SPEC-I18N-001 |
| Title   | Complete Multilingual Translation of Documentation |

---

## AC-1: English Translation Completion

### Scenario 1.1: All English pages display English content

```gherkin
Given the documentation site is running
When a user visits any page under the /en/ locale
Then all visible text content shall be in English
And no Korean, Chinese, or Japanese text shall appear in prose sections
```

### Scenario 1.2: Remaining 42 English files are translated

```gherkin
Given the 42 untranslated files in content/en/ contain Korean text
When the English translation phase is complete
Then all 42 files shall contain English prose
And the 13 previously translated files (getting-started + core-concepts) shall remain unchanged
And the total English file count shall be 55
```

### Scenario 1.3: Section-level verification (per section)

```gherkin
Given the section "advanced/" contains 9 MDX files
When translation of the advanced section is complete
Then each of the 9 files shall have English headings
And each file's frontmatter title shall be in English
And each file's frontmatter description shall be in English
```

---

## AC-2: Chinese Translation Completion

### Scenario 2.1: All Chinese pages display Simplified Chinese content

```gherkin
Given the documentation site is running
When a user visits any page under the /zh/ locale
Then all visible text content shall be in Simplified Chinese
And no English-only prose paragraphs shall remain (excluding code blocks and technical terms)
```

### Scenario 2.2: All 55 Chinese files are translated

```gherkin
Given all 55 files in content/zh/ contain English or Korean placeholder text
When the Chinese translation phase is complete
Then all 55 files shall contain Simplified Chinese prose
And each file's frontmatter title shall be in Chinese
And each file's frontmatter description shall be in Chinese
```

### Scenario 2.3: Chinese terminology consistency

```gherkin
Given the translation glossary defines standard Chinese terms
When any Chinese page uses a glossary term
Then the term shall match the glossary definition
And the English original shall appear in parentheses on first occurrence per page
```

---

## AC-3: Japanese Translation Completion

### Scenario 3.1: All Japanese pages display Japanese content

```gherkin
Given the documentation site is running
When a user visits any page under the /ja/ locale
Then all visible text content shall be in Japanese
And no English-only prose paragraphs shall remain (excluding code blocks and technical terms)
```

### Scenario 3.2: All 55 Japanese files are translated

```gherkin
Given all 55 files in content/ja/ contain English or Korean placeholder text
When the Japanese translation phase is complete
Then all 55 files shall contain Japanese prose
And each file's frontmatter title shall be in Japanese
And each file's frontmatter description shall be in Japanese
```

### Scenario 3.3: Japanese register appropriateness

```gherkin
Given the documentation targets a developer audience
When Japanese translation uses verb forms
Then the desu/masu polite form shall be used consistently
And casual or overly formal register shall be avoided
```

---

## AC-4: MDX Structure Preservation

### Scenario 4.1: Import statements preserved

```gherkin
Given a source file begins with "import { Callout } from 'nextra/components'"
When the file is translated to any target locale
Then the translated file shall begin with the identical import statement
And no import statements shall be added, removed, or modified
```

### Scenario 4.2: JSX components preserved

```gherkin
Given a source file contains <Callout type="info"> with content </Callout>
When the file is translated
Then the JSX tags shall remain identical (<Callout type="info"> and </Callout>)
And only the text content between the tags shall be translated
```

### Scenario 4.3: Code blocks preserved

```gherkin
Given a source file contains a fenced code block (```)
When the file is translated
Then the code block content shall remain completely unchanged
And the language identifier after ``` shall remain unchanged
And code comments inside the block shall remain unchanged
```

### Scenario 4.4: Frontmatter translated correctly

```gherkin
Given a source file has YAML frontmatter with title and description fields
When the file is translated to Chinese
Then the title field value shall be in Simplified Chinese
And the description field value shall be in Simplified Chinese
And all other frontmatter fields shall remain unchanged
```

### Scenario 4.5: Internal links preserved

```gherkin
Given a source file contains a markdown link [link text](/en/getting-started/installation)
When the file is translated to Chinese
Then the link text shall be translated to Chinese
And the link path shall remain unchanged (/en/getting-started/installation)
```

### Scenario 4.6: Mermaid diagrams preserved

```gherkin
Given a source file contains a Mermaid diagram code block
When the file is translated
Then the Mermaid syntax structure shall remain valid
And only human-readable node labels shall be translated
And diagram layout directives (TD, LR, etc.) shall remain unchanged
```

---

## AC-5: Technical Term Consistency

### Scenario 5.1: Product names remain in English

```gherkin
Given the translated content references "MoAI-ADK"
When the content is displayed in any locale
Then "MoAI-ADK" shall appear as-is without translation
And other product names (Claude Code, Nextra, Next.js) shall appear as-is
```

### Scenario 5.2: Command names remain unchanged

```gherkin
Given the translated content references the command "/moai plan"
When the content is displayed in any locale
Then "/moai plan" shall appear exactly as-is
And all other slash commands shall appear in their original English form
```

### Scenario 5.3: Localized terms include English parenthetical

```gherkin
Given the Chinese translation uses a localized term from the glossary
When the term appears for the first time on a page
Then the English original shall appear in parentheses, e.g., "代理 (Agent)"
And subsequent occurrences on the same page may omit the parenthetical
```

---

## AC-6: Build Integrity

### Scenario 6.1: Build passes after each phase

```gherkin
Given Phase 1 (English translation) is complete
When "npm run build" is executed
Then the build shall complete successfully with exit code 0
And no build errors shall be reported
```

### Scenario 6.2: All 220 pages accessible

```gherkin
Given all translation phases are complete
When all 220 pages (55 x 4 locales) are requested
Then each page shall return HTTP 200
And no page shall return 404 or 500
```

### Scenario 6.3: Language switcher functional

```gherkin
Given a user is viewing /en/getting-started/introduction
When the user switches to Chinese (zh) using the language selector
Then the browser shall navigate to /zh/getting-started/introduction
And the page shall display Chinese content
```

---

## AC-7: Unwanted Behavior Prevention

### Scenario 7.1: No machine-translation artifacts

```gherkin
Given a translated page is reviewed
When the prose is read by a native speaker
Then the text shall read naturally without awkward literal translations
And terminology shall be consistent within each page
And sentence structure shall follow target language conventions
```

### Scenario 7.2: Code blocks untouched

```gherkin
Given a translated file contains inline code (`moai init`)
When the file is rendered
Then the inline code shall display the original text unchanged
And no translation shall appear inside backtick-delimited spans
```

### Scenario 7.3: File structure unchanged

```gherkin
Given the original directory structure has 55 MDX files per locale
When all translations are complete
Then each locale directory shall still contain exactly 55 MDX files
And no files shall be added, removed, or renamed
And the directory hierarchy shall match the Korean source exactly
```

### Scenario 7.4: Existing English translations preserved

```gherkin
Given the 13 previously translated English files exist
When Phase 1 (English completion) is executed
Then the 13 files in getting-started/ and core-concepts/ shall not be modified
Unless a specific error is identified and documented
```

---

## Quality Gate Criteria

### Definition of Done

- [ ] All 42 English files translated from Korean source
- [ ] All 55 Chinese files translated from English source
- [ ] All 55 Japanese files translated from English source
- [ ] `npm run build` passes with zero errors
- [ ] All 220 pages return HTTP 200
- [ ] Language switcher navigates correctly between all 4 locales
- [ ] Spot-check: 5 random pages per locale read naturally to a native speaker
- [ ] No Korean text remains in en/, zh/, or ja/ prose sections
- [ ] No English-only prose remains in zh/ or ja/ sections (excluding code blocks)
- [ ] All `_meta.ts` files remain unchanged (already translated)
- [ ] All MDX imports and JSX components preserved identically across locales

### Verification Methods

| Criterion                     | Method                                                |
| ----------------------------- | ----------------------------------------------------- |
| Content language correctness  | Manual spot-check (5 pages per locale)                |
| MDX structure preservation    | Diff import lines between source and translated files |
| Build integrity               | `npm run build` exit code 0                           |
| Page accessibility            | HTTP status check for all 220 URLs                    |
| Language switcher              | Manual navigation test across all locale pairs        |
| Technical term consistency    | Grep for glossary terms across translated files       |
| File count integrity          | `find content/{en,zh,ja} -name "*.mdx" | wc -l` = 165 |
