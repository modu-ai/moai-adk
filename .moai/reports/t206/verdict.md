# t206 — audit enum option descriptions

Card: t206 (re-scoped) · Branch: `WT-i18n-enum-opts` · Worktree: `.claude/worktrees/t206`
Base: `origin/main` `f7d4b7824`

## Claim

The audit tab's enum options now carry an always-visible per-option description
in all four locales, while the English enum **labels** are unchanged — the G1-2
decision is upheld, not reversed. The card's original premise (a broken render
path printing raw enum values) was false and was rejected first; this entry
records what the re-scoped card actually delivered.

### Surface chosen, and why

The lead named three candidate surfaces (field `.desc` extension / inline text /
tooltip) and delegated the choice. **None was used.** A fourth surface already
exists and is purpose-built for exactly this: `OptionDef.OptionDesc`.

Setting it is documented as "the sole opt-in" for the stacked radio layout —
`schemaRadioRow` branches on its presence and renders
`<span class="seg__d" data-i18n={opt.OptionDesc}>` under each label
(`internal/web/fieldsets.templ:434-447`). Reasons it beats all three:

- **Always visible**, unlike a tooltip. The operator's complaint was not knowing
  what the options mean; a hover-only affordance they must discover first does
  not reliably answer that, and does not exist at all on touch input.
- **Per-option**, unlike extending the field-level `.desc` — one paragraph
  cannot attach a sentence to each of four values.
- **Zero new render or CSS code.** The layout, the `seg--stacked` class, and the
  `seg__d` style all already exist and already ship, with `report.format` as the
  precedent user.

### The constraint that decided the key shape

The lead's [HARD] instruction — new keys must not contain `.opt.` — is
load-bearing, and the existing precedent violates it. `report.format` names its
per-option keys `f.report.format.opt.html_md.desc`, which **does** contain
`.opt.`, so `applyI18n` resolves them against the English dictionary and their
ko/ja/zh translations never render. Copying that convention would have shipped a
feature that silently does nothing outside English.

The keys therefore use `.option.` — `f.workflow.audit.gate.option.<value>` and
`f.workflow.audit.model.option.<value>`. `.opt.` is not a substring of
`.option.` because the character after `opt` is `i`. A test asserts this rather
than leaving it to a reader to re-derive.

The three gate fields (claude/codex/glm) share one description set: the values
mean the same thing whichever backend the gate fronts, so three copies would be
three things to keep in sync.

## Evidence

### RED — the failure was observed before the fix

```
$ go test ./internal/web/ -run 'TestAuditOption|TestAuditRadio'
--- FAIL: TestAuditOptionsCarryOptionDesc (0.00s)
    field "workflow.audit.model" option "claude" has no OptionDesc — the option renders with no explanation
    field "workflow.audit.model" option "codex" has no OptionDesc …
    field "workflow.audit.model" option "glm" has no OptionDesc …
    field "workflow.audit.model" option "multi" has no OptionDesc …
    field "workflow.audit.gates.claude" option "off" / "advisory" / "required" has no OptionDesc …
    field "workflow.audit.gates.codex"  option "off" / "advisory" / "required" has no OptionDesc …
    field "workflow.audit.gates.glm"    option "off" / "advisory" / "required" has no OptionDesc …
    (13 failures)
--- FAIL: TestAuditOptionDescTranslatedInAllLocales (0.00s)
    i18n.js declares key "f.workflow.audit.model.option.claude" 0 times, want 4 (en/ko/ja/zh)
    … (7 keys, all 0 times)
```

The first RED pass had two tests passing **vacuously** — they skipped whenever
`OptionDesc` was empty, which was the very condition under test. They were
rewritten to derive the expected key from `prefix + opt.Value` instead, and only
the output above (after that rewrite) is quoted as the baseline.

