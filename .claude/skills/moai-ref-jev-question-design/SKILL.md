---
name: moai-ref-jev-question-design
description: >
  Question-design reference for the gated TypeSafe System One judgment
  capability. How to author typed questions that return useful, honest
  answers: compute the question in code, always admit a no-match answer,
  keep the supplied state small, phrase noul questions so both polarities
  are expressible. Carries the question-design rules ONLY — never the call
  path; where and how to reach the capability is the MCP catalogue's concern.

when_to_use: >
  Use when authoring or reviewing anything that composes typed questions for
  the gated judgment capability — new consumers, code review of question
  phrasing, or diagnosing answers that read as forced or uninformative.

user-invocable: false
metadata:
  version: "1.0.0"
  category: "domain"
  status: "active"
---

# Jev Question Design

The capability answers typed questions over a supplied state and returns a
typed answer with the model's probability. It generates no text and decides
nothing: the answer is a labelled model signal a person reads, never a
completion predicate, a merge approval, a queue mutation, or any other
decision that is hard to undo. Good questions keep that contract; bad
questions waste the call or manufacture false confidence. Four rules govern
question design.

## Rule 1 — Compute the question in code

The question must be computed from the caller's inputs by the caller's code,
not improvised as free-form prose at call time. A computed question is
reviewable, stable across calls, and testable: the same inputs always
compose the same question, and a change in phrasing shows up as a diff.
Compose question text with the project's language's string formatting from
named inputs — never hand-type a question into a prompt-shaped literal that
drifts from the code that acts on its answer.

## Rule 2 — Always admit a no-match answer

A choice question whose option set forces a pick manufactures confidence the
call cannot have. Every choice question carries an explicit no-match option
("none of these fit", "no listed option") so the answer can say the honest
thing. Same discipline for noul questions: ask only when both YES and NO are
answers a reader would treat as informative — a question with only one
meaningful answer is rhetoric, not a question.

## Rule 3 — Keep the state small

Input is charged per token and output is not, so cost tracks state size, not
question count. Several questions over one state in one call keep answers
mutually consistent and cost nothing extra. Supply the minimum denatured
facts the question needs — the axis being judged, the options, the counts
that matter — and nothing else. Oversize payloads are refused unsent; a
state that must be truncated to fit is a signal to redesign the question,
not to trim the payload by hand.

## Rule 4 — Phrase both noul polarities, and record which one you asked

A noul question has two phrasings for every underlying question ("is the
premise alive" vs "is the premise dead"), and the same YES means opposite
things under them. Compose the polarity deliberately, record it next to the
question id in the caller's own output, and never re-label an answer into
the opposite polarity after the fact. An answer read under the wrong
polarity is worse than no answer: it is a confident wrong signal attributed
to a model that never made that claim.

## What an absence means

Every non-answer is a typed unavailability (gate off, no credential,
endpoint unreachable, oversize, secret refused unsent). Each is a NO SIGNAL,
never a negative answer: an unavailable judgment means the consumer proceeds
without it, exactly as before the call was written. Silent defaulting —
treating an absent answer as either polarity, or as "no difference found" —
is the failure mode these rules exist to prevent.
