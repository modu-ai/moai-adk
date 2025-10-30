"""
Backend Plugin Commands - /init-fastapi, /db-setup, /resource-crud implementations

@CODE:BACKEND-INIT-CMD-001:COMMANDS
"""

from dataclasses import dataclass
from pathlib import Path
from typing import Optional, List
from datetime import datetime


# @CODE:BACKEND-COMMAND-RESULT-001:RESULT
@dataclass
class CommandResult:
    """Result object for command execution"""
    success: bool
    project_dir: Path
    files_created: List[str]
    message: str
    error: Optional[str] = None


class InitFastAPICommand:
    """
    /init-fastapi command implementation

    Initializes FastAPI project with standard structure
    """

    VALID_FRAMEWORKS = ["fastapi"]
    VALID_DATABASES = ["postgresql", "mysql", "sqlite", "mongodb"]
    MIN_PROJECT_NAME_LENGTH = 3
    MAX_PROJECT_NAME_LENGTH = 50

    def validate_project_name(self, project_name: str) -> bool:
        """
        Validate project name format

        @CODE:BACKEND-FASTAPI-VALIDATE-NAME-001:VALIDATION
        """
        if not project_name:
            raise ValueError("Project name cannot be empty")

        if len(project_name) < self.MIN_PROJECT_NAME_LENGTH:
            raise ValueError(
                f"Project name must be at least {self.MIN_PROJECT_NAME_LENGTH} characters"
            )

        if len(project_name) > self.MAX_PROJECT_NAME_LENGTH:
            raise ValueError(
                f"Project name cannot exceed {self.MAX_PROJECT_NAME_LENGTH} characters"
            )

        if not all(c.islower() or c.isdigit() or c == "-" for c in project_name):
            raise ValueError(
                "Project name must contain only lowercase letters, numbers, and hyphens"
            )

        if project_name.startswith("-") or project_name.endswith("-"):
            raise ValueError("Project name cannot start or end with a hyphen")

        if "--" in project_name:
            raise ValueError("Project name cannot contain consecutive hyphens")

        return True

    def validate_database(self, database: Optional[str]) -> bool:
        """
        Validate database type

        @CODE:BACKEND-FASTAPI-VALIDATE-DB-001:VALIDATION
        """
        if database is not None and database not in self.VALID_DATABASES:
            raise ValueError(
                f"Invalid database: {database}\nSupported: {', '.join(self.VALID_DATABASES)}"
            )
        return True

    def create_project_directory(self, output_dir: Path, project_name: str) -> Path:
        """
        Create FastAPI project directory

        @CODE:BACKEND-FASTAPI-CREATE-DIR-001:DIRECTORY
        """
        project_dir = output_dir / project_name

        if project_dir.exists():
            raise FileExistsError(f"Project already exists: {project_dir}")

        project_dir.mkdir(parents=True, exist_ok=False)
        return project_dir

    def create_fastapi_structure(
        self,
        project_dir: Path,
        include_db: bool = False,
        database: Optional[str] = None,
        include_auth: bool = False,
        include_cors: bool = False
    ) -> List[str]:
        """
        Create FastAPI project structure

        @CODE:BACKEND-FASTAPI-STRUCTURE-001:STRUCTURE
        """
        files_created = []

        # Create app directory
        app_dir = project_dir / "app"
        app_dir.mkdir(exist_ok=False)

        # Create __init__.py
        (app_dir / "__init__.py").touch()
        files_created.append(str((app_dir / "__init__.py").relative_to(project_dir.parent)))

        # Create main.py
        main_content = """from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

app = FastAPI(
    title="API",
    version="1.0.0",
    description="FastAPI Application"
)

# CORS Configuration
origins = [
    "http://localhost",
    "http://localhost:3000",
    "http://localhost:8000",
]

app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

@app.get("/")
async def root():
    return {"message": "FastAPI Application"}

@app.get("/health")
async def health():
    return {"status": "healthy"}
"""
        main_file = project_dir / "main.py"
        main_file.write_text(main_content)
        files_created.append(str(main_file.relative_to(project_dir.parent)))

        # Create requirements.txt
        requirements = """fastapi==0.104.1
uvicorn==0.24.0
pydantic==2.5.0
python-multipart==0.0.6
"""
        if include_db and database:
            if database == "postgresql":
                requirements += "sqlalchemy==2.0.23\npsycopg2-binary==2.9.9\nalembic==1.13.0\n"
            elif database == "mysql":
                requirements += "sqlalchemy==2.0.23\nmysql-connector-python==8.2.0\nalembic==1.13.0\n"
            elif database == "mongodb":
                requirements += "motor==3.3.2\npymongo==4.6.0\n"

        if include_auth:
            requirements += "python-jose==3.3.0\npasslib==1.7.4\nbcrypt==4.1.1\npyjwt==2.8.1\n"

        req_file = project_dir / "requirements.txt"
        req_file.write_text(requirements)
        files_created.append(str(req_file.relative_to(project_dir.parent)))

        # Create .env.example if database
        if include_db and database:
            env_content = f"""# Database Configuration
DATABASE_URL={self._get_database_url(database)}

# API Configuration
API_TITLE=API
API_VERSION=1.0.0

# Server
HOST=0.0.0.0
PORT=8000
"""
            env_file = project_dir / ".env.example"
            env_file.write_text(env_content)
            files_created.append(str(env_file.relative_to(project_dir.parent)))

            # Create database.py
            db_content = """\"\"\"
Database Configuration
\"\"\"

from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker, declarative_base
import os

DATABASE_URL = os.getenv("DATABASE_URL", "sqlite:///./test.db")

engine = create_engine(
    DATABASE_URL,
    connect_args={"check_same_thread": False} if "sqlite" in DATABASE_URL else {}
)

SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)

Base = declarative_base()

def get_db():
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()
"""
            db_file = app_dir / "database.py"
            db_file.write_text(db_content)
            files_created.append(str(db_file.relative_to(project_dir.parent)))

        return files_created

    def _get_database_url(self, database: str) -> str:
        """Get database URL template"""
        urls = {
            "postgresql": "postgresql://user:password@localhost/dbname",
            "mysql": "mysql://user:password@localhost/dbname",
            "sqlite": "sqlite:///./test.db",
            "mongodb": "mongodb://localhost:27017/dbname"
        }
        return urls.get(database, "")

    def execute(
        self,
        project_name: str,
        output_dir: Path,
        include_db: bool = False,
        database: Optional[str] = None,
        include_auth: bool = False,
        include_cors: bool = False
    ) -> CommandResult:
        """
        Execute /init-fastapi command

        @CODE:BACKEND-FASTAPI-EXECUTE-001:MAIN
        """
        # Validation
        self.validate_project_name(project_name)
        self.validate_database(database)

        try:
            # Create project directory
            project_dir = self.create_project_directory(Path(output_dir), project_name)

            # Create structure
            files_created = self.create_fastapi_structure(
                project_dir,
                include_db=include_db,
                database=database,
                include_auth=include_auth,
                include_cors=include_cors
            )

            message = f"✅ FastAPI project '{project_name}' initialized successfully\n"
            message += f"📁 Location: {project_dir}\n"
            message += f"📦 Files created: {len(files_created)}"

            return CommandResult(
                success=True,
                project_dir=project_dir,
                files_created=files_created,
                message=message
            )

        except FileExistsError:
            raise
        except Exception as e:
            return CommandResult(
                success=False,
                project_dir=None,
                files_created=[],
                message=f"❌ Error initializing FastAPI project",
                error=str(e)
            )


