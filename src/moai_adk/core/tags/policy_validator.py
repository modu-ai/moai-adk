#!/usr/bin/env python3
# @CODE:TAG-POLICY-VALIDATOR-001 | @SPEC:TAG-POLICY-001
"""Real-time TAG policy violation detection system

Real-time TAG policy validator enforcing MoAI-ADK's SPEC-first principle.
Integrates with Pre-Tool-Use hooks to fundamentally block SPEC-less code generation.

Key Features:
- Real-time TAG policy violation detection
- Block CODE generation without SPEC
- TAG chain integrity verification
- Immediate violation reporting and correction guidance

@SPEC:TAG-POLICY-001
"""

import json
import re
import time
from dataclasses import dataclass, field
from enum import Enum
from pathlib import Path
from typing import Any, Dict, List, Optional, Set

from moai_adk.core.tags.language_dirs import (
    detect_directories,
    get_exclude_patterns,
    is_code_directory,
)


class PolicyViolationLevel(Enum):
    """Policy violation severity levels

    CRITICAL: Critical violation requiring immediate task termination
    HIGH: High-level violation requiring user confirmation before proceeding
    MEDIUM: Warning-level violation (advisory)
    LOW: Information-level violation (recommendation)
    """
    CRITICAL = "critical"
    HIGH = "high"
    MEDIUM = "medium"
    LOW = "low"


class PolicyViolationType(Enum):
    """Policy violation types

    SPECLESS_CODE: CODE generation without SPEC (critical)
    MISSING_TAGS: Required TAG missing (high)
    CHAIN_BREAK: TAG chain disconnected (high)
    DUPLICATE_TAGS: Duplicate TAG (medium)
    FORMAT_INVALID: TAG format error (medium)
    NO_SPEC_REFERENCE: CODE has no SPEC reference (low)
    """
    SPECLESS_CODE = "specless_code"
    MISSING_TAGS = "missing_tags"
    CHAIN_BREAK = "chain_break"
    DUPLICATE_TAGS = "duplicate_tags"
    FORMAT_INVALID = "format_invalid"
    NO_SPEC_REFERENCE = "no_spec_reference"


