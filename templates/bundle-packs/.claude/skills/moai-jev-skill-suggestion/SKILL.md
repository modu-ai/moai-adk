---
name: moai-jev-skill-suggestion
description: >
  Guidance for a display-only skill-suggestion capability: how a caller
  should suggest a skill to the operator, how the emitted answer is read,
  and what must never happen (selection, auto-loading, judgment). The
  capability gate ships off; this skill carries mechanism and contract ONLY
  — never the call path; where and how to reach the capability is the
  catalogue's concern.

when_to_use: >
  Use when designing, reviewing, or wiring anything that suggests a skill to
  the operator from a model signal — consumer authors, reviewers of
  display-only contracts, or anyone diagnosing a suggestion that appeared to
  select, load, or decide on its own.

user-invocable: true
metadata:
  version: "1.0.0"
  category: "reference"
  status: "active"
---

# Skill Suggestion, Display-Only

## The gate comes first

The capability is gated behind `workflow.jev.enabled`, and the shipped
default is off. Confirm the gate before anything else. With the gate off, a
caller constructs no request, makes no network call, and proceeds with no
signal: an absent signal is NO SIGNAL — the same as "we did not ask" — never
a negative answer, and never a reason to change course.

Three readiness surfaces exist for the on state: the gate key itself, the
credentials file referenced by convention (`~/.moai/.env.typesafe`, read by
the credential loader — never inlined, never copied into settings), and the
doctor readiness check (a TCP-only probe that reports gate, credential, and
reachability in three lines without sending any judgment request). A caller
that cannot confirm all three proceeds as if the gate were off.

The reachable surface itself lives in the product catalogue, not here. This
skill owns mechanism and contract only; the catalogue owns where the call
lives.

## What a suggestion is

A suggestion is a signal a person reads: a labelled model answer that names
a skill that might be worth loading and carries a confidence, and it does
nothing else. Four things it is not:

- Not a selection. The existing selector keeps its selection authority; a
  suggestion never flips it.
- Not an auto-load. No skill content is loaded, attached, or injected
  because a suggestion was emitted.
- Not a judgment. Not a completion decision, not a merge approval, not a
  queue mutation — anything hard to undo stays with people.
- Not a code-path input. Nothing branches on the answer; a human reads it,
  and the reading is the only consumer.

## The loading mechanisms rejected here

Two mechanisms were evaluated and rejected; recording them keeps them from
returning quietly:

- Hiding the listing and injecting a chosen SKILL.md body directly. This
  breaks the display-only chain: a skill that never appeared in context has
  no reading human, and a direct content injection is precisely the
  auto-load the contract forbids.
- An ambient function-hook wiring that evaluates a suggestion on every
  prompt. It commercializes a model call per prompt on an unmeasured
  behavior, fragments the template distribution, and rides an
  early-experiment surface that can shift without notice. Not adopted.
  Revisit only with measured evidence that a constant-answer baseline has
  been beaten — the measurement bar belongs to a separate track, not to
  this skill.

## Operational checklist (for when a caller exists)

Every item below is guidance conditioned on a caller existing — it ships
no behavior:

1. Log each decision, and announce once per session that suggestion logging
   is active — one visible line per session, not a chorus per call.
2. Filter non-task origins. A prompt typed by the operator is not a task
   event; only genuine task origins may produce a suggestion.
3. Filter invocation-disabled skills. A skill marked
   disable-model-invocation stays invisible to suggestions; the filter runs
   before the call, never after the answer.
4. Cache the roster and skill bodies per session; re-reading listings and
   bodies on every prompt is the waste a session cache exists to prevent.
5. Guard against duplicate injection: a skill already present in context is
   never suggested to appear again. Check what is loaded before showing a
   name.
6. Canonicalize names: a display name and a directory id are two spellings
   of one skill; map both through one canonical form before deduplicating,
   logging, or displaying.

## Question design is a separate discipline

The suggestion call is a judgment over a typed question, and typed
questions carry their own rules — compute the question in code, always
admit a no-match answer, keep the supplied state small, phrase both
polarities. Those rules live in the question-design reference skill; this
skill does not restate them. The measurement regime above the surface is
likewise owned elsewhere: this skill records the contract, never baseline
numbers.

## What this skill is not

This skill is not a capability. It ships no caller, registers no command,
adds no hook, and changes no setting. It is not a verdict: suggestion
answers never close cards, approve merges, or mutate queues. And it is not
an availability promise: the gate ships off, and nothing in this skill makes
a suggestion surface exist — the checklist above governs only if a caller is
ever built elsewhere, and until then an absent suggestion is just NO SIGNAL.
