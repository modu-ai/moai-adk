"""
MoAI-ADK Skill Loading System

Formal skill loading architecture with validation, dependency management, and caching
for MoAI-ADK's modular documentation system.
"""

import logging
import os
import re
import threading
from collections import OrderedDict
from dataclasses import dataclass, field
from datetime import datetime, timedelta
from typing import Any, Dict, List, Optional

import yaml

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


class SkillLoadingError(Exception):
    """Base exception for skill loading errors"""

    pass


class SkillNotFoundError(SkillLoadingError):
    """Raised when skill file is not found"""

    pass


class SkillValidationError(SkillLoadingError):
    """Raised when skill validation fails"""

    pass


class DependencyError(SkillLoadingError):
    """Raised when skill dependencies cannot be resolved"""

    pass


@dataclass
class SkillData:
    """Data structure for loaded skills with metadata and content"""

    name: str
    frontmatter: Dict[str, Any]
    content: str
    loaded_at: datetime
    applied_filters: List[str] = field(default_factory=list)

    def get_capability(self, capability_name: str) -> Any:
        """Get specific capability from skill content"""
        return self.frontmatter.get("capabilities", {}).get(capability_name)

    def supports_effort(self, effort: int) -> bool:
        """Check if skill supports specific effort level"""
        supported_efforts = self.frontmatter.get("supported_efforts", [1, 3, 5])
        return effort in supported_efforts

    def apply_filter(self, filter_type: str) -> None:
        """Apply content filter based on effort level"""
        if filter_type not in self.applied_filters:
            self.content = self._filter_content(self.content, filter_type)
            self.applied_filters.append(filter_type)

    def _filter_content(self, content: str, filter_type: str) -> str:
        """Filter content based on effort level"""
        if filter_type == "basic":
            # Keep only quick reference and basic patterns
            lines = content.split("\n")
            filtered_lines = []
            in_basic_section = False

            for line in lines:
                if line.startswith("## Quick Reference") or line.startswith("### Core Patterns"):
                    in_basic_section = True
                elif line.startswith("## ") and not line.startswith("## Quick Reference"):
                    in_basic_section = False

                if in_basic_section or line.startswith("#"):
                    filtered_lines.append(line)

            return "\n".join(filtered_lines)

        elif filter_type == "comprehensive":
            # Return full content
            return content

        else:  # standard
            # Filter out advanced sections
            lines = content.split("\n")
            filtered_lines = []
            skip_section = False

            for line in lines:
                if line.startswith("## Advanced") or line.startswith("## Implementation Details"):
                    skip_section = True
                elif line.startswith("## ") and not line.startswith("## Advanced"):
                    skip_section = False

                if not skip_section:
                    filtered_lines.append(line)

            return "\n".join(filtered_lines)

    @classmethod
    def get_empty_skill(cls, skill_name: str) -> "SkillData":
        """Create empty fallback skill"""
        return cls(
            name=skill_name,
            frontmatter={"name": skill_name, "status": "fallback"},
            content=f"# Fallback Skill: {skill_name}\n\nThis skill failed to load. Using fallback mode.",
            loaded_at=datetime.now(),
        )


class LRUCache:
    """Thread-safe LRU cache with TTL support"""

    def __init__(self, maxsize: int = 100, ttl: int = 3600):
        self.maxsize = maxsize
        self.ttl = ttl  # Time to live in seconds
        self.cache: OrderedDict[str, tuple[SkillData, datetime]] = OrderedDict()  # key -> (data, timestamp)
        self.lock = threading.Lock()

    def get(self, key: str) -> Optional[SkillData]:
        """Get value from cache"""
        with self.lock:
            if key in self.cache:
                data, timestamp = self.cache[key]
                if datetime.now() - timestamp < timedelta(seconds=self.ttl):
                    # Move to end (most recently used)
                    self.cache.move_to_end(key)
                    return data
                else:
                    # Expired, remove it
                    del self.cache[key]
            return None

    def set(self, key: str, value: SkillData) -> None:
        """Set value in cache"""
        with self.lock:
            self.cache[key] = (value, datetime.now())
            # Move to end (most recently used)
            self.cache.move_to_end(key)

            # Remove oldest if over maxsize
            while len(self.cache) > self.maxsize:
                self.cache.popitem(last=False)

    def clear(self) -> None:
        """Clear all cache entries"""
        with self.lock:
            self.cache.clear()

    def keys(self) -> List[str]:
        """Get all cache keys"""
        with self.lock:
            return list(self.cache.keys())


