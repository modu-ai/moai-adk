"""t1257 role-naming inventory (document layer). Read-only.

Usage: python3 inv.py <scratchpad> <outdir>
Reads the git-tracked file list from <scratchpad>/files.txt (produced by
`git ls-files`), scans document-layer surfaces, and writes TSVs to <outdir>.
Heuristic per-match classification; per-line disposition is a run-phase job.
"""
import re, sys, os, collections

SP, OUT = sys.argv[1], sys.argv[2]
os.makedirs(OUT, exist_ok=True)
files = open(os.path.join(SP, 'files.txt')).read().split()

TEXT_EXT = ('.md', '.tmpl', '.toml', '.yaml', '.yml', '.json', '.txt', '.sh')

def surface(f):
    if f.endswith('.go'):
        return None
    if not f.endswith(TEXT_EXT) and os.path.basename(f) not in ('CLAUDE.md', 'AGENTS.md'):
        return None
    T = 'internal/template/templates/'
    if f.startswith(T):
        rest = f[len(T):]
        if rest.startswith('.claude/'):
            return 'T:.claude/' + rest.split('/')[1]
        if rest.startswith('.codex/'):
            return 'T:.codex'
        if rest.startswith('.agents/'):
            return 'T:.agents'
        if rest.startswith('.moai/'):
            return 'T:.moai'
        return 'T:root'
    if f.startswith('.claude/'):
        seg = f.split('/')[1]
        if seg in ('rules', 'agents', 'skills', 'output-styles', 'commands'):
            return 'L:.claude/' + seg
        return None
    if f in ('CLAUDE.md', 'AGENTS.md', 'CLAUDE.local.md'):
        return 'L:root'
    if f.startswith('.moai/config/'):
        return 'L:.moai/config'
    if f.startswith('docs-site/content/'):
        return 'docs:' + f.split('/')[2]
    if re.fullmatch(r'README(\.(ko|ja|zh))?\.md', f):
        return 'README'
    return None

TOK = {
    'lead': r"\blead(?:s|'s)?\b",
    'leader': r"\bleaders?(?:'s)?\b",
    'lane': r"\blanes?(?:'s)?\b",
    'worker': r"\bworkers?(?:'s)?\b",
    'companion': r"\bcompanions?\b",
    'foreman': r"\bforem[ae]n\b",
    'deputy': r"\bdeput(?:y|ies)\b",
    'coordinator': r"\bcoordinators?\b",
}
TOKRE = {k: re.compile(v, re.I) for k, v in TOK.items()}

PLAIN_LEAD = re.compile(r"\blead(?:s)?\s+(?:to|into|the|with|a|an|you|it|each|users?|reviewers?|readers?|me|us|them|nowhere|back|directly|straight)\b|\blead[- ]time\b|\blead-in\b|\bleads? by\b", re.I)
PLAIN_WORKER = re.compile(r"\b(?:service|web|background|thread|pool|cloudflare)\s+workers?\b|\bworkers?\s+(?:threads?|pool|process(?:es)?)\b|\bworker_threads\b", re.I)
PLAIN_LANE = re.compile(r"\b(?:swim|fast|slow)\s*lanes?\b", re.I)
SUBAGENT_WORKER = re.compile(r"\bleaf[- ]workers?\b|\bworker\s+(?:sub-?agents?|agents?)\b|\bsub-?agent\s+workers?\b", re.I)
DOC_COMPANION = re.compile(r"(?:detail|lazy|reference|sidecar)[ -]companion|companion[ -](?:file|doc|document|rule|sidecar)|companion (?:to|of) `|lazy-companion|\bcompanion\s*:\s*`", re.I)
AGENT_TEAMS = re.compile(r"\bteam(?:mate)?s?\b|Agent Teams|\bCG\b|moai cg|\bGLM\b|leader orchestrat|team lead", re.I)
KANBAN_CTX = re.compile(r"kanban|factory|\blanes?\b|\bcards?\b|-k\b|-f\b|board|dispatch", re.I)
HIST_LINE =re.compile(r"SUPERSEDED|~~|\bdeprecated\b|\blegacy\b|\bretired\b|\bformerly\b|\bwas renamed\b|\brenamed\b|\bHISTORY\b|폐기|旧|舊|レガシー|구 표기|旧式|已弃用|弃用|非推奨|폐기된|레거시", re.I)
HIST_FILE = re.compile(r"CHANGELOG|release-notes|/releases?/|changelog", re.I)

