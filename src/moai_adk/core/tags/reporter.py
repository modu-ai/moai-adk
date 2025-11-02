#!/usr/bin/env python3
# @CODE:DOC-TAG-004 | Component 4: Documentation & Reporting system
"""TAG reporting and documentation generation for MoAI-ADK

This module provides automated reporting for TAG system health and coverage:
- Generates TAG inventories across entire codebase
- Creates coverage matrices showing SPEC implementation status
- Analyzes SPEC→CODE→TEST→DOC chain completeness
- Produces statistics and metrics in multiple formats
- Formats reports as Markdown, JSON, CSV, and HTML (optional)

Architecture:
    ReportGenerator (orchestrator)
    ├── InventoryGenerator (tag-inventory.md)
    ├── MatrixGenerator (tag-matrix.md)
    ├── CoverageAnalyzer (coverage analysis)
    ├── StatisticsGenerator (tag-statistics.json)
    └── ReportFormatter (multi-format output)

Usage:
    generator = ReportGenerator()
    result = generator.generate_all_reports("/path/to/project", "/path/to/output")
    print(f"Generated reports: {result.inventory_path}, {result.matrix_path}")
"""

import json
import re
from dataclasses import dataclass, field
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Set

# ============================================================================
# Data Models
# ============================================================================

@dataclass
class TagInventory:
    """Single TAG inventory item with metadata

    Attributes:
        tag_id: TAG identifier (e.g., "DOC-TAG-001")
        file_path: File path where TAG is located
        line_number: Line number of TAG
        context: Surrounding code snippet
        related_tags: List of related TAG strings
        last_modified: Last modification timestamp
        status: TAG status (active|deprecated|orphan|incomplete)
    """
    tag_id: str
    file_path: str
    line_number: int
    context: str
    related_tags: List[str] = field(default_factory=list)
    last_modified: datetime = field(default_factory=datetime.now)
    status: str = "active"


@dataclass
class TagMatrix:
    """Coverage matrix showing implementation status

    Attributes:
        rows: Dict mapping SPEC ID to coverage status
              {
                  "AUTH-001": {
                      "SPEC": True,
                      "CODE": True,
                      "TEST": False,
                      "DOC": False
                  }
              }
        completion_percentages: Dict mapping SPEC ID to completion percentage
    """
    rows: Dict[str, Dict[str, bool]] = field(default_factory=dict)
    completion_percentages: Dict[str, float] = field(default_factory=dict)


@dataclass
class CoverageMetrics:
    """Coverage metrics for a single SPEC

    Attributes:
        spec_id: SPEC identifier
        has_code: Whether CODE implementation exists
        has_test: Whether TEST exists
        has_doc: Whether DOC exists
        coverage_percentage: Overall completion percentage
    """
    spec_id: str
    has_code: bool = False
    has_test: bool = False
    has_doc: bool = False
    coverage_percentage: float = 0.0


@dataclass
class StatisticsReport:
    """Overall TAG statistics

    Attributes:
        generated_at: Report generation timestamp
        total_tags: Total TAG count
        by_type: Count by TAG type (SPEC, CODE, TEST, DOC)
        by_domain: Count by domain (AUTH, USER, etc.)
        coverage: Coverage metrics
        issues: Issue counts (orphans, incomplete chains, etc.)
    """
    generated_at: datetime
    total_tags: int
    by_type: Dict[str, int] = field(default_factory=dict)
    by_domain: Dict[str, int] = field(default_factory=dict)
    coverage: Dict[str, float] = field(default_factory=dict)
    issues: Dict[str, int] = field(default_factory=dict)


@dataclass
class ReportResult:
    """Result of report generation

    Attributes:
        inventory_path: Path to generated inventory file
        matrix_path: Path to generated matrix file
        statistics_path: Path to generated statistics file
        success: Whether generation succeeded
        error_message: Error message if failed
    """
    inventory_path: Path
    matrix_path: Path
    statistics_path: Path
    success: bool = True
    error_message: str = ""


# ============================================================================
# Core Generators
# ============================================================================

