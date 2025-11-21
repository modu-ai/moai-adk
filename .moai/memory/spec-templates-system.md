# SPEC Template System

**Status**: Static Reference | **Version**: 1.0.0 | **Language**: English

---

## Overview

The SPEC-First workflow uses **3 automatic template levels** selected by Alfred based on complexity assessment. Each level balances thoroughness with efficiency.

| Level | Name | Complexity | Files | Time | Sections | AI Assistance |
|-------|------|-----------|-------|------|----------|-----------------|
| **1** | Minimal | LOW | 1-2 | 5-10 min | 5 | 80% |
| **2** | Standard | MEDIUM | 3-5 | 10-15 min | 7 (EARS) | 70% |
| **3** | Comprehensive | HIGH | 5+ | 20-30 min | 10+ | 60% |

---

## Level 1: Minimal Template

**Purpose**: Simple bug fixes, documentation updates, minor features

**When Used**:
- Single or 2-file modifications
- No architecture changes
- Estimated time: 30 minutes or less
- One-time tasks with minimal future extension needs

### Template Structure

```markdown
# SPEC-XXX: [Title]

## 1. Overview
One-line summary of requirement

## 2. Current Situation
- Problem 1
- Problem 2

## 3. Solution Approach
- Approach 1
- Approach 2

## 4. Verification Criteria
- [ ] Test case 1
- [ ] Test case 2

## 5. Completion Criteria
Clear, objective definition of "done"
```

### Example: i18n Message Fix

```markdown
# SPEC-001: Login Page Error Message Localization

## 1. Overview
Display error messages in user's browser language instead of English-only

## 2. Current Situation
- Login form shows "Invalid credentials" (English only)
- No Korean translation for error messages
- Users with non-English browsers see English text

## 3. Solution Approach
1. Add Korean translations to `src/i18n/messages.json`
2. Verify login form uses i18n() function for error display
3. Test with browser language set to Korean and English

## 4. Verification Criteria
- [ ] Korean browser displays Korean error message
- [ ] English browser displays English error message
- [ ] No existing functionality broken

## 5. Completion Criteria
Error messages display in user's browser language with proper localization
```

### Writing Time

| Activity | Duration |
|----------|----------|
| AI generates template (80%) | 2-4 min |
| User reviews/edits (20%) | 1-2 min |
| **Total** | **5-10 min** |

---

## Level 2: Standard Template (EARS Format)

**Purpose**: General feature development, moderate complexity

**When Used**:
- 3-5 files need modification
- Basic architectural decisions required
- Estimated time: 1-2 hours
- Future maintenance/extension likely

### Template Structure

```markdown
# SPEC-XXX: [Title]

## 📋 SPEC Metadata
| Field | Value |
|-------|-------|
| ID | SPEC-XXX |
| Complexity | MEDIUM |
| Estimated Time | X hours |
| Priority | HIGH/MEDIUM/LOW |

## 1. Overview (What & Why)

### Goals
- Goal 1
- Goal 2

### Scope
- Include: [In scope]
- Exclude: [Out of scope]

## 2. Evaluation (Requirements)

### Functional Requirements
1. REQ-1: [Detailed requirement]
2. REQ-2: [Detailed requirement]

### Non-Functional Requirements
1. NFREQ-1: [Performance/Security/Scalability]
2. NFREQ-2: [Performance/Security/Scalability]

## 3. Analysis (Current State)

### Current Situation
- [System state]
- [Known limitations]

### Technical Considerations
- [Dependencies]
- [Constraints]

## 4. Recommendation (Solution)

### Proposed Solution
- [Overview]
- [Technology stack]

### Alternative Solutions
- Alt 1: [Pros/Cons]
- Alt 2: [Pros/Cons]

## 5. Synthesis (Implementation)

### Implementation Strategy
1. Phase 1: [First steps]
2. Phase 2: [Next steps]
3. Phase 3: [Final steps]

### Risk Management
- Risk 1: [Mitigation]
- Risk 2: [Mitigation]

## 6. Verification Criteria

### Functional Tests
- [ ] Feature works as specified
- [ ] Edge cases handled
- [ ] User experience validated

### Quality Tests
- [ ] Test coverage ≥ 85%
- [ ] Code review approved
- [ ] No performance regression

## 7. Completion Criteria

Definition of Done:
- All verification criteria met
- Tests passing
- Code review approved
- Documentation complete
```

