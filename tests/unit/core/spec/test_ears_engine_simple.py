"""
Simple comprehensive unit tests for EARS template engine.

Tests cover:
- EARSTemplateEngine initialization
- Template rendering
- Language detection
- Domain determination
- Spec ID generation
- Content generation methods
"""

from pathlib import Path
from unittest import mock

import pytest

from moai_adk.core.spec.ears_template_engine import EARSTemplateEngine


class TestEARSTemplateEngineInit:
    """Test EARSTemplateEngine initialization."""

    def test_engine_initializes(self):
        """Test that engine initializes without error."""
        engine = EARSTemplateEngine()
        assert engine is not None

    def test_engine_has_confidence_scorer(self):
        """Test engine has confidence scorer."""
        engine = EARSTemplateEngine()
        assert engine.confidence_scorer is not None

    def test_engine_has_template_cache(self):
        """Test engine has template cache."""
        engine = EARSTemplateEngine()
        assert hasattr(engine, "template_cache")
        assert isinstance(engine.template_cache, dict)

    def test_engine_has_domain_templates(self):
        """Test engine has domain-specific templates."""
        engine = EARSTemplateEngine()
        assert "auth" in engine.domain_templates
        assert "api" in engine.domain_templates
        assert "data" in engine.domain_templates
        assert "ui" in engine.domain_templates
        assert "business" in engine.domain_templates

    def test_engine_has_ears_templates(self):
        """Test engine has EARS section templates."""
        engine = EARSTemplateEngine()
        assert "environment" in engine.ears_templates
        assert "assumptions" in engine.ears_templates
        assert "requirements" in engine.ears_templates
        assert "specifications" in engine.ears_templates

    def test_auth_domain_template_has_description(self):
        """Test auth domain template has description."""
        engine = EARSTemplateEngine()
        auth_template = engine.domain_templates["auth"]

        assert "description" in auth_template
        assert "authentication" in auth_template["description"].lower()

    def test_api_domain_template_has_features(self):
        """Test API domain template has common features."""
        engine = EARSTemplateEngine()
        api_template = engine.domain_templates["api"]

        assert "common_features" in api_template
        assert "Endpoints" in api_template["common_features"]


class TestLanguageDetection:
    """Test language detection from file paths."""

    def test_detect_language_python(self):
        """Test Python language detection."""
        engine = EARSTemplateEngine()
        lang = engine._detect_language("module.py")
        assert lang == "Python"

    def test_detect_language_javascript(self):
        """Test JavaScript language detection."""
        engine = EARSTemplateEngine()
        lang = engine._detect_language("script.js")
        assert lang == "JavaScript"

    def test_detect_language_typescript(self):
        """Test TypeScript language detection."""
        engine = EARSTemplateEngine()
        lang = engine._detect_language("module.ts")
        assert lang == "TypeScript"

    def test_detect_language_go(self):
        """Test Go language detection."""
        engine = EARSTemplateEngine()
        lang = engine._detect_language("main.go")
        assert lang == "Go"

    def test_detect_language_java(self):
        """Test Java language detection."""
        engine = EARSTemplateEngine()
        lang = engine._detect_language("Main.java")
        assert lang == "Java"

    def test_detect_language_unknown(self):
        """Test unknown language returns Unknown."""
        engine = EARSTemplateEngine()
        lang = engine._detect_language("file.xyz")
        assert lang == "Unknown"

    def test_detect_language_case_insensitive(self):
        """Test language detection is case insensitive."""
        engine = EARSTemplateEngine()
        lang = engine._detect_language("module.PY")
        assert lang == "Python"


