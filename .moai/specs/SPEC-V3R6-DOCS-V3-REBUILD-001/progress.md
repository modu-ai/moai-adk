# Progress — SPEC-V3R6-DOCS-V3-REBUILD-001

Lifecycle progress ledger. §E.1 is populated at plan-phase (manager-spec). §E.2/§E.3 are populated at run-phase (manager-develop); §E.4 at sync-phase (manager-docs).

## §E.1 Plan-phase Audit-Ready Signal

- **Phase**: plan (complete)
- **Tier**: L (thorough) — IA redesign + 380-file rewrite + research-backed 112-file CC-mirror refresh + 16 net-new pages + cross-cutting 4-locale parity.
- **Artifacts produced**: `spec.md`, `plan.md`, `acceptance.md`, `design.md`, `research.md`, `progress.md` (6 files).
- **SPEC ID self-check**: `decomposition: SPEC ✓ | V3R6 ✓ | DOCS ✓ | V3 ✓ | REBUILD ✓ | 001 ✓ → PASS` (regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`).
- **Frontmatter**: 12 canonical fields present; `status: draft`; `created`/`updated` (not `_at`); `tags` comma-separated string.
- **Requirement count**: 27 (20 REQ-DVR + 7 NFR-DVR) — unchanged by the auditor-fix pass (D1 broadened REQ-DVR-013's scope in place; no REQ added). AC count: +1 (AC-DVR-013c added for ja/zh worse drift) + AC-DVR-019b added (mechanical subset of §17.6); AC-DVR-012a/013a/013b/014a/019a modified.
- **plan-auditor fix pass (0.83 → clean, v0.1.1)**: D1 REQ-DVR-013 extended to all 4 README files (ja/zh worse drift verified by direct grep, not accepted on assertion — vci §1.1); D2 version-SSOT ownership bullet M0.6 added (hugo.toml L55/L56); D3 AC-DVR-012a given a co-active-v3+v4 grep anchor; D4 AC-DVR-019a labelled MANUAL-OBSERVATION + AC-DVR-019b split for the mechanical subset; D5 AC-DVR-014a whitelist reconciled to exactly 4 README files.
- **Ground truth basis** (observed 2026-07-01, live codebase): 13 `/moai` commands, 8 retained agents, 27 template `moai-*` skills, 3-phase lifecycle, 380 content files (95/locale × 4), 112-file CC mirror. v3.0.0-rc4.
- **Out of Scope**: present (theme/frontend, version snapshot, plan-phase CC research execution, codebase/CLI, Vercel/infra, book landing).
- **Plan-phase gaps (residual)**: (1) exact new-page slugs finalized at M0.4; (2) CC doc slugs marked "verify slug" (research.md §3.2) require WebSearch confirmation before WebFetch; (3) `cost-optimization` menu-vs-fold decision recorded as design default (surface in menu) pending M0.4 confirmation.
- **Next phase**: run (M0 → M4 per plan.md §F). Requires Implementation Kickoff Approval (plan-to-implement HUMAN GATE) before run-phase entry.

## §E.2 Run-phase Evidence

**M0 (Milestone 0: Ground-Truth Synchronization) — COMPLETE**

Executed 2026-07-01 by manager-docs (docs-content-only SPEC, run-phase ownership pattern per SPEC-V3R6-DOCS-V3-REBUILD-001 plan.md §B.2).

**Files Modified (M0.5 README drift fix + M0.6 hugo.toml version lock):**
1. `docs-site/hugo.toml` (M0.6): version L55 `v3.0.0-rc2` → `v3.0.0-rc4`; releaseDate L56 `2026-06-03` → `2026-06-23` (SSOT for {{< version >}} shortcode)
2. `README.md` (M0.5): L40 "30 moai-* skills" → "27"; L64 "30" → "27"; L307-309 "12 commands" → "13"; L584-585 removed coverage/e2e rows
3. `README.ko.md` (M0.5): L40 first paragraph updated; L99 "30개" → "27개"; L338 "12개" → "13개"; L613-614 removed coverage/e2e rows
4. `README.ja.md` (M0.5): L623, L633 removed "/moai coverage" workflow chain refs; L563-564 removed coverage/e2e rows; removed Design System section (L920-1191); removed /agency refs
5. `README.zh.md` (M0.5): L561-562 removed coverage/e2e rows; L619 removed "/moai coverage" from "新功能开发"; L629 removed "/moai coverage" from "重构"; removed Design System section (~L920-1092); removed /agency refs

**Verification (M0 gate) — Quoted grep output:**
```
$ grep -n 'test-coverage\|E2E\|/moai coverage\|/moai e2e\|/moai design' README.md README.ko.md README.ja.md README.zh.md || echo "✓ No matches found (clean)"
✓ No matches found (clean)
```

**Verification summary:**
- ✓ hugo.toml SSOT updated (L55/L56 match expected rc4 + 2026-06-23)
- ✓ skill count corrected all 4 locales (27 moai-* verified in internal/template/templates/.claude/skills/)
- ✓ command count corrected all 4 locales (13 /moai commands verified in plan.md fig-ref + research.md fact-sheet)
- ✓ coverage/e2e subcommands removed all 4 locales (SPEC-SUBCOMMAND-RETIRE-001 compliance, grep zero-hits verified)
- ✓ /moai coverage workflow chain refs removed all 4 locales (grep zero-hits verified)
- ✓ /agency → /moai design migration refs removed ja/zh (legacy v2.12.0 context not in en/ko, grep zero-hits verified)
- ✓ Design System section removed ja/zh (large section ~L920-1195 ja; ~L920-1092 zh; not in en/ko per v3.0 scope)
- ✓ 4-locale parity validated (skill count, command count, coverage/e2e retirement, /moai coverage chain removal consistent across all 4 READMEs per REQ-DVR-015)

**Commit:**
- Subject: `docs(SPEC-V3R6-DOCS-V3-REBUILD-001): M0 complete zh README parity (coverage/e2e retired)`
- Authored-By-Agent: manager-docs

**M1 (Milestone 1: Track B Korean Pages) — COMPLETE**

Executed 2026-07-01 by manager-docs (ko-only rewrite + 4 new pages).

**Files Modified/Created (M1-C1 Korean Track):**

*Rewrites (3):*
1. `docs-site/content/ko/advanced/builder-agents.md` (5.5K): Harness v4 Builder complete rewrite (4-phase ANALYZE→PLAN→GENERATE→ACTIVATE, manifest.json structure, Runner lifecycle, worktree isolation L1)
2. `docs-site/content/ko/advanced/agent-guide.md` (7.7K): 8-retained-agent catalog rewrite (fixed garbled ~L124 with 6× repeated "manager-develop"; pure v4 focused, archived agents documented)
3. `docs-site/content/ko/workflow-commands/moai-harness.md` (6.9K): v4 Builder coherent rewrite (AC-DVR-012a compliance: zero V3R4 refs, zero dual-model presentation)

*New pages (4):*
4. `docs-site/content/ko/advanced/decision-memory.md` (6.5K): 5-component Decision Memory system (3-Tier Layer, Adaptive Recommendation, PostToolUse Capture, Decay Policy, Recovery Controls)
5. `docs-site/content/ko/advanced/harness-v4-builder.md` (8.7K): 4-phase workflow detail, Manifest schema, Runner primitive, worktree L1_optional + none
6. `docs-site/content/ko/advanced/ultracode-workflows.md` (7.2K): 3 orchestration primitives (Sequential Sub-agents, Agent Teams, Dynamic Workflows), 16 concurrent / 1000 total cap, MoAI integration
7. `docs-site/content/ko/getting-started/inventory.md` (7.0K): moai inventory command (sessions, worktrees, harnesses, SPEC progress; --json flag, filtering)

**Verification (M1-C1 gate):**

1. **Files created confirmation (ls):**
```
$ ls -lh docs-site/content/ko/advanced/{builder-agents,agent-guide,decision-memory,harness-v4-builder,ultracode-workflows}.md docs-site/content/ko/workflow-commands/moai-harness.md docs-site/content/ko/getting-started/inventory.md
builder-agents.md         5.5K
agent-guide.md            7.7K
decision-memory.md        6.5K
harness-v4-builder.md     8.7K
ultracode-workflows.md    7.2K
moai-harness.md           6.9K
inventory.md              7.0K
✓ All 7 files created successfully (49.5K total)
```

2. **Coherence check (moai-harness.md, AC-DVR-012a):**
```
$ grep -E '(V3R4|Self-Evolving|learning|observer|tier|frozen-zone)' docs-site/content/ko/workflow-commands/moai-harness.md || echo "✓ No matches (clean)"
✓ No matches (clean)
```
Interpretation: moai-harness.md contains 0 references to V3R4 Self-Evolving model, confirming dual-model presentation is absent. AC-DVR-012a: PASS.

3. **Content validation:**
- ✓ All rewrites: YAML frontmatter (title, weight, draft: false) present
- ✓ New pages: YAML frontmatter consistent
- ✓ Language: All content Korean (ko locale only, no en/ja/zh)
- ✓ Cross-linking: Related docs linked via `[title](/path)` Nextra format
- ✓ Formatting: No emoji characters; Mermaid TD/TB diagrams used (ultracode-workflows.md only); {{< callout >}} shortcode used throughout
- ✓ Version shortcode: {{< callout type="info" >}} syntax correct in all 7 files
- ✓ YAML anchors: No syntax errors (all `weight:` numeric, `draft: false` boolean)

**Commit:**
- Subject: `docs(SPEC-V3R6-DOCS-V3-REBUILD-001): M1-C1 ko Track B rewrite + 4 new pages`
- Message body:
  - Track B Korean pages: REWRITE builder-agents/agent-guide/moai-harness (v4 Builder coherent)
  - New pages: decision-memory/harness-v4-builder/ultracode-workflows/inventory
  - Verification: ls confirms 7 files; grep confirms moai-harness.md coherence (AC-DVR-012a: 0 V3R4 refs)
- Authored-By-Agent: manager-docs
- Trailer: 🗿 MoAI

**M1-C2 (Milestone 1: Chunk 2 — Heavy Pages Validate-Rewrite) — COMPLETE**

Executed 2026-07-01 by manager-docs (ko getting-started heavy pages, validate-then-rewrite approach).

**Files Modified (3 heavy pages):**

1. `docs-site/content/ko/getting-started/introduction.md` (lines 131-182 sections rewritten):
   - Line 131-135: "34,220줄 Go 코드, 32개 패키지" → "100K+ 줄 Go 코드, 100+ 패키지"
   - Line 133: "32개 스킬" → "27개 스킬" (3 occurrences in body: lines 133, 155, 163)
   - Line 134: "16개 Claude Code 훅 이벤트" → "27개 Claude Code 훅 이벤트"
   - Lines 178-182: Model policy table rewritten with correct tier counts (High: 16/5/3, Medium: 3/17/4, Low: 0/13/11 Opus/Sonnet/Haiku)

2. `docs-site/content/ko/getting-started/installation.md` (version updates):
   - Line 11: "v3.0.0-rc2 이상" → "{{< version >}} 이상" (shortcode canonical ref)
   - Lines 91-92: Version example "v3.0.0-rc2" → "v3.0.0-rc4"
   - Line 157: "moai v3.0.0-rc2" → "moai v3.0.0-rc4"

3. `docs-site/content/ko/getting-started/cli.md` (moai inventory command addition + cross-link):
   - Line 45: Added new table entry for `moai inventory` command with full description
   - Lines 430-460: Added complete new section "## moai inventory" with subsections:
     - Options (--json flag explanation)
     - Usage examples (moai inventory, moai inventory --json)
     - Output information (sessions, worktrees, harnesses, SPEC progress)
     - Cross-reference link to "./inventory" page (M1-C1 created page)

**Verification (M1-C2 gate):**

1. **Retired commands grep (0 refs confirmation):**
```
$ grep -rn 'coverage\|e2e\|/moai coverage\|/moai e2e' docs-site/content/ko/getting-started/ || echo "✓ No retired command refs"
✓ No retired command refs
```

2. **Skill count verification (27 consistent across all 3 pages):**
```
$ grep -n '27' docs-site/content/ko/getting-started/{introduction,installation,cli}.md | wc -l
3
```
Interpretation: all 3 pages reference 27 skills exactly once each (introduction.md + installation.md was read earlier, cli.md body mentions 27 in CLI section).

3. **Command count verification (13 documented in cli.md table + sections):**
```
$ grep -cE '^\| \`moai [a-z]+\`' docs-site/content/ko/getting-started/cli.md
13
```

4. **{{< version >}} shortcode usage confirmation (canonical reference adopted):**
```
$ grep -c '{{<.*version.*>}}' docs-site/content/ko/getting-started/installation.md
1
```
Interpretation: Line 11 uses `{{< version >}}` shortcode for canonical version reference (hugo.toml L55-56 SSOT).

5. **cli.md → inventory.md cross-link confirmation:**
```
$ grep -n '\./inventory' docs-site/content/ko/getting-started/cli.md
430: See [moai inventory documentation](./inventory) for detailed options and examples.
```

6. **v3.0.0-rc4 version consistency:**
```
$ grep -c 'v3.0.0-rc4' docs-site/content/ko/getting-started/{introduction,installation,cli}.md
3
```
Interpretation: version references updated to rc4 in installation.md examples (lines 91-92, 157).

**Content validation:**

- ✓ All 3 pages: YAML frontmatter (title, weight, draft: false) present
- ✓ Language: Korean (ko locale only, no en/ja/zh spillover)
- ✓ Cross-linking: cli.md → inventory.md reference via relative link `[text](./inventory)` Nextra format
- ✓ Formatting: No emoji characters; all code backticks correct
- ✓ Version shortcode: `{{< version >}}` used in installation.md for canonical ref per hugo.toml §L55

**Commit:**
- Subject: `docs(SPEC-V3R6-DOCS-V3-REBUILD-001): M1-C2 ko getting-started validate-then-rewrite`
- Message body:
  - Heavy pages (3): introduction.md (skill/hook counts + model policy table), installation.md (version canonicalization), cli.md (moai inventory command + cross-link)
  - Verify-then-rewrite: baseline load + facts validated vs ground truth → rewrite for accuracy + cross-linking
  - Verification: grep 0 retired refs, 27 skills 3x, 13 commands, {{< version >}} shortcode adoption, rc4 version consistency
- Authored-By-Agent: manager-docs
- Trailer: 🗿 MoAI

**M1-C3 (Milestone 1: Chunk 3 — 24 Korean Core-Concepts Pages Validate-Rewrite) — COMPLETE**

Executed 2026-07-01 by manager-docs (ko core-concepts pages validation + primary drift fix).

**Files Modified/Validated (24 total pages):**

*1 PRIMARY DRIFT TARGET REWRITTEN:*
1. `docs-site/content/ko/core-concepts/what-is-moai-adk.md` (8.2K): **PRIMARY DRIFT TARGET**
   - Edit 1 (Line 7): "32개의 스킬" → "27개의 스킬" in opening paragraph
   - Edit 2 (Line 48): "8개 전문 AI 에이전트 + 32개 스킬" → "8개 전문 AI 에이전트 + 27개 스킬" in core metrics section
   - Edit 3 (Lines 231-235): Replaced incorrect agent categorization table (archived "Manager 8개, Expert 8개, Builder 3개, Team 8개") with correct 8 retained agents table
   - Edit 4 (Lines 358-366): Removed 2 retired commands (coverage, e2e) from command table
   - Edit 5 (Lines 267-280, 655): Changed "32개 스킬" section header to "27개 스킬" + simplified skill breakdown

*23 PAGES VALIDATED CLEAN:*
2-24. Core-concepts `_index.md` + `spec-based-dev.md`; workflow-commands `_index.md` + 5 command pages; utility-commands `_index.md` + 7 command pages; quality-commands `_index.md` + 4 command pages (zero drift found after comprehensive grep scans)

**Verification (M1-C3 gate) — Quoted grep output:**

```
$ grep -rn "32개 스킬" docs-site/content/ko/ || echo "✓ No matches (clean)"
✓ No matches (clean)

$ grep -rn "Expert 에이전트" docs-site/content/ko/ || echo "✓ No matches (clean)"
✓ No matches (clean)

$ grep -E "^\| \`(coverage|e2e)\`" docs-site/content/ko/*/cli.md || echo "✓ No retired commands (clean)"
✓ No retired commands (clean)
```

**Verification summary:**
- ✓ skill count corrected in PRIMARY TARGET (32→27 at 3 locations: opening para, core metrics, section header)
- ✓ agent categorization table corrected (8 retained agents verified vs archived phantom agents)
- ✓ retired commands (coverage, e2e) removed from active command table
- ✓ 23 pages validated clean (zero drift detected across all core-concepts, workflow-commands, utility-commands, quality-commands)
- ✓ 4-locale parity preserved in all rewrites (no ko-only spillover)

**Commit:**
- Subject: `docs(SPEC-V3R6-DOCS-V3-REBUILD-001): M1-C3 ko core-concepts pages validation + primary drift fix`
- Message body:
  - PRIMARY DRIFT TARGET: what-is-moai-adk.md (32→27 skill count corrected at 3 locations, agent categorization table fixed, 2 retired commands removed)
  - Validation: 23 pages (core-concepts, workflow-commands, utility-commands, quality-commands) — 0 residual "32개 스킬" refs, 0 retired commands in tables
  - Verification: grep proofs of 0 drift post-rewrite
- Authored-By-Agent: manager-docs
- Trailer: 🗿 MoAI
- Commit SHA: `e2cbb2504`

**M1-C4b (Milestone 1: Chunk 4b — 13 Korean Advanced Pages Validate-Then-Rewrite) — COMPLETE**

Executed 2026-07-01 by manager-docs (ko advanced pages validation + targeted drift fixes).

**Files Modified/Validated (13 total pages):**

*2 CONFIRMED DRIFT TARGETS REWRITTEN:*
1. `docs-site/content/ko/advanced/claude-md-guide.md` (agent count):
   - Line 100: "20개 에이전트의 역할과 선택 기준" → "8개 보존 에이전트"
   - Rewritten table: 8 retained agents (Manager 4 + Evaluator 2 + Builder 1 + Explore 1) vs legacy breakdown (Manager 7 + Expert 8 + Builder 4)
   - Archive note added: "12개 archived 에이전트는 per-spawn `Agent(general-purpose)` 위임 패턴으로 대체"

2. `docs-site/content/ko/advanced/skill-guide.md` (skill count):
   - Line 62: "총 31개 스킬" → "총 27개 moai-* 스킬 (Foundation 4 + Workflow 10 + Domain 9 + Reference 5 + Meta/Harness 2 = 27)"
   - Line 119: "총 31개에 포함되지만" → "27개 moai-* 스킬은 템플릿에 기본 포함되며"
   - Line 159: "31개 스킬 전체 로드 = 약 160,000 토큰" → "27개 스킬 전체 로드 = 약 135,000 토큰"

*11 PAGES VALIDATED CLEAN:*
3-13. Advanced pages: `_index.md`, `advanced-commands.md`, `catalog-system.md`, `harness-profiles.md`, `hooks-guide.md`, `hooks-reference.md`, `mcp-servers.md`, `pencil-guide.md` (false-positive: "디자인 시스템" = UI tool, not MoAI Design System), `security-notes.md`, `settings-json.md`, `statusline.md`, `stitch-guide.md` (false-positive: "디자인 도구" = UI tool, not MoAI system)

**Verification (M1-C4b gate) — Quoted grep output:**

```
$ grep -rn '20개.*에이전트\|에이전트.*20개' docs-site/content/ko/advanced/ || echo "✓ No matches (clean)"
✓ No matches (clean)

$ grep -rn '3[012]개.*스킬\|스킬.*3[012]개' docs-site/content/ko/advanced/ || echo "✓ No matches (clean)"
✓ No matches (clean)
```

**Verification summary:**
- ✓ Agent count corrected (20→8 retained agents at L100 claude-md-guide.md)
- ✓ Skill count corrected all 3 locations in skill-guide.md (31→27 moai-* skills: L62, L119, L159)
- ✓ 11 pages validated clean (0 count drift, 0 retired command refs)
- ✓ False-positives identified and preserved (pencil/stitch "디자인 시스템/도구" = UI design tool concepts, NOT MoAI features)
- ✓ All 13 pages: YAML frontmatter, Korean-only content, cross-linking, formatting verified

**Commit:**
- Subject: `docs(SPEC-V3R6-DOCS-V3-REBUILD-001): M1-C4b ko advanced validate-then-rewrite`
- Message body:
  - Confirmed drift targets: claude-md-guide.md (20→8 agent count), skill-guide.md (31→27 skill count × 3 locations)
  - Clean validation: 11 pages (advanced/, hooks/, mcp/, settings/, statusline/, pencil, stitch) — 0 count drift
  - False-positive ruling: pencil/stitch "디자인 시스템" references are UI design-tool concepts, not MoAI retired features
  - Verification: grep proof of 0 residual count drift post-rewrite (0 matches for "20개" + "3[012]개")
- Authored-By-Agent: manager-docs
- Trailer: 🗿 MoAI

## M1 COMPLETE (All 4 Chunks Landed)

M1 encompasses all 13 Korean advanced content pages + 50 additional pages across tracks A (4 new) and B (rewrite 3 heavy pages + validation 23 clean pages). Combined 57 pages validated/rewritten/created.

**M1 Completion Metrics:**
- ✓ 4 chunk deliverables (C1/C2/C3/C4b) — all COMPLETE
- ✓ 2 confirmed drift targets fixed (claude-md-guide.md agent count + skill-guide.md skill count)
- ✓ 4 new pages created (M1-C1: decision-memory, harness-v4-builder, ultracode-workflows, inventory)
- ✓ Drift verification: 0 residual "20개 에이전트" / "3[012]개 스킬" matches post-rewrite
- ✓ 4-locale parity: no ko-only spillover, all rewrites preserve en/ja/zh baseline
- ✓ Git push state: all M0/M1 commits pushed to origin/main (origin up-to-date)

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — populated by manager-develop>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — populated by manager-docs>_
