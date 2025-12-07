# Core Responsibilities Module

Detailed implementation patterns for the four core responsibilities of yoda-content-generator.

## 1. Knowledge Fetching (지식 수집)

**Purpose**: Fetch latest documentation and best practices from multiple sources in parallel

**Key Features**:
- Context7 MCP + WebSearch parallel execution
- Automatic deduplication (remove duplicate content)
- Relevance ranking (score 0-100)
- Caching (30 days stable, 7 days beta)

**Usage Pattern**:

```python
from generators.context7_injector import Context7Injector

injector = Context7Injector()

# Parallel fetch from Context7 + WebSearch
knowledge = await injector.parallel_fetch(
    topic="React 19 Hooks",
    sources=["context7", "websearch"]
)

# Returns:
# {
#   "context7": {...},
#   "websearch": [...],
#   "deduplicated": [...],
#   "ranked": [...]
# }
```

**Implementation Details**:

```python
class Context7Injector:
    """Context7 MCP + WebSearch parallel knowledge fetching."""
    
    async def parallel_fetch(
        self,
        topic: str,
        sources: List[str] = ["context7", "websearch"]
    ) -> Dict[str, Any]:
        """Fetch knowledge from multiple sources in parallel."""
        
        tasks = []
        
        if "context7" in sources:
            tasks.append(self.fetch_from_context7(topic))
        
        if "websearch" in sources:
            tasks.append(self.fetch_from_websearch(topic))
        
        results = await asyncio.gather(*tasks, return_exceptions=True)
        
        # Deduplicate and rank
        deduplicated = self.deduplicate_knowledge(results)
        ranked = self.rank_by_relevance(deduplicated, topic)
        
        return {
            "context7": results[0] if "context7" in sources else None,
            "websearch": results[1] if "websearch" in sources else None,
            "deduplicated": deduplicated,
            "ranked": ranked
        }
```

---

## 2. Content Expansion (콘텐츠 확장)

**Purpose**: Transform Plan agent results into detailed, comprehensive sections

**Key Features**:
- Plan → 25 sections (4 pages each = 100 pages)
- Template-based expansion (education/presentation/workshop)
- Knowledge injection (Context7 + WebSearch)
- Section validation (structure, quality, completeness)

**Usage Pattern**:

```python
from generators.plan_expander import PlanExpander

expander = PlanExpander()

# Expand plan to sections
sections = await expander.expand_plan_to_sections(
    plan=plan_result,
    template="education",
    knowledge=knowledge_fetched
)

# Returns: List[Section]
# Section = {
#   "title": "Introduction to React Hooks",
#   "content": "...",
#   "subsections": [...],
#   "code_examples": [...],
#   "diagrams": [...]
# }
```

**Expansion Strategy**:

```
Plan Agent Output (High-level):
├─ Section 1: Introduction (1 paragraph)
├─ Section 2: Core Concepts (3 bullets)
└─ Section 3: Advanced Topics (2 bullets)

↓ Plan Expander

Detailed Sections (Lecture-ready):
├─ Section 1: Introduction (4 pages)
│   ├─ 1.1 Overview (1 page)
│   ├─ 1.2 Learning Objectives (1 page)
│   ├─ 1.3 Prerequisites (1 page)
│   └─ 1.4 What You'll Learn (1 page)
│
├─ Section 2: Core Concepts (12 pages)
│   ├─ 2.1 Concept A (4 pages)
│   ├─ 2.2 Concept B (4 pages)
│   └─ 2.3 Concept C (4 pages)
│
└─ Section 3: Advanced Topics (8 pages)
    ├─ 3.1 Topic A (4 pages)
    └─ 3.2 Topic B (4 pages)

Total: 25 sections × 4 pages = 100 pages
```

---

## 3. Batch Processing (배치 처리)

**Purpose**: Generate sections in parallel batches for 70% performance improvement

**Key Features**:
- Batch size: 5 sections parallel (configurable)
- Automatic retry on failure
- Progress tracking
- Memory optimization

