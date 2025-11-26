# Google Gemini SDK Migration Guide: 0.8.5 → Unified SDK

**Target File**: `/Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-connector-nano-banana/modules/image_generator.py`
**Current SDK**: google-generativeai 0.8.5 (deprecated)
**Target SDK**: google-genai (unified)
**Priority**: HIGH (EOL: August 31, 2025)

---

## Problem Statement

### Current Error
```
AttributeError: module 'google.generativeai' has no attribute 'Client'
```

### Root Cause
The code attempts to use `genai.Client(api_key=...)` which doesn't exist in `google.generativeai` 0.8.5. This pattern is correct for the NEW unified `google.genai` SDK, but the imports are still using the old SDK.

### Solution Overview
1. Switch from `google.generativeai` to `google.genai`
2. Update client initialization
3. Use proper type-safe configuration objects
4. Update response parsing for correct API structure

---

## Step-by-Step Migration

### Part 1: Update Imports (Lines 1-25)

**CURRENT (WRONG)**:
```python
import google.generativeai as genai
from google.api_core import exceptions
```

**FIXED**:
```python
from google import genai
from google.genai import types
from google.api_core import exceptions
```

**Why**:
- The unified SDK is `google.genai`, not `google.generativeai`
- `types` module provides `GenerateContentConfig` and `ImageConfig`
- `exceptions` module location remains the same

---

### Part 2: Update Client Initialization (Lines 80-104)

**CURRENT (WRONG)**:
```python
def __init__(self, api_key: Optional[str] = None):
    if api_key is None:
        api_key = os.getenv("GOOGLE_API_KEY")

    if not api_key:
        raise ValueError(
            "API key not found. Set GOOGLE_API_KEY environment variable "
            "or pass api_key parameter"
        )

    genai.configure(api_key=api_key)  # ❌ Not needed in new SDK
    self.client = genai.Client(api_key=api_key)  # ✅ Correct for new SDK!
    logger.info("Nano Banana Image Generator initialized")
```

**FIXED**:
```python
def __init__(self, api_key: Optional[str] = None):
    if api_key is None:
        api_key = os.getenv("GEMINI_API_KEY")  # New standard env var

    if not api_key:
        raise ValueError(
            "API key not found. Set GEMINI_API_KEY environment variable "
            "or pass api_key parameter"
        )

    # Only line needed - genai.configure() removed
    self.client = genai.Client(api_key=api_key)
    logger.info("Nano Banana Image Generator initialized")
```

**Changes**:
- Remove: `genai.configure(api_key=api_key)` (not used in unified SDK)
- Change env var: `GOOGLE_API_KEY` → `GEMINI_API_KEY` (new standard)
- Keep: `self.client = genai.Client(api_key=api_key)` (this is correct!)

---

### Part 3: Update Text-to-Image Generation (Lines 155-179)

**CURRENT (WRONG)**:
```python
# Line 156-179
try:
    model_name = self.MODELS[model]

    generation_config = {
        "response_modalities": ["TEXT", "IMAGE"],
        "image_config": {
            "aspect_ratio": aspect_ratio,
            "image_size": resolution  # ❌ Deprecated field!
        }
    }

    tools = []
    if use_google_search:
        tools = [{"google_search": {}}]

    response = self.client.models.generate_content(
        model=model_name,
        contents=[{"parts": [{"text": prompt}]}],  # ❌ Wrong format
        config=generation_config,  # ❌ Raw dict instead of types
        tools=tools if tools else None
    )
```

**FIXED**:
```python
try:
    model_name = self.MODELS[model]

    # Use proper type-safe configuration
    config = types.GenerateContentConfig(
        response_modalities=["IMAGE"],  # Image only, no TEXT needed
        image_config=types.ImageConfig(
            aspect_ratio=aspect_ratio,
            # Note: no image_size field in v1beta API
        ),
    )

    # Build contents properly
    response = self.client.models.generate_content(
        model=model_name,
        contents=prompt,  # Simple string format
        config=config,  # Use types.GenerateContentConfig
    )
```

**Key Changes**:
- Use `types.GenerateContentConfig()` instead of raw dict
- Remove `image_size` parameter (deprecated)
- Response modalities: `["IMAGE"]` only (text not needed)
- Contents: Simple string `prompt` instead of nested dict
- Remove `tools` parameter (Google Search not available in v1beta)

---

### Part 4: Update Response Parsing (Lines 182-209)

