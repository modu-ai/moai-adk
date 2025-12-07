# Yoda Content Generator Examples

Real-world usage scenarios demonstrating the content generator.

## Example 1: React 19 Lecture Generation

```python
# Step 1: Plan agent creates outline
plan = {
    "topic": "React 19 최신 기능",
    "sections": [
        {"title": "Introduction", "subsections": [...]},
        {"title": "Core Concepts", "subsections": [...]},
        {"title": "Advanced Topics", "subsections": [...]}
    ]
}

# Step 2: Content generator expands
knowledge = await fetch_knowledge_parallel(
    topic="React 19",
    sources=["context7", "websearch"]
)

sections = await expand_plan_to_sections(
    plan=plan,
    template="education",
    knowledge=knowledge
)

# Step 3: Batch generate
generated = await build_sections_batch(
    sections=sections,
    batch_size=5
)

# Step 4: Embed diagrams
final_content = await embed_diagrams_inline(generated)

# Result: 100+ page lecture in 5 minutes
```

---

## Example 2: Large-Scale Workshop Material

```python
# Generate 200+ page workshop material
plan = {
    "topic": "Docker Container Optimization",
    "sections": [
        # 50 sections × 4 pages = 200 pages
    ]
}

# Parallel fetch (1 min)
knowledge = await fetch_knowledge_parallel("Docker Optimization")

# Expand plan (1 min)
sections = await expand_plan_to_sections(plan, "workshop", knowledge)

# Batch generate (6 min for 50 sections, batch_size=5)
generated = await build_sections_batch(sections, batch_size=5)

# Embed diagrams (1 min)
final_content = await embed_diagrams_inline(generated)

# Total: 9 minutes for 200 pages
```
