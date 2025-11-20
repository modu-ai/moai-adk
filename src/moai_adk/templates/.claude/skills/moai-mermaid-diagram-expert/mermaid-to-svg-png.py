#!/usr/bin/env python3
"""
Mermaid to SVG/PNG Converter - Production Edition
Version: 1.0.0
Requirements: Python 3.8+

Install:
  pip install playwright click pydantic pillow
  playwright install chromium
"""

import asyncio
import json
import logging
import re
import sys
import time
from dataclasses import asdict, dataclass
from pathlib import Path
from typing import Any, Dict, List, Literal, Optional

try:
    import click
    from playwright.async_api import Browser, async_playwright
    from pydantic import BaseModel, Field, validator
except ImportError as e:
    print(f"Error: {e}")
    print("\nPlease install required packages:")
    print("  pip install playwright click pydantic pillow")
    print("  playwright install chromium")
    sys.exit(1)

# ============================================================================
# Configuration & Validation
# ============================================================================

class MermaidConfig(BaseModel):
    """Configuration for Mermaid rendering"""
    input_path: Path
    output_dir: Path = Field(default_factory=Path.cwd)
    output_format: Literal['svg', 'png'] = 'svg'
    width: int = 1024
    height: int = 768
    theme: Literal['default', 'dark', 'forest'] = 'default'
    dpi: int = 96
    batch_mode: bool = False
    watch_mode: bool = False
    no_overwrite: bool = False
    dry_run: bool = False
    json_output: bool = False
    quiet: bool = False
    validate_only: bool = False

    @validator('width', 'height')
    def validate_dimensions(cls, v):  # noqa: N805
        if v < 100:
            raise ValueError('Minimum width/height: 100px')
        if v > 4096:
            raise ValueError('Maximum width/height: 4096px')
        return v

@dataclass
class ConversionResult:
    """Result of single file conversion"""
    input_file: Path
    output_file: Optional[Path] = None
    success: bool = False
    error_message: Optional[str] = None
    execution_time: float = 0.0
    file_size: Optional[int] = None
    diagram_type: Optional[str] = None

# ============================================================================
# Logging
# ============================================================================

def setup_logging(quiet: bool = False) -> logging.Logger:
    """Configure logging"""
    log_level = logging.WARNING if quiet else logging.INFO

    logger = logging.getLogger('mermaid_converter')
    logger.setLevel(log_level)

    # File handler
    fh = logging.FileHandler('mermaid-converter.log')
    fh.setLevel(logging.DEBUG)

    # Console handler
    ch = logging.StreamHandler()
    ch.setLevel(log_level)

    # Formatter
    formatter = logging.Formatter('[%(asctime)s] %(levelname)s - %(message)s')
    fh.setFormatter(formatter)
    ch.setFormatter(formatter)

    logger.addHandler(fh)
    logger.addHandler(ch)

    return logger

# ============================================================================
# Mermaid Syntax Validator
# ============================================================================

