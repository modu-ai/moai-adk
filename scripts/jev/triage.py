#!/usr/bin/env python3
"""Local-only card premise triage. Dev-only; not distributed.

Code measures, Jev judges one step, code applies the policy:

  1. read the card text from the queue (read-only)
  2. pull the symbols/paths the card names
  3. measure the tree: does the symbol still exist, did a sibling commit
     already deliver it, is that commit an ancestor of develop
  4. ask Jev whether the card's prose is refuted by that measurement
  5. print a verdict; auto-mark only above the confidence gate

Never mutates the queue. Never closes a card. With no API key it prints the
measurement and stops -- degraded, not failed.
"""
import json
import os
import re
import subprocess
import sys
import urllib.error
import urllib.request

API = os.environ.get("JEV_API", "https://api.typesafe.ai/v1/systemone")
MODEL = os.environ.get("JEV_MODEL", "jev-latest")
GATE = float(os.environ.get("JEV_GATE", "0.5"))
INTEGRATION = os.environ.get("JEV_INTEGRATION_BRANCH", "develop")
MAX_SYMBOLS = 4

PREMISE_Q = {
    "premise": {
        "type": "choice",
        "instructions": (
            "`card` states a defect someone recorded earlier. `observation` is what "
            "measuring the current tree shows. Judge the card's premise against the "
            "observation."
        ),
        "criteria": {
            "premise_dead": (
                "The observation shows the defect the card describes is not present in "
                "this tree - already repaired, never existed, or the card misread which "
                "two things are a pair."
            ),
            "premise_alive": (
                "The observation confirms the defect the card describes is present in "
                "this tree."
            ),
            "premise_partial": (
                "The observation confirms part of the card and refutes another part, so "
                "some work remains but not the work the card describes."
            ),
        },
    }
}


def run(args, limit=4000):
    """Run a command, return (rc, stdout) with stdout bounded."""
    try:
        proc = subprocess.run(
            args, capture_output=True, text=True, timeout=60, check=False
        )
    except (OSError, subprocess.TimeoutExpired) as exc:
        return 1, f"[command failed: {exc}]"
    return proc.returncode, proc.stdout[:limit]


# Sentinel returned by card_text when the queue LOOKUP itself failed — distinct
# from None (a verified absence), so callers never print "[not in queue]" for
# an unverified state (codex review gate round-7: rc=128 injection reproduced
# "[not in queue]" + exit 0 from a failed lookup).
UNVERIFIED = object()


QUEUE_LIMIT = 400000


def card_text(card_id):
    """Return the queue text for a card id, None if verified absent, or the
    UNVERIFIED sentinel when the lookup itself fails."""
    rc, out = run(["moai", "todo", "list", "--limit", "0"], limit=QUEUE_LIMIT)
    if rc != 0:
        # Lookup failure is NOT evidence of absence (AGENTS.md §1): surface it
        # on stderr so the caller's None is never read as "not in queue"
        # silently. (The --limit flag itself is supported on the current
        # binary — rc.23 verified — but a stale PATH binary would fail here,
        # and that failure must stay visible.)
        print(
            f"triage: moai todo list lookup failed (rc={rc}) — card {card_id} "
            "is UNVERIFIED, not absent",
            file=sys.stderr,
        )
        return UNVERIFIED
    if len(out) >= QUEUE_LIMIT:
        # run() truncates stdout at the limit: a card listed beyond the cut
        # would be reported absent from a PARTIAL view (codex review gate
        # round-12 reproduction). A full-length response is treated as
        # possibly-truncated → unverified, never as verified absence.
        print(
            f"triage: queue output hit the {QUEUE_LIMIT}-char read limit — "
            f"card {card_id} is UNVERIFIED (truncated view, not absent)",
            file=sys.stderr,
        )
        return UNVERIFIED
    for line in out.splitlines():
        parts = line.split("\t")
        if len(parts) >= 3 and parts[0] == card_id:
            return parts[2]
    return None


def symbols_from(text):
    """Symbols and paths a discriminator can look for, most specific first."""
    found = []
    for pattern in (
        r"`([A-Za-z_][A-Za-z0-9_./-]{3,})`",
        r"\b([a-z_]+/[a-z0-9_./-]+\.(?:go|md|yaml|sh))\b",
        r"\b([a-z][A-Za-z0-9]{6,})\(",
    ):
        for hit in re.findall(pattern, text):
            if hit not in found:
                found.append(hit)
    return found[:MAX_SYMBOLS]


