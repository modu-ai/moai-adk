# MoAI-ADK Memory Management Guide

## 🚨 Critical: Memory Leak Prevention

MoAI-ADK uses MCP (Model Context Protocol) servers which can cause memory leaks if not properly managed. This guide explains how to prevent and resolve memory issues.

## Root Cause Analysis

### MCP Server Proliferation
- Each Claude session spawns 3 MCP servers: context7, playwright, sequential-thinking
- Multiple Claude terminals = multiple server clusters
- Each cluster uses 1-1.5GB memory
- Servers may not terminate properly when sessions end

### Memory Usage Breakdown
- Claude process: 2-3GB per session
- MCP servers: 1-1.5GB per session  
- npm cache: 4GB+ (accumulates over time)
- Playwright browsers: 400-500MB each

## 🛠️ Prevention & Solutions

### 1. Automated Memory Management

The MCP Memory Manager automatically monitors and cleans up orphaned processes:

```bash
# Monitor current memory usage
python3 .claude/scripts/mcp-memory-manager.py monitor

# Clean up orphaned processes
python3 .claude/scripts/mcp-memory-manager.py cleanup

# Run as daemon (monitor every 5 minutes)
python3 .claude/scripts/mcp-memory-manager.py daemon 5
```

### 2. Emergency Cleanup

For critical memory situations:

```bash
# Quick emergency cleanup
./.claude/scripts/emergency-memory-cleanup.sh
```

### 3. Manual Process Management

```bash
# Check MCP processes
ps aux | grep -E "(context7|playwright|sequential-thinking)" | grep -v grep

# Kill specific MCP processes
pkill -f "context7-mcp"
pkill -f "mcp-server-playwright" 
pkill -f "mcp-server-sequential-thinking"

# Clean npm cache
npm cache clean --force
rm -rf ~/.npm/_npx/*
```

### 4. Claude Session Management

- **Limit concurrent Claude sessions**: Maximum 2-3 sessions
- **Close unused terminals**: Properly terminate Claude sessions
- **Restart periodically**: Restart Claude every few hours of heavy use

## ⚙️ Configuration

### Memory Manager Settings
Edit `.claude/scripts/mcp-memory-manager.py` to customize:

```python
self.max_mcp_per_session = 3          # Max MCP servers per session
self.max_memory_mb = 2048            # Max memory per MCP cluster  
self.session_timeout_minutes = 60    # Session timeout for cleanup
```

### Automatic Monitoring
Add to your shell startup (.zshrc or .bashrc):

```bash
# Start memory manager daemon in background
python3 ~/MoAI/MoAI-ADK/.claude/scripts/mcp-memory-manager.py daemon 10 &
```

## 🚨 Warning Signs

Monitor for these symptoms:

- Memory usage > 80% for extended periods
- "Aborted()" errors in Claude
- System becomes sluggish
- Multiple similar MCP processes running

## 🔍 Troubleshooting

### Check Current Status
```bash
# Memory usage
python3 -c "import psutil; m=psutil.virtual_memory(); print(f'{m.percent:.1f}% used, {m.available//1024//1024}MB free')"

# MCP processes
ps aux | grep -E "(context7|playwright|sequential-thinking)" | wc -l

# npm cache size
du -sh ~/.npm/_npx/
```

### Common Issues

1. **High memory usage**
   - Run emergency cleanup script
   - Check for orphaned MCP processes
   - Restart Claude sessions

2. **MCP servers not terminating**
   - Manual process killing
   - Memory manager daemon
   - System restart if needed

3. **npm cache growing**
   - Regular npm cache cleaning
   - Memory manager automatic cleanup

## 📊 Best Practices

1. **Daily cleanup**: Run memory manager cleanup daily
2. **Session discipline**: Close unused Claude terminals  
3. **Monitor memory**: Check usage before starting new sessions
4. **Regular restarts**: Restart Claude every 3-4 hours
5. **Cache management**: Clean npm cache weekly

## 🆘 Emergency Contacts

If memory issues persist:
1. Run emergency cleanup script
2. Restart system if needed
3. Check for other memory-intensive applications
4. Consider increasing system RAM if problem is chronic

Remember: Prevention is better than cure. Use the memory manager daemon for automatic protection.
