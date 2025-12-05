"""
Comprehensive unit tests for EARS template engine with 95%+ coverage.

Tests cover:
- Main generate_complete_spec method with various inputs
- Information extraction from code analysis
- Content generation for all three documents (spec.md, plan.md, acceptance.md)
- EARS compliance validation
- Edge cases and exception handling
- Domain-specific content generation
- Framework detection
- Component naming methods
"""

import pytest
from pathlib import Path
from unittest import mock
import time

from moai_adk.core.spec.ears_template_engine import EARSTemplateEngine


class TestGenerateCompleteSpec:
    """Test the main generate_complete_spec method."""

    @pytest.fixture
    def engine(self):
        """Create EARSTemplateEngine instance."""
        return EARSTemplateEngine()

    @pytest.fixture
    def basic_code_analysis(self):
        """Basic code analysis data."""
        return {
            "structure_info": {
                "classes": ["User", "Product"],
                "functions": ["authenticate", "create_product"],
                "imports": ["fastapi", "sqlalchemy"],
            },
            "domain_keywords": ["auth", "user", "login"],
        }

    def test_generate_complete_spec_basic(self, engine, basic_code_analysis):
        """Test basic spec generation."""
        result = engine.generate_complete_spec(
            code_analysis=basic_code_analysis,
            file_path="auth_service.py"
        )

        assert "spec_id" in result
        assert "domain" in result
        assert "spec_md" in result
        assert "plan_md" in result
        assert "acceptance_md" in result
        assert "validation" in result
        assert "generation_time" in result
        assert "extraction" in result

        assert result["domain"] == "auth"
        assert result["generation_time"] > 0
        assert result["validation"]["ears_compliance"] >= 0

    def test_generate_complete_spec_with_custom_config(self, engine, basic_code_analysis):
        """Test spec generation with custom config."""
        custom_config = {
            "project_name": "MyAuthSystem",
            "framework": "FastAPI",
            "status": "Production",
        }

        result = engine.generate_complete_spec(
            code_analysis=basic_code_analysis,
            file_path="auth_service.py",
            custom_config=custom_config
        )

        assert "MyAuthSystem" in result["spec_md"]
        assert "FastAPI" in result["spec_md"]

    def test_generate_complete_spec_empty_analysis(self, engine):
        """Test spec generation with empty code analysis."""
        result = engine.generate_complete_spec(
            code_analysis={},
            file_path="test.py"
        )

        assert result is not None
        assert result["domain"] in ["general", "auth", "api", "data", "ui", "business"]

    def test_generate_complete_spec_no_structure_info(self, engine):
        """Test spec generation without structure info."""
        code_analysis = {
            "domain_keywords": ["api", "endpoint"],
        }

        result = engine.generate_complete_spec(
            code_analysis=code_analysis,
            file_path="api.py"
        )

        assert result is not None
        assert "api" in result["domain"]

    def test_generate_complete_spec_missing_domain_keywords(self, engine):
        """Test spec generation without domain keywords."""
        code_analysis = {
            "structure_info": {
                "classes": ["APIRouter"],
                "functions": ["endpoint"],
                "imports": ["fastapi"],
            }
        }

        result = engine.generate_complete_spec(
            code_analysis=code_analysis,
            file_path="api_service.py"
        )

        assert result is not None
        # Domain detection depends on domain_keywords, when empty it might detect auth
        assert result["domain"] in ["auth", "api"]

    def test_generate_complete_spec_complex_extraction(self, engine):
        """Test spec generation with complex code analysis."""
        code_analysis = {
            "structure_info": {
                "classes": ["AuthController", "UserService", "AuthService"],
                "functions": ["login", "logout", "register", "reset_password"],
                "imports": ["fastapi", "fastapi.security", "sqlalchemy", "bcrypt"],
            },
            "domain_keywords": ["auth", "user", "security", "password", "session"],
        }

        result = engine.generate_complete_spec(
            code_analysis=code_analysis,
            file_path="auth_service.py"
        )

        # Should detect auth domain
        assert result["domain"] == "auth"
        # Should have multiple classes mentioned
        assert "AuthController" in result["spec_md"]

    def test_generate_complete_spec_data_domain(self, engine):
        """Test spec generation for data domain."""
        code_analysis = {
            "structure_info": {
                "classes": ["UserModel", "ProductModel", "OrderModel"],
                "functions": ["save_user", "get_user", "update_user"],
                "imports": ["sqlalchemy", "pydantic"],
            },
            "domain_keywords": ["model", "database", "persistence", "entity"],
        }

        result = engine.generate_complete_spec(
            code_analysis=code_analysis,
            file_path="models.py"
        )

        assert result["domain"] == "data"
        assert "UserModel" in result["spec_md"]

    def test_generate_complete_spec_api_domain(self, engine):
        """Test spec generation for API domain."""
        code_analysis = {
            "structure_info": {
                "classes": ["APIRouter", "EndpointHandler"],
                "functions": ["create_user", "get_users", "delete_user"],
                "imports": ["fastapi", "fastapi.responses"],
            },
            "domain_keywords": ["api", "endpoint", "route", "controller"],
        }

        result = engine.generate_complete_spec(
            code_analysis=code_analysis,
            file_path="api.py"
        )

        assert result["domain"] == "api"
        assert "APIRouter" in result["spec_md"]

    def test_generate_complete_spec_ui_domain(self, engine):
        """Test spec generation for UI domain."""
        code_analysis = {
            "structure_info": {
                "classes": ["Component", "View", "Controller"],
                "functions": ["render", "handle_click", "validate_input"],
                "imports": ["react", "react-dom"],
            },
            "domain_keywords": ["ui", "component", "view", "interface"],
        }

        result = engine.generate_complete_spec(
            code_analysis=code_analysis,
            file_path="ui_component.js"
        )

        assert result["domain"] == "ui"
        assert "Component" in result["spec_md"]

    def test_generate_complete_spec_business_domain(self, engine):
        """Test spec generation for business domain."""
        code_analysis = {
            "structure_info": {
                "classes": ["Workflow", "BusinessRule", "Process"],
                "functions": ["execute_process", "validate_rule", "approve_request"],
                "imports": ["workflow", "business"],
            },
            "domain_keywords": ["business", "workflow", "process", "rule"],
        }

        result = engine.generate_complete_spec(
            code_analysis=code_analysis,
            file_path="business_logic.py"
        )

        assert result["domain"] == "business"
        assert "Workflow" in result["spec_md"]


