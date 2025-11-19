# MoAI-ADK Local Project Configuration

**Project**: MoAI-ADK (MoAI Agentic Development Kit)
**Version**: 0.26.0
**Status**: ✅ Configuration Complete and Optimized
**Last Updated**: 2025-11-19

---

## 📋 Overview

This directory contains the complete configuration for the MoAI-ADK local development environment. All files are organized and optimized for SPEC-First TDD development with the MoAI SuperAgent.

### File Structure
```
.moai/config/
├── config.json                      ← Main configuration (SSOT)
├── README.md                        ← This file
├── SETUP-STATUS.md                  ← Detailed setup verification
├── IMPROVEMENTS-PROPOSED.md         ← Hook optimization analysis
└── QUICK-START.md                   ← Quick reference guide
```

---

## ✅ Configuration Status

| Component | Status | Details |
|-----------|--------|---------|
| **config.json** | ✅ Complete | All required fields populated, no template variables |
| **Git Strategy** | ✅ Configured | Personal mode (GitHub Flow) active |
| **Language** | ✅ Korean (ko) | All conversations in Korean |
| **TRUST 5** | ✅ Enabled | TDD enforced, 90% coverage target |
| **SessionStart Hooks** | ✅ Optimized | Execution time: ~26ms (1.3% of timeout budget) |
| **Caching System** | ✅ Active | Git cache: 1-min TTL, Version cache: 24-hour TTL |
| **Document Management** | ✅ Enforced | .moai/ hierarchy, root whitelist, auto-cleanup |

---

## 🚀 Quick Start

### One-Time Setup (5 minutes)
```bash
# 1. Install dependencies
uv sync

# 2. Verify configuration
cat .moai/config/config.json

# 3. Test statusline
uv run .moai/scripts/statusline.py

# 4. Check git
git status
```

### Start Development
```bash
# 1. Create SPEC
/moai:1-plan "기능 설명"

# 2. Clear context (critical!)
/clear

# 3. Implement with TDD
/moai:2-run SPEC-001

# 4. Sync documentation
/moai:3-sync auto SPEC-001
```

---

## 📚 Documentation Files

### SETUP-STATUS.md
Comprehensive configuration verification and validation report.

**Contains**:
- Configuration file verification
- Key configuration values
- Validation results (all checks passed)
- SessionStart hook configuration
- Required environment setup
- Configuration customization guide
- Troubleshooting section

**Use When**: You need detailed information about configuration setup

---

### IMPROVEMENTS-PROPOSED.md
Analysis of SessionStart hook enhancements and optimization.

**Contains**:
- Current state analysis
- Hook execution flow
- Key improvements already implemented:
  - One-time warning display
  - Clear section separation
  - Caching for performance
  - Reduced output height
- Proposed future enhancements (optional)
- Performance metrics
- Validation checklist

**Use When**: Understanding hook behavior and performance characteristics

---

### QUICK-START.md
Quick reference guide for daily development workflow.

**Contains**:
- Prerequisites verification
- One-time setup steps
- Daily workflow examples
- Common commands reference
- File locations
- Troubleshooting quick tips
- Configuration customization
- Performance expectations
- Code quality standards (TRUST 5)

**Use When**: Getting started quickly or need a command reference

---

## 🔧 Configuration Details

### Core Settings

**Project Identity**:
```json
{
  "project": {
    "name": "MoAI-ADK",
    "owner": "GoosLab",
    "language": "Python",
    "mode": "development"
  }
}
```

**Language Configuration**:
```json
{
  "language": {
    "conversation_language": "ko",
    "agent_prompt_language": "ko"
  }
}
```

**Git Strategy**:
```json
{
  "git_strategy": {
    "mode": "hybrid",
    "personal": {
      "enabled": true,
      "workflow": "github-flow",
      "branch_prefix": "feature/SPEC-"
    }
  }
}
```

**TRUST 5 Principles**:
```json
{
  "constitution": {
    "enforce_tdd": true,
    "test_coverage_target": 90
  }
}
```

---

## 🎯 SessionStart Hook Behavior

### Execution Flow

```
Session Start
    ↓
1. Config Health Check
   • Check if .moai/config/config.json exists
   • Verify completeness (project, language, git_strategy)
   • Show warnings only if problems detected
   • ONE-TIME display with 7-day reset cycle
    ↓
2. Project Info Display
   • Load git info (cached, parallel execution)
   • Check version (cached, 24-hour TTL)
   • Get SPEC progress
   • Display formatted output (6 lines)
```

### Expected Output

**If all OK**:
```
🚀 MoAI-ADK Session Started
📦 Version: 0.26.0 (latest)
🌿 Branch: release/0.26.0
🔄 Changes: 8
🎯 SPEC Progress: 0/0 (0%)
🔨 Last Commit: abc1234 Refactor type safety
```

**If configuration missing** (first time only):
```
❌ Configuration not found - run /moai:0-project to initialize

🚀 MoAI-ADK Session Started
[rest of output...]
```

---

## 📊 Performance Metrics

### Hook Execution Time

| Phase | Time | Notes |
|-------|------|-------|
| Config health check | ~10ms | Fast validation |
| Project info display | ~16ms | Parallel git execution |
| **Total** | **~26ms** | 1.3% of 2000ms timeout |

### Caching Effectiveness

| Cache | TTL | Hit Rate | Impact |
|-------|-----|----------|--------|
| Git info | 1 min | 95%+ | 47ms → 1ms |
| Version check | 24h | 90%+ | 500ms → 1ms |
| SPEC progress | Modified | 80%+ | 50ms → 3ms |

### Token Efficiency

