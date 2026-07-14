# SPEC-DESKTOP-NATIVE-E2E-001 — Acceptance Criteria

> Tier M. Sibling of `spec.md` (v0.1.2, status: in-progress). All executable commands live in § Executable Command Block (CMD-IDs) — never in table cells (pipe-escaping makes in-cell greps vacuous). ACs verify REACHABILITY (content anchored within its owning section, old text absent AND new text present, in BOTH trees), not bare token presence.

Path shorthands used below:

- `L-SKILL` = `.claude/skills/moai/workflows/e2e.md`
- `T-SKILL` = `internal/template/templates/.claude/skills/moai/workflows/e2e.md`
- `L-AGENT` = `.claude/agents/moai/e2e-tester.md`
- `T-AGENT` = `internal/template/templates/.claude/agents/moai/e2e-tester.md`
- `T-CMD` = `internal/template/templates/.claude/commands/moai/e2e.md.tmpl`
- `L-CMD` = `.claude/commands/moai/e2e.md`
- `EVID` = `.moai/state/verify/SPEC-DESKTOP-NATIVE-E2E-001/`

---

## §D AC Matrix

### Group A — Workflow-skill lane (REQ-DNE-001..009)

| AC | REQ | Criterion (observable) | Verify |
|----|-----|------------------------|--------|
| AC-DNE-001 | 001 | Detection Matrix section (windowed to the Detection Matrix heading range) contains the `desktop-native` row with all four positive toolkit marker families (AppKit, WinUI, Qt, GTK) — in BOTH trees | CMD-DNE-001 |
| AC-DNE-002 | 002 | Within the Detection Matrix, the Electron row and the Tauri row each appear on an earlier line than the `desktop-native` row — in BOTH trees | CMD-DNE-002 |
| AC-DNE-003 | 003 | The graceful-exit section routes `desktop-native` to the automation lane: the routing text (`desktop-native` + lane/recipe reference) appears ≥1 in BOTH trees, AND the old deferral paragraph is gone (see AC-DNE-004) | CMD-DNE-003 |
| AC-DNE-004 | 004 | The verbatim sentence `There is no opt-in automation path` returns 0 matches in BOTH skill files, AND the genuine no-target graceful-exit text ("no e2e target detected") remains ≥1 in BOTH | CMD-DNE-004 |
| AC-DNE-005 | 005 | Supported Flags: the `--platform` line contains `desktop-native` and the `--tool` line contains all 5 tokens (axcli, appium-mac2, flaui-webdriver, pywinauto, dogtail) — in BOTH trees | CMD-DNE-005 |
| AC-DNE-006 | 006 | Tool Matrix contains per-OS `desktop-native` rows naming the default toolchains (axcli / FlaUI / dogtail) with a token-cost cell — in BOTH trees | CMD-DNE-006 |
| AC-DNE-007 | 007 | Toolchain Probe table contains the desktop-native probe rows (axcli --version, appium driver list, /status, pywinauto import, dogtail import, ydotool --version) — in BOTH trees | CMD-DNE-007 |
| AC-DNE-008a | 008 | Execution Summary no longer mentions a `desktop-native` deferral (0 matches for the deferral token in the summary window) — in BOTH trees | CMD-DNE-008 |
| AC-DNE-008b | 009 | The host-OS rule text (non-host recipes are declarative; probes run only on the host OS) appears ≥1 in BOTH skill files | CMD-DNE-008 |
| AC-DNE-008c | 008 | Phase 5 report-template platform enum gains `desktop-native`: the `### Platform: {…}` template line contains `desktop-native` inside the braces (≥1) — in BOTH trees | CMD-DNE-008 |

### Group B — e2e-tester per-OS recipes (REQ-DNE-100..112)

