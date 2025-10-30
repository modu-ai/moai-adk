"""
DevOps Plugin Commands - Docker, CI/CD, Kubernetes setup implementations

@CODE:DEVOPS-INIT-CMD-001:COMMANDS
"""

from dataclasses import dataclass
from pathlib import Path
from typing import Optional, List
import yaml
from datetime import datetime


# @CODE:DEVOPS-COMMAND-RESULT-001:RESULT
@dataclass
class CommandResult:
    """Result object for command execution"""
    success: bool
    config_dir: Path
    files_created: List[str]
    message: str
    error: Optional[str] = None


class SetupDockerCommand:
    """
    /setup-docker command implementation

    Configures Docker containerization for projects
    """

    VALID_APP_TYPES = ["python", "nodejs", "go", "java"]

    def validate_app_type(self, app_type: str) -> bool:
        """
        Validate application type

        @CODE:DEVOPS-VALIDATE-APP-TYPE-001:VALIDATION
        """
        if app_type not in self.VALID_APP_TYPES:
            raise ValueError(
                f"Invalid app type: {app_type}\nSupported: {', '.join(self.VALID_APP_TYPES)}"
            )
        return True

    def create_dockerfile(self, project_dir: Path, app_type: str) -> str:
        """
        Create Dockerfile based on app type

        @CODE:DEVOPS-DOCKERFILE-001:DOCKER
        """
        dockerfiles = {
            "python": """FROM python:3.11-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

EXPOSE 8000
CMD ["python", "-m", "uvicorn", "main:app", "--host", "0.0.0.0"]
""",
            "nodejs": """FROM node:20-alpine

WORKDIR /app

COPY package*.json ./
RUN npm ci --only=production

COPY . .

EXPOSE 3000
CMD ["npm", "start"]
""",
            "go": """FROM golang:1.21-alpine as builder

WORKDIR /app
COPY . .
RUN go build -o app .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/app .

EXPOSE 8080
CMD ["./app"]
""",
            "java": """FROM openjdk:21-slim

WORKDIR /app
COPY target/*.jar app.jar

EXPOSE 8080
ENTRYPOINT ["java", "-jar", "app.jar"]
"""
        }
        return dockerfiles.get(app_type, dockerfiles["python"])

    def execute(
        self,
        project_name: str,
        output_dir: Path,
        app_type: str = "python",
        include_compose: bool = False
    ) -> CommandResult:
        """
        Execute /setup-docker command

        @CODE:DEVOPS-DOCKER-EXECUTE-001:MAIN
        """
        self.validate_app_type(app_type)

        try:
            project_dir = output_dir / project_name
            project_dir.mkdir(parents=True, exist_ok=True)

            files_created = []

            # Create Dockerfile
            dockerfile_content = self.create_dockerfile(project_dir, app_type)
            dockerfile = project_dir / "Dockerfile"
            dockerfile.write_text(dockerfile_content)
            files_created.append(str(dockerfile.relative_to(output_dir)))

            # Create .dockerignore
            dockerignore = """node_modules
npm-debug.log
.git
.gitignore
.env
.DS_Store
__pycache__
*.pyc
.venv
dist
build
"""
            (project_dir / ".dockerignore").write_text(dockerignore)
            files_created.append(str((project_dir / ".dockerignore").relative_to(output_dir)))

            # Create docker-compose if requested
            if include_compose:
                compose_config = {
                    "version": "3.8",
                    "services": {
                        "app": {
                            "build": ".",
                            "ports": ["8000:8000" if app_type == "python" else "3000:3000"],
                            "environment": {
                                "ENV": "development"
                            }
                        }
                    }
                }
                compose_file = project_dir / "docker-compose.yml"
                with open(compose_file, "w") as f:
                    yaml.dump(compose_config, f)
                files_created.append(str(compose_file.relative_to(output_dir)))

            message = f"✅ Docker configuration for {app_type} created\n"
            message += f"📁 Location: {project_dir}\n"
            message += f"📝 Files created: {len(files_created)}"

            return CommandResult(
                success=True,
                config_dir=project_dir,
                files_created=files_created,
                message=message
            )

        except Exception as e:
            return CommandResult(
                success=False,
                config_dir=None,
                files_created=[],
                message=f"❌ Error setting up Docker",
                error=str(e)
            )