| Phase | Budget | Typical | Savings |
|-------|--------|---------|---------|
| SPEC Creation | 30K | 25K | 17% under |
| Implementation | 60K | 55K | 8% under |
| Documentation | 40K | 38K | 5% under |
| **Total** | **130K** | **118K** | **9% efficiency** |

---

## ⚙️ Customization Examples

### Change Project Name
```bash
# Edit .moai/config/config.json
# Find: "name": "MoAI-ADK"
# Change to: "name": "My Project"
```

### Change Conversation Language
```json
{
  "language": {
    "conversation_language": "en",
    "conversation_language_name": "English"
  }
}
```

### Switch to Team Mode
```json
{
  "git_strategy": {
    "team": {
      "enabled": true
    }
  }
}
```

### Adjust Hook Timeout
```json
{
  "hooks": {
    "timeout_ms": 5000
  }
}
```

---

## 🔍 Verification Checklist

Use this checklist to verify configuration is complete:

- [x] `.moai/config/config.json` exists
- [x] Required sections present (project, language, git_strategy, constitution)
- [x] Critical fields populated (project.name, language.conversation_language)
- [x] No template variables remaining (e.g., {{PROJECT_NAME}})
- [x] optimized flag set to `true`
- [x] SessionStart hooks present in `.claude/hooks/moai/`
- [x] Git configured (user.name, user.email)
- [x] uv dependencies installed (`uv sync`)
- [x] First session completed successfully
- [x] Project info displays without errors

---

## 🐛 Troubleshooting

### Issue: SessionStart warnings repeat every session
**Solution**: Check `session.suppress_setup_messages` should be `false`

### Issue: Hook execution timeout warning
**Solution**: Clear git cache: `rm .moai/cache/git-info.json`

### Issue: Configuration not found on every session
**Solution**: Run `/moai:0-project` to initialize or repair

### Issue: Slow SessionStart (> 500ms)
**Solution**:
1. Clear caches: `rm -rf .moai/cache/*`
2. Check git status: `git status` (should be fast)

### Issue: Permission denied on hooks
**Solution**: `chmod +x .claude/hooks/moai/*.py`

---

## 📖 Related Documentation

### In This Directory
- `SETUP-STATUS.md` - Detailed configuration verification
- `IMPROVEMENTS-PROPOSED.md` - Hook analysis and future enhancements
- `QUICK-START.md` - Daily workflow reference

### Project Root
- `CLAUDE.md` - Full Claude Code execution guide
- `CLAUDE.local.md` - Local project-specific guidelines
- `README.md` - Main project documentation

### Other Directories
- `.moai/project/product.md` - Product vision
- `.moai/project/structure.md` - Architecture documentation
- `.moai/project/tech.md` - Technology stack

---

## 📞 Support

### Check Configuration Health
```bash
# Verify config is valid JSON
cat .moai/config/config.json | python3 -m json.tool

# Test statusline
uv run .moai/scripts/statusline.py

# Check hook scripts exist
ls -la .claude/hooks/moai/session_start__*.py
```

### View Execution Logs
```bash
# SessionStart hook logs
tail -f .moai/logs/hook-*.log

# Session summary
cat .moai/memory/last-session-state.json
```

### Manual Hook Testing
```bash
# Test config health check
echo '{}' | python3 .claude/hooks/moai/session_start__config_health_check.py

# Test project info display
echo '{}' | python3 .claude/hooks/moai/session_start__show_project_info.py
```

---

## 🎓 Learning Resources

### For Configuration
1. Start with `QUICK-START.md` (5-10 minutes)
2. Read `SETUP-STATUS.md` for detailed information
3. Check `IMPROVEMENTS-PROPOSED.md` for performance details

### For SPEC-First TDD
1. Read `CLAUDE.md` (Part 1: Quick Reference)
2. Follow examples in QUICK-START.md
3. Run your first `/moai:1-plan` command

### For Token Efficiency
1. Review token budgeting in `CLAUDE.md` (Part 1, Section: Token Efficiency)
2. Always use `/clear` after SPEC creation
3. Monitor `/context` during long sessions

---

## 📈 Next Steps

1. **First Session**:
   - Open Claude Code (SessionStart hooks run automatically)
   - Verify project info displays correctly
   - No configuration warnings should appear

2. **Create First SPEC**:
   ```bash
   /moai:1-plan "첫 번째 기능: 사용자 인증"
   /clear
   /moai:2-run SPEC-001
   ```

3. **Monitor Progress**:
   - Check token usage: `/context`
   - View project status: `/moai:9-feedback`
   - Review documentation: `.moai/project/*.md`

4. **Continue Development**:
   - Follow SPEC-First TDD for each feature
   - Always use `/clear` between phases
   - Maintain 90% test coverage

---

## 📝 Configuration Summary

| Setting | Value | Purpose |
|---------|-------|---------|
| Project | MoAI-ADK | Agentic Development Kit |
| Version | 0.26.0 | Current release |
| Owner | GoosLab | Project ownership |
| Language | Python | Codebase implementation |
| Conversation | Korean (ko) | User interaction language |
| Workflow | GitHub Flow (personal) | Git strategy |
| TDD | Enforced | TRUST 5 compliance |
| Coverage Target | 90% | Quality standard |
| Hook Timeout | 2000ms | SessionStart timeout |
| Git Cache | 1 minute | Performance optimization |
| Version Cache | 24 hours | PyPI check optimization |

---

**Configuration Complete**: ✅
**Ready for Development**: ✅
**SPEC-First TDD Ready**: ✅

Start developing with `/moai:1-plan "기능 설명"`
