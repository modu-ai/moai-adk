#!/usr/bin/env python3
"""@DOCS:AUTOMATION-001 Generate API reference pages

Automatically generates API documentation from source code using the
MoAI-ADK API generator module.
"""
import os
from pathlib import Path

from moai_adk.core.docs.api_generator import ApiGenerator


def main():
    """Generate API reference pages"""
    project_root = Path(__file__).parent.parent
    docs_dir = project_root / "docs"

    generator = ApiGenerator(str(project_root), "src")

    # Generate API documentation
    generator.generate_api_docs(str(docs_dir))

    # Update mkdocs navigation
    mkdocs_config = project_root / "mkdocs.yml"
    generator.update_mkdocs_nav(str(mkdocs_config))

    # Create API index
    api_index_content = generator.generate_module_index()
    (docs_dir / "api" / "index.md").write_text(api_index_content)

    print("✅ API documentation generated successfully")


if __name__ == "__main__":
    main()