#!/usr/bin/env python3
"""
Translation Status Checker

This script analyzes the translation coverage for MoAI-ADK documentation
across multiple languages (English, Japanese, Chinese).

It compares translated files against the base Korean documentation
and generates a comprehensive status report.
"""

import json
import os
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Set


class TranslationStatusChecker:
    """Check translation status across multiple languages."""

    def __init__(self, docs_src_dir: str = "src"):
        """
        Initialize the translation status checker.

        Args:
            docs_src_dir: Path to the documentation source directory
        """
        self.docs_src = Path(__file__).parent.parent / docs_src_dir
        self.languages = ["en", "ja", "zh"]
        self.stats: Dict[str, dict] = {}
        self.base_files: Set[str] = set()
        self.missing_files: Dict[str, List[str]] = {}

    def get_markdown_files(self, directory: Path) -> Set[str]:
        """
        Get all markdown files in a directory recursively.

        Args:
            directory: Directory to search

        Returns:
            Set of relative file paths
        """
        if not directory.exists():
            return set()

        md_files = set()
        for md_file in directory.rglob("*.md"):
            # Get relative path from the language directory
            rel_path = md_file.relative_to(directory)
            md_files.add(str(rel_path))

        return md_files

    def get_base_files(self) -> Set[str]:
        """
        Get all Korean (base) documentation files.

        Returns:
            Set of base file paths
        """
        base_files = set()

        # Get files from root docs/src directory (Korean files)
        for md_file in self.docs_src.rglob("*.md"):
            # Skip language-specific directories
            relative_path = md_file.relative_to(self.docs_src)
            parts = relative_path.parts

            if parts and parts[0] not in self.languages:
                base_files.add(str(relative_path))

        return base_files

    def analyze_language(self, lang: str) -> dict:
        """
        Analyze translation status for a specific language.

        Args:
            lang: Language code (en, ja, zh)

        Returns:
            Dictionary with translation statistics
        """
        lang_dir = self.docs_src / lang
        translated_files = self.get_markdown_files(lang_dir)

        total = len(self.base_files)
        translated = len(translated_files)
        completion = (translated / total * 100) if total > 0 else 0

        # Find missing files
        missing = []
        for base_file in self.base_files:
            if base_file not in translated_files:
                missing.append(base_file)

        self.missing_files[lang] = sorted(missing)

        return {
            "total": total,
            "translated": translated,
            "completion": round(completion, 2),
            "missing_count": len(missing),
        }

    def analyze_all(self) -> Dict[str, dict]:
        """
        Analyze translation status for all languages.

        Returns:
            Dictionary with statistics for all languages
        """
        self.base_files = self.get_base_files()
        print(f"Found {len(self.base_files)} base Korean documentation files")

        for lang in self.languages:
            print(f"Analyzing {lang} translations...")
            self.stats[lang] = self.analyze_language(lang)

        return self.stats

    def generate_json_report(self, output_path: str = ".moai/reports/translation-status.json"):
        """
        Generate JSON report file.

        Args:
            output_path: Path to save the JSON report
        """
        report = {
            "updated_at": datetime.now().isoformat(),
            "base_language": "ko",
            "base_files_count": len(self.base_files),
            "languages": self.stats,
            "missing_files": self.missing_files,
        }

        # Create directory if it doesn't exist
        output_file = Path(output_path)
        output_file.parent.mkdir(parents=True, exist_ok=True)

        with open(output_file, "w", encoding="utf-8") as f:
            json.dump(report, f, indent=2, ensure_ascii=False)

        print(f"\nJSON report saved to: {output_file}")

    def generate_markdown_dashboard(self, output_path: str = "translation-status.md"):
        """
        Generate Markdown dashboard.

        Args:
            output_path: Path to save the Markdown dashboard (relative to docs/src/)
        """
        lines = [
            "# Translation Status Dashboard",
            "",
            f"Last Updated: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}",
            "",
            "## Overview",
            "",
            f"Base Language: **Korean (ko)**",
            f"Total Documentation Files: **{len(self.base_files)}**",
            "",
            "## Translation Progress",
            "",
        ]

        # Progress table
        lines.extend([
            "| Language | Translated | Total | Completion | Missing |",
            "|----------|------------|-------|------------|---------|",
        ])

        for lang in self.languages:
            stats = self.stats[lang]
            lang_name = {"en": "English", "ja": "Japanese", "zh": "Chinese"}[lang]
            completion = stats["completion"]
            progress_bar = self._generate_progress_bar(completion)

            lines.append(
                f"| {lang_name} ({lang}) | {stats['translated']} | "
                f"{stats['total']} | {progress_bar} {completion}% | {stats['missing_count']} |"
            )

        lines.extend(["", "## Missing Files by Language", ""])

        # Missing files for each language
        for lang in self.languages:
            lang_name = {"en": "English", "ja": "Japanese", "zh": "Chinese"}[lang]
            missing = self.missing_files[lang]

            lines.extend([
                f"### {lang_name} ({lang})",
                "",
                f"Missing: {len(missing)} files",
                "",
            ])

            if missing:
                lines.append("```")
                for file in missing[:10]:  # Show first 10
                    lines.append(file)
                if len(missing) > 10:
                    lines.append(f"... and {len(missing) - 10} more files")
                lines.append("```")
            else:
                lines.append("✅ All files translated!")

            lines.append("")

        lines.extend([
            "## How to Contribute",
            "",
            "See [Translation Contributing Guide](guides/contributing-translations.md) for details.",
            "",
            "---",
            "",
            "*This dashboard is automatically updated by CI/CD pipeline.*",
        ])

        # Save file
        output_file = self.docs_src / output_path
        output_file.parent.mkdir(parents=True, exist_ok=True)

        with open(output_file, "w", encoding="utf-8") as f:
            f.write("\n".join(lines))

        print(f"Markdown dashboard saved to: {output_file}")

    @staticmethod
    def _generate_progress_bar(percentage: float, width: int = 10) -> str:
        """
        Generate a text-based progress bar.

        Args:
            percentage: Completion percentage (0-100)
            width: Width of the progress bar

        Returns:
            Progress bar string
        """
        filled = int(percentage / 100 * width)
        bar = "█" * filled + "░" * (width - filled)
        return f"{bar}"

    def print_summary(self):
        """Print summary to console."""
        print("\n" + "=" * 60)
        print("TRANSLATION STATUS SUMMARY")
        print("=" * 60)
        print(f"\nBase Files (Korean): {len(self.base_files)}")

        for lang in self.languages:
            stats = self.stats[lang]
            lang_name = {"en": "English", "ja": "Japanese", "zh": "Chinese"}[lang]
            print(f"\n{lang_name} ({lang}):")
            print(f"  Translated: {stats['translated']}/{stats['total']}")
            print(f"  Completion: {stats['completion']}%")
            print(f"  Missing: {stats['missing_count']} files")

        print("\n" + "=" * 60)


def main():
    """Main execution function."""
    # Change to docs directory
    docs_dir = Path(__file__).parent.parent
    os.chdir(docs_dir)

    print("Starting translation status check...\n")

    checker = TranslationStatusChecker()
    checker.analyze_all()
    checker.print_summary()

    # Generate reports
    print("\nGenerating reports...")
    checker.generate_json_report()
    checker.generate_markdown_dashboard()

    print("\n✅ Translation status check completed!")


if __name__ == "__main__":
    main()
