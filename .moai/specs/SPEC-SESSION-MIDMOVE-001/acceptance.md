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
```

**Read-time values.** Each value below is read by running its command **alone**, then assigning the printed value literally (for example `S=/var/folders/…/tmp.X`) at the start of every later command. The worktree-session guard refuses `$(…)` placed next to git calls, so no value is computed inside a check (research.md §R5, §R11).

| Value | Command run alone | Assign |
|---|---|---|
| `S` (scratch directory outside the repository; checks write only here) | `mktemp -d` | `S=<path>` |
| `CARD_BASE` (lane rule merge-base) | `git merge-base develop HEAD` | `CARD_BASE=<sha>` |
| `CAPS_COMMIT` | `git log --format=%H --diff-filter=A -- .moai/reports/t1279/m1-caps.md` | `CAPS_COMMIT=<sha>` |
| `CAPS_CT` | `git log -1 --format=%ct <CAPS_COMMIT>` | `CAPS_CT=<epoch>` |
| `FIRST_KT` | last line of `git log --no-merges --format=%H <CARD_BASE>..HEAD -- internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` | `FIRST_KT=<sha>` |
| `M1175 T1175`, `M1257 T1257` (develop merge commit, second parent) | `cat <S>/t1175.txt`, `cat <S>/t1257.txt` after the AC-SMM-022 derivation lines | `M1175=<sha> T1175=<sha>`, likewise for t1257 |
| `TX`, `CUTOFF` | `grep '^command: ' .moai/reports/t1279/m1-cost.md` (the path after `--until-line <N> `); `grep '^cutoff_line: ' .moai/reports/t1279/m1-cost.md` | `TX=<path>`, `CUTOFF=<N>` |

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
- **AC-SMM-022** (REQ-SMM-020) — The landing commits are derived from develop's merge commits, never from `$E`.
  - **Given** develop's merge log, **When** block (a) runs, **Then**:
    - the positive-control count is at least `1` (the grep form sees a known landed card merge);
    - `t1175.txt` has exactly `1` line.
  - **When** block (b) runs with `M1175 T1175` read from `t1175.txt`, **Then**:
    - the message grep over the merged-in commits prints at least `1` (a branch tip is often an absorb merge that carries no card id, so the whole merged-in range is read);
    - both `is-ancestor` calls print `rc=0`.
  - **t1257**, block (c):
    - when `t1257.txt` has `1` line, read `M1257 T1257` from it; the message grep prints at least `1` and the `is-ancestor` call prints `rc=0`;
    - when it has `0` lines, the two `$E` greps print `1` and `1` (not-landed, lead notified).
  - **RED-now**, recorded honestly. At the §R11 measurement (HEAD `bce7248ff`), `t1175.txt` had `0` lines because t1175 had not yet landed. **Correction 2026-09-27:** t1175 has since landed on develop as merge commit `7fe658815` (`git merge-base --is-ancestor 7fe658815 develop` exit `0`). This branch has not absorbed it (`… 7fe658815 HEAD` exit `1`), so the gate stays RED for that reason.
  ```bash
  # (a) derivation
  git log develop --merges --format='%H %P %s' > "$S/develop-merges.txt"
  grep -cE "^[0-9a-f]{40} [0-9a-f]{40} [0-9a-f]{40} Merge (branch ')?WT-statusline-landed-label'? into develop" "$S/develop-merges.txt"
  grep -E "^[0-9a-f]{40} [0-9a-f]{40} [0-9a-f]{40} Merge (branch ')?WT-rules-diet'? into develop" "$S/develop-merges.txt" | head -1 | cut -d' ' -f1,3 > "$S/t1175.txt"
  grep -E "^[0-9a-f]{40} [0-9a-f]{40} [0-9a-f]{40} Merge (branch ')?WT-role-naming-docs'? into develop" "$S/develop-merges.txt" | head -1 | cut -d' ' -f1,3 > "$S/t1257.txt"
  wc -l < "$S/t1175.txt"
  wc -l < "$S/t1257.txt"
  # (b) t1175 provenance (the commits that merge brought in name the card) and ancestry
  git log --format=%B "$M1175^1..$T1175" | grep -c 't1175'
  git merge-base --is-ancestor "$T1175" HEAD; echo "rc=$?"
  git merge-base --is-ancestor "$T1175" "$FIRST_KT"; echo "rc=$?"
  # (c) t1257: landed branch
  git log --format=%B "$M1257^1..$T1257" | grep -c 't1257'
  git merge-base --is-ancestor "$T1257" "$FIRST_KT"; echo "rc=$?"
  # (c) t1257: not-landed branch
  grep -c '^t1257_status: not-landed$' "$E"
  grep -c '^t1257_notified: yes$' "$E"
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
- **AC-SMM-002** (REQ-SMM-002, REQ-SMM-005, REQ-SMM-012) — Given `$E` and `$R/extract.txt` are committed, When the two heredocs run, Then each prints `0`.
  - The first checks `$E` for internal consistency. A gap path must carry `gap` in every value, `turn_tokens` included, and the decision keys must be derived as REQ-SMM-002 states.
  - The second re-derives every non-gap value from `extract.txt` for the recorded session id, and checks the P2→P3 transcript linkage and the `attribution_conflict` gap route.
  - Negative controls:

    | Control input | Heredoc | Prints |
    |---|---|---|
    | `listing_count_P1: 0` while `listing_trees_P1` stays `wt` | first | `2` |
    | a flipped `launcher_single_listing` | first | `1` |
    | a gap path carrying `turn_tokens_P3: 1,2` | first | `1` |
    | an `$E` whose `session_id_P3` names an unrelated session | second | ≥ `1` |
    | an `extract.txt` without the `clear` row while P3 is non-gap | second | ≥ `1` |
    | an `extract.txt` carrying an `attribution_conflict` row for a non-gap path | second | ≥ `1` |
    | an `$E` whose `listing_bytes_P1` differs from the extract | second | `1` |
  ```bash
  python3 - "$E" <<'PY'
  import sys,re
  d=dict(l.rstrip("\n").split(": ",1) for l in open(sys.argv[1],encoding="utf-8") if ": " in l)
  bad=0; sets=("primary","wt","primary+wt","none")
  for p in "123":
      c,t,b,k=(d.get(f"{x}_P{p}","") for x in ("listing_count","listing_trees","listing_bytes","turn_tokens"))
      if c in ("gap","0"):
          bad+= not (t=="gap" and b=="gap" and k=="gap"); continue
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
  python3 - "$E" "$R/extract.txt" <<'PY'
  import sys,re
  d=dict(l.rstrip("\n").split(": ",1) for l in open(sys.argv[1],encoding="utf-8") if ": " in l)
  blocks={}; cur=None
  for l in open(sys.argv[2],encoding="utf-8").read().split("\n"):
      m=re.match(r"session ([0-9a-f-]{36})$",l)
      if m: cur=blocks.setdefault(m.group(1),{"L":[],"T":[],"C":[],"X":[]}); continue
      if cur is None: continue
      if l.startswith("listing "): cur["L"].append(dict(x.split("=",1) for x in l.split()[1:]))
      elif l.startswith("turn "): cur["T"].append(dict(x.split("=",1) for x in l.split()[1:]))
      elif l.startswith("clear "): cur["C"].append(int(dict(x.split("=",1) for x in l.split()[1:])["line"]))
      elif l.startswith("attribution_conflict "): cur["X"].append(int(dict(x.split("=",1) for x in l.split()[1:])["line"]))
  bad=0; keys=("listing_count","listing_trees","listing_bytes","turn_tokens")
  def gapped(p,reason):
      return all(d.get(f"{k}_P{p}")=="gap" for k in keys) and d.get(f"gap_reason_listing_trees_P{p}")==reason
  for p in "123":
      sid=d.get(f"session_id_P{p}"); blk=blocks.get(sid)
      if p=="3" and (sid!=d.get("session_id_P2") or blk is None or len(blk["C"])!=1):
          bad+= not gapped("3","no_clear_link"); continue
      if blk is None: bad+=1; continue
      cl=blk["C"][0] if blk["C"] else None
      inr=lambda n: True if p=="1" or cl is None else (int(n)<cl if p=="2" else int(n)>cl)
      if any(inr(x) for x in blk["X"]):
          bad+= not gapped(p,"attribution_conflict"); continue
      L=[x for x in blk["L"] if inr(x["line"])]; T=[x for x in blk["T"] if inr(x["line"])]
      if not L:
          bad+= not all(d.get(f"{k}_P{p}")=="gap" for k in keys); continue
      want={"listing_count":str(len(L)),"listing_trees":";".join(x["trees"] for x in L),
            "listing_bytes":",".join(x["content_bytes"] for x in L),"turn_tokens":",".join(x["tokens"] for x in T)}
      bad+=sum(d.get(f"{k}_P{p}")!=v for k,v in want.items())
  print(bad)
  PY
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
- **AC-SMM-007** (REQ-SMM-006) — Given `$R/probes.txt` and `$E` are committed, When the commands run, Then the heredoc prints `N 0` with `N ≥ 1`, and the `git diff | grep` pipeline prints `0`.
  - The arguments after `claude` are checked against an allow-list (`shlex`), not a deny-list.
  - Negative controls, each a probes copy that prints `N 1`: a line ending `… --max-turns 4; rm -rf x`; a `cd --` target of `<fixture_root>-evil`; the flags `-d`, `--max-turns 40`, a duplicate `--max-turns 4 --max-turns 99`, `--dangerously-skip-permissions`, `--add-dir /x`, and `--settings /x.json`.
  ```bash
  python3 - "$R/probes.txt" "$E" <<'PY'
  import sys,re,os,shlex
  d=dict(l.rstrip("\n").split(": ",1) for l in open(sys.argv[2],encoding="utf-8") if ": " in l)
  r=os.path.realpath(d["fixture_root"])
  U="unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED CLAUDE_PROJECT_DIR && cd -- "
  pat=re.compile(re.escape(U)+r"(/[^ ;&|<>`$()]+) && timeout -k 10 300 claude ([^;&|<>`$()]*)")
  fixed={"--model":"haiku","--max-turns":"4","--output-format":"json","--allowedTools":"EnterWorktree"}
  def args_ok(s):
      try: t=shlex.split(s)
      except ValueError: return False
      seen={}; i=0
      while i<len(t):
          f=t[i]
          if f not in ("-p","--resume") and f not in fixed or i+1>=len(t): return False
          v=t[i+1]; seen[f]=seen.get(f,0)+1
          if f in fixed and v!=fixed[f]: return False
          if f=="--resume" and not re.fullmatch(r"[0-9a-f-]{36}",v): return False
          i+=2
      return all(c==1 for c in seen.values()) and all(seen.get(k)==1 for k in ("-p","--model","--max-turns","--output-format"))
  n=bad=0
  for l in open(sys.argv[1],encoding="utf-8").read().split("\n"):
      if not l or l.startswith("#"): continue
      n+=1; m=pat.fullmatch(l)
      if not m or os.path.commonpath([r,os.path.realpath(m.group(1))])!=r or not args_ok(m.group(2)): bad+=1
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

