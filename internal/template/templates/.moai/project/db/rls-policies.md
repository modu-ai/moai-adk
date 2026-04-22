# Row-Level Security Policies

_TBD — Define RLS policies after running `/moai db init`. Uncomment and customize the examples
below for your database engine._

---

## Supabase RLS Policies

<!--
Enable RLS on a table and define policies using Supabase's PostgreSQL-compatible syntax.

-- Enable RLS on the users table
ALTER TABLE users ENABLE ROW LEVEL SECURITY;

-- Policy: Users can only read their own row
CREATE POLICY "users_select_own"
  ON users
  FOR SELECT
  USING (auth.uid() = id);

-- Policy: Users can only update their own row
CREATE POLICY "users_update_own"
  ON users
  FOR UPDATE
  USING (auth.uid() = id);

-- Policy: Service role bypasses RLS (for admin operations)
-- Note: service_role key automatically bypasses RLS in Supabase
-->

| Table | Policy Name | Operation | Condition | Notes |
|-------|------------|-----------|-----------|-------|
| _TBD_ | _TBD_ | _TBD_ | _TBD_ | _TBD_ |

---

## PostgreSQL Policies

<!--
Standard PostgreSQL RLS policy syntax for non-Supabase deployments.

-- Enable RLS
ALTER TABLE orders ENABLE ROW LEVEL SECURITY;

-- Policy: Tenant isolation (multi-tenant schema)
CREATE POLICY "tenant_isolation"
  ON orders
  USING (tenant_id = current_setting('app.current_tenant')::uuid);

-- Policy: Admin role sees all rows
CREATE POLICY "admin_all_access"
  ON orders
  TO admin_role
  USING (true);
-->

| Table | Policy Name | Role | Operation | Condition |
|-------|------------|------|-----------|-----------|
| _TBD_ | _TBD_ | _TBD_ | _TBD_ | _TBD_ |

---

## Access Control Matrix

<!-- Map roles to permitted operations per table -->

| Table | anonymous | authenticated | service_role | admin |
|-------|-----------|---------------|--------------|-------|
| _TBD_ | _TBD_ | _TBD_ | _TBD_ | _TBD_ |

<!--
Example:
| users  | NONE     | SELECT (own row only) | ALL  | ALL   |
| posts  | SELECT   | SELECT + INSERT + UPDATE (own) | ALL | ALL |
| orders | NONE     | SELECT (own tenant)  | ALL  | ALL   |
-->
