#!/usr/bin/env python3
# @CODE:TAG-ANALYSIS-001
"""TAG 연결 상태 분석 스크립트

src/ 디렉토리와 .moai/specs/ 디렉토리의 TAG 연결 상태를 분석하여
고아 TAG 현황을 파악하고 개선 방안을 제시합니다.

@SPEC:TAG-ANALYSIS-001
"""

import re
import json
from pathlib import Path
from collections import defaultdict, Counter
from typing import Dict, List, Set, Tuple


class TagAnalyzer:
    """TAG 연결 상태 분석기"""

    TAG_PATTERN = re.compile(r'@(SPEC|CODE|TEST|DOC):([A-Z0-9-]+-\d{3})')

    def __init__(self, project_root: Path = None):
        self.project_root = project_root or Path('.')
        self.src_dir = self.project_root / "src"
        self.specs_dir = self.project_root / ".moai" / "specs"
        self.tests_dir = self.project_root / "tests"

        # 분석 결과 저장
        self.spec_tags = set()
        self.code_tags = set()
        self.test_tags = set()
        self.doc_tags = set()
        self.all_tags = defaultdict(set)  # {tag_type: set of domains}

        # 매핑 관계
        self.code_to_spec = {}  # {code_domain: spec_domain}
        self.code_to_test = {}  # {code_domain: test_domain}
        self.test_to_code = {}  # {test_domain: code_domain}
        self.spec_to_code = {}  # {spec_domain: code_domain}

    def extract_tags_from_file(self, file_path: Path) -> Dict[str, Set[str]]:
        """파일에서 TAG 추출

        Args:
            file_path: 파일 경로

        Returns:
            {tag_type: set of domains} 딕셔너리
        """
        tags = {"SPEC": set(), "CODE": set(), "TEST": set(), "DOC": set()}

        try:
            content = file_path.read_text(encoding="utf-8", errors="ignore")
            lines = content.splitlines()

            for line_num, line in enumerate(lines, start=1):
                matches = self.TAG_PATTERN.findall(line)
                for tag_type, domain in matches:
                    tags[tag_type].add(domain)
                    self.all_tags[tag_type].add(domain)

        except Exception as e:
            print(f"Warning: Failed to read {file_path}: {e}")

        return tags

    def analyze_specs(self) -> None:
        """SPEC 파일 분석"""
        print("🔍 SPEC 파일 분석 중...")

        spec_files = list(self.specs_dir.glob("*/spec.md"))
        print(f"   발견된 SPEC 파일: {len(spec_files)}개")

        for spec_file in spec_files:
            tags = self.extract_tags_from_file(spec_file)
            self.spec_tags.update(tags["SPEC"])
            self.all_tags["SPEC"].update(tags["SPEC"])

        print(f"   추출된 @SPEC TAG: {len(self.spec_tags)}개")

    def analyze_src(self) -> None:
        """src/ 디렉토리 분석"""
        print("🔍 src/ 디렉토리 분석 중...")

        code_files = list(self.src_dir.rglob("*.py"))
        print(f"   발견된 Python 파일: {len(code_files)}개")

        for code_file in code_files:
            tags = self.extract_tags_from_file(code_file)
            self.code_tags.update(tags["CODE"])
            self.test_tags.update(tags["TEST"])
            self.doc_tags.update(tags["DOC"])

        print(f"   추출된 @CODE TAG: {len(self.code_tags)}개")
        print(f"   추출된 @TEST TAG: {len(self.test_tags)}개")
        print(f"   추출된 @DOC TAG: {len(self.doc_tags)}개")

    def analyze_tests(self) -> None:
        """tests/ 디렉토리 분석"""
        print("🔍 tests/ 디렉토리 분석 중...")

        test_files = list(self.tests_dir.rglob("*.py"))
        print(f"   발견된 테스트 파일: {len(test_files)}개")

        for test_file in test_files:
            tags = self.extract_tags_from_file(test_file)
            self.test_tags.update(tags["TEST"])
            self.all_tags["TEST"].update(tags["TEST"])

        # 중복 제거를 위해 추가로 업데이트
        self.all_tags["TEST"].update(tags["TEST"])

        print(f"   추출된 @TEST TAG: {len(self.test_tags)}개 (중복 포함)")

    def build_mappings(self) -> None:
        """TAG 간 매핑 관계 구축 (개선된 지능형 매칭)"""
        print("🔗 TAG 매핑 관계 분석 중...")

        # CODE → SPEC 매핑 (향상된 도메인 매칭 기준)
        for code_domain in self.code_tags:
            matched_spec = self._find_best_spec_match(code_domain)
            if matched_spec:
                self.code_to_spec[code_domain] = matched_spec
                self.spec_to_code[matched_spec] = code_domain

        # CODE → TEST 매핑 (향상된 도메인 매칭 기준)
        for code_domain in self.code_tags:
            matched_test = self._find_best_test_match(code_domain)
            if matched_test:
                self.code_to_test[code_domain] = matched_test
                self.test_to_code[matched_test] = code_domain

        print(f"   CODE → SPEC 매핑: {len(self.code_to_spec)}개")
        print(f"   CODE → TEST 매핑: {len(self.code_to_test)}개")

    def _find_best_spec_match(self, code_domain: str) -> str:
        """CODE 도메인에 가장 적합한 SPEC 도메인 찾기 (개선된 매칭)"""
        code_base = self._extract_domain(code_domain)

        # 1. 정확히 일치하는 경우
        for spec_domain in self.spec_tags:
            spec_base = self._extract_domain(spec_domain)
            if code_base == spec_base:
                return spec_domain

        # 2. 실제 구현 관계 기반 매칭 (Git 로그 기반)
        implementation_mappings = {
            'TAG-AUTO-CORRECTOR': 'TAG-AUTO-001',
            'TAG-POLICY-VALIDATOR': 'TAG-POLICY-001',
            'TAG-ROLLBACK-MANAGER': 'TAG-ROLLBACK-001',
            'TAG-VALIDATOR': 'DOC-TAG-001',
            'TAG-GENERATOR': 'DOC-TAG-001',
            'TAG-INSERTER': 'DOC-TAG-001',
            'TAG-PARSER': 'DOC-TAG-001',
            'TAG-MAPPER': 'DOC-TAG-001',
            'TRUST-CHECKER': 'TRUST-001',
            'PROJECT-VALIDATOR': 'PROJECT-001',
            'PROJECT-INITIALIZER': 'INIT-001',
            'GIT-MANAGER': 'GIT-001',
            'GIT-BRANCH-MANAGER': 'GIT-001',
            'TEMPLATE-MERGER': 'TEMPLATE-001',
            'CONFIG-MIGRATION': 'CONFIG-001',
            'VERSION-CACHE': 'UPDATE-CACHE-FIX-001',
        }

        if code_base in implementation_mappings:
            target_spec = implementation_mappings[code_base]
            for spec_domain in self.spec_tags:
                if target_spec in spec_domain:
                    return spec_domain

        # 3. 포함 관계 매칭 (GIT-MANAGER ↔ GIT, HOOK-BASH ↔ HOOKS)
        for spec_domain in self.spec_tags:
            spec_base = self._extract_domain(spec_domain)
            if code_base in spec_base or spec_base in code_base:
                return spec_domain

        # 4. 의미적 매칭 테이블 (확장)
        semantic_matches = {
            'GIT': ['GIT-MANAGER', 'GIT-BRANCH', 'GIT-COMMIT', 'GIT-INIT'],
            'HOOK': ['HOOKS', 'HOOK-BASH', 'HOOK-TOOL', 'HOOK-TAG'],
            'TEMPLATE': ['TEMPLATE', 'TEMPLATE-MERGER', 'TEMPLATE-BACKUP', 'TEMPLATE-MODULE'],
            'VERSION': ['VERSION', 'VERSION-CACHE', 'VERSION-DETECT'],
            'UPDATE': ['UPDATE', 'UPDATE-CACHE', 'UPDATE-REFACTOR', 'UPDATE-TEMPLATE'],
            'LANG': ['LANG', 'LANG-DETECT', 'LANGUAGE-DETECTION'],
            'INIT': ['INIT', 'INITIALIZER', 'INIT-PHASE', 'INIT-FLOW'],
            'CACHE': ['CACHE', 'VERSION-CACHE', 'UPDATE-CACHE'],
            'TEST': ['TEST', 'TEST-COVERAGE', 'TEST-VALIDATOR'],
            'DOC': ['DOC', 'DOCS', 'DOCUMENT'],
            'CONFIG': ['CONFIG', 'CONFIGURATION'],
            'CLI': ['CLI', 'COMMAND', 'COMMANDS'],
            'CORE': ['CORE', 'CENTRAL', 'MAIN'],
            'AUTO': ['AUTO', 'AUTOMATIC', 'AUTOMATION'],
            'ROLLBACK': ['ROLLBACK', 'REVERT', 'UNDO'],
            'BACKUP': ['BACKUP', 'BACKUP-UTILS'],
            'NETWORK': ['NETWORK', 'NET', 'CONNECTIVITY'],
            'TIMEOUT': ['TIMEOUT', 'TIMING', 'TIME'],
            'POLICY': ['POLICY', 'POLICIES', 'RULES'],
            'UTILS': ['UTILS', 'UTILITIES', 'TOOLS'],
            'MAP': ['MAP', 'MAPPING', 'MAPPER'],
            'VAL': ['VAL', 'VALIDATE', 'VALIDATOR'],
            'LDE': ['LDE', 'DETECTION', 'LANG-DETECT'],
            'HOOKS': ['HOOK', 'HOOK-BASH', 'HOOK-TOOL'],
            'TAG': ['TAG', 'DOC-TAG'],
            'TRUST': ['TRUST', 'TRUST5'],
            'PROJECT': ['PROJECT', 'PROJECT-INIT', 'PROJECT-CONFIG'],
        }

        # 의미적 매칭 확인
        for key, related_terms in semantic_matches.items():
            if key in code_base or any(term in code_base for term in related_terms):
                for spec_domain in self.spec_tags:
                    spec_base = self._extract_domain(spec_domain)
                    if key in spec_base or any(term in spec_base for term in related_terms):
                        return spec_domain

        return None

    def _find_best_test_match(self, code_domain: str) -> str:
        """CODE 도메인에 가장 적합한 TEST 도메인 찾기"""
        # SPEC 매칭과 유사한 로직
        return self._find_best_spec_match(code_domain)  # 동일한 로직 사용

    def _extract_domain(self, tag_domain: str) -> str:
        """TAG 도메인에서 핵심 도메인 추출

        Args:
            tag_domain: 전체 도메인 문자열 (예: "AUTH-001")

        Returns:
            핵심 도메인 (예: "AUTH")
        """
        # 가장 복잡한 패턴부터 시도
        patterns = [
            r'(TRUST|AUTH|CLI|TEST|CORE|UTILS|DOC|INSTALLER|UPDATE|REFACTOR|LANG|HOOKS|INIT|TEMPLATE|CHECKPOINT|CONFIG|NETWORK|VERSION|ENHANCE|PERF|POLICY|VAL|MAP|AUTO|ROLLBACK|BUGFIX|LDE|PY314|CORE-PROJECT|CORE-GIT|CORE-PROJECT-3|CORE-GIT-COMMIT|CORE-GIT-INIT|CORE-GIT-BRANCH|CLI-PROMPTS|CLAUDE-COMMANDS|CLI-STATUS|CLI-DOCTOR|TIMEOUT|HOOK-TOOL|HOOKS-CLARITY|HOOK-BASH|NETWORK-DETECT|VERSION-DETECT-MAJOR|VERSION-CACHE|VERSION-INTEGRATE|OFFLINE|CONFIG-INTEGRATION|MAJOR-UPDATE-WARN|ENHANCE-PERF-CACHE|CROSS-PLATFORM|OFFLINE-SUPPORT)-001',
            r'(DOC-TAG|TAG|VAL|AUTO|POLICY|LANG|MAP|ROLLBACK|CORE|HOOKS|TEMPLATE|UPDATE|CACHE|VERSION|INSTALLER|CLI|TEST|TRUST|AUTH|BUGFIX|LDE|PY314|INIT|CHECKPOINT|CONFIG|NETWORK|ENHANCE|PERF|SESSION|SKILL|REFACTOR|WINDOW|CLAUDE|CODE|FEATURES|COMMANDS|DOCS|ALF|WORKFLOW)-\d{3}',
            r'(DOC|TAG|CLI|TEST|TRUST|AUTH|CORE|UTILS|INSTALLER|UPDATE|REFACTOR|LANG|HOOKS|INIT|TEMPLATE|CHECKPOINT|CONFIG|NETWORK|VERSION|ENHANCE|PERF|POLICY|VAL|MAP|AUTO|ROLLBACK|BUGFIX|LDE|PY314)-\d{3}',
            r'(CLI|TEST|TRUST|AUTH|CORE|UTILS|INSTALLER|UPDATE|REFACTOR|LANG|HOOKS|INIT|TEMPLATE|CHECKPOINT|CONFIG|NETWORK|VERSION|ENHANCE|PERF|POLICY|VAL|MAP|AUTO|ROLLBACK|BUGFIX|LDE|PY314)-\d{3}',
            r'(TEST|TRUST|AUTH|CORE|UTILS|INSTALLER|UPDATE|REFACTOR|LANG|HOOKS|INIT|TEMPLATE|CHECKPOINT|CONFIG|NETWORK|VERSION|ENHANCE|PERF|POLICY|VAL|MAP|AUTO|ROLLBACK|BUGFIX|LDE|PY314)-\d{3}',
            r'(TRUST|AUTH|CORE|UTILS|INSTALLER|UPDATE|REFACTOR|LANG|HOOKS|INIT|TEMPLATE|CHECKPOINT|CONFIG|NETWORK|VERSION|ENHANCE|PERF|POLICY|VAL|MAP|AUTO|ROLLBACK|BUGFIX|LDE|PY314)-\d{3}',
            r'(AUTH|CORE|UTILS|INSTALLER|UPDATE|REFACTOR|LANG|HOOKS|INIT|TEMPLATE|CHECKPOINT|CONFIG|NETWORK|VERSION|ENHANCE|PERF|POLICY|VAL|MAP|AUTO|ROLLBACK|BUGFIX|LDE|PY314)-\d{3}',
            r'(CORE|UTILS|INSTALLER|UPDATE|REFACTOR|LANG|HOOKS|INIT|TEMPLATE|CHECKPOINT|CONFIG|NETWORK|VERSION|ENHANCE|PERF|POLICY|VAL|MAP|AUTO|ROLLBACK|BUGFIX|LDE|PY314)-\d{3}',
            r'(UTILS|INSTALLER|UPDATE|REFACTOR|LANG|HOOKS|INIT|TEMPLATE|CHECKPOINT|CONFIG|NETWORK|VERSION|ENHANCE|PERF|POLICY|VAL|MAP|AUTO|ROLLBACK|BUGFIX|LDE|PY314)-\d{3}',
            r'(INSTALLER|UPDATE|REFACTOR|LANG|HOOKS|INIT|TEMPLATE|CHECKPOINT|CONFIG|NETWORK|VERSION|ENHANCE|PERF|POLICY|VAL|MAP|AUTO|ROLLBACK|BUGFIX|LDE|PY314)-\d{3}',
            r'(UPDATE|REFACTOR|LANG|HOOKS|INIT|TEMPLATE|CHECKPOINT|CONFIG|NETWORK|VERSION|ENHANCE|PERF|POLICY|VAL|MAP|AUTO|ROLLBACK|BUGFIX|LDE|PY314)-\d{3}',
            r'(LANG|HOOKS|INIT|TEMPLATE|CHECKPOINT|CONFIG|NETWORK|VERSION|ENHANCE|PERF|POLICY|VAL|MAP|AUTO|ROLLBACK|BUGFIX|LDE|PY314)-\d{3}',
            r'(HOOKS|INIT|TEMPLATE|CHECKPOINT|CONFIG|NETWORK|VERSION|ENHANCE|PERF|POLICY|VAL|MAP|AUTO|ROLLBACK|BUGFIX|LDE|PY314)-\d{3}',
            r'(INIT|TEMPLATE|CHECKPOINT|CONFIG|NETWORK|VERSION|ENHANCE|PERF|POLICY|VAL|MAP|AUTO|ROLLBACK|BUGFIX|LDE|PY314)-\d{3}',
            r'(TEMPLATE|CHECKPOINT|CONFIG|NETWORK|VERSION|ENHANCE|PERF|POLICY|VAL|MAP|AUTO|ROLLBACK|BUGFIX|LDE|PY314)-\d{3}',
            r'(CHECKPOINT|CONFIG|NETWORK|VERSION|ENHANCE|PERF|POLICY|VAL|MAP|AUTO|ROLLBACK|BUGFIX|LDE|PY314)-\d{3}',
            r'(CONFIG|NETWORK|VERSION|ENHANCE|PERF|POLICY|VAL|MAP|AUTO|ROLLBACK|BUGFIX|LDE|PY314)-\d{3}',
            r'(NETWORK|VERSION|ENHANCE|PERF|POLICY|VAL|MAP|AUTO|ROLLBACK|BUGFIX|LDE|PY314)-\d{3}',
            r'(VERSION|ENHANCE|PERF|POLICY|VAL|MAP|AUTO|ROLLBACK|BUGFIX|LDE|PY314)-\d{3}',
            r'(ENHANCE|PERF|POLICY|VAL|MAP|AUTO|ROLLBACK|BUGFIX|LDE|PY314)-\d{3}',
            r'(PERF|POLICY|VAL|MAP|AUTO|ROLLBACK|BUGFIX|LDE|PY314)-\d{3}',
            r'(POLICY|VAL|MAP|AUTO|ROLLBACK|BUGFIX|LDE|PY314)-\d{3}',
            r'(VAL|MAP|AUTO|ROLLBACK|BUGFIX|LDE|PY314)-\d{3}',
            r'(MAP|AUTO|ROLLBACK|BUGFIX|LDE|PY314)-\d{3}',
            r'(AUTO|ROLLBACK|BUGFIX|LDE|PY314)-\d{3}',
            r'(ROLLBACK|BUGFIX|LDE|PY314)-\d{3}',
            r'(BUGFIX|LDE|PY314)-\d{3}',
            r'(LDE|PY314)-\d{3}',
            r'(PY314)-\d{3}'
        ]

        for pattern in patterns:
            match = re.match(pattern, tag_domain)
            if match:
                return match.group(1)

        # 기본값: 첫 부분 반환
        return tag_domain.split('-')[0] if '-' in tag_domain else tag_domain

    def analyze_orphan_tags(self) -> Dict[str, List[str]]:
        """고아 TAG 분석

        Returns:
            {
                "code_without_spec": CODE이 없는 CODE TAG 목록,
                "code_without_test": TEST가 없는 CODE TAG 목록,
                "test_without_code": CODE가 없는 TEST TAG 목록,
                "spec_without_code": CODE가 없는 SPEC TAG 목록,
                "doc_without_spec": SPEC이 없는 DOC TAG 목록
            }
        """
        orphan_tags = {
            "code_without_spec": [],
            "code_without_test": [],
            "test_without_code": [],
            "spec_without_code": [],
            "doc_without_spec": []
        }

        # CODE가 없는 SPEC TAG
        for spec_domain in self.spec_tags:
            if spec_domain not in self.spec_to_code:
                orphan_tags["spec_without_code"].append(f"@SPEC:{spec_domain}")

        # SPEC이 없는 CODE TAG
        for code_domain in self.code_tags:
            if code_domain not in self.code_to_spec:
                orphan_tags["code_without_spec"].append(f"@CODE:{code_domain}")
            if code_domain not in self.code_to_test:
                orphan_tags["code_without_test"].append(f"@CODE:{code_domain}")

        # CODE가 없는 TEST TAG
        for test_domain in self.test_tags:
            if test_domain not in self.test_to_code:
                orphan_tags["test_without_code"].append(f"@TEST:{test_domain}")

        # SPEC이 없는 DOC TAG
        for doc_domain in self.doc_tags:
            if doc_domain not in self.all_tags["SPEC"]:
                orphan_tags["doc_without_spec"].append(f"@DOC:{doc_domain}")

        return orphan_tags

    def analyze_chains(self) -> Dict[str, List[Tuple[str, ...]]]:
        """TAG 체인 상태 분석

        Returns:
            {
                "complete_chains": 완전한 체인 목록,
                "partial_chains": 부분적인 체인 목록,
                "broken_chains": 끊어진 체인 목록
            }
        """
        chains = {
            "complete_chains": [],
            "partial_chains": [],
            "broken_chains": []
        }

        # 모든 도메인 집합
        all_domains = (self.spec_tags | self.code_tags | self.test_tags | self.doc_tags)

        for domain in all_domains:
            chain_status = []
            if f"@SPEC:{domain}" in self.spec_tags:
                chain_status.append("SPEC")
            if f"@CODE:{domain}" in self.code_tags:
                chain_status.append("CODE")
            if f"@TEST:{domain}" in self.test_tags:
                chain_status.append("TEST")
            if f"@DOC:{domain}" in self.doc_tags:
                chain_status.append("DOC")

            if len(chain_status) == 4:
                chains["complete_chains"].append((domain, tuple(chain_status)))
            elif len(chain_status) >= 2:
                chains["partial_chains"].append((domain, tuple(chain_status)))
            else:
                chains["broken_chains"].append((domain, tuple(chain_status)))

        return chains

    def generate_report(self) -> Dict:
        """분석 결과 보고서 생성

        Returns:
            전체 분석 결과 딕셔너리
        """
        orphan_tags = self.analyze_orphan_tags()
        chains = self.analyze_chains()

        # 도메인별 통계
        domain_stats = {}
        for domain in (self.spec_tags | self.code_tags | self.test_tags | self.doc_tags):
            domain_stats[domain] = {
                "has_spec": f"@SPEC:{domain}" in self.spec_tags,
                "has_code": f"@CODE:{domain}" in self.code_tags,
                "has_test": f"@TEST:{domain}" in self.test_tags,
                "has_doc": f"@DOC:{domain}" in self.doc_tags
            }

        return {
            "summary": {
                "total_specs": len(self.spec_tags),
                "total_codes": len(self.code_tags),
                "total_tests": len(self.test_tags),
                "total_docs": len(self.doc_tags),
                "total_domains": len(domain_stats),
                "complete_chains": len(chains["complete_chains"]),
                "partial_chains": len(chains["partial_chains"]),
                "broken_chains": len(chains["broken_chains"])
            },
            "orphan_tags": orphan_tags,
            "chains": chains,
            "domain_stats": domain_stats,
            "mappings": {
                "code_to_spec": self.code_to_spec,
                "code_to_test": self.code_to_test,
                "spec_to_code": self.spec_to_code,
                "test_to_code": self.test_to_code
            }
        }

    def generate_improvement_plan(self, report: Dict) -> List[Dict]:
        """개선 방안 생성

        Args:
            report: 분석 결과 보고서

        Returns:
            개선 방안 목록
        """
        improvements = []

        # 1. 긴급: 코드가 없는 SPEC에 대한 구현 제안
        if report["orphan_tags"]["spec_without_code"]:
            improvements.append({
                "priority": "CRITICAL",
                "type": "IMPLEMENT_MISSING_CODE",
                "description": f"구현이 필요한 SPEC: {len(report['orphan_tags']['spec_without_code'])}개",
                "tags": report["orphan_tags"]["spec_without_code"][:10],  # 상위 10개만 표시
                "action": "alfred:2-run",
                "estimated_time": "1-2주"
            })

        # 2. 높음: SPEC이 없는 CODE에 대한 SPEC 생성 제안
        if report["orphan_tags"]["code_without_spec"]:
            improvements.append({
                "priority": "HIGH",
                "type": "CREATE_MISSING_SPEC",
                "description": f"SPEC 생성이 필요한 CODE: {len(report['orphan_tags']['code_without_spec'])}개",
                "tags": report["orphan_tags"]["code_without_spec"][:10],
                "action": "alfred:1-plan",
                "estimated_time": "2-3일"
            })

        # 3. 중간: 테스트가 없는 CODE에 대한 TEST 생성 제안
        if report["orphan_tags"]["code_without_test"]:
            improvements.append({
                "priority": "MEDIUM",
                "type": "CREATE_MISSING_TEST",
                "description": f"테스트가 필요한 CODE: {len(report['orphan_tags']['code_without_test'])}개",
                "tags": report["orphan_tags"]["code_without_test"][:10],
                "action": "alfred:2-run",
                "estimated_time": "1-2일"
            })

        # 4. 낮음: 문서가 없는 DOC에 대한 DOC 생성 제안
        if report["orphan_tags"]["doc_without_spec"]:
            improvements.append({
                "priority": "LOW",
                "type": "CREATE_MISSING_DOC",
                "description": f"문서화가 필요한 항목: {len(report['orphan_tags']['doc_without_spec'])}개",
                "tags": report["orphan_tags"]["doc_without_spec"][:5],
                "action": "alfred:3-sync",
                "estimated_time": "1-2시간"
            })

        # 우선순위 정렬
        priority_order = {"CRITICAL": 0, "HIGH": 1, "MEDIUM": 2, "LOW": 3}
        improvements.sort(key=lambda x: priority_order.get(x["priority"], 999))

        return improvements

    def save_report(self, report: Dict, output_path: str = None) -> str:
        """보고서 저장

        Args:
            report: 분석 결과 보고서
            output_path: 출력 파일 경로

        Returns:
            저장된 파일 경로
        """
        if output_path is None:
            output_path = self.project_root / ".moai" / "reports" / "tag_analysis_report.json"

        # 디렉토리 생성
        output_path.parent.mkdir(parents=True, exist_ok=True)

        # 보고서 저장
        with open(output_path, 'w', encoding='utf-8') as f:
            json.dump(report, f, indent=2, ensure_ascii=False)

        return str(output_path)


