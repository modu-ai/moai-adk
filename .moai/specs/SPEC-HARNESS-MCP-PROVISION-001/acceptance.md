# Acceptance Criteria — SPEC-HARNESS-MCP-PROVISION-001

> SSOT for the AC matrix. 15 ACs covering all 11 REQs (100% AC→REQ coverage —
> every REQ-HMP-001..011 has at least one AC; REQ-HMP-003 → AC-HMP-014).
>
> **Verification conventions** (v0.1.2 — §C hardened for discrimination):
> 1. **Every AC discriminates.** Each AC carries at least one POSITIVE check that
>    measurably FAILS on the unmodified tree (`# baseline: 0`) and passes only once
>    the intended content is actually authored. Each command carries its measured
>    baseline inline so a future auditor can re-derive discrimination without
>    re-reading this preamble.
> 2. **No compound alternation as sole evidence.** A bare
>    `grep -c -i "A\|B\|C\|D"` with `expect >= 1` is PROHIBITED: one incidental match
>    of ONE alternative satisfies the whole check, so it passes on dead prose. Where
>    several distinct clauses must all exist, each clause gets its OWN grep with its
>    own expectation.
> 3. **Section-anchored, not file-anchored, where the requirement is positional.**
>    Positional clauses are verified inside a heading-delimited section range (§C.0),
>    never by a whole-file grep that an unrelated pre-existing section could satisfy.
> 4. **Content-token anchors, never line numbers** (line numbers drift).
> 5. **Case-sensitive** for tokens whose casing the implementation controls
>    (`Phase 3.6`, `Artifact 7`, `.mcp.json`); `-i` only where casing is genuinely
>    free. A `-i` grep on a short generic word is the primary vacuity vector.
> 6. Preservation / NO-WRITE ACs assert **absence** explicitly (`expect: 0`). These
>    are labelled `[absence guard]` and are non-discriminating BY DESIGN — they are
>    never an AC's sole evidence.
> 7. Template↔local parity is a byte `diff` on each touched file, **existence-checked
>    first** (parity between two absent files is vacuous).
> 8. This is a doc/config-only SPEC, so ACs are grep / diff / `make build` based —
>    no Go test / coverage ACs.

## §A. Given-When-Then Scenarios

### GWT-1 — Phase 3.6 provisions MCP between LSP and dev-mode

- **Given** a `/moai project` run whose Phase 3.5 (LSP) has completed and whose
  `harness-spec.yaml` declares `ui_surface: has-ui`,
- **When** the flow reaches the new Phase 3.6,
- **Then** it detects the web-frontend stack, selects recommended servers from
  `mcp-matrix.yaml` (Playwright + Chrome DevTools), surfaces them via the
  orchestrator AskUserQuestion, and — on approval — writes them into `.mcp.json` at
  the repo root; Phase 3.7 (dev-mode) then runs after Phase 3.6.

### GWT-2 — Credentialed server requires explicit per-server approval

- **Given** a backend-db stack whose matrix recommendation includes a credentialed
  server (read-only Postgres),
- **When** Phase 3.6 prepares to write that server,
- **Then** it requires an EXPLICIT per-server AskUserQuestion approval before writing
  it, and — when written — expresses the secret as `${DATABASE_URL}` (env-var form),
  never as an inlined literal token.

### GWT-3 — .mcp.json write is additive, never a clobber

- **Given** a repo that already has a `.mcp.json` with an unrelated user server,
- **When** Phase 3.6 writes the selected servers,
- **Then** the selected servers are MERGED into the existing `.mcp.json` (the
  pre-existing user server survives), and no file is written under `.moai/specs/**`.

### GWT-4 — Harness generation emits MCP fragment only when declared

- **Given** a harness whose PLAN declares MCP needs (from `harness-spec.yaml`
  `external_systems`),
- **When** the harness Builder runs GENERATE,
- **Then** it emits the OPTIONAL artifact 7 (`.mcp.json` fragment via
  `artifact_type=mcp-server`); and for a harness PLAN that declares NO MCP need,
  GENERATE omits artifact 7 and the output stays byte-identical to the
  without-artifact-7 baseline.