**Usage Pattern**:

```python
from generators.section_builder import SectionBuilder

builder = SectionBuilder()

# Build sections in batches
generated_sections = await builder.build_sections_batch(
    sections=sections,
    batch_size=5,
    pages_target=100
)

# Progress:
# Batch 1/5: [====    ] 5/25 sections (20%)
# Batch 2/5: [========] 10/25 sections (40%)
# ...
# Batch 5/5: [========] 25/25 sections (100%)
```

**Implementation Strategy**:

```python
class SectionBuilder:
    """Batch section generation with parallel processing."""
    
    async def build_sections_batch(
        self,
        sections: List[Dict],
        batch_size: int = 5
    ) -> List[str]:
        """Build sections in parallel batches."""
        
        # Split into batches
        batches = self.split_into_batches(sections, batch_size)
        
        # Process each batch in parallel
        all_generated = []
        
        for i, batch in enumerate(batches):
            print(f"Batch {i+1}/{len(batches)}: Processing {len(batch)} sections...")
            
            # Parallel generation within batch
            batch_results = await asyncio.gather(
                *[self.generate_section(section) for section in batch],
                return_exceptions=True
            )
            
            # Handle exceptions
            for j, result in enumerate(batch_results):
                if isinstance(result, Exception):
                    print(f"Error in section {batch[j]['title']}: {result}")
                    # Retry once
                    result = await self.generate_section(batch[j])
            
            all_generated.extend(batch_results)
        
        return all_generated
```

**Performance Comparison**:

```
Sequential Processing (1 at a time):
Section 1: [====] 30 sec
Section 2: [====] 30 sec
...
Section 25: [====] 30 sec
Total: 25 × 30 sec = 12.5 minutes

Batch Processing (5 parallel):
Batch 1 (5 sections): [====] 30 sec
Batch 2 (5 sections): [====] 30 sec
...
Batch 5 (5 sections): [====] 30 sec
Total: 5 × 30 sec = 2.5 minutes (80% faster!)
```

---

## 4. Diagram Embedding (다이어그램 임베딩)

**Purpose**: Convert Mermaid code blocks to inline SVG for portable documents

**Key Features**:
- Find all ```mermaid blocks
- Render to SVG using mermaid.js
- Embed inline (no external dependencies)
- Fallback to code block on error

**Usage Pattern**:

```python
from generators.mermaid_embedder import MermaidEmbedder

embedder = MermaidEmbedder()

# Embed all diagrams
final_content = await embedder.embed_diagrams_inline(
    content=markdown_content
)

# Before:
# ```mermaid
# graph TD
#     A[Start] --> B[End]
# ```

# After:
# <svg xmlns="http://www.w3.org/2000/svg" width="200" height="100">
#   <g>...</g>
# </svg>
```

**Implementation**:

```python
class MermaidEmbedder:
    """Mermaid diagram inline SVG embedding."""
    
    async def embed_diagrams_inline(self, content: str) -> str:
        """Find and embed all Mermaid diagrams as inline SVG."""
        
        # Find all mermaid code blocks
        mermaid_blocks = self.find_mermaid_blocks(content)
        
        # Render each to SVG
        for block in mermaid_blocks:
            try:
                svg = await self.render_mermaid_to_svg(block['code'])
                
                # Replace code block with inline SVG
                content = content.replace(
                    block['original'],
                    f"\n{svg}\n"
                )
            except Exception as e:
                print(f"Diagram rendering failed: {e}")
                # Keep original code block as fallback
        
        return content
    
    async def render_mermaid_to_svg(self, mermaid_code: str) -> str:
        """Render Mermaid code to SVG using mermaid.js."""
        
        # Use mermaid-cli (mmdc) or mermaid.js API
        # Example: subprocess call to mmdc
        import subprocess
        
        result = subprocess.run(
            ["mmdc", "-i", "-", "-o", "-", "-f", "svg"],
            input=mermaid_code.encode(),
            capture_output=True
        )
        
        return result.stdout.decode()
```
