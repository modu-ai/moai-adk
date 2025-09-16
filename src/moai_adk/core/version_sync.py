#!/usr/bin/env python3
"""
MoAI-ADK Automated Version Synchronization System
Automated version synchronization system for MoAI-ADK project files
"""

import re
import json
from pathlib import Path
from typing import Dict, List, Tuple, Optional
from .._version import __version__, VERSIONS, VERSION_FORMATS


class VersionSyncManager:
    """Version synchronization manager class"""
    
    def __init__(self, project_root: Optional[str] = None):
        """
        Initialize version sync manager
        
        Args:
            project_root: Project root directory. Auto-detected if None
        """
        self.project_root = Path(project_root) if project_root else self._find_project_root()
        self.current_version = __version__
        self.version_patterns = self._load_version_patterns()
        self.sync_log = []
        
    def _find_project_root(self) -> Path:
        """Find project root directory containing pyproject.toml"""
        current = Path(__file__).parent
        
        while current != current.parent:
            if (current / "pyproject.toml").exists():
                return current
            current = current.parent
            
        raise FileNotFoundError("pyproject.toml을 찾을 수 없습니다")
        
    def _load_version_patterns(self) -> Dict[str, List[Dict]]:
        """Define version patterns - file-specific replacement rules"""
        return {
            # Python package configuration
            "pyproject.toml": [
                {
                    "pattern": r'version\s*=\s*"[^"]*"',
                    "replacement": f'version = "{self.current_version}"',
                    "description": "Python package version"
                }
            ],
            
            # Python source files
            "**/*.py": [
                {
                    "pattern": r'__version__\s*=\s*"[^"]*"',
                    "replacement": f'__version__ = "0.1.17"',
                    "description": "Python module version"
                },
                {
                    "pattern": r'def get_version\([^)]*\):\s*return\s*"[^"]*"',
                    "replacement": f'def get_version(component="moai_adk"): return "0.1.17"',
                    "description": "Version function return value"
                },
                {
                    "pattern": r'"moai_version":\s*"[^"]*"',
                    "replacement": f'"moai_version": "0.1.17"',
                    "description": "Configuration moai_version"
                }
            ],
            
            # Markdown documents
            "**/*.md": [
                {
                    "pattern": r'MoAI-ADK \(MoAI Agentic Development Kit\) v[0-9]+\.[0-9]+\.[0-9]+',
                    "replacement": f'MoAI-ADK (MoAI Agentic Development Kit) v{self.current_version}',
                    "description": "MoAI-ADK full title version"
                },
                {
                    "pattern": r'MoAI-ADK v[0-9]+\.[0-9]+\.[0-9]+',
                    "replacement": f'MoAI-ADK v{self.current_version}',
                    "description": "MoAI-ADK version in documentation"
                },
                {
                    "pattern": r'version-[0-9]+\.[0-9]+\.[0-9]+-blue',
                    "replacement": f'version-{self.current_version}-blue',
                    "description": "Version badge"
                },
                {
                    "pattern": r'moai-adk-v[0-9]+\.[0-9]+\.[0-9]+',
                    "replacement": f'moai-adk-v{self.current_version}',
                    "description": "Release archive naming"
                },
                {
                    "pattern": r'\*\*MoAI-ADK 버전\*\*: v[0-9]+\.[0-9]+\.[0-9]+',
                    "replacement": f'**MoAI-ADK 버전**: v{self.current_version}',
                    "description": "Korean version footer"
                }
            ],
            
            # JSON configuration files
            "**/*.json": [
                {
                    "pattern": r'"version":\s*"[^"]*"',
                    "replacement": f'"version": "{self.current_version}"',
                    "description": "JSON version field"
                },
                {
                    "pattern": r'"moai_version":\s*"[^"]*"',
                    "replacement": f'"moai_version": "0.1.17"',
                    "description": "MoAI specific version field"
                },
                {
                    "pattern": r'"moai_adk_version":\s*"[^"]*"',
                    "replacement": f'"moai_adk_version": "{self.current_version}"',
                    "description": "MoAI ADK version field"
                }
            ],
            
            # GitHub Actions workflows
            "**/*.yml": [
                {
                    "pattern": r'v[0-9]+\.[0-9]+\.[0-9]+',
                    "replacement": f'v{self.current_version}',
                    "description": "YAML version tags"
                },
                {
                    "pattern": r'MoAI-ADK v[0-9]+\.[0-9]+\.[0-9]+',
                    "replacement": f'MoAI-ADK v{self.current_version}',
                    "description": "MoAI-ADK version in YAML"
                }
            ],
            
            # Makefile
            "Makefile": [
                {
                    "pattern": r'MoAI-ADK v[0-9]+\.[0-9]+\.[0-9]+',
                    "replacement": f'MoAI-ADK v{self.current_version}',
                    "description": "Makefile version display"
                }
            ],
            
            # CHANGELOG
            "CHANGELOG.md": [
                {
                    "pattern": r'MoAI-ADK v[0-9]+\.[0-9]+\.[0-9]+',
                    "replacement": f'MoAI-ADK v{self.current_version}',
                    "description": "Changelog version references"
                }
            ]
        }
    
    def sync_all_versions(self, dry_run: bool = False) -> Dict[str, List[str]]:
        """
        전체 프로젝트의 버전 정보 동기화
        
        Args:
            dry_run: True면 실제 변경하지 않고 시뮬레이션만
            
        Returns:
            Dict[파일패턴, 변경된파일리스트]
        """
        results = {}
        
        print(f"🗿 MoAI-ADK 버전 동기화 시작: v{self.current_version}")
        print(f"프로젝트 루트: {self.project_root}")
        
        for pattern, replacements in self.version_patterns.items():
            files_changed = self._sync_pattern(pattern, replacements, dry_run)
            if files_changed:
                results[pattern] = files_changed
                
        if dry_run:
            print("\\n✅ 드라이 런 완료 - 실제 파일은 변경되지 않았습니다")
        else:
            print("\\n✅ 버전 동기화 완료")
            
        return results
    
    def _sync_pattern(self, file_pattern: str, replacements: List[Dict], dry_run: bool) -> List[str]:
        """특정 파일 패턴에 대해 버전 동기화 수행"""
        changed_files = []
        
        if file_pattern.startswith("**"):
            # glob 패턴으로 파일 검색
            files = list(self.project_root.glob(file_pattern))
        else:
            # 단일 파일
            files = [self.project_root / file_pattern]
            files = [f for f in files if f.exists()]
        
        for file_path in files:
            if self._should_skip_file(file_path):
                continue
                
            try:
                changed = self._sync_file(file_path, replacements, dry_run)
                if changed:
                    changed_files.append(str(file_path.relative_to(self.project_root)))
                    print(f"  ✓ {file_path.relative_to(self.project_root)}")
                    
            except Exception as e:
                print(f"  ❌ {file_path.relative_to(self.project_root)}: {e}")
                
        return changed_files
    
    def _should_skip_file(self, file_path: Path) -> bool:
        """파일 스킵 조건 확인"""
        skip_patterns = [
            ".git/",
            "__pycache__/",
            ".pytest_cache/",
            "node_modules/",
            ".venv/",
            "venv/",
            "dist/",
            "build/",
            ".mypy_cache/"
        ]
        
        file_str = str(file_path)
        return any(pattern in file_str for pattern in skip_patterns)
    
    def _sync_file(self, file_path: Path, replacements: List[Dict], dry_run: bool) -> bool:
        """단일 파일의 버전 동기화"""
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                content = f.read()
        except UnicodeDecodeError:
            # 바이너리 파일은 스킵
            return False
            
        original_content = content
        changes_made = False
        
        for replacement_rule in replacements:
            pattern = replacement_rule["pattern"]
            replacement = replacement_rule["replacement"]
            
            if re.search(pattern, content):
                content = re.sub(pattern, replacement, content)
                changes_made = True
                
        if changes_made and not dry_run:
            with open(file_path, 'w', encoding='utf-8') as f:
                f.write(content)
                
        return changes_made
    
    def verify_sync(self) -> Dict[str, List[str]]:
        """버전 동기화 검증 - 남은 불일치 확인"""
        print(f"\\n🔍 버전 동기화 검증 중...")
        
        inconsistent_files = {}
        target_patterns = [
            (r'v[0-9]+\.[0-9]+\.[0-9]+', f"v{self.current_version}"),
            (r'version.*[0-9]+\.[0-9]+\.[0-9]+', f"version {self.current_version}"),
            (r'MoAI-ADK v[0-9]+\.[0-9]+\.[0-9]+', f"MoAI-ADK v{self.current_version}")
        ]
        
        for pattern, expected in target_patterns:
            mismatches = self._find_version_mismatches(pattern, expected)
            if mismatches:
                inconsistent_files[pattern] = mismatches
                
        if inconsistent_files:
            print("⚠️  다음 파일에서 버전 불일치가 발견되었습니다:")
            for pattern, files in inconsistent_files.items():
                print(f"  패턴: {pattern}")
                for file_info in files:
                    print(f"    {file_info}")
        else:
            print("✅ 모든 버전 정보가 일치합니다")
            
        return inconsistent_files
    
    def _find_version_mismatches(self, pattern: str, expected: str) -> List[str]:
        """버전 불일치 파일 찾기"""
        mismatches = []
        
        for file_path in self.project_root.glob("**/*"):
            if not file_path.is_file() or self._should_skip_file(file_path):
                continue
                
            try:
                with open(file_path, 'r', encoding='utf-8') as f:
                    content = f.read()
                    
                matches = re.findall(pattern, content, re.IGNORECASE)
                unexpected_matches = [m for m in matches if m != expected.split()[-1]]
                
                if unexpected_matches:
                    rel_path = file_path.relative_to(self.project_root)
                    mismatches.append(f"{rel_path}: {unexpected_matches}")
                    
            except (UnicodeDecodeError, OSError):
                continue
                
        return mismatches
    
    def create_version_update_script(self) -> str:
        """버전 업데이트용 스크립트 생성"""
        script_path = self.project_root / "scripts" / "update_version.py"
        script_path.parent.mkdir(exist_ok=True)
        
        script_content = f'''#!/usr/bin/env python3
"""
MoAI-ADK 버전 업데이트 자동화 스크립트
사용법: python scripts/update_version.py 0.2.0
"""

import sys
import re
from pathlib import Path

def update_version_in_file(file_path: Path, old_version: str, new_version: str) -> bool:
    """파일 내 버전 정보 업데이트"""
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()
            
        # 버전 패턴 교체
        patterns = [
            (r'__version__\\s*=\\s*"[^"]*"', f'__version__ = "0.1.17"'),
            (r'version\\s*=\\s*"[^"]*"', f'version = "{{new_version}}"'),
            (r'MoAI-ADK v[0-9]+\\.[0-9]+\\.[0-9]+', f'MoAI-ADK v{{new_version}}'),
            (r'"moai_version":\\s*"[^"]*"', f'"moai_version": "0.1.17"')
        ]
        
        original_content = content
        for pattern, replacement in patterns:
            content = re.sub(pattern, replacement.format(new_version=new_version), content)
            
        if content != original_content:
            with open(file_path, 'w', encoding='utf-8') as f:
                f.write(content)
            return True
            
    except Exception as e:
        print(f"Error updating {{file_path}}: {{e}}")
        
    return False

def main() -> None:
    if len(sys.argv) != 2:
        print("Usage: python scripts/update_version.py <new_version>")
        print("Example: python scripts/update_version.py 0.2.0")
        sys.exit(1)
        
    new_version = sys.argv[1]
    
    # 버전 형식 검증
    if not re.match(r'^[0-9]+\\.[0-9]+\\.[0-9]+$', new_version):
        print("Error: Version must be in format X.Y.Z")
        sys.exit(1)
        
    print(f"🗿 MoAI-ADK 버전 업데이트: v{{new_version}}")
    
    # 버전 동기화 실행
    from src.version_sync import VersionSyncManager
    
    # _version.py 먼저 업데이트
    version_file = Path("src/_version.py")
    update_version_in_file(version_file, None, new_version)
    
    # 전체 프로젝트 동기화
    sync_manager = VersionSyncManager()
    results = sync_manager.sync_all_versions()
    
    print(f"✅ 버전 업데이트 완료: v{{new_version}}")
    print("다음 단계:")
    print("1. git add .")  
    print("2. git commit -m 'bump version to v{new_version}'")
    print("3. git tag v{new_version}")
    print("4. git push origin main --tags")

if __name__ == "__main__":
    main()
'''
        
        with open(script_path, 'w', encoding='utf-8') as f:
            f.write(script_content)
            
        print(f"✅ 버전 업데이트 스크립트 생성: {script_path}")
        return str(script_path)


def main() -> None:
    """CLI entry point"""
    import argparse
    
    parser = argparse.ArgumentParser(description="MoAI-ADK 버전 동기화 도구")
    parser.add_argument("--dry-run", action="store_true", 
                       help="실제 변경하지 않고 시뮬레이션만 실행")
    parser.add_argument("--verify", action="store_true",
                       help="버전 동기화 검증만 실행")
    parser.add_argument("--create-script", action="store_true",
                       help="버전 업데이트 스크립트 생성")
    
    args = parser.parse_args()
    
    manager = VersionSyncManager()
    
    if args.verify:
        manager.verify_sync()
    elif args.create_script:
        manager.create_version_update_script()
    else:
        manager.sync_all_versions(dry_run=args.dry_run)
        if not args.dry_run:
            manager.verify_sync()


if __name__ == "__main__":
    main()
