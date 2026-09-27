# Acceptance — SPEC-SESSION-MIDMOVE-001

Verification layer. Every criterion is Given-When-Then and binary. Commands run from the worktree root. Criteria are evaluated before the card branch is integrated into develop (after integration `CARD_BASE` equals HEAD and AC-SMM-021 turns red by design).

## Variables (set in the same invocation as each check)

```bash
KT=internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md
KL=.claude/rules/moai/workflow/kanban-dispatch.md
DT=internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch-detail.md
DL=.claude/rules/moai/workflow/kanban-dispatch-detail.md
WT=internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md
WL=.claude/rules/moai/workflow/worktree-integration.md
AT=internal/template/templates/AGENTS.md.tmpl
AL=AGENTS.md
LP=.claude/rules/local/gitflow-lane-protocol.md
CL=CLAUDE.local.md
R=.moai/reports/t1279
E=$R/m1-measure.md
```

`CARD_BASE` is read at check time as the output of `git merge-base develop HEAD`. In a worktree-isolated session the guard refuses command substitution around git, so run `git merge-base develop HEAD` alone, assign the printed SHA literally (`CARD_BASE=<sha>`), and verify it with `git rev-parse --verify "$CARD_BASE^{commit}"` (exit 0). `CAPS_COMMIT` is read the same way from `git log --format=%H --diff-filter=A -- "$R/m1-caps.md"`.

Canonical phrases the doctrine must use verbatim (so checks are fixed-string):

| Id | Phrase |
|---|---|
| P-FLOW | ``move → `/clear` → re-send`` |
| P-HEAD | `### Card change in a standing session` |
| P-RESEND | `the lead re-sends` |
| P-MOVE | `adds the moved-into tree's skill listing` |
| P-LAUNCH | `gives one skill listing` |
| P-CLEAR | `leaves only the new tree's listing` |
| P-RELAUNCH | `relaunching it through the launcher` |
| P-EXEMPT | `mid-session move exemption` |

## §D — AC Matrix

### Range sanity

- **AC-SMM-021** (REQ-SMM-014, REQ-SMM-015, REQ-SMM-016, REQ-SMM-017) — Given `CARD_BASE` is read as above, When `git rev-parse --verify "$CARD_BASE^{commit}"` and `git diff --name-only "$CARD_BASE"..HEAD | wc -l` run, Then the first exits `0` and the second prints at least `1`. Every range-based criterion below is FAIL (not PASS) while this one is FAIL.

### Fixture mechanism check (M1)