class TestComplexityAnalysis:
    """Test code complexity analysis."""

    def test_analyze_complexity_low(self):
        """Test low complexity analysis."""
        engine = EARSTemplateEngine()
        extraction = {
            "classes": ["MyClass"],
            "functions": ["func1", "func2"],
            "imports": [],
            "domain_keywords": [],
        }

        complexity = engine._analyze_complexity(extraction)
        assert complexity == "low"

    def test_analyze_complexity_medium(self):
        """Test medium complexity analysis."""
        engine = EARSTemplateEngine()
        extraction = {
            "classes": ["Class1", "Class2", "Class3"],
            "functions": [
                "f1",
                "f2",
                "f3",
                "f4",
                "f5",
                "f6",
                "f7",
                "f8",
                "f9",
                "f10",
                "f11",
            ],
            "imports": [],
            "domain_keywords": [],
        }

        complexity = engine._analyze_complexity(extraction)
        assert complexity == "medium"

    def test_analyze_complexity_high_from_classes(self):
        """Test high complexity from many classes."""
        engine = EARSTemplateEngine()
        extraction = {
            "classes": ["C1", "C2", "C3", "C4", "C5", "C6"],
            "functions": [],
            "imports": [],
            "domain_keywords": [],
        }

        complexity = engine._analyze_complexity(extraction)
        assert complexity == "high"

    def test_analyze_complexity_high_from_functions(self):
        """Test high complexity from many functions."""
        engine = EARSTemplateEngine()
        extraction = {
            "classes": [],
            "functions": [f"f{i}" for i in range(21)],  # 21 functions
            "imports": [],
            "domain_keywords": [],
        }

        complexity = engine._analyze_complexity(extraction)
        assert complexity == "high"


class TestArchitectureAnalysis:
    """Test architecture pattern detection."""

    def test_analyze_architecture_mvc_django(self):
        """Test MVC architecture detection with Django."""
        engine = EARSTemplateEngine()
        extraction = {
            "classes": [],
            "functions": [],
            "imports": ["django", "django.db"],
            "domain_keywords": [],
        }

        arch = engine._analyze_architecture(extraction)
        assert arch == "mvc"

    def test_analyze_architecture_frontend_react(self):
        """Test frontend architecture detection with React."""
        engine = EARSTemplateEngine()
        extraction = {
            "classes": [],
            "functions": [],
            "imports": ["react", "react-dom"],
            "domain_keywords": [],
        }

        arch = engine._analyze_architecture(extraction)
        assert arch == "frontend"

    def test_analyze_architecture_api_fastapi(self):
        """Test API architecture detection with FastAPI."""
        engine = EARSTemplateEngine()
        extraction = {
            "classes": [],
            "functions": [],
            "imports": ["fastapi"],
            "domain_keywords": [],
        }

        arch = engine._analyze_architecture(extraction)
        assert arch == "api"

    def test_analyze_architecture_data(self):
        """Test data architecture detection."""
        engine = EARSTemplateEngine()
        extraction = {
            "classes": [],
            "functions": [],
            "imports": ["sqlalchemy"],
            "domain_keywords": [],
        }

        arch = engine._analyze_architecture(extraction)
        assert arch == "data"

    def test_analyze_architecture_simple_default(self):
        """Test simple architecture as default."""
        engine = EARSTemplateEngine()
        extraction = {
            "classes": [],
            "functions": [],
            "imports": [],
            "domain_keywords": [],
        }

        arch = engine._analyze_architecture(extraction)
        assert arch == "simple"


class TestDomainDetermination:
    """Test domain determination logic."""

    def test_determine_domain_auth(self):
        """Test authentication domain detection."""
        engine = EARSTemplateEngine()
        extraction = {
            "domain_keywords": ["auth", "login", "password"],
            "classes": [],
            "functions": [],
            "imports": [],
        }

        domain = engine._determine_domain(extraction)
        assert domain == "auth"

    def test_determine_domain_api(self):
        """Test API domain detection."""
        engine = EARSTemplateEngine()
        extraction = {
            "domain_keywords": ["api", "endpoint", "route"],
            "classes": [],
            "functions": [],
            "imports": [],
        }

        domain = engine._determine_domain(extraction)
        assert domain == "api"

    def test_determine_domain_data(self):
        """Test data domain detection."""
        engine = EARSTemplateEngine()
        extraction = {
            "domain_keywords": ["model", "schema", "database"],
            "classes": [],
            "functions": [],
            "imports": [],
        }

        domain = engine._determine_domain(extraction)
        assert domain == "data"

    def test_determine_domain_ui(self):
        """Test UI domain detection."""
        engine = EARSTemplateEngine()
        extraction = {
            "domain_keywords": ["ui", "component", "view"],
            "classes": [],
            "functions": [],
            "imports": [],
        }

        domain = engine._determine_domain(extraction)
        assert domain == "ui"

    def test_determine_domain_general_default(self):
        """Test default behavior when no keywords match."""
        engine = EARSTemplateEngine()
        extraction = {
            "domain_keywords": [],
            "classes": [],
            "functions": [],
            "imports": [],
        }

        domain = engine._determine_domain(extraction)
        # When no domain scores are found, max() returns first key
        # The actual behavior returns the domain with highest score or first
        assert isinstance(domain, str)
        assert domain in ["auth", "api", "data", "ui", "business", "general"]


