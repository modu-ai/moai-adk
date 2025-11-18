# Extended Troubleshooting Guide

**Common issues, error messages, and solutions for MoAI-ADK and Claude Code integration.**

> **See also**: CLAUDE.md → "Enhanced Troubleshooting" for quick FAQs

---

## Agent Execution Issues

### Agent Not Found

**Error**: `AgentNotFound: spec-builder not in registry`

**Solutions**:
```bash
# 1. Verify agent structure
ls -la .claude/agents/

# 2. Check YAML frontmatter
cat .claude/agents/alfred/cc-manager.md | head -20

# 3. Restart Claude Code
# Close and reopen Claude Code

# 4. Verify config.json
cat .moai/config/config.json | jq '.agents'
```

### Agent Timeout

**Error**: `TaskTimeout: spec-builder exceeded 300s`

**Solutions**:
```bash
# 1. Check system resources
top -l 1 | grep -E '^CPU|^Mem'

# 2. Reduce task complexity
# Split into smaller sub-tasks

# 3. Increase timeout (if supported)
# Task(timeout=600000)

# 4. Check agent logs
tail -100f .moai/logs/agent-execution.log
```

### Agent Memory Issues

**Error**: `OutOfMemory: Agent context exceeded limit`

**Solutions**:
```bash
# 1. Clear unnecessary files from context
/clear

# 2. Use `/compact` to compress conversation
/compact

# 3. Remove large files from context
/remove-dir node_modules/

# 4. Check memory usage
/memory
```

---

## Skill Loading Issues

### Skill Not Found

**Error**: `SkillNotFound: moai-lang-python`

**Solutions**:
```bash
# 1. Verify skill installation
ls -la .claude/skills/moai-lang-python/

# 2. Check SKILL.md existence
cat .claude/skills/moai-lang-python/SKILL.md | head -5

# 3. Restart Claude Code
# Skills auto-reload on restart

# 4. Update Claude Code
# Check Claude Code version compatibility
```

### Skill Initialization Failure

**Error**: `SkillInitFailed: moai-domain-backend`

**Solutions**:
```bash
# 1. Check Skill dependencies
cat .claude/skills/moai-domain-backend/.dependencies

# 2. Verify Context7 MCP connection
/mcp restart

# 3. Check Claude Code compatibility
claude --version

# 4. Review Skill logs
tail -50f .moai/logs/skill-init.log
```

---

## MCP Connection Issues

### MCP Server Offline

**Error**: `MCPConnectionError: context7-mcp failed to initialize`

**Solutions**:
```bash
# 1. Check MCP server status
claude mcp serve

# 2. Validate MCP configuration
cat .claude/mcp.json | jq '.mcpServers'

# 3. Restart MCP servers
/mcp restart

# 4. Check network connectivity
curl -I https://api.context7.io

# 5. Review MCP logs
cat .moai/logs/mcp-init.log
```

### Context7 Documentation Fetch Fails

**Error**: `DocumentFetchFailed: /facebook/react`

**Solutions**:
```bash
# 1. Verify library ID format
# Correct: /org/project/version
# Example: /facebook/react/19.0.0

# 2. Check Context7 session cache
ls -la .moai/sessions/context7-*

# 3. Clear outdated cache
rm -rf .moai/sessions/context7-*

# 4. Retry with explicit version
mcp__context7__get-library-docs("/facebook/react/18.2.0")
```

---

## Context & Memory Issues

### Context Window Overflow

**Error**: `ContextOverflow: 200,000 tokens exceeded`

**Solutions**:
```bash
# 1. Check current usage
/context

# 2. Clear conversation history
/clear

# 3. Use /compact before hitting limit
/compact

# 4. Phase implementation with /clear:
# Phase 1: SPEC (50K tokens) → /clear
# Phase 2: Implementation (100K tokens) → /clear
# Phase 3: Testing (50K tokens)
```

### Memory File Not Persisting

**Error**: Memory changes lost after session restart

**Solutions**:
```bash
# 1. Verify memory file location
cat .claude/memory.md

# 2. Ensure memory file is readable
ls -la .claude/memory.md

# 3. Check .gitignore (memory should NOT be ignored)
grep -n "memory.md" .gitignore

# 4. Manually add to Claude Code
# Claude Code → Settings → Memory File
```

---

## Git Workflow Issues

### Git Push Permission Denied

**Error**: `fatal: could not read Password`

**Solutions**:
```bash
# 1. Setup SSH keys
ssh-keygen -t ed25519 -C "your-email@example.com"

# 2. Add SSH key to GitHub
# GitHub → Settings → SSH Keys

# 3. Test SSH connection
ssh -T git@github.com

# 4. Use SSH for git remote
git remote set-url origin git@github.com:user/repo.git
```

### Branch Diverged

**Error**: `fatal: refusing to merge unrelated histories`

**Solutions**:
```bash
# 1. Check branch status
git log --oneline -10

# 2. Rebase instead of merge (for feature branches)
git rebase main

# 3. If merge necessary, use --allow-unrelated-histories
git merge --allow-unrelated-histories main

# 4. For Personal Mode, ensure based on main
git checkout main && git pull
```

---

## SPEC & TDD Issues

### SPEC File Syntax Error

**Error**: `SPECParseError: Invalid EARS format`

**Solutions**:
```markdown
# Verify EARS format:
✅ Ubiquitous: The system SHALL ...
✅ Event-Driven: WHEN ... → ...
✅ Unwanted: IF ... → THEN ...
✅ State-Driven: WHILE ... → ...
✅ Optional: WHERE ... → ...
```

### Test Failure in /moai:2-run

**Error**: `TestFailed: 10/100 tests failed`

**Solutions**:
```bash
# 1. Review test output
less .moai/reports/test-results.log

# 2. Run tests locally
uv run pytest -v

# 3. Debug specific test
uv run pytest -v tests/test_module.py::test_function

# 4. Check SPEC compliance
cat .moai/specs/SPEC-001/spec.md
```

---

## Performance Issues

### Slow Task Execution

**Solutions**:
```bash
# 1. Check system resources
# CPU, Memory, Disk usage

# 2. Monitor token usage
/context

# 3. Use Haiku for fast tasks
Task(subagent_type="...", model="haiku")

# 4. Reduce codebase size in context
/remove-dir tests/
```

### High API Costs

**Solutions**:
```bash
# 1. Check cost breakdown
/cost

# 2. Use Haiku for exploration
# Haiku: $0.0008/1K tokens vs Sonnet: $0.003/1K

# 3. Use agent delegation
# Smaller focused tasks = lower token usage

# 4. Monitor usage
/usage
```

---

## Documentation & Troubleshooting

### Find Log Files
```bash
# Agent execution logs
ls -la .moai/logs/agent-*.log

# MCP initialization
cat .moai/logs/mcp-init.log

# Performance reports
cat .moai/reports/performance-*.json

# Error reports
cat .moai/reports/errors-*.log
```

### Enable Debug Mode
```bash
# Claude Code debug
claude --debug

# Set environment variables
export DEBUG=moai-adk:*
export LOG_LEVEL=debug

# Check verbose output
/moai:2-run SPEC-001 --verbose
```

---

## Getting Help

1. **Check logs first**: `.moai/logs/`, `.moai/reports/`
2. **Review CLAUDE.md**: Comprehensive guides
3. **Check memory files**: `.moai/memory/`
4. **GitHub Issues**: https://github.com/anthropics/claude-code/issues

---

**Last Updated**: 2025-11-18
**Format**: Markdown | **Language**: English
**Scope**: MoAI-ADK & Claude Code v4.0 Integration Troubleshooting
