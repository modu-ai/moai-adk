# Design — SPEC-CLI-UIKIT-KERNEL-001

The uikit leaf-package contract, helper inventory, and caller-rewrite blast-radius
map. This document is the HOW; spec.md is the WHAT/WHY. It EXPANDS
SPEC-CLI-SUBPKG-SPLIT-001 design.md §C (deps provider-injection — N/A for uikit,
verified zero deps coupling this run) and §D (import-cycle resolution — THIS is
the section this SPEC implements). SPLIT-001 design.md §A/§B/§E/§F/§G are
cross-referenced, NOT duplicated.

## §A. The uikit Leaf-Package Contract

The kernel milestone (SPLIT-001 design.md §D) resolves the import-cycle blocker for
kernel-dependent clusters by extracting the shared TUI kernel into a NEW leaf package
`internal/cli/uikit` (package `uikit`).

> **D2 RESOLUTION (iter-2 plan-auditor fix)**: `settings.go` STAYS in `package cli` — it is
> NOT moved to uikit. `settings.go:28` `mutateSettingsLocal(path, func(*SettingsLocal))` operates
> on `SettingsLocal`, which is defined at `internal/cli/glm.go:97` (verified this run via
> `grep -nE 'type SettingsLocal' internal/ -r --include='*.go'`). glm.go is a STAYS file
> (kernel-USING, calls `renderSuccessCard` at :299/:339). Moving `settings.go` into uikit would
> require uikit to reference `SettingsLocal`, creating `uikit → package cli` import cycle (cycle
> forms because uikit/settings.go would import glm.go's package cli type). This violates
> REQ-CUK-001's leaf contract. The settings helpers (`mutateSettingsLocal`, `writeFileAtomic`,
> `stripGLMCredsAndSetTeammateMode`) are therefore NOT exported from uikit; they remain in
> `package cli` as internal helpers. The uikit leaf narrows to the **rendering + schema-bridge
> kernel** only (render.go + banner.go + schema_bridge.go helpers). Source LOC drops from 503 to
> ~365; blast-radius remains 13+ production files + ≥10 test files (Tier L still justified, see
> §D.9 corrected count). REQ-CUK-007 is updated to list `SettingsLocal` as a third cross-file
> type dependency whose resolution is "settings.go STAYS in package cli" (NOT a uikit move).

**The leaf contract (REQ-CUK-001)**:

```
internal/cli/uikit/          # package uikit — NEW leaf
├── render.go                # ← moved from internal/cli/render.go (120 LOC)
├── banner.go                # ← moved from internal/cli/banner.go (146 LOC)
├── schema_bridge.go         # ← moved from internal/cli/schema_bridge.go (99 LOC) — CONDITIONAL
│                              on profileSetupText resolution (§C.4 below); b-ii split resolution
│                              keeps the profileSetupText-referencing maps in package cli
├── types.go                 # NEW — carries CheckStatus type + consts moved from doctor.go:25-34
└── *_test.go                # ← moved from internal/cli/render_test.go + parts of misc_coverage_test.go
```

The leaf contract is binding: `package uikit` imports NEITHER `github.com/modu-ai/moai-adk/
internal/cli` (the parent) NOR any `github.com/modu-ai/moai-adk/internal/cli/*` (siblings). This
is verified mechanically by AC-CUK-005 (grep on the uikit source files). External imports are
unconstrained: uikit MAY import `internal/tui`, `internal/settings`, `lipgloss`, etc. (verified
this run: the 3 source files that move — render.go, banner.go, schema_bridge.go — import
`internal/tui`, `internal/settings`, `lipgloss`, `fmt`, `os`, `runtime`, `strings`; all external,
none import the parent or a sibling). `settings.go` no longer moves (D2 STAYS reclassification
above); its imports (`encoding/json`, `fmt`, `os`, `path/filepath`) are moot for the uikit
contract.

**Why the leaf contract matters** (SPLIT-001 design.md §D rationale, specialized for uikit):
`package cli` `root.go` imports command subpackages for `AddCommand` registration. Therefore no
subpackage can import `package cli` back — that would be a compile-time import cycle. But the
shared kernel helpers (`RenderError`, `RenderCard`, `PrintBanner`, `MutateSettingsLocal`,
`SchemaKeyToTUIField`) currently live in `package cli`. Any future cluster that consumes them
(doctor, migrate, update) cannot move to a subpackage until those helpers live in a NEUTRAL LEAF
that both `package cli` and the new cluster can import. `uikit` IS that leaf.

## §B. The Extraction Recipe (cross-reference SPLIT-001 design.md §B)

M1 follows the SPLIT-001 §B 7-step recipe verbatim, specialized for the uikit single-cluster
case:

