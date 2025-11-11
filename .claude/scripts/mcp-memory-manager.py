#!/usr/bin/env python3
"""
MCP Memory Manager - Prevents memory leaks from MCP server proliferation
Monitors and manages MCP (Model Context Protocol) servers to prevent memory exhaustion
"""

import os
import sys
import subprocess
import psutil
import time
import signal
import json
from pathlib import Path
from datetime import datetime, timedelta

class MCPMemoryManager:
    """Manages MCP server processes to prevent memory leaks"""
    
    def __init__(self):
        self.home_dir = Path.home()
        self.max_mcp_per_session = 3  # context7, playwright, sequential-thinking
        self.max_memory_mb = 2048  # Max memory per MCP cluster
        self.session_timeout_minutes = 60
        
    def get_mcp_processes(self):
        """Get all running MCP server processes"""
        mcp_processes = []
        for proc in psutil.process_iter(['pid', 'name', 'cmdline', 'memory_info', 'create_time']):
            try:
                cmdline = ' '.join(proc.info['cmdline'] or [])
                if any(keyword in cmdline.lower() for keyword in [
                    'context7-mcp', 'mcp-server-playwright', 'mcp-server-sequential-thinking'
                ]):
                    mcp_processes.append({
                        'pid': proc.info['pid'],
                        'name': proc.info['name'],
                        'cmdline': cmdline,
                        'memory_mb': proc.info['memory_info'].rss / 1024 / 1024,
                        'created': datetime.fromtimestamp(proc.info['create_time']),
                        'age_minutes': (datetime.now() - datetime.fromtimestamp(proc.info['create_time'])).total_seconds() / 60
                    })
            except (psutil.NoSuchProcess, psutil.AccessDenied):
                continue
        return mcp_processes
    
    def get_playwright_processes(self):
        """Get all Playwright browser processes"""
        playwright_procs = []
        for proc in psutil.process_iter(['pid', 'name', 'memory_info', 'create_time']):
            try:
                if proc.info['name'] and 'headless_shell' in proc.info['name']:
                    playwright_procs.append({
                        'pid': proc.info['pid'],
                        'name': proc.info['name'],
                        'memory_mb': proc.info['memory_info'].rss / 1024 / 1024,
                        'created': datetime.fromtimestamp(proc.info['create_time']),
                        'age_minutes': (datetime.now() - datetime.fromtimestamp(proc.info['create_time'])).total_seconds() / 60
                    })
            except (psutil.NoSuchProcess, psutil.AccessDenied):
                continue
        return playwright_procs
    
    def group_by_session(self, processes):
        """Group processes by Claude session (based on start time proximity)"""
        sessions = {}
        for proc in processes:
            # Group by minute-level granularity to identify session clusters
            session_key = proc['created'].strftime('%Y%m%d_%H%M')
            if session_key not in sessions:
                sessions[session_key] = []
            sessions[session_key].append(proc)
        return sessions
    
    def cleanup_orphaned_processes(self):
        """Clean up orphaned MCP and Playwright processes"""
        print("🔍 Scanning for orphaned MCP processes...")
        
        mcp_procs = self.get_mcp_processes()
        playwright_procs = self.get_playwright_processes()
        
        # Group MCP processes by session
        mcp_sessions = self.group_by_session(mcp_procs)
        
        cleaned_count = 0
        memory_freed = 0
        
        # Check each MCP session
        for session_key, session_procs in mcp_sessions.items():
            session_age = max(proc['age_minutes'] for proc in session_procs)
            session_memory = sum(proc['memory_mb'] for proc in session_procs)
            proc_count = len(session_procs)
            
            # Clean up if session is too old or has too many processes
            if session_age > self.session_timeout_minutes or proc_count > self.max_mcp_per_session:
                print(f"🧹 Cleaning session {session_key}: {proc_count} procs, {session_memory:.1f}MB, {session_age:.1f}min old")
                for proc in session_procs:
                    try:
                        os.kill(proc['pid'], signal.SIGTERM)
                        cleaned_count += 1
                        memory_freed += proc['memory_mb']
                        time.sleep(0.1)  # Gentle termination
                    except (OSError, ProcessLookupError):
                        pass
        
        # Clean up old Playwright processes
        for proc in playwright_procs:
            if proc['age_minutes'] > 30:  # Playwright processes older than 30 minutes
                try:
                    os.kill(proc['pid'], signal.SIGTERM)
                    cleaned_count += 1
                    memory_freed += proc['memory_mb']
                    print(f"🧹 Killed Playwright process {proc['pid']} ({proc['memory_mb']:.1f}MB)")
                except (OSError, ProcessLookupError):
                    pass
        
        if cleaned_count > 0:
            print(f"✅ Cleaned up {cleaned_count} processes, freed {memory_freed:.1f}MB memory")
        else:
            print("✅ No orphaned processes found")
            
        return cleaned_count, memory_freed
    
    def cleanup_npm_cache(self):
        """Clean up npm cache to prevent disk space issues"""
        npm_cache_dir = self.home_dir / '.npm' / '_npx'
        if npm_cache_dir.exists():
            try:
                size_before = sum(f.stat().st_size for f in npm_cache_dir.rglob('*') if f.is_file())
                # Keep only recent cache entries (last 7 days)
                cutoff_time = datetime.now() - timedelta(days=7)
                cleaned_dirs = 0
                
                for item in npm_cache_dir.iterdir():
                    if item.is_dir():
                        stat = item.stat()
                        if datetime.fromtimestamp(stat.st_mtime) < cutoff_time:
                            import shutil
                            shutil.rmtree(item)
                            cleaned_dirs += 1
                
                size_after = sum(f.stat().st_size for f in npm_cache_dir.rglob('*') if f.is_file())
                space_freed = (size_before - size_after) / 1024 / 1024
                
                if cleaned_dirs > 0:
                    print(f"🧹 Cleaned {cleaned_dirs} npm cache directories, freed {space_freed:.1f}MB")
                    
            except Exception as e:
                print(f"⚠️ Error cleaning npm cache: {e}")
    
    def monitor_memory_usage(self):
        """Monitor current memory usage"""
        memory = psutil.virtual_memory()
        print(f"💾 Current memory usage: {memory.percent:.1f}% ({memory.used/1024/1024/1024:.1f}GB used, {memory.available/1024/1024:.1f}MB free)")
        
        # Alert if memory usage is high
        if memory.percent > 80:
            print("⚠️ High memory usage detected! Running cleanup...")
            self.cleanup_orphaned_processes()
            self.cleanup_npm_cache()
    
    def run_daemon(self, interval_minutes=5):
        """Run as daemon to continuously monitor and clean"""
        print(f"🚀 MCP Memory Manager started (monitoring every {interval_minutes} minutes)")
        
        while True:
            try:
                self.monitor_memory_usage()
                time.sleep(interval_minutes * 60)
            except KeyboardInterrupt:
                print("\n👋 MCP Memory Manager stopped")
                break
            except Exception as e:
                print(f"⚠️ Error in monitoring loop: {e}")
                time.sleep(60)  # Wait before retrying

def main():
    if len(sys.argv) < 2:
        print("Usage: python mcp-memory-manager.py [cleanup|monitor|daemon]")
        sys.exit(1)
    
    manager = MCPMemoryManager()
    command = sys.argv[1].lower()
    
    if command == "cleanup":
        manager.cleanup_orphaned_processes()
        manager.cleanup_npm_cache()
        
    elif command == "monitor":
        manager.monitor_memory_usage()
        
    elif command == "daemon":
        interval = int(sys.argv[2]) if len(sys.argv) > 2 else 5
        manager.run_daemon(interval)
        
    else:
        print("Invalid command. Use: cleanup, monitor, or daemon")

if __name__ == "__main__":
    main()
