#!/usr/bin/env python3
"""
Unit tests for MermaidEmbedder

Author: GoosLab
Version: 1.0.0
"""

import pytest
import asyncio
from generators.mermaid_embedder import MermaidEmbedder, MermaidBlock


@pytest.mark.asyncio
async def test_find_mermaid_blocks():
    """Test finding Mermaid code blocks."""
    embedder = MermaidEmbedder()
    
    content = """
# Test Document

```mermaid
graph TD
    A --> B
```

Some text

```mermaid
sequenceDiagram
    User->>API: Request
```
"""
    
    blocks = embedder.find_mermaid_blocks(content)
    
    assert len(blocks) == 2
    assert "graph TD" in blocks[0].code
    assert "sequenceDiagram" in blocks[1].code


@pytest.mark.asyncio
async def test_extract_diagram_metadata():
    """Test diagram metadata extraction."""
    embedder = MermaidEmbedder()
    
    # Flowchart
    flowchart_code = "graph TD\n    A --> B"
    metadata = embedder.extract_diagram_metadata(flowchart_code)
    assert metadata['type'] == 'flowchart'
    
    # Sequence
    sequence_code = "sequenceDiagram\n    User->>API: Request"
    metadata = embedder.extract_diagram_metadata(sequence_code)
    assert metadata['type'] == 'sequence'


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
