# Skill Factory - 실전 예제

## 예제 1: 기본 스킬 생성

```python
# factory/basic_skill_factory.py
from moai_cc_skill_factory import SkillFactory

async def create_basic_skill():
    """기본 스킬 생성"""
    
    factory = SkillFactory()
    
    skill = await factory.create(
        name="moai-example-skill",
        description="예제 스킬",
        category="example",
        templates={
            "SKILL.md": "skill_template.md",
            "examples.md": "examples_template.md",
            "reference.md": "reference_template.md"
        }
    )
    
    return skill
```

## 예제 2: 언어별 스킬 생성

```python
# factory/language_skill_factory.py
from moai_cc_skill_factory import LanguageSkillFactory

async def create_language_skills():
    """언어별 스킬 생성"""
    
    factory = LanguageSkillFactory()
    
    skills = {}
    for lang in ["python", "javascript", "rust", "go"]:
        skill = await factory.create_for_language(
            language=lang,
            include_advanced_patterns=True,
            include_optimization=True
        )
        skills[lang] = skill
    
    return skills
```

## 예제 3: 도메인 스킬 생성

```python
# factory/domain_skill_factory.py
from moai_cc_skill_factory import DomainSkillFactory

async def create_domain_skills():
    """도메인별 스킬 생성"""
    
    factory = DomainSkillFactory()
    
    domains = {
        "backend": {
            "frameworks": ["FastAPI", "Django", "Spring"],
            "patterns": ["REST API", "GraphQL", "gRPC"]
        },
        "frontend": {
            "frameworks": ["React", "Vue", "Angular"],
            "patterns": ["Component", "State Management", "Routing"]
        },
        "devops": {
            "tools": ["Docker", "Kubernetes", "Terraform"],
            "patterns": ["CI/CD", "Infrastructure as Code"]
        }
    }
    
    skills = {}
    for domain, config in domains.items():
        skill = await factory.create_for_domain(
            domain=domain,
            config=config
        )
        skills[domain] = skill
    
    return skills
```

## 예제 4: 고급 패턴 및 최적화 포함

```python
# factory/advanced_skill_factory.py
from moai_cc_skill_factory import AdvancedSkillFactory

async def create_advanced_skill():
    """고급 패턴 포함 스킬"""
    
    factory = AdvancedSkillFactory()
    
    skill = await factory.create(
        name="moai-advanced-example",
        include_modules={
            "advanced-patterns.md": {
                "sections": [
                    "Enterprise Patterns",
                    "Distributed Systems",
                    "Performance Optimization",
                    "Security Patterns"
                ]
            },
            "optimization.md": {
                "sections": [
                    "Performance Tuning",
                    "Memory Management",
                    "Caching Strategies",
                    "Scalability Patterns"
                ]
            }
        }
    )
    
    return skill
```

## 예제 5: 배치 스킬 생성

```python
# factory/batch_skill_factory.py
from moai_cc_skill_factory import BatchSkillFactory

async def create_multiple_skills():
    """여러 스킬 배치 생성"""
    
    factory = BatchSkillFactory()
    
    skills_config = [
        {
            "name": "moai-lang-python",
            "type": "language",
            "config": {"version": "3.12"}
        },
        {
            "name": "moai-domain-backend",
            "type": "domain",
            "config": {"frameworks": ["FastAPI", "Django"]}
        },
        {
            "name": "moai-essentials-debug",
            "type": "essentials",
            "config": {"include_ai": True}
        }
    ]
    
    # 배치 생성
    skills = await factory.create_batch(
        configs=skills_config,
        parallel=True,
        validate=True
    )
    
    return skills
```

## 예제 6: 스킬 템플릿 커스터마이징

```python
# factory/custom_template_factory.py
from moai_cc_skill_factory import SkillFactory

async def create_custom_skill():
    """커스텀 템플릿으로 스킬 생성"""
    
    factory = SkillFactory()
    
    # 커스텀 템플릿 정의
    custom_templates = {
        "SKILL.md": """
# {skill_name}

**설명**: {description}

## 기능
- {feature1}
- {feature2}
- {feature3}

## 사용 사례
{use_cases}
""",
        "examples.md": """
# {skill_name} 실전 예제

{examples}
""",
        "reference.md": """
# {skill_name} 참고 자료

## API Reference
{api_reference}
"""
    }
    
    skill = await factory.create(
        name="moai-custom-skill",
        description="커스텀 스킬",
        templates=custom_templates,
        variables={
            "feature1": "기능 1",
            "feature2": "기능 2",
            "feature3": "기능 3",
            "use_cases": "사용 사례들"
        }
    )
    
    return skill
```

## 예제 7: 스킬 검증 및 배포

```python
# factory/deployment_factory.py
from moai_cc_skill_factory import SkillDeploymentFactory

async def deploy_skill():
    """스킬 검증 및 배포"""
    
    factory = SkillDeploymentFactory()
    
    skill_config = {
        "name": "moai-new-skill",
        "version": "1.0.0",
        "author": "Developer Name"
    }
    
    # 검증
    validation = await factory.validate(skill_config)
    
    if not validation.passed:
        return {
            "status": "validation_failed",
            "errors": validation.errors
        }
    
    # 배포
    deployment = await factory.deploy(skill_config)
    
    return {
        "status": "deployed",
        "skill_id": deployment.skill_id,
        "registry_url": deployment.registry_url
    }
```

---

**Last Updated**: 2025-11-22  
**Status**: Production Ready
