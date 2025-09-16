"""
🗿 MoAI-ADK Validation Utilities

Provides validation functions for environment, configuration, and project setup.
"""

import os
import sys
import subprocess
from pathlib import Path
from typing import Dict, List, Optional, Tuple
from colorama import Fore, Style

# Note: global_installer removed in favor of package-based resources


def validate_python_version(min_version: Tuple[int, int] = (3, 8)) -> bool:
    """
    Validate Python version meets minimum requirements.
    
    Args:
        min_version: Minimum required Python version as tuple
        
    Returns:
        True if version is valid, False otherwise
    """
    current = sys.version_info[:2]
    if current < min_version:
        print(f"{Fore.RED}❌ Python {min_version[0]}.{min_version[1]}+ required, "
              f"found {current[0]}.{current[1]}{Style.RESET_ALL}")
        return False
    return True


def validate_claude_code() -> bool:
    """
    Validate that Claude Code is available and properly configured.
    
    Returns:
        True if Claude Code is available, False otherwise
    """
    try:
        result = subprocess.run(
            ["claude", "--version"], 
            capture_output=True, 
            text=True, 
            timeout=10
        )
        if result.returncode == 0:
            return True
        else:
            print(f"{Fore.YELLOW}⚠️  Claude Code not found or not properly configured{Style.RESET_ALL}")
            return False
    except (subprocess.TimeoutExpired, FileNotFoundError):
        print(f"{Fore.YELLOW}⚠️  Claude Code not found in PATH{Style.RESET_ALL}")
        return False


def validate_git_repository(path: Path) -> bool:
    """
    Validate that the path is within a git repository.
    
    Args:
        path: Path to validate
        
    Returns:
        True if valid git repository, False otherwise
    """
    current_path = path.resolve()
    while current_path != current_path.parent:
        if (current_path / ".git").exists():
            return True
        current_path = current_path.parent
    return False


def validate_project_structure(project_path: Path) -> Dict[str, bool]:
    """
    Validate MoAI-ADK project structure.
    
    Args:
        project_path: Path to the project
        
    Returns:
        Dictionary of validation results
    """
    results = {}
    
    # Check for .claude directory
    claude_dir = project_path / ".claude"
    results["claude_config"] = claude_dir.exists()
    
    # Check for settings.json
    settings_file = claude_dir / "settings.json"
    results["settings_file"] = settings_file.exists()
    
    # Check for hooks directory
    hooks_dir = claude_dir / "hooks"
    results["hooks_directory"] = hooks_dir.exists()
    
    # Check for moai hooks
    moai_hooks_dir = hooks_dir / "moai"
    results["moai_hooks"] = moai_hooks_dir.exists()
    
    # Check for essential hook files
    if moai_hooks_dir.exists():
        results["session_start_hook"] = (moai_hooks_dir / "session_start_notice.py").exists()
        results["constitution_guard_hook"] = (moai_hooks_dir / "constitution_guard.py").exists()
        results["policy_block_hook"] = (moai_hooks_dir / "policy_block.py").exists()
        results["tag_validator_hook"] = (moai_hooks_dir / "tag_validator.py").exists()
        results["post_stage_guard_hook"] = (moai_hooks_dir / "post_stage_guard.py").exists()
    else:
        results["session_start_hook"] = False
        results["constitution_guard_hook"] = False
        results["policy_block_hook"] = False
        results["tag_validator_hook"] = False
        results["post_stage_guard_hook"] = False
    
    return results


def validate_environment() -> bool:
    """
    Comprehensive environment validation.

    Returns:
        True if environment is valid, False otherwise
    """
    all_valid = True

    print(f"{Fore.BLUE}🔍 Validating environment...{Style.RESET_ALL}")

    # Python version
    if not validate_python_version():
        all_valid = False
    else:
        version = sys.version_info
        print(f"{Fore.GREEN}✅ Python {version.major}.{version.minor}.{version.micro}{Style.RESET_ALL}")

    # Global resources validation
    if not validate_global_resources():
        all_valid = False
    else:
        print(f"{Fore.GREEN}✅ MoAI-ADK global resources available{Style.RESET_ALL}")

    # Claude Code availability
    if not validate_claude_code():
        print(f"{Fore.YELLOW}⚠️  Claude Code validation skipped (not required for installation){Style.RESET_ALL}")
    else:
        print(f"{Fore.GREEN}✅ Claude Code available{Style.RESET_ALL}")

    return all_valid


