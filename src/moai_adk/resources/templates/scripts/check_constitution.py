#!/usr/bin/env python3
"""
MoAI-ADK Constitution 5원칙 검증 스크립트 v0.1.12

Constitution 5원칙 준수 여부를 자동으로 검증합니다:
1. Simplicity: 동시 활성 프로젝트 ≤ 3개
2. Architecture: 모든 기능은 라이브러리로 구현
3. Testing: TDD RED-GREEN-REFACTOR 강제
4. Observability: 구조화된 로깅 시스템 필수
5. Versioning: MAJOR.MINOR.BUILD 체계 준수

사용법:
    python scripts/check_constitution.py [--fix] [--verbose]
    
옵션:
    --fix         자동 수정 가능한 위반사항 수정
    --verbose     상세한 분석 결과 출력
    --report      HTML 보고서 생성
"""

import json
import os
import re
import sys
import subprocess
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Optional, Any, Tuple
import argparse

class ConstitutionChecker:
    """Constitution 5원칙 준수 여부 검증 클래스"""
    
    def __init__(self, project_root: Path, verbose: bool = False):
        self.project_root = project_root
        self.verbose = verbose
        self.moai_dir = project_root / ".moai"
        self.results = {
            'simplicity': {'passed': False, 'score': 0, 'issues': []},
            'architecture': {'passed': False, 'score': 0, 'issues': []},
            'testing': {'passed': False, 'score': 0, 'issues': []},
            'observability': {'passed': False, 'score': 0, 'issues': []},
            'versioning': {'passed': False, 'score': 0, 'issues': []},
            'overall': {'passed': False, 'score': 0, 'compliance_level': ''}
        }
        
        # 체크 기준 설정
        self.thresholds = {
            'max_projects': 3,
            'min_test_coverage': 80,
            'required_patterns': {
                'logging': ['logging.getLogger', 'log.info', 'log.error', 'log.debug'],
                'versioning': ['__version__', 'VERSION', 'version'],
                'testing': ['test_', 'Test', 'pytest', 'unittest']
            }
        }

    def check_simplicity_principle(self) -> Dict[str, Any]:
        """원칙 1: Simplicity - 동시 활성 프로젝트 수 제한"""
        
        active_projects = []
        issues = []
        
        # .moai/specs 디렉토리에서 활성 프로젝트 찾기
        if self.moai_dir.exists():
            specs_dir = self.moai_dir / "specs"
            if specs_dir.exists():
                for spec_dir in specs_dir.glob("SPEC-*"):
                    if spec_dir.is_dir():
                        spec_file = spec_dir / "spec.md"
                        if spec_file.exists():
                            content = spec_file.read_text(encoding='utf-8')
                            # 활성 상태 확인 (NEEDS CLARIFICATION이나 완료 상태가 아닌 경우)
                            if not re.search(r'\[COMPLETED\]|\[CANCELLED\]', content, re.IGNORECASE):
                                active_projects.append(spec_dir.name)
        
        # 현재 작업 중인 branch 확인
        try:
            result = subprocess.run(['git', 'branch', '--show-current'], 
                                  capture_output=True, text=True, cwd=self.project_root)
            current_branch = result.stdout.strip()
            if current_branch != 'main' and current_branch != 'master':
                active_projects.append(f"branch:{current_branch}")
        except:
            pass
        
        project_count = len(active_projects)
        
        if project_count > self.thresholds['max_projects']:
            issues.append(f"활성 프로젝트 {project_count}개 > 최대 허용 {self.thresholds['max_projects']}개")
            issues.extend([f"  - {project}" for project in active_projects])
        
        if project_count == 0:
            issues.append("활성 프로젝트가 없음 (최소 1개 필요)")
        
        # 복잡도 스코어 계산 (0-100)
        complexity_score = min(100, max(0, 100 - (project_count - 1) * 20))
        
        return {
            'passed': project_count <= self.thresholds['max_projects'] and project_count > 0,
            'score': complexity_score,
            'issues': issues,
            'details': {
                'active_projects': active_projects,
                'project_count': project_count,
                'max_allowed': self.thresholds['max_projects']
            }
        }

    def check_architecture_principle(self) -> Dict[str, Any]:
        """원칙 2: Architecture - 모든 기능은 라이브러리로 구현"""
        
        issues = []
        library_patterns = []
        monolith_patterns = []
        
        # Python 파일들 스캔
        for py_file in self.project_root.rglob("*.py"):
            if 'test' in str(py_file) or '__pycache__' in str(py_file):
                continue
                
            try:
                content = py_file.read_text(encoding='utf-8')
                
                # 라이브러리 패턴 감지
                if re.search(r'class\s+\w+\(.*\):|def\s+\w+\(.*\).*->.*:|from\s+\.\w+\s+import', content):
                    library_patterns.append(str(py_file.relative_to(self.project_root)))
                
                # 모놀리스 패턴 감지
                if re.search(r'if\s+__name__\s*==\s*["\']__main__["\'].*\n.*\.run\(\)|app\.run\(|main\(\)', content, re.DOTALL):
                    monolith_patterns.append(str(py_file.relative_to(self.project_root)))
                    
            except Exception as e:
                if self.verbose:
                    print(f"파일 읽기 오류 {py_file}: {e}")
        
        # package.json 또는 pyproject.toml 체크 (의존성 관리)
        has_dependency_management = False
        dep_files = ['package.json', 'pyproject.toml', 'requirements.txt', 'setup.py']
        for dep_file in dep_files:
            if (self.project_root / dep_file).exists():
                has_dependency_management = True
                break

        # claude-code-manager 에이전트 존재 확인
        has_claude_code_manager = False
        claude_agents_dir = self.project_root / ".claude" / "agents" / "moai"
        if claude_agents_dir.exists():
            claude_code_manager_file = claude_agents_dir / "claude-code-manager.md"
            if claude_code_manager_file.exists():
                has_claude_code_manager = True
        
        if not has_dependency_management:
            issues.append("의존성 관리 파일 없음 (package.json, pyproject.toml, requirements.txt 중 하나 필요)")
        
        if len(library_patterns) == 0:
            issues.append("라이브러리 구조가 감지되지 않음")
        
        if len(monolith_patterns) > 2:
            issues.append(f"모놀리스 패턴 감지: {len(monolith_patterns)}개 파일")
            issues.extend([f"  - {pattern}" for pattern in monolith_patterns[:5]])

        if not has_claude_code_manager:
            issues.append("claude-code-manager 에이전트가 없음 (Claude Code 통합 관리 필수)")

        # 아키텍처 스코어 계산
        arch_score = 0
        if has_dependency_management:
            arch_score += 30
        if len(library_patterns) > 0:
            arch_score += 25
        if len(monolith_patterns) <= 2:
            arch_score += 25
        if has_claude_code_manager:
            arch_score += 20
        
        return {
            'passed': len(issues) == 0,
            'score': arch_score,
            'issues': issues,
            'details': {
                'library_patterns': len(library_patterns),
                'monolith_patterns': len(monolith_patterns),
                'has_dependency_management': has_dependency_management,
                'has_claude_code_manager': has_claude_code_manager
            }
        }

    def check_testing_principle(self) -> Dict[str, Any]:
        """원칙 3: Testing - TDD RED-GREEN-REFACTOR 강제"""
        
        issues = []
        test_files = []
        test_coverage = 0
        
        # 테스트 파일 찾기
        test_patterns = ['test_*.py', '*_test.py', 'tests.py']
        for pattern in test_patterns:
            test_files.extend(list(self.project_root.rglob(pattern)))
        
        # tests/ 디렉토리 체크
        tests_dir = self.project_root / "tests"
        if tests_dir.exists():
            test_files.extend(list(tests_dir.rglob("*.py")))
        
        # 중복 제거
        test_files = list(set(test_files))
        
        if len(test_files) == 0:
            issues.append("테스트 파일이 없음")
        
        # 테스트 커버리지 확인 (pytest-cov 사용)
        try:
            result = subprocess.run([
                'python', '-m', 'pytest', '--cov=src', '--cov=.', 
                '--cov-report=term-missing', '--tb=no', '-q'
            ], capture_output=True, text=True, cwd=self.project_root, timeout=60)
            
            if result.returncode == 0:
                # 커버리지 퍼센트 추출
                coverage_match = re.search(r'TOTAL.*?(\d+)%', result.stdout)
                if coverage_match:
                    test_coverage = int(coverage_match.group(1))
            else:
                # 테스트 실행 실패
                issues.append("테스트 실행 실패")
                
        except subprocess.TimeoutExpired:
            issues.append("테스트 실행 시간 초과")
        except Exception as e:
            if self.verbose:
                print(f"테스트 실행 오류: {e}")
        
        if test_coverage < self.thresholds['min_test_coverage']:
            issues.append(f"테스트 커버리지 {test_coverage}% < 최소 요구 {self.thresholds['min_test_coverage']}%")
        
        # TDD 패턴 감지 (Red-Green-Refactor)
        tdd_indicators = 0
        for test_file in test_files[:10]:  # 처음 10개 파일만 체크
            try:
                content = test_file.read_text(encoding='utf-8')
                # TDD 패턴 키워드 찾기
                if re.search(r'def test_.*fail|assert.*False|pytest\.raises', content):
                    tdd_indicators += 1
            except:
                pass
        
        if len(test_files) > 0 and tdd_indicators == 0:
            issues.append("TDD 패턴 (실패 테스트)이 감지되지 않음")
        
        # 테스트 스코어 계산
        test_score = 0
        if len(test_files) > 0:
            test_score += 30
        if test_coverage >= self.thresholds['min_test_coverage']:
            test_score += 50
        if tdd_indicators > 0:
            test_score += 20
        
        return {
            'passed': len(issues) == 0,
            'score': test_score,
            'issues': issues,
            'details': {
                'test_files_count': len(test_files),
                'test_coverage': test_coverage,
                'tdd_indicators': tdd_indicators,
                'required_coverage': self.thresholds['min_test_coverage']
            }
        }

    def check_observability_principle(self) -> Dict[str, Any]:
        """원칙 4: Observability - 구조화된 로깅 시스템 필수"""
        
        issues = []
        logging_files = []
        structured_logging = False
        
        # Python 파일에서 로깅 패턴 찾기
        for py_file in self.project_root.rglob("*.py"):
            if '__pycache__' in str(py_file):
                continue
                
            try:
                content = py_file.read_text(encoding='utf-8')
                
                # 로깅 import 확인
                if re.search(r'import logging|from logging', content):
                    logging_files.append(str(py_file.relative_to(self.project_root)))
                    
                    # 구조화된 로깅 확인
                    if re.search(r'logging\.basicConfig|logging\.getLogger.*\.info|log\.info.*{', content):
                        structured_logging = True
                        
            except:
                pass
        
        if len(logging_files) == 0:
            issues.append("로깅 시스템이 구현되지 않음")
        
        if not structured_logging:
            issues.append("구조화된 로깅 패턴이 없음")
        
        # 로깅 설정 파일 확인
        log_config_files = ['logging.json', 'logging.yaml', 'log_config.py']
        has_log_config = any((self.project_root / f).exists() for f in log_config_files)
        
        if len(logging_files) > 0 and not has_log_config:
            issues.append("로깅 설정 파일이 없음")
        
        # 관측성 스코어 계산
        obs_score = 0
        if len(logging_files) > 0:
            obs_score += 40
        if structured_logging:
            obs_score += 40
        if has_log_config:
            obs_score += 20
        
        return {
            'passed': len(issues) == 0,
            'score': obs_score,
            'issues': issues,
            'details': {
                'logging_files_count': len(logging_files),
                'structured_logging': structured_logging,
                'has_log_config': has_log_config
            }
        }

    def check_versioning_principle(self) -> Dict[str, Any]:
        """원칙 5: Versioning - MAJOR.MINOR.BUILD 체계 준수"""
        
        issues = []
        version_files = []
        valid_versions = []
        
        # 버전 정보가 있는 파일들 찾기
        version_patterns = [
            ('pyproject.toml', r'version\s*=\s*["\']([^"\']+)["\']'),
            ('setup.py', r'version\s*=\s*["\']([^"\']+)["\']'),
            ('package.json', r'"version"\s*:\s*"([^"]+)"'),
            ('__init__.py', r'__version__\s*=\s*["\']([^"\']+)["\']'),
            ('_version.py', r'__version__\s*=\s*["\']([^"\']+)["\']'),
            ('version.py', r'VERSION\s*=\s*["\']([^"\']+)["\']')
        ]
        
        for filename, pattern in version_patterns:
            for version_file in self.project_root.rglob(filename):
                try:
                    content = version_file.read_text(encoding='utf-8')
                    matches = re.findall(pattern, content)
                    for version in matches:
                        version_files.append(str(version_file.relative_to(self.project_root)))
                        
                        # MAJOR.MINOR.BUILD 형식 검증
                        if re.match(r'^\d+\.\d+\.\d+(-\w+(\.\d+)?)?$', version):
                            valid_versions.append((str(version_file.relative_to(self.project_root)), version))
                        else:
                            issues.append(f"잘못된 버전 형식 {version} in {version_file.name}")
                            
                except:
                    pass
        
        if len(version_files) == 0:
            issues.append("버전 정보가 없음")
        
        # 버전 일관성 확인
        if len(valid_versions) > 1:
            versions = [v[1] for v in valid_versions]
            if len(set(versions)) > 1:
                issues.append(f"버전 불일치: {set(versions)}")
        
        # Git 태그 버전 확인
        git_tags = []
        try:
            result = subprocess.run(['git', 'tag', '-l'], 
                                  capture_output=True, text=True, cwd=self.project_root)
            if result.returncode == 0:
                git_tags = [tag.strip() for tag in result.stdout.split('\n') if tag.strip()]
                valid_git_tags = [tag for tag in git_tags if re.match(r'^v?\d+\.\d+\.\d+', tag)]
                
                if len(git_tags) > 0 and len(valid_git_tags) == 0:
                    issues.append("Git 태그가 버전 체계를 따르지 않음")
        except:
            pass
        
        # 버전 스코어 계산
        version_score = 0
        if len(version_files) > 0:
            version_score += 30
        if len(valid_versions) > 0:
            version_score += 40
        if len(set([v[1] for v in valid_versions])) <= 1:  # 일관성
            version_score += 30
        
        return {
            'passed': len(issues) == 0,
            'score': version_score,
            'issues': issues,
            'details': {
                'version_files': version_files,
                'valid_versions': valid_versions,
                'git_tags': git_tags
            }
        }

    def run_full_check(self) -> Dict[str, Any]:
        """전체 Constitution 검증 실행"""
        
        print("🏛️ Constitution 5원칙 검증 시작...")
        
        # 각 원칙별 검증
        self.results['simplicity'] = self.check_simplicity_principle()
        self.results['architecture'] = self.check_architecture_principle()
        self.results['testing'] = self.check_testing_principle()
        self.results['observability'] = self.check_observability_principle()
        self.results['versioning'] = self.check_versioning_principle()
        
        # 전체 점수 계산
        total_score = sum([
            self.results[principle]['score'] 
            for principle in ['simplicity', 'architecture', 'testing', 'observability', 'versioning']
        ])
        overall_score = total_score / 5
        
        # 준수 등급 결정
        if overall_score >= 90:
            compliance_level = "EXCELLENT"
        elif overall_score >= 75:
            compliance_level = "GOOD"
        elif overall_score >= 60:
            compliance_level = "ADEQUATE"
        else:
            compliance_level = "NEEDS_IMPROVEMENT"
        
        # 전체 통과 여부
        all_passed = all([
            self.results[principle]['passed']
            for principle in ['simplicity', 'architecture', 'testing', 'observability', 'versioning']
        ])
        
        self.results['overall'] = {
            'passed': all_passed,
            'score': overall_score,
            'compliance_level': compliance_level,
            'total_issues': sum([len(self.results[p]['issues']) for p in ['simplicity', 'architecture', 'testing', 'observability', 'versioning']])
        }
        
        return self.results

    def generate_report(self, output_path: Optional[Path] = None) -> str:
        """검증 결과 리포트 생성"""
        
        report_lines = [
            "🏛️ MoAI-ADK Constitution 5원칙 검증 결과",
            "=" * 50,
            f"검증 일시: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}",
            f"프로젝트: {self.project_root.name}",
            f"전체 점수: {self.results['overall']['score']:.1f}/100",
            f"준수 등급: {self.results['overall']['compliance_level']}",
            f"전체 통과: {'✅ PASS' if self.results['overall']['passed'] else '❌ FAIL'}",
            ""
        ]
        
        # 원칙별 상세 결과
        principles = {
            'simplicity': '1. Simplicity (단순성)',
            'architecture': '2. Architecture (아키텍처)',
            'testing': '3. Testing (테스팅)',
            'observability': '4. Observability (관측성)',
            'versioning': '5. Versioning (버전관리)'
        }
        
        for key, name in principles.items():
            result = self.results[key]
            status = "✅ PASS" if result['passed'] else "❌ FAIL"
            
            report_lines.extend([
                f"## {name}",
                f"상태: {status} ({result['score']}/100)",
                ""
            ])
            
            if result['issues']:
                report_lines.extend([
                    "### 이슈:",
                ] + [f"  - {issue}" for issue in result['issues']] + [""])
            
            if self.verbose and 'details' in result:
                report_lines.extend([
                    "### 세부 정보:",
                    json.dumps(result['details'], indent=2, ensure_ascii=False),
                    ""
                ])
        
        # 권장사항
        report_lines.extend([
            "## 권장사항",
            ""
        ])
        
        if self.results['overall']['score'] < 75:
            report_lines.append("- 전체적인 Constitution 준수도 개선 필요")
        
        if not self.results['testing']['passed']:
            report_lines.append("- TDD 도입 및 테스트 커버리지 향상 우선 수행")
        
        if not self.results['observability']['passed']:
            report_lines.append("- 구조화된 로깅 시스템 도입")
        
        report_content = "\n".join(report_lines)
        
        # 파일 저장
        if output_path:
            output_path.write_text(report_content, encoding='utf-8')
            print(f"📄 보고서 저장: {output_path}")
        
        return report_content