class MermaidValidator:
    """Validate Mermaid diagram syntax"""

    DIAGRAM_TYPES = {
        'flowchart': r'^(graph|flowchart)\s+(TD|LR|RL|BT)',
        'sequence': r'^sequenceDiagram',
        'class': r'^classDiagram',
        'state': r'^stateDiagram-v2',
        'er': r'^erDiagram',
        'gantt': r'^gantt',
        'mindmap': r'^mindmap',
        'timeline': r'^timeline',
        'gitgraph': r'^gitGraph',
        'pie': r'^pie',
        'journey': r'^journey',
        'block': r'^block-beta',
        'c4': r'^C4Context',
        'sankey': r'^sankey-beta',
        'quadrant': r'^quadrantChart',
        'requirement': r'^requirementDiagram',
        'xychart': r'^xychart-beta',
        'kanban': r'^kanban',
        'packet': r'^packet-beta',
        'radar': r'^radar',
    }

    def __init__(self):
        self.errors: List[str] = []
        self.warnings: List[str] = []

    def validate(self, mermaid_code: str) -> Dict[str, Any]:
        """Comprehensive validation"""
        self.errors = []
        self.warnings = []

        if not mermaid_code.strip():
            self.errors.append("Empty diagram")
            return {"valid": False, "diagram_type": None, "errors": self.errors}

        diagram_type = self._detect_type(mermaid_code)
        if not diagram_type:
            self.errors.append("Unknown diagram type")

        if self._has_unbalanced_blocks(mermaid_code):
            self.errors.append("Unbalanced subgraph/block/end keywords")

        self._check_syntax(mermaid_code)

        return {
            "valid": len(self.errors) == 0,
            "diagram_type": diagram_type,
            "errors": self.errors,
            "warnings": self.warnings
        }

    def _detect_type(self, code: str) -> Optional[str]:
        """Detect diagram type"""
        for dtype, pattern in self.DIAGRAM_TYPES.items():
            if re.search(pattern, code, re.MULTILINE):
                return dtype
        return None

    def _has_unbalanced_blocks(self, code: str) -> bool:
        """Check for balanced blocks"""
        opens = code.count('subgraph') + code.count('block')
        closes = code.count('end')
        return opens != closes

    def _check_syntax(self, code: str) -> None:
        """Check for common syntax errors"""
        if "'" in code and '"' not in code:
            self.warnings.append("Apostrophes detected; consider using double quotes")

        # Check for incomplete connections
        if '-->' in code or '->' in code:
            lines = code.split('\n')
            for line in lines:
                if re.search(r'--+[>x]?\s*$', line):
                    self.warnings.append(f"Incomplete connection: {line.strip()}")

# ============================================================================
# Mermaid Converter
# ============================================================================

