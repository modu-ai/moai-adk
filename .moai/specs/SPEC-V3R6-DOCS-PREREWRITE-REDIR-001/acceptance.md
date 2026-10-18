# Acceptance Criteria — SPEC-V3R6-DOCS-PREREWRITE-REDIR-001 (revision 0.3.0; r3: plan-audit iter1 FAIL 0.72, D1 to D4 fixed)

## A. Pin, conventions, and the instrument

- **Document-level pin.** Tree commit `0ddf5ca48bed7aaeeb68ed4fc686bcfa098a82b1`, tree object `a2c0d16b90c0e49371b6efec3084e6e270781bd7`, `docs-site/vercel.json` blob `959e3de56575cf12306e1d787b9c9c9f2265e219` (E-R1). Every cell that reads a committed object names the commit in its command. The instrument cells (E-R4, E-R5, E-R6, E-R7) read the working tree; E-R1 shows that `docs-site` in the working tree equals the pinned commit (`git diff --stat 0ddf5ca48b -- docs-site` prints nothing, `git status --short` prints nothing), so those cells are attributed to the same tree. Measured on 2026-10-18 between 16:48Z and 17:01Z.
- **BASE.** In the run phase, `<BASE>` means the commit the run phase started from, recorded in progress.md §E.2 at M0. The RED-now cells below use the literal pin `0ddf5ca48b`; on a rebase the run phase re-measures and re-pins (verification-completeness §4).
- **Cell form (verification-completeness §2.1).** A RED-now cell is one read-only invocation with no `;`, `&&`, `|`, subshell, or redirection. Its verbatim stdout and exit code are separate fields. Where the Bash tool does not display a zero exit status, the status was taken from a measurement wrapper outside the cell (`runctl.py`, which runs the same command and prints the exit code); the ledger says so.
- **The instrument (E-R0).** The redirect-check script reads `<root>/vercel.json` and `<root>/content/<locale>/**.md`, derives the page URL set from the content tree, and runs one of two modes. `shadow` (REQ-PRC-011) lists every redirect source that matches an existing page or section URL. `resolve` (REQ-PRC-010) lists every redirect destination that resolves to no page, that names a locale different from the source's locale, or that is external and not admitted. Optional arguments after the root, in `resolve` mode only, name one admitted external URL and the exact sources it is admitted for. It prints `swept: redirects=<R> pages=<P> locales=<...>` first and exits 2 (fail closed) when R or P is zero, when `vercel.json` or `content/` cannot be read, or when `shadow` meets a source holding a character outside the two supported constructs (`:locale(...)` and `/:name*`; the characters rejected are `: * ? { } ( ) [ ] + ^ $ | . \` outside those constructs); it exits 1 on findings and 0 otherwise. The script text is the fenced block in E-R0; the run phase extracts it verbatim to `<scratch>/redirect-check.py` and confirms its SHA-256 before use. `<scratch>` is a session-scoped directory under `/private/tmp`, never inside the repository tree. In the plan phase it was `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/7b2f1a5a-e1bf-4787-be19-8084e37cb916/scratchpad`.
- **External cells** (E-R8) are labelled `external`, are not tree-attributed, and are never recorded as a pass (verification-completeness §2.1, undecidable disposition).
- **Classes.** `release-blocking (gated)`: re-executable on the tree; counts toward the verdict once its gate rows carry operator verdicts. `regression-guard`: a guard against a change this work could introduce; it cannot be red on the pre-change tree and is not counted toward the verdict.
- **Reading note on exit codes.** Findings exit 1. A zero-finding run with a non-empty sweep exits 0. An empty sweep or an unreadable input exits 2, so an empty result is never read as a pass.

## A.1 Traceability matrix (AC to REQ)

| AC | REQ | Gate rows | Class | Fails on the pre-change tree | Ledger entries |
|----|-----|-----------|-------|------------------------------|----------------|
| AC-PRC-011 | REQ-PRC-011 | Q23 (c, first), Q16, Q17 | release-blocking | yes (E-R5: `shadowed: 8`) | E-R5, E-R3, E-R2d; E-R6, E-R14 (controls, green path, mutants); E-R8, E-R9 (external, references) |
| AC-PRC-010 | REQ-PRC-010 | Q15, Q21 (b) | release-blocking | yes (E-R4, E-R14a) | E-R4, E-R2, E-R14a; E-R6, E-R14 |
| AC-REDIR-001 | REQ-REDIR-001 | Q21 (b) | regression-guard | no (references resolve today) | E-R9, E-R2e |
| AC-REDIR-002 | REQ-REDIR-002 | none | regression-guard | no (nothing has changed yet) | E-R2, E-R14 (numstat) |
| AC-REDIR-003 | REQ-REDIR-003 | Q16, Q17 | D0 gates (fails now); D1 and D2 are regression-guards | D0 yes (E-R3); D1 and D2 no | E-R3, E-R10, E-R10b, E-R13 |

