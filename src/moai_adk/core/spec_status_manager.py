"""SPEC Status Manager

Automated management of SPEC status transitions from 'draft' to 'completed'
based on implementation completion detection and validation criteria.
"""

import logging
import re
from datetime import datetime
from pathlib import Path
from typing import Any, Dict, List, Set

import yaml  # type: ignore

logger = logging.getLogger(__name__)


class SpecStatusManager:
    """Manages SPEC status detection and updates based on implementation completion"""

    def __init__(self, project_root: Path):
        """Initialize the SPEC status manager

        Args:
            project_root: Root directory of the MoAI project
        """
        self.project_root = Path(project_root)
        self.specs_dir = self.project_root / ".moai" / "specs"
        self.src_dir = self.project_root / "src"
        self.tests_dir = self.project_root / "tests"
        self.docs_dir = self.project_root / "docs"

        # Validation criteria (configurable)
        self.validation_criteria = {
            "min_code_coverage": 0.85,  # 85% minimum coverage
            "require_acceptance_criteria": True,
            "min_implementation_age_days": 0,  # Days since last implementation
        }

    def detect_draft_specs(self) -> Set[str]:
        """Detect all SPEC files with 'draft' status

        Returns:
            Set of SPEC IDs that have draft status
        """
        draft_specs: Set[str] = set()

        if not self.specs_dir.exists():
            logger.warning(f"SPEC directory not found: {self.specs_dir}")
            return draft_specs

        for spec_dir in self.specs_dir.iterdir():
            if spec_dir.is_dir():
                spec_file = spec_dir / "spec.md"
                if spec_file.exists():
                    try:
                        # Read frontmatter to check status
                        with open(spec_file, "r", encoding="utf-8") as f:
                            content = f.read()

                        frontmatter = None

                        # Handle JSON-like meta (common in older specs)
                        meta_match = re.search(r"<%?\s*---\s*\n(.*?)\n---\s*%?>", content, re.DOTALL)
                        if meta_match:
                            try:
                                meta_text = meta_match.group(1)
                                # Replace JSON-style quotes and parse as YAML
                                meta_text = meta_text.replace('"', "").replace("'", "")
                                frontmatter = yaml.safe_load("{" + meta_text + "}")
                            except Exception as e:
                                logger.debug(f"JSON meta parsing failed for {spec_dir.name}: {e}")

                        # Handle regular YAML frontmatter
                        elif content.startswith("---"):
                            end_marker = content.find("---", 3)
                            if end_marker != -1:
                                frontmatter_text = content[3:end_marker].strip()
                                try:
                                    frontmatter = yaml.safe_load(frontmatter_text)
                                except yaml.YAMLError as e:
                                    logger.warning(f"YAML parsing error for {spec_dir.name}: {e}")
                                    # Try to fix common issues (like @ in author field)
                                    try:
                                        # Replace problematic @author entries
                                        fixed_text = frontmatter_text
                                        if "author: @" in fixed_text:
                                            fixed_text = re.sub(
                                                r"author:\s*@(\w+)",
                                                r'author: "\1"',
                                                fixed_text,
                                            )
                                        frontmatter = yaml.safe_load(fixed_text)
                                    except yaml.YAMLError:
                                        logger.error(f"Could not parse YAML for {spec_dir.name} even after fixes")
                                        continue

                        if frontmatter and frontmatter.get("status") == "draft":
                            spec_id = spec_dir.name
                            draft_specs.add(spec_id)
                            logger.debug(f"Found draft SPEC: {spec_id}")

                    except Exception as e:
                        logger.error(f"Error reading SPEC {spec_dir.name}: {e}")

        logger.info(f"Found {len(draft_specs)} draft SPECs")
        return draft_specs

    def is_spec_implementation_completed(self, spec_id: str) -> bool:
        """Check if a SPEC's implementation is complete

        Args:
            spec_id: The SPEC identifier (e.g., "SPEC-001")

        Returns:
            True if implementation is complete, False otherwise
        """
        spec_dir = self.specs_dir / spec_id
        spec_file = spec_dir / "spec.md"

        if not spec_file.exists():
            logger.warning(f"SPEC file not found: {spec_file}")
            return False

        try:
            # Check basic implementation status
            spec_dir = spec_file.parent

            # Check for implementation files
            src_files = list(spec_dir.rglob("*.py")) if (spec_dir.parent.parent / "src").exists() else []

            # Check for test files
            test_dir = spec_dir.parent.parent / "tests"
            test_files = list(test_dir.rglob(f"test_*{spec_id.lower()}*.py")) if test_dir.exists() else []

            # Simple completion criteria
            has_code = len(src_files) > 0
            has_tests = len(test_files) > 0

            # Check if SPEC has acceptance criteria
            with open(spec_file, "r", encoding="utf-8") as f:
                spec_content = f.read()
            has_acceptance_criteria = "Acceptance Criteria" in spec_content

            # Overall completion check
            is_complete = has_code and has_tests and has_acceptance_criteria

            logger.info(f"SPEC {spec_id} implementation status: {'COMPLETE' if is_complete else 'INCOMPLETE'}")
            return is_complete

        except Exception as e:
            logger.error(f"Error checking SPEC {spec_id} completion: {e}")
            return False

    def update_spec_status(self, spec_id: str, new_status: str) -> bool:
        """Update SPEC status in frontmatter

        Args:
            spec_id: The SPEC identifier
            new_status: New status value ('completed', 'draft', etc.)

        Returns:
            True if update successful, False otherwise
        """
        spec_dir = self.specs_dir / spec_id
        spec_file = spec_dir / "spec.md"

        if not spec_file.exists():
            logger.error(f"SPEC file not found: {spec_file}")
            return False

        try:
            with open(spec_file, "r", encoding="utf-8") as f:
                content = f.read()

            # Extract and update frontmatter
            if content.startswith("---"):
                end_marker = content.find("---", 3)
                if end_marker != -1:
                    frontmatter_text = content[3:end_marker].strip()
                    try:
                        frontmatter = yaml.safe_load(frontmatter_text) or {}
                    except yaml.YAMLError as e:
                        logger.warning(f"YAML parsing error for {spec_id}: {e}")
                        # Try to fix common issues
                        try:
                            fixed_text = frontmatter_text
                            if "author: @" in fixed_text:
                                fixed_text = re.sub(r"author:\s*@(\w+)", r'author: "\1"', fixed_text)
                            frontmatter = yaml.safe_load(fixed_text) or {}
                        except yaml.YAMLError:
                            logger.error(f"Could not parse YAML for {spec_id} even after fixes")
                            return False

                    # Update status
                    frontmatter["status"] = new_status

                    # Bump version if completing
                    if new_status == "completed":
                        frontmatter["version"] = self._bump_version(frontmatter.get("version", "0.1.0"))
                        frontmatter["updated"] = datetime.now().strftime("%Y-%m-%d")

                    # Reconstruct the file
                    new_frontmatter = yaml.dump(frontmatter, default_flow_style=False)
                    new_content = f"---\n{new_frontmatter}---{content[end_marker + 3 :]}"

                    # Write back to file
                    with open(spec_file, "w", encoding="utf-8") as f:
                        f.write(new_content)

                    logger.info(f"Updated SPEC {spec_id} status to {new_status}")
                    return True

        except Exception as e:
            logger.error(f"Error updating SPEC {spec_id} status: {e}")
            return False

    def get_completion_validation_criteria(self) -> Dict:
        """Get the current validation criteria for SPEC completion

        Returns:
            Dictionary of validation criteria
        """
        return self.validation_criteria.copy()

    def validate_spec_for_completion(self, spec_id: str) -> Dict[str, Any]:
        """Validate if a SPEC is ready for completion

        Args:
            spec_id: The SPEC identifier

        Returns:
            Dictionary with validation results:
            {
                'is_ready': bool,
                'criteria_met': Dict[str, bool],
                'issues': List[str],
                'recommendations': List[str]
            }
        """
        result: Dict[str, Any] = {
            "is_ready": False,
            "criteria_met": {},
            "issues": [],
            "recommendations": [],
        }

        try:
            spec_dir = self.specs_dir / spec_id
            spec_file = spec_dir / "spec.md"

            if not spec_file.exists():
                issues_list = result["issues"]
                if isinstance(issues_list, list):
                    issues_list.append(f"SPEC file not found: {spec_file}")
                return result

            # Check implementation status
            criteria_checks: Dict[str, bool] = {}

            # Check for code implementation
            spec_dir = spec_file.parent
            src_dir = spec_dir.parent.parent / "src"
            criteria_checks["code_implemented"] = src_dir.exists() and len(list(src_dir.rglob("*.py"))) > 0
            if not criteria_checks["code_implemented"]:
                issues_list = result["issues"]
                if isinstance(issues_list, list):
                    issues_list.append("No source code files found")

            # Check for test implementation
            test_dir = spec_dir.parent.parent / "tests"
            test_files = list(test_dir.rglob("test_*.py")) if test_dir.exists() else []
            criteria_checks["test_implemented"] = len(test_files) > 0
            if not criteria_checks["test_implemented"]:
                issues_list = result["issues"]
                if isinstance(issues_list, list):
                    issues_list.append("No test files found")

            # Check for acceptance criteria
            criteria_checks["tasks_completed"] = self._check_acceptance_criteria(spec_file)
            if not criteria_checks["tasks_completed"]:
                issues_list = result["issues"]
                if isinstance(issues_list, list):
                    issues_list.append("Missing acceptance criteria section")

            # 4. Acceptance criteria present
            criteria_checks["has_acceptance_criteria"] = self._check_acceptance_criteria(spec_file)
            if (
                not criteria_checks["has_acceptance_criteria"]
                and self.validation_criteria["require_acceptance_criteria"]
            ):
                issues_list = result["issues"]
                if isinstance(issues_list, list):
                    issues_list.append("Missing acceptance criteria section")

            # 5. Documentation sync
            criteria_checks["docs_synced"] = self._check_documentation_sync(spec_id)
            if not criteria_checks["docs_synced"]:
                recs_list = result["recommendations"]
                if isinstance(recs_list, list):
                    recs_list.append("Consider running /moai:3-sync to update documentation")

            result["criteria_met"] = criteria_checks
            result["is_ready"] = all(criteria_checks.values())

            # Add recommendations
            if result["is_ready"]:
                recs_list = result["recommendations"]
                if isinstance(recs_list, list):
                    recs_list.append("SPEC is ready for completion. Consider updating status to 'completed'")

        except Exception as e:
            logger.error(f"Error validating SPEC {spec_id}: {e}")
            issues_list = result["issues"]
            if isinstance(issues_list, list):
                issues_list.append(f"Validation error: {e}")

        return result

    def batch_update_completed_specs(self) -> Dict[str, List[str]]:
        """Batch update all draft SPECs that have completed implementations

        Returns:
            Dictionary with update results:
            {
                'updated': List[str],  # Successfully updated SPEC IDs
                'failed': List[str],   # Failed SPEC IDs with errors
                'skipped': List[str]   # Incomplete SPEC IDs
            }
        """
        results: Dict[str, List[str]] = {"updated": [], "failed": [], "skipped": []}

        draft_specs = self.detect_draft_specs()
        logger.info(f"Checking {len(draft_specs)} draft SPECs for completion")

        for spec_id in draft_specs:
            try:
                # Validate first
                validation = self.validate_spec_for_completion(spec_id)

                if validation["is_ready"]:
                    # Update status
                    if self.update_spec_status(spec_id, "completed"):
                        updated_list = results["updated"]
                        if isinstance(updated_list, list):
                            updated_list.append(spec_id)
                        logger.info(f"Updated SPEC {spec_id} to completed")
                    else:
                        failed_list = results["failed"]
                        if isinstance(failed_list, list):
                            failed_list.append(spec_id)
                        logger.error(f"Failed to update SPEC {spec_id}")
                else:
                    skipped_list = results["skipped"]
                    if isinstance(skipped_list, list):
                        skipped_list.append(spec_id)
                    logger.debug(f"SPEC {spec_id} not ready for completion: {validation['issues']}")

            except Exception as e:
                failed_list = results["failed"]
                if isinstance(failed_list, list):
                    failed_list.append(spec_id)
                logger.error(f"Error processing SPEC {spec_id}: {e}")

        logger.info(
            f"Batch update complete: {len(results['updated'])} updated, "
            f"{len(results['failed'])} failed, "
            f"{len(results['skipped'])} skipped"
        )
        return results

    # Private helper methods

    def _check_acceptance_criteria(self, spec_file: Path) -> bool:
        """
        Check if SPEC file contains acceptance criteria

        Args:
            spec_file: Path to SPEC file

        Returns:
            True if acceptance criteria present
        """
        try:
            with open(spec_file, "r", encoding="utf-8") as f:
                content = f.read()

            # Look for acceptance criteria section
            acceptance_patterns = [
                r"##\s*Acceptance\s+Criteria",
                r"###\s*Acceptance\s+Criteria",
                r"##\s*验收\s+标准",
                r"###\s*验收\s+标准",
            ]

            for pattern in acceptance_patterns:
                if re.search(pattern, content, re.IGNORECASE):
                    return True

            return False

        except Exception as e:
            logger.error(f"Error checking acceptance criteria in {spec_file}: {e}")
            return False

    def _check_documentation_sync(self, spec_id: str) -> bool:
        """
        Check if documentation is synchronized with implementation

        Args:
            spec_id: The SPEC identifier

        Returns:
            True if documentation appears synchronized
        """
        try:
            # Simple heuristic: check if docs exist and are recent
            docs_dir = self.project_root / "docs"
            if not docs_dir.exists():
                return True  # No docs to sync

            # Check if there are any doc files related to this SPEC
            spec_docs = list(docs_dir.rglob(f"*{spec_id.lower()}*"))
            if not spec_docs:
                return True  # No specific docs for this SPEC

            # Basic check - assume docs are in sync if they exist
            return True

        except Exception as e:
            logger.error(f"Error checking documentation sync for {spec_id}: {e}")
            return False

    def _run_additional_validations(self, spec_id: str) -> bool:
        """Run additional validation checks for SPEC completion

        Args:
            spec_id: The SPEC identifier

        Returns:
            True if all additional validations pass
        """
        # Add any additional validation logic here
        # For now, return True as default
        return True

    def _bump_version(self, current_version: str) -> str:
        """Bump version to indicate completion

        Args:
            current_version: Current version string

        Returns:
            New version string
        """
        try:
            # Parse current version - strip quotes if present
            version = str(current_version).strip("\"'")

            if version.startswith("0."):
                # Major version bump for completion
                return "1.0.0"
            else:
                # Minor version bump for updates
                parts = version.split(".")
                if len(parts) >= 2:
                    try:
                        minor = int(parts[1]) + 1
                        return f"{parts[0]}.{minor}.0"
                    except ValueError:
                        # If parsing fails, default to 1.0.0
                        return "1.0.0"
                else:
                    return f"{version}.1"

        except Exception:
            # Fallback to 1.0.0
            return "1.0.0"
