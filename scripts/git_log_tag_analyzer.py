#!/usr/bin/env python3
"""
Git 로그 기반 TAG 연결 분석기
커밋 이력을 통해 실제 구현된 기능과 SPEC의 연결 관계를 추적
"""

import re
import subprocess
from pathlib import Path
from typing import Dict, List, Set, Tuple
from collections import defaultdict

class GitLogTagAnalyzer:
    def __init__(self):
        self.TAG_PATTERN = re.compile(r'@(SPEC|CODE|TEST|DOC):([A-Z0-9-]+-\d{3})')

        # Git 로그에서 찾은 실제 구현 관계
        self.implementations: Dict[str, str] = {}  # spec_domain -> code_file
        self.spec_implementations: Dict[str, List[str]] = defaultdict(list)  # spec -> [files]
        self.code_implementations: Dict[str, List[str]] = defaultdict(list)  # code_file -> [specs]

        # 커밋 메시지에서 추출한 TAG 정보
        self.commits_with_specs = []
        self.commits_with_features = []

    def run_git_command(self, command: str) -> str:
        """Git 명령어 실행"""
        try:
            result = subprocess.run(
                command,
                shell=True,
                capture_output=True,
                text=True,
                cwd=Path.cwd()
            )
            return result.stdout.strip()
        except Exception as e:
            print(f"Git command failed: {e}")
            return ""

    def analyze_commit_history(self) -> None:
        """Git 커밋 히스토리 분석"""
        print("🔍 Git 커밋 히스토리 분석 중...")

        # 1. SPEC 관련 커밋 찾기
        spec_commits = self.run_git_command(
            "git log --grep='SPEC' --oneline -50"
        )

        # 2. feat 커밋 찾기
        feat_commits = self.run_git_command(
            "git log --grep='feat' --oneline -50"
        )

        # 3. 최근 100개 커밋에서 파일 변경 내역 분석
        recent_commits = self.run_git_command(
            "git log --name-only --oneline -100"
        )

        print(f"   SPEC 관련 커밋: {len(spec_commits.splitlines())}개")
        print(f"   feat 커밋: {len(feat_commits.splitlines())}개")
        print(f"   최근 커밋: {len(recent_commits.splitlines())}개")

        self._parse_commit_messages(spec_commits, feat_commits)
        self._analyze_file_changes(recent_commits)

    def _parse_commit_messages(self, spec_commits: str, feat_commits: str) -> None:
        """커밋 메시지에서 SPEC/feature 정보 추출"""
        print("📝 커밋 메시지 파싱 중...")

        all_commits = (spec_commits + "\n" + feat_commits).splitlines()

        for commit_line in all_commits:
            if not commit_line.strip():
                continue

            # 태그 패턴 찾기
            tags = self.TAG_PATTERN.findall(commit_line)
            if tags:
                for tag_type, domain in tags:
                    if tag_type == "SPEC":
                        self.commits_with_specs.append((commit_line, domain))
                    elif tag_type == "CODE":
                        self.commits_with_features.append((commit_line, domain))

        print(f"   발견된 SPEC 커밋: {len(self.commits_with_specs)}개")
        print(f"   발견된 CODE 커밋: {len(self.commits_with_features)}개")

    def _analyze_file_changes(self, recent_commits: str) -> None:
        """최근 커밋의 파일 변경 내역 분석"""
        print("📁 파일 변경 내역 분석 중...")

        # 커밋별로 파일 변경 내역 파싱
        commit_blocks = recent_commits.split('\n\n')

        for block in commit_blocks:
            if not block.strip():
                continue

            lines = block.split('\n')
            if len(lines) < 2:
                continue

            commit_line = lines[0]
            changed_files = [line for line in lines[1:] if line.strip() and not line.startswith(' ')]

            # 변경된 파일에서 src/ 파일만 필터링
            src_files = [f for f in changed_files if f.startswith('src/')]

            if src_files:
                self._map_commit_to_implementation(commit_line, src_files)

        print(f"   분석된 구현 관계: {len(self.implementations)}개")

    def _map_commit_to_implementation(self, commit_line: str, src_files: List[str]) -> None:
        """커밋을 실제 구현된 파일과 매핑"""

        # 커밋 메시지에서 SPEC 도메인 추출
        tags = self.TAG_PATTERN.findall(commit_line)
        for tag_type, domain in tags:
            if tag_type == "SPEC":
                for src_file in src_files:
                    self.implementations[domain] = src_file
                    self.spec_implementations[domain].append(src_file)
                    self.code_implementations[src_file].append(domain)

    def find_missing_connections(self) -> Dict[str, List[str]]:
        """빠진 연결 관계 찾기"""
        print("🔗 빠진 연결 관계 분석 중...")

        # 실제 src/ 파일들에서 @CODE 태그 추출
        existing_code_tags = self._extract_code_tags_from_source()

        # Git 로그에서 찾은 구현 vs 현재 코드 태그 비교
        missing_connections = {
            "implemented_but_not_tagged": [],
            "tagged_but_not_in_git": []
        }

        for domain, src_file in self.implementations.items():
            if src_file in existing_code_tags:
                if domain not in existing_code_tags[src_file]:
                    missing_connections["implemented_but_not_tagged"].append((domain, src_file))
            else:
                missing_connections["tagged_but_not_in_git"].append((domain, src_file))

        print(f"   구현됐지만 태그 없음: {len(missing_connections['implemented_but_not_tagged'])}개")
        print(f"   태그됐지만 구현 없음: {len(missing_connections['tagged_but_not_in_git'])}개")

        return missing_connections

    def _extract_code_tags_from_source(self) -> Dict[str, Set[str]]:
        """소스 코드에서 @CODE 태그 추출"""
        code_tags = defaultdict(set)

        src_dir = Path("src")
        if not src_dir.exists():
            return code_tags

        for py_file in src_dir.rglob("*.py"):
            try:
                content = py_file.read_text(encoding='utf-8')
                tags = self.TAG_PATTERN.findall(content)
                for tag_type, domain in tags:
                    if tag_type == "CODE":
                        code_tags[str(py_file)].add(domain)
            except Exception as e:
                print(f"   파일 읽기 오류 {py_file}: {e}")

        return code_tags

    def generate_improved_mappings(self) -> Dict[str, str]:
        """개선된 매핑 제안 생성"""
        print("💡 개선된 매핑 제안 생성 중...")

        improved_mappings = {}

        # Git 로그에서 찾은 실제 구현 관계 기반 매핑
        for domain, src_file in self.implementations.items():
            # 파일 이름에서 CODE 도메인 추출 시도
            potential_code_domain = self._extract_domain_from_filename(src_file)
            if potential_code_domain:
                improved_mappings[f"@CODE:{potential_code_domain}"] = f"@SPEC:{domain}"

        print(f"   제안된 매핑: {len(improved_mappings)}개")
        return improved_mappings

    def _extract_domain_from_filename(self, filepath: str) -> str:
        """파일 경로에서 도메인 추출"""
        filename = Path(filepath).stem

        # 일반적인 패턴들
        patterns = {
            'tag': 'TAG',
            'git': 'GIT',
            'hook': 'HOOK',
            'template': 'TEMPLATE',
            'version': 'VERSION',
            'config': 'CONFIG',
            'project': 'PROJECT',
            'cache': 'CACHE',
            'lang': 'LANG',
            'update': 'UPDATE',
            'init': 'INIT',
            'rollback': 'ROLLBACK',
            'backup': 'BACKUP',
            'network': 'NETWORK',
            'timeout': 'TIMEOUT',
            'policy': 'POLICY',
            'util': 'UTILS',
            'map': 'MAP',
            'val': 'VAL'
        }

        for pattern, domain in patterns.items():
            if pattern in filename.lower():
                # 숫자 접미사 찾기
                numbers = re.findall(r'\d+', filename)
                if numbers:
                    return f"{domain}-{numbers[0]:03d}"
                else:
                    return f"{domain}-001"

        return ""

    def run_analysis(self) -> Dict:
        """전체 분석 실행"""
        print("=" * 60)
        print("🔍 Git 로그 기반 TAG 연결 분석")
        print("=" * 60)

        self.analyze_commit_history()
        missing_connections = self.find_missing_connections()
        improved_mappings = self.generate_improved_mappings()

        # 결과 요약
        print("\n📊 Git 로그 분석 결과 요약")
        print("-" * 40)
        print(f"🔍 발견된 SPEC-CODE 연결: {len(self.implementations)}개")
        print(f"📝 구현됐지만 태그 없음: {len(missing_connections['implemented_but_not_tagged'])}개")
        print(f"🏷️  태그됐지만 구현 없음: {len(missing_connections['tagged_but_not_in_git'])}개")
        print(f"💡 제안된 개선 매핑: {len(improved_mappings)}개")

        return {
            "implementations": self.implementations,
            "missing_connections": missing_connections,
            "improved_mappings": improved_mappings,
            "commits_with_specs": self.commits_with_specs,
            "commits_with_features": self.commits_with_features
        }

if __name__ == "__main__":
    analyzer = GitLogTagAnalyzer()
    results = analyzer.run_analysis()

    print(f"\n🎯 추천 조치:")
    print("1. 구현됐지만 태그 없는 파일에 @CODE 태그 추가")
    print("2. Git 로그에서 발견된 실제 구현 관계로 TAG 매핑 개선")
    print("3. 불필요한 태그 정리 및 중복 제거")