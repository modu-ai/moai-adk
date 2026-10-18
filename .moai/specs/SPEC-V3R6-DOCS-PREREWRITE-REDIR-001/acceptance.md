# Acceptance Criteria — SPEC-V3R6-DOCS-PREREWRITE-REDIR-001 (revision 0.1.0)

## A. Pin, conventions, and the instrument

- **Document-level pin.** Tree commit `0ddf5ca48bed7aaeeb68ed4fc686bcfa098a82b1`, tree object `a2c0d16b90c0e49371b6efec3084e6e270781bd7`, `docs-site/vercel.json` blob `959e3de56575cf12306e1d787b9c9c9f2265e219` (E-R1). Every cell that reads a committed object names the commit in its command. The instrument cells (E-R4, E-R5, E-R6, E-R7) read the working tree; E-R1 shows that `docs-site` in the working tree equals the pinned commit (`git diff --stat 0ddf5ca48b -- docs-site` prints nothing, `git status --short` prints nothing), so those cells are attributed to the same tree. Measured on 2026-10-18 between 16:48Z and 17:01Z.
- **BASE.** In the run phase, `<BASE>` means the commit the run phase started from, recorded in progress.md §E.2 at M0. The RED-now cells below use the literal pin `0ddf5ca48b`; on a rebase the run phase re-measures and re-pins (verification-completeness §4).
- **Cell form (verification-completeness §2.1).** A RED-now cell is one read-only invocation with no `;`, `&&`, `|`, subshell, or redirection. Its verbatim stdout and exit code are separate fields. Where the Bash tool does not display a zero exit status, the status was taken from a measurement wrapper outside the cell (`runctl.py`, which runs the same command and prints the exit code); the ledger says so.
- **The instrument (E-R0).** The redirect-check script reads `<root>/vercel.json` and `<root>/content/<locale>/**.md`, derives the page URL set from the content tree, and runs one of two modes. `shadow` (REQ-PRC-011) lists every redirect source that matches an existing page or section URL. `resolve` (REQ-PRC-010) lists every redirect destination that is external or resolves to no page. It prints `swept: redirects=<R> pages=<P> locales=<...>` first and exits 2 (fail closed) when R or P is zero, when `vercel.json` or `content/` cannot be read, or when a redirect source uses a pattern token outside the supported subset; it exits 1 on findings and 0 otherwise. The script text is the fenced block in E-R0; the run phase extracts it verbatim to `<scratch>/redirect-check.py` and confirms its SHA-256 before use. `<scratch>` is a session-scoped directory under `/private/tmp`, never inside the repository tree. In the plan phase it was `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/7b2f1a5a-e1bf-4787-be19-8084e37cb916/scratchpad`.
- **External cells** (E-R8) are labelled `external`, are not tree-attributed, and are never recorded as a pass (verification-completeness §2.1, undecidable disposition).
- **Classes.** `release-blocking (gated)`: re-executable on the tree; counts toward the verdict once its gate rows carry operator verdicts. `regression-guard`: a guard against a change this work could introduce; it cannot be red on the pre-change tree and is not counted toward the verdict.
- **Reading note on exit codes.** Findings exit 1. A zero-finding run with a non-empty sweep exits 0. An empty sweep or an unreadable input exits 2, so an empty result is never read as a pass.

## A.1 Traceability matrix (AC to REQ)

| AC | REQ | Gate (decision rows) | Class | Ledger entries |
|----|-----|----------------------|-------|----------------|
| AC-PRC-011 | REQ-PRC-011 | Q23 (first); Q16; coupled to Q21 | release-blocking (gated) | E-R5, E-R3 (RED-now); E-R6, E-R7 (instrument controls); E-R8, E-R9 (external and reference observations) |
| AC-PRC-010 | REQ-PRC-010 | Q15, Q21 | release-blocking (gated) | E-R4, E-R2 (RED-now); E-R6, E-R7 (controls) |
| AC-REDIR-001 | REQ-REDIR-002 | none | regression-guard | E-R2, E-R4 (count baseline) |
| AC-REDIR-002 | REQ-REDIR-001 | Q21 | regression-guard | E-R9 |
| AC-REDIR-003 | REQ-REDIR-003 | Q17 | regression-guard | E-R10, E-R10b, E-R13 |

The umbrella's AC-PRC-010 and AC-PRC-011 are imported by the umbrella (M6) by criterion ID and this child's commit SHA; AC-REDIR-001 to AC-REDIR-003 are local to this child and are not imported.

## B. Acceptance criteria

### AC-PRC-011 — no redirect shadows an existing page, including the moai-mx pair (REQ-PRC-011)

Recorded first, per the card's before-start condition (Q23; leader ruling B04, `d-20261014T060405Z-0fcd`).

- **Given** the tree at BASE, where `docs-site/vercel.json` holds a locale-aware source for `changelog` (line 113) and one for `utility-commands/moai-mx` (line 163), and the four `changelog/_index.md` and four `utility-commands/moai-mx.md` pages exist (E-R2, E-R3).
- **When** the instrument runs in `shadow` mode on `docs-site` after the change, and the page-presence commands of D2 to D4 run.
- **Then** the end state in § D (branch table) holds for the recorded Q23, Q16, and Q21 answers.
- **RED-now (E-R5).** `<scratch>/redirect-check.py shadow docs-site` exits 1 and prints `shadowed: 8`: four lines for source `/:locale(ko|en|ja|zh)/changelog` at line 113 and four for source `/:locale(ko|en|ja|zh)/utility-commands/moai-mx` at line 163. It is red for the stated reason that those two locale-aware sources each match four existing pages. It is not red because of the root-level sources at lines 118 and 168: those match no page and print no line.
- **Green path.** The moai-mx lines flip at M1 (the recorded Q23 answer is applied to `vercel.json`). The four changelog lines flip at M2 under the Q16 Default (the pages are deleted) or at M1 under the Q16 Alternate (the rule at line 113 is removed). Under Q23 option (b) the four moai-mx lines are the expected end state and the amended REQ-PRC-011 names them.
- **Done when** all of the following hold at the closing commit (each is one invocation; commands run from the repository root):
  - D1. `python3 -I <scratch>/redirect-check.py shadow docs-site` prints a first line `swept: redirects=<N> pages=<P> ...` with N and P from § D, then no finding lines (or, under Q23 option (b), exactly the four moai-mx lines), and exits 0 (or 1 under option (b)).
  - D2. `git ls-tree -r --name-only HEAD -- docs-site/content/ko/utility-commands/moai-mx.md docs-site/content/en/utility-commands/moai-mx.md docs-site/content/ja/utility-commands/moai-mx.md docs-site/content/zh/utility-commands/moai-mx.md` prints four lines, under every Q23 option. Its pre-change output is E-R3 (positive control).
  - D3. `git grep -n -E '^draft:' HEAD -- <the same four paths>` prints four lines, each `draft: false`. Its pre-change output is E-R3b.
  - D4. Under the Q16 Default: `git ls-tree -r --name-only HEAD -- docs-site/content/ko/changelog docs-site/content/en/changelog docs-site/content/ja/changelog docs-site/content/zh/changelog` prints nothing, and its pre-change output is four lines (E-R3, positive control for the same pathspec). Under the Q16 Alternate it prints those four lines.
  - D5. The instrument controls of E-R6 and E-R7, re-run with the extracted script, print the recorded outputs (the instrument is verified before its verdict is read).
