# Three-Tier SPEC Templates

**Created**: 2025-11-21
**Status**: Production Ready

---

## Overview

Based on Alfred's assessment, one of the following three templates is **automatically selected**:

| Level | Target | Files | Time | Sections |
|------|------|--------|------|------|
| **Level 1** | Simple tasks | 1-2 files | 5-10 min | 5 sections |
| **Level 2** | Standard features | 3-5 files | 10-15 min | 7 sections (EARS) |
| **Level 3** | Complex tasks | 5+ files | 20-30 min | 10+ sections |

---

## Level 1: Minimal (For Simple Tasks)

### Applicable To
- Bug fixes
- Simple feature additions
- Documentation updates
- Dependency upgrades

### Template Structure

```markdown
# SPEC-XXX: [Title]

## 1. Overview
One or two sentences summarizing the requirement

## 2. Current Situation (Problem)
- Problem 1
- Problem 2

## 3. Solution Approach
- Approach 1
- Approach 2

## 4. Acceptance Criteria
- [ ] Test case 1
- [ ] Test case 2

## 5. Completion Criteria
- Done when: [clear criteria]
```

### Real Example: Login Korean Message

```markdown
# SPEC-001: Fix Login Page Korean Message Display Error

## 1. Overview
Fix bug where error messages on login page only display in English.

## 2. Current Situation
- Error message: "Invalid credentials" (English only)
- Korean translation missing from i18n configuration

## 3. Solution Approach
1. Add Korean translations to `/src/i18n/messages.json`
2. Verify login form uses i18n functions

## 4. Acceptance Criteria
- [ ] Korean browser: Korean messages displayed
- [ ] English browser: English messages displayed
- [ ] No breaking changes to existing functionality

## 5. Completion Criteria
- Login failure displays messages according to user's browser language
```

---

## Level 2: Standard (For Regular Features)

### Applicable To
- Standard feature development
- Multiple file modifications required
- Basic component integration

### Template Structure (EARS Format)

```markdown
# SPEC-XXX: [Title]

## 📋 SPEC Metadata
| Item | Value |
|------|-----|
| ID | SPEC-XXX |
| Title | [Title] |
| Complexity | MEDIUM |
| Estimated Time | [X] hours |
| Priority | HIGH/MEDIUM/LOW |

## 1. Overview

### Objectives
- Objective 1 to achieve
- Objective 2 to achieve

### Scope
- Included: [In-scope items]
- Excluded: [Out-of-scope items]

## 2. Evaluation Criteria

### Functional Requirements
1. REQ-1: [Detailed requirement]
2. REQ-2: [Detailed requirement]

### Non-Functional Requirements
1. NFREQ-1: [Performance/Security]
2. NFREQ-2: [Scalability]

## 3. Analysis

### Current State
- [Current system state]
- [Issues]

### Considerations
- [Technical constraints]
- [Dependencies]

## 4. Recommendation

### Proposed Solution
- [Solution overview]
- [Technology stack]

### Alternative Review
- Alternative 1: [Pros/Cons]
- Alternative 2: [Pros/Cons]

## 5. Synthesis

### Implementation Strategy
1. Phase 1: [Stage 1]
2. Phase 2: [Stage 2]

### Risk Management
- Risk 1: [Mitigation plan]
- Risk 2: [Mitigation plan]

## 6. Acceptance Criteria

### Functional Validation
- [ ] [Validation item 1]
- [ ] [Validation item 2]

### Quality Validation
- [ ] Test coverage 85% or higher
- [ ] Code review approved

## 7. Completion Criteria

### Definition of Done
- All acceptance criteria met
- Tests passing
- Code review approved
- Documentation completed
```

### Real Example: Profile Image Upload

```markdown
# SPEC-003: User Profile Image Upload Feature

## 📋 SPEC Metadata
| Item | Value |
|------|-----|
| ID | SPEC-003 |
| Title | User Profile Image Upload |
| Complexity | MEDIUM |
| Estimated Time | 2 hours |
| Priority | HIGH |

## 1. Overview

### Objectives
- Enable users to upload profile pictures
- Image optimization and caching handling
- Securely store in database

### Scope
- Included: Profile page modifications, backend API, database
- Excluded: Image editing features, gallery management

## 2. Evaluation Criteria

### Functional Requirements
1. REQ-1: Users can upload images from profile page
2. REQ-2: Images support JPEG/PNG formats only (max 5MB)
3. REQ-3: Uploaded images automatically optimized (1024x1024)
4. REQ-4: Existing images replaced by new images

### Non-Functional Requirements
1. NFREQ-1: Image upload completes within 3 seconds
2. NFREQ-2: Images cached on CDN
3. NFREQ-3: User privacy protection

## 3. Analysis

### Current State
- Profile stores basic info only (name, email)
- No image upload functionality

### Considerations
- Need to add image_url column to profile table
- Configure file storage (S3 or local)
- Require image optimization library

## 4. Recommendation

### Proposed Solution
- Backend: Node.js + Express + multer
- Image processing: Sharp library
- Storage: AWS S3

### Alternative Review
- Alternative 1: Google Drive API - high complexity
- Alternative 2: Local storage - low scalability

## 5. Synthesis

### Implementation Strategy
1. Phase 1: DB schema modification
2. Phase 2: Backend API implementation
3. Phase 3: Frontend UI addition

### Risk Management
- Risk 1: Large files → implement chunked uploads
- Risk 2: Malicious files → add file validation logic

## 6. Acceptance Criteria

### Functional Validation
- [ ] File selection available
- [ ] Only JPEG/PNG accepted
- [ ] Image optimization verified
- [ ] Concurrent uploads supported

### Quality Validation
- [ ] Test coverage 85% or higher
- [ ] Code review approved

## 7. Completion Criteria
- All acceptance criteria met
- Production deployment ready
```

