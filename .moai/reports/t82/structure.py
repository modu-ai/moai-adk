"""Measure the document-structure component of AGENTS.md.

M1 assumed 3,300 B of structure (headings + framing prose + front matter) on top of
the compressed clause text; the assumption becomes measurable once the document has a
skeleton. Structure here = the front-matter block (everything above the first clause
section) + every heading line + every horizontal rule + every blank line. The
remainder is clause text. Run from the worktree root.
"""

raw = open('AGENTS.md', 'rb').read()
lines = raw.split(b'\n')

first_section = next(i for i, ln in enumerate(lines) if ln.startswith(b'## 1.'))
front = sum(len(ln) + 1 for ln in lines[:first_section])

structure = front
for ln in lines[first_section:]:
    if ln.startswith(b'#') or ln.strip() == b'---' or ln.strip() == b'':
        structure += len(ln) + 1

total = len(raw)
print('AGENTS.md total bytes: %d' % total)
print('front matter (above first clause section): %d' % front)
print('structure total (front matter + headings + rules + blank lines): %d' % structure)
print('clause text: %d' % (total - structure))
print('M1 assumption for structure: 3300  -> delta %+d' % (structure - 3300))
