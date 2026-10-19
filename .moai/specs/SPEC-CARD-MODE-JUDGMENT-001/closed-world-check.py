"""Closed-world checks for SPEC-CARD-MODE-JUDGMENT-001 (AC-MJ-020 and the commit-shape decision of AC-MJ-011).

Read-only. Run from the worktree root as one invocation:

  python3 -I <this file> lines <manifest> [<root>]   compare added/removed lines with the manifest
  python3 -I <this file> paths <manifest> [<base>]   changed or new paths must be manifest-listed or run-writable
  python3 -I <this file> paths <manifest> -- <path>...  judge an explicit path list instead of querying git
  python3 -I <this file> pairs <manifest> <commit>[:<n>]...  print `<commit> pairs <both sides> unpaired <one side only>`;
                                                             exit 1 if any unpaired, or if `:<n>` is given and the pair count differs

Semantics of `lines`: for each file of the manifest, on the template side and on the deployed side,
the non-blank lines of the working copy are compared with the non-blank lines of the same path at the
manifest's base commit, as multisets. A line present more often in the working copy than in the base is
an ADDED line; present less often is a REMOVED line. Five numbers are printed per side:
  <added but not listed> <listed but not added> <removed but not listed> <listed but not removed>
  <code-fence lines whose preceding non-blank line differs from the base>
and the command exits 1 unless every number is 0. Blank lines are ignored; the order of non-fence
lines is ignored; every other difference, with or without any particular word in it, is a number.
The fifth number reads a code-fence line's (one beginning with three backticks) nearest preceding
non-blank line, walking past the lines the manifest lists, and compares that pair with the base: a
fence moved behind different text keeps the line multiset the first four numbers compare, and is seen
only here. Listed lines added or removed anywhere leave every fence's walked predecessor unchanged,
so correct work prints 0 on the fifth number whatever position the listed rows land in.

Semantics of `paths`: every changed or new path of the working copy — against the base commit, or the
explicit list given after `--` — must be named by the manifest or lie under one of the four paths a
run of this SPEC legitimately writes (.moai/specs/SPEC-CARD-MODE-JUDGMENT-001/, .moai/reports/,
.moai/state/, .moai/logs/). Everything else, tracked .moai projections included, is printed as an
unexpected path and makes the exit 1.
"""
import subprocess
import sys
from collections import Counter


def git(*args):
    r = subprocess.run(['git'] + list(args), capture_output=True)
    if r.returncode != 0:
        sys.stderr.write('git %s failed (%d): %s\n' % (' '.join(args[:2]), r.returncode, r.stderr.decode('utf-8', 'replace')[:200]))
        raise SystemExit(2)
    return r.stdout


def nonblank(text):
    return [line.rstrip('\r') for line in text.split('\n') if line.strip()]


def load_manifest(path):
    base = strip = None
    files, also = [], []
    cur = None
    for raw in open(path, encoding='utf-8').read().split('\n'):
        if raw == '' or raw.startswith('#'):
            continue
        if raw.startswith('base: '):
            base = raw[6:].strip()
        elif raw.startswith('twin-strip: '):
            strip = raw[12:].strip()
        elif raw.startswith('also: '):
            also.append(raw[6:].strip())
        elif raw.startswith('file: '):
            letter, p = raw[6:].split(' ', 1)
            cur = {'k': letter, 'path': p.strip(), 'plus': Counter(), 'minus': Counter()}
            files.append(cur)
        elif raw[:2] == '+ ':
            cur['plus'][raw[2:]] += 1
        elif raw[:2] == '- ':
            cur['minus'][raw[2:]] += 1
        else:
            raise SystemExit('manifest: unreadable line: ' + raw[:60])
    if not (base and strip and files):
        raise SystemExit('manifest: base, twin-strip and at least one file are required')
    return base, strip, files, also


def fence_pairs(lines, listed):
    c = Counter()
    prev = ''
    for line in lines:
        if line.startswith('```'):
            c[(line, prev)] += 1
        if line not in listed:
            prev = line
    return c


def cmd_lines(manifest, root):
    base, strip, files, _ = load_manifest(manifest)
    bad = 0
    for f in files:
        parts = [f['k']]
        listed = f['plus'] | f['minus']
        for label, p in (('template', f['path']), ('deployed', f['path'][len(strip):])):
            b_lines = nonblank(git('show', base + ':' + p).decode('utf-8'))
            b = Counter(b_lines)
            try:
                w_lines = nonblank(open(root + p, 'rb').read().decode('utf-8'))
            except OSError:
                parts.append('%s missing-file' % label)
                bad = 1
                continue
            w = Counter(w_lines)
            added, removed = w - b, b - w
            n = [sum((added - f['plus']).values()), sum((f['plus'] - added).values()),
                 sum((removed - f['minus']).values()), sum((f['minus'] - removed).values()),
                 sum((fence_pairs(w_lines, listed) - fence_pairs(b_lines, listed)).values())]
            if any(n):
                bad = 1
            parts.append('%s %d %d %d %d %d' % (label, *n))
        print(' '.join(parts))
    return bad


def cmd_paths(manifest, base_override, simulated=()):
    base, strip, files, also = load_manifest(manifest)
    if base_override:
        base = base_override
    allowed = set(also)
    for f in files:
        allowed.add(f['path'])
        allowed.add(f['path'][len(strip):])
    if simulated:
        changed = set(simulated)
    else:
        changed = set(git('diff', '--name-only', '-z', base, '--').decode('utf-8').split('\0'))
        changed |= set(git('ls-files', '--others', '--exclude-standard', '-z').decode('utf-8').split('\0'))
    writable = ('.moai/specs/SPEC-CARD-MODE-JUDGMENT-001/', '.moai/reports/', '.moai/state/', '.moai/logs/')
    unexpected = sorted(p for p in changed if p and p not in allowed and not p.startswith(writable))
    print('unexpected-paths %d' % len(unexpected))
    for p in unexpected[:8]:
        print(p)
    return 1 if unexpected else 0


def cmd_pairs(manifest, commits):
    _, strip, files, _ = load_manifest(manifest)
    pairs = [(f['path'], f['path'][len(strip):]) for f in files]
    bad = 0
    for arg in commits:
        c, _, want = arg.partition(':')
        names = set(git('diff-tree', '--no-commit-id', '--name-only', '-r', '-z', c).decode('utf-8').split('\0'))
        lone = [t if t in names else d for t, d in pairs if (t in names) != (d in names)]
        both = sum(1 for t, d in pairs if t in names and d in names)
        tail = ' expected %s' % want if want else ''
        print('%s pairs %d unpaired %d%s' % (c[:10], both, len(lone), tail))
        for p in lone:
            print('  ' + p)
        if lone or (want and both != int(want)):
            bad = 1
    return bad


def main(argv):
    if len(argv) < 3:
        raise SystemExit(__doc__)
    mode, manifest = argv[1], argv[2]
    if mode == 'lines':
        root = argv[3] if len(argv) > 3 else ''
        if root and not root.endswith('/'):
            root += '/'
        return cmd_lines(manifest, root)
    if mode == 'paths':
        rest = argv[3:]
        if '--' in rest:
            cut = rest.index('--')
            return cmd_paths(manifest, rest[0] if cut else '', rest[cut + 1:])
        return cmd_paths(manifest, rest[0] if rest else '')
    if mode == 'pairs':
        return cmd_pairs(manifest, argv[3:])
    raise SystemExit('unknown mode: ' + mode)


if __name__ == '__main__':
    sys.exit(main(sys.argv))
