# MoAI-ADK Configuration Documentation Index

**Navigation Guide for Configuration Setup and Reference**
**Last Updated**: 2025-11-19
**Project Version**: 0.26.0

---

## Quick Navigation

### I'm New to MoAI-ADK - Where Do I Start?

**Start here** → [`QUICK-START.md`](QUICK-START.md) (5-10 minutes)

This file contains:
- Prerequisites verification (Python, git, uv)
- One-time setup steps (5 minutes)
- Daily workflow examples
- Common commands reference
- Quick troubleshooting tips

### I Need to Understand the Configuration

**Read this** → [`SETUP-STATUS.md`](SETUP-STATUS.md) (15-20 minutes)

This file contains:
- Configuration file verification results
- All required and optional settings
- Validation checklist (all items checked)
- Environment setup requirements
- Configuration customization guide
- Troubleshooting with solutions

### I Want to Understand Hook Performance

**Read this** → [`IMPROVEMENTS-PROPOSED.md`](IMPROVEMENTS-PROPOSED.md) (10-15 minutes)

This file contains:
- SessionStart hook execution flow
- Current output and formatting
- Key improvements already implemented
- Caching effectiveness metrics
- Performance targets (all met)
- Proposed future enhancements

### I Need the Big Picture

**Read this** → [`COMPLETION-REPORT.md`](COMPLETION-REPORT.md) (20-30 minutes)

This file contains:
- Executive summary of setup completion
- Configuration deliverables checklist
- Performance metrics validation
- Token efficiency budgeting
- Technology stack validation
- Next steps for development

### I Just Need Quick Facts

**Read this** → [`README.md`](README.md) (5 minutes)

This file contains:
- Configuration overview
- Status summary table
- Quick start command
- File structure reference
- Key settings table
- Performance expectations

### I Need a Specific Answer

