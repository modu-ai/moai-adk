# @CODE:VAL-002
"""TAG chain analysis tool for identifying broken chains.

Analyzes TAG chains across the codebase to identify broken links
between SPEC, CODE, TEST, and DOC elements.

@SPEC:DOCS-005: TAG 체인 분석 및 검증 도구
"""

import json
import re
from dataclasses import dataclass
from pathlib import Path
from typing import Dict, List, Optional


@dataclass
class TagChain:
    """Represents a complete TAG chain across all artifact types."""

    domain: str
    spec_id: Optional[str] = None
    code_id: Optional[str] = None
    test_id: Optional[str] = None
    doc_id: Optional[str] = None

    @property
    def is_complete(self) -> bool:
        """Check if chain has all required elements."""
        has_spec = self.spec_id is not None
        has_code = self.code_id is not None
        has_test = self.test_id is not None
        return has_spec and has_code and has_test

    @property
    def missing_elements(self) -> List[str]:
        """List missing elements in the chain."""
        missing = []
        if self.spec_id is None:
            missing.append("SPEC")
        if self.code_id is None:
            missing.append("CODE")
        if self.test_id is None:
            missing.append("TEST")
        return missing

    @property
    def completeness_score(self) -> float:
        """Calculate completeness score (0.0 to 1.0)."""
        elements = [self.spec_id, self.code_id, self.test_id]
        return sum(1 for el in elements if el is not None) / len(elements)


@dataclass
class ChainAnalysisResult:
    """Analysis result for TAG chains."""

    total_chains: int
    complete_chains: int
    partial_chains: int
    broken_chains: int
    orphans_by_type: Dict[str, List[str]]
    chains_by_domain: Dict[str, List[TagChain]]
    broken_chain_details: List[Dict[str, str]]