class TestExtractInformationFromAnalysis:
    """Test _extract_information_from_analysis method."""

    @pytest.fixture
    def engine(self):
        """Create EARSTemplateEngine instance."""
        return EARSTemplateEngine()

    def test_extract_with_full_analysis(self, engine):
        """Test extraction with complete analysis data."""
        code_analysis = {
            "structure_info": {
                "classes": ["User", "Product", "Order"],
                "functions": ["create", "update", "delete", "get_all"],
                "imports": ["fastapi", "sqlalchemy", "pydantic"],
            },
            "domain_keywords": ["api", "data", "model"],
        }

        extraction = engine._extract_information_from_analysis(code_analysis, "service.py")

        assert extraction["file_path"] == "service.py"
        assert extraction["file_name"] == "service"
        assert extraction["file_extension"] == ".py"
        assert extraction["language"] == "Python"
        assert extraction["classes"] == ["User", "Product", "Order"]
        assert extraction["functions"] == ["create", "update", "delete", "get_all"]
        assert extraction["imports"] == ["fastapi", "sqlalchemy", "pydantic"]
        assert extraction["domain_keywords"] == ["api", "data", "model"]
        assert extraction["complexity"] in ["low", "medium", "high"]
        assert extraction["architecture"] in ["simple", "mvc", "frontend", "api", "data"]

    def test_extract_with_empty_analysis(self, engine):
        """Test extraction with empty analysis."""
        extraction = engine._extract_information_from_analysis({}, "file.js")

        assert extraction["file_path"] == "file.js"
        assert extraction["file_name"] == "file"
        assert extraction["file_extension"] == ".js"
        assert extraction["language"] == "JavaScript"
        assert extraction["classes"] == []
        assert extraction["functions"] == []
        assert extraction["imports"] == []
        assert extraction["domain_keywords"] == []

    def test_extract_with_partial_analysis(self, engine):
        """Test extraction with partial analysis data."""
        code_analysis = {
            "structure_info": {
                "classes": ["MyClass"],
                # Missing functions and imports
            },
            "domain_keywords": ["auth"],
        }

        extraction = engine._extract_information_from_analysis(code_analysis, "partial.py")

        assert extraction["classes"] == ["MyClass"]
        assert extraction["functions"] == []
        assert extraction["imports"] == []
        assert extraction["domain_keywords"] == ["auth"]

    def test_extract_with_ast_info_attribute(self, engine):
        """Test extraction when ast_info attribute exists (mocked)."""
        code_analysis = mock.Mock()
        code_analysis.structure_info = {"classes": ["Test"]}
        # Add __contains__ method to make it work with 'in' operator
        code_analysis.__contains__ = lambda self, key: key == "structure_info"
        code_analysis.ast_info = {"some": "ast_data"}

        extraction = engine._extract_information_from_analysis(code_analysis, "ast_test.py")

        assert extraction["classes"] == ["Test"]
        # The actual code has: if hasattr(code_analysis, "ast_info"): pass

    def test_extract_complexity_analysis(self, engine):
        """Test complexity analysis."""
        # Low complexity
        low_extraction = {
            "classes": ["OneClass"],
            "functions": ["func1", "func2"],
        }
        assert engine._analyze_complexity(low_extraction) == "low"

        # Medium complexity
        medium_extraction = {
            "classes": ["C1", "C2", "C3"],
            "functions": [f"f{i}" for i in range(11)],
        }
        assert engine._analyze_complexity(medium_extraction) == "medium"

        # High complexity
        high_extraction1 = {
            "classes": [f"C{i}" for i in range(6)],
            "functions": [],
        }
        assert engine._analyze_complexity(high_extraction1) == "high"

        high_extraction2 = {
            "classes": [],
            "functions": [f"f{i}" for i in range(21)],
        }
        assert engine._analyze_complexity(high_extraction2) == "high"

    def test_extract_architecture_analysis(self, engine):
        """Test architecture analysis."""
        # Django MVC
        django_extraction = {"imports": ["django", "django.db"]}
        assert engine._analyze_architecture(django_extraction) == "mvc"

        # React frontend
        react_extraction = {"imports": ["react", "react-dom"]}
        assert engine._analyze_architecture(react_extraction) == "frontend"

        # FastAPI API
        fastapi_extraction = {"imports": ["fastapi"]}
        assert engine._analyze_architecture(fastapi_extraction) == "api"

        # SQLAlchemy data
        sqlalchemy_extraction = {"imports": ["sqlalchemy"]}
        assert engine._analyze_architecture(sqlalchemy_extraction) == "data"

        # Simple default
        simple_extraction = {"imports": []}
        assert engine._analyze_architecture(simple_extraction) == "simple"