- **AC-SMM-001** (REQ-SMM-001) — Given M1 has run, When `git show --name-only --format= "$CAPS_COMMIT"` runs, Then it prints exactly the one line `.moai/reports/t1279/m1-caps.md`; and When `git log -1 --format=%ct "$CAPS_COMMIT"` is compared with the `first_probe_epoch:` value of `$E`, Then the commit epoch is smaller; and `python3 -c 'import datetime,sys;print(int(datetime.datetime.fromisoformat(sys.argv[1].replace("Z","+00:00")).timestamp()))' "<first_probe_ts value from $E>"` prints the same integer as `first_probe_epoch:`; and `$R/m1-caps.md` contains the four strings `6`, `4 turns`, `timeout -k 10 300`, `45 minutes`.
- **AC-SMM-002** (REQ-SMM-002, REQ-SMM-012) — Given `$E` is committed, When `grep -E '^(fixture_trees|listing_count|listing_bytes|turn_tokens)_P[123]: ' "$E" | cut -d: -f1 | sort -u | wc -l` and `grep -cE '^(fixture_trees_P[123]: (primary|wt|primary,wt|gap)|listing_count_P[123]: ([0-9]+|gap)|(listing_bytes|turn_tokens)_P[123]: ([0-9]+(,[0-9]+)*|gap))$' "$E"` run, Then both print `12`; and `python3 -c 'import sys;d=dict(l.rstrip("\n").split(": ",1) for l in open(sys.argv[1]) if ": " in l);f=lambda t,one:"gap" if t=="gap" else ("yes" if one(t) else "no");ok=[f(d["fixture_trees_P1"],lambda t:"," not in t)==d["launcher_single_listing"],f(d["fixture_trees_P2"],lambda t:t=="primary,wt")==d["move_adds_listing"],f(d["fixture_trees_P3"],lambda t:"," not in t)==d["clear_restores_single_listing"]];print(sum(ok))' "$E"` prints `3`.
- **AC-SMM-003** (REQ-SMM-002, REQ-SMM-006) — Given `$E` is committed, When `grep -cE '^session_id_P[123]: [0-9a-f-]{36}$' "$E"` runs, Then it prints `3`; and When every `probe_cwd_P[123]:` value is compared with the `fixture_root:` value, Then `python3 -c 'import sys;d=dict(l.rstrip("\n").split(": ",1) for l in open(sys.argv[1]) if ": " in l);r=d["fixture_root"];print(sum(1 for p in "123" if not all(c.startswith(r) for c in d["probe_cwd_P"+p].split(","))))' "$E"` prints `0`; positive control: the same command on a copy of `$E` whose `probe_cwd_P1` is replaced by the repository root prints `1`.
- **AC-SMM-004** (REQ-SMM-003) — Given `$E` is committed, When `grep -c '^judge_source: transcript$' "$E"` runs, Then it prints `1`; and every non-comment line of `$R/extract.txt` that is a command starts with `python3 .moai/reports/t1279/extract_listing.py ` (`grep -E '^\$ ' "$R/extract.txt" | grep -vc '^\$ python3 .moai/reports/t1279/extract_listing.py '` prints `0`, and `grep -c '^\$ ' "$R/extract.txt"` prints at least `1`).
- **AC-SMM-005** (REQ-SMM-004) — Given `$R/m1-control.jsonl` and `$R/m1-control-neg.jsonl` are committed, When `python3 $R/extract_listing.py $R/m1-control.jsonl | grep -c 'scoped_names=[1-9]'` and `python3 $R/extract_listing.py $R/m1-control-neg.jsonl | grep -c 'scoped_names=[1-9]'` and `python3 $R/extract_listing.py $R/m1-control-neg.jsonl | grep -c 'scoped_names=0'` run, Then they print at least `1`, `0`, and at least `1`. This criterion is mandatory whenever a `fixture_trees_P*` value holds one tree or no path reports a scoped name.
- **AC-SMM-006** (REQ-SMM-005) — Given `$E` and `$R/probes.txt` are committed, When `python3 -c 'import sys;d=dict(l.rstrip("\n").split(": ",1) for l in open(sys.argv[1]) if ": " in l);print(sum(1 for k,v in d.items() if v=="gap" and not d.get("gap_reason_"+k,"").strip()))' "$E"` runs, Then it prints `0`; positive control: on a copy with one `gap` value and its reason line removed, it prints `1`. And `probes_run:` is between the count of distinct `session_id_P*` values and `6`; `wall_clock_min:` equals `(last_probe_epoch - first_probe_epoch) / 60` rounded up and is at most `45`; and every `# operator-run: session=<id>` line in `probes.txt` names an id that appears as a `session_id_P*` value.
- **AC-SMM-007** (REQ-SMM-006) — Given `$R/probes.txt` is committed, When `grep -v '^#' "$R/probes.txt" | grep -c .` and `grep -v '^#' "$R/probes.txt" | grep -vc '^unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && timeout -k 10 300 claude '` and `grep -cE '(^| )(--setting-sources|-d|--debug|--debug-file)( |=|$)' "$R/probes.txt"` run, Then they print at least `1`, `0`, `0`; positive control: `printf 'x claude -p y -d\nx claude --setting-sources user\n' | grep -cE '(^| )(--setting-sources|-d|--debug|--debug-file)( |=|$)'` prints `2`. And `$E` has `repo_head_before` equal to `repo_head_after` and `repo_branch_before` equal to `repo_branch_after`; and `git diff --name-only "$CARD_BASE"..HEAD -- .moai/reports/t1279/ | grep -E '\.jsonl$|debug' | grep -vcE '/m1-control(-neg)?\.jsonl$'` prints `0`.
- **AC-SMM-008** (REQ-SMM-007) — Given M1 is committed, When `test -s "$R/probes.txt" && test -s "$R/extract.txt" && test -s "$R/fixture.txt"` runs, Then it exits `0`; and `grep -c 'extract_listing.py' "$R/probes.txt"` and `grep -c 'claude -p' "$R/fixture.txt"` both print `0`.

### Real-session cost (M2)

