"""Tests for TAG parser module (T2: TAG Parser - Comment Extraction)."""

from pathlib import Path

import pytest

from moai_adk.tag_system import parser as tag_parser


class TestTAGExtraction:
    """Test TAG extraction from Python source code (T2: TAG Parser)."""

    def test_extract_single_tag(self):
        """Test extracting single TAG from Python code."""
        source = """
# @SPEC SPEC-AUTH-001
def authenticate_user():
    pass
"""

        tags = tag_parser.extract_tags_from_source(source, Path("auth.py"))

        assert len(tags) == 1
        assert tags[0].spec_id == "SPEC-AUTH-001"
        assert tags[0].verb == "impl"
        assert tags[0].line == 2

    def test_extract_tag_with_verb(self):
        """Test extracting TAG with explicit verb."""
        source = """
# @SPEC SPEC-AUTH-001 verify
def test_auth_flow():
    pass
"""

        tags = tag_parser.extract_tags_from_source(source, Path("test_auth.py"))

        assert len(tags) == 1
        assert tags[0].spec_id == "SPEC-AUTH-001"
        assert tags[0].verb == "verify"

    def test_extract_multiple_tags(self):
        """Test extracting multiple TAGs from same file."""
        source = """
# @SPEC SPEC-AUTH-001
def login():
    pass

# @SPEC SPEC-AUTH-002 verify
def test_login():
    pass
"""

        tags = tag_parser.extract_tags_from_source(source, Path("auth.py"))

        assert len(tags) == 2
        assert tags[0].spec_id == "SPEC-AUTH-001"
        assert tags[1].spec_id == "SPEC-AUTH-002"

    def test_extract_tag_with_description(self):
        """Test extracting TAG with descriptive text."""
        source = """
# @SPEC SPEC-AUTH-001 impl - User authentication flow
def authenticate():
    pass
"""

        tags = tag_parser.extract_tags_from_source(source, Path("auth.py"))

        assert len(tags) == 1
        assert tags[0].spec_id == "SPEC-AUTH-001"

    def test_extract_no_tags(self):
        """Test extracting from code without TAGs."""
        source = """
def regular_function():
    pass

# Just a regular comment
def another_function():
    pass
"""

        tags = tag_parser.extract_tags_from_source(source, Path("test.py"))

        assert len(tags) == 0

    def test_extract_tags_with_different_verbs(self):
        """Test extracting TAGs with different verbs."""
        source = """
# @SPEC SPEC-AUTH-001 impl
# @SPEC SPEC-AUTH-002 verify
# @SPEC SPEC-AUTH-003 depends
# @SPEC SPEC-AUTH-004 related
def complex_function():
    pass
"""

        tags = tag_parser.extract_tags_from_source(source, Path("test.py"))

        assert len(tags) == 4
        assert tags[0].verb == "impl"
        assert tags[1].verb == "verify"
        assert tags[2].verb == "depends"
        assert tags[3].verb == "related"


class TestTAGExtractionEdgeCases:
    """Test TAG extraction edge cases."""

    def test_tag_in_string_literal_ignored(self):
        """Test that TAGs in string literals are ignored."""
        source = '''
def help_text():
    """This is a @SPEC SPEC-AUTH-001 in a docstring."""
    return "Usage: @SPEC SPEC-AUTH-001"
'''

        tags = tag_parser.extract_tags_from_source(source, Path("test.py"))

        assert len(tags) == 0

    def test_tag_with_unicode_characters(self):
        """Test TAG extraction with Unicode characters."""
        source = """
# @SPEC SPEC-I18N-001 한글 태그
def internationalized():
    pass
"""

        tags = tag_parser.extract_tags_from_source(source, Path("test.py"))

        assert len(tags) == 1
        assert tags[0].spec_id == "SPEC-I18N-001"

    def test_duplicate_tags_recorded(self):
        """Test that duplicate TAGs are recorded (T2.1)."""
        source = """
# @SPEC SPEC-AUTH-001
def func1():
    pass

# @SPEC SPEC-AUTH-001
def func2():
    pass
"""

        tags = tag_parser.extract_tags_from_source(source, Path("test.py"))

        assert len(tags) == 2
        assert tags[0].spec_id == "SPEC-AUTH-001"
        assert tags[1].spec_id == "SPEC-AUTH-001"

    def test_invalid_tag_format_ignored(self):
        """Test that invalid TAG formats are ignored."""
        source = """
# @spec auth-001  # Wrong format
def bad_example():
    pass

# @SPEC SPEC-AUTH-001  # Valid format
def good_example():
    pass
"""

        tags = tag_parser.extract_tags_from_source(source, Path("test.py"))

        assert len(tags) == 1
        assert tags[0].spec_id == "SPEC-AUTH-001"

    def test_empty_source(self):
        """Test TAG extraction from empty source."""
        source = ""

        tags = tag_parser.extract_tags_from_source(source, Path("test.py"))

        assert len(tags) == 0

    def test_source_with_only_comments(self):
        """Test source with only comments and TAGs."""
        source = """
# @SPEC SPEC-AUTH-001
# Another comment
# @SPEC SPEC-AUTH-002 verify
"""

        tags = tag_parser.extract_tags_from_source(source, Path("test.py"))

        assert len(tags) == 2


