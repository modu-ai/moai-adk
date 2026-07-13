---
title: 开发方法论 (DDD/TDD)
weight: 50
draft: false
---

本文详细介绍 MoAI-ADK 的开发方法论。它是 Run 阶段智能体实现代码时遵循的纪律，根据项目状态选用 TDD 或 DDD。方法论明确了，智能体就不会迷路 — 测试本身就是完成条件，循环自行收敛，不在无谓的重试上浪费 token。

{{< callout type="info" >}}
  **一句话总结：** 新项目使用 **TDD**（RED-GREEN-REFACTOR），几乎没有测试的现有项目使用 **DDD**（ANALYZE-PRESERVE-IMPROVE）。也可以在 `quality.yaml` 中手动选择。
{{< /callout >}}

## 方法论概览

MoAI-ADK 会根据项目状态自动选择最优开发方法论。

```mermaid
flowchart TD
    A["项目分析"] --> B{"新项目？"}
    B -->|"是"| C["TDD\nRED-GREEN-REFACTOR"]
    B -->|"否"| D{"测试覆盖率？"}
    D -->|"10% 以上"| C
    D -->|"低于 10%"| E["DDD\nANALYZE-PRESERVE-IMPROVE"]

    style C fill:#4CAF50,color:#fff
    style E fill:#2196F3,color:#fff
```

| 项目类型 | 方法论 | 循环 | 说明 |
| ---------------------------------- | ------- | ------------------------- | ---------------------------------------- |
| **新项目** | **TDD** | RED-GREEN-REFACTOR | 先写测试再实现 |
| **现有项目**（覆盖率 ≥ 10%） | **TDD** | RED-GREEN-REFACTOR | 基于部分测试扩展 TDD |
| **现有项目**（覆盖率 < 10%） | **DDD** | ANALYZE-PRESERVE-IMPROVE | 用特性测试安全地渐进改进 |

{{< callout type="info" >}}
  **方法论可以手动选择：** 在 `.moai/config/sections/quality.yaml` 中把 `development_mode` 设为 `tdd` 或 `ddd`，即可忽略自动选择，使用想要的方法论。
{{< /callout >}}

## 什么是 TDD？

**TDD** (Test-Driven Development) 是**先写测试，再实现通过该测试的最少代码**的开发方法论。它是 MoAI-ADK 的默认方法论，大多数项目都在使用。

### RED-GREEN-REFACTOR 循环

TDD 以重复三个阶段的循环推进。

```mermaid
flowchart TD
    A["RED\n编写失败的测试"] --> B["GREEN\n用最少代码通过测试"]
    B --> C["REFACTOR\n改进代码质量\n测试持续通过"]
    C --> D{"全部需求\n实现完成？"}
    D -->|"否"| A
    D -->|"是"| E["确认测试覆盖率 85%+"]
```

### 第 1 步：RED（编写失败的测试）

**先写**要实现功能的测试。代码还不存在，测试必然失败。

**核心原则：**

- 一次只写一个测试
- 用 Given-When-Then 明确描述要实现的行为
- 确认测试失败（不失败的测试没有意义）

### 第 2 步：GREEN（用最少代码通过测试）

编写通过测试的**最简单代码**。

**核心原则：**

- 不提前优化或抽象
- 专注正确性，优雅留到之后
- 测试通过就停手

### 第 3 步：REFACTOR（改进代码质量）

在保持测试通过的状态下整理代码。

**核心原则：**

- 消除重复代码
- 改进变量名、函数名
- 应用 SOLID 原则
- 测试必须持续通过

### TDD 实战示例