## §B. AC ↔ REQ Mapping

| AC | REQ | Title |
|----|-----|-------|
| AC-HMP-001 | REQ-HMP-001 | Phase 3.6 heading present, ordered between Phase 3.5 and Phase 3.7 |
| AC-HMP-002 | REQ-HMP-001 | Phase 3.6 detects stack + references the MCP matrix |
| AC-HMP-003 | REQ-HMP-002 | orchestrator-held AskUserQuestion approval + subagent-no-prompt clause |
| AC-HMP-004 | REQ-HMP-004 | mcp-matrix.yaml exists with web / mobile / backend rows |
| AC-HMP-005 | REQ-HMP-005 | 3-5 server cap + vendor-maintained preference present |
| AC-HMP-006 | REQ-HMP-006 | credentialed server per-server approval + never-auto-write |
| AC-HMP-007 | REQ-HMP-007 | .mcp.json additive/merge + project-scope + `${VAR}` + no-literal-token |
| AC-HMP-008 | REQ-HMP-008 | harness-builder.md artifact-7 section present |
| AC-HMP-009 | REQ-HMP-009 | conditional-emission clause (emit iff MCP need / byte-identical omission) |
| AC-HMP-010 | REQ-HMP-010 | doctor tolerates optional manifest mcp block — documented-tolerance grep + `DisallowUnknownFields == 0` guard |
| AC-HMP-011 | REQ-HMP-011 | template↔local byte-parity on every touched file |
| AC-HMP-012 | REQ-HMP-011 | NO-SPEC guard: no .moai/specs/ write path in project flow |
| AC-HMP-013 | REQ-HMP-011 | template neutrality + internal-content-leak guard green + 16-lang neutral matrix |
| AC-HMP-014 | REQ-HMP-003 | on user approval, `.mcp.json` written at project scope (repo root) |
| AC-HMP-015 | REQ-HMP-008 | harness-builder.md "exactly 5" prose reconciled to reflect the added artifacts |

## §C. Verification Commands (per AC)

### §C.0 Section extractors (run ONCE per shell; every §C block below depends on them)

The three functions below are the anti-vacuity engine of this §C. Copy this block into
the shell before running any AC command.

```bash
DG=".claude/skills/moai/workflows/project/doc-generation.md"
HB=".claude/skills/moai/workflows/harness-builder.md"
MX=".moai/config/sections/mcp-matrix.yaml"
MXT="internal/template/templates/.moai/config/sections/mcp-matrix.yaml"

# Phase 3.6 section body: from the '## Phase 3.6' heading to the next H2. Flattened.
p36() { awk '/^## Phase 3\.6/{f=1;next} /^## /{f=0} f' "$DG" | tr '\n' ' '; }
# Artifact 7 section body: from the '### Artifact 7' heading to the next H2/H3. Flattened.
a7()  { awk '/^### Artifact 7/{f=1;next} /^## |^### /{f=0} f' "$HB" | tr '\n' ' '; }
# GENERATE Output Contract section body: from its H2 to the next H2. Flattened.
goc() { awk '/^## GENERATE Output Contract/{f=1;next} /^## /{f=0} f' "$HB" | tr '\n' ' '; }
```

**Why a heading-delimited range and not a fixed-N context window.** A `grep -A <N>`
window can TRUNCATE the target clause when the section grows past `N` lines, and can
SPILL into the following section when it shrinks — both produce wrong verdicts, and
neither failure is visible in the output. These extractors take the range from the
section's own heading to the **next heading of equal-or-higher level**, so the window
is exactly the section: truncation is structurally impossible regardless of how long
the authored section becomes, and spill is impossible because the next heading
terminates the range. No `N` is chosen, so no `N` can be wrong.

**Why the newline flattening (`tr '\n' ' '`).** Two effects, both load-bearing:
- It removes the **line-wrap false negative**: a markdown sentence such as "the optional
  manifest `mcp` block is tolerated by `moai harness doctor`" may wrap so that `mcp` and
  `doctor` land on different lines. A per-line grep would miss the co-location.
