# Expert Agents Testing Guide

## Overview

This guide explains how to verify that the expert agent proactive delegation system works correctly. Expert agents should automatically activate when implementation-planner detects domain-specific keywords in SPEC documents.

---

## Test Case: Full-Stack Dashboard SPEC

### Location
- SPEC: `.moai/specs/SPEC-EXPERT-TEST-001/spec.md`
- Type: Full-stack application with multiple domains

### Expected Expert Agent Activations

**SPEC Keywords Present**:
```
backend, api, authentication, database (→ backend-expert)
frontend, React, components, state management (→ frontend-expert)
design, accessibility, a11y, WCAG, Figma (→ ui-ux-expert)
deployment, Docker, Kubernetes, CI/CD (→ devops-expert)
```

**Expected Result**:
When running `/alfred:2-run SPEC-EXPERT-TEST-001`, the system should:
1. Detect all 4 domain keywords
2. Automatically invoke all 4 expert agents in sequence:
   - backend-expert (API & database design)
   - frontend-expert (component architecture)
   - ui-ux-expert (accessibility & design system)
   - devops-expert (CI/CD & infrastructure)
3. Generate implementation plan with `@EXPERT:BACKEND | @EXPERT:FRONTEND | @EXPERT:UIUX | @EXPERT:DEVOPS` tags

---

## How to Execute Test

### Step 1: Verify Test SPEC Exists

```bash
ls -la .moai/specs/SPEC-EXPERT-TEST-001/spec.md
```

### Step 2: Run Implementation Planning

```bash
/alfred:2-run SPEC-EXPERT-TEST-001
```

### Step 3: Verify Expert Agent Invocations

Look for these indicators in the implementation plan:

**✅ Backend-Expert Activated**:
- Mentions of JWT token strategy
- Database schema recommendations
- API endpoint design patterns
- Authentication architecture

**✅ Frontend-Expert Activated**:
- React component architecture suggestions
- State management patterns (Redux recommendations)
- Component library recommendations
- Performance optimization tips

**✅ UI/UX-Expert Activated**:
- WCAG 2.2 AA compliance guidance
- Design token integration strategy
- Figma MCP workflow recommendations
- Accessibility checklist

**✅ DevOps-Expert Activated**:
- Docker containerization strategy
- Kubernetes deployment patterns
- CI/CD pipeline recommendations
- Infrastructure-as-Code guidance

---

## Verification Checklist

### Implementation Plan Quality

- [ ] All 4 expert domains have recommendations
- [ ] Recommendations are specific and actionable
- [ ] No generic/boilerplate content
- [ ] Technology stack is justified

### TAG System Integration

- [ ] Plan includes `@EXPERT:BACKEND` tag
- [ ] Plan includes `@EXPERT:FRONTEND` tag
- [ ] Plan includes `@EXPERT:UIUX` tag
- [ ] Plan includes `@EXPERT:DEVOPS` tag

### Expert-Specific Content

**Backend**:
- [ ] JWT token implementation strategy
- [ ] Database schema design
- [ ] API versioning approach
- [ ] Authentication flow diagram

**Frontend**:
- [ ] Component hierarchy structure
- [ ] State management approach (Redux/Zustand/etc)
- [ ] Library recommendations (axios, react-query, etc)
- [ ] Performance optimization strategy

**UI/UX**:
- [ ] WCAG 2.2 AA compliance checklist
- [ ] Color contrast requirements
- [ ] Keyboard navigation plan
- [ ] Design system integration (DTCG tokens)

**DevOps**:
- [ ] Docker image strategy
- [ ] Kubernetes manifest approach
- [ ] CI/CD pipeline stages (build, test, deploy)
- [ ] Monitoring and logging strategy

---

## Expected Output Format

The implementation plan should show:

