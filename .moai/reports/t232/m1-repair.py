#!/usr/bin/env python3
"""M1 data repair for SPEC-ZONE-REGISTRY-RESYNC-001.

Applies the repair map (clause re-spans + anchor re-points + file re-points)
to the zone-registry, verifying EVERY constraint mechanically BEFORE writing:
  - each live clause occurs on exactly ONE line of its file in BOTH trees
  - each anchor resolves to a heading slug in BOTH trees (6-step slug rule)
  - no entry's file: is the registry itself (D13)
  - id set 101->101, zone/zone_class/canary_gate byte-identical (D2)
  - retired 4 entries untouched (option C)
Writes both mirrors byte-identically.
"""
import re
import sys

ROOT = '/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t232'
REG_REL = '.claude/rules/moai/core/zone-registry.md'
TMPL_PREFIX = 'internal/template/templates/'
REG_LOCAL = ROOT + '/' + REG_REL

MC = '.claude/rules/moai/core/moai-constitution.md'

# id -> {'file':..., 'anchor':..., 'clause':...}  (only changed fields present)
R = {}

def fix(i, **kw):
    R[i] = kw

# --- CLAUDE.md §1 -> moai-constitution.md migrations (D4) ---
fix('CONST-V3R2-008', file=MC, anchor='#response-language',
    clause='All user-facing responses MUST be in the user\'s conversation_language.')
fix('CONST-V3R2-009', file=MC, anchor='#parallel-execution',
    clause='Execute all independent tool calls in parallel when no dependencies exist.')
fix('CONST-V3R2-010', file=MC, anchor='#output-format',
    clause='XML tags are reserved for agent-to-agent data transfer')
fix('CONST-V3R2-011', file=MC, anchor='#output-format',
    clause='Use Markdown for all user-facing communication')

# --- CLAUDE.md in-place clause re-spans ---
fix('CONST-V3R2-007', clause='MoAI is the Strategic Orchestrator for Claude Code.')
fix('CONST-V3R2-012', clause='Every question directed at the user MUST be asked via AskUserQuestion.')
fix('CONST-V3R2-013', clause='When intent is unclear, conduct a Socratic interview before execution')
fix('CONST-V3R2-014', clause='Before non-trivial code, explain the approach + which files change + why; get user approval')
fix('CONST-V3R2-015', clause='When modifying 3+ files, split into logical units (TodoList), execute file-by-file, analyze dependencies before parallel execution, report progress per unit')
fix('CONST-V3R2-016', clause='After coding, provide potential-issue list (edge cases, error/concurrency scenarios), suggested test cases, known limitations/assumptions, additional-validation recommendations')
fix('CONST-V3R2-017', clause='Write a failing reproduction test first; confirm it fails; challenge the diagnosed root cause once')
fix('CONST-V3R2-018', clause='Every question directed at the user MUST be asked via AskUserQuestion. Free-form prose questions in response text are prohibited.')
fix('CONST-V3R2-019', clause='`AskUserQuestion`, `TaskCreate`, `TaskUpdate`, `TaskList`, `TaskGet` are **deferred tools** — schemas NOT loaded at session start')
fix('CONST-V3R2-020', clause='subagents run in the background by default (the runtime chooses foreground only when it needs the result; every permission prompt still surfaces in the main session); MoAI does not set `background:` — the retained safeguard is concurrency, not backgrounding')

# --- moai-constitution.md entries ---
fix('CONST-V3R2-002', clause='All code changes must pass TRUST 5 validation')
fix('CONST-V3R2-025', clause='AskUserQuestion is the sole user-facing question channel')
fix('CONST-V3R2-026', clause='used ONLY by the MoAI orchestrator (subagents must never prompt users)')
fix('CONST-V3R2-027', clause='Canonical reference: `.claude/rules/moai/core/askuser-protocol.md` § Channel Monopoly / § ToolSearch Preload Procedure / § Socratic Interview Structure / § Option Description Standards')
fix('CONST-V3R2-028', anchor='#opus-5-48-prompt-philosophy',
    clause='Principle 4 — fewer subagents by default**: 4.7+ does not auto-spawn')
fix('CONST-V3R2-029', anchor='#opus-5-48-prompt-philosophy',
    clause='Principle 5 — fewer tool calls by default**: specify when and why each')
