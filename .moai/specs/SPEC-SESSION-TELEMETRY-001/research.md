# SPEC-SESSION-TELEMETRY-001 — Research

The measurement record. Every row was produced by the command named, in this run, in worktree
`.claude/worktrees/t207` at `dfbf828a6`. Nothing here is carried over from the parent SPEC's
citations or from the split-design report — where a figure agrees with one of those, it agrees
because it was re-run, not because it was copied.

**Note on the base commit.** The split-design report measured at `52358e72f`. This tree's `HEAD`
is `dfbf828a6`; the intervening commits are report and SPEC-document commits on this card's
branch, touching no file this SPEC measures. Every measurement below was nonetheless re-run at
`dfbf828a6` rather than inherited.

## §1 The write path and the record

```
$ grep -n 'contextUsageRecord\|context-usage\|stateDir' internal/statusline/context_usage.go
27:  // contextUsageSchemaVersion is the on-disk schema version of context-usage.json.
47:  // contextUsageRecord is the on-disk schema for
48:  // <projectDir>/.moai/state/context-usage.json (REQ-THRESHOLD-010).
56:  type contextUsageRecord struct {
133:      stateDir := filepath.Join(projDir, ".moai", "state")
134:      path := filepath.Join(stateDir, "context-usage.json")
186: func readContextUsage(path string) (*contextUsageRecord, error) {
203: func sameSemanticPayload(a, b *contextUsageRecord) bool
216: func isRealSessionID(s string) bool
236: func isFreshForSession(rec *contextUsageRecord, curSession string, curWriterID int) bool
```

One file per project root; the session identity lives inside the payload, not in the path.

**Correction to a cited line number.** The split design placed the persistence call at
`builder.go:157`. Measured:

```
$ grep -n 'writeContextUsage(' internal/statusline/builder.go
168:	writeContextUsage(resolveProjectDir(input), sessionID, os.Getpid(), data.Memory, handoffGuideStage(data))
```

`:157` is inside the explanatory comment block that precedes the call. The call itself is at
`:168`. This SPEC cites `:168`.

## §2 The race, observed rather than cited

```
$ python3 … .moai/state/active-sessions.json        (2026-08-24)
2beac221…  pid 15207  …/worktrees/t219
c15d8434…  pid 51045  …/worktrees/t210
3db058e1…  pid 36912  …/worktrees/t207

$ cat .moai/state/context-usage.json                 (13:08)
{"schema_version":1,"session_id":"d281730e-…","writer_pid":71763,
 "captured_at":"2026-08-24T13:08:03+09:00","tokens_used":560000,"raw_pct":56,…}

$ ls -d .moai/state/context-usage
ls: .moai/state/context-usage: No such file or directory
```

Three live sessions registered; the single slot holds a fourth. Re-read eight minutes later the
same slot held `writer_pid 12889`, `raw_pct 57` — the same session churning the one file. None
of the three registered sessions' usage is readable from it, at any moment.

## §3 The launcher cannot supply model or effort

```
$ grep -rn '"-m"\|"--model"\|Model' internal/cli/cc.go | wc -l
0

$ grep -n "ANTHROPIC_DEFAULT\|--model" internal/cli/cc.go
36:  -m, --model <model>           Override model selection      ← help string only

$ sed -n '350,353p' internal/cli/glm.go
  os.Setenv(config.EnvAnthropicDefaultOpusModel,   glmConfig.Models.High)
  os.Setenv(config.EnvAnthropicDefaultSonnetModel, glmConfig.Models.Medium)
  os.Setenv(config.EnvAnthropicDefaultHaikuModel,  glmConfig.Models.Low)
  os.Setenv(config.EnvAnthropicDefaultFableModel,  glmConfig.Models.Fable)

$ sed -n '70,73p;92p' internal/config/profile.go
  // ModelEffort carries a {model, effort} assignment … Model is a Claude Code
  // short alias (opus/sonnet/fable/inherit) …
  func (l LLMConfig) EffectiveProfile() string {      ← returns high/medium/low
```

`moai cc` parses no model. `moai glm` populates four slots and cannot know which one the session
will run in. `EffectiveProfile` returns a profile name in a different vocabulary from a model
name. The parent SPEC's REQ-WC15-011 **as it stood in that SPEC's version 0.1.0** was
unimplementable on both backends — the finding the split acted on. The requirement no longer
exists: `SPEC-WEB-CONSOLE-015` 0.2.0 removed it rather than reworded it, so the identifier
resolves to nothing in the current tree and is cited here as history, not as a live reference.

## §4 The statusline does hold all three values

```
$ grep -n 'Effort \|EffortInfo\|DisplayName' internal/statusline/types.go
69:   Effort *EffortInfo `json:"effort"`   // Claude Code v2.1.139+ (nil if absent)
131:  type EffortInfo struct { Level string `json:"level"` }
148:  DisplayName string `json:"display_name"`

$ grep -n 'data.Effort\|input.Effort' internal/statusline/builder.go
286:  if input != nil && input.Effort != nil {
287:      data.Effort = input.Effort

$ grep -n 'resolveGLMModelName' internal/statusline/metrics.go
51:   modelName = resolveGLMModelName(modelName)
197:  func resolveGLMModelName(displayName string) string
```

`resolveGLMModelName`'s own comment states it strips a `[1m]` suffix and substitutes the actual
model when `ANTHROPIC_DEFAULT_*_MODEL` names a non-Claude one — the GLM half of D-5 already
exists and is already on the render path.

### Runtime payload, observed