1. Create `internal/cli/uikit/` with `package uikit`.
2. Resolve the 3 cross-file type dependencies FIRST (§C below) — this is the uikit-specific
   addition to the recipe (SPLIT-001 §B did not face this because agentlint had no cross-file
   type deps). The three resolutions are: (a) CheckStatus co-locates into uikit/types.go;
   (b) profileSetupText coupling resolves via b-ii split (keep profileSetupText-referencing maps
   in package cli; move only SchemaKeyToTUIField/FieldDefTUILabel helpers to uikit); (c)
   `SettingsLocal` coupling resolves by `settings.go STAYS in package cli` (D2 option a).
3. Move the 3 source files (`git mv`) — render.go, banner.go, schema_bridge.go (or its split per
   step 2-b-ii). `settings.go` STAYS in package cli.
4. Re-scope symbols (export the externally-consumed helpers; see §C inventory).
5. Rewrite every `package cli` caller to `uikit.<Helper>` (§D blast-radius map).
6. Move the test files that exercise the moved helpers.
7. Verify + commit (recipe step 6-7; see plan.md §F M1).

## §C. Helper Inventory Moving (the 3 source files + cross-file type deps)

### §C.1 `render.go` (120 LOC) — moved to `uikit/render.go`

Exported (already):
- `RenderError(err error) string` (`render.go:105`)

