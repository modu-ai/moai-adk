---

name: moai-domain-devops
description: CI/CD pipelines, Docker containerization, Kubernetes orchestration, and infrastructure as code. Use when working on DevOps automation scenarios.
allowed-tools:
  - Read
  - Bash
---

# DevOps Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand for platform and CI/CD topics |
| Trigger cues | Infrastructure as code, pipeline design, release automation, observability setup. |
| Tier | 4 |

## What it does

Provides expertise in continuous integration/deployment (CI/CD), Docker containerization, Kubernetes orchestration, and infrastructure as code (IaC) for automated deployment workflows.

## When to use

- Engages when DevOps, CI/CD, or infrastructure automation is required.
- “CI/CD pipeline”, “Docker containerization”, “Kubernetes deployment”, “infrastructure code”
- Automatically invoked when working with DevOps projects
- DevOps SPEC implementation (`/alfred:2-run`)

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
```bash
$ terraform fmt && terraform validate
$ ansible-playbook deploy.yml --check
```

## Inputs
- Domain-specific design documents and user requirements.
- Project technology stack and operational constraints.

## Outputs
- Domain-specific architecture or implementation guidelines.
- Recommended list of associated sub-agents/skills.

## Failure Modes
- When the domain document does not exist or is ambiguous.
- When the project strategy is unconfirmed and cannot be specified.

## Dependencies
- `.moai/project/` document and latest technical briefing are required.

## References
- Google SRE. "Site Reliability Engineering." https://sre.google/books/ (accessed 2025-03-29).
- HashiCorp. "Terraform Best Practices." https://developer.hashicorp.com/terraform/intro (accessed 2025-03-29).

## Changelog
- 2025-03-29: Codified input/output and failure responses for domain skills.

## Works well with

- alfred-trust-validation (deployment validation)
- shell-expert (shell scripting for automation)
- security-expert (secure deployments)

## Best Practices
- Record supporting documentation (version/link) for each domain decision.
- Review performance, security, and operational requirements simultaneously at an early stage.
