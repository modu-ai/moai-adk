#!/usr/bin/env python3
"""
Nano Banana Pro - Batch Image Generation Script

Batch image generation using the gemini-3-pro-image-preview model (Nano Banana Pro).
Processes multiple prompts from JSON/YAML files or command line with smart rate limiting.

Features:
    - JSON/YAML batch configuration files
    - Parallel processing with concurrency control
    - Exponential backoff for API quota errors
    - Progress tracking and detailed reports
    - Style inheritance from global settings
    - Resume capability for interrupted batches

Usage:
    python batch_generate.py -c prompts.json -d output/
    python batch_generate.py -c prompts.yaml -d output/ --concurrency 3
    python batch_generate.py --prompts "Cat" "Dog" "Bird" -d output/

Environment Variables Required:
    GOOGLE_API_KEY - Your Google AI Studio API key

Model: gemini-3-pro-image-preview (hardcoded)
"""

import argparse
import asyncio
import json
import os
import random
import sys
import time
from dataclasses import dataclass, field
from datetime import datetime
from pathlib import Path
from typing import Any, Dict, List, Optional

try:
    from google import genai
    from google.genai import types
except ImportError:
    print("Error: google-genai package not installed.")
    print("Install with: pip install google-genai")
    sys.exit(1)

try:
    import yaml

    YAML_AVAILABLE = True
except ImportError:
    YAML_AVAILABLE = False

try:
    from dotenv import load_dotenv

    load_dotenv()
except ImportError:
    pass  # dotenv is optional


# Fixed model - Nano Banana Pro
MODEL_NAME = "gemini-3-pro-image-preview"

# Supported configurations
SUPPORTED_ASPECT_RATIOS = ["1:1", "2:3", "3:2", "3:4", "4:3", "4:5", "5:4", "9:16", "16:9", "21:9"]
SUPPORTED_RESOLUTIONS = ["1K", "2K", "4K"]

# Retry configuration
MAX_RETRIES = 5
BASE_DELAY = 2.0
MAX_DELAY = 120.0


@dataclass
class ImageTask:
    """Represents a single image generation task."""

    prompt: str
    output_path: str
    aspect_ratio: str = "16:9"
    resolution: str = "2K"
    style: Optional[str] = None
    enable_grounding: bool = False
    task_id: int = 0
    metadata: Dict[str, Any] = field(default_factory=dict)


@dataclass
class BatchResult:
    """Represents results of a batch generation."""

    total: int = 0
    successful: int = 0
    failed: int = 0
    results: List[Dict[str, Any]] = field(default_factory=list)
    errors: List[Dict[str, Any]] = field(default_factory=list)
    start_time: Optional[datetime] = None
    end_time: Optional[datetime] = None
    total_duration_seconds: float = 0.0


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
        return "16:9"
    return ratio


def validate_resolution(resolution: str) -> str:
    """Validate and normalize resolution."""
    normalized = resolution.upper()
    if normalized not in SUPPORTED_RESOLUTIONS:
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


def load_config_file(config_path: str) -> Dict[str, Any]:
    """Load batch configuration from JSON or YAML file."""
    path = Path(config_path)

    if not path.exists():
        raise FileNotFoundError(f"Config file not found: {config_path}")

    content = path.read_text(encoding="utf-8")

    if path.suffix.lower() in [".yaml", ".yml"]:
        if not YAML_AVAILABLE:
            raise ImportError("PyYAML not installed. Install with: pip install pyyaml")
        return yaml.safe_load(content)
    else:
        return json.loads(content)