class InventoryGenerator:
    """Generates TAG inventory across codebase

    Scans entire codebase for TAGs and creates comprehensive inventory
    grouped by domain and type.
    """

    TAG_PATTERN = re.compile(r"@(SPEC|CODE|TEST|DOC):([A-Z]+(?:-[A-Z]+)*-\d{3})")
    IGNORE_PATTERNS = [".git/*", "node_modules/*", "__pycache__/*", "*.pyc", ".venv/*", "venv/*"]

    def generate_inventory(self, root_path: str) -> List[TagInventory]:
        """Scan directory and generate TAG inventory

        Args:
            root_path: Root directory to scan

        Returns:
            List of TagInventory objects
        """
        inventory: list[TagInventory] = []
        root = Path(root_path)

        if not root.exists() or not root.is_dir():
            return inventory

        # Scan all files recursively
        for filepath in root.rglob("*"):
            if not filepath.is_file():
                continue

            # Check ignore patterns
            if self._should_ignore(filepath, root):
                continue

            # Extract TAGs from file
            tags = self._extract_tags_from_file(filepath, root)
            inventory.extend(tags)

        return inventory

    def _should_ignore(self, filepath: Path, root: Path) -> bool:
        """Check if file should be ignored

        Args:
            filepath: File path to check
            root: Root directory

        Returns:
            True if file should be ignored
        """
        try:
            relative = filepath.relative_to(root)
            relative_str = str(relative)

            for pattern in self.IGNORE_PATTERNS:
                pattern_clean = pattern.replace("/*", "").replace("*", "")
                if pattern_clean in relative_str:
                    return True

            return False

        except ValueError:
            return True

    def _extract_tags_from_file(self, filepath: Path, root: Path) -> List[TagInventory]:
        """Extract TAGs from a single file

        Args:
            filepath: File to scan
            root: Root directory for relative paths

        Returns:
            List of TagInventory objects
        """
        inventory = []

        try:
            content = filepath.read_text(encoding="utf-8", errors="ignore")
            lines = content.splitlines()

            # Get file modification time
            last_modified = datetime.fromtimestamp(filepath.stat().st_mtime)

            for line_num, line in enumerate(lines, start=1):
                matches = self.TAG_PATTERN.findall(line)

                for tag_type, domain in matches:
                    tag_id = domain

                    # Extract context (±2 lines)
                    context_lines = []
                    for i in range(max(0, line_num - 3), min(len(lines), line_num + 2)):
                        if i < len(lines):
                            context_lines.append(lines[i])
                    context = "\n".join(context_lines)

                    # Create inventory item
                    relative_path = str(filepath.relative_to(root))
                    inventory.append(TagInventory(
                        tag_id=tag_id,
                        file_path=relative_path,
                        line_number=line_num,
                        context=context,
                        related_tags=[],  # Will be populated later
                        last_modified=last_modified,
                        status="active"
                    ))

        except Exception:
            pass

        return inventory

    def group_by_domain(self, inventory: List[TagInventory]) -> Dict[str, List[TagInventory]]:
        """Group inventory by domain

        Args:
            inventory: List of TagInventory objects

        Returns:
            Dict mapping domain prefix to list of tags
        """
        grouped: Dict[str, List[TagInventory]] = {}

        for item in inventory:
            # Extract domain prefix (e.g., "AUTH" from "AUTH-LOGIN-001")
            parts = item.tag_id.split("-")
            if parts:
                domain = parts[0]
                if domain not in grouped:
                    grouped[domain] = []
                grouped[domain].append(item)

        return grouped

    def format_as_markdown(self, grouped: Dict[str, List[TagInventory]]) -> str:
        """Format grouped inventory as markdown

        Args:
            grouped: Grouped inventory dict

        Returns:
            Markdown-formatted string
        """
        lines = []
        lines.append("# TAG Inventory")
        lines.append("")
        lines.append(f"Generated: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")

        # Calculate totals
        total_tags = sum(len(tags) for tags in grouped.values())
        lines.append(f"Total TAGs: {total_tags}")
        lines.append("")

        # Group by domain
        lines.append("## By Domain")
        lines.append("")

        for domain in sorted(grouped.keys()):
            lines.append(f"### {domain}")
            lines.append("")

            for item in sorted(grouped[domain], key=lambda x: x.tag_id):
                lines.append(f"- **{item.tag_id}** (`{item.file_path}:{item.line_number}`)")

            lines.append("")

        return "\n".join(lines)