- It makes every section grep a **binary presence check** (`grep -c` on one line returns
  0 or 1 — it counts matching LINES, not matches). This is precisely the property the
  old compound-alternation greps lacked: a multi-line `grep -c "A\|B\|C"` could report
  `>= 1` from several incidental matches of a single alternative `A`, while `B` and `C`
  were never authored at all. Here each required clause gets its own grep, each returns
  0 or 1, and ALL must return 1.

**Baseline behaviour.** On the unmodified tree neither `## Phase 3.6` nor
`### Artifact 7` exists, so `p36` and `a7` emit the empty string and every grep against
them returns `0`. That is the discrimination: these checks CANNOT pass until the
sections are actually authored.

---

### AC-HMP-001 (REQ-HMP-001) — Phase 3.6 present and ordered

```bash
# 1. The heading itself (case-sensitive, line-anchored — only a real H2 heading matches).
grep -c "^## Phase 3.6" "$DG"                       # baseline: 0 | expect: 1

# 2. Ordering asserted MECHANICALLY (3.5 < 3.6 < 3.7), not eyeballed from a grep -n dump.
L35=$(grep -n "^## Phase 3.5" "$DG" | cut -d: -f1)
L36=$(grep -n "^## Phase 3.6" "$DG" | cut -d: -f1)
L37=$(grep -n "^## Phase 3.7" "$DG" | cut -d: -f1)
{ [ -n "$L36" ] && [ "$L35" -lt "$L36" ] && [ "$L36" -lt "$L37" ]; } \
  && echo ORDER_OK || echo ORDER_FAIL               # baseline: ORDER_FAIL | expect: ORDER_OK
```

Expected: the Phase 3.6 heading exists and sits strictly between Phase 3.5 (LSP) and
Phase 3.7 (dev-mode). The line numbers here are DERIVED at run time from the
`^## Phase 3.N` content-token anchors and compared transiently — no line number is
stored as an anchor.

Hardening note: the prior form was `grep -c -i "Phase 3.6"` plus an unassertive
`grep -n` dump that a human had to read. Ordering is now a machine verdict.

### AC-HMP-002 (REQ-HMP-001) — stack detection + matrix reference

```bash
p36 | grep -c -F "mcp-matrix.yaml"     # baseline: 0 | expect: 1
p36 | grep -c -F "external_systems"    # baseline: 0 | expect: 1
p36 | grep -c -F "ui_surface"          # baseline: 0 | expect: 1
```

Expected: the Phase 3.6 section references the MCP matrix by filename AND names both
`harness-spec.yaml` stack signals it consumes.

Hardening note (the worst offender). The prior form was a whole-file
`grep -c -i "external_systems\|ui_surface\|stack\|framework detection"` which measured
**6 on the unmodified tree** — because Phase 3.2 ALREADY emits `harness-spec.yaml` and
already names `external_systems` / `ui_surface` in its schema block, and because `stack`
is a generic English word appearing throughout the file. The check passed with zero
implementation. Section-anchoring to §3.6 is what makes it discriminate: the tokens must
appear in the NEW section, not merely somewhere in the file.

### AC-HMP-003 (REQ-HMP-002) — orchestrator approval + subagent no-prompt

```bash
p36 | grep -c -F "AskUserQuestion"     # baseline: 0 | expect: 1
p36 | grep -c -i -F "subagent"         # baseline: 0 | expect: 1
p36 | grep -c -i -F "blocker report"   # baseline: 0 | expect: 1
```

Expected: approval routes through the orchestrator's AskUserQuestion channel, and the
subagent-no-prompt boundary is stated explicitly (a subagent returns a blocker report).

Hardening note: the prior form measured **7** on the unmodified tree — `doc-generation.md`
already calls `AskUserQuestion` in Phase 3.1 (plan-auditor retry) and Phase 3.5 (LSP
install prompt). Any file-anchored `AskUserQuestion` grep is therefore vacuous here.

### AC-HMP-004 (REQ-HMP-004) — matrix config file exists with rows

