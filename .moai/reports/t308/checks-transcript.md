# SPEC-CLOCAL-AUDIT-001 — Checks Transcript

Consolidated log for all mechanical checks behind `.moai/reports/t308/claims-inventory.md`.
Protocol: one numbered entry per executed check, appended by run-phase in execution order:

```
### CHK-NNN — <INV-NNN> <one-line description>
$ <command>
<verbatim output>
exit: <code>
```

Rules (REQ-CLOCAL-002):
- A DEFECT-CONFIRMED verdict without a CHK entry is a misclassification (AC-CLOCAL-001).
- Long outputs: redirect to file under this directory + bounded tail; exit code REQUIRED.
- Serial ordering required for dependent checks only.

## Plan-phase baseline entries (recorded 2026-08-27 by manager-spec)

### CHK-000g — run-phase basis equivalence (develop advance d29b8942e → 8d8da0b2b)
$ git rev-parse --short HEAD && git branch --show-current && git diff --stat d29b8942e..8d8da0b2b -- CLAUDE.local.md
8d8da0b2b
WT-clocal-audit
(diff-stat output empty — subject byte-identical across the develop advance)
exit: 0
Note: every SHA-pinned basis reference remains valid; all subject-file claims are measured
against the live worktree copy which equals the frozen d29b8942e copy for CLAUDE.local.md.
Reality-tree citations remain measured against THIS tree @ 8d8da0b2b (named per finding).

## Run-phase entries (2026-08-27, manager-develop, basis CHK-000g)

### CHK-000a — basis pinning (original plan-phase)
$ git rev-parse HEAD && git branch --show-current
d29b8942e5b237ef7180749dc04273d83647c745
WT-clocal-audit
exit: 0

### CHK-000b — subject copy integrity
$ wc -l CLAUDE.local.md && git diff --stat HEAD -- CLAUDE.local.md
     663 CLAUDE.local.md
(no diff-stat lines beyond the count line — clean vs HEAD)
exit: 0

### CHK-000c — SPEC ID pre-write self-check
$ ID="SPEC-CLOCAL-AUDIT-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL
PASS
exit: 0

