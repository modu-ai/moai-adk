# TUI Survey Examples - Real-World Scenarios

This document provides complete, production-ready examples of TUI surveys across common MoAI-ADK workflows.

_Last updated: 2025-10-22_

---

## Table of Contents

1. [Competition Completion Page](#example-1-competition-completion-page)
2. [State Management Library Selection](#example-2-state-management-library-selection)
3. [Auth System Refactor](#example-3-auth-system-refactor)
4. [Feature Approval Workflow](#example-4-feature-approval-workflow)
5. [API Design Decision](#example-5-api-design-decision)
6. [Database Migration Strategy](#example-6-database-migration-strategy)
7. [Testing Framework Selection](#example-7-testing-framework-selection)
8. [Performance Optimization](#example-8-performance-optimization)

---

## Example 1: Competition Completion Page

### Context

**User Request**: "Competition is over. Add a completion page."

**Current State**:
- Existing `/end` page (authentication required, shows results)
- No public completion page exists
- Need to communicate competition closure to all visitors

### Phase 1: Codebase Analysis

```typescript
// Alfred analyzes existing code
const endPageAnalysis = await Read("/app/end/page.tsx");

// Analysis results:
// - Next.js App Router
// - Requires authentication
// - Shows competition results
// - No environment-based gating
```

### Phase 2: Interactive Survey

**Question 1: Implementation Approach**

```typescript
AskUserQuestion({
  questions: [
    {
      question: "How should the completion page be implemented?",
      header: "Approach",
      multiSelect: false,
      options: [
        {
          label: "Create new public page (/competition-closed)",
          description: "Unguarded route visible to all visitors, no authentication required. Clean separation from existing /end page."
        },
        {
          label: "Modify existing /end page with conditional logic",
          description: "Add competition status check, redirect to completion message if inactive. Reuses existing route structure."
        },
        {
          label: "Use environment-based gating",
          description: "Set NEXT_PUBLIC_COMPETITION_CLOSED=true, redirect all traffic. Runtime configuration."
        }
      ]
    }
  ]
})
```

**User Selection**: "Create new public page (/competition-closed)"

**Question 2: User Behavior**

```typescript
AskUserQuestion({
  questions: [
    {
      question: "For logged-in participants accessing the new page, what should they see?",
      header: "UX",
      multiSelect: false,
      options: [
        {
          label: "Show submission history",
          description: "Redirect authenticated users to /end page, display full results and timeline."
        },
        {
          label: "Show simple completion message only",
          description: "Display 'Competition concluded' notice without historical data. Consistent experience for all users."
        },
        {
          label: "Custom behavior based on user role",
          description: "Different experience per role (admin/participants/visitors)."
        }
      ]
    }
  ]
})
```

**User Selection**: "Show simple completion message only"

### Phase 3: Review & Confirmation

```
Summary:
✓ Implementation: New public page (/competition-closed)
✓ User experience: Simple completion message for all users

Ready to proceed?
```

**User Selection**: "Approve and execute"

### Phase 4: Execution

```typescript
// /app/competition-closed/page.tsx
export const metadata = {
  title: "Competition Closed",
  description: "This competition has concluded"
};

export default function CompetitionClosedPage() {
  return (
    <div className="flex min-h-screen items-center justify-center">
      <div className="max-w-md text-center">
        <h1 className="text-3xl font-bold">Competition Concluded</h1>
        <p className="mt-4 text-muted-foreground">
          Thank you for your interest. This competition has ended.
        </p>
        <p className="mt-2 text-sm">
          Check back soon for future opportunities!
        </p>
      </div>
    </div>
  );
}
```

### Outcome

✅ Clean implementation matching user specifications
✅ No guessing or ambiguity
✅ Direct execution based on confirmed choices
✅ Simple public route with consistent UX

---

## Example 2: State Management Library Selection

### Context

**User Request**: "Add state management to the application."

**Current State**:
- React 18.x application
- No existing state management library
- Component prop drilling in 12+ files
- Need centralized state for user auth, theme, settings

### Phase 1: Codebase Analysis

```typescript
const analysis = {
  framework: "React 18.3.1",
  propDrillingDepth: 4,
  stateComplexity: "Medium",
  teamPreference: null
};
```

### Phase 2: Interactive Survey

**Question 1: Library Selection**

```typescript
AskUserQuestion({
  questions: [
    {
      question: "Which state management library should be used?",
      header: "Library",
      multiSelect: false,
      options: [
        {
          label: "Redux Toolkit (v2.x)",
          description: "Industry standard with extensive ecosystem. DevTools support, middleware. PRO: Battle-tested. CON: More boilerplate."
        },
        {
          label: "Zustand (v5.x)",
          description: "Modern, minimalist state management. Simple API with hooks. PRO: Minimal boilerplate. CON: Fewer resources."
        },
        {
          label: "Jotai (v2.x)",
          description: "Atomic state management. Fine-grained reactivity. PRO: Granular updates. CON: Paradigm shift."
        },
        {
          label: "Context API + useReducer (built-in)",
          description: "React's built-in solution. PRO: Zero dependencies. CON: Performance concerns."
        }
      ]
    }
  ]
})
```

**User Selection**: "Zustand (v5.x)"

**Question 2: Store Structure**

```typescript
AskUserQuestion({
  questions: [
    {
      question: "How should the Zustand store be structured?",
      header: "Structure",
      multiSelect: false,
      options: [
        {
          label: "Single global store",
          description: "All state in one store. Simple to reason about."
        },
        {
          label: "Multiple domain stores",
          description: "Separate stores per domain (useAuthStore, useThemeStore, useSettingsStore). Clear separation."
        },
        {
          label: "Hybrid (global + domain-specific)",
          description: "Core state in global store, specialized logic in domain stores."
        }
      ]
    }
  ]
})
```

**User Selection**: "Multiple domain stores"

**Question 3: Persistence**

```typescript
AskUserQuestion({
  questions: [
    {
      question: "Which state should be persisted to localStorage?",
      header: "Persistence",
      multiSelect: true,
      options: [
        {
          label: "Authentication state",
          description: "Persist user session, tokens. Resume logged-in state."
        },
        {
          label: "Theme preferences",
          description: "Persist dark/light mode selection."
        },
        {
          label: "User settings",
          description: "Persist app configuration (language, notifications)."
        },
        {
          label: "None (session-only)",
          description: "Don't persist any state."
        }
      ]
    }
  ]
})
```

**User Selections**: ["Theme preferences", "User settings"]

### Phase 3: Execution

```typescript
// stores/theme.ts
import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface ThemeState {
  theme: 'light' | 'dark';
  setTheme: (theme: 'light' | 'dark') => void;
  toggleTheme: () => void;
}

export const useThemeStore = create<ThemeState>()(
  persist(
    (set) => ({
      theme: 'light',
      setTheme: (theme) => set({ theme }),
      toggleTheme: () => set((state) => ({
        theme: state.theme === 'light' ? 'dark' : 'light'
      }))
    }),
    {
      name: 'theme-storage'
    }
  )
);

// stores/settings.ts
interface SettingsState {
  language: string;
  notifications: boolean;
  setLanguage: (language: string) => void;
  toggleNotifications: () => void;
}

export const useSettingsStore = create<SettingsState>()(
  persist(
    (set) => ({
      language: 'en',
      notifications: true,
      setLanguage: (language) => set({ language }),
      toggleNotifications: () => set((state) => ({
        notifications: !state.notifications
      }))
    }),
    {
      name: 'settings-storage'
    }
  )
);

// stores/auth.ts (no persistence)
interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  login: (user: User) => void;
  logout: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  isAuthenticated: false,
  login: (user) => set({ user, isAuthenticated: true }),
  logout: () => set({ user: null, isAuthenticated: false })
}));
```

### Outcome

✅ Zustand v5.x with minimal boilerplate
✅ Three domain-specific stores
✅ Selective persistence (theme + settings, not auth)
✅ Type-safe TypeScript implementation
✅ Exact match to user specifications

---

## Example 3: Auth System Refactor

### Context

**User Request**: "Refactor the auth system to use JWT instead of sessions."

**Current State**:
- Session-based authentication (Express + Redis)
- 20+ routes depend on session middleware
- ~500 active users in production
- No JWT infrastructure exists

### Risk Assessment

```typescript
const riskAnalysis = {
  activeUsers: 500,
  dependentRoutes: 23,
  currentUptime: "99.2%",
  rollbackComplexity: "High",
  estimatedDowntime: "2-4 hours (big bang) or 0 minutes (incremental)"
};
```

### Interactive Survey

**Question 1: Migration Strategy**

```typescript
AskUserQuestion({
  questions: [
    {
      question: "What migration strategy should be used for the auth refactor?",
      header: "Strategy",
      multiSelect: false,
      options: [
        {
          label: "Big bang migration (full rewrite)",
          description: "Complete replacement in one deployment. Timeline: 2-3 weeks. Risk: High (downtime, all users logged out)."
        },
        {
          label: "Incremental migration with feature flags",
          description: "Gradual rollout per route. Timeline: 4-6 weeks. Risk: Medium (code duplication)."
        },
        {
          label: "Parallel run with adapter pattern",
          description: "Support both methods simultaneously. Timeline: 6-8 weeks. Risk: Low (safe rollback, zero downtime)."
        }
      ]
    }
  ]
})
```

**User Selection**: "Parallel run with adapter pattern"

**Question 2: JWT Configuration**

```typescript
AskUserQuestion({
  questions: [
    {
      question: "How should JWT tokens be configured?",
      header: "JWT Config",
      multiSelect: false,
      options: [
        {
          label: "Short-lived access tokens only (15 min)",
          description: "Single token type. Simple but requires frequent logins."
        },
        {
          label: "Access + refresh tokens (15 min + 7 days)",
          description: "Access token for requests, refresh for renewal. Standard approach, better UX."
        },
        {
          label: "Long-lived access tokens (24 hours)",
          description: "Extended lifespan, no refresh tokens. Simplified but less secure."
        }
      ]
    }
  ]
})
```

**User Selection**: "Access + refresh tokens (15 min + 7 days)"

**Question 3: Rollout Plan**

```typescript
AskUserQuestion({
  questions: [
    {
      question: "What rollout percentage should be used for JWT migration?",
      header: "Rollout",
      multiSelect: false,
      options: [
        {
          label: "Canary (5% → 25% → 50% → 100%)",
          description: "Gradual rollout over 4 weeks with monitoring. Safest approach."
        },
        {
          label: "Blue-Green (0% → 100%)",
          description: "Full cutover after parallel testing. Faster but higher risk."
        },
        {
          label: "Opt-in beta (users choose)",
          description: "Users voluntarily test JWT auth. Slowest but zero forced risk."
        }
      ]
    }
  ]
})
```

**User Selection**: "Canary (5% → 25% → 50% → 100%)"

### Execution

```typescript
// lib/auth/adapter.ts
export class AuthAdapter {
  private sessionAuth: SessionAuthHandler;
  private jwtAuth: JWTAuthHandler;
  private rolloutPercentage: number;

  async authenticate(req: Request): Promise<User | null> {
    const useJWT = this.shouldUseJWT(req);

    if (useJWT) {
      return this.jwtAuth.authenticate(req);
    } else {
      return this.sessionAuth.authenticate(req);
    }
  }

  private shouldUseJWT(req: Request): boolean {
    // Sticky session: once JWT, always JWT
    if (req.cookies.auth_method === 'jwt') return true;

    // Gradual rollout based on user ID hash
    const userId = req.session?.userId;
    if (!userId) return false;

    const hash = hashUserId(userId);
    return (hash % 100) < this.rolloutPercentage;
  }
}

// Rollout schedule (via environment variable)
// Week 1-2: ROLLOUT_PERCENTAGE=5
// Week 3-4: ROLLOUT_PERCENTAGE=25
// Week 5-6: ROLLOUT_PERCENTAGE=50
// Week 7-8: ROLLOUT_PERCENTAGE=100
```

### Outcome

✅ Parallel auth system with adapter pattern
✅ Zero downtime migration over 6-8 weeks
✅ Safe rollback at any stage
✅ Canary deployment with monitoring
✅ Access (15 min) + refresh (7 days) tokens
✅ Matches confirmed risk tolerance

---

## Example 4: Feature Approval Workflow

### Context

**User Request**: "Clean up unused files from the last refactor."

**Current State**:
- 50 files not referenced in codebase (last modified >6 months ago)
- Files include: deprecated components, old test fixtures, legacy utils
- Risk: some files may be dynamically imported or used in scripts

### Phase 1: Analysis & Backup

```typescript
const unusedFiles = [
  "/src/components/deprecated/OldButton.tsx",
  "/src/utils/legacy/formatter.ts",
  // ... 48 more files
];

// Create backup
await Bash("git checkout -b backup/pre-cleanup-2025-10-22");
await Bash("tar -czf .moai-backups/pre-cleanup-$(date +%Y%m%d).tar.gz src/");
```

### Phase 2: Approval Survey

```typescript
AskUserQuestion({
  questions: [
    {
      question: "50 unused files identified. Review and approve deletion?",
      header: "Cleanup",
      multiSelect: false,
      options: [
        {
          label: "Show file list first",
          description: "Display all 50 files before deciding. Manual review of each path."
        },
        {
          label: "Approve and execute cleanup",
          description: "Delete all 50 files immediately. Backup created. Reversible via Git."
        },
        {
          label: "Approve deletion for specific directories only",
          description: "Select directories to clean. Partial cleanup with directory-level control."
        },
        {
          label: "Cancel operation",
          description: "Abort cleanup, keep all files."
        }
      ]
    }
  ]
})
```

**User Selection**: "Show file list first"

### Phase 3: Detailed Review

```typescript
const filesByCategory = {
  "Deprecated Components (15 files)": [
    "/src/components/deprecated/OldButton.tsx",
    "/src/components/deprecated/LegacyModal.tsx",
    // ... 13 more
  ],
  "Legacy Utilities (20 files)": [
    "/src/utils/legacy/formatter.ts",
    "/src/utils/legacy/validator.ts",
    // ... 18 more
  ],
  "Old Test Fixtures (15 files)": [
    "/tests/fixtures/old-data.json",
    "/tests/fixtures/deprecated-mock.ts",
    // ... 13 more
  ]
};

// Follow-up survey
AskUserQuestion({
  questions: [
    {
      question: "Which categories should be cleaned up?",
      header: "Categories",
      multiSelect: true,
      options: [
        {
          label: "Deprecated Components (15 files)",
          description: "Remove old UI components no longer used."
        },
        {
          label: "Legacy Utilities (20 files)",
          description: "Remove utility functions replaced by modern alternatives."
        },
        {
          label: "Old Test Fixtures (15 files)",
          description: "Remove outdated test data and mocks."
        }
      ]
    }
  ]
})
```

**User Selections**: ["Deprecated Components (15 files)", "Old Test Fixtures (15 files)"]

### Phase 4: Execution

```typescript
const filesToDelete = [
  ...filesByCategory["Deprecated Components (15 files)"],
  ...filesByCategory["Old Test Fixtures (15 files)"]
];

for (const file of filesToDelete) {
  await Bash(`rm ${file}`);
}

await Bash(`git add -A && git commit -m "chore: remove deprecated components and test fixtures

- Deleted 15 deprecated component files
- Deleted 15 old test fixture files
- Preserved 20 legacy utility files for gradual migration
- Backup available at .moai-backups/pre-cleanup-2025-10-22.tar.gz

Refs: cleanup-2025-10-22"`);
```

### Outcome

✅ 30 files deleted (selective cleanup)
✅ 20 files preserved (user decision)
✅ Complete backup and Git branch for rollback
✅ User had full control and visibility
✅ Explicit approval before destructive operation

---

## Example 5: API Design Decision

### Context

**User Request**: "Add pagination to the /api/posts endpoint."

**Current State**:
- Endpoint returns all posts (currently 200, growing to 1000+)
- No pagination implemented
- Frontend expects array of posts

### Interactive Survey

```typescript
AskUserQuestion({
  questions: [
    {
      question: "Which pagination strategy should be used?",
      header: "Pagination",
      multiSelect: false,
      options: [
        {
          label: "Offset-based (?page=1&limit=20)",
          description: "Traditional page numbers. Simple, familiar. CON: Performance degrades with large offsets."
        },
        {
          label: "Cursor-based (?cursor=abc123&limit=20)",
          description: "Use cursor (e.g., last post ID). Better performance, consistent results. CON: No random page access."
        },
        {
          label: "Hybrid (offset + cursor fallback)",
          description: "Offset for first 100 pages, cursor thereafter. Balances UX and performance. CON: Complex."
        }
      ]
    },
    {
      question: "Should the API return pagination metadata?",
      header: "Metadata",
      multiSelect: true,
      options: [
        {
          label: "Total count",
          description: "Include total number of posts. Allows 'Page X of Y'. Warning: COUNT(*) slow on large tables."
        },
        {
          label: "Has next/previous flags",
          description: "Boolean flags indicating more data. Efficient, no COUNT() needed."
        },
        {
          label: "Next/previous page URLs",
          description: "Full URLs (HATEOAS). Self-documenting but larger response."
        }
      ]
    }
  ]
})
```

**User Selections**:
- Pagination: "Cursor-based (?cursor=abc123&limit=20)"
- Metadata: ["Has next/previous flags", "Next/previous page URLs"]

### Execution

```typescript
// GET /api/posts?cursor=abc123&limit=20
{
  "data": [...],
  "pagination": {
    "hasNext": true,
    "hasPrevious": true,
    "nextUrl": "/api/posts?cursor=def456&limit=20",
    "previousUrl": "/api/posts?cursor=xyz789&limit=20"
  }
}
```

---

## Example 6: Database Migration Strategy

### Context

**User Request**: "Add full-text search to the posts table."

**Current State**:
- PostgreSQL 15.x database
- 500k rows in posts table
- Production database (24/7 availability required)

### Interactive Survey

```typescript
AskUserQuestion({
  questions: [
    {
      question: "Which full-text search implementation should be used?",
      header: "Search",
      multiSelect: false,
      options: [
        {
          label: "PostgreSQL GIN index with tsvector",
          description: "Native Postgres full-text search. PRO: No external dependencies. CON: Less flexible than dedicated engines."
        },
        {
          label: "PostgreSQL pg_trgm (trigram) index",
          description: "Fuzzy matching with trigrams. PRO: Typo-tolerant. CON: Higher storage overhead."
        },
        {
          label: "External search engine (Elasticsearch/Meilisearch)",
          description: "Dedicated search service. PRO: Advanced features. CON: Operational overhead, data sync."
        }
      ]
    },
    {
      question: "How should the index be created on production?",
      header: "Migration",
      multiSelect: false,
      options: [
        {
          label: "Online with CONCURRENTLY (zero downtime)",
          description: "CREATE INDEX CONCURRENTLY allows reads/writes during creation. Duration: ~2-4 hours. Risk: Low."
        },
        {
          label: "Offline during maintenance window",
          description: "Scheduled downtime. Duration: ~30-60 minutes. Risk: Medium (unavailable)."
        },
        {
          label: "Blue-green deployment with replica",
          description: "Create index on replica, promote to primary. Duration: ~1 hour prep + instant cutover. Risk: Low but complex."
        }
      ]
    }
  ]
})
```

**User Selections**:
- Implementation: "PostgreSQL GIN index with tsvector"
- Migration: "Online with CONCURRENTLY (zero downtime)"

### Execution

```sql
-- Migration: add_fulltext_search_to_posts.sql

-- Step 1: Add tsvector column
ALTER TABLE posts
ADD COLUMN search_vector tsvector;

-- Step 2: Populate search_vector (batched)
UPDATE posts
SET search_vector = to_tsvector('english', title || ' ' || content)
WHERE id BETWEEN 1 AND 100000;

-- Step 3: Create GIN index (CONCURRENTLY = zero downtime)
CREATE INDEX CONCURRENTLY idx_posts_search_vector
ON posts USING GIN(search_vector);

-- Step 4: Add trigger to keep search_vector updated
CREATE TRIGGER posts_search_vector_update
BEFORE INSERT OR UPDATE ON posts
FOR EACH ROW EXECUTE FUNCTION
  tsvector_update_trigger(search_vector, 'pg_catalog.english', title, content);
```

---

## Example 7: Testing Framework Selection

### Context

**User Request**: "Set up testing for the React application."

**Current State**:
- React 18.x + TypeScript
- No tests exist yet
- Need unit tests, component tests, and E2E tests

### Interactive Survey

```typescript
AskUserQuestion({
  questions: [
    {
      question: "Which unit/component testing framework should be used?",
      header: "Test FW",
      multiSelect: false,
      options: [
        {
          label: "Vitest + React Testing Library",
          description: "Modern, fast test runner with native ESM. PRO: Fastest execution. CON: Newer ecosystem."
        },
        {
          label: "Jest + React Testing Library",
          description: "Industry standard. PRO: Mature, many resources. CON: Slower, ESM issues."
        },
        {
          label: "Vitest + Playwright component testing",
          description: "Hybrid with real browser testing. PRO: True browser environment. CON: Slower, complex setup."
        }
      ]
    },
    {
      question: "Which E2E testing tool should be used?",
      header: "E2E Tool",
      multiSelect: false,
      options: [
        {
          label: "Playwright",
          description: "Modern E2E framework. PRO: Fast, reliable, trace viewer. CON: Node-only API."
        },
        {
          label: "Cypress",
          description: "Popular E2E tool. PRO: Great docs, visual testing. CON: Slower than Playwright."
        },
        {
          label: "None (unit/component tests only)",
          description: "Skip E2E initially. PRO: Faster setup. CON: No full user flow coverage."
        }
      ]
    }
  ]
})
```

**User Selections**:
- Unit/Component: "Vitest + React Testing Library"
- E2E: "Playwright"

---

## Example 8: Performance Optimization

### Context

**User Request**: "The dashboard is loading slowly."

**Current State**:
- Dashboard loads in 4-5 seconds
- Multiple data fetches on mount
- Large bundle size (500KB)
- No profiling data yet

### Interactive Survey

```typescript
AskUserQuestion({
  questions: [
    {
      question: "What is the primary optimization goal?",
      header: "Goal",
      multiSelect: false,
      options: [
        {
          label: "Reduce initial load time",
          description: "Focus on bundle size, code splitting, lazy loading. Target: <2s to interactive."
        },
        {
          label: "Reduce time to first data",
          description: "Optimize API requests, implement caching, parallel fetches. Target: <1s to first content."
        },
        {
          label: "Improve perceived performance",
          description: "Add loading states, skeleton screens, progressive rendering. Target: instant visual feedback."
        }
      ]
    },
    {
      question: "Which optimization techniques should be applied?",
      header: "Techniques",
      multiSelect: true,
      options: [
        {
          label: "Code splitting + lazy loading",
          description: "Split dashboard into chunks, load on demand. Impact: -200KB initial bundle."
        },
        {
          label: "Data fetching optimization",
          description: "Parallel requests, caching, React Query. Impact: -2s load time."
        },
        {
          label: "Image optimization",
          description: "WebP format, responsive images, lazy loading. Impact: -100KB assets."
        },
        {
          label: "Memo/useMemo for expensive renders",
          description: "Prevent unnecessary re-renders. Impact: smoother interactions."
        }
      ]
    }
  ]
})
```

**User Selections**:
- Goal: "Reduce initial load time"
- Techniques: ["Code splitting + lazy loading", "Data fetching optimization"]

---

## Best Practices Summary

### Effective TUI Surveys:

1. **Analyze first**: Examine codebase context before presenting options
2. **Limit options**: 2-4 choices per question (avoid decision fatigue)
3. **Show trade-offs**: PRO/CON, effort, risk for each option
4. **Allow "Other"**: Provide escape hatch for custom input
5. **Review before execution**: Summary step with "Go back" option
6. **Context is key**: Include current state, constraints, implications

### Common Patterns:

- **Architectural decisions**: Present 3 approaches with trade-offs
- **Technical choices**: Show maturity, ecosystem, pros/cons
- **Migration strategies**: Include timeline, risk, rollback plan
- **Approval workflows**: Show impact, backup plan, reversibility

### Integration with MoAI Workflow:

- **Plan phase**: Scope clarification, EARS pattern selection
- **Run phase**: Library selection, implementation approach
- **Sync phase**: Documentation format, merge strategy

---

**Examples Count**: 8 complete scenarios
**Total Lines**: 850+
**Coverage**: Architecture, libraries, refactoring, approvals, API design, databases, testing, performance
**Last Updated**: 2025-10-22
