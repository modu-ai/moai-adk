# SPEC-AGENT-PROGRESS-PUSH-001 — Acceptance Criteria

**50 criteria covering 37 requirements. Every criterion is machine-verifiable, per-item, and carries an explicitly measured baseline.**

Baselines were measured during plan-phase (research.md §A). Re-verify them at run-phase entry before asserting any delta.

---

## §A Shared Verification Helpers

Every 9-agent criterion uses these helpers. They exist to eliminate two failure modes this repo has hit: **vacuous compound greps** (one file matching N times passes a `≥ N` check) and **`-A N` window truncation** (a windowed grep silently cuts a section short).

```bash
# H1 — exact section extraction. Terminates at the next H2, so there is NO window to truncate.
sec() { awk '/^## Progress Reporting Contract$/{f=1;next} f && /^## /{exit} f{print}' "$1"; }

# H2 — per-FILE predicate: how many of the 9 agents' sections match a pattern.
#      Counts FILES, not occurrences, so one file matching 9 times cannot pass.
seccount() { # $1 = agents dir, $2 = grep -E pattern
  n=0; for f in "$1"/*.md; do sec "$f" | grep -qE "$2" && n=$((n+1)); done; echo "$n"
}

# H3 — generic section extraction for an arbitrary H2 heading.
sect() { awk -v h="$2" '$0 == h {f=1;next} f && /^## /{exit} f{print}' "$1"; }

LIVE=.claude/agents/moai
TMPL=internal/template/templates/.claude/agents/moai
SSOT=.claude/rules/moai/workflow/progress-reporting-protocol.md
ACP=.claude/rules/moai/core/agent-common-protocol.md
WTI=.claude/rules/moai/workflow/worktree-integration.md
Z=.claude/rules/moai/core/zone-registry.md
```

**File-count guard — prerequisite for every 9/9 assertion (proves 9/9, not 8/9):**

```bash
ls -1 $LIVE/*.md | wc -l   # MUST be 9
ls -1 $TMPL/*.md | wc -l   # MUST be 9
```

---

## §B Given-When-Then Scenarios

### Scenario 1 — Agent reports progress on both channels

- **Given** the orchestrator delegates to `sync-auditor`, whose contract declares 3 milestones
- **When** the agent starts, and then completes its second milestone
- **Then** the 3 milestones were registered on the shared task list at start via `TaskCreate`; milestone 2 is marked via `TaskUpdate`; and a `SendMessage({to: "main", ...})` push carrying `[2/3]` surfaces at the orchestrator's next tool-call boundary and is relayed to the user in `conversation_language` naming the emitter

### Scenario 2 — The undocumented channel breaks

- **Given** a future runtime removes the `"main"` recipient
- **When** an agent reaches a milestone boundary
- **Then** the `SendMessage` push fails, the agent continues and completes normally with no retry and no error surfaced, and the milestone is still visible to the orchestrator on the shared task list. Progress reporting degrades from *immediate* to *pull-visible* — it does not disappear

### Scenario 3 — Agent needs user input

- **Given** a delegated agent discovers a missing parameter it cannot proceed without
- **When** it reaches that point
- **Then** it returns a structured blocker report — and does **not** attempt to obtain the input through either progress channel. (The user-question tool is unavailable to subagents at the platform level regardless of the `tools:` field, so the blocker report is the only path.)

### Scenario 4 — Orchestrator does not idle after a background spawn

- **Given** the orchestrator has spawned a background agent
- **When** its current unit of work completes
- **Then** it continues independent **read-only** work rather than ending the turn, so queued pushes have a boundary to drain at — and it performs no write while a write-capable agent is in flight

### Scenario 5 — Long delegation is pre-narrated

- **Given** the orchestrator is about to delegate to an agent whose contract declares `N ≥ 3`
- **When** the delegation is issued
- **Then** a NOW / NEXT / LATER / GATE roadmap was emitted first, in `conversation_language`, and GATE names the next point at which the user will be asked

### Scenario 6 — A write-capable agent runs in the background

- **Given** the runtime chooses background execution for a write-capable agent (the v2.1.198 default)
- **When** the agent reaches a tool call needing permission
- **Then** the prompt surfaces in the main session naming the asking subagent, the write proceeds on approval, and no MoAI doctrine forbids the background write — while the concurrency safeguard still holds: no second write-capable agent is running, and the orchestrator's concurrent work is read-only

