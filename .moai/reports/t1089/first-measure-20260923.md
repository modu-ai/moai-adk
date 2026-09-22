# t1089 first measurements — 2026-09-23 (tree 17f71a13d, WT-opus-55-default)

## 1. zone-registry.md opus-5 hits
`grep -n -i 'opus-5\|opus 5' .claude/rules/moai/core/zone-registry.md` → lines 311, 319 (template copy identical).
Both are `anchor: "#opus-5-48-prompt-philosophy"` fields, not clause text:
- CONST-V3R2-028 — zone: Evolvable, zone_class: evolvable-tuning, canary_gate: false (Principle 4)
- CONST-V3R2-029 — zone: Evolvable, zone_class: evolvable-tuning, canary_gate: false (Principle 5)
No Frozen row hit → constitution amend gate NOT required. Caveat: renaming the
constitution heading "Opus 5 / 4.8 Prompt Philosophy" changes the anchor, so both
anchor values (local + template) must move with it.

## 2. moai web model/effort edit path
Exists already. `internal/web/fieldsets.templ:128-129` fieldsetLaunch renders
optSelect "model" + "effort_level"; schema `internal/settings/schema.go:358-378`,
both Persist PersistProfileStore; validation `internal/web/validate.go:154`.
- effort options `internal/settings/schema.go:231-234`: low/medium/high/xhigh/max (already the 5 required)
- effort empty label = runtime default ("opt.runtime_default"); no "medium" default shown
- model options = `template.ModelAliasPickerValues()` (`internal/template/model_policy.go:139`): fable[1m], opus[1m], sonnet[1m], haiku — aliases only, no full IDs
Gap for scope ④ is therefore default display/recommendation, not a missing edit path.

## 3. settings.json.tmpl effortLevel
`grep -n -i effort internal/template/templates/.claude/settings.json.tmpl` → no match (rc=1). No effortLevel key in the template.

## 4. model_policy.go
`ModelIDOpus5 = "claude-opus-5"` (:51), alias table `"opus": ModelIDOpus5` (:78), legacy map carries 4-6/4-7/4-8 → "opus".

## Base drift
Local develop moved to 783b74455 (t1088 merge) after dispatch; this tree stays on dispatched base 17f71a13d.
