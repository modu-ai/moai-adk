# scripts/jev — local-only TypeSafe (Jev) tooling

**Dev-only. Not distributed.** Nothing here is mirrored into
`internal/template/templates/`, so `moai init` / `moai update` never deploy it to a
user project. This directory follows the same local-only convention as
`scripts/ci-watch/` and `scripts/ci-autofix/`.

## What this is for

Jev returns a *typed judgment* (one of a fixed set, or a probability) plus a
confidence — it does not write prose. These scripts use it for the one step in
card triage that code cannot do: deciding whether a card's prose is refuted by
what the tree actually shows.

Everything mechanical stays in the script: the git measurements are run locally
and their output is what gets judged.

## Credential

The key lives in **one place, outside the repository**:

```
~/.moai/.env.typesafe        # chmod 600, single line: TYPESAFE_API_KEY=apikey_...
```

Never put the key in `.claude/settings*.json`, `.moai/config/`, a template, or a
commit. The scripts read that file and nothing else; with no key they print what
they measured and exit 0 (degraded, not failed) — the same fail-open posture
`glm_audit` uses.

## Usage

```bash
# one card: run the mechanical discriminator, then judge the premise
scripts/jev/triage.sh t784

# several cards
scripts/jev/triage.sh t784 t787 t790

# raw question against arbitrary text (debugging / experiments)
echo "some state" | scripts/jev/ask.sh noul "Does this report show a command and its output?"
```

`triage.sh` is **read-only**: it never calls `moai todo`, never closes a card,
never writes to the queue. It prints a verdict for a human to act on.

## Reading the verdict

```
t784  premise_dead   conf 0.94   [auto]      symbol hits: 0/3   sibling: 923353c66 (ancestor)
t787  premise_alive  conf 0.31   [ask human] symbol hits: 3/3   sibling: -
```

- `[auto]` — confidence ≥ 0.50. In the measured sample every answer above this
  line was correct (11/11 over 16 cards); every error sat at ≤ 0.30.
- `[ask human]` — below the gate. Treat as "not measured", not as "alive".

**The gate is provisional.** It was fitted on 16 cards whose text had been
translated to English by hand, and TypeSafe's own model card says non-English
accuracy is lower than English. Re-measure on Korean card text before trusting
the number — that is the first task of card t943.

## Cost

Charged per **input** token; output tokens are free. The model card lists
`$42 per Btok / $0.042 per Mtok`. One card triage sends roughly 600 input
tokens, so ~2.5e-5 USD per card at that rate. A 25-card batch is well under a
cent. Rate limits at the time of writing: 250,000 tokens/s and 1,200
requests/min, and the model card notes they change without notice.

## What is deliberately NOT here

- No queue mutation. `moai todo done` stays a human act.
- No auto-close, no auto-drop, no reordering.
- No report auditing. That experiment scored 5/8 on held-out data and is parked;
  see `.moai/reports/jev-lead-integration-20260919.md` §6.
