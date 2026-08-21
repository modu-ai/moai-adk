"""M2 clause -> AGENTS.md section trace.

Reads the M1 clause blocks and their C/K/P classification, applies M2's section
assignment, and emits the trace table plus a completeness check (no Codex-relevant
clause unmapped). Run from the worktree root.
"""
import json
import csv

blocks = json.load(open('.moai/reports/t82/blocks.json'))
cls = {r['id']: r['class'] for r in csv.DictReader(open('.moai/reports/t82/classification.tsv'), delimiter='\t')}

sec = {
    'B036': '§1', 'B037': '§1', 'B038': '§1',
    'B076': '§2', 'B077': '§2', 'B013': '§2', 'B012': '§2',
    'B066': '§3', 'B067': '§3', 'B068': '§3', 'B069': '§3', 'B070': '§3', 'B071': '§3',
    'B072': '§4', 'B073': '§4', 'B074': '§4', 'B010': '§4', 'B064': '§4',
    'B028': '§5', 'B029': '§5', 'B030': '§5', 'B031': '§5', 'B032': '§5', 'B033': '§5',
    'B002': '§6', 'B025': '§6', 'B034': '§6', 'B095': '§6', 'B003': '§6', 'B004': '§6', 'B014': '§6',
    'B009': '§7', 'B005': '§7', 'B040': '§7', 'B041': '§7', 'B042': '§7',
}
titles = {
    '§1': '§1 Evidence and verification claims',
    '§2': '§2 Git, branches, and the shared checkout',
    '§3': '§3 Worktrees',
    '§4': '§4 How verification is run',
    '§5': '§5 Core behaviors',
    '§6': '§6 Output, language, and format',
    '§7': '§7 Tools and command output',
}

rows = []
for i, b in enumerate(blocks):
    bid = 'B%03d' % i
    if bid in sec:
        rows.append((bid, cls[bid], b['file'], b['line'], b['bytes'], sec[bid]))

cblocks = ['B%03d' % i for i in range(len(blocks)) if cls['B%03d' % i] == 'C']
missing = [x for x in cblocks if x not in sec]
promoted = sorted(x for x in sec if cls[x] != 'C')

print('mapped rows: %d | C blocks: %d | unmapped C: %s | promoted from K: %s'
      % (len(rows), len(cblocks), missing or 'none', promoted or 'none'))
print('verbatim bytes mapped: %d' % sum(r[4] for r in rows))

order = {'§%d' % k: k for k in range(1, 8)}
rows.sort(key=lambda r: (order[r[5]], r[0]))
with open('.moai/reports/t82/trace.md', 'w') as out:
    out.write('| Clause | Class | Origin file | Line | Verbatim B | `AGENTS.md` section |\n')
    out.write('|---|---|---|---:|---:|---|\n')
    for bid, c, fn, ln, by, s in rows:
        out.write('| `%s` | %s | `%s` | %d | %d | %s |\n' % (bid, c, fn, ln, by, titles[s]))
