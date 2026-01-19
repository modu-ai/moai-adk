"""Comprehensive tests for migration module __init__.py.

Tests cover module exports and imports with 100% coverage.
"""



class TestModuleImports:
    """Test module can import all expected classes."""

    def test_import_version_migrator(self):
        """Test VersionMigrator can be imported."""
        from moai_adk.core.migration import VersionMigrator

        assert VersionMigrator is not None

    def test_import_version_detector(self):
        """Test VersionDetector can be imported."""
        from moai_adk.core.migration import VersionDetector

        assert VersionDetector is not None

    def test_import_backup_manager(self):
        """Test BackupManager can be imported."""
        from moai_adk.core.migration import BackupManager

        assert BackupManager is not None

    def test_import_file_migrator(self):
        """Test FileMigrator can be imported."""
        from moai_adk.core.migration import FileMigrator

        assert FileMigrator is not None

    def test_import_all_exports(self):
        """Test all __all__ exports are available."""
        from moai_adk.core.migration import __all__

        expected_exports = ["VersionMigrator", "VersionDetector", "BackupManager", "FileMigrator"]

        for export in expected_exports:
            assert export in __all__

    def test_all_exports_match_imports(self):
        """Test __all__ exports match actual imports."""
        import moai_adk.core.migration as migration_module
        from moai_adk.core.migration import __all__

        # Verify each export in __all__ can be imported
        for export in __all__:
            assert hasattr(migration_module, export), f"Export {export} not found in module"


class TestModuleStructure:
    """Test module structure and organization."""

    def test_module_has_docstring(self):
        """Test module has descriptive docstring."""
        import moai_adk.core.migration as migration_module

        assert migration_module.__doc__ is not None
        assert len(migration_module.__doc__) > 0
        assert "migration" in migration_module.__doc__.lower()

    def test_module_exports_count(self):
        """Test module exports expected number of items."""
        from moai_adk.core.migration import __all__

        assert len(__all__) == 4

    def test_module_exports_are_classes(self):
        """Test all exports are class types."""
        import inspect

        from moai_adk.core.migration import __all__

        for export in __all__:
            from moai_adk.core.migration import (
                BackupManager,
                FileMigrator,
                VersionDetector,
                VersionMigrator,
            )

            classes = {
                "VersionMigrator": VersionMigrator,
                "VersionDetector": VersionDetector,
                "BackupManager": BackupManager,
                "FileMigrator": FileMigrator,
            }
            assert export in classes
            assert inspect.isclass(classes[export])


class TestModuleDependencies:
    """Test module dependencies and imports."""

    def test_version_migrator_module_exists(self):
        """Test VersionMigrator can be imported from its module."""
        from moai_adk.core.migration.version_migrator import VersionMigrator

        assert VersionMigrator is not None

    def test_version_detector_module_exists(self):
        """Test VersionDetector can be imported from its module."""
        from moai_adk.core.migration.version_detector import VersionDetector

        assert VersionDetector is not None

    def test_backup_manager_module_exists(self):
        """Test BackupManager can be imported from its module."""
        from moai_adk.core.migration.backup_manager import BackupManager

        assert BackupManager is not None

    def test_file_migrator_module_exists(self):
        """Test FileMigrator can be imported from its module."""
        from moai_adk.core.migration.file_migrator import FileMigrator

        assert FileMigrator is not None


class TestModuleReExports:
    """Test module properly re-exports classes."""

    def test_version_migrator_is_re_exported(self):
        """Test VersionMigrator is re-exported from module."""
        from moai_adk.core.migration import VersionMigrator as Migrator
        from moai_adk.core.migration.version_migrator import VersionMigrator as OriginalMigrator

        # Both should reference the same class
        assert Migrator is OriginalMigrator

    def test_version_detector_is_re_exported(self):
        """Test VersionDetector is re-exported from module."""
        from moai_adk.core.migration import VersionDetector as Detector
        from moai_adk.core.migration.version_detector import VersionDetector as OriginalDetector

        assert Detector is OriginalDetector

    def test_backup_manager_is_re_exported(self):
        """Test BackupManager is re-exported from module."""
        from moai_adk.core.migration import BackupManager as Manager
        from moai_adk.core.migration.backup_manager import BackupManager as OriginalManager

        assert Manager is OriginalManager

    def test_file_migrator_is_re_exported(self):
        """Test FileMigrator is re-exported from module."""
        from moai_adk.core.migration import FileMigrator as Migrator
        from moai_adk.core.migration.file_migrator import FileMigrator as OriginalMigrator

        assert Migrator is OriginalMigrator


class TestModuleUsage:
    """Test typical usage patterns."""

    def test_import_all_migration_classes(self):
        """Test importing all migration classes at once."""
        from moai_adk.core.migration import (
            BackupManager,
            FileMigrator,
            VersionDetector,
            VersionMigrator,
        )

        # Verify all are imported and are classes
        assert VersionMigrator is not None
        assert VersionDetector is not None
        assert BackupManager is not None
        assert FileMigrator is not None

    def test_star_import_respects_all(self):
        """Test star import respects __all__ declaration."""
        # Execute star import in a controlled namespace
        namespace = {}
        exec("from moai_adk.core.migration import *", namespace)

        # Check that only __all__ exports are in namespace
        from moai_adk.core.migration import __all__

        for export in __all__:
            assert export in namespace

        # Check that no unexpected exports are in namespace (excluding __builtins__)
        expected_keys = set(__all__) | {"__builtins__", "__name__", "__doc__"}
        actual_keys = set(namespace.keys())

        # Allow for some Python builtins
        extra_keys = actual_keys - expected_keys
        # Filter out common dunder attributes that Python adds
        extra_keys = {k for k in extra_keys if not k.startswith("__")}
        assert len(extra_keys) == 0 or all(k in __all__ for k in extra_keys)


class TestModuleEdgeCases:
    """Test edge cases and error conditions."""

    def test_import_nonexistent_export_fails(self):
        """Test importing non-existent export raises ImportError."""
        # Use a try/except to verify the import error behavior
        try:
            from moai_adk.core.migration import NonExistentClass  # type: ignore

            assert False, "Expected ImportError but import succeeded"
        except ImportError:
            pass  # Expected behavior

    def test_module_name_is_correct(self):
        """Test module has correct name."""
        import moai_adk.core.migration as migration_module

        assert migration_module.__name__ == "moai_adk.core.migration"

    def test_module_file_attribute(self):
        """Test module has __file__ attribute."""
        import moai_adk.core.migration as migration_module

        assert hasattr(migration_module, "__file__")
        assert "migration" in migration_module.__file__
        assert migration_module.__file__.endswith("__init__.py")