@dataclass
class PolicyViolation:
    """TAG policy violation information

    Attributes:
        level: Violation severity level (CRITICAL|HIGH|MEDIUM|LOW)
        type: Violation type (PolicyViolationType)
        tag: Related TAG (if applicable)
        message: Violation description
        file_path: Related file path
        action: Suggested action (block|warn|suggest)
        guidance: Correction guidance
        auto_fix_possible: Whether automatic fix is possible
    """
    level: PolicyViolationLevel
    type: PolicyViolationType
    tag: Optional[str]
    message: str
    file_path: Optional[str]
    action: str  # block|warn|suggest
    guidance: str
    auto_fix_possible: bool = False

    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary"""
        return {
            "level": self.level.value,
            "type": self.type.value,
            "tag": self.tag,
            "message": self.message,
            "file_path": self.file_path,
            "action": self.action,
            "guidance": self.guidance,
            "auto_fix_possible": self.auto_fix_possible
        }

    def should_block_operation(self) -> bool:
        """Whether the operation should be blocked"""
        return self.level == PolicyViolationLevel.CRITICAL or self.action == "block"


@dataclass
class PolicyValidationConfig:
    """TAG policy validation configuration

    Attributes:
        strict_mode: Strict mode (block all violations)
        require_spec_before_code: Require SPEC before CODE generation
        require_test_for_code: Require TEST for CODE
        allow_duplicate_tags: Whether to allow duplicate TAGs
        validation_timeout: Validation timeout (seconds)
        auto_fix_enabled: Enable automatic fix feature
        file_types_to_validate: File extensions to validate
    """
    strict_mode: bool = True
    require_spec_before_code: bool = True
    require_test_for_code: bool = True
    allow_duplicate_tags: bool = False
    validation_timeout: int = 5
    auto_fix_enabled: bool = False
    file_types_to_validate: Set[str] = field(default_factory=lambda: {
        "py", "js", "ts", "jsx", "tsx", "md", "txt", "yml", "yaml"
    })


class TagPolicyValidator:
    """Real-time TAG policy validator

    Integrates with Pre-Tool-Use hooks to detect and block TAG policy violations
    during file creation/modification. Enforces SPEC-first principle for quality assurance.

    Usage:
        config = PolicyValidationConfig(strict_mode=True)
        validator = TagPolicyValidator(config=config)

        # Validate before file creation
        violations = validator.validate_before_creation(
            file_path="src/example.py",
            content="def example(): pass"
        )

        # Check if operation should be blocked
        should_block = any(v.should_block_operation() for v in violations)
    """

    # TAG regex pattern
    TAG_PATTERN = re.compile(r"@(SPEC|CODE|TEST|DOC):([A-Z0-9-]+-\d{3})")

    def __init__(self, config: Optional[PolicyValidationConfig] = None, project_config: Optional[Dict] = None):
        """Initialize validator

        Args:
            config: Policy validation configuration (default: PolicyValidationConfig())
            project_config: Project configuration (loaded from .moai/config/config.json, optional)
        """
        self.config = config or PolicyValidationConfig()
        self.project_config = project_config or self._load_project_config()
        self.code_directories = detect_directories(self.project_config)
        self.exclude_patterns = get_exclude_patterns(self.project_config)
        self._start_time = time.time()

    def validate_before_creation(self, file_path: str, content: str) -> List[PolicyViolation]:
        """Validate TAG policy before file creation

        Called from Pre-Tool-Use hooks. Detects policy violations at file creation time.

        Args:
            file_path: Path to file being created/modified
            content: File content

        Returns:
            List of PolicyViolation
        """
        violations: List[PolicyViolation] = []

        # Check timeout
        if time.time() - self._start_time > self.config.validation_timeout:
            return violations

        # Check file type
        if not self._should_validate_file(file_path):
            return violations

        # Extract existing TAGs
        existing_tags = self._extract_tags_from_content(content)

        # Check if new file
        is_new_file = not Path(file_path).exists()

        if is_new_file:
            # Validate new file creation
            violations.extend(self._validate_new_file_creation(file_path, existing_tags))
        else:
            # Validate file modification
            violations.extend(self._validate_file_modification(file_path, existing_tags))

        return violations

    def validate_after_modification(self, file_path: str, content: str) -> List[PolicyViolation]:
        """Validate TAG policy after file modification

        Called from Post-Tool-Use hooks. Validates final state after modification.

        Args:
            file_path: Modified file path
            content: Modified file content

        Returns:
            List of PolicyViolation (mostly warning level)
        """
        violations: List[PolicyViolation] = []

        # Check timeout
        if time.time() - self._start_time > self.config.validation_timeout:
            return violations

        # Check file type
        if not self._should_validate_file(file_path):
            return violations

        # Extract TAGs
        tags = self._extract_tags_from_content(content)

        # Check missing TAGs
        violations.extend(self._validate_missing_tags(file_path, tags))

        # Validate chain integrity
        violations.extend(self._validate_chain_integrity(file_path, tags))

        return violations

    def _should_validate_file(self, file_path: str) -> bool:
        """Check if file should be validated

        Args:
            file_path: File path

        Returns:
            True if file should be validated
        """
        path = Path(file_path)
        suffix = path.suffix.lstrip(".")

        # Check file extension
        if suffix not in self.config.file_types_to_validate:
            return False

        # Exclude optional file patterns (not TAG validation targets)
        optional_patterns = [
            "CLAUDE.md",
            "README.md",
            "CHANGELOG.md",
            "CONTRIBUTING.md",
            ".claude/",
            ".moai/docs/",
            ".moai/reports/",
            ".moai/analysis/",
            "docs/",
            "templates/",
            "examples/",
        ]

        file_path_str = str(path)
        if any(pattern in file_path_str for pattern in optional_patterns):
            return False

        return True

    def _extract_tags_from_content(self, content: str) -> Dict[str, List[str]]:
        """Extract TAGs from content

        Args:
            content: File content

        Returns:
            Dictionary of {tag_type: [domains]}
        """
        tags: Dict[str, List[str]] = {
            "SPEC": [], "CODE": [], "TEST": [], "DOC": []
        }

        matches = self.TAG_PATTERN.findall(content)
        for tag_type, domain in matches:
            tags[tag_type].append(domain)

        return tags

    def _validate_new_file_creation(self, file_path: str, tags: Dict[str, List[str]]) -> List[PolicyViolation]:
        """Validate policy for new file creation

        Args:
            file_path: File path
            tags: Extracted TAGs

        Returns:
            List of PolicyViolation
        """
        violations = []

        # Check SPEC requirement for CODE file creation
        if self._is_code_file(file_path) and self.config.require_spec_before_code:
            if not tags.get("CODE"):
                violations.append(PolicyViolation(
                    level=PolicyViolationLevel.CRITICAL,
                    type=PolicyViolationType.SPECLESS_CODE,
                    tag=None,
                    message="CODE file has no @TAG",
                    file_path=file_path,
                    action="block",
                    guidance="CODE files must have a @CODE:DOMAIN-XXX format TAG. Create SPEC first.",
                    auto_fix_possible=False
                ))
            else:
                # If CODE TAG exists, check for linked SPEC
                for domain in tags["CODE"]:
                    spec_file = self._find_spec_file(domain)
                    if not spec_file:
                        spec_path = f".moai/specs/SPEC-{domain}/spec.md"
                        guidance = f"Create {spec_path} file or add to existing SPEC."
                        violations.append(PolicyViolation(
                            level=PolicyViolationLevel.HIGH,
                            type=PolicyViolationType.NO_SPEC_REFERENCE,
                            tag=f"@CODE:{domain}",
                            message=f"@CODE:{domain} has no linked SPEC",
                            file_path=file_path,
                            action="block" if self.config.strict_mode else "warn",
                            guidance=guidance,
                            auto_fix_possible=True
                        ))

        # Check CODE requirement for TEST file creation
        if self._is_test_file(file_path) and tags.get("TEST"):
            for domain in tags["TEST"]:
                code_file = self._find_code_file(domain)
                if not code_file:
                    violations.append(PolicyViolation(
                        level=PolicyViolationLevel.HIGH,
                        type=PolicyViolationType.CHAIN_BREAK,
                        tag=f"@TEST:{domain}",
                        message=f"@TEST:{domain} has no linked CODE",
                        file_path=file_path,
                        action="warn",
                        guidance=f"Create implementation file with @CODE:{domain} first.",
                        auto_fix_possible=False
                    ))

        return violations

    def _validate_file_modification(self, file_path: str, tags: Dict[str, List[str]]) -> List[PolicyViolation]:
        """Validate policy for file modification

        Args:
            file_path: File path
            tags: Extracted TAGs

        Returns:
            List of PolicyViolation
        """
        violations = []

        # Check for duplicate TAGs
        if not self.config.allow_duplicate_tags:
            duplicates = self._find_duplicate_tags(file_path, tags)
            for duplicate in duplicates:
                violations.append(PolicyViolation(
                    level=PolicyViolationLevel.MEDIUM,
                    type=PolicyViolationType.DUPLICATE_TAGS,
                    tag=duplicate,
                    message=f"Duplicate TAG: {duplicate}",
                    file_path=file_path,
                    action="warn",
                    guidance="Remove duplicate TAG. Each TAG should be unique.",
                    auto_fix_possible=True
                ))

        return violations

    def _validate_missing_tags(self, file_path: str, tags: Dict[str, List[str]]) -> List[PolicyViolation]:
        """Check for missing TAGs

        Args:
            file_path: File path
            tags: Extracted TAGs

        Returns:
            List of PolicyViolation
        """
        violations = []

        # If CODE file has no TAG
        if self._is_code_file(file_path) and not tags.get("CODE"):
            violations.append(PolicyViolation(
                level=PolicyViolationLevel.HIGH,
                type=PolicyViolationType.MISSING_TAGS,
                tag=None,
                message="CODE file is missing @TAG",
                file_path=file_path,
                action="suggest",
                guidance="Add @CODE:DOMAIN-XXX format TAG at the top of the file.",
                auto_fix_possible=True
            ))

        return violations

    def _validate_chain_integrity(self, file_path: str, tags: Dict[str, List[str]]) -> List[PolicyViolation]:
        """Validate TAG chain integrity

        Args:
            file_path: File path
            tags: Extracted TAGs

        Returns:
            List of PolicyViolation
        """
        violations = []

        # If CODE exists but TEST doesn't
        if tags.get("CODE") and self.config.require_test_for_code:
            for domain in tags["CODE"]:
                test_file = self._find_test_file(domain)
                if not test_file:
                    violations.append(PolicyViolation(
                        level=PolicyViolationLevel.MEDIUM,
                        type=PolicyViolationType.CHAIN_BREAK,
                        tag=f"@CODE:{domain}",
                        message=f"@CODE:{domain} has no linked TEST",
                        file_path=file_path,
                        action="suggest",
                        guidance=f"Create test file with @TEST:{domain} in tests/ directory.",
                        auto_fix_possible=True
                    ))

        return violations

    def _is_code_file(self, file_path: str) -> bool:
        """Check if file is a code file (dynamic detection by language)

        Args:
            file_path: File path

        Returns:
            True if code file
        """
        path = Path(file_path)

        # Check file extension (code file extensions only)
        code_extensions = {".py", ".js", ".ts", ".jsx", ".tsx", ".go", ".rs", ".kt", ".rb", ".php", ".java", ".cs"}
        if path.suffix not in code_extensions:
            return False

        # Dynamic detection using language_dirs
        return is_code_directory(path, self.project_config)

    def _is_test_file(self, file_path: str) -> bool:
        """Check if file is a test file

        Args:
            file_path: File path

        Returns:
            True if test file
        """
        path = Path(file_path)
        test_patterns = ["test/", "tests/", "__tests__", "spec/", "_test.", "_spec."]
        return any(pattern in str(path) for pattern in test_patterns)

    def _find_spec_file(self, domain: str) -> Optional[Path]:
        """Find SPEC file for given domain

        Args:
            domain: TAG domain (e.g., USER-REG-001)

        Returns:
            SPEC file path or None
        """
        spec_patterns = [
            f".moai/specs/SPEC-{domain}/spec.md",
            f".moai/specs/SPEC-{domain}.md",
            f"specs/SPEC-{domain}.md"
        ]

        for pattern in spec_patterns:
            path = Path(pattern)
            if path.exists():
                return path

        return None

    def _find_code_file(self, domain: str) -> Optional[Path]:
        """Find CODE file for given domain

        Args:
            domain: TAG domain

        Returns:
            CODE file path or None
        """
        # Search for CODE TAG from project root
        for pattern in ["src/**/*.py", "lib/**/*.py", "**/*.py", "**/*.js", "**/*.ts"]:
            for path in Path(".").glob(pattern):
                if path.is_file():
                    try:
                        content = path.read_text(encoding="utf-8", errors="ignore")
                        if f"@CODE:{domain}" in content:
                            return path
                    except Exception:
                        continue

        return None

    def _find_test_file(self, domain: str) -> Optional[Path]:
        """Find TEST file for given domain

        Args:
            domain: TAG domain

        Returns:
            TEST file path or None
        """
        test_patterns = [
            f"tests/**/test_*{domain}*.py",
            f"test/**/test_*{domain}*.py",
            f"tests/**/*{domain}*_test.py",
            f"**/*test*{domain}*.py"
        ]

        for pattern in test_patterns:
            for path in Path(".").glob(pattern):
                if path.is_file():
                    try:
                        content = path.read_text(encoding="utf-8", errors="ignore")
                        if f"@TEST:{domain}" in content:
                            return path
                    except Exception:
                        continue

        return None

    def _find_duplicate_tags(self, file_path: str, tags: Dict[str, List[str]]) -> List[str]:
        """Find duplicate TAGs in file

        Args:
            file_path: File path
            tags: Extracted TAGs

        Returns:
            List of duplicate TAGs
        """
        duplicates: List[str] = []

        try:
            content = Path(file_path).read_text(encoding="utf-8", errors="ignore")

            for tag_type, domains in tags.items():
                for domain in domains:
                    tag = f"@{tag_type}:{domain}"
                    count = content.count(tag)
                    if count > 1:
                        duplicates.append(tag)

        except Exception:
            pass

        return duplicates

    def create_validation_report(self, violations: List[PolicyViolation]) -> str:
        """Generate validation result report

        Args:
            violations: List of policy violations

        Returns:
            Formatted report string
        """
        if not violations:
            return "✅ TAG policy validation passed"

        lines = []
        lines.append("❌ TAG policy violations found")
        lines.append("=" * 50)

        # Group by level
        by_level: Dict[PolicyViolationLevel, List[PolicyViolation]] = {
            PolicyViolationLevel.CRITICAL: [],
            PolicyViolationLevel.HIGH: [],
            PolicyViolationLevel.MEDIUM: [],
            PolicyViolationLevel.LOW: []
        }

        for violation in violations:
            by_level[violation.level].append(violation)

        # Output by level
        level_names = {
            PolicyViolationLevel.CRITICAL: "🚨 Critical",
            PolicyViolationLevel.HIGH: "⚠️ High",
            PolicyViolationLevel.MEDIUM: "⚡ Medium",
            PolicyViolationLevel.LOW: "ℹ️ Low"
        }

        for level in [PolicyViolationLevel.CRITICAL, PolicyViolationLevel.HIGH,
                     PolicyViolationLevel.MEDIUM, PolicyViolationLevel.LOW]:
            level_violations = by_level[level]
            if level_violations:
                lines.append(f"\n{level_names[level]} ({len(level_violations)} found):")
                lines.append("-" * 30)

                for violation in level_violations:
                    tag_info = f" - {violation.tag}" if violation.tag else ""
                    lines.append(f"  {violation.message}{tag_info}")
                    if violation.file_path:
                        lines.append(f"    File: {violation.file_path}")
                    lines.append(f"    Action: {violation.guidance}")
                    if violation.auto_fix_possible:
                        lines.append("    🤖 Auto-fix available")
                    lines.append("")

        return "\n".join(lines)

    def _load_project_config(self) -> Dict:
        """Load project configuration (.moai/config/config.json)

        Returns:
            Project configuration dictionary
        """
        config_path = Path(".moai/config/config.json")
        if config_path.exists():
            try:
                return json.loads(config_path.read_text(encoding="utf-8"))
            except Exception:
                pass

        # Return default configuration
        return {"project": {"language": "python"}}

    def _fix_duplicate_tags(self, content: str) -> str:
        """Remove duplicate TAGs

        When same TAG appears multiple times, keep only the first occurrence and remove the rest.

        Args:
            content: File content

        Returns:
            Modified content
        """
        lines = content.split("\n")
        seen_tags = set()
        result_lines = []

        for line in lines:
            # Extract all TAGs from this line
            tags = self.TAG_PATTERN.findall(line)
            modified_line = line

            for tag_type, domain in tags:
                tag = f"@{tag_type}:{domain}"
                if tag in seen_tags:
                    # Already seen TAG - remove from this line
                    modified_line = modified_line.replace(f"{tag} | ", "")
                    modified_line = modified_line.replace(f" | {tag}", "")
                    modified_line = modified_line.replace(tag, "")
                else:
                    seen_tags.add(tag)

            result_lines.append(modified_line)

        return "\n".join(result_lines)

    def _fix_format_errors(self, content: str) -> str:
        """Fix TAG format errors

        - Missing colon: @CODE AUTH-001 → @CODE:AUTH-001
        - Normalize whitespace: @CODE:AUTH-001  |  @SPEC:... → @CODE:AUTH-001 | @SPEC:...

        Args:
            content: File content

        Returns:
            Modified content
        """
        # Fix missing colon (e.g., @CODE AUTH-001 → @CODE:AUTH-001)
        content = re.sub(r"@(SPEC|CODE|TEST|DOC)\s+([A-Z0-9-]+-\d{3})", r"@\1:\2", content)

        # Normalize whitespace (around pipe)
        content = re.sub(r"\s*\|\s*", " | ", content)

        # Remove duplicate whitespace
        content = re.sub(r"  +", " ", content)

        return content

    def _apply_auto_fix(self, file_path: str, violations: List[PolicyViolation]) -> Dict[str, Any]:
        """Apply automatic fixes

        Automatically fix SAFE-level violations according to configuration.

        Args:
            file_path: File path
            violations: List of policy violations

        Returns:
            Fix result dictionary
        """
        result: Dict[str, Any] = {
            "success": False,
            "fixed_count": 0,
            "pending_count": 0,
            "fixed_violations": [],
            "pending_violations": []
        }

        try:
            content = Path(file_path).read_text(encoding="utf-8")
            modified = False

            for violation in violations:
                if violation.type == PolicyViolationType.DUPLICATE_TAGS:
                    content = self._fix_duplicate_tags(content)
                    result["fixed_count"] += 1
                    result["fixed_violations"].append(violation)
                    modified = True

                elif violation.type == PolicyViolationType.FORMAT_INVALID:
                    content = self._fix_format_errors(content)
                    result["fixed_count"] += 1
                    result["fixed_violations"].append(violation)
                    modified = True

                else:
                    # Violation cannot be fixed
                    result["pending_count"] += 1
                    result["pending_violations"].append(violation)

            # Save modified content
            if modified:
                Path(file_path).write_text(content, encoding="utf-8")
                result["success"] = True

        except Exception as e:
            result["error"] = str(e)

        return result