def validate_project_readiness(project_path: Path) -> bool:
    """
    Validate project is ready for MoAI-ADK integration.
    
    Args:
        project_path: Path to the project
        
    Returns:
        True if project is ready, False otherwise
    """
    if not project_path.exists():
        print(f"{Fore.RED}❌ Project path does not exist: {project_path}{Style.RESET_ALL}")
        return False
    
    if not project_path.is_dir():
        print(f"{Fore.RED}❌ Project path is not a directory: {project_path}{Style.RESET_ALL}")
        return False
    
    # Check if it's already a MoAI-ADK project
    claude_dir = project_path / ".claude"
    if claude_dir.exists():
        print(f"{Fore.YELLOW}⚠️  Claude Code configuration already exists{Style.RESET_ALL}")
        
        structure = validate_project_structure(project_path)
        if structure["moai_hooks"]:
            print(f"{Fore.YELLOW}⚠️  MoAI-ADK already initialized in this project{Style.RESET_ALL}")
            return False
    
    return True


def validate_moai_structure(project_path: Path) -> Dict[str, bool]:
    """
    Validate complete MoAI-ADK project structure.
    
    Args:
        project_path: Path to the project
        
    Returns:
        Dictionary of validation results for MoAI components
    """
    results = {}
    
    # Basic structure validation from existing function
    basic_results = validate_project_structure(project_path)
    results.update(basic_results)
    
    # MoAI-specific structure validation
    moai_dir = project_path / ".moai"
    results["moai_directory"] = moai_dir.exists()
    
    if moai_dir.exists():
        # Essential MoAI directories
        results["moai_steering"] = (moai_dir / "steering").exists()
        results["moai_specs"] = (moai_dir / "specs").exists()
        results["moai_memory"] = (moai_dir / "memory").exists()
        results["moai_templates"] = (moai_dir / "templates").exists()
        results["moai_indexes"] = (moai_dir / "indexes").exists()
        
        # Essential MoAI files
        results["moai_config"] = (moai_dir / "config.json").exists()
        results["constitution"] = (moai_dir / "memory" / "constitution.md").exists()
        
        # Steering documents
        steering_dir = moai_dir / "steering"
        if steering_dir.exists():
            results["product_md"] = (steering_dir / "product.md").exists()
            results["structure_md"] = (steering_dir / "structure.md").exists()
            results["tech_md"] = (steering_dir / "tech.md").exists()
        else:
            results["product_md"] = False
            results["structure_md"] = False
            results["tech_md"] = False
        
        # TAG system
        indexes_dir = moai_dir / "indexes"
        if indexes_dir.exists():
            results["tags_index"] = (indexes_dir / "tags.json").exists()
            results["traceability_index"] = (indexes_dir / "traceability.json").exists()
            results["state_index"] = (indexes_dir / "state.json").exists()
        else:
            results["tags_index"] = False
            results["traceability_index"] = False
            results["state_index"] = False
    else:
        # All MoAI components are missing
        for key in ["moai_steering", "moai_specs", "moai_memory", "moai_templates", 
                   "moai_indexes", "moai_config", "constitution", "product_md", 
                   "structure_md", "tech_md", "tags_index", "traceability_index", 
                   "state_index"]:
            results[key] = False
    
    return results