class SetupCICommand:
    """
    /setup-ci command implementation

    Configures CI/CD pipelines
    """

    VALID_PLATFORMS = ["github-actions", "gitlab-ci", "circleci"]

    def validate_platform(self, platform: str) -> bool:
        """
        Validate CI platform

        @CODE:DEVOPS-VALIDATE-CI-001:VALIDATION
        """
        if platform not in self.VALID_PLATFORMS:
            raise ValueError(
                f"Invalid platform: {platform}\nSupported: {', '.join(self.VALID_PLATFORMS)}"
            )
        return True

    def create_github_workflow(self, project_dir: Path, include_tests: bool) -> Path:
        """
        Create GitHub Actions workflow

        @CODE:DEVOPS-GITHUB-WORKFLOW-001:CI
        """
        github_dir = project_dir / ".github" / "workflows"
        github_dir.mkdir(parents=True, exist_ok=True)

        workflow = {
            "name": "CI",
            "on": {
                "push": {"branches": ["main", "develop"]},
                "pull_request": {"branches": ["main"]}
            },
            "jobs": {
                "build": {
                    "runs-on": "ubuntu-latest",
                    "steps": [
                        {"uses": "actions/checkout@v4"},
                        {"uses": "actions/setup-python@v4", "with": {"python-version": "3.11"}},
                        {"run": "pip install -r requirements.txt"}
                    ]
                }
            }
        }

        if include_tests:
            workflow["jobs"]["build"]["steps"].append({"run": "pytest"})

        workflow_file = github_dir / "ci.yml"
        with open(workflow_file, "w") as f:
            yaml.dump(workflow, f)

        return workflow_file

    def execute(
        self,
        project_name: str,
        output_dir: Path,
        ci_platform: str = "github-actions",
        include_tests: bool = False
    ) -> CommandResult:
        """
        Execute /setup-ci command

        @CODE:DEVOPS-CI-EXECUTE-001:MAIN
        """
        self.validate_platform(ci_platform)

        try:
            project_dir = output_dir / project_name
            project_dir.mkdir(parents=True, exist_ok=True)

            files_created = []

            if ci_platform == "github-actions":
                workflow_file = self.create_github_workflow(project_dir, include_tests)
                files_created.append(str(workflow_file.relative_to(output_dir)))

            elif ci_platform == "gitlab-ci":
                gitlab_config = """stages:
  - build
  - test
  - deploy

build:
  stage: build
  script:
    - echo "Building..."

test:
  stage: test
  script:
    - echo "Testing..."

deploy:
  stage: deploy
  script:
    - echo "Deploying..."
"""
                gitlab_file = project_dir / ".gitlab-ci.yml"
                gitlab_file.write_text(gitlab_config)
                files_created.append(str(gitlab_file.relative_to(output_dir)))

            message = f"✅ CI/CD setup with {ci_platform} completed\n"
            message += f"📁 Location: {project_dir}\n"
            message += f"📝 Files created: {len(files_created)}"

            return CommandResult(
                success=True,
                config_dir=project_dir,
                files_created=files_created,
                message=message
            )

        except Exception as e:
            return CommandResult(
                success=False,
                config_dir=None,
                files_created=[],
                message=f"❌ Error setting up CI/CD",
                error=str(e)
            )