class TagChainAnalyzer:
    """Analyzes TAG chains across the MoAI-ADK codebase."""

    def __init__(self, root_path: Path = Path(".")):
        self.root_path = root_path
        self.tag_pattern = re.compile(r'@[A-Z]+:[A-Z0-9-]+\d{3}')

    def analyze_all_chains(self) -> ChainAnalysisResult:
        """Analyze all TAG chains in the codebase."""
        # Scan for all TAGs
        all_tags = self._scan_all_tags()

        # Group by domain
        chains_by_domain = self._group_chains_by_domain(all_tags)

        # Analyze chain completeness
        complete_chains = 0
        partial_chains = 0
        broken_chains = 0
        broken_chain_details = []
        orphans_by_type = {
            "code_without_spec": [],
            "code_without_test": [],
            "test_without_code": [],
            "spec_without_code": [],
            "doc_without_spec": []
        }

        for domain, chains in chains_by_domain.items():
            for chain in chains:
                if chain.is_complete:
                    complete_chains += 1
                elif chain.completeness_score > 0:
                    partial_chains += 1
                    broken_chain_details.append({
                        "domain": domain,
                        "chain": str(chain),
                        "missing": chain.missing_elements,
                        "score": chain.completeness_score
                    })
                else:
                    broken_chains += 1
                    broken_chain_details.append({
                        "domain": domain,
                        "chain": str(chain),
                        "missing": chain.missing_elements,
                        "score": chain.completeness_score
                    })

        # Identify orphans
        orphans_by_type = self._identify_orphans(all_tags)

        total_chains = complete_chains + partial_chains + broken_chains

        return ChainAnalysisResult(
            total_chains=total_chains,
            complete_chains=complete_chains,
            partial_chains=partial_chains,
            broken_chains=broken_chains,
            orphans_by_type=orphans_by_type,
            chains_by_domain=chains_by_domain,
            broken_chain_details=broken_chain_details
        )

    def _scan_all_tags(self) -> Dict[str, List[str]]:
        """Scan all files for TAG occurrences."""
        all_tags = {}

        # Scan source files
        src_path = self.root_path / "src"
        if src_path.exists():
            for py_file in src_path.rglob("*.py"):
                tags = self._extract_tags_from_file(py_file)
                for tag in tags:
                    if tag not in all_tags:
                        all_tags[tag] = []
                    all_tags[tag].append(str(py_file))

        # Scan test files
        tests_path = self.root_path / "tests"
        if tests_path.exists():
            for py_file in tests_path.rglob("*.py"):
                tags = self._extract_tags_from_file(py_file)
                for tag in tags:
                    if tag not in all_tags:
                        all_tags[tag] = []
                    all_tags[tag].append(str(py_file))

        # Scan spec files
        specs_path = self.root_path / ".moai" / "specs"
        if specs_path.exists():
            for md_file in specs_path.rglob("*.md"):
                tags = self._extract_tags_from_file(md_file)
                for tag in tags:
                    if tag not in all_tags:
                        all_tags[tag] = []
                    all_tags[tag].append(str(md_file))

        return all_tags

    def _extract_tags_from_file(self, file_path: Path) -> List[str]:
        """Extract TAGs from a single file."""
        try:
            content = file_path.read_text(encoding='utf-8')
            matches = self.tag_pattern.findall(content)
            return list(set(matches))  # Remove duplicates
        except Exception:
            return []

    def _group_chains_by_domain(self, all_tags: Dict[str, List[str]]) -> Dict[str, List[TagChain]]:
        """Group TAGs by domain and create chain objects."""
        chains_by_domain = {}

        # Group by domain
        domain_groups = {}
        for tag, files in all_tags.items():
            domain = self._extract_domain_from_tag(tag)
            if domain not in domain_groups:
                domain_groups[domain] = []
            domain_groups[domain].append((tag, files))

        # Create chains for each domain
        for domain, tag_list in domain_groups.items():
            chains_by_domain[domain] = []

            # For each domain, create one chain per domain instance
            spec_tags = [tag for tag, _ in tag_list if tag.startswith("@SPEC:")]
            code_tags = [tag for tag, _ in tag_list if tag.startswith("@CODE:")]
            test_tags = [tag for tag, _ in tag_list if tag.startswith("@TEST:")]

            # Create chains based on highest number found
            max_number = self._get_max_number(spec_tags, code_tags, test_tags)

            for i in range(1, max_number + 1):
                spec_id = f"@SPEC:{domain}-{i:03d}" if f"@SPEC:{domain}-{i:03d}" in spec_tags else None
                code_id = f"@CODE:{domain}-{i:03d}" if f"@CODE:{domain}-{i:03d}" in code_tags else None
                test_id = f"@TEST:{domain}-{i:03d}" if f"@TEST:{domain}-{i:03d}" in test_tags else None

                chain = TagChain(
                    domain=domain,
                    spec_id=spec_id,
                    code_id=code_id,
                    test_id=test_id
                )
                chains_by_domain[domain].append(chain)

        return chains_by_domain

    def _extract_domain_from_tag(self, tag: str) -> str:
        """Extract domain from TAG (e.g., "@SPEC:AUTH-004" -> "AUTH")."""
        # Match pattern: @TYPE:DOMAIN-NUMBER where NUMBER is exactly 3 digits
        match = re.match(r'@[A-Z]+:(.+?)-\d{3}', tag)
        if match:
            return match.group(1)
        return tag.split(":")[1].rsplit("-", 1)[0]

    def _get_max_number(self, spec_tags: List[str], code_tags: List[str], test_tags: List[str]) -> int:
        """Get the highest number found across all tag types."""
        max_num = 0

        for tag_list in [spec_tags, code_tags, test_tags]:
            for tag in tag_list:
                match = re.search(r'-(\d{3})$', tag)
                if match:
                    num = int(match.group(1))
                    max_num = max(max_num, num)

        return max_num

    def _identify_orphans(self, all_tags: Dict[str, List[str]]) -> Dict[str, List[str]]:
        """Identify orphaned TAGs (missing linked elements)."""
        orphans = {
            "code_without_spec": [],
            "code_without_test": [],
            "test_without_code": [],
            "spec_without_code": [],
            "doc_without_spec": []
        }

        code_tags = [tag for tag in all_tags.keys() if tag.startswith("@CODE:")]
        test_tags = [tag for tag in all_tags.keys() if tag.startswith("@TEST:")]
        spec_tags = [tag for tag in all_tags.keys() if tag.startswith("@SPEC:")]

        # Find CODE without SPEC
        for code_tag in code_tags:
            domain = self._extract_domain_from_tag(code_tag)
            spec_tag = f"@SPEC:{domain}-{''.join(code_tag.split('-')[1:])}"
            if spec_tag not in spec_tags:
                orphans["code_without_spec"].append(code_tag)

        # Find CODE without TEST
        for code_tag in code_tags:
            domain = self._extract_domain_from_tag(code_tag)
            test_tag = f"@TEST:{domain}-{''.join(code_tag.split('-')[1:])}"
            if test_tag not in test_tags:
                orphans["code_without_test"].append(code_tag)

        # Find TEST without CODE
        for test_tag in test_tags:
            domain = self._extract_domain_from_tag(test_tag)
            code_tag = f"@CODE:{domain}-{''.join(test_tag.split('-')[1:])}"
            if code_tag not in code_tags:
                orphans["test_without_code"].append(test_tag)

        # Find SPEC without CODE
        for spec_tag in spec_tags:
            domain = self._extract_domain_from_tag(spec_tag)
            code_tag = f"@CODE:{domain}-{''.join(spec_tag.split('-')[1:])}"
            if code_tag not in code_tags:
                orphans["spec_without_code"].append(spec_tag)

        return orphans

    def generate_report(self, result: ChainAnalysisResult) -> str:
        """Generate a human-readable analysis report."""
        report = []
        report.append("# TAG Chain Analysis Report")
        report.append("")
        report.append("## Summary")
        report.append(f"- Total Chains: {result.total_chains}")
        complete_pct = result.complete_chains / result.total_chains * 100
        report.append(f"- Complete Chains: {result.complete_chains} ({complete_pct:.1f}%)")
        partial_pct = result.partial_chains / result.total_chains * 100
        report.append(f"- Partial Chains: {result.partial_chains} ({partial_pct:.1f}%)")
        broken_pct = result.broken_chains / result.total_chains * 100
        report.append(f"- Broken Chains: {result.broken_chains} ({broken_pct:.1f}%)")
        report.append("")

        report.append("## Orphan TAGs")
        for orphan_type, tags in result.orphans_by_type.items():
            report.append(f"- {len(tags)} {orphan_type.replace('_', ' ').title()}:")
            for tag in tags[:5]:  # Show first 5
                report.append(f"  - {tag}")
            if len(tags) > 5:
                report.append(f"  - ... and {len(tags) - 5} more")
            report.append("")

        report.append("## Broken Chain Details")
        for detail in result.broken_chain_details[:10]:  # Show first 10
            report.append(f"- {detail['domain']} ({detail['score']:.2f}): Missing {', '.join(detail['missing'])}")
        if len(result.broken_chain_details) > 10:
            report.append(f"- ... and {len(result.broken_chain_details) - 10} more")

        return "\n".join(report)


def analyze_tag_chains(root_path: Path = Path(".")) -> ChainAnalysisResult:
    """Convenience function to analyze TAG chains."""
    analyzer = TagChainAnalyzer(root_path)
    return analyzer.analyze_all_chains()


def main():
    """Main function for CLI usage."""
    import argparse

    parser = argparse.ArgumentParser(description="Analyze TAG chains in MoAI-ADK")
    parser.add_argument("--path", default=".", help="Path to analyze (default: current directory)")
    parser.add_argument("--output", help="Output file for JSON report")

    args = parser.parse_args()

    result = analyze_tag_chains(Path(args.path))
    analyzer = TagChainAnalyzer()

    if args.output:
        with open(args.output, 'w') as f:
            json.dump({
                "summary": {
                    "total_chains": result.total_chains,
                    "complete_chains": result.complete_chains,
                    "partial_chains": result.partial_chains,
                    "broken_chains": result.broken_chains
                },
                "orphans_by_type": result.orphans_by_type,
                "broken_chain_details": result.broken_chain_details
            }, f, indent=2)
    else:
        print(analyzer.generate_report(result))


if __name__ == "__main__":
    main()