Unexported → exported on move:
- `cardStyle() lipgloss.Style` (`render.go:15`) → `CardStyle` — IF externally consumed; the
  run-phase grep decides. (Likely stays unexported — it's a style factory internal to render.go.)
- `renderCard(title, content string) string` (`render.go:23`) → `RenderCard` — externally
  consumed (see §D blast radius).
- `renderKeyValue(key, value string, keyWidth int) string` (`render.go:30`) → `RenderKeyValue`.
- `renderKeyValueLines(pairs []kvPair) string` (`render.go:36`) → `RenderKeyValueLines`.
- `renderStatusLine(status CheckStatus, label, message string, labelWidth int) string`
  (`render.go:60`) → `RenderStatusLine` — NOTE consumes `CheckStatus` (see §C.5).
- `renderSuccessCard(title string, details ...string) string` (`render.go:67`) →
  `RenderSuccessCard`.
- `renderInfoCard(title string, details ...string) string` (`render.go:77`) → `RenderInfoCard`.
- `renderSummaryLine(ok, warn, fail int) string` (`render.go:86`) → `RenderSummaryLine`.

Supporting type (defined in render.go, moves with it):
- `kvPair struct { key, value string }` (`render.go:53-54`) → `KVPair` (exported, because
  `RenderKeyValueLines(pairs []KVPair)` takes a slice of it — external callers constructing the
  slice need the type).

### §C.2 `banner.go` (146 LOC) — moved to `uikit/banner.go`

Exported (already):
- `PrintBanner(version string)` (`banner.go:102`)
- `PrintWelcomeMessage()` (`banner.go:131`)

Unexported helpers (`resolveTheme`/`goVersion`/`claudeVersion`/`gitVersionOverride`/
`ghVersionOverride`/`goosArch` at `banner.go:41-85`): the run-phase grep decides whether any
external caller exists. If zero external callers, they stay unexported (leaf-internal helpers
that support `PrintBanner`/`PrintWelcomeMessage`).

### §C.3 `settings.go` (138 LOC) — STAYS in `package cli` (D2 STAYS reclassification)

**D2 RESOLUTION (BLOCKING iter-1 defect fixed)**: `settings.go` does NOT move to uikit.
Verified this run via `grep -nE 'type SettingsLocal' internal/ -r --include='*.go'`:

```
internal/cli/glm.go:97:type SettingsLocal struct {
```

`SettingsLocal` is defined at `glm.go:97` (a STAYS file per research.md §B.2 — glm.go is in the
kernel-USING set). `settings.go` consumes `SettingsLocal` at three sites:

```
internal/cli/settings.go:28:   func mutateSettingsLocal(path string, mutate func(*SettingsLocal)) error {
internal/cli/settings.go:48:       var settings SettingsLocal
internal/cli/settings.go:110: func stripGLMCredsAndSetTeammateMode(s *SettingsLocal) {
```

If `settings.go` moves to `package uikit`, `uikit/settings.go` references `SettingsLocal` which
is defined in `glm.go` (package cli). The result: `uikit` MUST import `package cli` to resolve
`SettingsLocal` → IMPORT CYCLE → defeats REQ-CUK-001's leaf contract.

**Option evaluation** (per plan-auditor D2 directive):

| Option | Description | Feasibility |
|--------|-------------|-------------|
| **(a) STAYS reclassification** | `settings.go` does NOT move; only `render.go`, `banner.go`, `schema_bridge.go` (or its b-ii split) move. Settings helpers remain in package cli. | **CHOSEN — see rationale below** |
| (b) Co-locate SettingsLocal into uikit | Move `SettingsLocal` type def from glm.go:97 into uikit alongside settings.go; glm.go then imports uikit for the type. | Rejected — glm.go:97's `SettingsLocal` is a 60+ field struct tied to `.claude/settings.local.json` schema (used at glm.go:513/590/940, launcher.go:251/635/662, settings.go:28/48/110); co-locating drags i18n-orthogonal settings-schema concerns into the "TUI kernel" leaf, violating the cohesion principle (AP-7 uikit-as-dumping-ground). |
| (c) Provider injection | Inject `SettingsLocal` via a provider (worktree.WorktreeProvider precedent). | Rejected — heaviest option; `mutateSettingsLocal`/`writeFileAtomic`/`stripGLMCredsAndSetTeammateMode` are file-I/O helpers tightly coupled to the on-disk schema; provider injection adds abstraction overhead with no caller-side benefit. |

**Rationale for (a)**: `settings.go`'s three exported helpers are NOT consumed outside
`package cli` (verified this run via `grep -rnE 'mutateSettingsLocal|writeFileAtomic|stripGLMCredsAndSetTeammateMode' internal/cli/*.go | grep -v _test.go` → only callers are within `internal/cli/{settings.go, launcher.go:206}`). No future cluster (migrate/doctor/update) needs these helpers
as a leaf export — they are `.claude/settings.local.json`-specific I/O routines, not part of the
TUI rendering kernel that future clusters depend on. The uikit leaf is therefore **narrowed** to
the rendering + schema-bridge kernel only (render.go + banner.go + schema_bridge.go helpers).
Source LOC: 503 → 365 (120 + 146 + 99).

**Impact on blast radius** (re-derived under option a):
- `launcher.go:206` `mutateSettingsLocal(settingsPath, stripGLMCredsAndSetTeammateMode)` call site
  does NOT rewrite — both helpers stay in package cli.
- `glm.go:513/590/940`, `launcher.go:251/635/662` `SettingsLocal` variable declarations do NOT
  rewrite — the type stays in package cli.
- The blast radius narrows by the settings-helper caller surface; corrected count in §D.9.

**Tier re-evaluation**: source LOC drops to 365 (Tier M range), BUT blast radius stays ≥13
production files + ≥10 test files = ≥23 files (well over the >15 Tier L threshold). plan-auditor
verdict: "Tier L is NOT over-scoping once the corrected blast radius is used." **Tier L stays.**

### §C.4 `schema_bridge.go` (99 LOC) — CONDITIONAL on profileSetupText resolution (b-ii split)

Unexported → exported on move:
- `schemaKeyToTUIField(schemaKey, locale string) (tuiLabel, bool)` (`schema_bridge.go:83`) →
  `SchemaKeyToTUIField`.
- `fieldDefTUILabel(f settings.FieldDef, locale string) (tuiLabel, bool)` (`schema_bridge.go:97`)
  → `FieldDefTUILabel`.

**Cross-file type dependency (CRITICAL — D5 correction: TWO maps, not one)**: `schema_bridge.go`
declares TWO `profileSetupText`-referencing map literals (verified this run):

- `schemaFieldBridge` (`schema_bridge.go:24-58`, 19 entries): every map entry signature is
  `func(t profileSetupText) tuiLabel` — every entry references the `profileSetupText` type.
- `schemaSegmentBridge` (`schema_bridge.go:60-77`, 16 entries): every map entry signature is
  `func(t profileSetupText) string` — every entry references the `profileSetupText` type.

Both maps span the line range `schema_bridge.go:24-77` (D5 iter-1 fix: the original "L24-65"
specification missed the second map). `profileSetupText` is defined at
`profile_setup_translations.go:10`; its getter `getProfileText` at `:604`. The
`profile_setup_translations.go` file is part of the FUTURE profile cluster (SPLIT-001 §C.4.4)
and STAYS in `package cli` for this SPEC.

**Two resolutions (run phase chooses; both preserve behavior, REQ-CUK-011)**:

- **(b-i) Co-locate `profileSetupText` + `getProfileText` into uikit.** Cleaner: the schema_bridge
  becomes self-contained in uikit. BUT it pulls translation data (609 LOC of i18n strings + the
  SegmentXxx fields) into the kernel leaf, which overreaches (uikit becomes TUI-kernel +
  i18n-data hybrid). The profile cluster (when it extracts) would then import `profileSetupText`
  from uikit — a reverse direction that is not the natural ownership.
- **(b-ii) Split BOTH `schemaFieldBridge` AND `schemaSegmentBridge` — keep the
  `profileSetupText`-referencing maps in package cli (D5 fix: iter-1 said "split schemaFieldBridge"
  only, missing the second map).** The profileSetupText-referencing map literals for BOTH maps
  stay in package cli (e.g. in a new `package cli` file `schema_bridge_profile.go`); only the
  leaf-able helpers (`SchemaKeyToTUIField`, `FieldDefTUILabel`) move to uikit. The helpers move
  with a signature refactor — they no longer close over `profileSetupText`; instead the caller
  looks up the bridge entry first, then passes the resolved `profileSetupText` to the returned
  closure (the closures themselves stay in package cli). This keeps uikit cohesive
  (TUI-rendering + schema-key-translation helpers only) and AVOIDS the cycle that would form if
  uikit imported profile_setup_translations.go.

**The b-ii resolution MUST cover BOTH maps** — a run-phase engineer following the iter-1
instruction literally would split only `schemaFieldBridge` and miss `schemaSegmentBridge`, which
would then either fail to compile in uikit (if the file moves as a whole) or silently re-create
the cycle (if `schemaSegmentBridge` is moved alongside `schemaFieldBridge` into uikit and both
reference `profileSetupText`).

The run-phase decision records which resolution was chosen in the M1 commit message.

### §C.5 Cross-file type dependency: `CheckStatus` (defined in doctor.go, consumed by render.go AND 6 other files)

**The hazard**: `render.go:60` `renderStatusLine(status CheckStatus, ...)` consumes the
`CheckStatus` type. But `CheckStatus` is defined at `doctor.go:25-26`:

```go
// doctor.go:25-34 (current location)
type CheckStatus string
const (
    CheckOK   CheckStatus = "ok"
    CheckWarn CheckStatus = "warn"
    CheckFail CheckStatus = "fail"
)
```

`doctor.go` is a kernel-dependent cluster file (SPLIT-001 §B doctor cluster) that STAYS in
`package cli` for this SPEC. If render.go moves to uikit carrying the `renderStatusLine` helper,
uikit must NOT import doctor.go's `CheckStatus` (cycle: uikit → cli → ... → uikit via root.go
registration of any future doctor cluster — actually the cycle doesn't form TODAY because
doctor.go is still in package cli, but it WILL form when doctor extracts to a sibling
subpackage).

