# Acceptance — SPEC-SESSION-MIDMOVE-001

Verification layer. Every criterion is Given-When-Then and binary; each is decided by the fenced commands under it and their stated expected stdout. Criteria are evaluated before the card branch is integrated into develop.

## Variables (set in the same shell as each check)

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
DS=.moai/specs/SPEC-SESSION-MIDMOVE-001/design.md
S=$(mktemp -d)   # scratch directory outside the repository; every check writes only here
```

**Read-time values.** Each value below is read by running its command alone and assigning the printed value literally in the next command (the worktree-session guard refuses `$(git …)`; research.md §R5). `CARD_BASE` is the lane rule's merge-base: `git merge-base develop HEAD` → `CARD_BASE=<sha>`. `CAPS_COMMIT`: `git log --format=%H --diff-filter=A -- .moai/reports/t1279/m1-caps.md` → `CAPS_COMMIT=<sha>`. `CAPS_CT`: `git log -1 --format=%ct <CAPS_COMMIT>` → `CAPS_CT=<epoch>`. `FIRST_KT`: the last line of `git log --no-merges --format=%H <CARD_BASE>..HEAD -- internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` → `FIRST_KT=<sha>`. `T1175`: the `t1175_absorbed:` value in `$E`. `TX`: the transcript path that follows `--cost ` on the `command:` line of `$R/m1-cost.md`.

**Canonical phrases** (doctrine must use them verbatim; checks are fixed-string):

| Id | Phrase |
|---|---|
| P-FLOW | ``move → `/clear` → re-send`` |
| P-CARD | ``On a card change, `/clear` exactly once, after the move`` |
| P-HEAD | `### Card change in a standing session` |
| P-RESEND | `the lead re-sends` |
| P-MOVE | `adds the moved-into tree's skill listing` |
| P-LAUNCH | `gives one skill listing` |
| P-CLEAR | `leaves only the new tree's listing` |
| P-RELAUNCH | `relaunching it through the launcher` |
| P-EXEMPT | `mid-session move exemption: the move into the integration or release worktree to merge, and the return move to the card worktree` |

## §D — AC Matrix

### Range sanity and ordering gate

- **AC-SMM-021** (REQ-SMM-014..017) — Given `CARD_BASE` is read, When the commands run, Then the first exits `0` and the second prints at least `1`. Every range-based criterion is FAIL while this one is FAIL.
  ```bash
  git rev-parse --verify "$CARD_BASE^{commit}"
  git diff --name-only "$CARD_BASE"..HEAD | wc -l
  ```
- **AC-SMM-022** (REQ-SMM-020) — Given `$E` records `t1175_absorbed: <sha>` and `t1257_status: absorbed <sha>` or `t1257_status: not-landed`, When the commands run, Then both `merge-base` calls exit `0`; for `absorbed`, `git merge-base --is-ancestor <t1257 sha> "$FIRST_KT"` also exits `0`; for `not-landed`, `grep -c '^t1257_notified: yes$' "$E"` prints `1`. RED-now: the t1175 branch tip is not an ancestor (`git merge-base --is-ancestor WT-rules-diet HEAD` exit `1`, research.md §R8).
  ```bash
  git merge-base --is-ancestor "$T1175" HEAD
  git merge-base --is-ancestor "$T1175" "$FIRST_KT"
  ```

### Fixture mechanism check (M1)

- **AC-SMM-001** (REQ-SMM-001) — Given M1 has run, When the commands run, Then the first prints exactly `.moai/reports/t1279/m1-caps.md`, the heredoc prints `0`, and `grep -c` prints `4`.
  ```bash
  git show --name-only --format= "$CAPS_COMMIT"
  python3 - "$E" "$CAPS_CT" <<'PY'
  import sys,datetime
  d=dict(l.rstrip("\n").split(": ",1) for l in open(sys.argv[1],encoding="utf-8") if ": " in l)
  iso=int(datetime.datetime.fromisoformat(d["first_probe_ts"].replace("Z","+00:00")).timestamp())
  print(sum([iso!=int(d["first_probe_epoch"]), not int(sys.argv[2])<int(d["first_probe_epoch"])]))
  PY
  grep -cE '6 probe sessions|4 turns|timeout -k 10 300|45 minutes' "$R/m1-caps.md"
  ```