class MermaidConverter:
    """Async Mermaid to SVG/PNG converter"""

    def __init__(self, config: MermaidConfig, logger: logging.Logger):
        self.config = config
        self.logger = logger
        self.browser: Optional[Browser] = None
        self.results: List[ConversionResult] = []
        self.validator = MermaidValidator()

    async def initialize(self) -> None:
        """Initialize browser"""
        try:
            playwright = await async_playwright().start()
            self.browser = await playwright.chromium.launch(headless=True)
            self.logger.info("Browser initialized")
        except Exception as e:
            self.logger.error(f"Browser initialization failed: {e}")
            raise

    async def cleanup(self) -> None:
        """Cleanup resources"""
        if self.browser:
            try:
                await self.browser.close()
                self.logger.info("Browser closed")
            except Exception as e:
                self.logger.error(f"Error closing browser: {e}")

    async def convert_single(self, input_file: Path) -> ConversionResult:
        """Convert single Mermaid file"""
        result = ConversionResult(input_file=input_file)
        start_time = time.time()

        try:
            # Check file exists
            if not input_file.exists():
                result.error_message = f"File not found: {input_file}"
                self.logger.warning(result.error_message)
                return result

            if not input_file.is_file():
                result.error_message = f"Not a file: {input_file}"
                self.logger.warning(result.error_message)
                return result

            # Read content
            mermaid_code = input_file.read_text(encoding='utf-8')

            # Validate
            validation = self.validator.validate(mermaid_code)
            result.diagram_type = validation['diagram_type']

            if not validation['valid']:
                result.error_message = f"Syntax error: {validation['errors'][0]}"
                self.logger.error(result.error_message)
                return result

            if validation['warnings']:
                self.logger.warning(f"Warnings: {validation['warnings']}")

            # Validate-only mode
            if self.config.validate_only:
                result.success = True
                result.execution_time = time.time() - start_time
                self.logger.info(f"Validated: {input_file.name}")
                return result

            # Determine output file
            output_file = self._get_output_path(input_file)

            # Check overwrite
            if output_file.exists() and self.config.no_overwrite:
                result.success = True
                result.output_file = output_file
                result.execution_time = time.time() - start_time
                self.logger.info(f"Skipped (exists): {output_file.name}")
                return result

            # Ensure output directory exists
            output_file.parent.mkdir(parents=True, exist_ok=True)

            # Render
            if self.config.output_format == 'svg':
                content = await self._render_svg(mermaid_code)
                if not self.config.dry_run:
                    output_file.write_text(content)
                result.file_size = len(content)
            else:  # PNG
                content = await self._render_png(mermaid_code)
                if not self.config.dry_run:
                    output_file.write_bytes(content)
                result.file_size = len(content)

            result.output_file = output_file
            result.success = True
            result.execution_time = time.time() - start_time

            self.logger.info(
                f"Converted: {input_file.name} -> {output_file.name} "
                f"({result.file_size} bytes, {result.execution_time:.2f}s)"
            )

        except Exception as e:
            result.error_message = str(e)
            result.execution_time = time.time() - start_time
            self.logger.error(f"Error converting {input_file}: {e}")

        return result

    async def _render_svg(self, mermaid_code: str) -> str:
        """Render to SVG using Playwright"""
        page = await self.browser.new_page(
            viewport={'width': self.config.width, 'height': self.config.height},
            device_scale_factor=self.config.dpi / 96
        )

        try:
            html = f"""
            <!DOCTYPE html>
            <html>
            <head>
                <meta charset="utf-8">
                <script src="https://cdn.jsdelivr.net/npm/mermaid@11/dist/mermaid.min.js"></script>
                <style>
                    body {{
                        margin: 0;
                        padding: 20px;
                        background: white;
                    }}
                    .mermaid {{
                        display: flex;
                        justify-content: center;
                        align-items: center;
                    }}
                </style>
            </head>
            <body>
                <div class="mermaid">
                {mermaid_code}
                </div>
                <script>
                    mermaid.initialize({{
                        startOnLoad: true,
                        theme: '{self.config.theme}',
                        securityLevel: 'loose'
                    }});
                    mermaid.contentLoaded();
                </script>
            </body>
            </html>
            """

            await page.set_content(html)

            # Wait for SVG to render
            try:
                await page.wait_for_selector('.mermaid svg', timeout=15000)
            except Exception as e:
                self.logger.warning(f"Timeout waiting for SVG: {e}")

            # Extract SVG
            svg = await page.evaluate("""
                () => {
                    const elem = document.querySelector('.mermaid svg');
                    return elem ? elem.outerHTML : null;
                }
            """)

            if not svg:
                raise RuntimeError("Failed to render SVG")

            return svg

        finally:
            await page.close()

    async def _render_png(self, mermaid_code: str) -> bytes:
        """Render to PNG"""
        page = await self.browser.new_page(
            viewport={'width': self.config.width, 'height': self.config.height},
            device_scale_factor=self.config.dpi / 96
        )

        try:
            html = f"""
            <!DOCTYPE html>
            <html>
            <head>
                <meta charset="utf-8">
                <script src="https://cdn.jsdelivr.net/npm/mermaid@11/dist/mermaid.min.js"></script>
                <style>
                    body {{
                        margin: 0;
                        padding: 20px;
                        background: white;
                    }}
                </style>
            </head>
            <body>
                <div class="mermaid">
                {mermaid_code}
                </div>
                <script>
                    mermaid.initialize({{
                        startOnLoad: true,
                        theme: '{self.config.theme}',
                        securityLevel: 'loose'
                    }});
                    mermaid.contentLoaded();
                </script>
            </body>
            </html>
            """

            await page.set_content(html)
            await page.wait_for_selector('.mermaid', timeout=15000)

            png_bytes = await page.screenshot(full_page=True)

            return png_bytes

        finally:
            await page.close()

    async def convert_batch(self, input_dir: Path) -> List[ConversionResult]:
        """Batch convert all .mmd files"""
        if not input_dir.is_dir():
            raise ValueError(f"Not a directory: {input_dir}")

        mmd_files = sorted(input_dir.rglob('*.mmd'))
        self.logger.info(f"Found {len(mmd_files)} Mermaid files")

        if not mmd_files:
            self.logger.warning("No .mmd files found")

        for input_file in mmd_files:
            result = await self.convert_single(input_file)
            self.results.append(result)

        return self.results

    def _get_output_path(self, input_file: Path) -> Path:
        """Generate output file path"""
        stem = input_file.stem
        ext = 'svg' if self.config.output_format == 'svg' else 'png'
        return self.config.output_dir / f"{stem}.{ext}"

    def print_summary(self, json_format: bool = False) -> None:
        """Print summary"""
        successful = sum(1 for r in self.results if r.success)
        total = len(self.results)

        if json_format:
            output = [asdict(r) for r in self.results]
            # Convert Path objects to strings
            for item in output:
                if item['input_file']:
                    item['input_file'] = str(item['input_file'])
                if item['output_file']:
                    item['output_file'] = str(item['output_file'])
            print(json.dumps(output, indent=2, default=str))

        summary_msg = f"\nConversion Summary: {successful}/{total} successful"
        if not self.config.quiet:
            click.echo(summary_msg)

        self.logger.info(summary_msg)

