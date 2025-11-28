#!/usr/bin/env python3
"""
TDD Tests for Phase 1: Project Initialization
Tests for Nextra project structure, configuration, and theme setup
"""

import pytest
import json
import os
from pathlib import Path


class TestPhase1ProjectInitialization:
    """Test suite for Phase 1: Project Initialization"""

    def setup_method(self):
        """Setup test fixtures"""
        self.docs_dir = Path(__file__).parent
        self.project_root = self.docs_dir.parent

    # RED Phase Tests: These tests will FAIL until we implement the code

    def test_package_json_exists(self):
        """Test that package.json exists in docs directory"""
        package_json_path = self.docs_dir / "package.json"
        assert package_json_path.exists(), "package.json must exist in docs directory"

    def test_package_json_has_required_dependencies(self):
        """Test that package.json contains Next.js and Nextra dependencies"""
        package_json_path = self.docs_dir / "package.json"
        with open(package_json_path, 'r') as f:
            pkg = json.load(f)

        required_deps = ["next", "nextra", "nextra-theme-docs", "react", "react-dom"]
        dependencies = pkg.get("dependencies", {})

        for dep in required_deps:
            assert dep in dependencies, f"{dep} must be in dependencies"
            # Verify version is acceptable
            version = dependencies[dep]
            assert isinstance(version, str), f"{dep} version must be a string"

    def test_next_config_exists(self):
        """Test that next.config.js exists"""
        next_config_path = self.docs_dir / "next.config.js"
        assert next_config_path.exists(), "next.config.js must exist"

    def test_next_config_has_nextra_integration(self):
        """Test that next.config.js includes Nextra configuration"""
        next_config_path = self.docs_dir / "next.config.js"
        with open(next_config_path, 'r') as f:
            config_content = f.read()

        assert "withNextra" in config_content, "next.config.js must use withNextra plugin"
        assert "theme" in config_content, "next.config.js must include theme configuration"

    def test_theme_config_exists(self):
        """Test that theme.config.tsx exists"""
        theme_config_path = self.docs_dir / "theme.config.tsx"
        assert theme_config_path.exists(), "theme.config.tsx must exist"

    def test_theme_config_has_moai_branding(self):
        """Test that theme.config.tsx includes MoAI-ADK branding"""
        theme_config_path = self.docs_dir / "theme.config.tsx"
        with open(theme_config_path, 'r') as f:
            config_content = f.read()

        assert "MoAI" in config_content, "theme.config.tsx must include MoAI branding"
        assert "logo" in config_content, "theme.config.tsx must define logo"

    def test_pages_directory_structure(self):
        """Test that pages directory has required structure"""
        pages_dir = self.docs_dir / "pages"
        assert pages_dir.exists(), "pages directory must exist"

        required_dirs = ["getting-started", "core-concepts", "advanced", "worktree", "reference"]
        for dir_name in required_dirs:
            dir_path = pages_dir / dir_name
            assert dir_path.is_dir(), f"{dir_name} directory must exist in pages"

    def test_index_mdx_exists(self):
        """Test that index.mdx (homepage) exists"""
        index_path = self.docs_dir / "pages" / "index.mdx"
        assert index_path.exists(), "index.mdx (homepage) must exist"

    def test_styles_directory_exists(self):
        """Test that styles directory exists with globals.css"""
        styles_dir = self.docs_dir / "styles"
        assert styles_dir.exists(), "styles directory must exist"

        globals_css = styles_dir / "globals.css"
        assert globals_css.exists(), "globals.css must exist"

    def test_globals_css_has_grayscale_theme(self):
        """Test that globals.css includes grayscale theme variables"""
        globals_css = self.docs_dir / "styles" / "globals.css"
        with open(globals_css, 'r') as f:
            css_content = f.read()

        required_vars = [
            "--color-primary-fg",
            "--color-primary-bg",
            "--font-ko",
            "--font-code"
        ]
        for var in required_vars:
            assert var in css_content, f"CSS variable {var} must be defined"

    def test_globals_css_has_dark_mode_support(self):
        """Test that globals.css includes dark mode (data-theme='dark')"""
        globals_css = self.docs_dir / "styles" / "globals.css"
        with open(globals_css, 'r') as f:
            css_content = f.read()

        assert "data-theme=\"dark\"" in css_content or "[data-theme=\"dark\"]" in css_content, \
            "globals.css must include dark mode CSS support"

    def test_components_directory_exists(self):
        """Test that components directory exists"""
        components_dir = self.docs_dir / "components"
        assert components_dir.exists(), "components directory must exist"

    def test_public_directory_exists(self):
        """Test that public directory exists for static assets"""
        public_dir = self.docs_dir / "public"
        assert public_dir.exists(), "public directory must exist for static assets"

    def test_tsconfig_exists(self):
        """Test that tsconfig.json exists for TypeScript configuration"""
        tsconfig_path = self.docs_dir / "tsconfig.json"
        assert tsconfig_path.exists(), "tsconfig.json must exist"

    def test_tsconfig_strict_mode(self):
        """Test that tsconfig.json has strict mode enabled"""
        tsconfig_path = self.docs_dir / "tsconfig.json"
        with open(tsconfig_path, 'r') as f:
            tsconfig = json.load(f)

        compiler_options = tsconfig.get("compilerOptions", {})
        assert compiler_options.get("strict", False) is True, \
            "TypeScript strict mode must be enabled"

    def test_meta_json_exists_for_navigation(self):
        """Test that meta.json exists for navigation structure"""
        meta_json_path = self.docs_dir / "pages" / "_meta.json"
        assert meta_json_path.exists(), "_meta.json must exist for navigation"

    def test_getting_started_meta_exists(self):
        """Test that getting-started/_meta.js exists"""
        meta_path = self.docs_dir / "pages" / "getting-started" / "_meta.js"
        assert meta_path.exists(), "_meta.js must exist in getting-started directory"

    def test_eslint_config_exists(self):
        """Test that ESLint configuration exists"""
        eslint_paths = [
            self.docs_dir / ".eslintrc.json",
            self.docs_dir / ".eslintrc.js",
            self.docs_dir / "eslint.config.js"
        ]
        assert any(p.exists() for p in eslint_paths), \
            "ESLint configuration must exist"

    def test_prettier_config_exists(self):
        """Test that Prettier configuration exists"""
        prettier_paths = [
            self.docs_dir / ".prettierrc.json",
            self.docs_dir / ".prettierrc.js"
        ]
        assert any(p.exists() for p in prettier_paths), \
            "Prettier configuration must exist"

    def test_gitignore_exists(self):
        """Test that .gitignore exists"""
        gitignore_path = self.docs_dir / ".gitignore"
        assert gitignore_path.exists(), ".gitignore must exist"

    def test_gitignore_excludes_node_modules(self):
        """Test that .gitignore excludes node_modules and build artifacts"""
        gitignore_path = self.docs_dir / ".gitignore"
        with open(gitignore_path, 'r') as f:
            gitignore_content = f.read()

        required_ignores = ["node_modules", ".next", "dist", "out"]
        for ignore in required_ignores:
            assert ignore in gitignore_content, f".gitignore must exclude {ignore}"

    def test_dev_script_in_package_json(self):
        """Test that package.json has dev script"""
        package_json_path = self.docs_dir / "package.json"
        with open(package_json_path, 'r') as f:
            pkg = json.load(f)

        scripts = pkg.get("scripts", {})
        assert "dev" in scripts, "package.json must have 'dev' script"
        assert "npm run dev" in scripts["dev"] or "next dev" in scripts["dev"], \
            "dev script must run next development server"

    def test_build_script_in_package_json(self):
        """Test that package.json has build script"""
        package_json_path = self.docs_dir / "package.json"
        with open(package_json_path, 'r') as f:
            pkg = json.load(f)

        scripts = pkg.get("scripts", {})
        assert "build" in scripts, "package.json must have 'build' script"

    def test_start_script_in_package_json(self):
        """Test that package.json has start script"""
        package_json_path = self.docs_dir / "package.json"
        with open(package_json_path, 'r') as f:
            pkg = json.load(f)

        scripts = pkg.get("scripts", {})
        assert "start" in scripts, "package.json must have 'start' script"