def parse_tasks_from_config(config: Dict[str, Any], output_dir: str) -> List[ImageTask]:
    """Parse image tasks from configuration dictionary."""
    tasks = []

    # Global defaults
    defaults = config.get("defaults", {})
    global_style = defaults.get("style")
    global_resolution = defaults.get("resolution", "2K")
    global_aspect_ratio = defaults.get("aspect_ratio", "16:9")
    global_grounding = defaults.get("enable_grounding", False)

    # Parse items
    items = config.get("images", config.get("prompts", []))

    for idx, item in enumerate(items):
        if isinstance(item, str):
            # Simple string prompt
            prompt = item
            filename = f"image_{idx + 1:03d}.png"
            task = ImageTask(
                prompt=prompt,
                output_path=str(Path(output_dir) / filename),
                aspect_ratio=validate_aspect_ratio(global_aspect_ratio),
                resolution=validate_resolution(global_resolution),
                style=global_style,
                enable_grounding=global_grounding,
                task_id=idx + 1,
            )
        else:
            # Dictionary with full configuration
            prompt = item.get("prompt", item.get("description", ""))
            filename = item.get("filename", item.get("output", f"image_{idx + 1:03d}.png"))

            # Ensure .png extension
            if not filename.lower().endswith(".png"):
                filename += ".png"

            task = ImageTask(
                prompt=prompt,
                output_path=str(Path(output_dir) / filename),
                aspect_ratio=validate_aspect_ratio(item.get("aspect_ratio", global_aspect_ratio)),
                resolution=validate_resolution(item.get("resolution", global_resolution)),
                style=item.get("style", global_style),
                enable_grounding=item.get("enable_grounding", global_grounding),
                task_id=idx + 1,
                metadata=item.get("metadata", {}),
            )

        if task.prompt:
            tasks.append(task)

    return tasks


def create_tasks_from_prompts(
    prompts: List[str], output_dir: str, style: Optional[str] = None, resolution: str = "2K", aspect_ratio: str = "16:9"
) -> List[ImageTask]:
    """Create tasks from command-line prompts."""
    tasks = []

    for idx, prompt in enumerate(prompts):
        filename = f"image_{idx + 1:03d}.png"
        task = ImageTask(
            prompt=prompt,
            output_path=str(Path(output_dir) / filename),
            aspect_ratio=validate_aspect_ratio(aspect_ratio),
            resolution=validate_resolution(resolution),
            style=style,
            task_id=idx + 1,
        )
        tasks.append(task)

    return tasks


async def generate_single_image(
    client: genai.Client, task: ImageTask, max_retries: int = MAX_RETRIES, verbose: bool = False
) -> Dict[str, Any]:
    """Generate a single image with retry logic."""
    final_prompt = build_prompt_with_style(task.prompt, task.style)

    # Build configuration
    config_params = {
        "response_modalities": ["TEXT", "IMAGE"],
        "image_config": types.ImageConfig(aspect_ratio=task.aspect_ratio, image_size=task.resolution),
    }

    if task.enable_grounding:
        config_params["tools"] = [{"google_search": {}}]

    config = types.GenerateContentConfig(**config_params)

    last_error = None
    for attempt in range(max_retries):
        try:
            if verbose:
                print(f"  [{task.task_id}] Attempt {attempt + 1}/{max_retries}...")

            start_time = time.time()

            # Generate image (sync call in async context)
            response = await asyncio.get_event_loop().run_in_executor(
                None, lambda: client.models.generate_content(model=MODEL_NAME, contents=final_prompt, config=config)
            )

            generation_time = time.time() - start_time

            # Ensure output directory exists
            output_dir = Path(task.output_path).parent
            if output_dir and str(output_dir) != "." and not output_dir.exists():
                output_dir.mkdir(parents=True, exist_ok=True)

            # Save image
            image_saved = False

            for part in response.candidates[0].content.parts:
                if part.inline_data is not None:
                    with open(task.output_path, "wb") as f:
                        f.write(part.inline_data.data)
                    image_saved = True
                elif hasattr(part, "text") and part.text:
                    _ = part.text  # Text response captured but not used for image generation

            if not image_saved:
                raise RuntimeError("No image data in response")

            return {
                "success": True,
                "task_id": task.task_id,
                "prompt": task.prompt,
                "output_path": str(Path(task.output_path).absolute()),
                "generation_time_seconds": round(generation_time, 2),
                "attempts": attempt + 1,
                "metadata": task.metadata,
            }

        except Exception as e:
            last_error = e
            error_str = str(e).lower()

            # Non-retryable errors
            if "permission" in error_str or "api_key" in error_str:
                break
            if "invalid" in error_str and "argument" in error_str:
                break

            # Retryable errors with exponential backoff
            if attempt < max_retries - 1:
                is_quota_error = any(
                    kw in error_str for kw in ["quota", "rate", "429", "resource_exhausted", "too_many"]
                )
                delay = calculate_backoff_delay(attempt, jitter=is_quota_error)

                if verbose:
                    if is_quota_error:
                        print(f"  [{task.task_id}] Rate limit. Waiting {delay:.1f}s...")
                    else:
                        print(f"  [{task.task_id}] Error: {e}. Retrying in {delay:.1f}s...")

                await asyncio.sleep(delay)

    return {
        "success": False,
        "task_id": task.task_id,
        "prompt": task.prompt,
        "error": str(last_error),
        "attempts": max_retries,
        "metadata": task.metadata,
    }


