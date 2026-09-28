"""Per-file CJK role-token counts for docs-site ko/ja/zh, README.{ko,ja,zh},
and Korean local surfaces. Writes cjk-per-file.tsv; prints totals."""
import sys, os, collections
SP, OUT = sys.argv[1], sys.argv[2]
files = open(os.path.join(SP, 'files.txt')).read().split()
TOK = {
 'ko': ['리드', '리더', '레인', '워커', '동반 세션', '포어맨', '코디네이터', '대리'],
 'ja': ['リード', 'リーダー', 'レーン', 'ワーカー', 'コンパニオン', 'フォアマン', 'コーディネーター', '代理'],
 'zh': ['主导', '领导', '负责人', '主控', '工作者', '泳道', '通道', '伴随', '工头', '协调者'],
}
def lang_of(f):
    for l in ('ko', 'ja', 'zh'):
        if f.startswith('docs-site/content/%s/' % l) or f == 'README.%s.md' % l:
            return l
    if f in ('CLAUDE.local.md', '.claude/rules/local/gitflow-lane-protocol.md'):
        return 'ko'
    return None
tot = collections.Counter()
with open(os.path.join(OUT, 'cjk-per-file.tsv'), 'w') as o:
    o.write('path\tlang\tcounts\n')
    for f in files:
        l = lang_of(f)
        if not l:
            continue
        try:
            t = open(f, encoding='utf-8').read()
        except Exception:
            continue
        c = {k: t.count(k) for k in TOK[l] if t.count(k)}
        if c:
            o.write('%s\t%s\t%s\n' % (f, l, ' '.join('%s=%d' % kv for kv in c.items())))
            for k, v in c.items():
                tot[(l, 'README' if f.startswith('README') else ('local' if not f.startswith('docs') else 'docs'), k)] += v
for k, v in sorted(tot.items()):
    print(k, v)
# CLAUDE.local.md §4.1 slice
t = open('CLAUDE.local.md', encoding='utf-8').read()
s = t.find('### §4.1'); e = t.find('\n## 5', s)
sl = t[s:e]
print('CLAUDE.local.md §4.1 lines', sl.count('\n'), {k: sl.count(k) for k in ('리드', '레인', '워커', '[HARD]')})