class SkillValidator:
    """Validates skill names, metadata, and compatibility"""

    def __init__(self, registry):
        self.registry = registry
        self.valid_skill_patterns = [
            r"^moai-[a-z]+-[a-z-]+$",  # Standard MoAI skills
            r"^[a-z]+-[a-z-]+$",  # Generic skills
        ]

    def validate_skill_name(self, skill_name: str) -> bool:
        """Validate skill name format and existence"""
        if not skill_name:
            raise SkillValidationError("Skill name cannot be empty")

        # Check against valid patterns
        for pattern in self.valid_skill_patterns:
            if re.match(pattern, skill_name):
                break
        else:
            raise SkillValidationError(f"Invalid skill name format: {skill_name}")

        # Check if skill exists in registry
        if skill_name not in self.registry.skills:
            raise SkillValidationError(f"Skill not found: {skill_name}")

        return True

    def validate_effort_parameter(self, skill_name: str, effort: int) -> bool:
        """Validate effort parameter against skill requirements"""
        metadata = self.registry.get_skill_metadata(skill_name)
        if not metadata:
            return False

        # Validate effort range
        if effort not in [1, 3, 5]:
            raise SkillValidationError(f"Invalid effort level: {effort}. Must be 1, 3, or 5")

        # Check if skill supports the requested effort level
        supported_efforts = metadata.get("supported_efforts", [1, 3, 5])
        if effort not in supported_efforts:
            raise SkillValidationError(
                f"Effort {effort} not supported by skill {skill_name}. Supported efforts: {supported_efforts}"
            )

        return True

    def validate_dependencies(self, skill_name: str, loaded_skills: List[str]) -> bool:
        """Validate skill dependencies against currently loaded skills"""
        metadata = self.registry.get_skill_metadata(skill_name)
        if not metadata:
            return False

        required_skills = metadata.get("requires", [])
        missing_skills = [skill for skill in required_skills if skill not in loaded_skills]

        if missing_skills:
            raise DependencyError(f"Skill {skill_name} requires missing dependencies: {missing_skills}")

        return True


class SkillRegistry:
    """Central registry for all available skills with metadata"""

    def __init__(self):
        self.skills: Dict[str, Dict[str, Any]] = {}
        self.dependencies: Dict[str, List[str]] = {}
        self.compatibility_matrix: Dict[str, Dict[str, bool]] = {}
        self._initialized = False

    def initialize_from_filesystem(self, skill_paths: List[str]) -> None:
        """Initialize registry by scanning filesystem for skills"""
        if self._initialized:
            return

        for skill_path in skill_paths:
            if os.path.exists(skill_path):
                self._scan_directory(skill_path)

        self._initialized = True
        logger.info(f"Initialized registry with {len(self.skills)} skills")

    def _scan_directory(self, directory: str) -> None:
        """Scan directory for skill files"""
        for root, dirs, files in os.walk(directory):
            for file in files:
                if file == "SKILL.md":
                    skill_path = os.path.join(root, file)
                    skill_name = self._extract_skill_name(skill_path)
                    if skill_name:
                        try:
                            metadata = self._parse_skill_metadata(skill_path)
                            self.register_skill(skill_name, metadata)
                        except Exception as e:
                            logger.warning(f"Failed to parse skill {skill_path}: {e}")

    def _extract_skill_name(self, skill_path: str) -> Optional[str]:
        """Extract skill name from file path"""
        # Extract from path like: .claude/skills/moai-foundation-claude/SKILL.md
        parts = skill_path.split(os.sep)
        if "skills" in parts:
            skills_idx = parts.index("skills")
            if skills_idx + 1 < len(parts):
                return parts[skills_idx + 1]
        return None

    def _parse_skill_metadata(self, skill_path: str) -> Dict[str, Any]:
        """Parse skill metadata from SKILL.md file"""
        with open(skill_path, "r", encoding="utf-8") as f:
            content = f.read()

        # Parse YAML frontmatter
        if content.startswith("---"):
            try:
                _, frontmatter, content = content.split("---", 2)
                metadata = yaml.safe_load(frontmatter.strip())
                return metadata or {}
            except Exception as e:
                logger.warning(f"Failed to parse frontmatter for {skill_path}: {e}")

        return {}

    def register_skill(self, skill_name: str, metadata: Dict[str, Any]) -> None:
        """Register a skill with its metadata"""
        self.skills[skill_name] = metadata

        # Extract dependencies
        self.dependencies[skill_name] = metadata.get("requires", [])

        # Initialize compatibility matrix
        if skill_name not in self.compatibility_matrix:
            self.compatibility_matrix[skill_name] = {}

    def get_skill_metadata(self, skill_name: str) -> Optional[Dict[str, Any]]:
        """Get skill metadata including dependencies"""
        return self.skills.get(skill_name)

    def check_compatibility(self, skill_name: str, other_skills: List[str]) -> bool:
        """Check if skill is compatible with currently loaded skills"""
        skill_compatibility = self.compatibility_matrix.get(skill_name, {})

        for other_skill in other_skills:
            if not skill_compatibility.get(other_skill, True):
                return False

        return True


