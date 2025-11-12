"""SPEC Status Manager

Automated management of SPEC status transitions from 'draft' to 'completed'
based on implementation completion detection and validation criteria.
"""

import os
import re
import yaml
from pathlib import Path
from typing import Dict, List, Set, Optional, Tuple
import logging
from datetime import datetime

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
            "require_all_tags": True,    # All @CODE and @TEST tags must be implemented
            "max_open_tasks": 0,         # No remaining @TASK tags
            "require_acceptance_criteria": True,
            "min_implementation_age_days": 0  # Days since last implementation
        }

    def detect_draft_specs(self) -> Set[str]:
        """Detect all SPEC files with 'draft' status

        Returns:
            Set of SPEC IDs that have draft status
        """
        draft_specs = set()

        if not self.specs_dir.exists():
            logger.warning(f"SPEC directory not found: {self.specs_dir}")
            return draft_specs

        for spec_dir in self.specs_dir.iterdir():
            if spec_dir.is_dir():
                spec_file = spec_dir / "spec.md"
                if spec_file.exists():
                    try:
                        # Read frontmatter to check status
                        with open(spec_file, 'r', encoding='utf-8') as f:
                            content = f.read()

                        # Extract frontmatter (YAML or @META format)
                        frontmatter = None

                            # Parse @META format (JSON-like)
                            if meta_match:
                                try:
                                    meta_text = meta_match.group(1)
                                    # Replace JSON-style quotes and parse as YAML
                                    meta_text = meta_text.replace('"', '').replace("'", '')
                                    frontmatter = yaml.safe_load('{' + meta_text + '}')
                                except Exception as e:
                                    logger.warning(f"Could not parse @META format for {spec_dir.name}: {e}")

                        # Handle regular YAML frontmatter
                        elif content.startswith('---'):
                            end_marker = content.find('---', 3)
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
                                        if 'author: @' in fixed_text:
                                            fixed_text = re.sub(r'author:\s*@(\w+)', r'author: "\1"', fixed_text)
                                        frontmatter = yaml.safe_load(fixed_text)
                                    except:
                                        logger.error(f"Could not parse YAML for {spec_dir.name} even after fixes")
                                        continue

                        if frontmatter and frontmatter.get('status') == 'draft':
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
            # Parse SPEC content for required TAGs
            required_tags = self._extract_required_tags(spec_file)

            # Check implementation status for each tag category
            code_tags = required_tags.get('CODE', set())
            test_tags = required_tags.get('TEST', set())
            task_tags = required_tags.get('TASK', set())

            # Verify all @CODE tags are implemented
            code_implemented = self._verify_code_tags_implementation(code_tags)

            # Verify all @TEST tags are implemented
            test_implemented = self._verify_test_tags_implementation(test_tags)

            # Verify no remaining @TASK tags (unless allowed)
            tasks_completed = len(task_tags) <= self.validation_criteria['max_open_tasks']

            # Additional validation checks
            validation_passed = self._run_additional_validations(spec_id)

            # Overall completion check
            is_complete = (
                code_implemented and
                test_implemented and
                tasks_completed and
                validation_passed
            )

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
            with open(spec_file, 'r', encoding='utf-8') as f:
                content = f.read()

            # Extract and update frontmatter
            if content.startswith('---'):
                end_marker = content.find('---', 3)
                if end_marker != -1:
                    frontmatter_text = content[3:end_marker].strip()
                    try:
                        frontmatter = yaml.safe_load(frontmatter_text) or {}
                    except yaml.YAMLError as e:
                        logger.warning(f"YAML parsing error for {spec_id}: {e}")
                        # Try to fix common issues
                        try:
                            fixed_text = frontmatter_text
                            if 'author: @' in fixed_text:
                                fixed_text = re.sub(r'author:\s*@(\w+)', r'author: "\1"', fixed_text)
                            frontmatter = yaml.safe_load(fixed_text) or {}
                        except:
                            logger.error(f"Could not parse YAML for {spec_id} even after fixes")
                            return False

                    # Update status
                    frontmatter['status'] = new_status

                    # Bump version if completing
                    if new_status == 'completed':
                        frontmatter['version'] = self._bump_version(frontmatter.get('version', '0.1.0'))
                        frontmatter['updated'] = datetime.now().strftime('%Y-%m-%d')

                    # Reconstruct the file
                    new_frontmatter = yaml.dump(frontmatter, default_flow_style=False)
                    new_content = f"---\n{new_frontmatter}---{content[end_marker+3:]}"

                    # Write back to file
                    with open(spec_file, 'w', encoding='utf-8') as f:
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

    def validate_spec_for_completion(self, spec_id: str) -> Dict:
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
        result = {
            'is_ready': False,
            'criteria_met': {},
            'issues': [],
            'recommendations': []
        }

        try:
            spec_dir = self.specs_dir / spec_id
            spec_file = spec_dir / "spec.md"

            if not spec_file.exists():
                result['issues'].append(f"SPEC file not found: {spec_file}")
                return result

            # Extract required TAGs
            required_tags = self._extract_required_tags(spec_file)

            # Check each validation criterion
            criteria_checks = {}

            # 1. All @CODE tags implemented
            code_tags = required_tags.get('CODE', set())
            criteria_checks['code_implemented'] = self._verify_code_tags_implementation(code_tags)
            if not criteria_checks['code_implemented']:
                result['issues'].append(f"Missing {len(code_tags)} @CODE tag implementations")

            # 2. All @TEST tags implemented
            test_tags = required_tags.get('TEST', set())
            criteria_checks['test_implemented'] = self._verify_test_tags_implementation(test_tags)
            if not criteria_checks['test_implemented']:
                result['issues'].append(f"Missing {len(test_tags)} @TEST tag implementations")

            # 3. No remaining @TASK tags
            task_tags = required_tags.get('TASK', set())
            criteria_checks['tasks_completed'] = len(task_tags) <= self.validation_criteria['max_open_tasks']
            if not criteria_checks['tasks_completed']:
                result['issues'].append(f"{len(task_tags)} @TASK tags remaining")

            # 4. Acceptance criteria present
            criteria_checks['has_acceptance_criteria'] = self._check_acceptance_criteria(spec_file)
            if not criteria_checks['has_acceptance_criteria'] and self.validation_criteria['require_acceptance_criteria']:
                result['issues'].append("Missing acceptance criteria section")

            # 5. Documentation sync
            criteria_checks['docs_synced'] = self._check_documentation_sync(spec_id)
            if not criteria_checks['docs_synced']:
                result['recommendations'].append("Consider running /alfred:3-sync to update documentation")

            result['criteria_met'] = criteria_checks
            result['is_ready'] = all(criteria_checks.values())

            # Add recommendations
            if result['is_ready']:
                result['recommendations'].append("SPEC is ready for completion. Consider updating status to 'completed'")

        except Exception as e:
            logger.error(f"Error validating SPEC {spec_id}: {e}")
            result['issues'].append(f"Validation error: {e}")

        return result

    def batch_update_completed_specs(self) -> Dict:
        """Batch update all draft SPECs that have completed implementations

        Returns:
            Dictionary with update results:
            {
                'updated': List[str],  # Successfully updated SPEC IDs
                'failed': List[str],   # Failed SPEC IDs with errors
                'skipped': List[str]   # Incomplete SPEC IDs
            }
        """
        results = {
            'updated': [],
            'failed': [],
            'skipped': []
        }

        draft_specs = self.detect_draft_specs()
        logger.info(f"Checking {len(draft_specs)} draft SPECs for completion")

        for spec_id in draft_specs:
            try:
                # Validate first
                validation = self.validate_spec_for_completion(spec_id)

                if validation['is_ready']:
                    # Update status
                    if self.update_spec_status(spec_id, 'completed'):
                        results['updated'].append(spec_id)
                        logger.info(f"Updated SPEC {spec_id} to completed")
                    else:
                        results['failed'].append(spec_id)
                        logger.error(f"Failed to update SPEC {spec_id}")
                else:
                    results['skipped'].append(spec_id)
                    logger.debug(f"SPEC {spec_id} not ready for completion: {validation['issues']}")

            except Exception as e:
                results['failed'].append(spec_id)
                logger.error(f"Error processing SPEC {spec_id}: {e}")

        logger.info(f"Batch update complete: {len(results['updated'])} updated, {len(results['failed'])} failed, {len(results['skipped'])} skipped")
        return results

    # Private helper methods

    def _extract_required_tags(self, spec_file: Path) -> Dict[str, Set[str]]:
        """Extract required TAGs from SPEC file

        Args:
            spec_file: Path to SPEC file

        Returns:
            Dictionary of tag categories and their IDs
        """
        required_tags = {
            'CODE': set(),
            'TEST': set(),
            'TASK': set(),
            'REQ': set(),
            'DESIGN': set()
        }

        try:
            with open(spec_file, 'r', encoding='utf-8') as f:
                content = f.read()

            # Find all TAG patterns
            tag_pattern = r'@(CODE|TEST|TASK|REQ|DESIGN):([A-Z0-9\-]+)'
            matches = re.findall(tag_pattern, content)

            for tag_type, tag_id in matches:
                if tag_type in required_tags:
                    required_tags[tag_type].add(tag_id)

        except Exception as e:
            logger.error(f"Error extracting tags from {spec_file}: {e}")

        return required_tags

    def _verify_code_tags_implementation(self, code_tags: Set[str]) -> bool:
        """Verify that all @CODE tags have corresponding implementations

        Args:
            code_tags: Set of CODE tag IDs

        Returns:
            True if all tags are implemented
        """
        if not code_tags:
            return True

        # Search through source files for CODE tags
        implemented_tags = set()

        try:
            for code_file in self.src_dir.rglob("*.py"):
                try:
                    with open(code_file, 'r', encoding='utf-8') as f:
                        lines = f.readlines()

                    for line_num, line in enumerate(lines, 1):
                        for tag_id in code_tags:
                            # Check if tag is present in a non-comment line
                                # Remove common comment markers and check if tag is still present
                                stripped_line = line.strip()
                                comment_free = stripped_line.replace('#', '').replace('//', '')

                                # Tag should be actual implementation, not just a comment
                                    # Further check - tag should be followed by actual code
                                    # Look for function definitions, class definitions, or significant code blocks
                                    tag_line_index = lines.index(line) if line in lines else -1
                                    if tag_line_index >= 0:
                                        # Look at next few lines for actual implementation
                                        found_implementation = False
                                        for i in range(tag_line_index, min(tag_line_index + 10, len(lines))):
                                            next_line = lines[i].strip()
                                            # Skip empty lines and pure comments
                                            if next_line and not next_line.startswith('#'):
                                                # Check if this looks like actual code
                                                if any(keyword in next_line for keyword in ['def ', 'class ', 'async def', 'return', 'import ', 'from ']):
                                                    found_implementation = True
                                                    break
                                                # Or if it has significant non-comment content
                                                if len(next_line) > 20 and not next_line.startswith('"""'):
                                                    found_implementation = True
                                                    break

                                        if found_implementation:
                                            implemented_tags.add(tag_id)
                                            logger.debug(f"Found implemented CODE tag {tag_id} in {code_file}:{line_num}")
                                        else:
                                            logger.debug(f"Found CODE tag {tag_id} but no implementation in {code_file}:{line_num}")

                except Exception as e:
                    logger.warning(f"Error reading code file {code_file}: {e}")

        except Exception as e:
            logger.error(f"Error scanning source directory: {e}")

        # Return False if any required CODE tags are missing
        all_implemented = code_tags.issubset(implemented_tags)
        if not all_implemented:
            missing_tags = code_tags - implemented_tags
            logger.debug(f"Missing CODE tags: {missing_tags}")
        return all_implemented

    def _verify_test_tags_implementation(self, test_tags: Set[str]) -> bool:
        """Verify that all @TEST tags have corresponding test implementations

        Args:
            test_tags: Set of TEST tag IDs

        Returns:
            True if all tags are implemented
        """
        if not test_tags:
            return True

        # Search through test files for TEST tags
        implemented_tags = set()

        try:
            for test_file in self.tests_dir.rglob("*.py"):
                try:
                    with open(test_file, 'r', encoding='utf-8') as f:
                        content = f.read()
                        for tag_id in test_tags:
                                implemented_tags.add(tag_id)
                except Exception as e:
                    logger.warning(f"Error reading test file {test_file}: {e}")

        except Exception as e:
            logger.error(f"Error scanning test directory: {e}")

        return test_tags.issubset(implemented_tags)

    def _check_acceptance_criteria(self, spec_file: Path) -> bool:
        """Check if SPEC has acceptance criteria section

        Args:
            spec_file: Path to SPEC file

        Returns:
            True if acceptance criteria are present
        """
        try:
            with open(spec_file, 'r', encoding='utf-8') as f:
                content = f.read()

            # Look for acceptance criteria section
            acceptance_patterns = [
                r'#+\s*acceptance\s+criteria',
                r'#+\s*acceptance\s+criteria',  # English (Korean removed)
                r'#+\s*acceptance\s+criteria',  # English (Japanese removed)
                r'#+\s*criterios\s+de\s+aceptación'
            ]

            return any(re.search(pattern, content, re.IGNORECASE) for pattern in acceptance_patterns)

        except Exception as e:
            logger.error(f"Error checking acceptance criteria in {spec_file}: {e}")
            return False

    def _check_documentation_sync(self, spec_id: str) -> bool:
        """Check if documentation is synchronized with implementation

        Args:
            spec_id: The SPEC identifier

        Returns:
            True if documentation appears synchronized
        """
        # This is a simplified check - in reality, you'd want more sophisticated
        # documentation sync detection
        try:
            spec_dir = self.specs_dir / spec_id

            # Check for recent sync reports
            reports_dir = self.project_root / ".moai" / "reports"
            if reports_dir.exists():
                sync_reports = list(reports_dir.glob("sync-report-*.md"))
                if sync_reports:
                    # Get the most recent sync report
                    latest_report = max(sync_reports, key=os.path.getctime)
                    # If sync report is recent (within 24 hours), consider docs synced
                    import time
                    if time.time() - os.path.getctime(latest_report) < 86400:  # 24 hours
                        return True

            # Fallback: check if docs directory exists and has content
            return self.docs_dir.exists() and any(self.docs_dir.iterdir())

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
            version = str(current_version).strip('"\'')

            if version.startswith('0.'):
                # Major version bump for completion
                return "1.0.0"
            else:
                # Minor version bump for updates
                parts = version.split('.')
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