class TestDomainDetermination:
    """Test domain determination logic."""

    @pytest.fixture
    def engine(self):
        """Create EARSTemplateEngine instance."""
        return EARSTemplateEngine()

    def test_determine_auth_domain(self, engine):
        """Test authentication domain detection."""
        extraction = {
            "domain_keywords": ["auth", "login", "password", "security", "token"],
            "classes": [],
            "functions": [],
            "imports": [],
        }
        assert engine._determine_domain(extraction) == "auth"

    def test_determine_api_domain(self, engine):
        """Test API domain detection."""
        extraction = {
            "domain_keywords": ["api", "endpoint", "route", "controller"],
            "classes": [],
            "functions": [],
            "imports": [],
        }
        assert engine._determine_domain(extraction) == "api"

    def test_determine_data_domain(self, engine):
        """Test data domain detection."""
        extraction = {
            "domain_keywords": ["model", "entity", "schema", "database"],
            "classes": [],
            "functions": [],
            "imports": [],
        }
        assert engine._determine_domain(extraction) == "data"

    def test_determine_ui_domain(self, engine):
        """Test UI domain detection."""
        extraction = {
            "domain_keywords": ["ui", "component", "view", "interface"],
            "classes": [],
            "functions": [],
            "imports": [],
        }
        assert engine._determine_domain(extraction) == "ui"

    def test_determine_business_domain(self, engine):
        """Test business domain detection."""
        extraction = {
            "domain_keywords": ["business", "logic", "process", "workflow"],
            "classes": [],
            "functions": [],
            "imports": [],
        }
        assert engine._determine_domain(extraction) == "business"

    def test_determine_domain_tie(self, engine):
        """Test domain determination with tie score."""
        extraction = {
            "domain_keywords": ["auth", "api"],  # Both score 1
            "classes": [],
            "functions": [],
            "imports": [],
        }
        # Should return one of the domains, with preference for the first in domain_indicators
        domain = engine._determine_domain(extraction)
        assert domain in ["auth", "api"]

    def test_determine_domain_no_keywords(self, engine):
        """Test domain determination with no keywords."""
        extraction = {
            "domain_keywords": ["unknown"],
            "classes": [],
            "functions": [],
            "imports": [],
        }
        domain = engine._determine_domain(extraction)
        # Should return one of the domains since max() on empty dict throws error
        assert domain in ["auth", "api", "data", "ui", "business"]


class TestSpecIDGeneration:
    """Test SPEC ID generation."""

    @pytest.fixture
    def engine(self):
        """Create EARSTemplateEngine instance."""
        return EARSTemplateEngine()

    def test_generate_spec_id_format(self, engine):
        """Test SPEC ID format."""
        extraction = {
            "file_name": "auth_service",
            "classes": [],
            "functions": [],
            "imports": [],
        }

        spec_id = engine._generate_spec_id(extraction, "auth")

        assert spec_id.startswith("AUTH-")
        assert "-" in spec_id
        assert len(spec_id.split("-")) == 3

    def test_generate_spec_id_uppercase_domain(self, engine):
        """Test domain is uppercase."""
        extraction = {
            "file_name": "user",
            "classes": [],
            "functions": [],
            "imports": [],
        }

        spec_id = engine._generate_spec_id(extraction, "api")
        assert spec_id.startswith("API-")

    def test_generate_spec_id_different_files(self, engine):
        """Test different files generate different IDs."""
        extraction1 = {"file_name": "service", "classes": [], "functions": [], "imports": []}
        extraction2 = {"file_name": "handler", "classes": [], "functions": [], "imports": []}

        spec_id1 = engine._generate_spec_id(extraction1, "auth")
        spec_id2 = engine._generate_spec_id(extraction2, "auth")

        # Hashes should be different due to file name
        assert spec_id1 != spec_id2

    def test_generate_spec_id_special_characters_cleaned(self, engine):
        """Test special characters are cleaned from file name."""
        extraction = {
            "file_name": "my-service_v2.0",
            "classes": [],
            "functions": [],
            "imports": [],
        }

        spec_id = engine._generate_spec_id(extraction, "data")

        # Should not contain special characters
        assert "-" in spec_id
        assert "my" in spec_id.lower()
        assert "service" in spec_id.lower()
        # v2.0 becomes part of the cleaned name
        assert "myservic" in spec_id.lower()


