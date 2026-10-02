# Baseline of question-tool gate rounds (M0, SPEC-AUTONOMY-BATCH-GATE-001, card t1344)

Tree HEAD measured on: 517566c39 (branch WT-batch-approval-gate). The transcript store is outside the git tree and is live, so the figures below are a reading of the store taken on 2026-10-02 (about 08:1xZ), not a property of the commit.

This file states what the numbers below measure and nothing more. It carries no before/after outcome figure and no reduction claim: the earlier throwaway count in research.md R3.0 is not a baseline (see Limits).

Command: python3 m0_measure.py m0_records.tsv (script reproduced verbatim in section "Script"; scratch copy, not committed) reads every transcript under the project directories named `*moai-adk-go*` in ~/.moai/claude-profiles/moai-adk/projects whose file mtime is on or after 2026-09-26, counts distinct question-tool tool_use ids with event timestamp in 2026-09-26..2026-10-02 UTC, and classifies each as a gate round or other; an independent grep cross-check is in section "Cross-check".
Observed output: 189 distinct question-tool calls in scope, of which 43 name the gate row "Kickoff" (gate rounds) and 146 name no gate row (other); 0 duplicate tool_use ids removed; 0 calls found in sub-agent transcripts (867 sub-agent files were in the file set); exit code 0; the grep cross-check finds 206 ids, the 17 extra ids all being events dated 2026-09-24 or 2026-09-25, outside the window.
Classification method: a call is a GATE round only when its header plus question text (the visible question, not the option labels) literally names a row of the gate inventory in .claude/rules/moai/workflow/auto-semantics.md section 9, matched case-insensitively by a fixed regex per row, first match wins in the row order given in section "Scope"; every other call is "other", and no call is classified by judgement.
Limits: the rule is a name match, so it under-counts gate rounds whose text names no row (observed: "착수 승인 ... 구현을 시작할까요?" style kickoff questions fall into other) and may over-count calls that merely mention Kickoff; six of the seven row patterns matched zero calls and have no positive control, so for those rows 0 means unmeasured, not none; only the Kickoff pattern has a positive control; the store is live and held 405 project directories on this run where the card text said 406 (369 in scope); no outcome figure such as the proposal's 176 is stated or implied.

## Scope

- Store: /Users/goos/.moai/claude-profiles/moai-adk/projects (405 project directories on this run; the card text said 406). In scope: the 369 directories whose name contains `moai-adk-go` (the primary checkout and its worktree directories). This is the same selector as research.md R3.0 (`-path '*moai-adk-go*'`).
- Window, by event time: the `timestamp` field of each transcript event, 2026-09-26T00:00:00Z inclusive to 2026-10-03T00:00:00Z exclusive (UTC; the machine zone is KST). The day buckets below are UTC days.
- File prefilter, by modification time: only files with mtime on or after local midnight 2026-09-26 are opened (`find -newermt 2026-09-26`). This cannot drop an in-window event, because a file's mtime is not earlier than its last event; it does admit older events from files modified later, and those are removed by the event-time test.
- Sub-agent transcripts (`<session>/subagents/*.jsonl`) are included in the file set (867 of 1115 files per the later `find` count), and none of them contained a question-tool call.
- Duplicates: a resumed or forked session can hold the same tool call twice, so calls are de-duplicated by the tool_use `id`. Raw occurrences 189, distinct ids 189, duplicates removed 0.
- What counts as a call: a JSON event line containing `"type":"tool_use"` and `"name":"AskUserQuestion"`, taken from `message.content[]` blocks of that type and name. A line that only mentions the tool name (for example a deferred-tools attachment) is not a call.
- Gate-row names, from the gate inventory table in auto-semantics.md section 9, and the regex used for each (applied case-insensitively, first match wins in this order):
  1. `factory decide` -> rows "factory decide: kickoff / push / resume,block,unblock / abandon"
  2. `kickoff|킥오프` -> row "plan->run Kickoff" (the pattern cannot tell this row from "factory decide: kickoff approve/reject"; both are counted under the Kickoff label)
  3. `contract sign` -> row "contract signing"
  4. `sync blocking` -> row "sync blocking approval"
  5. `card pick` -> row "card pick"
  6. `plan-audit bypass|skip-audit` -> row "plan-audit bypass flags"
  7. `jev capability` -> row "Jev capability gate"

## Command (verbatim, in the order run)