class SetupK8sCommand:
    """
    /setup-k8s command implementation

    Configures Kubernetes manifests
    """

    def validate_replicas(self, replicas: int) -> bool:
        """
        Validate replica count

        @CODE:DEVOPS-VALIDATE-REPLICAS-001:VALIDATION
        """
        if replicas < 1:
            raise ValueError("Replicas must be at least 1")
        return True

    def create_k8s_manifests(
        self,
        project_dir: Path,
        app_name: str,
        replicas: int,
        include_ingress: bool = False
    ) -> List[str]:
        """
        Create Kubernetes manifests

        @CODE:DEVOPS-K8S-MANIFESTS-001:K8S
        """
        k8s_dir = project_dir / "k8s"
        k8s_dir.mkdir(parents=True, exist_ok=True)

        files_created = []

        # Deployment
        deployment = {
            "apiVersion": "apps/v1",
            "kind": "Deployment",
            "metadata": {"name": app_name},
            "spec": {
                "replicas": replicas,
                "selector": {"matchLabels": {"app": app_name}},
                "template": {
                    "metadata": {"labels": {"app": app_name}},
                    "spec": {
                        "containers": [{
                            "name": app_name,
                            "image": f"{app_name}:latest",
                            "ports": [{"containerPort": 8000}]
                        }]
                    }
                }
            }
        }

        deployment_file = k8s_dir / "deployment.yaml"
        with open(deployment_file, "w") as f:
            yaml.dump(deployment, f)
        files_created.append(str(deployment_file.relative_to(project_dir.parent)))

        # Service
        service = {
            "apiVersion": "v1",
            "kind": "Service",
            "metadata": {"name": f"{app_name}-service"},
            "spec": {
                "selector": {"app": app_name},
                "ports": [{"port": 80, "targetPort": 8000}],
                "type": "LoadBalancer"
            }
        }

        service_file = k8s_dir / "service.yaml"
        with open(service_file, "w") as f:
            yaml.dump(service, f)
        files_created.append(str(service_file.relative_to(project_dir.parent)))

        # Ingress if requested
        if include_ingress:
            ingress = {
                "apiVersion": "networking.k8s.io/v1",
                "kind": "Ingress",
                "metadata": {"name": f"{app_name}-ingress"},
                "spec": {
                    "rules": [{
                        "host": f"{app_name}.example.com",
                        "http": {
                            "paths": [{
                                "path": "/",
                                "pathType": "Prefix",
                                "backend": {
                                    "service": {
                                        "name": f"{app_name}-service",
                                        "port": {"number": 80}
                                    }
                                }
                            }]
                        }
                    }]
                }
            }

            ingress_file = k8s_dir / "ingress.yaml"
            with open(ingress_file, "w") as f:
                yaml.dump(ingress, f)
            files_created.append(str(ingress_file.relative_to(project_dir.parent)))

        return files_created

    def execute(
        self,
        project_name: str,
        output_dir: Path,
        replicas: int = 3,
        include_ingress: bool = False
    ) -> CommandResult:
        """
        Execute /setup-k8s command

        @CODE:DEVOPS-K8S-EXECUTE-001:MAIN
        """
        self.validate_replicas(replicas)

        try:
            project_dir = output_dir / project_name
            project_dir.mkdir(parents=True, exist_ok=True)

            files_created = self.create_k8s_manifests(
                project_dir,
                project_name,
                replicas,
                include_ingress
            )

            message = f"✅ Kubernetes setup completed\n"
            message += f"📁 Location: {project_dir / 'k8s'}\n"
            message += f"📝 Files created: {len(files_created)}\n"
            message += f"🔄 Replicas: {replicas}"

            return CommandResult(
                success=True,
                config_dir=project_dir / "k8s",
                files_created=files_created,
                message=message
            )

        except Exception as e:
            return CommandResult(
                success=False,
                config_dir=None,
                files_created=[],
                message=f"❌ Error setting up Kubernetes",
                error=str(e)
            )


# Create module-level command instances
setup_docker = SetupDockerCommand()
setup_ci = SetupCICommand()
setup_k8s = SetupK8sCommand()
