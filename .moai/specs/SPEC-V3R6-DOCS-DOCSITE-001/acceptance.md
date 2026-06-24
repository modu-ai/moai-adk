# Acceptance — SPEC-V3R6-DOCS-DOCSITE-001

> Mechanically verifiable acceptance criteria for docs-site 4-locale content reconciliation.
> 모든 AC는 `grep`, `wc`, `diff` 명령으로 4 locale 트리에 걸쳐 기계적 검증 가능.

## §A. Verification Philosophy

각 AC는 두 축을 동시에 검증한다:
1. **Per-fact correctness** — 사실이 docs-truth.md가 인용한 primary source와 일치
2. **4-locale parity** — 사실이 en/ko/ja/zh 모두에서 균등하게 반영

부정 grep (stale facts 제거 확인) + 긍정 grep (canonical facts 존재 확인) 쌍으로 구성.

## §B. Test Environment

- Working directory: repo root (`/Users/goos/MoAI/moai-adk-go`)
- docs-site root: `docs-site/content/{en,ko,ja,zh}/`
- 4 locale 트리: en, ko, ja, zh (각 105페이지, 총 420페이지)
- Parity tool: `scripts/docs-i18n-check.sh`
- Baseline commit: `7666bd178` (README-001 close 이후)

## §C. Severity Model

- **Critical** (MUST-FIX): stale fact가 활성 context에 존재 (예: "28 agents" active catalog)
- **High**: canonical fact 누락 (예: `glm-5.2[1m]` 부재)
- **Medium**: parity gap (예: 1 locale만 table 보유, 3 locale은 stub)
- **Low**: adjacent drift (out of scope — research.md 기록만)

---

## §D. Acceptance Criteria Matrix

### AC-DOCSITE-001 — Agent count: stale "28" eradicated, 4-locale (Critical)

**Given** docs-site pages that reference the agent-catalog count
**When** the reconciliation runs per-locale stale-count greps
**Then** no docs-site `core-concepts/what-is-moai-adk.md` page carries stale "28 agents" / "28개 ... 에이전트" / "28个 ... Agent" / "28(の|個) ... エージェント" in any of the 4 locales

**Mechanical check (MUST PASS)** — per-locale stale "28" with digit-boundary protection (prevents false-positive on "128", "280", and prevents the "8 inside 28" substring bug caught in D1):

```bash
# Per-locale stale-28 count. Each regex tested live against tree 8108d4311 (2026-06-17).
# PRE-FIX expected counts (the drift this AC must eradicate):
#   en=0, ko=2, ja=5, zh=1   ← live-grep verified D1
# POST-FIX expected: all 0
for loc in en ko ja zh; do
  c=$(grep -cE '(^|[^0-9])28\s*(specialized\s*)?agents?|(^|[^0-9])28\s*개.*(전문\s*)?에이전트|(^|[^0-9])28\s*个.*Agent|(^|[^0-9])28\s*(の|個).*エージェント' \
    docs-site/content/$loc/core-concepts/what-is-moai-adk.md 2>/dev/null)
  echo "$loc stale-28: $c"
done
# Expected POST-FIX: 0 per locale
```

**Inline live-grep evidence (PRE-FIX baseline, tree `8108d4311`)**:
```
en stale-28: 0    ← clean (no edit needed for stale-28 axis)
ko stale-28: 2    ← L226 active prose "28개의 전문 에이전트", L666 code-block "28개 AI 에이전트 정의"
ja stale-28: 5    ← L7, L48, L226, L496, L670 (primary prose-drift locale)
zh stale-28: 1    ← L674 code-block "28个AI Agent定义" (prose L226 already "8个专业Agent")
```

