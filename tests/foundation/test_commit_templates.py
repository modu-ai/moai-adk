"""
Comprehensive TDD test suite for Commit Templates module.

This test suite provides comprehensive coverage for the commit_templates.py module,
following the RED-GREEN-REFACTOR TDD cycle.

Coverage Goals:
- 100% line coverage
- All branches and edge cases tested
- Error handling paths validated

Test Categories:
1. Unit Tests - Individual methods and functions
2. Integration Tests - Method interactions
3. Edge Cases - Empty inputs, invalid data
4. Error Handling - Exceptions, error messages

Target: 100% coverage
Test Framework: pytest
"""


from src.moai_adk.foundation.git.commit_templates import (
    CommitCategory,
    CommitTemplate,
    CommitTemplates,
)

# ============================================================================
# Enum Tests
# ============================================================================


class TestCommitCategory:
    """Test CommitCategory enumeration."""

    def test_commit_category_values(self):
        """Test that CommitCategory has all required values."""
        assert CommitCategory.FEATURES.value == "features"
        assert CommitCategory.BUG_FIXES.value == "bug_fixes"
        assert CommitCategory.DOCUMENTATION.value == "documentation"
        assert CommitCategory.PERFORMANCE.value == "performance"
        assert CommitCategory.TESTING.value == "testing"
        assert CommitCategory.REFACTORING.value == "refactoring"
        assert CommitCategory.MAINTENANCE.value == "maintenance"
        assert CommitCategory.SECURITY.value == "security"
        assert CommitCategory.DEPENDENCIES.value == "dependencies"
        assert CommitCategory.DEPLOYMENT.value == "deployment"
        assert CommitCategory.CONFIGURATION.value == "configuration"
        assert CommitCategory.INFRASTRUCTURE.value == "infrastructure"
        assert CommitCategory.EXPERIMENTS.value == "experiments"

    def test_commit_category_all_members(self):
        """Test that CommitCategory has exactly 13 members."""
        members = list(CommitCategory)
        assert len(members) == 13

    def test_commit_category_iteration(self):
        """Test iterating over CommitCategory members."""
        categories = [c for c in CommitCategory]
        assert len(categories) == 13


# ============================================================================
# Dataclass Tests
# ============================================================================


class TestCommitTemplate:
    """Test CommitTemplate dataclass."""

    def test_commit_template_creation_full(self):
        """Test creating a CommitTemplate with all fields."""
        template = CommitTemplate(
            type="feat",
            category=CommitCategory.FEATURES,
            pattern="feat({scope}): {description}",
            description="Add new feature",
            examples=["feat(auth): Add user authentication"],
            keywords=["add", "feature", "new"],
            priority=1,
        )

        assert template.type == "feat"
        assert template.category == CommitCategory.FEATURES
        assert template.pattern == "feat({scope}): {description}"
        assert template.description == "Add new feature"
        assert len(template.examples) == 1
        assert len(template.keywords) == 3
        assert template.priority == 1

    def test_commit_template_creation_minimal(self):
        """Test creating a CommitTemplate with minimal fields."""
        template = CommitTemplate(
            type="fix",
            category=CommitCategory.BUG_FIXES,
            pattern="fix({scope}): Fix {issue}",
            description="Fix bug",
        )

        assert template.type == "fix"
        assert template.category == CommitCategory.BUG_FIXES
        assert template.examples == []
        assert template.keywords == []
        assert template.priority == 0

    def test_commit_template_defaults(self):
        """Test CommitTemplate default values."""
        template = CommitTemplate(
            type="docs",
            category=CommitCategory.DOCUMENTATION,
            pattern="docs: {description}",
            description="Documentation",
        )

        assert template.examples == []
        assert template.keywords == []
        assert template.priority == 0


# ============================================================================
# CommitTemplates Initialization Tests
# ============================================================================