```python
# RED: 실패하는 테스트 먼저 작성
def test_user_registration():
    """
    GIVEN: 유효한 사용자 정보가 있고
    WHEN: 회원가입을 하면
    THEN: 사용자가 생성되고 환영 이메일이 발송되어야 함
    """
    user_service = UserService()
    result = user_service.register(
        email="newuser@example.com",
        password="SecurePass123!"
    )

    assert result.success is True
    assert result.user.id is not None
    assert email_service.welcome_email_sent("newuser@example.com") is True

# 테스트 실행 (실패 예상 - 구현 안 됨)
# > pytest test_user_service.py - test_user_registration FAILED

# ====================================

# GREEN: 최소한의 코드로 테스트 통과
class UserService:
    def register(self, email: str, password: str) -> RegistrationResult:
        user = User.create(email, password)
        user_repository.save(user)
        email_service.send_welcome(email)
        return RegistrationResult.success(user)

# 테스트 실행 (통과)
# > pytest test_user_service.py - test_user_registration PASSED

# ====================================

# REFACTOR: 코드 품질 개선 (테스트는 계속 통과)
class UserService:
    def __init__(
        self,
        user_repo: UserRepository,
        email_service: EmailService,
        password_validator: PasswordValidator
    ):
        self.user_repo = user_repo
        self.email_service = email_service
        self.password_validator = password_validator

    def register(self, email: str, password: str) -> RegistrationResult:
        if not self.password_validator.validate(password):
            return RegistrationResult.failure("비밀번호가 유효하지 않습니다")

        user = User.create(email, password)
        self.user_repo.save(user)
        self.email_service.send_welcome(email)
        return RegistrationResult.success(user)

# 테스트 실행 (여전히 통과)
# > pytest test_user_service.py - test_user_registration PASSED
```

### 在现有项目中使用 TDD (Brownfield Enhancement)

在已有代码的项目中使用 TDD 时，会增加 **Pre-RED 阶段**：

1. **(Pre-RED)** 阅读目标区域的现有代码，理解当前行为
2. **RED：** 基于对现有代码的理解编写失败的测试
3. **GREEN：** 用最少代码让测试通过
4. **REFACTOR：** 保持测试通过的同时改进代码

{{< callout type="info" >}}
  即使已有代码，只要测试覆盖率在 10% 以上就可以使用 TDD。因为在 Pre-RED 阶段先把握现有行为再写测试，所以可以在安全保存既有功能的同时添加新功能。
{{< /callout >}}

## 什么是 DDD？

**DDD** (Domain-Driven Development) 是**安全改进代码的方法**，一种尊重现有代码、渐进式改进的方式。用于几乎没有测试（低于 10%）的现有项目。

### 房屋改造比喻

为初次接触 DDD 的读者，用**房屋改造**来打比方。想象你要改造一栋住了 10 年的房子。

| 房屋改造阶段 | DDD 阶段 | 做的事 | 为什么重要 |
| --------------------- | --------------------- | ---------------------------------- | ----------------------------------------------------------- |
| 检查房子 | **ANALYZE**（分析） | 确认墙上的裂缝、管道状况 | 不知道哪里有问题就无从修起 |
| 给现状拍照 | **PRESERVE**（保存） | 给所有房间拍照存档 | 之后疑惑"原来这里有堵墙吗？"时可以核对 |
| 一间一间改造 | **IMPROVE**（改进） | 一次只施工一个房间，每次都验收 | 一下子全拆了，就不知道问题出在哪 |

**错误的方法 vs 正确的方法：**

```
잘못된 방법: "전체 코드를 한 번에 다 바꿀게요!"
  --> 기존 기능이 망가질 위험이 높습니다
  --> 문제가 생기면 어디서 잘못됐는지 찾기 어렵습니다

올바른 방법: "테스트로 현재 동작을 기록하고, 조금씩 바꿀게요!"
  --> 기존 기능이 망가지면 테스트가 바로 알려줍니다
  --> 문제가 생기면 마지막 변경만 되돌리면 됩니다
```

### ANALYZE-PRESERVE-IMPROVE 循环

MoAI-ADK 的 DDD 以重复三个阶段的循环推进。

```mermaid
flowchart TD
    A["ANALYZE\n分析代码结构\n把握问题点"] --> B["PRESERVE\n编写特性测试\n记录当前行为"]
    B --> C["IMPROVE\n渐进式代码改进\n确认测试通过"]
    C --> D{"全部测试\n通过了吗？"}
    D -->|"通过"| E["提交并\n进行下一项改进"]
    D -->|"失败"| F["回退\n最后一次变更"]
    F --> C
    E --> G{"全部需求\n实现完成？"}
    G -->|"还有剩余"| A
    G -->|"完成"| H["实现完成"]
```

### 第 1 步：ANALYZE（分析）

彻底分析现有代码的结构，就像医生给病人诊察。

**分析项目：**