| AC | REQ | Criterion (observable) | Verify |
|----|-----|------------------------|--------|
| AC-DNE-009 | 100 | The old stub text (`not provided by this agent`) returns 0 matches, and three OS recipe subsections (macOS / Windows / Linux) exist under the desktop-native heading — in BOTH trees | CMD-DNE-009A |
| AC-DNE-010 | 101 | macOS recipe: `axcli` default with `cargo install`, a pinned version token, and the `axcli --version` probe — in BOTH trees | CMD-DNE-009A |
| AC-DNE-011 | 102 | macOS fallback: appium-mac2 + WebdriverIO with Xcode prerequisite and `appium driver list --installed` probe — in BOTH trees | CMD-DNE-009A |
| AC-DNE-012 | 103 | TCC/Accessibility prerequisite present: grant-path text + blocker-report routing (no user prompt) — in BOTH trees | CMD-DNE-009A |
| AC-DNE-013 | 104/105 | Windows recipe: FlaUI.WebDriver default with EXPERIMENTAL caveat + pinned version + `/status` smoke probe; pywinauto fallback with `print_control_identifiers()` — in BOTH trees | CMD-DNE-009B |
| AC-DNE-014 | 106/107 | Linux recipe: dogtail 2.x default with at-spi2 prerequisite + `QT_LINUX_ACCESSIBILITY_ALWAYS_ON=1` + Wayland/ponytail caveat; ydotool/xdotool fallback PAIRED with screenshot verification — in BOTH trees | CMD-DNE-009B |
| AC-DNE-015 | 108/110 | Last-resort + token-cost doctrine present: AX-tree snapshot loop, screenshot-loop token-cost caveat with the NOT-for-CI-evidence statement, and the bounded-tail (≤50 lines / ≤2KB or equivalent) wiring — in BOTH trees | CMD-DNE-009C |
| AC-DNE-016 | 111 | Artifact Directory Conventions table contains an `e2e/desktop-native/` row — in BOTH trees | CMD-DNE-009C |
| AC-DNE-017 | 112/109 | Missing-toolchain sequence follows the probe→surface→install→re-probe pattern in the recipes, AND `WinAppDriver`/`appium-windows-driver` return 0 matches across all 4 skill/agent files | CMD-DNE-010 |

### Group C — Command surfaces (REQ-DNE-200..201)

| AC | REQ | Criterion (observable) | Verify |
|----|-----|------------------------|--------|
| AC-DNE-018 | 200 | T-CMD argument-hint `--platform` value list contains `desktop-native`; body stays <20 non-empty lines | CMD-DNE-012 |
| AC-DNE-019 | 201 | L-CMD argument-hint `--platform` value list contains `desktop-native` | CMD-DNE-012 |

### Group D — Integrity & guards (REQ-DNE-300..305)

| AC | REQ | Criterion (observable) | Verify |
|----|-----|------------------------|--------|
| AC-DNE-020 | 300 | Skill pair diff exit 0 AND agent pair diff exit 0 (byte-identical after edits) | CMD-DNE-013 |
| AC-DNE-021 | 301/303/304 | `make build` exit 0 AND `go test ./internal/template/...` exit 0 (thin-command, frontmatter-consistency, neutrality guards included) | CMD-DNE-014 |
| AC-DNE-022 | 302 | Zero diff on `model_policy.go`, `catalog_tier_audit_test.go`, `catalog_loader_test.go`; on `catalog.yaml` a `hash:`-line-only diff (content-hash regen of EXISTING entries via `make build`, REQ-DNE-302 carve-out) is PASS, while any entry/count/pin change (new/removed entry, `expectedAgentCount`/`expectedTotal` delta) is FAIL; no new file under `.claude/agents/` in either tree | CMD-DNE-015 |
| AC-DNE-023 | 305 | Deferral-wording family regex returns 0 matches across all 4 skill/agent files (baseline before edits: skill 3 + agent 1 matching lines per tree) | CMD-DNE-011 |
| AC-DNE-024 | (gate) | `moai spec lint --strict` on this SPEC's spec.md reports 0 errors (StatusGitConsistency WARNING is expected until sync close) | CMD-DNE-016 |
| AC-DNE-025 (OPTIONAL, non-gating) | 101 | macOS local smoke: `axcli --version` output captured if the tool is installed on the dev host; absence is NOT a failure (C-6) | CMD-DNE-017 |

---

## Given-When-Then Scenarios

### S1 — AppKit project routes into the macOS lane (was: graceful exit)

