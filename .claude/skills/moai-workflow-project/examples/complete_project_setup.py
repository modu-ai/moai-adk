#!/usr/bin/env python3
"""
MoAI Menu Project - Complete Project Setup Example

This example demonstrates how to use the integrated MoAI Menu Project system
to set up a complete project with documentation, multilingual support,
and template optimization.
"""

import sys
from pathlib import Path

# Add the moai-menu-project to Python path
sys.path.insert(0, str(Path(__file__).parent.parent))

from moai_menu_project import MoaiMenuProject, initialize_project


def example_basic_setup():
    """Example: Basic project setup with default settings."""

    print("=== Basic Project Setup Example ===")

    # Initialize project with default settings
    result = initialize_project(
        project_root="./example-project",
        language="en",
        user_name="John Doe",
        domains=["backend", "frontend"],
        project_type="web_application",
    )

    print(f"Setup Success: {result['success']}")
    print(f"Modules Initialized: {', '.join(result['modules_initialized'])}")
    print(f"Created Files: {len(result['created_files'])}")

    if result.get("optimization_results"):
        opt_results = result["optimization_results"]
        print(f"Optimized Files: {len(opt_results.get('optimized_files', []))}")

    return result


def example_multilingual_setup():
    """Example: Multilingual project setup."""

    print("\n=== Multilingual Project Setup Example ===")

    # Initialize Korean project
    result = initialize_project(
        project_root="./example-ko-project",
        language="ko",
        user_name="김개발",
        domains=["backend", "frontend", "mobile"],
        project_type="web_application",
    )

    print(f"설정 성공: {result['success']}")
    print(f"언어: {result['configuration_updates']['language']['language']}")
    print(f"사용자 이름: {result['configuration_updates']['language']['updated_config']['user']['name']}")

    # Check token impact
    token_impact = result["configuration_updates"]["language"]["token_cost_impact"]
    print(f"토큰 오버헤드: {token_impact['total_overhead_percentage']}%")

    return result


def example_documentation_generation():
    """Example: Generate documentation from SPEC."""

    print("\n=== Documentation Generation Example ===")

    # Create project instance
    project = MoaiMenuProject("./example-project")

    # Sample SPEC data
    spec_data = {
        "id": "SPEC-001",
        "title": "User Authentication System",
        "description": "Implement secure user authentication with JWT tokens",
        "requirements": [
            "User registration with email verification",
            "Login with email and password",
            "JWT token generation and validation",
            "Password reset functionality",
            "Session management",
        ],
        "status": "Planned",
        "priority": "High",
        "api_endpoints": [
            {
                "path": "/api/auth/register",
                "method": "POST",
                "description": "Register new user",
                "parameters": [
                    {"name": "email", "type": "string", "required": True},
                    {"name": "password", "type": "string", "required": True},
                ],
            },
            {
                "path": "/api/auth/login",
                "method": "POST",
                "description": "User login",
                "parameters": [
                    {"name": "email", "type": "string", "required": True},
                    {"name": "password", "type": "string", "required": True},
                ],
            },
        ],
    }

    # Generate documentation
    docs_result = project.generate_documentation_from_spec(spec_data)

    print(f"Documentation generated for SPEC: {docs_result['spec_id']}")
    print(f"Updated files: {', '.join(docs_result['updated_files'])}")

    if docs_result.get("api_documentation"):
        api_docs = docs_result["api_documentation"]
        print(f"API endpoints documented: {len(api_docs.get('endpoints', []))}")

    return docs_result


def example_template_optimization():
    """Example: Advanced template optimization."""

    print("\n=== Template Optimization Example ===")

    # Create project instance
    project = MoaiMenuProject("./example-project")

    # Define optimization options
    optimization_options = {
        "backup_first": True,
        "apply_size_optimizations": True,
        "apply_performance_optimizations": True,
        "apply_complexity_optimizations": True,
        "preserve_functionality": True,
    }

    # Run optimization
    optimization_result = project.optimize_project_templates(optimization_options)

    analysis = optimization_result["analysis"]
    optimization = optimization_result["optimization"]

    print(f"Template files analyzed: {len(analysis['template_files'])}")
    print(f"Optimization opportunities: {len(analysis['optimization_opportunities'])}")

    if optimization["success"]:
        print(f"Files optimized: {len(optimization['optimized_files'])}")
        print(f"Size reduction: {optimization.get('size_reduction', 0):.1f}%")

        # Show applied optimizations
        applied_opts = optimization.get("optimizations_applied", [])
        unique_opts = list(set(applied_opts))
        print(f"Optimization types applied: {', '.join(unique_opts)}")
    else:
        print(f"Optimization failed: {optimization.get('errors', [])}")

    return optimization_result