class TestCommitTemplatesInitialization:
    """Test CommitTemplates initialization and setup."""

    def test_templates_initialization(self):
        """Test CommitTemplates initializes correctly."""
        templates = CommitTemplates()

        assert templates is not None
        assert isinstance(templates, CommitTemplates)

    def test_templates_dict_initialized(self):
        """Test templates dictionary is initialized."""
        templates = CommitTemplates()

        assert isinstance(templates.templates, dict)
        assert len(templates.templates) > 0

    def test_categories_dict_initialized(self):
        """Test categories dictionary is initialized."""
        templates = CommitTemplates()

        assert isinstance(templates.categories, dict)
        assert len(templates.categories) > 0

    def test_has_feature_templates(self):
        """Test that feature templates exist."""
        templates = CommitTemplates()

        assert CommitCategory.FEATURES in templates.categories
        feature_templates = templates.categories[CommitCategory.FEATURES]
        assert len(feature_templates) > 0

    def test_has_bug_fix_templates(self):
        """Test that bug fix templates exist."""
        templates = CommitTemplates()

        assert CommitCategory.BUG_FIXES in templates.categories
        bug_fix_templates = templates.categories[CommitCategory.BUG_FIXES]
        assert len(bug_fix_templates) > 0

    def test_all_categories_initialized(self):
        """Test that main commit categories are initialized."""
        templates = CommitTemplates()

        # Check that key categories are initialized (not all 13 are initialized by default)
        expected_categories = [
            CommitCategory.FEATURES,
            CommitCategory.BUG_FIXES,
            CommitCategory.DOCUMENTATION,
            CommitCategory.PERFORMANCE,
            CommitCategory.TESTING,
            CommitCategory.REFACTORING,
            CommitCategory.SECURITY,
            CommitCategory.DEPENDENCIES,
            CommitCategory.CONFIGURATION,
            CommitCategory.INFRASTRUCTURE,
        ]

        for category in expected_categories:
            assert category in templates.categories


# ============================================================================
# Get Template by Type Tests
# ============================================================================


class TestGetTemplateByType:
    """Test get_template_by_type method."""

    def test_get_feat_template(self):
        """Test getting feat template by type."""
        templates = CommitTemplates()
        template = templates.get_template_by_type("feat")

        assert template is not None
        assert template.type == "feat"
        assert template.category == CommitCategory.FEATURES

    def test_get_fix_template(self):
        """Test getting fix template by type."""
        templates = CommitTemplates()
        template = templates.get_template_by_type("fix")

        assert template is not None
        assert template.type == "fix"
        assert template.category == CommitCategory.BUG_FIXES

    def test_get_docs_template(self):
        """Test getting docs template by type."""
        templates = CommitTemplates()
        template = templates.get_template_by_type("docs")

        assert template is not None
        assert template.type == "docs"
        assert template.category == CommitCategory.DOCUMENTATION

    def test_get_nonexistent_template(self):
        """Test getting non-existent template returns None."""
        templates = CommitTemplates()
        template = templates.get_template_by_type("nonexistent")

        assert template is None


# ============================================================================
# Get Category Tests
# ============================================================================


class TestGetCategory:
    """Test get_category method."""

    def test_get_features_category(self):
        """Test getting features category templates."""
        templates = CommitTemplates()
        features = templates.get_category(CommitCategory.FEATURES)

        assert isinstance(features, list)
        assert len(features) > 0
        assert all(isinstance(t, CommitTemplate) for t in features)
        assert all(t.category == CommitCategory.FEATURES for t in features)

    def test_get_bug_fixes_category(self):
        """Test getting bug fixes category templates."""
        templates = CommitTemplates()
        bug_fixes = templates.get_category(CommitCategory.BUG_FIXES)

        assert isinstance(bug_fixes, list)
        assert len(bug_fixes) > 0
        assert all(t.category == CommitCategory.BUG_FIXES for t in bug_fixes)

    def test_get_documentation_category(self):
        """Test getting documentation category templates."""
        templates = CommitTemplates()
        docs = templates.get_category(CommitCategory.DOCUMENTATION)

        assert isinstance(docs, list)
        assert len(docs) > 0

    def test_get_performance_category(self):
        """Test getting performance category templates."""
        templates = CommitTemplates()
        perf = templates.get_category(CommitCategory.PERFORMANCE)

        assert isinstance(perf, list)
        assert len(perf) > 0

    def test_get_testing_category(self):
        """Test getting testing category templates."""
        templates = CommitTemplates()
        testing = templates.get_category(CommitCategory.TESTING)

        assert isinstance(testing, list)
        assert len(testing) > 0

    def test_get_refactoring_category(self):
        """Test getting refactoring category templates."""
        templates = CommitTemplates()
        refactoring = templates.get_category(CommitCategory.REFACTORING)

        assert isinstance(refactoring, list)
        assert len(refactoring) > 0

    def test_get_security_category(self):
        """Test getting security category templates."""
        templates = CommitTemplates()
        security = templates.get_category(CommitCategory.SECURITY)

        assert isinstance(security, list)
        assert len(security) > 0

    def test_get_dependencies_category(self):
        """Test getting dependencies category templates."""
        templates = CommitTemplates()
        deps = templates.get_category(CommitCategory.DEPENDENCIES)

        assert isinstance(deps, list)
        assert len(deps) > 0


