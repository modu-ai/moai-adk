"""
PHASE 1: Metadata Normalization Tests
RED Phase - All tests should initially FAIL
"""

from pathlib import Path

from tests.conftest import (
    calculate_compliance_score,
    count_description_quality,
    validate_semantic_version,
    validate_skill_name,
)


class TestMetadataPresence:
    """AC-004-1: All skills have required metadata fields"""

    def test_all_skills_have_name_field(self, all_skills):
        """Each skill must have a 'name' field"""
        missing_name = []
        for skill in all_skills:
            if "name" not in skill.metadata or not skill.metadata["name"]:
                missing_name.append(skill.name)

        assert len(missing_name) == 0, f"Skills missing 'name' field: {missing_name}"

    def test_all_skills_have_description_field(self, all_skills):
        """Each skill must have a 'description' field"""
        missing_desc = []
        for skill in all_skills:
            if "description" not in skill.metadata or not skill.metadata["description"]:
                missing_desc.append(skill.name)

        assert len(missing_desc) == 0, f"Skills missing 'description' field: {missing_desc}"

    def test_all_skills_have_version_field(self, all_skills):
        """Each skill must have a 'version' field"""
        missing_version = []
        for skill in all_skills:
            if "version" not in skill.metadata:
                missing_version.append(skill.name)

        assert len(missing_version) == 0, f"Skills missing 'version' field: {missing_version}"

    def test_all_skills_have_modularized_field(self, all_skills):
        """Each skill must have a 'modularized' field"""
        missing_modularized = []
        for skill in all_skills:
            if "modularized" not in skill.metadata:
                missing_modularized.append(skill.name)

        assert len(missing_modularized) == 0, f"Skills missing 'modularized' field: {missing_modularized}"

    def test_all_skills_have_last_updated_field(self, all_skills):
        """Each skill must have a 'last_updated' field"""
        missing_updated = []
        for skill in all_skills:
            if "last_updated" not in skill.metadata:
                missing_updated.append(skill.name)

        assert len(missing_updated) == 0, f"Skills missing 'last_updated' field: {missing_updated}"

    def test_all_skills_have_compliance_score_field(self, all_skills):
        """Each skill must have a 'compliance_score' field"""
        missing_compliance = []
        for skill in all_skills:
            if "compliance_score" not in skill.metadata:
                missing_compliance.append(skill.name)

        assert len(missing_compliance) == 0, f"Skills missing 'compliance_score' field: {missing_compliance}"


class TestVersionFormat:
    """AC-004-2: Version field follows semantic versioning"""

    def test_all_versions_are_semantic(self, all_skills):
        """All versions must follow X.Y.Z format"""
        invalid_versions = []
        for skill in all_skills:
            version = skill.metadata.get("version", "")
            if version and not validate_semantic_version(version):
                invalid_versions.append((skill.name, version))

        assert len(invalid_versions) == 0, f"Invalid semantic versions: {invalid_versions}"


class TestDescriptionQuality:
    """AC-004-3: Description length optimization"""

    def test_description_length_optimal_ratio(self, all_skills):
        """At least 60% of descriptions should be 100-200 characters"""
        quality = count_description_quality(all_skills)
        optimal_ratio = quality["optimal"] / len(all_skills)

        assert optimal_ratio >= 0.60, f"Optimal ratio: {optimal_ratio:.2%} (expected ≥60%)"

    def test_description_length_too_short_ratio(self, all_skills):
        """No more than 20% of descriptions should be < 100 characters"""
        quality = count_description_quality(all_skills)
        too_short_ratio = quality["too_short"] / len(all_skills)

        assert too_short_ratio <= 0.20, f"Too short: {too_short_ratio:.2%} (expected ≤20%)"

    def test_no_extremely_long_descriptions(self, all_skills):
        """No descriptions should exceed 300 characters"""
        quality = count_description_quality(all_skills)

        assert quality["too_long"] == 0, f"Found {quality['too_long']} descriptions > 300 chars"


class TestSkillNaming:
    """AC-003-1: Skill naming compliance"""

    def test_all_skill_names_valid(self, all_skills):
        """All skill directory names must be valid"""
        invalid_names = []
        for skill in all_skills:
            if not validate_skill_name(skill.name):
                invalid_names.append(skill.name)

        assert len(invalid_names) == 0, f"Invalid skill names: {invalid_names}"

    def test_metadata_name_matches_directory(self, all_skills):
        """Skill metadata 'name' should match directory name"""
        mismatches = []
        for skill in all_skills:
            metadata_name = skill.metadata.get("name", "")
            if metadata_name and metadata_name != skill.name:
                mismatches.append((skill.name, metadata_name))

        assert len(mismatches) == 0, f"Name mismatches: {mismatches}"


class TestNanoBananaRename:
    """AC-003-2: Rename moai-domain-nano-banana to moai-google-nano-banana"""

    def test_nano_banana_renamed(self, all_skills):
        """Old name should not exist"""
        skill_names = {s.name for s in all_skills}
        assert "moai-domain-nano-banana" not in skill_names, "Old name still exists"

    def test_google_nano_banana_exists(self, all_skills):
        """New name should exist"""
        skill_names = {s.name for s in all_skills}
        assert "moai-google-nano-banana" in skill_names, "New name not found"

    def test_backward_compatibility_alias_exists(self):
        """Migration alias or symlink should exist for backward compatibility"""
        alias_path = Path("/Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-domain-nano-banana")
        # Can be symlink or actual directory with migration guide
        has_alias = alias_path.exists() or alias_path.is_symlink()
        assert has_alias, "No backward compatibility alias found"


class TestMetadataCompleteness:
    """Overall metadata quality metrics"""

    def test_overall_compliance_score_calculation(self, all_skills):
        """Overall compliance score should be calculated"""
        score = calculate_compliance_score(all_skills)
        assert 0 <= score <= 100, f"Invalid compliance score: {score}"
        # Currently should be around 72%, target is 95%
        assert score < 100, "Already at perfect score?"

    def test_compliance_score_field_is_numeric(self, all_skills):
        """compliance_score field must be numeric"""
        invalid_scores = []
        for skill in all_skills:
            score = skill.metadata.get("compliance_score", 0)
            try:
                float(score)
            except (ValueError, TypeError):
                invalid_scores.append((skill.name, score))

        assert len(invalid_scores) == 0, f"Invalid compliance scores: {invalid_scores}"