class SkillLoader:
    """Main skill loading implementation with caching and error handling"""

    def __init__(self, skill_paths: Optional[List[str]] = None):
        self.registry = SkillRegistry()
        self.validator = SkillValidator(self.registry)
        self.cache = LRUCache(maxsize=100, ttl=3600)  # 1 hour TTL
        self.loading_stack: List[str] = []  # Track loading order to prevent circular dependencies

        # Initialize registry
        default_paths = [
            ".claude/skills",
            "src/moai_adk/.claude/skills",
            os.path.expanduser("~/.claude/skills"),
        ]
        self.registry.initialize_from_filesystem(skill_paths or default_paths)

    def load_skill(self, skill_name: str, effort: Optional[int] = None, force_reload: bool = False) -> SkillData:
        """Primary skill loading function with comprehensive validation"""
        try:
            # Check cache first (unless force_reload)
            if not force_reload:
                cached_skill = self.cache.get(skill_name)
                if cached_skill:
                    return self._validate_cached_skill(cached_skill, effort)

            # Prevent circular dependencies
            if skill_name in self.loading_stack:
                raise SkillLoadingError(f"Circular dependency detected: {skill_name}")

            self.loading_stack.append(skill_name)

            try:
                # Validate skill name and existence
                self.validator.validate_skill_name(skill_name)

                # Validate effort parameter if provided
                if effort is not None:
                    self.validator.validate_effort_parameter(skill_name, effort)

                # Check dependencies
                loaded_skills = list(self.cache.keys())
                self.validator.validate_dependencies(skill_name, loaded_skills)

                # Load skill data from file system
                skill_data = self._load_skill_from_filesystem(skill_name)

                # Apply effort parameter customization
                if effort is not None:
                    skill_data = self._apply_effort_parameter(skill_data, effort)

                # Cache the loaded skill
                self.cache.set(skill_name, skill_data)

                logger.info(f"Successfully loaded skill: {skill_name}")
                return skill_data

            finally:
                self.loading_stack.pop()

        except SkillLoadingError as e:
            logger.error(f"Failed to load skill {skill_name}: {e}")
            return self._get_fallback_skill(skill_name, effort)

    def _validate_cached_skill(self, cached_skill: SkillData, effort: Optional[int] = None) -> SkillData:
        """Validate cached skill against current effort parameter"""
        if effort is not None and not cached_skill.supports_effort(effort):
            logger.warning(f"Cached skill {cached_skill.name} doesn't support effort {effort}, reloading")
            return self.load_skill(cached_skill.name, effort, force_reload=True)

        return cached_skill

    def _load_skill_from_filesystem(self, skill_name: str) -> SkillData:
        """Load skill data from the file system"""
        skill_path = self._get_skill_path(skill_name)

        if not os.path.exists(skill_path):
            raise SkillNotFoundError(f"Skill file not found: {skill_path}")

        try:
            with open(skill_path, "r", encoding="utf-8") as f:
                content = f.read()

            # Parse frontmatter and content
            frontmatter, content = self._parse_skill_file(content)

            return SkillData(
                name=skill_name,
                frontmatter=frontmatter,
                content=content,
                loaded_at=datetime.now(),
            )

        except Exception as e:
            raise SkillLoadingError(f"Failed to parse skill file {skill_path}: {e}")

    def _get_skill_path(self, skill_name: str) -> str:
        """Get the file system path for a skill"""
        possible_paths = [
            f".claude/skills/{skill_name}/SKILL.md",
            f"src/moai_adk/.claude/skills/{skill_name}/SKILL.md",
            f"{os.path.expanduser('~')}/.claude/skills/{skill_name}/SKILL.md",
        ]

        for path in possible_paths:
            if os.path.exists(path):
                return path

        # Return the first path even if it doesn't exist (for error messages)
        return possible_paths[0]

    def _parse_skill_file(self, content: str) -> tuple[Dict[str, Any], str]:
        """Parse skill file into frontmatter and content"""
        if content.startswith("---"):
            try:
                _, frontmatter, content = content.split("---", 2)
                frontmatter_data = yaml.safe_load(frontmatter.strip())
                return frontmatter_data or {}, content.strip()
            except Exception as e:
                logger.warning(f"Failed to parse frontmatter: {e}")

        return {}, content

    def _apply_effort_parameter(self, skill_data: SkillData, effort: int) -> SkillData:
        """Apply effort parameter customization to skill data"""
        # Customize skill behavior based on effort level
        if effort == 1:
            # Minimal effort: basic functionality only
            skill_data.apply_filter("basic")
        elif effort == 3:
            # Standard effort: full functionality with moderate depth
            skill_data.apply_filter("standard")
        elif effort == 5:
            # Deep effort: comprehensive functionality with maximum depth
            skill_data.apply_filter("comprehensive")

        return skill_data

    def _get_fallback_skill(self, skill_name: str, effort: Optional[int] = None) -> SkillData:
        """Get fallback skill when loading fails"""
        fallback_map = {
            "moai-foundation-core": "moai-toolkit-essentials",
            "moai-lang-unified": "moai-lang-python",
            "moai-domain-backend": "expert-backend",
            "moai-domain-frontend": "expert-frontend",
        }

        fallback_name = fallback_map.get(skill_name, "moai-toolkit-essentials")

        logger.warning(f"Using fallback skill {fallback_name} for {skill_name}")

        try:
            return self.load_skill(fallback_name, effort, force_reload=True)
        except Exception:
            # Ultimate fallback
            return SkillData.get_empty_skill(skill_name)

    def get_cache_stats(self) -> Dict[str, Any]:
        """Get cache statistics for debugging"""
        return {
            "cached_skills": list(self.cache.keys()),
            "cache_size": len(self.cache.keys()),
            "registry_skills": len(self.registry.skills),
        }