- **AC-SMM-009** (REQ-SMM-008) — Given `$R/m1-cost.md` is committed, When `grep -cE '^row: ts=[0-9T:.-]+Z skillCount=[0-9]+ scoped_names=[1-9][0-9]* content_bytes=[0-9]+ before=[0-9]+ after=[0-9]+ delta=-?[0-9]+ other_row_bytes=[0-9]+ bound=upper$' "$R/m1-cost.md"` and `grep -cE '^cost: gap reason=.+' "$R/m1-cost.md"` run, Then exactly one of them is non-zero (at least `1` rows, or exactly `1` gap line); and `grep -c '^command: python3 .moai/reports/t1279/extract_listing.py --cost ' "$R/m1-cost.md"` prints `1`. When rows are present, Then a row with `ts=2026-09-27T03:57:49.175Z` exists (plan-time reading: `content_bytes=43566 scoped_names=42 delta=20829 other_row_bytes=20526`).

### Doctrine (M4)

- **AC-SMM-010** (REQ-SMM-009) — Given M4 is committed, When `grep -c 'moai cc -w <card-id>'` runs on each of `$KT`, `$DT`, `$WT`, `$AT`, `$AL`, `$CL`, Then each prints at least `1` (at `2370c5b31`: all `0`).
- **AC-SMM-011** (REQ-SMM-010) — Given M4 is committed, When `grep -cF 'move → `/clear` → re-send'` runs on each of `$DT`, `$AT`, `$LP`, `$CL`, and on the `/clear` section of `$KT` extracted with `awk '/^## The `\/clear` handoff between phases/{f=1;next} /^## /{f=0} f' "$KT"`, Then each prints at least `1` (at `2370c5b31`: all `0`); and `grep -cF '### Card change in a standing session' "$DT"` prints `1` (baseline `0`); and within that subsection (`awk '/^### Card change in a standing session/{f=1;next} /^##/{f=0} f' "$DT"`), `grep -cF 'the lead re-sends'` prints at least `1` and `grep -ciE 'relaunch[^.]{0,60}(optional|not required)'` prints at least `1`.
- **AC-SMM-012** (REQ-SMM-011) — Given M4 is committed, When `grep -c 'EnterWorktree(<card-id>)'` runs on `$KT` and `$LP`, Then both print `0` (baseline `2`, `2`); and `grep -cE 'moai cc -w <card-id>` (또는|or) .*EnterWorktree' "$LP"` prints `0` (baseline `1`); and the move-without-clear count
  `python3 -c 'import re,sys;ex=re.compile(r"release|integration|develop|통합|auto-names|Return the same way|병합을 마치고");b=0
for f in sys.argv[1:]:
  for p in re.split(r"\n\s*\n",open(f,encoding="utf-8").read()):
    rows=[l for l in p.split("\n") if l.startswith("|")];pr="\n".join(l for l in p.split("\n") if not l.startswith("|"))
    b+=sum(1 for u in rows+[pr] if "EnterWorktree(" in u and not ex.search(u) and "/clear" not in u)
print(b)' "$KT" "$DT" "$AT" "$CL" "$LP"`
  prints `0` (baseline `5` at `2370c5b31`, the positive control).
- **AC-SMM-013** (REQ-SMM-012) — Given `$E` and M4 are committed, When each of `move_adds_listing`, `launcher_single_listing`, `clear_restores_single_listing` is read from `$E`, Then for each key whose value is `no` or `gap`, `grep -cF` of its phrase (P-MOVE, P-LAUNCH, P-CLEAR respectively) summed over `$KT`, `$DT`, `$WT`, `$AT` prints `0`; and when `clear_restores_single_listing` is `no` or `gap`, `grep -cF 'relaunching it through the launcher' "$DT"` prints at least `1`.
- **AC-SMM-014** (REQ-SMM-013) — Given DP-2 is recorded in progress.md §E.2 as `dp2: exempt` or `dp2: covered`, When the section of `$KT` from `## Integration into the release branch is self-served` to the next `## ` is extracted with the same `awk` form as AC-SMM-011, Then: for `exempt`, `grep -cF 'mid-session move exemption'` prints at least `1` on that section, on `$LP`, and on `$CL` (baseline `0` each); for `covered`, `grep -cF 'move → `/clear` → re-send'` prints at least `1` on that section.

### Where the doctrine lives