**D1 iter-1 BLOCKING fix — the CheckStatus blast radius is NOT just doctor.go**: the type is
referenced by SEVEN files across the cli tree. Verified this run via
`grep -cE '\bCheckStatus\b|\bCheckOK\b|\bCheckWarn\b|\bCheckFail\b'` on each file:

| File | Ref count | Type |
|------|----------:|------|
| `internal/cli/doctor.go` | (separately verified — type def + 4 ref sites: L25-34 type def, L40 field, L152-153 fn sig, L568 fn sig) | production |
| `internal/cli/doctor_cache.go` | 5 (L49, L60, L70, L80, L88 — `check.Status = CheckOK/CheckWarn`) | production |
| `internal/cli/doctor_harness.go` | 4 (L24, L96, L100, L104 — `check.Status = CheckOK/CheckFail/CheckWarn`) | production |
| `internal/cli/mcp_doctor_coverage_test.go` | 4 | test |
| `internal/cli/coverage_test.go` | 6 | test |
| `internal/cli/coverage_fixes_test.go` | 3 | test |
| `internal/cli/coverage_improvement_test.go` | 12 | test |
| `internal/cli/doctor_golden_test.go` | 9 | test |

**Total: 43 additional CheckStatus references across 7 files** (3 production + 4 test) that
iter-1 design.md §D.8 missed entirely. Every one of these references MUST rewrite to
`uikit.CheckStatus`/`uikit.CheckOK`/`uikit.CheckWarn`/`uikit.CheckFail` when the type moves to
`uikit/types.go` (per REQ-CUK-007 / AC-CUK-007 / AC-CUK-008).

**Reverse-cycle check**: every caller of `CheckStatus` is a CONSUMER (it reads/writes the type +
its consts). None of the 7 files is a file that `uikit` itself imports. uikit imports
`internal/tui`, `internal/settings`, `lipgloss`, `fmt`, `os`, `runtime`, `strings` — none of
these is a `package cli` file. The type move therefore creates NO reverse cycle (test files
importing `uikit` is fine — test files do not belong to the production import graph).

**Resolution (clean, REQ-CUK-007)**: the `CheckStatus` type is a generic `"ok" / "warn" / "fail"`
status enum — it has no doctor-specific semantics. It BELONGS in the uikit leaf. So:

1. Move the `type CheckStatus string` + the 3 consts from `doctor.go:25-34` into
   `internal/cli/uikit/types.go` (new file).