def validate_constitution_compliance(project_path: Path) -> Dict[str, Dict]:
    """
    Validate project compliance with MoAI Constitution 5 principles.
    
    Args:
        project_path: Path to the project
        
    Returns:
        Dictionary with compliance results for each principle
    """
    results = {
        "simplicity": {"compliant": True, "details": [], "score": 0},
        "architecture": {"compliant": True, "details": [], "score": 0},
        "testing": {"compliant": True, "details": [], "score": 0},
        "observability": {"compliant": True, "details": [], "score": 0},
        "versioning": {"compliant": True, "details": [], "score": 0}
    }
    
    # 1. Simplicity - check project complexity
    complexity_score = _calculate_project_complexity(project_path)
    results["simplicity"]["score"] = complexity_score
    if complexity_score > 3:
        results["simplicity"]["compliant"] = False
        results["simplicity"]["details"].append(f"Project complexity {complexity_score} exceeds maximum of 3")
    else:
        results["simplicity"]["details"].append(f"Project complexity {complexity_score} within limits")
    
    # 2. Architecture - check for modular structure
    modular_score = _check_architectural_modularity(project_path)
    results["architecture"]["score"] = modular_score
    results["architecture"]["details"].append(f"Modularity score: {modular_score}/100")
    
    # 3. Testing - check for test files and coverage setup
    testing_score = _check_testing_setup(project_path)
    results["testing"]["score"] = testing_score
    if testing_score < 60:
        results["testing"]["compliant"] = False
        results["testing"]["details"].append("Insufficient testing infrastructure")
    else:
        results["testing"]["details"].append("Testing infrastructure present")
    
    # 4. Observability - check for logging and monitoring
    observability_score = _check_observability_setup(project_path)
    results["observability"]["score"] = observability_score
    results["observability"]["details"].append(f"Observability score: {observability_score}/100")
    
    # 5. Versioning - check for proper versioning
    versioning_score = _check_versioning_setup(project_path)
    results["versioning"]["score"] = versioning_score
    if versioning_score < 80:
        results["versioning"]["compliant"] = False
        results["versioning"]["details"].append("Versioning not properly configured")
    else:
        results["versioning"]["details"].append("Proper versioning detected")
    
    return results


def _calculate_project_complexity(project_path: Path) -> int:
    """Calculate project complexity score (1-10)."""
    complexity = 0
    
    # Count different project types
    if (project_path / "package.json").exists():
        complexity += 1
    if (project_path / "pyproject.toml").exists():
        complexity += 1
    if (project_path / "Cargo.toml").exists():
        complexity += 1
    if (project_path / "go.mod").exists():
        complexity += 1
    if (project_path / "pom.xml").exists():
        complexity += 1
    
    # Count services (Dockerfiles)
    dockerfiles = list(project_path.rglob("Dockerfile"))
    complexity += len(dockerfiles)
    
    # Count microservices indicators
    if (project_path / "docker-compose.yml").exists():
        complexity += 1
    
    return min(complexity, 10)  # Cap at 10


def _check_architectural_modularity(project_path: Path) -> int:
    """Check architectural modularity (0-100)."""
    score = 50  # Base score
    
    # Check for src/ or lib/ directories (good modularity)
    if (project_path / "src").exists() or (project_path / "lib").exists():
        score += 20
    
    # Check for services/ or packages/ directories  
    if (project_path / "services").exists() or (project_path / "packages").exists():
        score += 15
    
    # Check for utils/ or shared/ directories
    if (project_path / "utils").exists() or (project_path / "shared").exists():
        score += 10
    
    # Penalty for monolithic indicators
    if len(list(project_path.rglob("*.py"))) > 100 and not (project_path / "src").exists():
        score -= 20
    
    return min(max(score, 0), 100)


def _check_testing_setup(project_path: Path) -> int:
    """Check testing infrastructure (0-100)."""
    score = 0
    
    # Check for test directories
    test_dirs = ["test", "tests", "__tests__", "spec"]
    for test_dir in test_dirs:
        if (project_path / test_dir).exists():
            score += 30
            break
    
    # Check for test configuration files
    test_configs = ["pytest.ini", "jest.config.js", "vitest.config.ts", "phpunit.xml"]
    for config in test_configs:
        if (project_path / config).exists():
            score += 20
            break
    
    # Check for test files
    test_files = list(project_path.rglob("*test*"))
    test_files.extend(list(project_path.rglob("*spec*")))
    if test_files:
        score += 30
    
    # Check for coverage configuration
    coverage_configs = [".coveragerc", "coverage.xml", "nyc.config.js"]
    for config in coverage_configs:
        if (project_path / config).exists():
            score += 20
            break
    
    return min(score, 100)