- **AC-SMM-015** (REQ-SMM-014) — Given M4 is committed, When `cmp "$KT" "$KL"`, `cmp "$DT" "$DL"`, `cmp "$WT" "$WL"` run, Then each exits `0`; When `sed -n '/^## 3. Worktrees/,/^## 4\./p' "$AL" | wc -l` and the same on `$AT` run, Then both print more than `0`, and `diff <(sed -n '/^## 3. Worktrees/,/^## 4\./p' "$AL") <(sed -n '/^## 3. Worktrees/,/^## 4\./p' "$AT")` exits `0` (HEAD only; no range); and `comm -23 <(git log --no-merges --format=%H "$CARD_BASE"..HEAD -- "$KL" "$DL" "$WL" | sort) <(git log --no-merges --format=%H "$CARD_BASE"..HEAD -- "$KT" "$DT" "$WT" | sort) | wc -l` prints `0` while `git log --no-merges --format=%H "$CARD_BASE"..HEAD -- "$KT" | wc -l` prints at least `1`.
- **AC-SMM-016** (REQ-SMM-015) — Given `always_loaded_before: <N>` is recorded in `$E` before the first doctrine commit, When `go test ./internal/config/ -run '^TestCodexContractByteCeiling$|^TestAlwaysLoadedTokenBudget$' -count=1 -v` runs, Then it exits `0`, its output contains lines starting `--- PASS: TestAlwaysLoadedTokenBudget (` and `--- PASS: TestCodexContractByteCeiling (`, and the `always-loaded surface = <M> tokens` figure satisfies `M ≤ N`; and `git diff "$CARD_BASE"..HEAD -- internal/config/token_budget_guard.go | wc -l` prints `0` while `grep -c '^const AlwaysLoadedTokenBudget = ' internal/config/token_budget_guard.go` prints `1`.
- **AC-SMM-017** (REQ-SMM-016) — Given M4 is committed, When `git diff -U0 "$CARD_BASE"..HEAD -- "$KT" "$DT" "$WT" "$AT" | grep '^+[^+]' | grep -cE '\bt[0-9]{2,}\b|SPEC-[A-Z]|[0-9]{4}-[0-9]{2}-[0-9]{2}|\b[0-9a-f]{9,40}\b'` runs, Then it prints `0`; positive control, one line per alternative: `printf 'x t1279 y\nSPEC-X-001\n2026-09-27\n2370c5b31\n' | grep -cE '\bt[0-9]{2,}\b|SPEC-[A-Z]|[0-9]{4}-[0-9]{2}-[0-9]{2}|\b[0-9a-f]{9,40}\b'` prints `4`; and `git diff -U0 "$CARD_BASE"..HEAD -- "$KT" | grep -c '^+[^+]'` prints at least `1`.
- **AC-SMM-018** (REQ-SMM-017) — Given M4 is committed, When the non-blank lines of the `/clear` section at `CARD_BASE` are written to a file with `git show "$CARD_BASE:$KT" | awk '/^## The `\/clear` handoff between phases/{f=1;next} /^## /{f=0} f && NF' > clear-base.txt` and `grep -vxF -f "$KT" clear-base.txt | wc -l` runs, Then `clear-base.txt` has at least `1` line and the count is `0`; positive control: the same check against a copy of `$KT` with `asks the operator` changed to `tells the operator` prints `1` (observed at `b59a5d69c`: 3 lines, `0`, `1`). And the `[HARD] **Card worktree branches carry the \`WT-\` prefix` line at `CARD_BASE` appears verbatim in `$KT` (`grep -cxF`, `1`); and each of `[HARD] **A new card starts in a new worktree — exit any previous one first.**`, `The fresh tree is created from the remote default branch rather than reused`, `a dependency is a reason to merge, never to reuse the tree` is found in `$KT` by `grep -cF` (`≥ 1`); and `grep -c '\[HARD\]'` on `$KT`, `$DT`, `$WT`, `$LP`, `$CL` is at least its value at `CARD_BASE` (`git show "$CARD_BASE:<path>" | grep -c '\[HARD\]'`).
- **AC-SMM-019** (REQ-SMM-018) — Given M4 is committed, When `test -e internal/template/templates/.claude/rules/local` and `grep -rl 'gitflow-lane-protocol' internal/template/templates/ | wc -l` run, Then the first exits non-zero and the second prints `0`, while `grep -rl 'gitflow-lane-protocol' .claude/rules/local/ | wc -l` prints at least `1` (the control). The lane-protocol and CLAUDE.local.md content checks are AC-SMM-010, AC-SMM-011, AC-SMM-012, AC-SMM-014.