- **Mutant probe.** (1) Delete the four moai-mx pages instead of the rules: D1 reports fewer findings for them, and D2 prints nothing, so the mutant fails D2 (E-R7 `mutant-mxpages` shows D1 alone passes the moai-mx lines). (2) Mark the four pages `draft: true`: the instrument excludes drafts and D1 passes; D3 prints `draft: true` and fails. (3) Rename the page files: D2 fails. (4) Keep the rule and add a trailing slash to its source: the instrument matches the source with and without a trailing slash, so D1 still prints the four lines. (5) Replace the source with a token the instrument does not support: it exits 2.
- **What this criterion does not detect** (also listed in § G): a source that differs from a page URL only by letter case (the instrument compares case-sensitively; whether Vercel does is not measured); a redirect rule carrying `has` or `missing` conditions (the instrument reads only `source`); a page whose URL is changed by `url:`, `slug:`, or `aliases:` front matter (none exists at BASE, E-R3d, but the instrument ignores them); a shadowed static file under `docs-site/static`; and any Hugo-generated URL that is not a content page. D2 to D4 detect a deleted, drafted, or renamed page only for the eight paths named.

### AC-PRC-010 — every redirect destination resolves (REQ-PRC-010)

- **Given** the tree at BASE, where three `agency` rules point at a `design` section that exists in no locale and two changelog rules point at an external URL (E-R2, E-R3c).
- **When** the instrument runs in `resolve` mode on `docs-site` after the change, and the source and destination commands of D2 and D3 run.
- **Then** the end state in § D holds for the recorded Q15 and Q21 answers.
- **RED-now (E-R4).** `<scratch>/redirect-check.py resolve docs-site` exits 1 and prints `unresolved: 14`: nine lines for the three `agency` rules (4 + 4 + 1, one line per locale for the two locale-aware rules) and five lines for the two changelog rules (4 + 1), each marked `external destination`. It is red for the stated reason: the destinations `/<locale>/design[/...]` and `/design/...` resolve to no page, and the two GitHub URLs are external. The other 32 rules print no line.
- **Green path.** M1 removes (or, under the Q15 Alternate, re-points) the `agency` rules and applies the recorded Q21 answer. Under Q21 option (b) the five external lines are the expected end state and the amended REQ-PRC-010 admits exactly the GitHub Releases string.
- **Done when** all of the following hold at the closing commit:
  - D1. `python3 -I <scratch>/redirect-check.py resolve docs-site` prints `swept: redirects=<N> pages=<P> ...` with N and P from § D, then no finding lines, and exits 0 (under Q21 option (b): exactly the five external lines naming the admitted destination, and `unresolved: 5`, exit 1).
  - D2. Under the Q15 Default: `git grep -n -E '"source": "[^"]*agency' HEAD -- docs-site/vercel.json` prints nothing (exit 1); its pre-change output is three lines (E-R2, positive control).
  - D3. Under Q21 option (a): `git grep -n -E '"destination": "http' HEAD -- docs-site/vercel.json` prints nothing (exit 1); its pre-change output is two lines (E-R2b, positive control).
  - D4. The controls of E-R6 and E-R7 print the recorded outputs.
- **Mutant probe.** (1) Delete every rule: the instrument exits 2 (empty sweep, E-R6 `empty`). (2) Delete one rule outside the verdict's set: N in D1 is lower than § D predicts, and AC-REDIR-001 D2 fails. (3) Re-point a changelog rule to an existing page instead of removing it or admitting it: a re-point is not one of the Q21 options. Under Q21 (a), D1 and D3 pass, and the mutant is caught by AC-REDIR-001 D2 (N is 2 higher than § D predicts). Under Q21 (b), D1 fails (fewer than five external lines). (4) Add a new rule with a resolving destination: N is higher than § D predicts, and AC-REDIR-001 D2 fails.
- **What this criterion does not detect:** a destination that resolves to an existing page that is the wrong page (the criterion checks existence, as REQ-PRC-010 states); a destination with a `:path*` tail whose base resolves while the tail does not (the instrument strips the tail and checks the base only); a destination with a query string or fragment (the instrument reports it as unresolved, a false positive that is visible); and the deployed behavior (E-R8 is external).

### AC-REDIR-001 — only the named files and rules change (REQ-REDIR-002) — regression-guard

- **Given** the tree at BASE.
- **When** the run phase has committed its changes.
- **Then** nothing outside the allowed set changed, and no redirect rule was added or removed beyond the recorded answers.
- **Done when**:
  - D1. `git diff --name-only <BASE>..HEAD -- . ':!docs-site/vercel.json' ':!docs-site/content/*/changelog/_index.md' ':!docs-site/layouts/partials/site-footer.html' ':!.moai/specs/SPEC-V3R6-DOCS-PREREWRITE-REDIR-001' ':!.moai/reports/t1718'` prints nothing. Its positive control is `git diff --name-only <BASE>..HEAD -- docs-site/vercel.json`, which prints `docs-site/vercel.json`.
  - D2. The `swept: redirects=<N>` first line of AC-PRC-010 D1 equals `39 - 3*[Q15 Default] - 2*[Q21 = (a)] - k`, with k = 2 for Q23 (a), 1 for (c), 0 for (b). The pre-change value is 39 (E-R4).
  - D3. When Q21 is not (a): `git diff --name-only <BASE>..HEAD -- docs-site/layouts/partials/site-footer.html` prints nothing.