```bash
test -f "$MX"  && echo MX_LOCAL_OK || echo MX_LOCAL_MISSING   # baseline: MX_LOCAL_MISSING | expect: MX_LOCAL_OK
test -f "$MXT" && echo MX_TPL_OK   || echo MX_TPL_MISSING     # baseline: MX_TPL_MISSING   | expect: MX_TPL_OK

# One grep PER required row (never a single alternation standing in for four rows):
grep -c -E "^[[:space:]]*web-frontend:"      "$MX"   # baseline: 0 | expect: 1
grep -c -E "^[[:space:]]*mobile:"            "$MX"   # baseline: 0 | expect: 1
grep -c -E "^[[:space:]]*backend-db:"        "$MX"   # baseline: 0 | expect: 1
grep -c -E "^[[:space:]]*universal_starter:" "$MX"   # baseline: 0 | expect: 1

# The skill POINTS to the matrix (pointer, not a copy):
p36 | grep -c -F "mcp-matrix.yaml"                   # baseline: 0 | expect: 1
# [absence guard] the matrix ROWS are NOT duplicated into skill prose (REQ-HMP-004):
grep -c -F -- "- { name:" "$DG"                      # baseline: 0 | expect: 0
```

Expected: `mcp-matrix.yaml` exists in BOTH trees and carries all four rows
(web-frontend / mobile / backend-db / universal_starter); the skill references it rather
than duplicating the rows.

Hardening note: `grep -c -i "backend-db\|backend"` collapsed to the substring `backend`,
and `grep -c -i "mobile"` was unanchored — a single stray word anywhere in the file
satisfied a "row exists" claim. Each row is now anchored to a YAML key at line start.

### AC-HMP-005 (REQ-HMP-005) — 3-5 cap + vendor-maintained

```bash
p36 | grep -c -F "3-5"                     # baseline: 0 | expect: 1
p36 | grep -c -i -F "vendor-maintained"    # baseline: 0 | expect: 1
```

Expected: both the 3-5 server cap AND the vendor-maintained preference are stated inside
Phase 3.6.

Hardening note: the prior cap grep measured **2** on the unmodified tree — the
alternation branch `max.*server` is a loose regex that matched incidental prose. The cap
literal `3-5` is the canonical token (spec.md §D.1: "3-5 servers max per row").

### AC-HMP-006 (REQ-HMP-006) — credential per-server approval, never auto-write

```bash
p36 | grep -c -i "credential"        # baseline: 0 | expect: 1   (credential gate present)
p36 | grep -c -i -F "per-server"     # baseline: 0 | expect: 1   (per-server granularity)
p36 | grep -c -i -F "auto-write"     # baseline: 0 | expect: 1   (never-auto-write clause)
```

Expected: a credentialed server requires an EXPLICIT per-server approval and is never
auto-written. Three distinct clauses → three distinct greps.

Hardening note: the prior form alternated `credential\|token\|requires_credentials\|secret`
— the single word `token` would satisfy the whole check from any unrelated sentence. It
measured 0 today only by luck; the alternation was structurally vacuous.

### AC-HMP-007 (REQ-HMP-007) — additive merge + project scope + env-var secrets

```bash
p36 | grep -c -i -F "additive"    # baseline: 0 | expect: 1   (additive write)
p36 | grep -c -i -F "clobber"     # baseline: 0 | expect: 1   (never-clobber clause)
p36 | grep -c -F ".mcp.json"      # baseline: 0 | expect: 1   (the write target)
p36 | grep -c -F '${'             # baseline: 0 | expect: 1   (env-var expansion form)
p36 | grep -c -i -F "literal"     # baseline: 0 | expect: 1   (never inline a literal token)
# [absence guard] no .mcp.json write target under .moai/specs/ (prohibition-aware — see AC-012):
grep -E "\.moai/specs/[^ ]*mcp" "$DG" | grep -v -i -E "MUST NOT|never|not be written|NO-SPEC" | wc -l
                                  # baseline: 0 | expect: 0
```

Expected: the `.mcp.json` write is additive (merge, never clobber) at project scope;
secrets are expressed as `${VAR}`; a literal token is never inlined.