class MatrixGenerator:
    """Generates TAG coverage matrix

    Creates matrix showing SPEC implementation status across
    CODE, TEST, and DOC components.
    """

    def generate_matrix(self, tags: Dict[str, Set[str]]) -> TagMatrix:
        """Generate coverage matrix from tags

        Args:
            tags: Dict mapping type to set of domain IDs
                  {"SPEC": {"AUTH-001"}, "CODE": {"AUTH-001"}, ...}

        Returns:
            TagMatrix object
        """
        matrix = TagMatrix()

        # Get all unique domains
        all_domains = set()
        for tag_set in tags.values():
            all_domains.update(tag_set)

        # Build matrix rows
        for domain in all_domains:
            matrix.rows[domain] = {
                "SPEC": domain in tags.get("SPEC", set()),
                "CODE": domain in tags.get("CODE", set()),
                "TEST": domain in tags.get("TEST", set()),
                "DOC": domain in tags.get("DOC", set())
            }

            # Calculate completion percentage
            matrix.completion_percentages[domain] = self.calculate_completion_percentage(domain, tags)

        return matrix

    def calculate_completion_percentage(self, spec_id: str, tags: Dict[str, Set[str]]) -> float:
        """Calculate completion percentage for a SPEC

        Args:
            spec_id: SPEC domain ID
            tags: Tags dict

        Returns:
            Completion percentage (0-100)
        """
        components = ["SPEC", "CODE", "TEST", "DOC"]
        present = sum(1 for comp in components if spec_id in tags.get(comp, set()))

        return (present / len(components)) * 100.0

    def format_as_markdown_table(self, matrix: TagMatrix) -> str:
        """Format matrix as markdown table

        Args:
            matrix: TagMatrix object

        Returns:
            Markdown table string
        """
        lines = []
        lines.append("# TAG Coverage Matrix")
        lines.append("")
        lines.append(f"Generated: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
        lines.append("")

        # Table header
        lines.append("| SPEC | CODE | TEST | DOC | Completion |")
        lines.append("|------|------|------|-----|------------|")

        # Table rows
        for domain in sorted(matrix.rows.keys()):
            row = matrix.rows[domain]
            spec_mark = "✅" if row["SPEC"] else "❌"
            code_mark = "✅" if row["CODE"] else "❌"
            test_mark = "✅" if row["TEST"] else "❌"
            doc_mark = "✅" if row["DOC"] else "❌"
            completion = f"{matrix.completion_percentages[domain]:.0f}%"

            lines.append(f"| {domain} ({spec_mark}) | {code_mark} | {test_mark} | {doc_mark} | {completion} |")

        lines.append("")

        # Summary
        total_specs = len(matrix.rows)
        fully_implemented = sum(1 for pct in matrix.completion_percentages.values() if pct == 100.0)

        lines.append("## Summary")
        lines.append("")
        lines.append(f"- Total SPECs: {total_specs}")
        lines.append(f"- Fully Implemented (100%): {fully_implemented}")
        lines.append("")

        return "\n".join(lines)

    def format_as_csv(self, matrix: TagMatrix) -> str:
        """Format matrix as CSV

        Args:
            matrix: TagMatrix object

        Returns:
            CSV string
        """
        lines = []
        lines.append("SPEC,CODE,TEST,DOC,Completion")

        for domain in sorted(matrix.rows.keys()):
            row = matrix.rows[domain]
            spec = "1" if row["SPEC"] else "0"
            code = "1" if row["CODE"] else "0"
            test = "1" if row["TEST"] else "0"
            "1" if row["DOC"] else "0"
            completion = f"{matrix.completion_percentages[domain]:.1f}"

            lines.append(f"{domain},{spec},{code},{test},{completion}")

        return "\n".join(lines)


class CoverageAnalyzer:
    """Analyzes TAG coverage and chain integrity

    Analyzes SPEC→CODE→TEST→DOC chains to identify
    coverage gaps and orphan TAGs.
    """

    TAG_PATTERN = re.compile(r"@(SPEC|CODE|TEST|DOC):([A-Z]+(?:-[A-Z]+)*-\d{3})")
    IGNORE_PATTERNS = [".git/*", "node_modules/*", "__pycache__/*", "*.pyc", ".venv/*", "venv/*"]

    def analyze_spec_coverage(self, spec_id: str, root_path: str) -> CoverageMetrics:
        """Analyze coverage for a specific SPEC

        Args:
            spec_id: SPEC domain ID
            root_path: Root directory to scan

        Returns:
            CoverageMetrics object
        """
        tags = self._collect_tags(root_path)

        metrics = CoverageMetrics(spec_id=spec_id)
        metrics.has_code = spec_id in tags.get("CODE", set())
        metrics.has_test = spec_id in tags.get("TEST", set())
        metrics.has_doc = spec_id in tags.get("DOC", set())

        # Calculate coverage percentage
        components = [metrics.has_code, metrics.has_test, metrics.has_doc]
        metrics.coverage_percentage = (sum(components) / 3.0) * 100.0

        return metrics

    def get_specs_without_code(self, root_path: str) -> List[str]:
        """Find SPECs without CODE implementation

        Args:
            root_path: Root directory to scan

        Returns:
            List of SPEC IDs without CODE
        """
        tags = self._collect_tags(root_path)

        specs = tags.get("SPEC", set())
        codes = tags.get("CODE", set())

        return list(specs - codes)

    def get_code_without_tests(self, root_path: str) -> List[str]:
        """Find CODE without TEST

        Args:
            root_path: Root directory to scan

        Returns:
            List of CODE IDs without TEST
        """
        tags = self._collect_tags(root_path)

        codes = tags.get("CODE", set())
        tests = tags.get("TEST", set())

        return list(codes - tests)

    def get_code_without_docs(self, root_path: str) -> List[str]:
        """Find CODE without DOC

        Args:
            root_path: Root directory to scan

        Returns:
            List of CODE IDs without DOC
        """
        tags = self._collect_tags(root_path)

        codes = tags.get("CODE", set())
        docs = tags.get("DOC", set())

        return list(codes - docs)

    def calculate_overall_coverage(self, root_path: str) -> float:
        """Calculate overall coverage percentage

        Args:
            root_path: Root directory to scan

        Returns:
            Overall coverage percentage (0-100)
        """
        tags = self._collect_tags(root_path)

        specs = tags.get("SPEC", set())
        if not specs:
            return 0.0 if tags.get("CODE", set()) else 100.0

        # Calculate average coverage for all SPECs
        total_coverage = 0.0
        for spec_id in specs:
            metrics = self.analyze_spec_coverage(spec_id, root_path)
            total_coverage += metrics.coverage_percentage

        return total_coverage / len(specs)

    def _collect_tags(self, root_path: str) -> Dict[str, Set[str]]:
        """Collect all TAGs from directory

        Args:
            root_path: Root directory to scan

        Returns:
            Dict mapping type to set of domain IDs
        """
        tags: Dict[str, Set[str]] = {
            "SPEC": set(),
            "CODE": set(),
            "TEST": set(),
            "DOC": set()
        }

        root = Path(root_path)
        if not root.exists():
            return tags

        for filepath in root.rglob("*"):
            if not filepath.is_file():
                continue

            # Check ignore patterns
            if self._should_ignore(filepath, root):
                continue

            # Extract tags
            try:
                content = filepath.read_text(encoding="utf-8", errors="ignore")
                matches = self.TAG_PATTERN.findall(content)

                for tag_type, domain in matches:
                    tags[tag_type].add(domain)

            except Exception:
                pass

        return tags

    def _should_ignore(self, filepath: Path, root: Path) -> bool:
        """Check if file should be ignored

        Args:
            filepath: File path
            root: Root directory

        Returns:
            True if should ignore
        """
        try:
            relative = filepath.relative_to(root)
            relative_str = str(relative)

            for pattern in self.IGNORE_PATTERNS:
                pattern_clean = pattern.replace("/*", "").replace("*", "")
                if pattern_clean in relative_str:
                    return True

            return False

        except ValueError:
            return True


class StatisticsGenerator:
    """Generates overall TAG statistics

    Produces aggregated statistics and metrics for TAG system health.
    """

    TAG_PATTERN = re.compile(r"@(SPEC|CODE|TEST|DOC):([A-Z]+(?:-[A-Z]+)*-\d{3})")

    def generate_statistics(self, tags: Dict[str, Set[str]]) -> StatisticsReport:
        """Generate statistics from tags

        Args:
            tags: Dict mapping type to set of domain IDs

        Returns:
            StatisticsReport object
        """
        report = StatisticsReport(
            generated_at=datetime.now(),
            total_tags=0,
            by_type={},
            by_domain={},
            coverage={},
            issues={}
        )

        # Count by type
        for tag_type, domains in tags.items():
            report.by_type[tag_type] = len(domains)
            report.total_tags += len(domains)

        # Count by domain
        all_domains = set()
        for domains in tags.values():
            all_domains.update(domains)

        for domain in all_domains:
            # Extract domain prefix
            parts = domain.split("-")
            if parts:
                domain_prefix = parts[0]
                if domain_prefix not in report.by_domain:
                    report.by_domain[domain_prefix] = 0
                report.by_domain[domain_prefix] += 1

        # Calculate coverage metrics
        specs = tags.get("SPEC", set())
        codes = tags.get("CODE", set())
        tests = tags.get("TEST", set())

        if specs:
            spec_to_code = len(specs & codes) / len(specs) * 100.0
            report.coverage["spec_to_code"] = round(spec_to_code, 2)

        if codes:
            code_to_test = len(codes & tests) / len(codes) * 100.0
            report.coverage["code_to_test"] = round(code_to_test, 2)

        # Calculate overall coverage
        if specs:
            total_coverage = 0.0
            for spec in specs:
                components = 0
                if spec in codes:
                    components += 1
                if spec in tests:
                    components += 1
                if spec in tags.get("DOC", set()):
                    components += 1
                total_coverage += (components / 3.0) * 100.0

            report.coverage["overall_percentage"] = round(total_coverage / len(specs), 2)
        else:
            report.coverage["overall_percentage"] = 0.0

        # Detect issues
        orphan_codes = codes - tests
        orphan_tests = tests - codes
        report.issues["orphan_count"] = len(orphan_codes) + len(orphan_tests)

        incomplete_specs = specs - codes
        incomplete_chains = len(incomplete_specs)
        for spec in specs & codes:
            if spec not in tests:
                incomplete_chains += 1

        report.issues["incomplete_chains"] = incomplete_chains
        report.issues["deprecated_count"] = 0  # Placeholder

        return report

    def format_as_json(self, stats: StatisticsReport) -> str:
        """Format statistics as JSON

        Args:
            stats: StatisticsReport object

        Returns:
            JSON string
        """
        data = {
            "generated_at": stats.generated_at.isoformat(),
            "total_tags": stats.total_tags,
            "by_type": stats.by_type,
            "by_domain": stats.by_domain,
            "coverage": stats.coverage,
            "issues": stats.issues
        }

        return json.dumps(data, indent=2)

    def format_as_human_readable(self, stats: StatisticsReport) -> str:
        """Format statistics as human-readable text

        Args:
            stats: StatisticsReport object

        Returns:
            Human-readable string
        """
        lines = []
        lines.append("# TAG Statistics")
        lines.append("")
        lines.append(f"Generated: {stats.generated_at.strftime('%Y-%m-%d %H:%M:%S')}")
        lines.append("")

        lines.append(f"Total TAGs: {stats.total_tags}")
        lines.append("")

        lines.append("## By Type")
        for tag_type, count in sorted(stats.by_type.items()):
            lines.append(f"- {tag_type}: {count}")
        lines.append("")

        lines.append("## By Domain")
        for domain, count in sorted(stats.by_domain.items()):
            lines.append(f"- {domain}: {count}")
        lines.append("")

        lines.append("## Coverage")
        for metric, value in sorted(stats.coverage.items()):
            lines.append(f"- {metric}: {value}%")
        lines.append("")

        return "\n".join(lines)


class ReportFormatter:
    """Formats reports in multiple output formats

    Provides formatting utilities for inventory, matrix, and statistics
    in Markdown, HTML, CSV, and JSON formats.
    """

    def format_inventory_md(self, inventory: List[TagInventory]) -> str:
        """Format inventory as markdown

        Args:
            inventory: List of TagInventory objects

        Returns:
            Markdown string
        """
        generator = InventoryGenerator()
        grouped = generator.group_by_domain(inventory)
        return generator.format_as_markdown(grouped)

    def format_matrix_md(self, matrix: TagMatrix) -> str:
        """Format matrix as markdown

        Args:
            matrix: TagMatrix object

        Returns:
            Markdown string
        """
        generator = MatrixGenerator()
        return generator.format_as_markdown_table(matrix)

    def format_table(self, headers: List[str], rows: List[List[str]]) -> str:
        """Format data as markdown table

        Args:
            headers: Table headers
            rows: Table rows

        Returns:
            Markdown table string
        """
        lines = []

        # Header row
        lines.append("| " + " | ".join(headers) + " |")

        # Separator row
        lines.append("| " + " | ".join(["---"] * len(headers)) + " |")

        # Data rows
        for row in rows:
            lines.append("| " + " | ".join(row) + " |")

        return "\n".join(lines)

    def format_html_dashboard(self, inventory: List[TagInventory]) -> str:
        """Format inventory as HTML dashboard (OPTIONAL)

        Args:
            inventory: List of TagInventory objects

        Returns:
            HTML string

        Raises:
            NotImplementedError: HTML formatting is optional
        """
        raise NotImplementedError("HTML dashboard formatting is optional")


class ReportGenerator:
    """Main orchestrator for report generation

    Coordinates all generators to produce complete reporting suite:
    - tag-inventory.md
    - tag-matrix.md
    - tag-statistics.json
    """

    def __init__(self):
        """Initialize report generator"""
        self.inventory_gen = InventoryGenerator()
        self.matrix_gen = MatrixGenerator()
        self.coverage_analyzer = CoverageAnalyzer()
        self.stats_gen = StatisticsGenerator()
        self.formatter = ReportFormatter()

    def generate_inventory_report(self, root_path: str) -> str:
        """Generate inventory report

        Args:
            root_path: Root directory to scan

        Returns:
            Markdown inventory report
        """
        inventory = self.inventory_gen.generate_inventory(root_path)
        return self.formatter.format_inventory_md(inventory)

    def generate_matrix_report(self, root_path: str) -> str:
        """Generate coverage matrix report

        Args:
            root_path: Root directory to scan

        Returns:
            Markdown matrix report
        """
        tags = self.coverage_analyzer._collect_tags(root_path)
        matrix = self.matrix_gen.generate_matrix(tags)
        return self.formatter.format_matrix_md(matrix)

    def generate_statistics_report(self, root_path: str) -> str:
        """Generate statistics report

        Args:
            root_path: Root directory to scan

        Returns:
            JSON statistics report
        """
        tags = self.coverage_analyzer._collect_tags(root_path)
        stats = self.stats_gen.generate_statistics(tags)
        return self.stats_gen.format_as_json(stats)

    def generate_combined_report(self, root_path: str) -> str:
        """Generate combined report with all sections

        Args:
            root_path: Root directory to scan

        Returns:
            Combined markdown report
        """
        lines = []
        lines.append("# MoAI-ADK TAG System Report")
        lines.append("")
        lines.append(f"Generated: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
        lines.append("")

        # Inventory section
        lines.append("---")
        lines.append("")
        lines.append(self.generate_inventory_report(root_path))
        lines.append("")

        # Matrix section
        lines.append("---")
        lines.append("")
        lines.append(self.generate_matrix_report(root_path))
        lines.append("")

        # Statistics section
        lines.append("---")
        lines.append("")
        lines.append("# Statistics")
        lines.append("")
        lines.append("```json")
        lines.append(self.generate_statistics_report(root_path))
        lines.append("```")
        lines.append("")

        return "\n".join(lines)

    def generate_all_reports(self, root_path: str, output_dir: str) -> ReportResult:
        """Generate all reports and save to output directory

        Args:
            root_path: Root directory to scan
            output_dir: Output directory for reports

        Returns:
            ReportResult with file paths
        """
        output = Path(output_dir)
        output.mkdir(parents=True, exist_ok=True)

        try:
            # Generate inventory
            inventory_path = output / "tag-inventory.md"
            inventory_report = self.generate_inventory_report(root_path)
            inventory_path.write_text(inventory_report, encoding="utf-8")

            # Generate matrix
            matrix_path = output / "tag-matrix.md"
            matrix_report = self.generate_matrix_report(root_path)
            matrix_path.write_text(matrix_report, encoding="utf-8")

            # Generate statistics
            statistics_path = output / "tag-statistics.json"
            statistics_report = self.generate_statistics_report(root_path)
            statistics_path.write_text(statistics_report, encoding="utf-8")

            return ReportResult(
                inventory_path=inventory_path,
                matrix_path=matrix_path,
                statistics_path=statistics_path,
                success=True
            )

        except Exception as e:
            return ReportResult(
                inventory_path=output / "tag-inventory.md",
                matrix_path=output / "tag-matrix.md",
                statistics_path=output / "tag-statistics.json",
                success=False,
                error_message=str(e)
            )
