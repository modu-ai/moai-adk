# SPEC-CLOCAL-AUDIT-001 — Frozen Claims Inventory

Audit basis: `CLAUDE.local.md` @ develop `d29b8942e` (663 lines, worktree branch `WT-clocal-audit`). Exclusion zone: lines 273–337 (§4.1 git-flow section — sibling cards t294/t295/t298/t303 own residuals there; pointer-validation from outside sections only).

Verdict classes (REQ-CLOCAL-007): DEFECT-CONFIRMED · PASS-ATTESTED · HISTORICAL-RECORD · EXTERNAL-UNVERIFIABLE · KNOWN-UNRESOLVED · AMBIGUOUS-PATH · OPEN-QUESTION.
Status at freeze: ALL `pending`. Run-phase populates verdict + evidence pointer per item; closure requires zero blanks (AC-CLOCAL-006). Total frozen: **76** items.
Anchor provenance: ALL line anchors re-derived mechanically against d29b8942e during the plan-audit iter-1 repair (CHK-000e/000f in `checks-transcript.md`); ~18 stale cells corrected, six new rows appended (Group X1).

## Group H1 — Header / formatting / numbering (M7)

| ID | CLAUDE.local.md anchor | Claim class | Verification method | Verdict |
|----|------------------------|-------------|---------------------|---------|
| INV-001 | L5 `Last Updated: 2026-05-25` | staleness vs August edits | `git log -1 --format=%ci -- CLAUDE.local.md` | DEFECT-CONFIRMED (CHK-011 → CHK-FIX-002) |
| INV-002 | L493 + L495 twin `---` (blank L494 between) closing §11 before §12 heading | empty-section formatting defect | sed render of L488–500 | DEFECT-CONFIRMED (CHK-012 → CHK-FIX-003) |
| INV-003 | References stub block L632+ §18–§27 consolidation; §17.1, §2.x sub-numbering | §-numbering drift from stubs | heading grep sweep | PASS-ATTESTED (CHK-013) |

## Group Q1 — Quick Start / CLI surface (M5/M7)

| ID | Anchor | Claim class | Verification | Verdict |
|----|--------|-------------|--------------|---------|
| INV-010 | L13–16 Work Location paths | path existence | ls internal/template/templates/ ; pwd of repo root | PASS-ATTESTED (CHK-014 (absolute PRIMARY path unprobed — Gap)) |
| INV-011 | L28–46 moai CLI verb list (init/update/hook/glm/version) | CLI surface accuracy | `moai help` or source command registry | PASS-ATTESTED (CHK-015) |
| INV-012 | L48–51 NO top-level `moai build`; //go:embed all:templates pointer | tooling + embed pointer | CLI source grep + embed.go read | PASS-ATTESTED (CHK-015) |

## Group T1 — Template-First / neutrality pointers (M7)

| ID | Anchor | Verification | Verdict |
|----|--------|--------------|---------|
| INV-020 | L85–89 five template-source roots exist | ls each root | DEFECT-CONFIRMED (CHK-016 → CHK-FIX-004) |
| INV-021 | L110 CI guard `.github/workflows/template-neutrality-check.yaml` | test -f | PASS-ATTESTED (CHK-017) |
| INV-022 | L112 `.moai/docs/template-internal-isolation-doctrine.md` §25.1/§25.3 anchors; sibling `internal_content_leak_test.go` | file existence + heading grep | PASS-ATTESTED (CHK-017) |

## Group LO — Local-Only Files block, L114–143 (M1/M5)

Existence class unless noted. Local paths checked IN WORKTREE (tracked surface) AND on-disk (worktree FS); template-counterpart checks against `internal/template/templates/…`.

