#!/usr/bin/env python3
# @TASK:LICENSE-CHECK-011
"""
MoAI-ADK License Compliance Checker v0.1.12
프로젝트 의존성 라이선스 검사 및 호환성 검증

이 스크립트는 프로젝트의 모든 의존성을 스캔하여:
- 라이선스 정보 수집 및 분석
- 제한적 라이선스 (GPL, AGPL 등) 감지
- 라이선스 정책 준수 확인
- 상세한 라이선스 리포트 생성
"""

import json
import re
import subprocess
import sys
from dataclasses import dataclass
from datetime import datetime
from pathlib import Path
from typing import Any


@dataclass
class LicenseInfo:
    """라이선스 정보 구조"""
    name: str
    spdx_id: str | None
    category: str  # permissive, copyleft, proprietary, unknown
    risk_level: str  # low, medium, high, critical
    restrictions: list[str]
    compatibility: dict[str, bool]  # 다른 라이선스와의 호환성

@dataclass
class PackageLicense:
    """패키지 라이선스 정보"""
    package: str
    version: str
    license: str
    license_info: LicenseInfo | None
    source: str  # requirements.txt, package.json, etc.
    status: str  # compliant, non-compliant, needs-review

class LicenseChecker:
    """라이선스 호환성 검사기"""

    def __init__(self, project_root: Path):
        self.project_root = project_root
        self.moai_dir = project_root / ".moai"

        # 라이선스 데이터베이스
        self.license_db = self.init_license_database()

        # 정책 설정
        self.policy = self.load_license_policy()

        # 결과 저장
        self.scan_results = []
        self.violations = []
        self.warnings = []

    def init_license_database(self) -> dict[str, LicenseInfo]:
        """라이선스 정보 데이터베이스 초기화"""

        return {
            # Permissive Licenses (Low Risk)
            "MIT": LicenseInfo(
                name="MIT License",
                spdx_id="MIT",
                category="permissive",
                risk_level="low",
                restrictions=["include-copyright"],
                compatibility={"GPL": True, "Apache": True, "BSD": True}
            ),
            "Apache-2.0": LicenseInfo(
                name="Apache License 2.0",
                spdx_id="Apache-2.0",
                category="permissive",
                risk_level="low",
                restrictions=["include-copyright", "include-license", "state-changes"],
                compatibility={"GPL": True, "MIT": True, "BSD": True}
            ),
            "BSD-3-Clause": LicenseInfo(
                name="BSD 3-Clause License",
                spdx_id="BSD-3-Clause",
                category="permissive",
                risk_level="low",
                restrictions=["include-copyright", "no-endorsement"],
                compatibility={"GPL": True, "Apache": True, "MIT": True}
            ),
            "ISC": LicenseInfo(
                name="ISC License",
                spdx_id="ISC",
                category="permissive",
                risk_level="low",
                restrictions=["include-copyright"],
                compatibility={"GPL": True, "Apache": True, "MIT": True}
            ),

            # Copyleft Licenses (Medium to High Risk)
            "GPL-2.0": LicenseInfo(
                name="GNU General Public License v2.0",
                spdx_id="GPL-2.0-only",
                category="copyleft",
                risk_level="high",
                restrictions=["disclose-source", "license-compatibility", "same-license"],
                compatibility={"Apache": False, "MIT": False, "BSD": False}
            ),
            "GPL-3.0": LicenseInfo(
                name="GNU General Public License v3.0",
                spdx_id="GPL-3.0-only",
                category="copyleft",
                risk_level="high",
                restrictions=["disclose-source", "license-compatibility", "same-license", "patent-grant"],
                compatibility={"Apache": True, "MIT": False, "BSD": False}
            ),
            "AGPL-3.0": LicenseInfo(
                name="GNU Affero General Public License v3.0",
                spdx_id="AGPL-3.0-only",
                category="copyleft",
                risk_level="critical",
                restrictions=["disclose-source", "network-copyleft", "same-license"],
                compatibility={"GPL": True, "Apache": False, "MIT": False}
            ),
            "LGPL-2.1": LicenseInfo(
                name="GNU Lesser General Public License v2.1",
                spdx_id="LGPL-2.1-only",
                category="weak-copyleft",
                risk_level="medium",
                restrictions=["disclose-source-modifications", "license-compatibility"],
                compatibility={"GPL": True, "Apache": True, "MIT": True}
            ),

            # Proprietary/Commercial
            "UNLICENSED": LicenseInfo(
                name="Unlicensed/Proprietary",
                spdx_id=None,
                category="proprietary",
                risk_level="critical",
                restrictions=["commercial-use-restricted", "distribution-restricted"],
                compatibility={}
            )
        }

    def load_license_policy(self) -> dict[str, Any]:
        """라이선스 정책 로드"""

        policy_file = self.moai_dir / "config.json"
        default_policy = {
            "allowed_licenses": ["MIT", "Apache-2.0", "BSD-3-Clause", "ISC"],
            "restricted_licenses": ["GPL-2.0", "GPL-3.0", "AGPL-3.0", "UNLICENSED"],
            "review_required": ["LGPL-2.1", "MPL-2.0", "CC-BY-4.0"],
            "max_risk_level": "medium",
            "allow_dual_license": True,
            "require_attribution": True
        }

        if policy_file.exists():
            try:
                config = json.loads(policy_file.read_text())
                return config.get("license_policy", default_policy)
            except:
                pass

        return default_policy

    def scan_python_dependencies(self) -> list[PackageLicense]:
        """Python 의존성 스캔"""
        results = []

        # pip list로 설치된 패키지 확인
        try:
            pip_result = subprocess.run([
                'pip', 'list', '--format=json'
            ], capture_output=True, text=True, timeout=30)

            if pip_result.returncode == 0:
                packages = json.loads(pip_result.stdout)

                for pkg in packages:
                    pkg_name = pkg['name']
                    pkg_version = pkg['version']

                    # 라이선스 정보 가져오기
                    license_info = self.get_package_license(pkg_name)

                    results.append(PackageLicense(
                        package=pkg_name,
                        version=pkg_version,
                        license=license_info.get('license', 'Unknown'),
                        license_info=self.license_db.get(license_info.get('license')),
                        source='pip',
                        status=self.evaluate_license_compliance(license_info.get('license'))
                    ))

        except Exception as error:
            self.warnings.append(f"Python dependency scan failed: {error}")

        return results

    def scan_nodejs_dependencies(self) -> list[PackageLicense]:
        """Node.js 의존성 스캔"""
        results = []

        package_json = self.project_root / "package.json"
        if not package_json.exists():
            return results

        try:
            # npm ls로 의존성 트리 확인
            npm_result = subprocess.run([
                'npm', 'ls', '--json', '--depth=0'
            ], capture_output=True, text=True, timeout=60, cwd=self.project_root)

            if npm_result.returncode == 0:
                npm_data = json.loads(npm_result.stdout)
                dependencies = npm_data.get('dependencies', {})

                for pkg_name, pkg_info in dependencies.items():
                    version = pkg_info.get('version', 'unknown')

                    # package.json에서 라이선스 정보 확인
                    license_info = self.get_npm_package_license(pkg_name)

                    results.append(PackageLicense(
                        package=pkg_name,
                        version=version,
                        license=license_info,
                        license_info=self.license_db.get(license_info),
                        source='npm',
                        status=self.evaluate_license_compliance(license_info)
                    ))

        except Exception as error:
            self.warnings.append(f"Node.js dependency scan failed: {error}")

        return results

    def get_package_license(self, package_name: str) -> dict[str, str]:
        """Python 패키지의 라이선스 정보 조회"""

        try:
            # pip show로 패키지 정보 확인
            show_result = subprocess.run([
                'pip', 'show', package_name
            ], capture_output=True, text=True, timeout=10)

            if show_result.returncode == 0:
                output = show_result.stdout

                # License 필드 추출
                license_match = re.search(r'License: (.+)', output)
                if license_match:
                    license_text = license_match.group(1).strip()

                    # 라이선스 정규화
                    normalized_license = self.normalize_license_name(license_text)

                    return {
                        'license': normalized_license,
                        'raw_license': license_text
                    }

        except Exception:
            pass

        return {'license': 'Unknown'}

    def get_npm_package_license(self, package_name: str) -> str:
        """NPM 패키지의 라이선스 정보 조회"""

        try:
            # node_modules에서 package.json 확인
            pkg_path = self.project_root / "node_modules" / package_name / "package.json"

            if pkg_path.exists():
                pkg_data = json.loads(pkg_path.read_text())
                license_info = pkg_data.get('license', 'Unknown')

                if isinstance(license_info, dict):
                    license_info = license_info.get('type', 'Unknown')

                return self.normalize_license_name(str(license_info))

        except Exception:
            pass

        return 'Unknown'

    def normalize_license_name(self, license_text: str) -> str:
        """라이선스 이름 정규화"""

        if not license_text or license_text.lower() in ['unknown', 'none', '']:
            return 'Unknown'

        # 일반적인 라이선스 별칭 처리
        license_aliases = {
            'MIT': 'MIT',
            'Apache': 'Apache-2.0',
            'Apache 2.0': 'Apache-2.0',
            'Apache-2': 'Apache-2.0',
            'BSD': 'BSD-3-Clause',
            'BSD-3': 'BSD-3-Clause',
            'GPL': 'GPL-3.0',
            'GPL-2': 'GPL-2.0',
            'GPL-3': 'GPL-3.0',
            'LGPL': 'LGPL-2.1',
            'AGPL': 'AGPL-3.0',
            'ISC': 'ISC',
            'UNLICENSED': 'UNLICENSED'
        }

        license_upper = license_text.upper()
        for alias, standard in license_aliases.items():
            if alias.upper() in license_upper:
                return standard

        return license_text

    def evaluate_license_compliance(self, license_name: str) -> str:
        """라이선스 컴플라이언스 평가"""

        if license_name in self.policy['allowed_licenses']:
            return 'compliant'
        elif license_name in self.policy['restricted_licenses']:
            return 'non-compliant'
        elif license_name in self.policy['review_required'] or license_name == 'Unknown':
            return 'needs-review'
        else:
            # 위험 수준으로 판단
            license_info = self.license_db.get(license_name)
            if license_info:
                if license_info.risk_level in ['critical', 'high']:
                    return 'non-compliant'
                elif license_info.risk_level == 'medium':
                    return 'needs-review'
                else:
                    return 'compliant'

            return 'needs-review'

    def generate_report(self, scan_results: list[PackageLicense]) -> dict[str, Any]:
        """라이선스 스캔 리포트 생성"""

        # 상태별 분류
        compliant = [r for r in scan_results if r.status == 'compliant']
        non_compliant = [r for r in scan_results if r.status == 'non-compliant']
        needs_review = [r for r in scan_results if r.status == 'needs-review']

        # 위험 분석
        critical_violations = []
        high_risk_packages = []

        for result in scan_results:
            if result.license_info and result.license_info.risk_level == 'critical':
                critical_violations.append(result)
            elif result.license_info and result.license_info.risk_level == 'high':
                high_risk_packages.append(result)

        return {
            'scan_summary': {
                'total_packages': len(scan_results),
                'compliant': len(compliant),
                'non_compliant': len(non_compliant),
                'needs_review': len(needs_review),
                'scan_date': datetime.now().isoformat()
            },
            'compliance_status': 'PASS' if len(non_compliant) == 0 and len(critical_violations) == 0 else 'FAIL',
            'critical_violations': [
                {
                    'package': v.package,
                    'version': v.version,
                    'license': v.license,
                    'reason': 'Critical license risk'
                } for v in critical_violations
            ],
            'license_distribution': self.get_license_distribution(scan_results),
            'recommendations': self.generate_recommendations(scan_results),
            'detailed_results': [
                {
                    'package': r.package,
                    'version': r.version,
                    'license': r.license,
                    'status': r.status,
                    'source': r.source,
                    'risk_level': r.license_info.risk_level if r.license_info else 'unknown'
                } for r in scan_results
            ]
        }

    def get_license_distribution(self, results: list[PackageLicense]) -> dict[str, int]:
        """라이선스 분포 통계"""
        distribution = {}

        for result in results:
            license_name = result.license
            distribution[license_name] = distribution.get(license_name, 0) + 1

        return distribution

    def generate_recommendations(self, results: list[PackageLicense]) -> list[str]:
        """개선 권장사항 생성"""
        recommendations = []

        non_compliant = [r for r in results if r.status == 'non-compliant']
        if non_compliant:
            recommendations.append(f"{len(non_compliant)}개의 비호환 라이선스 패키지를 대체하거나 제거하세요")

        needs_review = [r for r in results if r.status == 'needs-review']
        if needs_review:
            recommendations.append(f"{len(needs_review)}개의 패키지에 대한 라이선스 검토가 필요합니다")

        unknown_licenses = [r for r in results if r.license == 'Unknown']
        if unknown_licenses:
            recommendations.append(f"{len(unknown_licenses)}개의 패키지 라이선스 정보를 확인하세요")

        if not recommendations:
            recommendations.append("모든 라이선스가 정책에 준수합니다")

        return recommendations

    def run_scan(self) -> dict[str, Any]:
        """전체 라이선스 스캔 실행"""

        print("🔍 Starting license compliance scan...")

        all_results = []

        # Python 의존성 스캔
        print("  Scanning Python dependencies...")
        python_results = self.scan_python_dependencies()
        all_results.extend(python_results)

        # Node.js 의존성 스캔
        print("  Scanning Node.js dependencies...")
        nodejs_results = self.scan_nodejs_dependencies()
        all_results.extend(nodejs_results)

        print(f"  Found {len(all_results)} packages")

        # 리포트 생성
        report = self.generate_report(all_results)

        return report

