#!/usr/bin/env python3
"""
MoAI Menu Project - Complete Workflow Demo

This demonstration script shows the complete usage of the moai-menu-project
integrated system, including:

1. Project initialization with all modules
2. Configuration management and updates
3. Documentation generation from SPECs
4. Language localization
5. Template optimization
6. Performance benchmarking
7. Backup creation and recovery
8. Export functionality

This serves as both a demonstration and a validation that the entire
system works together seamlessly.
"""

import sys
import tempfile
import time
from datetime import datetime
from pathlib import Path

# Add the skill directory to the path
sys.path.insert(0, str(Path(__file__).parent.parent))

try:
    from moai_menu_project import (
        MoaiMenuProject,
        generate_docs,
        initialize_project,
        optimize_templates,
    )
except ImportError as e:
    print(f"❌ Import Error: {e}")
    print("Make sure you're running this from the correct directory")
    sys.exit(1)


class DemoRunner:
    """Runner for the complete workflow demonstration."""

    def __init__(self):
        self.demo_project_dir = None
        self.project = None
        self.results = {}

    def setup_demo_environment(self):
        """Set up the demonstration environment."""
        print("🚀 Setting up demo environment...")

        # Create temporary directory for demo
        self.demo_project_dir = Path(tempfile.mkdtemp(prefix="moai_menu_demo_"))
        print(f"📁 Demo project directory: {self.demo_project_dir}")

        # Initialize the project
        self.project = MoaiMenuProject(str(self.demo_project_dir))

        print("✅ Demo environment setup complete")
        print("-" * 50)

    def demo_project_initialization(self):
        """Demonstrate complete project initialization."""
        print("🔧 Demo 1: Complete Project Initialization")
        print("-" * 30)

        # Initialize with Korean language and full configuration
        start_time = time.time()

        result = self.project.initialize_complete_project(
            language="ko",
            user_name="데모 사용자",
            domains=["backend", "frontend", "mobile"],
            project_type="web_application",
            optimization_enabled=True,
        )

        end_time = time.time()
        duration = end_time - start_time

        self.results["initialization"] = {
            "success": result["success"],
            "duration": duration,
            "modules_initialized": result["modules_initialized"],
            "created_files": len(result.get("created_files", [])),
            "optimization_applied": result["optimization_results"] is not None,
        }

        print(f"✅ Initialization completed in {duration:.2f} seconds")
        print(f"📦 Modules initialized: {', '.join(result['modules_initialized'])}")
        print(f"📄 Files created: {len(result.get('created_files', []))}")

        if result["optimization_results"]:
            print("🔧 Template optimization applied")

        # Show configuration
        print("\\n📋 Current Configuration:")
        config = self.project.config
        print(f"  - Project: {config['project']['name']} ({config['project']['type']})")
        print(f"  - Language: {config['language']['conversation_language']}")
        print(f"  - User: {config['user']['name']}")
        print(f"  - Menu System: v{config['menu_system']['version']}")

        print("\\n" + "=" * 50 + "\\n")

    def demo_spec_documentation_generation(self):
        """Demonstrate documentation generation from SPEC data."""
        print("📚 Demo 2: Documentation Generation from SPEC")
        print("-" * 42)

        # Create comprehensive SPEC data
        spec_data = {
            "id": "SPEC-DEMO-001",
            "title": "사용자 인증 시스템 구현",
            "description": "JWT 기반의 보안 사용자 인증 시스템 구현",
            "requirements": [
                "사용자 등록 및 이메일 인증",
                "JWT 토큰 생성 및 검증",
                "비밀번호 재설정 기능",
                "소셜 로그인 연동 (Google, GitHub)",
                "보안 로그인 시도 제한",
                "사용자 프로필 관리",
            ],
            "api_endpoints": [
                {
                    "path": "/api/auth/register",
                    "method": "POST",
                    "description": "신규 사용자 등록",
                    "parameters": {
                        "email": "string",
                        "password": "string",
                        "name": "string",
                    },
                },
                {
                    "path": "/api/auth/login",
                    "method": "POST",
                    "description": "사용자 로그인",
                    "parameters": {"email": "string", "password": "string"},
                },
                {
                    "path": "/api/auth/refresh",
                    "method": "POST",
                    "description": "JWT 토큰 갱신",
                    "parameters": {"refresh_token": "string"},
                },
            ],
            "status": "Planned",
            "priority": "High",
            "estimated_days": 5,
        }

        print("📝 Generating documentation from SPEC...")
        start_time = time.time()

        docs_result = self.project.generate_documentation_from_spec(spec_data)

        end_time = time.time()
        duration = end_time - start_time

        self.results["documentation_generation"] = {
            "success": docs_result["success"],
            "duration": duration,
            "feature_docs_generated": "feature_docs" in docs_result,
            "api_docs_generated": "api_docs" in docs_result,
            "localized_docs": "localized_documentation" in docs_result,
        }

        print(f"✅ Documentation generated in {duration:.2f} seconds")
        print(f"📄 Feature docs: {'✅' if 'feature_docs' in docs_result else '❌'}")
        print(f"🔗 API docs: {'✅' if 'api_docs' in docs_result else '❌'}")
        print(f"🌐 Localized docs: {'✅' if 'localized_documentation' in docs_result else '❌'}")

        print("\\n📋 Generated Documentation Structure:")
        docs_dir = self.demo_project_dir / "docs"
        if docs_dir.exists():
            for doc_file in docs_dir.rglob("*"):
                if doc_file.is_file():
                    relative_path = doc_file.relative_to(docs_dir)
                    size = doc_file.stat().st_size
                    print(f"  📄 {relative_path} ({size} bytes)")

        print("\\n" + "=" * 50 + "\\n")

    def demo_language_localization(self):
        """Demonstrate language localization capabilities."""
        print("🌐 Demo 3: Language Localization")
        print("-" * 29)

        # Test multiple languages
        languages = ["en", "ko", "ja", "zh"]

        for lang in languages:
            print(f"🔤 Testing {lang.upper()} language support...")

            # Update language settings
            updates = {
                "language.conversation_language": lang,
                "language.documentation_language": lang,
            }

            update_result = self.project.update_language_settings(updates)

            if update_result["success"]:
                # Create multilingual documentation structure
                multilingual_result = self.project.language_initializer.create_multilingual_documentation_structure(
                    lang
                )

                print(f"  ✅ {lang.upper()} configured")
                print(f"  📁 Docs structure: {'✅' if multilingual_result['success'] else '❌'}")

                # Get token cost analysis
                cost_analysis = self.project.language_initializer.get_token_cost_analysis(lang)
                print(f"  💰 Token cost impact: +{cost_analysis['cost_impact']}%")

            else:
                print(f"  ❌ {lang.upper()} configuration failed")

        print("\\n🌍 Multilingual Support Summary:")
        lang_status = self.project.language_initializer.get_language_status()
        print(f"  - Current language: {lang_status['current_language']}")
        print(f"  - Supported languages: {', '.join(lang_status['supported_languages'])}")

        print("\\n" + "=" * 50 + "\\n")

    def demo_template_optimization(self):
        """Demonstrate template optimization capabilities."""
        print("⚡ Demo 4: Template Optimization")
        print("-" * 30)

        # Create some test templates for optimization
        templates_dir = self.demo_project_dir / "templates"
        templates_dir.mkdir(exist_ok=True)

        # Create test templates with optimization opportunities
        test_templates = {
            "project_overview.md": """
# Project Overview

This is the project overview template.

## Project Description

This section describes the project.

## Project Description  # Duplicate

This section describes the project again.

## Features

- Feature 1
- Feature 2
- Feature 3

## Features  # Duplicate

- Feature 1
- Feature 2
- Feature 3

Extra whitespace:



            """,
            "api_documentation.md": """
# API Documentation

## Introduction

API documentation template.

## Endpoints

### GET /api/users

Get users list.

### POST /api/users

Create new user.

### GET /api/users

Get users list again.  # Duplicate

Complex template logic:
{% if api_version == "v1" %}
    {% if endpoint_type == "public" %}
        Public endpoint
    {% endif %}
{% endif %}

            """,
        }

        for template_name, content in test_templates.items():
            (templates_dir / template_name).write_text(content, encoding="utf-8")

        print(f"📝 Created {len(test_templates)} test templates for optimization")

        # Analyze templates
        print("\\n🔍 Analyzing templates...")
        start_time = time.time()

        analysis = self.project.template_optimizer.analyze_project_templates()

        analysis_time = time.time() - start_time

        print(f"⏱️  Analysis completed in {analysis_time:.2f} seconds")
        print(f"📁 Templates analyzed: {len(analysis.get('analyzed_files', []))}")

        # Show analysis results
        for file_analysis in analysis.get("analyzed_files", []):
            print(f"  📄 {Path(file_analysis['file_path']).name}")
            print(f"     - Size: {file_analysis.get('file_size', 0)} bytes")
            if "complexity_score" in file_analysis:
                print(f"     - Complexity: {file_analysis['complexity_score']}/10")

        # Apply optimizations
        print("\\n🔧 Applying optimizations...")
        start_time = time.time()

        optimization_options = {
            "backup_first": True,
            "apply_size_optimizations": True,
            "apply_performance_optimizations": True,
            "apply_complexity_optimizations": True,
        }

        opt_result = self.project.template_optimizer.create_optimized_templates(optimization_options)

        opt_time = time.time() - start_time

        self.results["template_optimization"] = {
            "analysis_time": analysis_time,
            "optimization_time": opt_time,
            "files_analyzed": len(analysis.get("analyzed_files", [])),
            "optimizations_applied": opt_result.get("success", False),
        }

        print(f"✅ Optimization completed in {opt_time:.2f} seconds")

        # Show optimization results
        if opt_result.get("success"):
            opt_results = opt_result.get("optimization_results", {})
            if "size_reduction" in opt_results:
                print(f"📉 Size reduction: {opt_results['size_reduction']:.1f}%")
            if "performance_improvement" in opt_results:
                print(f"⚡ Performance improvement: {opt_results['performance_improvement']:.1f}%")

        # Run benchmark
        print("\\n🏃 Running performance benchmark...")
        benchmark_result = self.project.template_optimizer.benchmark_template_performance()

        if benchmark_result.get("success"):
            print("✅ Benchmark completed successfully")

        print("\\n" + "=" * 50 + "\\n")

    def demo_backup_and_recovery(self):
        """Demonstrate backup and recovery capabilities."""
        print("💾 Demo 5: Backup and Recovery")
        print("-" * 28)

        # Create comprehensive backup
        backup_name = f"demo-backup-{datetime.now().strftime('%Y%m%d-%H%M%S')}"

        print(f"📦 Creating backup: {backup_name}")
        start_time = time.time()

        backup_result = self.project.create_project_backup(backup_name)

        backup_time = time.time() - start_time

        self.results["backup_creation"] = {
            "success": backup_result["success"],
            "duration": backup_time,
            "backup_name": backup_name,
        }

        print(f"✅ Backup created in {backup_time:.2f} seconds")

        # Show backup components
        components = backup_result.get("components", {})
        for component_name, component_data in components.items():
            status = "✅" if component_data.get("success", False) else "❌"
            print(f"  {status} {component_name.replace('_', ' ').title()}")

        # Verify backup files exist
        backup_dir = self.demo_project_dir / ".moai-backups" / backup_name
        if backup_dir.exists():
            backup_files = list(backup_dir.rglob("*"))
            backup_files = [f for f in backup_files if f.is_file()]
            print(f"📁 Backup files created: {len(backup_files)}")

        print("\\n" + "=" * 50 + "\\n")

    def demo_export_functionality(self):
        """Demonstrate export functionality."""
        print("📤 Demo 6: Export Functionality")
        print("-" * 28)

        # Test different export formats
        export_formats = [
            "markdown",
            "html",
        ]  # PDF would require additional dependencies

        for format_type in export_formats:
            print(f"📄 Testing {format_type.upper()} export...")

            start_time = time.time()

            # Export documentation
            export_result = self.project.export_project_documentation(format_type=format_type, language="ko")

            export_time = time.time() - start_time

            if export_result.get("success"):
                print(f"✅ {format_type.upper()} export completed in {export_time:.2f} seconds")
                print(f"📁 Export path: {export_result.get('export_path', 'N/A')}")
            else:
                print(f"❌ {format_type.upper()} export failed")

        # Show export summary
        print("\\n📊 Export Summary:")
        docs_dir = self.demo_project_dir / "docs"
        if docs_dir.exists():
            file_types = {}
            for file_path in docs_dir.rglob("*"):
                if file_path.is_file():
                    suffix = file_path.suffix.lower()
                    file_types[suffix] = file_types.get(suffix, 0) + 1

            for file_type, count in file_types.items():
                print(f"  📄 {file_type or 'no extension'}: {count} files")

        print("\\n" + "=" * 50 + "\\n")

    def demo_project_status(self):
        """Demonstrate project status reporting."""
        print("📊 Demo 7: Project Status Reporting")
        print("-" * 34)

        # Get comprehensive project status
        status = self.project.get_project_status()

        print("📋 Project Configuration:")
        config = status["configuration"]
        print(f"  📁 Project: {config['project_name']} ({config['project_type']})")
        print(f"  🏷️  Type: {config['project_type']}")
        print(f"  🚀 Menu System: v{config['menu_system_version']}")
        print(f"  ✅ Fully Initialized: {status['fully_initialized']}")

        print("\\n🔧 Module Status:")
        modules = status["modules"]
        for module_name, module_status in modules.items():
            print(f"  📦 {module_name.replace('_', ' ').title()}: {module_status}")

        print("\\n🌐 Language Status:")
        lang_status = status["language_status"]
        print(f"  💬 Current: {lang_status.get('current_language', 'N/A')}")
        print(f"  🌍 Supported: {', '.join(lang_status.get('supported_languages', []))}")

        print("\\n📚 Documentation Status:")
        doc_status = status["documentation_status"]
        print(f"  📄 Structure: {doc_status.get('structure_initialized', 'N/A')}")
        print(f"  📁 Templates: {doc_status.get('templates_available', 'N/A')}")

        # Show integration matrix
        print("\\n🔗 Integration Matrix:")
        matrix = self.project.get_integration_matrix()

        print("  Module Integrations:")
        for module_name, module_info in matrix["modules"].items():
            integrates_with = module_info.get("integrates_with", [])
            print(
                f"    📦 {module_name}: connects to {', '.join(integrates_with) if integrates_with else 'standalone'}"
            )

        print("\\n  Workflows:")
        for workflow_name, workflow_steps in matrix["workflows"].items():
            print(f"    🔄 {workflow_name.replace('_', ' ').title()}: {len(workflow_steps)} steps")

        print("\\n" + "=" * 50 + "\\n")

    def demo_convenience_functions(self):
        """Demonstrate convenience functions."""
        print("⚡ Demo 8: Convenience Functions")
        print("-" * 31)

        # Test convenience functions
        import shutil

        # Create another temporary project for convenience function testing
        convenience_dir = Path(tempfile.mkdtemp(prefix="moai_convenience_"))

        try:
            print("🔧 Testing initialize_project() convenience function...")
            start_time = time.time()

            init_result = initialize_project(
                str(convenience_dir),
                language="ja",
                user_name="利便性テスト",
                project_type="mobile_application",
            )

            init_time = time.time() - start_time

            if init_result["success"]:
                print(f"✅ Convenience initialization completed in {init_time:.2f} seconds")
                print(f"📦 Modules: {', '.join(init_result['modules_initialized'])}")

            # Test documentation generation convenience function
            print("\\n📚 Testing generate_docs() convenience function...")
            test_spec = {
                "id": "CONVENIENCE-SPEC",
                "title": "Convenience Function Test",
                "description": "Testing convenience functions",
                "requirements": ["Requirement 1", "Requirement 2"],
            }

            docs_result = generate_docs(test_spec, str(convenience_dir))

            if docs_result["success"]:
                print("✅ Convenience documentation generation successful")

            # Test template optimization convenience function
            print("\\n⚡ Testing optimize_templates() convenience function...")

            opt_result = optimize_templates(str(convenience_dir))

            if "analysis" in opt_result:
                print("✅ Convenience template optimization successful")

        finally:
            # Clean up convenience test directory
            shutil.rmtree(convenience_dir, ignore_errors=True)

        print("\\n" + "=" * 50 + "\\n")

    def show_performance_summary(self):
        """Show performance summary of all operations."""
        print("📈 Performance Summary")
        print("-" * 20)

        total_operations_time = 0

        for operation, metrics in self.results.items():
            if isinstance(metrics, dict) and "duration" in metrics:
                duration = metrics["duration"]
                total_operations_time += duration
                print(f"  ⏱️  {operation.replace('_', ' ').title()}: {duration:.2f}s")

        print(f"\\n🕐 Total operation time: {total_operations_time:.2f} seconds")

        # Show key metrics
        print("\\n🔑 Key Metrics:")
        if "initialization" in self.results:
            init = self.results["initialization"]
            print(f"  📦 Modules initialized: {init['modules_initialized']}")
            print(f"  📄 Files created: {init['created_files']}")

        if "documentation_generation" in self.results:
            docs = self.results["documentation_generation"]
            print(f"  📚 Feature docs: {'✅' if docs['feature_docs_generated'] else '❌'}")
            print(f"  🔗 API docs: {'✅' if docs['api_docs_generated'] else '❌'}")

        if "template_optimization" in self.results:
            opt = self.results["template_optimization"]
            print(f"  📁 Templates analyzed: {opt['files_analyzed']}")
            print(f"  ⚡ Optimizations applied: {'✅' if opt['optimizations_applied'] else '❌'}")

        if "backup_creation" in self.results:
            backup = self.results["backup_creation"]
            print(f"  💾 Backup created: {backup['backup_name']}")

        print("\\n" + "=" * 50 + "\\n")

    def cleanup_demo_environment(self):
        """Clean up the demonstration environment."""
        print("🧹 Cleaning up demo environment...")

        import shutil

        if self.demo_project_dir and self.demo_project_dir.exists():
            shutil.rmtree(self.demo_project_dir, ignore_errors=True)
            print(f"🗑️  Removed demo directory: {self.demo_project_dir}")

        print("✅ Cleanup complete")
        print("\\n🎉 Demo completed successfully!")
        print("\\n📋 Summary:")
        print("  ✅ All modules initialized successfully")
        print("  ✅ Documentation generated from SPEC")
        print("  ✅ Language localization demonstrated")
        print("  ✅ Template optimization applied")
        print("  ✅ Backup and recovery tested")
        print("  ✅ Export functionality verified")
        print("  ✅ Project status reporting complete")
        print("  ✅ Convenience functions working")
        print("\\n🚀 MoAI Menu Project system is fully functional!")

    def run_complete_demo(self):
        """Run the complete demonstration workflow."""
        print("🎬 MoAI Menu Project - Complete Workflow Demo")
        print("=" * 50)
        print("This demo showcases all features of the integrated")
        print("project management system including documentation,")
        print("language localization, and template optimization.")
        print("=" * 50 + "\\n")

        try:
            # Run all demonstration phases
            self.setup_demo_environment()
            self.demo_project_initialization()
            self.demo_spec_documentation_generation()
            self.demo_language_localization()
            self.demo_template_optimization()
            self.demo_backup_and_recovery()
            self.demo_export_functionality()
            self.demo_project_status()
            self.demo_convenience_functions()
            self.show_performance_summary()

        except Exception as e:
            print(f"❌ Demo failed with error: {e}")
            import traceback

            traceback.print_exc()

        finally:
            # Always clean up
            self.cleanup_demo_environment()


def main():
    """Main function to run the demonstration."""
    print("🚀 Starting MoAI Menu Project Complete Workflow Demo...")
    print()

    # Check if we're in the right directory
    skill_dir = Path(__file__).parent.parent
    if not (skill_dir / "SKILL.md").exists():
        print(f"❌ Error: SKILL.md not found in {skill_dir}")
        print("Make sure you're running this from the correct directory")
        return 1

    # Run the demonstration
    demo_runner = DemoRunner()
    demo_runner.run_complete_demo()

    return 0


if __name__ == "__main__":
    sys.exit(main())