def measure(card_id, text):
    """The mechanical part: what does the tree say about this card's symbols?"""
    lines = []
    syms = symbols_from(text)
    lines.append(f"symbols named by the card: {syms or '(none extractable)'}")

    present = 0
    for sym in syms:
        rc, out = run(["git", "grep", "-c", "--", sym, INTEGRATION])
        if rc == 0:
            hits = len(out.splitlines())
            present += 1 if hits else 0
            lines.append(f"  '{sym}': {hits} file(s) contain it in {INTEGRATION}")
        elif rc == 1:
            # git grep exit 1 = no matches: a real absence, reportable as 0.
            lines.append(f"  '{sym}': 0 file(s) contain it in {INTEGRATION}")
        else:
            # Lookup failure (missing integration ref, repo error — e.g. exit
            # 128): report it as unverified, never as 0 hits. An unresolved
            # query must not read as evidence the defect is gone.
            # (AGENTS.md §1: no unobserved-claim as absence.)
            lines.append(
                f"  '{sym}': LOOKUP FAILED (git exit {rc}) in {INTEGRATION}"
                " — unverified, not absent"
            )

        rc, out = run(
            ["git", "log", "--all", "--oneline", "-5", "-S", sym]
        )
        if rc == 0 and out.strip():
            lines.append(f"  '{sym}': commits that added or removed it:")
            for row in out.strip().splitlines()[:5]:
                sha = row.split()[0]
                anc, _ = run(
                    ["git", "merge-base", "--is-ancestor", sha, INTEGRATION]
                )
                if anc == 0:
                    mark = "ancestor of " + INTEGRATION
                elif anc == 1:
                    # exit 1 = genuinely not an ancestor: an answer the query
                    # actually measured, reportable as such.
                    mark = "NOT an ancestor"
                else:
                    # Any other exit (unresolved integration ref, repo error —
                    # e.g. 128) never measured the ancestry: report it as
                    # unmeasured, never as "NOT an ancestor". A failed query
                    # must not read as evidence the lineage is broken.
                    # (AGENTS.md §1: no unobserved-claim as absence.)
                    mark = (
                        f"ANCESTOR LOOKUP FAILED (git exit {anc}) in {INTEGRATION}"
                        " — unmeasured, not absent"
                    )
                lines.append(f"      {row}   [{mark}]")
        elif rc == 0:
            # The query ran and genuinely found nothing: a real absence.
            lines.append(f"  '{sym}': no commit touched this string "
                         f"(the query ran clean and returned no rows)")
        else:
            # Query failure (repo error, exit 128, timeout): the query never
            # measured anything. Keep it as an explicit unmeasured state —
            # never an absence sentence — so Jev cannot read a broken query as
            # evidence the premise is dead. (AGENTS.md §1: no unobserved-claim.)
            lines.append(f"  '{sym}': HISTORY LOOKUP FAILED (git exit {rc})"
                         " — unmeasured, not absent")

    rc, out = run(["git", "log", "--all", "--oneline", "-5", "--grep", card_id])
    if rc == 0:
        lines.append(f"commits naming {card_id}: {out.strip() or '(none)'}")
    else:
        lines.append(f"commits naming {card_id}: LOOKUP FAILED (git exit {rc})"
                     " — unmeasured, not none")
    lines.append(
        "note: a delivering commit often carries a SIBLING card id, so the line above "
        "being empty does not mean the work was never done."
    )
    return f"{present}/{len(syms)}" if syms else "0/0", "\n".join(lines)


def ask(card, observation, key):
    body = json.dumps(
        {
            "model": MODEL,
            "state": {"card": card, "observation": observation},
            "questions": PREMISE_Q,
        }
    ).encode()
    req = urllib.request.Request(
        API,
        data=body,
        headers={"Authorization": f"Bearer {key}", "Content-Type": "application/json", "User-Agent": "moai-adk-jev/1.0"},
    )
    with urllib.request.urlopen(req, timeout=60) as resp:
        return json.loads(resp.read())


def load_key():
    key = os.environ.get("TYPESAFE_API_KEY", "")
    if key:
        return key
    path = os.environ.get(
        "JEV_ENV_FILE", os.path.expanduser("~/.moai/.env.typesafe")
    )
    try:
        with open(path, encoding="utf-8") as handle:
            for line in handle:
                if line.startswith("TYPESAFE_API_KEY="):
                    return line.split("=", 1)[1].strip()
    except OSError:
        return ""
    return ""


def main(argv):
    if not argv:
        print("usage: triage.py <card-id> [<card-id> ...]", file=sys.stderr)
        return 2
    key = load_key()
    verbose = os.environ.get("JEV_VERBOSE") == "1"
    for card_id in argv:
        text = card_text(card_id)
        if text is UNVERIFIED:
            print(f"{card_id}  [unverified — queue lookup failed]")
            continue
        if text is None:
            print(f"{card_id}  [not in queue]")
            continue
        ratio, observation = measure(card_id, text)
        if verbose:
            print(f"--- {card_id} observation ---\n{observation}\n")
        if not key:
            print(f"{card_id}  [no key: measurement only]  symbols present {ratio}")
            continue
        try:
            resp = ask(text[:6000], observation, key)
        except (urllib.error.URLError, OSError) as exc:
            print(f"{card_id}  [api unavailable: {exc}]  symbols present {ratio}")
            continue
        answer = resp["answers"]["premise"]
        gate = "auto" if answer["confidence"] >= GATE else "ask human"
        print(
            f"{card_id}  {answer['choice']:14s} conf {answer['confidence']:.2f}  "
            f"[{gate}]  symbols present {ratio}  "
            f"tokens {resp['usage']['input_tokens']}"
        )
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