def _check_observability_setup(project_path: Path) -> int:
    """Check observability setup (0-100)."""
    score = 30  # Base score for basic project
    
    # Check for logging configuration
    if any((project_path / f).exists() for f in ["logging.conf", "log4j.properties", "winston.config.js"]):
        score += 25
    
    # Check for environment configuration
    if (project_path / ".env").exists() or (project_path / ".env.example").exists():
        score += 20
    
    # Check for monitoring/metrics
    if any(path.exists() for path in [
        project_path / "prometheus.yml",
        project_path / "grafana",
        project_path / "metrics"
    ]):
        score += 25
    
    return min(score, 100)


def _check_versioning_setup(project_path: Path) -> int:
    """Check versioning setup (0-100)."""
    score = 0
    
    # Check for version files
    version_indicators = [
        "package.json",
        "pyproject.toml", 
        "Cargo.toml",
        "VERSION",
        "_version.py"
    ]
    
    for indicator in version_indicators:
        if (project_path / indicator).exists():
            score += 40
            break
    
    # Check for git
    if validate_git_repository(project_path):
        score += 30
    
    # Check for release workflow
    github_dir = project_path / ".github" / "workflows"
    if github_dir.exists():
        release_files = list(github_dir.glob("*release*")) + list(github_dir.glob("*version*"))
        if release_files:
            score += 30
    
    return min(score, 100)


def run_full_validation(project_path: Path, verbose: bool = False) -> Dict[str, any]:
    """
    Run complete MoAI-ADK validation suite.
    
    Args:
        project_path: Path to validate
        verbose: Whether to print detailed results
        
    Returns:
        Complete validation results
    """
    if verbose:
        print(f"\n{Fore.CYAN}🗿 MoAI-ADK 프로젝트 검증 시작{Style.RESET_ALL}")
        print("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
    
    results = {
        "environment": validate_environment() if verbose else True,
        "project_readiness": validate_project_readiness(project_path),
        "moai_structure": validate_moai_structure(project_path),
        "constitution_compliance": validate_constitution_compliance(project_path)
    }
    
    if verbose:
        print(f"\n{Fore.BLUE}📊 검증 결과 요약:{Style.RESET_ALL}")
        
        # Structure results
        structure = results["moai_structure"]
        structure_score = sum(structure.values()) / len(structure) * 100
        print(f"  • MoAI 구조 완성도: {structure_score:.1f}%")
        
        # Constitution compliance
        constitution = results["constitution_compliance"]
        compliant_count = sum(1 for p in constitution.values() if p["compliant"])
        print(f"  • Constitution 준수: {compliant_count}/5 원칙")
        
        # Overall health
        if structure_score >= 80 and compliant_count >= 4:
            print(f"{Fore.GREEN}  ✅ 전체 상태: 우수{Style.RESET_ALL}")
        elif structure_score >= 60 and compliant_count >= 3:
            print(f"{Fore.YELLOW}  ⚠️  전체 상태: 양호 (개선 권장){Style.RESET_ALL}")
        else:
            print(f"{Fore.RED}  ❌ 전체 상태: 문제 (수정 필요){Style.RESET_ALL}")

    return results


def validate_global_resources() -> bool:
    """
    Validate global MoAI-ADK resources and install if needed.

    Returns:
        bool: True if global resources are available, False otherwise
    """
    try:
        print(f"{Fore.BLUE}🔍 Checking MoAI-ADK global resources...{Style.RESET_ALL}")

        # Note: Resources are now embedded in package - no separate installation needed
        print(f"{Fore.GREEN}✅ Resources are embedded in the package{Style.RESET_ALL}")
        return True

    except Exception as error:
        print(f"{Fore.RED}❌ Error checking global resources: {error}{Style.RESET_ALL}")
        return False