- **Does not detect:** an edit inside a rule that stays (for example a changed `permanent` flag) that leaves the count unchanged. The run phase reads `git diff -U0 <BASE>..HEAD -- docs-site/vercel.json` and confirms every changed line belongs to a block whose source a recorded answer names; that reading is by eye, not by command.

### AC-REDIR-002 — no internal reference is orphaned (REQ-REDIR-001) — regression-guard

- **Given** the references of E-R9: one footer link to `/changelog/`, seventeen lines in thirteen files that link to `utility-commands/moai-mx`, and none to `/agency` or a `design` section.
- **When** the change is committed.
- **Then** each reference still resolves, and the changelog URL still serves something.
- **Done when**:
  - D1. Q21 is not (a): `git grep -n -E 'href="/changelog' HEAD -- docs-site` prints the same one footer line as E-R9a, and `python3 -I <scratch>/redirect-check.py shadow docs-site` plus the rule list show that the root-level rule at the former line 118 remains (its pre-change line is in E-R2).
  - D2. Q21 is (a): the footer line of E-R9a is absent or names a target that exists in the built output (`<scratch>/public/ko/index.html` contains an `href` whose target path has an `index.html` in `<scratch>/public`; the build is the one of AC-REDIR-003 D2).
  - D3. The change does not remove both the changelog rules and the changelog pages: `git grep -c -E '"source": "[^"]*changelog' HEAD -- docs-site/vercel.json` prints a count above zero, or D4 of AC-PRC-011 prints four lines.
  - D4. `git grep -c -E '/utility-commands/moai-mx' HEAD -- docs-site ':!docs-site/vercel.json'` prints the same thirteen file entries as E-R9c (the references stay), and AC-PRC-011 D1 to D3 hold (so the references resolve) under Q23 options (a) and (c). Under option (b) the references stay shadowed by the exempt rule; that is the recorded residual of the amendment.
- **Does not detect:** a reference written in a form the greps do not match (a `relref` shortcode, a relative link, a script-built URL, a link inside `static/`); E-R9d found no `relref` or `ref` shortcode use in `content/` or `layouts/` at BASE (the only matches are inside the vendored theme), but that is a pre-change observation. D3 and D4 detect a missing rule or page only by name.

### AC-REDIR-003 — locale parity and the docs build stay clean (REQ-REDIR-003) — regression-guard

- **Given** the strict i18n check reports `Errors: 0` and `Warnings: 0` at BASE (E-R10), and that a ko-only deletion yields three warnings while an en-only deletion yields an error (E-R10b).
- **When** the change is committed and the checks run from the repository root.
- **Then** both checks stay clean.
- **Done when**:
  - D1. `bash scripts/docs-i18n-check.sh` exits 0 and its summary reads `Errors:   0` and `Warnings: 0` (the tail of the output; output beyond 50 lines goes to a file). The exit code alone is not enough: a ko-only deletion exits 0 with warnings (E-R10b).
  - D2. `hugo --minify --gc --source docs-site --destination <scratch>/public` exits 0 and prints no `WARN` or `ERROR` line (the recipe of `.claude/skills/hns-oss-docs-verify/SKILL.md` §1, with the destination moved to scratch). The baseline for this command was not measured in the plan phase (G2); M0 records it on the kickoff tree before any edit, and this criterion is unadopted until that baseline is green. The build needs no Go toolchain: `hugo.toml` declares no module (E-R13).
- **Does not detect:** a deleted page whose absence only a content reader would notice; a Vercel-side build difference (the Vercel install step `hugo mod get -u` is not run locally); the deployed result.

## C. Evidence ledger

Each entry gives the command, exit code, stdout, and attribution. Entries marked (wrapper) took their exit code from a measurement wrapper outside the cell.

### E-R0 — the instrument text (sha256 `768a46f948d034a9ef5235495e4dae5ac5699f175b36b585ba1ee4920312a014`, 96 lines)

The run phase writes this block verbatim to `<scratch>/redirect-check.py` and checks `shasum -a 256 <scratch>/redirect-check.py` against the value above before use.