def classify(tok, line, m, path, in_history):
    s, e = m.start(), m.end()
    txt = m.group(0)
    before = line[max(0, s - 12):s]
    after = line[e:e + 12]
    # agent names / identifiers
    if re.search(r"manager[-_]$", before, re.I):
        return 'ident:agent-name'
    if txt.isupper() and len(txt) > 2:
        return 'ident:sentinel/env'
    if before.endswith('_') or after.startswith('_'):
        return 'ident:snake'
    if re.match(r"-(?:\d|N\b|<n>|<N>|\{n\}|\$|n\b|\*)", after):
        return 'ident:notation'
    if re.search(r"(?:-f|--name|--role|role[=:]\s*['\"]?|--lane)\s*['\"`]?$", before):
        return 'ident:cli-token'
    if re.match(r"-[a-z]", after) or re.search(r"[a-z]-$", before):
        return 'compound(slug-or-prose)'
    if tok == 'lead' and PLAIN_LEAD.search(line[max(0, s - 2):e + 25]):
        return 'plain-english'
    if tok == 'worker' and PLAIN_WORKER.search(line):
        return 'plain-english'
    if tok == 'lane' and PLAIN_LANE.search(line):
        return 'plain-english'
    if tok == 'companion' and DOC_COMPANION.search(line):
        return 'other-meaning:doc-companion'
    if tok == 'worker' and SUBAGENT_WORKER.search(line):
        return 'role:subagent-worker'
    if tok in ('lead', 'leader') and AGENT_TEAMS.search(line) and not KANBAN_CTX.search(line):
        return 'other-meaning:agent-teams/cg'
    if in_history or HIST_FILE.search(path) or HIST_LINE.search(line):
        return 'frozen/historical'
    return 'role'

rows = []            # path, surface, token, class, count
hard_rows = []       # path, lineno, token, line
heading_rows = []    # path, lineno, heading
per_file_tok = collections.defaultdict(collections.Counter)

for f in files:
    sf = surface(f)
    if not sf:
        continue
    try:
        lines = open(f, encoding='utf-8').read().split('\n')
    except Exception:
        continue
    in_history = False
    cnt = collections.Counter()
    for i, line in enumerate(lines, 1):
        if re.match(r"^#{1,6}\s", line):
            in_history = bool(re.search(r"history|changelog|이력|履歴|历史", line, re.I))
        hit_tokens = set()
        for tok, rx in TOKRE.items():
            for m in rx.finditer(line):
                c = classify(tok, line, m, f, in_history)
                cnt[(tok, c)] += 1
                hit_tokens.add(tok)
        if hit_tokens and '[HARD]' in line:
            hard_rows.append((f, i, ','.join(sorted(hit_tokens)), line.strip()[:220]))
        if hit_tokens and re.match(r"^#{1,6}\s", line):
            heading_rows.append((f, i, ','.join(sorted(hit_tokens)), line.strip()))
    for (tok, c), n in cnt.items():
        rows.append((f, sf, tok, c, n))
        per_file_tok[f][tok] += n

with open(os.path.join(OUT, 'per-file-class.tsv'), 'w') as o:
    o.write('path\tsurface\ttoken\tclass\tcount\n')
    for r in sorted(rows):
        o.write('\t'.join(map(str, r)) + '\n')