**Check here** → [Specific Topics Below](#specific-topics)

---

## Specific Topics

### Setup & Configuration

| Topic | File | Time | Questions Answered |
|-------|------|------|-------------------|
| **First-time setup** | QUICK-START.md | 5-10 min | How do I get started? |
| **Configuration details** | SETUP-STATUS.md | 15-20 min | What fields are configured? |
| **System overview** | README.md | 5 min | What's the configuration status? |
| **Verify everything** | COMPLETION-REPORT.md | 20-30 min | Is everything working? |

### Performance & Optimization

| Topic | File | Time | Questions Answered |
|-------|------|------|-------------------|
| **Hook execution** | IMPROVEMENTS-PROPOSED.md | 10-15 min | How fast are hooks? |
| **Caching system** | IMPROVEMENTS-PROPOSED.md | 10-15 min | What's cached? |
| **Token efficiency** | COMPLETION-REPORT.md | 15 min | How do I save tokens? |
| **Performance metrics** | COMPLETION-REPORT.md | 10 min | What are the metrics? |

### Troubleshooting

| Topic | File | Time | Questions Answered |
|-------|------|------|-------------------|
| **Common issues** | QUICK-START.md | 5-10 min | How do I fix...? |
| **Detailed solutions** | SETUP-STATUS.md | 15-20 min | How do I solve...? |
| **Verification** | COMPLETION-REPORT.md | 20-30 min | Is my setup correct? |

### Customization

| Topic | File | Time | Questions Answered |
|-------|------|------|-------------------|
| **Change settings** | SETUP-STATUS.md | 5-10 min | How do I customize...? |
| **Language & names** | QUICK-START.md | 5 min | How do I change...? |
| **Git workflow** | QUICK-START.md | 5 min | How do I use teams? |

---

## File Structure & Contents

### Core Configuration File

**`config.json`** (14KB)
- Main configuration (single source of truth)
- Fully optimized (no template variables)
- All sections: moai, constitution, git_strategy, language, project, user, session, hooks, document_management, github, pipeline, report_generation, auto_cleanup, daily_analysis
- Read-only after setup (modify via `/moai:0-project` or edit directly)

### Documentation Files

| File | Size | Purpose | Audience | Read Time |
|------|------|---------|----------|-----------|
| **QUICK-START.md** | 8KB | Daily reference | Everyone | 5-10 min |
| **SETUP-STATUS.md** | 12KB | Detailed setup | Developers | 15-20 min |
| **IMPROVEMENTS-PROPOSED.md** | 10KB | Hook analysis | Technical | 10-15 min |
| **COMPLETION-REPORT.md** | 14KB | Final verification | Project leads | 20-30 min |
| **README.md** | 10KB | Overview & index | Everyone | 5 min |
| **INDEX.md** | This file | Navigation guide | Everyone | 3-5 min |

**Total documentation size**: ~68KB (comprehensive coverage)

---

## Reading Recommendations

### By Role

**👨‍💻 Developer (Starting First Feature)**
1. QUICK-START.md (5 min) - Understand workflow
2. README.md (5 min) - Get the overview
3. Start developing with `/moai:1-plan`

**🏗️ Architect (Understanding System)**
1. README.md (5 min) - Overview
2. SETUP-STATUS.md (20 min) - Configuration details
3. IMPROVEMENTS-PROPOSED.md (15 min) - Performance analysis
4. COMPLETION-REPORT.md (30 min) - Full validation

**🔧 DevOps (Setup & Maintenance)**
1. SETUP-STATUS.md (20 min) - Environment setup
2. IMPROVEMENTS-PROPOSED.md (15 min) - Hook system
3. COMPLETION-REPORT.md (30 min) - Verification
4. Troubleshooting in QUICK-START.md

**👔 Project Manager (Status & Metrics)**
1. COMPLETION-REPORT.md (30 min) - Full report
2. README.md (5 min) - Summary table
3. QUICK-START.md (10 min) - Workflow overview

### By Time Available

**⏱️ 5 Minutes**
- README.md - Get the gist

**⏱️ 15 Minutes**
- QUICK-START.md - Understand workflow and commands

**⏱️ 30 Minutes**
- README.md (5 min)
- QUICK-START.md (15 min)
- SETUP-STATUS.md excerpt (10 min)

**⏱️ 1 Hour**
- QUICK-START.md (15 min)
- SETUP-STATUS.md (20 min)
- IMPROVEMENTS-PROPOSED.md (15 min)
- README.md (5 min)

**⏱️ 2+ Hours**
- Read all documentation files in order
- Understand every section
- Become an expert on the system

---

## Configuration Verification Checklist

Use this checklist to verify configuration is complete:

### Essential (Must Have)
- [ ] config.json exists at `.moai/config/config.json`
- [ ] All required sections present
- [ ] No template variables ({{VARIABLE}}) remaining
- [ ] project.name is not empty
- [ ] language.conversation_language is set
- [ ] File is valid JSON

### Important (Should Have)
- [ ] SessionStart hooks present in `.claude/hooks/moai/`
- [ ] Git configured (user.name, user.email)
- [ ] Dependencies installed (uv sync)
- [ ] Python 3.11+ available

### Verification (Check Working)
- [ ] First session runs without errors
- [ ] Project info displays correctly
- [ ] No repeated warnings shown
- [ ] Git cache working (fast sessions)
- [ ] Token budgeting understood

---

## Quick Command Reference

### Project Setup
```bash
uv sync                                    # Install dependencies
cat .moai/config/config.json              # View configuration
uv run .moai/scripts/statusline.py        # Test statusline
```

### Development Workflow
```bash
/moai:1-plan "기능 설명"                 # Create SPEC
/clear                                    # Save 45K tokens!
/moai:2-run SPEC-001                     # Implement with TDD
/moai:3-sync auto SPEC-001               # Sync documentation
```

### Monitoring
```bash
/context                                  # Check token usage
/cost                                     # View API spend
/moai:9-feedback                         # Get feedback
git status                                # Check git status
```

### Verification
```bash
pytest --cov=src                         # Check test coverage
mypy src/                                # Type checking
ruff check src/                          # Linting
```

---

## Documentation Cross-References

### By Topic

**Configuration Setup**
- QUICK-START.md → Step 1-4
- SETUP-STATUS.md → Required Environment Setup
- README.md → Customization Examples

**SessionStart Hooks**
- IMPROVEMENTS-PROPOSED.md → Hook Execution Flow
- SETUP-STATUS.md → SessionStart Hook Configuration
- README.md → Hook Behavior

**Performance Optimization**
- IMPROVEMENTS-PROPOSED.md → Caching, Performance metrics
- COMPLETION-REPORT.md → Token Efficiency Budget
- README.md → Performance Expectations

**Token Management**
- COMPLETION-REPORT.md → Token Efficiency Budget (detailed)
- QUICK-START.md → Token Management
- README.md → Token Efficiency

**Troubleshooting**
- QUICK-START.md → Troubleshooting section
- SETUP-STATUS.md → Troubleshooting section
- COMPLETION-REPORT.md → Known Issues

**Next Steps**
- QUICK-START.md → Next Steps
- COMPLETION-REPORT.md → Immediate Next Steps
- README.md → Next Steps

---

## Key Metrics Summary

### Configuration Status
- ✅ Sections complete: 10/10
- ✅ Required fields: All populated
- ✅ Template variables: 0 remaining
- ✅ Validation checks: 100% passed

### Performance Status
- Hook execution time: 26ms (target: < 2000ms)
- Timeout utilization: 1.3% (safety margin: 98.7%)
- Cache hit rate: 95%+ for git, 90%+ for version
- SessionStart latency: < 50ms

### Development Status
- TDD enforcement: Enabled
- Test coverage target: 90%
- Git workflow: GitHub Flow (personal mode)
- Token budget per feature: 130K

---

## Files in This Directory

```
.moai/config/
├── config.json                  ← Main configuration (SSOT)
├── INDEX.md                     ← This file (navigation)
├── README.md                    ← Overview and quick reference
├── QUICK-START.md               ← Daily workflow guide
├── SETUP-STATUS.md              ← Detailed setup verification
├── IMPROVEMENTS-PROPOSED.md     ← Hook performance analysis
└── COMPLETION-REPORT.md         ← Final setup verification
```

---

## Related Documentation

### In Project Root
- **CLAUDE.md** - Full Claude Code execution guide (recommended reading)
- **CLAUDE.local.md** - Local project-specific guidelines
- **README.md** - Main project documentation

### In .moai/project/
- **product.md** - Product vision
- **structure.md** - Architecture documentation
- **tech.md** - Technology stack details

### In .claude/hooks/moai/
- **session_start__config_health_check.py** - Hook implementation
- **session_start__show_project_info.py** - Hook implementation

---

## Getting Help

### Step 1: Check This Index
You're reading it! Use the navigation above to find relevant documentation.

### Step 2: Check QUICK-START.md
Most common questions answered in the troubleshooting section.

### Step 3: Check SETUP-STATUS.md
Detailed troubleshooting and configuration guidance.

### Step 4: Check COMPLETION-REPORT.md
System-wide validation and verification information.

### Step 5: Manual Verification
```bash
# Check if config is valid
cat .moai/config/config.json | python3 -m json.tool

# Check hook execution
echo '{}' | python3 .claude/hooks/moai/session_start__config_health_check.py

# View git info cache
cat .moai/cache/git-info.json
```

---

## Next Actions

### For New Users
1. Read: QUICK-START.md (5-10 minutes)
2. Run: One-time setup (5 minutes)
3. Create: First SPEC (10-15 minutes)
4. Review: SETUP-STATUS.md for details

### For Existing Users
1. Reference: QUICK-START.md for commands
2. Check: COMPLETION-REPORT.md for metrics
3. Optimize: Token efficiency using CLAUDE.md guidelines

### For System Maintenance
1. Review: IMPROVEMENTS-PROPOSED.md for performance
2. Validate: SETUP-STATUS.md verification checklist
3. Monitor: COMPLETION-REPORT.md metrics

---

## Summary

This directory contains complete configuration and setup documentation for MoAI-ADK:

- **config.json** - Working configuration (no edits needed)
- **QUICK-START.md** - Start developing immediately
- **SETUP-STATUS.md** - Detailed technical guide
- **IMPROVEMENTS-PROPOSED.md** - Performance analysis
- **COMPLETION-REPORT.md** - Full verification
- **README.md** - Overview and reference
- **INDEX.md** - This navigation guide

**Status**: ✅ Configuration Complete
**Ready for**: SPEC-First TDD Development
**Next Step**: Read QUICK-START.md or start with `/moai:1-plan "기능 설명"`

---

**Configuration Index Created**: 2025-11-19
**Version**: 0.26.0
**Total Documentation**: ~68KB across 7 files
**Navigation**: Use the sections above to find what you need
