---
id: {{id}}
title: "{{title}}"
status: draft
revision: 1
created: {{created}}
{{parent}}---

# {{title}}

## Intent

Why this change is needed, in one or two sentences.

## Requirements

<!--
Rule 1: exactly one `shall` per requirement line; a second behavior goes on its own REQ line (spec check refuses it as req_gears).
Rule 2: every When requirement needs an IF/THEN requirement that names it in parentheses, or its line ends with "· failure: none — reason" (spec check refuses it as req_failure_path).
-->
- REQ-1: When <trigger>, the <subject> shall <behavior>.
- REQ-2 (REQ-1): IF <the step REQ-1 needs fails>, THEN the <subject> shall <safe behavior>.

## Acceptance

- AC-1 (REQ-1, REQ-2): <observable result that shows REQ-1 and REQ-2 hold> · verify: `<command that checks it>`

## Scope

- write: `<path the change writes>`

## Plan

1. <first step>