Captured 2026-08-24 13:02, Claude Code 2.1.241, session `d281730e-a47e-4f82-878e-5fd0ddc4dcb9`,
by temporarily copying the statusline wrapper's stdin to a scratch file and restoring the
wrapper afterwards (both the worktree and primary copies were backed up, patched, and restored;
`diff -q` confirmed both restorations):

```json
{ "session_id": "d281730e-…",
  "effort":  { "level": "medium" },
  "model":   { "id": "claude-opus-5[1m]", "display_name": "Opus 5 (1M context)" },
  "version": "2.1.241",
  "context_window": { "used_percentage": 54, "context_window_size": 1000000 } }
```

Effort is genuinely delivered — which is why REQ-ST-003 says the record *carries* model and
effort rather than *carries them when available*. Both `id` and `display_name` arrive, which is
the fork D-5 resolves.

## §5 The second reader, and why it breaks silently

```
$ grep -n 'tokensContextSnapshotFilename\|type tokensContextSnapshot\|func readTokensContextSnapshot\|context-usage' internal/cli/tokens.go
28:  // tokensContextSnapshotFilename is the statusline-persisted context snapshot
30:  const tokensContextSnapshotFilename = "context-usage.json"
79:  // tokensContextSnapshot is the subset of the statusline context-usage.json
81:  type tokensContextSnapshot struct {
393: func readTokensContextSnapshot(stateDir string) *tokensContextSnapshot {
394:     data, err := os.ReadFile(filepath.Join(stateDir, tokensContextSnapshotFilename))
395:     if err != nil { return nil }
426:  "… embed the context-usage snapshot when present …"   ← command help string
```

Its own filename constant, its own copy of the schema, and a `nil` return on any read error. A
path move it does not know about removes the block from `moai tokens` output with no error of
any kind.

**Correction to a cited line number.** `:79` is the struct's doc comment; the declaration is at
`:81`. The parent SPEC and its acceptance document both cited `:79` as the declaration — the
defect the second audit recorded as F10. This SPEC cites `:81`, and names `:79` as the comment.

## §6 Baselines for the absence-criteria

```
$ grep -rn '"raw_pct"' internal/
internal/statusline/context_usage.go:63
internal/statusline/context_usage_test.go:150
internal/cli/tokens.go:86
internal/cli/tokens_test.go:283

$ grep -rln '"raw_pct"' internal/ | grep -v '^internal/statusline' | wc -l
2

$ grep -rEn '^func Read[A-Za-z]*ContextUsage' internal/statusline/*.go | wc -l
0

$ grep -rln "state/context-usage.json" .claude internal/template/templates
.claude/rules/moai/workflow/context-window-management.md
.claude/rules/moai/workflow/context-window-management-detail.md
internal/template/templates/.claude/rules/moai/workflow/context-window-management-detail.md
internal/template/templates/.claude/rules/moai/workflow/context-window-management.md

$ diff -q  <each mirror pair>
(no output, exit 0)  ×2

$ grep -rln "context-usage.json" docs-site/content | wc -l
12
$ grep -rln "context-usage.json" docs-site/content
content/{en,ja,ko,zh}/advanced/statusline.md
content/{en,ja,ko,zh}/advanced/token-budget.md
content/{en,ja,ko,zh}/cli-reference/tokens.md

$ grep -rn "context-usage" .moai/README.md internal/template/templates/.moai/README.md
:31  | `state/` | Runtime state snapshots, e.g. context-usage (gitignored — regenerated) |
```

Four hits of `"raw_pct"` in four files, two of them outside `internal/statusline`; zero exported
readers; four doctrine files, both mirror pairs currently identical; twelve docs-site pages,
exactly three per locale; and two README rows that name no filename and therefore need no change.

## §7 What the doctrine actually says

`context-window-management-detail.md` §1-§2 carries the read procedure the split invalidates:
the file is named as *the* authoritative snapshot at `<projectDir>/.moai/state/context-usage.json`,
its field list includes `session_id` / `writer_pid` / `captured_at` as "validity-guard inputs",
and §2 spells out the single-slot guard verbatim — *"valid only when the record's `session_id`
equals the current session id (last-writer-wins)"* and *"the `writer_pid` discriminator
distinguishes them"*. Both guards become unreachable once the path carries the session id, which
is what REQ-ST-008 requires be dropped rather than left as decoration.

## §8 Gaps — what was NOT observed

1. **No code was changed and nothing was compiled.** Every claim here is a read of the tree as it
   stands.
2. **The write-throttle interaction is unmeasured.** Whether adding model and effort to
   `sameSemanticPayload`'s comparison degrades the throttle into a write-per-render was not
   tested. `plan.md` §F item 1.
3. **No GLM-backed session was observed.** The runtime payload capture covers the Claude backend
   only; whether a GLM session's payload carries `effort` is unknown. `plan.md` §F item 2.
4. **The 23-file count is an enumeration, not a diff.** No implementation exists to measure a
   real diff against.
5. **The docs-site pages were counted, not read.** Twelve files mention the string; what each one
   says about it, and therefore how much of each page M6 must rewrite, was not examined.
6. **`moai tokens` was not run.** The silent-break mechanism is established from the source
   (`nil` on read error, own filename constant), not from an observed disappearance.

## §9 Residual risk

- **The record's name will no longer describe its contents.** Accepted deliberately (D-1) to
  avoid moving two things in one sweep, but it is a readability debt someone will meet later.
- **The doctrine is always-loaded.** An error in M5's edit reaches every session of every project
  that takes the template, which is why the mirror-pair `diff -q` is an acceptance criterion
  rather than a review note.
- **Twelve published pages in four locales** is exactly the shape of change that ships with one
  locale left behind. AC-ST-011's third half (three per locale) exists for that and nothing else.
