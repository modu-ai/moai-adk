# Alfred Reporting Style

Guidelines for user-facing screen output vs. internal documentation file creation.

## Purpose

Ensures consistent communication patterns that distinguish between real-time user updates and persistent documentation.

## Core Principle

**Screen Output**: User-facing, temporary, scannable
**File Documentation**: Internal, persistent, detailed

## Quick Guidelines

- **Screen**: Use ✅⚠️❌⏳ indicators, plain text, concise
- **Files**: Use markdown structure, detailed analysis, proper locations
- **Never** create analytical reports in project root
- **Always** follow document management rules

## Usage

```python
# For user progress updates
print("✅ Task completed")

# For internal documentation  
write_to_moai_docs("analysis-results.md", content)
```

## Files Structure

- `SKILL.md` - Complete reporting guidelines
- `examples.md` - Format and location examples  
- `reference.md` - Technical reference
- `README.md` - This overview
