# moai-alfred-interactive-questions - Quick Reference

_Last updated: 2025-10-23_

## AskUserQuestion API Essentials

### Minimal Template

```typescript
const answer = await AskUserQuestion({
  questions: [
    {
      question: "Your question here?",
      header: "Short label",  // max 12 chars
      multiSelect: false,     // or true for multi-select
      options: [
        {
          label: "Option 1",  // 1-5 words
          description: "Why you'd pick this option."
        },
        {
          label: "Option 2",
          description: "Alternative with different implications."
        }
      ]
    }
  ]
});

// Access result
const choice = answer["Short label"];  // "Option 1" or ["Option 1", "Option 2"]
```

---

## When to Ask (Checklist)

- [ ] User intent is ambiguous
- [ ] Multiple valid implementation paths exist
- [ ] Trade-off decision (speed vs quality, simple vs complete)
- [ ] Risky/destructive operation needs approval
- [ ] Feature scope needs definition

---

## Top 5 Patterns

| Pattern | Use When | Example |
|---------|----------|---------|
| **Implementation Approach** | "Add feature X" without specifics | New component vs extend existing |
| **Confirmation** | Destructive action (delete, migrate) | "Delete 50 records. Proceed?" |
| **Feature Selection** | "Which framework/library?" | next-intl vs react-i18next |
| **Multi-Select** | Independent combinable features | Unit + E2E + Visual tests |
| **Sequential** | Dependent questions (Q2 depends on Q1) | Enable auth? → Which provider? |

---

## Key Constraints

| What | Max | Why |
|------|-----|-----|
| Questions | 4 | Avoid fatigue |
| Options | 4 | Prevent choice paralysis |
| Header length | 12 chars | TUI layout |
| Label length | 5 words | Quick scanning |
| Nesting depth | 2 levels | Keep linear flow |

---

## Integration Points

### spec-builder (`/alfred:1-plan`)
Ask when SPEC title/scope is vague

### code-builder (`/alfred:2-run`)
Ask when implementation approach unclear or multiple paths exist

### doc-syncer (`/alfred:3-sync`)
Ask when sync mode or documentation scope unclear

---

## Error Handling

```typescript
// User cancels (ESC)
try {
  const answer = await AskUserQuestion({...});
} catch (error) {
  // Fallback or abort
}

// Validate custom input ("Other" option)
if (!VALID_OPTIONS.includes(answer["Header"])) {
  validateCustomInput(answer["Header"]);
}
```

---

## Best Practices (Quick)

✅ **DO**:
- Be specific ("Which database?" not "What should we use?")
- Provide context (file names, scope, impact)
- Explain trade-offs
- Order logically (safest first)
- Batch related questions

❌ **DON'T**:
- Overuse (only when ambiguous)
- Too many options (2-4 max)
- Skip descriptions
- Hide implications
- Make destructive default
- Manually add "Other" (auto-provided)

---

For full details, see **SKILL.md**
