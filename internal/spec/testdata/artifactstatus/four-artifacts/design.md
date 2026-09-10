---
id: SPEC-ASTF-001
status:draft
---

# SPEC-ASTF-001 — Design

This artifact writes `status:` with NO space after the colon. The counting
predicate the corpus figures were measured with (`^status:[[:space:]]`) does
not select this line; the lint rule deliberately does. The corpus carries zero
lines of this shape today, so the widening changes no present count — it is
there for the line a future author writes, which is what a recurrence guard is
for.