- **AC-SMM-002** (REQ-SMM-002, REQ-SMM-012) — Given `$E` and `$R/extract.txt` are committed, When the heredoc runs, Then it prints `0`, and `grep -c '^attribution_conflict ' "$R/extract.txt"` prints `0`. Negative controls: a copy of `$E` with `listing_count_P1: 0` while `listing_trees_P1` stays `wt` prints `2` (the count/tree mismatch, plus the derived key that no longer matches); a copy whose `launcher_single_listing` is flipped prints `1`.
  ```bash
  python3 - "$E" <<'PY'
  import sys,re
  d=dict(l.rstrip("\n").split(": ",1) for l in open(sys.argv[1],encoding="utf-8") if ": " in l)
  bad=0; sets=("primary","wt","primary+wt","none")
  for p in "123":
      c,t,b,k=(d.get(f"{x}_P{p}","") for x in ("listing_count","listing_trees","listing_bytes","turn_tokens"))
      if c in ("gap","0"):
          bad+= not (t=="gap" and b=="gap"); continue
      if not re.fullmatch(r"[1-9][0-9]*",c): bad+=1; continue
      ts,bs=t.split(";"),b.split(",")
      bad+= not (all(x in sets for x in ts) and len(ts)==int(c)==len(bs) and all(re.fullmatch(r"[0-9]+",x) for x in bs))
      bad+= not (re.fullmatch(r"[0-9]+(,[0-9]+)*",k) and len(k.split(","))<=4)
  def der(p,f):
      c=d.get(f"listing_count_P{p}","gap")
      return "gap" if c in ("gap","0") else ("yes" if f(d[f"listing_trees_P{p}"].split(";")) else "no")
  exp={"launcher_single_listing":der("1",lambda t:t==["wt"]),
       "move_adds_listing":der("2",lambda t:len(t)>=2 and t[0]=="primary" and any("wt" in x for x in t[1:])),
       "clear_restores_single_listing":der("3",lambda t:all(x=="wt" for x in t))}
  print(bad+sum(d.get(k)!=v for k,v in exp.items()))
  PY
  grep -c '^attribution_conflict ' "$R/extract.txt"
  ```
- **AC-SMM-003** (REQ-SMM-002, REQ-SMM-006) — Given `$E` is committed, When the commands run, Then `grep -cE` prints `3` and the heredoc prints `0`. Negative controls: a copy with `probe_cwd_P1` replaced by `<fixture_root>-evil/.claude/worktrees/w1` prints `2`; a copy with `probe_cwd_P2` starting at `<fixture_root>/.claude/worktrees/w1` prints `1`.
  ```bash
  grep -cE '^session_id_P[123]: [0-9a-f-]{36}$' "$E"
  python3 - "$E" <<'PY'
  import sys,os
  d=dict(l.rstrip("\n").split(": ",1) for l in open(sys.argv[1],encoding="utf-8") if ": " in l)
  r=os.path.realpath(d["fixture_root"]); w1=os.path.join(r,".claude","worktrees","w1"); bad=0
  inside=lambda c: os.path.commonpath([r,os.path.realpath(c)])==r
  for p,first in (("1",w1),("2",r),("3",None)):
      cs=d[f"probe_cwd_P{p}"].split(",")
      bad+=sum(not inside(c) for c in cs)
      if first and os.path.realpath(cs[0])!=first: bad+=1
  print(bad)
  PY
  ```
- **AC-SMM-004** (REQ-SMM-003) — Given `$E` and `$R/extract.txt` are committed, When the commands run, Then they print `1`, `0`, and at least `1`.
  ```bash
  grep -c '^judge_source: transcript$' "$E"
  grep -E '^\$ ' "$R/extract.txt" | grep -vc '^\$ python3 .moai/reports/t1279/extract_listing.py '
  grep -c '^\$ ' "$R/extract.txt"
  ```