---

## §C Edge Cases

| Edge case | Required behavior |
|---|---|
| Agent finishes in one milestone | One push; no padding to reach `N` |
| Agent exceeds its declared milestone list | The cap of 6 still applies; excess milestones are folded, not pushed |
| `SendMessage` succeeds but `TaskUpdate` fails | The agent continues; the push already delivered. Neither channel's failure is work-stopping |
| Agent run is aborted mid-flight | A push creates no ledger obligation; it is fire-and-forget |
| Two agents run concurrently (both read-only) | Each push names its emitter, so relays are attributable. Permitted — the concurrency safeguard binds *write-capable* agents only |
| `conversation_language` not in the locale table | Roadmap markers fall back to English labels; surrounding prose renders in the configured language |

---

## §D Acceptance Criteria Matrix

### D.1 — Layer 1: tool declarations (dual channel)

**AC-APP-001a** — `SendMessage` declared in all 9 **live** agent `tools:` CSVs.
Baseline: **0 of 9**. Target: **9 of 9**.
```bash
grep -c '^tools:.*SendMessage' $LIVE/*.md | grep -c ':1$'    # before: 0   after: 9
```
The `:1$` anchor requires exactly one match **per file**; with the file-count guard (§A) this cannot pass at 8/9.

**AC-APP-001b** — `SendMessage` declared in all 9 **template** agent `tools:` CSVs.
Baseline: **0 of 9**. Target: **9 of 9**.
```bash
grep -c '^tools:.*SendMessage' $TMPL/*.md | grep -c ':1$'    # before: 0   after: 9
```

**AC-APP-002a** — all four `Task*` tools declared in all 9 **live** agents.
Baseline: **5 of 9** (absent from `manager-design`, `plan-auditor`, `super-advisor`, `sync-auditor`). Target: **9 of 9**.
```bash
n=0; for f in $LIVE/*.md; do
  t=$(grep -m1 '^tools:' "$f"); ok=1
  for k in TaskCreate TaskUpdate TaskList TaskGet; do echo "$t" | grep -q "$k" || ok=0; done
  [ $ok -eq 1 ] && n=$((n+1))
done; echo "$n"                                              # before: 5   after: 9
```
Per-file, per-token — all four must be present in the same file. A file with only `TaskList` does not count.

**AC-APP-002b** — same in all 9 **template** agents.
Baseline: **5 of 9**. Target: **9 of 9**. (Same loop against `$TMPL`.)

**AC-APP-002c** — the 4 previously-lacking agents specifically now carry all four `Task*` tools, in both trees.
```bash
for d in $LIVE $TMPL; do for b in manager-design plan-auditor super-advisor sync-auditor; do
  t=$(grep -m1 '^tools:' "$d/$b.md")
  for k in TaskCreate TaskUpdate TaskList TaskGet; do
    echo "$t" | grep -q "$k" || echo "MISSING $k in $d/$b.md"
  done
done; done                                                   # MUST produce no output
```
Named explicitly so the sweep cannot silently skip the 4 that matter most.

**AC-APP-002d** — `tools:` remains a CSV string in all 18 files (no YAML array, no whitespace-only separation).
```bash
grep -h '^tools:' $LIVE/*.md $TMPL/*.md | grep -c '^tools: \['     # MUST be 0
go test ./internal/template/ -run TestAgentsFrontmatter_ToolsCSVFormat   # MUST PASS
```

### D.2 — Layer 1: contract section

**AC-APP-003a** — `## Progress Reporting Contract` present in all 9 **live** bodies.
Baseline: **0**. Target: **9**.
```bash
grep -c '^## Progress Reporting Contract$' $LIVE/*.md | grep -c ':1$'   # before: 0   after: 9
```

**AC-APP-003b** — same in all 9 **template** bodies. Baseline **0**, target **9**. (Same against `$TMPL`.)

**AC-APP-003c** — every section names its milestone count `N` and points to the SSOT by path.
```bash
seccount $LIVE 'N = [1-6]'                        # MUST be 9
seccount $LIVE 'progress-reporting-protocol\.md'  # MUST be 9
```

**AC-APP-004** — every section instructs the agent to register milestones on the shared task list (primary channel).
```bash
seccount $LIVE 'TaskCreate'   # MUST be 9
seccount $LIVE 'TaskUpdate'   # MUST be 9
```