Measurement (the summary is written to a scratch file; the echo records the interpreter's exit status):

```text
python3 /private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go--moai-worktrees-t1344/658f721d-4ea1-40a3-98f6-2fa1679f52da/scratchpad/m0_measure.py /private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go--moai-worktrees-t1344/658f721d-4ea1-40a3-98f6-2fa1679f52da/scratchpad/m0_records.tsv > /private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go--moai-worktrees-t1344/658f721d-4ea1-40a3-98f6-2fa1679f52da/scratchpad/m0_summary.txt 2>&1; echo "exit=$?"
```

Output of that echo: `exit=0`.

## Observed output (verbatim contents of m0_summary.txt)

```text
dirs_in_scope=369
files_scanned=1111
files_with_hit=65
raw_tool_use_occurrences=189
unique_tool_use_ids=189
duplicates_removed=0
by_class:
  other=146
  plan->run Kickoff=43
by_source:
  main=189
by_date_and_class (UTC date: gate/other):
  2026-09-26 gate=18 other=52
  2026-09-27 gate=3 other=20
  2026-09-28 gate=1 other=12
  2026-09-29 gate=7 other=11
  2026-09-30 gate=5 other=12
  2026-10-01 gate=5 other=22
  2026-10-02 gate=4 other=17
```

Totals check: 18+3+1+7+5+5+4 = 43 gate, 52+20+12+11+12+22+17 = 146 other, 189 in all.

## Positive control

Positive-control command: `awk -F'\t' '$4!="other"{print $1"\t"$2"\t"substr($5,1,110)}' m0_records.tsv | head -8` (scratch file; exit status not recorded). A real call that the rule classifies as a gate round: tool_use id `toolu_019vLHGsSmYbJDoYyyMfF3oZ`, dated 2026-09-26, classified under the row name "Kickoff" (matched token: `Kickoff`). Korean-text variants are also matched: id `toolu_013B1Yp97v2KrZx5vVJVpUPZ`, 2026-09-26, matched token `킥오프`. So the Kickoff pattern demonstrably selects real calls; a count of 0 for it would not have been silent.

The other six patterns (factory decide, contract signing, sync blocking approval, card pick, plan-audit bypass flags, Jev capability gate) matched zero calls in this scope. They have no positive control in this run, so their zero counts are "unmeasured", not "none".

Negative sample (calls classified other), command `awk -F'\t' '$4=="other"{print $2"\t"substr($5,1,90)}' m0_records.tsv | head -12` (exit status not recorded): the sample includes `착수 승인 t1224 (SPEC v0.2.2, 설계 결정 D1·D2·D6 확정) 구현을 시작할까요?` and `착수 승인 t1224 SPEC-HOOK-MATCHER-POWERSHELL-001 구현(run)을 시작할까요?`. These ask for approval to start implementation yet name no row of the inventory, so the name-match rule counts them as other. The 43 is therefore a floor under this rule, not the full number of kickoff-type questions.

## Cross-check

Independent id-level count with a different tool (grep, not the script); observed output `206`, pipeline exit status not recorded for this run (a later variant of the same pipeline that wrote the ids to a scratch file reported `exit=0`, which is the status of its last stage only, and also yielded 206 lines):

```text
cd /Users/goos/.moai/claude-profiles/moai-adk/projects && find . -name '*.jsonl' -newermt 2026-09-26 -path '*moai-adk-go*' -print0 | xargs -0 grep -h -o '"type":"tool_use","id":"[^"]*","name":"AskUserQuestion"' | sort -u | wc -l
```

The ids were then written to a scratch file and compared with the script's ids with `comm`: the script's 189 ids are all inside the grep's 206 (`comm -23` gives 0 lines) and the grep has 17 ids the script lacks (`comm -13` gives 17 lines). Timestamps of those 17, read with a `grep -F -f` over the same file set, are 2026-09-24 (8 events) and 2026-09-25 (9 events), so they fall before the window start. The two counts therefore agree once the event-time window is applied; the grep count has no window and is not a baseline.

## Script (verbatim m0_measure.py)

```python
#!/usr/bin/env python3
# M0 baseline measurement: question-tool (AskUserQuestion) tool_use calls in the
# local transcript store, classified as gate rounds or other. Read-only.
import json
import os
import re
import sys
from collections import Counter

STORE = "/Users/goos/.moai/claude-profiles/moai-adk/projects"
DIR_SUBSTR = "moai-adk-go"
WINDOW_START = "2026-09-26T00:00:00"  # event time, UTC, inclusive
WINDOW_END = "2026-10-03T00:00:00"    # event time, UTC, exclusive

import time
MTIME_FLOOR = time.mktime(time.strptime("2026-09-26", "%Y-%m-%d"))  # local midnight, lossless prefilter

# Gate-row names derived from the gate inventory table of
# .claude/rules/moai/workflow/auto-semantics.md section 9 (ten rows).
GATE_ROWS = [
    ("factory decide", re.compile(r"factory decide", re.I)),
    ("plan->run Kickoff", re.compile(r"kickoff|킥오프", re.I)),
    ("contract signing", re.compile(r"contract sign", re.I)),
    ("sync blocking approval", re.compile(r"sync blocking", re.I)),
    ("card pick", re.compile(r"card pick", re.I)),
    ("plan-audit bypass flags", re.compile(r"plan-audit bypass|skip-audit", re.I)),
    ("Jev capability gate", re.compile(r"jev capability", re.I)),
]


def classify(text):
    for name, rx in GATE_ROWS:
        if rx.search(text):
            return name
    return "other"


def main():
    out_tsv = sys.argv[1]
    dirs = sorted(d for d in os.listdir(STORE) if DIR_SUBSTR in d)
    files_scanned = 0
    files_with_hit = 0
    raw_calls = 0
    first_seen = {}  # tool_use id -> record
    for d in dirs:
        root = os.path.join(STORE, d)
        for dp, _dn, fns in os.walk(root):
            for fn in fns:
                if not fn.endswith(".jsonl"):
                    continue
                p = os.path.join(dp, fn)
                try:
                    if os.path.getmtime(p) < MTIME_FLOOR:
                        continue
                except OSError:
                    continue
                files_scanned += 1
                is_sub = "/subagents/" in p
                hit = False
                with open(p, "r", encoding="utf-8", errors="replace") as fh:
                    for line in fh:
                        if '"name":"AskUserQuestion"' not in line or '"type":"tool_use"' not in line:
                            continue
                        try:
                            ev = json.loads(line)
                        except ValueError:
                            continue
                        ts = ev.get("timestamp", "")
                        if not (WINDOW_START <= ts < WINDOW_END):
                            continue
                        msg = ev.get("message") or {}
                        content = msg.get("content")
                        if not isinstance(content, list):
                            continue
                        for blk in content:
                            if not isinstance(blk, dict):
                                continue
                            if blk.get("type") != "tool_use" or blk.get("name") != "AskUserQuestion":
                                continue
                            hit = True
                            raw_calls += 1
                            tid = blk.get("id", "")
                            if tid in first_seen:
                                continue
                            qs = (blk.get("input") or {}).get("questions") or []
                            text = " | ".join(
                                "%s %s" % (q.get("header", ""), q.get("question", ""))
                                for q in qs if isinstance(q, dict)
                            )
                            first_seen[tid] = {
                                "id": tid,
                                "date": ts[:10],
                                "sub": "sub" if is_sub else "main",
                                "row": classify(text),
                                "text": text.replace("\t", " ").replace("\n", " "),
                                "file": os.path.relpath(p, STORE),
                            }
                if hit:
                    files_with_hit += 1
    recs = sorted(first_seen.values(), key=lambda r: (r["date"], r["id"]))
    with open(out_tsv, "w", encoding="utf-8") as fo:
        for r in recs:
            fo.write("\t".join([r["id"], r["date"], r["sub"], r["row"], r["text"][:160], r["file"]]) + "\n")
    print("dirs_in_scope=%d" % len(dirs))
    print("files_scanned=%d" % files_scanned)
    print("files_with_hit=%d" % files_with_hit)
    print("raw_tool_use_occurrences=%d" % raw_calls)
    print("unique_tool_use_ids=%d" % len(recs))
    print("duplicates_removed=%d" % (raw_calls - len(recs)))
    print("by_class:")
    for k, v in sorted(Counter(r["row"] for r in recs).items()):
        print("  %s=%d" % (k, v))
    print("by_source:")
    for k, v in sorted(Counter(r["sub"] for r in recs).items()):
        print("  %s=%d" % (k, v))
    print("by_date_and_class (UTC date: gate/other):")
    days = sorted(set(r["date"] for r in recs))
    for dday in days:
        g = sum(1 for r in recs if r["date"] == dday and r["row"] != "other")
        o = sum(1 for r in recs if r["date"] == dday and r["row"] == "other")
        print("  %s gate=%d other=%d" % (dday, g, o))


if __name__ == "__main__":
    main()
```

## Limits (detail)

- Name-match classifier. It reads the question text only. A gate round whose text names no inventory row is counted as other (observed, section "Positive control"); a call that mentions "Kickoff" without being the Kickoff gate is counted as a gate round (not sampled one by one in this run, so the size of this over-count is unmeasured). The 43 and the 146 are therefore not an exact gate/non-gate split.
- The Kickoff pattern cannot separate the plan->run Kickoff row from the "factory decide: kickoff approve/reject" row; both are counted under the one label.
- Six rows have zero matches and no positive control: unmeasured for those rows.
- The store is live. Another `find` run after the script counted 1115 files in the file set against the script's 1111 scanned, so a few files were written or touched in between. The measurement is a reading at one moment.
- The 369 in-scope directories are a name selector: it takes every directory whose name contains `moai-adk-go`, including worktree directories of cards that are not this card's. It is not limited to gate-relevant sessions.
- Sub-agent transcripts were opened (867 files) and held no question-tool call; this is the observed result, and the script would have counted such a call had it existed. Whether any other transcript location exists outside these directories is not examined.
- Relation to research.md R3.0: that command counts matching lines by the first timestamp on the line, over every question-tool call (no classification), with no id de-duplication. Its daily numbers are not reproduced here and are not comparable with the figures above. This baseline does not confirm, reproduce, or replace the proposal's 176.
- Process note: the cross-check command began with `cd`, which the card instructions asked to avoid; it ran without refusal and changes nothing in the measurement.
- Tool-failure and refusal record: no command was refused by the worktree guard during this measurement; the pipeline form `find ... | xargs ... | sort | uniq` ran as written, so no substitution was made.
