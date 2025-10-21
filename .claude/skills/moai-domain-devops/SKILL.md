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
- 도메인 관련 설계 문서 및 사용자 요구사항.
- 프로젝트 기술 스택 및 운영 제약.

## Outputs
- 도메인 특화 아키텍처 또는 구현 가이드라인.
- 연관 서브 에이전트/스킬 권장 목록.

## Failure Modes
- 도메인 근거 문서가 없거나 모호할 때.
- 프로젝트 전략이 미확정이라 구체화할 수 없을 때.

## Dependencies
- `.moai/project/` 문서와 최신 기술 브리핑이 필요합니다.

## References
- Google SRE. "Site Reliability Engineering." https://sre.google/books/ (accessed 2025-03-29).
- HashiCorp. "Terraform Best Practices." https://developer.hashicorp.com/terraform/intro (accessed 2025-03-29).

## Changelog
- 2025-03-29: 도메인 스킬에 대한 입력/출력 및 실패 대응을 명문화했습니다.

## Works well with

- alfred-trust-validation (deployment validation)
- shell-expert (shell scripting for automation)
- security-expert (secure deployments)

## Best Practices
- 도메인 결정 사항마다 근거 문서(버전/링크)를 기록합니다.
- 성능·보안·운영 요구사항을 초기 단계에서 동시에 검토하세요.
