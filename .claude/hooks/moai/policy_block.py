#!/usr/bin/env python3
"""
MoAI-ADK Policy Block Hook - v0.1.12
위험한 명령어 차단, 정책 검증, Constitution 보호

PreToolUse Hook으로 위험한 Bash 명령어와 중요한 문서 수정을 차단합니다.
"""

import json
import sys
import re
import os
from pathlib import Path
from typing import List, Dict, Any, Optional

class PolicyBlocker:
    """MoAI-ADK 정책 차단 시스템"""
    
    def __init__(self, project_root: Path):
        self.project_root = project_root
        self.moai_config = project_root / ".moai" / "config.json"
        self.config = self.load_config()
        
        # 위험한 명령어 패턴 정의
        self.dangerous_commands = [
            # 시스템 파괴적 명령어
            r'rm\s+-rf\s+/',
            r'sudo\s+rm',
            r'dd\s+if=/dev/zero',
            r':\(\)\{:\|:&\};:',  # Fork bomb
            r'>\s*/dev/sd[a-z]',
            r'mkfs\.',
            r'fdisk\s+-l',
            
            # 네트워크 위험 명령어
            r'curl\s+.*\|\s*bash',
            r'wget\s+.*\|\s*sh',
            r'nc\s+.*-e',
            
            # 권한 변경 위험
            r'chmod\s+777\s+/',
            r'chown\s+-R\s+.*/',
            
            # 환경 변수 조작
            r'export\s+PATH=',
            r'unset\s+PATH',
        ]
        
        # 보호된 경로 패턴
        self.protected_paths = [
            r'\.moai/steering/',
            r'\.moai/memory/constitution\.md',
            r'\.claude/settings\.json',
            r'\.claude/hooks/',
            r'/etc/',
            r'/usr/',
            r'/var/',
            r'/root/',
        ]
    
    def load_config(self) -> Dict[str, Any]:
        """설정 파일 로드"""
        if self.moai_config.exists():
            try:
                with open(self.moai_config, 'r', encoding='utf-8') as f:
                    config = json.load(f)
                    return config.get('policy', {})
            except (json.JSONDecodeError, FileNotFoundError):
                pass
        return self.get_default_policy()
    
    def get_default_policy(self) -> Dict[str, Any]:
        """기본 정책 설정"""
        return {
            "enabled": True,
            "block_dangerous_commands": True,
            "protect_steering_docs": True,
            "protect_constitution": True,
            "block_system_paths": True,
            "require_constitution_checklist": True,
            "log_blocked_attempts": True
        }
    
    def check_dangerous_bash_command(self, command: str) -> tuple[bool, str]:
        """위험한 Bash 명령어 검사"""
        if not self.config.get("block_dangerous_commands", True):
            return True, "Dangerous command blocking disabled"
        
        for pattern in self.dangerous_commands:
            if re.search(pattern, command, re.IGNORECASE):
                return False, f"위험한 명령어 패턴 감지: {pattern}"
        
        return True, "Command safe"
    
    def check_protected_file_access(self, file_path: str) -> tuple[bool, str]:
        """보호된 파일 접근 검사"""
        if not self.config.get("protect_steering_docs", True) and not self.config.get("protect_constitution", True):
            return True, "File protection disabled"
        
        # 파일 경로 정규화 (Path Traversal 공격 방지)
        try:
            normalized_path = Path(file_path).resolve()
            project_root_resolved = self.project_root.resolve()
            
            # 프로젝트 루트 밖의 파일 접근 차단
            if not str(normalized_path).startswith(str(project_root_resolved)):
                return False, "프로젝트 외부 파일 접근이 차단되었습니다"
        except (OSError, ValueError):
            return False, "유효하지 않은 파일 경로입니다"
        
        # Steering 문서 보호
        if self.config.get("protect_steering_docs", True):
            steering_path = project_root_resolved / ".moai" / "steering"
            if str(normalized_path).startswith(str(steering_path)):
                return False, "Steering 문서는 /moai:project setting 명령으로만 수정 가능합니다"
        
        # Constitution 보호
        if self.config.get("protect_constitution", True):
            if normalized_path.name == 'constitution.md':
                checklist_path = self.project_root / '.moai' / 'memory' / 'constitution_update_checklist.md'
                if not checklist_path.exists():
                    return False, "Constitution 변경은 체크리스트 완성 후에만 가능합니다"
        
        # 시스템 경로 보호 (정규화된 경로 사용)
        if self.config.get("block_system_paths", True):
            normalized_path_str = str(normalized_path)
            for pattern in self.protected_paths:
                if re.search(pattern, normalized_path_str):
                    return False, f"보호된 경로에 대한 접근이 차단되었습니다: {pattern}"
        
        return True, "File access allowed"
    
    def check_hook_modification(self, file_path: str) -> tuple[bool, str]:
        """Hook 파일 수정 검사"""
        try:
            normalized_path = Path(file_path).resolve()
            project_root_resolved = self.project_root.resolve()
            hooks_path = project_root_resolved / ".claude" / "hooks"
            
            if str(normalized_path).startswith(str(hooks_path)):
                # Hook 파일은 claude-code-manager 에이전트를 통해서만 수정
                return False, "Hook 파일은 claude-code-manager 에이전트를 통해서만 수정 가능합니다"
        except (OSError, ValueError):
            return False, "유효하지 않은 파일 경로입니다"
        
        return True, "Hook modification allowed"
    
    def log_blocked_attempt(self, tool_name: str, reason: str, details: Dict[str, Any]):
        """차단된 시도 로깅"""
        if not self.config.get("log_blocked_attempts", True):
            return
        
        log_dir = self.project_root / '.claude' / 'logs'
        log_dir.mkdir(exist_ok=True)
        
        log_file = log_dir / 'policy_blocks.json'
        
        log_entry = {
            "timestamp": __import__('datetime').datetime.now().isoformat(),
            "tool": tool_name,
            "reason": reason,
            "details": details
        }
        
        logs = []
        if log_file.exists():
            try:
                with open(log_file, 'r', encoding='utf-8') as f:
                    logs = json.load(f)
            except (json.JSONDecodeError, FileNotFoundError):
                logs = []
        
        logs.append(log_entry)
        
        # 로그 파일이 너무 커지지 않도록 최대 1000개 항목 유지
        if len(logs) > 1000:
            logs = logs[-1000:]
        
        try:
            with open(log_file, 'w', encoding='utf-8') as f:
                json.dump(logs, f, indent=2, ensure_ascii=False)
        except Exception as e:
            print(f"Warning: Failed to write policy log: {e}", file=sys.stderr)

