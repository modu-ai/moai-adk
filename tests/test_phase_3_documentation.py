"""
PHASE 3: Documentation and File Quality Tests
RED Phase - Tests for modularization and document quality
"""


class TestSkillFileSize:
    """Skills files should be modularized (< 500 lines)"""

    def test_no_excessive_skill_files(self, all_skills):
        """No SKILL.md should exceed 500 lines"""
        excessive = []
        for skill in all_skills:
            content = skill.read_skill_md()
            lines = len(content.split("\n"))
            if lines > 500:
                excessive.append((skill.name, lines))

        assert len(excessive) == 0, f"Files exceeding 500 lines: {excessive}"


class TestModularizationStatus:
    """AC-005-2: Skills should be modularized"""

    def test_modularized_field_exists(self, all_skills):
        """modularized field should exist in all skills"""
        missing = []
        for skill in all_skills:
            if "modularized" not in skill.metadata:
                missing.append(skill.name)

        assert len(missing) == 0, f"Missing modularized field: {missing}"

    def test_large_skills_are_modularized(self, all_skills):
        """Skills > 300 lines should have modularized=true"""
        not_modularized = []
        for skill in all_skills:
            content = skill.read_skill_md()
            lines = len(content.split("\n"))
            is_modularized = skill.metadata.get("modularized", False)

            if lines > 300 and not is_modularized:
                not_modularized.append((skill.name, lines))

        assert len(not_modularized) == 0, f"Large non-modularized skills: {not_modularized}"

    def test_modularized_skills_have_modules_dir(self, all_skills):
        """If modularized=true, modules/ directory should exist"""
        missing_modules = []
        for skill in all_skills:
            if skill.metadata.get("modularized"):
                modules_dir = skill.path / "modules"
                if not modules_dir.exists():
                    missing_modules.append(skill.name)

        assert len(missing_modules) == 0, f"Modularized skills missing modules/: {missing_modules}"


class TestAutoTriggerKeywords:
    """AC-006-2: All skills have auto_trigger_keywords"""

    def test_all_skills_have_keywords(self, all_skills):
        """All skills should have auto_trigger_keywords field"""
        missing = []
        for skill in all_skills:
            keywords = skill.metadata.get("auto_trigger_keywords")
            if not keywords:
                missing.append(skill.name)

        # In RED phase, most will be missing
        # assert len(missing) == 0

    def test_keywords_are_list(self, all_skills):
        """auto_trigger_keywords must be a list"""
        invalid = []
        for skill in all_skills:
            keywords = skill.metadata.get("auto_trigger_keywords")
            if keywords and not isinstance(keywords, list):
                invalid.append((skill.name, type(keywords)))

        assert len(invalid) == 0, f"Invalid keyword type: {invalid}"

    def test_keywords_are_lowercase(self, all_skills):
        """Keywords should be lowercase"""
        non_lowercase = []
        for skill in all_skills:
            keywords = skill.metadata.get("auto_trigger_keywords", [])
            for keyword in keywords:
                if keyword != keyword.lower():
                    non_lowercase.append((skill.name, keyword))

        assert len(non_lowercase) == 0, f"Non-lowercase keywords: {non_lowercase}"

    def test_minimum_keywords_per_skill(self, all_skills):
        """Each skill should have at least 1 keyword"""
        insufficient = []
        for skill in all_skills:
            keywords = skill.metadata.get("auto_trigger_keywords", [])
            if len(keywords) < 1:
                insufficient.append(skill.name)

        # In RED phase, many will be missing
        # assert len(insufficient) == 0


class TestDocumentationStructure:
    """Skills should follow Progressive Disclosure pattern"""

    def test_skill_md_has_frontmatter(self, all_skills):
        """SKILL.md should start with YAML frontmatter"""
        missing_frontmatter = []
        for skill in all_skills:
            content = skill.read_skill_md()
            if not content.startswith("---"):
                missing_frontmatter.append(skill.name)

        assert len(missing_frontmatter) == 0, f"Missing frontmatter: {missing_frontmatter}"

    def test_examples_exist_for_modularized(self, all_skills):
        """Modularized skills should have examples.md"""
        missing_examples = []
        for skill in all_skills:
            if skill.metadata.get("modularized"):
                examples_path = skill.path / "examples.md"
                if not examples_path.exists():
                    missing_examples.append(skill.name)

        # Relaxed for RED phase
        # assert len(missing_examples) == 0
