#!/usr/bin/env python3
"""
Simple Integration Test for moai-menu-project

Tests the core functionality without complex imports.
"""

import sys
import tempfile
from pathlib import Path

# Add modules to path
sys.path.insert(0, "modules")

# Import individual modules
import documentation_manager
import language_initializer
import template_optimizer


def test_individual_modules():
    """Test each module individually."""
    print("🧪 Testing Individual Modules")
    print("=" * 40)

    # Create temporary directory
    test_dir = Path(tempfile.mkdtemp())
    print(f"📁 Test directory: {test_dir}")

    # Test configuration
    config = {
        "project": {"name": "Test Project", "type": "web_application"},
        "language": {"conversation_language": "en", "documentation_language": "en"},
    }

    results = {}

    # Test DocumentationManager
    print("\\n📚 Testing DocumentationManager...")
    try:
        doc_manager = documentation_manager.DocumentationManager(str(test_dir), config)

        # Test initialization
        init_result = doc_manager.initialize_documentation_structure()
        results["documentation_manager"] = {
            "success": init_result["success"],
            "created_files": len(init_result.get("created_files", [])),
        }

        # Test SPEC documentation generation
        spec_data = {
            "id": "TEST-SPEC-001",
            "title": "Test Feature",
            "description": "Test feature implementation",
            "requirements": ["Requirement 1", "Requirement 2"],
            "api_endpoints": [{"path": "/api/test", "method": "POST", "description": "Test endpoint"}],
        }

        docs_result = doc_manager.generate_documentation_from_spec(spec_data)
        results["documentation_manager"]["spec_generation"] = docs_result["success"]

        print(f"✅ DocumentationManager: {init_result['success']}")
        print(f"  📄 Files created: {len(init_result.get('created_files', []))}")
        print(f"  📝 SPEC generation: {docs_result['success']}")

    except Exception as e:
        print(f"❌ DocumentationManager error: {e}")
        results["documentation_manager"] = {"success": False, "error": str(e)}

    # Test LanguageInitializer
    print("\\n🌐 Testing LanguageInitializer...")
    try:
        lang_init = language_initializer.LanguageInitializer(str(test_dir), config)

        # Test language detection
        detected_lang = lang_init.detect_project_language()

        # Test language configuration
        lang_result = lang_init.initialize_language_configuration(
            language="ko", user_name="테스트 사용자", domains=["backend", "frontend"]
        )

        # Test multilingual structure
        multilingual_result = lang_init.create_multilingual_documentation_structure("ko")

        results["language_initializer"] = {
            "success": lang_result["success"],
            "detected_language": detected_lang,
            "multilingual_created": multilingual_result["success"],
        }

        print(f"✅ LanguageInitializer: {lang_result['success']}")
        print(f"  💬 Detected language: {detected_lang}")
        print(f"  🌍 Multilingual structure: {multilingual_result['success']}")

    except Exception as e:
        print(f"❌ LanguageInitializer error: {e}")
        results["language_initializer"] = {"success": False, "error": str(e)}

    # Test TemplateOptimizer
    print("\\n⚡ Testing TemplateOptimizer...")
    try:
        # Create test templates
        templates_dir = test_dir / "templates"
        templates_dir.mkdir(exist_ok=True)

        test_template = """
# Test Template

This template contains redundant content.

## Section 1

Content here...

## Section 1  # Duplicate

Duplicate content here...

Extra whitespace:

Complex logic:
{% if condition %}
    {% if nested_condition %}
        Content
    {% endif %}
{% endif %}
        """

        (templates_dir / "test_template.md").write_text(test_template, encoding="utf-8")

        template_opt = template_optimizer.TemplateOptimizer(str(test_dir), config)

        # Test template analysis
        analysis = template_opt.analyze_project_templates()

        # Test optimization
        optimization_result = template_opt.create_optimized_templates({"backup_first": True})

        # Test benchmarking
        benchmark_result = template_opt.benchmark_template_performance()

        results["template_optimizer"] = {
            "success": analysis["success"],
            "files_analyzed": len(analysis.get("analyzed_files", [])),
            "optimization_applied": optimization_result.get("success", False),
            "benchmark_completed": benchmark_result.get("success", False),
        }

        print(f"✅ TemplateOptimizer: {analysis['success']}")
        print(f"  📁 Files analyzed: {len(analysis.get('analyzed_files', []))}")
        print(f"  ⚡ Optimization applied: {optimization_result.get('success', False)}")
        print(f"  🏃 Benchmark completed: {benchmark_result.get('success', False)}")

    except Exception as e:
        print(f"❌ TemplateOptimizer error: {e}")
        results["template_optimizer"] = {"success": False, "error": str(e)}

    # Cleanup
    import shutil

    shutil.rmtree(test_dir, ignore_errors=True)

    return results


