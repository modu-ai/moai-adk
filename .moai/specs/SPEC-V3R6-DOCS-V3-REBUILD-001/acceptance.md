# Acceptance Criteria — SPEC-V3R6-DOCS-V3-REBUILD-001

All criteria are testable via shell command or observable artifact. `<L>` denotes the 4 locales `{ko,en,ja,zh}`. Commands assume CWD = repo root unless noted.

## §D AC Matrix (per requirement)

### REQ-DVR-001 — Version accuracy

- **AC-DVR-001a**: `grep -n 'v3.0.0-rc4' docs-site/hugo.toml` returns the `params.version` line; `grep -rn 'rc2\|rc3' docs-site/hugo.toml` returns zero matches.
- **AC-DVR-001b**: `grep -rln 'rc2\|릴리스 후보 2\|34,220\|34220' docs-site/content/*/_index.md docs-site/content/*/getting-started/introduction.md docs-site/content/*/getting-started/installation.md` returns zero files across all 4 locales.

### REQ-DVR-002 — Command count accuracy

- **AC-DVR-002a**: Each locale's command reference lists exactly 13 `/moai` commands; `grep -rn '12 total\|12개 명령\|12 commands' docs-site/content/` returns zero matches.
- **AC-DVR-002b**: `grep -rniE '/moai (coverage|e2e|design|brain|security)\b' docs-site/content/` returns zero matches (retired subcommands absent). (A prose mention that these were retired is permitted only in a changelog/migration context, not as an active command.)

### REQ-DVR-003 — Agent count accuracy

- **AC-DVR-003a**: Each locale documents 8 retained agents (7 MoAI-custom + Explore); the agent list contains manager-spec, manager-develop, manager-docs, manager-git, plan-auditor, sync-auditor, builder-harness, Explore.
- **AC-DVR-003b**: `grep -rniE 'expert-(backend|frontend|security|devops|performance|refactoring)' docs-site/content/` returns zero matches presenting these as active agents.

### REQ-DVR-004 — Skill count accuracy

- **AC-DVR-004a**: The doc-facing skill number equals the template source count: `ls -1d internal/template/templates/.claude/skills/moai-* | wc -l` (= 27) matches the number printed in `docs-site/content/<L>/getting-started/introduction.md`.
- **AC-DVR-004b**: `grep -rn '30 .moai-\|32개 스킬\|32 skills\|30 skills' docs-site/content/ README.md README.ko.md` returns zero matches.

### REQ-DVR-005 — Lifecycle accuracy

- **AC-DVR-005a**: Each locale's lifecycle description names exactly the 3 phases (plan → run → sync).
- **AC-DVR-005b**: `grep -rniE 'Mx (phase|단계|フェーズ|阶段)|4-phase|4단계' docs-site/content/` returns zero matches presenting Mx as a separate 4th lifecycle phase (a "Mx tag as a sync-phase cross-cutting concern" mention is permitted).

### REQ-DVR-006 — Full rebuild scope (validate-then-rewrite)

- **AC-DVR-006a**: All 380 content files (95/locale × 4) show modification within the rebuild commit range (`git log --oneline` for the SPEC shows edits touching every section).
- **AC-DVR-006b**: The 5 already-accurate sections (core-concepts, workflow-commands, utility-commands, quality-commands, multi-llm) were validated against the M0 fact-sheet before rewrite (design.md §C validate-then-rewrite log present).

### REQ-DVR-007 — New IA blueprint

- **AC-DVR-007a**: `data/menu/main.yaml` includes every content section, including `cost-optimization`; no content section exists on disk without a corresponding menu entry.
- **AC-DVR-007b**: design.md §A carries the approved IA blueprint/sitemap and the menu ↔ content reconciliation table.

### REQ-DVR-008 — CC mirror refresh

- **AC-DVR-008a**: All 28 `claude-code/` pages per locale exist and each page's content reflects the run-phase research delta (research.md §3 traceability note present per page).
- **AC-DVR-008b**: The latest-CC topics (goal, workflows, scheduled-tasks, agent-teams, agent-view, plugins, mcp, hooks) each have a refreshed page in all 4 locales.

### REQ-DVR-009 — New feature pages