| ID | Anchor entry | Special trap | Verdict |
|----|--------------|--------------|---------|
| INV-030 | L116 .claude/settings.local.json | untracked surface — name which surface | PASS-ATTESTED (CHK-018) |
| INV-031 | L117 settings.json rendered from .json.tmpl | template counterpart filename | PASS-ATTESTED (CHK-018) |
| INV-032 | L118 .claude/agent-memory/ | — | PASS-ATTESTED (CHK-018) |
| INV-033 | L119 .claude/hooks/moai/handle-*.sh "(not templates)" | TRAP iii contradiction vs §2.3 pair-drift text — resolve NOW framing via templates tree contents | DEFECT-CONFIRMED (CHK-001 → CHK-FIX-018) |
| INV-034 | L120 rules/local/lifecycle-sync-gate.md | no-template-mirror negative check too | PASS-ATTESTED (CHK-018) |
| INV-035 | L121 rules/local/repo-local-pr-policy.md | same negative check | PASS-ATTESTED (CHK-018) |
| INV-036 | L122 commands/harness/{release-update,github,release}* trio | glob count ≥3 | PASS-ATTESTED (CHK-018) |
| INV-037 | L123 commands/harness/release-update/manifest.json | — | PASS-ATTESTED (CHK-018) |
| INV-038 | L124 workflows/hns-release-update-run.js | — | PASS-ATTESTED (CHK-018) |
| INV-039 | L125 agents/harness hns-{release-update,github,release}-specialist.md triple | count ==3 | PASS-ATTESTED (CHK-018) |
| INV-040 | L126 scripts/ci-watch/ CLAIMED 5 files | COUNT files — count drift = defect | PASS-ATTESTED (CHK-007) |
| INV-041 | L127 scripts/ci-autofix/ CLAIMED 4 files | COUNT | PASS-ATTESTED (CHK-008) |
| INV-042 | L128 skills/hns-workflow-ci-loop/ mirror kept | dir non-empty | PASS-ATTESTED (CHK-018) |
| INV-043 | L129 rules/local/ci-watch-protocol.md | — | PASS-ATTESTED (CHK-018) |
| INV-044 | L130 ci-autofix-protocol.md twin doctrine; #1557 ed04e40e6 commit; deployed twin regenerated noise | commit-anchor check (`git cat-file -t ed04e40e6`) = historical anchor | HISTORICAL-RECORD (CHK-021 (commit ed04e40e6 exists)) |
| INV-045 | L132 .moai/state/last-cc-version.json | runtime surface | PASS-ATTESTED (CHK-019 (runtime state; template-absence checked)) |
| INV-046 | L133 .moai/research/cc-update-*.md | glob non-empty | PASS-ATTESTED (CHK-018) |
| INV-047 | L134–140 cache/logs/state/specs/plans/reports + manifest.json generated-at-runtime | dirs exist; manifest regeneration plausible | DEFECT-CONFIRMED (CHK-020 → CHK-FIX-005) |
| INV-048 | L141 status_line.sh rendered from .sh.tmpl | counterpart in templates tree | PASS-ATTESTED (CHK-018) |
| INV-049 | L142 astgrep story: move 2026-08-15; six yml named go/{concurrency,error-handling,idioms,resource-safety}.yml security/{secrets,web}.yml; gate.yaml rules_dir wiring; CLI --rules-dir default still old | COUNT six files; source-grep gate.go default path | DEFECT-CONFIRMED (CHK-004/005/006 → CHK-FIX-019) |

## Group REF — Pointers & cross-doc rows (M2)

| ID | Anchor | Verification | Verdict |
|----|--------|--------------|---------|
| INV-055 | L634–645 eleven `.moai/docs/*.md` refs (version-management, hook-development, git-workflow-doctrine, dev-only-commands-isolation, local-dev-settings-intent, git-local-workflow-doctrine, harness-namespace-doctrine, template-internal-isolation-doctrine, local-linear-integration) + L618/L624 docs-site i18n/design-components | existence per row | PASS-ATTESTED (CHK-034 (11/11 exist)) |
| INV-056 | L644 local-linear-integration.md labeled local-only | TRACKED-vs-untracked determination (`git ls-files` vs fs); name surfaces | PASS-ATTESTED (CHK-002) |
| INV-057 | L509 memory/audit_sweep_patterns.md Pattern A; L585 lessons.md #5; L614 lessons.md #4 — NO directory anchor | locate in-repo vs global Claude projects memory dir; AMBIGUOUS-PATH if unreproducible from repo root | AMBIGUOUS-PATH (CHK-003 → CHK-FIX-020 (referent = global auto-memory dir)) |
| INV-058 | L651 SKILL.md § Durable operations; L663 § Verification | file exists (tracked) + both headings present | PASS-ATTESTED (CHK-034) |
| INV-059 | L637 §19 row pointers askuser-protocol.md + orchestration-mode-selection.md §E | existence + §E heading | PASS-ATTESTED (CHK-034) |