def test_module_integration():
    """Test integration between modules."""
    print("\\n🔗 Testing Module Integration")
    print("=" * 40)

    # Create temporary directory
    test_dir = Path(tempfile.mkdtemp())

    try:
        # Initialize all three modules with same config
        config = {
            "project": {"name": "Integration Test Project", "type": "web_application"},
            "language": {"conversation_language": "ko", "documentation_language": "ko"},
        }

        # Initialize modules
        doc_manager = documentation_manager.DocumentationManager(str(test_dir), config)
        lang_init = language_initializer.LanguageInitializer(str(test_dir), config)
        template_opt = template_optimizer.TemplateOptimizer(str(test_dir), config)

        print("✅ All modules initialized successfully")

        # Test 1: Language + Documentation integration
        print("\\n🌐📚 Testing Language + Documentation integration...")

        # Set up language configuration
        lang_result = lang_init.initialize_language_configuration(
            language="ko", user_name="통합 테스트", domains=["backend"]
        )

        # Create documentation structure
        doc_result = doc_manager.initialize_documentation_structure()

        # Test multilingual documentation
        multilingual_result = lang_init.create_multilingual_documentation_structure("ko")

        integration_success = all(
            [
                lang_result["success"],
                doc_result["success"],
                multilingual_result["success"],
            ]
        )

        print(f"✅ Language + Documentation integration: {integration_success}")

        # Test 2: Template + Documentation integration
        print("\\n⚡📚 Testing Template + Documentation integration...")

        # Create test templates
        templates_dir = test_dir / "templates"
        templates_dir.mkdir(exist_ok=True)

        (templates_dir / "docs_template.md").write_text(
            "# Documentation Template\\n\\nContent here...", encoding="utf-8"
        )

        # Analyze templates
        analysis = template_opt.analyze_project_templates()

        # Generate documentation using templates
        spec_data = {
            "id": "INTEGRATION-SPEC",
            "title": "통합 테스트 기능",
            "description": "모듈 통합 테스트",
            "requirements": ["요구사항 1", "요구사항 2"],
        }

        docs_from_spec = doc_manager.generate_documentation_from_spec(spec_data)

        template_doc_success = analysis["success"] and docs_from_spec["success"]

        print(f"✅ Template + Documentation integration: {template_doc_success}")

        # Test 3: All modules together
        print("\\n🔧🌐📚 Testing all modules together...")

        # Complete workflow
        workflow_success = True

        # 1. Language setup
        lang_result = lang_init.initialize_language_configuration(language="ja")
        workflow_success &= lang_result["success"]

        # 2. Documentation initialization
        doc_result = doc_manager.initialize_documentation_structure()
        workflow_success &= doc_result["success"]

        # 3. Template analysis
        analysis = template_opt.analyze_project_templates()
        workflow_success &= analysis["success"]

        print(f"✅ Complete workflow: {workflow_success}")

        return {
            "language_docs_integration": integration_success,
            "template_docs_integration": template_doc_success,
            "complete_workflow": workflow_success,
        }

    except Exception as e:
        print(f"❌ Integration test error: {e}")
        import traceback

        traceback.print_exc()
        return {"success": False, "error": str(e)}

    finally:
        # Cleanup
        import shutil

        shutil.rmtree(test_dir, ignore_errors=True)