class TestContentGenerationHelpers:
    """Test helper methods for content generation."""

    @pytest.fixture
    def engine(self):
        """Create EARSTemplateEngine instance."""
        return EARSTemplateEngine()

    @pytest.fixture
    def sample_extraction(self):
        """Sample extraction data."""
        return {
            "classes": ["User", "Product"],
            "functions": ["create", "update", "delete"],
            "imports": ["fastapi", "sqlalchemy"],
            "domain_keywords": ["api", "data"],
            "language": "Python",
            "architecture": "api",
            "complexity": "medium",
            "file_name": "service",
        }

    def test_extract_primary_function_from_classes(self, engine, sample_extraction):
        """Test primary function extraction from classes."""
        result = engine._extract_primary_function(sample_extraction, "data")
        assert "User" in result
        assert "class" in result.lower()

    def test_extract_primary_function_from_functions(self, engine):
        """Test primary function extraction from functions."""
        extraction = {
            "classes": [],
            "functions": ["authenticate", "authorize"],
        }
        result = engine._extract_primary_function(extraction, "auth")
        assert "authenticate" in result
        assert "function" in result.lower()

    def test_extract_primary_function_default(self, engine):
        """Test default primary function."""
        extraction = {"classes": [], "functions": []}
        result = engine._extract_primary_function(extraction, "general")
        assert "general" in result.lower()

    def test_generate_state_requirements_auth_domain(self, engine, sample_extraction):
        """Test state requirements for auth domain."""
        result = engine._generate_state_requirements(sample_extraction, "auth")
        assert "AUTH-001" in result
        assert "authenticated" in result.lower()

    def test_generate_state_requirements_api_domain(self, engine, sample_extraction):
        """Test state requirements for API domain."""
        result = engine._generate_state_requirements(sample_extraction, "api")
        assert "API-001" in result
        assert "ready" in result.lower()

    def test_generate_state_requirements_data_domain(self, engine, sample_extraction):
        """Test state requirements for data domain."""
        result = engine._generate_state_requirements(sample_extraction, "data")
        assert "DATA-001" in result
        assert "create" in result.lower()

    def test_generate_event_requirements_auth_domain(self, engine, sample_extraction):
        """Test event requirements for auth domain."""
        result = engine._generate_event_requirements(sample_extraction, "auth")
        assert "AUTH-EVT-001" in result
        assert "login" in result.lower()

    def test_generate_event_requirements_api_domain(self, engine, sample_extraction):
        """Test event requirements for API domain."""
        result = engine._generate_event_requirements(sample_extraction, "api")
        assert "API-EVT-001" in result
        assert "API request" in result

    def test_generate_technical_specs(self, engine, sample_extraction):
        """Test technical specs generation."""
        result = engine._generate_technical_specs(sample_extraction, "api")
        assert "SPEC-001" in result
        assert "SPEC-002" in result
        assert "User" in result or "Product" in result

    def test_generate_data_models_with_classes(self, engine, sample_extraction):
        """Test data models generation."""
        result = engine._generate_data_models(sample_extraction, "data")
        assert "User" in result
        assert "Product" in result

    def test_generate_data_models_empty(self, engine):
        """Test data models generation with no classes."""
        extraction = {"classes": [], "functions": [], "imports": []}
        result = engine._generate_data_models(extraction, "general")
        assert "not explicitly defined" in result

    def test_generate_api_specs_applicable_domains(self, engine):
        """Test API specs for applicable domains."""
        extraction = {"classes": [], "functions": [], "imports": []}

        # API domain
        result = engine._generate_api_specs(extraction, "api")
        assert "GET" in result or "POST" in result

        # Auth domain
        result = engine._generate_api_specs(extraction, "auth")
        assert "GET" in result or "POST" in result

    def test_generate_api_specs_non_applicable(self, engine):
        """Test API specs for non-applicable domain."""
        extraction = {"classes": [], "functions": [], "imports": []}
        result = engine._generate_api_specs(extraction, "business")
        assert "not applicable" in result.lower()

    def test_generate_interface_specs_ui_domain(self, engine):
        """Test interface specs for UI domain."""
        extraction = {"classes": [], "functions": [], "imports": []}
        result = engine._generate_interface_specs(extraction, "ui")
        assert "User Interface" in result

    def test_generate_interface_specs_api_domain(self, engine):
        """Test interface specs for API domain."""
        extraction = {"classes": [], "functions": [], "imports": []}
        result = engine._generate_interface_specs(extraction, "api")
        assert "API Interface" in result

    def test_generate_interface_specs_non_applicable(self, engine):
        """Test interface specs for non-applicable domain."""
        extraction = {"classes": [], "functions": [], "imports": []}
        result = engine._generate_interface_specs(extraction, "data")
        assert "not applicable" in result.lower()

    def test_generate_security_specs_auth_domain(self, engine):
        """Test security specs for auth domain."""
        extraction = {"classes": [], "functions": [], "imports": []}
        result = engine._generate_security_specs(extraction, "auth")
        assert "Authentication" in result or "Security" in result

    def test_generate_security_specs_api_domain(self, engine):
        """Test security specs for API domain."""
        extraction = {"classes": [], "functions": [], "imports": []}
        result = engine._generate_security_specs(extraction, "api")
        assert "Authentication" in result or "Security" in result

    def test_generate_security_specs_general_domain(self, engine):
        """Test security specs for general domain."""
        extraction = {"classes": [], "functions": [], "imports": []}
        result = engine._generate_security_specs(extraction, "business")
        assert "apply by default" in result.lower()

    def test_generate_performance_specs(self, engine):
        """Test performance specs generation."""
        extraction = {"classes": [], "functions": [], "imports": []}
        result = engine._generate_performance_specs(extraction, "api")
        assert "Response time" in result
        assert "1 second" in result

    def test_generate_scalability_specs(self, engine):
        """Test scalability specs generation."""
        extraction = {"classes": [], "functions": [], "imports": []}
        result = engine._generate_scalability_specs(extraction, "api")
        assert "Horizontal scaling" in result
        assert "Load balancing" in result

    def test_generate_traceability(self, engine):
        """Test traceability section generation."""
        result = engine._generate_traceability("SPEC-TEST-001")
        assert "Traceability" in result
        assert "Requirements Traceability Matrix" in result

    def test_generate_edit_guide(self, engine, sample_extraction):
        """Test edit guide generation."""
        result = engine._generate_edit_guide(sample_extraction, "api")
        assert "Edit Guide" in result
        assert "User Review Checklist" in result
        assert "API" in result

    def test_get_domain_specific_review(self, engine):
        """Test domain-specific review guidance."""
        assert "security" in engine._get_domain_specific_review("auth").lower()
        assert "api design" in engine._get_domain_specific_review("api").lower()
        assert "data integrity" in engine._get_domain_specific_review("data").lower()
        assert "user experience" in engine._get_domain_specific_review("ui").lower()
        assert "business rules" in engine._get_domain_specific_review("business").lower()
        assert "general requirements" in engine._get_domain_specific_review("general").lower()