### Example: Profile Image Upload

```markdown
# SPEC-003: User Profile Image Upload Feature

## 📋 SPEC Metadata
| Field | Value |
|-------|-------|
| ID | SPEC-003 |
| Complexity | MEDIUM |
| Estimated Time | 2 hours |
| Priority | HIGH |

## 1. Overview

### Goals
- Allow users to upload profile images
- Optimize and cache images automatically
- Store securely in database

### Scope
- Include: Upload UI, backend API, database storage, image optimization
- Exclude: Image editing, social sharing, bulk uploads

## 2. Evaluation

### Functional Requirements
1. REQ-1: User can select and upload JPEG/PNG image (max 5MB)
2. REQ-2: Image automatically resizes to 1024x1024
3. REQ-3: Previous image replaced by new upload
4. REQ-4: Image accessible on user profile within 5 seconds

### Non-Functional Requirements
1. NFREQ-1: Upload completes within 3 seconds
2. NFREQ-2: Image cached via CDN for fast retrieval
3. NFREQ-3: User data protected per GDPR requirements

## 3. Analysis

### Current Situation
- Profile stores only text data (name, email)
- No image upload capability
- No file storage infrastructure

### Technical Considerations
- Need to add image_url column to profiles table
- File storage: Local vs. S3 decision required
- Image processing library needed (Sharp recommended)

## 4. Recommendation

### Proposed Solution
- Frontend: React component with drag-drop upload
- Backend: Node.js/Express with multer for file handling
- Image Processing: Sharp library for optimization
- Storage: AWS S3 with CloudFront CDN

### Alternative Solutions
- Alt 1: Google Drive API → Too complex, not recommended
- Alt 2: Local filesystem → Not scalable for production

## 5. Synthesis

### Implementation Strategy
1. **Phase 1**: Database schema modification (add image_url column)
2. **Phase 2**: Backend API endpoint for image upload with validation
3. **Phase 3**: Frontend UI component with upload handling

### Risk Management
- **Risk 1: Large file uploads** → Implement chunked upload with progress
- **Risk 2: Malicious files** → Add file type validation and virus scanning

## 6. Verification Criteria

### Functional Tests
- [ ] User can select image file
- [ ] File format validation (JPEG/PNG only)
- [ ] Image optimization applied (1024x1024)
- [ ] Concurrent uploads supported

### Quality Tests
- [ ] Test coverage ≥ 85%
- [ ] Code review approved
- [ ] E2E test validates full flow

## 7. Completion Criteria

Definition of Done:
- Profile page displays uploaded image
- Image optimized and cached
- All tests passing (unit, integration, E2E)
- Documentation updated
- Ready for production deployment
```

### Writing Time

| Activity | Duration |
|----------|----------|
| AI generates template (70%) | 4-8 min |
| User reviews/edits (30%) | 2-4 min |
| **Total** | **10-15 min** |

---

## Level 3: Comprehensive Template

**Purpose**: Architecture changes, migrations, complex integrations

**When Used**:
- 5+ files or major refactoring
- Architecture or data model changes
- Estimated time: 2+ hours
- Long-term maintenance and extension critical
- Cross-team coordination required

### Template Structure (Level 2 + Additional Sections)

```markdown
# SPEC-XXX: [Title]

## 📋 SPEC Metadata
[Same as Level 2]

## 1-7. [All Level 2 sections]

## 8. Architecture Design (NEW)

### System Architecture Diagram
[ASCII diagram or reference to design document]

### Component Interactions
- Component A → Component B: [Interaction]
- Component B → Component C: [Interaction]

### Data Model Changes
- Table 1: [Schema changes]
- Table 2: [Schema changes]

## 9. Migration Strategy (NEW)

### Detailed Phase Breakdown
1. **Phase 1**: [Specific tasks]
   - Task 1.1
   - Task 1.2
2. **Phase 2**: [Next phase]
3. **Phase 3**: [Final phase]

### Rollback Plan
- Scenario 1: [Rollback steps]
- Scenario 2: [Rollback steps]

### Data Migration
- How to migrate existing data
- Validation steps
- Rollback if needed

## 10. Performance Considerations (NEW)

### Performance Goals
- Metric 1: [Target]
- Metric 2: [Target]

### Optimization Strategies
- Strategy 1: [Approach]
- Strategy 2: [Approach]

## 11. Security Considerations (NEW)

### Security Requirements
- Requirement 1: [Details]
- Requirement 2: [Details]

### Threat Model
- Threat 1: [Mitigation]
- Threat 2: [Mitigation]

## 12. Verification Criteria (More Detailed)
[Similar to Level 2, but more comprehensive]

## 13. Completion Criteria
[Similar to Level 2, emphasizing architecture validation]
```