# Global skill loader instance
SKILL_LOADER = SkillLoader()


def load_skill(skill_name: str, effort: Optional[int] = None, force_reload: bool = False) -> SkillData:
    """Public API for loading skills"""
    return SKILL_LOADER.load_skill(skill_name, effort, force_reload)


def get_skill_cache_stats() -> Dict[str, Any]:
    """Get skill loading system statistics"""
    return SKILL_LOADER.get_cache_stats()


def clear_skill_cache() -> None:
    """Clear all cached skills"""
    SKILL_LOADER.cache.clear()
    logger.info("Skill cache cleared")


# Enhanced Task function with automatic skill loading
def detect_required_skills(subagent_type: str, prompt: str) -> List[str]:
    """Auto-detect required skills based on task parameters"""
    skill_mapping = {
        "expert-backend": ["moai-lang-unified", "moai-domain-backend"],
        "expert-frontend": ["moai-lang-unified", "moai-domain-frontend"],
        "manager-tdd": ["moai-workflow-testing", "moai-foundation-quality"],
        "manager-spec": ["moai-foundation-claude", "moai-workflow-docs"],
        "security-expert": ["moai-quality-security", "moai-foundation-context"],
        "expert-devops": ["moai-system-universal", "moai-platform-baas"],
        "builder-agent": ["moai-foundation-claude"],
        "builder-skill": ["moai-foundation-claude"],
        "mcp-context7": ["moai-connector-mcp"],
        "mcp-sequential-thinking": ["moai-connector-mcp"],
    }

    # Base skills for all tasks
    base_skills = ["moai-foundation-claude"]

    # Add specific skills based on subagent_type
    specific_skills = skill_mapping.get(subagent_type, [])

    # Add skills detected from prompt content
    prompt_skills = _detect_skills_from_prompt(prompt)

    # Remove duplicates and return
    all_skills = list(set(base_skills + specific_skills + prompt_skills))

    logger.debug(f"Detected required skills for {subagent_type}: {all_skills}")
    return all_skills


def _detect_skills_from_prompt(prompt: str) -> List[str]:
    """Detect skills needed based on prompt content"""
    prompt_lower = prompt.lower()
    detected_skills = []

    # Language detection
    if any(lang in prompt_lower for lang in ["python", "fastapi", "django"]):
        detected_skills.append("moai-lang-unified")

    if any(lang in prompt_lower for lang in ["typescript", "react", "next.js", "frontend"]):
        detected_skills.append("moai-lang-unified")

    # Domain detection
    if any(domain in prompt_lower for domain in ["api", "backend", "server"]):
        detected_skills.append("moai-domain-backend")

    if any(domain in prompt_lower for domain in ["ui", "frontend", "component"]):
        detected_skills.append("moai-domain-frontend")

    # Security detection
    if any(sec in prompt_lower for sec in ["security", "auth", "vulnerability"]):
        detected_skills.append("moai-quality-security")

    return detected_skills


if __name__ == "__main__":
    # Test skill loading system
    print("Testing MoAI-ADK Skill Loading System")
    print("=" * 50)

    # Load a skill
    try:
        skill = load_skill("moai-foundation-claude", effort=3)
        print(f"Loaded skill: {skill.name}")
        print(f"Loaded at: {skill.loaded_at}")
        print(f"Content length: {len(skill.content)} characters")
        print(f"Applied filters: {skill.applied_filters}")
    except Exception as e:
        print(f"Error loading skill: {e}")

    # Show cache stats
    stats = get_skill_cache_stats()
    print(f"\nCache stats: {stats}")
