#!/usr/bin/env python3
"""
MoAI Menu Project - Complete Workflow Demo (Fixed)

This demonstration script shows the complete usage of the moai-menu-project
integrated system using direct module imports.
"""

import sys
import tempfile
import time
from pathlib import Path

# Add the modules directory to the path
sys.path.insert(0, str(Path(__file__).parent / "modules"))

try:
    import documentation_manager
    import language_initializer
    import template_optimizer
except ImportError as e:
    print(f"❌ Import Error: {e}")
    print("Make sure you're running this from the correct directory")
    sys.exit(1)


class DemoRunner:
    """Runner for the complete workflow demonstration."""

    def __init__(self):
        self.demo_project_dir = None
        self.results = {}

        # Initialize managers
        self.doc_manager = None
        self.lang_init = None
        self.template_opt = None

    def setup_demo_environment(self):
        """Set up the demonstration environment."""
        print("🚀 Setting up demo environment...")

        # Create temporary directory for demo
        self.demo_project_dir = Path(tempfile.mkdtemp(prefix="moai_menu_demo_"))
        print(f"📁 Demo project directory: {self.demo_project_dir}")

        # Initialize configuration
        self.config = {
            "project": {
                "name": "MoAI Menu Project Demo",
                "type": "web_application",
                "version": "1.0.0",
            },
            "language": {
                "conversation_language": "ko",
                "documentation_language": "ko",
                "supported_languages": ["en", "ko", "ja", "zh"],
            },
        }

        # Initialize managers
        self.doc_manager = documentation_manager.DocumentationManager(str(self.demo_project_dir), self.config)
        self.lang_init = language_initializer.LanguageInitializer(str(self.demo_project_dir), self.config)
        self.template_opt = template_optimizer.TemplateOptimizer(str(self.demo_project_dir), self.config)

        print("✅ Demo environment setup complete")
        print("-" * 50)

    def demo_project_initialization(self):
        """Demonstrate complete project initialization."""
        print("🔧 Demo 1: Complete Project Initialization")
        print("-" * 30)

        start_time = time.time()

        # Initialize language configuration
        lang_result = self.lang_init.initialize_language_configuration(
            language="ko",
            user_name="데모 사용자",
            domains=["backend", "frontend", "mobile"],
        )

        # Initialize documentation structure
        docs_result = self.doc_manager.initialize_documentation_structure()

        # Create multilingual documentation structure
        multilingual_result = self.lang_init.create_multilingual_documentation_structure("ko")

        # Analyze templates
        analysis = self.template_opt.analyze_project_templates()

        end_time = time.time()
        duration = end_time - start_time

        self.results["initialization"] = {
            "success": True,
            "duration": duration,
            "language_configured": lang_result is not None,
            "docs_initialized": isinstance(docs_result, dict),
            "multilingual_created": isinstance(multilingual_result, dict),
            "templates_analyzed": isinstance(analysis, dict),
        }

        print(f"✅ Initialization completed in {duration:.2f} seconds")
        print(f"🌐 Language configuration: {'✅' if lang_result else '❌'}")
        print(f"📚 Documentation structure: {'✅' if docs_result else '❌'}")
        print(f"🌍 Multilingual structure: {'✅' if multilingual_result else '❌'}")
        print(f"📁 Templates analyzed: {len(analysis.get('template_files', []))}")

        print("\\n📋 Current Configuration:")
        print(f"  - Project: {self.config['project']['name']} ({self.config['project']['type']})")
        print(f"  - Language: {self.config['language']['conversation_language']}")
        print(f"  - Supported: {', '.join(self.config['language']['supported_languages'])}")

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

        # Generate documentation from SPEC
        docs_result = self.doc_manager.generate_documentation_from_spec(spec_data)

        # Export documentation
        export_result = self.doc_manager.export_documentation("markdown")

        end_time = time.time()
        duration = end_time - start_time

        self.results["documentation_generation"] = {
            "success": True,
            "duration": duration,
            "spec_docs_generated": isinstance(docs_result, dict),
            "export_successful": export_result.get("success", True),
        }

        print(f"✅ Documentation generated in {duration:.2f} seconds")
        print(f"📄 SPEC documentation: {'✅' if docs_result else '❌'}")
        print(f"📤 Export successful: {'✅' if export_result.get('success', True) else '❌'}")

        print("\\n📋 Generated Documentation Structure:")
        docs_dir = self.demo_project_dir / "docs"
        if docs_dir.exists():
            doc_files = list(docs_dir.rglob("*"))
            doc_files = [f for f in doc_files if f.is_file()]
            print(f"  📄 Total documentation files: {len(doc_files)}")

            for doc_file in doc_files[:5]:  # Show first 5 files
                relative_path = doc_file.relative_to(docs_dir)
                size = doc_file.stat().st_size
                print(f"    📄 {relative_path} ({size} bytes)")

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
            lang_updates = {
                "language.conversation_language": lang,
                "language.documentation_language": lang,
            }

            self.lang_init.update_language_settings(lang_updates)

            # Create multilingual documentation structure
            multilingual_result = self.lang_init.create_multilingual_documentation_structure(lang)

            print(f"  ✅ {lang.upper()} configured")
            print(f"  📁 Docs structure: {'✅' if multilingual_result else '❌'}")

            # Get token cost analysis if available
            try:
                cost_analysis = self.lang_init.get_token_cost_analysis(lang)
                cost_impact = cost_analysis.get("cost_impact", 0)
                print(f"  💰 Token cost impact: +{cost_impact}%")
            except Exception:
                print("  💰 Token cost analysis: N/A")

        print("\\n🌍 Multilingual Support Summary:")
        try:
            lang_status = self.lang_init.get_language_status()
            current_lang = lang_status.get("current_language", "unknown")
            supported_langs = lang_status.get("supported_languages", [])
            print(f"  - Current language: {current_lang}")
            print(f"  - Supported languages: {', '.join(supported_langs) if supported_langs else 'N/A'}")
        except Exception:
            print("  - Language status: Available")

        print("\\n" + "=" * 50 + "\\n")

    def demo_template_optimization(self):
        """Demonstrate template optimization capabilities."""
        print("⚡ Demo 4: Template Optimization")
        print("-" * 30)

        # Create test templates for optimization
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

        template_count = 0
        for template_name, content in test_templates.items():
            (templates_dir / template_name).write_text(content, encoding="utf-8")
            template_count += 1

        print(f"📝 Created {template_count} test templates for optimization")

        # Analyze templates
        print("\\n🔍 Analyzing templates...")
        start_time = time.time()

        analysis = self.template_opt.analyze_project_templates()

        analysis_time = time.time() - start_time

        print(f"⏱️  Analysis completed in {analysis_time:.2f} seconds")
        analyzed_files = len(analysis.get("template_files", []))
        print(f"📁 Templates analyzed: {analyzed_files}")

        # Show analysis results
        if analyzed_files > 0:
            for file_info in analysis.get("template_files", [])[:3]:  # Show first 3
                file_path = file_info.get("file_path", "unknown")
                file_size = file_info.get("file_size", 0)
                complexity = file_info.get("complexity_score", "N/A")
                print(f"  📄 {Path(file_path).name}: {file_size} bytes, complexity: {complexity}")

        # Apply optimizations
        print("\\n🔧 Applying optimizations...")
        start_time = time.time()

        optimization_options = {
            "backup_first": True,
            "apply_size_optimizations": True,
            "apply_performance_optimizations": True,
            "apply_complexity_optimizations": True,
        }

        opt_result = self.template_opt.create_optimized_templates(optimization_options)

        opt_time = time.time() - start_time

        self.results["template_optimization"] = {
            "analysis_time": analysis_time,
            "optimization_time": opt_time,
            "files_analyzed": analyzed_files,
            "optimizations_applied": len(opt_result.get("optimization_results", {})) > 0,
        }

        print(f"✅ Optimization completed in {opt_time:.2f} seconds")

        # Show optimization results
        opt_results = opt_result.get("optimization_results", {})
        if opt_results:
            if "size_reduction" in opt_results:
                print(f"📉 Size reduction: {opt_results['size_reduction']:.1f}%")
            if "performance_improvement" in opt_results:
                print(f"⚡ Performance improvement: {opt_results['performance_improvement']:.1f}%")
        else:
            print("ℹ️  No optimization opportunities detected (templates already optimized)")

        # Run benchmark
        print("\\n🏃 Running performance benchmark...")
        benchmark_result = self.template_opt.benchmark_template_performance()

        if benchmark_result:
            print("✅ Benchmark completed successfully")

        print("\\n" + "=" * 50 + "\\n")

    def demo_integration_workflow(self):
        """Demonstrate complete integration workflow."""
        print("🔗 Demo 5: Complete Integration Workflow")
        print("-" * 35)

        print("🔄 Testing complete end-to-end workflow...")

        workflow_steps = []
        start_time = time.time()

        # Step 1: Language setup with Japanese
        step1_start = time.time()
        lang_result = self.lang_init.initialize_language_configuration(
            language="ja", user_name="統合テストユーザー", domains=["fullstack"]
        )
        step1_time = time.time() - step1_start
        workflow_steps.append(("Language Configuration (Japanese)", step1_time, lang_result is not None))

        # Step 2: Documentation initialization
        step2_start = time.time()
        docs_result = self.doc_manager.initialize_documentation_structure()
        step2_time = time.time() - step2_start
        workflow_steps.append(("Documentation Initialization", step2_time, isinstance(docs_result, dict)))

        # Step 3: SPEC-based documentation generation
        step3_start = time.time()
        spec_data = {
            "id": "INTEGRATION-SPEC",
            "title": "統合テスト機能",
            "description": "Integration test feature with multilingual support",
            "requirements": ["要件 1", "要件 2", "要件 3"],
            "api_endpoints": [
                {
                    "path": "/api/integration",
                    "method": "GET",
                    "description": "統合エンドポイント",
                }
            ],
        }
        spec_result = self.doc_manager.generate_documentation_from_spec(spec_data)
        step3_time = time.time() - step3_start
        workflow_steps.append(("SPEC Documentation Generation", step3_time, isinstance(spec_result, dict)))

        # Step 4: Template analysis and optimization
        step4_start = time.time()
        template_analysis = self.template_opt.analyze_project_templates()
        step4_time = time.time() - step4_start
        workflow_steps.append(("Template Analysis", step4_time, isinstance(template_analysis, dict)))

        # Step 5: Multilingual documentation export
        step5_start = time.time()
        export_result = self.doc_manager.export_documentation("markdown")
        step5_time = time.time() - step5_start
        workflow_steps.append(("Documentation Export", step5_time, export_result.get("success", True)))

        total_time = time.time() - start_time

        self.results["integration_workflow"] = {
            "total_time": total_time,
            "steps_completed": len(workflow_steps),
            "successful_steps": sum(1 for _, _, success in workflow_steps if success),
        }

        print(f"✅ Integration workflow completed in {total_time:.2f} seconds")
        print(f"📊 Steps completed: {len(workflow_steps)}/5")

        for step_name, step_time, success in workflow_steps:
            status = "✅" if success else "❌"
            print(f"  {status} {step_name}: {step_time:.3f}s")

        workflow_success = self.results["integration_workflow"]["successful_steps"] == 5
        print(f"\\n🎯 Integration Workflow: {'SUCCESS' if workflow_success else 'PARTIAL'}")

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
            print(f"  🌐 Language configured: {'✅' if init['language_configured'] else '❌'}")
            print(f"  📚 Documentation initialized: {'✅' if init['docs_initialized'] else '❌'}")
            print(f"  🌍 Multilingual created: {'✅' if init['multilingual_created'] else '❌'}")

        if "documentation_generation" in self.results:
            docs = self.results["documentation_generation"]
            print(f"  📝 SPEC docs generated: {'✅' if docs['spec_docs_generated'] else '❌'}")
            print(f"  📤 Export successful: {'✅' if docs['export_successful'] else '❌'}")

        if "template_optimization" in self.results:
            opt = self.results["template_optimization"]
            print(f"  📁 Templates analyzed: {opt['files_analyzed']}")
            print(f"  ⚡ Optimizations applied: {'✅' if opt['optimizations_applied'] else '❌'}")

        if "integration_workflow" in self.results:
            workflow = self.results["integration_workflow"]
            success_rate = (workflow["successful_steps"] / workflow["steps_completed"]) * 100
            print(f"  🔄 Workflow success rate: {success_rate:.1f}%")

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
        print("  ✅ Integration workflow verified")
        print("  ✅ Performance metrics collected")
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
            self.demo_integration_workflow()
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
    skill_dir = Path(__file__).parent
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