**AC-APP-005** — every section carries the `SendMessage` call form with the `to: "main"` recipient (secondary channel).
```bash
seccount $LIVE 'SendMessage\('   # MUST be 9
seccount $LIVE 'to: "main"'      # MUST be 9
```

**AC-APP-010** — every section carries the `[n/N]` counter token and states the 2-line limit.
```bash
seccount $LIVE '\[n/N\]'                       # MUST be 9
seccount $LIVE 'Two lines|two lines|2 lines'   # MUST be 9
```

### D.3 — Layer 1: honest provenance (the undocumented channel)

**AC-APP-006** — the SSOT states the channel is undocumented, records the verified runtime version, and states it may break.
```bash
grep -cE 'undocumented' $SSOT                    # MUST be >= 1
grep -cE 'v?2\.1\.[0-9]+' $SSOT                  # MUST be >= 1  (a concrete version, not a placeholder)
grep -cE 'without notice|may break|may stop'     $SSOT   # MUST be >= 1
```

**AC-APP-006b** — every agent section also warns that `to: "main"` is undocumented (an agent reading only its own body must know).
```bash
seccount $LIVE 'undocumented'   # MUST be 9
```

**AC-APP-007** — no text anywhere implies `to: "main"` is official or documented.
Baseline (measured against the live tree — `$SSOT` not yet created at plan-phase): **0** in each direction.
```bash
# Forward direction — word-boundary anchored (\b) so `documented` inside `undocumented` cannot false-positive.
grep -rniE '\b(officially|documented|sanctioned|supported)\b[^.]{0,40}to: "main"' $SSOT $LIVE/ $TMPL/ $ACP CLAUDE.md | wc -l
# MUST be 0
# Reverse direction — the `is ` prefix already excludes `is undocumented`.
grep -rniE 'to: "main"[^.]{0,40}(is (officially|documented|sanctioned|supported))' $SSOT $LIVE/ $TMPL/ $ACP CLAUDE.md | wc -l
# MUST be 0
```
Both directions checked, since the claim could be phrased either way round.

**Regex false-positive guard (D1 fix).** REQ-APP-006/006b REQUIRE every agent section to describe `to: "main"` as *"undocumented"* — so the mandated honesty word would trip an un-anchored `documented` alternative. The forward pattern is `\b`-anchored precisely so it matches whole-word `documented` but NOT the substring inside `undocumented` (no word boundary falls before `documented` inside `undocumented`). Verified against both the live-tree `command grep` (BSD) and the ugrep wrapper:
```bash
# 0 — the mandated honesty text does NOT false-positive:
echo 'the undocumented `to: "main"` recipient' | grep -cE '\b(officially|documented|sanctioned|supported)\b[^.]{0,40}to: "main"'   # 0
# >=1 — a genuine violation is STILL caught:
echo 'officially supported `to: "main"`'       | grep -cE '\b(officially|documented|sanctioned|supported)\b[^.]{0,40}to: "main"'   # 1
```
The `\b` construct is portable across BSD grep / GNU grep / ugrep and is already relied on by AC-APP-015 in this same file.

**AC-APP-008** — best-effort degradation is stated in the SSOT and in every section, and names the fallback.
```bash
grep -cE '[Bb]est-effort' $SSOT                                  # MUST be >= 1
seccount $LIVE '[Bb]est-effort'                                  # MUST be 9
seccount $LIVE 'keep working|continue|still carries|task list'   # MUST be 9
```

### D.4 — Layer 1: boundary preservation (must-pass — cannot be weakened)

**AC-APP-009a** — no agent body invokes the user-question tool in call form; the existing subagent guard still passes.
Baseline: **0 files with a call form**. Target: **still 0**.
```bash
grep -rc 'AskUserQuestion(' $LIVE/ $TMPL/ | grep -vc ':0$'          # MUST be 0
go test ./internal/template/ -run TestNoAskUserQuestionInSubagents  # MUST PASS
```

**AC-APP-009b** — every one of the 9 sections restates the no-question prohibition **in prose**.
```bash
seccount $LIVE 'never a question|MUST NOT ask the user'   # MUST be 9
```

**AC-APP-014** — every section directs the agent to the blocker-report path for user input.
```bash
seccount $LIVE 'blocker report'   # MUST be 9
```