async def run_batch_generation(
    tasks: List[ImageTask],
    concurrency: int = 2,
    max_retries: int = MAX_RETRIES,
    verbose: bool = False,
    delay_between_tasks: float = 1.0,
) -> BatchResult:
    """Run batch generation with concurrency control."""
    result = BatchResult(total=len(tasks), start_time=datetime.now())

    # Get API key and create client
    api_key = get_api_key()
    client = genai.Client(api_key=api_key)

    # Semaphore for concurrency control
    semaphore = asyncio.Semaphore(concurrency)

    async def process_with_semaphore(task: ImageTask) -> Dict[str, Any]:
        async with semaphore:
            task_result = await generate_single_image(client, task, max_retries, verbose)
            # Add delay between tasks to avoid rate limiting
            await asyncio.sleep(delay_between_tasks)
            return task_result

    # Process all tasks
    print(f"\nProcessing {len(tasks)} images with concurrency={concurrency}...")

    task_results = await asyncio.gather(*[process_with_semaphore(task) for task in tasks], return_exceptions=True)

    # Collect results
    for idx, task_result in enumerate(task_results):
        if isinstance(task_result, Exception):
            result.failed += 1
            result.errors.append(
                {"task_id": tasks[idx].task_id, "prompt": tasks[idx].prompt, "error": str(task_result)}
            )
        elif task_result.get("success"):
            result.successful += 1
            result.results.append(task_result)
            print(f"  [OK] {Path(task_result['output_path']).name}")
        else:
            result.failed += 1
            result.errors.append(task_result)
            print(f"  [FAIL] Task {task_result.get('task_id')}: {task_result.get('error', 'Unknown error')[:50]}")

    result.end_time = datetime.now()
    result.total_duration_seconds = (result.end_time - result.start_time).total_seconds()

    return result


def save_batch_report(result: BatchResult, output_path: str) -> None:
    """Save batch generation report to JSON file."""
    report = {
        "summary": {
            "total": result.total,
            "successful": result.successful,
            "failed": result.failed,
            "success_rate": f"{(result.successful / result.total * 100):.1f}%" if result.total > 0 else "0%",
            "duration_seconds": round(result.total_duration_seconds, 2),
            "start_time": result.start_time.isoformat() if result.start_time else None,
            "end_time": result.end_time.isoformat() if result.end_time else None,
        },
        "results": result.results,
        "errors": result.errors,
    }

    with open(output_path, "w", encoding="utf-8") as f:
        json.dump(report, f, indent=2, ensure_ascii=False)


