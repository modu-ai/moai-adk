"""For every token-bearing heading, count cross-file references to it.

A reference is either `§ <heading text>` / `§<heading text>` (prose anchor
convention used across .claude/rules) or a markdown slug link `#<slug>`.
Scans git-tracked .md/.tmpl/.go/.toml/.yaml files (Go included: tests pin
anchors). Read-only.
"""
import re, sys, os, collections
SP, OUT = sys.argv[1], sys.argv[2]
files = [f for f in open(os.path.join(SP, 'files.txt')).read().split()
         if f.endswith(('.md', '.tmpl', '.go', '.toml', '.yaml', '.yml'))]
texts = {}
for f in files:
    try:
        texts[f] = open(f, encoding='utf-8').read()
    except Exception:
        pass
rows = []
for line in open(os.path.join(OUT, 'headings.tsv')).read().split('\n')[1:]:
    if not line.strip():
        continue
    path, ln, toks, h = line.split('\t', 3)
    title = re.sub(r'^#+\s*', '', h).strip()
    title_core = re.sub(r'\s*\{#.*\}$', '', title)
    slug = re.sub(r'[^\w\- ]', '', title_core.lower()).strip().replace(' ', '-')
    sect = re.compile(r'§\s*' + re.escape(title_core))
    slug_rx = re.compile(r'#' + re.escape(slug) + r'\b') if slug else None
    refs = []
    for f, t in texts.items():
        n = len(sect.findall(t)) + (len(slug_rx.findall(t)) if slug_rx and len(slug) > 3 else 0)
        if f == path:
            continue
        if n:
            refs.append('%s:%d' % (f, n))
    rows.append((path, ln, toks, title_core, len(refs), ' '.join(sorted(refs))[:1500]))
with open(os.path.join(OUT, 'anchors.tsv'), 'w') as o:
    o.write('path\tline\ttokens\theading\tref_files\trefs\n')
    for r in sorted(rows, key=lambda r: -r[4]):
        o.write('\t'.join(map(str, r)) + '\n')
c = sum(1 for r in rows if r[4])
print('headings', len(rows), 'referenced-elsewhere', c, 'total ref-files', sum(r[4] for r in rows))