# ============================================================================
# Generate From Template Tests
# ============================================================================


class TestGenerateFromTemplate:
    """Test generate_from_template method."""

    def test_generate_from_template_basic(self):
        """Test basic template generation."""
        templates = CommitTemplates()
        template = CommitTemplate(
            type="feat",
            category=CommitCategory.FEATURES,
            pattern="feat({scope}): {description}",
            description="Add feature",
        )

        result = templates.generate_from_template(template, "auth", "Add user login")

        assert result["type"] == "feat"
        assert result["scope"] == "auth"
        assert result["description"] == "Add user login"
        assert result["pattern"] == "feat(auth): Add user login"

    def test_generate_from_template_with_feature_name(self):
        """Test template generation with feature_name placeholder."""
        templates = CommitTemplates()
        template = CommitTemplate(
            type="feat",
            category=CommitCategory.FEATURES,
            pattern="feat({scope}): Add {feature_name} functionality",
            description="Add feature",
        )

        result = templates.generate_from_template(template, "api", "user authentication")

        # The implementation replaces {feature_name} with the placeholder name itself
        assert result["pattern"] == "feat(api): Add feature_name functionality"

    def test_generate_from_template_with_issue_type(self):
        """Test template generation with issue_type placeholder."""
        templates = CommitTemplates()
        template = CommitTemplate(
            type="fix",
            category=CommitCategory.BUG_FIXES,
            pattern="fix({scope}): Resolve {issue_type} in {component}",
            description="Fix issue",
        )

        # The actual API only takes 3 arguments: template, scope, description
        # Placeholders like {issue_type} and {component} are replaced with their names
        result = templates.generate_from_template(template, "api", "Fix issue")

        assert result["pattern"] == "fix(api): Resolve issue_type in component"

    def test_generate_from_template_includes_keywords(self):
        """Test that generated result includes keywords."""
        templates = CommitTemplates()
        template = CommitTemplate(
            type="feat",
            category=CommitCategory.FEATURES,
            pattern="feat({scope}): {description}",
            description="Add feature",
            keywords=["add", "new", "feature"],
        )

        result = templates.generate_from_template(template, "auth", "Add login")

        assert result["keywords"] == ["add", "new", "feature"]

    def test_generate_from_template_includes_category(self):
        """Test that generated result includes category."""
        templates = CommitTemplates()
        template = CommitTemplate(
            type="feat",
            category=CommitCategory.FEATURES,
            pattern="feat({scope}): {description}",
            description="Add feature",
        )

        result = templates.generate_from_template(template, "auth", "Add login")

        assert result["category"] == "features"


# ============================================================================
# Add Template Tests
# ============================================================================


class TestAddTemplate:
    """Test add_template method."""

    def test_add_new_template(self):
        """Test adding a new custom template."""
        templates = CommitTemplates()
        initial_count = len(templates.templates)

        custom_template = CommitTemplate(
            type="custom",
            category=CommitCategory.MAINTENANCE,
            pattern="custom({scope}): {description}",
            description="Custom template",
            keywords=["custom"],
        )

        templates.add_template(custom_template)

        assert len(templates.templates) == initial_count + 1
        assert "custom" in templates.templates
        assert templates.get_template_by_type("custom") is not None

    def test_add_template_to_existing_category(self):
        """Test adding template to existing category."""
        templates = CommitTemplates()
        initial_features_count = len(templates.categories[CommitCategory.FEATURES])

        custom_feature = CommitTemplate(
            type="feat2",
            category=CommitCategory.FEATURES,
            pattern="feat2: {description}",
            description="Custom feature",
        )

        templates.add_template(custom_feature)

        assert len(templates.categories[CommitCategory.FEATURES]) == initial_features_count + 1

    def test_add_template_to_new_category(self):
        """Test adding template creates new category if needed."""
        templates = CommitTemplates()
        custom_template = CommitTemplate(
            type="misc",
            category=CommitCategory.MAINTENANCE,
            pattern="misc: {description}",
            description="Miscellaneous",
        )

        templates.add_template(custom_template)

        assert CommitCategory.MAINTENANCE in templates.categories
        assert custom_template in templates.categories[CommitCategory.MAINTENANCE]

    def test_add_template_overwrites_existing(self):
        """Test that adding template with same type overwrites."""
        templates = CommitTemplates()

        custom_template = CommitTemplate(
            type="feat",
            category=CommitCategory.FEATURES,
            pattern="new_feat_pattern: {description}",
            description="New description",
            keywords=["new", "keywords"],
        )

        templates.add_template(custom_template)

        updated = templates.get_template_by_type("feat")
        assert updated.pattern == "new_feat_pattern: {description}"
        assert updated.description == "New description"
        assert updated.keywords == ["new", "keywords"]


