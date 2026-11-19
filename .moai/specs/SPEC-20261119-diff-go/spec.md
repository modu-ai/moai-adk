---
id: SPEC-20261119-diff-go
title: "differential tests against the Go build for behaviour the Rust build keeps (C-104)"
status: draft
revision: 1
created: 2026-11-19
---

# differential tests against the Go build for behaviour the Rust build keeps (C-104)

## Intent

Why this change is needed and who it serves, in a few sentences.

## Non-goals

- What this SPEC leaves out on purpose.

## Requirements

- REQ-1: When <trigger>, the <subject> shall <behavior>.
- REQ-2 (REQ-1): IF <the step REQ-1 needs fails>, THEN the <subject> shall <safe behavior>.

## Acceptance

- AC-1 (REQ-1): <observable result that shows REQ-1 holds> · verify: `<command that checks it>`
- AC-2 (REQ-2): <observable result that shows REQ-2 holds> · verify: `<command that checks it>`

## Scope

- write: `<path the change writes>`
- read-only: `<path the change reads>`
- forbidden: `<path the change must not touch>`

## Plan

1. <first step>

## Risks

- <what could go wrong and how it would be noticed>

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | <what was decided> | <who chose it and why> | <how to reverse it> |