2. Every CONSUMER of `CheckStatus`/`CheckOK`/`CheckWarn`/`CheckFail` rewrites to the
   `uikit.`-qualified form. The rewrite covers doctor.go (4 sites), doctor_cache.go (5 sites),
   doctor_harness.go (4 sites), AND the 4 test files (28 sites — test files rewrite the same way
   to keep compiling against the moved type). The struct field `Status CheckStatus` at
   `doctor.go:40` becomes `Status uikit.CheckStatus`; the function signatures
   `statusIcon(s CheckStatus)` (`doctor.go:568`) and `checkStatusToTUI(s CheckStatus)`
   (`doctor.go:152-153`) take `uikit.CheckStatus`.

This is part of the M1 caller-rewrite blast radius (§D.8) — CheckStatus-bearing files are 3
production + 4 test rewritten for the type move (NOT a doctor-cluster extraction; the doctor
cluster's heavier helpers stay in package cli).

### §C.6 Inventory summary

| Source file | LOC | Final classification | Exported helpers | Unexported→exported | Cross-file type dep |
|-------------|----:|----------------------|------------------|---------------------|---------------------|
| render.go | 120 | **MOVED to uikit** | RenderError | 8 render helpers + KVPair type | CheckStatus (§C.5) |
| banner.go | 146 | **MOVED to uikit** | PrintBanner, PrintWelcomeMessage | 0-6 (run-phase grep) | none |
| schema_bridge.go | 99 | **MOVED to uikit** (b-ii split: only `SchemaKeyToTUIField` + `FieldDefTUILabel` move; BOTH `schemaFieldBridge` AND `schemaSegmentBridge` stay in package cli per §C.4 D5 fix) | (none) | 2 bridge helpers | profileSetupText (§C.4) |
| settings.go | 138 | **STAYS in package cli** (D2 STAYS reclassification — `SettingsLocal` coupling to glm.go:97 defeats leaf contract; helpers remain package-cli-internal) | n/a (not moved) | n/a | SettingsLocal (resolved by STAYS) |
| **Total moved** | **365** | **3 source files** | **3 already-exported** | **10+ re-scoped** | **3 cross-file + 1 STAYS** |

## §D. Caller-Rewrite Blast-Radius Map (verbatim file:line, observed this run)

Every caller of a moved helper MUST rewrite to `uikit.<Helper>(...)`. The map below was
re-derived this run via `grep` (NOT assumed from SPLIT-001); it is the input to REQ-CUK-006 /
AC-CUK-008.

### §D.1 `RenderError` callers (1 site, 1 file)

```
internal/cli/migrate_agency.go:634:    fmt.Fprintln(os.Stderr, RenderError(err))
```

This is THE blocker SPLIT-001 §C.4.3(i) flagged. After uikit lands, this becomes
`fmt.Fprintln(os.Stderr, uikit.RenderError(err))` — and a future `internal/cli/migrate`
subpackage can make the same call without a cycle.

### §D.2 `renderCard` callers (5 sites, 4 files)

```
internal/cli/hook.go:292:              fmt.Fprintln(out, renderCard("Registered Hook Handlers", renderKeyValueLines(pairs)))
internal/cli/inventory.go:92:          fmt.Fprintln(w, renderCard(...))
internal/cli/inventory.go:96:          fmt.Fprintln(w, renderCard(...))
internal/cli/inventory.go:100:         fmt.Fprintln(w, renderCard(...))
internal/cli/research.go:69:           fmt.Fprintln(w, renderCard("Research Status", renderKeyValueLines(pairs)))
```

Post-move: every site becomes `uikit.RenderCard(...)`.

### §D.3 `renderKeyValue` / `renderKeyValueLines` callers (8 sites, 5 files)

```
internal/cli/hook.go:292:              ... renderKeyValueLines(pairs)
internal/cli/init_layout.go:48:        if details := renderKeyValueLines(pairs); ...
internal/cli/inventory.go:133:         return renderKeyValueLines(pairs)
internal/cli/inventory.go:148:         return renderKeyValueLines(pairs)
internal/cli/inventory.go:164:         return renderKeyValueLines(pairs)
internal/cli/init.go:435:              renderKeyValueLines([]kvPair{...})
internal/cli/research.go:69:           ... renderKeyValueLines(pairs)
```

Plus the `kvPair` literal construction sites (these become `uikit.KVPair` literals):

```
internal/cli/hook.go:275,285           var pairs []kvPair / pairs = append(pairs, kvPair{...})
internal/cli/init_layout.go:28,43      pairs := []kvPair{...} / pairs = append(pairs, kvPair{...})
internal/cli/inventory.go:126,128,141,143,156,162   pairs := make([]kvPair, ...), kvPair{...}
internal/cli/init.go:435               []kvPair{...}
internal/cli/research.go:55            pairs := []kvPair{...}
```

### §D.4 `renderSuccessCard` / `renderInfoCard` callers (6 sites, 4 files)

```
internal/cli/doctor.go:139:            fmt.Fprintln(out, renderInfoCard("Suggested Fixes", ...))
internal/cli/glm.go:299:               fmt.Fprintln(out, renderSuccessCard(...))
internal/cli/glm.go:339:               fmt.Fprintln(out, renderSuccessCard(...))
internal/cli/hook.go:269:              fmt.Fprintln(out, renderInfoCard("Registered Hook Handlers", "..."))
internal/cli/hook.go:290:              fmt.Fprintln(out, renderInfoCard("Registered Hook Handlers", "..."))
internal/cli/init.go:444:              fmt.Fprintln(cmd.OutOrStdout(), renderSuccessCard("MoAI project initialized", ...))
```

### §D.5 `renderStatusLine` / `renderSummaryLine` callers (0 non-test sites)

```
(no non-test callers — renderStatusLine is exercised only via render_test.go:54; renderSummaryLine has 0 sites in the grep)
```

These helpers still move (they are part of the render.go file; SPLIT-001 §B cluster inventory
listed them as kernel surface). Their tests move with them.

### §D.6 `PrintBanner` callers (3 sites, 3 files)

```
internal/cli/init.go:317:              PrintBanner(version.GetVersion())
internal/cli/root.go:27:               PrintBanner(version.GetVersion())
internal/cli/update.go:2538:           PrintBanner(version.GetVersion())
```

### §D.7 `PrintWelcomeMessage` callers (1 production site + 1 test file)

```
internal/cli/update.go:2543:           PrintWelcomeMessage()
internal/cli/misc_coverage_test.go:673-717   (PrintWelcomeMessage tests — move to uikit, OR rewrite to call uikit.PrintWelcomeMessage)
```

### §D.8 `CheckStatus` type references (3 production + 5 test files — type moves to uikit/types.go)

> **D1 iter-1 BLOCKING fix**: the original §D.8 listed only `doctor.go` + `render_test.go:54`.
> The verbatim re-grep this run (per SPLIT-001 M1-blocker lesson: "cluster characterization
> without verbatim file:line verification is unsound") surfaces 43 additional `CheckStatus` /
> `CheckOK` / `CheckWarn` / `CheckFail` references across 7 files. Every reference rewrites to
> the `uikit.`-qualified form when the type moves to `uikit/types.go`.

**Production files** (3 — type-def + 2 production consumers iter-1 missed):

```
internal/cli/doctor.go:25-34         type CheckStatus + 3 consts → MOVE to uikit/types.go
internal/cli/doctor.go:40            Status CheckStatus → Status uikit.CheckStatus
internal/cli/doctor.go:152-153       checkStatusToTUI(s CheckStatus) → (s uikit.CheckStatus)
internal/cli/doctor.go:568           statusIcon(s CheckStatus) → (s uikit.CheckStatus)
internal/cli/doctor_cache.go:49      check.Status = CheckOK → check.Status = uikit.CheckOK
internal/cli/doctor_cache.go:60      check.Status = CheckOK → check.Status = uikit.CheckOK
internal/cli/doctor_cache.go:70      check.Status = CheckOK → check.Status = uikit.CheckOK
internal/cli/doctor_cache.go:80      check.Status = CheckWarn → check.Status = uikit.CheckWarn
internal/cli/doctor_cache.go:88      check.Status = CheckOK → check.Status = uikit.CheckOK
internal/cli/doctor_harness.go:24    check.Status = CheckOK → check.Status = uikit.CheckOK
internal/cli/doctor_harness.go:96    check.Status = CheckFail → check.Status = uikit.CheckFail
internal/cli/doctor_harness.go:100   check.Status = CheckWarn → check.Status = uikit.CheckWarn
internal/cli/doctor_harness.go:104   check.Status = CheckOK → check.Status = uikit.CheckOK
```

**Test files** (5 — every test file referencing CheckStatus rewrites the same way to keep
compiling against the moved type):

```
internal/cli/mcp_doctor_coverage_test.go    4 refs → rewrite to uikit.* forms
internal/cli/coverage_test.go               6 refs → rewrite to uikit.* forms
internal/cli/coverage_fixes_test.go         3 refs → rewrite to uikit.* forms
internal/cli/coverage_improvement_test.go  12 refs → rewrite to uikit.* forms
internal/cli/doctor_golden_test.go          9 refs → rewrite to uikit.* forms
internal/cli/render_test.go:54              renderStatusLine(CheckOK, ...) → (uikit.CheckOK, ...) [test moves with render.go]
```

**Total for CheckStatus rewrite**: 3 production files (doctor.go + doctor_cache.go +
doctor_harness.go) + 6 test files (the 5 CheckStatus-bearing test files + render_test.go).

### §D.9 Blast-radius summary (D3 + D4 corrected)

| File | Rewrites | Reason | Iter-1 status |
|------|---------:|--------|---------------|
| migrate_agency.go | 1 | RenderError | OK |
| hook.go | 5 | renderCard, renderInfoCard ×2, renderKeyValueLines, kvPair | OK |
| inventory.go | 9 | renderCard ×3, renderKeyValueLines ×3, kvPair ×6 (some overlap) | OK |
| research.go | 2 | renderCard, renderKeyValueLines, kvPair | OK |
| init.go | 4 | renderKeyValueLines, kvPair, renderSuccessCard, PrintBanner | OK |
| init_layout.go | 3 | renderKeyValueLines, kvPair ×2 | OK |
| doctor.go | 4+ | CheckStatus type def + 4 ref sites (type moves OUT of doctor.go) | OK |
| **doctor_cache.go** | 5 | CheckStatus refs at L49/60/70/80/88 | **D1 iter-1 fix — newly added** |
| **doctor_harness.go** | 4 | CheckStatus refs at L24/96/100/104 | **D1 iter-1 fix — newly added** |
| root.go | 1 | PrintBanner | OK |
| update.go | 2 | PrintBanner, PrintWelcomeMessage | OK |
| glm.go | 2 | renderSuccessCard ×2 | OK |

**Production files re-derived under D2 option (a) (settings.go STAYS)**:
- launcher.go:206 was listed as a "missed production file" by the iter-1 plan-auditor, BUT it
  calls `mutateSettingsLocal` + `stripGLMCredsAndSetTeammateMode` — both helpers STAY in package
  cli under D2 option (a). The call site therefore does NOT rewrite. launcher.go is REMOVED from
  the blast radius under D2-a (corrected from the iter-1 plan-auditor's "13 production files"
  count, which assumed settings.go moves to uikit — that assumption is reversed by D2-a).
- profile_setup_translations.go:141 carries only a comment reference to a moved helper (not a
  call site); excluded from the rewrite count.

**Corrected production-file count (D2-a applied)**: 12 production files rewritten
(migrate_agency.go + hook.go + inventory.go + research.go + init.go + init_layout.go + doctor.go
+ doctor_cache.go + doctor_harness.go + root.go + update.go + glm.go). The original iter-1 count
of "10 production files" was an undercount (D1 + D3 added doctor_cache.go + doctor_harness.go).
The plan-auditor's alternative "13 production files" assumed settings.go moves; under D2-a
launcher.go drops out, netting 12.

**Tier L re-justification**: 12 production files + (see §D.10) ≥10 test files + 3 source files
moved + 1 new types.go = ≥26 files total. The >15 Tier L threshold is comfortably met on the
corrected blast radius even with source LOC at 365. Tier L stays.

### §D.10 Test-file blast radius (D4 iter-1 fix — ≥10, not "≥1")

> **D4 iter-1 SHOULD-FIX**: the iter-1 §D.9 summary listed only `render_test.go` +
> `misc_coverage_test.go`. The verbatim re-grep this run surfaces **19 test files** referencing
> at least one moved helper or the CheckStatus type. The run phase re-verifies each before the
> M1 commit (AC-CUK-008 / AC-CUK-012).

Verified via
`grep -lrnE 'renderCard|renderKeyValue|renderStatusLine|renderSuccessCard|renderInfoCard|renderSummaryLine|RenderError|PrintBanner|PrintWelcomeMessage|printWelcomeMessage|mutateSettingsLocal|writeFileAtomic|schemaKeyToTUIField|kvPair|CheckStatus|CheckOK|CheckWarn|CheckFail|profileSetupText|stripGLMCredsAndSetTeammateMode' internal/cli/*_test.go | sort -u`:

| # | Test file | Reason | Action |
|---|-----------|--------|--------|
| 1 | `banner_test.go` | PrintBanner / PrintWelcomeMessage | rewrite to `uikit.PrintBanner` / `uikit.PrintWelcomeMessage` |
| 2 | `cg_mode_hardening_test.go` | (settings helper or render helper refs — verify at run-phase) | rewrite or keep (run-phase decision per re-grep) |
| 3 | `coverage_fixes_test.go` | CheckStatus refs (3) | rewrite to `uikit.CheckStatus` / consts |
| 4 | `coverage_improvement_test.go` | CheckStatus refs (12) | rewrite to `uikit.CheckStatus` / consts |
| 5 | `coverage_test.go` | CheckStatus refs (6) | rewrite to `uikit.CheckStatus` / consts |
| 6 | `doctor_cache_test.go` | exercises doctor_cache.go which references CheckStatus | verify at run-phase (likely no direct CheckStatus ref — re-grep decides) |
| 7 | `doctor_constitution_test.go` | (doctor-cluster test — verify at run-phase) | rewrite or keep (run-phase decision) |
| 8 | `doctor_golden_test.go` | CheckStatus refs (9) | rewrite to `uikit.CheckStatus` / consts |
| 9 | `doctor_harness_test.go` | (exercises doctor_harness.go) | verify at run-phase |
| 10 | `doctor_new_test.go` | (doctor-cluster test — verify at run-phase) | rewrite or keep (run-phase decision) |
| 11 | `doctor_test.go` | (doctor-cluster test — verify at run-phase) | rewrite or keep (run-phase decision) |
| 12 | `mcp_doctor_coverage_test.go` | CheckStatus refs (4) | rewrite to `uikit.CheckStatus` / consts |
| 13 | `misc_coverage_test.go` | PrintWelcomeMessage test block (L673-717) | **MOVE** the PrintWelcomeMessage block to uikit (rewrite to `package uikit` / `package uikit_test`); rest of file stays |
| 14 | `profile_setup_model_policy_test.go` | (profile cluster test — verify at run-phase) | likely no rewrite (profile cluster STAYS) |
| 15 | `profile_setup_translations_test.go` | (profile cluster test — verify at run-phase) | likely no rewrite (profile cluster STAYS) |
| 16 | `remaining_coverage_test.go` | (coverage test — verify at run-phase) | rewrite or keep |
| 17 | `render_test.go` | white-box test of render helpers | **MOVE** to `package uikit` (becomes black-box `package uikit_test` or stays `package uikit` per run-phase decision) |
| 18 | `schema_bridge_test.go` | schemaKeyToTUIField / fieldDefTUILabel | rewrite to `uikit.SchemaKeyToTUIField` / `uikit.FieldDefTUILabel`; if the test closes over `profileSetupText`, the b-ii split decision governs whether it moves |
| 19 | `target_coverage_test.go` | (coverage test — verify at run-phase) | rewrite or keep |

**Test-file count**: ≥10 test files affected (the CheckStatus-bearing 4 — coverage_test,
coverage_fixes_test, coverage_improvement_test, mcp_doctor_coverage_test — plus doctor_golden_test
plus banner_test plus render_test (move) plus misc_coverage_test (move-block) plus
schema_bridge_test; the remaining test files are run-phase-re-verified). The iter-1 "≥1"
characterization was a 10× undercount.

**misc_coverage_test.go PrintWelcomeMessage resolution (D4 plan-phase decision)**: the
PrintWelcomeMessage test block (`misc_coverage_test.go:673-717`) **MOVES** to uikit (becomes
part of the moved test surface alongside render_test.go). The rest of misc_coverage_test.go
stays in package cli. This resolves the iter-1 "(rewrite or move)" ambiguity.

## §E. Anti-Patterns (cross-reference SPLIT-001 design.md §G)

The SPLIT-001 §G anti-pattern catalogue carries forward. The most relevant for uikit:

- **AP-2 — Exporting kernel helpers from `package cli` and importing back**: creates an import
  cycle (root imports uikit if uikit is a registered command, OR package cli imports uikit and
  uikit imports package cli back). The uikit leaf contract (REQ-CUK-001) is the structural
  prevention: uikit imports NOTHING from `internal/cli/*`. Verified by AC-CUK-005.
- **AP-3 — Moving a source file without its `_test.go` files**: leaves white-box tests
  referencing moved-away unexported symbols → compile break. Tests move WITH the source
  (render_test.go moves with render.go; misc_coverage_test.go's PrintWelcomeMessage block
  moves or rewrites). The run phase verifies test count before == after (AC-CUK-012).
- **AP-5 — Adding behavior tests during the move**: this is a pure refactor; new tests conflate
  refactor risk with feature risk. Characterization only (SPLIT-001 §E).
- **AP-7 (uikit-specific) — Using uikit as a dumping ground**: the leaf stays cohesive
  (TUI rendering + settings mutation + schema-bridge helpers). Unrelated helpers belong in
  their own leaf (e.g. `archiveutil` for the migrate/update shared helpers — SPLIT-001 axis-ii).
  Do NOT add non-kernel helpers to uikit just because "they need a home."

## §F. Cross-References

- `.moai/specs/SPEC-CLI-SUBPKG-SPLIT-001/design.md` §A — target layout (uikit leaf entry).
- `.moai/specs/SPEC-CLI-SUBPKG-SPLIT-001/design.md` §B — 7-step extraction recipe.
- `.moai/specs/SPEC-CLI-SUBPKG-SPLIT-001/design.md` §D — import-cycle resolution (THIS SPEC).
- `.moai/specs/SPEC-CLI-SUBPKG-SPLIT-001/design.md` §G — anti-pattern catalogue (AP-1..AP-6).
- `internal/cli/CLAUDE.md` — cobra registration, cross-platform, subagent boundary conventions.
- research.md §B — 14-file kernel survey classification (MOVED / STAYS / GATED).
