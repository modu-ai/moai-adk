---
id: SPEC-FIXTURE-NEW-001
status: draft
---

# File watcher

## 1. New file watcher

The watcher calls syscall.Kqueue to register file events.