```python
import json, os, re, sys

# redirect-check: modes "shadow" (REQ-PRC-011) and "resolve" (REQ-PRC-010).
# argv: <mode> <root>; <root> holds vercel.json and content/<locale>/**.md
# exit 0 = no finding, 1 = findings, 2 = unreadable input or unsupported pattern token
def die(msg):
    print("redirect-check: " + msg, file=sys.stderr)
    sys.exit(2)

if len(sys.argv) != 3 or sys.argv[1] not in ("shadow", "resolve"):
    die("usage: <shadow|resolve> <root>")
mode, root = sys.argv[1], sys.argv[2]
try:
    text = open(os.path.join(root, "vercel.json"), encoding="utf-8").read()
    cfg = json.loads(text)
except (OSError, ValueError) as e:
    die("cannot read vercel.json: %s" % e)
lines = text.splitlines()

def line_of(src):
    needle = '"source": ' + json.dumps(src)
    for i, l in enumerate(lines, 1):
        if needle in l:
            return i
    return 0

# page URL set derived from content/ (draft pages excluded; clean URLs, no trailing slash)
cdir = os.path.join(root, "content")
if not os.path.isdir(cdir):
    die("no content directory under " + root)
pages, sections = set(), set()
for loc in sorted(os.listdir(cdir)):
    ldir = os.path.join(cdir, loc)
    if not os.path.isdir(ldir):
        continue
    for dp, dn, fn in os.walk(ldir):
        for f in fn:
            if not f.endswith(".md"):
                continue
            body = open(os.path.join(dp, f), encoding="utf-8").read(4096)
            if re.search(r"^draft:\s*true\s*$", body.split("\n---", 1)[0], re.M):
                continue
            rel = os.path.relpath(dp, cdir).replace(os.sep, "/")
            url = "/" + rel if f in ("_index.md", "index.md") else "/" + rel + "/" + f[:-3]
            pages.add(url)
            d = rel
            while d and d != ".":
                sections.add("/" + d)
                d = d.rpartition("/")[0]
exists = pages | sections
locales = sorted(l for l in os.listdir(cdir) if os.path.isdir(os.path.join(cdir, l)))

LOC = re.compile(r":locale\(([a-z|]+)\)")
def expand_source(src):
    # path-to-regexp subset; any other token fails closed
    s = LOC.sub(lambda m: "(?:%s)" % m.group(1), src)
    s = re.sub(r"/:[A-Za-z]+\*", "(?:/.*)?", s)
    if re.search(r"[:*?{}()]", re.sub(r"\(\?:[a-z|]+\)|\(\?:/\.\*\)\?", "", s)):
        die("unsupported source token: " + src)
    return re.compile(s)

def src_locales(src):
    m = LOC.search(src)
    return m.group(1).split("|") if m else [None]

rules = cfg.get("redirects", [])
if not rules or not pages:
    die("empty sweep: redirects=%d pages=%d" % (len(rules), len(pages)))
print("swept: redirects=%d pages=%d locales=%s" % (len(rules), len(pages), ",".join(locales)))
found = 0
rewrite_sources = {r.get("source") for r in cfg.get("rewrites", [])}
for r in rules:
    src, dst = r["source"], r["destination"]
    ln = line_of(src)
    if mode == "shadow":
        rx = expand_source(src)
        for u in sorted(exists):
            if rx.fullmatch(u) or rx.fullmatch(u + "/"):
                found += 1
                print("%s:%d: source %s shadows page %s" % ("vercel.json", ln, src, u))
    else:
        for loc in src_locales(src):
            d = dst.replace(":locale", loc) if loc else dst
            if re.match(r"https?://", d):
                found += 1
                print("%s:%d: source %s -> external destination %s" % ("vercel.json", ln, src, d))
                continue
            base = re.sub(r"/:[A-Za-z]+\*$", "", d)
            if base == "":
                base = "/"
            if base in exists or (base == "/" and "/" in rewrite_sources):
                continue
            found += 1
            print("%s:%d: source %s -> destination %s resolves to no page%s" % ("vercel.json", ln, src, d, " (locale %s)" % loc if loc else ""))
print("%s: %d" % ("shadowed" if mode == "shadow" else "unresolved", found))
sys.exit(1 if found else 0)
```

Reading: `pages=612` counts one URL per non-draft `.md` file; section URLs (a directory that holds at least one page) are matched as well, so a rule that matches a section index is also reported.

### E-R1 — tree pin and cleanliness