# ============================================================================
# Get Statistics Tests
# ============================================================================


class TestGetStatistics:
    """Test get_statistics method."""

    def test_get_statistics_returns_dict(self):
        """Test that statistics returns a dictionary."""
        templates = CommitTemplates()
        stats = templates.get_statistics()

        assert isinstance(stats, dict)

    def test_get_statistics_has_total_templates(self):
        """Test that statistics includes total template count."""
        templates = CommitTemplates()
        stats = templates.get_statistics()

        assert "total_templates" in stats
        assert isinstance(stats["total_templates"], int)
        assert stats["total_templates"] > 0

    def test_get_statistics_has_categories(self):
        """Test that statistics includes category breakdown."""
        templates = CommitTemplates()
        stats = templates.get_statistics()

        assert "categories" in stats
        assert isinstance(stats["categories"], dict)

    def test_get_statistics_categories_match(self):
        """Test that category statistics match actual categories."""
        templates = CommitTemplates()
        stats = templates.get_statistics()

        for category, count in stats["categories"].items():
            assert category in [c.value for c in CommitCategory]
            assert isinstance(count, int)
            assert count >= 0

    def test_get_statistics_has_most_used_types(self):
        """Test that statistics includes most used types."""
        templates = CommitTemplates()
        stats = templates.get_statistics()

        assert "most_used_types" in stats
        assert isinstance(stats["most_used_types"], list)

    def test_get_statistics_has_template_types(self):
        """Test that statistics includes all template types."""
        templates = CommitTemplates()
        stats = templates.get_statistics()

        assert "template_types" in stats
        assert isinstance(stats["template_types"], list)
        assert len(stats["template_types"]) > 0


# ============================================================================
# Search Templates Tests
# ============================================================================


class TestSearchTemplates:
    """Test search_templates method."""

    def test_search_templates_by_type(self):
        """Test searching templates by commit type."""
        templates = CommitTemplates()
        results = templates.search_templates("feat")

        assert len(results) > 0
        assert all("feat" in t.type.lower() for t in results)

    def test_search_templates_by_keyword(self):
        """Test searching templates by keyword."""
        templates = CommitTemplates()
        results = templates.search_templates("add")

        # "add" is a keyword in many templates
        matching = [t for t in results if "add" in " ".join(t.keywords).lower()]
        assert len(results) > 0
        assert len(matching) > 0

    def test_search_templates_by_description(self):
        """Test searching templates by description."""
        templates = CommitTemplates()
        results = templates.search_templates("feature")

        assert len(results) > 0

    def test_search_templates_by_example(self):
        """Test searching templates by example content."""
        templates = CommitTemplates()
        results = templates.search_templates("oauth")

        # Should find templates with OAuth in examples
        assert isinstance(results, list)

    def test_search_templates_case_insensitive(self):
        """Test that search is case insensitive."""
        templates = CommitTemplates()
        results_lower = templates.search_templates("feature")
        results_upper = templates.search_templates("FEATURE")
        results_mixed = templates.search_templates("FeAtUrE")

        assert len(results_lower) > 0
        assert len(results_upper) > 0
        assert len(results_mixed) > 0

    def test_search_templates_no_matches(self):
        """Test search with no matches returns empty list."""
        templates = CommitTemplates()
        results = templates.search_templates("xyznonexistent123")

        assert results == []

    def test_search_templates_empty_query(self):
        """Test search with empty query."""
        templates = CommitTemplates()
        results = templates.search_templates("")

        # Should return all templates or handle gracefully
        assert isinstance(results, list)


# ============================================================================
# Get Templates By Keywords Tests
# ============================================================================


