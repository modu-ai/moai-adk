# Integration with Alfred Workflow

Patterns for integrating personas into Alfred's decision workflow.

## Alfred Integration

```python
# Step 1: Select persona
persona = select_persona(
    user_request,
    session_context,
    project_config
)

# Step 2: Adapt to context
adapted_persona = adapt_to_project_context(
    persona,
    project_context
)

# Step 3: Generate response
response = adapted_persona.communicate(topic)

# Step 4: Track feedback
if user_feedback:
    update_persona_preferences(
        user_id, persona, feedback
    )
```

## AskUserQuestion Integration

```python
# Technical Mentor (educational options)
AskUserQuestion(
    question="Feature type?",
    options=[
        {"label": "Learn types first", "desc": "Examples"},
        {"label": "Simple feature", "desc": "Beginner"},
        {"label": "Help decide", "desc": "Guidance"}
    ]
)

# Efficiency Coach (direct options)
AskUserQuestion(
    question="Type?",
    options=[
        {"label": "User feature", "desc": "Frontend"},
        {"label": "API feature", "desc": "Backend"}
    ]
)
```

---

**Integration Points**:
- User request analysis
- AskUserQuestion formatting
- Feedback collection
- Session persistence