### Hook (conditional on DP-1)

- **AC-SMM-020** (REQ-SMM-019) — Given DP-1 is recorded in progress.md §E.2 as `dp1: docs-only`, `dp1: warn`, or `dp1: block`, and `FX` is the absolute path of a scratch fixture with a worktree at `$FX/.claude/worktrees/w1`, and `P='{"hook_event_name":"PostToolUse","tool_name":"EnterWorktree","session_id":"ac-probe","cwd":"'"$FX"'/.claude/worktrees/w1","tool_input":{"path":".claude/worktrees/w1"},"tool_response":{}}'`:
  - **docs-only:** When `git diff --name-only "$CARD_BASE"..HEAD -- internal/hook/ | wc -l` runs, Then it prints `0` (range non-empty per AC-SMM-021).
  - **warn:** When `MOAI_KANBAN=1 go run ./cmd/moai hook post-tool <<< "$P" | jq -r '.hookSpecificOutput.additionalContext // ""' | grep -c '/clear'` runs, Then it prints `1`; without `MOAI_KANBAN` it prints `0`; with the payload's `cwd` set to an exempt integration worktree it prints `0`; and `$E` carries `additional_context_visible: yes` from the M5a visibility probe.
  - **block:** When the same payload with `"hook_event_name":"PreToolUse"` goes to `go run ./cmd/moai hook pre-tool`, Then `jq -r '.hookSpecificOutput.permissionDecision // "none"'` prints `none` with the flag disabled, `deny` with the flag enabled, with `permissionDecisionReason` starting with the sentinel, and `none` for an exempt target.

## §D.1 Edge cases

- A probe transcript with no `skill_listing` attachment: its `fixture_trees` value is `gap` with a reason, never an empty set.
- `/clear` keeps the same session id: `session_id_P3` equals `session_id_P2`, and P3 counts rows after the `/clear` command row. A new id: `session_id_P3` is the new id.
- A standing lane that must change cards mid-window (integration pending): DP-2 governs; the between-cards `/clear` stays after the move.

## §D.2 Quality gate

- `go run ./cmd/moai spec lint .moai/specs/SPEC-SESSION-MIDMOVE-001` exits `0`.
- `go run ./cmd/moai spec lint --baseline .moai/spec-lint-baseline.json` exits `0`.
- `make build` succeeds after template edits.
- Under DP-1 warn/block: `go test ./internal/hook/... -count=1` exits `0`.

## §D.3 Traceability

| REQ | AC |
|---|---|
| REQ-SMM-001 | AC-SMM-001 |
| REQ-SMM-002 | AC-SMM-002, AC-SMM-003 |
| REQ-SMM-003 | AC-SMM-004 |
| REQ-SMM-004 | AC-SMM-005 |
| REQ-SMM-005 | AC-SMM-006 |
| REQ-SMM-006 | AC-SMM-003, AC-SMM-007 |
| REQ-SMM-007 | AC-SMM-008 |
| REQ-SMM-008 | AC-SMM-009 |
| REQ-SMM-009 | AC-SMM-010 |
| REQ-SMM-010 | AC-SMM-011 |
| REQ-SMM-011 | AC-SMM-012 |
| REQ-SMM-012 | AC-SMM-002, AC-SMM-013 |
| REQ-SMM-013 | AC-SMM-014 |
| REQ-SMM-014 | AC-SMM-015, AC-SMM-021 |
| REQ-SMM-015 | AC-SMM-016, AC-SMM-021 |
| REQ-SMM-016 | AC-SMM-017, AC-SMM-021 |
| REQ-SMM-017 | AC-SMM-018, AC-SMM-021 |
| REQ-SMM-018 | AC-SMM-019 (+ AC-SMM-010, 011, 012, 014 on `$LP`/`$CL`) |
| REQ-SMM-019 | AC-SMM-020 |

## §D.4 Definition of Done

- `dp1:`, `dp2:`, `dp3:` recorded in progress.md §E.2 before M1.
- AC-SMM-001 … AC-SMM-021 each PASS, or N/A with the reason written. N/A is allowed only for the unselected branches of AC-SMM-014 and AC-SMM-020, and for AC-SMM-005 when neither trigger holds.
- The §D.2 quality gate is green; evidence files are force-added and committed; the lane pushes nothing.