| 分析对象 | 确认内容 | 比喻 |
| ---------- | ---------------------------------- | ------------------ |
| 文件结构 | 有哪些文件、如何关联 | 核对房屋图纸 |
| 依赖 | 哪个模块依赖哪个模块 | 检查管道与电气布线 |
| 测试现状 | 现有测试有多少 | 核对已有保险 |
| 问题点 | 重复代码、安全漏洞、性能瓶颈 | 检查裂缝的墙、漏水 |

**manager-develop 生成的分析报告示例：**

```markdown
## 코드 분석 보고서

- 대상: src/auth/ (인증 모듈)
- 파일: 8개 Python 파일
- 코드 줄 수: 1,850줄
- 테스트 커버리지: 5%

## 발견된 문제
1. 중복된 인증 로직 (3곳에서 동일한 코드 반복)
2. 하드코딩된 비밀키 (config.py에 직접 작성)
3. SQL Injection 취약점 (user_repository.py)
4. 불충분한 테스트 (5%, 목표 85%)
```

### 第 2 步：PRESERVE（保存）

为保存现有行为构建**安全网**。这一阶段的核心是编写**特性测试** (Characterization Tests)。

{{< callout type="info" >}}
  **特性测试是什么？**

  就像房屋改造前**给现状拍照存档**。

  一般的测试确认"这是否正确地工作？"。而特性测试记录的是"这现在是如何工作的？"。

  也就是说，不是判断对错，而是**记录"原本就是这样工作的"这一事实**。之后修改代码时测试若失败，就能立即知道既有行为被改变了。
{{< /callout >}}

**特性测试示例：**

```python
class TestExistingLoginBehavior:
    """기존 로그인 함수의 현재 동작을 기록하는 특성화 테스트"""

    def test_valid_login_returns_token(self):
        """
        GIVEN: 등록된 사용자가 있고
        WHEN: 올바른 비밀번호로 로그인하면
        THEN: 현재 구현이 반환하는 응답을 그대로 기록
        """
        user = create_test_user(
            email="test@example.com",
            password="password123"
        )

        result = login_service.login("test@example.com", "password123")

        # 현재 동작을 그대로 기록 (옳고 그름을 판단하지 않음)
        assert result["status"] == "success"
        assert result["token"] is not None
        assert result["expires_in"] == 3600  # 현재 만료 시간

    def test_wrong_password_returns_error(self):
        """잘못된 비밀번호로 로그인 시 현재 동작 기록"""
        create_test_user(email="test@example.com", password="password123")

        result = login_service.login("test@example.com", "wrongpassword")

        assert result["status"] == "error"
        assert result["code"] == 401
```

**测试编写策略：**

```mermaid
flowchart TD
    A["分析现有代码"] --> B["列出主要行为清单"]
    B --> C["为每个行为编写\n特性测试"]
    C --> D["运行全部测试"]
    D --> E{"全部测试\n通过？"}
    E -->|"通过"| F["安全网构建完成\n可以开始重构"]
    E -->|"失败"| G["修改测试\n调整为符合当前行为"]
    G --> D
```

### 第 3 步：IMPROVE（改进）

特性测试构建完成后，就可以安全地改进代码了。核心原则是**拆成小步骤地变更**。

**改进过程：**

```python
# BEFORE: 개선 전 코드
def login(email, password):
    # SQL Injection 취약점
    user = db.query("SELECT * FROM users WHERE email = '" + email + "'")
    if user and check_password(user.password, password):
        token = generate_token(user.id)
        return {"status": "success", "token": token}
    return {"status": "error", "code": 401}

# ====================================

# AFTER: 개선 후 코드 (3번의 반복을 거쳐 완성)
def login(email: str, password: str) -> LoginResult:
    """사용자 로그인을 처리합니다."""
    # 반복 1: 파라미터화된 쿼리로 SQL Injection 방지
    user = user_repository.find_by_email(email)

    if not user:
        return LoginResult.failure("자격증명이 유효하지 않습니다")

    # 반복 2: 인증 로직 중앙화
    if not auth_service.verify_password(user, password):
        return LoginResult.failure("자격증명이 유효하지 않습니다")

    # 반복 3: 토큰 서비스 분리
    token = token_service.generate(user.id)
    return LoginResult.success(token)
```

**渐进式改进步骤：**

