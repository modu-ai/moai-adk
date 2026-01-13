"""Comprehensive TDD tests for TAG parser module.

Targets 100% coverage for parser.py including:
- extract_tags_from_source edge cases (lines 19-61)
- extract_tags_from_file error handling (lines 64-107)
- extract_tags_from_files batch processing (lines 110-131)
- extract_tags_from_directory traversal (lines 134-168)
"""

from pathlib import Path

import pytest

from moai_adk.tag_system import parser


class TestExtractTagsFromSourceEdgeCases:
    """Test extract_tags_from_source with edge cases."""

    def test_extract_tags_from_source_with_none_input(self):
        """Test extract_tags_from_source with None input."""
        # type: ignore[arg-type]
        result = parser.extract_tags_from_source(None, Path("test.py"))
        assert result == []

    def test_extract_tags_from_source_with_non_string_input(self):
        """Test extract_tags_from_source with non-string input."""
        # Integer
        assert parser.extract_tags_from_source(123, Path("test.py")) == []  # type: ignore[arg-type]

        # List
        assert parser.extract_tags_from_source(["# @SPEC"], Path("test.py")) == []  # type: ignore[arg-type]

        # Dict
        assert parser.extract_tags_from_source({"code": "# @SPEC"}, Path("test.py")) == []  # type: ignore[arg-type]

        # Object
        assert parser.extract_tags_from_source(object(), Path("test.py")) == []  # type: ignore[arg-type]

    def test_extract_tags_from_source_with_empty_string(self):
        """Test extract_tags_from_source with empty string."""
        result = parser.extract_tags_from_source("", Path("test.py"))
        assert result == []

    def test_extract_tags_from_source_with_no_tags(self):
        """Test extract_tags_from_source with code but no TAGs."""
        source = """
def hello():
    print("Hello, world!")

class Foo:
    pass
"""
        result = parser.extract_tags_from_source(source, Path("test.py"))
        assert result == []

    def test_extract_tags_from_source_with_single_tag(self):
        """Test extract_tags_from_source with single TAG."""
        source = "# @SPEC SPEC-AUTH-001"
        result = parser.extract_tags_from_source(source, Path("test.py"))

        assert len(result) == 1
        assert result[0].spec_id == "SPEC-AUTH-001"
        assert result[0].verb == "impl"
        assert result[0].line == 1

    def test_extract_tags_from_source_with_multiple_tags(self):
        """Test extract_tags_from_source with multiple TAGs."""
        source = """
# @SPEC SPEC-AUTH-001
def auth():
    pass

# @SPEC SPEC-USER-002
class User:
    pass
"""
        result = parser.extract_tags_from_source(source, Path("test.py"))

        assert len(result) == 2
        assert result[0].spec_id == "SPEC-AUTH-001"
        assert result[1].spec_id == "SPEC-USER-002"

    def test_extract_tags_from_source_line_numbers_accurate(self):
        """Test that line numbers are accurate."""
        source = """
# Line 2
# Line 3
# @SPEC SPEC-AUTH-001
# Line 5
"""
        result = parser.extract_tags_from_source(source, Path("test.py"))

        assert len(result) == 1
        assert result[0].line == 4

    def test_extract_tags_from_source_with_inline_comment_tags(self):
        """Test TAGs in inline comments."""
        source = """
def func():  # @SPEC SPEC-AUTH-001
    pass

x = 1  # @SPEC SPEC-USER-002 verify
"""
        result = parser.extract_tags_from_source(source, Path("test.py"))

        assert len(result) == 2
        assert result[0].spec_id == "SPEC-AUTH-001"
        assert result[0].line == 2
        assert result[1].spec_id == "SPEC-USER-002"
        assert result[1].verb == "verify"

    def test_extract_tags_from_source_preserves_file_path(self):
        """Test that file_path is preserved."""
        source = "# @SPEC SPEC-AUTH-001"
        file_path = Path("my/module/auth.py")

        result = parser.extract_tags_from_source(source, file_path)

        assert len(result) == 1
        assert result[0].file_path == file_path

    def test_extract_tags_from_source_with_invalid_tags(self):
        """Test that invalid TAGs are skipped."""
        source = """
# @SPEC SPEC-AUTH-001
# @SPEC invalid-format
# @SPEC SPEC-USER-002
# Not a TAG
"""
        result = parser.extract_tags_from_source(source, Path("test.py"))

        # Only valid TAGs should be extracted
        assert len(result) == 2
        assert result[0].spec_id == "SPEC-AUTH-001"
        assert result[1].spec_id == "SPEC-USER-002"

    def test_extract_tags_from_source_with_all_verbs(self):
        """Test extracting all valid verbs."""
        source = """
# @SPEC SPEC-AUTH-001 impl
# @SPEC SPEC-AUTH-002 verify
# @SPEC SPEC-AUTH-003 depends
# @SPEC SPEC-AUTH-004 related
"""
        result = parser.extract_tags_from_source(source, Path("test.py"))

        assert len(result) == 4
        assert result[0].verb == "impl"
        assert result[1].verb == "verify"
        assert result[2].verb == "depends"
        assert result[3].verb == "related"

    def test_extract_tags_from_source_with_mixed_code_and_comments(self):
        """Test extracting TAGs from mixed code and comments."""
        source = """
import os

# @SPEC SPEC-IMPORT-001
def load_config():
    '''Load configuration.'''  # @SPEC SPEC-CONFIG-002 verify
    pass

class App:  # @SPEC SPEC-APP-003
    pass
"""
        result = parser.extract_tags_from_source(source, Path("test.py"))

        assert len(result) == 3
        assert result[0].line == 4
        assert result[1].line == 6
        assert result[2].line == 9

    def test_extract_tags_from_source_with_multiple_hashes_in_line(self):
        """Test handling of multiple # characters in a line."""
        source = """
# @SPEC SPEC-AUTH-001 # This is also a comment
string = "# Not a TAG comment"  # @SPEC SPEC-STR-002
"""
        result = parser.extract_tags_from_source(source, Path("test.py"))

        # Parser only finds first # in each line
        assert len(result) == 1
        assert result[0].spec_id == "SPEC-AUTH-001"
        # The second TAG in the inline comment won't be found because
        # parser.find("#") returns the first # only

    def test_extract_tags_from_source_with_tab_indented_comments(self):
        """Test TAGs in tab-indented comments."""
        source = "\t# @SPEC SPEC-AUTH-001\n\t\t# @SPEC SPEC-USER-002"
        result = parser.extract_tags_from_source(source, Path("test.py"))

        assert len(result) == 2
        assert result[0].line == 1
        assert result[1].line == 2

    def test_extract_tags_from_source_windows_line_endings(self):
        """Test source with Windows CRLF line endings."""
        source = "# @SPEC SPEC-AUTH-001\r\n# @SPEC SPEC-USER-002\r\n"
        result = parser.extract_tags_from_source(source, Path("test.py"))

        assert len(result) == 2
        assert result[0].line == 1
        assert result[1].line == 2

    def test_extract_tags_from_source_mac_classic_line_endings(self):
        """Test source with classic Mac CR line endings."""
        source = "# @SPEC SPEC-AUTH-001\r# @SPEC SPEC-USER-002\r"
        result = parser.extract_tags_from_source(source, Path("test.py"))

        assert len(result) == 2
        assert result[0].line == 1
        assert result[1].line == 2

    def test_extract_tags_from_source_unicode_in_source(self):
        """Test source with Unicode characters."""
        source = "# @SPEC SPEC-AUTH-001 인증 구현\n# @SPEC SPEC-USER-002 ユーザー"
        result = parser.extract_tags_from_source(source, Path("test.py"))

        assert len(result) == 2
        assert result[0].spec_id == "SPEC-AUTH-001"
        assert result[1].spec_id == "SPEC-USER-002"