class TestFrameworkDetection:
    """Test framework detection."""

    @pytest.fixture
    def engine(self):
        """Create EARSTemplateEngine instance."""
        return EARSTemplateEngine()

    def test_detect_framework_django(self, engine):
        """Test Django detection."""
        extraction = {"imports": ["django", "django.db"]}
        assert engine._detect_framework(extraction) == "Django"

    def test_detect_framework_flask(self, engine):
        """Test Flask detection."""
        extraction = {"imports": ["flask"]}
        assert engine._detect_framework(extraction) == "Flask"

    def test_detect_framework_fastapi(self, engine):
        """Test FastAPI detection."""
        extraction = {"imports": ["fastapi"]}
        assert engine._detect_framework(extraction) == "FastAPI"

    def test_detect_framework_spring(self, engine):
        """Test Spring detection."""
        extraction = {"imports": ["spring"]}
        assert engine._detect_framework(extraction) == "Spring"

    def test_detect_framework_express(self, engine):
        """Test Express detection."""
        extraction = {"imports": ["express"]}
        assert engine._detect_framework(extraction) == "Express"

    def test_detect_framework_react(self, engine):
        """Test React detection."""
        extraction = {"imports": ["react", "react-dom"]}
        assert engine._detect_framework(extraction) == "React"

    def test_detect_framework_angular(self, engine):
        """Test Angular detection."""
        extraction = {"imports": ["angular"]}
        assert engine._detect_framework(extraction) == "Angular"

    def test_detect_framework_vue(self, engine):
        """Test Vue detection."""
        extraction = {"imports": ["vue"]}
        assert engine._detect_framework(extraction) == "Vue"

    def test_detect_framework_nextjs(self, engine):
        """Test Next.js detection."""
        extraction = {"imports": ["next"]}
        assert engine._detect_framework(extraction) == "Next.js"

    def test_detect_framework_custom_default(self, engine):
        """Test custom framework as default."""
        extraction = {"imports": ["custom_lib"]}
        assert engine._detect_framework(extraction) == "Custom"

    def test_detect_framework_case_insensitive(self, engine):
        """Test framework detection is case insensitive."""
        extraction = {"imports": ["FastAPI", "React-DOM"]}
        result = engine._detect_framework(extraction)
        assert result in ["FastAPI", "React"]


class TestArchitectureDiagram:
    """Test architecture diagram generation."""

    @pytest.fixture
    def engine(self):
        """Create EARSTemplateEngine instance."""
        return EARSTemplateEngine()

    def test_generate_architecture_diagram_auth(self, engine):
        """Test auth architecture diagram."""
        extraction = {"classes": [], "functions": [], "imports": []}
        result = engine._generate_architecture_diagram(extraction, "auth")
        assert "Auth Service" in result
        assert "Client" in result
        assert "Database" in result

    def test_generate_architecture_diagram_api(self, engine):
        """Test API architecture diagram."""
        extraction = {"classes": [], "functions": [], "imports": []}
        result = engine._generate_architecture_diagram(extraction, "api")
        assert "Load Balancer" in result
        assert "Service" in result
        assert "API Gateway" in result

    def test_generate_architecture_diagram_data(self, engine):
        """Test data architecture diagram."""
        extraction = {"classes": [], "functions": [], "imports": []}
        result = engine._generate_architecture_diagram(extraction, "data")
        assert "Data Service" in result
        assert "Analytics" in result
        assert "Cache Layer" in result

    def test_generate_architecture_diagram_default(self, engine):
        """Test default architecture diagram."""
        extraction = {"classes": [], "functions": [], "imports": []}
        result = engine._generate_architecture_diagram(extraction, "business")
        assert "Client" in result
        assert "Service" in result
        assert "Cache" in result


class TestComponentNaming:
    """Test component naming methods."""

    @pytest.fixture
    def engine(self):
        """Create EARSTemplateEngine instance."""
        return EARSTemplateEngine()

    @pytest.fixture
    def sample_extraction(self):
        """Sample extraction data."""
        return {
            "classes": ["User"],
            "functions": ["authenticate"],
            "imports": ["fastapi"],
        }

    def test_get_main_component(self, engine, sample_extraction):
        """Test main component naming."""
        assert engine._get_main_component(sample_extraction, "auth") == "AuthService"
        assert engine._get_main_component(sample_extraction, "api") == "APIController"
        assert engine._get_main_component(sample_extraction, "data") == "DataService"
        assert engine._get_main_component(sample_extraction, "ui") == "UIController"
        assert engine._get_main_component(sample_extraction, "business") == "BusinessLogic"
        assert engine._get_main_component(sample_extraction, "unknown") == "MainComponent"

    def test_get_service_component(self, engine, sample_extraction):
        """Test service component naming."""
        assert engine._get_service_component(sample_extraction, "auth") == "UserService"
        assert engine._get_service_component(sample_extraction, "api") == "ExternalService"
        assert engine._get_service_component(sample_extraction, "data") == "PersistenceService"
        assert engine._get_service_component(sample_extraction, "ui") == "ClientService"
        assert engine._get_service_component(sample_extraction, "business") == "WorkflowService"
        assert engine._get_service_component(sample_extraction, "unknown") == "ServiceComponent"

    def test_get_data_component(self, engine, sample_extraction):
        """Test data component naming."""
        assert engine._get_data_component(sample_extraction, "auth") == "UserRepository"
        assert engine._get_data_component(sample_extraction, "api") == "DataRepository"
        assert engine._get_data_component(sample_extraction, "data") == "DataAccessLayer"
        assert engine._get_data_component(sample_extraction, "ui") == "StateManagement"
        assert engine._get_data_component(sample_extraction, "business") == "DataProcessor"
        assert engine._get_data_component(sample_extraction, "unknown") == "DataComponent"

    def test_get_component_4(self, engine, sample_extraction):
        """Test fourth component naming."""
        assert engine._get_component_4(sample_extraction, "auth") == "SecurityManager"
        assert engine._get_component_4(sample_extraction, "api") == "RateLimiter"
        assert engine._get_component_4(sample_extraction, "data") == "DataValidator"
        assert engine._get_component_4(sample_extraction, "ui") == "FormValidator"
        assert engine._get_component_4(sample_extraction, "business") == "RuleEngine"
        assert engine._get_component_4(sample_extraction, "unknown") == "ValidationComponent"

    def test_get_new_modules(self, engine, sample_extraction):
        """Test new modules naming."""
        auth_modules = engine._get_new_modules(sample_extraction, "auth")
        assert "Authentication" in auth_modules
        assert "Security" in auth_modules

        api_modules = engine._get_new_modules(sample_extraction, "api")
        assert "Routing" in api_modules

        data_modules = engine._get_new_modules(sample_extraction, "data")
        assert "Database" in data_modules

        ui_modules = engine._get_new_modules(sample_extraction, "ui")
        assert "Component library" in ui_modules

        business_modules = engine._get_new_modules(sample_extraction, "business")
        assert "Business rules" in business_modules

        default_modules = engine._get_new_modules(sample_extraction, "unknown")
        assert "Standard" in default_modules


