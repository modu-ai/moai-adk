"""t217 D2 measurement: what fraction of real files would the derived
literal pre-filter skip (i.e. NOT spawn sg for)?

Token sets are derived by hand from the SHIPPED ruleset's error-severity rules
(internal/template/templates/.moai/config/astgrep-rules), under the
alternation-extraction reading of spec.md C.2 (audit exit (a)):

  go     go-error-ignored-blank  `$_, $ERR = $FUNC($$$ARGS)` -> mandatory ',' '='
         (7 other go error rules contribute further tokens; ',' and '=' already
         dominate, and adding tokens can only LOWER the skip rate)
  js/ts  child_process.exec | cp.exec | credential prefixes
  py     subprocess.call | os.system | credential prefixes

Credential prefixes come from sec-hardcoded-credential's regex alternation:
  ^["'](sk-|AKIA...|ghp_...|xox[baprs]-|AIza...)
"""
import os
import sys

root = sys.argv[1]
skipdirs = {'.git', 'node_modules', 'vendor', 'public', 'resources', 'dist', 'worktrees'}
tok = {
    'go': [',', '='],
    'js': ['child_process.exec', 'cp.exec', 'sk-', 'AKIA', 'ghp_', 'xox', 'AIza'],
    'py': ['subprocess.call', 'os.system', 'sk-', 'AKIA', 'ghp_', 'xox', 'AIza'],
}
ext = {'.go': 'go', '.js': 'js', '.jsx': 'js', '.mjs': 'js', '.cjs': 'js',
       '.ts': 'js', '.tsx': 'js', '.mts': 'js', '.cts': 'js', '.py': 'py'}
stat = {}
for dp, dn, fn in os.walk(root):
    dn[:] = [d for d in dn if d not in skipdirs]
    for f in fn:
        e = os.path.splitext(f)[1]
        if e not in ext:
            continue
        lang = ext[e]
        try:
            c = open(os.path.join(dp, f), encoding='utf-8', errors='replace').read()
        except Exception:
            continue
        hit = any(t in c for t in tok[lang])
        s = stat.setdefault(lang, [0, 0])
        s[0] += 1
        if not hit:
            s[1] += 1
for k, (n, skip) in sorted(stat.items()):
    print(k, 'files=' + str(n), 'wouldSKIP=' + str(skip), 'rate=%.1f%%' % (100 * skip / n))