Hardening note: the prior additive grep measured **2** on the unmodified tree — the
alternation branch `merge` is a generic English word already present in the file. The
prior env-var grep measured **1** for the same reason (`env var` / `environment variable`
appear in unrelated prose). `additive` / `clobber` / `${` are tokens only the intended
content produces.

### AC-HMP-008 (REQ-HMP-008) — harness artifact-7 section present

```bash
grep -c "^### Artifact 7" "$HB"              # baseline: 0 | expect: 1
a7 | grep -c -F "artifact_type=mcp-server"   # baseline: 0 | expect: 1
a7 | grep -c -F ".mcp.json"                  # baseline: 0 | expect: 1
```

Expected: `harness-builder.md` carries an `### Artifact 7` section under the GENERATE
Output Contract, and that section names the reused `artifact_type=mcp-server` capability
and the `.mcp.json` fragment it emits.

### AC-HMP-009 (REQ-HMP-009) — conditional emission + byte-identical omission

```bash
a7 | grep -c -i -F "byte-identical"    # baseline: 0 | expect: 1   (omission is byte-identical)
a7 | grep -c -i -F "omit"              # baseline: 0 | expect: 1   (artifact 7 is omitted)
a7 | grep -c -F "external_systems"     # baseline: 0 | expect: 1   (the MCP-need derivation source)
```

Expected: artifact 7 is emitted ONLY when the harness PLAN declares an MCP need (derived
from `harness-spec.yaml` `external_systems`); when it does not, artifact 7 is OMITTED and
GENERATE output stays byte-identical to the without-artifact-7 baseline. The three greps
map 1:1 onto REQ-HMP-009's three assertions.

Hardening note (headline requirement, worst vacuity). The prior form was a single
whole-file `grep -c -i "optional\|only when\|declared.*mcp\|mcp.*need\|omit\|byte-identical\|when no MCP"`
measuring **8 on the unmodified tree** — the alternation branch `optional` alone matched
eight times in unrelated prose. The conditional-emission clause is the SPEC's headline
behaviour, and its AC passed with the feature entirely unimplemented. This is the exact
token-presence-vs-reachability failure mode that let a prior SPEC in this repo report
14/14 AC PASS with a completely inert headline feature.

### AC-HMP-010 (REQ-HMP-010) — doctor tolerates optional manifest mcp block