### Example Sections for Migration

When migrating to microservices, you would include:

- **System Architecture Diagram**: Shows monolith → microservices transition
- **Service Definitions**: User Service, Product Service, Payment Service
- **Communication Protocol**: REST APIs, message queues, event streaming
- **Database Migration**: How to split databases safely
- **Rollback Strategy**: How to revert if issues arise
- **Performance Targets**: Response time goals, throughput requirements
- **Security**: API authentication, data isolation, network security

### Writing Time

| Activity | Duration |
|----------|----------|
| AI generates template (60%) | 8-12 min |
| User reviews/edits (40%) | 5-8 min |
| **Total** | **20-30 min** |

---

## Auto-Selection Logic

Alfred selects template level based on complexity assessment:

```python
if criteria_met <= 1:
    template = Level1_Minimal
    complexity = "LOW"
elif criteria_met <= 3:
    template = Level2_Standard
    complexity = "MEDIUM"
else:
    template = Level3_Comprehensive
    complexity = "HIGH"
```

**Criteria Evaluated**:
1. File modification scope (1 vs. 2-3 vs. 4+)
2. Architecture impact (none vs. partial vs. major)
3. Component integration (1 vs. 2-3 vs. 4+)
4. Implementation time (< 30min vs. 30-120min vs. > 120min)
5. Maintenance needs (none vs. possible vs. certain)

---

## Template Flexibility

Generated SPECs can be modified:
- ✅ Add sections as needed
- ✅ Remove unused sections
- ✅ Reorganize content
- ✅ Adjust detail level

**Principle**: Templates provide structure, but user customization is always welcome.

---

## AI Assistance Ratios

How much AI help varies by template level:

| Phase | Level 1 | Level 2 | Level 3 |
|-------|---------|---------|---------|
| Initial generation | 90% | 75% | 60% |
| Content review | 80% | 70% | 50% |
| Final polish | 70% | 60% | 40% |
| **Average** | **80%** | **68%** | **50%** |

**Note**: Higher-level templates require more user input and thinking.

---

## Best Practices

### For All Levels

1. **Be Specific**: Use concrete examples and numbers
2. **Define Success**: Clear completion criteria matter
3. **Consider Risks**: Address potential problems upfront
4. **Validate Assumptions**: Question requirements if unclear

### For Level 1

- Keep it simple and focused
- One primary goal, minimal scope
- Quick turnaround expected

### For Level 2

- Follow EARS structure for clarity
- Include alternatives evaluated
- Balance detail with brevity

### For Level 3

- Diagram complex architecture clearly
- Plan migration in detail
- Include comprehensive rollback strategy
- Involve stakeholders in validation

---

## Common Pitfalls

### Pitfall 1: Wrong Level Selection
- **Problem**: Using Level 1 for complex work or Level 3 for simple fix
- **Solution**: Alfred's auto-selection, but user can adjust if needed

### Pitfall 2: Incomplete Templates
- **Problem**: Not filling all required sections
- **Solution**: All sections have purpose; at minimum provide rationale if skipping

### Pitfall 3: Ambiguous Criteria
- **Problem**: Completion criteria not specific enough to verify
- **Solution**: Use measurable, objective criteria with checklists

### Pitfall 4: Ignoring Non-Functional Requirements
- **Problem**: Only documenting what the feature does, not how well
- **Solution**: EARS format requires explicit performance/security/scalability considerations

---

## Customization for Teams

### Team with Strict Requirements
→ Use Level 3 for everything over 1 hour

### Startup Team with Fast Iteration
→ Prefer Level 1, escalate to Level 2 when integration challenges arise

### Enterprise with Complex Architecture
→ Always use Level 2 minimum, Level 3 for infrastructure changes

### Distributed Team
→ Invest in Level 2/3 templates for clear communication

---

**Document Version**: 1.0.0
**Last Updated**: 2025-11-21
**Status**: Production Ready
