---
name: project_migration
description: Auth middleware rewrite project context
type: project
---

Auth middleware rewrite is driven by legal/compliance requirements around session token storage.

**Why:** Legal flagged the old auth middleware for storing session tokens in a way that does not meet the new compliance requirements. This is not tech-debt cleanup.

**How to apply:** Scope decisions should favor compliance over ergonomics. Do not trade security for convenience in this area.
