#!/usr/bin/env python3
"""Route a lane's question to a decision class. Dev-only; not distributed.

A lane blocks and asks. The lead must decide three things fast:
  - can this be answered from doctrine already on disk, or does it need a
    measurement the lane has not taken?
  - is the decision the lead's, or does it belong to the operator?
  - if the lead's: is it reversible enough to answer immediately?

Those are judgments over prose, which is what Jev is for. The policy that
turns the judgment into an action stays here in code, and the hard boundaries
below are never delegated.

Input: the question text on stdin (the lane's message, trimmed).
Output: one line per axis plus a recommended class. Nothing is executed.
"""
import json
import os
import sys
import urllib.error
import urllib.request

API = os.environ.get("JEV_API", "https://api.typesafe.ai/v1/systemone")
MODEL = os.environ.get("JEV_MODEL", "jev-latest")
GATE = float(os.environ.get("JEV_GATE", "0.5"))

QUESTIONS = {
    "owner": {
        "type": "choice",
        "instructions": (
            "A worker on a software team has paused and asked its coordinator a "
            "question. Who must decide the thing being asked?"
        ),
        "criteria": {
            "lead": (
                "The coordinator can decide it from rules and evidence already "
                "written down: scope of a task, which check to run, ordering of "
                "work, how to word something, whether a measurement is sufficient."
            ),
            "operator": (
                "It changes what the team is working on or commits the project to "
                "something: adding/closing/dropping a work item, changing a "
                "user-facing behaviour, reversing an earlier human decision, "
                "spending money, or anything the worker frames as needing approval."
            ),
            "worker": (
                "The worker can answer it themselves by measuring; they are asking "
                "for permission they do not need, or for a fact their own tree holds."
            ),
        },
    },
    "needs_measurement": {
        "type": "noul",
        "instructions": (
            "Does answering this question require a NEW measurement that nobody has "
            "taken yet - running a command, reading a file, reproducing a failure? "
            "Answer no when the question is answerable from stated rules or from "
            "evidence the message already contains."
        ),
    },
    "reversible": {
        "type": "noul",
        "instructions": (
            "If the coordinator answers this immediately and the answer turns out to "
            "be wrong, is the consequence cheap to undo - an edit reverted, a check "
            "re-run, a message corrected? Answer no when a wrong answer would land "
            "something hard to take back: a merge, a push, a deletion, a released "
            "behaviour, or work thrown away."
        ),
    },
}


def ask(state, key):
    body = json.dumps(
        {"model": MODEL, "state": state, "questions": QUESTIONS}
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
    path = os.environ.get("JEV_ENV_FILE", os.path.expanduser("~/.moai/.env.typesafe"))
    try:
        with open(path, encoding="utf-8") as handle:
            for line in handle:
                if line.startswith("TYPESAFE_API_KEY="):
                    return line.split("=", 1)[1].strip()
    except OSError:
        return ""
    return ""


def classify(answers):
    """Policy in code: map the three judgments to one recommended class."""
    owner = answers["owner"]
    conf = owner["confidence"]
    needs = answers["needs_measurement"]["noul"]
    rev = answers["reversible"]["noul"]

    if conf < GATE:
        return "ASK-OPERATOR", "owner unclear — do not guess who decides"
    if owner["choice"] == "operator":
        return "ASK-OPERATOR", "the decision changes scope or commits the project"
    if owner["choice"] == "worker":
        return "RETURN-TO-LANE", "the lane can measure this itself"
    if needs >= GATE:
        return "MEASURE-FIRST", "answerable only after a new measurement"
    if rev < GATE:
        return "LEAD-DECIDE-CAREFULLY", "lead's call, but a wrong answer is costly"
    return "LEAD-ANSWER-NOW", "lead's call and cheap to correct"


def main():
    text = sys.stdin.read().strip()
    if not text:
        print("usage: route.sh < question.txt", file=sys.stderr)
        return 2
    key = load_key()
    if not key:
        print("[no key] route unavailable — decide by reading the doctrine")
        return 0
    try:
        resp = ask(text[:12000], key)
    except (urllib.error.URLError, OSError) as exc:
        print(f"[api unavailable: {exc}] — decide by reading the doctrine")
        return 0
    answers = resp["answers"]
    cls, why = classify(answers)
    owner = answers["owner"]
    print(f"owner              {owner['choice']:9s} conf {owner['confidence']:.2f}")
    print(f"needs_measurement  {answers['needs_measurement']['noul']:.2f}")
    print(f"reversible         {answers['reversible']['noul']:.2f}")
    print(f"=> {cls}  ({why})")
    print(f"   tokens {resp['usage']['input_tokens']}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
