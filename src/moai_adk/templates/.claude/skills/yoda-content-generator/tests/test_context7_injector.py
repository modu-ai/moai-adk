#!/usr/bin/env python3
"""
Unit tests for Context7Injector

Author: GoosLab
Version: 1.0.0
"""

import pytest
import asyncio
from generators.context7_injector import Context7Injector, KnowledgeSource


@pytest.mark.asyncio
async def test_parallel_fetch():
    """Test parallel knowledge fetching."""
    injector = Context7Injector()
    
    knowledge = await injector.parallel_fetch(
        topic="React 19",
        sources=["context7", "websearch"]
    )
    
    assert 'context7' in knowledge
    assert 'websearch' in knowledge
    assert 'deduplicated' in knowledge
    assert 'ranked' in knowledge


@pytest.mark.asyncio
async def test_deduplicate_knowledge():
    """Test knowledge deduplication."""
    injector = Context7Injector()
    
    sources = [
        KnowledgeSource(source="test1", content="duplicate content", relevance=90),
        KnowledgeSource(source="test2", content="duplicate content", relevance=80),
        KnowledgeSource(source="test3", content="unique content", relevance=70)
    ]
    
    deduplicated = injector.deduplicate_knowledge(sources)
    
    # Should remove duplicate content
    assert len(deduplicated) == 2


@pytest.mark.asyncio
async def test_rank_by_relevance():
    """Test relevance ranking."""
    injector = Context7Injector()
    
    sources = [
        KnowledgeSource(source="test1", content="content1", relevance=50),
        KnowledgeSource(source="test2", content="content2", relevance=90),
        KnowledgeSource(source="test3", content="content3", relevance=70)
    ]
    
    ranked = injector.rank_by_relevance(sources, "test")
    
    # Should be sorted by relevance (descending)
    assert ranked[0].relevance == 90
    assert ranked[1].relevance == 70
    assert ranked[2].relevance == 50


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
