# MoAI-ADK Local Setup - Quick Start Guide

**Setup Time**: 5 minutes
**Status**: ✅ Ready for Development

---

## Prerequisites Verification

### 1. Check Python & UV
```bash
python3 --version       # Python 3.11+
uv --version           # Latest version (0.5.x+)
```

### 2. Check Git
```bash
git config user.name
git config user.email
git status
```

### 3. Verify Project Structure
```bash
ls -la .moai/config/config.json    # Should exist
ls -la .claude/hooks/moai/         # Should contain hooks
ls -la src/moai_adk/               # Should contain source
```

---

## One-Time Setup

### Step 1: Install Dependencies
```bash
cd /Users/goos/MoAI/MoAI-ADK
uv sync
```

**Expected Output**:
```
Resolved 28 packages in 0.12s
Prepared 28 packages in 0.35s
Installed 28 packages in 0.82s
```

### Step 2: Verify Configuration
```bash
# Check if config exists
cat .moai/config/config.json

# Should show:
# {
#   "project": { "name": "MoAI-ADK", ... },
#   "language": { "conversation_language": "ko", ... },
#   ...
# }
```

### Step 3: Test Git Integration
```bash
# Check current branch
git branch --show-current
# Expected: release/0.26.0

# Check uncommitted changes
git status
# Expected: clean or list modified files
```

### Step 4: Test Statusline
```bash
uv run .moai/scripts/statusline.py
```

**Expected Output**:
```
✅ Version 0.26.0 (latest)
🌿 Branch: release/0.26.0
📦 MoAI-ADK by GoosLab
🔄 Changes: N (where N = number of modified files)
```

---

## Daily Workflow

### Start New Session
```bash
# Just open Claude Code - SessionStart hooks run automatically
# You'll see:
# 🚀 MoAI-ADK Session Started
# 📦 Version: 0.26.0 (latest)
# 🌿 Branch: release/0.26.0
# [other info...]
```

### Create New Feature
```bash
# Step 1: Write SPEC (EARS format)
/moai:1-plan "사용자 인증 시스템: JWT 토큰 발급, 비밀번호 해시, 로그인 유효성 검사"

# Step 2: MANDATORY - Clear context (save 45K tokens!)
/clear

# Step 3: Implement with TDD
/moai:2-run SPEC-001

# Step 4: Sync documentation
/moai:3-sync auto SPEC-001
```

### Token Management (Critical!)
```
Phase 1: SPEC Creation (30K tokens)
  /moai:1-plan "description"
  /clear  ← MANDATORY!

Phase 2: Implementation (60K tokens)
  /moai:2-run SPEC-XXX
  /clear  ← Optional if context > 150K

Phase 3: Documentation (40K tokens)
  /moai:3-sync auto SPEC-XXX
  /clear  ← After sync completion
```

---

## Common Commands

| Command | Purpose | Time | Notes |
|---------|---------|------|-------|
| `/moai:0-project` | Initialize project | 2min | One-time setup |
| `/moai:1-plan "..."` | Create SPEC | 5-10min | EARS format |
| `/clear` | Reset context | Instant | 45K token savings! |
| `/moai:2-run SPEC-XXX` | TDD implementation | 20-30min | Red→Green→Refactor |
| `/moai:3-sync auto SPEC-XXX` | Auto-documentation | 10-15min | Generate docs |
| `/context` | Check token usage | Instant | Monitor usage |
| `/cost` | View API spend | Instant | Budget tracking |

---

## File Locations

### Configuration
- `.moai/config/config.json` - Main configuration
- `.moai/config/SETUP-STATUS.md` - Detailed setup guide
- `.moai/config/IMPROVEMENTS-PROPOSED.md` - Enhancement analysis

### Project Documentation
- `.moai/project/product.md` - Product vision
- `.moai/project/structure.md` - Architecture
- `.moai/project/tech.md` - Technology stack

### Specifications
- `.moai/specs/SPEC-XXX/spec.md` - Feature specification
- `.moai/specs/SPEC-XXX/implementation.md` - Implementation guide

### Hooks (SessionStart execution)
- `.claude/hooks/moai/session_start__config_health_check.py`
- `.claude/hooks/moai/session_start__show_project_info.py`

### Scripts
- `.moai/scripts/statusline.py` - Status display
- `.moai/scripts/verify_schema.py` - Config validation