def handle_pre_tool_use():
    """PreToolUse Hook 메인 핸들러"""
    try:
        # stdin에서 툴 사용 데이터 읽기
        data = json.loads(sys.stdin.read())
        
        tool_name = data.get('tool_name', '')
        tool_input = data.get('tool_input', {})
        
        # 프로젝트 루트 찾기
        project_root = Path(os.environ.get('CLAUDE_PROJECT_DIR', Path.cwd()))
        current_dir = project_root
        
        # .moai 디렉토리를 찾을 때까지 상위로 올라가기
        while not (project_root / '.moai').exists() and project_root.parent != project_root:
            project_root = project_root.parent
        
        if not (project_root / '.moai').exists():
            # MoAI 프로젝트가 아니면 통과
            sys.exit(0)
        
        blocker = PolicyBlocker(project_root)
        
        # Bash 명령어 검사
        if tool_name == 'Bash':
            raw_command = tool_input.get('command', '')
            if isinstance(raw_command, list):
                command = " ".join(str(part) for part in raw_command)
            else:
                command = str(raw_command)
            is_safe, reason = blocker.check_dangerous_bash_command(command)
            
            if not is_safe:
                print(f"🚫 위험한 명령어가 차단되었습니다: {reason}", file=sys.stderr)
                blocker.log_blocked_attempt(tool_name, reason, {'command': command})
                sys.exit(2)  # 차단
        
        # 파일 수정 도구 검사
        if tool_name in ['Write', 'Edit', 'MultiEdit']:
            file_path = tool_input.get('file_path', '')
            
            if file_path:
                # 보호된 파일 접근 검사
                is_allowed, reason = blocker.check_protected_file_access(file_path)
                if not is_allowed:
                    print(f"🔒 파일 접근이 차단되었습니다: {reason}", file=sys.stderr)
                    blocker.log_blocked_attempt(tool_name, reason, {'file_path': file_path})
                    sys.exit(2)  # 차단
                
                # Hook 수정 검사
                is_allowed, reason = blocker.check_hook_modification(file_path)
                if not is_allowed:
                    print(f"⚙️ Hook 수정이 차단되었습니다: {reason}", file=sys.stderr)
                    blocker.log_blocked_attempt(tool_name, reason, {'file_path': file_path})
                    sys.exit(2)  # 차단
        
        # WebFetch 남용 검사
        if tool_name == 'WebFetch':
            url = tool_input.get('url', '')
            if url and any(danger in url.lower() for danger in ['exec', 'eval', 'script']):
                print("🌐 의심스러운 웹 요청이 차단되었습니다", file=sys.stderr)
                blocker.log_blocked_attempt(tool_name, "Suspicious URL pattern", {'url': url})
                sys.exit(2)  # 차단
        
        # 모든 검사 통과
        sys.exit(0)
        
    except KeyboardInterrupt:
        sys.exit(1)
    except Exception as e:
        print(f"Policy hook error: {e}", file=sys.stderr)
        # 에러가 발생해도 작업을 차단하지는 않음 (fail-open)
        sys.exit(0)

if __name__ == "__main__":
    handle_pre_tool_use()
