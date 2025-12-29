#!/usr/bin/env python3
"""
Nano Banana Pro - Single Image Generation Script

Generates images using the gemini-3-pro-image-preview model (Nano Banana Pro).
This script is the canonical implementation for single image generation tasks.

Features:
    - 1K, 2K, 4K resolution support
    - 10 aspect ratios (1:1, 16:9, etc.)
    - Style prefix support for consistent aesthetics
    - Exponential backoff retry for API quota errors
    - Google Search grounding integration

Usage:
    python generate_image.py -p "A futuristic cityscape" -o city.png
    python generate_image.py -p "Dashboard UI" -o ui.png -r 4K -a 16:9
    python generate_image.py -p "Logo design" -o logo.png --style "minimalist, vector"

Environment Variables Required:
    GOOGLE_API_KEY - Your Google AI Studio API key

Model: gemini-3-pro-image-preview (hardcoded)
"""

import argparse
import os
import random
import sys
import time
from datetime import datetime
from pathlib import Path
from typing import Optional

try:
    from google import genai
    from google.genai import types
except ImportError:
    print("Error: google-genai package not installed.")
    print("Install with: pip install google-genai")
    sys.exit(1)

try:
    from dotenv import load_dotenv

    load_dotenv()
except ImportError:
    pass  # dotenv is optional


# Fixed model - Nano Banana Pro
MODEL_NAME = "gemini-3-pro-image-preview"

# Supported configurations
SUPPORTED_ASPECT_RATIOS = ["1:1", "2:3", "3:2", "3:4", "4:3", "4:5", "5:4", "9:16", "16:9", "21:9"]
SUPPORTED_RESOLUTIONS = ["1K", "2K", "4K"]  # Must be uppercase

# Retry configuration
MAX_RETRIES = 5
BASE_DELAY = 2.0
MAX_DELAY = 60.0


def get_api_key() -> str:
    """Get API key from environment variable."""
    api_key = os.getenv("GOOGLE_API_KEY")
    if not api_key:
        print("Error: GOOGLE_API_KEY environment variable not set.")
        print("Set up your API key:")
        print("  1. Get your key from https://aistudio.google.com/apikey")
        print("  2. Create .env file with: GOOGLE_API_KEY=your_key_here")
        print("  3. Or set environment variable: export GOOGLE_API_KEY=your_key_here")
        sys.exit(1)
    return api_key


def validate_aspect_ratio(ratio: str) -> str:
    """Validate and normalize aspect ratio."""
    if ratio not in SUPPORTED_ASPECT_RATIOS:
        print(f"Warning: '{ratio}' is not a standard aspect ratio.")
        print(f"Supported ratios: {', '.join(SUPPORTED_ASPECT_RATIOS)}")
        print("Using default: 16:9")
        return "16:9"
    return ratio


def validate_resolution(resolution: str) -> str:
    """Validate and normalize resolution (must be uppercase K)."""
    normalized = resolution.upper()
    if normalized not in SUPPORTED_RESOLUTIONS:
        print(f"Warning: '{resolution}' is not a valid resolution.")
        print(f"Supported resolutions: {', '.join(SUPPORTED_RESOLUTIONS)}")
        print("Using default: 2K")
        return "2K"
    return normalized


def build_prompt_with_style(prompt: str, style: Optional[str] = None) -> str:
    """Build final prompt with optional style prefix."""
    if style:
        return f"{style}: {prompt}"
    return prompt


def calculate_backoff_delay(attempt: int, jitter: bool = True) -> float:
    """Calculate exponential backoff delay with optional jitter."""
    delay = min(BASE_DELAY * (2**attempt), MAX_DELAY)
    if jitter:
        delay = delay * (0.5 + random.random())
    return delay


