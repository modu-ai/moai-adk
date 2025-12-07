#!/usr/bin/env python3
"""
Section Builder: Batch Section Generation with Parallel Processing

Purpose:
- Generate sections in parallel batches for performance optimization
- Process 5 sections simultaneously (configurable)
- Automatic retry on failure
- Progress tracking and logging

Performance:
- Sequential: 25 sections × 30 sec = 12.5 minutes
- Batch (5 parallel): 5 batches × 30 sec = 2.5 minutes
- Improvement: 80%

Author: GoosLab
Version: 1.0.0
"""

import asyncio
import logging
from typing import Any, Dict, List, Optional
from dataclasses import dataclass

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


@dataclass
class GenerationResult:
    """Section generation result."""
    section_title: str
    content: str
    success: bool
    error: Optional[str] = None
    retry_count: int = 0


class SectionBuilder:
    """Batch section generation with parallel processing."""
    
    def __init__(
        self,
        batch_size: int = 5,
        timeout: int = 60,
        max_retries: int = 3
    ):
        """
        Initialize Section Builder.
        
        Args:
            batch_size: Number of sections to process in parallel
            timeout: Timeout per section in seconds
            max_retries: Maximum retry attempts on failure
        """
        self.batch_size = batch_size
        self.timeout = timeout
        self.max_retries = max_retries
    
    async def build_sections_batch(
        self,
        sections: List[Dict],
        batch_size: Optional[int] = None,
        pages_target: int = 100
    ) -> List[str]:
        """
        Build sections in parallel batches.
        
        Args:
            sections: List of section data (from PlanExpander)
            batch_size: Override default batch size
            pages_target: Target total pages
        
        Returns:
            List of generated section content (markdown strings)
        """
        if batch_size:
            self.batch_size = batch_size
        
        logger.info(f"Building {len(sections)} sections in batches of {self.batch_size}")
        logger.info(f"Target: {pages_target} pages total")
        
        # Split into batches
        batches = self._split_into_batches(sections)
        
        logger.info(f"Created {len(batches)} batches")
        
        # Process each batch
        all_generated = []
        
        for i, batch in enumerate(batches):
            logger.info(f"Processing batch {i+1}/{len(batches)} ({len(batch)} sections)")
            
            # Progress bar
            progress = (i / len(batches)) * 100
            progress_bar = self._create_progress_bar(progress)
            logger.info(f"Progress: {progress_bar} {progress:.1f}%")
            
            # Generate sections in parallel within batch
            batch_results = await self._process_batch(batch)
            
            # Collect results
            for result in batch_results:
                if result.success:
                    all_generated.append(result.content)
                else:
                    logger.error(f"Failed to generate section: {result.section_title} - {result.error}")
                    # Add placeholder
                    all_generated.append(f"# {result.section_title}\n\n[Generation failed: {result.error}]\n")
        
        logger.info(f"✓ Generated {len(all_generated)} sections")
        
        return all_generated
    
    def _split_into_batches(self, sections: List[Dict]) -> List[List[Dict]]:
        """
        Split sections into batches.
        
        Args:
            sections: List of section data
        
        Returns:
            List of batches
        """
        batches = []
        
        for i in range(0, len(sections), self.batch_size):
            batch = sections[i:i + self.batch_size]
            batches.append(batch)
        
        return batches
    
    async def _process_batch(self, batch: List[Dict]) -> List[GenerationResult]:
        """
        Process single batch in parallel.
        
        Args:
            batch: List of section data
        
        Returns:
            List of GenerationResult objects
        """
        # Create tasks for parallel execution
        tasks = [
            self._generate_section_with_retry(section)
            for section in batch
        ]
        
        # Execute in parallel
        results = await asyncio.gather(*tasks, return_exceptions=True)
        
        # Convert exceptions to GenerationResult
        processed_results = []
        
        for i, result in enumerate(results):
            if isinstance(result, Exception):
                processed_results.append(GenerationResult(
                    section_title=batch[i].get('title', 'Unknown'),
                    content="",
                    success=False,
                    error=str(result)
                ))
            else:
                processed_results.append(result)
        
        return processed_results
    
    async def _generate_section_with_retry(self, section: Dict) -> GenerationResult:
        """
        Generate section with automatic retry on failure.
        
        Args:
            section: Section data
        
        Returns:
            GenerationResult
        """
        section_title = section.get('title', 'Untitled')
        retry_count = 0
        
        while retry_count <= self.max_retries:
            try:
                # Generate section content
                content = await asyncio.wait_for(
                    self._generate_section(section),
                    timeout=self.timeout
                )
                
                return GenerationResult(
                    section_title=section_title,
                    content=content,
                    success=True,
                    retry_count=retry_count
                )
            
            except asyncio.TimeoutError:
                retry_count += 1
                logger.warning(f"Timeout generating section '{section_title}' (retry {retry_count}/{self.max_retries})")
                
                if retry_count > self.max_retries:
                    return GenerationResult(
                        section_title=section_title,
                        content="",
                        success=False,
                        error="Timeout exceeded",
                        retry_count=retry_count
                    )
            
            except Exception as e:
                retry_count += 1
                logger.error(f"Error generating section '{section_title}': {e} (retry {retry_count}/{self.max_retries})")
                
                if retry_count > self.max_retries:
                    return GenerationResult(
                        section_title=section_title,
                        content="",
                        success=False,
                        error=str(e),
                        retry_count=retry_count
                    )
        
        # Should not reach here
        return GenerationResult(
            section_title=section_title,
            content="",
            success=False,
            error="Unknown error"
        )
    
    async def _generate_section(self, section: Dict) -> str:
        """
        Generate section content.
        
        Args:
            section: Section data (from PlanExpander)
        
        Returns:
            Generated markdown content
        """
        # Simulate content generation (replace with actual generation)
        await asyncio.sleep(0.1)  # Simulate processing time
        
        # Extract section data
        title = section.get('title', 'Untitled')
        content = section.get('content', '')
        subsections = section.get('subsections', [])
        code_examples = section.get('code_examples', [])
        
        # Build markdown
        parts = []
        
        # Title
        parts.append(f"# {title}\n\n")
        
        # Content
        if content:
            parts.append(f"{content}\n\n")
        
        # Subsections
        for subsection in subsections:
            parts.append(f"## {subsection.get('title', 'Subsection')}\n\n")
            parts.append(f"{subsection.get('content', '')}\n\n")
        
        # Code examples
        if code_examples:
            parts.append("## Examples\n\n")
            for i, example in enumerate(code_examples, 1):
                parts.append(f"### Example {i}\n\n")
                parts.append(f"```python\n{example}\n```\n\n")
        
        return "".join(parts)
    
    def _create_progress_bar(self, progress: float, width: int = 20) -> str:
        """
        Create text-based progress bar.
        
        Args:
            progress: Progress percentage (0-100)
            width: Width of progress bar
        
        Returns:
            Progress bar string
        """
        filled = int((progress / 100) * width)
        empty = width - filled
        
        return f"[{'=' * filled}{' ' * empty}]"
    
    async def parallel_section_generation(
        self,
        sections: List[Dict]
    ) -> List[str]:
        """
        Alternative method: Generate all sections in parallel (no batching).
        
        Use this for small section counts (< 10).
        
        Args:
            sections: List of section data
        
        Returns:
            List of generated content
        """
        logger.info(f"Generating {len(sections)} sections in parallel (no batching)")
        
        # Create tasks
        tasks = [
            self._generate_section_with_retry(section)
            for section in sections
        ]
        
        # Execute all in parallel
        results = await asyncio.gather(*tasks, return_exceptions=True)
        
        # Extract content
        generated = []
        
        for result in results:
            if isinstance(result, GenerationResult) and result.success:
                generated.append(result.content)
            else:
                generated.append(f"[Generation failed]\n")
        
        return generated
    
    def merge_sections_to_markdown(self, sections: List[str]) -> str:
        """
        Merge generated sections into single markdown document.
        
        Args:
            sections: List of section markdown strings
        
        Returns:
            Complete markdown document
        """
        # Add page breaks between sections
        return "\n\n---\n\n".join(sections)
    
    def validate_section_structure(self, section: str) -> bool:
        """
        Validate section structure.
        
        Args:
            section: Section markdown content
        
        Returns:
            True if valid, False otherwise
        """
        # Basic validation
        if not section.strip():
            return False
        
        # Check for title (starts with #)
        if not section.strip().startswith('#'):
            return False
        
        # Check minimum length
        if len(section) < 100:
            return False
        
        return True


# Example usage
async def main():
    """Example usage of SectionBuilder."""
    
    builder = SectionBuilder(batch_size=5)
    
    # Sample sections
    sections = [
        {"title": f"Section {i}", "content": f"Content for section {i}"}
        for i in range(1, 26)  # 25 sections
    ]
    
    # Build in batches
    generated = await builder.build_sections_batch(
        sections=sections,
        pages_target=100
    )
    
    print(f"Generated {len(generated)} sections")
    
    # Merge to single document
    document = builder.merge_sections_to_markdown(generated)
    
    print(f"Final document: {len(document)} characters")


if __name__ == "__main__":
    asyncio.run(main())
