# Audit-agent eval: incidents that mattered

This folder grades the audit agent (`moai-audit`) on bugs that actually blocked pull requests of
this repository, not on invented ones. Every case replays a real audit FAIL: the repository as
it was at review time (the `head` and `base` the audit verdict named) and a one-sentence answer
key that states the bug the audit's blocking item found. Clean merged pull requests whose audits
passed are the controls: a finding on a control is over-flagging.

## Files

- `cases-tuning.json` — six incidents used while tuning the audit prompt or the agent.
- `cases-heldout.json` — six incidents never looked at while tuning; they measure the result.
- `controls.json` — merged PASS-audit pull requests, one per size bucket of the incidents
  (bucket = total changed lines within ±50%, else the nearest merged control, named in
  `matched_case` and `size_bucket` on each row).

The two case files never share an id: a case scored twice measures nothing, and the tool refuses
it (`run` and `score`).

## Adding a case

Read a FAIL audit comment, pick the blocking item that mattered, and write the case yourself —
the tool never writes answer keys. A case names its pull request (`pr`), the audit comment
(`audit_comment`), the review-time `head` and `base` (from the verdict line), and one sentence
(`answer_key`) that states the bug so a reader who never saw the audit could recognize a finding
that describes it. Verify both refs exist locally (`git cat-file -e <sha>^{commit}`); a case
whose ref the repository does not hold is reported `missing_ref` and skipped, never silently.
Put the case in the held-out file only when you are not tuning against it; otherwise it belongs
in the tuning file. Keep the files balanced (the split today is 6 and 6).

## Running and scoring

```sh
# Build the audit prompts for a set (dry-run: no model, no network; answer keys stay out).
python3 tools/audit_eval.py run --set tuning --out target/audit-eval/prompts

# Score a findings file against a set: recall of raw findings, recall of the final verdict,
# and the over-flagging rate on the controls.
python3 tools/audit_eval.py score --set heldout --findings findings.json
```

A findings file is a JSON list of `{ "case_id", "findings": ["…"], "verdict": "PASS"|"FAIL" }`
rows, one per case (controls included as case ids `ctl-…`). A finding counts as describing the
bug only when at least half of the answer key's distinctive tokens appear in it — the threshold
is a constant in the tool (`TOKEN_FRACTION`); it is a ruler for comparing runs, not a truth.

## When to stop tuning

Stop at the change whose tuning score rises while the held-out score does not: that next gain is
overfitting, and the held-out file is the only score that counts. After such a stop, move the
cases you tuned on back to the tuning file, add fresh incidents to the held-out file (from new
FAIL audits), and re-measure. Never edit an answer key to make a score move; edit it only when
the sentence is wrong about the bug, and say so in the commit.
