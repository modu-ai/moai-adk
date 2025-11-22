# 🍌 Nano Banana Pro Skill

**Google Nano Banana Pro Image Generation Expert Skill**

An enterprise-grade Skill providing text-to-image generation, image editing, and advanced prompt optimization.

**Version**: 1.0.0
**Status**: Production Ready
**API**: Gemini 3 Pro Image Preview
**License**: MIT

---

## 🎯 Key Features

### ✨ Image Generation (Text-to-Image)
- **Natural Language Prompts**: Intuitive support for Korean, English, Japanese, and Chinese prompts
- **Resolution Options**: 1K (1024x1024), 2K (2048x2048), 4K (4096x4096)
- **Style Templates**: photorealistic, artistic, cinematic, minimal, abstract, fantasy
- **Auto Optimization**: Automatic prompt improvement and photography element addition
- **SynthID Watermark**: Automatic AI-generated content marking

### 🖼️ Image Editing (Image-to-Image)
- **Image Transformation**: Modify style/content based on existing images
- **Multi-turn Editing**: Step-by-step image refinement
- **URL/File Support**: Accept web images or local files
- **Format Support**: JPEG, PNG, GIF, WebP

### 🔍 Prompt Optimization
- **Auto Language Detection**: Automatically recognize input language
- **Style Enhancement**: Automatically add photography technique terms
- **Length Optimization**: Automatic adjustment within API limits
- **Template Application**: Pre-defined style combinations

### 🛡️ Reliability and Stability
- **Auto Retry**: Automatic RESOURCE_EXHAUSTED handling
- **Error Classification**: Distinguish between Safety, Recitation, and Technical errors
- **Exponential Backoff**: Intelligent retry wait time calculation
- **Timeout Management**: Network reliability assurance

---

## 📦 Installation