fix('CONST-V3R2-030', clause='Before implementing anything non-trivial, list assumptions explicitly and wait for user confirmation')
fix('CONST-V3R2-031', clause='When encountering inconsistencies, conflicting requirements, or unclear specifications, STOP and surface the confusion before proceeding')
fix('CONST-V3R2-032', clause='Point out issues directly when an approach has clear problems. Sycophancy is a failure mode.')
fix('CONST-V3R2-033', clause='Actively resist overcomplexity. The natural tendency of code generation is toward over-engineering. Resist it.')
fix('CONST-V3R2-034', clause='Touch only what you were asked to touch. Drive-by refactors create noise and risk regressions.')
fix('CONST-V3R2-035', clause='Every task requires evidence of completion.')

# --- other rule-doc clause re-spans ---
fix('CONST-V3R2-001', anchor='#plan-phase',
    clause='Create comprehensive specification using EARS format.')
fix('CONST-V3R2-003', anchor='#scope',
    clause='This rule applies to all agents working with source code in the supported programming languages')
fix('CONST-V3R2-004', clause='All instruction documents must be in English:')
fix('CONST-V3R2-005', clause='All slash command files MUST be thin routing wrappers (under 20 LOC body).')
fix('CONST-V3R2-006', clause='`AskUserQuestion` is the **only** user-facing question channel')
fix('CONST-V3R2-037', clause='Preload `AskUserQuestion` via `ToolSearch(query:')
fix('CONST-V3R2-038', clause='AskUserQuestion is reserved exclusively for the MoAI orchestrator')
fix('CONST-V3R2-043', clause='Agents are invoked through MoAI\'s natural language delegation pattern')
fix('CONST-V3R2-044', clause='The retained safeguard is **concurrency, not backgrounding**')
fix('CONST-V3R2-049', clause='The reviewer mode operates as a fresh-judgment auditor')
fix('CONST-V3R2-062', clause='moai-domain-copywriting MUST adhere to brand voice, tone, and terminology from brand-voice.md')
fix('CONST-V3R2-063', clause='moai-domain-brand-design MUST use brand color palette, typography, and visual language from visual-identity.md')
fix('CONST-V3R2-064', clause='expert-frontend MUST implement design tokens derived from brand context')
fix('CONST-V3R2-066', clause='MUST auto-load human-authored design documents (research.md, system.md, spec.md) when present and not _TBD_')
fix('CONST-V3R2-068', clause='`moai-workflow-design` continues to write machine-generated artifacts to `.moai/design/`')
fix('CONST-V3R2-069', clause='Reserved file paths (canonical list): `tokens.json`, `components.json`, `assets/`, `import-warnings.json`, `brief/BRIEF-*.md`')
fix('CONST-V3R2-070', clause='Token budget for auto-loading is bounded by `.moai/config/sections/design.yaml` `design_docs.token_budget`; when the key is absent, the system MUST default to 20000')
fix('CONST-V3R2-072', clause='When both are present, brand constraints win on conflict.')
fix('CONST-V3R2-150', clause='The orchestrator MUST emit a paste-ready resume message when ANY of these conditions activate')
fix('CONST-V3R2-151', clause='Resume message MUST follow this exact 6-block structure, **bounded by cut-line markers**')
fix('CONST-V3R2-152', clause='Save the message to a memory project entry. Filename pattern: `project_<epic>_<spec>_<status>.md`')
fix('CONST-V3R2-153', clause='`✂` symbol (U+2702 BLACK SCISSORS) is **preserved verbatim across all locales** — never translate or substitute')
fix('CONST-V3R5-001', clause='Subagents MUST NOT invoke `AskUserQuestion`')
fix('CONST-V3R5-022', clause='Operational threshold is **model-specific**. Larger windows tolerate higher percentage utilization before stall risk dominates')
fix('CONST-V3R5-023', clause='When usage crosses the model-specific threshold:')
fix('CONST-V3R5-024', clause='The next action MUST be `/clear` — no further large work in the current session')
fix('CONST-V3R5-025', clause='Pre-clear announcement: When the orchestrator detects accumulated context (input + output) approaching the model-specific threshold')
fix('CONST-V3R5-026', clause='Resume message format: include all of the following so the next session is self-sufficient')
fix('CONST-V3R5-027', clause='Step 1 (plan) MUST execute in main checkout on BOTH routes. NO L2/L3 worktree at this step')
fix('CONST-V3R5-028', clause='Step 4 (cleanup) applies to **Route B only**. It MUST happen ONLY after BOTH run AND sync PRs are merged')
fix('CONST-V3R5-029', clause='AskUserQuestion is invoked by the **orchestrator only**.')
fix('CONST-V3R5-035', clause='Skill body BODP gate MUST follow the askuser-protocol Socratic structure: `(권장)` first, ≤4 options, conversation_language match')
fix('CONST-V3R5-037', clause='Comma-separated string ONLY (`tools: Read, Write, Edit`). YAML arrays NOT supported')
fix('CONST-V3R5-038', clause='allowed-tools format: [ZONE:Evolvable] [HARD] Comma-separated string ONLY. Space-separated values are PROHIBITED')
fix('CONST-V3R5-039', clause='When the work happened inside a worktree, the resume message MUST prepend **Block 0 (cwd anchoring)** before the standard 6-block structure')
fix('CONST-V3R5-040', clause='They SHALL route web search to `mcp__web_search_prime__webSearchPrime`, web fetch to `mcp__web_reader__webReader`, and image reading to a `mcp__zai-mcp-server__*` vision tool')
fix('CONST-V3R5-041', clause='While a session is GLM-backed, the built-in `WebSearch` / `WebFetch` tools and `Read`-on-an-image-file are **PROHIBITED** because they route through the 529-prone `api.z.ai/api/anthropic` gateway and the base64→422 image path')
fix('CONST-V3R6-001', clause='Stop/PostToolUse hooks SHOULD exit 0 (allow the turn to end / the tool call to proceed) rather than exit 2 (block), so that recovery turns are NOT placed into the `error → stop-hook-blocks → retry → error` loop')