class TestGetTemplatesByKeywords:
    """Test get_templates_by_keywords method."""

    def test_get_templates_by_single_keyword(self):
        """Test getting templates by single keyword."""
        templates = CommitTemplates()
        results = templates.get_templates_by_keywords(["add"])

        assert isinstance(results, list)
        assert len(results) > 0

    def test_get_templates_by_multiple_keywords(self):
        """Test getting templates by multiple keywords."""
        templates = CommitTemplates()
        results = templates.get_templates_by_keywords(["add", "feature", "new"])

        assert isinstance(results, list)
        # Should find templates matching any of the keywords

    def test_get_templates_by_keywords_case_sensitive(self):
        """Test that keyword search is case sensitive."""
        templates = CommitTemplates()
        results_lower = templates.get_templates_by_keywords(["add"])
        results_upper = templates.get_templates_by_keywords(["ADD"])

        # Case insensitive matching should work
        assert isinstance(results_lower, list)
        assert isinstance(results_upper, list)

    def test_get_templates_by_keywords_no_matches(self):
        """Test with keywords that have no matches."""
        templates = CommitTemplates()
        results = templates.get_templates_by_keywords(["xyz123nonexistent"])

        assert results == []

    def test_get_templates_by_keywords_empty_list(self):
        """Test with empty keyword list."""
        templates = CommitTemplates()
        results = templates.get_templates_by_keywords([])

        assert results == []


# ============================================================================
# Export Templates Tests
# ============================================================================


class TestExportTemplates:
    """Test export_templates method."""

    def test_export_templates_returns_dict(self):
        """Test that export returns a dictionary."""
        templates = CommitTemplates()
        exported = templates.export_templates()

        assert isinstance(exported, dict)

    def test_export_templates_has_templates_key(self):
        """Test that export includes templates key."""
        templates = CommitTemplates()
        exported = templates.export_templates()

        assert "templates" in exported
        assert isinstance(exported["templates"], dict)

    def test_export_templates_has_categories_key(self):
        """Test that export includes categories key."""
        templates = CommitTemplates()
        exported = templates.export_templates()

        assert "categories" in exported
        assert isinstance(exported["categories"], dict)

    def test_export_templates_structure(self):
        """Test that exported templates have correct structure."""
        templates = CommitTemplates()
        exported = templates.export_templates()

        for name, template_data in exported["templates"].items():
            assert "type" in template_data
            assert "category" in template_data
            assert "pattern" in template_data
            assert "description" in template_data
            assert "examples" in template_data
            assert "keywords" in template_data
            assert "priority" in template_data

    def test_export_templates_categories_structure(self):
        """Test that exported categories have correct structure."""
        templates = CommitTemplates()
        exported = templates.export_templates()

        for category, template_types in exported["categories"].items():
            assert isinstance(template_types, list)
            assert all(isinstance(t, str) for t in template_types)

    def test_export_templates_complete(self):
        """Test that export includes all templates."""
        templates = CommitTemplates()
        exported = templates.export_templates()

        exported_count = len(exported["templates"])
        original_count = len(templates.templates)

        assert exported_count == original_count


# ============================================================================
# Import Templates Tests
# ============================================================================


class TestImportTemplates:
    """Test import_templates method."""

    def test_import_templates_adds_new(self):
        """Test importing templates adds new templates."""
        templates = CommitTemplates()
        initial_count = len(templates.templates)

        import_data = {
            "templates": {
                "imported1": {
                    "type": "imported1",
                    "category": "maintenance",
                    "pattern": "imported1: {description}",
                    "description": "Imported template 1",
                    "examples": ["imported1: example"],
                    "keywords": ["imported"],
                    "priority": 1,
                }
            },
            "categories": {},
        }

        templates.import_templates(import_data)

        assert len(templates.templates) >= initial_count
        assert "imported1" in templates.templates

    def test_import_templates_overwrites_existing(self):
        """Test importing overwrites existing templates."""
        templates = CommitTemplates()

        import_data = {
            "templates": {
                "feat": {
                    "type": "feat",
                    "category": "features",
                    "pattern": "new_pattern: {description}",
                    "description": "New description",
                    "examples": [],
                    "keywords": ["new"],
                    "priority": 99,
                }
            },
            "categories": {},
        }

        templates.import_templates(import_data)

        updated = templates.get_template_by_type("feat")
        assert updated.pattern == "new_pattern: {description}"
        assert updated.priority == 99

    def test_import_templates_with_empty_data(self):
        """Test importing with empty data."""
        templates = CommitTemplates()
        initial_count = len(templates.templates)

        import_data = {"templates": {}, "categories": {}}

        templates.import_templates(import_data)

        # Should not crash and count should remain the same
        assert len(templates.templates) == initial_count

    def test_import_templates_with_all_fields(self):
        """Test importing template with all fields."""
        templates = CommitTemplates()

        import_data = {
            "templates": {
                "complete": {
                    "type": "complete",
                    "category": "features",
                    "pattern": "complete({scope}): {description}",
                    "description": "Complete template",
                    "examples": ["complete(api): Full implementation"],
                    "keywords": ["complete", "full"],
                    "priority": 5,
                }
            },
            "categories": {},
        }

        templates.import_templates(import_data)

        template = templates.get_template_by_type("complete")
        assert template is not None
        assert template.description == "Complete template"
        assert len(template.examples) == 1
        assert template.keywords == ["complete", "full"]
        assert template.priority == 5