---

## Troubleshooting

### SessionStart hooks not running
```bash
# Check hook execution
ls -la .claude/hooks/moai/session_start__*.py

# Test hook manually
cat .moai/config/config.json | python3 .claude/hooks/moai/session_start__config_health_check.py
```

### Git cache issues
```bash
# Clear cache
rm -f .moai/cache/git-info.json

# Verify cache creation on next session
cat .moai/cache/git-info.json
```

### Configuration outdated warning
```bash
# Run project initialization
/moai:0-project

# Or manually update config
# See SETUP-STATUS.md > Configuration Customization Guide
```

### Token usage exceeding budget
```bash
# Check current usage
/context

# Reduce context size
/compact   # Compress conversation

# Or full reset
/clear     # Full reset (save tokens, lose history)
```

---

## Configuration Customization

### Change Project Name
Edit `.moai/config/config.json`:
```json
{
  "project": {
    "name": "My New Project"
  }
}
```

### Change Conversation Language
Edit `.moai/config/config.json`:
```json
{
  "language": {
    "conversation_language": "en",  // "ko", "en", "ja", etc.
    "conversation_language_name": "English"
  }
}
```

### Adjust Hook Timeout
Edit `.moai/config/config.json`:
```json
{
  "hooks": {
    "timeout_ms": 3000  // Increase from 2000 if needed
  }
}
```

### Switch to Team Mode (Multi-contributor)
Edit `.moai/config/config.json`:
```json
{
  "git_strategy": {
    "team": {
      "enabled": true  // Switch to Git-Flow
    }
  }
}
```

---

## Performance Expectations

### First Session
- SessionStart execution: ~500ms (includes cache population)
- Project info display: 6 lines
- Configuration health check: ~100ms

### Subsequent Sessions (< 1 min from first)
- SessionStart execution: ~26ms (cached)
- Project info display: 6 lines
- Configuration health check: ~10ms

### Token Efficiency
- SPEC creation: ~30K tokens (with /clear)
- Implementation: ~60K tokens (with /clear if > 150K)
- Documentation: ~40K tokens (with /clear after sync)
- **Total**: ~130K tokens vs 300K+ (57% savings!)

---

## Code Quality Standards (TRUST 5)

All features must follow TRUST 5 principles:

| Principle | Standard | Check |
|-----------|----------|-------|
| **T**est-first | Tests → Code → Docs | `pytest --cov=85` |
| **R**eadable | Clear naming, formatted | `ruff format` |
| **U**nified | Design patterns, consistent | `mypy src/` |
| **S**ecured | Security scanning | `bandit -r src/` |
| **T**rackable | SPEC linked | `.moai/specs/SPEC-XXX` |

---

## Useful Resources

### Internal Documentation
- `CLAUDE.md` - Full Claude Code execution guide
- `CLAUDE.local.md` - Local project-specific guidelines
- `.moai/config/SETUP-STATUS.md` - Detailed setup analysis
- `.moai/config/IMPROVEMENTS-PROPOSED.md` - Hook improvements

### Quick Commands Reference
```bash
# Check version
cat .moai/config/config.json | grep -A2 '"moai"'

# List active SPECs
find .moai/specs -type d -name "SPEC-*"

# View last commit
git log -1 --pretty=format:"%h %s"

# Check test coverage
pytest --cov=src --cov-report=term-missing
```

---

## Success Checklist

- [x] `.moai/config/config.json` exists and is complete
- [x] `uv sync` completed successfully
- [x] Git configured with user.name and user.email
- [x] SessionStart hooks ready to execute
- [x] Project info displays correctly
- [x] No configuration warnings shown
- [x] Token budget understood
- [x] Ready to create first SPEC

---

## Next Steps

1. **Start Development**:
   ```bash
   /moai:1-plan "첫 번째 기능 설명"
   /clear
   /moai:2-run SPEC-001
   ```

2. **Check Progress**:
   ```bash
   /moai:9-feedback
   ```

3. **View Documentation**:
   - `.moai/project/product.md`
   - `.moai/project/structure.md`
   - `.moai/project/tech.md`

4. **Monitor Quality**:
   ```bash
   pytest --cov=src
   mypy src/
   ```

---

**Configuration Status**: ✅ Complete
**Ready for**: SPEC-First TDD Development
**Last Updated**: 2025-11-19
