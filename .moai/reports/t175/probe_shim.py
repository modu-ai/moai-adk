#!/usr/bin/env python3
"""t175 plan-phase §A probe: does the z.ai anthropic-compat shim honor
reasoning-depth controls?

Bounded: 3 POST calls to $ANTHROPIC_BASE_URL/v1/messages using the session's
own ANTHROPIC_AUTH_TOKEN (never printed). Only the `usage` block and any
thinking-bearing content-block types are observed, per AC-MTP-032b.

Probes:
  P1 thinking budget 2048  (small — Claude 'low'/'high' low end)
  P2 top-level reasoning_effort "max" (z.ai extension hypothesis; no thinking)
  P3 thinking budget 32768 (max-equivalent budget)
"""
import json
import os
import urllib.request

BASE = os.environ["ANTHROPIC_BASE_URL"].rstrip("/")
TOKEN = os.environ["ANTHROPIC_AUTH_TOKEN"]
URL = BASE + "/v1/messages"


def call(label, body):
    req = urllib.request.Request(
        URL,
        data=json.dumps(body).encode(),
        headers={
            "content-type": "application/json",
            "x-api-key": TOKEN,
            "authorization": "Bearer " + TOKEN,
            "anthropic-version": "2023-06-01",
        },
        method="POST",
    )
    try:
        with urllib.request.urlopen(req, timeout=90) as r:
            d = json.loads(r.read())
    except urllib.error.HTTPError as e:
        body_txt = e.read().decode()[:300]
        print(f"{label}: HTTP {e.code} {body_txt}")
        return
    usage = d.get("usage", {})
    stop = d.get("stop_reason")
    kinds = [b.get("type") for b in d.get("content", [])]
    print(f"{label}: usage={json.dumps(usage)} stop={stop} blocks={kinds}")


msg = [{"role": "user", "content": "Reply with the single word OK."}]

call(
    "P1(budget=2048)",
    {"model": "glm-5.3", "max_tokens": 4096,
     "thinking": {"type": "enabled", "budget_tokens": 2048}, "messages": msg},
)
call(
    "P2(reasoning_effort=max)",
    {"model": "glm-5.3", "max_tokens": 4096,
     "reasoning_effort": "max", "messages": msg},
)
call(
    "P3(budget=32768)",
    {"model": "glm-5.3", "max_tokens": 40000,
     "thinking": {"type": "enabled", "budget_tokens": 32768}, "messages": msg},
)