**AC-APP-025** — the SSOT cites the platform-level backing: the user-question tool is unavailable to subagents even when listed in `tools`.
```bash
grep -cE 'even when listed|unavailable to subagents|not available to subagents' $SSOT   # MUST be >= 1
grep -c 'tools' $SSOT                                                                    # MUST be >= 1
```

### D.5 — Layer 1: noise control

**AC-APP-011** — the cap of 6 is stated in the SSOT, and no per-agent `N` exceeds it.
```bash
grep -cE 'at most 6|maximum of 6|cap of 6' $SSOT   # MUST be >= 1
seccount $LIVE 'N = [7-9]'                         # MUST be 0  (no agent over the cap)
seccount $LIVE 'N = [1-6]'                         # MUST be 9  (every agent within it)
```

**AC-APP-012** — the milestone-only prohibition is stated in the SSOT and in every section.
```bash
grep -cE '[Mm]ilestone-only' $SSOT       # MUST be >= 1
seccount $LIVE '[Mm]ilestone-only'       # MUST be 9
```

**AC-APP-013** — the English-body / orchestrator-relays rule is stated in the SSOT and in every section.
```bash
grep -c 'English' $SSOT      # MUST be >= 1
seccount $LIVE 'English'     # MUST be 9
```

### D.6 — Layer 2: orchestrator obligations (stated in the SSOT)

**AC-APP-015** — roadmap, its four markers, the locale table, and the `N ≥ 3` trigger are all specified.
```bash
for m in NOW NEXT LATER GATE; do grep -qE "\b$m\b" $SSOT || echo "MISSING marker $m"; done   # no output
grep -c '지금' $SSOT                          # MUST be >= 1   (locale table present)
grep -cE 'N *(>=|≥) *3|3 or more'  $SSOT      # MUST be >= 1   (the trigger, not a new threshold)
```

**AC-APP-016** — the relay obligation names `conversation_language`.
```bash
grep -cE 'relay' $SSOT                  # MUST be >= 1
grep -c 'conversation_language' $SSOT   # MUST be >= 1
```

**AC-APP-017** — the SSOT names `TaskList` as the durable progress view.
```bash
grep -c 'TaskList' $SSOT                        # MUST be >= 1
grep -cE 'durable|primary|correctness' $SSOT    # MUST be >= 1
```

**AC-APP-018** — the non-idle obligation is stated, with the read-only constraint.
```bash
grep -cE 'not.{0,20}idle|shall not end its turn' $SSOT   # MUST be >= 1
grep -cE 'read-only' $SSOT                                # MUST be >= 1
```

**AC-APP-019** — the polling prohibition names both closed paths.
```bash
grep -cE 'transcript' $SSOT    # MUST be >= 1
grep -c 'TaskOutput' $SSOT     # MUST be >= 1
```

**AC-APP-020** — the named-spawn hazard is documented.
```bash
grep -cE 'name:' $SSOT                                # MUST be >= 1
grep -cE 'team file|team-runtime|team runtime' $SSOT  # MUST be >= 1
```

### D.7 — Layer 3: doctrine and registration (reachability, not mere existence)

**AC-APP-021** — the SSOT exists in both trees and is byte-identical.
Baseline: **absent in both**.
```bash
test -f $SSOT && echo OK
test -f internal/template/templates/$SSOT && echo OK
diff $SSOT internal/template/templates/$SSOT && echo BYTE-IDENTICAL   # MUST produce no diff output
```

**AC-APP-022** — the SSOT pointer is **inside** `## Background Agent Execution` of `agent-common-protocol.md`, not merely somewhere in the file.
Baseline: **0**.
```bash
sect $ACP '## Background Agent Execution' | grep -c 'progress-reporting-protocol.md'
# before: 0   after: >= 1
```
Section-scoped (terminates at the next H2) — no window to truncate.

**AC-APP-023** — the SSOT pointer is **inside** `## 14. Parallel Execution Safeguards` of `CLAUDE.md`.
Baseline: **0**.
```bash
sect CLAUDE.md '## 14. Parallel Execution Safeguards' | grep -c 'progress-reporting-protocol.md'
# before: 0   after: >= 1
```

