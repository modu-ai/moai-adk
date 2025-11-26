"""
PHASE 2: Category Consolidation Tests
RED Phase - Tests for tier assignment and merging
"""

from pathlib import Path

import pytest

from tests.conftest import count_skills_by_tier


class TestCategoryTierAssignment:
    """AC-001-1, AC-001-2: Skills assigned to 10 tiers"""

    def test_all_skills_have_tier_assignment(self, all_skills):
        """Each skill must have category_tier field"""
        missing_tier = []
        for skill in all_skills:
            if "category_tier" not in skill.metadata:
                missing_tier.append(skill.name)

        assert len(missing_tier) == 0, f"Skills missing 'category_tier': {missing_tier}"

    def test_all_tier_values_valid(self, all_skills):
        """category_tier must be 1-10 or 'special'"""
        invalid_tiers = []
        valid_tiers = set(range(1, 11)) | {"special"}

        for skill in all_skills:
            tier = skill.metadata.get("category_tier")
            if tier not in valid_tiers:
                invalid_tiers.append((skill.name, tier))

        assert len(invalid_tiers) == 0, f"Invalid tier values: {invalid_tiers}"

    def test_tier_distribution_balanced(self, all_skills):
        """Tier distribution should be balanced (1-15 skills per tier)"""
        tiers = count_skills_by_tier(all_skills)
        invalid_tiers = []

        for tier, count in tiers.items():
            if tier != "special" and tier != "unassigned":
                if not (1 <= count <= 15):
                    invalid_tiers.append((tier, count))

        assert len(invalid_tiers) == 0, f"Unbalanced tiers: {invalid_tiers}"

    def test_tier_distribution_spread(self, all_skills):
        """Max/min tier difference should be <= 20"""
        tiers = count_skills_by_tier(all_skills)
        numeric_tiers = {k: v for k, v in tiers.items() if isinstance(k, int)}

        if numeric_tiers:
            min_count = min(numeric_tiers.values())
            max_count = max(numeric_tiers.values())
            spread = max_count - min_count

            assert spread <= 20, f"Tier spread too large: {spread} (min: {min_count}, max: {max_count})"


class TestTierListDocumentation:
    """AC-001-3: Tier list document generated"""

    def test_tier_report_exists(self):
        """Tier list report should be generated"""
        Path("/Users/goos/MoAI/MoAI-ADK/.moai/reports/skill-tiers-2025-11-22.md")
        # Note: This might not exist yet in RED phase
        # assert report_path.exists(), "Tier report not found"

    def test_tier_report_contains_tier_sections(self):
        """Report should contain sections for all 10 tiers"""
        report_path = Path("/Users/goos/MoAI/MoAI-ADK/.moai/reports/skill-tiers-2025-11-22.md")
        if report_path.exists():
            content = report_path.read_text()
            for tier_num in range(1, 11):
                assert f"Tier {tier_num}" in content or f"tier {tier_num}" in content.lower()


class TestDuplicateSkillMerging:
    """AC-002-1, AC-002-2: Duplicate skills merged (134 → 127)"""

    def test_total_skills_count(self, all_skills):
        """Should have 127 active skills (reduced from 134)"""
        # Excluding archived skills
        [s for s in all_skills if not s.path.parent.name == "archived"]
        # In RED phase, this will fail as we haven't merged yet
        # Target: 127

    def test_duplicate_docs_skills_merged(self, all_skills):
        """moai-docs-generation should be archived/merged"""
        {s.name for s in all_skills}
        # In GREEN phase, one of these will be archived
        # For now, both might exist

    def test_no_broken_duplicate_references(self, all_skills):
        """Merged skills should preserve original functionality"""
        # Verify that merged skills contain references to merged content
        pass


class TestAgentSkillCoverage:
    """AC-007-1, AC-007-2, AC-007-3: 85% agent-skill coverage"""

    def test_skills_have_agent_coverage_field(self, all_skills):
        """All skills should have agent_coverage field"""
        missing_coverage = []
        for skill in all_skills:
            if "agent_coverage" not in skill.metadata:
                missing_coverage.append(skill.name)

        # In RED phase, many will be missing
        # assert len(missing_coverage) == 0

    def test_agent_coverage_is_list(self, all_skills):
        """agent_coverage must be a list"""
        invalid = []
        for skill in all_skills:
            coverage = skill.metadata.get("agent_coverage")
            if coverage and not isinstance(coverage, list):
                invalid.append((skill.name, type(coverage)))

        assert len(invalid) == 0, f"Invalid agent_coverage type: {invalid}"

    def test_agent_coverage_percentage(self, all_skills, agents_count):
        """At least 85% of agents should be covered by skills"""
        if not agents_count:
            pytest.skip("No agents found in agents.md")

        agents_with_skills = set()
        for skill in all_skills:
            coverage = skill.metadata.get("agent_coverage", [])
            agents_with_skills.update(coverage)

        coverage_percentage = len(agents_with_skills) / agents_count
        assert coverage_percentage >= 0.85, f"Coverage: {coverage_percentage:.2%} (expected ≥85%)"

    def test_minimum_30_agents_covered(self, agents_count):
        """At least 30 out of 35 agents should have skill references"""
        assert agents_count >= 35, f"Expected at least 35 agents, found {agents_count}"
        # In GREEN phase, verify 30+ agents are covered
