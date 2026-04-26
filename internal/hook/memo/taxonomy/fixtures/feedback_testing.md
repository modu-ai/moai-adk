---
name: feedback_testing
description: Feedback about test isolation approach
type: feedback
---

Do not mock the database in integration tests.

**Why:** We got burned last quarter when mocked tests passed but the prod migration failed. Mock/prod divergence masked a broken migration.

**How to apply:** Integration tests must hit a real database, not mocks. Use t.TempDir() for file-based stores.
