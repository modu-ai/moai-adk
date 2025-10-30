"""
Test suite for SPEC-NEXTRA-SITE-001: Nextra 4.0 + Next.js 14 Basic Structure
@TEST:NEXTRA-SITE-001

This test suite follows the TDD approach with RED-GREEN-REFACTOR cycles.
Each TAG is tested independently to ensure proper implementation.
"""

import os
import json
import pytest
from pathlib import Path


# Test configuration
PROJECT_ROOT = Path(__file__).parent.parent
DOCS_SITE_DIR = PROJECT_ROOT / "docs-site"


class TestNextraInitialization:
    """
    @TEST:NEXTRA-INIT-001
    Test Next.js 14 project initialization
    """

    def test_docs_site_directory_exists(self):
        """Verify that docs-site directory is created"""
        assert DOCS_SITE_DIR.exists(), "docs-site directory should exist"
        assert DOCS_SITE_DIR.is_dir(), "docs-site should be a directory"

    def test_package_json_exists(self):
        """Verify that package.json exists in docs-site"""
        package_json_path = DOCS_SITE_DIR / "package.json"
        assert package_json_path.exists(), "package.json should exist"

    def test_package_json_contains_dependencies(self):
        """Verify that package.json contains required dependencies"""
        package_json_path = DOCS_SITE_DIR / "package.json"

        with open(package_json_path, 'r', encoding='utf-8') as f:
            package_data = json.load(f)

        # Check required dependencies
        dependencies = package_data.get('dependencies', {})
        assert 'next' in dependencies, "next should be in dependencies"
        assert 'nextra' in dependencies, "nextra should be in dependencies"
        assert 'nextra-theme-docs' in dependencies, "nextra-theme-docs should be in dependencies"
        assert 'react' in dependencies, "react should be in dependencies"
        assert 'react-dom' in dependencies, "react-dom should be in dependencies"

        # Verify version constraints
        assert dependencies['next'].startswith('^14'), "next version should be ^14.x.x"
        assert dependencies['nextra'].startswith('^4'), "nextra version should be ^4.x.x"
        assert dependencies['react'].startswith('^18'), "react version should be ^18.x.x"

    def test_package_json_contains_scripts(self):
        """Verify that package.json contains required scripts"""
        package_json_path = DOCS_SITE_DIR / "package.json"

        with open(package_json_path, 'r', encoding='utf-8') as f:
            package_data = json.load(f)

        scripts = package_data.get('scripts', {})
        assert 'dev' in scripts, "dev script should exist"
        assert 'build' in scripts, "build script should exist"
        assert 'start' in scripts, "start script should exist"
        assert scripts['dev'] == 'next dev', "dev script should be 'next dev'"
        assert scripts['build'] == 'next build', "build script should be 'next build'"

    def test_tsconfig_json_exists(self):
        """Verify that tsconfig.json exists"""
        tsconfig_path = DOCS_SITE_DIR / "tsconfig.json"
        assert tsconfig_path.exists(), "tsconfig.json should exist"

    def test_gitignore_exists(self):
        """Verify that .gitignore exists"""
        gitignore_path = DOCS_SITE_DIR / ".gitignore"
        assert gitignore_path.exists(), ".gitignore should exist"


class TestNextraConfiguration:
    """
    @TEST:NEXTRA-CONFIG-001
    Test Nextra 4.0 basic configuration
    """

    def test_next_config_exists(self):
        """Verify that next.config.js exists"""
        next_config_path = DOCS_SITE_DIR / "next.config.js"
        assert next_config_path.exists(), "next.config.js should exist"

    def test_next_config_contains_nextra_plugin(self):
        """Verify that next.config.js contains Nextra plugin configuration"""
        next_config_path = DOCS_SITE_DIR / "next.config.js"

        with open(next_config_path, 'r', encoding='utf-8') as f:
            content = f.read()

        # Check for Nextra plugin import
        assert 'nextra' in content.lower(), "next.config.js should import nextra"
        assert 'nextra-theme-docs' in content, "next.config.js should reference nextra-theme-docs"

    def test_next_config_has_export_output(self):
        """Verify that next.config.js has 'output: export' for static site generation"""
        next_config_path = DOCS_SITE_DIR / "next.config.js"

        with open(next_config_path, 'r', encoding='utf-8') as f:
            content = f.read()

        # Check for static export configuration
        assert "output:" in content, "next.config.js should have output configuration"
        assert "'export'" in content or '"export"' in content, "output should be set to 'export'"

    def test_next_config_has_images_unoptimized(self):
        """Verify that next.config.js has images.unoptimized for static export"""
        next_config_path = DOCS_SITE_DIR / "next.config.js"

        with open(next_config_path, 'r', encoding='utf-8') as f:
            content = f.read()

        # Check for image optimization disabled (required for static export)
        assert "images:" in content, "next.config.js should have images configuration"
        assert "unoptimized" in content, "images should have unoptimized option"


class TestNextraTheme:
    """
    @TEST:NEXTRA-THEME-001
    Test Nextra theme configuration
    """

    def test_theme_config_exists(self):
        """Verify that theme.config.tsx exists"""
        theme_config_path = DOCS_SITE_DIR / "theme.config.tsx"
        assert theme_config_path.exists(), "theme.config.tsx should exist"

    def test_theme_config_has_logo(self):
        """Verify that theme.config.tsx has logo configuration"""
        theme_config_path = DOCS_SITE_DIR / "theme.config.tsx"

        with open(theme_config_path, 'r', encoding='utf-8') as f:
            content = f.read()

        assert "logo:" in content or "logo" in content, "theme.config.tsx should have logo"


class TestNextraHomepage:
    """
    @TEST:NEXTRA-HOMEPAGE-001
    Test homepage content
    """

    def test_pages_directory_exists(self):
        """Verify that pages directory exists"""
        pages_dir = DOCS_SITE_DIR / "pages"
        assert pages_dir.exists(), "pages directory should exist"

    def test_index_mdx_exists(self):
        """Verify that pages/index.mdx exists"""
        index_mdx = DOCS_SITE_DIR / "pages" / "index.mdx"
        assert index_mdx.exists(), "pages/index.mdx should exist"


class TestNextraBuild:
    """
    @TEST:NEXTRA-BUILD-001
    Test build and deployment configuration
    """

    def test_vercel_json_exists(self):
        """Verify that vercel.json exists"""
        vercel_json = DOCS_SITE_DIR / "vercel.json"
        assert vercel_json.exists(), "vercel.json should exist"

    def test_vercel_json_has_framework(self):
        """Verify that vercel.json specifies nextjs framework"""
        vercel_json_path = DOCS_SITE_DIR / "vercel.json"

        with open(vercel_json_path, 'r', encoding='utf-8') as f:
            vercel_config = json.load(f)

        assert 'framework' in vercel_config, "vercel.json should specify framework"
        assert vercel_config['framework'] == 'nextjs', "framework should be 'nextjs'"

    def test_vercel_json_has_security_headers(self):
        """Verify that vercel.json includes security headers"""
        vercel_json_path = DOCS_SITE_DIR / "vercel.json"

        with open(vercel_json_path, 'r', encoding='utf-8') as f:
            vercel_config = json.load(f)

        assert 'headers' in vercel_config, "vercel.json should have headers configuration"
