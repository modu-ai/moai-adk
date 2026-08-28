import re, sys

ID = re.compile(r'AC-(?:[A-Z0-9]+-)*[0-9]+')
TOK_SAMELINE = re.compile(r'\A[ \t]*(\[RETIRED\]|\[REF\])')
TOK_WS = re.compile(r'\A\s*(\[RETIRED\]|\[REF\])')


def adjacency(path, mode):
    txt = open(path, encoding='utf-8').read()
    pat = TOK_SAMELINE if mode == 'sameline' else TOK_WS
    st = {}
    for m in ID.finditer(txt):
        st.setdefault(m.group(0), []).append(bool(pat.match(txt[m.end():])))
    live = [k for k, v in st.items() if not any(v)]
    exc = [k for k, v in st.items() if all(v)]
    amb = [k for k, v in st.items() if any(v) and not all(v)]
    return live, exc, amb


def linesem(path):
    txt = open(path, encoding='utf-8').read()
    live = set()
    for line in txt.splitlines():
        marked = ('[REF]' in line) or ('[RETIRED]' in line)
        for m in ID.finditer(line):
            if not marked:
                live.add(m.group(0))
    return live


if __name__ == '__main__':
    path = sys.argv[1]
    which = sys.argv[2] if len(sys.argv) > 2 else 'sameline'
    if which == 'line':
        print('LINE live=%d' % len(linesem(path)))
    else:
        l, e, a = adjacency(path, which)
        if a:
            print('AMBIGUOUS %s (live=%d excluded=%d)' % (' '.join(sorted(a)), len(l), len(e)))
        else:
            print('COUNT %d (live=%d excluded=%d ambiguous=0)' % (len(l), len(l), len(e)))
