---
name: moai-domain-devops
description: CI/CD pipelines, Docker containerization, Kubernetes orchestration, and
  infrastructure as code
allowed-tools:
- Read
- Bash
---

# DevOps Expert

## What it does

Provides expertise in continuous integration/deployment (CI/CD), Docker containerization, Kubernetes orchestration, and infrastructure as code (IaC) for automated deployment workflows.

## When to use

- "CI/CD 파이프라인", "Docker 컨테이너화", "Kubernetes 배포", "인프라 코드"
- Automatically invoked when working with DevOps projects
- DevOps SPEC implementation (`/alfred:2-build`)

## How it works

**CI/CD Pipelines**:
- **GitHub Actions**: Workflow automation (.github/workflows)
- **GitLab CI**: .gitlab-ci.yml configuration
- **Jenkins**: Pipeline as code (Jenkinsfile)
- **CircleCI**: .circleci/config.yml
- **Pipeline stages**: Build → Test → Deploy

**Docker Containerization**:
- **Dockerfile**: Multi-stage builds for optimization
- **docker-compose**: Local development environments
- **Image optimization**: Layer caching, alpine base images
- **Container registries**: Docker Hub, GitHub Container Registry

**Kubernetes Orchestration**:
- **Deployments**: Rolling updates, rollbacks
- **Services**: LoadBalancer, ClusterIP, NodePort
- **ConfigMaps/Secrets**: Configuration management
- **Helm charts**: Package management
- **Ingress**: Traffic routing

**Infrastructure as Code (IaC)**:
- **Terraform**: Cloud-agnostic provisioning
- **Ansible**: Configuration management
- **CloudFormation**: AWS-specific IaC
- **Pulumi**: Programmatic infrastructure

**Monitoring & Logging**:
- **Prometheus**: Metrics collection
- **Grafana**: Visualization
- **ELK Stack**: Logging (Elasticsearch, Logstash, Kibana)

## Examples

### Example 1: GitHub Actions CI/CD
User: "/alfred:2-build CICD-001"
Claude: (creates RED workflow test, GREEN GitHub Actions workflow, REFACTOR)

### Example 2: Kubernetes deployment
User: "Kubernetes 배포 설정"
Claude: (creates deployment.yaml, service.yaml, ingress.yaml)

## Works well with

- alfred-trust-validation (deployment validation)
- shell-expert (shell scripting for automation)
- security-expert (secure deployments)