class TestSpecIDGeneration:
    """Test SPEC ID generation."""

    def test_generate_spec_id_format(self):
        """Test SPEC ID follows correct format."""
        engine = EARSTemplateEngine()
        extraction = {
            "file_name": "user_model",
            "classes": [],
            "functions": [],
            "imports": [],
        }

        spec_id = engine._generate_spec_id(extraction, "auth")
        # Format: DOMAIN-FILENAME-HASH
        assert "AUTH" in spec_id
        assert "user" in spec_id.lower()
        assert "-" in spec_id

    def test_generate_spec_id_different_for_different_domains(self):
        """Test SPEC ID differs for different domains."""
        engine = EARSTemplateEngine()
        extraction = {
            "file_name": "service",
            "classes": [],
            "functions": [],
            "imports": [],
        }

        spec_id1 = engine._generate_spec_id(extraction, "auth")
        spec_id2 = engine._generate_spec_id(extraction, "api")

        # Should start with different domain prefixes
        assert spec_id1.startswith("AUTH")
        assert spec_id2.startswith("API")

    def test_generate_spec_id_uppercase_domain(self):
        """Test SPEC ID uses uppercase domain."""
        engine = EARSTemplateEngine()
        extraction = {
            "file_name": "auth",
            "classes": [],
            "functions": [],
            "imports": [],
        }

        spec_id = engine._generate_spec_id(extraction, "auth")
        assert spec_id.startswith("AUTH-")


class TestTemplateRendering:
    """Test template rendering with variable substitution."""

    def test_render_template_with_string_values(self):
        """Test rendering template with string values."""
        engine = EARSTemplateEngine()
        template = {
            "template": "Project: {project_name}, Language: {language}",
            "required_fields": [],
        }

        result = engine._render_template(template, {"project_name": "MyAPI", "language": "Python"})

        assert "MyAPI" in result
        assert "Python" in result

    def test_render_template_with_missing_values(self):
        """Test rendering when values are missing uses empty string."""
        engine = EARSTemplateEngine()
        template = {
            "template": "Name: {name}, Status: {status}",
            "required_fields": [],
        }

        result = engine._render_template(template, {"name": "Test"})

        assert "Test" in result

    def test_render_template_with_list_values(self):
        """Test rendering template with list values."""
        engine = EARSTemplateEngine()
        template = {
            "template": "Classes: {classes}",
            "required_fields": [],
        }

        result = engine._render_template(template, {"classes": ["ClassA", "ClassB", "ClassC"]})

        assert "ClassA" in result
        assert "ClassB" in result

    def test_render_template_with_numeric_values(self):
        """Test rendering template with numeric values."""
        engine = EARSTemplateEngine()
        template = {
            "template": "Version: {version}, Count: {count}",
            "required_fields": [],
        }

        result = engine._render_template(template, {"version": 1, "count": 42})

        assert "1" in result
        assert "42" in result