with open(os.path.join(OUT, 'hard-lines.tsv'), 'w') as o:
    o.write('path\tline\ttokens\ttext\n')
    for r in hard_rows:
        o.write('\t'.join(map(str, r)) + '\n')
with open(os.path.join(OUT, 'headings.tsv'), 'w') as o:
    o.write('path\tline\ttokens\theading\n')
    for r in heading_rows:
        o.write('\t'.join(map(str, r)) + '\n')

# summaries
tot = collections.Counter(); tot_cls = collections.Counter(); tot_sf = collections.Counter()
for f, sf, tok, c, n in rows:
    tot[tok] += n; tot_cls[(tok, c)] += n; tot_sf[(sf, tok)] += n
print('TOKEN TOTALS', dict(tot))
print('CLASS'); [print(' ', k, v) for k, v in sorted(tot_cls.items())]
print('SURFACE x TOKEN'); [print(' ', k, v) for k, v in sorted(tot_sf.items())]
TOKS = list(TOK)
sfmap = {f: surface(f) for f in per_file_tok}
with open(os.path.join(OUT, 'per-file-pivot.tsv'), 'w') as o:
    o.write('path\tsurface\t' + '\t'.join(TOKS) + '\n')
    for f in sorted(per_file_tok):
        o.write(f + '\t' + sfmap[f] + '\t' + '\t'.join(str(per_file_tok[f][t]) for t in TOKS) + '\n')
fc = collections.Counter()
for f in per_file_tok:
    for t in TOKS:
        if per_file_tok[f][t]:
            fc[(sfmap[f], t)] += 1
print('FILE COUNTS surface x token'); [print(' ', k, v) for k, v in sorted(fc.items())]
# mirror pairs: templates/.claude/X vs .claude/X
T = 'internal/template/templates/'
pairs = []
allf = set(files)
for f in sorted(per_file_tok):
    if f.startswith(T + '.claude/'):
        loc = f[len(T):]
        lt = per_file_tok.get(loc)
        same = None
        if loc in allf:
            same = open(f, 'rb').read() == open(loc, 'rb').read()
        pairs.append((loc, 'present' if loc in allf else 'ABSENT', 'identical' if same else ('differs' if same is False else '-'),
                      ' '.join('%s=%d/%d' % (t, per_file_tok[f][t], (lt or {}).get(t, 0)) for t in TOKS if per_file_tok[f][t] or (lt or {}).get(t, 0))))
for f in sorted(per_file_tok):
    if f.startswith('.claude/') and (T + f) not in allf:
        pairs.append((f, 'LOCAL-ONLY', '-', ' '.join('%s=-/%d' % (t, per_file_tok[f][t]) for t in TOKS if per_file_tok[f][t])))
with open(os.path.join(OUT, 'mirror-pairs.tsv'), 'w') as o:
    o.write('local_path\tlocal_status\tbytes\ttoken=template/local\n')
    for p in pairs:
        o.write('\t'.join(p) + '\n')
print('MIRROR PAIRS', collections.Counter((p[1], p[2]) for p in pairs))
# docs-site per-page parity across locales
pages = collections.defaultdict(dict)
for f in per_file_tok:
    if f.startswith('docs-site/content/'):
        parts = f.split('/')
        pages['/'.join(parts[3:])][parts[2]] = sum(per_file_tok[f][t] for t in ('lead', 'leader', 'lane', 'worker', 'foreman', 'deputy', 'companion'))
with open(os.path.join(OUT, 'docs-locale-parity.tsv'), 'w') as o:
    o.write('page\ten\tko\tja\tzh\n')
    for p in sorted(pages):
        o.write(p + '\t' + '\t'.join(str(pages[p].get(l, 0)) for l in ('en', 'ko', 'ja', 'zh')) + '\n')
print('docs pages with hits', len(pages))
print('files with any hit', len(per_file_tok), 'HARD lines', len(hard_rows), 'headings', len(heading_rows))