**AC-APP-024** — five new constitution clauses registered, each complete, with the boundary clause Frozen and canary-gated.
Baseline: `grep -c '^- id: CONST-V3R6-00[2-6]$'` = **0**.
```bash
grep -c '^- id: CONST-V3R6-00[2-6]$' $Z                    # MUST be 5

# every entry carries all 6 fields (each entry is exactly 7 lines: id + 6 fields)
for id in 002 003 004 005 006; do
  blk=$(grep -A 6 "^- id: CONST-V3R6-$id\$" $Z)
  for k in zone zone_class file anchor clause canary_gate; do
    echo "$blk" | grep -qE "^  $k:" || echo "MISSING $k in CONST-V3R6-$id"
  done
done                                                       # MUST produce no output

# the question-boundary clause is Frozen + canary-gated
grep -A 6 '^- id: CONST-V3R6-002$' $Z | grep -c 'zone: Frozen'       # MUST be 1
grep -A 6 '^- id: CONST-V3R6-002$' $Z | grep -c 'canary_gate: true'  # MUST be 1
```
The `-A 6` window is exact by construction (7-line entries) and was dry-run against the existing `CONST-V3R6-001` entry.

### D.8 — Layer 4: background-default realignment

**AC-APP-026** — no doctrine surface still asserts the superseded write restriction.
Baseline: the phrase appears in `agent-common-protocol.md`, `worktree-integration.md` (both trees), and `zone-registry.md`.
```bash
grep -rn 'MUST NOT perform Write/Edit' \
  CLAUDE.md $ACP $WTI $Z \
  internal/template/templates/CLAUDE.md \
  internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md \
  internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md \
  | wc -l
# MUST be 0
```
A single absence-grep across every surface named in research.md §D.2 — if any one is missed, this fails.

**AC-APP-026b** — the CLAUDE.md §14 background-default assertion, which AC-APP-026 does **not** reach, is realigned in both trees.
Baseline (measured): **1** in live `CLAUDE.md`, **1** in the template mirror.
```bash
grep -cE 'run_in_background: false.*conservative default' CLAUDE.md                              # before: 1   after: 0
grep -cE 'run_in_background: false.*conservative default' internal/template/templates/CLAUDE.md  # before: 1   after: 0
```
Per-surface and discriminating — it targets §14's actual phrase (`run_in_background: false … conservative default`, present at `CLAUDE.md` line 239), NOT a vacuous compound `A|B` grep. Without this AC the §14 realignment (REQ-APP-027 / REQ-APP-030) is **ungated**: `MUST NOT perform Write/Edit` never appears in `CLAUDE.md` (measured 0 in both trees), so AC-APP-026 passes at 0 whether or not §14 is ever touched. The post-implementation target is **0** — the superseded assertion is removed or realigned to the documented v2.1.198 background default.

**AC-APP-027** — the alignment decision and its reasoning are stated, and the runtime version is named.
```bash
grep -cE '2\.1\.198' $ACP                        # MUST be >= 1
grep -cE '2\.1\.198' CLAUDE.md                   # MUST be >= 1
grep -ciE 'default|align' $ACP                   # MUST be >= 1
```
Baseline for the version token in both files: **0** (measured — no doctrine surface mentions v2.1.198).

**AC-APP-028** — the replacement safeguard is stated on every surface that previously carried the write restriction.
```bash
for f in CLAUDE.md $ACP $SSOT; do
  grep -qiE 'two write-capable|concurrent write|one writer' "$f" || echo "MISSING concurrency safeguard in $f"
done                                             # MUST produce no output
grep -ciE 'read-only' $ACP                       # MUST be >= 1
```
Removing a fence without putting one up is the failure this AC prevents.

**AC-APP-029** — the two registry clauses are amended, not stale.
```bash
grep -A 6 '^- id: CONST-V3R2-020$' $Z | grep -c '2\.1\.198'                  # MUST be >= 1
grep -A 6 '^- id: CONST-V3R2-044$' $Z | grep -ci 'MUST NOT perform Write'    # MUST be 0
grep -A 6 '^- id: CONST-V3R2-044$' $Z | grep -ciE 'concurrent|read-only'     # MUST be >= 1
```