Verified as TOLERATE-ONLY via a documented-tolerance clause + a `DisallowUnknownFields`
regression guard. The prior repo-wide `go run ./cmd/moai harness doctor` smoke was
already DROPPED at v0.1.1 (it took no manifest argument, constructed no `mcp`-block
fixture, and its exit code depended on unrelated harnesses' state).

```bash
# 1. Documented-tolerance clause, positioned INSIDE the artifact-7 section:
a7 | grep -c -i -F "doctor"    # baseline: 0 | expect: 1
a7 | grep -c -i "tolerat"      # baseline: 0 | expect: 1   (stem: tolerate/tolerated/tolerant)

# 2. [absence guard] no strict decoding in the manifest schema package:
grep -r "DisallowUnknownFields" internal/harness/v4manifest/ | wc -l              # baseline: 0 | expect: 0
# 3. [absence guard] the decode sites stay lenient:
grep -r "DisallowUnknownFields" internal/harness/applier.go internal/cli/harness/doctor.go | wc -l   # baseline: 0 | expect: 0
```

Expected: the tolerance clause is present in the artifact-7 section AND no
`DisallowUnknownFields` appears in the manifest schema package or at either decode site —
so an optional `mcp` block is silently tolerated with zero Go change. (Active `mcp`-block
validation is OUT OF SCOPE per the resolved TOLERATE-ONLY decision; see plan.md §A
Resolved clarifications + progress.md §E.1.)

Hardening note — TWO defects fixed:
1. The tolerance grep alternated `tolerate\|optional.*mcp\|mcp.*block\|lenient\|doctor`
   and measured **1** on the unmodified tree: it matched the bare word `doctor` on the
   pre-existing ACTIVATE-phase smoke-gate line (`moai harness doctor`). It passed with
   zero implementation. The clause is now section-anchored and requires BOTH `doctor`
   AND a `tolerat` stem inside the artifact-7 section.
2. The regression guards used `grep -c "DisallowUnknownFields" <path>/*.go`, which on a
   multi-file glob prints one `file:count` line PER FILE
   (`…/command_template.go:0`, `…/isolation_test.go:0`, …) and never a single number —
   so the `# expect 0` comparison was not evaluable at all. Replaced with
   `grep -r … | wc -l`, which yields a true scalar match count.

### AC-HMP-011 (REQ-HMP-011) — template↔local byte-parity

```bash
for pair in \
  ".moai/config/sections/mcp-matrix.yaml|internal/template/templates/.moai/config/sections/mcp-matrix.yaml" \
  ".claude/skills/moai/workflows/project/doc-generation.md|internal/template/templates/.claude/skills/moai/workflows/project/doc-generation.md" \
  ".claude/skills/moai/workflows/harness-builder.md|internal/template/templates/.claude/skills/moai/workflows/harness-builder.md"
do
  L="${pair%%|*}"; T="${pair##*|}"
  if   [ ! -f "$L" ] || [ ! -f "$T" ]; then echo "MISSING: $L"
  elif diff -q "$L" "$T" >/dev/null;   then echo "PARITY OK: $L"
  else                                      echo "DRIFT: $L"; fi
done
# baseline: MISSING (mcp-matrix.yaml) + PARITY OK x2
# expect:   PARITY OK x3 — zero MISSING, zero DRIFT
```

Expected: every touched file exists in BOTH trees and is byte-identical between them.

Hardening note: existence is now checked BEFORE parity. The prior form ran
`diff -q A B && echo OK || echo DRIFT` directly — parity between two files that are both
absent is a vacuous claim, and folding "missing" into the same DRIFT bucket hid which
failure had occurred.

### AC-HMP-012 (REQ-HMP-011) — NO-SPEC scope guard

```bash
# [absence guard] no .moai/specs/ WRITE PATH. A prohibition statement that NAMES the
# path in order to forbid it is PERMITTED and filtered out — the guard targets write
# targets, not mentions.
grep -F ".moai/specs/" "$DG" | grep -v -i -E "MUST NOT|never|not be written|NO-SPEC" | wc -l
                                   # baseline: 0 | expect: 0
# [positive, discriminating] the artifact target IS the repo-root .mcp.json:
p36 | grep -c -F ".mcp.json"       # baseline: 0 | expect: 1
```

Expected: `doc-generation.md` contains no `.moai/specs/` WRITE path (prohibition
statements naming the path are permitted); the Phase 3.6 write target is the repo-root
`.mcp.json`.

Hardening note — this AC was BROKEN, not merely vacuous. The prior form was
`grep -rn "\.moai/specs/" "$DG" | wc -l  # expect 0`, which measures **1** on the
unmodified tree: `doc-generation.md` line ~92 (Phase 3.2) legitimately STATES the guard —
"[HARD] The artifact MUST NOT be written anywhere under `.moai/specs/` — the
`/moai project` NO-SPEC scope guard applies to this artifact too." The old check
conflated *naming a path in order to forbid it* with *writing to it*, so it could never
pass. A run-phase agent chasing that false FAIL would most plausibly have "fixed" it by
DELETING the legitimate NO-SPEC guard statement — actively removing the very protection
REQ-HMP-011 exists to preserve. The guard is now prohibition-aware: it counts only
`.moai/specs/` occurrences that are NOT part of a prohibition.

### AC-HMP-013 (REQ-HMP-011) — neutrality + 16-language

```bash
make build ; echo "exit=$?"                       # baseline: exit=0 | expect: exit=0  [non-regression]
go test ./internal/template/... 2>&1 | tail -2    # baseline: ok     | expect: ok      [neutrality + leak guards]
# [absence guard] no internal SPEC ID leaked into the template tree:
grep -r "SPEC-HARNESS-MCP-PROVISION-001" internal/template/templates/ | wc -l   # baseline: 0 | expect: 0

# [positive, discriminating] the TEMPLATE matrix exists and is project-TYPE-keyed
# (one grep per row — never one alternation standing in for three):
grep -c -E "^[[:space:]]*web-frontend:" "$MXT"    # baseline: 0 | expect: 1
grep -c -E "^[[:space:]]*mobile:"       "$MXT"    # baseline: 0 | expect: 1
grep -c -E "^[[:space:]]*backend-db:"   "$MXT"    # baseline: 0 | expect: 1
# [absence guard] no privileged language in the matrix (16-language neutrality):
grep -c -i -E "primary[ _-]?language|primary: (go|python|typescript)" "$MXT"    # baseline: 0 | expect: 0
```

Expected: `make build` clean; template guards green; no internal SPEC ID in the template
tree; the matrix is keyed by project TYPE (not by a privileged language), so all 16
supported languages are treated equally.

Hardening note: `grep -c -i "web-frontend\|mobile\|backend"` with `expect >= 1` was
satisfiable by ANY ONE of the three rows — a matrix with only `mobile` would have passed
a check that claims to verify all three. Each row now has its own anchored grep.

### AC-HMP-014 (REQ-HMP-003) — write-on-approval at project scope

```bash
p36 | grep -c -i -E "(on|upon|once|after) approval"   # baseline: 0 | expect: 1   (approval → write trigger)
p36 | grep -c -i -F "project scope"                   # baseline: 0 | expect: 1
p36 | grep -c -i -E "repo[- ]root"                    # baseline: 0 | expect: 1
```

Expected: Phase 3.6 documents that, ON USER APPROVAL, the selected servers are written
into the repo-root `.mcp.json` at project scope (REQ-HMP-003). This asserts the write
EVENT; AC-HMP-007 asserts the write DISCIPLINE (additive/merge + `${VAR}`).

Alternation note: the first grep alternates GRAMMATICAL VARIANTS of a single clause (the
approval→write trigger: on/upon/once/after approval), not several distinct clauses — and
it is not this AC's sole evidence (two further independent greps follow). This is
permitted. The prohibited pattern is one alternation standing in for multiple distinct
required clauses, which is what the prior form did
(`on approval\|upon approval\|approv.*writ\|writ.*approv\|…` collapsed the trigger, the
scope, and the target into a single `>= 1`).

### AC-HMP-015 (REQ-HMP-008) — "exactly 5" prose reconciled

```bash
# [reverse delta — DISCRIMINATING] the bare uncontextualized claim must be gone:
grep -c -F "exactly 5 artifact types" "$HB"   # baseline: 1 | expect: 0
# [positive] the GENERATE Output Contract names the optional artifact 7:
goc | grep -c -i "artifact 7"                 # baseline: 0 | expect: 1
goc | grep -c -i -F "optional"                # baseline: 0 | expect: 1
```

Expected: the bare `emits exactly 5 artifact types` sentence no longer stands
uncontextualized against the added artifacts, and the GENERATE Output Contract section
names the optional artifact 7 (per SPEC-HARNESS-VERIFY-PROMOTE-001, the mandatory verify
skill is artifact 6).

Note: a reconciliation that reads "exactly 5 **base** artifact types" does NOT match the
literal `exactly 5 artifact types` and therefore correctly PASSES the first check — the
guard targets the uncontextualized count, not the number 5.

Hardening note: this AC is the only one in §C whose primary check is a REVERSE delta
(baseline 1 → expect 0). The prior form
(`grep -c -i "artifact 7\|optional.*mcp\|…\|5 base\|base artifact\|verify skill"`)
measured 0 today and so appeared to discriminate, but it verified only that SOME new
prose was added SOMEWHERE in the file — it never verified that the offending "exactly 5"
sentence was actually reconciled, which is the entire point of the AC.

## §D. Edge Cases

- **E1 ambiguous / unknown stack**: when the stack cannot be classified into
  web-frontend / mobile / backend-db, Phase 3.6 falls back to the
  `universal_starter` row (GitHub + Context7 + Playwright) rather than skipping MCP
  provisioning silently.
- **E2 user declines all servers**: when the user rejects the recommendation entirely
  via AskUserQuestion, Phase 3.6 writes NO `.mcp.json` entry and proceeds to Phase 3.7
  — a declined recommendation is not an error.
- **E3 existing `.mcp.json` with overlapping server**: when the existing file already
  contains a server with the same key, the additive merge keeps the existing entry
  (no duplicate, no clobber); Phase 3.6 does not silently overwrite a user-tuned
  entry.
- **E4 credentialed server with no env var available**: when a credentialed server is
  approved but the required `${VAR}` is not set in the environment, the entry is still
  written with the `${VAR}` placeholder (config-time); actual credential resolution is
  a runtime concern (out of scope) — never inline a literal.
- **E5 harness PLAN declares no MCP need**: artifact 7 is omitted and GENERATE output
  is byte-identical to the without-artifact-7 baseline (REQ-HMP-009) — a no-MCP harness
  is unchanged from today.
- **E6 mcp-matrix config surface (RESOLVED at plan-phase)**: the decision is settled —
  `mcp-matrix.yaml` is a standalone data resource read as prose-context, no Go loader
  (recorded in progress.md §E.1 + plan.md §A Resolved clarifications). Run-phase applies
  it as-is; do NOT silently add a Go config-struct field.
- **E7 doctor validate-vs-tolerate (RESOLVED at plan-phase)**: the decision is settled —
  TOLERATE-ONLY, no Go change (recorded in progress.md §E.1 + plan.md §A Resolved
  clarifications). Run-phase applies it as-is; do NOT add `DisallowUnknownFields` or a
  Go `MCP` struct field.

## §E. Quality Gates

- TRUST 5: Tested (doc/config-only SPEC — verification is grep / diff / `make build` /
  `moai harness doctor`, not unit tests; state test ACs as N/A honestly); Readable
  (workflow prose stays consistent with surrounding sections); Unified (byte-identical
  template mirror); Secured (no literal secrets — `${VAR}` only; credentialed servers
  gated on explicit approval; NO-SPEC guard preserved); Trackable (Conventional Commits
  per milestone, `🗿 MoAI` trailer).
- Neutrality: `template-neutrality-check.yaml` + `internal_content_leak_test.go` green
  on the final push; matrix is project-type-keyed (16-language neutral).
- Doctor tolerance: documented-tolerance clause present in `harness-builder.md` +
  `DisallowUnknownFields == 0` in `internal/harness/v4manifest/*.go` and at the decode
  sites (`applier.go` / `doctor.go`) — deterministic grep+guard, no live repo-wide
  doctor smoke.
- Non-regression: `go build ./...` + `go test ./...` exit 0 (no Go code changed).

## §F. Definition of Done

1. All 15 ACs PASS (or documented N/A / PASS-WITH-DEBT with rationale) with verbatim
   command output recorded in progress.md §E.2 (run-phase evidence, owned by
   manager-develop).
2. Phase 3.6 inserted between Phase 3.5 (LSP) and Phase 3.7 (dev-mode) in
   `doc-generation.md`: stack detect → matrix select (3-5 cap, vendor-maintained) →
   orchestrator approval (per-server for credentialed) → additive `.mcp.json` write at
   project scope (`${VAR}` secrets, never literal).
3. `mcp-matrix.yaml` (web / mobile / backend + universal_starter) created under
   `.moai/config/sections/`, referenced (not duplicated) by the skill.
4. `harness-builder.md` GENERATE documents the optional artifact 7 with the
   conditional-emission (emit iff MCP need / byte-identical omission) clause, and its
   "exactly 5" prose is reconciled to the canonical order (5 base + verify skill
   artifact 6 + optional MCP fragment artifact 7); the optional manifest `mcp` block is
   doctor-tolerant.
5. Every touched file byte-identical between local and template trees; `make build` +
   neutrality guards green; doctor tolerance verified via the documented-tolerance grep
   + `DisallowUnknownFields == 0` regression guard (no live repo-wide doctor smoke).
6. Both plan-phase clarifications resolved (mcp-matrix config surface = standalone data
   resource; doctor validate-vs-tolerate = tolerate-only) and recorded in progress.md
   §E.1 — no unresolved markers remain in plan.md.
7. Sync-phase close by manager-docs per the Status Transition Ownership Matrix.
