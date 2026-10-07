---
description: "Boundary verification methodology for detecting integration defects"
paths: "**/*_test.go,**/*test*.py,**/*test*.ts,**/*test*.rs,**/test/**,**/tests/**"
---

<!-- Source: revfactory/harness — Apache License 2.0 — see .claude/rules/moai/NOTICE.md -->

# Boundary Verification

Systematic methodology for detecting defects at component boundaries. These defects are missed by individual component testing but manifest when components interact.

## Overview

**Core Problem**: Individual components test correctly, but integration fails at boundaries. TypeScript generics, type casting, and absent cross-validation create boundary defects that unit tests cannot catch.

**Key Principle**: "Boundary bugs are not found by testing each side alone. Both sides must be read together."

---

## Case Studies

### Case 1: API Response Schema Mismatch

**Symptom**: API returns correct data structure, frontend hook receives it, but runtime fails because the shape doesn't match type expectation.

**Example**: 
- Backend API returns: `{ projects: [{ id, name }] }`
- Frontend hook expects generic type parameter: `SlideProject[]`
- TypeScript compilation passes (generic casting hides mismatch)
- Runtime error: `Cannot read property 'name' of undefined`

**Root Cause**: API and hook typed independently without cross-validation. Generic type parameters bypass compile-time checking.

**Boundary Condition to Test**:
- Actual API response must be unwrapped if wrapped in outer object
- Field names (camelCase vs snake_case) must match
- Required vs optional fields must align

**Verification Strategy**:
1. Locate API route and extract `NextResponse.json()` payload structure
2. Locate corresponding frontend hook and extract generic type `T`
3. Trace payload from source through any transformation
4. Verify hook calls `.json<T>()` with correct extraction (e.g., `response.projects` if API wraps)
5. Create test: invoke both simultaneously, capture actual response and expected type

---

### Case 2: File Path to Router Mismatch

**Symptom**: Link leads to non-existent URL, or router redirects fail silently.

**Example**:
- Page exists at: `src/app/dashboard/create/page.tsx` → URL `/dashboard/create`
- Link href specified as: `/create` (missing `/dashboard` prefix)
- Browser navigates to `/create`, which doesn't exist

**Root Cause**: File structure and link paths not cross-validated. Route groups and dynamic segments add complexity.

**Boundary Condition to Test**:
- All `href`, `router.push()`, `redirect()` values must map to actual page files
- Route group parentheses `(group)` must be excluded from URL path
- Dynamic segments `[id]` must be recognized as variable segments

**Verification Strategy**:
1. Extract all page files from `src/app/` and construct their URL paths
2. Extract all navigation calls (`href=`, `router.push(`, `redirect(`)
3. Verify each reference maps to an actual page URL
4. Test: click each link in real browser (Playwright/E2E), verify page loads

---

### Case 3: State Transition Incompleteness

**Symptom**: State transitions defined in spec but not fully implemented; intermediate states become dead ends.

**Example**:
- State transition map allows: `generating_template` → `template_approved`
- Implementation updates status to `generating_template` but never transitions to `template_approved`
- Component stuck in `generating_template` state indefinitely

**Root Cause**: State map and state update code not cross-validated. All transitions must have corresponding implementation.

**Boundary Condition to Test**:
- Every transition in the state map must have code that executes it
- No "dead" transitions (defined but unreachable)
- All intermediate states must have exit paths

**Verification Strategy**:
1. Extract state machine definition (transition map or state type)
2. Grep all code for `.update({ status: ... })` or equivalent state changes
3. Verify every allowed transition has an implementation
4. Verify every implemented transition is in the state map (no rogue transitions)
5. Test: trace state changes for each workflow, verify all transitions execute

---

### Case 4: Field Name Mismatch (camelCase ↔ snake_case)

**Symptom**: Field names differ between database (snake_case), API (camelCase), and frontend types (camelCase). One layer doesn't translate correctly.

**Example**:
- DB column: `thumbnail_url`
- API response: `{ thumbnailUrl: "..." }`
- Frontend type: `interface SlideProject { thumbnail_url: string }`
- Actual data: `obj.thumbnailUrl` exists but type expects `obj.thumbnail_url`

**Root Cause**: Automatic transformation skipped in one layer. No validation of field name consistency across layers.