- **AC-DVR-009a**: 4 new topics exist in all 4 locales (16 files): Decision Memory (`moai preference`), `moai inventory`, Harness v4 Builder, `/effort ultracode`. `find docs-site/content -path '*preference*' -o -path '*inventory*' -o -path '*harness-v4*' -o -path '*ultracode*'` confirms 4 × 4 = 16 files (path naming per design.md §A).
- **AC-DVR-009b**: Each new page appears in `data/menu/main.yaml` with all 4 locale name entries.

### REQ-DVR-010 — builder-agents rewrite

- **AC-DVR-010a**: `advanced/builder-agents.md` (all 4 locales) describes the Harness v4 Builder (4-phase ANALYZE→PLAN→GENERATE→ACTIVATE + manifest Runner + conditional worktree isolation).
- **AC-DVR-010b**: `grep -rn 'builder-skill\|builder-plugin\|builder-agent\b' docs-site/content/*/advanced/builder-agents.md` returns zero matches (obsolete 3-builder model removed).

### REQ-DVR-011 — agent-guide correction

- **AC-DVR-011a**: `advanced/agent-guide.md` has no garbled repeated-token line: `grep -n 'manager-develop, manager-develop' docs-site/content/*/advanced/agent-guide.md` returns zero matches.
- **AC-DVR-011b**: The page documents dynamic-team spawning (`Agent(general-purpose)` with domain context) and has no stray archived-agent reference presented as active.

### REQ-DVR-012 — harness reconcile

- **AC-DVR-012a**: `workflow-commands/moai-harness.md` (all 4 locales) presents one coherent harness model consistent with the v4 Builder. Binary anchor: no single file may present BOTH an active "v3 learning system" AND an active "v4 Builder" as the current model. Mechanical check — for each file, if `grep -liE 'v3 (learning|학습|学习|学習)' <file>` matches AND `grep -liE 'v4 (Builder|빌더|构建|ビルダー)' <file>` matches, then the v3 reference MUST be inside a superseded/historical clause (e.g. "superseded by v4" / "이전 v3"); a file presenting both as active fails the AC. Expected: zero files with a co-active v3+v4 claim.

### REQ-DVR-013 — README drift correction

- **AC-DVR-013a**: `grep -niE '/moai (coverage|e2e|design)\b|\| .(coverage|e2e|design). ' README.md README.ko.md README.ja.md README.zh.md` returns zero active-subcommand rows for coverage/e2e/design across all 4 README files.
- **AC-DVR-013b**: README command count reads 13 (not 12) and skill count reads 27; `grep -niE '12 total\|30 .moai-\|30 skills\|32 skill' README.md README.ko.md README.ja.md README.zh.md` returns zero matches across all 4 README files.
- **AC-DVR-013c** (ja/zh-specific worse drift, verified present at plan-phase by grep): `grep -niE '/moai design|/agency|Design System' README.ja.md README.zh.md` returns zero matches presenting `/moai design` (retired) or `/agency` as active; `grep -niE '/moai coverage' README.ja.md README.zh.md` returns zero matches inside workflow chains; and stale `v2.6.0`-as-current references are reconciled. The retired `/moai design` Design System section (~L926-1061) and the `/agency → /moai design` v2.12.0 rows (L57/L63) are removed.

### REQ-DVR-014 — Theme preservation

- **AC-DVR-014a**: `git diff --name-only` for the rebuild touches zero files under `docs-site/themes/`, `docs-site/static/*.css`, `docs-site/layouts/`. Only these paths change: `docs-site/content/`, `docs-site/data/menu/main.yaml`, `docs-site/hugo.toml` (version), and exactly the 4 repo-root README files (`README.md`, `README.ko.md`, `README.ja.md`, `README.zh.md`). The `README*.md` whitelist is scoped to exactly those 4 files — consistent with REQ-DVR-013's 4-file enumeration (D5 reconciliation).

### REQ-DVR-015 — 4-locale parity

- **AC-DVR-015a**: `for l in ko en ja zh; do find docs-site/content/$l -name '*.md' | wc -l; done` prints an identical count for all 4 locales.
- **AC-DVR-015b**: Relative file paths match across locales (`scripts/docs-i18n-check.sh` path-parity check passes).