class TestTAGExtractionFromFile:
    """Test TAG extraction from files (T2.2)."""

    def test_extract_from_file(self, tmp_path):
        """Test extracting TAGs from a Python file."""
        test_file = tmp_path / "test.py"
        test_file.write_text("""
# @SPEC SPEC-AUTH-001
def login():
    pass

# @SPEC SPEC-AUTH-002 verify
def test_login():
    pass
""")

        tags = tag_parser.extract_tags_from_file(test_file)

        assert len(tags) == 2
        assert tags[0].file_path == test_file
        assert tags[1].file_path == test_file

    def test_extract_from_nonexistent_file(self, tmp_path):
        """Test extracting from nonexistent file returns empty list."""
        nonexistent = tmp_path / "nonexistent.py"

        tags = tag_parser.extract_tags_from_file(nonexistent)

        assert tags == []

    def test_extract_from_file_with_syntax_error(self, tmp_path):
        """Test extracting from file with syntax errors (T2.2)."""
        test_file = tmp_path / "bad_syntax.py"
        test_file.write_text("""
# @SPEC SPEC-AUTH-001
def incomplete_function(
    # Missing closing parenthesis
""")

        # Should return empty list, not raise exception
        tags = tag_parser.extract_tags_from_file(test_file)

        assert tags == []

    def test_extract_from_multiple_files(self, tmp_path):
        """Test extracting TAGs from multiple files."""
        file1 = tmp_path / "file1.py"
        file2 = tmp_path / "file2.py"

        file1.write_text("# @SPEC SPEC-AUTH-001\ndef f(): pass")
        file2.write_text("# @SPEC SPEC-TAG-002\ndef g(): pass")

        tags = tag_parser.extract_tags_from_files([file1, file2])

        assert len(tags) == 2
        assert any(t.file_path == file1 for t in tags)
        assert any(t.file_path == file2 for t in tags)


class TestCommentHandling:
    """Test comment extraction and handling."""

    def test_inline_comment_tag(self):
        """Test TAG in inline comment."""
        source = "def func(): pass  # @SPEC SPEC-AUTH-001"

        tags = tag_parser.extract_tags_from_source(source, Path("test.py"))

        # Should extract the TAG
        assert len(tags) == 1
        assert tags[0].spec_id == "SPEC-AUTH-001"

    def test_block_comment_tag(self):
        """Test TAG in block comment."""
        source = '''
"""
@SPEC SPEC-AUTH-001
This is a block comment
"""
def func():
    pass
'''

        tags = tag_parser.extract_tags_from_source(source, Path("test.py"))

        # May or may not extract depending on implementation
        # At minimum should not crash

    def test_shebang_and_encoding(self):
        """Test file with shebang and encoding declaration."""
        source = """#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# @SPEC SPEC-AUTH-001
def authenticate():
    pass
"""

        tags = tag_parser.extract_tags_from_source(source, Path("auth.py"))

        assert len(tags) == 1
        assert tags[0].spec_id == "SPEC-AUTH-001"


class TestPerformance:
    """Test performance for large files."""

    def test_large_file_performance(self, tmp_path):
        """Test TAG extraction from large file (EC-2.3)."""
        import time

        # Create file with 1000 TAGs (using valid SPEC-ID format)
        lines = ["# @SPEC SPEC-PERF-{:03d}\n".format(i) for i in range(1000)]
        test_file = tmp_path / "large.py"
        test_file.write_text("\n".join(lines))

        start = time.time()
        tags = tag_parser.extract_tags_from_file(test_file)
        elapsed = time.time() - start

        assert len(tags) == 1000
        assert elapsed < 5.0  # Should complete in 5 seconds

    def test_large_codebase_performance(self, tmp_path):
        """Test TAG extraction from multiple files."""
        import time

        # Create 100 files with TAGs (using valid SPEC-ID format)
        files = []
        for i in range(100):
            test_file = tmp_path / f"file{i}.py"
            test_file.write_text(f"# @SPEC SPEC-PERF-{i:03d}\ndef f(): pass")
            files.append(test_file)

        start = time.time()
        all_tags = tag_parser.extract_tags_from_files(files)
        elapsed = time.time() - start

        assert len(all_tags) == 100
        assert elapsed < 10.0  # Should complete in 10 seconds
