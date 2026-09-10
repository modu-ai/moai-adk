---
id: SPEC-FIXTURE-NEW-001
status: draft
---

# File watcher

## 1. Existing lock helper

The existing helper file carries a `//go:build !windows` constraint and is unchanged by this SPEC.

## 2. New file watcher

The watcher calls syscall.Kqueue to register file events.