class TestPhase1CSSTheme:
    """Test suite for CSS theme system"""

    def setup_method(self):
        """Setup test fixtures"""
        self.docs_dir = Path(__file__).parent
        self.globals_css_path = self.docs_dir / "styles" / "globals.css"

    def test_light_mode_colors_defined(self):
        """Test that light mode colors are properly defined"""
        with open(self.globals_css_path, 'r') as f:
            css_content = f.read()

        light_mode_colors = [
            "--color-primary-fg: #000000",
            "--color-primary-bg: #FFFFFF",
        ]
        for color in light_mode_colors:
            assert color in css_content, f"Light mode {color} must be defined"

    def test_dark_mode_colors_defined(self):
        """Test that dark mode colors are properly defined"""
        with open(self.globals_css_path, 'r') as f:
            css_content = f.read()

        assert "--color-primary-fg: #FFFFFF" in css_content or \
               "--color-primary-fg:#FFFFFF" in css_content, \
            "Dark mode foreground color must be defined"

    def test_korean_typography_styles(self):
        """Test that Korean typography styles are included"""
        with open(self.globals_css_path, 'r') as f:
            css_content = f.read()

        assert "font-family" in css_content, "Font family must be defined"
        assert "line-height" in css_content, "Line height for Korean text must be defined"

    def test_code_block_styles(self):
        """Test that code block styles are defined"""
        with open(self.globals_css_path, 'r') as f:
            css_content = f.read()

        assert "code" in css_content.lower() or "pre" in css_content.lower(), \
            "Code block styles must be defined"

    def test_font_imports_or_definitions(self):
        """Test that web fonts are defined or imported"""
        with open(self.globals_css_path, 'r') as f:
            css_content = f.read()

        # Should have font imports (Pretendard, Inter, or from system)
        assert ("@import" in css_content or "@font-face" in css_content or
                "font-family" in css_content), \
            "Font setup must be included"


class TestPhase1Integration:
    """Integration tests for Phase 1"""

    def test_docs_directory_structure_complete(self):
        """Test that complete docs directory structure is in place"""
        docs_dir = Path(__file__).parent

        # Required files
        required_files = [
            "package.json",
            "next.config.js",
            "theme.config.tsx",
            "tsconfig.json",
            "styles/globals.css"
        ]

        for file_path in required_files:
            full_path = docs_dir / file_path
            assert full_path.exists(), f"Required file {file_path} must exist"

    def test_pages_have_navigation_metadata(self):
        """Test that main page sections have navigation metadata"""
        docs_dir = Path(__file__).parent
        pages_dir = docs_dir / "pages"

        sections = ["getting-started", "core-concepts", "advanced", "worktree", "reference"]
        for section in sections:
            section_dir = pages_dir / section
            meta_file = section_dir / "_meta.js"
            assert meta_file.exists(), \
                f"Navigation metadata _meta.js must exist in {section}"


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