# --- ci-autofix-protocol.md: 10 entries (anchor re-points; 9 clause re-spans) ---
fix('CONST-V3R5-004', anchor='#entry-condition',
    clause='The CI auto-fix loop MUST be entered ONLY when the orchestrator hands off')
fix('CONST-V3R5-005', anchor='#iteration-cap',
    clause='The auto-fix loop MUST attempt at most **3 iterations**')
fix('CONST-V3R5-006', anchor='#iteration-cap',
    clause='The AskUserQuestion at iteration > 3 MUST be a blocking call')
fix('CONST-V3R5-007', anchor='#patch-commit-rule-no-force-push',
    clause='Every auto-fix patch MUST be applied as a **new commit** on the PR branch')
fix('CONST-V3R5-008', anchor='#askuserquestion-boundary',
    clause='AskUserQuestion is the **exclusive user interaction channel**')
fix('CONST-V3R5-009', anchor='#askuserquestion-boundary',
    clause='The orchestrator MUST preload AskUserQuestion via')
fix('CONST-V3R5-010', anchor='#semantic-failure-no-auto-patch',
    clause='Semantic failures (data race, deadlock, panic, test assertion failure) MUST')
fix('CONST-V3R5-011', anchor='#secrets-and-credentials-protection',
    clause='The auto-fix loop MUST NOT modify `.env`, `.env.*`, credentials files')
fix('CONST-V3R5-012', anchor='#audit-log-requirement')
fix('CONST-V3R5-013', anchor='#ci-infrastructure-preservation',
    clause='The auto-fix loop MUST NOT modify CI watch infrastructure scripts or')

RETIRED = {'CONST-V3R2-021', 'CONST-V3R2-022', 'CONST-V3R2-023', 'CONST-V3R2-024'}


def slug(h):
    h = h.strip().lstrip('#').strip().replace('`', '').lower()
    h = re.sub(r'[^a-z0-9\s-]', '', h)
    return '#' + re.sub(r'\s+', '-', h.strip())


def heading_slugs(path):
    slugs = set()
    inf = False
    for ln in open(path, encoding='utf-8'):
        if ln.strip().startswith('```'):
            inf = not inf
            continue
        if not inf and ln.startswith('#'):
            slugs.add(slug(ln))
    return slugs


def hit_count(path, needle):
    """grep -F -c semantics: number of LINES containing needle."""
    n = 0
    for ln in open(path, encoding='utf-8'):
        if needle in ln:
            n += 1
    return n