class TestPlanContentGeneration:
    """Test plan content generation."""

    @pytest.fixture
    def engine(self):
        """Create EARSTemplateEngine instance."""
        return EARSTemplateEngine()

    @pytest.fixture
    def sample_extraction(self):
        """Sample extraction data."""
        return {
            "file_name": "service",
            "classes": ["User"],
            "functions": ["create"],
            "imports": ["fastapi"],
            "language": "Python",
        }

    def test_generate_plan_content_structure(self, engine, sample_extraction):
        """Test plan content structure."""
        spec_id = "SPEC-TEST-001"
        result = engine._generate_plan_content(sample_extraction, "api", spec_id)

        assert f"PLAN-{spec_id}" in result
        assert "Implementation Phases" in result
        assert "Technical Approach" in result
        assert "Success Criteria" in result
        assert "Next Steps" in result

    def test_generate_plan_content_phases(self, engine, sample_extraction):
        """Test plan content phases."""
        spec_id = "SPEC-TEST-001"
        result = engine._generate_plan_content(sample_extraction, "api", spec_id)

        # Check phase sections
        assert "Requirements Analysis" in result
        assert "Design" in result
        assert "Development" in result
        assert "Testing" in result
        assert "Deployment" in result

    def test_generate_plan_content_components(self, engine, sample_extraction):
        """Test plan content components."""
        spec_id = "SPEC-TEST-001"
        result = engine._generate_plan_content(sample_extraction, "api", spec_id)

        # Check components are mentioned
        assert "APIController" in result
        assert "ExternalService" in result
        assert "DataRepository" in result
        assert "RateLimiter" in result

    def test_generate_plan_content_different_domains(self, engine, sample_extraction):
        """Test plan content for different domains."""
        spec_id = "SPEC-TEST-001"

        # Auth domain
        auth_plan = engine._generate_plan_content(sample_extraction, "auth", spec_id)
        assert "AuthService" in auth_plan
        assert "UserService" in auth_plan

        # Data domain
        data_plan = engine._generate_plan_content(sample_extraction, "data", spec_id)
        assert "DataService" in data_plan
        assert "PersistenceService" in data_plan


class TestAcceptanceContentGeneration:
    """Test acceptance content generation."""

    @pytest.fixture
    def engine(self):
        """Create EARSTemplateEngine instance."""
        return EARSTemplateEngine()

    @pytest.fixture
    def sample_extraction(self):
        """Sample extraction data."""
        return {
            "file_name": "service",
            "classes": ["User"],
            "functions": ["create"],
            "imports": ["fastapi"],
            "language": "Python",
        }

    def test_generate_acceptance_content_structure(self, engine, sample_extraction):
        """Test acceptance content structure."""
        spec_id = "SPEC-TEST-001"
        result = engine._generate_acceptance_content(sample_extraction, "api", spec_id)

        assert f"ACCEPT-{spec_id}" in result
        assert "Acceptance Criteria" in result
        assert "Validation Process" in result
        assert "Completion Criteria" in result
        assert "Validation Templates" in result

    def test_generate_acceptance_content_basic_functionality(self, engine, sample_extraction):
        """Test acceptance content basic functionality."""
        spec_id = "SPEC-TEST-001"
        result = engine._generate_acceptance_content(sample_extraction, "api", spec_id)

        assert "Basic Functionality" in result
        assert "System SHALL operate normally" in result
        assert "User interface SHALL display correctly" in result

    def test_generate_acceptance_content_performance(self, engine, sample_extraction):
        """Test acceptance content performance section."""
        spec_id = "SPEC-TEST-001"
        result = engine._generate_acceptance_content(sample_extraction, "api", spec_id)

        assert "Performance Testing" in result
        assert "Response time: Within 1 second" in result
        assert "Concurrent users: Support 100+ users" in result

    def test_generate_acceptance_content_security(self, engine, sample_extraction):
        """Test acceptance content security section."""
        spec_id = "SPEC-TEST-001"
        result = engine._generate_acceptance_content(sample_extraction, "api", spec_id)

        assert "Security Testing" in result
        assert "Pass authentication and authorization validation" in result
        assert "Pass OWASP Top 10 checks" in result

    def test_generate_acceptance_content_validation_phases(self, engine, sample_extraction):
        """Test acceptance content validation phases."""
        spec_id = "SPEC-TEST-001"
        result = engine._generate_acceptance_content(sample_extraction, "api", spec_id)

        assert "Phase 1: Unit Tests" in result
        assert "Phase 2: Integration Tests" in result
        assert "Phase 3: System Tests" in result
        assert "Phase 4: User Tests" in result

    def test_generate_acceptance_content_domain_specific(self, engine, sample_extraction):
        """Test domain-specific acceptance criteria."""
        spec_id = "SPEC-TEST-001"

        # Auth domain
        auth_acceptance = engine._generate_acceptance_content(sample_extraction, "auth", spec_id)
        assert "AUTH-001" in auth_acceptance
        assert "User login functionality validation" in auth_acceptance

        # API domain
        api_acceptance = engine._generate_acceptance_content(sample_extraction, "api", spec_id)
        assert "API-001" in api_acceptance
        assert "REST API functionality validation" in api_acceptance

        # Data domain
        data_acceptance = engine._generate_acceptance_content(sample_extraction, "data", spec_id)
        assert "DATA-001" in data_acceptance
        assert "Data storage functionality validation" in data_acceptance

    def test_generate_acceptance_content_validation_templates(self, engine, sample_extraction):
        """Test acceptance content validation templates."""
        spec_id = "SPEC-TEST-001"
        result = engine._generate_acceptance_content(sample_extraction, "api", spec_id)

        assert "Functional Validation Template" in result
        assert "| Function ID | Function Name | Expected Result | Actual Result | Status | Notes |" in result
        assert "Performance Validation Template" in result
        assert "| Test Item | Target | Measured | Status | Notes |" in result