**Why the digit-boundary prefix `(^|[^0-9])`**: without it, the ja positive-parity regex `8\s*(つの|個).*エージェント` matched the "8" inside "28個" (D4 investigation), and the stale-28 regex would over-count on numbers like "128", "280". The `(^|[^0-9])` prefix requires the "28" or "8" to be preceded by start-of-line or a non-digit, which is the correct token-boundary semantics. This was the root cause of 6/10 AC failures in iter-1 (Verify-Don't-Assume violation — regexes authored from imagination, never run against the live tree).

**Coverage**: REQ-DOCSITE-001, REQ-DOCSITE-005

---

### AC-DOCSITE-002 — Archived agents historical framing + mermaid structural integrity (Critical)

**Given** docs-site pages that name archived agents AND mermaid diagrams that render the agent catalog
**When** the reconciliation inspects prose and mermaid `Managers` subgraph structure
**Then** (a) archived agents appear ONLY in explicit "archived/consolidated" historical context, AND (b) the mermaid `Managers` subgraph contains only the 4 legitimate manager nodes (manager-spec, manager-develop, manager-docs, manager-git) — NO phantom duplicate `M6/M7` slots, NO `sync-auditor` misclassified as a Manager

**Mechanical check (a) — archived-name historical framing (MUST PASS)**:
```bash
for loc in en ko ja zh; do
  c=$(grep -nE 'manager-strategy|manager-quality|manager-brain|manager-project|claude-code-guide|expert-backend|expert-frontend|expert-security|expert-devops|expert-performance|expert-refactoring' \
    docs-site/content/$loc/core-concepts/what-is-moai-adk.md \
    | grep -vE 'archived|consolidat|아카이브|보관|アーカイブ|保管|归档|破棄|历史|履歴|historical' \
    | wc -l)
  echo "$loc archived-unframed: $c"
done
# Expected: 0 per locale
```

**Mechanical check (b) — mermaid phantom-duplicate-slot detection (MUST PASS, D3 replacement)**:

The iter-1 archived-name-in-mermaid grep returned 0 for ALL locales (archived names are NOT in any mermaid block — that check was a no-op). The REAL mermaid drift is **structural**: ko/ja/zh mermaid `Managers (8)` subgraph uses `M6["manager-spec ..."]` and `M7["manager-docs ..."]` phantom duplicate slots to inflate the count to 8, while misclassifying `M5["sync-auditor ..."]` as a Manager (sync-auditor is an evaluator). en does NOT have this drift. The 8-retained flat catalog is: 4 managers (spec/develop/docs/git) + 2 evaluators (plan-auditor/sync-auditor) + 1 builder (builder-harness) + 1 Explore built-in — NOT 8 managers.

```bash
# Detect phantom duplicate M6/M7 manager-spec/docs slots (structural drift, D3).
# Live-grep verified tree 8108d4311 (2026-06-17):
#   en=0, ko=2, ja=2, zh=2   ← phantom duplicate slots present
# POST-FIX expected: all 0 (mermaid Managers subgraph must use unique M1..M4 IDs only)
for loc in en ko ja zh; do
  c=$(grep -cE 'M6\["manager-spec|M6\["manager-docs|M7\["manager-spec|M7\["manager-docs' \
    docs-site/content/$loc/core-concepts/what-is-moai-adk.md 2>/dev/null)
  echo "$loc phantom-M6/M7: $c"
done
# Expected POST-FIX: 0 per locale
```

**Inline live-grep evidence (PRE-FIX baseline)**:
```
en phantom-M6/M7: 0   ← en mermaid is clean (uses unique IDs)
ko phantom-M6/M7: 2   ← M6["manager-spec ..."], M7["manager-docs ..."] duplicate slots
ja phantom-M6/M7: 2   ← same phantom pattern
zh phantom-M6/M7: 2   ← same phantom pattern
```

**Why the iter-1 check was a no-op**: `grep -cE 'manager-strategy|...|expert-refactoring'` on the mermaid block returns 0 because archived names were already removed from the diagram — the drift is that the REMOVAL was papered over with phantom duplicate slots rather than honestly reducing the count to the real 4-manager + 4-other structure. The fix in M2 must rewrite the ko/ja/zh mermaid `Managers` subgraph to list exactly 4 unique manager nodes + a separate `Evaluators` subgraph (plan-auditor, sync-auditor) + `Builder` node + `Explore` built-in, matching the en mermaid structure and the actual `.claude/agents/moai/*.md` (7 files) + Explore catalog.

**Coverage**: REQ-DOCSITE-002, REQ-DOCSITE-005

---

### AC-DOCSITE-003 — GLM tier-models: stale glm-5.1 eradicated + canonical glm-5.2[1m] present, 4-locale (Critical)

**Given** docs-site `multi-llm/` pages that enumerate the GLM→Claude tier mapping
**When** the reconciliation runs per-locale greps for stale and canonical GLM model tokens
**Then** (a) no `multi-llm/` page in ANY of the 4 locales states stale `glm-5.1`/`GLM-5.1`, AND (b) after M3 expansion, each locale's `multi-llm/` carries the canonical `glm-5.2[1m]` token (matching `internal/config/defaults.go` `DefaultGLMHigh`/`DefaultGLMOpus`)

**Mechanical check (a) — stale glm-5.1 eradicated, 4-locale (MUST PASS)**:

The iter-1 check only grepped `{en,ko,ja,zh}/multi-llm/` as a glob and asserted total = 0, which masks which locale carries the stale token. D10 re-survey: the stale `glm-5.1` is concentrated in **ko only** (1 file: `ko/multi-llm/_index.md` L17, L23). en/ja/zh have NO glm-5.1 (they're near-empty stubs with no tier table at all — covered by AC-004 parity). Per-locale check:

```bash
# Live-grep verified tree 8108d4311 (2026-06-17):
#   en=0, ko=1, ja=0, zh=0   ← files with stale glm-5.1/GML-5.1
# POST-FIX expected: all 0
for loc in en ko ja zh; do
  c=$(grep -rcE 'glm-5\.1|GLM-5\.1' docs-site/content/$loc/multi-llm/ 2>/dev/null | grep -v ':0$' | wc -l)
  echo "$loc files-with-glm-5.1: $c"
done
# Expected POST-FIX: 0 per locale
```

**Mechanical check (b) — canonical glm-5.2 present, 4-locale (MUST PASS, post-M3)**:

The iter-1 "positive check" grepped only ko and asserted ≥1. Live re-survey: **NO locale currently carries `glm-5.2`** — `grep -rcE 'glm-5\.2' {en,ko,ja,zh}/multi-llm/` returns 0 files for every locale (ko has stale glm-5.1, en/ja/zh are stubs). The positive parity target is therefore POST-M3 (after M3 expands en/ja/zh stubs and corrects ko), NOT pre-fix. Each locale must carry the canonical `glm-5.2[1m]` token.

```bash
# POST-M3 expected: each locale ≥ 1 file with canonical glm-5.2[1m]
# PRE-FIX baseline (tree 8108d4311): all 0 — this is the gap M3 must close
for loc in en ko ja zh; do
  c=$(grep -rcE 'glm-5\.2' docs-site/content/$loc/multi-llm/ 2>/dev/null | grep -v ':0$' | wc -l)
  echo "$loc files-with-glm-5.2: $c"
done
# Expected POST-M3: ≥ 1 per locale
```

**Inline live-grep evidence (PRE-FIX baseline, tree `8108d4311`)**:
```
en files-with-glm-5.1: 0   en files-with-glm-5.2: 0   ← stub, no table
ko files-with-glm-5.1: 1   ko files-with-glm-5.2: 0   ← STALE glm-5.1 in _index.md L17,L23
ja files-with-glm-5.1: 0   ja files-with-glm-5.2: 0   ← stub, no table
zh files-with-glm-5.1: 0   zh files-with-glm-5.2: 0   ← stub, no table
```

**Why the iter-1 premise was inverted**: iter-1 assumed the drift was "wrong glm-5.1 in ko, correct glm-5.2 elsewhere". The live tree shows the drift is "stale glm-5.1 in ko, NO glm-5.2 ANYWHERE". The fix (M3) must (i) correct ko `glm-5.1` → `glm-5.2[1m]`, AND (ii) expand en/ja/zh stubs with the tier-models table containing `glm-5.2[1m]`. The 4-locale loop (not a ko-only grep) is the correct verification shape.

**Coverage**: REQ-DOCSITE-003

---

### AC-DOCSITE-004 — GLM facts 4-locale parity (High)

**Given** the GLM tier-models facts
**When** the reconciliation compares `multi-llm/_index.md` content across 4 locales
**Then** all 4 locales carry structurally equivalent GLM tier-models content (all carry the table, OR all uniformly omit it)

**Mechanical check (MUST PASS)** — structural parity:
```bash
# Each locale's multi-llm/_index.md must have comparable size (within 50% of the largest)
largest=$(for loc in en ko ja zh; do
  wc -l < docs-site/content/$loc/multi-llm/_index.md
done | sort -rn | head -1)
for loc in en ko ja zh; do
  size=$(wc -l < docs-site/content/$loc/multi-llm/_index.md)
  ratio=$(echo "scale=2; $size / $largest" | bc)
  echo "$loc: $size lines (ratio $ratio)"
done
# Expected: all 4 locales ratio ≥ 0.50 (no locale is a near-empty stub after reconciliation)
```

**`scripts/docs-i18n-check.sh` check (MUST PASS)**:
```bash
bash scripts/docs-i18n-check.sh 2>&1 | grep -i 'multi-llm' | grep -iE 'fail|missing|parity'
# Expected: empty (no multi-llm parity failures)
```

**Coverage**: REQ-DOCSITE-004

---

### AC-DOCSITE-005 — Agent catalog 4-locale parity: positive "8" + stale "28" (Critical)

**Given** the 8-retained-agents fact and archived historical framing
**When** the reconciliation compares `core-concepts/what-is-moai-adk.md` across 4 locales
**Then** (a) all 4 locales carry a positive "8 retained/specialized agents" claim (locale-native idiom, ≥1 per locale), AND (b) all 4 locales carry 0 stale "28 agents" / "52 skills"

**Mechanical check (a) — positive "8" parity, locale-native idioms (MUST PASS, D4 rewrite)**:

The iter-1 positive-parity regex used invented CJK translations (`8つの保持`, `8个保留`, `8保留`) that do NOT appear in the live files. D4 re-survey: actual idiomatic forms are ko `8개의 전문 AI 에이전트` / `8개 에이전트`, zh `8个专业AI Agent` / `8个Agent`, ja `8つの専門エージェント` (target state — currently 0 in ja), en `8 retained agents` / `8 specialized AI agents`. The digit-boundary prefix `(^|[^0-9])` is MANDATORY to prevent the "8 inside 28" substring false-positive (D4 investigation: the iter-1 regex matched "28個" because `\s*` allows zero whitespace and `28個` contains `8個`).

```bash
# Live-grep verified tree 8108d4311 (2026-06-17):
#   en=2, ko=4, ja=0, zh=5   ← positive "8" count (digit-boundary protected)
# POST-FIX expected: each locale ≥ 1 (ja must reach ≥1 after M2 introduces the positive form)
for loc in en ko ja zh; do
  c=$(grep -cE '(^|[^0-9])8\s*(retained|specialized)\s*agents?|(^|[^0-9])8\s*개.*(전문\s*)?(AI\s*)?에이전트|(^|[^0-9])8\s*个.*(专业\s*)?(AI\s*)?Agent|(^|[^0-9])8\s*(つの|個).*(専門\s*)?エージェント' \
    docs-site/content/$loc/core-concepts/what-is-moai-adk.md 2>/dev/null)
  echo "$loc positive-8: $c"
done
# Expected POST-FIX: each locale ≥ 1
```

**Inline live-grep evidence (PRE-FIX baseline)**:
```
en positive-8: 2   ← L7 "8 specialized AI agents", L226 "8 retained agents"
ko positive-8: 4   ← L7, L24, L48, L498 (mixed with 2 stale-28 — ko is partial-done)
ja positive-8: 0   ← ja has ZERO positive form (all 5 are stale "28") — worst drift locale
zh positive-8: 5   ← L7, L24, L48, L226, L500 (prose done; 1 code-block stale-28 remains)
```

**Mechanical check (b) — stale "28" parity (MUST PASS, D1 per-locale loop)**:

Same digit-boundary-protected regex as AC-001, scoped to `core-concepts/what-is-moai-adk.md`. See AC-001 for the per-locale regex and PRE-FIX baseline (en=0, ko=2, ja=5, zh=1). POST-FIX expected: 0 per locale. The "52 skills" check is retained but noted as lower-priority (skill count is out-of-scope ADJ-1; "52" appears only in ja L48 alongside the stale "28").

```bash
for loc in en ko ja zh; do
  c=$(grep -cE '(^|[^0-9])28\s*(specialized\s*)?agents?|(^|[^0-9])28\s*개.*(전문\s*)?에이전트|(^|[^0-9])28\s*个.*Agent|(^|[^0-9])28\s*(の|個).*エージェント|(^|[^0-9])52\s*(skills?|個.*スキル|个.*技能|개.*스킬)' \
    docs-site/content/$loc/core-concepts/what-is-moai-adk.md 2>/dev/null)
  echo "$loc stale-28-or-52: $c"
done
# Expected POST-FIX: 0 per locale
```

**Coverage**: REQ-DOCSITE-001, REQ-DOCSITE-002, REQ-DOCSITE-005

---

### AC-DOCSITE-006 — CLI /moai command count: 17 (High)

**Given** docs-site pages that enumerate the `/moai` slash-command set
**When** the reconciliation inspects command-count claims
**Then** no page states a stale `/moai` command count (historical "10", "12", etc.)

**Mechanical check (SHOULD PASS)**:
```bash
# Stale /moai command counts must be empty
grep -rEln '/moai .*(10|11|12|13|14|15|16) (commands?|subcommands?|slash)' \
  docs-site/content/{en,ko,ja,zh}/ | wc -l
# Expected: 0
```

**Coverage**: REQ-DOCSITE-006

---

### AC-DOCSITE-007 — SPEC status enum + frontmatter 4-locale (Medium)

**Given** docs-site pages that enumerate the SPEC status lifecycle or frontmatter fields
**When** the reconciliation inspects those pages
**Then** the 8-value status enum and 12 required frontmatter fields match `internal/spec/status.go` and `internal/spec/lint.go`

**Mechanical check (SHOULD PASS)** — status enum:
```bash
# Pages enumerating the status lifecycle must carry all 8 values
for loc in en ko ja zh; do
  # Find pages that mention 'draft' + 'planned' + 'implemented' (status-lifecycle pages)
  for f in $(grep -rlE 'draft.*planned|planned.*draft' docs-site/content/$loc/ --include='*.md' 2>/dev/null); do
    hits=$(grep -cE 'draft|planned|in-progress|implemented|completed|superseded|archived|rejected' "$f")
    echo "$f: $hits status keywords"
  done
done
# Expected: status-lifecycle pages carry all 8 values (≥8 keyword hits, modulo translation)
```

**Coverage**: REQ-DOCSITE-007

---

### AC-DOCSITE-008 — FAQ 4-locale model-assignment table parity (Medium)

**Given** the 8-retained-agents model-assignment table in `getting-started/faq.md`
**When** the reconciliation compares faq.md across 4 locales
**Then** all 4 locales carry the model-assignment markdown TABLE with ≥4 retained-agent rows (manager-spec / manager-develop / manager-docs / manager-git at minimum)

**Mechanical check (SHOULD PASS) — table-row detection (D5 rewrite)**:

The iter-1 keyword regex (`retained agents?|8 retained|model assignment|...`) returned ≥1 for ALL 4 locales, providing zero discrimination — it matched heading keywords, not actual table content. D5 re-survey: the markdown table itself (with `| manager-spec | ... |` rows) ALREADY exists in all 4 locales (awk row-count = 4 each). The real verification is row-presence, not keyword-presence. The AC is therefore downgraded from "add missing table" to "verify table-row parity" (D-007 severity Low).

```bash
# Live-grep verified tree 8108d4311 (2026-06-17):
#   en=4, ko=4, ja=4, zh=4   ← table rows per locale (already present)
# POST-FIX expected: each locale ≥ 4 (manager-spec/develop/docs/git rows)
for loc in en ko ja zh; do
  c=$(awk '/model.*assign|모델.*할당|モデル.*割り当て|模型.*分配|8\s*(个|保留|retained|개)/{f=1} f&&/^\|.*manager-(spec|develop|docs|git)/{n++} END{print n+0}' \
    docs-site/content/$loc/getting-started/faq.md 2>/dev/null)
  echo "$loc faq table-rows: $c"
done
# Expected: each locale ≥ 4
```

**Inline live-grep evidence (PRE-FIX baseline)**:
```
en faq table-rows: 4   ← manager-spec/develop/docs/git rows present
ko faq table-rows: 4   ← present
ja faq table-rows: 4   ← present (iter-1 incorrectly claimed ja missing — Verify-Don't-Assume violation)
zh faq table-rows: 4   ← present (iter-1 incorrectly claimed zh missing)
```

**Coverage**: REQ-DOCSITE-001, REQ-DOCSITE-005 (faq reinforcement)

---

### AC-DOCSITE-009 — i18n parity tool PASS (High)

**Given** the existing `scripts/docs-i18n-check.sh` parity tool
**When** the reconciliation completes and runs the tool
**Then** the tool exits 0 (or reports no parity failures in the facts-bearing pages)

**Mechanical check (MUST PASS)**:
```bash
bash scripts/docs-i18n-check.sh
# Expected: exit 0, or non-zero exit only for pre-existing out-of-scope drift (research.md §C ADJ-1)
```

**Coverage**: All REQs (parity umbrella)

---

### AC-DOCSITE-010 — No archived-agent active-context regression (Critical)

**Given** the reconciliation edits
**When** post-edit grep runs across all 4 locale trees
**Then** no NEW active-context references to archived agents were introduced

**Mechanical check (MUST PASS)** — diff-guard:
```bash
# Every archived-agent mention in the diff must be in a -archived/consolidat context line
git diff HEAD -- docs-site/content/ | \
  grep -E '^\+.*manager-strategy|^\+.*manager-quality|^\+.*expert-backend|^\+.*expert-frontend|^\+.*expert-security|^\+.*expert-devops|^\+.*expert-performance|^\+.*expert-refactoring' | \
  grep -vE 'archived|consolidat|아카이브|보관|アーカイブ|保管|归档|破棄|历史|履歴|migration|rejection' | \
  wc -l
# Expected: 0 (no added line introduces an archived agent as active)
```

**Coverage**: REQ-DOCSITE-002 (regression guard)

---

### AC-DOCSITE-011 — Language count: 16 (not 18), 4-locale (Medium, D6)

**Given** docs-site pages that state the supported-language count
**When** the reconciliation runs per-locale greps for stale "18 languages"
**Then** no docs-site page carries "18 languages" / "18개 언어" / "18言語" / "18种语言" — all must state 16, matching CLAUDE.local.md §15 and FROZEN rule CONST-V3R2-004

**Mechanical check (SHOULD PASS) — stale "18 languages" eradicated, 4-locale**:

```bash
# Live-grep verified tree 8108d4311 (2026-06-17):
#   en=1 occurrence, ko=3, ja=1, zh=0   ← "18 languages" stale claims
# POST-FIX expected: all 0
for loc in en ko ja zh; do
  c=$(grep -rE '(^|[^0-9])18\s*(languages?|개.*언어|言語|个.*语言|種.*言語)' \
    docs-site/content/$loc/ 2>/dev/null | wc -l)
  echo "$loc 18-languages-occurrences: $c"
done
# Expected POST-FIX: 0 per locale
```

**Inline live-grep evidence (PRE-FIX baseline)**:
```
en 18-languages-occurrences: 1   ← what-is-moai-adk.md L216 "18 languages supported"
ko 18-languages-occurrences: 3   ← L49 "18개 프로그래밍 언어", L216 "18개 언어 지원" ×2 contexts
ja 18-languages-occurrences: 1   ← L216 "18言語サポート"
zh 18-languages-occurrences: 0   ← zh already clean on this axis
```

**Why in-scope (unlike "31 skills")**: the 16-language count is a FROZEN HARD rule (CONST-V3R2-004, `.claude/rules/moai/development/coding-standards.md` § Language Policy). Leaving "18" in reconciled pages preserves a contradiction with a canonical FROZEN rule. "31 skills" is out-of-scope because skill count is uncoupled (not a FROZEN rule). This is a clean per-locale `18 → 16` text substitution (no IA change, no structural edit).

**Coverage**: REQ-DOCSITE-008

---

## §D.1..§D.7 Severity / Traceability

| AC ID | Severity | Covers REQ | Covers Milestone |
|-------|----------|------------|------------------|
| AC-001 | Critical | 001, 005 | M2 |
| AC-002 | Critical | 002, 005 | M2 |
| AC-003 | Critical | 003 | M3 |
| AC-004 | High | 004 | M3 |
| AC-005 | Critical | 001, 002, 005 | M2 |
| AC-006 | High | 006 | M4 |
| AC-007 | Medium | 007 | M4 |
| AC-008 | Medium | 001, 005 | M5 (verify-only — table already present) |
| AC-009 | High | (all) | M6 |
| AC-010 | Critical | 002 | M6 (regression guard) |
| AC-011 | Medium | 008 | M4 (D6 — 18→16 language reconciliation) |

## §E. Edge Cases

1. **"researcher" string** — 16 files mention "researcher". 대부분은 비-에이전트 일반 명사(researcher role, the researcher writes...)이다. archived 에이전트로서의 `researcher`만 정정 대상. per-file review 필요 (research.md §C D-003-note).
2. **"31 skills"** — 인접 drift, **OUT OF SCOPE** (skill count는 FROZEN rule이 아니며 uncoupled). research.md에 기록만 하고 수정 금지.
3. **"18 languages"** — **IN SCOPE** per REQ-DOCSITE-008 / AC-011 (D6). 16-language count는 FROZEN rule CONST-V3R2-004이므로 방치 시 모순 잔류. M4에서 4-locale `18 → 16` 조정.
4. **en 이미 correct한 페이지** — en `core-concepts/what-is-moai-adk.md` prose는 이미 "8 retained" + archived historical framing. 단 en은 mermaid 구조는 clean하나 (AC-002(b) phantom-M6/M7 = 0), ko/ja/zh 3개 locale은 phantom M6/M7 structural drift 보유 (M2에서 en 구조를 템플릿으로 재작성).
5. **multi-llm/_index.md 4-locale content 격차** — ko는 2680 bytes (stale glm-5.1 포함), en/ja/zh는 208 bytes (stub, table 부재). parity를 위해 ko 정정(glm-5.1→glm-5.2[1m]) + en/ja/zh 확장을 동시 수행. IA 변경 없이 내용 사실만 채운다.
6. **faq model-assignment table** — iter-1이 ja/zh "missing table"로 보고했으나 D5 live-grep 결과 **4-locale 모두 table 보유** (row-count = 4 each). AC-008은 verify-only로 downgrade.

## §F. Quality Gate Criteria

run-phase 종료 시 아래 전부 PASS:
- [ ] AC-001 ~ AC-011 전부 PASS (Critical AC은 0 fail, Medium AC은 ≤2 debt 허용)
- [ ] `scripts/docs-i18n-check.sh` facts-bearing sections parity PASS
- [ ] `moai spec audit --json` 본 SPEC lint 0 findings
- [ ] 4 locale × 6 docs-truth axes parity regression 없음

## §G. Definition of Done

1. 본 SPEC의 모든 Critical AC이 PASS
2. 4 locale 모두 docs-truth 6 axes (5 original + language count)에 대해 동일한 사실을 담고 있음
3. `scripts/docs-i18n-check.sh` exit 0 (또는 out-of-scope drift만 잔류)
4. research.md drift inventory (D-001..D-009 + D-001b/D-004b) 전부 resolved 또는 explicitly out-of-scope 처리
5. archived-agent active-context regression 없음 (AC-010)
6. plan-auditor independent audit PASS (≥0.85 skip-eligible OR ≥0.75 with documented debt)