- **AC-SMM-005** (REQ-SMM-004) — Given the controls are committed, When the commands run, Then they print at least `1`, `0`, and at least `1`. Required whenever a P1–P3 set is a single tree or no path reports a scoped name.
  ```bash
  python3 "$R/extract_listing.py" --fixture-root /nonexistent "$R/m1-control.jsonl" | grep -c 'scoped_names=[1-9]'
  python3 "$R/extract_listing.py" --fixture-root /nonexistent "$R/m1-control-neg.jsonl" | grep -c 'scoped_names=[1-9]'
  python3 "$R/extract_listing.py" --fixture-root /nonexistent "$R/m1-control-neg.jsonl" | grep -c 'scoped_names=0'
  ```
- **AC-SMM-006** (REQ-SMM-005, REQ-SMM-001) — Given `$E` and `$R/probes.txt` are committed, When the heredoc runs, Then it prints `0`. Negative controls: a copy of `$E` with one `gap` value and its reason line removed prints `1`; a copy with `wall_clock_min` off by one prints `1`.
  ```bash
  python3 - "$E" "$R/probes.txt" <<'PY'
  import sys,re,math
  d=dict(l.rstrip("\n").split(": ",1) for l in open(sys.argv[1],encoding="utf-8") if ": " in l)
  bad=sum(1 for k,v in d.items() if v=="gap" and not d.get("gap_reason_"+k,"").strip())
  ids={d[f"session_id_P{p}"] for p in "123"}
  bad+= not (len(ids)<=int(d["probes_run"])<=6)
  span=int(d["last_probe_epoch"])-int(d["first_probe_epoch"])
  bad+= int(d["wall_clock_min"])!=math.ceil(span/60) or int(d["wall_clock_min"])>45
  ops=re.findall(r"^# operator-run: session=([0-9a-f-]{36}) ",open(sys.argv[2],encoding="utf-8").read(),re.M)
  bad+=sum(1 for o in ops if o not in ids)
  print(bad)
  PY
  ```