**AC-APP-029b** — the constitution amendment is mechanically permissible and regression-free (self-proof, not prose).
Baseline (measured at plan-phase against the live registry): guard **exit 0**; `validate` reports **77** DRIFT errors.
```bash
# (a) CONST-V3R2-020 / CONST-V3R2-044 are amendable (zone: Evolvable, not Frozen) —
#     MUST exit 0 BOTH before and after the amendment. A Frozen-zone marking would exit 1.
moai constitution guard --violations CONST-V3R2-020,CONST-V3R2-044; echo "exit=$?"   # exit=0 ("No Frozen zone violations")

# (b) the amendment introduces NO NEW drift beyond the pre-existing baseline of 77.
N=$(moai constitution validate 2>&1 | grep -oE 'found [0-9]+ error' | grep -oE '[0-9]+' | tail -1)
echo "validate errors = $N"                                                         # before: 77   after: <= 77
[ "${N:-999}" -le 77 ] && echo PASS || echo FAIL                                     # 78+ = FAIL (a new DRIFT was introduced)
```
Pre-existing baseline stated explicitly: **77** DRIFT errors, all unrelated to this SPEC (e.g. `CONST-V3R5-036..041`, `CONST-V3R6-001`). The five new `CONST-V3R6-002..006` entries (AC-APP-024) and the two amended `CONST-V3R2-020/044` entries (AC-APP-029) MUST each validate cleanly, keeping the total at 77 or lower; an increase to **78+** means the run-phase left a clause whose `clause:` text does not match its source anchor. This mechanically self-proves what C1 previously asserted only in prose.

**AC-APP-030** — the inline zone markers are reconciled with the registry, and the unregistered clause is resolved.
```bash
# the two registered clauses no longer carry a Frozen inline marker (registry says Evolvable)
sect $ACP '## Background Agent Execution' | grep -c 'ZONE:Frozen'   # MUST be 0
sect CLAUDE.md '## 14. Parallel Execution Safeguards' | grep -c 'ZONE:Frozen'   # MUST be 0

# worktree-integration.md's unregistered HARD clause is gone (reduced to a cross-reference)
grep -c 'ZONE:Frozen\] \[HARD\] (clause updated' $WTI   # MUST be 0
grep -c 'agent-common-protocol.md' $WTI                 # MUST be >= 1  (cross-reference present)
```
Baselines: 1 / 1 / 1 / (measure) respectively.

**AC-APP-031** — no MoAI agent sets the `background:` frontmatter field.
Baseline: **0**. Target: **still 0**.
```bash
grep -lc '^background:' $LIVE/*.md $TMPL/*.md 2>/dev/null | wc -l   # MUST be 0
```

### D.9 — Layer 5: template mirroring and guards

**AC-APP-032** — every changed distributed file has a mirror; the live-only file does **not** gain one.
```bash
for p in \
  ".claude/rules/moai/workflow/progress-reporting-protocol.md" \
  ".claude/rules/moai/core/agent-common-protocol.md" \
  ".claude/rules/moai/workflow/worktree-integration.md" ; do
  test -f "internal/template/templates/$p" || echo "MISSING MIRROR $p"
done
test -f internal/template/templates/CLAUDE.md || echo "MISSING MIRROR CLAUDE.md"
# MUST produce no output

# zone-registry.md is intentionally live-only — a mirror is a DEFECT, not a gap
test -f internal/template/templates/.claude/rules/moai/core/zone-registry.md \
  && echo "DEFECT: zone-registry mirror created" || echo "OK: live-only preserved"
```

**AC-APP-033** — the lines this SPEC **adds** to the template tree contain zero internal identifiers.
```bash
git diff --unified=0 -- internal/template/templates/ | grep '^+' | grep -v '^+++' \
  | grep -cE 'SPEC-[A-Z0-9]+(-[A-Z0-9]+)*-[0-9]{3}|REQ-APP-|AC-APP-|CONST-V3R6-|[0-9]{4}-[0-9]{2}-[0-9]{2}|\b[0-9a-f]{9,40}\b'
# MUST be 0
go test ./internal/template/ -run 'TestTemplateNoInternalContentLeak'   # MUST PASS
```
Diff-scoped, so it cannot inherit or be masked by pre-existing template state.

**AC-APP-034** — the SSOT is enrolled in the byte-parity mirror-CI allowlist.
Baseline: **0**.
```bash
grep -c 'progress-reporting-protocol.md' internal/template/rule_template_mirror_test.go   # before: 0   after: >= 1
go test ./internal/template/ -run TestRuleTemplateMirror                                   # MUST PASS
```
Without this, a future single-tree edit ships stale — the cross-file-reachability failure mode.

**AC-APP-035a** — the 6 currently byte-identical agents remain byte-identical.
Baseline (measured): `manager-design`, `manager-develop`, `manager-docs`, `manager-git`, `plan-auditor`, `super-advisor`.
```bash
for b in manager-design manager-develop manager-docs manager-git plan-auditor super-advisor; do
  diff -q "$LIVE/$b.md" "$TMPL/$b.md" >/dev/null || echo "PARITY BROKEN: $b"
done                                             # MUST produce no output
```

