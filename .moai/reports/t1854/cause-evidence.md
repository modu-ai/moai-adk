# t1854 cause evidence (RED, pre-fix)

Root cause:
1. `auditReadyRecorded` (internal/homestate/card_audit_kickoff.go) splits each §E.1 line with `strings.Cut(strings.TrimSpace(line), ":")` and counts only lines whose key is exactly `audit_ready`.
2. A markdown bullet line `- audit_ready: true` (or `* audit_ready: true`) has key `- audit_ready`, so it is silently ignored.
3. `auditKickoffRefusal` then reports "the plan phase records no audit-ready status (progress.md §E.1)", with no hint that the signal was present in bullet form or that a look-alike key was ignored.

Command: `go test ./internal/homestate -run TestAuditReadyBulletForms -count=1`
Tree: HEAD fb749d372 plus the uncommitted new test file internal/homestate/card_audit_ready_bullet_test.go (no production change yet).

Verbatim output:

```
--- FAIL: TestAuditReadyBulletForms (9.93s)
    --- FAIL: TestAuditReadyBulletForms/dash_bullet (1.48s)
        card_audit_ready_bullet_test.go:41: want acceptance, got refusal: card evidence refused: audit kickoff refused: the plan phase records no audit-ready status (progress.md §E.1)
    --- FAIL: TestAuditReadyBulletForms/star_bullet (1.45s)
        card_audit_ready_bullet_test.go:41: want acceptance, got refusal: card evidence refused: audit kickoff refused: the plan phase records no audit-ready status (progress.md §E.1)
    --- FAIL: TestAuditReadyBulletForms/bullet_inside_fenced_yaml (0.93s)
        card_audit_ready_bullet_test.go:41: want acceptance, got refusal: card evidence refused: audit kickoff refused: the plan phase records no audit-ready status (progress.md §E.1)
    --- FAIL: TestAuditReadyBulletForms/look-alike_suffix_key (0.74s)
        card_audit_ready_bullet_test.go:49: refusal "card evidence refused: audit kickoff refused: the plan phase records no audit-ready status (progress.md §E.1)" does not name "ignored look-alike key \"audit_ready_note\""
    --- FAIL: TestAuditReadyBulletForms/look-alike_hyphen_key (0.67s)
        card_audit_ready_bullet_test.go:49: refusal "card evidence refused: audit kickoff refused: the plan phase records no audit-ready status (progress.md §E.1)" does not name "ignored look-alike key \"audit-ready\""
    --- FAIL: TestAuditReadyBulletForms/look-alike_capitalised_key (0.67s)
        card_audit_ready_bullet_test.go:49: refusal "card evidence refused: audit kickoff refused: the plan phase records no audit-ready status (progress.md §E.1)" does not name "ignored look-alike key \"Audit_Ready\""
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/homestate	10.306s
FAIL
```