- **AC-SMM-007** (REQ-SMM-006) — Given `$R/probes.txt` and `$E` are committed, When the commands run, Then the heredoc prints `N 0` with `N ≥ 1`, and the `git diff | grep` pipeline prints `0`. Negative controls: a probes copy whose line ends `… --max-turns 4; rm -rf x` prints `N 1`; a copy whose `cd --` target is `<fixture_root>-evil` prints `N 1`; a copy carrying `-d` prints `N 1`.
  ```bash
  python3 - "$R/probes.txt" "$E" <<'PY'
  import sys,re,os
  d=dict(l.rstrip("\n").split(": ",1) for l in open(sys.argv[2],encoding="utf-8") if ": " in l)
  r=os.path.realpath(d["fixture_root"])
  U="unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED CLAUDE_PROJECT_DIR && cd -- "
  pat=re.compile(re.escape(U)+r"(/[^ ;&|<>`$()]+) && timeout -k 10 300 claude ([^;&|<>`$()]*)")
  n=bad=0
  for l in open(sys.argv[1],encoding="utf-8").read().split("\n"):
      if not l or l.startswith("#"): continue
      n+=1; m=pat.fullmatch(l)
      if (not m or os.path.commonpath([r,os.path.realpath(m.group(1))])!=r or "--max-turns 4" not in m.group(2)
          or re.search(r"(^| )(--setting-sources|-d|--debug|--debug-file)( |=|$)",m.group(2))): bad+=1
  bad_state=sum(d[k+"_before"]!=d[k+"_after"] for k in ("repo_head","repo_branch","repo_status_sha"))
  print(n,bad+bad_state)
  PY
  git diff --name-only "$CARD_BASE"..HEAD -- .moai/reports/t1279/ | grep -E '\.jsonl$|debug' | grep -vcE '/m1-control(-neg)?\.jsonl$'
  ```
- **AC-SMM-008** (REQ-SMM-007) — Given M1 is committed, When the commands run, Then the first exits `0` and the next two print `0`.
  ```bash
  test -s "$R/probes.txt" && test -s "$R/extract.txt" && test -s "$R/fixture.txt"
  grep -c 'extract_listing.py' "$R/probes.txt"
  grep -c 'claude -p' "$R/fixture.txt"
  ```
  Also `fixture.txt` records the two disjointness checks with exit code 0: `grep -cE '^\$ test ! -e .*/\.claude/worktrees/w1/\.claude/skills/fx-primary-a +# rc=0$|^\$ test ! -e .*/\.claude/skills/fx-wt-a +# rc=0$' "$R/fixture.txt"` prints `2`.

### Real-session cost (M2)

- **AC-SMM-009** (REQ-SMM-008) — Given `$R/m1-cost.md` is committed, When the commands run, Then exactly one of the first two counts is non-zero, the third prints `1`, and while the transcript is readable the final `cmp` exits `0` (re-run equals file). When the transcript is gone, the file carries the `cost: gap` line and the `cmp` step is N/A with that reason.
  ```bash
  grep -cE '^row: ts=[0-9T:.-]+Z skillCount=[0-9]+ scoped_names=[1-9][0-9]* content_bytes=[0-9]+ before=[0-9]+ after=[0-9]+ delta=-?[0-9]+ other_row_bytes=[0-9]+ bound=(upper|confounded)$' "$R/m1-cost.md"
  grep -cE '^cost: gap reason=.+' "$R/m1-cost.md"
  grep -c '^command: python3 .moai/reports/t1279/extract_listing.py --cost ' "$R/m1-cost.md"
  python3 "$R/extract_listing.py" --cost "$TX" | grep '^row:' > "$S/cost-rerun.txt"
  grep '^row:' "$R/m1-cost.md" > "$S/cost-file.txt"
  cmp "$S/cost-rerun.txt" "$S/cost-file.txt"
  ```
  Reference (research.md §R2): the `03:57:49.175Z` row reads `content_bytes=43566 scoped_names=42 before=278878 after=299707 delta=20829 other_row_bytes=20526 bound=upper`.

### Doctrine (M4)

- **AC-SMM-010** (REQ-SMM-009) — Given M4 is committed, When the command runs, Then every file prints at least `1`. RED-now: `$KT`, `$DT`, `$WT`, `$AT`, `$AL`, `$CL` print `0`; `$LP` prints `1` (its line is a regression guard, not RED).
  ```bash
  grep -c 'moai cc -w <card-id>' "$KT" "$DT" "$WT" "$AT" "$AL" "$CL" "$LP"
  ```
- **AC-SMM-011** (REQ-SMM-010) — Given M4 is committed, When the commands run, Then every count is at least `1`. RED-now: all `0`.
  ```bash
  grep -cF 'move → `/clear` → re-send' "$DT" "$AT" "$LP" "$CL"
  awk '/^## The `\/clear` handoff between phases/{f=1;next} /^## /{f=0} f' "$KT" | grep -cF 'On a card change, `/clear` exactly once, after the move'
  grep -cF '### Card change in a standing session' "$DT"
  awk '/^### Card change in a standing session/{f=1;next} /^##/{f=0} f' "$DT" | grep -cF 'the lead re-sends'
  awk '/^### Card change in a standing session/{f=1;next} /^##/{f=0} f' "$DT" | grep -ciE 'relaunch[^.]{0,60}(optional|not required)'
  ```
- **AC-SMM-012** (REQ-SMM-011) — Given M4 is committed, When the commands run, Then the first command prints `1` for `$KT` and `0` for `$LP`, the second prints `0`, and the heredoc prints `0`. RED-now (research.md §R7): `2`, `2`, `1`, and `6`. Units counted include the new-card `[HARD]` paragraph. A unit is excused only by the literal P-FLOW phrase (which carries both `/clear` and re-send) or by the integration/`auto-names` exclusions. The remaining `EnterWorktree(<card-id>)` in `$KT` is the one inside the preserved `[HARD]` new-card line.
  ```bash
  grep -c 'EnterWorktree(<card-id>)' "$KT" "$LP"
  grep -cE 'moai cc -w <card-id>` (또는|or) .*EnterWorktree' "$LP"
  python3 - "$KT" "$DT" "$AT" "$CL" "$LP" <<'PY'
  import re,sys
  ex=re.compile(r"release|integration|develop|통합|auto-names|Return the same way|병합을 마치고")
  flow="move → `/clear` → re-send"; b=0
  for f in sys.argv[1:]:
      for p in re.split(r"\n\s*\n",open(f,encoding="utf-8").read()):
          ls=p.split("\n"); rows=[l for l in ls if l.startswith("|")]
          pr="\n".join(l for l in ls if not l.startswith("|"))
          b+=sum(1 for u in rows+[pr] if "EnterWorktree(" in u and not ex.search(u) and flow not in u)
  print(b)
  PY
  ```
  The first command's `$KT` expectation is `1`, not `0`: that remaining occurrence is the preserved `[HARD]` line.
- **AC-SMM-013** (REQ-SMM-012) — Given `$E` and M4 are committed, When the heredoc runs, Then it prints `0`. RED-now with `clear_restores_single_listing: gap`: `1` (the P-RELAUNCH phrase is not yet in `$DT`). Negative control: a copy of `$DT` with the P-CLEAR phrase appended raises the count by `1` (`1` → `2` observed).
  ```bash
  git diff -U0 "$CARD_BASE"..HEAD -- "$KT" "$DT" "$WT" "$AT" "$LP" "$CL" | grep '^+[^+]' > "$S/added.txt"
  python3 - "$E" "$S/added.txt" "$KT" "$DT" "$WT" "$AT" "$LP" "$CL" <<'PY'
  import sys
  d=dict(l.rstrip("\n").split(": ",1) for l in open(sys.argv[1],encoding="utf-8") if ": " in l)
  ph={"move_adds_listing":"adds the moved-into tree's skill listing","launcher_single_listing":"gives one skill listing",
      "clear_restores_single_listing":"leaves only the new tree's listing"}
  body="".join(open(f,encoding="utf-8").read() for f in sys.argv[3:])
  bad=sum(body.count(v) for k,v in ph.items() if d.get(k)!="yes")
  if d.get("clear_restores_single_listing")!="yes": bad+= "relaunching it through the launcher" not in open(sys.argv[6],encoding="utf-8").read()
  canon=list(ph.values())+["relaunching it through the launcher"]
  bad+=sum(1 for l in open(sys.argv[2],encoding="utf-8") if "listing" in l.lower() and not any(c in l for c in canon))
  print(bad)
  PY
  ```
- **AC-SMM-014** (REQ-SMM-013) — Given `dp2: exempt` or `dp2: covered` is recorded in progress.md §E.2, When the commands run, Then for `exempt` each of the three counts is exactly `1` (RED-now: `0`), and for `covered` the fourth prints at least `1`.
  ```bash
  awk '/^## Integration into the release branch is self-served/{f=1;next} /^## /{f=0} f' "$KT" | grep -cF 'mid-session move exemption: the move into the integration or release worktree to merge, and the return move to the card worktree'
  grep -cF 'mid-session move exemption: the move into the integration or release worktree to merge, and the return move to the card worktree' "$LP"
  grep -cF 'mid-session move exemption: the move into the integration or release worktree to merge, and the return move to the card worktree' "$CL"
  awk '/^## Integration into the release branch is self-served/{f=1;next} /^## /{f=0} f' "$KT" | grep -cF 'move → `/clear` → re-send'
  ```

### Where the doctrine lives

- **AC-SMM-015** (REQ-SMM-014) — Given M4 is committed, When the commands run (each git call on its own; no process substitution), Then:
  - the three `cmp` calls and the `diff` exit `0`;
  - each §3 extraction has more than `0` lines;
  - each `git log` exits `0`;
  - each per-pair `comm -23` prints `0` lines;
  - the `$KT` commit count is at least `1` (RED-now: `0`).
  ```bash
  cmp "$KT" "$KL"; cmp "$DT" "$DL"; cmp "$WT" "$WL"
  sed -n '/^## 3. Worktrees/,/^## 4\./p' "$AL" > "$S/al3.txt"
  sed -n '/^## 3. Worktrees/,/^## 4\./p' "$AT" > "$S/at3.txt"
  wc -l < "$S/al3.txt"; wc -l < "$S/at3.txt"; diff "$S/al3.txt" "$S/at3.txt"
  git log --no-merges --format=%H "$CARD_BASE"..HEAD -- "$KL" > "$S/c-kl.txt"
  git log --no-merges --format=%H "$CARD_BASE"..HEAD -- "$KT" > "$S/c-kt.txt"
  git log --no-merges --format=%H "$CARD_BASE"..HEAD -- "$DL" > "$S/c-dl.txt"
  git log --no-merges --format=%H "$CARD_BASE"..HEAD -- "$DT" > "$S/c-dt.txt"
  git log --no-merges --format=%H "$CARD_BASE"..HEAD -- "$WL" > "$S/c-wl.txt"
  git log --no-merges --format=%H "$CARD_BASE"..HEAD -- "$WT" > "$S/c-wt.txt"
  git log --no-merges --format=%H "$CARD_BASE"..HEAD -- "$AL" > "$S/c-al.txt"
  git log --no-merges --format=%H "$CARD_BASE"..HEAD -- "$AT" > "$S/c-at.txt"
  sort -o "$S/c-kl.txt" "$S/c-kl.txt"; sort -o "$S/c-kt.txt" "$S/c-kt.txt"; comm -23 "$S/c-kl.txt" "$S/c-kt.txt" | wc -l
  sort -o "$S/c-dl.txt" "$S/c-dl.txt"; sort -o "$S/c-dt.txt" "$S/c-dt.txt"; comm -23 "$S/c-dl.txt" "$S/c-dt.txt" | wc -l
  sort -o "$S/c-wl.txt" "$S/c-wl.txt"; sort -o "$S/c-wt.txt" "$S/c-wt.txt"; comm -23 "$S/c-wl.txt" "$S/c-wt.txt" | wc -l
  sort -o "$S/c-al.txt" "$S/c-al.txt"; sort -o "$S/c-at.txt" "$S/c-at.txt"; comm -23 "$S/c-al.txt" "$S/c-at.txt" | wc -l
  wc -l < "$S/c-kt.txt"
  ```
  Negative control: a `c-kl.txt` holding one SHA absent from `c-kt.txt` makes that pair print `1`.
- **AC-SMM-016** (REQ-SMM-015) — Given `always_loaded_before: <N>` and `claude_local_bytes_before: <B>` are recorded in `$E` before the first doctrine commit, When the commands run, Then:
  - the test exits `0` and its output contains lines starting `--- PASS: TestAlwaysLoadedTokenBudget (` and `--- PASS: TestCodexContractByteCeiling (`;
  - the heredoc prints `0`;
  - the diff line count prints `0`;
  - `grep -c` prints `1`.
  ```bash
  go test ./internal/config/ -run '^TestCodexContractByteCeiling$|^TestAlwaysLoadedTokenBudget$' -count=1 -v > "$S/budget.txt"; echo "rc=$?"
  python3 - "$E" "$S/budget.txt" "$CL" <<'PY'
  import sys,re,os
  d=dict(l.rstrip("\n").split(": ",1) for l in open(sys.argv[1],encoding="utf-8") if ": " in l)
  m=int(re.search(r"always-loaded surface = (\d+) tokens",open(sys.argv[2]).read()).group(1))
  print(int(m>int(d["always_loaded_before"]))+int(os.path.getsize(sys.argv[3])-int(d["claude_local_bytes_before"])>600))
  PY
  git diff "$CARD_BASE"..HEAD -- internal/config/token_budget_guard.go | wc -l
  grep -c '^const AlwaysLoadedTokenBudget = ' internal/config/token_budget_guard.go
  ```
- **AC-SMM-017** (REQ-SMM-016) — Given M4 is committed, When the commands run, Then the first prints `0`, the control prints `4`, and the last prints at least `1` (RED-now: `0`).
  ```bash
  git diff -U0 "$CARD_BASE"..HEAD -- "$KT" "$DT" "$WT" "$AT" | grep '^+[^+]' | grep -cE '\bt[0-9]{2,}\b|SPEC-[A-Z]|[0-9]{4}-[0-9]{2}-[0-9]{2}|\b[0-9a-f]{9,40}\b'
  printf 'x t1279 y\nSPEC-X-001\n2026-09-27\n2370c5b31\n' | grep -cE '\bt[0-9]{2,}\b|SPEC-[A-Z]|[0-9]{4}-[0-9]{2}-[0-9]{2}|\b[0-9a-f]{9,40}\b'
  git diff -U0 "$CARD_BASE"..HEAD -- "$KT" | grep -c '^+[^+]'
  ```
- **AC-SMM-018** (REQ-SMM-017) — Given M4 is committed, When the commands run for both copies (`$KT` shown; repeat with `$KL`), Then:
  - `old153.txt` and `new153.txt` each have `1` line;
  - the preservation heredoc prints `0`;
  - the old-line count prints `0`;
  - the section-scoped new-line heredoc prints `1`;
  - the file-wide new-line count prints `1`;
  - the non-kanban `[HARD]` count heredoc prints `0`.

  RED-now on `$KT`: preservation `0`, old line `1`, section new line `0`, file new line `0`. Controls (research.md §R11): a copy of the base file with the `MUST \`ExitWorktree\`` line moved to the file end gives preservation `1`; a copy with the old line replaced by the new one gives preservation `0` and section new line `1`.
  ```bash
  sed -n '/^<!-- old-153 -->$/,/^<!-- \/old-153 -->$/p' "$DS" | sed '1d;$d' > "$S/old153.txt"
  sed -n '/^<!-- amendment-153 -->$/,/^<!-- \/amendment-153 -->$/p' "$DS" | sed '1d;$d' > "$S/new153.txt"
  wc -l < "$S/old153.txt"; wc -l < "$S/new153.txt"
  git show "$CARD_BASE:$KT" > "$S/kt-base.md"
  python3 - "$S/kt-base.md" "$KT" "$S/old153.txt" <<'PY'
  import sys
  old=open(sys.argv[3],encoding="utf-8").read().rstrip("\n")
  def secs(p):
      s,o="",[]
      for l in open(p,encoding="utf-8").read().split("\n"):
          if l.startswith("## "): s=l
          o.append((s,l))
      return o
  h=set(secs(sys.argv[2]))
  print(sum(1 for s,l in secs(sys.argv[1]) if "[HARD]" in l and l!=old and (s,l) not in h))
  PY
  grep -cxF -f "$S/old153.txt" "$KT"
  python3 - "$KT" "$S/new153.txt" <<'PY'
  import sys
  new=open(sys.argv[2],encoding="utf-8").read().rstrip("\n"); s=""; n=0
  for l in open(sys.argv[1],encoding="utf-8").read().split("\n"):
      if l.startswith("## "): s=l
      n+= s=="## The `/clear` handoff between phases" and l==new
  print(n)
  PY
  grep -cxF -f "$S/new153.txt" "$KT"
  git show "$CARD_BASE:$DT" > "$S/b-dt.md"
  git show "$CARD_BASE:$WT" > "$S/b-wt.md"
  git show "$CARD_BASE:$LP" > "$S/b-lp.md"
  git show "$CARD_BASE:$CL" > "$S/b-cl.md"
  python3 - "$S/b-dt.md" "$DT" "$S/b-wt.md" "$WT" "$S/b-lp.md" "$LP" "$S/b-cl.md" "$CL" <<'PY'
  import sys
  c=lambda p: open(p,encoding="utf-8").read().count("[HARD]")
  a=sys.argv[1:]; print(sum(int(c(a[i+1])<c(a[i])) for i in range(0,len(a),2)))
  PY
  ```
- **AC-SMM-019** (REQ-SMM-018) — Given M4 is committed, When the commands run, Then the `test` exits non-zero, the template count prints `0`, and the local control prints at least `1`. Content checks for `$LP`/`$CL` are AC-SMM-010, 011, 012, 014; the CLAUDE.local.md growth cap is AC-SMM-016.
  ```bash
  test -e internal/template/templates/.claude/rules/local
  grep -rl 'gitflow-lane-protocol' internal/template/templates/ | wc -l
  grep -rl 'gitflow-lane-protocol' .claude/rules/local/ | wc -l
  ```

### Hook (conditional on DP-1)

- **AC-SMM-020** (REQ-SMM-019) — Given `dp1: docs-only`, `dp1: warn`, or `dp1: block` is recorded in progress.md §E.2, and `FX` is a scratch fixture with `$FX/.claude/worktrees/w1`, and `P='{"hook_event_name":"PostToolUse","tool_name":"EnterWorktree","session_id":"ac-probe","cwd":"'"$FX"'/.claude/worktrees/w1","tool_input":{"path":".claude/worktrees/w1"},"tool_response":{}}'`:
  - **docs-only:** the first command prints `0` (AC-SMM-021 guarantees the range is non-empty).
  - **warn:** the second prints `1`, the third prints `0` (variables unset in the same invocation), and the fourth — with `cwd` pointing at an exempt integration worktree — prints `0`; `$E` carries `additional_context_visible: yes`.
  - **block:** with `workflow.midmove_guard.enabled: true` written into the fixture's `.moai/config/sections/workflow.yaml`, the PreToolUse form prints `deny`, with `permissionDecisionReason` starting with the sentinel; with the key false or absent it prints `none`; for an exempt target, `none`.
  ```bash
  git diff --name-only "$CARD_BASE"..HEAD -- internal/hook/ | wc -l
  MOAI_KANBAN=1 go run ./cmd/moai hook post-tool <<< "$P" | jq -r '.hookSpecificOutput.additionalContext // ""' | grep -c '/clear'
  unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go run ./cmd/moai hook post-tool <<< "$P" | jq -r '.hookSpecificOutput.additionalContext // ""' | grep -c '/clear'
  ```

## §D.1 Edge cases

- A probe with no `skill_listing`: its listing count is `0` and every tree/byte value is `gap` (AC-SMM-002 rejects anything else).
- `/clear` keeps the same session id: `session_id_P3` equals `session_id_P2`, and only rows after the `/clear` row count. A new id: `session_id_P3` is that id.
- A standing lane that changes cards mid-window: DP-2 governs; the card-change `/clear` stays after the move.

## §D.2 Quality gate

- `go run ./cmd/moai spec lint .moai/specs/SPEC-SESSION-MIDMOVE-001` exits `0`.
- `go run ./cmd/moai spec lint --baseline .moai/spec-lint-baseline.json` exits `0`.
- `make build` succeeds after template edits.
- Under DP-1 warn/block: `go test ./internal/hook/... -count=1` exits `0`.

## §D.3 Traceability

| REQ | AC |
|---|---|
| REQ-SMM-001 | AC-SMM-001, AC-SMM-006 |
| REQ-SMM-002 | AC-SMM-002, AC-SMM-003, AC-SMM-008 |
| REQ-SMM-003 | AC-SMM-004 |
| REQ-SMM-004 | AC-SMM-005 |
| REQ-SMM-005 | AC-SMM-006 |
| REQ-SMM-006 | AC-SMM-003, AC-SMM-007 |
| REQ-SMM-007 | AC-SMM-008 |
| REQ-SMM-008 | AC-SMM-009 |
| REQ-SMM-009 | AC-SMM-010 |
| REQ-SMM-010 | AC-SMM-011, AC-SMM-018 |
| REQ-SMM-011 | AC-SMM-012 |
| REQ-SMM-012 | AC-SMM-002, AC-SMM-013 |
| REQ-SMM-013 | AC-SMM-014 |
| REQ-SMM-014 | AC-SMM-015, AC-SMM-021 |
| REQ-SMM-015 | AC-SMM-016, AC-SMM-021 |
| REQ-SMM-016 | AC-SMM-017, AC-SMM-021 |
| REQ-SMM-017 | AC-SMM-018, AC-SMM-021 |
| REQ-SMM-018 | AC-SMM-019 (+ AC-SMM-010, 011, 012, 014, 016 on `$LP`/`$CL`) |
| REQ-SMM-019 | AC-SMM-020 |
| REQ-SMM-020 | AC-SMM-022 |

## §D.4 Definition of Done

- `dp1:`, `dp2:` recorded in progress.md §E.2 before M1.
- AC-SMM-001 … AC-SMM-022 each PASS, or N/A with the reason written. N/A is allowed only for:
  - the unselected branches of AC-SMM-014 and AC-SMM-020;
  - AC-SMM-005 when neither trigger holds;
  - the `cmp` step of AC-SMM-009 when the transcript is unreadable.
- The §D.2 quality gate is green; evidence files are force-added and committed; the lane pushes nothing.

## §D.5 Plan-time RED-now ledger

Measured in this worktree at HEAD `59ecb6582`, with `CARD_BASE` = `b59a5d69c1862b08a8a9e4a48afc0ad33c8d951c`. Commands and verbatim outputs are in research.md §R11. Criteria that depend on run-phase artifacts that do not exist yet (AC-SMM-001–009) were exercised against synthetic evidence in the scratchpad. Their valid inputs print the PASS values and their negative controls print the FAIL values listed above. They are not recorded as passes.