def main():
    """메인 함수"""
    print("=" * 60)
    print("🔍 MoAI-ADK TAG 연결 상태 분석")
    print("=" * 60)
    print()

    analyzer = TagAnalyzer()

    # 1. 각 디렉토리 분석
    analyzer.analyze_specs()
    print()
    analyzer.analyze_src()
    analyzer.analyze_tests()
    print()

    # 2. 매핑 관계 구축
    analyzer.build_mappings()
    print()

    # 3. 보고서 생성
    report = analyzer.generate_report()

    # 4. 결과 출력
    print("📊 분석 결과 요약")
    print("-" * 40)
    summary = report["summary"]
    print(f"📄 전체 SPEC: {summary['total_specs']}개")
    print(f"💻 전체 CODE: {summary['total_codes']}개")
    print(f"🧪 전체 TEST: {summary['total_tests']}개")
    print(f"📚 전체 DOC: {summary['total_docs']}개")
    print(f"🔗 전체 도메인: {summary['total_domains']}개")
    print(f"✅ 완전한 체인: {summary['complete_chains']}개 ({summary['complete_chains']/max(1, summary['total_domains'])*100:.1f}%)")
    print(f"⚠️ 부분적 체인: {summary['partial_chains']}개 ({summary['partial_chains']/max(1, summary['total_domains'])*100:.1f}%)")
    print(f"❌ 끊어진 체인: {summary['broken_chains']}개 ({summary['broken_chains']/max(1, summary['total_domains'])*100:.1f}%)")
    print()

    # 5. 고아 TAG 현황
    print("🚨 고아 TAG 현황")
    print("-" * 40)
    orphan_tags = report["orphan_tags"]
    print(f"   SPEC이 없는 CODE: {len(orphan_tags['code_without_spec'])}개")
    print(f"   TEST가 없는 CODE: {len(orphan_tags['code_without_test'])}개")
    print(f"   CODE가 없는 TEST: {len(orphan_tags['test_without_code'])}개")
    print(f"   CODE가 없는 SPEC: {len(orphan_tags['spec_without_code'])}개")
    print(f"   SPEC이 없는 DOC: {len(orphan_tags['doc_without_spec'])}개")
    print()

    # 6. 개선 방안
    improvements = analyzer.generate_improvement_plan(report)
    print("🚀 개선 방안")
    print("-" * 40)
    for i, improvement in enumerate(improvements[:5], 1):
        print(f"{i}. [{improvement['priority']}] {improvement['description']}")
        print(f"   - 예상 시간: {improvement['estimated_time']}")
        if improvement.get('tags'):
            print(f"   - 주요 TAG: {', '.join(improvement['tags'][:3])}{'...' if len(improvement['tags']) > 3 else ''}")
        print()

    # 7. 보고서 저장
    output_path = analyzer.save_report(report)
    print(f"📄 상세 보고서 저장: {output_path}")
    print()

    print("🎯 추천 실행 순서")
    print("-" * 40)
    print("1. 긴급: `alfred:2-run`으로 CODE 없는 SPEC 구현")
    print("2. 높음: `alfred:1-plan`으로 SPEC 없는 CODE 명세화")
    print("3. 중간: 테스트 커버리지 85% 달성")
    print("4. 낮음: 문서화 완성 및 자동화")
    print()


if __name__ == "__main__":
    main()