class TestExtractTagsFromFileErrorHandling:
    """Test extract_tags_from_file error handling."""

    def test_extract_tags_from_file_with_path_object(self):
        """Test extract_tags_from_file with Path object."""
        # Create a temporary test file
        import tempfile

        with tempfile.NamedTemporaryFile(mode="w", suffix=".py", delete=False) as f:
            f.write("# @SPEC SPEC-AUTH-001\ndef func(): pass")
            temp_path = Path(f.name)

        try:
            result = parser.extract_tags_from_file(temp_path)
            assert len(result) == 1
            assert result[0].spec_id == "SPEC-AUTH-001"
        finally:
            temp_path.unlink()

    def test_extract_tags_from_file_with_string_path(self):
        """Test extract_tags_from_file with string path (auto-convert to Path)."""
        import tempfile

        with tempfile.NamedTemporaryFile(mode="w", suffix=".py", delete=False) as f:
            f.write("# @SPEC SPEC-AUTH-001")
            temp_path = f.name

        try:
            result = parser.extract_tags_from_file(temp_path)  # type: ignore[arg-type]
            assert len(result) == 1
            assert result[0].spec_id == "SPEC-AUTH-001"
        finally:
            Path(temp_path).unlink()

    def test_extract_tags_from_file_nonexistent_path(self):
        """Test extract_tags_from_file with nonexistent file."""
        result = parser.extract_tags_from_file(Path("nonexistent.py"))
        assert result == []

    def test_extract_tags_from_file_with_directory_path(self):
        """Test extract_tags_from_file with directory path."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            result = parser.extract_tags_from_file(Path(temp_dir))
            assert result == []

    def test_extract_tags_from_file_with_syntax_error(self):
        """Test extract_tags_from_file with Python syntax error."""
        import tempfile

        with tempfile.NamedTemporaryFile(mode="w", suffix=".py", delete=False) as f:
            f.write("# @SPEC SPEC-AUTH-001\ndef invalid syntax here")
            temp_path = Path(f.name)

        try:
            result = parser.extract_tags_from_file(temp_path)
            # Should return empty list on syntax error
            assert result == []
        finally:
            temp_path.unlink()

    def test_extract_tags_from_file_with_indentation_error(self):
        """Test extract_tags_from_file with indentation error."""
        import tempfile

        with tempfile.NamedTemporaryFile(mode="w", suffix=".py", delete=False) as f:
            f.write("# @SPEC SPEC-AUTH-001\ndef func():\n  badly indented")
            temp_path = Path(f.name)

        try:
            result = parser.extract_tags_from_file(temp_path)
            # Should return empty list on syntax error
            assert result == []
        finally:
            temp_path.unlink()

    def test_extract_tags_from_file_empty_file(self):
        """Test extract_tags_from_file with empty file."""
        import tempfile

        with tempfile.NamedTemporaryFile(mode="w", suffix=".py", delete=False) as f:
            temp_path = Path(f.name)

        try:
            result = parser.extract_tags_from_file(temp_path)
            assert result == []
        finally:
            temp_path.unlink()

    def test_extract_tags_from_file_with_only_comments(self):
        """Test extract_tags_from_file with only comments (no code)."""
        import tempfile

        with tempfile.NamedTemporaryFile(mode="w", suffix=".py", delete=False) as f:
            f.write("# @SPEC SPEC-AUTH-001\n# @SPEC SPEC-USER-002")
            temp_path = Path(f.name)

        try:
            result = parser.extract_tags_from_file(temp_path)
            assert len(result) == 2
        finally:
            temp_path.unlink()

    def test_extract_tags_from_file_read_permission_error(self):
        """Test extract_tags_from_file with read permission error."""
        import tempfile

        with tempfile.NamedTemporaryFile(mode="w", suffix=".py", delete=False) as f:
            f.write("# @SPEC SPEC-AUTH-001")
            temp_path = Path(f.name)

        # Make file unreadable
        temp_path.chmod(0o000)

        try:
            result = parser.extract_tags_from_file(temp_path)
            # Should return empty list on error
            assert result == []
        finally:
            # Restore permissions for cleanup
            temp_path.chmod(0o644)
            temp_path.unlink()

    def test_extract_tags_from_file_with_utf8_encoding(self):
        """Test extract_tags_from_file with UTF-8 encoded content."""
        import tempfile

        with tempfile.NamedTemporaryFile(mode="w", suffix=".py", delete=False, encoding="utf-8") as f:
            f.write("# @SPEC SPEC-AUTH-001 인증\n# @SPEC SPEC-USER-002")
            temp_path = Path(f.name)

        try:
            result = parser.extract_tags_from_file(temp_path)
            assert len(result) == 2
        finally:
            temp_path.unlink()

    def test_extract_tags_from_file_with_bom_encoding(self):
        """Test extract_tags_from_file with UTF-8 BOM encoding."""
        import tempfile

        with tempfile.NamedTemporaryFile(mode="wb", suffix=".py", delete=False) as f:
            # Write UTF-8 BOM + content
            content = "# @SPEC SPEC-AUTH-001\ndef func(): pass"
            f.write(content.encode("utf-8-sig"))
            temp_path = Path(f.name)

        try:
            result = parser.extract_tags_from_file(temp_path)
            # Python's read_text handles BOM, but ast.parse may have issues
            # The behavior depends on Python version, so we just verify it doesn't crash
            assert isinstance(result, list)
        finally:
            temp_path.unlink()


class TestExtractTagsFromFilesBatchProcessing:
    """Test extract_tags_from_files batch processing."""

    def test_extract_tags_from_files_empty_list(self):
        """Test extract_tags_from_files with empty list."""
        result = parser.extract_tags_from_files([])
        assert result == []

    def test_extract_tags_from_files_single_file(self):
        """Test extract_tags_from_files with single file."""
        import tempfile

        with tempfile.NamedTemporaryFile(mode="w", suffix=".py", delete=False) as f:
            f.write("# @SPEC SPEC-AUTH-001")
            temp_path = Path(f.name)

        try:
            result = parser.extract_tags_from_files([temp_path])
            assert len(result) == 1
            assert result[0].spec_id == "SPEC-AUTH-001"
        finally:
            temp_path.unlink()

    def test_extract_tags_from_files_multiple_files(self):
        """Test extract_tags_from_files with multiple files."""
        import tempfile

        temp_files = []
        try:
            for i in range(3):
                with tempfile.NamedTemporaryFile(mode="w", suffix=".py", delete=False) as f:
                    f.write(f"# @SPEC SPEC-AUTH-00{i+1}")
                    temp_files.append(Path(f.name))

            result = parser.extract_tags_from_files(temp_files)
            assert len(result) == 3

            # Verify all SPEC-IDs present
            spec_ids = {tag.spec_id for tag in result}
            assert "SPEC-AUTH-001" in spec_ids
            assert "SPEC-AUTH-002" in spec_ids
            assert "SPEC-AUTH-003" in spec_ids

        finally:
            for temp_path in temp_files:
                temp_path.unlink()

    def test_extract_tags_from_files_with_nonexistent_files(self):
        """Test extract_tags_from_files with some nonexistent files."""
        import tempfile

        with tempfile.NamedTemporaryFile(mode="w", suffix=".py", delete=False) as f:
            f.write("# @SPEC SPEC-AUTH-001")
            temp_path = Path(f.name)

        try:
            result = parser.extract_tags_from_files([
                temp_path,
                Path("nonexistent1.py"),
                Path("nonexistent2.py"),
            ])

            # Should only extract from existing file
            assert len(result) == 1
            assert result[0].spec_id == "SPEC-AUTH-001"
        finally:
            temp_path.unlink()

    def test_extract_tags_from_files_with_duplicates(self):
        """Test extract_tags_from_files handles duplicate files."""
        import tempfile

        with tempfile.NamedTemporaryFile(mode="w", suffix=".py", delete=False) as f:
            f.write("# @SPEC SPEC-AUTH-001")
            temp_path = Path(f.name)

        try:
            # Pass same file twice
            result = parser.extract_tags_from_files([temp_path, temp_path])

            # Should extract from both (duplicates preserved)
            assert len(result) == 2
            assert all(tag.spec_id == "SPEC-AUTH-001" for tag in result)
        finally:
            temp_path.unlink()

    def test_extract_tags_from_files_with_mixed_results(self):
        """Test extract_tags_from_files with mixed success/failure."""
        import tempfile

        temp_files = []
        try:
            # Valid file with TAGs
            with tempfile.NamedTemporaryFile(mode="w", suffix=".py", delete=False) as f:
                f.write("# @SPEC SPEC-AUTH-001")
                temp_files.append(Path(f.name))

            # Empty file
            with tempfile.NamedTemporaryFile(mode="w", suffix=".py", delete=False) as f:
                temp_files.append(Path(f.name))

            # File with syntax error
            with tempfile.NamedTemporaryFile(mode="w", suffix=".py", delete=False) as f:
                f.write("invalid syntax here")
                temp_files.append(Path(f.name))

            result = parser.extract_tags_from_files(temp_files)

            # Should only extract from valid file
            assert len(result) == 1
            assert result[0].spec_id == "SPEC-AUTH-001"

        finally:
            for temp_path in temp_files:
                temp_path.unlink()

    def test_extract_tags_from_files_preserves_file_paths(self):
        """Test that file paths are preserved in TAGs."""
        import tempfile

        temp_files = []
        try:
            for i in range(2):
                with tempfile.NamedTemporaryFile(mode="w", suffix=".py", delete=False) as f:
                    f.write(f"# @SPEC SPEC-AUTH-00{i+1}")
                    temp_path = Path(f.name)
                    temp_files.append(temp_path)

            result = parser.extract_tags_from_files(temp_files)

            # Verify file paths match
            assert len(result) == 2
            assert result[0].file_path == temp_files[0]
            assert result[1].file_path == temp_files[1]

        finally:
            for temp_path in temp_files:
                temp_path.unlink()

    def test_extract_tags_from_files_with_string_paths(self):
        """Test extract_tags_from_files with string paths."""
        import tempfile

        temp_files = []
        try:
            for i in range(2):
                with tempfile.NamedTemporaryFile(mode="w", suffix=".py", delete=False) as f:
                    f.write(f"# @SPEC SPEC-AUTH-00{i+1}")
                    temp_files.append(f.name)  # String, not Path

            result = parser.extract_tags_from_files(temp_files)

            assert len(result) == 2

        finally:
            for temp_path_str in temp_files:
                Path(temp_path_str).unlink()


class TestExtractTagsFromDirectoryTraversal:
    """Test extract_tags_from_directory traversal."""

    def test_extract_tags_from_directory_nonexistent_directory(self):
        """Test extract_tags_from_directory with nonexistent directory."""
        result = parser.extract_tags_from_directory(Path("nonexistent_dir/"))
        assert result == []

    def test_extract_tags_from_directory_with_file_path(self):
        """Test extract_tags_from_directory with file path instead of directory."""
        import tempfile

        with tempfile.NamedTemporaryFile(mode="w", suffix=".py", delete=False) as f:
            f.write("# @SPEC SPEC-AUTH-001")
            temp_path = Path(f.name)

        try:
            result = parser.extract_tags_from_directory(temp_path)
            assert result == []
        finally:
            temp_path.unlink()

    def test_extract_tags_from_directory_empty_directory(self):
        """Test extract_tags_from_directory with empty directory."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            result = parser.extract_tags_from_directory(Path(temp_dir))
            assert result == []

    def test_extract_tags_from_directory_single_file(self):
        """Test extract_tags_from_directory with single Python file."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            test_file = Path(temp_dir) / "test.py"
            test_file.write_text("# @SPEC SPEC-AUTH-001")

            result = parser.extract_tags_from_directory(Path(temp_dir))

            assert len(result) == 1
            assert result[0].spec_id == "SPEC-AUTH-001"

    def test_extract_tags_from_directory_multiple_files(self):
        """Test extract_tags_from_directory with multiple files."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            for i in range(3):
                test_file = Path(temp_dir) / f"test{i}.py"
                test_file.write_text(f"# @SPEC SPEC-AUTH-00{i+1}")

            result = parser.extract_tags_from_directory(Path(temp_dir))

            assert len(result) == 3

    def test_extract_tags_from_directory_recursive_true(self):
        """Test extract_tags_from_directory with recursive=True."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            # Create nested structure
            subdir = Path(temp_dir) / "subdir"
            subdir.mkdir()

            root_file = Path(temp_dir) / "root.py"
            root_file.write_text("# @SPEC SPEC-ROOT-001")

            sub_file = subdir / "sub.py"
            sub_file.write_text("# @SPEC SPEC-SUB-002")

            result = parser.extract_tags_from_directory(Path(temp_dir), recursive=True)

            assert len(result) == 2
            spec_ids = {tag.spec_id for tag in result}
            assert "SPEC-ROOT-001" in spec_ids
            assert "SPEC-SUB-002" in spec_ids

    def test_extract_tags_from_directory_recursive_false(self):
        """Test extract_tags_from_directory with recursive=False."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            # Create nested structure
            subdir = Path(temp_dir) / "subdir"
            subdir.mkdir()

            root_file = Path(temp_dir) / "root.py"
            root_file.write_text("# @SPEC SPEC-ROOT-001")

            sub_file = subdir / "sub.py"
            sub_file.write_text("# @SPEC SPEC-SUB-002")

            result = parser.extract_tags_from_directory(Path(temp_dir), recursive=False)

            # Should only find root file
            assert len(result) == 1
            assert result[0].spec_id == "SPEC-ROOT-001"

    def test_extract_tags_from_directory_with_custom_pattern(self):
        """Test extract_tags_from_directory with custom file pattern."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            # Create files with different extensions
            test_py = Path(temp_dir) / "test.py"
            test_py.write_text("# @SPEC SPEC-PY-001")

            test_txt = Path(temp_dir) / "test.txt"
            test_txt.write_text("# @SPEC SPEC-TXT-002")

            # Only match .py files
            result = parser.extract_tags_from_directory(Path(temp_dir), pattern="*.py")

            assert len(result) == 1
            assert result[0].spec_id == "SPEC-PY-001"

    def test_extract_tags_from_directory_with_deep_nesting(self):
        """Test extract_tags_from_directory with deeply nested directories."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            # Create deep nesting
            level1 = Path(temp_dir) / "level1"
            level2 = level1 / "level2"
            level3 = level2 / "level3"
            level3.mkdir(parents=True)

            deep_file = level3 / "deep.py"
            deep_file.write_text("# @SPEC SPEC-DEEP-003")

            result = parser.extract_tags_from_directory(Path(temp_dir), recursive=True)

            assert len(result) == 1
            assert result[0].spec_id == "SPEC-DEEP-003"
            assert result[0].file_path == deep_file

    def test_extract_tags_from_directory_with_non_python_files(self):
        """Test extract_tags_from_directory ignores non-Python files."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            # Create various file types
            (Path(temp_dir) / "test.py").write_text("# @SPEC SPEC-PY-001")
            (Path(temp_dir) / "test.txt").write_text("# @SPEC SPEC-TXT-002")
            (Path(temp_dir) / "test.md").write_text("# @SPEC SPEC-MD-003")
            (Path(temp_dir) / "README").write_text("# @SPEC SPEC-README-004")

            result = parser.extract_tags_from_directory(Path(temp_dir), pattern="*.py")

            # Should only extract from .py files
            assert len(result) == 1
            assert result[0].spec_id == "SPEC-PY-001"

    def test_extract_tags_from_directory_with_path_object(self):
        """Test extract_tags_from_directory with Path object."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            test_file = Path(temp_dir) / "test.py"
            test_file.write_text("# @SPEC SPEC-AUTH-001")

            # Pass as Path object
            result = parser.extract_tags_from_directory(Path(temp_dir))

            assert len(result) == 1

    def test_extract_tags_from_directory_with_string_path(self):
        """Test extract_tags_from_directory with string path."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            test_file = Path(temp_dir) / "test.py"
            test_file.write_text("# @SPEC SPEC-AUTH-001")

            # Pass as string
            result = parser.extract_tags_from_directory(str(temp_dir))  # type: ignore[arg-type]

            assert len(result) == 1

    def test_extract_tags_from_directory_with_symlinks(self):
        """Test extract_tags_from_directory with symbolic links."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            # Create actual file
            actual_file = Path(temp_dir) / "actual.py"
            actual_file.write_text("# @SPEC SPEC-ACTUAL-001")

            # Create symlink
            link_file = Path(temp_dir) / "link.py"
            try:
                link_file.symlink_to(actual_file)

                result = parser.extract_tags_from_directory(Path(temp_dir))

                # Should handle symlinks (behavior may vary by OS)
                assert len(result) >= 1
                spec_ids = {tag.spec_id for tag in result}
                assert "SPEC-ACTUAL-001" in spec_ids
            except OSError:
                # Symlinks not supported on this system
                pass

    def test_extract_tags_from_directory_with_permission_denied(self):
        """Test extract_tags_from_directory with permission denied subdirectory."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            # Create accessible file
            good_file = Path(temp_dir) / "good.py"
            good_file.write_text("# @SPEC SPEC-GOOD-001")

            # Create subdirectory with no permissions
            bad_dir = Path(temp_dir) / "bad_dir"
            bad_dir.mkdir()

            try:
                bad_dir.chmod(0o000)

                # Should still extract from accessible files
                result = parser.extract_tags_from_directory(Path(temp_dir), recursive=True)

                # At minimum, should have the good file
                assert any(tag.spec_id == "SPEC-GOOD-001" for tag in result)

            finally:
                # Restore permissions for cleanup
                bad_dir.chmod(0o755)
