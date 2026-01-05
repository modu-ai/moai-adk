---
name: row-level-security
description: RLS policies for multi-tenant data isolation and access control
parent-skill: moai-platform-supabase
version: 1.0.0
updated: 2026-01-06
---

# Row-Level Security (RLS) Module

## Overview

Row-Level Security provides automatic data isolation at the database level, ensuring users can only access data they are authorized to see.

## Basic Setup

Enable RLS on a table:

```sql
ALTER TABLE projects ENABLE ROW LEVEL SECURITY;
```

## Policy Types

RLS policies can be created for specific operations:

- SELECT: Controls read access
- INSERT: Controls creation
- UPDATE: Controls modification
- DELETE: Controls removal
- ALL: Applies to all operations

## Basic Tenant Isolation

### JWT-Based Tenant Isolation

Extract tenant ID from JWT claims:

```sql
CREATE POLICY "tenant_isolation" ON projects FOR ALL
  USING (tenant_id = (auth.jwt() ->> 'tenant_id')::UUID);
```

### Owner-Based Access

Restrict access to resource owners:

```sql
CREATE POLICY "owner_access" ON projects FOR ALL
  USING (owner_id = auth.uid());
```

## Hierarchical Access Patterns

### Organization Membership

Allow access based on organization membership:

```sql
CREATE POLICY "org_member_select" ON organizations FOR SELECT
  USING (id IN (SELECT org_id FROM org_members WHERE user_id = auth.uid()));
```

### Role-Based Modification

Restrict modifications to specific roles:

```sql
CREATE POLICY "org_admin_modify" ON organizations FOR UPDATE
  USING (id IN (
    SELECT org_id FROM org_members
    WHERE user_id = auth.uid() AND role IN ('owner', 'admin')
  ));
```

### Cascading Project Access

Grant project access through organization membership:

```sql
CREATE POLICY "project_access" ON projects FOR ALL
  USING (org_id IN (SELECT org_id FROM org_members WHERE user_id = auth.uid()));
```

## Service Role Bypass

Allow service role to bypass RLS for server-side operations:

```sql
CREATE POLICY "service_bypass" ON organizations FOR ALL TO service_role USING (true);
```

## Multi-Tenant SaaS Schema

### Complete Schema Setup

```sql
-- Organizations (tenants)
CREATE TABLE organizations (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  name TEXT NOT NULL,
  slug TEXT UNIQUE NOT NULL,
  plan TEXT DEFAULT 'free' CHECK (plan IN ('free', 'pro', 'enterprise')),
  settings JSONB DEFAULT '{}',
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Organization members with roles
CREATE TABLE organization_members (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
  user_id UUID NOT NULL,
  role TEXT NOT NULL CHECK (role IN ('owner', 'admin', 'member', 'viewer')),
  joined_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(organization_id, user_id)
);

-- Projects within organizations
CREATE TABLE projects (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  owner_id UUID NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW()
);
```

### Enable RLS on All Tables

```sql
ALTER TABLE organizations ENABLE ROW LEVEL SECURITY;
ALTER TABLE organization_members ENABLE ROW LEVEL SECURITY;
ALTER TABLE projects ENABLE ROW LEVEL SECURITY;
```

### Comprehensive RLS Policies

```sql
-- Organization read access
CREATE POLICY "org_member_select" ON organizations FOR SELECT
  USING (id IN (SELECT organization_id FROM organization_members WHERE user_id = auth.uid()));

-- Organization admin update
CREATE POLICY "org_admin_update" ON organizations FOR UPDATE
  USING (id IN (SELECT organization_id FROM organization_members
    WHERE user_id = auth.uid() AND role IN ('owner', 'admin')));

-- Project member access
CREATE POLICY "project_member_access" ON projects FOR ALL
  USING (organization_id IN (SELECT organization_id FROM organization_members WHERE user_id = auth.uid()));

-- Member management (admin only)
CREATE POLICY "member_admin_manage" ON organization_members FOR ALL
  USING (organization_id IN (SELECT organization_id FROM organization_members
    WHERE user_id = auth.uid() AND role IN ('owner', 'admin')));
```

## Helper Functions

### Check Organization Membership

```sql
CREATE OR REPLACE FUNCTION is_org_member(org_id UUID)
RETURNS BOOLEAN AS $$
BEGIN
  RETURN EXISTS (
    SELECT 1 FROM organization_members
    WHERE organization_id = org_id AND user_id = auth.uid()
  );
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
```

### Check Organization Role

```sql
CREATE OR REPLACE FUNCTION has_org_role(org_id UUID, required_roles TEXT[])
RETURNS BOOLEAN AS $$
BEGIN
  RETURN EXISTS (
    SELECT 1 FROM organization_members
    WHERE organization_id = org_id
      AND user_id = auth.uid()
      AND role = ANY(required_roles)
  );
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
```

### Usage in Policies

```sql
CREATE POLICY "project_admin_delete" ON projects FOR DELETE
  USING (has_org_role(organization_id, ARRAY['owner', 'admin']));
```

## Performance Optimization

### Index for RLS Queries

Create indexes on foreign keys used in RLS policies:

```sql
CREATE INDEX idx_org_members_user ON organization_members(user_id);
CREATE INDEX idx_org_members_org ON organization_members(organization_id);
CREATE INDEX idx_projects_org ON projects(organization_id);
```

### Materialized View for Complex Policies

For complex permission checks, use materialized views:

```sql
CREATE MATERIALIZED VIEW user_accessible_projects AS
SELECT p.id as project_id, om.user_id, om.role
FROM projects p
JOIN organization_members om ON p.organization_id = om.organization_id;

CREATE INDEX idx_uap_user ON user_accessible_projects(user_id);

REFRESH MATERIALIZED VIEW CONCURRENTLY user_accessible_projects;
```

## Testing RLS Policies

### Test as Authenticated User

```sql
SET request.jwt.claim.sub = 'user-uuid-here';
SET request.jwt.claims = '{"role": "authenticated"}';

SELECT * FROM projects;  -- Returns only accessible projects
```

### Verify Policy Restrictions

```sql
-- Should fail if not a member
INSERT INTO projects (organization_id, name, owner_id)
VALUES ('non-member-org-id', 'Test', auth.uid());
```

## Common Patterns

### Public Read, Owner Write

```sql
CREATE POLICY "public_read" ON posts FOR SELECT USING (true);
CREATE POLICY "owner_write" ON posts FOR INSERT WITH CHECK (author_id = auth.uid());
CREATE POLICY "owner_update" ON posts FOR UPDATE USING (author_id = auth.uid());
CREATE POLICY "owner_delete" ON posts FOR DELETE USING (author_id = auth.uid());
```

### Draft vs Published

```sql
CREATE POLICY "published_read" ON articles FOR SELECT
  USING (status = 'published' OR author_id = auth.uid());
```

### Time-Based Access

```sql
CREATE POLICY "active_subscription" ON premium_content FOR SELECT
  USING (
    EXISTS (
      SELECT 1 FROM subscriptions
      WHERE user_id = auth.uid()
        AND expires_at > NOW()
    )
  );
```

## Context7 Query Examples

For latest RLS documentation:

Topic: "row level security policies supabase"
Topic: "auth.uid auth.jwt functions"
Topic: "rls performance optimization"

---

Related Modules:
- auth-integration.md - Authentication patterns
- typescript-patterns.md - Client-side access patterns