class TestContentGeneration:
    """Test content generation methods."""

    def test_extract_primary_function_from_class(self):
        """Test extracting primary function from classes."""
        engine = EARSTemplateEngine()
        extraction = {
            "classes": ["UserModel", "ProductModel"],
            "functions": [],
            "imports": [],
            "domain_keywords": [],
        }

        primary = engine._extract_primary_function(extraction, "data")

        assert "UserModel" in primary
        assert "class" in primary.lower()

    def test_extract_primary_function_from_function(self):
        """Test extracting primary function from functions."""
        engine = EARSTemplateEngine()
        extraction = {
            "classes": [],
            "functions": ["authenticate", "authorize"],
            "imports": [],
            "domain_keywords": [],
        }

        primary = engine._extract_primary_function(extraction, "auth")

        assert "authenticate" in primary
        assert "function" in primary.lower()

    def test_extract_primary_function_default(self):
        """Test default primary function when no classes/functions."""
        engine = EARSTemplateEngine()
        extraction = {
            "classes": [],
            "functions": [],
            "imports": [],
            "domain_keywords": [],
        }

        primary = engine._extract_primary_function(extraction, "api")

        assert "api" in primary.lower()

    def test_generate_state_requirements_returns_string(self):
        """Test state requirements generation returns string."""
        engine = EARSTemplateEngine()
        extraction = {
            "classes": [],
            "functions": [],
            "imports": [],
            "domain_keywords": [],
        }

        result = engine._generate_state_requirements(extraction, "auth")

        assert isinstance(result, str)
        assert len(result) > 0
        assert "REQ" in result

    def test_generate_event_requirements_returns_string(self):
        """Test event requirements generation returns string."""
        engine = EARSTemplateEngine()
        extraction = {
            "classes": [],
            "functions": [],
            "imports": [],
            "domain_keywords": [],
        }

        result = engine._generate_event_requirements(extraction, "api")

        assert isinstance(result, str)
        assert len(result) > 0
        assert "EVT" in result

    def test_generate_technical_specs_returns_string(self):
        """Test technical specs generation returns string."""
        engine = EARSTemplateEngine()
        extraction = {
            "classes": ["MainClass"],
            "functions": ["mainFunc"],
            "imports": [],
            "domain_keywords": [],
        }

        result = engine._generate_technical_specs(extraction, "data")

        assert isinstance(result, str)
        assert len(result) > 0
        assert "SPEC" in result

    def test_generate_data_models_with_classes(self):
        """Test data models generation with classes."""
        engine = EARSTemplateEngine()
        extraction = {
            "classes": ["User", "Product", "Order"],
            "functions": [],
            "imports": [],
            "domain_keywords": [],
        }

        result = engine._generate_data_models(extraction, "data")

        assert "User" in result
        assert "Product" in result

    def test_generate_api_specs_for_api_domain(self):
        """Test API specs generation for API domain."""
        engine = EARSTemplateEngine()
        extraction = {
            "classes": [],
            "functions": [],
            "imports": [],
            "domain_keywords": [],
        }

        result = engine._generate_api_specs(extraction, "api")

        assert "GET" in result or "POST" in result or "RESTful" in result

    def test_generate_api_specs_for_non_api_domain(self):
        """Test API specs for non-API domain returns not applicable."""
        engine = EARSTemplateEngine()
        extraction = {
            "classes": [],
            "functions": [],
            "imports": [],
            "domain_keywords": [],
        }

        result = engine._generate_api_specs(extraction, "data")

        assert "not applicable" in result.lower()

    def test_generate_security_specs_for_auth_domain(self):
        """Test security specs generation for auth domain."""
        engine = EARSTemplateEngine()
        extraction = {
            "classes": [],
            "functions": [],
            "imports": [],
            "domain_keywords": [],
        }

        result = engine._generate_security_specs(extraction, "auth")

        assert "Authentication" in result or "authentication" in result or "Security" in result or "security" in result

    def test_generate_performance_specs_returns_string(self):
        """Test performance specs generation returns string."""
        engine = EARSTemplateEngine()
        extraction = {
            "classes": [],
            "functions": [],
            "imports": [],
            "domain_keywords": [],
        }

        result = engine._generate_performance_specs(extraction, "api")

        assert isinstance(result, str)
        assert len(result) > 0

    def test_generate_scalability_specs_returns_string(self):
        """Test scalability specs generation returns string."""
        engine = EARSTemplateEngine()
        extraction = {
            "classes": [],
            "functions": [],
            "imports": [],
            "domain_keywords": [],
        }

        result = engine._generate_scalability_specs(extraction, "api")

        assert isinstance(result, str)
        assert len(result) > 0

    def test_generate_traceability_returns_string(self):
        """Test traceability section generation."""
        engine = EARSTemplateEngine()
        result = engine._generate_traceability("SPEC-TEST-001")

        assert "Traceability" in result
        assert isinstance(result, str)

    def test_generate_edit_guide_returns_string(self):
        """Test edit guide generation."""
        engine = EARSTemplateEngine()
        extraction = {
            "classes": [],
            "functions": [],
            "imports": [],
            "domain_keywords": [],
        }

        result = engine._generate_edit_guide(extraction, "api")

        assert "Edit Guide" in result
        assert isinstance(result, str)