**AC-APP-035b** — the 3 sanitized agents retain **exactly** their pre-existing divergence.
Baseline (measured diff line counts): `builder-harness` = 8, `manager-spec` = 4, `sync-auditor` = 4.
```bash
for pair in builder-harness:8 manager-spec:4 sync-auditor:4; do
  b=${pair%%:*}; want=${pair##*:}
  got=$(diff "$LIVE/$b.md" "$TMPL/$b.md" | wc -l | tr -d ' ')
  [ "$got" = "$want" ] || echo "DIVERGENCE CHANGED: $b want=$want got=$got"
done                                             # MUST produce no output
```
A larger count means the edit hit only one tree, or carried an identifier. A smaller count means an existing sanitization was clobbered.

**AC-APP-035c** — `worktree-integration.md` retains its sanitized divergence (not collapsed to identical, not widened).
Baseline (measured): **20 diff lines**.
```bash
got=$(diff $WTI internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md | wc -l | tr -d ' ')
echo "$got"    # MUST be 20
```

**AC-APP-036** — build regenerates and the guard suites are green.
```bash
make build; echo "exit=$?"                       # MUST be exit=0
go test ./internal/template/...                  # MUST PASS
go test ./...                                    # MUST PASS (no new failure vs. baseline)
```

### D.10 — Layer 6: standing regression check

**AC-APP-037a** — the SSOT records the Claude Code version the undocumented channel was verified against, as a concrete version.
```bash
grep -cE 'v?2\.1\.[0-9]{3}' $SSOT   # MUST be >= 1
```
A placeholder or a version range fails this — the point is that a future reader can tell whether the verification is stale.

**AC-APP-037b** — the channel regression check is **executed and recorded as an observation**, for both spawn modes.
```bash
grep -ciE 'foreground|background' .moai/specs/SPEC-AGENT-PROGRESS-PUSH-001/progress.md   # MUST be >= 1
```
Manual gate (the load-bearing half): `progress.md` §E.2 carries, **verbatim**, (a) the observed Claude Code version, (b) the `SendMessage` return value observed from a **background** spawn, and (c) the return value **and the observed delivery timing** from a **foreground** spawn.

Item (c) is a genuine open measurement. The plan-phase probe established that the foreground call *succeeds*; it did **not** establish *when the user sees the message*. Because delivery is "queued for the main conversation's next turn" and a foreground orchestrator is blocked (issuing no tool calls) until the agent returns, the message may only surface at return — which would deliver no interim value. Run-phase must **measure** this, not infer it. An inferred or assumed answer fails this criterion.

---

## §E Traceability — every requirement is covered

| REQ | Covered by |
|---|---|
| REQ-APP-001 | AC-APP-001a, AC-APP-001b |
| REQ-APP-002 | AC-APP-002a, AC-APP-002b, AC-APP-002c, AC-APP-002d |
| REQ-APP-003 | AC-APP-003a, AC-APP-003b, AC-APP-003c |
| REQ-APP-004 | AC-APP-004 |
| REQ-APP-005 | AC-APP-005 |
| REQ-APP-006 | AC-APP-006, AC-APP-006b |
| REQ-APP-007 | AC-APP-007 |
| REQ-APP-008 | AC-APP-008 |
| REQ-APP-009 | AC-APP-009a, AC-APP-009b |
| REQ-APP-010 | AC-APP-010 |
| REQ-APP-011 | AC-APP-011 |
| REQ-APP-012 | AC-APP-012 |
| REQ-APP-013 | AC-APP-013 |
| REQ-APP-014 | AC-APP-014 |
| REQ-APP-015 | AC-APP-015 |
| REQ-APP-016 | AC-APP-016 |
| REQ-APP-017 | AC-APP-017 |
| REQ-APP-018 | AC-APP-018 |
| REQ-APP-019 | AC-APP-019 |
| REQ-APP-020 | AC-APP-020 |
| REQ-APP-021 | AC-APP-021 |
| REQ-APP-022 | AC-APP-022 |
| REQ-APP-023 | AC-APP-023 |
| REQ-APP-024 | AC-APP-024 |
| REQ-APP-025 | AC-APP-025 |
| REQ-APP-026 | AC-APP-026, AC-APP-026b |
| REQ-APP-027 | AC-APP-027 |
| REQ-APP-028 | AC-APP-028 |
| REQ-APP-029 | AC-APP-029, AC-APP-029b |
| REQ-APP-030 | AC-APP-030 |
| REQ-APP-031 | AC-APP-031 |
| REQ-APP-032 | AC-APP-032 |
| REQ-APP-033 | AC-APP-033 |
| REQ-APP-034 | AC-APP-034 |
| REQ-APP-035 | AC-APP-035a, AC-APP-035b, AC-APP-035c |
| REQ-APP-036 | AC-APP-036 |
| REQ-APP-037 | AC-APP-037a, AC-APP-037b |

