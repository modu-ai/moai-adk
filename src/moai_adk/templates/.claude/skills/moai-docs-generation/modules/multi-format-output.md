# Multi-Format Output Generation

## Overview
Comprehensive documentation generation supporting HTML sites, PDF documents, static sites, and various output formats with responsive design and professional styling.

## Quick Implementation

```python
from jinja2 import Template
import markdown

class HTMLDocGenerator:
 def __init__(self):
 self.templates = self.load_templates()

 def generate_html_site(self, documentation: Dict, output_dir: str):
 """Generate a complete HTML documentation site."""

 # Generate index page
 index_html = self.generate_index_page(documentation)
 self.write_file(output_dir, "index.html", index_html)

 # Generate API reference pages
 self.generate_api_pages(documentation, output_dir)

 # Generate tutorial pages
 self.generate_tutorial_pages(documentation, output_dir)

 # Generate CSS and assets
 self.generate_assets(output_dir)

 def generate_index_page(self, documentation: Dict) -> str:
 """Generate the main index page."""

 template = Template("""
<!DOCTYPE html>
<html lang="en">
<head>
 <meta charset="UTF-8">
 <meta name="viewport" content="width=device-width, initial-scale=1.0">
 <title>{{ title }}</title>
 <link rel="stylesheet" href="assets/style.css">
</head>
<body>
 <nav class="sidebar">
 <div class="logo">
 <h2>{{ title }}</h2>
 </div>
 <ul class="nav-menu">
 <li><a href="index.html" class="active">Home</a></li>
 <li><a href="api/index.html">API Reference</a></li>
 <li><a href="tutorials/index.html">Tutorials</a></li>
 <li><a href="cookbook.html">Cookbook</a></li>
 </ul>
 </nav>

 <main class="content">
 <section class="hero">
 <h1>{{ title }}</h1>
 <p>{{ description }}</p>
 <div class="cta-buttons">
 <a href="tutorials/getting-started.html" class="btn primary">Get Started</a>
 <a href="api/index.html" class="btn secondary">API Docs</a>
 </div>
 </section>

 <section class="features">
 <h2>Key Features</h2>
 <div class="feature-grid">
 {% for feature in features %}
 <div class="feature-card">
 <h3>{{ feature.name }}</h3>
 <p>{{ feature.description }}</p>
 </div>
 {% endfor %}
 </div>
 </section>

 <section class="quick-start">
 <h2>Quick Start</h2>
 <pre><code>{{ quick_start_example }}</code></pre>
 </section>
 </main>

 <script src="assets/script.js"></script>
</body>
</html>
 """)

 return template.render(
 title=documentation.get("title", "API Documentation"),
 description=documentation.get("description", ""),
 features=documentation.get("features", []),
 quick_start_example=documentation.get("quick_start", "")
 )

 def generate_pdf_documentation(self, markdown_content: str, output_path: str):
 """Generate PDF from Markdown content."""
 try:
 import weasyprint

 # Convert Markdown to HTML
 html_content = markdown.markdown(
 markdown_content,
 extensions=['tables', 'fenced_code', 'toc']
 )

 # Add CSS styling
 styled_html = f"""
 <!DOCTYPE html>
 <html>
 <head>
 <style>
 body {{ font-family: Arial, sans-serif; line-height: 1.6; margin: 40px; }}
 h1 {{ color: #333; border-bottom: 2px solid #333; }}
 h2 {{ color: #666; }}
 code {{ background: #f4f4f4; padding: 2px 4px; border-radius: 3px; }}
 pre {{ background: #f4f4f4; padding: 10px; border-radius: 5px; overflow-x: auto; }}
 table {{ border-collapse: collapse; width: 100%; }}
 th, td {{ border: 1px solid #ddd; padding: 8px; text-align: left; }}
 </style>
 </head>
 <body>
 {html_content}
 </body>
 </html>
 """

 # Generate PDF
 weasyprint.HTML(string=styled_html).write_pdf(output_path)

 except ImportError:
 raise ImportError("weasyprint is required for PDF generation. Install with: pip install weasyprint")
```

## Output Formats

### 1. HTML Documentation Sites
- Responsive design with sidebar navigation
- Interactive code examples
- Search functionality
- Dark/light theme support
- Mobile-optimized layouts

### 2. PDF Documentation
- Professional print-ready formatting
- Table of contents generation
- Code syntax highlighting
- Header/footer customization
- Page numbering and bookmarks

### 3. Static Site Generation
- Jekyll/Hugo compatible output
- SEO optimization
- Fast loading times
- GitHub Pages deployment
- Custom domain support

### 4. Markdown Variants
- GitHub Flavored Markdown
- CommonMark compliance
- Extension support (tables, footnotes)
- Cross-reference linking
- Image optimization

## Template System
- Jinja2-based templating
- Customizable themes
- Component-based design
- Asset management
- Multi-language support

## Integration Points
- Documentation platforms (GitBook, Docusaurus)
- CI/CD pipelines
- Static site hosting
- Content management systems
- Version control workflows