### Prerequisites
- Python 3.9 or higher
- Gemini API key (https://ai.google.dev)

### ✨ No External Dependencies!
This Skill uses **only Python standard library**.
No additional library installation required! 🎉

**Standard libraries used:**
- `urllib` - HTTP requests and API calls
- `json` - JSON processing
- `base64` - Image encoding
- `pathlib` - File path handling
- `typing` - Type hinting
- `time` - Retry delays

### API Key Setup
```bash
# 1. Create .env file
echo "GEMINI_API_KEY=gsk_YOUR_API_KEY" > .env

# 2. Or set up in Python
! uv run -c "
from modules.env_key_manager import EnvKeyManager
EnvKeyManager.set_api_key('gsk_YOUR_API_KEY')
"
```

---

## 🚀 Usage Guide

### 1️⃣ Basic Image Generation

**Running Python script:**
```bash
! uv run -c "
from modules.env_key_manager import EnvKeyManager
from modules.image_generator import ImageGenerator
from modules.prompt_generator import PromptGenerator

# API 키 로드
api_key = EnvKeyManager.get_api_key()

# 프롬프트 최적화
prompt = PromptGenerator.optimize(
    'beautiful mountain landscape at sunset',
    style='photorealistic'
)

# 이미지 생성
generator = ImageGenerator(api_key)
result = generator.generate_image(
    prompt=prompt,
    resolution='2048x2048'
)

print(f'Success: {result[\"success\"]}')
print(f'Image data length: {len(result.get(\"image_data\", \"\"))}')
"
```

**Save example:**
```bash
! uv run -c "
from modules.env_key_manager import EnvKeyManager
from modules.image_generator import ImageGenerator

api_key = EnvKeyManager.get_api_key()
generator = ImageGenerator(api_key)

result = generator.generate_image('serene ocean view', resolution='2048x2048')
if result['success']:
    generator.save_image(result, 'output/ocean_sunset.png')
    print('Image saved to output/ocean_sunset.png')
"
```

### 2️⃣ Multi-language Prompt Support

**Korean example:**
```bash
! uv run -c "
from modules.prompt_generator import PromptGenerator

# Auto language detection and optimization
prompt_ko = PromptGenerator.optimize('한국 고궁의 아름다운 건축물', style='photorealistic')
print(f'Korean: {prompt_ko}')

# Explicit language specification
prompt_ko = PromptGenerator.optimize(
    'cherry blossoms in spring',
    language='ja',
    style='artistic'
)
print(f'Japanese: {prompt_ko}')
"
```

### 3️⃣ Image Editing

**URL-based image editing:**
```bash
! uv run -c "
from modules.env_key_manager import EnvKeyManager
from modules.image_generator import ImageGenerator

api_key = EnvKeyManager.get_api_key()
generator = ImageGenerator(api_key)

# URL에서 이미지 로드 및 편집
result = generator.edit_image(
    image_input='https://example.com/image.jpg',
    instruction='change the sky to vibrant sunset colors',
    resolution='2048x2048'
)

if result['success']:
    print('Image edited successfully')
"
```

**Local file editing:**
```bash
! uv run -c "
from modules.image_generator import ImageGenerator
from modules.env_key_manager import EnvKeyManager

api_key = EnvKeyManager.get_api_key()
generator = ImageGenerator(api_key)

# Edit local file
result = generator.edit_image(
    image_input='input/original.png',
    instruction='apply warm lighting and enhance details',
    resolution='2048x2048'
)
"
```

### 4️⃣ Error Handling

**Auto retry and recovery:**
```bash
! uv run -c "
from modules.env_key_manager import EnvKeyManager
from modules.image_generator import ImageGenerator
from modules.error_handler import ErrorHandler
import time

api_key = EnvKeyManager.get_api_key()
generator = ImageGenerator(api_key)

max_retries = 3
attempt = 0

while attempt < max_retries:
    try:
        result = generator.generate_image('beautiful landscape')
        if result['success']:
            print('Image generated successfully')
            break
    except Exception as e:
        print(f'Error: {e}')
        attempt += 1
        if attempt < max_retries:
            print(f'Retrying in 5 seconds...')
            time.sleep(5)
"
```

---

## 📖 API Reference

### EnvKeyManager

**Methods:**
- `get_api_key() -> str | None`: Load API key
- `set_api_key(api_key: str) -> bool`: Set API key
- `validate_api_key(api_key: str) -> bool`: Validate API key format
- `is_configured() -> bool`: Check configuration status

**Usage example:**
```python
from modules.env_key_manager import EnvKeyManager

if not EnvKeyManager.is_configured():
    EnvKeyManager.set_api_key("gsk_...")

api_key = EnvKeyManager.get_api_key()
```

### PromptGenerator

**Methods:**
- `optimize(prompt, style, add_photographic, language) -> str`: Optimize prompt
- `validate(prompt) -> bool`: Validate prompt format
- `add_style(prompt, style) -> str`: Add style to prompt
- `get_style_list() -> list`: Get available styles
- `get_resolution_list() -> dict`: Get available resolutions

**Supported styles:**
- `photorealistic`: Photorealistic and realistic images
- `artistic`: Artistic style
- `cinematic`: Cinematic and dramatic cinematography
- `minimal`: Minimalist and clean composition
- `abstract`: Abstract art style
- `fantasy`: Fantasy and magical atmosphere

### ImageGenerator

**Methods:**
- `generate_image(prompt, resolution, max_retries) -> dict`: Generate image
- `edit_image(image_input, instruction, resolution, max_retries) -> dict`: Edit image
- `save_image(image_data, output_path) -> bool`: Save image to file

**Response structure:**
```python
{
    "success": bool,
    "image_data": "base64_encoded_string",
    "mime_type": "image/jpeg",
    "finish_reason": "STOP",
    "metadata": {
        "input_tokens": int,
        "output_tokens": int,
        "total_tokens": int,
        "synthetic_watermark": True
    }
}
```

### ErrorHandler

**Methods:**
- `is_retryable() -> bool`: Check if error is retryable
- `get_retry_delay() -> float`: Get retry wait time
- `get_message() -> str`: Get user message
- `get_resolution_action() -> str`: Get resolution action
- `get_error_details() -> dict`: Get error details

**Error types:**
| Code | Retryable | Description |
|------|-----------|-------------|
| RESOURCE_EXHAUSTED | ✅ | API rate limit (429) |
| SAFETY | ❌ | Safety policy violation |
| RECITATION | ❌ | Training data similarity |
| INVALID_ARGUMENT | ❌ | Invalid input |
| INTERNAL | ✅ | Server error |

---

## 🔧 Advanced Usage

### Custom Style Combination

```bash
! uv run -c "
from modules.prompt_generator import PromptGenerator

# Combine multiple styles
base_prompt = 'ancient temple in mountains'
optimized = PromptGenerator.optimize(base_prompt, style='photorealistic')
optimized = PromptGenerator.add_style(optimized, 'cinematic')

print(f'Combined styles: {optimized}')
"
```

### Batch Image Generation

```bash
! uv run -c "
from modules.env_key_manager import EnvKeyManager
from modules.image_generator import ImageGenerator
from modules.prompt_generator import PromptGenerator

api_key = EnvKeyManager.get_api_key()
generator = ImageGenerator(api_key)

prompts = [
    'serene mountain landscape',
    'bustling city street at night',
    'peaceful forest clearing'
]

for i, prompt in enumerate(prompts, 1):
    optimized = PromptGenerator.optimize(prompt)
    result = generator.generate_image(optimized)

    if result['success']:
        generator.save_image(result, f'output/image_{i}.png')
        print(f'Generated: image_{i}.png')
    else:
        print(f'Failed: {prompt}')
"
```

### Generation by Resolution

```bash
! uv run -c "
from modules.env_key_manager import EnvKeyManager
from modules.image_generator import ImageGenerator
from modules.prompt_generator import PromptGenerator

api_key = EnvKeyManager.get_api_key()
generator = ImageGenerator(api_key)
prompt = 'beautiful landscape'

for resolution_name, resolution in [('1K', '1024x1024'), ('2K', '2048x2048')]:
    result = generator.generate_image(prompt, resolution=resolution)
    if result['success']:
        generator.save_image(result, f'output/{resolution_name}.png')
"
```

---

## ⚠️ Troubleshooting

### API Key Error
```
Error: UNAUTHENTICATED
Solution: Check GEMINI_API_KEY in .env file and validate it
```

### Safety Violation
```
Error: SAFETY finish_reason
Solution: Remove prohibited content from prompt and retry
```

### Rate Limit
```
Error: RESOURCE_EXHAUSTED (429)
Solution: Use automatic retry or wait 5 minutes before retrying
```

### Insufficient Memory
```
Error: RESOURCE_EXHAUSTED (memory)
Solution: Reduce image resolution and retry
```

---

## 📊 Performance Optimization

### Recommended Settings

**Fast Generation (Development):**
```python
ImageGenerator(api_key).generate_image(
    prompt="landscape",
    resolution="1024x1024",  # Low resolution
    max_retries=1  # Minimal retries
)
```

**High Quality Generation (Production):**
```python
ImageGenerator(api_key).generate_image(
    prompt="landscape",
    resolution="4096x4096",  # High resolution
    max_retries=3  # Sufficient retries
)
```

### Cost Reduction

- Development stage: Use 1K resolution (1/16 cost)
- Testing: Generate multiple images at once in batch mode
- Retry: Automatic optimization with exponential backoff

---

## 🧪 Testing

### Run Tests
```bash
! uv run -m pytest tests/ -v
```

### Check Coverage
```bash
! uv run -m pytest tests/ --cov=modules --cov-report=html
```

### Lint Check
```bash
! uv run -m ruff check modules/
! uv run -m mypy modules/
```

---

## 📝 Logging and Debugging

### Enable Debug Mode
```python
import logging

logging.basicConfig(level=logging.DEBUG)

# All requests/responses will be logged
```

### Detailed Error Information
```python
from modules.error_handler import ErrorHandler

try:
    result = generator.generate_image(prompt)
except Exception as e:
    error_handler = ErrorHandler({"error": str(e)})
    details = error_handler.get_error_details()
    print(f"Full error: {details}")
```

---

## 🔗 References

- **Gemini API Official Documentation**: https://ai.google.dev/api/rest/v1beta
- **Nano Banana Pro Blog**: https://blog.google/intl/ko-kr/company-news/technology/nano-banana-pro/
- **Google AI Studio**: https://aistudio.google.com
- **Rate Limiting Guide**: https://ai.google.dev/models/gemini-3-pro-image#rate-limits

---

## 📄 License

MIT License - Free to use, modify, and distribute

---

## 🤝 Contributing

1. Bug Reports: Include detailed reproduction steps
2. Feature Requests: Describe use cases
3. Code Contributions: Follow PEP 8 style guide

---

**Last Updated**: 2025-11-22
**Maintained by**: MoAI-ADK Team
