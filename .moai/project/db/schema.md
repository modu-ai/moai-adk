---
engine: _TBD_
orm: _TBD_
last_synced_at: _TBD_
manifest_hash: _TBD_
---

# Database Schema

_TBD — Run `/moai db init` to configure the database engine and ORM, then edit this file or let
the auto-sync hook populate it from your migration files._

---

## Tables

<!-- For NoSQL databases, replace this section with ## Collections -->

| Table | Description |
|-------|-------------|
| _TBD_ | _TBD_ |

<!--
Example:
| users | Core user account table — authentication identity |
| posts | User-authored content items |
| comments | Threaded comment entries linked to posts |
-->

---

## Relationships

<!-- Cardinality notation: 1:1, 1:N, N:M -->

| From | To | Cardinality | FK Column | Notes |
|------|----|-------------|-----------|-------|
| _TBD_ | _TBD_ | _TBD_ | _TBD_ | _TBD_ |

<!--
Example:
| users | posts    | 1:N | posts.user_id    | A user owns many posts |
| posts | comments | 1:N | comments.post_id | A post has many comments |
| users | roles    | N:M | user_roles table | Via junction table |
-->

---

## Indexes

<!-- List standalone and composite indexes -->

| Table | Columns | Type | Purpose |
|-------|---------|------|---------|
| _TBD_ | _TBD_ | _TBD_ | _TBD_ |

<!--
Example:
| users | email          | UNIQUE  | Enforce unique emails for login |
| posts | (user_id, created_at) | COMPOSITE | Paginated user post queries |
| posts | title          | GIN/FTS | Full-text search on post titles |
-->

---

## Constraints

<!-- UNIQUE, CHECK, EXCLUSION, NOT NULL (non-obvious cases) -->

| Table | Constraint | Type | Definition |
|-------|-----------|------|-----------|
| _TBD_ | _TBD_ | _TBD_ | _TBD_ |

<!--
Example:
| users | users_email_unique | UNIQUE | email must be unique |
| posts | posts_status_check | CHECK  | status IN ('draft', 'published', 'archived') |
| bookings | no_overlap        | EXCLUSION | daterange(start_at, end_at) with &&  |
-->
