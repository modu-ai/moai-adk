---
id: SPEC-ASTF-001
title: "Artifact-statelessness fixture — plan"
version: "0.1.0"
status: draft
created: 2026-08-28
---

# SPEC-ASTF-001 — Implementation Plan

The `status:` field above is the violation. Every OTHER field in that block is
in-scope-free: the D1 cleanup removes the status line and leaves `id`, `title`,
`version`, and `created` untouched, so a checker that fired on them would be
reading a wider axis than the cleanup writes.

The line below is body text, not frontmatter, and must not be flagged:

status: this is prose about a status, not a frontmatter field