def generate_image(
    prompt: str,
    output_path: str,
    aspect_ratio: str = "16:9",
    resolution: str = "2K",
    style: Optional[str] = None,
    enable_grounding: bool = False,
    max_retries: int = MAX_RETRIES,
    verbose: bool = False,
) -> dict:
    """
    Generate an image using Nano Banana Pro (gemini-3-pro-image-preview).

    Args:
        prompt: The image generation prompt
        output_path: Path to save the generated image
        aspect_ratio: Image aspect ratio (default: 16:9)
        resolution: Image resolution - 1K, 2K, or 4K (default: 2K)
        style: Optional style prefix (e.g., "photorealistic", "minimalist")
        enable_grounding: Enable Google Search grounding (default: False)
        max_retries: Maximum retry attempts for transient errors (default: 5)
        verbose: Print detailed progress information

    Returns:
        dict with generation metadata including success status, file path, and timing
    """
    # Validate inputs
    aspect_ratio = validate_aspect_ratio(aspect_ratio)
    resolution = validate_resolution(resolution)
    final_prompt = build_prompt_with_style(prompt, style)

    # Get API key
    api_key = get_api_key()

    if verbose:
        print(f"  Model: {MODEL_NAME}")
        print(f"  Final Prompt: {final_prompt[:100]}{'...' if len(final_prompt) > 100 else ''}")
        print(f"  Style: {style or 'None'}")
        print("-" * 50)

    # Initialize client
    client = genai.Client(api_key=api_key)

    # Build configuration
    config_params = {
        "response_modalities": ["TEXT", "IMAGE"],
        "image_config": types.ImageConfig(aspect_ratio=aspect_ratio, image_size=resolution),
    }

    # Optionally enable Google Search grounding
    if enable_grounding:
        config_params["tools"] = [{"google_search": {}}]

    config = types.GenerateContentConfig(**config_params)

    # Retry logic with exponential backoff
    last_error = None
    for attempt in range(max_retries):
        try:
            if verbose:
                print(f"  Attempt {attempt + 1}/{max_retries}...")

            start_time = time.time()

            # Generate image
            response = client.models.generate_content(model=MODEL_NAME, contents=final_prompt, config=config)

            generation_time = time.time() - start_time

            # Ensure output directory exists
            output_dir = Path(output_path).parent
            if output_dir and str(output_dir) != "." and not output_dir.exists():
                output_dir.mkdir(parents=True, exist_ok=True)

            # Save image
            image_saved = False
            text_response = ""

            for part in response.candidates[0].content.parts:
                if part.inline_data is not None:
                    with open(output_path, "wb") as f:
                        f.write(part.inline_data.data)
                    image_saved = True
                elif hasattr(part, "text") and part.text:
                    text_response = part.text

            if not image_saved:
                raise RuntimeError("No image data in response. The model may have returned text-only response.")

            # Return success metadata
            return {
                "success": True,
                "model": MODEL_NAME,
                "prompt": prompt,
                "final_prompt": final_prompt,
                "style": style,
                "output_path": str(Path(output_path).absolute()),
                "aspect_ratio": aspect_ratio,
                "resolution": resolution,
                "generation_time_seconds": round(generation_time, 2),
                "timestamp": datetime.now().isoformat(),
                "text_response": text_response if text_response else None,
                "grounding_enabled": enable_grounding,
                "attempts": attempt + 1,
            }

        except Exception as e:
            last_error = e
            error_str = str(e).lower()

            # Check for non-retryable errors
            if "permission" in error_str or "api_key" in error_str or "invalid_api_key" in error_str:
                print(f"Error: API key issue - {e}")
                print("Check your GOOGLE_API_KEY in .env file or environment variable.")
                break

            if "invalid" in error_str and "argument" in error_str:
                print(f"Error: Invalid argument - {e}")
                print(f"Check aspect_ratio ({aspect_ratio}) and resolution ({resolution}).")
                break

            # Check for quota/rate limit errors - apply exponential backoff
            is_quota_error = any(kw in error_str for kw in ["quota", "rate", "429", "resource_exhausted", "too_many"])

            if attempt < max_retries - 1:
                delay = calculate_backoff_delay(attempt, jitter=is_quota_error)
                if is_quota_error:
                    print(f"  Rate limit hit (attempt {attempt + 1}). Waiting {delay:.1f}s...")
                else:
                    print(f"  Attempt {attempt + 1} failed: {e}")
                    print(f"  Retrying in {delay:.1f}s...")
                time.sleep(delay)
            else:
                print(f"  All {max_retries} attempts failed.")

    # Return failure metadata
    return {
        "success": False,
        "model": MODEL_NAME,
        "prompt": prompt,
        "final_prompt": final_prompt,
        "style": style,
        "error": str(last_error),
        "timestamp": datetime.now().isoformat(),
        "attempts": max_retries,
    }