The umbrella's AC-PRC-010 and AC-PRC-011 are imported by the umbrella (M6) by criterion ID and this child's commit SHA; AC-REDIR-001 to AC-REDIR-003 are local and are not imported. Suffixes agree: AC-REDIR-00n verifies REQ-REDIR-00n. Q23, Q21, and Q16 carry the verdicts of board record `d-20261018T170818Z-a207`; Q15 and Q17 carry `DEFAULT-APPLIED`.

## B. Acceptance criteria

The change set under test is exactly: remove the rule object with source `/:locale(ko|en|ja|zh)/utility-commands/moai-mx` (line 163) and the three `agency` rule objects (lines 88, 93, 98) from `docs-site/vercel.json`; delete the four `docs-site/content/{ko,en,ja,zh}/changelog/_index.md` pages; change nothing else. Rules kept on purpose: the two changelog rules (lines 113, 118; admitted external destination) and the root-level moai-mx rule (line 168; matches no page).

### AC-PRC-011 — no redirect shadows an existing page, including the moai-mx rule (REQ-PRC-011)

Recorded first, per the card's before-start condition (Q23; leader ruling B04, `d-20261014T060405Z-0fcd`).

- **Given** the tree at BASE, where `docs-site/vercel.json` holds a locale-aware source for `changelog` (line 113) and one for `utility-commands/moai-mx` (line 163), and the four `changelog/_index.md` and four `utility-commands/moai-mx.md` pages exist (E-R2, E-R3).
- **When** the check script runs in `shadow` mode on `docs-site` after the change, and the page-presence and rule-presence commands of D2 to D5 run.
- **Then** the end state of § D holds.
- **RED-now (E-R5).** `<scratch>/redirect-check.py shadow docs-site` exits 1 and prints `shadowed: 8`: four lines for source `/:locale(ko|en|ja|zh)/changelog` (line 113) and four for source `/:locale(ko|en|ja|zh)/utility-commands/moai-mx` (line 163). It is red for the stated reason that those two locale-aware sources each match four existing pages. The root-level sources at lines 118 and 168 print no line.
- **Green path.** The four moai-mx lines flip at M1 (line 163 removed); the four changelog lines flip at M2 (the pages deleted, the rule at line 113 kept). Observed on a scratch copy that applies exactly this change set: before `shadowed: 8` exit 1, after `shadowed: 0` exit 0 (E-R14).
- **Done when** all of the following hold at the closing commit (each is one invocation, run from the repository root):
  - D1. `python3 -I <scratch>/redirect-check.py shadow docs-site` prints `swept: redirects=35 pages=608 locales=en,ja,ko,zh`, then `shadowed: 0`, and exits 0.
  - D2. `git ls-tree -r --name-only HEAD -- docs-site/content/ko/utility-commands/moai-mx.md docs-site/content/en/utility-commands/moai-mx.md docs-site/content/ja/utility-commands/moai-mx.md docs-site/content/zh/utility-commands/moai-mx.md` prints four lines. Its pre-change output is E-R3 (positive control for the same form).
  - D3. `git grep -n -E '^draft:' HEAD -- <the same four paths>` prints four lines, each `draft: false` (pre-change: E-R3b).
  - D4. `git ls-tree -r --name-only HEAD -- docs-site/content/ko/changelog docs-site/content/en/changelog docs-site/content/ja/changelog docs-site/content/zh/changelog` prints nothing; the same pathspec pre-change prints four lines (E-R3, positive control).
  - D5. `git grep -n -F '"source": "/:locale(ko|en|ja|zh)/utility-commands/moai-mx"' HEAD -- docs-site/vercel.json` prints nothing, and `git grep -n -F '"source": "/utility-commands/moai-mx"' HEAD -- docs-site/vercel.json` prints one line (pre-change each prints one line: E-R2d).
  - D6. The controls of E-R6, E-R7, and E-R14, re-run with the extracted script, print the recorded outputs.
