# Gemini Image Generation API - Complete Reference

**Official Documentation**: https://ai.google.dev/gemini-api/docs/image-generation

**Last Updated**: 2025-11-27
**API Version**: v1beta
**SDK**: google-genai >= 1.0.0

---

## Table of Contents

1. [Models & Endpoints](#models--endpoints)
2. [Configuration Parameters](#configuration-parameters)
3. [API Methods](#api-methods)
4. [Response Structure](#response-structure)
5. [Error Handling](#error-handling)
6. [Rate Limits & Quotas](#rate-limits--quotas)
7. [Best Practices](#best-practices)

---

## Models & Endpoints

### Available Models

| Model ID | Performance | Resolution | Use Case | Processing Time |
|----------|-------------|------------|----------|-----------------|
| `gemini-2.5-flash-image` | Fast | Up to 1K | Quick iterations, prototyping | ~5-15s |
| `gemini-3-pro-image-preview` | Professional | Up to 4K | Production assets, advanced features | ~10-60s |

### API Endpoint

```
Base URL: https://generativelanguage.googleapis.com/v1beta/models/{model}:generateContent
```

### Authentication

**Python**:
```python
from google import genai

# Method 1: Direct API key
client = genai.Client(api_key="YOUR_API_KEY")

# Method 2: Environment variable
import os
api_key = os.getenv("GEMINI_API_KEY")
client = genai.Client(api_key=api_key)
```

**JavaScript**:
```javascript
import { GoogleGenAI } from "@google/genai";

const ai = new GoogleGenAI({
  apiKey: process.env.GEMINI_API_KEY
});
```

**REST/cURL**:
```bash
export GEMINI_API_KEY="your-api-key-here"

curl -X POST \
  "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-image:generateContent" \
  -H "x-goog-api-key: $GEMINI_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{...}'
```

---

## Configuration Parameters

### Image Configuration (`image_config`)

**`aspect_ratio`** (string)
- **Description**: Output image dimension ratio
- **Valid values**: "1:1", "2:3", "3:2", "3:4", "4:3", "4:5", "5:4", "9:16", "16:9", "21:9", "9:21"
- **Default**: Model-dependent (usually "1:1")
- **Example**:
  ```python
  image_config=types.ImageConfig(aspect_ratio="16:9")
  ```

**`image_size`** (string, Gemini 3 Pro only)
- **Description**: Output resolution
- **Valid values**: "1K", "2K", "4K"
- **Default**: "1K"
- **Note**: Must use uppercase 'K'. Lowercase or numeric values will be rejected.
- **Example**:
  ```python
  image_config=types.ImageConfig(
      aspect_ratio="16:9",
      image_size="4K"
  )
  ```

### Response Modalities

**`response_modalities`** (array)
- **Description**: Types of content to generate
- **Valid values**: ["TEXT"], ["IMAGE"], ["TEXT", "IMAGE"]
- **Default**: ["TEXT", "IMAGE"]
- **Example**:
  ```python
  config=types.GenerateContentConfig(
      response_modalities=["TEXT", "IMAGE"]
  )
  ```

### Tools Configuration

**`tools`** (array)
- **Description**: Optional tools for enhanced generation
- **Available tools**:
  - **Google Search**: `[{"google_search": {}}]` - Real-time information grounding
- **Example**:
  ```python
  config=types.GenerateContentConfig(
      response_modalities=["TEXT", "IMAGE"],
      tools=[{"google_search": {}}]
  )
  ```

---

## API Methods

### 1. Text-to-Image Generation

**Generate image from text prompt**

**Python**:
```python
response = client.models.generate_content(
    model="gemini-2.5-flash-image",
    contents="A serene mountain landscape at golden hour",
    config=types.GenerateContentConfig(
        response_modalities=["TEXT", "IMAGE"],
        image_config=types.ImageConfig(
            aspect_ratio="16:9"
        )
    )
)

for part in response.parts:
    if part.text:
        print(part.text)
    elif part.inline_data:
        image = part.as_image()
        image.save("output.png")
```

**JavaScript**:
```javascript
const response = await ai.models.generateContent({
  model: "gemini-2.5-flash-image",
  contents: "A serene mountain landscape at golden hour",
  config: {
    responseModalities: ["TEXT", "IMAGE"],
    imageConfig: {
      aspectRatio: "16:9"
    }
  }
});
```

**REST/cURL**:
```bash
curl -s -X POST \
  "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-image:generateContent" \
  -H "x-goog-api-key: $GEMINI_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "contents": [{
      "parts": [
        {"text": "A serene mountain landscape at golden hour"}
      ]
    }],
    "generationConfig": {
      "responseModalities": ["TEXT", "IMAGE"]
    }
  }' | jq -r '.candidates[0].content.parts[] | select(.inlineData) | .inlineData.data' | base64 -d > output.png
```

### 2. Image-to-Image Editing

**Edit existing image with instructions**

**Python**:
```python
from PIL import Image

original_image = Image.open("photo.png")

response = client.models.generate_content(
    model="gemini-2.5-flash-image",
    contents=[
        "Add a sunset in the background",
        original_image
    ],
    config=types.GenerateContentConfig(
        response_modalities=["TEXT", "IMAGE"]
    )
)

for part in response.parts:
    if part.inline_data:
        edited = part.as_image()
        edited.save("edited.png")
```

**JavaScript**:
```javascript
const imagePath = "photo.png";
const imageData = fs.readFileSync(imagePath);
const base64Image = imageData.toString("base64");

const response = await ai.models.generateContent({
  model: "gemini-2.5-flash-image",
  contents: [
    { text: "Add a sunset in the background" },
    {
      inlineData: {
        mimeType: "image/png",
        data: base64Image
      }
    }
  ]
});
```

**REST/cURL**:
```bash
IMG_PATH="photo.jpeg"
IMG_BASE64=$(base64 -w0 "$IMG_PATH")

curl -X POST \
  "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-image:generateContent" \
  -H "x-goog-api-key: $GEMINI_API_KEY" \
  -H 'Content-Type: application/json' \
  -d "{
    \"contents\": [{
      \"parts\":[
        {\"text\": \"Add a sunset in the background\"},
        {
          \"inline_data\": {
            \"mime_type\":\"image/jpeg\",
            \"data\": \"$IMG_BASE64\"
          }
        }
      ]
    }]
  }"
```

### 3. Multi-turn Chat

**Iterative refinement through conversation**

**Python**:
```python
chat = client.chats.create(
    model="gemini-3-pro-image-preview",
    config=types.GenerateContentConfig(
        response_modalities=["TEXT", "IMAGE"]
    )
)

# Turn 1
response1 = chat.send_message("Create a vibrant infographic about photosynthesis")

# Turn 2
response2 = chat.send_message("Make it more colorful")

# Turn 3
response3 = chat.send_message("Translate text to Spanish")
```

**JavaScript**:
```javascript
const chat = ai.chats.create({
  model: "gemini-3-pro-image-preview",
  config: {
    responseModalities: ["TEXT", "IMAGE"]
  }
});

const response1 = await chat.sendMessage({
  message: "Create a vibrant infographic about photosynthesis"
});

const response2 = await chat.sendMessage({
  message: "Make it more colorful"
});
```

### 4. Multi-Image References (Gemini 3 Pro only)

**Generate with multiple reference images**

**Python**:
```python
from PIL import Image

response = client.models.generate_content(
    model="gemini-3-pro-image-preview",
    contents=[
        "An office group photo of these people making funny faces",
        Image.open("person1.png"),
        Image.open("person2.png"),
        Image.open("person3.png"),
        Image.open("person4.png"),
        Image.open("person5.png")
    ],
    config=types.GenerateContentConfig(
        response_modalities=["TEXT", "IMAGE"],
        image_config=types.ImageConfig(
            aspect_ratio="5:4",
            image_size="2K"
        )
    )
)
```

**Limitations**:
- Maximum 14 reference images
- Up to 6 high-fidelity object images
- Up to 5 person images for character consistency

### 5. Google Search Grounding

**Generate with real-time information**

**Python**:
```python
response = client.models.generate_content(
    model="gemini-3-pro-image-preview",
    contents="Visualize current weather forecast for next 5 days in San Francisco",
    config=types.GenerateContentConfig(
        response_modalities=["TEXT", "IMAGE"],
        tools=[{"google_search": {}}],
        image_config=types.ImageConfig(
            aspect_ratio="16:9"
        )
    )
)
```

**Use Cases**:
- Weather forecasts
- Stock market information
- Recent events
- Current statistics

**Note**: Image-based search results are not passed to generation model.

---

## Response Structure

### Basic Response Format

```python
{
  "candidates": [{
    "content": {
      "parts": [
        {
          "text": "Description or explanation"
        },
        {
          "inlineData": {
            "mimeType": "image/png",
            "data": "base64_encoded_image_data"
          }
        }
      ]
    }
  }],
  "usageMetadata": {
    "totalTokenCount": 1234
  }
}
```

### Accessing Response Parts

**Python**:
```python
for part in response.parts:
    if part.text:
        print(f"Description: {part.text}")
    elif part.inline_data:
        image = part.as_image()  # Returns PIL.Image
        image.save("output.png")
```

**JavaScript**:
```javascript
for (const part of response.candidates[0].content.parts) {
  if (part.text) {
    console.log(part.text);
  } else if (part.inlineData) {
    const buffer = Buffer.from(part.inlineData.data, "base64");
    fs.writeFileSync("output.png", buffer);
  }
}
```

### Grounding Metadata

When using Google Search tool:

```json
{
  "groundingMetadata": {
    "searchEntryPoint": {
      "webSearchEntries": [...]
    },
    "groundingChunks": [
      {
        "web": {
          "title": "Source title",
          "uri": "https://example.com",
          "snippet": "Relevant excerpt"
        }
      }
    ]
  }
}
```

**Contains**:
- `searchEntryPoint`: Recommended searches (HTML/CSS rendering)
- `groundingChunks`: Top 3 web sources used

### Thought Process (Gemini 3 Pro)

**Thinking mode** (automatic, cannot be disabled):

```python
for part in response.parts:
    if part.thought:
        if part.text:
            print(f"Thinking: {part.text}")
        elif part.as_image():
            image = part.as_image()
            image.show()  # Intermediate reasoning image
```

**Thought Signatures**:
- Ensures consistency across multi-turn interactions
- Automatically managed by official SDKs
- No manual extraction needed when using chat functionality

---

## Error Handling

### Common Errors

**1. Resource Exhausted (Quota Exceeded)**

```python
from google.api_core import exceptions

try:
    response = client.models.generate_content(...)
except exceptions.ResourceExhausted:
    print("API quota exceeded. Please try again later.")
    # Implement exponential backoff
```

**2. Permission Denied**

```python
try:
    response = client.models.generate_content(...)
except exceptions.PermissionDenied:
    print("Invalid API key or insufficient permissions")
```

**3. Invalid Argument**

```python
try:
    response = client.models.generate_content(...)
except exceptions.InvalidArgument as e:
    print(f"Invalid parameter: {e}")
```

### Retry Logic with Exponential Backoff

```python
import time

def generate_with_retry(client, prompt, max_retries=3):
    for attempt in range(max_retries):
        try:
            return client.models.generate_content(
                model="gemini-2.5-flash-image",
                contents=prompt
            )
        except exceptions.ResourceExhausted:
            if attempt == max_retries - 1:
                raise
            wait_time = 2 ** attempt
            time.sleep(wait_time)
```

---

## Rate Limits & Quotas

### Rate Limits

- **Gemini 2.5 Flash**: 60 requests per minute (RPM)
- **Gemini 3 Pro**: 15 requests per minute (RPM)
- **Concurrent requests**: 5 maximum

### Quota Management

**Check quota status**:
```bash
curl -X GET \
  "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-image" \
  -H "x-goog-api-key: $GEMINI_API_KEY"
```

**Best Practices**:
1. Implement exponential backoff
2. Queue requests during high-volume operations
3. Use Batch API for bulk processing (24-hour window)

### Batch API

For high-volume generation:

```python
# Batch API allows up to 24-hour processing windows
# Higher rate limits in exchange for delayed results

# See: https://ai.google.dev/gemini-api/docs/batch-api
```

---

## Best Practices

### 1. Prompt Engineering

**Core Principle**: Describe scenes, not keyword lists

**Good Example** (descriptive paragraph):
```
A photorealistic close-up portrait of an elderly Japanese ceramicist
with sun-etched wrinkles and knowing smile inspecting a glazed tea bowl.
Rustic workshop with pottery wheels. Golden hour light through window.
85mm portrait lens with soft bokeh. Serene, masterful mood.
```

**Bad Example** (keyword list):
```
ceramicist, old, japan, tea, workshop, golden hour
```

### 2. Resolution Selection

| Resolution | Use Case | Processing Time | File Size |
|------------|----------|-----------------|-----------|
| 1K | Quick iterations, drafts | Fast (~10-20s) | ~500KB |
| 2K | Professional assets | Medium (~20-40s) | ~2MB |
| 4K | Print quality, hero images | Slow (~40-60s) | ~8MB |

### 3. Model Selection

**Choose Gemini 2.5 Flash when**:
- Quick iterations needed
- Prototyping and testing
- High volume generation
- Cost sensitive

**Choose Gemini 3 Pro when**:
- Professional quality required
- 2K/4K resolution needed
- Complex compositions
- Advanced text in images
- Multi-image references

### 4. Cost Optimization

**Strategies**:
1. Use Flash for iterations, Pro for final output
2. Cache prompts for similar generations
3. Batch process with rate limiting
4. Implement request queuing

---

## Security & Compliance

### User Responsibilities

**You must**:
- Own rights to uploaded images
- Comply with Google's "Generative AI Use Policy"
- Avoid prohibited content (deceptive, harassing, harmful)

**Prohibited**:
- Copyright infringement
- Deceptive deepfakes
- Harassment or harmful content
- Violating others' rights

### SynthID Watermarking

**All generated images** automatically include SynthID watermarks:
- Invisible to human eye
- Verifies AI-generated authenticity
- Cannot be removed via API

---

## Advanced Features

### 1. Thinking Mode (Gemini 3 Pro)

- Automatic (cannot be disabled)
- Generates up to 2 temporary reasoning images
- Final image includes last thinking image
- Reveals composition refinement logic

### 2. Thought Signatures

- Ensures multi-turn context continuity
- Automatically managed by official SDKs
- No manual handling required

### 3. Sequential Art (Comics/Storyboards)

**Example**:
```python
response = client.models.generate_content(
    model="gemini-3-pro-image-preview",
    contents="Make a 3-panel comic in gritty noir art style. "
             "Character in humorous scene.",
    config=types.GenerateContentConfig(
        response_modalities=["TEXT", "IMAGE"]
    )
)
```

Best used with Gemini 3 Pro for text accuracy and storytelling.

---

## Available Regions

Check supported regions:
- [Available Regions Documentation](https://ai.google.dev/gemini-api/docs/available-regions)

---

## Pricing

Check current pricing:
- [Pricing Page](https://ai.google.dev/pricing)

**General Structure**:
- Charged per image generation
- Higher resolution = higher cost
- Gemini 3 Pro more expensive than 2.5 Flash

---

## Additional Resources

- **Official Documentation**: https://ai.google.dev/gemini-api/docs/image-generation
- **SDK Libraries**: https://ai.google.dev/gemini-api/docs/libraries
- **Thinking Mode Guide**: https://ai.google.dev/gemini-api/docs/thinking
- **Google Search Grounding**: https://ai.google.dev/gemini-api/docs/grounding
- **Batch API**: https://ai.google.dev/gemini-api/docs/batch-api
- **Rate Limits**: https://ai.google.dev/gemini-api/docs/rate-limits

---

**Document Version**: 1.0.0
**Last Updated**: 2025-11-27
**Status**: Production Ready