def main():
    """Main entry point for CLI usage."""
    parser = argparse.ArgumentParser(
        description="Generate images using Nano Banana Pro (gemini-3-pro-image-preview)",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
    # Basic usage
    python generate_image.py -p "A fluffy cat eating a banana" -o cat.png

    # High resolution with specific aspect ratio
    python generate_image.py -p "Modern dashboard UI" -o dashboard.png -a 16:9 -r 4K

    # With style prefix
    python generate_image.py -p "Mountain landscape" -o landscape.png \\
        --style "photorealistic, dramatic lighting"

    # With Google Search grounding for factual content
    python generate_image.py -p "Mount Fuji at sunset" -o fuji.png -g

    # Verbose mode with all options
    python generate_image.py -p "Tech company logo" -o logo.png \\
        -r 4K -a 1:1 --style "minimalist, vector art" -v

Style Prefix Examples:
    - "photorealistic, 8K, detailed"
    - "minimalist, vector art, clean"
    - "watercolor painting, soft colors"
    - "3D render, cinematic lighting"
    - "anime style, vibrant colors"

Environment:
    GOOGLE_API_KEY - Required. Get from https://aistudio.google.com/apikey
        """,
    )

    parser.add_argument("-p", "--prompt", required=True, help="Image generation prompt (descriptive text)")

    parser.add_argument("-o", "--output", required=True, help="Output file path for the generated image (PNG format)")

    parser.add_argument(
        "-a", "--aspect-ratio", default="16:9", choices=SUPPORTED_ASPECT_RATIOS, help="Aspect ratio (default: 16:9)"
    )

    parser.add_argument(
        "-r",
        "--resolution",
        default="2K",
        choices=SUPPORTED_RESOLUTIONS,
        help="Resolution: 1K, 2K, or 4K (default: 2K)",
    )

    parser.add_argument("--style", help="Style prefix to prepend to prompt (e.g., 'photorealistic', 'minimalist')")

    parser.add_argument(
        "-g", "--enable-grounding", action="store_true", help="Enable Google Search grounding for factual content"
    )

    parser.add_argument(
        "--max-retries",
        type=int,
        default=MAX_RETRIES,
        help=f"Maximum retry attempts for transient errors (default: {MAX_RETRIES})",
    )

    parser.add_argument("-v", "--verbose", action="store_true", help="Print detailed progress information")

    args = parser.parse_args()

    print(f"Generating image with {MODEL_NAME}...")
    print(f"  Prompt: {args.prompt[:80]}{'...' if len(args.prompt) > 80 else ''}")
    if args.style:
        print(f"  Style: {args.style}")
    print(f"  Aspect Ratio: {args.aspect_ratio}")
    print(f"  Resolution: {args.resolution}")
    print(f"  Output: {args.output}")

    result = generate_image(
        prompt=args.prompt,
        output_path=args.output,
        aspect_ratio=args.aspect_ratio,
        resolution=args.resolution,
        style=args.style,
        enable_grounding=args.enable_grounding,
        max_retries=args.max_retries,
        verbose=args.verbose,
    )

    if result["success"]:
        print("\nSuccess!")
        print(f"  File: {result['output_path']}")
        print(f"  Time: {result['generation_time_seconds']}s")
        print(f"  Attempts: {result['attempts']}")
        if result.get("text_response"):
            print(f"  Model Note: {result['text_response'][:200]}")
    else:
        print(f"\nFailed: {result['error']}")
        sys.exit(1)


if __name__ == "__main__":
    main()
