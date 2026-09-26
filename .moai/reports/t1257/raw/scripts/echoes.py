"""Q3 clause-echo search (closed by measurement). Read-only.

Usage (from the tree root):
    python3 .moai/reports/t1257/raw/scripts/echoes.py > .moai/reports/t1257/raw/q3-clause-echoes.tsv

Searches every git-tracked-looking text file under the in-scope roots for
statements that echo the promotion / production / dispatch-only invariants
of kanban-dispatch.md. Python `re` is used instead of BSD grep because BSD
grep miscounts multibyte alternations (inventory §3). Output columns:
path, line, locale, pattern-id, text (truncated to 220 chars).
"""
import os, re, sys

ROOTS = ['.claude', 'internal/template/templates', 'docs-site/content',
         'README.md', 'README.ko.md', 'README.ja.md', 'README.zh.md',
         'CLAUDE.md', 'AGENTS.md', 'CLAUDE.local.md', 'internal/cli/todo.go']
EXT = ('.md', '.tmpl', '.toml', '.yaml', '.yml', '.go')

PATTERNS = {
    'en-operator-act': r"operator.s (own )?act",
    'en-sole-producer': r"sole producer",
    'en-operator-picks': r"operator picks|picking the next card|operator-picked|picks the next",
    'en-never-picks': r"never picks for|silently promotes?|looks ready",
    'en-marked-picked': r"marked `?picked`?",
    'en-routes-whole': r"routes each card",
    'ko': r"스스로 고르|골라둔|고른 카드|운영자의 (행위|몫|일)|유일한 생산자|카드를 집지|다음 카드를 dispatch|운영자가 고른|운영자가 이미",
    'ja': r"オペレーターの行為|唯一の生産者|選んだカード|選んでおいた|自分で選|カードを選ぶこと|次のカードを選ぶ|(運営者|オペレーター)が(すでに|カードを)?選|選ぶ主体|ディスパッチ周期を担う",
    'zh': r"操作员的行为|唯一的生产者|操作(员|者)(已经|已)?(选|挑|标)|自行(选|挑)|整张.{0,6}(路由|分派)|选下一张|挑卡片的主体|掌管派工循环",
    'any-picked': r"`picked`",
}
RX = {k: re.compile(v) for k, v in PATTERNS.items()}


def locale(p):
    for l in ('ko', 'ja', 'zh', 'en'):
        if p.startswith('docs-site/content/%s/' % l) or p == 'README.%s.md' % l:
            return l
    if p == 'README.md':
        return 'en'
    if p == 'CLAUDE.local.md' or '/rules/local/' in p:
        return 'ko-local'
    return 'en-doc'


def walk():
    for r in ROOTS:
        if os.path.isfile(r):
            yield r
            continue
        for d, _, fs in os.walk(r):
            if '/agent-memory' in d or '/worktrees' in d:
                continue
            for f in fs:
                if f.endswith(EXT):
                    yield os.path.join(d, f)


print('path\tline\tlocale\tpattern\ttext')
n = 0
files = set()
for p in sorted(walk()):
    try:
        lines = open(p, encoding='utf-8').read().split('\n')
    except Exception:
        continue
    for i, line in enumerate(lines, 1):
        for k, rx in RX.items():
            if rx.search(line):
                print('%s\t%d\t%s\t%s\t%s' % (p, i, locale(p), k, line.strip()[:220].replace('\t', ' ')))
                n += 1
                files.add(p)
                break
print('# total lines %d, files %d' % (n, len(files)), file=sys.stderr)