def main():
    """메인 실행 함수"""

    project_root = Path.cwd()

    # MoAI 프로젝트 확인
    if not (project_root / ".moai").exists():
        print("❌ This is not a MoAI-ADK project")
        sys.exit(1)

    try:
        # 라이선스 검사 실행
        checker = LicenseChecker(project_root)
        report = checker.run_scan()

        # 결과 출력
        print("\n" + "="*60)
        print("📋 LICENSE COMPLIANCE REPORT")
        print("="*60)

        summary = report['scan_summary']
        print(f"Total Packages: {summary['total_packages']}")
        print(f"Compliant: {summary['compliant']}")
        print(f"Non-Compliant: {summary['non_compliant']}")
        print(f"Needs Review: {summary['needs_review']}")
        print(f"Status: {'✅ PASS' if report['compliance_status'] == 'PASS' else '❌ FAIL'}")

        # 위반 사항 출력
        if report['critical_violations']:
            print(f"\n🚨 Critical Violations ({len(report['critical_violations'])}):")
            for violation in report['critical_violations']:
                print(f"  • {violation['package']} ({violation['license']}) - {violation['reason']}")

        # 라이선스 분포
        print("\n📊 License Distribution:")
        for license_name, count in sorted(report['license_distribution'].items()):
            print(f"  • {license_name}: {count}")

        # 권장사항
        print("\n💡 Recommendations:")
        for rec in report['recommendations']:
            print(f"  • {rec}")

        # 리포트 파일 저장
        report_file = project_root / ".moai" / "reports" / "license_scan.json"
        report_file.parent.mkdir(parents=True, exist_ok=True)
        report_file.write_text(json.dumps(report, indent=2))

        print(f"\n📄 Detailed report saved to: {report_file}")

        # Exit code
        sys.exit(0 if report['compliance_status'] == 'PASS' else 1)

    except Exception as error:
        print(f"❌ License scan failed: {error}")
        sys.exit(1)

if __name__ == "__main__":
    main()