- **Given** a project containing an `.xcodeproj` with a macOS app target and no `electron`/`tauri` dependencies and no web/mobile markers,
- **When** `/moai e2e` runs Phase 0 detection,
- **Then** the classification is `desktop-native` (AppKit), the workflow proceeds to Phase 0.5 toolchain selection offering `axcli` as the recommended default (appium-mac2/WDIO as fallback), and the graceful-exit/deferral path is NOT taken.

### S2 — Genuinely target-less project still exits gracefully (REQ-E2E-007 preserved)

- **Given** a pure library repository with no web, mobile, desktop, or desktop-native markers,
- **When** `/moai e2e` runs Phase 0 detection,
- **Then** the workflow reports "no e2e target detected" with the marker evidence consulted and exits WITHOUT creating any `e2e/` artifacts — identical to the parent behavior.

### S3 — Windows project on a macOS host (declarative recipe, no live probe)

- **Given** a `.vcxproj` WinUI project detected as `desktop-native` while the host OS is macOS,
- **When** the workflow reaches toolchain probing,
- **Then** the Windows recipe (FlaUI.WebDriver default) is surfaced as declarative documentation for this host, no Windows probe/install is attempted, and the report states the host-OS/target-OS mismatch.

### S4 — macOS TCC permission missing (blocker, not prompt)

- **Given** the macOS lane selected with `axcli` installed but Accessibility (TCC) permission not granted to the executing terminal,
- **When** the e2e-tester attempts the first AX-tree snapshot,
- **Then** it returns a structured `## Missing Inputs`-style blocker report naming the System Settings grant path, and the ORCHESTRATOR surfaces the remediation — the agent never prompts the user and never silently fails.

## Edge Cases

- **E1 — Electron repo with a stray `.xcodeproj`**: mixed markers must resolve via most-specific-first ordering — Electron row matches first; classification is `desktop` (electron), never `desktop-native` (AC-DNE-002 guards row order).
- **E2 — Qt project without accessibility enabled**: the Linux recipe documents `QT_LINUX_ACCESSIBILITY_ALWAYS_ON=1` as a prerequisite; without it the AT-SPI tree is empty — recipe text must state this failure mode (covered by AC-DNE-014 token set).
- **E3 — Wayland non-GNOME desktop**: dogtail's Wayland support is GNOME-only (ponytail); the recipe routes non-GNOME Wayland to the ydotool + screenshot-verification fallback.
- **E4 — Screenshot-loop misuse**: a run that produces ONLY screenshot-loop evidence for an AC is non-compliant — the recipes state screenshot loops are not CI-repeatable AC evidence (AC-DNE-015).

---

## § Executable Command Block

All commands run from the project root. Evidence redirected under `EVID` (`mkdir -p .moai/state/verify/SPEC-DESKTOP-NATIVE-E2E-001` first). `L_SKILL`/`T_SKILL`/`L_AGENT`/`T_AGENT` refer to the § path shorthands.

```bash
EVID=.moai/state/verify/SPEC-DESKTOP-NATIVE-E2E-001; mkdir -p "$EVID"
L_SKILL=.claude/skills/moai/workflows/e2e.md
T_SKILL=internal/template/templates/.claude/skills/moai/workflows/e2e.md
L_AGENT=.claude/agents/moai/e2e-tester.md
T_AGENT=internal/template/templates/.claude/agents/moai/e2e-tester.md

# CMD-DNE-001 — detection-row positive markers, anchored INSIDE the Detection Matrix section (both trees)
# PASS: each of the 4 toolkit families found ≥1 within the section window, per tree (8 PASS lines total)
# Window mechanics: flag-awk (start-flag set on the heading, exit on the NEXT ##/### heading). The naive
# range form `/^### X/,/^### /` self-terminates on its own start line (emits only the heading) — prohibited.
for f in "$L_SKILL" "$T_SKILL"; do
  awk '/^### Detection Matrix/{f=1;next} f&&/^#{2,3} /{exit} f' "$f" > /tmp/dne-dm.txt
  for tok in AppKit WinUI Qt GTK; do
    c=$(grep -c "$tok" /tmp/dne-dm.txt); echo "$f $tok=$c $([ "$c" -ge 1 ] && echo PASS || echo FAIL)"
  done
done | tee "$EVID/cmd-001-detection-row.txt"