### REQ-DVR-016 — Menu 4-locale rebuild

- **AC-DVR-016a**: Every `main.yaml` entry has non-empty `ko`, `en`, `ja`, `zh` name fields; every `ref` resolves to an existing page in all 4 locales.

### REQ-DVR-017 — Build gate

- **AC-DVR-017a**: `cd docs-site && hugo --minify` exits 0 with zero warnings in stderr.
- **AC-DVR-017b**: `docs-site/public/sitemap.xml` exists after build; `grep -r 'v3.0.0-rc4' docs-site/public/` returns matches; `grep -rc 'vv3\|v v3' docs-site/public/` returns zero (no double-v render regression).

### REQ-DVR-018 — i18n check gate

- **AC-DVR-018a**: `bash scripts/docs-i18n-check.sh` exits 0 (file count/path parity, non-empty titles, H1 present, glossary preserved).

### REQ-DVR-019 — Deploy verify

- **AC-DVR-019a** [MANUAL-OBSERVATION — NOT mechanically grep/CI-verifiable]: The §17.6 checklist is confirmed by human observation, NOT by an automated assertion — Mermaid TD diagrams render (visual), language switcher navigates locale→locale (browser interaction), Vercel preview serves all 4 locale home pages (browser). This AC MUST be reported as manually observed with a reviewer note; it MUST NOT be claimed as mechanically PASS (verification-claim-integrity §1 — no unobserved-verification-claim).
- **AC-DVR-019b** (the mechanically-verifiable subset of §17.6): `grep -l 'adk.mo.ai.kr' docs-site/static/robots.txt docs-site/static/llms.txt` confirms both files carry the canonical domain, AND `docs-site/public/sitemap.xml` exists after build. This subset is the only part of §17.6 that is grep/CI-verifiable.

### REQ-DVR-020 — Research capture (plan-phase deferral)

- **AC-DVR-020a**: `research.md` §3 contains the CC-latest research PLAN with concrete target URLs and a per-CC-page fetch mapping.
- **AC-DVR-020b**: No Claude Code documentation was fetched during plan-phase (the plan-phase deliverable is the PLAN only; git shows no CC-content edits at plan-phase).

### NFR acceptance

- **AC-DVR-N01** (URL): `grep -rniE 'docs\.moai-ai\.dev|adk\.moai\.com|adk\.moai\.kr([^.]|$)' docs-site/` returns zero matches.
- **AC-DVR-N02** (Mermaid): `grep -rniE 'flowchart LR|graph LR' docs-site/content/` returns zero matches.
- **AC-DVR-N03** (Emphasis spacing): No `**…(…)**` emphasis-glued-paren occurrences introduced (spot-check + i18n lint).
- **AC-DVR-N04** (No emoji): `scripts/docs-i18n-check.sh` emoji check (or equivalent grep) returns zero decorative-emoji hits in body content.
- **AC-DVR-N05** (Glossary): "MoAI-ADK" and other glossary terms appear untranslated in all 4 locales.
- **AC-DVR-N06** (Vercel): `docs-site/vercel.json` project binding + root + preset unchanged from the pre-rebuild baseline (`git diff docs-site/vercel.json` shows only domain-string corrections, if any).
- **AC-DVR-N07** (No snapshot): `find docs-site/content -maxdepth 2 -type d -name 'v3'` returns nothing (no snapshot dir created).

## §D.1 Given-When-Then Scenarios

### Scenario 1 — Command count reconciliation (REQ-DVR-002, -013)

- **Given** the live command dir `.claude/commands/moai/` contains 13 command files and README claims "12 total",
- **When** the rebuild corrects the command reference in docs and README,
- **Then** every command-count statement reads 13, the 13 commands are enumerated, and no retired subcommand (coverage/e2e/design/brain/security) appears as active — verified by `AC-DVR-002a/b` and `AC-DVR-013a/b/c` (013c covers the ja/zh-only retired `/moai design` section + `/agency` refs).

### Scenario 2 — Skill count derived from template source (REQ-DVR-004)