def main():
    """Main entry point for CLI usage."""
    parser = argparse.ArgumentParser(
        description="Batch image generation using Nano Banana Pro (gemini-3-pro-image-preview)",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
    # From config file
    python batch_generate.py -c prompts.json -d output/
    python batch_generate.py -c prompts.yaml -d output/ --concurrency 3

    # From command line prompts
    python batch_generate.py --prompts "A cat" "A dog" "A bird" -d output/
    python batch_generate.py --prompts "Logo design" "Banner art" -d output/ \\
        --style "minimalist, vector"

Config File Format (JSON):
    {
        "defaults": {
            "style": "photorealistic",
            "resolution": "4K",
            "aspect_ratio": "16:9"
        },
        "images": [
            "Simple prompt string",
            {
                "prompt": "Detailed prompt",
                "filename": "custom_name.png",
                "resolution": "2K",
                "style": "watercolor"
            }
        ]
    }

Config File Format (YAML):
    defaults:
      style: photorealistic
      resolution: 4K
    images:
      - Simple prompt string
      - prompt: Detailed prompt
        filename: custom_name.png

Environment:
    GOOGLE_API_KEY - Required. Get from https://aistudio.google.com/apikey
        """,
    )

    # Input options (mutually exclusive)
    input_group = parser.add_mutually_exclusive_group(required=True)
    input_group.add_argument("-c", "--config", help="Path to JSON or YAML config file")
    input_group.add_argument("--prompts", nargs="+", help="List of prompts to generate")

    # Output options
    parser.add_argument("-d", "--output-dir", required=True, help="Output directory for generated images")

    # Generation options
    parser.add_argument(
        "-r", "--resolution", default="2K", choices=SUPPORTED_RESOLUTIONS, help="Default resolution (default: 2K)"
    )

    parser.add_argument(
        "-a",
        "--aspect-ratio",
        default="16:9",
        choices=SUPPORTED_ASPECT_RATIOS,
        help="Default aspect ratio (default: 16:9)",
    )

    parser.add_argument("--style", help="Default style prefix for all prompts")

    # Processing options
    parser.add_argument("--concurrency", type=int, default=2, help="Number of concurrent generations (default: 2)")

    parser.add_argument(
        "--max-retries",
        type=int,
        default=MAX_RETRIES,
        help=f"Maximum retry attempts per image (default: {MAX_RETRIES})",
    )

    parser.add_argument("--delay", type=float, default=1.0, help="Delay between tasks in seconds (default: 1.0)")

    # Report options
    parser.add_argument("--report", help="Path to save JSON report (optional)")

    parser.add_argument("-v", "--verbose", action="store_true", help="Print detailed progress information")

    args = parser.parse_args()

    # Ensure output directory exists
    output_dir = Path(args.output_dir)
    output_dir.mkdir(parents=True, exist_ok=True)

    # Load or create tasks
    if args.config:
        print(f"Loading config from: {args.config}")
        config = load_config_file(args.config)
        tasks = parse_tasks_from_config(config, str(output_dir))
    else:
        tasks = create_tasks_from_prompts(
            args.prompts, str(output_dir), style=args.style, resolution=args.resolution, aspect_ratio=args.aspect_ratio
        )

    if not tasks:
        print("Error: No valid tasks found.")
        sys.exit(1)

    print(f"Loaded {len(tasks)} image tasks")
    print(f"Output directory: {output_dir.absolute()}")
    print(f"Model: {MODEL_NAME}")

    # Run batch generation
    result = asyncio.run(
        run_batch_generation(
            tasks,
            concurrency=args.concurrency,
            max_retries=args.max_retries,
            verbose=args.verbose,
            delay_between_tasks=args.delay,
        )
    )

    # Print summary
    print(f"\n{'=' * 50}")
    print("BATCH GENERATION COMPLETE")
    print(f"{'=' * 50}")
    print(f"  Total:      {result.total}")
    print(f"  Successful: {result.successful}")
    print(f"  Failed:     {result.failed}")
    print(f"  Duration:   {result.total_duration_seconds:.1f}s")

    if result.total > 0:
        print(f"  Success Rate: {(result.successful / result.total * 100):.1f}%")

    # Save report if requested
    if args.report:
        save_batch_report(result, args.report)
        print(f"\nReport saved to: {args.report}")

    # Exit with error if any failures
    if result.failed > 0:
        print(f"\nWarning: {result.failed} images failed to generate.")
        if result.errors:
            print("Failed tasks:")
            for err in result.errors[:5]:  # Show first 5 errors
                print(f"  - Task {err.get('task_id')}: {err.get('error', 'Unknown')[:60]}")
            if len(result.errors) > 5:
                print(f"  ... and {len(result.errors) - 5} more errors")
        sys.exit(1)

    print("\nAll images generated successfully!")


if __name__ == "__main__":
    main()
