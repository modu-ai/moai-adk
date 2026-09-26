"""Q3 clause-echo search (closed by measurement). Read-only.

Usage (from the tree root):
    python3 .moai/reports/t1257/raw/scripts/echoes.py > .moai/reports/t1257/raw/q3-clause-echoes.tsv

Pass 1 searches every text file under the in-scope roots for statements that
echo the promotion / production / dispatch-only invariants of
kanban-dispatch.md. Every concept is expressed in all four locales (en, ko,
ja, zh) so no locale can be caught while its siblings are missed.

Pass 2 (cross-locale mirror): docs-site pages are line-aligned across the four
locales. For every docs-site hit at <locale>/<page>:<line>, the same line of
the other three locale copies is listed with pattern id `mirror`, so a
locale-asymmetric pattern cannot leave a sibling line unlisted.

Python `re` is used instead of BSD grep because BSD grep miscounts multibyte
alternations (inventory §3). Output columns: path, line, locale, pattern-id,
text (truncated to 220 chars). Totals go to stderr.
"""
import os
import re
import sys

ROOTS = ['.claude', 'internal/template/templates', 'docs-site/content',
         'README.md', 'README.ko.md', 'README.ja.md', 'README.zh.md',
         'CLAUDE.md', 'AGENTS.md', 'CLAUDE.local.md', 'internal/cli/todo.go']
EXT = ('.md', '.tmpl', '.toml', '.yaml', '.yml', '.go')
LOCALES = ('en', 'ko', 'ja', 'zh')

# One entry per concept; each concept carries all four locales.
CONCEPTS = {
    'operator-act': [
        r"operator.s (own )?act",
        r"운영자의 (행위|몫|일)",
        r"(オペレーター|運営者)の行為",
        r"操作(员|者)的行为",
    ],
    'sole-producer': [
        r"sole producer",
        r"유일한 생산자",
        r"唯一の生産者",
        r"唯一的生产者",
    ],
    'who-picks': [
        r"operator picks|picking the next card|operator-picked|picks the next|"
        r"human does the picking|the one who picks|who picks|does the picking",
        r"고르는 주체|운영자가 고른|운영자가 (이미 )?골라|골라둔|고른 카드|다음 카드를 고르",
        r"選ぶ主体|(運営者|オペレーター)が(すでに|カードを)?選|選んだカード|選んでおいた|次のカードを選ぶ",
        r"挑选的主体|挑卡片的主体|操作(员|者)(已经|已)?(选|挑|标)|选下一张",
    ],
    'never-picks': [
        r"never picks for|silently promotes?|looks ready|lane never picks|does not pick",
        r"스스로 고르지|카드를 집지|스스로 승격",
        r"自分で選(ば|ん)|勝手に昇格",
        r"自行(选|挑)|擅自(提升|晋升)",
    ],
    'routes-whole': [
        r"routes each card|owns the dispatch cycle|deals the cards",
        r"통째로 (배정|라우팅)|배차 주기를 (맡|쥔)|다음 카드를 dispatch",
        r"ディスパッチ周期を担う|丸ごと(割り当て|ルーティング)",
        r"掌管派工循环|整张.{0,6}(路由|分派)",
    ],
    'picked-status': [r"`picked`"],
}
RX = []
for cid, pats in CONCEPTS.items():
    for pat in pats:
        RX.append((cid, re.compile(pat)))


def locale_of(path):
    for loc in LOCALES:
        if path.startswith('docs-site/content/%s/' % loc) or path == 'README.%s.md' % loc:
            return loc
    if path == 'README.md':
        return 'en'
    if path == 'CLAUDE.local.md' or '/rules/local/' in path:
        return 'ko-local'
    return 'en-doc'


def walk():
    for root in ROOTS:
        if os.path.isfile(root):
            yield root
            continue
        for d, _, fs in os.walk(root):
            if '/agent-memory' in d or '/worktrees' in d:
                continue
            for f in fs:
                if f.endswith(EXT):
                    yield os.path.join(d, f)


def read_lines(path, cache={}):
    if path not in cache:
        try:
            cache[path] = open(path, encoding='utf-8').read().split('\n')
        except Exception:
            cache[path] = None
    return cache[path]


hits = {}  # (path, line) -> (locale, pattern-id, text)
for p in sorted(walk()):
    lines = read_lines(p)
    if lines is None:
        continue
    for i, line in enumerate(lines, 1):
        for cid, rx in RX:
            if rx.search(line):
                hits[(p, i)] = (locale_of(p), cid, line.strip()[:220].replace('\t', ' '))
                break

# Pass 2: cross-locale mirror for docs-site pages.
mirrors = 0
for (p, i) in list(hits):
    if not p.startswith('docs-site/content/'):
        continue
    page = p.split('/', 3)[3]
    for loc in LOCALES:
        q = 'docs-site/content/%s/%s' % (loc, page)
        if (q, i) in hits:
            continue
        lines = read_lines(q)
        if lines is None or i > len(lines):
            continue
        hits[(q, i)] = (loc, 'mirror', lines[i - 1].strip()[:220].replace('\t', ' '))
        mirrors += 1

print('path\tline\tlocale\tpattern\ttext')
for (p, i) in sorted(hits):
    loc, cid, text = hits[(p, i)]
    print('%s\t%d\t%s\t%s\t%s' % (p, i, loc, cid, text))
files = {p for (p, _) in hits}
print('# total lines %d, files %d, mirror-added %d' % (len(hits), len(files), mirrors), file=sys.stderr)
