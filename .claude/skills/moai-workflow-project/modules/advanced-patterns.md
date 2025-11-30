# Advanced Implementation Patterns

Advanced patterns for custom template development, performance optimization, and integration workflows.

## Custom Template Development

Documentation Templates:
```python
# Custom template for specific project type
custom_template = {
 "project_type": "mobile_application",
 "language": "ko",
 "sections": {
 "product": {
 "mission": " ",
 "metrics": [" ", " ", " "],
 "success_criteria": " "
 },
 "tech": {
 "frameworks": ["Flutter", "React Native", "Swift"],
 "performance_targets": " "
 }
 }
}

# Generate custom documentation
docs = project.documentation_manager._generate_product_doc(
 "mobile_application", "ko"
)
```

Language-Specific Customization:
```python
# Add custom language support
custom_language_config = {
 "code": "de",
 "name": "German",
 "native_name": "Deutsch",
 "locale": "de_DE.UTF-8",
 "date_format": "%d.%m.%Y",
 "rtl": False
}

# Register custom language
project.language_initializer.LANGUAGE_CONFIG["de"] = custom_language_config
```

## Performance Optimization Strategies

Template Caching:
```python
# Enable template caching for performance
project.template_optimizer.optimization_cache = {}

# Cache optimization results
def cached_optimization(template_path):
 cache_key = f"opt_{template_path}_{datetime.now().strftime('%Y%m%d')}"
 
 if cache_key not in project.template_optimizer.optimization_cache:
 result = project.template_optimizer._optimize_template_file(template_path)
 project.template_optimizer.optimization_cache[cache_key] = result
 
 return project.template_optimizer.optimization_cache[cache_key]
```

Batch Processing:
```python
# Process multiple templates efficiently
def batch_optimize_templates(template_paths):
 results = []
 
 for template_path in template_paths:
 try:
 result = project.template_optimizer._optimize_template_file(template_path)
 results.append(result)
 except Exception as e:
 results.append({
 "file_path": template_path,
 "success": False,
 "error": str(e)
 })
 
 return results
```

## Integration Workflows

Complete Project Lifecycle:
```python
def full_project_lifecycle():
 """Complete project setup and management workflow."""
 
 # Phase 1: Project Initialization
 project = MoaiMenuProject("./new-project")
 init_result = project.initialize_complete_project(
 language="en",
 domains=["backend", "frontend"],
 optimization_enabled=True
 )
 
 # Phase 2: Feature Development with SPEC
 spec_data = {
 "id": "SPEC-001",
 "title": "Core Feature Implementation",
 "requirements": ["Requirement 1", "Requirement 2"],
 "api_endpoints": [/* ... */]
 }
 
 docs_result = project.generate_documentation_from_spec(spec_data)
 
 # Phase 3: Performance Optimization
 optimization_result = project.optimize_project_templates()
 
 # Phase 4: Documentation Export
 export_result = project.export_project_documentation("html")
 
 # Phase 5: Backup Creation
 backup_result = project.create_project_backup()
 
 return {
 "initialization": init_result,
 "documentation": docs_result,
 "optimization": optimization_result,
 "export": export_result,
 "backup": backup_result
 }
```

Multilingual Project Management:
```python
def multilingual_project_workflow():
 """Manage multilingual project with cost optimization."""
 
 project = MoaiMenuProject("./multilingual-project")
 
 # Initialize with primary language
 init_result = project.initialize_complete_project(
 language="ko",
 user_name=" ",
 domains=["backend", "frontend"]
 )
 
 # Optimize agent prompts for cost (use English)
 lang_update = project.update_language_settings({
 "language.agent_prompt_language": "english"
 })
 
 # Generate documentation in multiple languages
 for lang in ["ko", "en", "ja"]:
 export_result = project.export_project_documentation("markdown", lang)
 print(f"{lang} documentation exported: {export_result['success']}")
 
 return init_result
```

---

For detailed implementation patterns, refer to the main SKILL.md Implementation Guide section.
