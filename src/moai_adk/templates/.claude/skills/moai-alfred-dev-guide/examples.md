# Development Workflow Examples

## SPEC → TDD → Sync Cycle

**Scenario**: Implement user registration feature

```bash
# Step 1: Write SPEC
/alfred:1-plan "User Registration"
→ Creates .moai/specs/SPEC-AUTH-001/
→ Defines requirements in EARS syntax

# Step 2: TDD Implementation
/alfred:2-run SPEC-AUTH-001
RED:   Write tests for registration, validation, duplicate email
GREEN: Implement register() function
REFACTOR: Improve code quality

# Step 3: Sync Documentation
/alfred:3-sync
→ Auto-update README
→ Update CHANGELOG
→ Create @TAG links (SPEC→TEST→CODE→DOC)

Result: ✅ Complete, traceable feature
```

---

Learn more in `reference.md`.