**Boundary Condition to Test**:
- Field name transforms must be applied consistently
- API response must be serialized to frontend type correctly
- TypeScript `Pick` or destructuring must use correct names

**Verification Strategy**:
1. Document field naming convention per layer (DB, API, frontend)
2. Trace field through transformation: DB → API response → TS type
3. Verify transformation applied at each boundary
4. Create assertion: sample API response must serialize to frontend type without field errors
5. Test: `Object.keys(apiResponse) ===  Object.keys(frontendType)` after transformation

---

### Case 5: Missing API Endpoint Implementation

**Symptom**: Frontend calls API endpoint that is called but never executed (missing implementation or wrong route).

**Example**:
- Frontend hook calls `POST /api/projects/123/approve`
- API route `src/app/api/projects/[id]/approve/route.ts` not found
- Request returns 404, but frontend silently handles error

**Root Cause**: API defined in documentation but not implemented. Frontend assumes availability.

**Boundary Condition to Test**:
- Every frontend fetch must have a corresponding API route
- Route signature must match (HTTP method, path parameters)
- API must handle the request (not just exist as an empty file)

**Verification Strategy**:
1. Extract all frontend `fetch()` calls with URLs
2. Extract all API routes from `src/app/api/**`
3. Verify 1:1 correspondence
4. For each unimplemented endpoint: mark as "intentional stub" or implement
5. Test: E2E flow that exercises each endpoint

---

### Case 6: Async Response Shape Inconsistency

**Symptom**: API returns shape for immediate response (`202 Accepted`) but frontend expects final result shape. Polling or websocket updates with different shape cause runtime error.

**Example**:
- Initial response: `{ status: "processing", taskId: 123 }`
- Final result (via webhook): `{ status: "complete", data: [...], errors: [...] }`
- Frontend type union doesn't account for both shapes
- Code accesses `.data` on initial response, crashes

**Root Cause**: Async workflows have multiple response shapes. Type definitions cover only one shape.

**Boundary Condition to Test**:
- Initial response shape must be typed separately from final shape
- Frontend code must handle both shapes
- State machine must track which shape is current

**Verification Strategy**:
1. Identify async workflows (202 responses, polling, webhooks)
2. Document all possible response shapes at each stage
3. Create union type covering all shapes: `InitialResponse | FinalResponse`
4. Verify code uses discriminator field (`status`) to handle each shape
5. Test: invoke async endpoint, verify response at each stage matches expected shape

---

### Case 7: TypeScript Generic Casting Bypass

**Symptom**: Generic casting (`as T`, `as unknown as T`) bypasses type safety. Mismatch exists but compilation succeeds.

**Example**:
- API type: `interface ApiResponse { projects: Project[] }`
- Frontend code: `const data = await fetchJson<SlideProject[]>(url)`
- No intermediate unwrapping: `response.projects` forgotten
- Type works but runtime data is `{ projects: [...] }` not `SlideProject[]`

**Root Cause**: Generic type parameters are not validated at the boundaries. Cast operations bypass static checks.

**Boundary Condition to Test**:
- Generic types must match actual runtime data shape
- Type casting (especially `as any` or unconstrained `as T`) must be verified
- No use of generic type parameters without explicit data transformation

**Verification Strategy**:
1. Audit all `fetchJson<T>` calls and verify `T` matches actual API response shape
2. Grep for unsafe casts: `as any`, `as unknown`, unconstrained generic `as T`
3. For each cast, document why it's safe and verify with integration test
4. Disable TypeScript implicit `any` in `tsconfig.json`
5. Test: runtime data must deserialize to type without casting

---

## Verification Strategy Summary

1. **Extract Boundaries**: Identify component interaction points (API ↔ Frontend, DB ↔ API, State ↔ Code)
2. **Cross-Read**: Read both sides of each boundary simultaneously
3. **Trace Data**: Follow data transformation from source to consumer
4. **Verify Contracts**: Ensure implicit contracts (naming, shape, transitions) are honored
5. **Test Integration**: Unit tests don't catch boundary bugs; integration/E2E tests required
6. **Document Assumptions**: Explicitly state assumptions about field names, response shapes, state transitions

**Golden Rule**: If a boundary can fail, the test must exercise both sides of the boundary.