def main():
    lines = open(REG_LOCAL, encoding='utf-8').read().split('\n')

    # parse entries (field -> value) from original
    entries = []   # list of dicts in order
    cur = None
    for ln in lines:
        m = re.match(r'^- id: (\S+)', ln)
        if m:
            cur = {'id': m.group(1)}
            entries.append(cur)
            continue
        if cur is None:
            continue
        m = re.match(r'^  (\w+): (.*)$', ln)
        if m:
            k, v = m.group(1), m.group(2).strip()
            if v.startswith('"') and v.endswith('"') and len(v) >= 2:
                v = v[1:-1]
            elif v.startswith("'") and v.endswith("'") and len(v) >= 2:
                v = v[1:-1]
            cur[k] = v

    ids = [e['id'] for e in entries]
    assert len(ids) == 101, 'entry count %d != 101' % len(ids)
    retired_actual = [e['id'] for e in entries
                      if e.get('clause', '').lstrip().startswith('[SUPERSEDED')]
    assert set(retired_actual) == RETIRED, 'retired set mismatch: %s' % retired_actual
    for i in R:
        assert i in ids, 'repair target %s not in registry' % i
        assert i not in RETIRED, 'repair target %s is retired (must stay untouched)' % i

    # apply repairs to entry dicts (post state)
    post = {}
    for e in entries:
        d = dict(e)
        if e['id'] in R:
            d.update(R[e['id']])
        post[e['id']] = d

    # D2: zone/zone_class/canary_gate unchanged; id set unchanged
    for e in entries:
        p = post[e['id']]
        for f in ('zone', 'zone_class', 'canary_gate'):
            assert e[f] == p[f], '%s %s changed' % (e['id'], f)

    # D13: no self-reference
    for i, p in post.items():
        assert p['file'] != REG_REL, '%s self-references registry' % i

    # verify all live clauses + anchors in BOTH trees
    tmpl_reg = ROOT + '/' + TMPL_PREFIX + REG_REL
    assert open(REG_LOCAL, 'rb').read() == open(tmpl_reg, 'rb').read(), 'mirrors not identical before repair'

    file_cache = {}
    problems = []
    for e in entries:
        i = e['id']
        p = post[i]
        retired = i in RETIRED
        for prefix, label in ((ROOT, 'local'), (ROOT + '/' + TMPL_PREFIX, 'tmpl')):
            path = prefix + '/' + p['file']
            if path not in file_cache:
                try:
                    file_cache[path] = (open(path, encoding='utf-8').read(), heading_slugs(path))
                except OSError:
                    file_cache[path] = (None, set())
            src, slugs = file_cache[path]
            if src is None:
                problems.append('%s [%s] missing file %s' % (i, label, p['file']))
                continue
            if p['anchor'] not in slugs:
                problems.append('%s [%s] anchor %s unresolved' % (i, label, p['anchor']))
            if not retired:
                n = hit_count(path, p['clause'])
                if n != 1:
                    problems.append('%s [%s] clause hit=%d :: %s' % (i, label, n, p['clause'][:60]))
                if '"' in p['clause'] or len(p['clause']) < 20:
                    problems.append('%s clause has quote or <20 chars' % i)
    if problems:
        print('VERIFY FAIL (%d):' % len(problems))
        for x in problems:
            print('  ' + x)
        sys.exit(1)

    # rewrite registry lines
    out = []
    cur_id = None
    for ln in lines:
        m = re.match(r'^- id: (\S+)', ln)
        if m:
            cur_id = m.group(1)
            out.append(ln)
            continue
        if cur_id is None or cur_id not in R:
            out.append(ln)
            continue
        m = re.match(r'^  (file|anchor|clause): ', ln)
        if m and m.group(1) in R[cur_id]:
            k = m.group(1)
            v = R[cur_id][k]
            if k == 'file':
                out.append('  file: ' + v)
            else:
                out.append('  %s: "%s"' % (k, v))
        else:
            out.append(ln)
    new_text = '\n'.join(out)

    # post-write invariants: entry count, field deltas only
    n_ids = len(re.findall(r'^- id: CONST-', new_text, re.M))
    assert n_ids == 101, 'post count %d' % n_ids
    for e in entries:
        p = post[e['id']]
        for f in ('zone', 'zone_class', 'canary_gate'):
            assert e[f] == p[f]

    open(REG_LOCAL, 'w', encoding='utf-8').write(new_text)
    open(tmpl_reg, 'w', encoding='utf-8').write(new_text)
    assert open(REG_LOCAL, 'rb').read() == open(tmpl_reg, 'rb').read()

    n_clause = sum(1 for i in R if 'clause' in R[i])
    n_anchor = sum(1 for i in R if 'anchor' in R[i])
    n_file = sum(1 for i in R if 'file' in R[i])
    print('OK: repaired entries=%d (clause edits=%d, anchor edits=%d, file re-points=%d)'
          % (len(R), n_clause, n_anchor, n_file))
    print('OK: all live clauses hit exactly once in BOTH trees; all 101 anchors resolve in BOTH trees')
    print('OK: retired 4 untouched; ids/zone/zone_class/canary_gate unchanged; mirrors byte-identical')


main()