```mermaid
flowchart TD
    S1["迭代 1：小变更\n修复 SQL Injection"] --> T1["运行测试\n156 个全部通过"]
    T1 --> C1["提交：保存安全状态"]
    C1 --> S2["迭代 2：小变更\n认证逻辑中心化"]
    S2 --> T2["运行测试\n156 个全部通过"]
    T2 --> C2["提交：保存安全状态"]
    C2 --> S3["迭代 3：小变更\n分离令牌服务"]
    S3 --> T3["运行测试\n156 个全部通过"]
    T3 --> C3["提交：改进完成"]
```

{{< callout type="warning" >}}
  **核心原则：** 每次变更后必须运行测试。测试失败时只需回退最后一次变更。这就是"小步骤"的力量。一次改动太多，就很难找出问题出在哪。
{{< /callout >}}

## 方法论对比

| 视角 | TDD | DDD |
| ----------------- | --------------------------- | ---------------------------- |
| **测试时机** | 编写代码前 (RED) | 分析后 (PRESERVE) |
| **覆盖率取向** | 每次提交严格标准 | 渐进式改进 |
| **最佳场景** | 新项目、10%+ 覆盖率 | 覆盖率低于 10% 的遗留代码 |
| **风险水平** | 中（需要纪律） | 低（保存行为） |
| **覆盖率例外** | 不允许 | 允许 |
| **Run Phase 循环** | RED-GREEN-REFACTOR | ANALYZE-PRESERVE-IMPROVE |

{{< callout type="warning" >}}
  **方法论选择指南：**

  - **新项目**（绿地）：TDD（默认值）
  - **现有项目**（覆盖率 50% 以上）：TDD
  - **现有项目**（覆盖率 10-49%）：TDD（利用 Pre-RED 阶段）
  - **现有项目**（覆盖率低于 10%）：DDD（渐进式特性测试）
{{< /callout >}}

## 什么是特性测试？

特性测试是 DDD 的核心工具。让我们更详细地了解一下。

### 与一般测试的区别

| 类别 | 一般测试 | 特性测试 |
| ------------- | ------------------------------- | ------------------------------ |
| **目的** | "这是否正确地工作？" | "这现在是如何工作的？" |
| **编写时机** | 编写新代码前/后 | 重构现有代码前 |
| **标准** | 需求（设计书） | 当前的实际行为 |
| **比喻** | 确认是否按图纸建造 | 用照片记录房屋现状 |

### 编写原则

1. **只记录不评判**：即使当前代码有 bug，也照原样记录其行为
2. **包含边界情况**：不仅记录正常情况，也记录全部异常情况
3. **可重现**：测试运行多少次都应得到同样的结果
4. **要快**：特性测试必须跑得快，才能在每次变更后立即验证

## 执行方法

### 执行 TDD

SPEC 文档准备好后，用以下命令执行 TDD 循环。

```bash
# TDD 실행 (development_mode: tdd 일 때)
> /moai run SPEC-AUTH-001
```

执行该命令后，**manager-develop 智能体**会自动执行 RED-GREEN-REFACTOR 循环：

```mermaid
flowchart TD
    A["阅读 SPEC 文档\nSPEC-AUTH-001"] --> B["RED\n为每条需求编写失败测试"]
    B --> C["GREEN\n用最少代码通过测试"]
    C --> D["REFACTOR\n改进代码质量\n保持测试"]
    D --> E{"还有下一条\n需求吗？"}
    E -->|"有"| B
    E -->|"没有"| F["最终验证\n确认覆盖率 85%+\n通过 TRUST 5 门禁"]
    F --> G["实现完成\n可进入 Sync 阶段"]
```

### 执行 DDD

```bash
# DDD 실행 (development_mode: ddd 일 때)
> /moai run SPEC-AUTH-001
```

执行该命令后，**manager-develop 智能体**会自动执行 ANALYZE-PRESERVE-IMPROVE 循环：

```mermaid
flowchart TD
    A["阅读 SPEC 文档\nSPEC-AUTH-001"] --> B["ANALYZE\n分析代码结构\n把握依赖"]
    B --> C["PRESERVE\n编写特性测试\n确立基线"]
    C --> D["IMPROVE\n迭代 1：认证逻辑中心化\n确认测试通过"]
    D --> E["IMPROVE\n迭代 2：密钥环境变量化\n确认测试通过"]
    E --> F["IMPROVE\n迭代 3：修复 SQL Injection\n确认测试通过"]
    F --> G["最终验证\n确认覆盖率 85%+\n通过 TRUST 5 门禁"]
    G --> H["实现完成\n可进入 Sync 阶段"]
```