- **Given** the local dev project has 27 `moai-*` + 2 `harness-*` skill dirs but the template source ships only 27 `moai-*`,
- **When** M0.2 reconciles the doc-facing number from `internal/template/templates/.claude/skills/`,
- **Then** every doc surface prints 27 (not 30, not 32), and the 2 user-owned `harness-*` dirs are excluded — verified by `AC-DVR-004a/b`.

### Scenario 3 — CC mirror refreshed without fabrication (REQ-DVR-008, -020)

- **Given** the plan-phase produced a research PLAN (research.md §3) but no CC docs were fetched yet,
- **When** run-phase executes the research and rewrites the 112 CC-mirror files,
- **Then** each refreshed page traces back to a fetched CC source, and no feature claim exists without research evidence — verified by `AC-DVR-008a/b` and `AC-DVR-020a/b` (verification-claim-integrity §1).

### Scenario 4 — 4-locale parity holds after new pages (REQ-DVR-009, -015, -016)

- **Given** 4 new topics are authored in ko,
- **When** they propagate to en/ja/zh and the menu is rebuilt,
- **Then** all 4 locales have identical file counts, all 16 new files exist, and every menu `ref` resolves in every locale — verified by `AC-DVR-009a/b`, `AC-DVR-015a/b`, `AC-DVR-016a`.

## §D.2 Edge Cases

- **EC-1** — A page accurate in ko but with a stale en translation: validate-then-rewrite must re-validate EACH locale independently, not assume en inherits ko accuracy.
- **EC-2** — `hugo.toml params.version` already carries a `v` prefix; the `{{< version >}}` shortcode must not double-render (`vv3.0.0`). Guard: `AC-DVR-017b` double-v scan.
- **EC-3** — A CC feature deprecated between rc2 and rc4: the CC mirror must remove it, not merely refresh it (research delta note must flag removals).
- **EC-4** — `cost-optimization` section exists in content but is absent from the menu: IA rebuild must either surface it in the menu or fold it into a parent section (design.md §A decision).
- **EC-5** — A retired subcommand name appearing legitimately inside a migration/changelog note: the AC-DVR-002b grep must scope to active-command context, not prose history.

## §D.3 Severity Classification

- **MUST-FIX (blocks close)**: REQ-DVR-001, -002, -003, -004, -005 (accuracy foundation), -013 (README), -015 (parity), -017, -018 (build/i18n gates); NFR-DVR-001, -002.
- **SHOULD-FIX**: REQ-DVR-006, -007, -008, -009, -010, -011, -012, -016, -019, -020; NFR-DVR-003, -004, -005.
- **NICE-TO-HAVE**: cosmetic wording polish beyond factual accuracy.

## §D.4 Traceability

Every REQ-DVR-NNN maps to at least one AC-DVR-NNN and at least one milestone (plan.md §F). Every NFR maps to an AC-DVR-Nxx and the §17 doctrine SSOT.

## Definition of Done

- [ ] All MUST-FIX acceptance criteria pass (§D.3).
- [ ] `hugo --minify` builds clean (zero warnings) with sitemap + rc4 (AC-DVR-017).
- [ ] `scripts/docs-i18n-check.sh` passes (AC-DVR-018).
- [ ] 4-locale file-count parity confirmed (AC-DVR-015a).
- [ ] Forbidden-URL scan zero matches; Mermaid LR zero matches (AC-DVR-N01/N02).
- [ ] All 16 new-feature pages present + menu-linked in 4 locales (AC-DVR-009).
- [ ] CC mirror refreshed with research traceability, no fabricated claims (AC-DVR-008).
- [ ] All 4 README files corrected — retired coverage/e2e/design removed, 13 commands, 27 skills; ja/zh additionally have the retired `/moai design` Design System section + `/agency` refs removed (AC-DVR-013a/b/c).
- [ ] `docs-site/hugo.toml` version SSOT set to `v3.0.0-rc4` by plan M0.6 (AC-DVR-001a).
- [ ] Theme untouched, and only the 4 enumerated README files changed at repo root (AC-DVR-014a).
- [ ] §17.6 deploy checklist: manual-observation items reported as observed (AC-DVR-019a), mechanical subset verified (AC-DVR-019b).