```
IMPLEMENTATION PLAN: SPEC-EXPERT-TEST-001

Created by: implementation-planner
Expert Consultations: backend-expert, frontend-expert, ui-ux-expert, devops-expert

## 1. Overview
[Summary of full-stack requirements]

## 2. Technology Stack

### Backend (backend-expert consultation)
- Language: Python 3.13
- Framework: FastAPI
- Database: PostgreSQL 16
- Authentication: PyJWT
- ORM: SQLAlchemy 2.0

### Frontend (frontend-expert consultation)
- Framework: React 19
- State Management: Redux
- UI Library: React Query
- Testing: Vitest

### UI/UX (ui-ux-expert consultation)
- Design System: W3C DTCG 2025.10
- Accessibility: WCAG 2.2 AA
- Design Tool: Figma with MCP

### DevOps (devops-expert consultation)
- Container: Docker
- Orchestration: Kubernetes
- CI/CD: GitHub Actions
- Monitoring: Prometheus + Grafana

## 3. Implementation Phases

### Phase 1: Backend Infrastructure
[backend-expert recommendations]

### Phase 2: Frontend Architecture
[frontend-expert recommendations]

### Phase 3: UI/UX Implementation
[ui-ux-expert recommendations]

### Phase 4: Deployment Setup
[devops-expert recommendations]

## 4. Expert Consultations Summary

✅ backend-expert: Reviewed API design, database schema, authentication flow
✅ frontend-expert: Reviewed component architecture, state management
✅ ui-ux-expert: Reviewed accessibility compliance, design system integration
✅ devops-expert: Reviewed deployment strategy, CI/CD pipeline

## 5. Tagged Requirements

@SPEC:EXPERT-TEST-001
@EXPERT:BACKEND | @EXPERT:FRONTEND | @EXPERT:UIUX | @EXPERT:DEVOPS
```

---

## Troubleshooting

### Issue: Expert agents not activating

**Possible Cause 1: Keywords not detected**
- Verify SPEC contains trigger keywords
- Check keyword spelling and casing
- Ensure keywords are in main sections (not comments)

**Possible Cause 2: implementation-planner not running keyword scan**
- Verify `.claude/agents/alfred/implementation-planner.md` includes proactive delegation
- Check git log to ensure expert delegation commit exists
- Run `/alfred:2-run` with verbose output

**Fix**: Re-read implementation-planner.md and verify "Proactive Expert Delegation" section exists

### Issue: Expert recommendations are generic

**Possible Cause**: Expert agent receiving incomplete SPEC context
- Verify full SPEC content is passed to expert agent
- Check that SPEC includes specific requirements (not vague)

**Fix**: Add more specific details to SPEC requirements

### Issue: @EXPERT TAGs not appearing

**Possible Cause**: tag-agent not validating @EXPERT tags
- Verify `.claude/agents/alfred/tag-agent.md` includes @EXPERT TAG system
- Check tag-agent configuration for EXPERT domain support

**Fix**: Re-read tag-agent.md and verify "@EXPERT TAG System (NEW)" section exists

---

## Real-World SPEC Testing

### Test SPECs in MoAI-ADK Codebase

These SPECs should trigger expert agents:

| SPEC | Domain Keywords | Expected Experts |
|------|-----------------|-------------------|
| SPEC-CLI-001 | backend, command-line | backend-expert |
| SPEC-CLAUDE-CODE-FEATURES-001 | frontend, UI, component, design | frontend-expert, ui-ux-expert |
| (Future) SPEC-DEPLOYMENT-001 | deployment, Docker, CI/CD, pipeline | devops-expert |

### To Test Real SPECs

1. Select one of the SPECs above
2. Run: `/alfred:2-run SPEC-ID`
3. Verify expert agents activate based on keywords
4. Review implementation plan for expert-specific recommendations

---

## Success Criteria

✅ **System is working correctly when**:

1. All 4 expert agents activate for SPEC-EXPERT-TEST-001
2. Each expert provides domain-specific recommendations
3. Implementation plan includes all `@EXPERT:DOMAIN` tags
4. Tech stack recommendations vary by domain
5. Phase-by-phase implementation respects domain dependencies

❌ **System needs fixes if**:

1. Expert agents don't activate despite keywords present
2. Expert recommendations are generic/boilerplate
3. Missing `@EXPERT:DOMAIN` tags in plan
4. Single expert recommendation for multi-domain SPEC
5. Incorrect domain expert activated for given keywords

---

## Next Steps

After verifying expert agents work correctly:

1. **Tag Validation**: Run tag-agent to verify @EXPERT TAG integrity
2. **Implementation**: Execute `/alfred:2-run SPEC-EXPERT-TEST-001` for full TDD cycle
3. **Documentation**: Sync documentation with `/alfred:3-sync`
4. **Production Testing**: Apply to real project SPECs

---

**Last Updated**: 2025-11-04
**Status**: Ready for testing
**Related Files**:
- `.moai/specs/SPEC-EXPERT-TEST-001/spec.md`
- `.claude/agents/alfred/implementation-planner.md`
- `.claude/agents/alfred/tag-agent.md`
