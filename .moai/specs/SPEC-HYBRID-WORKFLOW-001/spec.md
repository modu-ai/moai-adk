# SPEC-HYBRID-WORKFLOW-001: Validate Hybrid Personal-Pro Workflow

## Overview
Comprehensive validation that the Hybrid Personal-Pro Git Workflow implementation works correctly for both Personal Mode (main-based GitHub Flow) and Team Mode (develop-based Git-Flow).

## Objectives
1. Verify Personal Mode configuration and behavior
2. Verify Team Mode configuration and behavior
3. Validate automatic mode detection based on contributor count
4. Test git-manager agent integration with hybrid workflow

## Requirements

### Ubiquitous Requirements
- System SHALL read hybrid mode configuration from `.moai/config/config.json`
- System SHALL support both Personal Mode (base_branch: main) and Team Mode (base_branch: develop)
- System SHALL auto-detect mode based on git log contributor count
- System SHALL create feature branches from configured base_branch

### Event-Driven Requirements
- WHEN git-manager is invoked → it SHALL read config.json git_strategy.mode
- WHEN git_strategy.mode = "hybrid" → it SHALL auto-detect current mode
- WHEN contributor_count < threshold → it SHALL use Personal Mode (main)
- WHEN contributor_count ≥ threshold → it SHALL use Team Mode (develop)

### Unwanted Scenarios (Error Handling)
- IF config.json is missing or invalid → THEN display error message with recovery instructions
- IF git log fails → THEN fall back to config setting or ask user for mode
- IF base_branch is missing from config → THEN use sensible defaults (main for personal, develop for team)

### State-Driven Requirements
- WHILE working in Personal Mode → feature branches SHALL originate from main
- WHILE working in Team Mode → feature branches SHALL originate from develop
- WHILE switching from Personal to Team Mode → ALL existing feature branches SHALL continue to work
- WHILE in Team Mode → PR base SHALL be develop (not main)

### Optional Features
- WHERE user wants manual mode override → config.json git_strategy.team.enabled can be set explicitly
- WHERE user wants custom threshold → auto_switch_threshold in config can be modified
- WHERE user prefers develop in Personal Mode → base_branch can be manually configured

## Test Scenarios

### Scenario 1: Personal Mode Detection
**Given**: MoAI-ADK project with 1-2 contributors
**When**: `/alfred:1-plan "test feature"` is executed
**Then**:
- Feature branch should be created from main
- Branch name should be feature/SPEC-TEST-001
- Configuration should show git_strategy.mode = "hybrid"

**Success Criteria**:
```bash
git rev-parse --abbrev-ref HEAD
# Should output: feature/SPEC-TEST-001

git merge-base --is-ancestor main HEAD
# Should return 0 (indicating branch originates from main)
```

### Scenario 2: Team Mode Configuration
**Given**: config.json has git_strategy.team.enabled = true
**When**: `/alfred:1-plan "test feature"` is executed
**Then**:
- Feature branch should be created from develop
- PR should target develop (not main)

**Success Criteria**:
```bash
git branch --show-current | grep feature/SPEC
# Should show feature branch name

# Verify it's based on develop
git merge-base develop HEAD | git diff --quiet
# Should show branch is based on develop
```

### Scenario 3: Automatic Mode Switching
**Given**: Initial state has 1 contributor (Personal Mode)
**When**: Second contributor is added to git history
**Then**:
- `contributor_count` should be 2
- System should remain in Personal Mode
- When third contributor added → System should switch to Team Mode

**Success Criteria**:
```bash
contributor_count=$(git log --format='%aN' | sort | uniq | wc -l)
echo "Contributors: $contributor_count"
# Should correctly count unique contributors
```

### Scenario 4: Git-Manager Integration
**Given**: New hybrid workflow configuration
**When**: `git-manager` agent is invoked via `/alfred:2-run SPEC-HYBRID-WORKFLOW-001`
**Then**:
- Agent should read config.json successfully
- Agent should detect current mode (Personal)
- Agent should execute mode-appropriate Git operations

**Success Criteria**:
- No errors during execution
- Feature branches created with correct base
- Commits include required signature

## Acceptance Criteria

### Configuration ✅
- [ ] `.moai/config/config.json` contains valid git_strategy configuration
- [ ] git_strategy.mode = "hybrid"
- [ ] Personal mode has base_branch = "main"
- [ ] Team mode has base_branch = "develop"
- [ ] auto_switch_threshold = 3

### Functionality ✅
- [ ] Personal Mode: feature/SPEC-* branches originate from main
- [ ] Team Mode: feature/SPEC-* branches originate from develop
- [ ] Auto-detection: Works based on contributor count
- [ ] Git-manager: Respects configuration settings
- [ ] CI/CD: Works for both branches (main and develop)

### Documentation ✅
- [ ] CLAUDE.md documents Hybrid workflow
- [ ] README.md explains Personal and Team modes
- [ ] git-manager.md updated with mode-awareness
- [ ] .github/ documentation complete

## Success Metrics

| Metric | Target |
|--------|--------|
| Configuration read success rate | 100% |
| Feature branch creation in correct location | 100% |
| Mode auto-detection accuracy | 100% |
| CI/CD pass rate (both branches) | 100% |
| Documentation completeness | 100% |

## Rollback Plan

If hybrid workflow implementation fails:
1. Revert to previous git_strategy configuration
2. Manually specify mode in config.json
3. Keep develop and main branches as fallback
4. Use simple GitHub Flow if needed

## Related Documents
- .moai/specs/SPEC-STATUSLINE-UVX-001 (Previous implementation)
- .github/CI_CD_HYBRID_WORKFLOW.md (CI/CD configuration)
- .github/BRANCH_PROTECTION_CONFIG.md (Branch protection)
- CLAUDE.md (Documentation)
- README.md (User guide)