## 方法论设置

在 `.moai/config/sections/quality.yaml` 文件中设置开发方法论。

### TDD 设置（默认值）

```yaml
constitution:
  development_mode: tdd  # TDD 방법론 사용

  tdd_settings:
    test_first_required: true         # 구현 전 테스트 작성 필수
    red_green_refactor: true          # RED-GREEN-REFACTOR 사이클 준수
    min_coverage_per_commit: 80       # 커밋당 최소 커버리지
    mutation_testing_enabled: false   # 뮤테이션 테스트 (선택)

  test_coverage_target: 85            # 전체 커버리지 목표
```

### DDD 设置

```yaml
constitution:
  development_mode: ddd  # DDD 방법론 사용

  ddd_settings:
    require_existing_tests: true      # 리팩토링 전 기존 테스트 필요
    characterization_tests: true      # 특성화 테스트 자동 생성
    behavior_snapshots: true          # 스냅샷 테스트 사용
    max_transformation_size: small    # 변경 크기 제한
    preserve_before_improve: true     # 보존 후 개선 필수

  test_coverage_target: 85            # 전체 커버리지 목표
```

**DDD max_transformation_size 选项：**

| 值 | 变更范围 | 推荐场景 |
| -------- | ------------------------ | -------------------------------- |
| `small` | 1-2 个文件，简单重构 | 一般的代码改进（推荐） |
| `medium` | 3-5 个文件，中等复杂度 | 模块结构变更 |
| `large` | 10 个以上文件 | 架构变更（需谨慎） |

{{< callout type="warning" >}}
  将 `max_transformation_size` 设为 `large` 会一次变更很多文件，出问题时难以定位原因。建议尽量保持 `small`。
{{< /callout >}}

## 实战示例：重构遗留代码

这是一个重构 3 年前编写的认证模块的场景。测试覆盖率只有 5%，非常低，因此使用 DDD 方法论。

### 现状

```
문제점:
- SQL Injection 취약점 2곳
- 하드코딩된 비밀키
- 중복된 인증 로직 3곳
- 테스트 커버리지 5%
- 코드 복잡도 높음
```

### 执行过程

```bash
# 1단계: SPEC 생성 (Plan)
> /moai plan "레거시 인증 시스템 리팩토링. SQL Injection 수정, 비밀키 환경 변수화, 인증 로직 중앙화"

# manager-spec이 SPEC-AUTH-REFACTOR-001 생성
```

```bash
# 2단계: DDD 실행 (Run)
> /moai run SPEC-AUTH-REFACTOR-001

# manager-develop가 ANALYZE-PRESERVE-IMPROVE 사이클 실행
# ANALYZE: 코드 분석, 문제점 목록 생성
# PRESERVE: 특성화 테스트 156개 작성
# IMPROVE: 3번의 반복으로 점진적 개선
```

```bash
# 3단계: 문서 동기화 (Sync)
> /moai sync SPEC-AUTH-REFACTOR-001

# manager-docs가 API 문서 업데이트, 리팩토링 보고서 생성
```

### 结果

| 指标 | Before | After | 变化 |
| ------------------ | ------ | -------- | -------------- |
| 测试覆盖率 | 5% | 87% | +82% |
| SQL Injection 漏洞 | 2 处 | 0 处 | 清除完成 |
| 硬编码密钥 | 有 | 无 | 环境变量化 |
| 重复代码 | 3 处 | 0 处 | 中心化完成 |
| 代码复杂度 | 高 | 降低 35% | 结构改进 |

{{< callout type="info" >}}
  **核心要点：** 重构过程中，既有行为没有发生任何一处改变。156 个特性测试在每次迭代中全部通过，因此在不影响现有用户的前提下大幅提升了代码质量。
{{< /callout >}}

## 相关文档

- [基于 SPEC 的开发](/zh/core-concepts/spec-based-dev) -- 执行开发方法论之前需要 SPEC 文档
- [TRUST 5 质量](/zh/core-concepts/trust-5) -- 确认实现完成后的质量验证标准
