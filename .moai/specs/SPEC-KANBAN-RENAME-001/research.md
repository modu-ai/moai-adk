---
id: SPEC-KANBAN-RENAME-001
title: "Research — measurements underlying the Factory Mode to Kanban Mode rename"
version: "0.3.0"
created: 2026-08-10
updated: 2026-08-11
author: manager-spec
priority: Medium
phase: "v3.1.0 target"
module: cli
lifecycle: spec-anchored
tags: "rename, research, measurements, evidence, template-mirror"
tier: L
---

## §A. What this file is, and how to read it

Every measured fact `spec.md`, `plan.md`, `acceptance.md`, and `design.md` rely on, recorded as **command → observed output → what it establishes**, so a later reader re-runs rather than re-derives. It is a Tier L artifact added at v0.2.0 with the promotion.

Two disciplines bind it. **Measurements are time-stamped by their tree, not trusted forever** — every count below is a fact about this repository at authoring time, and a run-phase that finds a different number trusts its own measurement and records the delta. And **an output is recorded even when it contradicts what the SPEC already asserts**; §H exists because three did.

Every command below was **re-run at v0.2.0 authoring time**, not transcribed from the existing prose. Working directory throughout: the worktree `/Users/goos/.moai/worktrees/kanban`, at HEAD `d39e3cdc6`, working tree clean except the four untracked kanban SPEC directories. `claude` version `2.1.226`.

The token pattern referred to below as `$TOK` is defined once, in `spec.md` §D.1, and copied byte-identically into `acceptance.md` AC-KR-021. It is not redefined here — a third copy would be a third thing to keep in sync, and `AC-KR-027` exists because two already drifted once.

---

## §B. The Kanban Mode surface census

### B.1 The whole surface, at file granularity

```
$ grep -rlniIE "$TOK" internal/ .claude/ .moai/project/ | wc -l
      28
```

The 28, enumerated:

```
.claude/rules/moai/workflow/goal-directive.md
.claude/skills/moai/workflows/factory.md
.claude/skills/moai/workflows/moai.md
.claude/skills/moai/workflows/run.md
.claude/skills/moai/workflows/run/mode-orchestration.md
.claude/skills/moai/workflows/sync/quality-gates-quality.md
.moai/project/codemaps/modules.md
.moai/project/structure.md
internal/cli/cc.go
internal/cli/cc_test.go
internal/cli/cg.go
internal/cli/cg_test.go
internal/cli/factory.go
internal/cli/glm.go
internal/cli/glm_test.go
internal/cli/launcher_blockcap_infinite.go
internal/cli/launcher_blockcap_infinite_test.go
internal/config/envkeys.go
internal/factory/record.go
internal/factory/record_test.go
internal/factory/revision.go
internal/factory/revision_test.go
internal/template/templates/.claude/rules/moai/workflow/goal-directive.md
internal/template/templates/.claude/skills/moai/workflows/factory.md
internal/template/templates/.claude/skills/moai/workflows/moai.md
internal/template/templates/.claude/skills/moai/workflows/run.md
internal/template/templates/.claude/skills/moai/workflows/run/mode-orchestration.md
internal/template/templates/.claude/skills/moai/workflows/sync/quality-gates-quality.md
```

Broken down by kind:

```
$ grep -rlniIE "$TOK" internal/ --include='*.go' | grep -v '_test.go' | wc -l
       8
$ grep -rlniIE "$TOK" internal/ --include='*_test.go' | wc -l
       6
```

**What this establishes.** The surface partitions cleanly as **8 Go non-test + 6 Go test + 6 harness docs + 6 template mirrors + 2 project documents = 28**, which is the denominator `AC-KR-021` measures against and the number `REQ-KR-021` drives to zero. The partition is exhaustive and the four groups do not overlap, so a milestone that clears one group can be verified independently of the others — which is what makes M1, M2, and M2b separably committable.

The Go non-test figure is **8**, confirming the v0.1.1 correction (D9) rather than the v0.1.0 figure of 7. The prose in `spec.md` §A.3 enumerated eight paths while the label said seven; the enumeration was right. See §H.1 for the correction's provenance.

