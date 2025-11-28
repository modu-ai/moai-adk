#!/usr/bin/env python3
"""
API Documentation Generator for MoAI-ADK

Automatically generates MDX API reference pages from Python source code.
Extracts docstrings, type hints, and function signatures.
"""

import ast
import os
import sys
from pathlib import Path
from typing import List, Dict, Any, Optional
import json


class APIDocGenerator:
    """Generate API documentation from Python source code"""

    def __init__(self, source_dir: Path, output_dir: Path):
        self.source_dir = source_dir
        self.output_dir = output_dir
        self.modules = []

    def scan_modules(self):
        """Scan source directory for Python modules"""
        print(f"📂 Scanning {self.source_dir} for Python modules...")

        for root, dirs, files in os.walk(self.source_dir):
            # Skip __pycache__ and hidden directories
            dirs[:] = [d for d in dirs if not d.startswith('__') and not d.startswith('.')]

            for file in files:
                if file.endswith('.py') and not file.startswith('__'):
                    module_path = Path(root) / file
                    rel_path = module_path.relative_to(self.source_dir)

                    self.modules.append({
                        'path': module_path,
                        'relative': rel_path,
                        'name': str(rel_path.with_suffix('')).replace(os.sep, '.')
                    })

        print(f"✅ Found {len(self.modules)} modules")
        return self.modules

    def parse_module(self, module_path: Path) -> Dict[str, Any]:
        """Parse Python module and extract API information"""
        try:
            with open(module_path, 'r', encoding='utf-8') as f:
                source = f.read()

            tree = ast.parse(source)

            module_info = {
                'path': str(module_path),
                'docstring': ast.get_docstring(tree) or '',
                'functions': [],
                'classes': []
            }

            for node in ast.walk(tree):
                if isinstance(node, ast.FunctionDef):
                    # Skip private functions
                    if not node.name.startswith('_'):
                        func_info = self.extract_function_info(node)
                        module_info['functions'].append(func_info)

                elif isinstance(node, ast.ClassDef):
                    # Skip private classes
                    if not node.name.startswith('_'):
                        class_info = self.extract_class_info(node)
                        module_info['classes'].append(class_info)

            return module_info

        except Exception as e:
            print(f"⚠️  Error parsing {module_path}: {e}")
            return None

    def extract_function_info(self, node: ast.FunctionDef) -> Dict[str, Any]:
        """Extract function information from AST node"""
        args = []
        for arg in node.args.args:
            arg_type = self.get_annotation_string(arg.annotation) if arg.annotation else 'Any'
            args.append({
                'name': arg.arg,
                'type': arg_type
            })

        return_type = self.get_annotation_string(node.returns) if node.returns else 'Any'

        return {
            'name': node.name,
            'docstring': ast.get_docstring(node) or '',
            'args': args,
            'returns': return_type,
            'is_async': isinstance(node, ast.AsyncFunctionDef)
        }

    def extract_class_info(self, node: ast.ClassDef) -> Dict[str, Any]:
        """Extract class information from AST node"""
        methods = []
        for item in node.body:
            if isinstance(item, (ast.FunctionDef, ast.AsyncFunctionDef)):
                if not item.name.startswith('_') or item.name == '__init__':
                    methods.append(self.extract_function_info(item))

        return {
            'name': node.name,
            'docstring': ast.get_docstring(node) or '',
            'methods': methods
        }

    def get_annotation_string(self, annotation) -> str:
        """Convert AST annotation to string"""
        if annotation is None:
            return 'Any'

        if isinstance(annotation, ast.Name):
            return annotation.id
        elif isinstance(annotation, ast.Constant):
            return str(annotation.value)
        elif isinstance(annotation, ast.Subscript):
            base = self.get_annotation_string(annotation.value)
            slice_val = self.get_annotation_string(annotation.slice)
            return f"{base}[{slice_val}]"
        else:
            return 'Any'

    def generate_mdx_page(self, module: Dict[str, Any], module_info: Dict[str, Any]) -> str:
        """Generate MDX content for a module"""
        module_name = module['name']

        mdx_content = f"""# {module_name}

{module_info['docstring'] or f'API reference for `{module_name}` module.'}

## Overview

**Module Path**: `{module['relative']}`

"""

        # Generate functions section
        if module_info['functions']:
            mdx_content += "## Functions\n\n"

            for func in module_info['functions']:
                mdx_content += self.generate_function_docs(func, module_name)

        # Generate classes section
        if module_info['classes']:
            mdx_content += "## Classes\n\n"

            for cls in module_info['classes']:
                mdx_content += self.generate_class_docs(cls, module_name)

        # Add footer
        mdx_content += f"""
---

**Module**: `moai_adk.{module_name}`
**Last Updated**: 2025-11-28
"""

        return mdx_content

    def generate_function_docs(self, func: Dict[str, Any], module_name: str) -> str:
        """Generate MDX documentation for a function"""
        func_name = func['name']
        is_async = func['is_async']

        # Build signature
        args_str = ', '.join([f"{arg['name']}: {arg['type']}" for arg in func['args']])
        async_prefix = 'async ' if is_async else ''
        signature = f"{async_prefix}def {func_name}({args_str}) -> {func['returns']}"

        docs = f"""### {func_name}()

{func['docstring'] or f'Function `{func_name}` in module `{module_name}`.'}

**Signature**:
```python
{signature}
```

"""

        # Add parameters
        if func['args']:
            docs += "**Parameters**:\n"
            for arg in func['args']:
                docs += f"- `{arg['name']}` ({arg['type']}): Parameter description\n"
            docs += "\n"

        # Add returns
        docs += f"""**Returns**:
- `{func['returns']}`: Return value description

**Example**:
```python
from moai_adk.{module_name} import {func_name}

# Usage example
result = {func_name}(...)
```

---

"""
        return docs

    def generate_class_docs(self, cls: Dict[str, Any], module_name: str) -> str:
        """Generate MDX documentation for a class"""
        cls_name = cls['name']

        docs = f"""### {cls_name}

{cls['docstring'] or f'Class `{cls_name}` in module `{module_name}`.'}

**Example**:
```python
from moai_adk.{module_name} import {cls_name}

# Initialize instance
instance = {cls_name}()
```

"""

        # Add methods
        if cls['methods']:
            docs += "**Methods**:\n\n"
            for method in cls['methods']:
                method_name = method['name']
                args_str = ', '.join([f"{arg['name']}: {arg['type']}" for arg in method['args']])

                docs += f"""#### {method_name}()

{method['docstring'] or f'Method `{method_name}` of class `{cls_name}`.'}

```python
{method_name}({args_str}) -> {method['returns']}
```

"""

        docs += "---\n\n"
        return docs

    def create_navigation(self, modules: List[Dict]) -> str:
        """Create _meta.js navigation file"""
        # Organize modules by directory structure
        nav_structure = {
            'index': 'API Overview',
            '---core': {'type': 'separator', 'title': 'Core'},
            'cli': 'CLI',
            'core': 'Core',
            '---modules': {'type': 'separator', 'title': 'Modules'},
            'foundation': 'Foundation',
            'project': 'Project',
            'utils': 'Utilities',
        }

        return f"""export default {json.dumps(nav_structure, indent=2)}
"""

    def generate_all_docs(self):
        """Generate all API documentation"""
        print("🚀 Starting API documentation generation...")

        # Scan modules
        self.scan_modules()

        # Create output directory
        api_dir = self.output_dir / 'api'
        api_dir.mkdir(parents=True, exist_ok=True)

        # Generate index page
        index_content = """# API Reference

Complete API reference for MoAI-ADK Python package.

## Overview

MoAI-ADK (Multi-Agent AI Application Development Kit) provides a comprehensive framework for building AI-powered applications with Claude Code.

## Modules

### Core Modules
- [CLI](/api/cli) - Command-line interface
- [Core](/api/core) - Core functionality
- [Foundation](/api/foundation) - Foundation utilities

### Utility Modules
- [Project](/api/project) - Project management
- [Utils](/api/utils) - Utility functions

---

**Package**: `moai_adk`
**Version**: 0.30.2
**Last Updated**: 2025-11-28
"""

        index_path = api_dir / 'index.mdx'
        with open(index_path, 'w', encoding='utf-8') as f:
            f.write(index_content)

        print(f"✅ Created index: {index_path}")

        # Generate module pages
        pages_created = 0
        total_lines = 0

        for module in self.modules[:30]:  # Limit to 30 modules for now
            module_info = self.parse_module(module['path'])

            if module_info and (module_info['functions'] or module_info['classes']):
                mdx_content = self.generate_mdx_page(module, module_info)

                # Create output path
                output_path = api_dir / module['relative'].with_suffix('.mdx')
                output_path.parent.mkdir(parents=True, exist_ok=True)

                with open(output_path, 'w', encoding='utf-8') as f:
                    f.write(mdx_content)

                lines = len(mdx_content.splitlines())
                total_lines += lines
                pages_created += 1

                print(f"✅ Created: {output_path} ({lines} lines)")

        # Create navigation
        meta_path = api_dir / '_meta.js'
        nav_content = self.create_navigation(self.modules)
        with open(meta_path, 'w', encoding='utf-8') as f:
            f.write(nav_content)

        print(f"\n✅ Created navigation: {meta_path}")
        print(f"\n📊 Summary:")
        print(f"   - Pages Created: {pages_created}")
        print(f"   - Total Lines: {total_lines:,}")
        print(f"   - Average Lines/Page: {total_lines // pages_created if pages_created > 0 else 0}")

        return pages_created, total_lines


def main():
    """Main entry point"""
    # Define paths
    project_root = Path(__file__).parent.parent
    source_dir = project_root / 'src' / 'moai_adk'
    output_dir = project_root / 'docs' / 'pages'

    print(f"📁 Source: {source_dir}")
    print(f"📁 Output: {output_dir}")

    if not source_dir.exists():
        print(f"❌ Source directory not found: {source_dir}")
        sys.exit(1)

    # Generate documentation
    generator = APIDocGenerator(source_dir, output_dir)
    pages_created, total_lines = generator.generate_all_docs()

    print(f"\n✅ API documentation generation complete!")
    print(f"   Total: {pages_created} pages, {total_lines:,} lines")


if __name__ == '__main__':
    main()