**CURRENT (MIGHT WORK BUT NEEDS VERIFICATION)**:
```python
image = None
description = ""

for part in response.candidates[0].content.parts:
    if hasattr(part, 'text') and part.text:
        description = part.text
    elif hasattr(part, 'inline_data') and part.inline_data:
        image_bytes = base64.b64decode(part.inline_data.data)
        from PIL import Image
        image = Image.open(BytesIO(image_bytes))

if not image:
    raise ValueError("No image data in response")

metadata = {
    "timestamp": datetime.now().isoformat(),
    "model": model,
    "resolution": resolution,
    "aspect_ratio": aspect_ratio,
    "prompt": prompt,
    "description": description,
    "finish_reason": response.candidates[0].finish_reason,
    "tokens_used": response.usage_metadata.total_token_count if hasattr(response, 'usage_metadata') else None,
    "use_google_search": use_google_search,
    "grounding_sources": []
}
```

**UPDATED**:
```python
image = None
description = ""

# response.parts is simpler in v1beta
for part in response.parts:
    if hasattr(part, 'text') and part.text:
        description = part.text
    elif hasattr(part, 'inline_data') and part.inline_data:
        image_bytes = base64.b64decode(part.inline_data.data)
        from PIL import Image
        image = Image.open(BytesIO(image_bytes))

if not image:
    raise ValueError("No image data in response")

# Build metadata (update to reflect actual fields)
metadata = {
    "timestamp": datetime.now().isoformat(),
    "model": model,
    "resolution": resolution,
    "aspect_ratio": aspect_ratio,
    "prompt": prompt,
    "description": description,
    "finish_reason": response.candidates[0].finish_reason if response.candidates else None,
    "tokens_used": response.usage_metadata.total_token_count if hasattr(response, 'usage_metadata') else None,
    "use_google_search": False,  # Not supported in current v1beta
    "grounding_sources": []  # Not available in current v1beta
}
```

**Changes**:
- Use `response.parts` instead of `response.candidates[0].content.parts`
- Remove Google Search handling (not in v1beta)
- Simplify metadata collection
- Add safety check for candidates list

---

### Part 5: Update Image-to-Image Editing (Lines 330-352)

**CURRENT (WRONG)**:
```python
response = self.client.models.generate_content(
    model=model_name,
    contents=[{
        "parts": [
            {
                "text": instruction
            },
            {
                "inline_data": {
                    "mime_type": mime_type,
                    "data": image_data
                }
            }
        ]
    }],
    config={
        "response_modalities": ["TEXT", "IMAGE"],
        "image_config": {
            "aspect_ratio": aspect_ratio,
            "image_size": resolution
        }
    }
)
```

**FIXED**:
```python
response = self.client.models.generate_content(
    model=model_name,
    contents=[
        types.Content(
            parts=[
                types.Part(text=instruction),
                types.Part(inline_data=types.Blob(
                    mime_type=mime_type,
                    data=image_data
                ))
            ]
        )
    ],
    config=types.GenerateContentConfig(
        response_modalities=["IMAGE"],
        image_config=types.ImageConfig(
            aspect_ratio=aspect_ratio,
        ),
    )
)
```

**Improvements**:
- Use proper type objects instead of dicts
- Remove `image_size` field
- Use `types.Content`, `types.Part`, `types.Blob` for type safety
- Simpler, more maintainable code

---

### Part 6: Update Error Handling (Lines 238-258)

**CURRENT**:
```python
except exceptions.ResourceExhausted:
    logger.error("API quota exceeded")
    print("❌ API 할당량 초과")
    print("   • 해상도를 1K로 다운그레이드하거나")
    print("   • 몇 분 후에 다시 시도하세요")
    raise

except exceptions.PermissionDenied:
    logger.error("Permission denied - check API key")
    print("❌ 권한 오류 - API 키 확인이 필요합니다")
    raise

except exceptions.InvalidArgument as e:
    logger.error(f"Invalid argument: {e}")
    print(f"❌ 잘못된 파라미터: {e}")
    raise
```