# CMD-DNE-002 — row ordering INSIDE the Detection Matrix window: electron AND tauri rows precede the
# desktop-native row (both trees). Windowing removes first-match drift from the Supported Flags line and
# the host-OS rule text that sit above the matrix in whole-file line numbering.
# PASS: for each tree, electron_line < native_line AND tauri_line < native_line (window-relative)
for f in "$L_SKILL" "$T_SKILL"; do
  awk '/^### Detection Matrix/{f=1;next} f&&/^#{2,3} /{exit} f' "$f" > /tmp/dne-dm-order.txt
  e=$(grep -n 'desktop` (electron)' /tmp/dne-dm-order.txt | head -1 | cut -d: -f1)
  t=$(grep -n 'desktop` (tauri)' /tmp/dne-dm-order.txt | head -1 | cut -d: -f1)
  n=$(grep -n '`desktop-native`' /tmp/dne-dm-order.txt | head -1 | cut -d: -f1)
  [ -n "$e" ] && [ -n "$t" ] && [ -n "$n" ] && [ "$e" -lt "$n" ] && [ "$t" -lt "$n" ] && echo "$f ORDER PASS" || echo "$f ORDER FAIL (e=$e t=$t n=$n)"
done | tee "$EVID/cmd-002-row-order.txt"

# CMD-DNE-003 — graceful-exit section routes desktop-native to the lane (both trees)
# PASS: ≥1 routing mention of desktop-native together with lane/recipe wording in the graceful-exit window
# (non-vacuity is carried by the AC-DNE-004 pairing: the old deferral sentence must simultaneously be 0)
for f in "$L_SKILL" "$T_SKILL"; do
  awk '/^### No-Target Graceful Exit/{f=1;next} f&&/^#{2,3} /{exit} f' "$f" | grep -icE 'desktop-native.*(lane|recipe|automation)|(lane|recipe|automation).*desktop-native'
done | tee "$EVID/cmd-003-graceful-route.txt"

# CMD-DNE-004 — old deferral sentence gone; genuine no-target text preserved (both trees)
# PASS: first grep 0 per tree; second grep ≥1 per tree
for f in "$L_SKILL" "$T_SKILL"; do
  echo "$f old=$(grep -c 'There is no opt-in automation path' "$f") keep=$(grep -c 'no e2e target detected' "$f")"
done | tee "$EVID/cmd-004-old-gone-keep-present.txt"

# CMD-DNE-005 — flags: --platform enum + 5 --tool tokens (both trees)
# PASS: platform line contains desktop-native; each of the 5 tool tokens ≥1 on the --tool line
for f in "$L_SKILL" "$T_SKILL"; do
  echo "$f platform=$(grep -c -- '--platform.*desktop-native' "$f")"
  for tok in axcli appium-mac2 flaui-webdriver pywinauto dogtail; do
    echo "$f tool:$tok=$(grep -- '--tool' "$f" | grep -c "$tok")"
  done
done | tee "$EVID/cmd-005-flags.txt"

# CMD-DNE-006 — Tool Matrix desktop-native default rows (both trees)
# PASS: within the Tool Matrix section window, ≥1 row each for axcli, FlaUI, dogtail
# (flag-awk window: ends at the next ##/### heading — currently `### MCP Escalation Ladder`)
for f in "$L_SKILL" "$T_SKILL"; do
  awk '/^## Tool Matrix/{f=1;next} f&&/^#{2,3} /{exit} f' "$f" > /tmp/dne-tm.txt
  for tok in axcli FlaUI dogtail; do echo "$f matrix:$tok=$(grep -c "$tok" /tmp/dne-tm.txt)"; done
done | tee "$EVID/cmd-006-tool-matrix.txt"

# CMD-DNE-007 — probe-table rows (both trees)
# PASS: each probe token ≥1 per tree
for f in "$L_SKILL" "$T_SKILL"; do
  for tok in 'axcli --version' 'driver list --installed' '/status' 'import pywinauto' 'import dogtail' 'ydotool --version'; do
    echo "$f probe:[$tok]=$(grep -c "$tok" "$f")"
  done
done | tee "$EVID/cmd-007-probe-rows.txt"