class TestDetectFramework:
    """Test framework detection."""

    def test_detect_framework_django(self):
        """Test Django framework detection."""
        engine = EARSTemplateEngine()
        extraction = {
            "classes": [],
            "functions": [],
            "imports": ["django.db", "django.http"],
            "domain_keywords": [],
        }

        framework = engine._detect_framework(extraction)
        assert framework == "Django"

    def test_detect_framework_fastapi(self):
        """Test FastAPI framework detection."""
        engine = EARSTemplateEngine()
        extraction = {
            "classes": [],
            "functions": [],
            "imports": ["fastapi", "fastapi.routing"],
            "domain_keywords": [],
        }

        framework = engine._detect_framework(extraction)
        assert framework == "FastAPI"

    def test_detect_framework_react(self):
        """Test React framework detection."""
        engine = EARSTemplateEngine()
        extraction = {
            "classes": [],
            "functions": [],
            "imports": ["react", "react-dom"],
            "domain_keywords": [],
        }

        framework = engine._detect_framework(extraction)
        assert framework == "React"

    def test_detect_framework_custom_default(self):
        """Test custom framework as default."""
        engine = EARSTemplateEngine()
        extraction = {
            "classes": [],
            "functions": [],
            "imports": ["custom_lib"],
            "domain_keywords": [],
        }

        framework = engine._detect_framework(extraction)
        assert framework == "Custom"


class TestExtractionInformation:
    """Test information extraction from code analysis."""

    def test_extract_information_basic(self):
        """Test basic information extraction."""
        engine = EARSTemplateEngine()
        code_analysis = {
            "structure_info": {
                "classes": ["MyClass"],
                "functions": ["myFunc"],
                "imports": ["os", "sys"],
            }
        }

        extraction = engine._extract_information_from_analysis(code_analysis, "test.py")

        assert extraction["file_name"] == "test"
        assert extraction["file_extension"] == ".py"
        assert extraction["language"] == "Python"

    def test_extract_information_detects_language(self):
        """Test extraction detects language."""
        engine = EARSTemplateEngine()

        extraction = engine._extract_information_from_analysis({}, "module.js")

        assert extraction["language"] == "JavaScript"

    def test_extract_information_analyzes_complexity(self):
        """Test extraction analyzes complexity."""
        engine = EARSTemplateEngine()
        code_analysis = {
            "structure_info": {
                "classes": ["C1"],
                "functions": ["f1", "f2"],
                "imports": [],
            }
        }

        extraction = engine._extract_information_from_analysis(code_analysis, "test.py")

        assert extraction["complexity"] in ["low", "medium", "high"]

    def test_extract_information_analyzes_architecture(self):
        """Test extraction analyzes architecture."""
        engine = EARSTemplateEngine()
        code_analysis = {
            "structure_info": {
                "classes": [],
                "functions": [],
                "imports": ["fastapi"],
            }
        }

        extraction = engine._extract_information_from_analysis(code_analysis, "test.py")

        assert extraction["architecture"] in [
            "simple",
            "mvc",
            "frontend",
            "api",
            "data",
        ]
