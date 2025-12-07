#!/usr/bin/env python3
"""
Unit tests for SectionBuilder

Author: GoosLab
Version: 1.0.0
"""

import pytest
import asyncio
from generators.section_builder import SectionBuilder, GenerationResult


@pytest.mark.asyncio
async def test_build_sections_batch():
    """Test batch section generation."""
    builder = SectionBuilder(batch_size=2)
    
    sections = [
        {"title": f"Section {i}", "content": f"Content {i}"}
        for i in range(1, 6)  # 5 sections
    ]
    
    generated = await builder.build_sections_batch(sections)
    
    assert len(generated) == 5
    for content in generated:
        assert len(content) > 0


@pytest.mark.asyncio
async def test_split_into_batches():
    """Test batch splitting."""
    builder = SectionBuilder(batch_size=3)
    
    sections = list(range(10))  # 10 items
    
    batches = builder._split_into_batches(sections)
    
    # Should create 4 batches: [3, 3, 3, 1]
    assert len(batches) == 4
    assert len(batches[0]) == 3
    assert len(batches[1]) == 3
    assert len(batches[2]) == 3
    assert len(batches[3]) == 1


@pytest.mark.asyncio
async def test_validate_section_structure():
    """Test section validation."""
    builder = SectionBuilder()
    
    # Valid section
    valid = "# Title\n\nContent goes here..."
    assert builder.validate_section_structure(valid) is True
    
    # Invalid: Empty
    invalid_empty = ""
    assert builder.validate_section_structure(invalid_empty) is False
    
    # Invalid: No title
    invalid_no_title = "Content without title"
    assert builder.validate_section_structure(invalid_no_title) is False
    
    # Invalid: Too short
    invalid_short = "# Title"
    assert builder.validate_section_structure(invalid_short) is False


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