## Group B — Code-symbol + line citations at d29b8942e (M3)

| ID | Citation under audit | Verification | Verdict |
|----|---------------------|--------------|---------|
| INV-070 | L149 deploy.go:29 CleanMoaiManagedPaths | symbol@line | DEFECT-CONFIRMED (CHK-022 → CHK-FIX-006) |
| INV-071 | L152–154 eight wipe roots; "(deploy.go:122)" covering .moai/config specifically | symbol@line + root list equality | DEFECT-CONFIRMED (CHK-022 → CHK-FIX-006 (8-root list itself matches)) |
| INV-072 | L188 plan.go:73 `if IsMoaiManaged(...) { continue }` | symbol@line | PASS-ATTESTED (CHK-023) |
| INV-073 | L199 update.go:513 archiveLegacySkills-after-wipe — CITATION ONLY, defect KNOWN-UNRESOLVED | symbol@line | DEFECT-CONFIRMED (CHK-024 → CHK-FIX-007 (citation only; code defect stays KNOWN-UNRESOLVED)) |
| INV-074 | L165 gate.go:160-161 empty-rules_dir fallback | symbol@lines | DEFECT-CONFIRMED (CHK-006 → CHK-FIX-019) |
| INV-075 | L487 Makefile:9 LDFLAGS; Makefile:38 install = go install $(LDFLAGS) ./cmd/moai | symbol@line | DEFECT-CONFIRMED (CHK-025 → CHK-FIX-008 (Makefile:9 held; :38 stale)) |
| INV-076 | L487 pkg/version compile defaults Commit="none", Date="unknown" | const values in pkg/version | PASS-ATTESTED (CHK-026) |

## Group C — Defaults/values (M4)