**UPDATED** (remains mostly the same, but add):
```python
except exceptions.ResourceExhausted:
    logger.error("API quota exceeded")
    print("❌ API 할당량 초과")
    print("   • 해상도를 1K로 다운그레이드하거나")
    print("   • 몇 분 후에 다시 시도하세요")
    raise

except exceptions.PermissionDenied:
    logger.error("Permission denied - check API key")
    print("❌ 권한 오류 - API 키 확인이 필요합니다")
    raise

except exceptions.InvalidArgument as e:
    logger.error(f"Invalid argument: {e}")
    print(f"❌ 잘못된 파라미터: {e}")
    raise

except exceptions.BadRequest as e:
    # New in v1beta - model validation errors
    logger.error(f"Bad request (model issue): {e}")
    print(f"❌ 모델 요청 오류: {e}")
    raise
```

---

## Complete Fixed Methods

### Fixed `__init__` Method
```python
def __init__(self, api_key: Optional[str] = None):
    """
    Initialize Nano Banana Image Generator

    Args:
        api_key: Google Gemini API key
                (if None, loads from environment variable)

    Example:
        >>> generator = NanoBananaImageGenerator()
        >>> # or
        >>> generator = NanoBananaImageGenerator("AIza...")
    """
    if api_key is None:
        api_key = os.getenv("GEMINI_API_KEY") or os.getenv("GOOGLE_API_KEY")

    if not api_key:
        raise ValueError(
            "API key not found. Set GEMINI_API_KEY or GOOGLE_API_KEY "
            "environment variable or pass api_key parameter"
        )

    self.client = genai.Client(api_key=api_key)
    logger.info("Nano Banana Image Generator initialized")
```

### Fixed `generate` Method (Lines 155-230)
```python
def generate(
    self,
    prompt: str,
    model: str = "flash",
    resolution: str = "2K",
    aspect_ratio: str = "16:9",
    use_google_search: bool = False,
    save_path: Optional[str] = None
) -> Tuple[Any, Dict[str, Any]]:
    """
    Text-to-Image 생성

    Args:
        prompt: 이미지 생성 프롬프트
        model: 모델 선택 ("flash" 또는 "pro")
        resolution: 해상도 ("1K", "2K", "4K")
        aspect_ratio: 종횡비 (기본: "16:9")
        use_google_search: Google Search 연동 여부 (현재 미지원)
        save_path: 이미지 저장 경로 (선택사항)

    Returns:
        Tuple[PIL.Image, Dict]: (생성된 이미지, 메타데이터)

    Raises:
        ValueError: 잘못된 파라미터
        Exception: API 호출 실패
    """
    # 파라미터 검증
    self._validate_params(model, resolution, aspect_ratio)

    print(f"\n{'='*70}")
    print(f"🎨 Nano Banana 이미지 생성 시작")
    print(f"{'='*70}")
    print(f"📝 프롬프트: {prompt[:50]}...")
    print(f"🎯 설정: {model.upper()} | {resolution} | {aspect_ratio}")
    print(f"⏳ 처리 중...\n")

    try:
        model_name = self.MODELS[model]

        # 요청 구성 - 새로운 타입 기반 설정
        config = types.GenerateContentConfig(
            response_modalities=["IMAGE"],
            image_config=types.ImageConfig(
                aspect_ratio=aspect_ratio,
            ),
        )

        # API 호출
        response = self.client.models.generate_content(
            model=model_name,
            contents=prompt,
            config=config,
        )

        # 응답 처리
        image = None
        description = ""

        for part in response.parts:
            if hasattr(part, 'text') and part.text:
                description = part.text
            elif hasattr(part, 'inline_data') and part.inline_data:
                # Base64 데이터를 PIL Image로 변환
                image_bytes = base64.b64decode(part.inline_data.data)
                from PIL import Image
                image = Image.open(BytesIO(image_bytes))

        if not image:
            raise ValueError("No image data in response")

        # 메타데이터 구성
        metadata = {
            "timestamp": datetime.now().isoformat(),
            "model": model,
            "resolution": resolution,
            "aspect_ratio": aspect_ratio,
            "prompt": prompt,
            "description": description,
            "finish_reason": response.candidates[0].finish_reason if response.candidates else None,
            "tokens_used": response.usage_metadata.total_token_count if hasattr(response, 'usage_metadata') else None,
            "use_google_search": False,  # Not supported in v1beta
            "grounding_sources": []
        }

        # 저장
        if save_path:
            Path(save_path).parent.mkdir(parents=True, exist_ok=True)
            image.save(save_path)
            metadata["saved_to"] = save_path
            print(f"✅ 이미지 저장: {save_path}\n")

        print(f"✅ 이미지 생성 완료!")
        print(f"   • 모델: {model.upper()}")
        print(f"   • 해상도: {resolution}")
        print(f"   • 토큰: {metadata['tokens_used']}")

        return image, metadata

    except exceptions.ResourceExhausted:
        logger.error("API quota exceeded")
        print("❌ API 할당량 초과")
        print("   • 해상도를 1K로 다운그레이드하거나")
        print("   • 몇 분 후에 다시 시도하세요")
        raise

    except exceptions.PermissionDenied:
        logger.error("Permission denied - check API key")
        print("❌ 권한 오류 - API 키 확인이 필요합니다")
        raise

    except exceptions.InvalidArgument as e:
        logger.error(f"Invalid argument: {e}")
        print(f"❌ 잘못된 파라미터: {e}")
        raise

    except Exception as e:
        logger.error(f"Error generating image: {e}")
        print(f"❌ 오류 발생: {e}")
        raise
```