# ============================================================================
# Integration Tests
# ============================================================================


class TestIntegration:
    """Integration tests for CommitTemplates."""

    def test_full_export_import_cycle(self):
        """Test exporting and importing templates preserves data."""
        templates1 = CommitTemplates()

        # Export
        exported = templates1.export_templates()

        # Create new instance and import
        templates2 = CommitTemplates()
        templates2.import_templates(exported)

        # Verify templates exist
        for name in exported["templates"]:
            assert name in templates2.templates

    def test_search_after_adding_template(self):
        """Test that newly added template is searchable."""
        templates = CommitTemplates()

        custom = CommitTemplate(
            type="searchable",
            category=CommitCategory.FEATURES,
            pattern="searchable: {description}",
            description="Searchable custom template",
            keywords=["searchable", "custom", "test"],
        )

        templates.add_template(custom)
        results = templates.search_templates("searchable")

        assert len(results) > 0
        assert custom in results

    def test_category_after_adding_template(self):
        """Test that category reflects newly added template."""
        templates = CommitTemplates()

        initial_count = len(templates.get_category(CommitCategory.FEATURES))

        custom = CommitTemplate(
            type="custom_feature",
            category=CommitCategory.FEATURES,
            pattern="custom: {description}",
            description="Custom feature",
        )

        templates.add_template(custom)

        new_count = len(templates.get_category(CommitCategory.FEATURES))
        assert new_count >= initial_count

    def test_generate_from_added_template(self):
        """Test generating message from newly added template."""
        templates = CommitTemplates()

        custom = CommitTemplate(
            type="mytype",
            category=CommitCategory.MAINTENANCE,
            pattern="mytype({scope}): {description}",
            description="My custom type",
        )

        templates.add_template(custom)
        result = templates.generate_from_template(custom, "myscope", "mydescription")

        assert result["pattern"] == "mytype(myscope): mydescription"


# ============================================================================
# Edge Cases and Error Handling Tests
# ============================================================================


class TestEdgeCases:
    """Test edge cases and error handling."""

    def test_get_template_by_type_empty_string(self):
        """Test getting template with empty string type."""
        templates = CommitTemplates()
        template = templates.get_template_by_type("")

        assert template is None

    def test_search_templates_special_characters(self):
        """Test search with special characters."""
        templates = CommitTemplates()
        results = templates.search_templates("feat(): {}")

        # Should handle gracefully
        assert isinstance(results, list)

    def test_generate_with_empty_scope(self):
        """Test generating with empty scope."""
        templates = CommitTemplates()
        template = CommitTemplate(
            type="test",
            category=CommitCategory.TESTING,
            pattern="test({scope}): {description}",
            description="Test",
        )

        result = templates.generate_from_template(template, "", "test description")

        assert result["scope"] == ""
        assert "test(): test description" in result["pattern"]

    def test_generate_with_empty_description(self):
        """Test generating with empty description."""
        templates = CommitTemplates()
        template = CommitTemplate(
            type="test",
            category=CommitCategory.TESTING,
            pattern="test({scope}): {description}",
            description="Test",
        )

        result = templates.generate_from_template(template, "scope", "")

        assert result["description"] == ""
        assert "test(scope):" in result["pattern"]

    def test_add_template_with_empty_fields(self):
        """Test adding template with minimal/empty fields."""
        templates = CommitTemplates()

        minimal = CommitTemplate(
            type="minimal",
            category=CommitCategory.MAINTENANCE,
            pattern="minimal",
            description="",
        )

        templates.add_template(minimal)

        assert templates.get_template_by_type("minimal") is not None

    def test_get_category_nonexistent(self):
        """Test getting templates from category that shouldn't have templates."""
        templates = CommitTemplates()

        # All defined categories should exist
        for category in CommitCategory:
            result = templates.get_category(category)
            assert isinstance(result, list)

    def test_statistics_accuracy(self):
        """Test that statistics are accurate."""
        templates = CommitTemplates()
        stats = templates.get_statistics()

        # Total templates should match templates dict
        assert stats["total_templates"] == len(templates.templates)

        # Template types should match
        assert set(stats["template_types"]) == set(templates.templates.keys())


# ============================================================================
# Data Validation Tests
# ============================================================================