| ID | Claim | Verification | Verdict |
|----|-------|--------------|---------|
| INV-080 | L201 AstGrepGate{Enabled:true, BlockOnError:false, WarnOnlyMode:true} in defaults.go; loader_gate.go exists; Loader.Load calls loadGateSection | literal field compare + call-chain grep | PASS-ATTESTED (CHK-009) |
| INV-081 | L486 hook timeout default 5초; recipe {"timeout":60} | template settings hook timeout actual values | PASS-ATTESTED (CHK-027 (surface: template settings.json.tmpl)) |
| INV-082 | L377–378 coverage 85% min; cli/template/hook 90%+ | locate enforcement (quality.yaml / evaluator-profiles / docs) | DEFECT-CONFIRMED (CHK-028 → CHK-FIX-009) |
| INV-083 | L201 released v3.0.1 lacked `loader_gate.go` (`loadGateSection`) → gate loader (issue #1265) [N2 folded 2026-08-27] | HISTORICAL anchor: tag v3.0.1 exists; cheap tag-tree absence check if feasible | HISTORICAL-RECORD (CHK-021 (tag v3.0.1 exists)) |

## Group D — Env-var surface (M4)

| ID | Claim | Verification | Verdict |
|----|-------|--------------|---------|
| INV-090 | L460–462 applyEnvOverrides reads exactly MOAI_DEVELOPMENT_MODE / MOAI_LOG_LEVEL / MOAI_LOG_FORMAT / MOAI_NO_COLOR / MOAI_CONFIG_DIR; constants in envkeys.go | switch-case compare; envkeys.go constants | DEFECT-CONFIRMED (CHK-029 → CHK-FIX-010) |
| INV-091 | L467–470 MOAI_USER_NAME / MOAI_CONVERSATION_LANG have ZERO readers in internal/, pkg/, cmd/ | full-tree grep WITH named scan scope recorded | PASS-ATTESTED (CHK-029 (scope: *.go under internal/, pkg/, cmd/ — 0 hits)) |
| INV-092 | L442–468 config layout (config.yaml main; sections quality/language/user/workflow.yaml …) + priority env > user yaml > template defaults | sections inventory + manager.go precedence read | DEFECT-CONFIRMED (CHK-030 → CHK-FIX-011) |

## Group E — Template-tree structure (M5)

| ID | Claim | Verification | Verdict |
|----|-------|--------------|---------|
| INV-100 | L225 embed.go carries BOTH //go:embed all:templates AND //go:embed catalog.yaml; NO generated embedded.go | directive grep + embedded.go absence | PASS-ATTESTED (CHK-031) |
| INV-101 | L402 §8 cites internal/template/templates/.claude/settings.json while L117 says .json.tmpl | ACTUAL filename — internal inconsistency candidate; fix imprecise side(s) | DEFECT-CONFIRMED (CHK-032 → CHK-FIX-012) |
| INV-102 | L417–420 struct + L431–435 constructor TemplateContext fields GoBinPath/HomeDir; NewTemplateContext(WithGoBinPath…, WithHomeDir…) | struct + constructor signatures | DEFECT-CONFIRMED (CHK-032 → CHK-FIX-013) |
| INV-103 | L435 deployer.Deploy(ctx, projectRoot, mgr, ctx) | function signature | DEFECT-CONFIRMED (CHK-032 → CHK-FIX-014) |
| INV-104 | L425–426 (code block L423–427) status_line.sh.tmpl contains export PATH="{{.GoBinPath}}:$PATH" | file content match | DEFECT-CONFIRMED (CHK-032 → CHK-FIX-015) |
| INV-105 | L538 posixPath registration; L540 renderer.go $HOME in claudeCodePassthroughTokens | identifier grep in renderer.go | PASS-ATTESTED (CHK-032) |
| INV-106 | L582 project_markers-based auto detection exists | identifier search (template/config code) | PASS-ATTESTED (CHK-033) |
| INV-107 | L568–575 sixteen-language list correctness; Dart canonical "flutter"; catalog.yaml parity | list compare + catalog.yaml language entries | PASS-ATTESTED (CHK-033 (16/16 lsp.servers entries; dart=0)) |

## Group F — Cross-document consistency (M2)

| ID | Claim | Verification | Verdict |
|----|-------|--------------|---------|
| INV-110 | L636 §18 row describes git-workflow-doctrine.md "(Enhanced GitHub Flow, branch protection enforce_admins:true, Hybrid Trunk RETIRED)" | row summary vs CURRENT doctrine content post-transition | DEFECT-CONFIRMED (CHK-035 → CHK-FIX-016) |
| INV-111 | L641 §23 row "PR-mandatory 1-person OSS, all tiers via PR" | row summary vs CURRENT doctrine content post-transition | DEFECT-CONFIRMED (CHK-035 → CHK-FIX-017) |
| INV-112 | L177–186 reinstall recipe presumes committed HEAD git-strategy.yaml holds workflow:gitflow keys | worktree HEAD .moai/config/sections/git-strategy.yaml content; restore-step validity — recipe sits OUTSIDE §4.1 so IN scope | PASS-ATTESTED (CHK-036) |
| INV-113 | pointers FROM outside sections INTO §4.1 subsections (the outside-section instances found by `grep -n '§4\.1' CLAUDE.local.md` — at the resumed basis: L247 and the §18 References row L638) [N1 folded 2026-08-27] | heading-target validation only | PASS-ATTESTED (CHK-037 (pointer-validation only)) |

## Group G — Tooling (M5/M7)

| ID | Claim | Verification | Verdict |
|----|-------|--------------|---------|
| INV-120 | (with INV-012) no top-level moai build | CLI help/source | PASS-ATTESTED (CHK-015) |
| INV-121 | L40–63 (terminal verbs L40–46; slash list L54–63; goal/todo audited repo-wide, absent from §1) slash commands wired [N3 folded 2026-08-27] (.claude/skills/moai/workflows + commands for plan/run/sync/fix/loop/project/feedback/goal/todo) | workflow file existence per verb | PASS-ATTESTED (CHK-038) |
| INV-122 | L639 SPLIT_HARNESS_NAMESPACE_LEAK sentinel string exists somewhere real | repo-wide grep with named scope | PASS-ATTESTED (CHK-038 (live in internal/template/split_namespace_test.go)) |
| INV-123 | L478 make targets build/install/help with ## help markers | Makefile target grep | PASS-ATTESTED (CHK-038) |

## Group H — LSEL §28 via PRIMARY surfaces (M6)

Every finding NAMES THE SURFACE checked.

| ID | Claim | Verification | Surface | Verdict |
|----|-------|--------------|---------|---------|
| INV-130 | L655 primary .claude/settings.local.json SessionStart has exactly 2 lsel entries each `"timeout": 30` wiring session_drain.sh wrapper + backlog_check.sh | jq extraction | PRIMARY | PASS-ATTESTED (CHK-039 — PRIMARY) |
| INV-131 | wrappers session_drain.sh / backlog_check.sh exist + tracked status + flags --inbox/--state-dir accepted | file existence + arg parsing read | BOTH surfaces named | PASS-ATTESTED (CHK-039 — WORKTREE(tracked) + PRIMARY wiring) |
| INV-132 | underlying drain.sh exists; unconditional clusters.json overwrite behavior matches prose | script read | named | PASS-ATTESTED (CHK-039 — WORKTREE) |
| INV-133 | L663 .moai/state/lsel/clusters-history/ exists (untracked runtime data) | directory listing | PRIMARY | PASS-ATTESTED (CHK-039 — PRIMARY (61 copies)) |
| INV-134 | default threshold 25 + env key LSEL_BACKLOG_THRESHOLD | constant location (backlog_check.sh or SKILL.md) | named | PASS-ATTESTED (CHK-039 — WORKTREE (backlog_check.sh:15)) |
| INV-135 | example path .moai/lessons-inbox.jsonl used by wiring exists | file existence | PRIMARY | PASS-ATTESTED (CHK-039 — PRIMARY) |

## Group X1 — Plan-audit iter-1 additions (D3 coverage holes; anchors from CHK-000f)

| ID | Anchor | Claim / cluster | Milestone | Verdict |
|----|--------|-----------------|-----------|---------|
| INV-136 | L230–238 §3 Code Standards: coding-standards.md auto-load pointer (L232); snake_case file-naming (L236); `fmt.Errorf("…: %w", err)` wrapping (L237); English-code policy (L238) | pointer existence vs `.claude/rules/moai/development/coding-standards.md`; convention text matches rules content | M5 | PASS-ATTESTED (CHK-040) |
| INV-137 | L499 §12 pointer pair `.claude/rules/moai/development/{skill-authoring.md, agent-authoring.md}` | file existence per name | M2 | PASS-ATTESTED (CHK-034) |
| INV-138 | L521 §13 `loadGLMKey()` symbol existence in internal/ tests + `~/.moai/.env.glm` path convention | symbol grep with named scope (internal/, cmd/) + path-grep of call site | M4 | PASS-ATTESTED (CHK-040) |
| INV-139 | L487 §11 inward cross-ref "binary lag 검증은 §6 검증 규율(clear → 측정)" | target §6 Go Test Execution Rules exists/reachable (L380ff) — VALIDATE REACHABILITY ONLY; Makefile:9/:38/pkg-version anchors on the same line are owned by INV-075/076 and linked here | M7 | PASS-ATTESTED (CHK-037) |
| INV-140 | L645 References §27 Agent-Skill Architecture row claims (≥1 skill set 4 elements; `/moai:<sub>` slash-wrapping; `/moai:harness` meta-harness v4 Builder; Analyze-First) | cheap surface validation (`/moai:harness` command presence, workflow skill names); prose policy itself not adjudicated | M2 | PASS-ATTESTED (CHK-034) |
| INV-141 | L638 References §20 Vercel pricing row ($0.0035/$0.126 per min; billing doc quotes) | PRE-CLASSIFIED `EXTERNAL-UNVERIFIABLE` — record-only row, zero further effort per REQ scope rule 5 | M7 | EXTERNAL-UNVERIFIABLE (CHK-041 (record-only per REQ-CLOCAL-007)) |

## Closure tally target

AC-CLOCAL-006: all 76 rows populated; classes sum to 76.