### CHK-000d — duplicate-ID scan
$ ls .moai/specs/ | grep -i CLOCAL          → no matches (exit 1)
$ grep -rn "CLOCAL" .moai/specs --include=spec.md -l → no matches
exit: 0 (absence established over .moai/specs/** scope)

### CHK-000e — plan-audit iter-1 D2 anchor re-derivation sweep @ d29b8942e
Commands executed (worktree root, subject CLAUDE.local.md):
1. `awk 'NR>=92 && NR<=144 {printf "%d:%s\n", NR, $0}' CLAUDE.local.md` — Local-Only block.
   Result: Local-Only heading L114; fence L115; settings.local.json L116; settings.json L117;
   agent-memory L118; handle-*.sh L119; lifecycle-sync-gate L120; repo-local-pr-policy L121;
   commands/harness trio L122; manifest.json L123; hns-release-update-run.js L124;
   specialist triple L125; scripts/ci-watch L126; scripts/ci-autofix L127;
   skills/hns-workflow-ci-loop L128; ci-watch-protocol L129; ci-autofix-protocol L130;
   last-cc-version.json L132; research/cc-update-*.md L133; runtime dirs L134–140;
   status_line.sh L141 exact; astgrep-rules entry L142; closing fence L143.
2. `grep -n "^## \|^### \|^#### \|^- \*\*§" CLAUDE.local.md` — full heading + References-row map.
   Result: §18 row L636; §19 row L637; §20 row L638; §21 row L639; §22 row L640;
   §23 row L641; §26 local-linear row L644; §27 row L645; §28 headings L649/653/657/661;
   §12 heading L497; §13 heading L513; exclusion zone boundaries: `### §4.1 …` L273,
   last §4.1.4 subsection L328, `## 5.` L338.
3. `awk 'NR>=481 && NR<=530 {…}'` — §11 close region. Result: twin `---` at L493 and L495
   (blank L494 between); §12 rule-pointer pair at L499; loadGLMKey() claim L521;
   inward cross-ref "§6 검증 규율(clear → 측정)" inside L487.
4. `awk 'NR>=228 && NR<=241 {…}'` and `awk 'NR>=412 && NR<=436 {…}'` and
   `awk 'NR>=170 && NR<=190 {…}'`. Results: §3 cluster L230–238 (coding-standards pointer
   L232); TemplateContext struct L417–420, constructor L431–434, Deploy call L435,
   status_line export snippet L424–427 (export line L425–426); git-strategy recipe block
   L177–186 (recipe bullet L177, verify L181, restore L183) — INV-112 anchor confirmed exact.
Dispositions applied to claims-inventory.md: ~18 stale cells re-pinned
(INV-002, 011, 021, 022, 033–047 range cells, 049, 055, 056, 059, 102, 104, 110, 111, 121).
Anchor cells NOT listed above were already derived from this same tree's grep in the
original freeze and were spot-re-confirmed by sweeps 1–4 (INV-010, 020, 030–032, 048,
057's L509/L585/L614, 058, 070–076 doc-line cites, 080–083, 090–092, 103, 105–107,
112, 113, 122–123, 130–135). exit codes: awk/grep all 0.

### CHK-000f — D3 new-row anchor derivation (Group X1)
Source: sweep outputs of CHK-000e items 2–4. Anchors frozen:
INV-136 → L230–238 (pointer L232); INV-137 → L499; INV-138 → L521;
INV-139 → L487 (cross-ref clause verified present in the exit-137 paragraph);
INV-140 → L645; INV-141 → L638. Row existence verified by inclusion in Group X1 table
(claims-inventory.md); mechanical existence checks themselves are run-phase work.
exit: 0

### CHK-001 — INV-033 handle-*.sh framing (templates tree surface)
$ ls internal/template/templates/.claude/hooks/moai/
48 entries including 24+ `handle-*.sh.tmpl` sources AND deployed `.sh` counterparts
for several handlers (handle-agent-hook.sh + .sh.tmpl pair; handle-stop-goal.sh +
.sh.tmpl both present; others .tmpl-only or .sh-only). exit: 0
Verdict basis: wrappers are template-derived PAIRS, so L119's "(not templates)" framing
is false → DEFECT-CONFIRMED (fix applied below, CHK-FIX-001).

### CHK-002 — INV-056 local-linear-integration.md tracked-vs-untracked
$ git -C <worktree> ls-files --error-unmatch .moai/docs/local-linear-integration.md
.moai/docs/local-linear-integration.md
exit: 0
$ test -f .moai/docs/local-linear-integration.md → FS_EXIT=0
Surface named: TRACKED worktree index (both surfaces checked: git-tracked AND on-disk).
"local-only" means dev-only/not-distributed — being tracked is consistent (unlike files
under CleanMoaiManagedPaths roots, `.moai/docs/` survives `moai update`). PASS-ATTESTED.

### CHK-003 — INV-057 legacy memory referents (global auto-memory dir)
$ ls memory/ in worktree → contains ONLY feedback_lsel_m3_smoke.md (exit 0)
$ ls /Users/goos/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/ | grep audit_sweep|lesson
audit_sweep_patterns.md ; lessons.md (also lesson_*.md split files) — FOUND.
$ grep -n "Pattern A" …/memory/audit_sweep_patterns.md → line 30 "### Pattern A: Space-separated allowed-tools detection" exit 0
$ grep -n "#4\|#5" …/memory/lessons.md → lines 43 "## #4", 87 "## #5" exit 0
Referent = GLOBAL Claude projects auto-memory dir, NOT repo-relative. Unreproducible from
repo root → AMBIGUOUS-PATH (defect class per method column); clarification fix applied
(recorded as CHK-FIX-020 in the resumed-session fix block below — the interrupted run's
forward-reference to "CHK-FIX-002..004" never materialized as entries and is superseded). Surface named: in-repo memory/ checked in WORKTREE; global dir is a
home-directory read (not a PRIMARY-checkout mutation).

### CHK-004 — INV-049(a) six named ymls exist in .moai/astgrep-rules (worktree surface)
$ ls .moai/astgrep-rules/go .moi… no — exact: ls .moai/astgrep-rules/go security
go/: concurrency.yml error-handling.yml hardcoding.yml idioms.yml resource-safety.yml (5)
security/: credentials.yml crypto.yml injection.yml secrets.yml web.yml (5)
All SIX historically-named ymls present (plus 5 newer additions). Historical anchors hold;
tree now holds 10 ymls total. HISTORICAL-RECORD portion PASS.

### CHK-005 — INV-049(b)/INV-074 rules_dir default logic @ this tree 8d8da0b2b
$ grep -rn "RulesDir\|rules-dir\|rules_dir" internal/cli/gate.go internal/hook/pre_tool.go internal/config/loader_gate.go internal/config/defaults.go
gate.go:138-139 "// copy (issue #1265…) An empty RulesDir stays empty (t50): gate.yaml
ast_grep_gate.rules_dir is the SSOT."; gate.go:208-209 "// t50: an empty gate.yaml
rules_dir maps through as empty — NO hardcoded path fallback." Same doctrine at
pre_tool.go:875-876. exit: 0

### CHK-006 — INV-049(c) CLI --rules-dir flag default
$ grep -rn "rules-dir" internal/ (non-test hits shown)
astgrep.go:69 cmd.Flags().StringVar(&flags.rulesDir,"rules-dir","","ast-grep rules directory
path (default: gate.yaml ast_grep_gate.rules_dir)") ; astgrep.go:105 empty-config message
"scanning with 0 rules". SAME for astedit.go:71. exit: 0
⇒ L142's "CLI --rules-dir 기본값은 여전히 구 경로" is FALSE now → DEFECT-CONFIRMED
(CHK-FIX-003). L165's "빈 값일 때만 기본 경로 폴백 — internal/cli/gate.go:160-161"
is FALSE twice over (behavior removed by t50; lines now hold mapConfigGateToQuality's
docstring) → DEFECT-CONFIRMED (CHK-FIX-004, INV-074).

### CHK-007 — INV-040 scripts/ci-watch claimed 5 files
$ find scripts/ci-watch -type f | wc -l → 5 (run.sh, lib/classify.sh, lib/timeout.sh,
lib/_common.sh, test/run_test.sh) exit: 0 — PASS-ATTESTED.

### CHK-008 — INV-041 scripts/ci-autofix claimed 4 files
$ find scripts/ci-autofix -type f | wc -l → 4 (classify.sh, log-fetch.sh,
test/classify_test.sh, test/log_fetch_test.sh) exit: 0 — PASS-ATTESTED.

### CHK-009 — INV-080 AstGrepGate defaults literal + loader chain
$ grep -A6 "AstGrepGate" internal/config/defaults.go → lines 581-585 literal
{Enabled: true, BlockOnError: false, WarnOnlyMode: true} — matches prose exactly.
$ ls internal/config/loader_gate.go → exists; loader.go:86 calls l.loadGateSection inside Load.
exit: 0 — PASS-ATTESTED. (Local config wiring cross-checked: .moai/config/sections/gate.yaml:49
carries ast_grep_gate with rules_dir SSOT comment.)

### CHK-FIX-001 — subject edit: L119 contradiction resolution (from CHK-001)
Comment reframed to template-derived pairs; line kept at 119 words on same line (no line-count change
above the §4.1 zone ⇒ live zone interval unchanged, re-derived as unchanged).


## Run-phase entries — resumed session (2026-08-27, manager-develop #2, basis re-pinned below)

### CHK-010 — basis re-pin + live exclusion-zone re-derivation (REQ-CLOCAL-001/004)
$ git rev-parse --short HEAD && git branch --show-current && grep -n '^### §4.1\|^## 5\.' CLAUDE.local.md && wc -l CLAUDE.local.md
8d8da0b2b
WT-clocal-audit
274:### §4.1 통합 레인 (git-flow, `develop` 원격 공유)
339:## 5. Version Management
     666 CLAUDE.local.md
exit: 0
Live exclusion interval at resume: L274–L337 (heading L274 → the `---` at L337 immediately
preceding `## 5.` at L339). Frozen-basis coordinates were 273–337; the +1 shift is the
partial fix diff already applied above the zone. Zone untouched for the remainder of the run.

### CHK-011 — INV-001 Last-Updated staleness
$ git log -1 --format='%h %ci' -- CLAUDE.local.md
20fab1786 2026-08-27 14:59:01 +0900
exit: 0
Header read `Last Updated: 2026-05-25` at freeze vs a 2026-08-27 content commit ⇒ DEFECT-CONFIRMED
(fix already applied in the interrupted run; attested here — CHK-FIX-002).

### CHK-012 — INV-002 twin `---` empty section
$ grep -n '^---$' CLAUDE.local.md
7 71 229 241 337 343 389 395 441 478 484 498 514 527 549 590 619 631 650
exit: 0
No two rule lines are separated by fewer than 5 lines ⇒ the L493/L495 twin pair is gone.
DEFECT-CONFIRMED at freeze, fix applied (CHK-FIX-003).

### CHK-013 — INV-003 §-numbering drift
$ grep -n '^## ' CLAUDE.local.md
## 1..## 17 sequential, then `## References` (L633, holding the §18–§27 consolidation rows),
then `## 28. LSEL 드레인 운영` (L652).
exit: 0
Numbering is self-consistent with the document's own stated consolidation (§18–§27 live as
rows inside References). No orphan or duplicated §-number. PASS-ATTESTED.

### CHK-014 — INV-010 work-location paths (worktree surface)
$ ls -d internal/template/templates/ && git rev-parse --show-toplevel
drwxr-xr-x@ 13 goos staff 416 internal/template/templates/
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t308
exit: 0
Relative structure holds. The absolute PRIMARY path in the subject was NOT probed
(REQ-CLOCAL-009 keeps the primary read-free outside the §28 exception) — recorded as a Gap.

### CHK-015 — INV-011 / INV-012 / INV-120 CLI verb surface
$ grep -rn 'Use:[[:space:]]*"' internal/cli/*.go | grep -v _test   [full output: raw/b2.txt]
init.go:47 `init [project-name]`; update.go:42 `update`; hook.go:31 `hook`;
glm.go:40 `glm [...]`; version.go:20 `version`; root.go:19 `moai`.
No top-level `build` verb — the only `build` uses are subcommands
(tool_policy.go:72 `build`, graph.go:265 `build`).
exit: 0 — PASS-ATTESTED for all three rows.

### CHK-016 — INV-020 template-source roots
$ ls -d internal/template/templates/.claude internal/template/templates/.moai internal/template/templates/CLAUDE.md; ls -d internal/template/templates/.agency
.claude OK / .moai OK / CLAUDE.md OK
ls: internal/template/templates/.agency: No such file or directory
exit: 1 (for the .agency probe)
The `.agency/` template root does not exist ⇒ DEFECT-CONFIRMED (fix applied, CHK-FIX-004).

### CHK-017 — INV-021 / INV-022 neutrality-guard anchors
$ ls -l .github/workflows/template-neutrality-check.yaml
-rw-r--r--@ 1 goos staff 3689 .github/workflows/template-neutrality-check.yaml
$ grep -n '25\.1\|25\.3' .moai/docs/template-internal-isolation-doctrine.md
9:### §25.1 정의 — Allowed vs Forbidden Content Classes
44:### §25.3 Pre-commit Self-Check (5-item Mandatory Checklist)
$ find . -name internal_content_leak_test.go -not -path './.git/*'
./internal/template/internal_content_leak_test.go
exit: 0 — both PASS-ATTESTED.

### CHK-018 — INV-030..INV-048 Local-Only block existence sweep
$ for p in <23 paths>; do [ -e "$p" ] && echo EXISTS || echo ABSENT; git ls-files --error-unmatch "$p"; done   [full output: raw/b3.txt, raw/b4.txt]
settings.local.json EXISTS/untracked (no template counterpart — `ls internal/template/templates/.claude/settings.local.json` → No such file); settings.json EXISTS + `settings.json.tmpl` present in templates;
agent-memory/ EXISTS untracked, no template mirror; lifecycle-sync-gate.md + repo-local-pr-policy.md
+ gitflow-lane-protocol.md EXIST tracked, and `internal/template/templates/.claude/rules/local/`
does not exist (negative check passes); commands/harness trio → github.md + release.md +
release-update.md (3); manifest.json EXISTS; hns-release-update-run.js EXISTS;
agents/harness specialists → hns-github-specialist.md, hns-release-specialist.md,
hns-release-update-specialist.md (exactly 3); hns-workflow-ci-loop/ holds SKILL.md (non-empty);
ci-watch-protocol.md + ci-autofix-protocol.md EXIST; cc-update-*.md → 3 files.
exit: 0 — PASS-ATTESTED for INV-030..039, 042, 043, 046, 048.

### CHK-019 — INV-045 last-cc-version.json
$ [ -e .moai/state/last-cc-version.json ] && echo EXISTS || echo ABSENT
ABSENT
$ grep -rln 'last-cc-version' internal/ .claude/
.claude/agents/harness/hns-release-update-specialist.md
exit: 0
The file is untracked runtime state produced by the release-update harness; it does not travel
into a worktree checkout, and it carries no template counterpart. The Local-Only claim
("never in templates") holds on the checked surface. PASS-ATTESTED (surface: worktree + template tree).

### CHK-020 — INV-047 runtime-directory rows vs template tree
$ for p in .moai/cache .moai/logs .moai/state .moai/specs .moai/plans .moai/reports .moai/manifest.json .moai/status_line.sh; do [ -e "internal/template/templates/$p" ] && echo MIRRORED || echo no-mirror; done
no-mirror .moai/cache · MIRRORED .moai/logs · MIRRORED .moai/state · no-mirror .moai/specs
no-mirror .moai/plans · MIRRORED .moai/reports · no-mirror .moai/manifest.json · no-mirror .moai/status_line.sh
$ ls -a internal/template/templates/.moai/logs internal/template/templates/.moai/state internal/template/templates/.moai/reports
logs/.gitkeep ; state/.gitkeep + state/chain/ ; reports/plan-audit/
exit: 0
Three of the eight rows sit under "Never in Templates" while the template tree DOES carry their
empty directory scaffold ⇒ DEFECT-CONFIRMED (precision defect; fix applied, CHK-FIX-005).

### CHK-021 — INV-044 / INV-083 historical anchors
$ git cat-file -t ed04e40e6 && git log -1 --format='%h %s' ed04e40e6
commit
ed04e40e6 chore(local): move local-only files out of the update-managed roots (#1557)
$ git tag -l v3.0.1 && git rev-parse -q --verify 'v3.0.1^{}' >/dev/null && echo TAG_OK
v3.0.1
TAG_OK
exit: 0 — both HISTORICAL-RECORD (anchors hold; narratives not re-executed).

### CHK-022 — INV-070 / INV-071 deploy.go citations
$ grep -n 'func CleanMoaiManagedPaths' internal/cli/update/deploy/deploy.go
107:func CleanMoaiManagedPaths(projectRoot string, out io.Writer, tmplFS fs.FS) error {
$ sed -n '180,200p' internal/cli/update/deploy/deploy.go
187: // Clean .moai/config/ entirely - backup was already done by the Backup step.
190: configDir := filepath.Join(projectRoot, defs.MoAIDir, defs.ConfigSubdir)
$ grep -n 'DisplayPath:' internal/cli/update/deploy/deploy.go
59 settings.json · 63 commands/moai · 67 agents/moai · 71 skills/moai* · 76 rules/moai
· 80 output-styles/moai · 84 hooks/moai   (7 targets + `.moai/config` handled separately = 8)
exit: 0
Symbol line 29 → 107 and `.moai/config` line 122 → 187-190: both citations were stale
⇒ DEFECT-CONFIRMED (fixes applied, CHK-FIX-006). The 8-root list itself matches the prose exactly.

### CHK-023 — INV-072 plan.go:73
$ sed -n '70,76p' internal/cli/update/plan/plan.go
		// Filter out MoAI-managed files - they are automatically installed
		if IsMoaiManaged(displayPath) {
			continue
		}
exit: 0 — line 73 holds `if IsMoaiManaged(displayPath) {`. PASS-ATTESTED.

### CHK-024 — INV-073 archiveLegacySkills call site (citation only)
$ grep -n 'archiveLegacySkills' internal/cli/update.go
554:		archived, archiveErr := archiveLegacySkills(cwd, out, getBoolFlag(cmd, "force"))
exit: 0
Cited `update.go:513` → actual 554 ⇒ citation DEFECT-CONFIRMED (fix applied, CHK-FIX-007).
The underlying code defect (call after the wipe) is NOT re-adjudicated — KNOWN-UNRESOLVED per REQ-CLOCAL-005.

### CHK-025 — INV-075 Makefile citations
$ sed -n '9p' Makefile
LDFLAGS := -ldflags "-s -w -X $(MODULE)/pkg/version.Version=$(VERSION) ..."
$ grep -n '^install:' Makefile && sed -n '41p' Makefile
40:install: ## Install the binary
41:	go install $(LDFLAGS) ./cmd/moai
exit: 0
`Makefile:9` holds. `Makefile:38` does NOT — line 38 is an `@echo "  Version: ..."` line;
the install recipe is at 40-41 ⇒ DEFECT-CONFIRMED (fix applied this run, CHK-FIX-008).

### CHK-026 — INV-076 pkg/version compile defaults
$ grep -rn 'Commit  =\|Date    =' pkg/version/version.go
9:	Commit  = "none"
10:	Date    = "unknown"
exit: 0 — PASS-ATTESTED.

### CHK-027 — INV-081 hook timeouts (template surface)
$ grep -o '"timeout":[0-9 ]*' internal/template/templates/.claude/settings.json.tmpl | sort | uniq -c
   5 "timeout": 10 ·  1 "timeout": 120 ·  2 "timeout": 30 · 23 "timeout": 5 · 3 "timeout": 60 · 2 "timeout": 900
exit: 0
The prevailing configured per-hook value is 5 (23 of 36 entries) and `{"timeout": 60}` is a real
configured value in the same file, so the recipe is actionable. PASS-ATTESTED (surface: template
settings.json.tmpl). Ambiguity vs Claude Code's own documented hook default recorded as residual risk.

### CHK-028 — INV-082 coverage-target enforcement
$ grep -n 'coverage' .moai/config/sections/quality.yaml
5:    test_coverage_target: 85
6:    coverage_threshold: 0
16:        min_coverage_per_commit: 80
$ grep -rn '90%' .moai/config/evaluator-profiles/strict.md
11:| Craft | 20% | Coverage >= 90%, mutation testing score >= 80% |
24:- Coverage below 90% = Craft FAIL
exit: 0
The 85% package-level figure is config-backed; the "cli/template/hook 90%+" per-package rule has
no located mechanical enforcer (only a global strict-profile gate) ⇒ DEFECT-CONFIRMED
(attribution defect; fix applied, CHK-FIX-009).

### CHK-029 — INV-090 / INV-091 env-var surface
$ grep -n 'func applyEnvOverrides' -A 30 internal/config/manager.go | grep 'os.Getenv'
399: EnvDevelopmentMode · 402: EnvLogLevel · 405: EnvLogFormat · 408: EnvNoColor   (four, not five)
$ grep -n 'EnvConfigDir' internal/config/manager.go
70, 106, 246   (config-dir resolver, separate from applyEnvOverrides)
$ grep -rn 'MOAI_USER_NAME\|MOAI_CONVERSATION_LANG' --include='*.go' internal/ pkg/ cmd/
(no matches) exit: 1
INV-090 ⇒ DEFECT-CONFIRMED: `MOAI_CONFIG_DIR` is NOT read by applyEnvOverrides (fix applied,
CHK-FIX-010). INV-091 ⇒ PASS-ATTESTED, scan scope named: `*.go` under internal/, pkg/, cmd/ —
zero readers. (Two prose hits exist in internal/config/CLAUDE.md, a doc, not code — recorded as a Gap.)

### CHK-030 — INV-092 config layout
$ ls .moai/config/config.yaml
ls: .moai/config/config.yaml: No such file or directory   exit: 1
$ ls .moai/config/sections/ | wc -l
32
exit: 0
The claimed main `config.yaml` does not exist on this tree, and the section list named a
`config.yaml - Main config` row that likewise does not exist ⇒ DEFECT-CONFIRMED
(fix applied this run, CHK-FIX-011). The env > user-yaml > template-defaults ordering itself is
consistent with CHK-029's observed resolver chain.

### CHK-031 — INV-100 embed directives
$ grep -n 'go:embed' internal/template/embed.go && ls internal/template/embedded.go
28://go:embed all:templates
29://go:embed catalog.yaml
ls: internal/template/embedded.go: No such file or directory
exit: 0 / 1 — PASS-ATTESTED (both directives present, no generated embedded.go).

### CHK-032 — INV-101 / INV-102 / INV-103 / INV-104 / INV-105 template-tree surface
$ ls internal/template/templates/.claude/settings.json.tmpl → EXISTS (INV-101: §8 cited the
  non-`.tmpl` name; DEFECT-CONFIRMED, fix applied CHK-FIX-012)
$ grep -n 'type TemplateContext struct' -A 12 internal/template/context.go → struct carries
  ProjectName/ProjectRoot/UserName/ConversationLanguage/... ; GoBinPath at :51, HomeDir at :52;
  NewTemplateContext :85, WithGoBinPath :230, WithHomeDir :237
  (INV-102: the two-field excerpt is a partial view of a much larger struct; DEFECT-CONFIRMED,
   excerpt-marker fix applied CHK-FIX-013)
$ grep -n 'func .*Deploy(' internal/template/deployer.go
137:func (d *deployer) Deploy(ctx context.Context, projectRoot string, m manifest.Manager, tmplCtx *TemplateContext) error
  (INV-103: the documented call `Deploy(ctx, projectRoot, mgr, ctx)` passes a *TemplateContext as
   the first parameter, which is context.Context ⇒ DEFECT-CONFIRMED, fix applied CHK-FIX-014)
$ grep -n 'GoBinPath\|PATH\|Resolved' internal/template/templates/.moai/status_line.sh.tmpl
22: if command -v moai; 25: {{if .ResolvedMoaiPath}}; 29-30: posixPath .ResolvedMoaiPath;
34-35: if [ -f "$HOME/go/bin/moai" ]; then exec "$HOME/go/bin/moai" statusline
  — NO `export PATH="{{.GoBinPath}}:$PATH"` line anywhere in the file
  (INV-104 ⇒ DEFECT-CONFIRMED, fix applied CHK-FIX-015; the replacement snippet quotes lines 34-35 verbatim)
$ grep -n 'posixPath\|claudeCodePassthroughTokens\|\$HOME' internal/template/renderer.go
27: "posixPath" func · 39: var claudeCodePassthroughTokens · 43: "$HOME"
  (INV-105 ⇒ PASS-ATTESTED)
exit: 0

### CHK-033 — INV-106 / INV-107 detection + language set
$ grep -rln 'project_markers\|ProjectMarkers' internal/ .moai/config/
internal/lsp/config/types.go · internal/template/templates/.moai/config/sections/lsp.yaml.tmpl · (+tests)
$ for l in go python typescript javascript rust java kotlin csharp ruby php elixir cpp scala r flutter swift dart; do grep -c "lsp.servers.$l\." internal/config/testdata/shipped_key_inventory.yaml; done
go:8 python:7 typescript:7 javascript:7 rust:8 java:7 kotlin:7 csharp:7 ruby:7 php:7 elixir:7
cpp:7 scala:7 r:7 flutter:7 swift:7 dart:0
exit: 0
All sixteen names resolve as `lsp.servers.<lang>` entries; `dart` resolves to zero, confirming
"flutter" as the canonical name. Both PASS-ATTESTED. (catalog.yaml carries no language entries at
all, so the inventory's suggested catalog-parity cross-check is not applicable — recorded as a Gap.)

### CHK-034 — INV-055 / INV-058 / INV-059 / INV-137 / INV-140 pointer existence
$ for f in version-management hook-development git-workflow-doctrine dev-only-commands-isolation local-dev-settings-intent git-local-workflow-doctrine harness-namespace-doctrine template-internal-isolation-doctrine local-linear-integration docs-site-i18n-rules docs-site-design-components; do [ -f ".moai/docs/$f.md" ] && echo OK || echo MISSING; done
OK ×11 (zero MISSING)
$ grep -n '^## \|^### ' .claude/skills/hns-lsel-curator/SKILL.md
132:## Durable operations — the session-start trigger (`session_drain.sh`)
170:## Verification (run before declaring a drain complete)
$ ls .claude/rules/moai/core/askuser-protocol.md .claude/rules/moai/workflow/orchestration-mode-selection.md → both exist; §E heading at orchestration-mode-selection.md:211 `## §E — Anti-Patterns`
$ ls .claude/rules/moai/development/{skill-authoring,agent-authoring}.md → both exist
$ ls .claude/commands/moai/harness.md → exists; .claude/skills/moai/workflows/harness.md → exists
exit: 0 — INV-055, 058, 059, 137, 140 all PASS-ATTESTED.

### CHK-035 — INV-110 / INV-111 cross-doc row accuracy
$ grep -n 'git-flow\|develop\|enforce_admins' .moai/docs/git-workflow-doctrine.md | head
5:> **[HARD] 상위 모델 변경 (2026-08-27) — GitHub Flow → git-flow.** … 정본은 `CLAUDE.local.md` §4.1 …
7-8: superseded 범위(= develop 없는 Enhanced GitHub Flow 전제) / 여전히 구속력 있는 항목(enforce_admins:true 등) 구분
$ grep -n 'PR-mandatory\|git-flow' .moai/docs/git-local-workflow-doctrine.md | head
5:> **[HARD] 상위 모델 변경 (2026-08-27) — GitHub Flow → git-flow.** 카드 작업은 develop 분기, 카드 PR 없음
150: (2026-07-20 신규) PR-mandatory: 모든 tier (S/M/L) 변경은 PR 경유
exit: 0
Both rows summarized the pre-transition state only ⇒ DEFECT-CONFIRMED for each
(fixes applied, CHK-FIX-016 / CHK-FIX-017).

### CHK-036 — INV-112 §2.3 restore recipe validity
$ git show HEAD:.moai/config/sections/git-strategy.yaml | sed -n '1,20p'
git_strategy:
    mode: manual
    …
    manual:
        workflow: gitflow
        …
        develop_branch: develop
exit: 0
HEAD carries `mode: manual` → `manual.workflow: gitflow` (NOT the template github-flow values),
so `git restore --source=HEAD -- .moai/config/sections/git-strategy.yaml` restores the committed
gitflow keys. The recipe is valid. PASS-ATTESTED.

### CHK-037 — INV-113 / INV-139 pointer reachability
$ grep -n '§4\.1' CLAUDE.local.md
247 (§4.1의 develop 통합 검증) · 274 (the heading itself) · 638 (§18 row: 정본은 §4.1)
$ sed -n '274,338p' CLAUDE.local.md | grep '^####'
§4.1.1 여섯 가지 규율 · §4.1.2 운영 절차 · §4.1.3 GitFlow 기각 사유 · §4.1.4 로컬 CI
$ sed -n '381p' CLAUDE.local.md
### Go Test Execution Rules
exit: 0
Both inward pointers (L247, L638) name §4.1 as a whole and the target exists with four real
subsections; the §11 → §6 cross-reference resolves to a live heading. Both PASS-ATTESTED.
Pointer-validation only — no §4.1 content was audited (REQ-CLOCAL-004).

### CHK-038 — INV-121 / INV-122 / INV-123 tooling surface
$ ls .claude/skills/moai/workflows/
plan.md run.md sync.md fix.md loop.md project.md feedback.md goal.md todo.md (+ clean, codemaps,
design, e2e, factory, gate, harness*, moai, mx, review, and the plan/run/sync/project subdirs)
$ grep -rl 'SPLIT_HARNESS_NAMESPACE_LEAK' --include='*.go' --include='*.md' --include='*.sh' --include='*.yaml' . | grep -v '^./.git' | grep -v '^./.moai/specs'
CLAUDE.local.md · .moai/docs/dev-only-commands-isolation.md · .moai/docs/harness-namespace-doctrine.md
· internal/template/split_namespace_test.go   ← the sentinel is live in real test code
$ grep -n '^install:\|^build:\|^help:\|##' Makefile
23:build: templ-generate ## Build the binary · 40:install: ## Install the binary · 83:help: ## Show this help
exit: 0 — all three PASS-ATTESTED.

### CHK-039 — INV-130..INV-135 LSEL §28 (PRIMARY surfaces named per finding)
SURFACE: PRIMARY checkout /Users/goos/MoAI/moai-adk-go (authorized §28 read exception)
$ jq -r '.hooks.SessionStart' <PRIMARY>/.claude/settings.local.json
one SessionStart group (matcher "startup|resume|clear|compact|fork") with exactly TWO hooks,
each `"timeout": 30`, args[2] = ${CLAUDE_PROJECT_DIR}/.claude/skills/hns-lsel-curator/session_drain.sh
and .../backlog_check.sh, each invoked with --inbox .../.moai/lessons-inbox.jsonl
--state-dir .../.moai/state/lsel
$ ls -d <PRIMARY>/.moai/state/lsel/clusters-history && ls <PRIMARY>/.moai/state/lsel/clusters-history | wc -l
directory exists; 61 archived copies
$ ls -l <PRIMARY>/.moai/lessons-inbox.jsonl
-rw-r--r--@ 1 goos staff 1293610 .../.moai/lessons-inbox.jsonl
SURFACE: WORKTREE (tracked)
$ ls .claude/skills/hns-lsel-curator/*.sh → session_drain.sh, backlog_check.sh, drain.sh (+ tests), all tracked, mode 0755
$ grep -n '\-\-inbox\|--state-dir' .claude/skills/hns-lsel-curator/session_drain.sh
53:    --inbox)     INBOX="$2"; shift 2;;
54:    --state-dir) STATE_DIR="$2"; shift 2;;
$ grep -n 'clusters-history' .claude/skills/hns-lsel-curator/session_drain.sh
21: # …<state-dir>/clusters-history/ BEFORE invoking drain.sh — drain.sh …
106:HISTORY_DIR="$STATE_DIR/clusters-history"
$ grep -n 'clusters.json' .claude/skills/hns-lsel-curator/drain.sh
49:CLUSTERS_FILE="$STATE_DIR/clusters.json"   (written on both the drain and no-op paths)
$ grep -rn 'LSEL_BACKLOG_THRESHOLD' .claude/skills/hns-lsel-curator/
backlog_check.sh:15:THRESHOLD="${LSEL_BACKLOG_THRESHOLD:-25}"
exit: 0 — INV-130..135 all PASS-ATTESTED.
Note (method): the first attempt selected on `.command`, which is `bash` for both entries, and
returned nothing. Absence on one query shape is not absence — the script path lives in `.args[2]`.

### CHK-040 — INV-136 / INV-138 §3 and §13 anchors
$ ls .claude/rules/moai/development/coding-standards.md → exists (11,276 bytes)
$ grep -n 'English' .claude/rules/moai/development/coding-standards.md
11:All instruction documents must be in English:
$ grep -rn 'func loadGLMKey' --include='*.go' internal/ cmd/
internal/cli/glm.go:963:func loadGLMKey() string {   → body: `return glmcred.Load()`
$ grep -n 'env.glm\|func Load' internal/glmcred/glmcred.go
42:// Path returns the absolute path to the GLM credential file (~/.moai/.env.glm).
95:// Load reads the GLM API key from ~/.moai/.env.glm.
exit: 0
§3's language-policy pointer resolves (the Go-specific conventions are stated as project-local,
not attributed to the rules file); §13's `loadGLMKey()` symbol exists and does read
`~/.moai/.env.glm`. Both PASS-ATTESTED.

### CHK-041 — INV-141 Vercel pricing row
No command executed. Pre-classified EXTERNAL-UNVERIFIABLE per REQ-CLOCAL-007 / spec §C
(external pricing figures are marked and left alone, zero further effort). exit: n/a

## Fix entries (run-phase, resumed session)

### CHK-FIX-002 — INV-001 Last-Updated stamp
`Last Updated: 2026-05-25` → `2026-08-27`. (Applied in the interrupted run; attested by CHK-011.)

### CHK-FIX-003 — INV-002 twin `---` removal
The duplicated rule between §11 and §12 removed. (Interrupted run; attested by CHK-012.)

### CHK-FIX-004 — INV-020 `.agency/` template root
Root line replaced with a dated removal note. (Interrupted run; attested by CHK-016.)

### CHK-FIX-005 — INV-047 runtime-directory precision (this run)
`.moai/logs/`, `.moai/state/`, `.moai/reports/` rows annotated: contents are local-only, but the
empty directory scaffold (`.gitkeep`, `state/chain/`, `reports/plan-audit/`) IS in the template tree.

### CHK-FIX-006 — INV-070 / INV-071 deploy.go anchors
`deploy.go:29` → `deploy.go:107`; `(deploy.go:122)` → `(deploy.go:187-190)`. (Interrupted run; CHK-022.)

### CHK-FIX-007 — INV-073 archiveLegacySkills anchor
`update.go:513` → `internal/cli/update.go:554`. (Interrupted run; CHK-024.)

### CHK-FIX-008 — INV-075 Makefile install anchor (this run)
`make install`(Makefile:38) → `make install`(Makefile:40-41; [2026-08-27 감사 정정]).

### CHK-FIX-009 — INV-082 coverage-target attribution
85% attributed to quality.yaml; the 90%+ per-package figure re-framed as a target whose only
mechanical backing is the strict evaluator profile's global gate. (Interrupted run; CHK-028.)

### CHK-FIX-010 — INV-090 MOAI_CONFIG_DIR split
applyEnvOverrides (manager.go:398-411, four vars) separated from the config-dir resolver
(manager.go:70). (Interrupted run; CHK-029.)

### CHK-FIX-011 — INV-092 config layout (this run)
The `.moai/config/config.yaml` "main configuration file" claim replaced with the measured state
(no such file; 32 section files), and the phantom `config.yaml - Main config` row dropped.

### CHK-FIX-012 — INV-101 template settings filename
§8 `internal/template/templates/.claude/settings.json` → `settings.json.tmpl`. (Interrupted run; CHK-032.)

### CHK-FIX-013 — INV-102 TemplateContext excerpt marker
Struct block annotated as a partial excerpt pointing at internal/template/context.go. (Interrupted run.)

### CHK-FIX-014 — INV-103 Deploy call example (this run)
`deployer.Deploy(ctx, projectRoot, mgr, ctx)` → the real signature as a comment plus
`deployer.Deploy(context.Background(), projectRoot, mgr, tmplCtx)`.

### CHK-FIX-015 — INV-104 status_line.sh.tmpl snippet
The non-existent `export PATH="{{.GoBinPath}}:$PATH"` example replaced with the file's actual
fallback-chain lines. (Interrupted run; CHK-032.)

### CHK-FIX-016 / CHK-FIX-017 — INV-110 / INV-111 References rows
§18 and §23 rows annotated with the 2026-08-27 git-flow transition and the §4.1 canonical pointer.
(Interrupted run; CHK-035.)

### CHK-FIX-018 — INV-033 companion (Local-Only additions, interrupted run)
`gitflow-lane-protocol.md` row added to the Local-Only block; `handle-*.sh` framing corrected
(the original CHK-FIX-001 edit).

### CHK-FIX-019 — INV-049 astgrep CLI default + INV-074 gate.go anchor
Both corrected to the t50 behavior (no fallback path; `astgrep.go:69,105`). (Interrupted run; CHK-005/006.)

### CHK-FIX-020 — INV-057 memory referent clarification
Three `memory/...` references re-pointed at the global project auto-memory directory. (Interrupted run; CHK-003.)

### CHK-042 — evidence-path repair: raw/b3.txt + raw/b4.txt regeneration
Provenance (stated plainly): the batch-3 and batch-4 invocations behind CHK-018 / CHK-019 /
CHK-020 were issued WITHOUT a file redirect — their output was read inline and never written to
disk, so the citations in verdict.md pointed at files that did not exist. The lead caught this
before commit. Because the tree had not moved, the commands were re-issued byte-identically and
their output persisted.

$ git rev-parse --short HEAD && git branch --show-current
8d8da0b2b
WT-clocal-audit
exit: 0
Same HEAD as CHK-010 (the resume basis) ⇒ the regenerated output is measured against the same
tree the original observation was, so the attribution chain for CHK-018/019/020 holds.

$ <batch-3 command, verbatim> > .moai/reports/t308/raw/b3.txt
exit: 0   (23-path EXISTS/ABSENT + tracked/untracked sweep, harness globs, specialist count,
           cc-update glob count, hns-workflow-ci-loop listing)
$ <batch-4 command, verbatim> > .moai/reports/t308/raw/b4.txt
exit: 0   (INV-039 triple, INV-042 SKILL.md, INV-030/034/035 template-absence probes,
           INV-045 producer grep, INV-047 cache-reference grep)

Each file carries a 3-line header stating it was regenerated in this session at HEAD 8d8da0b2b.
Cross-check against the values CHK-018/019/020 originally recorded: all reproduce identically
(3 harness specialists; SKILL.md present; `internal/template/templates/.claude/settings.local.json`
and `.../rules/local/` both absent; `.moai/state/last-cc-version.json` ABSENT with its producer at
`.claude/agents/harness/hns-release-update-specialist.md`; `.moai/cache` referenced by
`internal/worktree/state_guard.go:53`). No verdict changed.

Numbering note: the b5 and b6 slots are unused — no such file exists and nothing cites one.
(Written without the `raw/<name>` path shape on purpose: a path-resolution sweep over this pack
would otherwise read this sentence itself as a citation and report a dangling pointer.) Batches 5 and 6
(the template-mirror-absence sweep + historical anchors, and the Makefile/roots follow-up) were
likewise issued inline; their numbers were simply skipped when the redirect habit resumed at b7.
Their findings are recorded verbatim inside CHK-020 / CHK-021 / CHK-022 / CHK-025 rather than in a
raw file — those CHK entries quote the output directly and cite no external path.

$ grep -rn 'raw/b[0-9]*\.txt' .moai/reports/t308/*.md
checks-transcript.md:210 → raw/b2.txt   (exists)
checks-transcript.md:235 → raw/b3.txt, raw/b4.txt   (now exist)
verdict.md:70            → raw/b3.txt, raw/b4.txt   (now exist)
exit: 0 — every cited raw path resolves; no other raw citation exists in the pack.