- **Mutants observed (E-R7, E-R14):** (1) delete the four moai-mx pages instead of the rule: `shadow` stops listing them (E-R7 `mutant-mxpages`), D2 prints nothing and fails. (2) Mark the four pages `draft: true` with the rule kept: `shadow` stops listing them (`pages=608`, `shadowed: 4`), so D1 alone would pass if the rule were also removed; D3 fails. (3) Keep the rule and add a trailing slash to its source: `shadow` still prints the four lines. (4) Use a source character the script does not support (`b.ok`): `shadow` exits 2. Not built, so no detection is claimed: renaming the page files (D2 would fail by construction, not observed) and removing line 168 as well (D5's second command would print nothing, not observed).
- **What this criterion does not detect:** a source that differs from a page URL only by letter case (observed: a `Moai-mx` variant of the rule 163 source is not reported, E-R14; whether Vercel is case-sensitive is not measured); a rule that carries `has` or `missing` conditions (the script reads only `source`); a page whose URL is changed by `url:`, `slug:`, or `aliases:` front matter (none exists at BASE, E-R3d, but the script ignores them); a shadowed static file under `docs-site/static`; any Hugo-generated URL that is not a content page; and the deployed behavior (E-R8 is external). D2 to D5 detect a deleted, drafted, renamed, or re-added item only for the paths and rule strings named.

### AC-PRC-010 — every redirect destination resolves in the source's locale, or is the admitted one (REQ-PRC-010)

- **Given** the tree at BASE, where three `agency` rules point at a `design` section that exists in no locale (E-R3c) and two changelog rules point at the external GitHub Releases URL (E-R2b).
- **When** the check script runs in `resolve` mode on `docs-site` after the change, with the admitted URL and its two admitted sources as arguments (the amended REQ-PRC-010), and the rule-presence commands of D2 and D3 run.
- **Then** the end state of § D holds.
- **What `resolve` verifies.** For each redirect, and for each locale in the source's `:locale(...)` group: the destination (with `:locale` replaced) names an existing page or section; when the destination begins with a locale segment, that segment equals the source's locale; an external destination is accepted only when it equals the admitted URL and the rule's source is one of the admitted sources. A `:path*` tail on a destination is stripped and the base is checked. Sources without a `:locale` group are checked for existence only.
- **RED-now (E-R4, E-R14a).** `resolve docs-site` without the admitted arguments exits 1 with `unresolved: 14`: nine lines for the three `agency` rules (4 + 4 + 1) and five external lines for the two changelog rules. The done-command form (with the admitted arguments) exits 1 with `admitted external: 5`, `unresolved: 9` on the pre-change tree (E-R14a). Both are red for the stated reason: `/<locale>/design[/...]` and `/design/...` resolve to no page.
- **Green path.** M1 removes the three `agency` rules. Observed on a scratch copy built by deleting exactly the four rule blocks by line (E-R14): `resolve` with the admitted arguments exits 0, `admitted external: 5`, `unresolved: 0`. The leader's change set without the three `agency` rules leaves `unresolved: 9` (E-R14).
- **Done when** all of the following hold at the closing commit:
  - D1. `python3 -I <scratch>/redirect-check.py resolve docs-site https://github.com/modu-ai/moai-adk/releases '/:locale(ko|en|ja|zh)/changelog' /changelog` prints `swept: redirects=35 pages=608 locales=en,ja,ko,zh`, `admitted external: 5`, `unresolved: 0`, and exits 0.
  - D2. `git grep -n -E '"source": "[^"]*agency' HEAD -- docs-site/vercel.json` prints nothing (exit 1); pre-change it prints three lines (E-R2).
  - D3. `git grep -n -F -e '"source": "/changelog"' -e '"source": "/:locale(ko|en|ja|zh)/changelog"' HEAD -- docs-site/vercel.json` prints two lines, and `git grep -n -F '"destination": "https://github.com/modu-ai/moai-adk/releases"' HEAD -- docs-site/vercel.json` prints two lines. Pre-change each prints the same two lines (E-R2b, E-R2e), so D3 is a presence guard, not a flip.
  - D4. The controls of E-R6 and E-R14 print the recorded outputs.
- **Mutants observed (E-R14, each on a scratch copy):** (1) a locale-aware rule whose destination is hard-coded to `/ko/cli-reference/inventory`: D1 reports three lines (locales en, ja, zh) and exits 1. (2) The two changelog destinations swapped with two inventory destinations: `admitted external: 0`, `unresolved: 5`, exit 1. (3) Every rule deleted: exit 2 (E-R6 `empty`). Mutants not built, so no detection is claimed: pointing a kept changelog rule at a different external URL is expected to be reported by the same branch as (2) but was not run.
- **What this criterion does not detect:** a destination that resolves to an existing page that is the wrong page (existence and locale, not meaning); a `:path*` tail whose base resolves while the tail does not; a destination with a query string or fragment (reported as unresolved, a visible false positive); a hard-coded locale destination on a source with no `:locale` group (existence only); a destination whose first segment is not a known locale name; whether the GitHub URL is reachable; the deployed behavior (E-R8 is external). `resolve` mode does not validate the source grammar, so a rule with an unsupported source token is stopped by `shadow` mode (exit 2), not by `resolve` (E-R6 `tok`).

### AC-REDIR-001 — no internal reference is orphaned (REQ-REDIR-001) — regression-guard

- **Given** the references of E-R9: one footer link to `/changelog/`, seventeen lines in thirteen files that link to `utility-commands/moai-mx`, and none to `/agency` or a `design` section.
- **When** the change is committed.
- **Then** each reference still resolves, and the changelog URL still serves something.
- **Done when**:
  - D1. `git grep -n -E 'href="/changelog' HEAD -- docs-site` prints the same one footer line as E-R9a, and `git grep -n -F -e '"source": "/changelog"' -e '"source": "/:locale(ko|en|ja|zh)/changelog"' HEAD -- docs-site/vercel.json` prints two lines (pre-change: E-R2e), so the footer link still resolves through the kept rule at the former line 118.
  - D2. `git diff --name-only <BASE>..HEAD -- docs-site/layouts docs-site/i18n` prints nothing (the footer file and its strings are unchanged).
  - D3. `git grep -c -E '/utility-commands/moai-mx' HEAD -- docs-site ':!docs-site/vercel.json'` prints the same thirteen file entries as E-R9c, and AC-PRC-011 D1 to D3 hold, so those references land on existing pages.
- **Does not detect:** a reference written in a form the greps do not match (a `relref` shortcode, a relative link, a script-built URL, a link inside `static/`); E-R9d found no `relref` or `ref` shortcode use in `content/` or `layouts/` at BASE (matches only inside the vendored theme), a pre-change observation. D1 confirms the rule strings are present, not that the deployed redirect works (E-R8). D3 shows the page files are present and not shadowed in the content model, not that the deployed site serves them.

### AC-REDIR-002 — only the named files and rules change (REQ-REDIR-002) — regression-guard

- **Given** the tree at BASE.
- **When** the run phase has committed its changes.
- **Then** nothing outside the five files changed, `docs-site/vercel.json` lost exactly the four named rule blocks and gained nothing, and exactly four pages were deleted.
- **Done when**:
  - D1. `git diff --name-only <BASE>..HEAD -- . ':!docs-site/vercel.json' ':!docs-site/content/*/changelog/_index.md' ':!.moai/specs/SPEC-V3R6-DOCS-PREREWRITE-REDIR-001' ':!.moai/reports/t1718'` prints nothing. The footer file is not excluded, so any change to it shows. Positive control: `git diff --name-only <BASE>..HEAD -- docs-site/vercel.json docs-site/content` prints five lines.
  - D2. `git diff --numstat <BASE>..HEAD -- docs-site/vercel.json` prints `0`, `20`, and `docs-site/vercel.json` (tab-separated): no line added and twenty deleted. Derivation: the four named rule objects (three `agency` rules and the locale-aware moai-mx rule) are five lines each. Observed on the scratch copy that deletes exactly those blocks by line: `0	20` (E-R14). Observed on mutants: an in-rule edit of one destination gives `1	21`; a rename of one unrelated rule's source gives `1	21`; swapping four destinations gives `4	24` (E-R14).
  - D3. `git diff --name-only --diff-filter=D <BASE>..HEAD -- docs-site/content` prints exactly the four `changelog/_index.md` paths, and `git diff --name-only --diff-filter=AMR <BASE>..HEAD -- docs-site/content` prints nothing.
  - D4. With D2 and the absence of the four named rules (AC-PRC-011 D5, AC-PRC-010 D2): the named blocks account for all twenty deleted lines, so no other rule was removed. The rule count `swept: redirects=35` is an observation that agrees with this, not a proof: a replacement (one rule removed, one added) keeps the count at 35 and is caught by D2, not by the count (E-R14 `mut-replace`: both checks exit 0, numstat `1	21`).
- **Does not detect:** a change to `docs-site/vercel.json` that adds and deletes nothing (none exists for a text diff); a change outside `docs-site/vercel.json` and the four pages that lies inside the excluded SPEC or report paths; and which of two equal-sized sets of four blocks was removed if the named-rule greps were not also run (D4 depends on them).

### AC-REDIR-003 — the pages are deleted together, and parity and the docs build stay clean (REQ-REDIR-003)

REQ-REDIR-003 has two clauses. D0 is the deletion clause and fails on the pre-change tree; D1 and D2 are regression-guards that pass on the pre-change tree. The criterion as a whole is therefore a gate through D0 and a guard through D1 and D2.

- **Given** the strict i18n check reports `Errors: 0` and `Warnings: 0` at BASE (E-R10), and that a ko-only deletion yields three warnings while an en-only deletion yields an error (E-R10b).
- **When** the change is committed and the checks run from the repository root.
- **Then** the four pages are gone and both checks stay clean.
- **Done when**:
  - D0. `git ls-tree -r --name-only HEAD -- docs-site/content/ko/changelog docs-site/content/en/changelog docs-site/content/ja/changelog docs-site/content/zh/changelog` prints nothing. Pre-change the same command prints four lines (E-R3), so D0 is red now and green after M2. It is the same command as AC-PRC-011 D4.
  - D1. `bash scripts/docs-i18n-check.sh` exits 0 and its summary reads `Errors:   0` and `Warnings: 0` (the tail of the output; longer output goes to a file). The exit code alone is not enough: a ko-only deletion exits 0 with warnings (E-R10b).
  - D2. `hugo --minify --gc --source docs-site --destination <scratch>/public` exits 0 and prints no `WARN` or `ERROR` line (the recipe of `.claude/skills/hns-oss-docs-verify/SKILL.md` §1 with the destination moved to scratch). The iter1 plan audit (`.moai/reports/t1718/plan-audit-iter1.md`) reports a clean Hugo build and a clean i18n run on a post-deletion copy; the command as written was not run on the repository tree by this author. The build needs no Go toolchain: `hugo.toml` declares no module (E-R13).
- **Does not detect:** a ko-only deletion by D0 alone is detected only because D0 lists all four paths (a partial deletion leaves a non-empty listing); D1 and D2 do not show that the deleted pages were the right ones; a Vercel-side build difference (the Vercel install step `hugo mod get -u` is not run locally); the deployed result.

## C. Evidence ledger

Each entry gives the command, exit code, stdout, and attribution. Entries marked (wrapper) took their exit code from a measurement wrapper outside the cell.

### E-R0 — the instrument text (sha256 `d9074c46307802d5db8cffc7931fb2083440066c57a81d4fe6caeea9f4263ebf`, 110 lines)

The run phase writes this block verbatim to `<scratch>/redirect-check.py` and checks `shasum -a 256 <scratch>/redirect-check.py` against the value above before use.

```python
import json, os, re, sys

# redirect-check: modes "shadow" (REQ-PRC-011) and "resolve" (REQ-PRC-010).
# argv: <mode> <root> [<admitted-external-url> <source> [<source> ...]]; <root> holds vercel.json and content/<locale>/**.md
# the optional arguments (resolve mode only) name one external destination and the exact sources it is admitted for
# exit 0 = no finding, 1 = findings, 2 = unreadable input or unsupported pattern token
def die(msg):
    print("redirect-check: " + msg, file=sys.stderr)
    sys.exit(2)

if len(sys.argv) < 3 or len(sys.argv) == 4 or sys.argv[1] not in ("shadow", "resolve"):
    die("usage: <shadow|resolve> <root> [<admitted-external-url> <source> ...]")
mode, root = sys.argv[1], sys.argv[2]
admit = sys.argv[3] if len(sys.argv) > 3 else None
admit_sources = set(sys.argv[4:])
admitted = 0
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
    # path-to-regexp subset; sources holding any of : * ? { } ( ) [ ] + ^ $ | . \ outside the two supported constructs fail closed
    s = LOC.sub(lambda m: "(?:%s)" % m.group(1), src)
    s = re.sub(r"/:[A-Za-z]+\*", "(?:/.*)?", s)
    if re.search(r"[:*?{}()\[\]+^$|.\\]", re.sub(r"\(\?:[a-z|]+\)|\(\?:/\.\*\)\?", "", s)):
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
                if d == admit and src in admit_sources:
                    admitted += 1
                    continue
                found += 1
                print("%s:%d: source %s -> external destination %s" % ("vercel.json", ln, src, d))
                continue
            m = re.match(r"/([A-Za-z-]+)(?:/|$)", d)
            if loc and m and m.group(1) in locales and m.group(1) != loc:
                found += 1
                print("%s:%d: source %s -> destination %s is locale %s, source locale is %s" % ("vercel.json", ln, src, d, m.group(1), loc))
                continue
            base = re.sub(r"/:[A-Za-z]+\*$", "", d)
            if base == "":
                base = "/"
            if base in exists or (base == "/" and "/" in rewrite_sources):
                continue
            found += 1
            print("%s:%d: source %s -> destination %s resolves to no page%s" % ("vercel.json", ln, src, d, " (locale %s)" % loc if loc else ""))
if admit:
    print("admitted external: %d" % admitted)
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
- E-R2d. `git grep -n -F '"source": "/:locale(ko|en|ja|zh)/utility-commands/moai-mx"' 0ddf5ca48b -- docs-site/vercel.json` — exit 0 — `...:163:      "source": "/:locale(ko|en|ja|zh)/utility-commands/moai-mx",`; and `git grep -n -F '"source": "/utility-commands/moai-mx"' 0ddf5ca48b -- docs-site/vercel.json` — exit 0 — `...:168:      "source": "/utility-commands/moai-mx",`.
- E-R2e. `git grep -n -F -e '"source": "/changelog"' -e '"source": "/:locale(ko|en|ja|zh)/changelog"' 0ddf5ca48b -- docs-site/vercel.json` — exit 0 — lines 113 and 118. The same -F form for the destination string `"destination": "https://github.com/modu-ai/moai-adk/releases"` prints lines 114 and 119.
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

Reading: these copies predate the a207 verdicts and explore other answers; the copy that applies the recorded change set is E-R14. `mutant-mxpages` shows the shadow check alone cannot tell a removed rule from a deleted page: D2 of AC-PRC-011 exists for that reason.

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

### E-R14 — the planned change set, built by line deletion, with mutants (outside the repository)

Built by `mkexact2.py` (r3) from the pinned `docs-site`. `ADM` below means the three trailing arguments `https://github.com/modu-ai/moai-adk/releases '/:locale(ko|en|ja|zh)/changelog' /changelog`. Copies are made from the original file text by deleting whole lines, so formatting and line numbers are preserved and `git diff --no-index --numstat` is meaningful:

- `before`: the unchanged `vercel.json` and `content/`.
- `good`: lines 87 to 101 (the three `agency` rule blocks) and 162 to 166 (the locale-aware moai-mx block) deleted from `vercel.json`, and the four `changelog/_index.md` pages deleted. Nothing else changed. This is the planned change set.
- `leader-only`: only lines 162 to 166 deleted, and the four pages deleted (the leader's stated set, without the Q15 Default).
- Mutants, each derived from `good`: `mut-locale` (the destination of the locale-aware inventory rule hard-coded to `/ko/cli-reference/inventory`), `mut-swap` (the two changelog destinations swapped with the two inventory destinations), `mut-replace` (the source of the unrelated `/book` rule renamed to `/book-old`, so the rule count is unchanged), `mut-dot` (the source `/:locale(ko|en|ja|zh)/book` changed to `/:locale(ko|en|ja|zh)/b.ok`).

Summary lines (verbatim), exit codes printed by the same run. The finding lines between the first and last line are omitted where noted.

```text
before        shadow                  exit=1  swept: redirects=39 pages=612   shadowed: 8
before        resolve                 exit=1  (E-R4: unresolved: 14, run without ADM)
before        resolve ADM             exit=1  swept: redirects=39 pages=612   admitted external: 5   unresolved: 9      <- E-R14a
leader-only   shadow                  exit=0  swept: redirects=38 pages=608   shadowed: 0
leader-only   resolve ADM             exit=1  swept: redirects=38 pages=608   admitted external: 5   unresolved: 9
good          shadow                  exit=0  swept: redirects=35 pages=608   shadowed: 0
good          resolve ADM             exit=0  swept: redirects=35 pages=608   admitted external: 5   unresolved: 0
good          resolve (no ADM)        exit=1  swept: redirects=35 pages=608   unresolved: 5   (the five external lines)
mut-locale    resolve ADM             exit=1  3 lines: "vercel.json:28: source /:locale(ko|en|ja|zh)/getting-started/inventory -> destination /ko/cli-reference/inventory is locale ko, source locale is en" (and ja, zh)   admitted external: 5   unresolved: 3
mut-swap      resolve ADM             exit=1  swept: redirects=35 pages=608   admitted external: 0   unresolved: 5
mut-replace   shadow                  exit=0  swept: redirects=35 pages=608   shadowed: 0
mut-replace   resolve ADM             exit=0  swept: redirects=35 pages=608   admitted external: 5   unresolved: 0
mut-dot       shadow                  exit=2  stderr: redirect-check: unsupported source token: /:locale(ko|en|ja|zh)/b.ok
mut-dot       resolve ADM             exit=0  unresolved: 0     (resolve does not parse source grammar)
```

numstat (`git diff --no-index --numstat <original vercel.json> <copy>/vercel.json`, columns added, deleted):

```text
good          0   20
mut-locale    1   21
mut-swap      4   24
mut-replace   1   21
mut-dot       1   21
```

Further mutants on copies of `before` (`mutx.py`; `shadow` mode; first and last output lines):

```text
draft: true on the four moai-mx pages, rule 163 kept   exit=1  pages=608  shadowed: 4   (the moai-mx lines disappear: drafts are excluded from the page set)
trailing slash on source 163, rule kept                exit=1  pages=612  shadowed: 8   (still detected)
case variant of source 163 (Moai-mx), rule kept        exit=1  pages=612  shadowed: 4   (the moai-mx lines are NOT detected)
```

Reading:

- AC-PRC-011 (`shadow`) turns green with the planned change set and with the leader's set alone (`leader-only`: 38 rules, `shadowed: 0`). It needs neither the `agency` removal nor the admitted-URL arguments.
- AC-PRC-010 (`resolve ADM`) is green only for `good`. The leader's set alone leaves the nine `agency` lines (`unresolved: 9`).
- The three kept in-scope rules pass as before: the locale-aware changelog rule (line 113) matches no page once the four pages are deleted and carries the admitted URL for an admitted source; the root-level `/changelog` rule (line 118) matches no page and is admitted; the root-level moai-mx rule (line 168) matches no page and its destination `/ko/utility-commands` exists. The other 32 rules print no line in either mode.
- `mut-replace` is the case that a rule count cannot catch: `swept: redirects=35`, both modes exit 0, and only the numstat (`1 21` instead of `0 20`) differs from `good`. `mut-dot` shows `resolve` mode does not stop an unsupported source token; `shadow` mode does (exit 2).
- The draft mutant shows the check treats a draft page as absent; the case mutant shows a case-variant source is not seen (G3).
- Controls on fixtures (E-R6) were re-run with the r3 script: `pos` shadow exit 1 `shadowed: 2`, `pos` resolve exit 1 `unresolved: 2`, `neg` both exit 0, `empty` exit 2, `tok` shadow exit 2; outputs identical to E-R6. E-R7 was not re-run with the r3 script.

### E-R2f — an unrelated stale pointer

`sed -n 195p .claude/agents/harness/hns-release-specialist.md` — exit 0 — the printed line ends with the words: edit.** The four docs-site/content/{en,ko,ja,zh}/changelog/_index.md pages (the line instructs the release harness about the four pages this change deletes; see G15).

## D. End state and Definition of Done

### D.1 Expected end state

- Instrument first line: `swept: redirects=35 pages=608 locales=en,ja,ko,zh` (39 - 1 - 3 rules; 612 - 4 pages).
- `shadow docs-site`: no finding lines, `shadowed: 0`, exit 0.
- `resolve docs-site https://github.com/modu-ai/moai-adk/releases '/:locale(ko|en|ja|zh)/changelog' /changelog`: `admitted external: 5`, `unresolved: 0`, exit 0. Without the admitted arguments it prints the five external lines and `unresolved: 5`, exit 1, by design.
- `git diff --numstat <BASE>..HEAD -- docs-site/vercel.json` prints `0`, `20`.
- Intended public behavior: `/changelog` and `/<locale>/changelog` redirect to GitHub Releases (rules kept); `/<locale>/utility-commands/moai-mx` serves the page; `/utility-commands/moai-mx` still redirects to `/ko/utility-commands`; the `/agency` URLs end in 404 without a redirect chain. Whether the deployed site shows this is external (E-R8).
- Files changed: `docs-site/vercel.json` and the four `changelog/_index.md` pages; the footer file is unchanged.

### D.2 Definition of Done

- **Gating criteria** (a pass is required): AC-PRC-011 D1 to D6, AC-PRC-010 D1 to D4, and AC-REDIR-003 D0. Each criterion fails on the pre-change tree through D1 (AC-PRC-011, AC-PRC-010) or D0 (AC-REDIR-003); within them AC-PRC-011 D2 and D3 and AC-PRC-010 D3 are presence guards that pass before the change and are kept because they close the mutants listed under each criterion.
- **Guards, recorded as observations at the closing commit and not counted as passes of the verdict:** AC-REDIR-001 D1 to D3, AC-REDIR-002 D1 to D4, and AC-REDIR-003 D1 and D2. A guard failure still blocks the closing commit, because each guard states a condition the change must not break; it is not a pass for REQ-REDIR-001 or REQ-REDIR-002, which have no criterion that fails on the pre-change tree.
- The verbatim output and exit code of each command are copied into progress.md §E.2 with the commit SHA. AC-REDIR-003 D2 has a baseline recorded at M0.

## G. Gaps, unmeasured items, and residual risk

- **G1 — Deployed behavior is external.** E-R8 was observed on the deployed host between 16:48Z and 16:54Z on 2026-10-18. The deployed commit is not known to equal 0ddf5ca48b, and the observation is neither tree-attributed nor a pass. The claim that the deployed redirect takes precedence over the built page rests on E-R8 alone.
- **G2 — No Hugo build was run in this planning pass.** The build writes a `.hugo_build.lock` and `resources/_gen` into `docs-site` (both gitignored), and this pass wrote nothing outside the SPEC directory. AC-REDIR-003 D2 therefore has no baseline; M0 records one.
- **G3 — Instrument limits.** The instrument compares URL paths case-sensitively; reads only `source` and `destination` (not `has`, `missing`, `statusCode`, or `permanent`); expands only the `:locale(...)` group and `/:name*` tails; strips a `:path*` tail from destinations and checks the base; excludes `draft: true` pages and ignores `publishDate`, `expiryDate`, `url:`, `slug:`, and `aliases:` (E-R3d found none at BASE); does not know Hugo-generated URLs that are not content pages (taxonomy, RSS, `search.json`, sitemap, 404); does not read `docs-site/static`; `resolve` mode does not validate source tokens (E-R6 `tok`, E-R14 `mut-dot`); `shadow` mode rejects the characters listed in the instrument bullet but accepts any other literal text; the locale comparison in `resolve` applies only to sources that carry a `:locale(...)` group and to destinations whose first segment is a known locale name, so a hard-coded locale on a locale-less source is checked for existence only; and a case-variant source is not seen (observed, E-R14).
- **G4 — Root-level sources match no page by design.** Lines 118 (`/changelog`) and 168 (`/utility-commands/moai-mx`) are not reported by `shadow` mode, because no content page has a root-level URL. REQ-PRC-011 is therefore not violated by them. They are in scope because Q21 and Q23 name them, and because line 118 is what keeps the footer link alive (E-R9a).
- **G5 — Compound git forms were refused.** A single compound shell command that looped over local branches to repeat the in-flight overlap measurement was refused by the worktree-isolation guard ("too complex to verify that it stays inside the worktree"). The overlap statement in spec.md is therefore the lane's observation as relayed by the leader, not re-measured here. The fallback was to state it as relayed and to scope it: it covers the five card files only. The fallback was not substituted for a measurement.
- **G6 — Sibling child SPECs.** `.moai/specs` held only the umbrella among the SPEC-V3R6-DOCS-PREREWRITE-* directories at planning time, so this child's shape follows the umbrella and the Tier M rules, not a sibling's.
- **G7 — The t1655 split documents.** The path given for them, `/Users/goos/MoAI/moai-adk-go/.moai/reports/worktrees/t1808/t1631/split-proposal.md`, holds a different card's document (SPEC-CAPACITY-RECLAIM-001, card t1631). The t1655 files were found at `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/26c6d7b7-2473-4c40-b89a-e97d49cc0edc/scratchpad/t1655-reports/.moai/reports/t1655/{child-cards.md,split-proposal.md}`, a session scratch directory that may not persist. The REDIR card text there matches the card text quoted in this SPEC (child-cards.md item 2; split-proposal.md §2 and §5 rows Q15, Q16, Q21).
- **G8 — Vercel precedence and case rules.** Whether Vercel matches a `source` case-sensitively, and whether `trailingSlash: false` with `cleanUrls: true` normalizes every variant before the redirect, was not measured beyond the URLs in E-R8.
- **G9 — Reference scan limits.** E-R9 covers text patterns for `href`, markdown links, `ref:` keys, and `relref`. It does not cover URLs assembled by script or template functions, and does not cover `docs-site/static`.
- **G10 — The footer dependency is inferred.** That `/changelog/` is served by the root-level rule at the former line 118 follows from the content tree (no page has a root-level URL, `defaultContentLanguageInSubdir = true`) and from E-R8 (the live answer is the rule's redirect). No request was made with that rule removed. The chosen end state keeps the rule, so the inference matters only if a later change removes it. The same holds for `/utility-commands/moai-mx` and the kept rule at the former line 168.
- **G11 — Installed `moai` provenance for the lint claim.** `moai spec lint ... --strict` and the `spec_audit` MCP call ran on the installed build `moai_cp/20261007_195927-1888-g7705b365e2`, built 2026-10-18T10:45:08Z. Its commit is a strict ancestor of the tree HEAD (`git merge-base --is-ancestor 7705b365e2 0ddf5ca48b` exits 0; `git rev-list --count 7705b365e2..0ddf5ca48b` prints 42). `git diff --name-only 7705b365e2 0ddf5ca48b -- internal/spec` prints nothing, so the SPEC lint rules are byte-identical between the two commits; other packages were not compared.
- **G12 — Admitted-URL scope.** The script admits the external URL only for the sources given as arguments (r3); the done command passes the two changelog sources. A rule that carries the URL under a third source is reported (E-R14 `mut-swap`). The sources themselves are strings supplied by the done command, not read from the requirement text, so the requirement and the command must be kept equal by hand.
- **G13 — Scratch line numbers.** In E-R14 the scratch `vercel.json` is re-serialized, so finding lines carry different line numbers than the repository file; only counts are comparable.
- **G14 — Leader's change set versus AC-PRC-010.** The leader's stated change set (line 163 plus the four pages) does not by itself turn `resolve` green (E-R14: `unresolved: 14`). The SPEC's change set also removes the three `agency` rules under the `DEFAULT-APPLIED` Q15, which does (`unresolved: 0` with the admitted URL). If Q15's Default were struck, AC-PRC-010 would stay red on those nine lines.
- **Residual risk.** (1) The kept rules at lines 113, 118, and 168 are intentional exceptions to a plain reading of "no redirect hides a page" and "every destination is internal": the first is safe because the pages are deleted, the second two match no page; the external destination is admitted by amendment only for the two changelog sources (G12). (2) The 17 reference lines in E-R9c land on the pages after line 163 is removed only on the next deployment; E-R8 is a pre-change observation. (3) A deleted `changelog/_index.md` can reappear through a rebase that restores the file; AC-PRC-011 D4 and AC-REDIR-003 D0 guard it at the closing commit only. (4) Q15 and Q17 carry `DEFAULT-APPLIED` verdicts that rest on external observations (E-R8) taken against a deployment of unknown commit.
- **G15 — A stale pointer outside the five files.** `.claude/agents/harness/hns-release-specialist.md` line 195 tells the release harness to check, at each release, the four `changelog/_index.md` pages that this change deletes (E-R2f). The five-file scope forbids fixing it here, and the reference scan of E-R9 covers `docs-site` only. Owner: the release-harness maintainers (the `hns-release-specialist` file is user-owned under the `hns-*` namespace); a follow-up card should update the line after this change lands. Until then a release run may look for pages that no longer exist.
- **G16 — Prior-revision mutant claims.** Revision 0.2.0 stated that the rule count N = 35 proves no other rule was removed or added, and that a hard-coded locale destination and a swapped admitted URL were caught. Those statements were wrong; the r3 criteria replace them and every detection claim in § B now points at an observed mutant or is marked not built.