def example_project_status():
    """Example: Get comprehensive project status."""

    print("\n=== Project Status Example ===")

    # Create project instance
    project = MoaiMenuProject("./example-project")

    # Get status
    status = project.get_project_status()

    print(f"Project: {status['configuration']['project_name']}")
    print(f"Type: {status['configuration']['project_type']}")
    print(f"Fully Initialized: {status['fully_initialized']}")
    print(f"Language: {status['language_status']['configured_language']}")

    # Documentation status
    docs_status = status["documentation_status"]
    print(f"Documentation files: {docs_status['total_files']}")
    if docs_status["missing_sections"]:
        print(f"Missing sections: {', '.join(docs_status['missing_sections'])}")

    # Integration matrix
    integration_matrix = project.get_integration_matrix()
    print(f"Available modules: {', '.join(integration_matrix['modules'].keys())}")

    return status


def example_export_documentation():
    """Example: Export documentation in different formats."""

    print("\n=== Documentation Export Example ===")

    # Create project instance
    project = MoaiMenuProject("./example-project")

    # Export in different formats
    formats = ["markdown", "html"]

    for format_type in formats:
        export_result = project.export_project_documentation(format_type)

        print(f"\n{format_type.upper()} Export:")
        print(f"Success: {export_result.get('success', False)}")
        print(f"Files exported: {len(export_result.get('files', []))}")
        print(f"Output directory: {export_result.get('output_directory', 'N/A')}")

        if not export_result.get("success"):
            print(f"Error: {export_result.get('error', 'Unknown error')}")

    return export_result


def example_multilingual_workflow():
    """Example: Complete multilingual workflow."""

    print("\n=== Multilingual Workflow Example ===")

    # Initialize Korean project
    project = MoaiMenuProject("./multilingual-example")

    # Step 1: Initialize with Korean language
    init_result = project.initialize_complete_project(
        language="ko",
        user_name="박디벨로퍼",
        domains=["backend", "frontend"],
        project_type="web_application",
    )

    print(f"Korean project initialized: {init_result['success']}")

    # Step 2: Update language settings
    lang_update = project.update_language_settings(
        {"language.agent_prompt_language": "english"}  # Use English for agent prompts to save tokens
    )

    print(f"Language settings updated: {lang_update['success']}")

    # Step 3: Generate localized documentation
    spec_data = {
        "id": "SPEC-KO-001",
        "title": "사용자 관리 시스템",
        "description": "사용자 등록, 인증, 프로필 관리 기능 구현",
        "requirements": [
            "이메일을 통한 사용자 등록",
            "소셜 로그인 지원",
            "프로필 관리 및 수정",
            "사용자 권한 관리",
        ],
        "status": "Planned",
        "priority": "High",
    }

    docs_result = project.generate_documentation_from_spec(spec_data)

    print(f"Korean documentation generated: {bool(docs_result)}")

    # Step 4: Export documentation
    export_result = project.export_project_documentation("markdown", "ko")

    print(f"Korean documentation exported: {export_result.get('success', False)}")

    return {
        "initialization": init_result,
        "language_update": lang_update,
        "documentation": docs_result,
        "export": export_result,
    }


def main():
    """Run all examples."""

    print("MoAI Menu Project - Integrated Module System Examples")
    print("=" * 60)

    examples = [
        ("Basic Setup", example_basic_setup),
        ("Multilingual Setup", example_multilingual_setup),
        ("Documentation Generation", example_documentation_generation),
        ("Template Optimization", example_template_optimization),
        ("Project Status", example_project_status),
        ("Documentation Export", example_export_documentation),
        ("Multilingual Workflow", example_multilingual_workflow),
    ]

    results = {}

    for name, example_func in examples:
        try:
            print(f"\n{'=' * 20} {name} {'=' * 20}")
            results[name] = example_func()
        except Exception as e:
            print(f"Error in {name}: {e}")
            results[name] = {"error": str(e)}

    print(f"\n{'=' * 20} Summary {'=' * 20}")

    for name, result in results.items():
        if "error" in result:
            print(f"❌ {name}: Failed - {result['error']}")
        else:
            success = result.get("success", True)
            print(f"{'✅' if success else '⚠️'} {name}: {'Success' if success else 'Partial'}")

    return results


if __name__ == "__main__":
    results = main()
    sys.exit(0 if all("error" not in result for result in results.values()) else 1)