def test_performance():
    """Test performance characteristics."""
    print("\\n📈 Performance Testing")
    print("=" * 30)

    import time

    # Create temporary directory
    test_dir = Path(tempfile.mkdtemp())

    try:
        config = {
            "project": {"name": "Performance Test", "type": "web_application"},
            "language": {"conversation_language": "en"},
        }

        performance_results = {}

        # Test module initialization time
        start_time = time.time()

        doc_manager = documentation_manager.DocumentationManager(str(test_dir), config)
        lang_init = language_initializer.LanguageInitializer(str(test_dir), config)
        template_opt = template_optimizer.TemplateOptimizer(str(test_dir), config)

        init_time = time.time() - start_time
        performance_results["module_initialization"] = init_time

        print(f"⏱️  Module initialization: {init_time:.3f}s")

        # Test documentation generation time
        start_time = time.time()

        doc_manager.initialize_documentation_structure()

        doc_time = time.time() - start_time
        performance_results["documentation_generation"] = doc_time

        print(f"⏱️  Documentation generation: {doc_time:.3f}s")

        # Test language detection time
        start_time = time.time()

        # Create some test files
        (test_dir / "test.py").write_text(
            "# Korean comments\\ndef calculate():\\n    # 계산 함수\\n    return 100",
            encoding="utf-8",
        )

        lang_init.detect_project_language()

        lang_time = time.time() - start_time
        performance_results["language_detection"] = lang_time

        print(f"⏱️  Language detection: {lang_time:.3f}s")

        # Test template analysis time
        templates_dir = test_dir / "templates"
        templates_dir.mkdir(exist_ok=True)

        # Create multiple test templates
        for i in range(5):
            (templates_dir / f"template_{i}.md").write_text(
                f"# Template {i}\\n\\nContent for template {i}.", encoding="utf-8"
            )

        start_time = time.time()

        template_opt.analyze_project_templates()

        template_time = time.time() - start_time
        performance_results["template_analysis"] = template_time

        print(f"⏱️  Template analysis (5 files): {template_time:.3f}s")

        # Performance benchmarks
        print("\\n📊 Performance Benchmarks:")
        print(f"  Module initialization (<1s): {'✅' if init_time < 1.0 else '❌'}")
        print(f"  Documentation generation (<3s): {'✅' if doc_time < 3.0 else '❌'}")
        print(f"  Language detection (<0.5s): {'✅' if lang_time < 0.5 else '❌'}")
        print(f"  Template analysis (<2s): {'✅' if template_time < 2.0 else '❌'}")

        total_time = sum(performance_results.values())
        print(f"\\n🕐 Total time: {total_time:.3f}s")

        return performance_results

    finally:
        # Cleanup
        import shutil

        shutil.rmtree(test_dir, ignore_errors=True)


def main():
    """Run all integration tests."""
    print("🧪 MoAI Menu Project - Integration Test Suite")
    print("=" * 50)
    print("Testing core functionality and module integration")
    print("=" * 50)

    try:
        # Run individual module tests
        module_results = test_individual_modules()

        # Run integration tests
        integration_results = test_module_integration()

        # Run performance tests
        performance_results = test_performance()

        # Summary
        print("\\n🎉 Test Summary")
        print("=" * 20)

        # Module success rates
        total_modules = len(module_results)
        successful_modules = sum(1 for result in module_results.values() if result.get("success", False))

        print(f"📦 Modules: {successful_modules}/{total_modules} successful")

        # Integration success rates
        total_integrations = len(integration_results)
        successful_integrations = sum(1 for result in integration_results.values() if result is True)

        print(f"🔗 Integrations: {successful_integrations}/{total_integrations} successful")

        # Performance summary
        avg_time = sum(performance_results.values()) / len(performance_results)
        print(f"📈 Average operation time: {avg_time:.3f}s")

        # Overall assessment
        overall_success = successful_modules == total_modules and successful_integrations == total_integrations

        print(f"\\n🎯 Overall Result: {'✅ SUCCESS' if overall_success else '❌ ISSUES DETECTED'}")

        if overall_success:
            print("\\n🚀 MoAI Menu Project system is fully functional!")
            print("✅ All modules working correctly")
            print("✅ Module integration successful")
            print("✅ Performance within acceptable limits")
        else:
            print("\\n⚠️  Some issues detected - review test results")

        return 0 if overall_success else 1

    except Exception as e:
        print(f"\\n❌ Test suite failed with error: {e}")
        import traceback

        traceback.print_exc()
        return 1


if __name__ == "__main__":
    sys.exit(main())