- `git rev-parse HEAD` — exit 0 — `0ddf5ca48bed7aaeeb68ed4fc686bcfa098a82b1`
- `git rev-parse 0ddf5ca48b^{tree}` — exit 0 — `a2c0d16b90c0e49371b6efec3084e6e270781bd7`
- `git rev-parse 0ddf5ca48b:docs-site/vercel.json` — exit 0 — `959e3de56575cf12306e1d787b9c9c9f2265e219`
- `git status --short` — exit 0 — no output (clean tree).
- `git diff --stat 0ddf5ca48b -- docs-site` — exit 0 — no output (working `docs-site` equals the pin).
- `git diff --stat 8816de4dae 0ddf5ca48b -- docs-site/vercel.json docs-site/content/ko/changelog docs-site/content/en/changelog docs-site/content/ja/changelog docs-site/content/zh/changelog 'docs-site/content/*/utility-commands/moai-mx.md' docs-site/layouts docs-site/data` — exit 0 — no output (the umbrella's line numbers at 8816de4dae are unchanged at 0ddf5ca48b).

### E-R2 — redirect rule lines in scope (RED-now context)

- Command: `git grep -n -E '"source": "[^"]*(agency|changelog|moai-mx)' 0ddf5ca48b -- docs-site/vercel.json` — exit 0 — stdout:

```text
0ddf5ca48b:docs-site/vercel.json:88:      "source": "/:locale(ko|en|ja|zh)/agency/:path*",
0ddf5ca48b:docs-site/vercel.json:93:      "source": "/:locale(ko|en|ja|zh)/agency",
0ddf5ca48b:docs-site/vercel.json:98:      "source": "/agency/:path*",
0ddf5ca48b:docs-site/vercel.json:113:      "source": "/:locale(ko|en|ja|zh)/changelog",
0ddf5ca48b:docs-site/vercel.json:118:      "source": "/changelog",
0ddf5ca48b:docs-site/vercel.json:163:      "source": "/:locale(ko|en|ja|zh)/utility-commands/moai-mx",
0ddf5ca48b:docs-site/vercel.json:168:      "source": "/utility-commands/moai-mx",
```

- E-R2b. `git grep -n -E '"destination": "http' 0ddf5ca48b -- docs-site/vercel.json` — exit 0 — stdout:

```text
0ddf5ca48b:docs-site/vercel.json:114:      "destination": "https://github.com/modu-ai/moai-adk/releases",
0ddf5ca48b:docs-site/vercel.json:119:      "destination": "https://github.com/modu-ai/moai-adk/releases",
```

- E-R2c. `git grep -c '"source"' 0ddf5ca48b -- docs-site/vercel.json` — exit 0 — `0ddf5ca48b:docs-site/vercel.json:42` (39 redirects, one rewrite, two header blocks; the instrument's `swept: redirects=39` is the redirect count).
- Reading: the rule blocks run from line 87 (the `{` before the first `agency` source) to line 101, from 112 to 121, and from 162 to 171. The card's "lines 87 to 101", "113 to 121", and "163, 168" match the source lines above; no drift.

### E-R3 — the pages in scope exist (positive control for the page-presence commands)

- Command: `git ls-tree -r --name-only 0ddf5ca48b -- docs-site/content/ko/changelog docs-site/content/en/changelog docs-site/content/ja/changelog docs-site/content/zh/changelog docs-site/content/ko/utility-commands/moai-mx.md docs-site/content/en/utility-commands/moai-mx.md docs-site/content/ja/utility-commands/moai-mx.md docs-site/content/zh/utility-commands/moai-mx.md` — exit 0 — stdout (8 lines):

```text
docs-site/content/en/changelog/_index.md
docs-site/content/en/utility-commands/moai-mx.md
docs-site/content/ja/changelog/_index.md
docs-site/content/ja/utility-commands/moai-mx.md
docs-site/content/ko/changelog/_index.md
docs-site/content/ko/utility-commands/moai-mx.md
docs-site/content/zh/changelog/_index.md
docs-site/content/zh/utility-commands/moai-mx.md
```

- E-R3b. `git grep -n -E '^draft:' 0ddf5ca48b -- docs-site/content/ko/utility-commands/moai-mx.md docs-site/content/en/utility-commands/moai-mx.md docs-site/content/ja/utility-commands/moai-mx.md docs-site/content/zh/utility-commands/moai-mx.md` — exit 0 — four lines, each `...:4:draft: false`.
- E-R3c. `git ls-tree -r --name-only 0ddf5ca48b -- docs-site/content/en/design docs-site/content/en/design.md docs-site/content/ja/design docs-site/content/ja/design.md docs-site/content/ko/design docs-site/content/ko/design.md docs-site/content/zh/design docs-site/content/zh/design.md` — exit 0 — no output (no `design` path in any locale; positive control for the pathspec form is E-R3, which prints files for the same form).
- E-R3d. `git grep -n -E '^(url|slug|aliases):' 0ddf5ca48b -- docs-site/content` — exit 1 — no output (no page relocates itself by front matter).

### E-R4 — resolve mode, RED-now (AC-PRC-010)

- Command: `python3 -I <scratch>/redirect-check.py resolve docs-site` — exit 1 (wrapper) — stdout:

```text
swept: redirects=39 pages=612 locales=en,ja,ko,zh
vercel.json:88: source /:locale(ko|en|ja|zh)/agency/:path* -> destination /ko/design/:path* resolves to no page (locale ko)
vercel.json:88: source /:locale(ko|en|ja|zh)/agency/:path* -> destination /en/design/:path* resolves to no page (locale en)
vercel.json:88: source /:locale(ko|en|ja|zh)/agency/:path* -> destination /ja/design/:path* resolves to no page (locale ja)
vercel.json:88: source /:locale(ko|en|ja|zh)/agency/:path* -> destination /zh/design/:path* resolves to no page (locale zh)
vercel.json:93: source /:locale(ko|en|ja|zh)/agency -> destination /ko/design resolves to no page (locale ko)
vercel.json:93: source /:locale(ko|en|ja|zh)/agency -> destination /en/design resolves to no page (locale en)
vercel.json:93: source /:locale(ko|en|ja|zh)/agency -> destination /ja/design resolves to no page (locale ja)
vercel.json:93: source /:locale(ko|en|ja|zh)/agency -> destination /zh/design resolves to no page (locale zh)
vercel.json:98: source /agency/:path* -> destination /design/:path* resolves to no page
vercel.json:113: source /:locale(ko|en|ja|zh)/changelog -> external destination https://github.com/modu-ai/moai-adk/releases
vercel.json:113: source /:locale(ko|en|ja|zh)/changelog -> external destination https://github.com/modu-ai/moai-adk/releases
vercel.json:113: source /:locale(ko|en|ja|zh)/changelog -> external destination https://github.com/modu-ai/moai-adk/releases
vercel.json:113: source /:locale(ko|en|ja|zh)/changelog -> external destination https://github.com/modu-ai/moai-adk/releases
vercel.json:118: source /changelog -> external destination https://github.com/modu-ai/moai-adk/releases
unresolved: 14
```

Why red: 9 lines come from the `agency` rules (E-R3c: no `design` path), 5 from the external destinations (E-R2b). The remaining 32 rules print no line.

### E-R5 — shadow mode, RED-now (AC-PRC-011)

- Command: `python3 -I <scratch>/redirect-check.py shadow docs-site` — exit 1 (wrapper) — stdout:

```text
swept: redirects=39 pages=612 locales=en,ja,ko,zh
vercel.json:113: source /:locale(ko|en|ja|zh)/changelog shadows page /en/changelog
vercel.json:113: source /:locale(ko|en|ja|zh)/changelog shadows page /ja/changelog
vercel.json:113: source /:locale(ko|en|ja|zh)/changelog shadows page /ko/changelog
vercel.json:113: source /:locale(ko|en|ja|zh)/changelog shadows page /zh/changelog
vercel.json:163: source /:locale(ko|en|ja|zh)/utility-commands/moai-mx shadows page /en/utility-commands/moai-mx
vercel.json:163: source /:locale(ko|en|ja|zh)/utility-commands/moai-mx shadows page /ja/utility-commands/moai-mx
vercel.json:163: source /:locale(ko|en|ja|zh)/utility-commands/moai-mx shadows page /ko/utility-commands/moai-mx
vercel.json:163: source /:locale(ko|en|ja|zh)/utility-commands/moai-mx shadows page /zh/utility-commands/moai-mx
shadowed: 8
```

Why red: the two locale-aware sources at lines 113 and 163 each match four existing pages (E-R3). The root-level sources at lines 118 and 168 print nothing, because no content page has a root-level URL.

### E-R6 — instrument controls on synthetic fixtures (scratch, outside the repository)

Fixtures live under `<scratch>/ctl/`: each root holds two locales (ko, en), each with `a/_index.md` and `a/b.md` (both `draft: false`), and a `vercel.json` with the redirect list given. Built by `mkctl.py`; exit codes from the wrapper `runctl.py`.

| Fixture | `vercel.json` redirects | Command | Exit | Output (verbatim except the wrapper lines) |
|---------|-------------------------|---------|------|---------------------------------------------|
| `pos` (positive control: must go red) | `/:locale(ko\|en)/a/b` -> `/:locale/a`; `/:locale(ko\|en)/x` -> `/:locale/missing` | `... shadow ctl/pos` | 1 | `swept: redirects=2 pages=4 locales=en,ko` then two `shadows page /en/a/b` / `/ko/a/b` lines, `shadowed: 2` |
| `pos` | (same) | `... resolve ctl/pos` | 1 | `swept: redirects=2 pages=4 locales=en,ko`, two `resolves to no page` lines for `/ko/missing` and `/en/missing`, `unresolved: 2` |
| `neg` (negative control: must stay green with a non-empty sweep) | `/:locale(ko\|en)/old` -> `/:locale/a/b` | `... shadow ctl/neg` | 0 | `swept: redirects=1 pages=4 locales=en,ko`, `shadowed: 0` |
| `neg` | (same) | `... resolve ctl/neg` | 0 | `swept: redirects=1 pages=4 locales=en,ko`, `unresolved: 0` |
| `empty` (empty sweep: must fail closed) | `[]` | `... shadow ctl/empty` and `... resolve ctl/empty` | 2 | stderr `redirect-check: empty sweep: redirects=0 pages=4`, no stdout |
| `tok` (unsupported token: must fail closed) | `/ko/:slug` -> `/ko/a` | `... shadow ctl/tok` | 2 | stdout `swept: redirects=1 pages=4 locales=en,ko`; stderr `redirect-check: unsupported source token: /ko/:slug` |

Reading: `resolve` mode does not parse the source grammar except for the `:locale(...)` group, so `resolve ctl/tok` exits 0 (`unresolved: 0`); a rule with an unsupported source token is stopped by `shadow` mode, not by `resolve` mode (G3).

### E-R7 — scratch copies of the real redirect surface (green path and mutants, outside the repository)

Built by `mkreal.py`: each copy holds `docs-site/vercel.json` and `docs-site/content/` (without `_meta.yaml`) from the pinned tree, with the change named below applied to the copy only. Exit codes from the wrapper.

| Copy | Change applied in scratch | `shadow` | `resolve` |
|------|---------------------------|----------|-----------|
| `unchanged` | none | exit 1, `swept: redirects=39 pages=612`, `shadowed: 8` | exit 1, `unresolved: 14` |
| `q23a` | both moai-mx rules removed | exit 1, `swept: redirects=37 pages=612`, `shadowed: 4` (the four changelog pages) | exit 1, `unresolved: 14` |
| `allrules` | moai-mx pair, three `agency` rules, and two changelog rules removed (pages untouched) | exit 0, `swept: redirects=32 pages=612`, `shadowed: 0` | exit 0, `unresolved: 0` |
| `q16del` | four changelog pages deleted (rules untouched) | exit 1, `swept: redirects=39 pages=608`, `shadowed: 4` (the four moai-mx pages) | exit 1, `unresolved: 14` |
| `mutant-mxpages` | four moai-mx pages deleted, rules untouched | exit 1, `swept: redirects=39 pages=608`, `shadowed: 4` (the changelog pages only) | exit 1, `unresolved: 14` |

Reading: `allrules` is the observed green path (32 = 39 - 3 - 2 - 2) for the combination Q15 Default, Q21 (a), Q23 (a); it leaves the changelog pages in place, and § D shows why that is the only coherent pairing for Q21 (a). `mutant-mxpages` shows the shadow check alone cannot tell a removed rule from a deleted page: D2 of AC-PRC-011 exists for that reason.

### E-R8 — deployed site, external observation (never a pass; deployed commit unknown)

Command form (one invocation per URL; run from any directory): `curl -s -o /dev/null -m 15 -w 'http=%{http_code} location=%{redirect_url}\n' https://adk.mo.ai.kr<path>`. Observed 2026-10-18 between 16:48Z and 16:54Z:

```text
/ko/utility-commands/moai-mx => http=308 location=https://adk.mo.ai.kr/ko/utility-commands
/en/utility-commands/moai-mx => http=308 location=https://adk.mo.ai.kr/en/utility-commands
/ja/utility-commands/moai-mx => http=308 location=https://adk.mo.ai.kr/ja/utility-commands
/zh/utility-commands/moai-mx => http=308 location=https://adk.mo.ai.kr/zh/utility-commands
/utility-commands/moai-mx    => http=308 location=https://adk.mo.ai.kr/ko/utility-commands
/ko/changelog  => http=307 location=https://github.com/modu-ai/moai-adk/releases
/en/changelog  => http=307 location=https://github.com/modu-ai/moai-adk/releases
/ja/changelog  => http=307 location=https://github.com/modu-ai/moai-adk/releases
/zh/changelog  => http=307 location=https://github.com/modu-ai/moai-adk/releases
/changelog     => http=307 location=https://github.com/modu-ai/moai-adk/releases
/changelog/    => http=308 location=https://adk.mo.ai.kr/changelog
/ko/agency     => http=308 location=https://adk.mo.ai.kr/ko/design/
/ko/design/    => http=308 location=https://adk.mo.ai.kr/ko/design
/ko/design     => http=404 location=
/agency        => http=308 location=https://adk.mo.ai.kr/design/
/design        => http=404 location=
/ko/agency/x   => http=308 location=https://adk.mo.ai.kr/ko/design/x
```

Controls on the same host: `/ko/utility-commands/moai-fix` => `http=200`; `/ko/utility-commands/moai-gate` => `http=200`; `/ko/core-concepts/book` => `http=200`; `/en/book` => `http=308 location=https://adk.mo.ai.kr/en/core-concepts/book`.

Reading: the deployed redirect rules fire for URLs whose page file exists in the tree, so a Vercel redirect takes precedence over the built page at least for this deployment. The deployed commit is not shown by these requests and may differ from 0ddf5ca48b (G1).

### E-R9 — internal references to the URLs this child alters

- E-R9a. `git grep -n -E 'href="/changelog' 0ddf5ca48b -- docs-site` — exit 0 — stdout: `0ddf5ca48b:docs-site/layouts/partials/site-footer.html:9:        <a href="/changelog/">{{ i18n "footer_changelog" }}</a> ·`
- E-R9b. `git grep -n -E '\]\((/(ko|en|ja|zh))?/changelog' 0ddf5ca48b -- docs-site` — exit 1 — no output (no markdown link to the changelog URL).
- E-R9c. `git grep -n -E '/utility-commands/moai-mx' 0ddf5ca48b -- docs-site ':!docs-site/vercel.json'` — exit 0 — 17 lines in 13 files: per locale `advanced/mx-scanner-internals.md` (lines 7 and 78), `cli-reference/graph.md` (line 113), `workflow-commands/_index.md` (line 71; line 75 in ko); and `docs-site/data/menu/main.yaml` line 284 (`ref: /utility-commands/moai-mx`). The four `utility-commands/_meta.yaml` files also list the page as the key `"moai-mx":` (line 20), a key the pattern does not match. Whether `main.yaml` renders is not verified: `hugo.toml` sets `geekdocMenuBundle = false`.
- E-R9d. `git grep -n -E '(\]\(|href="|ref: |relref )/?((ko|en|ja|zh)/)?(agency|design)([/)" #]|$)' 0ddf5ca48b -- docs-site ':!docs-site/vercel.json'` — exit 1 — no output (no internal reference to `/agency` or a `/design` section). `git grep -n -E '\]\((/(ko|en|ja|zh))?/changelog|relref|\{\{< ref' 0ddf5ca48b -- docs-site` matches only the vendored theme (`themes/hugo-geekdoc/layouts/partials/menu-extra.html:33`, `.../shortcodes/button.html:13-14`).
- Reading: the footer link `/changelog/` resolves only through the root-level rule at line 118, because pages live under locale prefixes. The footer file is outside the card's five files.

### E-R10 — strict i18n parity script, baseline (AC-REDIR-003)

- Command: `bash scripts/docs-i18n-check.sh` — exit 0 — tail of stdout:

```text
=== Summary ===
Errors:   0
Warnings: 0
OK: all 4 locales pass parity, frontmatter, H1, and glossary checks.
```

- E-R10b (scratch copies of `scripts/` and `docs-site/content/`, built by `mkparity.py`, outside the repository): deleting `changelog/_index.md` in no locale: exit 0, `Errors: 0`, `Warnings: 0`. In ko only: exit 0, `Errors: 0`, `Warnings: 3`, with `orphan translation: en/changelog/_index.md has no ko counterpart` and the same for ja and zh. In en only: exit 1, `Errors: 1`, `missing translation: en/changelog/_index.md (exists in ko)`. In all four: exit 0, `Errors: 0`, `Warnings: 0`.

### E-R11 — history of the rules and pages (observed with `git log`)

- Rule for `utility-commands/moai-mx` added: `git log --format='%h %ad %s' --date=short -S'utility-commands/moai-mx' -- docs-site/vercel.json` — `e7e8115a21 2026-07-13 docs(docs-site): restructure sections — retire db/quality-commands, consolidate utility-commands (goal/codemaps), ...`
- The same commit deleted the `moai-mx.md` page (`git log --diff-filter=D --oneline -- docs-site/content/ko/utility-commands/moai-mx.md` prints `e7e8115a21`); the page was added again by `7feec50a22 2026-07-14 docs(docs-site): full 4-locale refresh aligned to shipped CLI/template reality`; `git merge-base --is-ancestor e7e8115a21 7feec50a22` exits 0.
- Changelog pages added by `263ea2581e 2026-07-30` (#1241); the redirect to GitHub Releases added by `47f5bfde91 2026-08-10` (#1422).
- `/agency` rules added by `4271fd8a8c 2026-04-21` (SPEC-AGENCY-ABSORB-001, #682).

### E-R12 — the changelog pages' own text

`git grep -n -i -e 'redirect' -e 'リダイレクト' -e '重定向' 0ddf5ca48b -- docs-site/content/ko/changelog/_index.md docs-site/content/en/changelog/_index.md docs-site/content/ja/changelog/_index.md docs-site/content/zh/changelog/_index.md` — exit 0 — three lines: en line 7 (`This address redirects there.`), ja line 7 (`このアドレスはそちらへリダイレクトされます。`), zh line 7 (`本地址会重定向到那里。`). The ko page has no match.

### E-R13 — build and mirror facts

- `git grep -n -E '^\[module|^\[\[module|^\[\[imports' 0ddf5ca48b -- docs-site/hugo.toml` — exit 1 — no output (no Hugo module; the theme is vendored under `docs-site/themes/hugo-geekdoc`).
- `hugo version` — exit 0 — `hugo v0.160.1+extended+withdeploy darwin/arm64 BuildDate=2026-04-08T14:02:42Z VendorInfo=Homebrew` (matches `HUGO_VERSION` 0.160.1 in `docs-site/vercel.json`).
- `git ls-files internal/template/templates` filtered for `docs-site` — 0 paths (docs-site is not template-mirrored).
- `.github/workflows/docs-i18n-check.yml` contains no `redirect` or `vercel` match (`grep -n -i 'redirect\|vercel'` printed nothing): no workflow checks `vercel.json` today.

## D. Branch table and Definition of Done

### D.1 Expected end state by recorded answers

`N` and `P` are the two numbers on the instrument's first output line: `swept: redirects=N pages=P`.

- N = 39 - 3 x [Q15 Default] - 2 x [Q21 = (a)] - k, with k = 2 for Q23 (a), k = 1 for Q23 (c), k = 0 for Q23 (b).
- P = 612 - 4 x [Q16 Default].
- `shadow` end state: no finding lines and exit 0; under Q23 (b) exactly the four moai-mx lines for the exempt source and exit 1.
- `resolve` end state: no finding lines and exit 0; under Q21 (b) exactly the five external lines naming `https://github.com/modu-ai/moai-adk/releases` and exit 1.

| Q21 | Q16 | Coherent? | Public changelog URL | Footer link `/changelog/` | Files changed |
|-----|-----|-----------|----------------------|----------------------------|---------------|
| (b) rules kept | Default (pages deleted) | yes | `/changelog` and `/<locale>/changelog` still redirect to GitHub Releases | resolves via line 118 | 5 (vercel.json + 4 pages) |
| (a) rules removed | Alternate (pages kept) | yes | `/<locale>/changelog` serves the page; `/changelog` serves nothing | dead unless the footer changes | 2 (vercel.json, plus the footer file, which the card does not name; no page is deleted) |
| (a) rules removed | Default (pages deleted) | no (REQ-REDIR-001) | serves nothing | dead | n/a |
| (b) rules kept | Alternate | no | the Alternate removes the rules (b) keeps | n/a | n/a |

Reading the second row: with Q21 (a) and the Q16 Alternate no page is deleted, so the change touches two files, and one of them (`site-footer.html`) is outside the card's list of five. The card's scope is exceeded in which file changes, not in how many.

### D.2 Definition of Done

- AC-PRC-011 and AC-PRC-010: D1 to D5 / D1 to D4 hold at the closing commit, with the recorded Q23, Q16, Q15, and Q21 answers, and the verbatim output and exit code of each command are copied into progress.md §E.2 with the commit SHA.
- AC-REDIR-001, AC-REDIR-002, AC-REDIR-003 hold; AC-REDIR-003 D2 has a green baseline recorded at M0.
- The decision rows Q21 and Q23 each carry an operator verdict attributed to the operator with a UTC timestamp (a `DEFAULT-APPLIED` line does not satisfy a product-level row).
- Regression-guard criteria are re-run as observations and recorded; they are not counted as passes of the verdict.

## G. Gaps, unmeasured items, and residual risk

- **G1 — Deployed behavior is external.** E-R8 was observed on the deployed host between 16:48Z and 16:54Z on 2026-10-18. The deployed commit is not known to equal 0ddf5ca48b, and the observation is neither tree-attributed nor a pass. The claim that the deployed redirect takes precedence over the built page rests on E-R8 alone.
- **G2 — No Hugo build was run in this planning pass.** The build writes a `.hugo_build.lock` and `resources/_gen` into `docs-site` (both gitignored), and this pass wrote nothing outside the SPEC directory. AC-REDIR-003 D2 therefore has no baseline; M0 records one.
- **G3 — Instrument limits.** The instrument compares URL paths case-sensitively; reads only `source` and `destination` (not `has`, `missing`, `statusCode`, or `permanent`); expands only the `:locale(...)` group and `/:name*` tails; strips a `:path*` tail from destinations and checks the base; excludes `draft: true` pages and ignores `publishDate`, `expiryDate`, `url:`, `slug:`, and `aliases:` (E-R3d found none at BASE); does not know Hugo-generated URLs that are not content pages (taxonomy, RSS, `search.json`, sitemap, 404); does not read `docs-site/static`; and `resolve` mode does not validate source tokens (E-R6, `tok`).
- **G4 — Root-level sources match no page by design.** Lines 118 (`/changelog`) and 168 (`/utility-commands/moai-mx`) are not reported by `shadow` mode, because no content page has a root-level URL. REQ-PRC-011 is therefore not violated by them. They are in scope because Q21 and Q23 name them, and because line 118 is what keeps the footer link alive (E-R9a).
- **G5 — Compound git forms were refused.** A single compound shell command that looped over local branches to repeat the in-flight overlap measurement was refused by the worktree-isolation guard ("too complex to verify that it stays inside the worktree"). The overlap statement in spec.md is therefore the lane's observation as relayed by the leader, not re-measured here. The fallback was to state it as relayed and to scope it: it covers the five card files only. The fallback was not substituted for a measurement.
- **G6 — Sibling child SPECs.** `.moai/specs` held only the umbrella among the SPEC-V3R6-DOCS-PREREWRITE-* directories at planning time, so this child's shape follows the umbrella and the Tier M rules, not a sibling's.
- **G7 — The t1655 split documents.** The path given for them, `/Users/goos/MoAI/moai-adk-go/.moai/reports/worktrees/t1808/t1631/split-proposal.md`, holds a different card's document (SPEC-CAPACITY-RECLAIM-001, card t1631). The t1655 files were found at `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/26c6d7b7-2473-4c40-b89a-e97d49cc0edc/scratchpad/t1655-reports/.moai/reports/t1655/{child-cards.md,split-proposal.md}`, a session scratch directory that may not persist. The REDIR card text there matches the card text quoted in this SPEC (child-cards.md item 2; split-proposal.md §2 and §5 rows Q15, Q16, Q21).
- **G8 — Vercel precedence and case rules.** Whether Vercel matches a `source` case-sensitively, and whether `trailingSlash: false` with `cleanUrls: true` normalizes every variant before the redirect, was not measured beyond the URLs in E-R8.
- **G9 — Reference scan limits.** E-R9 covers text patterns for `href`, markdown links, `ref:` keys, and `relref`. It does not cover URLs assembled by script or template functions, and does not cover `docs-site/static`.
- **G10 — The footer dependency is inferred.** That `/changelog/` is served only by the root-level rule at the former line 118 follows from the content tree (no page has a root-level URL, `defaultContentLanguageInSubdir = true`) and from E-R8 (the live answer is the rule's redirect). No request was made with that rule removed, so the dead-link outcome under Q21 (a) is an inference. The same holds for `/utility-commands/moai-mx` and the rule at the former line 168.
- **G11 — Installed `moai` provenance for the lint claim.** `moai spec lint ... --strict` and the `spec_audit` MCP call ran on the installed build `moai_cp/20261007_195927-1888-g7705b365e2`, built 2026-10-18T10:45:08Z. Its commit is a strict ancestor of the tree HEAD (`git merge-base --is-ancestor 7705b365e2 0ddf5ca48b` exits 0; `git rev-list --count 7705b365e2..0ddf5ca48b` prints 42). `git diff --name-only 7705b365e2 0ddf5ca48b -- internal/spec` prints nothing, so the SPEC lint rules are byte-identical between the two commits; other packages were not compared.
- **Residual risk.** (1) An answer of Q21 (a) enlarges the change beyond the card's five files; the card's scope is not self-consistent with that answer until the leader decides. (2) Under Q23 (b) the 17 reference lines in E-R9c keep landing on the section index; the amendment records that, and does not repair it. (3) A deleted `changelog/_index.md` can reappear through a rebase that restores the file; AC-PRC-011 D4 and AC-REDIR-003 D1 are the guards, and both run at the closing commit only. (4) Q15 and Q16 carry `DEFAULT-APPLIED` verdicts that rest on external observations (E-R8) taken against a deployment of unknown commit.