---

## Level 3: Comprehensive (For Complex Tasks)

### Applicable To
- Architecture changes
- Large-scale data model modifications
- Complex integration of multiple components
- Migration tasks

### Template Structure

All Level 2 sections + the following additions:

```markdown
## 6. Architecture Design (NEW)

### System Diagram
[ASCII diagram]

### Component Interactions
- [Component 1] → [Component 2]

## 7. Migration Strategy (NEW)

### Phased Execution Plan
1. Phase 1: [Details]
2. Phase 2: [Details]
3. Phase 3: [Details]

### Rollback Plan
- [Rollback scenarios]

## 8. Performance Considerations (NEW)

### Performance Goals
- [Goal 1]
- [Goal 2]

## 9. Security Considerations (NEW)

### Security Requirements
- [Requirement 1]

### Threat Model
- [Threat 1] → [Response]

## 10. Acceptance Criteria
[More detailed than Level 2]

## 11. Completion Criteria
[Similar to Level 2]
```

---

## Template Auto-Selection Logic

Alfred automatically selects based on the following rules:

```python
if complexity_assessment['strength'] == 'low':
    template = Level1_Minimal
elif complexity_assessment['strength'] == 'medium':
    template = Level2_Standard
else:  # 'high'
    template = Level3_Comprehensive
```

### Selection Criteria

| Complexity | Characteristics | Template |
|--------|------|--------|
| Low | 1-2 files, under 30 min | Level 1 |
| Medium | 3-5 files, 1-2 hours, basic integration | Level 2 |
| High | 5+ files, 2+ hours, architecture changes | Level 3 |

---

## Template Usage Flow

### Step 1: User Request
```
"Add user profile image upload feature"
```

### Step 2: Alfred Assessment
```
Complexity: MEDIUM
→ Level 2 (Standard) template selected
```

### Step 3: Auto-Generate SPEC
spec-builder fills Level 2 template to generate SPEC

### Step 4: User Review
User reviews generated SPEC and modifies if needed

### Step 5: Begin Implementation
```bash
/moai:2-run SPEC-XXX
```

---

## AI Assistance Ratio by Template

| Stage | Level 1 | Level 2 | Level 3 |
|------|---------|---------|---------|
| Draft generation | 90% | 75% | 60% |
| Content review | 80% | 70% | 50% |
| Final completion | 70% | 60% | 40% |
| **Average** | **80%** | **68%** | **50%** |

---

## Template Writing Time Comparison

| Activity | Level 1 | Level 2 | Level 3 |
|------|---------|---------|---------|
| AI generation (80%) | 2-4 min | 4-8 min | 8-12 min |
| User review (20%) | 1-2 min | 2-4 min | 5-8 min |
| **Total Time** | **5-10 min** | **10-15 min** | **20-30 min** |

---

## Template Selection Examples

### Example 1: Bug Fix → Level 1

```
Task: "Change button color"
Files: 1 (CSS)
Time: 5 min
Complexity: LOW

→ Level 1 selected (5-10 min to write)
```

### Example 2: Feature Addition → Level 2

```
Task: "User profile image upload"
Files: 4 (API, Frontend, DB, Middleware)
Time: 2 hours
Complexity: MEDIUM

→ Level 2 selected (10-15 min to write)
```

### Example 3: Architecture Change → Level 3

```
Task: "Migrate to microservices architecture"
Files: 15+ files
Time: 1+ week
Complexity: HIGH

→ Level 3 selected (20-30 min to write)
```

---

## Key Template Features

### ✅ Level 1 (Minimal)
- Fast and concise
- Essential information only
- Optimal for prototype modifications

### ✅ Level 2 (Standard)
- Follows EARS format
- Optimal for general feature development
- Sufficient structure and flexibility

### ✅ Level 3 (Comprehensive)
- Includes architecture design
- Optimal for complex projects
- Detailed risk management

---

## FAQ

**Q: Can templates be modified?**
A: Yes, generated SPECs can be modified anytime. You can add or remove sections as needed.

**Q: What if complexity changes during work?**
A: Alfred will detect this and suggest switching to a higher level template.

**Q: Does Level 1 also require tests?**
A: Yes, tests are mandatory at all levels. However, Level 1 only requires simple unit tests, while Level 3 includes integration tests.

---

**Document Version**: 1.0.0
**Last Updated**: 2025-11-21
**Status**: Production Ready