def main():
    """스크립트 진입점"""
    
    parser = argparse.ArgumentParser(description="Constitution 5원칙 준수 검증")
    parser.add_argument("--fix", action="store_true", help="자동 수정 실행")
    parser.add_argument("--verbose", "-v", action="store_true", help="상세 출력")
    parser.add_argument("--report", help="리포트 출력 파일 경로")
    parser.add_argument("--project-dir", help="프로젝트 루트 디렉토리")
    
    args = parser.parse_args()
    
    # 프로젝트 루트 설정
    if args.project_dir:
        project_root = Path(args.project_dir)
    else:
        project_root = Path(os.getcwd())
    
    if not project_root.exists():
        print(f"❌ 프로젝트 디렉토리가 존재하지 않음: {project_root}")
        sys.exit(1)
    
    # Constitution 검증 실행
    checker = ConstitutionChecker(project_root, verbose=args.verbose)
    results = checker.run_full_check()
    
    # 결과 출력
    report = checker.generate_report(
        Path(args.report) if args.report else None
    )
    
    if not args.report:
        print(report)
    
    # 종료 코드 설정
    if results['overall']['passed']:
        print("\n🎉 모든 Constitution 원칙을 준수합니다!")
        sys.exit(0)
    else:
        print(f"\n⚠️  {results['overall']['total_issues']}개 이슈 해결 필요")
        sys.exit(1)

if __name__ == "__main__":
    main()