class DBSetupCommand:
    """
    /db-setup command implementation

    Configures database for FastAPI project
    """

    VALID_DATABASES = ["postgresql", "mysql", "sqlite", "mongodb"]

    def validate_database(self, database: str) -> bool:
        """
        Validate database type

        @CODE:BACKEND-DB-VALIDATE-001:VALIDATION
        """
        if database not in self.VALID_DATABASES:
            raise ValueError(
                f"Invalid database: {database}\nSupported: {', '.join(self.VALID_DATABASES)}"
            )
        return True

    def execute(
        self,
        project_name: str,
        output_dir: Path,
        database: str
    ) -> CommandResult:
        """
        Execute /db-setup command

        @CODE:BACKEND-DB-EXECUTE-001:MAIN
        """
        self.validate_database(database)

        try:
            project_dir = output_dir / project_name

            files_created = []

            # Create .env.example
            env_file = project_dir / ".env.example"
            if not env_file.exists():
                env_file.write_text(f"DATABASE_URL=\n")
                files_created.append(str(env_file.relative_to(output_dir)))

            message = f"✅ Database setup for {database} completed\n"
            message += f"📁 Location: {project_dir}"

            return CommandResult(
                success=True,
                project_dir=project_dir,
                files_created=files_created,
                message=message
            )

        except Exception as e:
            return CommandResult(
                success=False,
                project_dir=None,
                files_created=[],
                message=f"❌ Error setting up database",
                error=str(e)
            )