The dictionary baseline is a clean **0 occurrences**, console-wide: before this
change no `.option.` key existed anywhere in `i18n.js`, so the per-option
tooltip seam `f.Description + ".option." + opt.Value` in `fieldsets.templ:409/
411/467` had never had a single key behind it. That makes the assertion a real
observation rather than a re-grep of text that already existed.

### GREEN

```
$ go test ./internal/web/ -run 'TestAuditOption|TestAuditRadio|TestOptionLabelsStayEnglish'
ok  	github.com/modu-ai/moai-adk/internal/web	1.907s

$ go test ./internal/web/... ./internal/settings/... -timeout 600s
ok  	github.com/modu-ai/moai-adk/internal/web            7.795s
ok  	github.com/modu-ai/moai-adk/internal/settings       0.809s
ok  	github.com/modu-ai/moai-adk/internal/settings/agentfm    1.681s
ok  	github.com/modu-ai/moai-adk/internal/settings/yamlpatch  1.262s

$ go vet ./internal/web/... ./internal/settings/...
(no output — clean)

$ grep -c 'audit\.\(gate\|model\)\.option\.' internal/web/assets/i18n.js
28                                    # 7 keys × 4 locales
```

`gofmt -l` flags `internal/settings/tier_test.go` and
`internal/web/viewmodel_ops.go`. Neither appears in `git diff --name-only`
(which lists only `internal/settings/schema_sections.go` and
`internal/web/assets/i18n.js`) — both were already unformatted on the base
commit and are left alone under scope discipline.

### Guard upheld

`TestOptionLabelsStayEnglish` (G1-2) is untouched and still passes.
`TestOptionLabelsStayEnglishStillGuarded` was added as a tripwire so a later
change cannot quietly reverse the decision this card declined to reverse.

## Baseline-attribution

Every figure was measured in this run, in this tree
(`.claude/worktrees/t206`), against base `origin/main` `f7d4b7824`. The RED
output is from this tree before the two source edits; the GREEN output is from
the same tree after them. Nothing is carried over from another package, tree, or
point in time.

## Gaps

- **The full suite was not run**, per the standing rule and the lead's [HARD]
  instruction. Only `internal/web/...` and `internal/settings/...` ran; the
  full-package verdict is CI's on the pushed head.
- **No browser render was performed.** The render lock is
  `TestAuditRadioRendersStackedOptionDesc`, which asserts the
  `data-i18n="…"` attributes are present in the served HTML. That the browser
  then paints them legibly under `seg--stacked` is inferred from the existing
  `report.format` precedent, not observed here.
- **The ko/ja/zh copy has had no second reader.** I wrote it; no native reviewer
  checked it.
- **The `seg--stacked` assertion is weak on its own** — `report.format` already
  renders that class, so it would pass even if the audit fields did not opt in.
  The per-key `data-i18n` assertions carry the real weight.
- **Only the audit fields were changed.** The card's original sweep list
  (`execution_mode`, `default_mode`, `harness.default_profile`,
  `harness.mode_defaults`) was NOT touched — those were named as further
  instances of the misdiagnosed "untranslated label" symptom, which the reject
  established is intended behaviour. They have no per-option descriptions
  either, but adding them is outside the re-scoped card.

## Residual-risk

- **The `report.format` precedent is still broken and was left as found.** Its
  per-option description keys carry `.opt.`, so its ko/ja/zh translations are
  resolved against the English dictionary and never appear. This is a live
  defect in a shipped field, discovered while reading the precedent, and it is
  out of this card's scope — it warrants its own card.
- The dangling `f.Description + ".option." + opt.Value` tooltip seam at
  `fieldsets.templ:409/411/467` remains dangling for every other field: those
  fields set no `Description`, so the emitted attribute is `.option.<value>`, a
  key in no dictionary. Harmless (the tooltip is simply absent) but it is a
  wired seam with nothing behind it.
- `withOptionDesc` mutates `f.Options` in place. Every current caller passes a
  freshly constructed `FieldDef`, so no sharing exists today; a future caller
  that passed a shared slice would see aliasing.