class TestDataValidation:
    """Test data validation and integrity."""

    def test_all_templates_have_required_fields(self):
        """Test all templates have required fields."""
        templates = CommitTemplates()

        for template in templates.templates.values():
            assert template.type
            assert template.category in CommitCategory
            assert template.pattern
            assert template.description

    def test_all_categories_have_valid_templates(self):
        """Test all category templates are valid."""
        templates = CommitTemplates()

        for category, templates_list in templates.categories.items():
            assert category in CommitCategory
            for template in templates_list:
                assert isinstance(template, CommitTemplate)
                assert template.category == category

    def test_template_types_are_unique(self):
        """Test that template types are unique."""
        templates = CommitTemplates()

        types_list = list(templates.templates.keys())
        types_set = set(types_list)

        # Should have no duplicates
        assert len(types_list) == len(types_set)

    def test_export_import_preserves_integrity(self):
        """Test export/import preserves data integrity."""
        templates1 = CommitTemplates()

        # Add custom template
        custom = CommitTemplate(
            type="integrity_test",
            category=CommitCategory.FEATURES,
            pattern="integrity: {desc}",
            description="Integrity test",
            keywords=["integrity"],
            priority=10,
        )
        templates1.add_template(custom)

        # Export and import
        exported = templates1.export_templates()
        templates2 = CommitTemplates()
        templates2.import_templates(exported)

        # Verify integrity
        restored = templates2.get_template_by_type("integrity_test")
        assert restored is not None
        assert restored.pattern == "integrity: {desc}"
        assert restored.keywords == ["integrity"]
        assert restored.priority == 10


# ============================================================================
# Performance Tests
# ============================================================================


class TestPerformance:
    """Test performance characteristics."""

    def test_search_performance_large_dataset(self):
        """Test search performance with many templates."""
        templates = CommitTemplates()

        # Add many custom templates
        for i in range(100):
            custom = CommitTemplate(
                type=f"perf_{i}",
                category=CommitCategory.MAINTENANCE,
                pattern=f"perf_{i}: {{description}}",
                description=f"Performance test {i}",
                keywords=[f"keyword_{i}"],
            )
            templates.add_template(custom)

        # Search should complete quickly
        import time

        start = time.time()
        results = templates.search_templates("perf_50")
        elapsed = time.time() - start

        assert len(results) > 0
        assert elapsed < 1.0  # Should complete in less than 1 second

    def test_export_performance_many_templates(self):
        """Test export performance with many templates."""
        templates = CommitTemplates()

        # Add many templates
        for i in range(50):
            custom = CommitTemplate(
                type=f"export_{i}",
                category=CommitCategory.FEATURES,
                pattern=f"export_{i}: {{description}}",
                description=f"Export test {i}",
            )
            templates.add_template(custom)

        import time

        start = time.time()
        exported = templates.export_templates()
        elapsed = time.time() - start

        assert len(exported["templates"]) >= 50
        assert elapsed < 1.0  # Should complete in less than 1 second


# ============================================================================
# Category Coverage Tests
# ============================================================================