**Coverage: 37 / 37 requirements. 50 criteria. Zero orphan criteria (every AC maps to a REQ).**

---

## §F Severity and Closure Gates

| Severity | Criteria | Rule |
|---|---|---|
| **Must-pass (blocking)** | AC-APP-009a, AC-APP-009b, AC-APP-024 (Frozen/canary fields), AC-APP-028, AC-APP-033, AC-APP-035a, AC-APP-035b, AC-APP-035c, AC-APP-036 | A failure here means the question boundary was eroded, a safety fence was removed without a replacement, the template leaked, mirror parity broke, or the build is red. No score compensates; close is blocked |
| **Core** | AC-APP-001a/b, AC-APP-002a/b/c/d, AC-APP-003a/b/c, AC-APP-004, AC-APP-005, AC-APP-021, AC-APP-022, AC-APP-023, AC-APP-026, AC-APP-026b, AC-APP-029, AC-APP-029b, AC-APP-030, AC-APP-034, AC-APP-037b | The headline capability, its reachability, and the realignment. All must pass to close |
| **Contract quality** | AC-APP-006, AC-APP-006b, AC-APP-007, AC-APP-008, AC-APP-010 .. AC-APP-020, AC-APP-025, AC-APP-027, AC-APP-031, AC-APP-032, AC-APP-037a | The bounding and honesty clauses. A single miss may be closed as documented debt only with an explicit follow-up entry |

### Definition of Done

- [ ] All 50 criteria executed with verbatim command output captured in `progress.md` §E.2
- [ ] Zero must-pass failures
- [ ] `make build` exit 0; `go test ./...` green with no new failure
- [ ] The channel regression check recorded as an **observation** for both spawn modes, including foreground delivery *timing* (AC-APP-037b)
- [ ] Parity table re-measured and unchanged (AC-APP-035a/b/c)
- [ ] Zero `[NEEDS CLARIFICATION]` markers (confirmed at plan-phase; re-confirm none were introduced)

### Accepted Residual Risks (recorded, not blocking)

Coverage limits the plan-auditor surfaced for which no cheap mechanical fix exists. They are accepted as residual risk and recorded here so run-phase and close do not mistake them for gaps.

- **D4 — AC-APP-009a call-form guard scope.** AC-APP-009a's `AskUserQuestion(`-with-paren grep protects only the literal (platform-impossible) tool call; it does **not** detect a question smuggled through `SendMessage` prose. The prose-question prohibition on the `SendMessage` channel is covered only by the prose ACs (AC-APP-009b restates the no-question rule in every section; REQ-APP-009). No runtime hook can inspect a `SendMessage` body, so mechanical enforcement of "no question via SendMessage prose" is infeasible — accepted residual risk.
- **D5 — AC-APP-034 string-presence vs reachability.** AC-APP-034's `grep 'progress-reporting-protocol.md' … rule_template_mirror_test.go` proves the SSOT string was **added** to the mirror allowlist, but the real reachability guard is the Go test `TestRuleTemplateMirror` (also asserted in AC-APP-034). The grep alone is string-presence; the Go test is what actually fails a future single-tree edit. The two are asserted together in AC-APP-034 by design — the grep is a fast pre-check, the Go test is the load-bearing guard. Recorded so the grep is never read as the sole reachability proof.

### Forward-looking checks (recorded at close, not blocking)

- Does an agent run surface pushes to the user end-to-end, relayed in `conversation_language`?
- Is the cap of 6 empirically right, or does `manager-develop` feel under- or over-chatty? (input to the deferred config knob)
- After the background-default realignment, does the runtime actually background write-capable agents — and do the permission prompts behave as documented?