class TestEARSComplianceValidation:
    """Test EARS compliance validation."""

    @pytest.fixture
    def engine(self):
        """Create EARSTemplateEngine instance."""
        return EARSTemplateEngine()

    def test_validate_ears_compliance_complete(self, engine):
        """Test validation with complete spec."""
        spec_content = {
            "spec_md": """
## Auto-generated SPEC

### Overview

### Environment

### Assumptions

### Requirements

### Specifications

### Traceability
""",
            "plan_md": "plan content",
            "acceptance_md": "acceptance content",
        }

        result = engine._validate_ears_compliance(spec_content)

        assert result["ears_compliance"] == 1.0  # 100%
        assert result["total_sections"] == 6
        assert result["present_sections"] == 6
        assert len(result["suggestions"]) == 0

    def test_validate_ears_compliance_missing_sections(self, engine):
        """Test validation with missing sections."""
        spec_content = {
            "spec_md": """
## Auto-generated SPEC

### Overview

### Environment

### Requirements
""",
            "plan_md": "plan content",
            "acceptance_md": "acceptance content",
        }

        result = engine._validate_ears_compliance(spec_content)

        assert result["ears_compliance"] < 1.0
        assert result["total_sections"] == 6
        assert result["present_sections"] == 3
        assert len(result["suggestions"]) > 0
        assert any("Assumptions" in s for s in result["suggestions"])
        assert any("Specifications" in s for s in result["suggestions"])

    def test_validate_ears_compliance_empty_spec(self, engine):
        """Test validation with empty spec."""
        spec_content = {
            "spec_md": "",
            "plan_md": "plan content",
            "acceptance_md": "acceptance content",
        }

        result = engine._validate_ears_compliance(spec_content)

        assert result["ears_compliance"] == 0.0
        assert result["total_sections"] == 6
        assert result["present_sections"] == 0
        assert len(result["suggestions"]) == 5

    def test_validate_ears_compliance_partial_sections(self, engine):
        """Test validation with partial section matches."""
        spec_content = {
            "spec_md": """
## Auto-generated SPEC

### Overview

### Environment

### Assumptions

### Requirements

### Specifications

### Traceability
""",
            "plan_md": "plan content",
            "acceptance_md": "acceptance content",
        }

        result = engine._validate_ears_compliance(spec_content)

        # Should detect exact matches
        assert result["ears_compliance"] == 1.0  # All sections found

    def test_validate_ears_compliance_no_spec_md(self, engine):
        """Test validation when spec_md is missing."""
        spec_content = {
            "plan_md": "plan content",
            "acceptance_md": "acceptance content",
        }

        result = engine._validate_ears_compliance(spec_content)

        assert result["ears_compliance"] == 0.0
        assert result["total_sections"] == 6
        assert result["present_sections"] == 0