# ============================================================================
# CLI
# ============================================================================

@click.command()
@click.argument('input', type=click.Path(exists=True), required=True)
@click.option('-o', '--output', type=click.Path(), help='Output file or directory')
@click.option('-f', '--format', type=click.Choice(['svg', 'png']), default='svg',
              help='Output format (default: svg)')
@click.option('-w', '--width', type=int, default=1024,
              help='Image width in pixels (default: 1024)')
@click.option('-h', '--height', type=int, default=768,
              help='Image height in pixels (default: 768)')
@click.option('-t', '--theme', type=click.Choice(['default', 'dark', 'forest']),
              default='default', help='Mermaid theme (default: default)')
@click.option('--dpi', type=int, default=96,
              help='DPI for PNG rendering (default: 96)')
@click.option('-b', '--batch', is_flag=True,
              help='Batch mode: process entire directory')
@click.option('--no-overwrite', is_flag=True,
              help='Skip files that already exist')
@click.option('--dry-run', is_flag=True,
              help='Preview without saving files')
@click.option('--json', is_flag=True,
              help='Output results as JSON')
@click.option('-q', '--quiet', is_flag=True,
              help='Suppress output messages')
@click.option('--validate', is_flag=True,
              help='Validate syntax only (no conversion)')
@click.version_option('1.0.0')
def main(input, output, format, width, height, theme, dpi, batch, no_overwrite,
         dry_run, json, quiet, validate):
    """
    Convert Mermaid diagrams to SVG or PNG
    
    Examples:
    \b
      # Single file to SVG
      mermaid-to-svg-png.py diagram.mmd --output diagram.svg
    \b
      # Batch PNG conversion
      mermaid-to-svg-png.py ./diagrams --format png --batch --output ./images
    \b
      # Validate only
      mermaid-to-svg-png.py diagram.mmd --validate
    \b
      # CI/CD usage
      mermaid-to-svg-png.py ./docs --format png --json --quiet > results.json
    """

    logger = setup_logging(quiet)

    try:
        input_path = Path(input)
        output_dir = Path(output) if output else Path.cwd()

        config = MermaidConfig(
            input_path=input_path,
            output_dir=output_dir,
            output_format=format,
            width=width,
            height=height,
            theme=theme,
            dpi=dpi,
            batch_mode=batch,
            no_overwrite=no_overwrite,
            dry_run=dry_run,
            json_output=json,
            quiet=quiet,
            validate_only=validate,
        )

        converter = MermaidConverter(config, logger)

        async def run():
            await converter.initialize()
            try:
                if batch or input_path.is_dir():
                    await converter.convert_batch(input_path)
                else:
                    result = await converter.convert_single(input_path)
                    converter.results = [result]
            finally:
                await converter.cleanup()

        asyncio.run(run())

        # Print summary
        converter.print_summary(json_format=json)

        # Exit code
        successful = sum(1 for r in converter.results if r.success)
        exit_code = 0 if successful == len(converter.results) else 1
        sys.exit(exit_code)

    except Exception as e:
        logger.error(f"Fatal error: {e}")
        sys.exit(1)

# ============================================================================
# Entry Point
# ============================================================================

if __name__ == '__main__':
    main()