- **AC-SMM-009** (REQ-SMM-008) — Given `$R/m1-cost.md` is committed, When the commands run, Then:
  - exactly one of the first two counts is non-zero;
  - the third and fourth each print `1`;
  - while the transcript is readable, the final `cmp` exits `0` (a re-run cut at the same `cutoff_line` equals the file, so the live transcript's later growth cannot change the result).

  When the transcript is gone, the file carries the `cost: gap` line and the `cmp` step is N/A with that reason. Negative control: a re-run without `--until-line` on the live transcript (which has grown past `CUTOFF`) makes `cmp` exit `1` whenever a later scoped listing exists.
  ```bash
  grep -cE '^row: ts=[0-9T:.-]+Z skillCount=[0-9]+ scoped_names=[1-9][0-9]* content_bytes=[0-9]+ before=[0-9]+ after=[0-9]+ delta=-?[0-9]+ other_row_bytes=[0-9]+ bound=(upper|confounded)$' "$R/m1-cost.md"
  grep -cE '^cost: gap reason=.+' "$R/m1-cost.md"
  grep -cE '^command: python3 \.moai/reports/t1279/extract_listing\.py --cost --until-line [1-9][0-9]* ' "$R/m1-cost.md"
  grep -cE '^cutoff_line: [1-9][0-9]*$' "$R/m1-cost.md"
  python3 "$R/extract_listing.py" --cost --until-line "$CUTOFF" "$TX" | grep '^row:' > "$S/cost-rerun.txt"
  grep '^row:' "$R/m1-cost.md" > "$S/cost-file.txt"
  cmp "$S/cost-rerun.txt" "$S/cost-file.txt"
  ```
  Reference (research.md §R2): the `03:57:49.175Z` row reads `content_bytes=43566 scoped_names=42 before=278878 after=299707 delta=20829 other_row_bytes=20526 bound=upper`.

### Doctrine (M4)

- **AC-SMM-010** (REQ-SMM-009) — Given M4 is committed, When the command runs, Then every file prints at least `1`. RED-now: `$KT`, `$DT`, `$WT`, `$AT`, `$AL`, `$CL` print `0`; `$LP` prints `1` (its line is a regression guard, not RED).
  ```bash
  grep -c 'moai cc -w <card-id>' "$KT" "$DT" "$WT" "$AT" "$AL" "$CL" "$LP"
  ```
- **AC-SMM-011** (REQ-SMM-010) — Given M4 is committed, When the commands run, Then every count is at least `1`. RED-now: all `0`. The last two lines (ND9) state that a standing session with no next card is cleared as before, and that no work other than the move happens between the move and the `/clear`.
  ```bash
  grep -cF 'move → `/clear` → re-send' "$DT" "$AT" "$LP" "$CL"
  awk '/^## The `\/clear` handoff between phases/{f=1;next} /^## /{f=0} f' "$KT" | grep -cF 'On a card change, `/clear` exactly once, after the move'
  grep -cF '### Card change in a standing session' "$DT"
  awk '/^### Card change in a standing session/{f=1;next} /^##/{f=0} f' "$DT" | grep -cF 'the lead re-sends'
  awk '/^### Card change in a standing session/{f=1;next} /^##/{f=0} f' "$DT" | grep -ciE 'relaunch[^.]{0,60}(optional|not required)'
  awk '/^### Card change in a standing session/{f=1;next} /^##/{f=0} f' "$DT" | grep -cF 'When no next card is ready, the session is cleared as before'
  awk '/^### Card change in a standing session/{f=1;next} /^##/{f=0} f' "$DT" | grep -cF 'Nothing but the move happens between the move and the /clear'
  ```
- **AC-SMM-012** (REQ-SMM-011) — Given M4 is committed, When the commands run, Then the first command prints `1` for `$KT` and `0` for `$LP`, the second prints `0`, and the heredoc prints `0`. RED-now (research.md §R7): `2`, `2`, `1`, and `6`. Units counted include the new-card `[HARD]` paragraph. A unit is excused only by the literal P-FLOW phrase (which carries both `/clear` and re-send), or by the integration exclusions. Those exclusions are path- and section-specific (`worktrees/develop`, `develop 워크트리`, `release`, `integration`, `통합`, the return and merge-back lines), plus the verbatim `[HARD]` `auto-names` line; a bare mention of `develop` no longer excuses a unit (ND11). The remaining `EnterWorktree(<card-id>)` in `$KT` is the one inside the preserved `[HARD]` new-card line.
  ```bash
  grep -c 'EnterWorktree(<card-id>)' "$KT" "$LP"
  grep -cE 'moai cc -w <card-id>` (또는|or) .*EnterWorktree' "$LP"
  python3 - "$KT" "$DT" "$AT" "$CL" "$LP" <<'PY'
  import re,sys
  ex=re.compile(r"release|integration|worktrees/develop|develop 워크트리|통합|auto-names|Return the same way|병합을 마치고")
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
- **AC-SMM-013** (REQ-SMM-012) — Given `$E` and M4 are committed, When the heredoc runs, Then it prints `0`. RED-now with `clear_restores_single_listing: gap`: `1` (the P-RELAUNCH phrase is not yet in `$DT`, argv[4]). Controls:
  - P-RELAUNCH present only in a `$DT` copy → `0`;
  - present only in an `$AT` copy → `1`;
  - the P-CLEAR phrase appended to a `$DT` copy raises the count by `1`;
  - an added-lines file carrying a non-canonical line mentioning `목록` raises it by `1`.
  ```bash
  git diff -U0 "$CARD_BASE"..HEAD -- "$KT" "$DT" "$WT" "$AT" "$LP" "$CL" | grep '^+[^+]' > "$S/added.txt"
  python3 - "$E" "$S/added.txt" "$KT" "$DT" "$WT" "$AT" "$LP" "$CL" <<'PY'
  import sys
  d=dict(l.rstrip("\n").split(": ",1) for l in open(sys.argv[1],encoding="utf-8") if ": " in l)
  ph={"move_adds_listing":"adds the moved-into tree's skill listing","launcher_single_listing":"gives one skill listing",
      "clear_restores_single_listing":"leaves only the new tree's listing"}
  body="".join(open(f,encoding="utf-8").read() for f in sys.argv[3:])
  bad=sum(body.count(v) for k,v in ph.items() if d.get(k)!="yes")
  if d.get("clear_restores_single_listing")!="yes": bad+= "relaunching it through the launcher" not in open(sys.argv[4],encoding="utf-8").read()
  canon=list(ph.values())+["relaunching it through the launcher"]
  bad+=sum(1 for l in open(sys.argv[2],encoding="utf-8") if ("listing" in l.lower() or "목록" in l) and not any(c in l for c in canon))
  print(bad)
  PY
  ```
- **AC-SMM-014** (REQ-SMM-013) — Given `dp2: exempt` or `dp2: covered` is recorded in progress.md §E.2, When the commands run, Then for `exempt` each of the first three counts is exactly `1` (RED-now: `0`). For `covered`, the fourth prints at least `1`, and the fifth and sixth print `0`: every integration-move `EnterWorktree(` line in `$LP` and `$CL` carries P-FLOW. RED-now for `covered`: `0`, `2`, `1` (research.md §R11).
  ```bash
  awk '/^## Integration into the release branch is self-served/{f=1;next} /^## /{f=0} f' "$KT" | grep -cF 'mid-session move exemption: the move into the integration or release worktree to merge, and the return move to the card worktree'
  grep -cF 'mid-session move exemption: the move into the integration or release worktree to merge, and the return move to the card worktree' "$LP"
  grep -cF 'mid-session move exemption: the move into the integration or release worktree to merge, and the return move to the card worktree' "$CL"
  awk '/^## Integration into the release branch is self-served/{f=1;next} /^## /{f=0} f' "$KT" | grep -cF 'move → `/clear` → re-send'
  grep -F 'EnterWorktree(' "$LP" | grep -E '통합|develop|병합' | grep -vcF 'move → `/clear` → re-send'
  grep -F 'EnterWorktree(' "$CL" | grep -E '통합|develop|병합' | grep -vcF 'move → `/clear` → re-send'
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
  - the test prints `rc=0`;
  - the two `grep -c` on `budget.txt` print `1` and `1`: the lines start with `--- PASS: TestAlwaysLoadedTokenBudget (` and `--- PASS: TestCodexContractByteCeiling (`, Go's name-plus-space separator;
  - the heredoc prints `0`;
  - the diff line count prints `0`;
  - `grep -c` prints `1`.
  ```bash
  go test ./internal/config/ -run '^TestCodexContractByteCeiling$|^TestAlwaysLoadedTokenBudget$' -count=1 -v > "$S/budget.txt"; echo "rc=$?"
  grep -c '^--- PASS: TestAlwaysLoadedTokenBudget (' "$S/budget.txt"
  grep -c '^--- PASS: TestCodexContractByteCeiling (' "$S/budget.txt"
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
  - the non-kanban `[HARD]` count heredoc prints `0`;
  - the non-`[HARD]` re-send sentence of the `/clear` section is present verbatim: `grep -cxF` prints `1` (ND11). A copy with that sentence deleted prints `0`.

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
  grep -cxF 'Where the next phase reuses a just-cleared session, the lead re-sends the full pointer instruction rather than assuming the session remembers.' "$KT"
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
  - The exempt payload `PX` is `P` with `.claude/worktrees/w1` replaced by `.claude/worktrees/develop`.
  - `PB` and `PBX` are `P` and `PX` with `"hook_event_name":"PreToolUse"`.
  - **docs-only:** block (d) prints `0` (AC-SMM-021 guarantees the range is non-empty).
  - **warn:** block (w) prints `1`, `0`, `0`, and `1` in order:
    - Kanban-mode card target;
    - variables unset in the same invocation;
    - Kanban-mode exempt target;
    - the visibility record.
  - **block:** block (b) prints `deny`, `1`, `none`, `none` in order:
    - flag on, card target;
    - sentinel prefix present;
    - flag off, card target;
    - flag on, exempt target.
  ```bash
  # (d) docs-only
  git diff --name-only "$CARD_BASE"..HEAD -- internal/hook/ | wc -l
  # (w) warn
  MOAI_KANBAN=1 go run ./cmd/moai hook post-tool <<< "$P" | jq -r '.hookSpecificOutput.additionalContext // ""' | grep -c '/clear'
  unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go run ./cmd/moai hook post-tool <<< "$P" | jq -r '.hookSpecificOutput.additionalContext // ""' | grep -c '/clear'
  MOAI_KANBAN=1 go run ./cmd/moai hook post-tool <<< "$PX" | jq -r '.hookSpecificOutput.additionalContext // ""' | grep -c '/clear'
  grep -c '^additional_context_visible: yes$' "$E"
  # (b) block
  mkdir -p "$FX/.moai/config/sections" && printf 'workflow:\n  midmove_guard:\n    enabled: true\n' > "$FX/.moai/config/sections/workflow.yaml"
  MOAI_KANBAN=1 CLAUDE_PROJECT_DIR="$FX" go run ./cmd/moai hook pre-tool <<< "$PB" | jq -r '.hookSpecificOutput.permissionDecision // "none"'
  MOAI_KANBAN=1 CLAUDE_PROJECT_DIR="$FX" go run ./cmd/moai hook pre-tool <<< "$PB" | jq -r '.hookSpecificOutput.permissionDecisionReason // ""' | grep -c '^MIDMOVE_GUARD_VIOLATION:'
  printf 'workflow:\n  midmove_guard:\n    enabled: false\n' > "$FX/.moai/config/sections/workflow.yaml"
  MOAI_KANBAN=1 CLAUDE_PROJECT_DIR="$FX" go run ./cmd/moai hook pre-tool <<< "$PB" | jq -r '.hookSpecificOutput.permissionDecision // "none"'
  printf 'workflow:\n  midmove_guard:\n    enabled: true\n' > "$FX/.moai/config/sections/workflow.yaml"
  MOAI_KANBAN=1 CLAUDE_PROJECT_DIR="$FX" go run ./cmd/moai hook pre-tool <<< "$PBX" | jq -r '.hookSpecificOutput.permissionDecision // "none"'
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
| REQ-SMM-005 | AC-SMM-002, AC-SMM-006 |
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