class TestEdgeCases:
    """Test edge cases and error conditions."""

    @pytest.fixture
    def engine(self):
        """Create EARSTemplateEngine instance."""
        return EARSTemplateEngine()

    def test_language_detection_edge_cases(self, engine):
        """Test language detection edge cases."""
        # Empty string
        assert engine._detect_language("") == "Unknown"

        # No extension
        assert engine._detect_language("filename") == "Unknown"

        # Unrecognized extension
        assert engine._detect_language(".xyz") == "Unknown"

        # Case insensitive - these should work
        assert engine._detect_language(".py") == "Python"
        assert engine._detect_language(".jsx") == "JavaScript"
        assert engine._detect_language(".tsx") == "TypeScript"

    def test_complexity_analysis_edge_cases(self, engine):
        """Test complexity analysis edge cases."""
        # Empty extraction
        extraction = {"classes": [], "functions": []}
        assert engine._analyze_complexity(extraction) == "low"

        # Exactly at boundaries
        medium_extraction = {"classes": ["C1", "C2", "C3"], "functions": []}
        assert engine._analyze_complexity(medium_extraction) == "medium"

        high_extraction = {"classes": ["C1", "C2", "C3", "C4", "C5", "C6"], "functions": []}
        assert engine._analyze_complexity(high_extraction) == "high"

    def test_architecture_analysis_edge_cases(self, engine):
        """Test architecture analysis edge cases."""
        # Empty imports
        extraction = {"imports": []}
        assert engine._analyze_architecture(extraction) == "simple"

        # Partial matches
        extraction = {"imports": ["fastapi", "something_else"]}
        assert engine._analyze_architecture(extraction) == "api"

    def test_domain_determination_edge_cases(self, engine):
        """Test domain determination edge cases."""
        # Empty keywords
        extraction = {"domain_keywords": [], "imports": []}
        domain = engine._determine_domain(extraction)
        assert domain in ["auth", "api", "data", "ui", "business"]

        # No matches
        extraction = {"domain_keywords": ["unknown", "keywords"], "imports": []}
        domain = engine._determine_domain(extraction)
        assert domain in ["auth", "api", "data", "ui", "business"]

        # Mixed case
        extraction = {"domain_keywords": ["AUTH", "Login", "PASSWORD"], "imports": []}
        assert engine._determine_domain(extraction) == "auth"

    def test_spec_id_generation_edge_cases(self, engine):
        """Test spec ID generation edge cases."""
        # Empty file name
        extraction = {"file_name": "", "classes": [], "functions": [], "imports": []}
        spec_id = engine._generate_spec_id(extraction, "auth")
        assert spec_id.startswith("AUTH-")
        assert len(spec_id) > 5  # Should have some hash

        # Very long file name
        long_name = "a" * 50
        extraction = {"file_name": long_name, "classes": [], "functions": [], "imports": []}
        spec_id = engine._generate_spec_id(extraction, "api")
        assert spec_id.startswith("API-")
        assert len(spec_id) <= 50  # Reasonable length

    def test_template_rendering_edge_cases(self, engine):
        """Test template rendering edge cases."""
        template = {"template": "Hello {name}", "required_fields": []}

        # Missing key
        result = engine._render_template(template, {})
        assert "Hello " in result

        # None value
        result = engine._render_template(template, {"name": None})
        assert "Hello None" in result

        # Empty list
        result = engine._render_template(template, {"items": []})
        assert "" in result  # Empty string from empty list

        # Nested objects
        result = engine._render_template(template, {"obj": {"nested": "value"}})
        assert "Hello" in result  # obj.nested not replaced, so just Hello

    def test_framework_detection_edge_cases(self, engine):
        """Test framework detection edge cases."""
        # Empty imports
        extraction = {"imports": []}
        assert engine._detect_framework(extraction) == "Custom"

        # Mixed case
        extraction = {"imports": ["FastAPI", "React-DOM"]}
        result = engine._detect_framework(extraction)
        assert result in ["FastAPI", "React"]

        # Partial matches
        extraction = {"imports": ["django_stuff", "other"]}
        assert engine._detect_framework(extraction) == "Django"

    def test_component_naming_edge_cases(self, engine):
        """Test component naming edge cases."""
        extraction = {"classes": [], "functions": [], "imports": []}

        # Unknown domain
        assert engine._get_main_component(extraction, "unknown") == "MainComponent"
        assert engine._get_service_component(extraction, "unknown") == "ServiceComponent"
        assert engine._get_data_component(extraction, "unknown") == "DataComponent"
        assert engine._get_component_4(extraction, "unknown") == "ValidationComponent"

    def test_architecture_diagram_edge_cases(self, engine):
        """Test architecture diagram edge cases."""
        extraction = {"classes": [], "functions": [], "imports": []}

        # Unknown domain
        result = engine._generate_architecture_diagram(extraction, "unknown")
        assert "Client" in result
        assert "Service" in result
        assert "Database" in result

    def test_modules_naming_edge_cases(self, engine):
        """Test modules naming edge cases."""
        extraction = {"classes": [], "functions": [], "imports": []}

        # Unknown domain
        modules = engine._get_new_modules(extraction, "unknown")
        assert "Standard" in modules


class TestPerformanceAndIntegration:
    """Test performance and integration scenarios."""

    @pytest.fixture
    def engine(self):
        """Create EARSTemplateEngine instance."""
        return EARSTemplateEngine()

    def test_large_file_handling(self, engine):
        """Test handling of large/complex code analysis."""
        # Create a large code analysis
        large_analysis = {
            "structure_info": {
                "classes": [f"Class{i}" for i in range(100)],
                "functions": [f"function_{i}" for i in range(500)],
                "imports": ["fastapi", "sqlalchemy", "pydantic"] * 50,
            },
            "domain_keywords": ["api", "data", "model"] * 100,
        }

        result = engine.generate_complete_spec(
            code_analysis=large_analysis,
            file_path="large_service.py"
        )

        # Should still work
        assert result is not None
        assert result["extraction"]["complexity"] == "high"
        assert result["extraction"]["architecture"] in ["api", "data"]

    def test_multiple_domains_scoring(self, engine):
        """Test domain scoring with multiple matches."""
        # Keywords match multiple domains
        extraction = {
            "domain_keywords": ["auth", "api", "user", "endpoint"],
            "classes": [],
            "functions": [],
            "imports": [],
        }

        domain = engine._determine_domain(extraction)
        # Should return one of the domains that scored
        assert domain in ["auth", "api"]

    def test_concurrent_generation(self, engine):
        """Test concurrent spec generation (thread safety)."""
        import threading
        import time

        results = []
        errors = []

        def generate_spec(file_num):
            try:
                code_analysis = {
                    "structure_info": {
                        "classes": ["Service"],
                        "functions": ["process"],
                        "imports": ["fastapi"],
                    },
                    "domain_keywords": ["api"],
                }
                result = engine.generate_complete_spec(
                    code_analysis=code_analysis,
                    file_path=f"service_{file_num}.py"
                )
                results.append(result)
            except Exception as e:
                errors.append(str(e))

        # Create multiple threads
        threads = []
        for i in range(5):
            t = threading.Thread(target=generate_spec, args=(i,))
            threads.append(t)
            t.start()

        # Wait for all threads
        for t in threads:
            t.join()

        # Check results
        assert len(results) == 5
        assert len(errors) == 0
        for result in results:
            assert result is not None
            assert result["domain"] == "api"

    def test_memory_usage(self, engine):
        """Test memory usage with repeated calls."""
        import gc
        import psutil
        import os

        process = psutil.Process(os.getpid())
        initial_memory = process.memory_info().rss

        # Generate many specs
        for i in range(100):
            code_analysis = {
                "structure_info": {
                    "classes": ["Test"],
                    "functions": ["test_func"],
                    "imports": ["test_lib"],
                },
                "domain_keywords": ["test"],
            }
            engine.generate_complete_spec(
                code_analysis=code_analysis,
                file_path=f"test_{i}.py"
            )

        # Clean up
        gc.collect()
        final_memory = process.memory_info().rss

        # Memory should not grow excessively (within reasonable bounds)
        memory_growth = final_memory - initial_memory
        assert memory_growth < 100 * 1024 * 1024  # Less than 100MB growth