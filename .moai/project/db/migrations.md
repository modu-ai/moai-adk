# Migrations

_TBD — Populate this file as migrations are created and applied. The `Applied Migrations` table
is partially auto-updated by the `moai-domain-db-docs` hook when migration files are detected._

---

## Applied Migrations

| Filename | Applied At | Checksum | Summary |
|----------|-----------|----------|---------|
| _TBD_ | _TBD_ | _TBD_ | _TBD_ |

<!--
Example rows:
| 20240101_001_create_users.sql     | 2024-01-01T10:00:00Z | sha256:abc123 | Initial users table |
| 20240115_002_add_posts_table.sql  | 2024-01-15T14:30:00Z | sha256:def456 | Add posts with user FK |
| 20240201_003_add_email_index.sql  | 2024-02-01T09:15:00Z | sha256:ghi789 | Perf index on users.email |
-->

---

## Pending Migrations

List migrations that exist in the codebase but have not yet been applied to production.

| Filename | Created At | Description | Blocking? |
|----------|-----------|-------------|-----------|
| _TBD_ | _TBD_ | _TBD_ | _TBD_ |

<!--
Example:
| 20240301_004_add_comments.sql | 2024-03-01 | Add comments table | No |
| 20240310_005_rls_enable.sql   | 2024-03-10 | Enable RLS on users | Yes — requires DBA review |
-->

---

## Rollback Notes

Document rollback procedures for each migration that is difficult or non-trivial to reverse.

| Migration | Risk Level | Rollback Steps | Data Loss? |
|-----------|-----------|----------------|------------|
| _TBD_ | _TBD_ | _TBD_ | _TBD_ |

<!--
Example:
| 20240201_003_add_email_index.sql | Low    | DROP INDEX users_email_idx; | No |
| 20240310_005_rls_enable.sql      | High   | Disable RLS, restore original policies | No |
| 20240401_006_drop_old_column.sql | Critical | Cannot roll back — column data lost | YES |
-->