# CMD-DNE-008 — Execution Summary deferral mention gone (0) + host-OS declarative rule present (≥1)
#              + Phase 5 report-template platform enum contains desktop-native (≥1) (both trees)
for f in "$L_SKILL" "$T_SKILL"; do
  awk '/^## Execution Summary/,0' "$f" > /tmp/dne-es.txt
  echo "$f summary-deferral=$(grep -c 'deferral' /tmp/dne-es.txt) hostos=$(grep -icE 'host OS.*declarative|declarative.*host OS' "$f") report-enum=$(grep -cE '^### Platform: \{[^}]*desktop-native[^}]*\}' "$f")"
done | tee "$EVID/cmd-008-summary-hostos.txt"

# CMD-DNE-009A — agent macOS recipe reachability (both trees)
# PASS per tree: stub text 0; per-OS recipe headings (^#### …macOS/Windows/Linux) ≥3; TCC×Accessibility
# co-occurrence ≥1; axcli install+probe; mac2 fallback probe; Xcode prereq; blocker routing ≥1.
# Re-anchored: bare macOS/Windows/Linux/Accessibility tokens already match the current tree (vacuous);
# the heading regex and the TCC compound anchor are 0 today and turn ≥threshold only post-implementation.
for f in "$L_AGENT" "$T_AGENT"; do
  echo "== $f"
  echo "  stub-gone=$(grep -c 'not provided by this agent' "$f")"
  echo "  os-recipe-headings=$(grep -cE '^#### .*(macOS|Windows|Linux)' "$f")"   # PASS: >=3
  echo "  TCC-accessibility=$(grep -cE 'TCC.*[Aa]ccessibility|[Aa]ccessibility.*TCC' "$f")"   # PASS: >=1
  for tok in 'cargo install axcli' 'axcli --version' 'appium driver list --installed' 'Xcode' 'blocker'; do
    echo "  [$tok]=$(grep -c "$tok" "$f")"
  done
done | tee "$EVID/cmd-009a-agent-macos.txt"

# CMD-DNE-009B — agent Windows + Linux declarative recipes (both trees)
# EXPERIMENTAL caveat is verified as a FlaUI co-occurrence (a bare EXPERIMENTAL grep matches the existing
# Electron-recipe caveat today — vacuous); the compound anchor is 0 today, >=1 post-implementation.
for f in "$L_AGENT" "$T_AGENT"; do
  echo "== $f"
  echo "  [FlaUI x EXPERIMENTAL]=$(grep -icE 'FlaUI.*EXPERIMENTAL|EXPERIMENTAL.*FlaUI' "$f")"   # PASS: >=1
  for tok in 'FlaUI' '/status' 'pywinauto' 'print_control_identifiers' 'dogtail' 'at-spi2' 'QT_LINUX_ACCESSIBILITY_ALWAYS_ON=1' 'ponytail' 'ydotool' 'xdotool' 'screenshot verification'; do
    echo "  [$tok]=$(grep -ic "$tok" "$f")"
  done
done | tee "$EVID/cmd-009b-agent-win-linux.txt"

# CMD-DNE-009C — last-resort + token-cost doctrine + artifact row (both trees)
for f in "$L_AGENT" "$T_AGENT"; do
  echo "== $f"
  for tok in 'AX-tree' 'screenshot loop' 'e2e/desktop-native/' 'tokens/frame'; do
    echo "  [$tok]=$(grep -ic "$tok" "$f")"
  done
done | tee "$EVID/cmd-009c-agent-lastresort.txt"

# CMD-DNE-010 — WinAppDriver family absence across all 4 skill/agent files (expect 0 everywhere)
grep -ricE 'WinAppDriver|appium-windows-driver' "$L_SKILL" "$T_SKILL" "$L_AGENT" "$T_AGENT" | tee "$EVID/cmd-010-winappdriver-absent.txt"

# CMD-DNE-011 — deferral-wording family invariance (expect 0 in all 4 files; pre-edit baseline: skill 3 + agent 1 per tree)
grep -cE 'deferral notice|deferred to a follow-up|not yet provided|no opt-in automation path|not provided by this agent' \
  "$L_SKILL" "$T_SKILL" "$L_AGENT" "$T_AGENT" | tee "$EVID/cmd-011-deferral-family.txt"

