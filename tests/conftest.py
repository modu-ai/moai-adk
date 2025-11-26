"""Pytest configuration and fixtures for SPEC-SKILL-PORTFOLIO-OPT-001."""

import re
import sys
from pathlib import Path
from typing import Dict, List, Optional, Set

import pytest
import yaml

# ===== CRITICAL: Add paths to sys.path for test imports =====
PROJECT_ROOT = Path(__file__).parent.parent
HOOKS_MOAI_DIR = PROJECT_ROOT / ".claude" / "hooks" / "moai"
HOOKS_LIB_DIR = HOOKS_MOAI_DIR / "lib"

# Add moai directory to sys.path so "from lib import X" works
# This allows test files to do: from session import handle_session_start
# and session.py can do: from lib import HookPayload, HookResult
if str(HOOKS_MOAI_DIR) not in sys.path:
    sys.path.insert(0, str(HOOKS_MOAI_DIR))

# Also add lib directly for backward compatibility with tests
if str(HOOKS_LIB_DIR) not in sys.path:
    sys.path.insert(0, str(HOOKS_LIB_DIR))

# Add project root for other imports
if str(PROJECT_ROOT) not in sys.path:
    sys.path.insert(0, str(PROJECT_ROOT))

# Add src for moai_adk imports
SRC_DIR = PROJECT_ROOT / "src"
if str(SRC_DIR) not in sys.path:
    sys.path.insert(0, str(SRC_DIR))


# ===== CONSTANTS =====
SKILLS_DIR = PROJECT_ROOT / ".claude" / "skills"
SPEC_DIR = PROJECT_ROOT / ".moai" / "specs" / "SPEC-SKILL-PORTFOLIO-OPT-001"
TESTS_DIR = PROJECT_ROOT / "tests"


# ===== SKILL LOADING =====
class SkillMetadata:
    """Represents skill metadata from YAML frontmatter."""

    def __init__(self, skill_name: str, metadata: Dict, skill_path: Path):
        self.name = skill_name
        self.path = skill_path
        self.metadata = metadata
        self.skill_md_path = skill_path / "SKILL.md"

    def read_skill_md(self) -> str:
        """Read the full SKILL.md content."""
        if self.skill_md_path.exists():
            return self.skill_md_path.read_text(encoding="utf-8")
        return ""

    def __repr__(self):
        return f"<Skill {self.name}>"


def load_skill_metadata(skill_path: Path) -> Optional[Dict]:
    """Load YAML frontmatter from SKILL.md."""
    skill_md = skill_path / "SKILL.md"
    if not skill_md.exists():
        return None

    content = skill_md.read_text(encoding="utf-8")

    # Extract YAML frontmatter
    if not content.startswith("---"):
        return None

    try:
        end_frontmatter = content.find("---", 3)
        if end_frontmatter == -1:
            return None

        frontmatter_str = content[3:end_frontmatter].strip()
        metadata = yaml.safe_load(frontmatter_str)
        return metadata if metadata else {}
    except Exception:
        return {}


def load_all_skills() -> List[SkillMetadata]:
    """Load all skills from .claude/skills directory."""
    skills = []

    if not SKILLS_DIR.exists():
        return skills

    for skill_dir in sorted(SKILLS_DIR.iterdir()):
        if not skill_dir.is_dir():
            continue
        if skill_dir.name.startswith("."):
            continue

        metadata = load_skill_metadata(skill_dir)
        if metadata:
            skills.append(SkillMetadata(skill_dir.name, metadata, skill_dir))

    return skills


# ===== FIXTURES =====
@pytest.fixture(scope="session")
def all_skills() -> List[SkillMetadata]:
    """Load all skills once per test session."""
    return load_all_skills()


@pytest.fixture(scope="session")
def skills_count(all_skills) -> int:
    """Total number of skills."""
    return len(all_skills)


@pytest.fixture(scope="session")
def skill_names(all_skills) -> Set[str]:
    """Set of all skill names."""
    return {skill.name for skill in all_skills}


@pytest.fixture(scope="session")
def skill_names_from_metadata(all_skills) -> Set[str]:
    """Set of skill names from metadata 'name' field."""
    return {skill.metadata.get("name", "") for skill in all_skills if skill.metadata.get("name")}


@pytest.fixture(scope="session")
def agents() -> List[Dict]:
    """Load agents from agents.md."""
    agents_file = Path(__file__).parent.parent / ".moai" / "memory" / "agents.md"
    if not agents_file.exists():
        return []

    # Simple extraction of agent names from markdown
    content = agents_file.read_text(encoding="utf-8")
    # Extract agent names from bullet points with backticks
    agent_names = re.findall(r"`([a-z0-9-]+):`", content)
    return [{"name": agent} for agent in agent_names]


@pytest.fixture(scope="session")
def agents_count(agents) -> int:
    """Total number of agents."""
    return len(agents)


# ===== HELPER FUNCTIONS =====
def count_skills_by_tier(all_skills: List[SkillMetadata]) -> Dict[str, int]:
    """Count skills by category_tier."""
    tiers = {}
    for skill in all_skills:
        tier = skill.metadata.get("category_tier", "unassigned")
        tiers[tier] = tiers.get(tier, 0) + 1
    return tiers


def validate_semantic_version(version: str) -> bool:
    """Check if version follows X.Y.Z format."""
    return bool(re.match(r"^\d+\.\d+\.\d+$", version))


def validate_skill_name(name: str) -> bool:
    """Check if skill name follows Claude Code standard."""
    if not re.match(r"^[a-z0-9-]+$", name):
        return False
    if len(name) > 64:
        return False
    return True


def count_description_quality(all_skills: List[SkillMetadata]) -> Dict[str, int]:
    """Count descriptions by length category."""
    optimal = 0
    too_short = 0
    too_long = 0
    medium = 0

    for skill in all_skills:
        desc = skill.metadata.get("description", "")
        length = len(desc)

        if 100 <= length <= 200:
            optimal += 1
        elif length < 100:
            too_short += 1
        elif length > 300:
            too_long += 1
        else:
            medium += 1

    return {"optimal": optimal, "too_short": too_short, "too_long": too_long, "medium": medium}


def calculate_compliance_score(all_skills: List[SkillMetadata]) -> float:
    """Calculate overall compliance score."""
    if not all_skills:
        return 0.0

    scores = [skill.metadata.get("compliance_score", 0) for skill in all_skills]
    return sum(scores) / len(scores)