---

## Testing Checklist

After applying changes, verify:

### Basic Functionality
- [ ] `NanoBananaImageGenerator(api_key)` initializes without error
- [ ] `generator.generate("test prompt")` produces an image
- [ ] Image is saved to disk correctly
- [ ] Metadata contains expected fields

### Model Testing
- [ ] Flash model generates images in ~5-15s
- [ ] Pro model generates images in ~10-60s
- [ ] 4K resolution supported with pro model
- [ ] 1K resolution works with flash model

### Aspect Ratio Testing
- [ ] 1:1 (square) works
- [ ] 16:9 (widescreen) works
- [ ] 9:16 (portrait) works
- [ ] All 11 aspect ratios supported

### Error Handling
- [ ] Missing API key raises clear error
- [ ] Invalid model name raises error
- [ ] Invalid aspect ratio raises error
- [ ] API errors caught and logged properly

### Response Validation
- [ ] Image data correctly decoded from Base64
- [ ] PIL Image opens without error
- [ ] Metadata timestamps are valid
- [ ] Token counts are accurate

---

## Comparison: Old vs New API

| Feature | Old (0.8.5) | New (Unified) |
|---------|------------|---------------|
| Package | google-generativeai | google-genai |
| Import | `import google.generativeai as genai` | `from google import genai` |
| Client Init | `genai.Client()` doesn't exist | `genai.Client(api_key=...)` ✅ |
| Config Type | Raw dict | `types.GenerateContentConfig()` |
| Image Config | dict with `image_size` | `types.ImageConfig()` with `aspect_ratio` |
| Response Parts | `response.candidates[0].content.parts` | `response.parts` |
| Error Classes | `google.api_core.exceptions` | Same, no change |
| Google Search | Supported in dict | Not in v1beta |

---

## Expected Results

After migration:

### Before (Broken)
```
AttributeError: module 'google.generativeai' has no attribute 'Client'
```

### After (Fixed)
```
🎨 Nano Banana 이미지 생성 시작
==================================================================
📝 프롬프트: A serene mountain landscape at golden hour...
🎯 설정: FLASH | 2K | 16:9
⏳ 처리 중...

✅ 이미지 생성 완료!
   • 모델: FLASH
   • 해상도: 2K
   • 토큰: 1234

✅ 이미지 저장: output.png
```

---

## Installation Instructions

### 1. Update requirements.txt

```bash
# Remove old SDK
pip uninstall google-generativeai -y

# Install new SDK
pip install google-genai>=0.1.0
```

### 2. Update imports in image_generator.py

Replace at top of file:
```python
from google import genai
from google.genai import types
```

### 3. Test basic functionality

```python
from modules.image_generator import NanoBananaImageGenerator

api_key = "your-api-key"
generator = NanoBananaImageGenerator(api_key)

# Should work without errors
image, metadata = generator.generate(
    "A test image",
    model="flash",
    resolution="2K",
    aspect_ratio="16:9"
)

print(f"✅ Success! Generated image: {metadata}")
```

---

## Deployment Checklist

- [ ] Install new google-genai package
- [ ] Update all import statements
- [ ] Fix client initialization
- [ ] Update generate_content calls with types.GenerateContentConfig
- [ ] Remove deprecated image_size parameter
- [ ] Test with both models
- [ ] Update documentation
- [ ] Update examples
- [ ] Run full test suite
- [ ] Verify error handling
- [ ] Monitor initial production usage

---

**Status**: Ready for Implementation
**Estimated Time**: 30-45 minutes
**Complexity**: Medium (API changes, but straightforward fixes)
**Risk Level**: Low (well-documented new SDK, clear migration path)