# CMD-DNE-012 — command argument-hints + thin body (template + local); actual counts teed to EVID
{
  echo "tmpl-hint=$(grep -c -- 'desktop-native' internal/template/templates/.claude/commands/moai/e2e.md.tmpl)"
  echo "tmpl-loc=$(grep -vc '^\s*$' internal/template/templates/.claude/commands/moai/e2e.md.tmpl)"   # PASS: <20
  echo "local-hint=$(grep -c -- '--platform.*desktop-native' .claude/commands/moai/e2e.md)"
} | tee "$EVID/cmd-012-commands.txt"

# CMD-DNE-013 — byte parity of both pairs (PASS: both diffs exit 0)
diff "$L_SKILL" "$T_SKILL" > "$EVID/cmd-013-skill-pair.diff"; echo "skill-pair exit=$?"
diff "$L_AGENT" "$T_AGENT" > "$EVID/cmd-013-agent-pair.diff"; echo "agent-pair exit=$?"

# CMD-DNE-014 — build + template CI guards
make build > "$EVID/cmd-014-make-build.log" 2>&1; echo "make exit=$?"
go test ./internal/template/... > "$EVID/cmd-014-go-test.log" 2>&1; echo "gotest exit=$?"; tail -20 "$EVID/cmd-014-go-test.log"

# CMD-DNE-015 — pin integrity + no new agent files (PASS: no-touch trio diff list empty; catalog.yaml non-hash diff
# lines empty — a hash:-line-only regen is permitted per the REQ-DNE-302 carve-out; agent file counts unchanged: 10 per tree)
git diff --name-only -- internal/template/model_policy.go \
  internal/template/catalog_tier_audit_test.go internal/template/catalog_loader_test.go | tee "$EVID/cmd-015-pins.txt"
git diff -U0 -- internal/template/catalog.yaml | grep -E '^[+-][^+-]' | grep -v 'hash:' | tee "$EVID/cmd-015-catalog-nonhash.txt"   # PASS: empty output (an entry/count/pin change surfaces here as FAIL)
find .claude/agents/moai -name '*.md' | wc -l; find internal/template/templates/.claude/agents/moai -name '*.md' | wc -l   # PASS: 10 / 10 (`.md`-scoped; `ls | wc -l` counts non-agent entries)

# CMD-DNE-016 — spec lint (PASS: 0 errors; StatusGitConsistency WARNING expected until sync close)
moai spec lint --strict .moai/specs/SPEC-DESKTOP-NATIVE-E2E-001/spec.md | tee "$EVID/cmd-016-lint.txt"

# CMD-DNE-017 — OPTIONAL macOS smoke (non-gating; absence of axcli is NOT a failure)
axcli --version > "$EVID/cmd-017-axcli.txt" 2>&1; echo "axcli-probe exit=$? (non-gating)"
```

---

## Quality Gates

- **G1**: All gating ACs (AC-DNE-001..024) PASS with CMD-cited verbatim evidence under `EVID`. AC-DNE-025 is evidence-if-available (non-gating, C-6).
- **G2**: `go test ./internal/template/...` exit 0 and `make build` exit 0 (AC-DNE-021).
- **G3**: `moai spec lint --strict` → **0 errors** on spec.md (the StatusGitConsistency WARNING is structurally present for draft/in-progress SPECs until sync close — do not lint.skip it).
- **G4**: No unrelated working-tree files staged in any run-phase commit (specific-path `git add` only).

## Definition of Done

1. All 6 in-scope files edited; skill + agent pairs byte-identical across trees; command pair carries the argument-hint delta.
2. The desktop-native deferral is fully discharged: old sentence 0, deferral family 0, lane reachable from Detection Matrix → Phase 0.5 → agent recipes in BOTH trees.
3. Per-OS recipes complete per spec.md §A.2 (macOS live-capable; Windows/Linux declarative; last-resort + token-cost doctrine encoded).
4. `make build` re-embedded the templates; all CI guards green; pinned counts untouched.
5. Evidence set persisted under `EVID`; §E self-verification report delivered; run-phase commits pushed with full SPEC-ID scopes.