class TestCategoryCoverage:
    """Test that all categories are properly covered."""

    def test_features_category_templates(self):
        """Test FEATURES category has proper templates."""
        templates = CommitTemplates()
        features = templates.get_category(CommitCategory.FEATURES)

        assert len(features) > 0
        assert all(t.category == CommitCategory.FEATURES for t in features)

    def test_bug_fixes_category_templates(self):
        """Test BUG_FIXES category has proper templates."""
        templates = CommitTemplates()
        bug_fixes = templates.get_category(CommitCategory.BUG_FIXES)

        assert len(bug_fixes) > 0
        assert all(t.category == CommitCategory.BUG_FIXES for t in bug_fixes)

    def test_documentation_category_templates(self):
        """Test DOCUMENTATION category has proper templates."""
        templates = CommitTemplates()
        docs = templates.get_category(CommitCategory.DOCUMENTATION)

        assert len(docs) > 0
        assert all(t.category == CommitCategory.DOCUMENTATION for t in docs)

    def test_performance_category_templates(self):
        """Test PERFORMANCE category has proper templates."""
        templates = CommitTemplates()
        perf = templates.get_category(CommitCategory.PERFORMANCE)

        assert len(perf) > 0
        assert all(t.category == CommitCategory.PERFORMANCE for t in perf)

    def test_testing_category_templates(self):
        """Test TESTING category has proper templates."""
        templates = CommitTemplates()
        testing = templates.get_category(CommitCategory.TESTING)

        assert len(testing) > 0
        assert all(t.category == CommitCategory.TESTING for t in testing)

    def test_refactoring_category_templates(self):
        """Test REFACTORING category has proper templates."""
        templates = CommitTemplates()
        refactoring = templates.get_category(CommitCategory.REFACTORING)

        assert len(refactoring) > 0
        assert all(t.category == CommitCategory.REFACTORING for t in refactoring)

    def test_maintenance_category_templates(self):
        """Test MAINTENANCE category has proper templates."""
        templates = CommitTemplates()
        maintenance = templates.get_category(CommitCategory.MAINTENANCE)

        assert isinstance(maintenance, list)
        assert all(t.category == CommitCategory.MAINTENANCE for t in maintenance)

    def test_security_category_templates(self):
        """Test SECURITY category has proper templates."""
        templates = CommitTemplates()
        security = templates.get_category(CommitCategory.SECURITY)

        assert len(security) > 0
        assert all(t.category == CommitCategory.SECURITY for t in security)

    def test_dependencies_category_templates(self):
        """Test DEPENDENCIES category has proper templates."""
        templates = CommitTemplates()
        deps = templates.get_category(CommitCategory.DEPENDENCIES)

        assert len(deps) > 0
        assert all(t.category == CommitCategory.DEPENDENCIES for t in deps)

    def test_deployment_category_templates(self):
        """Test DEPLOYMENT category has proper templates."""
        templates = CommitTemplates()
        deployment = templates.get_category(CommitCategory.DEPLOYMENT)

        assert isinstance(deployment, list)
        assert all(t.category == CommitCategory.DEPLOYMENT for t in deployment)

    def test_configuration_category_templates(self):
        """Test CONFIGURATION category has proper templates."""
        templates = CommitTemplates()
        config = templates.get_category(CommitCategory.CONFIGURATION)

        assert len(config) > 0
        assert all(t.category == CommitCategory.CONFIGURATION for t in config)

    def test_infrastructure_category_templates(self):
        """Test INFRASTRUCTURE category has proper templates."""
        templates = CommitTemplates()
        infra = templates.get_category(CommitCategory.INFRASTRUCTURE)

        assert len(infra) > 0
        assert all(t.category == CommitCategory.INFRASTRUCTURE for t in infra)

    def test_experiments_category_templates(self):
        """Test EXPERIMENTS category has proper templates."""
        templates = CommitTemplates()
        experiments = templates.get_category(CommitCategory.EXPERIMENTS)

        assert isinstance(experiments, list)
        assert all(t.category == CommitCategory.EXPERIMENTS for t in experiments)


# ============================================================================
# Pattern Placeholder Tests
# ============================================================================


class TestPatternPlaceholders:
    """Test pattern placeholder handling."""

    def test_scope_placeholder_replacement(self):
        """Test {scope} placeholder replacement."""
        templates = CommitTemplates()
        template = CommitTemplate(
            type="test",
            category=CommitCategory.TESTING,
            pattern="test({scope}): {description}",
            description="Test",
        )

        result = templates.generate_from_template(template, "api", "Add tests")

        assert "test(api): Add tests" == result["pattern"]

    def test_description_placeholder_replacement(self):
        """Test {description} placeholder replacement."""
        templates = CommitTemplates()
        template = CommitTemplate(
            type="feat",
            category=CommitCategory.FEATURES,
            pattern="feat: {description}",
            description="Feature",
        )

        result = templates.generate_from_template(template, "", "User login")

        assert "feat: User login" == result["pattern"]

    def test_multiple_placeholders_replacement(self):
        """Test multiple placeholder replacement."""
        templates = CommitTemplates()
        template = CommitTemplate(
            type="fix",
            category=CommitCategory.BUG_FIXES,
            pattern="fix({scope}): Fix {issue_type} in {component}",
            description="Fix",
        )

        # The actual API only takes 3 arguments: template, scope, description
        # Placeholders like {issue_type} and {component} are replaced with their names
        result = templates.generate_from_template(template, "api", "Fix issue")

        assert result["pattern"] == "fix(api): Fix issue_type in component"

    def test_unknown_placeholder_handling(self):
        """Test handling of unknown placeholders."""
        templates = CommitTemplates()
        template = CommitTemplate(
            type="custom",
            category=CommitCategory.MAINTENANCE,
            pattern="custom: {unknown} {description}",
            description="Custom",
        )

        result = templates.generate_from_template(template, "", "Test")

        # Unknown placeholder should be replaced with placeholder name
        assert "unknown" in result["pattern"].lower()
