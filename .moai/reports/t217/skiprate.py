"""t217 C.3 measurement: what fraction of real files would the derived literal
pre-filter skip (i.e. NOT dispatch a scan for)?

Token sets are derived by hand from the SHIPPED ruleset's error-severity rules
(internal/template/templates/.moai/config/astgrep-rules), applying every row of
spec.md C.2. The shipped ruleset was verified byte-identical to the primary
checkout's copy for injection.yml, the file the audit's E1 finding concerned.

SOUNDNESS NOTE. A token set is a UNION, and the pre-filter skips only when NONE
of its tokens is present. Adding a token can therefore only LOWER a reported
skip rate, never raise it. An incomplete token set inflates the saving, which is
why every rule below is enumerated rather than sampled. This is the defect the
plan audit found in v1 of this script: the python `any:` has FOUR branches and
only two were listed, inflating python's rate from 85.7% to 92.9%.

go -- 8 error rules
  go-error-ignored-blank         $_, $ERR = $FUNC($$$ARGS)         -> ',' '='
  go-defer-in-loop               for .. { .. defer $R.$M() .. }    -> 'for' 'defer'
  sec-hardcoded-api-key          const $NAME = "sk-$$$REST"        -> 'const'
  sec-hardcoded-jwt-signing-key  SignedString([]byte("$H"))        -> 'SignedString'
  sec-command-injection-shell    exec.Command("sh", "-c", $CMD)    -> 'exec.Command'
  sec-template-injection-html    template.HTML($USER_INPUT)        -> 'template.HTML'
  sec-weak-hash-md5              md5.New()                         -> 'md5.New'
  sec-hardcoded-credential       kind: + regex: alternation        -> credential prefixes
  (',' and '=' from the first rule dominate -- they appear in nearly every Go
   file -- so the go rate is bounded above by that one rule.)

javascript / typescript -- 2 error rules
  sec-command-injection-exec     any: child_process.exec | cp.exec -> BOTH branches
  sec-hardcoded-credential       kind: + regex: alternation        -> credential prefixes

python -- 2 error rules
  sec-command-injection-shell    any: subprocess.call | subprocess.run
                                    | subprocess.Popen | os.system -> ALL FOUR branches
  sec-hardcoded-credential       kind: + regex: alternation        -> credential prefixes

Credential prefixes are the mandatory literal prefix of each branch of
  ^["'](sk-|AKIA[0-9A-Z]{16}|ghp_[0-9A-Za-z]{36}|xox[baprs]-|AIza[0-9A-Za-z_-]{35})

Only whitespace-free literals are used as tokens. ast-grep normalizes whitespace
when matching, so a run such as " = " is not guaranteed to appear verbatim in the
source; '=' is.
"""
import os
import sys

CRED = ['sk-', 'AKIA', 'ghp_', 'xox', 'AIza']

tok = {
    'go': [',', '=', 'for', 'defer', 'const', 'SignedString', 'exec.Command',
           'template.HTML', 'md5.New'] + CRED,
    'js': ['child_process.exec', 'cp.exec'] + CRED,
    'py': ['subprocess.call', 'subprocess.run', 'subprocess.Popen', 'os.system'] + CRED,
}
ext = {'.go': 'go', '.js': 'js', '.jsx': 'js', '.mjs': 'js', '.cjs': 'js',
       '.ts': 'js', '.tsx': 'js', '.mts': 'js', '.cts': 'js', '.py': 'py'}

root = sys.argv[1]
skipdirs = {'.git', 'node_modules', 'vendor', 'public', 'resources', 'dist', 'worktrees'}
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