### B.2 The two files outside the original scope

```
$ grep -rnIE "$TOK" .moai/project/
.moai/project/structure.md:139:> **`internal/` package count.** Measured `49` top-level packages … Exactly one of the three-package gap — `internal/factory` — belongs to SPEC-FACTORY-MODE-001 …
.moai/project/codemaps/modules.md:157:### internal/factory
.moai/project/codemaps/modules.md:246:> 실측 명령: `ls -d internal/*/ | wc -l` → 49 … 세 패키지 차이 중 `internal/factory` 하나만 SPEC-FACTORY-MODE-001이 추가한 것이고 …
```

and the fuller picture from a bare-word grep of the same two files:

```
$ grep -n -i factory .moai/project/codemaps/modules.md
157:### internal/factory
158:**역할**: Factory 모드 상태 — `moai cc -f` / `moai glm -f`가 여는 plan→run→verify→sync 체인의 세션 기록과 중복 억제
161:**진입점**: `internal/cli/factory.go` (플래그 파싱, env 진입/복원), `internal/cli/launcher_blockcap_infinite.go` (Stop-hook block cap 상향)
246:> 실측 명령: … `internal/factory` 하나만 SPEC-FACTORY-MODE-001이 추가한 것이고 …

$ grep -n -i factory .moai/project/structure.md
139:> **`internal/` package count.** … `internal/factory` … `origin/main` already carries `48`, and `internal/factory` is absent there …
```

**What this establishes.** Both files name a package path and a flag that the rename deletes, so after the rename each would describe a package that does not exist and a flag that does not work. The v0.1.0 scope of `internal/` plus `.claude/` would have returned a clean completion grep over exactly that staleness. This is the measurement behind `design.md` §F.1 and `REQ-KR-024`.

It also establishes that both passages are **hand-authored measurement narrative** — `structure.md` line 139 and `modules.md` line 246 each record a package count with its command, its prior value, and the reason the prior value was wrong. That is the content a regeneration pass would not reliably reproduce, and the second ground on which `design.md` §F.1 rejects deferring these files to `/moai project`.

### B.3 docs-site is not involved

```
$ grep -rni factory docs-site/content/ | wc -l
       4
$ grep -rni factory docs-site/content/
docs-site/content/en/claude-code/agentic/best-practices.md:69:| **Point to sources** | `Why is the ExecutionFactory API weird?` | …
docs-site/content/ko/claude-code/agentic/best-practices.md:69:| **출처 지목** | `ExecutionFactory API가 왜 이상해?` | …
docs-site/content/ja/claude-code/agentic/best-practices.md:69:| **出典の指定** | `ExecutionFactory の API はなぜ変なの?` | …
docs-site/content/zh/claude-code/agentic/best-practices.md:69:| **指明出处** | `ExecutionFactory の git 履歴` … |
```

**What this establishes.** All four hits are the string `ExecutionFactory` in one Claude Code best-practices example table, one per locale, at the same line number in each. None names Kanban Mode. This SPEC commissions no docs-site work and `AC-KR-023` asserts the tree is untouched — which matters because docs-site work carries a 4-locale same-PR obligation this SPEC would otherwise inherit.

---

## §C. The template-mirror pair inventory

Six pairs, each measured by `diff .claude/<path> internal/template/templates/.claude/<path>`:

| Label | Path | `diff` lines | Relationship |
|---|---|---:|---|
| `contract` | `skills/moai/workflows/factory.md` | 0 | byte-identical |
| `run` | `skills/moai/workflows/run.md` | 0 | byte-identical |
| `goal` | `rules/moai/workflow/goal-directive.md` | 0 | byte-identical |
| `moaidoc` | `skills/moai/workflows/moai.md` | 5 | sanitized pair |
| `modeorch` | `skills/moai/workflows/run/mode-orchestration.md` | 7 | sanitized pair |
| `qgates` | `skills/moai/workflows/sync/quality-gates-quality.md` | 9 | sanitized pair |

The three sanitized deltas, verbatim:

```
$ diff .claude/skills/moai/workflows/moai.md internal/template/templates/.claude/skills/moai/workflows/moai.md
277,278c277
< Updated: 2026-07-09
< Source: SPEC-MOAI-001. Named pipeline gates + agentic completion loop + chaining policy (v3.0.0). …
---
> Named pipeline gates + agentic completion loop + chaining policy (v3.0.0). …

$ diff .claude/skills/moai/workflows/run/mode-orchestration.md internal/template/templates/.claude/skills/moai/workflows/run/mode-orchestration.md
47,48c47
< --spawn` teammate windows) is unaffected — only MoAI's static team-orchestration
< layer is retired.
---
> --spawn` teammate windows) is unaffected — only MoAI's static team-orchestration layer is retired.
147d145
< Updated: 2026-03-30

$ diff .claude/skills/moai/workflows/sync/quality-gates-quality.md internal/template/templates/.claude/skills/moai/workflows/sync/quality-gates-quality.md
162,164c162
< **Concurrent scheduling with the Phase 7-10 audit (A7 — SPEC-SYNC-PARALLEL-DOCS-001).** …
< 
< **[HARD] P1/P2 violations BLOCK sync, and halt BEFORE Phase 10 coverage (A7 — SPEC-SYNC-PARALLEL-DOCS-001).** …
---
> **[HARD] P1/P2 violations BLOCK sync.** …
166,167d163
< **No-false-abort guard (SPEC-SYNC-PARALLEL-DOCS-001 A7).** …
<
```

**What this establishes.** The premise "the local and template `.claude/**` counterparts are byte-identical" is **false at HEAD**: exactly half the pairs carry a delta. Every stripped item is a live instance of `CLAUDE.local.md` §25 Template Internal-Content Isolation — two `Updated:` internal dates, a `Source: SPEC-MOAI-001` line, and three paragraphs citing `SPEC-SYNC-PARALLEL-DOCS-001` — plus one line-wrap difference in `modeorch` that is not a §25 strip at all but is part of the same measured delta.

Two consequences the SPEC turns on. A parity invariant would have instructed the implementer to copy the local file over the template file, **re-introducing forbidden content and tripping the neutrality guard** — the invariant would have mandated the violation. And a pair that "becomes identical" after the rename is a **failure**, not an improvement, which is why `AC-KR-017` compares deltas rather than asserting equality.

The classification is **time-varying**: any commit touching either side moves it. It is re-measured at M0 before the first edit, because the pre-rename delta cannot be reconstructed once the edit has landed.

The `modeorch` line-wrap delta is worth one note of its own: it is the one difference with no §25 justification, so a future pass that "tidies" it would be closing a gap this SPEC's invariant requires preserving. Delta preservation is indifferent to whether a delta is justified — it preserves what is measured.

---

## §D. The unrelated "factory" vocabulary, and the pattern's falsification

### D.1 The bare-word census

```
$ grep -rliI factory internal/ .claude/ | wc -l
     134

$ comm -23 <(grep -rliI factory internal/ .claude/ | sort) \
           <(grep -rlniIE "$TOK" internal/ .claude/ | sort) | wc -l
     108
```

A sample of the 108:

```
.claude/rules/moai/core/agent-hooks.md
.claude/rules/moai/core/moai-constitution.md
.claude/rules/moai/development/agent-patterns.md
.claude/rules/moai/development/orchestrator-templates.md
.claude/rules/moai/development/skill-ab-testing.md
.claude/rules/moai/development/skill-writing-craft.md
.claude/rules/moai/languages/cpp.md
.claude/rules/moai/languages/flutter.md
```

**What this establishes.** A bare-word completion criterion would require an allowlist of **108** files — measured exactly, not estimated. `spec.md` §C and `plan.md` §B-2 describe this population as "roughly 110"; the measured figure is 108, and the approximation is fair (see §H.3). The token-scoped pattern excludes them **by construction** rather than by enumeration, which is what makes `REQ-KR-021` a mechanical criterion instead of a judgment call.

### D.2 Falsification against trees this SPEC does not touch

```
$ for t in internal/lsp internal/tui internal/hook internal/core docs-site/ \
           .claude/skills/moai/references/anti-patterns.md; do
    echo "$t -> $(grep -rlniIE "$TOK" "$t" | wc -l)"; done
internal/lsp -> 0
internal/tui -> 0
internal/hook -> 0
internal/core -> 0
docs-site/ -> 0
.claude/skills/moai/references/anti-patterns.md -> 0

$ grep -rli factory internal/lsp | wc -l
       9
$ grep -ci factory .claude/skills/moai/references/anti-patterns.md
       3
```

**What this establishes.** The pattern returns zero on every tree this SPEC does not touch, while the bare word still returns 9 files in `internal/lsp` and 3 occurrences in the anti-patterns reference. The pattern therefore discriminates: it is not matching nothing (which a typo would also achieve), and it is not sweeping the unrelated vocabulary (which would make `REQ-KR-021` unsatisfiable). `AC-KR-022` is the criterion that keeps both halves observable after the rename.

**Discrepancy, recorded in §H.2:** two of the falsification targets named in `spec.md` §D.1 do not resolve as written.

---

## §E. The `-k` collision probe

```
$ claude --version
2.1.226 (Claude Code)

$ claude --help > /tmp/kr-claude-help.txt 2>&1; echo "help-exit=$?"
help-exit=0

$ grep -E '(^|[^-])-k[ ,]' /tmp/kr-claude-help.txt; echo "k-grep-exit=$?"
k-grep-exit=1

$ grep -oE '(^|[[:space:]])-[a-zA-Z][,[:space:]]' /tmp/kr-claude-help.txt \
    | tr -d ' ,' | sort -u | tr '\n' ' '
-c -d -h -n -p -r -v -w

$ grep -rn '"-k"' --include='*.go' internal/ | wc -l
       0
```

**What this establishes.** On `claude` `2.1.226` the short-flag set is `-c -d -h -n -p -r -v -w` and `-k` is free; no Go source in `internal/` carries a `"-k"` literal either. The mapping in `spec.md` §A.1 is safe on this tree.

**Why the gate survives the answer.** The launcher **strips its own switch before pass-through**, so a `claude`-defined `-k` would be silently shadowed rather than rejected — a failure with no error message. The `claude` CLI surface drifts between versions and the run-phase tree may sit on a different one, so `REQ-KR-003` keeps M0 as a **re-confirmation** gate. The version is recorded above precisely so a run-phase on a different version knows what it is re-confirming against.

**Stated limitation, carried with the result.** The probe pattern matches `-k ` and `-k,` renderings and would miss a `-k=<value>` form. A null result is strong evidence, not proof. This limitation is recorded rather than engineered away because the alternative — a pattern loose enough to catch every rendering — would match prose.

---

## §F. Release exposure, and why no alias is needed

```
$ for t in v3.0.1 v3.1.0-rc.0 v3.1.0-rc.1; do
    git cat-file -e "$t:internal/cli/factory.go" 2>/dev/null \
      && echo "$t: present" || echo "$t: absent"; done
v3.0.1: absent
v3.1.0-rc.0: present
v3.1.0-rc.1: present

$ grep -ci factory CHANGELOG.md
0
```

**What this establishes.** The flag exists only in release candidates, never in a stable release, and no CHANGELOG entry has ever mentioned it. **No released user can depend on `-f`.** A deprecation alias would therefore be dead code from the moment it shipped, carrying a permanent maintenance cost for a population of zero. This is the measured basis for `REQ-KR-002` and `design.md` §C.2 — the decision rests on release exposure, not on a judgment about how much users like compatibility windows.

It also dates the rename's window: the cost of renaming rises the moment `v3.1.0` ships stable, which is why this SPEC runs ahead of the board work rather than inside it.

---

## §G. The identifier surface, verbatim

```
$ grep -n 'Factory' internal/config/envkeys.go
130:	// EnvMoaiFactory carries the Factory Mode signal from the launcher entry
142:	EnvMoaiFactory = "MOAI_FACTORY"
144:	// EnvMoaiFactorySpec names the SPEC a factory chain targets. It is set only
147:	EnvMoaiFactorySpec = "MOAI_FACTORY_SPEC"

$ grep -nE 'func |const |EnvMoaiFactory|factoryFlag|Sentinel' internal/cli/factory.go
26:const (
27:	factoryFlagLong  = "--factory"
28:	factoryFlagShort = "-f"
31:// factoryUnsupportedBackendSentinel is the machine-greppable marker on the
35:const factoryUnsupportedBackendSentinel = "FACTORY_MODE_UNSUPPORTED_BACKEND"
50:func parseFactoryFlag(args []string) (spec string, enabled bool, rest []string) {
59:		if arg != factoryFlagLong && arg != factoryFlagShort {
89:func enterFactoryMode(specID string) func() {
90:	restoreFactory := captureEnvState(config.EnvMoaiFactory)
91:	restoreSpec := captureEnvState(config.EnvMoaiFactorySpec)
93:	_ = os.Setenv(config.EnvMoaiFactory, "1")
95:		_ = os.Setenv(config.EnvMoaiFactorySpec, specID)
106:func captureEnvState(key string) func() {
125:func recordFactorySession(specID, backend string) {
136:func rejectFactoryOnCG(args []string) error {

$ grep -n 'stateDirSegments' internal/factory/record.go
43:var stateDirSegments = []string{".moai", "state", "factory"}

$ wc -l internal/cli/factory.go internal/factory/record.go internal/factory/revision.go
     143 internal/cli/factory.go
     184 internal/factory/record.go
     184 internal/factory/revision.go
     511 total

$ grep -rc 'AC-FM-' --include='*_test.go' internal/ | awk -F: '{s+=$2} END{print s}'
50
```

**What this establishes.** Every identifier `REQ-KR-005` and `REQ-KR-007` name exists at the line shown, and the three LOC figures in `spec.md` §A.3 (143 / 184 / 184) are confirmed unchanged. `captureEnvState` (line 106) carries no mode-specific token, which is the measured basis for `REQ-KR-006` leaving it alone. `stateDirSegments` is a single `[]string` literal, so `REQ-KR-009` is a one-line edit expressed through the existing constant rather than a new one.

The **50** `AC-FM-` occurrences are the baseline `AC-KR-013` compares against. The criterion is deliberately a count rather than an absence: an absence check passes a corpus where every citation was renamed, which is the exact defect `REQ-KR-012` forbids.

---

## §H. Where the re-measurement disagreed

Three disagreements surfaced when the measurements were re-run at v0.2.0. Each is recorded with what it does and does not change; none is repaired by editing a requirement or a criterion, since those passed audit and are not re-opened by this promotion.

### H.1 The Go non-test count: 7 → 8, confirmed rather than newly found

The v0.1.0 draft labelled the Go non-test surface **7**; v0.1.1 corrected it to **8** (audit delta D9). The re-measurement confirms **8**:

```
$ grep -rlniIE "$TOK" internal/ --include='*.go' | grep -v '_test.go' | wc -l
       8
```

The eight are `internal/cli/{cc,cg,factory,glm,launcher_blockcap_infinite}.go`, `internal/config/envkeys.go`, and `internal/factory/{record,revision}.go`. The v0.1.0 prose had already enumerated eight paths beside the label saying seven — the label was the error, not the enumeration. **Attribution note:** the correction belongs to v0.1.1 and its provenance is preserved as such; it is confirmed here, not re-discovered.

### H.2 Two falsification targets in `spec.md` §D.1 do not resolve

At v0.2.0 `spec.md` §D.1 stated the pattern was falsified against, among others, `NOTICE.md` and `references/anti-patterns.md`. Measured:

```
$ test -e NOTICE.md && echo EXISTS || echo ABSENT
ABSENT
$ test -e references/anti-patterns.md && echo EXISTS || echo ABSENT
ABSENT
$ find . -maxdepth 2 -iname 'NOTICE*' -not -path './.git/*'
(no output)
$ find . -name 'anti-patterns.md' -not -path './.git/*'
./.claude/skills/moai/references/anti-patterns.md
./internal/template/templates/.claude/skills/moai/references/anti-patterns.md
```

There is **no `NOTICE.md` anywhere in this tree**, at the root or below. The `revfactory/harness` Apache-2.0 attribution that `spec.md` §C cited as living there is instead found under `.moai/research/*.md` — a directory outside every grep scope this SPEC defines. And `references/anti-patterns.md` is a path shorthand that does not resolve from the repository root; the real path is `.claude/skills/moai/references/anti-patterns.md`.

The five files carrying the `revfactory/harness` attribution, measured: all under `.moai/research/` (`v3-redesign-blueprint-2026-05-22.md`, `harness-system-audit-2026-05-14.md`, `harness-autonomy-vision-2026-05-18.md`, `docs-site-v2.14-to-HEAD-update-plan-2026-05-20.md`, `v3.0-redesign-2026-05-23.md`) — a directory outside every grep scope this SPEC defines.

**What this changes, and what it does not.** A falsification run against a nonexistent path returns zero **vacuously** — it establishes nothing. Two of the seven falsification targets were therefore doing no work. The falsification itself survives, because the remaining targets are real and the pattern was re-run against the corrected path in §D.2: `internal/lsp`, `internal/tui`, `internal/hook`, `internal/core`, `docs-site/`, and `.claude/skills/moai/references/anti-patterns.md` all return **0** while the bare word still returns 9 and 3 respectively. The `spec.md` §C characterization of the unrelated vocabulary is unaffected in substance — the attribution exists, it is simply not in `NOTICE.md` — and the file is not in scope either way.

**Disposition, revised at v0.3.0: corrected in `spec.md`, not merely reported.** The v0.2.0 disposition was to report and leave the prose, on the ground that editing it would touch a requirement surface the audit had passed. That ground does not hold for a *false statement of fact*: `spec.md` §C asserted an attribution living in a file that does not exist, and §D.1 listed two falsification targets that resolve to nothing. Correctness of a claim about the repository is not immunised by a prior score, and the edit is a path correction touching no requirement, no criterion, and no count. Both were corrected at v0.3.0; this section remains as the provenance of the correction.

### H.3 The unrelated-vocabulary population: "roughly 110" is 108

```
$ comm -23 <(grep -rliI factory internal/ .claude/ | sort) \
           <(grep -rlniIE "$TOK" internal/ .claude/ | sort) | wc -l
     108
```

`spec.md` §C and `plan.md` §B-2 say "roughly 110 files". The measured figure is **108**. The approximation is honest and the argument does not move — the ground is that an allowlist of this order is unmaintainable, and 108 is as unmaintainable as 110. Recorded so a later reader who re-runs the command and gets 108 knows the two figures are the same measurement and not a drift.

### H.4 The token grep is file-accurate and line-incomplete on one file

This is not a disagreement with a stated figure but a property the SPEC does not record, found while measuring §B.2.

```
$ grep -rnIE "$TOK" .moai/project/codemaps/modules.md
157:### internal/factory
246:> 실측 명령: … `internal/factory` 하나만 …

$ grep -n -i factory .moai/project/codemaps/modules.md
157: … 158: … 161: … 246: …
```

Lines **158** and **161** carry `Factory 모드` (the mode phrase in Korean) and the path `internal/cli/factory.go`. Neither matches the pattern: `[Ff]actory [Mm]ode` requires the English phrase, and `internal/factory` does not match `internal/cli/factory.go`.

**What this establishes.** At **file** granularity the completion criterion is sound — lines 157 and 246 hold the file in the match set until they are fixed, so `AC-KR-021` cannot reach zero with the file untouched. At **line** granularity it is not: an implementer who edits only the two matching lines drives `AC-KR-028`'s token count to zero while lines 158 and 161 still name a deleted package and a dead flag. `plan.md` M2b enumerates all four lines explicitly, and `AC-KR-028`'s first command is a human-read grep for the renamed forms; the residual gap was that the mechanical half of `AC-KR-028` would not catch a partial edit.

**Closed at v0.3.0.** The measurement that closes it is a bare-word grep bounded to the two files:

```
$ grep -niI factory .moai/project/codemaps/modules.md .moai/project/structure.md | wc -l
       5
$ grep -cniI factory .moai/project/codemaps/modules.md .moai/project/structure.md
.moai/project/codemaps/modules.md:4
.moai/project/structure.md:1
```

All five matches are Kanban Mode — none is generic pattern vocabulary — so at this scope the bare word carries no false-positive population and is a sound instrument. That is the fact the v0.2.0 disposition missed: it declined to close the gap on a tree-scale objection (the ~110 files of §D.1) applied to a two-file scope where that population does not exist. `AC-KR-028` now carries the bounded grep as a third command, with `5` as its measured positive control and `0` as its target. No criterion and no requirement was added to close it. `design.md` §F.3 carries the same reversal as a design consequence.

---

## §I. Scale figures for the verification sweep

```
$ go list ./... | wc -l
     115
```

**What this establishes.** A full-suite run produces on the order of 115 package result lines, so `tail -20` shows under a fifth of it. This is the measured reason `AC-KR-011` and `AC-KR-019` assert a `^FAIL` count against the **whole** log rather than against a tail, and the reason `plan.md` §B-6 records the bounded-tail hazard alongside the read-`$?`-after-a-pipe hazard: the two failure modes compound, and a red suite piped to `tail` reports `exit=0` with the evidence of failure scrolled off.

---

## §J. The catalog

```
$ grep -c 'workflows/' internal/template/catalog.yaml
0
$ grep -ci factory internal/template/catalog.yaml
0
```

**What this establishes.** Both zeros are true and both are **misleading**. The catalog indexes skill **directories** with a content hash, not files by path, so no `workflows/` path and no `factory` token appears in it — and yet renaming `factory.md` inside `.claude/skills/moai/` changes the `moai` skill directory's hash. An implementer who greps the catalog and concludes it is unaffected skips `make build`, leaves the committed tree stale, and gets no local signal at all; the failure surfaces as a CI parity failure later.

This is why `AC-KR-020` asserts the hash **changed** rather than that the build ran, and why `plan.md` §B-3 records the misreading as a named hazard.

---

## §K. Out of Scope

### Out of Scope — what this file does not measure

- The runtime behavior of the renamed surface. Nothing is renamed yet; every measurement above is of the repository at HEAD `d39e3cdc6`, not of the rename's result.
- The correctness of the pre-rename Factory Mode implementation. `SPEC-FACTORY-MODE-001` is a closed record and this SPEC does not re-audit it; §G measures only where its identifiers live.
- Anything under `.moai/specs/`. Excluded from every grep scope by design (`spec.md` §C), so no census of it is taken.

### Out of Scope — measurements deferred to run-phase

- The six-pair mirror classification. It is **time-varying** (§C) and is re-captured at M0 as the delta baseline, never read from this document.
- The `-k` probe against the run-phase tree. §E records the answer on `claude` `2.1.226`; M0 re-confirms it against whatever version is present then.
- The post-rename value of every criterion's command. This file records baselines and positive controls only; `progress.md` §E.2 carries the run-phase evidence.

---

## §L. Cross-references

- `spec.md` §A.2 … §A.6, §D.1 — the requirements and verification surfaces these measurements support.
- `design.md` §B, §C, §E, §F — the decisions each measurement forced.
- `plan.md` §A (the baseline table), §B (B-1 … B-6), §F M0 — the run-phase re-measurement commands.
- `acceptance.md` AC-KR-004, AC-KR-013, AC-KR-015, AC-KR-017, AC-KR-020, AC-KR-021, AC-KR-022, AC-KR-026, AC-KR-028 — the criteria that consume them.
