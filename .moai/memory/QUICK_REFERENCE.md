# Quick Reference - Google Gemini SDK Migration

**Print this page or keep it open while implementing fixes**

---

## The 3-Step Fix

### Step 1: Update Imports (1 minute)
```python
# ❌ CHANGE FROM
import google.generativeai as genai

# ✅ CHANGE TO
from google import genai
from google.genai import types
```

### Step 2: Fix Client Init (2 minutes)
```python
# ❌ CHANGE FROM
genai.configure(api_key=api_key)
self.client = genai.Client(api_key=api_key)

# ✅ CHANGE TO
self.client = genai.Client(api_key=api_key)
```

### Step 3: Fix Config & API Call (5 minutes)
```python
# ❌ CHANGE FROM
response = self.client.models.generate_content(
    model=model_name,
    contents=[{"parts": [{"text": prompt}]}],
    config={
        "response_modalities": ["TEXT", "IMAGE"],
        "image_config": {
            "aspect_ratio": aspect_ratio,
            "image_size": resolution  # ❌ WRONG
        }
    },
)

# ✅ CHANGE TO
response = client.models.generate_content(
    model=model_name,
    contents=prompt,
    config=types.GenerateContentConfig(
        response_modalities=["IMAGE"],
        image_config=types.ImageConfig(
            aspect_ratio=aspect_ratio,
        ),
    ),
)
```

---

## Model Names

```python
'gemini-2.5-flash-image'      # Fast (5-15s), 1K resolution
'gemini-3-pro-image-preview'  # Quality (10-60s), 4K resolution
```

---

## Aspect Ratios (11 total)

```
1:1 (square)
2:3 (portrait)     3:2 (landscape)
3:4 (tall)         4:3 (wide)
4:5 (portrait)     5:4 (landscape)
9:16 (mobile)      16:9 (HD)
21:9 (ultrawide)
9:21 (ultrawide portrait)
```

---

## Response Parsing

```python
import base64
from PIL import Image
from io import BytesIO

for part in response.parts:
    if part.inline_data:
        image_bytes = base64.b64decode(part.inline_data.data)
        image = Image.open(BytesIO(image_bytes))
        image.save('output.png')
```

---

## Error Handling

```python
from google.api_core import exceptions

try:
    response = client.models.generate_content(...)
except exceptions.ResourceExhausted:
    print("Quota exceeded - wait and retry")
except exceptions.PermissionDenied:
    print("Invalid API key - check GEMINI_API_KEY")
except exceptions.InvalidArgument as e:
    print(f"Invalid parameter: {e}")
except Exception as e:
    print(f"Error: {e}")
```

---

## Validation

```python
# Validate before calling API
if model not in ['gemini-2.5-flash-image', 'gemini-3-pro-image-preview']:
    raise ValueError("Invalid model")

if aspect_ratio not in ['1:1', '16:9', '9:16', ...]:
    raise ValueError("Invalid aspect ratio")

if not isinstance(prompt, str) or len(prompt) < 10:
    raise ValueError("Prompt too short")
```

---

## Complete Minimal Example

```python
from google import genai
from google.genai import types
import base64
from PIL import Image
from io import BytesIO

# Initialize
client = genai.Client(api_key='YOUR_KEY')

# Generate
response = client.models.generate_content(
    model='gemini-2.5-flash-image',
    contents='A sunset over mountains',
    config=types.GenerateContentConfig(
        response_modalities=["IMAGE"],
        image_config=types.ImageConfig(aspect_ratio="16:9"),
    ),
)

# Extract image
for part in response.parts:
    if part.inline_data:
        image_bytes = base64.b64decode(part.inline_data.data)
        image = Image.open(BytesIO(image_bytes))
        image.save('output.png')
        break
```

---

## Installation

```bash
pip uninstall google-generativeai -y
pip install google-genai>=0.1.0
```

---

## Files to Change

**Target File**: `/Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-connector-nano-banana/modules/image_generator.py`

**Changes in**:
- Lines 20: Imports
- Lines 80-104: __init__ method
- Lines 155-179: generate() method config
- Lines 182-209: Response parsing
- Lines 330-352: edit() method config

---

## Testing Checklist (Quick)

- [ ] Import succeeds: `from google import genai`
- [ ] Client created: `genai.Client(api_key='...')`
- [ ] Flash model works: 5-15 seconds
- [ ] Pro model works: 10-60 seconds
- [ ] Image saved to disk
- [ ] All 11 aspect ratios work
- [ ] Errors caught properly

---

## Common Mistakes

### ❌ WRONG - Still using old SDK
```python
import google.generativeai as genai
client = genai.Client(api_key)  # AttributeError!
```

### ❌ WRONG - Using deprecated image_size
```python
image_config=types.ImageConfig(
    image_size="2K"  # REMOVED in v1beta
)
```

### ❌ WRONG - Nested contents dict
```python
contents=[{"parts": [{"text": prompt}]}]  # Overcomplicated
```

### ✅ RIGHT - New SDK patterns
```python
from google import genai
client = genai.Client(api_key)  # Works!

image_config=types.ImageConfig(
    aspect_ratio="16:9"  # Current field
)

contents=prompt  # Simple string
```

---

## Environment Variables

```bash
# Set this
export GEMINI_API_KEY="your-api-key"

# Or this
export GOOGLE_API_KEY="your-api-key"

# Then use
client = genai.Client()  # Auto-detects
```

---

## Time Estimates

| Task | Time |
|------|------|
| Read this page | 3 min |
| Update imports | 1 min |
| Fix client init | 2 min |
| Fix API calls | 5 min |
| Fix response parsing | 3 min |
| Test basic example | 5 min |
| Run full tests | 5 min |
| **Total** | **24 min** |

---

## Success Indicators

✅ When complete, you should see:
- No AttributeError
- Images generating
- No deprecated field warnings
- Tests passing
- Response parsed correctly

---

## If Stuck

1. Check: Are you using `from google import genai`?
2. Check: Do you have `types.GenerateContentConfig`?
3. Check: Did you remove `image_size`?
4. Check: Is your API key set?
5. Reference: Full guide in `gemini-sdk-migration-guide.md`

---

## Key Differences at a Glance

| Aspect | Old (0.8.5) | New (Unified) |
|--------|------------|---------------|
| Import | `import google.generativeai` | `from google import genai` |
| Client | ❌ Doesn't exist | ✅ `genai.Client(api_key)` |
| Config | Raw dict | `types.GenerateContentConfig()` |
| image_size | Supported | ❌ Removed (use aspect_ratio) |
| Error | AttributeError | Fixed! |

---

**Laminated Version**: Print and keep handy while implementing!

**Status**: Ready to use
**Last Updated**: 2025-11-26