class ResourceCRUDCommand:
    """
    /resource-crud command implementation

    Generates CRUD routes and models for a resource
    """

    def validate_resource_name(self, resource_name: str) -> bool:
        """
        Validate resource name

        @CODE:BACKEND-RESOURCE-VALIDATE-001:VALIDATION
        """
        if not resource_name:
            raise ValueError("Resource name cannot be empty")

        if not resource_name.islower() or " " in resource_name:
            raise ValueError(
                "Resource name must be lowercase without spaces"
            )

        return True

    def execute(
        self,
        project_name: str,
        resource_name: str,
        output_dir: Path
    ) -> CommandResult:
        """
        Execute /resource-crud command

        @CODE:BACKEND-RESOURCE-EXECUTE-001:MAIN
        """
        self.validate_resource_name(resource_name)

        try:
            project_dir = output_dir / project_name
            app_dir = project_dir / "app"
            resources_dir = app_dir / "resources"
            resources_dir.mkdir(parents=True, exist_ok=True)

            files_created = []

            # Create resource routes
            routes_content = f"""\"\"\"
{resource_name.title()} API Routes
\"\"\"

from fastapi import APIRouter, HTTPException
from typing import List, Optional

router = APIRouter(
    prefix="/{resource_name}",
    tags=["{resource_name}"]
)

@router.get("/")
async def list_{resource_name}():
    \"\"\"List all {resource_name}\"\"\"
    return {{"items": []}}

@router.get("/{{id}}")
async def get_{resource_name}(id: int):
    \"\"\"Get {resource_name} by ID\"\"\"
    return {{"id": id, "name": "example"}}

@router.post("/")
async def create_{resource_name}():
    \"\"\"Create new {resource_name}\"\"\"
    return {{"id": 1, "name": "created"}}

@router.put("/{{id}}")
async def update_{resource_name}(id: int):
    \"\"\"Update {resource_name}\"\"\"
    return {{"id": id, "name": "updated"}}

@router.delete("/{{id}}")
async def delete_{resource_name}(id: int):
    \"\"\"Delete {resource_name}\"\"\"
    return {{"status": "deleted"}}
"""
            routes_file = resources_dir / f"{resource_name}_routes.py"
            routes_file.write_text(routes_content)
            files_created.append(str(routes_file.relative_to(output_dir)))

            # Create resource model
            model_content = f"""\"\"\"
{resource_name.title()} Data Model
\"\"\"

from pydantic import BaseModel
from typing import Optional
from datetime import datetime

class {resource_name.title()}Base(BaseModel):
    name: str
    description: Optional[str] = None

class {resource_name.title()}Create({resource_name.title()}Base):
    pass

class {resource_name.title()}Update({resource_name.title()}Base):
    pass

class {resource_name.title()}(BaseModel):
    id: int
    name: str
    description: Optional[str] = None
    created_at: datetime

    class Config:
        from_attributes = True
"""
            model_file = resources_dir / f"{resource_name}_model.py"
            model_file.write_text(model_content)
            files_created.append(str(model_file.relative_to(output_dir)))

            message = f"✅ CRUD resources for '{resource_name}' generated\n"
            message += f"📁 Location: {resources_dir}\n"
            message += f"📝 Files created: {len(files_created)}"

            return CommandResult(
                success=True,
                project_dir=resources_dir,
                files_created=files_created,
                message=message
            )

        except Exception as e:
            return CommandResult(
                success=False,
                project_dir=None,
                files_created=[],
                message=f"❌ Error generating CRUD resources",
                error=str(e)
            )


# Create module-level command instances
init_fastapi = InitFastAPICommand()
db_setup = DBSetupCommand()
resource_crud = ResourceCRUDCommand()
